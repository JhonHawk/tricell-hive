#!/usr/bin/env python3
"""Count pilot projects in the Mac's everyday Engram store without printing names."""
import json
import os
import re
import subprocess
import sys


def parse_projects(output):
    lines = output.strip().splitlines()
    if lines == ["No projects found."]:
        return []
    if not lines or not re.fullmatch(r"Projects \(\d+\):", lines[0]):
        raise ValueError("unrecognized Engram 3.3.2 projects header")
    projects = []
    for line in lines[1:]:
        match = re.fullmatch(r"\s{2}(.+?)\s+\d+ obs\s+\d+ sessions?\s+\d+ prompts?\s*", line)
        if not match:
            raise ValueError("unrecognized Engram 3.3.2 project row")
        projects.append(match.group(1).strip())
    if len(projects) != int(re.search(r"\d+", lines[0]).group()):
        raise ValueError("Engram project count mismatch")
    return projects


def main():
    # Preserve the everyday store routing, but avoid an update check. Never
    # print stdout/stderr: either could include private project/record data.
    env = dict(os.environ, ENGRAM_NO_UPDATE_CHECK="1")
    result = subprocess.run(["engram", "projects", "list"], capture_output=True, text=True, env=env)
    if result.returncode:
        raise RuntimeError("Engram projects list failed; everyday store count is unverified")
    projects = parse_projects(result.stdout)
    exact = [sum(project == requested for project in projects) for requested in sys.argv[1:]]
    prefix = sum(project.startswith("flash-class-") for project in projects)
    print(json.dumps({"total_projects": len(projects), "flash_class_prefix_matches": prefix,
                      "exact_matches_in_argument_order": exact}))
    return 1 if prefix or any(exact) else 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (OSError, RuntimeError, ValueError) as error:
        print(f"error: {error}", file=sys.stderr)
        sys.exit(2)
