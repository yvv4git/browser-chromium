// Command links opens a page in the Chromium container, extracts the link
// list and prints the first links with their anchors, filtering out static
// resources and anchor links.
//
// The container must be running first (make compose-up).
//
//	go run ./links -addr http://localhost:9222 -url http://books.toscrape.com
package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

var staticExt = []string{".png", ".jpg", ".jpeg", ".gif", ".svg", ".css", ".js", ".woff", ".woff2", ".pdf", ".zip"}

func isStaticExt(link string) bool {
	ext := strings.ToLower(path.Ext(link))
	for _, s := range staticExt {
		if ext == s {
			return true
		}
	}
	return false
}

func main() {
	addr := flag.String("addr", "http://localhost:9222", "CDP endpoint of the Chromium container")
	baseURL := flag.String("url", "http://books.toscrape.com", "page to scan")
	limit := flag.Int("limit", 20, "max links to print")
	flag.Parse()

	browser := rod.New().ControlURL(*addr).MustConnect()

	page := browser.MustPage(*baseURL).MustWaitLoad()
	defer page.MustClose()

	if err := page.WaitStable(3 * time.Second); err != nil {
		log.Fatal(err)
	}

	links, err := page.Elements("a")
	if err != nil {
		log.Fatal(err)
	}

	base, err := url.Parse(*baseURL)
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	for _, link := range links {
		text := strings.TrimSpace(link.MustText())
		href, err := link.Attribute("href")
		if err != nil || href == nil || *href == "" {
			continue
		}
		if strings.HasPrefix(*href, "#") || isStaticExt(*href) {
			continue
		}

		ref, err := url.Parse(*href)
		if err != nil {
			continue
		}
		abs := base.ResolveReference(ref).String()
		if !strings.HasPrefix(abs, "http") {
			continue
		}

		fmt.Printf("  [%s] %s\n", text, abs)
		count++
		if count == *limit {
			break
		}
	}

	fmt.Printf("total links: %d, shown: %d\n", len(links), count)
}