import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/hive-rule-trackers/scripts/work_time.py"
sys.dont_write_bytecode = True  # keep __pycache__ out of the distributed skill folder


class WorkTimeTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.marks = Path(self.tmp.name) / "marks.jsonl"

    def tearDown(self):
        self.tmp.cleanup()

    def run_script(self, *args):
        return subprocess.run([sys.executable, "-I", str(SCRIPT), *args], capture_output=True, text=True, timeout=60)

    def mark(self, event, at):
        result = self.run_script("mark", str(self.marks), event, "--at", at)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_session_with_a_user_wait_and_a_ci_wait(self):
        # 40 min of work, a 15 min wait for the user, 10 of the 40 on remote CI.
        self.mark("start", "2026-01-01T10:00:00Z")
        self.mark("wait", "2026-01-01T10:10:00Z")
        self.mark("resume", "2026-01-01T10:25:00Z")
        self.mark("ci-start", "2026-01-01T10:35:00Z")
        self.mark("ci-end", "2026-01-01T10:45:00Z")
        self.mark("wait", "2026-01-01T10:55:00Z")
        out = json.loads(self.run_script("sum", str(self.marks)).stdout)
        self.assertEqual(out["active_seconds"], 40 * 60)
        self.assertEqual(out["user_wait_seconds"], 15 * 60)
        self.assertEqual(out["ci_wait_seconds"], 10 * 60)
        self.assertEqual(out["factor"], 2.0)
        self.assertEqual(out["converted_seconds"], 70 * 60)

    def test_converted_time_rounds_up_to_the_minute(self):
        self.mark("start", "2026-01-01T10:00:00Z")
        self.mark("wait", "2026-01-01T10:00:50Z")
        out = json.loads(self.run_script("sum", str(self.marks), "--factor", "1").stdout)
        self.assertEqual(out["active_seconds"], 50)
        self.assertEqual(out["converted_seconds"], 60)

    def test_since_starts_a_later_ticket_at_its_own_mark(self):
        self.mark("start", "2026-01-01T09:00:00Z")
        self.mark("start", "2026-01-01T10:00:00Z")
        self.mark("wait", "2026-01-01T10:20:00Z")
        out = json.loads(self.run_script("sum", str(self.marks), "--since", "2026-01-01T09:30:00Z").stdout)
        self.assertEqual(out["active_seconds"], 20 * 60)

    def test_rounding_matches_hive_time(self):
        # 1800 s at factor 1.1 is 1980 s in floating point plus a hair: 33 minutes, not 34.
        self.mark("start", "2026-01-01T10:00:00Z")
        self.mark("wait", "2026-01-01T10:30:00Z")
        out = json.loads(self.run_script("sum", str(self.marks), "--factor", "1.1").stdout)
        self.assertEqual(out["converted_seconds"], 1980)

    def test_ci_that_encloses_a_user_wait_counts_only_outside_it(self):
        # CI 10:05-10:37 encloses the wait 10:06-10:36: 2 min of CI, 10 min active.
        for event, at in (("start", "10:00"), ("ci-start", "10:05"), ("wait", "10:06"), ("resume", "10:36"), ("ci-end", "10:37"), ("wait", "10:40")):
            self.mark(event, f"2026-01-01T{at}:00Z")
        out = json.loads(self.run_script("sum", str(self.marks)).stdout)
        self.assertEqual(out["active_seconds"], 600)
        self.assertEqual(out["ci_wait_seconds"], 120)
        self.assertEqual(out["converted_seconds"], 18 * 60)

    def test_factor_outside_one_to_three_is_refused(self):
        self.mark("start", "2026-01-01T10:00:00Z")
        self.assertEqual(self.run_script("sum", str(self.marks), "--factor", "4").returncode, 2)
        self.assertEqual(self.run_script("sum", str(self.marks), "--factor", "0.5").returncode, 2)

    def test_marks_without_a_start_are_refused(self):
        self.mark("wait", "2026-01-01T10:00:00Z")
        result = self.run_script("sum", str(self.marks))
        self.assertEqual(result.returncode, 2)
        self.assertIn("no start mark", result.stderr)


if __name__ == "__main__":
    unittest.main()
