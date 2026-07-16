package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"event-memory-search-api/internal/domain"
)

func TestSearchResultHandler(t *testing.T) {
	server := newTestServer()

	server.Searches["srch_test"] = domain.SearchResponse{
		SearchID: "srch_test",
		Status:   "done",
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/search/srch_test",
		nil,
	)
	rec := httptest.NewRecorder()

	server.SearchResultHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestSearchResultHandlerNotFound(t *testing.T) {
	server := newTestServer()
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/search/not_found",
		nil,
	)
	rec := httptest.NewRecorder()

	server.SearchResultHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected %d got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}