package database

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/schoolgpt/backend/internal/errors"
	"github.com/schoolgpt/backend/internal/security"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

// DatabaseManager manages secure database connections
type DatabaseManager struct {
	masterDB      *firestore.Client
	schoolDBs     sync.Map // Thread-safe map[schoolID]*firestore.Client
	connPool      *ConnectionPool
	encryptionKey []byte
	audit         *security.AuditLogger
	monitor       *PerformanceMonitor
	healthChecker *HealthChecker
	logger        *zap.Logger
	config        *DatabaseConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	MasterProjectID     string
	SchoolProjectPrefix string
	Region              string
	MaxConnections      int
	IdleTimeout         time.Duration
	ConnectionTimeout   time.Duration
	HealthCheckInterval time.Duration
	RetryAttempts       int
	RetryDelay          time.Duration
	EnableEncryption    bool
	CredentialsPath     string
}

// ConnectionPool manages database connection pooling
type ConnectionPool struct {
	maxConnections      int
	idleTimeout         time.Duration
	connectionTimeout   time.Duration
	healthCheckInterval time.Duration
	connections         sync.Map // map[string]*ConnectionInfo
	metrics             *PoolMetrics
	mu                  sync.RWMutex
}

// ConnectionInfo holds connection metadata
type ConnectionInfo struct {
	Client    *firestore.Client
	SchoolID  string
	CreatedAt time.Time
	LastUsed  time.Time
	UseCount  int64
	IsHealthy bool
	ProjectID string
	mu        sync.RWMutex
}

// PoolMetrics tracks connection pool metrics
type PoolMetrics struct {
	TotalConnections   int64
	ActiveConnections  int64
	ConnectionsCreated int64
	ConnectionsReused  int64
	ConnectionsRemoved int64
	HealthChecksFailed int64
	mu                 sync.RWMutex
}

// PerformanceMonitor tracks database performance metrics
type PerformanceMonitor struct {
	queryLatencies  sync.Map // map[string]*LatencyTracker
	connectionTimes sync.Map // map[string]*LatencyTracker
	errorCounts     sync.Map // map[string]int64
	logger          *zap.Logger
}

// LatencyTracker tracks latency statistics
type LatencyTracker struct {
	Count int64
	Total time.Duration
	Min   time.Duration
	Max   time.Duration
	mu    sync.RWMutex
}

// HealthChecker performs health checks on database connections
type HealthChecker struct {
	interval time.Duration
	timeout  time.Duration
	logger   *zap.Logger
	stopCh   chan struct{}
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(config *DatabaseConfig, logger *zap.Logger) (*DatabaseManager, error) {
	// Generate encryption key
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Create master database connection
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectionTimeout)
	defer cancel()

	var opts []option.ClientOption
	if config.CredentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(config.CredentialsPath))
	}

	masterDB, err := firestore.NewClient(ctx, config.MasterProjectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create master database connection: %w", err)
	}

	dm := &DatabaseManager{
		masterDB:      masterDB,
		encryptionKey: encryptionKey,
		audit:         security.NewAuditLogger(),
		monitor:       NewPerformanceMonitor(logger),
		healthChecker: NewHealthChecker(config.HealthCheckInterval, logger),
		logger:        logger,
		config:        config,
		connPool: &ConnectionPool{
			maxConnections:      config.MaxConnections,
			idleTimeout:         config.IdleTimeout,
			connectionTimeout:   config.ConnectionTimeout,
			healthCheckInterval: config.HealthCheckInterval,
			metrics:             &PoolMetrics{},
		},
	}

	// Start health checker
	go dm.healthChecker.Start(dm)

	// Start connection cleanup
	go dm.cleanupConnections()

	logger.Info("Database manager initialized",
		zap.String("master_project", config.MasterProjectID),
		zap.Int("max_connections", config.MaxConnections),
	)

	return dm, nil
}

// GetSecureConnection gets a secure connection to a school database
func (dm *DatabaseManager) GetSecureConnection(schoolID string, ctx *security.SecurityContext) (*firestore.Client, error) {
	startTime := time.Now()
	defer func() {
		dm.monitor.RecordConnectionTime(schoolID, time.Since(startTime))
	}()

	// Validate security context
	if err := dm.validateSecurityContext(ctx); err != nil {
		dm.audit.LogSecurityViolation(ctx, "invalid_security_context")
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeUnauthorized,
			"Security context validation failed",
			errors.CategorySecurity,
			errors.SeverityHigh,
			false,
		).WithDetails(err.Error())
	}

	// Check connection pool first
	if connInfo := dm.getFromPool(schoolID); connInfo != nil {
		// Validate connection health
		if dm.isConnectionHealthy(connInfo.Client, schoolID) {
			dm.updateConnectionUsage(connInfo)
			dm.connPool.metrics.ConnectionsReused++
			dm.monitor.RecordConnectionReuse(schoolID)
			return connInfo.Client, nil
		}
		// Remove unhealthy connection
		dm.removeFromPool(schoolID)
	}

	// Create new secure connection
	client, err := dm.createSecureConnection(schoolID, ctx)
	if err != nil {
		dm.audit.LogEvent(security.AuditEvent{
			EventType: "connection_failure",
			UserID:    ctx.UserID,
			SchoolID:  schoolID,
			IPAddress: ctx.IPAddress,
			Success:   false,
			Severity:  security.SeverityHigh,
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return nil, fmt.Errorf("failed to create secure connection: %w", err)
	}

	// Store in pool
	dm.addToPool(schoolID, client)
	dm.connPool.metrics.ConnectionsCreated++
	dm.monitor.RecordNewConnection(schoolID)

	dm.logger.Info("New database connection created",
		zap.String("school_id", schoolID),
		zap.String("user_id", ctx.UserID),
	)

	return client, nil
}

// GetMasterConnection returns the master database connection
func (dm *DatabaseManager) GetMasterConnection() *firestore.Client {
	return dm.masterDB
}

// createSecureConnection creates a new secure connection to a school database
func (dm *DatabaseManager) createSecureConnection(schoolID string, ctx *security.SecurityContext) (*firestore.Client, error) {
	// Generate school project ID
	projectID := dm.generateSchoolProjectID(schoolID)

	// Create connection with timeout
	connCtx, cancel := context.WithTimeout(context.Background(), dm.config.ConnectionTimeout)
	defer cancel()

	var opts []option.ClientOption
	if dm.config.CredentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(dm.config.CredentialsPath))
	}

	client, err := firestore.NewClient(connCtx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client for school %s: %w", schoolID, err)
	}

	// Test connection
	if err := dm.testConnection(client); err != nil {
		client.Close()
		return nil, fmt.Errorf("connection test failed for school %s: %w", schoolID, err)
	}

	return client, nil
}

// generateSchoolProjectID generates a project ID for a school
func (dm *DatabaseManager) generateSchoolProjectID(schoolID string) string {
	return fmt.Sprintf("%s-%s", dm.config.SchoolProjectPrefix, schoolID)
}

// testConnection tests if a database connection is working
func (dm *DatabaseManager) testConnection(client *firestore.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Try to read a simple document to test connection
	_, err := client.Collection("health_check").Doc("test").Get(ctx)
	// We expect this to fail with "not found" which means connection is working
	if err != nil && !isNotFoundError(err) {
		return err
	}

	return nil
}

// isNotFoundError checks if error is a "not found" error
func isNotFoundError(err error) bool {
	// Implementation depends on Firestore error types
	return err.Error() == "document not found" || err.Error() == "collection not found"
}

// validateSecurityContext validates the security context
func (dm *DatabaseManager) validateSecurityContext(ctx *security.SecurityContext) error {
	if ctx == nil {
		return fmt.Errorf("security context is nil")
	}

	if ctx.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if ctx.SchoolID == "" {
		return fmt.Errorf("school ID is required")
	}

	if time.Now().After(ctx.ExpiresAt) {
		return fmt.Errorf("security context has expired")
	}

	return nil
}

// getFromPool retrieves a connection from the pool
func (dm *DatabaseManager) getFromPool(schoolID string) *ConnectionInfo {
	if connInterface, exists := dm.schoolDBs.Load(schoolID); exists {
		if connInfo, ok := connInterface.(*ConnectionInfo); ok {
			connInfo.mu.RLock()
			defer connInfo.mu.RUnlock()

			// Check if connection is not too old
			if time.Since(connInfo.LastUsed) < dm.config.IdleTimeout {
				return connInfo
			}
		}
	}
	return nil
}

// addToPool adds a connection to the pool
func (dm *DatabaseManager) addToPool(schoolID string, client *firestore.Client) {
	connInfo := &ConnectionInfo{
		Client:    client,
		SchoolID:  schoolID,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		UseCount:  1,
		IsHealthy: true,
		ProjectID: dm.generateSchoolProjectID(schoolID),
	}

	dm.schoolDBs.Store(schoolID, connInfo)
	dm.connPool.metrics.TotalConnections++
	dm.connPool.metrics.ActiveConnections++
}

// removeFromPool removes a connection from the pool
func (dm *DatabaseManager) removeFromPool(schoolID string) {
	if connInterface, exists := dm.schoolDBs.LoadAndDelete(schoolID); exists {
		if connInfo, ok := connInterface.(*ConnectionInfo); ok {
			connInfo.Client.Close()
			dm.connPool.metrics.ActiveConnections--
			dm.connPool.metrics.ConnectionsRemoved++
		}
	}
}

// updateConnectionUsage updates connection usage statistics
func (dm *DatabaseManager) updateConnectionUsage(connInfo *ConnectionInfo) {
	connInfo.mu.Lock()
	defer connInfo.mu.Unlock()

	connInfo.LastUsed = time.Now()
	connInfo.UseCount++
}

// isConnectionHealthy checks if a connection is healthy
func (dm *DatabaseManager) isConnectionHealthy(client *firestore.Client, schoolID string) bool {
	if err := dm.testConnection(client); err != nil {
		dm.logger.Warn("Connection health check failed",
			zap.String("school_id", schoolID),
			zap.Error(err),
		)
		dm.connPool.metrics.HealthChecksFailed++
		return false
	}
	return true
}

// cleanupConnections removes idle connections
func (dm *DatabaseManager) cleanupConnections() {
	ticker := time.NewTicker(dm.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		dm.schoolDBs.Range(func(key, value interface{}) bool {
			schoolID := key.(string)
			connInfo := value.(*ConnectionInfo)

			connInfo.mu.RLock()
			lastUsed := connInfo.LastUsed
			connInfo.mu.RUnlock()

			// Remove connections that have been idle too long
			if time.Since(lastUsed) > dm.config.IdleTimeout {
				dm.logger.Debug("Removing idle connection",
					zap.String("school_id", schoolID),
					zap.Duration("idle_time", time.Since(lastUsed)),
				)
				dm.removeFromPool(schoolID)
			}

			return true
		})
	}
}

// Close closes all database connections
func (dm *DatabaseManager) Close() error {
	// Close master connection
	if err := dm.masterDB.Close(); err != nil {
		dm.logger.Error("Failed to close master database connection", zap.Error(err))
	}

	// Close all school connections
	dm.schoolDBs.Range(func(key, value interface{}) bool {
		if connInfo, ok := value.(*ConnectionInfo); ok {
			if err := connInfo.Client.Close(); err != nil {
				dm.logger.Error("Failed to close school database connection",
					zap.String("school_id", connInfo.SchoolID),
					zap.Error(err),
				)
			}
		}
		return true
	})

	// Stop health checker
	dm.healthChecker.Stop()

	dm.logger.Info("Database manager closed")
	return nil
}

// GetConnectionMetrics returns connection pool metrics
func (dm *DatabaseManager) GetConnectionMetrics() *PoolMetrics {
	dm.connPool.metrics.mu.RLock()
	defer dm.connPool.metrics.mu.RUnlock()

	return &PoolMetrics{
		TotalConnections:   dm.connPool.metrics.TotalConnections,
		ActiveConnections:  dm.connPool.metrics.ActiveConnections,
		ConnectionsCreated: dm.connPool.metrics.ConnectionsCreated,
		ConnectionsReused:  dm.connPool.metrics.ConnectionsReused,
		ConnectionsRemoved: dm.connPool.metrics.ConnectionsRemoved,
		HealthChecksFailed: dm.connPool.metrics.HealthChecksFailed,
	}
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(logger *zap.Logger) *PerformanceMonitor {
	return &PerformanceMonitor{
		logger: logger,
	}
}

// RecordConnectionTime records connection establishment time
func (pm *PerformanceMonitor) RecordConnectionTime(schoolID string, duration time.Duration) {
	pm.recordLatency("connection_time", schoolID, duration)
}

// RecordConnectionReuse records connection reuse
func (pm *PerformanceMonitor) RecordConnectionReuse(schoolID string) {
	// Implementation for connection reuse metrics
}

// RecordNewConnection records new connection creation
func (pm *PerformanceMonitor) RecordNewConnection(schoolID string) {
	// Implementation for new connection metrics
}

// RecordQueryLatency records database query latency
func (pm *PerformanceMonitor) RecordQueryLatency(operation, schoolID string, duration time.Duration) {
	key := fmt.Sprintf("%s_%s", operation, schoolID)
	pm.recordLatency("query_latency", key, duration)
}

func (pm *PerformanceMonitor) recordLatency(metricType, key string, duration time.Duration) {
	fullKey := fmt.Sprintf("%s_%s", metricType, key)

	trackerInterface, _ := pm.queryLatencies.LoadOrStore(fullKey, &LatencyTracker{
		Min: duration,
		Max: duration,
	})

	tracker := trackerInterface.(*LatencyTracker)
	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	tracker.Count++
	tracker.Total += duration

	if duration < tracker.Min || tracker.Min == 0 {
		tracker.Min = duration
	}
	if duration > tracker.Max {
		tracker.Max = duration
	}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(interval time.Duration, logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		interval: interval,
		timeout:  time.Second * 5,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

// Start starts the health checker
func (hc *HealthChecker) Start(dm *DatabaseManager) {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.performHealthChecks(dm)
		case <-hc.stopCh:
			return
		}
	}
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}

// performHealthChecks performs health checks on all connections
func (hc *HealthChecker) performHealthChecks(dm *DatabaseManager) {
	dm.schoolDBs.Range(func(key, value interface{}) bool {
		schoolID := key.(string)
		connInfo := value.(*ConnectionInfo)

		if !dm.isConnectionHealthy(connInfo.Client, schoolID) {
			hc.logger.Warn("Connection failed health check, removing from pool",
				zap.String("school_id", schoolID),
			)
			dm.removeFromPool(schoolID)
		}

		return true
	})
}

// DefaultDatabaseConfig returns default database configuration
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		MasterProjectID:     "schoolgpt-master",
		SchoolProjectPrefix: "schoolgpt-school",
		Region:              "us-central1",
		MaxConnections:      100,
		IdleTimeout:         time.Minute * 30,
		ConnectionTimeout:   time.Second * 10,
		HealthCheckInterval: time.Minute * 5,
		RetryAttempts:       3,
		RetryDelay:          time.Second * 2,
		EnableEncryption:    true,
	}
}
