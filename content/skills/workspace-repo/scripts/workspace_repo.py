#!/usr/bin/env python3
"""Create or restore a workspace coordinator repository.

Subcommands:
  plan   (default) Print what apply would write; write nothing.
  apply  Write .gitignore, .ignore, AGENTS.md sections and the _support skeleton
         at the workspace root, then run `git init`. Never creates a remote,
         never pushes, never commits.
  setup  Clone every repository of the AGENTS.md `## Repositories` table that
         is missing from disk.
"""

import argparse
import os
import re
import subprocess
import sys

SUBCOMMANDS = ("plan", "apply", "setup")
IGNORED_SUPPORT = ("workspace", "data", "scripts", "secrets")
TABLE_HEADER = "| Repo | Path | Remote | Base | Role |"
TABLE_RULE = "| --- | --- | --- | --- | --- |"


class Failure(Exception):
    pass


def git(path, *args):
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    env["GIT_TERMINAL_PROMPT"] = "0"
    try:
        return subprocess.run(["git", "-C", path, *args], text=True, capture_output=True, env=env, timeout=600)
    except (OSError, subprocess.SubprocessError) as error:
        raise Failure(f"git {' '.join(args)} failed to run: {error}")


def read_text(path):
    try:
        with open(path, encoding="utf-8") as handle:
            return handle.read()
    except OSError:
        return ""


def norm_path(value):
    value = value.strip().strip("`")
    while value.startswith("./"):
        value = value[2:]
    return value.rstrip("/")


def parse_repo_args(root, items):
    repos = []
    for item in items:
        path, _, remote = item.partition("=")
        path = norm_path(path)
        if not path or path.startswith("/") or ".." in path.split("/"):
            raise Failure(f"--repo path must be relative to the root: {item}")
        child = os.path.join(root, path)
        if not os.path.isdir(child):
            raise Failure(f"repository directory not found: {child}")
        remote = remote.strip()
        if not remote:
            result = git(child, "config", "--get", "remote.origin.url")
            remote = result.stdout.strip() if result.returncode == 0 else ""
        if not remote:
            raise Failure(f"{path} has no remote.origin.url; pass --repo {path}=<remote>")
        result = git(child, "symbolic-ref", "--short", "HEAD")
        base = result.stdout.strip() if result.returncode == 0 else "—"
        repos.append({"path": path, "remote": remote, "base": base})
    return repos


def merge_lines(existing, wanted):
    """Keep every existing line; append wanted lines that are missing, once."""
    lines = existing.splitlines()
    present = {line.strip() for line in lines}
    added = []
    for line in wanted:
        if line not in present:
            lines.append(line)
            present.add(line)
            added.append(line)
    text = "\n".join(lines)
    return (text + "\n" if text else ""), added


def hive_section(args):
    lines = ["## Hive", "", f"- Project: {args.project}", f"- Base branch: {args.base_branch}",
             f"- Tracker: {args.tracker or '<tracker>'}"]
    if args.direct_base:
        lines.append("- Delivery: direct-base")
    lines.append(f"- Specs: {args.specs}")
    return "\n".join(lines) + "\n"


def repositories_section(repos, previous_roles):
    lines = ["## Repositories", "", TABLE_HEADER, TABLE_RULE]
    for repo in repos:
        role = previous_roles.get(repo["path"], "—")
        name = os.path.basename(repo["path"])
        lines.append(f"| {name} | {repo['path']} | {repo['remote']} | {repo['base']} | {role} |")
    return "\n".join(lines) + "\n"


def replace_section(text, heading, section):
    """Replace the section that starts with `heading`, or append it; keep all other text."""
    lines = text.splitlines()
    start = next((i for i, l in enumerate(lines) if re.match(rf"^{re.escape(heading)}\s*$", l)), None)
    if start is None:
        body = text.rstrip("\n")
        return (body + "\n\n" if body else "") + section
    end = next((i for i in range(start + 1, len(lines)) if re.match(r"^#{1,2}\s", lines[i])), len(lines))
    before = "\n".join(lines[:start]).rstrip("\n")
    after = "\n".join(lines[end:]).lstrip("\n")
    out = (before + "\n\n" if before else "") + section
    if after:
        out += "\n" + after + ("\n" if not after.endswith("\n") else "")
    return out


def parse_repositories(text):
    lines = text.splitlines()
    start = next((i + 1 for i, l in enumerate(lines) if re.match(r"^##\s+(Repositories|Repositorios)\s*$", l)), None)
    if start is None:
        return []
    header, rows = None, []
    for line in lines[start:]:
        if line.startswith("#"):
            break
        if not line.strip().startswith("|"):
            continue
        cells = [c.strip().strip("`") for c in line.strip().strip("|").split("|")]
        if all(re.match(r"^:?-+:?$", c) for c in cells if c):
            continue
        if header is None:
            header = [c.lower() for c in cells]
            continue
        if "path" in header and "remote" in header:
            rows.append(dict(zip(header, cells)))
    return rows


def build_plan(args):
    root = os.path.abspath(args.root)
    if not os.path.isdir(root):
        raise Failure(f"root is not a directory: {root}")
    repos = parse_repo_args(root, args.repo or [])
    files = {}
    gitignore_wanted = [f"/{r['path']}/" for r in repos] + ["/HIVE.local.md"]
    if os.path.isdir(os.path.join(root, ".engram")):
        gitignore_wanted.append("/.engram/")
    gitignore_wanted += [f"/_support/{name}/" for name in IGNORED_SUPPORT]
    files[".gitignore"], _ = merge_lines(read_text(os.path.join(root, ".gitignore")), gitignore_wanted)
    files[".ignore"], _ = merge_lines(read_text(os.path.join(root, ".ignore")), [f"!/{r['path']}/" for r in repos])
    agents_path = os.path.join(root, "AGENTS.md")
    existing = read_text(agents_path) if os.path.exists(agents_path) else ""
    if not existing:
        existing = f"# {args.project}\n\nRead `_support/docs/README.md` first.\n"
    previous_roles = {}
    for row in parse_repositories(existing):
        previous_roles[norm_path(row.get("path", ""))] = row.get("role", "—") or "—"
    text = replace_section(existing, "## Hive", hive_section(args))
    text = replace_section(text, "## Repositories", repositories_section(repos, previous_roles))
    files["AGENTS.md"] = text
    index = os.path.join(root, "_support", "docs", "README.md")
    if not os.path.exists(index):
        files["_support/docs/README.md"] = f"# {args.project} documentation index\n\nOne line per document, runbook and internal tool, with when to read or use it.\n"
    openspec = os.path.join(root, "_support", "openspec")
    if not os.path.isdir(openspec) or not os.listdir(openspec):
        files[os.path.join(args.specs, ".gitkeep")] = ""
    return root, files


def cmd_plan(args, apply=False):
    if apply:
        if not args.tracker:
            raise Failure("--tracker is required for apply")
        if not args.project:
            raise Failure("--project is required for apply")
    args.project = args.project or os.path.basename(os.path.abspath(args.root))
    root, files = build_plan(args)
    if apply and os.path.exists(os.path.join(root, ".git")):
        raise Failure(f"{root} already has .git; refusing to initialize a coordinator over an existing repository")
    for rel, content in files.items():
        path = os.path.join(root, rel)
        current = read_text(path) if os.path.exists(path) else None
        status = "create" if current is None else ("unchanged" if current == content else "update")
        print(f"{status} {rel}")
        if not apply:
            if status != "unchanged":
                print("".join(f"  | {line}\n" for line in content.splitlines()), end="")
            continue
        if status == "unchanged":
            continue
        os.makedirs(os.path.dirname(path), exist_ok=True)
        with open(path, "w", encoding="utf-8") as handle:
            handle.write(content)
    if not apply:
        print("plan only: nothing written. Re-run with `apply` to write it.")
        return 0
    result = git(root, "init", "-q", "-b", args.base_branch)
    if result.returncode != 0:
        raise Failure(f"git init failed: {result.stderr.strip()}")
    print(f"git init -b {args.base_branch}: done")
    print("Next (not run): git -C", root, "add -A && git -C", root, 'commit -m "chore: workspace coordinator"')
    print("A remote is not created and nothing is pushed; adding `origin` later changes the viewer project key.")
    return 0


def cmd_setup(args):
    root = os.path.abspath(args.root)
    rows = parse_repositories(read_text(os.path.join(root, "AGENTS.md")))
    if not rows:
        raise Failure(f"no `## Repositories` table in {os.path.join(root, 'AGENTS.md')}")
    failed = False
    for row in rows:
        rel, remote = norm_path(row.get("path", "")), row.get("remote", "").strip()
        child = os.path.join(root, rel)
        if os.path.isdir(os.path.join(child, ".git")) or (os.path.isdir(child) and os.listdir(child)):
            print(f"skip {rel}: already present")
            continue
        if not remote or remote in ("—", "-"):
            print(f"skip {rel}: no remote in the table")
            failed = True
            continue
        result = git(root, "clone", "-q", remote, child)
        if result.returncode != 0:
            print(f"error {rel}: clone failed: {result.stderr.strip()}", file=sys.stderr)
            failed = True
        else:
            print(f"cloned {rel} from {remote}")
    return 1 if failed else 0


def build_parser():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = parser.add_subparsers(dest="command")
    for name in ("plan", "apply"):
        p = sub.add_parser(name, help=f"{name} the coordinator files")
        p.add_argument("--root", required=True, help="workspace directory")
        p.add_argument("--repo", action="append", help="<path>=<remote>, path relative to root (repeatable)")
        p.add_argument("--project", help="project name for ## Hive")
        p.add_argument("--base-branch", default="main", help="coordinator's initial branch (default main)")
        p.add_argument("--tracker", help='tracker value, e.g. "GitHub Issues · owner/repo" (required for apply)')
        p.add_argument("--direct-base", action="store_true", help="write `- Delivery: direct-base`")
        p.add_argument("--specs", default="_support/openspec", help="Specs path (default _support/openspec)")
    p = sub.add_parser("setup", help="clone repositories missing from the table")
    p.add_argument("--root", required=True, help="workspace directory")
    return parser


def main(argv=None):
    argv = list(sys.argv[1:] if argv is None else argv)
    if argv and argv[0] not in SUBCOMMANDS and argv[0] not in ("-h", "--help"):
        argv.insert(0, "plan")
    args = build_parser().parse_args(argv)
    if not args.command:
        build_parser().print_help()
        return 0
    try:
        if args.command == "setup":
            return cmd_setup(args)
        return cmd_plan(args, apply=args.command == "apply")
    except Failure as error:
        print(f"error: {error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
