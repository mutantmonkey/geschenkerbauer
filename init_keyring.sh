#!/bin/sh
gpg2 --homedir /home/core/arch/gnupg --gen-key --batch <<EOF
%echo Generating geschenkerbauer keyring master key...
Key-Type: RSA
Key-Length: 2048
Key-Usage: sign
Name-Real: Geschenkerbauer Keyring Master Key
Name-Email: geschenkerbauer@localhost
Expire-Date: 0
%commit
%echo Done
EOF
