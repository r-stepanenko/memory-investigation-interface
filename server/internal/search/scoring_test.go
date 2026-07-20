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
		t.Fatalf("expected score 100, got %d", score)
	}

	if len(matched) != 1 {
		t.Fatalf("expected 1 matched hint, got %d", len(matched))
	}

	if contributions[0].Type != "exact" {
		t.Fatalf("expected exact match")
	}
}

func TestCalculateScoreSubstringMatch(t *testing.T) {
	event := domain.Event{
		UserID: "ivanov",
	}

	hints := domain.SearchHints{
		UserID: "ivan",
	}

	score, _, contributions := CalculateScore(event, hints)

	if score != 50 {
		t.Fatalf("expected score 50, got %d", score)
	}

	if contributions[0].Type != "substring" {
		t.Fatalf("expected substring match")
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

	score, _, _ := CalculateScore(event, hints)

	if score != 100 {
		t.Fatalf("expected score 100, got %d", score)
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
		t.Fatalf("expected score 0, got %d", score)
	}

	if len(matched) != 0 {
		t.Fatal("expected no matched hints")
	}

	if len(contributions) != 0 {
		t.Fatal("expected no contributions")
	}
}

func TestCalculateScoreNoHints(t *testing.T) {
	event := domain.Event{
		UserID: "ivan",
	}

	score, _, _ := CalculateScore(event, domain.SearchHints{})

	if score != 0 {
		t.Fatalf("expected score 0, got %d", score)
	}
}