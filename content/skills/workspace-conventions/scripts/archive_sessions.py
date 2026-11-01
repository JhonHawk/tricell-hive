#!/usr/bin/env python3
"""Archive explicitly closed support sessions; dry-run unless --apply."""
from __future__ import annotations

import argparse
import os
import re
from pathlib import Path
import shutil
import sys


# The first status field decides: `Status: closed`, `**Estado:** cerrado. ...`, or a flow-plan control-sheet row such as `| Estado | Completado · ... |`.
STATUS = re.compile(
    r"^[ \t]*(?:\|[ \t]*)?(?:\*\*)?(?:Status|Estado)(?:\*\*)?[ \t]*(?::(?:\*\*)?|\|)[ \t]*(?:\*\*)?(?P<value>[^\n]*)$",
    re.I | re.M,
)
# The closed word must end the value or be followed by a separator; `closed?`, `Closed-loop`, or `completado parcialmente` stay open.
CLOSED_VALUE = re.compile(r"(?:closed|completed|cancelled|cerrado|completado|cancelado)(?=[ \t]*(?:$|[·.;|]|\*\*))", re.I)
LINK = re.compile(r"(\]\()([^\s)#]+)(#[^)]*)?(\))")
SCHEME = re.compile(r"^[A-Za-z][A-Za-z0-9+.-]*:")


def closed_status(text: str) -> str | None:
    match = STATUS.search(text)
    if match and CLOSED_VALUE.match(match.group("value")):
        return match.group(0).strip()
    return None


def closure(session: Path) -> str | None:
    plans = list(dict.fromkeys(list(session.glob("*.plan.md")) + list(session.glob("*-plan.md"))))
    if plans:
        closed = []
        for record in plans:
            try:
                status = closed_status(record.read_text())
            except UnicodeDecodeError:
                return None
            if not status:
                return None
            closed.append(f"{record.relative_to(session)}: {status}")
        return closed[0]
    records = list(dict.fromkeys(list(session.glob("*.research.md")) + list(session.glob("*-research.md"))))
    for record in records:
        try:
            status = closed_status(record.read_text())
        except UnicodeDecodeError:
            continue
        if status:
            return f"{record.relative_to(session)}: {status}"
    return None


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--support-root", required=True)
    parser.add_argument("--session", action="append", required=True, help="Selected folder name under sessions/ (repeatable)")
    parser.add_argument("--apply", action="store_true")
    args = parser.parse_args(argv)
    support = Path(args.support_root).resolve()
    sessions = support / "sessions"
    archived = sessions / "archived"
    mappings: list[tuple[Path, Path, str]] = []
    blocked: list[dict[str, str]] = []
    if sessions.is_symlink() or archived.is_symlink():
        print("BLOCKED sessions: sessions or archive destination is a symlink")
        return 2
    seen: set[str] = set()
    for name in args.session:
        if name in seen:
            blocked.append({"session": name, "reason": "selected more than once"})
            continue
        seen.add(name)
        if Path(name).name != name or name in {".", "..", "archived"}:
            blocked.append({"session": name, "reason": "not a direct session folder"})
            continue
        source = sessions / name
        evidence = closure(source) if source.is_dir() and not source.is_symlink() else None
        if not evidence:
            blocked.append({"session": name, "reason": "missing explicit closed Status in plan or research record"})
            continue
        target = archived / name
        if target.exists() or target.is_symlink():
            blocked.append({"session": name, "reason": "archive destination exists"})
            continue
        mappings.append((source, target, evidence))
    if blocked:
        for item in blocked:
            print(f"BLOCKED {item['session']}: {item['reason']}")
        return 2
    for source, target, evidence in mappings:
        print(f"MOVE {source} -> {target} (closure: {evidence})")
    if not args.apply:
        print("DRY-RUN")
        return 0
    for markdown in support.rglob("*.md"):
        if markdown.is_symlink():
            print(f"BLOCKED {markdown}: markdown link is a symlink")
            return 2
    archived.mkdir(parents=True, exist_ok=True)
    moved_roots = {source.resolve(): target.resolve() for source, target, _ in mappings}
    edits: list[tuple[Path, Path, str]] = []
    for markdown in support.rglob("*.md"):
        original = markdown.read_text()
        future_file = markdown
        for source, target in moved_roots.items():
            try:
                future_file = target / markdown.relative_to(source)
                break
            except ValueError:
                pass

        def replace(match: re.Match[str]) -> str:
            href = match.group(2)
            if href.startswith("/") or SCHEME.match(href):
                return match.group(0)
            original_target = (markdown.parent / href).resolve()
            future_target = original_target
            for source, target in moved_roots.items():
                try:
                    future_target = target / original_target.relative_to(source)
                    break
                except ValueError:
                    pass
            if future_file == markdown and future_target == original_target:
                return match.group(0)
            relative = os.path.relpath(future_target, future_file.parent).replace(os.sep, "/")
            return match.group(1) + relative + (match.group(3) or "") + match.group(4)

        updated = LINK.sub(replace, original)
        if updated != original:
            edits.append((markdown, future_file, updated))
    for source, target, _ in mappings:
        shutil.move(str(source), str(target))
    for original, future_file, updated in edits:
        future_file.write_text(updated)
        print(f"LINKS {future_file}")
    print("APPLIED")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
