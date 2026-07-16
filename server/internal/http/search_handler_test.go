package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"event-memory-search-api/internal/domain"
)

func newTestServer() *Server {
	events := []domain.Event{
		{
			EventID:   "evt_1",
			UserID:    "ivan",
			Action:    "email_send",
			FileName:  "clients.xlsx",
		},
		{
			EventID:   "evt_2",
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
		Searches:  make(map[string]domain.SearchResponse),
		Events:    eventMap,
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