#!/usr/bin/python3

import glob
import re
import subprocess
import sys

line_re = re.compile(
    r"^PKGBUILD \((?P<pkgname>.+?)\) (?P<severity>[A-Z]): (?P<message>.+)$"
)
num_errors = 0

for pkgbuild in glob.glob("*/PKGBUILD"):
    output = subprocess.check_output(["namcap", pkgbuild])
    for line in output.splitlines():
        m = line_re.match(line.decode("utf-8"))
        if m.group("severity") == "E":
            severity = "error"
            num_errors += 1
        elif m.group("severity") == "W":
            severity = "warning"
        else:
            severity = "debug"

        print(
            "::{severity} file={pkgbuild},line=0,col=0::{message}".format(
                severity=severity, pkgbuild=pkgbuild, message=m.group("message")
            )
        )

if num_errors > 0:
    sys.exit(1)
