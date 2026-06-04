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
	"autorent-backend/internal/services"
	"autorent-backend/internal/storage"
	"autorent-backend/internal/translation"

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
	if cfg.DeepLAPIKey == "" {
		log.Println("DEEPL_API_KEY is not set. Ukrainian translation will be unavailable.")
	} else {
		log.Printf("DEEPL_API_KEY is set. Ukrainian translation will use endpoint: %s", cfg.DeepLAPIURL)
	}
	var imageStorage handlers.ImageStorage
	driveStorage, err := storage.NewGoogleDriveStorage(context.Background(), storage.GoogleDriveConfig{
		OAuthClientID:     cfg.GoogleDriveOAuthClientID,
		OAuthClientSecret: cfg.GoogleDriveOAuthClientSecret,
		OAuthRefreshToken: cfg.GoogleDriveOAuthRefreshToken,
		CarsFolderID:      cfg.GoogleDriveCarsFolderID,
		NewsFolderID:      cfg.GoogleDriveNewsFolderID,
	})
	if err != nil {
		log.Printf("Google Drive image storage disabled: %v", err)
	} else if driveStorage == nil {
		log.Println("Google Drive image storage disabled: folder ids are not configured.")
	} else {
		imageStorage = driveStorage
		log.Printf("Google Drive image storage enabled using %s auth.", driveStorage.AuthMode())
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
	r.Use(cors.New(corsConfig(os.Getenv("CORS_ALLOWED_ORIGINS"))))

	// Routes
	r.GET("/health", handlers.HealthHandler)
	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTTokenTTL)
	carRepository := repository.NewCarRepository(db)
	rentalOrderRepository := repository.NewRentalOrderRepository(db)
	carService := services.NewCarService(carRepository, rentalOrderRepository)
	rentalOrderService := services.NewRentalOrderService(rentalOrderRepository)
	userRepository := repository.NewUserRepository(db)
	newsRepository := repository.NewNewsRepository(db)
	var aiExtractor ai.CarFilterExtractor
	aiExtractor, err = ai.NewGeminiExtractor(context.Background(), cfg.GeminiAPIKey, cfg.GeminiModel)
	if err != nil {
		log.Printf("AI car assistant disabled: %v", err)
		aiExtractor = &ai.UnavailableExtractor{}
	}
	var translator handlers.Translator
	if cfg.DeepLAPIKey != "" {
		translator = translation.NewDeepLTranslator(cfg.DeepLAPIKey, cfg.DeepLAPIURL)
	}

	api := r.Group("/api")
	handlers.RegisterAuthRoutes(api.Group("/auth"), userRepository, tokenManager, cfg.AdminSetupToken, cfg.GoogleAuthClientID)
	handlers.RegisterCarRoutes(api, carService)
	handlers.RegisterRentalOrderRoutes(api, rentalOrderService, tokenManager)
	handlers.RegisterAIRoutes(api, carRepository, aiExtractor)
	handlers.RegisterTranslationRoutes(api, translator)
	handlers.RegisterNewsRoutes(api, newsRepository)
	handlers.RegisterImageRoutes(api, imageStorage)

	adminAPI := api.Group("/admin")
	adminAPI.Use(handlers.RequireAdmin(tokenManager))
	handlers.RegisterAdminUploadRoutes(adminAPI, imageStorage, cfg.ImageUploadMaxBytes)
	handlers.RegisterAdminCarRoutes(adminAPI, carService)
	handlers.RegisterAdminUserRoutes(adminAPI, userRepository)
	handlers.RegisterAdminRentalOrderRoutes(adminAPI, rentalOrderService)
	handlers.RegisterAdminNewsRoutes(adminAPI, newsRepository)

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

func corsConfig(allowedOriginsEnv string) cors.Config {
	allowedOrigins := strings.Split(allowedOriginsEnv, ",")
	allowCredentials := true
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "" {
		allowedOrigins = []string{"*"}
		allowCredentials = false
	}

	return cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Admin-Setup-Token"},
		AllowCredentials: allowCredentials,
	}
}
