package main

import (
	"context"
	"fmt"
	"log"
	"time"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// User represents a user with authentication and role information
type User struct {
	UID       string    `firestore:"uid"`
	Email     string    `firestore:"email"`
	Name      string    `firestore:"name"`
	Role      string    `firestore:"role"`      // teacher, parent, admin, student
	EntityID  string    `firestore:"entity_id"` // teacher_001, parent_001, etc.
	Active    bool      `firestore:"active"`
	CreatedAt time.Time `firestore:"created_at"`
	LastLogin time.Time `firestore:"last_login"`
}

// UserRole represents role-based permissions
type UserRole struct {
	Role        string   `firestore:"role"`
	Permissions []string `firestore:"permissions"`
	Description string   `firestore:"description"`
}

func main() {
	fmt.Println("🔐 Adding Authentication Roles to SchoolGPT Database")

	ctx := context.Background()

	// Initialize Firebase
	opt := option.WithCredentialsFile("firebase-credentials.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Error initializing app: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error getting Firestore client: %v", err)
	}
	defer client.Close()

	// Add Role Definitions
	fmt.Println("\n📋 Setting up role definitions...")

	roles := []UserRole{
		{
			Role: "admin",
			Permissions: []string{
				"view_all_students",
				"view_all_teachers",
				"view_all_attendance",
				"manage_users",
				"generate_reports",
				"system_admin",
			},
			Description: "School administrator with full access",
		},
		{
			Role: "teacher",
			Permissions: []string{
				"view_own_classes",
				"view_class_students",
				"mark_attendance",
				"view_student_attendance",
				"communicate_parents",
				"generate_class_reports",
			},
			Description: "Teacher with access to their classes and students",
		},
		{
			Role: "parent",
			Permissions: []string{
				"view_own_children",
				"view_child_attendance",
				"view_child_grades",
				"communicate_teachers",
				"receive_notifications",
			},
			Description: "Parent with access to their children's information",
		},
		{
			Role: "student",
			Permissions: []string{
				"view_own_attendance",
				"view_own_grades",
				"view_assignments",
				"submit_assignments",
			},
			Description: "Student with access to their own information",
		},
	}

	for _, role := range roles {
		_, err := client.Collection("roles").Doc(role.Role).Set(ctx, role)
		if err != nil {
			log.Printf("Error adding role %s: %v", role.Role, err)
		} else {
			fmt.Printf("✅ Added role: %s\n", role.Role)
		}
	}

	// Add Test Users
	fmt.Println("\n👥 Adding test users...")

	users := []User{
		// Admin User
		{
			UID:       "admin_test_uid_001",
			Email:     "admin@oakwoodschool.edu",
			Name:      "Dr. Jennifer Adams",
			Role:      "admin",
			EntityID:  "admin_001",
			Active:    true,
			CreatedAt: time.Now(),
			LastLogin: time.Now(),
		},

		// Teacher Users (matching our existing teachers)
		{
			UID:       "teacher_test_uid_001",
			Email:     "sarah.johnson@oakwoodschool.edu",
			Name:      "Ms. Sarah Johnson",
			Role:      "teacher",
			EntityID:  "teacher_001",
			Active:    true,
			CreatedAt: time.Now(),
			LastLogin: time.Now(),
		},
		{
			UID:       "teacher_test_uid_002",
			Email:     "david.chen@oakwoodschool.edu",
			Name:      "Mr. David Chen",
			Role:      "teacher",
			EntityID:  "teacher_002",
			Active:    true,
			CreatedAt: time.Now(),
			LastLogin: time.Now(),
		},
		{
			UID:       "teacher_test_uid_003",
			Email:     "maria.rodriguez@oakwoodschool.edu",
			Name:      "Dr. Maria Rodriguez",
			Role:      "teacher",
			EntityID:  "teacher_003",
			Active:    true,
			CreatedAt: time.Now(),
			LastLogin: time.Now(),
		},

		// Parent Users (matching our existing parents)
		{
			UID:       "parent_test_uid_001",
			Email:     "john.smith@email.com",
			Name:      "John Smith",
			Role:      "parent",
			EntityID:  "parent_001",
			Active:    true,
			CreatedAt: time.Now(),
			LastLogin: time.Now(),
		},
		{
			UID:       "parent_test_uid_002",
			Email:     "emily.wilson@email.com",
			Name:      "Emily Wilson",
			Role:      "parent",
			EntityID:  "parent_002",
			Active:    true,
			CreatedAt: time.Now(),
			LastLogin: time.Now(),
		},
		{
			UID:       "parent_test_uid_003",
			Email:     "michael.brown@email.com",
			Name:      "Michael Brown",
			Role:      "parent",
			EntityID:  "parent_003",
			Active:    true,
			CreatedAt: time.Now(),
			LastLogin: time.Now(),
		},

		// Student Users (matching our existing students)
		{
			UID:       "student_test_uid_001",
			Email:     "alex.smith@student.oakwoodschool.edu",
			Name:      "Alex Smith",
			Role:      "student",
			EntityID:  "student_001",
			Active:    true,
			CreatedAt: time.Now(),
			LastLogin: time.Now(),
		},
		{
			UID:       "student_test_uid_002",
			Email:     "emma.smith@student.oakwoodschool.edu",
			Name:      "Emma Smith",
			Role:      "student",
			EntityID:  "student_002",
			Active:    true,
			CreatedAt: time.Now(),
			LastLogin: time.Now(),
		},
		{
			UID:       "student_test_uid_003",
			Email:     "ryan.wilson@student.oakwoodschool.edu",
			Name:      "Ryan Wilson",
			Role:      "student",
			EntityID:  "student_003",
			Active:    true,
			CreatedAt: time.Now(),
			LastLogin: time.Now(),
		},
		{
			UID:       "student_test_uid_004",
			Email:     "sofia.brown@student.oakwoodschool.edu",
			Name:      "Sofia Brown",
			Role:      "student",
			EntityID:  "student_004",
			Active:    true,
			CreatedAt: time.Now(),
			LastLogin: time.Now(),
		},
	}

	for _, user := range users {
		_, err := client.Collection("users").Doc(user.UID).Set(ctx, user)
		if err != nil {
			log.Printf("Error adding user %s: %v", user.Name, err)
		} else {
			fmt.Printf("✅ Added user: %s (%s) - %s\n", user.Name, user.Email, user.Role)
		}
	}

	// Add some test auth tokens for easy testing
	fmt.Println("\n🔑 Creating test authentication tokens...")

	testTokens := map[string]map[string]interface{}{
		"test_admin_token": {
			"uid":        "admin_test_uid_001",
			"email":      "admin@oakwoodschool.edu",
			"role":       "admin",
			"entity_id":  "admin_001",
			"issued_at":  time.Now().Unix(),
			"expires_at": time.Now().Add(24 * time.Hour).Unix(),
		},
		"test_teacher_token": {
			"uid":        "teacher_test_uid_001",
			"email":      "sarah.johnson@oakwoodschool.edu",
			"role":       "teacher",
			"entity_id":  "teacher_001",
			"issued_at":  time.Now().Unix(),
			"expires_at": time.Now().Add(24 * time.Hour).Unix(),
		},
		"test_parent_token": {
			"uid":        "parent_test_uid_002",
			"email":      "emily.wilson@email.com",
			"role":       "parent",
			"entity_id":  "parent_002",
			"issued_at":  time.Now().Unix(),
			"expires_at": time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	for tokenName, tokenData := range testTokens {
		_, err := client.Collection("test_tokens").Doc(tokenName).Set(ctx, tokenData)
		if err != nil {
			log.Printf("Error adding test token %s: %v", tokenName, err)
		} else {
			fmt.Printf("✅ Added test token: %s\n", tokenName)
		}
	}

	fmt.Println("\n🎉 Authentication setup complete!")
	fmt.Println("\n📊 Summary:")
	fmt.Printf("  • %d Roles defined\n", len(roles))
	fmt.Printf("  • %d Users added\n", len(users))
	fmt.Printf("  • %d Test tokens created\n", len(testTokens))

	fmt.Println("\n🧪 Test Tokens Available:")
	fmt.Println("  • test_admin_token - Dr. Jennifer Adams (Admin)")
	fmt.Println("  • test_teacher_token - Ms. Sarah Johnson (Teacher)")
	fmt.Println("  • test_parent_token - Emily Wilson (Parent)")

	fmt.Println("\n🚀 Ready to test authenticated endpoints!")
	fmt.Println("\nExample usage:")
	fmt.Println(`  curl -X POST http://localhost:8080/api/v1/gpt/ask \
    -H "Authorization: Bearer test_teacher_token" \
    -H "Content-Type: application/json" \
    -d '{"query": "Show me attendance for my students"}'`)
}
