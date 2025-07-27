package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/schoolgpt/backend/internal/handlers"
	"github.com/schoolgpt/backend/internal/middleware"
)

// SetupRoutes sets up all routes for the application
func SetupRoutes(
	r *gin.Engine,
	authMiddleware *middleware.AuthMiddleware,
	gptHandler *handlers.GPTHandler,
	attendanceHandler *handlers.AttendanceHandler,
	voiceHandler *handlers.VoiceHandler,
	teacherTasksHandler *handlers.TeacherTasksHandler,
	schoolSetupHandler *handlers.SchoolSetupHandler,
	setupAgentHandler *handlers.SetupAgentHandler,
	educationalHandler *handlers.EducationalHandler,
	userManagementHandler *handlers.UserManagementHandler,
	teacherAnalyticsHandler *handlers.TeacherAnalyticsHandler,
) {
	// API v1 group
	api := r.Group("/api/v1")

	// Static files
	r.Static("/static", "./static")

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Test endpoint (no authentication required)
	r.POST("/test-gpt", gptHandler.HandleAskTest)

	// School setup endpoints (public for initial setup)
	setupRoutes := api.Group("/setup")
	{
		// Get available templates for school setup
		setupRoutes.GET("/templates", schoolSetupHandler.GetSchoolSetupTemplates)

		// Get schools by admin
		setupRoutes.GET("/schools", schoolSetupHandler.GetSchoolsByAdmin)

		// Create a new school
		setupRoutes.POST("/school", schoolSetupHandler.CreateSchool)

		// Get school configuration
		setupRoutes.GET("/school/:school_id", schoolSetupHandler.GetSchoolConfiguration)

		// Setup progress management
		setupRoutes.GET("/school/:school_id/progress", schoolSetupHandler.GetSetupProgress)
		setupRoutes.PUT("/school/:school_id/progress", schoolSetupHandler.UpdateSetupProgress)
		setupRoutes.POST("/school/:school_id/validate", schoolSetupHandler.ValidateSetupData)

		// Schema management endpoints
		setupRoutes.GET("/school/:school_id/schema/:entity_type", schoolSetupHandler.GetEntitySchema)
		setupRoutes.PUT("/school/:school_id/schema/:entity_type", schoolSetupHandler.UpdateEntitySchema)
		setupRoutes.POST("/school/:school_id/schema/:entity_type/fields", schoolSetupHandler.AddCustomField)
		setupRoutes.DELETE("/school/:school_id/schema/:entity_type/fields/:field_id", schoolSetupHandler.RemoveCustomField)

		// AI-powered setup agent endpoints (public for easy onboarding)
		setupRoutes.POST("/chat/start", setupAgentHandler.StartChat)
		setupRoutes.POST("/chat/:session_id/message", setupAgentHandler.SendMessage)
		setupRoutes.GET("/chat/:session_id/history", setupAgentHandler.GetChatHistory)
		setupRoutes.POST("/chat/:session_id/confirm", setupAgentHandler.ConfirmConfiguration)
		setupRoutes.GET("/chat/examples", setupAgentHandler.GetChatExamples)
		setupRoutes.GET("/chat/sessions", setupAgentHandler.GetActiveSessions)
	}

	// User Management endpoints (admin only)
	userRoutes := api.Group("/users")
	{
		// Create single user (admin only)
		userRoutes.POST("", authMiddleware.RoleBasedAccess("admin"), userManagementHandler.CreateUser)

		// Bulk create users (admin only)
		userRoutes.POST("/bulk", authMiddleware.RoleBasedAccess("admin"), userManagementHandler.BulkCreateUsers)

		// Upload CSV for bulk user creation (admin only)
		userRoutes.POST("/upload-csv", authMiddleware.RoleBasedAccess("admin"), userManagementHandler.UploadCSV)

		// Get users with filtering (admin and teachers can view)
		userRoutes.GET("", authMiddleware.RoleBasedAccess("admin", "teacher"), userManagementHandler.GetUsers)
	}

	// GPT endpoints
	gptRoutes := api.Group("/gpt")
	{
		// Ask GPT a question (requires authentication)
		gptRoutes.POST("/ask", authMiddleware.Authenticate(), gptHandler.HandleAsk)

		// Ask GPT about attendance (requires authentication and teacher role)
		gptRoutes.POST("/attendance", authMiddleware.RoleBasedAccess("teacher", "admin"), gptHandler.HandleAttendanceQuery)
	}

	// Attendance endpoints
	attendanceRoutes := api.Group("/attendance")
	{
		// Get attendance (requires authentication and appropriate role)
		attendanceRoutes.GET("/:student_id/:date", authMiddleware.Authenticate(), attendanceHandler.GetAttendance)

		// Create attendance (requires authentication and teacher role)
		attendanceRoutes.POST("", authMiddleware.RoleBasedAccess("teacher", "admin"), attendanceHandler.CreateAttendance)
	}

	// Voice endpoints
	voiceRoutes := api.Group("/voice")
	{
		// Transcribe audio (requires authentication)
		voiceRoutes.POST("/transcribe", authMiddleware.Authenticate(), voiceHandler.Transcribe)

		// Synthesize speech (requires authentication)
		voiceRoutes.POST("/synthesize", authMiddleware.Authenticate(), voiceHandler.Synthesize)
	}

	// Teacher automation endpoints
	teacherRoutes := api.Group("/teacher")
	{
		// Mark attendance with AI insights (teachers and admins only)
		teacherRoutes.POST("/mark-attendance", authMiddleware.RoleBasedAccess("teacher", "admin"), teacherTasksHandler.HandleMarkAttendance)

		// Get AI-powered student analysis (teachers and admins only)
		teacherRoutes.POST("/analyze-student", authMiddleware.RoleBasedAccess("teacher", "admin"), teacherTasksHandler.HandleStudentAnalysis)

		// Send AI-enhanced parent notifications (teachers and admins only)
		teacherRoutes.POST("/notify-parent", authMiddleware.RoleBasedAccess("teacher", "admin"), teacherTasksHandler.HandleParentNotification)
	}

	// Educational feature endpoints
	eduRoutes := api.Group("/education")
	{
		// Subject management (teachers and admins only)
		eduRoutes.POST("/subjects", authMiddleware.RoleBasedAccess("teacher", "admin"), educationalHandler.HandleCreateSubject)

		// Assignment management
		eduRoutes.POST("/assignments", authMiddleware.RoleBasedAccess("teacher", "admin"), educationalHandler.HandleCreateAssignment)
		eduRoutes.GET("/assignments/teacher", authMiddleware.RoleBasedAccess("teacher", "admin"), educationalHandler.HandleGetTeacherAssignments)
		eduRoutes.GET("/assignments/:assignment_id/submissions", authMiddleware.RoleBasedAccess("teacher", "admin"), educationalHandler.HandleGetAssignmentSubmissions)

		// Assignment submissions (students)
		eduRoutes.POST("/assignments/submit", authMiddleware.RoleBasedAccess("student"), educationalHandler.HandleSubmitAssignment)

		// Grading (teachers and admins only)
		eduRoutes.POST("/submissions/:submission_id/grade", authMiddleware.RoleBasedAccess("teacher", "admin"), educationalHandler.HandleGradeAssignment)

		// Grade viewing
		eduRoutes.GET("/students/:student_id/grades", authMiddleware.Authenticate(), educationalHandler.HandleGetStudentGrades)

		// Dashboards
		eduRoutes.GET("/dashboard/student", authMiddleware.RoleBasedAccess("student"), educationalHandler.HandleGetStudentDashboard)
		eduRoutes.GET("/dashboard/teacher", authMiddleware.RoleBasedAccess("teacher", "admin"), educationalHandler.HandleGetTeacherDashboard)
	}

	// Advanced Teacher Analytics endpoints
	analyticsRoutes := api.Group("/analytics")
	{
		// Student analysis endpoints (teachers and admins only)
		analyticsRoutes.GET("/student/:student_id/detailed", authMiddleware.RoleBasedAccess("teacher", "admin"), teacherAnalyticsHandler.GetStudentDetailedAnalysis)
		analyticsRoutes.GET("/student/:student_id/insights", authMiddleware.RoleBasedAccess("teacher", "admin"), teacherAnalyticsHandler.GenerateAttendanceInsights)
		analyticsRoutes.GET("/student/:student_id/recommendations", authMiddleware.RoleBasedAccess("teacher", "admin"), teacherAnalyticsHandler.GetInterventionRecommendations)

		// Class and teacher analytics (teachers and admins only)
		analyticsRoutes.GET("/class/overview", authMiddleware.RoleBasedAccess("teacher", "admin"), teacherAnalyticsHandler.GetClassOverview)
		analyticsRoutes.GET("/class/trends", authMiddleware.RoleBasedAccess("teacher", "admin"), teacherAnalyticsHandler.GetClassPerformanceTrends)
		analyticsRoutes.GET("/teacher/dashboard", authMiddleware.RoleBasedAccess("teacher", "admin"), teacherAnalyticsHandler.GetTeacherDashboard)
	}
}
