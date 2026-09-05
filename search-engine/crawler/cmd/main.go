package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"search-engine/crawler/crawler"
	"search-engine/crawler/repository"
	"search-engine/crawler/scrapper"
	domainExtractor "search-engine/crawler/utils"

	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()

	var firstURLs []string
	for _, u := range strings.Split(os.Getenv("FIRST_WEB_URLS"), ",") {
		if u = strings.TrimSpace(u); u != "" {
			firstURLs = append(firstURLs, u)
		}
	}
	if len(firstURLs) == 0 {
		panic("no seed URLs configured: set FIRST_WEB_URLS in your .env file")
	}

	maxPagesStr := os.Getenv("MAX_PAGES")
	maxPages, err := strconv.Atoi(maxPagesStr)
	if err != nil {
		panic(err)
	}

	simultaneousRequestsStr := os.Getenv("SIMULTANEOUS_REQUESTS")
	simultaneousRequests, err := strconv.Atoi(simultaneousRequestsStr)
	if err != nil {
		panic(err)
	}

	timeoutSeconds, err := strconv.Atoi(os.Getenv("REQUEST_TIMEOUT_SECONDS"))
	if err != nil || timeoutSeconds <= 0 {
		timeoutSeconds = 2
	}

	userAgent := strings.TrimSpace(os.Getenv("CRAWLER_USER_AGENT"))
	if userAgent == "" {
		userAgent = "DinoSearchBot/0.1 (+mailto:REPLACE_ME@example.com)"
	}

	allowedDomains := parseAllowedDomains(os.Getenv("ALLOWED_DOMAINS"), firstURLs)

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DisableKeepAlives = true
	httpClient := &http.Client{Transport: t}

	s := scrapper.NewScrapper(httpClient, time.Duration(timeoutSeconds)*time.Second, userAgent)

	repo, err := repository.NewRedisRepository()
	if err != nil {
		panic(err)
	}

	c := crawler.NewCrawler(s, repo, allowedDomains)
	c.Start(firstURLs, maxPages, simultaneousRequests)
}

func parseAllowedDomains(raw string, seedURLs []string) []string {
	if raw = strings.TrimSpace(raw); raw != "" {
		var domains []string
		for _, d := range strings.Split(raw, ",") {
			if d = strings.TrimSpace(d); d != "" {
				domains = append(domains, d)
			}
		}
		return domains
	}

	seen := map[string]bool{}
	var domains []string
	for _, u := range seedURLs {
		if d, err := domainExtractor.ExtractDomain(u); err == nil && !seen[d] {
			seen[d] = true
			domains = append(domains, d)
		}
	}
	return domains
}