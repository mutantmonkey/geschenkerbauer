FROM docker.io/library/golang:1.23.0-bookworm@sha256:89a1b98608f91fe47734cac497c0d7bd0c88239f90cccd00ea47057eb09604e2 AS builder

ADD autosign /usr/src/app/

WORKDIR /usr/src/app/cmd/autosign-receiver/

RUN go build

FROM ghcr.io/archlinux/archlinux:base@sha256:c404a543f3c04983b4f8a0ad14330f96cf9f582a9b673ce9b9b1c6e91e0309f2

RUN set -x \
        && pacman -Sy --noconfirm github-cli

COPY --from=builder /usr/src/app/cmd/autosign-receiver/autosign-receiver /usr/local/bin/autosign-receiver

CMD ["/usr/local/bin/autosign-receiver", "-d", "-config", "/etc/autosign/autosign-receiver.toml"]
