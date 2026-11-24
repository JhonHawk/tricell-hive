#!/usr/bin/env python3
"""Plan, apply, verify and undo the moves project-layout makes, deterministically.

  plan_layout.py [plan] --root DIR [--root DIR ...] [--override SRC=DEST ...]
                 [--keep SRC ...] [--home tools|infra|docs=PATH ...] [--json] [--out FILE]
  plan_layout.py apply  --plan FILE      move, rewrite references, write FILE.undo.json
  plan_layout.py verify --plan FILE      fail while a current reference to a moved path remains
  plan_layout.py undo   --plan FILE      reverse exactly what apply recorded

The plan reads observable signals only (who cites a file and from where, its name, its
extension and the folders the project already has) and applies the rules of
references/layout.md in a fixed order: the same tree gives the same plan, byte for byte.
`plan` changes nothing. `apply` never commits, pushes or stages files it does not move.
Never reads _support/secrets/ or _support/data/.
"""

import argparse
import collections
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import time
import urllib.parse

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import check_layout as cl  # noqa: E402  (python3 -I leaves the script's folder off sys.path)

VERSION = 1
STATUS_ALL = ("status", "--porcelain", "--untracked-files=all", "--ignored=matching")  # a `??` or `!!` line is a file no commit holds
UNDO_VERSION = 2  # 2 records the exact index entries of every row that leaves Git
IGNORED_HOMES = ("_support/workspace", "_support/data")  # the folders Git ignores: only a move there leaves Git
HIDING_FLAGS = 0x40000000 | 0x8000  # skip-worktree, assume-unchanged: edits that git status does not show
INDEX_FLAGS = 0x20000000 | 0x40000000 | 0x8000  # intent-to-add, skip-worktree, assume-unchanged
KEPT, KEPT_WITH = "kept by user", "kept with the kept files that use it"
MANIFESTS = (
    "package.json", "go.mod", "pyproject.toml", "setup.py", "composer.json", "Gemfile", "pom.xml",
    "build.gradle", "build.gradle.kts", "Cargo.toml", "pubspec.yaml", "mix.exs",
)
MANIFEST_SUFFIXES = (".csproj",)
TOOL_ROOTS = ("scripts", "script", "tools", "hack")
PACKAGE_TOOL_DIRS = ("scripts", "script", "tools")
IAC_HOMES = ("infra", "infrastructure", "terraform", "deployments")
GROUPS = {
    "ci": ("ci", "cicd", "pipeline", "pipelines"),
    "hooks": ("hooks", "git-hooks", "githooks"),
    "ops": ("ops", "operations", "runbooks"),
}
ONE_OFF_WORDS = {"backfill", "fix", "cutover", "migrate", "migration", "seed", "cleanup", "oneoff", "once", "repair"}
KEY_RE = re.compile(r"\b([A-Za-z]{2,10})-\d{2,}\b")
# Prefixes that look like a ticket key but name a runtime, an encoding or a version.
NOT_KEYS = {
    "node", "utf", "iso", "es", "ubuntu", "python", "py", "java", "jdk", "go", "sha", "md", "win",
    "macos", "debian", "alpine", "postgres", "pg", "mysql", "mongo", "redis", "http", "ipv", "rfc",
}
DATE_RE = re.compile(r"\d{4}-\d{2}-\d{2}")
DOC_EXTENSIONS = {".md", ".mdx", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp"}
MARKDOWN = (".md", ".mdx")
LEGACY_EVIDENCE = ("evidence", "reports")
LEGACY_DATA = ("backups", "migration-data", "exports", "dumps")
# Files that are a dump, backup or export of data: a record holding one is never deleted as aged.
DATA_SUFFIXES = frozenset((
    ".sql", ".dump", ".dmp", ".bak", ".backup", ".sqlite", ".sqlite3", ".db", ".csv", ".tsv", ".ndjson", ".jsonl", ".bson",
    ".parquet", ".avro", ".gz", ".tgz", ".zst", ".xz", ".bz2", ".7z", ".zip", ".tar", ".rar"))
DATA_NAMED_SUFFIXES = frozenset((".json", ".xml", ".xlsx"))
TEXT_FILE_LIMIT = 5_000_000
DATA_NAME_RE = re.compile(r"dump|export|backup", re.I)
# Homes of earlier layouts that hold dated records of past work: each child becomes a work folder.
LEGACY_RECORDS = ("sessions", "archive", "archived", "plan", "spec")
# The legacy homes a repository may keep at its own root, when its support home is the workspace's.
TOP_LEGACY = ("sessions", "evidence", "reports", "archive", "archived")
# Output that tooling regenerates: a legacy home holding it proposes its deletion.
REPRODUCIBLE = {"coverage", "htmlcov", "playwright-report", "test-results", ".pytest_cache", "Thumbs.db"}
RUNTIME_CONFIG_EXTENSIONS = {".json", ".yaml", ".yml", ".toml", ".ini", ".env", ".cfg", ".conf"}
CI_PATHS = re.compile(
    r"^(?:\.github/workflows/|\.circleci/)|(?:^|/)(?:\.gitlab-ci\.yml|azure-pipelines\.yml|"
    r"bitbucket-pipelines\.yml|Jenkinsfile|buildspec[^/]*\.ya?ml|amplify\.yml)$"
)
HOOK_PATHS = re.compile(r"^(?:\.husky/|\.githooks/)|(?:^|/)(?:lefthook\.ya?ml|\.pre-commit-config\.yaml)$")
# Records of what happened: references there describe the past and are never rewritten.
HISTORY_RE = re.compile(
    r"^(?:_support/)?(?:openspec/changes|sessions|evidence)/"
    r"|^_support/(?:workspace|reports|archive|archived|backups|plan|spec)/"
    r"|(?:^|/)CHANGELOG[^/]*$"
)
NEVER_READ = re.compile(r"(?:^|/)_support/(?:secrets|data)/|(?:^|/)\.engram/|(?:^|/)\.git/")
TEST_NAME = re.compile(r"^(?P<stem>.+?)(?:\.test|\.spec|_test)(?P<ext>\.[^.]+)$|^test_(?P<stem2>.+)$")
# A path that reaches into a support folder: optionally after a variable or command
# substitution and a slash (`$DIR/`, `"$ROOT"/`, `$(git rev-parse --show-toplevel)/`), with
# relative segments or a repository prefix. Boundaries keep a substring from being a token.
SUPPORT_TOKEN_RE = re.compile(
    r"(?<![\w@./-])(?P<pre>(?:\$\([^()\n]*\)|\"\$\{?\w+\}?\"|\$\{?\w+\}?)/)?"
    r"(?P<path>(?:\.{1,2}/)*(?:[\w.@+-]+/)*?_support/[\w.@+/*-]*)"
)
# A relative path with no support folder in it, such as a Markdown link `../openspec/x.md`.
RELATIVE_TOKEN_RE = re.compile(r"(?<![\w@./-])(?P<path>(?:\.{1,2}/)+[\w.@+-][\w.@+/*-]*)")
ANCHOR_RE = re.compile(r"BASH_SOURCE|dirname \"?\$0|\$\(dirname|__dirname|import\.meta\.url|__file__")
REL_LITERAL_RE = re.compile(r"(?<![\w.])((?:\.\./)*\.\.(?:/[\w.@-]+)*)(?![\w.])")
MAX_READ = 5 * 1024 * 1024
TOOL_RULES = ("R4", "R5", "R6", "R7", "R8")
# Files that decide what a build or a published package takes from their folder.
BUILD_IGNORES = (".dockerignore", ".gcloudignore", ".vercelignore", ".npmignore", ".helmignore")
GLOB_CHARS = re.compile(r"[*?\[]")
KEEP_REF_RE = re.compile(r"^(?P<path>.+?)(?::(?P<line>\d+))?$")


CASE_INSENSITIVE = None  # whether paths that differ only by case name one file; found once per run, a test may force it


def fold(path):
    """A path in the form destinations are compared in: case-folded on a case-insensitive filesystem."""
    global CASE_INSENSITIVE
    if CASE_INSENSITIVE is None:
        with tempfile.TemporaryDirectory(prefix="plan-layout-case-") as scratch:
            open(os.path.join(scratch, "probe"), "w").close()
            CASE_INSENSITIVE = os.path.exists(os.path.join(scratch, "PROBE"))
    path = os.path.normcase(path)
    return path.casefold() if CASE_INSENSITIVE else path


def fail(message):
    print(f"plan_layout.py: {message}", file=sys.stderr)
    return 1


def git(root, *args, **kwargs):
    return cl.git(root, *args, **kwargs)


def leaves_git_home(dest):
    """Whether a row moved to dest leaves Git: only a move into a folder Git ignores does; any other
    destination is a tracked home, and a move there stays tracked (git mv)."""
    if dest and dest.startswith("../"):
        return True  # outside the repository: no commit of it can hold the file, whatever the folder is called
    return bool(dest) and any(under(dest, home) is not None for home in IGNORED_HOMES)


def nested_collides(outer_source, outer_dest, dest):
    """Whether dest, which lies inside outer_dest, the destination of a moving folder, names a path the
    folder already holds at outer_source: the second move would then land on the first one's content.
    All three are absolute paths."""
    return os.path.lexists(os.path.join(outer_source, dest[len(outer_dest) + 1:]))


def is_build_ignore(rel):
    name = os.path.basename(rel)
    return name in BUILD_IGNORES or name.endswith(".dockerignore")


def ignore_base(ignore_file):
    """The folder an exclusion file's patterns start from. A `<name>.dockerignore` sits beside its
    Dockerfile but applies to the build context, which a monorepo sets to the repository root."""
    name = os.path.basename(ignore_file)
    return "" if name.endswith(".dockerignore") and name != ".dockerignore" else os.path.dirname(ignore_file)


def exclusion_lines(path, seen=()):
    """The patterns of an exclusion file as (negated, pattern), Docker's anchored to their base;
    a `.gcloudignore` also takes the lines of each file its `#!include:` lines name."""
    name = os.path.basename(path)
    docker = name.endswith(".dockerignore")
    lines = []
    for raw in cl.read_text(path).splitlines():
        line = raw.strip()
        if name == ".gcloudignore" and line.startswith("#!include:"):
            included = os.path.normpath(os.path.join(os.path.dirname(path), line[len("#!include:"):].strip()))
            if included not in seen:
                lines += exclusion_lines(included, (*seen, path))
            continue
        if not line or line.startswith("#"):
            continue
        negated = line.startswith("!")
        line = line[1:].strip() if negated else line
        if docker:  # Docker reads every pattern from the context root; Git's form for that is a leading slash
            line = "/" + re.sub(r"^(?:\./|/)+", "", line)
        lines.append((negated, line))
    return lines


def excluded_by(path, probes):
    """The probes (paths relative to the exclusion file's folder) it leaves out, matched by
    Git's own ignore rules in an empty repository. One difference with Docker: a `!` line
    cannot bring back a file under a folder an earlier line excludes."""
    if not probes:
        return set()
    scratch = tempfile.mkdtemp(prefix="plan-layout-")
    try:
        init = cl.git(scratch, "init", "-q")
        if init is None or init.returncode != 0:
            raise SystemExit(fail(f"cannot check {path}: git init failed in a temporary folder"))
        patterns = os.path.join(scratch, "exclude")
        with open(patterns, "w", encoding="utf-8") as handle:
            handle.writelines(("!" if negated else "") + line + "\n" for negated, line in exclusion_lines(path))
        result = cl.git(scratch, "-c", f"core.excludesFile={patterns}", "check-ignore", "--no-index", "-z", "--stdin",
                        input="\0".join(sorted(set(probes))) + "\0")
        if result is None or result.returncode not in (0, 1):  # 1: nothing matched
            raise SystemExit(fail(f"cannot check {path}: {result.stderr.strip() if result else 'git check-ignore failed'}"))
        return {item for item in result.stdout.split("\0") if item}
    finally:
        shutil.rmtree(scratch, ignore_errors=True)


def under(rel, base):
    """rel relative to the folder base, or None when it lies outside it."""
    if not base:
        return rel
    return rel[len(base) + 1:] if rel.startswith(base + "/") else None


def build_status(root, ignore_file, paths):
    """For each root-relative path: True when the exclusion file leaves it out of its build.
    A path outside the file's folder is not in that build, so it counts as left out. A folder
    path ending in / is probed through a child, since Git matches a folder pattern only then."""
    base = ignore_base(ignore_file)
    probes = {p: None if p.startswith("../") else under(p.rstrip("/"), base) for p in paths}
    asked = {p: (local + "/.layout-probe" if p.endswith("/") else local) for p, local in probes.items() if local}
    left_out = excluded_by(os.path.join(root, ignore_file), list(asked.values()))
    return {p: (asked[p] in left_out if p in asked else True) for p in paths}


def names_path(lines, ignore_file, dest, new_folder):
    """Whether the exclusion file's lines name dest or a folder of it the move created:
    whoever wrote that line decided about the new path."""
    base = ignore_base(ignore_file)
    accepted = {dest.rstrip("/")}
    if new_folder:
        parent = os.path.dirname(dest.rstrip("/"))
        while parent and (parent == new_folder or parent.startswith(new_folder + "/")):
            accepted.add(parent)
            parent = os.path.dirname(parent)
    accepted = {under(a, base) for a in accepted} - {None}
    for _, line in lines:
        clean = re.sub(r"^(?:\./|/)+", "", line).rstrip("/")
        if clean in accepted and not GLOB_CHARS.search(clean):
            return True
    return False


def left_ignored(root, folder):
    """What Git ignores under folder, a wholly ignored folder as one entry."""
    result = git(root, "ls-files", "-z", "--others", "--ignored", "--exclude-standard", "--directory", "--", folder)
    if result is None or result.returncode != 0:
        return []
    found = sorted(item for item in result.stdout.split("\0") if item)
    return [p for p in found if not any(q != p and q.endswith("/") and p.startswith(q) for q in found)]


def line_hash(line):
    return sha256_bytes(line.rstrip("\r\n").encode())


def plural(n, one, many):
    return f"{n} {one if n == 1 else many}"


def sha256_bytes(data):
    return hashlib.sha256(data).hexdigest()


def kebab(text):
    return re.sub(r"-+", "-", re.sub(r"[^a-z0-9]+", "-", text.lower())).strip("-")


def read_bytes(path):
    try:
        with open(path, "rb") as handle:
            return handle.read()
    except OSError:
        return None


def read_text_strict(path):
    """Text decoded as strict UTF-8 without NUL bytes, or None."""
    try:
        if os.path.getsize(path) > MAX_READ:
            return None
    except OSError:
        return None
    data = read_bytes(path)
    if data is None or b"\0" in data:
        return None
    try:
        return data.decode("utf-8")
    except UnicodeDecodeError:
        return None


def fingerprint(path):
    if os.path.isdir(path):
        entries = []
        for item in cl.list_files(path):
            try:
                entries.append(f"{os.path.relpath(item, path)}\0{os.path.getsize(item)}")
            except OSError:
                continue
        return sha256_bytes("\n".join(entries).encode())
    data = read_bytes(path)
    return sha256_bytes(data) if data is not None else None


def oldest_mtime(path):
    files = cl.list_files(path) if os.path.isdir(path) else [path]
    stamps = []
    for item in files:
        try:
            stamps.append(os.path.getmtime(item))
        except OSError:
            continue
    return time.strftime("%Y-%m-%d", time.gmtime(min(stamps))) if stamps else "0000-00-00"


def tokens(line, relative):
    """(column, prefix, path) of each path token in a line; with relative, also the
    relative paths that hold no support folder."""
    found = {}
    for match in SUPPORT_TOKEN_RE.finditer(line):
        found[match.start()] = (match.group("pre") or "", match.group("path").rstrip("."))
    if relative:
        for match in RELATIVE_TOKEN_RE.finditer(line):
            start = match.start("path")
            if not any(s <= start < s + len(p) + len(pre) for s, (pre, p) in found.items()):
                found[start] = ("", match.group("path").rstrip("."))
    return [(col, pre, path) for col, (pre, path) in sorted(found.items())]


class Root:
    def __init__(self, path):
        self.path = os.path.abspath(path)
        self.is_git = cl.is_git_toplevel(self.path)
        head = git(self.path, "rev-parse", "HEAD") if self.is_git else None
        self.head = head.stdout.strip() if head is not None and head.returncode == 0 else None
        self.files = [os.path.relpath(p, self.path).replace(os.sep, "/") for p in cl.list_files(self.path)]
        self.tracked = set(cl.tracked_files(self.path)) if self.is_git else set()
        index = cl.find_index(self.path)
        self.layout = cl.parse_layout(cl.read_text(index)) if index else {}
        self.packages = self._packages()

    def abs(self, rel):
        return os.path.normpath(os.path.join(self.path, rel))

    def _packages(self):
        found = set()
        for rel in self.files:
            if rel.startswith("_support/") or HISTORY_RE.search(rel):
                continue
            name = os.path.basename(rel)
            directory = os.path.dirname(rel)
            if directory and (name in MANIFESTS or name.endswith(MANIFEST_SUFFIXES)) and not cl.under_vendored(self.abs(rel), self.path):
                found.add(directory)
        return sorted(found)

    def package_of(self, rel):
        best = None
        for package in self.packages:
            if rel.startswith(package + "/") and (best is None or len(package) > len(best)):
                best = package
        return best

    def first_existing(self, base, names):
        for name in names:
            if os.path.isdir(os.path.join(self.path, base, name)):
                return name
        return None

    def first_commit_date(self, rel):
        result = git(self.path, "log", "--diff-filter=A", "--format=%cs", "--reverse", "--", rel)
        if result is not None and result.returncode == 0 and result.stdout.strip():
            return result.stdout.split()[0]
        return None

    def ignored(self, paths):
        """The subset of repository-relative paths Git ignores (a folder through a child probe)."""
        if not self.is_git or not paths:
            return set()
        probes = {}
        for rel in paths:
            probes[rel] = rel
            probes[rel + "/.layout-probe"] = rel
        # -z: without it Git quotes a path with non-ASCII characters and the match fails
        result = git(self.path, "check-ignore", "-z", "--stdin", input="\0".join(probes) + "\0")
        if result is None:
            return set()
        return {probes[item] for item in result.stdout.split("\0") if item in probes}


class Planner:
    def __init__(self, roots, overrides, homes, keeps=(), keep_refs=(), legacy="move"):
        self.legacy = legacy
        self.roots = [Root(r) for r in roots]
        self.overrides = overrides
        self.used_overrides = set()
        self.keeps = list(keeps)
        self.used_keeps = set()
        self.keep_refs = []  # (spec, path, line or None)
        for spec in keep_refs:
            match = KEEP_REF_RE.match(spec)
            self.keep_refs.append((spec, os.path.abspath(match["path"]) if os.path.isabs(match["path"])
                                   else os.path.normpath(match["path"]), int(match["line"]) if match["line"] else None))
        self.used_keep_refs = set()
        self.kept_refs = set()  # (root index, citing file, sha256 of the kept line or None for the whole file)
        self.homes_override = homes
        self.rows = {}  # (root index, source) -> row
        self.reviews = []
        self.rewrites = []
        self.contents = {}
        self.layout_lines = []
        self.index_lines = []
        self._homes = {}
        self._listing = {}

    # ---- homes -------------------------------------------------------------
    def homes(self, i):
        root = self.roots[i]
        if i not in self._homes:
            tools = self.homes_override.get("tools") or root.layout.get("tools") or root.first_existing("", TOOL_ROOTS) or "scripts"
            infra = self.homes_override.get("infra") or root.layout.get("infra") or root.first_existing("", IAC_HOMES) or "infra"
            docs = self.homes_override.get("docs") or root.layout.get("docs") or "_support/docs"
            groups = {key: root.first_existing(tools, aliases) or key for key, aliases in GROUPS.items()}
            self._homes[i] = {"tools": tools, "infra": infra, "docs": docs, "groups": groups}
        return self._homes[i]

    # ---- reading -------------------------------------------------------------
    def readable_files(self, i):
        """Files whose paths are references: not history, secrets, data or ignore lists."""
        for rel in self.roots[i].files:
            if NEVER_READ.search("/" + rel) or HISTORY_RE.search(rel) or os.path.basename(rel).endswith("ignore"):
                continue
            yield rel

    def content(self, i, rel):
        key = (i, rel)
        if key not in self.contents:
            self.contents[key] = read_text_strict(self.roots[i].abs(rel))
        return self.contents[key]

    def listing(self, path):
        if path not in self._listing:
            self._listing[path] = cl.list_files(path)
        return self._listing[path]

    @staticmethod
    def scans_relative(rel):
        # A Markdown link is relative to its file; a relative path in code usually is
        # relative to the working directory, so code gets the internal-path review instead.
        return rel.endswith(MARKDOWN)

    # ---- candidates ------------------------------------------------------------
    def collect(self):
        for i, root in enumerate(self.roots):
            local = cl.read_text(root.abs("HIVE.local.md"))
            named = {n.rstrip("/.,;:") for n in re.findall(r"_support/scripts/([^\s`'\")\]]+)", local)}
            for rel in root.files:
                if rel.startswith("_support/scripts/"):
                    sub = rel[len("_support/scripts/"):]
                    if sub in named or sub.split("/")[0] in named:
                        continue
                    self.add_file(i, rel, "_support/scripts/", "scripts")
            if root.is_git:
                for rel in cl.code_under_support(sorted(root.tracked)):
                    entry = rel.split("/")[1]
                    if entry in cl.LEGACY_HOMES:
                        continue  # its work folder carries the code; a row of its own would block that folder
                    self.add_file(i, rel, f"_support/{entry}/", entry)
            support = root.abs("_support")
            if os.path.isdir(support):
                for name in sorted(os.listdir(support)):
                    if name not in cl.SUPPORT_CHILDREN:
                        self.add_entry(i, name)
            if root.is_git:
                for name in TOP_LEGACY:
                    self.add_top_legacy(i, name)

    def new_row(self, i, rel, kind, entry, subpath=""):
        root = self.roots[i]
        if kind == "file":
            tracked = rel in root.tracked
        else:
            tracked = any(r.startswith(rel + "/") for r in root.tracked)
        return {
            "root": i, "source": rel, "kind": kind, "tracked": tracked, "subpath": subpath, "entry": entry,
            "rule": None, "destination": None, "notes": [], "override": False,
            "delete": False, "untrack": False,
        }

    def add_file(self, i, rel, prefix, entry):
        if (i, rel) not in self.rows:
            self.rows[(i, rel)] = self.new_row(i, rel, "file", entry, rel[len(prefix):])

    def add_entry(self, i, name):
        root = self.roots[i]
        base = f"_support/{name}"
        if not os.path.isdir(root.abs(base)):
            self.rows[(i, base)] = self.new_row(i, base, "file", name)
            return
        files = [r for r in root.files if r.startswith(base + "/")]
        if name in cl.IAC_FAMILY or any(r.endswith(".tf") for r in files):
            for rel in files:
                self.add_file(i, rel, base + "/", name)
            return
        if any((i, r) in self.rows for r in files):
            rest = [r for r in files if (i, r) not in self.rows]
            if rest and not self.kept_path(i, base):
                self.review(i, base, "no rule", f"{len(rest)} other files stay; classify them by hand")
            return
        self.rows[(i, base)] = self.new_row(i, base, "folder", name)

    def add_top_legacy(self, i, name):
        """A legacy home at the repository's own root: a folder of records, never one that holds code."""
        root = self.roots[i]
        files = [r for r in root.files if r.startswith(name + "/")]
        if os.path.islink(root.abs(name)) or not os.path.isdir(root.abs(name)):
            return
        if not files and not any(os.path.exists(os.path.join(root.abs(name), c, ".git")) for c in os.listdir(root.abs(name))):
            return
        if any(os.path.splitext(r)[1].lower() in cl.CODE_EXTENSIONS for r in files):
            return self.review(i, name, "no rule", f"{name}/ holds code, so it is not a legacy home of records; classify it by hand")
        self.rows[(i, name)] = self.new_row(i, name, "folder", name)

    def support_home(self, i, rel):
        """Where the work folders of root i go: the repository's own _support/, else, for a legacy home at
        the repository's root, the _support/ of the workspace folder when that folder is also a root of the plan."""
        root = self.roots[i]
        if rel.startswith("_support/") or os.path.isdir(root.abs("_support")):
            return "_support"
        parent = os.path.dirname(root.path)
        if any(other.path == parent for k, other in enumerate(self.roots) if k != i):
            return "../_support"
        return "_support"

    def review(self, i, path, reason, detail):
        self.reviews.append({"root": i, "path": path, "reason": reason, "detail": detail})

    # ---- references ------------------------------------------------------------
    def resolutions(self, i, citing, pre, path):
        """(base kind, absolute path) pairs a token may point to, most specific first.

        After a variable or command substitution, a relative path is read from the citing
        file's folder (the variable usually holds it) and any other path from the root;
        `./x` and `../x` from the citing file's folder, then from the root, where CI and
        package scripts run; `<repo>/x`, naming another root, from the workspace that holds
        it; any other path from the root."""
        root = self.roots[i]
        star = path.find("*")
        if star >= 0:
            path = path[:star]
        path = path.rstrip("/").rstrip(".")
        if not path:
            return []
        citing_dir = os.path.dirname(root.abs(citing))
        relative = path.startswith("./") or path.startswith("../")
        if pre:
            if relative:
                return [("file", os.path.normpath(os.path.join(citing_dir, path)))]
            return [("root", os.path.normpath(os.path.join(root.path, path)))]
        if relative:
            return [
                ("file", os.path.normpath(os.path.join(citing_dir, path))),
                ("root", os.path.normpath(os.path.join(root.path, path))),
            ]
        first = path.split("/")[0]
        found = []
        for k, other in enumerate(self.roots):
            if k != i and os.path.basename(other.path) == first:
                found.append((f"ws:{k}", os.path.normpath(os.path.join(os.path.dirname(other.path), path))))
        return found + [("root", os.path.normpath(os.path.join(root.path, path)))]

    def candidate_index(self):
        files, folders = {}, {}
        for key, row in self.rows.items():
            absolute = self.roots[row["root"]].abs(row["source"])
            (files if row["kind"] == "file" else folders)[absolute] = key
        return files, folders

    def candidate_dirs(self, files, folders):
        dirs = collections.defaultdict(list)
        roots = {r.path for r in self.roots}
        for absolute, key in list(files.items()) + list(folders.items()):
            parent = os.path.dirname(absolute)
            while parent not in roots and parent != os.path.dirname(parent):
                dirs[parent].append(key)
                parent = os.path.dirname(parent)
        return dirs

    @staticmethod
    def match_target(target, files, folders, candidate_dirs):
        """(kind, keys, inside) for a resolved path: a candidate, a file inside a candidate
        folder, or a folder below _support/<entry> holding candidates."""
        if target in files:
            return "literal", [files[target]], ""
        if target in folders:
            return "literal", [folders[target]], ""
        parent = os.path.dirname(target)
        while parent != os.path.dirname(parent):
            if parent in folders:
                return "literal", [folders[parent]], os.path.relpath(target, parent)
            parent = os.path.dirname(parent)
        if target in candidate_dirs and os.path.basename(target) != "_support":
            return "prefix", sorted(set(candidate_dirs[target])), ""
        return None, [], ""

    def references(self):
        """Every token in readable files that resolves to a candidate or a folder of candidates."""
        files, folders = self.candidate_index()
        candidate_dirs = self.candidate_dirs(files, folders)
        refs = []
        for i in range(len(self.roots)):
            for rel in self.readable_files(i):
                text = self.content(i, rel)
                if text is None:
                    data = read_bytes(self.roots[i].abs(rel)) or b""
                    if b"_support/" in data and (i, rel) not in self.rows:
                        self.review(i, rel, "unreadable", "binary or non-UTF-8 file cites _support/; not rewritten")
                    continue
                relative = self.scans_relative(rel)
                if "_support/" not in text and not relative:
                    continue
                for number, line in enumerate(text.splitlines(True), 1):
                    for col, pre, path in tokens(line, relative):
                        for base, target in self.resolutions(i, rel, pre, path):
                            kind, keys, inside = self.match_target(target, files, folders, candidate_dirs)
                            if kind:
                                refs.append({
                                    "root": i, "file": rel, "line": number, "col": col, "pre": pre, "path": path,
                                    "token": pre + path, "target": target, "base": base, "inside": inside,
                                    "kind": kind, "keys": keys, "text": line,
                                })
                                break
        return refs

    # ---- classification --------------------------------------------------------
    def caller_class(self, i, rel):
        if CI_PATHS.search(rel):
            return "ci"
        if HOOK_PATHS.search(rel):
            return "hook"
        if rel.endswith(MARKDOWN):
            return "doc"
        package = self.roots[i].package_of(rel)
        return f"package:{package}" if package else "root"

    def classify(self, refs):
        callers = collections.defaultdict(set)
        cites = collections.defaultdict(set)
        cited_by = collections.defaultdict(set)
        for ref in refs:
            citing_key = (ref["root"], ref["file"])
            for key in ref["keys"]:
                if citing_key == key:
                    continue
                if citing_key in self.rows:
                    cites[citing_key].add(key)
                    cited_by[key].add(citing_key)
                    continue
                if ref["kind"] == "prefix" and (
                    self.caller_class(ref["root"], ref["file"]) != "ci" or not self.is_test(self.rows[key])
                ):
                    continue  # a folder in CI is a test discovery: it calls the tests under it only
                callers[key].add(self.caller_class(ref["root"], ref["file"]))
        # a library or fixture is often sourced by name from its own folder, without a support path
        for key, row in self.rows.items():
            if row["kind"] == "file" and self.is_lib(row):
                name = os.path.basename(row["source"])
                for other, other_row in self.rows.items():
                    if (other != key and other[0] == key[0] and other_row["kind"] == "file"
                            and os.path.dirname(other[1]) == os.path.dirname(row["source"])
                            and name in (self.content(*other) or "")):
                        cites[other].add(key)
                        cited_by[key].add(other)
        self.callers, self.cites, self.cited_by = callers, cites, cited_by
        self.dependents = []
        tests, libs = [], []
        for key in sorted(self.rows):
            row = self.rows[key]
            if row["kind"] == "folder":
                self.place_folder(key, row)
            elif self.infra_tree(row):
                self.place_file(key, row)
            elif self.is_test(row):
                tests.append(key)
            elif self.is_lib(row):
                libs.append(key)
            else:
                self.place_file(key, row)
        self.dependents = tests + libs
        for key in self.dependents:
            self.place_dependent(key, self.rows[key])

    @staticmethod
    def is_test(row):
        return bool(TEST_NAME.match(os.path.basename(row["source"])))

    @staticmethod
    def is_lib(row):
        name = os.path.basename(row["source"]).lower()
        return ".lib." in name or "fixture" in name

    def infra_tree(self, row):
        entry = row["entry"]
        if entry == "scripts" or row["kind"] != "file":
            return False
        root = self.roots[row["root"]]
        return entry in cl.IAC_FAMILY or any(r.endswith(".tf") for r in root.files if r.startswith(f"_support/{entry}/"))

    def subject_of(self, key):
        row = self.rows[key]
        match = TEST_NAME.match(os.path.basename(row["source"]))
        if not match:
            return None
        name = match.group("stem2") or match.group("stem") + match.group("ext")
        subject = (row["root"], os.path.join(os.path.dirname(row["source"]), name).replace(os.sep, "/"))
        return subject if subject in self.rows else None

    def place_dependent(self, key, row):
        subject = self.subject_of(key) if self.is_test(row) else None
        if subject and not self.is_lib(self.rows[subject]):
            dest = self.rows[subject]["destination"]
            if self.is_kept(self.rows[subject]):
                return self.set_kept(key, "R1", follows=True)
            if dest:
                return self.set_dest(key, "R1", os.path.join(os.path.dirname(dest), os.path.basename(row["source"])))
            return self.set_review(key, "R1", "subject has no destination")
        related = self.cites.get(key, set()) if self.is_test(row) else self.cited_by.get(key, set())
        dirs = collections.Counter()
        for other in sorted(related):
            dest = self.rows[other]["destination"] if other in self.rows else None
            if dest and other[0] == key[0]:
                dirs[os.path.dirname(dest)] += 1
        if dirs:
            top = max(dirs.values())
            directory = sorted(d for d, n in dirs.items() if n == top)[0]
            return self.set_dest(key, "R3", os.path.join(directory, os.path.basename(row["source"])))
        if related and all(other in self.rows and self.is_kept(self.rows[other]) for other in related):
            return self.set_kept(key, "R3", follows=True)  # what only kept files use stays with them
        return self.place_file(key, row)

    def place_file(self, key, row):
        if not self.infra_tree(row):
            return self.place_code(key, row)
        i = row["root"]
        root = self.roots[i]
        homes = self.homes(i)
        if not root.is_git:
            return self.set_review(key, "R2", "no repository: infrastructure goes in a repository")
        if os.path.splitext(row["source"])[1].lower() in {".json", ".yaml", ".yml"} and not self.cited_anywhere(key):
            return self.set_review(key, "R2", "unreferenced: no tracked file cites it by path or name")
        if row["entry"] in IAC_HOMES:
            return self.set_dest(key, "R2", os.path.join(homes["infra"], row["subpath"]))
        return self.set_dest(key, "R2", os.path.join(homes["infra"], row["entry"], row["subpath"]))

    def place_code(self, key, row):
        i = row["root"]
        root = self.roots[i]
        homes = self.homes(i)
        sub = row["subpath"]
        classes = self.callers.get(key, set())
        code = {c for c in classes if c != "doc"}
        packages = sorted({c[len("package:"):] for c in code if c.startswith("package:")})
        first = sub.split("/")[0] if "/" in sub else None
        by_name = [p for p in root.packages if os.path.basename(p) == first] if first else []
        if not root.is_git:
            if not classes and self.one_off(row):
                return self.set_dest(key, "R9", self.session_dest(row))
            return self.set_review(key, "R4-R8", "no repository: code goes in the repository that uses it")
        if code and all(c.startswith("package:") for c in code) and len(packages) == 1:
            package = packages[0]
            rest = sub.split("/", 1)[1] if first and os.path.basename(package) == first else sub
            return self.set_dest(key, "R4", os.path.join(package, self.package_dir(i, package), rest))
        if not packages and len(by_name) == 1:
            package = by_name[0]
            return self.set_dest(key, "R4", os.path.join(package, self.package_dir(i, package), sub.split("/", 1)[1]))
        if len(by_name) > 1:
            self.rows[key]["notes"].append("ambiguous package")
        if "ci" in code:
            return self.set_dest(key, "R5", os.path.join(homes["tools"], homes["groups"]["ci"], sub))
        if "hook" in code:
            return self.set_dest(key, "R6", os.path.join(homes["tools"], homes["groups"]["hooks"], sub))
        if not code and "doc" in classes:
            return self.set_dest(key, "R7", os.path.join(homes["tools"], homes["groups"]["ops"], sub))
        if code:
            return self.set_dest(key, "R8", os.path.join(homes["tools"], sub))
        if self.one_off(row):
            return self.set_dest(key, "R9", self.session_dest(row))
        return self.set_review(key, "R10", "no rule: nothing cites it and its name is not one-off; keep or delete it")

    def package_dir(self, i, package):
        return self.roots[i].first_existing(package, PACKAGE_TOOL_DIRS) or "scripts"

    @staticmethod
    def one_off(row):
        name = os.path.basename(row["source"])
        if set(re.split(r"[-_.]", name.lower())) & ONE_OFF_WORDS:
            return True
        return any(m.group(1).lower() not in NOT_KEYS for m in KEY_RE.finditer(name))

    def session_dest(self, row):
        root = self.roots[row["root"]]
        date = root.first_commit_date(row["source"]) if row["tracked"] else None
        date = date or oldest_mtime(root.abs(row["source"]))
        stem = kebab(os.path.splitext(os.path.basename(row["source"]))[0])
        return os.path.join("_support/workspace", f"{date}-{stem}", os.path.basename(row["source"]))

    def corpus(self):
        if not hasattr(self, "_corpus"):
            self._corpus = [
                ((j, rel), self.content(j, rel) or "")
                for j in range(len(self.roots)) for rel in self.readable_files(j)
            ]
        return self._corpus

    def cited_anywhere(self, key):
        rel = key[1]
        name = os.path.basename(rel)
        return any(other != key and (rel in text or name in text) for other, text in self.corpus())

    def place_folder(self, key, row):
        i = row["root"]
        root = self.roots[i]
        homes = self.homes(i)
        name = row["entry"]
        base = row["source"]
        if name in LEGACY_EVIDENCE + LEGACY_DATA + LEGACY_RECORDS:
            if base == name:
                citing = self.top_citations(i, name)
                if citing:
                    return self.set_review(key, "F2", f"living path: a document or configuration cites {name}/: " + ", ".join(citing[:5]))
            citing = self.runtime_citations(root.abs(base))
            if citing:
                return self.set_review(key, "F2", "runtime path: code or configuration cites it: " + ", ".join(citing[:5]))
        if name in LEGACY_EVIDENCE + LEGACY_RECORDS:
            del self.rows[key]
            return self.place_records(i, base, name, "evidence" if name in LEGACY_EVIDENCE else "")
        if name in LEGACY_DATA:
            self.set_legacy_dest(key, "F4", f"_support/data/{self.record_date(i, base)}-{kebab(name)}")
            return self.review_current_knowledge(i, base, name)
        files = [r for r in root.files if r.startswith(base + "/")]
        if files and all(os.path.splitext(r)[1].lower() in DOC_EXTENSIONS for r in files):
            return self.set_dest(key, "F5", os.path.join(homes["docs"], name))
        return self.set_review(key, "F6", "no rule for this folder; classify it by hand")

    def runtime_citations(self, folder):
        """Non-comment lines of code or configuration whose support path resolves into folder."""
        found = []
        needle = "_support/" + os.path.basename(folder)
        for (j, rel), text in self.corpus():
            extension = os.path.splitext(rel)[1].lower()
            if extension not in cl.CODE_EXTENSIONS and extension not in RUNTIME_CONFIG_EXTENSIONS:
                continue
            if needle not in text:
                continue
            comment = cl.COMMENT_BY_EXTENSION.get(extension)
            for number, line in enumerate(text.splitlines(), 1):
                if needle not in line or (comment and comment.match(line)):
                    continue
                hits = [t for _, pre, path in tokens(line, False) for _, t in self.resolutions(j, rel, pre, path)]
                if any(t == folder or t.startswith(folder + os.sep) for t in hits):
                    found.append(f"{os.path.basename(self.roots[j].path)}/{rel}:{number}")
                    break
        return sorted(found)

    def top_citations(self, i, name):
        """Lines of other files, in any root, that cite the root's top-level folder as a path (`name/...`,
        also relative to the citing file); the files under the folder and history records do not count."""
        folder = self.roots[i].abs(name)
        pattern = re.compile(r"(?<![\w@.-])(?<!_support/)((?:\.{1,2}/)*" + re.escape(name) + r"(?:/[\w.@+*-]*)*)(?![\w-]|\.\w)")
        found = []
        for (j, rel), text in self.corpus():
            if j == i and (rel == name or rel.startswith(name + "/")):
                continue
            if not (rel.endswith(MARKDOWN) or os.path.splitext(rel)[1].lower() in cl.CODE_EXTENSIONS | RUNTIME_CONFIG_EXTENSIONS):
                continue
            if name not in text:
                continue
            extension = os.path.splitext(rel)[1].lower()
            prose = rel.endswith(MARKDOWN)  # a plain word in prose is no path; in code or configuration it is a value
            comment = cl.COMMENT_BY_EXTENSION.get(extension)
            citing_dir = os.path.dirname(self.roots[j].abs(rel))
            for number, line in enumerate(text.splitlines(), 1):
                if name not in line or (comment and comment.match(line)):
                    continue
                for match in pattern.finditer(line):
                    cited = match.group(1)
                    before = line[:match.start(1)]
                    if prose and not (cited.startswith(".") or cited.endswith("/") or "/" in cited[len(name):]
                                      or before.endswith(("`", "]("))):
                        continue  # a plain word, not a path
                    path = cited.rstrip("/.")
                    targets = {os.path.normpath(os.path.join(citing_dir, path)), os.path.normpath(os.path.join(self.roots[j].path, path))}
                    if any(t == folder or t.startswith(folder + os.sep) for t in targets):
                        found.append(f"{os.path.basename(self.roots[j].path)}/{rel}:{number}")
                        break
                else:
                    continue
                break
        return sorted(found)

    def record_date(self, i, rel):
        """The date a record carries in its name, else its first commit, else its oldest file."""
        match = DATE_RE.search(os.path.basename(rel))
        if match:
            return match.group(0)
        root = self.roots[i]
        tracked = any(r == rel or r.startswith(rel + "/") for r in root.tracked)
        return (root.first_commit_date(rel) if tracked else None) or oldest_mtime(root.abs(rel))

    def set_legacy_dest(self, key, rule, dest):
        """Place a legacy row; a tracked one leaves Git when it moves (the project's delivery commit removes it)."""
        self.set_dest(key, rule, dest)
        row = self.rows[key]
        row["untrack"] = bool(row["tracked"] and leaves_git_home(row["destination"]))

    def set_delete(self, key, rule, reason):
        row = self.rows[key]
        if self.keep_for(row):
            return self.set_kept(key, rule)
        override = self.override_for(row)
        row["rule"] = rule
        if override:
            row["destination"], row["override"] = override, True
            row["untrack"] = bool(row["tracked"] and leaves_git_home(override))
            return
        row["delete"], row["destination"], row["delete_note"] = True, None, reason
        row["notes"].append(reason)

    def review_current_knowledge(self, i, base, name):
        if name in ("plan", "spec"):
            self.review(i, base, "owner uncertain", f"{name}/ may hold current knowledge: promote it to the docs home, or keep it with --keep")

    def place_records(self, i, base, name, sub):
        """Each child of a legacy home becomes a work folder, dated by its name, else its first commit,
        else its oldest file; what regenerates is deleted. sub is the folder kept inside the work folder."""
        root = self.roots[i]
        home = self.support_home(i, base)
        loose = []
        for child in sorted(os.listdir(root.abs(base))):
            child_rel = f"{base}/{child}"
            key = (i, child_rel)
            folder = os.path.isdir(root.abs(child_rel))
            if self.embedded_repository(i, child_rel):
                continue
            if self.regenerable(i, child_rel, child, name):
                self.rows[key] = self.new_row(i, child_rel, "folder" if folder else "file", name)
                self.set_delete(key, "L3", "reproducible: tooling regenerates it")
            elif folder:
                work = kebab(DATE_RE.sub("", child, count=1)) or kebab(name)
                self.rows[key] = self.new_row(i, child_rel, "folder", name)
                folder_name = f"{self.record_date(i, child_rel)}-{work}"
                self.place_record(key, name, os.path.join(home, "workspace", folder_name, sub),
                                  os.path.join(home, "data", folder_name, sub))
            else:
                loose.append(child_rel)
        if loose:
            date = min(self.record_date(i, r) for r in loose)
            for rel in loose:
                key = (i, rel)
                self.rows[key] = self.new_row(i, rel, "file", name)
                folder_name = f"{date}-legacy-{kebab(name)}"
                self.place_record(key, name, os.path.join(home, "workspace", folder_name, sub, os.path.basename(rel)),
                                  os.path.join(home, "data", folder_name, sub, os.path.basename(rel)))
        self.review_current_knowledge(i, base, name)

    def regenerable(self, i, rel, child, home):
        """Whether the child is tooling output safe to delete by name: untracked, ignored by Git,
        and not under evidence/ or reports/, where a name alone says nothing about what it holds."""
        if child not in REPRODUCIBLE and child not in cl.IGNORED_NAMES:
            return False
        root = self.roots[i]
        if home in LEGACY_EVIDENCE or not root.is_git:
            return False
        tracked = rel in root.tracked or any(r.startswith(rel + "/") for r in root.tracked)
        return not tracked and rel in root.ignored([rel])

    def place_record(self, key, home, dest, data_dest):
        """A child of plan/ or spec/ may hold current knowledge: it waits for --keep or --override. With
        --legacy remove, a tracked child with no uncommitted or hidden edit is deleted (history and the
        archive tag hold all of it) unless it holds a dump or export, which goes to data/; any other
        child still moves."""
        row = self.rows[key]
        if home in ("plan", "spec") and not (self.keep_for(row) or self.override_for(row)):
            return self.set_review(key, "F3", f"{home}/ may hold current knowledge: promote it, or name --keep or --override")
        if self.legacy == "remove" and home not in ("plan", "spec"):
            if self.holds_data(row["root"], row["source"]):
                self.set_legacy_dest(key, "F4", data_dest)
                if row["destination"] == os.path.normpath(data_dest).replace(os.sep, "/"):
                    row["notes"].append("holds a dump or export: its removal is a decision of its own")
                return
            if self.unmodified(row["root"], row["source"]) is None:
                return self.set_delete(key, "F3", "aged record: history and the archive tag keep it")
        self.set_legacy_dest(key, "F3", dest)

    def holds_data(self, i, rel):
        """Whether the record is, or holds, a dump, backup or export: removing data is a decision of its own.
        Every dot-separated suffix of a file name counts, so `.sql.gz` and `.bak.old` are found."""
        base = self.roots[i].abs(rel)
        names = [os.path.basename(base)] if os.path.isfile(base) else [
            name for _, _, files in os.walk(base) for name in files]
        for name in names:
            parts = name.lower().split(".")
            suffixes = {"." + part for part in parts[1:]}
            if suffixes & DATA_SUFFIXES or (suffixes & DATA_NAMED_SUFFIXES and DATA_NAME_RE.search(parts[0])):
                return True
        return False

    def text_files(self):
        """(root index, relative path, text) of every text file of every root, ignored folders included, but
        for .git, node_modules, binary files and ignore lists: where a record's name may be written."""
        if not hasattr(self, "_text_files"):
            self._text_files, seen = [], set()
            for j, root in enumerate(self.roots):
                for folder, folders, files in os.walk(root.path):
                    folders[:] = [d for d in folders if d not in (".git", "node_modules")]
                    for name in files:
                        path = os.path.join(folder, name)
                        if path in seen or name.endswith("ignore") or os.path.islink(path):
                            continue
                        seen.add(path)
                        try:
                            if os.path.getsize(path) > TEXT_FILE_LIMIT:
                                continue
                            with open(path, "rb") as handle:
                                data = handle.read()
                        except OSError:
                            continue
                        if b"\0" in data[:8192]:
                            continue
                        self._text_files.append((j, os.path.relpath(path, root.path).replace(os.sep, "/"),
                                                 data.decode("utf-8", "replace")))
        return self._text_files

    def review_cited_deletions(self):
        """A delete row whose name occurs in a text file that stays goes to review with the citing line:
        a substring match, so a false hit only costs a review. The files inside rows this plan deletes do
        not count; a row sent to review stops excluding its own files, so the check repeats to a fixed point."""
        deleting = {key for key, row in self.rows.items() if row["delete"]}
        changed = bool(deleting)
        while changed:
            changed = False
            inside = [self.roots[i].abs(rel) for i, rel in deleting]
            cited = {}
            for j, other, text in self.text_files():
                path = self.roots[j].abs(other)
                if any(path == p or path.startswith(p + os.sep) for p in inside):
                    continue
                for key in deleting:
                    if key in cited:
                        continue
                    name = os.path.basename(key[1])
                    for needle in (name, urllib.parse.quote(name)):
                        at = text.find(needle)
                        if at >= 0:
                            line = text.count("\n", 0, at) + 1
                            cited[key] = f"{os.path.basename(self.roots[j].path)}/{other}:{line}"
                            break
            for key, where in sorted(cited.items()):
                row = self.rows[key]
                row["delete"] = False
                row["notes"] = [n for n in row["notes"] if n != row.get("delete_note")]
                self.set_review(key, row["rule"], f"living path: its name occurs in {where}")
                deleting.discard(key)
                changed = True

    # ---- the change register of an earlier layout ---------------------------------
    def declared(self, i, key):
        for name in ("AGENTS.md", "CLAUDE.md"):
            match = re.search(rf"^\s*[-*]?\s*{key}:\s*`?([^\s`·]+)", cl.read_text(self.roots[i].abs(name)), re.M)
            if match:
                return match.group(1)
        return None

    def specs_homes(self, i):
        root = self.roots[i]
        names = [cl.norm_path(self.declared(i, "Specs") or ""), "_support/openspec", "openspec"]
        found = []
        for name in names:
            if (name and not name.startswith(("..", "/")) and name not in found and os.path.isdir(root.abs(name + "/changes"))
                    and not self.in_other_repository(i, name)):
                found.append(name)
        return found

    def embedded_repository(self, i, rel):
        """Whether rel is not this root's to plan: a repository of its own, which a move would carry off with
        its history, is reviewed; a path another root of the plan covers is that root's, and is skipped."""
        if not self.in_other_repository(i, rel):
            return False
        absolute = self.roots[i].abs(rel)
        covered = any(k != i and len(other.path) > len(self.roots[i].path)
                      and (absolute == other.path or absolute.startswith(other.path + os.sep))
                      for k, other in enumerate(self.roots))
        if not covered and os.path.exists(os.path.join(absolute, ".git")):
            self.review(i, rel, "embedded repository",
                        "it holds its own .git: plan it as a root of its own, or move it by hand")
        return True

    def in_other_repository(self, i, rel):
        """Whether rel lies inside a repository of its own below root i, or inside another root of the plan,
        which plans it: a workspace folder outside Git never judges what its child repositories hold."""
        root = self.roots[i]
        absolute = root.abs(rel)
        if any(k != i and len(other.path) > len(root.path) and (absolute == other.path or absolute.startswith(other.path + os.sep))
               for k, other in enumerate(self.roots)):
            return True
        parent = absolute
        while parent != root.path and len(parent) > len(root.path):
            if os.path.exists(os.path.join(parent, ".git")):
                return True
            parent = os.path.dirname(parent)
        return False

    def base_ref(self, i):
        """origin/<base> from the declared Base branch, else the remote's default; None when neither resolves."""
        root = self.roots[i]
        if not root.is_git:
            return None
        base = self.declared(i, "Base branch")
        if not base:
            head = git(root.path, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
            base = head.stdout.strip().split("/", 1)[-1] if head is not None and head.returncode == 0 else None
        if not base:
            return None
        found = git(root.path, "rev-parse", "--verify", "-q", f"refs/remotes/origin/{base}")
        return f"origin/{base}" if found is not None and found.returncode == 0 else f"!{base}"

    def legacy_changes(self):
        """The change archive and each change folder an earlier layout kept: deleted when their last
        commit is on the remote base branch, kept and named otherwise."""
        for i, root in enumerate(self.roots):
            ref = self.base_ref(i)
            for home in self.specs_homes(i):
                changes = f"{home}/changes"
                for child in sorted(os.listdir(root.abs(changes))):
                    rel = f"{changes}/{child}"
                    folder = os.path.isdir(root.abs(rel))
                    if child == "archive" and folder and root.is_git:
                        self.archive_children(i, rel)
                        continue
                    key = (i, rel)
                    self.rows[key] = self.new_row(i, rel, "folder" if folder else "file", "changes")
                    if child == "archive" or not folder:
                        self.judge_record(key, i, rel, folder)
                        continue
                    verdict = self.integration(i, rel, ref)
                    if verdict is None:
                        self.set_delete(key, "L1", f"integrated: its last commit is on {ref}")
                    else:
                        self.set_keep(key, *verdict)

    def archive_children(self, i, archive):
        """One row per child of the change archive, so a child with an untracked, ignored or edited file
        stays without keeping the others: the manifest lists each archived change on its own."""
        root = self.roots[i]
        for child in sorted(os.listdir(root.abs(archive))):
            rel = f"{archive}/{child}"
            if self.embedded_repository(i, rel):
                continue
            folder = os.path.isdir(root.abs(rel))
            key = (i, rel)
            self.rows[key] = self.new_row(i, rel, "folder" if folder else "file", "changes")
            self.judge_record(key, i, rel, folder)

    def judge_record(self, key, i, rel, folder):
        """The change archive (or one of its children) and a loose file of changes/: deleted when tracked
        and clean, kept with the reason otherwise."""
        verdict = self.unmodified(i, rel)
        if verdict is None:
            self.set_delete(key, "L1", "change archive: a record of past work" if folder else "index of the change register")
        else:
            self.set_keep(key, *verdict)

    def hides_edits(self, i, rel):
        """Whether an index entry under rel carries assume-unchanged or skip-worktree, which hide edits
        from git status; None when Git cannot list them."""
        flags = git(self.roots[i].path, "ls-files", "--debug", "--", rel)
        if flags is None or flags.returncode != 0:
            return None
        return any(int(f, 16) & HIDING_FLAGS for f in re.findall(r"flags: ([0-9a-f]+)", flags.stdout))

    def unmodified(self, i, rel):
        """None when rel is tracked and has no uncommitted change, so history holds all of it; else (reason, detail)."""
        root = self.roots[i]
        if not root.is_git:
            return "owner uncertain", "no repository: no history holds it"
        if not (rel in root.tracked or any(r.startswith(rel + "/") for r in root.tracked)):
            return "owner uncertain", "it is not tracked: no history holds it"
        dirty = git(root.path, *STATUS_ALL, "--", rel)
        if dirty is None or dirty.returncode != 0:
            return "owner uncertain", "git status failed: cannot tell whether it has uncommitted changes"
        if dirty.stdout.strip():
            return "owner uncertain", "it has uncommitted or untracked changes: commit or discard them first"
        hidden = self.hides_edits(i, rel)
        if hidden is None or hidden:
            return "owner uncertain", "an index entry is assume-unchanged or skip-worktree, which hides edits from git status"
        return None

    def integration(self, i, rel, ref):
        """None when the folder's last commit is reachable from ref, else (reason, detail)."""
        root = self.roots[i]
        if not root.is_git:
            return "owner uncertain", "no repository: cannot tell whether this change landed"
        if ref is None or ref.startswith("!"):
            base = "its base branch" if ref is None else f"origin/{ref[1:]}"
            return "owner uncertain", f"{base} does not resolve here: cannot tell whether this change landed"
        last = git(root.path, "log", "-1", "--format=%H", "--", rel)
        sha = last.stdout.strip() if last is not None and last.returncode == 0 else ""
        if not sha:
            return "owner uncertain", "no commit holds it: nothing shows it landed"
        dirty = git(root.path, *STATUS_ALL, "--", rel)
        if dirty is None or dirty.returncode != 0:
            return "owner uncertain", "git status failed"
        if dirty.stdout.strip():
            return "owner uncertain", "it has uncommitted changes: commit or discard them first"
        hidden = self.hides_edits(i, rel)
        if hidden is None or hidden:
            return "owner uncertain", "an index entry is assume-unchanged or skip-worktree, which hides edits from git status"
        reachable = git(root.path, "merge-base", "--is-ancestor", sha, ref)
        if reachable is not None and reachable.returncode == 0:
            return None
        return "not integrated", f"its last commit {sha[:8]} is not reachable from {ref}"

    def set_keep(self, key, reason, detail):
        row = self.rows[key]
        if self.keep_for(row):
            return self.set_kept(key, "L1")
        override = self.override_for(row)
        if override:
            row["rule"], row["destination"], row["override"] = "L1", override, True
            row["untrack"] = bool(row["tracked"] and leaves_git_home(override))
            return
        row["rule"] = "L1"
        row["notes"].append(f"kept: {reason}")
        self.review(row["root"], row["source"], reason, detail)

    def report_unignored(self):
        """Moved legacy rows land in _support/workspace or _support/data: Git must ignore them. A ../
        destination lands in the workspace folder's own _support/, checked in that folder's repository."""
        for i, root in enumerate(self.roots):
            moved = [r["destination"] for k, r in self.rows.items() if k[0] == i and r["untrack"] and r["destination"]]
            tops = {"/".join(d.split("/")[:2]) for d in moved if not d.startswith("../")}
            for top in sorted(tops - root.ignored(sorted(tops))):
                self.review(i, top, "workspace not ignored",
                            f"add {top}/ to .gitignore before applying, or the moved files show up as untracked")
            outside = {"/".join(d.split("/")[1:3]) for d in moved
                       if any(d.startswith("../" + home + "/") for home in IGNORED_HOMES)}
            for k, parent in enumerate(self.roots):
                if k == i or parent.path != os.path.dirname(root.path) or not parent.is_git:
                    continue
                for top in sorted(outside - parent.ignored(sorted(outside))):
                    self.review(k, top, "workspace not ignored",
                                f"add {top}/ to .gitignore before applying, or the moved files show up as untracked")

    def override_for(self, row):
        root = self.roots[row["root"]]
        for spec, dest in self.overrides:
            if spec == row["source"] or os.path.abspath(spec) == root.abs(row["source"]):
                self.used_overrides.add(spec)
                return os.path.normpath(dest).replace(os.sep, "/")
        return None

    def keep_for(self, row):
        """Whether the user keeps this row where it is: its source, or a folder holding it."""
        return self.kept_path(row["root"], row["source"])

    def kept_path(self, i, rel):
        root = self.roots[i]
        source = root.abs(rel)
        for spec in self.keeps:
            path = os.path.abspath(spec) if os.path.isabs(spec) else root.abs(os.path.normpath(spec))
            if source == path or source.startswith(path.rstrip(os.sep) + os.sep):
                self.used_keeps.add(spec)
                return True
        return False

    def kept_ref(self, i, rel, line, text):
        """Whether the user keeps the reference on this line as it is (--keep-ref FILE[:LINE]).
        A kept line is recorded by its text, so lines added above it later do not lose it."""
        source = self.roots[i].abs(rel)
        for spec, path, wanted in self.keep_refs:
            if source == (path if os.path.isabs(path) else self.roots[i].abs(path)) and wanted in (None, line):
                self.used_keep_refs.add(spec)
                self.kept_refs.add((i, self.new_location(i, rel), line_hash(text) if wanted else None))
                return True
        return False

    def set_kept(self, key, rule, follows=False):
        """Keep a row in place: one the user kept, or with follows, one that goes where
        kept files are (the test of a kept file, what only kept files use), which an
        override still moves."""
        row = self.rows[key]
        override = self.override_for(row)
        if override and not follows:
            raise SystemExit(fail(f"--keep and --override both name {row['source']}: choose one"))
        if override:
            row["rule"], row["destination"], row["override"] = rule, override, True
            row["untrack"] = bool(row["tracked"] and leaves_git_home(override))
            return
        row["rule"], row["destination"], row["override"] = rule, None, False
        row["notes"].append(KEPT_WITH if follows else KEPT)

    @staticmethod
    def is_kept(row, by_user=False):
        return KEPT in row["notes"] or (not by_user and KEPT_WITH in row["notes"])

    def set_dest(self, key, rule, dest):
        row = self.rows[key]
        if self.keep_for(row):
            return self.set_kept(key, rule)
        override = self.override_for(row)
        row["rule"] = rule
        row["destination"] = override or os.path.normpath(dest).replace(os.sep, "/")
        row["override"] = bool(override)
        if override:
            row["untrack"] = bool(row["tracked"] and leaves_git_home(override))

    def set_review(self, key, rule, reason):
        row = self.rows[key]
        if self.keep_for(row):
            return self.set_kept(key, rule)
        override = self.override_for(row)
        row["rule"] = rule
        if override:
            row["destination"], row["override"] = override, True
            row["untrack"] = bool(row["tracked"] and leaves_git_home(override))
            return
        row["destination"] = None
        self.review(row["root"], row["source"], rule, reason)

    # ---- guards after classification ------------------------------------------------
    def guard(self):
        """Tracked rows going into ignored folders, then collisions: first for the rows
        placed on their own, then, once their subjects and users are settled, for the
        tests and libraries placed beside them. A dependent never displaces a row
        already settled: on a shared destination only the dependent goes to review."""
        dependents = set(self.dependents)
        settled = [k for k in sorted(self.rows) if k not in dependents]
        self.guard_ignored(settled)
        self.mark_collisions(settled, settled)
        for key in self.dependents:
            row = self.rows[key]
            if not row["override"] and row["rule"] in ("R1", "R3"):
                self.place_dependent(key, row)
        self.guard_ignored(self.dependents)
        self.mark_collisions(sorted(self.rows), self.dependents)

    def guard_ignored(self, keys):
        for i, root in enumerate(self.roots):
            tracked = {self.rows[k]["destination"]: k for k in keys
                       if self.rows[k]["root"] == i and self.rows[k]["tracked"] and not self.rows[k]["untrack"]
                       and self.rows[k]["destination"] and not self.rows[k]["override"]}
            for dest in sorted(root.ignored(sorted(tracked))):
                self.set_review(tracked[dest], self.rows[tracked[dest]]["rule"],
                                f"tracked: {dest} is ignored, so the move would drop it from Git; "
                                "remove it from the tree with the archive tag instead")

    def guard_builds(self):
        """A move that changes what a build or a published package takes: each moving row
        whose source and destination differ in an exclusion file's verdict keeps its
        destination, records that verdict for verify, and the plan asks about the new
        folders, one question per exclusion file and direction."""
        for i, root in enumerate(self.roots):
            ignore_files = [rel for rel in root.files if is_build_ignore(rel)]
            # a ../ destination leaves the repository, and with it every build the exclusion files shape
            moving = [k for k in sorted(self.rows) if k[0] == i and self.rows[k]["destination"]
                      and not self.rows[k]["destination"].startswith("../")]
            for ignore_file in ignore_files:
                paths = {}
                for key in moving:
                    row = self.rows[key]
                    suffix = "/" if row["kind"] != "file" else ""
                    paths[key] = (row["source"] + suffix, row["destination"] + suffix)
                status = build_status(root.path, ignore_file, sorted({p for pair in paths.values() for p in pair}))
                changed = collections.defaultdict(set)
                for key, (source, dest) in paths.items():
                    if status[source] == status[dest]:
                        continue
                    row = self.rows[key]
                    new = self.new_folder(root, row["destination"], row["kind"] != "file")
                    row.setdefault("builds", []).append({"file": ignore_file, "excluded": status[source], "new_folder": new})
                    changed[status[dest]].add((new or row["destination"]) + ("/" if new else ""))
                if changed[False]:
                    self.review(i, ignore_file, "build contents",
                                "now included: " + ", ".join(sorted(changed[False]))
                                + f"; add each one the build must not take to {ignore_file}, or name it with ! to keep it")
                if changed[True]:
                    self.review(i, ignore_file, "build contents",
                                "now excluded: " + ", ".join(sorted(changed[True]))
                                + f"; the build loses them: change {ignore_file}, or name each one in it to keep the change")

    def staying(self, i, root, entry):
        """The part of an ignored entry the plan does not move. A folder Git lists as one entry
        is opened when the plan moves a file inside it, so only the files that stay are named."""
        path = entry.rstrip("/")
        if self.moves(i, path):
            return []
        if not entry.endswith("/") or not any(k[0] == i and r["destination"] and k[1].startswith(path + "/")
                                              for k, r in self.rows.items()):
            return [entry]
        result = git(root.path, "ls-files", "-z", "--others", "--ignored", "--exclude-standard", "--", path)
        files = [f for f in result.stdout.split("\0") if f] if result is not None and result.returncode == 0 else []
        return [f for f in files if not self.moves(i, f)]

    def moves(self, i, rel):
        """Whether the plan moves rel, itself or with a folder that holds it."""
        while rel:
            row = self.rows.get((i, rel))
            if row and row["destination"]:
                return True
            rel = os.path.dirname(rel)
        return False

    @staticmethod
    def new_folder(root, dest, is_folder):
        """The highest folder of dest that does not exist yet, which the move creates."""
        parts = dest.split("/")
        for n in range(1, len(parts) + (1 if is_folder else 0)):
            candidate = "/".join(parts[:n])
            if not os.path.lexists(root.abs(candidate)):
                return candidate
        return None

    def report_left_ignored(self):
        """Files Git ignores stay behind when the move empties their folder of tracked files."""
        for i, root in enumerate(self.roots):
            if not root.is_git:
                continue
            moving = {r["source"] for k, r in self.rows.items()
                      if k[0] == i and r["destination"] and r["tracked"] and r["kind"] == "file"}
            folders = set()
            for source in moving:
                parent = os.path.dirname(source)
                while parent and parent != "_support":
                    folders.add(parent)
                    parent = os.path.dirname(parent)
            holding = set()  # folders that keep a tracked file the plan does not move
            for tracked in root.tracked:
                if tracked not in moving:
                    parent = os.path.dirname(tracked)
                    while parent and parent not in holding:
                        holding.add(parent)
                        parent = os.path.dirname(parent)
            emptied = folders - holding
            for folder in sorted(f for f in emptied if not any(f.startswith(e + "/") for e in emptied)):
                left = [p for entry in left_ignored(root.path, folder) for p in self.staying(i, root, entry)]
                if left:
                    self.review(i, folder, "ignored files stay",
                                f"Git ignores {plural(len(left), 'entry', 'entries')} that stay at {folder} after the move, "
                                f"such as {', '.join(left[:3])}: move or delete them by hand")

    def mark_collisions(self, claimants, movable):
        """Send to review the claimants of a destination that exists or is shared. On a
        shared one, a single row with the user's --override keeps it and the others go;
        otherwise only the rows of movable go, and every row when none is movable."""
        movable = set(movable)
        seen = collections.defaultdict(list)
        for key in claimants:
            row = self.rows[key]
            if row["destination"]:
                seen[fold(self.roots[row["root"]].abs(row["destination"]))].append(key)
        for _, keys in sorted(seen.items()):
            exists = os.path.lexists(self.roots[self.rows[keys[0]]["root"]].abs(self.rows[keys[0]]["destination"]))
            if not exists and len(keys) == 1:
                continue
            chosen = [k for k in keys if self.rows[k]["override"]]
            if not exists and len(chosen) == 1:
                losers = [k for k in keys if k not in chosen]
            elif chosen and not exists:
                losers = keys
            else:
                losers = [k for k in keys if k in movable] or keys
            for key in losers:
                row = self.rows[key]
                shown = row["destination"]
                row["notes"].append("collision")
                row["override"] = False
                row["destination"] = None
                self.review(row["root"], row["source"], row["rule"], f"collision at {shown}")
        for key in claimants:
            row = self.rows[key]
            if row["destination"] and self.nested_in_other(key):
                dest = row["destination"]
                row["notes"].append("collision")
                row["override"] = False
                row["destination"] = None
                self.review(row["root"], row["source"], row["rule"],
                            f"nested destination: {dest} lies inside the destination of another folder; "
                            "move one of them with --override, or keep it with --keep")

    def nested_in_other(self, key):
        """Whether the row's destination lies inside the destination of a folder row that moves elsewhere,
        unless that folder moves in the same plan and holds nothing at the path (apply moves it first)."""
        row = self.rows[key]
        dest = self.roots[row["root"]].abs(row["destination"])
        for other_key, other in self.rows.items():
            if other_key == key or other["kind"] != "folder" or not other["destination"]:
                continue
            outer = self.roots[other["root"]]
            outer_dest = outer.abs(other["destination"])
            same_tree = (other["root"] == row["root"] and row["source"].startswith(other["source"] + "/"))
            if fold(dest).startswith(fold(outer_dest) + os.sep) and not same_tree:
                if not nested_collides(outer.abs(other["source"]), outer_dest, dest):
                    continue
                return True
        return False

    def mark_split_tests(self):
        """A kept test whose subject moves may run it by a path that breaks: ask."""
        for key in sorted(self.rows):
            row = self.rows[key]
            if not self.is_kept(row, by_user=True) or not self.is_test(row):
                continue
            subject = self.subject_of(key)
            if subject and self.rows[subject]["destination"]:
                self.review(row["root"], row["source"], "kept test",
                            f"its subject moves to {self.rows[subject]['destination']}; keep both or move both")

    def mark_same_names(self):
        for key in sorted(self.rows):
            row = self.rows[key]
            if row["kind"] != "file":
                continue
            root = self.roots[row["root"]]
            stem, ext = os.path.splitext(os.path.basename(row["source"]))
            for rel in sorted(root.tracked):
                if rel.startswith("_support/") or HISTORY_RE.search(rel):
                    continue
                other_stem, other_ext = os.path.splitext(os.path.basename(rel))
                if other_stem == stem and other_ext != ext:
                    row["notes"].append(f"same name elsewhere: {rel}")
                    break

    def apply_homes(self):
        if not self.homes_override:
            return
        for row in self.rows.values():
            rule = row["rule"] or ""
            category = "tools" if rule in ("R5", "R6", "R7", "R8") else "infra" if rule == "R2" else "docs" if rule == "F5" else None
            if category in self.homes_override and row["destination"]:
                row["override"] = True
        for i, root in enumerate(self.roots):
            merged = dict(root.layout)
            merged.update(self.homes_override)
            parts = [f"{k} at {merged[k]}" for k in ("docs", "tools", "infra") if k in merged]
            self.layout_lines.append({"root": i, "line": "Layout: " + " · ".join(parts)})

    # ---- rewrites ----------------------------------------------------------------
    def new_location(self, i, rel):
        """Where a file of root i will be after the plan: its row's destination, the
        destination of a moving folder that holds it, or where it is."""
        row = self.rows.get((i, rel))
        if row and row["destination"]:
            return row["destination"]
        parent = os.path.dirname(rel)
        while parent:
            row = self.rows.get((i, parent))
            if row and row["destination"] and row["kind"] == "folder":
                return os.path.join(row["destination"], os.path.relpath(rel, parent)).replace(os.sep, "/")
            parent = os.path.dirname(parent)
        return rel

    def new_abs(self, key):
        row = self.rows.get(key)
        if not row or not row["destination"]:
            return None
        return self.roots[row["root"]].abs(row["destination"])

    def build_rewrites(self, refs):
        done = set()
        for ref in sorted(refs, key=lambda r: (r["root"], r["file"], r["line"], r["col"])):
            i = ref["root"]
            root = self.roots[i]
            if self.kept_ref(i, ref["file"], ref["line"], ref["text"]):
                done.add((i, ref["file"], ref["line"], ref["col"]))
                continue
            if ref["kind"] == "literal":
                new_target = self.new_abs(ref["keys"][0])
                if new_target is None and ref["keys"][0] in self.rows and self.is_kept(self.rows[ref["keys"][0]]):
                    cited = ref["keys"][0]  # it stays, in its own root; the citing file may move
                    new_target = self.roots[cited[0]].abs(cited[1])
                if new_target is None:
                    continue
                if ref["inside"]:
                    new_target = os.path.join(new_target, ref["inside"])
            else:
                new_target = self.prefix_destination(ref["target"], ref["keys"])
                if new_target is None and all(not self.rows[k]["destination"] and self.is_kept(self.rows[k]) for k in ref["keys"]):
                    new_target = ref["target"]  # nothing under it moves; the citing file may
                if new_target is None:
                    dirs = sorted({os.path.dirname(self.rows[k]["destination"]) if self.rows[k]["destination"] else "review"
                                   for k in ref["keys"]})
                    self.review(i, f"{ref['file']}:{ref['line']}", "prefix",
                                f"{ref['token']} covers files with different destinations: " + ", ".join(dirs))
                    continue
            citing_dir = os.path.dirname(root.abs(self.new_location(i, ref["file"])))
            new_token = ref["pre"] + self.render(ref["path"], ref["base"], new_target, citing_dir, root)
            done.add((i, ref["file"], ref["line"], ref["col"]))
            if new_token != ref["token"]:
                self.add_rewrite(i, ref["file"], ref["line"], ref["col"], ref["text"], ref["token"], new_token)
        self.relocate_markdown(done)

    def add_rewrite(self, i, rel, line, col, text, old, new):
        self.rewrites.append({
            "root": i, "file": rel, "line": line, "col": col,
            "sha256_line": sha256_bytes(text.encode()), "old": old, "new": new,
        })

    def relocate_markdown(self, done):
        """A moving Markdown file keeps its relative links pointing where they did."""
        for i, root in enumerate(self.roots):
            for rel in root.files:
                if not rel.endswith(MARKDOWN) or NEVER_READ.search("/" + rel) or HISTORY_RE.search(rel):
                    continue  # a record of the past keeps the links it was written with
                new_rel = self.new_location(i, rel)
                if new_rel == rel:
                    continue
                text = self.content(i, rel)
                if text is None:
                    continue
                old_dir = os.path.dirname(root.abs(rel))
                new_dir = os.path.dirname(root.abs(new_rel))
                for number, line in enumerate(text.splitlines(True), 1):
                    for col, pre, path in tokens(line, True):
                        if pre or (i, rel, number, col) in done or not (path.startswith("./") or path.startswith("../")):
                            continue
                        if self.kept_ref(i, rel, number, line):
                            continue
                        target = os.path.normpath(os.path.join(old_dir, path.split("*")[0].rstrip("/")))
                        if not os.path.lexists(target):
                            continue
                        moved = self.new_location(i, os.path.relpath(target, root.path).replace(os.sep, "/")) \
                            if target.startswith(root.path + os.sep) else None
                        final = root.abs(moved) if moved else target
                        new_path = self.render(path, "file", final, new_dir, root)
                        if new_path != path:
                            self.add_rewrite(i, rel, number, col, line, path, new_path)

    def prefix_destination(self, prefix, keys):
        base = None
        for key in keys:
            row = self.rows[key]
            if not row["destination"]:
                return None
            source = self.roots[row["root"]].abs(row["source"])
            rest = os.path.relpath(source, prefix)
            dest = self.roots[row["root"]].abs(row["destination"])
            if not dest.endswith(os.sep + rest):
                return None
            candidate = dest[: len(dest) - len(rest) - 1]
            if base is None:
                base = candidate
            elif base != candidate:
                return None
        moving = {self.roots[self.rows[k]["root"]].abs(self.rows[k]["source"]) for k in keys}
        for item in self.listing(prefix):
            if not any(item == m or item.startswith(m + os.sep) for m in moving):
                return None
        return base

    def render(self, path, base, new_target, citing_dir, root):
        """The new path, in the form the old one had."""
        star = path.find("*")
        head = path[:star] if star >= 0 else path
        stripped = head.rstrip("/")
        suffix = path[len(stripped):]
        dot = "./" if stripped.startswith("./") else ""
        if base == "file":
            rel = os.path.relpath(new_target, citing_dir).replace(os.sep, "/")
            return (dot if not rel.startswith("../") else "") + rel + suffix
        if base.startswith("ws:"):
            workspace = os.path.dirname(self.roots[int(base[3:])].path)
            return os.path.relpath(new_target, workspace).replace(os.sep, "/") + suffix
        return dot + os.path.relpath(new_target, root.path).replace(os.sep, "/") + suffix

    def internal_paths(self):
        """A moving script that finds files from its own folder: literals whose target changes."""
        files, _ = self.candidate_index()
        for key in sorted(self.rows):
            row = self.rows[key]
            if row["kind"] != "file" or not row["destination"]:
                continue
            root = self.roots[row["root"]]
            text = self.content(row["root"], row["source"])
            if not text or not ANCHOR_RE.search(text):
                continue
            old_dir = os.path.dirname(root.abs(row["source"]))
            new_dir = os.path.dirname(root.abs(row["destination"]))
            for number, line in enumerate(text.splitlines(), 1):
                covered = [(col, col + len(pre) + len(path)) for col, pre, path in tokens(line, False)]
                for match in REL_LITERAL_RE.finditer(line):
                    if any(a <= match.start(1) < b for a, b in covered):
                        continue  # a support path token: the rewrites already carry it
                    literal = match.group(1)
                    glued = match.start(1) > 0 and line[match.start(1) - 1] == "/"
                    if "/" not in literal and not glued:
                        continue  # a bare `..`, such as a comparison or a string, is no path
                    before = os.path.normpath(os.path.join(old_dir, literal))
                    if not os.path.lexists(before):
                        continue  # it names nothing today, so it is no path to this tree
                    expected = self.new_abs(files[before]) if before in files else before
                    after = os.path.normpath(os.path.join(new_dir, literal))
                    if expected and after != expected:
                        self.review(row["root"], f"{row['source']}:{number}", "internal path",
                                    f"{literal} resolves to {os.path.relpath(before, root.path)} today and to "
                                    f"{os.path.relpath(after, root.path)} from {row['destination']}")

    def build_index_lines(self):
        for key in sorted(self.rows):
            row = self.rows[key]
            if row["rule"] in TOOL_RULES and row["destination"] and not self.is_test(row):
                self.index_lines.append({"root": row["root"], "line": f"- `{row['destination']}` — <what it is for> · run: <command> · test: <its test or none>"})

    # ---- output --------------------------------------------------------------------
    def plan(self):
        self.collect()
        refs = self.references()
        self.classify(refs)
        self.legacy_changes()
        self.review_cited_deletions()
        self.guard()
        self.apply_homes()
        unused = sorted({spec for spec, _ in self.overrides} - self.used_overrides)
        if unused:
            raise SystemExit(fail("--override names no planned row: " + ", ".join(unused)))
        unused = sorted(set(self.keeps) - self.used_keeps)
        if unused:
            raise SystemExit(fail("--keep names no planned row: " + ", ".join(unused)))
        self.guard_builds()
        self.report_unignored()
        self.report_left_ignored()
        self.mark_split_tests()
        self.mark_same_names()
        self.build_rewrites(refs)
        unused = sorted({spec for spec, _, _ in self.keep_refs} - self.used_keep_refs)
        if unused:
            raise SystemExit(fail("--keep-ref names no reference: " + ", ".join(unused)))
        self.internal_paths()
        self.build_index_lines()
        rows = []
        for key in sorted(self.rows):
            row = self.rows[key]
            action = ("delete" if row["delete"] else "move" if row["destination"]
                      else "keep" if any(n.startswith("kept") for n in row["notes"]) else "review")
            rows.append({
                "root": row["root"], "source": row["source"], "kind": row["kind"], "tracked": row["tracked"],
                "action": action, "untrack": row["untrack"],
                "rule": row["rule"], "destination": row["destination"], "notes": sorted(set(row["notes"])),
                "override": row["override"], "fingerprint": fingerprint(self.roots[row["root"]].abs(row["source"])),
                **({"builds": sorted(row["builds"], key=lambda b: b["file"])} if row.get("builds") else {}),
            })
        reviews = sorted({json.dumps(r, sort_keys=True) for r in self.reviews})
        rewrites = sorted({json.dumps(r, sort_keys=True) for r in self.rewrites})
        return {
            "version": VERSION,
            "legacy": self.legacy,
            "roots": [{"path": r.path, "head": r.head, "git": r.is_git} for r in self.roots],
            "rows": rows,
            "rewrites": sorted((json.loads(r) for r in rewrites), key=lambda r: (r["root"], r["file"], r["line"], r["col"])),
            "reviews": [json.loads(r) for r in reviews],
            "index_lines": self.index_lines,
            "layout_lines": self.layout_lines,
            "kept_refs": [{"root": i, "file": rel, "sha256_line": digest}
                          for i, rel, digest in sorted(self.kept_refs, key=lambda k: (k[0], k[1], k[2] or ""))],
        }


def markdown(plan):
    lines = []
    for i, root in enumerate(plan["roots"]):
        lines += [f"## {root['path']}", "", "| Source | Rule | Destination | Notes |", "| --- | --- | --- | --- |"]
        for row in (r for r in plan["rows"] if r["root"] == i):
            dest = (f"`{row['destination']}`" if row["destination"] else "**delete**" if row["action"] == "delete"
                    else "kept" if row["action"] == "keep" else "review")
            notes = "; ".join(row["notes"] + (["leaves Git"] if row["untrack"] else []) + (["user override"] if row["override"] else []))
            lines.append(f"| `{row['source']}` | {row['rule']} | {dest} | {notes} |")
        rewrites = [r for r in plan["rewrites"] if r["root"] == i]
        lines += ["", f"Reference rewrites: {len(rewrites)} in {len({r['file'] for r in rewrites})} files."]
        reviews = [r for r in plan["reviews"] if r["root"] == i]
        if reviews:
            lines += ["", "For review:"]
            lines += [f"- `{r['path']}` — {r['reason']}: {r['detail']}" for r in reviews]
        for item in plan["layout_lines"]:
            if item["root"] == i:
                lines += ["", f"Index line to write: `{item['line']}`"]
        lines.append("")
    return "\n".join(lines)


def cmd_plan(args):
    overrides = []
    for spec in args.override or []:
        if "=" not in spec:
            return fail(f"--override needs SRC=DEST: {spec}")
        overrides.append(tuple(spec.split("=", 1)))
    homes = {}
    for spec in args.home or []:
        key, _, value = spec.partition("=")
        if key not in ("tools", "infra", "docs") or not value:
            return fail(f"--home needs tools|infra|docs=PATH: {spec}")
        homes[key] = cl.norm_path(value)
    for spec in args.keep_ref or []:
        if not KEEP_REF_RE.match(spec):
            return fail(f"--keep-ref needs FILE[:LINE]: {spec!r}")
    for root in args.root:
        if not os.path.isdir(root):
            return fail(f"not a directory: {root}")
    plan = Planner(args.root, overrides, homes, args.keep or [], args.keep_ref or [], args.legacy).plan()
    text = json.dumps(plan, indent=2, sort_keys=True, ensure_ascii=False) + "\n"
    if args.out:
        with open(args.out, "w", encoding="utf-8") as handle:
            handle.write(text)
    print(text if args.json else markdown(plan), end="" if args.json else "\n")
    return 0


# ---- apply, verify, undo -------------------------------------------------------------

def load(path):
    with open(path, encoding="utf-8") as handle:
        plan = json.load(handle)
    if plan.get("version") != VERSION:
        raise SystemExit(fail(f"unknown plan version in {path}"))
    return plan


def root_path(plan, i):
    return plan["roots"][i]["path"]


def write_lines(path, lines):
    with open(path, "r+b") as handle:  # in place, so the file keeps its mode
        handle.seek(0)
        handle.write("".join(lines).encode("utf-8"))
        handle.truncate()


def save_undo(path, undo):
    with open(path, "w", encoding="utf-8") as handle:
        json.dump(undo, handle, indent=2, sort_keys=True, ensure_ascii=False)


def validate(plan):
    """Errors that make apply refuse, and the current lines of each file it rewrites."""
    errors = []
    for root in plan["roots"]:
        if root["git"]:
            head = git(root["path"], "rev-parse", "HEAD")
            current = head.stdout.strip() if head is not None and head.returncode == 0 else None
            if current != root["head"]:
                errors.append(f"{root['path']}: HEAD is {current}, the plan was made at {root['head']}")
    for row in plan["rows"]:
        if not row["destination"]:
            continue
        base = root_path(plan, row["root"])
        if fingerprint(os.path.join(base, row["source"])) != row["fingerprint"]:
            errors.append(f"{row['source']}: changed since the plan")
        if os.path.lexists(os.path.join(base, row["destination"])):
            errors.append(f"{row['destination']}: destination exists")
        if row["tracked"] and row["destination"].startswith("../") and not row.get("untrack"):
            errors.append(f"{row['source']}: it is tracked and moves outside the repository, so it must leave Git: "
                          "plan it again")
        if row.get("untrack") and row["tracked"]:
            staged = git(base, "diff", "--cached", "--name-only", "--", row["source"])
            if staged is None or staged.returncode != 0 or staged.stdout.strip():
                errors.append(f"{row['source']}: it has staged changes (or Git cannot say): commit or unstage them first")
            flags = git(base, "ls-files", "--debug", "--", row["source"])
            if flags is None or flags.returncode != 0:
                errors.append(f"{row['source']}: Git cannot list its index entries")
            elif any(int(f, 16) & INDEX_FLAGS for f in re.findall(r"flags: ([0-9a-f]+)", flags.stdout)):
                errors.append(f"{row['source']}: an index entry is intent-to-add, skip-worktree or assume-unchanged, "
                              "which undo cannot restore: clear the flag (git update-index) or keep it with --keep")
    def absolute(row, field):
        return os.path.normpath(os.path.join(root_path(plan, row["root"]), row[field]))

    moving = sorted((r for r in plan["rows"] if r["destination"]), key=lambda r: fold(absolute(r, "destination")))
    for n, row in enumerate(moving):
        dest = absolute(row, "destination")
        for other in moving[n + 1:]:
            other_dest = absolute(other, "destination")
            if fold(other_dest) == fold(dest):
                errors.append(f"{other['destination']}: {row['source']} and {other['source']} move to the same path")
            elif fold(other_dest).startswith(fold(dest) + os.sep):
                if row["kind"] == "folder" and not nested_collides(absolute(row, "source"), dest, other_dest):
                    continue  # the outer folder moves first, in this plan, and holds nothing at that path
                errors.append(f"{other['destination']}: nested inside the destination {row['destination']}")
    for i, root in enumerate(plan["roots"]):
        removing = [r for r in plan["rows"] if r["root"] == i and (r.get("untrack") or r["action"] == "delete")]
        if root["git"] and removing and not tagged(root["path"], root["head"]):
            errors.append(f"{root['path']}: {len(removing)} rows leave Git or are deleted and no archive/* tag points at "
                          f"{(root['head'] or '')[:8]}: create it first, as project-layout step 5 sets")
    originals = {}
    for item in plan["rewrites"]:
        key = (item["root"], item["file"])
        if key not in originals:
            text = read_text_strict(os.path.join(root_path(plan, item["root"]), item["file"]))
            originals[key] = text.splitlines(True) if text is not None else None
        lines = originals[key]
        if lines is None:
            errors.append(f"{item['file']}: unreadable")
            continue
        line = lines[item["line"] - 1] if item["line"] <= len(lines) else ""
        if sha256_bytes(line.encode()) != item["sha256_line"] or line[item["col"]:item["col"] + len(item["old"])] != item["old"]:
            errors.append(f"{item['file']}:{item['line']}: line changed since the plan")
    return sorted(set(errors)), originals


def untracked_entries(root, source):
    """The index entries under source, as `git ls-files -s -z` prints them, or None when Git cannot say.
    Read as bytes: a name that is not UTF-8 raises RuntimeError, since the undo record holds text."""
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    try:
        result = subprocess.run(["git", "-C", root, "ls-files", "-s", "-z", "--", source],
                                capture_output=True, env=env, timeout=60)
    except (OSError, subprocess.SubprocessError):
        return None
    if result.returncode != 0:
        return None
    try:
        return result.stdout.decode("utf-8")
    except UnicodeDecodeError:
        raise RuntimeError(f"{source}: an index entry has a name that is not UTF-8, which the undo record cannot hold")


def restore_entries(root, entries):
    result = git(root, "update-index", "-z", "--index-info", input=entries)
    return result is not None and result.returncode == 0


def tagged(root, head):
    result = git(root, "tag", "--list", "--points-at", head, "archive/*")
    return result is not None and result.returncode == 0 and bool(result.stdout.strip())


def cmd_apply(args):
    plan = load(args.plan)
    errors, originals = validate(plan)
    if errors:
        for error in errors:
            print(f"refused: {error}", file=sys.stderr)
        return 1
    undo_path = args.plan + ".undo.json"
    undo = {"version": UNDO_VERSION, "rewrites": [], "moves": [], "removed_dirs": [], "created_dirs": []}
    save_undo(undo_path, undo)  # from here on, every step is recorded before the next one
    by_file = collections.defaultdict(list)
    for item in plan["rewrites"]:
        by_file[(item["root"], item["file"])].append(item)
    try:
        for (i, rel), items in sorted(by_file.items()):
            lines = originals[(i, rel)]
            before = list(lines)
            for item in sorted(items, key=lambda it: (it["line"], -it["col"])):
                line = lines[item["line"] - 1]
                lines[item["line"] - 1] = line[:item["col"]] + item["new"] + line[item["col"] + len(item["old"]):]
            write_lines(os.path.join(root_path(plan, i), rel), lines)
            for number in sorted({it["line"] for it in items}):
                undo["rewrites"].append({"root": i, "file": rel, "line": number,
                                         "before": before[number - 1], "after": lines[number - 1]})
            save_undo(undo_path, undo)
        moves = sorted((r for r in plan["rows"] if r["destination"]),
                       key=lambda r: (os.path.normpath(os.path.join(root_path(plan, r["root"]), r["destination"])).count(os.sep),
                                      r["root"], r["source"]))  # outer destinations first
        for row in moves:
            root = root_path(plan, row["root"])
            source = os.path.join(root, row["source"])
            dest = os.path.normpath(os.path.join(root, row["destination"]))
            if os.path.lexists(dest):
                raise RuntimeError(f"{row['destination']}: destination exists; {row['source']} stays")
            missing = []
            parent = os.path.dirname(dest)
            while not os.path.isdir(parent):
                missing.append(parent)
                parent = os.path.dirname(parent)
            os.makedirs(os.path.dirname(dest), exist_ok=True)
            undo["created_dirs"].extend({"root": row["root"], "dir": os.path.relpath(d, root)} for d in reversed(missing))
            leaves_git = bool(row["tracked"] and row.get("untrack"))
            if row["tracked"] and not leaves_git:
                result = git(root, "mv", "--", row["source"], row["destination"])
                if result is None or result.returncode != 0:
                    raise RuntimeError(f"git mv {row['source']}: {result.stderr.strip() if result else 'failed'}")
            else:
                if leaves_git:  # the index entries are recorded before the files move: they are the only copy
                    entries = untracked_entries(root, row["source"])
                    if not entries:
                        raise RuntimeError(f"git ls-files {row['source']}: failed")
                    undo["moves"].append({"root": row["root"], "source": row["source"], "destination": row["destination"],
                                          "git": False, "untracked": True, "index": entries})
                    save_undo(undo_path, undo)
                try:
                    os.replace(source, dest)
                except OSError:
                    if leaves_git:
                        undo["moves"].pop()
                        save_undo(undo_path, undo)
                    raise
                if leaves_git:
                    result = git(root, "rm", "-r", "-q", "--cached", "--", row["source"])
                    if result is None or result.returncode != 0:
                        raise RuntimeError(f"git rm --cached {row['source']}: {result.stderr.strip() if result else 'failed'}")
            if not leaves_git:
                undo["moves"].append({"root": row["root"], "source": row["source"], "destination": row["destination"],
                                      "git": bool(row["tracked"]), "untracked": False})
            parent = os.path.dirname(source)
            while os.path.basename(parent) != "_support" and parent.startswith(root + os.sep):
                if os.listdir(parent):
                    break
                os.rmdir(parent)
                undo["removed_dirs"].append({"root": row["root"], "dir": os.path.relpath(parent, root)})
                parent = os.path.dirname(parent)
            save_undo(undo_path, undo)
    except (OSError, RuntimeError) as error:
        save_undo(undo_path, undo)
        print(f"apply stopped: {error}; undo what it did with: plan_layout.py undo --plan {args.plan}", file=sys.stderr)
        return 1
    print(f"moved {len(undo['moves'])}, rewrote {len(undo['rewrites'])} lines; undo record: {undo_path}")
    for review in plan["reviews"]:
        print(f"review: {review['path']} — {review['reason']}: {review['detail']}")
    return 0


def cmd_verify(args):
    plan = load(args.plan)
    roots = [p["path"] for p in plan["roots"]]
    gone = set()
    sources = {}  # absolute source of a moved or deleted row, now absent -> the row
    moved_markdown = []
    for row in plan["rows"]:
        if not row["destination"] and row["action"] != "delete":
            continue
        source = os.path.join(roots[row["root"]], row["source"])
        if not os.path.lexists(source):
            gone.add(source)
            sources[source] = row
            parent = os.path.dirname(source)
            while parent not in roots and parent != os.path.dirname(parent) and not os.path.lexists(parent):
                gone.add(parent)
                parent = os.path.dirname(parent)
        if not row["destination"]:
            continue
        dest = os.path.join(roots[row["root"]], row["destination"])
        if not HISTORY_RE.search(row["source"]):
            for f in ([dest] if os.path.isfile(dest) else cl.list_files(dest)):
                if f.endswith(MARKDOWN):
                    old = source if f == dest else os.path.join(source, os.path.relpath(f, dest))
                    moved_markdown.append((f, os.path.dirname(old), row["root"]))

    def existed_inside(target):
        """Whether target, below a gone row source, existed before apply: at the moved row's destination,
        or at the plan's HEAD for a deleted one."""
        for source, row in sources.items():
            if target.startswith(source + os.sep):
                inner = os.path.relpath(target, source).replace(os.sep, "/")
                if row["destination"]:
                    return os.path.lexists(os.path.join(roots[row["root"]], row["destination"], inner))
                head = plan["roots"][row["root"]]["head"]
                found = git(roots[row["root"]], "cat-file", "-e", f"{head}:{row['source']}/{inner}") if head else None
                return found is not None and found.returncode == 0
        return False

    residual = []
    kept = {(k["root"], k["file"], k["sha256_line"]) for k in plan.get("kept_refs", [])}

    def is_kept(i, rel, line):
        return (i, rel, None) in kept or (i, rel, line_hash(line)) in kept

    planner = Planner(roots, [], {})
    for i in range(len(roots)):
        for rel in planner.readable_files(i):
            text = planner.content(i, rel)
            relative = planner.scans_relative(rel)
            if not text or ("_support/" not in text and not relative):
                continue
            for number, line in enumerate(text.splitlines(), 1):
                if is_kept(i, rel, line):
                    continue
                for _, pre, path in tokens(line, relative):
                    targets = [t for _, t in planner.resolutions(i, rel, pre, path)]
                    if not targets or any(os.path.lexists(t) for t in targets):
                        continue
                    if any(c in path for c in "*<>{}$") or "YYYY" in path:
                        continue  # a glob or a placeholder names no single path
                    if any(t in gone or existed_inside(t) for t in targets):
                        residual.append(f"{roots[i]}/{rel}:{number}: {pre}{path}")
    for path, old_dir, owner in moved_markdown:
        text = read_text_strict(path) or ""
        for number, line in enumerate(text.splitlines(), 1):
            if is_kept(owner, os.path.relpath(path, roots[owner]).replace(os.sep, "/"), line):
                continue
            for _, pre, link in tokens(line, True):
                if pre or not link.startswith("../"):
                    continue
                clean = link.split("*")[0].rstrip("/")
                if os.path.lexists(os.path.normpath(os.path.join(os.path.dirname(path), clean))):
                    continue
                if not os.path.lexists(os.path.normpath(os.path.join(old_dir, clean))):
                    continue  # already broken before the move: not this plan's doing
                residual.append(f"{path}:{number}: {link} points at nothing")
    for i, root in enumerate(plan["roots"]):
        tools = [r for r in plan["rows"] if r["root"] == i and r["rule"] in TOOL_RULES and r["destination"]
                 and not TEST_NAME.match(os.path.basename(r["source"]))]
        if not tools:
            continue
        index = cl.find_index(root["path"])
        if not index:
            residual.append(f"{root['path']}: moved tools and no index; create _support/docs/README.md with a line per tool")
            continue
        text = cl.read_text(index)
        residual += [f"{index}: no line for {r['destination']}" for r in tools if os.path.basename(r["destination"]) not in text]
    residual += build_residuals(plan)
    for item in sorted(set(residual)):
        print(item)
    for review in plan["reviews"]:
        if review["reason"] != "ignored files stay":
            continue
        left = left_ignored(root_path(plan, review["root"]), review["path"])
        if left:
            print(f"note: {review['path']} still holds {plural(len(left), 'entry', 'entries')} Git ignores, "
                  f"such as {', '.join(left[:3])}")
    return 1 if residual else 0


def build_residuals(plan):
    """Moved rows an exclusion file now judges otherwise than their source, unless a line
    of it names the new path."""
    by_file = collections.defaultdict(list)
    for row in plan["rows"]:
        for build in row.get("builds", []):
            if row["destination"].startswith("../"):
                continue  # a plan an earlier script wrote: a ../ destination left the repository, so no build here holds it
            suffix = "/" if row["kind"] != "file" else ""
            by_file[(row["root"], build["file"])].append((row["destination"] + suffix, build))
    residual = []
    for (i, ignore_file), items in sorted(by_file.items()):
        root = root_path(plan, i)
        status = build_status(root, ignore_file, sorted({dest for dest, _ in items}))
        lines = exclusion_lines(os.path.join(root, ignore_file))
        for dest, build in items:
            if status[dest] == build["excluded"] or names_path(lines, ignore_file, dest, build["new_folder"]):
                continue
            verdict = "no longer excludes" if build["excluded"] else "now excludes"
            residual.append(f"{root}/{ignore_file} {verdict} {dest.rstrip('/')}: change it back, "
                            f"or name {build['new_folder'] or dest.rstrip('/')} in it to keep the change")
    return residual


def cmd_undo(args):
    undo_path = args.plan + ".undo.json"
    try:
        with open(undo_path, encoding="utf-8") as handle:
            undo = json.load(handle)
    except OSError:
        return fail(f"no undo record at {undo_path}")
    plan = load(args.plan)
    old = [m for m in undo["moves"] if m.get("untracked") and not m.get("index")]
    if old:
        for move in old:
            print(f"refused: {move['source']}: the record was written by an earlier script and holds no index entries; "
                  f"undo would leave it untracked. Move {move['destination']} back to {move['source']} by hand, "
                  f"then run `git add -- {move['source']}`", file=sys.stderr)
        return 1
    conflicts = []
    for removed in reversed(undo["removed_dirs"]):
        os.makedirs(os.path.join(root_path(plan, removed["root"]), removed["dir"]), exist_ok=True)
    for move in reversed(undo["moves"]):
        root = root_path(plan, move["root"])
        if move["git"]:
            result = git(root, "mv", "--", move["destination"], move["source"])
            if result is None or result.returncode != 0:
                conflicts.append(f"git mv back failed for {move['destination']}")
        else:
            moved, back = os.path.join(root, move["destination"]), os.path.join(root, move["source"])
            if not os.path.lexists(moved) and os.path.lexists(back):
                continue  # recorded before the files moved, and apply stopped first
            os.replace(moved, back)
            if move.get("untracked") and not restore_entries(root, move["index"]):
                conflicts.append(f"restoring the index entries of {move['source']} failed")
    for created in reversed(undo["created_dirs"]):
        path = os.path.join(root_path(plan, created["root"]), created["dir"])
        try:
            if os.path.isdir(path) and not os.listdir(path):
                os.rmdir(path)
        except OSError:
            pass
    by_file = collections.defaultdict(list)
    for item in undo["rewrites"]:
        by_file[(item["root"], item["file"])].append(item)
    for (i, rel), items in sorted(by_file.items()):
        path = os.path.join(root_path(plan, i), rel)
        text = read_text_strict(path)
        if text is None:
            conflicts.append(f"{rel}: unreadable or missing after apply; its rewrites are left as they are")
            continue
        lines = text.splitlines(True)
        for item in items:
            if item["line"] <= len(lines) and lines[item["line"] - 1] == item["after"]:
                lines[item["line"] - 1] = item["before"]
            else:
                conflicts.append(f"{rel}:{item['line']} changed after apply; left as is")
        write_lines(path, lines)
    for conflict in conflicts:
        print(conflict, file=sys.stderr)
    if not conflicts:
        os.remove(undo_path)
    print(f"undid {len(undo['moves'])} moves and {len(undo['rewrites'])} line rewrites")
    return 1 if conflicts else 0


def main(argv=None):
    argv = list(sys.argv[1:] if argv is None else argv)
    if not argv or argv[0] not in ("plan", "apply", "verify", "undo", "-h", "--help"):
        argv.insert(0, "plan")
    parser = argparse.ArgumentParser(description="Deterministic placement plan for project-layout.")
    sub = parser.add_subparsers(dest="command", required=True)
    plan = sub.add_parser("plan", help="print the plan; changes nothing")
    plan.add_argument("--root", action="append", required=True)
    plan.add_argument("--override", action="append", metavar="SRC=DEST")
    plan.add_argument("--keep", action="append", metavar="SRC", help="leave this row, or every row under this folder, where it is")
    plan.add_argument("--keep-ref", action="append", metavar="FILE[:LINE]",
                      help="leave the references in this file, or on this line, as they are; verify accepts them")
    plan.add_argument("--home", action="append", metavar="KEY=PATH")
    plan.add_argument("--legacy", choices=("move", "remove"), default="move",
                      help="aged legacy records: move them to workspace/ (default), or delete the tracked, unmodified ones")
    plan.add_argument("--json", action="store_true")
    plan.add_argument("--out")
    for name in ("apply", "verify", "undo"):
        sub.add_parser(name).add_argument("--plan", required=True)
    args = parser.parse_args(argv)
    return {"plan": cmd_plan, "apply": cmd_apply, "verify": cmd_verify, "undo": cmd_undo}[args.command](args)


if __name__ == "__main__":
    sys.exit(main())
