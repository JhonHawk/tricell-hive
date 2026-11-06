from pathlib import Path
import re
import subprocess
import sys
import tempfile
import textwrap
import unittest


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/flow-close/scripts/apply_deltas.py"
sys.dont_write_bytecode = True  # keep __pycache__ out of the distributed skill folder

SPEC = """
    # cap

    ## Purpose

    The capability.

    ## Requirements

    ### Requirement: Alpha
    Alpha SHALL exist.

    #### Scenario: First
    - **WHEN** a
    - **THEN** b

    #### Scenario: Second
    - **WHEN** c
    - **THEN** d

    ### Requirement: Beta
    Beta SHALL exist.

    ### Requirement: Gamma
    Gamma SHALL exist.
    """

DELTA = """
    ## MODIFIED Requirements

    ### Requirement: Alpha
    Alpha SHALL exist now.

    #### Scenario: First
    - **WHEN** a
    - **THEN** b

    ## REMOVED Requirements

    ### Requirement: Beta
    Obsolete.

    ## ADDED Requirements

    ### Requirement: Delta
    Delta SHALL exist.

    ```markdown
    ### Requirement: Inside a fence
    ```
    """

PROPOSAL = """
    # demo — change

    | Field | Value |
    | --- | --- |
    | Status | In build · candidate verified |
    """


def write(path, text):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(textwrap.dedent(text).lstrip())


class ApplyDeltasTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.specs = Path(self.tmp.name) / "openspec"
        self.change = self.specs / "changes" / "demo"
        write(self.specs / "specs/cap/spec.md", SPEC)
        write(self.change / "proposal.md", PROPOSAL)
        write(self.change / "tasks.md", "# Tasks\n")
        write(self.change / "specs/cap/spec.md", DELTA)

    def tearDown(self):
        self.tmp.cleanup()

    def run_script(self, *args):
        return subprocess.run(
            [sys.executable, "-B", str(SCRIPT), str(self.change), "--specs", str(self.specs), *args],
            text=True,
            capture_output=True,
        )

    def snapshot(self):
        return {p: p.read_bytes() for p in sorted(self.specs.rglob("*")) if p.is_file()}

    def archived(self):
        return list((self.specs / "changes/archive").glob("*-demo"))

    def test_preview_lists_operations_and_target_and_writes_nothing(self):
        before = self.snapshot()
        result = self.run_script()
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertIn('cap: MODIFIED "Alpha"', result.stdout)
        self.assertIn('cap: drops scenario "Second" from "Alpha"', result.stdout)
        self.assertIn('cap: REMOVED "Beta"', result.stdout)
        self.assertIn('cap: ADDED "Delta"', result.stdout)
        self.assertNotIn("Inside a fence", result.stdout)
        self.assertRegex(result.stdout, r"archive: .*changes/archive/\d{4}-\d{2}-\d{2}-\d{4}-demo")
        self.assertEqual(self.snapshot(), before)

    def test_apply_merges_marks_completed_and_moves(self):
        result = self.run_script("--apply")
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        spec = (self.specs / "specs/cap/spec.md").read_text()
        outside_fences = re.sub(r"^```.*?^```$", "", spec, flags=re.S | re.M)
        names = re.findall(r"^### Requirement: (.+)$", outside_fences, re.M)
        self.assertEqual(names, ["Alpha", "Gamma", "Delta"])
        self.assertIn("Alpha SHALL exist now.", spec)
        self.assertNotIn("Scenario: Second", spec)
        self.assertIn("### Requirement: Inside a fence\n```", spec)
        self.assertFalse(self.change.exists())
        [target] = self.archived()
        self.assertRegex(target.name, r"^\d{4}-\d{2}-\d{2}-\d{4}-demo$")
        self.assertIn("| Status | Completed · candidate verified |", (target / "proposal.md").read_text())

    def test_new_capability_gets_the_spec_skeleton(self):
        write(self.change / "specs/fresh/spec.md", "## ADDED Requirements\n\n### Requirement: One\nOne SHALL exist.\n")
        result = self.run_script("--apply")
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        fresh = (self.specs / "specs/fresh/spec.md").read_text()
        self.assertTrue(fresh.startswith("# fresh\n\n## Purpose\n\n"), fresh)
        self.assertIn("## Requirements\n\n### Requirement: One\nOne SHALL exist.\n", fresh)

    def test_refusals_write_nothing(self):
        cases = {
            "missing MODIFIED target": (DELTA.replace("### Requirement: Alpha", "### Requirement: Omega"), "Omega"),
            "ADDED with other text": (DELTA.replace("### Requirement: Delta\n    Delta", "### Requirement: Gamma\n    Other"), "Gamma"),
            "repeated in one section": (DELTA + "\n    ### Requirement: Delta\n    Again.\n", "Delta"),
            "in two sections": (DELTA.replace("### Requirement: Beta", "### Requirement: Alpha"), "Alpha"),
            "no operations": ("Nothing here.\n", "no operations"),
        }
        for label, (delta, expected) in cases.items():
            with self.subTest(label):
                write(self.change / "specs/cap/spec.md", delta)
                before = self.snapshot()
                result = self.run_script("--apply")
                self.assertEqual(result.returncode, 1, result.stdout)
                self.assertIn(expected, result.stdout)
                self.assertEqual(self.snapshot(), before)
                self.assertTrue(self.change.exists())

    def test_new_capability_with_modified_is_refused(self):
        write(self.change / "specs/fresh/spec.md", "## MODIFIED Requirements\n\n### Requirement: One\nOne.\n")
        before = self.snapshot()
        result = self.run_script("--apply")
        self.assertEqual(result.returncode, 1)
        self.assertIn("fresh", result.stdout)
        self.assertEqual(self.snapshot(), before)

    def test_spec_without_requirements_section_is_refused(self):
        write(self.specs / "specs/cap/spec.md", "# cap\n\n## Purpose\n\nNo requirements here.\n")
        before = self.snapshot()
        result = self.run_script("--apply")
        self.assertEqual(result.returncode, 1)
        self.assertIn("## Requirements", result.stdout)
        self.assertEqual(self.snapshot(), before)

    def test_existing_archive_target_is_refused(self):
        preview = self.run_script()
        target = Path(re.search(r"archive: (\S+)", preview.stdout).group(1))
        target.mkdir(parents=True)
        before = self.snapshot()
        result = self.run_script("--apply")
        if result.returncode == 0:  # the minute turned between the two runs
            self.skipTest("archive minute changed between runs")
        self.assertEqual(result.returncode, 1)
        self.assertIn("already exists", result.stdout)
        self.assertEqual(self.snapshot(), before)

    def test_absent_removed_and_identical_added_are_no_ops(self):
        write(self.change / "specs/cap/spec.md", "## REMOVED Requirements\n\n### Requirement: Nope\nGone.\n\n## ADDED Requirements\n\n### Requirement: Gamma\nGamma SHALL exist.\n")
        result = self.run_script("--apply")
        self.assertEqual(result.returncode, 0, result.stdout)
        self.assertEqual((self.specs / "specs/cap/spec.md").read_text(), textwrap.dedent(SPEC).lstrip())

    def test_lossy_or_unknown_structure_is_refused(self):
        cases = {
            "other ### heading in a block": (DELTA.replace("Alpha SHALL exist now.", "Alpha SHALL exist now.\n\n    ### Notes\n    A note."), "### Notes"),
            "RENAMED section": (DELTA + "\n    ## RENAMED Requirements\n\n    - FROM: `### Requirement: Gamma`\n    - TO: `### Requirement: Zeta`\n", "RENAMED"),
            "case variant": (DELTA.replace("## ADDED Requirements", "## Added requirements"), "Added requirements"),
        }
        for label, (delta, expected) in cases.items():
            with self.subTest(label):
                write(self.change / "specs/cap/spec.md", delta)
                before = self.snapshot()
                result = self.run_script("--apply")
                self.assertEqual(result.returncode, 1, result.stdout)
                self.assertIn(expected, result.stdout)
                self.assertEqual(self.snapshot(), before)

    def test_other_sections_are_listed_not_applied(self):
        write(self.change / "specs/cap/spec.md", DELTA + "\n    ## Operational constraint\n\n    Text.\n")
        result = self.run_script()
        self.assertEqual(result.returncode, 0, result.stdout)
        self.assertIn('cap: section "## Operational constraint" is not applied', result.stdout)

    def test_repeated_requirement_in_the_spec_is_refused(self):
        write(self.specs / "specs/cap/spec.md", SPEC + "\n    ### Requirement: Gamma\n    Gamma again.\n")
        before = self.snapshot()
        result = self.run_script("--apply")
        self.assertEqual(result.returncode, 1)
        self.assertIn('"Gamma"', result.stdout)
        self.assertEqual(self.snapshot(), before)

    def test_new_capability_with_only_removed_is_refused(self):
        write(self.change / "specs/typo/spec.md", "## REMOVED Requirements\n\n### Requirement: Q\nGone.\n")
        before = self.snapshot()
        result = self.run_script("--apply")
        self.assertEqual(result.returncode, 1)
        self.assertIn("typo", result.stdout)
        self.assertEqual(self.snapshot(), before)

    def test_proposal_without_status_is_refused(self):
        write(self.change / "proposal.md", "# demo\n\nNo status here.\n")
        before = self.snapshot()
        result = self.run_script("--apply")
        self.assertEqual(result.returncode, 1)
        self.assertIn("Status", result.stdout)
        self.assertEqual(self.snapshot(), before)

    def test_spanish_status_keeps_its_details(self):
        write(self.change / "proposal.md", "# demo\n\n| Campo | Valor |\n| --- | --- |\n| Estado | En build · integrado por el PR #9 |\n")
        result = self.run_script("--apply")
        self.assertEqual(result.returncode, 0, result.stdout)
        [target] = self.archived()
        self.assertIn("| Estado | Completado · integrado por el PR #9 |", (target / "proposal.md").read_text())

    def test_folder_outside_the_specs_changes_is_refused(self):
        other = Path(self.tmp.name) / "elsewhere"
        other.mkdir()
        result = subprocess.run([sys.executable, "-B", str(SCRIPT), str(self.change), "--specs", str(other), "--apply"], text=True, capture_output=True)
        self.assertEqual(result.returncode, 1)
        self.assertIn("changes", result.stdout)
        self.assertFalse((other / "specs").exists())

    def test_nested_fences_keep_headings_as_text(self):
        nested = "## ADDED Requirements\n\n### Requirement: Delta\nDelta SHALL exist.\n\n````markdown\n```js\n### Requirement: Fake\n```\n````\n"
        write(self.change / "specs/cap/spec.md", nested)
        result = self.run_script()
        self.assertEqual(result.returncode, 0, result.stdout)
        self.assertNotIn("Fake", result.stdout)

    def test_missing_proposal_exits_2(self):
        (self.change / "proposal.md").unlink()
        result = self.run_script()
        self.assertEqual(result.returncode, 2)
        self.assertIn("proposal.md", result.stderr)


if __name__ == "__main__":
    unittest.main()
