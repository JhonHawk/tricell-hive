#!/usr/bin/env python3
"""Derive the universal (Codex/opencode-clean) copy of the skills.

Source of truth: global/skills/ (Claude Code format).
Output: a mirror tree where each SKILL.md is cleaned for universal consumers:
  - frontmatter keeps ONLY the universally honored keys
    (name, description, license, compatibility, metadata)
  - Claude-only keys are stripped (argument-hint, disable-model-invocation,
    user-invocable, allowed-tools, context, agent, model, effort, paths, shell)
  - body rewrites: ${CLAUDE_SKILL_DIR} -> ~/.agents/skills/<name>
                   ~/.claude/skills/   -> ~/.agents/skills/
  - every other file (references/, scripts/, assets/, agents/openai.yaml)
    copies verbatim

Run by /deploy-global before copying to ~/.agents/skills/. Never edit the
generated tree by hand — edit the canonical skill and redeploy.

Usage: convert-skills.py <skills-src-dir> <out-dir>
"""
import re
import shutil
import sys
from pathlib import Path

KEEP_KEYS = {"name", "description", "license", "compatibility", "metadata"}


def clean_frontmatter(fm_text: str) -> str:
    """Keep only KEEP_KEYS, preserving their indented continuation lines."""
    out, keeping = [], False
    for line in fm_text.splitlines():
        key_match = re.match(r"^([A-Za-z][\w-]*):", line)
        if key_match:
            keeping = key_match.group(1) in KEEP_KEYS
        elif not line.startswith((" ", "\t")) and line.strip():
            keeping = False  # non-key, non-continuation line (comments at col 0)
        if keeping:
            out.append(line)
    return "\n".join(out)


def clean_skill_md(text: str, skill_name: str) -> str:
    m = re.match(r"^---\n(.*?)\n---\n?(.*)$", text, re.S)
    if not m:
        return text
    fm, body = m.groups()
    fm = clean_frontmatter(fm)
    body = body.replace("${CLAUDE_SKILL_DIR}", f"~/.agents/skills/{skill_name}")
    body = body.replace("~/.claude/skills/", "~/.agents/skills/")
    return f"---\n{fm}\n---\n{body}"


def main():
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    src, out = Path(sys.argv[1]), Path(sys.argv[2])
    if out.exists():
        shutil.rmtree(out)
    count = 0
    for skill_dir in sorted(p for p in src.iterdir() if p.is_dir()):
        dst = out / skill_dir.name
        shutil.copytree(skill_dir, dst)
        skill_md = dst / "SKILL.md"
        if skill_md.exists():
            skill_md.write_text(
                clean_skill_md(skill_md.read_text(encoding="utf-8"), skill_dir.name),
                encoding="utf-8",
            )
            count += 1
    print(f"converted {count} skills -> {out}")


if __name__ == "__main__":
    main()
