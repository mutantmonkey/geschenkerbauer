FROM archlinux:latest

# This is not a typo; we need git and hub, which are separate tools
RUN set -x \
        && pacman -Syu --noconfirm \
        && pacman -S git hub

COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]
