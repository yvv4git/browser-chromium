# status_code

Loads a list of URLs in the Chromium container and prints the HTTP status code
of each main document response. The example demonstrates request interception
with `HijackRequests`:

- the `*` route matches document requests
- the request is continued and the response is loaded with `LoadResponse`
- the status code is read from the response payload

## Usage

The container must be running first:

```bash
cd .. && make compose-up
```

Run the example:

```bash
go run ./status_code -addr http://localhost:9222
```

## Flags

| Flag    | Default                 | Description                            |
| ------- | ----------------------- | -------------------------------------- |
| `-addr` | `http://localhost:9222` | CDP endpoint of the Chromium container |

Run `go run ./status_code -h` for the full list.
