// Command stress_testing continuously opens sites from a list in parallel tabs
// against a running Chromium container.
//
// The container must be running first (make compose-up).
//
//	go run ./stress_testing -addr http://localhost:9222
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"github.com/yvv4git/browser-chromium/examples/internal/cdp"
)

const (
	defaultAddr    = "http://localhost:9222"
	defaultWorkers = 5
	defaultList    = "examples/stress_testing/list.txt"
	defaultTimeout = 10 * time.Second
)

var blockedMediaURLs = []string{
	"*.png", "*.png?*", "*.jpg", "*.jpg?*", "*.jpeg", "*.jpeg?*",
	"*.gif", "*.gif?*", "*.webp", "*.webp?*", "*.svg", "*.svg?*",
	"*.ico", "*.ico?*", "*.bmp", "*.bmp?*", "*.avif", "*.avif?*",
	"*.mp4", "*.mp4?*", "*.webm", "*.webm?*",
	"*.mp3", "*.mp3?*", "*.m4a", "*.m4a?*",
	"*.ogg", "*.ogg?*", "*.wav", "*.wav?*",
	"*.avi", "*.avi?*", "*.mov", "*.mov?*",
	"*.flv", "*.flv?*", "*.mkv", "*.mkv?*",
	"*.m3u8", "*.m3u8?*", "*.ts", "*.ts?*",
	"*.woff", "*.woff?*", "*.woff2", "*.woff2?*",
	"*.ttf", "*.ttf?*", "*.otf", "*.otf?*",
}

type pageResult struct {
	url      string
	duration time.Duration
	status   int
	bytes    int64
	err      string
}

type stats struct {
	mu      sync.Mutex
	results []pageResult
}

func (s *stats) add(r pageResult) {
	s.mu.Lock()
	s.results = append(s.results, r)
	s.mu.Unlock()
}

func (s *stats) snapshot() []pageResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]pageResult, len(s.results))
	copy(out, s.results)
	return out
}

func main() {
	start := time.Now()

	addr := flag.String("addr", defaultAddr, "CDP endpoint of the Chromium container")
	workers := flag.Int("workers", defaultWorkers, "number of parallel tabs")
	proxy := flag.String("proxy", "", "HTTP proxy URL (empty = direct)")
	listPath := flag.String("list", defaultList, "path to file with URLs")
	blockMedia := flag.Bool("block-media", false, "block media requests (images, video, audio, fonts)")
	pageTimeout := flag.Duration("timeout", defaultTimeout, "per-page load timeout")
	flag.Parse()

	if *workers < 1 {
		fmt.Fprintln(os.Stderr, "workers must be >= 1")
		os.Exit(1)
	}

	targets, err := readList(*listPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(targets) == 0 {
		fmt.Fprintln(os.Stderr, "list is empty")
		os.Exit(1)
	}

	browser := cdp.Connect(*addr)

	var browserContextID proto.BrowserBrowserContextID
	if *proxy != "" {
		id, createErr := proto.TargetCreateBrowserContext{ProxyServer: *proxy}.Call(browser)
		if createErr != nil {
			fmt.Fprintf(os.Stderr, "create browser context: %v\n", createErr)
			os.Exit(1)
		}
		browserContextID = id.BrowserContextID
		defer func() {
			_ = proto.TargetDisposeBrowserContext{BrowserContextID: browserContextID}.Call(browser)
		}()
		fmt.Printf("proxy: %s\n", *proxy)
	} else {
		fmt.Println("proxy: direct")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	st := &stats{}
	jobs := make(chan string)
	var wg sync.WaitGroup
	for i := range *workers {
		wg.Add(1)
		go worker(ctx, &wg, i+1, browser, browserContextID, jobs, st, *blockMedia, *pageTimeout)
	}

	fmt.Printf("workers: %d, targets: %d\n", *workers, len(targets))
loop:
	for {
		for _, target := range targets {
			select {
			case <-ctx.Done():
				break loop
			case jobs <- target:
			}
		}
	}
	close(jobs)
	wg.Wait()
	printStats(start, st)
}

func worker(ctx context.Context, wg *sync.WaitGroup, id int, browser *rod.Browser, browserContextID proto.BrowserBrowserContextID, jobs <-chan string, st *stats, blockMedia bool, pageTimeout time.Duration) {
	defer wg.Done()
	for target := range jobs {
		if ctx.Err() != nil {
			return
		}
		openAndClose(ctx, id, browser, browserContextID, target, st, blockMedia, pageTimeout)
	}
}

func openAndClose(ctx context.Context, workerID int, browser *rod.Browser, browserContextID proto.BrowserBrowserContextID, target string, st *stats, blockMedia bool, pageTimeout time.Duration) {
	started := time.Now()
	var status atomic.Int32
	var bytes atomic.Int64
	var pageErr string
	defer func() {
		st.add(pageResult{url: target, duration: time.Since(started), status: int(status.Load()), bytes: bytes.Load(), err: pageErr})
	}()

	page, err := browser.Page(proto.TargetCreateTarget{
		URL:              "about:blank",
		BrowserContextID: browserContextID,
	})
	if err != nil {
		pageErr = err.Error()
		fmt.Printf("worker %d: open %s: %v\n", workerID, target, err)
		return
	}
	defer func() {
		if closeErr := page.Close(); closeErr != nil {
			fmt.Printf("worker %d: close %s: %v\n", workerID, target, closeErr)
		}
	}()

	navCtx, navCancel := context.WithTimeout(ctx, pageTimeout)
	defer navCancel()
	page = page.Context(navCtx)
	go func() {
		<-navCtx.Done()
		_ = page.StopLoading()
		_ = page.Close()
	}()

	enableErr := proto.NetworkEnable{}.Call(page)
	if enableErr != nil {
		fmt.Printf("worker %d: network enable %s: %v\n", workerID, target, enableErr)
	}
	var docRequestID proto.NetworkRequestID
	waitStatus := page.EachEvent(func(e *proto.NetworkResponseReceived) bool {
		if e.Type == proto.NetworkResourceTypeDocument {
			docRequestID = e.RequestID
			status.Store(int32(e.Response.Status))
		}
		return navCtx.Err() != nil
	}, func(e *proto.NetworkLoadingFinished) bool {
		if docRequestID != "" && e.RequestID == docRequestID {
			bytes.Store(int64(e.EncodedDataLength))
		}
		return navCtx.Err() != nil
	})
	go waitStatus()

	if blockMedia {
		if err := page.SetBlockedURLs(blockedMediaURLs); err != nil {
			fmt.Printf("worker %d: block media %s: %v\n", workerID, target, err)
		}
	}

	if err := page.Navigate(target); err != nil {
		if ctx.Err() != nil {
			pageErr = "canceled"
			return
		}
		if navCtx.Err() != nil {
			pageErr = "timeout"
			fmt.Printf("worker %d: timeout %s\n", workerID, target)
			return
		}
		pageErr = err.Error()
		fmt.Printf("worker %d: navigate %s: %v\n", workerID, target, err)
		return
	}
	if err := page.WaitLoad(); err != nil {
		if ctx.Err() != nil {
			pageErr = "canceled"
			return
		}
		if navCtx.Err() != nil {
			pageErr = "timeout"
			fmt.Printf("worker %d: timeout %s\n", workerID, target)
			return
		}
		pageErr = err.Error()
		fmt.Printf("worker %d: load %s: %v\n", workerID, target, err)
		return
	}

	info, err := page.Info()
	if err != nil {
		pageErr = err.Error()
		fmt.Printf("worker %d: info %s: %v\n", workerID, target, err)
		return
	}
	elapsed := time.Since(started).Round(time.Millisecond)
	fmt.Printf("worker %d: %s - %s [status=%d time=%s size=%s]\n", workerID, info.URL, info.Title, status.Load(), elapsed, formatBytes(bytes.Load()))
}

func printStats(start time.Time, st *stats) {
	results := st.snapshot()
	total := time.Since(start)

	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "METRIC\tVALUE")
	fmt.Fprintf(w, "Runtime\t%s\n", total.Round(time.Millisecond))
	fmt.Fprintf(w, "Pages processed\t%d\n", len(results))
	if len(results) > 0 {
		durations := make([]time.Duration, len(results))
		statusCounts := map[int]int{}
		for i, r := range results {
			durations[i] = r.duration
			statusCounts[r.status]++
		}
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })

		var sum time.Duration
		for _, d := range durations {
			sum += d
		}
		avg := sum / time.Duration(len(durations))

		fmt.Fprintf(w, "Min page time\t%s\n", durations[0].Round(time.Millisecond))
		fmt.Fprintf(w, "Max page time\t%s\n", durations[len(durations)-1].Round(time.Millisecond))
		fmt.Fprintf(w, "Avg page time\t%s\n", avg.Round(time.Millisecond))
		fmt.Fprintf(w, "Median page time\t%s\n", medianDuration(durations).Round(time.Millisecond))

		codes := make([]int, 0, len(statusCounts))
		for code := range statusCounts {
			codes = append(codes, code)
		}
		sort.Ints(codes)
		for _, code := range codes {
			fmt.Fprintf(w, "Status %d\t%d\n", code, statusCounts[code])
		}
	}
	_ = w.Flush()

	tableResults := make([]pageResult, len(results))
	copy(tableResults, results)
	sort.Slice(tableResults, func(i, j int) bool {
		if tableResults[i].url != tableResults[j].url {
			return tableResults[i].url < tableResults[j].url
		}
		if tableResults[i].status != tableResults[j].status {
			return tableResults[i].status < tableResults[j].status
		}
		return tableResults[i].bytes > tableResults[j].bytes
	})

	fmt.Println()
	pw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(pw, "#\tURL\tRESULT")
	for i, r := range tableResults {
		result := fmt.Sprintf("%s status=%d size=%s", r.duration.Round(time.Millisecond), r.status, formatBytes(r.bytes))
		if r.status == 0 && r.err != "" {
			result = fmt.Sprintf("%s status=0 size=%s error=%s", r.duration.Round(time.Millisecond), formatBytes(r.bytes), r.err)
		}
		fmt.Fprintf(pw, "%d\t%s\t%s\n", i+1, r.url, result)
	}
	_ = pw.Flush()
}

func medianDuration(ds []time.Duration) time.Duration {
	n := len(ds)
	if n%2 == 1 {
		return ds[n/2]
	}
	return (ds[n/2-1] + ds[n/2]) / 2
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func readList(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open list: %w", err)
	}
	defer file.Close()

	var targets []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		targets = append(targets, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read list: %w", err)
	}
	return targets, nil
}
