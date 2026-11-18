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
import sys
import time

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import check_layout as cl  # noqa: E402  (python3 -I leaves the script's folder off sys.path)

VERSION = 1
# Names in check_layout.GATED that hold dependencies rather than build output.
DEPENDENCY_DIRS = ("vendor", "deps")
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
LEGACY_DATA = ("backups", "migration-data")
RUNTIME_CONFIG_EXTENSIONS = {".json", ".yaml", ".yml", ".toml", ".ini", ".env", ".cfg", ".conf"}
CI_PATHS = re.compile(
    r"^(?:\.github/workflows/|\.circleci/)|(?:^|/)(?:\.gitlab-ci\.yml|azure-pipelines\.yml|"
    r"bitbucket-pipelines\.yml|Jenkinsfile|buildspec[^/]*\.ya?ml|amplify\.yml)$"
)
HOOK_PATHS = re.compile(r"^(?:\.husky/|\.githooks/)|(?:^|/)(?:lefthook\.ya?ml|\.pre-commit-config\.yaml)$")
# Records of what happened: references there describe the past and are never rewritten.
HISTORY_RE = re.compile(
    r"^(?:_support/)?(?:openspec/changes/archive|sessions|evidence)/|^_support/(?:workspace|reports)/"
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


def fail(message):
    print(f"plan_layout.py: {message}", file=sys.stderr)
    return 1


def git(root, *args, **kwargs):
    return cl.git(root, *args, **kwargs)


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
        return os.path.join(self.path, rel)

    def _packages(self):
        found = set()
        for rel in self.files:
            if rel.startswith("_support/") or HISTORY_RE.search(rel):
                continue
            name = os.path.basename(rel)
            directory = os.path.dirname(rel)
            if directory and (name in MANIFESTS or name.endswith(MANIFEST_SUFFIXES)) and not self.in_dependencies(directory):
                found.add(directory)
        return sorted(found)

    def in_dependencies(self, rel):
        """Whether a folder lies in a committed dependency folder, such as Go's vendor/
        beside go.mod, which holds other projects' manifests. Build output a repository
        commits, such as build/, is its own source and still counts."""
        parent = self.path
        for name in rel.split("/"):
            if name in cl.PRUNE or (name in DEPENDENCY_DIRS and cl.is_dependency_dir(parent, name)):
                return True
            parent = os.path.join(parent, name)
        return False

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
    def __init__(self, roots, overrides, homes, keeps=()):
        self.roots = [Root(r) for r in roots]
        self.overrides = overrides
        self.used_overrides = set()
        self.keeps = list(keeps)
        self.used_keeps = set()
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
                    self.add_file(i, rel, f"_support/{entry}/", entry)
            support = root.abs("_support")
            if os.path.isdir(support):
                for name in sorted(os.listdir(support)):
                    if name not in cl.SUPPORT_CHILDREN:
                        self.add_entry(i, name)

    def new_row(self, i, rel, kind, entry, subpath=""):
        root = self.roots[i]
        if kind == "file":
            tracked = rel in root.tracked
        else:
            tracked = any(r.startswith(rel + "/") for r in root.tracked)
        return {
            "root": i, "source": rel, "kind": kind, "tracked": tracked, "subpath": subpath, "entry": entry,
            "rule": None, "destination": None, "notes": [], "override": False,
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
            if "kept by user" in self.rows[subject]["notes"]:
                return self.set_kept(key, "R1")
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
        if any(other in self.rows and "kept by user" in self.rows[other]["notes"] for other in related):
            return self.set_kept(key, "R3")  # what only kept files use stays with them
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
        return os.path.join("_support/sessions", f"{date}-{stem}", os.path.basename(row["source"]))

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
        if name in LEGACY_EVIDENCE + LEGACY_DATA:
            citing = self.runtime_citations(root.abs(base))
            if citing:
                return self.set_review(key, "F2", "runtime path: code or configuration cites it: " + ", ".join(citing[:5]))
        if name in LEGACY_EVIDENCE:
            del self.rows[key]
            return self.place_evidence(i, base, name)
        if name in LEGACY_DATA:
            return self.set_dest(key, "F4", f"_support/data/{oldest_mtime(root.abs(base))}-{kebab(name)}")
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

    def place_evidence(self, i, base, name):
        root = self.roots[i]
        loose = []
        for child in sorted(os.listdir(root.abs(base))):
            child_rel = f"{base}/{child}"
            if os.path.isdir(root.abs(child_rel)):
                match = DATE_RE.search(child)
                date = match.group(0) if match else oldest_mtime(root.abs(child_rel))
                work = kebab(DATE_RE.sub("", child, count=1)) or kebab(name)
                key = (i, child_rel)
                self.rows[key] = self.new_row(i, child_rel, "folder", name)
                self.set_dest(key, "F3", f"_support/workspace/{date}-{work}/evidence")
            elif child not in cl.IGNORED_NAMES:
                loose.append(child_rel)
        if loose:
            date = min(oldest_mtime(root.abs(r)) for r in loose)
            for rel in loose:
                key = (i, rel)
                self.rows[key] = self.new_row(i, rel, "file", name)
                self.set_dest(key, "F3", f"_support/workspace/{date}-legacy-{kebab(name)}/evidence/{os.path.basename(rel)}")

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

    def set_kept(self, key, rule):
        row = self.rows[key]
        if self.override_for(row):
            raise SystemExit(fail(f"--keep and --override both name {row['source']}: choose one"))
        row["rule"], row["destination"], row["override"] = rule, None, False
        row["notes"].append("kept by user")

    def set_dest(self, key, rule, dest):
        row = self.rows[key]
        if self.keep_for(row):
            return self.set_kept(key, rule)
        override = self.override_for(row)
        row["rule"] = rule
        row["destination"] = override or os.path.normpath(dest).replace(os.sep, "/")
        row["override"] = bool(override)

    def set_review(self, key, rule, reason):
        row = self.rows[key]
        if self.keep_for(row):
            return self.set_kept(key, rule)
        override = self.override_for(row)
        row["rule"] = rule
        if override:
            row["destination"], row["override"] = override, True
            return
        row["destination"] = None
        self.review(row["root"], row["source"], rule, reason)

    # ---- guards after classification ------------------------------------------------
    def guard(self):
        """Tracked rows going into ignored folders, then collisions; dependents follow."""
        for i, root in enumerate(self.roots):
            tracked = {r["destination"]: k for k, r in self.rows.items()
                       if r["root"] == i and r["tracked"] and r["destination"] and not r["override"]}
            for dest in sorted(root.ignored(sorted(tracked))):
                self.set_review(tracked[dest], self.rows[tracked[dest]]["rule"],
                                f"tracked: {dest} is ignored, so the move would drop it from Git; "
                                "remove it from the tree with the archive tag instead")
        self.mark_collisions()
        for key in self.dependents:
            row = self.rows[key]
            if not row["override"] and row["rule"] in ("R1", "R3"):
                self.place_dependent(key, row)
        self.mark_collisions()

    def mark_collisions(self):
        seen = collections.defaultdict(list)
        for key in sorted(self.rows):
            row = self.rows[key]
            if row["destination"]:
                seen[(row["root"], row["destination"])].append(key)
        for (i, dest), keys in sorted(seen.items()):
            if os.path.lexists(self.roots[i].abs(dest)) or len(keys) > 1:
                for key in keys:
                    row = self.rows[key]
                    row["notes"].append("collision")
                    row["override"] = False
                    row["destination"] = None
                    self.review(row["root"], row["source"], row["rule"], f"collision at {dest}")

    def mark_split_tests(self):
        """A kept test whose subject moves may run it by a path that breaks: ask."""
        for key in sorted(self.rows):
            row = self.rows[key]
            if "kept by user" not in row["notes"] or not self.is_test(row):
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
            if ref["kind"] == "literal":
                new_target = self.new_abs(ref["keys"][0])
                if new_target is None:
                    continue
                if ref["inside"]:
                    new_target = os.path.join(new_target, ref["inside"])
            else:
                new_target = self.prefix_destination(ref["target"], ref["keys"])
                if new_target is None and all(not self.rows[k]["destination"] and "kept by user" in self.rows[k]["notes"] for k in ref["keys"]):
                    continue  # nothing under it moves
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
        self.guard()
        self.apply_homes()
        unused = sorted({spec for spec, _ in self.overrides} - self.used_overrides)
        if unused:
            raise SystemExit(fail("--override names no planned row: " + ", ".join(unused)))
        unused = sorted(set(self.keeps) - self.used_keeps)
        if unused:
            raise SystemExit(fail("--keep names no planned row: " + ", ".join(unused)))
        self.mark_split_tests()
        self.mark_same_names()
        self.build_rewrites(refs)
        self.internal_paths()
        self.build_index_lines()
        rows = []
        for key in sorted(self.rows):
            row = self.rows[key]
            rows.append({
                "root": row["root"], "source": row["source"], "kind": row["kind"], "tracked": row["tracked"],
                "rule": row["rule"], "destination": row["destination"], "notes": sorted(set(row["notes"])),
                "override": row["override"], "fingerprint": fingerprint(self.roots[row["root"]].abs(row["source"])),
            })
        reviews = sorted({json.dumps(r, sort_keys=True) for r in self.reviews})
        rewrites = sorted({json.dumps(r, sort_keys=True) for r in self.rewrites})
        return {
            "version": VERSION,
            "roots": [{"path": r.path, "head": r.head, "git": r.is_git} for r in self.roots],
            "rows": rows,
            "rewrites": sorted((json.loads(r) for r in rewrites), key=lambda r: (r["root"], r["file"], r["line"], r["col"])),
            "reviews": [json.loads(r) for r in reviews],
            "index_lines": self.index_lines,
            "layout_lines": self.layout_lines,
        }


def markdown(plan):
    lines = []
    for i, root in enumerate(plan["roots"]):
        lines += [f"## {root['path']}", "", "| Source | Rule | Destination | Notes |", "| --- | --- | --- | --- |"]
        for row in (r for r in plan["rows"] if r["root"] == i):
            dest = f"`{row['destination']}`" if row["destination"] else "kept" if "kept by user" in row["notes"] else "review"
            notes = "; ".join(row["notes"] + (["user override"] if row["override"] else []))
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
    for root in args.root:
        if not os.path.isdir(root):
            return fail(f"not a directory: {root}")
    plan = Planner(args.root, overrides, homes, args.keep or []).plan()
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


def cmd_apply(args):
    plan = load(args.plan)
    errors, originals = validate(plan)
    if errors:
        for error in errors:
            print(f"refused: {error}", file=sys.stderr)
        return 1
    undo_path = args.plan + ".undo.json"
    undo = {"rewrites": [], "moves": [], "removed_dirs": [], "created_dirs": []}
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
        moves = sorted((r for r in plan["rows"] if r["destination"]), key=lambda r: (r["root"], r["source"]))
        for row in moves:
            root = root_path(plan, row["root"])
            source = os.path.join(root, row["source"])
            dest = os.path.join(root, row["destination"])
            missing = []
            parent = os.path.dirname(dest)
            while not os.path.isdir(parent):
                missing.append(parent)
                parent = os.path.dirname(parent)
            os.makedirs(os.path.dirname(dest), exist_ok=True)
            undo["created_dirs"].extend({"root": row["root"], "dir": os.path.relpath(d, root)} for d in reversed(missing))
            if row["tracked"]:
                result = git(root, "mv", "--", row["source"], row["destination"])
                if result is None or result.returncode != 0:
                    raise RuntimeError(f"git mv {row['source']}: {result.stderr.strip() if result else 'failed'}")
            else:
                os.replace(source, dest)
            undo["moves"].append({"root": row["root"], "source": row["source"], "destination": row["destination"], "git": bool(row["tracked"])})
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
    moved_markdown = []
    for row in plan["rows"]:
        if not row["destination"]:
            continue
        source = os.path.join(roots[row["root"]], row["source"])
        if not os.path.lexists(source):
            gone.add(source)
            parent = os.path.dirname(source)
            while parent not in roots and parent != os.path.dirname(parent) and not os.path.lexists(parent):
                gone.add(parent)
                parent = os.path.dirname(parent)
        dest = os.path.join(roots[row["root"]], row["destination"])
        if not HISTORY_RE.search(row["source"]):
            for f in ([dest] if os.path.isfile(dest) else cl.list_files(dest)):
                if f.endswith(MARKDOWN):
                    old = source if f == dest else os.path.join(source, os.path.relpath(f, dest))
                    moved_markdown.append((f, os.path.dirname(old)))
    residual = []
    planner = Planner(roots, [], {})
    for i in range(len(roots)):
        for rel in planner.readable_files(i):
            text = planner.content(i, rel)
            relative = planner.scans_relative(rel)
            if not text or ("_support/" not in text and not relative):
                continue
            for number, line in enumerate(text.splitlines(), 1):
                for _, pre, path in tokens(line, relative):
                    targets = [t for _, t in planner.resolutions(i, rel, pre, path)]
                    if targets and not any(os.path.lexists(t) for t in targets) and any(t in gone for t in targets):
                        residual.append(f"{roots[i]}/{rel}:{number}: {pre}{path}")
    for path, old_dir in moved_markdown:
        text = read_text_strict(path) or ""
        for number, line in enumerate(text.splitlines(), 1):
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
    for item in sorted(set(residual)):
        print(item)
    return 1 if residual else 0


def cmd_undo(args):
    undo_path = args.plan + ".undo.json"
    try:
        with open(undo_path, encoding="utf-8") as handle:
            undo = json.load(handle)
    except OSError:
        return fail(f"no undo record at {undo_path}")
    plan = load(args.plan)
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
            os.replace(os.path.join(root, move["destination"]), os.path.join(root, move["source"]))
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
    plan.add_argument("--home", action="append", metavar="KEY=PATH")
    plan.add_argument("--json", action="store_true")
    plan.add_argument("--out")
    for name in ("apply", "verify", "undo"):
        sub.add_parser(name).add_argument("--plan", required=True)
    args = parser.parse_args(argv)
    return {"plan": cmd_plan, "apply": cmd_apply, "verify": cmd_verify, "undo": cmd_undo}[args.command](args)


if __name__ == "__main__":
    sys.exit(main())
