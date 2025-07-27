package storage

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/schoolgpt/backend/internal/config"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FirestoreDB handles Firestore database operations
type FirestoreDB struct {
	client *firestore.Client
	config *config.Config
}

// AttendanceRecord represents a student's attendance record
type AttendanceRecord struct {
	StudentID string    `firestore:"student_id" json:"student_id"`
	Date      string    `firestore:"date" json:"date"`
	Status    string    `firestore:"status" json:"status"`
	MarkedBy  string    `firestore:"marked_by" json:"marked_by"`
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
	Reason    string    `firestore:"reason,omitempty" json:"reason,omitempty"`
	ClassID   string    `firestore:"class_id" json:"class_id"`
}

// Student represents a student
type Student struct {
	ID        string    `firestore:"id" json:"id"`
	Name      string    `firestore:"name" json:"name"`
	Email     string    `firestore:"email" json:"email"`
	ParentID  string    `firestore:"parent_id" json:"parent_id"`
	Grade     string    `firestore:"grade" json:"grade"`
	StudentID string    `firestore:"student_id" json:"student_id"`
	Phone     string    `firestore:"phone" json:"phone"`
	Address   string    `firestore:"address" json:"address"`
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}

// Teacher represents a teacher
type Teacher struct {
	ID         string    `firestore:"id" json:"id"`
	Name       string    `firestore:"name" json:"name"`
	Email      string    `firestore:"email" json:"email"`
	Subject    string    `firestore:"subject" json:"subject"`
	Department string    `firestore:"department" json:"department"`
	EmployeeID string    `firestore:"employee_id" json:"employee_id"`
	Phone      string    `firestore:"phone" json:"phone"`
	Address    string    `firestore:"address" json:"address"`
	CreatedAt  time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt  time.Time `firestore:"updated_at" json:"updated_at"`
}

// Parent represents a parent
type Parent struct {
	ID        string    `firestore:"id" json:"id"`
	Name      string    `firestore:"name" json:"name"`
	Email     string    `firestore:"email" json:"email"`
	Phone     string    `firestore:"phone" json:"phone"`
	Address   string    `firestore:"address" json:"address"`
	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}

// Class represents a class
type Class struct {
	ID        string `firestore:"id" json:"id"`
	Name      string `firestore:"name" json:"name"`
	TeacherID string `firestore:"teacher_id" json:"teacher_id"`
	Subject   string `firestore:"subject" json:"subject"`
}

// AttendanceSummary represents attendance summary for a student
type AttendanceSummary struct {
	StudentID      string             `json:"student_id"`
	StudentName    string             `json:"student_name"`
	TotalDays      int                `json:"total_days"`
	PresentDays    int                `json:"present_days"`
	AbsentDays     int                `json:"absent_days"`
	LateDays       int                `json:"late_days"`
	AttendanceRate float64            `json:"attendance_rate"`
	RecentAbsences []AttendanceRecord `json:"recent_absences,omitempty"`
}

// New creates a new FirestoreDB instance
func New(cfg *config.Config) (*FirestoreDB, error) {
	ctx := context.Background()

	// If credentials are provided, use them
	var client *firestore.Client
	var err error

	if cfg.GoogleCredentialsPath != "" {
		// Use credentials file
		opt := option.WithCredentialsFile(cfg.GoogleCredentialsPath)
		client, err = firestore.NewClient(ctx, cfg.FirebaseProjectID, opt)
	} else if cfg.FirebasePrivateKey != "" && cfg.FirebaseClientEmail != "" {
		// Use service account credentials from environment variables
		// In a real application, you would use option.WithCredentialsJSON with a proper JSON credential
		// For simplicity, we'll use the default credentials
		client, err = firestore.NewClient(ctx, cfg.FirebaseProjectID)
	} else if cfg.Env == "development" {
		// In development, try to use application default credentials
		client, err = firestore.NewClient(ctx, cfg.FirebaseProjectID)
	} else {
		return nil, fmt.Errorf("firestore credentials not provided")
	}

	if err != nil {
		return nil, fmt.Errorf("error initializing firestore client: %v", err)
	}

	return &FirestoreDB{
		client: client,
		config: cfg,
	}, nil
}

// Close closes the Firestore client
func (db *FirestoreDB) Close() error {
	return db.client.Close()
}

// GetClient returns the Firestore client
func (db *FirestoreDB) GetClient() *firestore.Client {
	return db.client
}

// FetchAttendance fetches a student's attendance record for a specific date
func (db *FirestoreDB) FetchAttendance(ctx context.Context, studentID, date string) (*AttendanceRecord, error) {
	// Set a timeout for the operation
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Query the attendance collection
	query := db.client.Collection("attendance").
		Where("student_id", "==", studentID).
		Where("date", "==", date).
		Limit(1)

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error fetching attendance: %v", err)
	}

	if len(docs) == 0 {
		return nil, status.Errorf(codes.NotFound, "attendance record not found")
	}

	// Parse the document into an AttendanceRecord
	var record AttendanceRecord
	if err := docs[0].DataTo(&record); err != nil {
		return nil, fmt.Errorf("error parsing attendance record: %v", err)
	}

	return &record, nil
}

// SaveAttendance saves a student's attendance record
func (db *FirestoreDB) SaveAttendance(ctx context.Context, record *AttendanceRecord) error {
	// Set a timeout for the operation
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Set timestamps
	now := time.Now()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	record.UpdatedAt = now

	// Save to Firestore
	_, err := db.client.Collection("attendance").
		Doc(fmt.Sprintf("%s_%s", record.StudentID, record.Date)).
		Set(ctx, record)

	if err != nil {
		return fmt.Errorf("error saving attendance record: %v", err)
	}

	return nil
}

// GetStudentsByParent gets all students for a specific parent
func (db *FirestoreDB) GetStudentsByParent(ctx context.Context, parentID string) ([]Student, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	query := db.client.Collection("students").Where("parent_id", "==", parentID)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error fetching students: %v", err)
	}

	var students []Student
	for _, doc := range docs {
		var student Student
		if err := doc.DataTo(&student); err != nil {
			continue
		}
		students = append(students, student)
	}

	return students, nil
}

// GetStudentsByTeacher gets all students for classes taught by a teacher
func (db *FirestoreDB) GetStudentsByTeacher(ctx context.Context, teacherID string) ([]Student, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// First get all classes for this teacher
	classQuery := db.client.Collection("classes").Where("teacher_id", "==", teacherID)
	classDocs, err := classQuery.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error fetching teacher classes: %v", err)
	}

	var classIDs []string
	for _, doc := range classDocs {
		var class Class
		if err := doc.DataTo(&class); err == nil {
			classIDs = append(classIDs, class.ID)
		}
	}

	if len(classIDs) == 0 {
		return []Student{}, nil
	}

	// Get all students enrolled in these classes
	var students []Student
	for _, classID := range classIDs {
		enrollmentQuery := db.client.Collection("enrollments").Where("class_id", "==", classID)
		enrollmentDocs, err := enrollmentQuery.Documents(ctx).GetAll()
		if err != nil {
			continue
		}

		for _, enrollDoc := range enrollmentDocs {
			data := enrollDoc.Data()
			if studentID, ok := data["student_id"].(string); ok {
				student, err := db.GetStudent(ctx, studentID)
				if err == nil {
					students = append(students, *student)
				}
			}
		}
	}

	return students, nil
}

// GetStudent gets a specific student by ID
func (db *FirestoreDB) GetStudent(ctx context.Context, studentID string) (*Student, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	doc, err := db.client.Collection("students").Doc(studentID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching student: %v", err)
	}

	var student Student
	if err := doc.DataTo(&student); err != nil {
		return nil, fmt.Errorf("error parsing student: %v", err)
	}

	return &student, nil
}

// GetAttendanceSummary gets attendance summary for a student
func (db *FirestoreDB) GetAttendanceSummary(ctx context.Context, studentID string, days int) (*AttendanceSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get student info
	student, err := db.GetStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}

	// Get attendance records
	query := db.client.Collection("attendance").
		Where("student_id", "==", studentID).
		OrderBy("date", firestore.Desc)

	if days > 0 {
		query = query.Limit(days)
	}

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error fetching attendance records: %v", err)
	}

	var records []AttendanceRecord
	presentCount := 0
	absentCount := 0
	lateCount := 0

	for _, doc := range docs {
		var record AttendanceRecord
		if err := doc.DataTo(&record); err != nil {
			continue
		}
		records = append(records, record)

		switch record.Status {
		case "present":
			presentCount++
		case "absent":
			absentCount++
		case "late":
			lateCount++
		}
	}

	totalDays := len(records)
	attendanceRate := 0.0
	if totalDays > 0 {
		attendanceRate = float64(presentCount) / float64(totalDays) * 100
	}

	// Get recent absences (last 5 absences)
	var recentAbsences []AttendanceRecord
	count := 0
	for _, record := range records {
		if record.Status == "absent" && count < 5 {
			recentAbsences = append(recentAbsences, record)
			count++
		}
	}

	return &AttendanceSummary{
		StudentID:      studentID,
		StudentName:    student.Name,
		TotalDays:      totalDays,
		PresentDays:    presentCount,
		AbsentDays:     absentCount,
		LateDays:       lateCount,
		AttendanceRate: attendanceRate,
		RecentAbsences: recentAbsences,
	}, nil
}

// GetAttendanceByDateRange gets attendance records for a date range
func (db *FirestoreDB) GetAttendanceByDateRange(ctx context.Context, startDate, endDate string, teacherID string) ([]AttendanceRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	query := db.client.Collection("attendance").
		Where("date", ">=", startDate).
		Where("date", "<=", endDate)

	if teacherID != "" {
		query = query.Where("marked_by", "==", teacherID)
	}

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error fetching attendance records: %v", err)
	}

	var records []AttendanceRecord
	for _, doc := range docs {
		var record AttendanceRecord
		if err := doc.DataTo(&record); err != nil {
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

// GetTeacher gets a specific teacher by ID
func (db *FirestoreDB) GetTeacher(ctx context.Context, teacherID string) (*Teacher, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	doc, err := db.client.Collection("teachers").Doc(teacherID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching teacher: %v", err)
	}

	var teacher Teacher
	if err := doc.DataTo(&teacher); err != nil {
		return nil, fmt.Errorf("error parsing teacher: %v", err)
	}

	return &teacher, nil
}

// GetAllStudents gets all students (admin only)
func (db *FirestoreDB) GetAllStudents(ctx context.Context) ([]Student, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	docs, err := db.client.Collection("students").Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error fetching all students: %v", err)
	}

	var students []Student
	for _, doc := range docs {
		var student Student
		if err := doc.DataTo(&student); err != nil {
			continue
		}
		students = append(students, student)
	}

	return students, nil
}

// GetUserByEmail retrieves a user by email from any collection
func (db *FirestoreDB) GetUserByEmail(ctx context.Context, email string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Check students collection
	query := db.client.Collection("students").Where("email", "==", email).Limit(1)
	docs, err := query.Documents(ctx).GetAll()
	if err == nil && len(docs) > 0 {
		var student Student
		if err := docs[0].DataTo(&student); err == nil {
			return &student, nil
		}
	}

	// Check teachers collection
	query = db.client.Collection("teachers").Where("email", "==", email).Limit(1)
	docs, err = query.Documents(ctx).GetAll()
	if err == nil && len(docs) > 0 {
		var teacher Teacher
		if err := docs[0].DataTo(&teacher); err == nil {
			return &teacher, nil
		}
	}

	// Check parents collection
	query = db.client.Collection("parents").Where("email", "==", email).Limit(1)
	docs, err = query.Documents(ctx).GetAll()
	if err == nil && len(docs) > 0 {
		var parent Parent
		if err := docs[0].DataTo(&parent); err == nil {
			return &parent, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "user not found")
}

// CreateStudent creates a new student
func (db *FirestoreDB) CreateStudent(ctx context.Context, student *Student) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Set timestamps and generate ID if not provided
	now := time.Now()
	if student.ID == "" {
		student.ID = fmt.Sprintf("student_%d", now.Unix())
	}
	student.CreatedAt = now
	student.UpdatedAt = now

	_, err := db.client.Collection("students").Doc(student.ID).Set(ctx, student)
	return err
}

// CreateTeacher creates a new teacher
func (db *FirestoreDB) CreateTeacher(ctx context.Context, teacher *Teacher) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Set timestamps and generate ID if not provided
	now := time.Now()
	if teacher.ID == "" {
		teacher.ID = fmt.Sprintf("teacher_%d", now.Unix())
	}
	teacher.CreatedAt = now
	teacher.UpdatedAt = now

	_, err := db.client.Collection("teachers").Doc(teacher.ID).Set(ctx, teacher)
	return err
}

// CreateParent creates a new parent
func (db *FirestoreDB) CreateParent(ctx context.Context, parent *Parent) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Set timestamps and generate ID if not provided
	now := time.Now()
	if parent.ID == "" {
		parent.ID = fmt.Sprintf("parent_%d", now.Unix())
	}
	parent.CreatedAt = now
	parent.UpdatedAt = now

	_, err := db.client.Collection("parents").Doc(parent.ID).Set(ctx, parent)
	return err
}

// GetUsersWithFilters retrieves users with optional filtering
func (db *FirestoreDB) GetUsersWithFilters(ctx context.Context, role, grade, department string) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var users []interface{}

	// If role filter is specified, only query that collection
	if role != "" {
		switch role {
		case "student":
			students, err := db.getStudentsWithFilters(ctx, grade)
			if err != nil {
				return nil, err
			}
			for _, student := range students {
				users = append(users, student)
			}
		case "teacher":
			teachers, err := db.getTeachersWithFilters(ctx, department)
			if err != nil {
				return nil, err
			}
			for _, teacher := range teachers {
				users = append(users, teacher)
			}
		case "parent":
			parents, err := db.getParentsWithFilters(ctx)
			if err != nil {
				return nil, err
			}
			for _, parent := range parents {
				users = append(users, parent)
			}
		}
	} else {
		// Query all collections if no role filter
		students, _ := db.getStudentsWithFilters(ctx, grade)
		for _, student := range students {
			users = append(users, student)
		}

		teachers, _ := db.getTeachersWithFilters(ctx, department)
		for _, teacher := range teachers {
			users = append(users, teacher)
		}

		parents, _ := db.getParentsWithFilters(ctx)
		for _, parent := range parents {
			users = append(users, parent)
		}
	}

	return users, nil
}

// Helper functions for filtered queries
func (db *FirestoreDB) getStudentsWithFilters(ctx context.Context, grade string) ([]Student, error) {
	collection := db.client.Collection("students")
	var query firestore.Query

	if grade != "" {
		query = collection.Where("grade", "==", grade)
	} else {
		query = collection.Query
	}

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var students []Student
	for _, doc := range docs {
		var student Student
		if err := doc.DataTo(&student); err == nil {
			students = append(students, student)
		}
	}

	return students, nil
}

func (db *FirestoreDB) getTeachersWithFilters(ctx context.Context, department string) ([]Teacher, error) {
	collection := db.client.Collection("teachers")
	var query firestore.Query

	if department != "" {
		query = collection.Where("department", "==", department)
	} else {
		query = collection.Query
	}

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var teachers []Teacher
	for _, doc := range docs {
		var teacher Teacher
		if err := doc.DataTo(&teacher); err == nil {
			teachers = append(teachers, teacher)
		}
	}

	return teachers, nil
}

func (db *FirestoreDB) getParentsWithFilters(ctx context.Context) ([]Parent, error) {
	docs, err := db.client.Collection("parents").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var parents []Parent
	for _, doc := range docs {
		var parent Parent
		if err := doc.DataTo(&parent); err == nil {
			parents = append(parents, parent)
		}
	}

	return parents, nil
}
