FROM ghcr.io/archlinux/archlinux:base-devel@sha256:cba3690d1b5251461cc3418397f4d8f0668acb28ba80ebf81a27eb907cdcacfd

RUN set -x \
        && useradd -m -g users -u 1000 builduser \
        && install -d -o builduser -g users /build /srcdest /srcpkgdest /logdest /startdir

COPY builduser-pacman /etc/sudoers.d/builduser-pacman
RUN chmod 440 /etc/sudoers.d/builduser-pacman

COPY makepkg.conf /etc/makepkg.conf
COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]

USER builduser
