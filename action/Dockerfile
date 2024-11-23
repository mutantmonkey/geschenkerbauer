FROM ghcr.io/archlinux/archlinux:base-devel@sha256:94f5379e59ffb97d42c6e58e328dc7eba42fde6cdb81af92e55f89947871c6ed

RUN set -x \
        && useradd -m -g users -u 1000 builduser \
        && install -d -o builduser -g users /build /srcdest /srcpkgdest /logdest /startdir

COPY builduser-pacman /etc/sudoers.d/builduser-pacman
RUN chmod 440 /etc/sudoers.d/builduser-pacman

COPY makepkg.conf /etc/makepkg.conf
COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

USER builduser
