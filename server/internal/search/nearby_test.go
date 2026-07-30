package search_test

import (
	"testing"
	"time"

	"event-memory-search-api/internal/domain"
	"event-memory-search-api/internal/search"
)

func TestFindNearbyEvents(t *testing.T) {

	event := domain.Event{
		EventID:   "evt1",
		UserID:    "ivan",
		Action:    "login",
		Timestamp: "2026-06-20T11:00:00Z",
	}

	events := []domain.Event{
		event,

		{
			EventID:   "evt2",
			UserID:    "ivan",
			Action:    "create_archive",
			Timestamp: "2026-06-20T10:55:00Z",
		},

		{
			EventID:   "evt3",
			UserID:    "ivan",
			Action:    "logout",
			Timestamp: "2026-06-20T12:00:00Z",
		},
	}

	found := search.FindNearbyEvents(
		events,
		event,
		10*time.Minute,
		10*time.Minute,
		[]string{"create_archive"},
	)

	if len(found) == 0 {
		t.Fatal("nearby event was not found")
	}

	if found[0].EventID != "evt2" {

		t.Fatalf(
			"expected evt2, got %s",
			found[0].EventID,
		)
	}
}
