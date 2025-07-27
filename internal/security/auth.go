package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

// SecurityContext represents the security context for a request
type SecurityContext struct {
	UserID       string    `json:"user_id"`
	SchoolID     string    `json:"school_id"`
	Role         string    `json:"role"`
	Permissions  []string  `json:"permissions"`
	SessionToken string    `json:"session_token"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	RequestID    string    `json:"request_id"`
	LoginTime    time.Time `json:"login_time"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// RoleBasedAccessControl manages permissions and roles
type RoleBasedAccessControl struct {
	rolePermissions map[string][]string
	resourceActions map[string][]string
	mu              sync.RWMutex
}

// Permission constants
const (
	// School management permissions
	PermissionSchoolCreate = "school:create"
	PermissionSchoolRead   = "school:read"
	PermissionSchoolUpdate = "school:update"
	PermissionSchoolDelete = "school:delete"

	// User management permissions
	PermissionUserCreate = "user:create"
	PermissionUserRead   = "user:read"
	PermissionUserUpdate = "user:update"
	PermissionUserDelete = "user:delete"

	// Student management permissions
	PermissionStudentCreate = "student:create"
	PermissionStudentRead   = "student:read"
	PermissionStudentUpdate = "student:update"
	PermissionStudentDelete = "student:delete"

	// Teacher management permissions
	PermissionTeacherCreate = "teacher:create"
	PermissionTeacherRead   = "teacher:read"
	PermissionTeacherUpdate = "teacher:update"
	PermissionTeacherDelete = "teacher:delete"

	// Configuration permissions
	PermissionConfigRead   = "config:read"
	PermissionConfigUpdate = "config:update"

	// Analytics permissions
	PermissionAnalyticsRead = "analytics:read"

	// System permissions
	PermissionSystemAdmin = "system:admin"
)

// Role constants
const (
	RoleSystemAdmin    = "system_admin"
	RoleSchoolAdmin    = "school_admin"
	RoleTeacher        = "teacher"
	RoleParent         = "parent"
	RoleStudent        = "student"
	RoleAccountManager = "account_manager"
)

// NewRBAC creates a new role-based access control system
func NewRBAC() *RoleBasedAccessControl {
	rbac := &RoleBasedAccessControl{
		rolePermissions: make(map[string][]string),
		resourceActions: make(map[string][]string),
	}

	// Define role permissions
	rbac.rolePermissions[RoleSystemAdmin] = []string{
		PermissionSystemAdmin,
		PermissionSchoolCreate, PermissionSchoolRead, PermissionSchoolUpdate, PermissionSchoolDelete,
		PermissionUserCreate, PermissionUserRead, PermissionUserUpdate, PermissionUserDelete,
		PermissionConfigRead, PermissionConfigUpdate,
		PermissionAnalyticsRead,
	}

	rbac.rolePermissions[RoleSchoolAdmin] = []string{
		PermissionSchoolRead, PermissionSchoolUpdate,
		PermissionUserCreate, PermissionUserRead, PermissionUserUpdate, PermissionUserDelete,
		PermissionStudentCreate, PermissionStudentRead, PermissionStudentUpdate, PermissionStudentDelete,
		PermissionTeacherCreate, PermissionTeacherRead, PermissionTeacherUpdate, PermissionTeacherDelete,
		PermissionConfigRead, PermissionConfigUpdate,
		PermissionAnalyticsRead,
	}

	rbac.rolePermissions[RoleTeacher] = []string{
		PermissionStudentRead, PermissionStudentUpdate,
		PermissionConfigRead,
	}

	rbac.rolePermissions[RoleParent] = []string{
		PermissionStudentRead, // Only for their own children
		PermissionConfigRead,
	}

	rbac.rolePermissions[RoleStudent] = []string{
		PermissionConfigRead,
	}

	rbac.rolePermissions[RoleAccountManager] = []string{
		PermissionSchoolRead,
		PermissionUserRead,
		PermissionAnalyticsRead,
	}

	return rbac
}

// HasPermission checks if a role has a specific permission
func (rbac *RoleBasedAccessControl) HasPermission(role, resource, action string) bool {
	rbac.mu.RLock()
	defer rbac.mu.RUnlock()

	permissions, exists := rbac.rolePermissions[role]
	if !exists {
		return false
	}

	permissionNeeded := fmt.Sprintf("%s:%s", resource, action)

	for _, permission := range permissions {
		if permission == permissionNeeded || permission == PermissionSystemAdmin {
			return true
		}
	}

	return false
}

// RateLimiter manages request rate limiting
type RateLimiter struct {
	limiters sync.Map // map[string]*rate.Limiter
	config   *RateLimitConfig
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		config: config,
	}
}

// Allow checks if a request should be allowed based on rate limits
func (rl *RateLimiter) Allow(userID, ipAddress string) bool {
	// Rate limit by user ID
	if !rl.allowForKey(userID, rl.config.RequestsPerMinute) {
		return false
	}

	// Rate limit by IP address (more permissive)
	if !rl.allowForKey(ipAddress, rl.config.RequestsPerMinute*2) {
		return false
	}

	return true
}

func (rl *RateLimiter) allowForKey(key string, requestsPerMinute int) bool {
	limiterInterface, exists := rl.limiters.Load(key)
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(requestsPerMinute)/60, rl.config.BurstSize)
		rl.limiters.Store(key, limiter)
		return limiter.Allow()
	}

	limiter, ok := limiterInterface.(*rate.Limiter)
	if !ok {
		return false
	}

	return limiter.Allow()
}

// AuditLogger logs security events
type AuditLogger struct {
	events chan AuditEvent
	buffer []AuditEvent
	mu     sync.Mutex
}

// AuditEvent represents a security audit event
type AuditEvent struct {
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	SchoolID  string                 `json:"school_id"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	Resource  string                 `json:"resource"`
	Action    string                 `json:"action"`
	Success   bool                   `json:"success"`
	Details   map[string]interface{} `json:"details"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id"`
	Severity  string                 `json:"severity"`
}

// Audit event types
const (
	AuditEventLogin                 = "login"
	AuditEventLogout                = "logout"
	AuditEventPermissionDenied      = "permission_denied"
	AuditEventRateLimitExceeded     = "rate_limit_exceeded"
	AuditEventSuspiciousActivity    = "suspicious_activity"
	AuditEventSecurityViolation     = "security_violation"
	AuditEventDataAccess            = "data_access"
	AuditEventConfigChange          = "config_change"
	AuditEventSchoolIsolationBreach = "school_isolation_breach"
)

// Severity levels
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	al := &AuditLogger{
		events: make(chan AuditEvent, 1000),
		buffer: make([]AuditEvent, 0, 100),
	}

	// Start background processor
	go al.processEvents()

	return al
}

// LogEvent logs an audit event
func (al *AuditLogger) LogEvent(event AuditEvent) {
	event.Timestamp = time.Now()

	select {
	case al.events <- event:
	default:
		// Channel is full, log to buffer
		al.mu.Lock()
		al.buffer = append(al.buffer, event)
		if len(al.buffer) > 100 {
			al.buffer = al.buffer[1:] // Remove oldest event
		}
		al.mu.Unlock()
	}
}

// LogUnauthorizedAccess logs unauthorized access attempts
func (al *AuditLogger) LogUnauthorizedAccess(ctx *SecurityContext, resource, action string) {
	al.LogEvent(AuditEvent{
		EventType: AuditEventPermissionDenied,
		UserID:    ctx.UserID,
		SchoolID:  ctx.SchoolID,
		IPAddress: ctx.IPAddress,
		UserAgent: ctx.UserAgent,
		Resource:  resource,
		Action:    action,
		Success:   false,
		RequestID: ctx.RequestID,
		Severity:  SeverityMedium,
		Details: map[string]interface{}{
			"role":        ctx.Role,
			"permissions": ctx.Permissions,
		},
	})
}

// LogSuspiciousActivity logs suspicious activity
func (al *AuditLogger) LogSuspiciousActivity(ctx *SecurityContext, activityType string) {
	al.LogEvent(AuditEvent{
		EventType: AuditEventSuspiciousActivity,
		UserID:    ctx.UserID,
		SchoolID:  ctx.SchoolID,
		IPAddress: ctx.IPAddress,
		UserAgent: ctx.UserAgent,
		Success:   false,
		RequestID: ctx.RequestID,
		Severity:  SeverityHigh,
		Details: map[string]interface{}{
			"activity_type": activityType,
		},
	})
}

// LogSecurityViolation logs security violations
func (al *AuditLogger) LogSecurityViolation(ctx *SecurityContext, violationType string) {
	al.LogEvent(AuditEvent{
		EventType: AuditEventSecurityViolation,
		UserID:    ctx.UserID,
		SchoolID:  ctx.SchoolID,
		IPAddress: ctx.IPAddress,
		UserAgent: ctx.UserAgent,
		Success:   false,
		RequestID: ctx.RequestID,
		Severity:  SeverityCritical,
		Details: map[string]interface{}{
			"violation_type": violationType,
		},
	})
}

// LogSuccessfulLogin logs successful login attempts
func (al *AuditLogger) LogSuccessfulLogin(ctx *SecurityContext) {
	al.LogEvent(AuditEvent{
		EventType: AuditEventLogin,
		UserID:    ctx.UserID,
		SchoolID:  ctx.SchoolID,
		IPAddress: ctx.IPAddress,
		UserAgent: ctx.UserAgent,
		Success:   true,
		RequestID: ctx.RequestID,
		Severity:  SeverityLow,
	})
}

// processEvents processes audit events in the background
func (al *AuditLogger) processEvents() {
	for event := range al.events {
		// Here you would typically write to a database, log file, or external service
		// For now, we'll just print to stdout (in production, use proper logging)
		fmt.Printf("[AUDIT] %s: %s@%s - %s:%s (Success: %t)\n",
			event.Timestamp.Format(time.RFC3339),
			event.UserID,
			event.IPAddress,
			event.Resource,
			event.Action,
			event.Success,
		)
	}
}

// AccessControl provides the main access control functionality
type AccessControl struct {
	rbac        *RoleBasedAccessControl
	audit       *AuditLogger
	rateLimiter *RateLimiter
	validator   *SecurityValidator
}

// NewAccessControl creates a new access control system
func NewAccessControl() *AccessControl {
	return &AccessControl{
		rbac:        NewRBAC(),
		audit:       NewAuditLogger(),
		rateLimiter: NewRateLimiter(DefaultRateLimitConfig()),
		validator:   NewSecurityValidator(),
	}
}

// ValidateAccess validates access for a security context
func (ac *AccessControl) ValidateAccess(ctx *SecurityContext, resource string, action string) error {
	// Rate limiting check
	if !ac.rateLimiter.Allow(ctx.UserID, ctx.IPAddress) {
		ac.audit.LogSuspiciousActivity(ctx, "rate_limit_exceeded")
		return NewSecurityError("RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	}

	// Permission validation
	if !ac.rbac.HasPermission(ctx.Role, resource, action) {
		ac.audit.LogUnauthorizedAccess(ctx, resource, action)
		return NewSecurityError("INSUFFICIENT_PERMISSIONS", "Insufficient permissions")
	}

	// School isolation check (for non-system-admin roles)
	if ctx.Role != RoleSystemAdmin && resource != "system" {
		if err := ac.validateSchoolAccess(ctx.UserID, ctx.SchoolID); err != nil {
			ac.audit.LogSecurityViolation(ctx, "school_isolation_breach")
			return NewSecurityError("SCHOOL_ISOLATION_VIOLATION", "School isolation violation")
		}
	}

	// Log successful access
	ac.audit.LogEvent(AuditEvent{
		EventType: AuditEventDataAccess,
		UserID:    ctx.UserID,
		SchoolID:  ctx.SchoolID,
		IPAddress: ctx.IPAddress,
		UserAgent: ctx.UserAgent,
		Resource:  resource,
		Action:    action,
		Success:   true,
		RequestID: ctx.RequestID,
		Severity:  SeverityLow,
	})

	return nil
}

// validateSchoolAccess checks if user has access to the specified school
func (ac *AccessControl) validateSchoolAccess(userID, schoolID string) error {
	// This would typically check against the database
	// For now, we'll implement a basic check
	if userID == "" || schoolID == "" {
		return errors.New("invalid user or school ID")
	}
	return nil
}

// SecurityError represents a security-related error
type SecurityError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (se *SecurityError) Error() string {
	return fmt.Sprintf("security error [%s]: %s", se.Code, se.Message)
}

// NewSecurityError creates a new security error
func NewSecurityError(code, message string) *SecurityError {
	return &SecurityError{
		Code:    code,
		Message: message,
	}
}

// PasswordManager handles password hashing and validation
type PasswordManager struct {
	minLength     int
	requireUpper  bool
	requireLower  bool
	requireDigit  bool
	requireSymbol bool
}

// NewPasswordManager creates a new password manager
func NewPasswordManager() *PasswordManager {
	return &PasswordManager{
		minLength:     8,
		requireUpper:  true,
		requireLower:  true,
		requireDigit:  true,
		requireSymbol: true,
	}
}

// HashPassword hashes a password using bcrypt
func (pm *PasswordManager) HashPassword(password string) (string, error) {
	if err := pm.ValidatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword verifies a password against its hash
func (pm *PasswordManager) VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ValidatePassword validates password strength
func (pm *PasswordManager) ValidatePassword(password string) error {
	if len(password) < pm.minLength {
		return fmt.Errorf("password must be at least %d characters long", pm.minLength)
	}

	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		default:
			hasSymbol = true
		}
	}

	if pm.requireUpper && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if pm.requireLower && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if pm.requireDigit && !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if pm.requireSymbol && !hasSymbol {
		return errors.New("password must contain at least one symbol")
	}

	return nil
}

// SessionManager manages user sessions
type SessionManager struct {
	sessions sync.Map // map[string]*SecurityContext
	timeout  time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(timeout time.Duration) *SessionManager {
	sm := &SessionManager{
		timeout: timeout,
	}

	// Start cleanup goroutine
	go sm.cleanupExpiredSessions()

	return sm
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(userID, schoolID, role string, permissions []string, ipAddress, userAgent string) (*SecurityContext, error) {
	sessionToken, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	ctx := &SecurityContext{
		UserID:       userID,
		SchoolID:     schoolID,
		Role:         role,
		Permissions:  permissions,
		SessionToken: sessionToken,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		LoginTime:    time.Now(),
		ExpiresAt:    time.Now().Add(sm.timeout),
	}

	sm.sessions.Store(sessionToken, ctx)
	return ctx, nil
}

// GetSession retrieves a session by token
func (sm *SessionManager) GetSession(token string) (*SecurityContext, error) {
	sessionInterface, exists := sm.sessions.Load(token)
	if !exists {
		return nil, errors.New("session not found")
	}

	ctx, ok := sessionInterface.(*SecurityContext)
	if !ok {
		return nil, errors.New("invalid session data")
	}

	if time.Now().After(ctx.ExpiresAt) {
		sm.sessions.Delete(token)
		return nil, errors.New("session expired")
	}

	return ctx, nil
}

// InvalidateSession invalidates a session
func (sm *SessionManager) InvalidateSession(token string) {
	sm.sessions.Delete(token)
}

// cleanupExpiredSessions cleans up expired sessions periodically
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(time.Minute * 15)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		sm.sessions.Range(func(key, value interface{}) bool {
			ctx, ok := value.(*SecurityContext)
			if ok && now.After(ctx.ExpiresAt) {
				sm.sessions.Delete(key)
			}
			return true
		})
	}
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsValidIP checks if an IP address is valid
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// SanitizeUserAgent sanitizes user agent string
func SanitizeUserAgent(userAgent string) string {
	// Remove potentially dangerous characters
	sanitized := strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, userAgent)

	// Limit length
	if len(sanitized) > 500 {
		sanitized = sanitized[:500]
	}

	return sanitized
}
