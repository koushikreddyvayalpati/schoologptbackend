package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🎓 Testing SchoolGPT Teacher Automation Features")
	fmt.Println(repeat("=", 60))

	baseURL := "http://localhost:8080"
	teacherToken := "test_teacher_token" // Ms. Sarah Johnson

	// Test 1: Mark student absent with reason
	fmt.Println("\n📝 Test 1: Mark Student Absent with AI Insights")
	markAbsent := map[string]interface{}{
		"student_id": "student_003", // Ryan Wilson
		"class_id":   "class_001",   // Algebra II
		"status":     "absent",
		"reason":     "Family emergency",
	}

	result1 := makeRequest("POST", baseURL+"/api/v1/teacher/mark-attendance", teacherToken, markAbsent)
	fmt.Printf("✅ Result: %s\n", result1)

	time.Sleep(2 * time.Second)

	// Test 2: Analyze student performance
	fmt.Println("\n🔍 Test 2: AI-Powered Student Analysis")
	analyzeStudent := map[string]interface{}{
		"student_id": "student_003", // Ryan Wilson
		"subject":    "Mathematics",
		"days":       30,
	}

	result2 := makeRequest("POST", baseURL+"/api/v1/teacher/analyze-student", teacherToken, analyzeStudent)
	fmt.Printf("✅ Result: %s\n", result2)

	time.Sleep(2 * time.Second)

	// Test 3: Send parent notification
	fmt.Println("\n📧 Test 3: AI-Enhanced Parent Notification")
	notifyParent := map[string]interface{}{
		"student_id": "student_003", // Ryan Wilson
		"message":    "Ryan was absent today due to family emergency. Please ensure he catches up on today's algebra lesson.",
		"type":       "attendance",
		"urgent":     false,
	}

	result3 := makeRequest("POST", baseURL+"/api/v1/teacher/notify-parent", teacherToken, notifyParent)
	fmt.Printf("✅ Result: %s\n", result3)

	time.Sleep(2 * time.Second)

	// Test 4: Mark student present
	fmt.Println("\n✅ Test 4: Mark Student Present")
	markPresent := map[string]interface{}{
		"student_id": "student_001", // Alex Smith
		"class_id":   "class_001",   // Algebra II
		"status":     "present",
	}

	result4 := makeRequest("POST", baseURL+"/api/v1/teacher/mark-attendance", teacherToken, markPresent)
	fmt.Printf("✅ Result: %s\n", result4)

	time.Sleep(2 * time.Second)

	// Test 5: Analyze another student
	fmt.Println("\n📊 Test 5: Analyze High-Performing Student")
	analyzeGoodStudent := map[string]interface{}{
		"student_id": "student_001", // Alex Smith
		"subject":    "Mathematics",
		"days":       30,
	}

	result5 := makeRequest("POST", baseURL+"/api/v1/teacher/analyze-student", teacherToken, analyzeGoodStudent)
	fmt.Printf("✅ Result: %s\n", result5)

	fmt.Println("\n🎉 Teacher Automation Tests Complete!")
	fmt.Println("\n💡 Key Features Demonstrated:")
	fmt.Println("  • Automated attendance marking with AI insights")
	fmt.Println("  • Comprehensive student performance analysis")
	fmt.Println("  • AI-enhanced parent communication")
	fmt.Println("  • Automatic parent notifications for absences")
	fmt.Println("  • Actionable recommendations for teachers")

	fmt.Println("\n🔧 Real-world Integration Points:")
	fmt.Println("  • Email/SMS services (SendGrid, Twilio)")
	fmt.Println("  • WhatsApp Business API")
	fmt.Println("  • School management systems")
	fmt.Println("  • Gradebook integration")
	fmt.Println("  • Parent portal updates")
}

func makeRequest(method, url, token string, payload map[string]interface{}) string {
	var body io.Reader
	if payload != nil {
		jsonData, _ := json.Marshal(payload)
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Sprintf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response: %v", err)
	}

	// Pretty print JSON response
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, responseBody, "", "  ")
	if err != nil {
		return string(responseBody)
	}

	return fmt.Sprintf("Status: %d\n%s", resp.StatusCode, prettyJSON.String())
}

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
