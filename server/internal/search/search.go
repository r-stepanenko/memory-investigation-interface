package search

import "strings"

type MatchType string

const (
	Exact     MatchType = "exact"
	Substring MatchType = "substring"
	Fuzzy     MatchType = "fuzzy"
	None      MatchType = "none"
)

func Normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	s = strings.NewReplacer(
		"_", " ",
		"-", " ",
		".", " ",
	).Replace(s)

	return s
}

func MatchUserID(value string, query string) (MatchType, float64) {
	value = Normalize(value)
	query = Normalize(query)

	if value == query {
		return Exact, 1
	}

	if strings.Contains(value, query) {
		return Substring, 0.5
	}

	similarity := Similarity(value, query)

	if similarity >= 0.6 {
		return Fuzzy, 0.45
	}

	return None, 0
}

func MatchFileName(value string, query string) (MatchType, float64) {
	value = Normalize(value)
	query = Normalize(query)

	if value == query {
		return Exact, 1
	}

	if strings.Contains(value, query) {
		return Substring, 0.5
	}

	similarity := Similarity(value, query)

	if similarity >= 0.75 {
		return Fuzzy, 0.4
	}

	return None, 0
}

func MatchExact(value string, query string) (MatchType, float64) {
	value = Normalize(value)
	query = Normalize(query)

	if value == query {
		return Exact, 1
	}

	return None, 0
}
