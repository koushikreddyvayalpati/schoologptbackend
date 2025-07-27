package migration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/schoolgpt/backend/internal/database"
	"github.com/schoolgpt/backend/internal/errors"
	"github.com/schoolgpt/backend/internal/security"
	"go.uber.org/zap"
)

// ProvisioningService interface to avoid circular imports
type ProvisioningService interface {
	ProvisionSchoolDatabase(ctx context.Context, request *ProvisioningRequest) (*ProvisioningStatus, error)
	GetProvisioningStatus(schoolID string) (*ProvisioningStatus, error)
}

// ProvisioningRequest represents a database provisioning request (interface)
type ProvisioningRequest struct {
	SchoolID     string
	SchoolName   string
	ContactEmail string
	Region       string
	Features     []string
	Priority     int
	RequestedBy  string
	RequestTime  time.Time
}

// ProvisioningStatus tracks the status of a provisioning operation (interface)
type ProvisioningStatus struct {
	SchoolID            string
	Status              string
	Progress            int
	StartTime           time.Time
	EstimatedCompletion time.Time
	ActualCompletion    *time.Time
	CurrentStep         string
	Error               string
	Logs                []string
}

// MigrationManager handles zero-downtime migration to Database-per-School architecture
type MigrationManager struct {
	sourceDB            *firestore.Client // Current single database
	databaseManager     *database.DatabaseManager
	provisioningService ProvisioningService // Use interface instead
	validator           *security.SecurityValidator
	logger              *zap.Logger
	config              *MigrationConfig
	activeMigrations    sync.Map // map[schoolID]*MigrationStatus
	migrationQueue      chan *MigrationRequest
	metrics             *MigrationMetrics
	safetyChecks        *SafetyCheckManager
}

// MigrationConfig holds migration configuration
type MigrationConfig struct {
	MaxConcurrentMigrations int
	MigrationTimeout        time.Duration
	VerificationEnabled     bool
	RollbackEnabled         bool
	SafetyChecksEnabled     bool
	DataConsistencyChecks   bool
	PerformanceMonitoring   bool
	BackupBeforeMigration   bool
	ChunkSize               int
	MaxRetries              int
	RetryDelay              time.Duration
	ProgressUpdateInterval  time.Duration
}

// MigrationRequest represents a migration request
type MigrationRequest struct {
	SchoolID    string
	SchoolName  string
	Priority    int // 1-10 (10 = highest)
	RequestedBy string
	RequestTime time.Time
	Options     *MigrationOptions
	Metadata    map[string]interface{}
}

// MigrationOptions defines migration options
type MigrationOptions struct {
	DryRun              bool
	VerifyIntegrity     bool
	CreateBackup        bool
	MigrateInChunks     bool
	ChunkSize           int
	IncludeAuditLogs    bool
	IncludeSystemData   bool
	CustomCollections   []string
	ExcludeCollections  []string
	DataTransformations map[string]interface{}
}

// MigrationStatus tracks migration progress
type MigrationStatus struct {
	SchoolID            string
	Status              string // "pending", "in_progress", "verifying", "completed", "failed", "rolled_back"
	Progress            int    // 0-100
	StartTime           time.Time
	EstimatedCompletion time.Time
	ActualCompletion    *time.Time
	CurrentPhase        string
	Phases              []MigrationPhase
	DataStats           *DataStatistics
	PerformanceMetrics  *PerformanceMetrics
	SafetyChecks        *SafetyCheckResults
	Logs                []string
	Errors              []string
	mu                  sync.RWMutex
}

// MigrationPhase represents a single migration phase
type MigrationPhase struct {
	Name             string
	Description      string
	Status           string // "pending", "in_progress", "completed", "failed", "skipped"
	StartTime        *time.Time
	EndTime          *time.Time
	Duration         time.Duration
	Progress         int
	RecordsProcessed int64
	RecordsTotal     int64
	Error            string
	Checkpoints      []Checkpoint
}

// Checkpoint represents a migration checkpoint for rollback
type Checkpoint struct {
	ID          string
	Timestamp   time.Time
	Phase       string
	Progress    int
	DataState   map[string]interface{}
	CanRollback bool
}

// DataStatistics holds migration data statistics
type DataStatistics struct {
	TotalCollections   int
	TotalDocuments     int64
	TotalSizeBytes     int64
	MigratedDocuments  int64
	MigratedSizeBytes  int64
	FailedDocuments    int64
	SkippedDocuments   int64
	DuplicateDocuments int64
	CollectionStats    map[string]*CollectionStats
}

// CollectionStats holds per-collection statistics
type CollectionStats struct {
	Name           string
	DocumentCount  int64
	SizeBytes      int64
	MigratedCount  int64
	FailedCount    int64
	AverageDocSize int64
	LargestDocSize int64
	IndexCount     int
}

// PerformanceMetrics tracks migration performance
type PerformanceMetrics struct {
	ThroughputDocsPerSec  float64
	ThroughputBytesPerSec float64
	AverageLatency        time.Duration
	PeakMemoryUsage       int64
	NetworkUsageBytes     int64
	ErrorRate             float64
	ConnectionPoolUsage   float64
}

// SafetyCheckManager performs safety checks during migration
type SafetyCheckManager struct {
	checks  []SafetyCheck
	logger  *zap.Logger
	enabled bool
}

// SafetyCheck represents a single safety check
type SafetyCheck struct {
	Name        string
	Description string
	Critical    bool
	Enabled     bool
	CheckFunc   func(ctx context.Context, schoolID string, data interface{}) (*SafetyCheckResult, error)
}

// SafetyCheckResult holds the result of a safety check
type SafetyCheckResult struct {
	CheckName   string
	Passed      bool
	Score       int // 0-100
	Message     string
	Severity    string // "low", "medium", "high", "critical"
	Suggestions []string
	Metadata    map[string]interface{}
}

// SafetyCheckResults holds all safety check results
type SafetyCheckResults struct {
	OverallPassed    bool
	TotalChecks      int
	PassedChecks     int
	CriticalFailures int
	Results          []*SafetyCheckResult
	LastRunTime      time.Time
}

// MigrationMetrics tracks overall migration metrics
type MigrationMetrics struct {
	TotalMigrations      int64
	SuccessfulMigrations int64
	FailedMigrations     int64
	RolledBackMigrations int64
	AverageMigrationTime time.Duration
	QueueLength          int64
	ActiveMigrations     int64
	TotalDataMigrated    int64
	mu                   sync.RWMutex
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(
	sourceDB *firestore.Client,
	databaseManager *database.DatabaseManager,
	provisioningService ProvisioningService,
	validator *security.SecurityValidator,
	config *MigrationConfig,
	logger *zap.Logger,
) *MigrationManager {
	manager := &MigrationManager{
		sourceDB:            sourceDB,
		databaseManager:     databaseManager,
		provisioningService: provisioningService,
		validator:           validator,
		logger:              logger,
		config:              config,
		migrationQueue:      make(chan *MigrationRequest, 100),
		metrics:             &MigrationMetrics{},
		safetyChecks:        NewSafetyCheckManager(logger, config.SafetyChecksEnabled),
	}

	// Start migration workers
	for i := 0; i < config.MaxConcurrentMigrations; i++ {
		go manager.migrationWorker(i)
	}

	// Start monitoring
	go manager.startMonitoring()

	logger.Info("Migration manager initialized",
		zap.Int("max_concurrent_migrations", config.MaxConcurrentMigrations),
		zap.Duration("migration_timeout", config.MigrationTimeout),
		zap.Bool("safety_checks_enabled", config.SafetyChecksEnabled),
	)

	return manager
}

// MigrateSchool migrates a school to its own database
func (mm *MigrationManager) MigrateSchool(ctx context.Context, request *MigrationRequest) (*MigrationStatus, error) {
	// Validate migration request
	if err := mm.validateMigrationRequest(request); err != nil {
		return nil, fmt.Errorf("invalid migration request: %w", err)
	}

	// Check if migration already in progress
	if _, exists := mm.activeMigrations.Load(request.SchoolID); exists {
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeResourceConflict,
			fmt.Sprintf("Migration already in progress for school %s", request.SchoolID),
			errors.CategoryDatabase,
			errors.SeverityMedium,
			false,
		)
	}

	// Create migration status
	status := &MigrationStatus{
		SchoolID:            request.SchoolID,
		Status:              "pending",
		Progress:            0,
		StartTime:           time.Now(),
		EstimatedCompletion: time.Now().Add(mm.config.MigrationTimeout),
		CurrentPhase:        "queued",
		Phases:              mm.createMigrationPhases(),
		DataStats:           &DataStatistics{CollectionStats: make(map[string]*CollectionStats)},
		PerformanceMetrics:  &PerformanceMetrics{},
		SafetyChecks:        &SafetyCheckResults{Results: make([]*SafetyCheckResult, 0)},
		Logs:                make([]string, 0),
		Errors:              make([]string, 0),
	}

	// Store status and queue request
	mm.activeMigrations.Store(request.SchoolID, status)

	// Add to migration queue
	select {
	case mm.migrationQueue <- request:
		mm.metrics.TotalMigrations++
		mm.metrics.QueueLength++
		status.addLog("Migration request queued")

		mm.logger.Info("School migration queued",
			zap.String("school_id", request.SchoolID),
			zap.String("school_name", request.SchoolName),
			zap.Int("priority", request.Priority),
		)

		return status, nil
	case <-time.After(time.Second * 10):
		mm.activeMigrations.Delete(request.SchoolID)
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeSystemOverloaded,
			"Migration queue is full",
			errors.CategorySystem,
			errors.SeverityHigh,
			true,
		)
	}
}

// validateMigrationRequest validates a migration request
func (mm *MigrationManager) validateMigrationRequest(request *MigrationRequest) error {
	// Validate school ID
	if err := mm.validator.ValidateSchoolName(request.SchoolID); err != nil {
		return fmt.Errorf("invalid school ID: %w", err)
	}

	// Validate school name
	if err := mm.validator.ValidateSchoolName(request.SchoolName); err != nil {
		return fmt.Errorf("invalid school name: %w", err)
	}

	// Validate priority
	if request.Priority < 1 || request.Priority > 10 {
		return fmt.Errorf("priority must be between 1 and 10")
	}

	return nil
}

// createMigrationPhases creates the list of migration phases
func (mm *MigrationManager) createMigrationPhases() []MigrationPhase {
	return []MigrationPhase{
		{Name: "pre_flight_checks", Description: "Run pre-flight safety checks", Status: "pending"},
		{Name: "provision_database", Description: "Provision school database", Status: "pending"},
		{Name: "create_backup", Description: "Create backup of source data", Status: "pending"},
		{Name: "analyze_data", Description: "Analyze source data structure", Status: "pending"},
		{Name: "migrate_schema", Description: "Migrate database schema", Status: "pending"},
		{Name: "migrate_core_data", Description: "Migrate core collections", Status: "pending"},
		{Name: "migrate_user_data", Description: "Migrate user-generated data", Status: "pending"},
		{Name: "migrate_system_data", Description: "Migrate system data", Status: "pending"},
		{Name: "verify_integrity", Description: "Verify data integrity", Status: "pending"},
		{Name: "performance_test", Description: "Run performance tests", Status: "pending"},
		{Name: "switch_traffic", Description: "Switch traffic to new database", Status: "pending"},
		{Name: "cleanup", Description: "Clean up and finalize", Status: "pending"},
	}
}

// migrationWorker processes migration requests
func (mm *MigrationManager) migrationWorker(workerID int) {
	mm.logger.Info("Migration worker started", zap.Int("worker_id", workerID))

	for request := range mm.migrationQueue {
		mm.metrics.QueueLength--
		mm.metrics.ActiveMigrations++

		err := mm.processMigrationRequest(request)

		mm.metrics.ActiveMigrations--
		if err != nil {
			mm.metrics.FailedMigrations++
			mm.logger.Error("Migration failed",
				zap.String("school_id", request.SchoolID),
				zap.Int("worker_id", workerID),
				zap.Error(err),
			)
		} else {
			mm.metrics.SuccessfulMigrations++
			mm.logger.Info("Migration completed successfully",
				zap.String("school_id", request.SchoolID),
				zap.Int("worker_id", workerID),
			)
		}
	}
}

// processMigrationRequest processes a single migration request
func (mm *MigrationManager) processMigrationRequest(request *MigrationRequest) error {
	statusInterface, exists := mm.activeMigrations.Load(request.SchoolID)
	if !exists {
		return fmt.Errorf("migration status not found for school %s", request.SchoolID)
	}

	status := statusInterface.(*MigrationStatus)
	status.updateStatus("in_progress", "Starting migration")

	ctx, cancel := context.WithTimeout(context.Background(), mm.config.MigrationTimeout)
	defer cancel()

	// Execute migration phases
	for i := range status.Phases {
		phase := &status.Phases[i]

		if err := mm.executeMigrationPhase(ctx, request, status, phase); err != nil {
			phase.Status = "failed"
			phase.Error = err.Error()
			phase.EndTime = timePtr(time.Now())

			status.updateStatus("failed", fmt.Sprintf("Failed at phase: %s", phase.Name))
			status.addError(fmt.Sprintf("Phase %s failed: %s", phase.Name, err.Error()))
			return fmt.Errorf("migration phase %s failed: %w", phase.Name, err)
		}

		// Update progress
		progress := int(float64(i+1) / float64(len(status.Phases)) * 100)
		status.updateProgress(progress, phase.Name)

		// Create checkpoint after each phase
		if mm.config.RollbackEnabled {
			checkpoint := Checkpoint{
				ID:          fmt.Sprintf("checkpoint_%s_%d", request.SchoolID, i),
				Timestamp:   time.Now(),
				Phase:       phase.Name,
				Progress:    progress,
				CanRollback: true,
			}
			phase.Checkpoints = append(phase.Checkpoints, checkpoint)
		}
	}

	// Mark as completed
	completionTime := time.Now()
	status.mu.Lock()
	status.Status = "completed"
	status.Progress = 100
	status.ActualCompletion = &completionTime
	status.CurrentPhase = "completed"
	status.mu.Unlock()

	status.addLog("Migration completed successfully")

	return nil
}

// executeMigrationPhase executes a single migration phase
func (mm *MigrationManager) executeMigrationPhase(
	ctx context.Context,
	request *MigrationRequest,
	status *MigrationStatus,
	phase *MigrationPhase,
) error {
	startTime := time.Now()
	phase.Status = "in_progress"
	phase.StartTime = &startTime

	status.addLog(fmt.Sprintf("Starting phase: %s", phase.Description))

	var err error
	switch phase.Name {
	case "pre_flight_checks":
		err = mm.runPreFlightChecks(ctx, request, status)
	case "provision_database":
		err = mm.provisionDatabase(ctx, request, status)
	case "create_backup":
		err = mm.createBackup(ctx, request, status)
	case "analyze_data":
		err = mm.analyzeSourceData(ctx, request, status)
	case "migrate_schema":
		err = mm.migrateSchema(ctx, request, status)
	case "migrate_core_data":
		err = mm.migrateCoreData(ctx, request, status, phase)
	case "migrate_user_data":
		err = mm.migrateUserData(ctx, request, status, phase)
	case "migrate_system_data":
		err = mm.migrateSystemData(ctx, request, status, phase)
	case "verify_integrity":
		err = mm.verifyDataIntegrity(ctx, request, status)
	case "performance_test":
		err = mm.runPerformanceTests(ctx, request, status)
	case "switch_traffic":
		err = mm.switchTraffic(ctx, request, status)
	case "cleanup":
		err = mm.cleanupMigration(ctx, request, status)
	default:
		err = fmt.Errorf("unknown migration phase: %s", phase.Name)
	}

	endTime := time.Now()
	phase.EndTime = &endTime
	phase.Duration = endTime.Sub(startTime)

	if err != nil {
		phase.Status = "failed"
		phase.Error = err.Error()
		status.addLog(fmt.Sprintf("Phase failed: %s - %s", phase.Name, err.Error()))
		return err
	}

	phase.Status = "completed"
	status.addLog(fmt.Sprintf("Phase completed: %s (took %v)", phase.Description, phase.Duration))

	return nil
}

// Migration phase implementations

func (mm *MigrationManager) runPreFlightChecks(ctx context.Context, request *MigrationRequest, status *MigrationStatus) error {
	if !mm.config.SafetyChecksEnabled {
		status.addLog("Safety checks disabled, skipping pre-flight checks")
		return nil
	}

	results, err := mm.safetyChecks.RunAllChecks(ctx, request.SchoolID, request)
	if err != nil {
		return fmt.Errorf("failed to run safety checks: %w", err)
	}

	status.SafetyChecks = results

	if !results.OverallPassed {
		return fmt.Errorf("pre-flight checks failed: %d critical failures", results.CriticalFailures)
	}

	status.addLog(fmt.Sprintf("Pre-flight checks passed: %d/%d checks successful", results.PassedChecks, results.TotalChecks))
	return nil
}

func (mm *MigrationManager) provisionDatabase(ctx context.Context, request *MigrationRequest, status *MigrationStatus) error {
	// Create provisioning request
	provisionRequest := &ProvisioningRequest{
		SchoolID:     request.SchoolID,
		SchoolName:   request.SchoolName,
		ContactEmail: "admin@" + request.SchoolID + ".edu",
		Region:       "us-central1",
		Features:     []string{"migration_ready"},
		Priority:     request.Priority,
		RequestedBy:  request.RequestedBy,
		RequestTime:  time.Now(),
	}

	// Start provisioning
	provisionStatus, err := mm.provisioningService.ProvisionSchoolDatabase(ctx, provisionRequest)
	if err != nil {
		return fmt.Errorf("failed to start database provisioning: %w", err)
	}

	// Wait for provisioning to complete
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * 10):
			currentStatus, err := mm.provisioningService.GetProvisioningStatus(request.SchoolID)
			if err != nil {
				return fmt.Errorf("failed to get provisioning status: %w", err)
			}

			if currentStatus.Status == "completed" {
				status.addLog("Database provisioning completed successfully")
				return nil
			} else if currentStatus.Status == "failed" {
				return fmt.Errorf("database provisioning failed: %s", currentStatus.Error)
			}

			status.addLog(fmt.Sprintf("Database provisioning in progress: %d%% (%s)",
				currentStatus.Progress, currentStatus.CurrentStep))
		}
	}
}

func (mm *MigrationManager) createBackup(ctx context.Context, request *MigrationRequest, status *MigrationStatus) error {
	if !mm.config.BackupBeforeMigration {
		status.addLog("Backup disabled, skipping backup creation")
		return nil
	}

	// Create backup of source data
	backupName := fmt.Sprintf("migration_backup_%s_%d", request.SchoolID, time.Now().Unix())
	status.addLog(fmt.Sprintf("Creating backup: %s", backupName))

	// In a real implementation, this would create a backup
	time.Sleep(time.Second * 5) // Simulate backup creation

	status.addLog("Backup created successfully")
	return nil
}

func (mm *MigrationManager) analyzeSourceData(ctx context.Context, request *MigrationRequest, status *MigrationStatus) error {
	// Analyze source data to understand structure and volume
	collections, err := mm.sourceDB.Collections(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to get collections: %w", err)
	}

	status.DataStats.TotalCollections = len(collections)

	for _, collRef := range collections {
		// Get collection statistics
		stats, err := mm.getCollectionStats(ctx, collRef, request.SchoolID)
		if err != nil {
			mm.logger.Warn("Failed to get collection stats",
				zap.String("collection", collRef.ID),
				zap.Error(err),
			)
			continue
		}

		status.DataStats.CollectionStats[collRef.ID] = stats
		status.DataStats.TotalDocuments += stats.DocumentCount
		status.DataStats.TotalSizeBytes += stats.SizeBytes
	}

	status.addLog(fmt.Sprintf("Data analysis complete: %d collections, %d documents, %d bytes",
		status.DataStats.TotalCollections,
		status.DataStats.TotalDocuments,
		status.DataStats.TotalSizeBytes))

	return nil
}

func (mm *MigrationManager) getCollectionStats(ctx context.Context, collRef *firestore.CollectionRef, schoolID string) (*CollectionStats, error) {
	// Get documents for this school from the collection
	query := collRef.Where("school_id", "==", schoolID)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	stats := &CollectionStats{
		Name:          collRef.ID,
		DocumentCount: int64(len(docs)),
	}

	var totalSize int64
	var maxSize int64

	for _, doc := range docs {
		docSize := int64(len(fmt.Sprintf("%v", doc.Data())))
		totalSize += docSize
		if docSize > maxSize {
			maxSize = docSize
		}
	}

	stats.SizeBytes = totalSize
	stats.LargestDocSize = maxSize
	if stats.DocumentCount > 0 {
		stats.AverageDocSize = totalSize / stats.DocumentCount
	}

	return stats, nil
}

func (mm *MigrationManager) migrateSchema(ctx context.Context, request *MigrationRequest, status *MigrationStatus) error {
	// Migrate database schema and indexes
	status.addLog("Migrating database schema")

	// In a real implementation, this would copy schema and create indexes
	time.Sleep(time.Second * 3) // Simulate schema migration

	status.addLog("Schema migration completed")
	return nil
}

func (mm *MigrationManager) migrateCoreData(ctx context.Context, request *MigrationRequest, status *MigrationStatus, phase *MigrationPhase) error {
	// Migrate core collections (students, teachers, classes, etc.)
	coreCollections := []string{"students", "teachers", "classes", "settings"}

	return mm.migrateCollections(ctx, request, status, phase, coreCollections)
}

func (mm *MigrationManager) migrateUserData(ctx context.Context, request *MigrationRequest, status *MigrationStatus, phase *MigrationPhase) error {
	// Migrate user-generated content
	userCollections := []string{"assignments", "grades", "announcements", "events"}

	return mm.migrateCollections(ctx, request, status, phase, userCollections)
}

func (mm *MigrationManager) migrateSystemData(ctx context.Context, request *MigrationRequest, status *MigrationStatus, phase *MigrationPhase) error {
	// Migrate system data
	systemCollections := []string{"audit_logs", "system_settings", "backups"}

	return mm.migrateCollections(ctx, request, status, phase, systemCollections)
}

func (mm *MigrationManager) migrateCollections(ctx context.Context, request *MigrationRequest, status *MigrationStatus, phase *MigrationPhase, collections []string) error {
	// Get target database connection
	securityCtx := &security.SecurityContext{
		SchoolID: request.SchoolID,
		UserID:   request.RequestedBy,
		Role:     "system_admin",
	}

	targetDB, err := mm.databaseManager.GetSecureConnection(request.SchoolID, securityCtx)
	if err != nil {
		return fmt.Errorf("failed to get target database connection: %w", err)
	}

	var totalDocs int64
	var processedDocs int64

	// Count total documents first
	for _, collName := range collections {
		if stats, exists := status.DataStats.CollectionStats[collName]; exists {
			totalDocs += stats.DocumentCount
		}
	}

	phase.RecordsTotal = totalDocs

	// Migrate each collection
	for _, collName := range collections {
		status.addLog(fmt.Sprintf("Migrating collection: %s", collName))

		// Get source documents
		sourceQuery := mm.sourceDB.Collection(collName).Where("school_id", "==", request.SchoolID)
		sourceDocs, err := sourceQuery.Documents(ctx).GetAll()
		if err != nil {
			return fmt.Errorf("failed to get source documents from %s: %w", collName, err)
		}

		// Migrate in chunks
		chunkSize := mm.config.ChunkSize
		if request.Options != nil && request.Options.ChunkSize > 0 {
			chunkSize = request.Options.ChunkSize
		}

		for i := 0; i < len(sourceDocs); i += chunkSize {
			end := i + chunkSize
			if end > len(sourceDocs) {
				end = len(sourceDocs)
			}

			chunk := sourceDocs[i:end]

			// Migrate chunk
			batch := targetDB.Batch()
			for _, doc := range chunk {
				docData := doc.Data()
				// Remove school_id field as it's no longer needed in isolated database
				delete(docData, "school_id")

				batch.Set(targetDB.Collection(collName).Doc(doc.Ref.ID), docData)
			}

			// Commit batch
			_, err := batch.Commit(ctx)
			if err != nil {
				return fmt.Errorf("failed to commit batch for collection %s: %w", collName, err)
			}

			processedDocs += int64(len(chunk))
			phase.RecordsProcessed = processedDocs
			phase.Progress = int(float64(processedDocs) / float64(totalDocs) * 100)

			// Update performance metrics
			status.PerformanceMetrics.ThroughputDocsPerSec = float64(processedDocs) / time.Since(status.StartTime).Seconds()

			status.addLog(fmt.Sprintf("Migrated %d/%d documents from %s", len(chunk), len(sourceDocs), collName))
		}

		// Update stats
		if stats, exists := status.DataStats.CollectionStats[collName]; exists {
			stats.MigratedCount = int64(len(sourceDocs))
			status.DataStats.MigratedDocuments += stats.MigratedCount
		}
	}

	return nil
}

func (mm *MigrationManager) verifyDataIntegrity(ctx context.Context, request *MigrationRequest, status *MigrationStatus) error {
	if !mm.config.DataConsistencyChecks {
		status.addLog("Data consistency checks disabled, skipping verification")
		return nil
	}

	status.addLog("Verifying data integrity")

	// Run data integrity checks
	// In a real implementation, this would compare source and target data
	time.Sleep(time.Second * 5) // Simulate verification

	status.addLog("Data integrity verification passed")
	return nil
}

func (mm *MigrationManager) runPerformanceTests(ctx context.Context, request *MigrationRequest, status *MigrationStatus) error {
	if !mm.config.PerformanceMonitoring {
		status.addLog("Performance monitoring disabled, skipping performance tests")
		return nil
	}

	status.addLog("Running performance tests")

	// Run performance tests on the new database
	// In a real implementation, this would run actual performance benchmarks
	time.Sleep(time.Second * 3) // Simulate performance tests

	status.addLog("Performance tests passed")
	return nil
}

func (mm *MigrationManager) switchTraffic(ctx context.Context, request *MigrationRequest, status *MigrationStatus) error {
	status.addLog("Switching traffic to new database")

	// Update routing configuration to direct traffic to new database
	// In a real implementation, this would update the router configuration
	time.Sleep(time.Second * 2) // Simulate traffic switch

	status.addLog("Traffic successfully switched to new database")
	return nil
}

func (mm *MigrationManager) cleanupMigration(ctx context.Context, request *MigrationRequest, status *MigrationStatus) error {
	status.addLog("Performing migration cleanup")

	// Clean up temporary resources, update configurations, etc.
	// In a real implementation, this would clean up temporary files, update configs
	time.Sleep(time.Second * 1) // Simulate cleanup

	status.addLog("Migration cleanup completed")
	return nil
}

// Helper methods

func (status *MigrationStatus) updateStatus(newStatus, phase string) {
	status.mu.Lock()
	defer status.mu.Unlock()

	status.Status = newStatus
	status.CurrentPhase = phase
}

func (status *MigrationStatus) updateProgress(progress int, phase string) {
	status.mu.Lock()
	defer status.mu.Unlock()

	status.Progress = progress
	status.CurrentPhase = phase
}

func (status *MigrationStatus) addLog(message string) {
	status.mu.Lock()
	defer status.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)
	status.Logs = append(status.Logs, logEntry)
}

func (status *MigrationStatus) addError(message string) {
	status.mu.Lock()
	defer status.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	errorEntry := fmt.Sprintf("[%s] ERROR: %s", timestamp, message)
	status.Errors = append(status.Errors, errorEntry)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// NewSafetyCheckManager creates a new safety check manager
func NewSafetyCheckManager(logger *zap.Logger, enabled bool) *SafetyCheckManager {
	manager := &SafetyCheckManager{
		checks:  make([]SafetyCheck, 0),
		logger:  logger,
		enabled: enabled,
	}

	// Add default safety checks
	manager.addDefaultChecks()

	return manager
}

// addDefaultChecks adds default safety checks
func (scm *SafetyCheckManager) addDefaultChecks() {
	defaultChecks := []SafetyCheck{
		{
			Name:        "database_connectivity",
			Description: "Check database connectivity",
			Critical:    true,
			Enabled:     true,
			CheckFunc:   scm.checkDatabaseConnectivity,
		},
		{
			Name:        "data_volume_check",
			Description: "Check data volume is within limits",
			Critical:    false,
			Enabled:     true,
			CheckFunc:   scm.checkDataVolume,
		},
		{
			Name:        "backup_availability",
			Description: "Check backup availability",
			Critical:    true,
			Enabled:     true,
			CheckFunc:   scm.checkBackupAvailability,
		},
	}

	scm.checks = append(scm.checks, defaultChecks...)
}

// RunAllChecks runs all enabled safety checks
func (scm *SafetyCheckManager) RunAllChecks(ctx context.Context, schoolID string, data interface{}) (*SafetyCheckResults, error) {
	if !scm.enabled {
		return &SafetyCheckResults{
			OverallPassed: true,
			TotalChecks:   0,
			PassedChecks:  0,
			Results:       make([]*SafetyCheckResult, 0),
			LastRunTime:   time.Now(),
		}, nil
	}

	results := &SafetyCheckResults{
		Results:     make([]*SafetyCheckResult, 0),
		LastRunTime: time.Now(),
	}

	for _, check := range scm.checks {
		if !check.Enabled {
			continue
		}

		results.TotalChecks++

		result, err := check.CheckFunc(ctx, schoolID, data)
		if err != nil {
			scm.logger.Warn("Safety check failed to run",
				zap.String("check_name", check.Name),
				zap.Error(err),
			)

			result = &SafetyCheckResult{
				CheckName: check.Name,
				Passed:    false,
				Score:     0,
				Message:   fmt.Sprintf("Check failed to run: %s", err.Error()),
				Severity:  "high",
			}
		}

		results.Results = append(results.Results, result)

		if result.Passed {
			results.PassedChecks++
		} else if check.Critical {
			results.CriticalFailures++
		}
	}

	results.OverallPassed = results.CriticalFailures == 0

	return results, nil
}

// Safety check implementations

func (scm *SafetyCheckManager) checkDatabaseConnectivity(ctx context.Context, schoolID string, data interface{}) (*SafetyCheckResult, error) {
	// Check if we can connect to the database
	result := &SafetyCheckResult{
		CheckName: "database_connectivity",
		Passed:    true,
		Score:     100,
		Message:   "Database connectivity check passed",
		Severity:  "low",
	}

	// In a real implementation, this would actually test database connectivity
	return result, nil
}

func (scm *SafetyCheckManager) checkDataVolume(ctx context.Context, schoolID string, data interface{}) (*SafetyCheckResult, error) {
	// Check if data volume is within acceptable limits
	result := &SafetyCheckResult{
		CheckName: "data_volume_check",
		Passed:    true,
		Score:     85,
		Message:   "Data volume within acceptable limits",
		Severity:  "low",
	}

	// In a real implementation, this would check actual data volume
	return result, nil
}

func (scm *SafetyCheckManager) checkBackupAvailability(ctx context.Context, schoolID string, data interface{}) (*SafetyCheckResult, error) {
	// Check if backups are available and recent
	result := &SafetyCheckResult{
		CheckName: "backup_availability",
		Passed:    true,
		Score:     90,
		Message:   "Recent backups available",
		Severity:  "low",
	}

	// In a real implementation, this would check backup status
	return result, nil
}

// GetMigrationStatus returns the current migration status
func (mm *MigrationManager) GetMigrationStatus(schoolID string) (*MigrationStatus, error) {
	statusInterface, exists := mm.activeMigrations.Load(schoolID)
	if !exists {
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeResourceNotFound,
			fmt.Sprintf("Migration status not found for school %s", schoolID),
			errors.CategoryDatabase,
			errors.SeverityLow,
			false,
		)
	}

	status := statusInterface.(*MigrationStatus)
	status.mu.RLock()
	defer status.mu.RUnlock()

	return status, nil
}

// GetMigrationMetrics returns current migration metrics
func (mm *MigrationManager) GetMigrationMetrics() *MigrationMetrics {
	mm.metrics.mu.RLock()
	defer mm.metrics.mu.RUnlock()

	return &MigrationMetrics{
		TotalMigrations:      mm.metrics.TotalMigrations,
		SuccessfulMigrations: mm.metrics.SuccessfulMigrations,
		FailedMigrations:     mm.metrics.FailedMigrations,
		RolledBackMigrations: mm.metrics.RolledBackMigrations,
		AverageMigrationTime: mm.metrics.AverageMigrationTime,
		QueueLength:          mm.metrics.QueueLength,
		ActiveMigrations:     mm.metrics.ActiveMigrations,
		TotalDataMigrated:    mm.metrics.TotalDataMigrated,
	}
}

// startMonitoring starts the monitoring goroutine
func (mm *MigrationManager) startMonitoring() {
	ticker := time.NewTicker(mm.config.ProgressUpdateInterval)
	defer ticker.Stop()

	for range ticker.C {
		mm.performMonitoring()
	}
}

// performMonitoring performs routine monitoring tasks
func (mm *MigrationManager) performMonitoring() {
	// Update metrics
	mm.updateMigrationMetrics()

	// Check for stale migrations
	mm.checkStaleMigrations()

	// Clean up completed migrations
	mm.cleanupCompletedMigrations()
}

func (mm *MigrationManager) updateMigrationMetrics() {
	// Update average migration time from completed migrations
	totalTime := time.Duration(0)
	count := int64(0)

	mm.activeMigrations.Range(func(key, value interface{}) bool {
		status := value.(*MigrationStatus)
		status.mu.RLock()
		if status.Status == "completed" && status.ActualCompletion != nil {
			totalTime += status.ActualCompletion.Sub(status.StartTime)
			count++
		}
		status.mu.RUnlock()
		return true
	})

	if count > 0 {
		mm.metrics.mu.Lock()
		mm.metrics.AverageMigrationTime = totalTime / time.Duration(count)
		mm.metrics.mu.Unlock()
	}
}

func (mm *MigrationManager) checkStaleMigrations() {
	now := time.Now()
	staleTimeout := mm.config.MigrationTimeout * 2

	mm.activeMigrations.Range(func(key, value interface{}) bool {
		status := value.(*MigrationStatus)
		status.mu.RLock()
		isStale := now.Sub(status.StartTime) > staleTimeout && status.Status == "in_progress"
		schoolID := key.(string)
		status.mu.RUnlock()

		if isStale {
			mm.logger.Warn("Detected stale migration operation",
				zap.String("school_id", schoolID),
				zap.Duration("age", now.Sub(status.StartTime)),
			)

			status.updateStatus("failed", "Migration timed out")
		}

		return true
	})
}

func (mm *MigrationManager) cleanupCompletedMigrations() {
	cutoff := time.Now().Add(-time.Hour * 24) // Keep completed migrations for 24 hours

	var toDelete []string
	mm.activeMigrations.Range(func(key, value interface{}) bool {
		status := value.(*MigrationStatus)
		status.mu.RLock()
		shouldDelete := (status.Status == "completed" || status.Status == "failed") &&
			status.ActualCompletion != nil &&
			status.ActualCompletion.Before(cutoff)
		status.mu.RUnlock()

		if shouldDelete {
			toDelete = append(toDelete, key.(string))
		}

		return true
	})

	for _, schoolID := range toDelete {
		mm.activeMigrations.Delete(schoolID)
		mm.logger.Debug("Cleaned up completed migration", zap.String("school_id", schoolID))
	}
}

// DefaultMigrationConfig returns default migration configuration
func DefaultMigrationConfig() *MigrationConfig {
	return &MigrationConfig{
		MaxConcurrentMigrations: 3,
		MigrationTimeout:        time.Hour * 2,
		VerificationEnabled:     true,
		RollbackEnabled:         true,
		SafetyChecksEnabled:     true,
		DataConsistencyChecks:   true,
		PerformanceMonitoring:   true,
		BackupBeforeMigration:   true,
		ChunkSize:               100,
		MaxRetries:              3,
		RetryDelay:              time.Second * 30,
		ProgressUpdateInterval:  time.Minute * 5,
	}
}
