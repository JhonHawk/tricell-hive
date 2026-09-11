"""Behavioral tests for the portable Hive plan validator."""

from __future__ import annotations

import hashlib
import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


SCRIPT = Path(__file__).with_name("plan.py")
SPEC = importlib.util.spec_from_file_location("hive_plan", SCRIPT)
assert SPEC is not None and SPEC.loader is not None
PLAN = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = PLAN
SPEC.loader.exec_module(PLAN)


def render_plan(
    root: Path,
    *,
    status: str = "planned",
    grants: list[dict] | None = None,
    revocations: list[dict] | None = None,
    execution: str = "| T1 | pending | |",
    contract: str = "\n# Fixture plan\n\n## Contract\n\nGoal: exercise the validator.\n",
    recovery: dict | None = None,
    include_recovery: bool = False,
) -> Path:
    contract_bytes = contract.encode("utf-8")
    authorization = {
        "schema": PLAN.SCHEMA_AUTHORIZATION,
        "plan_id": "fixture-plan",
        "contract_sha256": hashlib.sha256(contract_bytes).hexdigest(),
        "approval": {
            "evidence": "The user explicitly authorized this fixture.",
            "date": "2026-09-10",
            "revision": 1,
        },
        "grants": grants
        if grants is not None
        else [
            {
                "id": "grant-implement",
                "action": "implement",
                "targets": ["repo:fixture"],
                "conditions": [],
                "evidence": "The user explicitly authorized implementation.",
            }
        ],
        "revocations": revocations if revocations is not None else [],
    }
    document = (
        f"# Fixture\n\nStatus: {status}\n\n"
        "<!-- hive-plan:contract:start -->"
    ).encode("utf-8")
    document += contract_bytes
    document += (
        "<!-- hive-plan:contract:end -->\n\n## Execution\n\n"
        "| Task | State | Evidence |\n|---|---|---|\n"
    ).encode("utf-8")
    document += execution.encode("utf-8")
    if include_recovery:
        document += (
            "\n\n<!-- hive-plan:recovery:start -->\n```json\n"
            + json.dumps(recovery, indent=2, sort_keys=True)
            + "\n```\n<!-- hive-plan:recovery:end -->"
        ).encode("utf-8")
    document += (
        "\n\n<!-- hive-plan:authorization:start -->\n```json\n"
        + json.dumps(authorization, indent=2, sort_keys=True)
        + "\n```\n<!-- hive-plan:authorization:end -->\n"
    ).encode("utf-8")
    path = root / "plan.md"
    path.write_bytes(document)
    return path


class PlanValidatorTests(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory(prefix="hive-plan-")
        self.addCleanup(self.scratch.cleanup)
        self.root = Path(self.scratch.name)
        self.plan = render_plan(self.root)

    def test_exact_grant_validates_and_cli_emits_json(self):
        result = PLAN.validate_plan(self.plan, "implement", "repo:fixture")
        self.assertEqual(result["grant_ids"], ["grant-implement"])

        cli = subprocess.run(
            [
                sys.executable,
                str(SCRIPT),
                "validate",
                str(self.plan),
                "--action",
                "implement",
                "--target",
                "repo:fixture",
            ],
            check=False,
            capture_output=True,
            text=True,
        )
        self.assertEqual(cli.returncode, 0, cli.stderr)
        self.assertTrue(json.loads(cli.stdout)["valid"])

    def test_malformed_document_is_rejected(self):
        self.plan.write_text("<!-- hive-plan:contract:start -->\n", encoding="utf-8")
        with self.assertRaises(PLAN.PlanError) as raised:
            PLAN.inspect_plan(self.plan)
        self.assertEqual(raised.exception.code, "malformed_document")

    def test_draft_is_inspectable_but_not_executable(self):
        draft = render_plan(self.root, status="draft")
        inspected = PLAN.inspect_plan(draft)
        self.assertEqual(inspected["status"], "draft")
        self.assertFalse(inspected["can_implement"])
        with self.assertRaises(PLAN.PlanError) as raised:
            PLAN.validate_plan(draft, "implement", "repo:fixture")
        self.assertEqual(raised.exception.code, "draft_not_executable")

    def test_progress_table_mutation_does_not_change_contract_digest(self):
        original = PLAN.digest_plan(self.plan)
        self.plan.write_text(
            self.plan.read_text(encoding="utf-8").replace(
                "| T1 | pending | |", "| T1 | done | tests: 1 passing |"
            ),
            encoding="utf-8",
        )
        self.assertEqual(PLAN.digest_plan(self.plan), original)
        self.assertTrue(PLAN.validate_plan(self.plan, "implement", "repo:fixture")["valid"])

    def test_contract_mutation_unbinds_authorization(self):
        original = PLAN.digest_plan(self.plan)
        self.plan.write_text(
            self.plan.read_text(encoding="utf-8").replace(
                "Goal: exercise the validator.", "Goal: changed after approval."
            ),
            encoding="utf-8",
        )
        self.assertNotEqual(PLAN.digest_plan(self.plan), original)
        inspected = PLAN.inspect_plan(self.plan)
        self.assertEqual(inspected["authorization"]["status"], "unbound")
        with self.assertRaises(PLAN.PlanError) as raised:
            PLAN.validate_plan(self.plan, "implement", "repo:fixture")
        self.assertEqual(raised.exception.code, "contract_drift")

    def test_validate_rejects_recovery_history_bound_to_another_plan(self):
        plan = render_plan(
            self.root,
            recovery={
                "schema": PLAN.SCHEMA_RECOVERY,
                "plan_id": "another-plan",
                "attempts": [],
            },
            include_recovery=True,
        )
        with self.assertRaises(PLAN.PlanError) as raised:
            PLAN.validate_plan(plan, "implement", "repo:fixture")
        self.assertEqual(raised.exception.code, "recovery_plan_mismatch")

        cli = subprocess.run(
            [
                sys.executable,
                str(SCRIPT),
                "validate",
                str(plan),
                "--action",
                "implement",
                "--target",
                "repo:fixture",
            ],
            check=False,
            capture_output=True,
            text=True,
        )
        self.assertNotEqual(cli.returncode, 0)
        self.assertEqual(json.loads(cli.stdout)["error"]["code"], "recovery_plan_mismatch")

    def test_undeclared_operation_and_destination_are_distinct(self):
        with self.assertRaises(PLAN.PlanError) as operation:
            PLAN.validate_plan(self.plan, "push", "repo:fixture")
        self.assertEqual(operation.exception.code, "undeclared_operation")

        with self.assertRaises(PLAN.PlanError) as destination:
            PLAN.validate_plan(self.plan, "implement", "repo:other")
        self.assertEqual(destination.exception.code, "undeclared_destination")

    def test_revocation_is_effective_for_current_grant(self):
        revoked = render_plan(
            self.root,
            revocations=[
                {
                    "grant_id": "grant-implement",
                    "reason": "approval withdrawn",
                    "evidence": "The user explicitly withdrew the grant.",
                    "revoked_at": "2026-09-10T18:00:00-06:00",
                }
            ],
        )
        with self.assertRaises(PLAN.PlanError) as raised:
            PLAN.validate_plan(revoked, "implement", "repo:fixture")
        self.assertEqual(raised.exception.code, "grant_revoked")

    def test_unresolved_conditions_fail_closed(self):
        conditional = render_plan(
            self.root,
            grants=[
                {
                    "id": "grant-implement",
                    "action": "implement",
                    "targets": ["repo:fixture"],
                    "conditions": [
                        {"requirement": "affected tests are green", "evidence": ""}
                    ],
                    "evidence": "The user explicitly authorized implementation.",
                }
            ],
        )
        with self.assertRaises(PLAN.PlanError) as raised:
            PLAN.validate_plan(conditional, "implement", "repo:fixture")
        self.assertEqual(raised.exception.code, "conditions_unresolved")

    def test_satisfied_condition_uses_recorded_evidence_without_new_approval(self):
        conditional = render_plan(
            self.root,
            grants=[
                {
                    "id": "grant-implement",
                    "action": "implement",
                    "targets": ["repo:fixture"],
                    "conditions": [
                        {
                            "requirement": "affected tests are green",
                            "evidence": "pytest: 4 passed",
                        }
                    ],
                    "evidence": "The user explicitly authorized implementation.",
                }
            ],
        )
        result = PLAN.validate_plan(conditional, "implement", "repo:fixture")
        self.assertEqual(result["conditions"][0]["evidence"], "pytest: 4 passed")

    def test_unrelated_pending_delivery_condition_does_not_block_implementation(self):
        grants = [
            {
                "id": "grant-implement",
                "action": "implement",
                "targets": ["repo:fixture"],
                "conditions": [],
                "evidence": "The user explicitly authorized implementation.",
            },
            {
                "id": "grant-push",
                "action": "push",
                "targets": ["origin:feature"],
                "conditions": [
                    {"requirement": "verification is complete", "evidence": ""}
                ],
                "evidence": "The user explicitly authorized publication.",
            },
        ]
        plan = render_plan(self.root, grants=grants)
        self.assertTrue(PLAN.inspect_plan(plan)["can_implement"])
        self.assertTrue(PLAN.validate_plan(plan, "implement", "repo:fixture")["valid"])
        with self.assertRaises(PLAN.PlanError) as raised:
            PLAN.validate_plan(plan, "push", "origin:feature")
        self.assertEqual(raised.exception.code, "invalid_action_state")

    def test_satisfied_grant_alternative_ignores_pending_same_target_grant(self):
        grants = [
            {
                "id": "grant-implement-open",
                "action": "implement",
                "targets": ["repo:fixture"],
                "conditions": [],
                "evidence": "The user explicitly authorized implementation.",
            },
            {
                "id": "grant-implement-conditional",
                "action": "implement",
                "targets": ["repo:fixture"],
                "conditions": [
                    {"requirement": "review is complete", "evidence": ""}
                ],
                "evidence": "The user also preauthorized the reviewed path.",
            },
        ]
        plan = render_plan(self.root, grants=grants)
        result = PLAN.validate_plan(plan, "implement", "repo:fixture")
        self.assertEqual(result["grant_ids"], ["grant-implement-open"])
        self.assertEqual(result["conditions"], [])

    def test_design_only_approval_has_no_implementation_authority(self):
        design_only = render_plan(self.root, grants=[])
        inspected = PLAN.inspect_plan(design_only)
        self.assertTrue(inspected["ok"])
        self.assertFalse(inspected["can_implement"])
        with self.assertRaises(PLAN.PlanError) as raised:
            PLAN.validate_plan(design_only, "implement", "repo:fixture")
        self.assertEqual(raised.exception.code, "undeclared_operation")

    def test_delivery_waits_for_verified_status(self):
        push_grant = [
            {
                "id": "grant-push",
                "action": "push",
                "targets": ["origin:feature"],
                "conditions": [],
                "evidence": "The user explicitly authorized publication.",
            }
        ]
        planned = render_plan(self.root, grants=push_grant)
        with self.assertRaises(PLAN.PlanError) as raised:
            PLAN.validate_plan(planned, "push", "origin:feature")
        self.assertEqual(raised.exception.code, "invalid_action_state")

        verified = render_plan(self.root, status="verified", grants=push_grant)
        self.assertTrue(PLAN.validate_plan(verified, "push", "origin:feature")["valid"])

    def test_verified_plan_does_not_reopen_implementation(self):
        verified = render_plan(self.root, status="verified")
        with self.assertRaises(PLAN.PlanError) as raised:
            PLAN.validate_plan(verified, "implement", "repo:fixture")
        self.assertEqual(raised.exception.code, "invalid_action_state")

    def test_duplicate_json_keys_are_malformed(self):
        text = self.plan.read_text(encoding="utf-8")
        text = text.replace('"plan_id": "fixture-plan",', '"plan_id": "fixture-plan",\n  "plan_id": "other",')
        self.plan.write_text(text, encoding="utf-8")
        with self.assertRaises(PLAN.PlanError) as raised:
            PLAN.inspect_plan(self.plan)
        self.assertEqual(raised.exception.code, "duplicate_json_key")

    def test_legacy_plan_reports_unknown_recovery_history(self):
        inspected = PLAN.inspect_plan(self.plan)
        self.assertEqual(inspected["recovery"]["status"], "unknown")
        self.assertFalse(inspected["recovery"]["history_known"])
        self.assertEqual(inspected["recovery"]["attempts"], 0)

        result = PLAN.recovery_check(self.plan, "delegation", "T1")
        self.assertFalse(result["allowed"])
        self.assertEqual(result["decision"], "reconcile")
        self.assertEqual(result["status"], "unknown")

    def test_empty_recovery_history_is_known_and_does_not_change_digest(self):
        recovery = {
            "schema": PLAN.SCHEMA_RECOVERY,
            "plan_id": "fixture-plan",
            "attempts": [],
        }
        plan = render_plan(
            self.root,
            recovery=recovery,
            include_recovery=True,
        )
        original = PLAN.digest_plan(plan)
        inspected = PLAN.inspect_plan(plan)
        self.assertEqual(inspected["recovery"]["status"], "known")
        self.assertTrue(inspected["recovery"]["history_known"])
        self.assertEqual(inspected["recovery"]["attempts"], 0)
        self.assertEqual(PLAN.digest_plan(plan), original)

        result = PLAN.recovery_check(plan, "delegation", "T1")
        self.assertTrue(result["allowed"])
        self.assertEqual(result["decision"], "continue")

    def test_recovery_counts_delegation_rerun_and_preserves_independent_scope(self):
        recovery = {
            "schema": PLAN.SCHEMA_RECOVERY,
            "plan_id": "fixture-plan",
            "attempts": [
                {
                    "id": "attempt-1",
                    "sequence": 1,
                    "kind": "delegation",
                    "scope": "T1",
                    "contract_sha256": PLAN.digest_plan(self.plan),
                    "started_at": "2026-09-10T18:00:00-06:00",
                    "ended_at": "2026-09-10T18:01:00-06:00",
                    "outcome": "failed",
                    "evidence": "child returned no usable result",
                },
                {
                    "id": "attempt-2",
                    "sequence": 2,
                    "kind": "delegation",
                    "scope": "T2",
                    "contract_sha256": "0" * 64,
                    "started_at": "2026-09-10T18:02:00-06:00",
                    "ended_at": "2026-09-10T18:03:00-06:00",
                    "outcome": "failed",
                    "evidence": "independent task failed",
                },
            ],
        }
        plan = render_plan(self.root, recovery=recovery, include_recovery=True)
        self.assertTrue(PLAN.recovery_check(plan, "delegation", "T1")["allowed"])
        self.assertEqual(PLAN.recovery_check(plan, "delegation", "T1")["attempts"]["charged"], 1)
        self.assertTrue(PLAN.recovery_check(plan, "delegation", "T2")["allowed"])

    def test_recovery_exhausts_delegation_after_one_rerun(self):
        digest = PLAN.digest_plan(self.plan)
        attempts = [
            {
                "id": f"attempt-{index}",
                "sequence": index,
                "kind": "delegation",
                "scope": "T1",
                "contract_sha256": digest,
                "started_at": f"2026-09-10T18:0{index}:00-06:00",
                "ended_at": f"2026-09-10T18:0{index}:30-06:00",
                "outcome": "failed",
                "evidence": f"delegation failure {index}",
            }
            for index in (1, 2)
        ]
        plan = render_plan(
            self.root,
            recovery={
                "schema": PLAN.SCHEMA_RECOVERY,
                "plan_id": "fixture-plan",
                "attempts": attempts,
            },
            include_recovery=True,
        )
        result = PLAN.recovery_check(plan, "delegation", "T1")
        self.assertFalse(result["allowed"])
        self.assertEqual(result["decision"], "exhausted")
        self.assertEqual(result["limit"], 2)

    def test_open_recovery_attempt_requires_reconciliation_and_cli_returns_nonzero(self):
        plan = render_plan(
            self.root,
            recovery={
                "schema": PLAN.SCHEMA_RECOVERY,
                "plan_id": "fixture-plan",
                "attempts": [
                    {
                        "id": "attempt-open",
                        "sequence": 1,
                        "kind": "remote",
                        "scope": "push",
                        "contract_sha256": PLAN.digest_plan(self.plan),
                        "started_at": "2026-09-10T18:00:00-06:00",
                        "outcome": "started",
                        "evidence": "push process was launched",
                    }
                ],
            },
            include_recovery=True,
        )
        before = plan.read_bytes()
        result = PLAN.recovery_check(plan, "remote", "push")
        self.assertFalse(result["allowed"])
        self.assertEqual(result["decision"], "reconcile")
        self.assertEqual(result["attempts"]["open"], 1)
        self.assertEqual(plan.read_bytes(), before)

        self.assertTrue(PLAN.validate_plan(plan, "implement", "repo:fixture")["valid"])

        cli = subprocess.run(
            [
                sys.executable,
                str(SCRIPT),
                "recovery-check",
                str(plan),
                "--kind",
                "remote",
                "--scope",
                "push",
            ],
            check=False,
            capture_output=True,
            text=True,
        )
        self.assertNotEqual(cli.returncode, 0)
        self.assertEqual(json.loads(cli.stdout)["decision"], "reconcile")

    def test_remote_verified_transient_failure_is_exempt_once(self):
        digest = PLAN.digest_plan(self.plan)
        attempts = [
            {
                "id": "remote-transient",
                "sequence": 1,
                "kind": "remote",
                "scope": "deploy",
                "contract_sha256": digest,
                "started_at": "2026-09-10T18:00:00-06:00",
                "ended_at": "2026-09-10T18:01:00-06:00",
                "outcome": "failed",
                "evidence": "transport timed out before response",
                "exception": {
                    "type": "verified_transient",
                    "evidence": "provider status confirmed a transient timeout",
                },
            },
            {
                "id": "remote-failed",
                "sequence": 2,
                "kind": "remote",
                "scope": "deploy",
                "contract_sha256": "0" * 64,
                "started_at": "2026-09-10T18:02:00-06:00",
                "ended_at": "2026-09-10T18:03:00-06:00",
                "outcome": "failed",
                "evidence": "deployment rejected by provider",
            },
        ]
        plan = render_plan(
            self.root,
            recovery={
                "schema": PLAN.SCHEMA_RECOVERY,
                "plan_id": "fixture-plan",
                "attempts": attempts,
            },
            include_recovery=True,
        )
        result = PLAN.recovery_check(plan, "remote", "deploy")
        self.assertTrue(result["allowed"])
        self.assertEqual(result["attempts"]["charged"], 1)
        self.assertEqual(result["attempts"]["transient_exemptions"], 1)

    def test_fix_and_review_budgets_count_failed_corrections_by_scope(self):
        digest = PLAN.digest_plan(self.plan)
        attempts = []
        sequence = 0
        for kind, scope, count in (("fix", "bug-1", 3), ("review", "cycle-1", 2)):
            for _ in range(count):
                sequence += 1
                attempts.append(
                    {
                        "id": f"attempt-{sequence}",
                        "sequence": sequence,
                        "kind": kind,
                        "scope": scope,
                        "contract_sha256": digest,
                        "started_at": f"2026-09-10T18:{sequence:02d}:00-06:00",
                        "ended_at": f"2026-09-10T18:{sequence:02d}:30-06:00",
                        "outcome": "failed",
                        "evidence": f"{kind} correction failed",
                    }
                )
        plan = render_plan(
            self.root,
            recovery={
                "schema": PLAN.SCHEMA_RECOVERY,
                "plan_id": "fixture-plan",
                "attempts": attempts,
            },
            include_recovery=True,
        )
        self.assertEqual(PLAN.recovery_check(plan, "fix", "bug-1")["decision"], "exhausted")
        self.assertEqual(PLAN.recovery_check(plan, "review", "cycle-1")["decision"], "exhausted")
        self.assertTrue(PLAN.recovery_check(plan, "fix", "other-bug")["allowed"])

    def test_successful_review_completes_scope_across_digest_revisions(self):
        plan = render_plan(
            self.root,
            recovery={
                "schema": PLAN.SCHEMA_RECOVERY,
                "plan_id": "fixture-plan",
                "attempts": [
                    {
                        "id": "review-success",
                        "sequence": 1,
                        "kind": "review",
                        "scope": "cycle-1",
                        "contract_sha256": "0" * 64,
                        "started_at": "2026-09-10T18:00:00-06:00",
                        "ended_at": "2026-09-10T18:01:00-06:00",
                        "outcome": "succeeded",
                        "evidence": "review approved the current correction",
                    }
                ],
            },
            include_recovery=True,
        )
        result = PLAN.recovery_check(plan, "review", "cycle-1")
        self.assertFalse(result["allowed"])
        self.assertEqual(result["decision"], "completed")
        self.assertEqual(result["attempts"]["charged"], 1)
        self.assertTrue(PLAN.recovery_check(plan, "review", "new-cycle")["allowed"])

        cli = subprocess.run(
            [
                sys.executable,
                str(SCRIPT),
                "recovery-check",
                str(plan),
                "--kind",
                "review",
                "--scope",
                "cycle-1",
            ],
            check=False,
            capture_output=True,
            text=True,
        )
        self.assertNotEqual(cli.returncode, 0)
        self.assertEqual(json.loads(cli.stdout)["decision"], "completed")

    def test_successful_remote_does_not_close_milestone_scope(self):
        digest = PLAN.digest_plan(self.plan)
        attempts = [
            {
                "id": "remote-success",
                "sequence": 1,
                "kind": "remote",
                "scope": "milestone-1",
                "contract_sha256": "0" * 64,
                "started_at": "2026-09-10T18:00:00-06:00",
                "ended_at": "2026-09-10T18:01:00-06:00",
                "outcome": "succeeded",
                "evidence": "provider accepted the first operation",
            },
            {
                "id": "remote-failure-1",
                "sequence": 2,
                "kind": "remote",
                "scope": "milestone-1",
                "contract_sha256": digest,
                "started_at": "2026-09-10T18:02:00-06:00",
                "ended_at": "2026-09-10T18:03:00-06:00",
                "outcome": "failed",
                "evidence": "provider rejected the second operation",
            },
        ]
        plan = render_plan(
            self.root,
            recovery={
                "schema": PLAN.SCHEMA_RECOVERY,
                "plan_id": "fixture-plan",
                "attempts": attempts,
            },
            include_recovery=True,
        )
        result = PLAN.recovery_check(plan, "remote", "milestone-1")
        self.assertTrue(result["allowed"])
        self.assertEqual(result["decision"], "continue")
        self.assertEqual(result["attempts"]["charged"], 1)
        self.assertTrue(PLAN.recovery_check(plan, "remote", "new-milestone")["allowed"])

        attempts.append(
            {
                "id": "remote-failure-2",
                "sequence": 3,
                "kind": "remote",
                "scope": "milestone-1",
                "contract_sha256": digest,
                "started_at": "2026-09-10T18:04:00-06:00",
                "ended_at": "2026-09-10T18:05:00-06:00",
                "outcome": "failed",
                "evidence": "provider rejected the third operation",
            }
        )
        render_plan(
            self.root,
            recovery={
                "schema": PLAN.SCHEMA_RECOVERY,
                "plan_id": "fixture-plan",
                "attempts": attempts,
            },
            include_recovery=True,
        )
        exhausted = PLAN.recovery_check(plan, "remote", "milestone-1")
        self.assertFalse(exhausted["allowed"])
        self.assertEqual(exhausted["decision"], "exhausted")
        self.assertEqual(exhausted["attempts"]["charged"], 2)

    def test_recovery_structure_rejects_duplicate_sequence_and_unknown_predecessor(self):
        digest = PLAN.digest_plan(self.plan)
        base = {
            "schema": PLAN.SCHEMA_RECOVERY,
            "plan_id": "fixture-plan",
            "attempts": [
                {
                    "id": "attempt-1",
                    "sequence": 1,
                    "kind": "fix",
                    "scope": "bug-1",
                    "contract_sha256": digest,
                    "started_at": "2026-09-10T18:00:00-06:00",
                    "ended_at": "2026-09-10T18:01:00-06:00",
                    "outcome": "failed",
                    "evidence": "test still fails",
                },
                {
                    "id": "attempt-2",
                    "sequence": 1,
                    "kind": "fix",
                    "scope": "bug-1",
                    "contract_sha256": digest,
                    "started_at": "2026-09-10T18:02:00-06:00",
                    "ended_at": "2026-09-10T18:03:00-06:00",
                    "outcome": "failed",
                    "evidence": "test still fails",
                    "predecessor": "missing-attempt",
                },
            ],
        }
        plan = render_plan(self.root, recovery=base, include_recovery=True)
        with self.assertRaises(PLAN.PlanError) as raised:
            PLAN.inspect_plan(plan)
        self.assertEqual(raised.exception.code, "invalid_recovery_sequence")

    def test_recovery_structure_rejects_duplicate_ids(self):
        digest = PLAN.digest_plan(self.plan)
        attempt = {
            "id": "same-id",
            "sequence": 1,
            "kind": "task",
            "scope": "T1",
            "contract_sha256": digest,
            "started_at": "2026-09-10T18:00:00-06:00",
            "ended_at": "2026-09-10T18:01:00-06:00",
            "outcome": "succeeded",
            "evidence": "task evidence",
        }
        duplicate = dict(attempt)
        duplicate["sequence"] = 2
        plan = render_plan(
            self.root,
            recovery={
                "schema": PLAN.SCHEMA_RECOVERY,
                "plan_id": "fixture-plan",
                "attempts": [attempt, duplicate],
            },
            include_recovery=True,
        )
        with self.assertRaises(PLAN.PlanError) as raised:
            PLAN.inspect_plan(plan)
        self.assertEqual(raised.exception.code, "duplicate_recovery_id")


if __name__ == "__main__":
    unittest.main()
