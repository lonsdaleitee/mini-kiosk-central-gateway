package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	db *sql.DB
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{
		db: db,
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
	checkQuery := `SELECT id FROM users WHERE username = $1 OR email = $2 LIMIT 1`
	err := h.db.QueryRow(checkQuery, req.Username, req.Email).Scan(&existingUserID)

	if err != sql.ErrNoRows {
		if err == nil {
			// User exists
			c.JSON(http.StatusConflict, gin.H{
				"error": "User with this username or email already exists",
			})
			return
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
		INSERT INTO users (username, email, first_name, last_name, password_hash) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id`

	err = h.db.QueryRow(insertQuery, req.Username, req.Email, req.FirstName, req.LastName, string(hashedPassword)).Scan(&userID)
	if err != nil {
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
