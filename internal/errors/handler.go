package errors

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SchoolGPTError represents a structured error in the SchoolGPT system
type SchoolGPTError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   string                 `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	SchoolID  string                 `json:"school_id,omitempty"`
	Stack     string                 `json:"stack,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Severity  string                 `json:"severity"`
	Category  string                 `json:"category"`
	Retryable bool                   `json:"retryable"`
}

func (e *SchoolGPTError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Category, e.Message)
}

// Error codes
const (
	// Database errors
	ErrCodeDatabaseConnection = "DB_CONNECTION_FAILED"
	ErrCodeDatabaseTimeout    = "DB_TIMEOUT"
	ErrCodeDatabaseConstraint = "DB_CONSTRAINT_VIOLATION"
	ErrCodeDatabaseDeadlock   = "DB_DEADLOCK"
	ErrCodeSchoolNotFound     = "SCHOOL_NOT_FOUND"
	ErrCodeUserNotFound       = "USER_NOT_FOUND"
	ErrCodeStudentNotFound    = "STUDENT_NOT_FOUND"
	ErrCodeTeacherNotFound    = "TEACHER_NOT_FOUND"

	// Authentication & Authorization errors
	ErrCodeUnauthorized       = "UNAUTHORIZED_ACCESS"
	ErrCodeForbidden          = "FORBIDDEN_ACCESS"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeSessionExpired     = "SESSION_EXPIRED"
	ErrCodeInvalidToken       = "INVALID_TOKEN"
	ErrCodeAccountLocked      = "ACCOUNT_LOCKED"

	// Validation errors
	ErrCodeValidationFailed  = "VALIDATION_FAILED"
	ErrCodeInvalidInput      = "INVALID_INPUT"
	ErrCodeMissingField      = "MISSING_REQUIRED_FIELD"
	ErrCodeInvalidFormat     = "INVALID_FORMAT"
	ErrCodeDuplicateEntry    = "DUPLICATE_ENTRY"
	ErrCodeDuplicateResource = "DUPLICATE_RESOURCE"
	ErrCodeResourceNotFound  = "RESOURCE_NOT_FOUND"

	// Business logic errors
	ErrCodeBusinessRule            = "BUSINESS_RULE_VIOLATION"
	ErrCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
	ErrCodeResourceLimit           = "RESOURCE_LIMIT_EXCEEDED"
	ErrCodeOperationNotAllowed     = "OPERATION_NOT_ALLOWED"

	// External service errors
	ErrCodeExternalService   = "EXTERNAL_SERVICE_ERROR"
	ErrCodeAPIRateLimit      = "API_RATE_LIMIT_EXCEEDED"
	ErrCodeThirdPartyTimeout = "THIRD_PARTY_TIMEOUT"

	// System errors
	ErrCodeInternalServer     = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout            = "REQUEST_TIMEOUT"
	ErrCodeResourceExhausted  = "RESOURCE_EXHAUSTED"
	ErrCodeSystemOverloaded   = "SYSTEM_OVERLOADED"

	// Security errors
	ErrCodeSecurityThreat        = "SECURITY_THREAT_DETECTED"
	ErrCodeRateLimitExceeded     = "RATE_LIMIT_EXCEEDED"
	ErrCodeSuspiciousActivity    = "SUSPICIOUS_ACTIVITY"
	ErrCodeSchoolIsolationBreach = "SCHOOL_ISOLATION_BREACH"

	// Configuration errors
	ErrCodeConfigurationError   = "CONFIGURATION_ERROR"
	ErrCodeFeatureNotEnabled    = "FEATURE_NOT_ENABLED"
	ErrCodeUnsupportedOperation = "UNSUPPORTED_OPERATION"
	ErrCodeResourceConflict     = "RESOURCE_CONFLICT"
)

// Error categories
const (
	CategoryDatabase      = "database"
	CategoryAuth          = "authentication"
	CategoryValidation    = "validation"
	CategoryBusiness      = "business"
	CategoryExternal      = "external"
	CategorySystem        = "system"
	CategorySecurity      = "security"
	CategoryConfiguration = "configuration"
	CategoryMonitoring    = "monitoring"
)

// Severity levels
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// Predefined error instances
var (
	ErrDatabaseConnection = &SchoolGPTError{
		Code:      ErrCodeDatabaseConnection,
		Message:   "Database connection failed",
		Category:  CategoryDatabase,
		Severity:  SeverityCritical,
		Retryable: true,
	}

	ErrSchoolNotFound = &SchoolGPTError{
		Code:      ErrCodeSchoolNotFound,
		Message:   "School not found",
		Category:  CategoryDatabase,
		Severity:  SeverityMedium,
		Retryable: false,
	}

	ErrUnauthorized = &SchoolGPTError{
		Code:      ErrCodeUnauthorized,
		Message:   "Unauthorized access",
		Category:  CategoryAuth,
		Severity:  SeverityMedium,
		Retryable: false,
	}

	ErrValidationFailed = &SchoolGPTError{
		Code:      ErrCodeValidationFailed,
		Message:   "Input validation failed",
		Category:  CategoryValidation,
		Severity:  SeverityLow,
		Retryable: false,
	}

	ErrRateLimitExceeded = &SchoolGPTError{
		Code:      ErrCodeRateLimitExceeded,
		Message:   "Rate limit exceeded",
		Category:  CategorySecurity,
		Severity:  SeverityMedium,
		Retryable: true,
	}

	ErrInternalServer = &SchoolGPTError{
		Code:      ErrCodeInternalServer,
		Message:   "Internal server error",
		Category:  CategorySystem,
		Severity:  SeverityCritical,
		Retryable: true,
	}
)

// ErrorMonitor tracks error patterns and metrics
type ErrorMonitor struct {
	errorCounts     sync.Map // map[string]int64
	errorRates      sync.Map // map[string]*ErrorRate
	patterns        sync.Map // map[string]*ErrorPattern
	alertThresholds map[string]int64
	mu              sync.RWMutex
}

// ErrorRate tracks error rates over time
type ErrorRate struct {
	WindowSize time.Duration
	Count      int64
	LastReset  time.Time
	mu         sync.Mutex
}

// ErrorPattern detects error patterns
type ErrorPattern struct {
	ErrorCode      string
	Count          int64
	LastOccurrence time.Time
	UserIDs        map[string]int64
	SchoolIDs      map[string]int64
	IPAddresses    map[string]int64
}

// NewErrorMonitor creates a new error monitor
func NewErrorMonitor() *ErrorMonitor {
	em := &ErrorMonitor{
		alertThresholds: map[string]int64{
			ErrCodeDatabaseConnection: 5,
			ErrCodeSecurityThreat:     1,
			ErrCodeRateLimitExceeded:  10,
			ErrCodeInternalServer:     10,
		},
	}

	// Start background cleanup
	go em.cleanup()

	return em
}

// RecordError records an error occurrence
func (em *ErrorMonitor) RecordError(err *SchoolGPTError) {
	// Update error count
	count, _ := em.errorCounts.LoadOrStore(err.Code, int64(0))
	newCount := count.(int64) + 1
	em.errorCounts.Store(err.Code, newCount)

	// Update error rate
	em.updateErrorRate(err.Code)

	// Update error pattern
	em.updateErrorPattern(err)

	// Check alert thresholds
	if threshold, exists := em.alertThresholds[err.Code]; exists && newCount >= threshold {
		em.triggerAlert(err, newCount)
	}
}

func (em *ErrorMonitor) updateErrorRate(errorCode string) {
	rateInterface, exists := em.errorRates.Load(errorCode)
	if !exists {
		rate := &ErrorRate{
			WindowSize: time.Minute * 5,
			Count:      1,
			LastReset:  time.Now(),
		}
		em.errorRates.Store(errorCode, rate)
		return
	}

	rate := rateInterface.(*ErrorRate)
	rate.mu.Lock()
	defer rate.mu.Unlock()

	now := time.Now()
	if now.Sub(rate.LastReset) > rate.WindowSize {
		rate.Count = 1
		rate.LastReset = now
	} else {
		rate.Count++
	}
}

func (em *ErrorMonitor) updateErrorPattern(err *SchoolGPTError) {
	patternInterface, exists := em.patterns.Load(err.Code)
	if !exists {
		pattern := &ErrorPattern{
			ErrorCode:      err.Code,
			Count:          1,
			LastOccurrence: time.Now(),
			UserIDs:        make(map[string]int64),
			SchoolIDs:      make(map[string]int64),
			IPAddresses:    make(map[string]int64),
		}
		if err.UserID != "" {
			pattern.UserIDs[err.UserID] = 1
		}
		if err.SchoolID != "" {
			pattern.SchoolIDs[err.SchoolID] = 1
		}
		em.patterns.Store(err.Code, pattern)
		return
	}

	pattern := patternInterface.(*ErrorPattern)
	pattern.Count++
	pattern.LastOccurrence = time.Now()

	if err.UserID != "" {
		pattern.UserIDs[err.UserID]++
	}
	if err.SchoolID != "" {
		pattern.SchoolIDs[err.SchoolID]++
	}
}

func (em *ErrorMonitor) triggerAlert(err *SchoolGPTError, count int64) {
	// This would typically send alerts to monitoring systems like PagerDuty, Slack, etc.
	fmt.Printf("[ALERT] Error threshold exceeded: %s (count: %d)\n", err.Code, count)
}

func (em *ErrorMonitor) cleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-24 * time.Hour)

		em.patterns.Range(func(key, value interface{}) bool {
			pattern := value.(*ErrorPattern)
			if pattern.LastOccurrence.Before(cutoff) {
				em.patterns.Delete(key)
			}
			return true
		})
	}
}

// AlertManager handles error alerts
type AlertManager struct {
	channels map[string]AlertChannel
	rules    []AlertRule
	mu       sync.RWMutex
}

// AlertChannel represents an alert destination
type AlertChannel interface {
	Send(alert Alert) error
}

// AlertRule defines when to send alerts
type AlertRule struct {
	ErrorCode  string
	Threshold  int64
	TimeWindow time.Duration
	Severity   string
	Channels   []string
}

// Alert represents an alert to be sent
type Alert struct {
	Title     string
	Message   string
	Severity  string
	Timestamp time.Time
	ErrorCode string
	Count     int64
	Metadata  map[string]interface{}
}

// ErrorHandler provides comprehensive error handling
type ErrorHandler struct {
	logger        *zap.Logger
	monitor       *ErrorMonitor
	alerting      *AlertManager
	recovery      *RecoveryManager
	httpStatusMap map[string]int
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *zap.Logger) *ErrorHandler {
	eh := &ErrorHandler{
		logger:   logger,
		monitor:  NewErrorMonitor(),
		alerting: NewAlertManager(),
		recovery: NewRecoveryManager(),
		httpStatusMap: map[string]int{
			ErrCodeDatabaseConnection: http.StatusServiceUnavailable,
			ErrCodeDatabaseTimeout:    http.StatusGatewayTimeout,
			ErrCodeSchoolNotFound:     http.StatusNotFound,
			ErrCodeUserNotFound:       http.StatusNotFound,
			ErrCodeUnauthorized:       http.StatusUnauthorized,
			ErrCodeForbidden:          http.StatusForbidden,
			ErrCodeValidationFailed:   http.StatusBadRequest,
			ErrCodeInvalidInput:       http.StatusBadRequest,
			ErrCodeBusinessRule:       http.StatusUnprocessableEntity,
			ErrCodeExternalService:    http.StatusBadGateway,
			ErrCodeAPIRateLimit:       http.StatusTooManyRequests,
			ErrCodeInternalServer:     http.StatusInternalServerError,
			ErrCodeServiceUnavailable: http.StatusServiceUnavailable,
			ErrCodeTimeout:            http.StatusRequestTimeout,
			ErrCodeRateLimitExceeded:  http.StatusTooManyRequests,
			ErrCodeSecurityThreat:     http.StatusForbidden,
			ErrCodeConfigurationError: http.StatusInternalServerError,
		},
	}

	return eh
}

// HandleError handles an error and returns appropriate response
func (eh *ErrorHandler) HandleError(err error, ctx *gin.Context) {
	requestID := ctx.GetString("request_id")
	userID := ctx.GetString("user_id")
	schoolID := ctx.GetString("school_id")

	// Convert to SchoolGPT error
	schoolGPTErr := eh.convertError(err, requestID, userID, schoolID)

	// Add stack trace for critical errors
	if schoolGPTErr.Severity == SeverityCritical {
		schoolGPTErr.Stack = eh.getStackTrace()
	}

	// Log error with context
	eh.logError(schoolGPTErr, ctx)

	// Record error metrics
	eh.monitor.RecordError(schoolGPTErr)

	// Attempt recovery for retryable errors
	if schoolGPTErr.Retryable {
		if eh.recovery.ShouldRetry(schoolGPTErr) {
			// In a real implementation, you might queue for retry
			eh.logger.Info("Error marked for retry", zap.String("error_code", schoolGPTErr.Code))
		}
	}

	// Alert on critical errors
	if eh.isCriticalError(schoolGPTErr) {
		eh.sendAlert(schoolGPTErr)
	}

	// Return sanitized error to client
	response := eh.sanitizeErrorForClient(schoolGPTErr)
	httpStatus := eh.getHTTPStatus(schoolGPTErr)

	ctx.JSON(httpStatus, gin.H{
		"error":      response,
		"request_id": requestID,
		"timestamp":  time.Now().UTC(),
	})
}

// convertError converts a generic error to SchoolGPTError
func (eh *ErrorHandler) convertError(err error, requestID, userID, schoolID string) *SchoolGPTError {
	// If it's already a SchoolGPTError, enhance it
	if schoolGPTErr, ok := err.(*SchoolGPTError); ok {
		schoolGPTErr.Timestamp = time.Now()
		schoolGPTErr.RequestID = requestID
		schoolGPTErr.UserID = userID
		schoolGPTErr.SchoolID = schoolID
		return schoolGPTErr
	}

	// Detect error type from message
	errMsg := err.Error()

	var code, category, severity string
	retryable := false

	switch {
	case strings.Contains(errMsg, "connection"):
		code = ErrCodeDatabaseConnection
		category = CategoryDatabase
		severity = SeverityCritical
		retryable = true
	case strings.Contains(errMsg, "timeout"):
		code = ErrCodeTimeout
		category = CategorySystem
		severity = SeverityHigh
		retryable = true
	case strings.Contains(errMsg, "not found"):
		code = ErrCodeSchoolNotFound
		category = CategoryDatabase
		severity = SeverityMedium
		retryable = false
	case strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "forbidden"):
		code = ErrCodeUnauthorized
		category = CategoryAuth
		severity = SeverityMedium
		retryable = false
	case strings.Contains(errMsg, "validation") || strings.Contains(errMsg, "invalid"):
		code = ErrCodeValidationFailed
		category = CategoryValidation
		severity = SeverityLow
		retryable = false
	default:
		code = ErrCodeInternalServer
		category = CategorySystem
		severity = SeverityCritical
		retryable = true
	}

	return &SchoolGPTError{
		Code:      code,
		Message:   errMsg,
		Category:  category,
		Severity:  severity,
		Retryable: retryable,
		Timestamp: time.Now(),
		RequestID: requestID,
		UserID:    userID,
		SchoolID:  schoolID,
	}
}

// logError logs the error with appropriate level
func (eh *ErrorHandler) logError(err *SchoolGPTError, ctx *gin.Context) {
	fields := []zap.Field{
		zap.String("error_code", err.Code),
		zap.String("category", err.Category),
		zap.String("severity", err.Severity),
		zap.String("request_id", err.RequestID),
		zap.String("user_id", err.UserID),
		zap.String("school_id", err.SchoolID),
		zap.String("path", ctx.Request.URL.Path),
		zap.String("method", ctx.Request.Method),
		zap.String("ip", ctx.ClientIP()),
		zap.String("user_agent", ctx.Request.UserAgent()),
		zap.Bool("retryable", err.Retryable),
	}

	if err.Stack != "" {
		fields = append(fields, zap.String("stack", err.Stack))
	}

	switch err.Severity {
	case SeverityCritical:
		eh.logger.Error(err.Message, fields...)
	case SeverityHigh:
		eh.logger.Warn(err.Message, fields...)
	case SeverityMedium:
		eh.logger.Info(err.Message, fields...)
	default:
		eh.logger.Debug(err.Message, fields...)
	}
}

// getStackTrace gets the current stack trace
func (eh *ErrorHandler) getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// isCriticalError checks if an error is critical
func (eh *ErrorHandler) isCriticalError(err *SchoolGPTError) bool {
	return err.Severity == SeverityCritical ||
		err.Code == ErrCodeSecurityThreat ||
		err.Code == ErrCodeSchoolIsolationBreach
}

// sendAlert sends an alert for critical errors
func (eh *ErrorHandler) sendAlert(err *SchoolGPTError) {
	// Implementation would send to monitoring systems
	eh.logger.Error("CRITICAL ERROR ALERT",
		zap.String("error_code", err.Code),
		zap.String("message", err.Message),
		zap.String("request_id", err.RequestID),
	)
}

// sanitizeErrorForClient removes sensitive information from error
func (eh *ErrorHandler) sanitizeErrorForClient(err *SchoolGPTError) *SchoolGPTError {
	sanitized := &SchoolGPTError{
		Code:      err.Code,
		Message:   err.Message,
		Timestamp: err.Timestamp,
		RequestID: err.RequestID,
		Severity:  err.Severity,
		Category:  err.Category,
		Retryable: err.Retryable,
	}

	// Don't expose sensitive information
	if err.Severity == SeverityCritical {
		sanitized.Message = "An internal error occurred. Please try again later."
		sanitized.Details = ""
	}

	return sanitized
}

// getHTTPStatus returns appropriate HTTP status code
func (eh *ErrorHandler) getHTTPStatus(err *SchoolGPTError) int {
	if status, exists := eh.httpStatusMap[err.Code]; exists {
		return status
	}
	return http.StatusInternalServerError
}

// RecoveryManager handles error recovery
type RecoveryManager struct {
	retryAttempts map[string]int
	backoffConfig map[string]time.Duration
	mu            sync.RWMutex
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager() *RecoveryManager {
	return &RecoveryManager{
		retryAttempts: make(map[string]int),
		backoffConfig: map[string]time.Duration{
			ErrCodeDatabaseConnection: time.Second * 5,
			ErrCodeExternalService:    time.Second * 2,
			ErrCodeTimeout:            time.Second * 3,
		},
	}
}

// ShouldRetry determines if an error should be retried
func (rm *RecoveryManager) ShouldRetry(err *SchoolGPTError) bool {
	if !err.Retryable {
		return false
	}

	rm.mu.RLock()
	attempts := rm.retryAttempts[err.RequestID]
	rm.mu.RUnlock()

	maxAttempts := 3
	return attempts < maxAttempts
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		channels: make(map[string]AlertChannel),
		rules:    make([]AlertRule, 0),
	}
}

// NewSchoolGPTError creates a new SchoolGPT error
func NewSchoolGPTError(code, message, category, severity string, retryable bool) *SchoolGPTError {
	return &SchoolGPTError{
		Code:      code,
		Message:   message,
		Category:  category,
		Severity:  severity,
		Retryable: retryable,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// WithDetails adds details to an error
func (e *SchoolGPTError) WithDetails(details string) *SchoolGPTError {
	e.Details = details
	return e
}

// WithMetadata adds metadata to an error
func (e *SchoolGPTError) WithMetadata(key string, value interface{}) *SchoolGPTError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithContext adds context information to an error
func (e *SchoolGPTError) WithContext(requestID, userID, schoolID string) *SchoolGPTError {
	e.RequestID = requestID
	e.UserID = userID
	e.SchoolID = schoolID
	return e
}
