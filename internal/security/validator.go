package security

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

// SecurityValidator provides comprehensive input validation and sanitization
type SecurityValidator struct {
	maxStringLength   int
	allowedCharacters map[string]*regexp.Regexp
	sqlInjectionRules []*regexp.Regexp
	xssPatterns       []*regexp.Regexp
	validator         *validator.Validate
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", ve.Field, ve.Message)
}

// NewSecurityValidator creates a new security validator with predefined rules
func NewSecurityValidator() *SecurityValidator {
	sv := &SecurityValidator{
		maxStringLength: 500,
		validator:       validator.New(),
		allowedCharacters: map[string]*regexp.Regexp{
			"school_name":  regexp.MustCompile(`^[a-zA-Z0-9\s\-.'&()]+$`),
			"school_code":  regexp.MustCompile(`^[A-Z0-9]{6,12}$`),
			"email":        regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
			"name":         regexp.MustCompile(`^[a-zA-Z\s\-.']+$`),
			"phone":        regexp.MustCompile(`^[\+]?[1-9][\d]{0,15}$`),
			"alphanumeric": regexp.MustCompile(`^[a-zA-Z0-9]+$`),
			"numeric":      regexp.MustCompile(`^[0-9]+$`),
			"address":      regexp.MustCompile(`^[a-zA-Z0-9\s\-.,#/()]+$`),
		},
	}

	// SQL injection prevention patterns
	sv.sqlInjectionRules = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`),
		regexp.MustCompile(`(?i)(script|javascript|vbscript|onload|onerror|onclick)`),
		regexp.MustCompile(`(?i)(\||&|\<|\>|;|'|"|--|\*|%|@)`),
		regexp.MustCompile(`(?i)(0x[0-9a-f]+)`), // Hex encoding
		regexp.MustCompile(`(?i)(char\(|ascii\(|substring\()`),
	}

	// XSS prevention patterns
	sv.xssPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`),
		regexp.MustCompile(`(?i)<object[^>]*>.*?</object>`),
		regexp.MustCompile(`(?i)<embed[^>]*>.*?</embed>`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)vbscript:`),
		regexp.MustCompile(`(?i)data:text/html`),
		regexp.MustCompile(`(?i)on\w+\s*=`), // Event handlers
	}

	return sv
}

// ValidateSchoolName validates and sanitizes school name input
func (sv *SecurityValidator) ValidateSchoolName(name string) error {
	if err := sv.validateBasicString(name, "school_name", 2, 100); err != nil {
		return err
	}

	// Character validation
	if !sv.allowedCharacters["school_name"].MatchString(name) {
		return &ValidationError{
			Field:   "school_name",
			Value:   name,
			Message: "school name contains invalid characters",
			Code:    "INVALID_CHARACTERS",
		}
	}

	return sv.validateSecurityThreats(name, "school_name")
}

// ValidateSchoolCode validates school code format
func (sv *SecurityValidator) ValidateSchoolCode(code string) error {
	if len(code) < 6 || len(code) > 12 {
		return &ValidationError{
			Field:   "school_code",
			Value:   code,
			Message: "school code must be 6-12 characters long",
			Code:    "INVALID_LENGTH",
		}
	}

	if !sv.allowedCharacters["school_code"].MatchString(code) {
		return &ValidationError{
			Field:   "school_code",
			Value:   code,
			Message: "school code must contain only uppercase letters and numbers",
			Code:    "INVALID_FORMAT",
		}
	}

	return nil
}

// ValidateEmail validates email format and security
func (sv *SecurityValidator) ValidateEmail(email string) error {
	if err := sv.validateBasicString(email, "email", 5, 255); err != nil {
		return err
	}

	if !sv.allowedCharacters["email"].MatchString(email) {
		return &ValidationError{
			Field:   "email",
			Value:   email,
			Message: "invalid email format",
			Code:    "INVALID_EMAIL_FORMAT",
		}
	}

	return sv.validateSecurityThreats(email, "email")
}

// ValidatePersonName validates person name (teachers, students, etc.)
func (sv *SecurityValidator) ValidatePersonName(name string) error {
	if err := sv.validateBasicString(name, "name", 2, 100); err != nil {
		return err
	}

	if !sv.allowedCharacters["name"].MatchString(name) {
		return &ValidationError{
			Field:   "name",
			Value:   name,
			Message: "name contains invalid characters",
			Code:    "INVALID_CHARACTERS",
		}
	}

	return sv.validateSecurityThreats(name, "name")
}

// ValidatePhoneNumber validates phone number format
func (sv *SecurityValidator) ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return nil // Optional field
	}

	if len(phone) < 10 || len(phone) > 20 {
		return &ValidationError{
			Field:   "phone",
			Value:   phone,
			Message: "phone number must be 10-20 digits",
			Code:    "INVALID_LENGTH",
		}
	}

	if !sv.allowedCharacters["phone"].MatchString(phone) {
		return &ValidationError{
			Field:   "phone",
			Value:   phone,
			Message: "invalid phone number format",
			Code:    "INVALID_FORMAT",
		}
	}

	return nil
}

// ValidateAddress validates address format
func (sv *SecurityValidator) ValidateAddress(address string) error {
	if address == "" {
		return nil // Optional field
	}

	if err := sv.validateBasicString(address, "address", 5, 200); err != nil {
		return err
	}

	if !sv.allowedCharacters["address"].MatchString(address) {
		return &ValidationError{
			Field:   "address",
			Value:   address,
			Message: "address contains invalid characters",
			Code:    "INVALID_CHARACTERS",
		}
	}

	return sv.validateSecurityThreats(address, "address")
}

// ValidateStudentGrade validates grade level
func (sv *SecurityValidator) ValidateStudentGrade(grade string) error {
	if err := sv.validateBasicString(grade, "grade", 1, 20); err != nil {
		return err
	}

	// Allow common grade formats: K, 1, 2, 3...12, Pre-K, etc.
	gradePattern := regexp.MustCompile(`^(Pre-K|K|[1-9]|1[0-2]|Nursery|LKG|UKG|Class [1-9]|Class 1[0-2])$`)
	if !gradePattern.MatchString(grade) {
		return &ValidationError{
			Field:   "grade",
			Value:   grade,
			Message: "invalid grade format",
			Code:    "INVALID_GRADE",
		}
	}

	return nil
}

// ValidateCustomFieldValue validates custom field values based on type
func (sv *SecurityValidator) ValidateCustomFieldValue(value interface{}, fieldType string) error {
	switch fieldType {
	case "text", "textarea":
		if str, ok := value.(string); ok {
			return sv.validateBasicString(str, "custom_field", 0, 500)
		}
	case "number":
		if _, ok := value.(float64); !ok {
			return &ValidationError{
				Field:   "custom_field",
				Message: "value must be a number",
				Code:    "INVALID_TYPE",
			}
		}
	case "email":
		if str, ok := value.(string); ok {
			return sv.ValidateEmail(str)
		}
	case "phone":
		if str, ok := value.(string); ok {
			return sv.ValidatePhoneNumber(str)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return &ValidationError{
				Field:   "custom_field",
				Message: "value must be true or false",
				Code:    "INVALID_TYPE",
			}
		}
	}

	return nil
}

// validateBasicString performs basic string validation
func (sv *SecurityValidator) validateBasicString(str, fieldName string, minLen, maxLen int) error {
	// Null/empty check
	if strings.TrimSpace(str) == "" && minLen > 0 {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s is required", fieldName),
			Code:    "REQUIRED_FIELD",
		}
	}

	// Length validation
	if len(str) < minLen {
		return &ValidationError{
			Field:   fieldName,
			Value:   str,
			Message: fmt.Sprintf("%s must be at least %d characters", fieldName, minLen),
			Code:    "MIN_LENGTH",
		}
	}

	if len(str) > maxLen {
		return &ValidationError{
			Field:   fieldName,
			Value:   str,
			Message: fmt.Sprintf("%s must be at most %d characters", fieldName, maxLen),
			Code:    "MAX_LENGTH",
		}
	}

	// UTF-8 validation
	if !utf8.ValidString(str) {
		return &ValidationError{
			Field:   fieldName,
			Value:   str,
			Message: "contains invalid UTF-8 characters",
			Code:    "INVALID_ENCODING",
		}
	}

	return nil
}

// validateSecurityThreats checks for SQL injection and XSS attacks
func (sv *SecurityValidator) validateSecurityThreats(input, fieldName string) error {
	lowerInput := strings.ToLower(input)

	// SQL injection detection
	for _, rule := range sv.sqlInjectionRules {
		if rule.MatchString(lowerInput) {
			return &ValidationError{
				Field:   fieldName,
				Message: "potential SQL injection detected",
				Code:    "SECURITY_THREAT_SQL",
			}
		}
	}

	// XSS detection
	for _, pattern := range sv.xssPatterns {
		if pattern.MatchString(input) {
			return &ValidationError{
				Field:   fieldName,
				Message: "potential XSS attack detected",
				Code:    "SECURITY_THREAT_XSS",
			}
		}
	}

	return nil
}

// SanitizeString sanitizes input string by removing dangerous characters
func (sv *SecurityValidator) SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newline and tab
	sanitized := strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\t' && r != '\r' {
			return -1
		}
		return r
	}, input)

	// Trim whitespace
	return strings.TrimSpace(sanitized)
}

// ValidateStruct validates a struct using struct tags
func (sv *SecurityValidator) ValidateStruct(s interface{}) error {
	return sv.validator.Struct(s)
}

// BatchValidate validates multiple fields at once
func (sv *SecurityValidator) BatchValidate(validations map[string]func() error) []ValidationError {
	var errors []ValidationError

	for field, validationFunc := range validations {
		if err := validationFunc(); err != nil {
			if ve, ok := err.(*ValidationError); ok {
				errors = append(errors, *ve)
			} else {
				errors = append(errors, ValidationError{
					Field:   field,
					Message: err.Error(),
					Code:    "VALIDATION_ERROR",
				})
			}
		}
	}

	return errors
}

// Rate limiting constants
const (
	MaxRequestsPerMinute = 60
	MaxRequestsPerHour   = 1000
	MaxRequestsPerDay    = 10000
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
	BurstSize         int
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerMinute: MaxRequestsPerMinute,
		RequestsPerHour:   MaxRequestsPerHour,
		RequestsPerDay:    MaxRequestsPerDay,
		BurstSize:         10,
	}
}
