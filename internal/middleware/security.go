package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/schoolgpt/backend/internal/errors"
	"github.com/schoolgpt/backend/internal/security"
	"go.uber.org/zap"
)

// SecurityMiddleware provides comprehensive security middleware
type SecurityMiddleware struct {
	accessControl   *security.AccessControl
	sessionManager  *security.SessionManager
	passwordManager *security.PasswordManager
	validator       *security.SecurityValidator
	errorHandler    *errors.ErrorHandler
	logger          *zap.Logger
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(logger *zap.Logger) *SecurityMiddleware {
	return &SecurityMiddleware{
		accessControl:   security.NewAccessControl(),
		sessionManager:  security.NewSessionManager(time.Hour * 24), // 24 hour sessions
		passwordManager: security.NewPasswordManager(),
		validator:       security.NewSecurityValidator(),
		errorHandler:    errors.NewErrorHandler(logger),
		logger:          logger,
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func (sm *SecurityMiddleware) RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func (sm *SecurityMiddleware) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict transport security (HTTPS only)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Feature Policy
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		// Cross-Origin-Opener-Policy - Use unsafe-none for Firebase Auth compatibility
		c.Header("Cross-Origin-Opener-Policy", "unsafe-none")

		c.Next()
	}
}

// CORSMiddleware handles CORS with security considerations
func (sm *SecurityMiddleware) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Define allowed origins (in production, this should be configurable)
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173", // Vite dev server
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173", // Vite dev server
			"https://admin.schoolgpt.com",
			"https://schoolgpt.com",
		}

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AuthenticationMiddleware validates user authentication
func (sm *SecurityMiddleware) AuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for public endpoints
		if sm.isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			sm.handleAuthError(c, "missing authorization header")
			return
		}

		// Parse Bearer token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			sm.handleAuthError(c, "invalid authorization format")
			return
		}

		token := tokenParts[1]

		// Validate session
		ctx, err := sm.sessionManager.GetSession(token)
		if err != nil {
			sm.handleAuthError(c, "invalid or expired session")
			return
		}

		// Update security context
		ctx.RequestID = c.GetString("request_id")
		ctx.IPAddress = c.ClientIP()
		ctx.UserAgent = security.SanitizeUserAgent(c.Request.UserAgent())

		// Store security context in gin context
		c.Set("security_context", ctx)
		c.Set("user_id", ctx.UserID)
		c.Set("school_id", ctx.SchoolID)
		c.Set("role", ctx.Role)

		c.Next()
	}
}

// AuthorizationMiddleware validates user permissions
func (sm *SecurityMiddleware) AuthorizationMiddleware(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get security context
		securityCtx, exists := c.Get("security_context")
		if !exists {
			sm.handleAuthError(c, "security context not found")
			return
		}

		ctx := securityCtx.(*security.SecurityContext)

		// Validate access
		if err := sm.accessControl.ValidateAccess(ctx, resource, action); err != nil {
			sm.logger.Warn("Authorization failed",
				zap.String("user_id", ctx.UserID),
				zap.String("school_id", ctx.SchoolID),
				zap.String("role", ctx.Role),
				zap.String("resource", resource),
				zap.String("action", action),
				zap.Error(err),
			)

			if secErr, ok := err.(*security.SecurityError); ok {
				switch secErr.Code {
				case "RATE_LIMIT_EXCEEDED":
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error": "Rate limit exceeded",
						"code":  "RATE_LIMIT_EXCEEDED",
					})
				case "INSUFFICIENT_PERMISSIONS":
					c.JSON(http.StatusForbidden, gin.H{
						"error": "Insufficient permissions",
						"code":  "INSUFFICIENT_PERMISSIONS",
					})
				default:
					c.JSON(http.StatusForbidden, gin.H{
						"error": "Access denied",
						"code":  "ACCESS_DENIED",
					})
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Authorization check failed",
					"code":  "AUTHORIZATION_ERROR",
				})
			}

			c.Abort()
			return
		}

		c.Next()
	}
}

// InputValidationMiddleware validates and sanitizes input
func (sm *SecurityMiddleware) InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate request size
		if c.Request.ContentLength > 10*1024*1024 { // 10MB limit
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request body too large",
				"code":  "REQUEST_TOO_LARGE",
			})
			c.Abort()
			return
		}

		// Validate Content-Type for POST/PUT requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if !sm.isValidContentType(contentType) {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error": "Unsupported content type",
					"code":  "UNSUPPORTED_MEDIA_TYPE",
				})
				c.Abort()
				return
			}
		}

		// Validate common query parameters
		if err := sm.validateQueryParams(c); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"code":  "INVALID_QUERY_PARAMS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitingMiddleware applies rate limiting
func (sm *SecurityMiddleware) RateLimitingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		ipAddress := c.ClientIP()

		// Create a temporary security context for rate limiting
		ctx := &security.SecurityContext{
			UserID:    userID,
			IPAddress: ipAddress,
		}

		// Check rate limits
		if err := sm.accessControl.ValidateAccess(ctx, "api", "request"); err != nil {
			if secErr, ok := err.(*security.SecurityError); ok && secErr.Code == "RATE_LIMIT_EXCEEDED" {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "Rate limit exceeded",
					"code":        "RATE_LIMIT_EXCEEDED",
					"retry_after": 60,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ErrorHandlingMiddleware handles errors with proper logging and response
func (sm *SecurityMiddleware) ErrorHandlingMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		var err error

		switch v := recovered.(type) {
		case error:
			err = v
		default:
			err = errors.NewSchoolGPTError(
				errors.ErrCodeInternalServer,
				"Internal server error",
				errors.CategorySystem,
				errors.SeverityCritical,
				false,
			)
		}

		sm.errorHandler.HandleError(err, c)
	})
}

// IPWhitelistMiddleware restricts access to whitelisted IPs (for admin endpoints)
func (sm *SecurityMiddleware) IPWhitelistMiddleware(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Check if IP is whitelisted
		allowed := false
		for _, ip := range allowedIPs {
			if clientIP == ip {
				allowed = true
				break
			}
		}

		if !allowed {
			sm.logger.Warn("IP not whitelisted",
				zap.String("ip", clientIP),
				zap.String("path", c.Request.URL.Path),
			)

			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
				"code":  "IP_NOT_WHITELISTED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper methods

func (sm *SecurityMiddleware) isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/health",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/forgot-password",
		"/static/",
	}

	for _, endpoint := range publicEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}

	return false
}

func (sm *SecurityMiddleware) handleAuthError(c *gin.Context, message string) {
	sm.logger.Warn("Authentication failed",
		zap.String("message", message),
		zap.String("ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.String("path", c.Request.URL.Path),
	)

	c.JSON(http.StatusUnauthorized, gin.H{
		"error": "Authentication required",
		"code":  "AUTHENTICATION_REQUIRED",
	})
	c.Abort()
}

func (sm *SecurityMiddleware) isValidContentType(contentType string) bool {
	validTypes := []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
	}

	for _, validType := range validTypes {
		if strings.Contains(contentType, validType) {
			return true
		}
	}

	return false
}

func (sm *SecurityMiddleware) validateQueryParams(c *gin.Context) error {
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			// Validate query parameter names
			if err := sm.validator.ValidatePersonName(key); err != nil {
				return err
			}

			// Validate query parameter values (basic sanitization)
			if strings.Contains(value, "<script") || strings.Contains(value, "javascript:") {
				return &security.ValidationError{
					Field:   key,
					Message: "potentially malicious query parameter",
					Code:    "MALICIOUS_QUERY_PARAM",
				}
			}
		}
	}

	return nil
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// AuthenticationHandler provides authentication endpoints
type AuthenticationHandler struct {
	sessionManager  *security.SessionManager
	passwordManager *security.PasswordManager
	validator       *security.SecurityValidator
	logger          *zap.Logger
}

// NewAuthenticationHandler creates a new authentication handler
func NewAuthenticationHandler(logger *zap.Logger) *AuthenticationHandler {
	return &AuthenticationHandler{
		sessionManager:  security.NewSessionManager(time.Hour * 24),
		passwordManager: security.NewPasswordManager(),
		validator:       security.NewSecurityValidator(),
		logger:          logger,
	}
}

// Login handles user login
func (ah *AuthenticationHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	// Validate input
	if err := ah.validator.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "VALIDATION_FAILED",
		})
		return
	}

	// In a real implementation, you would:
	// 1. Look up user in database
	// 2. Verify password hash
	// 3. Get user permissions

	// For demo purposes, we'll use hardcoded values
	if req.Email == "admin@schoolgpt.com" && req.Password == "Admin123!" {
		// Create session
		ctx, err := ah.sessionManager.CreateSession(
			"admin-user-id",
			"demo-school-id",
			security.RoleSystemAdmin,
			[]string{security.PermissionSystemAdmin},
			c.ClientIP(),
			security.SanitizeUserAgent(c.Request.UserAgent()),
		)
		if err != nil {
			ah.logger.Error("Failed to create session", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create session",
				"code":  "SESSION_CREATION_FAILED",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"token":   ctx.SessionToken,
			"user": gin.H{
				"id":        ctx.UserID,
				"email":     req.Email,
				"role":      ctx.Role,
				"school_id": ctx.SchoolID,
			},
			"expires_at": ctx.ExpiresAt,
		})

		ah.logger.Info("User logged in successfully",
			zap.String("email", req.Email),
			zap.String("ip", c.ClientIP()),
		)

		return
	}

	// Invalid credentials
	c.JSON(http.StatusUnauthorized, gin.H{
		"error": "Invalid credentials",
		"code":  "INVALID_CREDENTIALS",
	})
}

// Logout handles user logout
func (ah *AuthenticationHandler) Logout(c *gin.Context) {
	securityCtx, exists := c.Get("security_context")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
			"code":  "NOT_AUTHENTICATED",
		})
		return
	}

	ctx := securityCtx.(*security.SecurityContext)
	ah.sessionManager.InvalidateSession(ctx.SessionToken)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// GetCurrentUser returns current user information
func (ah *AuthenticationHandler) GetCurrentUser(c *gin.Context) {
	securityCtx, exists := c.Get("security_context")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Not authenticated",
			"code":  "NOT_AUTHENTICATED",
		})
		return
	}

	ctx := securityCtx.(*security.SecurityContext)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":          ctx.UserID,
			"school_id":   ctx.SchoolID,
			"role":        ctx.Role,
			"permissions": ctx.Permissions,
			"login_time":  ctx.LoginTime,
			"expires_at":  ctx.ExpiresAt,
		},
	})
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
