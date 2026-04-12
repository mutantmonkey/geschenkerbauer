FROM docker.io/library/golang:1.26.2-trixie@sha256:da3943074756e8d6f109ce7a84be16e98bffa39e9d369a0447f016b56db84e8f AS builder

ADD autosign /usr/src/app/

WORKDIR /usr/src/app/cmd/autosign-receiver/

RUN go build

FROM ghcr.io/archlinux/archlinux:base@sha256:3b7c36a72d3e5b5996e2f29c911066a14efbb63f43491cde0dd5fef5c14b15e9

RUN set -x \
        && pacman -Sy --noconfirm github-cli

COPY --from=builder /usr/src/app/cmd/autosign-receiver/autosign-receiver /usr/local/bin/autosign-receiver

CMD ["/usr/local/bin/autosign-receiver", "-d", "-config", "/etc/autosign/autosign-receiver.toml"]
