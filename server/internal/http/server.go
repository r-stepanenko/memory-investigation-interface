package http

import (
	"event-memory-search-api/internal/domain"
	"event-memory-search-api/internal/search"
	"sync"
)

type Server struct {
	Datasets    map[string][]domain.Event
	UserIndexes map[string]*search.UserIndex
	Searches    map[string]domain.SearchResponse // результаты поиска
	Events      map[string]domain.Event          // быстрый поиск события по event_id
	EventIndex  map[string]int

	mu sync.RWMutex
}
