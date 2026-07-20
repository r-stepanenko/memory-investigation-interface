package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"event-memory-search-api/internal/domain"
)

func (s *Server) ContextHandler(w http.ResponseWriter, r *http.Request) {

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

	// /api/events/{event_id}/context

	path := strings.TrimPrefix(
		r.URL.Path,
		"/api/events/",
	)

	parts := strings.Split(path, "/")

	if len(parts) != 2 || parts[1] != "context" {
		WriteError(
			w,
			http.StatusNotFound,
			"INVALID_PATH",
			"invalid event context path",
		)
		return
	}

	id := parts[0]

	event, ok := s.Events[id]

	if !ok {
		WriteError(
			w,
			http.StatusNotFound,
			"EVENT_NOT_FOUND",
			"event not found",
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

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
