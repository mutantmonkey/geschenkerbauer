FROM ghcr.io/archlinux/archlinux:base-devel@sha256:e57a29572b0c6ff1e8ea79a4fe6475af89ef5cc281604623bb421591fe0e5805

RUN set -x \
        && pacman -Syu --noconfirm \
        && pacman -S --noconfirm devtools nvchecker \
        && useradd -m -g users -u 1000 builduser

COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

#USER builduser
