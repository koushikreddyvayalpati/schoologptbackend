package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/schoolgpt/backend/internal/services"
)

// SetupAgentHandler handles AI-powered school setup conversations
type SetupAgentHandler struct {
	setupAgent *services.SetupAgent
}

// NewSetupAgentHandler creates a new setup agent handler
func NewSetupAgentHandler(setupAgent *services.SetupAgent) *SetupAgentHandler {
	return &SetupAgentHandler{
		setupAgent: setupAgent,
	}
}

// StartChatRequest represents a request to start a new chat session
type StartChatRequest struct {
	AdminName  string `json:"admin_name" binding:"required"`
	AdminEmail string `json:"admin_email" binding:"required"`
}

// ChatMessageRequest represents a message in the conversation
type ChatMessageRequest struct {
	Message string `json:"message" binding:"required"`
}

// ConfirmConfigRequest represents a configuration confirmation
type ConfirmConfigRequest struct {
	Confirmed bool `json:"confirmed" binding:"required"`
}

// StartChat initiates a new conversation with the setup agent
func (h *SetupAgentHandler) StartChat(c *gin.Context) {
	var req StartChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate email format (basic validation)
	if !strings.Contains(req.AdminEmail, "@") {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid email format",
		})
		return
	}

	// Start chat session
	session, err := h.setupAgent.StartChatSession(c.Request.Context(), req.AdminName, req.AdminEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to start chat session",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":         true,
		"message":         "Chat session started successfully",
		"session_id":      session.ID,
		"welcome_message": session.Messages[0].Content,
		"next_step":       "Tell me about your school's basic information",
		"suggestions":     h.generateWelcomeSuggestions(),
	})
}

// SendMessage processes a message from the admin
func (h *SetupAgentHandler) SendMessage(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Session ID is required",
		})
		return
	}

	var req ChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Process message
	session, err := h.setupAgent.ProcessMessage(c.Request.Context(), sessionID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to process message",
			"details": err.Error(),
		})
		return
	}

	// Get the latest agent response
	var latestResponse string
	if len(session.Messages) > 0 {
		latestResponse = session.Messages[len(session.Messages)-1].Content
	}

	response := gin.H{
		"success":              true,
		"agent_response":       latestResponse,
		"session_status":       session.Status,
		"confirmation_pending": session.ConfirmationPending,
		"message_count":        len(session.Messages),
	}

	// Generate contextual suggestions based on current state
	suggestions := h.generateSuggestions(session, req.Message)
	response["suggestions"] = suggestions

	// If configuration is pending, add configuration details
	if session.ConfirmationPending && session.GeneratedConfig != nil {
		response["generated_config"] = gin.H{
			"school_name":      session.GeneratedConfig.SchoolName,
			"region":           session.GeneratedConfig.Region,
			"education_system": session.GeneratedConfig.EducationSystem,
			"features":         session.GeneratedConfig.Features,
		}
		response["next_step"] = "Review the configuration and confirm or request changes"
	} else {
		response["next_step"] = "Continue describing your school's requirements"
	}

	c.JSON(http.StatusOK, response)
}

// GetChatHistory retrieves the full conversation history
func (h *SetupAgentHandler) GetChatHistory(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Session ID is required",
		})
		return
	}

	// Get chat session from database
	session, err := h.setupAgent.GetChatSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Chat session not found",
			"details": err.Error(),
		})
		return
	}

	// Prepare response with extracted information
	response := gin.H{
		"success":              true,
		"session_id":           session.ID,
		"admin_name":           session.AdminName,
		"admin_email":          session.AdminEmail,
		"status":               session.Status,
		"messages":             session.Messages,
		"confirmation_pending": session.ConfirmationPending,
		"created_at":           session.CreatedAt,
		"updated_at":           session.UpdatedAt,
	}

	// Add extracted information if available
	if session.ExtractedInfo.BasicInfo.Name != "" {
		response["extracted_info"] = session.ExtractedInfo
	}

	// Add generated configuration if available
	if session.GeneratedConfig != nil {
		response["generated_config"] = session.GeneratedConfig
	}

	c.JSON(http.StatusOK, response)
}

// ConfirmConfiguration handles admin's confirmation of the generated configuration
func (h *SetupAgentHandler) ConfirmConfiguration(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Session ID is required",
		})
		return
	}

	var req ConfirmConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Confirm configuration
	config, err := h.setupAgent.ConfirmConfiguration(c.Request.Context(), sessionID, req.Confirmed)
	if err != nil {
		if strings.Contains(err.Error(), "no configuration pending") {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "No configuration pending confirmation",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to process confirmation",
			"details": err.Error(),
		})
		return
	}

	if !req.Confirmed {
		// Configuration was rejected, continue conversation
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"message":   "Configuration rejected. You can now specify changes.",
			"status":    "modification_mode",
			"next_step": "Tell me what you'd like to change about the configuration",
		})
		return
	}

	// Configuration was confirmed and school created
	c.JSON(http.StatusCreated, gin.H{
		"success":     true,
		"message":     "School created successfully!",
		"school_id":   config.ID,
		"school_code": config.SchoolCode,
		"school_name": config.SchoolName,
		"status":      "completed",
		"next_steps": []string{
			"Import your users (teachers, students, parents)",
			"Set up user permissions and roles",
			"Configure class schedules and assignments",
			"Start using SchoolGPT features",
		},
	})
}

// GetActiveSessions retrieves all active chat sessions (for admin monitoring)
func (h *SetupAgentHandler) GetActiveSessions(c *gin.Context) {
	// This would typically require admin authentication
	// For now, returning a placeholder response

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Active chat sessions",
		"sessions": []gin.H{
			{
				"id":         "sample_session",
				"admin_name": "Demo Admin",
				"status":     "active",
				"created_at": "2025-06-25T10:00:00Z",
			},
		},
		"total_sessions": 1,
	})
}

// GetChatExamples returns example conversations for documentation
func (h *SetupAgentHandler) GetChatExamples(c *gin.Context) {
	examples := gin.H{
		"success": true,
		"examples": []gin.H{
			{
				"scenario": "CBSE School in India",
				"conversation": []gin.H{
					{
						"sender":  "admin",
						"message": "Hi, I need to set up our school system. We're Bright Future International School in Mumbai, following CBSE curriculum.",
					},
					{
						"sender":  "agent",
						"message": "Hello! Great to hear from Bright Future International School. CBSE curriculum and Mumbai location noted. Can you tell me more about your specific needs? For example, what information do you typically track for students?",
					},
					{
						"sender":  "admin",
						"message": "We need to track student blood groups for medical emergencies, house assignments for sports events, and transportation method. Also need financial management for fee collection.",
					},
					{
						"sender":  "agent",
						"message": "Perfect! I'll set up custom fields for blood group, house assignment, and transportation. Financial management will be enabled. Do you also need attendance tracking and parent communication features?",
					},
				},
			},
			{
				"scenario": "Small School - Basic Setup",
				"conversation": []gin.H{
					{
						"sender":  "admin",
						"message": "I'm setting up a small school system. We just need basic attendance and grade management.",
					},
					{
						"sender":  "agent",
						"message": "I understand you need a simple setup. What's your school name and which education system do you follow? Also, how many students do you have approximately?",
					},
					{
						"sender":  "admin",
						"message": "We're Little Stars Academy, about 200 students, following ICSE system in Bangalore.",
					},
					{
						"sender":  "agent",
						"message": "Great! ICSE system with attendance and grade management for 200 students. Would you like AI-powered insights to help teachers with student performance analysis?",
					},
				},
			},
		},
		"tips": []string{
			"Be specific about your school's unique needs",
			"Mention any special fields you track (blood group, house, etc.)",
			"Tell us about your preferred features (financial, transport, etc.)",
			"The AI can understand natural conversation - just chat normally!",
		},
	}

	c.JSON(http.StatusOK, examples)
}

// generateSuggestions creates contextual suggestions based on conversation state
func (h *SetupAgentHandler) generateSuggestions(session *services.ChatSession, lastMessage string) []string {
	// If configuration is pending, show confirmation options
	if session.ConfirmationPending {
		return []string{
			"✅ Yes, create the school with these settings",
			"📝 Let me modify something first",
			"❌ No, let's start over",
		}
	}

	// If school is completed, show next steps
	if session.Status == "completed" {
		return []string{
			"📊 Upload student data via Excel/CSV",
			"👨‍🏫 Create teacher accounts in bulk",
			"👨‍👩‍👧‍👦 Import parent/guardian information",
			"⚙️ Configure additional settings",
		}
	}

	// Generate suggestions based on conversation context
	extractedInfo := session.ExtractedInfo
	lastMessageLower := strings.ToLower(lastMessage)

	// If basic info is incomplete, suggest basic information
	if extractedInfo.BasicInfo.Name == "" {
		return []string{
			"🏫 Tell you our school name",
			"📍 Describe our location and curriculum",
			"🎓 Explain our grade levels and student count",
			"📊 Upload our existing data files",
		}
	}

	// If region/education system is missing
	if extractedInfo.BasicInfo.Region == "" || extractedInfo.BasicInfo.EducationSystem == "" {
		return []string{
			"🇮🇳 We're located in India with CBSE curriculum",
			"🇺🇸 We're in the United States with K-12 system",
			"🇬🇧 We're in the United Kingdom",
			"🌍 We're an international school",
		}
	}

	// If features haven't been discussed much
	if !extractedInfo.Features.FinancialManagement && !extractedInfo.Features.TransportManagement {
		return []string{
			"💰 We need financial management for fee collection",
			"🚌 We need transport management for bus routes",
			"📊 We track student blood groups and house assignments",
			"📱 We want parent communication features",
		}
	}

	// If discussing data/files
	if strings.Contains(lastMessageLower, "data") || strings.Contains(lastMessageLower, "excel") || strings.Contains(lastMessageLower, "csv") {
		return []string{
			"📊 I have Excel files with student data",
			"👨‍🏫 I have teacher information in spreadsheets",
			"👨‍👩‍👧‍👦 I have parent/guardian data to import",
			"📋 I need templates for data organization",
		}
	}

	// Default suggestions based on conversation progress
	return []string{
		"📊 I have Excel files with school data to upload",
		"⚙️ Show me all available features",
		"🎯 Let's proceed with basic setup",
		"💡 What other information do you need?",
	}
}

// generateWelcomeSuggestions creates suggestions for the welcome message
func (h *SetupAgentHandler) generateWelcomeSuggestions() []string {
	return []string{
		"🏫 Tell you about our school name and location",
		"📚 Describe our grade levels and student count",
		"🎓 Explain our education system (CBSE/ICSE/K-12/IB)",
		"📊 Upload existing Excel/CSV data files",
		"⚙️ Start with basic school information",
	}
}


