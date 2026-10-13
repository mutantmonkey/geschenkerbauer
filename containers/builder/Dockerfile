FROM ghcr.io/archlinux/archlinux:base-devel@sha256:31659a2c121c7da1ad668457cc0c1eed906f2cd484e9aa572cddcf5d660ae8f6

RUN set -x \
        && useradd -m -g users -u 1000 builduser \
        && install -d -o builduser -g users /build /srcdest /srcpkgdest /logdest /startdir

COPY builduser-pacman /etc/sudoers.d/builduser-pacman
RUN chmod 440 /etc/sudoers.d/builduser-pacman

COPY makepkg.conf /etc/makepkg.conf
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

USER builduser
