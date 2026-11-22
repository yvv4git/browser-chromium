# stress_testing

Continuously opens sites from `list.txt` in parallel tabs against a running
Chromium container. Useful for load testing headless or default browser modes.

## Usage

Start a browser container first:

```bash
cd ../.. && make compose-up
```

Run with defaults (5 workers, direct connection, `examples/stress_testing/list.txt`):

```bash
go run ./stress_testing
```

## Flags

| Flag          | Default                          | Description                               |
| ------------- | -------------------------------- | ----------------------------------------- |
| `-addr`       | `http://localhost:9222`          | CDP endpoint of the container             |
| `-workers`    | `5`                              | Number of parallel tabs                   |
| `-proxy`      | empty                            | HTTP proxy URL, e.g. `http://host:8138`   |
| `-list`       | `examples/stress_testing/list.txt` | Path to file with URLs                  |
| `-block-media`| `false`                          | Block media requests (images/video/audio/fonts) |
| `-timeout`    | `10s`                            | Per-page load timeout                      |

## Examples

Run with 10 workers:

```bash
go run ./stress_testing -workers 10
```

Run with 2 workers:

```bash
go run ./stress_testing -workers 2
```

Run through proxy:

```bash
go run ./stress_testing -proxy http://host.docker.internal:8138
```

Stop with `Ctrl+C`. The app performs graceful shutdown, closes active tabs, and prints a statistics table (runtime, processed pages, min/max/avg/median page time, HTTP status code counts; status `0` means no response, e.g. DNS resolution failure).
