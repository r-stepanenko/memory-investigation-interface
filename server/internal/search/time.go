package search

import (
	"strconv"
	"strings"
	"time"
)

func ParseTolerance(s string) (time.Duration, error) {

	s = strings.TrimSpace(
		strings.ToLower(s),
	)

	if strings.HasSuffix(s, "d") {

		days := strings.TrimSuffix(s, "d")

		n, err := strconv.Atoi(days)

		if err != nil || n < 0 {
			return 0, err
		}

		return time.Duration(n) * 24 * time.Hour, nil
	}

	return time.ParseDuration(s)
}

func MatchTime(eventTime, around, tolerance string) bool {

	if around == "" && tolerance == "" {
		return true
	}

	if around == "" || tolerance == "" {
		return false
	}

	event, err := time.Parse(time.RFC3339, eventTime)
	if err != nil {
		return false
	}

	center, err := time.Parse(time.RFC3339, around)
	if err != nil {
		return false
	}

	d, err := ParseTolerance(tolerance)
	if err != nil {
		return false
	}

	diff := event.Sub(center)
	if diff < 0 {
		diff = -diff
	}

	return diff <= d
}
