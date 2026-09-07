package models

// SearchResponse is the JSON body returned by GET /api/search.
type SearchResponse struct {
	Query        string             `json:"query"`         // the search terms that were echoed back
	Pages        []SearchResultPage `json:"pages"`         // the ranked results for this one page, ready to render
	TotalResults int                `json:"totalResults"`  // how many results matched in total, across all pages
	Page         int                `json:"page"`          // which page this response is (1-indexed)
	PageSize     int                `json:"pageSize"`      // results per page for this response
	PagesCount   int                `json:"pagesCount"`    // total pages available: ceil(TotalResults / PageSize)
	TimeMs       int64              `json:"timeMs"`        // how long the search took server-side, in milliseconds
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