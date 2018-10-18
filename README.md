# geschenkerbauer

A Dockerfile and build scripts to build directories containing Arch packages.

## Usage
Just run `build_pkgtree.py` in the same directory as the Arch package you'd
like to build; you'll just need to specify `--buildhost` and the name of the
package to build.

## Known Issues
* You will need to manually initialize a GnuPG keyring. The included
  init_keyring.sh may be helpful for this.
