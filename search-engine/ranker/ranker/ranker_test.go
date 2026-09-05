package ranker

import (
	"math"
	"testing"
	"time"

	"search-engine/ranker/document"
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestDefaultWeightsSumToOne(t *testing.T) {
	got := DefaultWeights().Sum()
	if !approxEqual(got, 1.0) {
		t.Fatalf("DefaultWeights().Sum() = %v, want 1.0", got)
	}
}

func TestDefaultConfigValidates(t *testing.T) {
	if err := DefaultConfig().Validate(); err != nil {
		t.Fatalf("DefaultConfig() should validate, got error: %v", err)
	}
}

func TestConfigValidateCatchesBadWeights(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Weights.BM25 = 0.9 // now sums to > 1.0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected Validate to reject weights that don't sum to 1.0")
	}
}

func TestConfigValidateCatchesNegativeDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FreshnessHalfLife = 0
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected Validate to reject a non-positive FreshnessHalfLife")
	}
}

func TestNormalizeBM25(t *testing.T) {
	candidates := []Candidate{
		{BM25Score: 1.0},
		{BM25Score: 3.0},
		{BM25Score: 5.0},
	}
	got := normalizeBM25(candidates)
	want := []float64{0.0, 0.5, 1.0}
	for i := range want {
		if !approxEqual(got[i], want[i]) {
			t.Errorf("normalizeBM25()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNormalizeBM25AllEqual(t *testing.T) {
	candidates := []Candidate{{BM25Score: 2.0}, {BM25Score: 2.0}}
	got := normalizeBM25(candidates)
	for i, v := range got {
		if !approxEqual(v, 1.0) {
			t.Errorf("normalizeBM25()[%d] = %v, want 1.0 for a zero-spread set", i, v)
		}
	}
}

func TestNormalizeBM25Empty(t *testing.T) {
	got := normalizeBM25(nil)
	if len(got) != 0 {
		t.Fatalf("normalizeBM25(nil) = %v, want empty", got)
	}
}

func TestFreshnessScoreDecay(t *testing.T) {
	now := time.Now()
	halfLife := 14 * 24 * time.Hour

	fresh := document.Listing{PostedAt: now}
	if got := freshnessScore(fresh, now, halfLife); !approxEqual(got, 1.0) {
		t.Errorf("freshnessScore(age=0) = %v, want 1.0", got)
	}

	atHalfLife := document.Listing{PostedAt: now.Add(-halfLife)}
	if got := freshnessScore(atHalfLife, now, halfLife); !approxEqual(got, 0.5) {
		t.Errorf("freshnessScore(age=halfLife) = %v, want 0.5", got)
	}

	old := document.Listing{PostedAt: now.Add(-10 * halfLife)}
	if got := freshnessScore(old, now, halfLife); got >= 0.01 {
		t.Errorf("freshnessScore(age=10*halfLife) = %v, want close to 0", got)
	}

	noTimestamp := document.Listing{}
	if got := freshnessScore(noTimestamp, now, halfLife); !approxEqual(got, 0.5) {
		t.Errorf("freshnessScore(no timestamp) = %v, want neutral 0.5", got)
	}
}

func TestFreshnessScoreFallsBackToLastCrawled(t *testing.T) {
	now := time.Now()
	l := document.Listing{LastCrawledAt: now.Add(-24 * time.Hour)}
	got := freshnessScore(l, now, 14*24*time.Hour)
	if got <= 0.9 || got >= 1.0 {
		t.Errorf("freshnessScore fallback to LastCrawledAt = %v, want just under 1.0", got)
	}
}

func TestDeadlineUrgencyNoDeadline(t *testing.T) {
	now := time.Now()
	l := document.Listing{DeadlineAt: nil}
	got := deadlineUrgencyScore(l, now, 60*24*time.Hour, 0.5)
	if !approxEqual(got, 0.5) {
		t.Errorf("deadlineUrgencyScore(no deadline) = %v, want the neutral default 0.5", got)
	}
}

func TestDeadlineUrgencyImminent(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	l := document.Listing{DeadlineAt: &tomorrow}
	got := deadlineUrgencyScore(l, now, 60*24*time.Hour, 0.5)
	if got <= 0.95 {
		t.Errorf("deadlineUrgencyScore(due tomorrow, 60d window) = %v, want close to 1.0", got)
	}
}

func TestDeadlineUrgencyFarOff(t *testing.T) {
	now := time.Now()
	inFiveMonths := now.Add(150 * 24 * time.Hour)
	l := document.Listing{DeadlineAt: &inFiveMonths}
	got := deadlineUrgencyScore(l, now, 60*24*time.Hour, 0.5)
	if !approxEqual(got, 0) {
		t.Errorf("deadlineUrgencyScore(150d out, 60d window) = %v, want 0 (not urgent yet)", got)
	}
}

func TestDeadlineUrgencyExpiredIsDefensiveZero(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	l := document.Listing{DeadlineAt: &yesterday}
	got := deadlineUrgencyScore(l, now, 60*24*time.Hour, 0.5)
	if !approxEqual(got, 0) {
		t.Errorf("deadlineUrgencyScore(expired) = %v, want 0", got)
	}
}

func TestTrustTierScoreOrdering(t *testing.T) {
	primary := trustTierScore(document.TrustTierPrimary, document.TrustTierUnverified)
	verified := trustTierScore(document.TrustTierVerified, document.TrustTierUnverified)
	unverified := trustTierScore(document.TrustTierUnverified, document.TrustTierUnverified)

	if !(primary > verified && verified > unverified) {
		t.Errorf("expected primary(%v) > verified(%v) > unverified(%v)", primary, verified, unverified)
	}
	if !approxEqual(primary, 1.0) {
		t.Errorf("trustTierScore(Primary) = %v, want 1.0", primary)
	}
	if !approxEqual(unverified, 0.0) {
		t.Errorf("trustTierScore(least trusted tier) = %v, want 0.0", unverified)
	}
}

func TestTrustTierScoreClampsOutOfRange(t *testing.T) {
	got := trustTierScore(document.TrustTier(99), document.TrustTierUnverified)
	if !approxEqual(got, 0.0) {
		t.Errorf("trustTierScore(tier beyond max) = %v, want clamped to 0.0", got)
	}
}

func TestFieldMatchScoreTitleBeatsNoMatch(t *testing.T) {
	terms := tokenize("software internship")
	titleMatch := document.Listing{Title: "Software Internship Program", Organization: "Acme"}
	noMatch := document.Listing{Title: "Marketing Fellowship", Organization: "Acme"}

	titleScore := fieldMatchScore(terms, titleMatch)
	noMatchScore := fieldMatchScore(terms, noMatch)

	if titleScore <= noMatchScore {
		t.Errorf("expected a title match (%v) to score above no match (%v)", titleScore, noMatchScore)
	}
	if titleScore <= 0 {
		t.Errorf("fieldMatchScore with full title match = %v, want > 0", titleScore)
	}
}

func TestFieldMatchScoreEmptyQuery(t *testing.T) {
	got := fieldMatchScore(nil, document.Listing{Title: "Anything"})
	if !approxEqual(got, 0) {
		t.Errorf("fieldMatchScore(empty query) = %v, want 0", got)
	}
}

func TestRankSortsDescendingByFinalScore(t *testing.T) {
	now := time.Now()
	tenDaysOut := now.Add(10 * 24 * time.Hour)

	candidates := []Candidate{
		{
			BM25Score: 1.0,
			Listing: document.Listing{
				ID: "weak", Title: "Marketing Fellowship", Organization: "Random Org",
				TrustTier: document.TrustTierUnverified, PostedAt: now.Add(-200 * 24 * time.Hour),
			},
		},
		{
			BM25Score: 5.0,
			Listing: document.Listing{
				ID: "strong", Title: "Software Engineering Internship", Organization: "Acme",
				Tags: []string{"software", "internship"}, TrustTier: document.TrustTierPrimary,
				PostedAt: now, DeadlineAt: &tenDaysOut,
			},
		},
	}

	results := Rank(DefaultConfig(), "software engineering internship", candidates, document.Profile{}, now)

	if len(results) != 2 {
		t.Fatalf("Rank returned %d results, want 2", len(results))
	}
	if results[0].Listing.ID != "strong" {
		t.Errorf("expected the stronger candidate first, got %q first", results[0].Listing.ID)
	}
	if results[0].Breakdown.Final <= results[1].Breakdown.Final {
		t.Errorf("results not sorted descending: %v then %v", results[0].Breakdown.Final, results[1].Breakdown.Final)
	}
}

func TestRankEmptyCandidates(t *testing.T) {
	results := Rank(DefaultConfig(), "anything", nil, document.Profile{}, time.Now())
	if len(results) != 0 {
		t.Fatalf("Rank(nil candidates) = %d results, want 0", len(results))
	}
}

func TestFieldOfStudyMatchScoreNeutralWithNoProfile(t *testing.T) {
	l := document.Listing{FieldsOfStudy: []string{"biology"}}
	got := fieldOfStudyMatchScore("", l)
	if !approxEqual(got, 0.5) {
		t.Errorf("fieldOfStudyMatchScore(no major given) = %v, want neutral 0.5", got)
	}
}

func TestFieldOfStudyMatchScoreNeutralWithNoListingData(t *testing.T) {
	l := document.Listing{} // FieldsOfStudy unset
	got := fieldOfStudyMatchScore("computer science", l)
	if !approxEqual(got, 0.5) {
		t.Errorf("fieldOfStudyMatchScore(listing has no FieldsOfStudy) = %v, want neutral 0.5", got)
	}
}

func TestFieldOfStudyMatchScoreMatch(t *testing.T) {
	l := document.Listing{FieldsOfStudy: []string{"computer science", "software engineering"}}
	got := fieldOfStudyMatchScore("computer science", l)
	if !approxEqual(got, 1.0) {
		t.Errorf("fieldOfStudyMatchScore(matching major) = %v, want 1.0", got)
	}
}

func TestFieldOfStudyMatchScoreMismatchIsLowNotZero(t *testing.T) {
	l := document.Listing{FieldsOfStudy: []string{"biology"}}
	got := fieldOfStudyMatchScore("computer science", l)
	if got <= 0 || got >= 0.5 {
		t.Errorf("fieldOfStudyMatchScore(explicit mismatch) = %v, want a low but nonzero soft penalty", got)
	}
}

func TestRankIsProfileAwareOnMajor(t *testing.T) {
	now := time.Now()
	candidates := []Candidate{
		{
			BM25Score: 3.0,
			Listing: document.Listing{
				ID: "cs", Title: "Research Internship", Organization: "Acme Labs",
				FieldsOfStudy: []string{"computer science"}, TrustTier: document.TrustTierPrimary, PostedAt: now,
			},
		},
		{
			BM25Score: 3.0,
			Listing: document.Listing{
				ID: "bio", Title: "Research Internship", Organization: "Acme Labs",
				FieldsOfStudy: []string{"biology"}, TrustTier: document.TrustTierPrimary, PostedAt: now,
			},
		},
	}

	// Identical on every signal except FieldsOfStudy -- a CS-major profile
	// should push the CS listing above the otherwise-tied biology listing.
	results := Rank(DefaultConfig(), "research internship", candidates, document.Profile{Major: "computer science"}, now)

	if results[0].Listing.ID != "cs" {
		t.Errorf("expected the field-of-study match to rank first with a matching profile, got %q first", results[0].Listing.ID)
	}

	// The same two candidates, with no profile, should be an exact tie in
	// Final score, proving FieldOfStudyMatch doesn't leak bias in when
	// there's nothing to compare against.
	noProfileResults := Rank(DefaultConfig(), "research internship", candidates, document.Profile{}, now)
	if !approxEqual(noProfileResults[0].Breakdown.Final, noProfileResults[1].Breakdown.Final) {
		t.Errorf("expected a tie with no profile, got %v vs %v",
			noProfileResults[0].Breakdown.Final, noProfileResults[1].Breakdown.Final)
	}
}