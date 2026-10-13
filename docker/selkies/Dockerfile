FROM debian:bookworm-slim

ARG SELKIES_VERSION=2.0.0

RUN apt-get update && apt-get install -y --no-install-recommends \
    chromium \
    chromium-common \
    chromium-l10n \
    socat \
    xvfb \
    x11-utils \
    x11-xserver-utils \
    ca-certificates \
    curl \
    dnsutils \
    dumb-init \
    fonts-liberation \
    fonts-noto-color-emoji \
    fontconfig \
    libva2 \
    libva-drm2 \
    libva-x11-2 \
    python3 \
    python3-venv \
    && rm -rf /var/lib/apt/lists/*

RUN python3 -m venv /opt/selkies \
    && /opt/selkies/bin/pip install --no-cache-dir "selkies==${SELKIES_VERSION}"

ENV PATH="/opt/selkies/bin:${PATH}"

ENV DISPLAY=:99
ENV RESOLUTION=1920x1080
ENV CHROME_ARGS=""
ENV PROXY=""
ENV SELKIES_PORT=8080

EXPOSE 9222 9223 8080

COPY docker/selkies/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/entrypoint.sh"]
