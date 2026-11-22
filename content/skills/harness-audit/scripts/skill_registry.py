#!/usr/bin/env python3
"""Check a project's `## Skills` registry and propose official skills for its stack.

A registry line is a list item whose backticked token ending in `SKILL.md` is the
skill to read, resolved from the folder that holds the AGENTS.md:

    - Before editing files under `apps/mobile/`: `.agents/skills/expo-router/SKILL.md`

Subcommands:
  check <path>    Compare the registry of a repository, or of a declared workspace
                  and each of its child repositories, with the skills installed in
                  `.agents/skills`, `.claude/skills`, `.grok/skills` and
                  `skills-lock.json`. Skills that Hive installed (their folder holds
                  `assets/hive-managed.json`) are not counted. Exit 0 when they
                  match, 1 with findings, 2 when an input cannot be read.
  propose <repo>  Print each stack the repository's files show, with the file and
                  field that showed it, and its official skill sources from
                  `assets/stack-skills.json`. Exit 0, also when no stack is found.

The script only reads; it installs nothing and uses no network.
"""

import argparse
import importlib.util
import json
import pathlib
import re
import sys

SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
TABLE = SCRIPT_DIR.parent / "assets" / "stack-skills.json"
WORKSPACE_SCRIPT = SCRIPT_DIR.parents[1] / "workspace-repo" / "scripts" / "workspace_repo.py"

SKILL_DIRS = (".agents/skills", ".claude/skills", ".grok/skills")
LOCK = "skills-lock.json"
HIVE_MARKER = ("assets", "hive-managed.json")
HIVE_OWNER = "tricell-hive"

HEADING_RE = re.compile(r"^(#{1,6})\s+(.*?)\s*#*\s*$")
FENCE_RE = re.compile(r"^\s*(```|~~~)")
ITEM_RE = re.compile(r"^\s*[-*]\s+")
TICK_RE = re.compile(r"`([^`]+)`")

MAX_DEPTH = 4
SKIP_DIRS = {"node_modules", "Pods", "build", ".git", ".gradle", ".dart_tool", "DerivedData", "dist", ".next", "vendor", ".expo"}
WRAPPERS = ("flutter", "expo", "react-native", "kmp")


class Unreadable(Exception):
    pass


class Report:
    def __init__(self):
        self.findings = []
        self.errors = []
        self.notes = []

    def code(self):
        if self.errors:
            return 2
        return 1 if self.findings else 0


def _workspace_module():
    spec = importlib.util.spec_from_file_location("workspace_repo", WORKSPACE_SCRIPT)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def read_text(path):
    try:
        return path.read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError) as exc:
        raise Unreadable(f"{path}: cannot be read ({exc.__class__.__name__})")


def outside_fences(lines):
    """Yield (index, line) for the lines that are not inside a fenced code block."""
    fence = None
    for i, line in enumerate(lines):
        m = FENCE_RE.match(line)
        if m:
            if fence is None:
                fence = m.group(1)
            elif m.group(1) == fence:
                fence = None
            continue
        if fence is None:
            yield i, line


def sections(text):
    """Map each level-2 heading's title to its lines, ignoring headings inside fences."""
    lines = text.splitlines()
    out, current = {}, None
    for _, line in outside_fences(lines):
        h = HEADING_RE.match(line)
        if h and len(h.group(1)) <= 2:
            current = h.group(2).strip() if len(h.group(1)) == 2 else None
            if current is not None:
                out.setdefault(current, [])
            continue
        if current is not None:
            out[current].append(line)
    return out


def safe_relative(token):
    p = pathlib.PurePosixPath(token)
    return not (token.startswith(("/", "~", "\\")) or re.match(r"^[A-Za-z]:", token) or ".." in p.parts)


def registry_lines(text):
    """Return the registry entries: (skill token, other backticked tokens, line)."""
    entries = []
    for line in sections(text).get("Skills", []):
        if not ITEM_RE.match(line):
            continue
        tokens = TICK_RE.findall(line)
        skills = [t for t in tokens if t.endswith("SKILL.md")]
        for skill in skills:
            entries.append((skill, [t for t in tokens if t != skill], line.strip()))
    return entries


def misplaced(text):
    """True when the registry sits inside `## Hive` as a `### Skills` heading or a `Skills:` key."""
    for line in sections(text).get("Hive", []):
        if re.match(r"^###\s+Skills\s*$", line) or re.match(r"^\s*[-*]\s+Skills\s*:", line):
            return True
    return False


def hive_managed(folder):
    marker = folder.joinpath(*HIVE_MARKER)
    if not marker.is_file():
        return False
    try:
        return json.loads(marker.read_text(encoding="utf-8")).get("managed_by") == HIVE_OWNER
    except (OSError, ValueError, AttributeError):
        return False


def installed_skills(repo, report, label):
    """Names of installed skills that Hive did not install, from the skill folders."""
    names = {}
    for rel in SKILL_DIRS:
        base = repo / rel
        if base.is_symlink() and not base.exists():
            report.findings.append(f"{label}{rel}: broken link")
            continue
        if not base.is_dir():
            continue
        for entry in sorted(base.iterdir(), key=lambda p: p.name):
            if entry.name.startswith("."):
                continue
            if entry.is_symlink() and not entry.exists():
                report.findings.append(f"{label}{rel}/{entry.name}: broken link")
                continue
            if not entry.is_dir() or not (entry / "SKILL.md").is_file():
                continue
            if hive_managed(entry):
                continue
            names.setdefault(entry.name, f"{rel}/{entry.name}")
    return names


def lock_entries(repo):
    path = repo / LOCK
    if not path.exists():
        return []
    text = read_text(path)
    try:
        data = json.loads(text)
    except ValueError:
        raise Unreadable(f"{path}: not valid JSON (merge conflict markers or a broken edit)")
    if not isinstance(data, dict) or data.get("version") != 2 or not isinstance(data.get("skills"), dict):
        raise Unreadable(f"{path}: expected version 2 with a skills object")
    return sorted(data["skills"])


def check_repo(repo, report, label=""):
    """Check one repository; return its registry entries for the workspace comparison."""
    agents = repo / "AGENTS.md"
    installed = installed_skills(repo, report, label)
    try:
        locked = lock_entries(repo)
    except Unreadable as exc:
        report.errors.append(str(exc))
        locked = []
    for name in locked:
        if name not in installed and not any((repo / rel / name).is_dir() for rel in SKILL_DIRS):
            report.findings.append(f"{label}{LOCK}: {name} is in the lock but not installed")
    if not agents.is_file():
        if installed:
            report.findings.append(f"{label}AGENTS.md is missing; installed skills have no ## Skills registry: {', '.join(sorted(installed))}")
        return []
    try:
        text = read_text(agents)
    except Unreadable as exc:
        report.errors.append(str(exc))
        return []
    if misplaced(text):
        report.findings.append(f"{label}AGENTS.md: the skills registry is inside ## Hive; move it to its own ## Skills section")
    entries = registry_lines(text)
    listed = set()
    for token, _, line in entries:
        if not safe_relative(token):
            report.findings.append(f"{label}AGENTS.md: {token} must be a path relative to the AGENTS.md folder, without ~ or ..")
            continue
        path = repo / token
        if not path.is_file():
            report.findings.append(f"{label}AGENTS.md: {token} does not exist")
            continue
        listed.add(path.parent.name)
        listed.add(path.resolve().parent.name)
    for name in sorted(installed):
        if name not in listed:
            report.findings.append(f"{label}{installed[name]} is installed but no ## Skills line lists it")
    return entries


def norm(text):
    while text.startswith("./"):
        text = text[2:]
    return str(pathlib.PurePosixPath(text))


def check(path):
    report = Report()
    root = pathlib.Path(path)
    if not root.is_dir():
        report.errors.append(f"{root}: not a directory")
        return report
    agents = root / "AGENTS.md"
    rows = []
    if agents.is_file():
        try:
            rows = _workspace_module().parse_repositories(read_text(agents))
        except Unreadable as exc:
            report.errors.append(str(exc))
    root_entries = check_repo(root, report)
    if not rows:
        return report
    workspace_skills = {norm(token) for token, _, _ in root_entries}
    for row in rows:
        child_path = row.get("path", "").strip().strip("/")
        child = root / child_path
        if not child_path or not child.is_dir():
            report.notes.append(f"{child_path or '(empty path)'}: not cloned; skipped")
            continue
        label = f"{child_path}/"
        for token, _, line in check_repo(child, report, label):
            if not safe_relative(token):
                continue
            wanted = norm(f"{child_path}/{norm(token)}")
            if wanted not in workspace_skills:
                report.findings.append(f"AGENTS.md: ## Skills lacks {wanted}, from {label}AGENTS.md: {line}")
    return report


def load_table():
    data = json.loads(TABLE.read_text(encoding="utf-8"))
    return data


def walk(repo):
    """Yield files and folders up to MAX_DEPTH, skipping dependency and build folders."""
    stack = [(repo, 0)]
    while stack:
        folder, depth = stack.pop()
        try:
            entries = sorted(folder.iterdir(), key=lambda p: p.name)
        except OSError:
            continue
        for entry in entries:
            if entry.is_symlink():
                continue
            if entry.is_dir():
                if entry.name in SKIP_DIRS or entry.name.startswith("."):
                    continue
                yield entry
                if depth + 1 < MAX_DEPTH:
                    stack.append((entry, depth + 1))
            else:
                yield entry


def plugin_aliases(toml_text):
    """Map Gradle version-catalog plugin aliases (dotted) to plugin ids."""
    aliases, in_plugins = {}, False
    for line in toml_text.splitlines():
        s = line.strip()
        if s.startswith("["):
            in_plugins = s == "[plugins]"
            continue
        if not in_plugins or "=" not in s or s.startswith("#"):
            continue
        key, value = s.split("=", 1)
        m = re.search(r'id\s*=\s*"([^"]+)"', value) or re.match(r'\s*"([^":]+)', value)
        if m:
            aliases[re.sub(r"[-_]", ".", key.strip().strip('"'))] = m.group(1)
    return aliases


def gradle_plugins(text, aliases):
    ids = set(re.findall(r'id\s*\(?\s*["\']([\w.\-]+)["\']', text))
    if re.search(r'kotlin\s*\(\s*["\']multiplatform["\']', text):
        ids.add("org.jetbrains.kotlin.multiplatform")
    for alias in re.findall(r"alias\s*\(\s*libs\.plugins\.([\w.]+)\s*\)", text):
        if alias in aliases:
            ids.add(aliases[alias])
    return ids


def gradle_root(folder, repo):
    """The nearest folder from `folder` up to `repo` that holds a Gradle settings file."""
    for candidate in [folder, *folder.parents]:
        if any((candidate / s).is_file() for s in ("settings.gradle", "settings.gradle.kts")):
            return candidate
        if candidate == repo:
            break
    return folder


def detect(repo):
    """Return a list of (stack id, evidence, root folder) the repository's files show."""
    found = []
    files = list(walk(repo))
    aliases = {}
    for f in files:
        if f.name == "libs.versions.toml" and f.is_file():
            try:
                aliases.update(plugin_aliases(f.read_text(encoding="utf-8")))
            except (OSError, UnicodeDecodeError):
                pass
    for f in files:
        rel = f.relative_to(repo).as_posix()
        if f.is_dir():
            if f.suffix == ".xcodeproj":
                found.append(("ios", f"{rel}", f.parent))
            continue
        name = f.name
        try:
            if name == "pubspec.yaml":
                text = f.read_text(encoding="utf-8")
                if re.search(r"^\s*sdk:\s*flutter\s*$", text, re.M):
                    found.append(("flutter", f"{rel}: dependencies.flutter", f.parent))
                else:
                    found.append(("dart", f"{rel}", f.parent))
            elif name == "package.json":
                deps = json.loads(f.read_text(encoding="utf-8")).get("dependencies") or {}
                if not isinstance(deps, dict):
                    continue
                if "expo" in deps:
                    found.append(("expo", f"{rel}: dependencies.expo", f.parent))
                elif "react-native" in deps:
                    found.append(("react-native", f"{rel}: dependencies.react-native", f.parent))
                if "next" in deps:
                    found.append(("nextjs", f"{rel}: dependencies.next", f.parent))
                if "@nestjs/core" in deps:
                    found.append(("nestjs", f"{rel}: dependencies.@nestjs/core", f.parent))
            elif name in ("build.gradle", "build.gradle.kts", "settings.gradle", "settings.gradle.kts"):
                ids = gradle_plugins(f.read_text(encoding="utf-8"), aliases)
                if "org.jetbrains.kotlin.multiplatform" in ids:
                    found.append(("kmp", f"{rel}: plugin org.jetbrains.kotlin.multiplatform", gradle_root(f.parent, repo)))
                for pid in ("com.android.application", "com.android.library"):
                    if pid in ids:
                        found.append(("android", f"{rel}: plugin {pid}", f.parent))
                        break
            elif name == "Package.swift":
                found.append(("ios", f"{rel}", f.parent))
        except (OSError, UnicodeDecodeError, ValueError, AttributeError):
            continue
    wrapper_roots = [root for sid, _, root in found if sid in WRAPPERS]
    kept = []
    for sid, evidence, root in found:
        if sid in ("android", "ios") and any(root == w or w in root.parents for w in wrapper_roots):
            continue
        kept.append((sid, evidence))
    return kept


def propose(path):
    repo = pathlib.Path(path)
    if not repo.is_dir():
        print(f"error: {repo}: not a directory")
        return 2
    table = load_table()
    found = detect(repo)
    if not found:
        print("no recognized stack")
        return 0
    evidence = {}
    for sid, ev in found:
        evidence.setdefault(sid, []).append(ev)
    print(f"stack-skills.json reviewed {table['reviewed']}; installer: {table['installer']} (version chosen by check_package.py; pin each source to a commit)")
    for stack in table["stacks"]:
        sid = stack["id"]
        if sid not in evidence:
            continue
        print(f"stack: {stack['label']} ({'; '.join(evidence[sid])})")
        if not stack["sources"]:
            print(f"  no official skills: {stack.get('note', '')}".rstrip())
        for source in stack["sources"]:
            print(f"  source: {source['repo']} · docs: {source['docs']}")
        if stack["sources"] and stack.get("note"):
            print(f"  note: {stack['note']}")
    return 0


def main(argv=None):
    parser = argparse.ArgumentParser(description="Check a ## Skills registry or propose official skills for a stack.")
    sub = parser.add_subparsers(dest="command", required=True)
    sub.add_parser("check").add_argument("path")
    sub.add_parser("propose").add_argument("path")
    args = parser.parse_args(argv)
    if args.command == "propose":
        return propose(args.path)
    report = check(args.path)
    for line in report.errors:
        print(f"error: {line}")
    for line in report.findings:
        print(f"finding: {line}")
    for line in report.notes:
        print(f"note: {line}")
    if report.code() == 0:
        print("ok: the registry matches the installed skills")
    return report.code()


if __name__ == "__main__":
    sys.exit(main())
