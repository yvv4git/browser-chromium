// Command search runs a Wikipedia search in the Chromium container: fills the
// search input, clicks search, reads the article title and saves the first
// article image.
//
// The container must be running first (make compose-up).
//
//	go run ./search -addr http://localhost:9222 -q Earth
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/utils"

	"github.com/yvv4git/browser-chromium/examples/internal/cdp"
)

func main() {
	addr := flag.String("addr", "http://localhost:9222", "CDP endpoint of the Chromium container")
	query := flag.String("q", "Earth", "search query")
	output := flag.String("output", "article.png", "article image output file")
	flag.Parse()

	browser := cdp.Connect(*addr)

	page := browser.MustPage("https://www.wikipedia.org").MustWaitLoad()
	defer page.MustClose()

	searchInput := page.MustWaitStable().MustElement("#searchInput")
	searchInput.MustInput(*query)
	page.MustElement("#search-form > fieldset > button").MustClick()
	page.MustWaitLoad().MustWaitStable()

	fmt.Printf("article: %s\n", page.MustElement("#firstHeading").MustText())

	err := rod.Try(func() {
		img := page.MustWaitStable().MustElement("#mw-content-text img")
		data := img.MustResource()
		if err := utils.OutputFile(*output, data); err != nil {
			log.Printf("save: %v", err)
		} else {
			log.Printf("article image saved: %s", *output)
		}
	})
	if err != nil {
		log.Printf("image: %v", err)
	}
}