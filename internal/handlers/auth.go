package handlers

import (
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/harrywijaya/mini-kiosk-central-gateway/internal/config"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	db     *sql.DB
	config *config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *sql.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		db:     db,
		config: cfg,
	}
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Password  string `json:"password" binding:"required,min=6"`
}

// RegisterResponse represents the response body for user registration
type RegisterResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type Claims struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Fullname string `json:"full_name"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
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

	// Check if user already exists by username or email
	var existingUserID string
	checkQuery := `SELECT id FROM "user" WHERE username = $1 OR email = $2 LIMIT 1`
	err := h.db.QueryRow(checkQuery, req.Username, req.Email).Scan(&existingUserID)

	if err != sql.ErrNoRows {
		if err == nil {
			// User exists
			c.JSON(http.StatusConflict, gin.H{
				"error": "User with this username or email already exists",
			})
			return
		}
		if gin.Mode() == "debug" {
			fmt.Printf("SELECT id FROM user WHERE username = %s OR email = %s LIMIT 1\n", req.Username, req.Email)
			fmt.Printf("Error found during checking to database : %s\n", err)
		}
		// Database error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check existing user",
		})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process password",
		})
		return
	}

	// Insert new user
	var userID string
	insertQuery := `
		INSERT INTO "user" (username, email, first_name, last_name, password_hash) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id`

	err = h.db.QueryRow(insertQuery, req.Username, req.Email, req.FirstName, req.LastName, string(hashedPassword)).Scan(&userID)
	if err != nil {
		if gin.Mode() == "debug" {
			fmt.Printf("Error during inserting to database : %s\n", err)
			fmt.Printf("Query executed : %s\n", insertQuery)
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	// Return success response with user ID
	c.JSON(http.StatusCreated, RegisterResponse{
		ID:      userID,
		Message: "User registered successfully",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request"})
		return
	}

	var userId, hashedPassword, firstName, lastName, email string
	query := `SELECT id, password_hash, first_name, last_name, email FROM "user" WHERE username = $1`
	err := h.db.QueryRow(query, req.Username).Scan(&userId, &hashedPassword, &firstName, &lastName, &email)

	if err != nil || bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	claims := &Claims{
		Username: req.Username,
		Email:    email,
		Fullname: firstName + " " + lastName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	privateKey, err := LoadRSAPrivateKey(h.config.Keys.PrivateKeyPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read private key"})
		return
	}
	accessToken, err := token.SignedString(privateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	refreshToken := uuid.New().String()

	_, err = h.db.Exec(`
		INSERT INTO "refresh_tokens"(user_id, token, expires_at)
	VALUES ($1, $2, $3);
	`, userId, refreshToken, time.Now().Add(7*24*time.Hour))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save new refresh token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	})
}

// RefreshTokenRequest represents the request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken handles access token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate refresh token from database
	var userID string
	var expiresAt time.Time

	query := `
		SELECT user_id, expires_at 
		FROM refresh_tokens 
		WHERE token = $1 AND expires_at > $2
	`
	err := h.db.QueryRow(query, req.RefreshToken, time.Now()).Scan(&userID, &expiresAt)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Get user details for new access token
	var username, email, firstName, lastName string
	userQuery := `SELECT username, email, first_name, last_name FROM "user" WHERE id = $1`
	err = h.db.QueryRow(userQuery, userID).Scan(&username, &email, &firstName, &lastName)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	// Generate new access token
	claims := &Claims{
		Username: username,
		Email:    email,
		Fullname: firstName + " " + lastName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	privateKey, err := LoadRSAPrivateKey(h.config.Keys.PrivateKeyPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read private key"})
		return
	}

	newAccessToken, err := token.SignedString(privateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new token"})
		return
	}

	// Update last_used_at
	h.db.Exec(`UPDATE refresh_tokens SET last_used_at = $1 WHERE token = $2`, time.Now(), req.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
		"expires_in":   900,
	})
}

// Logout handles user logout by revoking refresh token
func (h *AuthHandler) Logout(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Delete refresh token from database
	_, err := h.db.Exec(`DELETE FROM refresh_tokens WHERE token = $1`, req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func LoadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func LoadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub.(*rsa.PublicKey), nil
}
