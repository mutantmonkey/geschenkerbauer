FROM archlinux:base

RUN set -x \
        && useradd -m -g users -u 1000 builduser \
        && pacman -Sy --noconfirm signify

COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

USER builduser
