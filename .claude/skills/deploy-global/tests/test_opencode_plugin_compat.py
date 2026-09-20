#!/usr/bin/env python3
"""Check the OpenCode session plugin is deployed on V2 and skipped on V1."""

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


ABSENT = None
VERSION_PROBE_FAILS = "__fails__"


def run_plugin_deploy(version: str | None, home: Path) -> subprocess.CompletedProcess[str]:
    """Run step_deploy_opencode_plugin against a fake `opencode` on PATH.

    version=ABSENT drops opencode from PATH entirely; VERSION_PROBE_FAILS
    installs one that exits non-zero, which is what a broken install or a
    dangling shim looks like.
    """
    source_only = SCRIPT.read_text(encoding="utf-8").split(MAIN_MARKER, 1)[0]
    with tempfile.TemporaryDirectory() as temp:
        fake_bin = Path(temp) / "bin"
        fake_bin.mkdir(parents=True, exist_ok=True)
        if version is VERSION_PROBE_FAILS:
            write_executable(fake_bin / "opencode", "#!/bin/sh\nexit 1\n")
        elif version is not ABSENT:
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
            # ABSENT needs a PATH with no real opencode on it; the step only
            # needs the coreutils under /usr/bin and /bin.
            base_path = "/usr/bin:/bin" if version is ABSENT else os.environ.get("PATH", "")
            environment = {
                **os.environ,
                "HOME": str(home),
                "PATH": f"{fake_bin}:{base_path}",
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
    """The plugin targets the V2 API only, so the version gate decides both ways."""

    def test_v2_deploys_the_session_plugin(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            result = run_plugin_deploy("opencode v2.0.9", home)
            self.assertEqual(result.returncode, 0, result.stderr)
            target = home / ".config" / "opencode" / "plugins" / PLUGIN_SOURCE.name
            self.assertEqual(target.read_bytes(), PLUGIN_SOURCE.read_bytes())

    def test_v2_replaces_a_stale_copy(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            target = home / ".config" / "opencode" / "plugins" / PLUGIN_SOURCE.name
            target.parent.mkdir(parents=True)
            target.write_text("stale V1 copy\n", encoding="utf-8")

            result = run_plugin_deploy("opencode v2.0.9", home)

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(target.read_bytes(), PLUGIN_SOURCE.read_bytes())

    def test_v1_skips_the_v2_plugin_and_preserves_an_existing_copy(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            target = home / ".config" / "opencode" / "plugins" / PLUGIN_SOURCE.name
            target.parent.mkdir(parents=True)
            target.write_text("operator-owned legacy copy\n", encoding="utf-8")

            result = run_plugin_deploy("opencode v1.18.31", home)

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("targets the OpenCode 2 plugin API", result.stderr)
            self.assertIn("opencode v1.18.31", result.stderr)
            self.assertEqual(target.read_text(encoding="utf-8"), "operator-owned legacy copy\n")

    def test_an_absent_opencode_still_deploys(self) -> None:
        """The file is inert until opencode exists, and V2 is the current major."""
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            result = run_plugin_deploy(ABSENT, home)
            self.assertEqual(result.returncode, 0, result.stderr)
            target = home / ".config" / "opencode" / "plugins" / PLUGIN_SOURCE.name
            self.assertEqual(target.read_bytes(), PLUGIN_SOURCE.read_bytes())

    def test_an_unreadable_version_skips_without_aborting_the_deploy(self) -> None:
        """A bare assignment here would kill the run before the manifest is written."""
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp)
            result = run_plugin_deploy(VERSION_PROBE_FAILS, home)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("detected opencode is unknown", result.stderr)
            target = home / ".config" / "opencode" / "plugins" / PLUGIN_SOURCE.name
            self.assertFalse(target.exists())

    def test_the_plugin_source_exports_the_v2_default_definition(self) -> None:
        """A named-export regression would deploy a file OpenCode 2 refuses to load."""
        source = PLUGIN_SOURCE.read_text(encoding="utf-8")
        self.assertRegex(source, r"(?m)^export default \w+$")
        self.assertIn('import type { Plugin } from "@opencode/plugin"', source)
        self.assertIn('ctx.session.hook("context"', source)
        # The V1 shape as it would appear in CODE: a named plugin export, or the
        # transform registered as an object key. The header comment names the same
        # hook in prose to explain why V1 is unsupported, so match the punctuation
        # only a registration carries — either quote style.
        self.assertNotRegex(source, r"export const \w+\s*:\s*Plugin\b")
        self.assertNotRegex(source, r"""['"]experimental\.chat\.system\.transform['"]\s*:""")


if __name__ == "__main__":
    unittest.main()
