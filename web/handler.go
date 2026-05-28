package web

import (
	"embed"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"xray-go/config"
	"xray-go/latency"
	"xray-go/region"
	"xray-go/singbox"
	"xray-go/subscription"
	"xray-go/xrayproxy"
)

//go:embed static/*
var staticFS embed.FS

var staticSubFS = mustSub(staticFS, "static")

func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Static files served via SPA handler
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
		}
		w.Header().Set("Content-Type", contentType)
		w.Write(data)
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

// Auth handlers
func (s *Server) handleAuthInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if s.auth.HasUser() {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "user already exists"})
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := s.auth.CreateUser(req.Username, req.Password); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	token, err := s.auth.ValidateUser(req.Username, req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":    token,
		"username": req.Username,
	})
}

func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	token, err := s.auth.ValidateUser(req.Username, req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":    token,
		"username": req.Username,
	})
}

func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{
		"initialized": s.auth.HasUser(),
	})
}

func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "logged out",
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

// Protected handlers
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, s.cfg)
}

func (s *Server) getAllNodes() []*subscription.Node {
	var nodes []*subscription.Node
	for _, sub := range s.cfg.Subscriptions {
		nodes = append(nodes, sub.Nodes...)
	}
	nodes = append(nodes, s.cfg.StandaloneNodes...)
	return nodes
}

func (s *Server) handleSubscriptions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, s.cfg.Subscriptions)
		return
	}
	if r.Method == http.MethodPost {
		var req struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.URL) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and url required"})
			return
		}
		sub := s.cfg.AddSubscription(req.Name, req.URL)
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, sub)
		return
	}
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (s *Server) handleSubscriptionDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/subscriptions/")
	parts := strings.Split(path, "/")
	name := parts[0]

	if r.Method == http.MethodDelete {
		if !s.cfg.RemoveSubscription(name) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "subscription not found"})
			return
		}
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
		return
	}

	if r.Method == http.MethodPost && len(parts) > 1 && parts[1] == "refresh" {
		sub := s.cfg.FindSubscription(name)
		if sub == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "subscription not found"})
			return
		}
		data, err := subscription.Fetch(sub.URL)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		nodes, err := subscription.Parse(data)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		sub.Nodes = nodes
		sub.LastFetched = time.Now()
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, sub)
		return
	}

	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, s.getAllNodes())
		return
	}
	if r.Method == http.MethodPost {
		var req struct {
			Link string `json:"link"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		node, err := subscription.ParseNode(req.Link)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		s.cfg.AddStandaloneNode(node)
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, node)
		return
	}
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (s *Server) handleNodeRegions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, region.GroupByRegion(s.getAllNodes()))
}

func (s *Server) handleProxyStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "proxy already running"})
		return
	}

	var req struct {
		NodeName  string           `json:"node_name,omitempty"`
		Region    string           `json:"region,omitempty"`
		RouteMode config.RouteMode `json:"route_mode,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
	}

	if req.RouteMode != "" {
		switch req.RouteMode {
		case config.RouteModeGlobal, config.RouteModeWhitelist, config.RouteModeBlacklist:
			s.cfg.RouteMode = req.RouteMode
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid route_mode"})
			return
		}
	}

	var node *subscription.Node
	allNodes := s.getAllNodes()

	if req.NodeName != "" {
		for _, n := range allNodes {
			if n.Name == req.NodeName {
				node = n
				break
			}
		}
		if node == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "node not found"})
			return
		}
	} else {
		var targetNodes []*subscription.Node
		if req.Region != "" {
			groups := region.GroupByRegion(allNodes)
			targetNodes = groups[req.Region]
		} else {
			targetNodes = allNodes
		}
		if len(targetNodes) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no nodes available"})
			return
		}
		var err error
		node, _, err = latency.FindBest(targetNodes)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	var proxy ProxyServer
	var err error
	httpPort := 16708
	socksPort := 16709

	if node.Protocol == "anytls" {
		proxy, err = singbox.Start(node, socksPort, httpPort, s.cfg.RouteMode, s.cfg.Whitelist, s.cfg.Blacklist)
	} else {
		proxy, err = xrayproxy.Start(node, socksPort, httpPort, s.cfg.RouteMode, s.cfg.Whitelist, s.cfg.Blacklist)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	s.proxy = proxy
	s.currentNode = node
	s.isRunning = true
	s.httpPort = httpPort
	s.socksPort = socksPort

	if err := s.cfg.Save(); err != nil {
		log.Printf("failed to save config: %v", err)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "proxy started",
		"node":       node,
		"http_port":  httpPort,
		"socks_port": socksPort,
	})
}

func (s *Server) handleProxyStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.proxy != nil {
		if err := s.proxy.Stop(); err != nil {
			log.Printf("proxy stop error: %v", err)
		}
	}
	s.proxy = nil
	s.currentNode = nil
	s.isRunning = false
	s.httpPort = 0
	s.socksPort = 0
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "proxy stopped",
	})
}

func (s *Server) handleProxyStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"running":    s.isRunning,
		"http_port":  s.httpPort,
		"socks_port": s.socksPort,
		"route_mode": s.cfg.RouteMode,
		"node":       s.currentNode,
	})
}

func (s *Server) handleProxyTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		Region string `json:"region,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
	}

	var targetNodes []*subscription.Node
	if req.Region != "" {
		groups := region.GroupByRegion(s.getAllNodes())
		targetNodes = groups[req.Region]
	} else {
		targetNodes = s.getAllNodes()
	}
	if len(targetNodes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no nodes available"})
		return
	}

	results := latency.TestAll(targetNodes, 5)
	var resp []map[string]interface{}
	for _, res := range results {
		item := map[string]interface{}{
			"name":    res.Node.Name,
			"latency": res.Latency.Milliseconds(),
		}
		if res.Err != nil {
			item["error"] = res.Err.Error()
		}
		resp = append(resp, item)
	}
	writeJSON(w, http.StatusOK, resp)
}
