"""Regression checks for explicit PI approval and the legacy hook boundary."""

from concurrent.futures import ThreadPoolExecutor
import hashlib
import json
from pathlib import Path
import subprocess
import tempfile
import unittest


HOOK = Path(__file__).with_name("flow-plan-capture.sh")


class PiCaptureTests(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.root = Path(self.scratch.name)

    def invoke(self, plan, *, explicit=True, raw=None):
        payload = raw if raw is not None else json.dumps(
            {"harness": "pi", "cwd": str(self.root), "tool_response": {"plan": plan}}
        )
        return subprocess.run(
            ["/bin/bash", str(HOOK), *(["--from-pi-command"] if explicit else [])],
            input=payload, text=True, capture_output=True, check=False,
        )

    def make_repo(self):
        subprocess.run(["git", "init", "-q", str(self.root)], check=True)

    def test_standalone_exact_bytes_and_repeat(self):
        self.make_repo()
        plan = "# A plan\n\nKeep trailing whitespace.  \n\n\n"
        result = self.invoke(plan)
        self.assertEqual(result.returncode, 0, result.stderr)
        response = json.loads(result.stdout)
        self.assertEqual(response["status"], "captured")
        self.assertEqual(response["sha256"], hashlib.sha256(plan.encode()).hexdigest())
        target = Path(response["path"])
        self.assertEqual(target.read_bytes().split(b"\n", 4)[4], plan.encode())
        self.assertEqual(json.loads(self.invoke(plan).stdout)["path"], str(target))
        self.assertNotIn("commit", result.stdout.lower())

    def test_flow_capture_and_advanced_revision(self):
        (self.root / "_support").mkdir()
        (self.root / "_support/PROJECT.md").write_text("# Workspace\n")
        first = json.loads(self.invoke("# Plan\nfirst").stdout)
        target = Path(first["path"])
        target.write_text(target.read_text().replace("Status: planned", "Status: building", 1))
        second = json.loads(self.invoke("# Plan\nchanged").stdout)
        self.assertNotEqual(first["path"], second["path"])
        self.assertIn("Status: building", target.read_text())

    def test_non_repository_is_explicitly_session_only(self):
        response = json.loads(self.invoke("# Plan\n").stdout)
        self.assertEqual(response, {"status": "session_only", "reason": "no-repository"})
        self.assertFalse((self.root / "_support").exists())

    def test_session_opt_out(self):
        self.make_repo()
        response = json.loads(self.invoke("# Plan\nSession: no\n").stdout)
        self.assertEqual(response, {"status": "skipped", "reason": "session-no"})
        self.assertFalse((self.root / "_support").exists())

    def test_opt_out_outside_repository_is_explicit(self):
        response = json.loads(self.invoke("# Plan\nSession: no\n").stdout)
        self.assertEqual(response, {"status": "skipped", "reason": "session-no"})

    def test_legacy_non_flow_hook_stays_silent(self):
        self.make_repo()
        result = self.invoke("# Plan\n", explicit=False)
        self.assertEqual(result.returncode, 0)
        self.assertEqual(result.stdout, "")
        self.assertFalse((self.root / "_support").exists())

    def test_invalid_payload_is_failure(self):
        for payload in ("{", json.dumps({"cwd": str(self.root), "tool_response": {"plan": 42}})):
            with self.subTest(payload=payload):
                result = self.invoke(None, raw=payload)
                self.assertNotEqual(result.returncode, 0)
                self.assertEqual(json.loads(result.stdout)["status"], "error")

    def test_index_failure_never_becomes_success_on_retry(self):
        self.make_repo()
        (self.root / "_support/sessions/README.md").mkdir(parents=True)
        for _ in range(2):
            result = self.invoke("# Plan\n")
            self.assertNotEqual(result.returncode, 0)
            self.assertEqual(json.loads(result.stdout)["status"], "error")

    def test_concurrent_approvals_keep_distinct_exact_snapshots(self):
        self.make_repo()
        plans = [f"# Concurrent plan\nVersion {index}\n" for index in range(8)]
        with ThreadPoolExecutor(max_workers=8) as pool:
            results = list(pool.map(self.invoke, plans))
        targets = []
        for plan, result in zip(plans, results):
            self.assertEqual(result.returncode, 0, result.stderr)
            target = Path(json.loads(result.stdout)["path"])
            targets.append(target)
            self.assertEqual(target.read_bytes().split(b"\n", 4)[4], plan.encode())
        self.assertEqual(len(set(targets)), len(plans))
        index = (self.root / "_support/sessions/README.md").read_text()
        for target in targets:
            self.assertIn(target.parent.name, index)

    def test_multi_digit_snapshot_suffix_is_reused(self):
        self.make_repo()
        for version in range(11):
            result = self.invoke(f"# Plan\nVersion {version}")
        original = json.loads(result.stdout)["path"]
        self.assertEqual(json.loads(self.invoke("# Plan\nVersion 10").stdout)["path"], original)

    def test_relative_cwd_terminates(self):
        result = subprocess.run(
            ["/bin/bash", str(HOOK.resolve()), "--from-pi-command"],
            input=json.dumps({"cwd": ".", "tool_response": {"plan": "# Plan"}}),
            cwd=self.root, text=True, capture_output=True, timeout=5, check=True,
        )
        self.assertEqual(json.loads(result.stdout)["status"], "session_only")

    def test_legacy_index_error_keeps_context(self):
        (self.root / "_support/sessions/README.md").mkdir(parents=True)
        (self.root / "_support/PROJECT.md").write_text("# Workspace")
        result = self.invoke("# Legacy", explicit=False)
        self.assertEqual(result.returncode, 0)
        self.assertIn("additionalContext", result.stdout)

    def test_unwritable_destination_is_failure(self):
        self.make_repo()
        (self.root / "_support").write_text("occupied")
        result = self.invoke("# Plan\n")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(json.loads(result.stdout)["status"], "error")


if __name__ == "__main__":
    unittest.main()
