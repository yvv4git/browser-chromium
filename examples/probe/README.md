# probe

Prints browser fingerprint signals that anti-bot systems check:

- `navigator.webdriver`, `navigator.plugins`, `navigator.languages`
- `navigator.hardwareConcurrency`, `navigator.deviceMemory`
- `window.chrome`, WebGL renderer string
- `performance.memory` JS heap size

Use it to compare fingerprints between the `default` and `optimized` images.

## Usage

The container must be running first:

```bash
cd .. && make compose-up
```

Run the example:

```bash
go run ./probe -addr http://localhost:9222
```

## Flags

| Flag    | Default                 | Description                            |
| ------- | ----------------------- | -------------------------------------- |
| `-addr` | `http://localhost:9222` | CDP endpoint of the Chromium container |

Run `go run ./probe -h` for the full list.
