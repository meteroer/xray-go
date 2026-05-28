package web

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

type wsHub struct {
	clients map[*wsClient]struct{}
	mu      sync.RWMutex
}

func newWsHub() *wsHub {
	return &wsHub{
		clients: make(map[*wsClient]struct{}),
	}
}

func (h *wsHub) Register(client *wsClient) {
	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()
	log.Printf("WebSocket client connected, total: %d", len(h.clients))
}

func (h *wsHub) Unregister(client *wsClient) {
	h.mu.Lock()
	delete(h.clients, client)
	h.mu.Unlock()
	close(client.send)
	log.Printf("WebSocket client disconnected, total: %d", len(h.clients))
}

func (h *wsHub) Broadcast(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws broadcast marshal error: %v", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			go h.Unregister(client)
		}
	}
}

func (c *wsClient) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *wsClient) readPump() {
	defer c.conn.Close()
	c.conn.SetReadLimit(512)
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}
