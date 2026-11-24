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

TASKS_MD = textwrap.dedent(
    """\
    # Tasks

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

PROPOSAL = "# demo\n\n- AC1 — one · false today · Observed in: here\n- AC2 — two · false today · Observed in: here\n"


class TasksScriptTest(unittest.TestCase):
    def run_tasks(self, folder, *args):
        return subprocess.run(
            [sys.executable, "-I", "-B", str(TASKS), str(folder), *args],
            text=True,
            capture_output=True,
            timeout=60,
        )

    def folder(self, tmp, tasks=TASKS_MD):
        change = Path(tmp) / "demo"
        change.mkdir()
        (change / "tasks.md").write_text(tasks, encoding="utf-8")
        (change / "proposal.md").write_text(PROPOSAL, encoding="utf-8")
        return change

    def text(self, change):
        return (change / "tasks.md").read_text(encoding="utf-8")

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
            empty = TASKS_MD.replace("2026-11-08 — Plan written.\n", "")
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
            change.chmod(0o500)  # the temporary file next to tasks.md cannot be created
            try:
                result = self.run_tasks(change, "done", "T2", "--commit", "abc1234", "--unverified", "none")
            finally:
                change.chmod(0o700)
            self.assertEqual(result.returncode, 2)
            self.assertEqual(result.stdout, "")
            self.assertEqual(self.text(change), TASKS_MD)

    def test_done_exits_2_for_a_missing_task_or_progress_section(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            result = self.run_tasks(change, "done", "T9", "--commit", "abc1234", "--unverified", "none")
            self.assertEqual(result.returncode, 2)
            self.assertEqual(self.text(change), TASKS_MD)
        with tempfile.TemporaryDirectory() as tmp:
            bare = TASKS_MD.split("## Review status")[0]
            change = self.folder(tmp, tasks=bare)
            result = self.run_tasks(change, "done", "T2", "--commit", "abc1234", "--unverified", "none")
            self.assertEqual(result.returncode, 2)
            self.assertEqual(self.text(change), bare)

    def test_done_rejects_a_missing_unverified_flag(self):
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            result = self.run_tasks(change, "done", "T2", "--commit", "abc1234")
            self.assertEqual(result.returncode, 2)
            self.assertEqual(self.text(change), TASKS_MD)

    def test_status_counts_tasks(self):
        with tempfile.TemporaryDirectory() as tmp:
            result = self.run_tasks(self.folder(tmp), "status")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(
                result.stdout.splitlines(),
                ["T1 [x] 2/2 · First", "T2 [ ] 0/2 · Second thing", "T3 [ ] 0/1 · Third", "T4 [ ] 0/1 · Fourth", "1/4 tasks"],
            )

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
            path = change / "tasks.md"
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
        proposal = textwrap.dedent(
            """\
            # demo

            - **D4 A** — Old way.
            - **D12 A** — New way. **Replaces:** D4-A
            - AC1 — one · false today · Observed in: here
            - AC2 — two · false today · Observed in: here
            """
        )
        with tempfile.TemporaryDirectory() as tmp:
            change = self.folder(tmp)
            (change / "proposal.md").write_text(proposal, encoding="utf-8")
            self.checkpoint(change, "build", "T2", "run T3", "--finding", "H1: keep D4-A")
            self.assertIn("D4-A", self.text(change))
            check = ROOT / "content/skills/flow-plan/scripts/check_change.py"
            result = subprocess.run(
                [sys.executable, "-I", "-B", str(check), str(change)], text=True, capture_output=True
            )
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)


class ExecutionRecordTest(unittest.TestCase):
    """The CLI owns this persisted contract; fixtures use real Git and files."""

    run_tasks = TasksScriptTest.run_tasks
    folder = TasksScriptTest.folder
    text = TasksScriptTest.text
    checkpoint = TasksScriptTest.checkpoint

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

    def setup_execution(self, tmp, role="hive-build-owner", record_root=None):
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
        self.git(repo, "add", "demo")
        self.git(repo, "commit", "-m", "plan")
        base = self.git(repo, "rev-parse", "HEAD")
        worktree = Path(tmp) / "owner"
        self.git(repo, "worktree", "add", "-b", "unit-a", str(worktree), base)
        record = (record_root or repo) / "_support/workspace/run/execution.json"
        record.parent.mkdir(parents=True)
        (record.parent / "check.txt").write_text("observed command exit 0\n")
        unit = dict(id="U1", tasks=["T2"], criteria=["AC2"], depends_on=[],
                    assignment_id="A1", owner_id="owner-1", role=role,
                    repo=str(repo), worktree=str(worktree), branch="unit-a", base_sha=base,
                    allowed_paths=["a.txt"], resources=["port:fixture-a"], state="assigned",
                    returned_head=None, evidence=[], verifier=None, limitations=[],
                    next_action="implement", cleanup=None)
        manifest = dict(grant_ref="proposal.md#decisions",
                        integration_target=dict(repo=str(repo), worktree=str(repo), branch="main"),
                        units=[unit], publication=None)
        source = Path(tmp) / "input.json"
        source.write_text(json.dumps(manifest))
        result = self.run_tasks(change, "execution", "init", "--record", str(record),
                                "--input", str(source), "--coordinator", "root")
        self.assertEqual(result.returncode, 0, result.stderr)
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

    def test_execution_record_lives_in_the_coordinator_checkout(self):
        with tempfile.TemporaryDirectory() as tmp:
            coordinator = self.coordinator_repo(tmp)
            change, record, source, _, worktree = self.setup_execution(
                tmp, role="hive-ticket-owner", record_root=coordinator)
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
            (change / "proposal.md").write_text(PROPOSAL + "changed requirement\n")
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
            (change / "proposal.md").write_text(PROPOSAL.replace("AC2", "AC3"))
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
            (change / "proposal.md").write_text(PROPOSAL + "\nUpdated design context.\n")
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
                proposal = change / "proposal.md"
                proposal.write_text(proposal.read_text() + "\nUpdated during transition.\n")
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
            self.git(repo, "add", "demo/tasks.md")
            self.git(repo, "commit", "-m", "checkpoint")
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


GIT_CONFIG = ["-c", "user.name=t", "-c", "user.email=t@example.invalid", "-c", "init.defaultBranch=main",
              "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null"]


class VersionedTest(unittest.TestCase):
    """`versioned` reads real repositories: a bare origin and its clones, in a temporary folder."""

    def setUp(self):
        self.tmp = Path(tempfile.mkdtemp()).resolve()
        self.addCleanup(shutil.rmtree, self.tmp)

    def git(self, cwd, *args):
        result = subprocess.run(["git", *GIT_CONFIG, *args], cwd=cwd, text=True, capture_output=True, timeout=60)
        self.assertEqual(result.returncode, 0, result.stderr)
        return result.stdout.strip()

    def repo(self, name, hive=""):
        """A clone of a fresh bare origin, with one pushed commit holding AGENTS.md."""
        origin = self.tmp / f"{name}.git"
        self.git(self.tmp, "init", "--bare", str(origin))
        self.git(self.tmp, "clone", str(origin), name)
        repo = self.tmp / name
        (repo / "AGENTS.md").write_text(f"# Guide\n\n## Hive\n\n{hive}", encoding="utf-8")
        self.git(repo, "add", "AGENTS.md")
        self.git(repo, "commit", "-m", "init")
        self.git(repo, "push", "-q", "origin", "main")
        self.git(repo, "remote", "set-head", "origin", "main")
        return repo

    def change(self, specs, name="demo"):
        folder = specs / "changes" / name
        folder.mkdir(parents=True)
        (folder / "tasks.md").write_text(TASKS_MD, encoding="utf-8")
        (folder / "proposal.md").write_text(PROPOSAL, encoding="utf-8")
        return folder

    def commit(self, repo, folder, push=None):
        self.git(repo, "add", str(folder))
        self.git(repo, "commit", "-m", "change folder")
        if push:
            self.git(repo, "push", "-q", "origin", push)

    def versioned(self, folder, *args, cwd=None):
        return subprocess.run([sys.executable, "-I", "-B", str(TASKS), str(folder), "versioned", *args],
                              cwd=cwd or self.tmp, text=True, capture_output=True, timeout=60)

    def assertVersioned(self, result, code, *lines, absent=()):
        self.assertEqual(result.returncode, code, result.stdout + result.stderr)
        for line in lines:
            self.assertIn(line, result.stdout)
        for line in absent:
            self.assertNotIn(line, result.stdout)

    def test_specs_in_the_work_repository_pass_with_later_edits_as_warnings(self):
        repo = self.repo("app", "- Base branch: main\n- Specs: openspec\n")
        folder = self.change(repo / "openspec")
        self.commit(repo, folder, push="main")
        self.assertVersioned(self.versioned(folder, "--work-repo", str(repo)), 0, "versioned: yes", absent=("warning",))
        (folder / "tasks.md").write_text(TASKS_MD + "\nCheckpoint line\n", encoding="utf-8")
        (folder / "design.md").write_text("# Design\n", encoding="utf-8")
        self.assertVersioned(self.versioned(folder), 0, "versioned: yes",
                             "warning: modified openspec/changes/demo/tasks.md: name it in the close report",
                             "warning: untracked openspec/changes/demo/design.md: name it in the close report",
                             absent=("later edits wait for the archive",))

    def test_a_checkpoint_left_in_another_worktree_is_named(self):
        repo = self.repo("app", "- Base branch: main\n- Specs: openspec\n")
        folder = self.change(repo / "openspec")
        self.commit(repo, folder, push="main")
        other = self.tmp / "wt"
        self.git(repo, "worktree", "add", "-q", "--detach", str(other), "HEAD")
        other_tasks = other / "openspec" / "changes" / "demo" / "tasks.md"
        other_tasks.write_text(TASKS_MD + "\n## Checkpoint\n\n- Stage: verify · x\n", encoding="utf-8")
        result = self.versioned(folder)
        self.assertVersioned(result, 0, "versioned: yes",
                             f"warning: modified openspec/changes/demo/tasks.md in {other}: name it in the close report",
                             absent=(f"git -C {other} restore",))

    def test_a_checkpoint_only_edit_is_restored(self):
        repo = self.repo("app", "- Base branch: main\n- Specs: openspec\n")
        folder = self.change(repo / "openspec")
        self.commit(repo, folder, push="main")
        script = [sys.executable, "-I", "-B", str(TASKS), str(folder), "checkpoint", "--stage", "build",
                  "--step", "T2", "--commit", "abc1234", "--next", "go on"]
        self.assertEqual(subprocess.run(script, text=True, capture_output=True, timeout=60).returncode, 0)
        self.assertVersioned(self.versioned(folder), 0, "versioned: yes",
                             "warning: modified openspec/changes/demo/tasks.md: only its checkpoint changed;"
                             f" restore it with git -C {repo} restore -- openspec/changes/demo/tasks.md")

    def test_monorepo_folder_checked_from_a_package_directory(self):
        repo = self.repo("mono", "- Specs: _support/openspec\n")
        folder = self.change(repo / "_support/openspec")
        self.git(repo, "switch", "-c", "feat/work")
        self.commit(repo, folder, push="feat/work")
        package = repo / "packages/api"
        package.mkdir(parents=True)
        result = self.versioned(os.path.relpath(folder, package), cwd=package)
        self.assertVersioned(result, 0, f"repo: {repo} ·", "base: main (origin/HEAD)", "versioned: yes")

    def test_sibling_direct_base_specs_from_a_workspace_root_outside_git(self):
        specs = self.repo("specs", "- Base branch: main\n- Delivery: direct-base\n")
        code = self.repo("code", "- Specs: ../specs/openspec · https://example.invalid/specs.git\n")
        folder = self.change(specs / "openspec")
        args = ("--work-repo", "code")
        self.assertVersioned(self.versioned("specs/openspec/changes/demo", *args), 1, "versioned: no",
                             "untracked: openspec/changes/demo/tasks.md", "not committed: the folder has no commit",
                             "next: commit and deliver the folder as git-workflow step 4 sets", absent=("warning",))
        self.git(specs, "switch", "-c", "feat/work")
        self.commit(specs, folder, push="feat/work")
        self.assertVersioned(self.versioned("specs/openspec/changes/demo", *args), 1, "delivery: direct-base",
                             "is not on origin/main")
        self.git(specs, "switch", "main")
        self.git(specs, "merge", "-q", "--ff-only", "feat/work")
        self.git(specs, "push", "-q", "origin", "main")
        self.assertVersioned(self.versioned("specs/openspec/changes/demo", *args), 0, "versioned: yes",
                             f"work repo: {code}", absent=("warning",))
        (folder / "tasks.md").write_text(TASKS_MD + "\nedit\n", encoding="utf-8")
        self.assertVersioned(self.versioned("specs/openspec/changes/demo", *args), 1,
                             "modified: openspec/changes/demo/tasks.md")

    def test_squash_merged_folder_with_its_branch_deleted_passes(self):
        for delivery in ("", "- Delivery: direct-base\n"):
            with self.subTest(delivery=delivery):
                repo = self.repo(f"app{len(delivery)}", f"- Base branch: `main`\n{delivery}- Specs: `openspec`\n")
                folder = self.change(repo / "openspec")
                self.git(repo, "switch", "-c", "feat/work")
                self.commit(repo, folder, push="feat/work")
                self.git(repo, "switch", "main")
                self.git(repo, "merge", "-q", "--squash", "feat/work")
                self.git(repo, "commit", "-m", "squash")
                self.git(repo, "push", "-q", "origin", "main", ":feat/work")
                self.git(repo, "branch", "-D", "feat/work")
                self.git(repo, "fetch", "-q", "--prune", "origin")
                result = self.versioned(folder, "--work-repo", str(repo))
                self.assertVersioned(result, 0, "versioned: yes", "base: main (Hive)", absent=("warning",))

    def test_folder_without_tasks_md_and_absent_work_repo(self):
        repo = self.repo("app", "- Specs: openspec\n")
        folder = self.change(repo / "openspec")
        (folder / "tasks.md").unlink()
        self.commit(repo, folder, push="main")
        self.assertVersioned(self.versioned(folder), 0, "versioned: yes",
                             "warning: no --work-repo given; the folder's repository is taken as the work's")

    def test_commit_on_no_remote_branch_fails(self):
        repo = self.repo("app")
        folder = self.change(repo / "openspec")
        self.commit(repo, folder)
        self.assertVersioned(self.versioned(folder), 1, "is on no remote branch")

    def test_direct_base_without_a_resolvable_base_fails(self):
        repo = self.tmp / "bare-hive"
        repo.mkdir()
        self.git(repo, "init")
        (repo / "AGENTS.md").write_text("## Hive\n\n- Delivery: direct-base\n", encoding="utf-8")
        folder = self.change(repo / "openspec")
        self.git(repo, "add", "AGENTS.md")
        self.commit(repo, folder)
        self.assertVersioned(self.versioned(folder), 1, "no base branch resolves")

    def test_folder_outside_git_cannot_be_versioned(self):
        folder = self.change(self.tmp / "loose")
        self.assertVersioned(self.versioned(folder), 1, "is in no Git repository")

    def test_work_repo_specs_warnings(self):
        specs = self.repo("specs", "- Base branch: main\n")
        folder = self.change(specs / "openspec")
        self.commit(specs, folder, push="main")
        self.repo("bare-code")
        self.repo("elsewhere", "- Specs: docs/openspec\n")
        self.assertVersioned(self.versioned(folder, "--work-repo", "bare-code"), 0,
                             "declares no Specs; declare Specs: ../specs/openspec")
        self.assertVersioned(self.versioned(folder, "--work-repo", "elsewhere"), 0,
                             "Specs docs/openspec does not hold the folder")
        (self.tmp / "plain").mkdir()
        self.assertEqual(self.versioned(folder, "--work-repo", "plain").returncode, 2)

    def test_worktree_and_main_checkout_count_as_one_repository(self):
        repo = self.repo("app", "- Specs: openspec\n")
        self.git(repo, "worktree", "add", "-q", "-b", "feat/work", str(self.tmp / "wt"))
        folder = self.change(self.tmp / "wt/openspec")
        self.commit(self.tmp / "wt", folder, push="feat/work")
        (folder / "tasks.md").write_text(TASKS_MD + "\nedit\n", encoding="utf-8")
        self.assertVersioned(self.versioned(folder, "--work-repo", str(repo)), 0, "work repo: same",
                             "warning: modified openspec/changes/demo/tasks.md")


if __name__ == "__main__":
    unittest.main()


DONE_MD = TASKS_MD.replace("- [ ]", "- [x]")


class IntegratedFolderTest(unittest.TestCase):
    """After its merge, with every task done, nothing is written to the change folder."""

    setUp, git, repo, change, commit = (VersionedTest.setUp, VersionedTest.git, VersionedTest.repo,
                                        VersionedTest.change, VersionedTest.commit)

    def integrated(self, tasks=DONE_MD, hive="- Base branch: main\n- Specs: openspec\n"):
        repo = self.repo("app", hive)
        folder = self.change(repo / "openspec")
        (folder / "tasks.md").write_text(tasks, encoding="utf-8")
        self.commit(repo, folder, push="main")
        return repo, folder

    def run_tasks(self, folder, *args, cwd=None):
        return subprocess.run([sys.executable, "-I", "-B", str(TASKS), str(folder), *args],
                              cwd=cwd or self.tmp, text=True, capture_output=True, timeout=60)

    def checkpoint(self, folder, *extra):
        return self.run_tasks(folder, "checkpoint", "--stage", "close", "--step", "integrated",
                              "--commit", "abc1234", "--next", "close", *extra)

    def assertRefused(self, result, folder, before, out=True):
        self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
        text = result.stdout + result.stderr
        self.assertIn("is on origin/main with every task done", text)
        self.assertEqual("--out" in text, out, text)
        self.assertEqual((folder / "tasks.md").read_bytes(), before)

    def test_checkpoint_is_refused_on_the_base(self):
        repo, folder = self.integrated()
        before = (folder / "tasks.md").read_bytes()
        self.assertRefused(self.checkpoint(folder), folder, before)

    def test_checkpoint_is_refused_on_no_branch(self):
        repo, folder = self.integrated()
        self.git(repo, "switch", "-q", "--detach", "HEAD")
        before = (folder / "tasks.md").read_bytes()
        self.assertRefused(self.checkpoint(folder), folder, before)

    def test_checkpoint_is_refused_on_the_merged_branch(self):
        repo = self.repo("app", "- Base branch: main\n- Specs: openspec\n")
        self.git(repo, "switch", "-q", "-c", "feat/x")
        folder = self.change(repo / "openspec")
        (folder / "tasks.md").write_text(DONE_MD, encoding="utf-8")
        self.commit(repo, folder, push="feat/x")
        self.git(repo, "push", "-q", "origin", "feat/x:main")
        self.git(repo, "fetch", "-q", "origin")
        before = (folder / "tasks.md").read_bytes()
        self.assertRefused(self.checkpoint(folder), folder, before)

    def test_checkpoint_is_refused_after_a_squash_merge(self):
        repo = self.repo("app", "- Base branch: main\n- Specs: openspec\n")
        self.git(repo, "switch", "-q", "-c", "feat/x")
        folder = self.change(repo / "openspec")
        (folder / "tasks.md").write_text(DONE_MD, encoding="utf-8")
        self.commit(repo, folder, push="feat/x")
        self.git(repo, "switch", "-q", "main")
        self.git(repo, "merge", "-q", "--squash", "feat/x")
        self.git(repo, "commit", "-q", "-m", "squash")
        self.git(repo, "push", "-q", "origin", "main")
        self.git(repo, "switch", "-q", "feat/x")
        before = (folder / "tasks.md").read_bytes()
        self.assertRefused(self.checkpoint(folder), folder, before)

    def test_done_is_refused(self):
        repo, folder = self.integrated()
        before = (folder / "tasks.md").read_bytes()
        result = self.run_tasks(folder, "done", "T1", "--commit", "abc1234", "--unverified", "none")
        self.assertRefused(result, folder, before, out=False)

    def test_an_open_task_keeps_writing(self):
        repo, folder = self.integrated(tasks=TASKS_MD)
        self.assertEqual(self.checkpoint(folder).returncode, 0)
        self.assertIn("## Checkpoint", (folder / "tasks.md").read_text(encoding="utf-8"))

    def test_a_later_unit_records_its_last_task_then_its_checkpoint(self):
        last_open = DONE_MD.replace("- [x] gamma", "- [ ] gamma")
        repo, folder = self.integrated(tasks=last_open)
        self.git(repo, "switch", "-q", "-c", "unit-2")
        done = self.run_tasks(folder, "done", "T3", "--commit", "abc1234", "--unverified", "none")
        self.assertEqual(done.returncode, 0, done.stdout + done.stderr)
        self.assertEqual(self.checkpoint(folder).returncode, 0)

    def test_a_folder_only_on_its_branch_keeps_writing(self):
        repo = self.repo("app", "- Base branch: main\n- Specs: openspec\n")
        self.git(repo, "switch", "-q", "-c", "feat/x")
        folder = self.change(repo / "openspec")
        (folder / "tasks.md").write_text(DONE_MD, encoding="utf-8")
        self.commit(repo, folder, push="feat/x")
        self.assertEqual(self.checkpoint(folder).returncode, 0)

    def test_a_repository_without_commits_or_remote_keeps_writing(self):
        bare = self.tmp / "bare"
        self.git(self.tmp, "init", "-q", str(bare))
        folder = bare / "openspec" / "changes" / "demo"
        folder.mkdir(parents=True)
        (folder / "tasks.md").write_text(DONE_MD, encoding="utf-8")
        self.assertEqual(self.checkpoint(folder).returncode, 0)
        self.git(bare, "add", ".")
        self.git(bare, "commit", "-q", "-m", "local")
        self.assertEqual(self.checkpoint(folder).returncode, 0)

    def test_out_writes_only_the_checkpoint_elsewhere(self):
        repo, folder = self.integrated()
        before = (folder / "tasks.md").read_bytes()
        out = self.tmp / "outside" / "demo-checkpoint.md"
        out.parent.mkdir()
        first = self.checkpoint(folder, "--out", str(out), "--check", "suite · pass · abc1234")
        self.assertEqual(first.returncode, 0, first.stdout + first.stderr)
        self.assertIn(f"checkpoint written to {out}", first.stdout)
        second = self.checkpoint(folder, "--out", str(out))
        self.assertEqual(second.returncode, 0, second.stdout + second.stderr)
        text = out.read_text(encoding="utf-8")
        self.assertTrue(text.startswith("## Checkpoint"), text)
        self.assertIn("- suite · pass · abc1234", text)
        self.assertNotIn("### T1", text)
        self.assertEqual((folder / "tasks.md").read_bytes(), before)

    def test_a_task_reopened_in_the_working_copy_keeps_writing(self):
        repo, folder = self.integrated()
        self.git(repo, "switch", "-q", "-c", "reopen")
        text = (folder / "tasks.md").read_text(encoding="utf-8").replace("- [x] delta", "- [ ] delta")
        (folder / "tasks.md").write_text(text, encoding="utf-8")
        done = self.run_tasks(folder, "done", "T4", "--commit", "abc1234", "--unverified", "none")
        self.assertEqual(done.returncode, 0, done.stdout + done.stderr)
        again = self.checkpoint(folder)
        self.assertEqual(again.returncode, 0, again.stdout + again.stderr)

    def test_out_starts_an_empty_file(self):
        repo, folder = self.integrated()
        out = self.tmp / "empty.md"
        out.write_text("", encoding="utf-8")
        result = self.checkpoint(folder, "--out", str(out))
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertTrue(out.read_text(encoding="utf-8").startswith("## Checkpoint"))

    def test_out_does_not_replace_a_file_that_is_not_a_checkpoint(self):
        repo, folder = self.integrated()
        notes = self.tmp / "notes.md"
        notes.write_text("my notes\nimportant\n", encoding="utf-8")
        result = self.checkpoint(folder, "--out", str(notes))
        self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
        self.assertIn("not a checkpoint file", result.stdout + result.stderr)
        self.assertEqual(notes.read_text(encoding="utf-8"), "my notes\nimportant\n")

    def test_out_inside_the_repository_is_refused(self):
        repo, folder = self.integrated()
        result = self.checkpoint(folder, "--out", str(repo / "checkpoint.md"))
        self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
        self.assertIn("outside the repository", result.stdout + result.stderr)
        self.assertFalse((repo / "checkpoint.md").exists())
