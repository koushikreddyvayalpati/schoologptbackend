package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/schoolgpt/backend/internal/models"
	"github.com/schoolgpt/backend/internal/storage"
)

// SchoolSetupService handles school setup operations with separate databases
type SchoolSetupService struct {
	db            *storage.FirestoreDB
	schoolManager *storage.SchoolManager
}

// NewSchoolSetupService creates a new school setup service with SchoolManager
func NewSchoolSetupService(db *storage.FirestoreDB) *SchoolSetupService {
	// Initialize SchoolManager for multi-tenant architecture
	schoolManager, err := storage.NewSchoolManager("schoolgpt-backend", "")
	if err != nil {
		fmt.Printf("Warning: Could not initialize SchoolManager: %v\n", err)
		// Fall back to single database mode
		return &SchoolSetupService{db: db}
	}
	
	return &SchoolSetupService{
		db:            db,
		schoolManager: schoolManager,
	}
}

// CreateSchoolConfiguration creates a new school with its own database
func (s *SchoolSetupService) CreateSchoolConfiguration(ctx context.Context, req *CreateSchoolRequest) (*models.SchoolConfiguration, error) {
	// Generate unique identifiers
	schoolID := generateSchoolID(req.SchoolName)
	schoolCode := generateSchoolCode(req.SchoolName, req.Region)

	// Get basic template
	template := s.getBasicTemplate()

	// Create school configuration
	config := &models.SchoolConfiguration{
		ID:              schoolID,
		SchoolName:      req.SchoolName,
		SchoolCode:      schoolCode,
		AdminEmail:      req.AdminEmail,
		Region:          req.Region,
		EducationSystem: req.EducationSystem,
		Timezone:        req.Timezone,
		Language:        req.Language,
		Currency:        req.Currency,
		AcademicYear:    template.AcademicYear,
		Schemas:         make(map[string]models.EntitySchema),
		Features:        req.Features,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Status:          "active",
	}

	// Set up basic entity schemas
	config.Schemas["student"] = template.Schemas["student"]
	config.Schemas["teacher"] = template.Schemas["teacher"]
	config.Schemas["parent"] = template.Schemas["parent"]

	// **NEW: Use SchoolManager for separate database creation**
	if s.schoolManager != nil {
		// Create school metadata for master database
		schoolMetadata := storage.SchoolMetadata{
			SchoolID:         schoolID,
			SchoolName:       req.SchoolName,
			SchoolCode:       schoolCode,
			Region:           req.Region,
			Status:           "active",
			CreatedAt:        time.Now(),
			LastActive:       time.Now(),
			SubscriptionPlan: "basic",
			Features: map[string]bool{
				"attendance_tracking":    req.Features.AttendanceTracking,
				"grade_management":       req.Features.GradeManagement,
				"parent_communication":   req.Features.ParentCommunication,
				"assignment_tracking":    req.Features.AssignmentTracking,
				"behavior_tracking":      req.Features.BehaviorTracking,
				"ai_insights":            req.Features.AIInsights,
				"online_exams":           req.Features.OnlineExams,
				"multi_language_support": req.Features.MultiLanguageSupport,
			},
			Limits: map[string]int{
				"max_students": 1000,
				"max_teachers": 100,
				"max_classes":  50,
			},
		}

		// Create school with its own database
		if err := s.schoolManager.CreateSchool(ctx, schoolMetadata); err != nil {
			return nil, fmt.Errorf("error creating school database: %v", err)
		}

		// Get the school-specific database client
		schoolDB, err := s.schoolManager.GetSchoolDatabase(ctx, schoolID)
		if err != nil {
			return nil, fmt.Errorf("error accessing school database: %v", err)
		}

		// Save school configuration to the school's own database
		_, err = schoolDB.Collection("config").Doc("school").Set(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("error saving school configuration: %v", err)
		}
	} else {
		// Fallback: Save to shared database (old method)
		client := s.db.GetClient()
		_, err := client.Collection("school_configurations").Doc(schoolID).Set(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("error creating school configuration: %v", err)
		}
	}

	return config, nil
}

// getBasicTemplate returns a basic template for school setup
func (s *SchoolSetupService) getBasicTemplate() models.SchoolConfiguration {
	// Get education system template (default to CBSE)
	template, exists := models.EducationSystemTemplates["cbse"]
	if !exists {
		// Create a minimal template if not found
		template = models.SchoolConfiguration{
			AcademicYear: models.AcademicYear{
				StartDate: "2024-04-01",
				EndDate:   "2025-03-31",
				Terms: []models.Term{
					{
						ID:        "term1",
						Name:      "First Term",
						StartDate: "2024-04-01",
						EndDate:   "2024-09-30",
						IsActive:  false,
					},
					{
						ID:        "term2", 
						Name:      "Second Term",
						StartDate: "2024-10-01",
						EndDate:   "2025-03-31",
						IsActive:  true,
					},
				},
				WorkingDays: []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday"},
			},
			Schemas: map[string]models.EntitySchema{
				"student": models.GetDefaultStudentSchema(),
				"teacher": models.GetDefaultTeacherSchema(),
				"parent":  models.GetDefaultParentSchema(),
			},
		}
	}
	return template
}

// GetSchoolConfiguration retrieves a school's configuration
func (s *SchoolSetupService) GetSchoolConfiguration(ctx context.Context, schoolID string) (*models.SchoolConfiguration, error) {
	client := s.db.GetClient()
	doc, err := client.Collection("school_configurations").Doc(schoolID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting school configuration: %v", err)
	}

	var config models.SchoolConfiguration
	if err := doc.DataTo(&config); err != nil {
		return nil, fmt.Errorf("error parsing school configuration: %v", err)
	}

	return &config, nil
}

// UpdateSchoolSchema updates a specific entity schema for a school
func (s *SchoolSetupService) UpdateSchoolSchema(ctx context.Context, schoolID, entityType string, schema models.EntitySchema) error {
	// Get current configuration
	config, err := s.GetSchoolConfiguration(ctx, schoolID)
	if err != nil {
		return err
	}

	// Update schema
	schema.SchoolID = schoolID
	schema.UpdatedAt = time.Now()
	schema.Version = config.Schemas[entityType].Version + 1
	config.Schemas[entityType] = schema
	config.UpdatedAt = time.Now()

	// Save to database
	client := s.db.GetClient()
	_, err = client.Collection("school_configurations").Doc(schoolID).Set(ctx, config)
	if err != nil {
		return fmt.Errorf("error updating school schema: %v", err)
	}

	return nil
}

// AddCustomField adds a custom field to an entity schema
func (s *SchoolSetupService) AddCustomField(ctx context.Context, schoolID, entityType string, field models.CustomField) error {
	config, err := s.GetSchoolConfiguration(ctx, schoolID)
	if err != nil {
		return err
	}

	schema := config.Schemas[entityType]

	// Set field metadata
	field.CreatedAt = time.Now()
	field.UpdatedAt = time.Now()

	// Check if field with same key already exists
	for _, existingField := range schema.CustomFields {
		if existingField.Key == field.Key {
			return fmt.Errorf("field with key '%s' already exists", field.Key)
		}
	}

	// Add field
	schema.CustomFields = append(schema.CustomFields, field)
	return s.UpdateSchoolSchema(ctx, schoolID, entityType, schema)
}

// RemoveCustomField removes a custom field from an entity schema
func (s *SchoolSetupService) RemoveCustomField(ctx context.Context, schoolID, entityType, fieldID string) error {
	config, err := s.GetSchoolConfiguration(ctx, schoolID)
	if err != nil {
		return err
	}

	schema := config.Schemas[entityType]

	// Find and remove field
	var updatedFields []models.CustomField
	for _, field := range schema.CustomFields {
		if field.ID != fieldID {
			updatedFields = append(updatedFields, field)
		}
	}

	schema.CustomFields = updatedFields
	return s.UpdateSchoolSchema(ctx, schoolID, entityType, schema)
}

// UpdateSetupProgress updates the school setup progress
func (s *SchoolSetupService) UpdateSetupProgress(ctx context.Context, schoolID string, step int, stepName string) error {
	progress := &models.SetupProgress{
		SchoolID:    schoolID,
		CurrentStep: step,
		TotalSteps:  6, // Total setup steps
		StepName:    stepName,
		Completed:   false,
		UpdatedAt:   time.Now(),
	}

	if step >= 6 {
		progress.Completed = true
		// Update school status to active
		if err := s.updateSchoolStatus(ctx, schoolID, "active"); err != nil {
			return err
		}
	}

	client := s.db.GetClient()
	_, err := client.Collection("setup_progress").Doc(schoolID).Set(ctx, progress)
	return err
}

// GetSetupProgress retrieves the setup progress for a school
func (s *SchoolSetupService) GetSetupProgress(ctx context.Context, schoolID string) (*models.SetupProgress, error) {
	client := s.db.GetClient()
	doc, err := client.Collection("setup_progress").Doc(schoolID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting setup progress: %v", err)
	}

	var progress models.SetupProgress
	if err := doc.DataTo(&progress); err != nil {
		return nil, fmt.Errorf("error parsing setup progress: %v", err)
	}

	return &progress, nil
}

// ValidateSchemaField validates a custom field configuration
func (s *SchoolSetupService) ValidateSchemaField(field models.CustomField) error {
	// Check required fields
	if field.Name == "" {
		return fmt.Errorf("field name is required")
	}
	if field.Key == "" {
		return fmt.Errorf("field key is required")
	}
	if field.Type == "" {
		return fmt.Errorf("field type is required")
	}

	// Validate field key format (snake_case)
	if !isValidFieldKey(field.Key) {
		return fmt.Errorf("field key must be in snake_case format")
	}

	// Validate field type
	validTypes := []models.FieldType{
		models.FieldTypeText, models.FieldTypeNumber, models.FieldTypeDate,
		models.FieldTypeEmail, models.FieldTypePhone, models.FieldTypeSelect,
		models.FieldTypeBoolean, models.FieldTypeFile, models.FieldTypeTextArea,
	}

	var isValidType bool
	for _, validType := range validTypes {
		if field.Type == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		return fmt.Errorf("invalid field type: %s", field.Type)
	}

	// Validate select field options
	if field.Type == models.FieldTypeSelect && len(field.Options) == 0 {
		return fmt.Errorf("select field must have options")
	}

	return nil
}

// GetAvailableTemplates returns available schema templates for different regions and education systems
func (s *SchoolSetupService) GetAvailableTemplates() map[string]interface{} {
	return map[string]interface{}{
		"regions": map[string][]string{
			"india": {"CBSE", "ICSE", "State Board"},
			"usa":   {"K-12", "Common Core"},
			"uk":    {"National Curriculum"},
		},
		"education_systems": models.EducationSystemTemplates,
		"regional_fields":   models.RegionalSchemaTemplates,
		"default_schemas": map[string]models.EntitySchema{
			"student": models.GetDefaultStudentSchema(),
			"teacher": models.GetDefaultTeacherSchema(),
			"parent":  models.GetDefaultParentSchema(),
		},
	}
}

// Helper functions

func generateSchoolID(schoolName string) string {
	// Convert to lowercase, remove spaces and special characters
	id := strings.ToLower(schoolName)
	id = strings.ReplaceAll(id, " ", "_")
	id = strings.ReplaceAll(id, "-", "_")

	// Add timestamp for uniqueness
	timestamp := time.Now().Unix()
	return fmt.Sprintf("school_%s_%d", id, timestamp)
}

func generateSchoolCode(schoolName, region string) string {
	// Take first 3 letters of school name + region code
	nameCode := strings.ToUpper(schoolName)
	if len(nameCode) > 3 {
		nameCode = nameCode[:3]
	}

	regionCode := strings.ToUpper(region)
	if len(regionCode) > 2 {
		regionCode = regionCode[:2]
	}

	// Add random number
	timestamp := time.Now().Unix() % 1000
	return fmt.Sprintf("%s%s%03d", nameCode, regionCode, timestamp)
}

func isValidFieldKey(key string) bool {
	// Check if key is valid snake_case
	if strings.Contains(key, " ") || strings.Contains(key, "-") {
		return false
	}
	return strings.ToLower(key) == key
}

func (s *SchoolSetupService) updateSchoolStatus(ctx context.Context, schoolID, status string) error {
	client := s.db.GetClient()
	_, err := client.Collection("school_configurations").Doc(schoolID).Update(ctx, []firestore.Update{
		{Path: "status", Value: status},
		{Path: "updated_at", Value: time.Now()},
	})
	return err
}

// Request and Response types

type CreateSchoolRequest struct {
	SchoolName      string                `json:"school_name" binding:"required"`
	AdminEmail      string                `json:"admin_email" binding:"required"`
	Region          string                `json:"region" binding:"required"`
	EducationSystem string                `json:"education_system" binding:"required"`
	Timezone        string                `json:"timezone" binding:"required"`
	Language        string                `json:"language" binding:"required"`
	Currency        string                `json:"currency" binding:"required"`
	Features        models.SchoolFeatures `json:"features"`
}

type AddFieldRequest struct {
	EntityType string             `json:"entity_type" binding:"required"`
	Field      models.CustomField `json:"field" binding:"required"`
}

type UpdateSchemaRequest struct {
	EntityType string              `json:"entity_type" binding:"required"`
	Schema     models.EntitySchema `json:"schema" binding:"required"`
}

// GetSchoolsByAdmin retrieves all schools created by a specific admin using SchoolManager
func (s *SchoolSetupService) GetSchoolsByAdmin(ctx context.Context, adminEmail string) ([]models.SchoolConfiguration, error) {
	var schools []models.SchoolConfiguration

	if s.schoolManager != nil {
		// **NEW: Use master database to find schools by admin**
		// Get active schools from master database
		activeSchools, err := s.schoolManager.GetActiveSchools(ctx)
		if err != nil {
			return nil, fmt.Errorf("error fetching schools from master database: %v", err)
		}

		// Filter schools by admin email and get detailed configurations
		for _, schoolMeta := range activeSchools {
			// Get school-specific database
			schoolDB, err := s.schoolManager.GetSchoolDatabase(ctx, schoolMeta.SchoolID)
			if err != nil {
				continue // Skip this school if database is not accessible
			}

			// Get school configuration from the school's own database
			doc, err := schoolDB.Collection("config").Doc("school").Get(ctx)
			if err != nil {
				continue // Skip if config not found
			}

			var config models.SchoolConfiguration
			if err := doc.DataTo(&config); err != nil {
				continue // Skip if parsing fails
			}

			// Filter by admin email
			if config.AdminEmail == adminEmail {
				schools = append(schools, config)
			}
		}
	} else {
		// Fallback: Query shared database (old method)
		iter := s.db.GetClient().Collection("school_configurations").Documents(ctx)
		for {
			doc, err := iter.Next()
			if err != nil {
				break
			}

			var config models.SchoolConfiguration
			if err := doc.DataTo(&config); err != nil {
				continue
			}

			if config.AdminEmail == adminEmail {
				schools = append(schools, config)
			}
		}
	}

	return schools, nil
}
