package search

import (
	"time"

	"event-memory-search-api/internal/domain"
)

func FindNearbyEvents(
	events []domain.Event,
	target domain.Event,
	before time.Duration,
	after time.Duration,
	actions []string,
) []domain.Event {

	result := []domain.Event{}

	targetTime, _ := time.Parse(
		time.RFC3339,
		target.Timestamp,
	)

	for _, e := range events {

		if e.EventID == target.EventID {
			continue
		}

		// важно
		if e.UserID != target.UserID {
			continue
		}

		eventTime, err := time.Parse(
			time.RFC3339,
			e.Timestamp,
		)

		if err != nil {
			continue
		}

		diff := eventTime.Sub(targetTime)

		if diff < -before || diff > after {
			continue
		}

		for _, action := range actions {
			if e.Action == action {
				result = append(result, e)
			}
		}
	}

	return result
}