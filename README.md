# geschenkerbauer

A Dockerfile and build scripts to build directories containing Arch packages.

## Known Issues
* There are several things that are hard coded that you probably will need to
  change in order to use this:
    * The following variables: buildhost, buildsrcdir, repodir, PACKAGER
    * The rsync line that copies ~/arch/packages/ to `$buildsrcdir` on
      `$buildhost`
* You will need to manually initialize a GnuPG keyring. The included
  init_keyring.sh may be helpful for this.
