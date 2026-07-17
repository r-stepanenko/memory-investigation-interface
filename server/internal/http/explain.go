package http

import (
	"encoding/json"
	"event-memory-search-api/internal/domain"
	"net/http"
	"strings"
)

func (s *Server) ExplainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	path := strings.TrimPrefix( // /api/search/{search_id}/candidates/{event_id}/explain
		r.URL.Path,
		"/api/search/",
	)

	parts := strings.Split(path, "/")

	if len(parts) != 4 ||
		parts[1] != "candidates" ||
		parts[3] != "explain" {

		http.NotFound(w, r)
		return
	}

	searchID := parts[0]
	eventID := parts[2]
	s.mu.RLock()
	searchResult, ok := s.Searches[searchID]
	s.mu.RUnlock()

	if !ok {
		http.Error(
			w,
			"search not found",
			http.StatusNotFound,
		)
		return
	}

	for _, candidate := range searchResult.Candidates {
		if candidate.Event.EventID == eventID {
			response := domain.ExplainResponse{
				SearchID:      searchID,
				EventID:       eventID,
				Score:         candidate.Score,
				Contributions: candidate.Contributions,
			}

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(
					w,
					"failed to encode response",
					http.StatusInternalServerError,
				)
				return
			}

			return
		}
	}

	http.Error(
		w,
		"candidate not found",
		http.StatusNotFound,
	)
}
