package http

import (
	"encoding/json"
	"event-memory-search-api/internal/domain"
	"event-memory-search-api/internal/search"
	"net/http"
	"sort"
	"strings"
	"time"
)

func (s *Server) SearchHandler(w http.ResponseWriter, r *http.Request) {
	Cors(w)

	if r.Method == http.MethodOptions {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req domain.SearchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"bad json",
			http.StatusBadRequest,
		)
		return
	}

	events, ok := s.Datasets[req.DatasetID]

	if !ok {
		http.Error(
			w,
			"dataset not found",
			http.StatusNotFound,
		)
		return
	}

	results := make([]domain.SearchResult, 0)

	for _, event := range events {
		score := 0

		matched := make([]string, 0)

		userScore := search.MatchScore(
			event.UserID,
			req.Hints.UserID,
		)

		if userScore > 0 {
			score += userScore

			matched = append(
				matched,
				"user_id",
			)
		}

		fileScore := search.MatchScore(
			event.FileName,
			req.Hints.FileName,
		)

		if fileScore > 0 {
			score += fileScore

			matched = append(
				matched,
				"file_name",
			)
		}

		actionScore := search.MatchScore(
			event.Action,
			req.Hints.Action,
		)

		if actionScore > 0 {
			score += actionScore

			matched = append(
				matched,
				"action",
			)
		}

		if req.Hints.DestinationType != "" {
			if strings.EqualFold(
				event.DestinationType,
				req.Hints.DestinationType,
			) {

				score += 20

				matched = append(
					matched,
					"destination_type",
				)
			}
		}

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
				nearbyEvents := search.FindNearbyEvents(
					events,
					event,
					before,
					after,
					actions,
				)

				if len(nearbyEvents) > 0 {
					score += 10
					matched = append(
						matched,
						"nearby event found",
					)
				}
			}
		}

		if score > 0 {
			results = append(
				results,
				domain.SearchResult{
					Score:        score,
					MatchedHints: matched,
					Event:        event,
				},
			)
		}
	}

	//сначала самые подходящие
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
	s.Searches[searchID] = resp
	json.NewEncoder(w).Encode(resp)
}
