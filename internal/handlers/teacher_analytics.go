package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/schoolgpt/backend/internal/storage"
	"github.com/schoolgpt/backend/pkg/gpt"
)

// TeacherAnalyticsHandler handles advanced analytics for teachers
type TeacherAnalyticsHandler struct {
	analytics *storage.TeacherAnalytics
	gptClient *gpt.Client
}

// NewTeacherAnalyticsHandler creates a new teacher analytics handler
func NewTeacherAnalyticsHandler(db *storage.FirestoreDB, gptClient *gpt.Client) *TeacherAnalyticsHandler {
	return &TeacherAnalyticsHandler{
		analytics: storage.NewTeacherAnalytics(db),
		gptClient: gptClient,
	}
}

// GetStudentDetailedAnalysis provides comprehensive student analysis
func (h *TeacherAnalyticsHandler) GetStudentDetailedAnalysis(c *gin.Context) {
	studentID := c.Param("student_id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Student ID is required"})
		return
	}

	// Get analysis period from query param (default 30 days)
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		days = 30
	}

	// Get detailed attendance pattern
	pattern, err := h.analytics.GetDetailedAttendancePattern(c.Request.Context(), studentID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get attendance pattern"})
		return
	}

	// Get comprehensive insight
	insight, err := h.analytics.GetStudentPerformanceInsight(c.Request.Context(), studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get performance insight"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"student_id":        studentID,
		"analysis_period":   days,
		"attendance_pattern": pattern,
		"performance_insight": insight,
		"generated_at":      pattern,
	})
}

// GetClassOverview provides class-level analytics
func (h *TeacherAnalyticsHandler) GetClassOverview(c *gin.Context) {
	teacherID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teacher ID not found"})
		return
	}

	// Optional date range filter
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	var dateRange []string
	if startDate != "" && endDate != "" {
		dateRange = []string{startDate, endDate}
	}

	overviews, err := h.analytics.GetClassAttendanceOverview(c.Request.Context(), teacherID.(string), dateRange...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get class overview"})
		return
	}

	// Calculate summary stats
	totalStudents := 0
	totalCritical := 0
	totalNeedsAttention := 0
	classCount := len(overviews)

	for _, overview := range overviews {
		totalStudents += overview.TotalStudents
		totalCritical += overview.CriticalConcerns
		totalNeedsAttention += overview.NeedsAttention
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"teacher_id":   teacherID,
		"class_count":  classCount,
		"summary": gin.H{
			"total_students":     totalStudents,
			"critical_concerns":  totalCritical,
			"needs_attention":    totalNeedsAttention,
			"healthy_students":   totalStudents - totalCritical - totalNeedsAttention,
		},
		"class_overviews": overviews,
		"date_range":      dateRange,
	})
}

// GetTeacherDashboard provides comprehensive dashboard data
func (h *TeacherAnalyticsHandler) GetTeacherDashboard(c *gin.Context) {
	teacherID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teacher ID not found"})
		return
	}

	dashboard, err := h.analytics.GetTeacherDashboardData(c.Request.Context(), teacherID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"dashboard": dashboard,
	})
}

// GenerateAttendanceInsights uses AI to analyze attendance patterns
func (h *TeacherAnalyticsHandler) GenerateAttendanceInsights(c *gin.Context) {
	studentID := c.Param("student_id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Student ID is required"})
		return
	}

	// Get detailed pattern
	pattern, err := h.analytics.GetDetailedAttendancePattern(c.Request.Context(), studentID, 60) // 60 days for better analysis
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get attendance pattern"})
		return
	}

	// Generate AI insights
	insight, err := h.generateAIInsights(c, pattern)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate AI insights"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"student_id":    studentID,
		"pattern":       pattern,
		"ai_insights":   insight,
		"generated_at":  pattern,
	})
}

// GetInterventionRecommendations provides AI-powered intervention suggestions
func (h *TeacherAnalyticsHandler) GetInterventionRecommendations(c *gin.Context) {
	studentID := c.Param("student_id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Student ID is required"})
		return
	}

	teacherID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teacher ID not found"})
		return
	}

	// Get comprehensive insight
	insight, err := h.analytics.GetStudentPerformanceInsight(c.Request.Context(), studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get performance insight"})
		return
	}

	// Generate detailed recommendations
	recommendations, err := h.generateDetailedRecommendations(c, insight, teacherID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate recommendations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"student_id":       studentID,
		"risk_level":       insight.OverallRisk,
		"recommendations":  recommendations,
		"intervention_plan": insight.InterventionPlan,
	})
}

// GetClassPerformanceTrends analyzes trends across all classes
func (h *TeacherAnalyticsHandler) GetClassPerformanceTrends(c *gin.Context) {
	teacherID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Teacher ID not found"})
		return
	}

	// Get class overviews for the last 30 days
	overviews, err := h.analytics.GetClassAttendanceOverview(c.Request.Context(), teacherID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get class trends"})
		return
	}

	// Generate trend analysis
	trends := h.analyzeTrends(overviews)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"teacher_id": teacherID,
		"trends":     trends,
		"classes":    overviews,
	})
}

// Helper functions for AI insights

func (h *TeacherAnalyticsHandler) generateAIInsights(c *gin.Context, pattern *storage.AttendancePattern) (string, error) {
	prompt := `You are an AI assistant specializing in educational analytics for Indian schools.

Analyze this student's attendance pattern and provide insights:

Student: %s
Total Days: %d
Present: %d (%.1f%%)
Absent: %d
Late: %d
Consecutive Absences: %d

Absence Patterns by Day:
%v

Flags: %v

Please provide:
1. **Key Insights**: What patterns do you notice?
2. **Risk Assessment**: Is this concerning? Why?
3. **Root Cause Analysis**: What might be causing these patterns?
4. **Immediate Actions**: What should the teacher do right away?
5. **Long-term Strategy**: How to improve attendance over time?
6. **Parent Communication**: Key points to discuss with parents

Keep insights practical and culturally appropriate for Indian families.`

	formattedPrompt := fmt.Sprintf(prompt,
		pattern.StudentName,
		pattern.TotalDays,
		pattern.PresentDays,
		pattern.AttendanceRate,
		pattern.AbsentDays,
		pattern.LateDays,
		pattern.ConsecutiveAbsences,
		pattern.AbsencePattern,
		pattern.Flags,
	)

	return h.gptClient.HandleGPTQuery(c.Request.Context(), formattedPrompt)
}

func (h *TeacherAnalyticsHandler) generateDetailedRecommendations(c *gin.Context, insight *storage.StudentPerformanceInsight, teacherID string) (string, error) {
	prompt := `You are an expert educational counselor providing intervention recommendations for a student.

Student Profile:
Name: %s
Grade: %s
Attendance Score: %.1f%%
Behavior Score: %.1f%%
Academic Score: %.1f%%
Overall Risk: %s

Current Strengths: %v
Current Concerns: %v

Based on this profile, provide a comprehensive intervention plan with:

1. **Immediate Actions** (Next 1-2 weeks):
   - Specific steps for the teacher
   - Parent communication strategy
   - Student engagement activities

2. **Short-term Goals** (1-2 months):
   - Measurable objectives
   - Support strategies
   - Progress monitoring

3. **Long-term Development** (3-6 months):
   - Skill building plans
   - Habit formation strategies
   - Success milestones

4. **Resource Allocation**:
   - Additional support needed
   - Specialist referrals
   - Family involvement

5. **Success Indicators**:
   - What to measure
   - How often to review
   - When to adjust the plan

Make recommendations specific, actionable, and appropriate for Indian educational context.`

	formattedPrompt := fmt.Sprintf(prompt,
		insight.StudentName,
		insight.Grade,
		insight.AttendanceScore,
		insight.BehaviorScore,
		insight.AcademicScore,
		insight.OverallRisk,
		insight.Strengths,
		insight.Concerns,
	)

	return h.gptClient.HandleGPTQuery(c.Request.Context(), formattedPrompt)
}

func (h *TeacherAnalyticsHandler) analyzeTrends(overviews []storage.ClassAttendanceOverview) map[string]interface{} {
	if len(overviews) == 0 {
		return map[string]interface{}{
			"status": "no_data",
		}
	}

	totalStudents := 0
	totalCritical := 0
	totalNeedsAttention := 0
	var attendanceRates []float64

	for _, overview := range overviews {
		totalStudents += overview.TotalStudents
		totalCritical += overview.CriticalConcerns
		totalNeedsAttention += overview.NeedsAttention
		attendanceRates = append(attendanceRates, overview.OverallRate)
	}

	// Calculate average attendance
	avgAttendance := 0.0
	for _, rate := range attendanceRates {
		avgAttendance += rate
	}
	avgAttendance = avgAttendance / float64(len(attendanceRates))

	// Determine overall health
	healthyStudents := totalStudents - totalCritical - totalNeedsAttention
	healthPercentage := float64(healthyStudents) / float64(totalStudents) * 100

	var status string
	var priority string
	var recommendations []string

	if healthPercentage >= 90 {
		status = "excellent"
		priority = "maintain"
		recommendations = []string{
			"Continue current successful strategies",
			"Recognize and celebrate student achievements",
			"Share best practices with other teachers",
		}
	} else if healthPercentage >= 80 {
		status = "good"
		priority = "monitor"
		recommendations = []string{
			"Monitor students in 'needs attention' category",
			"Implement preventive measures for at-risk students",
			"Strengthen parent communication",
		}
	} else if healthPercentage >= 70 {
		status = "concerning"
		priority = "intervention"
		recommendations = []string{
			"Immediate intervention for critical cases",
			"Develop individual support plans",
			"Increase frequency of parent meetings",
			"Consider additional resources or support staff",
		}
	} else {
		status = "critical"
		priority = "urgent"
		recommendations = []string{
			"Emergency intervention required",
			"Administrative support needed",
			"Comprehensive review of teaching strategies",
			"Possible need for additional training or resources",
		}
	}

	return map[string]interface{}{
		"status":               status,
		"priority":             priority,
		"total_students":       totalStudents,
		"healthy_students":     healthyStudents,
		"needs_attention":      totalNeedsAttention,
		"critical_concerns":    totalCritical,
		"health_percentage":    healthPercentage,
		"average_attendance":   avgAttendance,
		"recommendations":      recommendations,
		"class_count":          len(overviews),
	}
} 