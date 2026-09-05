
import (
"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"

	"search-engine/searcher/models"

	"github.com/redis/go-redis/v9"
)

// RedisRepository wraps the Redis client. It intentionally holds no mutex:
// *redis.Client already manages its own connection pool and is safe for
// concurrent use by multiple goroutines (that's the whole point of a
// connection pool). The previous version wrapped every single method in a
// shared *sync.Mutex, which serialized every Redis call in the process -
// including the per-query-token goroutines fired off in services/search.go
// specifically to run those lookups in parallel. The mutex silently
// defeated that concurrency rather than protecting anything.
type RedisRepository struct {
	rh *redis.Client
}

var keyPrefixes = struct {
	Word            string
	Pages           string
	PageToId        string
	IdToPage        string
	OutboundLinks   string
	WordFrequencies string
}{
	Word:            "w:",
	Pages:           "pages",
	PageToId:        "p-i:",
	IdToPage:        "i-p:",
	OutboundLinks:   "out:",
	WordFrequencies: "words-freq",
}

// GetPagesForWord returns, for a single query token, the map of matching
// page ID (as a string) -> raw hit-data blob stored at index time.
func (r *RedisRepository) GetPagesForWord(ctx context.Context, word string) map[string]string {
	res, err := r.rh.HGetAll(ctx, keyPrefixes.Word+word).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Printf("repository: GetPagesForWord(%q): %v", word, err)
		}
		return nil
	}
	return res
}

// GetPage fetches and decodes a single page by ID. It returns nil if the
// page doesn't exist or the stored value can't be decoded - callers must
// nil-check the result before dereferencing it.
func (r *RedisRepository) GetPage(ctx context.Context, id int64) *models.Page {
	key := strconv.FormatInt(id, 10)
	res, err := r.rh.HGet(ctx, keyPrefixes.Pages, key).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Printf("repository: GetPage(%d): %v", id, err)
		}
		return nil
	}

	var page models.Page
	if err := json.Unmarshal([]byte(res), &page); err != nil {
		log.Printf("repository: GetPage(%d): decoding stored page: %v", id, err)
		return nil
	}
	return &page
}

// GetPages batch-fetches and decodes multiple pages by ID in a single
// round trip via HMGet. Used to resolve a whole merged candidate set's
// full listing data at once (see pipes.Resolve), rather than one HGet per
// candidate: ranking needs every candidate's full data up front (trust
// tier, dates, field of study) to score and hard-filter it, not just the
// handful of results a page of results will actually show. IDs that don't
// exist or fail to decode are simply omitted from the result rather than
// failing the whole batch.
func (r *RedisRepository) GetPages(ctx context.Context, pageIds []string) map[string]models.Page {
	if len(pageIds) == 0 {
		return map[string]models.Page{}
	}

	res, err := r.rh.HMGet(ctx, keyPrefixes.Pages, pageIds...).Result()
	if err != nil {
		log.Printf("repository: GetPages: %v", err)
		return nil
	}

	pages := make(map[string]models.Page, len(res))
	for i, raw := range res {
		if raw == nil {
			continue
		}
		rawStr, ok := raw.(string)
		if !ok {
			continue
		}
		var page models.Page
		if err := json.Unmarshal([]byte(rawStr), &page); err != nil {
			log.Printf("repository: GetPages: decoding page %s: %v", pageIds[i], err)
			continue
		}
		pages[pageIds[i]] = page
	}

	return pages
}

// GetAllWords returns every indexed word and its document frequency. Used
// once at startup to build the fuzzy-search dictionary.
func (r *RedisRepository) GetAllWords(ctx context.Context) ([]models.WordFrequency, error) {
	res, err := r.rh.HGetAll(ctx, keyPrefixes.WordFrequencies).Result()
	if err != nil {
		return nil, err
	}

	frequencies := make([]models.WordFrequency, 0, len(res))
	for word, count := range res {
		countInt, err := strconv.Atoi(count)
		if err != nil {
			continue
		}
		frequencies = append(frequencies, models.WordFrequency{
			Word:  word,
			Count: countInt,
		})
	}
	return frequencies, nil
}

// Ping verifies connectivity to Redis. Used at startup (to fail fast with a
// clear error rather than a confusing downstream failure) and by the
// /api/health endpoint (to report real readiness, not just "the process is
// running").
func (r *RedisRepository) Ping(ctx context.Context) error {
	return r.rh.Ping(ctx).Err()
}

// Close releases the underlying connection pool. Call during graceful
// shutdown.
func (r *RedisRepository) Close() error {
	return r.rh.Close()
}

func NewRedisRepository() *RedisRepository {
	connectionAddress := os.Getenv("REDIS_CONNECTION_ADDRESS")
	password := os.Getenv("REDIS_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     connectionAddress,
		Password: password,
		DB:       0,
	})

	return &RedisRepository{rh: rdb}
}