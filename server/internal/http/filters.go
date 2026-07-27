package http

import (
	"encoding/json"
	"net/http"
	"strings"
)

type DatasetFilters struct {
	UserID          []string `json:"user_id"`
	FileName        []string `json:"file_name"`
	Action          []string `json:"action"`
	DestinationType []string `json:"destination_type"`
	Channel         []string `json:"channel"`
	Severity        []string `json:"severity"`
}

func (s *Server) DatasetFiltersHandler(w http.ResponseWriter, r *http.Request) {
	Cors(w)

	if r.Method == http.MethodOptions {
		return
	}

	if !strings.HasSuffix(r.URL.Path, "/filters") {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/datasets/")
	path = strings.TrimSuffix(path, "/filters")

	events, ok := s.Datasets[path]
	if !ok {
		http.Error(w, "dataset not found", http.StatusNotFound)
		return
	}

	users := map[string]struct{}{}
	files := map[string]struct{}{}
	actions := map[string]struct{}{}
	destinations := map[string]struct{}{}
	channels := map[string]struct{}{}
	severities := map[string]struct{}{}

	for _, e := range events {
		users[e.UserID] = struct{}{}
		files[e.FileName] = struct{}{}
		actions[e.Action] = struct{}{}
		destinations[e.DestinationType] = struct{}{}
		channels[e.Channel] = struct{}{}
		severities[e.Severity] = struct{}{}
	}

	response := DatasetFilters{
		UserID:          mapKeys(users),
		FileName:        mapKeys(files),
		Action:          mapKeys(actions),
		DestinationType: mapKeys(destinations),
		Channel:         mapKeys(channels),
		Severity:        mapKeys(severities),
	}

	json.NewEncoder(w).Encode(response)
}

func mapKeys(m map[string]struct{}) []string {
	result := make([]string, 0, len(m))

	for key := range m {
		if key != "" {
			result = append(result, key)
		}
	}

	return result
}
