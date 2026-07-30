package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"event-memory-search-api/internal/datasets"
	"event-memory-search-api/internal/domain"
	"event-memory-search-api/internal/search"
)

type SearchCLIConfig struct {
	Dataset     string
	UserID      string
	Action      string
	DatasetsDir string
	EventsFile  string
	QueryFile   string
	Out         string
	Limit       int
}

type CLIResult = domain.SearchResult

func RunSearchCLI(args []string) {

	fs := flag.NewFlagSet("search", flag.ExitOnError)

	datasetName := fs.String(
		"dataset",
		"control",
		"dataset name",
	)

	datasetsDir := fs.String(
		"datasets",
		"./internal/datasets",
		"datasets directory",
	)

	eventsFile := fs.String(
		"events",
		"",
		"events jsonl file",
	)

	queryFile := fs.String(
		"query",
		"",
		"query json file",
	)

	userID := fs.String(
		"user",
		"",
		"user id",
	)

	action := fs.String(
		"action",
		"",
		"event action",
	)

	out := fs.String(
		"out",
		"",
		"output file",
	)

	limit := fs.Int(
		"limit",
		10,
		"max results",
	)

	fs.Parse(args)

	config := SearchCLIConfig{
		Dataset:     *datasetName,
		UserID:      *userID,
		Action:      *action,
		DatasetsDir: *datasetsDir,
		EventsFile:  *eventsFile,
		QueryFile:   *queryFile,
		Out:         *out,
		Limit:       *limit,
	}

	var err error
	var query CLIQuery

	if config.QueryFile != "" {

		query, err = LoadCLIQuery(
			config.QueryFile,
		)

		if err != nil {
			log.Fatal(err)
		}

		if query.DatasetID != "" {
			config.Dataset = query.DatasetID
		}
	}

	var (
		events []domain.Event
	)

	if config.EventsFile != "" {
		var stats datasets.LoadStats

		events, stats, err = datasets.LoadEvents(config.EventsFile)
		if err != nil {
			log.Fatal(err)
		}

		if stats.Skipped > 0 {
			log.Printf(
				"%s: loaded=%d skipped=%d",
				config.EventsFile,
				stats.Loaded,
				stats.Skipped,
			)
		}

	} else {

		events, err = loadCLIDataset(
			config.Dataset,
			config.DatasetsDir,
		)
	}

	if err != nil {
		log.Fatal(err)
	}

	hints := domain.SearchHints{
		UserID: config.UserID,
		Action: config.Action,
	}

	if config.QueryFile != "" {

		if query.Hints.UserID != "" {
			hints.UserID = query.Hints.UserID
		}

		if query.Hints.Action != "" {
			hints.Action = query.Hints.Action
		}

		if query.Hints.FileName != "" {
			hints.FileName = query.Hints.FileName
		}

		if query.Hints.DestinationType != "" {
			hints.DestinationType = query.Hints.DestinationType
		}

		if query.Hints.Channel != "" {
			hints.Channel = query.Hints.Channel
		}

		if query.Hints.Severity != "" {
			hints.Severity = query.Hints.Severity
		}
	}

	results := make([]CLIResult, 0)

	if config.QueryFile != "" {

		events = search.FilterByTime(
			events,
			query.Time,
		)

	}

	for _, event := range events {

		var nearbyEvents []domain.Event

		if query.Context.RequireNearby != nil {
			nearbyEvents = events
		}

		score, matched, contributions, missed := search.CalculateScore(
			event,
			hints,
			nearbyEvents,
			query.Context.RequireNearby != nil,
		)

		minScore := 0.0

		if config.QueryFile != "" {
			minScore = query.Scoring.MinScore
		}

		if score > 0 &&
			score >= minScore {

			results = append(results, CLIResult{
				Score:         score,
				MatchedHints:  matched,
				Contributions: contributions,
				MissedHints:   missed,
				Event:         event,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}

		return results[i].Event.Timestamp > results[j].Event.Timestamp
	})

	resultLimit := config.Limit

	if config.QueryFile != "" &&
		query.Scoring.Limit > 0 {

		resultLimit = query.Scoring.Limit

	}

	if resultLimit > 0 && len(results) > resultLimit {
		results = results[:resultLimit]
	}

	data, err := json.MarshalIndent(
		results,
		"",
		"  ",
	)

	if err != nil {
		log.Fatal(err)
	}

	if config.Out != "" {

		err := os.WriteFile(
			config.Out,
			data,
			0644,
		)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("saved:", config.Out)

	} else {

		fmt.Println(string(data))

	}
}

func loadCLIDataset(name string, dir string) ([]domain.Event, error) {

	switch name {

	case "control":
		events, _, err := datasets.LoadEvents(
			dir + "/events.jsonl",
		)
		return events, err

	case "test":
		events, _, err := datasets.LoadEvents(
			dir + "/testEvents.jsonl",
		)
		return events, err

	default:
		return nil, fmt.Errorf(
			"unknown dataset: %s",
			name,
		)
	}
}
