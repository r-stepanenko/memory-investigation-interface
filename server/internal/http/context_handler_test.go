package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContextHandler(t *testing.T) {
	server := newTestServer()
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/events/evt_1/context",
		nil,
	)

	rec := httptest.NewRecorder()

	server.ContextHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestContextHandlerNotFound(t *testing.T) {
	server := newTestServer()
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/events/unknown/context",
		nil,
	)

	rec := httptest.NewRecorder()

	server.ContextHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected %d got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestContextHandlerMethodNotAllowed(t *testing.T) {
	server := newTestServer()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/events/evt_1/context",
		nil,
	)
	rec := httptest.NewRecorder()
	server.ContextHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected %d got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}