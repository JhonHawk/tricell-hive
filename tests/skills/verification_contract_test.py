import copy
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/flow-verify/scripts/check_contract.py"


class VerificationContractTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.repo = Path(self.tmp.name) / "repo"
        self.repo.mkdir()
        self.path = self.repo / "_support/docs/verification/project.json"
        (self.repo / "check.py").write_text("raise SystemExit(0)\n")
        (self.repo / "run.txt").write_text("observed exit 0\n")

    def cli(self, action, path=None):
        return subprocess.run(
            [sys.executable, "-I", "-B", str(SCRIPT), action,
             "--repo", str(self.repo), "--path", str(path or self.path)],
            capture_output=True, text=True, timeout=10,
        )

    def initialized(self):
        result = self.cli("init")
        self.assertEqual(result.returncode, 0, result.stderr)
        return json.loads(self.path.read_text())

    def check(self, data):
        self.path.write_text(json.dumps(data))
        return self.cli("check")

    def populated(self):
        data = self.initialized()
        target = data["targets"][0]
        for name in ("drive", "evidence"):
            target["capabilities"][name].update(
                references=[{"id": name + "-check", "path": "check.py",
                             "command": "python3 check.py"}],
                success="exit code 0",
            )
        target["features"] = [{"id": "export", "sources": ["check.py"],
                               "drive": ["drive-check"], "evidence": ["evidence-check"]}]
        return data

    def test_init_unverified_and_no_overwrite(self):
        data = self.initialized()
        self.assertEqual(data["schema_version"], 1)
        self.assertEqual(set(data["targets"][0]["capabilities"]),
                         {"launch", "doctor", "drive", "evidence", "cleanup"})
        for capability in data["targets"][0]["capabilities"].values():
            self.assertEqual(capability["status"], "unverified")
            self.assertEqual(capability["references"], [])
            self.assertIsNone(capability["evidence"])
        before = self.path.read_bytes()
        self.assertNotEqual(self.cli("init").returncode, 0)
        self.assertEqual(self.path.read_bytes(), before)
        self.assertEqual(self.cli("check").returncode, 0)

    def test_multiple_targets_and_features(self):
        data = self.populated()
        other = copy.deepcopy(data["targets"][0])
        other["id"] = "worker"
        data["targets"].append(other)
        self.assertEqual(self.check(data).returncode, 0)
        other["features"][0]["drive"] = ["missing"]
        self.assertNotEqual(self.check(data).returncode, 0)

    def test_statuses_and_verified_evidence(self):
        data = self.populated()
        capability = data["targets"][0]["capabilities"]["drive"]
        for status in ("unsupported", "not_applicable"):
            capability["status"] = status
            self.assertNotEqual(self.check(data).returncode, 0)
            capability["reason"] = "No supported driver in this project"
            self.assertEqual(self.check(data).returncode, 0)
            capability["reason"] = None
        capability["status"] = "verified"
        for evidence in (None, {}, {"revision": "abc", "result": "pass"},
                         {"revision": "abc", "result": "failed", "path": "run.txt"}):
            capability["evidence"] = evidence
            self.assertNotEqual(self.check(data).returncode, 0)
        capability["evidence"] = {"revision": "abc", "result": "pass", "path": "run.txt"}
        self.assertEqual(self.check(data).returncode, 0)

    def test_duplicate_json_version_ids_and_missing_fields(self):
        data = self.populated()
        self.path.write_text('{"schema_version":1,"schema_version":1,"targets":[]}')
        self.assertNotEqual(self.cli("check").returncode, 0)
        for version in (2, True, "1"):
            data["schema_version"] = version
            self.assertNotEqual(self.check(data).returncode, 0)
        data["schema_version"] = 1
        target = data["targets"][0]
        for key in ("cwd", "capabilities", "features"):
            broken = copy.deepcopy(data)
            del broken["targets"][0][key]
            self.assertNotEqual(self.check(broken).returncode, 0)
        target["features"].append(copy.deepcopy(target["features"][0]))
        self.assertNotEqual(self.check(data).returncode, 0)
        target["features"].pop()
        data["targets"].append(copy.deepcopy(target))
        self.assertNotEqual(self.check(data).returncode, 0)

    def test_paths_and_symlinks_confined(self):
        data = self.populated()
        outside = Path(self.tmp.name) / "outside"
        outside.mkdir()
        (outside / "file").write_text("private")
        (self.repo / "escape").symlink_to(outside, target_is_directory=True)
        for path in ("../outside/file", str(outside / "file"), "escape/file", "absent"):
            data["targets"][0]["features"][0]["sources"] = [path]
            self.assertNotEqual(self.check(data).returncode, 0)
        data["targets"][0]["features"][0]["sources"] = ["check.py"]
        data["targets"][0]["cwd"] = "escape"
        self.assertNotEqual(self.check(data).returncode, 0)
        for path in (outside / "new.json", self.repo / "escape/new.json",
                     self.repo / "../outside/new.json"):
            self.assertNotEqual(self.cli("init", path).returncode, 0)
        self.assertFalse((outside / "new.json").exists())

    def test_commands_are_data(self):
        data = self.populated()
        data["targets"][0]["capabilities"]["drive"]["references"][0]["command"] = (
            "touch " + str(self.repo / "executed"))
        self.assertEqual(self.check(data).returncode, 0)
        self.assertFalse((self.repo / "executed").exists())

    def test_init_preserves_dangling_symlink(self):
        self.path.parent.mkdir(parents=True)
        destination = self.path.parent / "missing.json"
        self.path.symlink_to(destination)
        self.assertNotEqual(self.cli("init").returncode, 0)
        self.assertTrue(self.path.is_symlink())
        self.assertFalse(destination.exists())

    def test_invalid_reference_and_evidence_paths(self):
        data = self.populated()
        capability = data["targets"][0]["capabilities"]["drive"]
        for path in ("absent", "../outside", str(self.repo / "check.py")):
            capability["references"][0]["path"] = path
            self.assertNotEqual(self.check(data).returncode, 0)
        capability["references"][0]["path"] = "check.py"
        capability["references"].append(copy.deepcopy(capability["references"][0]))
        self.assertNotEqual(self.check(data).returncode, 0)
        capability["references"].pop()
        capability["status"] = "verified"
        capability["evidence"] = {"revision": "abc", "result": "pass", "path": "absent"}
        self.assertNotEqual(self.check(data).returncode, 0)

    def test_malformed_inputs_fail_without_traceback(self):
        self.initialized()
        for value in ('{"targets":', 'null', '[]', '{"schema_version": NaN,"targets":[]}'):
            self.path.write_text(value)
            result = self.cli("check")
            self.assertNotEqual(result.returncode, 0)
            self.assertNotIn("Traceback", result.stderr)


if __name__ == "__main__":
    unittest.main()
