import pathlib
import subprocess
import sys
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
TASKS = ROOT / "content/skills/flow-plan/scripts/tasks.py"
CHECK = ROOT / "content/skills/flow-plan/scripts/check_change.py"

PLAN = """# 2026-01-02-example — Example

| Estado | In build · T1 de 1 |
| --- | --- |

## Alcance y criterios de aceptación

- AC1 — The export lists every row · false today because it stops at 100 · Observed in: the export page

## Tareas

### T1 — Export all rows

- [ ] The export lists every row

**Closes:** AC1

**Verification:** the export test passes

## Checkpoint

## Review status and progress
"""


def run(script, *args):
    return subprocess.run([sys.executable, "-I", str(script), *args], capture_output=True, text=True)


class PlanRecordVariantsTest(unittest.TestCase):
    def setUp(self):
        self.dir = tempfile.TemporaryDirectory()
        self.plan = pathlib.Path(self.dir.name) / "2026-01-02-example.plan.md"
        self.plan.write_text(PLAN)

    def tearDown(self):
        self.dir.cleanup()

    def test_status_set_accepts_a_translated_field_and_keeps_the_rest_of_the_cell(self):
        result = run(TASKS, str(self.plan), "status-set", "Completed")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("| Estado | Completed · T1 de 1 |", self.plan.read_text())

    def test_criteria_are_read_under_another_acceptance_heading(self):
        result = run(CHECK, str(self.plan))
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.plan.write_text(PLAN.replace(" · Observed in: the export page", ""))
        result = run(CHECK, str(self.plan))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Observed in", result.stdout + result.stderr)


if __name__ == "__main__":
    unittest.main()
