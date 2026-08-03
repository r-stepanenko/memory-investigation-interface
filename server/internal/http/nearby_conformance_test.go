package http

import (
	"encoding/json"
	"net/http"
	"testing"

	"event-memory-search-api/internal/domain"
)

func TestNearbyConformance(t *testing.T) {

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

				"before": "30m",

				"after": "30m",

				"require_nearby": []map[string]string{
					{
						"action": "create_archive",
						"within": "10m",
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

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if len(response.Candidates) == 0 {
		t.Fatal("expected nearby candidate")
	}

	foundNearby := false

	for _, candidate := range response.Candidates {

		for _, contribution := range candidate.Contributions {

			if contribution.Hint == "nearby" {

				foundNearby = true

				if !contribution.Matched {
					t.Fatal(
						"nearby contribution must be matched",
					)
				}

				if contribution.Points != 10 {
					t.Fatalf(
						"expected nearby points 10 got %v",
						contribution.Points,
					)
				}
			}
		}
	}

	if !foundNearby {
		t.Fatal(
			"nearby contribution missing",
		)
	}
}


func TestSearchNearbyWithBeforeAfter(t *testing.T) {

	server := newTestServer()

	rec := executeSearch(
		t,
		server,
		map[string]any{

			"dataset_id": "control",

			"hints": map[string]string{
				"user_id": "ivan",
			},

			"context": map[string]any{

				"before": "30m",

				"after": "30m",

				"require_nearby": []map[string]string{
					{
						"action": "create_archive",
						"within": "10m",
					},
				},
			},
		},
	)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200 got %d",
			rec.Code,
		)
	}

	var response domain.SearchResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if len(response.Candidates) == 0 {
		t.Fatal(
			"expected candidate with nearby context",
		)
	}

	found := false

	for _, candidate := range response.Candidates {

		for _, contribution := range candidate.Contributions {

			if contribution.Hint == "nearby" &&
				contribution.Matched &&
				contribution.Points == 10 {

				found = true
			}
		}
	}

	if !found {
		t.Fatal(
			"expected matched nearby contribution from before/after window",
		)
	}
}