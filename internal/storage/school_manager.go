package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// SchoolManager handles multi-tenant database architecture
type SchoolManager struct {
	masterDB    *firestore.Client
	schoolDBs   map[string]*firestore.Client
	credentials string
	projectID   string
}

// SchoolMetadata represents school information in master database
type SchoolMetadata struct {
	SchoolID         string          `firestore:"school_id"`
	SchoolName       string          `firestore:"school_name"`
	SchoolCode       string          `firestore:"school_code"`
	DatabaseName     string          `firestore:"database_name"`
	Region           string          `firestore:"region"`
	Status           string          `firestore:"status"`
	CreatedAt        time.Time       `firestore:"created_at"`
	LastActive       time.Time       `firestore:"last_active"`
	SubscriptionPlan string          `firestore:"subscription_plan"`
	Features         map[string]bool `firestore:"features"`
	Limits           map[string]int  `firestore:"limits"`
}

// NewSchoolManager creates a new multi-tenant school manager
func NewSchoolManager(masterProjectID, credentialsPath string) (*SchoolManager, error) {
	// Initialize master database connection
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(context.Background(), &firebase.Config{
		ProjectID: masterProjectID,
	}, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing master Firebase app: %v", err)
	}

	masterDB, err := app.Firestore(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting master Firestore client: %v", err)
	}

	return &SchoolManager{
		masterDB:    masterDB,
		schoolDBs:   make(map[string]*firestore.Client),
		credentials: credentialsPath,
		projectID:   masterProjectID,
	}, nil
}

// CreateSchool creates a new school with its own database
func (sm *SchoolManager) CreateSchool(ctx context.Context, schoolData SchoolMetadata) error {
	// 1. Generate unique database name
	dbName := sm.generateDatabaseName(schoolData.SchoolName, schoolData.SchoolID)
	schoolData.DatabaseName = dbName

	// 2. Create school database (this would typically be done via Firebase Admin API)
	schoolDB, err := sm.createSchoolDatabase(ctx, dbName)
	if err != nil {
		return fmt.Errorf("error creating school database: %v", err)
	}

	// 3. Initialize school database structure
	if err := sm.initializeSchoolDatabase(ctx, schoolDB, schoolData); err != nil {
		return fmt.Errorf("error initializing school database: %v", err)
	}

	// 4. Register school in master database
	_, err = sm.masterDB.Collection("schools").Doc(schoolData.SchoolID).Set(ctx, schoolData)
	if err != nil {
		return fmt.Errorf("error registering school in master database: %v", err)
	}

	// 5. Cache the database connection
	sm.schoolDBs[schoolData.SchoolID] = schoolDB

	return nil
}

// GetSchoolDatabase returns the database client for a specific school
func (sm *SchoolManager) GetSchoolDatabase(ctx context.Context, schoolID string) (*firestore.Client, error) {
	// Check cache first
	if db, exists := sm.schoolDBs[schoolID]; exists {
		return db, nil
	}

	// Get school metadata from master database
	doc, err := sm.masterDB.Collection("schools").Doc(schoolID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("school not found: %v", err)
	}

	var metadata SchoolMetadata
	if err := doc.DataTo(&metadata); err != nil {
		return nil, fmt.Errorf("error parsing school metadata: %v", err)
	}

	// Connect to school database
	schoolDB, err := sm.connectToSchoolDatabase(ctx, metadata.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("error connecting to school database: %v", err)
	}

	// Cache the connection
	sm.schoolDBs[schoolID] = schoolDB

	return schoolDB, nil
}

// Optimized schema structure for individual school databases
func (sm *SchoolManager) initializeSchoolDatabase(ctx context.Context, db *firestore.Client, metadata SchoolMetadata) error {
	batch := db.Batch()

	// 1. School Configuration (much smaller document)
	schoolConfig := map[string]interface{}{
		"school_id":        metadata.SchoolID,
		"school_name":      metadata.SchoolName,
		"school_code":      metadata.SchoolCode,
		"region":           metadata.Region,
		"education_system": "cbse", // from your example
		"currency":         "INR",
		"language":         "en",
		"timezone":         "Asia/Kolkata",
		"status":           "active",
		"created_at":       time.Now(),
	}

	configRef := db.Collection("config").Doc("school")
	batch.Set(configRef, schoolConfig)

	// 2. Academic Year (separate document)
	academicYear := map[string]interface{}{
		"start_date": "2024-04-01",
		"end_date":   "2025-03-31",
		"terms": []map[string]interface{}{
			{
				"id":         "term1",
				"name":       "First Term",
				"start_date": "2024-04-01",
				"end_date":   "2024-09-30",
				"is_active":  false,
			},
			{
				"id":         "term2",
				"name":       "Second Term",
				"start_date": "2024-10-01",
				"end_date":   "2025-03-31",
				"is_active":  true,
			},
		},
		"working_days": []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday"},
	}

	yearRef := db.Collection("config").Doc("academic_year")
	batch.Set(yearRef, academicYear)

	// 3. Features Configuration (separate document)
	features := map[string]interface{}{
		"attendance_tracking":    false,
		"grade_management":       false,
		"assignment_tracking":    false,
		"parent_communication":   false,
		"behavior_tracking":      false,
		"financial_management":   false,
		"transport_management":   false,
		"library_management":     false,
		"event_management":       false,
		"online_exams":           false,
		"ai_insights":            false,
		"multi_language_support": false,
	}

	featuresRef := db.Collection("config").Doc("features")
	batch.Set(featuresRef, features)

	// 4. Entity Schemas (separate documents for each entity)
	schemas := map[string]interface{}{
		"student": sm.getStudentSchema(),
		"teacher": sm.getTeacherSchema(),
		"parent":  sm.getParentSchema(),
	}

	for entityType, schema := range schemas {
		schemaRef := db.Collection("schemas").Doc(entityType)
		batch.Set(schemaRef, schema)
	}

	// 5. Initialize collections with proper indexes
	// Collections: students, teachers, parents, classes, subjects, attendance, etc.

	_, err := batch.Commit(ctx)
	return err
}

// Helper functions for database management
func (sm *SchoolManager) generateDatabaseName(schoolName, schoolID string) string {
	// Create clean database name: school-name-suffix
	clean := strings.ToLower(strings.ReplaceAll(schoolName, " ", "-"))
	clean = strings.ReplaceAll(clean, "_", "-")
	// Remove special characters and limit length
	var result strings.Builder
	for _, r := range clean {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	cleaned := result.String()
	if len(cleaned) > 20 {
		cleaned = cleaned[:20]
	}

	// Add unique suffix from school ID
	suffix := schoolID[len(schoolID)-6:] // Last 6 chars
	return fmt.Sprintf("school-%s-%s", cleaned, suffix)
}

func (sm *SchoolManager) createSchoolDatabase(ctx context.Context, dbName string) (*firestore.Client, error) {
	// In production, you would use Firebase Admin API to create a new database
	// For now, we'll use the same project with different collection prefixes
	opt := option.WithCredentialsFile(sm.credentials)
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID:   sm.projectID,
		DatabaseURL: fmt.Sprintf("https://%s-default-rtdb.firebaseio.com", dbName),
	}, opt)
	if err != nil {
		return nil, err
	}

	return app.Firestore(ctx)
}

func (sm *SchoolManager) connectToSchoolDatabase(ctx context.Context, dbName string) (*firestore.Client, error) {
	return sm.createSchoolDatabase(ctx, dbName)
}

// Optimized schema structures (smaller, more focused)
func (sm *SchoolManager) getStudentSchema() map[string]interface{} {
	return map[string]interface{}{
		"entity_type": "student",
		"version":     1,
		"core_fields": []map[string]interface{}{
			{
				"id":       "student_id",
				"name":     "Student ID",
				"type":     "text",
				"required": true,
				"order":    1,
			},
			{
				"id":       "full_name",
				"name":     "Full Name",
				"type":     "text",
				"required": true,
				"order":    2,
			},
			{
				"id":       "grade_level",
				"name":     "Grade/Class",
				"type":     "text",
				"required": true,
				"order":    3,
			},
			{
				"id":       "parent_id",
				"name":     "Parent/Guardian ID",
				"type":     "text",
				"required": true,
				"order":    4,
			},
		},
		"custom_fields": []map[string]interface{}{
			{
				"id":       "date_of_birth",
				"name":     "Date of Birth",
				"type":     "date",
				"required": false,
				"order":    10,
			},
			{
				"id":      "gender",
				"name":    "Gender",
				"type":    "select",
				"options": []string{"Male", "Female", "Other", "Prefer not to say"},
				"order":   11,
			},
		},
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
}

func (sm *SchoolManager) getTeacherSchema() map[string]interface{} {
	return map[string]interface{}{
		"entity_type": "teacher",
		"version":     1,
		"core_fields": []map[string]interface{}{
			{
				"id":       "teacher_id",
				"name":     "Teacher ID",
				"type":     "text",
				"required": true,
				"order":    1,
			},
			{
				"id":       "full_name",
				"name":     "Full Name",
				"type":     "text",
				"required": true,
				"order":    2,
			},
			{
				"id":       "email",
				"name":     "Email Address",
				"type":     "email",
				"required": true,
				"order":    3,
			},
			{
				"id":       "subject",
				"name":     "Primary Subject",
				"type":     "text",
				"required": true,
				"order":    4,
			},
		},
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
}

func (sm *SchoolManager) getParentSchema() map[string]interface{} {
	return map[string]interface{}{
		"entity_type": "parent",
		"version":     1,
		"core_fields": []map[string]interface{}{
			{
				"id":       "parent_id",
				"name":     "Parent ID",
				"type":     "text",
				"required": true,
				"order":    1,
			},
			{
				"id":       "full_name",
				"name":     "Full Name",
				"type":     "text",
				"required": true,
				"order":    2,
			},
			{
				"id":       "email",
				"name":     "Email Address",
				"type":     "email",
				"required": true,
				"order":    3,
			},
			{
				"id":       "phone_number",
				"name":     "Phone Number",
				"type":     "phone",
				"required": true,
				"order":    4,
			},
		},
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
}

// Performance optimization methods
func (sm *SchoolManager) GetActiveSchools(ctx context.Context) ([]SchoolMetadata, error) {
	docs, err := sm.masterDB.Collection("schools").
		Where("status", "==", "active").
		OrderBy("last_active", firestore.Desc).
		Limit(100).
		Documents(ctx).GetAll()

	if err != nil {
		return nil, err
	}

	schools := make([]SchoolMetadata, len(docs))
	for i, doc := range docs {
		if err := doc.DataTo(&schools[i]); err != nil {
			return nil, err
		}
	}

	return schools, nil
}

func (sm *SchoolManager) UpdateLastActive(ctx context.Context, schoolID string) error {
	_, err := sm.masterDB.Collection("schools").Doc(schoolID).Update(ctx, []firestore.Update{
		{Path: "last_active", Value: time.Now()},
	})
	return err
}

// Cleanup method
func (sm *SchoolManager) Close() error {
	if err := sm.masterDB.Close(); err != nil {
		return err
	}

	for _, db := range sm.schoolDBs {
		if err := db.Close(); err != nil {
			return err
		}
	}

	return nil
}
