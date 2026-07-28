---
name: memory-policy
description: >
  Codex/opencode: load before first Engram op (mem_save/search/context/summary) or
  answering pending state from memory. Project identity, topic_key, ground-truth.
  Skip Claude Code (always-on). Plugin protocol wins over this policy layer.
---

# memory-policy — the policy layer over the Engram plugin protocol

The full canonical policy lives in `references/memory-routing.md` (injected at build
time from `global/rules/workflow/` — single source of truth). The Engram plugin
hook-injects the PROTOCOL (tools, save/search triggers, session summary) every session;
this skill carries the POLICY that keeps those saves clean: project identity, upserts,
invalidation, ground truth.

## Rules of use

- Read `references/memory-routing.md` in full on first load; apply the Engram side. Its
  native file-memory half (`MEMORY.md`, memory directories) is Claude-Code-only —
  Codex/opencode sessions skip it rather than simulating it through other storage.
- This skill parameterizes Engram's own affordances (`topic_key`, `mem_update`,
  `.engram/config.json`) — it never overrides the plugin-injected protocol. On conflict,
  the plugin's protocol wins and this skill gets fixed, not argued past.
- Session without Engram tools → this skill has nothing for the task; never simulate
  memory behavior through unrelated storage.
