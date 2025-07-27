package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/schoolgpt/backend/internal/models"
	"github.com/schoolgpt/backend/internal/services"
)

// SchoolSetupHandler handles school setup and configuration endpoints
type SchoolSetupHandler struct {
	setupService *services.SchoolSetupService
}

// NewSchoolSetupHandler creates a new school setup handler
func NewSchoolSetupHandler(setupService *services.SchoolSetupService) *SchoolSetupHandler {
	return &SchoolSetupHandler{
		setupService: setupService,
	}
}

// GetSchoolSetupTemplates returns available templates for school setup
func (h *SchoolSetupHandler) GetSchoolSetupTemplates(c *gin.Context) {
	templates := h.setupService.GetAvailableTemplates()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    templates,
		"message": "Available templates retrieved successfully",
	})
}

// CreateSchool handles the initial school creation
func (h *SchoolSetupHandler) CreateSchool(c *gin.Context) {
	var req services.CreateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if req.SchoolName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "School name is required",
		})
		return
	}

	// Create school configuration
	config, err := h.setupService.CreateSchoolConfiguration(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create school configuration",
			"details": err.Error(),
		})
		return
	}

	// Update setup progress
	err = h.setupService.UpdateSetupProgress(c.Request.Context(), config.ID, 1, "School Created")
	if err != nil {
		// Log error but don't fail the request
		// log.Printf("Failed to update setup progress: %v", err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":     true,
		"message":     "School created successfully",
		"school_id":   config.ID,
		"school_code": config.SchoolCode,
		"setup_step":  1,
		"next_step":   "Configure data fields for Students, Teachers, and Parents",
		"data":        config,
	})
}

// GetSchoolConfiguration retrieves the school's current configuration
func (h *SchoolSetupHandler) GetSchoolConfiguration(c *gin.Context) {
	schoolID := c.Param("school_id")
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "School ID is required",
		})
		return
	}

	config, err := h.setupService.GetSchoolConfiguration(c.Request.Context(), schoolID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "School configuration not found",
			"details": err.Error(),
		})
		return
	}

	// Get setup progress
	progress, err := h.setupService.GetSetupProgress(c.Request.Context(), schoolID)
	if err != nil {
		// Set default progress if not found
		progress = &models.SetupProgress{
			SchoolID:    schoolID,
			CurrentStep: 1,
			TotalSteps:  6,
			StepName:    "School Created",
			Completed:   false,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"data":           config,
		"setup_progress": progress,
		"message":        "School configuration retrieved successfully",
	})
}

// GetEntitySchema retrieves the schema for a specific entity type
func (h *SchoolSetupHandler) GetEntitySchema(c *gin.Context) {
	schoolID := c.Param("school_id")
	entityType := c.Param("entity_type")

	if schoolID == "" || entityType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "School ID and entity type are required",
		})
		return
	}

	config, err := h.setupService.GetSchoolConfiguration(c.Request.Context(), schoolID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "School configuration not found",
		})
		return
	}

	schema, exists := config.Schemas[entityType]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Entity schema not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    schema,
		"message": "Schema retrieved successfully",
	})
}

// UpdateEntitySchema updates the schema for a specific entity type
func (h *SchoolSetupHandler) UpdateEntitySchema(c *gin.Context) {
	schoolID := c.Param("school_id")
	entityType := c.Param("entity_type")

	if schoolID == "" || entityType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "School ID and entity type are required",
		})
		return
	}

	var req services.UpdateSchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Override entity type from URL
	req.EntityType = entityType

	err := h.setupService.UpdateSchoolSchema(c.Request.Context(), schoolID, entityType, req.Schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update schema",
			"details": err.Error(),
		})
		return
	}

	// Update setup progress if this is step 2
	err = h.setupService.UpdateSetupProgress(c.Request.Context(), schoolID, 2, "Data Schema Configured")
	if err != nil {
		// Log error but don't fail the request
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Schema updated successfully",
		"entity_type": entityType,
		"setup_step":  2,
		"next_step":   "Add custom fields if needed",
	})
}

// AddCustomField adds a custom field to an entity schema
func (h *SchoolSetupHandler) AddCustomField(c *gin.Context) {
	schoolID := c.Param("school_id")
	entityType := c.Param("entity_type")

	if schoolID == "" || entityType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "School ID and entity type are required",
		})
		return
	}

	var req services.AddFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Override entity type from URL
	req.EntityType = entityType

	// Validate field
	if err := h.setupService.ValidateSchemaField(req.Field); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid field configuration",
			"details": err.Error(),
		})
		return
	}

	// Add field
	err := h.setupService.AddCustomField(c.Request.Context(), schoolID, entityType, req.Field)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to add custom field",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Custom field added successfully",
		"field":   req.Field,
	})
}

// RemoveCustomField removes a custom field from an entity schema
func (h *SchoolSetupHandler) RemoveCustomField(c *gin.Context) {
	schoolID := c.Param("school_id")
	entityType := c.Param("entity_type")
	fieldID := c.Param("field_id")

	if schoolID == "" || entityType == "" || fieldID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "School ID, entity type, and field ID are required",
		})
		return
	}

	err := h.setupService.RemoveCustomField(c.Request.Context(), schoolID, entityType, fieldID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to remove custom field",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Custom field removed successfully",
		"field_id": fieldID,
	})
}

// UpdateSetupProgress manually updates the setup progress
func (h *SchoolSetupHandler) UpdateSetupProgress(c *gin.Context) {
	schoolID := c.Param("school_id")
	stepStr := c.Query("step")
	stepName := c.Query("step_name")

	if schoolID == "" || stepStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "School ID and step are required",
		})
		return
	}

	step, err := strconv.Atoi(stepStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid step number",
		})
		return
	}

	if stepName == "" {
		stepNames := []string{
			"School Created",
			"Data Schema Configured",
			"Custom Fields Added",
			"Users Imported",
			"Permissions Set",
			"Setup Complete",
		}
		if step <= len(stepNames) {
			stepName = stepNames[step-1]
		}
	}

	err = h.setupService.UpdateSetupProgress(c.Request.Context(), schoolID, step, stepName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update setup progress",
			"details": err.Error(),
		})
		return
	}

	// Get updated progress
	progress, err := h.setupService.GetSetupProgress(c.Request.Context(), schoolID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Setup progress updated successfully",
			"step":    step,
		})
		return
	}

	nextSteps := map[int]string{
		1: "Configure data fields for your entities",
		2: "Add any custom fields needed",
		3: "Import your users (teachers, students, parents)",
		4: "Set up permissions and roles",
		5: "Complete final configuration",
		6: "School setup is complete!",
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Setup progress updated successfully",
		"progress":  progress,
		"next_step": nextSteps[step+1],
	})
}

// GetSetupProgress retrieves the current setup progress
func (h *SchoolSetupHandler) GetSetupProgress(c *gin.Context) {
	schoolID := c.Param("school_id")

	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "School ID is required",
		})
		return
	}

	progress, err := h.setupService.GetSetupProgress(c.Request.Context(), schoolID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Setup progress not found",
			"details": err.Error(),
		})
		return
	}

	// Calculate progress percentage
	progressPercentage := (float64(progress.CurrentStep) / float64(progress.TotalSteps)) * 100

	steps := []string{
		"Create School",
		"Configure Data Schema",
		"Add Custom Fields",
		"Import Users",
		"Set Permissions",
		"Complete Setup",
	}

	c.JSON(http.StatusOK, gin.H{
		"success":             true,
		"data":                progress,
		"progress_percentage": progressPercentage,
		"steps":               steps,
		"message":             "Setup progress retrieved successfully",
	})
}

// ValidateSetupData validates the setup data before proceeding
func (h *SchoolSetupHandler) ValidateSetupData(c *gin.Context) {
	schoolID := c.Param("school_id")

	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "School ID is required",
		})
		return
	}

	config, err := h.setupService.GetSchoolConfiguration(c.Request.Context(), schoolID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "School configuration not found",
		})
		return
	}

	// Validate configuration
	validationErrors := []string{}

	// Check if basic information is complete
	if config.SchoolName == "" {
		validationErrors = append(validationErrors, "School name is required")
	}
	if config.Region == "" {
		validationErrors = append(validationErrors, "Region is required")
	}
	if config.EducationSystem == "" {
		validationErrors = append(validationErrors, "Education system is required")
	}

	// Check if schemas are configured
	requiredEntities := []string{"student", "teacher", "parent"}
	for _, entity := range requiredEntities {
		if _, exists := config.Schemas[entity]; !exists {
			validationErrors = append(validationErrors, "Schema for "+entity+" is not configured")
		}
	}

	isValid := len(validationErrors) == 0

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"is_valid":          isValid,
		"validation_errors": validationErrors,
		"message":           "Validation completed",
		"school_status":     config.Status,
	})
}

// GetSchoolsByAdmin retrieves all schools created by the current admin
func (h *SchoolSetupHandler) GetSchoolsByAdmin(c *gin.Context) {
	// In a real application, you would get the admin ID from authentication middleware
	// For now, we'll use a query parameter or header
	adminEmail := c.Query("admin_email")
	if adminEmail == "" {
		adminEmail = c.GetHeader("X-Admin-Email")
	}
	
	if adminEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Admin email is required",
		})
		return
	}

	schools, err := h.setupService.GetSchoolsByAdmin(c.Request.Context(), adminEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve schools",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"data":        schools,
		"count":       len(schools),
		"admin_email": adminEmail,
		"message":     "Schools retrieved successfully",
	})
}
