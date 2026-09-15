# links

Opens a page in the Chromium container, collects all links and prints them as
`anchor URL` pairs. The example filters out:

- empty and anchor (`#`) links
- static resources (images, css, js, fonts, pdf, zip)
- relative links are resolved to absolute URLs

## Usage

The container must be running first:

```bash
cd .. && make compose-up
```

Run the example:

```bash
go run ./links -addr http://localhost:9222 -url http://books.toscrape.com
```

## Flags

| Flag     | Default                     | Description                            |
| -------- | --------------------------- | -------------------------------------- |
| `-addr`  | `http://localhost:9222`     | CDP endpoint of the Chromium container |
| `-url`   | `http://books.toscrape.com` | Page to scan                           |
| `-limit` | `20`                        | Max links to print                     |

Run `go run ./links -h` for the full list.
