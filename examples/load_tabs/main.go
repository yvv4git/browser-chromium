// Command load_tabs opens several tabs at once, holds them for a random time
// and closes them in a loop. It reproduces the load pattern used to measure
// the RAM and CPU of the Chromium container.
//
// The container must be running first (make compose-up).
//
//	go run ./load_tabs -addr http://localhost:9222
package main

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"github.com/yvv4git/browser-chromium/examples/internal/cdp"
)

const (
	defaultAddr    = "http://localhost:9222"
	defaultTabs    = 5
	defaultCycles  = 10
	defaultHoldMin = 10 * time.Second
	defaultHoldMax = 40 * time.Second
)

var targets = []string{
	"https://www.wikipedia.org",
	"https://example.com",
	"https://go.dev",
	"https://www.python.org",
	"https://www.kernel.org",
}

func main() {
	addr := flag.String("addr", defaultAddr, "CDP endpoint of the Chromium container")
	tabs := flag.Int("tabs", defaultTabs, "how many tabs to open at once")
	cycles := flag.Int("cycles", defaultCycles, "how many open-hold-close cycles")
	holdMin := flag.Duration("hold-min", defaultHoldMin, "minimum tab hold time")
	holdMax := flag.Duration("hold-max", defaultHoldMax, "maximum tab hold time")
	flag.Parse()

	browser := cdp.Connect(*addr)
	closeStrays(browser)

	for cycle := range *cycles {
		hold := randomHold(*holdMin, *holdMax)
		pages := openTabs(browser, *tabs)
		fmt.Printf("cycle %d: opened %d tabs, hold %s\n", cycle+1, len(pages), hold)

		time.Sleep(hold)
		closePages(pages)
		fmt.Printf("cycle %d: closed\n", cycle+1)
	}
}

// closeStrays keeps the first page as an anchor and closes pages left by
// previous runs, so every run starts from the same tab count.
func closeStrays(browser *rod.Browser) {
	pages, err := browser.Pages()
	if err != nil {
		fmt.Printf("list pages: %v\n", err)

		return
	}
	if len(pages) <= 1 {
		return
	}

	for _, page := range pages[1:] {
		if closeErr := page.Close(); closeErr != nil {
			fmt.Printf("close stray: %v\n", closeErr)
		}
	}
	fmt.Printf("strays closed: %d pages left\n", len(pages)-1)
}

// openTabs opens up to count tabs at once and skips targets that fail.
func openTabs(browser *rod.Browser, count int) []*rod.Page {
	pages := make([]*rod.Page, 0, count)
	for index := range count {
		target := targets[index%len(targets)]
		page, openErr := browser.Page(proto.TargetCreateTarget{URL: target})
		if openErr != nil {
			fmt.Printf("open failed: %s: %v\n", target, openErr)

			continue
		}
		pages = append(pages, page)
	}

	return pages
}

// closePages closes every page and reports failures.
func closePages(pages []*rod.Page) {
	for _, page := range pages {
		if closeErr := page.Close(); closeErr != nil {
			fmt.Printf("close failed: %v\n", closeErr)
		}
	}
}

// randomHold returns a random hold duration in [min, max].
func randomHold(min, max time.Duration) time.Duration {
	if max <= min {
		return min
	}

	return min + time.Duration(rand.Int64N(int64(max-min)))
}
