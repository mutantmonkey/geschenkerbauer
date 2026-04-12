FROM docker.io/library/golang:1.23.0-bookworm@sha256:89a1b98608f91fe47734cac497c0d7bd0c88239f90cccd00ea47057eb09604e2 AS builder

ADD autosign /usr/src/app/

WORKDIR /usr/src/app/cmd/autosign-receiver/

RUN go build

FROM ghcr.io/archlinux/archlinux:base@sha256:3b7c36a72d3e5b5996e2f29c911066a14efbb63f43491cde0dd5fef5c14b15e9

RUN set -x \
        && pacman -Sy --noconfirm github-cli

COPY --from=builder /usr/src/app/cmd/autosign-receiver/autosign-receiver /usr/local/bin/autosign-receiver

CMD ["/usr/local/bin/autosign-receiver", "-d", "-config", "/etc/autosign/autosign-receiver.toml"]
