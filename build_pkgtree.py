#!/usr/bin/python3

import argparse
import os.path
import re
import subprocess
import sys
import tarfile


class Package(object):
    def __init__(self, pkgname=None, pkgver=None):
        self.name = pkgname
        self.version = pkgver
        self.provides = set([])


def parse_pkgdesc(f):
    pkg = Package()
    current_section = None

    for line in f:
        if type(line) == str:
            line = line.strip()
        else:
            line = line.decode('utf-8').strip()

        if current_section is not None and len(line) <= 0:
            current_section = None
        elif line == '%NAME%':
            current_section = 'name'
        elif line == '%VERSION%':
            current_section = 'version'
        elif line == '%PROVIDES%':
            current_section = 'provides'
        elif current_section == 'name':
            pkg.name = line
        elif current_section == 'version':
            pkg.version = line
        elif current_section == 'provides':
            pkg.provides.add(line.split('=')[0])

    return pkg


def list_db(dbpath):
    pkgs = set([])
    with tarfile.open(dbpath) as tf:
        for member in tf.getmembers():
            if os.path.basename(member.name) == 'desc':
                f = tf.extractfile(member)
                pkg = parse_pkgdesc(f)
                pkgs.add(pkg.name)
                pkgs = pkgs.union(pkg.provides)

    return pkgs


def list_installed_packages(dbpath):
    pkgs = set([])
    for rootdirname, dirnames, _ in os.walk(os.path.join(dbpath, 'local')):
        for dirname in dirnames:
            pkgname, pkgver, pkgrel = dirname.rsplit('-', 2)
            pkgs.add(pkgname)

            with open(os.path.join(dbpath, dirname, 'desc')) as f:
                pkg = parse_pkgdesc(f)
                pkgs.add(pkg.name)
                pkgs = pkgs.union(pkg.provides)

    return pkgs


def get_dependencies(pkgname):
    pkgs = []
    split_re = re.compile('[<>=]+')

    output = subprocess.check_output(['makepkg', '--printsrcinfo'],
                                     cwd=pkgname)
    for line in output.decode('utf-8').splitlines():
        data = line.split(' = ', 1)
        if len(data) < 2:
            continue

        if data[0] in ('\tdepends', '\tmakedepends'):
            pkgs.append(split_re.split(data[1], 1)[0])

    return pkgs


def build_deptree(pkg, skip_pkgs):
    pkgs = []
    deps = get_dependencies(pkg)
    for dep in deps:
        if os.path.exists(dep):
            deptree = build_deptree(dep, skip_pkgs)
            for d in deptree:
                pkgs.append(d)
        elif dep not in skip_pkgs:
            print("Warning: {0} not found".format(dep), file=sys.stderr)

    pkgs.append(pkg)
    return pkgs


if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description="Build a package (and its dependencies) with "
                    "geschenkerbauer")
    parser.add_argument('--buildhost', help="Build host")
    parser.add_argument('--buildscript',
                        default="arch/geschenkerbauer/build_pkg_server.sh",
                        help="Script used to run the build")
    parser.add_argument('--dbpath', default="/var/lib/pacman",
                        help="Path to pacman database directory")
    parser.add_argument('pkgs', nargs='+')
    args = parser.parse_args()

    skip_pkgs = set([])
    for syncdb in ['core', 'extra', 'community']:
        skip_pkgs = skip_pkgs.union(list_db(
            os.path.join(args.dbpath, 'sync', '{0}.db'.format(syncdb))))

    #skip_pkgs = skip_pkgs.union(list_installed_packages(args.dbpath))

    pkgs_to_build = []
    for pkg in args.pkgs:
        pkg_deptree = build_deptree(pkg, skip_pkgs)
        print(pkg, pkg_deptree)
        for bpkg in pkg_deptree:
            if bpkg not in pkgs_to_build:
                pkgs_to_build.append(bpkg)

    print(pkgs_to_build)

    # TODO: copy arch/packages to buildsrcdir

    subprocess.run(['ssh', args.buildhost, args.buildscript],
                   input='\n'.join(pkgs_to_build), encoding='utf-8')
