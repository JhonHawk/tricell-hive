---
name: harness-audit
description: Audit coding-agent instruction files (AGENTS.md and CLAUDE.md at workspace, repository, and nested package levels), skills, and subagent definitions for harness-engineering quality, and run native static validators. Use when asked to audit, review, prune, or check agent instructions, skills, or agents, or to explain what a host loads from an entry directory.
---

# Harness audit

Audit what coding-agent hosts actually load and whether each instruction, skill, or agent earns its place. Report findings with evidence; do not edit. The caller decides and applies changes.

## Scope

Establish before reading deeply:

- **Targets and artifact types**: instruction files, skills, agents, or a subset.
- **Entry points**: the directories where sessions start. A file's effect depends on the entry point, not on the file alone.
- **Hosts in use**: Claude Code and Codex unless the caller names others. Treat other hosts only on explicit request and do not assert their loading behavior without their own evidence.
- **Authority**: reading another project is allowed; proposing changes for it requires the caller's assignment of that audit.

## Procedure

1. For instruction files, read [hierarchy](references/hierarchy.md) and compute the effective set per host and entry point before judging any single file. Then read [instruction files](references/instruction-files.md) and apply its admission test and rules.
2. For skills or agents, read [skills and agents](references/skills-and-agents.md), together with the project's own authoring conventions.
3. Before running any native tool, relying on its output, or asserting version-sensitive loading behavior, read [native tools](references/native-tools.md).
4. Before proposing `delete`, `demote`, or `soften`, or reporting `already lean`, for any artifact type, read [instruction files](references/instruction-files.md) for HA-IF-10, HA-IF-13, and HA-IF-14, and recover the rule's rationale (HA-IF-10).
5. Report in the format below. Prefer fewer, consequential findings over exhaustive restatement.

## Outcomes

| Outcome | Use when |
| --- | --- |
| `delete` | Covered by the project's own layers or a deterministic mechanism (HA-IF-13), or a functionless note such as editing advice addressed to maintainers inside a loaded file. |
| `discoverable` | The repository already states it; removing the line changes nothing the agent would do. |
| `stale` | A factual claim that does not match reality: a path, a command, or a statement about what a host loads (HA-HI-02). Rewrite per host when hosts differ. |
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

## Boundaries

Stay read-only: do not edit files, configuration, trackers, or memory, and do not commit. Never propose writing `CLAUDE.local.md`. Run only static tools your boundary allows, and request the rest from the caller; run a model-invoking evaluation only when the caller authorizes it with a cost cap (HA-ME-01).
