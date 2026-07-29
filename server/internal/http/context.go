package http

import (
	"encoding/json"

	"net/http"
	"strings"

	"event-memory-search-api/internal/domain"
	"sort"
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

	datasetID := r.URL.Query().Get("dataset")

	var events []domain.Event

	if datasetID != "" {

		var ok bool

		events, ok = s.Datasets[datasetID]

		if !ok {
			WriteError(
				w,
				http.StatusNotFound,
				"DATASET_NOT_FOUND",
				"dataset not found",
			)
			return
		}

	} else {

		found := false

		for _, datasetEvents := range s.Datasets {

			for _, e := range datasetEvents {

				if e.EventID == event.EventID {

					events = datasetEvents
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			WriteError(
				w,
				http.StatusNotFound,
				"DATASET_NOT_FOUND",
				"dataset for event not found",
			)
			return
		}
	}

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

	beforeWindow := 30 * time.Minute
	afterWindow := 30 * time.Minute

	if value := r.URL.Query().Get("before"); value != "" {
		window, err := time.ParseDuration(value)

		if err != nil {
			WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_BEFORE_WINDOW",
				"invalid before duration",
			)
			return
		}

		beforeWindow = window
	}

	if value := r.URL.Query().Get("after"); value != "" {
		window, err := time.ParseDuration(value)

		if err != nil {
			WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_AFTER_WINDOW",
				"invalid after duration",
			)
			return
		}

		afterWindow = window
	}

	for _, e := range events {

		if e.EventID == event.EventID {
			continue
		}

		if e.UserID != event.UserID {
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

	sort.Slice(before, func(i, j int) bool {
		return before[i].Timestamp < before[j].Timestamp
	})

	sort.Slice(after, func(i, j int) bool {
		return after[i].Timestamp < after[j].Timestamp
	})

	response := domain.EventContext{
		Event:  event,
		Before: before,
		After:  after,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
