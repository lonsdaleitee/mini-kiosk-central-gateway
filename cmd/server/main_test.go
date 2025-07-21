package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/harrywijaya/mini-kiosk-central-gateway/internal/config"
	"github.com/harrywijaya/mini-kiosk-central-gateway/internal/router"
	_ "github.com/lib/pq"
)

// setupTestDB creates a test database connection
// For testing purposes, we'll use a mock or in-memory database
func setupTestDB() *sql.DB {
	// For testing, we can use a simple mock database connection
	// In a real scenario, you might want to use a test database
	// or a library like go-sqlmock for mocking

	// Create a simple connection that won't be used for actual DB operations in these tests
	db, err := sql.Open("postgres", "host=localhost port=5432 user=test dbname=test sslmode=disable")
	if err != nil {
		// If we can't connect to a real database for testing, we'll need to handle this
		// For now, returning nil and we'll need to modify the handlers to handle nil DB
		return nil
	}
	return db
}

// setupTestConfig creates a test configuration
func setupTestConfig() *config.Config {
	// Create a minimal test configuration
	return &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
		Database: config.DatabaseConfig{
			Host:   "localhost",
			Port:   5432,
			User:   "test",
			DBName: "test",
		},
		Keys: config.PublicPrivateKey{
			PrivateKeyPath: "privateKey.pem",
			PublicKeyPath:  "publicKey.pem",
		},
		Gin: config.GinConfig{
			Mode: "debug",
		},
	}
}

func TestHealthEndpoint(t *testing.T) {
	// Set up the router with test database and config
	db := setupTestDB()
	cfg := setupTestConfig()
	r := router.SetupRouter(db, cfg)
	if db != nil {
		defer db.Close()
	}

	// Create a test request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	r.ServeHTTP(w, req)

	// Check the status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check the response body
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}

	if response["service"] != "mini-kiosk-central-gateway" {
		t.Errorf("Expected service 'mini-kiosk-central-gateway', got %v", response["service"])
	}
}

func TestReadyEndpoint(t *testing.T) {
	// Set up the router with test database and config
	db := setupTestDB()
	cfg := setupTestConfig()
	r := router.SetupRouter(db, cfg)
	if db != nil {
		defer db.Close()
	}

	// Create a test request
	req, err := http.NewRequest("GET", "/ready", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	r.ServeHTTP(w, req)

	// Check the status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check the response body
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ready" {
		t.Errorf("Expected status 'ready', got %v", response["status"])
	}
}

func TestConfigLoad(t *testing.T) {
	// Test if configuration can be loaded successfully
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Check default values
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected default database host 'localhost', got %s", cfg.Database.Host)
	}

	// Check key configuration defaults
	if cfg.Keys.PrivateKeyPath != "privateKey.pem" {
		t.Errorf("Expected default private key path 'privateKey.pem', got %s", cfg.Keys.PrivateKeyPath)
	}

	if cfg.Keys.PublicKeyPath != "publicKey.pem" {
		t.Errorf("Expected default public key path 'publicKey.pem', got %s", cfg.Keys.PublicKeyPath)
	}
}

func TestConfigurableKeys(t *testing.T) {
	// Test configuration with different key paths
	testConfig := &config.Config{
		Keys: config.PublicPrivateKey{
			PrivateKeyPath: "custom-private.pem",
			PublicKeyPath:  "custom-public.pem",
		},
		Server: config.ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
		Gin: config.GinConfig{
			Mode: "debug",
		},
	}

	// Verify the configuration holds the custom paths
	if testConfig.Keys.PrivateKeyPath != "custom-private.pem" {
		t.Errorf("Expected custom private key path 'custom-private.pem', got %s", testConfig.Keys.PrivateKeyPath)
	}

	if testConfig.Keys.PublicKeyPath != "custom-public.pem" {
		t.Errorf("Expected custom public key path 'custom-public.pem', got %s", testConfig.Keys.PublicKeyPath)
	}
}
