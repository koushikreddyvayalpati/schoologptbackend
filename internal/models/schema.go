package models

import (
	"time"
)

// FieldType represents the type of a custom field
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeNumber   FieldType = "number"
	FieldTypeDate     FieldType = "date"
	FieldTypeEmail    FieldType = "email"
	FieldTypePhone    FieldType = "phone"
	FieldTypeSelect   FieldType = "select"
	FieldTypeBoolean  FieldType = "boolean"
	FieldTypeFile     FieldType = "file"
	FieldTypeTextArea FieldType = "textarea"
)

// CustomField represents a configurable field in the school's data model
type CustomField struct {
	ID           string      `firestore:"id" json:"id"`
	Name         string      `firestore:"name" json:"name"`         // Display name
	Key          string      `firestore:"key" json:"key"`           // Database key (snake_case)
	Type         FieldType   `firestore:"type" json:"type"`         // Field type
	Required     bool        `firestore:"required" json:"required"` // Is field required
	DefaultValue interface{} `firestore:"default_value" json:"default_value,omitempty"`
	Options      []string    `firestore:"options" json:"options,omitempty"`       // For select fields
	Validation   string      `firestore:"validation" json:"validation,omitempty"` // Regex or validation rule
	Description  string      `firestore:"description" json:"description,omitempty"`
	Category     string      `firestore:"category" json:"category"` // personal, academic, administrative, etc.
	Order        int         `firestore:"order" json:"order"`       // Display order
	CreatedAt    time.Time   `firestore:"created_at" json:"created_at"`
	UpdatedAt    time.Time   `firestore:"updated_at" json:"updated_at"`
}

// EntitySchema represents the complete schema for an entity type (Student, Teacher, etc.)
type EntitySchema struct {
	ID           string        `firestore:"id" json:"id"`
	SchoolID     string        `firestore:"school_id" json:"school_id"`
	EntityType   string        `firestore:"entity_type" json:"entity_type"`     // student, teacher, parent, class
	CoreFields   []CustomField `firestore:"core_fields" json:"core_fields"`     // Required system fields
	CustomFields []CustomField `firestore:"custom_fields" json:"custom_fields"` // School-specific fields
	Version      int           `firestore:"version" json:"version"`
	Active       bool          `firestore:"active" json:"active"`
	CreatedAt    time.Time     `firestore:"created_at" json:"created_at"`
	UpdatedAt    time.Time     `firestore:"updated_at" json:"updated_at"`
}

// SchoolConfiguration represents a school's complete data configuration
type SchoolConfiguration struct {
	ID              string                  `firestore:"id" json:"id"`
	SchoolName      string                  `firestore:"school_name" json:"school_name"`
	SchoolCode      string                  `firestore:"school_code" json:"school_code"`
	AdminEmail      string                  `firestore:"admin_email" json:"admin_email"` // Email of admin who created school
	Region          string                  `firestore:"region" json:"region"`
	EducationSystem string                  `firestore:"education_system" json:"education_system"` // CBSE, ICSE, State Board, etc.
	Timezone        string                  `firestore:"timezone" json:"timezone"`
	Language        string                  `firestore:"language" json:"language"`
	Currency        string                  `firestore:"currency" json:"currency"`
	AcademicYear    AcademicYear            `firestore:"academic_year" json:"academic_year"`
	Schemas         map[string]EntitySchema `firestore:"schemas" json:"schemas"` // entity_type -> schema
	Features        SchoolFeatures          `firestore:"features" json:"features"`
	Status          string                  `firestore:"status" json:"status"`         // setup, active, suspended
	SetupStep       int                     `firestore:"setup_step" json:"setup_step"` // Track setup progress
	CreatedAt       time.Time               `firestore:"created_at" json:"created_at"`
	UpdatedAt       time.Time               `firestore:"updated_at" json:"updated_at"`
}

// AcademicYear represents the school's academic year configuration
type AcademicYear struct {
	StartDate   string    `firestore:"start_date" json:"start_date"` // YYYY-MM-DD
	EndDate     string    `firestore:"end_date" json:"end_date"`     // YYYY-MM-DD
	Terms       []Term    `firestore:"terms" json:"terms"`           // Semesters/Terms
	Holidays    []Holiday `firestore:"holidays" json:"holidays"`
	WorkingDays []string  `firestore:"working_days" json:"working_days"` // ["monday", "tuesday", ...]
}

// Term represents a term/semester
type Term struct {
	ID        string `firestore:"id" json:"id"`
	Name      string `firestore:"name" json:"name"`
	StartDate string `firestore:"start_date" json:"start_date"`
	EndDate   string `firestore:"end_date" json:"end_date"`
	IsActive  bool   `firestore:"is_active" json:"is_active"`
}

// Holiday represents a school holiday
type Holiday struct {
	ID          string `firestore:"id" json:"id"`
	Name        string `firestore:"name" json:"name"`
	Date        string `firestore:"date" json:"date"`
	Type        string `firestore:"type" json:"type"` // national, regional, school
	Description string `firestore:"description" json:"description,omitempty"`
}

// SchoolFeatures represents enabled features for the school
type SchoolFeatures struct {
	AttendanceTracking   bool `firestore:"attendance_tracking" json:"attendance_tracking"`
	GradeManagement      bool `firestore:"grade_management" json:"grade_management"`
	AssignmentTracking   bool `firestore:"assignment_tracking" json:"assignment_tracking"`
	ParentCommunication  bool `firestore:"parent_communication" json:"parent_communication"`
	BehaviorTracking     bool `firestore:"behavior_tracking" json:"behavior_tracking"`
	FinancialManagement  bool `firestore:"financial_management" json:"financial_management"`
	TransportManagement  bool `firestore:"transport_management" json:"transport_management"`
	LibraryManagement    bool `firestore:"library_management" json:"library_management"`
	EventManagement      bool `firestore:"event_management" json:"event_management"`
	OnlineExams          bool `firestore:"online_exams" json:"online_exams"`
	AIInsights           bool `firestore:"ai_insights" json:"ai_insights"`
	MultiLanguageSupport bool `firestore:"multi_language_support" json:"multi_language_support"`
}

// SetupProgress represents the school setup progress
type SetupProgress struct {
	SchoolID    string    `firestore:"school_id" json:"school_id"`
	CurrentStep int       `firestore:"current_step" json:"current_step"`
	TotalSteps  int       `firestore:"total_steps" json:"total_steps"`
	StepName    string    `firestore:"step_name" json:"step_name"`
	Completed   bool      `firestore:"completed" json:"completed"`
	UpdatedAt   time.Time `firestore:"updated_at" json:"updated_at"`
}

// Subscription represents a school's subscription plan
type Subscription struct {
	ID                string            `firestore:"id" json:"id"`
	SchoolID          string            `firestore:"school_id" json:"school_id"`
	PlanType          string            `firestore:"plan_type" json:"plan_type"` // "basic", "professional", "enterprise"
	BillingCycle      string            `firestore:"billing_cycle" json:"billing_cycle"` // "monthly", "annual"
	Status            string            `firestore:"status" json:"status"` // "active", "suspended", "cancelled"
	StartDate         time.Time         `firestore:"start_date" json:"start_date"`
	EndDate           time.Time         `firestore:"end_date" json:"end_date"`
	NextBillingDate   time.Time         `firestore:"next_billing_date" json:"next_billing_date"`
	Amount            float64           `firestore:"amount" json:"amount"`
	Currency          string            `firestore:"currency" json:"currency"`
	StudentCount      int               `firestore:"student_count" json:"student_count"`
	TeacherCount      int               `firestore:"teacher_count" json:"teacher_count"`
	Features          map[string]bool   `firestore:"features" json:"features"`
	Limits            SubscriptionLimits `firestore:"limits" json:"limits"`
	PaymentMethod     string            `firestore:"payment_method" json:"payment_method"`
	StripeCustomerID  string            `firestore:"stripe_customer_id" json:"stripe_customer_id"`
	StripeSubscriptionID string         `firestore:"stripe_subscription_id" json:"stripe_subscription_id"`
	CreatedAt         time.Time         `firestore:"created_at" json:"created_at"`
	UpdatedAt         time.Time         `firestore:"updated_at" json:"updated_at"`
}

// SubscriptionLimits defines usage limits for different subscription tiers
type SubscriptionLimits struct {
	MaxStudents        int `firestore:"max_students" json:"max_students"`
	MaxTeachers        int `firestore:"max_teachers" json:"max_teachers"`
	MaxClasses         int `firestore:"max_classes" json:"max_classes"`
	MaxAssignments     int `firestore:"max_assignments" json:"max_assignments"`
	MaxStorageGB       int `firestore:"max_storage_gb" json:"max_storage_gb"`
	MaxAIQueries       int `firestore:"max_ai_queries" json:"max_ai_queries"`
	MaxVoiceMinutes    int `firestore:"max_voice_minutes" json:"max_voice_minutes"`
	MaxParentAccounts  int `firestore:"max_parent_accounts" json:"max_parent_accounts"`
	MaxReportsPerMonth int `firestore:"max_reports_per_month" json:"max_reports_per_month"`
}

// BillingEvent represents billing events and transactions
type BillingEvent struct {
	ID            string                 `firestore:"id" json:"id"`
	SchoolID      string                 `firestore:"school_id" json:"school_id"`
	EventType     string                 `firestore:"event_type" json:"event_type"` // "payment", "refund", "upgrade", "downgrade"
	Amount        float64                `firestore:"amount" json:"amount"`
	Currency      string                 `firestore:"currency" json:"currency"`
	Status        string                 `firestore:"status" json:"status"` // "success", "failed", "pending"
	PaymentMethod string                 `firestore:"payment_method" json:"payment_method"`
	StripeEventID string                 `firestore:"stripe_event_id" json:"stripe_event_id"`
	Description   string                 `firestore:"description" json:"description"`
	Metadata      map[string]interface{} `firestore:"metadata" json:"metadata"`
	CreatedAt     time.Time              `firestore:"created_at" json:"created_at"`
}

// UsageMetrics tracks feature usage for billing
type UsageMetrics struct {
	SchoolID         string    `firestore:"school_id" json:"school_id"`
	Month            string    `firestore:"month" json:"month"` // "2024-01"
	StudentsActive   int       `firestore:"students_active" json:"students_active"`
	TeachersActive   int       `firestore:"teachers_active" json:"teachers_active"`
	AIQueriesUsed    int       `firestore:"ai_queries_used" json:"ai_queries_used"`
	VoiceMinutesUsed int       `firestore:"voice_minutes_used" json:"voice_minutes_used"`
	StorageUsedGB    float64   `firestore:"storage_used_gb" json:"storage_used_gb"`
	AssignmentsCreated int     `firestore:"assignments_created" json:"assignments_created"`
	ReportsGenerated int       `firestore:"reports_generated" json:"reports_generated"`
	LastUpdated      time.Time `firestore:"last_updated" json:"last_updated"`
}

// GetDefaultStudentSchema returns the default student schema template
func GetDefaultStudentSchema() EntitySchema {
	return EntitySchema{
		EntityType: "student",
		CoreFields: []CustomField{
			{
				ID: "student_id", Name: "Student ID", Key: "student_id", Type: FieldTypeText,
				Required: true, Category: "administrative", Order: 1,
			},
			{
				ID: "full_name", Name: "Full Name", Key: "full_name", Type: FieldTypeText,
				Required: true, Category: "personal", Order: 2,
			},
			{
				ID: "email", Name: "Email Address", Key: "email", Type: FieldTypeEmail,
				Required: false, Category: "personal", Order: 3,
			},
			{
				ID: "grade_level", Name: "Grade/Class", Key: "grade_level", Type: FieldTypeText,
				Required: true, Category: "academic", Order: 4,
			},
			{
				ID: "parent_id", Name: "Parent/Guardian ID", Key: "parent_id", Type: FieldTypeText,
				Required: true, Category: "administrative", Order: 5,
			},
		},
		CustomFields: []CustomField{
			{
				ID: "date_of_birth", Name: "Date of Birth", Key: "date_of_birth", Type: FieldTypeDate,
				Required: false, Category: "personal", Order: 10,
			},
			{
				ID: "gender", Name: "Gender", Key: "gender", Type: FieldTypeSelect,
				Required: false, Options: []string{"Male", "Female", "Other", "Prefer not to say"},
				Category: "personal", Order: 11,
			},
			{
				ID: "phone_number", Name: "Phone Number", Key: "phone_number", Type: FieldTypePhone,
				Required: false, Category: "personal", Order: 12,
			},
			{
				ID: "address", Name: "Address", Key: "address", Type: FieldTypeTextArea,
				Required: false, Category: "personal", Order: 13,
			},
		},
		Version: 1,
		Active:  true,
	}
}

// GetDefaultTeacherSchema returns the default teacher schema template
func GetDefaultTeacherSchema() EntitySchema {
	return EntitySchema{
		EntityType: "teacher",
		CoreFields: []CustomField{
			{
				ID: "teacher_id", Name: "Teacher ID", Key: "teacher_id", Type: FieldTypeText,
				Required: true, Category: "administrative", Order: 1,
			},
			{
				ID: "full_name", Name: "Full Name", Key: "full_name", Type: FieldTypeText,
				Required: true, Category: "personal", Order: 2,
			},
			{
				ID: "email", Name: "Email Address", Key: "email", Type: FieldTypeEmail,
				Required: true, Category: "personal", Order: 3,
			},
			{
				ID: "subject", Name: "Primary Subject", Key: "subject", Type: FieldTypeText,
				Required: true, Category: "academic", Order: 4,
			},
		},
		CustomFields: []CustomField{
			{
				ID: "phone_number", Name: "Phone Number", Key: "phone_number", Type: FieldTypePhone,
				Required: false, Category: "personal", Order: 10,
			},
			{
				ID: "qualification", Name: "Qualifications", Key: "qualification", Type: FieldTypeTextArea,
				Required: false, Category: "academic", Order: 11,
			},
			{
				ID: "experience_years", Name: "Years of Experience", Key: "experience_years", Type: FieldTypeNumber,
				Required: false, Category: "academic", Order: 12,
			},
			{
				ID: "joining_date", Name: "Joining Date", Key: "joining_date", Type: FieldTypeDate,
				Required: false, Category: "administrative", Order: 13,
			},
		},
		Version: 1,
		Active:  true,
	}
}

// GetDefaultParentSchema returns the default parent schema template
func GetDefaultParentSchema() EntitySchema {
	return EntitySchema{
		EntityType: "parent",
		CoreFields: []CustomField{
			{
				ID: "parent_id", Name: "Parent ID", Key: "parent_id", Type: FieldTypeText,
				Required: true, Category: "administrative", Order: 1,
			},
			{
				ID: "full_name", Name: "Full Name", Key: "full_name", Type: FieldTypeText,
				Required: true, Category: "personal", Order: 2,
			},
			{
				ID: "email", Name: "Email Address", Key: "email", Type: FieldTypeEmail,
				Required: true, Category: "personal", Order: 3,
			},
			{
				ID: "phone_number", Name: "Phone Number", Key: "phone_number", Type: FieldTypePhone,
				Required: true, Category: "personal", Order: 4,
			},
		},
		CustomFields: []CustomField{
			{
				ID: "secondary_phone", Name: "Secondary Phone", Key: "secondary_phone", Type: FieldTypePhone,
				Required: false, Category: "personal", Order: 10,
			},
			{
				ID: "occupation", Name: "Occupation", Key: "occupation", Type: FieldTypeText,
				Required: false, Category: "personal", Order: 11,
			},
			{
				ID: "address", Name: "Address", Key: "address", Type: FieldTypeTextArea,
				Required: false, Category: "personal", Order: 12,
			},
			{
				ID: "emergency_contact", Name: "Emergency Contact", Key: "emergency_contact", Type: FieldTypePhone,
				Required: false, Category: "personal", Order: 13,
			},
		},
		Version: 1,
		Active:  true,
	}
}

// RegionalSchemaTemplates provides schema templates for different regions
var RegionalSchemaTemplates = map[string]map[string][]CustomField{
	"india": {
		"student": {
			{ID: "aadhar_number", Name: "Aadhar Number", Key: "aadhar_number", Type: FieldTypeText, Category: "government"},
			{ID: "caste_category", Name: "Caste Category", Key: "caste_category", Type: FieldTypeSelect,
				Options: []string{"General", "OBC", "SC", "ST", "Other"}, Category: "government"},
			{ID: "religion", Name: "Religion", Key: "religion", Type: FieldTypeText, Category: "personal"},
			{ID: "mother_tongue", Name: "Mother Tongue", Key: "mother_tongue", Type: FieldTypeText, Category: "personal"},
			{ID: "transportation", Name: "Transportation", Key: "transportation", Type: FieldTypeSelect,
				Options: []string{"School Bus", "Private Vehicle", "Public Transport", "Walking"}, Category: "administrative"},
		},
		"teacher": {
			{ID: "pan_number", Name: "PAN Number", Key: "pan_number", Type: FieldTypeText, Category: "government"},
			{ID: "aadhar_number", Name: "Aadhar Number", Key: "aadhar_number", Type: FieldTypeText, Category: "government"},
			{ID: "bank_account", Name: "Bank Account Number", Key: "bank_account", Type: FieldTypeText, Category: "financial"},
			{ID: "salary", Name: "Monthly Salary", Key: "salary", Type: FieldTypeNumber, Category: "financial"},
		},
	},
	"usa": {
		"student": {
			{ID: "ssn", Name: "Social Security Number", Key: "ssn", Type: FieldTypeText, Category: "government"},
			{ID: "grade_level", Name: "Grade Level", Key: "grade_level", Type: FieldTypeSelect,
				Options: []string{"K", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"}, Category: "academic"},
			{ID: "lunch_program", Name: "Free/Reduced Lunch", Key: "lunch_program", Type: FieldTypeBoolean, Category: "administrative"},
		},
		"teacher": {
			{ID: "ssn", Name: "Social Security Number", Key: "ssn", Type: FieldTypeText, Category: "government"},
			{ID: "teaching_license", Name: "Teaching License Number", Key: "teaching_license", Type: FieldTypeText, Category: "academic"},
		},
	},
}

// EducationSystemTemplates provides templates for different education systems
var EducationSystemTemplates = map[string]SchoolConfiguration{
	"cbse": {
		EducationSystem: "CBSE",
		AcademicYear: AcademicYear{
			StartDate: "2024-04-01",
			EndDate:   "2025-03-31",
			Terms: []Term{
				{ID: "term1", Name: "First Term", StartDate: "2024-04-01", EndDate: "2024-09-30"},
				{ID: "term2", Name: "Second Term", StartDate: "2024-10-01", EndDate: "2025-03-31"},
			},
			WorkingDays: []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday"},
		},
	},
	"icse": {
		EducationSystem: "ICSE",
		AcademicYear: AcademicYear{
			StartDate: "2024-06-01",
			EndDate:   "2025-04-30",
			Terms: []Term{
				{ID: "term1", Name: "First Term", StartDate: "2024-06-01", EndDate: "2024-12-31"},
				{ID: "term2", Name: "Second Term", StartDate: "2025-01-01", EndDate: "2025-04-30"},
			},
			WorkingDays: []string{"monday", "tuesday", "wednesday", "thursday", "friday"},
		},
	},
}
