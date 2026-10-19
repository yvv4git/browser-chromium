// Command spa_probe scrolls SPA pages to the bottom in a loop, measures DOM
// growth (scrollHeight, HTML size, links, text) and logs XHR/Fetch requests
// triggered by the scrolling. It helps to classify content-loading mechanics
// of a site: infinite scroll, pagination links, lazy-load, or a static stub.
//
//	go run ./spa_probe -addr http://localhost:9222
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/yvv4git/browser-chromium/examples/internal/cdp"
)

type metrics struct {
	ScrollHeight int `json:"scrollHeight"`
	HTMLLen      int `json:"htmlLen"`
	Links        int `json:"links"`
	TextLen      int `json:"textLen"`
}

type stepResult struct {
	Step    int     `json:"step"`
	URL     string  `json:"url"`
	Metrics metrics `json:"metrics"`
}

type siteResult struct {
	URL          string       `json:"url"`
	InitialURL   string       `json:"initialURL"`
	FinalURL     string       `json:"finalURL"`
	Error        string       `json:"error,omitempty"`
	Steps        []stepResult `json:"steps"`
	Requests     []string     `json:"requests"`
	GrewByScroll bool         `json:"grewByScroll"`
	URLChanged   bool         `json:"urlChanged"`
}

type site struct {
	name string
	url  string
	// afterLoad is an optional JS snippet executed after page load (e.g. a
	// search button click for demos that do not fetch on scroll until then).
	afterLoad string
}

var sites = []site{
	{name: "tumblr", url: "https://www.tumblr.com/"},
	{name: "pexels", url: "https://www.pexels.com/"},
	{name: "imgur", url: "https://imgur.com/"},
	{name: "wikitok", url: "https://wikitok.vercel.app/"},
	{name: "vc.ru", url: "https://vc.ru/"},
	{name: "infinite-scroll-full-page", url: "https://infinite-scroll.com/demo/full-page/"},
	{name: "infinite-scroll-unsplash", url: "https://alobuuls.github.io/infinite-scroll-unsplash/",
		afterLoad: `() => { const b = document.getElementById('btn-search'); if (b) b.click(); }`},
	{name: "devmuzaky-infinite-scroll", url: "https://devmuzaky.github.io/Infinite-Scroll/"},
	{name: "infinite-scroll-masonry", url: "https://infinite-scroll.com/demo/masonry/"},
	{name: "scrollmagic-infinite", url: "https://scrollmagic.io/examples/advanced/infinite_scrolling.html"},
	{name: "htmx-infinite-scroll", url: "https://htmx.org/examples/infinite-scroll/"},
}

const (
	maxSteps    = 12
	stepDelay   = 1500 * time.Millisecond
	initialWait = 4 * time.Second
	pageTimeout = 60 * time.Second
	maxReqLog   = 40
)

func main() {
	addr := flag.String("addr", "http://localhost:9222", "CDP endpoint of the Chromium container")
	outPath := flag.String("out", "", "write results to this JSON file")
	only := flag.String("only", "", "comma-separated site name substrings to probe (default: all)")
	timeout := flag.Duration("timeout", pageTimeout, "per-page timeout")
	noHijack := flag.Bool("no-hijack", false, "do not intercept Fetch/XHR requests")
	flag.Parse()

	browser := cdp.Connect(*addr)
	defer browser.MustClose()

	selected := sites
	if *only != "" {
		selected = nil
		for _, s := range sites {
			if strings.Contains(s.name, *only) || strings.Contains(s.url, *only) {
				selected = append(selected, s)
			}
		}
	}

	results := make([]siteResult, 0, len(selected))
	for _, s := range selected {
		fmt.Printf("probing %s ...\n", s.url)
		res := probeSite(browser, s, *timeout, !*noHijack)
		results = append(results, res)
		printResult(res)
	}

	if *outPath != "" {
		out, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(*outPath, out, 0o600); err != nil {
			panic(err)
		}
	}
}

func probeSite(browser *rod.Browser, s site, timeout time.Duration, hijack bool) (res siteResult) {
	res.URL = s.url
	defer func() {
		if r := recover(); r != nil {
			res.Error = fmt.Sprint(r)
		}
	}()

	page := browser.MustPage("about:blank")
	defer page.MustClose()
	page = page.Timeout(timeout)

	requests := make([]string, 0, maxReqLog)
	var mu sync.Mutex
	seenReq := make(map[string]bool)
	if hijack {
		router := page.HijackRequests()
		logReq := func(u string) {
			mu.Lock()
			defer mu.Unlock()
			if !seenReq[u] && len(requests) < maxReqLog {
				seenReq[u] = true
				requests = append(requests, u)
			}
		}
		_ = router.Add("*", proto.NetworkResourceTypeFetch, func(ctx *rod.Hijack) {
			logReq(ctx.Request.URL().String())
			ctx.ContinueRequest(&proto.FetchContinueRequest{})
		})
		_ = router.Add("*", proto.NetworkResourceTypeXHR, func(ctx *rod.Hijack) {
			logReq(ctx.Request.URL().String())
			ctx.ContinueRequest(&proto.FetchContinueRequest{})
		})
		go router.Run()
		defer router.Stop()
	}

	page.MustNavigate(s.url)
	// Some SPAs never fire the full load event; wait best-effort and fall back
	// to a fixed delay.
	if err := page.WaitLoad(); err != nil {
		fmt.Printf("  WaitLoad: %v, falling back to sleep\n", err)
	}
	time.Sleep(initialWait)

	if s.afterLoad != "" {
		if _, err := page.Eval(s.afterLoad); err != nil {
			res.Error = "afterLoad: " + err.Error()
		}
		time.Sleep(2 * time.Second)
	}

	res.InitialURL = page.MustInfo().URL
	prev := metrics{}
	var steps []stepResult
	stable := 0

	for i := 0; i < maxSteps; i++ {
		m := pageMetrics(page)
		steps = append(steps, stepResult{Step: i, URL: page.MustInfo().URL, Metrics: m})
		if i > 0 && m == prev {
			stable++
			if stable >= 2 {
				break
			}
		} else {
			stable = 0
		}
		prev = m
		scrollBottom(page)
		time.Sleep(stepDelay)
	}
	// Final sample after the last scroll.
	m := pageMetrics(page)
	steps = append(steps, stepResult{Step: len(steps), URL: page.MustInfo().URL, Metrics: m})

	res.Steps = steps
	mu.Lock()
	res.Requests = append([]string(nil), requests...)
	mu.Unlock()
	res.FinalURL = page.MustInfo().URL
	res.URLChanged = res.InitialURL != res.FinalURL
	if len(steps) > 1 {
		res.GrewByScroll = m.HTMLLen > steps[0].Metrics.HTMLLen*3/2 || m.Links > steps[0].Metrics.Links*3/2
	}
	return res
}

func pageMetrics(page *rod.Page) metrics {
	var m metrics
	v, err := page.Eval(`() => {
		const el = document.scrollingElement || document.documentElement;
		return JSON.stringify({
			scrollHeight: el.scrollHeight,
			htmlLen: document.documentElement.outerHTML.length,
			links: document.querySelectorAll('a').length,
			textLen: (document.body && document.body.innerText || '').length,
		});
	}`)
	if err != nil {
		return m
	}
	_ = json.Unmarshal([]byte(v.Value.String()), &m)
	return m
}

func scrollBottom(page *rod.Page) {
	_, _ = page.Eval(`() => {
		const el = document.scrollingElement || document.documentElement;
		window.scrollTo(0, el.scrollHeight);
	}`)
}

func printResult(r siteResult) {
	fmt.Printf("\n=== %s ===\n", r.URL)
	if r.Error != "" {
		fmt.Printf("error: %s\n", r.Error)
	}
	for _, st := range r.Steps {
		fmt.Printf("step %2d: scrollH=%-7d htmlLen=%-9d links=%-5d textLen=%-7d %s\n",
			st.Step, st.Metrics.ScrollHeight, st.Metrics.HTMLLen,
			st.Metrics.Links, st.Metrics.TextLen, shorten(st.URL, 100))
	}
	fmt.Printf("grewByScroll=%v urlChanged=%v\n", r.GrewByScroll, r.URLChanged)
	if len(r.Requests) > 0 {
		fmt.Printf("requests (%d):\n", len(r.Requests))
		for _, u := range r.Requests {
			fmt.Printf("  %s\n", shorten(u, 120))
		}
	}
}

func shorten(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}
