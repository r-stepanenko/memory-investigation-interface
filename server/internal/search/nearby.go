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
	requiredActions []string,
) []domain.Event {

	result := make([]domain.Event, 0)

	targetTime, err := time.Parse(
		time.RFC3339,
		target.Timestamp,
	)

	if err != nil {
		return result
	}

	start := targetTime.Add(-before)

	end := targetTime.Add(after)

	for _, event := range events {

		if event.EventID == target.EventID {
			continue
		}

		eventTime, err := time.Parse(
			time.RFC3339,
			event.Timestamp,
		)

		if err != nil {
			continue
		}

		if eventTime.Before(start) ||
			eventTime.After(end) {

			continue
		}

		for _, action := range requiredActions {
			if event.Action == action {
				result = append(
					result,
					event,
				)
				break
			}
		}
	}
	return result
}