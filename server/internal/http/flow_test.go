package http

import (
	"encoding/json"
	"event-memory-search-api/internal/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchContextExplainFlow(t *testing.T) {

	server := &Server{
		Datasets: map[string][]domain.Event{
			"control": {
				{
					EventID:   "evt_33",
					Timestamp: "2026-06-20T11:40:00Z",
					UserID:    "ivan",
					Action:    "email_send",
				},
			},
		},

		Events: map[string]map[string]domain.Event{
			"control": {
				"evt_33": {
					EventID:   "evt_33",
					Timestamp: "2026-06-20T11:40:00Z",
					UserID:    "ivan",
					Action:    "email_send",
				},
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

	err := json.NewDecoder(rec.Body).Decode(&searchResp)

	if err != nil {
		t.Fatal(err)
	}

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
		"/api/events/evt_33/context?dataset=control&before=30m&after=30m",
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

func TestExplainNearbyContribution(t *testing.T) {

	server := &Server{
		Datasets: map[string][]domain.Event{
			"control": {
				{
					EventID:   "evt_1",
					Timestamp: "2026-06-20T11:00:00Z",
					UserID:    "ivan",
					Action:    "email_send",
				},
				{
					EventID:   "evt_2",
					Timestamp: "2026-06-20T11:05:00Z",
					UserID:    "ivan",
					Action:    "file_copy",
				},
			},
		},

		Events: map[string]map[string]domain.Event{

			"control": {

				"evt_1": {
					EventID:   "evt_1",
					Timestamp: "2026-06-20T11:00:00Z",
					UserID:    "ivan",
					Action:    "email_send",
				},

				"evt_2": {
					EventID:   "evt_2",
					Timestamp: "2026-06-20T11:05:00Z",
					UserID:    "ivan",
					Action:    "file_copy",
				},
			},
		},

		Searches: make(map[string]domain.SearchResponse),
	}

	// SEARCH ONLY BY REQUIRE_NEARBY

	reqBody := `
	{
		"dataset_id":"control",
		"context":{
			"require_nearby":[
				{
					"action":"file_copy",
					"within":"10m"
				}
			]
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
		t.Fatalf(
			"search failed: %d",
			rec.Code,
		)
	}

	var searchResp domain.SearchResponse

	err := json.NewDecoder(rec.Body).Decode(&searchResp)

	if err != nil {
		t.Fatal(err)
	}

	if len(searchResp.Candidates) == 0 {
		t.Fatal("no candidates found")
	}

	candidate := searchResp.Candidates[0]

	if candidate.Score != 10 {

		t.Fatalf(
			"expected score 10 got %v",
			candidate.Score,
		)
	}

	// EXPLAIN

	req = httptest.NewRequest(
		http.MethodGet,
		"/api/search/"+searchResp.SearchID+
			"/candidates/"+candidate.Event.EventID+
			"/explain",
		nil,
	)

	rec = httptest.NewRecorder()

	server.ExplainHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"explain failed: %d",
			rec.Code,
		)
	}

	var explain domain.ExplainResponse

	err = json.NewDecoder(rec.Body).Decode(&explain)

	if err != nil {
		t.Fatal(err)
	}

	if explain.Score != candidate.Score {

		t.Fatalf(
			"score mismatch: search=%v explain=%v",
			candidate.Score,
			explain.Score,
		)
	}

	nearbyFound := false

	for _, contribution := range explain.Contributions {

		if contribution.Hint == "nearby" {

			nearbyFound = true

			if contribution.Points != 10 {

				t.Fatalf(
					"expected nearby points 10 got %v",
					contribution.Points,
				)
			}

			if !contribution.Matched {

				t.Fatal(
					"nearby contribution should be matched",
				)
			}
		}
	}

	if !nearbyFound {

		t.Fatal(
			"nearby contribution not found",
		)
	}
}