package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"event-memory-search-api/internal/domain"
	apphttp "event-memory-search-api/internal/http"
)

func createTestServer() *apphttp.Server {

	events := []domain.Event{
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
	}

	return &apphttp.Server{

		Datasets: map[string][]domain.Event{
			"control": events,
		},

		Searches: make(
			map[string]domain.SearchResponse,
		),

		EventIndex: map[string]map[string]int{

			"control": {
				"evt1": 0,
				"evt2": 1,
				"evt3": 2,
				"evt4": 3,
			},
		},
	}
}

func TestSearchTimeFilter(t *testing.T) {

	server := createTestServer()

	request := map[string]interface{}{

		"dataset_id": "control",

		"time": map[string]string{
			"around":    "2026-06-20T11:00:00Z",
			"tolerance": "10m",
		},

		"hints": map[string]string{
			"user_id": "ivan",
			"action":  "login",
		},
	}

	body, err := json.Marshal(request)

	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/search",
		bytes.NewReader(body),
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

	if rec.Code != http.StatusOK {

		t.Fatalf(
			"expected status 200, got %d body=%s",
			rec.Code,
			rec.Body.String(),
		)
	}

	var response domain.SearchResponse

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(response.Candidates) == 0 {

		t.Fatal(
			"expected candidates",
		)
	}

	for _, candidate := range response.Candidates {

		ts := candidate.Event.Timestamp

		if ts < "2026-06-20T10:50:00Z" ||
			ts > "2026-06-20T11:10:00Z" {

			t.Fatalf(
				"event outside time window: %s",
				ts,
			)
		}
	}
}
