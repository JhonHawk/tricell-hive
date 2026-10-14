# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
the project uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Each release sets [VERSION](VERSION) to its version; between releases it reads `dev`.
[CONTRIBUTING.md](CONTRIBUTING.md#releases-and-version-numbers) says which changes
raise the minor or the patch version.

## [Unreleased]

### Added

- Grok user scope installs `~/.grok/hooks/hive-read-whole-references.json`, a
  `PreToolUse` hook on `read_file` that runs the hidden `hive grok-hook
  read-whole` subcommand to drop `offset` and `limit` when a session reads a
  Markdown file under the shared skills directory (at most 1000 lines and 64 KiB),
  so references are read whole. It fails open and never denies a call. State gains
  the `hook` record kind, which older managers reject; a plan built from a `go run`
  or temporary binary that would install it is refused.
- Optional OpenCode V2 context loader, delivered inactive at user scope as
  `hive/context-loader/{index.js,context-loader.json}`, outside auto-discovered
  `plugins/`, with no `package.json`. Explicit
  native activation supplies canonical references only for `hive-review-plan` and
  `hive-review-harness`; it does not classify prompts, change permissions or
  enforce compliance. The manager generates its script snapshots and metadata
  schema 2 from the validated release payload; requests read no reference files
  or metadata. Canonical references retain one editorial home. Refresh requires
  regeneration and plugin reload, not editing installed Markdown. The derived
  files use existing ownership, drift and recovery machinery; their SHA-256
  digests record provenance, not cryptographic authentication. Activate the
  installed `hive/context-loader` directory, not the source template. Remove the user-owned plugin entry
  before uninstalling or downgrading to content without it.
- Declared reading-footprint scenarios report unique source bytes and characters,
  including required procedures and selected activity entrypoints, separately
  from actual tokens. New offline packages use manifest v2 and include the loader
  source; the current manager still verifies v1 packages. Older v1-only managers
  cannot verify v2 packages. This optional integration meets the minor-version
  feature criterion; no release number is assigned here.

- The route question can offer an iterative route when the work alternates
  investigation and implementation toward one measurable objective. Its option
  states the objective, how it is measured, the number of rounds the agent
  proposes, and when it stops. Once you choose it, the agent reports each round
  in one line and keeps going within local editing and verification; it stops
  and asks when the objective is met, the rounds run out, a round brings no
  measurable improvement, a step needs an effect the route does not cover, a
  round needs a plan, the scope changes, or a decision is yours.

### Changed

- A research child cites a live observation by the exact read-only command that
  produced it, so the parent reruns it unchanged instead of rewriting the query, and
  children that share a question get disjoint scopes so none repeats another's reads.

- The browser tool follows the situation: `agent-browser` for journeys, UX review,
  screenshots and measurements, Chrome DevTools MCP for audits and performance traces
  or an independent reproduction, Playwright MCP only without `agent-browser`. A child
  never installs a missing browser tool; it reports the install commands and the user
  authorizes once. The frontend, UI design and task-verification roles now read the
  browser rules and use only the browser tool their brief assigns.

- `git-workflow` no longer prescribes a worktree location: agents use the host's own
  worktree feature or the repository's documented place, never a temporary directory,
  make sure an in-repository worktree path is ignored, and record each worktree they
  create in persistent memory so another session or host can resume it; `flow-close`
  updates that record when it removes the worktree. Agents create a worktree only when
  the current checkout holds others' changes or a non-base branch, another task writes
  in the repository at the same time, or the task needs two checkouts at once; otherwise
  they create the branch in the current checkout.

- OpenCode V2 role profiles now use OpenCode Go models instead of GitHub Copilot:
  `opencode-go/deepseek-v4.1-flash` for execution, `opencode-go/deepseek-v4-pro`
  for reasoning and `opencode-go/glm-5.3` for the verifier.

- The global core retains essential boundaries and explicit read conditions while
  canonical skill references hold reporting, questions, delegation, artifact
  details and project settings. Reporting remains required before every chat
  reply, including ordinary answers without citations. Smaller global text is
  not a claim of lower total context or better model reliability; the declared
  plan/build source union can grow.

- A research round about work you could start now ends with the route
  question directly: the routes with the flow each one runs, keep
  investigating, or stop at the findings. The agent no longer first offers to
  show recommendations. With a question card that waits for your answer, the
  message keeps the explanation and one summary line per question with the
  recommended option, and the options are in the card; without one, or after a
  dismissal or timeout, the options are written in the message, and a
  dismissed card is asked again in text. After each answer the agent says in
  one line whether it settles the question, opens the next round, or returns
  the work to an earlier flow, and what that invalidates. The size check on the
  global guidance now only reports its size and no longer fails.
- A ready plan is summarized in the conversation, not only saved to its file:
  the objective, the decisions in force, each task in one line, the checks,
  the delivery, what is still open, and the plan's path. The summary ends with
  the question whether to start Build now or stop with the plan ready.
- Findings from a review that runs inside work you already authorized are fixed
  as part of that work, without asking again; a review you request on its own
  still does not authorize fixes.
- Work that will not end in a Git delivery gets the review question once, if
  that choice is still open, before the first edit instead of a separate stop
  when the review is due.
- After the one follow-up review, a new finding whose fix is small, inside the
  change and touches no contract is fixed with a test, without another review
  round, and the report says it was fixed without review; anything larger comes
  to you as a decision.
- `hive-research` returns a bounded reply: the answer, a status with its stop
  reason, each deciding claim with a locator and excerpt, counterevidence and what
  it did not check. `flow-research` checks those claims by opening their locators,
  gives each research child a stop condition, routes long state collection to
  `hive-read-state`, and uses the global evidence labels; independent children
  launch together.

### Fixed

- UI review captures go to the work's `_support/workspace/<date>-<work>/images/`
  folder again: `flow-build`'s verification reference still pointed to a global
  rule that had moved to `artifact-layout`, and asked for a temporary destination,
  so a session saved its captures where a restart erased them.
- A side finding becomes a proposed ticket only when its failure can start from an
  input or state the running system can produce; one traced only inside a component
  stays in the report as unconfirmed. The close question marks which findings it
  recommends opening.
- The `automatic` delivery mode never stops for human validation. When the project's
  guidance requires one for a change, such as a visual check of a changed screen, the
  delivery question marks `automatic` unavailable for that change and names the rule,
  instead of offering it and stopping before the merge anyway.
- UI review and in-vivo rounds are shorter: the agent seeds the walk's data and
  passes its records to the children instead of having each one create them through
  the interface, the UI reviewer no longer repeats the functional journeys the in-vivo
  check owns, and each child stops after one observation per area, viewport and theme,
  reporting a check that hangs or fails twice as not verified.
- A stuck browser command no longer stalls a UI walk for minutes: each `agent-browser`
  command runs under a wrapper that ends it after about 45 seconds, a command that
  hits that limit is treated as a wedged session (doctor, close, fresh session), and
  options in custom selects are chosen by click rather than keyboard.
- Read-only roles on Grok can no longer write files with Grok's `write` tool: the
  observe deny list named only `Write`, which Grok ignores, so a review child wrote
  its notes file. The list now names the tool ID `write`.

## [0.3.3] - 2026-10-10

### Added

- An urgent production fix has its own route: it enters the base branch and is
  promoted at once through each environment, in order. When a promotion would
  carry work you are holding back, the agent offers a hotfix branch from
  production as a separate decision, only where the project's branch policy
  accepts it, and merges that same branch back instead of cherry-picking. A
  short postmortem is offered at close.

### Changed

- When the work needs its own checkout, the agent creates the Git worktree
  where the repository's guidance places worktrees, otherwise in the sibling
  folder `../<repo>-worktrees/<work>`, never inside the repository or in a
  temporary directory. It names each worktree in its report and removes it
  with its branch at close. Worktrees made by a host's own feature keep the
  host's location.

### Fixed

- Environment branches after the base branch, such as `qa` and `production`,
  again receive changes only through the promotion path: no direct commit,
  push, cherry-pick, or hotfix, and every deviation is asked per push, even
  under unattended delegation.

## [0.3.2] - 2026-10-09

### Added

- The documentation site includes a new architectural Hive social preview image
  with Open Graph and X image metadata on every page, plus a branded homepage
  sharing title.

### Changed

- A message that waits on the user ends with one fixed pause line, such as
  "**Esperando tu decisión:**", instead of burying the choice in a report or the
  pending list. Research ends each round of findings by offering to show the
  recommendations, and a plan decision that changes the design is presented at
  once under a pause label.
- A work item is a ticket or the scope you authorized, not each task of a
  plan: between tasks the agent reports progress and continues. When work
  continues a plan in progress, the closing question's continue option covers
  all remaining tasks and asks once for the permissions they still need, and
  the agent no longer suggests a new session in the middle of that plan.
- While authorized work remains, the agent sends each progress line together
  with its next action, so the turn does not end on an announcement, and it
  answers a question or notification in one line and keeps working. Each build,
  deployment, or CI run it starts gets a completion signal before its turn ends,
  preferably the host's own monitor or watch tool.
- When work resumes or continues a plan, for example after a close, the agent
  lists the remaining tasks in the host's task list again.
- A plan's change record quotes the user's request verbatim, and the
  acceptance criteria of each ticket it names with the date they were read, so
  requirements are not lost to paraphrase on the way to the implementer.
- An implementer of a plan task reads the task, the quoted request, and the
  acceptance criteria it closes from the change folder; the delegation brief
  points to them instead of paraphrasing them.
- When the delivery mode makes commits, the agent commits each plan task once
  it is verified, so every commit holds one verified task and its fix rounds.

### Fixed

- Every choice the agent asks for, including decisions, routes, the close
  question, and yes-or-no offers, is written complete in the message: context,
  options with their effect, and the recommended one with its reason. The native
  question card of each host is used as a shortcut and written to stand alone,
  so the decision survives when the card or the message is hidden.

## [0.3.1] - 2026-10-08

### Changed

- When diagnosing a failure that can still occur, the agent first runs a command
  or observation that shows the exact symptom, and only then forms hypotheses or
  edits code. When it cannot build one, it says what it tried and asks for a
  reproducing environment, a captured artifact, or temporary instrumentation.
  Incident mitigation does not wait for it.

### Fixed

- Work folders under `_support/` carry the change id with words that identify
  the work, shortened when long, never the ticket key alone; without a change
  they use the ticket key followed by a short description.

## [0.3.0] - 2026-10-07

### Changed

- Unattended delegation also starts when you say you are leaving, such as going
  to sleep, while authorized work is running. It lasts until that work completes
  or you return, decides open questions itself and reports each decision with how
  to revert it. Irreversible or destructive actions stay forbidden.
- Before handing a pending item to you, the agent checks its premise and whether
  the session can do it itself. Closing a work item also restores local state the
  task changed.
- A restriction enters project guidance, a plan, or a question option only when a
  source states it or you decided it.
- Questions asked in text give each option one line; shared effects and
  unavailable routes are stated once after the options.
- The README and `llms.txt` link the documentation site, and the site header
  links the GitHub repository.

## [0.2.1] - 2026-10-06

### Added

- Documentation site in `site/`, with one page per flow (research, plan, build,
  close), published at https://hive.tricell.tech/ from this release on.

### Changed

- The `starlight-docs-site` template pins Astro 7.3.5, Starlight 0.42.4, and
  pnpm 12.8.1, denies the `esbuild` install script it does not need, and ships a
  `biome.json` so `pnpm lint` ignores build output.
- `README.md`, `llms.txt`, and pages of the documentation site are updated in
  the same pull request as any change to what they state.
- `CONTRIBUTING.md` defines what the version number tracks and when a release
  raises the minor or the patch version.

## [0.2.0] - 2026-10-05

### Added

- One-line installation: `curl -fsSL https://hive.tricell.tech/install.sh | sh`
  runs `get-hive.sh`, which downloads the latest release package, verifies its
  checksum, keeps it in `~/.local/share/hive/packages/`, and runs its
  `install.sh`. `--dry-run` previews the installation without changing any host
  (the package is still downloaded and kept; add `--hosts` when no terminal is
  available), and `--version` selects a release (#11).

### Changed

- `git-workflow` asks pull-request descriptions to state whether a revert fully
  undoes the merge and what the change would affect if it were wrong.

## [0.1.0] - 2026-10-05

First public release.

### Added

- Shared guidance layer for deliberate work with coding agents: authorization,
  preservation, evidence, and artifact-placement rules.
- Activity skills: `flow-research`, `flow-plan`, `flow-build`, and `flow-close`.
- Specialist role definitions with bounded tasks, owned paths, and evidence to
  return.
- Go manager with the commands `install`, `status`, `doctor`, `update`, `plan`,
  `apply`, and `recover`. It previews destinations, checks ownership and drift,
  preserves user content outside managed resources, and supports recovery.
- Six host adapters: Claude Code, Codex, Grok Build, Pi, OpenCode V2, and
  Cursor CLI.
- Downloadable packages on GitHub Releases for macOS Apple Silicon (arm64) and
  Linux arm64 and amd64. Each package includes the executable, content,
  verification inventory, `LICENSE`, and `THIRD_PARTY_NOTICES.md`, and installs
  offline with `./install.sh`. Installation refuses a package that lacks the
  license files, and a content test keeps the notices in sync with `go.mod` and
  the Go toolchain.

### Removed

- The `bootstrap.sh` online installer and the `hive bootstrap` subcommand.
  Install from a complete package or a source checkout instead. State written by
  the retired subcommand (an `installer` field) is still read and ignored.
