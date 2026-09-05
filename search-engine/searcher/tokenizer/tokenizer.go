package tokenizer

import "strings"

// defaultLanguage is used for every query. Per-query language detection

// Sanitize, RemoveStopWords, and StemWords all still take a language
// parameter, so real multi-language support - e.g. accepting an explicit
// `lang` query parameter from the API once more stopword/stemmer data is
// bundled - can be layered back in without restructuring this pipeline.
const defaultLanguage = "english"

// Tokenize turns free-text search input into a normalized slice of terms:
// whitespace-split, sanitized, stopword-filtered, and stemmed.
func Tokenize(text string) []string {
	result := strings.Fields(text)
	result = Sanitize(result, defaultLanguage)
	result = RemoveStopWords(result, defaultLanguage)
	result = StemWords(result, defaultLanguage)
return result
}