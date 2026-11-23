// Command proxy opens a page through an HTTP proxy in the Chromium container.
//
// The container must be running first (make compose-up).
//
// Run from the examples directory:
//
//	cd examples && go run ./proxy
//	cd examples && go run ./proxy -proxy http://host.docker.internal:8138 -url https://example.com
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-rod/rod/lib/proto"

	"github.com/yvv4git/browser-chromium/examples/internal/cdp"
)

func main() {
	addr := flag.String("addr", "http://localhost:9222", "CDP endpoint of the Chromium container")
	proxyURL := flag.String("proxy", "http://host.docker.internal:8138", "HTTP proxy URL reachable from the container")
	pageURL := flag.String("url", "http://ipecho.net/plain", "page to load through the proxy")
	flag.Parse()

	browser := cdp.Connect(*addr)

	id, err := proto.TargetCreateBrowserContext{ProxyServer: *proxyURL}.Call(browser)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create browser context: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = proto.TargetDisposeBrowserContext{BrowserContextID: id.BrowserContextID}.Call(browser)
	}()

	// rod.Browser.Page overwrites BrowserContextID with the value from Browser,
	// so bind the proxy context to the browser itself.
	incognito := *browser
	incognito.BrowserContextID = id.BrowserContextID
	browser = &incognito

	page, err := browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "open page: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = page.Close()
	}()

	if err := page.Navigate(*pageURL); err != nil {
		fmt.Fprintf(os.Stderr, "navigate %s: %v\n", *pageURL, err)
		os.Exit(1)
	}
	if err := page.WaitLoad(); err != nil {
		fmt.Fprintf(os.Stderr, "wait load %s: %v\n", *pageURL, err)
		os.Exit(1)
	}

	info, err := page.Info()
	if err != nil {
		fmt.Fprintf(os.Stderr, "page info: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("proxy: %s\n", *proxyURL)
	fmt.Printf("url: %s\n", info.URL)
	fmt.Printf("title: %s\n", info.Title)

	body, err := page.Eval(`() => document.body ? document.body.innerText.trim() : ""`)
	if err == nil {
		fmt.Printf("body: %v\n", body.Value.Val())
	}
}
