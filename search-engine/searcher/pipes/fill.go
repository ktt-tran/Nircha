package pipes

import (
	"search-engine/ranker/ranker"
	"search-engine/searcher/models"
)

// Fill slices ranker.Rank's already-sorted results down to one page and
// converts each surviving entry to the API's response shape.
func Fill(results []ranker.Result, size, page int) []models.SearchResultPage {
	out := make([]models.SearchResultPage, 0)

	if size <= 0 || page <= 0 {
		return out
	}

	start := (page - 1) * size
	if start >= len(results) {
		return out
	}

	end := start + size
	if end > len(results) {
		end = len(results)
	}

	for i := start; i < end; i++ {
		r := results[i]
		out = append(out, models.SearchResultPage{
			Page: models.Page{
				Title:         r.Listing.Title,
				Abstract:      r.Listing.Description,
				Url:           r.Listing.URL,
				Organization:  r.Listing.Organization,
				TrustTier:     int(r.Listing.TrustTier),
				PostedAt:      r.Listing.PostedAt,
				LastCrawledAt: r.Listing.LastCrawledAt,
				DeadlineAt:    r.Listing.DeadlineAt,
				FieldsOfStudy: r.Listing.FieldsOfStudy,
				DegreeLevels:  r.Listing.DegreeLevels,
				Location:      r.Listing.Location,
				Remote:        r.Listing.Remote,
			},
			Score: r.Breakdown.Final,
		})
	}

	return out
}