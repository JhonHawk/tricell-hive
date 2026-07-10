#!/usr/bin/env python3
"""Regenerate every generated tree under harness/ from the canonical sources.

Single entry point (run from the repo root):
    python3 harness/build.py

Generates (delete-and-recreate, never incremental):
    harness/agents-skills/        <- global/skills   (cleaned universal skills)
    harness/codex/agents/         <- global/agents   (Codex TOML subagents)
    harness/opencode/agents/      <- global/agents   (opencode markdown subagents)
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

# AGENTS.md is the condensed, portable rule set shared by Codex and opencode.
# Two thresholds, two purposes (see CLAUDE.md > Routing Maintenance):
#   BUDGET  — editorial limit: it is a *condensed* file; at this size, prune
#             before adding. Protects conciseness (context cost in BOTH harnesses).
#   HARD    — Codex's project_doc_max_bytes airbag: past this, Codex truncates
#             silently (dropping the LAST sections). Must match the snippet value.
AGENTS_BUDGET_BYTES = 30 * 1024
AGENTS_HARD_LIMIT_BYTES = 48 * 1024

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
            f"{AGENTS_HARD_LIMIT_BYTES // 1024} KiB hard limit. Codex truncates "
            f"silently past project_doc_max_bytes (the last sections — Git, "
            f"Session Execution Mode — drop first). Prune before deploying."
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
        ):
            regen_dir(dst)
            for f in sorted((Path(tmp) / src_name).iterdir()):
                shutil.copy2(f, dst / f.name)
            (dst / "README.md").write_text(GENERATED_README, encoding="utf-8")

    # Path-scoped language rules -> opencode-rules plugin format
    rules_out = ROOT / "harness" / "opencode" / "rules"
    regen_dir(rules_out)
    subprocess.run(
        [sys.executable, str(BUILD / "convert-rules.py"),
         str(ROOT / "global" / "rules" / "languages"), str(rules_out)],
        check=True,
    )
    (rules_out / "README.md").write_text(GENERATED_README, encoding="utf-8")

    # language-rules router skill (Codex leg): inject the canonical rule files as
    # frontmatter-stripped references so the skill body's routing table resolves.
    lang_refs = skills_out / "language-rules" / "references"
    lang_refs.mkdir(parents=True, exist_ok=True)
    injected = 0
    for rule in sorted((ROOT / "global" / "rules" / "languages").glob("*.md")):
        text = rule.read_text(encoding="utf-8")
        m = re.match(r"^---\n.*?\n---\n?", text, re.S)
        (lang_refs / rule.name).write_text(text[m.end():] if m else text, encoding="utf-8")
        injected += 1
    print(f"injected {injected} language-rule references -> {lang_refs}")

    check_agents_size()

    print("harness/ regenerated. If `git status` shows changes, commit them — "
          "a dirty tree after build means canonical sources changed without a rebuild.")


if __name__ == "__main__":
    main()
