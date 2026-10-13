# Chromium Browser

![Chromium Browser logo](./assets/logo.jpeg)

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Image-2496ED?style=flat-square&logo=docker&logoColor=white)](docker/default/Dockerfile)
[![Browser](https://img.shields.io/badge/Browser-Chromium-0098EA?style=flat-square)](https://www.chromium.org)
[![Buy me a TON](https://img.shields.io/badge/Buy%20me%20a%20TON-0098EA?style=flat-square)](#support)

Docker image with Chromium browser with VNC/noVNC support
for remote access and automation, plus a variant with
[Selkies](https://github.com/selkies-project/selkies) (H.264/WebSocket
streaming) instead of VNC.

## Table of contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Usage](#usage)
- [Ports](#ports)
- [Project structure](#project-structure)
- [License](#license)
- [Support](#support)

## Features

- Chromium with headless support via Xvfb
- Interactive VNC access via noVNC web interface
- Alternative variant: Selkies low-latency H.264 streaming web UI (`docker/selkies`)
- Chrome DevTools Protocol (CDP) for automation
- HTTP proxy support
- Profile persistence via volume
- Log size limits (10 MB x 3 files)

## Quick Start

```bash
# Build image
make image-build

# Start container
make compose-up

# Stop container
make compose-down
```

Open noVNC in browser:

```text
http://localhost:3000
```

Or build and run the Selkies variant instead (H.264 over WebSocket,
port 8080):

```bash
make image-build-selkies
BROWSER_VARIANT=selkies make compose-up
```

```text
http://localhost:8080
```

Check CDP:

```bash
curl http://localhost:9222/json/version
```

## Configuration

Create a `.env` file (see `.env.example`):

| Variable            | Description                                       |
| ------------------- | ------------------------------------------------- |
| `BROWSER_VARIANT`   | Browser image version: `default`, `optimized` or `selkies` |
| `RESOLUTION`        | Screen resolution                                 |
| `PROXY`             | Proxy URL e.g. `http://host.docker.internal:8138` |
| `CHROME_ARGS_EXTRA` | Extra Chrome flags                                |
| `SELKIES_BASIC_AUTH_PASSWORD` | Selkies web UI login (empty = no login)  |

| Variable            | Default     |
| ------------------- | ----------- |
| `BROWSER_VARIANT`   | `default`   |
| `RESOLUTION`        | `1920x1080` |
| `PROXY`             | empty       |
| `CHROME_ARGS_EXTRA` | empty       |

Set `BROWSER_VARIANT=optimized` in `.env` to run the optimized image.

## Usage

Start with proxy:

```bash
docker compose up -d
```

Run an automation script against CDP:

```bash
curl http://localhost:9222/json/version
```

## Ports

| Port | Service                        |
| ---- | ------------------------------ |
| 9222 | Chrome DevTools Protocol (CDP) |
| 3000 | noVNC web interface            |
| 8080 | Selkies web interface (selkies variant) |

## Project structure

| Path                             | Description                                             |
| -------------------------------- | ------------------------------------------------------- |
| `docker/default/Dockerfile`      | Base image: Debian bookworm-slim with chromium, VNC     |
| `docker/default/entrypoint.sh`   | Starts Xvfb, VNC, noVNC and Chromium                    |
| `docker/optimized/Dockerfile`    | Resource-optimized image variant                        |
| `docker/optimized/entrypoint.sh` | Launcher for the optimized variant                      |
| `docker/selkies/Dockerfile`      | Selkies variant: chromium + Selkies instead of VNC      |
| `docker/selkies/entrypoint.sh`   | Starts Xvfb, Selkies and Chromium                       |
| `docker-compose.yaml`            | Service definition with ports and volumes               |
| `Makefile`                       | Build, run, stop commands                               |
| `.env.example`                   | Environment variables template                          |

## License

MIT

## Support

<p align="center">
  <a href="https://tonviewer.com/UQCcbp-mue-7HTjDNQ_ZrKtg-tUxIFu817APmItjXasiBGP3">
    <img src="https://img.shields.io/badge/Buy%20me%20a%20TON-0098EA?style=for-the-badge">
  </a>
</p>

<p align="center">
  If this tool helps you, consider buying me a coffee! ☕
</p>
