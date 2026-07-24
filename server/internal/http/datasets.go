package http

import (
	"encoding/json"
	"net/http"

	"event-memory-search-api/internal/domain"
)

type DatasetInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        int    `json:"size"`
	Period      string `json:"period"`
	Description string `json:"description"`
}

func (s *Server) DatasetsHandler(w http.ResponseWriter, r *http.Request) {
	Cors(w)

	if r.Method == http.MethodOptions {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	result := make([]DatasetInfo, 0)

	for name, events := range s.Datasets {

		result = append(result, DatasetInfo{
			ID:          name,
			Name:        name,
			Size:        len(events),
			Period:      getDatasetPeriod(events),
			Description: "Набор событий пользователей",
		})
	}

	json.NewEncoder(w).Encode(result)
}

func getDatasetPeriod(events []domain.Event) string {
	if len(events) == 0 {
		return "нет данных"
	}

	min := events[0].Timestamp
	max := events[0].Timestamp

	for _, event := range events {

		if event.Timestamp < min {
			min = event.Timestamp
		}

		if event.Timestamp > max {
			max = event.Timestamp
		}
	}

	return min + " — " + max
}