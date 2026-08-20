#!/usr/bin/env python3
"""Regenerate every generated tree under harness/ from the canonical sources.

Single entry point (run from the repo root):
    python3 harness/build.py

Generates (delete-and-recreate, never incremental):
    harness/agents-skills/        <- global/skills   (cleaned universal skills)
    harness/codex/agents/         <- global/agents   (Codex TOML subagents)
    harness/opencode/agents/      <- global/agents   (opencode markdown subagents)
    harness/grok/agents/          <- global/agents   (Grok Build markdown agents)
    harness/opencode/rules/       <- global/rules/languages (opencode-rules plugin format)

Hand-written sources are never touched: harness/AGENTS.md, harness/codex/{README,
*.snippet}, harness/opencode/{README, *.snippet, commands/}.

Run this after ANY edit to global/agents or global/skills, before committing —
/deploy-global also runs it and flags a dirty harness/ as a missed rebuild.
"""
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
BUILD = ROOT / "harness" / "build"

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
        ("rules/languages", "*.md"),
        ("rules/quality", "development-principles.md"),
        # Extracted out of development-principles.md + debugging.md, both injected
        # here: without this line Codex and opencode silently lose both sections.
        ("rules/quality", "reporting-integrity.md"),
        ("rules/quality", "testing.md"),
        ("rules/quality", "debugging.md"),
        ("rules/quality", "patterns-antipatterns.md"),
        ("rules/workflow", "devops-principles.md"),
        ("rules/tools", "browser-automation.md"),
        ("rules/tools", "code-search.md"),
        # Claude Code and Grok get this always-on, but on Codex and opencode the
        # agent Role rules table is the ONLY pointer to the Context7 protocol —
        # and it pointed at a file this map never generated.
        ("rules/tools", "context7.md"),
    ],
    "workspace-conventions": [
        ("rules/workflow", "project-structure.md"),
        ("rules/workflow", "session-capture.md"),
        ("rules/workflow", "support-artifacts.md"),
        ("rules/workflow", "cross-service-workflow.md"),
        ("rules/workflow", "infra-naming.md"),
    ],
    # Sources live in rules-situational/: their trigger is an intent (delegating,
    # planning), which `paths:` cannot express, so a router is their only channel.
    "task-routing": [
        ("rules-situational", "agent-routing.md"),
        ("rules-situational", "gap-resolution.md"),
        # The core gates behavior on "trivial" (review scaling, test ritual) but
        # only this file defines it — and no router carried it, so Codex/opencode
        # judged the threshold with nothing to judge it by.
        ("rules/quality", "critical-thinking.md"),
    ],
    "git-mechanics": [
        ("rules-situational", "git-mechanics.md"),
    ],
    "memory-policy": [
        ("rules-situational", "memory-routing.md"),
    ],
    "flow-report": [
        ("rules/quality", "communication-format.md"),
    ],
    "unattended-delegation": [
        ("rules/workflow", "unattended-autonomy.md"),
    ],
}

GENERATED_README = """# Generated — do not edit

Everything in this directory is derived from canonical sources in `global/`
by `harness/build.py`. Edit the canonical file and rebuild; hand edits here
are overwritten on the next build.
"""


def regen_dir(path: Path):
    if path.exists():
        shutil.rmtree(path)
    path.mkdir(parents=True)


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


def main():
    # Universal skills (Codex + opencode read ~/.agents/skills)
    skills_out = ROOT / "harness" / "agents-skills"
    regen_dir(skills_out)
    subprocess.run(
        [sys.executable, str(BUILD / "convert-skills.py"),
         str(ROOT / "global" / "skills"), str(skills_out)],
        check=True,
    )
    (skills_out / "README.md").write_text(GENERATED_README, encoding="utf-8")

    # Per-harness agents
    with tempfile.TemporaryDirectory() as tmp:
        subprocess.run(
            [sys.executable, str(BUILD / "convert-agents.py"),
             str(ROOT / "global" / "agents"), tmp],
            check=True,
        )
        for src_name, dst in (
            ("codex", ROOT / "harness" / "codex" / "agents"),
            ("opencode", ROOT / "harness" / "opencode" / "agents"),
            ("grok", ROOT / "harness" / "grok" / "agents"),
        ):
            regen_dir(dst)
            for f in sorted((Path(tmp) / src_name).iterdir()):
                shutil.copy2(f, dst / f.name)
            (dst / "README.md").write_text(GENERATED_README, encoding="utf-8")

    # Path-scoped rules -> opencode-rules plugin format (the converter skips
    # alwaysApply files, so the workflow pass only picks up path-scoped ones)
    rules_out = ROOT / "harness" / "opencode" / "rules"
    regen_dir(rules_out)
    for rules_src in ("languages", "workflow"):
        subprocess.run(
            [sys.executable, str(BUILD / "convert-rules.py"),
             str(ROOT / "global" / "rules" / rules_src), str(rules_out)],
            check=True,
        )
    (rules_out / "README.md").write_text(GENERATED_README, encoding="utf-8")

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
            matches = sorted((ROOT / "global" / subdir).glob(pattern))
            if not matches:
                sys.exit(f"ERROR: no files match global/{subdir}/{pattern} "
                         f"(reference injection for '{skill_name}').")
            for rule in matches:
                text = rule.read_text(encoding="utf-8")
                m = re.match(r"^---\n.*?\n---\n?", text, re.S)
                (refs / rule.name).write_text(
                    text[m.end():] if m else text, encoding="utf-8")
                injected += 1
        print(f"injected {injected} references -> {refs}")

    report_agents_size()
    report_codex_chain()

    print("harness/ regenerated. If `git status` shows changes, commit them — "
          "a dirty tree after build means canonical sources changed without a rebuild.")


if __name__ == "__main__":
    main()
