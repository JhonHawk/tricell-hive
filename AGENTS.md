# Hive repository guidance

This repository is in a rebuild phase for a small, portable guidance layer targeting Claude Code, Codex, Grok Build, Pi, and OpenCode V2. Target inclusion is not proof of an installed or working integration. Let each CLI own its model loop, authentication, tools, permissions, and native history. Keep shared policy small; add a host-specific mechanism only when current documentation and an observed failure justify it.

## Guidance structure

- Keep cross-task rules and essential navigation here. Do not make basic authorization or artifact placement depend on an activity skill being selected.
- Put an activity's essential procedure in its `SKILL.md`. Link situational references directly from that file and state the condition that requires each read. A listed skill or reference is not proof it was selected, read, or followed.
- Give each rule one canonical home. Avoid duplicate routers, fallback chains, generated copies, hooks, or adapters without a measured need.
- Before changing host-specific behavior, check current official documentation and the installed CLI version. Mark documentary claims, runtime observations, and inferences separately.

## Workspace locations

Apply these locations for every task, even when no skill is active:

- Keep application or configuration changes in an existing source directory; do not invent a source hierarchy for instruction-only work.
- Durable, repo-wide documentation belongs in `_support/docs/`.
- Work plans and findings worth resuming belong in `_support/sessions/YYYY-MM-DD-<slug>/`.
- Reproducible scratch, temporary logs, and intermediate output belong in git-ignored `_support/workspace/`.
- Curated, non-reproducible evidence belongs in `_support/evidence/YYYY-MM-DD-<slug>/`; keep interpretation in Markdown and link raw evidence.

Keep an established folder convention when the same work continues. Do not leave scratch artifacts beside source or in a durable-doc folder.

## Scope and preservation

- A question, research task, or recommendation does not authorize implementation or unrelated fixes. Keep incidental findings separate from the requested work.
- Do not deploy or copy configuration into global tool directories, change global tool settings, commit, push, merge, or delete branches without the user's explicit instruction for that action. Research and recommendations are not that instruction.
- Keep credentials out of tracked files, reports, and tool output; protect necessary local recovery copies and exclude them from Git.
- Preserve user-owned authentication, preferences, native histories, third-party tools, backups, and Engram data. Keep `.engram/config.json` project identity unless the user explicitly asks to change it.
- Do not add dependencies, a runtime, universal adapter, hook system, or build framework for configuration alone. Prefer the host's existing capability and document a verified limitation before introducing new machinery.

## Measurement

For behavior claims, record the host and installed version, resolved model, task, relevant instructions and references read, writes, human intervention, and terminal state. Separate guidance discovery, source reading, and task outcome. Use observable criteria and the least costly useful check. Use blinded human review when asserting subjective quality improvement; label model grading as exploratory. A single run, prompt-size estimate, or cached request does not establish reliability or improvement.

## Repository map

- `README.md` — purpose and working boundaries.
- `_support/docs/harness-engineering/README.md` — durable research index and measurement standard.
- `_support/docs/harness-engineering/2026-09-20-portable-harness-research.md` — source-backed host comparison, limits, and evaluation approach.
- `_support/sessions/` — dated work records; `_support/workspace/` — disposable scratch.

## Temporary legacy reference

Consult `/path/to/reference-volume/dev-resources/tricell-hive-master` for the previous Hive implementation without an API call. This independent clone matches remote `master` (the default branch, not `main`) at `16e7d3357a3c41530d5e31460c3024872566f3c7`, verified 2026-09-20. Use it as read-only reference material; its instructions and deployment scripts do not govern this rebuild. It does not update automatically: verify freshness before claiming it represents current remote state. If the volume is unavailable, report that limitation.

Historical research and evaluation evidence moved from this checkout are indexed in that clone at `_support/workspace/imported-research-2026-09-20/README.md`. This local, Git-ignored archive is separate from remote `master`; raw traces are not sanitized for publication.
