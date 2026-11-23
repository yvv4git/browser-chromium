# Headless Chromium (`chrome-headless-shell`)

Минимальный Docker-образ для автоматизации через Chrome DevTools Protocol (CDP) без GUI.

## Отличия от `docker/optimized`

| | `docker/headless` | `docker/optimized` |
|---|---|---|
| База | `node:22-bookworm-slim` | `debian:bookworm-slim` |
| Браузер | `chrome-headless-shell@stable` (Puppeteer) | системный `chromium` из apt |
| Режимы | только headless | `BROWSER_MODE=vnc` (по умолчанию) или `headless` |
| GUI | нет Xvfb/x11vnc/noVNC | Xvfb + x11vnc + noVNC |
| Порты | `9222`, `9223` (CDP) | `9222`, `9223`, `5900`, `3000` |
| Прокси | не настраивается в entrypoint | `PROXY`, `http_proxy/https_proxy`, bypass для `*.ru,localhost` |
| Профиль | нет persistent profile | `/config/chromium-data`, очистка `Singleton*` |
| Флаги запуска | минимальные + `CHROME_ARGS_EXTRA` | большой набор флагов тюнинга ресурсов/фич |
| Стартовая страница | `about:blank` | задаётся через `CHROME_CLI` |
| Назначение | лёгкий CDP-only контейнер | универсальный образ: UI/debug/headless + proxy |

## Что делает образ

- Устанавливает зависимости Chrome и `chrome-headless-shell` в `/usr/local/bin/chrome-headless-shell`.
- Пробрасывает внешний CDP-порт `9223` на внутренний `9222` через `socat`.
- Запускает браузер с `--remote-debugging-port=9222 --no-sandbox --disable-gpu`.
- Дополнительные флаги можно передать через `CHROME_ARGS_EXTRA`.

## Сборка и запуск

Из корня репозитория:

```bash
make image-build-headless
BROWSER_VARIANT=headless make compose-up
```

Проверка CDP:

```bash
curl http://localhost:9222/json/version
```

## Файлы

- `docker/headless/Dockerfile` — сборка образа.
- `docker/headless/entrypoint.sh` — запуск `socat` и `chrome-headless-shell`.
