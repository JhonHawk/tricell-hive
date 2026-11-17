#!/usr/bin/env python3
"""Validate and mutate HIVE.local.md, ACCOUNTS.md, and _support/secrets/ deterministically.

Subcommands:
  check         Validate git-ignore exclusions, secret permissions (0700/0600),
                HIVE.local.md line length (<=200 chars), dates, <=50 line cap,
                and absence of secret values in HIVE.local.md and ACCOUNTS.md.
  ensure-ignore Ensure $(git rev-parse --git-common-dir)/info/exclude ignores
                HIVE.local.md, _support/secrets/, and _support/scripts/.
  add           Ensure ignore, validate a new entry line (<=200 chars, dated,
                secret-free), append it to main checkout's HIVE.local.md, and
                prune STALE entries when total entries exceed 50.
  mark-stale    Prefix a matching entry with STALE <YYYY-MM-DD> (<cause>): (or
                add failure date/cause to an unverified entry).
"""

import argparse
import importlib.util
import json
import os
import pathlib
import re
import shutil
import subprocess
import sys

SCRIPT_DIR = pathlib.Path(__file__).resolve().parent


def _load_extract_candidates():
    path = SCRIPT_DIR / "extract_candidates.py"
    spec = importlib.util.spec_from_file_location("extract_candidates", path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {path}")
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


EXTRACT = _load_extract_candidates()
SECRET_FLAG = EXTRACT.SECRET_FLAG
ENTROPY = EXTRACT.ENTROPY
MIXED = EXTRACT.MIXED
secret_like = EXTRACT.secret_like

DATE_RE = re.compile(r"\b\d{4}-\d{2}-\d{2}\b")
STALE_PREFIX_RE = re.compile(r"^STALE \d{4}-\d{2}-\d{2} \([^)]*\):\s*")
UNVERIFIED_FAIL_RE = re.compile(r"\bunverified \d{4}-\d{2}-\d{2} \([^)]*\):\s*")
KNOWN_SECRET_RE = re.compile(
    r"\b(?:gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,}|glpat-[A-Za-z0-9_-]{20,}|"
    r"(?:AKIA|ASIA)[0-9A-Z]{16}|AIza[0-9A-Za-z_-]{35}|xox[baprs]-[A-Za-z0-9-]{10,}|"
    r"sk-(?:ant-|proj-|live-)?[A-Za-z0-9_-]{20,})\b|-----BEGIN (?:[A-Z0-9 ]+ )?PRIVATE KEY-----"
)
REQUIRED_EXCLUDES = ("HIVE.local.md", "_support/secrets/", "_support/scripts/")
DEFAULT_HEADER = (
    "Per-machine facts of this repository, untracked, kept by agents under the global guidance's "
    "«Local instructions»; facts that hold in any project live in ~/.config/hive/HIVE.local.md.\n"
)


def run_git(cwd, *args):
    git = shutil.which("git") or "git"
    proc = subprocess.run(
        [git, "-C", str(cwd), *args],
        stdin=subprocess.DEVNULL,
        capture_output=True,
        text=True,
    )
    return proc.returncode, proc.stdout, proc.stderr.strip()


def resolve_main_checkout(repo):
    code, out, _ = run_git(repo, "worktree", "list", "--porcelain")
    if code == 0:
        for line in out.splitlines():
            if line.startswith("worktree "):
                return pathlib.Path(line[len("worktree ") :].strip())
    return pathlib.Path(repo).resolve()


def line_has_secret(text):
    if KNOWN_SECRET_RE.search(text):
        return True
    for raw_tok in re.split(r"[\s|,;]+", text):
        tok = raw_tok.strip("'\"`()[]{}")
        if not tok or tok.startswith(("http://", "https://", "engram:", "/", "~/", "./")):
            continue
        if "=" in tok:
            k, v = tok.split("=", 1)
            if SECRET_FLAG.search(k) and len(v) >= 8:
                return True
            if secret_like(v):
                return True
        elif len(tok) >= 20 and secret_like(tok):
            return True
    return False


def effective_entry_length(line):
    trimmed = line.strip()
    rest = trimmed[2:] if trimmed.startswith("- ") else trimmed
    rest = STALE_PREFIX_RE.sub("", rest)
    rest = UNVERIFIED_FAIL_RE.sub("unverified ", rest)
    return len("- " + rest)


def validate_entry_line(line):
    errors = []
    trimmed = line.strip()
    if not trimmed.startswith("- "):
        errors.append("entry must start with '- '")
    if "\n" in line.rstrip("\r\n"):
        errors.append("entry must be a single line")
    if effective_entry_length(trimmed) > 200:
        errors.append(f"entry exceeds 200 characters ({effective_entry_length(trimmed)} chars)")
    if not DATE_RE.search(trimmed):
        errors.append("entry is missing a YYYY-MM-DD verification date")
    if line_has_secret(trimmed):
        errors.append("entry contains a secret-like value")
    return errors


def ensure_ignore(repo):
    code, common_dir, err = run_git(repo, "rev-parse", "--git-common-dir")
    if code != 0 or not common_dir.strip():
        return {"ok": False, "error": err or "not a git repository"}
    cpath = pathlib.Path(common_dir.strip())
    if not cpath.is_absolute():
        cpath = (pathlib.Path(repo) / cpath).resolve()
    exclude_file = cpath / "info" / "exclude"
    exclude_file.parent.mkdir(parents=True, exist_ok=True)
    existing = exclude_file.read_text(encoding="utf-8") if exclude_file.is_file() else ""
    existing_lines = {line.strip() for line in existing.splitlines()}
    added = []
    for item in REQUIRED_EXCLUDES:
        check_target = item.rstrip("/")
        chk_code, _, _ = run_git(repo, "check-ignore", "-q", check_target)
        if chk_code != 0 and item not in existing_lines and check_target not in existing_lines:
            added.append(item)
    if added:
        prefix = "" if (not existing or existing.endswith("\n")) else "\n"
        with exclude_file.open("a", encoding="utf-8") as f:
            f.write(prefix + "\n".join(added) + "\n")
    return {"ok": True, "exclude_file": str(exclude_file), "added": added}


def check_project_local(repo):
    main_root = resolve_main_checkout(repo)
    blocking = []
    warnings = []

    for rel in ("HIVE.local.md", "_support/secrets", "_support/scripts"):
        target = main_root / rel
        if target.exists():
            code, _, _ = run_git(main_root, "check-ignore", "-q", rel)
            if code != 0:
                blocking.append(f"{rel}: not ignored by Git; add it to .git/info/exclude")

    secrets_dir = main_root / "_support" / "secrets"
    if secrets_dir.is_dir():
        perm = secrets_dir.stat().st_mode & 0o777
        if perm != 0o700:
            blocking.append(f"_support/secrets: permissions are {perm:04o}; expected 0700")
        for item in sorted(secrets_dir.iterdir()):
            if not item.is_file():
                continue
            if item.name != "ACCOUNTS.md":
                fperm = item.stat().st_mode & 0o777
                if fperm != 0o600:
                    blocking.append(f"_support/secrets/{item.name}: permissions are {fperm:04o}; expected 0600")
            else:
                for idx, line in enumerate(item.read_text(encoding="utf-8").splitlines(), 1):
                    if line_has_secret(line):
                        blocking.append(f"_support/secrets/ACCOUNTS.md: line {idx} contains a secret-like value")

    local_file = main_root / "HIVE.local.md"
    if local_file.is_file():
        total_entries = 0
        stale_entries = 0
        past_intro = False
        for idx, raw in enumerate(local_file.read_text(encoding="utf-8").splitlines(), 1):
            trimmed = raw.strip()
            if not trimmed or trimmed.startswith("#"):
                past_intro = True
                continue
            if not trimmed.startswith("- "):
                if line_has_secret(trimmed):
                    blocking.append(f"HIVE.local.md: line {idx} contains a secret-like value")
                if past_intro:
                    blocking.append(f"HIVE.local.md: line {idx} is not a '- ' entry")
                continue
            past_intro = True
            total_entries += 1
            if trimmed[2:].startswith("STALE "):
                stale_entries += 1
            for err in validate_entry_line(trimmed):
                blocking.append(f"HIVE.local.md: line {idx}: {err}")
        if total_entries > 50:
            if stale_entries > 0:
                blocking.append(f"HIVE.local.md: {total_entries} entries (> 50) with unpruned STALE entries")
            else:
                warnings.append(
                    f"HIVE.local.md: {total_entries} entries (> 50 with no STALE entries); curate with harness-audit"
                )

    return {"ok": len(blocking) == 0, "blocking": blocking, "warnings": warnings}


def add_entry(repo, entry, target_file=None):
    trimmed = entry.strip()
    if not trimmed.startswith("- "):
        trimmed = "- " + trimmed
    errors = validate_entry_line(trimmed)
    if errors:
        return {"ok": False, "errors": errors}

    if target_file:
        path = pathlib.Path(target_file).expanduser().resolve()
        path.parent.mkdir(parents=True, exist_ok=True)
        code, _, _ = run_git(path.parent, "rev-parse", "--is-inside-work-tree")
        if code == 0:
            ensure_ignore(str(path.parent))
    else:
        main_root = resolve_main_checkout(repo)
        ig = ensure_ignore(str(main_root))
        if not ig.get("ok"):
            return {"ok": False, "errors": [ig.get("error", "ensure_ignore failed")]}
        path = main_root / "HIVE.local.md"

    if path.is_file():
        lines = path.read_text(encoding="utf-8").splitlines()
    else:
        lines = [DEFAULT_HEADER.rstrip("\n"), ""]

    lines.append(trimmed)
    entry_indices = [i for i, l in enumerate(lines) if l.strip().startswith("- ")]
    stale_indices = [i for i in entry_indices if lines[i].strip()[2:].startswith("STALE ")]

    pruned = 0
    while len(entry_indices) - pruned > 50 and stale_indices:
        drop_idx = stale_indices.pop(0)
        lines[drop_idx] = None
        pruned += 1

    final_lines = [l for l in lines if l is not None]
    path.write_text("\n".join(final_lines) + "\n", encoding="utf-8")
    total_after = sum(1 for l in final_lines if l.strip().startswith("- "))
    return {
        "ok": True,
        "file": str(path),
        "pruned_stale": pruned,
        "total_entries": total_after,
        "needs_curation": total_after > 50,
    }


def mark_stale(repo, match_substring, date_str, cause, target_file=None):
    if not DATE_RE.fullmatch(date_str.strip()):
        return {"ok": False, "error": f"invalid YYYY-MM-DD date: {date_str!r}"}
    main_root = resolve_main_checkout(repo)
    path = pathlib.Path(target_file).expanduser().resolve() if target_file else (main_root / "HIVE.local.md")
    if not path.is_file():
        return {"ok": False, "error": f"{path} does not exist"}

    lines = path.read_text(encoding="utf-8").splitlines()
    updated = 0
    for i, line in enumerate(lines):
        trimmed = line.strip()
        if not trimmed.startswith("- ") or match_substring not in trimmed:
            continue
        rest = trimmed[2:]
        if rest.startswith("STALE "):
            continue
        if "unverified" in rest.lower():
            cleaned = UNVERIFIED_FAIL_RE.sub("", rest)
            cleaned = re.sub(r"\s*\bunverified\b", "", cleaned).strip()
            lines[i] = f"- unverified {date_str.strip()} ({cause.strip()}): {cleaned}"
        else:
            lines[i] = f"- STALE {date_str.strip()} ({cause.strip()}): {rest}"
        updated += 1

    if updated == 0:
        return {"ok": False, "error": f"no matching active entry for {match_substring!r}"}
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return {"ok": True, "file": str(path), "updated": updated}


def main(argv=None):
    parser = argparse.ArgumentParser(description="Validate and mutate HIVE.local.md and _support/secrets/.")
    sub = parser.add_subparsers(dest="cmd", required=True)

    p_check = sub.add_parser("check")
    p_check.add_argument("--repo", default=".")

    p_ig = sub.add_parser("ensure-ignore")
    p_ig.add_argument("--repo", default=".")

    p_add = sub.add_parser("add")
    p_add.add_argument("--repo", default=".")
    p_add.add_argument("--file", default=None)
    p_add.add_argument("entry")

    p_stale = sub.add_parser("mark-stale")
    p_stale.add_argument("--repo", default=".")
    p_stale.add_argument("--file", default=None)
    p_stale.add_argument("--date", required=True)
    p_stale.add_argument("--cause", required=True)
    p_stale.add_argument("match")

    args = parser.parse_args(argv)
    if args.cmd == "check":
        res = check_project_local(args.repo)
        print(json.dumps(res, indent=2))
        return 0 if res["ok"] else 1
    if args.cmd == "ensure-ignore":
        res = ensure_ignore(args.repo)
        print(json.dumps(res, indent=2))
        return 0 if res["ok"] else 1
    if args.cmd == "add":
        res = add_entry(args.repo, args.entry, target_file=args.file)
        print(json.dumps(res, indent=2))
        return 0 if res["ok"] else 1
    if args.cmd == "mark-stale":
        res = mark_stale(args.repo, args.match, args.date, args.cause, target_file=args.file)
        print(json.dumps(res, indent=2))
        return 0 if res["ok"] else 1
    return 2


if __name__ == "__main__":
    sys.exit(main())
