package database

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MigrationConfig stores Flyway configuration options
type MigrationConfig struct {
	URL        string
	User       string
	Password   string
	Location   string
	Command    string // "migrate", "clean", "info", etc.
	OutOfOrder bool   // Allow out-of-order migrations
}

// DefaultMigrationConfig returns a default migration configuration
func DefaultMigrationConfig() MigrationConfig {
	return MigrationConfig{
		URL:        "jdbc:postgresql://localhost:5432/central_gateway_mini_kiosk",
		User:       "harrywijaya",
		Password:   "",
		Location:   "filesystem:db/migrations",
		Command:    "migrate",
		OutOfOrder: true,
	}
}

// RunFlyway executes Flyway with the provided configuration
func RunFlyway(config MigrationConfig) error {
	// Find the absolute path to the project root
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Make sure the migration directory exists
	migrationsDir := strings.TrimPrefix(config.Location, "filesystem:")
	migrationPath := filepath.Join(currentDir, migrationsDir)

	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", migrationPath)
	}

	// Prepare Flyway command
	args := []string{
		"-url=" + config.URL,
		"-user=" + config.User,
		"-password=" + config.Password,
		"-locations=" + config.Location,
		config.Command,
	}

	if config.OutOfOrder {
		args = append(args, "-outOfOrder=true")
	}

	// Execute Flyway command
	cmd := exec.Command("flyway", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Executing migration command:", strings.Join(cmd.Args, " "))
	return cmd.Run()
}

// MigrateDatabase runs all pending migrations
func MigrateDatabase(config MigrationConfig) error {
	config.Command = "migrate"
	return RunFlyway(config)
}

// CleanDatabase drops all objects in the schema
func CleanDatabase(config MigrationConfig) error {
	config.Command = "clean"
	return RunFlyway(config)
}

// ValidateMigrations checks migration checksums
func ValidateMigrations(config MigrationConfig) error {
	config.Command = "validate"
	return RunFlyway(config)
}

// InfoMigrations prints the details and status of all migrations
func InfoMigrations(config MigrationConfig) error {
	config.Command = "info"
	return RunFlyway(config)
}

// MigrateDatabase runs all pending migrations
func RepairMigrations(config MigrationConfig) error {
	config.Command = "repair"
	return RunFlyway(config)
}
