#!/usr/bin/python3

import argparse
import os.path
import requests
import sys
import yaml


class GitHubEventDispatcher(object):
    def __init__(self, owner, repo, gh_auth):
        self.owner = owner
        self.repo = repo
        self.gh_auth = gh_auth

    def dispatch(self, event_type, client_payload=None):
        event_data = {
            'event_type': event_type,
        }

        if client_payload is not None:
            event_data['client_payload'] = client_payload

        return requests.post(
            'https://api.github.com/repos/{owner}/{repo}/dispatches'.format(
                owner=self.owner,
                repo=self.repo),
            auth=(self.gh_auth['user'], self.gh_auth['oauth_token']),
            json=event_data,
            headers={
                'Accept': "application/vnd.github.v3+json",
                'User-Agent': "geschenkerbauer-event-dispatch",
            })


parser = argparse.ArgumentParser(
    description="Use GitHub Actions to perform package-related actions")
group = parser.add_mutually_exclusive_group(required=True)
group.add_argument('--build', action='store_true',
                   help="Build the specified package(s)")
group.add_argument('--check-updates', action='store_true',
                   help="Check the AUR for package updates")
parser.add_argument('--nodeps', action='store_true',
                    help="Disable dependency checking when building")
parser.add_argument('pkgname', nargs='*')
args = parser.parse_args()

try:
    import xdg.BaseDirectory
    configpath = xdg.BaseDirectory.load_first_config('hub')
except ImportError:
    configpath = os.path.expanduser('~/.config/hub')

with open(configpath) as f:
    config = yaml.safe_load(f)
    gh_auth = config['github.com'][0]

dispatcher = GitHubEventDispatcher('mutantmonkey', 'aur', gh_auth)

if args.build:
    for pkgname in args.pkgname:
        data = {'pkgname': pkgname}
        if args.nodeps:
            data['nodeps'] = "1"

        r = dispatcher.dispatch('build-package', data)
        if r.status_code == 204:
            print("Build for {0} triggered".format(pkgname), file=sys.stderr)
        else:
            print(
                "Build for {0} returned unknown status code {1}".format(
                    pkgname, r.status_code),
                file=sys.stderr)
elif args.check_updates:
    if len(args.pkgname) > 0:
        print(
            "Limiting the update check by package name is not supported; all "
            "packages will be checked.",
            file=sys.stderr)

    r = dispatcher.dispatch('check-aur-for-updates')
    if r.status_code == 204:
        print("Update check triggered", file=sys.stderr)
    else:
        print(
            "Update check returned unknown status code {0}".format(
                r.status_code),
            file=sys.stderr)
