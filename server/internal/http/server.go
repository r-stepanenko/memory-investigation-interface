package http

import (
	"event-memory-search-api/internal/domain"
	"event-memory-search-api/internal/search"
	"sync"
)

type Server struct {
	Datasets    map[string][]domain.Event
	UserIndexes map[string]*search.UserIndex
	Searches    map[string]domain.SearchResponse   // результаты поиска
	EventIndex  map[string]map[string]int

	mu sync.RWMutex
}
