FROM archlinux:base-devel

RUN set -x \
        && useradd -m -g users -u 1000 builduser \
        && pacman -S signify

COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

USER builduser
