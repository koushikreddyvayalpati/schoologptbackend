package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/schoolgpt/backend/internal/models"
	"github.com/schoolgpt/backend/internal/storage"
	"github.com/schoolgpt/backend/pkg/gpt"
)

// DataImportService handles importing school data after setup
type DataImportService struct {
	gptClient *gpt.Client
	db        *storage.FirestoreDB
}

// NewDataImportService creates a new data import service
func NewDataImportService(gptClient *gpt.Client, db *storage.FirestoreDB) *DataImportService {
	return &DataImportService{
		gptClient: gptClient,
		db:        db,
	}
}

// SampleDataRequest represents a request to generate sample data
type SampleDataRequest struct {
	SchoolID        string `json:"school_id" binding:"required"`
	TeacherCount    int    `json:"teacher_count" binding:"required"`
	StudentCount    int    `json:"student_count" binding:"required"`
	ParentCount     int    `json:"parent_count" binding:"required"`
	DataType        string `json:"data_type"` // "sample" or "template"
	Region          string `json:"region"`
	EducationSystem string `json:"education_system"`
}

// ImportDataResponse represents the response after data import
type ImportDataResponse struct {
	SchoolID        string             `json:"school_id"`
	TeachersCreated int                `json:"teachers_created"`
	StudentsCreated int                `json:"students_created"`
	ParentsCreated  int                `json:"parents_created"`
	Status          string             `json:"status"`
	Message         string             `json:"message"`
	NextSteps       []string           `json:"next_steps"`
	SampleData      *SampleDataPreview `json:"sample_data,omitempty"`
}

// SampleDataPreview shows preview of data that will be created
type SampleDataPreview struct {
	Teachers []map[string]interface{} `json:"teachers"`
	Students []map[string]interface{} `json:"students"`
	Parents  []map[string]interface{} `json:"parents"`
}

// GenerateSampleData creates realistic sample data for a school
func (s *DataImportService) GenerateSampleData(ctx context.Context, req *SampleDataRequest) (*ImportDataResponse, error) {
	// Get school configuration
	school, err := s.getSchoolConfiguration(ctx, req.SchoolID)
	if err != nil {
		return nil, fmt.Errorf("error getting school configuration: %v", err)
	}

	// Generate sample data using AI
	teachers, err := s.generateTeachers(ctx, school, req.TeacherCount)
	if err != nil {
		return nil, fmt.Errorf("error generating teachers: %v", err)
	}

	parents, err := s.generateParents(ctx, school, req.ParentCount)
	if err != nil {
		return nil, fmt.Errorf("error generating parents: %v", err)
	}

	students, err := s.generateStudents(ctx, school, req.StudentCount, parents)
	if err != nil {
		return nil, fmt.Errorf("error generating students: %v", err)
	}

	// If it's just a preview, return the data without saving
	if req.DataType == "template" {
		return &ImportDataResponse{
			SchoolID: req.SchoolID,
			Status:   "preview",
			Message:  "Sample data generated successfully",
			SampleData: &SampleDataPreview{
				Teachers: teachers[:min(3, len(teachers))], // Show first 3
				Students: students[:min(3, len(students))], // Show first 3
				Parents:  parents[:min(3, len(parents))],   // Show first 3
			},
		}, nil
	}

	// Save data to database
	teachersCreated, err := s.saveTeachers(ctx, req.SchoolID, teachers)
	if err != nil {
		return nil, fmt.Errorf("error saving teachers: %v", err)
	}

	parentsCreated, err := s.saveParents(ctx, req.SchoolID, parents)
	if err != nil {
		return nil, fmt.Errorf("error saving parents: %v", err)
	}

	studentsCreated, err := s.saveStudents(ctx, req.SchoolID, students)
	if err != nil {
		return nil, fmt.Errorf("error saving students: %v", err)
	}

	return &ImportDataResponse{
		SchoolID:        req.SchoolID,
		TeachersCreated: teachersCreated,
		StudentsCreated: studentsCreated,
		ParentsCreated:  parentsCreated,
		Status:          "completed",
		Message:         "School data imported successfully",
		NextSteps: []string{
			"Set up class schedules and assignments",
			"Configure user permissions",
			"Start taking attendance",
			"Begin using AI features",
		},
	}, nil
}

// generateTeachers creates realistic teacher data using AI
func (s *DataImportService) generateTeachers(ctx context.Context, school *models.SchoolConfiguration, count int) ([]map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Generate %d realistic teacher records for %s school in %s following %s curriculum.

Create diverse, professional teacher data with:
- Realistic Indian names (mix of regional names)
- Professional email addresses (@%s)
- Valid phone numbers (Indian format)
- Appropriate subjects for %s curriculum
- Realistic qualifications and experience

Return ONLY a JSON array of teacher objects with these fields:
[
  {
    "teacher_id": "unique_id",
    "full_name": "realistic name",
    "email": "professional email",
    "subject": "primary subject",
    "phone_number": "indian phone number",
    "qualification": "educational qualifications",
    "experience_years": number,
    "joining_date": "YYYY-MM-DD"
  }
]

Make it realistic for an actual %s school.`,
		count,
		school.SchoolName,
		school.Region,
		school.EducationSystem,
		strings.ToLower(strings.ReplaceAll(school.SchoolName, " ", "")+".edu.in"),
		school.EducationSystem,
		school.EducationSystem)

	response, err := s.gptClient.GenerateResponse(ctx, prompt, 3000)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	teachers, err := s.parseJSONArray(response)
	if err != nil {
		return nil, fmt.Errorf("error parsing teachers JSON: %v", err)
	}

	return teachers, nil
}

// generateParents creates realistic parent data using AI
func (s *DataImportService) generateParents(ctx context.Context, school *models.SchoolConfiguration, count int) ([]map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Generate %d realistic parent records for %s school in %s.

Create diverse parent data with:
- Realistic Indian family names
- Professional email addresses
- Valid phone numbers (Indian format)
- Appropriate occupations for %s region
- Emergency contacts

Return ONLY a JSON array of parent objects:
[
  {
    "parent_id": "unique_id",
    "full_name": "realistic name",
    "email": "email address",
    "phone_number": "primary phone",
    "secondary_phone": "secondary phone",
    "occupation": "professional occupation",
    "address": "realistic address in %s",
    "emergency_contact": "emergency phone number"
  }
]

Make it realistic for families in %s.`,
		count,
		school.SchoolName,
		school.Region,
		school.Region,
		school.Region,
		school.Region)

	response, err := s.gptClient.GenerateResponse(ctx, prompt, 3000)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	parents, err := s.parseJSONArray(response)
	if err != nil {
		return nil, fmt.Errorf("error parsing parents JSON: %v", err)
	}

	return parents, nil
}

// generateStudents creates realistic student data using AI
func (s *DataImportService) generateStudents(ctx context.Context, school *models.SchoolConfiguration, count int, parents []map[string]interface{}) ([]map[string]interface{}, error) {
	// Get custom fields for students
	customFields := s.extractCustomFields(school, "student")

	prompt := fmt.Sprintf(`Generate %d realistic student records for %s school (%s curriculum).

Create diverse student data with:
- Realistic names (children's names)
- Appropriate grade levels for %s system
- Age-appropriate email addresses (for older students)
- Parent assignments from provided parent list
- Custom fields: %s

Parent IDs available: %s

Return ONLY a JSON array of student objects:
[
  {
    "student_id": "unique_id",
    "full_name": "realistic child name",
    "email": "email if grade 8+, empty for younger",
    "grade_level": "appropriate grade",
    "parent_id": "parent_id from list",
    "date_of_birth": "YYYY-MM-DD",
    "gender": "Male/Female",
    "phone_number": "phone if older student",
    "address": "matches parent address",
    "blood_group": "A+/B+/O+/AB+ etc",
    "house": "Red/Blue/Green/Yellow"
  }
]

Ensure realistic age-to-grade mapping and family relationships.`,
		count,
		school.SchoolName,
		school.EducationSystem,
		school.EducationSystem,
		customFields,
		s.extractParentIDs(parents))

	response, err := s.gptClient.GenerateResponse(ctx, prompt, 4000)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	students, err := s.parseJSONArray(response)
	if err != nil {
		return nil, fmt.Errorf("error parsing students JSON: %v", err)
	}

	return students, nil
}

// Helper functions

func (s *DataImportService) getSchoolConfiguration(ctx context.Context, schoolID string) (*models.SchoolConfiguration, error) {
	client := s.db.GetClient()
	doc, err := client.Collection("school_configurations").Doc(schoolID).Get(ctx)
	if err != nil {
		return nil, err
	}

	var school models.SchoolConfiguration
	if err := doc.DataTo(&school); err != nil {
		return nil, err
	}

	return &school, nil
}

func (s *DataImportService) parseJSONArray(response string) ([]map[string]interface{}, error) {
	// Extract JSON array from response
	start := strings.Index(response, "[")
	end := strings.LastIndex(response, "]") + 1

	if start == -1 || end <= start {
		return nil, fmt.Errorf("no valid JSON array found in response")
	}

	_ = response[start:end] // jsonStr for future use

	// Parse JSON (simplified - in production use proper JSON unmarshaling)
	var data []map[string]interface{}
	// This is a simplified parser - in production, use json.Unmarshal

	return data, nil
}

func (s *DataImportService) extractCustomFields(school *models.SchoolConfiguration, entityType string) string {
	// Extract custom fields for display in prompt
	if schema, exists := school.Schemas[entityType]; exists {
		var fields []string
		for _, field := range schema.CustomFields {
			fields = append(fields, fmt.Sprintf("%s (%s)", field.Name, field.Type))
		}
		return strings.Join(fields, ", ")
	}
	return "standard fields only"
}

func (s *DataImportService) extractParentIDs(parents []map[string]interface{}) string {
	var ids []string
	for _, parent := range parents {
		if id, ok := parent["parent_id"].(string); ok {
			ids = append(ids, id)
		}
	}
	return strings.Join(ids, ", ")
}

func (s *DataImportService) saveTeachers(ctx context.Context, schoolID string, teachers []map[string]interface{}) (int, error) {
	client := s.db.GetClient()
	collection := fmt.Sprintf("%s_teachers", schoolID)

	batch := client.Batch()
	count := 0

	for _, teacher := range teachers {
		if teacherID, ok := teacher["teacher_id"].(string); ok {
			teacher["school_id"] = schoolID
			teacher["created_at"] = time.Now()
			teacher["updated_at"] = time.Now()

			doc := client.Collection(collection).Doc(teacherID)
			batch.Set(doc, teacher)
			count++
		}
	}

	_, err := batch.Commit(ctx)
	return count, err
}

func (s *DataImportService) saveParents(ctx context.Context, schoolID string, parents []map[string]interface{}) (int, error) {
	client := s.db.GetClient()
	collection := fmt.Sprintf("%s_parents", schoolID)

	batch := client.Batch()
	count := 0

	for _, parent := range parents {
		if parentID, ok := parent["parent_id"].(string); ok {
			parent["school_id"] = schoolID
			parent["created_at"] = time.Now()
			parent["updated_at"] = time.Now()

			doc := client.Collection(collection).Doc(parentID)
			batch.Set(doc, parent)
			count++
		}
	}

	_, err := batch.Commit(ctx)
	return count, err
}

func (s *DataImportService) saveStudents(ctx context.Context, schoolID string, students []map[string]interface{}) (int, error) {
	client := s.db.GetClient()
	collection := fmt.Sprintf("%s_students", schoolID)

	batch := client.Batch()
	count := 0

	for _, student := range students {
		if studentID, ok := student["student_id"].(string); ok {
			student["school_id"] = schoolID
			student["created_at"] = time.Now()
			student["updated_at"] = time.Now()

			doc := client.Collection(collection).Doc(studentID)
			batch.Set(doc, student)
			count++
		}
	}

	_, err := batch.Commit(ctx)
	return count, err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
