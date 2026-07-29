package main

import (
	"encoding/json"
	"os"

	"event-memory-search-api/internal/domain"
)

type CLIQuery struct {
	DatasetID string `json:"dataset_id"`
	Time domain.TimeFilter `json:"time"`
	Hints domain.SearchHints `json:"hints"`
	Context domain.SearchContext `json:"context"`
	Scoring domain.Scoring `json:"scoring"`
}

func LoadCLIQuery(path string) (CLIQuery, error) {

	data, err := os.ReadFile(path)

	if err != nil {
		return CLIQuery{}, err
	}

	var query CLIQuery

	err = json.Unmarshal(
		data,
		&query,
	)

	return query, err
}
