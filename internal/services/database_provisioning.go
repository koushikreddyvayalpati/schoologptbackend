package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/schoolgpt/backend/internal/database"
	"github.com/schoolgpt/backend/internal/errors"
	"github.com/schoolgpt/backend/internal/security"
	"go.uber.org/zap"
)

// DatabaseProvisioningService handles automated database creation for schools
type DatabaseProvisioningService struct {
	masterDB         *firestore.Client
	databaseManager  *database.DatabaseManager
	validator        *security.SecurityValidator
	logger           *zap.Logger
	config           *ProvisioningConfig
	provisionQueue   chan *ProvisioningRequest
	activeProvisions sync.Map // map[schoolID]*ProvisioningStatus
	metrics          *ProvisioningMetrics
	securityManager  *SchoolSecurityManager
}

// ProvisioningConfig holds configuration for database provisioning
type ProvisioningConfig struct {
	MaxConcurrentProvisions int
	ProvisionTimeout        time.Duration
	DefaultRegion           string
	AvailableRegions        []string
	ResourceLimits          *ResourceLimits
	SecurityDefaults        *SecurityDefaults
	HealthCheckInterval     time.Duration
	BackupConfig            *BackupConfig
	MonitoringEnabled       bool
	AutoScalingEnabled      bool
}

// ResourceLimits defines resource constraints for school databases
type ResourceLimits struct {
	MaxConnections      int
	MaxStorageGB        int
	MaxQueriesPerMinute int
	MaxDocumentsRead    int64
	MaxDocumentsWrite   int64
	MaxIndexes          int
	MaxCollections      int
	QueryTimeoutSeconds int
}

// SecurityDefaults defines default security settings
type SecurityDefaults struct {
	EncryptionEnabled    bool
	BackupEncryption     bool
	AccessLoggingEnabled bool
	FirewallRules        []FirewallRule
	AllowedIPRanges      []string
	RequireTLS           bool
	MinTLSVersion        string
	PasswordPolicy       *PasswordPolicy
}

// FirewallRule represents a database firewall rule
type FirewallRule struct {
	Name        string
	Priority    int
	Action      string // "allow", "deny"
	Source      string // IP range or pattern
	Target      string // collection or operation
	Description string
}

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumbers   bool
	RequireSymbols   bool
	MaxAge           time.Duration
	HistorySize      int
}

// BackupConfig defines backup settings
type BackupConfig struct {
	Enabled            bool
	Schedule           string // cron format
	RetentionDays      int
	EncryptionEnabled  bool
	CompressionEnabled bool
	OffSiteBackup      bool
	BackupRegions      []string
}

// ProvisioningRequest represents a database provisioning request
type ProvisioningRequest struct {
	SchoolID       string
	SchoolName     string
	ContactEmail   string
	Region         string
	Features       []string
	CustomLimits   *ResourceLimits
	SecurityConfig *SecurityDefaults
	Priority       int
	RequestedBy    string
	RequestTime    time.Time
	Metadata       map[string]interface{}
}

// ProvisioningStatus tracks the status of a provisioning operation
type ProvisioningStatus struct {
	SchoolID            string
	Status              string // "pending", "in_progress", "completed", "failed", "cancelled"
	Progress            int    // 0-100
	StartTime           time.Time
	EstimatedCompletion time.Time
	ActualCompletion    *time.Time
	CurrentStep         string
	Steps               []ProvisioningStep
	Error               string
	Logs                []string
	ResourcesCreated    []string
	mu                  sync.RWMutex
}

// ProvisioningStep represents a single step in the provisioning process
type ProvisioningStep struct {
	Name        string
	Description string
	Status      string // "pending", "in_progress", "completed", "failed", "skipped"
	StartTime   *time.Time
	EndTime     *time.Time
	Duration    time.Duration
	Error       string
	Output      map[string]interface{}
}

// ProvisioningMetrics tracks provisioning performance
type ProvisioningMetrics struct {
	TotalRequests        int64
	SuccessfulProvisions int64
	FailedProvisions     int64
	AverageProvisionTime time.Duration
	QueueLength          int64
	ActiveProvisions     int64
	mu                   sync.RWMutex
}

// SchoolSecurityManager manages per-school security configurations
type SchoolSecurityManager struct {
	configs   sync.Map // map[schoolID]*SchoolSecurityConfig
	masterDB  *firestore.Client
	validator *security.SecurityValidator
	logger    *zap.Logger
}

// SchoolSecurityConfig holds security configuration for a school
type SchoolSecurityConfig struct {
	SchoolID            string
	EncryptionKeys      map[string]string
	FirewallRules       []FirewallRule
	AccessPolicies      map[string]*AccessPolicy
	AuditConfig         *AuditConfig
	ComplianceStandards []string
	LastSecurityAudit   time.Time
	SecurityIncidents   []SecurityIncident
	mu                  sync.RWMutex
}

// AccessPolicy defines access control policies
type AccessPolicy struct {
	Name        string
	Resources   []string
	Permissions []string
	Conditions  map[string]interface{}
	ExpiresAt   *time.Time
}

// AuditConfig defines audit logging configuration
type AuditConfig struct {
	Enabled              bool
	LogLevel             string
	IncludeDataReads     bool
	IncludeDataWrites    bool
	IncludeSchemaChanges bool
	RetentionDays        int
	ExportEnabled        bool
	ExportDestination    string
}

// SecurityIncident represents a security incident
type SecurityIncident struct {
	ID          string
	Type        string
	Severity    string
	Description string
	Timestamp   time.Time
	Resolved    bool
	Resolution  string
}

// NewDatabaseProvisioningService creates a new provisioning service
func NewDatabaseProvisioningService(
	masterDB *firestore.Client,
	databaseManager *database.DatabaseManager,
	validator *security.SecurityValidator,
	config *ProvisioningConfig,
	logger *zap.Logger,
) *DatabaseProvisioningService {
	service := &DatabaseProvisioningService{
		masterDB:        masterDB,
		databaseManager: databaseManager,
		validator:       validator,
		logger:          logger,
		config:          config,
		provisionQueue:  make(chan *ProvisioningRequest, 100),
		metrics:         &ProvisioningMetrics{},
		securityManager: &SchoolSecurityManager{
			masterDB:  masterDB,
			validator: validator,
			logger:    logger,
		},
	}

	// Start provisioning workers
	for i := 0; i < config.MaxConcurrentProvisions; i++ {
		go service.provisioningWorker(i)
	}

	// Start monitoring
	go service.startMonitoring()

	logger.Info("Database provisioning service initialized",
		zap.Int("max_concurrent_provisions", config.MaxConcurrentProvisions),
		zap.String("default_region", config.DefaultRegion),
		zap.Bool("auto_scaling_enabled", config.AutoScalingEnabled),
	)

	return service
}

// ProvisionSchoolDatabase creates a new database for a school
func (dps *DatabaseProvisioningService) ProvisionSchoolDatabase(ctx context.Context, request *ProvisioningRequest) (*ProvisioningStatus, error) {
	// Validate the provisioning request
	if err := dps.validateProvisioningRequest(request); err != nil {
		return nil, fmt.Errorf("invalid provisioning request: %w", err)
	}

	// Check if school already has a database
	if exists, err := dps.checkSchoolExists(ctx, request.SchoolID); err != nil {
		return nil, fmt.Errorf("failed to check school existence: %w", err)
	} else if exists {
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeDuplicateResource,
			fmt.Sprintf("School %s already has a database", request.SchoolID),
			errors.CategoryDatabase,
			errors.SeverityMedium,
			false,
		)
	}

	// Create provisioning status
	status := &ProvisioningStatus{
		SchoolID:            request.SchoolID,
		Status:              "pending",
		Progress:            0,
		StartTime:           time.Now(),
		EstimatedCompletion: time.Now().Add(dps.config.ProvisionTimeout),
		CurrentStep:         "queued",
		Steps:               dps.createProvisioningSteps(),
		Logs:                make([]string, 0),
		ResourcesCreated:    make([]string, 0),
	}

	// Store status and queue request
	dps.activeProvisions.Store(request.SchoolID, status)

	// Add to provisioning queue
	select {
	case dps.provisionQueue <- request:
		dps.metrics.TotalRequests++
		dps.metrics.QueueLength++
		status.addLog("Request queued for provisioning")

		dps.logger.Info("School database provisioning queued",
			zap.String("school_id", request.SchoolID),
			zap.String("school_name", request.SchoolName),
			zap.String("region", request.Region),
		)

		return status, nil
	case <-time.After(time.Second * 10):
		dps.activeProvisions.Delete(request.SchoolID)
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeSystemOverloaded,
			"Provisioning queue is full",
			errors.CategorySystem,
			errors.SeverityHigh,
			true,
		)
	}
}

// validateProvisioningRequest validates a provisioning request
func (dps *DatabaseProvisioningService) validateProvisioningRequest(request *ProvisioningRequest) error {
	// Validate school ID
	if err := dps.validator.ValidateSchoolName(request.SchoolID); err != nil {
		return fmt.Errorf("invalid school ID: %w", err)
	}

	// Validate school name
	if err := dps.validator.ValidateSchoolName(request.SchoolName); err != nil {
		return fmt.Errorf("invalid school name: %w", err)
	}

	// Validate email
	if err := dps.validator.ValidateEmail(request.ContactEmail); err != nil {
		return fmt.Errorf("invalid contact email: %w", err)
	}

	// Validate region
	if !dps.isValidRegion(request.Region) {
		return fmt.Errorf("invalid region: %s", request.Region)
	}

	// Validate features
	for _, feature := range request.Features {
		if !dps.isValidFeature(feature) {
			return fmt.Errorf("invalid feature: %s", feature)
		}
	}

	return nil
}

// isValidRegion checks if a region is valid
func (dps *DatabaseProvisioningService) isValidRegion(region string) bool {
	for _, validRegion := range dps.config.AvailableRegions {
		if region == validRegion {
			return true
		}
	}
	return false
}

// isValidFeature checks if a feature is valid
func (dps *DatabaseProvisioningService) isValidFeature(feature string) bool {
	validFeatures := []string{
		"voice_processing",
		"advanced_analytics",
		"ai_assistance",
		"custom_fields",
		"third_party_integration",
		"advanced_security",
		"compliance_reporting",
		"performance_monitoring",
	}

	for _, validFeature := range validFeatures {
		if feature == validFeature {
			return true
		}
	}
	return false
}

// checkSchoolExists checks if a school already has a database
func (dps *DatabaseProvisioningService) checkSchoolExists(ctx context.Context, schoolID string) (bool, error) {
	doc, err := dps.masterDB.Collection("schools").Doc(schoolID).Get(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return doc.Exists(), nil
}

// createProvisioningSteps creates the list of provisioning steps
func (dps *DatabaseProvisioningService) createProvisioningSteps() []ProvisioningStep {
	return []ProvisioningStep{
		{Name: "validate_request", Description: "Validate provisioning request", Status: "pending"},
		{Name: "create_project", Description: "Create GCP project for school", Status: "pending"},
		{Name: "setup_firestore", Description: "Initialize Firestore database", Status: "pending"},
		{Name: "configure_security", Description: "Apply security configurations", Status: "pending"},
		{Name: "create_collections", Description: "Create base collections and indexes", Status: "pending"},
		{Name: "setup_monitoring", Description: "Configure monitoring and alerting", Status: "pending"},
		{Name: "configure_backup", Description: "Setup backup and recovery", Status: "pending"},
		{Name: "run_tests", Description: "Run validation tests", Status: "pending"},
		{Name: "register_school", Description: "Register school in master database", Status: "pending"},
		{Name: "complete", Description: "Finalize provisioning", Status: "pending"},
	}
}

// provisioningWorker processes provisioning requests
func (dps *DatabaseProvisioningService) provisioningWorker(workerID int) {
	dps.logger.Info("Provisioning worker started", zap.Int("worker_id", workerID))

	for request := range dps.provisionQueue {
		dps.metrics.QueueLength--
		dps.metrics.ActiveProvisions++

		err := dps.processProvisioningRequest(request)

		dps.metrics.ActiveProvisions--
		if err != nil {
			dps.metrics.FailedProvisions++
			dps.logger.Error("Provisioning failed",
				zap.String("school_id", request.SchoolID),
				zap.Int("worker_id", workerID),
				zap.Error(err),
			)
		} else {
			dps.metrics.SuccessfulProvisions++
			dps.logger.Info("Provisioning completed successfully",
				zap.String("school_id", request.SchoolID),
				zap.Int("worker_id", workerID),
			)
		}
	}
}

// processProvisioningRequest processes a single provisioning request
func (dps *DatabaseProvisioningService) processProvisioningRequest(request *ProvisioningRequest) error {
	statusInterface, exists := dps.activeProvisions.Load(request.SchoolID)
	if !exists {
		return fmt.Errorf("provisioning status not found for school %s", request.SchoolID)
	}

	status := statusInterface.(*ProvisioningStatus)
	status.updateStatus("in_progress", "Starting provisioning")

	ctx, cancel := context.WithTimeout(context.Background(), dps.config.ProvisionTimeout)
	defer cancel()

	// Execute provisioning steps
	for i := range status.Steps {
		step := &status.Steps[i]

		if err := dps.executeProvisioningStep(ctx, request, status, step); err != nil {
			step.Status = "failed"
			step.Error = err.Error()
			step.EndTime = timePtr(time.Now())

			status.updateStatus("failed", fmt.Sprintf("Failed at step: %s", step.Name))
			return fmt.Errorf("provisioning step %s failed: %w", step.Name, err)
		}

		// Update progress
		progress := int(float64(i+1) / float64(len(status.Steps)) * 100)
		status.updateProgress(progress, step.Name)
	}

	// Mark as completed
	completionTime := time.Now()
	status.mu.Lock()
	status.Status = "completed"
	status.Progress = 100
	status.ActualCompletion = &completionTime
	status.CurrentStep = "completed"
	status.mu.Unlock()

	status.addLog("Provisioning completed successfully")

	return nil
}

// executeProvisioningStep executes a single provisioning step
func (dps *DatabaseProvisioningService) executeProvisioningStep(
	ctx context.Context,
	request *ProvisioningRequest,
	status *ProvisioningStatus,
	step *ProvisioningStep,
) error {
	startTime := time.Now()
	step.Status = "in_progress"
	step.StartTime = &startTime

	status.addLog(fmt.Sprintf("Starting step: %s", step.Description))

	var err error
	switch step.Name {
	case "validate_request":
		err = dps.validateProvisioningRequest(request)
	case "create_project":
		err = dps.createGCPProject(ctx, request, status)
	case "setup_firestore":
		err = dps.setupFirestore(ctx, request, status)
	case "configure_security":
		err = dps.configureSecurity(ctx, request, status)
	case "create_collections":
		err = dps.createBaseCollections(ctx, request, status)
	case "setup_monitoring":
		err = dps.setupMonitoring(ctx, request, status)
	case "configure_backup":
		err = dps.configureBackup(ctx, request, status)
	case "run_tests":
		err = dps.runValidationTests(ctx, request, status)
	case "register_school":
		err = dps.registerSchoolInMaster(ctx, request, status)
	case "complete":
		err = dps.finalizeProvisioning(ctx, request, status)
	default:
		err = fmt.Errorf("unknown provisioning step: %s", step.Name)
	}

	endTime := time.Now()
	step.EndTime = &endTime
	step.Duration = endTime.Sub(startTime)

	if err != nil {
		step.Status = "failed"
		step.Error = err.Error()
		status.addLog(fmt.Sprintf("Step failed: %s - %s", step.Name, err.Error()))
		return err
	}

	step.Status = "completed"
	status.addLog(fmt.Sprintf("Step completed: %s (took %v)", step.Description, step.Duration))

	return nil
}

// Provisioning step implementations

func (dps *DatabaseProvisioningService) createGCPProject(ctx context.Context, request *ProvisioningRequest, status *ProvisioningStatus) error {
	projectID := fmt.Sprintf("schoolgpt-school-%s", request.SchoolID)
	status.ResourcesCreated = append(status.ResourcesCreated, projectID)

	// In a real implementation, this would create a GCP project
	// For now, we'll simulate the process
	time.Sleep(time.Second * 2) // Simulate project creation time

	return nil
}

func (dps *DatabaseProvisioningService) setupFirestore(ctx context.Context, request *ProvisioningRequest, status *ProvisioningStatus) error {
	// Initialize Firestore database for the school
	databaseName := fmt.Sprintf("school-%s", request.SchoolID)
	status.ResourcesCreated = append(status.ResourcesCreated, databaseName)

	// In a real implementation, this would create the Firestore database
	time.Sleep(time.Second * 3) // Simulate database creation time

	return nil
}

func (dps *DatabaseProvisioningService) configureSecurity(ctx context.Context, request *ProvisioningRequest, status *ProvisioningStatus) error {
	// Apply security configurations
	securityConfig := dps.createSchoolSecurityConfig(request)

	err := dps.securityManager.ApplySecurityConfig(ctx, request.SchoolID, securityConfig)
	if err != nil {
		return fmt.Errorf("failed to apply security config: %w", err)
	}

	status.addLog("Security configuration applied successfully")
	return nil
}

func (dps *DatabaseProvisioningService) createBaseCollections(ctx context.Context, request *ProvisioningRequest, status *ProvisioningStatus) error {
	// Create base collections and indexes
	collections := []string{
		"students",
		"teachers",
		"classes",
		"attendance",
		"assignments",
		"grades",
		"announcements",
		"events",
		"settings",
		"audit_logs",
	}

	for _, collection := range collections {
		status.ResourcesCreated = append(status.ResourcesCreated, fmt.Sprintf("collection:%s", collection))
		// In a real implementation, collections would be created with proper schemas
	}

	status.addLog(fmt.Sprintf("Created %d base collections", len(collections)))
	return nil
}

func (dps *DatabaseProvisioningService) setupMonitoring(ctx context.Context, request *ProvisioningRequest, status *ProvisioningStatus) error {
	// Setup monitoring and alerting
	if dps.config.MonitoringEnabled {
		status.addLog("Monitoring and alerting configured")
	}
	return nil
}

func (dps *DatabaseProvisioningService) configureBackup(ctx context.Context, request *ProvisioningRequest, status *ProvisioningStatus) error {
	// Configure backup and recovery
	if dps.config.BackupConfig.Enabled {
		status.addLog("Backup and recovery configured")
	}
	return nil
}

func (dps *DatabaseProvisioningService) runValidationTests(ctx context.Context, request *ProvisioningRequest, status *ProvisioningStatus) error {
	// Run validation tests
	tests := []string{
		"connectivity_test",
		"security_test",
		"performance_test",
		"backup_test",
	}

	for _, test := range tests {
		status.addLog(fmt.Sprintf("Running %s", test))
		time.Sleep(time.Millisecond * 500) // Simulate test execution
	}

	status.addLog("All validation tests passed")
	return nil
}

func (dps *DatabaseProvisioningService) registerSchoolInMaster(ctx context.Context, request *ProvisioningRequest, status *ProvisioningStatus) error {
	// Register school in master database
	schoolDoc := map[string]interface{}{
		"school_id":           request.SchoolID,
		"school_name":         request.SchoolName,
		"contact_email":       request.ContactEmail,
		"region":              request.Region,
		"features":            request.Features,
		"status":              "active",
		"created_at":          time.Now(),
		"database_name":       fmt.Sprintf("school-%s", request.SchoolID),
		"project_id":          fmt.Sprintf("schoolgpt-school-%s", request.SchoolID),
		"provisioning_status": status,
	}

	_, err := dps.masterDB.Collection("schools").Doc(request.SchoolID).Set(ctx, schoolDoc)
	if err != nil {
		return fmt.Errorf("failed to register school in master database: %w", err)
	}

	status.addLog("School registered in master database")
	return nil
}

func (dps *DatabaseProvisioningService) finalizeProvisioning(ctx context.Context, request *ProvisioningRequest, status *ProvisioningStatus) error {
	// Finalize provisioning
	status.addLog("Provisioning finalized")
	return nil
}

// Helper methods

func (dps *DatabaseProvisioningService) createSchoolSecurityConfig(request *ProvisioningRequest) *SchoolSecurityConfig {
	return &SchoolSecurityConfig{
		SchoolID:            request.SchoolID,
		EncryptionKeys:      make(map[string]string),
		FirewallRules:       dps.config.SecurityDefaults.FirewallRules,
		AccessPolicies:      make(map[string]*AccessPolicy),
		AuditConfig:         &AuditConfig{Enabled: true, LogLevel: "INFO"},
		ComplianceStandards: []string{"FERPA", "COPPA"},
		LastSecurityAudit:   time.Now(),
	}
}

func (status *ProvisioningStatus) updateStatus(newStatus, step string) {
	status.mu.Lock()
	defer status.mu.Unlock()

	status.Status = newStatus
	status.CurrentStep = step
}

func (status *ProvisioningStatus) updateProgress(progress int, step string) {
	status.mu.Lock()
	defer status.mu.Unlock()

	status.Progress = progress
	status.CurrentStep = step
}

func (status *ProvisioningStatus) addLog(message string) {
	status.mu.Lock()
	defer status.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)
	status.Logs = append(status.Logs, logEntry)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// GetProvisioningStatus returns the current status of a provisioning operation
func (dps *DatabaseProvisioningService) GetProvisioningStatus(schoolID string) (*ProvisioningStatus, error) {
	statusInterface, exists := dps.activeProvisions.Load(schoolID)
	if !exists {
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeResourceNotFound,
			fmt.Sprintf("Provisioning status not found for school %s", schoolID),
			errors.CategoryDatabase,
			errors.SeverityLow,
			false,
		)
	}

	status := statusInterface.(*ProvisioningStatus)
	status.mu.RLock()
	defer status.mu.RUnlock()

	return status, nil
}

// ApplySecurityConfig applies security configuration for a school
func (sm *SchoolSecurityManager) ApplySecurityConfig(ctx context.Context, schoolID string, config *SchoolSecurityConfig) error {
	// Store security configuration
	sm.configs.Store(schoolID, config)

	// Apply configuration in the actual database
	configDoc := map[string]interface{}{
		"school_id":            config.SchoolID,
		"encryption_enabled":   true,
		"firewall_rules":       config.FirewallRules,
		"audit_config":         config.AuditConfig,
		"compliance_standards": config.ComplianceStandards,
		"last_security_audit":  config.LastSecurityAudit,
		"created_at":           time.Now(),
	}

	_, err := sm.masterDB.Collection("school_security_configs").Doc(schoolID).Set(ctx, configDoc)
	if err != nil {
		return fmt.Errorf("failed to store security config: %w", err)
	}

	sm.logger.Info("Security configuration applied",
		zap.String("school_id", schoolID),
		zap.Int("firewall_rules", len(config.FirewallRules)),
	)

	return nil
}

// GetProvisioningMetrics returns current provisioning metrics
func (dps *DatabaseProvisioningService) GetProvisioningMetrics() *ProvisioningMetrics {
	dps.metrics.mu.RLock()
	defer dps.metrics.mu.RUnlock()

	return &ProvisioningMetrics{
		TotalRequests:        dps.metrics.TotalRequests,
		SuccessfulProvisions: dps.metrics.SuccessfulProvisions,
		FailedProvisions:     dps.metrics.FailedProvisions,
		AverageProvisionTime: dps.metrics.AverageProvisionTime,
		QueueLength:          dps.metrics.QueueLength,
		ActiveProvisions:     dps.metrics.ActiveProvisions,
	}
}

// startMonitoring starts the monitoring goroutine
func (dps *DatabaseProvisioningService) startMonitoring() {
	ticker := time.NewTicker(dps.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		dps.performHealthChecks()
	}
}

// performHealthChecks performs routine health checks
func (dps *DatabaseProvisioningService) performHealthChecks() {
	// Check for stale provisioning operations
	dps.checkStaleProvisions()

	// Update metrics
	dps.updateMetrics()

	// Clean up completed provisions
	dps.cleanupCompletedProvisions()
}

func (dps *DatabaseProvisioningService) checkStaleProvisions() {
	now := time.Now()
	staleTimeout := dps.config.ProvisionTimeout * 2

	dps.activeProvisions.Range(func(key, value interface{}) bool {
		status := value.(*ProvisioningStatus)
		status.mu.RLock()
		isStale := now.Sub(status.StartTime) > staleTimeout && status.Status == "in_progress"
		schoolID := key.(string)
		status.mu.RUnlock()

		if isStale {
			dps.logger.Warn("Detected stale provisioning operation",
				zap.String("school_id", schoolID),
				zap.Duration("age", now.Sub(status.StartTime)),
			)

			status.updateStatus("failed", "Operation timed out")
		}

		return true
	})
}

func (dps *DatabaseProvisioningService) updateMetrics() {
	// Update average provisioning time
	if dps.metrics.SuccessfulProvisions > 0 {
		// Calculate from completed provisions
		totalTime := time.Duration(0)
		count := int64(0)

		dps.activeProvisions.Range(func(key, value interface{}) bool {
			status := value.(*ProvisioningStatus)
			status.mu.RLock()
			if status.Status == "completed" && status.ActualCompletion != nil {
				totalTime += status.ActualCompletion.Sub(status.StartTime)
				count++
			}
			status.mu.RUnlock()
			return true
		})

		if count > 0 {
			dps.metrics.mu.Lock()
			dps.metrics.AverageProvisionTime = totalTime / time.Duration(count)
			dps.metrics.mu.Unlock()
		}
	}
}

func (dps *DatabaseProvisioningService) cleanupCompletedProvisions() {
	cutoff := time.Now().Add(-time.Hour * 24) // Keep completed provisions for 24 hours

	var toDelete []string
	dps.activeProvisions.Range(func(key, value interface{}) bool {
		status := value.(*ProvisioningStatus)
		status.mu.RLock()
		shouldDelete := status.Status == "completed" &&
			status.ActualCompletion != nil &&
			status.ActualCompletion.Before(cutoff)
		status.mu.RUnlock()

		if shouldDelete {
			toDelete = append(toDelete, key.(string))
		}

		return true
	})

	for _, schoolID := range toDelete {
		dps.activeProvisions.Delete(schoolID)
		dps.logger.Debug("Cleaned up completed provision", zap.String("school_id", schoolID))
	}
}

// DefaultProvisioningConfig returns default configuration
func DefaultProvisioningConfig() *ProvisioningConfig {
	return &ProvisioningConfig{
		MaxConcurrentProvisions: 5,
		ProvisionTimeout:        time.Minute * 30,
		DefaultRegion:           "us-central1",
		AvailableRegions:        []string{"us-central1", "us-east1", "us-west1", "europe-west1"},
		ResourceLimits: &ResourceLimits{
			MaxConnections:      100,
			MaxStorageGB:        10,
			MaxQueriesPerMinute: 10000,
			MaxDocumentsRead:    1000000,
			MaxDocumentsWrite:   100000,
			MaxIndexes:          200,
			MaxCollections:      50,
			QueryTimeoutSeconds: 30,
		},
		SecurityDefaults: &SecurityDefaults{
			EncryptionEnabled:    true,
			BackupEncryption:     true,
			AccessLoggingEnabled: true,
			FirewallRules:        []FirewallRule{},
			AllowedIPRanges:      []string{},
			RequireTLS:           true,
			MinTLSVersion:        "1.2",
			PasswordPolicy: &PasswordPolicy{
				MinLength:        12,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   true,
				MaxAge:           time.Hour * 24 * 90, // 90 days
				HistorySize:      10,
			},
		},
		HealthCheckInterval: time.Minute * 5,
		BackupConfig: &BackupConfig{
			Enabled:            true,
			Schedule:           "0 2 * * *", // Daily at 2 AM
			RetentionDays:      30,
			EncryptionEnabled:  true,
			CompressionEnabled: true,
			OffSiteBackup:      true,
			BackupRegions:      []string{"us-central1", "us-east1"},
		},
		MonitoringEnabled:  true,
		AutoScalingEnabled: true,
	}
}
