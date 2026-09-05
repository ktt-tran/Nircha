package tokenizer

import (
"embed"
"encoding/json"
"fmt"
"strings"
"sync"
)

var stopwordsFS embed.FS

var languageAbrv = map[string]string{
"afrikaans":      "af",
"albanian":       "sq",
"arabic":         "ar",
"armenian":       "hy",
"azerbaijani":    "az",
"basque":         "eu",
"belarusian":     "be",
"bengali":        "bn",
"bosnian":        "bs",
"bulgarian":      "bg",
"catalan":        "ca",
"chinese":        "zh",
"croatian":       "hr",
"czech":          "cs",
"danish":         "da",
"dutch":          "nl",
"english":        "en",
"esperanto":      "eo",
"estonian":       "et",
"finnish":        "fi",
"french":         "fr",
"galician":       "gl",
"georgian":       "ka",
"german":         "de",
"greek":          "el",
"gujarati":       "gu",
"haitian creole": "ht",
"hebrew":         "he",
"hindi":          "hi",
"hungarian":      "hu",
"icelandic":      "is",
"indonesian":     "id",
"irish":          "ga",
"italian":        "it",
"japanese":       "ja",
"javanese":       "jv",
"kannada":        "kn",
"kazakh":         "kk",
"khmer":          "km",
"korean":         "ko",
"kurdish":        "ku",
"kyrgyz":         "ky",
"lao":            "lo",
"latin":          "la",
"latvian":        "lv",
"lithuanian":     "lt",
"luxembourgish":  "lb",
"macedonian":     "mk",
"malagasy":       "mg",
"malay":          "ms",
"malayalam":      "ml",
"maltese":        "mt",
"maori":          "mi",
"marathi":        "mr",
"mongolian":      "mn",
"myanmar":        "my",
"nepali":         "ne",
"norwegian":      "no",
"odia":           "or",
"pashto":         "ps",
"persian":        "fa",
"polish":         "pl",
"portuguese":     "pt",
"punjabi":        "pa",
"romanian":       "ro",
"russian":        "ru",
"samoan":         "sm",
"scots gaelic":   "gd",
"serbian":        "sr",
"sesotho":        "st",
"shona":          "sn",
"sindhi":         "sd",
"sinhala":        "si",
"slovak":         "sk",
"slovenian":      "sl",
"somali":         "so",
"spanish":        "es",
"sundanese":      "su",
"swahili":        "sw",
"swedish":        "sv",
"tajik":          "tg",
"tamil":          "ta",
"tatar":          "tt",
"telugu":         "te",
"thai":           "th",
"turkish":        "tr",
"turkmen":        "tk",
"ukrainian":      "uk",
"urdu":           "ur",
"uyghur":         "ug",
"uzbek":          "uz",
"vietnamese":     "vi",
"welsh":          "cy",
"xhosa":          "xh",
"yiddish":        "yi",
"yoruba":         "yo",
"zulu":           "zu",
}

var (
	stopWordsMu sync.RWMutex
	stopWords   = map[string]map[string]struct{}{}
)

// RemoveStopWords filters common, low-information words (as bundled for the
// given language) out of words. If no stopword list is bundled for the
// language, words is returned unchanged - that is treated as a normal,
// expected case (most languages don't have a bundled list yet) rather than
// an error.
func RemoveStopWords(words []string, language string) []string {
	language = strings.ToLower(language)

	currentStopWords := getStopWords(language)

	if len(currentStopWords) == 0 {
			return words
	}

	result := make([]string, 0, len(words))
	for _, word := range words {
		if _, ok := currentStopWords[word]; ok {
			continue
		}
		result = append(result, word)
	}

	return result
}

// getStopWords returns the cached stopword set for language, loading and
// caching it on first use.
func getStopWords(language string) map[string]struct{} {
stopWordsMu.RLock()
sw, ok := stopWords[language]
stopWordsMu.RUnlock()
if ok {
		return sw
}

stopWordsMu.Lock()
defer stopWordsMu.Unlock()

if sw, ok := stopWords[language]; ok {
		return sw
}

sw, err := loadStopWords(language)
if err != nil {
		sw = map[string]struct{}{}
}
stopWords[language] = sw
	return sw
}

func loadStopWords(language string) (map[string]struct{}, error) {
	lng, ok := languageAbrv[language]
	if !ok {
		return nil, fmt.Errorf("tokenizer: unrecognized language %q", language)
	}

	data, err := stopwordsFS.ReadFile("stopwords/" + lng + ".json")
	if err != nil {
		return nil, fmt.Errorf("tokenizer: no bundled stopwords for language %q: %w", language, err)
	}

	var words []string
	if err := json.Unmarshal(data, &words); err != nil {
		return nil, fmt.Errorf("tokenizer: malformed stopwords file for language %q: %w", language, err)
	}

	sw := make(map[string]struct{}, len(words))
	for _, w := range words {
		sw[w] = struct{}{}
	}

	return sw, nil
}