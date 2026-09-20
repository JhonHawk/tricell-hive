#!/usr/bin/env python3
"""Check OpenCode V1 plugin deployment is skipped on V2."""

from __future__ import annotations

import os
import subprocess
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[4]
SCRIPT = REPO_ROOT / ".claude" / "skills" / "deploy-global" / "scripts" / "deploy-global.sh"
PLUGIN_SOURCE = REPO_ROOT / "global" / "hooks" / "flow-session-context" / "flow-session-context.ts"
MAIN_MARKER = "# ---------------------------------------------------------------------------\n# Main\n"


def write_executable(path: Path, content: str) -> Path:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")
    path.chmod(0o755)
    return path


def run_plugin_deploy(version: str, home: Path) -> subprocess.CompletedProcess[str]:
    source_only = SCRIPT.read_text(encoding="utf-8").split(MAIN_MARKER, 1)[0]
    with tempfile.TemporaryDirectory() as temp:
        fake_bin = Path(temp) / "bin"
        write_executable(fake_bin / "opencode", f"#!/bin/sh\nprintf '%s\\n' '{version}'\n")
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
                "HOME": str(home),
                "PATH": f"{fake_bin}:{os.environ.get('PATH', '')}",
            }
            return subprocess.run(
                [
                    "/bin/bash",
                    "-c",
                    'source "$1"; RUN_OPENCODE=1; APPLY=1; step_deploy_opencode_plugin',
                    "opencode-plugin-test",
                    str(source_path),
                ],
                cwd=REPO_ROOT,
                env=environment,
                check=False,
                capture_output=True,
                text=True,
            )
        finally:
            source_path.unlink(missing_ok=True)


class OpenCodePluginCompatibilityTests(unittest.TestCase):
    def test_v1_still_deploys_the_session_plugin(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            result = run_plugin_deploy("opencode v1.18.18", home)
            self.assertEqual(result.returncode, 0, result.stderr)
            target = home / ".config" / "opencode" / "plugins" / PLUGIN_SOURCE.name
            self.assertEqual(target.read_bytes(), PLUGIN_SOURCE.read_bytes())

    def test_v2_skips_the_v1_plugin_and_preserves_an_existing_copy(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            target = home / ".config" / "opencode" / "plugins" / PLUGIN_SOURCE.name
            target.parent.mkdir(parents=True)
            target.write_text("operator-owned legacy copy\n", encoding="utf-8")

            result = run_plugin_deploy("opencode v2.0.9", home)

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("OpenCode 2 does not load V1 plugins", result.stderr)
            self.assertEqual(target.read_text(encoding="utf-8"), "operator-owned legacy copy\n")


if __name__ == "__main__":
    unittest.main()
