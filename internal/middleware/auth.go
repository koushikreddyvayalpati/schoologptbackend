package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/schoolgpt/backend/internal/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthMiddleware handles authentication and authorization
type AuthMiddleware struct {
	authClient      *auth.Client
	firestoreClient *firestore.Client
	config          *config.Config
}

// New creates a new AuthMiddleware
func New(authClient *auth.Client, firestoreClient *firestore.Client, cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		authClient:      authClient,
		firestoreClient: firestoreClient,
		config:          cfg,
	}
}

// Authenticate verifies the Firebase ID token or test token
func (am *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// Extract the token
		idToken := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
		if idToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "ID token required"})
			return
		}

		// Check if it's a test token first
		if strings.HasPrefix(idToken, "test_") {
			err := am.verifyTestToken(c, idToken)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid test token: %v", err)})
				return
			}
			c.Next()
			return
		}

		// Verify the Firebase token
		token, err := am.authClient.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
			return
		}

		// Set the user ID in the context
		c.Set("user_id", token.UID)
		c.Set("user_email", token.Claims["email"])

		c.Next()
	}
}

// RequireRoles requires the user to have one of the specified roles
func (am *AuthMiddleware) RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user ID from the context
		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Check if role is already set (from test token)
		var role string
		if existingRole, hasRole := c.Get("user_role"); hasRole {
			role = existingRole.(string)
		} else {
			// Get the user's role from Firestore
			var err error
			role, err = am.getUserRole(c.Request.Context(), userID.(string))
			if err != nil {
				if status.Code(err) == codes.NotFound {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User has no role assigned"})
				} else {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error getting user role: %v", err)})
				}
				return
			}
		}

		// Check if the user has one of the required roles
		hasRole := false
		for _, r := range roles {
			if r == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}

		// Set the user's role in the context
		c.Set("user_role", role)

		c.Next()
	}
}

// verifyTestToken verifies a test token and sets user context
func (am *AuthMiddleware) verifyTestToken(c *gin.Context, tokenName string) error {
	// Get the test token from Firestore
	doc, err := am.firestoreClient.Collection("test_tokens").Doc(tokenName).Get(c.Request.Context())
	if err != nil {
		return fmt.Errorf("test token not found: %v", err)
	}

	data := doc.Data()

	// Check token expiration
	expiresAt, ok := data["expires_at"].(int64)
	if !ok {
		return fmt.Errorf("invalid token expiration")
	}

	if expiresAt < time.Now().Unix() {
		return fmt.Errorf("test token expired")
	}

	// Set user context from token data
	if uid, ok := data["uid"].(string); ok {
		c.Set("user_id", uid)
	}
	if email, ok := data["email"].(string); ok {
		c.Set("user_email", email)
	}
	if role, ok := data["role"].(string); ok {
		c.Set("user_role", role)
	}
	if entityID, ok := data["entity_id"].(string); ok {
		c.Set("entity_id", entityID)
	}

	return nil
}

// getUserRole gets a user's role from Firestore
func (am *AuthMiddleware) getUserRole(ctx context.Context, userID string) (string, error) {
	// Get the user document from Firestore
	doc, err := am.firestoreClient.Collection("users").Doc(userID).Get(ctx)
	if err != nil {
		return "", err
	}

	// Get the role field
	data := doc.Data()
	role, ok := data["role"].(string)
	if !ok {
		return "", fmt.Errorf("user has no role field")
	}

	return role, nil
}

// RoleBasedAccess creates middleware for role-based access control
func (am *AuthMiddleware) RoleBasedAccess(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// Extract the token
		idToken := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
		if idToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "ID token required"})
			return
		}

		// Check if it's a test token first
		if strings.HasPrefix(idToken, "test_") {
			err := am.verifyTestToken(c, idToken)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid test token: %v", err)})
				return
			}
		} else {
			// Verify the Firebase token
			token, err := am.authClient.VerifyIDToken(c.Request.Context(), idToken)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
				return
			}
			// Set the user ID in the context
			c.Set("user_id", token.UID)
			c.Set("user_email", token.Claims["email"])
		}

		// Get the user ID from the context
		userID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Check if role is already set (from test token)
		var role string
		if existingRole, hasRole := c.Get("user_role"); hasRole {
			role = existingRole.(string)
		} else {
			// Get the user's role from Firestore
			var err error
			role, err = am.getUserRole(c.Request.Context(), userID.(string))
			if err != nil {
				if status.Code(err) == codes.NotFound {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User has no role assigned"})
				} else {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error getting user role: %v", err)})
				}
				return
			}
		}

		// Check if the user has one of the required roles
		hasRole := false
		for _, r := range allowedRoles {
			if r == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}

		// Set the user's role in the context
		c.Set("user_role", role)

		c.Next()
	}
}
