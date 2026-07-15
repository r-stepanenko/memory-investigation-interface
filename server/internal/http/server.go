package http

import "event-memory-search-api/internal/domain"

type Server struct {
	Datasets map[string][]domain.Event
	Searches map[string]domain.SearchResponse // результаты поиска
	Events map[string]domain.Event // быстрый поиск события по event_id
	EventIndex map[string]int
}