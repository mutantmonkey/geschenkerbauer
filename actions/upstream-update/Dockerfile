FROM archlinux:base-devel

RUN set -x \
        && pacman -Syu --noconfirm \
        && pacman -S --noconfirm git python python-requests \
        && useradd -m -g users -u 1000 builduser

COPY docker-entrypoint.sh /
COPY update_check_pypi.py /
ENTRYPOINT ["/docker-entrypoint.sh"]

USER builduser
