package http

import (
	"encoding/json"
	"testing"
	"net/http"
	"event-memory-search-api/internal/domain"
)

func TestScoringRankingAndMatchedHints(t *testing.T) {

	server := newTestServer()

	rec := executeSearch(
		t,
		server,
		map[string]any{
			"dataset_id": "control",
			"hints": map[string]string{
				"user_id": "ivan",
			},
		},
	)

	if rec.Code != 200 {
		t.Fatalf("search failed: %d", rec.Code)
	}

	var response domain.SearchResponse

	err := json.NewDecoder(rec.Body).Decode(&response)

	if err != nil {
		t.Fatal(err)
	}

	if len(response.Candidates) == 0 {
		t.Fatal("no candidates")
	}

	for _, c := range response.Candidates {

		if c.Score <= 0 {
			t.Fatalf(
				"invalid score %v",
				c.Score,
			)
		}

		if len(c.MatchedHints) == 0 {
			t.Fatalf(
				"matched hints empty",
			)
		}
	}
}

func TestScoringConformance(t *testing.T) {

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
		},
	)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}

	var response domain.SearchResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if len(response.Candidates) == 0 {
		t.Fatal("expected candidates")
	}

	candidate := response.Candidates[0]

	if candidate.Score <= 0 {
		t.Fatalf(
			"expected positive score got %v",
			candidate.Score,
		)
	}

	if len(candidate.MatchedHints) == 0 {
		t.Fatal(
			"expected matched hints",
		)
	}
}
