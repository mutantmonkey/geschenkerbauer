FROM ghcr.io/archlinux/archlinux:base@sha256:a02f5675c33bd07bacc758531e4cd3ed473463081fe75304214baececb422ba1

RUN set -x \
        && pacman -Sy --noconfirm namcap

COPY docker-entrypoint.py /
ENTRYPOINT ["/docker-entrypoint.py"]
