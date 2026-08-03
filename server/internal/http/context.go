package http

import (
	"encoding/json"
	"event-memory-search-api/internal/domain"
	"net/http"
	"sort"
	"strings"
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
		event  domain.Event
		events []domain.Event
		ok     bool
		idx    int
	)

	if datasetID != "" {

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

		indexMap, exists := s.EventIndex[datasetID]

		if !exists {
			WriteError(
				w,
				http.StatusInternalServerError,
				"EVENT_INDEX_NOT_FOUND",
				"event index not found",
			)
			return
		}

		idx, exists = indexMap[id]

		if !exists {
			WriteError(
				w,
				http.StatusNotFound,
				"EVENT_NOT_FOUND",
				"event not found",
			)
			return
		}

		event = events[idx]

	} else {

		for name, indexMap := range s.EventIndex {

			foundIdx, exists := indexMap[id]

			if exists {
				events = s.Datasets[name]
				event = events[foundIdx]
				idx = foundIdx
				datasetID = name
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

	before := []domain.Event{}
	after := []domain.Event{}

	for i := idx - 1; i >= 0; i-- {

		e := events[i]

		eventTime, err := time.Parse(
			time.RFC3339,
			e.Timestamp,
		)

		if err != nil {
			continue
		}

		if targetTime.Sub(eventTime) > beforeWindow {
			break
		}

		before = append(before, e)
	}

	for i := idx + 1; i < len(events); i++ {

		e := events[i]

		eventTime, err := time.Parse(
			time.RFC3339,
			e.Timestamp,
		)

		if err != nil {
			continue
		}

		if eventTime.Sub(targetTime) > afterWindow {
			break
		}

		after = append(after, e)
	}

	sort.Slice(before, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, before[i].Timestamp)
		t2, _ := time.Parse(time.RFC3339, before[j].Timestamp)

		return t1.Before(t2)
	})

	sort.Slice(after, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, after[i].Timestamp)
		t2, _ := time.Parse(time.RFC3339, after[j].Timestamp)

		return t1.Before(t2)
	})

	response := domain.EventContext{
		Event:  event,
		Before: before,
		After:  after,
	}

	json.NewEncoder(w).Encode(response)
}
