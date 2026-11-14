#!/usr/bin/env python3
"""Prepare explicit .engram identities without inferring a workspace."""
from __future__ import annotations

import argparse
import json
from pathlib import Path
import sys


def inside(child: Path, root: Path) -> bool:
    try:
        child.relative_to(root)
        return True
    except ValueError:
        return False


def parse_config(path: Path) -> str | None:
    try:
        value = json.loads(path.read_text())
    except (OSError, json.JSONDecodeError):
        return None
    name = value.get("project_name") if isinstance(value, dict) else None
    return name if isinstance(name, str) and name else None


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", required=True, help="Declared workspace or repository root")
    parser.add_argument("--name", required=True, help="Requested Engram project_name")
    parser.add_argument("--target", action="append", default=[], help="Root-relative target (repeatable; default is root)")
    parser.add_argument("--apply", action="store_true", help="Create missing matching configs")
    args = parser.parse_args(argv)

    root = Path(args.root).resolve()
    if not root.is_dir():
        print(json.dumps({"status": "blocked", "reason": "declared root does not exist"}))
        return 2
    if not args.name.strip() or "/" in args.name or "\\" in args.name:
        print(json.dumps({"status": "blocked", "reason": "invalid project_name"}))
        return 2

    requested = args.target or ["."]
    targets: list[Path] = []
    for raw in requested:
        candidate = (root / raw).resolve()
        if not candidate.is_dir() or not inside(candidate, root):
            print(json.dumps({"status": "blocked", "reason": "target is missing or outside declared root", "target": raw}))
            return 2
        if candidate not in targets:
            targets.append(candidate)

    records = []
    blocked = []
    for target in targets:
        config = target / ".engram" / "config.json"
        engram_dir = target / ".engram"
        if engram_dir.is_symlink() or (engram_dir.exists() and not engram_dir.is_dir()) or config.is_symlink():
            existing, state = None, "blocked"
            blocked.append(str(config))
        elif config.exists():
            existing = parse_config(config)
            state = "preserved" if existing == args.name else "blocked"
            if state == "blocked":
                blocked.append(str(config))
        else:
            existing, state = None, "create"
        records.append({"target": str(target), "config": str(config), "state": state, "existing": existing})

    output = {"root": str(root), "project_name": args.name, "apply": args.apply, "targets": records}
    if blocked:
        output["status"] = "blocked"
        output["reason"] = "existing configuration differs or is malformed; no targets changed"
        print(json.dumps(output, indent=2))
        return 2
    if not args.apply:
        output["status"] = "dry-run"
        print(json.dumps(output, indent=2))
        return 0
    for record in records:
        if record["state"] == "create":
            config = Path(record["config"])
            config.parent.mkdir(mode=0o700, exist_ok=True)
            config.write_text(json.dumps({"project_name": args.name}) + "\n")
            record["state"] = "created"
    output["status"] = "applied"
    print(json.dumps(output, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
