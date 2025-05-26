package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Config holds database connection configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DefaultConfig returns a default database configuration
func DefaultConfig() Config {
	return Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "kiosk_gateway",
		SSLMode:  "disable",
	}
}

// Connect establishes a connection to the database
func Connect(config Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// GetFlywayDSN converts a database Config to a Flyway JDBC connection string
func (c Config) GetFlywayDSN() string {
	return fmt.Sprintf("jdbc:postgresql://%s:%d/%s", c.Host, c.Port, c.DBName)
}
