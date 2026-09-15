// Command check connects to the Chromium container over CDP, loads a page,
// prints its title, URL and HTML size, then saves a screenshot.
//
// The container must be running first (make compose-up).
//
//	go run ./check -addr http://localhost:9222 https://www.wikipedia.org
//	go run ./check -addr http://localhost:9222 -output wikipedia.png https://example.com
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/yvv4git/browser-chromium/examples/internal/cdp"
)

func main() {
	addr := flag.String("addr", "http://localhost:9222", "CDP endpoint of the Chromium container")
	output := flag.String("output", "check.png", "screenshot output file")
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	target := flag.Arg(0)

	browser := cdp.Connect(*addr)

	page := browser.MustPage(target).MustWaitLoad()
	defer page.MustClose()

	page.MustWaitStable()

	info := page.MustInfo()
	fmt.Printf("title: %s\n", info.Title)
	fmt.Printf("url:   %s\n", info.URL)

	html, err := page.HTML()
	if err != nil {
		log.Fatalf("html: %v", err)
	}
	fmt.Printf("html:  %d bytes\n", len(html))

	page.MustScreenshot(*output)
	fmt.Printf("screenshot saved: %s\n", *output)
}