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
		{
			EventID:   "evt_1",
			Timestamp: "2026-06-20T10:00:00Z",
			UserID:    "ivan",
			Action:    "email_send",
			FileName:  "clients.xlsx",
		},
		{
			EventID:   "evt_2",
			Timestamp: "2026-06-20T10:20:00Z",
			UserID:    "alex",
			Action:    "file_copy",
			FileName:  "salary.xlsx",
		},
	}

	eventMap := map[string]domain.Event{
		"evt_1": events[0],
		"evt_2": events[1],
	}

	eventIndex := map[string]int{
		"evt_1": 0,
		"evt_2": 1,
	}

	return &Server{
		Datasets: map[string][]domain.Event{
			"control": events,
		},
		Searches:   make(map[string]domain.SearchResponse),
		Events:     eventMap,
		EventIndex: eventIndex,
	}
}

func TestSearchHandler(t *testing.T) {
	server := newTestServer()
	reqBody := domain.SearchRequest{
		DatasetID: "control",
	}
	reqBody.Hints.UserID = "ivan"
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/search",
		bytes.NewBuffer(body),
	)

	rec := httptest.NewRecorder()

	server.SearchHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
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

	server.SearchHandler(rec, req)

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
	reqBody := domain.SearchRequest{
		DatasetID: "unknown",
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/search",
		bytes.NewBuffer(body),
	)
	rec := httptest.NewRecorder()
	server.SearchHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected %d got %d",
			http.StatusNotFound,
			rec.Code,
		)
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

	server.Datasets["control"] = append(
		server.Datasets["control"],
		domain.Event{
			EventID:   "evt_3",
			Timestamp: "2026-06-20T10:20:00Z",
			UserID:    "ivan",
			Action:    "create_archive",
			FileName:  "archive.zip",
		},
	)

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

	for _, hint := range response.Candidates[0].MatchedHints {

		if hint == "nearby event found" {

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
