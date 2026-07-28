# Skill catalog budget (Codex / OpenCode / Claude)

**Date:** 2026-07-24  
**Status:** applied after Grok↔Claude consensus (working tree + `~/.agents/skills`; **not** live `~/.claude/` until user `/deploy-global`)  
**Commit:** not committed — review and commit when ready  
**Related:** Codex [Build skills](https://learn.chatgpt.com/docs/build-skills), [Agent Skills spec](https://agentskills.io/specification), OpenCode [Skills](https://opencode.ai/docs/skills/)

## Why

Codex loads a **skills catalog** (name + description + path) into context before any body:

- Cap: **at most ~2% of the model context window**, or **8 000 characters** when the window is unknown.
- Under pressure it **shortens descriptions** and may **omit** skills from the initial list (with a warning).
- Implicit match uses that catalog; a bloated catalog → worse routing + missing skills.

An audit of `~/.agents/skills` (2026-07-24) found:

| Pool | Est. listing size | vs 8k floor |
|------|-------------------|-------------|
| All skills (~39) | ~19 300 chars | **~241%** |
| Implicit-only (~30) | ~14 800 chars | **~185%** |
| One description | `language-rules` **1087** chars | **>1024** (Agent Skills / OpenCode hard max) |

Several design/image skills also had **implicit on** and multi-thousand-line bodies — a wrong auto-match would dump ~5–20k tokens into the session.

## Consensus (Grok ↔ Claude, same day)

Debate in Herdr (`w9:p1` Grok / `w9:p3` Claude). Package **A–F accepted** with Claude’s precision amendments.

### Critical distinction (do not re-break)

| Action | Meaning |
|--------|---------|
| **Load the skill** | Model reads `SKILL.md` instructions — required for unattended protocol |
| **Activate the mode** | User has explicitly handed over control — guarded by body + description, not by blocking model invocation |

**Wrong gate (reverted):** `disable-model-invocation: true` / `allow_implicit_invocation: false` on `unattended-delegation`. That blocked Codex from loading the skill when the user says “control total…”, so the mode could be declared **without** decision-log / dedicated branch / gates / expiry — the highest-risk path without policy.

**Correct guard:** short description + body (“Never activates from silence…”, declare scope/gates/log first). `harness/AGENTS.md` still says: on explicit handover, **invoke** `unattended-delegation` before declaring the mode accepted.

### Deploy boundary

- **Never** copy into `~/.claude/` without user `/deploy-global` (or explicit confirm).
- An intermediate selective sync to `~/.claude/skills/` was **reverted** from `git show HEAD:global/skills/<name>/SKILL.md`.
- **Assumption:** last `/deploy-global` matched HEAD; if the last deploy was older, restored files may be newer than the true prior live state — best available approximation; next full deploy converges.
- Live **Claude Code** stays on HEAD copies until the user deploys.
- Live **Codex/opencode** path `~/.agents/skills` **is** updated (session-authorized for this work).

## What remains applied

### 1. Hive routers — shorter descriptions (implicit where designed)

Source: `global/skills/*/SKILL.md` → `python3 harness/build.py` → `harness/agents-skills/` → `~/.agents/skills/`.

| Skill | Role | Notes |
|-------|------|--------|
| `language-rules` | Implicit (Codex router) | Desc short; wording: do not load on Claude Code/opencode own channels; opencode quality refs only |
| `workspace-conventions` | Implicit | Short desc |
| `memory-policy` | Implicit | Short desc |
| `flow-report` | Implicit | Short desc; body declines non-report uses |
| `memory-sync` | Implicit | Short desc |
| `starlight-docs-site` | Implicit | Short desc |
| `engram-init-workspace` | Implicit | Short desc |
| `unattended-delegation` | **Implicit** (model may load) | Short desc **without** dmi; activation still user-explicit only |

### 2. Non-hive skills — explicit-only (`allow_implicit_invocation: false`)

Local `~/.agents/skills/<name>/agents/openai.yaml` only (not in hive `global/`):

`brandkit`, `design-taste-frontend`, `image-to-code`, `imagegen-frontend-mobile`, `imagegen-frontend-web`, `high-end-visual-design`, `redesign-existing-projects`, `industrial-brutalist-ui`, `minimalist-ui`, `gpt-taste`, `stitch-design-taste`, `shadcn`, `firecrawl`, `herdr`, `full-output-enforcement`, `notion-cli`, **`herdr-pre-release-audit`** (17 total).

Reinstall of those packages can wipe `openai.yaml` — re-apply manually or add a re-apply script later (deferred).

### 3. Hook agent-line-count

- `.claude/hooks/agent-line-count.sh` (new)
- `.claude/settings.json`:  
  `bash "${CLAUDE_PROJECT_DIR}/.claude/hooks/agent-line-count.sh"`  
- Commit **settings + script together** when committing.

### 4. Flow pack / gates

Unchanged: already explicit-only where designed.

## Post-consensus checklist (done)

- [x] A — Revert dmi on unattended; consensus description; no blocking `openai.yaml` in harness or `~/.agents`
- [x] B — Restore `~/.claude/skills/{8 hive skills}` from HEAD
- [x] C — Keep short descriptions + non-hive explicit + hook
- [x] D — herdr-pre-release-audit explicit; language-rules wording; `CLAUDE_PROJECT_DIR` in hook
- [x] E — This doc updated
- [x] F — No commit

## Files (hive working tree)

- `global/skills/language-rules/SKILL.md`
- `global/skills/workspace-conventions/SKILL.md`
- `global/skills/memory-policy/SKILL.md`
- `global/skills/unattended-delegation/SKILL.md` (**no** `disable-model-invocation`)
- `global/skills/flow-report/SKILL.md`
- `global/skills/memory-sync/SKILL.md`
- `global/skills/starlight-docs-site/SKILL.md`
- `global/skills/engram-init-workspace/SKILL.md`
- `.claude/hooks/agent-line-count.sh`
- `.claude/settings.json`
- `harness/agents-skills/**` (regenerated)
- `_support/docs/skill-catalog-budget.md` (this file)

## Out of scope / deferred

- Re-apply script for non-hive `openai.yaml` list
- OpenCode `permission.skill` deny lists
- Always-on `harness/AGENTS.md` size budget
- User-run `/deploy-global` (required for live Claude to pick up hive skill edits)
