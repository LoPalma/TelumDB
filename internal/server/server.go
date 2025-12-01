package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/telumdb/telumdb/internal/config"
	"github.com/telumdb/telumdb/pkg/storage"
	"go.uber.org/zap"
)

// Server represents the TelumDB server
type Server struct {
	config     *config.Config
	storage    storage.Engine
	logger     *zap.Logger
	httpServer *http.Server
	listener   net.Listener
}

// New creates a new server instance
func New(cfg *config.Config, storageEngine storage.Engine, logger *zap.Logger) (*Server, error) {
	srv := &Server{
		config:  cfg,
		storage: storageEngine,
		logger:  logger,
	}

	// Initialize HTTP server for API endpoints
	srv.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.HTTPPort),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Setup routes
	srv.setupRoutes()

	return srv, nil
}

// Start starts the server
func (s *Server) Start(ctx context.Context) error {
	// Start storage engine
	if err := s.storage.Start(ctx); err != nil {
		return fmt.Errorf("failed to start storage engine: %w", err)
	}

	// Start HTTP server
	go func() {
		s.logger.Info("Starting HTTP server",
			zap.String("address", s.httpServer.Addr),
		)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Start database protocol server
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.config.Server.Port, err)
	}
	s.listener = listener

	s.logger.Info("TelumDB server started",
		zap.String("host", s.config.Server.Host),
		zap.Int("port", s.config.Server.Port),
		zap.Int("http_port", s.config.Server.HTTPPort),
	)

	// Accept connections
	go s.acceptConnections(ctx)

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Error shutting down HTTP server", zap.Error(err))
	}

	// Close listener
	if s.listener != nil {
		s.listener.Close()
	}

	// Shutdown storage engine
	if err := s.storage.Shutdown(ctx); err != nil {
		s.logger.Error("Error shutting down storage engine", zap.Error(err))
	}

	s.logger.Info("Server shutdown complete")
	return nil
}

// setupRoutes sets up HTTP routes
func (s *Server) setupRoutes() {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	// Metrics endpoint
	mux.HandleFunc("/metrics", s.handleMetrics)

	// API endpoints
	mux.HandleFunc("/api/v1/", s.handleAPI)

	s.httpServer.Handler = mux
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
}

// handleMetrics handles metrics requests
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement metrics collection
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "# TelumDB metrics\n# TODO: Implement metrics\n")
}

// handleAPI handles API requests
func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement API endpoints
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprint(w, `{"error":"API not yet implemented"}`)
}

// acceptConnections accepts database connections
func (s *Server) acceptConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					s.logger.Error("Error accepting connection", zap.Error(err))
					continue
				}
			}

			// Handle connection in goroutine
			go s.handleConnection(conn)
		}
	}
}

// handleConnection handles a single database connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	s.logger.Info("New connection established",
		zap.String("remote_addr", conn.RemoteAddr().String()),
	)

	// TODO: Implement database protocol handling
	// For now, just echo back messages
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			s.logger.Debug("Connection closed", zap.Error(err))
			return
		}

		// Echo back the message
		if _, err := conn.Write(buffer[:n]); err != nil {
			s.logger.Error("Error writing to connection", zap.Error(err))
			return
		}
	}
}
