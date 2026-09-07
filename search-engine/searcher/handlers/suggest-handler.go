package handlers

import (
	"net/http"

	"search-engine/searcher/models"
	"search-engine/searcher/tokenizer"
)

// HandleSuggest serves GET /api/suggest?q=...
//
// It returns spelling-correction candidates for tokens that don't already
// match the index - "did you mean X" rather than prefix autocomplete
// (that's a different problem needing a different data structure; the
// fuzzy searcher underneath this is a BK-tree indexed by edit distance,
// which is well suited to typo correction, not prefix completion).
func (h *Handler) HandleSuggest(w http.ResponseWriter, r *http.Request) {
	query, err := parseQueryText(r.URL.Query().Get("q"), h.cfg.MaxQueryLength)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	tokens := tokenizer.Tokenize(query)
	suggestions := h.searchService.Suggest(r.Context(), tokens, h.cfg.SuggestLimit)

	writeJSON(w, http.StatusOK, models.SuggestResponse{
		Query:       query,
		Suggestions: suggestions,
	})
}