package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"event-memory-search-api/internal/domain"
)

func main() {
	count := flag.Int(
		"count",
		100000,
		"number of events",
	)

	out := flag.String(
		"out",
		"events.jsonl",
		"output jsonl file",
	)

	flag.Parse()

	file, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	users := []string{
		"ivan",
		"ivanov",
		"petr",
		"alex",
		"maria",
		"anna",
	}

	actions := []string{
		"login",
		"logout",
		"file_copy",
		"email_send",
		"create_archive",
	}

	files := []string{
		"report.docx",
		"contract.xlsx",
		"clients.zip",
		"database.sql",
	}

	destinations := []string{
		"internal",
		"external",
		"usb",
	}

	channels := []string{
		"email",
		"web",
		"usb",
		"network",
	}

	severities := []string{
		"low",
		"medium",
		"high",
	}

	contentClasses := [][]string{
		{"public"},
		{"internal"},
		{"confidential"},
		{"secret"},
	}

	machines := []string{
		"pc-01",
		"pc-02",
		"pc-03",
		"pc-04",
		"pc-05",
	}

	start := time.Date(
		2026,
		6,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < *count; i++ {

		timestamp := start.Add(
			time.Duration(i) * time.Minute,
		)

		fileName := files[rnd.Intn(len(files))]

		event := domain.Event{
			EventID:         fmt.Sprintf("evt_%d", i+1),
			Timestamp:       timestamp.Format(time.RFC3339),
			UserID:          users[rnd.Intn(len(users))],
			MachineID:       machines[rnd.Intn(len(machines))],
			Action:          actions[rnd.Intn(len(actions))],
			Channel:         channels[rnd.Intn(len(channels))],
			FileName:        fileName,
			FileExt:         fileName[len(fileName)-4:],
			ContentClasses:  contentClasses[rnd.Intn(len(contentClasses))],
			DestinationType: destinations[rnd.Intn(len(destinations))],
			Destination:     fmt.Sprintf("dest-%d", rnd.Intn(100)),
			Severity:        severities[rnd.Intn(len(severities))],
		}

		if err := encoder.Encode(event); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf(
		"Generated %d events into %s\n",
		*count,
		*out,
	)
}
