import importlib.util
import os
import pathlib
import shutil
import subprocess
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/harness-audit/scripts/local_instructions.py"
GIT = shutil.which("git")


def load_module():
    spec = importlib.util.spec_from_file_location("local_instructions", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class LocalInstructionsScriptTests(unittest.TestCase):
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
        self.repo = self.tmp / "repo"
        self.repo.mkdir()
        subprocess.run([GIT, "init", "-q", "-b", "main"], cwd=self.repo, env=self.env, check=True)
        subprocess.run(
            [GIT, "commit", "-q", "--allow-empty", "-m", "init"], cwd=self.repo, env=self.env, check=True
        )

    def tearDown(self):
        self.tmpdir.cleanup()

    def test_add_ensures_git_exclude_validates_length_and_secrets_and_prunes_stale_past_50(self):
        mod = load_module()
        # Add first entry: ensures info/exclude ignores HIVE.local.md, _support/secrets/, _support/scripts/
        res = mod.add_entry(
            str(self.repo),
            "- Go is installed in /usr/local/go · 2026-11-16",
        )
        self.assertTrue(res["ok"])
        chk = subprocess.run([GIT, "-C", str(self.repo), "check-ignore", "-q", "HIVE.local.md"])
        self.assertEqual(chk.returncode, 0)

        # Reject secret-like entry
        token = "ghp_" + "A1b2C3d4E5f6G7h8I9j0K1l2M3n4O5p6Q7r8"
        bad_secret = mod.add_entry(str(self.repo), f"- token {token} · 2026-11-16")
        self.assertFalse(bad_secret["ok"])
        self.assertTrue(any("secret" in e for e in bad_secret["errors"]))

        # Reject over 200 chars
        overlong = "- " + ("x" * 200) + " · 2026-11-16"
        bad_len = mod.add_entry(str(self.repo), overlong)
        self.assertFalse(bad_len["ok"])

        # Mark entry stale and prune when exceeding 50 entries
        marked = mod.mark_stale(str(self.repo), "Go is installed", "2026-11-17", "moved to homebrew")
        self.assertTrue(marked["ok"])
        for i in range(1, 51):
            r = mod.add_entry(str(self.repo), f"- fact {i} is active · 2026-11-16")
            self.assertTrue(r["ok"])
        content = (self.repo / "HIVE.local.md").read_text(encoding="utf-8")
        self.assertNotIn("STALE 2026-11-17", content)

    def test_check_validates_permissions_and_accounts_secrets(self):
        mod = load_module()
        mod.ensure_ignore(str(self.repo))
        secrets_dir = self.repo / "_support" / "secrets"
        secrets_dir.mkdir(parents=True)
        os.chmod(secrets_dir, 0o755)  # wrong dir mode
        token = "ghp_" + "A1b2C3d4E5f6G7h8I9j0K1l2M3n4O5p6Q7r8"
        (secrets_dir / "ACCOUNTS.md").write_text(f"| local | admin | a@b.c | {token} | all | ok 2026-11-16 |\n")

        report = mod.check_project_local(str(self.repo))
        self.assertFalse(report["ok"])
        joined = "\n".join(report["blocking"])
        self.assertIn("0700", joined)
        self.assertIn("ACCOUNTS.md", joined)


if __name__ == "__main__":
    unittest.main()
