# search

Runs a Wikipedia search in the Chromium container and demonstrates:

- finding the search input by CSS selector and typing a query
- clicking the search button
- waiting for the article page to load
- reading the article title
- downloading the first article image to a file

## Usage

The container must be running first:

```bash
cd .. && make compose-up
```

Run the example:

```bash
go run ./search -addr http://localhost:9222 -q Earth
```

With a custom image path:

```bash
go run ./search -addr http://localhost:9222 -q Earth -output earth.png
```

## Flags

| Flag      | Default                 | Description                            |
| --------- | ----------------------- | -------------------------------------- |
| `-addr`   | `http://localhost:9222` | CDP endpoint of the Chromium container |
| `-q`      | `Earth`                 | Search query                           |
| `-output` | `article.png`           | Article image output file              |

Run `go run ./search -h` for the full list.
