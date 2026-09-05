package crawler

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func (c *Crawler) showStatus(pagesAdded *atomic.Int64, maxPages int, done <-chan struct{}) {

	timeStart := time.Now()
	p := message.NewPrinter(language.English)

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
		}

		i := pagesAdded.Load()

		percentage := 0
		if maxPages > 0 {
			percentage = int((i * 100) / int64(maxPages))
		}
		if percentage > 100 {
			percentage = 100
		} else if percentage < 0 {
			percentage = 0
		}

		fmt.Print("\033[H\033[2J")
		p.Printf("Crawling...\n%d / %d\n", i, maxPages)
		fmt.Printf("[%s%s]%d%%\n", strings.Repeat("=", percentage), strings.Repeat(" ", 100-percentage), percentage)

		if i > 0 {
			timeElapsed := time.Since(timeStart).Seconds()
			timeRemaining := (timeElapsed / float64(i)) * float64(int64(maxPages)-i)
			if timeRemaining < 0 {
				timeRemaining = 0
			}
			hoursRemaining := int(timeRemaining) / 3600
			minutesRemaining := int(timeRemaining) % 3600 / 60
			secondsRemaining := int(timeRemaining) % 60
			fmt.Printf("Time remaining: %02dh:%02dm:%02ds\n", hoursRemaining, minutesRemaining, secondsRemaining)
		} else {
			fmt.Println("Time remaining: calculating...")
		}
	}
}