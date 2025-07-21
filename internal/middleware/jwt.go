package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/harrywijaya/mini-kiosk-central-gateway/internal/handlers"
)

// JWTAuthMiddleware creates a JWT authentication middleware
func JWTAuthMiddleware(publicKeyPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &handlers.Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			// Load public key
			return handlers.LoadRSAPublicKey(publicKeyPath)
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*handlers.Claims); ok {
			// Add user info to headers for downstream services
			c.Request.Header.Set("X-User-ID", claims.Username)
			c.Request.Header.Set("X-User-Email", claims.Email)
			c.Request.Header.Set("X-User-Name", claims.Fullname)

			// Set in context for current request
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Set("fullname", claims.Fullname)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
		}
	}
}
