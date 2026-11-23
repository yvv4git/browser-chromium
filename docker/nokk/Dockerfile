FROM debian:bookworm-slim

ARG NOKK_VERSION=v0.1.50-alpha
ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    dumb-init \
    libstdc++6 \
    socat \
    tar \
    && rm -rf /var/lib/apt/lists/* \
    && case "$TARGETARCH" in \
         amd64) NOKK_ARCH=x86_64 ;; \
         arm64) NOKK_ARCH=aarch64 ;; \
         *) echo "unsupported TARGETARCH=$TARGETARCH" >&2; exit 1 ;; \
       esac \
    && curl -fsSL "https://github.com/koloss777/nokk/releases/download/${NOKK_VERSION}/nokk-${NOKK_VERSION}-linux-${NOKK_ARCH}.tar.gz" -o /tmp/nokk.tar.gz \
    && tar -xzf /tmp/nokk.tar.gz -C /usr/local/bin \
    && rm /tmp/nokk.tar.gz \
    && chmod +x /usr/local/bin/nokk

ENV NOKK_ARGS=""

EXPOSE 9222 9223

COPY docker/nokk/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/entrypoint.sh"]
