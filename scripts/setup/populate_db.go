package main

import (
	"context"
	"fmt"
	"log"
	"time"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// Student represents a student in the school
type Student struct {
	ID        string    `firestore:"id"`
	Name      string    `firestore:"name"`
	Email     string    `firestore:"email"`
	Grade     string    `firestore:"grade"`
	ClassID   string    `firestore:"class_id"`
	ParentID  string    `firestore:"parent_id"`
	CreatedAt time.Time `firestore:"created_at"`
}

// Teacher represents a teacher in the school
type Teacher struct {
	ID        string    `firestore:"id"`
	Name      string    `firestore:"name"`
	Email     string    `firestore:"email"`
	Subject   string    `firestore:"subject"`
	ClassIDs  []string  `firestore:"class_ids"`
	CreatedAt time.Time `firestore:"created_at"`
}

// Class represents a class in the school
type Class struct {
	ID         string    `firestore:"id"`
	Name       string    `firestore:"name"`
	Subject    string    `firestore:"subject"`
	Grade      string    `firestore:"grade"`
	TeacherID  string    `firestore:"teacher_id"`
	StudentIDs []string  `firestore:"student_ids"`
	Schedule   string    `firestore:"schedule"`
	CreatedAt  time.Time `firestore:"created_at"`
}

// Parent represents a parent/guardian
type Parent struct {
	ID         string    `firestore:"id"`
	Name       string    `firestore:"name"`
	Email      string    `firestore:"email"`
	Phone      string    `firestore:"phone"`
	StudentIDs []string  `firestore:"student_ids"`
	CreatedAt  time.Time `firestore:"created_at"`
}

// AttendanceRecord represents an attendance record
type AttendanceRecord struct {
	ID        string    `firestore:"id"`
	StudentID string    `firestore:"student_id"`
	ClassID   string    `firestore:"class_id"`
	Date      string    `firestore:"date"`
	Status    string    `firestore:"status"` // present, absent, late, excused
	MarkedBy  string    `firestore:"marked_by"`
	Notes     string    `firestore:"notes"`
	CreatedAt time.Time `firestore:"created_at"`
}

func main() {
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

	fmt.Println("🏫 Populating SchoolGPT database with sample data...")

	// Add Teachers
	teachers := []Teacher{
		{
			ID:        "teacher_001",
			Name:      "Ms. Sarah Johnson",
			Email:     "sarah.johnson@oakwoodschool.edu",
			Subject:   "Mathematics",
			ClassIDs:  []string{"class_001", "class_002"},
			CreatedAt: time.Now(),
		},
		{
			ID:        "teacher_002",
			Name:      "Mr. David Chen",
			Email:     "david.chen@oakwoodschool.edu",
			Subject:   "English Literature",
			ClassIDs:  []string{"class_003"},
			CreatedAt: time.Now(),
		},
		{
			ID:        "teacher_003",
			Name:      "Dr. Maria Rodriguez",
			Email:     "maria.rodriguez@oakwoodschool.edu",
			Subject:   "Science",
			ClassIDs:  []string{"class_004"},
			CreatedAt: time.Now(),
		},
	}

	for _, teacher := range teachers {
		_, err := client.Collection("teachers").Doc(teacher.ID).Set(ctx, teacher)
		if err != nil {
			log.Printf("Error adding teacher %s: %v", teacher.Name, err)
		} else {
			fmt.Printf("✅ Added teacher: %s\n", teacher.Name)
		}
	}

	// Add Parents
	parents := []Parent{
		{
			ID:         "parent_001",
			Name:       "John Smith",
			Email:      "john.smith@email.com",
			Phone:      "+1-555-0101",
			StudentIDs: []string{"student_001", "student_002"},
			CreatedAt:  time.Now(),
		},
		{
			ID:         "parent_002",
			Name:       "Emily Wilson",
			Email:      "emily.wilson@email.com",
			Phone:      "+1-555-0102",
			StudentIDs: []string{"student_003"},
			CreatedAt:  time.Now(),
		},
		{
			ID:         "parent_003",
			Name:       "Michael Brown",
			Email:      "michael.brown@email.com",
			Phone:      "+1-555-0103",
			StudentIDs: []string{"student_004"},
			CreatedAt:  time.Now(),
		},
	}

	for _, parent := range parents {
		_, err := client.Collection("parents").Doc(parent.ID).Set(ctx, parent)
		if err != nil {
			log.Printf("Error adding parent %s: %v", parent.Name, err)
		} else {
			fmt.Printf("✅ Added parent: %s\n", parent.Name)
		}
	}

	// Add Classes
	classes := []Class{
		{
			ID:         "class_001",
			Name:       "Algebra II",
			Subject:    "Mathematics",
			Grade:      "10th",
			TeacherID:  "teacher_001",
			StudentIDs: []string{"student_001", "student_003"},
			Schedule:   "Mon/Wed/Fri 9:00-10:00 AM",
			CreatedAt:  time.Now(),
		},
		{
			ID:         "class_002",
			Name:       "Geometry",
			Subject:    "Mathematics",
			Grade:      "9th",
			TeacherID:  "teacher_001",
			StudentIDs: []string{"student_002", "student_004"},
			Schedule:   "Tue/Thu 10:00-11:00 AM",
			CreatedAt:  time.Now(),
		},
		{
			ID:         "class_003",
			Name:       "American Literature",
			Subject:    "English",
			Grade:      "11th",
			TeacherID:  "teacher_002",
			StudentIDs: []string{"student_001", "student_003"},
			Schedule:   "Mon/Wed/Fri 11:00 AM-12:00 PM",
			CreatedAt:  time.Now(),
		},
		{
			ID:         "class_004",
			Name:       "Biology",
			Subject:    "Science",
			Grade:      "10th",
			TeacherID:  "teacher_003",
			StudentIDs: []string{"student_002", "student_004"},
			Schedule:   "Tue/Thu 1:00-2:30 PM",
			CreatedAt:  time.Now(),
		},
	}

	for _, class := range classes {
		_, err := client.Collection("classes").Doc(class.ID).Set(ctx, class)
		if err != nil {
			log.Printf("Error adding class %s: %v", class.Name, err)
		} else {
			fmt.Printf("✅ Added class: %s\n", class.Name)
		}
	}

	// Add Students
	students := []Student{
		{
			ID:        "student_001",
			Name:      "Alex Smith",
			Email:     "alex.smith@student.oakwoodschool.edu",
			Grade:     "10th",
			ClassID:   "class_001",
			ParentID:  "parent_001",
			CreatedAt: time.Now(),
		},
		{
			ID:        "student_002",
			Name:      "Emma Smith",
			Email:     "emma.smith@student.oakwoodschool.edu",
			Grade:     "9th",
			ClassID:   "class_002",
			ParentID:  "parent_001",
			CreatedAt: time.Now(),
		},
		{
			ID:        "student_003",
			Name:      "Ryan Wilson",
			Email:     "ryan.wilson@student.oakwoodschool.edu",
			Grade:     "10th",
			ClassID:   "class_001",
			ParentID:  "parent_002",
			CreatedAt: time.Now(),
		},
		{
			ID:        "student_004",
			Name:      "Sofia Brown",
			Email:     "sofia.brown@student.oakwoodschool.edu",
			Grade:     "9th",
			ClassID:   "class_002",
			ParentID:  "parent_003",
			CreatedAt: time.Now(),
		},
	}

	for _, student := range students {
		_, err := client.Collection("students").Doc(student.ID).Set(ctx, student)
		if err != nil {
			log.Printf("Error adding student %s: %v", student.Name, err)
		} else {
			fmt.Printf("✅ Added student: %s\n", student.Name)
		}
	}

	// Add Attendance Records (last 5 days)
	attendanceRecords := []AttendanceRecord{}
	studentIDs := []string{"student_001", "student_002", "student_003", "student_004"}
	classIDs := []string{"class_001", "class_002", "class_003", "class_004"}
	statuses := []string{"present", "present", "present", "absent", "late"}

	for i := 0; i < 5; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		for j, studentID := range studentIDs {
			for k, classID := range classIDs {
				recordID := fmt.Sprintf("attendance_%s_%s_%s", studentID, classID, date)
				status := statuses[(i+j+k)%len(statuses)]

				record := AttendanceRecord{
					ID:        recordID,
					StudentID: studentID,
					ClassID:   classID,
					Date:      date,
					Status:    status,
					MarkedBy:  fmt.Sprintf("teacher_%03d", (k%3)+1),
					Notes:     "",
					CreatedAt: time.Now(),
				}

				if status == "absent" {
					record.Notes = "Family vacation"
				} else if status == "late" {
					record.Notes = "Traffic delay"
				}

				attendanceRecords = append(attendanceRecords, record)
			}
		}
	}

	for _, record := range attendanceRecords {
		_, err := client.Collection("attendance").Doc(record.ID).Set(ctx, record)
		if err != nil {
			log.Printf("Error adding attendance record %s: %v", record.ID, err)
		}
	}

	fmt.Printf("✅ Added %d attendance records\n", len(attendanceRecords))

	fmt.Println("\n🎉 Database population complete!")
	fmt.Println("\n📊 Summary:")
	fmt.Printf("  • %d Teachers\n", len(teachers))
	fmt.Printf("  • %d Parents\n", len(parents))
	fmt.Printf("  • %d Classes\n", len(classes))
	fmt.Printf("  • %d Students\n", len(students))
	fmt.Printf("  • %d Attendance Records\n", len(attendanceRecords))
	fmt.Println("\n🚀 Ready to test SchoolGPT with real data!")
}
