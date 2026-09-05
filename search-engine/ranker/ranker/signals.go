package ranker

import (
	"math"
	"strings"
	"time"

	"search-engine/ranker/document"
)

// normalizeBM25 min-max normalizes raw relevance scores to [0,1] across the
// current candidate set.
func normalizeBM25(candidates []Candidate) []float64 {
	normalized := make([]float64, len(candidates))
	if len(candidates) == 0 {
		return normalized
	}

	lo, hi := candidates[0].BM25Score, candidates[0].BM25Score
	for _, c := range candidates {
		if c.BM25Score < lo {
			lo = c.BM25Score
		}
		if c.BM25Score > hi {
			hi = c.BM25Score
		}
	}

	spread := hi - lo
	for i, c := range candidates {
		if spread == 0 {
			// Every candidate scored identically (often just a single hit,
			// or a query that matched everything equally well). Treat them
			// as equally, maximally relevant rather than dividing by zero.
			normalized[i] = 1
			continue
		}
		normalized[i] = (c.BM25Score - lo) / spread
	}
	return normalized
}

func tokenize(s string) []string {
	return strings.Fields(strings.ToLower(s))
}

func containsToken(tokens []string, term string) bool {
	for _, t := range tokens {
		if t == term {
			return true
		}
	}
	return false
}

// fieldMatchScore rewards listings whose Title, Tags, or Organization
// directly contain the query's terms.
func fieldMatchScore(queryTerms []string, l document.Listing) float64 {
	if len(queryTerms) == 0 {
		return 0
	}

	titleTokens := tokenize(l.Title)
	orgTokens := tokenize(l.Organization)
	tagTokens := tokenize(strings.Join(l.Tags, " "))

	var titleHits, orgHits, tagHits int
	for _, term := range queryTerms {
		if containsToken(titleTokens, term) {
			titleHits++
		}
		if containsToken(orgTokens, term) {
			orgHits++
		}
		if containsToken(tagTokens, term) {
			tagHits++
		}
	}

	n := float64(len(queryTerms))
	titleScore := float64(titleHits) / n
	tagScore := float64(tagHits) / n
	orgScore := float64(orgHits) / n

	combined := 0.6*titleScore + 0.25*tagScore + 0.15*orgScore
	if combined > 1 {
		combined = 1
	}
	return combined
}

// fieldOfStudyMatchScore compares a person's declared major against a
// listing's FieldsOfStudy. It stays neutral (0.5) whenever there's not
// enough information to say anything either way: no profile
// major given, or a listing that doesn't state one. An explicit mismatch
// (both sides known, no overlap) scores low but not zero.
func fieldOfStudyMatchScore(major string, l document.Listing) float64 {
	const (
		neutral   = 0.5
		mismatch  = 0.15
		fullMatch = 1.0
	)

	major = strings.ToLower(strings.TrimSpace(major))
	if major == "" || len(l.FieldsOfStudy) == 0 {
		return neutral
	}

	majorTokens := tokenize(major)
	for _, fos := range l.FieldsOfStudy {
		fosTokens := tokenize(fos)
		for _, term := range majorTokens {
			if containsToken(fosTokens, term) {
				return fullMatch
			}
		}
	}
	return mismatch
}

// trustTierScore normalizes a listing's manually-assigned trust tier into
// [0,1], where 1.0 is most trusted (TrustTierPrimary) and 0.0 is the least
// trusted tier in use (Config.MaxTrustTier).
func trustTierScore(tier document.TrustTier, maxTier document.TrustTier) float64 {
	if tier < 1 {
		tier = 1
	}
	if tier > maxTier {
		tier = maxTier
	}
	if maxTier <= 1 {
		return 1
	}
	return 1 - float64(tier-1)/float64(maxTier-1)
}

// freshnessScore applies exponential decay based on how long ago a listing
// effectively went live (Listing.EffectiveFreshnessTime). The score is 1.0
// the moment a listing appears, decays to 0.5 after halfLife, and
// asymptotically approaches 0 as the listing ages further. A listing with
// no timestamp at all scores neutrally (0.5) rather than being penalized
// for an ingestion gap that isn't the listing's fault.
func freshnessScore(l document.Listing, now time.Time, halfLife time.Duration) float64 {
	effective := l.EffectiveFreshnessTime()
	if effective.IsZero() {
		return 0.5
	}
	if halfLife <= 0 {
		return 1
	}

	age := now.Sub(effective)
	if age < 0 {
		age = 0
	}
	return math.Pow(0.5, age.Hours()/halfLife.Hours())
}

// deadlineUrgencyScore rewards listings whose deadline is close at hand.
// Listings with no deadline get a neutral default rather than being
// penalized for not having one. A deadline further out than window isn't
// urgent yet and also scores 0. An already-passed deadline scores 0 as a
// defensive fallback.
func deadlineUrgencyScore(l document.Listing, now time.Time, window time.Duration, noDeadlineDefault float64) float64 {
	if l.DeadlineAt == nil {
		return noDeadlineDefault
	}

	remaining := l.DeadlineAt.Sub(now)
	if remaining <= 0 {
		return 0
	}
	if window <= 0 || remaining >= window {
		return 0
	}
	return 1 - remaining.Hours()/window.Hours()
}