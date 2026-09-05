package document

import "strings"

// Profile is the optional set of criteria the frontend collects about a
// person's: major, degree level, and location and so results can be
// filtered and ranked for them specifically. Every field is optional: the
// zero value means "not specified," and every check below is a no-op when
// its corresponding field is empty. A caller with no profile at all can
// just pass the zero-value Profile{}.
type Profile struct {
	Major       string
	DegreeLevel string
	Location    string
}

// MatchesProfile applies the hard-filter half of profile-based
// personalization: degree level and location eligibility. Call this to
// filter a candidate list *before* handing it to ranker.Rank.
func (l Listing) MatchesProfile(p Profile) bool {
	if p.DegreeLevel != "" && len(l.DegreeLevels) > 0 {
		want := normalize(p.DegreeLevel)
		matched := false
		for _, dl := range l.DegreeLevels {
			if normalize(dl) == want {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if p.Location != "" && !l.Remote && l.Location != "" {
		want := normalize(p.Location)
		have := normalize(l.Location)
		// Substring match in both directions covers "Boston" matching
		// "Boston, MA" and vice versa without needing real geocoding.
		if want != have && !strings.Contains(have, want) && !strings.Contains(want, have) {
			return false
		}
	}

	return true
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}