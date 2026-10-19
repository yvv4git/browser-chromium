# spa_probe

Scrolls SPA pages to the bottom in a loop and records what happens:

- DOM growth per step: `scrollHeight`, `outerHTML` length, `<a>` count, `innerText` length
- XHR/Fetch requests triggered by scrolling
- URL changes caused by scrolling (`history.pushState` based demos)

Use it to classify content-loading mechanics of a site: infinite scroll,
pagination links, lazy-load, JS click-driven loading, or a static stub.

Sites to probe are hard-coded in `sites` inside `main.go`.

## Usage

The container must be running first:

```bash
cd .. && make compose-up
```

Run the example:

```bash
go run ./spa_probe -addr http://localhost:9222 -out /tmp/spa_probe.json
```

## Flags

| Flag    | Default                 | Description                            |
| ------- | ----------------------- | -------------------------------------- |
| `-addr` | `http://localhost:9222` | CDP endpoint of the Chromium container |
| `-out`  |                         | Write full results to a JSON file      |

Run `go run ./spa_probe -h` for the full list.
