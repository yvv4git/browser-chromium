# stress_testing

Continuously opens sites from a URL list in parallel tabs against a running
Chromium container. Useful for load testing headless or default browser modes.

## Usage

Start a browser container first from the repository root:

```bash
make compose-up
```

Run with default config (`examples/stress_testing/config.toml`):

```bash
go run ./examples/stress_testing
```

Run with a custom config:

```bash
go run ./examples/stress_testing -config /path/to/config.toml
```

## Config

Settings are loaded from `config.toml` (default: `examples/stress_testing/config.toml`).

```toml
addr = "http://localhost:9222"
workers = 5
proxy = ""
list = "examples/stress_testing/list.txt"
block_media = false
timeout = "10s"
loop = false
save_html = false
```

| Field         | Default                            | Description                                             |
| ------------- | ---------------------------------- | ------------------------------------------------------- |
| `addr`        | `http://localhost:9222`            | CDP endpoint of the Chromium container                  |
| `workers`     | `5`                                | Number of parallel tabs                                 |
| `proxy`       | empty                              | HTTP proxy URL, e.g. `http://host:8138`                 |
| `list`        | `examples/stress_testing/list.txt` | Path to file with URLs                                  |
| `block_media` | `false`                            | Block media requests (images, video, audio, fonts)      |
| `timeout`     | `10s`                              | Per-page load timeout                                   |
| `loop`        | `false`                            | Iterate list repeatedly until interrupted               |
| `save_html`   | `false`                            | Save page HTML to `examples/stress_testing/result/[hostname].html` |

## Flags

| Flag       | Default                                      | Description         |
| ---------- | -------------------------------------------- | ------------------- |
| `-config`  | `examples/stress_testing/config.toml`        | Path to config.toml |

## URL list format

`-list` points to a text file with one URL per line. Empty lines and lines starting with `#` are ignored.

```text
# comment
https://example.com
https://chromium.org
```

## Examples

Run with default config:

```bash
go run ./examples/stress_testing
```

Use a custom config file:

```bash
go run ./examples/stress_testing -config examples/stress_testing/config.toml
```

Example `config.toml` for proxy + HTML saving:

```toml
addr = "http://localhost:9222"
workers = 5
proxy = "http://host.docker.internal:8138"
list = "examples/stress_testing/good_hostnames.txt"
block_media = true
timeout = "10s"
loop = false
save_html = true
```

Example `config.toml` for a long-running loop:

```toml
addr = "http://localhost:9222"
workers = 8
proxy = ""
list = "examples/stress_testing/list.txt"
block_media = true
timeout = "15s"
loop = true
save_html = false
```

## Output

Stop with `Ctrl+C`. The app performs graceful shutdown, closes active tabs, and prints:

- summary statistics: runtime, pages processed, min/max/avg/median page time, HTTP status code counts;
- per-page results table sorted by URL, status, and size.

Status `0` means no HTTP response was captured, e.g. DNS resolution failure or navigation error.
