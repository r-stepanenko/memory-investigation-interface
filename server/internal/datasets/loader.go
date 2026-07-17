package datasets

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"event-memory-search-api/internal/domain"
)

func LoadEvents(path string) []domain.Event {

	file, err := os.Open(path)

	if err != nil {
		log.Fatalf(
			"failed to open dataset: %v",
			err,
		)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	events := make([]domain.Event, 0)

	for scanner.Scan() {

		var event domain.Event

		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			log.Printf(
				"skip invalid event: %v",
				err,
			)
			continue
		}

		events = append(
			events,
			event,
		)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf(
			"failed reading dataset: %v",
			err,
		)
	}

	return events
}
