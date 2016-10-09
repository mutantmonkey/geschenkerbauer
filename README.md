# geschenkerbauer

A Dockerfile and build scripts to build directories containing Arch packages.

## Known Issues
* There are several things that are hard coded that you probably will need to
  change in order to use this:
    * The following variables: buildhost, buildsrcdir, repodir, PACKAGER
    * The rsync line that copies ~/arch/packages/ to `$buildsrcdir` on
      `$buildhost`
* There is currently no support for verifying package signatures. Any PKGBUILDs
  that contain signature files in the sources will fail at the moment. I'm
  working on a solution for a persistent keyring.
