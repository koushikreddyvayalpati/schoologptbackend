package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/schoolgpt/backend/internal/storage"
	"github.com/schoolgpt/backend/pkg/gpt"
)

// GPTHandler handles GPT-related requests
type GPTHandler struct {
	gptClient *gpt.Client
	db        *storage.FirestoreDB
}

// NewGPTHandler creates a new GPT handler
func NewGPTHandler(gptClient *gpt.Client, db *storage.FirestoreDB) *GPTHandler {
	return &GPTHandler{
		gptClient: gptClient,
		db:        db,
	}
}

// AskRequest represents a request to ask GPT a question
type AskRequest struct {
	Query string `json:"query" binding:"required"`
}

// AskResponse represents a response from GPT
type AskResponse struct {
	Answer string `json:"answer"`
}

// HandleAsk handles a request to ask GPT a question with database context
func (h *GPTHandler) HandleAsk(c *gin.Context) {
	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user context from auth middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "User role not found"})
		return
	}

	entityID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Entity ID not found"})
		return
	}

	fmt.Printf("GPT query from %s (%s): %s\n", userRole.(string), userID.(string), req.Query)

	// Get relevant data based on user role and create context
	contextData, err := h.getContextDataForUser(c.Request.Context(), userRole.(string), entityID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error getting context data: %v", err)})
		return
	}

	// Create enhanced query with context
	enhancedQuery := h.createEnhancedQuery(req.Query, userRole.(string), contextData)

	// Call GPT with enhanced context
	answer, err := h.gptClient.HandleGPTQuery(c.Request.Context(), enhancedQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the response
	c.JSON(http.StatusOK, gin.H{
		"answer":           answer,
		"model":            h.gptClient.GetModelName(),
		"context_provided": true,
	})
}

// HandleAskTest handles a test request to ask GPT a question (no authentication required)
func (h *GPTHandler) HandleAskTest(c *gin.Context) {
	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the query (test mode)
	fmt.Printf("GPT test query: %s\n", req.Query)

	// Call AI (Gemini or OpenAI)
	answer, err := h.gptClient.HandleGPTQuery(c.Request.Context(), req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the response with model info
	c.JSON(http.StatusOK, gin.H{
		"answer":       answer,
		"model":        h.gptClient.GetModelName(),
		"using_gemini": h.gptClient.IsUsingGemini(),
	})
}

// HandleAttendanceQuery handles a request to get attendance information using GPT function calling
func (h *GPTHandler) HandleAttendanceQuery(c *gin.Context) {
	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user context from auth middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "User role not found"})
		return
	}

	entityID, exists := c.Get("entity_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Entity ID not found"})
		return
	}

	fmt.Printf("Attendance query from %s (%s): %s\n", userRole.(string), userID.(string), req.Query)

	// Get attendance data based on user role
	attendanceData, err := h.getAttendanceDataForUser(c.Request.Context(), userRole.(string), entityID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error getting attendance data: %v", err)})
		return
	}

	// Create enhanced query with attendance context
	enhancedQuery := h.createAttendanceQuery(req.Query, userRole.(string), attendanceData)

	// Call GPT with attendance context
	answer, err := h.gptClient.HandleGPTQuery(c.Request.Context(), enhancedQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the response
	c.JSON(http.StatusOK, gin.H{
		"query":    req.Query,
		"response": answer,
		"model":    h.gptClient.GetModelName(),
		"provider": "google-gemini-free",
	})
}

// getContextDataForUser gets relevant data based on user role
func (h *GPTHandler) getContextDataForUser(ctx context.Context, role, entityID string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	switch role {
	case "admin":
		// Admins get comprehensive data
		students, err := h.db.GetAllStudents(ctx)
		if err != nil {
			return nil, err
		}
		data["students"] = students
		data["role"] = "admin"
		data["permissions"] = "full_access"

	case "teacher":
		// Teachers get data for their students
		students, err := h.db.GetStudentsByTeacher(ctx, entityID)
		if err != nil {
			return nil, err
		}
		teacher, err := h.db.GetTeacher(ctx, entityID)
		if err != nil {
			return nil, err
		}
		data["students"] = students
		data["teacher"] = teacher
		data["role"] = "teacher"
		data["permissions"] = "own_classes_only"

	case "parent":
		// Parents get data for their children only
		students, err := h.db.GetStudentsByParent(ctx, entityID)
		if err != nil {
			return nil, err
		}
		data["children"] = students
		data["role"] = "parent"
		data["permissions"] = "own_children_only"

	default:
		return nil, fmt.Errorf("unknown role: %s", role)
	}

	return data, nil
}

// createEnhancedQuery creates an enhanced query with context data
func (h *GPTHandler) createEnhancedQuery(originalQuery, role string, contextData map[string]interface{}) string {
	// Convert context data to JSON string
	contextJSON, _ := json.MarshalIndent(contextData, "", "  ")

	basePrompt := fmt.Sprintf(`You are SchoolGPT, an AI assistant for Oakwood Elementary School. 

USER ROLE: %s
QUERY: %s

AVAILABLE DATA:
%s

INSTRUCTIONS:
- Answer the query using ONLY the data provided above
- Respect the user's role permissions:
  * Admins: Can access all student data
  * Teachers: Can only access data for students in their classes
  * Parents: Can only access data for their own children
- If the user asks for data they don't have permission to access, politely explain the limitation
- Provide helpful, accurate information based on the available data
- If specific data isn't available, say so clearly
- Use a friendly, professional tone appropriate for a school environment
- Format responses clearly with bullet points or tables when helpful

Please answer the query based on the available data and role permissions.`,
		role, originalQuery, string(contextJSON))

	return basePrompt
}

// getAttendanceDataForUser gets attendance data based on user role
func (h *GPTHandler) getAttendanceDataForUser(ctx context.Context, role, entityID string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	switch role {
	case "admin", "teacher":
		// Get recent attendance data
		endDate := time.Now().Format("2006-01-02")
		startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02") // Last 30 days

		var teacherID string
		if role == "teacher" {
			teacherID = entityID
		}

		records, err := h.db.GetAttendanceByDateRange(ctx, startDate, endDate, teacherID)
		if err != nil {
			return nil, err
		}

		data["attendance_records"] = records
		data["date_range"] = fmt.Sprintf("%s to %s", startDate, endDate)
		data["role"] = role

	case "parent":
		// Parents can only see their children's attendance
		students, err := h.db.GetStudentsByParent(ctx, entityID)
		if err != nil {
			return nil, err
		}

		var allRecords []storage.AttendanceSummary
		for _, student := range students {
			summary, err := h.db.GetAttendanceSummary(ctx, student.ID, 30) // Last 30 days
			if err != nil {
				continue
			}
			allRecords = append(allRecords, *summary)
		}

		data["attendance_summaries"] = allRecords
		data["role"] = "parent"

	default:
		return nil, fmt.Errorf("unknown role: %s", role)
	}

	return data, nil
}

// createAttendanceQuery creates an enhanced attendance query
func (h *GPTHandler) createAttendanceQuery(originalQuery, role string, attendanceData map[string]interface{}) string {
	contextJSON, _ := json.MarshalIndent(attendanceData, "", "  ")

	basePrompt := fmt.Sprintf(`You are SchoolGPT's attendance specialist for Oakwood Elementary School.

USER ROLE: %s
ATTENDANCE QUERY: %s

ATTENDANCE DATA:
%s

INSTRUCTIONS:
- Analyze the attendance data to answer the query
- Provide specific statistics and insights
- Highlight any attendance concerns (below 90%% attendance rate)
- Suggest actions for poor attendance when appropriate
- Use clear formatting with tables or bullet points
- Be professional and helpful
- Only discuss attendance data the user has permission to access

Please provide a comprehensive attendance analysis based on the query and available data.`,
		role, originalQuery, string(contextJSON))

	return basePrompt
}
