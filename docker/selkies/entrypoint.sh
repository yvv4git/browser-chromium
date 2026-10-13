#!/bin/bash
set -e

cleanup() {
    kill $CHROME_PID $SELKIES_PID $XVFB_PID $WINFIT_PID 2>/dev/null
    wait $CHROME_PID $SELKIES_PID $XVFB_PID 2>/dev/null
}
trap cleanup SIGTERM SIGINT

RESOLUTION=${RESOLUTION:-1920x1080}
MAX_RESOLUTION=${MAX_RESOLUTION:-3840x2160}
SELKIES_PORT=${SELKIES_PORT:-8080}
SELKIES_ENABLE_RESIZE=${SELKIES_ENABLE_RESIZE:-true}
PROFILE_DIR=/config/chromium-data

mkdir -p "$PROFILE_DIR"
rm -f "$PROFILE_DIR"/Singleton*

Xvfb :99 -screen 0 ${MAX_RESOLUTION}x24 -ac -nolisten tcp -s 0 -dpms -noreset \
    +extension RANDR +extension MIT-SHM +extension XFIXES +extension XTEST \
    +extension RENDER +extension DAMAGE &
XVFB_PID=$!

for i in $(seq 10); do
    if xdpyinfo -display :99 >/dev/null 2>&1; then
        break
    fi
    sleep 1
done

socat TCP-LISTEN:9223,fork,reuseaddr TCP:127.0.0.1:9222 &

export DISPLAY=:99
export SELKIES_PORT
export SELKIES_PUBLIC=${SELKIES_PUBLIC:-true}
export SELKIES_ENABLE_HTTPS=${SELKIES_ENABLE_HTTPS:-false}
export SELKIES_AUDIO_ENABLED=${SELKIES_AUDIO_ENABLED:-false}
export SELKIES_ENABLE_RESIZE

if [ -n "$SELKIES_BASIC_AUTH_PASSWORD" ]; then
    export SELKIES_BASIC_AUTH_USER=${SELKIES_BASIC_AUTH_USER:-ubuntu}
    export SELKIES_ENABLE_BASIC_AUTH=${SELKIES_ENABLE_BASIC_AUTH:-true}
else
    export SELKIES_ENABLE_BASIC_AUTH=${SELKIES_ENABLE_BASIC_AUTH:-false}
fi

if [ "$SELKIES_ENABLE_RESIZE" = "true" ] || [ "$SELKIES_ENABLE_RESIZE" = "1" ]; then
    selkies-resize "$RESOLUTION" || echo "selkies-resize failed, keeping ${MAX_RESOLUTION}"
else
    xrandr --display :99 --fb "${RESOLUTION/x/,}" 2>/dev/null || true
fi

env -u LD_PRELOAD selkies $SELKIES_CLI &
SELKIES_PID=$!

(
    while true; do
        cur=$(xrandr --display :99 2>/dev/null | awk '/\*/{print $1; exit}')
        w=${cur%x*}; h=${cur#*x}
        if [ "$w" -gt 100 ] 2>/dev/null && [ "$h" -gt 100 ] 2>/dev/null; then
            xdotool search --onlyvisible --name 'Chromium$' 2>/dev/null | while read -r id; do
                xdotool windowsize "$id" "$w" "$h" 2>/dev/null || true
            done
        fi
        sleep 2
    done
) &
WINFIT_PID=$!

CHROME_ARGS=(
    --no-first-run
    --no-sandbox
    --password-store=basic
    --disable-gpu
    --disable-dev-shm-usage
    --disable-software-rasterizer
    --disable-async-dns
    --disable-features=WebRTC
    --simulate-outdated-no-au='Tue, 31 Dec 2099 23:59:59 GMT'
    --window-size=${RESOLUTION/x/,}
    --test-type
    --remote-debugging-port=9222
    --user-data-dir="$PROFILE_DIR"
)

if [ -n "$PROXY" ]; then
    echo "proxy: $PROXY"
    CHROME_ARGS+=(--proxy-server="$PROXY")
    CHROME_ARGS+=(--proxy-bypass-list="*.ru,<-loopback>")
    export http_proxy="$PROXY"
    export https_proxy="$PROXY"
    export no_proxy=".ru,localhost,127.0.0.1,[::1]"
else
    echo "proxy: none"
fi

if [ -n "$CHROME_ARGS_EXTRA" ]; then
    CHROME_ARGS+=($CHROME_ARGS_EXTRA)
fi

/usr/bin/chromium "${CHROME_ARGS[@]}" $CHROME_CLI &
CHROME_PID=$!
wait $CHROME_PID
