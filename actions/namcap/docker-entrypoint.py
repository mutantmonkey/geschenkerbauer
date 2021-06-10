#!/usr/bin/python3

import glob
import re
import subprocess

line_re = re.compile(
    r"^PKGBUILD \((?P<pkgname>.+?)\) (?P<severity>[A-Z]): (?P<message>.+)$"
)

for pkgbuild in glob.glob("*/PKGBUILD"):
    output = subprocess.check_output(["namcap", pkgbuild])
    for line in output.splitlines():
        m = line_re.match(line.decode("utf-8"))
        if m.group("severity") == "E":
            severity = "error"
        elif m.group("severity") == "W":
            severity = "warning"
        else:
            severity = "debug"

        print(
            "::{severity} file={pkgbuild}::{message}".format(
                severity=severity, pkgbuild=pkgbuild, message=m.group("message")
            )
        )
