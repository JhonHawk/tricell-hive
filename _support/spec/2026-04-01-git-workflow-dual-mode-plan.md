# Git Workflow Dual Mode + AskUserQuestion Enforcement — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use optional reference project:subagent-driven-development (recommended) or optional reference project:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restructure git-workflow.md into dual mode (direct/isolated) and enforce `AskUserQuestion` for discrete choices in CLAUDE.md.

**Architecture:** Two file edits — one structural rewrite (git-workflow.md) and one surgical line change (CLAUDE.md). No new files. Config-only repo — no build, no runtime, no tests to run. Verification is manual read-back against the spec.

**Tech Stack:** Markdown configuration files (Claude Code rules system)

---

### Task 1: Rewrite git-workflow.md with dual mode structure

**Files:**
- Modify: `global/rules/workflow/git-workflow.md` (full rewrite, preserving frontmatter)

**Spec reference:** `_support/spec/2026-04-01-git-workflow-dual-mode-design.md` sections: Mode Definitions, Checkpoints, Safety Rules, Conventions.

- [ ] **Step 1: Replace the entire content of git-workflow.md**

Keep the existing frontmatter (`alwaysApply: true`) and replace everything below it with the new dual-mode structure:

```markdown
---
alwaysApply: true
---

## Git Workflow

> Universal git conventions with two operating modes. Project-specific overrides (branching model, protected branches) live in each project's CLAUDE.md.

### Modes of Operation

#### Direct Mode (default)
Active when working on the user's main workspace — not in a worktree, not on a dedicated feature branch that the agent created or was instructed to use.

- **Commit workflow:** after implementing a change, let the user validate first. Only commit after user confirmation.
- **Pre-commit gate:** when the user says "commit", ensure lint, types, and tests pass before executing. If any fail, report and fix first.
- **Next steps after commit:** suggest the next step (PR, push, deploy) but don't execute without confirmation.
- **No proactive git actions;** the user specifies `commit`, `push`, `pull`, etc.

#### Isolated Mode (worktree / feature branch)
Activated automatically when **either** condition is true:
1. Working inside a **git worktree** (detectable via `git worktree list`)
2. Working on a **dedicated feature branch** that the agent created or was explicitly instructed to work on (not `main`/`master`/`develop`)

- **Commits are autonomous** — they function as safety checkpoints during execution. No user confirmation required per commit.
- **Branch creation is autonomous** — required for isolation.
- **Control gate:** the merge/PR/keep/discard decision at the end (pattern: `finishing-a-development-branch`). This is where the user reviews and decides.

##### Checkpoints
For plans with **4+ tasks**, intermediate checkpoints are mandatory:
- **Frequency:** every 3 completed tasks, pause and ask the user via `AskUserQuestion`.
- **Content:** brief summary of completed tasks, any concerns or deviations, test status.
- **Options:** continue / adjust / pause.
- **Not a code review** — spec and quality reviewers already handle that. This is a progress pulse.

Plans with 1-3 tasks: final gate only (no intermediate checkpoints).

### Conventions (both modes)
- **Conventional commits required.** Prefixes: `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `test:`, `perf:`, `ci:`. Scope optional: `feat(auth):`.
- **Commits must be small and focused.** One concern per commit. If a task touches multiple concerns, split into separate commits.
- **Never include attribution, "Generated with Claude Code", "Co-Authored-By", or any AI credit** in commits, PRs, messages, or file outputs.

### Branches
- **Branch naming:** `<type>/<short-description>` in kebab-case. Examples: `feat/user-auth`, `fix/invoice-rounding`, `chore/upgrade-deps`.
- **Never commit directly to `main` or `master`** unless the project explicitly allows it (e.g., config-only repos with no CI).

### Pull Requests
- **PR title follows conventional commit format.** Same prefixes as commits.
- **PR description:** summary of changes (what and why) + test plan or verification steps. No boilerplate filler.
- **One PR per logical change.** Don't bundle unrelated work. If a refactor enables a feature, consider splitting into two PRs.

### Safety (both modes)
- **Never force-push** without explicit user confirmation. Present what will be overwritten and the risk.
- **Never merge to main/master** without user confirmation.
- **Never push to remote** without user confirmation.
- **Never rewrite published history** (`rebase` on shared branches, `reset --hard`, `amend` on pushed commits) without confirmation.
- **Never delete branches** (local or remote) without confirmation.
```

- [ ] **Step 2: Verify rewrite against spec**

Read back `global/rules/workflow/git-workflow.md` and confirm:
- Frontmatter `alwaysApply: true` is preserved
- Direct Mode contains: commit confirmation, pre-commit gate, next steps suggestion, no proactive git actions
- Isolated Mode contains: autonomous commits, autonomous branch creation, control gate at end
- Isolated Mode activation: worktree OR dedicated feature branch (agent-created or explicitly instructed)
- Checkpoints section: 4+ tasks threshold, every 3 tasks, AskUserQuestion, progress pulse (not code review)
- Conventions section: conventional commits, small/focused, no AI attribution — all marked "both modes"
- Branches section: kebab-case naming, never commit to main/master
- Pull Requests section: conventional format title, description, one PR per change
- Safety section: force-push, merge to main, push to remote, rewrite history, delete branches — all require confirmation, all marked "both modes"
- Old rule "Never ask for git actions proactively" no longer exists as a standalone Safety rule (moved to Direct Mode as "No proactive git actions")

- [ ] **Step 3: Commit**

```bash
git add global/rules/workflow/git-workflow.md
git commit -m "refactor: restructure git-workflow into dual mode (direct/isolated)"
```

---

### Task 2: Enforce AskUserQuestion in CLAUDE.md Communication section

**Files:**
- Modify: `global/CLAUDE.md:59` (single line replacement)

**Spec reference:** `_support/spec/2026-04-01-git-workflow-dual-mode-design.md` section: AskUserQuestion Enforcement.

- [ ] **Step 1: Replace the Communication bullet on line 59**

Current line 59:
```markdown
- **Use interactive questions** (structured choices) when the decision has discrete options.
```

Replace with:
```markdown
- **Use `AskUserQuestion`** for any question with 2+ discrete options. Never present choices as plain numbered text. Reserve prose for open-ended questions.
```

- [ ] **Step 2: Verify the change**

Read back `global/CLAUDE.md` lines 58-63 and confirm:
- Line 59 now references `AskUserQuestion` explicitly
- The surrounding lines (58, 60-63) are unchanged
- The instruction is clear: 2+ discrete options → `AskUserQuestion`, never plain numbered text

- [ ] **Step 3: Commit**

```bash
git add global/CLAUDE.md
git commit -m "docs: enforce AskUserQuestion for discrete choices in Communication rules"
```
