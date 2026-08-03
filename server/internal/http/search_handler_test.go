package http

import (
	"bytes"
	"encoding/json"
	"event-memory-search-api/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer() *Server {

	events := []domain.Event{
		// time filter
		{
			EventID:   "evt_1",
			Timestamp: "2026-06-20T10:55:00Z",
			UserID:    "ivan",
			Action:    "login",
		},
		{
			EventID:   "evt_2",
			Timestamp: "2026-06-20T11:00:00Z",
			UserID:    "ivan",
			Action:    "file_copy",
		},
		{
			EventID:   "evt_3",
			Timestamp: "2026-06-20T11:05:00Z",
			UserID:    "ivan",
			Action:    "logout",
		},
		{
			EventID:   "evt_4",
			Timestamp: "2026-06-20T12:00:00Z",
			UserID:    "ivan",
			Action:    "login",
		},

		// nearby
		{
			EventID:   "evt_33",
			Timestamp: "2026-06-20T11:40:00Z",
			UserID:    "ivan",
			Action:    "email_send",
		},
		{
			EventID:   "evt_34",
			Timestamp: "2026-06-20T11:35:00Z",
			UserID:    "ivan",
			Action:    "create_archive",
		},
	}

	return &Server{
		Datasets: map[string][]domain.Event{
			"control": events,
		},

		Searches: make(map[string]domain.SearchResponse),

		EventIndex: map[string]map[string]int{
			"control": {
				"evt_1":  0,
				"evt_2":  1,
				"evt_3":  2,
				"evt_4":  3,
				"evt_33": 4,
				"evt_34": 5,
			},
		},
	}
}

func executeSearch(
	t *testing.T,
	server *Server,
	body any,
) *httptest.ResponseRecorder {

	data, err := json.Marshal(body)

	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/search",
		bytes.NewReader(data),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	rec := httptest.NewRecorder()

	server.SearchHandler(
		rec,
		req,
	)

	return rec
}

func TestSearchHandler(t *testing.T) {

	server := newTestServer()

	reqBody := domain.SearchRequest{
		DatasetID: "control",
	}

	reqBody.Hints.UserID = "ivan"

	rec := executeSearch(
		t,
		server,
		reqBody,
	)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			rec.Code,
		)
	}
}

func TestSearchHandlerMethodNotAllowed(t *testing.T) {

	server := newTestServer()

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/search",
		nil,
	)

	rec := httptest.NewRecorder()

	server.SearchHandler(
		rec,
		req,
	)

	if rec.Code != http.StatusMethodNotAllowed {

		t.Fatalf(
			"expected %d got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestSearchHandlerDatasetNotFound(t *testing.T) {

	server := newTestServer()

	rec := executeSearch(
		t,
		server,
		map[string]any{
			"dataset_id": "unknown",
		},
	)

	if rec.Code != http.StatusNotFound {

		t.Fatalf(
			"expected %d got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestSearchDatasetRequired(t *testing.T) {

	server := newTestServer()

	rec := executeSearch(
		t,
		server,
		map[string]any{},
	)

	if rec.Code != http.StatusBadRequest {

		t.Fatalf(
			"expected 400 got %d",
			rec.Code,
		)
	}
}

func TestSearchInvalidTime(t *testing.T) {

	server := newTestServer()

	rec := executeSearch(
		t,
		server,
		map[string]any{

			"dataset_id": "control",

			"time": map[string]any{
				"around": "2026-06-20T10:00:00Z",
			},
		},
	)

	if rec.Code != http.StatusBadRequest {

		t.Fatalf(
			"expected 400 got %d",
			rec.Code,
		)
	}
}

func TestSearchNearbyFound(t *testing.T) {

	server := newTestServer()

	event := domain.Event{
		EventID:   "evt_3",
		Timestamp: "2026-06-20T10:10:00Z",
		UserID:    "ivan",
		Action:    "create_archive",
		FileName:  "archive.zip",
	}

	server.Datasets["control"] =
		append(
			server.Datasets["control"],
			event,
		)

	server.EventIndex["control"]["evt_3"] = 2

	rec := executeSearch(
		t,
		server,
		map[string]any{

			"dataset_id": "control",

			"hints": map[string]string{
				"user_id": "ivan",
				"action":  "email_send",
			},

			"context": map[string]any{

				"before": "30m",
				"after":  "30m",

				"require_nearby": []map[string]string{
					{
						"action": "create_archive",
					},
				},
			},
		},
	)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			rec.Code,
		)
	}

	var response domain.SearchResponse

	err := json.NewDecoder(
		rec.Body,
	).Decode(&response)

	if err != nil {
		t.Fatal(err)
	}

	if len(response.Candidates) == 0 {
		t.Fatal(
			"expected nearby candidate",
		)
	}

	found := false

	for _, contribution := range response.Candidates[0].Contributions {

		if contribution.Hint == "nearby" &&
			contribution.Matched {

			found = true
		}
	}

	if !found {

		t.Fatal(
			"nearby contribution missing",
		)
	}
}

func TestSearchNearbyMissing(t *testing.T) {

	server := newTestServer()

	rec := executeSearch(
		t,
		server,
		map[string]any{

			"dataset_id": "control",

			"hints": map[string]string{
				"user_id": "ivan",
				"action":  "email_send",
			},

			"context": map[string]any{

				"before": "5m",
				"after":  "5m",

				"require_nearby": []map[string]string{
					{
						"action": "file_delete",
					},
				},
			},
		},
	)

	var response domain.SearchResponse

	err := json.NewDecoder(
		rec.Body,
	).Decode(&response)

	if err != nil {
		t.Fatal(err)
	}

	if len(response.Candidates) != 0 {

		t.Fatalf(
			"expected zero candidates got %d",
			len(response.Candidates),
		)
	}
}

func TestSearchLimit(t *testing.T) {

	server := newTestServer()

	rec := executeSearch(
		t,
		server,
		map[string]any{

			"dataset_id": "control",

			"scoring": map[string]int{
				"limit": 1,
			},
		},
	)

	var response domain.SearchResponse

	err := json.NewDecoder(
		rec.Body,
	).Decode(&response)

	if err != nil {
		t.Fatal(err)
	}

	if len(response.Candidates) > 1 {

		t.Fatalf(
			"limit ignored: got %d",
			len(response.Candidates),
		)
	}
}
