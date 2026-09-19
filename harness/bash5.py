#!/usr/bin/env python3
"""Find and validate an installed Bash 5+ without changing machine state."""

from __future__ import annotations

import argparse
import os
import re
import shutil
import subprocess
import sys
from collections.abc import Iterable
from pathlib import Path


DEFAULT_PREFIXES = (Path("/opt/homebrew"), Path("/usr/local"))
VERSION_PROBE = 'printf "%s\\n" "${BASH_VERSINFO[0]}"'
PROBE_TIMEOUT_SECONDS = 2


def bash_major(path: str | os.PathLike[str], *, timeout: float = PROBE_TIMEOUT_SECONDS) -> int | None:
    """Return the executable's Bash major version, or None when it is not Bash."""

    candidate = os.fspath(path)
    if not candidate or "\0" in candidate:
        return None
    shell = Path(candidate)
    if not shell.is_file() or not os.access(shell, os.X_OK):
        return None
    try:
        result = subprocess.run(
            [candidate, "--noprofile", "--norc", "-c", VERSION_PROBE],
            check=False,
            capture_output=True,
            env=_probe_environment(),
            text=True,
            timeout=timeout,
        )
    except (OSError, subprocess.SubprocessError):
        return None
    if result.returncode != 0:
        return None
    output = result.stdout.strip()
    if not re.fullmatch(r"\d+", output):
        return None
    return int(output)


def _probe_environment() -> dict[str, str]:
    """Keep shell startup hooks and exported functions out of the probe."""

    environment = dict(os.environ)
    for name in tuple(environment):
        if name in {"BASH_ENV", "ENV", "SHELLOPTS", "BASHOPTS"} or name.startswith("BASH_FUNC_"):
            environment.pop(name, None)
    environment["LC_ALL"] = "C"
    return environment


def is_bash5(path: str | os.PathLike[str]) -> bool:
    """Return whether path executes Bash major version 5 or newer."""

    major = bash_major(path)
    return major is not None and major >= 5


def _brew_bash_prefix(
    brew_executable: str | None,
    *,
    timeout: float = PROBE_TIMEOUT_SECONDS,
) -> Path | None:
    if brew_executable is None:
        brew_executable = shutil.which("brew")
    if not brew_executable:
        return None
    try:
        result = subprocess.run(
            [brew_executable, "--prefix", "bash"],
            check=False,
            capture_output=True,
            env=_probe_environment(),
            text=True,
            timeout=timeout,
        )
    except (OSError, subprocess.SubprocessError):
        return None
    if result.returncode != 0:
        return None
    prefix_text = result.stdout.rstrip("\r\n")
    if not prefix_text or "\n" in prefix_text or "\r" in prefix_text:
        return None
    prefix = Path(prefix_text)
    return prefix if prefix.is_absolute() else None


def discover_bash5(
    *,
    brew_executable: str | None = None,
    prefixes: Iterable[str | os.PathLike[str]] = DEFAULT_PREFIXES,
    path_bash: str | None = None,
) -> str | None:
    """Return the first executable Bash 5+ from Homebrew and standard locations.

    Homebrew's own formula prefix is authoritative when available. The two
    published macOS prefixes and PATH then cover installations without a
    callable Homebrew command. The function never installs or edits anything.
    """

    candidates: list[str] = []
    prefix = _brew_bash_prefix(brew_executable)
    if prefix is not None:
        candidates.append(str(prefix / "bin/bash"))
    candidates.extend(str(Path(item) / "bin/bash") for item in prefixes)
    if path_bash is None:
        path_bash = shutil.which("bash")
    if path_bash:
        candidates.append(path_bash)

    seen: set[str] = set()
    for candidate in candidates:
        candidate = os.path.abspath(candidate)
        if candidate in seen:
            continue
        seen.add(candidate)
        if is_bash5(candidate):
            return candidate
    return None


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("discover", "validate"))
    parser.add_argument("path", nargs="?", help="Bash executable to validate")
    args = parser.parse_args(argv)
    if args.command == "validate":
        if not args.path or not is_bash5(args.path):
            print("ERROR: configured executable is not Bash 5+", file=sys.stderr)
            return 2
        return 0
    shell = discover_bash5()
    if shell is None:
        print("ERROR: no executable Bash 5+ found; install Bash or configure a valid shell value", file=sys.stderr)
        return 2
    print(shell)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
