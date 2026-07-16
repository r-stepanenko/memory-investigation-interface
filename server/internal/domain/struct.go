package domain

type Event struct {
	EventID         string   `json:"event_id"`
	Timestamp       string   `json:"timestamp"`
	UserID          string   `json:"user_id"`
	MachineID       string   `json:"machine_id"`
	Action          string   `json:"action"`
	Channel         string   `json:"channel"`
	FileName        string   `json:"file_name"`
	FileExt         string   `json:"file_ext"`
	ContentClasses  []string `json:"content_classes"`
	DestinationType string   `json:"destination_type"`
	Destination     string   `json:"destination"`
	Severity        string   `json:"severity"`
}

type SearchHints struct {
	UserID          string `json:"user_id"`
	FileName        string `json:"file_name"`
	Action          string `json:"action"`
	DestinationType string `json:"destination_type"`
}

type NearbyRule struct {
	Action string `json:"action"`
}

// контекст поиска
type SearchContext struct {
	Before        string       `json:"before"`
	After         string       `json:"after"`
	RequireNearby []NearbyRule `json:"require_nearby"`
}

// запрос поиска
type SearchRequest struct {
	DatasetID string        `json:"dataset_id"`
	Hints     SearchHints   `json:"hints"`
	Context   SearchContext `json:"context"`
}

// один кандидат
type SearchResult struct {
	Score         int            `json:"score"`
	MatchedHints  []string       `json:"matched_hints"`
	Contributions []Contribution `json:"-"`
	Event         Event          `json:"event"`
}

// ответ поиска
type SearchResponse struct {
	SearchID        string         `json:"search_id"`
	Status          string         `json:"status"`
	DatasetID       string         `json:"dataset_id"`
	TotalCandidates int            `json:"total_candidates"`
	Candidates      []SearchResult `json:"candidates"`
}

// контекст события
type EventContext struct {
	Event  Event   `json:"event"`
	Before []Event `json:"before"`
	After  []Event `json:"after"`
}

type Contribution struct {
	Hint   string `json:"hint"`
	Type   string `json:"type"`
	Value  string `json:"value"`
	Query  string `json:"query,omitempty"`
	Points int    `json:"points"`
}

type ExplainResponse struct {
	SearchID      string         `json:"search_id"`
	EventID       string         `json:"event_id"`
	Score         int            `json:"score"`
	Contributions []Contribution `json:"contributions"`
}
