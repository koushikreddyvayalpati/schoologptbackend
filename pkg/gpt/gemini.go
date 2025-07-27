package gpt

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"github.com/schoolgpt/backend/internal/config"
	"google.golang.org/api/option"
)

// GeminiClient handles interactions with Google Gemini API
type GeminiClient struct {
	client *genai.Client
	config *config.Config
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(cfg *config.Config) (*GeminiClient, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GeminiAPIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	return &GeminiClient{
		client: client,
		config: cfg,
	}, nil
}

// HandleGPTQuery processes a query using Gemini (compatible with existing interface)
func (gc *GeminiClient) HandleGPTQuery(ctx context.Context, query string) (string, error) {
	// Use Gemini 2.5 Flash for the generous free tier (1,500 requests/day)
	model := gc.client.GenerativeModel("gemini-2.5-flash")

	// Configure for free tier optimization
	model.SetTemperature(0.3)     // Consistent responses
	model.SetMaxOutputTokens(800) // Increased tokens for better responses

	// Create school-focused prompt
	prompt := fmt.Sprintf(`You are a school AI assistant. Be concise and helpful.

User: %s

Please provide a helpful, educational response.`, query)

	// Call Gemini API (FREE!)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("error generating content with Gemini: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	// Extract the response text
	candidate := resp.Candidates[0]
	if candidate.Content == nil {
		return "", fmt.Errorf("empty response content")
	}

	var responseText string
	for _, part := range candidate.Content.Parts {
		// Handle different part types
		switch p := part.(type) {
		case genai.Text:
			responseText += string(p)
		default:
			// Handle other types or convert to string
			responseText += fmt.Sprintf("%v", p)
		}
	}

	if responseText == "" {
		return "", fmt.Errorf("no text content in response")
	}

	return responseText, nil
}

// GetAttendanceWithGPT handles attendance queries using Gemini
func (gc *GeminiClient) GetAttendanceWithGPT(ctx context.Context, query string) (map[string]interface{}, error) {
	// Use Gemini 2.5 Flash for attendance queries
	model := gc.client.GenerativeModel("gemini-2.5-flash")

	// Configure for structured output
	model.SetTemperature(0.3)
	model.SetMaxOutputTokens(800)

	// Create attendance-specific prompt
	attendancePrompt := fmt.Sprintf(`You are SchoolGPT's attendance assistant. Analyze this query: "%s"

Provide a helpful response about attendance management, patterns, or data requirements.
Be educational and specific to school contexts.`, query)

	resp, err := model.GenerateContent(ctx, genai.Text(attendancePrompt))
	if err != nil {
		return nil, fmt.Errorf("error generating attendance response: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no attendance response generated")
	}

	// Extract response text
	candidate := resp.Candidates[0]
	if candidate.Content == nil {
		return nil, fmt.Errorf("empty attendance response content")
	}

	var responseText string
	for _, part := range candidate.Content.Parts {
		// Handle different part types
		switch p := part.(type) {
		case genai.Text:
			responseText += string(p)
		default:
			// Handle other types or convert to string
			responseText += fmt.Sprintf("%v", p)
		}
	}

	// Return in expected format
	return map[string]interface{}{
		"response": responseText,
		"query":    query,
		"model":    "gemini-2.5-flash",
		"provider": "google-gemini-free",
	}, nil
}

// GenerateResponseWithTokens generates a response with specified max tokens
func (gc *GeminiClient) GenerateResponseWithTokens(ctx context.Context, prompt string, maxTokens int) (string, error) {
	// Use Gemini 2.5 Flash for the generous free tier
	model := gc.client.GenerativeModel("gemini-2.5-flash")

	// Configure with the specified max tokens (be more generous for complex responses)
	model.SetTemperature(0.3)
	model.SetMaxOutputTokens(int32(maxTokens))

	// Call Gemini API directly with the prompt
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("error generating content with Gemini: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	// Extract the response text
	candidate := resp.Candidates[0]
	if candidate.Content == nil {
		return "", fmt.Errorf("empty response content")
	}

	var responseText string
	for _, part := range candidate.Content.Parts {
		// Handle different part types
		switch p := part.(type) {
		case genai.Text:
			responseText += string(p)
		default:
			// Handle other types or convert to string
			responseText += fmt.Sprintf("%v", p)
		}
	}

	if responseText == "" {
		return "", fmt.Errorf("no text content in response")
	}

	return responseText, nil
}

// Close closes the Gemini client
func (gc *GeminiClient) Close() error {
	return gc.client.Close()
}
