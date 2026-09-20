FROM ghcr.io/archlinux/archlinux:base-devel@sha256:5eec215aabd3912afe2649fe6e7fe2912d12692d6b5d4cae6c950a2b065793cd

RUN set -x \
        && useradd -m -g users -u 1000 builduser \
        && install -d -o builduser -g users /build /srcdest /srcpkgdest /logdest /startdir

COPY builduser-pacman /etc/sudoers.d/builduser-pacman
RUN chmod 440 /etc/sudoers.d/builduser-pacman

COPY makepkg.conf /etc/makepkg.conf
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

USER builduser
