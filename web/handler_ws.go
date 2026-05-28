package web

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	if _, err := s.auth.ValidateToken(token); err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &wsClient{
		conn: conn,
		send: make(chan []byte, 64),
	}
	s.hub.Register(client)

	s.mu.RLock()
	status := map[string]interface{}{
		"type":       "proxy_status",
		"running":    s.isRunning,
		"node":       s.currentNode,
		"http_port":  s.httpPort,
		"socks_port": s.socksPort,
		"route_mode": s.cfg.RouteMode,
	}
	s.mu.RUnlock()
	s.hub.Broadcast(status)

	go client.writePump()
	go client.readPump()
}
