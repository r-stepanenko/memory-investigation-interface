package http

import (
	"encoding/json"
	"event-memory-search-api/internal/domain"
	"event-memory-search-api/internal/search"
	"fmt"
	"net/http"
	"sort"
	"strings"
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

	fmt.Printf("%+v\n", req.Hints)

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

	if req.Hints.UserID != "" {
		if index, ok := s.UserIndexes[req.DatasetID]; ok {
			indexed := index.Find(req.Hints.UserID)
			if len(indexed) > 0 {
				events = indexed
			}
		}
	}
	
	if (req.Time.Around == "") != (req.Time.Tolerance == "") {
		WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_TIME",
			"time.around and time.tolerance must be specified together",
		)
		return
	}

	if (req.Context.Before == "") != (req.Context.After == "") {
		WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_CONTEXT",
			"context.before and context.after must be specified together",
		)
		return
	}

	var (
		before       time.Duration
		after        time.Duration
		nearbyWithin time.Duration
		actions      []string
	)

	if req.Context.Before != "" {
		var err error

		before, err = search.ParseTolerance(req.Context.Before)
		if err != nil {
			WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_DURATION",
				"context.before must be duration",
			)
			return
		}

		after, err = search.ParseTolerance(req.Context.After)
		if err != nil {
			WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_DURATION",
				"context.after must be duration",
			)
			return
		}
	}

	if len(req.Context.RequireNearby) > 0 {

		for _, rule := range req.Context.RequireNearby {

			if strings.TrimSpace(rule.Action) == "" {
				WriteError(
					w,
					http.StatusBadRequest,
					"INVALID_CONTEXT",
					"nearby action is required",
				)
				return
			}

			actions = append(actions, rule.Action)

			if rule.Within != "" {

				duration, err := search.ParseTolerance(rule.Within)

				if err != nil {
					WriteError(
						w,
						http.StatusBadRequest,
						"INVALID_DURATION",
						"nearby.within must be duration",
					)
					return
				}

				if duration <= 0 {
					WriteError(
						w,
						http.StatusBadRequest,
						"INVALID_CONTEXT",
						"nearby duration must be positive",
					)
					return
				}

				nearbyWithin = duration
			}
		}
	}

	if req.Time.Around != "" {

		_, err := time.Parse(
			time.RFC3339,
			req.Time.Around,
		)

		if err != nil {
			WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_TIME",
				"time.around must be RFC3339",
			)
			return
		}
	}

	if req.Time.Tolerance != "" {

		_, err := search.ParseTolerance(
			req.Time.Tolerance,
		)

		if err != nil {
			WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_DURATION",
				"time.tolerance must be duration",
			)
			return
		}
	}

	results := make([]domain.SearchResult, 0)

	for _, event := range events {

		if !search.MatchTime(
			event.Timestamp,
			req.Time.Around,
			req.Time.Tolerance,
		) {
			continue
		}

		baseScore, _, _, _ := search.CalculateScore(
			event,
			req.Hints,
			nil,
			false,
		)

		if baseScore == 0 && len(actions) == 0 {
			continue
		}

		var nearby []domain.Event
		nearbyMatched := false

		if len(actions) > 0 {

			nearbyBefore := before
			nearbyAfter := after

			if nearbyBefore == 0 && nearbyAfter == 0 {
				nearbyBefore = nearbyWithin
				nearbyAfter = nearbyWithin
			}

			if nearbyWithin > 0 {
				if nearbyBefore == 0 || nearbyWithin < nearbyBefore {
					nearbyBefore = nearbyWithin
				}

				if nearbyAfter == 0 || nearbyWithin < nearbyAfter {
					nearbyAfter = nearbyWithin
				}
			}

			nearby = search.FindNearbyEvents(
				events,
				event,
				nearbyBefore,
				nearbyAfter,
				actions,
			)

			foundActions := make(map[string]bool)

			for _, e := range nearby {

				if e.EventID == event.EventID {
					continue
				}

				foundActions[e.Action] = true
			}

			missing := false

			for _, action := range actions {

				if !foundActions[action] {
					missing = true
					break
				}

			}

			if missing {
				continue
			}

			nearbyMatched = true
		}

		score, matched, contributions, missedHints := search.CalculateScore(
			event,
			req.Hints,
			nearby,
			nearbyMatched,
		)

		if score > 100 {
			score = 100
		}

		if score > 0 {

			if req.Scoring.MinScore > 0 &&
				score < req.Scoring.MinScore {
				continue
			}

			results = append(
				results,
				domain.SearchResult{
					Score:         score,
					MatchedHints:  matched,
					Contributions: contributions,
					MissedHints:   missedHints,
					Event:         event,
				},
			)
		}
	}

	sort.SliceStable(results, func(i, j int) bool {

		if results[i].Score == results[j].Score {
			return results[i].Event.Timestamp > results[j].Event.Timestamp
		}

		return results[i].Score > results[j].Score
	})

	searchID := search.NewSearchID()
	totalCandidates := len(results)

	if req.Scoring.Limit > 0 &&
		len(results) > req.Scoring.Limit {

		results = results[:req.Scoring.Limit]
	}

	resp := domain.SearchResponse{
		SearchID:        searchID,
		Status:          "done",
		DatasetID:       req.DatasetID,
		TotalCandidates: totalCandidates,
		Candidates:      results,
	}

	s.mu.Lock()
	s.Searches[searchID] = resp
	s.mu.Unlock()

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return
	}
}
