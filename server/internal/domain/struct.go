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

type TimeFilter struct {
	Around    string `json:"around"`
	Tolerance string `json:"tolerance"`
}

// запрос поиска
type Scoring struct {
	MinScore float64 `json:"min_score"`
	Limit    int     `json:"limit"`
}

type SearchRequest struct {
	DatasetID string        `json:"dataset_id"`
	Time      TimeFilter    `json:"time"`
	Hints     SearchHints   `json:"hints"`
	Context   SearchContext `json:"context"`
	Scoring   Scoring       `json:"scoring"`
}

// один кандидат
type SearchResult struct {
	Score         float64        `json:"score"`
	MatchedHints  []string       `json:"matched_hints"`
	Contributions []Contribution `json:"-"`
	MissedHints   []MissedHint   `json:"-"`
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
	Hint    string  `json:"hint"`
	Type    string  `json:"type"`
	Value   string  `json:"value"`
	Query   string  `json:"query,omitempty"`
	Points  float64 `json:"points"`
	Matched bool    `json:"matched"`
	Reason  string  `json:"reason"`
}

type MissedHint struct {
	Hint   string `json:"hint"`
	Reason string `json:"reason"`
}

type ExplainResponse struct {
	SearchID      string         `json:"search_id"`
	EventID       string         `json:"event_id"`
	Score         float64        `json:"score"`
	Contributions []Contribution `json:"contributions"`
	MissedHints   []MissedHint   `json:"missed_hints,omitempty"`
}
