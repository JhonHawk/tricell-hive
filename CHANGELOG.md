# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
the project uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Each release sets [VERSION](VERSION) to its version; between releases it reads `dev`.
[CONTRIBUTING.md](CONTRIBUTING.md#releases-and-version-numbers) says which changes
raise the minor or the patch version.

## [Unreleased]

### Added

- `flow-close` takes an inventory of the support material its work left, such
  as research records, working files, evidence or an abandoned change folder,
  proposes an action for each item (commit, archive, promote a conclusion into
  `docs/`, delete or keep) and asks it right after the viewer close, before
  archiving and cleanup, applying nothing before the answer. Committed items
  go with the archived change record, or in a support-only commit on a new
  branch, delivered the way the work was; kept items in a worktree move to the
  main checkout before it is removed; paused and held work only list them.
- What a work item leaves open on purpose, such as a post-merge task of its
  plan or a later phase, is commented on its ticket and kept in one
  `pending/<change-id>` persistent-memory record titled `Pending:`;
  `flow-status` lists those records, marking those ready to resume, along with change
  folders not committed under `Specs`, and each close soft-deletes a record
  left with nothing open.

### Changed

- A change folder is committed with the code it plans: on the work's branch
  before the push, moved into the worktree when there is one, or in a
  separate specs repository at the completion point, a `ready` stop
  included. It no longer stays untracked until `flow-close` archives it.
- Every flow delegates by one criterion, kept in
  `flow-build/references/delegation.md`: the reviews Hive launches always
  run in a child, briefed with the artifact, the requirements and check
  outputs but never the producer's reasoning; other work goes to a child
  only when it is large (more than five large reads, or in a build more than
  three edited files). Read-only children run side by side; children that
  write do so only against a frozen typed contract with an integration check
  after. `global.md` defines a large read and one list of risky changes
  (#311).
- `hive-review-plan` reviews every plan with a change folder, and the UX
  reviewer also covers how the changed regions affect the layouts and
  components around them (#311).
- After a deployment, the served-build check runs in the main thread unless
  it is large (#311).
- A card that closed after a timeout still counts as a waiting question tool,
  so the next round goes through a card again instead of only in text (#311).
- A maintainer viewer restart reports that open viewer tabs lost access, since
  the restart issues a new token, and offers the reply *open the viewer*,
  which runs `hive view open`.
- The offer to open the viewer never goes as a question of its own: it joins
  the round its stop already asks, or a reply line of the message's waiting
  line or next-step block, with no decision ID or question card of its own
  (#311).
- Build hands its check evidence to Verify, which reuses passing results only
  while the inputs they cover remain unchanged and reruns missing or invalidated
  checks. The real-path walk and required independent verification remain.
  Code findings use an available, authorized local reproduction or conclusive
  source evidence, including when adjudicating a native review's output.
- Every request to do work now enters `flow-research`, which settles what the
  user needs before any how: it reads before asking, asks one open question
  per message in the text, checks a complete request for blockers and does it
  or continues into `flow-build` in the same turn, and sends to `flow-plan`
  only when the how has a decision for the user or the work cannot be finished
  from the request and the `git diff`. An analysis ends at its findings with
  no flow recommended. `flow-research` no longer publishes to the viewer,
  which starts its session at plan or build with a one-paragraph summary;
  `flow-plan` presents two or three approaches before the tasks; and after
  integration `flow-close` moves the ticket, reports and ends with a footer
  recommending to close plus replies from the context, running the archive and
  cleanup on the user's reply, when other work starts or when the session ends.
  A polite request phrased as a question that names the action and its target,
  such as "can you add this user to the repository?", counts as a request for
  that change; "can we…?" and "how would we…?" stay questions (#311).
- Archiving a viewer work shows a temporary toast naming that work with Undo,
  replacing the persistent banner that could appear above another active work.
- `flow-close` never stops the viewer service or closes the viewer the user
  has open: closing a work's viewer session changes only its record. No agent
  quits, restarts or kills a program other sessions or projects may use, such
  as Docker Desktop or a local database server, even when it hangs: it reports
  a blocker for the user to restart it. Starting or stopping a container the
  task did not start needs the user's confirmation, and the processes a task
  started are stopped by their process ID, never by a name or pattern that can
  match another session's processes.
- `flow-research` writes the aspects of a question, their sources and which go
  to a child in its message before the first read, counts its first reads
  toward the delegation limit, delegates each aspect that needs more than five
  large reads (a whole document or more than 100 lines; filtered queries stay
  in the main thread), does not reread what a child covers, and hands over the
  rest of an aspect when its sixth large read in the main thread still has no
  answer (#311).
- Viewer navigation groups registered source roots with the same explicit
  `.engram/config.json` project name, keeping each work's original source visible
  and its identity, history and collaboration permissions unchanged. Configured
  non-Git workspaces can be registered directly.
- GitHub now enforces that `master` receives changes only by promotion from
  `development`: a ruleset blocks direct pushes, force pushes and deletion, and
  requires the `source-branch` check, which fails for a pull request from any
  other branch. Only the repository admin can bypass it, by merging a pull
  request.

## [0.8.0] - 2026-11-03

### Added

- A `test-audit` skill trims one area's tests to one owner per contract. It
  records a baseline (test and production lines, pass or fail, coverage when
  measured, suite time), marks every test retain, repair, consolidate or
  delete in a read-only ledger, adds the keeper tests before deleting what
  they absorb, and has an independent reviewer and one deliberate code change
  per affected contract show that nothing lost its only test. Product defects
  it finds go to you as a decision. Adapted from OpenClaw's `test-audit`
  (MIT).
- Lists the user reads, such as reports, closes and pending items, open each
  point with what it is in plain words and carry one task per point; an ID it
  cites goes last, in parentheses, such as "**Move the plan summary to the
  top** (D2)". Decision titles and options keep their format.
- The UX reviewer runs an automated accessibility audit on each screen and
  state it reviews, including states reached by interaction, with
  `agent-browser a11y`: WCAG A/AA violations count as defects, best-practice
  rules as P3 findings, and results axe cannot decide as pending manual
  review. A UI linter the project names also runs when it is already
  installed; its style rules are informational only.
- `hive view contract --for <flow>` prints a brief contract with only that
  flow's operations (research, plan, build, verify, delivery, close, collab, local):
  usage, required and optional fields, allowed values and an example.
  `--list` prints one line per operation with its flows, and
  `--section rules` the shared rules. A rejected input names its field, such
  as `invalid timeline input: invalid_text on reason` (#328).
- A `ready` delivery mode: the agent reviews, pushes, opens a non-draft pull
  request, waits for CI and answers its reviews, then stops for you to merge.
  It never merges, enables auto-merge or joins a merge queue, and asks first
  when the repository's automation would merge the pull request. The open
  pull request is the work's completion point, and the viewer shows it as in
  review (#312).
- After the push, review bots are read once when CI ends: only their output
  for the latest commit counts, each finding is checked against the code,
  dismissed ones get a reply with the reason, and their fixes get a single
  round. A bot that has not reviewed yet does not hold the merge. When GitHub's
  GraphQL quota runs out, the pull request and its checks are read through
  REST and CI is awaited with `gh run watch` (#312).

### Fixed

- Before asking you to run a command or take a step, the agent checks whether
  its access and the authorization it already has let it do it, including the
  command's own non-interactive options, and does it when they do. You are asked
  only for what it cannot do, such as a browser sign-in, and for authorizations
  as decisions; it never gives a confirmation the guidance requires from you.
  The skills rebuild had dropped this check, so an agent asked you to run
  `hive update` when it could have run it itself.
- `flow-close` closes the work's viewer session as its first step, before any
  cleanup, and reads it back: closed when the work was integrated or
  abandoned, paused when it was paused. It was a note above the steps and a
  close could skip it.
- UI verification and the UX reviewer prefer `agent-browser` again when it is
  installed, use Chrome DevTools MCP only for audits, traces and an independent
  reproduction, bound each browser command to 45 seconds, and close the
  sessions they open. The skill cut had replaced that rule with "the available
  browser automation capability", so a session used another browser tool.

### Changed

- Build has a Tests section again, narrower than the test-first rule the
  skills rebuild removed. A test must fail first only for a fixed defect, as
  one regression at the layer that owns it, and for permissions, money, data
  integrity or a public or persisted contract; other changes get a direct
  check or extend the test that already owns the behavior. Each new test
  needs a realistic change that would make it fail, one owner per contract at
  the strongest boundary the harness runs cheaply, test doubles only for
  external services, time, randomness and a client's view of the project's
  own server API, and no fixed sleeps; a failure, limit, ordering or rollback
  the change handles is covered too. In a project
  without a test runner, a bug fix or consequential change first pins the current
  behavior for every caller, and adding test infrastructure is asked once.
  The roles that write tests follow it.
- Compacting is offered at two moments instead of by context size: when an
  approved plan moves to its build in the same session, and when another work
  item starts after a close. It is one more reply of the next-step block, with
  what to keep; after you compact, the agent shows the same next step again.
  The old rule recommended compacting past 40% of the window, but Claude Code
  does not tell the model its window, so the agent estimated against a smaller
  one and recommended it at about 20% of a 1M window (#328).
- A flow's command is offered only when the next step needs that flow. A single
  action that changes no repository file, such as creating the ticket a route
  needs, closing a ticket as not planned or deleting a merged branch, is asked
  directly ("Do you want me to create the ticket?") instead of being sent
  through a flow command, and your answer authorizes only that action (#328).
- The viewer publication procedure is read once per session and again after a
  compaction, not at each stage change. Each flow reads its own part of the
  contract once, uses the versions every write returns instead of reading the
  session after each one, and reads the session back before saying something
  is visible, at close, and when resuming. Publishing took 5–16% of the tokens
  of a session (#328).
- `hive view` prints compact JSON, with the same content (#328).

## [0.7.0] - 2026-11-01

### Added

- A shared-work viewer with an ascending timeline, declared research/plan/build/
  verification/delivery stages, independent task segments, visible review
  evidence and a compact Tasks/Go-to-current toolbar. One common detail panel
  supports immutable Markdown revisions, galleries, task review and CI attempts
  with manual snapshot updates.
- Native producers join the same work using their own verified Source and
  accepted producer IDs. Canonical ticket/delivery registries and immutable
  candidates replace repeated inline entities; one metadata patch updates
  current references while preserving publication context and history.
  Conditional operations retain CAS, audit, durable receipts and recovery.
  `hive view contract` describes commands, schemas and limits offline.
- Protocol/schema/storage3 starts on a fresh state directory and rejects earlier
  formats before writes. No legacy scanning, migration or preference import is
  performed. Standalone HTML fallback remains separate. Retention is manual,
  deprecated --ttl is ignored, and the viewer defaults to 4483 with a free-port
  fallback. Browser rename/archive/pin/delete manage the work without changing
  source files or native host history.
- Viewer system labels follow the primary browser language: English default/
  fallback and Spanish for es/es-* locales, including accessible names and
  dates. Stage keys render one localized heading with optional verbatim context;
  producer titles, captions and document history are unchanged. The offline
  catalog exposes the reserved labels and supported languages.
- A [viewer manual](site/src/content/docs/viewer.md) covers published work,
  stage-specific tasks, evidence, current/published references and delivery CI.
  The canonical publisher guide documents shared producers and ID-based
  publication without duplicating task state or provider observations.
- A new `flow-status` skill reports the project's status and backlog, read
  only: open tickets grouped by product module, unfinished local work (branch,
  uncommitted changes, unmerged branches, worktrees, open pull requests), and
  numbered recommendations. A ticket another session has in progress is listed
  apart and never recommended. Asking "what's next?" ends with the ticket to
  assess and the command that starts `flow-research` on it.
- A repository test checks that every heading an instruction link in
  `content/` points to exists, with the same link grammar and fence rule as
  release validation, so renaming a heading can no longer break a link
  without a failing test.
- `harness-audit` compares the memory each host loads at every entry point,
  not only the instruction files: Claude Code, Grok Build and Codex key their
  own memory by where the session starts, so a project whose memory lives at
  the workspace behaves differently from one started in a repository. It
  proposes a home for each fact that only one entry point loads, flags an
  Engram project that sessions split across two names, finds memory that
  restates the guidance in use, and flags an entry point where Grok Build
  loads no project instructions because that folder has no trust grant. It
  now also asks the decisions its findings need and applies the accepted
  outcomes, including creating or updating a repository's `## Hive` section
  and checking it with `hive doctor --project`.
- Unattended work that repeats cycles has a loop contract: a measured
  objective, a cap, a check every cycle, a stop after two flat cycles, and a
  state file a new session resumes from.
- Before each push of a change that is not mechanical, the agent runs locally
  the checks CI would run on that push: the jobs from the workflow files that
  are deterministic and need no service, secret or deployment, for the
  packages the change touches and those that depend on them. Jobs that need a
  service, a secret or a deployment stay in CI and the report names them as
  CI-only.

### Changed

- The skills are rebuilt as a small core. `content/guidance/global.md` shrinks
  from about 39 KB to about 20 KB and holds the rules every session needs,
  each stated once, in one place: language, scope and authorization,
  preservation, secrets, proportionality, questions and routes, delivery,
  project settings, memory and scope changes. The flow skills (`flow-research`,
  `flow-plan`, `flow-build`, `flow-verify`, `flow-deploy`, `flow-close`,
  `flow-walkthrough`) keep their main path in `SKILL.md` and lose most of
  their shared references. Rules a skill needs stay with that skill.
- Questions are asked one way everywhere. Each decision has an ID (`D1`, `D2`)
  and a title that holds the question; findings are `H1`, plan tasks `T1` and
  acceptance criteria `AC1`. Each option takes one line, with the recommended
  one first and marked with a star. The agent writes the round as text, then
  asks it in one native question card, and replying "ok" applies every
  starred option while `D2 B` changes one. When a card closes without an
  answer, one line says so and the answer goes in the chat. A message that
  asks, reports a result or a blocker, or closes work opens with a context
  line (stage, ticket, progress) and ends by saying what it waits for. A
  blocker comes with the smallest fix and, when it needs your choice, that fix
  as the starred option.
- The delivery mode and the code review mechanism are asked once per work
  item, before the first edit, from the global guidance, in one round. The
  review question has fixed options: the host's native review, the
  `hive-review-code` agent, or no dedicated review, with the recommendation
  following what the change touches. A repository that declares
  `Delivery: direct-base` is not asked. Before the validation or the push, the
  agent compares the branch with its base and, when it carries commits that
  are not this work's, names them with their count and asks whether they go in
  the same publication.
- Moving from one flow to the next is no longer a question card. A report ends
  with a separate next-step block: the recommended flow, its reason, and the
  command to type, written once in the syntax of the host you run in (for
  example `/flow-plan COM-833` in Claude Code, `$flow-plan COM-833` in
  Codex), with a real ticket or a recommendation to create it first. Only your
  reply starts the next flow.
- Every child (delegated role) starts with a fresh context window, not a copy
  of the conversation, and its brief carries what it needs. Children never
  commit, push, open pull requests, write trackers or launch children; the
  main thread commits and checks the evidence they return. Findings from every
  review use one scale: P0 blocks a release or a core flow, P1 must be fixed
  before merge, P2 should be fixed but does not block, P3 is minor.
- After a context compaction the agent resumes from the work's checkpoint
  (the change's `tasks.md`, which records the stage and step, the last commit,
  the open findings and the next action), does not repeat a report the history
  shows already done, and skips the hook-suggested session-summary save.
- `git-workflow` is cut to a minimal core that analyzes the situation first:
  it fetches, names commits that are not this work's, and, when conventions
  and state do not settle a Git choice (which branch, a worktree, how to split
  commits, a diverged branch, a conflict, a rejected push), asks one decision
  with concrete options instead of picking silently. It stages named paths
  only, lets hooks run, checks each effect after it happens, and merges only
  on a current base.
- `flow-plan` writes the change folder from templates (proposal, tasks,
  design) with a prerequisites section, numbered acceptance criteria that must
  be false on the current base, and, for a screen change, UI decisions settled
  against the real design system with ASCII sketches. It ends with a plan
  summary and the build route, and offers compacting first when the context is
  past 40%.
- `flow-build` keeps a native task list and a checkpoint, fixes a failing
  check by first deciding whether the code, the criterion or the test is
  wrong (a test is changed only when the test is wrong), repairs the local
  environment without asking, shows the first rendered view of a new screen
  for acceptance, and lets a delegated child work only where its writes do not
  overlap other work. `flow-verify` validates while the reviews run, in one
  walk and one fix round: the user's corrections and each confirmed blocking
  finding are fixed, a CI job that fails is diagnosed from its log before any
  rerun, and the second review covers only the fixes.
- `flow-close` ends with one final report (integrated, resolved, omitted,
  repaired, findings, next), proposes findings as a table answered by ID with
  P2 findings of one cause or area grouped, and asks one close question: end
  the session, see the findings, or find the next ticket.
- `flow-research` lists every plausible cause of a failure and checks each one
  before concluding, reporting the ones it ruled out with their evidence. Code
  review checks that a diff is minimal and rates an unrelated refactor or
  added feature as P2. Verification names what is not evidence: a sign-in
  page, an error or an empty list on a changed screen, a claim that checks
  passed without their output, a capture taken before the last change, and a
  child's report the main thread did not check.
- The documentation and UX review children use a lighter model profile. On
  Codex that is GPT-6.1 Terra for state collection, documentation and UX
  review, and routine task verification uses GPT-6.1 Sol high instead of GPT-6
  Astra high; other hosts keep their role models. Roles on Pi and OpenCode run
  on the session's model by default (OpenCode roles also take its effort)
  instead of models that need subscriptions that are often inactive; use
  `hive models set` to choose a model per role. Rebuild the manager before
  `hive update` to support the new profile.
- Cursor CLI gets the Hive guidance as an always-applied rule at
  `~/.cursor/rules/hive.mdc`, because it never loaded `~/.cursor/AGENTS.md`.
  The rule reaches workspaces under the home folder only. An update retires
  the block an earlier release wrote into `~/.cursor/AGENTS.md`.
- Unattended delegation no longer ends when you add a resource, feedback, a
  correction or a question about the work within its goal; changing the goal,
  the scope or an earlier answer still ends it. When you delegate or say you
  are leaving before the delivery, review and authorization questions were
  asked, the agent asks them at once in text and keeps working, stopping at a
  verified local result without your answers.
- Walkthroughs use screenshots of relevant states, including open dialogs and
  menus, instead of recording flows. Existing video manifests remain
  compatible. UI verification checks only the widths, themes and neighboring
  regions the change affects.
- The `## Hive` section of `AGENTS.md` now holds facts only: `Project`,
  `Base branch`, `Tracker`, `Specs`, `Environments`, `Delivery: direct-base`,
  `Completion` and `Hive guidance`. The delivery mode and the review mechanism
  are asked every work item; `Review`, `UI verification` and
  `Delivery: automatic|interactive` no longer answer or authorize anything. A
  later completion point is the new `Completion` key, written
  `<state> on <ref> · <state> on <ref>`.
- `hive doctor` warns about retired keys and a malformed `Completion`, and exits
  non-zero, printing `hive: Project section has blocking findings`, when the
  section has a blocking finding. `hive project set` and the Project view no
  longer offer `Review` or `UI verification`.
- The manager keeps only the newest 10 finished recovery journals under its
  `transactions/` folder, and the install summary says so.

- `flow-build` waits for helper agents with the host's wait tool, which returns
  when a helper finishes or writes, never with a loop of sleeps and status
  checks (observed in a Codex 0.166 session: 74 sleep calls, no wait).
- A plan's `proposal.md` records each user decision on its own line, in plain
  words: the question, the answer and what it changes, with the decision ID
  last. Field names follow the session language; the first field stays
  `Status` (or `Estado`), which the archive helper reads.
- The documentation site describes the rebuilt flows in plain words and links
  each skill; new pages cover Status, Security review and Host limits.

### Removed

- The `flow-security` and `workspace-archive` skills, and the
  `hive-refute-claim` role. A security review is the `hive-review-security`
  role, run only when you ask; a check of whether an advisory affects your
  repositories is `security-advisory`. Archiving closed `sessions/` folders is
  part of `workspace-conventions`. Checking a claim is part of `hive-research`
  and of the main thread's own check of what a child returns.
- The "matching declared job" source of unattended delegation, which no rule
  defined: a scheduled or background turn activates no unattended delegation.

### Fixed

- `hive update --out`, `hive plan --out`, `hive voice --out`, `hive models --out`
  and `hive apply --plan` accept a plan path whose ancestor is a symlink, such
  as macOS `/var` for `mktemp -d`. The file is written and read at the
  resolved path. A symlink that remains after resolution is still rejected.
- The walkthrough viewer keeps a project's resolved identity when a worktree
  disappears or its remote cannot be read, including after a restart, instead
  of splitting its saved copies into another project with the same name.
- The viewer timeline places work-level activity, such as a pause, where it
  happened between stages instead of above every stage, and shows a server
  reason once, in a title that wraps, instead of three times.

### Migration

- Run `hive update` and restart open sessions, since they still carry the
  previous guidance and skills. Rebuild the manager first
  (`go build -o "$(command -v hive)" ./tooling/cli` from a checkout) when you
  update from a checkout, because an older binary can reject newer content.
- Stop invoking `flow-security` and `workspace-archive`: ask for a security
  review (`hive-review-security`) or an archive of closed sessions in plain
  words, or use `security-advisory` and `workspace-conventions`. Stop
  invoking `hive-refute-claim`.
- The viewer moves to protocol 3, which does not read earlier viewer data. Stop
  the running viewer (`hive view stop`), move the `viewer` folder out of the
  Hive state directory (on macOS `~/Library/Application Support/tricell-hive`)
  to keep it, rebuild the binary, start the viewer again and register each
  repository with `hive view register --dir <root>`. Earlier works stay in the
  moved folder and do not appear in the new viewer; restoring that folder and
  the previous binary is the way back.
- Remove `Review`, `UI verification` and `Delivery: automatic|interactive` from
  each `## Hive` section with `hive project set --unset <Key>`, or run
  `harness-audit`; sections that keep them still work, and `hive doctor` only
  warns.
- Move a later completion point written in prose into `Completion`.
- Scripts that run `hive doctor` must expect a non-zero exit when the section
  has a blocking finding, and, with `--project` given explicitly, when the
  repository has no section or the path cannot be checked.

## [0.6.0] - 2026-10-25

### Security

- The documentation site and the `starlight-docs-site` templates move to
  Astro 7.4.0, Starlight 0.42.6 and Biome 2.5.15, and enforce a 7-day release
  cooldown. The site also overrides patched undici, source-map-js and zod and
  drops an unused dependency that brought extract-zip, so its `pnpm audit`
  goes from 11 high advisories to none.

### Added

- Stage skills `flow-verify`, `flow-deploy` and `flow-security`, with an
  "Entry and exit" contract in every stage skill. `flow-verify` checks a built
  candidate at a level the project settings and the request set, and a check
  skipped at your request counts as accepted without verification; `flow-security` runs the
  security review only when you ask; `flow-deploy` is a separate stage after
  integration that promotes through the declared environments and checks what
  is served. The README, `llms.txt` and the documentation site describe them
  (#169).
- Open topics: the agent keeps a list of the topics you add while work is open
  or closing, tells you in one line how it classifies each, closes none
  without a destination, and names them in its reports and route or close
  questions (#160).
- Test guidance: `tdd.md` now sets which kind of test to write and a filter
  before writing one (the test that already owns the behavior, low-value
  patterns, no test-only entry points in production code), recommends
  property-based tests where an independent oracle exists, and verification
  compares the old and the new path on the same inputs when a change replaces
  or migrates something (#157).
- Options of a question carry their estimated size when one starts work on
  another ticket, adds a component or dependency, or is about ten times
  another, and plans that edit agent guidance get a guidance-consistency
  reviewer. This repository declares `Delivery: automatic` and
  `Review: /code-review` (#158).
- Codex wait floor: an interactive `hive install` or `hive update` with the
  Codex host offers, selected by default while it is not applied and Codex is
  0.164 or later, to write `min_wait_timeout_ms = 300000` and
  `default_wait_timeout_ms = 600000` under `[features.multi_agent_v2]` in
  Codex's `config.toml`, so a parent agent waiting for a delegated child does
  not wake every 60 seconds. Hive saves `config.toml.hive-backup` (0600) before
  the first write and its journal keeps only fingerprints and the lines it
  added; `hive plan remove` for the codex host removes only the unchanged lines
  Hive wrote, then the backup, after a clean revert; a file Hive cannot edit safely becomes a manual
  step; `hive doctor` adds a row for it. This replaces the rule
  that Hive edits no host configuration file itself. The minimum applies to
  Codex's v2 multi-agent runtime (v1 unverified), and an older Hive binary that
  rewrites Hive's state loses the record, so `plan remove` would leave the
  lines. A default-selected capability raises the minor version.
- Check a dependency before installing: before adding, updating or installing
  a package, including a one-off `npx` run, the agent confirms the package,
  picks a version the release cooldown accepts, looks up advisories for the
  exact version without sending private package names to third parties, scans
  the lockfile, and asks before installing when a critical or high advisory
  has no safe path.
- The `security-advisory` skill checks, read-only, which repositories a
  published advisory or package compromise affects, with evidence for each
  one, and leaves the fixes to each project's own session.
- The viewer combines walkthroughs with the same project, host and session ID
  into one sidebar entry, with chronological ticket segments, internal jumps
  and independent copy actions. The session header collects explicit PR or
  commit delivery links while keeping each segment's association.
- CodeGraph is an optional, manual capability: the installer offers it for
  Claude Code, Codex, Cursor and OpenCode with its official source, and
  `hive doctor` shows whether `codegraph` is on `PATH` without running it. The
  global guidance says when to use a CodeGraph index, when to run
  `codegraph sync -q` so its relations are current, and which results to
  confirm with text search (#104).
- Plan review adds a User experience area for plans that create or change a
  screen, dialog or form where the end user enters or chooses data. Its
  reviewer walks the end user's task looking for values the system already
  holds or entries repeated per record, and UI planning records where each
  value a view asks for comes from.
- Before a production deployment of code or infrastructure, the build records
  its rollback triggers: each signal to watch with its threshold and window,
  and the recovery action it calls for. After the deployment, the health check
  compares the signals against them and reports a trigger that fires.

### Changed

- Workflow rules no longer contradict each other across plan, build, verify and
  close: the dedicated review runs once on the local candidate before the
  validation request or the push (a reviewer that only works on a pull request
  runs right after it), and under `automatic` a review is always offered and
  never recommended away; `automatic` stays automatic for UI changes and
  reports the first rendered view without asking; `interactive` stops before
  the push even when verify is skipped; a later completion point is defined
  once and authorizes both ticket moves; a question that depends on another's
  answer waits for a later turn; the findings question offers resolving them in
  this work; unattended work starts on silence after the route answer when
  delivery and review are declared; branch and worktree deletion wording is
  safe; one delegation criterion; role briefs match role contracts. The site,
  README, `llms.txt` and CONTRIBUTING describe the result (#180).

- `flow-build` ends at the built candidate and hands it to `flow-verify`; the
  independent UI children, the full suite, the dedicated code review and the
  walkthrough delivery move there, and promotion and deployment checks move to
  `flow-deploy`. `flow-walkthrough/references/ui-gate.md` and
  `ui-review-criteria.md` move to `flow-verify/references/`.
- The security review (`hive-review-security`) runs only when you ask for it.
  A declared `Review` value no longer authorizes it, and the accepted-candidate
  gate, the role description and the `security-advisory` scope assessment no
  longer launch it. When a change crosses a security boundary, the agent names
  the boundaries and offers the review in one line, without a question that
  waits and without stopping the work: in the delivery statement, also when
  `Delivery` and `Review` are declared, or in the next message once a boundary
  appears. A requested review runs before validation, or before the push or
  merge when there is no validation stop (#167).
- Verification at checkpoints: a plan with numbered criteria groups its tasks
  into checkpoints chosen by risk, and `flow-build` verifies the tasks of a
  checkpoint together instead of after each task. A task that depends on a
  task of its own checkpoint starts without waiting for a verdict, and tasks
  that share a file never run at the same time. Each task is committed when it
  is written, with fixes in additional commits that name it, and the work is
  published for merge only when every task is verified or accepted by you. A
  plan without checkpoints is still verified task by task (#168).
- The UI review keeps the UI library's design system as its baseline. The
  reviewer first reads the library configuration, theme and installed
  components, the project UI guide and any declared accessibility requirement.
  A library default or a recorded project decision is not a finding in the
  checks that measure that value, such as contrast, even below a WCAG
  threshold, unless the project declares an accessibility
  requirement or the ticket asks for one. Non-text contrast follows WCAG 2.2
  SC 1.4.11 precisely, a fix that customizes the theme or an installed
  component needs your decision or the project's, and a global finding is
  reported once and recorded in the project UI guide so it does not reappear
  for every view (#173).
- Incidental findings at a close: a P3 finding or a wording detail is no
  longer offered as a ticket, related P2 findings are proposed as one group, a
  UI finding is proposed only after it was seen rendered, and the close
  question does not recommend a ticket opened from the work's own findings. UX
  reviews report earlier problems only on the screens the change touches, and a
  deployed review whose code matches the local one starts from the local
  report (#159).
- A task whose second verification still fails no longer stops to ask you
  when the remaining fix is small, inside the task and obvious: the
  implementer makes that one correction and a third verification checks it;
  you are asked when the fix needs your decision or that verification also
  fails (#156).
- Fewer questions in the flows: effects beyond the local machine are
  authorized by class (local, non-production, production, critical) instead of
  one at a time. The test data of a non-production environment takes one
  question for the whole set, while each critical effect keeps its own. The
  steps of one objective are asked together and answers already in force are
  reused. Minor findings from plan reviewers are applied and listed in the plan
  summary, follow-up fix pull requests share the ticket's delivery, a condition
  on a grant lasts only for its stated scope, and a research route option now
  says whether it opens a ticket instead of a separate ticket question (#151).
- Unattended delegation keeps running when you write only to tell the agent
  to continue, acknowledge progress or ask how it is going; any other
  message ends it (#143).
- A costly check now has a definition: it spends money or a third-party
  service's quota, changes shared or remote resources, or is expected to run
  longer than 30 minutes; ordinary reads such as fetching dependencies do not
  count. A local check of minutes with no external cost and no irreversible
  effect is routine. The authorizations a plan
  foresees for later steps are asked in the delivery and review round, while
  you are present, and one left unanswered stays unauthorized (#144).
- A task added to a retained plan during the build or after validation is
  planned like any other and, when its test approach is `tdd` or `check`,
  verified by `hive-verify-task` before the work is published for merge; one that arrives after
  the code review ran is reported and you are asked whether to review it.
- Closing a change record now checks, after applying its deltas, that every
  added or modified requirement matches its delta and every removed one is
  gone, and marks the change completed only after that check. An abandoned
  change archives its never-applied deltas under an `unapplied-` folder (#117).

- A ready plan's summary shows each acceptance criterion and changed
  requirement, and choosing to start Build at the route question approves
  them by ID; a free-text correction revises the plan and asks again. A
  pull request now states what stayed unverified and the assumptions the
  result relies on and, when the work has acceptance criteria, the recorded
  evidence for each one (#134).

- In a maintainer checkout, refreshing viewer runtime or embedded-asset changes
  now includes restarting an already running local viewer with its existing
  port, state directory and TTL, followed by health and runtime verification.
  A stopped viewer stays stopped.

- The message before a question card must be visible text in the response,
  not only in the agent's reasoning; a brief output style shortens it but does
  not remove it (#70).
- Repository guidance: the two-session bar for new guidance rules applies only
  to corrections of a rule that already covers the situation. New features or
  requirements, including ideas from external references, and gaps no rule
  addresses are adopted by the maintainer's decision.
- Walkthrough publication keeps JSON only when the `hive` executable is on
  `PATH`; standalone HTML is generated only without that binary. A stopped
  viewer does not select HTML, and previous pages remain unchanged. Retaining
  the manifest retains its referenced images and recordings.
- Upgrade the binary before the content that produces `delivery` metadata and
  explicitly restart an existing viewer to activate its new reader and UI.
  Older strict readers reject the optional field; historical manifests without
  it remain valid and are not rewritten automatically.
- Research, Plan and Build check what depends on a component before
  recommending, planning or building a change that removes, closes,
  disables, migrates, renames or moves something running in production, or
  that changes untested production behavior. The new `change-impact` reference always searches for
  the literal value involved and runs its topology, reachable-path and
  observation steps only when their signal appears. The interface rule now
  asks to search for consumers first, and a bounded observation of real
  traffic only chooses the approach. `hive-review-code` reports a pre-existing
  defect that the change makes reachable as introduced by the change (#110).
- The security policy supports only the latest published minor release,
  instead of naming 0.1.x and 0.2.x (#108).
- A child or the main thread no longer runs `git stash` while another agent
  works in any worktree of the same repository, because every worktree
  shares the stash list (#118).
- A plan lists in-scope work whose question cannot be stated precisely yet
  under "Not yet specified", apart from exclusions and open decisions; no
  task depends on it until it can be stated (#120).
- Diagnosis confirms a forced failure with a diff, shrinks the reproduction
  and compares it with the closest case that works, raises the rate of an
  intermittent failure before hypotheses, changes one variable per
  experiment, and tags temporary instrumentation for removal. A new
  `ticket-body` reference sets what a ticket Hive opens carries: current and
  expected behavior, dated evidence, criteria and what is out of scope.
  The Inspiration page credits Skills For Real Engineers.
- Build's test-first cycle takes one behavior at a time, writing the next
  test only after the previous one passes, and replaces with a test double
  only what crosses the system boundary, never the project's own modules.
- After a compaction, the active skill's `SKILL.md` is reread in full, since
  a host may re-attach only its start; Claude Code keeps the first 5,000
  tokens of each invoked skill. Each `SKILL.md` stays at or under 16,000
  bytes, checked by a content test and by harness-audit's HA-SK-02:
  flow-build moves its retained-plan procedure, test-first criterion, tool
  and dependency rules and other situational rules to references, and
  flow-plan moves its interface and plan-revision rules.

### Fixed

- Plan order: the plan format now states that a task depends only on earlier
  tasks that exist and that checkpoints are numbered without gaps, the last
  `tdd` or `check` task carrying the highest; plan review, the readiness check
  and tasks added during the build block a plan with numbered criteria that
  breaks this order, which the build's start rules could not handle (#176).
- The viewer uses subtle theme borders instead of inherited white outlines,
  and all dialogs inherit its shared base font while walkthroughs retain
  their own typography.
- Installing or removing pi-subagents no longer waits forever on a hung `pi`:
  `pi install` and `pi remove` stop after 15 minutes or on Ctrl-C, a closed
  terminal or SIGTERM, together with npm and every process they started, and
  the operation fails with an error that `hive recover` can resume (#109).

## [0.5.0] - 2026-10-23

### Added

- The walkthrough viewer groups compact sessions directly under projects, with
  session counts, full descriptive titles, ticket/date metadata and a route icon.
  Its grouped toolbar shows work status and actions to archive, restore, pin and
  delete a copy. Archived copies are hidden by default and can be shown with the
  archive icon beside Projects. Archiving is independent of pinning, retention
  and manifest status. Projects initially expand only when they contain work in
  progress; manual expansion, sidebar visibility, archive filtering and selection
  survive refresh within the same browser tab and origin.

- `hive view`, a local walkthrough viewer the user starts and stops
  (`start`, `stop`, `status`, `open`, `serve`, `register`). Agents register repositories and use
  `hive view status --dir <work folder>` to confirm closure; the server on 127.0.0.1 copies each repository's
  `walkthrough.json` and its files into Hive's state directory, indexes project, ticket and session metadata and presents sessions grouped by
  project in a compact sidebar, purges unpinned copies after 14
  days (`--ttl`), and deletes a copy on request without touching the
  repository. A new manager command raises the minor version of the next
  release.
- `flow-walkthrough` renders its page from data: the agent writes
  `walkthrough.json` as the new `references/walkthrough-manifest.md` sets and
  runs `scripts/render_walkthrough.py`, which validates it and writes
  `walkthrough.html` from a fixed template with a summary, a section index and
  a capture carousel; the page's fixed labels follow the manifest's `lang`
  (`es` or `en`). The agent no longer copies and fills an HTML skeleton.

- `codex-question-card`, an optional capability that `hive install` offers
  with the Codex host, selected by default in interactive installs when the
  installed Codex reports the feature in `codex features list` (uncheck it or
  type `none` to skip; non-interactive installs never run it; an older Codex
  gets a manual step that says to update it). After consent it runs `codex features enable
  default_mode_request_user_input`, Codex's own configuration writer, so Codex
  shows its question card in Default mode instead of asking in text; the step
  is verified by reading the Codex `config.toml`, `hive doctor` shows a
  `Codex question card` row with the record and the enable, disable and list
  commands, and `codex features disable …` turns it off. The text installer
  gains a default selection ("Press Enter to accept the default selection, or
  type none to skip."), an offer can show an effect instead of a version
  phrase, and the installer's boundary is stated in the tool-dependencies
  document: Hive edits no host configuration file itself and, with consent,
  runs the host's own command that writes one. A capability selected by
  default raises the minor version of the next release.

- `flow-walkthrough`, a skill that owns the human validation of a user
  interface: the first rendered view of a new layout, the correction rounds,
  the captures, a short recorded pass when a flow needs one (`agent-browser
  record`, checked with `ffprobe`) and the handoff as a message and a
  dark-mode HTML page rendered from a JSON manifest (see below). Its reference
  `ui-gate.md` is the one home of the UI gate: the implementer's rendered
  inspection, the independent children with their briefs and stops, and the
  accepted-candidate gate; `ui-review-criteria.md` moved beside it. Build's
  verification reference now points to them instead of holding the handoff
  and the gate.
- `UI verification`, an optional `## Hive` key with `full`, `functional` or
  `none`, selects which independent children run on a built UI change; without
  it, Build asks once per repository and session in the review round.
  `hive doctor` and the project form accept the key.
- The documentation site adds Getting started, Skills, and Inspiration pages
  alongside the complete release history from this changelog, with unreleased
  work distinguished from published versions.

- A code repository can declare `Delivery: automatic` or `Delivery: interactive`
  and a `Review` mechanism (or `none`) in its `## Hive` section; the agent then
  states the mode and the review in one line before the first edit instead of
  asking, and `hive doctor` and the project form accept the new values.

### Changed

- UI evidence is refreshed for affected views and journeys after any later fix,
  including code-review fixes after approval. A changed visible result needs
  approval again; unaffected evidence remains valid. Adopted after the Facti
  Claude Code session `c574e5a1-ec3e-44da-995e-2a9a89ba2e6b` of 2026-10-23.
- Gaps found in two sessions of 2026-10-19 (Grok `01a15706`, Codex `01a156af`;
  issue #93). A notice that only reports the exit of a process the agent
  stopped on purpose gets no message, also while a question waits. Close stops
  the processes the task's children left running, which the agent records when
  each child returns. The implementer of a UI change loads the changed view in
  a real browser with no test doubles before reporting it. Each UI gate child
  verifies a fixed candidate, named by commit or tree hash and served from its
  own build and port; a fix during the gate makes a new candidate. The first
  rendered view and the final validation offer to open the viewer or approve,
  with corrections written in the answer's free text. When an answer leaves a
  selection to the agent, it lists what goes in and what stays out and asks for
  confirmation before acting.
- Two costs seen in both Grok cam sessions of 2026-10-20 (`01a1579f`,
  `01a157db`; issue #94): after a context compaction the agent rereads only a
  reference whose condition applies to its next action, plus the active
  `SKILL.md` when its steps are not in the summary, instead of every reference
  read earlier; and it writes to persistent memory only at a user decision, a
  ready plan, a work-item close, a worktree's creation or removal, and a session
  summary when the user ends the session, a cadence that holds over memory
  tools' own save instructions. Close writes its memory records in one pass.
- The pilot runner (`tests/pilot`, not installed) accepts a text-only close
  question only on hosts whose headless mode has no question tool (Codex,
  Claude, Pi); on Grok and OpenCode it now requires the native question (#66).

- The local walkthrough viewer replaces its vanilla shell with React and public
  MIT shadcn/ui and ReUI components in a custom compact project/session sidebar.
  Its versioned bundle runs inside the Go binary without Node, a ReUI license
  or third-party network access. Frontend development uses Node >=24 and pnpm
  12.8.1, with frozen dependencies and read-only bundle freshness checks.
- Walkthroughs carry optional explicit `session.status` (`in_progress` or
  `closed`); historical manifests without it remain `unknown`. Badges are
  independent of Git classification. `flow-close` publishes closure and confirms
  the exact current viewer copy before eligible source cleanup; an unconfirmed
  copy keeps its evidence. See the
  [canonical close procedure](content/skills/flow-close/SKILL.md#close-the-current-walkthrough).
  Update the viewer binary and installed content together, and coordinate a
  restart of an existing older viewer before using the new producers: replacing
  files does not update its running process.

- `flow-walkthrough` takes captures by default and records a flow only when it
  validates an interaction or a transition that captures cannot show, or when
  the user asks; it no longer records every flow at the validation handoff.
  The walkthrough skeleton leaves `flow-report`, and the walkthrough's
  retention unit now includes `walkthrough.json`.
- `flow-walkthrough` opens nothing by itself: the message gives one line with
  the page's path, and the agent runs `hive view open` only when the user
  chooses that option in the validation question.

- Inspiration opens one page containing OpenSpec, gentle-ai, and Superpowers
  instead of a collapsible sidebar group.

- Waits, messages and the first rendered view follow one rule each (the
  analysis of cam session `e2c8d39f`, 2026-10-18). Every progress or decision
  message opens with a bold context line, stage · ticket · progress, and a
  question card opens with the stage and ticket. A turn that a notification
  starts without changing what you need is that line followed by the change,
  without listing the work still running or repeating an open question. A
  child that returns, or that the host reports as stopped, while it waits on
  work of its own is treated as stopped and checked at once. A task brief
  carries the skills and project documents that the project's rules direct for
  that work. The first rendered view of a layout is delivered as
  `walkthrough.html`, and the message gives the command to open it before the question. A task
  with a rendered UI effect is committed after its first rendered view and the
  UI review, and the UI review, the full suite and, when the change crosses a
  security boundary, the security review run before you are asked to validate; later fixes go in an additional commit per
  task, never an amend. The Build, Close, Research and Plan pages describe
  this, and the delivery notes record Claude Code's notice for a stopped child.

- The language rule names URL paths and route segments, including those of
  public pages, among the identifiers written in English: a URL path is an
  identifier, not user-facing text. `hive-review-code` reports a new
  identifier, file name, URL path or route segment in another language as a
  finding, and `hive-verify-task` notes it beside its verdicts.
- The global guidance points to the browser-automation reference before any
  browser is driven, including recordings and demos outside a build.
- Unattended delegation has a fourth source: the user's silence after they
  answered the route, delivery and review questions of the running work
  delegates that work up to the delivery point they chose. The skill names
  the force-push with an exact lease and the pull-request text edits it
  allows, each with its way back, and reports an inferred activation as
  inferred.
- The labels the host shows while the agent works, such as a command's
  description, a task's subject and a child's label, follow the session
  language; a contribution to an external or open-source repository first
  checks that repository's language and uses it for what reaches it.
- The documentation site uses the Ion theme with Hive branding and an original
  hexagonal hero illustration; contributing guidance covers site content updates,
  preview checks, and the single changelog source. Flow pages match the canonical reporting reference's applicability.

- Before merging a pull request, the agent confirms that the passing checks ran
  against the current base; when the base moved, it merges the base into the
  branch, pushes, and waits for the new run.
- `hive-review-code` reports a repeated defect once with every location and
  says in its verdict whether the diff crosses a security boundary, which
  triggers `hive-review-security`; the urgent-fix postmortem names the layer
  that should have caught the failure and places the preventive action there.
- At close, the agent may delete merged work branches from this work or
  earlier, with their clean worktrees, as the close procedure already did;
  the global grant now says so, and environment branches are never work
  branches (observed: a repository with seven merged work branches left on
  GitHub).
- The global guidance no longer states how review-gate findings are fixed; the
  severity rules in Build's verification reference are their only home, and
  Build reads that reference for a code review too. A review requested on its
  own still authorizes no fixes.
- The global guidance keeps one home per rule: the `openspec` location is one
  sentence plus the change-records reference, and two sentences that the
  retention and reporting references already carry left the global text.
- The `hive-write-docs` role (formerly `hive-write-spec`, see Removed) writes
  documentation and site pages from sources the plan already verified and
  reports what a source does not cover as not verified; the change record
  stays with the main thread.
- Plan has one home for revising a retained plan while the work runs
  ("Revise a plan in progress"): who edits, how the request, criteria, tasks
  and deltas stay coherent, when the revision returns to the plan reviewers,
  and when a change of intent is a new change. The seven places that stated
  parts of it now link there.
- Plan review records each reviewer's duration and token usage when the host
  reports them, so the gate's cost can be weighed against what it caught.
- The pilot runner (`tests/pilot`, not installed) keeps shell text outside the
  bounded-read grammar and glob/grep patterns in a private `0600` file beside
  `run.json`, types every inventory entry, and compares reference coverage on
  reads normalized of OpenCode's line prefixes, so a batch can be inspected
  and arms A/B can be observed (issues #48, #49).

### Removed

- The `hive-write-spec` role name. `hive update` retires the installed file;
  a retained plan or brief that names it should name `hive-write-docs`
  instead. Renaming a role raises the minor version of the next release.

### Fixed

- The local viewer and documentation site serve the Hive favicon without
  missing-resource requests. The viewer uses the correct image MIME type and
  also supports the default `/favicon.ico` request, without third-party requests.

- A pull request is merged with the method the repository declares or its base
  branch's recent merges show, a merge commit by default (Grok cam session
  `01a1579f`, 2026-10-20: PR #239 was squashed while the base branch's earlier
  merges were merge commits, collapsing the per-task commits).

- Close deletes a branch merged by squash or rebase, with its clean worktree,
  once `gh` shows its pull request merged from the same head; `git branch -d`
  refused them, so both Grok cam sessions of 2026-10-20 (`01a1579f`,
  `01a157db`) left orphan branches and worktrees. The retained checkout leaves
  the branch and eligible clean worktrees are removed before branch deletion.

- A process the agent waits on gets a bounded watch besides its exit signal, so
  a hung run still wakes the agent, and a turn that ends with work running says
  what will wake it and the duration or deadline of the current wait (Grok cam
  session `01a1579f`, 2026-10-20: a hung backend
  integration suite left the session silent until the user asked; issue #95).

- A stalled local integration run is handed to the pull request's CI only after
  confirming that CI runs that suite on pull requests; otherwise it is rerun
  locally or reported as not run (Grok cam session `01a157db`, 2026-10-20:
  a never-run test was committed on the assumption that PR CI would run it).

- The walkthrough viewer distinguishes local and unknown sources, skips new
  manifests confirmed as tracked in Git, and sets existing tracked copies apart
  without rendering their instructions or changing their pins and retention.
  Ingestion checks opened manifests and resources for hard links. Desktop capture
  previews use the full strip width on screen while preserving narrow-screen
  navigation and the print layout.
- A yes-or-no question, such as an authorization, lists its two options with
  IDs and effects like any other question, and a question tool that returns
  before the answer, such as Codex's asynchronous card, never replaces the
  options in the text. In a Codex Default-mode session on 2026-10-18 the
  Bugbot authorization went out as a bare "do you authorize it?" and a
  history-rewrite authorization put its options only in the asynchronous
  card, which closed at the end of the turn, so the user had to answer twice
  without options.
- The Codex `observe` profile no longer declares `sandbox_mode = "read-only"`
  in the installed role files: Codex restores the session's permissions after
  applying a role (verified on 0.163.0; the 0.155.1 note is consistent with
  this reading), so the key did not reach the child. Profiles frozen in
  earlier releases stay readable, and the renderer drops the key for Codex. A child whose browser is blocked by
  the host's sandbox, such as Codex's default `workspace-write` without network,
  reports the UI gate as not passed naming the block, and the user lifts it in
  the host's settings (issue #50).

## [0.4.0] - 2026-10-15

### Added

- When a human reviewer is assigned or requested on a pull request, the agent
  reads each of their comments against the code before changing anything, asks
  in the thread when one is ambiguous, applies them blocking-first and verifies
  each one, disagrees with evidence when the code already does what was asked
  or the request would break declared behavior, and replies in each thread.
- The completion report lists each implementation decision the agent took
  without asking, when the user could reasonably have chosen otherwise, with
  its reason and how to revert it, in attended and unattended work alike.
- Pi user-scope installs run `pi install npm:pi-subagents` without a version
  pin, so npm resolves the latest release; an existing pin is left unchanged.

- `hive install` offers `agent-browser` as an optional capability: in an
  interactive install on macOS arm64 (`brew install agent-browser`) or Linux
  amd64 with Node 24+ and a writable npm prefix (`npm install -g
  agent-browser`), it shows the exact commands, installs the latest version
  after your consent, and runs `agent-browser install` to download Chrome only
  when no browser is found. Commands run without a shell or `sudo`, with a time
  limit; a failed step is reported without touching the installed core.
  `hive doctor` now names `hive install` for a missing `agent-browser` or
  browser. Elsewhere the capability stays manual. This adds an installation
  route, so the next release raises the minor version.
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

- Children never receive the parent's conversation history: when the launch
  tool offers to pass it, the agent launches without it, since the brief
  carries what the child needs and inherited history never delivers the skills
  the parent read. Children never launch children of their own, and a child
  that cannot finish alone reports it; the ten implementer roles say so.
- Before calling a change mechanical, the agent checks whether the code it
  alters also serves behavior the request did not name, such as a shared
  helper, parser or option; if it does, the change alters a contract and is
  not mechanical.
- A failure you report, such as a bug, an error or a regression, with or
  without evidence, enters through Research unless the same message asks for
  the fix: the agent confirms the cause first, then asks in one round whether
  to track it and how to continue, and Build asks the delivery and review
  questions when it starts.
- `hive-verify-task` runs on its configured verifier model where the host sets
  one, as a preference rather than a condition: the agent no longer compares
  models before each launch, asks which model to use, or discounts a verdict on
  the child's own statement of its model. What makes the verdict independent is
  a fresh context and a brief without the implementer's report or claims; the
  verifier never runs on a smaller model or lower effort than the implementer,
  and when it cannot run it is retried once on the implementer's model before
  the agent asks.
- When an MCP server or other integration the task needs fails to connect or
  authenticate, the agent says so in its next message, names what it uses
  instead and what that source cannot show, and the action that restores it,
  instead of substituting another source silently.

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

- OpenCode roles no longer fail to launch with `Variant unavailable`: the execution profile uses `high`, the profiles declare the variants each OpenCode model offers, and an effort a model lacks (a role's `medium`) maps to the nearest offered variant; a test renders every role and fails on an unoffered variant. Existing installations need the rebuilt binary and `hive update` with this content: an installed release keeps its frozen profiles, and an older manager rejects the new `variants` key.
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
