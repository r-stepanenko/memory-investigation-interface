package search

import (
	"event-memory-search-api/internal/domain"
	"time"
)

func FilterByTime(events []domain.Event, filter domain.TimeFilter) []domain.Event {

	if filter.Around == "" || filter.Tolerance == "" {
		return events
	}

	center, err := time.Parse(time.RFC3339, filter.Around)
	if err != nil {
		return events
	}

	duration, err := time.ParseDuration(filter.Tolerance)
	if err != nil {
		return events
	}

	result := make([]domain.Event, 0)

	for _, event := range events {

		t, err := time.Parse(time.RFC3339, event.Timestamp)

		if err != nil {
			continue
		}

		diff := t.Sub(center)

		if diff < 0 {
			diff = -diff
		}

		if diff <= duration {
			result = append(result, event)
		}
	}

	return result
}
