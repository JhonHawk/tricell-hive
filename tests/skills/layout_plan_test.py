from pathlib import Path
import json
import os
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/project-layout/scripts/plan_layout.py"
FIXED = 1767312000  # 2026-01-02 UTC


def write(base, rel, text="", mode=None):
    path = Path(base) / rel
    path.parent.mkdir(parents=True, exist_ok=True)
    if isinstance(text, bytes):
        path.write_bytes(text)
    else:
        path.write_text(text)
    if mode:
        path.chmod(mode)
    return path


def git(path, *args, check=True):
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    return subprocess.run(["git", "-C", str(path), *args], text=True, capture_output=True, env=env, check=check)


def monorepo(repo, backend="backend", web="web"):
    """A monorepo with the support folder shapes the real carvento run met."""
    repo.mkdir(parents=True)
    git(repo, "init", "-q", "-b", "main")
    git(repo, "config", "user.email", "t@example.invalid")
    git(repo, "config", "user.name", "t")
    write(repo, ".gitignore", "node_modules/\n_support/evidence/*\n_support/reports/\n_support/backups/\n_support/workspace/\n")
    write(repo, "scripts/prepare.sh", "echo prepare\n")
    write(repo, f"apps/{backend}/package.json", json.dumps({"name": f"@x/{backend}", "scripts": {"contracts:check": f"bash ../../_support/scripts/{backend}/contracts-check.sh"}}) + "\n")
    write(repo, f"apps/{backend}/src/limits.ts", "// see _support/contracts/api.md for the limits\nexport const LIMIT = 1\n")
    write(repo, f"apps/{backend}/src/report.ts", "const DIR = '_support/reports/run'\nexport default DIR\n")
    write(repo, f"apps/{web}/package.json", json.dumps({"name": f"@x/{web}", "scripts": {"i18n": f"node ../../_support/scripts/{web}/check-i18n.mts"}}) + "\n")
    write(repo, ".github/workflows/ci.yml",
          "jobs:\n  t:\n    env:\n      TF_ROOT: _support/infrastructure/terraform\n    steps:\n"
          "      - run: _support/scripts/detect-changes.sh \"$base\"\n"
          "      - run: find _support/scripts -name '*.test.sh' -exec bash {} \\;\n")
    write(repo, "lefthook.yml", "pre-push:\n  commands:\n    v:\n      run: ./_support/scripts/verify-push.sh\n")
    write(repo, "_support/docs/README.md", "# Index\n\n- [Runbook](runbook.md) when rolling back\n")
    write(repo, "_support/docs/runbook.md", f"Run `_support/scripts/rollback.sh qa`, then `_support/scripts/{backend}/set-key.sh`.\n")
    write(repo, "site/package.json", "{}\n")
    write(repo, "site/ops.md", "Rollback: `_support/scripts/rollback.sh`\n")
    write(repo, f"_support/scripts/{backend}/contracts-check.sh", "echo check\n", 0o755)
    write(repo, f"_support/scripts/{backend}/set-key.sh", "echo key\n", 0o755)
    write(repo, f"_support/scripts/{web}/check-i18n.mts", "export {}\n")
    write(repo, "_support/scripts/detect-changes.sh", "case \"$f\" in\n  x) echo;;\nesac\r\necho done\r\n", 0o755)
    write(repo, "_support/scripts/detect-changes.test.sh", ". \"$(dirname \"$0\")/release-fixture.lib.sh\"\n")
    write(repo, "_support/scripts/release-fixture.lib.sh", "echo lib\n")
    write(repo, "_support/scripts/deploy-workflows.test.sh", "bash \"$REPO_ROOT/_support/scripts/detect-changes.sh\"\n")
    write(repo, "_support/scripts/verify-push.sh", "echo push\n", 0o755)
    write(repo, "_support/scripts/rollback.sh",
          "HERE=\"$(cd \"$(dirname \"${BASH_SOURCE[0]}\")\" && pwd)\"\ncat \"$HERE/../docs/runbook.md\"\n", 0o755)
    write(repo, "_support/scripts/backfill-users.sh", "echo once\n")
    for name in ("docs-scripts.sha256.sh", "node20-check.sh", "fixture-loader.sh"):
        write(repo, f"_support/scripts/{name}", "echo\n")
    write(repo, "_support/infrastructure/terraform/main.tf", "resource \"x\" \"y\" {}\n")
    write(repo, "_support/infrastructure/terraform/environments/qa/x.tfvars.example", "a = 1\n")
    write(repo, "_support/infrastructure/backend/iam/unused-policy.json", "{}\n")
    write(repo, "_support/aws/role/bootstrap.sh", "aws iam put-role-policy --policy-document file://policy.json\n", 0o755)
    write(repo, "_support/aws/role/policy.json", "{}\n")
    write(repo, "_support/aws/role/README.md", "# role\n")
    write(repo, "_support/contracts/api.md", "# API\n")
    write(repo, "openspec/changes/archive/2026-01-01-x/tasks.md", "- ran _support/scripts/detect-changes.sh\n")
    write(repo, f"apps/{backend}/data.bin", b"\xff\xfe _support/scripts/rollback.sh\n")
    git(repo, "add", "-A")
    git(repo, "commit", "-q", "-m", "fixture")
    # ignored material
    write(repo, "_support/evidence/2026-01-02-abc-1-x/shot.png", b"png")
    write(repo, "_support/evidence/ABC-2-2026-03-04-y/log.txt", "log\n")
    write(repo, "_support/reports/run/out.txt", "out\n")
    write(repo, "_support/backups/db.dump", "dump\n")
    for path in (repo / "_support/backups/db.dump",):
        os.utime(path, (FIXED, FIXED))
    return repo


class PlanLayoutTest(unittest.TestCase):
    def tmp(self):
        d = tempfile.TemporaryDirectory()
        self.addCleanup(d.cleanup)
        return Path(d.name)

    def plan(self, *roots, extra=(), check=True):
        args = ["python3", "-I", str(SCRIPT), "plan", "--json"]
        for root in roots:
            args += ["--root", str(root)]
        result = subprocess.run([*args, *extra], text=True, capture_output=True)
        if check:
            self.assertEqual(result.returncode, 0, result.stderr)
        return result

    def plan_json(self, *roots, extra=()):
        return json.loads(self.plan(*roots, extra=extra).stdout)

    @staticmethod
    def dest(plan, source):
        rows = [r for r in plan["rows"] if r["source"] == source]
        assert rows, f"no row for {source}"
        return rows[0]["destination"], rows[0]["rule"]

    def test_destinations_follow_the_rule_table(self):
        repo = monorepo(self.tmp() / "shop")
        plan = self.plan_json(repo)
        expected = {
            "_support/scripts/backend/contracts-check.sh": ("apps/backend/scripts/contracts-check.sh", "R4"),
            "_support/scripts/backend/set-key.sh": ("apps/backend/scripts/set-key.sh", "R4"),
            "_support/scripts/web/check-i18n.mts": ("apps/web/scripts/check-i18n.mts", "R4"),
            "_support/scripts/detect-changes.sh": ("scripts/ci/detect-changes.sh", "R5"),
            "_support/scripts/detect-changes.test.sh": ("scripts/ci/detect-changes.test.sh", "R1"),
            "_support/scripts/deploy-workflows.test.sh": ("scripts/ci/deploy-workflows.test.sh", "R3"),
            "_support/scripts/release-fixture.lib.sh": ("scripts/ci/release-fixture.lib.sh", "R3"),
            "_support/scripts/verify-push.sh": ("scripts/hooks/verify-push.sh", "R6"),
            "_support/scripts/rollback.sh": ("scripts/ops/rollback.sh", "R7"),
            "_support/infrastructure/terraform/main.tf": ("infra/terraform/main.tf", "R2"),
            "_support/infrastructure/terraform/environments/qa/x.tfvars.example": ("infra/terraform/environments/qa/x.tfvars.example", "R2"),
            "_support/aws/role/bootstrap.sh": ("infra/aws/role/bootstrap.sh", "R2"),
            "_support/aws/role/policy.json": ("infra/aws/role/policy.json", "R2"),
            "_support/aws/role/README.md": ("infra/aws/role/README.md", "R2"),
            "_support/contracts": ("_support/docs/contracts", "F5"),
            "_support/evidence/2026-01-02-abc-1-x": ("_support/workspace/2026-01-02-abc-1-x/evidence", "F3"),
            "_support/evidence/ABC-2-2026-03-04-y": ("_support/workspace/2026-03-04-abc-2-y/evidence", "F3"),
            "_support/backups": ("_support/data/2026-01-02-backups", "F4"),
        }
        for source, want in expected.items():
            self.assertEqual(self.dest(plan, source), want, source)
        destination, rule = self.dest(plan, "_support/scripts/backfill-users.sh")
        self.assertEqual(rule, "R9")
        self.assertRegex(destination, r"^_support/sessions/\d{4}-\d{2}-\d{2}-backfill-users/backfill-users\.sh$")
        reviews = {(r["path"], r["reason"]) for r in plan["reviews"]}
        for name in ("docs-scripts.sha256.sh", "node20-check.sh", "fixture-loader.sh"):
            self.assertIn((f"_support/scripts/{name}", "R10"), reviews)
        self.assertIn(("_support/infrastructure/backend/iam/unused-policy.json", "R2"), reviews)
        self.assertIn(("_support/reports", "F2"), reviews)
        self.assertIn(("apps/backend/data.bin", "unreadable"), reviews)

    def test_rewrites_keep_the_form_and_spare_history(self):
        repo = monorepo(self.tmp() / "shop")
        plan = self.plan_json(repo)
        pairs = {(r["file"], r["old"], r["new"]) for r in plan["rewrites"]}
        self.assertIn(("apps/backend/package.json", "../../_support/scripts/backend/contracts-check.sh", "scripts/contracts-check.sh"), pairs)
        self.assertIn(("apps/web/package.json", "../../_support/scripts/web/check-i18n.mts", "scripts/check-i18n.mts"), pairs)
        self.assertIn((".github/workflows/ci.yml", "_support/scripts/detect-changes.sh", "scripts/ci/detect-changes.sh"), pairs)
        self.assertIn((".github/workflows/ci.yml", "_support/infrastructure/terraform", "infra/terraform"), pairs)
        self.assertIn(("lefthook.yml", "./_support/scripts/verify-push.sh", "./scripts/hooks/verify-push.sh"), pairs)
        self.assertIn(("_support/scripts/deploy-workflows.test.sh", "$REPO_ROOT/_support/scripts/detect-changes.sh", "$REPO_ROOT/scripts/ci/detect-changes.sh"), pairs)
        self.assertIn(("apps/backend/src/limits.ts", "_support/contracts/api.md", "_support/docs/contracts/api.md"), pairs)
        self.assertFalse([p for p in pairs if p[0].startswith("openspec/changes/archive/")], pairs)
        reviews = {(r["path"], r["reason"]) for r in plan["reviews"]}
        self.assertIn((".github/workflows/ci.yml:7", "prefix"), reviews)  # the test discovery spans several destinations
        self.assertIn(("_support/scripts/rollback.sh:2", "internal path"), reviews)

    def test_the_same_tree_gives_the_same_plan_and_names_do_not_matter(self):
        base = self.tmp()
        first = monorepo(base / "shop")
        second = monorepo(base / "store")
        one = self.plan(first).stdout
        self.assertEqual(one, self.plan(first).stdout)

        def shape(text, repo):
            plan = json.loads(text)
            return (
                sorted((r["source"], r["rule"], r["destination"]) for r in plan["rows"]),
                sorted((r["file"], r["old"], r["new"]) for r in plan["rewrites"]),
            )
        self.assertEqual(shape(one, first), shape(self.plan(second).stdout, second))

    def test_existing_homes_are_reused(self):
        repo = monorepo(self.tmp() / "shop")
        git(repo, "mv", "scripts", "tools")
        write(repo, "tools/pipelines/.keep", "")
        write(repo, "infrastructure/README.md", "# infra\n")
        plan = self.plan_json(repo)
        self.assertEqual(self.dest(plan, "_support/scripts/detect-changes.sh")[0], "tools/pipelines/detect-changes.sh")
        self.assertEqual(self.dest(plan, "_support/scripts/verify-push.sh")[0], "tools/hooks/verify-push.sh")
        self.assertEqual(self.dest(plan, "_support/infrastructure/terraform/main.tf")[0], "infrastructure/terraform/main.tf")

    def test_overrides_and_home_lines(self):
        repo = monorepo(self.tmp() / "shop")
        plan = self.plan_json(repo, extra=["--override", "_support/scripts/rollback.sh=scripts/rollback.sh", "--home", "infra=platform"])
        row = [r for r in plan["rows"] if r["source"] == "_support/scripts/rollback.sh"][0]
        self.assertEqual((row["destination"], row["override"]), ("scripts/rollback.sh", True))
        self.assertEqual(self.dest(plan, "_support/infrastructure/terraform/main.tf")[0], "platform/terraform/main.tf")
        self.assertEqual([l["line"] for l in plan["layout_lines"]], ["Layout: infra at platform"])
        plain = self.plan_json(repo)
        self.assertEqual(plain["layout_lines"], [])

    def test_workspace_root_without_git_places_no_code(self):
        base = self.tmp() / "ws"
        write(base, "_support/scripts/seed-abc-12.cjs", "x\n")
        write(base, "_support/scripts/helper.sh", "x\n")
        write(base, "AGENTS.md", "run _support/scripts/helper.sh\n")
        plan = self.plan_json(base)
        self.assertEqual(self.dest(plan, "_support/scripts/seed-abc-12.cjs")[1], "R9")
        self.assertIsNone(self.dest(plan, "_support/scripts/helper.sh")[0])

    def test_plan_changes_nothing(self):
        repo = monorepo(self.tmp() / "shop")
        before = git(repo, "status", "--porcelain", "--ignored").stdout
        listing = sorted(str(p) for p in repo.rglob("*") if ".git" not in p.parts)
        self.plan(repo)
        self.assertEqual(git(repo, "status", "--porcelain", "--ignored").stdout, before)
        self.assertEqual(sorted(str(p) for p in repo.rglob("*") if ".git" not in p.parts), listing)


if __name__ == "__main__":
    unittest.main()
