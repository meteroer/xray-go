package web

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"xray-go/config"
	"xray-go/region"
	"xray-go/singbox"
	"xray-go/subscription"
	"xray-go/xrayproxy"
)

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
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.URL) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and url required"})
			return
		}
		sub := s.cfg.AddSubscription(req.Name, req.URL)
		for _, node := range sub.Nodes {
			if node.Region == "" {
				node.Region = region.DetectRegion(node)
			}
		}
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
	rawPath := r.URL.RawPath
	if rawPath == "" {
		rawPath = r.URL.Path
	}
	rawPath = strings.TrimPrefix(rawPath, "/api/subscriptions/")
	parts := strings.Split(rawPath, "/")
	rawName := parts[0]
	if r.URL.RawQuery != "" {
		rawName = rawName + "?" + r.URL.RawQuery
	}
	name, err := url.PathUnescape(rawName)
	if err != nil {
		name = rawName
	}

	if r.Method == http.MethodGet {
		sub := s.cfg.FindSubscription(name)
		if sub == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "subscription not found"})
			return
		}
		writeJSON(w, http.StatusOK, sub)
		return
	}

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

	if r.Method == http.MethodPut {
		sub := s.cfg.FindSubscription(name)
		if sub == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "subscription not found"})
			return
		}
		var req struct {
			URL string `json:"url"`
		}
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		if strings.TrimSpace(req.URL) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url required"})
			return
		}
		sub.URL = strings.TrimSpace(req.URL)
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, sub)
		return
	}

	if r.Method == http.MethodPost && len(parts) > 1 && parts[1] == "refresh" {
		sub := s.cfg.FindSubscription(name)
		if sub == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "subscription not found"})
			return
		}
		oldNames := make(map[string]bool)
		for _, n := range sub.Nodes {
			oldNames[n.Name] = true
		}
		data, err := subscription.Fetch(sub.URL)
		if err != nil {
			log.Printf("Direct fetch failed for '%s': %v, trying fallback...", sub.Name, err)
			data, err = s.fetchWithFallback(sub)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		}
		nodes, err := subscription.Parse(data)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		newNames := make(map[string]bool)
		for _, n := range nodes {
			newNames[n.Name] = true
		}
		added := 0
		for n := range newNames {
			if !oldNames[n] {
				added++
			}
		}
		removed := 0
		for n := range oldNames {
			if !newNames[n] {
				removed++
			}
		}
		sub.Nodes = nodes
		for _, node := range sub.Nodes {
			if node.Region == "" {
				node.Region = region.DetectRegion(node)
			}
		}
		sub.LastFetched = time.Now()
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"subscription": sub,
			"added":        added,
			"removed":      removed,
			"total":        len(nodes),
		})
		return
	}

	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (s *Server) fetchWithFallback(sub *config.Subscription) ([]byte, error) {
	var fallbackNode *subscription.Node
	var fallbackSub *config.Subscription

	fallbackSub = s.cfg.FindFallbackSub(sub.Name)
	if fallbackSub != nil {
		fallbackNode = fallbackSub.FindNode(fallbackSub.LastNode)
	}
	if fallbackNode == nil {
		for _, candidate := range s.cfg.Subscriptions {
			if candidate.Name == sub.Name {
				continue
			}
			if len(candidate.Nodes) > 0 {
				fallbackNode = candidate.Nodes[0]
				fallbackSub = candidate
				break
			}
		}
	}
	if fallbackNode == nil {
		return nil, fmt.Errorf("no fallback node available")
	}

	socksPort, err := xrayproxy.GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("get free port: %w", err)
	}
	httpPort, err := xrayproxy.GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("get free port: %w", err)
	}

	log.Printf("Starting fallback proxy with node '%s' on ports socks=%d http=%d", fallbackNode.Name, socksPort, httpPort)
	var proxyServer ProxyServer
	if fallbackNode.Protocol == "anytls" {
		proxyServer, err = singbox.Start(fallbackNode, socksPort, httpPort, config.RouteModeGlobal, nil, nil)
	} else {
		proxyServer, err = xrayproxy.Start(fallbackNode, socksPort, httpPort, config.RouteModeGlobal, nil, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("start fallback proxy: %w", err)
	}
	defer proxyServer.Stop()
	time.Sleep(300 * time.Millisecond)

	proxyAddr := fmt.Sprintf("0.0.0.0:%d", socksPort)
	data, err := subscription.FetchWithProxy(sub.URL, proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("fallback fetch failed: %w", err)
	}
	log.Printf("Fallback fetch succeeded for '%s'", sub.Name)
	return data, nil
}