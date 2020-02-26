#!/usr/bin/python3

import os.path
import re
import requests
import subprocess
import sys
from collections import defaultdict
from distutils.version import StrictVersion

base_pkgdir = os.environ.get('GITHUB_WORKSPACE')
repo_path = os.environ.get('GITHUB_REPOSITORY')
github_user = os.environ.get('GITHUB_USER')
github_token = os.environ.get('GITHUB_TOKEN')

packages = []
for pkg in os.environ.get('INPUT_PYPI_PACKAGES').splitlines():
    pkg = pkg.strip()
    if len(pkg) > 0:
        packages.append(pkg.split(': ', 1))


def parse_srcinfo(path):
    """Parse a SRCINFO file according to the specification:
    https://wiki.archlinux.org/index.php/.SRCINFO

    This function does not implement strict parsing as described in the
    specification but instead only provides sufficient functionality for this
    particular use case."""

    parsed = defaultdict(list)
    with open(path) as f:
        for line in f:
            line = line.strip()

            # skip commented and blank lines
            if line.startswith('#') or len(line) == 0:
                continue

            k, v = line.split(' = ', 1)
            if k in ('pkgbase', 'pkgver', 'pkgrel', 'epoch'):
                parsed[k] = v
            else:
                parsed[k].append(v)

    return parsed


digest_to_srcinfo_key = {
    'md5': 'md5sums',
    'sha1': 'sha1sums',
    'sha256': 'sha256sums',
    'sha384': 'sha384sums',
    'sha512': 'sha512sums',
}

for pkg in packages:
    arch_pkgname = pkg[0]
    pypi_pkgname = pkg[1]
    pkgdir = os.path.join(base_pkgdir, arch_pkgname)
    has_update = False

    r = requests.get('https://pypi.org/pypi/{0}/json'.format(pypi_pkgname))
    if r.status_code != 200:
        print(
            "::error::Failed to check {0} for updates on PyPI "
            "(HTTP status {1})".format(pypi_pkgname, r.status_code),
            file=sys.stderr)
        continue

    data = r.json()
    latest_version = data['info']['version']
    srcinfo = parse_srcinfo(os.path.join(pkgdir, '.SRCINFO'))

    # find updated release package and get new digest values
    for release_pkg in data['releases'][latest_version]:
        if release_pkg['packagetype'] == 'sdist':
            if StrictVersion(latest_version) > StrictVersion(
                    srcinfo['pkgver']):
                has_update = True
                srcinfo['pkgver'] = latest_version

                if len(srcinfo['source']) > 1:
                    for k, url in srcinfo['source'].items():
                        if url.contains('pythonhosted.org'):
                            source_key = k
                else:
                    source_key = 0

                for digest_key, digest_value in release_pkg['digests'].items():
                    if digest_key in digest_to_srcinfo_key:
                        srcinfo_key = digest_to_srcinfo_key[digest_key]
                        if len(srcinfo[srcinfo_key]) <= 1:
                            srcinfo[srcinfo_key] = [digest_value]
                        else:
                            srcinfo[srcinfo_key][source_key] = digest_value

    if has_update:
        # read existing PKGBUILD file into memory
        with open(os.path.join(pkgdir, 'PKGBUILD')) as f:
            contents = f.read()

        # replace pkgver in PKGBUILD
        contents = re.sub(
            r'^pkgver=.*$', 'pkgver={0}'.format(latest_version),
            contents, flags=re.MULTILINE)

        # update integrity checks in PKGBUILD with values from PyPI
        for k in digest_to_srcinfo_key.values():
            if len(srcinfo[k]) == 1:
                repl = "{0}=('{1}')".format(k, srcinfo[k][0])
            elif len(srcinfo[k]) > 1:
                prefix = ' ' * (len(k) + 2)
                repl = "{0}=('{1}'\n".format(k, srcinfo[k][0])
                for i in range(1, len(srcinfo[k])):
                    repl += "{0}'{1}'".format(prefix, srcinfo[k][i])
                    if i == len(srcinfo[k]) - 1:
                        repl += ')'
                    else:
                        repl += '\n'

            repl += '\n'
            contents = re.sub(
                r'^{0}=\([\s\S]*\)\n'.format(k),
                repl, contents, flags=re.MULTILINE)

        format_args = {
            'epoch': '',
            'pkgname': arch_pkgname,
            'pkgver': srcinfo['pkgver'],
            'pkgrel': srcinfo['pkgrel'],
        }

        if srcinfo.get('epoch') is not None:
            format_args['epoch'] = "{0}:".format(srcinfo['epoch'])

        branch_name = "pypi-updates/{pkgname}/{pkgver}".format(**format_args)

        subprocess.run(
            ["git", "checkout", "-B", branch_name],
            cwd=pkgdir,
            check=True)

        # write updated PKGBUILD file
        with open(os.path.join(pkgdir, 'PKGBUILD'), 'w') as f:
            f.write(contents)

        # generate updated .SRCINFO
        with open(os.path.join(pkgdir, '.SRCINFO'), 'w') as f:
            subprocess.run(
                ["makepkg", "--printsrcinfo"],
                cwd=pkgdir,
                stdout=f,
                check=True)

        print('{pkgname} {epoch}{pkgver}-{pkgrel}'.format(**format_args))

        subprocess.run(
            ["git", "commit", "-a", "-n", "-m", """\
upgpkg: {pkgname} {epoch}{pkgver}-{pkgrel}

Upstream release on PyPI (automatic update)
""".format(**format_args)],
            cwd=pkgdir,
            check=True)

        # check if there were any changes made that we need to push
        output = subprocess.check_output(["git", "status", "-uno"], cwd=pkgdir)
        if len(output) > 0:
            subprocess.run(
                ["git", "push", "origin", branch_name],
                cwd=pkgdir,
                check=True)

            # open pull request
            r = requests.post(
                'https://api.github.com/repos/{repo}/pulls'.format(
                    repo=repo_path),
                json={
                    'title': "Update {pkgname} to {pkgver}".format(
                        **format_args),
                    'head': branch_name,
                    'base': "master",
                },
                auth=(github_user, github_token),
                headers={
                    'Accept': "application/vnd.github.v3+json",
                })
            result = r.json()
            print(result['html_url'])

        subprocess.run(
            ["git", "checkout", "master"],
            cwd=pkgdir,
            check=True)
