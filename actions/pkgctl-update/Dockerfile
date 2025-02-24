FROM ghcr.io/archlinux/archlinux:base-devel@sha256:d4d292c2614c0577ba9ba3240631fd3004eb8fddb83a743ff72bea9c4a96f985

RUN set -x \
        && pacman -Syu --noconfirm
        && pacman -S --noconfirm devtools
        && useradd -m -g users -u 1000 builduser

COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

USER builduser
