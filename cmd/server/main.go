package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/harrywijaya/mini-kiosk-central-gateway/internal/config"
	"github.com/harrywijaya/mini-kiosk-central-gateway/internal/database"
	"github.com/harrywijaya/mini-kiosk-central-gateway/internal/router"
	"github.com/harrywijaya/mini-kiosk-central-gateway/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if cfg.Gin.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database configuration from config
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	// Override configuration from environment variables if provided
	if host := os.Getenv("DB_HOST"); host != "" {
		dbConfig.Host = host
	}
	if user := os.Getenv("DB_USER"); user != "" {
		dbConfig.User = user
	}
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		dbConfig.Password = pass
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		dbConfig.DBName = name
	}

	// Set up migration configuration
	migrationConfig := database.DefaultMigrationConfig()
	migrationConfig.URL = dbConfig.GetFlywayDSN()
	migrationConfig.User = dbConfig.User
	migrationConfig.Password = dbConfig.Password
	migrationConfig.Location = cfg.Flyway.Locations
	migrationConfig.OutOfOrder = cfg.Flyway.OutOfOrder

	// Run database migrations
	fmt.Println("Running database migrations...")
	if err := database.MigrateDatabase(migrationConfig); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	fmt.Println("Migrations completed successfully")

	// Connect to the database
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("Connected to database successfully")

	// Set up router
	r := router.SetupRouter(db)

	// Create and start server
	srv := server.NewServer(r, cfg)

	fmt.Printf("Starting mini-kiosk central gateway on port %d...\n", cfg.Server.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
