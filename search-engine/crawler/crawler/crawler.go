package crawler

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"search-engine/crawler/repository"
	"search-engine/crawler/scrapper"
	domainExtractor "search-engine/crawler/utils"
	"search-engine/crawler/webqueue"
)

type Crawler struct {
	scrapper   *scrapper.Scrapper
	repository *repository.Repository
	mu         *sync.Mutex
	allowed *allowlist
}

const defaultTrustTier = 3

type allowlist struct {
	domains map[string]int
}

func newAllowlist(domains []string) *allowlist {
	a := &allowlist{domains: make(map[string]int)}
	for _, d := range domains {
		d = strings.ToLower(strings.TrimSpace(d))
		if d == "*" {
			return &allowlist{domains: make(map[string]int)} // empty map => unrestricted, see allowed()
		}
		if d == "" {
			continue
		}

		domain, tier := d, defaultTrustTier
		if idx := strings.LastIndex(d, ":"); idx != -1 {
			if parsed, err := strconv.Atoi(d[idx+1:]); err == nil && parsed >= 1 {
				domain, tier = d[:idx], parsed
			}
		}
		a.domains[domain] = tier
	}
	return a
}

func (a *allowlist) allowed(rawURL string) bool {
	if a == nil || len(a.domains) == 0 {
		return true
	}
	domain, err := domainExtractor.ExtractDomain(rawURL)
	if err != nil {
		return false
	}
	_, ok := a.domains[domain]
	return ok
}

func (a *allowlist) tierFor(rawURL string) int {
	if a == nil {
		return defaultTrustTier
	}
	domain, err := domainExtractor.ExtractDomain(rawURL)
	if err != nil {
		return defaultTrustTier
	}
	if tier, ok := a.domains[domain]; ok {
		return tier
	}
	return defaultTrustTier
}

func (c *Crawler) Crawl(maxPages, simultaneousRequests int, wq *webqueue.WebQueue) {

	showStatusFlag := flag.Bool("show-status", false, "Show status of the crawler")
	flag.Parse()

	sem := make(chan struct{}, simultaneousRequests)

	var pagesAdded atomic.Int64
	var inFlight atomic.Int64
	var wg sync.WaitGroup

	done := make(chan struct{})
	if *showStatusFlag {
		go c.showStatus(&pagesAdded, maxPages, done)
	}

	for pagesAdded.Load() < int64(maxPages) {

		c.mu.Lock()
		url := wq.Dequeue()
		c.mu.Unlock()

		if url == "" {
			if inFlight.Load() == 0 {
				break
			}
			time.Sleep(50 * time.Millisecond)
			continue
		}

		sem <- struct{}{}
		inFlight.Add(1)
		wg.Add(1)
		go func(pageURL string) {
			defer wg.Done()
			defer inFlight.Add(-1)
			defer func() { <-sem }()

			page, err := c.scrapper.Scrape(pageURL)
			if err != nil {
				return
			}

			page.LastCrawledAt = time.Now()
			page.TrustTier = c.allowed.tierFor(pageURL)

			for _, link := range page.Links {
				if _, err := domainExtractor.ExtractDomain(link); err != nil {
					continue
				}
				if !c.allowed.allowed(link) {
					continue
				}
				c.mu.Lock()
				wq.Enqueue(link)
				c.mu.Unlock()
			}

			added, err := (*c.repository).AddPage(page)
			if err != nil {
				return
			}
			if added {
				pagesAdded.Add(1)
			}
		}(url)
	}

	wg.Wait()
	close(done)
}

func NewCrawler(scrapper *scrapper.Scrapper, repo repository.Repository, allowedDomains []string) *Crawler {
	mu := &sync.Mutex{}
	return &Crawler{scrapper: scrapper, repository: &repo, mu: mu, allowed: newAllowlist(allowedDomains)}
}

func (c *Crawler) Start(firstURLs []string, maxPages, simultaneousRequests int) {

	wq := webqueue.NewWebQueue()
	for _, url := range firstURLs {
		wq.Enqueue(url)
	}

	now := time.Now()
	c.Crawl(maxPages, simultaneousRequests, wq)
	elapsed := time.Since(now)

	hours := int(elapsed.Hours())
	elapsed -= time.Duration(hours) * time.Hour
	minutes := int(elapsed.Minutes())
	elapsed -= time.Duration(minutes) * time.Minute
	seconds := int(elapsed.Seconds())

	fmt.Printf("Crawling time: %dh:%dm:%ds\n", hours, minutes, seconds)
}