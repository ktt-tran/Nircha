package models

import "time"

// Page is a single opportunity listing as stored in Redis by the crawler
// and read back here. This is the storage shape - it deliberately carries
// no query-dependent data (a score, a rank) since a page's own properties
// don't depend on what someone searched for; see SearchResultPage for the
// per-query wrapper used in API responses.
//
// Field names and JSON tags mirror search-engine/crawler's webpage.WebPage
// exactly (see that struct's doc comment for what each field means and
// how it's populated) - the Pages Redis hash is the single shared wire
// format between the crawler and this module, not something each side
// interprets differently.
type Page struct {
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
	Url      string `json:"url"`

	Organization  string     `json:"organization"`
	TrustTier     int        `json:"trustTier"`
	PostedAt      time.Time  `json:"postedAt"`
	LastCrawledAt time.Time  `json:"lastCrawledAt"`
	DeadlineAt    *time.Time `json:"deadlineAt,omitempty"`
	FieldsOfStudy []string   `json:"fieldsOfStudy,omitempty"`
	DegreeLevels  []string   `json:"degreeLevels,omitempty"`
	Location      string     `json:"location,omitempty"`
	Remote        bool       `json:"remote"`
}

// SearchResultPage is one entry in a search response: a Page plus the
// final composite score it received for that specific query and profile.
// Score is never stored - it only exists per-query, computed fresh by the
// ranker on every request (see pipes.Resolve / handlers.HandleSearch).
type SearchResultPage struct {
	Page
	Score float64 `json:"score"`
}