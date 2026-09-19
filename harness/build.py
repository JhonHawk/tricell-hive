#!/usr/bin/env python3
"""Regenerate every generated tree under harness/ from the canonical sources.

Single entry point (run from the repo root):
    python3 harness/build.py                    # assemble + regenerate + parity checks
    python3 harness/build.py --check            # parity checks only, nothing written
    python3 harness/build.py --bless-core-size  # deliberately re-bless the core-size ratchet

Core outputs are guarded at assembly time: a stale core (matches HEAD, differs
from the new assembly) regenerates silently; a hand-edited core (differs from
both) aborts the build with the edit preserved on disk.

Generates (delete-and-recreate, never incremental):
    global/CLAUDE.md              <- global/core-sections (assembled always-on core, Claude Code)
    harness/AGENTS.md             <- global/core-sections (assembled always-on core, Codex/opencode/Grok)
    harness/agents-skills/        <- global/skills   (cleaned universal skills)
    harness/claude/agents/        <- global/agents   (Claude Code agents, packs inlined)
    harness/rule-manifest.json    <- global/rules(-situational) (rule delivery map)
    harness/codex/agents/         <- global/agents   (Codex TOML subagents)
    harness/opencode/agents/      <- global/agents   (opencode markdown subagents)
    harness/grok/agents/          <- global/agents   (Grok Build markdown agents)
    harness/pi/agents/            <- global/agents   (PI pi-subagents agents)

Hand-written sources are never touched: harness/codex/{README, *.snippet},
harness/opencode/{README, *.snippet, commands/}. The two always-on cores are
GENERATED — edit global/core-sections/ (format: its README), never the outputs.

Run this after ANY edit to global/agents or global/skills, before committing —
/deploy-global also runs it and flags a dirty harness/ as a missed rebuild.
"""
import importlib.util
import json
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

# AGENTS.md is the always-on core shared by Codex, opencode and Grok — paid in
# full at the start of EVERY session, in every project. That cost, not a byte
# cap, is what these thresholds govern.
#
# The placement rule they enforce: the core holds SAFETY GATES and policy whose
# trigger is an ACTION rather than a file; the mechanics of how to carry that
# action out live behind the router skill, which the core points at. Gate and
# pointer here, mechanics there — never both, or the core pays twice for what
# the router already carries.
#
# There is NO size threshold on this file, by design. It had two (an editorial
# budget and a build-failing hard limit); both were justified by a truncation
# risk that does not exist, and in practice they turned every core edit into
# byte-counting — trimming prose that earned its place to fit a number nobody
# enforces. What governs growth is the PLACEMENT rule above (gate and pointer
# here, mechanics in the router), a quality question, not a size one. The build
# reports the file's size and its per-session token cost as a FACT, so the cost
# stays visible without pretending to be a gate.
# The core-size RATCHET (harness/.core-size-baseline) is NOT that budget coming
# back: it imposes no number of its own — it freezes the last BLESSED state, so
# growth is a deliberate, visible act (`--bless-core-size`) instead of drift,
# and it re-blesses downward automatically when a core shrinks.
# NO harness caps the global instruction file — verified 2026-08-18 in source
# and reproduced empirically, so neither threshold below is about truncation:
#   Codex — the global is read whole (`codex-home/src/instructions/mod.rs`) into
#     a field separate from the `project_doc_max_bytes` counter, which is
#     initialized AFTER the global is appended (`core/src/agents_md.rs`) and only
#     consumes the PROJECT chain (git-root → cwd). Reproduced: an 80,824 B project
#     AGENTS.md truncated at the 65,536 cap while the 30,322 B global stayed
#     intact — 96,087 B total, above the cap, so they share no budget.
#   opencode — read via `fs.readFileString()` and concatenated verbatim, no slice
#     or max_bytes anywhere in the pipeline; a request for a configurable cap was
#     closed "not planned".
#   Grok — documented ("no character cap and no truncation") with a source test
#     asserting full content, applied to user- and repo-level alike.
# The one REAL cap is Codex's `project_doc_max_bytes` over the PROJECT chain —
# which this repo hits only when a session opens under harness/ (see below).
CHAIN_CAP_BYTES = 32 * 1024  # Codex default; the conservative number for a public repo

# Router skills (Codex/opencode leg): canonical rule files injected as
# frontmatter-stripped references so each skill's routing table resolves.
# Keys are skill dir names under global/skills/; entries are (dir, glob)
# relative to global/. The skill body names the trigger; the reference IS
# the single canonical source — never fork content into the skill.
SKILL_REFERENCE_INJECTIONS = {
    "language-rules": [
        # Named one by one: rules-situational/ is a flat store shared with the
        # other routers, so a wildcard there would pull their references in too.
        ("rules-situational", "angular-patterns.md"),
        ("rules-situational", "iac-devops.md"),
        ("rules-situational", "identifier-language.md"),
        ("rules-situational", "java-kotlin.md"),
        ("rules-situational", "nestjs-patterns.md"),
        ("rules-situational", "python-standards.md"),
        ("rules-situational", "react-nextjs.md"),
        ("rules-situational", "shell-standards.md"),
        ("rules-situational", "sql-migrations.md"),
        ("rules-situational", "tailwind.md"),
        ("rules-situational", "typescript-standards.md"),
        ("rules-situational", "ui-visual-design.md"),
        ("rules/quality", "development-principles.md"),
        # Extracted out of development-principles.md + debugging.md, both injected
        # here: without this line Codex and opencode silently lose both sections.
        ("rules/quality", "reporting-integrity.md"),
        ("rules/quality", "testing.md"),
        ("rules/quality", "debugging.md"),
        ("rules-situational", "patterns-antipatterns.md"),
        ("rules-situational", "devops-principles.md"),
        ("rules/tools", "browser-automation.md"),
        # F2 demotion: the CLI mechanics moved out of the always-on rule; the
        # stub above keeps the gate, this file carries the reference.
        ("rules-situational", "browser-automation-reference.md"),
        ("rules/tools", "code-search.md"),
        # Claude Code and Grok get this always-on, but on Codex and opencode the
        # agent Role rules table is the ONLY pointer to the Context7 protocol —
        # and it pointed at a file this map never generated.
        ("rules/tools", "context7.md"),
    ],
    "workspace-conventions": [
        ("rules-situational", "project-structure.md"),
        ("rules-situational", "session-capture.md"),
        ("rules-situational", "support-artifacts.md"),
        ("rules-situational", "cross-service-workflow.md"),
        ("rules-situational", "infra-naming.md"),
    ],
    # Sources live in rules-situational/: their trigger is an intent (delegating,
    # planning), which `paths:` cannot express, so a router is their only channel.
    "task-routing": [
        ("rules-situational", "agent-routing.md"),
        ("rules-situational", "gap-resolution.md"),
        # F3: the git-mechanics router merged into task-routing; the reference
        # file keeps its name and section anchors.
        ("rules-situational", "git-mechanics.md"),
        # The core gates behavior on "trivial" (review scaling, test ritual) but
        # only this file defines it — and no router carried it, so Codex/opencode
        # judged the threshold with nothing to judge it by.
        ("rules/quality", "critical-thinking.md"),
    ],
    # /flow-plan writes a plan file, never project code, so the `task-routing`
    # trigger (first Write/Edit on project code) never fires during planning —
    # leaving the gap analysis, and above all the in-vivo prerequisite checklist,
    # reachable from no path the planning run takes.
    "flow-plan": [
        ("rules-situational", "gap-resolution.md"),
    ],
    "memory-policy": [
        ("rules-situational", "memory-routing.md"),
    ],
    "flow-report": [
        ("rules/quality", "communication-format.md"),
        ("rules-situational", "communication-format-mechanics.md"),
    ],
    "unattended-delegation": [
        ("rules-situational", "unattended-autonomy-mode.md"),
    ],
}

# The two always-on cores are assembled from one canonical section corpus, so a
# policy edit touches exactly one file. Section format (order/targets/join
# frontmatter, per-target order overrides, `<name>.<target>.md` body overlays):
# global/core-sections/README.md.
CORE_SECTIONS_DIR = ROOT / "global" / "core-sections"
CORE_TARGETS = {
    "claude": ROOT / "global" / "CLAUDE.md",
    "agents": ROOT / "harness" / "AGENTS.md",
}
CORE_MARKER = "<!-- generated: global/core-sections -->"
CORE_NOTICE = ("> Generated from `global/core-sections/` by `harness/build.py` — "
               "edit the sections, not this file.")
CORE_OVERLAY_SUFFIXES = tuple(f".{t}.md" for t in CORE_TARGETS)


def _core_section_files():
    """Primary section files, after failing closed on structural defects."""
    if not CORE_SECTIONS_DIR.is_dir():
        sys.exit("ERROR: core assembly: global/core-sections/ does not exist — "
                 "refusing to regenerate the always-on cores from nothing.")
    files = sorted(p for p in CORE_SECTIONS_DIR.glob("*.md") if p.name != "README.md")
    primaries = [p for p in files if not p.name.endswith(CORE_OVERLAY_SUFFIXES)]
    if not primaries:
        sys.exit("ERROR: core assembly: global/core-sections/ holds no section files — "
                 "refusing to write empty cores.")
    for stray in CORE_SECTIONS_DIR.iterdir():
        if stray.is_file() and stray.suffix != ".md" and not stray.name.startswith("."):
            print(f"WARNING: core assembly: {stray.name} is not a .md file and is "
                  "ignored by the assembly.")
    names = {p.stem for p in primaries}
    for overlay in (p for p in files if p.name.endswith(CORE_OVERLAY_SUFFIXES)):
        stem, target = overlay.stem.rsplit(".", 1)
        if stem not in names:
            sys.exit(f"ERROR: core assembly: overlay {overlay.name} has no primary "
                     f"section {stem}.md.")
        primary_targets = _parse_core_section(CORE_SECTIONS_DIR / f"{stem}.md")[0]
        if target not in primary_targets:
            sys.exit(f"ERROR: core assembly: overlay {overlay.name} targets {target!r} "
                     f"but {stem}.md declares targets: {primary_targets} — the overlay "
                     "is dead. Add the target to the primary or delete the overlay.")
    return primaries


def _parse_core_section(path):
    """-> (targets, {target: order}, join, body). Exits on malformed frontmatter."""
    text = path.read_text(encoding="utf-8")
    m = re.match(r"^---\n(.*?)\n---\n", text, re.S)
    if not m:
        sys.exit(f"ERROR: core assembly: {path.name} lacks a frontmatter block.")
    meta = {}
    for raw in m.group(1).splitlines():
        line = raw.split("#", 1)[0].strip()
        if not line:
            continue
        key, sep, val = line.partition(":")
        if not sep or not val.strip():
            sys.exit(f"ERROR: core assembly: malformed frontmatter line {raw!r} "
                     f"in {path.name}.")
        meta[key.strip()] = val.strip()
    targets = [t.strip() for t in meta.get("targets", "").strip("[]").split(",")
               if t.strip()]
    if not targets or any(t not in CORE_TARGETS for t in targets):
        sys.exit(f"ERROR: core assembly: {path.name} needs `targets:` as a non-empty "
                 f"subset of {sorted(CORE_TARGETS)} (got {targets or 'none'}).")
    orders = {}
    for t in targets:
        v = meta.get(f"order_{t}", meta.get("order"))
        if v is None or not v.isdigit():
            sys.exit(f"ERROR: core assembly: {path.name} has no integer order for "
                     f"target {t!r} — set `order:` or `order_{t}:`.")
        orders[t] = int(v)
    join = meta.get("join", "loose")
    if join not in ("loose", "tight"):
        sys.exit(f"ERROR: core assembly: {path.name} `join:` must be 'loose' or "
                 f"'tight' (default 'loose'; got {join!r}).")
    return targets, orders, join, text[m.end():].strip("\n")


def _assemble_core(target):
    parts = []
    for path in _core_section_files():
        targets, orders, join, body = _parse_core_section(path)
        if target not in targets:
            continue
        overlay = path.with_name(f"{path.stem}.{target}.md")
        if overlay.exists():
            otext = overlay.read_text(encoding="utf-8")
            om = re.match(r"^---\n.*?\n---\n", otext, re.S)
            body = (otext[om.end():] if om else otext).strip("\n")
        parts.append((orders[target], path.stem, join, body))
    seen = {}
    for order, stem, _, _ in parts:
        if order in seen:
            sys.exit(f"ERROR: core assembly: {seen[order]}.md and {stem}.md both use "
                     f"order {order} for target {target!r} — assembly position would "
                     "be filename-dependent. Renumber one of them.")
        seen[order] = stem
    parts.sort(key=lambda p: (p[0], p[1]))
    out = []
    for _, _, join, body in parts:
        if out:
            out.append("\n" if join == "tight" else "\n\n")
        out.append(body)
    return f"{CORE_MARKER}\n{CORE_NOTICE}\n\n{''.join(out)}\n"


def _git_head_content(rel):
    """Content of the file at HEAD, or None (new file, or no repo)."""
    try:
        r = subprocess.run(["git", "-C", str(ROOT), "show", f"HEAD:{rel}"],
                           capture_output=True, text=True)
        return r.stdout if r.returncode == 0 else None
    except OSError:
        return None


def assemble_cores():
    """Write the assembled cores — refusing to destroy a hand edit.

    Per output: disk == assembly -> up to date, nothing written. Disk differs
    from the assembly but matches HEAD -> stale after a section edit,
    regenerated silently. Disk differs from BOTH -> a hand edit (or an
    uncommitted state the build cannot vouch for): exit non-zero with the edit
    preserved on disk, never overwrite it.
    """
    for target, dest in CORE_TARGETS.items():
        content = _assemble_core(target)
        rel = dest.relative_to(ROOT)
        on_disk = dest.read_text(encoding="utf-8") if dest.exists() else None
        if on_disk == content:
            print(f"core up to date: {rel} (target: {target})")
            continue
        if on_disk is not None and on_disk != _git_head_content(rel):
            sys.exit(
                f"ERROR: core assembly: {rel} differs from BOTH its regeneration and "
                "HEAD — it holds a hand edit this rebuild would destroy. The edit is "
                "still on disk: move it into global/core-sections/, restore the output "
                f"(git checkout -- {rel}), and rerun harness/build.py. If the file is "
                "itself an uncommitted regeneration with no hand edits, delete it and "
                "rerun.")
        dest.write_text(content, encoding="utf-8")
        print(f"assembled {rel} <- global/core-sections/ "
              f"({len(content.encode('utf-8')) / 1024:.1f} KiB, target: {target})")


def check_core_assembly_parity():
    """Fail when a core on disk differs from its regeneration. Read-only.

    In a full build assemble_cores() ran first, so this is a belt check for
    write integrity and determinism; its real teeth are the --check mode
    (deploy preflight, pre-commit), which runs it WITHOUT assembling — there
    it catches both a stale output and a hand-edited one.
    """
    for target, dest in CORE_TARGETS.items():
        expected = _assemble_core(target)
        actual = dest.read_text(encoding="utf-8") if dest.exists() else ""
        if actual != expected:
            sys.exit(f"ERROR: core-assembly parity: {dest.relative_to(ROOT)} does not "
                     "match the assembly of global/core-sections/. It is a generated "
                     "file — run python3 harness/build.py to regenerate (that run "
                     "refuses to destroy a hand edit); never hand-edit the output.")
    print(f"core assembly: {len(CORE_TARGETS)} generated cores match "
          "global/core-sections/.")


CORE_SIZE_BASELINE = ROOT / "harness" / ".core-size-baseline"

CORE_SIZE_BASELINE_HEADER = """\
# Core-size ratchet baseline — blessed per-session token counts (bytes/4) for
# the generated always-on cores. The build FAILS when a core exceeds its
# blessed count; update deliberately with:
#     python3 harness/build.py --bless-core-size
# Shrinking re-blesses automatically on the next full build (the ratchet only
# moves down on its own).
#
# What this does NOT catch: it counts tokens, not quality. A policy loss that
# SHRINKS the core blesses itself silently, and growth that earned its place
# still fails until blessed. The ratchet freezes the blessed state — it
# imposes no number and no budget (see the header note in harness/build.py).
"""


def _core_token_counts():
    return {
        dest.relative_to(ROOT).as_posix(): (dest.stat().st_size if dest.exists() else 0) // 4
        for dest in CORE_TARGETS.values()
    }


def _read_core_size_baseline():
    if not CORE_SIZE_BASELINE.exists():
        return None
    blessed = {}
    for raw in CORE_SIZE_BASELINE.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        parts = line.split()
        if len(parts) != 2 or not parts[1].isdigit():
            sys.exit(f"ERROR: core-size ratchet: malformed baseline line {raw!r} in "
                     f"{CORE_SIZE_BASELINE.relative_to(ROOT)} — expected '<path> <tokens>'.")
        blessed[parts[0]] = int(parts[1])
    return blessed


def _write_core_size_baseline(counts):
    body = "".join(f"{path} {tokens}\n" for path, tokens in sorted(counts.items()))
    CORE_SIZE_BASELINE.write_text(CORE_SIZE_BASELINE_HEADER + body, encoding="utf-8")


def check_core_size_ratchet(allow_write=True):
    """Fail when a generated core EXCEEDS its blessed token count.

    Not a budget: the ratchet imposes no number — it freezes the blessed state
    so growth is a deliberate act (`--bless-core-size`), never drift. Shrinking
    re-blesses in a full build (allow_write) and passes read-only in --check.
    """
    current = _core_token_counts()
    blessed = _read_core_size_baseline()
    if blessed is None:
        sys.exit(f"ERROR: core-size ratchet: {CORE_SIZE_BASELINE.relative_to(ROOT)} is "
                 "missing — bless the current sizes with: "
                 "python3 harness/build.py --bless-core-size")
    failures = []
    shrank = False
    for path, tokens in sorted(current.items()):
        b = blessed.get(path)
        if b is None:
            failures.append(f"  {path}: ~{tokens:,} tokens, no blessed entry")
        elif tokens > b:
            failures.append(f"  {path}: ~{tokens:,} tokens vs blessed ~{b:,} "
                            f"(+{tokens - b:,})")
        elif tokens < b:
            shrank = True
    if failures:
        sys.exit("ERROR: core-size ratchet: a core exceeds its blessed size:\n"
                 + "\n".join(failures)
                 + "\nGrowth is a deliberate act: re-audit placement (gate and pointer "
                 "in the core, mechanics in the router), then bless the new state "
                 "with: python3 harness/build.py --bless-core-size")
    if shrank and allow_write:
        _write_core_size_baseline(current)
        print("core-size ratchet: cores shrank — baseline re-blessed downward.")
    else:
        print("core-size ratchet: cores within blessed sizes.")


GENERATED_README = """# Generated — do not edit

Everything in this directory is derived from canonical sources in `global/`
by `harness/build.py`. Edit the canonical file and rebuild; hand edits here
are overwritten on the next build.
"""

GENERATED_TREE_RELATIVE_PATHS = (
    Path("harness/agents-skills"),
    Path("harness/codex/agents"),
    Path("harness/opencode/agents"),
    Path("harness/grok/agents"),
    Path("harness/pi/agents"),
    Path("harness/claude"),
)

# Generated files that belong to no tree. Same parity treatment: regenerated to
# a temporary root and compared byte-for-byte.
GENERATED_FILE_RELATIVE_PATHS = (
    Path("harness/rule-manifest.json"),
)

# --- Rule manifest -------------------------------------------------------
#
# One entry per rule TEXT, for the consumers that deliver rule text without
# Claude Code's native path-scoping: the rule-delivery hook (by glob, per
# touched file) and the pack inliner (by name). The manifest is derived, never
# authored — every field is read off the rule file or off
# SKILL_REFERENCE_INJECTIONS above, so a rule that moves or changes its globs
# cannot leave a consumer pointing at the old shape.
RULE_MANIFEST_RELATIVE_PATH = Path("harness/rule-manifest.json")
RULE_SOURCE_DIRS = (Path("global/rules"), Path("global/rules-situational"))
RULE_MANIFEST_NOTE = ("Generated by harness/build.py from global/rules/** and "
                      "global/rules-situational/** — do not edit by hand.")
# Deployed roots for an injected reference. Which one a harness reads:
#   ~/.claude/skills — Claude Code (deployed there) and Grok (scans it)
#   ~/.agents/skills — Codex, opencode and PI (shared universal skills)
RULE_REFERENCE_ROOTS = {
    "claude": "~/.claude/skills",
    "agents": "~/.agents/skills",
}
# Rules a READER needs, not only a writer: a reviewer opening `_support/**` has
# to know where artifacts belong before it can judge placement. Every other
# rule is delivered on a write, so a read-only agent never receives it.
RULE_READER_NAMES = frozenset({"project-structure.md", "session-capture.md"})
# `paths:` is Claude Code's key and the ONLY one it reads; `globs:` is the
# store's spelling (global/rules-situational/), which no harness reads natively.
# The manifest emits one `globs` list from whichever key the file uses, but only
# `paths:` decides always-on — and `globs:` inside global/rules/ is an error.


def _rule_frontmatter(text: str) -> str:
    """The first frontmatter block, or '' when the file opens with content."""
    match = re.match(r"^---\n(.*?)\n---\n", text, re.S)
    return match.group(1) if match else ""


def _split_outside_braces(value: str):
    """Split on commas that are NOT inside a `{…}` group.

    A glob's brace group is itself comma-separated (`**/*.{ts,tsx,mts}`), so a
    naive split shreds one glob into three broken ones — and the manifest then
    scopes a rule to patterns that match nothing, silently.
    """
    parts, depth, current = [], 0, []
    for char in value:
        if char == "{":
            depth += 1
        elif char == "}":
            depth = max(depth - 1, 0)
        if char == "," and depth == 0:
            parts.append("".join(current))
            current = []
        else:
            current.append(char)
    parts.append("".join(current))
    return [p.strip().strip("\"'") for p in parts if p.strip()]


def _declared_globs(frontmatter: str, key: str):
    """Globs under one frontmatter key, in file order. None when absent."""
    block = re.search(rf"^{key}:[ \t]*(.*)$((?:\n[ \t]+[-#].*)*)", frontmatter, re.M)
    if not block:
        return None
    inline, items = block.group(1).strip(), block.group(2)
    if inline:
        # Single-line form: `paths: "**/*.ts"` or `paths: ["a", "b"]`.
        return _split_outside_braces(inline.strip("[]"))
    return [line.split("-", 1)[1].strip().strip("\"'")
            for line in items.splitlines()
            if line.strip().startswith("-")]


def _rule_reference_skill(relative: Path) -> str | None:
    """The router skill whose `references/` receives this rule, or None.

    A rule injected into several skills (gap-resolution reaches task-routing
    and flow-plan) resolves to the FIRST in declaration order: the manifest
    names one deployed path per root, and any of them is the same text.
    """
    for skill_name, sources in SKILL_REFERENCE_INJECTIONS.items():
        for subdir, pattern in sources:
            if (relative.parent == Path("global") / subdir
                    and relative.match(pattern)):
                return skill_name
    return None


def _load_agent_converter():
    """The agent converter as a module: one parser, one set of validation rules."""
    spec = importlib.util.spec_from_file_location(
        "hive_convert_agents", ROOT / "harness" / "build" / "convert-agents.py"
    )
    converter = importlib.util.module_from_spec(spec)
    # Importing a source file writes harness/build/__pycache__ as a side
    # effect. The build's job is to produce declared outputs and nothing else —
    # a stray directory in a tree the pre-commit hook stages wholesale is not a
    # thing to leave to .gitignore.
    dont_write_bytecode = sys.dont_write_bytecode
    sys.dont_write_bytecode = True
    try:
        spec.loader.exec_module(converter)
    finally:
        sys.dont_write_bytecode = dont_write_bytecode
    return converter


def _agent_packs(source_root: Path) -> dict:
    """Agent name -> declared packs, for the agents that declare any.

    The rule-delivery hook must not push a rule at an agent already carrying it
    inlined, and `packs:` is stripped from every generated output — so without
    this section the build's own knowledge dies with the render.
    """
    agents_dir = source_root / "global" / "agents"
    if not agents_dir.is_dir():
        return {}
    converter = _load_agent_converter()
    packs = {}
    for path in sorted(agents_dir.rglob("*.md")):
        agent = converter.parse_agent(path)
        if agent["packs"]:
            packs[agent["name"]] = agent["packs"]
    return dict(sorted(packs.items()))


def _read_only_agents(source_root: Path) -> list:
    """Agents that cannot write, by the converter's own capability verdict.

    The rule-delivery hook gives these only the rules a READER needs
    (`readers`), and only on a read: every other rule is delivered on a write
    they will never make. Derived, never authored — an agent that gains or
    loses Write/Edit changes lists with the next build.
    """
    agents_dir = source_root / "global" / "agents"
    if not agents_dir.is_dir():
        return []
    converter = _load_agent_converter()
    names = []
    for path in sorted(agents_dir.rglob("*.md")):
        agent = converter.parse_agent(path)
        if not converter.can_write(agent):
            names.append(agent["name"])
    return sorted(names)


def build_rule_manifest(source_root: Path = ROOT) -> dict:
    """Describe every rule text: where it lives, what it scopes to, who ships it."""
    entries = []
    for rules_dir in RULE_SOURCE_DIRS:
        in_core = rules_dir == Path("global/rules")
        for path in sorted((source_root / rules_dir).rglob("*.md")):
            if path.name == "README.md":
                continue
            relative = path.relative_to(source_root)
            frontmatter = _rule_frontmatter(path.read_text(encoding="utf-8"))
            if _declared_globs(frontmatter, "paths") is not None:
                sys.exit(
                    f"ERROR: rule manifest: {relative.as_posix()} declares "
                    f"`paths:`. That was Claude Code's native path-scoping key "
                    f"and the mechanism is retired — rules are delivered by the "
                    f"rule-delivery hook, off `globs:`. A file still carrying it "
                    f"reads as scoped while loading unconditionally. Use "
                    f"`globs:` in global/rules-situational/, or nothing at all "
                    f"in global/rules/, which is always-on by definition."
                )
            store_globs = _declared_globs(frontmatter, "globs")
            if in_core and store_globs is not None:
                sys.exit(
                    f"ERROR: rule manifest: {relative.as_posix()} declares "
                    f"`globs:`, but global/rules/ loads unconditionally — the "
                    f"file would claim a scope nothing applies. Drop the key, or "
                    f"move the rule to global/rules-situational/, which is "
                    f"reachable only through a router skill or this hook."
                )
            globs = store_globs or []
            skill = _rule_reference_skill(relative)
            entries.append({
                "name": path.stem,
                "source": relative.as_posix(),
                "globs": globs,
                # Rules whose globs overlap because two frameworks share a file
                # convention (`*.service.ts`). Carrying the named pack proves
                # which framework this is, and this rule is then not delivered.
                "exclusive_with": _declared_globs(frontmatter, "exclusive-with") or [],
                # With `paths:` retired, the directory IS the criterion:
                # global/rules/ loads unconditionally, and a rules-situational
                # file is never always-on — nothing loads that directory by
                # itself — however its `globs:` read.
                "always_on": in_core,
                "references": {
                    root: f"{base}/{skill}/references/{path.name}"
                    for root, base in RULE_REFERENCE_ROOTS.items()
                } if skill else None,
                "readers": path.name in RULE_READER_NAMES,
            })
    entries.sort(key=lambda entry: (entry["name"], entry["source"]))
    # `exclusive-with` is a cross-reference between rule files, and a typo in it
    # fails silently forever: the hook keeps delivering a rule nobody meant it
    # to deliver. Resolve every name against the set the build just read.
    known = {entry["name"] for entry in entries}
    for entry in entries:
        for name in entry["exclusive_with"]:
            if name not in known:
                sys.exit(
                    f"ERROR: rule manifest: {entry['source']} declares "
                    f"`exclusive-with: {name}`, which is not a rule under "
                    f"{' or '.join(d.as_posix() for d in RULE_SOURCE_DIRS)}. "
                    f"Name the rule file's stem, without the extension."
                )
    return {
        "_generated": RULE_MANIFEST_NOTE,
        "agents": _agent_packs(source_root),
        "read_only_agents": _read_only_agents(source_root),
        "rules": entries,
    }


def write_rule_manifest(output_root: Path, source_root: Path = ROOT):
    """Emit the manifest deterministically — key order and indent are fixed."""
    destination = output_root / RULE_MANIFEST_RELATIVE_PATH
    destination.parent.mkdir(parents=True, exist_ok=True)
    manifest = build_rule_manifest(source_root)
    payload = json.dumps(manifest, indent=2, ensure_ascii=False)
    destination.write_text(payload + "\n", encoding="utf-8")
    scoped = sum(1 for entry in manifest["rules"] if entry["globs"])
    print(f"rule manifest: {len(manifest['rules'])} rules ({scoped} glob-scoped) -> "
          f"{RULE_MANIFEST_RELATIVE_PATH}")


def regen_dir(path: Path):
    if path.exists():
        shutil.rmtree(path)
    path.mkdir(parents=True)


def generate_generated_trees(output_root: Path, source_root: Path = ROOT):
    """Generate every non-core derived tree beneath output_root.

    source_root supplies the canonical inputs and converter scripts. Keeping the
    output root independent lets --check use this exact pipeline in a temporary
    directory without touching the working tree.
    """
    build_dir = source_root / "harness" / "build"

    # Universal skills (Codex + opencode read ~/.agents/skills)
    skills_out = output_root / "harness" / "agents-skills"
    regen_dir(skills_out)
    subprocess.run(
        [sys.executable, str(build_dir / "convert-skills.py"),
         str(source_root / "global" / "skills"), str(skills_out)],
        check=True,
    )
    (skills_out / "README.md").write_text(GENERATED_README, encoding="utf-8")

    # Per-harness agents
    with tempfile.TemporaryDirectory() as tmp:
        subprocess.run(
            [sys.executable, str(build_dir / "convert-agents.py"),
             str(source_root / "global" / "agents"), tmp],
            check=True,
        )
        for src_name, relative in (
            ("codex", Path("harness/codex/agents")),
            ("opencode", Path("harness/opencode/agents")),
            ("grok", Path("harness/grok/agents")),
            ("pi", Path("harness/pi/agents")),
        ):
            destination = output_root / relative
            regen_dir(destination)
            for generated_file in sorted((Path(tmp) / src_name).iterdir()):
                shutil.copy2(generated_file, destination / generated_file.name)
            (destination / "README.md").write_text(
                GENERATED_README, encoding="utf-8"
            )

        # Claude Code keeps the source's role subfolders, so this one copies as
        # a tree. It carries NO README.md: /deploy-global copies this directory
        # wholesale into ~/.claude/agents/, where a file without agent
        # frontmatter is not inert — the marker lives one level up instead.
        regen_dir(output_root / Path("harness/claude"))
        claude_agents = output_root / Path("harness/claude/agents")
        claude_agents.mkdir(parents=True)
        for generated_file in sorted((Path(tmp) / "claude").rglob("*.md")):
            relative_path = generated_file.relative_to(Path(tmp) / "claude")
            (claude_agents / relative_path).parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(generated_file, claude_agents / relative_path)
        (output_root / "harness" / "claude" / "README.md").write_text(
            GENERATED_README, encoding="utf-8"
        )

    # Router skills: inject canonical rule files as frontmatter-stripped
    # references so each skill body's routing table resolves.
    for skill_name, sources in SKILL_REFERENCE_INJECTIONS.items():
        refs = skills_out / skill_name / "references"
        if not (skills_out / skill_name).exists():
            sys.exit(f"ERROR: SKILL_REFERENCE_INJECTIONS names '{skill_name}' but "
                     f"global/skills/{skill_name}/ does not exist.")
        refs.mkdir(parents=True, exist_ok=True)
        injected = 0
        for subdir, pattern in sources:
            matches = sorted((source_root / "global" / subdir).glob(pattern))
            if not matches:
                sys.exit(f"ERROR: no files match global/{subdir}/{pattern} "
                         f"(reference injection for '{skill_name}').")
            for rule in matches:
                text = rule.read_text(encoding="utf-8")
                match = re.match(r"^---\n.*?\n---\n?", text, re.S)
                (refs / rule.name).write_text(
                    text[match.end():] if match else text, encoding="utf-8"
                )
                injected += 1
        print(f"injected {injected} references -> {refs}")

    # Last: the manifest cites the reference paths the injection above created,
    # so it is derived from a state that already exists on disk.
    write_rule_manifest(output_root, source_root)


def _tree_entries(tree: Path):
    """Return relative node names mapped to type and byte content."""
    entries = {}
    if not tree.is_dir():
        return entries
    for path in sorted(tree.rglob("*")):
        relative = path.relative_to(tree).as_posix()
        if path.is_symlink():
            entries[relative] = ("symlink", str(path.readlink()).encode())
        elif path.is_dir():
            entries[relative] = ("directory", b"")
        elif path.is_file():
            entries[relative] = ("file", path.read_bytes())
        else:
            entries[relative] = ("other", b"")
    return entries


def check_generated_tree_parity(actual_root: Path = ROOT):
    """Regenerate to a temporary root and compare every derived tree read-only."""
    issues = []
    with tempfile.TemporaryDirectory(prefix="hive-generated-check-") as tmp:
        expected_root = Path(tmp)
        generate_generated_trees(expected_root)
        for relative in GENERATED_TREE_RELATIVE_PATHS:
            expected_tree = expected_root / relative
            actual_tree = actual_root / relative
            if not actual_tree.exists() and not actual_tree.is_symlink():
                issues.append(("missing", relative.as_posix()))
                continue
            if not actual_tree.is_dir() or actual_tree.is_symlink():
                issues.append(("stale", relative.as_posix()))
                continue

            expected = _tree_entries(expected_tree)
            actual = _tree_entries(actual_tree)
            for node in sorted(expected.keys() - actual.keys()):
                issues.append(("missing", (relative / node).as_posix()))
            for node in sorted(actual.keys() - expected.keys()):
                issues.append(("extra", (relative / node).as_posix()))
            for node in sorted(expected.keys() & actual.keys()):
                if expected[node] != actual[node]:
                    issues.append(("stale", (relative / node).as_posix()))

        for relative in GENERATED_FILE_RELATIVE_PATHS:
            expected_file = expected_root / relative
            actual_file = actual_root / relative
            if not actual_file.is_file():
                issues.append(("missing", relative.as_posix()))
            elif actual_file.read_bytes() != expected_file.read_bytes():
                issues.append(("stale", relative.as_posix()))

    if issues:
        details = "\n".join(f"  {kind}: {path}" for kind, path in issues)
        sys.exit(
            "ERROR: generated-tree parity: derived harness outputs do not match "
            "their canonical global/ sources:\n"
            f"{details}\n"
            "Run python3 harness/build.py to regenerate them; never hand-edit "
            "generated trees."
        )
    print(
        f"generated trees: {len(GENERATED_TREE_RELATIVE_PATHS)} trees and "
        f"{len(GENERATED_FILE_RELATIVE_PATHS)} standalone files match "
        "their canonical global/ sources."
    )


def report_agents_size():
    """Report the always-on core's size and per-session cost. Never blocks.

    No harness caps this file (see the header note), so there is nothing here to
    fail the build on. The number is published because the cost is real — it is
    paid at the start of every session in every project — and because a rising
    trend is the cue to run the placement audit, not to reword paragraphs.
    """
    agents = ROOT / "harness" / "AGENTS.md"
    size = agents.stat().st_size
    print(f"harness/AGENTS.md: {size / 1024:.1f} KiB (~{size // 4:,} tokens "
          f"per session, every project). No harness caps it; growth is governed "
          f"by placement, not by a byte budget.")


def report_codex_chain():
    """Report this repo's own Codex instruction chain — advisory, never fatal.

    Codex concatenates <git-root>/AGENTS.md with every intermediate AGENTS.md
    down to the CWD, so the chain size depends on where the session is opened:
    at the repo root it is the root file alone; under harness/ the shared core
    loads a second time (once as the deployed global, once as a project doc).

    Report BOTH, always. A single number labelled "the chain" is read as the
    current state, and the worst case then gets quoted as if it were today's
    reading — which is how a 94%-of-cap figure for a directory nobody opens
    ends up in a status report about the repo at large.

    `project_doc_max_bytes` budgets ONLY this project chain. The global
    ~/.codex/AGENTS.md is exempt (verified in codex source + reproduced
    empirically; bibliography, 2026-08-18) — never add it to these numbers.
    """
    root_agents = ROOT / "AGENTS.md"
    if not root_agents.exists():
        return
    at_root = root_agents.stat().st_size
    under_harness = at_root + (ROOT / "harness" / "AGENTS.md").stat().st_size
    cap, origin = _codex_chain_cap()

    def fmt(size):
        return f"{size / 1024:.1f} KiB, {size / cap * 100:.0f}% of cap"

    header = (f"codex project chain — cap {cap // 1024} KiB ({origin}); "
              f"the global ~/.codex/AGENTS.md is exempt and never counts here")
    if at_root > cap:
        print(f"NOTE: {header}\n"
              f"  EVERY session truncates silently: at repo root {fmt(at_root)}. "
              f"Move content behind a router skill, or raise "
              f"project_doc_max_bytes in ~/.codex/config.toml.")
    elif under_harness > cap:
        print(f"NOTE: {header}\n"
              f"  at repo root: {fmt(at_root)} — the normal case\n"
              f"  under harness/: {fmt(under_harness)} — a session opened THERE "
              f"truncates silently. Open Codex at the root, or move content "
              f"behind a router skill.")
    else:
        print(f"{header}:\n"
              f"  at repo root: {fmt(at_root)} — the normal case\n"
              f"  under harness/: {fmt(under_harness)} — worst case; "
              f"open Codex at the root")


def _codex_chain_cap():
    """The cap this machine actually enforces, not the vendor default.

    A warning that fires against a number nobody uses is a warning nobody reads:
    `project_doc_max_bytes` is routinely raised, and comparing against the 32 KiB
    default then reports a truncation risk that does not exist here. Falls back to
    the default when the key is absent — which is also what a fresh clone gets.
    """
    cfg = Path.home() / ".codex" / "config.toml"
    if cfg.exists():
        for raw in cfg.read_text(encoding="utf-8", errors="replace").splitlines():
            line = raw.split("#", 1)[0].strip()
            if line.startswith("project_doc_max_bytes"):
                digits = "".join(c for c in line.split("=", 1)[-1] if c.isdigit())
                if digits:
                    return int(digits), "configured"
    return CHAIN_CAP_BYTES, "default"


# The delegation thresholds are stated twice on purpose: the canonical rule
# carries them with their rationale, the always-on core restates them condensed.
# Nothing tied the two copies together, so a number could change in one and
# silently contradict the other in every Codex/opencode/Grok session. This check
# is that tie — the canonical file stays the source, and the core must not
# disagree with it. Deterministic: the build exits non-zero on drift.
DELEGATION_CANON = ROOT / "global" / "rules-situational" / "agent-routing.md"
DELEGATION_CANON_START = "### Delegation Gates"
DELEGATION_CANON_END = "Inline vs delegate"
DELEGATION_CORE_ANCHOR = "- **Delegation gates are hard.**"

# Only STRUCTURED thresholds ("4+", "~20") are compared. The canonical section
# also carries bare sub-thresholds the core deliberately folds away (5
# exploratory reads, 2 non-mechanical edits inside the ~20 call gate); they are
# canon-only by design, so including them would fail a file that is correct.
THRESHOLD_TOKEN = re.compile(r"~\d+|\d+\+")


def check_delegation_threshold_parity():
    """Fail the build when the core's delegation numbers drift from the canon.

    Fails closed on the anchors too: a renamed section or bullet means the check
    can no longer see what it claims to verify, which is not a pass.
    """
    canon_text = DELEGATION_CANON.read_text(encoding="utf-8")
    start = canon_text.find(DELEGATION_CANON_START)
    end = canon_text.find(DELEGATION_CANON_END, start + 1) if start != -1 else -1
    if start == -1 or end == -1:
        sys.exit(f"ERROR: delegation-threshold parity: anchor "
                 f"{DELEGATION_CANON_START!r}..{DELEGATION_CANON_END!r} not found in "
                 f"{DELEGATION_CANON.relative_to(ROOT)}. Restore it or update "
                 f"DELEGATION_CANON_START/END in harness/build.py.")
    canon_tokens = set(THRESHOLD_TOKEN.findall(canon_text[start:end]))

    core = ROOT / "harness" / "AGENTS.md"
    core_bullet = next((line for line in core.read_text(encoding="utf-8").splitlines()
                        if line.startswith(DELEGATION_CORE_ANCHOR)), None)
    if core_bullet is None:
        sys.exit(f"ERROR: delegation-threshold parity: anchor "
                 f"{DELEGATION_CORE_ANCHOR!r} not found in harness/AGENTS.md. "
                 f"Restore it or update DELEGATION_CORE_ANCHOR in harness/build.py.")
    core_tokens = set(THRESHOLD_TOKEN.findall(core_bullet))

    if canon_tokens != core_tokens:
        sys.exit(
            "ERROR: delegation-threshold parity: the always-on core contradicts the "
            "canonical rule.\n"
            f"  canon  global/rules-situational/agent-routing.md > {DELEGATION_CANON_START}: "
            f"{sorted(canon_tokens)}\n"
            f"  core   harness/AGENTS.md > {DELEGATION_CORE_ANCHOR}: {sorted(core_tokens)}\n"
            f"  only in canon: {sorted(canon_tokens - core_tokens) or 'none'}\n"
            f"  only in core:  {sorted(core_tokens - canon_tokens) or 'none'}\n"
            "Change the canonical rule first, then mirror it into the core section "
            "(global/core-sections/work-style-delegation.md) and rebuild.")

    print(f"delegation thresholds: canon and core agree ({', '.join(sorted(canon_tokens))}).")


# The router table in global/CLAUDE.md names, for each situational router, the
# observable act that fires it. Claude Code and Grok read that file; Codex and
# opencode read the condensed core instead. A router added to one and forgotten
# in the other is invisible in every session of the harnesses that missed it —
# the exact failure mode this table exists to prevent. The check is directional:
# every router the table names must also appear in the core, and must exist on
# disk. Deterministic: the build exits non-zero on drift.
ROUTER_TABLE_START = "**The routers and the act that fires each one.**"
ROUTER_TABLE_END = "**A trigger is written as an act"
ROUTER_ROW = re.compile(r"^\|\s*`([a-z0-9-]+)`\s*\|")


def check_router_index_parity():
    """Fail the build when a router is listed for one harness family but not the other."""
    claude = ROOT / "global" / "CLAUDE.md"
    text = claude.read_text(encoding="utf-8")
    start = text.find(ROUTER_TABLE_START)
    end = text.find(ROUTER_TABLE_END, start + 1) if start != -1 else -1
    if start == -1 or end == -1:
        sys.exit(f"ERROR: router-index parity: anchor {ROUTER_TABLE_START!r}.."
                 f"{ROUTER_TABLE_END!r} not found in global/CLAUDE.md. Restore it "
                 f"(source: global/core-sections/skill-routing.md) or update "
                 f"ROUTER_TABLE_START/END in harness/build.py.")
    routers = [m.group(1) for line in text[start:end].splitlines()
               for m in [ROUTER_ROW.match(line)] if m]
    if not routers:
        sys.exit("ERROR: router-index parity: the router table in global/CLAUDE.md "
                 "matched zero rows. A table that names nothing is not a pass.")

    missing_on_disk = [r for r in routers if not (ROOT / "global" / "skills" / r).is_dir()]
    if missing_on_disk:
        sys.exit("ERROR: router-index parity: the table names skills that do not exist "
                 f"under global/skills/: {sorted(missing_on_disk)}. Fix the name or "
                 "remove the row.")

    core_text = (ROOT / "harness" / "AGENTS.md").read_text(encoding="utf-8")
    missing_in_core = [r for r in routers if f"`{r}`" not in core_text]
    if missing_in_core:
        sys.exit(
            "ERROR: router-index parity: routers listed for Claude Code/Grok never reach "
            "Codex/opencode.\n"
            f"  in global/CLAUDE.md table: {sorted(routers)}\n"
            f"  absent from harness/AGENTS.md: {sorted(missing_in_core)}\n"
            "Add each missing router to the always-on core (source: "
            "global/core-sections/on-demand-rules.md), stating the act that fires it.")

    print(f"router index: {len(routers)} routers, table and core agree "
          f"({', '.join(routers)}).")


def main():
    # --check: run every parity check read-only (deploy preflight, pre-commit)
    # — nothing is written, so a stale or hand-edited core FAILS here instead
    # of being regenerated or refused mid-write.
    if "--check" in sys.argv[1:]:
        check_delegation_threshold_parity()
        check_router_index_parity()
        check_core_assembly_parity()
        check_generated_tree_parity()
        check_core_size_ratchet(allow_write=False)
        print("check-only run: nothing written.")
        return

    # --bless-core-size: deliberately update the ratchet baseline — only from
    # a verified regenerated state, never from a stale or hand-edited core.
    if "--bless-core-size" in sys.argv[1:]:
        check_core_assembly_parity()
        counts = _core_token_counts()
        _write_core_size_baseline(counts)
        for path, tokens in sorted(counts.items()):
            print(f"blessed {path}: ~{tokens:,} tokens")
        return

    # Always-on cores first, so the size reports below measure generated output
    assemble_cores()
    generate_generated_trees(ROOT)

    report_agents_size()
    report_codex_chain()

    print("harness/ regenerated. If `git status` shows changes, commit them — "
          "a dirty tree after build means canonical sources changed without a rebuild.")

    # Last: the trees are already regenerated and the commit reminder already
    # printed, so a drift exit signals only the drift.
    check_delegation_threshold_parity()
    check_router_index_parity()
    check_core_assembly_parity()
    check_generated_tree_parity()
    check_core_size_ratchet(allow_write=True)


if __name__ == "__main__":
    main()
