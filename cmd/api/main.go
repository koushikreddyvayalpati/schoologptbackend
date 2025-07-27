package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/schoolgpt/backend/internal/auth"
	"github.com/schoolgpt/backend/internal/config"
	"github.com/schoolgpt/backend/internal/handlers"
	"github.com/schoolgpt/backend/internal/middleware"
	"github.com/schoolgpt/backend/internal/routes"
	"github.com/schoolgpt/backend/internal/security"
	"github.com/schoolgpt/backend/internal/services"
	"github.com/schoolgpt/backend/internal/storage"
	"github.com/schoolgpt/backend/pkg/gpt"
	"github.com/schoolgpt/backend/pkg/voice"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Set up Gin
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Configure CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173", "http://127.0.0.1:3000", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add security headers middleware
	r.Use(func(c *gin.Context) {
		// Cross-Origin-Opener-Policy for Firebase Auth compatibility
		c.Header("Cross-Origin-Opener-Policy", "unsafe-none")
		c.Next()
	})

	// Initialize Firebase Auth
	firebaseAuth, err := auth.New(cfg)
	if err != nil {
		log.Fatalf("Error initializing Firebase Auth: %v", err)
	}

	// Initialize Firestore
	firestoreDB, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("Error initializing Firestore: %v", err)
	}
	defer firestoreDB.Close()

	// Initialize GPT client
	gptClient := gpt.New(cfg)

	// Initialize Voice processor
	voiceProcessor, err := voice.New(cfg)
	if err != nil {
		log.Fatalf("Error initializing Voice processor: %v", err)
	}
	defer voiceProcessor.Close()

	// Initialize middleware
	authMiddleware := middleware.New(firebaseAuth.GetClient(), firestoreDB.GetClient(), cfg)

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	defer logger.Sync()

	// Initialize security validator
	securityValidator := security.NewSecurityValidator()

	// Initialize services
	schoolSetupService := services.NewSchoolSetupService(firestoreDB)
	setupAgent := services.NewSetupAgent(gptClient, schoolSetupService, firestoreDB)
	educationalService := services.NewEducationalFeaturesService(firestoreDB.GetClient(), securityValidator, logger)

	// Initialize handlers
	gptHandler := handlers.NewGPTHandler(gptClient, firestoreDB)
	attendanceHandler := handlers.NewAttendanceHandler(firestoreDB)
	voiceHandler := handlers.NewVoiceHandler(voiceProcessor)
	teacherTasksHandler := handlers.NewTeacherTasksHandler(gptClient, firestoreDB)
	schoolSetupHandler := handlers.NewSchoolSetupHandler(schoolSetupService)
	setupAgentHandler := handlers.NewSetupAgentHandler(setupAgent)
	educationalHandler := handlers.NewEducationalHandler(educationalService, logger)
	userManagementHandler := handlers.NewUserManagementHandler(firestoreDB, logger)
	teacherAnalyticsHandler := handlers.NewTeacherAnalyticsHandler(firestoreDB, gptClient)

	// Set up routes
	routes.SetupRoutes(r, authMiddleware, gptHandler, attendanceHandler, voiceHandler, teacherTasksHandler, schoolSetupHandler, setupAgentHandler, educationalHandler, userManagementHandler, teacherAnalyticsHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
