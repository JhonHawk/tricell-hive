# Native gate for the optional agent-browser capability (change
# agent-browser-optional-install, task T5). Linux amd64 only, the one Linux
# pair the installer offers. It carries Node, a global npm prefix writable by
# UID 1000 and Chrome's shared libraries, but no chromium package: a chromium
# on PATH would suppress the Chrome step. Build with
#   docker build --platform linux/amd64 -f agent-browser-gate.Dockerfile .
# and add --build-arg NODE_VERSION=v22.23.3 for the variant that must stay manual.
FROM debian:bookworm-slim@sha256:a4672c0cb26fbdde88e38fa2dfb6c681942306680e41e4378b28770b6e79ee91
LABEL hive.gate="agent-browser-optional-install"
ARG NODE_VERSION=v24.21.0
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl xz-utils procps \
    fonts-liberation libasound2 libatk-bridge2.0-0 libatk1.0-0 libcairo2 libcups2 libdbus-1-3 libdrm2 \
    libgbm1 libglib2.0-0 libnspr4 libnss3 libpango-1.0-0 libx11-6 libxcb1 libxcomposite1 libxdamage1 \
    libxext6 libxfixes3 libxkbcommon0 libxrandr2 \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /tmp/install
RUN curl -fsSLO "https://nodejs.org/dist/${NODE_VERSION}/node-${NODE_VERSION}-linux-x64.tar.xz" \
    && curl -fsSLO "https://nodejs.org/dist/${NODE_VERSION}/SHASUMS256.txt" \
    && grep " node-${NODE_VERSION}-linux-x64.tar.xz\$" SHASUMS256.txt | sha256sum -c - \
    && tar -xJf "node-${NODE_VERSION}-linux-x64.tar.xz" -C /usr/local --strip-components=1 \
    && rm -rf /tmp/install
RUN groupadd --gid 1000 gate && useradd --uid 1000 --gid 1000 --create-home gate \
    && mkdir -p /home/gate/.npm-global /work && chown -R 1000:1000 /home/gate /work
ENV HOME=/home/gate NPM_CONFIG_PREFIX=/home/gate/.npm-global PATH=/home/gate/.npm-global/bin:/usr/local/bin:/usr/bin:/bin
USER 1000:1000
WORKDIR /work
CMD ["sleep", "infinity"]
