package search

import (
	"path/filepath"
	"testing"

	"event-memory-search-api/internal/datasets"
	"event-memory-search-api/internal/domain"
)

func BenchmarkSearch100K(b *testing.B) {

	events, _, err := datasets.LoadEvents(
		filepath.Join("..", "datasets", "events100k.jsonl"),
	)

	if err != nil {
		b.Fatal(err)
	}

	index := BuildUserIndex(events)

	hints := domain.SearchHints{
		UserID: "ivan",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		candidates := index.Find(hints.UserID)

		for _, event := range candidates {

			CalculateScore(
				event,
				hints,
				nil,
				false,
			)
		}
	}
}

func BenchmarkSearch1M(b *testing.B) {

	events, _, err := datasets.LoadEvents(
		filepath.Join("..", "datasets", "events1m.jsonl"),
	)

	if err != nil {
		b.Fatal(err)
	}

	index := BuildUserIndex(events)

	hints := domain.SearchHints{
		UserID: "ivan",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		candidates := index.Find(hints.UserID)

		for _, event := range candidates {

			CalculateScore(
				event,
				hints,
				nil,
				false,
			)
		}
	}
}