// Command screenshot opens a page in the Chromium container and saves both a
// viewport screenshot and a full-page screenshot.
//
// The container must be running first (make compose-up).
//
//	go run ./screenshot -addr http://localhost:9222 -url https://www.wikipedia.org
//	go run ./screenshot -addr http://localhost:9222 -full
package main

import (
	"flag"
	"log"

	"github.com/go-rod/rod/lib/utils"

	"github.com/yvv4git/browser-chromium/examples/internal/cdp"
)

func main() {
	addr := flag.String("addr", "http://localhost:9222", "CDP endpoint of the Chromium container")
	pageURL := flag.String("url", "https://www.wikipedia.org", "page to load")
	output := flag.String("output", "screenshot.png", "output file")
	full := flag.Bool("full", false, "capture full page instead of viewport")
	flag.Parse()

	browser := cdp.Connect(*addr)

	page := browser.MustPage(*pageURL).MustWaitLoad()
	defer page.MustClose()

	page.MustWaitStable()

	var (
		data []byte
		err  error
	)
	if *full {
		data, err = page.Screenshot(true, nil)
	} else {
		data, err = page.Screenshot(false, nil)
	}
	if err != nil {
		log.Fatalf("screenshot: %v", err)
	}
	if err := utils.OutputFile(*output, data); err != nil {
		log.Fatalf("write %s: %v", *output, err)
	}
	log.Printf("saved %s (%d bytes)", *output, len(data))
}