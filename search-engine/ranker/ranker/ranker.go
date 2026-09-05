// Package ranker implements the composite ranking formula for the
// opportunity search engine: a single weighted combination of normalized
// BM25 relevance, field match, field-of-study match, trust tier, freshness
// decay, and deadline urgency. It replaces three-pass sequential ranking
// and PageRank (see pagerank.go) with one formula that's straightforward to
// reason about and debug alone.
package ranker

import (
	"sort"
	"time"

	"search-engine/ranker/document"
)

type ScoreBreakdown struct {
	BM25              float64
	FieldMatch        float64
	FieldOfStudyMatch float64
	TrustTier         float64
	Freshness         float64
	DeadlineUrgency   float64
	Final             float64
}

// Result pairs a listing with its score breakdown, ready to sort and serve.
type Result struct {
	Listing   document.Listing
	Breakdown ScoreBreakdown
}

// Score combines one candidate's already-normalized BM25 score with its
// other signals into a single weighted value using cfg.Weights.
func Score(cfg Config, candidate Candidate, normalizedBM25 float64, queryTerms []string, profile document.Profile, now time.Time) ScoreBreakdown {
	b := ScoreBreakdown{
		BM25:              normalizedBM25,
		FieldMatch:        fieldMatchScore(queryTerms, candidate.Listing),
		FieldOfStudyMatch: fieldOfStudyMatchScore(profile.Major, candidate.Listing),
		TrustTier:         trustTierScore(candidate.Listing.TrustTier, cfg.MaxTrustTier),
		Freshness:         freshnessScore(candidate.Listing, now, cfg.FreshnessHalfLife),
		DeadlineUrgency:   deadlineUrgencyScore(candidate.Listing, now, cfg.UrgencyWindow, cfg.NoDeadlineUrgency),
	}
	w := cfg.Weights
	b.Final = w.BM25*b.BM25 +
		w.FieldMatch*b.FieldMatch +
		w.FieldOfStudyMatch*b.FieldOfStudyMatch +
		w.TrustTier*b.TrustTier +
		w.Freshness*b.Freshness +
		w.DeadlineUrgency*b.DeadlineUrgency
	return b
}

// Rank scores every candidate in an already-retrieved, already-filtered
// candidate set and returns them sorted best-first. queryString is
// tokenized once and reused for every candidate's field-match signal; raw
// BM25 scores are min-max normalized across candidates before weighting.
// profile is the optional, frontend-collected major/degree-level/location
// profile.
func Rank(cfg Config, queryString string, candidates []Candidate, profile document.Profile, now time.Time) []Result {
	queryTerms := tokenize(queryString)
	normalizedScores := normalizeBM25(candidates)

	results := make([]Result, len(candidates))
	for i, c := range candidates {
		results[i] = Result{
			Listing:   c.Listing,
			Breakdown: Score(cfg, c, normalizedScores[i], queryTerms, profile, now),
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Breakdown.Final > results[j].Breakdown.Final
	})
	return results
}