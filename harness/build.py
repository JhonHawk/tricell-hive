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

# AGENTS.md is the always-on core shared by Codex and opencode (pruned to
# ~18 KiB 2026-07-14; situational depth lives behind the router skills).
# Two thresholds, two purposes:
#   BUDGET  — editorial limit: at this size, prune before adding. Re-bloat is
#             a deliberate act, not drift.
#   HARD    — build failure: the core has no business growing past this.
# `project_doc_max_bytes` covers ONLY the repo chain (git-root → cwd); the
# deployed global file does not count against it. Measured 2026-08-15 with
# `codex debug prompt-input`: from harness/, loaded docs total 81,278 B against
# a 65,536 cap with no truncation — only the 54,347 B chain is charged.
CHAIN_CAP_BYTES = 32 * 1024  # Codex default; the conservative number for a public repo
AGENTS_BUDGET_BYTES = 28 * 1024
AGENTS_HARD_LIMIT_BYTES = 32 * 1024

# Router skills (Codex/opencode leg): canonical rule files injected as
# frontmatter-stripped references so each skill's routing table resolves.
# Keys are skill dir names under global/skills/; entries are (dir, glob)
# relative to global/. The skill body names the trigger; the reference IS
# the single canonical source — never fork content into the skill.
SKILL_REFERENCE_INJECTIONS = {
    "language-rules": [
        ("rules/languages", "*.md"),
        ("rules/quality", "development-principles.md"),
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
    ],
    "memory-policy": [
        ("rules/workflow", "memory-routing.md"),
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


def check_agents_size():
    """Guard the shared AGENTS.md against context bloat and Codex truncation.

    Warns at the editorial budget; fails the build past the hard limit so a
    Codex-truncating file never reaches deploy. Returns True if within budget.
    """
    agents = ROOT / "harness" / "AGENTS.md"
    size = agents.stat().st_size
    kib = size / 1024
    if size > AGENTS_HARD_LIMIT_BYTES:
        sys.exit(
            f"ERROR: harness/AGENTS.md is {kib:.1f} KiB, over the "
            f"{AGENTS_HARD_LIMIT_BYTES // 1024} KiB hard limit. It is the "
            f"always-on core — move situational content behind a router skill "
            f"(SKILL_REFERENCE_INJECTIONS) instead of growing it."
        )
    if size > AGENTS_BUDGET_BYTES:
        print(
            f"WARNING: harness/AGENTS.md is {kib:.1f} KiB, over the "
            f"{AGENTS_BUDGET_BYTES // 1024} KiB editorial budget. It is a "
            f"condensed file — prune or consolidate before adding more."
        )
        return False
    print(f"harness/AGENTS.md: {kib:.1f} KiB (budget "
          f"{AGENTS_BUDGET_BYTES // 1024} KiB, hard limit "
          f"{AGENTS_HARD_LIMIT_BYTES // 1024} KiB).")
    return True


def report_codex_chain():
    """Report this repo's own Codex instruction chain — advisory, never fatal.

    Codex concatenates <git-root>/AGENTS.md with every intermediate AGENTS.md
    down to the cwd, so a session opened under harness/ loads the shared core
    a second time (once as the deployed global, once as a project doc). The
    file-level budget above cannot see that: the cap applies to the sum.
    """
    root_agents = ROOT / "AGENTS.md"
    if not root_agents.exists():
        return
    chain = root_agents.stat().st_size + (ROOT / "harness" / "AGENTS.md").stat().st_size
    pct = chain / CHAIN_CAP_BYTES * 100
    line = (f"codex chain (AGENTS.md + harness/AGENTS.md): {chain / 1024:.1f} KiB, "
            f"{pct:.0f}% of the {CHAIN_CAP_BYTES // 1024} KiB default cap")
    if chain > CHAIN_CAP_BYTES:
        print(f"NOTE: {line} — a Codex session opened under harness/ truncates "
              f"silently unless project_doc_max_bytes is raised.")
    else:
        print(line + ".")


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

    check_agents_size()
    report_codex_chain()

    print("harness/ regenerated. If `git status` shows changes, commit them — "
          "a dirty tree after build means canonical sources changed without a rebuild.")


if __name__ == "__main__":
    main()
