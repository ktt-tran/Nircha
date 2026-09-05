package services

import (
	"context"
	"log"
	"strconv"
	"strings"
	"sync"

	fuzzysearch "search-engine/searcher/fuzzy-search"
	"search-engine/searcher/models"
	"search-engine/searcher/repository"
)

type SearchService struct {
	repo          *repository.RedisRepository
	fuzzySearcher *fuzzysearch.FuzzySearcher
}

func NewSearchService(repo *repository.RedisRepository, fuzzySearcher *fuzzysearch.FuzzySearcher) *SearchService {
	return &SearchService{repo, fuzzySearcher}
}

// Search looks up each token in parallel (one goroutine per token) and
// returns a map of token -> matching pages. A token with no direct matches
// falls back to its best fuzzy-search correction before being counted as
// empty.
func (ss *SearchService) Search(ctx context.Context, tokens []string) (map[string][]models.SearchDBResult, error) {
	results := map[string][]models.SearchDBResult{}

	wg := sync.WaitGroup{}
	wg.Add(len(tokens))
	mu := sync.Mutex{}
	for _, token := range tokens {
		go func(token string) {
			defer wg.Done()
			// A panic inside a spawned goroutine is NOT recovered by
			// net/http's per-request recovery - that only guards the
			// goroutine net/http itself started for the request. An
			// unrecovered panic here would take down the entire process,
			// mid-flight requests and all. Given that, this goroutine gets
			// its own safety net: worst case, this one token contributes
			// no results instead of the whole server going down.
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("services: search token %q panicked: %v", token, rec)
				}
			}()

			res := ss.searchToken(ctx, token)
			if len(res) == 0 {
				return
			}

			mu.Lock()
			results[token] = res
			mu.Unlock()
		}(token)
	}
	wg.Wait()
	return results, nil
}

// Suggest returns, for each token that has no direct index match, up to
// limit candidate spelling corrections (closest edit distance first).
// Tokens that already match the index exactly are omitted from the result,
// since "did you mean 'computer'?" for a query that already found
// "computer" isn't a useful suggestion.
func (ss *SearchService) Suggest(ctx context.Context, tokens []string, limit int) map[string][]string {
	suggestions := make(map[string][]string, len(tokens))

	for _, token := range tokens {
		if pages := ss.repo.GetPagesForWord(ctx, token); len(pages) > 0 {
			continue
		}

		candidates := ss.fuzzySearcher.Search(token)
		if len(candidates) == 0 {
			continue
		}
		if limit > 0 && len(candidates) > limit {
			candidates = candidates[:limit]
		}
		suggestions[token] = candidates
	}

	return suggestions
}

// searchToken resolves a single token to its matching pages, falling back
// to the closest fuzzy-search correction when there's no direct match.
func (ss *SearchService) searchToken(ctx context.Context, token string) []models.SearchDBResult {
	pages := ss.repo.GetPagesForWord(ctx, token)
	if len(pages) == 0 {
		fuzzTerms := ss.fuzzySearcher.Search(token)
		if len(fuzzTerms) == 0 {
			return nil
		}
		pages = ss.repo.GetPagesForWord(ctx, fuzzTerms[0])
	}
	if len(pages) == 0 {
		return nil
	}

	res := make([]models.SearchDBResult, 0, len(pages))
	for page, info := range pages {
		pageInt, err := strconv.Atoi(page)
		if err != nil {
			log.Printf("services: search token %q: bad page id %q: %v", token, page, err)
			continue
		}

		// info is stored as "score:position". Guard against anything else
		// ending up in that field instead of indexing blindly into a
		// 1-element slice, which previously panicked (and, unrecovered,
		// would have killed the whole process - see the recover() above).
		parts := strings.SplitN(info, ":", 2)
		if len(parts) != 2 {
			log.Printf("services: search token %q: malformed hit data %q for page %q", token, info, page)
			continue
		}
		scoreInt, errScore := strconv.Atoi(parts[0])
		positionInt, errPos := strconv.Atoi(parts[1])
		if errScore != nil || errPos != nil {
			log.Printf("services: search token %q: malformed hit data %q for page %q", token, info, page)
			continue
		}

		res = append(res, models.SearchDBResult{
			PageId:   int64(pageInt),
			Score:    scoreInt,
			Position: positionInt,
		})
	}

	return res
}