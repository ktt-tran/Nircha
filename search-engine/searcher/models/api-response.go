package models

// SearchResponse is the JSON body returned by GET /api/search.
type SearchResponse struct {
	Query        string             `json:"query"`
	Pages        []SearchResultPage `json:"pages"`
	TotalResults int                `json:"totalResults"`
	Page         int                `json:"page"`
	PageSize     int                `json:"pageSize"`
	PagesCount   int                `json:"pagesCount"`
	TimeMs       int64              `json:"timeMs"`
}

// SuggestResponse is the JSON body returned by GET /api/suggest. Suggestions
// maps each processed query token to its candidate spelling corrections
// (closest edit distance first), for tokens that had any. A token that was
// already an exact index match, or had no reasonably close correction, is
// simply absent from the map.
type SuggestResponse struct {
	Query       string              `json:"query"`
	Suggestions map[string][]string `json:"suggestions"`
}

// HealthResponse is the JSON body returned by GET /api/health.
type HealthResponse struct {
	Status string `json:"status"`
	Redis  string `json:"redis"`
}

// ErrorResponse is the JSON body returned for any non-2xx response.
type ErrorResponse struct {
	Error string `json:"error"`
}