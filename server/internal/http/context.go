package http

import (
	"encoding/json"
	
	"net/http"
	"strings"
	
	"event-memory-search-api/internal/domain"
	"time"
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

	events := s.Datasets["control"]

	before := []domain.Event{}
	after := []domain.Event{}

	targetTime, err := time.Parse(
		time.RFC3339,
		event.Timestamp,
	)

	if err != nil {
		WriteError(
			w,
			http.StatusInternalServerError,
			"INVALID_TIMESTAMP",
			"event timestamp invalid",
		)
		return
	}

	// временные окна
	beforeWindow := 30 * time.Minute
	afterWindow := 30 * time.Minute

	for _, e := range events {

		if e.EventID == event.EventID {
			continue
		}

		if e.UserID != event.UserID {
			continue
		}

		if e.FileName != event.FileName {
			continue
		}

		eventTime, err := time.Parse(
			time.RFC3339,
			e.Timestamp,
		)

		if err != nil {
			continue
		}

		// события до
		if eventTime.Before(targetTime) {

			diff := targetTime.Sub(eventTime)

			if diff <= beforeWindow {
				before = append(before, e)
			}
		}

		// события после
		if eventTime.After(targetTime) {

			diff := eventTime.Sub(targetTime)

			if diff <= afterWindow {
				after = append(after, e)
			}
		}
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
