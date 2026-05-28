package web

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed static/*
var staticFS embed.FS

func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(mustSub(staticFS, "static")))))
	mux.Handle("/", s.spaHandler())

	// Auth APIs (no middleware)
	mux.HandleFunc("/api/auth/init", s.handleAuthInit)
	mux.HandleFunc("/api/auth/login", s.handleAuthLogin)
	mux.HandleFunc("/api/auth/status", s.handleAuthStatus)
	mux.HandleFunc("/api/auth/logout", s.handleAuthLogout)

	// Config APIs (protected)
	mux.HandleFunc("/api/config", s.authMiddleware(s.handleConfig))

	// Subscription APIs (protected)
	mux.HandleFunc("/api/subscriptions", s.authMiddleware(s.handleSubscriptions))
	mux.HandleFunc("/api/subscriptions/", s.authMiddleware(s.handleSubscriptionDetail))

	// Node APIs (protected)
	mux.HandleFunc("/api/nodes", s.authMiddleware(s.handleNodes))
	mux.HandleFunc("/api/nodes/regions", s.authMiddleware(s.handleNodeRegions))

	// Proxy APIs (protected)
	mux.HandleFunc("/api/proxy/start", s.authMiddleware(s.handleProxyStart))
	mux.HandleFunc("/api/proxy/stop", s.authMiddleware(s.handleProxyStop))
	mux.HandleFunc("/api/proxy/status", s.authMiddleware(s.handleProxyStatus))
	mux.HandleFunc("/api/proxy/test", s.authMiddleware(s.handleProxyTest))
}

func mustSub(fsys embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}

func (s *Server) spaHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		staticSub, err := fs.Sub(staticFS, "static")
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if r.URL.Path == "/" {
			r.URL.Path = "/index.html"
		}
		http.FileServer(http.FS(staticSub)).ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Auth handlers
func (s *Server) handleAuthInit(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.auth.extractToken(r)
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		username, err := s.auth.ValidateToken(token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}
		r.Header.Set("X-Username", username)
		next(w, r)
	}
}

// Protected handlers (stubs for Task 3)
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleSubscriptions(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleSubscriptionDetail(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleNodeRegions(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleProxyStart(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleProxyStop(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleProxyStatus(w http.ResponseWriter, r *http.Request) {}
func (s *Server) handleProxyTest(w http.ResponseWriter, r *http.Request) {}
