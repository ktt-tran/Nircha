package tokenizer

import (
	"strings"

	"github.com/rylans/getlang"
)

func Tokenize(text string) []string {
	info := getlang.FromString(text)
	language := info.LanguageName()
	confidence := info.Confidence()

	if confidence < 0.6 {
		language = "english"
	}

	result := strings.Fields(text)
	result = Sanitize(result, language)

	if filtered, err := RemoveStopWords(result, language); err == nil {
		result = filtered
	}

	result = StemWords(result, language)

	return result
}