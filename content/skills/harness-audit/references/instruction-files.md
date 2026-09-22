# Instruction-file content

Apply these rules after computing the effective set in [hierarchy](hierarchy.md). Judge each rule by its effect in the sessions that load it.

## Admission test (HA-IF-01)

An always-loaded file keeps:

- one sentence on what the project is;
- build, test, and verification commands that are not the ecosystem default;
- gotchas: what an agent would get wrong reasoning from the code alone;
- upward pointers that sessions cannot reach on their own.

Everything else needs a stated reason to be resident. Classify the rest:

- derivable from the repository → `discoverable`;
- true but situational → `demote`;
- a prohibition without a nameable failure mode → `soften`.

## Content rules

- **HA-IF-02 — Derivable content.** Repository overviews, directory trees, dependency and script lists, and framework names that manifests and configuration already state are `discoverable`. Apply the no-op test: remove the line and name what the agent would do differently. Verify against disk, not against the rule's wording.
- **HA-IF-03 — Size as cost.** Report files over 200 lines and Codex project chains near `project_doc_max_bytes` as context cost per entry point. Size alone does not predict adherence; the admission test decides what goes.
- **HA-IF-04 — Map, not manual.** Multi-step procedures, inventories, and environment detail belong in a document or skill read when needed. The resident file keeps the gate and a one-line pointer that states when to read the target.
- **HA-IF-05 — Deterministic enforcement.** A rule that a linter, formatter, test, hook, or CI job can enforce belongs there. When such a mechanism already enforces it, the prose rule is covered; propose `delete`, except a line that tells the agent not to duplicate what the mechanism already does.
- **HA-IF-06 — Checkable criteria.** Verification guidance names exact non-standard commands and observable pass conditions. Adjectives such as "clean" or "robust" without a check are `soften` or need a concrete criterion.
- **HA-IF-07 — Soften unexplained prohibitions.** Rewrite `never X` or `always Y` that cannot name the failure it prevents as the criterion it stood for. A prohibition whose failure is nameable stays.
- **HA-IF-08 — One canonical home.** Flag duplicates and paraphrases across files and levels; propose one owner and `delete` or `re-anchor` the rest. Flag contradictory pairs and never recommend keeping both; an intentional override must say what it overrides and why.
- **HA-IF-09 — Stale paths.** Check backtick paths only when the token contains `/` and ends in an extension; skip tokens starting with `@`, `/`, `http`, or `~`, and tokens containing `://`, `<`, `>`, `*`, or spaces. This filter deliberately skips bare root-level names to avoid false hits; check those by hand when a file names one as a path. Retry unresolved paths by basename before reporting: a moved path is rewritten, an absent one is `stale` and deleted.
- **HA-IF-10 — Recover the rationale.** Before `delete`, `demote`, or `soften`, read the rule's history (`git log -S "<line>"`, `git blame`) and any ticket it cites. A rule whose reason still holds is not a candidate. When the reason is gone or was never recorded, say which. Without version history, mark the rationale unrecoverable and lower the finding's confidence.
- **HA-IF-11 — Surviving rules keep their reason.** A gotcha keeps one short line of reasoning; other surviving rules keep a ticket or commit pointer. A comment that restates the rule without a reason is noise.
- **HA-IF-12 — HTML comments.** Comments in `AGENTS.md` reach the model in every host and cost tokens; Claude Code strips them only from `CLAUDE.md`. Hiding rationale in comments is not free.
- **HA-IF-13 — Coverage for deletion.** Count as coverage only the project's own instruction layers and its deterministic mechanisms. Overlap with a personal or machine-wide global layer is reported as cost, not `delete`, because collaborators and CI may run without that layer, unless the project declares that layer a requirement. This holds even when the repository authors that global layer: its deployed copy reaches sessions only after deployment, so it is not the repository's own coverage.
- **HA-IF-14 — Floor.** A file holding only what the admission test keeps is `already lean`. Do not propose cutting its project description, non-default commands, or needed pointers.
- **HA-IF-15 — Local files.** Read and verify `CLAUDE.local.md` like any other file; its stale claims outrank correct ones because it loads last. Report its findings as `info` for its owner and never propose writing it.
- **HA-IF-16 — Imports.** `@path` imports load at launch and stop after four hops; splitting a file into imports organizes it but does not reduce context. A backticked `` `@path` `` is literal text, not an import.

## Evidence for each finding

Quote the line, name the effective sets it affects, and state the rationale found (HA-IF-10). For `discoverable` and `stale`, cite the file or command output that already states or contradicts it.
