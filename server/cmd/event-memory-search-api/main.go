package main

import (
	"log"
	"time"

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
		Searches:   make(map[string]domain.SearchResponse),
		Events:     eventMap,
		EventIndex: eventIndex,
	}

	nethttp.HandleFunc("/api/search", server.SearchHandler)
	nethttp.HandleFunc("/api/health", server.HealthHandler)
	nethttp.HandleFunc("/api/datasets", server.DatasetsHandler)
	nethttp.HandleFunc("/api/search/", server.SearchRouter)
	nethttp.HandleFunc("/api/events/", server.ContextHandler)

	log.Println("Starting server on :8080")
	httpServer := &nethttp.Server{
		Addr: ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	err := httpServer.ListenAndServe()
	log.Fatal(err)
}
