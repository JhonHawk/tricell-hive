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


if __name__ == "__main__":
    unittest.main()
