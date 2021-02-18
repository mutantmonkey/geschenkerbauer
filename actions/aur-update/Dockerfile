FROM debian:buster

# This is not a typo; we need git and hub, which are separate tools
RUN set -x \
        && apt-get update \
        && apt-get install -y git hub \
        && apt-get clean \
        && rm -rf /var/lib/apt/lists/*

COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]
