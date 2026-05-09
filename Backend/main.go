package main

import (
	"database/sql"
	"log"
	"os"

	"autorent-backend/internal/config"
	"autorent-backend/internal/handlers"

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
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins for now (adjust for production)
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Routes
	r.GET("/health", handlers.HealthHandler)

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
