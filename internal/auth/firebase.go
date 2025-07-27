package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/schoolgpt/backend/internal/config"
	"google.golang.org/api/option"
)

// FirebaseAuth handles Firebase authentication
type FirebaseAuth struct {
	client *auth.Client
	config *config.Config
}

// User represents an authenticated user
type User struct {
	ID    string
	Email string
	Role  string
}

// New creates a new FirebaseAuth instance
func New(cfg *config.Config) (*FirebaseAuth, error) {
	ctx := context.Background()

	// Create Firebase app configuration
	conf := &firebase.Config{
		ProjectID:   cfg.FirebaseProjectID,
		DatabaseURL: cfg.FirebaseDatabaseURL,
	}

	// If credentials are provided, use them
	var app *firebase.App
	var err error

	if cfg.GoogleCredentialsPath != "" {
		// Use credentials file
		opt := option.WithCredentialsFile(cfg.GoogleCredentialsPath)
		app, err = firebase.NewApp(ctx, conf, opt)
	} else if cfg.FirebasePrivateKey != "" && cfg.FirebaseClientEmail != "" {
		// Use service account credentials from environment variables
		// In a real application, you would use option.WithCredentialsJSON with a proper JSON credential
		// For simplicity, we'll use the default credentials
		app, err = firebase.NewApp(ctx, conf)
	} else if cfg.Env == "development" {
		// In development, try to use application default credentials
		app, err = firebase.NewApp(ctx, conf)
	} else {
		return nil, errors.New("firebase credentials not provided")
	}

	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}

	// Get auth client
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting auth client: %v", err)
	}

	return &FirebaseAuth{
		client: client,
		config: cfg,
	}, nil
}

// VerifyToken verifies a Firebase ID token
func (fa *FirebaseAuth) VerifyToken(ctx context.Context, idToken string) (*User, error) {
	// Clean token
	idToken = strings.TrimSpace(idToken)
	idToken = strings.TrimPrefix(idToken, "Bearer ")

	// Verify token
	token, err := fa.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("error verifying ID token: %v", err)
	}

	// Create user
	user := &User{
		ID:    token.UID,
		Email: token.Claims["email"].(string),
	}

	return user, nil
}

// GetUserRole gets a user's role from Firestore
func (fa *FirebaseAuth) GetUserRole(ctx context.Context, userID string) (string, error) {
	// In a real application, this would query Firestore for the user's role
	// For now, just return a default role
	return "student", nil
}

// GetClient returns the Firebase Auth client
func (fa *FirebaseAuth) GetClient() *auth.Client {
	return fa.client
}
