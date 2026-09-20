#!/usr/bin/env python3
"""Verify the deploy wrapper accepts only the tested Pi 0.86 patch line."""

from __future__ import annotations

import os
import subprocess
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[4]
SCRIPT = REPO_ROOT / ".claude" / "skills" / "deploy-global" / "scripts" / "deploy-global.sh"
MAIN_MARKER = "# ---------------------------------------------------------------------------\n# Main\n"


def write_executable(path: Path, content: str) -> Path:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")
    path.chmod(0o755)
    return path


def run_pi_dependency_check(version: str) -> subprocess.CompletedProcess[str]:
    source_only = SCRIPT.read_text(encoding="utf-8").split(MAIN_MARKER, 1)[0]
    with tempfile.TemporaryDirectory() as temp:
        root = Path(temp)
        fake_bin = root / "bin"
        write_executable(fake_bin / "pi", f"#!/bin/sh\nprintf '%s\\n' '{version}'\n")
        with tempfile.NamedTemporaryFile(
            mode="w",
            encoding="utf-8",
            dir=SCRIPT.parent,
            prefix=".deploy-global-test-",
            suffix=".sh",
            delete=False,
        ) as source_file:
            source_file.write(source_only)
            source_path = Path(source_file.name)
        try:
            environment = {
                **os.environ,
                "PATH": f"{fake_bin}:{os.environ.get('PATH', '')}",
            }
            return subprocess.run(
                ["/bin/bash", "-c", 'source "$1"; check_pi_deps', "pi-version-test", str(source_path)],
                cwd=REPO_ROOT,
                env=environment,
                check=False,
                capture_output=True,
                text=True,
            )
        finally:
            source_path.unlink(missing_ok=True)


class PiVersionGateTests(unittest.TestCase):
    def test_accepts_pi_086_patch_releases(self) -> None:
        for version in ("0.86.0", "0.86.7"):
            with self.subTest(version=version):
                result = run_pi_dependency_check(version)
                self.assertEqual(result.returncode, 0, result.stderr)

    def test_rejects_other_minor_lines_and_prereleases(self) -> None:
        for version in ("0.85.1", "0.86.0-beta.1", "0.87.0", "1.0.0"):
            with self.subTest(version=version):
                result = run_pi_dependency_check(version)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("Unsupported PI version", result.stderr)


if __name__ == "__main__":
    unittest.main()
