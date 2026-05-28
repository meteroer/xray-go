package web

import (
	"errors"
	"io"
	"log"
	"net/http"

	"xray-go/config"
	"xray-go/latency"
	"xray-go/region"
	"xray-go/singbox"
	"xray-go/subscription"
	"xray-go/xrayproxy"
)

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
	if err := readJSON(r, &req); err != nil {
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

	s.hub.Broadcast(map[string]interface{}{
		"type":       "proxy_status",
		"running":    true,
		"node":       node,
		"http_port":  httpPort,
		"socks_port": socksPort,
		"route_mode": s.cfg.RouteMode,
	})

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

	s.hub.Broadcast(map[string]interface{}{
		"type":       "proxy_status",
		"running":    false,
		"node":       nil,
		"http_port":  0,
		"socks_port": 0,
		"route_mode": s.cfg.RouteMode,
	})

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
	if err := readJSON(r, &req); err != nil {
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

	s.hub.Broadcast(map[string]interface{}{
		"type":    "latency_progress",
		"status":  "started",
		"total":   len(targetNodes),
	})

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

	s.hub.Broadcast(map[string]interface{}{
		"type":   "latency_progress",
		"status": "completed",
		"count":  len(results),
	})

	writeJSON(w, http.StatusOK, resp)
}
