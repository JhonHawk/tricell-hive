import importlib.util
import json
import os
import pathlib
import shutil
import subprocess
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/flow-close/scripts/cleanup_branches.py"
GIT = shutil.which("git")


def load_module():
    spec = importlib.util.spec_from_file_location("cleanup_branches", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class CleanupBranchesTests(unittest.TestCase):
    def setUp(self):
        self.tmpdir = tempfile.TemporaryDirectory()
        self.tmp = pathlib.Path(self.tmpdir.name)
        self.env = dict(
            os.environ,
            GIT_CONFIG_GLOBAL=os.devnull,
            GIT_CONFIG_NOSYSTEM="1",
            GIT_AUTHOR_NAME="Ana",
            GIT_AUTHOR_EMAIL="ana@example.com",
            GIT_COMMITTER_NAME="Ana",
            GIT_COMMITTER_EMAIL="ana@example.com",
        )
        self.origin = self.tmp / "origin.git"
        self.main_repo = self.tmp / "main"
        self.git(self.tmp, "init", "--bare", "-b", "development", str(self.origin))
        self.git(self.tmp, "clone", "-q", str(self.origin), str(self.main_repo))
        (self.main_repo / ".gitignore").write_text(".env\n__pycache__/\n.DS_Store\n*.pyc\nnode_modules/\n")
        (self.main_repo / "app.txt").write_text("base\n")
        (self.main_repo / ".env").write_text("PORT=3000\n")
        self.git(self.main_repo, "add", ".gitignore", "app.txt")
        self.git(self.main_repo, "commit", "-q", "-m", "initial")
        self.git(self.main_repo, "push", "-q", "origin", "development")

    def tearDown(self):
        self.tmpdir.cleanup()

    def git(self, cwd, *args):
        return subprocess.run(
            [GIT, *args], cwd=cwd, env=self.env, check=True, capture_output=True, text=True
        ).stdout.strip()

    def test_merged_worktree_with_disposable_and_identical_ignored_files_is_removed_on_apply(self):
        mod = load_module()
        wt = self.tmp / "wt-merged"
        self.git(self.main_repo, "worktree", "add", "-q", "-b", "feat/merged", str(wt), "development")
        (wt / "feat.txt").write_text("done\n")
        self.git(wt, "add", "feat.txt")
        self.git(wt, "commit", "-q", "-m", "feat")
        head_sha = self.git(wt, "rev-parse", "HEAD")
        self.git(wt, "push", "-q", "origin", "feat/merged")
        self.git(self.main_repo, "merge", "-q", "--no-edit", "feat/merged")
        self.git(self.main_repo, "push", "-q", "origin", "development")

        # Add disposable ignored files and an identical .env
        (wt / ".DS_Store").write_text("junk")
        (wt / "__pycache__").mkdir()
        (wt / "__pycache__" / "mod.cpython-312.pyc").write_bytes(b"123")
        (wt / ".env").write_text("PORT=3000\n")

        plan = mod.plan_cleanup(
            str(self.main_repo),
            base="development",
            merged_prs={"feat/merged": head_sha},
            lsof_runner=lambda: (0, ""),
        )
        wt_item = next(item for item in plan["worktrees"] if item["branch"] == "feat/merged")
        self.assertEqual(wt_item["action"], "remove_worktree")
        br_item = next(item for item in plan["branches"] if item["branch"] == "feat/merged")
        self.assertEqual(br_item["action"], "delete_branch")

        applied = mod.apply_cleanup(str(self.main_repo), plan)
        self.assertTrue(applied["ok"])
        self.assertFalse(wt.exists())

    def test_worktree_kept_when_ignored_file_differs_or_lsof_fails_or_process_active(self):
        mod = load_module()
        wt = self.tmp / "wt-diff-env"
        self.git(self.main_repo, "worktree", "add", "-q", "-b", "feat/diff-env", str(wt), "development")
        (wt / ".env").write_text("PORT=9999\n")

        plan_diff = mod.plan_cleanup(
            str(self.main_repo),
            base="development",
            merged_prs={"feat/diff-env": self.git(wt, "rev-parse", "HEAD")},
            lsof_runner=lambda: (0, ""),
        )
        wt_item = next(item for item in plan_diff["worktrees"] if item["branch"] == "feat/diff-env")
        self.assertEqual(wt_item["action"], "keep")
        self.assertIn(".env", wt_item["reason"])

        # Same .env, but lsof fails (cannot be checked) -> must keep
        (wt / ".env").write_text("PORT=3000\n")
        plan_lsof_err = mod.plan_cleanup(
            str(self.main_repo),
            base="development",
            merged_prs={"feat/diff-env": self.git(wt, "rev-parse", "HEAD")},
            lsof_runner=lambda: (1, ""),
        )
        wt_item2 = next(item for item in plan_lsof_err["worktrees"] if item["branch"] == "feat/diff-env")
        self.assertEqual(wt_item2["action"], "keep")
        self.assertIn("process check", wt_item2["reason"])

    def test_Protected_or_moved_remote_branch_is_kept(self):
        mod = load_module()
        wt = self.tmp / "wt-moved"
        self.git(self.main_repo, "worktree", "add", "-q", "-b", "feat/moved", str(wt), "development")
        (wt / "a.txt").write_text("1\n")
        self.git(wt, "add", "a.txt")
        self.git(wt, "commit", "-q", "-m", "c1")
        merged_sha = self.git(wt, "rev-parse", "HEAD")
        # Push a newer commit after the squash-merged PR headRefOid
        (wt / "a.txt").write_text("2\n")
        self.git(wt, "add", "a.txt")
        self.git(wt, "commit", "-q", "-m", "c2")
        self.git(wt, "push", "-q", "origin", "feat/moved")

        plan = mod.plan_cleanup(
            str(self.main_repo),
            base="development",
            merged_prs={"feat/moved": merged_sha},
            lsof_runner=lambda: (0, ""),
        )
        br_item = next(item for item in plan["branches"] if item["branch"] == "feat/moved")
        self.assertEqual(br_item["action"], "keep")
        self.assertIn("headRefOid", br_item["reason"])


if __name__ == "__main__":
    unittest.main()
