package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/schoolgpt/backend/internal/storage"
)

// AttendanceHandler handles attendance-related requests
type AttendanceHandler struct {
	db *storage.FirestoreDB
}

// NewAttendanceHandler creates a new attendance handler
func NewAttendanceHandler(db *storage.FirestoreDB) *AttendanceHandler {
	return &AttendanceHandler{
		db: db,
	}
}

// GetAttendanceRequest represents a request to get attendance
type GetAttendanceRequest struct {
	StudentID string `uri:"student_id" binding:"required"`
	Date      string `uri:"date" binding:"required"`
}

// CreateAttendanceRequest represents a request to create attendance
type CreateAttendanceRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	Date      string `json:"date" binding:"required"`
	Status    string `json:"status" binding:"required,oneof=present absent late excused"`
}

// GetAttendance handles a request to get attendance
func (h *AttendanceHandler) GetAttendance(c *gin.Context) {
	var req GetAttendanceRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user role from context (set by auth middleware)
	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	// Check if the user is authorized to access this attendance record
	// Parents can only access their own children's records
	// Teachers and admins can access any record
	if userRole.(string) == "parent" {
		// In a real application, this would check if the student belongs to the parent
		// For now, just log a warning
		fmt.Printf("Parent %s accessing attendance for student %s\n", userID.(string), req.StudentID)
	}

	// Get the attendance record
	record, err := h.db.FetchAttendance(c.Request.Context(), req.StudentID, req.Date)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attendance record not found"})
		return
	}

	// Return the record
	c.JSON(http.StatusOK, record)
}

// CreateAttendance handles a request to create attendance
func (h *AttendanceHandler) CreateAttendance(c *gin.Context) {
	var req CreateAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get user role from context (set by auth middleware)
	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	// Only teachers and admins can create attendance records
	if userRole.(string) != "teacher" && userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only teachers and admins can create attendance records"})
		return
	}

	// Create the attendance record
	record := &storage.AttendanceRecord{
		StudentID: req.StudentID,
		Date:      req.Date,
		Status:    req.Status,
		MarkedBy:  userID.(string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save the record
	if err := h.db.SaveAttendance(c.Request.Context(), record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the record
	c.JSON(http.StatusCreated, record)
}
