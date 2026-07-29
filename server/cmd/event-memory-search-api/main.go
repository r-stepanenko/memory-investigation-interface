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

	if len(os.Args) > 1 && os.Args[1] == "search" {
		RunSearchCLI(os.Args[2:])
		return
	}

	config := ParseCLI()

	events, err := datasets.LoadEvents(
		filepath.Join(config.DatasetsDir, "events.jsonl"),
	)
	if err != nil {
		log.Fatal(err)
	}

	testEvents, err := datasets.LoadEvents(
		filepath.Join(config.DatasetsDir, "testEvents.jsonl"),
	)
	if err != nil {
		log.Fatal(err)
	}

	events100k, err := datasets.LoadEvents(
		filepath.Join(config.DatasetsDir, "events100k.jsonl"),
	)
	if err != nil {
		log.Fatal(err)
	}

	events1m, err := datasets.LoadEvents(
		filepath.Join(config.DatasetsDir, "events1m.jsonl"),
	)
	if err != nil {
		log.Fatal(err)
	}

	eventMap := make(map[string]map[string]domain.Event)
	eventMap["control"] = make(map[string]domain.Event)

	for _, event := range events {
		eventMap["control"][event.EventID] = event
	}

	eventMap["test"] = make(map[string]domain.Event)

	for _, event := range testEvents {
		eventMap["test"][event.EventID] = event
	}

	eventMap["large100k"] = make(map[string]domain.Event)

	for _, event := range events100k {
		eventMap["large100k"][event.EventID] = event
	}

	eventMap["large1m"] = make(map[string]domain.Event)

	for _, event := range events1m {
		eventMap["large1m"][event.EventID] = event
	}
	eventIndex := make(map[string]int)

	for i, event := range events {
		eventIndex[event.EventID] = i
	}

	log.Printf("evt_32 in global map: %+v", eventMap["evt_32"])

	controlIndex := search.BuildUserIndex(events)
	testIndex := search.BuildUserIndex(testEvents)
	index100k := search.BuildUserIndex(events100k)
	index1m := search.BuildUserIndex(events1m)

	server := &myhttp.Server{
		Datasets: map[string][]domain.Event{
			"control":   events,
			"test":      testEvents,
			"large100k": events100k,
			"large1m":   events1m,
		},

		UserIndexes: map[string]*search.UserIndex{
			"control":   controlIndex,
			"test":      testIndex,
			"large100k": index100k,
			"large1m":   index1m,
		},

		Searches:   make(map[string]domain.SearchResponse),
		Events:     eventMap,
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
