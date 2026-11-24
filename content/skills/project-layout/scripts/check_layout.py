#!/usr/bin/env python3
"""Check a project's support layout without judgment (read-only).

Rules (see references/layout.md):
  LAYOUT-01  script under _support/scripts/ that HIVE.local.md does not name
  LAYOUT-02  tracked file under _support/{scripts,secrets,workspace,data}
  LAYOUT-03  tracked code file outside _support/, not a test, whose code (not a comment)
             cites work material: _support/{workspace,data,scripts,secrets,evidence,reports}
             (review: a script cannot tell application code from a tool that manages _support/)
  LAYOUT-04  top-level _support/ entry outside the convention, a legacy home among them
             (sessions, evidence, reports, archive, archived, backups, plan, spec), or a
             change register under _support/openspec/changes (review)
  LAYOUT-05  document missing from the index, broken index link, tool not indexed
  LAYOUT-06  coordinator table row whose repo is missing, has another remote,
             is not ignored or is not re-included in .ignore
  LAYOUT-07  undeclared root docs/ folder with Markdown (review)
  LAYOUT-08  tracked code or infrastructure under _support/ outside docs, openspec
             and the LAYOUT-02 homes

Files inside dependency or build output folders of any ecosystem are never read:
inside Git, only tracked and non-ignored files are listed, plus the ignored files
under _support/; outside Git, child repositories, PRUNE and GATED folders are skipped.

Output: one finding per line, `<rule> <severity> <path>: <problem> - <correction>`.
Exit 1 when any finding is `fail`, otherwise 0. Never opens _support/secrets/ nor
reads the content of _support/data/.
"""

import argparse
import fnmatch
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
    "docs", "openspec", "workspace", "data", "scripts", "secrets",
    "README.md", ".gitkeep", ".DS_Store",
}
# Homes of earlier layouts, which Hive no longer creates; project-layout migrates them.
LEGACY_HOMES = {"sessions", "evidence", "reports", "archive", "archived", "backups", "plan", "spec"}
UNTRACKED_SUPPORT = ("scripts", "secrets", "workspace", "data")
IGNORED_NAMES = {"__pycache__", ".DS_Store"}
# Dependency and build output, from the github/gitignore templates of each ecosystem.
PRUNE = {
    ".git", "node_modules", ".pnpm-store", ".next", ".nuxt", ".turbo", ".venv", "venv",
    "__pycache__", ".tox", ".mypy_cache", ".pytest_cache", ".terraform", ".gradle",
    ".dart_tool", "_build", "Pods", "site-packages", ".bundle",
}
# Common names that are build output only when the manifest that produces them sits beside them.
GATED = {
    "target": ("Cargo.toml", "pom.xml"),
    "vendor": ("composer.json", "go.mod", "Gemfile"),
    "dist": ("package.json", "pyproject.toml", "setup.py", "build.gradle", "build.gradle.kts", "pubspec.yaml"),
    "build": ("package.json", "pyproject.toml", "setup.py", "build.gradle", "build.gradle.kts", "pubspec.yaml"),
    "bin": ("*.csproj", "*.fsproj"),
    "obj": ("*.csproj", "*.fsproj"),
    "deps": ("mix.exs",),
}
# The GATED names that hold other projects' code rather than build output.
DEPENDENCY_NAMES = ("vendor", "deps")
IAC_EXTENSIONS = {".tf", ".tfvars", ".hcl"}
IAC_FAMILY = {"infra", "infrastructure", "terraform", "deployments", "aws", "gcp", "azure", "iam", "cloud"}
CONFIG_EXTENSIONS = {".json", ".yaml", ".yml"}
NOT_CODE_HOMES = {"docs", "openspec"}
LINK_RE = re.compile(r"\]\((<[^>]+>|[^)\s]+)")
SCHEME_RE = re.compile(r"^(?:[a-zA-Z][a-zA-Z0-9+.-]*:|#|/)")
TEST_PATH_RE = re.compile(r"(?:^|/)(?:tests?|__tests__|spec|e2e)/|[._-](?:test|spec)\.[a-z]+$|_test\.[a-z]+$|(?:^|/)test_[^/]*$")
# Code may read versioned knowledge (_support/docs, _support/openspec); it must not
# depend on work material, which a fresh clone does not have.
WORK_MATERIAL_RE = re.compile(r"_support/(?:workspace|data|scripts|secrets|evidence|reports|sessions)\b")
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


def git(root, *args, input=None):
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    env["GIT_TERMINAL_PROMPT"] = "0"
    try:
        return subprocess.run(
            ["git", "-C", root, *args], text=True, capture_output=True, env=env, timeout=60, input=input
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
            key_match = re.match(r"^(docs|tools|infra) at (\S+)\s*$", part.strip())
            if key_match:
                layout[key_match.group(1)] = norm_path(key_match.group(2))
        break
    return layout


def git_toplevel(path):
    result = git(path, "rev-parse", "--show-toplevel")
    if result is None or result.returncode != 0:
        return None
    return os.path.realpath(result.stdout.strip())


def has_manifest(directory, patterns):
    try:
        names = os.listdir(directory)
    except OSError:
        return False
    return any(fnmatch.fnmatch(name, pattern) for pattern in patterns for name in names)


def is_dependency_dir(parent, name):
    if name in PRUNE:
        return True
    return name in GATED and has_manifest(parent, GATED[name])


def _crosses(path, top, test):
    rel = os.path.relpath(path, top)
    parent = top
    for name in rel.split(os.sep)[:-1]:
        if test(parent, name):
            return True
        parent = os.path.join(parent, name)
    return False


def under_dependency(path, top, gated=True):
    """Whether a path below top crosses a dependency or build output folder. Inside Git,
    pass gated=False: Git already dropped ignored output, so a `build/` it lists is source."""
    return _crosses(path, top, lambda parent, name: name in PRUNE or (gated and is_dependency_dir(parent, name)))


def under_vendored(path, top):
    """Whether a path below top crosses a folder of other projects' code, even one the
    repository commits, such as Go's `vendor/` beside `go.mod`. Build output it commits,
    such as `build/`, is its own source and does not count."""
    return _crosses(path, top, lambda parent, name: name in PRUNE or (name in DEPENDENCY_NAMES and is_dependency_dir(parent, name)))


def walk_pruned(base):
    found = []
    for dirpath, dirnames, filenames in os.walk(base):
        dirnames[:] = sorted(
            d for d in dirnames
            if d not in IGNORED_NAMES and not is_dependency_dir(dirpath, d)
            and not os.path.exists(os.path.join(dirpath, d, ".git"))
        )
        found.extend(os.path.join(dirpath, n) for n in sorted(filenames) if n not in IGNORED_NAMES)
    return found


def list_files(base):
    """Files under base, sorted, without dependency or build output of any ecosystem.

    Paths keep the spelling of base, so callers can take relpaths against it."""
    real = os.path.realpath(base)
    return [os.path.join(base, os.path.relpath(p, real)) for p in _list_real(real)]


def _list_real(base):
    if not os.path.isdir(base):
        return []
    top = git_toplevel(base)
    if top is None:
        return walk_pruned(base)
    rel_base = os.path.relpath(base, top)
    prefix = "" if rel_base == "." else rel_base + "/"
    paths = set()
    listed = git(top, "ls-files", "-z", "--cached", "--others", "--exclude-standard", "--", rel_base)
    if listed is not None and listed.returncode == 0:
        paths.update(p for p in listed.stdout.split("\0") if p and not p.endswith("/"))
    # The convention keeps unversioned material ignored under _support/: list it too.
    # --directory names an ignored folder without walking it, such as node_modules.
    if "_support" in rel_base.split(os.sep):
        specs = [rel_base]
    else:
        specs = [f"{prefix}_support", f":(glob){prefix}**/_support/**"]
    ignored = git(top, "ls-files", "-z", "--others", "--ignored", "--exclude-standard", "--directory", "--", *specs)
    if ignored is not None and ignored.returncode == 0:
        for entry in (p for p in ignored.stdout.split("\0") if p):
            if "_support" not in entry.rstrip("/").split("/"):
                continue  # --directory also names ignored folders the pathspec does not cover
            if entry.endswith("/"):
                folder = os.path.join(top, entry.rstrip("/"))
                if not under_dependency(os.path.join(folder, "x"), top) and os.path.basename(folder) not in PRUNE:
                    paths.update(os.path.relpath(f, top) for f in walk_pruned(folder))
            else:
                paths.add(entry)
    found = []
    for rel in sorted(paths):
        path = os.path.join(top, rel)
        if os.path.basename(rel) in IGNORED_NAMES or not os.path.isfile(path) or under_dependency(path, top, gated=False):
            continue
        found.append(path)
    return found


def check_scripts(root, out):
    scripts = os.path.join(root, "_support", "scripts")
    if not os.path.isdir(scripts):
        return
    local = read_text(os.path.join(root, "HIVE.local.md"))
    named = set(re.findall(r"_support/scripts/([^\s`'\")\]]+)", local))
    named = {n.rstrip("/.,;:") for n in named}
    for path in list_files(scripts):
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


def code_under_support(tracked):
    """Tracked code or infrastructure under _support/ that LAYOUT-08 reports, sorted."""
    tf_entries = {
        rel.split("/")[1] for rel in tracked
        if rel.startswith("_support/") and rel.count("/") > 1 and os.path.splitext(rel)[1] == ".tf"
    }
    found = []
    for rel in tracked:
        parts = rel.split("/")
        if parts[0] != "_support" or len(parts) <= 2:
            continue
        if parts[1] in NOT_CODE_HOMES or parts[1] in UNTRACKED_SUPPORT:
            continue
        extension = os.path.splitext(rel)[1].lower()
        infra_tree = parts[1] in IAC_FAMILY or parts[1] in tf_entries
        if extension in CODE_EXTENSIONS or extension in IAC_EXTENSIONS or (infra_tree and extension in CONFIG_EXTENSIONS):
            found.append(rel)
    return sorted(found)


def check_tracked(root, out):
    tracked = tracked_files(root)
    for rel in code_under_support(tracked):
        out.append(Finding(
            "LAYOUT-08", "fail", os.path.join(root, rel),
            "tracked code or infrastructure under _support/",
            f"run plan_layout.py plan --root {root} and move it with the project's code",
        ))
    for rel in tracked:
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
        if name in LEGACY_HOMES:
            out.append(Finding(
                "LAYOUT-04", "review", os.path.join(support, name),
                "legacy home of an earlier layout",
                f"run plan_layout.py plan --root {root} and migrate it with project-layout",
            ))
        elif name not in SUPPORT_CHILDREN:
            out.append(Finding(
                "LAYOUT-04", "review", os.path.join(support, name),
                "top-level _support/ entry outside the convention",
                "classify it with project-layout organize mode",
            ))
    changes = os.path.join(support, "openspec", "changes")
    if os.path.isdir(changes):
        out.append(Finding(
            "LAYOUT-04", "review", changes,
            "legacy change register of an earlier layout",
            f"run plan_layout.py plan --root {root} and migrate it with project-layout",
        ))


def check_index(root, index, layout, out):
    index_text = read_text(index)
    index_dir = os.path.dirname(index)
    docs_rel = layout.get("docs", "_support/docs")
    docs_home = os.path.join(root, docs_rel)
    if os.path.isdir(docs_home):
        for path in list_files(docs_home):
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
    if any(p.lower().endswith(".md") for p in list_files(docs)):
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
