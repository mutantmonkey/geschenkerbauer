FROM archlinux:base

RUN set -x \
        && pacman -Sy --noconfirm namcap

COPY docker-entrypoint.py /
ENTRYPOINT ["/docker-entrypoint.py"]
