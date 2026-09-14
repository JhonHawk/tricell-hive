#!/usr/bin/env python3
"""Inspect and validate a harness-neutral Hive plan.

The contract block remains ordinary Markdown written for a human. Its exact
UTF-8 bytes are the only bytes covered by the digest. The status line,
authorization, recovery metadata, and execution table live outside that block
so the plan can advance without changing the approved contract.
"""

from __future__ import annotations

import argparse
from dataclasses import dataclass
from datetime import datetime
import hashlib
import json
from pathlib import Path
import re
import sys
from typing import Any, Mapping


SCHEMA_AUTHORIZATION = "hive-plan/authorization.v1"
SCHEMA_RECOVERY = "hive-plan/recovery.v1"
STATUSES = ("draft", "planned", "building", "built", "verified")
ACTION_PATTERN = re.compile(r"^[a-z][a-z0-9._-]*$")
DIGEST_PATTERN = re.compile(r"^[0-9a-f]{64}$")
DATE_PATTERN = re.compile(r"^\d{4}-\d{2}-\d{2}(?:[T ]\S+)?$")
RECOVERY_KINDS = frozenset(("task", "delegation", "fix", "review", "remote"))
RECOVERY_OUTCOMES = frozenset(
    ("started", "succeeded", "failed", "blocked", "interrupted", "unknown")
)
RECOVERY_TERMINAL_OUTCOMES = frozenset(
    ("succeeded", "failed", "blocked", "interrupted", "unknown")
)
RECOVERY_LIMITS: dict[str, int | None] = {
    "task": None,
    "delegation": 2,
    "fix": 3,
    "review": 2,
    "remote": 2,
}
WILDCARD_CHARS = frozenset("*?[]")
ACTION_ALLOWED_STATUSES = {
    "implement": frozenset(("planned", "building")),
    "verify": frozenset(("built", "verified")),
    "commit": frozenset(("verified",)),
    "push": frozenset(("verified",)),
    "pr": frozenset(("verified",)),
    "merge": frozenset(("verified",)),
    "deploy": frozenset(("verified",)),
}
HEADING_PATTERN = re.compile(r"(?m)^#{1,6}\s+\S")
STATUS_PATTERN = re.compile(
    r"(?m)^Status:\s*(draft|planned|building|built|verified)\s*$"
)

BLOCK_MARKERS = {
    "contract": (
        b"<!-- hive-plan:contract:start -->",
        b"<!-- hive-plan:contract:end -->",
    ),
    "authorization": (
        b"<!-- hive-plan:authorization:start -->",
        b"<!-- hive-plan:authorization:end -->",
    ),
    "recovery": (
        b"<!-- hive-plan:recovery:start -->",
        b"<!-- hive-plan:recovery:end -->",
    ),
}


class PlanError(ValueError):
    """A plan cannot be parsed or does not authorize the requested action."""

    def __init__(self, code: str, message: str):
        super().__init__(message)
        self.code = code
        self.message = message


@dataclass(frozen=True)
class PlanDocument:
    path: Path
    raw: bytes
    contract_bytes: bytes
    contract_markdown: str
    authorization: Mapping[str, Any]
    status: str
    recovery: Mapping[str, Any] | None

    @property
    def contract_sha256(self) -> str:
        return hashlib.sha256(self.contract_bytes).hexdigest()


def _duplicate_key_object(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for key, value in pairs:
        if key in result:
            raise PlanError("duplicate_json_key", f"duplicate JSON key: {key}")
        result[key] = value
    return result


def _reject_json_constant(value: str) -> None:
    raise PlanError("invalid_json", f"non-standard JSON constant: {value}")


def _extract_block(raw: bytes, block_name: str) -> bytes:
    try:
        start_marker, end_marker = BLOCK_MARKERS[block_name]
    except KeyError as exc:
        raise PlanError("internal_error", f"unknown plan block: {block_name}") from exc

    if raw.count(start_marker) != 1 or raw.count(end_marker) != 1:
        raise PlanError(
            "malformed_document",
            f"plan must contain exactly one {block_name} block",
        )

    start = raw.index(start_marker) + len(start_marker)
    try:
        end = raw.index(end_marker, start)
    except ValueError as exc:
        raise PlanError(
            "malformed_document",
            f"{block_name} block has no closing marker after its opening marker",
        ) from exc
    if end < start:
        raise PlanError("malformed_document", f"{block_name} block markers are reversed")
    return raw[start:end]


def _parse_json_fence(body: bytes, block_name: str) -> Mapping[str, Any]:
    try:
        text = body.decode("utf-8")
    except UnicodeDecodeError as exc:
        raise PlanError("invalid_utf8", f"{block_name} block is not UTF-8") from exc

    prefix = "\n```json\n"
    if not text.startswith(prefix):
        raise PlanError(
            "invalid_fence",
            f"{block_name} block must start with a JSON fence on the next line",
        )

    suffix = "\n```\n"
    if text.endswith(suffix):
        payload = text[len(prefix) : -len(suffix)]
    elif text.endswith("\n```"):
        payload = text[len(prefix) : -len("\n```")]
    else:
        raise PlanError(
            "invalid_fence",
            f"{block_name} block must end with a closing JSON fence",
        )

    if not payload.strip():
        raise PlanError("invalid_json", f"{block_name} block is empty")

    try:
        parsed = json.loads(
            payload,
            object_pairs_hook=_duplicate_key_object,
            parse_constant=_reject_json_constant,
        )
    except PlanError:
        raise
    except (TypeError, ValueError, json.JSONDecodeError) as exc:
        raise PlanError("invalid_json", f"{block_name} block is not valid JSON") from exc

    if not isinstance(parsed, dict):
        raise PlanError("invalid_json", f"{block_name} block must contain a JSON object")
    return parsed


def _read_block(raw: bytes, block_name: str) -> tuple[bytes, Mapping[str, Any]]:
    body = _extract_block(raw, block_name)
    return body, _parse_json_fence(body, block_name)


def _read_optional_block(
    raw: bytes,
    block_name: str,
    *,
    must_be_outside: tuple[str, ...] = (),
) -> Mapping[str, Any] | None:
    start_marker, end_marker = BLOCK_MARKERS[block_name]
    start_count = raw.count(start_marker)
    end_count = raw.count(end_marker)
    if start_count == 0 and end_count == 0:
        return None
    if start_count != 1 or end_count != 1:
        raise PlanError(
            "malformed_document",
            f"plan must contain exactly one {block_name} block when present",
        )

    marker_start = raw.index(start_marker)
    marker_end = raw.index(end_marker, marker_start) + len(end_marker)
    for frozen_block in must_be_outside:
        frozen_start_marker, frozen_end_marker = BLOCK_MARKERS[frozen_block]
        frozen_start = raw.index(frozen_start_marker)
        frozen_end = raw.index(frozen_end_marker, frozen_start) + len(frozen_end_marker)
        if frozen_start < marker_start < frozen_end or frozen_start < marker_end < frozen_end:
            raise PlanError(
                "malformed_document",
                f"{block_name} block must remain outside the {frozen_block} block",
            )

    return _read_block(raw, block_name)[1]


def _require_mapping(
    value: Any,
    label: str,
    *,
    error_code: str = "incomplete_authorization",
) -> Mapping[str, Any]:
    if not isinstance(value, dict):
        raise PlanError(error_code, f"{label} must be an object")
    return value


def _require_string(
    value: Any,
    label: str,
    *,
    error_code: str = "incomplete_authorization",
) -> str:
    if not isinstance(value, str) or not value.strip():
        raise PlanError(error_code, f"{label} must be a non-empty string")
    return value


def _require_string_list(value: Any, label: str, *, allow_empty: bool = False) -> list[str]:
    if not isinstance(value, list):
        raise PlanError("incomplete_authorization", f"{label} must be a list")
    if not allow_empty and not value:
        raise PlanError("incomplete_authorization", f"{label} must not be empty")
    result: list[str] = []
    for index, item in enumerate(value):
        result.append(_require_string(item, f"{label}[{index}]"))
    return result


def _validate_condition(value: Any, label: str) -> dict[str, str]:
    condition = _require_mapping(value, label)
    requirement = _require_string(condition.get("requirement"), f"{label}.requirement")
    evidence = condition.get("evidence", "")
    if not isinstance(evidence, str):
        raise PlanError("incomplete_authorization", f"{label}.evidence must be a string")
    return {"requirement": requirement, "evidence": evidence}


def _validate_conditions(value: Any, label: str) -> list[dict[str, str]]:
    if not isinstance(value, list):
        raise PlanError("incomplete_authorization", f"{label} must be a list")
    return [_validate_condition(item, f"{label}[{index}]") for index, item in enumerate(value)]


def _reject_wildcard(value: str, label: str) -> None:
    if any(character in value for character in WILDCARD_CHARS):
        raise PlanError(
            "wildcard_target",
            f"{label} cannot contain wildcard characters; list exact targets",
        )


def _parse_recovery_time(value: Any, label: str) -> datetime:
    timestamp = _require_string(value, label, error_code="incomplete_recovery")
    if not DATE_PATTERN.fullmatch(timestamp):
        raise PlanError("invalid_recovery_time", f"{label} is not ISO-like")
    try:
        return datetime.fromisoformat(timestamp.replace("Z", "+00:00"))
    except ValueError as exc:
        raise PlanError("invalid_recovery_time", f"{label} is not a valid timestamp") from exc


def _validate_recovery_evidence(value: Any, label: str) -> Any:
    if isinstance(value, str):
        if not value.strip():
            raise PlanError("incomplete_recovery", f"{label} must not be empty")
        return value
    if isinstance(value, list) and value:
        for index, item in enumerate(value):
            _require_string(item, f"{label}[{index}]", error_code="incomplete_recovery")
        return value
    raise PlanError(
        "incomplete_recovery",
        f"{label} must be a non-empty string or list of strings",
    )


def _validate_recovery(
    recovery: Mapping[str, Any],
) -> dict[str, Any]:
    if recovery.get("schema") != SCHEMA_RECOVERY:
        raise PlanError(
            "unsupported_recovery_schema",
            f"recovery schema must be {SCHEMA_RECOVERY}",
        )

    plan_id = _require_string(
        recovery.get("plan_id"),
        "recovery.plan_id",
        error_code="incomplete_recovery",
    )
    attempts = recovery.get("attempts")
    if not isinstance(attempts, list):
        raise PlanError("incomplete_recovery", "recovery.attempts must be a list")

    normalized_attempts: list[dict[str, Any]] = []
    attempt_ids: set[str] = set()
    sequence_values: set[int] = set()
    transient_exception_scopes: set[tuple[str, str]] = set()
    previous_sequence = 0
    for index, attempt in enumerate(attempts):
        label = f"recovery.attempts[{index}]"
        attempt_map = _require_mapping(
            attempt,
            label,
            error_code="incomplete_recovery",
        )
        attempt_id = _require_string(
            attempt_map.get("id"),
            f"{label}.id",
            error_code="incomplete_recovery",
        )
        if attempt_id in attempt_ids:
            raise PlanError("duplicate_recovery_id", f"duplicate recovery attempt id: {attempt_id}")
        attempt_ids.add(attempt_id)

        sequence = attempt_map.get("sequence")
        if not isinstance(sequence, int) or isinstance(sequence, bool) or sequence < 1:
            raise PlanError(
                "invalid_recovery_sequence",
                f"{label}.sequence must be a positive integer",
            )
        if sequence in sequence_values or sequence <= previous_sequence:
            raise PlanError(
                "invalid_recovery_sequence",
                "recovery attempt sequence must be strictly increasing",
            )
        sequence_values.add(sequence)
        previous_sequence = sequence

        kind = _require_string(
            attempt_map.get("kind"),
            f"{label}.kind",
            error_code="incomplete_recovery",
        )
        if kind not in RECOVERY_KINDS:
            raise PlanError("unsupported_recovery_kind", f"unsupported recovery kind: {kind}")
        scope = _require_string(
            attempt_map.get("scope"),
            f"{label}.scope",
            error_code="incomplete_recovery",
        )
        attempt_digest = _require_string(
            attempt_map.get("contract_sha256"),
            f"{label}.contract_sha256",
            error_code="incomplete_recovery",
        )
        if not DIGEST_PATTERN.fullmatch(attempt_digest):
            raise PlanError("incomplete_recovery", f"{label}.contract_sha256 is not SHA-256")

        started_at = _parse_recovery_time(attempt_map.get("started_at"), f"{label}.started_at")
        ended_at_value = attempt_map.get("ended_at")
        ended_at: datetime | None = None
        if ended_at_value is not None:
            ended_at = _parse_recovery_time(ended_at_value, f"{label}.ended_at")
            try:
                ends_before_start = ended_at < started_at
            except TypeError as exc:
                raise PlanError(
                    "invalid_recovery_time",
                    f"{label}.started_at and ended_at must use compatible timezone forms",
                ) from exc
            if ends_before_start:
                raise PlanError("invalid_recovery_time", f"{label}.ended_at precedes started_at")

        outcome = _require_string(
            attempt_map.get("outcome"),
            f"{label}.outcome",
            error_code="incomplete_recovery",
        )
        if outcome not in RECOVERY_OUTCOMES:
            raise PlanError(
                "unsupported_recovery_outcome",
                f"unsupported recovery outcome: {outcome}",
            )
        if outcome == "started" and ended_at is not None:
            raise PlanError("invalid_recovery_outcome", f"{label}.started cannot have ended_at")
        if outcome in RECOVERY_TERMINAL_OUTCOMES and ended_at is None:
            raise PlanError("invalid_recovery_outcome", f"{label}.{outcome} requires ended_at")

        evidence = _validate_recovery_evidence(attempt_map.get("evidence"), f"{label}.evidence")
        predecessor = attempt_map.get("predecessor")
        if predecessor is not None:
            predecessor = _require_string(
                predecessor,
                f"{label}.predecessor",
                error_code="incomplete_recovery",
            )
            if predecessor not in attempt_ids:
                raise PlanError(
                    "invalid_recovery_reference",
                    f"{label}.predecessor must reference an earlier attempt",
                )

        task = attempt_map.get("task")
        if task is not None:
            task = _require_string(task, f"{label}.task", error_code="incomplete_recovery")

        exception = attempt_map.get("exception")
        if exception is not None:
            exception_map = _require_mapping(
                exception,
                f"{label}.exception",
                error_code="incomplete_recovery",
            )
            exception_type = _require_string(
                exception_map.get("type"),
                f"{label}.exception.type",
                error_code="incomplete_recovery",
            )
            if exception_type != "verified_transient":
                raise PlanError(
                    "unsupported_recovery_exception",
                    f"unsupported recovery exception: {exception_type}",
                )
            if kind != "remote" or outcome != "failed":
                raise PlanError(
                    "invalid_recovery_exception",
                    f"{label}.exception is valid only for failed remote attempts",
                )
            scope_key = (kind, scope)
            if scope_key in transient_exception_scopes:
                raise PlanError(
                    "duplicate_recovery_exception",
                    f"scope {scope} has more than one verified transient exception",
                )
            transient_exception_scopes.add(scope_key)
            exception_evidence = _validate_recovery_evidence(
                exception_map.get("evidence"),
                f"{label}.exception.evidence",
            )
            exception = {"type": exception_type, "evidence": exception_evidence}

        normalized_attempts.append(
            {
                "id": attempt_id,
                "sequence": sequence,
                "kind": kind,
                "scope": scope,
                "contract_sha256": attempt_digest,
                "started_at": attempt_map["started_at"],
                "ended_at": attempt_map.get("ended_at"),
                "outcome": outcome,
                "evidence": evidence,
                "predecessor": predecessor,
                "task": task,
                "exception": exception,
            }
        )

    return {
        "schema": SCHEMA_RECOVERY,
        "plan_id": plan_id,
        "attempts": normalized_attempts,
    }


def _validate_contract_markdown(contract_markdown: str) -> str:
    """Validate only the human-readable shape, not the plan's domain content."""

    if not contract_markdown.strip():
        raise PlanError("incomplete_plan", "contract Markdown cannot be empty")
    if not HEADING_PATTERN.search(contract_markdown):
        raise PlanError("incomplete_plan", "contract Markdown must contain a heading")
    if contract_markdown.lstrip().startswith("```json"):
        raise PlanError("invalid_contract", "contract must remain human-readable Markdown")

    for line in contract_markdown.splitlines():
        if re.match(r"^#{1,6}\s+\S", line):
            return line.strip().lstrip("#").strip()
    raise PlanError("incomplete_plan", "contract Markdown must contain a heading")


READABLE_LINE_MAX = 200


def _readability_warnings(document: "PlanDocument") -> list[dict[str, Any]]:
    """Advisory only: long contract lines outside fenced code, numbered by file line.
    Never rejects a plan."""

    start_marker = BLOCK_MARKERS["contract"][0]
    lines_before = document.raw[: document.raw.index(start_marker)].count(b"\n")
    warnings: list[dict[str, Any]] = []
    in_fence = False
    for number, line in enumerate(document.contract_markdown.splitlines(), start=lines_before + 1):
        if line.lstrip().startswith("```"):
            in_fence = not in_fence
            continue
        if not in_fence and len(line) > READABLE_LINE_MAX:
            warnings.append({"code": "long_line", "line": number, "length": len(line)})
    return warnings


def _parse_status(raw: bytes) -> str:
    try:
        raw.decode("utf-8")
    except UnicodeDecodeError as exc:
        raise PlanError("invalid_utf8", "plan must be UTF-8") from exc

    start_marker, end_marker = BLOCK_MARKERS["contract"]
    start = raw.index(start_marker) + len(start_marker)
    end = raw.index(end_marker, start)
    outside = raw[:start] + raw[end + len(end_marker) :]
    outside_text = outside.decode("utf-8")
    matches = list(STATUS_PATTERN.finditer(outside_text))
    if len(matches) != 1:
        raise PlanError(
            "incomplete_plan",
            "plan must contain exactly one Status line outside the contract block",
        )
    status = matches[0].group(1)
    if status not in STATUSES:
        raise PlanError("invalid_state", f"unsupported plan status: {status}")
    return status


def _validate_approval(approval: Mapping[str, Any]) -> None:
    _require_string(approval.get("evidence"), "authorization.approval.evidence")
    date = _require_string(approval.get("date"), "authorization.approval.date")
    if not DATE_PATTERN.fullmatch(date):
        raise PlanError("incomplete_authorization", "authorization.approval.date is not ISO-like")
    revision = approval.get("revision")
    if not isinstance(revision, int) or isinstance(revision, bool) or revision < 1:
        raise PlanError(
            "incomplete_authorization",
            "authorization.approval.revision must be a positive integer",
        )


def _validate_authorization(
    authorization: Mapping[str, Any],
    contract_sha256: str,
) -> dict[str, Any]:
    if authorization.get("schema") != SCHEMA_AUTHORIZATION:
        raise PlanError(
            "unsupported_schema",
            f"authorization schema must be {SCHEMA_AUTHORIZATION}",
        )

    plan_id = _require_string(authorization.get("plan_id"), "authorization.plan_id")
    stored_digest = _require_string(
        authorization.get("contract_sha256"), "authorization.contract_sha256"
    )
    if not DIGEST_PATTERN.fullmatch(stored_digest):
        raise PlanError("incomplete_authorization", "authorization digest is not SHA-256")
    if stored_digest != contract_sha256:
        raise PlanError(
            "contract_drift",
            "authorization digest does not match the current contract bytes",
        )

    approval = _require_mapping(authorization.get("approval"), "authorization.approval")
    _validate_approval(approval)

    grants = authorization.get("grants")
    if not isinstance(grants, list):
        raise PlanError("incomplete_authorization", "authorization.grants must be a list")
    grant_ids: set[str] = set()
    normalized_grants: list[dict[str, Any]] = []
    for index, grant in enumerate(grants):
        grant_map = _require_mapping(grant, f"authorization.grants[{index}]")
        grant_id = _require_string(grant_map.get("id"), f"authorization.grants[{index}].id")
        if grant_id in grant_ids:
            raise PlanError("duplicate_grant", f"duplicate grant id: {grant_id}")
        grant_ids.add(grant_id)
        action = _require_string(
            grant_map.get("action"), f"authorization.grants[{index}].action"
        )
        if not ACTION_PATTERN.fullmatch(action):
            raise PlanError("invalid_action", f"invalid action name: {action}")
        if action not in ACTION_ALLOWED_STATUSES:
            raise PlanError("unsupported_action", f"unsupported action name: {action}")
        targets = _require_string_list(
            grant_map.get("targets"),
            f"authorization.grants[{index}].targets",
        )
        for target_index, target in enumerate(targets):
            _reject_wildcard(target, f"authorization.grants[{index}].targets[{target_index}]")
        grant_evidence = _require_string(
            grant_map.get("evidence"),
            f"authorization.grants[{index}].evidence",
        )
        conditions = _validate_conditions(
            grant_map.get("conditions"),
            f"authorization.grants[{index}].conditions",
        )
        normalized_grants.append(
            {
                "id": grant_id,
                "action": action,
                "targets": targets,
                "conditions": conditions,
                "evidence": grant_evidence,
            }
        )

    revocations = authorization.get("revocations")
    if not isinstance(revocations, list):
        raise PlanError(
            "incomplete_authorization", "authorization.revocations must be a list"
        )
    revoked_ids: set[str] = set()
    for index, revocation in enumerate(revocations):
        revocation_map = _require_mapping(
            revocation, f"authorization.revocations[{index}]"
        )
        grant_id = _require_string(
            revocation_map.get("grant_id"),
            f"authorization.revocations[{index}].grant_id",
        )
        if grant_id not in grant_ids:
            raise PlanError("unknown_revocation", f"revoked grant does not exist: {grant_id}")
        if grant_id in revoked_ids:
            raise PlanError("duplicate_revocation", f"duplicate revocation: {grant_id}")
        revoked_ids.add(grant_id)
        _require_string(
            revocation_map.get("reason"),
            f"authorization.revocations[{index}].reason",
        )
        _require_string(
            revocation_map.get("evidence"),
            f"authorization.revocations[{index}].evidence",
        )
        revoked_at = _require_string(
            revocation_map.get("revoked_at"),
            f"authorization.revocations[{index}].revoked_at",
        )
        if not DATE_PATTERN.fullmatch(revoked_at):
            raise PlanError("incomplete_authorization", "revocation date is not ISO-like")

    active_grants = [grant for grant in normalized_grants if grant["id"] not in revoked_ids]
    return {
        "plan_id": plan_id,
        "grant_ids": grant_ids,
        "grants": normalized_grants,
        "active_grants": active_grants,
        "revoked_ids": revoked_ids,
    }


def _validate_action_state(action: str, status: str) -> None:
    if action not in ACTION_ALLOWED_STATUSES:
        raise PlanError("unsupported_action", f"unsupported requested action: {action}")
    if status not in ACTION_ALLOWED_STATUSES[action]:
        allowed = ", ".join(sorted(ACTION_ALLOWED_STATUSES[action]))
        raise PlanError(
            "invalid_action_state",
            f"action {action} is not allowed while plan status is {status}; "
            f"allowed statuses: {allowed}",
        )


def _bind_recovery(
    recovery: Mapping[str, Any] | None,
    authorization_info: Mapping[str, Any],
) -> Mapping[str, Any] | None:
    if recovery is None:
        return None
    plan_id = authorization_info.get("plan_id")
    if isinstance(plan_id, str) and recovery.get("plan_id") != plan_id:
        raise PlanError(
            "recovery_plan_mismatch",
            "recovery.plan_id does not match authorization.plan_id",
        )
    return recovery


def _recovery_summary(recovery: Mapping[str, Any] | None) -> dict[str, Any]:
    if recovery is None:
        return {
            "status": "unknown",
            "history_known": False,
            "schema": None,
            "attempts": 0,
            "recorded_entries": 0,
            "open": 0,
            "ambiguous": 0,
            "completed": 0,
        }

    attempts = list(recovery["attempts"])
    open_attempts = [attempt for attempt in attempts if attempt["outcome"] == "started"]
    ambiguous_attempts = [
        attempt
        for attempt in attempts
        if attempt["outcome"] in {"unknown", "interrupted"}
    ]
    completed_attempts = [
        attempt
        for attempt in attempts
        if attempt["outcome"] in {"succeeded", "failed", "blocked"}
    ]
    return {
        "status": "known",
        "history_known": True,
        "schema": recovery["schema"],
        "attempts": len(attempts),
        "recorded_entries": len(attempts),
        "open": len(open_attempts),
        "ambiguous": len(ambiguous_attempts),
        "completed": len(completed_attempts),
    }


def _recovery_attempt_stats(attempts: list[Mapping[str, Any]], kind: str) -> dict[str, Any]:
    open_count = sum(attempt["outcome"] == "started" for attempt in attempts)
    ambiguous_count = sum(
        attempt["outcome"] in {"unknown", "interrupted"} for attempt in attempts
    )
    transient_exemptions = sum(
        (attempt.get("exception") or {}).get("type") == "verified_transient"
        for attempt in attempts
    )
    if kind == "delegation":
        charged = len(attempts)
    elif kind == "review":
        charged = len(attempts)
    elif kind == "fix":
        charged = sum(attempt["outcome"] == "failed" for attempt in attempts)
    elif kind == "remote":
        charged = sum(
            attempt["outcome"] == "failed"
            and (attempt.get("exception") or {}).get("type") != "verified_transient"
            for attempt in attempts
        )
    else:
        charged = 0
    return {
        "total": len(attempts),
        "completed": len(attempts) - open_count - ambiguous_count,
        "open": open_count,
        "ambiguous": ambiguous_count,
        "failed": sum(attempt["outcome"] == "failed" for attempt in attempts),
        "succeeded": sum(attempt["outcome"] == "succeeded" for attempt in attempts),
        "charged": charged,
        "transient_exemptions": transient_exemptions,
    }


def recovery_check(path: str | Path, kind: str, scope: str) -> dict[str, Any]:
    """Return a read-only continuation decision for one recovery scope."""

    document = _load(Path(path))
    authorization_info = _validate_authorization(
        document.authorization,
        document.contract_sha256,
    )
    recovery = _bind_recovery(document.recovery, authorization_info)
    if not isinstance(kind, str) or kind not in RECOVERY_KINDS:
        raise PlanError("unsupported_recovery_kind", f"unsupported recovery kind: {kind}")
    _require_string(scope, "requested recovery scope", error_code="incomplete_recovery")

    if recovery is None:
        return {
            "ok": True,
            "valid": False,
            "allowed": False,
            "decision": "reconcile",
            "status": "unknown",
            "history_known": False,
            "reason": "legacy plan has no recovery history; reconcile before repeating work",
            "plan_id": authorization_info["plan_id"],
            "kind": kind,
            "scope": scope,
            "contract_sha256": document.contract_sha256,
            "attempts": {
                "total": 0,
                "completed": 0,
                "open": 0,
                "ambiguous": 0,
                "failed": 0,
                "succeeded": 0,
                "charged": 0,
                "transient_exemptions": 0,
            },
            "recorded_entries": 0,
            "limit": RECOVERY_LIMITS[kind],
        }

    matching_attempts = [
        attempt
        for attempt in recovery["attempts"]
        if attempt["kind"] == kind and attempt["scope"] == scope
    ]
    attempts = _recovery_attempt_stats(matching_attempts, kind)
    limit = RECOVERY_LIMITS[kind]
    if attempts["open"] or attempts["ambiguous"]:
        decision = "reconcile"
        status = "reconcile"
        reason = "open or ambiguous attempt requires reconciliation before repeating work"
        allowed = False
    elif kind == "review" and attempts["succeeded"]:
        decision = "completed"
        status = "completed"
        reason = "the review scope already has a successful round and cannot be reopened"
        allowed = False
    elif limit is not None and attempts["charged"] >= limit:
        decision = "exhausted"
        status = "exhausted"
        reason = f"{kind} recovery budget is exhausted for scope {scope}"
        allowed = False
    else:
        decision = "continue"
        status = "ready"
        reason = "no open or ambiguous attempt and the recovery budget remains available"
        allowed = True

    return {
        "ok": True,
        "valid": allowed,
        "allowed": allowed,
        "decision": decision,
        "status": status,
        "history_known": True,
        "reason": reason,
        "plan_id": authorization_info["plan_id"],
        "kind": kind,
        "scope": scope,
        "contract_sha256": document.contract_sha256,
        "attempts": attempts,
        "recorded_entries": len(recovery["attempts"]),
        "limit": limit,
    }


def _load(path: Path) -> PlanDocument:
    try:
        raw = path.read_bytes()
    except OSError as exc:
        raise PlanError("read_error", f"cannot read plan: {path}") from exc

    contract_bytes = _extract_block(raw, "contract")
    try:
        contract_markdown = contract_bytes.decode("utf-8")
    except UnicodeDecodeError as exc:
        raise PlanError("invalid_utf8", "contract block is not UTF-8") from exc
    _validate_contract_markdown(contract_markdown)
    status = _parse_status(raw)
    _, authorization = _read_block(raw, "authorization")
    recovery_payload = _read_optional_block(
        raw,
        "recovery",
        must_be_outside=("contract", "authorization"),
    )
    recovery = (
        _validate_recovery(recovery_payload)
        if recovery_payload is not None
        else None
    )
    return PlanDocument(
        path=path,
        raw=raw,
        contract_bytes=contract_bytes,
        contract_markdown=contract_markdown,
        authorization=authorization,
        status=status,
        recovery=recovery,
    )


def digest_plan(path: str | Path) -> str:
    """Return the SHA-256 of the exact bytes between contract markers."""

    plan_path = Path(path)
    try:
        raw = plan_path.read_bytes()
    except OSError as exc:
        raise PlanError("read_error", f"cannot read plan: {plan_path}") from exc
    contract_bytes = _extract_block(raw, "contract")
    try:
        contract_bytes.decode("utf-8")
    except UnicodeDecodeError as exc:
        raise PlanError("invalid_utf8", "contract block is not UTF-8") from exc
    return hashlib.sha256(contract_bytes).hexdigest()


def inspect_plan(path: str | Path) -> dict[str, Any]:
    """Return the immutable digest and current mutable authorization state."""

    document = _load(Path(path))
    title = _validate_contract_markdown(document.contract_markdown)
    try:
        authorization_info = _validate_authorization(
            document.authorization,
            document.contract_sha256,
        )
        _bind_recovery(document.recovery, authorization_info)
        has_conditions = any(
            not condition["evidence"].strip()
            for grant in authorization_info["active_grants"]
            for condition in grant["conditions"]
        )
        authorization_status = "conditional" if has_conditions else "valid"
    except PlanError as exc:
        authorization_info = {"error": exc.code, "message": exc.message}
        authorization_status = "unbound" if exc.code == "contract_drift" else "invalid"

    can_implement = any(
        document.status in ACTION_ALLOWED_STATUSES["implement"]
        and grant["action"] == "implement"
        and all(condition["evidence"].strip() for condition in grant["conditions"])
        for grant in authorization_info.get("active_grants", [])
    )

    return {
        "ok": True,
        "plan_id": authorization_info.get("plan_id"),
        "title": title,
        "status": document.status,
        "can_implement": (
            document.status != "draft"
            and can_implement
        ),
        "contract_sha256": document.contract_sha256,
        "authorization": {
            "status": authorization_status,
            "active_grants": len(authorization_info.get("active_grants", [])),
            "revoked_grants": len(authorization_info.get("revoked_ids", [])),
        },
        "recovery": _recovery_summary(document.recovery),
        "warnings": _readability_warnings(document),
    }


def validate_plan(path: str | Path, action: str, target: str) -> dict[str, Any]:
    """Validate one exact action/target against current plan authorization."""

    document = _load(Path(path))
    if document.status == "draft":
        raise PlanError("draft_not_executable", "draft plans cannot authorize execution")

    authorization_info = _validate_authorization(
        document.authorization,
        document.contract_sha256,
    )
    _bind_recovery(document.recovery, authorization_info)

    if not isinstance(action, str) or not ACTION_PATTERN.fullmatch(action):
        raise PlanError("invalid_action", f"invalid requested action: {action}")
    _require_string(target, "requested target")
    _reject_wildcard(target, "requested target")

    all_action_grants = [
        grant for grant in authorization_info["grants"] if grant["action"] == action
    ]
    if not all_action_grants:
        raise PlanError("undeclared_operation", f"requested action is not granted: {action}")

    matching_grants = [
        grant for grant in all_action_grants if target in grant["targets"]
    ]
    if not matching_grants:
        raise PlanError(
            "undeclared_destination",
            f"requested target is not granted for {action}: {target}",
        )

    active_matching = [
        grant
        for grant in authorization_info["active_grants"]
        if grant["action"] == action and target in grant["targets"]
    ]
    if not active_matching:
        raise PlanError(
            "grant_revoked",
            f"all grants for action/target are revoked: {action} {target}",
        )

    _validate_action_state(action, document.status)

    satisfied_matching = [
        grant
        for grant in active_matching
        if all(condition["evidence"].strip() for condition in grant["conditions"])
    ]
    if not satisfied_matching:
        unresolved_conditions = [
            condition["requirement"]
            for grant in active_matching
            for condition in grant["conditions"]
            if not condition["evidence"].strip()
        ]
        raise PlanError(
            "conditions_unresolved",
            "authorization conditions require independent evidence: "
            + "; ".join(unresolved_conditions),
        )

    return {
        "ok": True,
        "valid": True,
        "plan_id": authorization_info["plan_id"],
        "status": document.status,
        "action": action,
        "target": target,
        "contract_sha256": document.contract_sha256,
        "grant_ids": [grant["id"] for grant in satisfied_matching],
        "conditions": [
            {
                "requirement": condition["requirement"],
                "evidence": condition["evidence"],
            }
            for grant in satisfied_matching
            for condition in grant["conditions"]
        ],
    }


def _error_payload(error: PlanError) -> dict[str, Any]:
    return {
        "ok": False,
        "valid": False,
        "error": {"code": error.code, "message": error.message},
    }


def _build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)

    for command in ("inspect", "digest"):
        subparser = commands.add_parser(command)
        subparser.add_argument("plan", type=Path)

    validate = commands.add_parser("validate")
    validate.add_argument("plan", type=Path)
    validate.add_argument("--action", required=True)
    validate.add_argument("--target", required=True)

    recovery_check_parser = commands.add_parser("recovery-check")
    recovery_check_parser.add_argument("plan", type=Path)
    recovery_check_parser.add_argument("--kind", required=True)
    recovery_check_parser.add_argument("--scope", required=True)
    return parser


def main(argv: list[str] | None = None) -> int:
    args = _build_parser().parse_args(argv)
    try:
        if args.command == "digest":
            result: dict[str, Any] = {"ok": True, "sha256": digest_plan(args.plan)}
        elif args.command == "inspect":
            result = inspect_plan(args.plan)
        elif args.command == "recovery-check":
            result = recovery_check(args.plan, args.kind, args.scope)
        else:
            result = validate_plan(args.plan, args.action, args.target)
    except PlanError as exc:
        print(json.dumps(_error_payload(exc), ensure_ascii=False, sort_keys=True))
        return 1

    print(json.dumps(result, ensure_ascii=False, sort_keys=True))
    if args.command == "recovery-check" and result.get("decision") != "continue":
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main())
