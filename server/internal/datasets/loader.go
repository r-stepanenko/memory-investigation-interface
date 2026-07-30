package datasets

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"event-memory-search-api/internal/domain"
)

type LoadStats struct {
	Loaded  int
	Skipped int
}

func LoadEvents(path string) ([]domain.Event, LoadStats, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, LoadStats{}, fmt.Errorf("failed to open dataset: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	events := make([]domain.Event, 0)
	stats := LoadStats{}

	for scanner.Scan() {
		var event domain.Event

		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			stats.Skipped++
			continue
		}

		events = append(events, event)
		stats.Loaded++
	}

	if err := scanner.Err(); err != nil {
		return nil, stats, fmt.Errorf("failed reading dataset: %w", err)
	}

	return events, stats, nil
}