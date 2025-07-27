package monitoring

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

// DatabaseMonitor provides comprehensive monitoring for school databases
type DatabaseMonitor struct {
	databaseManager *database.DatabaseManager
	router          *database.DatabaseRouter
	logger          *zap.Logger
	config          *MonitoringConfig
	metrics         *MonitoringMetrics
	collectors      map[string]*MetricCollector
	alertManager    *AlertManager
	mu              sync.RWMutex
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Enabled                bool
	CollectionInterval     time.Duration
	AlertingEnabled        bool
	PerformanceMonitoring  bool
	HealthCheckInterval    time.Duration
	MetricsRetentionDays   int
	AlertThresholds        *AlertThresholds
	OptimizationEnabled    bool
	RecommendationsEnabled bool
}

// AlertThresholds defines alert thresholds
type AlertThresholds struct {
	HighConnectionUsage     float64 // Percentage
	HighQueryLatency        time.Duration
	HighErrorRate           float64 // Percentage
	LowAvailability         float64 // Percentage
	HighDiskUsage           float64 // Percentage
	SlowQueryThreshold      time.Duration
	ConnectionPoolExhausted int
}

// MonitoringMetrics tracks overall monitoring metrics
type MonitoringMetrics struct {
	TotalSchools             int64
	ActiveConnections        int64
	TotalQueries             int64
	SuccessfulQueries        int64
	FailedQueries            int64
	AverageLatency           time.Duration
	PeakLatency              time.Duration
	AlertsTriggered          int64
	RecommendationsGenerated int64
	mu                       sync.RWMutex
}

// MetricCollector collects metrics for a specific school
type MetricCollector struct {
	SchoolID        string
	DatabaseName    string
	Metrics         *SchoolMetrics
	LastCollected   time.Time
	HealthStatus    string
	Alerts          []*Alert
	Recommendations []*Recommendation
	mu              sync.RWMutex
}

// SchoolMetrics holds metrics for a specific school database
type SchoolMetrics struct {
	ConnectionMetrics  *ConnectionMetrics
	QueryMetrics       *QueryMetrics
	PerformanceMetrics *PerformanceMetrics
	ResourceMetrics    *ResourceMetrics
	HealthMetrics      *HealthMetrics
	LastUpdated        time.Time
}

// ConnectionMetrics tracks database connections
type ConnectionMetrics struct {
	ActiveConnections   int64
	TotalConnections    int64
	MaxConnections      int64
	ConnectionPoolUsage float64
	IdleConnections     int64
	ConnectionErrors    int64
	AvgConnectionTime   time.Duration
	ConnectionTimeouts  int64
}

// QueryMetrics tracks query performance
type QueryMetrics struct {
	TotalQueries      int64
	SuccessfulQueries int64
	FailedQueries     int64
	ReadQueries       int64
	WriteQueries      int64
	AverageLatency    time.Duration
	P95Latency        time.Duration
	P99Latency        time.Duration
	SlowQueries       int64
	QueryErrors       int64
	QueryTimeouts     int64
}

// PerformanceMetrics tracks overall performance
type PerformanceMetrics struct {
	ThroughputQPS   float64
	ThroughputRPS   float64
	CPUUsage        float64
	MemoryUsage     float64
	DiskIOPS        float64
	NetworkBytesIn  int64
	NetworkBytesOut int64
	CacheHitRatio   float64
	IndexUsageRatio float64
}

// ResourceMetrics tracks resource usage
type ResourceMetrics struct {
	StorageUsedBytes    int64
	StorageTotalBytes   int64
	StorageUsagePercent float64
	DocumentCount       int64
	CollectionCount     int64
	IndexCount          int64
	BackupSizeBytes     int64
	BackupCount         int64
	LastBackupTime      time.Time
}

// HealthMetrics tracks database health
type HealthMetrics struct {
	AvailabilityPercent float64
	UptimeSeconds       int64
	LastDowntime        *time.Time
	HealthScore         int // 0-100
	ErrorRate           float64
	WarningCount        int64
	CriticalIssues      int64
	LastHealthCheck     time.Time
}

// Alert represents a monitoring alert
type Alert struct {
	ID             string
	SchoolID       string
	Type           string
	Severity       string // "low", "medium", "high", "critical"
	Title          string
	Message        string
	Metric         string
	Threshold      interface{}
	ActualValue    interface{}
	TriggeredAt    time.Time
	ResolvedAt     *time.Time
	AcknowledgedAt *time.Time
	Status         string // "active", "resolved", "acknowledged"
	Actions        []string
}

// Recommendation represents an optimization recommendation
type Recommendation struct {
	ID               string
	SchoolID         string
	Type             string // "performance", "cost", "security", "maintenance"
	Priority         string // "low", "medium", "high", "critical"
	Title            string
	Description      string
	Impact           string
	Effort           string // "low", "medium", "high"
	Category         string
	CreatedAt        time.Time
	ExpiresAt        *time.Time
	Applied          bool
	AppliedAt        *time.Time
	EstimatedSavings map[string]interface{}
}

// AlertManager manages alerts and notifications
type AlertManager struct {
	alerts          sync.Map // map[alertID]*Alert
	thresholds      *AlertThresholds
	notificationURL string
	logger          *zap.Logger
	enabled         bool
}

// NewDatabaseMonitor creates a new database monitor
func NewDatabaseMonitor(
	databaseManager *database.DatabaseManager,
	router *database.DatabaseRouter,
	config *MonitoringConfig,
	logger *zap.Logger,
) *DatabaseMonitor {
	monitor := &DatabaseMonitor{
		databaseManager: databaseManager,
		router:          router,
		logger:          logger,
		config:          config,
		metrics:         &MonitoringMetrics{},
		collectors:      make(map[string]*MetricCollector),
		alertManager:    NewAlertManager(config.AlertThresholds, logger, config.AlertingEnabled),
	}

	if config.Enabled {
		// Start monitoring
		go monitor.startMonitoring()

		// Start health checks
		go monitor.startHealthChecks()

		logger.Info("Database monitor initialized",
			zap.Duration("collection_interval", config.CollectionInterval),
			zap.Bool("alerting_enabled", config.AlertingEnabled),
			zap.Bool("optimization_enabled", config.OptimizationEnabled),
		)
	}

	return monitor
}

// startMonitoring starts the main monitoring loop
func (dm *DatabaseMonitor) startMonitoring() {
	ticker := time.NewTicker(dm.config.CollectionInterval)
	defer ticker.Stop()

	for range ticker.C {
		dm.collectAllMetrics()
	}
}

// startHealthChecks starts the health check loop
func (dm *DatabaseMonitor) startHealthChecks() {
	ticker := time.NewTicker(dm.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		dm.performHealthChecks()
	}
}

// collectAllMetrics collects metrics for all schools
func (dm *DatabaseMonitor) collectAllMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	// Get list of schools from master database
	schools, err := dm.getActiveSchools(ctx)
	if err != nil {
		dm.logger.Error("Failed to get active schools", zap.Error(err))
		return
	}

	dm.metrics.mu.Lock()
	dm.metrics.TotalSchools = int64(len(schools))
	dm.metrics.mu.Unlock()

	// Collect metrics for each school
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit concurrent collections

	for _, schoolID := range schools {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := dm.collectSchoolMetrics(ctx, id); err != nil {
				dm.logger.Warn("Failed to collect metrics for school",
					zap.String("school_id", id),
					zap.Error(err),
				)
			}
		}(schoolID)
	}

	wg.Wait()
	dm.updateOverallMetrics()
}

// getActiveSchools gets list of active schools
func (dm *DatabaseMonitor) getActiveSchools(ctx context.Context) ([]string, error) {
	masterDB := dm.databaseManager.GetMasterConnection()
	docs, err := masterDB.Collection("schools").Where("status", "==", "active").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	schools := make([]string, len(docs))
	for i, doc := range docs {
		schools[i] = doc.Ref.ID
	}

	return schools, nil
}

// collectSchoolMetrics collects metrics for a specific school
func (dm *DatabaseMonitor) collectSchoolMetrics(ctx context.Context, schoolID string) error {
	// Get or create metric collector
	collector := dm.getOrCreateCollector(schoolID)

	// Collect connection metrics
	connectionMetrics, err := dm.collectConnectionMetrics(ctx, schoolID)
	if err != nil {
		return fmt.Errorf("failed to collect connection metrics: %w", err)
	}

	// Collect query metrics
	queryMetrics, err := dm.collectQueryMetrics(ctx, schoolID)
	if err != nil {
		return fmt.Errorf("failed to collect query metrics: %w", err)
	}

	// Collect performance metrics
	performanceMetrics, err := dm.collectPerformanceMetrics(ctx, schoolID)
	if err != nil {
		return fmt.Errorf("failed to collect performance metrics: %w", err)
	}

	// Collect resource metrics
	resourceMetrics, err := dm.collectResourceMetrics(ctx, schoolID)
	if err != nil {
		return fmt.Errorf("failed to collect resource metrics: %w", err)
	}

	// Collect health metrics
	healthMetrics, err := dm.collectHealthMetrics(ctx, schoolID)
	if err != nil {
		return fmt.Errorf("failed to collect health metrics: %w", err)
	}

	// Update collector metrics
	collector.mu.Lock()
	collector.Metrics = &SchoolMetrics{
		ConnectionMetrics:  connectionMetrics,
		QueryMetrics:       queryMetrics,
		PerformanceMetrics: performanceMetrics,
		ResourceMetrics:    resourceMetrics,
		HealthMetrics:      healthMetrics,
		LastUpdated:        time.Now(),
	}
	collector.LastCollected = time.Now()
	collector.mu.Unlock()

	// Check for alerts
	dm.checkAlerts(schoolID, collector.Metrics)

	// Generate recommendations
	if dm.config.RecommendationsEnabled {
		dm.generateRecommendations(schoolID, collector.Metrics)
	}

	return nil
}

// getOrCreateCollector gets or creates a metric collector for a school
func (dm *DatabaseMonitor) getOrCreateCollector(schoolID string) *MetricCollector {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if collector, exists := dm.collectors[schoolID]; exists {
		return collector
	}

	collector := &MetricCollector{
		SchoolID:        schoolID,
		DatabaseName:    fmt.Sprintf("school-%s", schoolID),
		HealthStatus:    "unknown",
		Alerts:          make([]*Alert, 0),
		Recommendations: make([]*Recommendation, 0),
	}

	dm.collectors[schoolID] = collector
	return collector
}

// Metric collection implementations

func (dm *DatabaseMonitor) collectConnectionMetrics(ctx context.Context, schoolID string) (*ConnectionMetrics, error) {
	// Get router metrics for connection info
	routerMetrics := dm.router.GetRouterMetrics()

	return &ConnectionMetrics{
		ActiveConnections:   routerMetrics.ActiveConnections,
		TotalConnections:    routerMetrics.SuccessfulRoutes,
		MaxConnections:      100, // From config
		ConnectionPoolUsage: float64(routerMetrics.ActiveConnections) / 100.0 * 100,
		IdleConnections:     100 - routerMetrics.ActiveConnections,
		ConnectionErrors:    routerMetrics.ConnectionErrors,
		AvgConnectionTime:   routerMetrics.AverageRouteTime,
		ConnectionTimeouts:  0, // Would be tracked separately
	}, nil
}

func (dm *DatabaseMonitor) collectQueryMetrics(ctx context.Context, schoolID string) (*QueryMetrics, error) {
	// In a real implementation, this would collect actual query metrics
	// For now, we'll return simulated metrics
	return &QueryMetrics{
		TotalQueries:      1000,
		SuccessfulQueries: 950,
		FailedQueries:     50,
		ReadQueries:       800,
		WriteQueries:      200,
		AverageLatency:    time.Millisecond * 50,
		P95Latency:        time.Millisecond * 100,
		P99Latency:        time.Millisecond * 200,
		SlowQueries:       5,
		QueryErrors:       45,
		QueryTimeouts:     5,
	}, nil
}

func (dm *DatabaseMonitor) collectPerformanceMetrics(ctx context.Context, schoolID string) (*PerformanceMetrics, error) {
	return &PerformanceMetrics{
		ThroughputQPS:   20.5,
		ThroughputRPS:   15.2,
		CPUUsage:        45.5,
		MemoryUsage:     60.2,
		DiskIOPS:        150.0,
		NetworkBytesIn:  1024000,
		NetworkBytesOut: 512000,
		CacheHitRatio:   85.5,
		IndexUsageRatio: 92.1,
	}, nil
}

func (dm *DatabaseMonitor) collectResourceMetrics(ctx context.Context, schoolID string) (*ResourceMetrics, error) {
	return &ResourceMetrics{
		StorageUsedBytes:    1024 * 1024 * 500,       // 500MB
		StorageTotalBytes:   1024 * 1024 * 1024 * 10, // 10GB
		StorageUsagePercent: 5.0,
		DocumentCount:       10000,
		CollectionCount:     15,
		IndexCount:          45,
		BackupSizeBytes:     1024 * 1024 * 100, // 100MB
		BackupCount:         7,
		LastBackupTime:      time.Now().Add(-time.Hour * 24),
	}, nil
}

func (dm *DatabaseMonitor) collectHealthMetrics(ctx context.Context, schoolID string) (*HealthMetrics, error) {
	return &HealthMetrics{
		AvailabilityPercent: 99.9,
		UptimeSeconds:       86400 * 7, // 7 days
		LastDowntime:        nil,
		HealthScore:         95,
		ErrorRate:           0.1,
		WarningCount:        2,
		CriticalIssues:      0,
		LastHealthCheck:     time.Now(),
	}, nil
}

// checkAlerts checks metrics against alert thresholds
func (dm *DatabaseMonitor) checkAlerts(schoolID string, metrics *SchoolMetrics) {
	thresholds := dm.config.AlertThresholds

	// Check connection pool usage
	if metrics.ConnectionMetrics.ConnectionPoolUsage > thresholds.HighConnectionUsage {
		alert := &Alert{
			ID:          fmt.Sprintf("high_connection_usage_%s_%d", schoolID, time.Now().Unix()),
			SchoolID:    schoolID,
			Type:        "connection_pool",
			Severity:    "high",
			Title:       "High Connection Pool Usage",
			Message:     fmt.Sprintf("Connection pool usage is %.1f%%, threshold is %.1f%%", metrics.ConnectionMetrics.ConnectionPoolUsage, thresholds.HighConnectionUsage),
			Metric:      "connection_pool_usage",
			Threshold:   thresholds.HighConnectionUsage,
			ActualValue: metrics.ConnectionMetrics.ConnectionPoolUsage,
			TriggeredAt: time.Now(),
			Status:      "active",
			Actions:     []string{"scale_connections", "investigate_leaks"},
		}
		dm.alertManager.TriggerAlert(alert)
	}

	// Check query latency
	if metrics.QueryMetrics.AverageLatency > thresholds.HighQueryLatency {
		alert := &Alert{
			ID:          fmt.Sprintf("high_query_latency_%s_%d", schoolID, time.Now().Unix()),
			SchoolID:    schoolID,
			Type:        "performance",
			Severity:    "medium",
			Title:       "High Query Latency",
			Message:     fmt.Sprintf("Average query latency is %v, threshold is %v", metrics.QueryMetrics.AverageLatency, thresholds.HighQueryLatency),
			Metric:      "query_latency",
			Threshold:   thresholds.HighQueryLatency,
			ActualValue: metrics.QueryMetrics.AverageLatency,
			TriggeredAt: time.Now(),
			Status:      "active",
			Actions:     []string{"optimize_queries", "add_indexes", "check_resources"},
		}
		dm.alertManager.TriggerAlert(alert)
	}

	// Check error rate
	errorRate := float64(metrics.QueryMetrics.FailedQueries) / float64(metrics.QueryMetrics.TotalQueries) * 100
	if errorRate > thresholds.HighErrorRate {
		alert := &Alert{
			ID:          fmt.Sprintf("high_error_rate_%s_%d", schoolID, time.Now().Unix()),
			SchoolID:    schoolID,
			Type:        "reliability",
			Severity:    "critical",
			Title:       "High Error Rate",
			Message:     fmt.Sprintf("Error rate is %.2f%%, threshold is %.2f%%", errorRate, thresholds.HighErrorRate),
			Metric:      "error_rate",
			Threshold:   thresholds.HighErrorRate,
			ActualValue: errorRate,
			TriggeredAt: time.Now(),
			Status:      "active",
			Actions:     []string{"investigate_errors", "check_database_health", "escalate"},
		}
		dm.alertManager.TriggerAlert(alert)
	}

	// Check availability
	if metrics.HealthMetrics.AvailabilityPercent < thresholds.LowAvailability {
		alert := &Alert{
			ID:          fmt.Sprintf("low_availability_%s_%d", schoolID, time.Now().Unix()),
			SchoolID:    schoolID,
			Type:        "availability",
			Severity:    "critical",
			Title:       "Low Availability",
			Message:     fmt.Sprintf("Availability is %.2f%%, threshold is %.2f%%", metrics.HealthMetrics.AvailabilityPercent, thresholds.LowAvailability),
			Metric:      "availability",
			Threshold:   thresholds.LowAvailability,
			ActualValue: metrics.HealthMetrics.AvailabilityPercent,
			TriggeredAt: time.Now(),
			Status:      "active",
			Actions:     []string{"check_infrastructure", "escalate_immediately", "prepare_failover"},
		}
		dm.alertManager.TriggerAlert(alert)
	}

	// Check disk usage
	if metrics.ResourceMetrics.StorageUsagePercent > thresholds.HighDiskUsage {
		alert := &Alert{
			ID:          fmt.Sprintf("high_disk_usage_%s_%d", schoolID, time.Now().Unix()),
			SchoolID:    schoolID,
			Type:        "resource",
			Severity:    "high",
			Title:       "High Disk Usage",
			Message:     fmt.Sprintf("Disk usage is %.1f%%, threshold is %.1f%%", metrics.ResourceMetrics.StorageUsagePercent, thresholds.HighDiskUsage),
			Metric:      "disk_usage",
			Threshold:   thresholds.HighDiskUsage,
			ActualValue: metrics.ResourceMetrics.StorageUsagePercent,
			TriggeredAt: time.Now(),
			Status:      "active",
			Actions:     []string{"cleanup_old_data", "increase_storage", "optimize_storage"},
		}
		dm.alertManager.TriggerAlert(alert)
	}
}

// generateRecommendations generates optimization recommendations
func (dm *DatabaseMonitor) generateRecommendations(schoolID string, metrics *SchoolMetrics) {
	recommendations := make([]*Recommendation, 0)

	// Index optimization recommendation
	if metrics.PerformanceMetrics.IndexUsageRatio < 80.0 {
		recommendations = append(recommendations, &Recommendation{
			ID:          fmt.Sprintf("optimize_indexes_%s_%d", schoolID, time.Now().Unix()),
			SchoolID:    schoolID,
			Type:        "performance",
			Priority:    "medium",
			Title:       "Optimize Database Indexes",
			Description: "Index usage ratio is low. Consider reviewing and optimizing indexes.",
			Impact:      "Improved query performance, reduced latency",
			Effort:      "medium",
			Category:    "database_optimization",
			CreatedAt:   time.Now(),
			EstimatedSavings: map[string]interface{}{
				"latency_reduction": "20-40%",
				"cost_savings":      "$50-100/month",
			},
		})
	}

	// Cache optimization recommendation
	if metrics.PerformanceMetrics.CacheHitRatio < 85.0 {
		recommendations = append(recommendations, &Recommendation{
			ID:          fmt.Sprintf("optimize_cache_%s_%d", schoolID, time.Now().Unix()),
			SchoolID:    schoolID,
			Type:        "performance",
			Priority:    "high",
			Title:       "Improve Cache Hit Ratio",
			Description: "Cache hit ratio is below optimal. Consider cache tuning or increasing cache size.",
			Impact:      "Faster response times, reduced database load",
			Effort:      "low",
			Category:    "performance_tuning",
			CreatedAt:   time.Now(),
			EstimatedSavings: map[string]interface{}{
				"latency_reduction": "15-30%",
				"cost_savings":      "$25-75/month",
			},
		})
	}

	// Storage optimization recommendation
	if metrics.ResourceMetrics.StorageUsagePercent > 70.0 {
		recommendations = append(recommendations, &Recommendation{
			ID:          fmt.Sprintf("storage_cleanup_%s_%d", schoolID, time.Now().Unix()),
			SchoolID:    schoolID,
			Type:        "cost",
			Priority:    "medium",
			Title:       "Storage Cleanup and Optimization",
			Description: "Storage usage is high. Consider archiving old data or implementing compression.",
			Impact:      "Reduced storage costs, improved performance",
			Effort:      "medium",
			Category:    "storage_optimization",
			CreatedAt:   time.Now(),
			EstimatedSavings: map[string]interface{}{
				"storage_savings": "30-50%",
				"cost_savings":    "$100-200/month",
			},
		})
	}

	// Connection pool optimization
	if metrics.ConnectionMetrics.ConnectionPoolUsage > 80.0 {
		recommendations = append(recommendations, &Recommendation{
			ID:          fmt.Sprintf("connection_pool_tune_%s_%d", schoolID, time.Now().Unix()),
			SchoolID:    schoolID,
			Type:        "performance",
			Priority:    "high",
			Title:       "Optimize Connection Pool",
			Description: "Connection pool usage is high. Consider increasing pool size or optimizing connection usage.",
			Impact:      "Better scalability, reduced connection timeouts",
			Effort:      "low",
			Category:    "connection_optimization",
			CreatedAt:   time.Now(),
			EstimatedSavings: map[string]interface{}{
				"reliability_improvement": "15-25%",
				"cost_savings":            "$30-60/month",
			},
		})
	}

	// Store recommendations
	if len(recommendations) > 0 {
		collector := dm.getOrCreateCollector(schoolID)
		collector.mu.Lock()
		collector.Recommendations = append(collector.Recommendations, recommendations...)
		collector.mu.Unlock()

		dm.metrics.mu.Lock()
		dm.metrics.RecommendationsGenerated += int64(len(recommendations))
		dm.metrics.mu.Unlock()

		dm.logger.Info("Generated recommendations",
			zap.String("school_id", schoolID),
			zap.Int("count", len(recommendations)),
		)
	}
}

// performHealthChecks performs health checks for all schools
func (dm *DatabaseMonitor) performHealthChecks() {
	dm.mu.RLock()
	collectors := make([]*MetricCollector, 0, len(dm.collectors))
	for _, collector := range dm.collectors {
		collectors = append(collectors, collector)
	}
	dm.mu.RUnlock()

	for _, collector := range collectors {
		dm.performSchoolHealthCheck(collector)
	}
}

// performSchoolHealthCheck performs health check for a specific school
func (dm *DatabaseMonitor) performSchoolHealthCheck(collector *MetricCollector) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// Try to connect to the school database
	securityCtx := &security.SecurityContext{
		SchoolID: collector.SchoolID,
		UserID:   "system",
		Role:     "system_admin",
	}

	client, err := dm.databaseManager.GetSecureConnection(collector.SchoolID, securityCtx)
	if err != nil {
		collector.mu.Lock()
		collector.HealthStatus = "unhealthy"
		collector.mu.Unlock()

		dm.logger.Warn("Health check failed for school",
			zap.String("school_id", collector.SchoolID),
			zap.Error(err),
		)
		return
	}

	// Perform a simple read operation
	_, err = client.Collection("health_check").Doc("test").Get(ctx)
	if err != nil && !isNotFoundError(err) {
		collector.mu.Lock()
		collector.HealthStatus = "degraded"
		collector.mu.Unlock()

		dm.logger.Warn("Health check degraded for school",
			zap.String("school_id", collector.SchoolID),
			zap.Error(err),
		)
		return
	}

	collector.mu.Lock()
	collector.HealthStatus = "healthy"
	collector.mu.Unlock()
}

// updateOverallMetrics updates overall monitoring metrics
func (dm *DatabaseMonitor) updateOverallMetrics() {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	var totalConnections int64
	var totalQueries int64
	var successfulQueries int64
	var failedQueries int64
	var totalLatency time.Duration
	var count int64

	for _, collector := range dm.collectors {
		collector.mu.RLock()
		if collector.Metrics != nil {
			totalConnections += collector.Metrics.ConnectionMetrics.ActiveConnections
			totalQueries += collector.Metrics.QueryMetrics.TotalQueries
			successfulQueries += collector.Metrics.QueryMetrics.SuccessfulQueries
			failedQueries += collector.Metrics.QueryMetrics.FailedQueries
			totalLatency += collector.Metrics.QueryMetrics.AverageLatency
			count++
		}
		collector.mu.RUnlock()
	}

	dm.metrics.mu.Lock()
	dm.metrics.ActiveConnections = totalConnections
	dm.metrics.TotalQueries = totalQueries
	dm.metrics.SuccessfulQueries = successfulQueries
	dm.metrics.FailedQueries = failedQueries
	if count > 0 {
		dm.metrics.AverageLatency = totalLatency / time.Duration(count)
	}
	dm.metrics.mu.Unlock()
}

// Helper functions

func isNotFoundError(err error) bool {
	return err != nil && (err.Error() == "not found" ||
		fmt.Sprintf("%v", err) == "rpc error: code = NotFound desc = No document to get")
}

// NewAlertManager creates a new alert manager
func NewAlertManager(thresholds *AlertThresholds, logger *zap.Logger, enabled bool) *AlertManager {
	return &AlertManager{
		thresholds: thresholds,
		logger:     logger,
		enabled:    enabled,
	}
}

// TriggerAlert triggers a new alert
func (am *AlertManager) TriggerAlert(alert *Alert) {
	if !am.enabled {
		return
	}

	am.alerts.Store(alert.ID, alert)

	am.logger.Warn("Alert triggered",
		zap.String("alert_id", alert.ID),
		zap.String("school_id", alert.SchoolID),
		zap.String("type", alert.Type),
		zap.String("severity", alert.Severity),
		zap.String("title", alert.Title),
	)

	// In a real implementation, this would send notifications
}

// GetSchoolMetrics returns metrics for a specific school
func (dm *DatabaseMonitor) GetSchoolMetrics(schoolID string) (*SchoolMetrics, error) {
	dm.mu.RLock()
	collector, exists := dm.collectors[schoolID]
	dm.mu.RUnlock()

	if !exists {
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeResourceNotFound,
			fmt.Sprintf("No metrics found for school %s", schoolID),
			errors.CategoryMonitoring,
			errors.SeverityLow,
			false,
		)
	}

	collector.mu.RLock()
	defer collector.mu.RUnlock()

	return collector.Metrics, nil
}

// GetMonitoringMetrics returns overall monitoring metrics
func (dm *DatabaseMonitor) GetMonitoringMetrics() *MonitoringMetrics {
	dm.metrics.mu.RLock()
	defer dm.metrics.mu.RUnlock()

	return &MonitoringMetrics{
		TotalSchools:             dm.metrics.TotalSchools,
		ActiveConnections:        dm.metrics.ActiveConnections,
		TotalQueries:             dm.metrics.TotalQueries,
		SuccessfulQueries:        dm.metrics.SuccessfulQueries,
		FailedQueries:            dm.metrics.FailedQueries,
		AverageLatency:           dm.metrics.AverageLatency,
		PeakLatency:              dm.metrics.PeakLatency,
		AlertsTriggered:          dm.metrics.AlertsTriggered,
		RecommendationsGenerated: dm.metrics.RecommendationsGenerated,
	}
}

// GetSchoolAlerts returns active alerts for a school
func (dm *DatabaseMonitor) GetSchoolAlerts(schoolID string) ([]*Alert, error) {
	dm.mu.RLock()
	collector, exists := dm.collectors[schoolID]
	dm.mu.RUnlock()

	if !exists {
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeResourceNotFound,
			fmt.Sprintf("No alerts found for school %s", schoolID),
			errors.CategoryMonitoring,
			errors.SeverityLow,
			false,
		)
	}

	collector.mu.RLock()
	defer collector.mu.RUnlock()

	return collector.Alerts, nil
}

// GetSchoolRecommendations returns recommendations for a school
func (dm *DatabaseMonitor) GetSchoolRecommendations(schoolID string) ([]*Recommendation, error) {
	dm.mu.RLock()
	collector, exists := dm.collectors[schoolID]
	dm.mu.RUnlock()

	if !exists {
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeResourceNotFound,
			fmt.Sprintf("No recommendations found for school %s", schoolID),
			errors.CategoryMonitoring,
			errors.SeverityLow,
			false,
		)
	}

	collector.mu.RLock()
	defer collector.mu.RUnlock()

	return collector.Recommendations, nil
}

// DefaultMonitoringConfig returns default monitoring configuration
func DefaultMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		Enabled:               true,
		CollectionInterval:    time.Minute * 5,
		AlertingEnabled:       true,
		PerformanceMonitoring: true,
		HealthCheckInterval:   time.Minute * 2,
		MetricsRetentionDays:  30,
		AlertThresholds: &AlertThresholds{
			HighConnectionUsage:     80.0,
			HighQueryLatency:        time.Millisecond * 1000,
			HighErrorRate:           5.0,
			LowAvailability:         99.0,
			HighDiskUsage:           85.0,
			SlowQueryThreshold:      time.Second * 2,
			ConnectionPoolExhausted: 95,
		},
		OptimizationEnabled:    true,
		RecommendationsEnabled: true,
	}
}
