package handlers

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/schoolgpt/backend/internal/storage"
	"go.uber.org/zap"
)

// UserManagementHandler handles user creation and management endpoints
type UserManagementHandler struct {
	db     *storage.FirestoreDB
	logger *zap.Logger
}

// NewUserManagementHandler creates a new user management handler
func NewUserManagementHandler(db *storage.FirestoreDB, logger *zap.Logger) *UserManagementHandler {
	return &UserManagementHandler{
		db:     db,
		logger: logger,
	}
}

// CreateUserRequest represents a request to create a single user
type CreateUserRequest struct {
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required"`
	Role       string `json:"role" binding:"required,oneof=teacher student parent"`
	Grade      string `json:"grade,omitempty"`
	Subject    string `json:"subject,omitempty"`
	Department string `json:"department,omitempty"`
	StudentID  string `json:"student_id,omitempty"`
	EmployeeID string `json:"employee_id,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Address    string `json:"address,omitempty"`
}

// BulkCreateUsersRequest represents a request to create multiple users
type BulkCreateUsersRequest struct {
	Users []CreateUserRequest `json:"users" binding:"required"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	Grade      string `json:"grade,omitempty"`
	Subject    string `json:"subject,omitempty"`
	Department string `json:"department,omitempty"`
	StudentID  string `json:"student_id,omitempty"`
	EmployeeID string `json:"employee_id,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Address    string `json:"address,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// CreateUser handles single user creation
func (h *UserManagementHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate role-specific fields
	if err := h.validateUserRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	// Check if user already exists
	existingUser, _ := h.db.GetUserByEmail(c.Request.Context(), req.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "User with this email already exists",
		})
		return
	}

	// Create user based on role
	user, err := h.createUserByRole(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User created successfully",
		"user":    h.formatUserResponse(user),
	})
}

// BulkCreateUsers handles multiple user creation
func (h *UserManagementHandler) BulkCreateUsers(c *gin.Context) {
	var req BulkCreateUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if len(req.Users) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No users provided",
		})
		return
	}

	if len(req.Users) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Maximum 100 users allowed per bulk operation",
		})
		return
	}

	var createdUsers []UserResponse
	var failedUsers []map[string]interface{}

	for i, userReq := range req.Users {
		// Validate user request
		if err := h.validateUserRequest(&userReq); err != nil {
			failedUsers = append(failedUsers, map[string]interface{}{
				"index": i,
				"email": userReq.Email,
				"error": err.Error(),
			})
			continue
		}

		// Check if user already exists
		existingUser, _ := h.db.GetUserByEmail(c.Request.Context(), userReq.Email)
		if existingUser != nil {
			failedUsers = append(failedUsers, map[string]interface{}{
				"index": i,
				"email": userReq.Email,
				"error": "User with this email already exists",
			})
			continue
		}

		// Create user
		user, err := h.createUserByRole(c.Request.Context(), &userReq)
		if err != nil {
			h.logger.Error("Failed to create user in bulk operation",
				zap.Error(err),
				zap.String("email", userReq.Email))
			failedUsers = append(failedUsers, map[string]interface{}{
				"index": i,
				"email": userReq.Email,
				"error": err.Error(),
			})
			continue
		}

		createdUsers = append(createdUsers, h.formatUserResponse(user))
	}

	response := gin.H{
		"success":       true,
		"message":       fmt.Sprintf("Bulk operation completed. %d users created, %d failed", len(createdUsers), len(failedUsers)),
		"created_count": len(createdUsers),
		"failed_count":  len(failedUsers),
		"created_users": createdUsers,
	}

	if len(failedUsers) > 0 {
		response["failed_users"] = failedUsers
	}

	c.JSON(http.StatusCreated, response)
}

// UploadCSV handles CSV file upload for bulk user creation
func (h *UserManagementHandler) UploadCSV(c *gin.Context) {
	file, header, err := c.Request.FormFile("csv_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No CSV file provided",
		})
		return
	}
	defer file.Close()

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Only CSV files are allowed",
		})
		return
	}

	// Parse CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to parse CSV file",
			"details": err.Error(),
		})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "CSV file must contain headers and at least one data row",
		})
		return
	}

	// Parse headers
	headers := records[0]
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	// Required headers
	requiredHeaders := []string{"name", "email", "role"}
	for _, required := range requiredHeaders {
		if _, exists := headerMap[required]; !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Missing required header: %s", required),
			})
			return
		}
	}

	// Convert CSV rows to user requests
	var users []CreateUserRequest
	var parseErrors []string

	for rowIdx, record := range records[1:] {
		if len(record) < len(headers) {
			parseErrors = append(parseErrors, fmt.Sprintf("Row %d: Insufficient columns", rowIdx+2))
			continue
		}

		user := CreateUserRequest{
			Name:  strings.TrimSpace(record[headerMap["name"]]),
			Email: strings.TrimSpace(record[headerMap["email"]]),
			Role:  strings.TrimSpace(record[headerMap["role"]]),
		}

		// Optional fields
		if idx, exists := headerMap["grade"]; exists && idx < len(record) {
			user.Grade = strings.TrimSpace(record[idx])
		}
		if idx, exists := headerMap["subject"]; exists && idx < len(record) {
			user.Subject = strings.TrimSpace(record[idx])
		}
		if idx, exists := headerMap["department"]; exists && idx < len(record) {
			user.Department = strings.TrimSpace(record[idx])
		}
		if idx, exists := headerMap["student_id"]; exists && idx < len(record) {
			user.StudentID = strings.TrimSpace(record[idx])
		}
		if idx, exists := headerMap["employee_id"]; exists && idx < len(record) {
			user.EmployeeID = strings.TrimSpace(record[idx])
		}
		if idx, exists := headerMap["phone"]; exists && idx < len(record) {
			user.Phone = strings.TrimSpace(record[idx])
		}
		if idx, exists := headerMap["address"]; exists && idx < len(record) {
			user.Address = strings.TrimSpace(record[idx])
		}

		users = append(users, user)
	}

	if len(parseErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "CSV parsing errors",
			"details": parseErrors,
		})
		return
	}

	// Process users (same logic as BulkCreateUsers)
	var createdUsers []UserResponse
	var failedUsers []map[string]interface{}

	for i, userReq := range users {
		// Validate user request
		if err := h.validateUserRequest(&userReq); err != nil {
			failedUsers = append(failedUsers, map[string]interface{}{
				"row":   i + 2, // +2 because we skip header and array is 0-indexed
				"email": userReq.Email,
				"error": err.Error(),
			})
			continue
		}

		// Check if user already exists
		existingUser, _ := h.db.GetUserByEmail(c.Request.Context(), userReq.Email)
		if existingUser != nil {
			failedUsers = append(failedUsers, map[string]interface{}{
				"row":   i + 2,
				"email": userReq.Email,
				"error": "User with this email already exists",
			})
			continue
		}

		// Create user
		user, err := h.createUserByRole(c.Request.Context(), &userReq)
		if err != nil {
			h.logger.Error("Failed to create user from CSV",
				zap.Error(err),
				zap.String("email", userReq.Email))
			failedUsers = append(failedUsers, map[string]interface{}{
				"row":   i + 2,
				"email": userReq.Email,
				"error": err.Error(),
			})
			continue
		}

		createdUsers = append(createdUsers, h.formatUserResponse(user))
	}

	response := gin.H{
		"success":       true,
		"message":       fmt.Sprintf("CSV upload completed. %d users created, %d failed", len(createdUsers), len(failedUsers)),
		"filename":      header.Filename,
		"total_rows":    len(users),
		"created_count": len(createdUsers),
		"failed_count":  len(failedUsers),
		"created_users": createdUsers,
	}

	if len(failedUsers) > 0 {
		response["failed_users"] = failedUsers
	}

	c.JSON(http.StatusCreated, response)
}

// GetUsers retrieves users with optional filtering
func (h *UserManagementHandler) GetUsers(c *gin.Context) {
	role := c.Query("role")
	grade := c.Query("grade")
	department := c.Query("department")

	users, err := h.db.GetUsersWithFilters(c.Request.Context(), role, grade, department)
	if err != nil {
		h.logger.Error("Failed to get users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve users",
		})
		return
	}

	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, h.formatUserResponse(user))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   len(userResponses),
		"users":   userResponses,
		"filters": gin.H{
			"role":       role,
			"grade":      grade,
			"department": department,
		},
	})
}

// validateUserRequest validates user creation request
func (h *UserManagementHandler) validateUserRequest(req *CreateUserRequest) error {
	// Basic validation
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !strings.Contains(req.Email, "@") {
		return fmt.Errorf("invalid email format")
	}
	if req.Role == "" {
		return fmt.Errorf("role is required")
	}

	// Role-specific validation
	switch req.Role {
	case "student":
		if req.Grade == "" {
			return fmt.Errorf("grade is required for students")
		}
	case "teacher":
		if req.Subject == "" {
			return fmt.Errorf("subject is required for teachers")
		}
	case "parent":
		// Parents don't have specific required fields beyond the basics
	default:
		return fmt.Errorf("invalid role: %s", req.Role)
	}

	return nil
}

// createUserByRole creates a user based on their role
func (h *UserManagementHandler) createUserByRole(ctx context.Context, req *CreateUserRequest) (interface{}, error) {
	switch req.Role {
	case "student":
		student := &storage.Student{
			Name:      req.Name,
			Email:     req.Email,
			Grade:     req.Grade,
			StudentID: req.StudentID,
			Phone:     req.Phone,
			Address:   req.Address,
		}
		err := h.db.CreateStudent(ctx, student)
		return student, err

	case "teacher":
		teacher := &storage.Teacher{
			Name:       req.Name,
			Email:      req.Email,
			Subject:    req.Subject,
			Department: req.Department,
			EmployeeID: req.EmployeeID,
			Phone:      req.Phone,
			Address:    req.Address,
		}
		err := h.db.CreateTeacher(ctx, teacher)
		return teacher, err

	case "parent":
		parent := &storage.Parent{
			Name:    req.Name,
			Email:   req.Email,
			Phone:   req.Phone,
			Address: req.Address,
		}
		err := h.db.CreateParent(ctx, parent)
		return parent, err

	default:
		return nil, fmt.Errorf("unsupported role: %s", req.Role)
	}
}

// formatUserResponse formats a user object for API response
func (h *UserManagementHandler) formatUserResponse(user interface{}) UserResponse {
	switch u := user.(type) {
	case *storage.Student:
		return UserResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Role:      "student",
			Grade:     u.Grade,
			StudentID: u.StudentID,
			Phone:     u.Phone,
			Address:   u.Address,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	case *storage.Teacher:
		return UserResponse{
			ID:         u.ID,
			Name:       u.Name,
			Email:      u.Email,
			Role:       "teacher",
			Subject:    u.Subject,
			Department: u.Department,
			EmployeeID: u.EmployeeID,
			Phone:      u.Phone,
			Address:    u.Address,
			CreatedAt:  u.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:  u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	case *storage.Parent:
		return UserResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Role:      "parent",
			Phone:     u.Phone,
			Address:   u.Address,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	default:
		return UserResponse{}
	}
}
