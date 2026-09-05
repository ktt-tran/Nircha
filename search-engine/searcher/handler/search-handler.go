package handlers

import (
	"context"
	"net/http"
	"time"

	"search-engine/ranker/ranker"
	"search-engine/searcher/models"
	"search-engine/searcher/pipes"
	"search-engine/searcher/tokenizer"
)

// HandleSearch serves GET /api/search?q=...&page=...&size=...&major=...
// &degree=...&location=....
//
// The pipeline: Search (per-token Redis fan-out, unchanged) -> Merge
// (aggregate hit stats per page, unchanged) -> Resolve (batch-fetch full
// listing data, new) -> Filter (degree/location hard filter, new) ->
// ranker.Rank (the composite formula, new - replaces the old Rank+Sort
// placeholder) -> Fill (paginate, now pure in-memory slicing). See
// pipes.Resolve's doc comment for why full data has to be fetched before
// ranking now, unlike the old pipeline.
func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	timeStart := time.Now()

	query, err := parseQueryText(r.URL.Query().Get("q"), h.cfg.MaxQueryLength)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	page, err := parsePageParam(r.URL.Query().Get("page"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	size, err := parseSizeParam(r.URL.Query().Get("size"), h.cfg.DefaultPageSize, h.cfg.MaxPageSize)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	profile := parseProfile(r, h.cfg.MaxQueryLength)

	// Bounds how long the per-token Redis fan-out and the batch page
	// resolve below are allowed to run. Without this, a slow/hung Redis
	// connection would let the request hang until the client or a load
	// balancer gave up, rather than failing fast with a clear error.
	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.SearchTimeout)
	defer cancel()

	tokens := tokenizer.Tokenize(query)

	tokenResults, _ := h.searchService.Search(ctx, tokens)
	if ctx.Err() != nil {
		writeError(w, http.StatusGatewayTimeout, "search timed out, please try again")
		return
	}

	merged := pipes.Merge(&tokenResults)
	candidates := pipes.Resolve(ctx, h.repo, merged)
	if ctx.Err() != nil {
		writeError(w, http.StatusGatewayTimeout, "search timed out, please try again")
		return
	}

	eligible := pipes.Filter(candidates, profile)

	ranked := ranker.Rank(h.rankerConfig, query, eligible, profile, time.Now())

	results := pipes.Fill(ranked, size, page)

	totalResults := len(ranked)
	pagesCount := (totalResults + size - 1) / size // integer ceiling division; size >= 1 is guaranteed above

	writeJSON(w, http.StatusOK, models.SearchResponse{
		Query:        query,
		Pages:        results,
		TotalResults: totalResults,
		Page:         page,
		PageSize:     size,
		PagesCount:   pagesCount,
		TimeMs:       time.Since(timeStart).Milliseconds(),
	})
}