package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"search-engine/searcher/models"
	"search-engine/searcher/tokenizer"

	"github.com/redis/go-redis/v9"
)

type seedPage struct {
	id            int64
	title         string
	abstract      string
	url           string
	organization  string
	trustTier     int
	postedDaysAgo int
	deadlineDays  int
	fieldsOfStudy []string
	degreeLevels  []string
	location      string
	remote        bool
}

var samplePages = []seedPage{
	{
		id: 1, title: "Summer Research Internship in Computer Science",
		abstract: "A 10-week paid summer research internship for undergraduates in computer science.",
		url:      "https://example.edu/op/1", organization: "Example University",
		trustTier: 1, postedDaysAgo: 5, deadlineDays: 21,
		fieldsOfStudy: []string{"computer science"}, degreeLevels: []string{"undergraduate"},
		location: "Cambridge, MA",
	},
	{
		id: 2, title: "Undergraduate Research Opportunity in Biology",
		abstract: "Work in a molecular biology lab over the summer with faculty mentorship.",
		url:      "https://example.edu/op/2", organization: "Example University",
		trustTier: 1, postedDaysAgo: 60, deadlineDays: 0,
		fieldsOfStudy: []string{"biology"}, degreeLevels: []string{"undergraduate", "graduate"},
		location: "Cambridge, MA",
	},
	{
		id: 3, title: "Software Engineering Internship",
		abstract: "Join our engineering team for a 12-week software internship building production systems.",
		url:      "https://example.com/careers/3", organization: "Example Corp",
		trustTier: 2, postedDaysAgo: 2, deadlineDays: 45,
		fieldsOfStudy: []string{"computer science", "software engineering"}, degreeLevels: []string{"undergraduate"},
		remote: true,
	},
	{
		id: 4, title: "Data Science Fellowship",
		abstract: "A year-long fellowship applying data science methods to public policy research.",
		url:      "https://example.org/fellowship/4", organization: "Example Policy Institute",
		trustTier: 3, postedDaysAgo: 150, deadlineDays: 0,
		fieldsOfStudy: []string{"data science", "public health"}, degreeLevels: []string{"phd"},
		location: "New York, NY",
	},
}

func main() {
	ctx := context.Background()

	addr := os.Getenv("REDIS_CONNECTION_ADDRESS")
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("seed: cannot reach Redis at %s: %v", addr, err)
	}

	wordFreq := map[string]int{}
	now := time.Now()

	for _, p := range samplePages {
		idStr := fmt.Sprintf("%d", p.id)

		page := models.Page{
			Title: p.title, Abstract: p.abstract, Url: p.url,
			Organization:  p.organization,
			TrustTier:     p.trustTier,
			PostedAt:      now.Add(-time.Duration(p.postedDaysAgo) * 24 * time.Hour),
			LastCrawledAt: now,
			FieldsOfStudy: p.fieldsOfStudy,
			DegreeLevels:  p.degreeLevels,
			Location:      p.location,
			Remote:        p.remote,
		}
		if p.deadlineDays > 0 {
			deadline := now.Add(time.Duration(p.deadlineDays) * 24 * time.Hour)
			page.DeadlineAt = &deadline
		}

		data, err := json.Marshal(page)
		if err != nil {
			log.Fatal(err)
		}
		if err := rdb.HSet(ctx, "pages", idStr, data).Err(); err != nil {
			log.Fatal(err)
		}

		tokens := tokenizer.Tokenize(p.title)
		fmt.Printf("page %d %-55q -> %v\n", p.id, p.title, tokens)
		for pos, tok := range tokens {
			const score = 1 // placeholder term-weight; the real indexer decides this
			value := fmt.Sprintf("%d:%d", score, pos)
			if err := rdb.HSet(ctx, "w:"+tok, idStr, value).Err(); err != nil {
				log.Fatal(err)
			}
			wordFreq[tok]++
		}
	}

	for word, count := range wordFreq {
		if err := rdb.HSet(ctx, "words-freq", word, fmt.Sprintf("%d", count)).Err(); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("seed complete: %d pages, %d unique words\n", len(samplePages), len(wordFreq))
}
