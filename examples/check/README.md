# check

Connects to the Chromium container over CDP, opens a page and demonstrates the
basic interactions:

- prints the page title and URL
- prints the size of the page HTML
- saves a screenshot to a PNG file

## Usage

The container must be running first:

```bash
cd .. && make compose-up
```

Run the example:

```bash
go run ./check -addr http://localhost:9222 https://www.wikipedia.org
```

With a custom screenshot path:

```bash
go run ./check -addr http://localhost:9222 -output wikipedia.png https://example.com
```

## Flags

| Flag      | Default                 | Description                            |
| --------- | ----------------------- | -------------------------------------- |
| `-addr`   | `http://localhost:9222` | CDP endpoint of the Chromium container |
| `-output` | `check.png`             | Screenshot output file                 |

Run `go run ./check -h` for the full list.
