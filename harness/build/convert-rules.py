#!/usr/bin/env python3
"""Derive opencode-rules files from the path-scoped Claude Code language rules.

Source of truth: global/rules/languages/ (Claude Code path-scoped format:
`paths:` glob list in the frontmatter, conditional loading by touched file).
Output: a mirror tree consumable by the opencode-rules plugin
(github.com/frap129/opencode-rules, audited & pinned 0.6.4), deployed by
/deploy-global to ~/.config/opencode/rules/:
  - frontmatter maps `paths:` -> `globs:` (same glob syntax) + `match: any`
  - every other frontmatter key is dropped (the plugin ignores unknown keys,
    but leaking Claude-only keys invites misreads)
  - body copies verbatim

Only language rules convert — the alwaysApply rules (quality/workflow/tools)
stay represented by the condensed harness/AGENTS.md; duplicating them here
would double their always-on token cost in opencode for zero conditionality.

Usage: convert-rules.py <languages-src-dir> <out-dir>
"""
from __future__ import annotations

import re
import sys
from pathlib import Path


def convert(text: str) -> str | None:
    m = re.match(r"^---\n(.*?)\n---\n?(.*)$", text, re.S)
    if not m:
        return None  # no frontmatter -> not path-scoped -> skip
    fm, body = m.groups()
    paths = re.findall(r'^\s*-\s*["\']?([^"\'\n]+?)["\']?\s*$', fm, re.M)
    if not paths:
        return None
    globs = "\n".join(f'  - "{p}"' for p in paths)
    return f"---\nglobs:\n{globs}\nmatch: any\n---\n{body}"


def main():
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    src, out = Path(sys.argv[1]), Path(sys.argv[2])
    out.mkdir(parents=True, exist_ok=True)
    count, skipped = 0, []
    for rule in sorted(src.glob("*.md")):
        converted = convert(rule.read_text(encoding="utf-8"))
        if converted is None:
            skipped.append(rule.name)
            continue
        (out / rule.name).write_text(converted, encoding="utf-8")
        count += 1
    print(f"converted {count} rules -> {out}")
    if skipped:
        print(f"skipped (no paths: frontmatter): {', '.join(skipped)}")


if __name__ == "__main__":
    main()
