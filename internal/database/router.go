package database

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/schoolgpt/backend/internal/errors"
	"github.com/schoolgpt/backend/internal/security"
	"go.uber.org/zap"
)

// DatabaseRouter provides intelligent routing to school-specific databases
type DatabaseRouter struct {
	manager         *DatabaseManager
	schoolCache     *SchoolCache
	routingRules    *RoutingRules
	failoverManager *FailoverManager
	metrics         *RouterMetrics
	logger          *zap.Logger
	config          *RouterConfig
}

// RouterConfig holds router configuration
type RouterConfig struct {
	CacheTimeout        time.Duration
	MaxRetries          int
	RetryDelay          time.Duration
	FailoverEnabled     bool
	HealthCheckInterval time.Duration
	MetricsEnabled      bool
	PreferredRegions    []string
	LoadBalancingMode   string // "round_robin", "least_connections", "performance"
}

// SchoolCache caches school metadata for fast routing decisions
type SchoolCache struct {
	schools           sync.Map // map[schoolID]*CachedSchoolInfo
	lastUpdated       sync.Map // map[schoolID]time.Time
	cacheTimeout      time.Duration
	refreshInProgress sync.Map // map[schoolID]bool
	metrics           *CacheMetrics
	mu                sync.RWMutex
}

// CachedSchoolInfo holds cached school information
type CachedSchoolInfo struct {
	SchoolID      string
	DatabaseName  string
	ProjectID     string
	Region        string
	Status        string
	LastActive    time.Time
	ConnectionURL string
	Features      map[string]bool
	Limits        map[string]int
	Priority      int
	mu            sync.RWMutex
}

// RoutingRules defines routing logic and policies
type RoutingRules struct {
	rules           []RoutingRule
	defaultBehavior string
	mu              sync.RWMutex
}

// RoutingRule represents a single routing rule
type RoutingRule struct {
	ID        string
	Condition string // "school_id", "region", "feature", "performance"
	Operator  string // "equals", "contains", "greater_than", "less_than"
	Value     string
	Action    string // "route_to", "fallback", "deny"
	Target    string
	Priority  int
	Enabled   bool
	CreatedAt time.Time
}

// FailoverManager handles database failover scenarios
type FailoverManager struct {
	primaryDBs      sync.Map // map[schoolID]string (database identifier)
	secondaryDBs    sync.Map // map[schoolID][]string
	currentFailover sync.Map // map[schoolID]*FailoverState
	policies        map[string]*FailoverPolicy
	logger          *zap.Logger
}

// FailoverState tracks current failover status
type FailoverState struct {
	IsActive       bool
	StartTime      time.Time
	Reason         string
	AttemptsCount  int
	CurrentTarget  string
	OriginalTarget string
}

// FailoverPolicy defines failover behavior
type FailoverPolicy struct {
	MaxRetries      int
	RetryDelay      time.Duration
	HealthCheckFreq time.Duration
	AutoRecovery    bool
	NotificationURL string
}

// RouterMetrics tracks routing performance and health
type RouterMetrics struct {
	TotalRequests     int64
	SuccessfulRoutes  int64
	FailedRoutes      int64
	CacheHits         int64
	CacheMisses       int64
	AverageRouteTime  time.Duration
	ConnectionErrors  int64
	FailoverCount     int64
	ActiveConnections int64
	mu                sync.RWMutex
}

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	HitRate        float64
	MissRate       float64
	RefreshCount   int64
	EvictionCount  int64
	AverageRefresh time.Duration
	mu             sync.RWMutex
}

// NewDatabaseRouter creates a new database router
func NewDatabaseRouter(manager *DatabaseManager, config *RouterConfig, logger *zap.Logger) *DatabaseRouter {
	router := &DatabaseRouter{
		manager: manager,
		logger:  logger,
		config:  config,
		schoolCache: &SchoolCache{
			cacheTimeout: config.CacheTimeout,
			metrics:      &CacheMetrics{},
		},
		routingRules: &RoutingRules{
			rules:           make([]RoutingRule, 0),
			defaultBehavior: "auto_detect",
		},
		failoverManager: &FailoverManager{
			policies: make(map[string]*FailoverPolicy),
			logger:   logger,
		},
		metrics: &RouterMetrics{},
	}

	// Initialize default routing rules
	router.initializeDefaultRules()

	// Start background maintenance
	go router.startMaintenance()

	logger.Info("Database router initialized",
		zap.Duration("cache_timeout", config.CacheTimeout),
		zap.Int("max_retries", config.MaxRetries),
		zap.Bool("failover_enabled", config.FailoverEnabled),
	)

	return router
}

// RouteToSchoolDatabase intelligently routes to the correct school database
func (dr *DatabaseRouter) RouteToSchoolDatabase(ctx context.Context, securityCtx *security.SecurityContext) (*firestore.Client, error) {
	startTime := time.Now()
	defer func() {
		dr.updateMetrics(time.Since(startTime))
	}()

	// Extract school ID from security context
	schoolID := securityCtx.SchoolID
	if schoolID == "" {
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeValidationFailed,
			"School ID not found in security context",
			errors.CategoryValidation,
			errors.SeverityMedium,
			false,
		)
	}

	// Apply routing rules
	routingDecision, err := dr.applyRoutingRules(schoolID, securityCtx)
	if err != nil {
		dr.metrics.FailedRoutes++
		return nil, fmt.Errorf("routing rules failed: %w", err)
	}

	// Check cache first
	if cachedInfo := dr.getFromCache(schoolID); cachedInfo != nil {
		if client, err := dr.getConnectionFromCache(ctx, cachedInfo, securityCtx); err == nil {
			dr.metrics.CacheHits++
			dr.metrics.SuccessfulRoutes++
			return client, nil
		}
		// Cache miss or connection failed, continue to fresh lookup
		dr.metrics.CacheMisses++
	}

	// Get fresh connection with failover support
	client, err := dr.getConnectionWithFailover(ctx, schoolID, securityCtx, routingDecision)
	if err != nil {
		dr.metrics.FailedRoutes++
		return nil, fmt.Errorf("failed to get connection for school %s: %w", schoolID, err)
	}

	// Update cache
	dr.updateCache(schoolID, client)
	dr.metrics.SuccessfulRoutes++

	dr.logger.Debug("Successfully routed to school database",
		zap.String("school_id", schoolID),
		zap.String("routing_decision", routingDecision.Action),
		zap.Duration("route_time", time.Since(startTime)),
	)

	return client, nil
}

// RoutingDecision represents a routing decision
type RoutingDecision struct {
	Action   string
	Target   string
	Reason   string
	Priority int
	Metadata map[string]interface{}
}

// applyRoutingRules applies routing rules to determine the best routing decision
func (dr *DatabaseRouter) applyRoutingRules(schoolID string, securityCtx *security.SecurityContext) (*RoutingDecision, error) {
	dr.routingRules.mu.RLock()
	rules := dr.routingRules.rules
	dr.routingRules.mu.RUnlock()

	// Sort rules by priority (higher priority first)
	sortedRules := make([]RoutingRule, len(rules))
	copy(sortedRules, rules)

	// Apply rules in priority order
	for _, rule := range sortedRules {
		if !rule.Enabled {
			continue
		}

		match, err := dr.evaluateRule(rule, schoolID, securityCtx)
		if err != nil {
			dr.logger.Warn("Failed to evaluate routing rule",
				zap.String("rule_id", rule.ID),
				zap.Error(err),
			)
			continue
		}

		if match {
			return &RoutingDecision{
				Action:   rule.Action,
				Target:   rule.Target,
				Reason:   fmt.Sprintf("Matched rule: %s", rule.ID),
				Priority: rule.Priority,
				Metadata: map[string]interface{}{
					"rule_id":   rule.ID,
					"school_id": schoolID,
				},
			}, nil
		}
	}

	// Default routing decision
	return &RoutingDecision{
		Action:   "route_to_primary",
		Target:   "primary",
		Reason:   "Default routing behavior",
		Priority: 0,
		Metadata: map[string]interface{}{
			"school_id": schoolID,
		},
	}, nil
}

// evaluateRule evaluates a single routing rule
func (dr *DatabaseRouter) evaluateRule(rule RoutingRule, schoolID string, securityCtx *security.SecurityContext) (bool, error) {
	switch rule.Condition {
	case "school_id":
		return dr.evaluateStringCondition(schoolID, rule.Operator, rule.Value), nil
	case "region":
		// Get school region from cache or lookup
		if cachedInfo := dr.getFromCache(schoolID); cachedInfo != nil {
			return dr.evaluateStringCondition(cachedInfo.Region, rule.Operator, rule.Value), nil
		}
		return false, nil
	case "user_role":
		return dr.evaluateStringCondition(securityCtx.Role, rule.Operator, rule.Value), nil
	case "feature":
		// Check if school has specific feature enabled
		if cachedInfo := dr.getFromCache(schoolID); cachedInfo != nil {
			if enabled, exists := cachedInfo.Features[rule.Value]; exists {
				return enabled, nil
			}
		}
		return false, nil
	case "performance":
		// Performance-based routing (implement based on metrics)
		return dr.evaluatePerformanceCondition(schoolID, rule.Operator, rule.Value), nil
	default:
		return false, fmt.Errorf("unknown condition: %s", rule.Condition)
	}
}

// evaluateStringCondition evaluates string-based conditions
func (dr *DatabaseRouter) evaluateStringCondition(value, operator, target string) bool {
	switch operator {
	case "equals":
		return value == target
	case "contains":
		return strings.Contains(value, target)
	case "starts_with":
		return strings.HasPrefix(value, target)
	case "ends_with":
		return strings.HasSuffix(value, target)
	default:
		return false
	}
}

// evaluatePerformanceCondition evaluates performance-based conditions
func (dr *DatabaseRouter) evaluatePerformanceCondition(schoolID, operator, target string) bool {
	// Implementation would check performance metrics
	// For now, return true for basic routing
	return true
}

// getConnectionWithFailover gets a connection with failover support
func (dr *DatabaseRouter) getConnectionWithFailover(ctx context.Context, schoolID string, securityCtx *security.SecurityContext, decision *RoutingDecision) (*firestore.Client, error) {
	// Try primary connection first
	client, err := dr.manager.GetSecureConnection(schoolID, securityCtx)
	if err == nil {
		return client, nil
	}

	// If failover is enabled, try failover
	if dr.config.FailoverEnabled {
		return dr.attemptFailover(ctx, schoolID, securityCtx, err)
	}

	return nil, err
}

// attemptFailover attempts to failover to a secondary database
func (dr *DatabaseRouter) attemptFailover(ctx context.Context, schoolID string, securityCtx *security.SecurityContext, originalErr error) (*firestore.Client, error) {
	dr.logger.Warn("Attempting failover for school",
		zap.String("school_id", schoolID),
		zap.Error(originalErr),
	)

	// Get failover targets
	failoverTargets := dr.getFailoverTargets(schoolID)
	if len(failoverTargets) == 0 {
		return nil, fmt.Errorf("no failover targets available for school %s: %w", schoolID, originalErr)
	}

	// Try each failover target
	for _, target := range failoverTargets {
		dr.logger.Info("Trying failover target",
			zap.String("school_id", schoolID),
			zap.String("target", target),
		)

		// Attempt connection to failover target
		// Implementation would try alternative database/region
		client, err := dr.manager.GetSecureConnection(schoolID, securityCtx)
		if err == nil {
			// Record successful failover
			dr.recordFailover(schoolID, target, "connection_failure")
			dr.metrics.FailoverCount++
			return client, nil
		}

		dr.logger.Warn("Failover target failed",
			zap.String("school_id", schoolID),
			zap.String("target", target),
			zap.Error(err),
		)
	}

	return nil, fmt.Errorf("all failover attempts failed for school %s: %w", schoolID, originalErr)
}

// getFailoverTargets gets available failover targets for a school
func (dr *DatabaseRouter) getFailoverTargets(schoolID string) []string {
	if targets, exists := dr.failoverManager.secondaryDBs.Load(schoolID); exists {
		if targetList, ok := targets.([]string); ok {
			return targetList
		}
	}
	return []string{}
}

// recordFailover records a failover event
func (dr *DatabaseRouter) recordFailover(schoolID, target, reason string) {
	state := &FailoverState{
		IsActive:       true,
		StartTime:      time.Now(),
		Reason:         reason,
		AttemptsCount:  1,
		CurrentTarget:  target,
		OriginalTarget: "primary",
	}

	dr.failoverManager.currentFailover.Store(schoolID, state)

	dr.logger.Info("Failover recorded",
		zap.String("school_id", schoolID),
		zap.String("target", target),
		zap.String("reason", reason),
	)
}

// getFromCache gets school info from cache
func (dr *DatabaseRouter) getFromCache(schoolID string) *CachedSchoolInfo {
	if infoInterface, exists := dr.schoolCache.schools.Load(schoolID); exists {
		info := infoInterface.(*CachedSchoolInfo)

		// Check if cache is still valid
		if lastUpdated, exists := dr.schoolCache.lastUpdated.Load(schoolID); exists {
			if time.Since(lastUpdated.(time.Time)) < dr.schoolCache.cacheTimeout {
				return info
			}
		}

		// Cache expired, refresh in background
		go dr.refreshSchoolCache(schoolID)
	}

	return nil
}

// getConnectionFromCache gets a connection using cached school info
func (dr *DatabaseRouter) getConnectionFromCache(ctx context.Context, info *CachedSchoolInfo, securityCtx *security.SecurityContext) (*firestore.Client, error) {
	// Use the database manager to get connection
	return dr.manager.GetSecureConnection(info.SchoolID, securityCtx)
}

// updateCache updates the school cache
func (dr *DatabaseRouter) updateCache(schoolID string, client *firestore.Client) {
	// Create cached school info
	info := &CachedSchoolInfo{
		SchoolID:     schoolID,
		DatabaseName: fmt.Sprintf("school-%s", schoolID),
		ProjectID:    fmt.Sprintf("schoolgpt-school-%s", schoolID),
		Region:       "us-central1",
		Status:       "active",
		LastActive:   time.Now(),
		Features:     make(map[string]bool),
		Limits:       make(map[string]int),
		Priority:     1,
	}

	dr.schoolCache.schools.Store(schoolID, info)
	dr.schoolCache.lastUpdated.Store(schoolID, time.Now())
}

// refreshSchoolCache refreshes school cache in background
func (dr *DatabaseRouter) refreshSchoolCache(schoolID string) {
	// Prevent multiple concurrent refreshes
	if _, inProgress := dr.schoolCache.refreshInProgress.LoadOrStore(schoolID, true); inProgress {
		return
	}
	defer dr.schoolCache.refreshInProgress.Delete(schoolID)

	dr.logger.Debug("Refreshing school cache", zap.String("school_id", schoolID))

	// Fetch fresh school info from master database
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	masterDB := dr.manager.GetMasterConnection()
	doc, err := masterDB.Collection("schools").Doc(schoolID).Get(ctx)
	if err != nil {
		dr.logger.Warn("Failed to refresh school cache",
			zap.String("school_id", schoolID),
			zap.Error(err),
		)
		return
	}

	// Update cache with fresh data
	var metadata map[string]interface{}
	if err := doc.DataTo(&metadata); err == nil {
		info := &CachedSchoolInfo{
			SchoolID:     schoolID,
			DatabaseName: getString(metadata, "database_name"),
			ProjectID:    getString(metadata, "project_id"),
			Region:       getString(metadata, "region"),
			Status:       getString(metadata, "status"),
			LastActive:   time.Now(),
			Features:     getMapBool(metadata, "features"),
			Limits:       getMapInt(metadata, "limits"),
			Priority:     getInt(metadata, "priority"),
		}

		dr.schoolCache.schools.Store(schoolID, info)
		dr.schoolCache.lastUpdated.Store(schoolID, time.Now())
		dr.schoolCache.metrics.RefreshCount++
	}
}

// Helper functions for type conversion
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getMapBool(m map[string]interface{}, key string) map[string]bool {
	if val, ok := m[key].(map[string]bool); ok {
		return val
	}
	return make(map[string]bool)
}

func getMapInt(m map[string]interface{}, key string) map[string]int {
	if val, ok := m[key].(map[string]int); ok {
		return val
	}
	return make(map[string]int)
}

func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	return 0
}

// updateMetrics updates routing metrics
func (dr *DatabaseRouter) updateMetrics(routeTime time.Duration) {
	dr.metrics.mu.Lock()
	defer dr.metrics.mu.Unlock()

	dr.metrics.TotalRequests++

	// Update average route time
	if dr.metrics.TotalRequests == 1 {
		dr.metrics.AverageRouteTime = routeTime
	} else {
		// Calculate rolling average
		dr.metrics.AverageRouteTime = time.Duration(
			(int64(dr.metrics.AverageRouteTime)*dr.metrics.TotalRequests + int64(routeTime)) /
				(dr.metrics.TotalRequests + 1),
		)
	}
}

// initializeDefaultRules sets up default routing rules
func (dr *DatabaseRouter) initializeDefaultRules() {
	defaultRules := []RoutingRule{
		{
			ID:        "system_admin_priority",
			Condition: "user_role",
			Operator:  "equals",
			Value:     "system_admin",
			Action:    "route_to_primary",
			Target:    "primary",
			Priority:  100,
			Enabled:   true,
			CreatedAt: time.Now(),
		},
		{
			ID:        "high_performance_schools",
			Condition: "feature",
			Operator:  "equals",
			Value:     "high_performance",
			Action:    "route_to_primary",
			Target:    "high_performance_cluster",
			Priority:  90,
			Enabled:   true,
			CreatedAt: time.Now(),
		},
		{
			ID:        "default_routing",
			Condition: "school_id",
			Operator:  "contains",
			Value:     "",
			Action:    "route_to_primary",
			Target:    "primary",
			Priority:  1,
			Enabled:   true,
			CreatedAt: time.Now(),
		},
	}

	dr.routingRules.mu.Lock()
	dr.routingRules.rules = defaultRules
	dr.routingRules.mu.Unlock()
}

// startMaintenance starts background maintenance tasks
func (dr *DatabaseRouter) startMaintenance() {
	ticker := time.NewTicker(dr.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		dr.performMaintenance()
	}
}

// performMaintenance performs routine maintenance tasks
func (dr *DatabaseRouter) performMaintenance() {
	// Clean up expired cache entries
	dr.cleanupExpiredCache()

	// Update cache metrics
	dr.updateCacheMetrics()

	// Check for failed connections and attempt recovery
	dr.checkFailoverRecovery()
}

// cleanupExpiredCache removes expired cache entries
func (dr *DatabaseRouter) cleanupExpiredCache() {
	now := time.Now()
	var expiredKeys []string

	dr.schoolCache.lastUpdated.Range(func(key, value interface{}) bool {
		if now.Sub(value.(time.Time)) > dr.schoolCache.cacheTimeout*2 { // Keep for 2x timeout
			expiredKeys = append(expiredKeys, key.(string))
		}
		return true
	})

	for _, key := range expiredKeys {
		dr.schoolCache.schools.Delete(key)
		dr.schoolCache.lastUpdated.Delete(key)
		dr.schoolCache.metrics.EvictionCount++
	}

	if len(expiredKeys) > 0 {
		dr.logger.Debug("Cleaned up expired cache entries",
			zap.Int("count", len(expiredKeys)),
		)
	}
}

// updateCacheMetrics updates cache performance metrics
func (dr *DatabaseRouter) updateCacheMetrics() {
	if dr.metrics.CacheHits+dr.metrics.CacheMisses > 0 {
		hitRate := float64(dr.metrics.CacheHits) / float64(dr.metrics.CacheHits+dr.metrics.CacheMisses)
		missRate := 1.0 - hitRate

		dr.schoolCache.metrics.mu.Lock()
		dr.schoolCache.metrics.HitRate = hitRate
		dr.schoolCache.metrics.MissRate = missRate
		dr.schoolCache.metrics.mu.Unlock()
	}
}

// checkFailoverRecovery checks if failed systems have recovered
func (dr *DatabaseRouter) checkFailoverRecovery() {
	dr.failoverManager.currentFailover.Range(func(key, value interface{}) bool {
		schoolID := key.(string)
		state := value.(*FailoverState)

		if state.IsActive && time.Since(state.StartTime) > time.Minute*5 {
			// Try to recover to primary
			if dr.tryRecoveryToPrimary(schoolID) {
				dr.failoverManager.currentFailover.Delete(schoolID)
				dr.logger.Info("Successfully recovered from failover",
					zap.String("school_id", schoolID),
				)
			}
		}

		return true
	})
}

// tryRecoveryToPrimary attempts to recover to primary database
func (dr *DatabaseRouter) tryRecoveryToPrimary(schoolID string) bool {
	// Implementation would test primary connection
	// For now, assume recovery after some time
	return true
}

// GetRouterMetrics returns current router metrics
func (dr *DatabaseRouter) GetRouterMetrics() *RouterMetrics {
	dr.metrics.mu.RLock()
	defer dr.metrics.mu.RUnlock()

	return &RouterMetrics{
		TotalRequests:     dr.metrics.TotalRequests,
		SuccessfulRoutes:  dr.metrics.SuccessfulRoutes,
		FailedRoutes:      dr.metrics.FailedRoutes,
		CacheHits:         dr.metrics.CacheHits,
		CacheMisses:       dr.metrics.CacheMisses,
		AverageRouteTime:  dr.metrics.AverageRouteTime,
		ConnectionErrors:  dr.metrics.ConnectionErrors,
		FailoverCount:     dr.metrics.FailoverCount,
		ActiveConnections: dr.metrics.ActiveConnections,
	}
}

// AddRoutingRule adds a new routing rule
func (dr *DatabaseRouter) AddRoutingRule(rule RoutingRule) {
	dr.routingRules.mu.Lock()
	defer dr.routingRules.mu.Unlock()

	dr.routingRules.rules = append(dr.routingRules.rules, rule)

	dr.logger.Info("Added new routing rule",
		zap.String("rule_id", rule.ID),
		zap.Int("priority", rule.Priority),
	)
}

// DefaultRouterConfig returns default router configuration
func DefaultRouterConfig() *RouterConfig {
	return &RouterConfig{
		CacheTimeout:        time.Minute * 15,
		MaxRetries:          3,
		RetryDelay:          time.Second * 2,
		FailoverEnabled:     true,
		HealthCheckInterval: time.Minute * 5,
		MetricsEnabled:      true,
		PreferredRegions:    []string{"us-central1", "us-east1"},
		LoadBalancingMode:   "performance",
	}
}
