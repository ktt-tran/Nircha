package main

import (
	"flag"
	"fmt"
	"log"

	"search-engine/ranker/document"
	"search-engine/ranker/ranker"
	"search-engine/ranker/repository"

	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()

	pageRankMode := flag.Bool("pagerank", false, "run the legacy offline PageRank batch job instead of the composite ranker demo")
	query := flag.String("query", "software engineering internship", "query to run through the composite ranker demo")
	major := flag.String("major", "", "optional profile: major, e.g. \"computer science\"")
	degree := flag.String("degree", "", "optional profile: degree level, e.g. \"undergraduate\"")
	location := flag.String("location", "", "optional profile: location, e.g. \"Boston\"")
	flag.Parse()

	if *pageRankMode {
		runPageRank()
		return
	}

	profile := document.Profile{Major: *major, DegreeLevel: *degree, Location: *location}
	runCompositeDemo(*query, profile)
}

func runPageRank() {
	repo := repository.NewRedisRepository()
	r := ranker.NewPageRankRanker(repo)
	if err := r.Rank(); err != nil {
		panic(err)
	}
	log.Println("PageRank ranking completed")
}

func runCompositeDemo(query string, profile document.Profile) {
	now := time.Now()
	weekAgo := now.Add(-7 * 24 * time.Hour)
	sixMonthsAgo := now.Add(-180 * 24 * time.Hour)
	inTenDays := now.Add(10 * 24 * time.Hour)
	inFiveMonths := now.Add(150 * 24 * time.Hour)

	candidates := []ranker.Candidate{
		{
			BM25Score: 5.0,
			Listing: document.Listing{
				ID:            "1",
				Title:         "Software Engineering Summer Internship",
				Organization:  "Acme Robotics",
				Description:   "Build tooling for our robotics platform alongside the firmware team.",
				Tags:          []string{"software", "internship", "summer"},
				FieldsOfStudy: []string{"computer science", "software engineering"},
				DegreeLevels:  []string{"undergraduate"},
				Location:      "Boston, MA",
				TrustTier:     document.TrustTierPrimary,
				PostedAt:      weekAgo,
				LastCrawledAt: weekAgo,
				DeadlineAt:    &inTenDays,
			},
		},
		{
			BM25Score: 1.0,
			Listing: document.Listing{
				ID:            "2",
				Title:         "Research Assistant, Computational Biology",
				Organization:  "State University",
				Description:   "Support a wet lab team using software tools for genomic analysis.",
				Tags:          []string{"research", "biology"},
				FieldsOfStudy: []string{"biology"},
				DegreeLevels:  []string{"graduate", "phd"},
				Location:      "Seattle, WA",
				TrustTier:     document.TrustTierPrimary,
				PostedAt:      sixMonthsAgo,
				LastCrawledAt: weekAgo,
				DeadlineAt:    nil,
			},
		},
		{
			BM25Score: 3.5,
			Listing: document.Listing{
				ID:           "3",
				Title:        "General Internship Listings Roundup",
				Organization: "SomeAggregatorSite",
				Description:  "A roundup mentioning software, research, and other internship categories.",
				Tags:         []string{"internship"},
				// FieldsOfStudy, DegreeLevels, and Location deliberately
				// left unset here to demo the "unknown passes through"
				// rule. This listing should never be excluded by
				// profile filtering, only ever ranked (and ranked as
				// FieldOfStudyMatch-neutral).
				Remote:        true,
				TrustTier:     document.TrustTierUnverified,
				PostedAt:      weekAgo,
				LastCrawledAt: weekAgo,
				DeadlineAt:    &inFiveMonths,
			},
		},
	}

	// Hard filter: degree level and location eligibility. This is where
	// MatchesProfile plugs in after retrieval, before ranking. Major is
	// intentionally not applied here; see ranker.Rank's doc comment.
	filtered := make([]ranker.Candidate, 0, len(candidates))
	for _, c := range candidates {
		if c.Listing.MatchesProfile(profile) {
			filtered = append(filtered, c)
		}
	}

	cfg := ranker.DefaultConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid ranker config: %v", err)
	}

	results := ranker.Rank(cfg, query, filtered, profile, now)

	fmt.Printf("Query: %q  profile: %+v  (%d candidates, %d after profile filter)\n\n",
		query, profile, len(candidates), len(filtered))
	for i, r := range results {
		fmt.Printf("%d. %s — %s  (final=%.3f)\n", i+1, r.Listing.Title, r.Listing.Organization, r.Breakdown.Final)
		fmt.Printf("   bm25=%.3f  field=%.3f  fieldOfStudy=%.3f  trust=%.3f  fresh=%.3f  urgency=%.3f\n\n",
			r.Breakdown.BM25, r.Breakdown.FieldMatch, r.Breakdown.FieldOfStudyMatch,
			r.Breakdown.TrustTier, r.Breakdown.Freshness, r.Breakdown.DeadlineUrgency)
	}
}