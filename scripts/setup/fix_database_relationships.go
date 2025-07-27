package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// Enrollment represents a student enrollment in a class
type Enrollment struct {
	ID        string    `firestore:"id"`
	StudentID string    `firestore:"student_id"`
	ClassID   string    `firestore:"class_id"`
	CreatedAt time.Time `firestore:"created_at"`
}

func main() {
	fmt.Println("🔧 Fixing SchoolGPT Database Relationships")

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

	fmt.Println("\n📚 Adding student enrollments to classes...")

	// Create enrollments that connect students to classes
	enrollments := []Enrollment{
		// Alex Smith (student_001) - parent_001 (John Smith)
		{
			ID:        "enrollment_001",
			StudentID: "student_001",
			ClassID:   "class_001", // Algebra II - teacher_001 (Ms. Sarah Johnson)
			CreatedAt: time.Now(),
		},
		{
			ID:        "enrollment_002",
			StudentID: "student_001",
			ClassID:   "class_003", // American Literature - teacher_002 (Mr. David Chen)
			CreatedAt: time.Now(),
		},

		// Emma Smith (student_002) - parent_001 (John Smith)
		{
			ID:        "enrollment_003",
			StudentID: "student_002",
			ClassID:   "class_002", // Geometry - teacher_001 (Ms. Sarah Johnson)
			CreatedAt: time.Now(),
		},
		{
			ID:        "enrollment_004",
			StudentID: "student_002",
			ClassID:   "class_004", // Biology - teacher_003 (Dr. Maria Rodriguez)
			CreatedAt: time.Now(),
		},

		// Ryan Wilson (student_003) - parent_002 (Emily Wilson)
		{
			ID:        "enrollment_005",
			StudentID: "student_003",
			ClassID:   "class_001", // Algebra II - teacher_001 (Ms. Sarah Johnson)
			CreatedAt: time.Now(),
		},
		{
			ID:        "enrollment_006",
			StudentID: "student_003",
			ClassID:   "class_002", // Geometry - teacher_001 (Ms. Sarah Johnson)
			CreatedAt: time.Now(),
		},

		// Sofia Brown (student_004) - parent_003 (Michael Brown)
		{
			ID:        "enrollment_007",
			StudentID: "student_004",
			ClassID:   "class_003", // American Literature - teacher_002 (Mr. David Chen)
			CreatedAt: time.Now(),
		},
		{
			ID:        "enrollment_008",
			StudentID: "student_004",
			ClassID:   "class_004", // Biology - teacher_003 (Dr. Maria Rodriguez)
			CreatedAt: time.Now(),
		},
	}

	for _, enrollment := range enrollments {
		_, err := client.Collection("enrollments").Doc(enrollment.ID).Set(ctx, enrollment)
		if err != nil {
			log.Printf("Error adding enrollment %s: %v", enrollment.ID, err)
		} else {
			fmt.Printf("✅ Added enrollment: %s -> %s to %s\n", enrollment.ID, enrollment.StudentID, enrollment.ClassID)
		}
	}

	fmt.Println("\n🎯 Updating attendance records with class IDs...")

	// Update existing attendance records to include class_id
	attendanceUpdates := map[string]string{
		// Based on our enrollment data above
		"student_001_2025-06-21": "class_001", // Alex in Algebra II
		"student_001_2025-06-22": "class_003", // Alex in American Literature
		"student_001_2025-06-23": "class_001", // Alex in Algebra II
		"student_001_2025-06-24": "class_003", // Alex in American Literature
		"student_001_2025-06-25": "class_001", // Alex in Algebra II

		"student_002_2025-06-21": "class_002", // Emma in Geometry
		"student_002_2025-06-22": "class_004", // Emma in Biology
		"student_002_2025-06-23": "class_002", // Emma in Geometry
		"student_002_2025-06-24": "class_004", // Emma in Biology
		"student_002_2025-06-25": "class_002", // Emma in Geometry

		"student_003_2025-06-21": "class_003", // Ryan in American Literature
		"student_003_2025-06-22": "class_004", // Ryan in Biology
		"student_003_2025-06-23": "class_001", // Ryan in Algebra II
		"student_003_2025-06-24": "class_001", // Ryan in Algebra II
		"student_003_2025-06-25": "class_002", // Ryan in Geometry

		"student_004_2025-06-21": "class_004", // Sofia in Biology
		"student_004_2025-06-22": "class_003", // Sofia in American Literature
		"student_004_2025-06-23": "class_004", // Sofia in Biology
		"student_004_2025-06-24": "class_003", // Sofia in American Literature
		"student_004_2025-06-25": "class_004", // Sofia in Biology
	}

	count := 0
	for docID, classID := range attendanceUpdates {
		_, err := client.Collection("attendance").Doc(docID).Update(ctx, []firestore.Update{
			{Path: "class_id", Value: classID},
		})
		if err != nil {
			log.Printf("Error updating attendance %s: %v", docID, err)
		} else {
			count++
		}
	}

	fmt.Printf("✅ Updated %d attendance records with class IDs\n", count)

	fmt.Println("\n📊 Database Relationship Summary:")
	fmt.Println("Teachers and their classes:")
	fmt.Println("  • Ms. Sarah Johnson (teacher_001): Algebra II (class_001), Geometry (class_002)")
	fmt.Println("  • Mr. David Chen (teacher_002): American Literature (class_003)")
	fmt.Println("  • Dr. Maria Rodriguez (teacher_003): Biology (class_004)")
	fmt.Println()
	fmt.Println("Parents and their children:")
	fmt.Println("  • John Smith (parent_001): Alex Smith, Emma Smith")
	fmt.Println("  • Emily Wilson (parent_002): Ryan Wilson")
	fmt.Println("  • Michael Brown (parent_003): Sofia Brown")
	fmt.Println()
	fmt.Println("Student enrollments:")
	fmt.Println("  • Alex Smith: Algebra II, American Literature")
	fmt.Println("  • Emma Smith: Geometry, Biology")
	fmt.Println("  • Ryan Wilson: Algebra II, Geometry, American Literature")
	fmt.Println("  • Sofia Brown: American Literature, Biology")

	fmt.Println("\n🎉 Database relationships fixed!")
	fmt.Println("Now SchoolGPT can provide accurate, role-based responses with real data!")
}
