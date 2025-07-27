package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/schoolgpt/backend/internal/storage"
	"github.com/schoolgpt/backend/pkg/gpt"
)

// TeacherTasksHandler handles teacher-specific automation tasks
type TeacherTasksHandler struct {
	gptClient *gpt.Client
	db        *storage.FirestoreDB
}

// NewTeacherTasksHandler creates a new teacher tasks handler
func NewTeacherTasksHandler(gptClient *gpt.Client, db *storage.FirestoreDB) *TeacherTasksHandler {
	return &TeacherTasksHandler{
		gptClient: gptClient,
		db:        db,
	}
}

// MarkAttendanceRequest represents a request to mark student attendance
type MarkAttendanceRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	ClassID   string `json:"class_id" binding:"required"`
	Status    string `json:"status" binding:"required"` // present, absent, late, excused
	Reason    string `json:"reason,omitempty"`
	Date      string `json:"date,omitempty"` // optional, defaults to today
}

// StudentAnalysisRequest represents a request to analyze student performance
type StudentAnalysisRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	Subject   string `json:"subject,omitempty"`
	Days      int    `json:"days,omitempty"` // number of days to analyze, default 30
}

// ParentNotificationRequest represents a request to send parent notification
type ParentNotificationRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	Message   string `json:"message" binding:"required"`
	Type      string `json:"type"` // attendance, performance, general
	Urgent    bool   `json:"urgent,omitempty"`
}

// HandleMarkAttendance marks student attendance and provides AI insights
func (h *TeacherTasksHandler) HandleMarkAttendance(c *gin.Context) {
	var req MarkAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get teacher context
	teacherID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teacher ID not found"})
		return
	}

	// Set default date if not provided
	if req.Date == "" {
		req.Date = time.Now().Format("2006-01-02")
	}

	// Create attendance record
	record := &storage.AttendanceRecord{
		StudentID: req.StudentID,
		Date:      req.Date,
		Status:    req.Status,
		Reason:    req.Reason,
		ClassID:   req.ClassID,
		MarkedBy:  teacherID.(string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	err := h.db.SaveAttendance(c.Request.Context(), record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error saving attendance: %v", err)})
		return
	}

	// Get student info for AI analysis
	student, err := h.db.GetStudent(c.Request.Context(), req.StudentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error getting student info: %v", err)})
		return
	}

	// Generate AI insights for attendance
	insights, err := h.generateAttendanceInsights(c.Request.Context(), req.StudentID, req.Status, req.Reason)
	if err != nil {
		// Don't fail the request if insights fail
		insights = "Attendance marked successfully. AI insights temporarily unavailable."
	}

	// If student is absent, check if parent notification is needed
	var notificationSent bool
	if req.Status == "absent" {
		notificationSent = h.sendParentNotification(c.Request.Context(), req.StudentID, req.Reason, req.ClassID)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "Attendance marked successfully",
		"student":         student.Name,
		"status":          req.Status,
		"date":            req.Date,
		"insights":        insights,
		"parent_notified": notificationSent,
	})
}

// HandleStudentAnalysis provides AI-powered student performance analysis
func (h *TeacherTasksHandler) HandleStudentAnalysis(c *gin.Context) {
	var req StudentAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get teacher context
	teacherID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teacher ID not found"})
		return
	}

	// Set default analysis period
	if req.Days == 0 {
		req.Days = 30
	}

	// Get comprehensive student data
	student, err := h.db.GetStudent(c.Request.Context(), req.StudentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error getting student: %v", err)})
		return
	}

	// Get attendance summary
	attendanceSummary, err := h.db.GetAttendanceSummary(c.Request.Context(), req.StudentID, req.Days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error getting attendance: %v", err)})
		return
	}

	// Generate AI analysis
	analysis, err := h.generateStudentAnalysis(c.Request.Context(), student, attendanceSummary, teacherID.(string), req.Subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error generating analysis: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"student_name":    student.Name,
		"analysis_period": fmt.Sprintf("Last %d days", req.Days),
		"attendance_rate": attendanceSummary.AttendanceRate,
		"ai_analysis":     analysis,
		"recommendations": h.generateRecommendations(attendanceSummary.AttendanceRate),
	})
}

// HandleParentNotification sends AI-enhanced notifications to parents
func (h *TeacherTasksHandler) HandleParentNotification(c *gin.Context) {
	var req ParentNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get teacher context
	teacherID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teacher ID not found"})
		return
	}

	// Get student and teacher info
	student, err := h.db.GetStudent(c.Request.Context(), req.StudentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error getting student: %v", err)})
		return
	}

	teacher, err := h.db.GetTeacher(c.Request.Context(), teacherID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error getting teacher: %v", err)})
		return
	}

	// Enhance message with AI
	enhancedMessage, err := h.enhanceParentMessage(c.Request.Context(), req.Message, student, teacher, req.Type)
	if err != nil {
		enhancedMessage = req.Message // Fallback to original message
	}

	// In a real implementation, this would send email/SMS/WhatsApp
	// For now, we'll return the formatted message
	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"message":          "Notification prepared successfully",
		"recipient":        fmt.Sprintf("Parent of %s", student.Name),
		"from_teacher":     teacher.Name,
		"subject":          fmt.Sprintf("Update about %s", student.Name),
		"enhanced_message": enhancedMessage,
		"type":             req.Type,
		"urgent":           req.Urgent,
		"timestamp":        time.Now().Format("2006-01-02 15:04:05"),
	})
}

// generateAttendanceInsights creates AI-powered insights for attendance marking
func (h *TeacherTasksHandler) generateAttendanceInsights(ctx context.Context, studentID, status, reason string) (string, error) {
	// Get recent attendance pattern
	summary, err := h.db.GetAttendanceSummary(ctx, studentID, 30)
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`You are an AI assistant for teachers at Oakwood Elementary School. 

A student's attendance was just marked:
- Status: %s
- Reason: %s
- Student: %s
- Recent attendance rate: %.1f%%
- Total days: %d
- Absent days: %d

Please provide:
1. Brief insight about this attendance pattern
2. Any concerns or positive notes
3. Suggested next action for the teacher

Keep response concise and actionable for a busy teacher.`,
		status, reason, summary.StudentName, summary.AttendanceRate, summary.TotalDays, summary.AbsentDays)

	return h.gptClient.HandleGPTQuery(ctx, prompt)
}

// generateStudentAnalysis creates comprehensive AI analysis of student performance
func (h *TeacherTasksHandler) generateStudentAnalysis(ctx context.Context, student *storage.Student, attendance *storage.AttendanceSummary, teacherID, subject string) (string, error) {
	prompt := fmt.Sprintf(`You are an AI teaching assistant analyzing student performance.

Student: %s
Grade: %s
Attendance Rate: %.1f%% (over %d days)
Present: %d days, Absent: %d days, Late: %d days

Teacher requesting analysis: %s
Subject focus: %s

Provide a comprehensive analysis including:
1. **Attendance Patterns**: What trends do you see?
2. **Areas of Strength**: What is this student doing well?
3. **Areas for Improvement**: What needs attention?
4. **Recommendations**: Specific actions for the teacher
5. **Parent Communication**: Suggested talking points for parent meeting

Make this practical and actionable for a teacher's daily work.`,
		student.Name, student.Grade, attendance.AttendanceRate, attendance.TotalDays,
		attendance.PresentDays, attendance.AbsentDays, attendance.LateDays,
		teacherID, subject)

	return h.gptClient.HandleGPTQuery(ctx, prompt)
}

// enhanceParentMessage uses AI to create professional, empathetic parent communication
func (h *TeacherTasksHandler) enhanceParentMessage(ctx context.Context, originalMessage string, student *storage.Student, teacher *storage.Teacher, messageType string) (string, error) {
	prompt := fmt.Sprintf(`You are helping a teacher write a professional, empathetic message to a parent.

Teacher: %s (%s subject)
Student: %s
Message Type: %s
Original Message: "%s"

Please enhance this message to be:
- Professional but warm
- Clear and specific
- Action-oriented if needed
- Culturally sensitive for Indian parents
- Include next steps or how parent can help

Format as a complete email that's ready to send.`,
		teacher.Name, teacher.Subject, student.Name, messageType, originalMessage)

	return h.gptClient.HandleGPTQuery(ctx, prompt)
}

// generateRecommendations provides quick recommendations based on attendance rate
func (h *TeacherTasksHandler) generateRecommendations(attendanceRate float64) []string {
	if attendanceRate >= 95 {
		return []string{
			"Excellent attendance! Consider this student for perfect attendance recognition.",
			"Continue positive reinforcement for consistent attendance.",
		}
	} else if attendanceRate >= 90 {
		return []string{
			"Good attendance overall. Monitor for any developing patterns.",
			"Consider checking if there are specific days with more absences.",
		}
	} else if attendanceRate >= 80 {
		return []string{
			"Attendance needs attention. Schedule parent meeting to discuss.",
			"Look for patterns (specific days, subjects, times).",
			"Consider additional support or interventions.",
		}
	} else {
		return []string{
			"Critical attendance issue. Immediate intervention required.",
			"Schedule urgent parent meeting.",
			"Consider referral to school counselor or social worker.",
			"Document all communications and interventions.",
		}
	}
}

// sendParentNotification simulates sending notification to parent
func (h *TeacherTasksHandler) sendParentNotification(ctx context.Context, studentID, reason, classID string) bool {
	// In a real implementation, this would integrate with:
	// - Email service (SendGrid, AWS SES)
	// - SMS service (Twilio)
	// - WhatsApp Business API
	// - School's existing communication system

	// For now, we'll just log that notification would be sent
	fmt.Printf("📧 Parent notification would be sent for student %s (absent: %s)\n", studentID, reason)
	return true
}
