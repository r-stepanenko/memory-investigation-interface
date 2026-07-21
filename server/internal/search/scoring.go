package search

import "event-memory-search-api/internal/domain"

func CalculateScore(
	event domain.Event,
	hints domain.SearchHints,
) (
	float64,
	[]string,
	[]domain.Contribution,
) {

	score := 0.0

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

	weight := 100.0 / float64(hintCount)

	// USER_ID

	if hints.UserID != "" {

		matchType, multiplier := MatchUserID(
			event.UserID,
			hints.UserID,
		)

		if matchType != None {

			points := weight * multiplier

			score += points

			matched = append(
				matched,
				"user_id "+string(matchType),
			)

			contributions = append(
				contributions,
				domain.Contribution{
					Hint:   "user_id",
					Type:   string(matchType),
					Value:  event.UserID,
					Query:  hints.UserID,
					Points: points,
				},
			)
		}
	}

	// FILE NAME

	if hints.FileName != "" {

		matchType, multiplier := MatchFileName(
			event.FileName,
			hints.FileName,
		)

		if matchType != None {

			points := weight * multiplier

			score += points

			matched = append(
				matched,
				"file_name "+string(matchType),
			)

			contributions = append(
				contributions,
				domain.Contribution{
					Hint:   "file_name",
					Type:   string(matchType),
					Value:  event.FileName,
					Query:  hints.FileName,
					Points: points,
				},
			)
		}
	}

	// ACTION

	if hints.Action != "" {

		matchType, multiplier := MatchExact(
			event.Action,
			hints.Action,
		)

		if matchType != None {

			points := weight * multiplier

			score += points

			matched = append(
				matched,
				"action "+string(matchType),
			)

			contributions = append(
				contributions,
				domain.Contribution{
					Hint:   "action",
					Type:   string(matchType),
					Value:  event.Action,
					Query:  hints.Action,
					Points: points,
				},
			)
		}
	}

	// DESTINATION TYPE

	if hints.DestinationType != "" {

		matchType, multiplier := MatchExact(
			event.DestinationType,
			hints.DestinationType,
		)

		if matchType != None {

			points := weight * multiplier

			score += points

			matched = append(
				matched,
				"destination_type "+string(matchType),
			)

			contributions = append(
				contributions,
				domain.Contribution{
					Hint:   "destination_type",
					Type:   string(matchType),
					Value:  event.DestinationType,
					Query:  hints.DestinationType,
					Points: points,
				},
			)
		}
	}

	return score, matched, contributions
}
