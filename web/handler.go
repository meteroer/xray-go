package web

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"json encode error"}`, http.StatusInternalServerError)
	}
}

func readJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return json.Unmarshal([]byte{}, v)
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
