"""flow-context.sh must emit the hookSpecificOutput JSON envelope.

Claude Code accepts bare stdout for UserPromptSubmit, but the PI adapter
(harness/pi/src/hook-runner.ts) reports bare text as "Hook returned invalid
JSON". The hook therefore emits the same envelope as its sibling hooks.
"""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
import uuid
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[4]
HOOK = REPO_ROOT / "global" / "hooks" / "flow-context" / "flow-context.sh"


def run_hook(cwd: Path, session_id: str) -> str:
    payload = json.dumps({"harness": "pi", "cwd": str(cwd), "session_id": session_id})
    result = subprocess.run(
        ["bash", str(HOOK)], input=payload, capture_output=True, text=True, check=False
    )
    assert result.returncode == 0, result.stderr
    return result.stdout


class FlowContextHookTests(unittest.TestCase):
    def test_flow_workspace_emits_user_prompt_submit_envelope(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            session_dir = root / "_support" / "sessions" / "2026-09-11-x"
            session_dir.mkdir(parents=True)
            (root / "_support" / "PROJECT.md").write_text("| Current stage | build |\n")
            (session_dir / "x-plan.md").write_text("Status: building\nSession: yes\n")

            stdout = run_hook(root, f"envelope-{uuid.uuid4().hex}")

        envelope = json.loads(stdout)
        specific = envelope["hookSpecificOutput"]
        self.assertEqual(specific["hookEventName"], "UserPromptSubmit")
        context = specific["additionalContext"]
        self.assertIn("x-plan.md (Status: building)", context)
        self.assertIn("portable planning conventions", context)

    def test_non_flow_workspace_stays_silent(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            stdout = run_hook(Path(tmp), f"silent-{uuid.uuid4().hex}")
        self.assertEqual(stdout, "")


if __name__ == "__main__":
    unittest.main()
