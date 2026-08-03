package search

import (
	"time"

	"event-memory-search-api/internal/domain"
)

func FindNearbyEvents(
	events []domain.Event,
	target domain.Event,
	requirements []domain.NearbyRequirement,
	before time.Duration,
	after time.Duration,
) []domain.Event {

	result := []domain.Event{}

	targetTime, err := time.Parse(
		time.RFC3339,
		target.Timestamp,
	)

	if err != nil {
		return nil
	}

	startWindow := targetTime.Add(-before)
	endWindow := targetTime.Add(after)

	for _, e := range events {

		if e.EventID == target.EventID {
			continue
		}

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

		if before > 0 || after > 0 {

			if eventTime.Before(startWindow) ||
				eventTime.After(endWindow) {
				continue
			}
		}

		diff := eventTime.Sub(targetTime)

		if diff < 0 {
			diff = -diff
		}

		for _, req := range requirements {

			if e.Action != req.Action {
				continue
			}

			if req.Within != "" {

				within, err := ParseTolerance(req.Within)

				if err != nil {
					continue
				}

				if diff > within {
					continue
				}
			}

			result = append(result, e)
		}
	}

	return result
}
