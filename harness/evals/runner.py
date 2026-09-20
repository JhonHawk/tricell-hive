#!/usr/bin/env python3
"""Small, isolated, deterministic pilot runner for Hive skill evaluations.

The runner deliberately has no third-party dependencies.  It treats a missing
terminal event or missing skill-read evidence as unknown; unknown evidence is
never promoted to a passing result.
"""

from __future__ import annotations

import argparse
import fnmatch
import hashlib
import json
import os
import pwd
import re
import selectors
import shlex
import shutil
import signal
import subprocess
import sys
import tempfile
import time
from pathlib import Path
from typing import Any, Callable


ROOT = Path(__file__).resolve().parents[2]
DEFAULT_MANIFEST = ROOT / "harness" / "evals" / "cases.json"
HARNESS_NAMES = ("codex", "claude", "pi", "grok", "opencode")
EFFORTS = ("low", "medium", "high")
TERMINAL_TYPES = {
    "turn.completed",
    "response.completed",
    "run.completed",
    "task.completed",
    "session.completed",
    "agent.completed",
    "agent_end",
    "agent_settled",
    "session_end",
    "agent.end",
    "agent.settled",
    "session.end",
    "result",
}
FAILURE_WORDS = ("failed", "failure", "error", "cancelled", "canceled")
READ_NAMES = {"read", "read_file", "readfile", "cat", "sed", "view"}


class EvalError(ValueError):
    """Raised when a manifest or runner option is unsafe or malformed."""


def _read_native_claude_oauth_token(environment: dict[str, str]) -> str:
    """Read only the access token from Claude Code's default macOS Keychain item."""
    if "CLAUDE_CONFIG_DIR" in environment or "CLAUDE_SECURESTORAGE_CONFIG_DIR" in environment:
        raise EvalError(
            "Claude uses a custom config directory; supply CLAUDE_CODE_OAUTH_TOKEN explicitly"
        )

    account = environment.get("USER") or pwd.getpwuid(os.getuid()).pw_name
    if not re.fullmatch(r"[A-Za-z0-9._-]+", account):
        account = "claude-code-user"
    try:
        result = subprocess.run(
            [
                "/usr/bin/security",
                "find-generic-password",
                "-a",
                account,
                "-w",
                "-s",
                "Claude Code-credentials",
            ],
            capture_output=True,
            text=True,
            timeout=5,
            check=False,
        )
    except (OSError, subprocess.SubprocessError) as exc:
        raise EvalError("Claude native Keychain credential could not be read") from exc
    if result.returncode != 0:
        raise EvalError("Claude native Keychain credential is unavailable")
    try:
        credentials = json.loads(result.stdout)
    except (TypeError, json.JSONDecodeError) as exc:
        raise EvalError("Claude native Keychain credential has an unsupported format") from exc
    oauth = credentials.get("claudeAiOauth") if isinstance(credentials, dict) else None
    token = oauth.get("accessToken") if isinstance(oauth, dict) else None
    if not isinstance(token, str) or not token:
        raise EvalError("Claude native Keychain credential has no OAuth access token")
    return token


def _safe_relative(value: Any, label: str) -> str:
    if not isinstance(value, str) or not value or "\x00" in value:
        raise EvalError(f"{label} must be a non-empty relative path")
    portable = value.replace("\\", "/")
    path = Path(portable)
    if path.is_absolute() or portable.startswith("/") or ".." in path.parts:
        raise EvalError(f"{label} must stay below the evaluation workspace: {value!r}")
    if portable.startswith("~"):
        raise EvalError(f"{label} may not use home expansion: {value!r}")
    return portable


def _normalize_case(case: Any, index: int, arm: str | None = None) -> dict[str, Any]:
    if not isinstance(case, dict):
        raise EvalError(f"case {index} must be an object")
    for key in ("id", "skill", "category", "prompt"):
        if not isinstance(case.get(key), str) or not case[key]:
            raise EvalError(f"case {index} requires a non-empty {key}")
    if not isinstance(case.get("expected_activation"), bool):
        raise EvalError(f"case {case['id']} expected_activation must be boolean")

    skill = case["skill"]
    skill_by_arm = case.get("skill_by_arm", {})
    if not isinstance(skill_by_arm, dict) or any(key not in {"baseline", "candidate"} for key in skill_by_arm):
        raise EvalError(f"case {case['id']} skill_by_arm must map baseline/candidate to skill paths")
    if arm is not None and skill_by_arm:
        if arm not in skill_by_arm:
            raise EvalError(f"case {case['id']} has no target skill for arm {arm}")
        skill = skill_by_arm[arm]
    skill = _safe_relative(skill, f"case {case['id']} skill")
    fields = {
        "fixtures": {},
        "protected_files": [],
        "required_reads": [],
        "reads_before_write": [],
        "allowed_writes": [],
        "required_files": [],
        "output_contains": [],
        "file_contains": {},
    }
    fields.update({key: case.get(key, default) for key, default in fields.items()})
    if not isinstance(fields["fixtures"], dict):
        raise EvalError(f"case {case['id']} fixtures must be an object")
    fixtures: dict[str, str] = {}
    for path, content in fields["fixtures"].items():
        relative = _safe_relative(path, f"case {case['id']} fixture path")
        if not isinstance(content, str):
            raise EvalError(f"case {case['id']} fixture {path} content must be text")
        fixtures[relative] = content

    def paths(name: str) -> list[str]:
        values = fields[name]
        if not isinstance(values, list):
            raise EvalError(f"case {case['id']} {name} must be a list")
        result = []
        for value in values:
            result.append(_safe_relative(value, f"case {case['id']} {name}"))
        return result

    protected = paths("protected_files")
    reads = paths("required_reads")
    reads_before_write = paths("reads_before_write")
    allowed = paths("allowed_writes")
    required = paths("required_files")
    outputs = fields["output_contains"]
    if not isinstance(outputs, list) or not all(isinstance(item, str) for item in outputs):
        raise EvalError(f"case {case['id']} output_contains must be a list of strings")
    file_contains = fields["file_contains"]
    if not isinstance(file_contains, dict):
        raise EvalError(f"case {case['id']} file_contains must be an object")
    normalized_contains: dict[str, list[str]] = {}
    for path, values in file_contains.items():
        relative = _safe_relative(path, f"case {case['id']} file_contains path")
        if not isinstance(values, list) or not all(isinstance(item, str) for item in values):
            raise EvalError(f"case {case['id']} file_contains[{path!r}] must be strings")
        normalized_contains[relative] = values

    timeout = case.get("timeout_seconds")
    if timeout is not None and (not isinstance(timeout, (int, float)) or timeout <= 0):
        raise EvalError(f"case {case['id']} timeout_seconds must be positive")
    return {
        "id": case["id"],
        "skill": skill,
        "category": case["category"],
        "prompt": case["prompt"],
        "fixtures": fixtures,
        "protected_files": protected,
        "expected_activation": case["expected_activation"],
        "required_reads": reads,
        "reads_before_write": reads_before_write,
        "allowed_writes": allowed,
        "required_files": required,
        "output_contains": outputs,
        "file_contains": normalized_contains,
        **({"timeout_seconds": timeout} if timeout is not None else {}),
    }


def load_manifest(path: Path, arm: str | None = None) -> dict[str, Any]:
    """Load and validate the versioned pilot case manifest."""
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        raise EvalError(f"cannot read manifest {path}: {exc}") from exc
    if not isinstance(data, dict) or data.get("version") != 1:
        raise EvalError("manifest must be an object with version: 1")
    cases = data.get("cases")
    if not isinstance(cases, list) or not cases:
        raise EvalError("manifest cases must be a non-empty list")
    normalized = []
    seen: set[str] = set()
    for index, case in enumerate(cases):
        normalized_case = _normalize_case(case, index, arm)
        if normalized_case["id"] in seen:
            raise EvalError(f"duplicate case id: {normalized_case['id']}")
        seen.add(normalized_case["id"])
        normalized.append(normalized_case)
    return {"version": 1, "cases": normalized}


def _sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def _tree_snapshot(root: Path) -> dict[str, dict[str, Any]]:
    if not root.exists():
        return {}
    snapshot: dict[str, dict[str, Any]] = {}
    for path in sorted(root.rglob("*")):
        relative = path.relative_to(root).as_posix()
        if ".git" in Path(relative).parts:
            continue
        if path.is_symlink():
            snapshot[relative] = {"type": "symlink", "target": os.readlink(path)}
        elif path.is_file():
            snapshot[relative] = {
                "type": "file",
                "size": path.stat().st_size,
                "sha256": _sha256_file(path),
            }
        elif path.is_dir():
            snapshot[relative] = {"type": "directory"}
    return snapshot


def _tree_digest(root: Path) -> str:
    payload = json.dumps(_tree_snapshot(root), sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(payload.encode("utf-8")).hexdigest()


def _inside(path: Path, root: Path) -> bool:
    try:
        path.resolve().relative_to(root.resolve())
    except ValueError:
        return False
    return True


def _event_type(event: dict[str, Any]) -> str:
    value = event.get("type")
    if isinstance(value, str):
        return value.lower()
    item = event.get("item")
    if isinstance(item, dict) and isinstance(item.get("type"), str):
        return item["type"].lower()
    return ""


def _walk_values(value: Any, keys: set[str]) -> list[str]:
    found: list[str] = []
    if isinstance(value, dict):
        for key, child in value.items():
            if key.lower() in keys and isinstance(child, str):
                found.append(child)
            found.extend(_walk_values(child, keys))
    elif isinstance(value, list):
        for child in value:
            found.extend(_walk_values(child, keys))
    return found


def _simple_command_reads(command: str) -> tuple[str, list[str]]:
    """Recognize one bounded file-read command without executing it."""
    try:
        tokens = shlex.split(command)
    except ValueError:
        return "", []
    if not tokens:
        return "", []
    name = Path(tokens[0]).name
    if any(token in {"--help", "--version", "-h"} for token in tokens[1:]):
        return "", []
    if name == "cat":
        paths: list[str] = []
        options_ended = False
        for token in tokens[1:]:
            if token == "--":
                options_ended = True
                continue
            if not options_ended and token.startswith("-"):
                continue
            paths.append(token)
        return name, paths
    if name == "sed" and len(tokens) >= 4 and tokens[1] == "-n" and re.fullmatch(r"\d+(,\d+)?p", tokens[2]):
        paths = [token for token in tokens[3:] if token != "--" and not token.startswith("-")]
        return name, paths
    return "", []


def _split_shell_reads(command: str) -> tuple[list[str], list[str]] | None:
    """Split a shell command at safe separators, preserving quoted content."""
    segments: list[str] = []
    separators: list[str] = []
    current: list[str] = []
    quote = ""
    escaped = False
    index = 0
    while index < len(command):
        character = command[index]
        if escaped:
            current.append(character)
            escaped = False
        elif quote:
            current.append(character)
            if character == "\\" and quote == '"':
                escaped = True
            elif character == quote:
                quote = ""
        elif character in {"'", '"'}:
            current.append(character)
            quote = character
        elif character == "\\":
            current.append(character)
            escaped = True
        elif character == "&" and index + 1 < len(command) and command[index + 1] == "&":
            segment = "".join(current).strip()
            if not segment:
                return None
            segments.append(segment)
            separators.append("&&")
            current = []
            index += 1
        elif character in {";", "\n"}:
            segment = "".join(current).strip()
            if not segment:
                return None
            segments.append(segment)
            separators.append(";")
            current = []
        elif character in {"|", "<", ">", "`", "$", "#", "(", ")"} or character == "&":
            return None
        else:
            current.append(character)
        index += 1
    if quote or escaped:
        return None
    segment = "".join(current).strip()
    if not segment:
        return None
    segments.append(segment)
    if len(separators) != len(segments) - 1:
        return None
    return segments, separators


def _command_reads(command: str) -> tuple[str, list[str]]:
    """Recognize simple reads and safe straight-line command chains.

    An all-``&&`` chain with a successful shell exit proves each whitelisted
    read. For a semicolon/newline chain, only the final ``&&`` suffix is
    credited because the shell exit status does not prove earlier commands.
    """
    try:
        tokens = shlex.split(command)
    except ValueError:
        return "", []
    if len(tokens) == 3 and Path(tokens[0]).name in {"sh", "bash", "zsh"} and tokens[1] in {"-c", "-lc"}:
        command = tokens[2]
    split = _split_shell_reads(command)
    if split is None:
        return "", []
    segments, separators = split
    reads = [_simple_command_reads(segment) for segment in segments]
    if any(not name or not paths for name, paths in reads):
        return "", []
    if ";" in separators:
        start = max(index + 1 for index, separator in enumerate(separators) if separator == ";")
        reads = reads[start:]
    paths = [path for _, segment_paths in reads for path in segment_paths]
    names = {name for name, _ in reads}
    return (next(iter(names)) if len(names) == 1 else "read"), paths


def _claude_tool_paths(name: str, tool_input: Any) -> list[str]:
    if not isinstance(tool_input, dict):
        return []
    paths: list[str] = []
    for key in ("file_path", "path", "filePath"):
        value = tool_input.get(key)
        if isinstance(value, str):
            paths.append(value)
        elif isinstance(value, list):
            paths.extend(item for item in value if isinstance(item, str))
    if name.casefold() == "bash":
        command = tool_input.get("command")
        if isinstance(command, str):
            _, read_paths = _command_reads(command)
            paths.extend(read_paths)
    return paths


def _normalise_claude_trace(
    stdout: str, stderr: str, skill: str
) -> tuple[list[dict[str, Any]], list[dict[str, Any]], list[str], bool, bool]:
    """Normalize Claude Code stream-json events using completed tool results."""
    events: list[dict[str, Any]] = []
    tool_events: list[dict[str, Any]] = []
    skill_loads: list[str] = []
    pending: dict[str, dict[str, Any]] = {}
    terminal = False
    terminal_failure = False
    sequence = 0

    for source, text in (("stdout", stdout), ("stderr", stderr)):
        for line in text.splitlines():
            try:
                raw = json.loads(line)
            except json.JSONDecodeError:
                continue
            if not isinstance(raw, dict):
                continue
            event = {"source": source, **raw}
            events.append(event)
            event_type = str(event.get("type", "")).casefold()
            if event_type == "result":
                terminal = True
                subtype = str(event.get("subtype", "")).casefold()
                terminal_failure = bool(event.get("is_error")) or subtype not in {"", "success"}
            elif event_type == "assistant":
                message = event.get("message")
                blocks = message.get("content") if isinstance(message, dict) else None
                if isinstance(blocks, list):
                    for block in blocks:
                        if not isinstance(block, dict) or block.get("type") != "tool_use":
                            continue
                        call_id = block.get("id")
                        name = block.get("name")
                        if not isinstance(call_id, str) or not isinstance(name, str):
                            continue
                        pending[call_id] = {
                            "name": name,
                            "input": block.get("input"),
                            "sequence": sequence,
                        }
            elif event_type == "user":
                message = event.get("message")
                blocks = message.get("content") if isinstance(message, dict) else None
                if isinstance(blocks, list):
                    for block in blocks:
                        if not isinstance(block, dict) or block.get("type") != "tool_result":
                            continue
                        call_id = block.get("tool_use_id")
                        call = pending.pop(call_id, None) if isinstance(call_id, str) else None
                        if call is None:
                            continue
                        name = call["name"]
                        succeeded = block.get("is_error") is not True
                        paths = _claude_tool_paths(name, call.get("input"))
                        if name.casefold() == "skill" and isinstance(call.get("input"), dict):
                            skill_name = call["input"].get("skill") or call["input"].get("name")
                            if isinstance(skill_name, str):
                                paths.append(f"{skill_name}/SKILL.md")
                        tool_events.append({
                            "name": name,
                            "status": "completed" if succeeded else "failed",
                            "paths": paths,
                            "event_type": "tool_result",
                            "sequence": call["sequence"],
                        })
                        if succeeded and name.casefold() in {"read", "bash", "skill"}:
                            skill_loads.extend(paths)
            sequence += 1

    return events, tool_events, sorted(set(skill_loads)), terminal, terminal_failure


def _normalise_trace(
    stdout: str, stderr: str, skill: str, harness: str | None = None
) -> tuple[list[dict[str, Any]], list[dict[str, Any]], list[str], bool, bool]:
    if harness == "claude":
        return _normalise_claude_trace(stdout, stderr, skill)
    events: list[dict[str, Any]] = []
    tool_events: list[dict[str, Any]] = []
    skill_loads: list[str] = []
    terminal = False
    terminal_failure = False
    tool_state: dict[str, dict[str, Any]] = {}
    sequence = 0
    for source, text in (("stdout", stdout), ("stderr", stderr)):
        for line in text.splitlines():
            try:
                event = json.loads(line)
            except json.JSONDecodeError:
                continue
            if not isinstance(event, dict):
                continue
            event = {"source": source, **event}
            events.append(event)
            event_type = _event_type(event)
            compact_type = event_type.replace("_", ".")
            status_value = str(event.get("status", "")).lower()
            if not status_value and isinstance(event.get("item"), dict):
                status_value = str(event["item"].get("status", "")).lower()
            name_values = _walk_values(event, {"name", "tool", "tool_name", "toolname", "function"})
            name = next((item for item in name_values if isinstance(item, str)), "")
            item = event.get("item")
            item_type = item.get("type", "").lower() if isinstance(item, dict) else ""
            command_item = (
                item
                if item_type == "command_execution"
                else event
                if event_type == "command_execution"
                else None
            )
            is_tool = (
                "tool" in event_type
                or "function" in event_type
                or event_type == "command_execution"
                or name.lower() in READ_NAMES
                or item_type in {"tool_call", "function_call", "command_execution", "file_read", "read_file", "mcp_tool_call"}
            )
            if compact_type in TERMINAL_TYPES:
                terminal = True
            stop_reasons = [item.lower() for item in _walk_values(event, {"stop_reason", "stopreason"})]
            if any(reason in {"error", "aborted", "length", "incomplete"} for reason in stop_reasons):
                terminal_failure = True
            if any(word in event_type for word in FAILURE_WORDS) or status_value in FAILURE_WORDS:
                terminal_failure = terminal_failure or not is_tool
            if is_tool:
                if "started" in event_type or "start" in event_type or status_value in {"started", "in_progress"}:
                    tool_status = "started"
                elif any(word in event_type for word in FAILURE_WORDS) or status_value in FAILURE_WORDS or event.get("success") is False or event.get("isError") is True:
                    tool_status = "failed"
                else:
                    tool_status = "completed" if (
                        "completed" in event_type
                        or event_type.endswith("_end")
                        or event.get("success") is True
                        or status_value in {"completed", "succeeded", "success"}
                    ) else "unknown"
                paths = _walk_values(event, {"path", "file_path", "filepath", "filename"})
                if command_item is not None:
                    paths = []
                    exit_code = command_item.get("exit_code")
                    command_status = str(command_item.get("status", status_value)).lower()
                    if tool_status == "completed" and command_status in {"completed", "succeeded", "success"} and exit_code == 0:
                        name, paths = _command_reads(str(command_item.get("command", "")))
                    elif exit_code not in {None, 0} or command_status in FAILURE_WORDS:
                        tool_status = "failed"
                    else:
                        tool_status = "unknown"
                call_ids = _walk_values(event, {"tool_call_id", "toolcallid", "call_id", "callid"})
                call_id = call_ids[0] if call_ids else ""
                if call_id and call_id in tool_state:
                    prior = tool_state[call_id]
                    if not name:
                        name = str(prior.get("name", ""))
                    if not paths:
                        paths = list(prior.get("paths", []))
                record = {"name": name, "status": tool_status, "paths": paths, "event_type": event_type, "sequence": sequence}
                tool_events.append(record)
                if call_id:
                    tool_state[call_id] = {"name": name, "paths": paths}
                if tool_status == "completed" and name.lower() in READ_NAMES:
                    for candidate in paths:
                        portable = candidate.replace("\\", "/")
                        skill_loads.append(portable)
            sequence += 1
    return events, tool_events, sorted(set(skill_loads)), terminal, terminal_failure


def _status(status: str, detail: str | None = None) -> dict[str, Any]:
    result: dict[str, Any] = {"status": status}
    if detail:
        result["detail"] = detail
    return result


def _matches_any(path: str, patterns: list[str]) -> bool:
    return any(fnmatch.fnmatchcase(path, pattern) for pattern in patterns)


def _visible_output(stdout: str, stderr: str) -> str:
    """Return assistant-facing text while excluding tool arguments and traces."""
    visible: list[str] = []
    for text in (stdout, stderr):
        for line in text.splitlines():
            try:
                event = json.loads(line)
            except json.JSONDecodeError:
                visible.append(line)
                continue
            if not isinstance(event, dict):
                visible.append(line)
                continue
            event_type = _event_type(event)
            item = event.get("item")
            item_type = item.get("type", "").lower() if isinstance(item, dict) else ""
            if "tool" in event_type or "function" in event_type or item_type in {
                "tool_call", "function_call", "command_execution", "file_read", "read_file", "mcp_tool_call",
            }:
                continue
            values = _walk_values(event, {"text", "content", "message", "output", "delta", "result"})
            visible.extend(values)
    return "\n".join(visible)


def _resolved_models(events: list[dict[str, Any]]) -> list[str]:
    values = set(_walk_values(events, {"model", "model_name", "model_id"}))
    for event in events:
        usage = event.get("modelUsage")
        if isinstance(usage, dict):
            values.update(name for name in usage if isinstance(name, str))
    return sorted(values)


def _terminal_types(events: list[dict[str, Any]]) -> list[str]:
    return sorted({
        _event_type(event)
        for event in events
        if _event_type(event).replace("_", ".") in TERMINAL_TYPES
    })


def _sandbox_argv(argv: list[str], run_root: Path, workspace: Path, harness: str) -> list[str]:
    """Wrap harness processes with run-scoped writes and harness-specific local sockets."""
    sandbox = shutil.which("sandbox-exec") or "/usr/bin/sandbox-exec"
    if not Path(sandbox).is_file():
        raise EvalError("sandbox-exec is required for harness write containment")
    profile = run_root / "harness-sandbox.sb"
    root = str(run_root.resolve()).replace("\\", "\\\\").replace('"', '\\"')
    workspace_path = str(workspace.resolve()).replace("\\", "\\\\").replace('"', '\\"')
    network_rules = ""
    if harness == "grok":
        network_rules = (
            f'(allow network-bind (local unix-socket (subpath "{root}")))\n'
            f'(allow network-inbound (local unix-socket (subpath "{root}")))\n'
        )
    elif harness == "opencode":
        # Seatbelt's localhost matcher includes all host-owned IPs; runtime listener inspection is required.
        network_rules = (
            '(allow network-bind (local ip "localhost:*"))\n'
            '(allow network-inbound (local ip "localhost:*"))\n'
        )
    profile.write_text(
        "(version 1)\n"
        "(deny default)\n"
        "(allow process*)\n"
        "(allow sysctl-read)\n"
        "(allow file-read*)\n"
        f'(allow file-write* (subpath "{root}") (subpath "{workspace_path}"))\n'
        '(allow file-write* (literal "/dev/null"))\n'
        "(allow network-outbound)\n" + network_rules,
        encoding="utf-8",
    )
    return [sandbox, "-f", str(profile), *argv]


def _kill_process_group(process: subprocess.Popen[bytes], sig: signal.Signals = signal.SIGKILL) -> None:
    try:
        os.killpg(process.pid, sig)
    except (OSError, ProcessLookupError):
        if process.poll() is None:
            try:
                process.kill()
            except OSError:
                pass


class EvalRunner:
    """Run cases in disposable workspaces and emit machine-readable evidence."""

    def __init__(
        self,
        repo_root: Path = ROOT,
        source_root: Path | None = None,
        arm: str | None = None,
        harness: str = "codex",
        model: str = "gpt-5.6-luna",
        effort: str = "high",
        timeout: float = 120.0,
        command_builder: Callable[..., list[str]] | None = None,
        stream_log: Callable[[str], None] | None = None,
    ) -> None:
        if harness not in HARNESS_NAMES:
            raise EvalError(f"unsupported harness: {harness}")
        if effort not in EFFORTS:
            raise EvalError(f"unsupported effort: {effort}")
        if not model:
            raise EvalError("model must be non-empty")
        if timeout <= 0:
            raise EvalError("timeout must be positive")
        self.repo_root = repo_root.resolve()
        self.source_root = (source_root or self.repo_root).resolve()
        if arm not in {None, "baseline", "candidate"}:
            raise EvalError("arm must be baseline or candidate")
        self.arm = arm
        self.harness = harness
        self.model = model
        self.effort = effort
        self.runner_digest = _sha256_file(Path(__file__))
        self.timeout = timeout
        self.command_builder = command_builder or default_command_builder
        self.stream_log = stream_log or self._stream_to_stderr
        self._version_cache: dict[str, str | None] = {}
        self._active_process: subprocess.Popen[bytes] | None = None
        self._claude_oauth_token: str | None = None

    @staticmethod
    def _stream_to_stderr(chunk: str) -> None:
        sys.stderr.write(chunk)
        sys.stderr.flush()

    def _stage_skills(self, run_root: Path) -> tuple[Path, Path, dict[str, Any]]:
        generated = self.source_root / "harness" / "agents-skills"
        if not generated.is_dir():
            raise EvalError(f"generated skills directory is missing: {generated}")
        codex_core = self.source_root / "harness" / "AGENTS.md"
        claude_core = self.source_root / "global" / "CLAUDE.md"
        if not codex_core.is_file() or not claude_core.is_file():
            raise EvalError(f"source root is missing generated harness cores: {self.source_root}")
        for path in generated.rglob("*"):
            if path.is_symlink():
                target = path.resolve()
                if not _inside(target, generated):
                    raise EvalError(f"generated skill symlink escapes source root: {path}")
                raise EvalError(f"generated skill symlinks are not supported: {path}")
        staged = run_root / "staged-skills"
        shutil.copytree(generated, staged)
        codex_home = run_root / "codex-home"
        pi_home = run_root / "pi-agent"
        home = run_root / "home"
        claude_home = home / ".claude"
        grok_home = home / ".grok"
        opencode_config = run_root / "xdg_config" / "opencode" / "opencode.json"
        opencode_data_home = run_root / "xdg_data" / "opencode"
        workspace_core = run_root / "workspace-core" / "AGENTS.md"
        codex_home.mkdir()
        pi_home.mkdir()
        claude_home.mkdir(parents=True)
        grok_home.mkdir()
        opencode_data_home.mkdir(parents=True)
        shutil.copytree(staged, codex_home / "skills")
        shutil.copytree(staged, pi_home / "skills")
        shutil.copytree(staged, claude_home / "skills")
        shutil.copy2(codex_core, codex_home / "AGENTS.md")
        shutil.copy2(codex_core, pi_home / "AGENTS.md")
        shutil.copy2(claude_core, claude_home / "CLAUDE.md")
        workspace_core.parent.mkdir(parents=True)
        shutil.copy2(codex_core, workspace_core)
        (codex_home / "config.toml").write_text("# Hive eval isolated config\n", encoding="utf-8")
        (pi_home / "settings.json").write_text("{}\n", encoding="utf-8")
        skill_resource_paths = {
            f"{staged.resolve()}/*",
            f"{(home / '.agents' / 'skills')}/*",
            f"{(claude_home / 'skills')}/*",
        }
        opencode_permissions = [
            {"action": "skill", "resource": "flow-*", "effect": "ask"},
            {"action": "skill", "resource": "flow-research", "effect": "allow"},
            {"action": "websearch", "resource": "*", "effect": "deny"},
            {"action": "webfetch", "resource": "*", "effect": "deny"},
        ]
        for resource in sorted(skill_resource_paths):
            opencode_permissions.extend(
                [
                    {"action": "external_directory", "resource": resource, "effect": "allow"},
                    {"action": "read", "resource": resource, "effect": "allow"},
                ]
            )
        opencode_config.parent.mkdir(parents=True, exist_ok=True)
        opencode_config.write_text(
            json.dumps(
                {
                    "$schema": "https://opencode.ai/config.json",
                    "share": "disabled",
                    "plugins": [],
                    "mcp": {"servers": {}},
                    "permissions": opencode_permissions,
                },
                indent=2,
            ) + "\n",
            encoding="utf-8",
        )
        return staged, codex_home, {
            "codex_home": codex_home,
            "pi_home": pi_home,
            "claude_home": claude_home,
            "grok_home": grok_home,
            "opencode_data_home": opencode_data_home,
            "opencode_config": opencode_config,
            "codex_core": codex_home / "AGENTS.md",
            "claude_core": claude_home / "CLAUDE.md",
            "pi_core": pi_home / "AGENTS.md",
            "grok_core": claude_home / "CLAUDE.md",
            "workspace_core": workspace_core,
        }

    def _link_auth(self, env: dict[str, str], roots: dict[str, Path]) -> dict[str, Any]:
        if self.harness == "codex":
            old_root = Path(env.get("CODEX_HOME", "")) if env.get("CODEX_HOME") else Path.home() / ".codex"
            target_root = roots["codex_home"]
            candidates = (old_root / "auth.json", Path.home() / ".codex" / "auth.json")
        elif self.harness == "pi":
            old_root = Path(env.get("PI_CODING_AGENT_DIR", "")) if env.get("PI_CODING_AGENT_DIR") else Path.home() / ".pi" / "agent"
            target_root = roots["pi_home"]
            candidates = (old_root / "auth.json", Path.home() / ".pi" / "agent" / "auth.json")
        elif self.harness == "claude":
            supplied_token = env.get("CLAUDE_CODE_OAUTH_TOKEN")
            if supplied_token:
                return {"status": "credential-reused", "source": "environment"}
            if self._claude_oauth_token is None:
                self._claude_oauth_token = _read_native_claude_oauth_token(env)
            env["CLAUDE_CODE_OAUTH_TOKEN"] = self._claude_oauth_token
            return {"status": "credential-reused", "source": "native-keychain"}
        elif self.harness == "grok":
            old_root = Path(env.get("GROK_HOME", "")) if env.get("GROK_HOME") else Path.home() / ".grok"
            target_root = roots["grok_home"]
            candidates = (old_root / "auth.json", Path.home() / ".grok" / "auth.json")
        else:
            old_data_home = Path(env.get("XDG_DATA_HOME", "")) if env.get("XDG_DATA_HOME") else Path.home() / ".local" / "share"
            target_root = roots["opencode_data_home"]
            candidates = (
                old_data_home / "opencode" / "auth.json",
                Path.home() / ".local" / "share" / "opencode" / "auth.json",
            )
        for source in candidates:
            if source.is_file() or source.is_symlink():
                target_root.mkdir(parents=True, exist_ok=True)
                link = target_root / source.name
                if self.harness == "pi":
                    if link.exists() or link.is_symlink():
                        raise EvalError("isolated Pi auth destination already exists")
                    shutil.copyfile(source, link)
                    os.chmod(link, 0o600)
                    return {"status": "copied-isolated"}
                if not link.exists() and not link.is_symlink():
                    link.symlink_to(source)
                return {"status": "symlinked"}
        return {"status": "not-found"}

    def _environment(self, run_root: Path, roots: dict[str, Path]) -> tuple[dict[str, str], dict[str, Any]]:
        env = os.environ.copy()
        auth = self._link_auth(env, roots)
        env["HOME"] = str(run_root / "home")
        Path(env["HOME"]).mkdir(exist_ok=True)
        shared_skills = Path(env["HOME"]) / ".agents" / "skills"
        shared_skills.parent.mkdir(parents=True, exist_ok=True)
        shared_skills.symlink_to((run_root / "staged-skills").resolve(), target_is_directory=True)
        env["CODEX_HOME"] = str(roots["codex_home"])
        env["PI_CODING_AGENT_DIR"] = str(roots["pi_home"])
        env["GROK_HOME"] = str(roots["grok_home"])
        if self.harness == "opencode":
            env["OPENCODE_DISABLE_DEFAULT_PLUGINS"] = "1"
        env.pop("CLAUDE_CONFIG_DIR", None)
        if self.harness == "claude":
            claude_tmpdir = run_root / "t"
            claude_tmpdir.mkdir(exist_ok=True)
            env["CLAUDE_CODE_TMPDIR"] = str(claude_tmpdir)
            env["TMPDIR"] = str(claude_tmpdir)
        env["CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC"] = "1"
        env["HIVE_EVAL_SKILLS_DIR"] = str(run_root / "staged-skills")
        env["HIVE_EVAL_ISOLATED"] = "1"
        env["GIT_CONFIG_NOSYSTEM"] = "1"
        env["GIT_CONFIG_GLOBAL"] = str(run_root / "gitconfig")
        (run_root / "gitconfig").write_text("", encoding="utf-8")
        for name in ("XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_CACHE_HOME", "XDG_STATE_HOME"):
            value = run_root / name.lower().replace("_home", "")
            value.mkdir(parents=True, exist_ok=True)
            env[name] = str(value)
        for key in ("CLAUDE_PROJECT_DIR", "CLAUDE_CODE_ENTRYPOINT", "CODEX_SESSION_ID"):
            env.pop(key, None)
        isolation = {
            "config": "isolated",
            "auth": auth,
            "skills": "isolated HOME/.agents/skills and HOME/.claude/skills contain staged generated skills",
            "cores": "source-root generated cores staged in isolated harness roots",
            "memory": "disabled-by-isolation",
            "mcp": "disabled-by-isolation",
        }
        if self.harness == "opencode":
            isolation["network_listener"] = {
                "status": "not_verified",
                "limitation": "SBPL localhost filters match host-owned addresses, not loopback only; inspect the actual listener and stop if it is non-loopback.",
            }
        if self.harness == "pi" and auth["status"] == "copied-isolated":
            isolation["auth_file"] = "pi-agent/auth.json copied with mode 0600; excluded from retained artifacts"
        return env, isolation

    @staticmethod
    def _init_workspace_repo(workspace: Path, env: dict[str, str]) -> None:
        git = shutil.which("git")
        if git is None:
            raise EvalError("git is required to create the disposable repository fixture")
        commands = [
            [git, "init", "--quiet", "--template="],
            [git, "add", "--all"],
            [git, "-c", "core.hooksPath=/dev/null", "-c", "commit.gpgSign=false",
             "-c", "user.name=Fixture", "-c", "user.email=fixture@localhost",
             "commit", "--quiet", "--allow-empty", "-m", "test: initialize disposable fixture"],
        ]
        for command in commands:
            result = subprocess.run(
                command, cwd=workspace, env=env, capture_output=True, text=True, check=False
            )
            if result.returncode != 0:
                raise EvalError(f"disposable git repository initialization failed: {result.stderr.strip()}")

    def _version(self, executable: str, env: dict[str, str], cwd: Path) -> str | None:
        if executable in self._version_cache:
            return self._version_cache[executable]
        resolved = shutil.which(executable, path=env.get("PATH")) or executable
        version_env = env
        if self.harness == "claude" and "CLAUDE_CODE_OAUTH_TOKEN" in env:
            version_env = env.copy()
            version_env.pop("CLAUDE_CODE_OAUTH_TOKEN", None)
        try:
            result = subprocess.run(
                [resolved, "--version"], cwd=cwd, env=version_env, capture_output=True, text=True, timeout=5, check=False
            )
        except (OSError, subprocess.SubprocessError):
            value = None
        else:
            value = (result.stdout.strip() or result.stderr.strip() or None) if result.returncode == 0 else None
        self._version_cache[executable] = value
        return value

    def _run_process(self, argv: list[str], cwd: Path, env: dict[str, str], timeout: float | None = None) -> dict[str, Any]:
        try:
            return self._run_process_impl(argv, cwd, env, timeout)
        except BaseException:
            process = self._active_process
            if process is not None:
                _kill_process_group(process)
                try:
                    process.wait(timeout=1)
                except subprocess.TimeoutExpired:
                    process.kill()
                for stream in (process.stdout, process.stderr):
                    if stream is not None:
                        stream.close()
            self._active_process = None
            raise

    def _run_process_impl(self, argv: list[str], cwd: Path, env: dict[str, str], timeout: float | None = None) -> dict[str, Any]:
        if not argv or any(not isinstance(arg, str) for arg in argv):
            raise EvalError("command builder must return a non-empty argv list of strings")
        started_at = time.time()
        try:
            process = subprocess.Popen(
                argv,
                cwd=cwd,
                env=env,
                stdin=subprocess.DEVNULL,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                start_new_session=True,
            )
        except OSError as exc:
            return {
                "status": "failed",
                "exit_code": None,
                "terminal_event": False,
                "timed_out": False,
                "error": str(exc),
                "stdout": "",
                "stderr": "",
                "duration_seconds": 0,
            }
        self._active_process = process
        assert process.stdout is not None and process.stderr is not None
        selector = selectors.DefaultSelector()
        selector.register(process.stdout, selectors.EVENT_READ, "stdout")
        selector.register(process.stderr, selectors.EVENT_READ, "stderr")
        chunks = {"stdout": [], "stderr": []}
        claude_token = env.get("CLAUDE_CODE_OAUTH_TOKEN") if self.harness == "claude" else None
        timed_out = False
        deadline = time.monotonic() + (timeout if timeout is not None else self.timeout)
        killed_at: float | None = None
        while selector.get_map() or process.poll() is None:
            now = time.monotonic()
            if now >= deadline and killed_at is None:
                timed_out = True
                killed_at = now
                try:
                    os.killpg(process.pid, signal.SIGKILL)
                except (OSError, ProcessLookupError):
                    if process.poll() is None:
                        process.kill()
            if killed_at is not None and now - killed_at >= 1.0:
                break
            if not selector.get_map():
                time.sleep(0.05)
                continue
            remaining = max(0.01, deadline - now) if killed_at is None else 0.1
            events = selector.select(min(0.1, remaining))
            if not events and process.poll() is not None:
                # Pipes still need draining; the next select iteration does so.
                continue
            for key, _ in events:
                data = os.read(key.fileobj.fileno(), 65536)
                if not data:
                    selector.unregister(key.fileobj)
                    key.fileobj.close()
                    continue
                text = data.decode("utf-8", errors="replace")
                chunks[key.data].append(text)
                if claude_token is None:
                    try:
                        self.stream_log(text)
                    except BaseException:
                        _kill_process_group(process)
                        for open_key in list(selector.get_map().values()):
                            try:
                                selector.unregister(open_key.fileobj)
                            except KeyError:
                                pass
                            open_key.fileobj.close()
                        selector.close()
                        try:
                            process.wait(timeout=1)
                        except subprocess.TimeoutExpired:
                            process.kill()
                        raise
        for key in list(selector.get_map().values()):
            try:
                selector.unregister(key.fileobj)
            except KeyError:
                pass
            key.fileobj.close()
        selector.close()
        _kill_process_group(process)
        try:
            exit_code = process.wait(timeout=1)
        except subprocess.TimeoutExpired:
            process.kill()
            exit_code = process.wait()
        stdout = "".join(chunks["stdout"])
        stderr = "".join(chunks["stderr"])
        if claude_token is not None:
            stdout = stdout.replace(claude_token, "[REDACTED]")
            stderr = stderr.replace(claude_token, "[REDACTED]")
            self.stream_log(stdout)
            self.stream_log(stderr)
        trace, tool_events, skill_loads, terminal, terminal_failure = _normalise_trace(
            stdout, stderr, "", self.harness
        )
        if timed_out:
            process_status = "failed"
        elif terminal and not terminal_failure and exit_code == 0:
            process_status = "passed"
        elif terminal_failure or (exit_code is not None and exit_code != 0):
            process_status = "failed"
        else:
            process_status = "unknown"
        self._active_process = None
        return {
            "status": process_status,
            "exit_code": exit_code,
            "terminal_event": terminal,
            "terminal_event_types": _terminal_types(trace),
            "terminal_failure_event": terminal_failure,
            "timed_out": timed_out,
            "stdout": stdout,
            "stderr": stderr,
            "trace": trace,
            "tool_events": tool_events,
            "skill_loads": skill_loads,
            "duration_seconds": round(max(0, time.time() - started_at), 3),
        }

    def run_case(
        self,
        raw_case: dict[str, Any],
        parent_dir: Path | None = None,
        retain_dir: Path | None = None,
        artifact_label: str = "1",
    ) -> dict[str, Any]:
        case = _normalize_case(raw_case, 0, self.arm)
        temporary_kwargs = {"dir": parent_dir} if parent_dir is not None else {}
        if parent_dir is None and self.harness in {"claude", "grok"}:
            temporary_kwargs["dir"] = Path("/tmp")
        temporary_prefix = {
            "claude": "hive-cc-",
            "grok": "hive-gk-",
        }.get(self.harness, f"hive-eval-{case['id']}-")
        with tempfile.TemporaryDirectory(prefix=temporary_prefix, **temporary_kwargs) as temporary:
            run_root = Path(temporary)
            workspace = run_root / "workspace"
            workspace.mkdir()
            staged, _, roots = self._stage_skills(run_root)
            env, isolation = self._environment(run_root, roots)
            if self.harness == "opencode":
                shutil.copy2(roots["workspace_core"], workspace / "AGENTS.md")
            for relative, content in case["fixtures"].items():
                target = workspace / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                if not _inside(target, workspace):
                    raise EvalError(f"fixture escapes workspace: {relative}")
                target.write_text(content, encoding="utf-8")
            self._init_workspace_repo(workspace, env)
            isolation["git_fixture"] = "committed-clean"
            before = _tree_snapshot(workspace)
            source_skill = self.source_root / "global" / "skills" / case["skill"]
            generated_skill = self.source_root / "harness" / "agents-skills" / case["skill"]
            if not source_skill.is_dir() or not generated_skill.is_dir():
                raise EvalError(f"skill source/generated directory is missing: {case['skill']}")
            source_digest = _tree_digest(source_skill)
            generated_digest = _tree_digest(generated_skill)
            argv = self.command_builder(
                harness=self.harness,
                model=self.model,
                effort=self.effort,
                skill=case["skill"],
                category=case["category"],
                prompt=case["prompt"],
                cwd=workspace,
                environment=env,
            )
            argv = list(argv)
            if not argv:
                raise EvalError("command builder returned an empty argv")
            cli_version = self._version(argv[0], env, workspace)
            if self.harness in {"claude", "pi", "grok", "opencode"}:
                argv = _sandbox_argv(argv, run_root, workspace, self.harness)
            execution = self._run_process(argv, workspace, env, case.get("timeout_seconds"))
            events, tool_events, skill_loads, terminal, terminal_failure = _normalise_trace(
                execution["stdout"], execution["stderr"], case["skill"], self.harness
            )
            execution.update(
                {
                    "trace": events,
                    "tool_events": tool_events,
                    "skill_loads": skill_loads,
                    "terminal_event": terminal,
                    "terminal_event_types": _terminal_types(events),
                    "terminal_failure_event": terminal_failure,
                }
            )
            after = _tree_snapshot(workspace)
            protected: dict[str, Any] = {}
            for relative in case["protected_files"]:
                old = before.get(relative)
                new = after.get(relative)
                if old is None:
                    protected[relative] = _status("unknown", "protected fixture was absent")
                elif new is None:
                    protected[relative] = _status("failed", "protected file was deleted")
                elif old == new:
                    protected[relative] = _status("passed")
                else:
                    protected[relative] = _status("failed", "protected file changed")
            write_tool_events = [
                event for event in tool_events
                if event.get("status") == "completed"
                and str(event.get("name", "")).casefold() in {"write", "edit", "apply_patch", "write_file"}
            ]
            out_of_workspace_tool_writes: set[str] = set()
            for event in write_tool_events:
                for raw_path in event.get("paths", []):
                    target = Path(str(raw_path))
                    if not target.is_absolute():
                        target = workspace / target
                    if not _inside(target, workspace):
                        out_of_workspace_tool_writes.add(str(target).replace("\\", "/"))

            unexpected = []
            for relative in sorted(set(before) | set(after)):
                old = before.get(relative)
                new = after.get(relative)
                if old == new or (old and old["type"] == "directory" and new and new["type"] == "directory"):
                    continue
                if old is None and new and new["type"] == "directory" and any(
                    pattern.startswith(relative + "/") for pattern in case["allowed_writes"]
                ):
                    continue
                if old is None and new and new["type"] == "directory" and any(
                    candidate.startswith(relative + "/") and _matches_any(candidate, case["allowed_writes"])
                    for candidate, candidate_value in after.items()
                    if candidate_value["type"] == "file"
                ):
                    continue
                if (new and new["type"] == "symlink") or (old and old["type"] == "symlink") or not _matches_any(relative, case["allowed_writes"]):
                    unexpected.append(relative)
            unexpected.extend(sorted(out_of_workspace_tool_writes))
            unexpected = sorted(set(unexpected))
            write_status = _status("passed") if not unexpected else _status("failed", "unexpected or out-of-workspace writes")
            written_files = sorted(
                relative
                for relative, value in after.items()
                if value["type"] == "file" and before.get(relative) != value
            )
            activation_status = "unknown"
            activation_detail = "no successful skill read event observed"
            loaded_target = any(
                path == f"{case['skill']}/SKILL.md"
                or path.endswith(f"/{case['skill']}/SKILL.md")
                for path in skill_loads
            )
            if loaded_target and case["expected_activation"]:
                activation_status, activation_detail = "passed", None
            elif loaded_target and not case["expected_activation"]:
                activation_status, activation_detail = "failed", "skill was loaded"
            elif not loaded_target and not case["expected_activation"] and execution["status"] == "passed":
                activation_status, activation_detail = "passed", None
            activation = _status(activation_status, activation_detail)

            checks: dict[str, Any] = {"output_contains": {}, "required_files": {}, "file_contains": {}}
            combined_output = _visible_output(execution["stdout"], execution["stderr"])
            for needle in case["output_contains"]:
                if needle in combined_output:
                    checks["output_contains"][needle] = _status("passed")
                elif execution["status"] == "unknown":
                    checks["output_contains"][needle] = _status("unknown", "process did not provide terminal evidence")
                else:
                    checks["output_contains"][needle] = _status("failed", "substring not found")
            for pattern in case["required_files"]:
                matches = [relative for relative in after if fnmatch.fnmatchcase(relative, pattern) and after[relative]["type"] == "file"]
                fresh = [relative for relative in matches if relative not in before or before[relative] != after[relative]]
                if fresh:
                    checks["required_files"][pattern] = _status("passed")
                elif execution["status"] == "unknown":
                    checks["required_files"][pattern] = _status("unknown", "process did not provide terminal evidence")
                else:
                    checks["required_files"][pattern] = _status("failed", "required file was not created or changed")
            for pattern, needles in case["file_contains"].items():
                matches = [relative for relative in after if fnmatch.fnmatchcase(relative, pattern) and after[relative]["type"] == "file"]
                if not matches:
                    checks["file_contains"][pattern] = _status("failed", "file not found")
                    continue
                fresh_matches = [relative for relative in matches if relative not in before or before[relative] != after[relative]]
                candidates = fresh_matches or matches
                valid = False
                missing_by_file: dict[str, list[str]] = {}
                for relative in candidates:
                    text = (workspace / relative).read_text(encoding="utf-8", errors="replace")
                    missing = [needle for needle in needles if needle not in text]
                    missing_by_file[relative] = missing
                    valid = valid or not missing
                checks["file_contains"][pattern] = _status("passed") if valid else _status("failed", f"missing: {missing_by_file}")
            read_checks: dict[str, Any] = {}
            for required in case["required_reads"]:
                found = any(path.endswith(required) or path == required for path in skill_loads)
                read_checks[required] = _status("passed") if found else _status("unknown", "no successful read evidence")
            checks["required_reads"] = read_checks
            first_write_sequence = min(
                (
                    event["sequence"]
                    for event in tool_events
                    if event.get("status") == "completed"
                    and str(event.get("name", "")).casefold() in {"write", "edit", "apply_patch", "write_file"}
                    and event.get("paths")
                ),
                default=None,
            )
            order_checks: dict[str, Any] = {}
            for required in case["reads_before_write"]:
                read_sequences = [
                    event["sequence"]
                    for event in tool_events
                    if event.get("status") == "completed"
                    and any(str(path).replace("\\", "/").endswith(required) for path in event.get("paths", []))
                ]
                if not read_sequences or first_write_sequence is None:
                    order_checks[required] = _status("unknown", "completed read or file-target tool sequence is not observable")
                elif min(read_sequences) < first_write_sequence:
                    order_checks[required] = _status("passed")
                elif min(read_sequences) == first_write_sequence:
                    order_checks[required] = _status("unknown", "read and write were requested in the same harness event")
                else:
                    order_checks[required] = _status("failed", "file target was chosen before the required reference read")
            checks["reads_before_write"] = order_checks
            read_events = [
                event for event in tool_events
                if event.get("status") == "completed"
                and str(event.get("name", "")).casefold() in READ_NAMES | {"bash"}
                and event.get("paths")
            ]
            operations = {
                "terminal": {
                    "status": "observed" if terminal else "not_verified",
                    "event_types": _terminal_types(events),
                    "failure": terminal_failure,
                },
                "reads": {
                    "status": "observed" if read_events else "not_verified",
                    "completed_calls": len(read_events),
                    "paths": sorted({
                        str(path).replace("\\", "/")
                        for event in read_events
                        for path in event.get("paths", [])
                    }),
                },
                "write_tool_calls": {
                    "status": "observed" if write_tool_events else "not_verified",
                    "completed_calls": len(write_tool_events),
                    "paths": sorted({
                        str(path).replace("\\", "/")
                        for event in write_tool_events
                        for path in event.get("paths", [])
                    }),
                    "out_of_workspace_targets": sorted(out_of_workspace_tool_writes),
                },
                "workspace_writes": {
                    "status": write_status["status"],
                    "files_changed": written_files,
                },
            }
            check_statuses = [value["status"] for group in checks.values() for value in group.values()]
            check_statuses.append(write_status["status"])
            check_statuses.append(activation["status"])
            check_statuses.extend(item["status"] for item in protected.values())
            if execution["status"] == "failed" or "failed" in check_statuses:
                outcome_status = "failed"
            elif execution["status"] == "unknown" or "unknown" in check_statuses:
                outcome_status = "unknown"
            else:
                outcome_status = "passed"
            core_hashes = {
                name: _sha256_file(path)
                for name, path in (
                    ("codex", roots["codex_core"]),
                    ("claude", roots["claude_core"]),
                    ("pi", roots["pi_core"]),
                    ("grok", roots["grok_core"]),
                    ("opencode", roots["workspace_core"]),
                )
            }
            config_payload = {
                "harness": self.harness,
                "arm": self.arm,
                "source_root": str(self.source_root),
                "model": self.model,
                "effort": self.effort,
                "argv": argv,
                "cores": core_hashes,
                "isolated": True,
                "config_files": {
                    "codex": _sha256_file(roots["codex_home"] / "config.toml"),
                    "pi": _sha256_file(roots["pi_home"] / "settings.json"),
                    "opencode": _sha256_file(roots["opencode_config"]),
                },
            }
            config_digest = hashlib.sha256(json.dumps(config_payload, sort_keys=True).encode("utf-8")).hexdigest()
            artifact_evidence: dict[str, Any] = {"status": "not-retained", "files": [], "omitted_files": []}
            if retain_dir is not None:
                safe_case = "".join(character if character.isalnum() or character in "-_" else "_" for character in case["id"])
                artifact_root = retain_dir / safe_case / artifact_label
                artifact_root.mkdir(parents=True, exist_ok=True)
                retained: list[str] = []
                omitted_secret_files: list[dict[str, str]] = []
                if self.harness == "pi" and isolation.get("auth", {}).get("status") == "copied-isolated":
                    omitted_secret_files.append({
                        "path": "pi-agent/auth.json",
                        "reason": "isolated Pi credentials are excluded from retained artifacts",
                    })
                secret = env.get("CLAUDE_CODE_OAUTH_TOKEN")
                secret_bytes = secret.encode("utf-8") if secret else None
                for relative, value in after.items():
                    if value["type"] != "file":
                        continue
                    if relative not in before or before[relative] != value:
                        source = workspace / relative
                        if secret_bytes is not None:
                            overlap = b""
                            contains_secret = False
                            with source.open("rb") as artifact_file:
                                while True:
                                    chunk = artifact_file.read(65536)
                                    if not chunk:
                                        break
                                    block = overlap + chunk
                                    if secret_bytes in block:
                                        contains_secret = True
                                        break
                                    overlap = block[-(len(secret_bytes) - 1):] if len(secret_bytes) > 1 else b""
                            if contains_secret:
                                omitted_secret_files.append({
                                    "path": relative,
                                    "reason": "contains Claude OAuth token",
                                })
                                continue
                        destination = artifact_root / relative
                        destination.parent.mkdir(parents=True, exist_ok=True)
                        shutil.copy2(source, destination)
                        os.chmod(destination, 0o600)
                        retained.append(relative)
                artifact_evidence = {
                    "status": "retained-with-security-omissions" if omitted_secret_files else "retained",
                    "root": str(artifact_root),
                    "files": sorted(retained),
                    "omitted_files": omitted_secret_files,
                }
            return {
                "id": case["id"],
                "skill": case["skill"],
                "category": case["category"],
                "arm": self.arm,
                "source_root": str(self.source_root),
                "harness": self.harness,
                "model": self.model,
                "effort": self.effort,
                "outcome": {"status": outcome_status, "checks": checks},
                "activation": activation,
                "process": {key: value for key, value in execution.items() if key not in {"stdout", "stderr", "trace", "tool_events", "skill_loads"}},
                "style": _status("unverified", "initial pilot has no LLM judge"),
                "evidence": {
                    "trace": execution["trace"],
                    "stdout": execution["stdout"],
                    "stderr": execution["stderr"],
                    "tool_events": execution["tool_events"],
                    "skill_loads": execution["skill_loads"],
                    "fixture_before": before,
                    "fixture_after": after,
                    "protected_files": protected,
                    "unexpected_writes": unexpected,
                    "writes": write_status,
                    "operations": operations,
                    "isolation": isolation,
                    "artifacts": artifact_evidence,
                },
                "hashes": {
                    "runner": self.runner_digest,
                    "source_skill": source_digest,
                    "generated_skill": generated_digest,
                    "cores": core_hashes,
                    "staged_skills": _tree_digest(staged),
                    "config": config_digest,
                    "model": {"requested": self.model, "resolved": _resolved_models(execution["trace"]) or None, "cli_version": cli_version},
                },
            }

    def dry_run(self, cases: list[dict[str, Any]], repeat: int = 1) -> dict[str, Any]:
        results = []
        for case in cases:
            argv = self.command_builder(
                harness=self.harness,
                model=self.model,
                effort=self.effort,
                skill=case["skill"],
                category=case["category"],
                prompt=case["prompt"],
                cwd=Path("<isolated-workspace>"),
                environment={"HIVE_EVAL_SKILLS_DIR": "<staged-skills>"},
            )
            results.append({
                "id": case["id"],
                "skill": case["skill"],
                "category": case["category"],
                "repeat": repeat,
                "argv": list(argv),
                "source_root": str(self.source_root),
                "arm": self.arm,
                "isolation": "fresh temporary workspace; source-root cores and generated skills staged; user hooks/MCP/memory isolated",
                "run": False,
            })
        return {
            "version": 1,
            "mode": "dry-run",
            "arm": self.arm,
            "source_root": str(self.source_root),
            "harness": self.harness,
            "model": self.model,
            "effort": self.effort,
            "cases": results,
        }


def default_command_builder(*, harness: str, model: str, effort: str, prompt: str, skill: str | None = None, category: str | None = None, **kwargs: Any) -> list[str]:
    override_key = f"HIVE_EVAL_{harness.upper()}_COMMAND"
    override = os.environ.get(override_key)
    if override:
        values = shlex.split(override)
        if not values:
            raise EvalError(f"{override_key} is empty")
        return [item.replace("{model}", model).replace("{effort}", effort).replace("{prompt}", prompt) for item in values]
    if harness == "codex":
        effective_prompt = prompt
        if category == "explicit" and skill:
            skill_path = Path(kwargs["environment"]["HIVE_EVAL_SKILLS_DIR"]) / skill / "SKILL.md"
            effective_prompt += f"\n\nEval harness delivery: read and apply the skill at {skill_path}."
        return [
            "codex", "exec", "--ignore-user-config", "--ignore-rules", "--ephemeral",
            "--sandbox", "workspace-write", "--json", "--cd", str(kwargs.get("cwd", "<workspace>")),
            "--skip-git-repo-check", "--model", model, "--config", f"model_reasoning_effort={effort}", effective_prompt,
        ]
    if harness == "claude":
        claude_skills = (Path(kwargs["environment"]["HOME"]) / ".claude" / "skills").resolve()
        claude_skills_pattern = f"//{str(claude_skills).lstrip('/')}/**"
        return [
            "claude", "-p", "--output-format", "stream-json", "--verbose",
            "--no-session-persistence", "--permission-mode", "acceptEdits",
            "--permission-prompts", "none", "--setting-sources", "user,project",
            "--strict-mcp-config", "--mcp-config", '{"mcpServers":{}}',
            "--add-dir", str(claude_skills),
            "--disallowedTools",
            f"WebSearch,WebFetch,Task,Edit({claude_skills_pattern})",
            "--model", model, "--effort", effort, prompt,
        ]
    if harness == "pi":
        values = [
            "pi", "--no-extensions", "--no-prompt-templates", "--no-themes",
            "--no-session", "--mode", "json", "--no-approve", "-p",
            "--model", model, "--thinking", effort,
        ]
        if category == "explicit" and skill:
            values.extend(["--skill", str(Path(kwargs["environment"]["HIVE_EVAL_SKILLS_DIR"]) / skill / "SKILL.md")])
        values.append(prompt)
        return values
    if harness == "grok":
        cwd = Path(kwargs.get("cwd", "<workspace>"))
        return [
            "grok", "--cwd", str(cwd), "--leader-socket", str(cwd.parent / "grok-leader.sock"),
            "--model", model, "--reasoning-effort", effort, "--permission-mode", "dontAsk",
            "--allow", "Bash", "--allow", "Read", "--allow", "Grep", "--allow", "Glob",
            "--allow", "Edit", "--allow", "Write",
            "--output-format", "streaming-json", "--max-turns", "12", "--no-subagents",
            "--no-memory", "--no-auto-update", "--disable-web-search", "--single", prompt,
        ]
    if harness == "opencode":
        return [
            "opencode", "run", "--standalone", "--format", "json", "--model", model, prompt,
        ]
    raise EvalError(f"unsupported harness: {harness}")


def _parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Run the Hive skill evaluation pilot")
    parser.add_argument("--manifest", type=Path, default=DEFAULT_MANIFEST)
    parser.add_argument("--list", action="store_true", help="list case ids and exit")
    parser.add_argument("--dry-run", action="store_true", help="show the plan without executing (default)")
    parser.add_argument("--run", action="store_true", help="opt in to actual harness execution")
    parser.add_argument("--harness", choices=HARNESS_NAMES, default="codex")
    parser.add_argument("--source-root", type=Path, default=ROOT, help="repository snapshot whose cores and generated skills are staged")
    parser.add_argument("--arm", choices=("baseline", "candidate"), help="select the target skill named for this comparison arm")
    parser.add_argument("--model", default="gpt-5.6-luna")
    parser.add_argument("--effort", choices=EFFORTS, default="high")
    parser.add_argument("--case", action="append", dest="case_ids", default=[])
    parser.add_argument("--repeat", type=int, default=1)
    parser.add_argument("--output", type=Path, default=Path("-"))
    parser.add_argument("--timeout", type=float, default=120.0)
    return parser


def main(argv: list[str] | None = None) -> int:
    args = _parser().parse_args(argv)
    if args.run and args.dry_run:
        raise SystemExit("--run and --dry-run are mutually exclusive")
    if args.repeat < 1:
        raise SystemExit("--repeat must be at least 1")
    manifest = load_manifest(args.manifest, arm=args.arm)
    cases = manifest["cases"]
    if args.case_ids:
        wanted = set(args.case_ids)
        unknown = wanted - {case["id"] for case in cases}
        if unknown:
            raise SystemExit(f"unknown case id(s): {', '.join(sorted(unknown))}")
        cases = [case for case in cases if case["id"] in wanted]
    if args.list:
        for case in cases:
            print(f"{case['id']}\t{case['skill']}\t{case['category']}")
        return 0
    runner = EvalRunner(
        source_root=args.source_root,
        arm=args.arm,
        harness=args.harness,
        model=args.model,
        effort=args.effort,
        timeout=args.timeout,
    )
    if not args.run:
        report = runner.dry_run(cases, args.repeat)
    else:
        results = []
        retain_dir = None if str(args.output) == "-" else args.output.parent / f"{args.output.stem}-artifacts"
        for case in cases:
            for repeat_index in range(1, args.repeat + 1):
                result = runner.run_case(case, retain_dir=retain_dir, artifact_label=str(repeat_index))
                result["repeat_index"] = repeat_index
                results.append(result)
        report = {
            "version": 1,
            "mode": "run",
            "arm": args.arm,
            "source_root": str(args.source_root.resolve()),
            "harness": args.harness,
            "model": args.model,
            "effort": args.effort,
            "cases": results,
        }
    rendered = json.dumps(report, ensure_ascii=False, indent=2) + "\n"
    if str(args.output) == "-":
        sys.stdout.write(rendered)
    else:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(rendered, encoding="utf-8")
        os.chmod(args.output, 0o600)
    if args.run:
        return 0 if all(case["outcome"]["status"] == "passed" for case in report["cases"]) else 1
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except EvalError as exc:
        print(f"error: {exc}", file=sys.stderr)
        raise SystemExit(2)
