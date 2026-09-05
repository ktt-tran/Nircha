// Package document defines the shape of a single opportunity listing
// (internship, job, research position, or university program) as the
// ranker sees it.
package document

import "time"

// TrustTier is a manually-assigned, curated measure of how authoritative a
// listing's source is. Lower numbers are more trusted. This is the
// deliberately simple stand-in for a computed link-graph authority score:
// PageRank was evaluated and set aside for v1 because the opportunity-
// listing link graph is too sparse for it to produce a meaningful signal,
// so trust is assigned by hand per seed site instead.
type TrustTier int

const (
	// TrustTierPrimary is a listing published directly by the org offering
	// it: a university's own admissions/financial-aid pages, a company's
	// own careers page, an official government listing (e.g. USAJobs).
	TrustTierPrimary TrustTier = 1

	// TrustTierVerified is a reputable secondary source that reliably
	// mirrors primary listings (a known, curated aggregator).
	TrustTierVerified TrustTier = 2

	// TrustTierUnverified is a lower-confidence source: a newly added
	// seed, a general aggregator, or a site without a track record yet.
	TrustTierUnverified TrustTier = 3
)

// Listing is one opportunity document after crawling.
type Listing struct {
	ID           string
	Title        string
	Organization string
	Description  string
	Tags         []string
	URL          string
	SourceDomain string
	TrustTier    TrustTier
	PostedAt time.Time
	LastCrawledAt time.Time
	DeadlineAt *time.Time
	FieldsOfStudy []string
	DegreeLevels []string
	Location string
	Remote   bool
}

// EffectiveFreshnessTime returns the timestamp freshness scoring should
// measure age from: PostedAt when known, otherwise the first-crawl time as
// a best-effort proxy.
func (l Listing) EffectiveFreshnessTime() time.Time {
	if !l.PostedAt.IsZero() {
		return l.PostedAt
	}
	return l.LastCrawledAt
}