package router

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/harrywijaya/mini-kiosk-central-gateway/internal/config"
	"github.com/harrywijaya/mini-kiosk-central-gateway/internal/handlers"
	"github.com/harrywijaya/mini-kiosk-central-gateway/internal/middleware"
)

// SetupRouter sets up the main router with all routes and middleware
func SetupRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	// Create Gin router
	r := gin.New()

	// Add middleware
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(gin.Recovery())

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(db, cfg)

	// Health check routes
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/ready", healthHandler.ReadinessCheck)

	// API routes
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			// Authentication routes (to be proxied to auth service)
			auth := v1.Group("/auth")
			{
				auth.POST("/login", authHandler.Login)
				auth.POST("/register", authHandler.Register)
				auth.POST("/logout", authHandler.Logout)
				auth.GET("/refresh", authHandler.RefreshToken)
			}

			// Order routes (to be proxied to order service)
			orders := v1.Group("/orders")
			orders.Use(middleware.JWTAuthMiddleware(cfg.Keys.PublicKeyPath))
			{
				orders.GET("/", placeholderHandler("orders", "list"))
				orders.POST("/", placeholderHandler("orders", "create"))
				orders.GET("/:id", placeholderHandler("orders", "get"))
				orders.PUT("/:id", placeholderHandler("orders", "update"))
				orders.DELETE("/:id", placeholderHandler("orders", "delete"))
			}

			// Inventory routes (to be proxied to inventory service)
			inventory := v1.Group("/inventory")
			inventory.Use(middleware.JWTAuthMiddleware(cfg.Keys.PublicKeyPath))
			{
				inventory.GET("/", placeholderHandler("inventory", "list"))
				inventory.GET("/:id", placeholderHandler("inventory", "get"))
				inventory.PUT("/:id", placeholderHandler("inventory", "update"))
			}

			// Payment routes (to be proxied to payment service)
			payments := v1.Group("/payments")
			{
				payments.POST("/", placeholderHandler("payments", "create"))
				payments.GET("/:id", placeholderHandler("payments", "get"))
				payments.POST("/:id/refund", placeholderHandler("payments", "refund"))
			}
		}
	}

	return r
}

// placeholderHandler is a temporary handler for routes that will be implemented later
func placeholderHandler(service, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "This endpoint will proxy to " + service + " service",
			"action":  action,
			"status":  "not_implemented",
		})
	}
}
