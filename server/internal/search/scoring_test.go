package search

import (
	"event-memory-search-api/internal/domain"
	"testing"
)

func TestCalculateScoreExactMatch(t *testing.T) {

	event := domain.Event{
		UserID: "ivan",
	}

	hints := domain.SearchHints{
		UserID: "ivan",
	}

	score, matched, contributions, missed := CalculateScore(
		event,
		hints,
		nil,
		false,
	)

	if score != 100 {
		t.Fatalf("expected 100, got %f", score)
	}

	if len(matched) != 1 {
		t.Fatalf("expected 1 matched hint")
	}

	if matched[0] != "user_id exact" {
		t.Fatalf("unexpected match %s", matched[0])
	}

	if len(contributions) != 1 {
		t.Fatalf("expected 1 contribution")
	}

	if contributions[0].Type != "exact" {
		t.Fatalf("expected exact")
	}

	if contributions[0].Points != 100 {
		t.Fatalf("expected 100 points")
	}

	if len(missed) != 0 {
		t.Fatalf("expected no missed hints")
	}
}

func TestCalculateScoreSubstringMatch(t *testing.T) {

	event := domain.Event{
		UserID: "ivanov",
	}

	hints := domain.SearchHints{
		UserID: "ivan",
	}

	score, matched, contributions, _ := CalculateScore(
		event,
		hints,
		nil,
		false,
	)

	if score != 50 {
		t.Fatalf("expected 50 got %f", score)
	}

	if matched[0] != "user_id substring" {
		t.Fatalf("expected substring")
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

	score, matched, contributions, _ := CalculateScore(
		event,
		hints,
		nil,
		false,
	)

	if score != 45 {
		t.Fatalf("expected 45 got %f", score)
	}

	if matched[0] != "user_id fuzzy" {
		t.Fatalf("expected fuzzy")
	}

	if contributions[0].Type != "fuzzy" {
		t.Fatalf("expected fuzzy type")
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

	score, _, contributions, _ := CalculateScore(
		event,
		hints,
		nil,
		false,
	)

	if score != 100 {
		t.Fatalf("expected 100")
	}

	var total float64

	for _, c := range contributions {
		total += c.Points
	}

	if total != score {
		t.Fatalf(
			"score mismatch %f != %f",
			total,
			score,
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

	score, matched, contributions, missed := CalculateScore(
		event,
		hints,
		nil,
		false,
	)

	if score != 0 {
		t.Fatalf("expected zero")
	}

	if len(matched) != 0 {
		t.Fatalf("expected no matches")
	}

	if contributions[0].Matched {
		t.Fatalf("must be false")
	}

	if contributions[0].Points != 0 {
		t.Fatalf("expected zero points")
	}

	if len(missed) != 1 {
		t.Fatalf("expected missed hint")
	}
}

func TestCalculateScoreNoHints(t *testing.T) {

	score, _, contributions, missed := CalculateScore(
		domain.Event{
			UserID: "ivan",
		},
		domain.SearchHints{},
		nil,
		false,
	)

	if score != 0 {
		t.Fatalf("expected zero")
	}

	if len(contributions) != 0 {
		t.Fatalf("expected no contributions")
	}

	if len(missed) != 0 {
		t.Fatalf("expected no missed")
	}
}

func TestCalculateScoreExplainEqualsScore(t *testing.T) {

	event := domain.Event{
		UserID: "ivan",
		Action: "upload",
	}

	hints := domain.SearchHints{
		UserID: "ivan",
		Action: "upload",
	}

	score, _, contributions, _ := CalculateScore(
		event,
		hints,
		nil,
		false,
	)

	var total float64

	for _, c := range contributions {
		total += c.Points
	}

	if total != score {
		t.Fatalf(
			"explain mismatch %f != %f",
			total,
			score,
		)
	}
}

func TestCalculateScoreActionAndDestination(t *testing.T) {

	event := domain.Event{
		Action:          "upload",
		DestinationType: "cloud",
	}

	hints := domain.SearchHints{
		Action:          "upload",
		DestinationType: "cloud",
	}

	score, _, contributions, _ := CalculateScore(
		event,
		hints,
		nil,
		false,
	)

	if score != 100 {
		t.Fatalf("expected 100")
	}

	for _, c := range contributions {

		if !c.Matched {
			t.Fatalf("expected matched")
		}

		if c.Points != 50 {
			t.Fatalf("expected 50")
		}
	}
}

func TestCalculateScoreActionOnly(t *testing.T) {

	score, _, contributions, _ := CalculateScore(
		domain.Event{
			Action: "upload",
		},
		domain.SearchHints{
			Action: "upload",
		},
		nil,
		false,
	)

	if score != 100 {
		t.Fatalf("expected 100")
	}

	if contributions[0].Hint != "action" {
		t.Fatalf("expected action")
	}
}

func TestCalculateScoreNearbyRequired(t *testing.T) {

	event := domain.Event{
		UserID: "ivan",
	}

	hints := domain.SearchHints{
		UserID: "ivan",
	}

	nearby := []domain.Event{
		{
			Action: "login",
		},
	}

	score, _, contributions, missed := CalculateScore(
		event,
		hints,
		nearby,
		true,
	)

	if score != 100 {
		t.Fatalf(
			"expected capped score 100 got %f",
			score,
		)
	}

	if len(contributions) != 2 {
		t.Fatalf("expected two contributions")
	}

	if contributions[1].Hint != "nearby" {
		t.Fatalf("expected nearby")
	}

	if contributions[1].Points != 10 {
		t.Fatalf("expected 10 nearby points")
	}

	if len(missed) != 0 {
		t.Fatalf("unexpected missed")
	}
}

func TestCalculateScoreNearbyMissing(t *testing.T) {

	event := domain.Event{
		UserID: "ivan",
	}

	hints := domain.SearchHints{
		UserID: "ivan",
	}

	score, _, _, missed := CalculateScore(
		event,
		hints,
		nil,
		true,
	)

	if score != 50 {
		t.Fatalf(
			"expected only user score 50, got %f",
			score,
		)
	}


	if len(missed) != 1 {
		t.Fatalf(
			"expected missed nearby",
		)
	}


	if missed[0].Hint != "nearby" {
		t.Fatalf(
			"expected nearby missed",
		)
	}
}

func TestCalculateScoreNearbyFound(t *testing.T) {

	event := domain.Event{
		UserID: "ivan",
	}

	hints := domain.SearchHints{
		UserID: "ivan",
	}

	nearby := []domain.Event{
		{
			Action: "create_archive",
		},
	}

	score, matched, contributions, missed := CalculateScore(
		event,
		hints,
		nearby,
		true,
	)

	if score != 100 {
		t.Fatalf("expected 100, got %f", score)
	}

	if len(missed) != 0 {
		t.Fatalf("expected no missed hints")
	}

	if len(matched) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matched))
	}

	last := contributions[len(contributions)-1]

	if last.Hint != "nearby" {
		t.Fatalf("expected nearby contribution")
	}

	if !last.Matched {
		t.Fatalf("expected nearby matched")
	}

	if last.Points != 10 {
		t.Fatalf(
			"expected 10 points, got %f",
			last.Points,
		)
	}
}

