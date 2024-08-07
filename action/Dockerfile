FROM ghcr.io/archlinux/archlinux:base-devel@sha256:e79be3531579b3ce13d80c6a5ce907d3e90acea9ac3411f0d663fe92fc2fb07b

RUN set -x \
        && useradd -m -g users -u 1000 builduser \
        && install -d -o builduser -g users /build /srcdest /srcpkgdest /logdest

COPY builduser-pacman /etc/sudoers.d/builduser-pacman
RUN chmod 440 /etc/sudoers.d/builduser-pacman

COPY makepkg.conf /etc/makepkg.conf
COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

USER builduser
