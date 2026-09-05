package webpage

import "time"

// WebPage is a single crawled page. Fields through OutboundLinks are the
// original crawl-time content used for indexing and link-graph tracking;
// everything from Organization onward is opportunity-specific structured
// data extracted for the ranker.
type WebPage struct {
	URL           string   `json:"url"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Words         []Word   `json:"words"`
	Links         []string `json:"links"`
	Abstract      string   `json:"abstract"`
	OutboundLinks []string `json:"outbound_links"`
	Organization string `json:"organization"` 		// Organization is who's offering the opportunity
	// TrustTier is NOT extracted from page content. It's assigned from the
	// crawler's seed/domain configuration at crawl time and stamped onto 
	// every page crawled from a given domain.
	TrustTier int `json:"trustTier"`
	PostedAt time.Time `json:"postedAt"` 			// PostedAt is when the opportunity was first published
	LastCrawledAt time.Time `json:"lastCrawledAt"` 	// LastCrawledAt is when this page was fetched
	DeadlineAt *time.Time `json:"deadlineAt,omitempty"`
	FieldsOfStudy []string `json:"fieldsOfStudy,omitempty"`
	DegreeLevels []string `json:"degreeLevels,omitempty"`
	Location string `json:"location,omitempty"`
	Remote bool `json:"remote"`
}

type Word struct {
	Word     string `json:"word"`
	Score    int    `json:"score"`
	Position int    `json:"position"`
}