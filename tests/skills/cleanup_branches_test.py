import contextlib
import importlib.util
import io
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

    def test_fresh_and_environment_branches_at_base_are_never_scheduled(self):
        mod = load_module()
        wt = self.tmp / "wt-fresh"
        self.git(self.main_repo, "worktree", "add", "-q", "-b", "feat/fresh", str(wt), "development")
        self.git(self.main_repo, "branch", "qa", "development")
        plan = mod.plan_cleanup(
            str(self.main_repo),
            base="development",
            env_branches=["qa"],
            candidates=["qa"],
            lsof_runner=lambda: (0, ""),
        )
        wt_item = next(item for item in plan["worktrees"] if item["branch"] == "feat/fresh")
        self.assertEqual(wt_item["action"], "keep")
        self.assertFalse(any(b["action"] == "delete_branch" for b in plan["branches"]), plan["branches"])

    def test_failed_worktree_removal_keeps_its_local_and_remote_branch(self):
        mod = load_module()
        wt = self.tmp / "wt-locked"
        self.git(self.main_repo, "worktree", "add", "-q", "-b", "feat/locked", str(wt), "development")
        (wt / "l.txt").write_text("1\n")
        self.git(wt, "add", "l.txt")
        self.git(wt, "commit", "-q", "-m", "locked")
        head_sha = self.git(wt, "rev-parse", "HEAD")
        self.git(wt, "push", "-q", "origin", "feat/locked")
        self.git(self.main_repo, "worktree", "lock", str(wt))
        plan = mod.plan_cleanup(
            str(self.main_repo),
            base="development",
            merged_prs={"feat/locked": head_sha},
            lsof_runner=lambda: (0, ""),
        )
        applied = mod.apply_cleanup(str(self.main_repo), plan)
        self.assertFalse(applied["ok"])
        self.assertNotIn("feat/locked", applied["deleted_branches"])
        remote = self.git(self.main_repo, "ls-remote", "--heads", "origin", "feat/locked")
        self.assertTrue(remote, "remote branch must survive a failed worktree removal")


class BaseFastForwardTests(unittest.TestCase):
    """The close brings the main checkout's base up to origin when that is safe."""

    tearDown = CleanupBranchesTests.tearDown
    git = CleanupBranchesTests.git

    def setUp(self):
        CleanupBranchesTests.setUp(self)
        self.other = self.tmp / "other"
        self.git(self.tmp, "clone", "-q", str(self.origin), str(self.other))

    def push_upstream(self, path, text):
        target = self.other / path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(text)
        self.git(self.other, "add", "-f", path)
        self.git(self.other, "commit", "-q", "-m", f"upstream {path}")
        self.git(self.other, "push", "-q", "origin", "development")
        return self.git(self.other, "rev-parse", "HEAD")

    def run_main(self, *args):
        mod = load_module()
        out = io.StringIO()
        with contextlib.redirect_stdout(out):
            code = mod.main(["--repo", str(self.main_repo), "--base", "development", *args])
        return code, json.loads(out.getvalue())

    def head(self):
        return self.git(self.main_repo, "rev-parse", "HEAD")

    def test_base_behind_and_clean_is_fast_forwarded(self):
        before = self.head()
        self.push_upstream("one.txt", "1\n")
        tip = self.push_upstream("two.txt", "2\n")
        code, plan = self.run_main("--apply")
        self.assertEqual(code, 0, plan)
        self.assertEqual(plan["base_update"]["action"], "fast_forward", plan["base_update"])
        self.assertEqual(plan["base_update"]["commits"], 2)
        self.assertEqual(plan["base_update"]["from"], before)
        self.assertEqual(plan["base_update"]["to"], tip)
        self.assertEqual(self.head(), tip)

    def test_base_already_current(self):
        code, plan = self.run_main("--apply")
        self.assertEqual(code, 0, plan)
        self.assertEqual(plan["base_update"]["action"], "current", plan["base_update"])

    def test_incoming_change_to_a_modified_file_keeps_the_base(self):
        before = self.head()
        self.push_upstream("app.txt", "upstream\n")
        (self.main_repo / "app.txt").write_text("local\n")
        code, plan = self.run_main("--apply")
        self.assertEqual(plan["base_update"]["action"], "keep", plan["base_update"])
        self.assertIn("app.txt", plan["base_update"]["paths"])
        self.assertEqual(self.head(), before)
        self.assertEqual((self.main_repo / "app.txt").read_text(), "local\n")

    def test_incoming_file_matching_an_untracked_file_keeps_the_base(self):
        before = self.head()
        self.push_upstream("notes.txt", "upstream\n")
        (self.main_repo / "notes.txt").write_text("mine\n")
        code, plan = self.run_main("--apply")
        self.assertEqual(plan["base_update"]["action"], "keep", plan["base_update"])
        self.assertIn("notes.txt", plan["base_update"]["paths"])
        self.assertEqual(self.head(), before)
        self.assertEqual((self.main_repo / "notes.txt").read_text(), "mine\n")

    def test_incoming_file_that_is_ignored_locally_keeps_the_base(self):
        before = self.head()
        self.push_upstream(".env", "PORT=1\n")
        code, plan = self.run_main("--apply")
        self.assertEqual(plan["base_update"]["action"], "keep", plan["base_update"])
        self.assertIn(".env", plan["base_update"]["paths"])
        self.assertEqual(self.head(), before)
        self.assertEqual((self.main_repo / ".env").read_text(), "PORT=3000\n")

    def test_incoming_file_inside_an_ignored_local_folder_keeps_the_base(self):
        before = self.head()
        self.push_upstream("node_modules/pkg.js", "upstream\n")
        (self.main_repo / "node_modules").mkdir()
        (self.main_repo / "node_modules" / "pkg.js").write_text("local\n")
        code, plan = self.run_main("--apply")
        self.assertEqual(plan["base_update"]["action"], "keep", plan["base_update"])
        self.assertIn("node_modules/pkg.js", plan["base_update"]["paths"])
        self.assertEqual(self.head(), before)
        self.assertEqual((self.main_repo / "node_modules" / "pkg.js").read_text(), "local\n")

    def test_unrelated_local_change_survives_the_fast_forward(self):
        tip = self.push_upstream("one.txt", "1\n")
        (self.main_repo / "app.txt").write_text("local\n")
        code, plan = self.run_main("--apply")
        self.assertEqual(plan["base_update"]["action"], "fast_forward", plan["base_update"])
        self.assertEqual(self.head(), tip)
        self.assertEqual((self.main_repo / "app.txt").read_text(), "local\n")

    def test_local_base_with_own_commits_is_kept(self):
        self.push_upstream("one.txt", "1\n")
        (self.main_repo / "mine.txt").write_text("x\n")
        self.git(self.main_repo, "add", "mine.txt")
        self.git(self.main_repo, "commit", "-q", "-m", "local only")
        before = self.head()
        code, plan = self.run_main("--apply")
        self.assertEqual(plan["base_update"]["action"], "keep", plan["base_update"])
        self.assertIn("not on origin", plan["base_update"]["reason"])
        self.assertEqual(self.head(), before)

    def test_main_checkout_on_another_branch_is_kept(self):
        self.push_upstream("one.txt", "1\n")
        self.git(self.main_repo, "switch", "-q", "-c", "feat/elsewhere")
        code, plan = self.run_main("--apply")
        self.assertEqual(plan["base_update"]["action"], "keep", plan["base_update"])
        self.assertIn("feat/elsewhere", plan["base_update"]["reason"])
        self.assertEqual(self.git(self.main_repo, "branch", "--show-current"), "feat/elsewhere")

    def test_repo_given_as_linked_worktree_moves_the_main_checkout(self):
        wt = self.tmp / "wt-linked"
        self.git(self.main_repo, "worktree", "add", "-q", "-b", "feat/linked", str(wt), "development")
        wt_head = self.git(wt, "rev-parse", "HEAD")
        tip = self.push_upstream("one.txt", "1\n")
        mod = load_module()
        out = io.StringIO()
        with contextlib.redirect_stdout(out):
            mod.main(["--repo", str(wt), "--base", "development", "--apply"])
        plan = json.loads(out.getvalue())
        self.assertEqual(plan["base_update"]["action"], "fast_forward", plan["base_update"])
        self.assertEqual(self.head(), tip)
        self.assertEqual(self.git(wt, "rev-parse", "HEAD"), wt_head)

    def test_branch_merged_only_upstream_is_deleted_after_the_fast_forward(self):
        self.git(self.other, "switch", "-q", "-c", "feat/up")
        (self.other / "up.txt").write_text("u\n")
        self.git(self.other, "add", "up.txt")
        self.git(self.other, "commit", "-q", "-m", "up")
        self.git(self.other, "push", "-q", "origin", "feat/up")
        self.git(self.other, "switch", "-q", "development")
        self.git(self.other, "merge", "-q", "--no-ff", "--no-edit", "feat/up")
        self.git(self.other, "push", "-q", "origin", "development")
        self.git(self.main_repo, "fetch", "-q", "origin", "feat/up:feat/up")
        code, plan = self.run_main("--apply", "--candidate", "feat/up")
        self.assertEqual(plan["base_update"]["action"], "fast_forward", plan["base_update"])
        self.assertIn("feat/up", plan["applied"]["deleted_branches"], plan)

    def test_base_missing_on_origin_is_kept_without_an_error(self):
        mod = load_module()
        out = io.StringIO()
        with contextlib.redirect_stdout(out):
            code = mod.main(["--repo", str(self.main_repo), "--base", "trunk", "--apply"])
        plan = json.loads(out.getvalue())
        self.assertEqual(code, 0, plan)
        self.assertEqual(plan["base_update"]["action"], "keep", plan["base_update"])
        self.assertIn("no origin/trunk", plan["base_update"]["reason"])
        self.assertTrue(plan["applied"]["ok"], plan["applied"])

    def test_dry_run_neither_fetches_nor_moves(self):
        before = self.head()
        self.push_upstream("one.txt", "1\n")
        code, plan = self.run_main()
        self.assertEqual(code, 0, plan)
        self.assertEqual(plan["base_update"]["action"], "current", plan["base_update"])
        self.assertEqual(self.head(), before)
        self.assertEqual(self.git(self.main_repo, "rev-parse", "origin/development"), before)


if __name__ == "__main__":
    unittest.main()
