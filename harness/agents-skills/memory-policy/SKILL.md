---
name: memory-policy
description: >
  Load BEFORE recording or recalling anything that must outlive this session — saving a
  decision, a root cause or a convention, recalling prior work, or the close-time summary.
  Covers project identity, save cadence, topic_key upserts, invalidation and tracker sync.
  The user asks to "remember", "save this", "what did we decide"; they never name the tool.
  Triggers: remember, save, recall, "what did we decide", session close;
  recuerda, guarda esto, anota, "qué decidimos".
---

# memory-policy — the policy layer over the Engram plugin protocol

The Engram plugin hook-injects the PROTOCOL (tools, save/search triggers, session
summary) every session; this skill carries the POLICY that keeps those saves clean:
project identity, save cadence, upserts, invalidation, tracker sync. No harness loads it
always-on — it is dead weight in a session that never touches memory.

**Where the reference lives** — same canonical file, two paths:

| Harness | Read from |
|---|---|
| Codex / opencode | `references/memory-routing.md` (injected at build time) |

## Rules of use

- Read it in full on first load. Its native file-memory half (`MEMORY.md`, memory
  directories) is Claude-Code-only — Codex/opencode sessions skip that half rather than
  simulating it through other storage.
- **Not in this skill and never conditional:** how to ANSWER a question about state
  ("what's pending / where are we?") is always-on in `quality/debugging.md > Reporting
  state from ground truth` — verify against git/disk/DB before reporting, whether or not
  this skill ever loads.
- This skill parameterizes Engram's own affordances (`topic_key`, `mem_update`,
  `.engram/config.json`) — it never overrides the plugin-injected protocol. On conflict,
  the plugin's protocol wins and this skill gets fixed, not argued past.
- Session without Engram tools → this skill has nothing for the task; never simulate
  memory behavior through unrelated storage.
