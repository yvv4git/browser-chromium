#!/bin/bash
set -e

cleanup() {
    kill $CHROME_PID $X11VNC_PID $NOVNC_PID $XVFB_PID 2>/dev/null
    wait $CHROME_PID $X11VNC_PID $NOVNC_PID $XVFB_PID 2>/dev/null
}
trap cleanup SIGTERM SIGINT

RESOLUTION=${RESOLUTION:-1920x1080}
PROFILE_DIR=/config/chromium-data

mkdir -p "$PROFILE_DIR"
rm -f "$PROFILE_DIR"/Singleton*

Xvfb :99 -screen 0 ${RESOLUTION}x24 -ac -nolisten tcp &
XVFB_PID=$!

for i in $(seq 10); do
    if xdpyinfo -display :99 >/dev/null 2>&1; then
        break
    fi
    sleep 1
done

socat TCP-LISTEN:9223,fork,reuseaddr TCP:127.0.0.1:9222 &

x11vnc -display :99 -forever -nopw -quiet &
X11VNC_PID=$!

/usr/share/novnc/utils/novnc_proxy --vnc 127.0.0.1:5900 --listen 3000 --web /usr/share/novnc &
NOVNC_PID=$!

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
