package ranker

import (
	"fmt"
	"math"
	"time"

	"search-engine/ranker/document"
)

// Weights controls how much each signal contributes to a listing's final
// score. They are expected to sum to 1.0 -- see Validate.
type Weights struct {
	BM25              float64
	FieldMatch        float64
	FieldOfStudyMatch float64
	TrustTier         float64
	Freshness         float64
	DeadlineUrgency   float64
}

// DefaultWeights returns the v2 weighting.
func DefaultWeights() Weights {
	return Weights{
		BM25:              0.35,
		FieldMatch:        0.15,
		FieldOfStudyMatch: 0.15,
		TrustTier:         0.10,
		Freshness:         0.15,
		DeadlineUrgency:   0.10,
	}
}

// Sum reports the total of all weights, so a customized Weights value can
// be checked before it's wired in.
func (w Weights) Sum() float64 {
	return w.BM25 + w.FieldMatch + w.FieldOfStudyMatch + w.TrustTier + w.Freshness + w.DeadlineUrgency
}

type Config struct {
	Weights Weights

	// FreshnessHalfLife is how long it takes a listing's freshness score
	// to decay to 0.5. Two weeks fits fast-moving postings like internships
	// and research positions; lengthen it for slower-moving content (e.g.
	// standing graduate programs).
	FreshnessHalfLife time.Duration

	// UrgencyWindow is how far out a deadline has to be before it starts
	// contributing urgency at all. Inside the window, urgency ramps
	// linearly from 0 up to 1 as the deadline approaches. Outside it (or
	// with no deadline at all), a listing isn't "urgent" yet.
	UrgencyWindow time.Duration

	// NoDeadlineUrgency is the score given to listings with no deadline
	// (rolling admissions, always-open positions). Neutral by default, so
	// simply not having a deadline neither helps nor hurts a listing.
	NoDeadlineUrgency float64

	// MaxTrustTier is the lowest (least trusted) tier value in use, needed
	// to normalize TrustTier into [0,1]. Raise it if more tiers are added
	// later; tiers beyond it are clamped rather than erroring.
	MaxTrustTier document.TrustTier
}

// DefaultConfig returns DefaultWeights plus reasonable starting values for
// the remaining knobs, tuned for opportunity listings with day-to-month
// scale deadlines and week-to-month scale freshness windows.
func DefaultConfig() Config {
	return Config{
		Weights:           DefaultWeights(),
		FreshnessHalfLife: 14 * 24 * time.Hour,
		UrgencyWindow:     60 * 24 * time.Hour,
		NoDeadlineUrgency: 0.5,
		MaxTrustTier:      document.TrustTierUnverified,
	}
}

// Validate catches configuration mistakes early. A mistyped weight or a
// negative duration should fail loudly at startup.
func (c Config) Validate() error {
	const tolerance = 1e-9
	if math.Abs(c.Weights.Sum()-1.0) > tolerance {
		return fmt.Errorf("ranker: weights must sum to 1.0, got %.6f", c.Weights.Sum())
	}
	if c.Weights.BM25 < 0 || c.Weights.FieldMatch < 0 || c.Weights.FieldOfStudyMatch < 0 ||
		c.Weights.TrustTier < 0 || c.Weights.Freshness < 0 || c.Weights.DeadlineUrgency < 0 {
		return fmt.Errorf("ranker: weights must be non-negative: %+v", c.Weights)
	}
	if c.FreshnessHalfLife <= 0 {
		return fmt.Errorf("ranker: FreshnessHalfLife must be positive, got %v", c.FreshnessHalfLife)
	}
	if c.UrgencyWindow <= 0 {
		return fmt.Errorf("ranker: UrgencyWindow must be positive, got %v", c.UrgencyWindow)
	}
	if c.NoDeadlineUrgency < 0 || c.NoDeadlineUrgency > 1 {
		return fmt.Errorf("ranker: NoDeadlineUrgency must be in [0,1], got %v", c.NoDeadlineUrgency)
	}
	if c.MaxTrustTier < 1 {
		return fmt.Errorf("ranker: MaxTrustTier must be >= 1, got %v", c.MaxTrustTier)
	}
	return nil
}