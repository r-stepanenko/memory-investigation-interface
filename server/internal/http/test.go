package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
	"event-memory-search-api/internal/domain"
)

func TestSearchContextExplainFlow(t *testing.T) {

	server := &Server{
		Datasets: map[string][]domain.Event{
			"control": {
				{
					EventID: "evt_33",
					Timestamp: "2026-06-20T11:40:00Z",
					UserID: "ivan",
					Action: "email_send",
				},
			},
		},
		Events: map[string]domain.Event{
			"evt_33": {
				EventID: "evt_33",
				Timestamp: "2026-06-20T11:40:00Z",
				UserID: "ivan",
				Action: "email_send",
			},
		},
		Searches: make(map[string]domain.SearchResponse),
	}


	// SEARCH
	reqBody := `
	{
		"dataset_id":"control",
		"hints":{
			"user_id":"ivan",
			"action":"email_send"
		}
	}
	`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/search",
		strings.NewReader(reqBody),
	)

	rec := httptest.NewRecorder()

	server.SearchHandler(rec, req)


	if rec.Code != http.StatusOK {
		t.Fatalf("search failed: %d", rec.Code)
	}


	var searchResp domain.SearchResponse

	json.NewDecoder(rec.Body).Decode(&searchResp)


	if searchResp.SearchID == "" {
		t.Fatal("empty search id")
	}


	// RESULT
	req = httptest.NewRequest(
		http.MethodGet,
		"/api/search/"+searchResp.SearchID,
		nil,
	)

	rec = httptest.NewRecorder()

	server.SearchResultHandler(rec, req)


	if rec.Code != http.StatusOK {
		t.Fatalf("result failed: %d", rec.Code)
	}


	// CONTEXT
	req = httptest.NewRequest(
		http.MethodGet,
		"/api/events/evt_33/context",
		nil,
	)

	rec = httptest.NewRecorder()

	server.ContextHandler(rec, req)


	if rec.Code != http.StatusOK {
		t.Fatalf("context failed: %d", rec.Code)
	}


	// EXPLAIN
	req = httptest.NewRequest(
		http.MethodGet,
		"/api/search/"+searchResp.SearchID+
			"/candidates/evt_33/explain",
		nil,
	)

	rec = httptest.NewRecorder()

	server.ExplainHandler(rec, req)


	if rec.Code != http.StatusOK {
		t.Fatalf("explain failed: %d", rec.Code)
	}
}