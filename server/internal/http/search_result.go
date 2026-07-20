package http

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Server) SearchResultHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if strings.Contains(r.URL.Path, "/candidates/") {
		s.ExplainHandler(w, r)
		return
	}

	if r.Method != http.MethodGet {
		WriteError(
			w,
			http.StatusMethodNotAllowed,
			"METHOD_NOT_ALLOWED",
			"method not allowed",
		)
		return
	}

	searchID := strings.TrimPrefix(
		r.URL.Path,
		"/api/search/",
	)

	if searchID == "" {
		WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"search_id is required",
		)
		return
	}

	s.mu.RLock()
	result, ok := s.Searches[searchID]
	s.mu.RUnlock()

	if !ok {
		WriteError(
			w,
			http.StatusNotFound,
			"SEARCH_NOT_FOUND",
			"search not found",
		)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"failed to encode response",
		)
		return
	}
}