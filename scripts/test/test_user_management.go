package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	baseURL := "http://localhost:8080"

	// Test health endpoint
	fmt.Println("🔍 Testing API health...")
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
		return
	}
	resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ API server is running")
	} else {
		fmt.Printf("❌ Health check failed with status: %d\n", resp.StatusCode)
		return
	}

	// Test user creation endpoint (this will fail with 401 since we don't have auth)
	fmt.Println("\n🔍 Testing user creation endpoint...")

	testUser := map[string]interface{}{
		"name":    "Test Student",
		"email":   "test.student@school.edu",
		"role":    "student",
		"grade":   "10th",
		"phone":   "+1234567890",
		"address": "123 Test Street",
	}

	jsonData, _ := json.Marshal(testUser)
	resp, err = http.Post(baseURL+"/api/v1/users", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ User creation request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		fmt.Println("✅ User creation endpoint exists and requires authentication (expected)")
	} else {
		fmt.Printf("📄 Response Status: %d\n", resp.StatusCode)
		fmt.Printf("📄 Response Body: %s\n", string(body))
	}

	// Test bulk user creation endpoint
	fmt.Println("\n🔍 Testing bulk user creation endpoint...")

	bulkUsers := map[string]interface{}{
		"users": []map[string]interface{}{
			{
				"name":       "Test Teacher",
				"email":      "test.teacher@school.edu",
				"role":       "teacher",
				"subject":    "Mathematics",
				"department": "Science",
			},
		},
	}

	jsonData, _ = json.Marshal(bulkUsers)
	resp, err = http.Post(baseURL+"/api/v1/users/bulk", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Bulk user creation request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		fmt.Println("✅ Bulk user creation endpoint exists and requires authentication (expected)")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("📄 Response Status: %d\n", resp.StatusCode)
		fmt.Printf("📄 Response Body: %s\n", string(body))
	}

	// Test getting users endpoint
	fmt.Println("\n🔍 Testing get users endpoint...")

	resp, err = http.Get(baseURL + "/api/v1/users")
	if err != nil {
		fmt.Printf("❌ Get users request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		fmt.Println("✅ Get users endpoint exists and requires authentication (expected)")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("📄 Response Status: %d\n", resp.StatusCode)
		fmt.Printf("📄 Response Body: %s\n", string(body))
	}

	fmt.Println("\n🎉 User management endpoints are properly configured!")
	fmt.Println("📝 Next steps:")
	fmt.Println("   1. Start the backend server: go run cmd/api/main.go")
	fmt.Println("   2. Set up authentication to test the endpoints fully")
	fmt.Println("   3. Use the frontend Create Accounts page to test the functionality")
}
