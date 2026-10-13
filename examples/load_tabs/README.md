# load_tabs

Load pattern for RAM/CPU measurements: opens several tabs at once, holds them
for a random time and closes them in a loop.

## Usage

The container must be running first:

```bash
cd .. && make compose-up
```

Run the example:

```bash
go run ./load_tabs -addr http://localhost:9222
```

## Flags

| Flag        | Default                 | Description                       |
| ----------- | ----------------------- | --------------------------------- |
| `-addr`     | `http://localhost:9222` | CDP endpoint of the container     |
| `-tabs`     | `5`                     | How many tabs to open at once     |
| `-cycles`   | `10`                    | How many open-hold-close cycles   |
| `-hold-min` | `10s`                   | Minimum tab hold time             |
| `-hold-max` | `40s`                   | Maximum tab hold time             |

Failed targets are skipped and logged, the run continues.

Run `go run ./load_tabs -h` for the full list.
