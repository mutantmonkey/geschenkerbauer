#!/usr/bin/python3

import datetime
import dateutil.parser
import os.path
import requests
import subprocess
import tempfile
import yaml

allowed_workflows = set(["build_package.yml", "build_updated_packages.yml"])
headers = {
    "Accept": "application/vnd.github.v3+json",
    "User-Agent": "geschenkerbauer-artifact-fetch",
}

try:
    with open(os.path.expanduser("~/.config/hub")) as f:
        config = yaml.safe_load(f)
except IOError:
    config = None

if config is not None and "github.com" in config:
    auth_tuple = (
        config["github.com"][0]["user"],
        config["github.com"][0]["oauth_token"],
    )
else:
    auth_tuple = None


def get_workflows(only_workflow: str = None):
    r = requests.get(
        "https://api.github.com/repos/mutantmonkey/aur/actions/workflows",
        headers=headers,
        auth=auth_tuple,
    )
    r.raise_for_status()

    for workflow in r.json()["workflows"]:
        workflow_base = os.path.basename(workflow["path"])
        if only_workflow is not None:
            if workflow_base == only_workflow:
                yield workflow
        elif workflow_base in allowed_workflows:
            yield workflow


def get_runs_for_workflow(
    workflow: dict, min_run_number: int = None, time_window: datetime.timedelta = None
):
    runs_url = workflow["url"] + "/runs"

    r = requests.get(
        runs_url,
        headers=headers,
        auth=auth_tuple,
    )
    r.raise_for_status()
    runs = r.json()["workflow_runs"]

    if time_window is not None:
        min_created_at = datetime.datetime.utcnow().replace(tzinfo=None) - time_window
    else:
        min_created_at = None

    for run in runs:
        if min_run_number is not None and run["run_number"] < min_run_number:
            break

        if min_created_at is not None:
            created_at = dateutil.parser.parse(run["created_at"], ignoretz=True)
            if created_at < min_created_at:
                break

        yield run


def get_artifacts_for_run(run: dict):
    r = requests.get(
        run["artifacts_url"],
        headers=headers,
        auth=auth_tuple,
    )
    r.raise_for_status()
    for artifact in r.json()["artifacts"]:
        yield artifact


def download_and_unpack_artifact(artifact):
    output_fd, output_filename = tempfile.mkstemp(".zip")
    with open(output_fd, "wb") as f:
        r = requests.get(
            artifact["archive_download_url"],
            headers=headers,
            auth=auth_tuple,
            stream=True,
        )
        for chunk in r.iter_content(8192):
            f.write(chunk)

    subprocess.call(["./unpack_and_sign.sh", output_filename])
    os.remove(output_filename)


if __name__ == "__main__":
    import argparse
    from colorama import Fore, Back, Style, init

    parser = argparse.ArgumentParser()
    group = parser.add_mutually_exclusive_group()
    group.add_argument("--in-last-hours", type=int, default=4)
    group.add_argument("--min-run-number", type=int)
    parser.add_argument(
        "--workflow",
        type=str,
        help="Workflow basename (e.g. build_package.yml). Required for --min-run-number.",
    )
    args = parser.parse_args()

    only_workflow = args.workflow
    run_filter = {}

    if args.min_run_number is not None:
        run_filter["min_run_number"] = args.min_run_number
        if only_workflow is None:
            raise Exception("--workflow must be specified when using --min-run-number.")

    if args.min_run_number is None and args.in_last_hours is not None:
        run_filter["time_window"] = datetime.timedelta(hours=args.in_last_hours)

    for workflow in get_workflows(only_workflow):
        for run in get_runs_for_workflow(workflow, **run_filter):
            print(
                f"{Style.BRIGHT}{Fore.GREEN}==>{Fore.WHITE} Processing workflow run: {run['name']} #{run['run_number']}{Style.RESET_ALL}"
            )
            for artifact in get_artifacts_for_run(run):
                print(
                    f"{Style.BRIGHT}{Fore.GREEN}==>{Fore.WHITE} Processing artifact: '{artifact['name']}'...{Style.RESET_ALL}"
                )
                download_and_unpack_artifact(artifact)
