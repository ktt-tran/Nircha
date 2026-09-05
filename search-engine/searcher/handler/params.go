package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"search-engine/ranker/document"
)

// parseQueryText validates the `q` parameter shared by /api/search and
// /api/suggest: required, trimmed, and bounded in length.
func parseQueryText(raw string, maxLen int) (string, error) {
	q := strings.TrimSpace(raw)
	if q == "" {
		return "", fmt.Errorf("'q' is required")
	}
	if utf8.RuneCountInString(q) > maxLen {
		return "", fmt.Errorf("'q' must be %d characters or fewer", maxLen)
	}
	return q, nil
}

// parsePageParam validates the `page` parameter: an omitted value defaults
// to 1, but an explicitly invalid one (non-numeric, zero, or negative) is a
// 400 rather than being silently coerced.
func parsePageParam(raw string) (int, error) {
	if raw == "" {
		return 1, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("'page' must be a whole number")
	}
	if n < 1 {
		return 0, fmt.Errorf("'page' must be 1 or greater")
	}
	return n, nil
}

// parseSizeParam validates the `size` parameter against [1, maxSize].
func parseSizeParam(raw string, defaultSize, maxSize int) (int, error) {
	if raw == "" {
		return defaultSize, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("'size' must be a whole number")
	}
	if n < 1 {
		return 0, fmt.Errorf("'size' must be 1 or greater")
	}
	if n > maxSize {
		return 0, fmt.Errorf("'size' must be %d or less", maxSize)
	}
	return n, nil
}

// parseProfile reads the optional major/degree/location query params into
// a document.Profile. All three are optional - see document.Profile's doc
// comment - so there's nothing to reject here, only to trim. maxLen reuses
// MAX_QUERY_LENGTH as a defensive cap; these are meant to be short values
// like "computer science" or "Boston", not user-supplied prose.
func parseProfile(r *http.Request, maxLen int) document.Profile {
	get := func(key string) string {
		v := strings.TrimSpace(r.URL.Query().Get(key))
		if utf8.RuneCountInString(v) > maxLen {
			v = string([]rune(v)[:maxLen])
		}
		return v
	}
	return document.Profile{
		Major:       get("major"),
		DegreeLevel: get("degree"),
		Location:    get("location"),
	}
}