from pathlib import Path
import subprocess
import sys
import tempfile
import textwrap
import unittest


ROOT = Path(__file__).resolve().parents[2]
CHECK = ROOT / "content/skills/flow-plan/scripts/check_change.py"
sys.dont_write_bytecode = True  # keep __pycache__ out of the distributed skill folder


def write(folder, name, text):
    path = folder / name
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(textwrap.dedent(text).lstrip())


PROPOSAL = """
    # demo — change

    ## Decisions

    - **D4 A** — Commit per task point? Yes, one joint commit.
    - **D12 A** — Commit each task when reported. **Replaces:** D4-A

    ## Scope and acceptance

    - AC1 — The export button appears · false today
    - AC2 — Retired criterion **Retired:** already true on the base
    - AC3 — The export runs · false today
    - AC4 — Pagination works · false today

    ## Estado de la revisión y avance

    2026-10-25 — D4-A applied before the revision.
    """

TASKS = """
    # Tasks

    ### T1 — Export

    - [x] Export button shown
    - [ ] Export file written

    **Closes:** AC1, AC3–AC4

    **Changes:** commit by D12-A, which replaced D4-A.

    ## Review status and progress

    2026-10-26 — T1 done; D4-A no longer applies.
    """


class ChangeCheckTest(unittest.TestCase):
    def run_check(self, folder, *args):
        return subprocess.run(
            [sys.executable, "-B", str(CHECK), str(folder), *args],
            text=True,
            capture_output=True,
        )

    def folder(self, tmp, proposal=PROPOSAL, tasks=TASKS):
        change = Path(tmp) / "demo"
        write(change, "proposal.md", proposal)
        write(change, "tasks.md", tasks)
        return change

    def test_coherent_folder_passes(self):
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp))
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertEqual(result.stdout, "")

    def test_reports_a_criterion_no_task_closes(self):
        tasks = TASKS.replace("AC1, AC3–AC4", "AC1, AC3")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, tasks=tasks))
            self.assertEqual(result.returncode, 1)
            self.assertIn("proposal.md:13: AC4", result.stdout)
            self.assertNotIn("AC2", result.stdout)

    def test_reports_a_closes_line_naming_an_undefined_criterion(self):
        tasks = TASKS.replace("AC1, AC3–AC4", "AC1, AC3–AC4, AC9")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, tasks=tasks))
            self.assertEqual(result.returncode, 1)
            self.assertIn("tasks.md:8: AC9", result.stdout)

    def test_reports_each_stale_mention_of_a_replaced_or_retired_id(self):
        tasks = TASKS.replace(
            "**Changes:** commit by D12-A, which replaced D4-A.",
            "**Changes:** commit by D4-A and close AC2.",
        )
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp, tasks=tasks)
            write(change, "design.md", "# Design\n\nCommits follow D4 A, not D40.\n")
            write(change, "specs/export/spec.md", "### Requirement: Export\n\nJoint commit per D4-A.\n")
            result = self.run_check(change)
            self.assertEqual(result.returncode, 1)
            lines = result.stdout.splitlines()
            self.assertIn("tasks.md:10: D4-A", result.stdout)
            self.assertIn("tasks.md:10: AC2", result.stdout)
            self.assertIn("design.md:3: D4-A", result.stdout)
            self.assertIn("specs/export/spec.md:3: D4-A", result.stdout)
            self.assertEqual(len(lines), 4, result.stdout)

    def test_archive_lists_each_unchecked_box(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            plain = self.run_check(change)
            archive = self.run_check(change, "--archive")
            self.assertEqual(plain.returncode, 0)
            self.assertEqual(archive.returncode, 1)
            self.assertIn("tasks.md:6: unchecked box: Export file written", archive.stdout)
            self.assertIn("1 unchecked box", archive.stdout.splitlines()[-1])

    def test_a_changed_option_letter_is_a_replacement(self):
        proposal = PROPOSAL.replace("- **D12 A** — Commit each task when reported. **Replaces:** D4-A", "- **D4 B** — Commit each task when reported. **Replaces:** D4-A")
        tasks = TASKS.replace("**Changes:** commit by D12-A, which replaced D4-A.", "**Changes:** commit per D4-A; D4-B is current.")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, proposal=proposal, tasks=tasks))
            self.assertEqual(result.returncode, 0, result.stdout)
            tasks = TASKS.replace("**Changes:** commit by D12-A, which replaced D4-A.", "**Changes:** commit per D4-A.")
            result = self.run_check(self.folder(tmp, proposal=proposal, tasks=tasks))
            self.assertIn("tasks.md:10: D4-A", result.stdout)

    def test_only_the_proposal_definition_keeps_an_old_id(self):
        tasks = TASKS.replace("**Changes:** commit by D12-A, which replaced D4-A.", "- AC2 — check the retired thing still")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, tasks=tasks))
            self.assertIn("tasks.md:10: AC2", result.stdout)

    def test_prose_after_the_replaced_ids_is_not_replaced(self):
        proposal = PROPOSAL.replace("**Replaces:** D4-A", "**Replaces:** D4-A (asked again after T1)")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, proposal=proposal))
            self.assertEqual(result.returncode, 0, result.stdout)

    def test_markers_with_the_colon_outside_the_bold(self):
        proposal = PROPOSAL.replace("**Replaces:**", "**Replaces**:")
        tasks = TASKS.replace("**Closes:** AC1, AC3–AC4", "**Closes**: AC4–AC3, AC1").replace(
            "**Changes:** commit by D12-A, which replaced D4-A.", "**Changes:** commit per D4-A.")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, proposal=proposal, tasks=tasks))
            self.assertEqual(result.stdout.splitlines(), [
                "tasks.md:10: D4-A was replaced (proposal.md:6) but is still named here"])

    def test_criteria_are_defined_outside_the_progress_section_only(self):
        proposal = PROPOSAL + "- AC9 verified in the browser\n"
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, proposal=proposal))
            self.assertEqual(result.returncode, 0, result.stdout)

    def test_closes_lines_without_criteria_bullets_are_reported(self):
        proposal = "# demo\n\n| AC1 | export |\n"
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, proposal=proposal))
            self.assertEqual(result.returncode, 1)
            self.assertIn("no criterion", result.stdout)

    def test_unreadable_folder_exits_2(self):
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(Path(tmp) / "missing")
            self.assertEqual(result.returncode, 2)
            self.assertIn("proposal.md", result.stderr)


if __name__ == "__main__":
    unittest.main()
