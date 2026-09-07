package handlers

import (
	"log"
	"net/http"
	"slices"
	"time"
)

// CORS returns middleware that allows the configured origins to call this
// API from a browser.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				switch {
				case allowAll:
					w.Header().Set("Access-Control-Allow-Origin", "*")
				case slices.Contains(allowedOrigins, origin):
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Recover returns middleware that catches a panic anywhere in the wrapped
// handler and responds with a clean {"error": "..."} 500 instead of
// net/http's default behavior of logging it and abruptly closing the
// connection.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("handlers: panic serving %s %s: %v", r.Method, r.URL.Path, rec)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// responseRecorder captures the status code written through it, since
// http.ResponseWriter has no way to ask afterward what was sent - Logging
// needs it for the access-log line below.
type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *responseRecorder) WriteHeader(status int) {
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

// Logging returns middleware that logs one line per request: method, path,
// status code, and duration.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

// Chain applies middleware around h in the order given: Chain(h, A, B, C)
// wraps as A(B(C(h))), so for an incoming request A runs first, then B,
// then C, then h.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}