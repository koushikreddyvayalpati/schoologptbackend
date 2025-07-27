package gpt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/schoolgpt/backend/internal/config"
)

// Client handles interactions with AI models (OpenAI and Gemini)
type Client struct {
	openaiClient *openai.Client
	geminiClient *GeminiClient
	config       *config.Config
	useGemini    bool
}

// FunctionCall represents a function call from GPT
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// New creates a new AI client (prefers Gemini for free usage)
func New(cfg *config.Config) *Client {
	c := &Client{
		config: cfg,
	}

	// Prefer Gemini if API key is available (generous free tier)
	if cfg.GeminiAPIKey != "" {
		geminiClient, err := NewGeminiClient(cfg)
		if err == nil {
			c.geminiClient = geminiClient
			c.useGemini = true
			return c
		}
		// If Gemini fails, fall back to OpenAI
	}

	// Fallback to OpenAI
	if cfg.OpenAIAPIKey != "" {
		c.openaiClient = openai.NewClient(cfg.OpenAIAPIKey)
		c.useGemini = false
	}

	return c
}

// HandleGPTQuery handles a query using the preferred AI model (Gemini or OpenAI)
func (c *Client) HandleGPTQuery(ctx context.Context, query string) (string, error) {
	// Use Gemini if available (free tier with 1,500 requests/day)
	if c.useGemini && c.geminiClient != nil {
		return c.geminiClient.HandleGPTQuery(ctx, query)
	}

	// Fallback to OpenAI
	if c.openaiClient != nil {
		return c.handleOpenAIQuery(ctx, query)
	}

	return "", errors.New("no AI client available - please configure GEMINI_API_KEY or OPENAI_API_KEY")
}

// handleOpenAIQuery handles OpenAI-specific requests
func (c *Client) handleOpenAIQuery(ctx context.Context, query string) (string, error) {
	resp, err := c.openaiClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4.1-nano", // Using the ultra-cost-effective model
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a school AI assistant. Be concise and helpful.", // Shorter prompt = less cost
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: query,
				},
			},
			MaxTokens:   300, // Limit response length = lower cost
			Temperature: 0.3, // Lower temperature = more consistent, cheaper
		},
	)

	if err != nil {
		return "", fmt.Errorf("error creating chat completion: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// CallWithFunctionSchema calls GPT with a function schema (optimized for cost)
func (c *Client) CallWithFunctionSchema(ctx context.Context, query string, functionName string, parameters map[string]interface{}) (*FunctionCall, error) {
	// Define function schema
	functions := []openai.FunctionDefinition{
		{
			Name:        functionName,
			Description: fmt.Sprintf("Call %s with parameters", functionName), // Shorter description
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": parameters,
				"required":   getRequiredFields(parameters),
			},
		},
	}

	// Call OpenAI with function calling enabled
	resp, err := c.openaiClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4.1-nano", // Ultra-cost-effective model
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "Use the provided function when appropriate.", // Minimal system prompt
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: query,
				},
			},
			Functions:    functions,
			FunctionCall: "auto",
			MaxTokens:    200, // Lower limit for function calls
			Temperature:  0.1, // Very consistent for function calling
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error creating chat completion: %v", err)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response from GPT")
	}

	// Check if GPT called a function
	if resp.Choices[0].Message.FunctionCall == nil {
		return nil, fmt.Errorf("GPT did not call the function, response: %s", resp.Choices[0].Message.Content)
	}

	// Parse function call arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.FunctionCall.Arguments), &args); err != nil {
		return nil, fmt.Errorf("error parsing function arguments: %v", err)
	}

	return &FunctionCall{
		Name:      resp.Choices[0].Message.FunctionCall.Name,
		Arguments: args,
	}, nil
}

// GetAttendanceWithGPT handles attendance queries using the preferred AI model
func (c *Client) GetAttendanceWithGPT(ctx context.Context, query string) (map[string]interface{}, error) {
	// Use Gemini if available (simpler approach, no function calling needed)
	if c.useGemini && c.geminiClient != nil {
		return c.geminiClient.GetAttendanceWithGPT(ctx, query)
	}

	// Fallback to OpenAI with function calling
	if c.openaiClient != nil {
		return c.handleOpenAIAttendanceQuery(ctx, query)
	}

	return nil, errors.New("no AI client available - please configure GEMINI_API_KEY or OPENAI_API_KEY")
}

// handleOpenAIAttendanceQuery handles OpenAI-specific attendance queries
func (c *Client) handleOpenAIAttendanceQuery(ctx context.Context, query string) (map[string]interface{}, error) {
	// Define parameters for get_attendance function
	parameters := map[string]interface{}{
		"student_id": map[string]interface{}{
			"type":        "string",
			"description": "Student ID",
		},
		"date": map[string]interface{}{
			"type":        "string",
			"description": "Date in YYYY-MM-DD format",
		},
	}

	// Call OpenAI with function schema
	functionCall, err := c.CallWithFunctionSchema(ctx, query, "get_attendance", parameters)
	if err != nil {
		return nil, err
	}

	// Mock function to fetch attendance
	studentID, _ := functionCall.Arguments["student_id"].(string)
	date, _ := functionCall.Arguments["date"].(string)

	// This would call an actual service in a real application
	attendanceData := mockFetchAttendance(studentID, date)
	return attendanceData, nil
}

// Helper function to get required fields from parameters
func getRequiredFields(parameters map[string]interface{}) []string {
	var required []string
	for name := range parameters {
		required = append(required, name)
	}
	return required
}

// IsUsingGemini returns true if Gemini is being used
func (c *Client) IsUsingGemini() bool {
	return c.useGemini
}

// IsUsingOpenAI returns true if OpenAI is being used
func (c *Client) IsUsingOpenAI() bool {
	return !c.useGemini && c.openaiClient != nil
}

// GetModelName returns the name of the model being used
func (c *Client) GetModelName() string {
	if c.useGemini {
		return "gemini-2.5-flash"
	}
	return "gpt-4.1-nano"
}

// GenerateResponse generates a response with optional max tokens (for setup agent compatibility)
func (c *Client) GenerateResponse(ctx context.Context, prompt string, maxTokens int) (string, error) {
	// Use Gemini if available
	if c.useGemini && c.geminiClient != nil {
		return c.geminiClient.GenerateResponseWithTokens(ctx, prompt, maxTokens)
	}

	// Fallback to OpenAI
	if c.openaiClient != nil {
		return c.generateOpenAIResponse(ctx, prompt, maxTokens)
	}

	return "", errors.New("no AI client available - please configure GEMINI_API_KEY or OPENAI_API_KEY")
}

// generateOpenAIResponse handles OpenAI-specific response generation
func (c *Client) generateOpenAIResponse(ctx context.Context, prompt string, maxTokens int) (string, error) {
	resp, err := c.openaiClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4.1-nano",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   maxTokens,
			Temperature: 0.3,
		},
	)

	if err != nil {
		return "", fmt.Errorf("error creating chat completion: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// Mock function to fetch attendance
func mockFetchAttendance(studentID, date string) map[string]interface{} {
	return map[string]interface{}{
		"student_id": studentID,
		"date":       date,
		"status":     "present",
		"marked_by":  "Teacher A",
	}
}
