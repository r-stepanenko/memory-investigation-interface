package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"event-memory-search-api/internal/domain"
)

func (s *Server) ContextHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	// /api/events/{event_id}/context

	path := strings.TrimPrefix(
		r.URL.Path,
		"/api/events/",
	)

	parts := strings.Split(path, "/")

	if len(parts) != 2 || parts[1] != "context" {
		http.NotFound(w, r)
		return
	}

	id := parts[0]

	event, ok := s.Events[id]

	if !ok {
		http.Error(
			w,
			"event not found",
			http.StatusNotFound,
		)
		return
	}

	index := s.EventIndex[id]

	events := s.Datasets["control"]

	before := []domain.Event{}

	start := index - 2

	if start < 0 {
		start = 0
	}

	for i := start; i < index; i++ {
		before = append(before, events[i])
	}

	after := []domain.Event{}

	end := index + 3

	if end > len(events) {
		end = len(events)
	}

	for i := index + 1; i < end; i++ {
		after = append(after, events[i])
	}

	response := domain.EventContext{
		Event:  event,
		Before: before,
		After:  after,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(response)
}
