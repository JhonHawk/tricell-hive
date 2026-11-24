from pathlib import Path
import argparse
import importlib.util
import json
import os
import shutil
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
    """A monorepo with the support folder shapes the real project-a run met."""
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
    write(repo, "lefthook.yml", b"pre-push:\r\n  commands:\r\n    v:\r\n      run: ./_support/scripts/verify-push.sh\r\n")
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
    write(repo, "_support/scripts/deploy-workflows.test.sh", "bash \"$REPO_ROOT/_support/scripts/detect-changes.sh\"\n", 0o755)
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
    write(repo, "_support/scripts/backfill-orders.sh", "echo once, never committed\n")
    write(repo, "_support/evidence/2026-01-02-abc-1-x/shot.png", b"png")
    write(repo, "_support/evidence/ABC-2-2026-03-04-y/log.txt", "log\n")
    write(repo, "_support/reports/run/out.txt", "out\n")
    write(repo, "_support/backups/db.dump", "dump\n")
    for path in (repo / "_support/backups/db.dump",):
        os.utime(path, (FIXED, FIXED))
    return repo


class Fixture:
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



class PlanLayoutTest(Fixture, unittest.TestCase):
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
        destination, rule = self.dest(plan, "_support/scripts/backfill-orders.sh")
        self.assertEqual(rule, "R9")
        self.assertRegex(destination, r"^_support/workspace/\d{4}-\d{2}-\d{2}-backfill-orders/backfill-orders\.sh$")
        tracked = [r for r in plan["reviews"] if r["path"] == "_support/scripts/backfill-users.sh"]
        self.assertEqual(len(tracked), 1, "a tracked one-off script cannot move into the ignored workspace")
        self.assertEqual(tracked[0]["reason"], "R9")
        self.assertIn("_support/workspace/", tracked[0]["detail"])
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

    def test_kept_rows_stay_where_they_are(self):
        repo = monorepo(self.tmp() / "shop")
        keeps = ["_support/scripts/detect-changes.sh", "_support/infrastructure", "_support/evidence/2026-01-02-abc-1-x"]
        plan = self.plan_json(repo, extra=[a for k in keeps for a in ("--keep", k)])
        kept = {r["source"] for r in plan["rows"] if any(n.startswith("kept") for n in r["notes"])}
        self.assertEqual(kept, {
            "_support/scripts/detect-changes.sh", "_support/scripts/detect-changes.test.sh",
            "_support/scripts/release-fixture.lib.sh", "_support/scripts/deploy-workflows.test.sh",
            "_support/infrastructure/terraform/main.tf", "_support/infrastructure/terraform/environments/qa/x.tfvars.example",
            "_support/infrastructure/backend/iam/unused-policy.json", "_support/evidence/2026-01-02-abc-1-x",
        })
        self.assertTrue(all(r["destination"] is None for r in plan["rows"] if r["source"] in kept))
        self.assertFalse({r["path"] for r in plan["reviews"]} & kept)
        self.assertFalse([r for r in plan["reviews"] if "_support/infrastructure" in r["detail"] or "release-fixture" in r["path"]], plan["reviews"])
        self.assertFalse([r for r in plan["rewrites"] if "detect-changes.sh" in r["old"]])
        self.assertEqual(self.dest(plan, "_support/scripts/verify-push.sh")[0], "scripts/hooks/verify-push.sh")
        unknown = self.plan(repo, extra=["--keep", "_support/nothing"], check=False)
        self.assertNotEqual(unknown.returncode, 0)
        self.assertIn("_support/nothing", unknown.stderr)

    def test_a_kept_folder_asks_nothing_and_a_kept_test_flags_its_moving_subject(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, ".github/workflows/deploy.yml", "steps:\n  - run: _support/scripts/deploy.sh\n")
        write(repo, "_support/scripts/deploy.sh", "echo\n")
        write(repo, "_support/scripts/deploy.test.sh", "bash \"$(dirname \"$0\")/deploy.sh\"\n")
        write(repo, "_support/tools/run.py", "print()\n")
        write(repo, "_support/tools/notes.md", "# notes\n")
        commit(repo)
        plan = self.plan_json(repo, extra=["--keep", "_support/tools", "--keep", "_support/scripts/deploy.test.sh"])
        self.assertFalse([r for r in plan["reviews"] if r["path"].startswith("_support/tools")], plan["reviews"])
        self.assertIn("_support/scripts/deploy.test.sh", {r["path"] for r in plan["reviews"]})
        both = self.plan(repo, extra=["--keep", "_support/scripts/deploy.sh", "--override", "_support/scripts/deploy.sh=x/deploy.sh"], check=False)
        self.assertNotEqual(both.returncode, 0)
        self.assertIn("--keep and --override", both.stderr)

    def test_what_follows_a_kept_file_yields_to_an_override_and_needs_every_user_kept(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, ".github/workflows/deploy.yml", "steps:\n  - run: _support/scripts/deploy.sh\n")
        write(repo, "_support/scripts/deploy.sh", ". \"$(dirname \"$0\")/common.lib.sh\"\n")
        write(repo, "_support/scripts/common.lib.sh", "echo\n")
        write(repo, "_support/scripts/deploy.test.sh", "echo\n")
        commit(repo)
        keep = ["--keep", "_support/scripts/deploy.sh"]
        plan = self.plan_json(repo, extra=keep + ["--override", "_support/scripts/common.lib.sh=scripts/lib/common.lib.sh",
                                                  "--override", "_support/scripts/deploy.test.sh=tests/deploy.test.sh"])
        self.assertEqual(self.dest(plan, "_support/scripts/common.lib.sh")[0], "scripts/lib/common.lib.sh")
        self.assertEqual(self.dest(plan, "_support/scripts/deploy.test.sh")[0], "tests/deploy.test.sh")
        write(repo, "_support/scripts/other.sh", ". \"$(dirname \"$0\")/common.lib.sh\"\n")
        commit(repo)
        plan = self.plan_json(repo, extra=keep)
        lib = [r for r in plan["rows"] if r["source"] == "_support/scripts/common.lib.sh"][0]
        self.assertFalse([n for n in lib["notes"] if n.startswith("kept")], lib)

    def test_a_moving_file_keeps_its_path_to_a_kept_folder(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, "_support/scripts/a.sh", 'ls "$DIR"/../../_support/tools/\n')
        write(repo, "_support/tools/run.py", "print()\n")
        commit(repo)
        plan = self.plan_json(repo, extra=["--keep", "_support/tools", "--override", "_support/scripts/a.sh=a.sh"])
        news = [r["new"] for r in plan["rewrites"] if r["file"] == "_support/scripts/a.sh"]
        self.assertEqual(news, ['"$DIR"/_support/tools/'], plan["rewrites"])

    def test_workspace_root_without_git_places_no_code(self):
        base = self.tmp() / "ws"
        write(base, "_support/scripts/seed-abc-12.cjs", "x\n")
        write(base, "_support/scripts/helper.sh", "x\n")
        write(base, "AGENTS.md", "run _support/scripts/helper.sh\n")
        plan = self.plan_json(base)
        self.assertEqual(self.dest(plan, "_support/scripts/seed-abc-12.cjs")[1], "R9")
        self.assertIsNone(self.dest(plan, "_support/scripts/helper.sh")[0])

    def test_runtime_citations_count_only_for_the_folder_they_reach(self):
        ws = self.tmp() / "ws"
        write(ws, "_support/evidence/2026-01-02-abc-1-x/a.txt", "x\n")
        repo = monorepo(ws / "shop")  # its report.ts reads its own _support/reports, not the workspace's
        write(ws, "_support/reports/2026-01-05-run/out.txt", "x\n")
        plan = self.plan_json(ws, repo)
        workspace = [r for r in plan["rows"] if r["root"] == 0]
        self.assertEqual({(r["source"], r["rule"]) for r in workspace},
                         {("_support/evidence/2026-01-02-abc-1-x", "F3"), ("_support/reports/2026-01-05-run", "F3")})
        self.assertIn((1, "_support/reports", "F2"), {(r["root"], r["path"], r["reason"]) for r in plan["reviews"]})

    def test_plan_changes_nothing(self):
        repo = monorepo(self.tmp() / "shop")
        before = git(repo, "status", "--porcelain", "--ignored").stdout
        listing = sorted(str(p) for p in repo.rglob("*") if ".git" not in p.parts)
        self.plan(repo)
        self.assertEqual(git(repo, "status", "--porcelain", "--ignored").stdout, before)
        self.assertEqual(sorted(str(p) for p in repo.rglob("*") if ".git" not in p.parts), listing)



class ApplyLayoutTest(Fixture, unittest.TestCase):
    def run_cmd(self, command, plan_file):
        return subprocess.run(["python3", "-I", str(SCRIPT), command, "--plan", str(plan_file)], text=True, capture_output=True)

    def planned(self, repo, *extra):
        plan_file = repo.parent / "plan.json"
        self.plan(repo, extra=["--out", str(plan_file), *extra])
        git(repo, "tag", "archive/test")  # rows that leave Git or are deleted need it
        return plan_file

    def state(self, repo):
        return (git(repo, "status", "--porcelain", "--ignored").stdout, git(repo, "diff").stdout, git(repo, "diff", "--cached").stdout)

    def test_apply_moves_rewrites_and_never_commits(self):
        repo = monorepo(self.tmp() / "shop")
        head = git(repo, "rev-parse", "HEAD").stdout
        plan_file = self.planned(repo)
        result = self.run_cmd("apply", plan_file)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(git(repo, "rev-parse", "HEAD").stdout, head)
        self.assertTrue((repo / "apps/backend/scripts/contracts-check.sh").exists())
        self.assertIn("bash scripts/contracts-check.sh", (repo / "apps/backend/package.json").read_text())
        self.assertTrue((repo / "_support/workspace/2026-01-02-abc-1-x/evidence/shot.png").exists())
        self.assertFalse((repo / "_support/evidence/2026-01-02-abc-1-x").exists())
        self.assertIn(b"./scripts/hooks/verify-push.sh\r\n", (repo / "lefthook.yml").read_bytes())
        self.assertTrue(os.access(repo / "scripts/ci/deploy-workflows.test.sh", os.X_OK))
        self.assertIn("$REPO_ROOT/scripts/ci/detect-changes.sh", (repo / "scripts/ci/deploy-workflows.test.sh").read_text())
        self.assertIn("_support/scripts/detect-changes.sh", (repo / "openspec/changes/archive/2026-01-01-x/tasks.md").read_text())
        for old in ("_support/scripts/detect-changes.sh", "_support/scripts/verify-push.sh", "_support/infrastructure/terraform"):
            hits = git(repo, "grep", "-n", "-F", old, "--", ".", ":!openspec/changes/archive", check=False).stdout
            self.assertEqual(hits, "", old)

    def test_verify_wants_index_lines_and_an_index(self):
        repo = monorepo(self.tmp() / "shop")
        plan_file = self.planned(repo)
        self.assertEqual(self.run_cmd("apply", plan_file).returncode, 0)
        result = self.run_cmd("verify", plan_file)
        self.assertEqual(result.returncode, 1)
        self.assertIn("no line for scripts/ci/detect-changes.sh", result.stdout)
        index = repo / "_support/docs/README.md"
        names = ["contracts-check.sh", "set-key.sh", "check-i18n.mts", "detect-changes.sh", "verify-push.sh", "rollback.sh"]
        index.write_text(index.read_text() + "".join(f"- `{n}` — tool\n" for n in names))
        result = self.run_cmd("verify", plan_file)
        self.assertEqual(result.returncode, 0, result.stdout)
        index.unlink()
        result = self.run_cmd("verify", plan_file)
        self.assertIn("no index", result.stdout)

    def test_apply_refuses_a_stale_plan_or_a_taken_destination(self):
        repo = monorepo(self.tmp() / "shop")
        plan_file = self.planned(repo)
        (repo / "_support/scripts/verify-push.sh").write_text("echo changed\n")
        before = self.state(repo)
        result = self.run_cmd("apply", plan_file)
        self.assertEqual(result.returncode, 1)
        self.assertIn("changed since the plan", result.stderr)
        self.assertEqual(self.state(repo), before)
        repo2 = monorepo(self.tmp() / "store")
        plan_file = self.planned(repo2)
        write(repo2, "scripts/ci/detect-changes.sh", "taken\n")
        result = self.run_cmd("apply", plan_file)
        self.assertEqual(result.returncode, 1)
        self.assertIn("destination exists", result.stderr)

    TOOLS = ["contracts-check.sh", "set-key.sh", "check-i18n.mts", "detect-changes.sh", "verify-push.sh", "rollback.sh"]

    def index_tools(self, repo):
        index = repo / "_support/docs/README.md"
        index.write_text(index.read_text() + "".join(f"- `{n}` — tool\n" for n in self.TOOLS))

    def test_a_move_out_of_a_docker_exclusion_is_reviewed_until_the_file_names_it(self):
        repo = monorepo(self.tmp() / "shop")
        write(repo, ".dockerignore", "_support\nnode_modules\n")
        commit(repo)
        plan_file = self.planned(repo)
        plan = json.loads(plan_file.read_text())
        builds = [r for r in plan["reviews"] if r["reason"] == "build contents"]
        self.assertTrue(any(r["path"] == ".dockerignore" and "infra/," in r["detail"]
                            and "now included" in r["detail"] for r in builds), builds)
        self.assertEqual(self.run_cmd("apply", plan_file).returncode, 0)
        self.index_tools(repo)
        result = self.run_cmd("verify", plan_file)
        self.assertEqual(result.returncode, 1)
        self.assertIn(".dockerignore no longer excludes infra/terraform/main.tf", result.stdout)
        (repo / ".dockerignore").write_text("_support\nnode_modules\ninfra/\nscripts/\n!apps/backend/scripts\n!apps/web/scripts\n")
        result = self.run_cmd("verify", plan_file)
        self.assertEqual(result.returncode, 0, result.stdout)

    def test_a_move_into_a_package_exclusion_is_reviewed(self):
        repo = monorepo(self.tmp() / "shop")
        write(repo, ".npmignore", "/scripts/\n")
        commit(repo)
        plan = json.loads(self.planned(repo).read_text())
        self.assertTrue(any(r["reason"] == "build contents" and r["path"] == ".npmignore"
                            and "now excluded" in r["detail"] and "scripts/ci/" in r["detail"]
                            for r in plan["reviews"]), plan["reviews"])

    def test_a_dockerfile_ignore_and_a_gcloud_include_read_from_where_their_tools_do(self):
        repo = monorepo(self.tmp() / "shop")
        write(repo, "apps/backend/Dockerfile.dockerignore", "_support\n")  # built with the repository root as context
        write(repo, ".gitignore", (repo / ".gitignore").read_text() + "/_support/scripts/rollback.sh\n")
        write(repo, ".gcloudignore", "#!include:.gitignore\n")
        commit(repo)
        reviews = [r for r in json.loads(self.planned(repo).read_text())["reviews"] if r["reason"] == "build contents"]
        self.assertTrue(any(r["path"] == "apps/backend/Dockerfile.dockerignore" and "infra/," in r["detail"]
                            for r in reviews), reviews)
        self.assertTrue(any(r["path"] == ".gcloudignore" and "now included" in r["detail"] and "scripts/ops/" in r["detail"]
                            for r in reviews), reviews)

    def test_ignored_files_left_in_a_moved_folder_are_reported(self):
        repo = monorepo(self.tmp() / "shop")
        write(repo, ".gitignore", (repo / ".gitignore").read_text() + ".terraform/\n")
        commit(repo)
        write(repo, "_support/infrastructure/terraform/environments/qa/.terraform/terraform.tfstate", "{}\n")
        # ignored, and the plan moves them: one beside tracked files, one alone in its folder
        write(repo, "_support/infrastructure/terraform/prod.auto.tfvars", "secret = 1\n")
        write(repo, "_support/infrastructure/terraform/secrets/qa.auto.tfvars", "secret = 2\n")
        write(repo, ".git/info/exclude", "*.auto.tfvars\n")
        plan_file = self.planned(repo)
        plan = json.loads(plan_file.read_text())
        stay = [r for r in plan["reviews"] if r["reason"] == "ignored files stay"]
        self.assertTrue(any(r["path"] == "_support/infrastructure/terraform" and ".terraform" in r["detail"]
                            for r in stay), plan["reviews"])
        for name, listed in (("prod.auto.tfvars", "prod.auto.tfvars"), ("secrets/qa.auto.tfvars", "secrets/")):
            if any(r["source"] == f"_support/infrastructure/terraform/{name}" and r["destination"] for r in plan["rows"]):
                self.assertFalse([r for r in stay if listed in r["detail"]], stay)
        self.assertEqual(self.run_cmd("apply", plan_file).returncode, 0)
        self.index_tools(repo)
        result = self.run_cmd("verify", plan_file)
        self.assertEqual(result.returncode, 0, result.stdout)
        self.assertIn("note: _support/infrastructure/terraform still holds 1 entry Git ignores", result.stdout)

    def test_a_reference_the_user_keeps_stays_and_passes_verify(self):
        repo = monorepo(self.tmp() / "shop")
        write(repo, ".github/scripts/select.mjs", "const OLD = '_support/scripts/verify-push.sh'\n")
        commit(repo)
        plan_file = self.planned(repo, "--keep-ref", ".github/scripts/select.mjs:1")
        plan = json.loads(plan_file.read_text())
        self.assertFalse([r for r in plan["rewrites"] if r["file"] == ".github/scripts/select.mjs"])
        self.assertEqual(self.run_cmd("apply", plan_file).returncode, 0)
        self.index_tools(repo)
        selector = repo / ".github/scripts/select.mjs"
        selector.write_text("// a line added after the plan\n" + selector.read_text())
        result = self.run_cmd("verify", plan_file)
        self.assertEqual(result.returncode, 0, result.stdout)
        self.assertIn("_support/scripts/verify-push.sh", selector.read_text())
        for spec in ("nothing.txt", ""):
            result = self.plan(repo, extra=["--keep-ref", spec], check=False)
            self.assertEqual(result.returncode, 1)
            self.assertIn("--keep-ref", result.stderr)
            self.assertNotIn("Traceback", result.stderr)

    def test_undo_restores_the_tree_and_keeps_other_work(self):
        repo = monorepo(self.tmp() / "shop")
        ci = repo / ".github/workflows/ci.yml"
        ci.write_text(ci.read_text() + "      - run: echo other session\n")  # another work's uncommitted edit
        before = self.state(repo)
        plan_file = self.planned(repo)
        self.assertEqual(self.run_cmd("apply", plan_file).returncode, 0)
        self.assertNotEqual(self.state(repo), before)
        result = self.run_cmd("undo", plan_file)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.state(repo), before)
        self.assertIn("echo other session", ci.read_text())


def small_repo(repo):
    repo.mkdir(parents=True)
    git(repo, "init", "-q", "-b", "main")
    git(repo, "config", "user.email", "t@example.invalid")
    git(repo, "config", "user.name", "t")
    write(repo, ".gitignore", "_support/backups/\n_support/data/\n_support/workspace/\n")
    write(repo, ".github/workflows/ci.yml", "steps:\n  - run: find _support/scripts -name '*.test.sh'\n")
    return repo


def commit(repo):
    git(repo, "add", "-A")
    git(repo, "commit", "-q", "-m", "x")


class ReviewFindingsTest(Fixture, unittest.TestCase):
    def run_cmd(self, command, plan_file):
        return subprocess.run(["python3", "-I", str(SCRIPT), command, "--plan", str(plan_file)], text=True, capture_output=True)

    def rewrites(self, plan):
        return {(r["file"], r["old"], r["new"]) for r in plan["rewrites"]}

    def test_references_after_a_substitution_and_relative_markdown_links(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, "run.sh", 'bash "$(git rev-parse --show-toplevel)/_support/scripts/deploy.sh"\nbash "$DIR"/_support/scripts/deploy.sh\n')
        write(repo, ".github/workflows/deploy.yml", "steps:\n  - run: _support/scripts/deploy.sh\n")
        write(repo, "_support/scripts/deploy.sh", "echo\n")
        write(repo, "_support/infrastructure/main.tf", "x\n")
        write(repo, "_support/docs/ops.md", "[main](../infrastructure/main.tf)\n")
        commit(repo)
        plan = self.plan_json(repo)
        pairs = self.rewrites(plan)
        self.assertIn(("run.sh", "$(git rev-parse --show-toplevel)/_support/scripts/deploy.sh", "$(git rev-parse --show-toplevel)/scripts/ci/deploy.sh"), pairs)
        self.assertIn(("run.sh", '"$DIR"/_support/scripts/deploy.sh', '"$DIR"/scripts/ci/deploy.sh'), pairs)
        self.assertIn(("_support/docs/ops.md", "../infrastructure/main.tf", "../../infra/main.tf"), pairs)

    def test_a_code_folder_named_sessions_is_not_history(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, "src/sessions/start.sh", "bash _support/scripts/deploy.sh\n")
        write(repo, "_support/scripts/deploy.sh", "echo\n")
        commit(repo)
        self.assertIn(("src/sessions/start.sh", "_support/scripts/deploy.sh", "scripts/deploy.sh"), self.rewrites(self.plan_json(repo)))

    def test_a_failed_apply_leaves_an_undo_record(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, ".github/workflows/deploy.yml", "steps:\n  - run: _support/scripts/deploy.sh\n")
        write(repo, "_support/scripts/deploy.sh", "echo\n")
        write(repo, "scripts/keep.sh", "echo\n")
        commit(repo)
        before = (git(repo, "status", "--porcelain").stdout, git(repo, "diff").stdout)
        plan_file = repo.parent / "plan.json"
        self.plan(repo, extra=["--out", str(plan_file)])
        (repo / "scripts").chmod(0o555)  # the move into scripts/ci/ fails after the rewrites
        self.addCleanup((repo / "scripts").chmod, 0o755)
        result = self.run_cmd("apply", plan_file)
        self.assertEqual(result.returncode, 1)
        self.assertIn("undo", result.stderr)
        self.assertTrue(Path(str(plan_file) + ".undo.json").exists())
        (repo / "scripts").chmod(0o755)
        self.assertEqual(self.run_cmd("undo", plan_file).returncode, 0)
        self.assertEqual((git(repo, "status", "--porcelain").stdout, git(repo, "diff").stdout), before)

    def test_tracked_folders_rename_and_never_drop_out_of_git(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, "_support/contracts/api.md", "[spec](../openspec/x.md)\n")
        write(repo, "_support/openspec/x.md", "# x\n")
        write(repo, "_support/backups/db.dump", "dump\n")
        commit(repo)
        git(repo, "add", "-f", "_support/backups/db.dump")
        git(repo, "commit", "-q", "-m", "tracked backup")
        plan_file = repo.parent / "plan.json"
        plan = json.loads(self.plan(repo, extra=["--out", str(plan_file)]).stdout)
        git(repo, "tag", "archive/test")
        self.assertIn(("_support/contracts/api.md", "../openspec/x.md", "../../openspec/x.md"), self.rewrites(plan))
        backups = [r for r in plan["rows"] if r["source"] == "_support/backups"][0]
        self.assertTrue(backups["destination"].startswith("_support/data/"), backups)
        self.assertTrue(backups["untrack"])
        self.assertEqual(self.run_cmd("apply", plan_file).returncode, 0)
        self.assertEqual(git(repo, "ls-files", "_support/backups").stdout, "")
        status = git(repo, "status", "--porcelain").stdout
        self.assertRegex(status, r"(?m)^RM? _support/contracts/api\.md -> _support/docs/contracts/api\.md$")
        self.assertEqual(self.run_cmd("verify", plan_file).returncode, 0, self.run_cmd("verify", plan_file).stdout)

    def test_another_root_cited_by_its_folder_name(self):
        ws = self.tmp() / "ws"
        mono = small_repo(ws / "mono")
        write(mono, ".github/workflows/deploy.yml", "steps:\n  - run: _support/scripts/x.sh\n")
        write(mono, "_support/scripts/x.sh", "echo\n")
        commit(mono)
        specs = small_repo(ws / "specs")
        write(specs, "docs/run.md", "Run `mono/_support/scripts/x.sh`.\n")
        commit(specs)
        plan = self.plan_json(mono, specs)
        self.assertIn(("docs/run.md", "mono/_support/scripts/x.sh", "mono/scripts/ci/x.sh"),
                      {(r["file"], r["old"], r["new"]) for r in plan["rewrites"] if r["root"] == 1})

    def test_a_test_of_a_library_and_the_library_follow_ci(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, "_support/scripts/common.lib.sh", "echo\n")
        write(repo, "_support/scripts/common.lib.test.sh", ". ./common.lib.sh\n")
        commit(repo)
        plan = self.plan_json(repo)
        self.assertEqual(self.dest(plan, "_support/scripts/common.lib.test.sh")[0], "scripts/ci/common.lib.test.sh")
        self.assertEqual(self.dest(plan, "_support/scripts/common.lib.sh")[0], "scripts/ci/common.lib.sh")

    def test_an_override_carries_the_test_along(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, ".github/workflows/deploy.yml", "steps:\n  - run: _support/scripts/deploy.sh\n")
        write(repo, "_support/scripts/deploy.sh", "echo\n")
        write(repo, "_support/scripts/deploy.test.sh", "echo\n")
        commit(repo)
        plan = self.plan_json(repo, extra=["--override", "_support/scripts/deploy.sh=tools/release/deploy.sh"])
        self.assertEqual(self.dest(plan, "_support/scripts/deploy.test.sh")[0], "tools/release/deploy.test.sh")

    def test_a_tracked_build_folder_is_a_caller(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, "package.json", "{}\n")
        write(repo, "build/release.sh", "bash _support/scripts/sign.sh\n")
        write(repo, "_support/scripts/sign.sh", "echo\n")
        commit(repo)
        self.assertIn(("build/release.sh", "_support/scripts/sign.sh", "scripts/sign.sh"), self.rewrites(self.plan_json(repo)))

    def test_version_names_are_not_one_off_and_loose_files_are_files(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, "_support/scripts/setup-node-18.sh", "echo\n")
        write(repo, "_support/scripts/load-PRJ-645.sh", "echo\n")
        write(repo, "_support/notes.txt", "x\n")
        commit(repo)
        plan = self.plan_json(repo)
        self.assertEqual(self.dest(plan, "_support/scripts/setup-node-18.sh")[1], "R10")
        self.assertEqual(self.dest(plan, "_support/scripts/load-PRJ-645.sh")[1], "R9")
        self.assertEqual([r["kind"] for r in plan["rows"] if r["source"] == "_support/notes.txt"], ["file"])

    @staticmethod
    def dest(plan, source):
        rows = [r for r in plan["rows"] if r["source"] == source]
        assert rows, f"no row for {source}"
        return rows[0]["destination"], rows[0]["rule"]


class SecondReviewTest(Fixture, unittest.TestCase):
    def run_cmd(self, command, plan_file):
        return subprocess.run(["python3", "-I", str(SCRIPT), command, "--plan", str(plan_file)], text=True, capture_output=True)

    def test_ignored_folders_outside_support_are_not_read(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, ".gitignore", "logs/\n_support/workspace/\n")
        write(repo, ".github/workflows/deploy.yml", "steps:\n  - run: _support/scripts/x.sh\n")
        write(repo, "_support/scripts/x.sh", "echo\n")
        commit(repo)
        write(repo, "logs/gen.sh", "bash _support/scripts/x.sh\n")
        plan = self.plan_json(repo)
        self.assertFalse([r for r in plan["rewrites"] if r["file"].startswith("logs/")], plan["rewrites"])

    def test_a_committed_vendor_library_is_not_a_package(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, "go.mod", "module x\n")
        write(repo, "vendor/acme/composer.json", "{}\n")
        write(repo, "_support/scripts/acme/check.sh", "echo\n")
        commit(repo)
        dest = [r["destination"] for r in self.plan_json(repo)["rows"] if r["source"] == "_support/scripts/acme/check.sh"][0]
        self.assertFalse((dest or "").startswith("vendor/"), dest)
        # a committed build/ is source inside Git, so its manifest is still a package
        other = small_repo(self.tmp() / "b")
        write(other, "package.json", "{}\n")
        write(other, "build/web/package.json", "{}\n")
        write(other, "_support/scripts/web/x.sh", "echo\n")
        commit(other)
        dest = [r["destination"] for r in self.plan_json(other)["rows"] if r["source"] == "_support/scripts/web/x.sh"][0]
        self.assertTrue((dest or "").startswith("build/web/"), dest)

    def test_a_replaced_dependent_drops_its_old_collision(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, ".github/workflows/deploy.yml", "steps:\n  - run: _support/scripts/deploy.sh\n")
        write(repo, "scripts/ci/deploy.sh", "echo old\n")
        write(repo, "scripts/ci/deploy.test.sh", "echo old\n")
        write(repo, "_support/scripts/deploy.sh", "echo\n")
        write(repo, "_support/scripts/deploy.test.sh", "echo\n")
        commit(repo)
        plan = self.plan_json(repo)
        row = [r for r in plan["rows"] if r["source"] == "_support/scripts/deploy.test.sh"][0]
        self.assertIsNone(row["destination"])
        self.assertNotIn("collision", row["notes"])
        details = [r["detail"] for r in plan["reviews"] if r["path"] == "_support/scripts/deploy.test.sh"]
        self.assertEqual(details, ["subject has no destination"], details)

    def test_a_dependent_neither_lands_in_an_ignored_path_nor_takes_an_overridden_one(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, ".gitignore", "_support/workspace/\nscripts/ci/*.test.sh\n")
        write(repo, ".github/workflows/deploy.yml", "steps:\n  - run: _support/scripts/deploy.sh\n")
        write(repo, "_support/scripts/deploy.sh", "echo\n")
        write(repo, "_support/scripts/deploy.test.sh", "echo\n")
        write(repo, "_support/scripts/other.sh", "echo\n")
        git(repo, "add", "-f", "_support/scripts/deploy.test.sh")
        commit(repo)
        plan = self.plan_json(repo)
        self.assertIsNone(PlanLayoutTest.dest(plan, "_support/scripts/deploy.test.sh")[0])
        git(repo, "rm", "-q", "--cached", ".gitignore")
        write(repo, ".gitignore", "_support/workspace/\n")
        commit(repo)
        plan = self.plan_json(repo, extra=["--override", "_support/scripts/other.sh=scripts/ci/deploy.test.sh"])
        other = [r for r in plan["rows"] if r["source"] == "_support/scripts/other.sh"][0]
        self.assertEqual((other["destination"], other["override"]), ("scripts/ci/deploy.test.sh", True))
        self.assertIsNone(PlanLayoutTest.dest(plan, "_support/scripts/deploy.test.sh")[0])

    def test_a_user_override_wins_a_shared_destination(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, ".github/workflows/deploy.yml", "steps:\n  - run: _support/scripts/deploy.sh\n  - run: _support/scripts/other.sh\n")
        for name in ("deploy.sh", "deploy.test.sh", "other.sh"):
            write(repo, f"_support/scripts/{name}", "echo\n")
        commit(repo)
        plan = self.plan_json(repo, extra=["--override", "_support/scripts/deploy.test.sh=scripts/ci/other.sh"])
        test = [r for r in plan["rows"] if r["source"] == "_support/scripts/deploy.test.sh"][0]
        self.assertEqual((test["destination"], test["override"]), ("scripts/ci/other.sh", True))
        self.assertIsNone(PlanLayoutTest.dest(plan, "_support/scripts/other.sh")[0])

    def test_undo_never_empties_a_file_it_cannot_read(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, ".github/workflows/deploy.yml", "steps:\n  - run: _support/scripts/x.sh\n")
        write(repo, "_support/scripts/x.sh", "echo\n")
        commit(repo)
        plan_file = repo.parent / "plan.json"
        self.plan(repo, extra=["--out", str(plan_file)])
        self.assertEqual(self.run_cmd("apply", plan_file).returncode, 0)
        workflow = repo / ".github/workflows/deploy.yml"
        workflow.write_bytes(workflow.read_bytes() + b"\0")
        result = self.run_cmd("undo", plan_file)
        self.assertEqual(result.returncode, 1)
        self.assertGreater(workflow.stat().st_size, 1)

    def test_a_non_ascii_ignored_destination_is_still_guarded(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, ".gitignore", "infra/revisi\u00f3n/\n")
        write(repo, "_support/infrastructure/revisi\u00f3n/main.tf", "x\n")
        commit(repo)
        plan = self.plan_json(repo)
        row = [r for r in plan["rows"] if r["source"].endswith("main.tf")][0]
        self.assertIsNone(row["destination"])

    def test_verify_ignores_links_broken_before_the_move(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, "_support/contracts/a.md", "[gone](../old/gone.md)\n")
        commit(repo)
        plan_file = repo.parent / "plan.json"
        self.plan(repo, extra=["--out", str(plan_file)])
        self.assertEqual(self.run_cmd("apply", plan_file).returncode, 0)
        result = self.run_cmd("verify", plan_file)
        self.assertEqual(result.returncode, 0, result.stdout)

    def test_history_records_keep_their_links_when_they_move(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, ".gitignore", "_support/evidence/*\n_support/workspace/\n")
        write(repo, "_support/docs/x.md", "# x\n")
        commit(repo)
        write(repo, "_support/evidence/2026-01-02-abc-1-x/report.md", "[x](../../docs/x.md)\n")
        plan = self.plan_json(repo)
        self.assertFalse([r for r in plan["rewrites"] if "report.md" in r["file"]], plan["rewrites"])


def dated_commit(repo, date, message="x"):
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    env.update(GIT_COMMITTER_DATE=f"{date}T12:00:00Z", GIT_AUTHOR_DATE=f"{date}T12:00:00Z")
    subprocess.run(["git", "-C", str(repo), "add", "-A"], env=env, check=True)
    subprocess.run(["git", "-C", str(repo), "commit", "-q", "-m", message], env=env, check=True)


def legacy_repo(base):
    """A repository with an origin, tracked legacy homes and change folders in three states."""
    repo = small_repo(base / "r")
    origin = base / "origin.git"
    subprocess.run(["git", "init", "-q", "--bare", "-b", "main", str(origin)], check=True)
    git(repo, "remote", "add", "origin", str(origin))
    write(repo, "AGENTS.md", "# Guide\n\n## Hive\n\n- Base branch: main\n- Specs: _support/openspec\n")
    write(repo, "_support/docs/README.md", "# Index\n")
    write(repo, "_support/sessions/2026-01-02-x/a.md", "# session\n")
    write(repo, "_support/evidence/b.txt", "evidence\n")
    write(repo, "_support/openspec/changes/archive/2026-01-03-y/proposal.md", "# y\n")
    write(repo, "_support/openspec/changes/archive/README.md", "# index\n")
    write(repo, "_support/openspec/changes/done-change/proposal.md", "# done\n")
    git(repo, "add", "-A")
    git(repo, "add", "-f", "_support/evidence/b.txt")
    dated_commit(repo, "2026-02-03")
    git(repo, "push", "-q", "origin", "main")
    write(repo, "_support/openspec/changes/open-change/proposal.md", "# open\n")
    git(repo, "add", "-A")
    git(repo, "commit", "-q", "-m", "local only")
    return repo


class LegacyMigrationTest(Fixture, unittest.TestCase):
    def run_cmd(self, command, plan_file):
        return subprocess.run(["python3", "-I", str(SCRIPT), command, "--plan", str(plan_file)], text=True, capture_output=True)

    @staticmethod
    def row(plan, source):
        rows = [r for r in plan["rows"] if r["source"] == source]
        assert rows, f"no row for {source}: {[r['source'] for r in plan['rows']]}"
        return rows[0]

    def test_tracked_legacy_homes_move_to_the_work_folder_and_leave_git(self):
        plan = self.plan_json(legacy_repo(self.tmp()))
        session = self.row(plan, "_support/sessions/2026-01-02-x")
        self.assertEqual((session["action"], session["destination"], session["untrack"]),
                         ("move", "_support/workspace/2026-01-02-x", True))
        evidence = self.row(plan, "_support/evidence/b.txt")
        self.assertEqual((evidence["action"], evidence["untrack"]), ("move", True))
        self.assertEqual(evidence["destination"], "_support/workspace/2026-02-03-legacy-evidence/evidence/b.txt")

    def test_the_change_archive_and_integrated_folders_are_deleted_and_the_unintegrated_one_is_kept(self):
        plan = self.plan_json(legacy_repo(self.tmp()))
        for gone in ("_support/openspec/changes/archive/2026-01-03-y", "_support/openspec/changes/archive/README.md", "_support/openspec/changes/done-change"):
            row = self.row(plan, gone)
            self.assertEqual((row["action"], row["destination"]), ("delete", None), gone)
        kept = self.row(plan, "_support/openspec/changes/open-change")
        self.assertEqual((kept["action"], kept["destination"]), ("keep", None))
        reviews = [r for r in plan["reviews"] if r["path"] == "_support/openspec/changes/open-change"]
        self.assertEqual(len(reviews), 1, plan["reviews"])
        self.assertIn("origin/main", reviews[0]["detail"])

    def test_a_change_folder_with_uncommitted_edits_is_kept_even_when_its_commit_is_integrated(self):
        repo = legacy_repo(self.tmp())
        write(repo, "_support/openspec/changes/done-change/proposal.md", "# done, edited\n")
        row = self.row(self.plan_json(repo), "_support/openspec/changes/done-change")
        self.assertEqual(row["action"], "keep")

    def test_without_a_base_on_the_remote_every_change_folder_is_kept(self):
        repo = legacy_repo(self.tmp())
        git(repo, "remote", "remove", "origin")
        plan = self.plan_json(repo)
        self.assertEqual(self.row(plan, "_support/openspec/changes/done-change")["action"], "keep")
        self.assertEqual(self.row(plan, "_support/openspec/changes/archive/2026-01-03-y")["action"], "delete")

    def test_an_undated_item_takes_its_first_commit_date(self):
        repo = legacy_repo(self.tmp())
        write(repo, "_support/sessions/notes/plan.md", "# plan\n")
        dated_commit(repo, "2026-04-05")
        row = self.row(self.plan_json(repo), "_support/sessions/notes")
        self.assertEqual(row["destination"], "_support/workspace/2026-04-05-notes")

    def test_dumps_go_to_data_and_reproducible_output_is_deleted(self):
        repo = legacy_repo(self.tmp())
        write(repo, "_support/backups/db.dump", "dump\n")
        write(repo, "_support/reports/coverage/index.html", "<html>\n")
        write(repo, "_support/reports/2026-05-06-run/out.txt", "out\n")
        plan = self.plan_json(repo)
        backups = self.row(plan, "_support/backups")
        self.assertEqual(backups["action"], "move")
        self.assertTrue(backups["destination"].startswith("_support/data/"), backups)
        self.assertEqual(self.row(plan, "_support/reports/coverage")["action"], "move", "never deleted under reports/")
        self.assertEqual(self.row(plan, "_support/reports/2026-05-06-run")["destination"],
                         "_support/workspace/2026-05-06-run/evidence")

    def test_other_legacy_homes_are_recognised(self):
        repo = legacy_repo(self.tmp())
        for home in ("archive", "archived", "plan", "spec"):
            write(repo, f"_support/{home}/2026-06-07-{home}-item/a.md", "# a\n")
        plan = self.plan_json(repo)
        for home in ("archive", "archived"):
            row = self.row(plan, f"_support/{home}/2026-06-07-{home}-item")
            self.assertEqual((row["action"], row["destination"]), ("move", f"_support/workspace/2026-06-07-{home}-item"), home)
        for home in ("plan", "spec"):
            row = self.row(plan, f"_support/{home}/2026-06-07-{home}-item")
            self.assertEqual((row["action"], row["destination"]), ("review", None), home)

    def test_a_support_home_outside_git_moves_and_deletes_without_a_tag_and_keeps_every_change_folder(self):
        ws = self.tmp() / "ws"
        write(ws, "_support/sessions/2026-01-02-x/a.md", "# a\n")
        write(ws, "_support/openspec/changes/archive/2026-01-03-y/p.md", "# y\n")
        write(ws, "_support/openspec/changes/active/p.md", "# active\n")
        plan = self.plan_json(ws)
        self.assertFalse(plan["roots"][0]["git"])
        self.assertEqual(self.row(plan, "_support/sessions/2026-01-02-x")["action"], "move")
        kept = self.row(plan, "_support/openspec/changes/active")
        self.assertEqual(kept["action"], "keep")
        self.assertIn("owner uncertain", [r["reason"] for r in plan["reviews"] if r["path"] == kept["source"]])
        self.assertEqual(self.row(plan, "_support/openspec/changes/archive")["action"], "keep", "no history shows an archive outside Git")

    def test_apply_untracks_what_it_moves_and_undo_tracks_it_again(self):
        repo = legacy_repo(self.tmp())
        head = git(repo, "rev-parse", "HEAD").stdout
        plan_file = repo.parent / "plan.json"
        self.plan(repo, extra=["--out", str(plan_file)])
        git(repo, "tag", "archive/test")
        self.assertEqual(self.run_cmd("apply", plan_file).returncode, 0)
        self.assertEqual(git(repo, "rev-parse", "HEAD").stdout, head)
        self.assertEqual(git(repo, "ls-files", "_support/sessions", "_support/evidence").stdout, "")
        self.assertTrue((repo / "_support/workspace/2026-01-02-x/a.md").exists())
        self.assertTrue((repo / "_support/openspec/changes/archive").exists(), "apply never deletes")
        self.assertEqual(self.run_cmd("undo", plan_file).returncode, 0)
        self.assertEqual(sorted(git(repo, "ls-files", "_support/sessions", "_support/evidence").stdout.split()),
                         ["_support/evidence/b.txt", "_support/sessions/2026-01-02-x/a.md"])
        self.assertTrue((repo / "_support/sessions/2026-01-02-x/a.md").exists())
        self.assertEqual(git(repo, "status", "--porcelain").stdout, "")

    def test_verify_reports_a_dead_link_to_a_deleted_row(self):
        repo = legacy_repo(self.tmp())
        plan_file = self.make_plan(repo)
        write(repo, "docs/guide.md", "See [y](../_support/openspec/changes/archive/2026-01-03-y/proposal.md).\n")
        self.assertEqual(self.apply_after_tag(repo, plan_file).returncode, 0)
        shutil.rmtree(repo / "_support/openspec/changes/archive/2026-01-03-y")
        result = self.run_cmd("verify", plan_file)
        self.assertEqual(result.returncode, 1, result.stdout)
        self.assertIn("docs/guide.md:1", result.stdout)

    def test_verify_reports_a_dead_link_to_a_record_removed_by_legacy_remove_and_accepts_a_repoint(self):
        repo = legacy_repo(self.tmp())
        write(repo, "docs/guide.md", "See `_support/sessions/2026-01-02-x/a.md` and `_support/evidence/b.txt`.\n")
        commit(repo)
        plan_file = self.make_plan(repo, "--legacy", "remove")
        self.assertEqual(self.apply_after_tag(repo, plan_file).returncode, 0)
        for gone in ("_support/sessions/2026-01-02-x", "_support/evidence/b.txt"):
            git(repo, "rm", "-rq", gone)
        result = self.run_cmd("verify", plan_file)
        self.assertEqual(result.returncode, 1, result.stdout)
        self.assertIn("docs/guide.md:1: _support/sessions/2026-01-02-x/a.md", result.stdout)
        self.assertIn("docs/guide.md:1: _support/evidence/b.txt", result.stdout)
        head = git(repo, "rev-parse", "HEAD").stdout.strip()
        write(repo, "docs/guide.md", "See `archive/test:_support/sessions/2026-01-02-x/a.md` and "
                                     f"`{head}:_support/evidence/b.txt`, and the home `_support/sessions`.\n")
        result = self.run_cmd("verify", plan_file)
        self.assertEqual((result.returncode, result.stdout), (0, ""))

    def test_a_repoint_is_accepted_only_when_its_ref_holds_the_path(self):
        repo = legacy_repo(self.tmp())
        write(repo, "docs/guide.md", "See `_support/sessions/2026-01-02-x/a.md`.\n")
        commit(repo)
        plan_file = self.make_plan(repo, "--legacy", "remove")
        self.assertEqual(self.apply_after_tag(repo, plan_file).returncode, 0)
        git(repo, "rm", "-rq", "_support/sessions/2026-01-02-x")
        empty = git(repo, "commit-tree", "-m", "empty", git(repo, "hash-object", "-t", "tree", "/dev/null").stdout.strip()).stdout.strip()
        git(repo, "tag", "archive/empty", empty)
        for line, holds in (("archive/test:_support/sessions/2026-01-02-x/a.md", True),
                            ("archive/nothing:_support/sessions/2026-01-02-x/a.md", False),
                            ("archive/empty:_support/sessions/2026-01-02-x/a.md", False)):
            write(repo, "docs/guide.md", f"See `{line}`.\n")
            result = self.run_cmd("verify", plan_file)
            self.assertEqual(result.returncode, 0 if holds else 1, (line, result.stdout))
            if not holds:
                self.assertIn("ref does not hold the path", result.stdout)

    def test_an_untracked_delete_row_that_is_cited_goes_to_review(self):
        repo = legacy_repo(self.tmp())
        write(repo, ".gitignore", "_support/workspace/\n_support/archive/coverage/\n_support/archive/htmlcov/\n")
        write(repo, "_support/archive/coverage/index.html", "ignored\n")
        write(repo, "_support/archive/htmlcov/index.html", "ignored\n")
        write(repo, "docs/guide.md", "See the coverage report.\n")
        commit(repo)
        plan = self.plan_json(repo)
        cited = self.row(plan, "_support/archive/coverage")
        self.assertEqual((cited["action"], cited["destination"], cited["tracked"]), ("review", None, False))
        self.assertNotIn("cited_by", cited)
        reviews = [r for r in plan["reviews"] if r["path"] == "_support/archive/coverage"]
        self.assertEqual([r["reason"] for r in reviews], ["cited and not in Git: no tag restores it"])
        self.assertEqual(self.row(plan, "_support/archive/htmlcov")["action"], "delete")
        self.assertEqual(self.row(plan, "_support/openspec/changes/archive/2026-01-03-y")["action"], "delete")

    def test_verify_reports_a_citation_of_a_top_level_record_written_without_a_support_path(self):
        for legacy, removal in (("remove", "delete"), ("move", "move")):
            with self.subTest(legacy):
                repo = small_repo(self.tmp() / "r")
                write(repo, "_support/docs/README.md", "# Index\n")
                write(repo, "evidence/2026-03-04-audit/shot.txt", "shot\n")
                write(repo, "sessions/2026-03-05-review/a.md", "# a\n")
                write(repo, "README.md", "Captures: `evidence/2026-03-04-audit/shot.txt`, "
                                         "[s](sessions/2026-03-05-review/a.md), a new `evidence/2026-03-04-audit/never.txt`.\n")
                commit(repo)
                plan_file = self.make_plan(repo, "--legacy", legacy)
                plan = json.loads(plan_file.read_text())
                self.assertEqual(self.row(plan, "evidence/2026-03-04-audit")["action"], removal)
                self.assertEqual(self.apply_after_tag(repo, plan_file).returncode, 0)
                if legacy == "remove":
                    git(repo, "rm", "-rq", "evidence", "sessions")
                result = self.run_cmd("verify", plan_file)
                self.assertEqual(result.returncode, 1, result.stdout)
                self.assertIn("README.md:1: evidence/2026-03-04-audit/shot.txt", result.stdout)
                self.assertIn("README.md:1: sessions/2026-03-05-review/a.md", result.stdout)
                self.assertNotIn("never.txt", result.stdout)
                head = git(repo, "rev-parse", "HEAD").stdout.strip()
                write(repo, "README.md", "Captures: `archive/test:evidence/2026-03-04-audit/shot.txt`, "
                                         f"[s]({head}:sessions/2026-03-05-review/a.md).\n")
                result = self.run_cmd("verify", plan_file)
                self.assertEqual((result.returncode, result.stdout), (0, ""))
                write(repo, "_support/workspace/2026-03-04-audit/evidence/own.md", "evidence/2026-03-04-audit/shot.txt\n")
                result = self.run_cmd("verify", plan_file)
                self.assertEqual((result.returncode, result.stdout), (0, ""), "a moved destination is no citer")

    def test_verify_counts_what_an_untracked_deleted_row_held_as_existing(self):
        repo = legacy_repo(self.tmp())
        write(repo, ".gitignore", "_support/workspace/\n_support/archive/coverage/\n")
        write(repo, "_support/archive/coverage/index.html", "ignored\n")
        commit(repo)
        plan_file = self.make_plan(repo)
        self.assertEqual(self.row(json.loads(plan_file.read_text()), "_support/archive/coverage")["action"], "delete")
        write(repo, "docs/guide.md", "Open `_support/archive/coverage/index.html`.\n")
        self.assertEqual(self.apply_after_tag(repo, plan_file).returncode, 0)
        shutil.rmtree(repo / "_support/archive/coverage")
        result = self.run_cmd("verify", plan_file)
        self.assertEqual(result.returncode, 1, result.stdout)
        self.assertIn("docs/guide.md:1: _support/archive/coverage/index.html", result.stdout)

    def test_a_path_after_a_git_ref_is_a_repoint_only_for_a_real_ref(self):
        repo = legacy_repo(self.tmp())
        write(repo, "docs/env.md", "PYTHONPATH=_support/archive/lib:_support/sessions/2026-01-02-x/a.md\n"
                                   "at 1700000000:_support/sessions/2026-01-02-x/a.md\n"
                                   "see xarchive/t:_support/sessions/2026-01-02-x/a.md\n")
        commit(repo)
        plan_file = self.make_plan(repo, "--legacy", "remove")
        self.assertEqual(self.apply_after_tag(repo, plan_file).returncode, 0)
        git(repo, "rm", "-rq", "_support/sessions/2026-01-02-x")
        result = self.run_cmd("verify", plan_file)
        self.assertEqual(result.returncode, 1, result.stdout)
        for line in (1, 2, 3):
            self.assertIn(f"docs/env.md:{line}: _support/sessions/2026-01-02-x/a.md", result.stdout)

    def test_verify_ignores_the_manifest_of_the_skills_own_folder(self):
        repo = legacy_repo(self.tmp())
        folder = repo / "_support/workspace/2026-11-24-project-layout"
        folder.mkdir(parents=True)
        plan_file = folder / "plan.json"
        self.plan(repo, extra=["--legacy", "remove", "--out", str(plan_file)])
        self.assertEqual(self.apply_after_tag(repo, plan_file).returncode, 0)
        git(repo, "rm", "-rq", "_support/sessions/2026-01-02-x")
        write(repo, "_support/workspace/2026-11-24-project-layout/manifest.md", "`_support/sessions/2026-01-02-x/a.md`\n")
        result = self.run_cmd("verify", plan_file)
        self.assertEqual((result.returncode, result.stdout), (0, ""))

    def test_verify_passes_after_a_plain_move_whatever_a_placeholder_or_a_new_path_says(self):
        repo = legacy_repo(self.tmp())
        write(repo, "docs/guide.md", "Layout: `_support/sessions/YYYY-MM-DD-work/`, `_support/sessions/*`, "
                                     "`_support/sessions/2026-01-02-x/never-existed.md`.\n")
        commit(repo)
        plan_file = self.make_plan(repo)
        self.assertEqual(self.apply_after_tag(repo, plan_file).returncode, 0)
        result = self.run_cmd("verify", plan_file)
        self.assertEqual((result.returncode, result.stdout), (0, ""))

    def test_a_workspace_folder_git_does_not_ignore_is_reviewed(self):
        repo = legacy_repo(self.tmp())
        write(repo, ".gitignore", "")
        git(repo, "add", "-A")
        git(repo, "commit", "-q", "-m", "no ignore")
        plan = self.plan_json(repo)
        self.assertTrue([r for r in plan["reviews"] if r["reason"] == "workspace not ignored"], plan["reviews"])

    def apply_after_tag(self, repo, plan_file, tag="archive/test"):
        git(repo, "tag", tag)
        return self.run_cmd("apply", plan_file)

    def make_plan(self, repo, *extra):
        plan_file = repo.parent / "plan.json"
        self.plan(repo, extra=["--out", str(plan_file), *extra])
        return plan_file

    def test_a_work_in_sessions_and_evidence_moves_together_and_undo_restores_both(self):
        repo = legacy_repo(self.tmp())
        write(repo, "_support/evidence/2026-01-02-x/shot.txt", "shot\n")
        commit(repo)
        before = git(repo, "ls-files", "-s").stdout
        plan_file = self.make_plan(repo)
        plan = json.loads(plan_file.read_text())
        self.assertEqual(self.row(plan, "_support/evidence/2026-01-02-x")["destination"], "_support/workspace/2026-01-02-x/evidence")
        result = self.apply_after_tag(repo, plan_file)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue((repo / "_support/workspace/2026-01-02-x/a.md").exists())
        self.assertTrue((repo / "_support/workspace/2026-01-02-x/evidence/shot.txt").exists())
        self.assertFalse((repo / "_support/evidence/2026-01-02-x").exists())
        undone = self.run_cmd("undo", plan_file)
        self.assertEqual(undone.returncode, 0, undone.stderr)
        self.assertTrue((repo / "_support/sessions/2026-01-02-x/a.md").exists())
        self.assertTrue((repo / "_support/evidence/2026-01-02-x/shot.txt").exists())
        self.assertFalse((repo / "_support/workspace/2026-01-02-x").exists())
        self.assertEqual(git(repo, "ls-files", "-s").stdout, before)
        self.assertEqual(git(repo, "status", "--porcelain").stdout, "")

    def test_apply_refuses_a_nested_destination_the_outer_folder_already_holds(self):
        repo = legacy_repo(self.tmp())
        plan_file = self.make_plan(repo)
        plan = json.loads(plan_file.read_text())
        session = self.row(plan, "_support/sessions/2026-01-02-x")
        evidence = self.row(plan, "_support/evidence/b.txt")
        evidence["destination"] = session["destination"] + "/a.md"
        plan_file.write_text(json.dumps(plan))
        result = self.apply_after_tag(repo, plan_file)
        self.assertEqual(result.returncode, 1)
        self.assertIn("nested", result.stderr)
        self.assertTrue((repo / "_support/sessions/2026-01-02-x/a.md").exists())
        self.assertTrue((repo / "_support/evidence/b.txt").exists())

    def test_apply_records_the_index_entries_before_the_files_move(self):
        repo = legacy_repo(self.tmp())
        plan_file = self.make_plan(repo)
        git(repo, "tag", "archive/test")
        spec = importlib.util.spec_from_file_location("plan_layout_under_test", SCRIPT)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        undo_path = Path(str(plan_file) + ".undo.json")
        seen, real = [], os.replace

        def replace(src, dst):
            if str(src).endswith("_support/sessions/2026-01-02-x"):
                seen.append(json.loads(undo_path.read_text())["moves"])
            return real(src, dst)

        module.os.replace = replace
        try:
            self.assertEqual(module.cmd_apply(argparse.Namespace(plan=str(plan_file))), 0)
        finally:
            module.os.replace = real
        self.assertEqual(len(seen), 1)
        self.assertEqual([m["source"] for m in seen[0]], ["_support/sessions/2026-01-02-x"], "recorded before the move")
        self.assertIn("a.md", seen[0][0]["index"])

    def test_a_move_that_fails_leaves_the_files_and_no_recorded_move(self):
        repo = legacy_repo(self.tmp())
        plan_file = self.make_plan(repo)
        git(repo, "tag", "archive/test")
        spec = importlib.util.spec_from_file_location("plan_layout_under_test", SCRIPT)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        real = os.replace

        def refuse(src, dst):
            raise OSError("no space left")

        module.os.replace = refuse
        try:
            self.assertEqual(module.cmd_apply(argparse.Namespace(plan=str(plan_file))), 1)
        finally:
            module.os.replace = real
        self.assertEqual(json.loads(Path(str(plan_file) + ".undo.json").read_text())["moves"], [])
        self.assertTrue((repo / "_support/sessions/2026-01-02-x/a.md").exists())
        self.assertIn("_support/sessions/2026-01-02-x/a.md", git(repo, "ls-files").stdout)
        self.assertEqual(self.run_cmd("undo", plan_file).returncode, 0)

    def test_an_index_entry_with_a_name_that_is_not_utf8_stops_before_the_move(self):
        repo = legacy_repo(self.tmp())
        blob = subprocess.run(["git", "-C", str(repo), "hash-object", "-w", "/dev/null"], capture_output=True, text=True, check=True).stdout.strip()
        git(repo, "update-index", "--add", "--cacheinfo", f"100644,{blob},_support/sessions/2026-01-02-x/bad\udcff.md")
        spec = importlib.util.spec_from_file_location("plan_layout_under_test", SCRIPT)
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        with self.assertRaises(RuntimeError) as caught:
            module.untracked_entries(str(repo), "_support/sessions/2026-01-02-x")
        self.assertIn("UTF-8", str(caught.exception))

    def test_apply_refuses_a_source_whose_index_flags_undo_cannot_restore(self):
        for flag in ("skip-worktree", "assume-unchanged", "intent-to-add"):
            with self.subTest(flag=flag):
                repo = legacy_repo(self.tmp())
                if flag == "intent-to-add":
                    write(repo, "_support/sessions/2026-01-02-x/new.md", "# new\n")
                    git(repo, "add", "-N", "_support/sessions/2026-01-02-x/new.md")
                else:
                    git(repo, "update-index", f"--{flag}", "_support/sessions/2026-01-02-x/a.md")
                plan_file = self.make_plan(repo)
                result = self.apply_after_tag(repo, plan_file)
                self.assertEqual(result.returncode, 1, result.stdout)
                self.assertIn("intent-to-add, skip-worktree or assume-unchanged", result.stderr)
                self.assertTrue((repo / "_support/sessions/2026-01-02-x/a.md").exists())

    def test_undo_refuses_a_record_an_earlier_script_wrote_without_index_entries(self):
        repo = legacy_repo(self.tmp())
        plan_file = self.make_plan(repo)
        self.assertEqual(self.apply_after_tag(repo, plan_file).returncode, 0)
        undo_path = Path(str(plan_file) + ".undo.json")
        record = json.loads(undo_path.read_text())
        self.assertEqual(record["version"], 2)
        del record["version"]
        for move in record["moves"]:
            move.pop("index", None)
        undo_path.write_text(json.dumps(record))
        result = self.run_cmd("undo", plan_file)
        self.assertEqual(result.returncode, 1, result.stdout)
        self.assertIn("_support/sessions/2026-01-02-x", result.stderr)
        self.assertIn("git add --", result.stderr)
        self.assertTrue((repo / "_support/workspace/2026-01-02-x/a.md").exists(), "nothing is moved back half way")
        self.assertTrue(undo_path.exists())

    def test_uncommitted_untracked_and_ignored_files_keep_a_change_folder_whatever_the_status_config(self):
        for name, ignored in (("untracked", False), ("ignored", True)):
            with self.subTest(name):
                repo = legacy_repo(self.tmp())
                git(repo, "config", "status.showUntrackedFiles", "no")
                write(repo, ".gitignore", "_support/backups/\n_support/data/\n_support/workspace/\n*.log\n")
                commit(repo)
                extra = "x.log" if ignored else "x.md"
                write(repo, f"_support/openspec/changes/done-change/{extra}", "mine\n")
                write(repo, f"_support/openspec/changes/archive/{extra}", "mine\n")
                plan = self.plan_json(repo)
                self.assertEqual(self.row(plan, "_support/openspec/changes/done-change")["action"], "keep")
                self.assertEqual(self.row(plan, f"_support/openspec/changes/archive/{extra}")["action"], "keep")
                self.assertEqual(self.row(plan, "_support/openspec/changes/archive/2026-01-03-y")["action"], "delete", "a loose file keeps only itself")

    def test_an_override_into_a_tracked_home_stays_tracked_and_one_into_the_workspace_leaves_git(self):
        repo = legacy_repo(self.tmp())
        plan_file = self.make_plan(repo, "--override", "_support/openspec/changes/open-change=_support/docs/open-change",
                                   "--override", "_support/openspec/changes/archive/2026-01-03-y=_support/docs/arch",
                                   "--override", "_support/sessions/2026-01-02-x=_support/workspace/2026-01-02-kept")
        plan = json.loads(plan_file.read_text())
        self.assertFalse(self.row(plan, "_support/openspec/changes/open-change")["untrack"])
        self.assertFalse(self.row(plan, "_support/openspec/changes/archive/2026-01-03-y")["untrack"])
        self.assertTrue(self.row(plan, "_support/sessions/2026-01-02-x")["untrack"])
        self.assertEqual(self.apply_after_tag(repo, plan_file).returncode, 0)
        tracked = git(repo, "ls-files").stdout
        self.assertIn("_support/docs/open-change/proposal.md", tracked)
        self.assertIn("_support/docs/arch/proposal.md", tracked)
        self.assertNotIn("_support/workspace", tracked)

    def test_reproducible_output_is_deleted_only_when_untracked_ignored_and_outside_records_of_evidence(self):
        repo = legacy_repo(self.tmp())
        write(repo, ".gitignore", "_support/backups/\n_support/data/\n_support/workspace/\n_support/archive/coverage/\n_support/evidence/playwright-report/\n")
        write(repo, "_support/archive/htmlcov/x.html", "tracked\n")
        write(repo, "_support/archive/test-results/y.xml", "not ignored\n")
        commit(repo)
        write(repo, "_support/archive/coverage/index.html", "ignored\n")
        write(repo, "_support/evidence/playwright-report/z.xml", "evidence\n")
        plan = self.plan_json(repo)
        self.assertEqual(self.row(plan, "_support/archive/coverage")["action"], "delete")
        for source in ("_support/archive/htmlcov", "_support/archive/test-results", "_support/evidence/playwright-report"):
            row = self.row(plan, source)
            self.assertEqual(row["action"], "move", source)
            self.assertTrue(row["destination"], source)

    def test_undo_restores_the_index_entries_and_never_stages_ignored_files(self):
        repo = legacy_repo(self.tmp())
        write(repo, ".gitignore", "_support/backups/\n_support/data/\n_support/workspace/\n*.log\n")
        commit(repo)
        write(repo, "_support/sessions/2026-01-02-x/skip.log", "never tracked\n")
        before = git(repo, "ls-files", "-s").stdout
        plan_file = self.make_plan(repo)
        self.assertEqual(self.apply_after_tag(repo, plan_file).returncode, 0)
        self.assertEqual(self.run_cmd("undo", plan_file).returncode, 0)
        self.assertEqual(git(repo, "ls-files", "-s").stdout, before)
        self.assertEqual(git(repo, "ls-files", "--cached", "--", "*.log").stdout, "")

    def test_apply_refuses_a_source_with_staged_changes(self):
        repo = legacy_repo(self.tmp())
        write(repo, "_support/sessions/2026-01-02-x/a.md", "# staged edit\n")
        git(repo, "add", "_support/sessions/2026-01-02-x/a.md")
        plan_file = self.make_plan(repo)
        result = self.apply_after_tag(repo, plan_file)
        self.assertEqual(result.returncode, 1)
        self.assertIn("staged", result.stderr)
        self.assertTrue((repo / "_support/sessions/2026-01-02-x/a.md").exists())
        self.assertIn("a.md", git(repo, "diff", "--cached", "--name-only").stdout)

    def test_tracked_code_in_a_legacy_home_moves_with_its_work_folder(self):
        repo = legacy_repo(self.tmp())
        write(repo, "_support/sessions/2026-01-02-x/run.sh", "echo run\n")
        commit(repo)
        plan = self.plan_json(repo)
        self.assertEqual(self.row(plan, "_support/sessions/2026-01-02-x")["action"], "move")
        self.assertFalse([r for r in plan["rows"] if r["source"].endswith("run.sh")], plan["rows"])

    def test_an_archive_is_deleted_only_when_tracked_and_clean(self):
        repo = legacy_repo(self.tmp())
        write(repo, "_support/openspec/changes/archive/README.md", "# index, edited\n")
        plan = self.plan_json(repo)
        self.assertEqual(self.row(plan, "_support/openspec/changes/archive/README.md")["action"], "keep", "edited")
        self.assertEqual(self.row(plan, "_support/openspec/changes/archive/2026-01-03-y")["action"], "delete")
        repo = legacy_repo(self.tmp())
        write(repo, "_support/openspec/changes/archive/2026-02-02-new/p.md", "# untracked\n")
        plan = self.plan_json(repo)
        self.assertEqual(self.row(plan, "_support/openspec/changes/archive/2026-02-02-new")["action"], "keep", "untracked child")
        self.assertEqual(self.row(plan, "_support/openspec/changes/archive/2026-01-03-y")["action"], "delete", "its siblings still go")
        repo = legacy_repo(self.tmp())
        write(repo, "_support/openspec/changes/INDEX.md", "# loose, untracked\n")
        plan = self.plan_json(repo)
        self.assertEqual(self.row(plan, "_support/openspec/changes/INDEX.md")["action"], "keep")
        self.assertIn("owner uncertain", [r["reason"] for r in plan["reviews"] if r["path"].endswith("INDEX.md")])

    def test_apply_needs_an_archive_tag_on_the_head_when_a_row_leaves_git_or_is_deleted(self):
        repo = legacy_repo(self.tmp())
        plan_file = self.make_plan(repo)
        result = self.run_cmd("apply", plan_file)
        self.assertEqual(result.returncode, 1)
        self.assertIn("archive/", result.stderr)
        self.assertTrue((repo / "_support/sessions/2026-01-02-x/a.md").exists())
        git(repo, "tag", "other/x")
        self.assertEqual(self.run_cmd("apply", plan_file).returncode, 1, "only archive/* tags count")
        self.assertEqual(self.apply_after_tag(repo, plan_file).returncode, 0)

    def test_apply_outside_git_needs_no_tag(self):
        ws = self.tmp() / "ws"
        write(ws, "_support/sessions/2026-01-02-x/a.md", "# a\n")
        plan_file = ws.parent / "plan.json"
        self.plan(ws, extra=["--out", str(plan_file)])
        self.assertEqual(self.run_cmd("apply", plan_file).returncode, 0)

    def test_plan_and_spec_children_wait_for_keep_or_override(self):
        repo = legacy_repo(self.tmp())
        write(repo, "_support/plan/2026-06-07-a/a.md", "# a\n")
        write(repo, "_support/spec/2026-06-07-b/b.md", "# b\n")
        write(repo, "_support/spec/loose.md", "# loose\n")
        commit(repo)
        plan = self.plan_json(repo)
        for source in ("_support/plan/2026-06-07-a", "_support/spec/2026-06-07-b", "_support/spec/loose.md"):
            row = self.row(plan, source)
            self.assertEqual((row["action"], row["destination"]), ("review", None), source)
        plan = self.plan_json(repo, extra=["--keep", "_support/plan/2026-06-07-a",
                                           "--override", "_support/spec/2026-06-07-b=_support/workspace/2026-06-07-b"])
        self.assertEqual(self.row(plan, "_support/plan/2026-06-07-a")["action"], "keep")
        moved = self.row(plan, "_support/spec/2026-06-07-b")
        self.assertEqual((moved["action"], moved["destination"], moved["untrack"]),
                         ("move", "_support/workspace/2026-06-07-b", True))

    def test_keep_and_override_name_an_unintegrated_change_folder(self):
        repo = legacy_repo(self.tmp())
        source = "_support/openspec/changes/open-change"
        kept = self.plan(repo, extra=["--keep", source])
        plan = json.loads(kept.stdout)
        self.assertEqual(self.row(plan, source)["action"], "keep")
        self.assertFalse([r for r in plan["reviews"] if r["path"] == source and r["reason"] == "not integrated"])
        plan = self.plan_json(repo, extra=["--override", f"{source}=_support/workspace/2026-02-03-open"])
        row = self.row(plan, source)
        self.assertEqual((row["action"], row["destination"]), ("move", "_support/workspace/2026-02-03-open"))

    def test_a_failing_git_status_keeps_the_change_folder(self):
        repo = legacy_repo(self.tmp())
        (repo / ".git/index").write_bytes(b"not an index")
        plan = self.plan_json(repo)
        row = self.row(plan, "_support/openspec/changes/done-change")
        self.assertEqual(row["action"], "keep")
        self.assertIn("git status failed", " ".join(r["detail"] for r in plan["reviews"] if r["path"] == row["source"]))


def specs_workspace(base, clean=3):
    """A workspace folder outside Git holding a specs repository with a tracked change archive, one
    archived change with untracked files, and tracked top-level evidence/ and sessions/ folders."""
    ws = base / "ws"
    write(ws, "AGENTS.md", "# Workspace\n\n## Hive\n\n- Specs: specs-repo/openspec\n")
    specs = ws / "specs-repo"
    specs.mkdir(parents=True)
    git(specs, "init", "-q", "-b", "main")
    git(specs, "config", "user.email", "t@example.invalid")
    git(specs, "config", "user.name", "t")
    origin = base / "origin.git"
    subprocess.run(["git", "init", "-q", "--bare", "-b", "main", str(origin)], check=True)
    git(specs, "remote", "add", "origin", str(origin))
    write(specs, "AGENTS.md", "# Specs\n\n## Hive\n\n- Base branch: main\n")
    for n in range(clean):
        write(specs, f"openspec/changes/archive/2026-01-0{n + 1}-done-{n}/proposal.md", "# done\n")
        write(specs, f"openspec/changes/archive/2026-01-0{n + 1}-done-{n}/tasks.md", "- [x] ok\n")
    write(specs, "openspec/changes/archive/2026-02-01-dirty/proposal.md", "# dirty\n")
    write(specs, "evidence/2026-03-04-audit/shot.txt", "shot\n")
    write(specs, "evidence/loose.txt", "loose\n")
    write(specs, "sessions/2026-03-05-review/a.md", "# a\n")
    git(specs, "add", "-A")
    dated_commit(specs, "2026-03-06")
    git(specs, "push", "-q", "origin", "main")
    write(specs, "openspec/changes/live/proposal.md", "# live, not on the base\n")
    git(specs, "add", "-A")
    git(specs, "commit", "-q", "-m", "live")
    for n in range(4):
        write(specs, f"openspec/changes/archive/2026-02-01-dirty/extra-{n}.md", "# untracked\n")
    return ws, specs


class LegacyGapsTest(Fixture, unittest.TestCase):
    @staticmethod
    def row(plan, root, source):
        rows = [r for r in plan["rows"] if r["root"] == root and r["source"] == source]
        assert rows, f"no row for {source}: {[(r['root'], r['source']) for r in plan['rows']]}"
        return rows[0]

    def test_the_change_archive_is_planned_per_child(self):
        ws, specs = specs_workspace(self.tmp())
        plan = self.plan_json(specs)
        archive = "openspec/changes/archive"
        for n in range(3):
            row = self.row(plan, 0, f"{archive}/2026-01-0{n + 1}-done-{n}")
            self.assertEqual((row["action"], row["destination"]), ("delete", None), row["source"])
        dirty = self.row(plan, 0, f"{archive}/2026-02-01-dirty")
        self.assertEqual(dirty["action"], "keep")
        reasons = [r for r in plan["reviews"] if r["path"] == dirty["source"]]
        self.assertEqual(len(reasons), 1, plan["reviews"])
        self.assertIn("untracked", reasons[0]["detail"])
        self.assertFalse([r for r in plan["rows"] if r["source"] == archive], "no row judges the archive as a whole")
        self.assertEqual(self.row(plan, 0, "openspec/changes/live")["action"], "keep", "not on the base: unchanged rule")

    def test_a_loose_file_inside_the_archive_is_judged_like_a_child(self):
        ws, specs = specs_workspace(self.tmp(), clean=1)
        write(specs, "openspec/changes/archive/README.md", "# untracked index\n")
        plan = self.plan_json(specs)
        self.assertEqual(self.row(plan, 0, "openspec/changes/archive/README.md")["action"], "keep")
        self.assertEqual(self.row(plan, 0, "openspec/changes/archive/2026-01-01-done-0")["action"], "delete")

    def test_a_workspace_root_outside_git_skips_the_repositories_inside_it(self):
        ws, specs = specs_workspace(self.tmp())
        plan = self.plan_json(ws, specs)
        self.assertEqual([r["source"] for r in plan["rows"] if r["root"] == 0 and "specs-repo" in r["source"]], [])
        self.assertTrue([r for r in plan["rows"] if r["root"] == 1 and r["action"] == "delete"])
        self.assertFalse([r for r in plan["reviews"] if r["root"] == 0 and "no repository" in r["detail"]], plan["reviews"])

    def test_top_level_legacy_homes_move_to_the_workspace_support_home_and_leave_git(self):
        ws, specs = specs_workspace(self.tmp())
        plan = self.plan_json(ws, specs)
        shot = self.row(plan, 1, "evidence/2026-03-04-audit")
        self.assertEqual((shot["action"], shot["destination"], shot["untrack"]),
                         ("move", "../_support/workspace/2026-03-04-audit/evidence", True))
        session = self.row(plan, 1, "sessions/2026-03-05-review")
        self.assertEqual((session["action"], session["destination"], session["untrack"]),
                         ("move", "../_support/workspace/2026-03-05-review", True))
        loose = self.row(plan, 1, "evidence/loose.txt")
        self.assertEqual((loose["action"], loose["untrack"]), ("move", True))
        self.assertTrue(loose["destination"].startswith("../_support/workspace/"), loose)
        self.assertFalse([r for r in plan["reviews"] if r["reason"] == "workspace not ignored"], plan["reviews"])

    def test_top_level_legacy_homes_use_the_repository_support_home_when_it_exists(self):
        ws, specs = specs_workspace(self.tmp())
        write(specs, "_support/docs/README.md", "# Index\n")
        write(specs, ".gitignore", "_support/workspace/\n")
        commit(specs)
        plan = self.plan_json(specs)
        row = self.row(plan, 0, "sessions/2026-03-05-review")
        self.assertEqual((row["destination"], row["untrack"]), ("_support/workspace/2026-03-05-review", True))

    def test_every_top_level_legacy_home_is_found(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, "_support/docs/README.md", "# Index\n")
        for home in ("sessions", "evidence", "reports", "archive", "archived"):
            write(repo, f"{home}/2026-06-07-{home}-item/a.md", "# a\n")
        commit(repo)
        plan = self.plan_json(repo)
        for home in ("sessions", "evidence", "reports", "archive", "archived"):
            row = self.row(plan, 0, f"{home}/2026-06-07-{home}-item")
            self.assertEqual((row["action"], row["untrack"]), ("move", True), home)
            self.assertTrue(row["destination"].startswith("_support/workspace/2026-06-07-"), row)

    def test_legacy_remove_deletes_only_tracked_unmodified_records(self):
        base = self.tmp()
        repo = legacy_repo(base)
        write(repo, "_support/backups/db.dump", "dump\n")
        write(repo, "_support/sessions/2026-07-08-untracked/a.md", "# u\n")
        write(repo, "_support/sessions/2026-01-02-x/extra.md", "# untracked inside\n")
        plan = self.plan_json(repo, extra=["--legacy", "remove"])
        self.assertEqual(plan["legacy"], "remove")
        dirty = self.row(plan, 0, "_support/sessions/2026-01-02-x")
        self.assertEqual((dirty["action"], dirty["untrack"]), ("move", True), "an untracked file inside keeps the move")
        self.assertEqual(self.row(plan, 0, "_support/sessions/2026-07-08-untracked")["action"], "move")
        self.assertEqual(self.row(plan, 0, "_support/backups")["action"], "move")
        self.assertTrue(self.row(plan, 0, "_support/backups")["destination"].startswith("_support/data/"))
        clean = self.row(plan, 0, "_support/evidence/b.txt")
        self.assertEqual((clean["action"], clean["destination"], clean["untrack"]), ("delete", None, False))
        self.assertIn("aged record: history and the archive tag keep it", clean["notes"])
        ws, specs = specs_workspace(self.tmp())
        write(specs, "evidence/2026-07-01-new/a.txt", "untracked\n")
        write(specs, "sessions/2026-03-05-review/extra.md", "# untracked inside\n")
        top = self.plan_json(specs, extra=["--legacy", "remove"])
        self.assertEqual(self.row(top, 0, "evidence/2026-03-04-audit")["action"], "delete")
        self.assertEqual(self.row(top, 0, "evidence/loose.txt")["action"], "delete")
        self.assertEqual(self.row(top, 0, "evidence/2026-07-01-new")["action"], "move")
        self.assertEqual(self.row(top, 0, "sessions/2026-03-05-review")["action"], "move")

    def test_legacy_remove_keeps_every_row_that_needs_a_decision(self):
        repo = legacy_repo(self.tmp())
        for home in ("plan", "spec"):
            write(repo, f"_support/{home}/2026-06-07-{home}-item/a.md", "# a\n")
        commit(repo)
        plan = self.plan_json(repo, extra=["--legacy", "remove"])
        for home in ("plan", "spec"):
            self.assertEqual(self.row(plan, 0, f"_support/{home}/2026-06-07-{home}-item")["action"], "review", home)
        ws, specs = specs_workspace(self.tmp())
        write(specs, "config.json", '{"out": "evidence"}\n')
        commit(specs)
        cited = self.plan_json(ws, specs, extra=["--legacy", "remove"])
        self.assertEqual(self.row(cited, 1, "evidence")["action"], "review")
        outside = self.plan_json(ws, extra=["--legacy", "remove"])
        self.assertFalse([r for r in outside["rows"] if r["action"] == "delete" and "specs-repo" in r["source"]])

    def test_legacy_remove_sends_a_record_that_holds_data_to_the_data_home(self):
        repo = legacy_repo(self.tmp())
        held = {"2026-01-05-dbfix": "prod.sql", "2026-01-06-gz": "dump.sql.gz", "2026-01-07-cols": "rows.parquet",
                "2026-01-09-zip": "files.zip", "2026-01-10-named": "users-backup.json"}
        for name, data in held.items():
            write(repo, f"_support/sessions/{name}/{data}", "x\n")
        write(repo, "_support/evidence/2026-01-06-run/export.csv", "a,b\n")
        write(repo, "_support/sessions/2026-01-08-plain/notes.md", "# n\n")
        commit(repo)
        plan = self.plan_json(repo, extra=["--legacy", "remove"])
        for name in held:
            row = self.row(plan, 0, f"_support/sessions/{name}")
            self.assertEqual((row["action"], row["destination"]), ("move", f"_support/data/{name}"), name)
            self.assertIn("holds a dump or export: its removal is a decision of its own", row["notes"])
        csv = self.row(plan, 0, "_support/evidence/2026-01-06-run")
        self.assertEqual(csv["destination"], "_support/data/2026-01-06-run/evidence")
        self.assertEqual(self.row(plan, 0, "_support/sessions/2026-01-08-plain")["action"], "delete")

    def test_a_record_a_document_or_code_cites_is_deleted_and_lists_its_citations(self):
        repo = legacy_repo(self.tmp())
        for name in ("2026-03-01-ws", "2026-03-02-enc", "2026-03-04-free"):
            write(repo, f"_support/sessions/{name}/a.md", "# a\n")
        write(repo, "_support/sessions/2026-03-05 spaced/a.md", "# a\n")
        write(repo, "_support/evidence/2026-01-05-run/shot.png", b"png")
        write(repo, "_support/workspace/2026-03-09-x/notes.md", "see the 2026-03-01-ws run\n")
        write(repo, "docs/guide.md", "![shot](../_support/evidence/2026-01-05-run/shot.png)\n")
        write(repo, "src/app.py", "print('2026-03-02-enc')\n")
        write(repo, "_support/sessions/2026-03-04-free/a.md", "# a, mentions 2026-03-04-free itself\n")
        commit(repo)
        plan = self.plan_json(repo, extra=["--legacy", "remove"])
        name = os.path.basename(repo)
        for source, where in (("_support/sessions/2026-03-01-ws", "_support/workspace/2026-03-09-x/notes.md:1"),
                              ("_support/sessions/2026-03-02-enc", "src/app.py:1"),
                              ("_support/evidence/2026-01-05-run", "docs/guide.md:1")):
            row = self.row(plan, 0, source)
            self.assertEqual((row["action"], row["destination"]), ("delete", None), source)
            self.assertEqual(row["cited_by"], [f"{name}/{where}"], source)
            self.assertIn(f"cited by 1: {name}/{where}", row["notes"])
            self.assertFalse([r for r in plan["reviews"] if r["path"] == source], source)
        free = self.row(plan, 0, "_support/sessions/2026-03-04-free")
        self.assertEqual((free["action"], free["cited_by"]), ("delete", []), "its own files are not citations")
        self.assertNotIn("cited_by_more", free)

    def test_a_cited_change_archive_child_is_deleted_with_its_citations_and_cited_untracked_output_is_reviewed(self):
        repo = legacy_repo(self.tmp())
        write(repo, ".gitignore", "_support/workspace/\n_support/archive/coverage/\n")
        write(repo, "_support/archive/coverage/index.html", "ignored\n")
        write(repo, "docs/guide.md", "Archived: openspec/changes/archive/2026-01-03-y and the coverage report.\n")
        commit(repo)
        plan = self.plan_json(repo)
        child = self.row(plan, 0, "_support/openspec/changes/archive/2026-01-03-y")
        self.assertEqual((child["action"], len(child["cited_by"])), ("delete", 1))
        coverage = self.row(plan, 0, "_support/archive/coverage")
        self.assertEqual(coverage["action"], "review", "untracked and cited: no tag restores it")

    def test_a_citation_list_is_sorted_capped_and_counts_the_rest(self):
        repo = legacy_repo(self.tmp())
        write(repo, "_support/sessions/2026-03-01-popular/a.md", "# a\n")
        for n in range(12):
            write(repo, f"docs/g{n:02d}.md", "\nsee 2026-03-01-popular\n")
        commit(repo)
        row = self.row(self.plan_json(repo, extra=["--legacy", "remove"]), 0, "_support/sessions/2026-03-01-popular")
        name = os.path.basename(repo)
        self.assertEqual(row["action"], "delete")
        self.assertEqual(row["cited_by"], [f"{name}/docs/g{n:02d}.md:2" for n in range(10)])
        self.assertEqual(row["cited_by_more"], 2)
        self.assertIn(f"cited by 12: {name}/docs/g00.md:2, {name}/docs/g01.md:2, {name}/docs/g02.md:2, ...", row["notes"])

    def test_citations_of_a_moved_row_are_listed_without_changing_its_action(self):
        repo = legacy_repo(self.tmp())
        write(repo, "docs/guide.md", "the 2026-01-02-x run\n")
        commit(repo)
        row = self.row(self.plan_json(repo), 0, "_support/sessions/2026-01-02-x")
        self.assertEqual(row["action"], "move")
        self.assertEqual(row["cited_by"], [f"{os.path.basename(repo)}/docs/guide.md:1"])

    def test_a_config_cited_top_level_folder_still_goes_to_review_and_a_document_does_not(self):
        ws, specs = specs_workspace(self.tmp())
        write(specs, "README.md", "Captures live in `evidence/2026-03-04-audit/shot.txt`.\n")
        commit(specs)
        plan = self.plan_json(ws, specs, extra=["--legacy", "remove"])
        self.assertFalse([r for r in plan["rows"] if r["source"] == "evidence"], "a document alone holds nothing back")
        self.assertEqual(self.row(plan, 1, "evidence/2026-03-04-audit")["action"], "delete")
        write(specs, "scripts/build.sh", 'cp out.log "$ROOT/evidence/out.log"\n')
        commit(specs)
        plan = self.plan_json(ws, specs, extra=["--legacy", "remove"])
        row = self.row(plan, 1, "evidence")
        self.assertEqual((row["action"], row["destination"]), ("review", None))
        self.assertIn("scripts/build.sh:1", [r for r in plan["reviews"] if r["path"] == "evidence"][0]["detail"])
        kept = self.plan_json(ws, specs, extra=["--keep", str(specs / "evidence")])
        self.assertEqual(self.row(kept, 1, "evidence")["action"], "keep")

    def test_re_planning_with_a_manifest_in_the_skills_own_folder_changes_nothing(self):
        repo = legacy_repo(self.tmp())
        extra = ["--legacy", "remove"]
        first = self.plan(repo, extra=extra).stdout
        folder = repo / "_support/workspace/2026-11-24-project-layout"
        out = folder / "plan.json"
        folder.mkdir(parents=True)
        self.plan(repo, extra=[*extra, "--out", str(out)])
        write(repo, "_support/workspace/2026-11-24-project-layout/manifest.md", "| 2026-01-02-x | delete |\n")
        self.assertEqual(self.plan(repo, extra=[*extra, "--out", str(out)]).stdout, first)
        in_root = repo / "plan.json"
        self.plan(repo, extra=[*extra, "--out", str(in_root)])
        self.assertEqual(self.plan(repo, extra=[*extra, "--out", str(in_root)]).stdout, first, "the --out file is no citer")
        in_root.unlink()
        write(repo, "docs/2026-11-24-project-layout-extra/x.md", "2026-01-02-x\n")
        row = self.row(self.plan_json(repo, extra=extra), 0, "_support/sessions/2026-01-02-x")
        self.assertEqual(len(row["cited_by"]), 1, "only the exact folder name is the skill's own")

    def test_citations_scan_a_large_file_quickly(self):
        import time
        repo = legacy_repo(self.tmp())
        write(repo, "_support/sessions/2026-03-01-big/a.md", "# a\n")
        write(repo, "docs/big.txt", "2026-03-01-big and padding padding padding padding padding padding\n" * 40000)
        self.assertGreater((repo / "docs/big.txt").stat().st_size, 2_500_000)
        commit(repo)
        row = self.row(self.plan_json(repo, extra=["--legacy", "remove"]), 0, "_support/sessions/2026-03-01-big")
        self.assertEqual(row["cited_by_more"], 40000 - 10)
        planner = load_planner().Planner([str(repo)], [], {})
        planner.text_files()
        started = time.monotonic()
        count, first = planner.citations([(0, "_support/sessions/2026-03-01-big")])[(0, "_support/sessions/2026-03-01-big")]
        self.assertLess(time.monotonic() - started, 10)
        self.assertEqual((count, len(first)), (40000, 10))

    def test_legacy_remove_moves_a_record_with_an_ignored_or_hidden_file_and_deletes_a_clean_folder(self):
        repo = legacy_repo(self.tmp())
        write(repo, ".gitignore", "_support/workspace/\n*.log\n")
        for name in ("ignored", "skipped", "assumed", "clean"):
            write(repo, f"_support/sessions/2026-02-01-{name}/a.md", "# a\n")
        commit(repo)
        write(repo, "_support/sessions/2026-02-01-ignored/trace.log", "ignored\n")
        git(repo, "update-index", "--skip-worktree", "_support/sessions/2026-02-01-skipped/a.md")
        git(repo, "update-index", "--assume-unchanged", "_support/sessions/2026-02-01-assumed/a.md")
        plan = self.plan_json(repo, extra=["--legacy", "remove"])
        for name in ("ignored", "skipped", "assumed"):
            self.assertEqual(self.row(plan, 0, f"_support/sessions/2026-02-01-{name}")["action"], "move", name)
        clean = self.row(plan, 0, "_support/sessions/2026-02-01-clean")
        self.assertEqual((clean["kind"], clean["action"]), ("folder", "delete"))

    def test_a_shell_variable_before_the_folder_is_still_a_citation_of_it(self):
        for line in ('cp out.log "$ROOT/evidence/out.log"', 'ls ${REPO}/evidence', 'ls "$(dirname "$0")/../evidence"'):
            ws, specs = specs_workspace(self.tmp())
            write(specs, "scripts/ci.sh", line + "\n")
            commit(specs)
            self.assertEqual(self.row(self.plan_json(ws, specs), 1, "evidence")["action"], "review", line)

    def test_a_support_path_is_no_citation_of_the_root_folder_of_the_same_name(self):
        ws, specs = specs_workspace(self.tmp())
        write(specs, "README.md", "Captures: `../_support/evidence/2026-03-04-audit/shot.txt`.\n")
        commit(specs)
        row = self.row(self.plan_json(ws, specs), 1, "evidence/2026-03-04-audit")
        self.assertEqual(row["action"], "move")

    def test_the_legacy_policy_defaults_to_move_and_a_removal_needs_the_archive_tag(self):
        repo = legacy_repo(self.tmp())
        plan = self.plan_json(repo)
        self.assertEqual(plan["legacy"], "move")
        self.assertEqual(self.row(plan, 0, "_support/evidence/b.txt")["action"], "move")
        plan_file = repo.parent / "remove.json"
        self.plan(repo, extra=["--legacy", "remove", "--out", str(plan_file)])
        result = subprocess.run(["python3", "-I", str(SCRIPT), "apply", "--plan", str(plan_file)], text=True, capture_output=True)
        self.assertEqual(result.returncode, 1)
        self.assertIn("archive/", result.stderr)
        git(repo, "tag", "archive/test")
        result = subprocess.run(["python3", "-I", str(SCRIPT), "apply", "--plan", str(plan_file)], text=True, capture_output=True)
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_apply_moves_top_level_homes_into_the_workspace_and_undo_restores_them(self):
        ws, specs = specs_workspace(self.tmp())
        plan_file = ws.parent / "plan.json"
        self.plan(ws, specs, extra=["--out", str(plan_file)])
        before = git(specs, "ls-files", "-s").stdout
        git(specs, "tag", "archive/test")
        run = lambda c: subprocess.run(["python3", "-I", str(SCRIPT), c, "--plan", str(plan_file)], text=True, capture_output=True)
        applied = run("apply")
        self.assertEqual(applied.returncode, 0, applied.stderr)
        self.assertTrue((ws / "_support/workspace/2026-03-05-review/a.md").exists())
        self.assertTrue((ws / "_support/workspace/2026-03-04-audit/evidence/shot.txt").exists())
        self.assertFalse((specs / "sessions").exists())
        self.assertEqual(git(specs, "ls-files", "sessions", "evidence").stdout, "")
        undone = run("undo")
        self.assertEqual(undone.returncode, 0, undone.stderr)
        self.assertTrue((specs / "sessions/2026-03-05-review/a.md").exists())
        self.assertEqual(git(specs, "ls-files", "-s").stdout, before)
        self.assertFalse((ws / "_support/workspace/2026-03-05-review").exists())

    def cited_workspace(self, citation, home="evidence", name="inventory-2026-06-25.md"):
        """A workspace README citing a top-level home of the specs repository, which its own configuration also cites."""
        ws, specs = specs_workspace(self.tmp())
        write(specs, f"{home}/{name}", "# inventory\n")
        write(specs, "NOTES.md", f"- see `{home}/{name}`\n")
        write(specs, "config.json", json.dumps({"out": home}) + "\n")
        git(specs, "add", "-A")
        git(specs, "commit", "-q", "-m", "home")
        write(ws, "_support/README.md", f"- see `{citation}` for the audit\n")
        return ws, specs

    def test_a_kept_top_level_home_cited_from_another_root_is_not_rewritten(self):
        ws, specs = self.cited_workspace("../specs-repo/evidence/inventory-2026-06-25.md")
        plan = self.plan_json(ws, specs, extra=["--keep", "evidence", "--keep", "sessions"])
        self.assertEqual(self.row(plan, 1, "evidence")["action"], "keep")
        self.assertEqual([r for r in plan["rewrites"] if r["file"] == "_support/README.md"], [])

    def test_a_moved_top_level_home_cited_from_another_root_resolves_to_its_destination(self):
        ws, specs = self.cited_workspace("../specs-repo/evidence/inventory-2026-06-25.md")
        plan = self.plan_json(ws, specs, extra=["--override", "evidence=_support/workspace/old-evidence"])
        row = self.row(plan, 1, "evidence")
        self.assertEqual(row["action"], "move")
        rewrites = [r for r in plan["rewrites"] if r["file"] == "_support/README.md"]
        self.assertEqual(len(rewrites), 1, plan["rewrites"])
        resolved = os.path.normpath(os.path.join(ws, "_support", rewrites[0]["new"]))
        self.assertEqual(resolved, os.path.normpath(os.path.join(specs, row["destination"], "inventory-2026-06-25.md")))


def sibling_repos(base, names=("a", "b"), with_workspace_sessions=False):
    """Two repositories under a workspace folder outside Git, each with an untracked sessions/notes.md
    of the same date, so the work folder both would use is the same absolute path."""
    ws = base / "ws"
    ws.mkdir(parents=True)
    repos = []
    for name in names:
        repo = small_repo(ws / name)
        write(repo, "sessions/notes.md", f"{name}\n")
        os.utime(repo / "sessions/notes.md", (FIXED, FIXED))
        repos.append(repo)
    if with_workspace_sessions:
        write(ws, "_support/sessions/notes.md", "workspace\n")
        os.utime(ws / "_support/sessions/notes.md", (FIXED, FIXED))
    return ws, repos


class LegacyReviewTest(Fixture, unittest.TestCase):
    @staticmethod
    def rows(plan, root=None):
        return [r for r in plan["rows"] if root is None or r["root"] == root]

    def run_cmd(self, command, plan_file):
        return subprocess.run(["python3", "-I", str(SCRIPT), command, "--plan", str(plan_file)], text=True, capture_output=True)

    def test_destinations_are_compared_by_absolute_path_across_roots(self):
        ws, (a, b) = sibling_repos(self.tmp(), with_workspace_sessions=True)
        plan = self.plan_json(ws, a, b)
        moving = [os.path.normpath(os.path.join(plan["roots"][r["root"]]["path"], r["destination"]))
                  for r in plan["rows"] if r["destination"]]
        self.assertEqual(len(moving), len(set(moving)), "two rows share one destination: " + str(moving))
        notes = [r for r in plan["rows"] if r["source"].endswith("notes.md")]
        self.assertEqual(sorted(r["action"] for r in notes), ["review"] * 3, notes)

    def test_apply_refuses_two_rows_that_share_a_destination_and_overwrites_nothing(self):
        ws, (a, b) = sibling_repos(self.tmp())
        write(b, "sessions/notes.md", "b\n")
        os.rename(b / "sessions/notes.md", b / "sessions/other.md")
        os.utime(b / "sessions/other.md", (FIXED, FIXED))
        plan_file = ws.parent / "plan.json"
        self.plan(ws, a, b, extra=["--out", str(plan_file)])
        plan = json.loads(plan_file.read_text())
        shared = "../_support/workspace/2026-01-02-legacy-sessions/z.md"
        for row in plan["rows"]:
            if row["action"] == "move":
                row["destination"] = shared
        plan_file.write_text(json.dumps(plan))
        result = self.run_cmd("apply", plan_file)
        self.assertNotEqual(result.returncode, 0, result.stdout)
        self.assertEqual((a / "sessions/notes.md").read_text(), "a\n")
        self.assertEqual((b / "sessions/other.md").read_text(), "b\n")
        self.assertFalse((ws / "_support/workspace/2026-01-02-legacy-sessions/z.md").exists())

    def test_a_move_out_of_the_repository_is_not_a_build_change(self):
        ws, specs = specs_workspace(self.tmp())
        write(specs, ".dockerignore", "node_modules\n")
        git(specs, "add", ".dockerignore")
        git(specs, "commit", "-q", "-m", "ignore")
        plan_file = ws.parent / "plan.json"
        plan = json.loads(self.plan(ws, specs, extra=["--out", str(plan_file)]).stdout)
        self.assertFalse([r for r in plan["rows"] if r.get("builds")], "a ../ destination leaves every build")
        self.assertFalse([r for r in plan["reviews"] if r["reason"] == "build contents"], plan["reviews"])
        git(specs, "tag", "archive/test")
        self.assertEqual(self.run_cmd("apply", plan_file).returncode, 0)
        verified = self.run_cmd("verify", plan_file)
        self.assertNotIn(".dockerignore", verified.stdout + verified.stderr)

    def test_a_plain_word_is_not_a_citation_of_a_top_level_folder(self):
        ws, specs = specs_workspace(self.tmp())
        write(specs, "reports/2026-04-05-q1/a.md", "# q1\n")
        write(specs, "AGENTS.md", "# Specs\n\n## Hive\n\n- Base branch: main\n\nWeekly reports are mailed. The evidence-based approach and sessions-heavy work stay.\n")
        commit(specs)
        plan = self.plan_json(ws, specs)
        for name in ("reports", "evidence", "sessions"):
            self.assertFalse([r for r in plan["rows"] if r["source"] == name], name)
        self.assertEqual([r for r in plan["rows"] if r["source"] == "reports/2026-04-05-q1"][0]["action"], "move")
        write(specs, "AGENTS.md", "# Specs\n\nSee `reports` and the evidence/ folder.\n")
        commit(specs)
        plan = self.plan_json(ws, specs)
        for name in ("reports", "evidence"):
            self.assertFalse([r for r in plan["rows"] if r["source"] == name], f"a document is no citation of {name}")
        write(specs, "app.py", 'OUT = "reports"\n')
        commit(specs)
        row = [r for r in self.plan_json(ws, specs)["rows"] if r["source"] == "reports"][0]
        self.assertEqual(row["action"], "review")

    def test_a_coordinator_repository_must_ignore_the_workspace_folder_a_move_creates(self):
        ws, specs = specs_workspace(self.tmp())
        git(ws, "init", "-q", "-b", "main")
        write(ws, ".gitignore", "specs-repo/\n")
        plan = self.plan_json(ws, specs)
        reviews = [r for r in plan["reviews"] if r["reason"] == "workspace not ignored"]
        self.assertEqual([(r["root"], r["path"]) for r in reviews], [(0, "_support/workspace")], plan["reviews"])
        write(ws, ".gitignore", "specs-repo/\n_support/workspace/\n")
        plan = self.plan_json(ws, specs)
        self.assertFalse([r for r in plan["reviews"] if r["reason"] == "workspace not ignored"], plan["reviews"])

    def test_an_embedded_clone_inside_a_legacy_home_is_reviewed_and_never_moved(self):
        ws, specs = specs_workspace(self.tmp())
        write(specs, "archive/2026-05-01-x/a.md", "# a\n")
        clone = specs / "archive/old-repo"
        clone.mkdir()
        git(clone, "init", "-q", "-b", "main")
        write(clone, "f.md", "# f\n")
        inner = specs / "openspec/changes/archive/old-repo"
        inner.mkdir()
        git(inner, "init", "-q", "-b", "main")
        write(inner, "f.md", "# f\n")
        plan = self.plan_json(ws, specs)
        for source in ("archive/old-repo", "openspec/changes/archive/old-repo"):
            self.assertFalse([r for r in plan["rows"] if r["source"] == source], source)
            self.assertEqual([r["reason"] for r in plan["reviews"] if r["path"] == source], ["embedded repository"], source)
        self.assertEqual([r for r in plan["rows"] if r["source"] == "archive/2026-05-01-x"][0]["action"], "move")

    def test_a_file_that_hides_its_edits_from_git_status_is_never_deleted(self):
        for flag in ("--assume-unchanged", "--skip-worktree"):
            with self.subTest(flag):
                ws, specs = specs_workspace(self.tmp())
                child = "openspec/changes/archive/2026-01-01-done-0"
                git(specs, "update-index", flag, f"{child}/proposal.md")
                write(specs, f"{child}/proposal.md", "# edited locally\n")
                self.assertEqual(git(specs, "status", "--porcelain", "--", child).stdout, "", "git status hides it")
                plan = self.plan_json(specs)
                row = [r for r in plan["rows"] if r["source"] == child][0]
                self.assertEqual(row["action"], "keep")

    def test_a_dotdot_override_leaves_git_whatever_its_name(self):
        repo = legacy_repo(self.tmp())
        plan = self.plan_json(repo, extra=["--override", "_support/sessions/2026-01-02-x=../elsewhere/x"])
        row = [r for r in plan["rows"] if r["source"] == "_support/sessions/2026-01-02-x"][0]
        self.assertEqual((row["destination"], row["untrack"]), ("../elsewhere/x", True))


def load_planner():
    spec = importlib.util.spec_from_file_location("plan_layout_under_test", SCRIPT)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class LegacyLastRoundTest(Fixture, unittest.TestCase):
    def run_cmd(self, command, plan_file):
        return subprocess.run(["python3", "-I", str(SCRIPT), command, "--plan", str(plan_file)], text=True, capture_output=True)

    @staticmethod
    def row(plan, source, root=0):
        rows = [r for r in plan["rows"] if r["root"] == root and r["source"] == source]
        assert rows, f"no row for {source}"
        return rows[0]

    def reports_repo(self, **files):
        repo = small_repo(self.tmp() / "r")
        write(repo, "_support/docs/README.md", "# Index\n")
        write(repo, "reports/2026-04-05-q1/a.md", "# q1\n")
        for rel, text in files.items():
            write(repo, rel.replace("__", "/").replace("_dot_", "."), text)
        commit(repo)
        return repo

    def test_a_bare_folder_name_in_code_or_configuration_is_a_citation(self):
        for rel, text in (("config.yml", "output_dir: reports\n"), ("app.py", 'OUT = "reports"\n'),
                          ("settings.json", '{"dir": "reports"}\n')):
            with self.subTest(rel):
                repo = small_repo(self.tmp() / "r")
                write(repo, "_support/docs/README.md", "# Index\n")
                write(repo, "reports/2026-04-05-q1/a.md", "# q1\n")
                write(repo, rel, text)
                commit(repo)
                self.assertEqual(self.row(self.plan_json(repo), "reports")["action"], "review")

    def test_a_file_name_that_starts_with_the_folder_name_is_not_a_citation(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, "_support/docs/README.md", "# Index\n")
        write(repo, "reports/2026-04-05-q1/a.md", "# q1\n")
        write(repo, "AGENTS.md", "# A\n\nSee `reports.md` and reports.md.\n")
        commit(repo)
        row = self.row(self.plan_json(repo), "reports/2026-04-05-q1")
        self.assertEqual(row["action"], "move")

    def test_destinations_that_differ_only_by_case_collide_on_a_case_insensitive_filesystem(self):
        module = load_planner()
        module.CASE_INSENSITIVE = True
        base = self.tmp()
        ws = base / "ws"
        ws.mkdir()
        repos = []
        for name, file in (("a", "Notes.md"), ("b", "notes.md")):
            repo = small_repo(ws / name)
            write(repo, f"evidence/{file}", name + "\n")
            os.utime(repo / f"evidence/{file}", (FIXED, FIXED))
            repos.append(repo)
        plan = module.Planner([str(ws), str(repos[0]), str(repos[1])], [], {}).plan()
        notes = [r for r in plan["rows"] if r["source"].lower().endswith("notes.md")]
        self.assertEqual([r["action"] for r in notes], ["review", "review"], notes)
        plan_file = base / "plan.json"
        for row, name in zip(notes, ("Notes.md", "notes.md")):
            row["action"], row["destination"] = "move", f"../_support/workspace/2026-01-02-x/evidence/{name}"
        plan_file.write_text(json.dumps(plan))
        errors, _ = module.validate(json.loads(plan_file.read_text()))
        self.assertTrue([e for e in errors if "same path" in e], errors)

    def test_a_dotdot_override_on_a_review_or_dest_row_leaves_git(self):
        repo = small_repo(self.tmp() / "r")
        write(repo, "_support/docs/README.md", "# Index\n")
        write(repo, "_support/notes/a.md", "# a\n")
        write(repo, "_support/misc/a.bin", "x\n")
        commit(repo)
        extra = ["--override", "_support/notes=../elsewhere/notes", "--override", "_support/misc=../elsewhere/misc"]
        plan = self.plan_json(repo, extra=extra)
        for source in ("_support/notes", "_support/misc"):
            row = self.row(plan, source)
            self.assertEqual((row["destination"], row["untrack"]), (f"../elsewhere/{source.split('/')[1]}", True), source)

    def test_apply_refuses_a_tracked_row_outside_the_repository_that_is_not_marked_to_leave_git(self):
        repo = legacy_repo(self.tmp())
        plan_file = repo.parent / "plan.json"
        self.plan(repo, extra=["--out", str(plan_file)])
        plan = json.loads(plan_file.read_text())
        row = [r for r in plan["rows"] if r["source"] == "_support/sessions/2026-01-02-x"][0]
        row["destination"], row["untrack"] = "../elsewhere/x", False
        plan_file.write_text(json.dumps(plan))
        git(repo, "tag", "archive/test")
        result = self.run_cmd("apply", plan_file)
        self.assertEqual(result.returncode, 1, result.stdout)
        self.assertIn("leave Git", result.stderr)
        self.assertTrue((repo / "_support/sessions/2026-01-02-x/a.md").exists())

    def test_the_outside_ignore_check_covers_only_the_workspace_and_data_homes(self):
        ws, specs = specs_workspace(self.tmp())
        git(ws, "init", "-q", "-b", "main")
        write(ws, ".gitignore", "specs-repo/\n_support/workspace/\n")
        plan = self.plan_json(ws, specs, extra=["--override", f"{specs}/evidence/loose.txt=../_support/other/loose.txt"])
        self.assertEqual(self.row(plan, "evidence/loose.txt", 1)["untrack"], True)
        self.assertFalse([r for r in plan["reviews"] if r["reason"] == "workspace not ignored"], plan["reviews"])

    def test_a_path_another_root_plans_is_skipped_without_a_review(self):
        ws, specs = specs_workspace(self.tmp())
        write(specs, "archive/2026-05-01-x/a.md", "# a\n")
        inner = specs / "archive/inner"
        inner.mkdir()
        git(inner, "init", "-q", "-b", "main")
        plain = specs / "archive/plain"
        write(plain, "p.md", "# p\n")
        plan = self.plan_json(ws, specs, inner, plain)
        self.assertFalse([r for r in plan["reviews"] if r["reason"] == "embedded repository"], plan["reviews"])
        self.assertFalse([r for r in plan["rows"] if r["root"] == 1 and r["source"] in ("archive/inner", "archive/plain")])


if __name__ == "__main__":
    unittest.main()
