from pathlib import Path
import subprocess
import sys
import tempfile
import textwrap
import unittest


ROOT = Path(__file__).resolve().parents[2]
CHECK = ROOT / "content/skills/flow-plan/scripts/check_change.py"
PLAN_NAME = "2026-11-24-demo.plan.md"
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

    - AC1 — The export button appears · false today · Observed in: the export page served locally
    - AC2 — Retired criterion **Retired:** already true on the base
    - AC3 — The export runs · false today · **Observed in:** the exported file
    - AC4 — Pagination works · false today
      **Observed in**: the list page served locally

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


def plan_text(proposal, tasks):
    """One plan file: the proposal's sections, then `## Tareas` with the task blocks and their progress section."""
    head = textwrap.dedent(proposal).lstrip().split("## Estado de la revisión y avance")[0].rstrip()
    body = textwrap.dedent(tasks).lstrip().replace("# Tasks\n\n", "", 1)
    return f"{head}\n\n## Tareas\n\n{body}"


class ChangeCheckTest(unittest.TestCase):
    def run_check(self, plan, *args):
        return subprocess.run(
            [sys.executable, "-B", str(CHECK), str(plan), *args],
            text=True,
            capture_output=True,
        )

    def folder(self, tmp, proposal=PROPOSAL, tasks=TASKS):
        """Write the plan file in its work folder and return its path."""
        plan = Path(tmp) / "2026-11-24-demo" / PLAN_NAME
        plan.parent.mkdir(parents=True, exist_ok=True)
        plan.write_text(plan_text(proposal, tasks))
        return plan

    def test_coherent_plan_passes(self):
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp))
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            self.assertEqual(result.stdout, "")

    def test_reports_a_criterion_no_task_closes(self):
        tasks = TASKS.replace("AC1, AC3–AC4", "AC1, AC3")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, tasks=tasks))
            self.assertEqual(result.returncode, 1)
            self.assertIn(f"{PLAN_NAME}:13: AC4", result.stdout)
            self.assertNotIn("AC2", result.stdout)

    def test_reports_a_closes_line_naming_an_undefined_criterion(self):
        tasks = TASKS.replace("AC1, AC3–AC4", "AC1, AC3–AC4, AC9")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, tasks=tasks))
            self.assertEqual(result.returncode, 1)
            self.assertIn(f"{PLAN_NAME}:23: AC9", result.stdout)

    def test_reports_each_stale_mention_of_a_replaced_or_retired_id(self):
        tasks = TASKS.replace(
            "**Changes:** commit by D12-A, which replaced D4-A.",
            "**Changes:** commit by D4-A and close AC2.",
        )
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, tasks=tasks))
            self.assertEqual(result.returncode, 1)
            lines = result.stdout.splitlines()
            self.assertIn(f"{PLAN_NAME}:25: D4-A", result.stdout)
            self.assertIn(f"{PLAN_NAME}:25: AC2", result.stdout)
            self.assertEqual(len(lines), 2, result.stdout)

    def test_a_changed_option_letter_is_a_replacement(self):
        proposal = PROPOSAL.replace("- **D12 A** — Commit each task when reported. **Replaces:** D4-A", "- **D4 B** — Commit each task when reported. **Replaces:** D4-A")
        tasks = TASKS.replace("**Changes:** commit by D12-A, which replaced D4-A.", "**Changes:** commit per D4-A; D4-B is current.")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, proposal=proposal, tasks=tasks))
            self.assertEqual(result.returncode, 0, result.stdout)
            tasks = TASKS.replace("**Changes:** commit by D12-A, which replaced D4-A.", "**Changes:** commit per D4-A.")
            result = self.run_check(self.folder(tmp, proposal=proposal, tasks=tasks))
            self.assertIn(f"{PLAN_NAME}:25: D4-A", result.stdout)

    def test_only_the_defining_line_keeps_an_old_id(self):
        tasks = TASKS.replace("**Changes:** commit by D12-A, which replaced D4-A.", "- AC2 — check the retired thing still")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, tasks=tasks))
            self.assertIn(f"{PLAN_NAME}:25: AC2", result.stdout)

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
                f"{PLAN_NAME}:25: D4-A was replaced ({PLAN_NAME}:6) but is still named here"])

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

    def test_replaced_id_lists_in_ranges_and_markup(self):
        cases = {
            "**Replaces:** AC3–AC4": "AC4",
            "**Replaces:** `D4-A`": "D4-A",
            "**Replaces:** [**D4-A**](#d4)": "D4-A",
            "**Replaces:** D4 In the revision": "D4",
        }
        for marker, stale in cases.items():
            proposal = PROPOSAL.replace("**Replaces:** D4-A", marker)
            tasks = TASKS.replace("**Changes:** commit by D12-A, which replaced D4-A.", f"**Changes:** still {stale} here.")
            with self.subTest(marker=marker), tempfile.TemporaryDirectory() as tmp:
                result = self.run_check(self.folder(tmp, proposal=proposal, tasks=tasks))
                self.assertIn(f"{PLAN_NAME}:25: {stale}", result.stdout)
                self.assertNotIn("named by no", result.stdout)

    def test_naming_the_replacement_without_its_letter_exempts_the_line(self):
        tasks = TASKS.replace("which replaced D4-A", "D12, which replaced D4-A")
        tasks = tasks.replace("commit by D12-A, ", "commit by ")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, tasks=tasks))
            self.assertEqual(result.returncode, 0, result.stdout)

    def test_closes_none_without_criteria_passes(self):
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, proposal="# demo\n", tasks="### T1\n\n**Closes:** none (refactor)\n"))
            self.assertEqual(result.returncode, 0, result.stdout)

    def test_checkpoint_section_may_keep_a_replaced_id(self):
        checkpoint = """
    ## Checkpoint

    - Open findings: H1: keep D4-A

    ### Evidence

    - suite · pass · D4-A

    ## Review status and progress"""
        tasks = TASKS.replace("\n    ## Review status and progress", checkpoint)
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, tasks=tasks))
            self.assertEqual(result.returncode, 0, result.stdout)
            outside = tasks.replace("**Changes:** commit by D12-A, which replaced D4-A.", "**Changes:** commit by D12-A, which replaced D4-A.\n\n**Verification:** D4-A.")
            result = self.run_check(self.folder(tmp, tasks=outside))
            self.assertEqual(result.returncode, 1)

    def test_checkpoint_rules_and_principles_lines_are_checked(self):
        def with_lines(*extra):
            block = "\n    ## Checkpoint\n\n    - Next: T2\n" + "".join(f"    {line}\n" for line in extra) + "\n    ## Review status and progress"
            return TASKS.replace("\n    ## Review status and progress", block)
        good = with_lines("- Rules: flow-verify/stage-entry.md, flow-verify/stage-exit.md · extra: defect-fix, tests", "- Principles: boundary-discipline, hive-typescript")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, tasks=good))
            self.assertEqual(result.returncode, 0, result.stdout)
        for bad in ("- Rules:", "- Rules: stage.md", "- Rules: flow-build/stage.md · extra:", "- Principles:", "- Principles: none, tests", "- Principles: Tests"):
            with tempfile.TemporaryDirectory() as tmp:
                result = self.run_check(self.folder(tmp, tasks=with_lines(bad)))
                self.assertEqual(result.returncode, 1, bad)
                self.assertIn(bad.strip(), result.stdout)

    def test_reports_an_active_criterion_without_its_observed_place(self):
        proposal = PROPOSAL.replace(" · **Observed in:** the exported file", "")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, proposal=proposal))
            self.assertEqual(result.returncode, 1, result.stdout)
            self.assertEqual(result.stdout.splitlines(), [f"{PLAN_NAME}:12: AC3 names no **Observed in:** place"])

    def test_the_observed_place_may_sit_in_a_sub_bullet_or_after_a_blank_line(self):
        proposal = PROPOSAL.replace(
            "    - AC1 — The export button appears · false today · Observed in: the export page served locally\n",
            "    - AC1 — The export button appears\n      - false today\n      - Observed in: the export page served locally\n",
        ).replace(
            "    - AC3 — The export runs · false today · **Observed in:** the exported file\n",
            "    - AC3 — The export runs · false today\n\n      **Observed in:** the exported file\n\n",
        )
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, proposal=proposal))
            self.assertEqual(result.stdout, "")

    def test_an_observed_marker_without_a_place_is_reported(self):
        for empty in (" · Observed in:", " · Observed in: <where its real effect shows>"):
            proposal = PROPOSAL.replace(" · **Observed in:** the exported file", empty)
            with self.subTest(marker=empty), tempfile.TemporaryDirectory() as tmp:
                result = self.run_check(self.folder(tmp, proposal=proposal))
                self.assertIn("AC3 names no **Observed in:** place", result.stdout)

    def test_observed_place_edge_cases_that_name_a_place(self):
        cases = {
            "autolink": (" · **Observed in:** the exported file", " · Observed in: <http://localhost:3000/export>"),
            "wrapped": ("    - AC3 — The export runs · false today · **Observed in:** the exported file\n",
                        "    - AC3 — The export runs as issue\n      #431 asks · **Observed in:**\n      the exported file\n"),
            "marker named earlier": (" · **Observed in:** the exported file", " · requires `Observed in: <place>` · Observed in: the check's output"),
            "quoted": ("    ## Scope and acceptance\n", "    ## Request\n\n    - AC1 — quoted from the ticket\n\n    ## Scope and acceptance\n"),
        }
        for name, (old, new) in cases.items():
            with self.subTest(case=name), tempfile.TemporaryDirectory() as tmp:
                self.assertIn(old, PROPOSAL)
                result = self.run_check(self.folder(tmp, proposal=PROPOSAL.replace(old, new)))
                self.assertNotIn("Observed in", result.stdout)

    def test_observed_place_edge_cases_that_name_none(self):
        marker = " · **Observed in:** the exported file"
        cases = {
            "placeholder with text": " · Observed in: <place> (TBD)",
            "dash": " · Observed in: —",
            "nested criterion": "\n      - AC9 — nested · Observed in: its own page",
            "ordered item": "\n    1. A step. Observed in: elsewhere",
        }
        for name, new in cases.items():
            with self.subTest(case=name), tempfile.TemporaryDirectory() as tmp:
                result = self.run_check(self.folder(tmp, proposal=PROPOSAL.replace(marker, new)))
                self.assertIn("AC3 names no **Observed in:** place", result.stdout)

    UNITS = """
    ## Delivery units

    ### E1 — PR 1 · depende de: ninguna · archivos: a.py, b.py

    - [ ] unit: python3 -m unittest
    - [ ] live: open the page · pass: the button renders
    - [ ] perf: p95 latency · base: 120ms · falla: 200ms

    ### E2 — PR 2 · depends on: E1 · files: c.py

    - [ ] live: n/a: no interaction changes
    - [ ] perf: n/a: no hot path changes
    - [ ] perf: p95 latency · base: 120ms · fail: 200ms
    """

    def units_folder(self, tmp, units):
        tasks = textwrap.dedent(TASKS).replace("## Review status", textwrap.dedent(units).lstrip() + "\n## Review status")
        return self.folder(tmp, tasks=tasks)

    def test_valid_delivery_units_pass_in_both_languages(self):
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.units_folder(tmp, self.UNITS))
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_live_box_without_pass_is_rejected_with_its_line(self):
        units = self.UNITS.replace(" · pass: the button renders", "")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.units_folder(tmp, units))
            self.assertEqual(result.returncode, 1)
            self.assertIn(f"{PLAN_NAME}:32:", result.stdout)
            self.assertEqual(len(result.stdout.splitlines()), 1, result.stdout)

    def test_perf_box_without_base_or_threshold_is_rejected(self):
        for broken in ("p95 latency · falla: 200ms", "p95 latency · base: 120ms", "p95 latency"):
            units = self.UNITS.replace("p95 latency · base: 120ms · falla: 200ms", broken)
            with tempfile.TemporaryDirectory() as tmp:
                result = self.run_check(self.units_folder(tmp, units))
                self.assertEqual(result.returncode, 1, broken)
                self.assertIn(f"{PLAN_NAME}:33:", result.stdout)

    def test_unit_box_needs_a_command_and_unknown_kinds_are_rejected(self):
        for broken in ("- [ ] unit:", "- [ ] manual: look at it"):
            units = self.UNITS.replace("- [ ] unit: python3 -m unittest", broken)
            with tempfile.TemporaryDirectory() as tmp:
                result = self.run_check(self.units_folder(tmp, units))
                self.assertEqual(result.returncode, 1, broken)
                self.assertIn(f"{PLAN_NAME}:31:", result.stdout)

    def test_unit_heading_needs_its_dependencies_and_files(self):
        units = self.UNITS.replace(" · archivos: a.py, b.py", "")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.units_folder(tmp, units))
            self.assertEqual(result.returncode, 1)
            self.assertIn(f"{PLAN_NAME}:29:", result.stdout)

    def test_n_a_needs_a_reason(self):
        units = self.UNITS.replace("live: n/a: no interaction changes", "live: n/a:")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.units_folder(tmp, units))
            self.assertEqual(result.returncode, 1)

    def test_n_a_only_replaces_a_whole_live_or_perf_predicate(self):
        for bad in ("- [ ] manual: n/a: not a kind", "- [ ] unit: n/a: no command", "- [ ] unit: N/A: none",
                    "- [ ] perf: p95 · base: n/a: none · falla: 200ms", "- [ ] live: login · pass: n/a: nothing"):
            units = self.UNITS.replace("- [ ] live: n/a: no interaction changes", bad)
            with self.subTest(box=bad), tempfile.TemporaryDirectory() as tmp:
                result = self.run_check(self.units_folder(tmp, units))
                self.assertEqual(result.returncode, 1, result.stdout)

    def test_upper_case_n_a_replaces_a_whole_predicate(self):
        units = self.UNITS.replace("- [ ] live: n/a: no interaction changes", "- [ ] live: N/A: no interaction changes")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.units_folder(tmp, units))
            self.assertEqual(result.stdout, "")

    def test_archive_option_and_a_folder_are_refused(self):
        with tempfile.TemporaryDirectory() as tmp:
            plan = self.folder(tmp)
            archive = self.run_check(plan, "--archive")
            self.assertEqual(archive.returncode, 2)
            self.assertEqual(archive.stdout, "")
            folder = self.run_check(plan.parent)
            self.assertEqual(folder.returncode, 2)
            self.assertIn("not a plan file", folder.stderr)
            missing = self.run_check(Path(tmp) / "missing.plan.md")
            self.assertEqual(missing.returncode, 2)

    def test_a_criterion_may_be_a_plain_line_and_a_task_line_is_not_a_criterion(self):
        proposal = PROPOSAL.replace("    - AC1 —", "    AC1 —")
        tasks = TASKS.replace("**Changes:** commit by D12-A, which replaced D4-A.", "AC9 — a note inside a task")
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_check(self.folder(tmp, proposal=proposal, tasks=tasks))
            self.assertEqual(result.returncode, 0, result.stdout)

    def test_valid_business_rules_section_passes_and_malformed_lines_fail(self):
        valid_rules = """
    ## Business rules

    ticket section: PRJ-833 · Reglas de negocio · export
    ticket write: 2026-11-16 · build start · PRJ-833
    rule: Only admins may export CSV · screen Export · approved 2026-11-16: D2 A · found by plan
    rule: Default format is UTF-8 · screen none · assumption · found by build
    rule: Old rule · screen Export · cancelled 2026-11-16: replaced by admin rule · replaced by Only admins may export CSV · found by research
    card stale: Export · Old rule
    frame stale: 12:34 · Old rule
"""
        proposal_ok = PROPOSAL.replace("    ## Scope and acceptance\n", valid_rules + "\n    ## Scope and acceptance\n")
        with tempfile.TemporaryDirectory() as tmp:
            res_ok = self.run_check(self.folder(tmp, proposal=proposal_ok))
            self.assertEqual(res_ok.returncode, 0, res_ok.stdout)

        bad_stated = valid_rules.replace("assumption", "stated: ticket description")
        proposal_stated = PROPOSAL.replace("    ## Scope and acceptance\n", bad_stated + "\n    ## Scope and acceptance\n")
        with tempfile.TemporaryDirectory() as tmp:
            res_stated = self.run_check(self.folder(tmp, proposal=proposal_stated))
            self.assertEqual(res_stated.returncode, 1)
            self.assertIn("Business rules", res_stated.stdout)

        bad_line = valid_rules + "    - random bullet not following grammar\n"
        proposal_bad = PROPOSAL.replace("    ## Scope and acceptance\n", bad_line + "\n    ## Scope and acceptance\n")
        with tempfile.TemporaryDirectory() as tmp:
            res_bad = self.run_check(self.folder(tmp, proposal=proposal_bad))
            self.assertEqual(res_bad.returncode, 1)
            self.assertIn("Business rules", res_bad.stdout)


if __name__ == "__main__":
    unittest.main()
