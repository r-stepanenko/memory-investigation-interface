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
	events, err := datasets.LoadEvents("internal/datasets/events.jsonl")
	if err != nil {
		log.Fatal(err)
	}

	testEvents, err := datasets.LoadEvents("internal/datasets/testEvents.jsonl")
	if err != nil {
		log.Fatal(err)
	}

	eventMap := make(map[string]domain.Event)
	eventIndex := make(map[string]int)

	allEvents := append(events, testEvents...)

	for i, event := range allEvents {
		eventMap[event.EventID] = event
		eventIndex[event.EventID] = i
	}

	server := &myhttp.Server{
		Datasets: map[string][]domain.Event{
			"control": events,
			"test":    testEvents,
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
		Addr:         ":8080",
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
