package handlers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
	users         []MockUser
	refreshTokens []MockRefreshToken
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
}

// MockRefreshToken represents a refresh token in our mock database
type MockRefreshToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// NewMockAuthHandler creates a new mock auth handler
func NewMockAuthHandler() *MockAuthHandler {
	// Generate test RSA keys
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := &privateKey.PublicKey

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	return &MockAuthHandler{
		users: []MockUser{
			{
				ID:           "existing-user-id",
				Username:     "existinguser",
				Email:        "existing@example.com",
				FirstName:    "Existing",
				LastName:     "User",
				PasswordHash: string(hashedPassword),
			},
		},
		refreshTokens: []MockRefreshToken{},
		privateKey:    privateKey,
		publicKey:     publicKey,
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

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	// Create new user
	userID := h.createUser(req, string(hashedPassword))

	// Return success response
	c.JSON(http.StatusCreated, RegisterResponse{
		ID:      userID,
		Message: "User registered successfully",
	})
}

// Login handles user login (mock version)
func (h *MockAuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request"})
		return
	}

	// Find user
	var user *MockUser
	for _, u := range h.users {
		if u.Username == req.Username {
			user = &u
			break
		}
	}

	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	claims := &Claims{
		Username: user.Username,
		Email:    user.Email,
		Fullname: user.FirstName + " " + user.LastName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	accessToken, err := token.SignedString(h.privateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	refreshToken := uuid.New().String()

	// Store refresh token in mock storage
	h.refreshTokens = append(h.refreshTokens, MockRefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	})

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	})
}

// JWTAuthMiddleware for testing
func (h *MockAuthHandler) JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			c.Abort()
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return h.publicKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*Claims); ok {
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
		}
	}
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

// ============ LOGIN TESTS ============

func TestAuthHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockAuthHandler := NewMockAuthHandler()
	router.POST("/login", mockAuthHandler.Login)

	loginReq := LoginRequest{
		Username: "existinguser",
		Password: "password123",
	}

	jsonBody, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.AccessToken == "" {
		t.Error("Expected access token to be returned")
	}

	if response.RefreshToken == "" {
		t.Error("Expected refresh token to be returned")
	}

	if response.ExpiresIn != 900 {
		t.Errorf("Expected expires_in to be 900, got %d", response.ExpiresIn)
	}

	// Verify refresh token was stored
	if len(mockAuthHandler.refreshTokens) != 1 {
		t.Errorf("Expected 1 refresh token in storage, got %d", len(mockAuthHandler.refreshTokens))
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockAuthHandler := NewMockAuthHandler()
	router.POST("/login", mockAuthHandler.Login)

	loginReq := LoginRequest{
		Username: "existinguser",
		Password: "wrongpassword",
	}

	jsonBody, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["error"] != "Invalid credentials" {
		t.Errorf("Expected 'Invalid credentials', got '%s'", response["error"])
	}
}

func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockAuthHandler := NewMockAuthHandler()
	router.POST("/login", mockAuthHandler.Login)

	loginReq := LoginRequest{
		Username: "nonexistentuser",
		Password: "password123",
	}

	jsonBody, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockAuthHandler := NewMockAuthHandler()
	router.POST("/login", mockAuthHandler.Login)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// ============ JWT MIDDLEWARE TESTS ============

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockAuthHandler := NewMockAuthHandler()

	// Protected route
	router.GET("/protected", mockAuthHandler.JWTAuthMiddleware(), func(c *gin.Context) {
		username := c.GetString("username")
		email := c.GetString("email")
		c.JSON(http.StatusOK, gin.H{
			"message":  "Access granted",
			"username": username,
			"email":    email,
		})
	})

	// First login to get a valid token
	loginReq := LoginRequest{
		Username: "existinguser",
		Password: "password123",
	}

	// Create login request
	jsonBody, _ := json.Marshal(loginReq)
	loginRequest, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	loginRequest.Header.Set("Content-Type", "application/json")

	// Add login route temporarily
	router.POST("/login", mockAuthHandler.Login)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, loginRequest)

	var loginResponse LoginResponse
	json.Unmarshal(w.Body.Bytes(), &loginResponse)

	// Now test protected route with valid token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "Access granted" {
		t.Errorf("Expected 'Access granted', got '%s'", response["message"])
	}

	if response["username"] != "existinguser" {
		t.Errorf("Expected 'existinguser', got '%s'", response["username"])
	}
}

func TestJWTAuthMiddleware_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockAuthHandler := NewMockAuthHandler()

	router.GET("/protected", mockAuthHandler.JWTAuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Access granted"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["error"] != "No token provided" {
		t.Errorf("Expected 'No token provided', got '%s'", response["error"])
	}
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockAuthHandler := NewMockAuthHandler()

	router.GET("/protected", mockAuthHandler.JWTAuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Access granted"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["error"] != "Invalid token" {
		t.Errorf("Expected 'Invalid token', got '%s'", response["error"])
	}
}

func TestJWTAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockAuthHandler := NewMockAuthHandler()

	router.GET("/protected", mockAuthHandler.JWTAuthMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Access granted"})
	})

	// Create an expired token
	claims := &Claims{
		Username: "existinguser",
		Email:    "existing@example.com",
		Fullname: "Existing User",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	expiredToken, _ := token.SignedString(mockAuthHandler.privateKey)

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
