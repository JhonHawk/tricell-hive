from pathlib import Path
import os
import json
import copy
import time
import shutil
import subprocess
import sys
import tempfile
import textwrap
import unittest


ROOT = Path(__file__).resolve().parents[2]
TASKS = ROOT / "content/skills/flow-plan/scripts/tasks.py"
sys.dont_write_bytecode = True  # keep __pycache__ out of the distributed skill folder

PLAN_MD = textwrap.dedent(
    """\
    # 2026-11-24-demo — demo plan

    ## Decisiones

    - **D1 A** — Old way.

    ## Alcance y aceptación

    - AC1 — one · false today · Observed in: here
    - AC2 — two · false today · Observed in: here

    ## Tareas

    ### T1 — First

    - [x] one
    - [x] two

    **Closes:** AC1

    ### T2 — Second thing

    - [ ] alpha
    - [ ] beta

    **Closes:** AC2

    ### T3 — Third

    - [ ] gamma

    ### T4 — Fourth

    - [ ] delta

    ## Review status and progress

    2026-11-08 — Plan written.
    """
)

PLAN_NAME = "2026-11-24-demo.plan.md"


class TasksScriptTest(unittest.TestCase):
    def run_tasks(self, folder, *args):
        return subprocess.run(
            [sys.executable, "-I", "-B", str(TASKS), str(folder), *args],
            text=True,
            capture_output=True,
            timeout=60,
        )

    def folder(self, tmp, tasks=PLAN_MD):
        """Write the plan file in its work folder under _support/workspace and return its path."""
        plan = Path(tmp) / "_support/workspace/2026-11-24-demo" / PLAN_NAME
        plan.parent.mkdir(parents=True)
        plan.write_text(tasks, encoding="utf-8")
        return plan

    def text(self, change):
        return change.read_text(encoding="utf-8")

    def revise(self, change, extra="changed requirement"):
        """Change the plan above its tasks, outside `## Checkpoint` and the progress section, as a revision does."""
        change.write_text(self.text(change).replace("## Alcance y aceptación", extra + "\n\n## Alcance y aceptación"), encoding="utf-8")

    def test_done_ticks_only_that_task_and_records_one_line(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            result = self.run_tasks(change, "done", "T2", "--commit", "abc1234", "--unverified", "UI walk")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(result.stdout.strip(), "✓ T2 · 2/4 · Second thing")
            text = self.text(change)
            self.assertIn("- [x] alpha\n- [x] beta", text)
            self.assertIn("- [ ] gamma", text)
            self.assertIn("- [ ] delta", text)
            lines = [l for l in text.splitlines() if "T2 · abc1234" in l]
            self.assertEqual(len(lines), 1)
            self.assertRegex(lines[0], r"^\d{4}-\d{2}-\d{2} — T2 · abc1234 · unverified: UI walk$")
            self.assertLess(text.index("Plan written."), text.index("T2 · abc1234"))

    def test_done_twice_does_not_duplicate_the_line(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            args = ("done", "T2", "--commit", "abc1234", "--unverified", "none")
            self.run_tasks(change, *args)
            first = self.text(change)
            result = self.run_tasks(change, *args)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(self.text(change), first)
            self.assertEqual(first.count("T2 · abc1234"), 1)

    def test_done_without_a_commit_under_hold(self):
        with tempfile.TemporaryDirectory() as tmp:
            empty = PLAN_MD.replace("2026-11-08 — Plan written.\n", "")
            change = self.folder(tmp, tasks=empty)
            result = self.run_tasks(change, "done", "T2", "--commit", "none", "--unverified", "login flow")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertRegex(self.text(change),
                             r"## Review status and progress\n\n\d{4}-\d{2}-\d{2} — T2 · uncommitted · unverified: login flow\n\Z",
                             "one blank line under the heading, none after the entry")
            result = self.run_tasks(change, "done", "T2", "--commit", "none", "--unverified", "none")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertRegex(self.text(change), r"unverified: login flow\n\d{4}-\d{2}-\d{2} — T2 · uncommitted · unverified: none\n\Z",
                             "a later result without a commit is recorded, not taken for a repeat")

    @unittest.skipIf(hasattr(os, "geteuid") and os.geteuid() == 0, "root writes into a read-only folder")
    def test_a_failed_write_prints_no_progress_line(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            change.parent.chmod(0o500)  # the temporary file next to the plan cannot be created
            try:
                result = self.run_tasks(change, "done", "T2", "--commit", "abc1234", "--unverified", "none")
            finally:
                change.parent.chmod(0o700)
            self.assertEqual(result.returncode, 2)
            self.assertEqual(result.stdout, "")
            self.assertEqual(self.text(change), PLAN_MD)

    def test_done_exits_2_for_a_missing_task_or_progress_section(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            result = self.run_tasks(change, "done", "T9", "--commit", "abc1234", "--unverified", "none")
            self.assertEqual(result.returncode, 2)
            self.assertEqual(self.text(change), PLAN_MD)
        with tempfile.TemporaryDirectory() as tmp:
            bare = PLAN_MD.split("## Review status")[0]
            change = self.folder(tmp, tasks=bare)
            result = self.run_tasks(change, "done", "T2", "--commit", "abc1234", "--unverified", "none")
            self.assertEqual(result.returncode, 2)
            self.assertEqual(self.text(change), bare)

    def test_done_rejects_a_missing_unverified_flag(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            result = self.run_tasks(change, "done", "T2", "--commit", "abc1234")
            self.assertEqual(result.returncode, 2)
            self.assertEqual(self.text(change), PLAN_MD)

    def checkpoint(self, change, stage, step, next_action, *extra):
        return self.run_tasks(
            change, "checkpoint", "--stage", stage, "--step", step, "--commit", "abc1234", "--next", next_action, *extra
        )

    def test_checkpoint_creates_replaces_and_keeps_evidence(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            first = self.checkpoint(change, "build", "T2", "run T3", "--check", "go test ./... · pass · abc1234")
            self.assertEqual(first.returncode, 0, first.stderr)
            text = self.text(change)
            self.assertLess(text.index("## Checkpoint"), text.index("## Review status and progress"))
            for fixed in ("- Stage: build · T2", "- Last commit: abc1234", "- Open findings: none", "- Next: run T3", "### Evidence"):
                self.assertIn(fixed, text)
            self.assertIn("- go test ./... · pass · abc1234", text)

            second = self.checkpoint(
                change, "verify", "round 1", "fix H1", "--finding", "H1: replaced D4", "--check", "lint · pass · def5678"
            )
            self.assertEqual(second.returncode, 0, second.stderr)
            text = self.text(change)
            self.assertEqual(text.count("## Checkpoint"), 1)
            self.assertEqual(text.count("- Stage:"), 1)
            self.assertIn("- Stage: verify · round 1", text)
            self.assertNotIn("run T3", text)
            self.assertIn("- Open findings: H1: replaced D4", text)
            self.assertIn("- go test ./... · pass · abc1234", text)
            self.assertIn("- lint · pass · def5678", text)
            self.assertIn("Plan written.", text)

    def test_checkpoint_keeps_the_worktree_until_a_new_one_is_given(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            first = self.checkpoint(change, "build", "T1", "run T2", "--worktree", "/repo/.claude/worktrees/w1 · feat/w1")
            self.assertEqual(first.returncode, 0, first.stderr)
            self.assertIn("- Worktree: /repo/.claude/worktrees/w1 · feat/w1", self.text(change))

            second = self.checkpoint(change, "build", "T2", "run T3")
            self.assertEqual(second.returncode, 0, second.stderr)
            text = self.text(change)
            self.assertEqual(text.count("- Worktree:"), 1)
            self.assertIn("- Worktree: /repo/.claude/worktrees/w1 · feat/w1", text)

            third = self.checkpoint(change, "build", "T3", "run T4", "--worktree", "/repo/.claude/worktrees/w2 · feat/w2")
            self.assertEqual(third.returncode, 0, third.stderr)
            text = self.text(change)
            self.assertEqual(text.count("- Worktree:"), 1)
            self.assertIn("- Worktree: /repo/.claude/worktrees/w2 · feat/w2", text)

            fourth = self.checkpoint(change, "close", "cleanup", "archive", "--worktree", "none")
            self.assertEqual(fourth.returncode, 0, fourth.stderr)
            self.assertNotIn("- Worktree:", self.text(change))

    def test_checkpoint_keeps_the_rules_line_until_a_new_one_is_given(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            first = self.checkpoint(change, "build", "T1", "run T2", "--rules", "flow-build/stage.md · extra: defect-fix")
            self.assertEqual(first.returncode, 0, first.stderr)
            second = self.checkpoint(change, "build", "T2", "run T3")
            self.assertEqual(second.returncode, 0, second.stderr)
            text = self.text(change)
            self.assertEqual(text.count("- Rules:"), 1)
            self.assertIn("- Rules: flow-build/stage.md · extra: defect-fix", text)
            third = self.checkpoint(change, "verify", "1", "walk", "--rules", "flow-verify/stage-entry.md")
            self.assertEqual(third.returncode, 0, third.stderr)
            self.assertIn("- Rules: flow-verify/stage-entry.md\n", self.text(change))
            both = self.checkpoint(change, "verify", "2", "reviews", "--rules", "flow-verify/stage-entry.md, flow-verify/stage-exit.md")
            self.assertEqual(both.returncode, 0, both.stderr)
            self.assertIn("- Rules: flow-verify/stage-entry.md, flow-verify/stage-exit.md\n", self.text(change))
            bad = self.checkpoint(change, "verify", "1", "walk", "--rules", "stage.md")
            self.assertEqual(bad.returncode, 2)

    def test_checkpoint_rules_merge_within_a_flow_and_reset_on_a_new_one(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            steps = [
                ("flow-build/stage.md · extra: defect-fix", "- Rules: flow-build/stage.md · extra: defect-fix"),
                ("flow-verify/stage-entry.md", "- Rules: flow-verify/stage-entry.md"),
                ("flow-verify/stage-exit.md · extra: tests", "- Rules: flow-verify/stage-entry.md, flow-verify/stage-exit.md · extra: tests"),
                ("flow-verify/stage-exit.md · extra: defect-fix", "- Rules: flow-verify/stage-entry.md, flow-verify/stage-exit.md · extra: tests, defect-fix"),
            ]
            for n, (value, want) in enumerate(steps, 1):
                result = self.checkpoint(change, "verify", str(n), "next", "--rules", value)
                self.assertEqual(result.returncode, 0, result.stderr)
                text = self.text(change)
                self.assertEqual(text.count("- Rules:"), 1, text)
                self.assertIn(want + "\n", text)
            removed = self.checkpoint(change, "close", "5", "next", "--rules", "none")
            self.assertEqual(removed.returncode, 0, removed.stderr)
            self.assertNotIn("- Rules:", self.text(change))

    def test_checkpoint_merges_principles_and_never_removes_them(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            steps = [
                (["--principles", "none"], "- Principles: none"),
                (["--principles", "boundary-discipline"], "- Principles: boundary-discipline"),
                ([], "- Principles: boundary-discipline"),
                (["--principles", "model-the-domain, boundary-discipline"], "- Principles: boundary-discipline, model-the-domain"),
                (["--principles", "none"], "- Principles: boundary-discipline, model-the-domain"),
            ]
            for n, (extra, want) in enumerate(steps, 1):
                result = self.checkpoint(change, "build", f"T{n}", "next", *extra)
                self.assertEqual(result.returncode, 0, result.stderr)
                text = self.text(change)
                self.assertEqual(text.count("- Principles:"), 1, text)
                self.assertIn(want + "\n", text)
            for bad in ("", "Boundary", "none, model-the-domain", "a,b"):
                result = self.checkpoint(change, "build", "T9", "next", "--principles", bad)
                self.assertEqual(result.returncode, 2, bad)

    def test_checkpoint_keeps_unwritten_until_replaced_or_none(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            first = self.checkpoint(change, "build", "T1", "run T2", "--unwritten", "the retry idea for T4")
            self.assertEqual(first.returncode, 0, first.stderr)
            self.assertIn("- Unwritten: the retry idea for T4", self.text(change))

            second = self.checkpoint(change, "build", "T2", "run T3")
            self.assertEqual(second.returncode, 0, second.stderr)
            text = self.text(change)
            self.assertEqual(text.count("- Unwritten:"), 1)
            self.assertIn("- Unwritten: the retry idea for T4", text)

            third = self.checkpoint(change, "build", "T3", "run T4", "--unwritten", "another thought")
            text = self.text(change)
            self.assertEqual(text.count("- Unwritten:"), 1)
            self.assertIn("- Unwritten: another thought", text)

            fourth = self.checkpoint(change, "build", "T4", "close", "--unwritten", "none")
            self.assertEqual(fourth.returncode, 0, fourth.stderr)
            self.assertNotIn("Unwritten", self.text(change))

    def test_checkpoint_without_unwritten_is_unchanged(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            self.checkpoint(change, "build", "T1", "run T2")
            self.assertNotIn("Unwritten", self.text(change))

    def test_done_never_ticks_delivery_unit_boxes(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            path = change
            text = path.read_text().replace(
                "## Review status",
                "## Delivery units\n\n### E1 — PR 1 · depende de: ninguna · archivos: a.py\n\n"
                "- [ ] unit: make test\n- [ ] live: go · pass: ok\n\n## Review status",
            )
            path.write_text(text)
            done = self.run_tasks(change, "done", "T4", "--commit", "abc1234", "--unverified", "none")
            self.assertEqual(done.returncode, 0, done.stderr)
            after = path.read_text()
            self.assertIn("- [ ] unit: make test", after)
            self.assertIn("- [ ] live: go · pass: ok", after)
            self.assertIn("- [x] delta", after)

    def test_check_change_exits_0_when_the_checkpoint_names_a_replaced_decision(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            change.write_text(self.text(change).replace(
                "- **D1 A** — Old way.", "- **D1 A** — Old way.\n- **D2 A** — New way. **Replaces:** D1-A"), encoding="utf-8")
            self.checkpoint(change, "build", "T2", "run T3", "--finding", "H1: keep D1-A")
            self.assertIn("D1-A", self.text(change))
            check = ROOT / "content/skills/flow-plan/scripts/check_change.py"
            result = subprocess.run(
                [sys.executable, "-I", "-B", str(check), str(change)], text=True, capture_output=True
            )
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_a_folder_argument_is_refused_and_writes_nothing(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            result = self.run_tasks(change.parent, "done", "T2", "--commit", "abc1234", "--unverified", "none")
            self.assertEqual(result.returncode, 2)
            self.assertEqual(self.text(change), PLAN_MD)

    def test_removed_subcommands_and_options_are_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            for args in (("status",), ("versioned",),
                         ("checkpoint", "--stage", "b", "--step", "s", "--commit", "abc1234", "--next", "n", "--out", str(Path(tmp) / "o.md"))):
                with self.subTest(args=args):
                    self.assertEqual(self.run_tasks(change, *args).returncode, 2)
            self.assertEqual(self.text(change), PLAN_MD)

    def test_a_plan_on_the_base_branch_still_takes_writes(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            subprocess.run(["git", "-C", tmp, "init", "-q", "-b", "main"], check=True)
            done = self.run_tasks(change, "done", "T4", "--commit", "abc1234", "--unverified", "none")
            self.assertEqual(done.returncode, 0, done.stdout + done.stderr)
            self.assertEqual(self.checkpoint(change, "build", "T4", "next").returncode, 0)


    def test_status_set_writes_the_header_status_value(self):
        plan_text = "| Field | Value |\n| --- | --- |\n| Status | In build |\n| Tracker | none |\n\n" + PLAN_MD
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp, tasks=plan_text)
            result = self.run_tasks(change, "status-set", "Completed")
            self.assertEqual(result.returncode, 0, result.stderr)
            text = self.text(change)
            self.assertIn("| Status | Completed |\n| Tracker | none |", text)
            self.assertEqual(self.run_tasks(change, "status-set", "Done").returncode, 2)
            self.assertEqual(self.run_tasks(self.folder(Path(tmp) / "x"), "status-set", "Completed").returncode, 2)


class ExecutionRecordTest(unittest.TestCase):
    """The CLI owns this persisted contract; fixtures use real Git and files."""

    run_tasks = TasksScriptTest.run_tasks
    folder = TasksScriptTest.folder
    text = TasksScriptTest.text
    checkpoint = TasksScriptTest.checkpoint
    revise = TasksScriptTest.revise

    def git(self, repo, *args):
        result = subprocess.run(["git", "-C", str(repo), *args], text=True, capture_output=True,
                                env={**os.environ, "GIT_CONFIG_NOSYSTEM": "1"}, timeout=20)
        self.assertEqual(result.returncode, 0, result.stderr)
        return result.stdout.strip()

    def coordinator_repo(self, tmp, ignore=True):
        coordinator = Path(tmp) / "coordinator"
        coordinator.mkdir()
        self.git(coordinator, "init", "-b", "main")
        self.git(coordinator, "config", "user.email", "fixture@example.invalid")
        self.git(coordinator, "config", "user.name", "Fixture")
        (coordinator / ".gitignore").write_text("_support/workspace/\n" if ignore else "\n")
        self.git(coordinator, "add", ".gitignore")
        self.git(coordinator, "commit", "-m", "base")
        return coordinator

    def setup_execution(self, tmp, role="hive-build-owner", record_root=None, criteria=("AC2",), quoted=False, expect=0):
        repo = Path(tmp) / "repo"
        repo.mkdir()
        self.git(repo, "init", "-b", "main")
        self.git(repo, "config", "user.email", "fixture@example.invalid")
        self.git(repo, "config", "user.name", "Fixture")
        (repo / ".gitignore").write_text("_support/workspace/\n")
        (repo / "a.txt").write_text("base\n")
        (repo / "b.txt").write_text("base\n")
        self.git(repo, "add", ".")
        self.git(repo, "commit", "-m", "base")
        base = self.git(repo, "rev-parse", "HEAD")
        change = self.folder(repo)
        if quoted:
            change.write_text(self.text(change).replace("## Alcance y aceptación", "## Request\n\n- AC9 — quoted from the ticket\n\n## Alcance y aceptación"), encoding="utf-8")
        base = self.git(repo, "rev-parse", "HEAD")
        worktree = Path(tmp) / "owner"
        self.git(repo, "worktree", "add", "-b", "unit-a", str(worktree), base)
        record = (record_root or repo) / "_support/workspace/run/execution.json"
        record.parent.mkdir(parents=True)
        (record.parent / "check.txt").write_text("observed command exit 0\n")
        unit = dict(id="U1", tasks=["T2"], criteria=list(criteria), depends_on=[],
                    assignment_id="A1", owner_id="owner-1", role=role,
                    repo=str(repo), worktree=str(worktree), branch="unit-a", base_sha=base,
                    allowed_paths=["a.txt"], resources=["port:fixture-a"], state="assigned",
                    returned_head=None, evidence=[], verifier=None, limitations=[],
                    next_action="implement", cleanup=None)
        manifest = dict(grant_ref=f"{PLAN_NAME}#decisiones",
                        integration_target=dict(repo=str(repo), worktree=str(repo), branch="main"),
                        units=[unit], publication=None)
        source = Path(tmp) / "input.json"
        source.write_text(json.dumps(manifest))
        result = self.run_tasks(change, "execution", "init", "--record", str(record),
                                "--input", str(source), "--coordinator", "root")
        self.assertEqual(result.returncode, expect, result.stderr)
        return change, record, source, repo, worktree

    def operation(self, change, record, operation, *extra):
        return self.run_tasks(change, "execution", operation, "--record", str(record), *extra)

    def update_record(self, change, record, source, mutate, expected=None):
        data = json.loads(record.read_text())
        generation = data["generation"] if expected is None else expected
        mutate(data)
        source.write_text(json.dumps(data))
        return self.operation(change, record, "update", "--input", str(source),
                              "--expected-generation", str(generation), "--coordinator", "root")

    def returned(self, change, record, source, worktree):
        result = self.update_record(change, record, source, lambda d: d["units"][0].update(state="running"))
        self.assertEqual(result.returncode, 0, result.stderr)
        (worktree / "a.txt").write_text("implemented\n")
        self.git(worktree, "add", "a.txt")
        self.git(worktree, "commit", "-m", "implement")
        head = self.git(worktree, "rev-parse", "HEAD")
        result = self.update_record(change, record, source, lambda d: d["units"][0].update(
            state="returned", returned_head=head, evidence=["check.txt"]))
        self.assertEqual(result.returncode, 0, result.stderr)
        return head

    def verdict(self, head):
        return dict(id="verifier-1", head=head, criteria=["AC2"],
                    checks=[dict(command="fixture-check", result="pass", path="check.txt")])

    def test_execution_ignores_criteria_quoted_outside_the_acceptance_section(self):
        with tempfile.TemporaryDirectory() as tmp:
            self.setup_execution(tmp, criteria=("AC9",), quoted=True, expect=2)

    def test_execution_accepts_the_two_owner_roles_and_rejects_others(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, _, _ = self.setup_execution(tmp, role="hive-ticket-owner")
            self.assertEqual(self.operation(change, record, "check").returncode, 0)
            manifest = json.loads(source.read_text())
            manifest["units"][0]["role"] = "hive-build-backend"
            source.write_text(json.dumps(manifest))
            other = record.parent.parent / "run-2" / "execution.json"
            other.parent.mkdir()
            result = self.run_tasks(change, "execution", "init", "--record", str(other),
                                    "--input", str(source), "--coordinator", "root")
            self.assertEqual(result.returncode, 2, result.stdout)
            self.assertIn("unit role", result.stdout + result.stderr)
            self.assertFalse(other.exists())

    def test_execution_binds_grant_and_record_to_the_plan(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, repo, _ = self.setup_execution(tmp)
            manifest = json.loads(source.read_text())
            data = json.loads(record.read_text())
            self.assertEqual(data["plan_ref"], str(change.resolve()))
            self.assertEqual(data["grant_ref"], f"{PLAN_NAME}#decisiones")
            cases = {
                "grant in another file": dict(grant_ref="other.md#decisiones"),
                "grant outside the folder": dict(grant_ref=f"../{PLAN_NAME}#decisiones"),
            }
            for name, patch in cases.items():
                with self.subTest(name):
                    (change.parent / "other.md").write_text("# other\n")
                    source.write_text(json.dumps({**manifest, **patch}))
                    other = record.parent.parent / name.replace(" ", "-") / "execution.json"
                    other.parent.mkdir()
                    result = self.run_tasks(change, "execution", "init", "--record", str(other),
                                            "--input", str(source), "--coordinator", "root")
                    self.assertEqual(result.returncode, 2, result.stdout)
                    self.assertFalse(other.exists())
            inside = change.parent / "execution.json"
            source.write_text(json.dumps(manifest))
            result = self.run_tasks(change, "execution", "init", "--record", str(inside),
                                    "--input", str(source), "--coordinator", "root")
            self.assertEqual(result.returncode, 2, result.stdout)
            self.assertIn("outside the plan's work folder", result.stderr)
            self.assertFalse(inside.exists())

    def test_execution_record_lives_in_the_coordinator_checkout(self):
        with tempfile.TemporaryDirectory() as tmp:
            coordinator = self.coordinator_repo(tmp)
            change, record, source, _, worktree = self.setup_execution(
                tmp, role="hive-ticket-owner", record_root=coordinator)
            self.assertIn("- Execution record:", self.text(change))
            self.assertEqual(self.operation(change, record, "check").returncode, 0)
            self.returned(change, record, source, worktree)
            result = self.operation(change, record, "validate-handoff", "--unit", "U1")
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            check = ROOT / "content/skills/flow-plan/scripts/check_change.py"
            result = subprocess.run([sys.executable, "-I", "-B", str(check), str(change)],
                                    text=True, capture_output=True, timeout=60)
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_execution_record_in_another_checkout_keeps_location_rules(self):
        cases = {
            "not ignored": dict(ignore=False, relative="_support/workspace/run/execution.json", track=False),
            "tracked": dict(ignore=True, relative="_support/workspace/run/execution.json", track=True),
            "outside workspace": dict(ignore=True, relative="notes/run/execution.json", track=False),
        }
        for name, case in cases.items():
            with self.subTest(name), tempfile.TemporaryDirectory() as tmp:
                coordinator = self.coordinator_repo(tmp, ignore=case["ignore"])
                change, _, source, _, _ = self.setup_execution(tmp)
                record = coordinator / case["relative"]
                record.parent.mkdir(parents=True)
                if case["track"]:
                    record.write_text("{}\n")
                    self.git(coordinator, "add", "-f", case["relative"])
                    self.git(coordinator, "commit", "-m", "track record")
                    record.unlink()
                result = self.run_tasks(change, "execution", "init", "--record", str(record),
                                        "--input", str(source), "--coordinator", "root")
                self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
                self.assertFalse(record.exists())

    def test_execution_record_behind_a_symlink_is_refused(self):
        with tempfile.TemporaryDirectory() as tmp:
            coordinator = self.coordinator_repo(tmp)
            change, _, source, _, _ = self.setup_execution(tmp)
            outside = Path(tmp) / "outside"
            outside.mkdir()
            workspace = coordinator / "_support/workspace"
            workspace.mkdir(parents=True)
            (workspace / "run").symlink_to(outside, target_is_directory=True)
            record = workspace / "run" / "execution.json"
            result = self.run_tasks(change, "execution", "init", "--record", str(record),
                                    "--input", str(source), "--coordinator", "root")
            self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
            self.assertFalse((outside / "execution.json").exists())

    def test_execution_init_links_checkpoint_without_overwrite(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, _, _ = self.setup_execution(tmp)
            self.assertIn("- Execution record:", self.text(change))
            self.assertEqual(json.loads(record.read_text())["generation"], 0)
            before = record.read_bytes()
            result = self.operation(change, record, "init", "--input", str(source), "--coordinator", "root")
            self.assertEqual(result.returncode, 2)
            self.assertEqual(record.read_bytes(), before)
            self.assertEqual(self.operation(change, record, "check").returncode, 0)
            self.checkpoint(change, "build", "T2", "continue")
            self.assertIn("- Execution record:", self.text(change))

    def test_check_change_validates_checkpoint_execution_pointer(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, _, _, _ = self.setup_execution(tmp)
            check = ROOT / "content/skills/flow-plan/scripts/check_change.py"
            command = [sys.executable, "-I", "-B", str(check), str(change)]
            result = subprocess.run(command, text=True, capture_output=True, timeout=60)
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            data = json.loads(record.read_text())
            data["schema_version"] = 99
            record.write_text(json.dumps(data))
            result = subprocess.run(command, text=True, capture_output=True, timeout=60)
            self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
            self.assertIn("execution", result.stdout)

    def test_execution_invalid_structure_and_conflicts_preserve_record(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, _, _ = self.setup_execution(tmp)
            mutations = [
                lambda d: d.update(schema_version=99),
                lambda d: d["units"][0].update(tasks=["T99"]),
                lambda d: d["units"][0].update(criteria=["AC99"]),
                lambda d: d["units"][0].update(depends_on=["U99"]),
                lambda d: d["units"][0].update(depends_on=["U1"]),
                lambda d: d["units"].append(copy.deepcopy(d["units"][0])),
                lambda d: d["units"].append({**d["units"][0], "id": "U2", "assignment_id": "A2"}),
            ]
            for mutate in mutations:
                with self.subTest(mutate=mutate):
                    before = record.read_bytes()
                    result = self.update_record(change, record, source, mutate)
                    self.assertEqual(result.returncode, 2, result.stdout)
                    self.assertEqual(record.read_bytes(), before)
            source.write_text('{"units": [], "units": []}')
            result = self.operation(change, record, "update", "--input", str(source),
                                    "--expected-generation", "0", "--coordinator", "root")
            self.assertEqual(result.returncode, 2)

    def test_execution_stale_generation_writer_lock_and_plan_drift(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, _, _ = self.setup_execution(tmp)
            result = self.update_record(change, record, source, lambda d: None)
            self.assertEqual(result.returncode, 0, result.stderr)
            before = record.read_bytes()
            result = self.update_record(change, record, source, lambda d: None, expected=0)
            self.assertEqual(result.returncode, 2)
            self.assertIn("generation", result.stderr)
            lock = record.with_name(record.name + ".lock")
            lock.mkdir()
            result = self.update_record(change, record, source, lambda d: None)
            self.assertEqual(result.returncode, 2)
            self.assertTrue(lock.is_dir())
            self.assertEqual(record.read_bytes(), before)
            lock.rmdir()
            self.revise(change)
            result = self.operation(change, record, "check")
            self.assertEqual(result.returncode, 2)
            self.assertIn("revision", result.stderr)

    def test_execution_handoff_and_exact_head_evidence(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, _, worktree = self.setup_execution(tmp)
            head = self.returned(change, record, source, worktree)
            valid = self.operation(change, record, "validate-handoff", "--unit", "U1")
            self.assertEqual(valid.returncode, 0, valid.stderr)
            for bad in (None, self.verdict("0" * 40), {**self.verdict(head), "id": "owner-1"},
                        {**self.verdict(head), "checks": []},
                        {**self.verdict(head), "checks": [dict(command="check", result="fail", path="check.txt")]},
                        {**self.verdict(head), "checks": [dict(command="check", result="pass", path="missing.txt")]}):
                result = self.update_record(change, record, source, lambda d: d["units"][0].update(state="verified", verifier=bad))
                self.assertEqual(result.returncode, 2, result.stdout)
            result = self.update_record(change, record, source, lambda d: d["units"][0].update(
                state="verified", verifier=self.verdict(head)))
            self.assertEqual(result.returncode, 0, result.stderr)
            result = self.update_record(change, record, source, lambda d: d["units"][0].update(state="running"))
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIsNone(json.loads(record.read_text())["units"][0]["verifier"])

    def test_execution_dirty_branch_and_out_of_scope_handoffs(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, _, worktree = self.setup_execution(tmp)
            self.returned(change, record, source, worktree)
            (worktree / "b.txt").write_text("dirty\n")
            result = self.operation(change, record, "validate-handoff", "--unit", "U1")
            self.assertEqual(result.returncode, 2)
            self.assertIn("clean", result.stderr)
            self.git(worktree, "add", "b.txt")
            self.git(worktree, "commit", "-m", "outside")
            result = self.update_record(change, record, source, lambda d: d["units"][0].update(
                returned_head=self.git(worktree, "rev-parse", "HEAD")))
            self.assertEqual(result.returncode, 0, result.stderr)
            result = self.operation(change, record, "validate-handoff", "--unit", "U1")
            self.assertEqual(result.returncode, 2)
            self.assertIn("allowed", result.stderr)
            self.git(worktree, "branch", "-m", "wrong")
            result = self.operation(change, record, "validate-handoff", "--unit", "U1")
            self.assertEqual(result.returncode, 2)
            self.assertIn("branch", result.stderr)

    def test_execution_replacement_requires_confirmed_closure_preserves_history(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, _, worktree = self.setup_execution(tmp)
            head = self.returned(change, record, source, worktree)
            def replace(d):
                d["units"][0].update(assignment_id="A2", owner_id="owner-2", state="assigned")
            result = self.update_record(change, record, source, replace)
            self.assertEqual(result.returncode, 2)
            def closed_replace(d):
                replace(d)
                d["units"][0]["closure"] = dict(assignment_id="A1", owner_id="owner-1", kind="completed", path="check.txt")
            result = self.update_record(change, record, source, closed_replace)
            self.assertEqual(result.returncode, 0, result.stderr)
            unit = json.loads(record.read_text())["units"][0]
            self.assertEqual(unit["history"][0]["returned_head"], head)
            self.assertEqual(unit["history"][0]["assignment_id"], "A1")
            self.assertIsNone(unit["verifier"])
            for state in ("running", "returned"):
                result = self.update_record(change, record, source, lambda d: d["units"][0].update(state=state))
                self.assertEqual(result.returncode, 0, result.stderr)
            before = record.read_bytes()
            # The replacement still returns the former owner's partial head.
            prior_author_verdict = {**self.verdict(head), "id": "owner-1"}
            result = self.update_record(change, record, source, lambda d: d["units"][0].update(
                state="verified", verifier=prior_author_verdict))
            self.assertEqual(result.returncode, 2, result.stdout)
            self.assertIn("verifier must differ", result.stderr)
            self.assertEqual(record.read_bytes(), before)
            result = self.update_record(change, record, source, lambda d: d["units"][0].update(
                state="verified", verifier=self.verdict(head)))
            self.assertEqual(result.returncode, 0, result.stderr)

    def test_execution_concurrent_and_interrupted_writers_keep_prior_state(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, _, _ = self.setup_execution(tmp)
            source.write_text(record.read_text())
            command = [sys.executable, "-I", "-B", str(TASKS), str(change), "execution", "update",
                       "--record", str(record), "--input", str(source), "--coordinator", "root",
                       "--expected-generation", "0"]
            processes = [subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
                         for _ in range(2)]
            try:
                results = [p.communicate(timeout=30) for p in processes]
                self.assertEqual(sorted(p.returncode for p in processes), [0, 2], results)
                self.assertEqual(json.loads(record.read_text())["generation"], 1)
            finally:
                for proc in processes:
                    if proc.poll() is None:
                        proc.kill()
                        proc.wait(timeout=5)
            before = record.read_bytes()
            fifo = Path(tmp) / "blocked-input"
            os.mkfifo(fifo)
            command[command.index(str(source))] = str(fifo)
            command[-1] = "1"
            proc = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            lock = record.with_name(record.name + ".lock")
            try:
                deadline = time.monotonic() + 5
                while not lock.exists() and proc.poll() is None and time.monotonic() < deadline:
                    time.sleep(0.01)
                self.assertTrue(lock.is_dir(), "writer never acquired lock")
                proc.terminate()
                proc.communicate(timeout=5)
                self.assertTrue(lock.is_dir(), "interrupted ownership must require reconciliation")
                self.assertEqual(record.read_bytes(), before)
            finally:
                if proc.poll() is None:
                    proc.kill()
                    proc.communicate(timeout=5)

    def test_execution_rejects_escaped_storage_evidence_and_other_writer(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, _, _ = self.setup_execution(tmp)
            source.write_text(record.read_text())
            result = self.operation(change, record, "update", "--input", str(source),
                                    "--coordinator", "owner-1", "--expected-generation", "0")
            self.assertEqual(result.returncode, 2)
            escaped = record.parent / "escape.txt"
            escaped.symlink_to(source)
            result = self.update_record(change, record, source, lambda d: d["units"][0].update(evidence=["escape.txt"]))
            self.assertEqual(result.returncode, 2)
            outside = Path(tmp) / "execution.json"
            result = self.operation(change, outside, "init", "--input", str(source), "--coordinator", "root")
            self.assertEqual(result.returncode, 2)
            self.assertFalse(outside.exists())

    def test_execution_rejects_cycles_paths_branches_and_exclusive_resources(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, repo, _ = self.setup_execution(tmp)
            second = Path(tmp) / "second"
            base = json.loads(record.read_text())["units"][0]["base_sha"]
            self.git(repo, "worktree", "add", "-b", "unit-b", str(second), base)
            data = json.loads(record.read_text())
            other = {**data["units"][0], "id": "U2", "assignment_id": "A2", "owner_id": "owner-2",
                     "worktree": str(second), "branch": "unit-b", "allowed_paths": ["b.txt"], "resources": ["port:b"]}
            data["units"].append(other)
            for mutate, message in (
                (lambda d: d["units"][1].update(allowed_paths=["a.txt"]), "paths"),
                (lambda d: d["units"][1].update(branch="unit-a"), "branches"),
                (lambda d: d["units"][1].update(resources=["port:fixture-a"]), "resources"),
                (lambda d: (d["units"][0].update(depends_on=["U2"]), d["units"][1].update(depends_on=["U1"])), "cycle"),
            ):
                candidate = copy.deepcopy(data)
                mutate(candidate)
                new_record = record.parent / "case" / "execution.json"
                new_record.parent.mkdir(exist_ok=True)
                source.write_text(json.dumps(candidate))
                result = self.operation(change, new_record, "init", "--input", str(source), "--coordinator", "root")
                self.assertEqual(result.returncode, 2, result.stdout)
                self.assertIn(message, result.stderr)
                self.assertFalse(new_record.exists())

    def test_execution_reconciles_changed_plan_without_reusing_verdict(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, _, worktree = self.setup_execution(tmp)
            head = self.returned(change, record, source, worktree)
            result = self.update_record(change, record, source, lambda d: d["units"][0].update(
                state="verified", verifier=self.verdict(head)))
            self.assertEqual(result.returncode, 0, result.stderr)
            previous_revision = json.loads(record.read_text())["plan_revision"]
            change.write_text(self.text(change).replace("AC2", "AC3"), encoding="utf-8")
            data = json.loads(record.read_text())
            generation = data["generation"]
            data["units"][0].update(state="blocked", criteria=["AC3"])
            source.write_text(json.dumps(data))
            result = self.operation(change, record, "update", "--input", str(source), "--coordinator", "root",
                                    "--expected-generation", str(generation), "--reconcile-plan")
            self.assertEqual(result.returncode, 0, result.stderr)
            data = json.loads(record.read_text())
            self.assertNotEqual(previous_revision, data["plan_revision"])
            self.assertIsNone(data["units"][0]["verifier"])
            self.assertEqual(data["units"][0]["state"], "blocked")
            self.assertEqual(self.operation(change, record, "check").returncode, 0)

    def test_execution_plan_reconciliation_preserves_writer_and_resource_boundaries(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, repo, worktree = self.setup_execution(tmp)
            self.revise(change, "Updated design context.")
            before = record.read_bytes()
            original = json.loads(before)
            mutations = {
                "owner_id": "replacement-without-closure",
                "role": "another-role",
                "repo": str(worktree),
                "worktree": str(repo),
                "branch": "replacement-branch",
                "base_sha": "0" * 40,
                "allowed_paths": ["b.txt"],
                "resources": ["replacement-resource"],
            }
            for field, value in mutations.items():
                with self.subTest(field=field):
                    record.write_bytes(before)
                    candidate = copy.deepcopy(original)
                    candidate["units"][0].update(state="blocked", **{field: value})
                    source.write_text(json.dumps(candidate))
                    result = self.operation(change, record, "update", "--input", str(source),
                                            "--coordinator", "root", "--expected-generation", "0", "--reconcile-plan")
                    self.assertEqual(result.returncode, 2, result.stdout)
                    self.assertIn(f"assignment {field} changed without replacement", result.stderr)
                    self.assertEqual(record.read_bytes(), before)
            candidate = copy.deepcopy(original)
            candidate["units"][0]["state"] = "blocked"
            source.write_text(json.dumps(candidate))
            result = self.operation(change, record, "update", "--input", str(source), "--coordinator", "root",
                                    "--expected-generation", "0", "--reconcile-plan")
            self.assertEqual(result.returncode, 0, result.stderr)
            def confirmed_replacement(data):
                data["units"][0].update(
                    state="assigned", assignment_id="A2", owner_id="replacement-with-closure",
                    allowed_paths=["b.txt"], resources=["replacement-resource"],
                    closure=dict(assignment_id="A1", owner_id="owner-1", kind="stop_ack", path="check.txt"))
            result = self.update_record(change, record, source, confirmed_replacement)
            self.assertEqual(result.returncode, 0, result.stderr)
            unit = json.loads(record.read_text())["units"][0]
            self.assertEqual(unit["allowed_paths"], ["b.txt"])
            self.assertEqual(unit["history"][0]["allowed_paths"], ["a.txt"])
            self.assertEqual(unit["history"][0]["owner_id"], "owner-1")

    def test_execution_rejects_plan_changed_during_transition(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, repo, worktree = self.setup_execution(tmp)
            integration = Path(tmp) / "integration"
            self.git(repo, "worktree", "add", "-b", "joint", str(integration), "HEAD")
            # Configure a separate local target so editing the plan does not dirty the joint checkout.
            data = json.loads(record.read_text())
            data["integration_target"].update(worktree=str(integration), branch="joint")
            record.write_text(json.dumps(data))
            head = self.returned(change, record, source, worktree)
            result = self.update_record(change, record, source, lambda d: d["units"][0].update(
                state="verified", verifier=self.verdict(head)))
            self.assertEqual(result.returncode, 0, result.stderr)
            self.git(integration, "merge", "--ff-only", "unit-a")
            data = json.loads(record.read_text())
            before = record.read_bytes()
            generation = data["generation"]
            data["units"][0]["state"] = "integrated"
            data.update(integration_scope="local", integration_head=head,
                        joint_check=dict(head=head, command="joint", result="pass", path="check.txt"))
            source.write_text(json.dumps(data))
            gate, release = Path(tmp) / "git-gate", Path(tmp) / "git-release"
            shim_dir = Path(tmp) / "bin"
            shim_dir.mkdir()
            shim = shim_dir / "git"
            real_git = shutil.which("git")
            shim.write_text(
                f"#!{sys.executable}\nimport os, sys, time\nfrom pathlib import Path\n"
                f"gate, release = Path({str(gate)!r}), Path({str(release)!r})\n"
                "if 'symbolic-ref' in sys.argv and not gate.exists():\n"
                "    gate.touch()\n    deadline = time.monotonic() + 10\n"
                "    while not release.exists() and time.monotonic() < deadline:\n        time.sleep(0.01)\n"
                "    if not release.exists(): sys.exit(75)\n"
                f"os.execv({real_git!r}, [{real_git!r}, *sys.argv[1:]])\n")
            shim.chmod(0o755)
            command = [sys.executable, "-I", "-B", str(TASKS), str(change), "execution", "update",
                       "--record", str(record), "--input", str(source), "--coordinator", "root",
                       "--expected-generation", str(generation)]
            proc = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True,
                                    env={**os.environ, "PATH": str(shim_dir) + os.pathsep + os.environ["PATH"]})
            try:
                deadline = time.monotonic() + 5
                while not gate.exists() and proc.poll() is None and time.monotonic() < deadline:
                    time.sleep(0.01)
                self.assertTrue(gate.exists(), "transition never reached the external Git boundary")
                self.revise(change, "Updated during transition.")
                release.touch()
                out, err = proc.communicate(timeout=30)
                self.assertEqual(proc.returncode, 2, out + err)
                self.assertIn("revision", err)
                self.assertEqual(record.read_bytes(), before)
            finally:
                if proc.poll() is None:
                    proc.kill()
                    proc.communicate(timeout=5)

    def test_execution_can_block_after_old_evidence_becomes_unavailable(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, _, worktree = self.setup_execution(tmp)
            self.returned(change, record, source, worktree)
            (record.parent / "check.txt").unlink()
            result = self.operation(change, record, "check")
            self.assertEqual(result.returncode, 2)
            result = self.update_record(change, record, source, lambda d: d["units"][0].update(
                state="blocked", evidence=[], verifier=None, limitations=["Prior evidence unavailable"],
                next_action="Recover native output before verifying"))
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(json.loads(record.read_text())["units"][0]["state"], "blocked")

    def test_execution_shape_rejects_invalid_optional_integration_fields(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, _, _ = self.setup_execution(tmp)
            for fields in ({"integration_scope": "invented"}, {"integration_head": "abc1234"}):
                result = self.update_record(change, record, source, lambda d: d.update(fields))
                self.assertEqual(result.returncode, 2, result.stdout)

    def test_execution_handoff_wrong_head_ancestry_and_repository(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, repo, worktree = self.setup_execution(tmp)
            self.returned(change, record, source, worktree)
            data = json.loads(record.read_text())
            before = record.read_text()
            orphan = self.git(worktree, "commit-tree", "HEAD^{tree}", "-m", "unrelated base")
            cases = [({"returned_head": data["units"][0]["base_sha"]}, "head"),
                     ({"base_sha": orphan}, "descend")]
            foreign = Path(tmp) / "foreign"
            foreign.mkdir()
            self.git(foreign, "init", "-b", "main")
            cases.append(({"repo": str(foreign)}, "repository"))
            for fields, message in cases:
                candidate = json.loads(before)
                candidate["units"][0].update(fields)
                record.write_text(json.dumps(candidate))
                result = self.operation(change, record, "validate-handoff", "--unit", "U1")
                self.assertEqual(result.returncode, 2)
                self.assertIn(message, result.stderr)
            record.write_text(before)

    def test_execution_integrated_requires_real_joint_head_and_result(self):
        with tempfile.TemporaryDirectory() as tmp:
            change, record, source, repo, worktree = self.setup_execution(tmp)
            head = self.returned(change, record, source, worktree)
            result = self.update_record(change, record, source, lambda d: d["units"][0].update(
                state="verified", verifier=self.verdict(head)))
            self.assertEqual(result.returncode, 0, result.stderr)
            def integrate(d):
                d["units"][0]["state"] = "integrated"
                d.update(integration_scope="local", integration_head=head,
                         joint_check=dict(command="joint", result="pass", path="check.txt", head=head))
            result = self.update_record(change, record, source, integrate)
            self.assertEqual(result.returncode, 2)
            def missing_owner(d):
                integrate(d)
                current = self.git(repo, "rev-parse", "HEAD")
                d["integration_head"] = current
                d["joint_check"]["head"] = current
            result = self.update_record(change, record, source, missing_owner)
            self.assertEqual(result.returncode, 2)
            self.assertIn("does not contain", result.stderr)
            self.git(repo, "merge", "--no-edit", "unit-a")
            integrated_head = self.git(repo, "rev-parse", "HEAD")
            original_integrate = integrate
            def integrate(d):
                original_integrate(d)
                d["integration_head"] = integrated_head
                d["joint_check"]["head"] = integrated_head
            result = self.update_record(change, record, source, integrate)
            self.assertEqual(result.returncode, 0, result.stderr)
            data = json.loads(record.read_text())
            self.assertEqual(data["units"][0]["state"], "integrated")
            self.assertIsNone(data["publication"])
            self.git(repo, "worktree", "remove", str(worktree))
            result = self.operation(change, record, "check")
            self.assertEqual(result.returncode, 0, result.stderr)


if __name__ == "__main__":
    unittest.main()
