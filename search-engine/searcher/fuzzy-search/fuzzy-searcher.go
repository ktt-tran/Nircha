package fuzzysearch

import (
	"context"
	"log"
	"sort"

	"search-engine/searcher/repository"
)

type FuzzySearcher struct {
	repo      *repository.RedisRepository
	bkt       *BKTree
	tolerance int
}

func (f *FuzzySearcher) loadDictionary(ctx context.Context) error {
	words, err := f.repo.GetAllWords(ctx)
	if err != nil {
		return err
	}

	for i := range words {
		f.bkt.Add(&words[i])
	}
	log.Printf("fuzzysearch: dictionary loaded (%d words)", len(words))
	return nil
}

// Search returns candidate corrections for word, trying increasing edit
// distances (1..tolerance) and stopping at the first distance that yields
// any match, then sorting those matches by distance and, as a tiebreaker,
// by how frequently the candidate word appears in the index.
func (f *FuzzySearcher) Search(word string) []string {
	var results []Result
	for i := 1; i <= f.tolerance; i++ {
		// NOTE: this used to be `results := f.bkt.Search(word, i)`, which
		// declares a new, inner-scoped `results` that shadows the outer
		// one instead of assigning to it. The outer `results` therefore
		// stayed nil forever, and Search() always returned no matches
		// regardless of input. `=` (not `:=`) fixes that.
		results = f.bkt.Search(word, i)
		if len(results) > 0 {
			break
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].distance != results[j].distance {
			return results[i].distance < results[j].distance
		}
		return results[i].frequency > results[j].frequency
	})

	var res []string
	min := -1
	for _, r := range results {
		if min == -1 {
			min = r.distance
		}
		if r.distance > min {
			break
		}
		res = append(res, r.word)
	}

	return res
}

func NewFuzzySearcher(ctx context.Context, repo *repository.RedisRepository, tolerance int) (*FuzzySearcher, error) {
	bkt := NewBKTree()
	f := &FuzzySearcher{repo, bkt, tolerance}
	err := f.loadDictionary(ctx)
	return f, err
}