package services

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/schoolgpt/backend/internal/models"
	"github.com/schoolgpt/backend/internal/security"
	"go.uber.org/zap"
)

// BillingService handles subscription billing and management
type BillingService struct {
	db        *firestore.Client
	validator *security.SecurityValidator
	logger    *zap.Logger
}

// NewBillingService creates a new billing service
func NewBillingService(db *firestore.Client, validator *security.SecurityValidator, logger *zap.Logger) *BillingService {
	return &BillingService{
		db:        db,
		validator: validator,
		logger:    logger,
	}
}

// SubscriptionPlan represents a subscription plan configuration
type SubscriptionPlan struct {
	ID                string                      `json:"id"`
	Name              string                      `json:"name"`
	Description       string                      `json:"description"`
	MonthlyPrice      float64                     `json:"monthly_price"`
	AnnualPrice       float64                     `json:"annual_price"`
	Features          map[string]bool             `json:"features"`
	Limits            models.SubscriptionLimits   `json:"limits"`
	PopularFeatures   []string                    `json:"popular_features"`
	TargetAudience    string                      `json:"target_audience"`
	SupportLevel      string                      `json:"support_level"`
	SetupFee          float64                     `json:"setup_fee"`
	DiscountPercent   float64                     `json:"discount_percent"`
}

// GetSubscriptionPlans returns all available subscription plans
func (bs *BillingService) GetSubscriptionPlans() []SubscriptionPlan {
	return []SubscriptionPlan{
		{
			ID:           "basic",
			Name:         "Basic Plan",
			Description:  "Perfect for small schools getting started",
			MonthlyPrice: 29.99,
			AnnualPrice:  299.99, // 2 months free
			Features: map[string]bool{
				"attendance_tracking":    true,
				"grade_management":       true,
				"parent_communication":   true,
				"basic_analytics":        true,
				"assignment_tracking":    false,
				"behavior_tracking":      false,
				"ai_insights":            false,
				"voice_processing":       false,
				"advanced_analytics":     false,
				"multi_language_support": false,
				"financial_management":   false,
				"transport_management":   false,
				"library_management":     false,
				"online_exams":           false,
			},
			Limits: models.SubscriptionLimits{
				MaxStudents:        500,
				MaxTeachers:        25,
				MaxClasses:         20,
				MaxAssignments:     100,
				MaxStorageGB:       5,
				MaxAIQueries:       0,
				MaxVoiceMinutes:    0,
				MaxParentAccounts:  500,
				MaxReportsPerMonth: 10,
			},
			PopularFeatures: []string{"Student Management", "Basic Reporting", "Parent Portal"},
			TargetAudience:  "Small schools (up to 500 students)",
			SupportLevel:    "Email support",
			SetupFee:        99.99,
			DiscountPercent: 0,
		},
		{
			ID:           "professional",
			Name:         "Professional Plan",
			Description:  "Advanced features for growing schools",
			MonthlyPrice: 99.99,
			AnnualPrice:  999.99, // 2 months free
			Features: map[string]bool{
				"attendance_tracking":    true,
				"grade_management":       true,
				"parent_communication":   true,
				"basic_analytics":        true,
				"assignment_tracking":    true,
				"behavior_tracking":      true,
				"ai_insights":            true,
				"voice_processing":       true,
				"advanced_analytics":     true,
				"multi_language_support": false,
				"financial_management":   true,
				"transport_management":   true,
				"library_management":     true,
				"online_exams":           true,
			},
			Limits: models.SubscriptionLimits{
				MaxStudents:        2000,
				MaxTeachers:        100,
				MaxClasses:         50,
				MaxAssignments:     500,
				MaxStorageGB:       25,
				MaxAIQueries:       1000,
				MaxVoiceMinutes:    500,
				MaxParentAccounts:  2000,
				MaxReportsPerMonth: 50,
			},
			PopularFeatures: []string{"AI Insights", "Advanced Analytics", "Voice Processing", "Assignment Management"},
			TargetAudience:  "Medium schools (500-2000 students)",
			SupportLevel:    "Priority email + chat support",
			SetupFee:        199.99,
			DiscountPercent: 15,
		},
		{
			ID:           "enterprise",
			Name:         "Enterprise Plan",
			Description:  "Complete solution for large institutions",
			MonthlyPrice: 299.99,
			AnnualPrice:  2999.99, // 2 months free
			Features: map[string]bool{
				"attendance_tracking":    true,
				"grade_management":       true,
				"parent_communication":   true,
				"basic_analytics":        true,
				"assignment_tracking":    true,
				"behavior_tracking":      true,
				"ai_insights":            true,
				"voice_processing":       true,
				"advanced_analytics":     true,
				"multi_language_support": true,
				"financial_management":   true,
				"transport_management":   true,
				"library_management":     true,
				"online_exams":           true,
			},
			Limits: models.SubscriptionLimits{
				MaxStudents:        10000,
				MaxTeachers:        500,
				MaxClasses:         200,
				MaxAssignments:     -1, // unlimited
				MaxStorageGB:       100,
				MaxAIQueries:       5000,
				MaxVoiceMinutes:    2000,
				MaxParentAccounts:  10000,
				MaxReportsPerMonth: -1, // unlimited
			},
			PopularFeatures: []string{"Unlimited Features", "Multi-language", "Dedicated Support", "Custom Integration"},
			TargetAudience:  "Large schools (2000+ students)",
			SupportLevel:    "Dedicated account manager + phone support",
			SetupFee:        499.99,
			DiscountPercent: 25,
		},
	}
}

// CreateSubscription creates a new subscription for a school
func (bs *BillingService) CreateSubscription(ctx context.Context, schoolID string, planType string, billingCycle string) (*models.Subscription, error) {
	plans := bs.GetSubscriptionPlans()
	var selectedPlan *SubscriptionPlan
	
	for _, plan := range plans {
		if plan.ID == planType {
			selectedPlan = &plan
			break
		}
	}
	
	if selectedPlan == nil {
		return nil, fmt.Errorf("invalid plan type: %s", planType)
	}
	
	amount := selectedPlan.MonthlyPrice
	endDate := time.Now().AddDate(0, 1, 0) // 1 month from now
	
	if billingCycle == "annual" {
		amount = selectedPlan.AnnualPrice
		endDate = time.Now().AddDate(1, 0, 0) // 1 year from now
	}
	
	subscription := &models.Subscription{
		ID:              fmt.Sprintf("sub_%s_%d", schoolID, time.Now().Unix()),
		SchoolID:        schoolID,
		PlanType:        planType,
		BillingCycle:    billingCycle,
		Status:          "active",
		StartDate:       time.Now(),
		EndDate:         endDate,
		NextBillingDate: endDate,
		Amount:          amount,
		Currency:        "USD",
		Features:        selectedPlan.Features,
		Limits:          selectedPlan.Limits,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	
	// Save to database
	_, err := bs.db.Collection("subscriptions").Doc(subscription.ID).Set(ctx, subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %v", err)
	}
	
	bs.logger.Info("Subscription created",
		zap.String("school_id", schoolID),
		zap.String("plan_type", planType),
		zap.String("billing_cycle", billingCycle),
		zap.Float64("amount", amount),
	)
	
	return subscription, nil
}

// GetSchoolSubscription retrieves a school's current subscription
func (bs *BillingService) GetSchoolSubscription(ctx context.Context, schoolID string) (*models.Subscription, error) {
	query := bs.db.Collection("subscriptions").
		Where("school_id", "==", schoolID).
		Where("status", "==", "active").
		OrderBy("created_at", firestore.Desc).
		Limit(1)
	
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to query subscription: %v", err)
	}
	
	if len(docs) == 0 {
		return nil, fmt.Errorf("no active subscription found for school %s", schoolID)
	}
	
	var subscription models.Subscription
	if err := docs[0].DataTo(&subscription); err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %v", err)
	}
	
	return &subscription, nil
}

// CheckFeatureAccess checks if a school has access to a specific feature
func (bs *BillingService) CheckFeatureAccess(ctx context.Context, schoolID string, feature string) (bool, error) {
	subscription, err := bs.GetSchoolSubscription(ctx, schoolID)
	if err != nil {
		// If no subscription found, allow basic features only
		basicFeatures := map[string]bool{
			"attendance_tracking":  true,
			"grade_management":     true,
			"parent_communication": true,
		}
		return basicFeatures[feature], nil
	}
	
	// Check if subscription is expired
	if time.Now().After(subscription.EndDate) {
		return false, fmt.Errorf("subscription expired on %s", subscription.EndDate.Format("2006-01-02"))
	}
	
	return subscription.Features[feature], nil
}

// CheckUsageLimit checks if a school has exceeded usage limits
func (bs *BillingService) CheckUsageLimit(ctx context.Context, schoolID string, limitType string, currentUsage int) (bool, error) {
	subscription, err := bs.GetSchoolSubscription(ctx, schoolID)
	if err != nil {
		// Default limits for schools without subscription
		defaultLimits := models.SubscriptionLimits{
			MaxStudents:        100,
			MaxTeachers:        10,
			MaxClasses:         5,
			MaxAssignments:     20,
			MaxStorageGB:       1,
			MaxAIQueries:       0,
			MaxVoiceMinutes:    0,
			MaxParentAccounts:  100,
			MaxReportsPerMonth: 5,
		}
		return bs.checkLimit(defaultLimits, limitType, currentUsage), nil
	}
	
	return bs.checkLimit(subscription.Limits, limitType, currentUsage), nil
}

// checkLimit helper function to check specific limits
func (bs *BillingService) checkLimit(limits models.SubscriptionLimits, limitType string, currentUsage int) bool {
	switch limitType {
	case "students":
		return limits.MaxStudents == -1 || currentUsage < limits.MaxStudents
	case "teachers":
		return limits.MaxTeachers == -1 || currentUsage < limits.MaxTeachers
	case "classes":
		return limits.MaxClasses == -1 || currentUsage < limits.MaxClasses
	case "assignments":
		return limits.MaxAssignments == -1 || currentUsage < limits.MaxAssignments
	case "ai_queries":
		return limits.MaxAIQueries == -1 || currentUsage < limits.MaxAIQueries
	case "voice_minutes":
		return limits.MaxVoiceMinutes == -1 || currentUsage < limits.MaxVoiceMinutes
	case "parent_accounts":
		return limits.MaxParentAccounts == -1 || currentUsage < limits.MaxParentAccounts
	case "reports":
		return limits.MaxReportsPerMonth == -1 || currentUsage < limits.MaxReportsPerMonth
	default:
		return false
	}
}

// RecordUsage records usage metrics for billing
func (bs *BillingService) RecordUsage(ctx context.Context, schoolID string, usageType string, amount int) error {
	month := time.Now().Format("2006-01")
	usageID := fmt.Sprintf("%s_%s", schoolID, month)
	
	// Get or create usage metrics document
	doc := bs.db.Collection("usage_metrics").Doc(usageID)
	docSnap, err := doc.Get(ctx)
	
	var usage models.UsageMetrics
	if err != nil {
		// Create new usage metrics
		usage = models.UsageMetrics{
			SchoolID:    schoolID,
			Month:       month,
			LastUpdated: time.Now(),
		}
	} else {
		if err := docSnap.DataTo(&usage); err != nil {
			return fmt.Errorf("failed to parse usage metrics: %v", err)
		}
	}
	
	// Update usage based on type
	switch usageType {
	case "ai_query":
		usage.AIQueriesUsed += amount
	case "voice_minute":
		usage.VoiceMinutesUsed += amount
	case "assignment":
		usage.AssignmentsCreated += amount
	case "report":
		usage.ReportsGenerated += amount
	case "storage":
		usage.StorageUsedGB += float64(amount) / 1024 // Convert MB to GB
	}
	
	usage.LastUpdated = time.Now()
	
	// Save updated usage
	_, err = doc.Set(ctx, usage)
	if err != nil {
		return fmt.Errorf("failed to record usage: %v", err)
	}
	
	return nil
}

// GetUsageMetrics retrieves usage metrics for a school
func (bs *BillingService) GetUsageMetrics(ctx context.Context, schoolID string, month string) (*models.UsageMetrics, error) {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	
	usageID := fmt.Sprintf("%s_%s", schoolID, month)
	doc, err := bs.db.Collection("usage_metrics").Doc(usageID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage metrics: %v", err)
	}
	
	var usage models.UsageMetrics
	if err := doc.DataTo(&usage); err != nil {
		return nil, fmt.Errorf("failed to parse usage metrics: %v", err)
	}
	
	return &usage, nil
}

// CalculateRevenue calculates potential revenue based on current usage
func (bs *BillingService) CalculateRevenue(ctx context.Context, schoolCount int, avgStudentsPerSchool int) map[string]interface{} {
	plans := bs.GetSubscriptionPlans()
	
	// Conservative distribution estimate
	basicSchools := int(float64(schoolCount) * 0.4)      // 40% basic
	professionalSchools := int(float64(schoolCount) * 0.45) // 45% professional
	enterpriseSchools := int(float64(schoolCount) * 0.15)   // 15% enterprise
	
	monthlyRevenue := float64(basicSchools)*plans[0].MonthlyPrice +
		float64(professionalSchools)*plans[1].MonthlyPrice +
		float64(enterpriseSchools)*plans[2].MonthlyPrice
	
	annualRevenue := monthlyRevenue * 12
	
	return map[string]interface{}{
		"monthly_revenue":       monthlyRevenue,
		"annual_revenue":        annualRevenue,
		"total_schools":         schoolCount,
		"avg_students_per_school": avgStudentsPerSchool,
		"plan_distribution": map[string]int{
			"basic":        basicSchools,
			"professional": professionalSchools,
			"enterprise":   enterpriseSchools,
		},
		"revenue_breakdown": map[string]float64{
			"basic":        float64(basicSchools) * plans[0].MonthlyPrice,
			"professional": float64(professionalSchools) * plans[1].MonthlyPrice,
			"enterprise":   float64(enterpriseSchools) * plans[2].MonthlyPrice,
		},
	}
} 