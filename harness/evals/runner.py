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
HARNESS_NAMES = ("codex", "pi")
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
}
FAILURE_WORDS = ("failed", "failure", "error", "cancelled", "canceled")
READ_NAMES = {"read", "read_file", "readfile", "cat", "sed", "view"}


class EvalError(ValueError):
    """Raised when a manifest or runner option is unsafe or malformed."""


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


def _normalize_case(case: Any, index: int) -> dict[str, Any]:
    if not isinstance(case, dict):
        raise EvalError(f"case {index} must be an object")
    for key in ("id", "skill", "category", "prompt"):
        if not isinstance(case.get(key), str) or not case[key]:
            raise EvalError(f"case {index} requires a non-empty {key}")
    if not isinstance(case.get("expected_activation"), bool):
        raise EvalError(f"case {case['id']} expected_activation must be boolean")

    skill = _safe_relative(case["skill"], f"case {case['id']} skill")
    fields = {
        "fixtures": {},
        "protected_files": [],
        "required_reads": [],
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
        "allowed_writes": allowed,
        "required_files": required,
        "output_contains": outputs,
        "file_contains": normalized_contains,
        **({"timeout_seconds": timeout} if timeout is not None else {}),
    }


def load_manifest(path: Path) -> dict[str, Any]:
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
        normalized_case = _normalize_case(case, index)
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


def _normalise_trace(stdout: str, stderr: str, skill: str) -> tuple[list[dict[str, Any]], list[dict[str, Any]], list[str], bool, bool]:
    events: list[dict[str, Any]] = []
    tool_events: list[dict[str, Any]] = []
    skill_loads: list[str] = []
    terminal = False
    terminal_failure = False
    tool_state: dict[str, dict[str, Any]] = {}
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
                record = {"name": name, "status": tool_status, "paths": paths, "event_type": event_type}
                tool_events.append(record)
                if call_id:
                    tool_state[call_id] = {"name": name, "paths": paths}
                if tool_status == "completed" and name.lower() in READ_NAMES:
                    for candidate in paths:
                        portable = candidate.replace("\\", "/")
                        skill_loads.append(portable)
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
            values = _walk_values(event, {"text", "content", "message", "output", "delta"})
            visible.extend(values)
    return "\n".join(visible)


def _sandbox_argv(argv: list[str], run_root: Path, workspace: Path) -> list[str]:
    """Wrap PI in a macOS sandbox whose only writable area is this run root."""
    sandbox = shutil.which("sandbox-exec") or "/usr/bin/sandbox-exec"
    if not Path(sandbox).is_file():
        raise EvalError("PI runs require /usr/bin/sandbox-exec for write containment")
    profile = run_root / "pi-sandbox.sb"
    root = str(run_root.resolve()).replace("\\", "\\\\").replace('"', '\\"')
    workspace_path = str(workspace.resolve()).replace("\\", "\\\\").replace('"', '\\"')
    profile.write_text(
        "(version 1)\n"
        "(deny default)\n"
        "(allow process*)\n"
        "(allow sysctl-read)\n"
        "(allow file-read*)\n"
        f'(allow file-write* (subpath "{root}") (subpath "{workspace_path}"))\n'
        '(allow file-write* (literal "/dev/null"))\n'
        "(allow network-outbound)\n",
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
        self.harness = harness
        self.model = model
        self.effort = effort
        self.runner_digest = _sha256_file(Path(__file__))
        self.timeout = timeout
        self.command_builder = command_builder or default_command_builder
        self.stream_log = stream_log or self._stream_to_stderr
        self._version_cache: dict[str, str | None] = {}
        self._active_process: subprocess.Popen[bytes] | None = None

    @staticmethod
    def _stream_to_stderr(chunk: str) -> None:
        sys.stderr.write(chunk)
        sys.stderr.flush()

    def _stage_skills(self, run_root: Path) -> tuple[Path, Path, dict[str, Any]]:
        generated = self.repo_root / "harness" / "agents-skills"
        if not generated.is_dir():
            raise EvalError(f"generated skills directory is missing: {generated}")
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
        codex_home.mkdir()
        pi_home.mkdir()
        shutil.copytree(staged, codex_home / "skills")
        shutil.copytree(staged, pi_home / "skills")
        (codex_home / "config.toml").write_text("# Hive eval isolated config\n", encoding="utf-8")
        (pi_home / "settings.json").write_text("{}\n", encoding="utf-8")
        return staged, codex_home, {"codex_home": codex_home, "pi_home": pi_home}

    def _link_auth(self, env: dict[str, str], roots: dict[str, Path]) -> dict[str, Any]:
        if self.harness == "codex":
            old_root = Path(env.get("CODEX_HOME", "")) if env.get("CODEX_HOME") else Path.home() / ".codex"
            target_root = roots["codex_home"]
            candidates = (old_root / "auth.json", Path.home() / ".codex" / "auth.json")
        else:
            old_root = Path(env.get("PI_CODING_AGENT_DIR", "")) if env.get("PI_CODING_AGENT_DIR") else Path.home() / ".pi" / "agent"
            target_root = roots["pi_home"]
            candidates = (old_root / "auth.json", Path.home() / ".pi" / "agent" / "auth.json")
        for source in candidates:
            if source.is_file() or source.is_symlink():
                link = target_root / source.name
                if not link.exists() and not link.is_symlink():
                    link.symlink_to(source)
                return {"status": "symlinked"}
        return {"status": "not-found"}

    def _environment(self, run_root: Path, roots: dict[str, Path]) -> tuple[dict[str, str], dict[str, Any]]:
        env = os.environ.copy()
        auth = self._link_auth(env, roots)
        env["HOME"] = str(run_root / "home")
        Path(env["HOME"]).mkdir()
        shared_skills = Path(env["HOME"]) / ".agents" / "skills"
        shared_skills.parent.mkdir(parents=True, exist_ok=True)
        shared_skills.symlink_to((run_root / "staged-skills").resolve(), target_is_directory=True)
        env["CODEX_HOME"] = str(roots["codex_home"])
        env["PI_CODING_AGENT_DIR"] = str(roots["pi_home"])
        env["HIVE_EVAL_SKILLS_DIR"] = str(run_root / "staged-skills")
        env["HIVE_EVAL_ISOLATED"] = "1"
        env["GIT_CONFIG_NOSYSTEM"] = "1"
        env["GIT_CONFIG_GLOBAL"] = str(run_root / "gitconfig")
        (run_root / "gitconfig").write_text("", encoding="utf-8")
        for name in ("XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_CACHE_HOME"):
            value = run_root / name.lower().replace("_home", "")
            value.mkdir(parents=True, exist_ok=True)
            env[name] = str(value)
        for key in ("CLAUDE_PROJECT_DIR", "CLAUDE_CODE_ENTRYPOINT", "CODEX_SESSION_ID"):
            env.pop(key, None)
        return env, {
            "config": "isolated",
            "auth": auth,
            "skills": "isolated HOME/.agents/skills symlinked to staged skills",
            "memory": "disabled-by-isolation",
            "mcp": "disabled-by-isolation",
        }

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
        try:
            result = subprocess.run(
                [resolved, "--version"], cwd=cwd, env=env, capture_output=True, text=True, timeout=5, check=False
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
        trace, tool_events, skill_loads, terminal, terminal_failure = _normalise_trace(stdout, stderr, "")
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
        case = _normalize_case(raw_case, 0)
        temporary_kwargs = {"dir": parent_dir} if parent_dir is not None else {}
        with tempfile.TemporaryDirectory(prefix=f"hive-eval-{case['id']}-", **temporary_kwargs) as temporary:
            run_root = Path(temporary)
            workspace = run_root / "workspace"
            workspace.mkdir()
            staged, _, roots = self._stage_skills(run_root)
            env, isolation = self._environment(run_root, roots)
            for relative, content in case["fixtures"].items():
                target = workspace / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                if not _inside(target, workspace):
                    raise EvalError(f"fixture escapes workspace: {relative}")
                target.write_text(content, encoding="utf-8")
            self._init_workspace_repo(workspace, env)
            isolation["git_fixture"] = "committed-clean"
            before = _tree_snapshot(workspace)
            source_skill = self.repo_root / "global" / "skills" / case["skill"]
            generated_skill = self.repo_root / "harness" / "agents-skills" / case["skill"]
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
            if self.harness == "pi":
                argv = _sandbox_argv(argv, run_root, workspace)
            execution = self._run_process(argv, workspace, env, case.get("timeout_seconds"))
            events, tool_events, skill_loads, terminal, terminal_failure = _normalise_trace(
                execution["stdout"], execution["stderr"], case["skill"]
            )
            execution.update(
                {
                    "trace": events,
                    "tool_events": tool_events,
                    "skill_loads": skill_loads,
                    "terminal_event": terminal,
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
            write_status = _status("passed") if not unexpected else _status("failed", "unexpected workspace writes")
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
            config_payload = {
                "harness": self.harness,
                "model": self.model,
                "effort": self.effort,
                "argv": argv,
                "isolated": True,
                "config_files": {
                    "codex": _sha256_file(roots["codex_home"] / "config.toml"),
                    "pi": _sha256_file(roots["pi_home"] / "settings.json"),
                },
            }
            config_digest = hashlib.sha256(json.dumps(config_payload, sort_keys=True).encode("utf-8")).hexdigest()
            artifact_evidence: dict[str, Any] = {"status": "not-retained", "files": []}
            if retain_dir is not None:
                safe_case = "".join(character if character.isalnum() or character in "-_" else "_" for character in case["id"])
                artifact_root = retain_dir / safe_case / artifact_label
                retained: list[str] = []
                for relative, value in after.items():
                    if value["type"] != "file":
                        continue
                    if relative not in before or before[relative] != value:
                        source = workspace / relative
                        destination = artifact_root / relative
                        destination.parent.mkdir(parents=True, exist_ok=True)
                        shutil.copy2(source, destination)
                        os.chmod(destination, 0o600)
                        retained.append(relative)
                artifact_evidence = {"status": "retained", "root": str(artifact_root), "files": sorted(retained)}
            return {
                "id": case["id"],
                "skill": case["skill"],
                "category": case["category"],
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
                    "isolation": isolation,
                    "artifacts": artifact_evidence,
                },
                "hashes": {
                    "runner": self.runner_digest,
                    "source_skill": source_digest,
                    "generated_skill": generated_digest,
                    "staged_skills": _tree_digest(staged),
                    "config": config_digest,
                    "model": {"requested": self.model, "resolved": sorted(set(_walk_values(execution["trace"], {"model", "model_name", "model_id"}))) or None, "cli_version": cli_version},
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
                "isolation": "fresh temporary workspace; generated skills staged; auth symlinked if present",
                "run": False,
            })
        return {"version": 1, "mode": "dry-run", "harness": self.harness, "model": self.model, "effort": self.effort, "cases": results}


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
    values = [
            "pi", "--no-extensions", "--no-context-files", "--no-prompt-templates",
            "--no-themes", "--no-session", "--mode", "json", "--no-approve", "-p",
            "--model", model, "--thinking", effort,
        ]
    if category == "explicit" and skill:
        values.extend(["--skill", str(Path(kwargs["environment"]["HIVE_EVAL_SKILLS_DIR"]) / skill / "SKILL.md")])
    values.append(prompt)
    return values


def _parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Run the Hive skill evaluation pilot")
    parser.add_argument("--manifest", type=Path, default=DEFAULT_MANIFEST)
    parser.add_argument("--list", action="store_true", help="list case ids and exit")
    parser.add_argument("--dry-run", action="store_true", help="show the plan without executing (default)")
    parser.add_argument("--run", action="store_true", help="opt in to actual harness execution")
    parser.add_argument("--harness", choices=HARNESS_NAMES, default="codex")
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
    manifest = load_manifest(args.manifest)
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
    runner = EvalRunner(harness=args.harness, model=args.model, effort=args.effort, timeout=args.timeout)
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
        report = {"version": 1, "mode": "run", "harness": args.harness, "model": args.model, "effort": args.effort, "cases": results}
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
