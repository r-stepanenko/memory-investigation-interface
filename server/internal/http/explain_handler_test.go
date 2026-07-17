package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"event-memory-search-api/internal/domain"
)

func TestExplainHandler(t *testing.T) {
	server := newTestServer()
	searchID := "srch_test_1"

	server.Searches[searchID] = domain.SearchResponse{
		SearchID: searchID,
		Status:   "done",
		Candidates: []domain.SearchResult{
			{
				Score: 50,
				Event: domain.Event{
					EventID: "evt_1",
					UserID:  "ivan",
					Action:  "email_send",
				},
				Contributions: []domain.Contribution{
					{
						Hint:   "user_id",
						Type:   "substring",
						Query:  "ivan",
						Value:  "ivan",
						Points: 50,
					},
				},
			},
		},
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/search/srch_test_1/candidates/evt_1/explain",
		nil,
	)

	rec := httptest.NewRecorder()

	server.ExplainHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestExplainHandlerSearchNotFound(t *testing.T) {
	server := newTestServer()
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/search/unknown/candidates/evt_1/explain",
		nil,
	)
	rec := httptest.NewRecorder()

	server.ExplainHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected %d got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestExplainHandlerCandidateNotFound(t *testing.T) {

	server := newTestServer()

	searchID := "srch_test_1"

	server.Searches[searchID] = domain.SearchResponse{
		SearchID: searchID,
		Status: "done",
		Candidates: []domain.SearchResult{
			{
				Event: domain.Event{
					EventID: "evt_1",
				},
			},
		},
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/search/srch_test_1/candidates/evt_999/explain",
		nil,
	)
	rec := httptest.NewRecorder()

	server.ExplainHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected %d got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestExplainHandlerMethodNotAllowed(t *testing.T) {
	server := newTestServer()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/search/srch_test_1/candidates/evt_1/explain",
		nil,
	)
	rec := httptest.NewRecorder()

	server.ExplainHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected %d got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestExplainRoutingThroughMux(t *testing.T) {

	server := newTestServer()

	server.Searches["srch_test_1"] = domain.SearchResponse{
		SearchID: "srch_test_1",
		Status:   "done",
		Candidates: []domain.SearchResult{
			{
				Event: domain.Event{
					EventID: "evt_1",
				},
			},
		},
	}

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/api/search/",
		server.SearchResultHandler,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/search/srch_test_1/candidates/evt_1/explain",
		nil,
	)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}