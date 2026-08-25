#!/usr/bin/env python3
"""hive-compile.py — per-project hive profile generator.

Classifies a target repo from detectable facts (lockfiles, framework configs,
monorepo markers, infra files, repo-class markers) and writes an idempotent
managed block into the target's AGENTS.md between two single-line markers:

    <!-- hive-profile:start -->
    ...
    <!-- hive-profile:end -->

The stamp inside the block is ONE visible line (HTML comments are stripped only
by Claude Code — Codex, opencode and Grok inject them verbatim, so the two
minimal fences above are the only comment lines this block pays for).

Contract mirrors deploy-global: DRY-RUN is the default (prints the block),
`--apply` writes, `--create` allows creating a missing AGENTS.md, `--check`
exits non-zero when the target's existing block is stale (the hive moved past
the stamped SHA touching global/rules/ or this classifier).

Optional per-repo override file `.hive-profile.yaml` (flat `key: value` lines;
keys: class, tracker, exclusions, notes) — overrides win over detection. A
tracker line is emitted ONLY from an explicit override declaration: tracker
declarations are confirm-gated, the tool only mirrors one already made.

Usage:
    python3 harness/hive-compile.py <target-repo> [--apply] [--create] [--check] [--force]
    python3 harness/hive-compile.py --self-test   # idempotency + fail-closed proof

`--apply` refuses two unsafe states unless repaired (or, for the second, forced):
malformed fences (a lone start or end marker, duplicates, end before start) are
never rewritten over — repair by hand first; a target whose AGENTS.md no harness
would load (CLAUDE.md present without the `@AGENTS.md` import and no other
harness marker) is refused with the fix named — `--force` overrides.
"""
import datetime
import re
import subprocess
import sys
from pathlib import Path

HIVE_ROOT = Path(__file__).resolve().parent.parent
MARK_START = "<!-- hive-profile:start -->"
MARK_END = "<!-- hive-profile:end -->"
OVERRIDE_FILE = ".hive-profile.yaml"
# Paths whose movement makes an emitted profile stale (also used by the
# session-hygiene-report staleness advisory — keep the two in sync).
STALE_PATHS = ["global/rules/", "harness/hive-compile.py"]

# Exclusion set for classes with no runtime — mirrors the hand-written pattern
# of tricell-hive's AGENTS.md > Rule Exclusions.
NO_RUNTIME_EXCLUSIONS = [
    ("`quality/testing.md`", "no runtime code to test; files here are reviewed by reading"),
    ("`Build & Lint` (global `CLAUDE.md`)", "no build system; validation is read-review/diff review"),
    ("`security.md` > Supply Chain Security", "no installable dependencies; no OSV checks to run"),
    ("`patterns-antipatterns.md`", "code-pattern rules do not apply to markdown"),
    ("`critical-thinking.md` > Pre-ship ownership test (questions 1-2)",
     "no runtime load or customer impact; risk-surfacing and tradeoffs still apply"),
]


def sh(args, cwd=None):
    res = subprocess.run(args, cwd=cwd, capture_output=True, text=True)
    return res.returncode, res.stdout.strip()


def hive_sha():
    rc, out = sh(["git", "rev-parse", "--short", "HEAD"], cwd=HIVE_ROOT)
    if rc != 0:
        sys.exit("ERROR: cannot resolve the hive repo's HEAD (is this a git checkout?).")
    return out


def read_override(target: Path):
    """Flat `key: value` parser for .hive-profile.yaml — no external deps."""
    path = target / OVERRIDE_FILE
    if not path.is_file():
        return {}
    allowed = {"class", "tracker", "exclusions", "notes"}
    override = {}
    for ln, raw in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        m = re.match(r"^([A-Za-z_-]+)\s*:\s*(.+)$", line)
        if not m:
            sys.exit(f"ERROR: {path}:{ln}: not a flat 'key: value' line: {raw!r}")
        key, value = m.group(1).lower(), m.group(2).strip().strip("'\"")
        if key not in allowed:
            sys.exit(f"ERROR: {path}:{ln}: unknown key '{key}' (allowed: {sorted(allowed)})")
        override[key] = value
    return override


def globs(target: Path, *patterns, depth_dirs=("", "apps/*", "packages/*")):
    """Match patterns at the root and one level under common monorepo dirs."""
    hits = []
    for base in depth_dirs:
        for pattern in patterns:
            hits.extend(target.glob(f"{base}/{pattern}" if base else pattern))
    return hits


def pnpm_workspace_has_packages(path: Path):
    """pnpm ≥10 uses pnpm-workspace.yaml for general config too — the file is a
    monorepo marker only when it declares a non-empty `packages:` key."""
    try:
        text = path.read_text(encoding="utf-8")
    except OSError:
        return False
    m = re.search(r"^packages\s*:\s*(.*)$", text, re.M)
    if not m:
        return False
    inline = m.group(1).split("#", 1)[0].strip()
    if inline:
        return inline != "[]"
    for line in text[m.end():].splitlines():
        s = line.strip()
        if not s or s.startswith("#"):
            continue
        if not line[:1].isspace():
            return False          # next top-level key — the list was empty
        if s.startswith("- "):
            return True
    return False


def detect(target: Path):
    facts = {"stack": [], "rules": [], "infra": [], "monorepo": []}

    # Package manager: lockfile decides; absent one, packageManager; else pnpm.
    pkg_json = target / "package.json"
    pkg_text = pkg_json.read_text(encoding="utf-8") if pkg_json.is_file() else ""
    if (target / "pnpm-lock.yaml").is_file():
        facts["pm"] = "pnpm (pnpm-lock.yaml)"
    elif (target / "yarn.lock").is_file():
        facts["pm"] = "yarn (yarn.lock)"
    elif (target / "package-lock.json").is_file():
        facts["pm"] = "npm (package-lock.json)"
    elif pkg_json.is_file():
        m = re.search(r'"packageManager"\s*:\s*"([^@"]+)', pkg_text)
        facts["pm"] = f"{m.group(1)} (packageManager field)" if m else "pnpm (default — no lockfile)"
    else:
        facts["pm"] = None

    ws_yaml = target / "pnpm-workspace.yaml"
    if ws_yaml.is_file() and pnpm_workspace_has_packages(ws_yaml):
        facts["monorepo"].append("pnpm-workspace.yaml")
    for marker in ("turbo.json", "nx.json"):
        if (target / marker).is_file():
            facts["monorepo"].append(marker)

    def add(stack, rule):
        if stack and stack not in facts["stack"]:
            facts["stack"].append(stack)
        if rule and rule not in facts["rules"]:
            facts["rules"].append(rule)

    has_ts = pkg_json.is_file() or bool(globs(target, "tsconfig*.json"))
    if has_ts:
        add("TypeScript/JS", "typescript-standards")
    ui = False
    if globs(target, "next.config.*"):
        add("Next.js", "react-nextjs"); ui = True
    # Angular: angular.json (classic CLI), @angular/core in the root manifest,
    # or Nx project.json files (depth ≤2) naming @nx/angular or @angular
    # executors — Nx workspaces carry no angular.json.
    angular = bool(globs(target, "angular.json")) or '"@angular/core"' in pkg_text
    if not angular:
        for pj in globs(target, "project.json", depth_dirs=("", "*", "*/*")):
            if "node_modules" in pj.parts:
                continue
            try:
                pj_text = pj.read_text(encoding="utf-8")
            except OSError:
                continue
            if "@nx/angular" in pj_text or "@angular" in pj_text:
                angular = True
                break
    if angular:
        add("Angular", "angular-patterns"); ui = True
    if globs(target, "nest-cli.json"):
        add("NestJS", "nestjs-patterns")
    if globs(target, "astro.config.*"):
        # react-nextjs applies only when Astro actually hosts React islands.
        if '"@astrojs/react"' in pkg_text or '"react"' in pkg_text:
            add("Astro", "react-nextjs")
        else:
            add("Astro", None)
        ui = True
    if globs(target, "vite.config.*"):
        add("Vite", None); ui = True
    if globs(target, "schema.prisma", "prisma/schema.prisma", "drizzle.config.*"):
        add("Prisma/Drizzle", "sql-migrations")
    if globs(target, "tailwind.config.*") or '"tailwindcss"' in pkg_text:
        add("Tailwind", "tailwind")
    if (target / "pyproject.toml").is_file() or globs(target, "*.py", "scripts/*.py", "harness/*.py"):
        add("Python", "python-standards")
    if globs(target, "pom.xml", "build.gradle*"):
        add("Java/Kotlin", "java-kotlin")
    # Bounded lookup (no rglob: node_modules in a monorepo makes it minutes).
    if globs(target, "*.sh", "scripts/*.sh", "_support/scripts/*.sh", "global/hooks/*/*.sh"):
        add("shell scripts", "shell-standards")
    if ui:
        add(None, "ui-visual-design")

    if globs(target, "Dockerfile*"):
        facts["infra"].append("Dockerfile")
    if globs(target, "*.tf", "infrastructure/*.tf", "terraform/*.tf"):
        facts["infra"].append("Terraform")
    if (target / ".github" / "workflows").is_dir():
        facts["infra"].append("GitHub Actions")
    if facts["infra"]:
        facts["rules"].append("iac-devops")

    # Repo class.
    if (target / "global").is_dir() and (target / "harness").is_dir():
        facts["class"] = "config-hub"
    elif facts["monorepo"]:
        facts["class"] = "monorepo"
    elif pkg_json.is_file():
        facts["class"] = "app"
    else:
        # md-ratio over TEXT files only — images, binaries and .gitkeep dilute
        # the signal without saying anything about the repo's nature.
        non_text = {".webp", ".png", ".jpg", ".jpeg", ".svg", ".gif", ".ico", ".pdf", ".zip"}
        tracked = [p for p in target.rglob("*") if p.is_file()
                   and ".git" not in p.parts and "node_modules" not in p.parts
                   and p.suffix.lower() not in non_text and p.name != ".gitkeep"]
        md = [p for p in tracked if p.suffix == ".md"]
        ratio = len(md) / len(tracked) if tracked else 0.0
        # Name-signal tiebreak: a `<x>-specs` basename biases to specs from 0.4.
        threshold = 0.4 if target.name.endswith("-specs") else 0.6
        facts["class"] = "specs" if tracked and ratio >= threshold else "other"
    return facts


def render_block(facts, override):
    repo_class = override.get("class", facts["class"])
    today = datetime.date.today().isoformat()
    lines = [MARK_START,
             f"Hive profile v1 · hive@{hive_sha()} · {today} · class: {repo_class}",
             "",
             "## Hive Profile",
             ""]
    lines.append(f"- **Class:** {repo_class}"
                 + (f" ({', '.join(facts['monorepo'])})" if facts["monorepo"] else ""))
    if facts["pm"]:
        lines.append(f"- **Package manager:** {facts['pm']}")
    if facts["stack"]:
        lines.append(f"- **Stack detected:** {', '.join(facts['stack'])}")
    if facts["infra"]:
        lines.append(f"- **Infra:** {', '.join(facts['infra'])}")
    if override.get("tracker"):
        lines.append(f"- **Tracker (declared):** {override['tracker']}")
    if override.get("notes"):
        lines.append(f"- **Notes:** {override['notes']}")

    if facts["rules"]:
        lines += ["",
                  "**Path-scoped rule families that apply here** (load on touching matching files): "
                  + ", ".join(f"`{r}`" for r in facts["rules"]) + "."]

    no_runtime = repo_class in ("specs", "config-hub")
    if no_runtime or override.get("exclusions"):
        lines += ["", f"**Rule Exclusions (class: {repo_class} — no runtime):**"]
        if no_runtime:
            lines += [f"- {rule} — {why}." for rule, why in NO_RUNTIME_EXCLUSIONS]
        for extra in filter(None, (s.strip() for s in override.get("exclusions", "").split(","))):
            lines.append(f"- {extra}")
        lines += ["", "Build/test/lint enforcement is restored automatically in any repo with runtime code."]

    lines += ["",
              "**Creating the FIRST file of a kind in a session:** read its rule from "
              "`~/.claude/rules/languages/` first — path-scoped rules fire on read/edit, not on create.",
              MARK_END]
    return "\n".join(lines) + "\n"


def check_fences(agents_md: Path, text: str):
    """Fail-closed on any malformed block: a lone fence, duplicates, or end
    before start. Rewriting over an unparseable block risks regenerating from
    an empty base and destroying user prose — refuse and ask for a repair."""
    n_start, n_end = text.count(MARK_START), text.count(MARK_END)
    if n_start == 0 and n_end == 0:
        return
    problems = []
    if n_start != n_end:
        problems.append(f"{n_start}× start vs {n_end}× end fences")
    if n_start > 1 or n_end > 1:
        problems.append("duplicate fences")
    if n_start == 1 and n_end == 1 and text.index(MARK_START) > text.index(MARK_END):
        problems.append("end fence before start fence")
    if problems:
        sys.exit(f"ERROR: {agents_md} holds a malformed hive-profile block "
                 f"({'; '.join(problems)}) — repair the fences by hand first; "
                 "refusing to rewrite over an unparseable block.")


def upsert(agents_md: Path, block: str, create: bool):
    if not agents_md.is_file():
        if not create:
            sys.exit(f"ERROR: {agents_md} does not exist. Re-run with --create to create it, "
                     "or add the block to the file that repo actually loads.")
        agents_md.write_text(f"# Agent Configuration\n\n{block}", encoding="utf-8")
        return "created"
    text = agents_md.read_text(encoding="utf-8")
    check_fences(agents_md, text)
    if MARK_START in text:
        pattern = re.escape(MARK_START) + r".*?" + re.escape(MARK_END) + r"\n?"
        new = re.sub(pattern, block, text, count=1, flags=re.S)
        verdict = "unchanged" if new == text else "replaced"
    else:
        new = text.rstrip("\n") + "\n\n" + block
        verdict = "appended"
    if verdict != "unchanged":
        agents_md.write_text(new, encoding="utf-8")
    return verdict


def target_loads_agents_md(target: Path):
    """Would ANY harness load this repo's AGENTS.md?

    Codex/opencode/Grok read AGENTS.md natively; Claude Code reads it only
    through a CLAUDE.md that imports `@AGENTS.md`. The dead spot: a repo whose
    CLAUDE.md lacks the import and shows no other-harness marker — there the
    profile would be written where nothing reads it."""
    claude_md = target / "CLAUDE.md"
    if not claude_md.is_file():
        return True   # AGENTS.md is the repo's primary instruction doc
    if "@AGENTS.md" in claude_md.read_text(encoding="utf-8"):
        return True
    markers = (".codex", ".opencode", "opencode.jsonc", "opencode.json", ".grok")
    return any((target / m).exists() for m in markers)


def check_stale(agents_md: Path):
    if not agents_md.is_file():
        print(f"no AGENTS.md at {agents_md} — nothing to check.")
        return 0
    text = agents_md.read_text(encoding="utf-8")
    m = re.search(r"hive-profile:start -->\nHive profile v\d+ · hive@([0-9a-f]+)", text)
    if not m:
        print("no hive-profile block (or unparseable stamp) — nothing to check.")
        return 0
    stamp = m.group(1)
    current = hive_sha()
    if stamp == current:
        print(f"hive-profile fresh (hive@{stamp} == HEAD).")
        return 0
    rc, _ = sh(["git", "rev-parse", "--verify", f"{stamp}^{{commit}}"], cwd=HIVE_ROOT)
    if rc != 0:
        print(f"STALE: stamp hive@{stamp} is unknown to this hive checkout — regenerate: "
              "python3 harness/hive-compile.py <repo> --apply")
        return 1
    rc, out = sh(["git", "log", "--name-only", f"{stamp}..HEAD", "--"] + STALE_PATHS, cwd=HIVE_ROOT)
    if rc == 0 and out:
        print(f"STALE: hive moved {stamp}→{current} touching {' / '.join(STALE_PATHS)} — regenerate: "
              "python3 harness/hive-compile.py <repo> --apply")
        return 1
    print(f"hive-profile fresh enough (hive moved {stamp}→{current} without touching rules or the classifier).")
    return 0


def self_test():
    """Executable proof of the injection contract: second --apply is a
    byte-identical no-op, malformed fences refuse, target-loads detects the
    Claude-only dead spot. Runs against throwaway temp repos only."""
    import tempfile
    checks = 0

    def expect_exit(fn, label):
        nonlocal checks
        try:
            fn()
        except SystemExit as e:
            assert e.code, f"{label}: exited zero instead of refusing"
            checks += 1
            return
        raise AssertionError(f"{label}: did not refuse")

    with tempfile.TemporaryDirectory() as tmp:
        repo = Path(tmp) / "fake-app"
        repo.mkdir()
        (repo / "package.json").write_text('{"name": "fake"}\n', encoding="utf-8")
        agents = repo / "AGENTS.md"
        agents.write_text("# Fake repo\n\nUser prose stays intact.\n", encoding="utf-8")

        # Idempotency: first apply appends; second is a byte-identical no-op.
        v1 = upsert(agents, render_block(detect(repo), {}), create=False)
        assert v1 == "appended", f"first apply: {v1}"
        after1 = agents.read_bytes()
        v2 = upsert(agents, render_block(detect(repo), {}), create=False)
        after2 = agents.read_bytes()
        assert v2 == "unchanged", f"second apply: {v2}"
        assert after1 == after2, "second apply was not byte-identical"
        assert b"User prose stays intact." in after2, "user prose lost"
        checks += 3

        # Fail-closed fences: lone start, lone end, end-before-start, duplicates.
        for label, content in (
            ("lone-start", f"# X\n{MARK_START}\norphan\n"),
            ("lone-end", f"# X\norphan\n{MARK_END}\n"),
            ("end-before-start", f"# X\n{MARK_END}\nmid\n{MARK_START}\n"),
            ("duplicate-blocks", f"{MARK_START}\na\n{MARK_END}\n{MARK_START}\nb\n{MARK_END}\n"),
        ):
            agents.write_text(content, encoding="utf-8")
            expect_exit(lambda: upsert(agents, "BLOCK\n", create=False), label)
            assert agents.read_text(encoding="utf-8") == content, f"{label}: file was modified"

        # Target-loads: CLAUDE.md without the import and no harness marker -> not loaded;
        # adding the import (or a harness marker, or having no CLAUDE.md) -> loaded.
        assert target_loads_agents_md(repo), "no CLAUDE.md should count as loaded"
        (repo / "CLAUDE.md").write_text("# Standalone\n", encoding="utf-8")
        assert not target_loads_agents_md(repo), "importless CLAUDE.md should refuse"
        (repo / "opencode.jsonc").write_text("{}\n", encoding="utf-8")
        assert target_loads_agents_md(repo), "harness marker should count as loaded"
        (repo / "opencode.jsonc").unlink()
        (repo / "CLAUDE.md").write_text("@AGENTS.md\n", encoding="utf-8")
        assert target_loads_agents_md(repo), "@AGENTS.md import should count as loaded"
        checks += 4

    print(f"self-test: {checks} checks passed (idempotency, fail-closed fences, target-loads).")


def main():
    args = sys.argv[1:]
    flags = {a for a in args if a.startswith("--")}
    unknown = flags - {"--apply", "--create", "--check", "--dry-run", "--force", "--self-test"}
    if unknown:
        sys.exit(f"ERROR: unknown flag(s): {sorted(unknown)}\n{__doc__}")
    positional = [a for a in args if not a.startswith("--")]
    if "--self-test" in flags:
        if positional:
            sys.exit("ERROR: --self-test takes no target (it runs against temp repos).")
        self_test()
        return
    if len(positional) != 1:
        sys.exit(f"ERROR: exactly one target repo path required.\n{__doc__}")
    target = Path(positional[0]).resolve()
    if not target.is_dir():
        sys.exit(f"ERROR: target '{target}' is not a directory.")
    agents_md = target / "AGENTS.md"

    if "--check" in flags:
        sys.exit(check_stale(agents_md))

    override = read_override(target)
    facts = detect(target)
    block = render_block(facts, override)

    if "--apply" not in flags:
        print(f"# dry-run (default) — nothing written. Target: {agents_md}")
        print(block, end="")
    else:
        if not target_loads_agents_md(target) and "--force" not in flags:
            sys.exit(
                f"ERROR: {target}/CLAUDE.md does not import @AGENTS.md and no other "
                "harness marker (.codex/, .opencode/, opencode.json[c], .grok/) is "
                "present — the profile written to AGENTS.md would be loaded by NO "
                "harness in this repo. Fix: add an `@AGENTS.md` line to CLAUDE.md "
                "(the Claude Code import), then re-run; or pass --force to write anyway.")
        verdict = upsert(agents_md, block, create="--create" in flags)
        print(f"hive-profile {verdict} -> {agents_md}")

    repo_class = override.get("class", facts["class"])
    if repo_class in ("specs", "config-hub"):
        print()
        print("# optional manual step (never written by this tool): trim what CLAUDE.md pulls in")
        print(f"# for this no-runtime repo via {target}/.claude/settings.local.json:")
        print('#   { "claudeMdExcludes": ["rules/quality/testing.md", "rules/quality/patterns-antipatterns.md"] }')


if __name__ == "__main__":
    main()
