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
	viper.AddConfigPath("../../configs")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		log.Println("No config file found")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}
