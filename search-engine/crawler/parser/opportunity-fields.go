package htmlparser

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"search-engine/crawler/webpage"
)

type jsonLDJobPosting struct {
	Type                  string              `json:"@type"`
	DatePosted            string              `json:"datePosted"`
	ValidThrough          string              `json:"validThrough"`
	JobLocationType       string              `json:"jobLocationType"`
	HiringOrganization    *jsonLDOrganization `json:"hiringOrganization"`
	JobLocation           json.RawMessage     `json:"jobLocation"`
	EducationRequirements json.RawMessage     `json:"educationRequirements"`
}

type jsonLDOrganization struct {
	Name string `json:"name"`
}

type jsonLDPlace struct {
	Address *jsonLDAddress `json:"address"`
}

type jsonLDAddress struct {
	AddressLocality string `json:"addressLocality"`
	AddressRegion   string `json:"addressRegion"`
}

var jsonLDDateLayouts = []string{
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02",
}

func parseJSONLDDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range jsonLDDateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func extractJobLocationString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var single jsonLDPlace
	if err := json.Unmarshal(raw, &single); err == nil && single.Address != nil {
		if loc := formatAddress(single.Address); loc != "" {
			return loc
		}
	}

	var multiple []jsonLDPlace
	if err := json.Unmarshal(raw, &multiple); err == nil {
		for _, place := range multiple {
			if place.Address == nil {
				continue
			}
			if loc := formatAddress(place.Address); loc != "" {
				return loc
			}
		}
	}

	return ""
}

func formatAddress(a *jsonLDAddress) string {
	switch {
	case a.AddressLocality != "" && a.AddressRegion != "":
		return a.AddressLocality + ", " + a.AddressRegion
	case a.AddressLocality != "":
		return a.AddressLocality
	case a.AddressRegion != "":
		return a.AddressRegion
	default:
		return ""
	}
}

func extractEducationRequirementString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var obj struct {
		CredentialCategory string `json:"credentialCategory"`
		Name               string `json:"name"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil {
		if obj.CredentialCategory != "" {
			return obj.CredentialCategory
		}
		return obj.Name
	}
	return ""
}

func parseJobPosting(raw string) (*jsonLDJobPosting, bool) {
	var single jsonLDJobPosting
	if err := json.Unmarshal([]byte(raw), &single); err == nil && strings.EqualFold(single.Type, "JobPosting") {
		return &single, true
	}

	var multiple []jsonLDJobPosting
	if err := json.Unmarshal([]byte(raw), &multiple); err == nil {
		for i := range multiple {
			if strings.EqualFold(multiple[i].Type, "JobPosting") {
				return &multiple[i], true
			}
		}
	}

	return nil, false
}

var fieldsOfStudyKeywords = []string{
	"computer science", "software engineering", "electrical engineering",
	"mechanical engineering", "civil engineering", "chemical engineering",
	"biomedical engineering", "data science", "statistics", "mathematics",
	"physics", "chemistry", "biology", "computational biology",
	"environmental science", "public health", "economics", "business",
	"finance", "psychology", "political science", "sociology", "history",
	"english", "linguistics", "art", "design",
}

var degreeLevelKeywords = map[string][]string{
	"undergraduate": {"undergraduate", "bachelor's", "bachelors", "bachelor"},
	"graduate":      {"graduate student", "master's", "masters", "master's degree"},
	"phd":           {"phd", "ph.d", "doctoral", "doctorate"},
	"postdoc":       {"postdoc", "postdoctoral", "post-doctoral"},
}

var remoteKeywords = []string{
	"remote", "work from home", "telecommute", "virtual position", "work-from-home",
}

type keywordMatcher struct {
	label string
	re    *regexp.Regexp
}

func buildMatchers(keywords []string) []keywordMatcher {
	matchers := make([]keywordMatcher, len(keywords))
	for i, kw := range keywords {
		matchers[i] = keywordMatcher{label: kw, re: regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(kw) + `s?\b`)}
	}
	return matchers
}

var (
	fieldsOfStudyMatchers = buildMatchers(fieldsOfStudyKeywords)
	remoteMatchers        = buildMatchers(remoteKeywords)
	degreeLevelMatchers   = buildDegreeLevelMatchers(degreeLevelKeywords)
)

type degreeLevelMatcher struct {
	canonical string
	matchers  []keywordMatcher
}

func buildDegreeLevelMatchers(byLevel map[string][]string) []degreeLevelMatcher {
	result := make([]degreeLevelMatcher, 0, len(byLevel))
	for canonical, keywords := range byLevel {
		result = append(result, degreeLevelMatcher{canonical: canonical, matchers: buildMatchers(keywords)})
	}
	return result
}

// deadlinePhrases are checked for proximity to a date match - a bare date
// elsewhere on the page (an event date, a founding year) isn't a deadline
// just because it happens to parse as one.
var deadlinePhrases = []string{
	"deadline", "apply by", "applications close", "applications due",
	"due by", "must apply by",
}

// datePattern matches "Month D, YYYY" / "Month D YYYY" (case-insensitive
// month name) and "M/D/YYYY" style dates - the two formats real
// opportunity pages overwhelmingly use in free text. ISO dates in free
// text are rare enough outside JSON-LD to not be worth a third branch.
var datePattern = regexp.MustCompile(
	`(?i)(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d{1,2}),?\s+(\d{4})|(\d{1,2})/(\d{1,2})/(\d{4})`,
)

var monthNames = map[string]time.Month{
	"january": time.January, "february": time.February, "march": time.March,
	"april": time.April, "may": time.May, "june": time.June, "july": time.July,
	"august": time.August, "september": time.September, "october": time.October,
	"november": time.November, "december": time.December,
}

func extractDeadlineHeuristic(text string) (time.Time, bool) {
	lower := strings.ToLower(text)

	for _, phrase := range deadlinePhrases {
		idx := strings.Index(lower, phrase)
		if idx == -1 {
			continue
		}
		windowEnd := idx + len(phrase) + 60
		if windowEnd > len(text) {
			windowEnd = len(text)
		}
		window := text[idx:windowEnd]

		if t, ok := parseFreeTextDate(window); ok {
			return t, true
		}
	}
	return time.Time{}, false
}

func parseFreeTextDate(text string) (time.Time, bool) {
	match := datePattern.FindStringSubmatch(text)
	if match == nil {
		return time.Time{}, false
	}

	// Two alternatives in datePattern: "Month D, YYYY" (groups 1-3) or
	// "M/D/YYYY" (groups 4-6). Exactly one set is populated per match.
	if match[1] != "" {
		month, ok := monthNames[strings.ToLower(match[1])]
		if !ok {
			return time.Time{}, false
		}
		day, errDay := strconv.Atoi(match[2])
		year, errYear := strconv.Atoi(match[3])
		if errDay != nil || errYear != nil {
			return time.Time{}, false
		}
		return time.Date(year, month, day, 0, 0, 0, 0, time.UTC), true
	}

	if match[4] != "" {
		monthNum, errMonth := strconv.Atoi(match[4])
		day, errDay := strconv.Atoi(match[5])
		year, errYear := strconv.Atoi(match[6])
		if errMonth != nil || errDay != nil || errYear != nil || monthNum < 1 || monthNum > 12 {
			return time.Time{}, false
		}
		return time.Date(year, time.Month(monthNum), day, 0, 0, 0, 0, time.UTC), true
	}

	return time.Time{}, false
}

func extractFieldsOfStudyHeuristic(text string) []string {
	var found []string
	for _, m := range fieldsOfStudyMatchers {
		if m.re.MatchString(text) {
			found = append(found, m.label)
		}
	}
	return found
}

func extractDegreeLevelsHeuristic(text string) []string {
	var found []string
	for _, dm := range degreeLevelMatchers {
		for _, m := range dm.matchers {
			if m.re.MatchString(text) {
				found = append(found, dm.canonical)
				break
			}
		}
	}
	return found
}

func extractRemoteHeuristic(text string) bool {
	for _, m := range remoteMatchers {
		if m.re.MatchString(text) {
			return true
		}
	}
	return false
}

// --- orchestration ---

// extractOpportunityFields populates the opportunity-specific fields on page
func extractOpportunityFields(page *webpage.WebPage, jsonLDBlocks []string) {
	var jobPosting *jsonLDJobPosting
	for _, block := range jsonLDBlocks {
		if jp, ok := parseJobPosting(block); ok {
			jobPosting = jp
			break
		}
	}

	if jobPosting != nil {
		if t, ok := parseJSONLDDate(jobPosting.DatePosted); ok {
			page.PostedAt = t
		}
		if t, ok := parseJSONLDDate(jobPosting.ValidThrough); ok {
			page.DeadlineAt = &t
		}
		if jobPosting.HiringOrganization != nil && jobPosting.HiringOrganization.Name != "" {
			page.Organization = jobPosting.HiringOrganization.Name
		}
		if loc := extractJobLocationString(jobPosting.JobLocation); loc != "" {
			page.Location = loc
		}
		if strings.EqualFold(jobPosting.JobLocationType, "TELECOMMUTE") {
			page.Remote = true
		}
		if edu := extractEducationRequirementString(jobPosting.EducationRequirements); edu != "" {
			if levels := extractDegreeLevelsHeuristic(edu); len(levels) > 0 {
				page.DegreeLevels = levels
			}
		}
	}

	combinedText := strings.Join([]string{page.Title, page.Abstract, page.Content}, " . ")

	if page.DeadlineAt == nil {
		if t, ok := extractDeadlineHeuristic(combinedText); ok {
			page.DeadlineAt = &t
		}
	}
	if len(page.FieldsOfStudy) == 0 {
		page.FieldsOfStudy = extractFieldsOfStudyHeuristic(combinedText)
	}
	if len(page.DegreeLevels) == 0 {
		page.DegreeLevels = extractDegreeLevelsHeuristic(combinedText)
	}
	if !page.Remote {
		page.Remote = extractRemoteHeuristic(combinedText)
	}
}