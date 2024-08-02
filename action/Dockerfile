FROM ghcr.io/archlinux/archlinux:base-devel@sha256:42c83e3cbbe60a5f6afb7025a86eb9f4f39cd7555e5ddcab12662e3ea551ca91

RUN set -x \
        && useradd -m -g users -u 1000 builduser \
        && install -d -o builduser -g users /build /srcdest /srcpkgdest /logdest

COPY builduser-pacman /etc/sudoers.d/builduser-pacman
RUN chmod 440 /etc/sudoers.d/builduser-pacman

COPY makepkg.conf /etc/makepkg.conf
COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

USER builduser
