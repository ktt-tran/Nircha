package htmlparser

import (
	"net/url"
	"strings"

	"search-engine/crawler/tokenizer"
	"search-engine/crawler/webpage"

	"golang.org/x/net/html"
)

const maxContentLength = 20000

func GetPageInfo(doc *html.Node, pageURL string) *webpage.WebPage {
	page := &webpage.WebPage{URL: pageURL}

	// base is used to resolve relative hrefs (see resolveLink). If pageURL
	// itself somehow fails to parse, base is nil and resolveLink safely
	// rejects every href rather than panicking.
	base, _ := url.Parse(pageURL)

	currentPosition := 0
	jsonLDBlocks := doTraverse(doc, page, &currentPosition, base)

	if len(page.Content) > maxContentLength {
		page.Content = page.Content[:maxContentLength]
	}

	// Structured-data extraction (dates, location, field of study, degree
	// level, remote) runs after traversal, once page.Title/Abstract/Content
	// are final.
	extractOpportunityFields(page, jsonLDBlocks)

	return page
}

// resolveLink turns a raw href attribute value into an absolute, crawlable
// URL, or reports ok=false if it isn't one.
func resolveLink(base *url.URL, href string) (string, bool) {
	href = strings.TrimSpace(href)
	if href == "" || base == nil {
		return "", false
	}

	ref, err := url.Parse(href)
	if err != nil {
		return "", false
	}

	resolved := base.ResolveReference(ref)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", false
	}
	resolved.Fragment = ""

	return resolved.String(), true
}

// textContent concatenates all descendant text nodes under n.
func textContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(textContent(c))
	}
	return sb.String()
}

func doTraverse(doc *html.Node, page *webpage.WebPage, currentPosition *int, base *url.URL) []string {

	var metaDescription, ogDescription, twitterDescription, firstParagraph string

	var jsonLDBlocks []string

	textElements := map[string]bool{
		"title": true,
		"h1":    true,
		"h2":    true,
		"h3":    true,
		"h4":    true,
		"h5":    true,
		"h6":    true,
		"p":     true,
	}

	var traverse func(n *html.Node)
	traverse = func(n *html.Node) {

		for c := n.FirstChild; c != nil; c = c.NextSibling {

			if c.Type == html.ElementNode && c.Data == "a" {
				for _, a := range c.Attr {
					if a.Key == "href" {
						if resolved, ok := resolveLink(base, a.Val); ok {
							page.Links = append(page.Links, resolved)
						}
					}
				}
			}

			if c.Type == html.ElementNode && c.Data == "title" && page.Title == "" {
				page.Title = strings.TrimSpace(textContent(c))
			}

			if c.Type == html.ElementNode && c.Data == "script" {
				for _, a := range c.Attr {
					if a.Key == "type" && a.Val == "application/ld+json" {
						if text := strings.TrimSpace(textContent(c)); text != "" {
							jsonLDBlocks = append(jsonLDBlocks, text)
						}
						break
					}
				}
			}

			if c.Type == html.ElementNode && c.Data == "meta" {
				attrs := make(map[string]string, len(c.Attr))
				for _, a := range c.Attr {
					attrs[a.Key] = a.Val
				}
				switch {
				case attrs["name"] == "description" && metaDescription == "":
					metaDescription = attrs["content"]
				case attrs["property"] == "og:description" && ogDescription == "":
					ogDescription = attrs["content"]
				case attrs["name"] == "twitter:description" && twitterDescription == "":
					twitterDescription = attrs["content"]
				}
			}

			if c.Type == html.ElementNode && c.Data == "p" && firstParagraph == "" {
				if text := strings.TrimSpace(textContent(c)); len(text) > 20 {
					firstParagraph = text
				}
			}

			if c.Type == html.ElementNode && textElements[c.Data] {
				extractWords(c, page, currentPosition)

				if text := textContent(c); text != "" {
					if page.Content != "" {
						page.Content += " "
					}
					page.Content += text
				}
			}

			traverse(c)
		}
	}

	traverse(doc)

	switch {
	case metaDescription != "":
		page.Abstract = metaDescription
	case ogDescription != "":
		page.Abstract = ogDescription
	case twitterDescription != "":
		page.Abstract = twitterDescription
	default:
		page.Abstract = firstParagraph
	}
	if len(page.Abstract) > 150 {
		page.Abstract = page.Abstract[:150]
	}

	return jsonLDBlocks
}

func extractWords(c *html.Node, page *webpage.WebPage, currentPosition *int) {

	scoreTable := map[string]int{
		"title": 20,
		"h1":    10,
		"h2":    5,
		"h3":    4,
		"h4":    3,
		"h5":    2,
		"p":     1,
	}

	sentence := textContent(c)
	if sentence == "" {
		return
	}

	wordsMap := make(map[string]struct {
		Score    int
		Position int
	})

	words := tokenizer.Tokenize(sentence)
	tag := c.Data

	for _, word := range words {
		if word == "" {
			continue
		}
		if wordInfo, ok := wordsMap[word]; ok && scoreTable[tag] < wordInfo.Score {
			continue
		}
		wordsMap[word] = struct {
			Score    int
			Position int
		}{Score: scoreTable[tag], Position: *currentPosition}
		(*currentPosition)++
	}
	for word, wordData := range wordsMap {
		page.Words = append(page.Words, webpage.Word{Word: word, Score: wordData.Score, Position: wordData.Position})
	}
}