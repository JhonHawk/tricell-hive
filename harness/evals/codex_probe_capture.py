"""Parse Codex probe events into a small, payload-free evidence summary."""

from __future__ import annotations

import json
import os
import re
import shlex
from datetime import date
from typing import Any

from runner import _command_reads


_RULE_PATH = re.compile(r"(?:^|\s)([^\s\"']*/references/([A-Za-z0-9._-]+\.md))")
_HOLD_MARKER = "rule-delivery"
_DECISIONS = {"deny", "denied", "blocked"}
_PRETOOL_HOLD_MARKER = "blocked by pretooluse hook"
_RULE_BASENAMES = {
    "development-principles.md",
    "identifier-language.md",
    "patterns-antipatterns.md",
    "security.md",
    "test-gate.md",
    "typescript-standards.md",
}
_THREAD_ID = re.compile(r"[A-Za-z0-9_-]{1,128}\Z")
_PATCH_CALL = re.compile(
    r'\btools\.apply_patch\s*\(\s*("(?:\\.|[^"\\])*"\s*)\)', re.DOTALL
)
_PATCH_FILE_HEADER = re.compile(r"^\*\*\* (?:Add|Update|Delete) File: (.+?)\s*$", re.MULTILINE)


def _jsonl_objects(text: str) -> list[dict[str, Any]]:
    events: list[dict[str, Any]] = []
    for line in text.splitlines():
        try:
            event = json.loads(line)
        except (json.JSONDecodeError, TypeError):
            continue
        if isinstance(event, dict):
            events.append(event)
    return events


def _rule_path(value: str) -> tuple[str, str] | None:
    match = _RULE_PATH.search(value)
    if not match:
        return None
    if match.group(2) not in _RULE_BASENAMES:
        return None
    return _normal_path(match.group(1)), match.group(2)


def _normal_path(value: str, base_dir: str | None = None) -> str:
    if value.startswith("~/"):
        value = os.path.expanduser(value)
    elif base_dir and not os.path.isabs(value):
        value = os.path.join(base_dir, value)
    return os.path.normpath(value).replace("\\", "/")


def _change_paths(item: dict[str, Any], base_dir: str) -> set[str]:
    changes = item.get("changes")
    if not isinstance(changes, list):
        return set()
    return {
        _normal_path(change["path"], base_dir)
        for change in changes
        if isinstance(change, dict) and isinstance(change.get("path"), str)
    }


def _explicit_hold(error: Any) -> tuple[int | None, str | None, str | None] | None:
    """Accept only a structured deny with a rule-delivery reference."""
    if not isinstance(error, dict):
        return None
    decision = error.get("permissionDecision", error.get("decision"))
    reason = error.get("permissionDecisionReason", error.get("reason"))
    if not isinstance(decision, str) or decision.lower() not in _DECISIONS:
        return None
    if not isinstance(reason, str) or _HOLD_MARKER not in reason.lower():
        return None
    path = _rule_path(reason)
    return (len(reason), path[0] if path else None, path[1] if path else None)


def _read_paths(item: dict[str, Any], base_dir: str) -> list[tuple[str, str]]:
    if item.get("type") != "command_execution":
        return []
    if str(item.get("status", "")).lower() not in {"completed", "succeeded", "success"}:
        return []
    if item.get("exit_code") != 0:
        return []
    command = item.get("command")
    if isinstance(command, list) and all(isinstance(value, str) for value in command):
        command = shlex.join(command)
    if not isinstance(command, str):
        return []
    _, paths = _command_reads(command)
    recognized = []
    for value in paths:
        portable = _normal_path(value, base_dir)
        marker = "/references/"
        if marker not in portable:
            continue
        basename = portable.rsplit("/", 1)[-1]
        if basename in _RULE_BASENAMES:
            recognized.append((portable, basename))
    return recognized


def _rule_references(text: str) -> dict[str, str]:
    references: dict[str, str] = {}
    for match in _RULE_PATH.finditer(text):
        basename = match.group(2)
        if basename in _RULE_BASENAMES:
            path = _normal_path(match.group(1))
            references[path] = basename
    return references


def _wrapper_hold(output: Any) -> tuple[int, dict[str, str]] | None:
    if not isinstance(output, list):
        return None
    for block in output:
        if not isinstance(block, dict) or block.get("type") not in {"input_text", "text"}:
            continue
        text = block.get("text")
        if not isinstance(text, str) or _PRETOOL_HOLD_MARKER not in text.casefold():
            continue
        references = _rule_references(text)
        if references:
            return len(text), references
    return None


def _wrapper_patch_paths(call_input: str, base_dir: str) -> set[str]:
    """Read file headers from a literal apply_patch argument without evaluating code."""
    match = _PATCH_CALL.search(call_input)
    if not match:
        return set()
    try:
        patch_text = json.loads(match.group(1).strip())
    except json.JSONDecodeError:
        return set()
    if not isinstance(patch_text, str):
        return set()
    return {
        _normal_path(header.group(1).strip(), base_dir)
        for header in _PATCH_FILE_HEADER.finditer(patch_text)
    }


def _structured_rollout_holds(text: str) -> dict[str, tuple[int | None, str | None, str | None]]:
    events = _jsonl_objects(text)
    calls: dict[str, str] = {}
    holds: dict[str, tuple[int | None, str | None, str | None]] = {}
    for event in events:
        if event.get("type") != "response_item":
            continue
        payload = event.get("payload")
        if not isinstance(payload, dict):
            continue
        call_id = payload.get("call_id")
        if not isinstance(call_id, str):
            continue
        kind = payload.get("type")
        if kind in {"custom_tool_call", "function_call"}:
            name = payload.get("name")
            if isinstance(name, str):
                calls[call_id] = name
        elif kind in {"custom_tool_call_output", "function_call_output"}:
            if calls.get(call_id) not in {"apply_patch", "write_file"}:
                continue
            output = payload.get("output", payload.get("content"))
            candidates = [output]
            if isinstance(output, list):
                candidates.extend(
                    block.get("text")
                    for block in output
                    if isinstance(block, dict) and block.get("type") in {"input_text", "text"}
                )
            for candidate in candidates:
                if isinstance(candidate, str):
                    try:
                        candidate = json.loads(candidate)
                    except json.JSONDecodeError:
                        continue
                hold = _explicit_hold(candidate)
                if hold is not None:
                    holds[call_id] = hold
                    break
    return holds


def select_exact_rollout(
    sessions_root: str | os.PathLike[str], thread_id: str, session_date: date
) -> str | None:
    """Select one rollout by filename and matching first-line session metadata."""
    if not isinstance(thread_id, str) or not _THREAD_ID.fullmatch(thread_id):
        return None
    root = os.path.realpath(os.fspath(sessions_root))
    date_dir = os.path.join(
        root,
        f"{session_date.year:04d}",
        f"{session_date.month:02d}",
        f"{session_date.day:02d}",
    )
    if not os.path.isdir(date_dir):
        return None
    matches: list[str] = []
    for candidate in sorted(
        os.path.join(date_dir, name)
        for name in os.listdir(date_dir)
        if name.startswith("rollout-") and thread_id in name and name.endswith(".jsonl")
    ):
        resolved = os.path.realpath(candidate)
        if os.path.commonpath((root, resolved)) != root or not os.path.isfile(resolved):
            continue
        try:
            with open(resolved, encoding="utf-8") as stream:
                metadata = json.loads(stream.readline())
        except (OSError, json.JSONDecodeError):
            continue
        if (
            isinstance(metadata, dict)
            and metadata.get("type") == "session_meta"
            and isinstance(metadata.get("payload"), dict)
            and metadata["payload"].get("id") == thread_id
        ):
            matches.append(resolved)
    return matches[0] if len(matches) == 1 else None


def _rollout_change_paths(item: dict[str, Any], base_dir: str) -> set[str]:
    changes = item.get("changes")
    if isinstance(changes, dict):
        return {_normal_path(path, base_dir) for path in changes if isinstance(path, str)}
    return _change_paths(item, base_dir)


def parse_rollout_probe(
    text: str, *, target_path: str | os.PathLike[str]
) -> dict[str, Any]:
    """Extract hold/read/retry evidence from one already-selected Codex rollout."""
    events = _jsonl_objects(text)
    target = _normal_path(os.fspath(target_path))
    base_dir = os.path.dirname(target)
    writes: list[dict[str, Any]] = []
    outputs: dict[str, tuple[int, Any]] = {}
    reads: list[tuple[str, str, int]] = []
    completed_targets: list[int] = []

    for sequence, event in enumerate(events):
        if event.get("type") == "response_item":
            payload = event.get("payload")
            if not isinstance(payload, dict):
                continue
            kind = payload.get("type")
            call_id = payload.get("call_id")
            if not isinstance(call_id, str):
                continue
            if kind == "custom_tool_call":
                tool_name = payload.get("name")
                call_input = payload.get("input")
                if (
                    tool_name == "exec"
                    and isinstance(call_input, str)
                    and target in _wrapper_patch_paths(call_input, base_dir)
                ):
                    writes.append({
                        "call_id": call_id,
                        "input": call_input,
                        "sequence": sequence,
                        "output_sequence": None,
                        "hold": None,
                    })
            elif kind == "custom_tool_call_output":
                outputs[call_id] = (sequence, payload.get("output"))
                matching_write = next(
                    (write for write in writes if write["call_id"] == call_id), None
                )
                if matching_write is not None:
                    matching_write["output_sequence"] = sequence
                    matching_write["hold"] = _wrapper_hold(payload.get("output"))
            continue

        if event.get("type") != "event_msg":
            continue
        payload = event.get("payload")
        item = payload.get("item") if isinstance(payload, dict) else None
        if not isinstance(item, dict):
            continue
        item_type = str(item.get("type", "")).lower()
        status = str(item.get("status", "")).lower()
        if item_type == "commandexecution":
            normalized = {**item, "type": "command_execution"}
            for path, basename in _read_paths(normalized, base_dir):
                reads.append((path, basename, sequence))
        elif item_type == "filechange" and status in {"completed", "succeeded", "success"}:
            if target in _rollout_change_paths(item, base_dir):
                completed_targets.append(sequence)

    holds = [
        (write, write["hold"])
        for write in writes
        if write["hold"] is not None
    ]
    required_paths = {
        path
        for _, (_, references) in holds
        for path in references
    }
    retry_matches: bool | None = None
    retry_succeeded: bool | None = None
    if holds:
        held_write, (reason_length, held_references) = holds[-1]
        retry_matches = False
        retry_succeeded = False if held_references else None
        if held_references:
            for retry in writes:
                if retry["sequence"] <= held_write["sequence"]:
                    continue
                same_input = retry["input"] == held_write["input"]
                retry_matches = retry_matches or same_input
                read_paths = {
                    path
                    for path, _, read_sequence in reads
                    if held_write["hold"] is not None
                    and held_write["output_sequence"] is not None
                    and held_write["output_sequence"] < read_sequence < retry["sequence"]
                }
                has_completed_file_change = any(
                    retry["sequence"] < change_sequence < retry["output_sequence"]
                    for change_sequence in completed_targets
                ) if retry["output_sequence"] is not None else False
                output_sequence = retry["output_sequence"]
                if (
                    same_input
                    and set(held_references).issubset(read_paths)
                    and has_completed_file_change
                    and output_sequence is not None
                ):
                    retry_succeeded = True
                    break
        if retry_succeeded is False and retry_matches is None:
            retry_matches = False

    first_write_sequence = writes[0]["sequence"] if writes else None
    reads_before_first_write = [
        basename
        for _, basename, read_sequence in reads
        if first_write_sequence is not None and read_sequence < first_write_sequence
    ]
    basenames = {basename for _, basename, _ in reads}
    basenames.update(
        basename for _, (_, references) in holds for basename in references.values()
    )
    return {
        "write_attempts": len(writes) if writes else None,
        "hold_rounds": len(holds) if holds else None,
        "reason_lengths_chars": None,
        "hook_output_block_lengths_chars": [length for _, (length, _) in holds] if holds else None,
        "rule_reads": len(reads) if reads else None,
        "rule_basenames": sorted(basenames),
        "reads_before_first_write": reads_before_first_write,
        "required_rule_basenames": sorted(
            basename for _, (_, references) in holds for basename in references.values()
        ),
        "retry_succeeded": retry_succeeded,
        "retry_input_matches": retry_matches,
        "write_mechanism": "apply_patch" if writes else None,
        "hook_output_category": "pretool_hook_hold_text" if holds else None,
        "target_change_completed": bool(completed_targets),
        "unclassified_error_outputs": None,
    }


def parse_probe_stream(
    text: str, *, target_path: str | os.PathLike[str], rollout_text: str | None = None
) -> tuple[dict[str, Any], str | None]:
    """Return safe event evidence and the thread ID for in-memory correlation.

    The caller must not persist or print the returned thread ID. Unknown event
    types and unrecognized payloads are ignored rather than guessed at.
    """
    events = _jsonl_objects(text)
    rollout_holds = _structured_rollout_holds(rollout_text) if rollout_text is not None else {}
    thread_id = next(
        (
            event.get("thread_id")
            for event in events
            if event.get("type") == "thread.started"
            and isinstance(event.get("thread_id"), str)
        ),
        None,
    )
    target = _normal_path(os.fspath(target_path))
    base_dir = os.path.dirname(target)
    write_ids: set[str] = set()
    write_start_sequences: dict[str, list[int]] = {}
    successful_writes: list[tuple[str, int]] = []
    holds: list[tuple[int | None, str | None, str | None, int]] = []
    reads: list[tuple[str, str, int]] = []
    terminal_seen = False

    for sequence, event in enumerate(events):
        event_type = event.get("type")
        if event_type == "turn.completed":
            terminal_seen = True
        item = event.get("item")
        if not isinstance(item, dict):
            continue
        item_type = item.get("type")
        item_id = item.get("id")
        status = str(item.get("status", "")).lower()
        if item_type == "file_change":
            is_target = target in _change_paths(item, base_dir)
            if event_type == "item.started" and is_target and isinstance(item_id, str):
                write_ids.add(item_id)
                write_start_sequences.setdefault(item_id, []).append(sequence)
            if event_type == "item.completed" and is_target and status in {"completed", "succeeded", "success"}:
                if isinstance(item_id, str):
                    successful_writes.append((item_id, sequence))
            if event_type == "item.completed" and is_target:
                hold = _explicit_hold(item.get("error"))
                if hold is None and isinstance(item_id, str):
                    hold = rollout_holds.get(item_id)
                if hold is not None:
                    holds.append((hold[0], hold[1], hold[2], sequence))
        if event_type == "item.completed":
            reads.extend(
                (path, basename, sequence)
                for path, basename in _read_paths(item, base_dir)
            )

    retry_succeeded: bool | None = None
    if holds:
        latest_hold_sequence = holds[-1][3]
        required_paths = {path for _, path, _, _ in holds if path}
        retry_succeeded = None
        if required_paths:
            for item_id, write_sequence in successful_writes:
                retry_starts = [
                    start_sequence
                    for start_sequence in write_start_sequences.get(item_id, [])
                    if latest_hold_sequence < start_sequence < write_sequence
                ]
                for start_sequence in retry_starts:
                    read_paths = {
                        path
                        for path, _, read_sequence in reads
                        if latest_hold_sequence < read_sequence < start_sequence
                    }
                    if required_paths.issubset(read_paths):
                        retry_succeeded = True
                        break
                if retry_succeeded is True:
                    break
            if retry_succeeded is None:
                retry_succeeded = False

    reason_lengths = [length for length, _, _, _ in holds if length is not None]
    basenames = {name for _, _, name, _ in holds if name}
    basenames.update(name for _, name, _ in reads)
    summary: dict[str, Any] = {
        "write_attempts": len(write_ids) if write_ids else None,
        "hold_rounds": len(holds) if holds else None,
        "rule_reads": len(reads) if reads else None,
        "retry_succeeded": retry_succeeded,
        "reason_lengths_chars": reason_lengths if holds else None,
        "rule_basenames": sorted(basenames),
        "terminal_event": terminal_seen if events else None,
    }
    if rollout_text is not None:
        rollout_summary = parse_rollout_probe(rollout_text, target_path=target)
        for key, value in rollout_summary.items():
            if key != "rule_basenames" and (
                value is not None or key == "unclassified_error_outputs"
            ):
                summary[key] = value
        if rollout_summary["rule_basenames"]:
            summary["rule_basenames"] = rollout_summary["rule_basenames"]
    return summary, thread_id


def parse_rollout(text: str, *, target_path: str | os.PathLike[str]) -> dict[str, Any]:
    """Compatibility name for safe summary of an exact rollout and target."""
    return parse_rollout_probe(text, target_path=target_path)
