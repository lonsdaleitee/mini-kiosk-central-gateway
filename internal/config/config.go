package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Services ServicesConfig `mapstructure:"services"`
	Gin      GinConfig      `mapstructure:"gin"`
	Flyway   FlywayConfig   `mapstructure:"flyway"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// ServicesConfig holds downstream services configuration
type ServicesConfig struct {
	AuthService      ServiceConfig `mapstructure:"auth_service"`
	OrderService     ServiceConfig `mapstructure:"order_service"`
	InventoryService ServiceConfig `mapstructure:"inventory_service"`
	PaymentService   ServiceConfig `mapstructure:"payment_service"`
}

// ServiceConfig holds individual service configuration
type ServiceConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Timeout int    `mapstructure:"timeout"`
}

// GinConfig holds Gin framework manual configured value
type GinConfig struct {
	Mode string `mapstructure:"mode"`
}

// FlywayConfig holds Flyway migration tools configuration
type FlywayConfig struct {
	Url                     string `mapstructure:"url"`
	User                    string `mapstructure:"user"`
	Password                string `mapstructure:"password"`
	Locations               string `mapstructure:"locations"`
	ConnecRetries           int    `mapstructure:"connectRetries"`
	OutOfOrder              bool   `mapstructure:"outOfOrder"`
	ValidateMigrationNaming bool   `mapstructure:"validateMigrationNaming"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("local.config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")     // From project root
	viper.AddConfigPath("../configs")    // From cmd/server
	viper.AddConfigPath("../../configs") // From internal/config (current location)
	viper.AddConfigPath(".")             // Current directory

	// Enable environment variable support
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GATEWAY") // Environment variables will be prefixed with GATEWAY_

	// Set default values
	setDefaults()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		log.Println("No config file found, using defaults and environment variables")
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "harrywijaya")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "central_gateway_mini_kiosk")
	viper.SetDefault("database.sslmode", "disable")

	// Gin defaults
	viper.SetDefault("gin.mode", "debug")
}
