package services

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/schoolgpt/backend/internal/errors"
	"github.com/schoolgpt/backend/internal/security"
	"go.uber.org/zap"
)

// EducationalFeaturesService handles all educational features
type EducationalFeaturesService struct {
	db        *firestore.Client
	validator *security.SecurityValidator
	logger    *zap.Logger
}

// NewEducationalFeaturesService creates a new educational features service
func NewEducationalFeaturesService(db *firestore.Client, validator *security.SecurityValidator, logger *zap.Logger) *EducationalFeaturesService {
	return &EducationalFeaturesService{
		db:        db,
		validator: validator,
		logger:    logger,
	}
}

// Subject represents a school subject
type Subject struct {
	ID          string    `firestore:"id" json:"id"`
	Name        string    `firestore:"name" json:"name"`
	Code        string    `firestore:"code" json:"code"`
	Description string    `firestore:"description" json:"description"`
	Department  string    `firestore:"department" json:"department"`
	Credits     int       `firestore:"credits" json:"credits"`
	TeacherID   string    `firestore:"teacher_id" json:"teacher_id"`
	GradeLevels []string  `firestore:"grade_levels" json:"grade_levels"`
	IsActive    bool      `firestore:"is_active" json:"is_active"`
	CreatedAt   time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at" json:"updated_at"`
}

// Grade represents a student's grade for a subject
type Grade struct {
	ID           string    `firestore:"id" json:"id"`
	StudentID    string    `firestore:"student_id" json:"student_id"`
	SubjectID    string    `firestore:"subject_id" json:"subject_id"`
	TeacherID    string    `firestore:"teacher_id" json:"teacher_id"`
	AssignmentID string    `firestore:"assignment_id" json:"assignment_id,omitempty"`
	ExamID       string    `firestore:"exam_id" json:"exam_id,omitempty"`
	GradeType    string    `firestore:"grade_type" json:"grade_type"` // assignment, quiz, exam, project, participation
	Score        float64   `firestore:"score" json:"score"`
	MaxScore     float64   `firestore:"max_score" json:"max_score"`
	Percentage   float64   `firestore:"percentage" json:"percentage"`
	LetterGrade  string    `firestore:"letter_grade" json:"letter_grade"`
	Comments     string    `firestore:"comments" json:"comments"`
	Term         string    `firestore:"term" json:"term"`
	AcademicYear string    `firestore:"academic_year" json:"academic_year"`
	GradedAt     time.Time `firestore:"graded_at" json:"graded_at"`
	CreatedAt    time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt    time.Time `firestore:"updated_at" json:"updated_at"`
}

// Assignment represents a homework or project assignment
type Assignment struct {
	ID                  string    `firestore:"id" json:"id"`
	Title               string    `firestore:"title" json:"title"`
	Description         string    `firestore:"description" json:"description"`
	SubjectID           string    `firestore:"subject_id" json:"subject_id"`
	TeacherID           string    `firestore:"teacher_id" json:"teacher_id"`
	GradeLevels         []string  `firestore:"grade_levels" json:"grade_levels"`
	ClassIDs            []string  `firestore:"class_ids" json:"class_ids"`
	AssignmentType      string    `firestore:"assignment_type" json:"assignment_type"` // homework, project, essay, lab
	MaxScore            float64   `firestore:"max_score" json:"max_score"`
	DueDate             time.Time `firestore:"due_date" json:"due_date"`
	AssignedDate        time.Time `firestore:"assigned_date" json:"assigned_date"`
	Instructions        string    `firestore:"instructions" json:"instructions"`
	Attachments         []string  `firestore:"attachments" json:"attachments"`
	SubmissionType      string    `firestore:"submission_type" json:"submission_type"` // online, paper, presentation
	AllowLateSubmission bool      `firestore:"allow_late_submission" json:"allow_late_submission"`
	IsActive            bool      `firestore:"is_active" json:"is_active"`
	CreatedAt           time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt           time.Time `firestore:"updated_at" json:"updated_at"`
}

// AssignmentSubmission represents a student's assignment submission
type AssignmentSubmission struct {
	ID           string     `firestore:"id" json:"id"`
	AssignmentID string     `firestore:"assignment_id" json:"assignment_id"`
	StudentID    string     `firestore:"student_id" json:"student_id"`
	SubmittedAt  time.Time  `firestore:"submitted_at" json:"submitted_at"`
	Status       string     `firestore:"status" json:"status"` // submitted, late, pending, graded
	Content      string     `firestore:"content" json:"content"`
	Attachments  []string   `firestore:"attachments" json:"attachments"`
	Score        float64    `firestore:"score" json:"score"`
	Feedback     string     `firestore:"feedback" json:"feedback"`
	GradedAt     *time.Time `firestore:"graded_at" json:"graded_at,omitempty"`
	GradedBy     string     `firestore:"graded_by" json:"graded_by"`
	CreatedAt    time.Time  `firestore:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `firestore:"updated_at" json:"updated_at"`
}

// AcademicReport represents a student's academic report
type AcademicReport struct {
	ID              string                  `firestore:"id" json:"id"`
	StudentID       string                  `firestore:"student_id" json:"student_id"`
	Term            string                  `firestore:"term" json:"term"`
	AcademicYear    string                  `firestore:"academic_year" json:"academic_year"`
	SubjectGrades   map[string]SubjectGrade `firestore:"subject_grades" json:"subject_grades"`
	OverallGPA      float64                 `firestore:"overall_gpa" json:"overall_gpa"`
	TotalCredits    int                     `firestore:"total_credits" json:"total_credits"`
	Rank            int                     `firestore:"rank" json:"rank"`
	TotalStudents   int                     `firestore:"total_students" json:"total_students"`
	Attendance      AttendanceReport        `firestore:"attendance" json:"attendance"`
	TeacherComments []TeacherComment        `firestore:"teacher_comments" json:"teacher_comments"`
	GeneratedAt     time.Time               `firestore:"generated_at" json:"generated_at"`
	CreatedAt       time.Time               `firestore:"created_at" json:"created_at"`
}

// SubjectGrade represents grades for a specific subject
type SubjectGrade struct {
	SubjectName     string  `firestore:"subject_name" json:"subject_name"`
	TeacherName     string  `firestore:"teacher_name" json:"teacher_name"`
	TotalScore      float64 `firestore:"total_score" json:"total_score"`
	MaxScore        float64 `firestore:"max_score" json:"max_score"`
	Percentage      float64 `firestore:"percentage" json:"percentage"`
	LetterGrade     string  `firestore:"letter_grade" json:"letter_grade"`
	GradePoint      float64 `firestore:"grade_point" json:"grade_point"`
	Credits         int     `firestore:"credits" json:"credits"`
	AssignmentCount int     `firestore:"assignment_count" json:"assignment_count"`
	Improvement     string  `firestore:"improvement" json:"improvement"` // improved, declined, maintained
}

// AttendanceReport represents attendance summary for a report
type AttendanceReport struct {
	TotalDays      int     `firestore:"total_days" json:"total_days"`
	PresentDays    int     `firestore:"present_days" json:"present_days"`
	AbsentDays     int     `firestore:"absent_days" json:"absent_days"`
	LateDays       int     `firestore:"late_days" json:"late_days"`
	AttendanceRate float64 `firestore:"attendance_rate" json:"attendance_rate"`
}

// TeacherComment represents a teacher's comment on a student
type TeacherComment struct {
	TeacherID   string `firestore:"teacher_id" json:"teacher_id"`
	TeacherName string `firestore:"teacher_name" json:"teacher_name"`
	SubjectName string `firestore:"subject_name" json:"subject_name"`
	Comment     string `firestore:"comment" json:"comment"`
	Rating      int    `firestore:"rating" json:"rating"` // 1-5 scale
}

// CreateSubject creates a new subject
func (efs *EducationalFeaturesService) CreateSubject(ctx context.Context, subject *Subject) error {
	// Validate subject data
	if err := efs.validator.ValidatePersonName(subject.Name); err != nil {
		return errors.NewSchoolGPTError(
			errors.ErrCodeValidationFailed,
			"Invalid subject name",
			errors.CategoryValidation,
			errors.SeverityMedium,
			false,
		).WithDetails(err.Error())
	}

	// Generate ID if not provided
	if subject.ID == "" {
		subject.ID = fmt.Sprintf("subj_%d", time.Now().Unix())
	}

	// Set timestamps
	now := time.Now()
	subject.CreatedAt = now
	subject.UpdatedAt = now
	subject.IsActive = true

	// Save to database
	_, err := efs.db.Collection("subjects").Doc(subject.ID).Set(ctx, subject)
	if err != nil {
		return errors.NewSchoolGPTError(
			errors.ErrCodeDatabaseConnection,
			"Failed to create subject",
			errors.CategoryDatabase,
			errors.SeverityHigh,
			true,
		).WithDetails(err.Error())
	}

	efs.logger.Info("Subject created successfully",
		zap.String("subject_id", subject.ID),
		zap.String("subject_name", subject.Name),
		zap.String("teacher_id", subject.TeacherID),
	)

	return nil
}

// CreateAssignment creates a new assignment
func (efs *EducationalFeaturesService) CreateAssignment(ctx context.Context, assignment *Assignment) error {
	// Validate assignment data
	if err := efs.validator.ValidatePersonName(assignment.Title); err != nil {
		return errors.NewSchoolGPTError(
			errors.ErrCodeValidationFailed,
			"Invalid assignment title",
			errors.CategoryValidation,
			errors.SeverityMedium,
			false,
		).WithDetails(err.Error())
	}

	// Generate ID if not provided
	if assignment.ID == "" {
		assignment.ID = fmt.Sprintf("assgn_%d", time.Now().Unix())
	}

	// Set timestamps
	now := time.Now()
	assignment.CreatedAt = now
	assignment.UpdatedAt = now
	assignment.AssignedDate = now
	assignment.IsActive = true

	// Validate due date
	if assignment.DueDate.Before(now) {
		return errors.NewSchoolGPTError(
			errors.ErrCodeValidationFailed,
			"Due date cannot be in the past",
			errors.CategoryValidation,
			errors.SeverityMedium,
			false,
		)
	}

	// Save to database
	_, err := efs.db.Collection("assignments").Doc(assignment.ID).Set(ctx, assignment)
	if err != nil {
		return errors.NewSchoolGPTError(
			errors.ErrCodeDatabaseConnection,
			"Failed to create assignment",
			errors.CategoryDatabase,
			errors.SeverityHigh,
			true,
		).WithDetails(err.Error())
	}

	efs.logger.Info("Assignment created successfully",
		zap.String("assignment_id", assignment.ID),
		zap.String("title", assignment.Title),
		zap.String("teacher_id", assignment.TeacherID),
		zap.String("subject_id", assignment.SubjectID),
	)

	return nil
}

// SubmitAssignment handles student assignment submission
func (efs *EducationalFeaturesService) SubmitAssignment(ctx context.Context, submission *AssignmentSubmission) error {
	// Validate submission data
	if submission.AssignmentID == "" || submission.StudentID == "" {
		return errors.NewSchoolGPTError(
			errors.ErrCodeValidationFailed,
			"Assignment ID and Student ID are required",
			errors.CategoryValidation,
			errors.SeverityMedium,
			false,
		)
	}

	// Check if assignment exists and is active
	assignmentDoc, err := efs.db.Collection("assignments").Doc(submission.AssignmentID).Get(ctx)
	if err != nil {
		return errors.NewSchoolGPTError(
			errors.ErrCodeResourceNotFound,
			"Assignment not found",
			errors.CategoryDatabase,
			errors.SeverityMedium,
			false,
		)
	}

	var assignment Assignment
	if err := assignmentDoc.DataTo(&assignment); err != nil {
		return errors.NewSchoolGPTError(
			errors.ErrCodeInternalServer,
			"Failed to parse assignment data",
			errors.CategorySystem,
			errors.SeverityHigh,
			true,
		)
	}

	// Generate submission ID if not provided
	if submission.ID == "" {
		submission.ID = fmt.Sprintf("sub_%s_%s_%d", submission.AssignmentID, submission.StudentID, time.Now().Unix())
	}

	// Set timestamps and status
	now := time.Now()
	submission.CreatedAt = now
	submission.UpdatedAt = now
	submission.SubmittedAt = now

	// Determine submission status
	if now.After(assignment.DueDate) {
		if assignment.AllowLateSubmission {
			submission.Status = "late"
		} else {
			return errors.NewSchoolGPTError(
				errors.ErrCodeOperationNotAllowed,
				"Late submissions are not allowed for this assignment",
				errors.CategoryBusiness,
				errors.SeverityMedium,
				false,
			)
		}
	} else {
		submission.Status = "submitted"
	}

	// Save submission
	_, err = efs.db.Collection("assignment_submissions").Doc(submission.ID).Set(ctx, submission)
	if err != nil {
		return errors.NewSchoolGPTError(
			errors.ErrCodeDatabaseConnection,
			"Failed to submit assignment",
			errors.CategoryDatabase,
			errors.SeverityHigh,
			true,
		)
	}

	efs.logger.Info("Assignment submitted successfully",
		zap.String("submission_id", submission.ID),
		zap.String("assignment_id", submission.AssignmentID),
		zap.String("student_id", submission.StudentID),
		zap.String("status", submission.Status),
	)

	return nil
}

// GradeAssignment grades a student's assignment
func (efs *EducationalFeaturesService) GradeAssignment(ctx context.Context, submissionID string, score float64, feedback string, gradedBy string) error {
	// Get submission
	submissionDoc, err := efs.db.Collection("assignment_submissions").Doc(submissionID).Get(ctx)
	if err != nil {
		return errors.NewSchoolGPTError(
			errors.ErrCodeResourceNotFound,
			"Submission not found",
			errors.CategoryDatabase,
			errors.SeverityMedium,
			false,
		)
	}

	var submission AssignmentSubmission
	if err := submissionDoc.DataTo(&submission); err != nil {
		return errors.NewSchoolGPTError(
			errors.ErrCodeInternalServer,
			"Failed to parse submission data",
			errors.CategorySystem,
			errors.SeverityHigh,
			true,
		)
	}

	// Get assignment details for max score
	assignmentDoc, err := efs.db.Collection("assignments").Doc(submission.AssignmentID).Get(ctx)
	if err != nil {
		return errors.NewSchoolGPTError(
			errors.ErrCodeResourceNotFound,
			"Assignment not found",
			errors.CategoryDatabase,
			errors.SeverityMedium,
			false,
		)
	}

	var assignment Assignment
	if err := assignmentDoc.DataTo(&assignment); err != nil {
		return errors.NewSchoolGPTError(
			errors.ErrCodeInternalServer,
			"Failed to parse assignment data",
			errors.CategorySystem,
			errors.SeverityHigh,
			true,
		)
	}

	// Validate score
	if score < 0 || score > assignment.MaxScore {
		return errors.NewSchoolGPTError(
			errors.ErrCodeValidationFailed,
			fmt.Sprintf("Score must be between 0 and %v", assignment.MaxScore),
			errors.CategoryValidation,
			errors.SeverityMedium,
			false,
		)
	}

	// Update submission with grade
	now := time.Now()
	_, err = efs.db.Collection("assignment_submissions").Doc(submissionID).Update(ctx,
		[]firestore.Update{
			{Path: "score", Value: score},
			{Path: "feedback", Value: feedback},
			{Path: "graded_at", Value: now},
			{Path: "graded_by", Value: gradedBy},
			{Path: "status", Value: "graded"},
			{Path: "updated_at", Value: now},
		})
	if err != nil {
		return errors.NewSchoolGPTError(
			errors.ErrCodeDatabaseConnection,
			"Failed to update grade",
			errors.CategoryDatabase,
			errors.SeverityHigh,
			true,
		)
	}

	// Create grade record
	percentage := (score / assignment.MaxScore) * 100
	letterGrade := efs.calculateLetterGrade(percentage)

	grade := &Grade{
		ID:           fmt.Sprintf("grade_%s_%d", submissionID, time.Now().Unix()),
		StudentID:    submission.StudentID,
		SubjectID:    assignment.SubjectID,
		TeacherID:    assignment.TeacherID,
		AssignmentID: assignment.ID,
		GradeType:    assignment.AssignmentType,
		Score:        score,
		MaxScore:     assignment.MaxScore,
		Percentage:   percentage,
		LetterGrade:  letterGrade,
		Comments:     feedback,
		Term:         efs.getCurrentTerm(),
		AcademicYear: efs.getCurrentAcademicYear(),
		GradedAt:     now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = efs.db.Collection("grades").Doc(grade.ID).Set(ctx, grade)
	if err != nil {
		efs.logger.Error("Failed to create grade record", zap.Error(err))
		// Don't return error as the main grading was successful
	}

	efs.logger.Info("Assignment graded successfully",
		zap.String("submission_id", submissionID),
		zap.String("student_id", submission.StudentID),
		zap.Float64("score", score),
		zap.Float64("percentage", percentage),
		zap.String("letter_grade", letterGrade),
	)

	return nil
}

// GetStudentGrades retrieves all grades for a student
func (efs *EducationalFeaturesService) GetStudentGrades(ctx context.Context, studentID string, subjectID string, term string) ([]Grade, error) {
	query := efs.db.Collection("grades").Where("student_id", "==", studentID)

	if subjectID != "" {
		query = query.Where("subject_id", "==", subjectID)
	}

	if term != "" {
		query = query.Where("term", "==", term)
	}

	docs, err := query.OrderBy("graded_at", firestore.Desc).Documents(ctx).GetAll()
	if err != nil {
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeDatabaseConnection,
			"Failed to retrieve grades",
			errors.CategoryDatabase,
			errors.SeverityHigh,
			true,
		).WithDetails(err.Error())
	}

	var grades []Grade
	for _, doc := range docs {
		var grade Grade
		if err := doc.DataTo(&grade); err != nil {
			efs.logger.Error("Failed to parse grade", zap.Error(err))
			continue
		}
		grades = append(grades, grade)
	}

	return grades, nil
}

// GetAssignmentsByTeacher retrieves assignments created by a teacher
func (efs *EducationalFeaturesService) GetAssignmentsByTeacher(ctx context.Context, teacherID string, subjectID string, active bool) ([]Assignment, error) {
	query := efs.db.Collection("assignments").Where("teacher_id", "==", teacherID)

	if subjectID != "" {
		query = query.Where("subject_id", "==", subjectID)
	}

	if active {
		query = query.Where("is_active", "==", true)
	}

	docs, err := query.OrderBy("created_at", firestore.Desc).Documents(ctx).GetAll()
	if err != nil {
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeDatabaseConnection,
			"Failed to retrieve assignments",
			errors.CategoryDatabase,
			errors.SeverityHigh,
			true,
		).WithDetails(err.Error())
	}

	var assignments []Assignment
	for _, doc := range docs {
		var assignment Assignment
		if err := doc.DataTo(&assignment); err != nil {
			efs.logger.Error("Failed to parse assignment", zap.Error(err))
			continue
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

// GetAssignmentSubmissions retrieves submissions for an assignment
func (efs *EducationalFeaturesService) GetAssignmentSubmissions(ctx context.Context, assignmentID string) ([]AssignmentSubmission, error) {
	docs, err := efs.db.Collection("assignment_submissions").
		Where("assignment_id", "==", assignmentID).
		OrderBy("submitted_at", firestore.Desc).
		Documents(ctx).GetAll()
	if err != nil {
		return nil, errors.NewSchoolGPTError(
			errors.ErrCodeDatabaseConnection,
			"Failed to retrieve submissions",
			errors.CategoryDatabase,
			errors.SeverityHigh,
			true,
		).WithDetails(err.Error())
	}

	var submissions []AssignmentSubmission
	for _, doc := range docs {
		var submission AssignmentSubmission
		if err := doc.DataTo(&submission); err != nil {
			efs.logger.Error("Failed to parse submission", zap.Error(err))
			continue
		}
		submissions = append(submissions, submission)
	}

	return submissions, nil
}

// GenerateAcademicReport generates a comprehensive academic report for a student
func (efs *EducationalFeaturesService) GenerateAcademicReport(ctx context.Context, studentID string, term string, academicYear string) (*AcademicReport, error) {
	// Get all grades for the student in the specified term
	grades, err := efs.GetStudentGrades(ctx, studentID, "", term)
	if err != nil {
		return nil, err
	}

	// Group grades by subject
	subjectGrades := make(map[string][]Grade)
	for _, grade := range grades {
		subjectGrades[grade.SubjectID] = append(subjectGrades[grade.SubjectID], grade)
	}

	// Calculate subject-wise performance
	reportSubjectGrades := make(map[string]SubjectGrade)
	totalGradePoints := 0.0
	totalCredits := 0

	for subjectID, subjectGradeList := range subjectGrades {
		// Get subject details
		subjectDoc, err := efs.db.Collection("subjects").Doc(subjectID).Get(ctx)
		if err != nil {
			continue
		}

		var subject Subject
		if err := subjectDoc.DataTo(&subject); err != nil {
			continue
		}

		// Calculate subject performance
		totalScore := 0.0
		maxScore := 0.0
		assignmentCount := len(subjectGradeList)

		for _, grade := range subjectGradeList {
			totalScore += grade.Score
			maxScore += grade.MaxScore
		}

		percentage := 0.0
		if maxScore > 0 {
			percentage = (totalScore / maxScore) * 100
		}

		letterGrade := efs.calculateLetterGrade(percentage)
		gradePoint := efs.calculateGradePoint(percentage)

		reportSubjectGrades[subjectID] = SubjectGrade{
			SubjectName:     subject.Name,
			TeacherName:     "", // TODO: Get teacher name
			TotalScore:      totalScore,
			MaxScore:        maxScore,
			Percentage:      percentage,
			LetterGrade:     letterGrade,
			GradePoint:      gradePoint,
			Credits:         subject.Credits,
			AssignmentCount: assignmentCount,
			Improvement:     "maintained", // TODO: Calculate improvement
		}

		totalGradePoints += gradePoint * float64(subject.Credits)
		totalCredits += subject.Credits
	}

	// Calculate overall GPA
	overallGPA := 0.0
	if totalCredits > 0 {
		overallGPA = totalGradePoints / float64(totalCredits)
	}

	// TODO: Calculate rank (requires comparing with other students)
	rank := 0
	totalStudents := 0

	// TODO: Get attendance data
	attendanceReport := AttendanceReport{
		TotalDays:      100,
		PresentDays:    95,
		AbsentDays:     5,
		LateDays:       2,
		AttendanceRate: 95.0,
	}

	// TODO: Get teacher comments
	teacherComments := []TeacherComment{}

	report := &AcademicReport{
		ID:              fmt.Sprintf("report_%s_%s_%s", studentID, term, academicYear),
		StudentID:       studentID,
		Term:            term,
		AcademicYear:    academicYear,
		SubjectGrades:   reportSubjectGrades,
		OverallGPA:      overallGPA,
		TotalCredits:    totalCredits,
		Rank:            rank,
		TotalStudents:   totalStudents,
		Attendance:      attendanceReport,
		TeacherComments: teacherComments,
		GeneratedAt:     time.Now(),
		CreatedAt:       time.Now(),
	}

	// Save report to database
	_, err = efs.db.Collection("academic_reports").Doc(report.ID).Set(ctx, report)
	if err != nil {
		efs.logger.Error("Failed to save academic report", zap.Error(err))
		// Don't return error as report generation was successful
	}

	return report, nil
}

// Helper functions
func (efs *EducationalFeaturesService) calculateLetterGrade(percentage float64) string {
	switch {
	case percentage >= 90:
		return "A+"
	case percentage >= 85:
		return "A"
	case percentage >= 80:
		return "A-"
	case percentage >= 75:
		return "B+"
	case percentage >= 70:
		return "B"
	case percentage >= 65:
		return "B-"
	case percentage >= 60:
		return "C+"
	case percentage >= 55:
		return "C"
	case percentage >= 50:
		return "C-"
	case percentage >= 40:
		return "D"
	default:
		return "F"
	}
}

func (efs *EducationalFeaturesService) calculateGradePoint(percentage float64) float64 {
	switch {
	case percentage >= 90:
		return 4.0
	case percentage >= 85:
		return 3.7
	case percentage >= 80:
		return 3.3
	case percentage >= 75:
		return 3.0
	case percentage >= 70:
		return 2.7
	case percentage >= 65:
		return 2.3
	case percentage >= 60:
		return 2.0
	case percentage >= 55:
		return 1.7
	case percentage >= 50:
		return 1.3
	case percentage >= 40:
		return 1.0
	default:
		return 0.0
	}
}

func (efs *EducationalFeaturesService) getCurrentTerm() string {
	// TODO: Get current term from school configuration
	return "term2"
}

func (efs *EducationalFeaturesService) getCurrentAcademicYear() string {
	// TODO: Get current academic year from school configuration
	return "2024-25"
}
