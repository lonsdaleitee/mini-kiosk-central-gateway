package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yourusername/mini-kiosk-central-gateway/internal/database"
)

func main() {
	// Initialize database configuration
	dbConfig := database.DefaultConfig()

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

	// Start your application...
	fmt.Println("Starting mini-kiosk central gateway...")

	// Add your application code here
}
