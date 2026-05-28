package main

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	"autorent-backend/internal/auth"
	"autorent-backend/internal/config"
	"autorent-backend/internal/handlers"
	airepositories "autorent-backend/internal/repositories"
	"autorent-backend/internal/repository"
	aiservice "autorent-backend/internal/services/ai"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	log.Printf("Gemini API key configured: %t", cfg.GeminiAPIKey != "")
	log.Printf("Gemini model: %s", cfg.GeminiModel)
	log.Printf("Use mock cars for AI: %t", cfg.UseMockCars)

	db, err := openDatabase(cfg.DatabaseDSN)
	if err != nil {
		if !cfg.UseMockCars {
			log.Fatal("Failed to connect to database:", err)
		}
		log.Printf("Database unavailable, continuing with AI mock cars only: %v", err)
	} else {
		defer db.Close()
		log.Println("Database connection established")
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Admin-Setup-Token"},
		AllowCredentials: allowCredentials(),
	}))

	api := r.Group("/api")
	registerDatabaseRoutes(api, db, cfg)
	registerAIRoutes(api, db, cfg)

	r.GET("/health", handlers.HealthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func openDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func registerDatabaseRoutes(api *gin.RouterGroup, db *sql.DB, cfg *config.Config) {
	if db == nil {
		log.Println("Database-backed auth and car routes are disabled")
		return
	}

	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTTokenTTL)
	carRepository := repository.NewCarRepository(db)
	userRepository := repository.NewUserRepository(db)

	handlers.RegisterAuthRoutes(api.Group("/auth"), userRepository, tokenManager, cfg.AdminSetupToken)
	handlers.RegisterCarRoutes(api, carRepository)

	adminAPI := api.Group("/admin")
	adminAPI.Use(handlers.RequireAdmin(tokenManager))
	handlers.RegisterAdminCarRoutes(adminAPI, carRepository)
}

func registerAIRoutes(api *gin.RouterGroup, db *sql.DB, cfg *config.Config) {
	var carRepository airepositories.CarRepository
	if cfg.UseMockCars || db == nil {
		log.Println("Using mock car repository for AI recommendations")
		carRepository = airepositories.NewMockCarRepository()
	} else {
		log.Println("Using TiDB/MySQL car repository for AI recommendations")
		carRepository = airepositories.NewMySQLCarRepository(db)
	}

	carRecommendationService := aiservice.NewCarRecommendationService(
		cfg.GeminiAPIKey,
		cfg.GeminiModel,
		carRepository,
	)
	aiHandler := handlers.NewAIHandler(carRecommendationService)

	api.POST("/ai/car-recommendation", aiHandler.RecommendCar)
}

func allowedOrigins() []string {
	origins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")
	if len(origins) == 1 && origins[0] == "" {
		return []string{"*"}
	}

	return origins
}

func allowCredentials() bool {
	origins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")
	return !(len(origins) == 1 && origins[0] == "")
}
