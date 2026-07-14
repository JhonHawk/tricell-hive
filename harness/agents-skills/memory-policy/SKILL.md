---
name: memory-policy
description: >
  Load before the FIRST Engram memory operation of a session (mem_save, mem_search,
  mem_context, mem_session_summary) in any multi-repo workspace under
  projects/<group>/<project>/, and before answering "what's pending / where are we" from
  memory. Covers unified project naming via .engram/config.json, deterministic topic_key
  upserts for status facts, session-slug tagging, rewriting stale memories instead of
  appending, ground-truth-before-memory verification by claim type, and declared-tracker
  consultation. The Engram plugin's injected protocol always wins; this is only the
  policy layer on top. Codex/opencode channel; Claude Code receives these rules
  always-on via its own rules — do not load there.
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
