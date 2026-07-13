# Gap Resolution in Plans — Design Spec

> **For agentic workers:** This is a design spec, not an implementation plan. It describes WHAT to build and WHY. The implementation plan will be created separately.

**Goal:** Ensure that gaps detected during planning always become actionable tasks or are explicitly escalated to the user — never left as passive notes.

**Mechanism:** A new global rule at `global/rules/workflow/gap-resolution.md` that introduces a mandatory Gap Analysis phase before writing plan tasks.

**Scope:** All planning activities — `writing-plans` skill, plan mode, brainstorming specs, and any session where implementation tasks are being designed.

---

## Problem Statement

When plans are written (via `writing-plans` or similar), the model detects gaps — prerequisite work needed for the feature to succeed — but treats them as observations rather than actionable work. Gaps are mentioned in notes, assumptions sections, or self-review comments, but never materialize as tasks in the plan.

This is a prompt engineering problem: current instructions say "list any gaps" (validation check) rather than "for every gap, create a task or escalate" (action trigger). The result is consistent across sessions because no global rule contradicts this behavior.

## Principle

**A gap detected during planning that does not become an actionable task or get escalated to the user is a defect in the plan.**

## What Is a Gap

Any preexisting condition that must be resolved for the planned feature/change to work correctly:

| Category | Examples |
|---|---|
| **Existing code** | Function needing refactor, incomplete interface, coupled module requiring decoupling |
| **Spec/Design** | Unresolved design decision, ambiguous requirement, undefined contract between services |
| **Infrastructure/Deps** | Missing dependency, required DB migration, incomplete env config, CI/CD setup |
| **Testing** | Missing tests for code being modified, nonexistent fixtures, coverage gaps in critical areas |

### What Is NOT a Gap (but must be surfaced)

- **Unrelated tech debt:** Detected during exploration but does not block the current change. Presented to the user as an observation — user decides whether to include in plan, create a separate plan, or ignore.
- **"Nice to have" improvements:** Enhancement opportunities that are not prerequisites. Same treatment — reported with a recommendation, user decides.

**Rule: never silently discard.** Gaps become tasks. Tech debt and improvements are presented to the user with a recommendation, but the decision is theirs.

## Gap Analysis Phase

A structured phase that runs **after** reading the spec/requirements and **before** writing the first task of the plan.

### Step 1: Active Codebase Exploration

Deliberately explore the codebase looking for problems — not a quick glance, but targeted investigation:

- **Files to modify** — current state, dependencies, existing tests
- **Interfaces/contracts** the feature needs — do they exist, are they complete?
- **Dependencies** — installed, correct version, configuration present?
- **Infrastructure** — migrations needed, environment variables, CI/CD adjustments?

### Step 2: Classify Findings

Each finding is classified into one of three categories:

| Classification | Criteria | Action |
|---|---|---|
| **Inline gap** | Prerequisite for the feature, ≤3 tasks to resolve | Include as first tasks in the plan |
| **Prerequisite gap** | Prerequisite for the feature, >3 tasks to resolve | Propose as a separate plan to the user before continuing |
| **Observation** | Non-blocking tech debt or improvement | Present to user with recommendation; user decides |

### Step 3: Present to User

Before writing plan tasks, present all findings:

- List of gaps with classification and justification
- Observations (tech debt/improvements) with recommendation
- Request user confirmation before proceeding

Only after confirmation: write plan tasks, starting with inline gaps as the first tasks.

### When No Gaps Are Found

Declare explicitly: "No gaps or pending prerequisites found." This confirms the analysis was performed — not that it was skipped.

## Integration with Existing Rules

### Complements `writing-plans` skill

The skill's self-review says "list any gaps, add missing tasks." This rule does not contradict it — it reinforces it by moving gap analysis from a post-hoc check to a mandatory prior phase. If the self-review catches an additional gap missed during analysis, the same principle applies: convert to task or escalate.

### Complements `critical-thinking.md`

The existing rule says "Surface risks before proceeding." Gaps are a specific type of risk — this rule operationalizes them with a concrete process instead of leaving it as a general principle.

### Complements `cross-service-workflow.md`

That rule covers gaps in cross-service contracts specifically. This rule is broader — covers gaps in code, infra, deps, and testing within a single service. No overlap because cross-service activates only with 2+ repos/services.

## Files to Create/Modify

| File | Action |
|---|---|
| `global/rules/workflow/gap-resolution.md` | **Create** — the complete rule |
| All other files | No changes needed |

## Success Criteria

- Plans include prerequisite tasks for detected gaps before feature tasks
- Large gaps (>3 tasks) trigger a proposal for a separate prerequisite plan
- Tech debt and improvements found during exploration are presented to the user, not silently discarded
- "No gaps found" is explicitly declared when exploration finds nothing
- The gap analysis phase is identifiable in the planning output (not invisible)
