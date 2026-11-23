FROM node:22-bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    dumb-init \
    fonts-liberation \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libcairo2 \
    libcups2 \
    libdbus-1-3 \
    libdrm2 \
    libgbm1 \
    libglib2.0-0 \
    libnspr4 \
    libnss3 \
    libpango-1.0-0 \
    libx11-6 \
    libxcb1 \
    libxcomposite1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxkbcommon0 \
    libxrandr2 \
    socat \
    unzip \
    && rm -rf /var/lib/apt/lists/* \
    && npx --yes @puppeteer/browsers install chrome-headless-shell@stable --path /opt/chrome \
    && CHROME_BIN=$(find /opt/chrome -type f -name chrome-headless-shell | head -n 1) \
    && ln -s "$CHROME_BIN" /usr/local/bin/chrome-headless-shell

ENV CHROME_ARGS=""

EXPOSE 9222 9223

COPY docker/headless/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/entrypoint.sh"]
