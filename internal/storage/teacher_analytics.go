package storage

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
)

// TeacherAnalytics provides optimized queries for teacher automation features
type TeacherAnalytics struct {
	db *FirestoreDB
}

// NewTeacherAnalytics creates a new teacher analytics instance
func NewTeacherAnalytics(db *FirestoreDB) *TeacherAnalytics {
	return &TeacherAnalytics{db: db}
}

// AttendancePattern represents attendance analysis data
type AttendancePattern struct {
	StudentID           string                 `json:"student_id"`
	StudentName         string                 `json:"student_name"`
	TotalDays           int                    `json:"total_days"`
	PresentDays         int                    `json:"present_days"`
	AbsentDays          int                    `json:"absent_days"`
	LateDays            int                    `json:"late_days"`
	AttendanceRate      float64                `json:"attendance_rate"`
	ConsecutiveAbsences int                    `json:"consecutive_absences"`
	AbsencePattern      map[string]int         `json:"absence_pattern"` // day of week -> count
	MonthlyTrends       []MonthlyAttendance    `json:"monthly_trends"`
	RecentAbsences      []AttendanceRecord     `json:"recent_absences"`
	Flags               []string               `json:"flags"`
}

// MonthlyAttendance represents monthly attendance data
type MonthlyAttendance struct {
	Month          string  `json:"month"`
	AttendanceRate float64 `json:"attendance_rate"`
	TotalDays      int     `json:"total_days"`
	PresentDays    int     `json:"present_days"`
}

// ClassAttendanceOverview represents class-level attendance data
type ClassAttendanceOverview struct {
	ClassID           string  `json:"class_id"`
	ClassName         string  `json:"class_name"`
	TeacherID         string  `json:"teacher_id"`
	TotalStudents     int     `json:"total_students"`
	OverallRate       float64 `json:"overall_rate"`
	AboveThreshold    int     `json:"above_threshold"`    // students with >90% attendance
	NeedsAttention    int     `json:"needs_attention"`    // students with <80% attendance
	CriticalConcerns  int     `json:"critical_concerns"`  // students with <70% attendance
	TrendDirection    string  `json:"trend_direction"`    // improving, declining, stable
}

// StudentPerformanceInsight represents comprehensive student analysis
type StudentPerformanceInsight struct {
	StudentID         string                     `json:"student_id"`
	StudentName       string                     `json:"student_name"`
	Grade             string                     `json:"grade"`
	AttendanceScore   float64                    `json:"attendance_score"`
	BehaviorScore     float64                    `json:"behavior_score"`
	AcademicScore     float64                    `json:"academic_score"`
	OverallRisk       string                     `json:"overall_risk"` // low, medium, high, critical
	Strengths         []string                   `json:"strengths"`
	Concerns          []string                   `json:"concerns"`
	Recommendations   []string                   `json:"recommendations"`
	ParentContacts    []ParentContact            `json:"parent_contacts"`
	InterventionPlan  *InterventionPlan          `json:"intervention_plan,omitempty"`
}

// ParentContact represents communication with parents
type ParentContact struct {
	Date    time.Time `json:"date"`
	Type    string    `json:"type"`    // email, phone, meeting, sms
	Reason  string    `json:"reason"`  // attendance, behavior, academic, general
	Outcome string    `json:"outcome"` // positive, neutral, follow_up_needed
	Notes   string    `json:"notes"`
}

// InterventionPlan represents a plan to help struggling students
type InterventionPlan struct {
	CreatedAt     time.Time              `json:"created_at"`
	CreatedBy     string                 `json:"created_by"`
	Priority      string                 `json:"priority"` // low, medium, high, urgent
	Goals         []string               `json:"goals"`
	Actions       []InterventionAction   `json:"actions"`
	Timeline      string                 `json:"timeline"`
	ReviewDate    time.Time              `json:"review_date"`
	Status        string                 `json:"status"` // active, completed, paused
	Progress      map[string]interface{} `json:"progress"`
}

// InterventionAction represents a specific action in an intervention plan
type InterventionAction struct {
	Description string    `json:"description"`
	Responsible string    `json:"responsible"` // teacher, counselor, parent, admin
	DueDate     time.Time `json:"due_date"`
	Status      string    `json:"status"` // pending, in_progress, completed
	Notes       string    `json:"notes"`
}

// GetDetailedAttendancePattern provides comprehensive attendance analysis for a student
func (ta *TeacherAnalytics) GetDetailedAttendancePattern(ctx context.Context, studentID string, days int) (*AttendancePattern, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Get student info
	student, err := ta.db.GetStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}

	// Use optimized query with composite index: student_id + date (DESC)
	query := ta.db.client.Collection("attendance").
		Where("student_id", "==", studentID).
		OrderBy("date", firestore.Desc)

	if days > 0 {
		query = query.Limit(days)
	}

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error fetching attendance records: %v", err)
	}

	// Analyze attendance patterns
	pattern := &AttendancePattern{
		StudentID:   studentID,
		StudentName: student.Name,
		TotalDays:   len(docs),
	}

	dayOfWeekPattern := make(map[string]int)
	monthlyData := make(map[string]*MonthlyAttendance)
	consecutiveAbsences := 0
	maxConsecutive := 0
	var recentAbsences []AttendanceRecord

	for _, doc := range docs {
		var record AttendanceRecord
		if err := doc.DataTo(&record); err != nil {
			continue
		}

		// Count by status
		switch record.Status {
		case "present":
			pattern.PresentDays++
			consecutiveAbsences = 0
		case "absent":
			pattern.AbsentDays++
			consecutiveAbsences++
			if consecutiveAbsences > maxConsecutive {
				maxConsecutive = consecutiveAbsences
			}
			if len(recentAbsences) < 5 {
				recentAbsences = append(recentAbsences, record)
			}
		case "late":
			pattern.LateDays++
			consecutiveAbsences = 0
		}

		// Analyze day of week patterns
		if record.Status == "absent" {
			date, err := time.Parse("2006-01-02", record.Date)
			if err == nil {
				dayOfWeek := date.Weekday().String()
				dayOfWeekPattern[dayOfWeek]++

				// Monthly trends
				monthKey := date.Format("2006-01")
				if monthlyData[monthKey] == nil {
					monthlyData[monthKey] = &MonthlyAttendance{
						Month: monthKey,
					}
				}
				monthlyData[monthKey].TotalDays++
				if record.Status == "present" {
					monthlyData[monthKey].PresentDays++
				}
			}
		}
	}

	// Calculate rates and trends
	if pattern.TotalDays > 0 {
		pattern.AttendanceRate = float64(pattern.PresentDays) / float64(pattern.TotalDays) * 100
	}
	pattern.ConsecutiveAbsences = maxConsecutive
	pattern.AbsencePattern = dayOfWeekPattern
	pattern.RecentAbsences = recentAbsences

	// Convert monthly data
	for _, monthly := range monthlyData {
		if monthly.TotalDays > 0 {
			monthly.AttendanceRate = float64(monthly.PresentDays) / float64(monthly.TotalDays) * 100
		}
		pattern.MonthlyTrends = append(pattern.MonthlyTrends, *monthly)
	}

	// Generate flags
	pattern.Flags = ta.generateAttendanceFlags(pattern)

	return pattern, nil
}

// GetClassAttendanceOverview provides class-level attendance analytics
func (ta *TeacherAnalytics) GetClassAttendanceOverview(ctx context.Context, teacherID string, dateRange ...string) ([]ClassAttendanceOverview, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// Get teacher's classes using optimized index
	classQuery := ta.db.client.Collection("classes").Where("teacher_id", "==", teacherID)
	classDocs, err := classQuery.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var overviews []ClassAttendanceOverview

	for _, classDoc := range classDocs {
		var class Class
		if err := classDoc.DataTo(&class); err != nil {
			continue
		}

		overview := ClassAttendanceOverview{
			ClassID:   class.ID,
			ClassName: class.Name,
			TeacherID: teacherID,
		}

		// Get students in this class
		enrollmentQuery := ta.db.client.Collection("enrollments").Where("class_id", "==", class.ID)
		enrollmentDocs, err := enrollmentQuery.Documents(ctx).GetAll()
		if err != nil {
			continue
		}

		overview.TotalStudents = len(enrollmentDocs)
		var totalRates []float64

		// Analyze each student's attendance in this class
		for _, enrollDoc := range enrollmentDocs {
			data := enrollDoc.Data()
			if studentID, ok := data["student_id"].(string); ok {
				// Use optimized query: class_id + date
				attendanceQuery := ta.db.client.Collection("attendance").
					Where("student_id", "==", studentID).
					Where("class_id", "==", class.ID)

				if len(dateRange) >= 2 {
					attendanceQuery = attendanceQuery.
						Where("date", ">=", dateRange[0]).
						Where("date", "<=", dateRange[1])
				}

				attendanceDocs, err := attendanceQuery.Documents(ctx).GetAll()
				if err != nil {
					continue
				}

				presentCount := 0
				for _, attendanceDoc := range attendanceDocs {
					var record AttendanceRecord
					if err := attendanceDoc.DataTo(&record); err == nil && record.Status == "present" {
						presentCount++
					}
				}

				if len(attendanceDocs) > 0 {
					rate := float64(presentCount) / float64(len(attendanceDocs)) * 100
					totalRates = append(totalRates, rate)

					// Categorize students
					if rate >= 90 {
						overview.AboveThreshold++
					} else if rate >= 80 {
						// Good, no action needed
					} else if rate >= 70 {
						overview.NeedsAttention++
					} else {
						overview.CriticalConcerns++
					}
				}
			}
		}

		// Calculate overall class rate
		if len(totalRates) > 0 {
			sum := 0.0
			for _, rate := range totalRates {
				sum += rate
			}
			overview.OverallRate = sum / float64(len(totalRates))
		}

		// Determine trend (simplified - could be enhanced with historical data)
		if overview.OverallRate >= 90 {
			overview.TrendDirection = "stable"
		} else if overview.OverallRate >= 80 {
			overview.TrendDirection = "stable"
		} else {
			overview.TrendDirection = "declining"
		}

		overviews = append(overviews, overview)
	}

	return overviews, nil
}

// GetStudentPerformanceInsight provides comprehensive student analysis
func (ta *TeacherAnalytics) GetStudentPerformanceInsight(ctx context.Context, studentID string) (*StudentPerformanceInsight, error) {
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	student, err := ta.db.GetStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}

	insight := &StudentPerformanceInsight{
		StudentID:   studentID,
		StudentName: student.Name,
		Grade:       student.Grade,
	}

	// Get attendance score
	pattern, err := ta.GetDetailedAttendancePattern(ctx, studentID, 30)
	if err == nil {
		insight.AttendanceScore = pattern.AttendanceRate
	}

	// Analyze behavior (placeholder - would integrate with behavior tracking)
	insight.BehaviorScore = 85.0 // Default good behavior

	// Analyze academic performance (placeholder - would integrate with grade system)
	insight.AcademicScore = 78.0 // Default average

	// Calculate overall risk
	insight.OverallRisk = ta.calculateRiskLevel(insight.AttendanceScore, insight.BehaviorScore, insight.AcademicScore)

	// Generate insights
	insight.Strengths, insight.Concerns, insight.Recommendations = ta.generateInsights(insight)

	// Get parent contact history (placeholder)
	insight.ParentContacts = ta.getRecentParentContacts(ctx, studentID)

	// Create intervention plan if needed
	if insight.OverallRisk == "high" || insight.OverallRisk == "critical" {
		insight.InterventionPlan = ta.createInterventionPlan(insight)
	}

	return insight, nil
}

// Helper functions for analytics

func (ta *TeacherAnalytics) generateAttendanceFlags(pattern *AttendancePattern) []string {
	var flags []string

	if pattern.AttendanceRate < 70 {
		flags = append(flags, "CRITICAL_ATTENDANCE")
	} else if pattern.AttendanceRate < 80 {
		flags = append(flags, "LOW_ATTENDANCE")
	}

	if pattern.ConsecutiveAbsences >= 3 {
		flags = append(flags, "CONSECUTIVE_ABSENCES")
	}

	// Check for patterns (e.g., frequently absent on Mondays)
	for day, count := range pattern.AbsencePattern {
		if count >= 3 && float64(count)/float64(pattern.AbsentDays) > 0.4 {
			flags = append(flags, fmt.Sprintf("PATTERN_%s", day))
		}
	}

	return flags
}

func (ta *TeacherAnalytics) calculateRiskLevel(attendance, behavior, academic float64) string {
	// Weighted average: attendance 40%, behavior 30%, academic 30%
	score := (attendance*0.4 + behavior*0.3 + academic*0.3)

	if score >= 85 {
		return "low"
	} else if score >= 75 {
		return "medium"
	} else if score >= 65 {
		return "high"
	} else {
		return "critical"
	}
}

func (ta *TeacherAnalytics) generateInsights(insight *StudentPerformanceInsight) ([]string, []string, []string) {
	var strengths, concerns, recommendations []string

	// Attendance analysis
	if insight.AttendanceScore >= 95 {
		strengths = append(strengths, "Excellent attendance record")
	} else if insight.AttendanceScore >= 90 {
		strengths = append(strengths, "Good attendance consistency")
	} else if insight.AttendanceScore < 80 {
		concerns = append(concerns, "Attendance below school standards")
		recommendations = append(recommendations, "Schedule parent meeting to discuss attendance")
	}

	// Behavior analysis
	if insight.BehaviorScore >= 90 {
		strengths = append(strengths, "Positive classroom behavior")
	} else if insight.BehaviorScore < 70 {
		concerns = append(concerns, "Classroom behavior needs improvement")
		recommendations = append(recommendations, "Implement behavior intervention plan")
	}

	// Academic analysis
	if insight.AcademicScore >= 85 {
		strengths = append(strengths, "Strong academic performance")
	} else if insight.AcademicScore < 70 {
		concerns = append(concerns, "Academic performance below grade level")
		recommendations = append(recommendations, "Consider additional academic support")
	}

	// Overall recommendations
	if insight.OverallRisk == "high" || insight.OverallRisk == "critical" {
		recommendations = append(recommendations, "Priority case - immediate intervention needed")
	}

	return strengths, concerns, recommendations
}

func (ta *TeacherAnalytics) getRecentParentContacts(ctx context.Context, studentID string) []ParentContact {
	// Placeholder - would query communication logs
	return []ParentContact{
		{
			Date:    time.Now().AddDate(0, 0, -7),
			Type:    "email",
			Reason:  "attendance",
			Outcome: "positive",
			Notes:   "Parent acknowledged attendance concerns, committed to improvement",
		},
	}
}

func (ta *TeacherAnalytics) createInterventionPlan(insight *StudentPerformanceInsight) *InterventionPlan {
	plan := &InterventionPlan{
		CreatedAt: time.Now(),
		Priority:  insight.OverallRisk,
		Status:    "active",
		Timeline:  "4 weeks",
		ReviewDate: time.Now().AddDate(0, 0, 28),
	}

	// Set goals based on concerns
	if insight.AttendanceScore < 80 {
		plan.Goals = append(plan.Goals, "Improve attendance to 90% or higher")
		plan.Actions = append(plan.Actions, InterventionAction{
			Description: "Daily attendance monitoring with parent notification",
			Responsible: "teacher",
			DueDate:     time.Now().AddDate(0, 0, 14),
			Status:      "pending",
		})
	}

	if insight.AcademicScore < 70 {
		plan.Goals = append(plan.Goals, "Raise academic performance to grade level")
		plan.Actions = append(plan.Actions, InterventionAction{
			Description: "Weekly tutoring sessions",
			Responsible: "teacher",
			DueDate:     time.Now().AddDate(0, 0, 7),
			Status:      "pending",
		})
	}

	return plan
}

// GetTeacherDashboardData provides comprehensive data for teacher dashboard
func (ta *TeacherAnalytics) GetTeacherDashboardData(ctx context.Context, teacherID string) (map[string]interface{}, error) {
	// Get class overviews
	classOverviews, err := ta.GetClassAttendanceOverview(ctx, teacherID)
	if err != nil {
		return nil, err
	}

	// Calculate summary statistics
	totalStudents := 0
	totalCritical := 0
	totalNeedsAttention := 0
	var avgAttendance float64

	for _, overview := range classOverviews {
		totalStudents += overview.TotalStudents
		totalCritical += overview.CriticalConcerns
		totalNeedsAttention += overview.NeedsAttention
		avgAttendance += overview.OverallRate
	}

	if len(classOverviews) > 0 {
		avgAttendance = avgAttendance / float64(len(classOverviews))
	}

	return map[string]interface{}{
		"teacher_id":              teacherID,
		"total_students":          totalStudents,
		"critical_concerns":       totalCritical,
		"needs_attention":         totalNeedsAttention,
		"average_attendance":      avgAttendance,
		"class_overviews":         classOverviews,
		"recommendations_count":   totalCritical + totalNeedsAttention,
	}, nil
} 