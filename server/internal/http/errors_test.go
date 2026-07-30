package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchInvalidDatasetError(t *testing.T) {

	server := createTestServer()

	request := `{
		"dataset_id":"unknown",
		"hints":{
			"user_id":"ivan"
		}
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/search",
		strings.NewReader(request),
	)

	rec := httptest.NewRecorder()

	server.SearchHandler(
		rec,
		req,
	)

	if rec.Code != http.StatusNotFound {

		t.Fatalf(
			"expected status 404, got %d body=%s",
			rec.Code,
			rec.Body.String(),
		)
	}

	var response map[string]interface{}

	err := json.NewDecoder(
		rec.Body,
	).Decode(&response)

	if err != nil {

		t.Fatalf(
			"invalid json response: %v",
			err,
		)
	}

	if len(response) == 0 {

		t.Fatal(
			"empty error response",
		)
	}

	// Проверяем, что API действительно вернул структуру ошибки,
	// а не обычный успешный ответ

	_, hasError := response["error"]

	_, hasCode := response["code"]

	_, hasMessage := response["message"]

	if !hasError && !hasCode && !hasMessage {

		t.Fatalf(
			"structured error fields not found: %+v",
			response,
		)
	}
}
