# Gap Resolution Rule — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use optional reference project:subagent-driven-development (recommended) or optional reference project:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a global rule that ensures gaps detected during planning always become actionable tasks or get escalated to the user.

**Architecture:** Single markdown rule file in `global/rules/workflow/` following existing conventions (frontmatter with `alwaysApply: true`, H2 heading, blockquote context line). The rule defines a principle and a 3-step Gap Analysis phase that runs before writing plan tasks.

**Tech Stack:** Markdown (Claude Code rule format)

---

### Task 1: Create the gap-resolution rule

**Files:**
- Create: `global/rules/workflow/gap-resolution.md`

**Reference files (read for convention alignment):**
- `global/rules/workflow/cross-service-workflow.md` — frontmatter and structure pattern
- `global/rules/workflow/git-workflow.md` — table and subsection pattern

- [ ] **Step 1: Read existing workflow rules for convention reference**

Read both reference files to confirm the frontmatter format, heading level, and general structure before writing.

Run:
```bash
head -5 global/rules/workflow/cross-service-workflow.md
head -5 global/rules/workflow/git-workflow.md
```
Expected: Both start with `---\nalwaysApply: true\n---` followed by an H2 heading.

- [ ] **Step 2: Create the rule file**

Create `global/rules/workflow/gap-resolution.md` with the following exact content:

```markdown
---
alwaysApply: true
---

## Gap Resolution in Plans

> Ensures gaps detected during planning become actionable tasks — never passive notes. Complements `writing-plans` self-review and `critical-thinking.md` risk surfacing.

### Principle

**A gap detected during planning that does not become an actionable task or get escalated to the user is a defect in the plan.**

### What Is a Gap

Any preexisting condition that must be resolved for the planned feature/change to work correctly:

| Category | Examples |
|---|---|
| **Existing code** | Function needing refactor, incomplete interface, coupled module requiring decoupling |
| **Spec/Design** | Unresolved design decision, ambiguous requirement, undefined contract between services |
| **Infrastructure/Deps** | Missing dependency, required DB migration, incomplete env config, CI/CD setup |
| **Testing** | Missing tests for code being modified, nonexistent fixtures, coverage gaps in critical areas |

### What Is NOT a Gap (but must be surfaced)

- **Unrelated tech debt:** Does not block the current change. Present to the user as an observation — user decides whether to include in plan, create a separate plan, or ignore.
- **"Nice to have" improvements:** Enhancement opportunities that are not prerequisites. Same treatment — report with a recommendation, user decides.

**Never silently discard.** Gaps become tasks. Tech debt and improvements are presented to the user with a recommendation, but the decision is theirs.

### Mandatory Gap Analysis Phase

Runs **after** reading the spec/requirements and **before** writing the first plan task. Applies to all planning activities: `writing-plans` skill, plan mode, brainstorming specs, and any session where implementation tasks are being designed.

#### Step 1: Active Codebase Exploration

Deliberately explore the codebase looking for problems — not a quick glance, but targeted investigation with glob/grep/read:

- **Files to modify** — current state, dependencies, existing tests
- **Interfaces/contracts** the feature needs — do they exist, are they complete?
- **Dependencies** — installed, correct version, configuration present?
- **Infrastructure** — migrations needed, environment variables, CI/CD adjustments?

#### Step 2: Classify Findings

| Classification | Criteria | Action |
|---|---|---|
| **Inline gap** | Prerequisite for the feature, ≤3 tasks to resolve | Include as first tasks in the plan |
| **Prerequisite gap** | Prerequisite for the feature, >3 tasks to resolve | Propose as a separate prerequisite plan to the user before continuing |
| **Observation** | Non-blocking tech debt or improvement | Present to user with recommendation; user decides |

#### Step 3: Present to User

Before writing plan tasks, present all findings:

- List of gaps with classification and justification
- Observations (tech debt/improvements) with recommendation
- Request user confirmation before proceeding

Only after confirmation: write plan tasks, starting with inline gaps as the first tasks.

#### When No Gaps Are Found

Declare explicitly: "No gaps or pending prerequisites found." This confirms the analysis was performed — not that it was skipped.
```

- [ ] **Step 3: Verify the file was created correctly**

Run:
```bash
cat global/rules/workflow/gap-resolution.md
```
Expected: File starts with `---\nalwaysApply: true\n---`, contains H2 `## Gap Resolution in Plans`, and includes all sections: Principle, What Is a Gap, What Is NOT a Gap, Mandatory Gap Analysis Phase (with Steps 1-3 and "When No Gaps Are Found").

- [ ] **Step 4: Verify no conflicts with existing rules**

Confirm no other workflow rule covers the same topic:
```bash
grep -rl "gap.*resolution\|gap.*analysis\|prerequisite.*task" global/rules/workflow/
```
Expected: Only `gap-resolution.md` matches.

- [ ] **Step 5: Commit**

```bash
git add global/rules/workflow/gap-resolution.md
git commit -m "feat: add gap resolution rule for planning workflows"
```

---

## Self-Review

**1. Spec coverage:**
- Principle ("gap without task = defect") → ✅ Section "Principle"
- Gap definition with categories table → ✅ Section "What Is a Gap"
- Non-gaps that must be surfaced (tech debt, improvements) → ✅ Section "What Is NOT a Gap"
- 3-step Gap Analysis phase → ✅ Section "Mandatory Gap Analysis Phase" (Steps 1-3)
- Size-based classification (≤3 inline, >3 separate plan) → ✅ Step 2 table
- "No gaps found" explicit declaration → ✅ "When No Gaps Are Found"
- Integration notes (complements writing-plans, critical-thinking, cross-service) → ✅ Blockquote in heading
- Files to create/modify → ✅ Single file, no other changes
- Success criteria → ✅ All 5 criteria are addressed by the rule content

**2. Placeholder scan:** No TBD, TODO, "implement later", "add appropriate", or "similar to Task N" found.

**3. Type consistency:** N/A — no code types, functions, or method signatures. Terminology is consistent: "gap", "inline gap", "prerequisite gap", "observation" used uniformly.
