package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/schoolgpt/backend/internal/models"
	"github.com/schoolgpt/backend/internal/storage"
	"github.com/schoolgpt/backend/pkg/gpt"
)

// SetupAgent handles AI-powered school configuration conversations
type SetupAgent struct {
	gptClient          *gpt.Client
	schoolSetupService *SchoolSetupService
	db                 *storage.FirestoreDB
}

// NewSetupAgent creates a new setup agent
func NewSetupAgent(gptClient *gpt.Client, schoolSetupService *SchoolSetupService, db *storage.FirestoreDB) *SetupAgent {
	return &SetupAgent{
		gptClient:          gptClient,
		schoolSetupService: schoolSetupService,
		db:                 db,
	}
}

// ChatSession represents an ongoing conversation with a school admin
type ChatSession struct {
	ID                  string                      `firestore:"id" json:"id"`
	AdminName           string                      `firestore:"admin_name" json:"admin_name"`
	AdminEmail          string                      `firestore:"admin_email" json:"admin_email"`
	SchoolName          string                      `firestore:"school_name" json:"school_name"`
	Status              string                      `firestore:"status" json:"status"` // active, completed, cancelled
	Messages            []ChatMessage               `firestore:"messages" json:"messages"`
	ExtractedInfo       ExtractedSchoolInfo         `firestore:"extracted_info" json:"extracted_info"`
	GeneratedConfig     *models.SchoolConfiguration `firestore:"generated_config,omitempty" json:"generated_config,omitempty"`
	ConfirmationPending bool                        `firestore:"confirmation_pending" json:"confirmation_pending"`
	CreatedAt           time.Time                   `firestore:"created_at" json:"created_at"`
	UpdatedAt           time.Time                   `firestore:"updated_at" json:"updated_at"`
}

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	ID        string    `firestore:"id" json:"id"`
	Sender    string    `firestore:"sender" json:"sender"` // admin, agent
	Content   string    `firestore:"content" json:"content"`
	Type      string    `firestore:"type" json:"type"` // text, suggestion, confirmation
	Timestamp time.Time `firestore:"timestamp" json:"timestamp"`
}

// ExtractedSchoolInfo contains information extracted from the conversation
type ExtractedSchoolInfo struct {
	BasicInfo      BasicSchoolInfo            `firestore:"basic_info" json:"basic_info"`
	RequiredFields map[string][]RequiredField `firestore:"required_fields" json:"required_fields"`
	Features       models.SchoolFeatures      `firestore:"features" json:"features"`
	CustomRequests []string                   `firestore:"custom_requests" json:"custom_requests"`
	Confidence     float64                    `firestore:"confidence" json:"confidence"`
}

// BasicSchoolInfo contains basic school information
type BasicSchoolInfo struct {
	Name            string `firestore:"name" json:"name"`
	Region          string `firestore:"region" json:"region"`
	EducationSystem string `firestore:"education_system" json:"education_system"`
	StudentCount    string `firestore:"student_count" json:"student_count"`
	GradeLevels     string `firestore:"grade_levels" json:"grade_levels"`
	Timezone        string `firestore:"timezone" json:"timezone"`
	Language        string `firestore:"language" json:"language"`
	Currency        string `firestore:"currency" json:"currency"`
}

// RequiredField represents a field requirement extracted from conversation
type RequiredField struct {
	Name     string   `firestore:"name" json:"name"`
	Type     string   `firestore:"type" json:"type"`
	Required bool     `firestore:"required" json:"required"`
	Options  []string `firestore:"options,omitempty" json:"options,omitempty"`
	Reason   string   `firestore:"reason" json:"reason"`
	Category string   `firestore:"category" json:"category"`
}

// StartChatSession initiates a new chat session with a school admin
func (agent *SetupAgent) StartChatSession(ctx context.Context, adminName, adminEmail string) (*ChatSession, error) {
	sessionID := fmt.Sprintf("chat_%d", time.Now().Unix())

	session := &ChatSession{
		ID:         sessionID,
		AdminName:  adminName,
		AdminEmail: adminEmail,
		Status:     "active",
		Messages:   []ChatMessage{},
		ExtractedInfo: ExtractedSchoolInfo{
			RequiredFields: make(map[string][]RequiredField),
			Features:       models.SchoolFeatures{},
		},
		ConfirmationPending: false,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Add welcome message
	welcomeMessage := agent.generateWelcomeMessage(adminName)
	session.Messages = append(session.Messages, ChatMessage{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Sender:    "agent",
		Content:   welcomeMessage,
		Type:      "text",
		Timestamp: time.Now(),
	})

	// Save session to database
	client := agent.db.GetClient()
	_, err := client.Collection("chat_sessions").Doc(sessionID).Set(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("error saving chat session: %v", err)
	}

	return session, nil
}

// ProcessMessage processes an admin's message and generates an AI response
func (agent *SetupAgent) ProcessMessage(ctx context.Context, sessionID, adminMessage string) (*ChatSession, error) {
	// Get current session
	session, err := agent.getChatSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Add admin message to session
	adminMsg := ChatMessage{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Sender:    "admin",
		Content:   adminMessage,
		Type:      "text",
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, adminMsg)

	// Generate AI response
	agentResponse, extractedInfo, err := agent.generateResponse(ctx, session, adminMessage)
	if err != nil {
		return nil, fmt.Errorf("error generating response: %v", err)
	}

	// Update extracted information
	if extractedInfo != nil {
		agent.mergeExtractedInfo(&session.ExtractedInfo, *extractedInfo)
	}

	// Add agent response to session
	agentMsg := ChatMessage{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Sender:    "agent",
		Content:   agentResponse,
		Type:      "text",
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, agentMsg)

	// Check if we have enough information to generate configuration
	if agent.hasEnoughInformation(session.ExtractedInfo) && !session.ConfirmationPending {
		config, err := agent.generateSchoolConfiguration(session.ExtractedInfo)
		if err != nil {
			return nil, fmt.Errorf("error generating configuration: %v", err)
		}

		session.GeneratedConfig = config
		session.ConfirmationPending = true

		// Add configuration preview message
		previewMsg := agent.generateConfigurationPreview(config, session.ExtractedInfo)
		configMsg := ChatMessage{
			ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
			Sender:    "agent",
			Content:   previewMsg,
			Type:      "confirmation",
			Timestamp: time.Now(),
		}
		session.Messages = append(session.Messages, configMsg)
	}

	session.UpdatedAt = time.Now()

	// Save updated session
	client := agent.db.GetClient()
	_, err = client.Collection("chat_sessions").Doc(sessionID).Set(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("error updating chat session: %v", err)
	}

	return session, nil
}

// ConfirmConfiguration creates the school in the database after admin confirmation
func (agent *SetupAgent) ConfirmConfiguration(ctx context.Context, sessionID string, confirmed bool) (*models.SchoolConfiguration, error) {
	session, err := agent.getChatSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if !session.ConfirmationPending {
		return nil, fmt.Errorf("no configuration pending confirmation")
	}

	if !confirmed {
		// Reset for modifications
		session.ConfirmationPending = false
		session.GeneratedConfig = nil

		modifyMsg := ChatMessage{
			ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
			Sender:    "agent",
			Content:   "No problem! Let's modify the configuration. What would you like to change? You can tell me about any specific fields you need, features you want to enable/disable, or any other requirements.",
			Type:      "text",
			Timestamp: time.Now(),
		}
		session.Messages = append(session.Messages, modifyMsg)

		// Save updated session
		client := agent.db.GetClient()
		_, err = client.Collection("chat_sessions").Doc(sessionID).Set(ctx, session)
		return nil, err
	}

	// Create the school configuration
	req := &CreateSchoolRequest{
		SchoolName:      session.GeneratedConfig.SchoolName,
		Region:          session.GeneratedConfig.Region,
		EducationSystem: session.GeneratedConfig.EducationSystem,
		Timezone:        session.GeneratedConfig.Timezone,
		Language:        session.GeneratedConfig.Language,
		Currency:        session.GeneratedConfig.Currency,
		Features:        session.GeneratedConfig.Features,
	}

	config, err := agent.schoolSetupService.CreateSchoolConfiguration(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error creating school configuration: %v", err)
	}

	// Add custom fields based on extracted requirements
	for entityType, fields := range session.ExtractedInfo.RequiredFields {
		for _, field := range fields {
			customField := models.CustomField{
				ID:          strings.ToLower(strings.ReplaceAll(field.Name, " ", "_")),
				Name:        field.Name,
				Key:         strings.ToLower(strings.ReplaceAll(field.Name, " ", "_")),
				Type:        models.FieldType(field.Type),
				Required:    field.Required,
				Options:     field.Options,
				Category:    field.Category,
				Description: field.Reason,
				Order:       20 + len(session.ExtractedInfo.RequiredFields[entityType]),
			}

			err = agent.schoolSetupService.AddCustomField(ctx, config.ID, entityType, customField)
			if err != nil {
				// Log error but don't fail the entire process
				fmt.Printf("Warning: failed to add custom field %s: %v\n", field.Name, err)
			}
		}
	}

	// Mark session as completed
	session.Status = "completed"
	session.ConfirmationPending = false
	session.UpdatedAt = time.Now()

	// Add completion message with data import suggestion
	completionMsg := ChatMessage{
		ID:     fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Sender: "agent",
		Content: fmt.Sprintf("🎉 Excellent! Your school '%s' has been successfully created with ID: %s\n\nYour school is configured with custom fields for:\n%s\n\n🚀 **Next Step: Populate Your School Data**\n\nWould you like me to:\n1. **Generate sample data** (teachers, students, parents) to get started quickly?\n2. **Create data templates** for you to fill manually?\n3. **Skip for now** and add users later?\n\nReply with '1', '2', or '3' to proceed!",
			config.SchoolName,
			config.ID,
			agent.formatSchoolFeatures(session.ExtractedInfo)),
		Type:      "text",
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, completionMsg)

	// Save final session
	client := agent.db.GetClient()
	_, err = client.Collection("chat_sessions").Doc(sessionID).Set(ctx, session)
	if err != nil {
		return config, err // Return config even if session save fails
	}

	return config, nil
}

// generateWelcomeMessage creates a personalized welcome message
func (agent *SetupAgent) generateWelcomeMessage(adminName string) string {
	return fmt.Sprintf(`Hello %s! 👋 Welcome to SchoolGPT Setup Assistant!

I'm here to help you configure your school's system perfectly. I'll ask you a few questions to understand your school's specific needs, and then automatically create a customized configuration.

Let's start with some basic information:

1. What's the name of your school?
2. Which country/region is your school located in?
3. What education system do you follow? (e.g., CBSE, ICSE, K-12, etc.)

Feel free to tell me everything in one message or we can go step by step. I'm here to make this as easy as possible for you! 😊`, adminName)
}

// generateResponse uses AI to understand requirements and ask follow-up questions
func (agent *SetupAgent) generateResponse(ctx context.Context, session *ChatSession, adminMessage string) (string, *ExtractedSchoolInfo, error) {
	// Build conversation context
	conversationHistory := agent.buildConversationContext(session)

	// Create prompt for AI
	prompt := fmt.Sprintf(`You are a SchoolGPT Setup Assistant helping a school administrator configure their school management system. 

Your role:
1. Extract school information from the admin's messages
2. Ask intelligent follow-up questions to understand their specific needs
3. Identify what custom fields they might need for students, teachers, and parents
4. Understand what features they want (attendance tracking, grade management, parent communication, etc.)
5. Be friendly, professional, and helpful

Current conversation:
%s

Admin's latest message: "%s"

Based on the conversation so far, provide:
1. A helpful response to continue the conversation
2. Extract any new information about their school requirements

Available field types: text, number, email, phone, date, select, boolean, textarea, file
Available features: attendance_tracking, grade_management, assignment_tracking, parent_communication, behavior_tracking, financial_management, transport_management, library_management, event_management, online_exams, ai_insights, multi_language_support

Respond in a conversational, helpful tone. Ask about specific needs like:
- What information do you track for students? (blood group, house assignment, etc.)
- Do you need financial management features?
- What communication methods do you use with parents?
- Do you track behavior or disciplinary actions?

If you have enough information to suggest a configuration, ask for confirmation before proceeding.

Response format should be natural conversation text.`, conversationHistory, adminMessage)

	// Get AI response with higher token limit for conversational responses
	response, err := agent.gptClient.GenerateResponse(ctx, prompt, 2500)
	if err != nil {
		return "", nil, fmt.Errorf("error generating AI response: %v", err)
	}

	// Extract information from the conversation
	extractedInfo, err := agent.extractInformationFromConversation(ctx, conversationHistory, adminMessage)
	if err != nil {
		// Continue even if extraction fails
		fmt.Printf("Warning: failed to extract information: %v\n", err)
	}

	return response, extractedInfo, nil
}

// extractInformationFromConversation uses AI to extract structured information
func (agent *SetupAgent) extractInformationFromConversation(ctx context.Context, conversation, latestMessage string) (*ExtractedSchoolInfo, error) {
	prompt := fmt.Sprintf(`Extract school setup information from this conversation between a SchoolGPT Setup Assistant and a school admin.

Conversation:
%s

Latest message: "%s"

Extract and return ONLY a JSON object with this exact structure:
{
  "basic_info": {
    "name": "school name if mentioned",
    "region": "country/region if mentioned", 
    "education_system": "education system if mentioned (cbse, icse, k-12, etc.)",
    "student_count": "approximate number of students if mentioned",
    "grade_levels": "grade levels if mentioned",
    "timezone": "timezone if mentioned or inferred from region",
    "language": "primary language if mentioned", 
    "currency": "currency if mentioned or inferred from region"
  },
  "required_fields": {
    "student": [
      {
        "name": "field name",
        "type": "field type (text, number, email, phone, date, select, boolean, textarea)",
        "required": true/false,
        "options": ["option1", "option2"] if select type,
        "reason": "why this field is needed",
        "category": "personal/academic/administrative/medical/government"
      }
    ],
    "teacher": [],
    "parent": []
  },
  "features": {
    "attendance_tracking": true/false,
    "grade_management": true/false,
    "assignment_tracking": true/false,
    "parent_communication": true/false,
    "behavior_tracking": true/false,
    "financial_management": true/false,
    "transport_management": true/false,
    "library_management": true/false,
    "event_management": true/false,
    "online_exams": true/false,
    "ai_insights": true/false,
    "multi_language_support": true/false
  },
  "custom_requests": ["any specific custom requirements mentioned"],
  "confidence": 0.0-1.0
}

Only include information that was explicitly mentioned or clearly implied. Use empty strings for missing basic info and false for unmentioned features.`, conversation, latestMessage)

	// Use higher token limit for JSON extraction (complex structured output)
	response, err := agent.gptClient.GenerateResponse(ctx, prompt, 3000)
	if err != nil {
		return nil, err
	}

	// Clean response to extract JSON
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}") + 1
	if jsonStart == -1 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("no valid JSON found in response")
	}
	jsonStr := response[jsonStart:jsonEnd]

	// Parse JSON response
	var extractedInfo ExtractedSchoolInfo
	err = json.Unmarshal([]byte(jsonStr), &extractedInfo)
	if err != nil {
		return nil, fmt.Errorf("error parsing extracted information: %v", err)
	}

	return &extractedInfo, nil
}

// Helper functions

func (agent *SetupAgent) buildConversationContext(session *ChatSession) string {
	var context strings.Builder
	for _, msg := range session.Messages {
		if msg.Sender == "admin" {
			context.WriteString(fmt.Sprintf("Admin: %s\n", msg.Content))
		} else {
			context.WriteString(fmt.Sprintf("Assistant: %s\n", msg.Content))
		}
	}
	return context.String()
}

func (agent *SetupAgent) mergeExtractedInfo(existing *ExtractedSchoolInfo, new ExtractedSchoolInfo) {
	// Merge basic info
	if new.BasicInfo.Name != "" {
		existing.BasicInfo.Name = new.BasicInfo.Name
	}
	if new.BasicInfo.Region != "" {
		existing.BasicInfo.Region = new.BasicInfo.Region
	}
	if new.BasicInfo.EducationSystem != "" {
		existing.BasicInfo.EducationSystem = new.BasicInfo.EducationSystem
	}
	if new.BasicInfo.StudentCount != "" {
		existing.BasicInfo.StudentCount = new.BasicInfo.StudentCount
	}
	if new.BasicInfo.GradeLevels != "" {
		existing.BasicInfo.GradeLevels = new.BasicInfo.GradeLevels
	}
	if new.BasicInfo.Timezone != "" {
		existing.BasicInfo.Timezone = new.BasicInfo.Timezone
	}
	if new.BasicInfo.Language != "" {
		existing.BasicInfo.Language = new.BasicInfo.Language
	}
	if new.BasicInfo.Currency != "" {
		existing.BasicInfo.Currency = new.BasicInfo.Currency
	}

	// Merge required fields
	for entityType, fields := range new.RequiredFields {
		if existing.RequiredFields == nil {
			existing.RequiredFields = make(map[string][]RequiredField)
		}
		existing.RequiredFields[entityType] = fields
	}

	// Merge features (only update if new feature is true)
	if new.Features.AttendanceTracking {
		existing.Features.AttendanceTracking = true
	}
	if new.Features.GradeManagement {
		existing.Features.GradeManagement = true
	}
	if new.Features.AssignmentTracking {
		existing.Features.AssignmentTracking = true
	}
	if new.Features.ParentCommunication {
		existing.Features.ParentCommunication = true
	}
	if new.Features.BehaviorTracking {
		existing.Features.BehaviorTracking = true
	}
	if new.Features.FinancialManagement {
		existing.Features.FinancialManagement = true
	}
	if new.Features.TransportManagement {
		existing.Features.TransportManagement = true
	}
	if new.Features.LibraryManagement {
		existing.Features.LibraryManagement = true
	}
	if new.Features.EventManagement {
		existing.Features.EventManagement = true
	}
	if new.Features.OnlineExams {
		existing.Features.OnlineExams = true
	}
	if new.Features.AIInsights {
		existing.Features.AIInsights = true
	}
	if new.Features.MultiLanguageSupport {
		existing.Features.MultiLanguageSupport = true
	}

	// Merge custom requests
	existing.CustomRequests = append(existing.CustomRequests, new.CustomRequests...)

	// Update confidence
	if new.Confidence > existing.Confidence {
		existing.Confidence = new.Confidence
	}
}

func (agent *SetupAgent) hasEnoughInformation(info ExtractedSchoolInfo) bool {
	return info.BasicInfo.Name != "" &&
		info.BasicInfo.Region != "" &&
		info.BasicInfo.EducationSystem != "" &&
		info.Confidence > 0.7
}

func (agent *SetupAgent) generateSchoolConfiguration(info ExtractedSchoolInfo) (*models.SchoolConfiguration, error) {
	// Get education system template
	template, exists := models.EducationSystemTemplates[info.BasicInfo.EducationSystem]
	if !exists {
		template = models.EducationSystemTemplates["cbse"] // Default
	}

	// Set defaults for missing information
	timezone := info.BasicInfo.Timezone
	if timezone == "" {
		if strings.Contains(strings.ToLower(info.BasicInfo.Region), "india") {
			timezone = "Asia/Kolkata"
		} else {
			timezone = "UTC"
		}
	}

	language := info.BasicInfo.Language
	if language == "" {
		language = "en"
	}

	currency := info.BasicInfo.Currency
	if currency == "" {
		if strings.Contains(strings.ToLower(info.BasicInfo.Region), "india") {
			currency = "INR"
		} else {
			currency = "USD"
		}
	}

	// Create school configuration
	config := &models.SchoolConfiguration{
		SchoolName:      info.BasicInfo.Name,
		Region:          info.BasicInfo.Region,
		EducationSystem: info.BasicInfo.EducationSystem,
		Timezone:        timezone,
		Language:        language,
		Currency:        currency,
		AcademicYear:    template.AcademicYear,
		Features:        info.Features,
		Status:          "setup",
		SetupStep:       1,
	}

	return config, nil
}

func (agent *SetupAgent) generateConfigurationPreview(config *models.SchoolConfiguration, info ExtractedSchoolInfo) string {
	return fmt.Sprintf(`🎯 Perfect! Based on our conversation, I've prepared your school configuration:

**School Information:**
• Name: %s
• Region: %s
• Education System: %s
• Academic Year: %s to %s
• Language: %s
• Currency: %s

**Enabled Features:**
%s

**Custom Fields to be Added:**
%s

This configuration will be created in your database with all the custom fields you mentioned. Does this look correct? 

Reply with:
• "Yes" or "Confirm" to create the school
• "No" or "Modify" to make changes
• Tell me specifically what you'd like to change`,
		config.SchoolName,
		config.Region,
		config.EducationSystem,
		config.AcademicYear.StartDate,
		config.AcademicYear.EndDate,
		config.Language,
		config.Currency,
		agent.formatFeatures(config.Features),
		agent.formatCustomFields(info.RequiredFields))
}

func (agent *SetupAgent) formatFeatures(features models.SchoolFeatures) string {
	var enabled []string
	if features.AttendanceTracking {
		enabled = append(enabled, "✅ Attendance Tracking")
	}
	if features.GradeManagement {
		enabled = append(enabled, "✅ Grade Management")
	}
	if features.AssignmentTracking {
		enabled = append(enabled, "✅ Assignment Tracking")
	}
	if features.ParentCommunication {
		enabled = append(enabled, "✅ Parent Communication")
	}
	if features.BehaviorTracking {
		enabled = append(enabled, "✅ Behavior Tracking")
	}
	if features.FinancialManagement {
		enabled = append(enabled, "✅ Financial Management")
	}
	if features.TransportManagement {
		enabled = append(enabled, "✅ Transport Management")
	}
	if features.LibraryManagement {
		enabled = append(enabled, "✅ Library Management")
	}
	if features.EventManagement {
		enabled = append(enabled, "✅ Event Management")
	}
	if features.OnlineExams {
		enabled = append(enabled, "✅ Online Exams")
	}
	if features.AIInsights {
		enabled = append(enabled, "✅ AI Insights")
	}
	if features.MultiLanguageSupport {
		enabled = append(enabled, "✅ Multi-Language Support")
	}

	if len(enabled) == 0 {
		return "• Basic school management features"
	}

	return strings.Join(enabled, "\n")
}

func (agent *SetupAgent) formatCustomFields(requiredFields map[string][]RequiredField) string {
	if len(requiredFields) == 0 {
		return "• Standard fields (name, email, etc.)"
	}

	var result strings.Builder
	for entityType, fields := range requiredFields {
		if len(fields) > 0 {
			result.WriteString(fmt.Sprintf("\n**%s Fields:**\n", strings.Title(entityType)))
			for _, field := range fields {
				result.WriteString(fmt.Sprintf("• %s (%s) - %s\n", field.Name, field.Type, field.Reason))
			}
		}
	}

	if result.Len() == 0 {
		return "• Standard fields (name, email, etc.)"
	}

	return result.String()
}

func (agent *SetupAgent) getChatSession(ctx context.Context, sessionID string) (*ChatSession, error) {
	client := agent.db.GetClient()
	doc, err := client.Collection("chat_sessions").Doc(sessionID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting chat session: %v", err)
	}

	var session ChatSession
	if err := doc.DataTo(&session); err != nil {
		return nil, fmt.Errorf("error parsing chat session: %v", err)
	}

	return &session, nil
}

// GetChatSession is a public method to get chat session (for handlers)
func (agent *SetupAgent) GetChatSession(ctx context.Context, sessionID string) (*ChatSession, error) {
	return agent.getChatSession(ctx, sessionID)
}

// formatSchoolFeatures formats the extracted school information for display
func (agent *SetupAgent) formatSchoolFeatures(info ExtractedSchoolInfo) string {
	var features []string

	// Add basic info
	features = append(features, fmt.Sprintf("• School: %s (%s)", info.BasicInfo.Name, info.BasicInfo.EducationSystem))

	// Add custom fields
	for entityType, fields := range info.RequiredFields {
		for _, field := range fields {
			features = append(features, fmt.Sprintf("• %s %s: %s", strings.Title(entityType), field.Name, field.Reason))
		}
	}

	// Add enabled features
	var enabledFeatures []string
	if info.Features.FinancialManagement {
		enabledFeatures = append(enabledFeatures, "Financial Management")
	}
	if info.Features.TransportManagement {
		enabledFeatures = append(enabledFeatures, "Transport Management")
	}
	if info.Features.AIInsights {
		enabledFeatures = append(enabledFeatures, "AI Insights")
	}
	if info.Features.LibraryManagement {
		enabledFeatures = append(enabledFeatures, "Library Management")
	}
	if info.Features.BehaviorTracking {
		enabledFeatures = append(enabledFeatures, "Behavior Tracking")
	}

	if len(enabledFeatures) > 0 {
		features = append(features, fmt.Sprintf("• Features: %s", strings.Join(enabledFeatures, ", ")))
	}

	return strings.Join(features, "\n")
}
