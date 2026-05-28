package web

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"xray-go/config"
	"xray-go/subscription"
)

// ProxyServer matches the interface in main.go
type ProxyServer interface {
	Stop() error
}

// Server manages the web UI HTTP server and proxy state
type Server struct {
	httpServer  *http.Server
	auth        *AuthManager
	cfg         *config.Config
	proxy       ProxyServer
	currentNode *subscription.Node
	isRunning   bool
	httpPort    int
	socksPort   int
	mu          sync.RWMutex
}

// NewServer creates a new web server
func NewServer(addr string, cfg *config.Config) (*Server, error) {
	auth, err := NewAuthManager()
	if err != nil {
		return nil, fmt.Errorf("auth init: %w", err)
	}

	s := &Server{
		auth: auth,
		cfg:  cfg,
	}

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

// Start begins listening
func (s *Server) Start() error {
	fmt.Printf("Web UI running at http://%s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Stop shuts down the server
func (s *Server) Stop() error {
	if s.proxy != nil {
		if err := s.proxy.Stop(); err != nil {
			return fmt.Errorf("proxy stop: %w", err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
