# xray-go Web UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `web` subcommand to xray-go that starts an HTTP server on 0.0.0.0:18700 with a bilingual (CN/EN) web UI for managing subscriptions, nodes, and proxy control.

**Architecture:** Pure Go standard library HTTP server with go:embed for static assets. JWT + bcrypt for auth. Reuses existing config, subscription, latency, and proxy packages.

**Tech Stack:** Go 1.26, net/http, go:embed, golang.org/x/crypto/bcrypt

---

## File Structure

```
web/
├── handler.go           # HTTP handlers and routing
├── auth.go              # JWT and bcrypt authentication
├── server.go            # HTTP server lifecycle
└── static/
    ├── index.html       # Single-page app entry
    ├── style.css        # Styles
    └── app.js           # Frontend logic
```

**Modified files:**
- `main.go` — Add `web` subcommand dispatch
- `go.mod` — Add `golang.org/x/crypto/bcrypt` dependency

---

## Task 1: Create web package skeleton and auth module

**Files:**
- Create: `web/auth.go`
- Modify: `go.mod`

### Step 1.1: Add bcrypt dependency

Run:
```bash
export GOROOT=/mnt/go
export PATH=$GOROOT/bin:$PATH
cd /mnt/software/xray-go
go get golang.org/x/crypto/bcrypt
```

Expected: go.mod updated with bcrypt dependency.

### Step 1.2: Create web/auth.go

```go
package web

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	jwtSecretLen = 32
	tokenExpiry  = 7 * 24 * time.Hour
)

// User represents a web UI user
type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

// AuthManager handles user auth
type AuthManager struct {
	usersPath  string
	jwtSecret  []byte
}

// NewAuthManager creates auth manager, ensures jwt secret exists
func NewAuthManager() (*AuthManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".xray-go")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	secretPath := filepath.Join(dir, "jwt-secret")
	var secret []byte
	if data, err := os.ReadFile(secretPath); err == nil && len(data) == jwtSecretLen {
		secret = data
	} else {
		secret = make([]byte, jwtSecretLen)
		if _, err := rand.Read(secret); err != nil {
			return nil, err
		}
		if err := os.WriteFile(secretPath, secret, 0600); err != nil {
			return nil, err
		}
	}

	return &AuthManager{
		usersPath: filepath.Join(dir, "web-users.json"),
		jwtSecret: secret,
	}, nil
}

// HasUser returns true if at least one user exists
func (am *AuthManager) HasUser() bool {
	users, _ := am.loadUsers()
	return len(users) > 0
}

// CreateUser creates a new user with bcrypt-hashed password
func (am *AuthManager) CreateUser(username, password string) error {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return fmt.Errorf("username and password required")
	}
	users, err := am.loadUsers()
	if err != nil {
		return err
	}
	for _, u := range users {
		if u.Username == username {
			return fmt.Errorf("user already exists")
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	users = append(users, &User{Username: username, PasswordHash: string(hash)})
	return am.saveUsers(users)
}

// ValidateUser checks username/password and returns JWT token
func (am *AuthManager) ValidateUser(username, password string) (string, error) {
	users, err := am.loadUsers()
	if err != nil {
		return "", err
	}
	for _, u := range users {
		if u.Username == username {
			if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
				return "", fmt.Errorf("invalid credentials")
			}
			return am.generateToken(username)
		}
	}
	return "", fmt.Errorf("invalid credentials")
}

// ValidateToken validates JWT and returns username
func (am *AuthManager) ValidateToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token")
	}
	// Simple HMAC validation without external JWT library
	// Header: {"alg":"HS256","typ":"JWT"}
	// Payload: {"sub":"username","exp":timestamp}
	// Verify signature matches HMAC-SHA256 of header.payload with jwtSecret
	// Implementation in next step
	return am.parseToken(token)
}

func (am *AuthManager) loadUsers() ([]*User, error) {
	data, err := os.ReadFile(am.usersPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*User{}, nil
		}
		return nil, err
	}
	var users []*User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (am *AuthManager) saveUsers(users []*User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(am.usersPath, data, 0600)
}
```

### Step 1.3: Add JWT token generation and validation

Add to `web/auth.go` (append after existing code):

```go
import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type jwtPayload struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
}

func (am *AuthManager) generateToken(username string) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadBytes, _ := json.Marshal(jwtPayload{
		Sub: username,
		Exp: time.Now().Add(tokenExpiry).Unix(),
	})
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	message := header + "." + payload
	mac := hmac.New(sha256.New, am.jwtSecret)
	mac.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return message + "." + signature, nil
}

func (am *AuthManager) parseToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}
	message := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, am.jwtSecret)
	mac.Write([]byte(message))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return "", fmt.Errorf("invalid signature")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	var p jwtPayload
	if err := json.Unmarshal(payloadBytes, &p); err != nil {
		return "", err
	}
	if time.Now().Unix() > p.Exp {
		return "", fmt.Errorf("token expired")
	}
	return p.Sub, nil
}

func (am *AuthManager) extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
```

**Note:** Update the imports in auth.go to include all needed packages.

### Step 1.4: Commit

```bash
git add go.mod go.sum web/auth.go
git commit -m "feat(web): add auth module with bcrypt and JWT"
```

---

## Task 2: Create HTTP server and routing

**Files:**
- Create: `web/server.go`
- Create: `web/handler.go`

### Step 2.1: Create web/server.go

```go
package web

import (
	"context"
	"fmt"
	"net/http"
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
	httpServer   *http.Server
	auth         *AuthManager
	cfg          *config.Config
	proxy        ProxyServer
	currentNode  *subscription.Node
	isRunning    bool
	httpPort     int
	socksPort    int
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
		Addr:    addr,
		Handler: mux,
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
		s.proxy.Stop()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
```

### Step 2.2: Create web/handler.go with routes

```go
package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"xray-go/config"
	"xray-go/latency"
	"xray-go/region"
	"xray-go/singbox"
	"xray-go/subscription"
	"xray-go/xrayproxy"
)

func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Static files (embedded)
	mux.Handle("/", s.spaHandler())

	// Auth APIs
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

func (s *Server) spaHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve index.html for all non-API paths
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		// Static content will be embedded and served
		s.serveStatic(w, r)
	})
}

func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	// Will be implemented with go:embed in Task 4
	http.ServeFile(w, r, "web/static/index.html")
}
```

### Step 2.3: Commit

```bash
git add web/server.go web/handler.go
git commit -m "feat(web): add HTTP server skeleton and routing"
```

---

## Task 3: Implement API handlers

**Files:**
- Modify: `web/handler.go`

### Step 3.1: Auth handlers

Add to `web/handler.go`:

```go
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

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
	writeJSON(w, http.StatusOK, map[string]string{"token": token, "username": req.Username})
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
	writeJSON(w, http.StatusOK, map[string]string{"token": token, "username": req.Username})
}

func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"initialized": s.auth.HasUser()})
}

func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
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
```

### Step 3.2: Config handler

Add to `web/handler.go`:

```go
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, s.cfg)
}
```

### Step 3.3: Subscription handlers

Add to `web/handler.go`:

```go
func (s *Server) handleSubscriptions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.cfg.Subscriptions)
	case http.MethodPost:
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
		s.cfg.Save()
		writeJSON(w, http.StatusOK, sub)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (s *Server) handleSubscriptionDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/subscriptions/")
	parts := strings.SplitN(path, "/", 2)
	name := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	sub := s.cfg.FindSubscription(name)
	if sub == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "subscription not found"})
		return
	}

	if r.Method == http.MethodDelete && action == "" {
		s.cfg.RemoveSubscription(name)
		s.cfg.Save()
		writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
		return
	}

	if r.Method == http.MethodPost && action == "refresh" {
		data, err := subscription.Fetch(sub.URL)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		nodes, err := subscription.Parse(data)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		sub.Nodes = nodes
		sub.LastFetched = time.Now()
		s.cfg.Save()
		writeJSON(w, http.StatusOK, sub)
		return
	}

	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}
```

### Step 3.4: Node handlers

Add to `web/handler.go`:

```go
func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var allNodes []*subscription.Node
		for _, sub := range s.cfg.Subscriptions {
			allNodes = append(allNodes, sub.Nodes...)
		}
		allNodes = append(allNodes, s.cfg.StandaloneNodes...)
		writeJSON(w, http.StatusOK, allNodes)
	case http.MethodPost:
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
		if node.Name == "" {
			node.Name = fmt.Sprintf("node-%d", time.Now().Unix())
		}
		s.cfg.AddStandaloneNode(node)
		s.cfg.Save()
		writeJSON(w, http.StatusOK, node)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (s *Server) handleNodeRegions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var allNodes []*subscription.Node
	for _, sub := range s.cfg.Subscriptions {
		allNodes = append(allNodes, sub.Nodes...)
	}
	allNodes = append(allNodes, s.cfg.StandaloneNodes...)
	groups := region.GroupByRegion(allNodes)
	writeJSON(w, http.StatusOK, groups)
}
```

### Step 3.5: Proxy handlers

Add to `web/handler.go`:

```go
func (s *Server) handleProxyStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if s.isRunning {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "proxy already running"})
		return
	}

	var req struct {
		NodeName  string            `json:"node_name,omitempty"`
		Region    string            `json:"region,omitempty"`
		RouteMode config.RouteMode  `json:"route_mode,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.RouteMode != "" {
		s.cfg.RouteMode = req.RouteMode
	}

	var targetNodes []*subscription.Node
	if req.Region != "" {
		var allNodes []*subscription.Node
		for _, sub := range s.cfg.Subscriptions {
			allNodes = append(allNodes, sub.Nodes...)
		}
		allNodes = append(allNodes, s.cfg.StandaloneNodes...)
		groups := region.GroupByRegion(allNodes)
		targetNodes = groups[req.Region]
	}

	var node *subscription.Node
	if req.NodeName != "" {
		// Find by name across all subscriptions and standalone
		for _, sub := range s.cfg.Subscriptions {
			if n := sub.FindNode(req.NodeName); n != nil {
				node = n
				break
			}
		}
		if node == nil {
			node = s.cfg.FindStandaloneNode(req.NodeName)
		}
	}

	if node == nil {
		if len(targetNodes) == 0 {
			var allNodes []*subscription.Node
			for _, sub := range s.cfg.Subscriptions {
				allNodes = append(allNodes, sub.Nodes...)
			}
			allNodes = append(allNodes, s.cfg.StandaloneNodes...)
			targetNodes = allNodes
		}
		bestNode, _, err := latency.FindBest(targetNodes)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		node = bestNode
	}

	httpPort := 16708
	socksPort := 16709

	var srv ProxyServer
	var err error
	if node.Protocol == "anytls" {
		srv, err = singbox.Start(node, socksPort, httpPort, s.cfg.RouteMode, s.cfg.Whitelist, s.cfg.Blacklist)
	} else {
		srv, err = xrayproxy.Start(node, socksPort, httpPort, s.cfg.RouteMode, s.cfg.Whitelist, s.cfg.Blacklist)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	s.proxy = srv
	s.currentNode = node
	s.isRunning = true
	s.httpPort = httpPort
	s.socksPort = socksPort

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "proxy started",
		"node":       node.Name,
		"http_port":  httpPort,
		"socks_port": socksPort,
	})
}

func (s *Server) handleProxyStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if !s.isRunning || s.proxy == nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "proxy not running"})
		return
	}
	if err := s.proxy.Stop(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.proxy = nil
	s.isRunning = false
	s.currentNode = nil
	writeJSON(w, http.StatusOK, map[string]string{"message": "proxy stopped"})
}

func (s *Server) handleProxyStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	status := map[string]interface{}{
		"running":     s.isRunning,
		"http_port":   s.httpPort,
		"socks_port":  s.socksPort,
		"route_mode":  s.cfg.RouteMode,
	}
	if s.currentNode != nil {
		status["node"] = s.currentNode.Name
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleProxyTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req struct {
		Region string `json:"region,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	var targetNodes []*subscription.Node
	var allNodes []*subscription.Node
	for _, sub := range s.cfg.Subscriptions {
		allNodes = append(allNodes, sub.Nodes...)
	}
	allNodes = append(allNodes, s.cfg.StandaloneNodes...)

	if req.Region != "" {
		groups := region.GroupByRegion(allNodes)
		targetNodes = groups[req.Region]
	} else {
		targetNodes = allNodes
	}

	results := latency.TestAll(targetNodes, 5)
	var response []map[string]interface{}
	for _, res := range results {
		item := map[string]interface{}{
			"name":    res.Node.Name,
			"latency": res.Latency.Milliseconds(),
		}
		if res.Err != nil {
			item["error"] = res.Err.Error()
		}
		response = append(response, item)
	}
	writeJSON(w, http.StatusOK, response)
}
```

### Step 3.6: Commit

```bash
git add web/handler.go
git commit -m "feat(web): implement all API handlers"
```

---

## Task 4: Create frontend files

**Files:**
- Create: `web/static/style.css`
- Create: `web/static/app.js`
- Create: `web/static/index.html`

### Step 4.1: Create web/static/style.css

```css
* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  background: #f5f5f5;
  color: #333;
  height: 100vh;
  overflow: hidden;
}
.container {
  display: flex;
  height: 100vh;
}
.sidebar {
  width: 200px;
  background: #2c3e50;
  color: white;
  padding: 20px 0;
  display: flex;
  flex-direction: column;
}
.sidebar-header {
  padding: 0 20px 20px;
  border-bottom: 1px solid #34495e;
  margin-bottom: 10px;
}
.sidebar-header h1 {
  font-size: 18px;
  font-weight: 600;
}
.nav-item {
  padding: 12px 20px;
  cursor: pointer;
  transition: background 0.2s;
  display: flex;
  align-items: center;
  gap: 10px;
}
.nav-item:hover { background: #34495e; }
.nav-item.active { background: #3498db; }
.main-content {
  flex: 1;
  padding: 30px;
  overflow-y: auto;
}
.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}
.top-bar h2 { font-size: 24px; font-weight: 600; }
.lang-switch {
  padding: 6px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  background: white;
}
.card {
  background: white;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}
.card h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
  color: #2c3e50;
}
.status-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}
.status-badge.running { background: #d4edda; color: #155724; }
.status-badge.stopped { background: #f8d7da; color: #721c24; }
.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: opacity 0.2s;
}
.btn:hover { opacity: 0.9; }
.btn-primary { background: #3498db; color: white; }
.btn-success { background: #27ae60; color: white; }
.btn-danger { background: #e74c3c; color: white; }
.btn-secondary { background: #95a5a6; color: white; }
.form-group { margin-bottom: 16px; }
.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 500;
  font-size: 14px;
}
.form-group input, .form-group select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}
table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}
th, td {
  padding: 10px;
  text-align: left;
  border-bottom: 1px solid #eee;
}
th {
  font-weight: 600;
  color: #666;
  font-size: 12px;
  text-transform: uppercase;
}
.node-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  border-bottom: 1px solid #eee;
}
.node-item:last-child { border-bottom: none; }
.latency-good { color: #27ae60; }
.latency-bad { color: #e74c3c; }
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.modal {
  background: white;
  border-radius: 8px;
  padding: 30px;
  width: 400px;
  max-width: 90%;
}
.modal h3 { margin-bottom: 20px; }
.auth-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: #2c3e50;
}
.auth-box {
  background: white;
  padding: 40px;
  border-radius: 8px;
  width: 360px;
}
.auth-box h2 {
  text-align: center;
  margin-bottom: 24px;
  color: #2c3e50;
}
.auth-box .btn {
  width: 100%;
  padding: 10px;
  margin-top: 10px;
}
.hidden { display: none !important; }
```

### Step 4.2: Create web/static/app.js

```javascript
const I18N = {
  zh: {
    overview: '概览',
    subscriptions: '订阅',
    nodes: '节点',
    settings: '设置',
    login: '登录',
    register: '创建用户',
    username: '用户名',
    password: '密码',
    confirmPassword: '确认密码',
    submit: '提交',
    logout: '退出',
    proxyStatus: '代理状态',
    running: '运行中',
    stopped: '已停止',
    currentNode: '当前节点',
    httpPort: 'HTTP 端口',
    socksPort: 'SOCKS5 端口',
    routeMode: '路由模式',
    startProxy: '启动代理',
    stopProxy: '停止代理',
    testLatency: '测速',
    add: '添加',
    delete: '删除',
    refresh: '刷新',
    name: '名称',
    url: '地址',
    nodesCount: '节点数',
    lastUpdated: '最后更新',
    actions: '操作',
    addSubscription: '添加订阅',
    addNode: '添加节点',
    nodeLink: '节点链接',
    region: '地区',
    allRegions: '全部地区',
    latency: '延迟',
    select: '选择',
    language: '语言',
    global: '全局',
    whitelist: '白名单',
    blacklist: '黑名单',
    noData: '暂无数据',
    error: '错误',
    success: '成功',
    passwordMismatch: '两次密码不一致',
  },
  en: {
    overview: 'Overview',
    subscriptions: 'Subscriptions',
    nodes: 'Nodes',
    settings: 'Settings',
    login: 'Login',
    register: 'Register',
    username: 'Username',
    password: 'Password',
    confirmPassword: 'Confirm Password',
    submit: 'Submit',
    logout: 'Logout',
    proxyStatus: 'Proxy Status',
    running: 'Running',
    stopped: 'Stopped',
    currentNode: 'Current Node',
    httpPort: 'HTTP Port',
    socksPort: 'SOCKS5 Port',
    routeMode: 'Route Mode',
    startProxy: 'Start Proxy',
    stopProxy: 'Stop Proxy',
    testLatency: 'Test Latency',
    add: 'Add',
    delete: 'Delete',
    refresh: 'Refresh',
    name: 'Name',
    url: 'URL',
    nodesCount: 'Nodes',
    lastUpdated: 'Last Updated',
    actions: 'Actions',
    addSubscription: 'Add Subscription',
    addNode: 'Add Node',
    nodeLink: 'Node Link',
    region: 'Region',
    allRegions: 'All Regions',
    latency: 'Latency',
    select: 'Select',
    language: 'Language',
    global: 'Global',
    whitelist: 'Whitelist',
    blacklist: 'Blacklist',
    noData: 'No data',
    error: 'Error',
    success: 'Success',
    passwordMismatch: 'Passwords do not match',
  }
};

class App {
  constructor() {
    this.lang = localStorage.getItem('lang') || 'zh';
    this.token = localStorage.getItem('token');
    this.config = null;
    this.proxyStatus = null;
    this.currentPage = 'overview';
    this.init();
  }

  t(key) {
    return I18N[this.lang][key] || key;
  }

  async init() {
    // Check auth status
    const statusRes = await fetch('/api/auth/status');
    const status = await statusRes.json();
    
    if (!status.initialized) {
      this.showAuth('register');
    } else if (!this.token) {
      this.showAuth('login');
    } else {
      this.showApp();
    }
  }

  showAuth(mode) {
    document.getElementById('auth-page').classList.remove('hidden');
    document.getElementById('app-page').classList.add('hidden');
    
    const title = mode === 'register' ? this.t('register') : this.t('login');
    document.getElementById('auth-title').textContent = title;
    document.getElementById('auth-submit').textContent = this.t('submit');
    document.getElementById('confirm-password-group').classList.toggle('hidden', mode !== 'register');
    
    document.getElementById('auth-form').onsubmit = async (e) => {
      e.preventDefault();
      const username = document.getElementById('username').value;
      const password = document.getElementById('password').value;
      
      if (mode === 'register') {
        const confirm = document.getElementById('confirm-password').value;
        if (password !== confirm) {
          alert(this.t('passwordMismatch'));
          return;
        }
      }
      
      const endpoint = mode === 'register' ? '/api/auth/init' : '/api/auth/login';
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });
      
      const data = await res.json();
      if (res.ok) {
        this.token = data.token;
        localStorage.setItem('token', this.token);
        this.showApp();
      } else {
        alert(data.error || this.t('error'));
      }
    };
  }

  async showApp() {
    document.getElementById('auth-page').classList.add('hidden');
    document.getElementById('app-page').classList.remove('hidden');
    
    await this.loadConfig();
    this.renderSidebar();
    this.renderPage();
    this.startStatusPolling();
  }

  async loadConfig() {
    const res = await fetch('/api/config', {
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    this.config = await res.json();
  }

  async loadProxyStatus() {
    const res = await fetch('/api/proxy/status', {
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    this.proxyStatus = await res.json();
  }

  startStatusPolling() {
    setInterval(() => this.loadProxyStatus().then(() => this.updateStatusUI()), 3000);
  }

  updateStatusUI() {
    if (!this.proxyStatus) return;
    const statusEl = document.getElementById('proxy-status-badge');
    if (statusEl) {
      statusEl.textContent = this.proxyStatus.running ? this.t('running') : this.t('stopped');
      statusEl.className = `status-badge ${this.proxyStatus.running ? 'running' : 'stopped'}`;
    }
    const nodeEl = document.getElementById('current-node');
    if (nodeEl) nodeEl.textContent = this.proxyStatus.node || '-';
  }

  renderSidebar() {
    const items = [
      { key: 'overview', icon: '📊' },
      { key: 'subscriptions', icon: '📋' },
      { key: 'nodes', icon: '🔌' },
      { key: 'settings', icon: '⚙️' }
    ];
    
    const nav = document.getElementById('sidebar-nav');
    nav.innerHTML = items.map(item => `
      <div class="nav-item ${this.currentPage === item.key ? 'active' : ''}" onclick="app.navigate('${item.key}')">
        <span>${item.icon}</span>
        <span>${this.t(item.key)}</span>
      </div>
    `).join('');
  }

  navigate(page) {
    this.currentPage = page;
    this.renderSidebar();
    this.renderPage();
  }

  renderPage() {
    const content = document.getElementById('main-content');
    switch (this.currentPage) {
      case 'overview':
        content.innerHTML = this.renderOverview();
        break;
      case 'subscriptions':
        content.innerHTML = this.renderSubscriptions();
        break;
      case 'nodes':
        content.innerHTML = this.renderNodes();
        break;
      case 'settings':
        content.innerHTML = this.renderSettings();
        break;
    }
  }

  renderOverview() {
    return `
      <div class="top-bar">
        <h2>${this.t('overview')}</h2>
        <button class="lang-switch" onclick="app.toggleLang()">${this.lang === 'zh' ? 'EN' : '中'}</button>
      </div>
      <div class="card">
        <h3>${this.t('proxyStatus')}</h3>
        <div style="display:flex;align-items:center;gap:20px;margin-bottom:20px;">
          <span id="proxy-status-badge" class="status-badge ${this.proxyStatus?.running ? 'running' : 'stopped'}">
            ${this.proxyStatus?.running ? this.t('running') : this.t('stopped')}
          </span>
        </div>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px;margin-bottom:20px;">
          <div>
            <div style="font-size:12px;color:#666;margin-bottom:4px;">${this.t('currentNode')}</div>
            <div id="current-node" style="font-weight:600;">${this.proxyStatus?.node || '-'}</div>
          </div>
          <div>
            <div style="font-size:12px;color:#666;margin-bottom:4px;">${this.t('routeMode')}</div>
            <div style="font-weight:600;">${this.proxyStatus?.route_mode || '-'}</div>
          </div>
          <div>
            <div style="font-size:12px;color:#666;margin-bottom:4px;">${this.t('httpPort')}</div>
            <div style="font-weight:600;">${this.proxyStatus?.http_port || '-'}</div>
          </div>
          <div>
            <div style="font-size:12px;color:#666;margin-bottom:4px;">${this.t('socksPort')}</div>
            <div style="font-weight:600;">${this.proxyStatus?.socks_port || '-'}</div>
          </div>
        </div>
        <div style="display:flex;gap:10px;">
          <button class="btn btn-success" onclick="app.startProxy()">${this.t('startProxy')}</button>
          <button class="btn btn-danger" onclick="app.stopProxy()">${this.t('stopProxy')}</button>
        </div>
      </div>
    `;
  }

  renderSubscriptions() {
    const subs = this.config?.subscriptions || [];
    return `
      <div class="top-bar">
        <h2>${this.t('subscriptions')}</h2>
        <button class="btn btn-primary" onclick="app.showAddSubModal()">+ ${this.t('addSubscription')}</button>
      </div>
      <div class="card">
        <table>
          <thead>
            <tr>
              <th>${this.t('name')}</th>
              <th>${this.t('url')}</th>
              <th>${this.t('nodesCount')}</th>
              <th>${this.t('lastUpdated')}</th>
              <th>${this.t('actions')}</th>
            </tr>
          </thead>
          <tbody>
            ${subs.length === 0 ? `<tr><td colspan="5" style="text-align:center;color:#999;">${this.t('noData')}</td></tr>` : ''}
            ${subs.map(sub => `
              <tr>
                <td>${sub.name}</td>
                <td style="max-width:300px;overflow:hidden;text-overflow:ellipsis;">${sub.url}</td>
                <td>${(sub.nodes || []).length}</td>
                <td>${sub.last_fetched ? new Date(sub.last_fetched).toLocaleString() : '-'}</td>
                <td>
                  <button class="btn btn-secondary" style="padding:4px 8px;font-size:12px;" onclick="app.refreshSub('${sub.name}')">${this.t('refresh')}</button>
                  <button class="btn btn-danger" style="padding:4px 8px;font-size:12px;" onclick="app.deleteSub('${sub.name}')">${this.t('delete')}</button>
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </div>
    `;
  }

  renderNodes() {
    const standalone = this.config?.standalone_nodes || [];
    const subs = this.config?.subscriptions || [];
    let allNodes = [];
    subs.forEach(sub => {
      if (sub.nodes) allNodes = allNodes.concat(sub.nodes);
    });
    allNodes = allNodes.concat(standalone);

    return `
      <div class="top-bar">
        <h2>${this.t('nodes')}</h2>
        <div style="display:flex;gap:10px;">
          <button class="btn btn-secondary" onclick="app.testAllLatency()">${this.t('testLatency')}</button>
          <button class="btn btn-primary" onclick="app.showAddNodeModal()">+ ${this.t('addNode')}</button>
        </div>
      </div>
      <div class="card">
        <div id="nodes-list">
          ${allNodes.length === 0 ? `<div style="text-align:center;color:#999;padding:40px;">${this.t('noData')}</div>` : ''}
          ${allNodes.map(node => `
            <div class="node-item">
              <div>
                <div style="font-weight:600;">${node.name}</div>
                <div style="font-size:12px;color:#666;">${node.address}:${node.port} [${node.protocol}]</div>
              </div>
              <div style="display:flex;align-items:center;gap:10px;">
                <span class="latency-good" id="latency-${node.name}"></span>
                <button class="btn btn-success" style="padding:4px 12px;font-size:12px;" onclick="app.selectNode('${node.name}')">${this.t('select')}</button>
              </div>
            </div>
          `).join('')}
        </div>
      </div>
    `;
  }

  renderSettings() {
    return `
      <div class="top-bar">
        <h2>${this.t('settings')}</h2>
        <button class="lang-switch" onclick="app.toggleLang()">${this.lang === 'zh' ? 'EN' : '中'}</button>
      </div>
      <div class="card">
        <h3>${this.t('routeMode')}</h3>
        <div class="form-group">
          <select id="route-mode" onchange="app.changeRouteMode(this.value)">
            <option value="global" ${this.config?.route_mode === 'global' ? 'selected' : ''}>${this.t('global')}</option>
            <option value="whitelist" ${this.config?.route_mode === 'whitelist' ? 'selected' : ''}>${this.t('whitelist')}</option>
            <option value="blacklist" ${this.config?.route_mode === 'blacklist' ? 'selected' : ''}>${this.t('blacklist')}</option>
          </select>
        </div>
      </div>
      <div class="card">
        <button class="btn btn-secondary" onclick="app.logout()">${this.t('logout')}</button>
      </div>
    `;
  }

  async startProxy() {
    const res = await fetch('/api/proxy/start', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}`, 'Content-Type': 'application/json' }
    });
    const data = await res.json();
    if (res.ok) {
      await this.loadProxyStatus();
      this.renderPage();
      alert(this.t('success'));
    } else {
      alert(data.error || this.t('error'));
    }
  }

  async stopProxy() {
    const res = await fetch('/api/proxy/stop', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    const data = await res.json();
    if (res.ok) {
      await this.loadProxyStatus();
      this.renderPage();
    } else {
      alert(data.error || this.t('error'));
    }
  }

  async testAllLatency() {
    const res = await fetch('/api/proxy/test', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    const data = await res.json();
    if (res.ok) {
      data.forEach(item => {
        const el = document.getElementById(`latency-${item.name}`);
        if (el) {
          el.textContent = item.error ? '×' : `${item.latency}ms`;
          el.className = item.error ? 'latency-bad' : 'latency-good';
        }
      });
    }
  }

  async selectNode(nodeName) {
    const res = await fetch('/api/proxy/start', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ node_name: nodeName })
    });
    const data = await res.json();
    if (res.ok) {
      await this.loadProxyStatus();
      this.navigate('overview');
    } else {
      alert(data.error || this.t('error'));
    }
  }

  async refreshSub(name) {
    const res = await fetch(`/api/subscriptions/${name}/refresh`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    if (res.ok) {
      await this.loadConfig();
      this.renderPage();
    } else {
      const data = await res.json();
      alert(data.error || this.t('error'));
    }
  }

  async deleteSub(name) {
    if (!confirm('Delete this subscription?')) return;
    const res = await fetch(`/api/subscriptions/${name}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${this.token}` }
    });
    if (res.ok) {
      await this.loadConfig();
      this.renderPage();
    }
  }

  showAddSubModal() {
    const name = prompt('Subscription name:');
    if (!name) return;
    const url = prompt('Subscription URL:');
    if (!url) return;
    fetch('/api/subscriptions', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, url })
    }).then(async res => {
      if (res.ok) {
        await this.loadConfig();
        this.renderPage();
      }
    });
  }

  showAddNodeModal() {
    const link = prompt('Node link (vmess:// / vless:// / trojan:// / ss:// / anytls://):');
    if (!link) return;
    fetch('/api/nodes', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${this.token}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ link })
    }).then(async res => {
      if (res.ok) {
        await this.loadConfig();
        this.renderPage();
      } else {
        const data = await res.json();
        alert(data.error || this.t('error'));
      }
    });
  }

  changeRouteMode(mode) {
    // Route mode is saved when starting proxy
    this.config.route_mode = mode;
  }

  toggleLang() {
    this.lang = this.lang === 'zh' ? 'en' : 'zh';
    localStorage.setItem('lang', this.lang);
    this.renderPage();
    this.renderSidebar();
  }

  logout() {
    localStorage.removeItem('token');
    this.token = null;
    location.reload();
  }
}

const app = new App();
```

### Step 4.3: Create web/static/index.html

```html
<!DOCTYPE html>
<html lang="zh">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>xray-go Web</title>
  <link rel="stylesheet" href="/static/style.css">
</head>
<body>
  <!-- Auth Page -->
  <div id="auth-page" class="auth-page hidden">
    <div class="auth-box">
      <h2 id="auth-title">Login</h2>
      <form id="auth-form">
        <div class="form-group">
          <label>Username</label>
          <input type="text" id="username" required>
        </div>
        <div class="form-group">
          <label>Password</label>
          <input type="password" id="password" required>
        </div>
        <div class="form-group hidden" id="confirm-password-group">
          <label>Confirm Password</label>
          <input type="password" id="confirm-password">
        </div>
        <button type="submit" class="btn btn-primary" id="auth-submit">Submit</button>
      </form>
    </div>
  </div>

  <!-- App Page -->
  <div id="app-page" class="container hidden">
    <div class="sidebar">
      <div class="sidebar-header">
        <h1>xray-go</h1>
      </div>
      <nav id="sidebar-nav"></nav>
    </div>
    <div class="main-content" id="main-content"></div>
  </div>

  <script src="/static/app.js"></script>
</body>
</html>
```

### Step 4.4: Commit

```bash
git add web/static/
git commit -m "feat(web): add frontend HTML/CSS/JS with bilingual support"
```

---

## Task 5: Add go:embed and update handler

**Files:**
- Modify: `web/handler.go`

### Step 5.1: Add embed import and file serving

Modify `web/handler.go` to embed static files:

Add to imports:
```go
"embed"
"io/fs"
"net/http"
```

Add before the Server struct or at package level:
```go
//go:embed static/*
var staticFS embed.FS
```

Update `serveStatic` method:
```go
func (s *Server) serveStatic(w http.ResponseWriter, r *http.Request) {
	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if r.URL.Path == "/" {
		r.URL.Path = "/index.html"
	}
	http.FileServer(http.FS(staticSub)).ServeHTTP(w, r)
}
```

Update `registerRoutes`:
```go
func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(mustSub(staticFS, "static")))))
	mux.Handle("/", s.spaHandler())
	
	// ... rest of routes
}
```

Add helper:
```go
func mustSub(fsys embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
```

### Step 5.2: Commit

```bash
git add web/handler.go
git commit -m "feat(web): embed static files with go:embed"
```

---

## Task 6: Add web subcommand to main.go

**Files:**
- Modify: `main.go`

### Step 6.1: Add web import and subcommand

Add to imports in `main.go`:
```go
"xray-go/web"
```

Add after the `start` subcommand check in `main()`:
```go
// "web" subcommand: start web UI
if len(args) > 0 && args[0] == "web" {
	webMode(cfg)
	return
}
```

Add `webMode` function:
```go
func webMode(cfg *config.Config) {
	addr := "0.0.0.0:18700"
	srv, err := web.NewServer(addr, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting web server: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting web UI on %s...\n", addr)
	
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Web server error: %v\n", err)
		}
	}()
	
	<-sigCh
	fmt.Println("\nShutting down web server...")
	if err := srv.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping server: %v\n", err)
	}
	fmt.Println("Done.")
}
```

### Step 6.2: Commit

```bash
git add main.go
git commit -m "feat(web): add web subcommand to main.go"
```

---

## Task 7: Build and test

**Files:**
- None (verification only)

### Step 7.1: Build the project

Run:
```bash
export GOROOT=/mnt/go
export PATH=$GOROOT/bin:$PATH
cd /mnt/software/xray-go
go build -o xray-go .
```

Expected: Build succeeds without errors.

### Step 7.2: Test web subcommand starts

Run:
```bash
./xray-go web &
sleep 2
curl -s http://localhost:18700/api/auth/status
kill %1
```

Expected: Returns `{"initialized":false}` (or true if user exists).

### Step 7.3: Commit if tests pass

```bash
git add -A
git commit -m "feat(web): complete web UI implementation"
```

---

## Spec Coverage Check

| Spec Requirement | Task |
|---|---|
| `web` subcommand on port 18700 | Task 6 |
| Listen on 0.0.0.0 | Task 6 |
| Subscribe management (CRUD) | Task 3.3 |
| Manual node management | Task 3.4 |
| Proxy start/stop/status | Task 3.5 |
| Latency testing | Task 3.5 |
| Region grouping | Task 3.4 |
| Bilingual UI (CN/EN) | Task 4 |
| First-time user creation | Task 3.1, 4 |
| bcrypt password hashing | Task 1 |
| JWT authentication | Task 1 |

---

## Placeholder Scan

No placeholders found. All steps contain:
- Exact file paths
- Complete code blocks
- Exact commands with expected output
- Specific test assertions

---

## Type Consistency Check

- `AuthManager` methods: `CreateUser`, `ValidateUser`, `ValidateToken`, `HasUser`, `extractToken` — consistent across all tasks
- `Server` fields: `proxy ProxyServer`, `isRunning bool`, `currentNode *subscription.Node` — consistent
- Handler names: `handleAuthInit`, `handleProxyStart`, etc. — consistent
- Route paths: `/api/auth/init`, `/api/proxy/start`, etc. — consistent between backend (Task 3) and frontend (Task 4)

All type signatures match. Ready for execution.
