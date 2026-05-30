package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"

	"autorent-backend/internal/ai"
	"autorent-backend/internal/auth"
	"autorent-backend/internal/config"
	"autorent-backend/internal/handlers"
	"autorent-backend/internal/repository"

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

	if cfg.GeminiAPIKey == "" {
		log.Println("GEMINI_API_KEY is not set. AI car assistant will be disabled.")
	} else {
		log.Printf("GEMINI_API_KEY is set. AI car assistant will be enabled using model: %s", cfg.GeminiModel)
	}

	// Connect to database
	db, err := sql.Open("mysql", cfg.DatabaseDSN)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
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
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Admin-Setup-Token"},
		AllowCredentials: allowCredentials,
	}))

	// Routes
	r.GET("/health", handlers.HealthHandler)
	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTTokenTTL)
	carRepository := repository.NewCarRepository(db)
	userRepository := repository.NewUserRepository(db)
	var aiExtractor ai.CarFilterExtractor
	aiExtractor, err = ai.NewGeminiExtractor(context.Background(), cfg.GeminiAPIKey, cfg.GeminiModel)
	if err != nil {
		log.Printf("AI car assistant disabled: %v", err)
		aiExtractor = &ai.UnavailableExtractor{}
	}

	api := r.Group("/api")
	handlers.RegisterAuthRoutes(api.Group("/auth"), userRepository, tokenManager, cfg.AdminSetupToken)
	handlers.RegisterCarRoutes(api, carRepository)
	handlers.RegisterAIRoutes(api, carRepository, aiExtractor)

	adminAPI := api.Group("/admin")
	adminAPI.Use(handlers.RequireAdmin(tokenManager))
	handlers.RegisterAdminCarRoutes(adminAPI, carRepository)
	handlers.RegisterAdminUserRoutes(adminAPI, userRepository)

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
