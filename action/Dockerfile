FROM ghcr.io/archlinux/archlinux:base-devel@sha256:0d379416265205cfc23f5ad29aaae98f9deb9baeae733d93d514bfa98c9f46d4

RUN set -x \
        && useradd -m -g users -u 1000 builduser \
        && install -d -o builduser -g users /build /srcdest /srcpkgdest /logdest

COPY builduser-pacman /etc/sudoers.d/builduser-pacman
RUN chmod 440 /etc/sudoers.d/builduser-pacman

COPY makepkg.conf /etc/makepkg.conf
COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

USER builduser
