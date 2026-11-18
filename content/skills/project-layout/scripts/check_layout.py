#!/usr/bin/env python3
"""Check a project's support layout without judgment (read-only).

Rules (see references/layout.md):
  LAYOUT-01  script under _support/scripts/ that HIVE.local.md does not name
  LAYOUT-02  tracked file under _support/{scripts,secrets,workspace,data}
  LAYOUT-03  tracked code file outside _support/, not a test, whose code (not a comment)
             cites work material: _support/{workspace,data,scripts,secrets,sessions,evidence,reports}
             (review: a script cannot tell application code from a tool that manages _support/)
  LAYOUT-04  top-level _support/ entry outside the convention (review)
  LAYOUT-05  document missing from the index, broken index link, tool not indexed
  LAYOUT-06  coordinator table row whose repo is missing, has another remote,
             is not ignored or is not re-included in .ignore
  LAYOUT-07  undeclared root docs/ folder with Markdown (review)

Output: one finding per line, `<rule> <severity> <path>: <problem> - <correction>`.
Exit 1 when any finding is `fail`, otherwise 0. Never opens _support/secrets/ nor
reads the content of _support/data/.
"""

import argparse
import json
import os
import re
import subprocess
import sys
from urllib.parse import unquote

CODE_EXTENSIONS = {
    ".js", ".mjs", ".cjs", ".ts", ".mts", ".tsx", ".jsx", ".py", ".go", ".rb",
    ".php", ".java", ".kt", ".cs", ".sh", ".sql", ".ps1",
}
SUPPORT_CHILDREN = {
    "docs", "openspec", "sessions", "workspace", "data", "scripts", "secrets",
    "README.md", ".gitkeep", ".DS_Store",
}
UNTRACKED_SUPPORT = ("scripts", "secrets", "workspace", "data")
IGNORED_NAMES = {"__pycache__", ".DS_Store"}
LINK_RE = re.compile(r"\]\((<[^>]+>|[^)\s]+)")
SCHEME_RE = re.compile(r"^(?:[a-zA-Z][a-zA-Z0-9+.-]*:|#|/)")
TEST_PATH_RE = re.compile(r"(?:^|/)(?:tests?|__tests__|spec|e2e)/|[._-](?:test|spec)\.[a-z]+$|_test\.[a-z]+$|(?:^|/)test_[^/]*$")
# Code may read versioned knowledge (_support/docs, _support/openspec); it must not
# depend on work material, which a fresh clone does not have.
WORK_MATERIAL_RE = re.compile(r"_support/(?:workspace|data|scripts|secrets|sessions|evidence|reports)\b")
# Whole-line comment markers by language, so a shell flag line such as
# `  --input _support/data/x` is still code.
HASH_COMMENT = re.compile(r"^\s*#")
C_COMMENT = re.compile(r"^\s*(?://|/\*|\*)")
SQL_COMMENT = re.compile(r"^\s*--")
COMMENT_BY_EXTENSION = {
    **{ext: HASH_COMMENT for ext in (".py", ".rb", ".sh", ".ps1")},
    **{ext: C_COMMENT for ext in (".js", ".mjs", ".cjs", ".ts", ".mts", ".tsx", ".jsx", ".go", ".php", ".java", ".kt", ".cs")},
    ".sql": SQL_COMMENT,
}


class Finding:
    def __init__(self, rule, severity, path, problem, correction):
        self.rule = rule
        self.severity = severity
        self.path = path
        self.problem = problem
        self.correction = correction

    def line(self):
        return f"{self.rule} {self.severity} {self.path}: {self.problem} — {self.correction}"

    def as_dict(self):
        return {
            "rule": self.rule, "severity": self.severity, "path": self.path,
            "problem": self.problem, "correction": self.correction,
        }


def git(root, *args):
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    env["GIT_TERMINAL_PROMPT"] = "0"
    try:
        return subprocess.run(
            ["git", "-C", root, *args], text=True, capture_output=True, env=env, timeout=60
        )
    except (OSError, subprocess.SubprocessError):
        return None


def is_git_toplevel(path):
    result = git(path, "rev-parse", "--show-toplevel")
    if result is None or result.returncode != 0:
        return False
    return os.path.realpath(result.stdout.strip()) == os.path.realpath(path)


def read_text(path):
    try:
        with open(path, encoding="utf-8", errors="replace") as handle:
            return handle.read()
    except OSError:
        return ""


def norm_path(value):
    value = value.strip().strip("`")
    while value.startswith("./"):
        value = value[2:]
    return value.rstrip("/")


def find_index(root):
    for rel in ("_support/docs/README.md", "docs/README.md"):
        if os.path.isfile(os.path.join(root, rel)):
            return os.path.join(root, rel)
    for line in read_text(os.path.join(root, "AGENTS.md")).splitlines():
        match = re.match(r"^\s*[-*]?\s*Index:\s*`?([^`\s]+)`?\s*$", line)
        if match:
            candidate = os.path.join(root, norm_path(match.group(1)))
            if os.path.isfile(candidate):
                return candidate
    return None


def parse_layout(index_text):
    layout = {}
    for line in index_text.splitlines():
        match = re.match(r"^\s*Layout:\s*(.+?)\s*$", line)
        if not match:
            continue
        for part in match.group(1).split(" · "):
            key_match = re.match(r"^(docs|tools) at (\S+)\s*$", part.strip())
            if key_match:
                layout[key_match.group(1)] = norm_path(key_match.group(2))
        break
    return layout


def walk_files(base):
    for dirpath, dirnames, filenames in os.walk(base):
        dirnames[:] = sorted(d for d in dirnames if d not in IGNORED_NAMES)
        for name in sorted(filenames):
            if name in IGNORED_NAMES:
                continue
            yield os.path.join(dirpath, name)


def check_scripts(root, out):
    scripts = os.path.join(root, "_support", "scripts")
    if not os.path.isdir(scripts):
        return
    local = read_text(os.path.join(root, "HIVE.local.md"))
    named = set(re.findall(r"_support/scripts/([^\s`'\")\]]+)", local))
    named = {n.rstrip("/.,;:") for n in named}
    for path in walk_files(scripts):
        rel = os.path.relpath(path, scripts).replace(os.sep, "/")
        top = rel.split("/")[0]
        if rel in named or top in named:
            continue
        out.append(Finding(
            "LAYOUT-01", "fail", os.path.join(root, "_support", "scripts", rel),
            "local script not named in HIVE.local.md",
            "name it in HIVE.local.md, move it to the owning repository's code, or delete it if it was one-off",
        ))


def tracked_files(root):
    result = git(root, "ls-files", "-z")
    if result is None or result.returncode != 0:
        return []
    return [p for p in result.stdout.split("\0") if p]


def check_tracked(root, out):
    for rel in tracked_files(root):
        parts = rel.split("/")
        if parts[0] == "_support":
            if len(parts) > 2 and parts[1] in UNTRACKED_SUPPORT and parts[-1] != ".gitkeep":
                out.append(Finding(
                    "LAYOUT-02", "fail", os.path.join(root, rel),
                    f"tracked file under _support/{parts[1]}/ (never versioned)",
                    f"git rm --cached it and ignore _support/{parts[1]}/",
                ))
            continue
        if parts[0] == ".engram":
            continue
        extension = os.path.splitext(rel)[1].lower()
        if extension not in CODE_EXTENSIONS or TEST_PATH_RE.search(rel):
            continue
        comment = COMMENT_BY_EXTENSION.get(extension)
        for number, text in enumerate(read_text(os.path.join(root, rel)).splitlines(), 1):
            if not (comment and comment.match(text)) and WORK_MATERIAL_RE.search(text):
                out.append(Finding(
                    "LAYOUT-03", "review", os.path.join(root, f"{rel}:{number}"),
                    "code cites work material under _support/",
                    "when this is application code or a tool every clone runs, remove the dependency and keep what it needs in its own code or configuration; a tool that manages _support/ itself may keep it",
                ))
                break


def check_support_entries(root, out):
    support = os.path.join(root, "_support")
    if not os.path.isdir(support):
        return
    for name in sorted(os.listdir(support)):
        if name not in SUPPORT_CHILDREN:
            out.append(Finding(
                "LAYOUT-04", "review", os.path.join(support, name),
                "top-level _support/ entry outside the convention",
                "classify it with project-layout organize mode",
            ))


def check_index(root, index, layout, out):
    index_text = read_text(index)
    index_dir = os.path.dirname(index)
    docs_rel = layout.get("docs", "_support/docs")
    docs_home = os.path.join(root, docs_rel)
    if os.path.isdir(docs_home):
        for path in walk_files(docs_home):
            if not path.lower().endswith(".md") or os.path.realpath(path) == os.path.realpath(index):
                continue
            rel = os.path.relpath(path, index_dir).replace(os.sep, "/")
            if rel in index_text or os.path.basename(path) in index_text:
                continue
            out.append(Finding(
                "LAYOUT-05", "fail", path, "document missing from the index",
                f"add a line for it to {os.path.relpath(index, root)} saying when to read it",
            ))
    for target in LINK_RE.findall(index_text):
        if target.startswith("<") and target.endswith(">"):
            target = target[1:-1]
        if SCHEME_RE.match(target):
            continue
        clean = unquote(target.split("#")[0].split("?")[0])
        if clean and not os.path.exists(os.path.join(index_dir, clean)):
            out.append(Finding(
                "LAYOUT-05", "fail", index, f"index link to a missing file: {target}",
                "fix the link or remove the line",
            ))
    tools = layout.get("tools")
    if tools:
        tools_dir = os.path.join(root, tools)
        if os.path.isdir(tools_dir):
            for name in sorted(os.listdir(tools_dir)):
                path = os.path.join(tools_dir, name)
                if name in IGNORED_NAMES or not os.path.isfile(path) or "test" in name.lower():
                    continue
                if name not in index_text:
                    out.append(Finding(
                        "LAYOUT-05", "fail", path, "internal tool missing from the index",
                        f"add a line for it to {os.path.relpath(index, root)} saying when to use it",
                    ))


def parse_repositories(agents_text):
    rows = []
    lines = agents_text.splitlines()
    start = None
    for i, line in enumerate(lines):
        if re.match(r"^##\s+(Repositories|Repositorios)\s*$", line):
            start = i + 1
            break
    if start is None:
        return None
    header = None
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
            row = dict(zip(header, cells))
            rows.append((norm_path(row.get("path", "")), row.get("remote", "")))
    if header is None or "path" not in header or "remote" not in header:
        return None
    return rows


def norm_remote(value):
    value = value.strip().rstrip("/")
    if value.endswith(".git"):
        value = value[:-4]
    return value.rstrip("/")


def check_coordinator(root, out):
    rows = parse_repositories(read_text(os.path.join(root, "AGENTS.md")))
    if not rows:
        return
    root_is_repo = is_git_toplevel(root)
    ignore_lines = {l.strip() for l in read_text(os.path.join(root, ".ignore")).splitlines()}
    for rel, remote in rows:
        child = os.path.join(root, rel)
        if not os.path.isdir(child) or not is_git_toplevel(child):
            out.append(Finding(
                "LAYOUT-06", "fail", child, "listed repository is missing or is not a Git repository",
                "run workspace-repo setup to clone it from its Remote",
            ))
            continue
        result = git(child, "config", "--get", "remote.origin.url")
        actual = result.stdout.strip() if result is not None and result.returncode == 0 else ""
        if norm_remote(actual) != norm_remote(remote):
            out.append(Finding(
                "LAYOUT-06", "fail", child,
                f"remote origin is '{actual or 'none'}', the table says '{remote}'",
                "fix the Remote column or the repository's origin",
            ))
        if root_is_repo:
            ignored = git(root, "check-ignore", "-q", rel + "/")
            if ignored is None or ignored.returncode != 0:
                out.append(Finding(
                    "LAYOUT-06", "fail", child, "listed repository is not ignored by the coordinator",
                    f"add /{rel}/ to the coordinator's .gitignore",
                ))
        if f"!{rel}/" not in ignore_lines and f"!/{rel}/" not in ignore_lines:
            out.append(Finding(
                "LAYOUT-06", "fail", os.path.join(root, ".ignore"),
                f"listed repository {rel} is not re-included for search",
                f"add !/{rel}/ to .ignore",
            ))


def check_undeclared_docs(root, layout, out):
    docs = os.path.join(root, "docs")
    if not os.path.isdir(docs) or layout.get("docs") == "docs":
        return
    if any(p.lower().endswith(".md") for p in walk_files(docs)):
        out.append(Finding(
            "LAYOUT-07", "review", docs, "documentation home docs/ is not declared",
            "migrate it to _support/docs or declare 'Layout: docs at docs/' in the index",
        ))


def check_root(root):
    out = []
    root = os.path.abspath(root)
    index = find_index(root)
    layout = {}
    if index:
        layout = parse_layout(read_text(index))
    else:
        out.append(Finding(
            "LAYOUT-05", "review", root, "no documentation index found",
            "create _support/docs/README.md (or declare its path with an 'Index: <path>' line in AGENTS.md)",
        ))
    check_scripts(root, out)
    if is_git_toplevel(root):
        check_tracked(root, out)
    check_support_entries(root, out)
    if index:
        check_index(root, index, layout, out)
    check_coordinator(root, out)
    check_undeclared_docs(root, layout, out)
    return out


def main(argv=None):
    parser = argparse.ArgumentParser(description="Read-only check of a project's support layout.")
    parser.add_argument("--root", action="append", required=True, help="repository, workspace or coordinator directory (repeatable)")
    parser.add_argument("--json", action="store_true", help="print findings as a JSON list")
    args = parser.parse_args(argv)
    findings = []
    for root in args.root:
        if not os.path.isdir(root):
            print(f"error: not a directory: {root}", file=sys.stderr)
            return 2
        findings.extend(check_root(root))
    if args.json:
        print(json.dumps([f.as_dict() for f in findings], indent=2))
    else:
        for finding in findings:
            print(finding.line())
    return 1 if any(f.severity == "fail" for f in findings) else 0


if __name__ == "__main__":
    sys.exit(main())
