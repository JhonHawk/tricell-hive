from pathlib import Path
import os
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/project-layout/scripts/check_layout.py"


def write(base, rel, text=""):
    path = Path(base) / rel
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text)
    return path


def git(path, *args):
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    return subprocess.run(["git", "-C", str(path), *args], text=True, capture_output=True, env=env, check=True)


def init_repo(path, track=True):
    Path(path).mkdir(parents=True, exist_ok=True)
    git(path, "init", "-q", "-b", "main")
    git(path, "config", "user.email", "t@example.invalid")
    git(path, "config", "user.name", "t")


def commit_all(path, force=()):
    git(path, "add", "-A")
    for item in force:
        git(path, "add", "-f", item)
    git(path, "commit", "-q", "-m", "x")


def clean_repo(path):
    init_repo(path)
    write(path, "_support/docs/README.md", "# Index\n\n- [Guide](guide.md) read first\n")
    write(path, "_support/docs/guide.md", "# Guide\n")
    write(path, ".gitignore", "/HIVE.local.md\n_support/scripts/\n")
    write(path, "src/app.py", "print('hi')\n")
    commit_all(path)


class CheckLayoutTest(unittest.TestCase):
    def run_check(self, *roots, extra=()):
        args = ["python3", "-I", str(SCRIPT)]
        for root in roots:
            args += ["--root", str(root)]
        return subprocess.run([*args, *extra], text=True, capture_output=True)

    def tmp(self):
        d = tempfile.TemporaryDirectory()
        self.addCleanup(d.cleanup)
        return Path(d.name)

    def assertFinding(self, result, rule, severity="fail", fragment=""):
        hits = [l for l in result.stdout.splitlines() if l.startswith(f"{rule} {severity} ") and fragment in l]
        self.assertTrue(hits, f"no {rule} {severity} {fragment!r} in:\n{result.stdout}")
        self.assertIn(" — ", hits[0])

    def test_clean_fixture_exits_zero(self):
        repo = self.tmp() / "r"
        clean_repo(repo)
        result = self.run_check(repo)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertEqual(result.stdout.strip(), "")

    def test_layout_01_unregistered_script(self):
        repo = self.tmp() / "r"
        clean_repo(repo)
        write(repo, "_support/scripts/seed.sh", "echo\n")
        write(repo, "_support/scripts/named.sh", "echo\n")
        write(repo, "HIVE.local.md", "- `_support/scripts/named.sh` seeds the db 2026-11-18\n")
        result = self.run_check(repo)
        self.assertEqual(result.returncode, 1)
        self.assertFinding(result, "LAYOUT-01", fragment="seed.sh")
        self.assertNotIn("named.sh", result.stdout)

    def test_layout_02_tracked_under_untracked_homes(self):
        repo = self.tmp() / "r"
        clean_repo(repo)
        write(repo, "_support/secrets/token.txt", "x\n")
        write(repo, "_support/data/dump.sql", "x\n")
        write(repo, "_support/workspace/.gitkeep", "")
        commit_all(repo, force=["_support/secrets/token.txt", "_support/data/dump.sql", "_support/workspace/.gitkeep"])
        result = self.run_check(repo)
        self.assertEqual(result.returncode, 1)
        self.assertFinding(result, "LAYOUT-02", fragment="token.txt")
        self.assertFinding(result, "LAYOUT-02", fragment="dump.sql")
        self.assertNotIn(".gitkeep", result.stdout)

    def test_layout_03_code_cites_support_and_exemptions(self):
        repo = self.tmp() / "r"
        clean_repo(repo)
        write(repo, "src/bad.py", "x = 1\nopen('_support/data/x')\n")
        write(repo, ".gitignore", "_support/scripts/\n")
        write(repo, ".ignore", "!/a/\n_support/\n")
        write(repo, "AGENTS.md", "see _support/docs\n")
        write(repo, "README.md", "see _support/docs\n")
        write(repo, "ci.yml", "paths: _support/openspec\n")
        write(repo, "scripts/check_specs.sh", "ls _support/openspec/specs _support/docs/README.md\n")
        write(repo, "src/documented.ts", "// see _support/scripts/old.sh, retired\n/* _support/workspace/x */\nexport const a = 1\n")
        commit_all(repo)
        result = self.run_check(repo)
        self.assertEqual(result.returncode, 1)
        self.assertFinding(result, "LAYOUT-03", fragment="bad.py:2")
        self.assertNotIn("check_specs.sh", result.stdout)
        self.assertNotIn("documented.ts", result.stdout)
        lines = [l for l in result.stdout.splitlines() if l.startswith("LAYOUT-03")]
        self.assertEqual(len(lines), 1, lines)

    def test_layout_03_exempt_files_alone_pass(self):
        repo = self.tmp() / "r"
        clean_repo(repo)
        write(repo, ".gitignore", "_support/scripts/\n")
        write(repo, "AGENTS.md", "see _support/docs\n")
        write(repo, "README.md", "see _support/docs\n")
        commit_all(repo)
        result = self.run_check(repo)
        self.assertEqual(result.returncode, 0, result.stdout)

    def test_layout_04_unknown_support_entry_is_review(self):
        repo = self.tmp() / "r"
        clean_repo(repo)
        write(repo, "_support/evidence/a.txt", "x\n")
        result = self.run_check(repo)
        self.assertEqual(result.returncode, 0, result.stdout)
        self.assertFinding(result, "LAYOUT-04", "review", "evidence")

    def test_layout_05_missing_document_and_broken_link(self):
        repo = self.tmp() / "r"
        clean_repo(repo)
        write(repo, "_support/docs/orphan.md", "# Orphan\n")
        write(repo, "_support/docs/README.md", "# Index\n\n- [Guide](guide.md)\n- [Gone](gone.md)\n- [Web](https://example.com/x)\n- [A](#a)\n")
        result = self.run_check(repo)
        self.assertEqual(result.returncode, 1)
        self.assertFinding(result, "LAYOUT-05", fragment="orphan.md")
        self.assertFinding(result, "LAYOUT-05", fragment="gone.md")
        self.assertNotIn("example.com", result.stdout)

    def test_no_index_is_review_only(self):
        d = self.tmp() / "plain"
        d.mkdir()
        result = self.run_check(d)
        self.assertEqual(result.returncode, 0)
        self.assertFinding(result, "LAYOUT-05", "review", "index")

    def test_layout_06_coordinator_table(self):
        base = self.tmp()
        remote = base / "origin.git"
        subprocess.run(["git", "init", "-q", "--bare", str(remote)], check=True)
        coord = base / "coord"
        clean_repo(coord)
        child = coord / "api"
        init_repo(child)
        git(child, "remote", "add", "origin", str(remote))
        table = (
            "## Repositories\n\n| Repo | Path | Remote | Base | Role |\n| --- | --- | --- | --- | --- |\n"
            f"| api | api | {remote} | main | — |\n"
            f"| web | web | {remote} | main | — |\n"
        )
        write(coord, "AGENTS.md", "# A\n\n" + table)
        write(coord, ".gitignore", "_support/scripts/\n")
        result = self.run_check(coord)
        self.assertEqual(result.returncode, 1)
        self.assertFinding(result, "LAYOUT-06", fragment="web")  # missing
        self.assertFinding(result, "LAYOUT-06", fragment="not ignored")
        self.assertFinding(result, "LAYOUT-06", fragment="re-included")
        # wrong remote
        write(coord, "AGENTS.md", "# A\n\n" + table.replace(f"api | {remote}", "api | https://example.invalid/other.git").replace("| web | web |", "| x | web |"))
        result = self.run_check(coord)
        self.assertFinding(result, "LAYOUT-06", fragment="remote origin")

    def test_layout_06_ok_when_ignored_and_reincluded(self):
        base = self.tmp()
        remote = base / "origin.git"
        subprocess.run(["git", "init", "-q", "--bare", str(remote)], check=True)
        coord = base / "coord"
        clean_repo(coord)
        child = coord / "api"
        init_repo(child)
        git(child, "remote", "add", "origin", "https://example.invalid/org/api.git")
        write(coord, "AGENTS.md", f"## Repositories\n\n| Repo | Path | Remote |\n| - | - | - |\n| api | `api` | https://example.invalid/org/api/ |\n")
        write(coord, ".gitignore", "/api/\n_support/scripts/\n")
        write(coord, ".ignore", "!/api/\n")
        result = self.run_check(coord)
        self.assertEqual(result.returncode, 0, result.stdout)

    def test_layout_07_undeclared_docs_and_declared_layout(self):
        repo = self.tmp() / "r"
        init_repo(repo)
        write(repo, "docs/README.md", "# Index\n\n- [A](a.md)\n")
        write(repo, "docs/a.md", "# A\n")
        write(repo, "docs/b.md", "# B\n")
        write(repo, "scripts/tool.sh", "echo\n")
        write(repo, "scripts/tool_test.sh", "echo\n")
        commit_all(repo)
        result = self.run_check(repo)
        self.assertFinding(result, "LAYOUT-07", "review", "docs")
        self.assertEqual(result.returncode, 0, result.stdout)
        write(repo, "docs/README.md", "# Index\n\nLayout: docs at docs/ · tools at scripts/\n\n- [A](a.md)\n")
        result = self.run_check(repo)
        self.assertNotIn("LAYOUT-07", result.stdout)
        self.assertFinding(result, "LAYOUT-05", fragment="b.md")
        self.assertFinding(result, "LAYOUT-05", fragment="tool.sh")
        self.assertNotIn("tool_test.sh", result.stdout)
        self.assertEqual(result.returncode, 1)

    def test_agents_index_line_is_honored(self):
        repo = self.tmp() / "r"
        init_repo(repo)
        write(repo, "AGENTS.md", "# A\n\nIndex: handbook/README.md\n")
        write(repo, "handbook/README.md", "Layout: docs at handbook\n\n- [a](a.md)\n")
        write(repo, "handbook/a.md", "# A\n")
        write(repo, "handbook/b.md", "# B\n")
        result = self.run_check(repo)
        self.assertNotIn("no documentation index", result.stdout)
        self.assertFinding(result, "LAYOUT-05", fragment="b.md")

    def test_json_output(self):
        repo = self.tmp() / "r"
        clean_repo(repo)
        write(repo, "_support/scripts/seed.sh", "x\n")
        result = self.run_check(repo, extra=["--json"])
        import json
        data = json.loads(result.stdout)
        self.assertEqual(data[0]["rule"], "LAYOUT-01")
        self.assertEqual(result.returncode, 1)


if __name__ == "__main__":
    unittest.main()
