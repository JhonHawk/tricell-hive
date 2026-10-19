#!/usr/bin/env python3
"""Validate <folder>/walkthrough.json and write <folder>/walkthrough.html.

    python3 render_walkthrough.py <folder>

The page is a fixed template (assets/walkthrough.html, .css, .mjs) with the
manifest embedded as an inert JSON data block. Python 3 standard library only.
Exit 0 and one line on stdout (the written path) on success; exit 1 and an
error naming the field on stderr otherwise, and no page is written.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import sys
import tempfile
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
        if isinstance(seconds, bool) or not isinstance(seconds, (int, float)) or seconds < 0:
            raise fail(f"{field}.duration_seconds", "must be a non-negative number")
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
    session = expect_object(root["session"], "session", {"id"}, {"host"})
    expect_string(session["id"], "session.id")
    if "host" in session:
        expect_string(session["host"], "session.host")
    candidate = expect_object(root["candidate"], "candidate", {"branch", "commit"}, {"base"})
    for key, value in candidate.items():
        expect_string(value, f"candidate.{key}")
    created = expect_string(root["created_at"], "created_at")
    try:
        datetime.fromisoformat(created)
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


def load_manifest(folder: Path) -> object:
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
    try:
        return json.loads(raw.decode("utf-8"), parse_constant=reject_constant)
    except (ValueError, UnicodeError) as exc:
        raise fail("walkthrough.json", f"is not valid JSON: {exc}") from None


def inert_json(manifest: object) -> str:
    text = json.dumps(manifest, ensure_ascii=False, separators=(",", ":"))
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


def write_atomic(path: Path, content: str) -> None:
    # mkstemp creates a fresh file (O_EXCL, mode 0600) under an unpredictable
    # name, so a link planted at a fixed temporary name is never followed.
    descriptor, name = tempfile.mkstemp(dir=path.parent, prefix=".walkthrough-", suffix=".tmp")
    temporary = Path(name)
    try:
        with os.fdopen(descriptor, "w", encoding="utf-8") as handle:
            handle.write(content)
        os.chmod(temporary, 0o644)
        os.replace(temporary, path)
    finally:
        if temporary.exists():
            temporary.unlink()


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("folder", type=Path)
    args = parser.parse_args()
    folder = args.folder
    try:
        if not folder.is_dir():
            raise fail("folder", f"not a directory: {folder}")
        manifest = load_manifest(folder)
        validate(folder, manifest)
        page = build_page(manifest)
        target = folder / "walkthrough.html"
        write_atomic(target, page)
    except ManifestError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1
    except OSError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1
    print(target)
    return 0


if __name__ == "__main__":
    sys.exit(main())
