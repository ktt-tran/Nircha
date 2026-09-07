package handlers

import (
	"context"
	"net/http"
	"time"

	"search-engine/searcher/models"
)

// HandleHealth serves GET /api/health. It checks real Redis connectivity
// (not just "the process is running"), so a Docker healthcheck, uptime
// monitor.
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.repo.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, models.HealthResponse{
			Status: "degraded",
			Redis:  "unreachable",
		})
		return
	}

	writeJSON(w, http.StatusOK, models.HealthResponse{
		Status: "ok",
		Redis:  "ok",
	})
}