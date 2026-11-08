from pathlib import Path
import subprocess
import sys
import tempfile
import textwrap
import unittest


ROOT = Path(__file__).resolve().parents[2]
TASKS = ROOT / "content/skills/flow-plan/scripts/tasks.py"
sys.dont_write_bytecode = True  # keep __pycache__ out of the distributed skill folder

TASKS_MD = textwrap.dedent(
    """\
    # Tasks

    ### T1 — First

    - [x] one
    - [x] two

    **Closes:** AC1

    ### T2 — Second thing

    - [ ] alpha
    - [ ] beta

    **Closes:** AC2

    ### T3 — Third

    - [ ] gamma

    ### T4 — Fourth

    - [ ] delta

    ## Review status and progress

    2026-11-08 — Plan written.
    """
)

PROPOSAL = "# demo\n\n- AC1 — one · false today\n- AC2 — two · false today\n"


class TasksScriptTest(unittest.TestCase):
    def run_tasks(self, folder, *args):
        return subprocess.run(
            [sys.executable, "-I", "-B", str(TASKS), str(folder), *args],
            text=True,
            capture_output=True,
        )

    def folder(self, tmp, tasks=TASKS_MD):
        change = Path(tmp) / "demo"
        change.mkdir()
        (change / "tasks.md").write_text(tasks, encoding="utf-8")
        (change / "proposal.md").write_text(PROPOSAL, encoding="utf-8")
        return change

    def text(self, change):
        return (change / "tasks.md").read_text(encoding="utf-8")

    def test_done_ticks_only_that_task_and_records_one_line(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            result = self.run_tasks(change, "done", "T2", "--commit", "abc1234", "--unverified", "UI walk")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(result.stdout.strip(), "✓ T2 · 2/4 · Second thing")
            text = self.text(change)
            self.assertIn("- [x] alpha\n- [x] beta", text)
            self.assertIn("- [ ] gamma", text)
            self.assertIn("- [ ] delta", text)
            lines = [l for l in text.splitlines() if "T2 · abc1234" in l]
            self.assertEqual(len(lines), 1)
            self.assertRegex(lines[0], r"^\d{4}-\d{2}-\d{2} — T2 · abc1234 · unverified: UI walk$")
            self.assertLess(text.index("Plan written."), text.index("T2 · abc1234"))

    def test_done_twice_does_not_duplicate_the_line(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            args = ("done", "T2", "--commit", "abc1234", "--unverified", "none")
            self.run_tasks(change, *args)
            first = self.text(change)
            result = self.run_tasks(change, *args)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(self.text(change), first)
            self.assertEqual(first.count("T2 · abc1234"), 1)

    def test_done_exits_2_for_a_missing_task_or_progress_section(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            result = self.run_tasks(change, "done", "T9", "--commit", "abc1234", "--unverified", "none")
            self.assertEqual(result.returncode, 2)
            self.assertEqual(self.text(change), TASKS_MD)
        with tempfile.TemporaryDirectory() as tmp:
            bare = TASKS_MD.split("## Review status")[0]
            change = self.folder(tmp, tasks=bare)
            result = self.run_tasks(change, "done", "T2", "--commit", "abc1234", "--unverified", "none")
            self.assertEqual(result.returncode, 2)
            self.assertEqual(self.text(change), bare)

    def test_done_rejects_a_missing_unverified_flag(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            result = self.run_tasks(change, "done", "T2", "--commit", "abc1234")
            self.assertEqual(result.returncode, 2)
            self.assertEqual(self.text(change), TASKS_MD)

    def test_status_counts_tasks(self):
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_tasks(self.folder(tmp), "status")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(
                result.stdout.splitlines(),
                ["T1 [x] 2/2 · First", "T2 [ ] 0/2 · Second thing", "T3 [ ] 0/1 · Third", "T4 [ ] 0/1 · Fourth", "1/4 tasks"],
            )

    def checkpoint(self, change, stage, step, next_action, *extra):
        return self.run_tasks(
            change, "checkpoint", "--stage", stage, "--step", step, "--commit", "abc1234", "--next", next_action, *extra
        )

    def test_checkpoint_creates_replaces_and_keeps_evidence(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            first = self.checkpoint(change, "build", "T2", "run T3", "--check", "go test ./... · pass · abc1234")
            self.assertEqual(first.returncode, 0, first.stderr)
            text = self.text(change)
            self.assertLess(text.index("## Checkpoint"), text.index("## Review status and progress"))
            for fixed in ("- Stage: build · T2", "- Last commit: abc1234", "- Open findings: none", "- Next: run T3", "### Evidence"):
                self.assertIn(fixed, text)
            self.assertIn("- go test ./... · pass · abc1234", text)

            second = self.checkpoint(
                change, "verify", "round 1", "fix H1", "--finding", "H1: replaced D4", "--check", "lint · pass · def5678"
            )
            self.assertEqual(second.returncode, 0, second.stderr)
            text = self.text(change)
            self.assertEqual(text.count("## Checkpoint"), 1)
            self.assertEqual(text.count("- Stage:"), 1)
            self.assertIn("- Stage: verify · round 1", text)
            self.assertNotIn("run T3", text)
            self.assertIn("- Open findings: H1: replaced D4", text)
            self.assertIn("- go test ./... · pass · abc1234", text)
            self.assertIn("- lint · pass · def5678", text)
            self.assertIn("Plan written.", text)

    def test_check_change_exits_0_when_the_checkpoint_names_a_replaced_decision(self):
        proposal = textwrap.dedent(
            """\
            # demo

            - **D4 A** — Old way.
            - **D12 A** — New way. **Replaces:** D4-A
            - AC1 — one · false today
            - AC2 — two · false today
            """
        )
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            (change / "proposal.md").write_text(proposal, encoding="utf-8")
            self.checkpoint(change, "build", "T2", "run T3", "--finding", "H1: keep D4-A")
            self.assertIn("D4-A", self.text(change))
            check = ROOT / "content/skills/flow-plan/scripts/check_change.py"
            result = subprocess.run(
                [sys.executable, "-I", "-B", str(check), str(change)], text=True, capture_output=True
            )
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)


if __name__ == "__main__":
    unittest.main()
