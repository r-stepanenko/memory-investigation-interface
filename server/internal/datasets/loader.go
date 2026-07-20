package datasets

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"event-memory-search-api/internal/domain"
)

func LoadEvents(path string) ([]domain.Event, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open dataset: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	events := make([]domain.Event, 0)

	for scanner.Scan() {
		var event domain.Event

		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("invalid event: %w", err)
		}

		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading dataset: %w", err)
	}

	return events, nil
}