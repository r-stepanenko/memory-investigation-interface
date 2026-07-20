package http

import (
	"encoding/json"
	"event-memory-search-api/internal/domain"
	"event-memory-search-api/internal/search"
	"net/http"
	"sort"
	"time"
)

func (s *Server) SearchHandler(w http.ResponseWriter, r *http.Request) {
	Cors(w)

	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		WriteError(
			w,
			http.StatusMethodNotAllowed,
			"METHOD_NOT_ALLOWED",
			"method not allowed",
		)
		return
	}

	var req domain.SearchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"bad json",
		)
		return
	}

	if req.DatasetID == "" {
		WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"dataset_id is required",
		)
		return
	}

	events, ok := s.Datasets[req.DatasetID]

	if !ok {
		WriteError(
			w,
			http.StatusNotFound,
			"DATASET_NOT_FOUND",
			"dataset not found",
		)
		return
	}

	results := make([]domain.SearchResult, 0)

	for _, event := range events {

		score, matched, contributions := search.CalculateScore(
			event,
			req.Hints,
		)

		if len(req.Context.RequireNearby) > 0 {

			before, err1 := time.ParseDuration(
				req.Context.Before,
			)

			after, err2 := time.ParseDuration(
				req.Context.After,
			)

			if err1 == nil && err2 == nil {

				actions := make([]string, 0)

				for _, rule := range req.Context.RequireNearby {
					actions = append(
						actions,
						rule.Action,
					)
				}

				nearby := search.FindNearbyEvents(
					events,
					event,
					before,
					after,
					actions,
				)

				if len(nearby) > 0 {

					score += 10

					matched = append(
						matched,
						"nearby event found",
					)

					contributions = append(
						contributions,
						domain.Contribution{
							Hint:   "nearby",
							Type:   "context",
							Value:  nearby[0].Action,
							Points: 10,
						},
					)
				}
			}
		}

		if score > 100 {
			score = 100
		}

		if score > 0 {

			results = append(
				results,
				domain.SearchResult{
					Score:         score,
					MatchedHints:  matched,
					Contributions: contributions,
					Event:         event,
				},
			)
		}
	}

	sort.Slice(
		results,
		func(i, j int) bool {
			return results[i].Score > results[j].Score
		},
	)

	searchID := search.NewSearchID()

	resp := domain.SearchResponse{
		SearchID:        searchID,
		Status:          "done",
		DatasetID:       req.DatasetID,
		TotalCandidates: len(results),
		Candidates:      results,
	}

	s.mu.Lock()
	s.Searches[searchID] = resp
	s.mu.Unlock()

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return
	}
}
