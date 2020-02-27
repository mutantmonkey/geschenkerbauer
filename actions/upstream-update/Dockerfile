FROM archlinux:latest

RUN set -x \
        && pacman -Syu --noconfirm \
        && useradd -m -g users -u 1000 builduser \
        && pacman -S --noconfirm git python python-requests

COPY docker-entrypoint.sh /
COPY update_check_pypi.py /
ENTRYPOINT ["/docker-entrypoint.sh"]

USER builduser
