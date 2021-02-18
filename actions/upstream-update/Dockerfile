FROM python:3

RUN set -x \
        && pip install --no-cache-dir requests \
        && useradd -m -g users -u 1000 builduser

COPY docker-entrypoint.sh /
COPY update_check_pypi.py /
ENTRYPOINT ["/docker-entrypoint.sh"]

USER builduser
