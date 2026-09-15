// Command status_code loads a list of URLs in the Chromium container and
// prints the HTTP status code of each response using request interception.
//
// The container must be running first (make compose-up).
//
//	go run ./status_code -addr http://localhost:9222
package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"github.com/yvv4git/browser-chromium/examples/internal/cdp"
)

func main() {
	addr := flag.String("addr", "http://localhost:9222", "CDP endpoint of the Chromium container")
	flag.Parse()

	urls := []string{
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/404",
		"https://httpbin.org/status/403",
		"https://httpbin.org/status/500",
	}

	browser := cdp.Connect(*addr)

	for _, link := range urls {
		fmt.Printf("\n=== %s ===\n", link)
		checkURL(browser, link)
	}
}

func checkURL(browser *rod.Browser, target string) {
	page := browser.MustPage("")
	defer page.MustClose()

	router := page.HijackRequests()
	defer router.Stop()

	var statusCode int
	var gotResponse bool

	router.Add("*", proto.NetworkResourceTypeDocument, func(ctx *rod.Hijack) {
		ctx.ContinueRequest(&proto.FetchContinueRequest{})
		ctx.LoadResponse(&http.Client{}, true)
		if ctx.Response != nil {
			statusCode = ctx.Response.Payload().ResponseCode
			gotResponse = true
			fmt.Printf("status code: %d\n", statusCode)
		}
	})

	go router.Run()

	page.Navigate(target)

	if !gotResponse {
		fmt.Println("no response")
	}
}