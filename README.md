# geschenkerbauer

A Dockerfile and build scripts to build directories containing Arch packages.

## Usage
In order to build this, you will need an existing "arch" image in your local
Docker registry. You can use my
[docker-arch](https://github.com/mutantmonkey/docker-arch) tools to create one.

## Known Issues
* You will need to manually initialize a GnuPG keyring. The included
  init_keyring.sh may be helpful for this.
