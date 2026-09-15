// Command connect connects to the Chromium container over CDP and lists the
// browser version and the currently open tabs.
//
// The container must be running first (make compose-up).
//
//	go run ./connect -addr http://localhost:9222
package main

import (
	"flag"
	"fmt"

	"github.com/yvv4git/browser-chromium/examples/internal/cdp"
)

func main() {
	addr := flag.String("addr", "http://localhost:9222", "CDP endpoint of the Chromium container")
	flag.Parse()

	browser := cdp.Connect(*addr)

	version, err := browser.Version()
	if err != nil {
		panic(err)
	}
	fmt.Printf("browser: %s\n", version)

	pages := browser.MustPages()
	fmt.Printf("tabs: %d\n", len(pages))
	for i, p := range pages {
		info := p.MustInfo()
		fmt.Printf("  [%d] %s - %s\n", i+1, info.URL, info.Title)
	}
}