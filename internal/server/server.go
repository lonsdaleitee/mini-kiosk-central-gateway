package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harrywijaya/mini-kiosk-central-gateway/internal/config"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	config     *config.Config
}

// NewServer creates a new server instance
func NewServer(router *gin.Engine, cfg *config.Config) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:        router,
			ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		config: cfg,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create a channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop

	log.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("Server exited gracefully")
	return nil
}
