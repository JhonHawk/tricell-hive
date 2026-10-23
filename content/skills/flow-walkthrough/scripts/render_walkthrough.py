#!/usr/bin/env python3
"""Validate <folder>/walkthrough.json and write <folder>/walkthrough.html.

    python3 render_walkthrough.py <folder> [--session-status in_progress|closed]
    python3 render_walkthrough.py <folder> --session-status closed --confirm-viewer

The page is a fixed template (assets/walkthrough.html, .css, .mjs) with the
manifest embedded as an inert JSON data block. Python 3 standard library only.
Exit 0 and the written path on success; exit 1 on validation/publication error.
Exit 2 means closed was written but the viewer did not confirm its current
copy: retain the source evidence unit. No viewer is started or restarted.
"""

from __future__ import annotations

import argparse
import json
import math
import os
import re
import subprocess
import sys
import tempfile
import time
from datetime import datetime
from pathlib import Path

ASSETS = Path(__file__).resolve().parent.parent / "assets"

MAX_MANIFEST = 1024 * 1024
MAX_CAPTURE = 20 * 1024 * 1024
MAX_RECORDING = 200 * 1024 * 1024
MAX_FLOWS = 50
MAX_CAPTURES = 100

FLOW_ID = re.compile(r"^[a-z0-9-]{1,64}$")
SEGMENT = re.compile(r"^[A-Za-z0-9._-]+$")
LANG = re.compile(r"^[A-Za-z]{2,3}(-[A-Za-z0-9]{1,8})*$")
CAPTURE_EXTENSIONS = {".png", ".jpg", ".jpeg", ".webp"}
RECORDING_EXTENSIONS = {".mp4"}
CHECK_STATUSES = {"pass", "fail", "warn", "info", "not_run"}
VIEWPORTS = {"desktop", "mobile"}
KINDS = {"first_view", "final"}
SESSION_STATUSES = {"in_progress", "closed"}
CONFIRM_TIMEOUT = 15

RECORDING_FIELDS = {
    "none": ({"status"}, set()),
    "recorded": ({"status", "path", "duration_seconds", "candidate"}, {"path", "duration_seconds", "candidate"}),
    "not_recorded": ({"status", "reason"}, {"reason"}),
    "stale": ({"status", "path", "candidate"}, {"path", "candidate"}),
}


MAX_ERROR_TEXT = 200


def bounded(text: str) -> str:
    """Drop control characters and cut to MAX_ERROR_TEXT: an unknown field name
    or a path comes from an untrusted manifest and reaches stderr."""
    printable = "".join(ch for ch in text if ch.isprintable())
    if len(printable) > MAX_ERROR_TEXT:
        return printable[:MAX_ERROR_TEXT] + "\u2026"
    return printable


class ManifestError(Exception):
    def __init__(self, field: str, message: str) -> None:
        field, message = bounded(field), bounded(message)
        super().__init__(f"{field}: {message}")
        self.field = field



def is_one_of(value, allowed):
    """Return whether value is a string in allowed; other types never match."""
    return isinstance(value, str) and value in allowed

def is_finite(number: int | float) -> bool:
    """Return whether number is finite; an int too large for a float is not."""
    try:
        return math.isfinite(float(number))
    except OverflowError:
        return False


def fail(field: str, message: str) -> ManifestError:
    return ManifestError(field, message)


def expect_object(value: object, field: str, required: set[str], optional: set[str]) -> dict:
    if not isinstance(value, dict):
        raise fail(field, "must be an object")
    for key in value:
        if key not in required and key not in optional:
            name = f"{field}.{key}" if field else key
            raise fail(name, "unknown field")
    for key in sorted(required):
        if key not in value:
            name = f"{field}.{key}" if field else key
            raise fail(name, "required field is missing")
    return value


def expect_string(value: object, field: str, allow_empty: bool = False) -> str:
    if not isinstance(value, str):
        raise fail(field, "must be a string")
    if not allow_empty and not value.strip():
        raise fail(field, "must not be empty")
    return value


def expect_string_list(value: object, field: str) -> list[str]:
    if not isinstance(value, list):
        raise fail(field, "must be an array of strings")
    for index, item in enumerate(value):
        expect_string(item, f"{field}[{index}]")
    return value


def check_path(folder: Path, value: object, field: str, extensions: set[str], limit: int) -> None:
    text = expect_string(value, field)
    if text.startswith(("/", "\\")) or re.match(r"^[A-Za-z]:", text):
        raise fail(field, "path must be relative to the folder")
    segments = text.split("/")
    for segment in segments:
        if segment in ("", ".", ".."):
            raise fail(field, f"path must not contain empty, '.' or '..' segments: {text!r}")
        if not SEGMENT.fullmatch(segment):
            raise fail(field, f"path segment {segment!r} must match {SEGMENT.pattern}")
    suffix = Path(segments[-1]).suffix.lower()
    if suffix not in extensions:
        allowed = ", ".join(sorted(extensions))
        raise fail(field, f"extension {suffix or '(none)'} is not allowed; use one of {allowed}")
    current = folder
    for segment in segments:
        current = current / segment
        if current.is_symlink():
            raise fail(field, f"path crosses a symbolic link: {text!r}")
    root = folder.resolve()
    resolved = current.resolve()
    if root != resolved and root not in resolved.parents:
        raise fail(field, f"path escapes the folder: {text!r}")
    if not current.is_file():
        raise fail(field, f"referenced file does not exist: {text!r}")
    if current.stat().st_size > limit:
        raise fail(field, f"referenced file exceeds {limit // (1024 * 1024)} MiB: {text!r}")


def validate_recording(folder: Path, value: object, field: str) -> None:
    if not isinstance(value, dict):
        raise fail(field, "must be an object")
    status = value.get("status")
    if not is_one_of(status, RECORDING_FIELDS):
        raise fail(f"{field}.status", "must be one of none, recorded, not_recorded, stale")
    allowed, required = RECORDING_FIELDS[status]
    for key in value:
        if key not in allowed:
            raise fail(f"{field}.{key}", f"unknown field for status {status}")
    for key in sorted(required):
        if key not in value:
            raise fail(f"{field}.{key}", f"required for status {status}")
    if "path" in value:
        check_path(folder, value["path"], f"{field}.path", RECORDING_EXTENSIONS, MAX_RECORDING)
    if "duration_seconds" in value:
        seconds = value["duration_seconds"]
        if isinstance(seconds, bool) or not isinstance(seconds, (int, float)) or not is_finite(seconds) or seconds < 0:
            raise fail(f"{field}.duration_seconds", "must be a finite non-negative number")
    for key in ("candidate", "reason"):
        if key in value:
            expect_string(value[key], f"{field}.{key}")


def validate_flow(folder: Path, flow: object, index: int, seen: set[str]) -> None:
    field = f"flows[{index}]"
    flow = expect_object(
        flow, field, {"id", "title", "captures"}, {"steps", "recording"}
    )
    flow_id = expect_string(flow["id"], f"{field}.id")
    if not FLOW_ID.fullmatch(flow_id):
        raise fail(f"{field}.id", f"must match {FLOW_ID.pattern}")
    if flow_id in seen:
        raise fail(f"{field}.id", f"duplicate flow id {flow_id!r}")
    seen.add(flow_id)
    expect_string(flow["title"], f"{field}.title")
    if "steps" in flow:
        expect_string_list(flow["steps"], f"{field}.steps")
    captures = flow["captures"]
    if not isinstance(captures, list):
        raise fail(f"{field}.captures", "must be an array")
    if len(captures) > MAX_CAPTURES:
        raise fail(f"{field}.captures", f"at most {MAX_CAPTURES} captures per flow")
    for number, capture in enumerate(captures):
        item = f"{field}.captures[{number}]"
        capture = expect_object(capture, item, {"path", "caption"}, {"alt", "viewport"})
        check_path(folder, capture["path"], f"{item}.path", CAPTURE_EXTENSIONS, MAX_CAPTURE)
        expect_string(capture["caption"], f"{item}.caption")
        if "alt" in capture:
            expect_string(capture["alt"], f"{item}.alt")
        if "viewport" in capture and not is_one_of(capture["viewport"], VIEWPORTS):
            raise fail(f"{item}.viewport", "must be desktop or mobile")
    if "recording" in flow:
        validate_recording(folder, flow["recording"], f"{field}.recording")


def validate(folder: Path, manifest: object) -> None:
    root = expect_object(
        manifest,
        "",
        {"schema_version", "lang", "kind", "title", "session", "candidate", "created_at", "flows"},
        {"ticket", "result", "checks", "action_required", "environment", "provenance"},
    )
    version = root["schema_version"]
    if isinstance(version, bool) or version != 1:
        raise fail("schema_version", "must be 1")
    lang = expect_string(root["lang"], "lang")
    if not LANG.fullmatch(lang):
        raise fail("lang", "must be a language tag such as es or en")
    if not is_one_of(root["kind"], KINDS):
        raise fail("kind", "must be first_view or final")
    expect_string(root["title"], "title")
    session = expect_object(root["session"], "session", {"id"}, {"host", "status"})
    expect_string(session["id"], "session.id")
    if "host" in session:
        expect_string(session["host"], "session.host")
    if "status" in session and not is_one_of(session["status"], SESSION_STATUSES):
        raise fail("session.status", "must be in_progress or closed")
    candidate = expect_object(root["candidate"], "candidate", {"branch", "commit"}, {"base"})
    for key, value in candidate.items():
        expect_string(value, f"candidate.{key}")
    created = expect_string(root["created_at"], "created_at")
    # datetime.fromisoformat rejects a trailing "Z" before Python 3.11.
    normalized = created[:-1] + "+00:00" if created[-1:] in ("Z", "z") else created
    try:
        datetime.fromisoformat(normalized)
    except ValueError:
        raise fail("created_at", "must be an ISO 8601 date-time") from None
    for key in ("ticket", "result"):
        if key in root:
            expect_string(root[key], key)
    if "checks" in root:
        if not isinstance(root["checks"], list):
            raise fail("checks", "must be an array")
        for index, check in enumerate(root["checks"]):
            field = f"checks[{index}]"
            check = expect_object(check, field, {"name", "status"}, {"evidence"})
            expect_string(check["name"], f"{field}.name")
            if not is_one_of(check["status"], CHECK_STATUSES):
                raise fail(f"{field}.status", "must be one of " + ", ".join(sorted(CHECK_STATUSES)))
            if "evidence" in check:
                expect_string(check["evidence"], f"{field}.evidence")
    for key in ("action_required", "provenance"):
        if key in root:
            expect_string_list(root[key], key)
    if "environment" in root:
        if not isinstance(root["environment"], list):
            raise fail("environment", "must be an array")
        for index, entry in enumerate(root["environment"]):
            field = f"environment[{index}]"
            entry = expect_object(entry, field, {"label", "value"}, set())
            expect_string(entry["label"], f"{field}.label")
            expect_string(entry["value"], f"{field}.value")
    flows = root["flows"]
    if not isinstance(flows, list) or not flows:
        raise fail("flows", "must be a non-empty array")
    if len(flows) > MAX_FLOWS:
        raise fail("flows", f"at most {MAX_FLOWS} flows")
    seen: set[str] = set()
    for index, flow in enumerate(flows):
        validate_flow(folder, flow, index, seen)


def reject_constant(name: str) -> None:
    raise ValueError(f"{name} is not a valid JSON number")


def read_manifest(folder: Path) -> bytes:
    path = folder / "walkthrough.json"
    if not path.exists() and not path.is_symlink():
        raise fail("walkthrough.json", "file not found in the folder")
    if path.is_symlink() or not path.is_file():
        raise fail("walkthrough.json", "must be a regular file, not a link")
    try:
        # O_NOFOLLOW closes the gap between the check above and the read.
        descriptor = os.open(path, os.O_RDONLY | getattr(os, "O_NOFOLLOW", 0))
    except OSError:
        raise fail("walkthrough.json", "must be a regular file, not a link") from None
    with os.fdopen(descriptor, "rb") as handle:
        if os.fstat(handle.fileno()).st_size > MAX_MANIFEST:
            raise fail("walkthrough.json", "exceeds 1 MiB")
        raw = handle.read(MAX_MANIFEST + 1)
    if len(raw) > MAX_MANIFEST:
        raise fail("walkthrough.json", "exceeds 1 MiB")
    return raw


def parse_manifest(raw: bytes) -> object:
    try:
        return json.loads(raw.decode("utf-8"), parse_constant=reject_constant)
    except (ValueError, UnicodeError) as exc:
        raise fail("walkthrough.json", f"is not valid JSON: {exc}") from None


def load_manifest(folder: Path) -> object:
    return parse_manifest(read_manifest(folder))


def inert_json(manifest: object) -> str:
    text = json.dumps(manifest, ensure_ascii=False, separators=(",", ":"), allow_nan=False)
    return (
        text.replace("&", "\\u0026")
        .replace("<", "\\u003c")
        .replace(">", "\\u003e")
        .replace(" ", "\\u2028")
        .replace(" ", "\\u2029")
    )


def escape_html(text: str) -> str:
    return text.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;").replace('"', "&quot;")


def build_page(manifest: dict) -> str:
    shell = (ASSETS / "walkthrough.html").read_text(encoding="utf-8")
    module = (ASSETS / "walkthrough.mjs").read_text(encoding="utf-8")
    if "</script" in module.lower():
        raise fail("walkthrough.mjs", "module source must not contain a closing script tag")
    values = {
        "CSS": (ASSETS / "walkthrough.css").read_text(encoding="utf-8"),
        "MODULE": module,
        "DATA": inert_json(manifest),
        "TITLE": escape_html(manifest["title"]),
        "LANG": escape_html(manifest["lang"]),
    }
    # One pass, so injected text is never scanned for markers again.
    return re.sub(r"__WT_([A-Z]+)__", lambda match: values[match.group(1)], shell)


def stage_atomic(path: Path, content: bytes) -> Path:
    # mkstemp creates a fresh file (O_EXCL, mode 0600) under an unpredictable
    # name, so a link planted at a fixed temporary name is never followed.
    descriptor, name = tempfile.mkstemp(dir=path.parent, prefix=".walkthrough-", suffix=".tmp")
    temporary = Path(name)
    try:
        with os.fdopen(descriptor, "wb") as handle:
            handle.write(content)
        mode = path.stat().st_mode & 0o777 if path.name == "walkthrough.json" else 0o644
        os.chmod(temporary, mode)
        return temporary
    except BaseException:
        temporary.unlink(missing_ok=True)
        raise


def write_atomic(path: Path, content: bytes) -> None:
    temporary = stage_atomic(path, content)
    try:
        os.replace(temporary, path)
    finally:
        temporary.unlink(missing_ok=True)


def render(folder: Path, session_status: str | None = None) -> Path:
    """Validate and prepare both outputs before publishing explicit status.

    Each replacement is atomic. If publishing the page fails, restore the
    original manifest rather than leaving it ahead of the page. This is not a
    cross-file crash transaction; a retry renders from the authoritative JSON.
    """
    if not folder.is_dir():
        raise fail("folder", "must be a directory")
    original = read_manifest(folder)
    manifest = parse_manifest(original)
    validate(folder, manifest)
    source = folder / "walkthrough.json"
    if session_status is not None:
        if not is_one_of(session_status, SESSION_STATUSES):
            raise fail("session.status", "must be in_progress or closed")
        manifest["session"]["status"] = session_status
    validate(folder, manifest)
    page = build_page(manifest)
    serialized = json.dumps(manifest, ensure_ascii=False, indent=2, allow_nan=False) + "\n"
    if len(serialized.encode("utf-8")) > MAX_MANIFEST and session_status is not None:
        raise fail("walkthrough.json", "updated manifest exceeds 1 MiB")
    target = folder / "walkthrough.html"
    changed = session_status is not None and parse_manifest(original) != manifest
    staged_source = None
    staged_page = None
    try:
        # Stage every write before replacing either public file.
        if changed:
            staged_source = stage_atomic(source, serialized.encode("utf-8"))
        staged_page = stage_atomic(target, page.encode("utf-8"))
        if read_manifest(folder) != original:
            raise fail("walkthrough.json", "changed during rendering; retry from the current source")
        if changed:
            os.replace(staged_source, source)
        try:
            os.replace(staged_page, target)
        except OSError:
            # Do not restore over another writer's update.
            if changed and read_manifest(folder) == serialized.encode("utf-8"):
                write_atomic(source, original)
            raise
    finally:
        for temporary in (staged_source, staged_page):
            if temporary is not None:
                temporary.unlink(missing_ok=True)
    return target


def confirm_closed(folder: Path) -> bool:
    """Ask the existing authenticated CLI for the exact current source copy.

    The CLI owns identity/hash checks before and after its API query. Never
    inspect its store or echo its output, which can include private fields in
    an older runtime's diagnostics. Elapsed time alone never proves ingestion.
    """
    deadline = time.monotonic() + CONFIRM_TIMEOUT
    while True:
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            return False
        try:
            result = subprocess.run(
                ["hive", "view", "status", "--dir", str(folder.resolve())],
                stdin=subprocess.DEVNULL, capture_output=True, text=True,
                timeout=remaining, check=False,
            )
        except (OSError, subprocess.TimeoutExpired, UnicodeError):
            return False
        if result.returncode != 0:
            return False
        try:
            status = json.loads(result.stdout)
        except (ValueError, UnicodeError):
            return False
        if not isinstance(status, dict) or "found" not in status or "session_status" not in status:
            return False
        if status["found"] is True and status["session_status"] == "closed":
            return True
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            return False
        time.sleep(min(0.25, remaining))


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("folder", type=Path)
    parser.add_argument("--session-status", choices=sorted(SESSION_STATUSES),
                        help="explicitly update only this work's lifecycle and render it")
    parser.add_argument("--confirm-viewer", action="store_true",
                        help="with closed, require the viewer's exact current copy before cleanup")
    args = parser.parse_args()
    if args.confirm_viewer and args.session_status != "closed":
        parser.error("--confirm-viewer requires --session-status closed")
    folder = args.folder
    try:
        target = render(folder, args.session_status)
    except ManifestError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1
    except OSError:
        print("error: walkthrough files could not be read or published", file=sys.stderr)
        return 1
    if args.confirm_viewer and not confirm_closed(folder):
        print("error: current closed viewer copy not confirmed; retain the source evidence unit", file=sys.stderr)
        return 2
    print(target)
    return 0


if __name__ == "__main__":
    sys.exit(main())
