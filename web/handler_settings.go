package web

import (
	"net/http"
	"strings"

	"xray-go/config"
)

func (s *Server) handleRouteMode(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]string{"route_mode": string(s.cfg.RouteMode)})
		return
	}
	if r.Method == http.MethodPut {
		var req struct {
			RouteMode config.RouteMode `json:"route_mode"`
		}
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		switch req.RouteMode {
		case config.RouteModeGlobal, config.RouteModeWhitelist, config.RouteModeBlacklist:
			s.cfg.RouteMode = req.RouteMode
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid route_mode"})
			return
		}
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"route_mode": string(s.cfg.RouteMode)})
		return
	}
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (s *Server) handleWhitelist(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string][]string{"whitelist": s.cfg.Whitelist})
		return
	}
	if r.Method == http.MethodPut {
		var req struct {
			Whitelist []string `json:"whitelist"`
		}
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		s.cfg.Whitelist = req.Whitelist
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string][]string{"whitelist": s.cfg.Whitelist})
		return
	}
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (s *Server) handleBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string][]string{"blacklist": s.cfg.Blacklist})
		return
	}
	if r.Method == http.MethodPut {
		var req struct {
			Blacklist []string `json:"blacklist"`
		}
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		s.cfg.Blacklist = req.Blacklist
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string][]string{"blacklist": s.cfg.Blacklist})
		return
	}
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (s *Server) handleDeleteStandaloneNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	path := r.URL.Path
	name := strings.TrimPrefix(path, "/api/nodes/")
	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "node name required"})
		return
	}
	node := s.cfg.FindStandaloneNode(name)
	if node == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "node not found"})
		return
	}
	for i, n := range s.cfg.StandaloneNodes {
		if n.Name == name {
			s.cfg.RemoveStandaloneNode(i)
			break
		}
	}
	if err := s.cfg.Save(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
