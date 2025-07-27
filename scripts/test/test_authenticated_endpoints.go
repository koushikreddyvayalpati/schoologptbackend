package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TestCase represents a test case for API endpoints
type TestCase struct {
	Name         string
	Method       string
	Endpoint     string
	Token        string
	Payload      map[string]interface{}
	ExpectedCode int
	Description  string
}

func main() {
	fmt.Println("🧪 Testing SchoolGPT Authenticated Endpoints")
	fmt.Println(repeat("=", 50))

	baseURL := "http://localhost:8080"

	// Define test cases
	testCases := []TestCase{
		// Test 1: Admin asking about all students
		{
			Name:     "Admin - View All Students",
			Method:   "POST",
			Endpoint: "/api/v1/gpt/ask",
			Token:    "test_admin_token",
			Payload: map[string]interface{}{
				"query": "Show me a summary of all students and their attendance rates",
			},
			ExpectedCode: 200,
			Description:  "Admin should have access to all student data",
		},

		// Test 2: Teacher asking about their students
		{
			Name:     "Teacher - View Own Students",
			Method:   "POST",
			Endpoint: "/api/v1/gpt/ask",
			Token:    "test_teacher_token",
			Payload: map[string]interface{}{
				"query": "Show me attendance for my students in Algebra II",
			},
			ExpectedCode: 200,
			Description:  "Teacher should see their own class data",
		},

		// Test 3: Parent asking about their child
		{
			Name:     "Parent - View Own Child",
			Method:   "POST",
			Endpoint: "/api/v1/gpt/ask",
			Token:    "test_parent_token",
			Payload: map[string]interface{}{
				"query": "How is my child Ryan Wilson doing in school? Show me his attendance",
			},
			ExpectedCode: 200,
			Description:  "Parent should see their own child's data",
		},

		// Test 4: Teacher using attendance endpoint
		{
			Name:     "Teacher - Attendance Query",
			Method:   "POST",
			Endpoint: "/api/v1/gpt/attendance",
			Token:    "test_teacher_token",
			Payload: map[string]interface{}{
				"query": "Who was absent in my classes yesterday?",
			},
			ExpectedCode: 200,
			Description:  "Teacher should access attendance data through specialized endpoint",
		},

		// Test 5: Parent trying to access teacher endpoint (should fail)
		{
			Name:     "Parent - Unauthorized Access",
			Method:   "POST",
			Endpoint: "/api/v1/gpt/attendance",
			Token:    "test_parent_token",
			Payload: map[string]interface{}{
				"query": "Show me all attendance data",
			},
			ExpectedCode: 403,
			Description:  "Parent should NOT have access to teacher-only endpoints",
		},

		// Test 6: No token (should fail)
		{
			Name:     "No Authentication",
			Method:   "POST",
			Endpoint: "/api/v1/gpt/ask",
			Token:    "",
			Payload: map[string]interface{}{
				"query": "Tell me about students",
			},
			ExpectedCode: 401,
			Description:  "Should fail without authentication token",
		},

		// Test 7: Invalid token (should fail)
		{
			Name:     "Invalid Token",
			Method:   "POST",
			Endpoint: "/api/v1/gpt/ask",
			Token:    "invalid_token_123",
			Payload: map[string]interface{}{
				"query": "Tell me about students",
			},
			ExpectedCode: 401,
			Description:  "Should fail with invalid token",
		},

		// Test 8: Admin accessing all features
		{
			Name:     "Admin - Full Access",
			Method:   "POST",
			Endpoint: "/api/v1/gpt/ask",
			Token:    "test_admin_token",
			Payload: map[string]interface{}{
				"query": "Generate a comprehensive school report including all teachers, students, and attendance statistics",
			},
			ExpectedCode: 200,
			Description:  "Admin should have full system access",
		},
	}

	// Run tests
	fmt.Printf("\n🏃 Running %d test cases...\n\n", len(testCases))

	passed := 0
	failed := 0

	for i, test := range testCases {
		fmt.Printf("Test %d: %s\n", i+1, test.Name)
		fmt.Printf("📝 %s\n", test.Description)

		success := runTest(baseURL, test)
		if success {
			passed++
			fmt.Printf("✅ PASSED\n")
		} else {
			failed++
			fmt.Printf("❌ FAILED\n")
		}

		fmt.Println(repeat("-", 40))
		time.Sleep(1 * time.Second) // Brief pause between tests
	}

	// Summary
	fmt.Printf("\n🎯 Test Results Summary:\n")
	fmt.Printf("  ✅ Passed: %d\n", passed)
	fmt.Printf("  ❌ Failed: %d\n", failed)
	fmt.Printf("  📊 Success Rate: %.1f%%\n", float64(passed)/float64(len(testCases))*100)

	if failed == 0 {
		fmt.Println("\n🎉 All tests passed! Authentication system is working perfectly.")
	} else {
		fmt.Printf("\n⚠️  %d tests failed. Please check the authentication system.\n", failed)
	}
}

func runTest(baseURL string, test TestCase) bool {
	// Prepare the request
	var body io.Reader
	if test.Payload != nil {
		jsonData, _ := json.Marshal(test.Payload)
		body = bytes.NewBuffer(jsonData)
	}

	// Create the request
	req, err := http.NewRequest(test.Method, baseURL+test.Endpoint, body)
	if err != nil {
		fmt.Printf("❌ Error creating request: %v\n", err)
		return false
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if test.Token != "" {
		req.Header.Set("Authorization", "Bearer "+test.Token)
	}

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Error making request: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Error reading response: %v\n", err)
		return false
	}

	// Check status code
	if resp.StatusCode != test.ExpectedCode {
		fmt.Printf("❌ Expected status %d, got %d\n", test.ExpectedCode, resp.StatusCode)
		fmt.Printf("📄 Response: %s\n", string(responseBody))
		return false
	}

	// Print response for successful tests (200s)
	if resp.StatusCode == 200 {
		var response map[string]interface{}
		if err := json.Unmarshal(responseBody, &response); err == nil {
			if answer, ok := response["response"].(string); ok {
				// Truncate long responses for readability
				if len(answer) > 200 {
					answer = answer[:200] + "..."
				}
				fmt.Printf("💬 Response: %s\n", answer)
			}
		}
	} else {
		// Print error responses
		fmt.Printf("📄 Response: %s\n", string(responseBody))
	}

	return true
}

// Helper function to repeat a string (since Go doesn't have it built-in)
func repeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	result := make([]byte, 0, len(s)*count)
	for i := 0; i < count; i++ {
		result = append(result, s...)
	}
	return string(result)
}
