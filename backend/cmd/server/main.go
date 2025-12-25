package main

import (
	"log"
	"os"

	"github.com/L0g1cV/CodeEcho/backend/internal/database"
	"github.com/L0g1cV/CodeEcho/backend/internal/handlers"
	"github.com/L0g1cV/CodeEcho/backend/internal/middleware"
	"github.com/L0g1cV/CodeEcho/backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Setup Gin router
	router := gin.Default()

	// Global middleware
	router.Use(middleware.CORSMiddleware())

	// Initialize MinIO storage (optional - if MinIO is not available, uploads will fail)
	var uploadHandler *handlers.UploadHandler
	minioClient, err := storage.NewMinIOClientFromEnv()
	if err != nil {
		log.Printf("Warning: MinIO not available, image uploads disabled: %v", err)
	} else {
		uploadHandler = handlers.NewUploadHandler(minioClient)
		log.Println("MinIO storage initialized successfully")
	}

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", handlers.GetCurrentUser)
				users.PUT("/me/llm-config", handlers.UpdateLLMConfig)
			}

			// Unified Note routes (handles notebooks, folders, and notes)
			notes := protected.Group("/notes")
			{
				notes.GET("", handlers.GetNotes)
				notes.POST("", handlers.CreateNote)
				notes.GET("/:id", handlers.GetNote)
				notes.PUT("/:id", handlers.UpdateNote)
				notes.DELETE("/:id", handlers.DeleteNote)
			}

			// Upload route (requires MinIO)
			if uploadHandler != nil {
				protected.POST("/upload", uploadHandler.Upload)
			}
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
