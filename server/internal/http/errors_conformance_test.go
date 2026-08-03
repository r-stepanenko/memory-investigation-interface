package http

import (
	"net/http"
	"testing"
	"encoding/json"
)

func TestSearchUnknownDatasetReturnsStructuredError(
	t *testing.T,
) {

	server := newTestServer()

	rec := executeSearch(
		t,
		server,
		map[string]any{
			"dataset_id": "wrong",
		},
	)

	if rec.Code != 404 {

		t.Fatalf(
			"expected 404 got %d",
			rec.Code,
		)
	}

	body := rec.Body.String()

	if body == "" {

		t.Fatal(
			"empty error response",
		)
	}

}

func TestStructuredErrorConformance(t *testing.T) {

	server := newTestServer()

	rec := executeSearch(
		t,
		server,
		map[string]any{
			"dataset_id": "unknown",
		},
	)

	if rec.Code == http.StatusOK {
		t.Fatal(
			"expected error",
		)
	}

	var errResp map[string]any

	json.NewDecoder(rec.Body).Decode(&errResp)

	if _, ok := errResp["error"]; !ok {
		t.Fatal(
			"expected structured error field",
		)
	}
}
