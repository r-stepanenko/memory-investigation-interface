package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	myhttp "event-memory-search-api/internal/http"
	nethttp "net/http"

	"event-memory-search-api/internal/datasets"
	"event-memory-search-api/internal/domain"
	"event-memory-search-api/internal/search"
)

func main() {

	if len(os.Args) > 1 {

		switch os.Args[1] {

		case "serve":
			os.Args = append([]string{os.Args[0]}, os.Args[2:]...)

		case "search":
			RunSearchCLI(os.Args[2:])
			return

		default:
			log.Fatalf("unknown command: %s", os.Args[1])
		}
	}

	config := ParseCLI()

	events, stats, err := datasets.LoadEvents(
		filepath.Join(config.DatasetsDir, "events.jsonl"),
	)
	if err != nil {
		log.Fatal(err)
	}

	if stats.Skipped > 0 {
		log.Printf(
			"Loaded control dataset: %d events (%d skipped)",
			stats.Loaded,
			stats.Skipped,
		)
	} else {
		log.Printf(
			"Loaded control dataset: %d events",
			stats.Loaded,
		)
	}

	testEvents, testStats, err := datasets.LoadEvents(
		filepath.Join(config.DatasetsDir, "testEvents.jsonl"),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(
		"Loaded test dataset: %d events (%d skipped)",
		testStats.Loaded,
		testStats.Skipped,
	)

	events100k, _ := loadOptionalDataset(
		"large100k",
		filepath.Join(config.DatasetsDir, "events100k.jsonl"),
	)

	events1m, _ := loadOptionalDataset(
		"large1m",
		filepath.Join(config.DatasetsDir, "events1m.jsonl"),
	)

	eventIndex := make(map[string]map[string]int)

	datasetsMap := map[string][]domain.Event{
		"control": events,
		"test":    testEvents,
	}

	userIndexes := map[string]*search.UserIndex{
		"control": search.BuildUserIndex(events),
		"test":    search.BuildUserIndex(testEvents),
	}

	eventIndex["control"] = buildIndexes(events)
	eventIndex["test"] = buildIndexes(testEvents)

	if len(events100k) > 0 {

		eventIndex["large100k"] =
			buildIndexes(events100k)

		datasetsMap["large100k"] = events100k

		userIndexes["large100k"] =
			search.BuildUserIndex(events100k)
	}

	if len(events1m) > 0 {

		eventIndex["large1m"] =
			buildIndexes(events1m)

		datasetsMap["large1m"] = events1m

		userIndexes["large1m"] =
			search.BuildUserIndex(events1m)
	}

	server := &myhttp.Server{
		Datasets:    datasetsMap,
		UserIndexes: userIndexes,

		Searches:   make(map[string]domain.SearchResponse),
		EventIndex: eventIndex,
	}

	nethttp.HandleFunc("/api/search", server.SearchHandler)
	nethttp.HandleFunc("/api/health", server.HealthHandler)
	nethttp.HandleFunc("/api/datasets", server.DatasetsHandler)
	nethttp.HandleFunc("/api/search/", server.SearchRouter)
	nethttp.HandleFunc("/api/events/", server.ContextHandler)
	nethttp.HandleFunc("/api/datasets/", server.DatasetFiltersHandler)

	log.Printf("Starting server on %s", config.Addr)

	httpServer := &nethttp.Server{
		Addr:         config.Addr,
		Handler:      cors(nethttp.DefaultServeMux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Fatal(httpServer.ListenAndServe())
}

func cors(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {

		w.Header().Set(
			"Access-Control-Allow-Origin",
			"http://localhost:5173",
		)

		w.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, OPTIONS",
		)

		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type",
		)

		if r.Method == nethttp.MethodOptions {
			w.WriteHeader(nethttp.StatusOK)
			return
		}

		next.ServeHTTP(w, r)

	})
}

func buildIndexes(events []domain.Event) map[string]int {

	eventIndex := make(map[string]int)

	for i, event := range events {
		eventIndex[event.EventID] = i
	}

	return eventIndex
}

func loadOptionalDataset(
	name string,
	path string,
) ([]domain.Event, *datasets.LoadStats) {

	events, stats, err := datasets.LoadEvents(path)

	if err != nil {
		log.Printf(
			"Skipping dataset %s: %v",
			name,
			err,
		)

		return nil, nil
	}

	log.Printf(
		"Loaded %s dataset: %d events (%d skipped)",
		name,
		stats.Loaded,
		stats.Skipped,
	)

	return events, &stats
}
