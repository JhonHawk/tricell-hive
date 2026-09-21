import json
from pathlib import Path
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
HELPER = ROOT / "content/skills/engram-init-workspace/scripts/init_workspace.py"


class EngramWorkspaceHelperTest(unittest.TestCase):
    def run_helper(self, *args):
        return subprocess.run(["python3", str(HELPER), *args], text=True, capture_output=True)

    def test_dry_run_requires_explicit_contained_target(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp) / "workspace"
            child = root / "app"
            child.mkdir(parents=True)
            result = self.run_helper("--root", str(root), "--name", "acme", "--target", "app")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(json.loads(result.stdout)["status"], "dry-run")
            self.assertFalse((child / ".engram/config.json").exists())
            outside = self.run_helper("--root", str(root), "--name", "acme", "--target", "../outside")
            self.assertEqual(outside.returncode, 2)

    def test_conflicting_target_blocks_all_apply(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp) / "workspace"
            first, second = root / "first", root / "second"
            (second / ".engram").mkdir(parents=True)
            second.joinpath(".engram/config.json").write_text('{"project_name":"other"}\n')
            first.mkdir(parents=True)
            result = self.run_helper("--root", str(root), "--name", "shared", "--target", "first", "--target", "second", "--apply")
            self.assertEqual(result.returncode, 2)
            self.assertFalse((first / ".engram/config.json").exists())
            self.assertEqual(json.loads(second.joinpath(".engram/config.json").read_text())["project_name"], "other")

    def test_explicit_standalone_root_is_the_only_default_target(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp) / "repo"
            sibling = Path(tmp) / "sibling"
            root.mkdir(); sibling.mkdir()
            result = self.run_helper("--root", str(root), "--name", "only-this-repo", "--apply")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertTrue((root / ".engram/config.json").exists())
            self.assertFalse((sibling / ".engram/config.json").exists())

    def test_symlinked_engram_directory_blocks_without_following_it(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp) / "workspace"; target = root / "app"; outside = Path(tmp) / "outside"
            target.mkdir(parents=True); outside.mkdir()
            (target / ".engram").symlink_to(outside, target_is_directory=True)
            result = self.run_helper("--root", str(root), "--name", "shared", "--target", "app", "--apply")
            self.assertEqual(result.returncode, 2)
            self.assertFalse((outside / "config.json").exists())

    def test_regular_engram_path_blocks_all_targets_before_apply(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp) / "workspace"; first = root / "first"; second = root / "second"
            first.mkdir(parents=True); second.mkdir()
            second.joinpath(".engram").write_text("not a directory")
            result = self.run_helper("--root", str(root), "--name", "shared", "--target", "first", "--target", "second", "--apply")
            self.assertEqual(result.returncode, 2)
            self.assertFalse((first / ".engram/config.json").exists())


if __name__ == "__main__":
    unittest.main()
