package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/schoolgpt/backend/internal/services"
	"go.uber.org/zap"
)

// EducationalHandler handles educational feature endpoints
type EducationalHandler struct {
	eduService *services.EducationalFeaturesService
	logger     *zap.Logger
}

// NewEducationalHandler creates a new educational handler
func NewEducationalHandler(eduService *services.EducationalFeaturesService, logger *zap.Logger) *EducationalHandler {
	return &EducationalHandler{
		eduService: eduService,
		logger:     logger,
	}
}

// CreateSubjectRequest represents a request to create a subject
type CreateSubjectRequest struct {
	Name        string   `json:"name" binding:"required"`
	Code        string   `json:"code" binding:"required"`
	Description string   `json:"description"`
	Department  string   `json:"department"`
	Credits     int      `json:"credits"`
	TeacherID   string   `json:"teacher_id" binding:"required"`
	GradeLevels []string `json:"grade_levels"`
}

// CreateAssignmentRequest represents a request to create an assignment
type CreateAssignmentRequest struct {
	Title               string   `json:"title" binding:"required"`
	Description         string   `json:"description" binding:"required"`
	SubjectID           string   `json:"subject_id" binding:"required"`
	GradeLevels         []string `json:"grade_levels"`
	ClassIDs            []string `json:"class_ids"`
	AssignmentType      string   `json:"assignment_type"` // homework, project, essay, lab
	MaxScore            float64  `json:"max_score" binding:"required"`
	DueDate             string   `json:"due_date" binding:"required"` // ISO format
	Instructions        string   `json:"instructions"`
	Attachments         []string `json:"attachments"`
	SubmissionType      string   `json:"submission_type"` // online, paper, presentation
	AllowLateSubmission bool     `json:"allow_late_submission"`
}

// SubmitAssignmentRequest represents a request to submit an assignment
type SubmitAssignmentRequest struct {
	AssignmentID string   `json:"assignment_id" binding:"required"`
	Content      string   `json:"content" binding:"required"`
	Attachments  []string `json:"attachments"`
}

// GradeAssignmentRequest represents a request to grade an assignment
type GradeAssignmentRequest struct {
	Score    float64 `json:"score" binding:"required"`
	Feedback string  `json:"feedback"`
}

// HandleCreateSubject creates a new subject
func (eh *EducationalHandler) HandleCreateSubject(c *gin.Context) {
	var req CreateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get teacher ID from context (must be a teacher to create subjects)
	teacherID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teacher ID not found"})
		return
	}

	role, exists := c.Get("role")
	if !exists || role != "teacher" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only teachers can create subjects"})
		return
	}

	// Create subject
	subject := &services.Subject{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Department:  req.Department,
		Credits:     req.Credits,
		TeacherID:   teacherID.(string),
		GradeLevels: req.GradeLevels,
	}

	err := eh.eduService.CreateSubject(c.Request.Context(), subject)
	if err != nil {
		eh.logger.Error("Failed to create subject", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subject"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":    true,
		"message":    "Subject created successfully",
		"subject_id": subject.ID,
		"subject":    subject,
	})
}

// HandleCreateAssignment creates a new assignment
func (eh *EducationalHandler) HandleCreateAssignment(c *gin.Context) {
	var req CreateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get teacher ID from context
	teacherID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teacher ID not found"})
		return
	}

	role, exists := c.Get("role")
	if !exists || role != "teacher" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only teachers can create assignments"})
		return
	}

	// Parse due date
	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due date format. Use ISO 8601 format."})
		return
	}

	// Set default assignment type if not provided
	if req.AssignmentType == "" {
		req.AssignmentType = "homework"
	}

	// Set default submission type if not provided
	if req.SubmissionType == "" {
		req.SubmissionType = "online"
	}

	// Create assignment
	assignment := &services.Assignment{
		Title:               req.Title,
		Description:         req.Description,
		SubjectID:           req.SubjectID,
		TeacherID:           teacherID.(string),
		GradeLevels:         req.GradeLevels,
		ClassIDs:            req.ClassIDs,
		AssignmentType:      req.AssignmentType,
		MaxScore:            req.MaxScore,
		DueDate:             dueDate,
		Instructions:        req.Instructions,
		Attachments:         req.Attachments,
		SubmissionType:      req.SubmissionType,
		AllowLateSubmission: req.AllowLateSubmission,
	}

	err = eh.eduService.CreateAssignment(c.Request.Context(), assignment)
	if err != nil {
		eh.logger.Error("Failed to create assignment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create assignment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"message":       "Assignment created successfully",
		"assignment_id": assignment.ID,
		"assignment":    assignment,
	})
}

// HandleSubmitAssignment handles student assignment submission
func (eh *EducationalHandler) HandleSubmitAssignment(c *gin.Context) {
	var req SubmitAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get student ID from context
	studentID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Student ID not found"})
		return
	}

	role, exists := c.Get("role")
	if !exists || role != "student" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only students can submit assignments"})
		return
	}

	// Create submission
	submission := &services.AssignmentSubmission{
		AssignmentID: req.AssignmentID,
		StudentID:    studentID.(string),
		Content:      req.Content,
		Attachments:  req.Attachments,
	}

	err := eh.eduService.SubmitAssignment(c.Request.Context(), submission)
	if err != nil {
		eh.logger.Error("Failed to submit assignment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit assignment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"message":       "Assignment submitted successfully",
		"submission_id": submission.ID,
		"status":        submission.Status,
		"submitted_at":  submission.SubmittedAt,
	})
}

// HandleGradeAssignment handles assignment grading by teachers
func (eh *EducationalHandler) HandleGradeAssignment(c *gin.Context) {
	submissionID := c.Param("submission_id")
	if submissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Submission ID is required"})
		return
	}

	var req GradeAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get teacher ID from context
	teacherID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teacher ID not found"})
		return
	}

	role, exists := c.Get("role")
	if !exists || role != "teacher" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only teachers can grade assignments"})
		return
	}

	err := eh.eduService.GradeAssignment(c.Request.Context(), submissionID, req.Score, req.Feedback, teacherID.(string))
	if err != nil {
		eh.logger.Error("Failed to grade assignment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to grade assignment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Assignment graded successfully",
		"score":    req.Score,
		"feedback": req.Feedback,
	})
}

// HandleGetStudentGrades retrieves grades for a student
func (eh *EducationalHandler) HandleGetStudentGrades(c *gin.Context) {
	studentID := c.Param("student_id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Student ID is required"})
		return
	}

	// Check authorization
	userRole, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Role not found"})
		return
	}

	userEntityID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Entity ID not found"})
		return
	}

	// Students can only see their own grades, teachers can see their students' grades
	if userRole == "student" && userEntityID.(string) != studentID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only view your own grades"})
		return
	}

	// Get optional query parameters
	subjectID := c.Query("subject_id")
	term := c.Query("term")

	grades, err := eh.eduService.GetStudentGrades(c.Request.Context(), studentID, subjectID, term)
	if err != nil {
		eh.logger.Error("Failed to get student grades", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve grades"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"student_id": studentID,
		"grades":     grades,
		"count":      len(grades),
	})
}

// HandleGetTeacherAssignments retrieves assignments for a teacher
func (eh *EducationalHandler) HandleGetTeacherAssignments(c *gin.Context) {
	// Get teacher ID from context
	teacherID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teacher ID not found"})
		return
	}

	role, exists := c.Get("role")
	if !exists || role != "teacher" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only teachers can view their assignments"})
		return
	}

	// Get optional query parameters
	subjectID := c.Query("subject_id")
	activeOnly := c.Query("active") == "true"

	assignments, err := eh.eduService.GetAssignmentsByTeacher(c.Request.Context(), teacherID.(string), subjectID, activeOnly)
	if err != nil {
		eh.logger.Error("Failed to get teacher assignments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve assignments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"teacher_id":  teacherID,
		"assignments": assignments,
		"count":       len(assignments),
	})
}

// HandleGetAssignmentSubmissions retrieves submissions for an assignment
func (eh *EducationalHandler) HandleGetAssignmentSubmissions(c *gin.Context) {
	assignmentID := c.Param("assignment_id")
	if assignmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Assignment ID is required"})
		return
	}

	// Only teachers can view assignment submissions
	role, exists := c.Get("role")
	if !exists || role != "teacher" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only teachers can view assignment submissions"})
		return
	}

	submissions, err := eh.eduService.GetAssignmentSubmissions(c.Request.Context(), assignmentID)
	if err != nil {
		eh.logger.Error("Failed to get assignment submissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve submissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"assignment_id": assignmentID,
		"submissions":   submissions,
		"count":         len(submissions),
	})
}

// HandleGetStudentDashboard provides a comprehensive dashboard for students
func (eh *EducationalHandler) HandleGetStudentDashboard(c *gin.Context) {
	// Get student ID from context
	studentID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Student ID not found"})
		return
	}

	role, exists := c.Get("role")
	if !exists || role != "student" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only students can view their dashboard"})
		return
	}

	// Get recent grades (last 30 days)
	grades, err := eh.eduService.GetStudentGrades(c.Request.Context(), studentID.(string), "", "")
	if err != nil {
		eh.logger.Error("Failed to get student grades for dashboard", zap.Error(err))
		grades = []services.Grade{} // Continue with empty grades
	}

	// Calculate grade statistics
	totalGrades := len(grades)
	var totalPercentage float64
	gradeDistribution := make(map[string]int)

	for _, grade := range grades {
		totalPercentage += grade.Percentage
		gradeDistribution[grade.LetterGrade]++
	}

	averagePercentage := 0.0
	if totalGrades > 0 {
		averagePercentage = totalPercentage / float64(totalGrades)
	}

	// TODO: Get pending assignments, upcoming exams, etc.

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"student_id": studentID,
		"dashboard": gin.H{
			"recent_grades": grades[:min(len(grades), 10)], // Last 10 grades
			"statistics": gin.H{
				"total_grades":       totalGrades,
				"average_percentage": averagePercentage,
				"grade_distribution": gradeDistribution,
			},
			"pending_assignments": []interface{}{}, // TODO: Implement
			"upcoming_exams":      []interface{}{}, // TODO: Implement
		},
	})
}

// HandleGetTeacherDashboard provides a comprehensive dashboard for teachers
func (eh *EducationalHandler) HandleGetTeacherDashboard(c *gin.Context) {
	// Get teacher ID from context
	teacherID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teacher ID not found"})
		return
	}

	role, exists := c.Get("role")
	if !exists || role != "teacher" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only teachers can view their dashboard"})
		return
	}

	// Get teacher's assignments
	assignments, err := eh.eduService.GetAssignmentsByTeacher(c.Request.Context(), teacherID.(string), "", true)
	if err != nil {
		eh.logger.Error("Failed to get teacher assignments for dashboard", zap.Error(err))
		assignments = []services.Assignment{} // Continue with empty assignments
	}

	// Calculate assignment statistics
	totalAssignments := len(assignments)
	pendingGrading := 0
	upcomingDueDates := 0

	now := time.Now()
	for _, assignment := range assignments {
		if assignment.DueDate.After(now) && assignment.DueDate.Before(now.AddDate(0, 0, 7)) {
			upcomingDueDates++
		}
	}

	// TODO: Get pending submissions for grading

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"teacher_id": teacherID,
		"dashboard": gin.H{
			"recent_assignments": assignments[:min(len(assignments), 10)], // Last 10 assignments
			"statistics": gin.H{
				"total_assignments":  totalAssignments,
				"pending_grading":    pendingGrading,
				"upcoming_due_dates": upcomingDueDates,
			},
			"pending_submissions": []interface{}{}, // TODO: Implement
			"class_performance":   []interface{}{}, // TODO: Implement
		},
	})
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
