package search

import (
	"strings"

	"event-memory-search-api/internal/domain"
)

func CalculateScore(
	event domain.Event,
	hints domain.SearchHints,
) (
	int,
	[]string,
	[]domain.Contribution,
) {

	score := 0

	matched := make([]string, 0)

	contributions := make([]domain.Contribution, 0)

	// считаем количество активных подсказок

	hintCount := 0

	if hints.UserID != "" {
		hintCount++
	}

	if hints.FileName != "" {
		hintCount++
	}

	if hints.Action != "" {
		hintCount++
	}

	if hints.DestinationType != "" {
		hintCount++
	}

	if hintCount == 0 {
		return 0, matched, contributions
	}

	weight := 100 / hintCount

	// USER_ID

	if hints.UserID != "" {

		value := strings.ToLower(event.UserID)
		query := strings.ToLower(hints.UserID)

		if value == query {

			score += weight

			matched = append(
				matched,
				"user_id exact",
			)

			contributions = append(
				contributions,
				domain.Contribution{
					Hint:   "user_id",
					Type:   "exact",
					Value:  event.UserID,
					Query:  hints.UserID,
					Points: weight,
				},
			)

		} else if strings.Contains(value, query) {

			points := weight / 2

			score += points

			matched = append(
				matched,
				"user_id substring",
			)

			contributions = append(
				contributions,
				domain.Contribution{
					Hint:   "user_id",
					Type:   "substring",
					Value:  event.UserID,
					Query:  hints.UserID,
					Points: points,
				},
			)
		}
	}

	// FILE NAME

	if hints.FileName != "" {

		value := strings.ToLower(event.FileName)
		query := strings.ToLower(hints.FileName)

		if value == query {

			score += weight

			contributions = append(
				contributions,
				domain.Contribution{
					Hint:   "file_name",
					Type:   "exact",
					Value:  event.FileName,
					Query:  hints.FileName,
					Points: weight,
				},
			)
			matched = append(
				matched,
				"file_name exact",
			)

		} else if strings.Contains(value, query) {

			points := weight / 2

			score += points

			matched = append(
				matched,
				"file_name substring",
			)

			contributions = append(
				contributions,
				domain.Contribution{
					Hint:   "file_name",
					Type:   "substring",
					Value:  event.FileName,
					Query:  hints.FileName,
					Points: points,
				},
			)
		}
	}

	// ACTION

	if hints.Action != "" {

		value := strings.ToLower(event.Action)
		query := strings.ToLower(hints.Action)

		if value == query {

			score += weight

			contributions = append(
				contributions,
				domain.Contribution{
					Hint:   "action",
					Type:   "exact",
					Value:  event.Action,
					Query:  hints.Action,
					Points: weight,
				},
			)
			matched = append(
				matched,
				"action exact",
			)

		} else if strings.Contains(value, query) {

			points := weight / 2

			score += points

			matched = append(
				matched,
				"action substring",
			)

			contributions = append(
				contributions,
				domain.Contribution{
					Hint:   "action",
					Type:   "substring",
					Value:  event.Action,
					Query:  hints.Action,
					Points: points,
				},
			)
		}
	}

	if hints.DestinationType != "" {

		value := strings.ToLower(event.DestinationType)
		query := strings.ToLower(hints.DestinationType)

		if value == query {

			score += weight

			matched = append(
				matched,
				"destination_type exact",
			)

			contributions = append(
				contributions,
				domain.Contribution{
					Hint:   "destination_type",
					Type:   "exact",
					Value:  event.DestinationType,
					Query:  hints.DestinationType,
					Points: weight,
				},
			)
		}
	}

	return score, matched, contributions
}
