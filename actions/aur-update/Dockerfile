FROM debian:buster

# This is not a typo; we need git and hub, which are separate tools
RUN set -x \
        && apt install -y git hub

COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]
