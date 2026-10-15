FROM debian:bookworm-slim@sha256:a4672c0cb26fbdde88e38fa2dfb6c681942306680e41e4378b28770b6e79ee91
LABEL hive.pilot="flash-class-guidance"
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl git procps \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /tmp/install
RUN curl -fsSLO https://github.com/Gentleman-Programming/engram/releases/download/v3.3.2/engram_3.3.2_linux_amd64.tar.gz \
    && curl -fsSLO https://github.com/Gentleman-Programming/engram/releases/download/v3.3.2/checksums.txt \
    && grep '  engram_3.3.2_linux_amd64.tar.gz$' checksums.txt > selected-checksum.txt \
    && grep '^8694d3817b63c5fc5607b5919e88aa929bffa0bafda330513364e10c8ab61445 ' selected-checksum.txt \
    && sha256sum -c selected-checksum.txt \
    && tar -xzf engram_3.3.2_linux_amd64.tar.gz \
    && install -m 0755 engram /usr/local/bin/engram
# npm dist.integrity, confirmed against the registry for this exact version.
RUN curl -fsSL https://registry.npmjs.org/@opencode/cli-linux-x64/-/cli-linux-x64-2.0.27.tgz -o opencode.tgz \
    && echo '40f1abcc1492b390131b79cce178d2e2d9965d4076d838b7e87d2f7ccba12b093a8b73d1dcd6f8b26fa5d3482f4dd732091a2162c2b0aa3bc3ebb4f40cf7102d opencode.tgz' | sha512sum -c - \
    && tar -xzf opencode.tgz \
    && install -m 0755 package/bin/opencode /usr/local/bin/opencode \
    && rm -rf /tmp/install
# The device login tries to open a browser and aborts when xdg-open is missing;
# the user opens the printed URL on their own machine instead.
RUN printf '#!/bin/sh\nexit 0\n' > /usr/local/bin/xdg-open && chmod 0755 /usr/local/bin/xdg-open
RUN groupadd --gid 1000 pilot && useradd --uid 1000 --gid 1000 --create-home pilot \
    && mkdir -p /work/runs /work/input /home/pilot/.cache /home/pilot/.config/opencode /run/hive-pilot/secrets \
    && chown -R 1000:1000 /work /home/pilot /run/hive-pilot \
    && printf '%s\n' '{"mcp":{"servers":{"engram":{"type":"local","command":["engram","mcp"]}}}}' > /home/pilot/.config/opencode/opencode.json
ENV HOME=/home/pilot XDG_DATA_HOME=/run/hive-pilot/secrets/real-data ENGRAM_NO_UPDATE_CHECK=1
USER 1000:1000
WORKDIR /work
CMD ["sleep", "infinity"]
