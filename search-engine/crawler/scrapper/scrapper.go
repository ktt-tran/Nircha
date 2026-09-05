package scrapper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	htmlparser "search-engine/crawler/parser"
	domainExtractor "search-engine/crawler/utils"
	"search-engine/crawler/webpage"

	"golang.org/x/net/html"
)

type Scrapper struct {
	httpClient *http.Client
	timeout    time.Duration
	userAgent  string
}

func NewScrapper(httpClient *http.Client, timeout time.Duration, userAgent string) *Scrapper {
	return &Scrapper{httpClient: httpClient, timeout: timeout, userAgent: userAgent}
}

func (s *Scrapper) Scrape(siteUrl string) (*webpage.WebPage, error) {

	outboundLinks := make(map[string]bool)
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", siteUrl, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", s.userAgent)

	res, err := s.httpClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("scrape %s: unexpected status %d", siteUrl, res.StatusCode)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}

	page := htmlparser.GetPageInfo(doc, siteUrl)

	for _, link := range page.Links {
		if _, err := domainExtractor.ExtractDomain(link); err == nil {
			outboundLinks[link] = true
		}
	}

	for domain := range outboundLinks {
		page.OutboundLinks = append(page.OutboundLinks, domain)
	}

	return page, nil
}