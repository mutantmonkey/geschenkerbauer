FROM archlinux:base-devel

RUN set -x \
        && useradd -m -g users -u 1000 builduser \
        && install -d -o builduser -g users /build /srcdest /srcpkgdest /logdest

COPY builduser-pacman /etc/sudoers.d/builduser-pacman
RUN chmod 440 /etc/sudoers.d/builduser-pacman

# add local repo to pacman.conf
RUN sed -i '/\[testing\]/i \
        [geschenkerbauer]\nSigLevel = Never\nServer = file:///repo\n' /etc/pacman.conf

VOLUME ["/buildsrc", "/repo"]

COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.sh"]

USER builduser
