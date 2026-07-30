package http

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

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

	datasetID := r.URL.Query().Get("dataset")

	var (
		event         domain.Event
		datasetEvents map[string]domain.Event
		ok            bool
	)

	// если dataset передан
	if datasetID != "" {

		datasetEvents, ok = s.Events[datasetID]

		if !ok {
			WriteError(
				w,
				http.StatusNotFound,
				"DATASET_NOT_FOUND",
				"dataset not found",
			)
			return
		}

		event, ok = datasetEvents[id]

		if !ok {
			WriteError(
				w,
				http.StatusNotFound,
				"EVENT_NOT_FOUND",
				"event not found",
			)
			return
		}

	} else {

		// ищем событие во всех dataset

		for _, dataset := range s.Events {

			if found, exists := dataset[id]; exists {
				event = found
				datasetEvents = dataset
				ok = true
				break
			}
		}

		if !ok {
			WriteError(
				w,
				http.StatusNotFound,
				"EVENT_NOT_FOUND",
				"event not found",
			)
			return
		}
	}

	// превращаем map событий в slice

	events := make([]domain.Event, 0, len(datasetEvents))

	for _, e := range datasetEvents {
		events = append(events, e)
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

		if eventTime.Before(targetTime) {

			diff := targetTime.Sub(eventTime)

			if diff <= beforeWindow {
				before = append(before, e)
			}
		}

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

	json.NewEncoder(w).Encode(response)
}
