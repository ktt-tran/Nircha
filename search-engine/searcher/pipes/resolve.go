package pipes

import (
	"context"
	"strconv"

	"search-engine/ranker/document"
	"search-engine/ranker/ranker"
	"search-engine/searcher/models"
)

// PageRepository is the subset of repository.RedisRepository Resolve
// needs - defined here, at the point of use, rather than depending on the
// concrete repository package directly. Keeps pipes testable against a
// fake and avoids coupling it to one specific storage backend.
type PageRepository interface {
	GetPages(ctx context.Context, pageIds []string) map[string]models.Page
}

// Resolve turns a merged candidate set into ranker.Candidates ready to
// score: fetches every RawPage's full stored Page data in a single batch
// round trip (see PageRepository.GetPages), then pairs each one with a
// relevance score derived from its aggregate hit statistics (see
// relevanceScore).
func Resolve(ctx context.Context, repo PageRepository, rawPages *[]models.RawPage) []ranker.Candidate {
	ids := make([]string, len(*rawPages))
	byID := make(map[string]models.RawPage, len(*rawPages))
	for i, rp := range *rawPages {
		id := strconv.FormatInt(rp.PageId, 10)
		ids[i] = id
		byID[id] = rp
	}

	pages := repo.GetPages(ctx, ids)

	candidates := make([]ranker.Candidate, 0, len(pages))
	for id, page := range pages {
		candidates = append(candidates, ranker.Candidate{
			Listing:   toListing(id, page),
			BM25Score: relevanceScore(byID[id]),
		})
	}
	return candidates
}

// relevanceScore turns a RawPage's aggregate hit statistics into the
// single comparable number ranker.Candidate.BM25Score expects (see that
// field's doc comment for why "BM25" here doesn't mean literal BM25
// math). HitsCount - how many distinct query tokens matched - dominates:
// matching more of the query is the strongest relevance signal available
// here. HitsSum - the sum of each match's HTML-tag-based importance score
// (see the crawler's extractWords) - breaks ties within that.
//
// This is a weighted-sum approximation of the multi-key sort it replaces
// (HitsCount, then PositionalDistance, then HitsSum, then a stored Rank
// value), not an exact equivalent - no single weighted sum can reproduce
// a true lexicographic order in every case, and it doesn't need to:
// normalizeBM25 (in the ranker package) only needs the relative order
// this induces to closely track the original intent, not match it
// key-for-key. PositionalDistance is deliberately dropped rather than
// folded in as a third term - it was a minor tie-breaker in the old
// scheme, and the composite ranker's field-match signal already rewards
// compact, prominent matches in its own way. The Rank key is dropped too:
// it only ever reflected the deferred PageRank job's output (see
// document.TrustTier's doc comment in the ranker module for why that was
// set aside), which TrustTier now covers as this project's actual
// authority signal.
func relevanceScore(rp models.RawPage) float64 {
	return float64(rp.HitsCount)*100 + float64(rp.HitsSum)
}

// toListing adapts a stored Page into the ranker's document.Listing. This
// is the one place a Redis page record and the ranker's domain type meet
// - see models.Page's doc comment for why their fields already line up
// directly. FieldsOfStudy and DegreeLevels are folded into Tags too, on
// top of populating their own dedicated Listing fields: they're exactly
// the kind of short, topical labels the ranker's field-match signal
// checks title/tags/organization for, so a query like "biology research"
// gets credit there as well as through plain text relevance.
func toListing(id string, page models.Page) document.Listing {
	tags := make([]string, 0, len(page.FieldsOfStudy)+len(page.DegreeLevels))
	tags = append(tags, page.FieldsOfStudy...)
	tags = append(tags, page.DegreeLevels...)

	return document.Listing{
		ID:            id,
		Title:         page.Title,
		Organization:  page.Organization,
		Description:   page.Abstract,
		Tags:          tags,
		URL:           page.Url,
		TrustTier:     document.TrustTier(page.TrustTier),
		PostedAt:      page.PostedAt,
		LastCrawledAt: page.LastCrawledAt,
		DeadlineAt:    page.DeadlineAt,
		FieldsOfStudy: page.FieldsOfStudy,
		DegreeLevels:  page.DegreeLevels,
		Location:      page.Location,
		Remote:        page.Remote,
	}
}