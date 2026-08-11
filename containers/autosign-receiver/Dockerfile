FROM docker.io/library/golang:1.26.2-trixie@sha256:da3943074756e8d6f109ce7a84be16e98bffa39e9d369a0447f016b56db84e8f AS builder

ADD autosign /usr/src/app/

WORKDIR /usr/src/app/cmd/autosign-receiver/

RUN go build

FROM ghcr.io/archlinux/archlinux:base@sha256:2c17dd486378e9549e2b891d0c6fe87b63b865cc7a0a19fca8e72897a044a11e

RUN set -x \
        && pacman -Sy --noconfirm github-cli

COPY --from=builder /usr/src/app/cmd/autosign-receiver/autosign-receiver /usr/local/bin/autosign-receiver

CMD ["/usr/local/bin/autosign-receiver", "-d", "-config", "/etc/autosign/autosign-receiver.toml"]
