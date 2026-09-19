#!/usr/bin/env python3
"""Deploy the Hive PI layer without touching any other harness.

The shell deploy entry point intentionally delegates PI state handling here.  PI
has a different configuration root and JSON model from the other harnesses;
keeping the merge, manifest, and rollback logic in one stdlib-only module makes
the ``--only pi`` route independently testable and keeps the shell wrapper
macOS bash 3.2 compatible.
"""

from __future__ import annotations

import argparse
import copy
import hashlib
import json
import os
import re
import stat
import subprocess
import sys
import tempfile
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Iterable


PACKAGE_PINS = (
    "npm:pi-subagents@0.67.0",
    "npm:gentle-engram@0.1.13",
    "npm:pi-mcp-adapter@2.33.0",
    "npm:@juicesharp/rpiv-ask-user-question@2.9.0",
    "npm:pi-web-access@0.29.0",
)
MCP_ALLOWLIST_REL = "harness/pi/src/mcp-allowlist.json"
# scriptMode false hides the adapter's mcpScript tool, which no Hive guard covers
# in the parent session.
MANAGED_MCP_SETTINGS = {"scriptMode": False}
# Keyed by path plus mtime/size so an edited allowlist is never served stale.
_MANAGED_MCP_CACHE: dict[tuple[Path, int, int], dict[str, dict[str, Any]]] = {}


def _managed_mcp_servers(repo_root: Path) -> dict[str, dict[str, Any]]:
    """Build the managed ``mcp.json`` entries from the deployed allowlist.

    The Hive guard in ``src/hooks.ts`` and the adapter-side ``includeTools`` cut
    read the same file out of the repository being deployed, so they cannot drift.
    Loaded on demand: a subcommand that writes no config (``rollback``) never
    needs it.
    """

    path = (repo_root.resolve() / MCP_ALLOWLIST_REL)
    try:
        stats = path.stat()
        cache_key = (path, stats.st_mtime_ns, stats.st_size)
        cached = _MANAGED_MCP_CACHE.get(cache_key)
        if cached is not None:
            return cached
        allowlist = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, UnicodeError, json.JSONDecodeError) as error:
        raise DeployError(f"Cannot parse the PI MCP allowlist {path}: {error}") from error
    if not isinstance(allowlist, dict):
        raise DeployError(f"Expected a JSON object in the PI MCP allowlist {path}")
    servers: dict[str, dict[str, Any]] = {}
    for name, policy in allowlist.items():
        if not isinstance(policy, dict) or not isinstance(policy.get("url"), str) or not isinstance(policy.get("tools"), dict):
            raise DeployError(f'Malformed "{name}" entry in the PI MCP allowlist {path}')
        entry: dict[str, Any] = {
            "url": policy["url"],
            "protocolVersion": "auto",
            "directTools": False,
            "includeTools": list(policy["tools"]),
            "lifecycle": "lazy",
        }
        if "auth" in policy:
            entry["auth"] = policy["auth"]
        if "headers" in policy:
            entry["headers"] = copy.deepcopy(policy["headers"])
        servers[name] = entry
    _MANAGED_MCP_CACHE[cache_key] = servers
    return servers


ENV_PLACEHOLDER_PATTERN = re.compile(r"^\$\{([A-Z0-9_]+)\}$")


def _header_env_warnings(servers: dict[str, dict[str, Any]], relative: str) -> list[str]:
    """Report `${VAR}` header placeholders the current environment cannot fill."""

    warnings: list[str] = []
    for name, entry in servers.items():
        headers = entry.get("headers")
        if not isinstance(headers, dict):
            continue
        for value in headers.values():
            match = ENV_PLACEHOLDER_PATTERN.match(value) if isinstance(value, str) else None
            if match is None or match.group(1) in os.environ:
                continue
            warnings.append(
                f"{relative} {name}: {match.group(1)} is not set in this shell; "
                "the header will be empty until the environment that launches pi exports it"
            )
    return warnings


SHELL_PATH = "/opt/homebrew/bin/bash"  # bash 5: parity with CLAUDE_CODE_SHELL and opencode `shell`
WEB_SEARCH_FIELDS = {
    "provider": "openai",
    "openaiSearchProviders": ["openai-codex"],
    "workflow": "none",
}
PI_HOOK_RELATIVES = (
    "bash-policy/bash-policy.sh",
    "flow-context/flow-context.sh",
    "flow-session-context/flow-session-context.sh",
    "post-tool-hub/post-tool-hub.sh",
    "reviewer-guard/reviewer-guard.sh",
    "rule-delivery/rule-delivery.py",
    "session-hygiene-report/session-hygiene-report.sh",
)
# Generated, not authored under global/hooks — but rule-delivery reads it from
# beside itself and exits silently without it, so it ships in the same bundle.
PI_HOOK_DATA_FILES = (
    ("global/hooks/rule-delivery/rule-manifest.json", "harness/rule-manifest.json"),
)
PI_REQUIRED_FILES = (
    ("harness/AGENTS.md", "generated PI core"),
    ("harness/pi/src/index.ts", "PI runtime entrypoint"),
    (MCP_ALLOWLIST_REL, "PI MCP allowlist"),
    ("harness/pi/extensions/hive-hooks.ts", "general PI extension entrypoint"),
    ("harness/pi/extensions/hive/reviewer-guard.ts", "reviewer PI extension entrypoint"),
)
MANAGED_CONFIGS = (
    "settings.json",
    "mcp.json",
    "web-search.json",
    "extensions/subagent/config.json",
)
MANIFEST_NAME = ".hive-deploy-manifest.json"
BACKUP_DIR_NAME = ".hive-deploy-backups"
MANIFEST_VERSION = 1
SHARED_MANIFEST_NAME = ".hive-deploy-manifest.json"
SHARED_SCOPE = "shared-skills"
SHARED_SOURCE_ROOT = "harness/agents-skills"
SHARED_TARGET_ROOT = "skills"


class DeployError(RuntimeError):
    """A source, manifest, or configuration error that must stop the run."""


def _package_source(value: Any) -> str | None:
    if isinstance(value, str):
        return value
    if isinstance(value, dict) and isinstance(value.get("source"), str):
        return value["source"]
    return None


def _package_name(value: Any) -> str:
    """Return an npm package identity from a Pi package spec or object."""

    source = _package_source(value)
    if source is None:
        return ""
    spec = source[4:] if source.startswith("npm:") else source
    if spec.startswith("@"):
        separator = spec.find("@", 1)
        return spec if separator < 0 else spec[:separator]
    return spec.split("@", 1)[0]


PACKAGE_NAMES = frozenset(_package_name(pin) for pin in PACKAGE_PINS)
PATCH_PACKAGE = "pi-subagents"
PATCH_VERSION = "0.67.0"
PATCH_FILE_REL = "harness/pi/patches/pi-subagents-0.67.0.patch"
PATCH_METADATA_REL = "harness/pi/patches/pi-subagents-0.67.0.json"
PATCH_TARGET_PREFIX = f"npm/node_modules/{PATCH_PACKAGE}/"
PATCH_SHA_PATTERN = re.compile(r"^[0-9a-f]{64}$")
PI_REQUIRED_PATCH_FILES = (
    (PATCH_FILE_REL, "reviewed pi-subagents patch"),
    (PATCH_METADATA_REL, "reviewed pi-subagents patch metadata"),
)


def _sha256_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def _sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def _json_bytes(value: Any) -> bytes:
    return (json.dumps(value, ensure_ascii=False, indent=2, sort_keys=True) + "\n").encode("utf-8")


def _read_json(path: Path) -> dict[str, Any]:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, UnicodeError, json.JSONDecodeError) as error:
        raise DeployError(f"Cannot parse {path}: {error}") from error
    if not isinstance(value, dict):
        raise DeployError(f"Expected a JSON object in {path}")
    return value


def _patch_sha(value: Any, field: str) -> str:
    if not isinstance(value, str) or not PATCH_SHA_PATTERN.fullmatch(value):
        raise DeployError(f"Invalid {field} in {PATCH_METADATA_REL}")
    return value


def _patch_relative(value: Any) -> str:
    if not isinstance(value, str) or not value or "\\" in value or "\x00" in value:
        raise DeployError(f"Invalid patch target in {PATCH_METADATA_REL}")
    relative = Path(value)
    if relative.is_absolute() or any(part in {"", ".", ".."} for part in relative.parts):
        raise DeployError(f"Invalid patch target in {PATCH_METADATA_REL}: {value}")
    return relative.as_posix()


def _patch_record_hash(record: dict[str, Any], name: str, relative: str) -> str:
    value = record.get(name)
    if value is None:
        aliases = {"sha256Before": "beforeSha256", "sha256After": "afterSha256"}
        value = record.get(aliases.get(name, name))
    if value is None:
        raise DeployError(f"Missing {name} for patch target {relative}")
    return _patch_sha(value, f"{name} for {relative}")


def _patch_path_label(value: str) -> str:
    label = value.split("\t", 1)[0].strip()
    if label == "/dev/null":
        return label
    if label.startswith("a/") or label.startswith("b/"):
        return label[2:]
    return label


def _parse_hunk_header(line: str) -> tuple[int, int, int, int]:
    match = re.match(r"^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@", line)
    if match is None:
        raise DeployError(f"Malformed patch hunk header in {PATCH_FILE_REL}: {line.rstrip()}")
    old_start = int(match.group(1))
    old_count = int(match.group(2) or "1")
    new_start = int(match.group(3))
    new_count = int(match.group(4) or "1")
    return old_start, old_count, new_start, new_count


def _parse_unified_patch(patch: bytes) -> dict[str, list[tuple[int, int, list[str]]]]:
    try:
        lines = patch.decode("utf-8").splitlines(keepends=True)
    except UnicodeDecodeError as error:
        raise DeployError(f"Patch is not valid UTF-8: {PATCH_FILE_REL}") from error

    sections: dict[str, list[tuple[int, int, list[str]]]] = {}
    index = 0
    while index < len(lines):
        if not lines[index].startswith("--- "):
            index += 1
            continue
        old_label = _patch_path_label(lines[index][4:])
        index += 1
        if index >= len(lines) or not lines[index].startswith("+++ "):
            raise DeployError(f"Patch is missing its new-file header in {PATCH_FILE_REL}")
        new_label = _patch_path_label(lines[index][4:])
        index += 1
        if old_label == "/dev/null" or new_label == "/dev/null" or old_label != new_label:
            raise DeployError(f"Patch must modify an existing file in place: {old_label} -> {new_label}")
        target = _patch_relative(new_label)
        hunks: list[tuple[int, int, list[str]]] = []
        while index < len(lines) and not lines[index].startswith("--- "):
            if not lines[index].startswith("@@ "):
                index += 1
                continue
            old_start, old_count, _new_start, _new_count = _parse_hunk_header(lines[index].rstrip("\r\n"))
            index += 1
            body: list[str] = []
            old_seen = 0
            new_seen = 0
            while index < len(lines) and not lines[index].startswith("@@ ") and not lines[index].startswith("--- "):
                line = lines[index]
                if line.startswith("\\ No newline at end of file"):
                    body.append(line)
                    index += 1
                    continue
                if not line or line[0] not in {" ", "+", "-"}:
                    raise DeployError(f"Malformed patch hunk body in {PATCH_FILE_REL}")
                body.append(line)
                if line[0] != "+":
                    old_seen += 1
                if line[0] != "-":
                    new_seen += 1
                index += 1
            if old_seen != old_count or new_seen != _new_count:
                raise DeployError(f"Patch hunk line counts do not match in {PATCH_FILE_REL}")
            hunks.append((old_start, old_count, body))
        if not hunks:
            raise DeployError(f"Patch has no hunks for {target}")
        if target in sections:
            raise DeployError(f"Patch contains duplicate target: {target}")
        sections[target] = hunks
    if not sections:
        raise DeployError(f"Patch contains no file sections: {PATCH_FILE_REL}")
    return sections


def _remove_line_ending(line: str) -> str:
    if line.endswith("\r\n"):
        return line[:-2]
    if line.endswith(("\n", "\r")):
        return line[:-1]
    return line


def _apply_unified_hunks(original: bytes, hunks: list[tuple[int, int, list[str]]]) -> bytes:
    try:
        source_lines = original.decode("utf-8").splitlines(keepends=True)
    except UnicodeDecodeError as error:
        raise DeployError("Patch target is not valid UTF-8") from error

    output: list[str] = []
    cursor = 0
    for old_start, _old_count, body in hunks:
        old_index = max(old_start - 1, 0)
        if old_index < cursor or old_index > len(source_lines):
            raise DeployError("Patch hunk does not match its target context")
        output.extend(source_lines[cursor:old_index])
        cursor = old_index
        last_output_kind: str | None = None
        for line in body:
            if line.startswith("\\ No newline at end of file"):
                if last_output_kind == "add" and output:
                    output[-1] = _remove_line_ending(output[-1])
                continue
            marker = line[0]
            content = line[1:]
            if marker in {" ", "-"}:
                if cursor >= len(source_lines) or source_lines[cursor] != content:
                    raise DeployError("Patch hunk does not match its target context")
                if marker == " ":
                    output.append(source_lines[cursor])
                    last_output_kind = "context"
                else:
                    last_output_kind = "remove"
                cursor += 1
            else:
                output.append(content)
                last_output_kind = "add"
    output.extend(source_lines[cursor:])
    return "".join(output).encode("utf-8")


def _patch_bundle(repo_root: Path) -> tuple[Path, dict[str, Any], dict[str, list[tuple[int, int, list[str]]]]]:
    patch_path = repo_root / PATCH_FILE_REL
    metadata_path = repo_root / PATCH_METADATA_REL
    if patch_path.is_symlink() or metadata_path.is_symlink() or not patch_path.is_file() or not metadata_path.is_file():
        raise DeployError(f"Missing required PI package patch bundle: {patch_path} and {metadata_path}")
    metadata = _read_json(metadata_path)
    if metadata.get("schemaVersion", 1) != 1:
        raise DeployError(f"Unsupported patch metadata schema: {metadata_path}")
    if metadata.get("package") != PATCH_PACKAGE or metadata.get("version") != PATCH_VERSION:
        raise DeployError(f"Patch metadata must target {PATCH_PACKAGE}@{PATCH_VERSION}: {metadata_path}")
    if metadata.get("patchFile") != Path(PATCH_FILE_REL).name or metadata.get("patchFormat") != "unified-diff":
        raise DeployError(f"Patch metadata does not describe the reviewed unified diff: {metadata_path}")
    patch_bytes = patch_path.read_bytes()
    declared_patch_sha = metadata.get("patchSha256")
    if declared_patch_sha is not None and _patch_sha(declared_patch_sha, "patchSha256") != _sha256_bytes(patch_bytes):
        raise DeployError(f"Patch bytes do not match patchSha256: {patch_path}")
    raw_files = metadata.get("files")
    if raw_files is None:
        raw_targets = metadata.get("targets")
        if not isinstance(raw_targets, list) or not raw_targets:
            raise DeployError(f"Patch metadata targets must be a non-empty array: {metadata_path}")
        raw_files = {}
        for raw_target in raw_targets:
            if not isinstance(raw_target, dict) or not isinstance(raw_target.get("path"), str):
                raise DeployError(f"Malformed patch target metadata: {metadata_path}")
            path = raw_target["path"]
            if path in raw_files:
                raise DeployError(f"Duplicate patch target metadata: {path}")
            raw_files[path] = raw_target
    if not isinstance(raw_files, dict) or not raw_files:
        raise DeployError(f"Patch metadata files must be a non-empty object: {metadata_path}")
    files: dict[str, dict[str, Any]] = {}
    for raw_relative, raw_record in raw_files.items():
        relative = _patch_relative(raw_relative)
        if not isinstance(raw_record, dict):
            raise DeployError(f"Malformed patch metadata for {relative}")
        files[relative] = {
            "sha256Before": _patch_record_hash(raw_record, "sha256Before", relative),
            "sha256After": _patch_record_hash(raw_record, "sha256After", relative),
        }
    sections = _parse_unified_patch(patch_bytes)
    if set(sections) != set(files):
        raise DeployError(f"Patch target list does not match metadata: {metadata_path}")
    metadata["files"] = files
    return patch_path, metadata, sections


def _package_root(pi_dir: Path) -> Path:
    npm_root = pi_dir / "npm"
    if npm_root.is_symlink() and not npm_root.resolve().is_dir():
        raise DeployError(f"PI npm staging link is not a directory: {npm_root}")
    if not npm_root.is_dir():
        raise DeployError(
            f"PI prerequisite missing: install {PATCH_PACKAGE}@{PATCH_VERSION} under {npm_root} before apply"
        )
    package_root = _installed_package_root(pi_dir, PATCH_PACKAGE)
    package_json = package_root / "package.json"
    package = _read_json(package_json)
    if package.get("name") != PATCH_PACKAGE or package.get("version") != PATCH_VERSION:
        raise DeployError(f"Installed package must be {PATCH_PACKAGE}@{PATCH_VERSION}: {package_json}")
    return package_root


def _installed_package_root(pi_dir: Path, package_name: str) -> Path:
    npm_root = pi_dir / "npm"
    package_root = npm_root / "node_modules" / package_name
    if package_root.is_symlink() or not package_root.is_dir():
        raise DeployError(f"PI prerequisite missing: install {package_name} under {npm_root}")
    try:
        package_root.resolve().relative_to(npm_root.resolve())
    except ValueError as error:
        raise DeployError(f"Refusing package link outside PI npm staging root: {package_root}") from error
    package_json = package_root / "package.json"
    if package_json.is_symlink() or not package_json.is_file():
        raise DeployError(f"PI package manifest is not a regular file: {package_json}")
    return package_root


def _validate_shell_path() -> None:
    """Refuse to point Pi's bash tool at a shell that is not there."""

    shell = Path(SHELL_PATH)
    if not (shell.is_file() and os.access(shell, os.X_OK)):
        raise DeployError(f"{SHELL_PATH} is missing or not executable; install Homebrew bash 5 before deploying shellPath")


def _validate_installed_packages(pi_dir: Path) -> None:
    """Require every pinned package before a selected PI deployment."""

    for pin in PACKAGE_PINS:
        source = _package_source(pin)
        if source is None:
            raise DeployError(f"Invalid managed package pin: {pin}")
        package_name = _package_name(pin)
        package_root = _installed_package_root(pi_dir, package_name)
        package_json = package_root / "package.json"
        package = _read_json(package_json)
        expected_version = source.rsplit("@", 1)[-1]
        if package.get("name") != package_name or package.get("version") != expected_version:
            raise DeployError(
                f"Installed package must be {package_name}@{expected_version}: {package_json}"
            )


def _validate_enabled_settings(pi_dir: Path) -> None:
    """Require all managed packages to be present and enabled in settings."""

    settings_path = pi_dir / "settings.json"
    if settings_path.is_symlink() or not settings_path.is_file():
        raise DeployError(f"PI settings prerequisite missing: {settings_path}")
    settings = _read_json(settings_path)
    packages = settings.get("packages")
    if not isinstance(packages, list):
        raise DeployError(f"{settings_path}.packages must list all managed PI packages")
    by_name: dict[str, list[Any]] = {}
    for item in packages:
        name = _package_name(item)
        if name in PACKAGE_NAMES:
            by_name.setdefault(name, []).append(item)
    for pin in PACKAGE_PINS:
        name = _package_name(pin)
        values = by_name.get(name, [])
        if len(values) != 1 or _package_source(values[0]) != pin:
            raise DeployError(f"PI settings do not enable the exact managed pin: {pin}")
        if isinstance(values[0], dict) and values[0].get("autoload") is False:
            raise DeployError(f"PI settings disable the managed package: {pin}")


def _package_patch_records(
    repo_root: Path, pi_dir: Path
) -> tuple[dict[str, dict[str, Any]], dict[str, bytes]]:
    patch_path, metadata, sections = _patch_bundle(repo_root)
    package_root = _package_root(pi_dir)
    resolved_package_root = package_root.resolve()
    records: dict[str, dict[str, Any]] = {}
    patch_data: dict[str, bytes] = {}
    for relative, hashes in metadata["files"].items():
        package_relative = _patch_relative(relative)
        target_relative = f"{PATCH_TARGET_PREFIX}{package_relative}"
        target = _safe_target(pi_dir, target_relative)
        try:
            target.parent.resolve().relative_to(resolved_package_root)
        except (OSError, RuntimeError, ValueError) as error:
            raise DeployError(
                f"Refusing package patch target outside installed package root: {target}"
            ) from error
        if target.is_symlink() or not target.is_file():
            raise DeployError(f"Patch target is not a regular file: {target}")
        before = target.read_bytes()
        current_hash = _sha256_bytes(before)
        expected_before = hashes["sha256Before"]
        expected_after = hashes["sha256After"]
        if current_hash == expected_after:
            after = before
        elif current_hash == expected_before:
            after = _apply_unified_hunks(before, sections[package_relative])
            if _sha256_bytes(after) != expected_after:
                raise DeployError(f"Patch output hash does not match metadata: {target}")
        else:
            raise DeployError(
                f"Patch target is neither pristine nor already patched: {target}"
            )
        records[target_relative] = {
            "sha256": expected_after,
            "mode": _file_mode(target),
            "source": PATCH_METADATA_REL,
            "kind": "package-patch",
            "package": PATCH_PACKAGE,
            "version": PATCH_VERSION,
            "packageRelative": package_relative,
            "sha256Before": expected_before,
            "sha256After": expected_after,
        }
        patch_data[target_relative] = after
    return records, patch_data


def _safe_target(root: Path, relative: str) -> Path:
    """Resolve a manifest path and reject traversal or absolute paths."""

    relative_path = Path(relative)
    if relative_path.is_absolute():
        raise DeployError(f"Refusing absolute target path: {relative}")
    candidate = root / relative_path
    # Resolve only the parent.  Keeping the leaf unresolved lets callers detect
    # and preserve a user-created symlink instead of following it.
    resolved_parent = candidate.parent.resolve()
    try:
        resolved_parent.relative_to(root.resolve())
    except ValueError as error:
        if relative.startswith(PATCH_TARGET_PREFIX):
            try:
                resolved_parent.relative_to((root / "npm").resolve())
                return candidate
            except ValueError:
                pass
        raise DeployError(f"Refusing target outside PI directory: {relative}") from error
    return candidate


def _root_path(path: Path, label: str) -> Path:
    expanded = path.expanduser()
    if expanded.is_symlink():
        raise DeployError(f"{label} root is a symlink; refusing to follow it: {expanded}")
    return expanded.resolve()


def _relative_source(repo_root: Path, path: Path) -> str:
    try:
        return path.resolve().relative_to(repo_root.resolve()).as_posix()
    except ValueError:
        return path.as_posix()


def _file_mode(path: Path, default: int = 0o644) -> int:
    try:
        return stat.S_IMODE(path.stat().st_mode)
    except OSError:
        return default


def _deploy_mode(source: Path, relative: str) -> int:
    mode = _file_mode(source)
    # A hook the PI bridge cannot execute is skipped in silence, and a checkout
    # need not preserve the bit — so it is forced here, whatever the language.
    if relative.startswith("global/hooks/") and source.suffix in {".sh", ".py"}:
        return 0o755
    return mode


def _atomic_write(path: Path, data: bytes, mode: int) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, temporary = tempfile.mkstemp(prefix=f".{path.name}.hive-", dir=str(path.parent))
    temporary_path = Path(temporary)
    try:
        os.fchmod(fd, mode)
        with os.fdopen(fd, "wb") as stream:
            stream.write(data)
            stream.flush()
            os.fsync(stream.fileno())
        os.replace(temporary_path, path)
    finally:
        if temporary_path.exists():
            temporary_path.unlink()


def _copy_atomic(source: Path, target: Path) -> None:
    _atomic_write(target, source.read_bytes(), _file_mode(source))


def _matches_file_state(path: Path, expected_hash: str | None, expected_mode: int | None) -> bool:
    if expected_hash is None:
        return not path.exists() and not path.is_symlink()
    return (
        path.is_file()
        and not path.is_symlink()
        and _sha256_file(path) == expected_hash
        and _file_mode(path) == expected_mode
    )


@dataclass
class FileAction:
    relative: str
    action: str
    source: Path | None = None
    data: bytes | None = None
    mode: int | None = None


@dataclass
class ConfigAction:
    relative: str
    action: str
    data: dict[str, Any] | None = None


@dataclass
class DeployPlan:
    desired_files: dict[str, dict[str, Any]] = field(default_factory=dict)
    next_files: dict[str, dict[str, Any]] = field(default_factory=dict)
    next_owned: dict[str, dict[str, Any]] = field(default_factory=dict)
    file_actions: list[FileAction] = field(default_factory=list)
    config_actions: list[ConfigAction] = field(default_factory=list)
    conflicts: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)
    errors: list[str] = field(default_factory=list)


def _canonical_hook_sources(repo_root: Path) -> Iterable[tuple[str, Path]]:
    hooks_root = repo_root / "global" / "hooks"
    if not hooks_root.is_dir():
        raise DeployError(f"Missing canonical hooks directory: {hooks_root}")
    for relative in PI_HOOK_RELATIVES:
        source = hooks_root / relative
        if not source.is_file():
            raise DeployError(f"Missing required PI hook source: {source}")
        # The deployed runtime resolves this root from the PI agent directory.
        # Keep canonical scripts there so the parent and pi-subagents' explicit
        # extension paths resolve the same files.
        yield f"global/hooks/{relative}", source
    for target, source_relative in PI_HOOK_DATA_FILES:
        source = repo_root / source_relative
        # Absent only in a tree where the build has never run: the hook then
        # finds no manifest and stays silent, which is its designed failure.
        if source.is_file():
            yield target, source


def _source_files(repo_root: Path) -> list[tuple[str, Path]]:
    """Return the PI runtime/agents/entrypoints plus canonical hook scripts."""

    result: list[tuple[str, Path]] = []
    pi_root = repo_root / "harness" / "pi"
    for source_root, target_root in (
        (pi_root / "agents", "agents"),
        (pi_root / "src", "src"),
        (pi_root / "extensions", "extensions"),
    ):
        if not source_root.is_dir():
            raise DeployError(f"Missing required PI source directory: {source_root}")
        source_count = 0
        for source in sorted(source_root.rglob("*")):
            if not source.is_file() or "__pycache__" in source.parts:
                continue
            if source_root.name == "agents" and source.name == "README.md":
                continue
            source_count += 1
            result.append(((Path(target_root) / source.relative_to(source_root)).as_posix(), source))
        if source_root.name == "agents" and source_count == 0:
            raise DeployError(f"Missing generated PI agents: {source_root}")

    for relative, description in PI_REQUIRED_FILES:
        source = repo_root / relative
        if not source.is_file():
            raise DeployError(f"Missing {description}: {source}")
    for relative, description in PI_REQUIRED_PATCH_FILES:
        source = repo_root / relative
        if source.is_symlink() or not source.is_file():
            raise DeployError(f"Missing {description}: {source}")
    core = repo_root / "harness" / "AGENTS.md"
    result.insert(0, ("AGENTS.md", core))

    result.extend(_canonical_hook_sources(repo_root))
    return result


def _load_manifest(path: Path, scope: str = "pi") -> dict[str, Any]:
    if not path.exists():
        return {}
    value = _read_json(path)
    if value.get("schemaVersion") != MANIFEST_VERSION or value.get("scope") != scope:
        raise DeployError(f"Unsupported or foreign {scope} manifest: {path}")
    if not isinstance(value.get("managedFiles", {}), dict) or not isinstance(value.get("managedConfig", {}), dict):
        raise DeployError(f"Malformed PI manifest: {path}")
    return value


def _rendered_source(source: Path, pi_dir: Path) -> bytes:
    data = source.read_bytes()
    return data.replace(b"__HIVE_PI_ROOT__", str(pi_dir).encode("utf-8"))


def _source_records(
    repo_root: Path, pi_dir: Path
) -> tuple[dict[str, dict[str, Any]], dict[str, bytes]]:
    records: dict[str, dict[str, Any]] = {}
    for relative, source in _source_files(repo_root):
        if relative in records:
            raise DeployError(f"Two PI sources map to the same target: {relative}")
        records[relative] = {
            "sha256": _sha256_bytes(_rendered_source(source, pi_dir)),
            "mode": _deploy_mode(source, relative),
            "source": _relative_source(repo_root, source),
        }
    patch_records, patch_data = _package_patch_records(repo_root, pi_dir)
    for relative, record in patch_records.items():
        if relative in records:
            raise DeployError(f"Two PI sources map to the same target: {relative}")
        records[relative] = record
    return records, patch_data


def _shared_source_records(repo_root: Path) -> tuple[dict[str, dict[str, Any]], dict[str, bytes]]:
    """Describe generated universal skills without claiming foreign files."""

    source_root = repo_root / SHARED_SOURCE_ROOT
    if not source_root.is_dir():
        raise DeployError(f"Missing generated shared skills: {source_root}")
    records: dict[str, dict[str, Any]] = {}
    data_by_relative: dict[str, bytes] = {}
    for source in sorted(source_root.rglob("*")):
        if not source.is_file():
            continue
        if source.is_symlink():
            raise DeployError(f"Shared skill source is a symlink: {source}")
        relative_source = source.relative_to(source_root).as_posix()
        relative = f"{SHARED_TARGET_ROOT}/{relative_source}"
        if relative in records:
            raise DeployError(f"Two shared skill sources map to the same target: {relative}")
        data = source.read_bytes()
        records[relative] = {
            "sha256": _sha256_bytes(data),
            "mode": _file_mode(source),
            "source": f"{SHARED_SOURCE_ROOT}/{relative_source}",
        }
        data_by_relative[relative] = data
    if not records:
        raise DeployError(f"Generated shared skills tree is empty: {source_root}")
    return records, data_by_relative


def _legacy_source_blob(repo_root: Path, commit: str, relative_source: str) -> bytes | None:
    """Read one exact generated blob from the legacy manifest's source commit."""

    completed = subprocess.run(
        ["git", "-C", str(repo_root), "show", f"{commit}:{relative_source}"],
        check=False,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    return completed.stdout if completed.returncode == 0 else None


def _legacy_shared_prior(
    repo_root: Path,
    shared_root: Path,
    legacy_manifest: Path,
) -> dict[str, Any]:
    """Adopt only legacy shared files whose bytes are provably generated."""

    if legacy_manifest.is_symlink():
        raise DeployError(f"Legacy shared manifest is a symlink; refusing to follow it: {legacy_manifest}")
    if not legacy_manifest.exists():
        return {"schemaVersion": MANIFEST_VERSION, "scope": SHARED_SCOPE, "managedFiles": {}, "managedConfig": {}}
    if not legacy_manifest.is_file():
        raise DeployError(f"Legacy shared manifest is not a regular file: {legacy_manifest}")
    try:
        lines = legacy_manifest.read_text(encoding="utf-8").splitlines()
    except (OSError, UnicodeError) as error:
        raise DeployError(f"Cannot read legacy shared manifest: {legacy_manifest}: {error}") from error
    commit = ""
    for line in lines:
        if line.startswith("# source_commit:"):
            commit = line.split(":", 1)[1].strip()
            break
    if commit and not re.fullmatch(r"[0-9a-fA-F]{7,64}", commit):
        raise DeployError(f"Legacy shared manifest has an invalid source_commit: {commit}")

    records: dict[str, dict[str, Any]] = {}
    for line in lines:
        if not line.startswith("agents-skills/"):
            continue
        relative_source = line.removeprefix("agents-skills/")
        if not relative_source or "\\" in relative_source or any(
            part in {"", ".", ".."} for part in Path(relative_source).parts
        ):
            raise DeployError(f"Invalid legacy shared manifest entry: {line}")
        relative = f"{SHARED_TARGET_ROOT}/{relative_source}"
        target = _safe_target(shared_root, relative)
        if not target.exists() and not target.is_symlink():
            continue
        if target.is_symlink() or target.is_dir() or not target.is_file():
            raise DeployError(f"Legacy shared target is not a regular file: {target}")
        current_source = repo_root / SHARED_SOURCE_ROOT / relative_source
        candidates: list[bytes] = []
        if current_source.is_file() and not current_source.is_symlink():
            candidates.append(current_source.read_bytes())
        if commit:
            historical = _legacy_source_blob(repo_root, commit, f"{SHARED_SOURCE_ROOT}/{relative_source}")
            if historical is not None:
                candidates.append(historical)
        target_bytes = target.read_bytes()
        if not any(target_bytes == candidate for candidate in candidates):
            raise DeployError(
                f"Legacy shared target cannot be adopted safely; unknown edits: {target}"
            )
        records[relative] = {
            "sha256": _sha256_bytes(target_bytes),
            "mode": _file_mode(target),
            "source": f"{SHARED_SOURCE_ROOT}/{relative_source}",
        }
    return {
        "schemaVersion": MANIFEST_VERSION,
        "scope": SHARED_SCOPE,
        "managedFiles": records,
        "managedConfig": {},
    }


def _shared_prior(
    repo_root: Path,
    shared_root: Path,
    manifest_path: Path,
    legacy_manifest: Path,
) -> dict[str, Any]:
    if manifest_path.is_symlink():
        raise DeployError(f"Shared manifest is a symlink; refusing to follow or replace it: {manifest_path}")
    if manifest_path.exists():
        return _load_manifest(manifest_path, scope=SHARED_SCOPE)
    return _legacy_shared_prior(repo_root, shared_root, legacy_manifest)


def _plan_file_records(
    repo_root: Path,
    target_root: Path,
    desired: dict[str, dict[str, Any]],
    data_by_relative: dict[str, bytes],
    prior: dict[str, Any],
    plan: DeployPlan,
    *,
    strict_conflicts: bool = False,
) -> None:
    """Plan managed files for one owner, without mutating its target root."""

    previous = prior.get("managedFiles", {}) if prior else {}
    if not isinstance(previous, dict):
        raise DeployError("Malformed managedFiles in manifest")
    plan.desired_files = desired
    plan.next_files = dict(desired)

    def conflict(message: str, relative: str, previous_record: Any = None) -> None:
        plan.file_actions.append(FileAction(relative, "conflict"))
        plan.conflicts.append(message)
        if strict_conflicts:
            plan.errors.append(message)
        if isinstance(previous_record, dict):
            plan.next_files[relative] = previous_record
        else:
            plan.next_files.pop(relative, None)

    def write_action(relative: str, record: dict[str, Any]) -> None:
        data = data_by_relative.get(relative)
        if data is None:
            raise DeployError(f"Missing source data for managed file: {relative}")
        plan.file_actions.append(
            FileAction(
                relative,
                "write",
                source=None if record.get("kind") == "package-patch" else _source_path(repo_root, record),
                data=data,
                mode=record["mode"],
            )
        )

    for relative, record in desired.items():
        target = _safe_target(target_root, relative)
        previous_record = previous.get(relative)
        if record.get("kind") == "package-patch":
            if target.is_symlink() or target.is_dir() or not target.is_file():
                plan.errors.append(f"Package patch target is not a regular file: {target}")
                continue
            current_hash = _sha256_file(target)
            if current_hash == record["sha256"]:
                if _file_mode(target) != record["mode"]:
                    plan.errors.append(f"Package patch target mode changed: {target}")
                continue
            if current_hash != record["sha256Before"]:
                plan.errors.append(f"Package patch target changed from its reviewed preimage: {target}")
                continue
            write_action(relative, record)
            continue
        if target.is_symlink():
            conflict(f"{relative}: target is a symlink; preserved", relative, previous_record)
            continue
        if target.exists() and target.is_dir():
            plan.errors.append(f"Target is a directory, expected a file: {target}")
            continue
        if not target.exists():
            write_action(relative, record)
            continue

        current_hash = _sha256_file(target)
        if current_hash == record["sha256"]:
            current_mode = _file_mode(target)
            if current_mode == record["mode"]:
                continue
            previous_mode = previous_record.get("mode") if isinstance(previous_record, dict) else None
            if previous_mode is not None and current_mode != previous_mode:
                conflict(
                    f"{relative}: target mode changed outside the last deploy; preserved",
                    relative,
                    previous_record,
                )
                continue
            write_action(relative, record)
            continue
        if isinstance(previous_record, dict) and current_hash == previous_record.get("sha256"):
            write_action(relative, record)
            continue
        conflict(f"{relative}: target changed outside the last deploy; preserved", relative, previous_record)

    for relative, previous_record in previous.items():
        if relative in desired:
            continue
        if not isinstance(previous_record, dict):
            plan.errors.append(f"Malformed managed file entry: {relative}")
            continue
        target = _safe_target(target_root, relative)
        if not target.exists() and not target.is_symlink():
            continue
        if target.is_symlink() or target.is_dir():
            message = f"{relative}: orphan target is not a regular file; preserved"
            plan.conflicts.append(message)
            plan.next_files[relative] = previous_record
            if strict_conflicts:
                plan.errors.append(message)
            continue
        if _sha256_file(target) == previous_record.get("sha256"):
            plan.file_actions.append(FileAction(relative, "delete"))
            plan.next_files.pop(relative, None)
        else:
            message = f"{relative}: orphan was modified after deployment; preserved"
            plan.file_actions.append(FileAction(relative, "conflict"))
            plan.conflicts.append(message)
            plan.next_files[relative] = previous_record
            if strict_conflicts:
                plan.errors.append(message)


def _plan_files(repo_root: Path, pi_dir: Path, prior: dict[str, Any], plan: DeployPlan) -> None:
    desired, patch_data = _source_records(repo_root, pi_dir)
    data_by_relative = dict(patch_data)
    for relative, record in desired.items():
        if record.get("kind") != "package-patch":
            source = _source_path(repo_root, record)
            data_by_relative[relative] = _rendered_source(source, pi_dir)
    _plan_file_records(repo_root, pi_dir, desired, data_by_relative, prior, plan)


def _source_path(repo_root: Path, record: dict[str, Any]) -> Path:
    source = record.get("source")
    if not isinstance(source, str):
        raise DeployError("Managed file record is missing its source")
    return repo_root / source


def _owned_value(previous: dict[str, Any], relative: str, key: str) -> Any:
    value = previous.get(relative, {}) if isinstance(previous, dict) else {}
    return value.get(key) if isinstance(value, dict) else None


def _merge_scalar(
    data: dict[str, Any],
    key: str,
    desired: Any,
    previous: dict[str, Any],
    relative: str,
    conflicts: list[str],
) -> None:
    if key not in data:
        owned = previous.get(relative, {}) if isinstance(previous, dict) else {}
        if isinstance(owned, dict) and key in owned:
            conflicts.append(f"{relative}.{key}: user removed the managed value; preserved")
            return
        data[key] = copy.deepcopy(desired)
        return
    before = _owned_value(previous, relative, key)
    if before is not None and data[key] != before and data[key] != desired:
        conflicts.append(f"{relative}.{key}: user value differs from the last managed value; preserved")
        return
    data[key] = copy.deepcopy(desired)


def _merge_packages(data: dict[str, Any], previous: dict[str, Any], conflicts: list[str], relative: str) -> None:
    previous_packages = _owned_value(previous, relative, "packages")
    if "packages" not in data:
        if isinstance(previous_packages, list):
            conflicts.append(f"{relative}.packages: user removed the managed package list; preserved")
            return
        data["packages"] = list(PACKAGE_PINS)
        return
    current = data["packages"]
    if not isinstance(current, list) or any(_package_source(item) is None for item in current):
        raise DeployError(
            f"{relative}.packages must be an array of strings or objects with a string source"
        )

    previous_by_name = {
        _package_name(item): _package_source(item)
        for item in previous_packages
        if _package_source(item) is not None
    } if isinstance(previous_packages, list) else {}
    current_by_name: dict[str, list[Any]] = {}
    for item in current:
        name = _package_name(item)
        if name in PACKAGE_NAMES:
            current_by_name.setdefault(name, []).append(item)

    accepted: dict[str, str] = {}
    for pin in PACKAGE_PINS:
        name = _package_name(pin)
        values = current_by_name.get(name, [])
        old = previous_by_name.get(name)
        if len(values) > 1:
            conflicts.append(f"{relative}.packages.{name}: duplicate entries; preserved")
            continue
        if not values and old is not None:
            conflicts.append(f"{relative}.packages.{name}: user removed the managed package; preserved")
            continue
        current_source = _package_source(values[0]) if values else None
        if values and old is not None and current_source != old and current_source != pin:
            conflicts.append(f"{relative}.packages.{name}: user version differs from the last managed pin; preserved")
            continue
        accepted[name] = pin

    merged: list[str] = []
    seen: set[str] = set()
    for item in current:
        name = _package_name(item)
        if name in PACKAGE_NAMES:
            if name in accepted and name not in seen:
                if isinstance(item, dict):
                    preserved = copy.deepcopy(item)
                    preserved["source"] = accepted[name]
                    merged.append(preserved)
                else:
                    merged.append(accepted[name])
                seen.add(name)
            elif name not in accepted:
                merged.append(item)
            continue
        merged.append(item)
    for pin in PACKAGE_PINS:
        name = _package_name(pin)
        if name not in seen and name in accepted:
            merged.append(pin)
            seen.add(name)
    data["packages"] = merged


def _merge_mcp_settings(
    data: dict[str, Any],
    previous_settings: Any,
    conflicts: list[str],
    relative: str,
) -> None:
    if "settings" not in data:
        if isinstance(previous_settings, dict):
            conflicts.append(f"{relative}.settings: user removed the managed settings map; preserved")
            return
        data["settings"] = copy.deepcopy(MANAGED_MCP_SETTINGS)
        return
    settings = data["settings"]
    if not isinstance(settings, dict):
        conflicts.append(f"{relative}.settings: user value is not an object; preserved")
        return
    for key, desired in MANAGED_MCP_SETTINGS.items():
        before = previous_settings.get(key) if isinstance(previous_settings, dict) else None
        if key not in settings:
            if before is not None:
                conflicts.append(f"{relative}.settings.{key}: user removed the managed value; preserved")
                continue
            settings[key] = copy.deepcopy(desired)
            continue
        if before is not None and settings[key] != before and settings[key] != desired:
            conflicts.append(
                f"{relative}.settings.{key}: user value differs from the last managed value; preserved"
            )
            continue
        settings[key] = copy.deepcopy(desired)


def _merge_mcp(
    data: dict[str, Any],
    previous: dict[str, Any],
    conflicts: list[str],
    relative: str,
    managed_servers: dict[str, dict[str, Any]],
) -> None:
    _merge_mcp_settings(data, _owned_value(previous, relative, "settings"), conflicts, relative)
    previous_servers = _owned_value(previous, relative, "mcpServers")
    if "mcpServers" not in data:
        if isinstance(previous_servers, dict):
            conflicts.append(f"{relative}.mcpServers: user removed the managed server map; preserved")
            return
        servers = {}
        data["mcpServers"] = servers
    else:
        servers = data["mcpServers"]
    if not isinstance(servers, dict):
        raise DeployError(f"{relative}.mcpServers must be an object")
    for name, desired_server in managed_servers.items():
        current = servers.get(name)
        previous_server = previous_servers.get(name) if isinstance(previous_servers, dict) else None
        if name not in servers:
            if isinstance(previous_server, dict):
                conflicts.append(f"{relative}.mcpServers.{name}: user removed the managed server; preserved")
                continue
            servers[name] = copy.deepcopy(desired_server)
            continue
        if not isinstance(current, dict):
            conflicts.append(f"{relative}.mcpServers.{name}: user value is not an object; preserved")
            continue
        for key, desired in desired_server.items():
            before = previous_server.get(key) if isinstance(previous_server, dict) else None
            if before is not None and current.get(key) != before and current.get(key) != desired:
                conflicts.append(
                    f"{relative}.mcpServers.{name}.{key}: user value differs from the last managed value; preserved"
                )
                continue
            current[key] = copy.deepcopy(desired)


def _config_desired(
    relative: str,
    data: dict[str, Any],
    previous: dict[str, Any],
    conflicts: list[str],
    managed_servers: dict[str, dict[str, Any]],
) -> dict[str, Any]:
    merged = copy.deepcopy(data)
    if relative == "settings.json":
        _merge_packages(merged, previous, conflicts, relative)
        _merge_scalar(merged, "shellPath", SHELL_PATH, previous, relative, conflicts)
    elif relative == "mcp.json":
        _merge_mcp(merged, previous, conflicts, relative, managed_servers)
    elif relative == "web-search.json":
        for key, desired in WEB_SEARCH_FIELDS.items():
            _merge_scalar(merged, key, desired, previous, relative, conflicts)
    elif relative == "extensions/subagent/config.json":
        _merge_scalar(merged, "forceTopLevelAsync", True, previous, relative, conflicts)
    else:
        raise DeployError(f"Unknown PI managed config: {relative}")
    return merged


def _plan_configs(repo_root: Path, pi_dir: Path, prior: dict[str, Any], plan: DeployPlan) -> None:
    managed_servers = _managed_mcp_servers(repo_root)
    previous_owned = prior.get("managedConfig", {}) if prior else {}
    if not isinstance(previous_owned, dict):
        raise DeployError("Malformed managedConfig in PI manifest")
    for relative in MANAGED_CONFIGS:
        target = _safe_target(pi_dir, relative)
        if target.is_symlink():
            plan.config_actions.append(ConfigAction(relative, "conflict"))
            plan.conflicts.append(f"{relative}: config target is a symlink; preserved")
            previous_record = previous_owned.get(relative)
            if isinstance(previous_record, dict):
                plan.next_owned[relative] = previous_record
            continue
        if target.exists() and target.is_dir():
            plan.errors.append(f"Config target is a directory: {target}")
            continue
        if target.exists():
            current = _read_json(target)
        else:
            current = {}
        conflicts_before = len(plan.conflicts)
        merged = _config_desired(relative, current, previous_owned, plan.conflicts, managed_servers)
        if relative == "settings.json":
            owned = {"packages": list(PACKAGE_PINS)}
        elif relative == "mcp.json":
            owned = {
                "mcpServers": copy.deepcopy(managed_servers),
                "settings": copy.deepcopy(MANAGED_MCP_SETTINGS),
            }
            plan.warnings.extend(_header_env_warnings(managed_servers, relative))
        elif relative == "web-search.json":
            owned = dict(WEB_SEARCH_FIELDS)
        else:
            owned = {"forceTopLevelAsync": True}
        plan.next_owned[relative] = owned
        if merged != current:
            plan.config_actions.append(ConfigAction(relative, "write", data=merged))
        elif len(plan.conflicts) == conflicts_before:
            plan.config_actions.append(ConfigAction(relative, "unchanged", data=merged))


def _manifest_for(plan: DeployPlan, scope: str = "pi") -> dict[str, Any]:
    return {
        "schemaVersion": MANIFEST_VERSION,
        "scope": scope,
        "managedFiles": {key: plan.next_files[key] for key in sorted(plan.next_files)},
        "managedConfig": {key: plan.next_owned[key] for key in sorted(plan.next_owned)},
    }


def _backup_stamp(root: Path) -> Path:
    timestamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    candidate = root / timestamp
    suffix = 1
    while candidate.exists():
        candidate = root / f"{timestamp}-{suffix}"
        suffix += 1
    return candidate


def _ensure_private_dir(path: Path) -> None:
    if path.is_symlink():
        raise DeployError(f"Backup directory is a symlink; refusing to follow it: {path}")
    path.mkdir(mode=0o700, parents=True, exist_ok=True)
    if not path.is_dir():
        raise DeployError(f"Backup path is not a directory: {path}")
    path.chmod(0o700)


def _create_backup(
    root: Path,
    manifest_path: Path,
    relatives: Iterable[str],
    *,
    manifest_name: str = MANIFEST_NAME,
    scope: str = "pi",
) -> Path:
    backup_root = root / BACKUP_DIR_NAME
    _ensure_private_dir(backup_root)
    backup_dir = _backup_stamp(backup_root)
    _ensure_private_dir(backup_dir)
    payload = backup_dir / "payload"
    _ensure_private_dir(payload)
    entries: dict[str, dict[str, Any]] = {}
    unique_relatives = sorted(set(relatives) | {manifest_name})
    for relative in unique_relatives:
        target = manifest_path if relative == manifest_name else _safe_target(root, relative)
        existed = target.is_file()
        before_hash = _sha256_file(target) if existed else None
        before_mode = _file_mode(target) if existed else None
        entry: dict[str, Any] = {
            "existed": existed,
            "applied": False,
            "sha256After": before_hash,
            "modeAfter": before_mode,
        }
        if existed:
            destination = payload / relative
            _ensure_private_dir(destination.parent)
            _atomic_write(destination, target.read_bytes(), 0o600)
            entry["sha256Before"] = before_hash
            entry["mode"] = before_mode
        entries[relative] = entry
    metadata = {
        "schemaVersion": 1,
        "scope": scope,
        "manifest": manifest_name,
        "root": str(root.resolve()),
        "files": entries,
    }
    _atomic_write(backup_dir / "metadata.json", _json_bytes(metadata), 0o600)
    return backup_dir


def _journal_pending(backup_dir: Path, relative: str, after_hash: str | None, after_mode: int | None) -> None:
    metadata_path = backup_dir / "metadata.json"
    metadata = _read_json(metadata_path)
    entries = metadata.get("files", {})
    entry = entries.get(relative)
    if not isinstance(entry, dict):
        raise DeployError(f"Missing backup journal entry: {relative}")
    entry["pendingSha256After"] = after_hash
    entry["pendingModeAfter"] = after_mode
    _atomic_write(metadata_path, _json_bytes(metadata), 0o600)


def _journal_applied(backup_dir: Path, relative: str, after_hash: str | None, after_mode: int | None) -> None:
    metadata_path = backup_dir / "metadata.json"
    metadata = _read_json(metadata_path)
    entries = metadata.get("files", {})
    entry = entries.get(relative)
    if not isinstance(entry, dict):
        raise DeployError(f"Missing backup journal entry: {relative}")
    entry["sha256After"] = after_hash
    entry["modeAfter"] = after_mode
    entry["applied"] = True
    entry.pop("pendingSha256After", None)
    entry.pop("pendingModeAfter", None)
    _atomic_write(metadata_path, _json_bytes(metadata), 0o600)


def _assert_before_state(backup_dir: Path, relative: str, target: Path) -> None:
    metadata = _read_json(backup_dir / "metadata.json")
    entries = metadata.get("files", {})
    entry = entries.get(relative)
    if not isinstance(entry, dict):
        raise DeployError(f"Missing backup journal entry: {relative}")
    if entry.get("existed"):
        if not _matches_file_state(target, entry.get("sha256Before"), entry.get("mode")):
            raise DeployError(f"Target changed after backup; refusing to overwrite: {target}")
    elif target.exists() or target.is_symlink():
        raise DeployError(f"Target appeared after backup; refusing to overwrite: {target}")


def _apply_plan(
    root: Path,
    manifest_path: Path,
    prior: dict[str, Any],
    plan: DeployPlan,
    manifest: dict[str, Any],
    *,
    manifest_name: str = MANIFEST_NAME,
    scope: str = "pi",
) -> Path | None:
    changed_files = [action.relative for action in plan.file_actions if action.action in {"write", "delete"}]
    changed_configs = [action.relative for action in plan.config_actions if action.action == "write"]
    manifest_bytes = _json_bytes(manifest)
    manifest_changed = not manifest_path.exists() or manifest_path.read_bytes() != manifest_bytes
    if not changed_files and not changed_configs and not manifest_changed:
        return None

    backup_relatives = changed_files + changed_configs
    if manifest_name == MANIFEST_NAME and scope == "pi":
        # Preserve the small positional seam used by callers that inject a
        # backup race in tests and by older local wrappers.
        backup_dir = _create_backup(root, manifest_path, backup_relatives)
    else:
        backup_dir = _create_backup(
            root,
            manifest_path,
            backup_relatives,
            manifest_name=manifest_name,
            scope=scope,
        )
    try:
        for action in plan.file_actions:
            if action.action == "write":
                if action.source is None and action.data is None:
                    raise DeployError(f"Missing source data for {action.relative}")
                target = _safe_target(root, action.relative)
                if target.is_symlink() or target.is_dir():
                    raise DeployError(f"Target changed before apply: {target}")
                _assert_before_state(backup_dir, action.relative, target)
                data = action.data if action.data is not None else action.source.read_bytes()
                mode = action.mode if action.mode is not None else _deploy_mode(action.source, action.relative)
                after_hash = _sha256_bytes(data)
                _journal_pending(backup_dir, action.relative, after_hash, mode)
                _assert_before_state(backup_dir, action.relative, target)
                _atomic_write(target, data, mode)
                _journal_applied(backup_dir, action.relative, after_hash, mode)
            elif action.action == "delete":
                target = _safe_target(root, action.relative)
                if target.is_symlink() or target.is_dir():
                    raise DeployError(f"Target changed before apply: {target}")
                _assert_before_state(backup_dir, action.relative, target)
                _journal_pending(backup_dir, action.relative, None, None)
                _assert_before_state(backup_dir, action.relative, target)
                if target.exists():
                    target.unlink()
                _journal_applied(backup_dir, action.relative, None, None)
        for action in plan.config_actions:
            if action.action == "write" and action.data is not None:
                target = _safe_target(root, action.relative)
                if target.is_symlink() or target.is_dir():
                    raise DeployError(f"Config target changed before apply: {target}")
                _assert_before_state(backup_dir, action.relative, target)
                data = _json_bytes(action.data)
                mode = _file_mode(target, 0o600) if target.exists() else 0o600
                after_hash = _sha256_bytes(data)
                _journal_pending(backup_dir, action.relative, after_hash, mode)
                _assert_before_state(backup_dir, action.relative, target)
                _atomic_write(target, data, mode)
                _journal_applied(backup_dir, action.relative, after_hash, mode)
        if manifest_changed:
            if manifest_path.is_symlink() or manifest_path.is_dir():
                raise DeployError(f"{scope} manifest changed before apply: {manifest_path}")
            _assert_before_state(backup_dir, manifest_name, manifest_path)
            _journal_pending(backup_dir, manifest_name, _sha256_bytes(manifest_bytes), 0o600)
            _assert_before_state(backup_dir, manifest_name, manifest_path)
            _atomic_write(manifest_path, manifest_bytes, 0o600)
            _journal_applied(backup_dir, manifest_name, _sha256_bytes(manifest_bytes), 0o600)
    except (DeployError, OSError) as error:
        failure_label = "PI" if scope == "pi" else scope
        raise DeployError(f"{failure_label} apply failed after backup {backup_dir}: {error}") from error
    return backup_dir


def _print_plan(
    plan: DeployPlan,
    root: Path,
    apply: bool,
    backup_dir: Path | None = None,
    *,
    label: str = "PI deploy",
) -> None:
    mode = "APPLY" if apply else "DRY-RUN"
    print(f"{label} — {mode} — {root}")
    writes = sum(action.action == "write" for action in plan.file_actions)
    deletes = sum(action.action == "delete" for action in plan.file_actions)
    file_conflicts = sum(action.action == "conflict" for action in plan.file_actions)
    config_writes = sum(action.action == "write" for action in plan.config_actions)
    print(f"files: {writes} write, {deletes} delete, {file_conflicts} conflict")
    print(f"configs: {config_writes} write")
    if backup_dir is not None:
        print(f"backup: {backup_dir}")
    for error in plan.errors:
        print(f"ERROR: {error}")
    for warning in plan.warnings:
        print(f"WARNING: {warning}")
    for conflict in plan.conflicts:
        print(f"CONFLICT: {conflict}")
    if not apply:
        for action in plan.file_actions:
            if action.action in {"write", "delete"}:
                print(f"would {action.action} {action.relative}")
        for action in plan.config_actions:
            if action.action == "write":
                print(f"would write {action.relative}")


def deploy(repo_root: Path, pi_dir: Path, apply: bool) -> int:
    repo_root = repo_root.resolve()
    pi_dir = _root_path(pi_dir, "PI agent")
    manifest_path = pi_dir / MANIFEST_NAME
    if manifest_path.is_symlink():
        raise DeployError(f"PI manifest is a symlink; refusing to follow or replace it: {manifest_path}")
    prior = _load_manifest(manifest_path)
    plan = DeployPlan()
    _plan_files(repo_root, pi_dir, prior, plan)
    _plan_configs(repo_root, pi_dir, prior, plan)
    if plan.errors:
        _print_plan(plan, pi_dir, apply)
        return 2
    manifest = _manifest_for(plan)
    backup_dir: Path | None = None
    if apply:
        backup_dir = _apply_plan(pi_dir, manifest_path, prior, plan, manifest)
    _print_plan(plan, pi_dir, apply, backup_dir)
    return 0


def _pi_preflight(repo_root: Path, pi_dir: Path) -> int:
    """Validate the installed PI and Hive package surface without writing."""

    repo_root = repo_root.resolve()
    pi_dir = _root_path(pi_dir, "PI agent")
    _validate_installed_packages(pi_dir)
    _validate_enabled_settings(pi_dir)
    _validate_shell_path()
    manifest_path = pi_dir / MANIFEST_NAME
    if manifest_path.is_symlink():
        raise DeployError(f"PI manifest is a symlink; refusing to follow or replace it: {manifest_path}")
    prior = _load_manifest(manifest_path)
    plan = DeployPlan()
    _plan_files(repo_root, pi_dir, prior, plan)
    _plan_configs(repo_root, pi_dir, prior, plan)
    if plan.errors:
        _print_plan(plan, pi_dir, False, label="PI preflight")
        return 2
    print(f"PI preflight — OK — {pi_dir}")
    return 0


def deploy_shared(
    repo_root: Path,
    shared_root: Path,
    legacy_manifest: Path,
    apply: bool,
) -> int:
    """Deploy generated universal skills under their neutral owner."""

    repo_root = repo_root.resolve()
    shared_root = _root_path(shared_root, "Shared")
    manifest_path = shared_root / SHARED_MANIFEST_NAME
    prior = _shared_prior(repo_root, shared_root, manifest_path, legacy_manifest.expanduser())
    desired, data_by_relative = _shared_source_records(repo_root)
    plan = DeployPlan()
    _plan_file_records(
        repo_root,
        shared_root,
        desired,
        data_by_relative,
        prior,
        plan,
        strict_conflicts=True,
    )
    if plan.errors:
        _print_plan(plan, shared_root, apply, label="shared skills preflight")
        return 2
    manifest = _manifest_for(plan, scope=SHARED_SCOPE)
    backup_dir: Path | None = None
    if apply:
        backup_dir = _apply_plan(
            shared_root,
            manifest_path,
            prior,
            plan,
            manifest,
            manifest_name=SHARED_MANIFEST_NAME,
            scope=SHARED_SCOPE,
        )
    _print_plan(plan, shared_root, apply, backup_dir, label="shared skills deploy")
    return 0


def _rollback(
    backup_dir: Path,
    root: Path,
    apply: bool,
    *,
    manifest_name: str = MANIFEST_NAME,
    label: str = "PI rollback",
    scope: str = "pi",
) -> int:
    backup_dir = backup_dir.expanduser()
    if backup_dir.is_symlink():
        raise DeployError(f"Backup directory is a symlink; refusing to follow it: {backup_dir}")
    backup_dir = backup_dir.resolve()
    root = _root_path(root, "Rollback")
    metadata_path = backup_dir / "metadata.json"
    if metadata_path.is_symlink() or not metadata_path.is_file():
        raise DeployError(f"Backup metadata not found: {metadata_path}")
    metadata = _read_json(metadata_path)
    if metadata.get("scope") != scope:
        raise DeployError(f"Backup scope {metadata.get('scope')!r} does not match rollback scope {scope!r}")
    if metadata.get("manifest") != manifest_name:
        raise DeployError(f"Backup manifest identity does not match rollback target: {metadata_path}")
    if metadata.get("root") != str(root):
        raise DeployError(f"Backup root does not match rollback target: {root}")
    files = metadata.get("files")
    if not isinstance(files, dict):
        raise DeployError(f"Malformed backup metadata: {metadata_path}")

    conflicts: list[str] = []
    restorable: list[str] = []
    payload = backup_dir / "payload"
    if payload.is_symlink():
        raise DeployError(f"Backup payload is a symlink; refusing to follow it: {payload}")
    for relative, entry in files.items():
        if not isinstance(entry, dict):
            raise DeployError(f"Malformed backup entry: {relative}")
        target = root / manifest_name if relative == manifest_name else _safe_target(root, relative)
        if target.is_symlink() or target.is_dir():
            conflicts.append(f"{relative}: target is not a regular file; preserved")
            continue

        should_restore = False
        pending = "pendingSha256After" in entry or "pendingModeAfter" in entry
        if pending:
            # A pending journal entry means the mutation may not have reached
            # os.replace yet.  Accept either the complete after-state (the
            # replace happened before journaling failed) or the complete
            # before-state (the write failed before mutation).  Anything else
            # is an external edit and must block the all-or-nothing rollback.
            after_hash = entry.get("pendingSha256After")
            after_mode = entry.get("pendingModeAfter")
            before_hash = entry.get("sha256Before") if entry.get("existed") else None
            before_mode = entry.get("mode") if entry.get("existed") else None
            if _matches_file_state(target, after_hash, after_mode):
                should_restore = True
            elif _matches_file_state(target, before_hash, before_mode):
                continue
            else:
                conflicts.append(f"{relative}: changed after the backed-up deploy; preserved")
                continue
        elif entry.get("applied") or (
            "applied" not in entry and "sha256After" in entry
        ):
            expected = entry.get("sha256After")
            expected_mode = entry.get("modeAfter")
            if _matches_file_state(target, expected, expected_mode):
                should_restore = True
            elif not target.exists() and not target.is_symlink():
                # Preserve a user deletion after a successfully applied write.
                continue
            else:
                conflicts.append(f"{relative}: changed after the backed-up deploy; preserved")
                continue
        else:
            # No pending journal means this mutation was never started.  The
            # target must still be at its backed-up before-state for rollback
            # to proceed without treating it as a user change.
            before_hash = entry.get("sha256Before") if entry.get("existed") else None
            before_mode = entry.get("mode") if entry.get("existed") else None
            if _matches_file_state(target, before_hash, before_mode):
                continue
            if not target.exists() and not target.is_symlink():
                continue
            conflicts.append(f"{relative}: changed after the backed-up deploy; preserved")
            continue

        if should_restore and entry.get("existed"):
            source = _safe_target(payload, relative)
            if source.is_symlink() or not source.is_file():
                conflicts.append(f"{relative}: backup payload entry is not a regular file; preserved")
                continue
        if should_restore:
            restorable.append(relative)

    print(f"{label} — {'APPLY' if apply else 'DRY-RUN'} — {backup_dir}")
    for conflict in conflicts:
        print(f"CONFLICT: {conflict}")
    for relative in restorable:
        print(f"would restore {relative}" if not apply else f"restore {relative}")
    if conflicts:
        return 2
    if not apply:
        return 0

    for relative in restorable:
        target = root / manifest_name if relative == manifest_name else _safe_target(root, relative)
        entry = files[relative]
        if entry.get("existed"):
            source = _safe_target(payload, relative)
            _copy_atomic(source, target)
            if isinstance(entry.get("mode"), int):
                target.chmod(entry["mode"])
        elif target.exists() or target.is_symlink():
            target.unlink()
    return 0


def _parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Deploy or roll back the Hive PI layer")
    subparsers = parser.add_subparsers(dest="command")
    deploy_parser = subparsers.add_parser("deploy", help="deploy generated PI files and additive config")
    rollback_parser = subparsers.add_parser("rollback", help="restore one PI backup")
    for command_parser in (deploy_parser, rollback_parser):
        command_parser.add_argument("--pi-dir", default=os.environ.get("PI_CODING_AGENT_DIR", "~/.pi/agent"))
        mode = command_parser.add_mutually_exclusive_group()
        mode.add_argument("--apply", action="store_true", help="write changes")
        mode.add_argument("--dry-run", action="store_true", help="report changes without writing (default)")
        command_parser.add_argument("--verbose", action="store_true")
    deploy_parser.add_argument("--repo-root", default=str(Path(__file__).resolve().parents[2]))
    rollback_parser.add_argument("--backup-dir", required=True)

    shared_parser = subparsers.add_parser("shared", help="manage the neutral universal-skills owner")
    shared_subparsers = shared_parser.add_subparsers(dest="shared_command", required=True)
    shared_deploy = shared_subparsers.add_parser("deploy", help="deploy generated universal skills")
    shared_deploy.add_argument("--repo-root", default=str(Path(__file__).resolve().parents[2]))
    shared_deploy.add_argument("--shared-root", default="~/.agents")
    shared_deploy.add_argument("--legacy-manifest", default="~/.claude/.deploy-manifest")
    shared_deploy_mode = shared_deploy.add_mutually_exclusive_group()
    shared_deploy_mode.add_argument("--apply", action="store_true", help="write changes")
    shared_deploy_mode.add_argument("--dry-run", action="store_true", help="report changes without writing (default)")
    shared_deploy.add_argument("--verbose", action="store_true")

    shared_rollback = shared_subparsers.add_parser("rollback", help="restore one shared-skills backup")
    shared_rollback.add_argument("--shared-root", default="~/.agents")
    shared_rollback.add_argument("--backup-dir", required=True)
    shared_rollback_mode = shared_rollback.add_mutually_exclusive_group()
    shared_rollback_mode.add_argument("--apply", action="store_true", help="write changes")
    shared_rollback_mode.add_argument("--dry-run", action="store_true", help="report changes without writing (default)")
    shared_rollback.add_argument("--verbose", action="store_true")

    preflight_parser = subparsers.add_parser("preflight", help="validate installed PI prerequisites without writing")
    preflight_parser.add_argument("--repo-root", default=str(Path(__file__).resolve().parents[2]))
    preflight_parser.add_argument("--pi-dir", default=os.environ.get("PI_CODING_AGENT_DIR", "~/.pi/agent"))
    preflight_parser.add_argument("--verbose", action="store_true")
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = _parser()
    raw_args = list(sys.argv[1:] if argv is None else argv)
    if raw_args and raw_args[0] not in {"deploy", "rollback", "shared", "preflight", "-h", "--help"}:
        raw_args.insert(0, "deploy")
    args = parser.parse_args(raw_args)
    command = args.command or "deploy"
    apply = bool(getattr(args, "apply", False))
    try:
        if command == "rollback":
            return _rollback(
                Path(args.backup_dir),
                Path(args.pi_dir),
                apply,
            )
        if command == "preflight":
            return _pi_preflight(Path(args.repo_root), Path(args.pi_dir))
        if command == "shared":
            shared_root = Path(args.shared_root)
            if args.shared_command == "rollback":
                return _rollback(
                    Path(args.backup_dir),
                    shared_root,
                    apply,
                    manifest_name=SHARED_MANIFEST_NAME,
                    label="shared skills rollback",
                    scope=SHARED_SCOPE,
                )
            return deploy_shared(
                Path(args.repo_root),
                shared_root,
                Path(args.legacy_manifest),
                apply,
            )
        repo_root = Path(getattr(args, "repo_root", str(Path(__file__).resolve().parents[2])))
        return deploy(
            repo_root,
            Path(getattr(args, "pi_dir", os.environ.get("PI_CODING_AGENT_DIR", "~/.pi/agent"))),
            apply,
        )
    except DeployError as error:
        print(f"ERROR: {error}", file=sys.stderr)
        return 2
    except OSError as error:
        print(f"ERROR: filesystem operation failed: {error}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
