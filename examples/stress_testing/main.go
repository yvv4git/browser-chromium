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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"github.com/yvv4git/browser-chromium/examples/internal/cdp"
)

const (
	defaultConfig     = "config.toml"
	metricsTimeLayout = "2006-01-02_15_04"
	statusOK          = 200
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

type config struct {
	Addr       string        `toml:"addr"`
	Workers    int           `toml:"workers"`
	Proxy      string        `toml:"proxy"`
	List       string        `toml:"list"`
	BlockMedia bool          `toml:"block_media"`
	Timeout    time.Duration `toml:"timeout"`
	Loop       bool          `toml:"loop"`
	SaveHTML   bool          `toml:"save_html"`
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

type runMetrics struct {
	RuntimeS        float64          `json:"runtime_s"`
	PagesProcessed  int              `json:"pages_processed"`
	MinPageTimeS    float64          `json:"min_page_time_s"`
	MaxPageTimeS    float64          `json:"max_page_time_s"`
	AvgPageTimeS    float64          `json:"avg_page_time_s"`
	MedianPageTimeS float64          `json:"median_page_time_s"`
	StatusCounts    map[int]int      `json:"status_counts"`
	Non200URLs      map[int][]string `json:"non_200_urls,omitempty"`
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

	configPath := flag.String("config", filepath.Join(sourceDir(), defaultConfig), "path to config.toml")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if cfg.Workers < 1 {
		fmt.Fprintln(os.Stderr, "workers must be >= 1")
		os.Exit(1)
	}

	targets, err := readList(cfg.List)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(targets) == 0 {
		fmt.Fprintln(os.Stderr, "list is empty")
		os.Exit(1)
	}

	browser := cdp.Connect(cfg.Addr)

	var browserContextID proto.BrowserBrowserContextID
	if cfg.Proxy != "" {
		id, createErr := proto.TargetCreateBrowserContext{ProxyServer: cfg.Proxy}.Call(browser)
		if createErr != nil {
			fmt.Fprintf(os.Stderr, "create browser context: %v\n", createErr)
			os.Exit(1)
		}
		browserContextID = id.BrowserContextID
		incognito := *browser
		incognito.BrowserContextID = browserContextID
		browser = &incognito
		defer func() {
			_ = proto.TargetDisposeBrowserContext{BrowserContextID: browserContextID}.Call(browser)
		}()
		fmt.Printf("proxy: %s\n", cfg.Proxy)
	} else {
		fmt.Println("proxy: direct")
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
		<-sigCh
		fmt.Fprintln(os.Stderr, "forced exit")
		os.Exit(1)
	}()
	defer signal.Stop(sigCh)
	defer cancel()

	metricsDir := filepath.Join(sourceDir(), "result")
	if err := os.MkdirAll(metricsDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "create result dir: %v\n", err)
		os.Exit(1)
	}

	resultDir := ""
	if cfg.SaveHTML {
		resultDir = metricsDir
		fmt.Printf("save html: %s\n", resultDir)
	}

	st := &stats{}
	jobs := make(chan string)
	var wg sync.WaitGroup
	for i := range cfg.Workers {
		wg.Add(1)
		go worker(ctx, &wg, i+1, browser, browserContextID, jobs, st, cfg.BlockMedia, cfg.Timeout, resultDir)
	}

	fmt.Printf("workers: %d, targets: %d\n", cfg.Workers, len(targets))
loop:
	for {
		for _, target := range targets {
			select {
			case <-ctx.Done():
				break loop
			case jobs <- target:
			}
		}
		if !cfg.Loop {
			break
		}
	}
	close(jobs)
	wg.Wait()
	printStats(start, st, metricsDir)
}

func worker(ctx context.Context, wg *sync.WaitGroup, id int, browser *rod.Browser, browserContextID proto.BrowserBrowserContextID, jobs <-chan string, st *stats, blockMedia bool, pageTimeout time.Duration, resultDir string) {
	defer wg.Done()
	for target := range jobs {
		if ctx.Err() != nil {
			return
		}
		openAndClose(ctx, id, browser, browserContextID, target, st, blockMedia, pageTimeout, resultDir)
	}
}

func openAndClose(ctx context.Context, workerID int, browser *rod.Browser, browserContextID proto.BrowserBrowserContextID, target string, st *stats, blockMedia bool, pageTimeout time.Duration, resultDir string) {
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
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()
		p := page.Context(closeCtx)
		if closeErr := p.Close(); closeErr != nil {
			fmt.Printf("worker %d: close %s: %v\n", workerID, target, closeErr)
		}
	}()

	navCtx, navCancel := context.WithTimeout(ctx, pageTimeout)
	defer navCancel()
	page = page.Context(navCtx)
	go func() {
		<-navCtx.Done()
		if ctx.Err() != nil || errors.Is(navCtx.Err(), context.DeadlineExceeded) {
			stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
			_ = proto.PageStopLoading{}.Call(page.Context(stopCtx))
			stopCancel()
		}
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

	navigateErr := page.Navigate(target)
	loadErr := error(nil)
	if navigateErr == nil {
		loadErr = page.WaitLoad()
	}

	canceled := ctx.Err() != nil
	timedOut := !canceled && errors.Is(navCtx.Err(), context.DeadlineExceeded)

	switch {
	case canceled:
		pageErr = "canceled"
	case timedOut:
		pageErr = "timeout"
		fmt.Printf("worker %d: timeout %s\n", workerID, target)
	case navigateErr != nil:
		pageErr = navigateErr.Error()
		fmt.Printf("worker %d: navigate %s: %v\n", workerID, target, navigateErr)
	case loadErr != nil:
		pageErr = loadErr.Error()
		fmt.Printf("worker %d: load %s: %v\n", workerID, target, loadErr)
	default:
		stableCtx, stableCancel := context.WithTimeout(navCtx, time.Second)
		_ = page.Context(stableCtx).WaitDOMStable(time.Second, 0)
		stableCancel()
	}

	if resultDir != "" || bytes.Load() == 0 {
		saveCtx, saveCancel := context.WithTimeout(context.Background(), 30*time.Second)
		html, htmlErr := pageHTML(page.Context(saveCtx))
		if htmlErr != nil {
			if resultDir != "" {
				fmt.Printf("worker %d: save html %s: %v\n", workerID, target, htmlErr)
			}
		} else {
			if bytes.Load() == 0 {
				bytes.Store(int64(len(html)))
			}
			if resultDir != "" {
				if saveErr := writePageHTML(target, resultDir, html); saveErr != nil {
					fmt.Printf("worker %d: save html %s: %v\n", workerID, target, saveErr)
				}
			}
		}
		saveCancel()
	}

	if pageErr != "" && !timedOut {
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

func printStats(start time.Time, st *stats, metricsDir string) {
	results := st.snapshot()
	total := time.Since(start)

	metrics := runMetrics{
		RuntimeS:       seconds(total),
		PagesProcessed: len(results),
	}

	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "METRIC\tVALUE")
	fmt.Fprintf(w, "Runtime\t%s\n", total.Round(time.Millisecond))
	fmt.Fprintf(w, "Pages processed\t%d\n", len(results))
	if len(results) > 0 {
		durations := make([]time.Duration, len(results))
		statusCounts := map[int]int{}
		non200URLs := map[int][]string{}
		for i, r := range results {
			durations[i] = r.duration
			statusCounts[r.status]++
			if r.status != statusOK {
				non200URLs[r.status] = append(non200URLs[r.status], r.url)
			}
		}
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		for code := range non200URLs {
			sort.Strings(non200URLs[code])
		}

		var sum time.Duration
		for _, d := range durations {
			sum += d
		}
		avg := sum / time.Duration(len(durations))
		median := medianDuration(durations)

		metrics.MinPageTimeS = seconds(durations[0])
		metrics.MaxPageTimeS = seconds(durations[len(durations)-1])
		metrics.AvgPageTimeS = seconds(avg)
		metrics.MedianPageTimeS = seconds(median)
		metrics.StatusCounts = statusCounts
		metrics.Non200URLs = non200URLs

		fmt.Fprintf(w, "Min page time\t%s\n", durations[0].Round(time.Millisecond))
		fmt.Fprintf(w, "Max page time\t%s\n", durations[len(durations)-1].Round(time.Millisecond))
		fmt.Fprintf(w, "Avg page time\t%s\n", avg.Round(time.Millisecond))
		fmt.Fprintf(w, "Median page time\t%s\n", median.Round(time.Millisecond))

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

	if err := saveMetrics(metricsDir, metrics); err != nil {
		fmt.Fprintf(os.Stderr, "save metrics: %v\n", err)
	}
}

func saveMetrics(dir string, metrics runMetrics) error {
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metrics: %w", err)
	}

	path := filepath.Join(dir, time.Now().Format(metricsTimeLayout)+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write metrics %s: %w", path, err)
	}

	return nil
}

func seconds(d time.Duration) float64 {
	return float64(d.Round(time.Millisecond)) / float64(time.Second)
}

func medianDuration(ds []time.Duration) time.Duration {
	n := len(ds)
	if n%2 == 1 {
		return ds[n/2]
	}
	return (ds[n/2-1] + ds[n/2]) / 2
}

func loadConfig(path string) (config, error) {
	cfg := config{
		Addr:    "http://localhost:9222",
		Workers: 5,
		List:    "examples/stress_testing/list.txt",
		Timeout: 10 * time.Second,
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return config{}, fmt.Errorf("load config %s: %w", path, err)
	}
	return cfg, nil
}

func sourceDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	return filepath.Dir(file)
}

func writePageHTML(target, resultDir, html string) error {
	u, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	host := u.Hostname()
	if host == "" {
		host = "page"
	}
	host = strings.ReplaceAll(host, ":", "_")

	path := filepath.Join(resultDir, host+".html")
	return os.WriteFile(path, []byte(html), 0o644)
}

func pageHTML(page *rod.Page) (string, error) {
	doc, err := proto.DOMGetDocument{}.Call(page)
	if err == nil && doc.Root != nil {
		res, err := proto.DOMGetOuterHTML{BackendNodeID: doc.Root.BackendNodeID}.Call(page)
		if err == nil {
			return res.OuterHTML, nil
		}
	}

	res, err := proto.RuntimeEvaluate{
		Expression:    "document.documentElement ? document.documentElement.outerHTML : ''",
		ReturnByValue: true,
	}.Call(page)
	if err != nil {
		return "", err
	}
	if res.ExceptionDetails != nil {
		return "", fmt.Errorf("evaluate html: %s", res.ExceptionDetails.Text)
	}
	if res.Result.Value.Val() == nil {
		return "", nil
	}
	return fmt.Sprint(res.Result.Value.Val()), nil
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
