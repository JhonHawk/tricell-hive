#!/usr/bin/env python3
"""Plan, apply, verify and undo the moves project-layout makes, deterministically.

  plan_layout.py [plan] --root DIR [--root DIR ...] [--override SRC=DEST ...]
                 [--home tools|infra|docs=PATH ...] [--json] [--out FILE]
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
import subprocess
import sys
import time

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import check_layout as cl  # noqa: E402

VERSION = 1
MANIFESTS = (
    "package.json", "go.mod", "pyproject.toml", "setup.py", "composer.json", "Gemfile", "pom.xml",
    "build.gradle", "build.gradle.kts", "Cargo.toml", "pubspec.yaml", "mix.exs",
)
MANIFEST_GLOBS = ("*.csproj",)
TOOL_ROOTS = ("scripts", "script", "tools", "hack")
PACKAGE_TOOL_DIRS = ("scripts", "script", "tools")
IAC_HOMES = ("infra", "infrastructure", "terraform", "deployments")
GROUPS = {
    "ci": ("ci", "cicd", "pipeline", "pipelines"),
    "hooks": ("hooks", "git-hooks", "githooks"),
    "ops": ("ops", "operations", "runbooks"),
}
ONE_OFF_WORDS = {"backfill", "fix", "cutover", "migrate", "migration", "seed", "cleanup", "oneoff", "once", "repair"}
KEY_RE = re.compile(r"\b[A-Za-z]{2,10}-\d+\b")
DATE_RE = re.compile(r"\d{4}-\d{2}-\d{2}")
DOC_EXTENSIONS = {".md", ".mdx", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp"}
LEGACY_EVIDENCE = ("evidence", "reports")
LEGACY_DATA = ("backups", "migration-data")
RUNTIME_CONFIG_EXTENSIONS = {".json", ".yaml", ".yml", ".toml", ".ini", ".env", ".cfg", ".conf"}
CI_PATHS = re.compile(
    r"^(?:\.github/workflows/|\.circleci/)|(?:^|/)(?:\.gitlab-ci\.yml|azure-pipelines\.yml|"
    r"bitbucket-pipelines\.yml|Jenkinsfile|buildspec[^/]*\.ya?ml|amplify\.yml)$"
)
HOOK_PATHS = re.compile(r"^(?:\.husky/|\.githooks/)|(?:^|/)(?:lefthook\.ya?ml|\.pre-commit-config\.yaml)$")
HISTORY_RE = re.compile(
    r"(?:^|/)openspec/changes/archive/|(?:^|/)sessions/|(?:^|/)evidence/|(?:^|/)_support/workspace/"
    r"|(?:^|/)CHANGELOG[^/]*$"
)
NEVER_READ = re.compile(r"(?:^|/)_support/(?:secrets|data)/|(?:^|/)\.engram/|(?:^|/)\.git/")
TEST_NAME = re.compile(r"^(?P<stem>.+?)(?:\.test|\.spec|_test)(?P<ext>\.[^.]+)$|^test_(?P<stem2>.+)$")
# A path token that reaches into a support folder, optionally with a variable, a relative
# prefix or a repository prefix. Boundaries keep a substring from being read as a token.
TOKEN_RE = re.compile(
    r"(?<![\w@./-])((?:\$\{?\w+\}?/)?(?:\.{1,2}/)*(?:[\w.@+-]+/)*?_support/[\w.@+/*-]*)"
)
ANCHOR_RE = re.compile(r"BASH_SOURCE|dirname \"?\$0|\$\(dirname|__dirname|import\.meta\.url|__file__")
REL_LITERAL_RE = re.compile(r"(?<![\w.])((?:\.\./)*\.\.(?:/[\w.@-]+)*)(?![\w.])")
MAX_READ = 5 * 1024 * 1024


def fail(message):
    print(f"plan_layout.py: {message}", file=sys.stderr)
    return 1


def git(root, *args):
    return cl.git(root, *args)


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
    """(text, ok): text decoded as strict UTF-8 without NUL bytes, else (None, False)."""
    try:
        if os.path.getsize(path) > MAX_READ:
            return None, False
    except OSError:
        return None, False
    data = read_bytes(path)
    if data is None or b"\0" in data:
        return None, False
    try:
        return data.decode("utf-8"), True
    except UnicodeDecodeError:
        return None, False


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


def gm_date(seconds):
    return time.strftime("%Y-%m-%d", time.gmtime(seconds))


def oldest_mtime(path):
    files = cl.list_files(path) if os.path.isdir(path) else [path]
    stamps = []
    for item in files:
        try:
            stamps.append(os.path.getmtime(item))
        except OSError:
            continue
    return gm_date(min(stamps)) if stamps else "0000-00-00"


class Root:
    def __init__(self, path):
        self.path = os.path.abspath(path)
        self.is_git = cl.is_git_toplevel(self.path)
        head = git(self.path, "rev-parse", "HEAD") if self.is_git else None
        self.head = head.stdout.strip() if head is not None and head.returncode == 0 else None
        self.files = [os.path.relpath(p, self.path).replace(os.sep, "/") for p in cl.list_files(self.path)]
        self.tracked = set(cl.tracked_files(self.path)) if self.is_git else set()
        index = cl.find_index(self.path)
        self.index = index
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
            if directory and (name in MANIFESTS or any(re.fullmatch(g.replace("*", ".*"), name) for g in MANIFEST_GLOBS)):
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
            if os.path.isdir(os.path.join(self.path, base, name) if base else self.abs(name)):
                return name
        return None

    def first_commit_date(self, rel):
        result = git(self.path, "log", "--diff-filter=A", "--format=%cs", "--reverse", "--", rel)
        if result is not None and result.returncode == 0 and result.stdout.strip():
            return result.stdout.split()[0]
        return None


class Planner:
    def __init__(self, roots, overrides, homes):
        self.roots = [Root(r) for r in roots]
        self.overrides = overrides
        self.homes_override = homes
        self.rows = {}  # (root index, source) -> row
        self.reviews = []
        self.rewrites = []
        self.contents = {}  # (root index, rel) -> text, or None when unreadable
        self.layout_lines = []
        self.index_lines = []
        self._homes = {}

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
        root = self.roots[i]
        for rel in root.files:
            if NEVER_READ.search("/" + rel) or HISTORY_RE.search(rel):
                continue
            yield rel

    def content(self, i, rel):
        key = (i, rel)
        if key not in self.contents:
            text, ok = read_text_strict(self.roots[i].abs(rel))
            self.contents[key] = text if ok else None
        return self.contents[key]

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
                    if name in cl.SUPPORT_CHILDREN:
                        continue
                    self.add_folder(i, name)

    def add_file(self, i, rel, prefix, entry):
        key = (i, rel)
        if key in self.rows:
            return
        root = self.roots[i]
        self.rows[key] = {
            "root": i, "source": rel, "kind": "file", "tracked": rel in root.tracked,
            "subpath": rel[len(prefix):], "entry": entry, "rule": None, "destination": None,
            "notes": [], "override": False,
        }

    def add_folder(self, i, name):
        root = self.roots[i]
        base = f"_support/{name}"
        path = root.abs(base)
        files = [r for r in root.files if r.startswith(base + "/")]
        has_tf = any(r.endswith(".tf") for r in files)
        if os.path.isdir(path) and (name in cl.IAC_FAMILY or has_tf):
            for rel in files:
                self.add_file(i, rel, base + "/", name)
            return
        if any((i, r) in self.rows for r in files):
            rest = [r for r in files if (i, r) not in self.rows]
            if rest:
                self.review(i, base, "no rule", f"{len(rest)} other files stay; classify them by hand")
            return
        self.rows[(i, base)] = {
            "root": i, "source": base, "kind": "folder", "tracked": any(r in root.tracked for r in files),
            "subpath": "", "entry": name, "rule": None, "destination": None, "notes": [], "override": False,
        }

    def review(self, i, path, reason, detail):
        self.reviews.append({"root": i, "path": path, "reason": reason, "detail": detail})

    # ---- references ------------------------------------------------------------
    def resolutions(self, i, citing, token):
        """(base kind, absolute path) pairs a token may point to, most specific first.

        `$VAR/x` is read from the root; `./x` and `../x` from the citing file's folder, then
        from the root, where CI and package scripts run; any other path from the root."""
        root = self.roots[i]
        path = token
        star = path.find("*")
        if star >= 0:
            path = path[:star]
        path = path.rstrip("/").rstrip(".")
        if not path:
            return []
        var = re.match(r"^\$\{?\w+\}?/", path)
        if var:
            return [("var", os.path.normpath(os.path.join(root.path, path[var.end():])))]
        if path.startswith("./") or path.startswith("../"):
            return [
                ("file", os.path.normpath(os.path.join(os.path.dirname(root.abs(citing)), path))),
                ("root", os.path.normpath(os.path.join(root.path, path))),
            ]
        return [("root", os.path.normpath(os.path.join(root.path, path)))]

    def candidate_index(self):
        files, folders = {}, {}
        for key, row in self.rows.items():
            absolute = self.roots[row["root"]].abs(row["source"])
            (files if row["kind"] == "file" else folders)[absolute] = key
        return files, folders

    def match_target(self, target, files, folders, candidate_dirs):
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

    def candidate_dirs(self, files, folders):
        dirs = collections.defaultdict(list)
        roots = {r.path for r in self.roots}
        for absolute, key in list(files.items()) + list(folders.items()):
            parent = os.path.dirname(absolute)
            while parent not in roots and parent != os.path.dirname(parent):
                dirs[parent].append(key)
                parent = os.path.dirname(parent)
        return dirs

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
                if "_support/" not in text:
                    continue
                for number, line in enumerate(text.splitlines(True), 1):
                    if "_support/" not in line:
                        continue
                    for match in TOKEN_RE.finditer(line):
                        spelled = match.group(1).rstrip(".")
                        for base, target in self.resolutions(i, rel, spelled):
                            kind, keys, inside = self.match_target(target, files, folders, candidate_dirs)
                            if kind:
                                refs.append({
                                    "root": i, "file": rel, "line": number, "col": match.start(1), "token": spelled,
                                    "target": target, "base": base, "inside": inside, "kind": kind, "keys": keys,
                                    "text": line,
                                })
                                break
        return refs

    # ---- classification --------------------------------------------------------
    def caller_class(self, i, rel):
        if CI_PATHS.search(rel):
            return "ci"
        if HOOK_PATHS.search(rel):
            return "hook"
        if rel.lower().endswith((".md", ".mdx")):
            return "doc"
        package = self.roots[i].package_of(rel)
        if package:
            return f"package:{package}"
        return "root"

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
        self.callers, self.cites, self.cited_by = callers, cites, cited_by
        ordered = sorted(self.rows)
        # tests follow what they test, and libraries and fixtures follow the tests that use them
        tests, libs = [], []
        for key in ordered:
            row = self.rows[key]
            if row["kind"] == "folder":
                self.place_folder(key, row, refs)
            elif self.is_test(row):
                tests.append(key)
            elif self.is_lib(row):
                libs.append(key)
            else:
                self.place_file(key, row)
        for key in tests + libs:
            self.place_dependent(key, self.rows[key])

    @staticmethod
    def is_test(row):
        return bool(TEST_NAME.match(os.path.basename(row["source"])))

    @staticmethod
    def is_lib(row):
        name = os.path.basename(row["source"]).lower()
        return ".lib." in name or "fixture" in name

    def subject_of(self, key):
        row = self.rows[key]
        match = TEST_NAME.match(os.path.basename(row["source"]))
        if not match:
            return None
        if match.group("stem2"):
            subject_name = match.group("stem2")
        else:
            subject_name = match.group("stem") + match.group("ext")
        subject = (row["root"], os.path.join(os.path.dirname(row["source"]), subject_name).replace(os.sep, "/"))
        return subject if subject in self.rows else None

    def place_dependent(self, key, row):
        subject = self.subject_of(key) if self.is_test(row) else None
        if subject:
            dest = self.rows[subject]["destination"]
            if dest:
                self.set_dest(key, "R1", os.path.join(os.path.dirname(dest), os.path.basename(row["source"])))
            else:
                self.set_review(key, "R1", "subject has no destination")
            return
        if self.is_test(row):
            related = self.cites.get(key, set())
        else:
            # a library is often sourced by name from its own folder, without a support path
            name = os.path.basename(row["source"])
            related = set(self.cited_by.get(key, set())) | {
                other for other in self.rows
                if other != key and other[0] == key[0] and self.rows[other]["kind"] == "file"
                and os.path.dirname(other[1]) == os.path.dirname(row["source"])
                and name in (self.content(*other) or "")
            }
        dirs = collections.Counter()
        for other in sorted(related):
            dest = self.rows[other]["destination"] if other in self.rows else None
            if dest and other[0] == key[0]:
                dirs[os.path.dirname(dest)] += 1
        if dirs:
            top = max(dirs.values())
            directory = sorted(d for d, n in dirs.items() if n == top)[0]
            self.set_dest(key, "R3", os.path.join(directory, os.path.basename(row["source"])))
            return
        self.place_file(key, row)

    def place_file(self, key, row):
        i = row["root"]
        root = self.roots[i]
        homes = self.homes(i)
        sub = row["subpath"]
        entry = row["entry"]
        infra_tree = entry != "scripts" and (
            entry in cl.IAC_FAMILY or any(r.endswith(".tf") for r in root.files if r.startswith(f"_support/{entry}/"))
        )
        if not infra_tree:
            return self.place_code(key, row)
        # R2: a tree of the infrastructure family moves whole
        if not root.is_git:
            return self.set_review(key, "R2", "no repository: infrastructure goes in a repository")
        if os.path.splitext(row["source"])[1].lower() in {".json", ".yaml", ".yml"} and not self.cited_anywhere(key):
            return self.set_review(key, "R2", "unreferenced: no tracked file cites it by path or name")
        if entry in IAC_HOMES:
            return self.set_dest(key, "R2", os.path.join(homes["infra"], sub))
        return self.set_dest(key, "R2", os.path.join(homes["infra"], entry, sub))

    def place_code(self, key, row):
        i = row["root"]
        root = self.roots[i]
        homes = self.homes(i)
        sub = row["subpath"]
        classes = self.callers.get(key, set())
        code = {c for c in classes if c != "doc"}
        packages = sorted(c[len("package:"):] for c in code if c.startswith("package:"))
        first = sub.split("/")[0] if "/" in sub else None
        by_name = [p for p in root.packages if os.path.basename(p) == first] if first else []
        if not root.is_git:
            if not classes and self.one_off(row):
                return self.set_dest(key, "R9", self.session_dest(row))
            return self.set_review(key, "R4-R8", "no repository: code goes in the repository that uses it")
        if code and all(c.startswith("package:") for c in code) and len(set(packages)) == 1:
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
        tokens = set(re.split(r"[-_.]", name.lower()))
        return bool(tokens & ONE_OFF_WORDS) or bool(KEY_RE.search(name))

    def session_dest(self, row):
        root = self.roots[row["root"]]
        date = root.first_commit_date(row["source"]) if row["tracked"] else None
        date = date or oldest_mtime(root.abs(row["source"]))
        stem = kebab(os.path.splitext(os.path.basename(row["source"]))[0])
        return os.path.join("_support/sessions", f"{date}-{stem}", os.path.basename(row["source"]))

    def cited_anywhere(self, key):
        i, rel = key
        name = os.path.basename(rel)
        for j in range(len(self.roots)):
            for other in self.readable_files(j):
                if (j, other) == key:
                    continue
                text = self.content(j, other)
                if text and (rel in text or name in text):
                    return True
        return False

    def place_folder(self, key, row, refs):
        i = row["root"]
        root = self.roots[i]
        homes = self.homes(i)
        name = row["entry"]
        base = row["source"]
        if name in LEGACY_EVIDENCE + LEGACY_DATA:
            citing = self.runtime_citations(name)
            if citing:
                return self.set_review(key, "F2", "runtime path: code or configuration cites it: " + ", ".join(citing[:5]))
        if name in LEGACY_EVIDENCE:
            del self.rows[key]
            self.place_evidence(i, base, name)
            return
        if name in LEGACY_DATA:
            return self.set_dest(key, "F4", f"_support/data/{oldest_mtime(root.abs(base))}-{kebab(name)}")
        files = [r for r in root.files if r.startswith(base + "/")]
        if files and all(os.path.splitext(r)[1].lower() in DOC_EXTENSIONS for r in files):
            return self.set_dest(key, "F5", os.path.join(homes["docs"], name))
        return self.set_review(key, "F6", "no rule for this folder; classify it by hand")

    def runtime_citations(self, name):
        found = []
        needle = f"_support/{name}"
        for j in range(len(self.roots)):
            for rel in self.readable_files(j):
                extension = os.path.splitext(rel)[1].lower()
                if extension not in cl.CODE_EXTENSIONS and extension not in RUNTIME_CONFIG_EXTENSIONS:
                    continue
                text = self.content(j, rel)
                if not text or needle not in text:
                    continue
                comment = cl.COMMENT_BY_EXTENSION.get(extension)
                for number, line in enumerate(text.splitlines(), 1):
                    if needle in line and not (comment and comment.match(line)):
                        found.append(f"{os.path.basename(self.roots[j].path)}/{rel}:{number}")
                        break
        return sorted(found)

    def place_evidence(self, i, base, name):
        root = self.roots[i]
        path = root.abs(base)
        loose = []
        for child in sorted(os.listdir(path)):
            child_rel = f"{base}/{child}"
            if os.path.isdir(root.abs(child_rel)):
                match = DATE_RE.search(child)
                date = match.group(0) if match else oldest_mtime(root.abs(child_rel))
                work = kebab(DATE_RE.sub("", child, count=1)) or kebab(name)
                files = [r for r in root.files if r.startswith(child_rel + "/")]
                key = (i, child_rel)
                self.rows[key] = {
                    "root": i, "source": child_rel, "kind": "folder", "tracked": any(r in root.tracked for r in files),
                    "subpath": "", "entry": name, "rule": None, "destination": None, "notes": [], "override": False,
                }
                if self.rows[key]["tracked"]:
                    self.set_review(key, "F3", "tracked: moving it into an ignored folder would drop it from Git; remove it from the tree with the archive tag instead")
                else:
                    self.set_dest(key, "F3", f"_support/workspace/{date}-{work}/evidence")
            elif child not in cl.IGNORED_NAMES:
                loose.append(child_rel)
        if loose:
            date = min(oldest_mtime(root.abs(r)) for r in loose)
            for rel in loose:
                key = (i, rel)
                self.rows[key] = {
                    "root": i, "source": rel, "kind": "file", "tracked": rel in root.tracked, "subpath": "",
                    "entry": name, "rule": None, "destination": None, "notes": [], "override": False,
                }
                if self.rows[key]["tracked"]:
                    self.set_review(key, "F3", "tracked: remove it from the tree with the archive tag instead")
                else:
                    self.set_dest(key, "F3", f"_support/workspace/{date}-legacy-{kebab(name)}/evidence/{os.path.basename(rel)}")

    def set_dest(self, key, rule, dest):
        row = self.rows[key]
        row["rule"] = rule
        row["destination"] = os.path.normpath(dest).replace(os.sep, "/")

    def set_review(self, key, rule, reason):
        row = self.rows[key]
        row["rule"] = rule
        row["destination"] = None
        self.review(row["root"], row["source"], rule, reason)

    # ---- overrides, collisions, notes ---------------------------------------------
    def apply_overrides(self):
        for spec, dest in self.overrides:
            hits = [k for k, r in self.rows.items()
                    if r["source"] == spec or self.roots[r["root"]].abs(r["source"]) == os.path.abspath(spec)]
            if not hits:
                raise SystemExit(fail(f"--override names no planned row: {spec}"))
            for key in hits:
                row = self.rows[key]
                row["destination"] = os.path.normpath(dest).replace(os.sep, "/")
                row["override"] = True
                row["rule"] = row["rule"] or "override"
                self.reviews = [r for r in self.reviews if not (r["root"] == key[0] and r["path"] == row["source"])]
        if self.homes_override:
            for key, row in self.rows.items():
                rule = row["rule"] or ""
                category = "tools" if rule in ("R5", "R6", "R7", "R8") else "infra" if rule == "R2" else "docs" if rule == "F5" else None
                if category in self.homes_override:
                    row["override"] = True
            for i, root in enumerate(self.roots):
                merged = dict(root.layout)
                merged.update(self.homes_override)
                parts = [f"{k} at {merged[k]}" for k in ("docs", "tools", "infra") if k in merged]
                self.layout_lines.append({"root": i, "line": "Layout: " + " · ".join(parts)})

    def mark_collisions(self):
        seen = collections.defaultdict(list)
        for key in sorted(self.rows):
            row = self.rows[key]
            if row["destination"]:
                seen[(row["root"], row["destination"])].append(key)
        for (i, dest), keys in seen.items():
            exists = os.path.lexists(self.roots[i].abs(dest))
            if exists or len(keys) > 1:
                for key in keys:
                    row = self.rows[key]
                    row["notes"].append("collision")
                    self.set_review(key, row["rule"], f"collision at {dest}")

    def mark_same_names(self):
        for key in sorted(self.rows):
            row = self.rows[key]
            if row["kind"] != "file":
                continue
            root = self.roots[row["root"]]
            stem = os.path.splitext(os.path.basename(row["source"]))[0]
            for rel in sorted(root.tracked):
                if rel.startswith("_support/") or HISTORY_RE.search(rel) or rel == row["source"]:
                    continue
                if os.path.splitext(os.path.basename(rel))[0] == stem and os.path.splitext(rel)[1] != os.path.splitext(row["source"])[1]:
                    row["notes"].append(f"same name elsewhere: {rel}")
                    break

    # ---- rewrites ----------------------------------------------------------------
    def new_abs(self, key):
        row = self.rows.get(key)
        if not row or not row["destination"]:
            return None
        return self.roots[row["root"]].abs(row["destination"])

    def build_rewrites(self, refs):
        files, folders = self.candidate_index()
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
                if new_target is None:
                    self.review(i, f"{ref['file']}:{ref['line']}", "prefix",
                                f"{ref['token']} covers files with different destinations: "
                                + ", ".join(sorted({os.path.dirname(self.rows[k]['destination']) if self.rows[k]['destination'] else 'review' for k in ref['keys']})))
                    continue
            citing_key = (i, ref["file"])
            citing_dir = os.path.dirname(root.abs(ref["file"]))
            if citing_key in self.rows and self.rows[citing_key]["destination"]:
                citing_dir = os.path.dirname(root.abs(self.rows[citing_key]["destination"]))
            new_token = self.render(ref["token"], ref["base"], new_target, root, citing_dir)
            if new_token != ref["token"]:
                self.rewrites.append({
                    "root": i, "file": ref["file"], "line": ref["line"], "col": ref["col"],
                    "sha256_line": sha256_bytes(ref["text"].encode()), "old": ref["token"], "new": new_token,
                })

    def prefix_destination(self, prefix, keys):
        base = None
        for key in keys:
            row = self.rows[key]
            if not row["destination"]:
                return None
            source = self.roots[row["root"]].abs(row["source"])
            rest = os.path.relpath(source, prefix)
            dest = self.roots[row["root"]].abs(row["destination"])
            if not dest.endswith(os.sep + rest) and rest != ".":
                return None
            candidate = dest[: len(dest) - len(rest) - 1] if rest != "." else dest
            if base is None:
                base = candidate
            elif base != candidate:
                return None
        # every listed file under the prefix must move with it
        moving = {self.roots[self.rows[k]["root"]].abs(self.rows[k]["source"]) for k in keys}
        for item in cl.list_files(prefix):
            if not any(item == m or item.startswith(m + os.sep) for m in moving):
                return None
        return base

    @staticmethod
    def render(token, base, new_target, root, citing_dir):
        """The new token, in the form the old one had."""
        star = token.find("*")
        head = token[:star] if star >= 0 else token
        stripped = head.rstrip("/")
        suffix = token[len(stripped):]
        dot = "./" if stripped.startswith("./") else ""
        if base == "var":
            var = re.match(r"^\$\{?\w+\}?/", stripped)
            return var.group(0) + os.path.relpath(new_target, root.path).replace(os.sep, "/") + suffix
        if base == "file":
            rel = os.path.relpath(new_target, citing_dir).replace(os.sep, "/")
            return (dot if not rel.startswith("../") else "") + rel + suffix
        return dot + os.path.relpath(new_target, root.path).replace(os.sep, "/") + suffix

    def internal_paths(self):
        files, _ = self.candidate_index()
        for key in sorted(self.rows):
            row = self.rows[key]
            if row["kind"] != "file" or not row["destination"]:
                continue
            root = self.roots[row["root"]]
            text = self.content(row["root"], row["source"]) if not NEVER_READ.search("/" + row["source"]) else None
            if text is None:
                text, _ = read_text_strict(root.abs(row["source"]))
            if not text or not ANCHOR_RE.search(text):
                continue
            old_dir = os.path.dirname(root.abs(row["source"]))
            new_dir = os.path.dirname(root.abs(row["destination"]))
            for number, line in enumerate(text.splitlines(), 1):
                tokens = {m.start(1) for m in TOKEN_RE.finditer(line)}
                for match in REL_LITERAL_RE.finditer(line):
                    if match.start(1) in tokens:
                        continue  # a support path token: the rewrites already carry it
                    literal = match.group(1)
                    before = os.path.normpath(os.path.join(old_dir, literal))
                    expected = self.new_abs(files[before]) if before in files else before
                    after = os.path.normpath(os.path.join(new_dir, literal))
                    if expected and after != expected:
                        self.review(row["root"], f"{row['source']}:{number}", "internal path",
                                    f"{literal} resolves to {os.path.relpath(before, root.path)} today and to "
                                    f"{os.path.relpath(after, root.path)} from {row['destination']}")

    def build_index_lines(self):
        for key in sorted(self.rows):
            row = self.rows[key]
            if row["rule"] in ("R4", "R5", "R6", "R7", "R8") and row["destination"] and not self.is_test(row):
                self.index_lines.append({"root": row["root"], "line": f"- `{row['destination']}` — <what it is for> · run: <command> · test: <its test or none>"})

    # ---- output --------------------------------------------------------------------
    def plan(self):
        self.collect()
        refs = self.references()
        self.classify(refs)
        self.apply_overrides()
        self.mark_collisions()
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
        return {
            "version": VERSION,
            "roots": [{"path": r.path, "head": r.head, "git": r.is_git} for r in self.roots],
            "rows": rows,
            "rewrites": sorted(self.rewrites, key=lambda r: (r["root"], r["file"], r["line"], r["col"])),
            "reviews": [json.loads(r) for r in reviews],
            "index_lines": self.index_lines,
            "layout_lines": self.layout_lines,
        }


def markdown(plan):
    lines = []
    for i, root in enumerate(plan["roots"]):
        rows = [r for r in plan["rows"] if r["root"] == i]
        lines.append(f"## {root['path']}")
        lines.append("")
        lines.append("| Source | Rule | Destination | Notes |")
        lines.append("| --- | --- | --- | --- |")
        for row in rows:
            dest = row["destination"] or "review"
            notes = "; ".join(row["notes"] + (["user override"] if row["override"] else []))
            lines.append(f"| `{row['source']}` | {row['rule']} | {dest if dest == 'review' else '`' + dest + '`'} | {notes} |")
        rewrites = [r for r in plan["rewrites"] if r["root"] == i]
        lines.append("")
        lines.append(f"Reference rewrites: {len(rewrites)} in {len({r['file'] for r in rewrites})} files.")
        reviews = [r for r in plan["reviews"] if r["root"] == i]
        if reviews:
            lines.append("")
            lines.append("For review:")
            lines.extend(f"- `{r['path']}` — {r['reason']}: {r['detail']}" for r in reviews)
        for item in plan["layout_lines"]:
            if item["root"] == i:
                lines.append("")
                lines.append(f"Index line to write: `{item['line']}`")
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
    plan = Planner(args.root, overrides, homes).plan()
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


def is_ignored(root, rel):
    result = git(root, "check-ignore", "-q", "--", rel)
    return result is not None and result.returncode == 0


def cmd_apply(args):
    plan = load(args.plan)
    errors = []
    for i, root in enumerate(plan["roots"]):
        if root["git"]:
            head = git(root["path"], "rev-parse", "HEAD")
            current = head.stdout.strip() if head is not None and head.returncode == 0 else None
            if current != root["head"]:
                errors.append(f"{root['path']}: HEAD is {current}, the plan was made at {root['head']}")
    moves = [r for r in plan["rows"] if r["destination"]]
    for row in moves:
        source = os.path.join(root_path(plan, row["root"]), row["source"])
        if fingerprint(source) != row["fingerprint"]:
            errors.append(f"{row['source']}: changed since the plan")
        if os.path.lexists(os.path.join(root_path(plan, row["root"]), row["destination"])):
            errors.append(f"{row['destination']}: destination exists")
    by_file = collections.defaultdict(list)
    for rewrite in plan["rewrites"]:
        by_file[(rewrite["root"], rewrite["file"])].append(rewrite)
    originals = {}
    for (i, rel), items in sorted(by_file.items()):
        path = os.path.join(root_path(plan, i), rel)
        data = read_bytes(path)
        try:
            lines = data.decode("utf-8").splitlines(True) if data is not None else None
        except UnicodeDecodeError:
            lines = None
        if lines is None:
            errors.append(f"{rel}: unreadable")
            continue
        originals[(i, rel)] = lines
        for item in items:
            if item["line"] > len(lines) or sha256_bytes(lines[item["line"] - 1].encode()) != item["sha256_line"]:
                errors.append(f"{rel}:{item['line']}: line changed since the plan")
    if errors:
        for error in errors:
            print(f"refused: {error}", file=sys.stderr)
        return 1
    undo = {"plan": os.path.abspath(args.plan), "rewrites": [], "moves": [], "removed_dirs": []}
    for (i, rel), items in sorted(by_file.items()):
        lines = originals[(i, rel)]
        for number in sorted({it["line"] for it in items}):
            line = lines[number - 1]
            for item in sorted((it for it in items if it["line"] == number), key=lambda it: -it["col"]):
                start = item["col"]
                if line[start:start + len(item["old"])] != item["old"]:
                    print(f"refused mid-way: {rel}:{number} token moved", file=sys.stderr)
                    return 1
                line = line[:start] + item["new"] + line[start + len(item["old"]):]
            undo["rewrites"].append({"root": i, "file": rel, "line": number, "before": lines[number - 1], "after": line})
            lines[number - 1] = line
        path = os.path.join(root_path(plan, i), rel)
        with open(path, "r+b") as handle:
            handle.seek(0)
            handle.write("".join(lines).encode("utf-8"))
            handle.truncate()
    for row in sorted(moves, key=lambda r: (r["root"], r["source"])):
        root = root_path(plan, row["root"])
        source = os.path.join(root, row["source"])
        dest = os.path.join(root, row["destination"])
        os.makedirs(os.path.dirname(dest), exist_ok=True)
        used_git = False
        if row["tracked"] and row["kind"] == "file" and not is_ignored(root, row["destination"]):
            result = git(root, "mv", "--", row["source"], row["destination"])
            used_git = result is not None and result.returncode == 0
            if not used_git:
                print(f"git mv failed for {row['source']}: {result.stderr.strip() if result else ''}", file=sys.stderr)
                return 1
        else:
            os.replace(source, dest)
        undo["moves"].append({"root": row["root"], "source": row["source"], "destination": row["destination"], "git": used_git})
        parent = os.path.dirname(source)
        while os.path.basename(parent) != "_support" and parent.startswith(root + os.sep):
            try:
                if os.listdir(parent):
                    break
                os.rmdir(parent)
                undo["removed_dirs"].append({"root": row["root"], "dir": os.path.relpath(parent, root)})
            except OSError:
                break
            parent = os.path.dirname(parent)
    undo_path = args.plan + ".undo.json"
    with open(undo_path, "w", encoding="utf-8") as handle:
        json.dump(undo, handle, indent=2, sort_keys=True, ensure_ascii=False)
    print(f"moved {len(undo['moves'])}, rewrote {len(undo['rewrites'])} lines; undo record: {undo_path}")
    for review in plan["reviews"]:
        print(f"review: {review['path']} — {review['reason']}: {review['detail']}")
    return 0


def cmd_verify(args):
    plan = load(args.plan)
    residual = []
    roots = [p["path"] for p in plan["roots"]]
    moved = collections.defaultdict(set)
    for row in plan["rows"]:
        if row["destination"]:
            moved[row["root"]].add(os.path.join(roots[row["root"]], row["source"]))
    gone = {i: {m for m in paths if not os.path.lexists(m)} for i, paths in moved.items()}
    gone_dirs = set()
    for paths in gone.values():
        for item in paths:
            parent = os.path.dirname(item)
            while parent not in roots and parent != os.path.dirname(parent):
                if not os.path.lexists(parent):
                    gone_dirs.add(parent)
                parent = os.path.dirname(parent)
    all_gone = set().union(*gone.values()) | gone_dirs if gone else set()
    planner = Planner(roots, [], {})
    for i in range(len(roots)):
        for rel in planner.readable_files(i):
            text = planner.content(i, rel)
            if not text or "_support/" not in text:
                continue
            for number, line in enumerate(text.splitlines(), 1):
                for match in TOKEN_RE.finditer(line):
                    targets = [t for _, t in planner.resolutions(i, rel, match.group(1).rstrip("."))]
                    if targets and not any(os.path.lexists(t) for t in targets) and any(t in all_gone for t in targets):
                        residual.append(f"{roots[i]}/{rel}:{number}: {match.group(1)}")
    for i, root in enumerate(plan["roots"]):
        tools = [r for r in plan["rows"] if r["root"] == i and r["rule"] in ("R4", "R5", "R6", "R7", "R8") and r["destination"]
                 and not TEST_NAME.match(os.path.basename(r["source"]))]
        if not tools:
            continue
        index = cl.find_index(root["path"])
        if not index:
            residual.append(f"{root['path']}: moved tools and no index; create _support/docs/README.md with a line per tool")
            continue
        text = cl.read_text(index)
        for row in tools:
            if os.path.basename(row["destination"]) not in text:
                residual.append(f"{index}: no line for {row['destination']}")
    for item in residual:
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
        os.makedirs(os.path.dirname(os.path.join(root, move["source"])), exist_ok=True)
        if move["git"]:
            result = git(root, "mv", "--", move["destination"], move["source"])
            if result is None or result.returncode != 0:
                conflicts.append(f"git mv back failed for {move['destination']}")
        else:
            os.replace(os.path.join(root, move["destination"]), os.path.join(root, move["source"]))
        parent = os.path.dirname(os.path.join(root, move["destination"]))
        while parent.startswith(root + os.sep):
            try:
                if os.listdir(parent):
                    break
                os.rmdir(parent)
            except OSError:
                break
            parent = os.path.dirname(parent)
    by_file = collections.defaultdict(list)
    for item in undo["rewrites"]:
        by_file[(item["root"], item["file"])].append(item)
    for (i, rel), items in sorted(by_file.items()):
        path = os.path.join(root_path(plan, i), rel)
        lines = read_bytes(path).decode("utf-8").splitlines(True)
        for item in items:
            if lines[item["line"] - 1] == item["after"]:
                lines[item["line"] - 1] = item["before"]
            else:
                conflicts.append(f"{rel}:{item['line']} changed after apply; left as is")
        with open(path, "r+b") as handle:
            handle.seek(0)
            handle.write("".join(lines).encode("utf-8"))
            handle.truncate()
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
    plan.add_argument("--home", action="append", metavar="KEY=PATH")
    plan.add_argument("--json", action="store_true")
    plan.add_argument("--out")
    for name in ("apply", "verify", "undo"):
        command = sub.add_parser(name)
        command.add_argument("--plan", required=True)
    args = parser.parse_args(argv)
    return {"plan": cmd_plan, "apply": cmd_apply, "verify": cmd_verify, "undo": cmd_undo}[args.command](args)


if __name__ == "__main__":
    sys.exit(main())
