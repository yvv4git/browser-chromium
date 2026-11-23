#!/bin/bash
set -e

cleanup() {
    kill $CHROME_PID $SOCAT_PID 2>/dev/null
    wait $CHROME_PID $SOCAT_PID 2>/dev/null
}
trap cleanup SIGTERM SIGINT

socat TCP-LISTEN:9223,fork,reuseaddr TCP:127.0.0.1:9222 &
SOCAT_PID=$!

CHROME_ARGS_DEFAULT=(
    --remote-debugging-port=9222
    --no-sandbox
    --disable-gpu
)

if [ -n "$CHROME_ARGS_EXTRA" ]; then
    CHROME_ARGS_DEFAULT+=($CHROME_ARGS_EXTRA)
fi

/usr/local/bin/chrome-headless-shell "${CHROME_ARGS_DEFAULT[@]}" about:blank &
CHROME_PID=$!
wait $CHROME_PID
