package tokenizer

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

var stopWordsFS embed.FS

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

// stopWords caches the parsed stopword set for each language that has been
// requested so far.
var (
	stopWordsMu sync.RWMutex
	stopWords   = map[string]map[string]struct{}{}
)

func RemoveStopWords(words []string, language string) ([]string, error) {

	language = strings.ToLower(language)

	currentStopWords, err := getStopWords(language)
	if err != nil {
		return words, err
	}

	if len(currentStopWords) == 0 {
		return words, nil
	}

	result := []string{}

	for _, word := range words {

		if _, ok := currentStopWords[word]; ok {
			continue
		}
		result = append(result, word)
	}

	return result, nil
}

// getStopWords returns the cached stopword set for language, loading and
// caching it on first use.
func getStopWords(language string) (map[string]struct{}, error) {
	stopWordsMu.RLock()
	sw, ok := stopWords[language]
	stopWordsMu.RUnlock()
	if ok {
		return sw, nil
	}

	stopWordsMu.Lock()
	defer stopWordsMu.Unlock()

	if sw, ok := stopWords[language]; ok {
		return sw, nil
	}

	sw, err := loadStopWords(language)
	if err != nil {
		return nil, err
	}
	stopWords[language] = sw
	return sw, nil
}

// loadStopWords reads tokenizer/stopwords/<lang>.json for the given language
// name (e.g. "english") out of the binary's embedded filesystem.
func loadStopWords(language string) (map[string]struct{}, error) {

	lng, ok := languageAbrv[language]
	if !ok {
		return nil, fmt.Errorf("no stopword list available for language %q", language)
	}

	data, err := stopWordsFS.ReadFile("stopwords/" + lng + ".json")
	if err != nil {
		return nil, err
	}

	var sw []string
	if err := json.Unmarshal(data, &sw); err != nil {
		return nil, err
	}

	mapSw := make(map[string]struct{}, len(sw))
	for _, word := range sw {
		mapSw[word] = struct{}{}
	}

	return mapSw, nil
}