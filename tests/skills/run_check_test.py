import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import time
import unittest


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/hive-core/scripts/checks/run_check.py"
sys.dont_write_bytecode = True  # keep __pycache__ out of the distributed skill folder


def alive(pid):
    try:
        os.kill(pid, 0)
    except ProcessLookupError:
        return False
    return True


class RunCheckTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.dir = Path(self.tmp.name)
        self.log = self.dir / "check.log"

    def tearDown(self):
        self.tmp.cleanup()

    def run_check(self, *args, timeout=60):
        return subprocess.run(
            [sys.executable, "-I", str(SCRIPT), "--log", str(self.log), *args],
            capture_output=True,
            timeout=timeout,
        )

    def test_success_prints_the_log_path_first_and_keeps_the_output(self):
        result = self.run_check("--", "sh", "-c", "echo one; echo two")
        self.assertEqual(result.returncode, 0, result.stderr)
        lines = result.stdout.decode().splitlines()
        self.assertEqual(lines[0], f"log: {self.log}")
        self.assertEqual(lines[1:], ["one", "two"])
        self.assertEqual(self.log.read_bytes(), b"one\ntwo\n")

    def test_failure_returns_the_command_status(self):
        result = self.run_check("--", "sh", "-c", "echo broken; exit 3")
        self.assertEqual(result.returncode, 3)

    def test_log_holds_only_the_command_output(self):
        script = "i=1; while [ $i -le 1000 ]; do echo line$i; i=$((i+1)); done; exit 1"
        result = self.run_check("--", "sh", "-c", script)
        self.assertEqual(result.returncode, 1)
        expected = "".join(f"line{i}\n" for i in range(1, 1001)).encode()
        self.assertEqual(self.log.read_bytes(), expected)
        shown = result.stdout.decode().splitlines()[1:]
        self.assertEqual(shown, [f"line{i}" for i in range(961, 1001)])

    def test_stderr_goes_to_the_log(self):
        result = self.run_check("--", "sh", "-c", "echo out; echo err >&2")
        self.assertEqual(result.returncode, 0)
        self.assertIn(b"err\n", self.log.read_bytes())

    def test_last_line_without_newline_and_invalid_utf8_are_shown(self):
        result = self.run_check("--", "printf", "first\\nlast\\377")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue(result.stdout.endswith(b"last\xff\n") or result.stdout.endswith(b"last\xff"))

    def test_shown_tail_is_bounded(self):
        result = self.run_check("--", sys.executable, "-c", "print('x' * 200000)")
        self.assertEqual(result.returncode, 0)
        self.assertLess(len(result.stdout), 17 * 1024)
        self.assertEqual(self.log.stat().st_size, 200001)

    def test_grep_shows_matching_lines(self):
        result = self.run_check("--grep", "FAIL", "--", "sh", "-c", "echo ok a; echo FAIL b; echo ok c; echo FAIL d")
        self.assertEqual(result.stdout.decode().splitlines()[1:], ["FAIL b", "FAIL d"])

    def test_missing_command_exits_127(self):
        result = self.run_check("--", str(self.dir / "no-such-command"))
        self.assertEqual(result.returncode, 127)
        self.assertIn(b"run_check:", result.stderr)

    def test_command_killed_by_a_signal_exits_128_plus_signal(self):
        result = self.run_check("--", "sh", "-c", "kill -TERM $$")
        self.assertEqual(result.returncode, 128 + signal.SIGTERM)

    def test_cap_stops_a_runaway_output(self):
        cap = 1024 * 1024
        result = self.run_check("--max-bytes", str(cap), "--", "yes")
        self.assertEqual(result.returncode, 125)
        self.assertIn(str(self.log).encode(), result.stderr)
        self.assertIn(str(cap).encode(), result.stderr)
        self.assertLessEqual(self.log.stat().st_size, cap)
        self.assertGreater(self.log.stat().st_size, cap // 2)

    def test_cap_escalates_to_kill_when_term_is_ignored(self):
        cap = 64 * 1024
        start = time.monotonic()
        result = self.run_check("--max-bytes", str(cap), "--", "sh", "-c", "trap '' TERM; yes")
        self.assertEqual(result.returncode, 125)
        self.assertLess(time.monotonic() - start, 30)

    def test_background_grandchild_does_not_survive(self):
        pidfile = self.dir / "pid"
        result = self.run_check("--", "sh", "-c", f"sleep 60 & echo $! > {pidfile}; echo done")
        self.assertEqual(result.returncode, 0)
        pid = int(pidfile.read_text())
        self.assertFalse(alive(pid), "background grandchild still running")

    def test_existing_log_or_symlink_is_refused(self):
        self.log.write_text("keep\n")
        result = self.run_check("--", "true")
        self.assertEqual(result.returncode, 2)
        self.assertEqual(self.log.read_text(), "keep\n")
        self.log.unlink()
        target = self.dir / "target"
        target.write_text("keep\n")
        self.log.symlink_to(target)
        result = self.run_check("--", "true")
        self.assertEqual(result.returncode, 2)
        self.assertEqual(target.read_text(), "keep\n")

    def test_term_to_the_helper_stops_the_command_group(self):
        pidfile = self.dir / "pid"
        proc = subprocess.Popen(
            [sys.executable, "-I", str(SCRIPT), "--log", str(self.log), "--", "sh", "-c", f"echo $$ > {pidfile}; sleep 60"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        deadline = time.monotonic() + 10
        while not (pidfile.exists() and pidfile.read_text().strip()):
            self.assertLess(time.monotonic(), deadline, "command never started")
            time.sleep(0.05)
        pid = int(pidfile.read_text())
        proc.send_signal(signal.SIGTERM)
        proc.communicate(timeout=30)
        self.assertNotEqual(proc.returncode, 0)
        deadline = time.monotonic() + 5
        while alive(pid) and time.monotonic() < deadline:
            time.sleep(0.05)
        self.assertFalse(alive(pid), "command survived the helper")

    def test_invalid_grep_is_refused_before_the_command_runs(self):
        marker = self.dir / "ran"
        result = self.run_check("--grep", "FAIL[", "--", "sh", "-c", f"touch {marker}")
        self.assertEqual(result.returncode, 2)
        self.assertFalse(marker.exists(), "the command ran despite the invalid pattern")

    def test_command_dies_when_the_helper_is_killed(self):
        pidfile = self.dir / "pid"
        proc = subprocess.Popen(
            [sys.executable, "-I", str(SCRIPT), "--log", str(self.log), "--", "sh", "-c", f"echo $$ > {pidfile}; exec sleep 60"],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )
        deadline = time.monotonic() + 10
        while not (pidfile.exists() and pidfile.read_text().strip()):
            self.assertLess(time.monotonic(), deadline, "command never started")
            time.sleep(0.05)
        pid = int(pidfile.read_text())
        proc.kill()
        proc.wait(timeout=10)
        deadline = time.monotonic() + 10
        while alive(pid) and time.monotonic() < deadline:
            time.sleep(0.1)
        self.assertFalse(alive(pid), "command outlived a killed helper")

    def test_command_that_reaps_all_its_children_finishes(self):
        reaper = (
            "import os, subprocess\n"
            "subprocess.Popen(['true'])\n"
            "while True:\n"
            "    try:\n"
            "        os.wait()\n"
            "    except ChildProcessError:\n"
            "        break\n"
            "print('reaped')\n"
        )
        result = self.run_check("--", sys.executable, "-c", reaper, timeout=30)
        self.assertEqual(result.returncode, 0)
        self.assertIn(b"reaped", result.stdout)

    def test_default_log_is_created_in_the_temp_dir(self):
        result = subprocess.run(
            [sys.executable, "-I", str(SCRIPT), "--", "true"],
            capture_output=True,
            timeout=60,
            env={**os.environ, "TMPDIR": str(self.dir)},
        )
        self.assertEqual(result.returncode, 0)
        path = Path(result.stdout.decode().splitlines()[0].removeprefix("log: "))
        self.assertEqual(path.parent.resolve(), self.dir.resolve())
        self.assertTrue(path.name.startswith("hive-check-"))


if __name__ == "__main__":
    unittest.main()
