package http

import (
	"encoding/json"
	"event-memory-search-api/internal/domain"
	"net/http"
	"testing"
)

func TestTimeAroundToleranceConformance(t *testing.T) {

	server := newTestServer()

	rec := executeSearch(
		t,
		server,
		map[string]any{
			"dataset_id": "control",

			"time": map[string]string{
				"around": "2026-06-20T11:00:00Z",

				"tolerance": "10m",
			},

			"hints": map[string]string{
				"user_id": "ivan",
			},
		},
	)

	if rec.Code != 200 {
		t.Fatalf(
			"expected 200 got %d",
			rec.Code,
		)
	}

	var result domain.SearchResponse

	json.NewDecoder(
		rec.Body,
	).Decode(&result)

	for _, candidate := range result.Candidates {

		if candidate.Event.EventID == "evt4" {

			t.Fatal(
				"event outside tolerance included",
			)
		}
	}

}

func TestTimeFilterConformance(t *testing.T) {

	server := newTestServer()

	rec := executeSearch(
		t,
		server,
		map[string]any{
			"dataset_id": "control",

			"hints": map[string]string{
				"user_id": "ivan",
			},

			"time": map[string]string{
				"around":    "2026-06-20T11:00:00Z",
				"tolerance": "10m",
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

	json.NewDecoder(rec.Body).Decode(&response)

	if len(response.Candidates) == 0 {
		t.Fatal(
			"expected candidates inside time window",
		)
	}
}
