package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed static/*
var staticFS embed.FS

var staticSubFS = mustSub(staticFS, "static")

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.Handle("/", s.spaHandler())

	mux.HandleFunc("/api/auth/init", s.handleAuthInit)
	mux.HandleFunc("/api/auth/login", s.handleAuthLogin)
	mux.HandleFunc("/api/auth/status", s.handleAuthStatus)
	mux.HandleFunc("/api/auth/logout", s.handleAuthLogout)

	mux.HandleFunc("/api/config", s.authMiddleware(s.handleGetConfig))

	mux.HandleFunc("/api/subscriptions", s.authMiddleware(s.handleSubscriptions))
	mux.HandleFunc("/api/subscriptions/", s.authMiddleware(s.handleSubscriptionDetail))

	mux.HandleFunc("/api/nodes", s.authMiddleware(s.handleNodes))
	mux.HandleFunc("/api/nodes/", s.authMiddleware(s.handleNodesOrDelete))

	mux.HandleFunc("/api/proxy/start", s.authMiddleware(s.handleProxyStart))
	mux.HandleFunc("/api/proxy/stop", s.authMiddleware(s.handleProxyStop))
	mux.HandleFunc("/api/proxy/status", s.authMiddleware(s.handleProxyStatus))
	mux.HandleFunc("/api/proxy/test", s.authMiddleware(s.handleProxyTest))

	mux.HandleFunc("/api/ws", s.handleWebSocket)

	mux.HandleFunc("/api/settings/route-mode", s.authMiddleware(s.handleRouteMode))
	mux.HandleFunc("/api/settings/whitelist", s.authMiddleware(s.handleWhitelist))
	mux.HandleFunc("/api/settings/blacklist", s.authMiddleware(s.handleBlacklist))
	mux.HandleFunc("/api/settings/proxy-ports", s.authMiddleware(s.handleProxyPorts))
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

		filePath := strings.TrimPrefix(r.URL.Path, "/")
		filePath = strings.TrimPrefix(filePath, "static/")
		if filePath == "" {
			filePath = "index.html"
		}

		data, err := fs.ReadFile(staticSubFS, filePath)
		if err != nil {
			data, err = fs.ReadFile(staticSubFS, "index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			filePath = "index.html"
		}

		contentType := "text/html; charset=utf-8"
		if strings.HasSuffix(filePath, ".css") {
			contentType = "text/css; charset=utf-8"
		} else if strings.HasSuffix(filePath, ".js") {
			contentType = "application/javascript; charset=utf-8"
		} else if strings.HasSuffix(filePath, ".svg") {
			contentType = "image/svg+xml"
		} else if strings.HasSuffix(filePath, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(filePath, ".ico") {
			contentType = "image/x-icon"
		} else if strings.HasSuffix(filePath, ".woff2") {
			contentType = "font/woff2"
		} else if strings.HasSuffix(filePath, ".woff") {
			contentType = "font/woff"
		}
		w.Header().Set("Content-Type", contentType)
		w.Write(data)
	})
}

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
