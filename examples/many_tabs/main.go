// Command many_tabs opens several tabs in the Chromium container, switches
// between them and lists all open tabs.
//
// The container must be running first (make compose-up).
//
//	go run ./many_tabs -addr http://localhost:9222
package main

import (
	"flag"
	"fmt"

	"github.com/go-rod/rod"

	"github.com/yvv4git/browser-chromium/examples/internal/cdp"
)

func main() {
	addr := flag.String("addr", "http://localhost:9222", "CDP endpoint of the Chromium container")
	flag.Parse()

	browser := cdp.Connect(*addr)

	tabs := []string{
		"https://www.wikipedia.org",
		"https://golang.org",
		"https://example.com",
	}

	var pages []*rod.Page
	for _, target := range tabs {
		page := browser.MustPage(target).MustWaitLoad()
		pages = append(pages, page)
		fmt.Printf("opened: %s\n", page.MustInfo().Title)
	}
	defer func() {
		for _, page := range pages {
			_ = page.Close()
		}
	}()

	pages[1].MustActivate()
	fmt.Println("activated tab 2")

	pages[0].MustActivate()
	fmt.Println("activated tab 1")

	open := browser.MustPages()
	fmt.Printf("open tabs: %d\n", len(open))
	for i, page := range open {
		info := page.MustInfo()
		fmt.Printf("  [%d] %s - %s\n", i+1, info.URL, info.Title)
	}
}