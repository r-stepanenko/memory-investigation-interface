package http

import (
	"encoding/json"
	"event-memory-search-api/internal/domain"
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) ExplainHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if r.Method != http.MethodGet {
		WriteError(
			w,
			http.StatusMethodNotAllowed,
			"METHOD_NOT_ALLOWED",
			"method not allowed",
		)
		return
	}

	path := strings.TrimPrefix( // /api/search/{search_id}/candidates/{event_id}/explain
		r.URL.Path,
		"/api/search/",
	)

	parts := strings.Split(path, "/")

	fmt.Println("RAW PATH:", r.URL.Path)
	fmt.Println("PATH:", path)
	fmt.Println("PARTS:", parts)

	if len(parts) != 4 ||
		parts[1] != "candidates" ||
		parts[3] != "explain" {

		WriteError(
			w,
			http.StatusNotFound,
			"INVALID_PATH",
			"invalid explain path",
		)
		return
	}

	searchID := parts[0]
	eventID := parts[2]

	s.mu.RLock()
	searchResult, ok := s.Searches[searchID]
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

	for _, candidate := range searchResult.Candidates {

		if candidate.Event.EventID == eventID {

			response := domain.ExplainResponse{
				SearchID:      searchID,
				EventID:       eventID,
				Score:         candidate.Score,
				Contributions: candidate.Contributions,
				MissedHints:   candidate.MissedHints,
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				WriteError(
					w,
					http.StatusInternalServerError,
					"INTERNAL_ERROR",
					"failed to encode response",
				)
				return
			}

			return
		}
	}

	WriteError(
		w,
		http.StatusNotFound,
		"CANDIDATE_NOT_FOUND",
		"candidate not found",
	)
}
