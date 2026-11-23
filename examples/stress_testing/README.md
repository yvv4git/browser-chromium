# stress_testing

Continuously opens sites from a URL list in parallel tabs against a running
Chromium container. Useful for load testing headless or default browser modes.

## Usage

Start a browser container first from the repository root:

```bash
make compose-up
```

Run with defaults (5 workers, direct connection, `examples/stress_testing/list.txt`):

```bash
go run ./examples/stress_testing
```

## Flags

| Flag           | Default                            | Description                                             |
| -------------- | ---------------------------------- | ------------------------------------------------------- |
| `-addr`        | `http://localhost:9222`            | CDP endpoint of the Chromium container                  |
| `-workers`     | `5`                                | Number of parallel tabs                                 |
| `-proxy`       | empty                              | HTTP proxy URL, e.g. `http://host:8138`                 |
| `-list`        | `examples/stress_testing/list.txt` | Path to file with URLs                                  |
| `-block-media` | `false`                            | Block media requests (images, video, audio, fonts)      |
| `-timeout`     | `10s`                              | Per-page load timeout                                   |
| `-loop`        | `false`                            | Iterate list repeatedly until interrupted               |

## URL list format

`-list` points to a text file with one URL per line. Empty lines and lines starting with `#` are ignored.

```text
# comment
https://example.com
https://chromium.org
```

## Examples

Run with default settings:

```bash
go run ./examples/stress_testing
```

Run with a specific number of workers:

```bash
go run ./examples/stress_testing -workers 2
```

```bash
go run ./examples/stress_testing -workers 10
```

Run against a custom CDP endpoint:

```bash
go run ./examples/stress_testing -addr http://localhost:9222
```

Run through an HTTP proxy:

```bash
go run ./examples/stress_testing -proxy http://localhost:8138
```

```bash
go run ./examples/stress_testing -proxy http://host.docker.internal:8138
```

Use a custom URL list:

```bash
go run ./examples/stress_testing -list examples/stress_testing/good_hostnames.txt
```

```bash
go run ./examples/stress_testing -list /absolute/path/to/urls.txt
```

Block media to reduce bandwidth and speed up page loads:

```bash
go run ./examples/stress_testing -block-media
```

Set a custom per-page timeout:

```bash
go run ./examples/stress_testing -timeout 5s
```

```bash
go run ./examples/stress_testing -timeout 30s
```

Repeat the URL list until interrupted:

```bash
go run ./examples/stress_testing -loop
```

Combine flags:

```bash
go run ./examples/stress_testing -workers 10 -timeout 15s -block-media
```

```bash
go run ./examples/stress_testing -addr http://localhost:9222 -workers 5 -proxy http://localhost:8138 -list examples/stress_testing/list.txt -timeout 10s
```

```bash
go run ./examples/stress_testing -workers 8 -proxy http://host.docker.internal:8138 -block-media -loop
```

## Output

Stop with `Ctrl+C`. The app performs graceful shutdown, closes active tabs, and prints:

- summary statistics: runtime, pages processed, min/max/avg/median page time, HTTP status code counts;
- per-page results table sorted by URL, status, and size.

Status `0` means no HTTP response was captured, e.g. DNS resolution failure or navigation error.
