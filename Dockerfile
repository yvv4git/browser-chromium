FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    chromium \
    chromium-common \
    chromium-l10n \
    socat \
    xvfb \
    x11vnc \
    novnc \
    x11-utils \
    ca-certificates \
    curl \
    dnsutils \
    dumb-init \
    fonts-liberation \
    fonts-noto-color-emoji \
    fontconfig \
    && rm -rf /var/lib/apt/lists/*

ENV DISPLAY=:99
ENV RESOLUTION=1920x1080
ENV CHROME_ARGS=""
ENV PROXY=""

EXPOSE 9222 9223 5900 3000

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/entrypoint.sh"]
