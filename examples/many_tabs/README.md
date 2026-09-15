# many_tabs

Opens several tabs in the Chromium container and demonstrates multi-tab
work:

- opening pages in several tabs
- switching between tabs with `MustActivate`
- listing all currently open tabs with URLs and titles

## Usage

The container must be running first:

```bash
cd .. && make compose-up
```

Run the example:

```bash
go run ./many_tabs -addr http://localhost:9222
```

## Flags

| Flag    | Default                 | Description                            |
| ------- | ----------------------- | -------------------------------------- |
| `-addr` | `http://localhost:9222` | CDP endpoint of the Chromium container |

Run `go run ./many_tabs -h` for the full list.
