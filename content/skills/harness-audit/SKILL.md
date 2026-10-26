---
name: harness-audit
description: Audit coding-agent instruction files (AGENTS.md and CLAUDE.md at workspace, repository, and nested package levels), skills, and subagent definitions for harness-engineering quality, run native static validators, ask the decisions the findings need, and apply the accepted outcomes in the current project. Use when asked to audit, review, prune, fix, or check agent instructions, skills, or agents, to set up or update a repository's Hive settings section (onboarding a project to Hive), to explain what a host loads from an entry directory, or to prepare a project's migration to the `openspec` change-record layout.
---

# Harness audit

Audit what coding-agent hosts actually load and whether each instruction, skill, or agent earns its place. Report findings with evidence, ask the decisions they need, and, after the user's answer, apply the accepted outcomes in the current project, as [Apply accepted outcomes](#apply-accepted-outcomes) sets. A read-only role, such as `hive-review-harness`, only reports and returns those decisions to its parent, which asks and applies.

## Scope

Before resolving project settings needed by the audit or migration, read [project settings](../flow-plan/references/project-settings.md). Before applying an outcome that writes `Specs`, creates a specs directory, or creates or continues a change record, read [change records](../flow-plan/references/change-records.md); before writing or cleaning supporting artifacts, read [artifact layout and retention](../workspace-conventions/references/artifact-layout.md). An audit that applies nothing needs neither. Before findings or messages with citations, links or externally defined IDs, read [reporting](../flow-close/references/reporting.md); before a user decision, read [questions](../flow-plan/references/questions.md).

Establish before reading deeply:

- **Targets and artifact types**: instruction files, skills, agents, or a subset.
- **Entry points**: the directories where sessions start. A file's effect depends on the entry point, not on the file alone. Derive them from native session histories (see [hierarchy](references/hierarchy.md)) and report the distribution per host. When the caller declares none, audit the directories that histories show in use, not only the one the audit runs from. Report as a finding a declared entry point that differs from the observed ones, or an audit directory that is not the most used entry point.
- **Hosts in use**: Claude Code and Codex, plus any host whose history shows sessions for the targets, or that the caller names. A repository `.cursor/` directory suggests Cursor CLI; ask rather than assume it. Do not assert a host's loading behavior without its own evidence.
- **Authority**: reading another project is allowed; proposing changes for it requires the caller's assignment of that audit.

## Procedure

1. For instruction files, read [hierarchy](references/hierarchy.md) and compute the effective set per host and entry point before judging any single file. Then read [instruction files](references/instruction-files.md) and apply its admission test and rules.
2. For skills or agents, read [skills and agents](references/skills-and-agents.md), together with the project's own authoring conventions.
3. Before running any native tool, relying on its output, or asserting version-sensitive loading behavior, read [native tools](references/native-tools.md).
4. For a project's repositories, or when the caller asks to prepare a layout migration, read [project layout](references/project-layout.md) and include its migration manifest.
5. Before proposing `delete`, `demote`, or `soften`, or reporting `already lean`, for any artifact type, read [instruction files](references/instruction-files.md) for HA-IF-10, HA-IF-13, and HA-IF-14, and recover the rule's rationale (HA-IF-10).
6. Report in the format below. Prefer fewer, consequential findings over exhaustive restatement.
7. Ask the decisions and apply the accepted outcomes, as [Apply accepted outcomes](#apply-accepted-outcomes) sets.

When the caller asks only to set up or update a repository's `## Hive` section, such as onboarding a project to Hive, limit the audit to HA-IF-17 and HA-PL-01 for each repository it names and for the workspace section when sessions start there, and to the prose those rules relocate; then continue at step 6.

## Outcomes

| Outcome | Use when |
| --- | --- |
| `delete` | Covered by the project's own layers or a deterministic mechanism (HA-IF-13), or a note that changes nothing the agent does, such as maintainer editing advice or change history. |
| `discoverable` | The repository already states it; removing the line changes nothing the agent would do. |
| `stale` | A factual claim that does not match reality, such as a path, command, version, or statement about what a host loads (HA-HI-02). Rewrite per host when hosts differ. |
| `demote` | True and useful but situational: move it to a document or skill, leaving a one-line pointer with its read condition. |
| `soften` | Only for a prohibition without a nameable failure mode: rewrite as the criterion it proxied. |
| `keep` | Project-specific and a gotcha, or a required import or pointer. |
| `re-anchor` | Only for a copy of another rule that drifted from its canonical source, or an override that must read as intentional; never for a factual error. |
| `relocate` | Content sits at the wrong level; move it, never copy it. |
| `propose-global` | Belongs in a shared or personal global layer; report it for that layer's owner. |
| `fix` | A skill or agent defect with a concrete correction. |

Report a file already at its floor as `already lean` (HA-IF-14).

## Severity

- `high`: guidance a repository needs to operate reaches no host in use by any path; a contradiction with consequence; a broken required resource; a consequential skill or agent without its boundary. Content reachable through a repository's upward pointer counts as conditionally reached and is not `high`.
- `medium`: stale or misleading content, including loading claims that do not match a host; per-session cost without benefit.
- `low`: redundancy, surplus pointers, minor cost.
- `info`: observations without a proposed action, including anything about `CLAUDE.local.md`.

## Findings format

For each finding give severity, `file:line`, rule ID, outcome, evidence (quoted line, command output, or source), and whether it is confirmed or a hypothesis; for a hypothesis, name the exact check that would confirm it.

Close with coverage: targets, entry points, hosts, native outputs used, what was not verified, and decisions needed from the caller. Report `no material findings` when that is the result; it does not certify uninspected areas.

## Apply accepted outcomes

1. **Ask.** In one round, by following [questions](../flow-plan/references/questions.md), ask the decisions the findings need: each HA-IF-08 choice of which version to keep and each conflict with a global layer; each standing grant written in prose, to remove or keep as a stated exception; each `## Hive` value the repository's evidence does not settle; and which outcomes to apply, one option per group of findings with its files. When the applied edits will end in a Git delivery in a code repository, add to that round the delivery and review questions of [delivery decisions](../flow-plan/references/delivery-decisions.md), which must be answered before the first edit; `hold` keeps the edits uncommitted. Wait for the answers; a delegated auditor returns the decisions to its parent, which asks.
2. **Apply.** After the answer, edit only the current project, with the host's file-editing tool, and only the outcomes the answer accepted: `delete`, `discoverable`, `stale`, `relocate`, `re-anchor`, `soften`, `demote` with its one-line pointer, `fix`, `insert`, and the `## Hive` section with the facts [project settings](../flow-plan/references/project-settings.md) defines. Leave `keep` and `manual` entries untouched and list the `manual` ones as remaining work. Never apply `propose-global`, never write `CLAUDE.local.md`, and never edit another project: report its findings for that project's own session.
3. **Validate.** For each repository whose `## Hive` section you wrote, run `hive doctor --project <repository>`: exit status 0 and no line naming a retired key mean the section is valid; any other status means it is not, so fix it and run again. When `hive` is not on the `PATH`, and for a workspace section, which `hive doctor` does not check, reread the section against [project settings](../flow-plan/references/project-settings.md) and report that the command check did not run.
4. **Deliver.** Commits, pushes and pull requests follow the mode chosen in step 1, through [flow-build](../flow-build/SKILL.md) and [git-workflow](../git-workflow/SKILL.md).

Under [unattended delegation](../unattended-delegation/SKILL.md), leave unapplied and report the decisions that belong to the user: which version of a rule to keep, a conflict with a global layer, and each standing grant in prose.

## Boundaries

A role launched read-only, such as `hive-review-harness`, edits nothing: it returns findings, decisions and any migration manifest to its parent, which asks and applies. In the main thread, write only what [Apply accepted outcomes](#apply-accepted-outcomes) allows after the user's answer, and an approved migration manifest, whose mechanical entries run through the `workspace-conventions` procedure for applying a layout migration, stopping at `manual` and `ask` entries. Do not edit configuration outside the project, trackers, or memory. Never propose writing `CLAUDE.local.md`. Run only static tools your boundary allows, and request the rest from the caller; run a model-invoking evaluation only when the caller authorizes it with a cost cap (HA-ME-01).
