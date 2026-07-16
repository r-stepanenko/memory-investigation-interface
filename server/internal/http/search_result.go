package http

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Server) SearchResultHandler(w http.ResponseWriter, r *http.Request) {

	if strings.Contains(r.URL.Path, "/candidates/") {
		s.ExplainHandler(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	searchID := strings.TrimPrefix(
		r.URL.Path,
		"/api/search/",
	)

	result, ok := s.Searches[searchID]
	if !ok {
		http.Error(
			w,
			"search not found",
			http.StatusNotFound,
		)
		return
	}

	json.NewEncoder(w).Encode(result)
}