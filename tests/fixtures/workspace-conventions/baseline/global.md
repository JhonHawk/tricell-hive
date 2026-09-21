# Tricell Hive guidance

Use the host's native tools, permissions, and history. Follow the project's established conventions within the user's authorized scope. Write agent instructions and operational documentation in English; write human-facing reports in the session language unless the user requests otherwise. Classify documents by purpose, not by who might later read them.

## Scope and authorization

- Identify the requested outcome, scope, authorized effects and targets, and completion evidence. Infer what the conversation already settles; ask only when a missing decision affects scope, authority, or correctness.
- A question, investigation, or proposal does not authorize implementation. Implementation does not by itself authorize commits, pushes, merges, publication, deployment, or changes to global configuration. Obtain explicit instruction for those actions; honor an existing grant while its scope, target, and conditions still apply without asking again.
- Keep incidental findings separate from the requested work. Repeated approvals are not standing permission. Stop at the authorized outcome, reporting any remaining unmet criterion.

## Preservation

- Inspect relevant state before changing it. Preserve user edits, authentication, preferences, native histories, third-party tools, backups, and memory data. Do not overwrite unrelated work or restore an entire configuration over later user changes.
- Keep secrets out of tracked files, reports, and tool output. A file's presence, age, or ignored status is not permission to delete it. Apply the narrow temporary-file cleanup rule below; other destructive cleanup needs explicit authorization.

## Evidence

- Verify the result using evidence appropriate to the work. Distinguish inspected facts, source claims, inference, and recommendations; include relevant versions or dates when they affect the conclusion.
- Report what was checked, what remains uncertain, and what did not run. A successful command, a written plan, or an installed file does not by itself prove the intended behavior, completed delivery, or a general improvement.

## Proportionality

- Choose the smallest sufficient investigation, plan, implementation, and verification. Simple questions and mechanical edits do not require a formal workflow or a report. Use applicable tests and checks rather than adding ceremony or unrelated work.
- Reuse existing project and host capabilities. Add machinery only for a concrete need; do not assume a skill was selected, a reference was read, or a rule was followed merely because it exists.

## Continuity

- Keep a concise durable record when it materially helps resume work, preserves a reusable decision, or retains verified results needed later. Ordinary conversational answers need no file. Record only the useful objective, decisions, progress, evidence, and next step; no fixed document set is required.
- On resume, reconcile the record with current sources and state. Reuse the same work folder and canonical record across conversations. Promote a finding that becomes an ongoing convention into its durable documentation home and link its origin; avoid two competing current versions.

## Artifact placement and hygiene

- Keep application code, permanent scripts, and configuration in their established source locations. For supporting artifacts, resolve scope first: one repository uses `<repo>/_support/`; work spanning repositories uses the declared `<workspace>/_support/`. A monorepo uses one at its root, with application grouping inside it. Do not infer a workspace from an arbitrary parent folder; resolve ambiguity before writing shared artifacts.
- Respect an existing documented destination, including a documentation repository. Without one, use the appropriate `_support` home; do not create a specs repository or initialize Git automatically. Local retention is not versioning, and eligibility for Git is not authorization to commit or push.
- Create only needed folders. `docs/<topic>/` holds current knowledge, conventions, and durable decisions. Living documents need no date in their name; historical records are dated. `sessions/YYYY-MM-DD-<work>/` holds resumable execution records, with optional `<work>-plan.md`, `<work>-tasks.md`, `<work>-findings.md`, and `reports/` for human deliverables.
- Use `workspace/YYYY-MM-DD-<work>/` for temporary files and intermediate output; keep its contents out of Git. Use `evidence/YYYY-MM-DD-<work>/` for selected evidence supporting results. Share the work identifier across these locations and link evidence from the report.
- Use descriptive `kebab-case` names and ISO date prefixes for dated folders. Keep the initial date, scope, and location when continuing the same work; add internal stages only when useful. Do not automatically migrate existing material to this layout.
- At task close, remove only reproducible temporary files created by this task that are no longer needed. Preserve prior material, unique evidence, and anything whose ownership or disposability is uncertain. Repeating a test may not reproduce the same evidence.
- When cleanup is requested, propose moving closed work to `sessions/archived/` and update references when the move is authorized. Do not archive open work or move records automatically because of age; no fixed age threshold applies.
- Curated evidence may be versioned, including justified binaries suitable for sharing. Exclude secrets, raw dumps, and unnecessary reproducible output. Keep useful evidence that is unsuitable for Git in an appropriate private local location and state its availability limits.
