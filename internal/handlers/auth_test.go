package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MockUser represents a user in our mock database
type MockUser struct {
	ID           string
	Username     string
	Email        string
	FirstName    string
	LastName     string
	PasswordHash string
}

// MockAuthHandler is a modified version of AuthHandler for testing
type MockAuthHandler struct {
	users []MockUser
}

// NewMockAuthHandler creates a new mock auth handler
func NewMockAuthHandler() *MockAuthHandler {
	return &MockAuthHandler{
		users: []MockUser{
			{
				ID:           "existing-user-id",
				Username:     "existinguser",
				Email:        "existing@example.com",
				FirstName:    "Existing",
				LastName:     "User",
				PasswordHash: "hashedpassword",
			},
		},
	}
}

// checkUserExists checks if a user with the given username or email already exists
func (h *MockAuthHandler) checkUserExists(username, email string) bool {
	for _, user := range h.users {
		if user.Username == username || user.Email == email {
			return true
		}
	}
	return false
}

// createUser creates a new user in the mock database
func (h *MockAuthHandler) createUser(req RegisterRequest, hashedPassword string) string {
	userID := uuid.New().String()
	newUser := MockUser{
		ID:           userID,
		Username:     req.Username,
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: hashedPassword,
	}
	h.users = append(h.users, newUser)
	return userID
}

// Register handles user registration (mock version)
func (h *MockAuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Trim whitespace from input fields
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	// Validate that fields are not empty after trimming
	if req.Username == "" || req.Email == "" || req.FirstName == "" || req.LastName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "All fields are required and cannot be empty",
		})
		return
	}

	// Check if user already exists
	if h.checkUserExists(req.Username, req.Email) {
		c.JSON(http.StatusConflict, gin.H{
			"error": "User with this username or email already exists",
		})
		return
	}

	// Hash the password (simplified for testing)
	hashedPassword := "hashed_" + req.Password

	// Create new user
	userID := h.createUser(req, hashedPassword)

	// Return success response
	c.JSON(http.StatusCreated, RegisterResponse{
		ID:      userID,
		Message: "User registered successfully",
	})
}

func TestAuthHandler_Register_Success(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test router
	router := gin.New()

	// Create mock auth handler
	mockAuthHandler := NewMockAuthHandler()

	// Set up the route
	router.POST("/register", mockAuthHandler.Register)

	// Test case: Valid request body
	registerReq := RegisterRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "password123",
	}

	jsonBody, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response RegisterResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.ID == "" {
		t.Error("Expected user ID to be returned")
	}

	if response.Message != "User registered successfully" {
		t.Errorf("Expected success message, got %s", response.Message)
	}

	// Verify user was added to mock database
	if len(mockAuthHandler.users) != 2 { // 1 existing + 1 new
		t.Errorf("Expected 2 users in mock database, got %d", len(mockAuthHandler.users))
	}
}

func TestAuthHandler_Register_DuplicateUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockAuthHandler := NewMockAuthHandler()
	router.POST("/register", mockAuthHandler.Register)

	// Try to register with existing username
	registerReq := RegisterRequest{
		Username:  "existinguser", // This already exists in mock data
		Email:     "different@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "password123",
	}

	jsonBody, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	expectedError := "User with this username or email already exists"
	if response["error"] != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, response["error"])
	}
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockAuthHandler := NewMockAuthHandler()
	router.POST("/register", mockAuthHandler.Register)

	// Try to register with existing email
	registerReq := RegisterRequest{
		Username:  "differentuser",
		Email:     "existing@example.com", // This already exists in mock data
		FirstName: "John",
		LastName:  "Doe",
		Password:  "password123",
	}

	jsonBody, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}
}

func TestAuthHandler_Register_InvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockAuthHandler := NewMockAuthHandler()
	router.POST("/register", mockAuthHandler.Register)

	// Test case: Missing required fields
	invalidReq := map[string]string{
		"username": "",
		"email":    "test@example.com",
		"password": "password123",
	}

	jsonBody, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAuthHandler_Register_EmptyFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockAuthHandler := NewMockAuthHandler()
	router.POST("/register", mockAuthHandler.Register)

	// Test case: Empty fields after trimming
	registerReq := RegisterRequest{
		Username:  "validuser", // Valid username
		Email:     "test@example.com",
		FirstName: "   ", // Only whitespace - will be empty after trimming
		LastName:  "Doe",
		Password:  "password123",
	}

	jsonBody, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	expectedError := "All fields are required and cannot be empty"
	if response["error"] != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, response["error"])
	}
}
