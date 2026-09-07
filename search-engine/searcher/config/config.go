package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// Port the HTTP server listens on.
	Port string
	// RedisAddress is the host:port of the Redis instance backing the
	// search index.
	RedisAddress string
	// RedisPassword authenticates to Redis. Empty means no auth (typical
	// for local dev).
	RedisPassword string

	// CORSAllowedOrigins is the set of origins allowed to call this API
	// from a browser (e.g. the React dev server / deployed frontend). A
	// single "*" allows any origin.
	CORSAllowedOrigins []string

	// FuzzyTolerance is the maximum edit distance the fuzzy searcher will
	// widen to when looking for a correction to a term with no direct
	// matches.
	FuzzyTolerance int
	// DefaultPageSize is used when a search request omits `size`.
	DefaultPageSize int
	// MaxPageSize caps `size` on incoming requests so a client can't
	// force an arbitrarily expensive page fetch.
	MaxPageSize int
	// MaxQueryLength caps the length (in runes) of `q` on incoming
	// requests.
	MaxQueryLength int
	// SuggestLimit caps how many candidate corrections /api/suggest
	// returns per token.
	SuggestLimit int

	// SearchTimeout bounds how long a single /api/search request is
	// allowed to run (covers the fan-out of per-token Redis lookups)
	// before the handler gives up and returns 504.
	SearchTimeout time.Duration
	// StartupTimeout bounds how long building the fuzzy-search dictionary
	// is allowed to take at process startup.
	StartupTimeout time.Duration
}

// Load reads configuration from the process environment, applying a
// documented default for anything unset. It does not call godotenv.Load()
// itself, that's a process-bootstrapping side effect that belongs in
// main(), before Load() runs, so this function has one job: turn whatever
// is already in the environment into a validated Config.
func Load() (*Config, error) {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		RedisAddress:       getEnv("REDIS_CONNECTION_ADDRESS", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		CORSAllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:3001"}),
		FuzzyTolerance:     getEnvInt("FUZZY_TOLERANCE", 2),
		DefaultPageSize:    getEnvInt("DEFAULT_PAGE_SIZE", 10),
		MaxPageSize:        getEnvInt("MAX_PAGE_SIZE", 50),
		MaxQueryLength:     getEnvInt("MAX_QUERY_LENGTH", 200),
		SuggestLimit:       getEnvInt("SUGGEST_LIMIT", 5),
		SearchTimeout:      getEnvSeconds("SEARCH_TIMEOUT_SECONDS", 5*time.Second),
		StartupTimeout:     getEnvSeconds("STARTUP_TIMEOUT_SECONDS", 30*time.Second),
	}

	if cfg.FuzzyTolerance < 1 {
		return nil, fmt.Errorf("config: FUZZY_TOLERANCE must be >= 1, got %d", cfg.FuzzyTolerance)
	}
	if cfg.DefaultPageSize < 1 || cfg.DefaultPageSize > cfg.MaxPageSize {
		return nil, fmt.Errorf("config: DEFAULT_PAGE_SIZE (%d) must be between 1 and MAX_PAGE_SIZE (%d)", cfg.DefaultPageSize, cfg.MaxPageSize)
	}
	if cfg.MaxQueryLength < 1 {
		return nil, fmt.Errorf("config: MAX_QUERY_LENGTH must be >= 1, got %d", cfg.MaxQueryLength)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvSeconds(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return time.Duration(n) * time.Second
}

func getEnvList(key string, fallback []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}