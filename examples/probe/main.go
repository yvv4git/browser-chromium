// Command probe prints browser fingerprint signals that anti-bot systems check.
//
//	go run ./probe -addr http://localhost:9222
package main

import (
	"flag"
	"fmt"

	"github.com/yvv4git/browser-chromium/examples/internal/cdp"
)

var signals = []string{
	`() => "webdriver: " + navigator.webdriver`,
	`() => "userAgent: " + navigator.userAgent`,
	`() => "plugins: " + navigator.plugins.length`,
	`() => "languages: " + navigator.languages.join(",")`,
	`() => "hardwareConcurrency: " + navigator.hardwareConcurrency`,
	`() => "deviceMemory: " + navigator.deviceMemory`,
	`() => "chrome: " + typeof window.chrome`,
	`() => "webdriverProp: " + ("webdriver" in navigator)`,
	`() => "webglRenderer: " + (() => { const gl = document.createElement("canvas").getContext("webgl"); if (!gl) return "no-webgl"; const d = gl.getExtension("WEBGL_debug_renderer_info"); return d ? gl.getParameter(d.UNMASKED_RENDERER_WEBGL) : "no-debug-info"; })()`,
	`() => "jsHeap: " + performance.memory.usedJSHeapSize`,
}

func main() {
	addr := flag.String("addr", "http://localhost:9222", "CDP endpoint of the Chromium container")
	flag.Parse()

	browser := cdp.Connect(*addr)
	page := browser.MustPage("https://example.com").MustWaitLoad()
	defer page.MustClose()

	for _, signal := range signals {
		result, err := page.Eval(signal)
		if err != nil {
			fmt.Printf("ERROR %v\n", err)

			continue
		}
		fmt.Println(result.Value.Val())
	}
}
