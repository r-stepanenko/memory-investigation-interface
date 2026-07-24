package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"event-memory-search-api/internal/domain"
	myhttp "event-memory-search-api/internal/http"
)

func TestSearchAPIIntegration(t *testing.T) {
	server := &myhttp.Server{
		Datasets: map[string][]domain.Event{
			"control": {
				{
					EventID: "evt_1",
					UserID:  "ivan",
					Action:  "file_copy",
				},
				{
					EventID: "evt_2",
					UserID:  "alex",
					Action:  "login",
				},
			},
		},

		Searches: make(map[string]domain.SearchResponse),

		Events: map[string]domain.Event{
			"evt_1": {
				EventID: "evt_1",
				UserID:  "ivan",
				Action:  "file_copy",
			},
			"evt_2": {
				EventID: "evt_2",
				UserID:  "alex",
				Action:  "login",
			},
		},

		EventIndex: map[string]int{
			"evt_1": 0,
			"evt_2": 1,
		},
	}

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/api/search",
		server.SearchHandler,
	)

	mux.HandleFunc(
		"/api/search/",
		server.SearchRouter,
	)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	request := map[string]interface{}{
		"dataset_id": "control",
		"hints": map[string]string{
			"user_id": "ivan",
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(
		ts.URL+"/api/search",
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			resp.StatusCode,
		)
	}

	var searchResp domain.SearchResponse

	err = json.NewDecoder(resp.Body).Decode(&searchResp)

	if err != nil {
		t.Fatal(err)
	}

	if searchResp.SearchID == "" {
		t.Fatal("expected search id")
	}

	if searchResp.TotalCandidates != 1 {
		t.Fatalf(
			"expected 1 candidate got %d",
			searchResp.TotalCandidates,
		)
	}

	// GET /api/search/{id}

	resp2, err := http.Get(
		ts.URL + "/api/search/" + searchResp.SearchID,
	)

	if err != nil {
		t.Fatal(err)
	}

	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected 200 got %d",
			resp2.StatusCode,
		)
	}

	var stored domain.SearchResponse

	err = json.NewDecoder(resp2.Body).Decode(&stored)

	if err != nil {
		t.Fatal(err)
	}

	if stored.SearchID != searchResp.SearchID {
		t.Fatal("search ids mismatch")
	}

	if len(stored.Candidates) != 1 {
		t.Fatalf(
			"expected 1 stored candidate got %d",
			len(stored.Candidates),
		)
	}

	if stored.Candidates[0].Event.EventID != "evt_1" {
		t.Fatalf(
			"expected evt_1 got %s",
			stored.Candidates[0].Event.EventID,
		)
	}
}