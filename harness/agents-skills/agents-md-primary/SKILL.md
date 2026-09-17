---
name: agents-md-primary
description: >
  Convert a project to the AGENTS.md-primary pattern: AGENTS.md becomes the single canonical
  harness-instructions file and CLAUDE.md becomes its `@AGENTS.md` import (plus genuinely
  Claude-specific content below the import). Use on projects where AGENTS.md and CLAUDE.md
  duplicate content, where only one of the pair exists, or to find conversion candidates across
  many projects (scan). Also audits a project's AGENTS.md/CLAUDE.md against the deployed global
  canon — harness coverage, duplication, stale forks, content quality, and instruction budget —
  and checks every declared claim against git and disk rather than taking the document's word for
  it. Subcommands: audit | apply. Idempotent.
disable-model-invocation: true
---

# /agents-md-primary — one canonical file, every harness ambient

Why this pattern: Claude Code reads CLAUDE.md; Codex, opencode, and the rest of the
AGENTS-compatible ecosystem read AGENTS.md. Duplicating content across both guarantees
drift. The inversion makes AGENTS.md canonical and CLAUDE.md a one-line importer
(`@AGENTS.md` — official Claude Code import, loaded at session start), with
Claude-specific instructions below the import when they genuinely exist.

**The import law (empirical, Claude Code 2.1.233 — undocumented upstream):** `@AGENTS.md`
resolves in the cwd's own CLAUDE.md and in lazily-loaded subtree CLAUDE.md files; it does
NOT resolve in eagerly-loaded ANCESTOR CLAUDE.md files (their literal text loads, the
import is skipped). Consequence for `convert`: at a repo root the pattern is complete for
sessions opened there and for parent-workspace sessions that touch the repo (lazy path
resolves); at a WORKSPACE root the pattern's content operates only in sessions opened at
that root — warn this trade in the convert report when the target is a workspace root.

## `scan <root>` — find candidates (read-only)

Walk `<root>` (default `~/Development/projects`, depth ≤ 4 to cover
`projects/<group>/<project>/<repo>`) and classify every directory that has AGENTS.md
and/or CLAUDE.md at its root:

- `converted` — CLAUDE.md carries the `@AGENTS.md` import (or is a symlink to AGENTS.md).
  Detect the import ANYWHERE in the file, not just on line 1: a heading or preamble above it
  resolves identically. Matching only the first line misreports conformant repos as `both`.
- `duplicated` — both exist with substantially overlapping content ← the targets
- `divergent` — both exist with conflicting or deliberately different content
- `claude-only` / `agents-only` — half the pair missing
- `blind` — the half that is missing makes the root invisible to a harness in use:
  `AGENTS.md` with no sibling `CLAUDE.md` (Claude Code does NOT fall back), or `CLAUDE.md`
  with no `AGENTS.md` (Codex/opencode/Grok). Count only REPO and WORKSPACE roots — a
  subtree file, `_support/backup/**`, and a deployable source tree (`global/`, `harness/`)
  are not findings.
- `unmanaged` — an active root (any child repo with a commit < 90 days) carrying NO
  instruction file at all. `convert` has nothing to trigger on here, so this is the only
  subcommand that ever finds it; the conventions usually exist in `docs/CONTRIBUTING.md`
  or `README.md`, which no harness loads. Propose promoting that file, never a stub.

Report as a table with a suggested action per row. Change nothing.

## Convert (default; `$1` = project path, default cwd)

1. **Classify** the pair at the target (same buckets as scan). `converted` → say so and
   stop (idempotency). A CLAUDE.md symlink to AGENTS.md is already a valid form of the
   pattern — report it, don't rewrite it.
2. **Single-file cases (proceed directly — fully reversible):**
   - `claude-only` → move the content to AGENTS.md verbatim; CLAUDE.md becomes the
     import. Fix self-references ("this CLAUDE.md" → "this AGENTS.md").
   - `agents-only` → create CLAUDE.md containing exactly `@AGENTS.md`.
3. **Both-files case — merge analysis before touching anything.** Diff
   section-by-section into three buckets:
   - **Shared/duplicated** → lives once, in AGENTS.md.
   - **Harness-neutral but unique to either file** → AGENTS.md.
   - **Genuinely Claude-specific** (plan-mode behavior, Claude skills/hooks/subagent
     names, AskUserQuestion, model/effort settings) → CLAUDE.md, below the import.
     Test: would this line mislead a Codex/opencode session? Yes → it stays Claude-side.
   - **Contradictions** (both files rule differently on the same topic) → never pick
     silently; list each with both versions (divergence protocol) for the user to
     resolve in the confirmation step.
4. **Gate by signal, not by case.** A clean both-files merge (no contradictions, no
   divergent-by-design pair) → proceed-and-report like the single-file cases: fully
   reversible (uncommitted diff in a git repo, `.bak` files otherwise — step 6), and the
   summary diff is the review surface. Confirm only on signal: **contradictions** from
   step 3 (present both versions — divergence protocol), or a pair `divergent` by design
   (mostly disjoint scopes — e.g. AGENTS.md as a local compatibility guide, not a
   duplicate) → the pattern may not apply; stop unless the user overrides.
5. **Write:**
   - `AGENTS.md` — the canonical content. Preserve the original language and wording;
     translating or rewriting prose is out of scope for this skill.
   - `CLAUDE.md` — first line `@AGENTS.md`, then a blank line, then the Claude-specific
     block if any (under a `## Claude Code` heading).
6. **Verify, no silent loss:** every section of both originals must be accounted for —
   moved, kept Claude-side, or dropped with the reason stated in the report. Show a
   summary diff. If the project is a git repo, leave the changes uncommitted and suggest
   the commit; if it is NOT a git repo, write `CLAUDE.md.bak`/`AGENTS.md.bak` first.

## `audit [path]` — dedup project rules against the global canon (read-only)

Target: one project/workspace root (default cwd), its child repos' `AGENTS.md`/`CLAUDE.md`,
**and every `AGENTS.md` NESTED inside those repos** (`find <repo> -name AGENTS.md`, excluding
`node_modules`). A monorepo's `apps/*/AGENTS.md` are subdirectories, not child repos — the
scope that stops at repo roots misses exactly where instructions accumulate, and in a
monorepo-by-default portfolio that is most of the volume. **Coverage canon = the DEPLOYED layers — what sessions actually load:**
`~/.claude/CLAUDE.md` + always-on `~/.claude/rules/`, the condensed core at
`~/.codex/AGENTS.md` / `~/.config/opencode/AGENTS.md`, Grok's `~/.grok/rules/` symlinks,
the deployed router-skill references, and opencode's rules plugin. **The audit is
self-contained on those layers: never locate, search for, or read the hub repo the
deployment came from.** Promote-to-core / inject-to-router outcomes are hub edits by
nature — report them as proposals for the user to take to a session opened in the hub;
this skill neither finds nor touches it.

**Per rule, compute the harness-coverage matrix — never a boolean, and with a LEVEL
axis.** Which DEPLOYED always-on layer already carries it: `~/.claude/rules` (Claude ✓,
Grok ✓ via symlink) · the condensed deployed core (Codex ✓, opencode ✓) · a router-skill
injection or the opencode rules plugin (situational reach). A project rule duplicating a
global-rules-only item is still LOAD-BEARING for Codex/opencode — deletion requires
coverage in every harness the project uses. The LEVEL of the audited file changes what
"coverage" means: a REPO-level file loads in child-repo sessions (all harnesses); a
WORKSPACE-ROOT file operates ONLY in sessions opened at the workspace root — from a
child-repo session no harness auto-loads it, and Claude Code's ancestor walk loads the
workspace CLAUDE.md but does NOT resolve its `@AGENTS.md` import (the import law below).

**The admission test, applied before the table below.** An always-loaded instruction file
stays *lightweight and briefly describes what the repo is for*, and what it keeps beyond that
is **gotchas** — what the agent would get wrong reasoning from the code alone. Everything else
has a home that is not the resident context: derivable from the repo → `discoverable`; true
but situational → `demote`; a prohibition with no nameable failure mode → `soften`. Vendor
guidance for the Claude 5 generation (`_support/docs/methodology-bibliography.md`): ~80% of
Claude Code's own system prompt was removed with no measurable eval loss, and the shift is
from rules to judgment. **Two rules that contradict each other cost more than either one
alone** — flag the pair, never keep both.

**The manifest is WRITTEN TO DISK, never left in the conversation.** Target:
`<audited-root>/_support/workspace/agents-md-audit-YYYY-MM-DD.md` (create the folder if the
workspace lacks one; fall back to `<audited-root>/_support/` where `workspace/` is not the
convention). One entry per rule: file, line, evidence, outcome, proposed action — the same
table the report shows. An audit is expensive to produce and outlives the session that ran
it, like a portable Hive plan: day-2 continuity lives in a repo file, not harness state. It also makes the manifest reviewable outside the terminal, which is where
the user reads everything else. Same-day re-audit overwrites; a later date gets its own file.

**The floor — what a root file keeps even when everything else goes.** One sentence saying
what the project is; the package manager only if it is not the ecosystem default; the
build/typecheck/test commands that are NON-standard. Plus this repo's own additions: the
`## Git Workflow` declarations and the pointers a child-repo session cannot reach on its own
(ledger, sibling map). That is the stopping condition — an audit that proposes cutting into
this floor has gone from pruning to breaking, and a file already at it is reported **already
lean**, not squeezed further. A package-level file inside a monorepo takes the same shape for
its own scope: what the package is, its stack, its conventions — never a copy of the root's.

**Recover WHY before proposing any removal.** Every `delete`, `demote` and `soften` first
reads the rule's rationale — `git log -S "<line>"` or blame on it, plus the ticket it cites —
and the manifest entry carries what it found. A rule whose reason still holds is not a
candidate however redundant it looks; one whose reason is **gone or was never recorded** is
the cheapest removal there is, and the manifest must say which of the two it is. Skipping this
is the mechanism that makes these files grow: appending costs nothing, deleting safely costs
remembering, so nobody deletes (`_support/docs/methodology-bibliography.md` — measured at
+226% per file lifetime, and older instructions get *harder* to delete, not easier).
**Surviving rules earn their reason recorded — in the form the rule's own difficulty
decides.** A rule the model would comply with WRONG without knowing why (a gotcha) keeps one
short line of reasoning inline: it is measured to raise instruction-following, so it buys
compliance, not only future deletability. A rule that is obvious to follow and merely needs to
be retirable tomorrow takes a **pointer** — a ticket key or commit, ~10 characters against a
paragraph. What never works is the middle: a comment that restates the rule without encoding
why ("comment-shaped noise") grows the file at nearly the rate of no comment at all. And there
is no free channel — an HTML comment reaches the model verbatim in every harness (measured on
Codex `prompt-input`), so `<!-- -->` hides nothing and costs the same as plain text.

**Classify each rule into exactly one outcome:**

| Outcome | When | Proposed action |
|---|---|---|
| **delete** | Full coverage — pure noise (and reclaims Codex's per-repo `project_doc_max_bytes` budget). Coverage is not only another instruction layer: a rule a DETERMINISTIC mechanism in the repo now enforces (hook job, CI gate) is covered the same way | Remove from the project file |
| **promote-to-core** | Universal (gate, every-session procedure), partial coverage | Condensed line into `harness/AGENTS.md` (28 KiB budget is the gate) → then delete from EVERY project |
| **inject-to-router** | Situational (language, workspace, memory policy), partial coverage | Add the owning global rule to `SKILL_REFERENCE_INJECTIONS` → delete from the project |
| **discoverable** | The repo's own files already state it — package manager (lockfile), scripts (`package.json`), framework (its config), directory inventory | Remove. Discriminator is the no-op test: delete the line and name what the agent would do differently. Nothing → it is a no-op. Verify against disk before proposing, never from the rule's wording |
| **stale** | An implementation detail that no longer matches the repo — distinct from `re-anchor`, which is drift against a GLOBAL rule | Resolve every backtick path against disk (below). Resolves elsewhere → **moved**: rewrite the path. Nowhere → **absent**: delete the claim |
| **demote** | Project canon, but NOT a gotcha: procedure, inventories, command lists, environment detail — true, useful, and not needed on every turn | Move to a doc the agent reads when it acts (`_support/docs/`, the specs repo), leaving a ONE-LINE pointer. Progressive disclosure: the always-loaded file keeps the gate and the pointer, the mechanics live where they are read. This is the outcome for content too valuable to delete and too situational to resident-load |
| **soften** | A prohibition (`never X`, `always Y`) that cannot name the failure mode it prevents | Rewrite as the criterion the rule was proxying for — *"write code that reads like the surrounding code: match its comment density, naming, and idiom"*, not *"never write multi-line comment blocks"*. Guardrails written for older models' worst case are the bulk of this; a rule whose incident IS nameable stays a rule |
| **keep — project canon** | Genuinely project-specific, needed cross-harness, **and a gotcha**: something an agent would get WRONG by reasoning from the code alone (module load order that breaks routing, an id validated in one place only, a type layout nothing hints at). Canon that is merely true is `demote`, not `keep` | Split by LEVEL: a REPO file stays — it is the canonical channel for all harnesses; a WORKSPACE-ROOT file operates only in workspace-root sessions — its cross-harness content is proposed DOWN to the governing repo (plus a per-repo pointer to ledger/workspace root). **Record why it is canon** — a ticket reference (`TRI-514`) is the preferred form; it holds the full reason outside the context budget. That note is what lets the NEXT audit delete the rule instead of re-deriving it |
| **re-anchor / explicit override** | Stale fork (paraphrased copy of an evolved global rule — the drift that produces contradictory instructions) or a deliberate contradiction | Rewrite quoting the canonical text, or as `overrides global <rule> because <reason>` — overrides legitimately WIN (`git-workflow.md` precedence); they must read as intentional |

**Resolving backtick paths (feeds the `stale` outcome).** A naive resolver is unusable — ~90% of
its hits are false. Keep only tokens that contain `/` AND end in a `.ext`, discarding anything
starting with `@ * / http ~` or containing `:// < > *` or spaces — that drops TS aliases, URL
routes, slash-commands, regexes and hostnames. Unresolved → retry by basename (`find -name`)
before reporting: most broken paths are MOVED, not absent, and the two take opposite fixes.

**Completeness checks — same audit pass, per child repo of a workspace:**

- **Project pointers + SIBLING MAP**: the repo `AGENTS.md` carries `Ledger: <relative path>`
  and the workspace-root path, with the instruction to open them for project-scope tasks
  (child-repo sessions never auto-load the workspace files), PLUS the siblings — name, role
  (runtime / specs / mocks / archived) and that they live one level up. Derive the map from
  `find -maxdepth 2 -name .git`, never from what the file already claims: an active runtime
  missing from a sibling list is how a session concludes the archived repos ARE the product.
  Missing or wrong → propose the block (~300 B).
- **Workspace content that a child-repo session needs MOVES down — it is never copied.** A
  session opened in a child repo does not load the workspace files at all (Codex starts at
  the git root; Claude Code loads an ancestor CLAUDE.md's literal text but skips its
  `@AGENTS.md` import — the import law above). Split by ownership, not by convenience:
  **down** goes what a single-repo session needs to operate (commands, code conventions, its
  git workflow, verification); **up stays** only what crosses repos and no repo can own (the
  ledger, the promotion chain, cross-repo resolution rules, the deployment map). What goes
  down is DELETED from the workspace file and replaced by nothing — the per-repo pointer
  above is what carries the reader back up. Two copies of one rule with no arbiter is the
  drift this skill exists to remove; creating it while conforming the pair is a regression.
- **`## Git Workflow` declarations**: `Base branch:` / `PR review:` /
  `Issue tracker:` per `git-workflow.md`'s declaration block. A declared value settles a fact
  or seeds a recommendation; it never silences a decision. `Base branch` is a repo FACT: verify it against branch topology, never ask. The
  other three are DECISIONS no repo file can settle on its own. **The block lives in the
  REPO's `AGENTS.md`, one per repo — never at the workspace root:** the values differ per repo
  and a child-repo session never loads the root file; the root keeps only what crosses repos
  (the promotion chain, the project's tracker). Each decision is confirmed in the apply, never
  inferred silently:
  - **`Git mode` is no longer declared — an existing one is a FINDING to remove.** The mode is
    asked every session and nothing silences it (`git-mechanics.md > Commits`), so a declared
    token is a stale default that answers nothing. Removing it keeps whatever the line carried
    BEYOND the token — which surface needs visual validation, a standing push authorization —
    as its own line; that knowledge is the repo's, not the mode's.
  - **`PR review`** — the review apps wired to the repo tell you WHICH app to RECOMMEND;
    the route is still asked every session that opens a PR, with the declared app first and
    marked recommended and `none` among the options, and a paid app's trigger rides the
    user's sign-off (`git-mechanics.md > PRs & promotion`). A declaration phrased as a
    standing order ("after the push, comment `bugbot run`") is a FINDING: rewrite it as the
    recommendation it is.
  - **`Issue tracker`** — standing write authorization to an external system, confirm-gated in
    its own right (`memory-routing.md > Tracker sync`); never carried over as settled because a
    previous version of the file already said it.
- **Instruction budget — measure the CHAIN a session actually loads, never a file alone.**
  Codex concatenates global → git-root → cwd, so a nested file is charged with everything
  above it; `codex debug prompt-input` renders that sum exactly and is the measurement of
  record (Claude Code has no equivalent). Report the chain per entry point, and the global
  file's share of it — a core paid in every session of every project outranks a repo file in
  cost even when it is smaller on disk. `project_doc_max_bytes` (32 KiB default) is a
  truncation cap, NOT the budget: the real ceiling is how many instructions a model reliably
  follows (~150–200, with the harness system prompt already spending part of it), which is
  why a chain can be well under the cap and still be over-instructed.
  **The pruning criterion is `keep vs demote`, not line count** (`> Pruning`).
- **Verification parity**: derive the repo's layers from disk — `package.json` scripts,
  `lefthook.yml` / `.husky/`, `.github/workflows/` — and compute two gaps against CI.
  **Invocation**: pre-commit does not run on commits created by `cherry-pick` or `rebase`
  (measured: 0 runs), nor under `-n`, so a check living only there is absent from exactly the
  operations that prepare a PR. **Coverage**: repo paths no job's `glob` matches — shell
  scripts, root configs, workflow files. Widening globs never closes the invocation gap; only
  a pre-push job does. **The fix belongs in the repo's EXISTING hook manager, never in
  prose** — never introduce one — scoped to what CI already runs minus the full test suite
  (tests stay on the affected-subset rule, `testing.md > Execution Scope`).
  `AGENTS.md` takes only what the hook cannot enforce: in a repo with no hook manager, the
  imperative instruction (declared prompt-convention). Every verification line already in the
  file is then re-checked against the real map, **both directions**. One promising coverage
  the repo does not have is `stale` — correct or delete it; that lie is what produces the
  failure this check prevents. One prescribing by hand what a deterministic layer has since
  taken over is `delete` — the tool arrived after the rule and the rule never retired. The
  exception that stays: a line telling the agent NOT to duplicate what the hook already does
  (an autocorrecting pre-commit), which changes behaviour instead of restating it.
**Ground-truth checks — same audit pass. Each compares a CLAIM against the system that
decides it; all are deterministic.**

- **`CLAUDE.local.md` — READ it, never write it.** Its state claims (git mode, branch model,
  package manager, stack, ports, deploy target) are verified against disk and git exactly
  like any other file's, and outrank nothing: it is the highest-precedence local file, so a
  stale claim there wins over every correct one. Findings are report-only in the manifest;
  `apply` still never touches it.
- **Executable instruction layer** (`.claude/skills/**`, `.claude/rules/**`, per-repo and
  per-workspace): every disk path a skill routes to resolves; the tracker, branch and
  commands it names match the repo's own declaration. **Invocation gate is judged by
  CONSEQUENCE, not by name** — a skill whose body deploys, promotes, pushes, merges or
  deletes carries `disable-model-invocation: true`. A consequential skill without it is P0;
  so is a skill routing to a path that no longer exists.
- **Repo inventory**: `find -maxdepth 2 -name .git` against the repo names the instruction
  files actually mention. An active repo mentioned nowhere is a finding — NO inventory at
  all is the bigger one, because it makes the check unrunnable; propose the inventory block.
- **Declared vs git**: base and production branch against `git ls-remote --heads`; merge
  strategy against parent counts on the production branch (`git log --format=%p` — one
  parent = squash/rebase, two = merge commit). A declared branch that does not exist is P0.
- **Retired command names**: `rg` the workspace for `/`-prefixed command names no longer
  installed under `~/.claude/skills` and `~/.agents/skills`. Retiring a command elsewhere
  never rewrites the files that offer it.
- **Ledger ground truth** (`_support/PROJECT.md`): `Last updated` against the newest child
  HEAD; every pointer in the CURRENT handoff resolves; rows marked `current` still exist on
  disk; and the file does not contradict itself across sections. A false pending in the
  active handoff is the most expensive stale record in the workspace — it is the first thing
  the next session reads.

- **Tracker declaration**: a flow project declares in its LEDGER (the three fields —
  `memory-routing.md` owns the home and the confirm gate); the audit only FLAGS absence
  and proposes the ledger edit — never writes it, and the proposal rides the same
  confirm-gated batch.

Delegate the per-project reading to subagents (context hygiene; the semantic comparison
is judgment work — paraphrases count as duplicates). Output: a per-file table
(rule · current home · coverage CC/Grok/Codex/opencode · outcome · proposed diff), the
completeness proposals per repo, plus the hub-destined promotion/injection proposals.
Audit changes nothing.

**The manifest is Markdown, never HTML** — it is the `apply` step's input, not a document to
read, so it takes `communication-format.md`'s agent-handoff carve-out. This declaration is
what discharges the flow-report trigger; without it a long manifest keeps tripping it.

**Resolve the harness set from disk, never by asking.** Which harnesses are in play is
observable (`~/.claude`, `~/.codex`, `~/.config/opencode`, `~/.grok` — the last one's
`sessions/` and `memtrace/` show real use). And before asking whether a harness would lose a
rule, check how it reaches it: a **situational/path-scoped** rule reaches Codex AND Grok the
same way, through a router skill's injected references under `~/.agents/skills` — Grok is not
a special case, and it additionally gets every always-on rule as a flat symlink, which Codex
does not. **Verify Grok with `grok inspect`, which lists both the instruction files and the
skills it actually loads — never by looking at `~/.grok/skills/`:** Grok scans
`~/.agents/skills` and `~/.claude/skills`, so that directory is near-empty by design and
reading it as "no skills reach Grok" is a false negative. Ask only what disk cannot answer.

**A capability a connected MCP provides is NOT deducible from the rule layers — ask the
harness.** A tool's own description reaches the model directly, so a rule teaching its
mechanics can be fully redundant while every coverage matrix says otherwise. One-shot probes
settle it in seconds: `codex exec "<q>"`, `grok -p "<q>"`, `opencode run "<q>"`. Run one before
proposing to inject a tool rule into a router or to keep a per-repo line about a tool —
measured 2026-08-15, all three named context7 and its exact call sequence unprompted.

## `apply` — execute the confirmed audit manifest

One approval covers the batch; contradictions and overrides are listed individually
inside it. Project edits ride each repo's session git mode. A hook-config edit proposed by
**verification parity** rides that same approval — it is the one non-instruction file this
skill writes; verify it with the manager's own runner (`lefthook run pre-push`) before
reporting it done. Hub-destined proposals
(promote-to-core, injection-map edits) are OUT of apply's scope — the manifest carries
them for the user to execute in a session opened in the hub, where `build.py` and its
validation run. **Read the manifest from disk** (`> audit` writes it), so a run days later needs no
re-audit. **Revalidate before writing, entry by entry:** every path still resolves and every
quoted line still matches. A tree that moved underneath — a migration, an earlier partial
apply, someone else's commit — invalidates the entries it touched; those are re-audited, never
applied blind, and the report says which ones went stale. No manifest on disk and none in this
session → run `audit` first; nothing left to apply → say so rather than re-deriving one.

## Out of scope

- Nested `CLAUDE.md`/`AGENTS.md` in subdirectories — convert one root per invocation
  (point the skill at the subdirectory if needed).
- **Writing `CLAUDE.local.md`** — personal file, never written by any subcommand. `audit`
  READS it (above): excluding it from reading is what let a false git mode, a false merge
  strategy and a dead stack survive in three separate workspaces.
- Content quality in `convert`: it relocates, it does not rewrite or prune — pruning and
  dedup are `audit`/`apply`'s job.
- **Rewriting the user's prose, in any subcommand.** The skill proposes deleting, re-anchoring,
  or adding declarative blocks; it never recomposes, compresses, or restyles wording that stays.
- **Rationale as an inline comment.** Instruction files are tokenized verbatim by Codex,
  opencode and Grok (Claude Code strips HTML comments from `CLAUDE.md`; nothing strips them from
  `AGENTS.md`, which is where the canonical content lives). A rule's reason belongs in the
  ticket it cites or the commit body — never in the loaded file.
