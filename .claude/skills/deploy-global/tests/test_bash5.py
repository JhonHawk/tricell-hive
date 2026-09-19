#!/usr/bin/env python3
"""Bash 5 discovery and its OpenCode deploy integration."""

from __future__ import annotations

import json
import os
import subprocess
import tempfile
import unittest
from pathlib import Path
from importlib.util import module_from_spec, spec_from_file_location
from unittest import mock


REPO_ROOT = Path(__file__).resolve().parents[4]
SCRIPT = REPO_ROOT / ".claude/skills/deploy-global/scripts/deploy-global.sh"
BASH5_MODULE = REPO_ROOT / "harness/bash5.py"


def load_bash5_module():
    spec = spec_from_file_location("hive_bash5_test", BASH5_MODULE)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"Unable to load {BASH5_MODULE}")
    module = module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def write_executable(path: Path, content: str) -> Path:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")
    path.chmod(0o755)
    return path


def make_fake_bash(path: Path, major: str) -> Path:
    return write_executable(
        path,
        "#!/bin/sh\n"
        'if [ "$1" != "--noprofile" ] || [ "$2" != "--norc" ] || [ "$3" != "-c" ]; then exit 2; fi\n'
        'if ! printf "%s" "$4" | grep -q BASH_VERSINFO; then exit 2; fi\n'
        f"printf '%s\\n' '{major}'\n",
    )


def source_only_script() -> Path:
    source_only = SCRIPT.read_text(encoding="utf-8").split(
        "# ---------------------------------------------------------------------------\n# Main\n", 1
    )[0]
    with tempfile.NamedTemporaryFile(
        mode="w", encoding="utf-8", dir=SCRIPT.parent,
        prefix=".deploy-global-test-", suffix=".sh", delete=False,
    ) as source_file:
        source_file.write(source_only)
        return Path(source_file.name)


class BashDiscoveryTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)

    def test_discovers_bash_from_brew_prefix_with_spaces(self) -> None:
        bash5 = load_bash5_module()
        prefix = self.root / "Homebrew with spaces"
        candidate = make_fake_bash(prefix / "bin/bash", "5")
        brew = write_executable(
            self.root / "brew with spaces",
            "#!/bin/sh\n"
            '[ "$1" = "--prefix" ] && [ "$2" = "bash" ] || exit 2\n'
            f"printf '%s\\n' '{prefix}'\n",
        )

        self.assertEqual(
            bash5.discover_bash5(
                brew_executable=str(brew), prefixes=(), path_bash=None
            ),
            str(candidate),
        )

    def test_standard_prefixes_cover_apple_silicon_and_intel(self) -> None:
        bash5 = load_bash5_module()
        for prefix in (self.root / "opt/homebrew", self.root / "usr/local"):
            with self.subTest(prefix=prefix):
                candidate = make_fake_bash(prefix / "bin/bash", "5")
                self.assertEqual(
                    bash5.discover_bash5(
                        brew_executable="", prefixes=(prefix,), path_bash=None
                    ),
                    str(candidate),
                )

    def test_rejects_bash_three_and_non_executable_candidates(self) -> None:
        bash5 = load_bash5_module()
        old_bash = make_fake_bash(self.root / "old/bin/bash", "3")
        not_executable = self.root / "not executable/bin/bash"
        not_executable.parent.mkdir(parents=True)
        not_executable.write_text("#!/bin/sh\nexit 0\n", encoding="utf-8")

        self.assertFalse(bash5.is_bash5(str(old_bash)))
        self.assertFalse(bash5.is_bash5(str(not_executable)))
        self.assertIsNone(
            bash5.discover_bash5(
                brew_executable="", prefixes=(self.root / "old",), path_bash=str(not_executable)
            )
        )

    def test_version_probe_does_not_run_bash_env(self) -> None:
        bash5 = load_bash5_module()
        marker = self.root / "startup was not run"
        bash_env = self.root / "user startup file"
        bash_env.write_text(f"printf 'ran' > '{marker}'\n", encoding="utf-8")

        with mock.patch.dict(os.environ, {"BASH_ENV": str(bash_env)}):
            self.assertIsNotNone(bash5.bash_major("/bin/bash"))

        self.assertFalse(marker.exists())

    def test_opencode_adds_discovered_bash_and_preserves_user_shell(self) -> None:
        home = self.root / "home"
        config_dir = home / ".config/opencode"
        config_dir.mkdir(parents=True)
        prefix = self.root / "Homebrew with spaces"
        candidate = make_fake_bash(prefix / "bin/bash", "5")
        brew = write_executable(
            self.root / "bin/brew",
            "#!/bin/sh\n"
            '[ "$1" = "--prefix" ] && [ "$2" = "bash" ] || exit 2\n'
            f"printf '%s\\n' '{prefix}'\n",
        )
        (config_dir / "opencode.json").write_text(
            json.dumps({"permission": {"external_directory": {}}}) + "\n",
            encoding="utf-8",
        )
        source = source_only_script()
        report = self.root / "report.log"
        report.touch()
        try:
            result = subprocess.run(
                [
                    "/bin/bash", "-c",
                    'source "$1"; APPLY=1; RUN_OPENCODE=1; '
                    'REPORT_LOG="$2"; step_merge_opencode_permissions',
                    "bash5-opencode-test", str(source), str(report),
                ],
                cwd=REPO_ROOT,
                env={**os.environ, "HOME": str(home), "PATH": f"{brew.parent}:{os.environ['PATH']}"},
                check=False,
                capture_output=True,
                text=True,
            )
        finally:
            source.unlink(missing_ok=True)

        self.assertEqual(result.returncode, 0, result.stderr)
        merged = json.loads((config_dir / "opencode.json").read_text(encoding="utf-8"))
        self.assertEqual(merged["shell"], str(candidate))

        custom = make_fake_bash(self.root / "custom shell/bin/bash", "5")
        (config_dir / "opencode.json").write_text(
            json.dumps({"permission": {}, "shell": str(custom)}) + "\n",
            encoding="utf-8",
        )
        source = source_only_script()
        try:
            result = subprocess.run(
                [
                    "/bin/bash", "-c",
                    'source "$1"; APPLY=1; RUN_OPENCODE=1; '
                    'REPORT_LOG="$2"; step_merge_opencode_permissions',
                    "bash5-opencode-test", str(source), str(report),
                ],
                cwd=REPO_ROOT,
                env={**os.environ, "HOME": str(home), "PATH": f"{brew.parent}:{os.environ['PATH']}"},
                check=False,
                capture_output=True,
                text=True,
            )
        finally:
            source.unlink(missing_ok=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        merged = json.loads((config_dir / "opencode.json").read_text(encoding="utf-8"))
        self.assertEqual(merged["shell"], str(custom))


if __name__ == "__main__":
    unittest.main()
