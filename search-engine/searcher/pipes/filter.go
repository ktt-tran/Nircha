package pipes

import (
	"search-engine/ranker/document"
	"search-engine/ranker/ranker"
)

// Filter applies the hard-filter half of profile-based personalization -
// degree level and location eligibility - to an already-resolved
// candidate set (see Resolve).
// profile may be the zero value (document.Profile{}) for an anonymous,
// profile-less search, in which case every candidate passes through
// unchanged.
func Filter(candidates []ranker.Candidate, profile document.Profile) []ranker.Candidate {
	filtered := make([]ranker.Candidate, 0, len(candidates))
	for _, c := range candidates {
		if c.Listing.MatchesProfile(profile) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}