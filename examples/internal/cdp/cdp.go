// Package cdp provides a helper to connect to a remote Chromium browser over
// the Chrome DevTools Protocol for the examples.
package cdp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-rod/rod"
)

// Connect connects to the Chromium container. The addr can be either the plain
// HTTP endpoint (e.g. http://localhost:9222) or the full WebSocket URL
// (e.g. ws://localhost:9222/devtools/browser/<id>).
func Connect(addr string) *rod.Browser {
	wsURL := addr
	if !strings.HasPrefix(wsURL, "ws://") && !strings.HasPrefix(wsURL, "wss://") {
		var err error
		wsURL, err = resolveWSURL(addr)
		if err != nil {
			panic(err)
		}
	}
	return rod.New().ControlURL(wsURL).MustConnect()
}

func resolveWSURL(addr string) (string, error) {
	endpoint := strings.TrimSuffix(addr, "/") + "/json/version"

	resp, err := http.Get(endpoint)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch %s: %s", endpoint, resp.Status)
	}

	var data struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("parse %s: %w", endpoint, err)
	}
	if data.WebSocketDebuggerURL == "" {
		return "", errors.New("no webSocketDebuggerUrl in " + endpoint)
	}
	return data.WebSocketDebuggerURL, nil
}