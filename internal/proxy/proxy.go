package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harrywijaya/mini-kiosk-central-gateway/internal/config"
)

// ProxyHandler handles proxying requests to downstream services
type ProxyHandler struct {
	client   *http.Client
	services *config.ServicesConfig
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(services *config.ServicesConfig) *ProxyHandler {
	return &ProxyHandler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		services: services,
	}
}

// ProxyToService forwards a request to the specified service
func (p *ProxyHandler) ProxyToService(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var baseURL string

		// Get the base URL for the service
		switch serviceName {
		case "auth":
			baseURL = p.services.AuthService.BaseURL
		case "order":
			baseURL = p.services.OrderService.BaseURL
		case "inventory":
			baseURL = p.services.InventoryService.BaseURL
		case "payment":
			baseURL = p.services.PaymentService.BaseURL
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown service"})
			return
		}

		// Build the target URL
		targetURL := baseURL + c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}

		// Read the request body
		var bodyBytes []byte
		var err error
		if c.Request.Body != nil {
			bodyBytes, err = io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
				return
			}
		}

		// Create the proxy request
		req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewReader(bodyBytes))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy request"})
			return
		}

		// Copy headers from original request
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Add request ID for tracing
		if requestID, exists := c.Get("request_id"); exists {
			req.Header.Set("X-Request-ID", requestID.(string))
		}

		// Make the request to the downstream service
		resp, err := p.client.Do(req)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service unavailable",
				"service": serviceName,
				"message": fmt.Sprintf("Failed to connect to %s service", serviceName),
			})
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Copy response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
			return
		}

		// Return the response
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}
