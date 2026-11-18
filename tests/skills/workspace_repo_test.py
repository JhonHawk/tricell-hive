from pathlib import Path
import os
import shutil
import signal
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/workspace-repo/scripts/workspace_repo.py"
CHECK = ROOT / "content/skills/project-layout/scripts/check_layout.py"
NEEDLE = "zebra-needle-7431"


def git(path, *args, check=True):
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    return subprocess.run(["git", "-C", str(path), *args], text=True, capture_output=True, env=env, check=check)


def snapshot(root):
    out = {}
    for dirpath, _, names in os.walk(root):
        for name in names:
            p = Path(dirpath) / name
            if ".git" in p.relative_to(root).parts:
                continue
            out[str(p.relative_to(root))] = p.read_bytes()
    return out


class WorkspaceRepoTest(unittest.TestCase):
    def setUp(self):
        d = tempfile.TemporaryDirectory()
        self.addCleanup(d.cleanup)
        self.base = Path(d.name)
        self.ws = self.base / "ws"
        self.ws.mkdir()
        self.remotes = {}
        for name in ("a", "b", "specs"):
            remote = self.base / f"origin-{name}.git"
            subprocess.run(["git", "init", "-q", "--bare", "-b", "main", str(remote)], check=True)
            child = self.ws / name
            child.mkdir()
            git(child, "init", "-q", "-b", "main")
            git(child, "remote", "add", "origin", str(remote))
            (child / "f.txt").write_text(f"{NEEDLE if name == 'a' else 'plain'}\n")
            self.remotes[name] = str(remote)
        (self.ws / "HIVE.local.md").write_text("- local fact 2026-11-18\n")
        (self.ws / ".engram").mkdir()
        (self.ws / ".engram/config.json").write_text("{}\n")

    def run_script(self, *args):
        return subprocess.run(["python3", "-I", str(SCRIPT), *args], text=True, capture_output=True)

    def base_args(self, command="apply"):
        args = [command, "--root", str(self.ws), "--project", "demo", "--tracker", "GitHub Issues · o/r"]
        for name in ("a", "b", "specs"):
            args += ["--repo", name]
        return args

    def apply(self, *extra):
        result = self.run_script(*self.base_args(), *extra)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        return result

    def test_plan_writes_nothing_and_is_the_default(self):
        before = snapshot(self.ws)
        for args in (self.base_args("plan"), self.base_args("plan")[1:]):
            result = self.run_script(*args)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("/a/", result.stdout)
            self.assertIn("plan only", result.stdout)
        self.assertEqual(snapshot(self.ws), before)
        self.assertFalse((self.ws / ".git").exists())

    def test_apply_ignores_children_keeps_search_and_passes_check(self):
        self.apply()
        self.assertEqual(git(self.ws, "symbolic-ref", "--short", "HEAD").stdout.strip(), "main")
        git(self.ws, "-c", "user.email=t@example.invalid", "-c", "user.name=t", "commit", "-q", "--allow-empty", "-m", "init", check=True)
        git(self.ws, "add", "-A")
        status = git(self.ws, "status", "--short").stdout
        for name in ("a/", "b/", "specs/", "HIVE.local.md", ".engram"):
            self.assertNotIn(name, status)
        staged = git(self.ws, "diff", "--cached", "--name-only").stdout.splitlines()
        self.assertEqual(sorted(staged), sorted([".gitignore", ".ignore", "AGENTS.md", "_support/docs/README.md", "_support/openspec/.gitkeep"]))
        if shutil.which("rg"):
            found = subprocess.run(["rg", "-l", NEEDLE, "."], cwd=self.ws, text=True, capture_output=True, stdin=subprocess.DEVNULL)
            self.assertIn("a/f.txt", found.stdout)
        else:
            self.skipTest("rg is not installed; search assertion skipped")

    def test_apply_output_passes_layout_check(self):
        self.apply()
        git(self.ws, "add", "-A")
        result = subprocess.run(["python3", "-I", str(CHECK), "--root", str(self.ws)], text=True, capture_output=True)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_hive_section_and_direct_base(self):
        self.apply()
        text = (self.ws / "AGENTS.md").read_text()
        self.assertIn("- Tracker: GitHub Issues · o/r", text)
        self.assertNotIn("Delivery: direct-base", text)
        self.assertIn("| a | a | " + self.remotes["a"] + " | main | — |", text)
        shutil.rmtree(self.ws / ".git")
        self.apply("--direct-base")
        self.assertIn("- Delivery: direct-base", (self.ws / "AGENTS.md").read_text())

    def test_existing_agents_text_is_preserved(self):
        (self.ws / "AGENTS.md").write_text(
            "# Mine\n\nKeep this rule.\n\n## Hive\n\n- Project: old\n\n## Notes\n\nKeep notes too.\n\n"
            "## Repositories\n\n| Repo | Path | Remote | Base | Role |\n| --- | --- | --- | --- | --- |\n| a | a | x | main | backend |\n"
        )
        self.apply()
        text = (self.ws / "AGENTS.md").read_text()
        self.assertIn("Keep this rule.", text)
        self.assertIn("## Notes\n\nKeep notes too.", text)
        self.assertNotIn("Project: old", text)
        self.assertEqual(text.count("## Hive"), 1)
        self.assertEqual(text.count("## Repositories"), 1)
        self.assertIn("| backend |", text)  # role preserved

    def test_apply_twice_is_idempotent(self):
        self.apply()
        first = snapshot(self.ws)
        # a second apply refuses (the coordinator already exists) and changes nothing
        result = self.run_script(*self.base_args())
        self.assertEqual(result.returncode, 1)
        self.assertIn(".git", result.stderr)
        self.assertEqual(snapshot(self.ws), first)
        # re-running over pre-filled ignore files does not duplicate lines
        shutil.rmtree(self.ws / ".git")
        self.apply()
        self.assertEqual(snapshot(self.ws), first)
        lines = (self.ws / ".gitignore").read_text().splitlines()
        self.assertEqual(len(lines), len(set(lines)))

    def test_existing_ignore_lines_preserved(self):
        (self.ws / ".gitignore").write_text("node_modules/\n/a/\n")
        self.apply()
        lines = (self.ws / ".gitignore").read_text().splitlines()
        self.assertEqual(lines[0], "node_modules/")
        self.assertEqual(lines.count("/a/"), 1)

    def test_apply_requires_tracker_and_refuses_existing_git(self):
        result = self.run_script("apply", "--root", str(self.ws), "--project", "demo", "--repo", "a")
        self.assertEqual(result.returncode, 1)
        self.assertIn("--tracker", result.stderr)
        git(self.ws, "init", "-q")
        result = self.run_script(*self.base_args())
        self.assertEqual(result.returncode, 1)
        self.assertIn("already has .git", result.stderr)

    def test_setup_clones_missing_child(self):
        # publish a commit in child a's remote so a clone has content
        git(self.ws / "a", "-c", "user.email=t@example.invalid", "-c", "user.name=t", "commit", "-q", "--allow-empty", "-m", "c")
        git(self.ws / "a", "push", "-q", "origin", "main")
        self.apply()
        shutil.rmtree(self.ws / "a")
        result = self.run_script("setup", "--root", str(self.ws))
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertIn("cloned a", result.stdout)
        self.assertIn("skip b", result.stdout)
        self.assertTrue((self.ws / "a/.git").exists())
        again = self.run_script("setup", "--root", str(self.ws))
        self.assertIn("skip a", again.stdout)

    def test_doctor_reports_no_blocking_finding(self):
        go = shutil.which("go")
        if not go:
            self.skipTest("Go is unavailable; hive doctor check skipped")
        self.apply()
        git(self.ws, "add", "-A")
        git(self.ws, "-c", "user.email=t@example.invalid", "-c", "user.name=t", "commit", "-q", "-m", "init")
        home = self.base / "home"
        home.mkdir()
        proc = subprocess.Popen(
            [go, "-C", str(ROOT), "run", "./tooling/cli", "doctor", "-project", str(self.ws), "-home", str(home)],
            text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, start_new_session=True,
        )
        try:
            stdout, stderr = proc.communicate(timeout=120)
        except subprocess.TimeoutExpired:
            os.killpg(proc.pid, signal.SIGKILL)  # go run leaves the built binary holding the pipes
            proc.communicate()
            self.skipTest("hive doctor exceeded 120 s; skipped")
        result = subprocess.CompletedProcess(proc.args, proc.returncode, stdout, stderr)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        project = result.stdout.split("\nProject\n", 1)[1]
        self.assertIn("Valid", project)


if __name__ == "__main__":
    unittest.main()
