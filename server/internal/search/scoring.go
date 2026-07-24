package search

import (
	"event-memory-search-api/internal/domain"
	"strings"
)

func CalculateScore(
	event domain.Event,
	hints domain.SearchHints,
	nearby []domain.Event,
	requireNearby bool,
) (
	float64,
	[]string,
	[]domain.Contribution,
	[]domain.MissedHint,
) {

	score := 0.0

	matched := make([]string, 0)
	contributions := make([]domain.Contribution, 0)
	missedHints := make([]domain.MissedHint, 0)

	// считаем активные hints

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

	if hints.Channel != "" {
		hintCount++
	}

	if hints.Severity != "" {
		hintCount++
	}

	if hintCount == 0 && !requireNearby {
		return 0, matched, contributions, missedHints
	}

	weight := 0.0

	if hintCount > 0 {
		weight = 100 / float64(hintCount)
	}

	// USER_ID

	if hints.UserID != "" {

		matchType, multiplier := MatchUserID(
			event.UserID,
			hints.UserID,
		)

		points := weight * multiplier

		contributions = append(
			contributions,
			domain.Contribution{
				Hint:    "user_id",
				Type:    string(matchType),
				Value:   event.UserID,
				Query:   hints.UserID,
				Points:  points,
				Matched: matchType != None,
				Reason:  string(matchType) + " user id match",
			},
		)

		if matchType != None {

			score += points

			matched = append(
				matched,
				"user_id "+string(matchType),
			)

		} else {

			missedHints = append(
				missedHints,
				domain.MissedHint{
					Hint:   "user_id",
					Reason: "value does not match",
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

		points := weight * multiplier

		contributions = append(
			contributions,
			domain.Contribution{
				Hint:    "file_name",
				Type:    string(matchType),
				Value:   event.FileName,
				Query:   hints.FileName,
				Points:  points,
				Matched: matchType != None,
				Reason:  string(matchType) + " filename match",
			},
		)

		if matchType != None {

			score += points

			matched = append(
				matched,
				"file_name "+string(matchType),
			)

		} else {

			missedHints = append(
				missedHints,
				domain.MissedHint{
					Hint:   "file_name",
					Reason: "value does not match",
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

		points := weight * multiplier

		contributions = append(
			contributions,
			domain.Contribution{
				Hint:    "action",
				Type:    string(matchType),
				Value:   event.Action,
				Query:   hints.Action,
				Points:  points,
				Matched: matchType != None,
				Reason:  string(matchType) + " action match",
			},
		)

		if matchType != None {

			score += points

			matched = append(
				matched,
				"action "+string(matchType),
			)

		} else {

			missedHints = append(
				missedHints,
				domain.MissedHint{
					Hint:   "action",
					Reason: "value does not match",
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

		points := weight * multiplier

		contributions = append(
			contributions,
			domain.Contribution{
				Hint:    "destination_type",
				Type:    string(matchType),
				Value:   event.DestinationType,
				Query:   hints.DestinationType,
				Points:  points,
				Matched: matchType != None,
				Reason:  string(matchType) + " destination type match",
			},
		)

		if matchType != None {

			score += points

			matched = append(
				matched,
				"destination_type "+string(matchType),
			)

		} else {

			missedHints = append(
				missedHints,
				domain.MissedHint{
					Hint:   "destination_type",
					Reason: "value does not match",
				},
			)
		}
	}

	// CHANNEL

	if hints.Channel != "" {

		matchType, multiplier := MatchExact(
			event.Channel,
			hints.Channel,
		)

		points := weight * multiplier

		contributions = append(
			contributions,
			domain.Contribution{
				Hint:    "channel",
				Type:    string(matchType),
				Value:   event.Channel,
				Query:   hints.Channel,
				Points:  points,
				Matched: matchType != None,
				Reason:  string(matchType) + " channel match",
			},
		)

		if matchType != None {

			score += points

			matched = append(
				matched,
				"channel "+string(matchType),
			)

		} else {

			missedHints = append(
				missedHints,
				domain.MissedHint{
					Hint:   "channel",
					Reason: "value does not match",
				},
			)
		}
	}

	// SEVERITY

	if hints.Severity != "" {

		matchType, multiplier := MatchExact(
			event.Severity,
			hints.Severity,
		)

		points := weight * multiplier

		contributions = append(
			contributions,
			domain.Contribution{
				Hint:    "severity",
				Type:    string(matchType),
				Value:   event.Severity,
				Query:   hints.Severity,
				Points:  points,
				Matched: matchType != None,
				Reason:  string(matchType) + " severity match",
			},
		)

		if matchType != None {

			score += points

			matched = append(
				matched,
				"severity "+string(matchType),
			)

		} else {

			missedHints = append(
				missedHints,
				domain.MissedHint{
					Hint:   "severity",
					Reason: "value does not match",
				},
			)
		}
	}

	// NEARBY

	if len(nearby) > 0 {

		values := make([]string, 0)

		for _, e := range nearby {
			values = append(values, e.Action)
		}

		points := 10.0

		if score+points > 100 {
			points = 100 - score
		}

		if points > 0 {
			score += points
		}

		contributions = append(
			contributions,
			domain.Contribution{
				Hint:    "nearby",
				Type:    "context",
				Value:   strings.Join(values, ","),
				Query:   "required nearby",
				Points:  10,
				Matched: true,
				Reason:  "nearby events found",
			},
		)

		matched = append(
			matched,
			"nearby event found",
		)

	} else if requireNearby {

		missedHints = append(
			missedHints,
			domain.MissedHint{
				Hint:   "nearby",
				Reason: "required nearby event not found",
			},
		)

	}

	if score > 100 {
		score = 100
	}

	return score, matched, contributions, missedHints
}
