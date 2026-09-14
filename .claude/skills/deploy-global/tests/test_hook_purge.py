#!/usr/bin/env python3
"""Regression tests for surgical removal of retired Hive hook entries."""

from __future__ import annotations

import json
import shutil
import subprocess
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[4]
FILTER = REPO_ROOT / ".claude/skills/deploy-global/filters/hook-purge.jq"


def purge(payload: dict[str, object], basename: str, kind: str) -> dict[str, object]:
    jq = shutil.which("jq")
    if jq is None:
        raise unittest.SkipTest("jq is required by deploy-global")
    result = subprocess.run(
        [jq, "--arg", "bn", basename, "--arg", "kind", kind, "-f", str(FILTER)],
        input=json.dumps(payload),
        text=True,
        capture_output=True,
        check=True,
    )
    return json.loads(result.stdout)


class HookPurgeTests(unittest.TestCase):
    def test_claude_removes_only_exact_command_and_keeps_co_located_hooks(self) -> None:
        payload = {
            "hooks": {
                "PostToolUse": [
                    {
                        "matcher": "*",
                        "hooks": [
                            {"type": "command", "command": "$HOME/.claude/hooks/flow-plan-capture.sh"},
                            {"type": "command", "command": "$HOME/.claude/hooks/user-owned.sh"},
                        ],
                    },
                    {
                        "hooks": [
                            {"type": "command", "command": "/tmp/flow-plan-capture.sh --operator"},
                        ],
                    },
                ],
            },
        }

        result = purge(payload, "flow-plan-capture.sh", "claude")

        entries = result["hooks"]["PostToolUse"]
        self.assertEqual(len(entries), 2)
        self.assertEqual(entries[0]["hooks"][0]["command"], "$HOME/.claude/hooks/user-owned.sh")
        self.assertEqual(entries[1]["hooks"][0]["command"], "/tmp/flow-plan-capture.sh --operator")

    def test_codex_removes_known_hive_command_without_touching_same_path_with_other_args(self) -> None:
        payload = {
            "hooks": {
                "UserPromptSubmit": [
                    {
                        "hooks": [
                            {
                                "type": "command",
                                "command": '"$HOME/.codex/hooks/flow-plan-capture.sh" --from-codex-prompt',
                            },
                            {
                                "type": "command",
                                "command": '"$HOME/.codex/hooks/flow-plan-capture.sh" --operator',
                            },
                        ],
                    },
                ],
            },
        }

        result = purge(payload, "flow-plan-capture.sh", "codex")

        hooks = result["hooks"]["UserPromptSubmit"][0]["hooks"]
        self.assertEqual([entry["command"] for entry in hooks], ['"$HOME/.codex/hooks/flow-plan-capture.sh" --operator'])

    def test_empty_event_is_removed_after_its_hive_entry_is_removed(self) -> None:
        payload = {
            "hooks": {
                "PostToolUse": [
                    {
                        "hooks": [
                            {"type": "command", "command": "$HOME/.claude/hooks/flow-plan-capture.sh"},
                        ],
                    },
                ],
            },
        }

        result = purge(payload, "flow-plan-capture.sh", "claude")

        self.assertNotIn("PostToolUse", result["hooks"])


if __name__ == "__main__":
    unittest.main()
