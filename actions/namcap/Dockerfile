FROM archlinux:base

RUN set -x \
        && pacman -Sy --noconfirm namcap

COPY docker-entrypoint.sh /
ENTRYPOINT ["/docker-entrypoint.py"]
