package main

import (
	"log"

	myhttp "event-memory-search-api/internal/http"
	nethttp "net/http"

	"event-memory-search-api/internal/datasets"
	"event-memory-search-api/internal/domain"
)

func main() {
	events := datasets.LoadEvents("internal/datasets/events.jsonl")

	eventMap := make(map[string]domain.Event)
	eventIndex := make(map[string]int)

	for i, event := range events {
		eventMap[event.EventID] = event
		eventIndex[event.EventID] = i
	}

	server := &myhttp.Server{
		Datasets: map[string][]domain.Event{
			"control": events,
		},
		Searches: make(map[string]domain.SearchResponse),
		Events:   eventMap,
		EventIndex: eventIndex,
	}

	nethttp.HandleFunc("/api/search", server.SearchHandler)
	nethttp.HandleFunc("/api/health", server.HealthHandler)
	nethttp.HandleFunc("/api/datasets", server.DatasetsHandler)
	nethttp.HandleFunc("/api/search/", server.SearchResultHandler)
	nethttp.HandleFunc("/api/events/", server.ContextHandler)

	log.Println("Starting server on :8080")
	err := nethttp.ListenAndServe(":8080", nil)
	log.Fatal(err)
}
