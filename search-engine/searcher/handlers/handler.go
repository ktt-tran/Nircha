// Package handlers implements the HTTP handlers for the searcher REST API.
package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"search-engine/ranker/ranker"
	"search-engine/searcher/config"
	fuzzysearch "search-engine/searcher/fuzzy-search"
	"search-engine/searcher/models"
	"search-engine/searcher/repository"
	"search-engine/searcher/services"
)

type Handler struct {
	searchService *services.SearchService
	fuzzySearcher *fuzzysearch.FuzzySearcher
	repo          *repository.RedisRepository
	cfg           *config.Config
	rankerConfig  ranker.Config
}

func NewHandler(
	searchService *services.SearchService,
	fuzzySearcher *fuzzysearch.FuzzySearcher,
	repo *repository.RedisRepository,
	cfg *config.Config,
) *Handler {
	
	rankerConfig := ranker.DefaultConfig()
	if err := rankerConfig.Validate(); err != nil {
		log.Fatalf("handlers: invalid ranker config: %v", err)
	}

	return &Handler{
		searchService: searchService,
		fuzzySearcher: fuzzySearcher,
		repo:          repo,
		cfg:           cfg,
		rankerConfig:  rankerConfig,
	}
}

// writeJSON writes payload as a JSON response with the given status code.
// Every handler in this package should return through here (or writeError)
// so the response envelope and Content-Type header stay consistent.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// The status line and headers are already sent at this point, so
		// there's nothing left to do but log it - the client will just
		// see a truncated body.
		log.Printf("handlers: encoding response: %v", err)
	}
}

// writeError writes a consistent {"error": "..."} JSON body. message should
// be a client-safe description: no internal error text, stack traces, or
// Redis error details, which get logged server-side instead.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, models.ErrorResponse{Error: message})
}