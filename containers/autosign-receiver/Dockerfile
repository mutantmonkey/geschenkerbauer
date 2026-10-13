FROM docker.io/library/golang:1.27-trixie@sha256:9baa6b4187bbb98d240372a8a235ac0bb6b5ddd52bba1431dc2f7c0705862728 AS builder

ADD autosign /usr/src/app/

WORKDIR /usr/src/app/cmd/autosign-receiver/

RUN go build

FROM ghcr.io/archlinux/archlinux:base@sha256:86092797e32aa19d3f45a831b2873cf557c21c5280e9d197ca7a43c8ea79bcd3

RUN set -x \
        && pacman -Sy --noconfirm github-cli

COPY --from=builder /usr/src/app/cmd/autosign-receiver/autosign-receiver /usr/local/bin/autosign-receiver

CMD ["/usr/local/bin/autosign-receiver", "-d", "-config", "/etc/autosign/autosign-receiver.toml"]
