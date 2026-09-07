package tokenizer

import "strings"

// defaultLanguage is used for every query. Per-query language detection

// Sanitize, RemoveStopWords, and StemWords supports multi languag by adding 
// to stopwords directory.
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