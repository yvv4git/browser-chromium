#!/bin/bash
set -e

cleanup() {
    kill $NOKK_PID $SOCAT_PID 2>/dev/null
    wait $NOKK_PID $SOCAT_PID 2>/dev/null
}
trap cleanup SIGTERM SIGINT

socat TCP-LISTEN:9223,fork,reuseaddr TCP:127.0.0.1:9222 &
SOCAT_PID=$!

NOKK_ARGS_DEFAULT=(
    --host 127.0.0.1
    --port 9222
)

if [ -n "$NOKK_ARGS" ]; then
    NOKK_ARGS_DEFAULT+=($NOKK_ARGS)
fi

/usr/local/bin/nokk "${NOKK_ARGS_DEFAULT[@]}" &
NOKK_PID=$!
wait $NOKK_PID
