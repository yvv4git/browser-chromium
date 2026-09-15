# Chromium Browser

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Image-2496ED?style=flat-square&logo=docker&logoColor=white)](Dockerfile)
[![Browser](https://img.shields.io/badge/Browser-Chromium-0098EA?style=flat-square)](https://www.chromium.org)
[![Buy me a TON](https://img.shields.io/badge/Buy%20me%20a%20TON-0098EA?style=flat-square)](#support)

Docker image with Chromium browser with VNC/noVNC support
for remote access and automation.

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

Check CDP:

```bash
curl http://localhost:9222/json/version
```

## Configuration

Create a `.env` file (see `.env.example`):

| Variable            | Description                                       |
| ------------------- | ------------------------------------------------- |
| `RESOLUTION`        | Screen resolution                                 |
| `PROXY`             | Proxy URL e.g. `http://host.docker.internal:8138` |
| `CHROME_ARGS_EXTRA` | Extra Chrome flags                                |

| Variable            | Default     |
| ------------------- | ----------- |
| `RESOLUTION`        | `1920x1080` |
| `PROXY`             | empty       |
| `CHROME_ARGS_EXTRA` | empty       |

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

## Project structure

| File                  | Description                                             |
| --------------------- | ------------------------------------------------------- |
| `Dockerfile`          | Debian bookworm-slim with chromium, xvfb, x11vnc, novnc |
| `entrypoint.sh`       | Starts Xvfb, VNC, noVNC and Chromium                    |
| `docker-compose.yaml` | Service definition with ports and volumes               |
| `Makefile`            | Build, run, stop commands                               |
| `.env.example`        | Environment variables template                          |

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
