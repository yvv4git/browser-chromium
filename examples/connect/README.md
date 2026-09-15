# connect

Connects to the Chromium container over CDP and prints:

- the browser version
- the list of currently open tabs with their URLs and titles

## Usage

The container must be running first:

```bash
cd .. && make compose-up
```

Run the example:

```bash
go run ./connect -addr http://localhost:9222
```

## Flags

| Flag    | Default                 | Description                            |
| ------- | ----------------------- | -------------------------------------- |
| `-addr` | `http://localhost:9222` | CDP endpoint of the Chromium container |

Run `go run ./connect -h` for the full list.
