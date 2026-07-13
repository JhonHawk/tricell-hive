# Git Workflow Dual Mode + AskUserQuestion Enforcement

## Problem

The global `git-workflow.md` rules assume a single context: the agent working directly on the user's workspace. When optional reference project (or any orchestrated workflow) operates on an isolated branch/worktree, two rules conflict with the autonomous execution model:

1. "Only commit after user confirmation" — subagents commit autonomously as safety checkpoints
2. "Never ask for git actions proactively" — worktree/branch creation is autonomous

The user confirmed the autonomous behavior works well in practice. The issue is rule inconsistency, not behavioral problems.

## Decision

Adopt **Dual Mode** in `git-workflow.md`: two explicit contexts with distinct rules, unified by shared safety constraints.

## Design

### Mode Definitions

**Direct Mode (default)**
- Active when the agent works on the user's main workspace (not in a worktree, not on a dedicated feature branch)
- All current rules apply unchanged:
  - Commits require user confirmation
  - Branch creation requires user confirmation
  - No proactive git actions

**Isolated Mode**
- Activated automatically when **either** condition is true:
  1. Working inside a git worktree (detectable via `git worktree list`)
  2. Working on a dedicated feature branch created for the current task (not `main`/`master`/`develop`)
- Rules:
  - Commits are autonomous (function as safety checkpoints)
  - Branch creation is autonomous (required for isolation)
  - Conventional commit format still applies
  - Commits remain small and focused (one concern per commit)
  - **Control gate**: merge/PR/keep/discard decision at the end (pattern: `finishing-a-development-branch`)

**Activation criteria are observable**, not subjective — the agent checks git state, not intent.

### Checkpoints (Isolated Mode only)

For plans with **4+ tasks**, intermediate checkpoints are mandatory:

- **Frequency**: every 3 completed tasks
- **Content**:
  - Brief summary of completed tasks
  - Any concerns or deviations from the plan
  - Test status
  - Options: continue / adjust / pause
- **Mechanism**: use `AskUserQuestion` with structured options
- **Not a code review** — spec reviewer and code quality reviewer already handle that. This is a progress pulse.

Plans with 1-3 tasks: final gate only (no intermediate checkpoints).

### Safety Rules (both modes)

These rules apply regardless of mode — they are never relaxed:

- Never force-push without explicit user confirmation
- Never merge to main/master without user confirmation
- Never push to remote without user confirmation
- Never delete branches (local or remote) without confirmation
- Never rewrite published history without confirmation

### Conventions (both modes)

These apply unchanged in both modes:

- Conventional commits required (`feat:`, `fix:`, `refactor:`, etc.)
- Commits must be small and focused (one concern per commit)
- Branch naming: `<type>/<short-description>` in kebab-case
- Never commit directly to `main` or `master`
- PR title follows conventional commit format
- One PR per logical change
- No AI attribution in commits, PRs, or messages

## AskUserQuestion Enforcement

**Separate change** in `global/CLAUDE.md` under the Communication section.

Current rule (vague):
> "Use interactive questions (structured choices) when the decision has discrete options."

New rule (explicit):
> "**Use `AskUserQuestion`** for any question with 2+ discrete options. Never present choices as plain numbered text. Reserve prose for open-ended questions."

**Rationale**: `AskUserQuestion` provides preview support for comparing artifacts, `multiSelect` for non-exclusive choices, and renders as an interactive component — faster to answer and consistent with the optional reference project experience.

## Files Impacted

| File | Change |
|------|--------|
| `global/rules/workflow/git-workflow.md` | Restructure into dual mode + shared conventions/safety + checkpoints section |
| `global/CLAUDE.md` | Update Communication bullet to enforce `AskUserQuestion` for discrete options |

### Files NOT impacted

- optional reference project skills — external plugin, already compatible with isolated mode
- `agent-routing.md` — no routing changes
- Other rules/ files — no changes needed

## Risks

1. **Mode detection ambiguity**: an agent might be "on a feature branch" without it being a dedicated task branch (e.g., the user manually checked out a branch earlier). Mitigation: isolated mode activates when the agent itself created the branch or was explicitly instructed to work on it, not just when it detects any non-main branch.
2. **Checkpoint interruption cost**: pausing every 3 tasks adds latency. Mitigation: the checkpoint is lightweight (summary + question), not a full review cycle.
