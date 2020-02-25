FROM archlinux:latest

RUN set -x \
        && pacman -Syu --noconfirm \
        && pacman -S --noconfirm git python python-requests

COPY docker-entrypoint.sh /
COPY update_check_pypi.py /
ENTRYPOINT ["/docker-entrypoint.sh"]
