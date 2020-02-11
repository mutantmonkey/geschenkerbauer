FROM mutantmonkey/geschenkerbauer:latest

ENV MMKEYRING_VERSION 20170105

# reinitialize the pacman keyring
RUN pacman-key --init && pacman-key --populate archlinux

COPY --chown=1000 sha512sums.txt /tmp/
RUN cd /tmp \
        && curl -O https://deadbeef.ninja/archlinux-user/mutantmonkey/os/x86_64/mutantmonkey-keyring-$MMKEYRING_VERSION-1-any.pkg.tar.xz \
        && sha512sum -c sha512sums.txt \
        && sudo pacman --noconfirm -U mutantmonkey-keyring-$MMKEYRING_VERSION-1-any.pkg.tar.xz \
        && rm mutantmonkey-keyring-$MMKEYRING_VERSION-1-any.pkg.tar.xz sha512sums.txt
        && rm -rf etc/pacman.d/gnupg/{openpgp-revocs.d/,private-keys-v1.d/,pugring.gpg~,gnupg.S.}*

USER root
RUN sed -i -e '$a \
        \\n[mutantmonkey]\nServer = https://deadbeef.ninja/archlinux-user/mutantmonkey/os/$arch' /etc/pacman.conf
USER builduser
