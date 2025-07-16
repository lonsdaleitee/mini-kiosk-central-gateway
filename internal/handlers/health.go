package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck returns the health status of the gateway
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "mini-kiosk-central-gateway",
		"version": "1.0.0",
	})
}

// ReadinessCheck returns the readiness status of the gateway
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// TODO: Add checks for database connectivity and downstream services
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": gin.H{
			"database": "ok",
			// Add more service checks as needed
		},
	})
}
