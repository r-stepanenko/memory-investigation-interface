package search

import (
	"testing"
	"event-memory-search-api/internal/domain"
)

func TestCalculateScoreExactMatch(t *testing.T) {
	event := domain.Event{
		UserID: "ivan",
	}

	hints := domain.SearchHints{
		UserID: "ivan",
	}

	score, matched, contributions := CalculateScore(event, hints)

	if score != 100 {
		t.Fatalf("expected 100, got %f", score)
	}

	if matched[0] != "user_id exact" {
		t.Fatalf("expected exact, got %s", matched[0])
	}

	if contributions[0].Type != "exact" {
		t.Fatalf("expected exact contribution")
	}
}

func TestCalculateScoreSubstringMatch(t *testing.T) {
	event := domain.Event{
		UserID: "ivanov",
	}

	hints := domain.SearchHints{
		UserID: "ivan",
	}

	score, matched, contributions := CalculateScore(event, hints)

	if score != 50 {
		t.Fatalf("expected 50, got %f", score)
	}

	if matched[0] != "user_id substring" {
		t.Fatalf("expected substring, got %s", matched[0])
	}

	if contributions[0].Type != "substring" {
		t.Fatalf("expected substring contribution")
	}
}

func TestCalculateScoreFuzzyMatch(t *testing.T) {
	event := domain.Event{
		UserID: "ivan",
	}

	hints := domain.SearchHints{
		UserID: "ivn",
	}

	score, matched, contributions := CalculateScore(event, hints)

	if score != 45 {
		t.Fatalf("expected 45, got %f", score)
	}

	if matched[0] != "user_id fuzzy" {
		t.Fatalf("expected fuzzy, got %s", matched[0])
	}

	if contributions[0].Type != "fuzzy" {
		t.Fatalf("expected fuzzy contribution")
	}
}

func TestCalculateScoreTwoHints(t *testing.T) {
	event := domain.Event{
		UserID: "ivan",
		Action: "email_send",
	}

	hints := domain.SearchHints{
		UserID: "ivan",
		Action: "email_send",
	}

	score, _, contributions := CalculateScore(event, hints)

	if score != 100 {
		t.Fatalf("expected 100, got %f", score)
	}

	if len(contributions) != 2 {
		t.Fatalf(
			"expected 2 contributions, got %d",
			len(contributions),
		)
	}
}

func TestCalculateScoreNoMatch(t *testing.T) {
	event := domain.Event{
		UserID: "alex",
	}

	hints := domain.SearchHints{
		UserID: "ivan",
	}

	score, matched, contributions := CalculateScore(event, hints)

	if score != 0 {
		t.Fatalf("expected 0, got %f", score)
	}

	if len(matched) != 0 {
		t.Fatalf("expected no matched hints")
	}

	if len(contributions) != 0 {
		t.Fatalf("expected no contributions")
	}
}

func TestCalculateScoreNoHints(t *testing.T) {
	event := domain.Event{
		UserID: "ivan",
	}

	score, _, _ := CalculateScore(
		event,
		domain.SearchHints{},
	)

	if score != 0 {
		t.Fatalf("expected 0, got %f", score)
	}
}
