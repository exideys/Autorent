package main

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"autorent-backend/internal/config"
	"autorent-backend/internal/handlers"
	"autorent-backend/internal/repositories"
	aiservice "autorent-backend/internal/services/ai"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	log.Printf("Gemini API key configured: %t", cfg.GeminiAPIKey != "")
	log.Printf("Gemini model: %s", cfg.GeminiModel)
	// Connect to database
	/*db, err := sql.Open("mysql", cfg.DatabaseDSN)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}*/

	// Connect to database if it is available.
	// For now, AI car recommendations use mock data, so the backend can run without DB.
	db, err := sql.Open("mysql", cfg.DatabaseDSN)
	if err != nil {
		log.Printf("Database initialization skipped: %v", err)
	} else {
		defer db.Close()

		if err := db.Ping(); err != nil {
			log.Printf("Database is unavailable, continuing without DB: %v", err)
		} else {
			log.Println("Database connection established")
		}
	}

	// Initialize Gin router
	r := gin.Default()

	// CORS middleware
	allowedOrigins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")
	allowCredentials := true
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "" {
		allowedOrigins = []string{"*"}
		allowCredentials = false
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: allowCredentials,
	}))

	// Dependencies
	carRepository := repositories.NewMockCarRepository()
	carRecommendationService := aiservice.NewCarRecommendationService(
		cfg.GeminiAPIKey,
		cfg.GeminiModel,
		carRepository,
	)
	aiHandler := handlers.NewAIHandler(carRecommendationService)

	// Routes
	r.GET("/health", handlers.HealthHandler)

	api := r.Group("/api")
	{
		api.POST("/ai/car-recommendation", aiHandler.RecommendCar)
	}

	// Get port from environment or default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
