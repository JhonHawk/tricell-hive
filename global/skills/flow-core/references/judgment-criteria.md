# Workspace judgment criteria (the center)

The judgment core `/flow-hygiene audit` applies to a workspace, and the lens `/flow-adopt`
reuses while tiering and routing artifacts. Four questions, all judgment — never a pass/fail
conformance sweep:

1. **Git state.** Uncommitted session artifacts in a versioned home (captured plans, findings,
   reports sitting untracked) → propose the standing-authorized `chore(sessions): <slug>` commit.
   Moves use `git mv` inside a repo, plain `mv` across a repo boundary. Stale branches / leftover
   worktrees are noted, not force-resolved.
2. **Loose or stray files — preserve or not?** Judge each: keep in place, relocate, or expire.
   **Archive on doubt** — any hesitation archives (never a silent delete). Stale captured plans
   (`Status: planned`, older than ~2 weeks with no matching execution) → recommend adopt via
   `/flow-build`, conclude, or expire. Deletions execute ONLY after the user types `eliminar`
   plus the IDs. Preserve the raw (`_support`, gitignored) / curated (specs, versioned) split.
3. **Broken pointers.** Session back-references that dangle, stale `PROJECT.md` rows, a missing
   `Tracker` / `Tracker access` field. Detect the tracker from observable signals — ticket-key
   patterns in commits and branches (`FAC-48` ~ Linear, `ATSCL-2406` ~ Jira), conventions in
   AGENTS.md/CLAUDE.md, which MCP/CLI responds — and propose the field values with the evidence;
   the user confirms.
4. **Forward-formality declared.** When the workspace has a specs repo (`Specs repo` ledger row
   or a sibling `*-specs/`), check the ambient pair declares a forward-documentation convention
   for operator-facing work — spec before code, a product-state home (`product/<module>/<vista>`),
   docs/manual updated in the same change-group. Missing → propose installing the declaration in
   the workspace `AGENTS.md`; never retro-audit past work against it.

**Convention is suggested, never enforced.** Canonical structure, ISO date renames, naming
schemes, initiative-vs-flat layout: SUGGEST them when they clearly help navigation, never audit
them as findings. A workspace that deviates from the convention without harm produces NO finding.
