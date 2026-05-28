package web

import (
	"net/http"
	"strings"

	"xray-go/region"
	"xray-go/subscription"
)

func (s *Server) getAllNodes() []*subscription.Node {
	var nodes []*subscription.Node
	for _, sub := range s.cfg.Subscriptions {
		nodes = append(nodes, sub.Nodes...)
	}
	nodes = append(nodes, s.cfg.StandaloneNodes...)
	return nodes
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
		if err := readJSON(r, &req); err != nil {
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

func (s *Server) handleNodesOrDelete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/regions") {
		s.handleNodeRegions(w, r)
		return
	}
	if r.Method == http.MethodDelete {
		s.handleDeleteStandaloneNode(w, r)
		return
	}
	http.NotFound(w, r)
}
