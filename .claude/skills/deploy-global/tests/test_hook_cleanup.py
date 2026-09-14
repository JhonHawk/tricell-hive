#!/usr/bin/env python3
"""Regression tests for retiring the native plan-capture hook."""

from __future__ import annotations

import json
import os
import subprocess
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[4]
SCRIPT = REPO_ROOT / ".claude/skills/deploy-global/scripts/deploy-global.sh"
RETIRED_PATH = "global/hooks/flow-plan-capture/flow-plan-capture.sh"
HOOK_COMMAND = "$HOME/.claude/hooks/flow-plan-capture.sh"


def retired_bytes() -> bytes:
    """Read the retired bytes from repository history for the ownership test."""

    revisions = subprocess.run(
        ["git", "log", "--all", "--diff-filter=AM", "--format=%H", "--", RETIRED_PATH],
        cwd=REPO_ROOT,
        check=True,
        capture_output=True,
        text=True,
    ).stdout.splitlines()
    if not revisions:
        raise AssertionError("the retired hook has no reachable add/modify revision")
    result = subprocess.run(
        ["git", "show", f"{revisions[0]}:{RETIRED_PATH}"],
        cwd=REPO_ROOT,
        check=True,
        capture_output=True,
    )
    return result.stdout


def run_cleanup(home: Path, target_bytes: bytes) -> subprocess.CompletedProcess[str]:
    hooks = home / ".claude/hooks"
    hooks.mkdir(parents=True)
    (hooks / "flow-plan-capture.sh").write_bytes(target_bytes)
    settings = {
        "hooks": {
            "PostToolUse": [
                {
                    "matcher": "*",
                    "hooks": [
                        {"type": "command", "command": HOOK_COMMAND},
                        {"type": "command", "command": "$HOME/.claude/hooks/user-owned.sh"},
                        {"type": "command", "command": "/tmp/flow-plan-capture.sh --operator"},
                    ],
                }
            ]
        }
    }
    (home / ".claude/settings.json").write_text(json.dumps(settings), encoding="utf-8")
    manifest = home / ".claude/.deploy-manifest"
    manifest.write_text("hooks/flow-plan-capture.sh\n", encoding="utf-8")

    # Source the real script without invoking main. Keeping the temporary source
    # beside the script preserves its repository-root calculation and filter paths.
    source_only = SCRIPT.read_text(encoding="utf-8").split(
        "# ---------------------------------------------------------------------------\n# Main\n", 1
    )[0]
    with tempfile.NamedTemporaryFile(
        mode="w", encoding="utf-8", dir=SCRIPT.parent, prefix=".deploy-global-test-", suffix=".sh", delete=False
    ) as source_file:
        source_file.write(source_only)
        source_path = Path(source_file.name)

    report = home / "report.log"
    report.touch()
    try:
        command = (
            'source "$1"; '
            'APPLY=1; REPORT_LOG="$2"; '
            'ORPHANS=("hooks/flow-plan-capture.sh"); '
            'step_delete_orphans'
        )
        return subprocess.run(
            ["/bin/bash", "-c", command, "cleanup-test", str(source_path), str(report)],
            cwd=REPO_ROOT,
            env={**os.environ, "HOME": str(home)},
            check=False,
            capture_output=True,
            text=True,
        )
    finally:
        source_path.unlink(missing_ok=True)


class HookCleanupTests(unittest.TestCase):
    def test_edited_retired_hook_is_preserved_but_legacy_registration_is_removed(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp) / "home"
            home.mkdir()
            result = run_cleanup(home, b"#!/bin/sh\n# operator edit\n")

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(
                (home / ".claude/hooks/flow-plan-capture.sh").read_bytes(),
                b"#!/bin/sh\n# operator edit\n",
            )
            settings = json.loads((home / ".claude/settings.json").read_text(encoding="utf-8"))
            commands = [
                hook["command"]
                for entry in settings["hooks"]["PostToolUse"]
                for hook in entry["hooks"]
            ]
            self.assertNotIn(HOOK_COMMAND, commands)
            self.assertIn("$HOME/.claude/hooks/user-owned.sh", commands)
            self.assertIn("/tmp/flow-plan-capture.sh --operator", commands)
            self.assertIn("preserved edited retired hook", (home / "report.log").read_text())

    def test_unmodified_historical_hook_is_deleted_with_only_exact_registration(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            home = Path(temp) / "home"
            home.mkdir()
            result = run_cleanup(home, retired_bytes())

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertFalse((home / ".claude/hooks/flow-plan-capture.sh").exists())
            settings = json.loads((home / ".claude/settings.json").read_text(encoding="utf-8"))
            commands = [
                hook["command"]
                for entry in settings["hooks"]["PostToolUse"]
                for hook in entry["hooks"]
            ]
            self.assertNotIn(HOOK_COMMAND, commands)
            self.assertIn("$HOME/.claude/hooks/user-owned.sh", commands)
            self.assertIn("/tmp/flow-plan-capture.sh --operator", commands)
            self.assertIn("orphans: 1 deleted", (home / "report.log").read_text())


if __name__ == "__main__":
    unittest.main()
