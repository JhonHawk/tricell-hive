# Hive repository guidance

> **Audience:** this file governs agents that maintain Hive inside this repository. It is not guidance for using Hive in other projects and not a file to copy: Hive's distributable guidance lives in `content/`, and [README.md](README.md) and [llms.txt](llms.txt) describe the project for users and agents.

This repository is in a rebuild phase for a small, portable guidance layer targeting Claude Code, Codex, Grok Build, Pi, OpenCode V2, and Cursor CLI. Target inclusion is not proof of an installed or working integration. Let each CLI own its model loop, authentication, tools, permissions, and native history. Keep shared policy small; add a host-specific mechanism only when current documentation and an observed failure justify it.

## Hive

- Project: tricell-hive
- Base branch: development
- Environments: development → master
- Tracker: GitHub Issues · JhonHawk/tricell-hive
- Specs: _support/openspec

## Guidance structure

- Keep cross-task rules and essential navigation here. Do not make basic authorization or artifact placement depend on an activity skill being selected.
- Put an activity's essential procedure in its `SKILL.md`, which stays at or under 16,000 bytes because a host may keep only its start after compaction; `tests/content/skill_size_test.go` checks it, and also keeps each reference at or under 16,000 bytes, so a reference past it is split by responsibility. Link situational references directly from that file and state the condition that requires each read. A listed skill or reference is not proof it was selected, read, or followed.
- When authoring, moving, or removing distributed skills, agents, or their resources, read `_support/docs/architecture/instruction-resources.md` for canonical references, dependency checks, and runtime resolution.
- In distributed content, name the capability rather than a host's tool, such as "the host's native question tool", and give the fallback when it may be absent: tool names change across releases and modes. Keep a proper name only where the behavior itself differs by host and a capability-first rule has been shown insufficient; issue #30 found delegation needs no per-host table, so the host dialects live in `_support/docs/architecture/agent-delivery.md`.
- State a rule's criterion with its scope explicit, since current models apply instructions literally. Add the reason when it marks where the rule applies, and use examples only to disambiguate a boundary: a lone example reads as the rule's whole scope. Give each criterion an observable test or threshold instead of a qualitative word such as "small", "important" or "substantive", since literal and smaller models act on what they can check and drop what they cannot. Keep a fixed sequence only where the order matters or an error is severe or irreversible, and move a rule with no exceptions into a test, hook or script when one can enforce it.
- Write distributed content for the least capable model that runs it: the `execution` profile models in `integrations/agent-profiles.json`, which run delegated roles on every host that sets one, and, on hosts whose profiles inherit the session model, such as Pi, OpenCode, and Cursor, whatever model the user's session runs. A frontier model's prompting guide or default behavior alone does not justify removing a rule; removal needs evidence that those models behave correctly without it.
- Give each rule one canonical home, which alone states its full steps, their order, and its exceptions. Where the rule applies elsewhere, add only a pointer that names the rule and when to read it, such as "before asking the user, read and follow [questions](…)"; never restate some of its steps there, because a model follows the shorter version and skips the rest. Restate a rule in `content/guidance/global.md` only when it must act without its home being read, and then state its minimal invariant whole. An exception that applies at only one point of use stays there and names every element of the rule it changes. Correct an existing copy when a change edits that copy or the rule it restates. Avoid duplicate routers, fallback chains, generated copies, hooks, or adapters without a measured need.
- Before changing a distributed rule or its wording, review the whole behavior it governs, not only the sentence you edit. Find every place that states or depends on that behavior: the global guidance, skills and their references, agent roles, `site/` pages, `_support/docs/architecture/`, and the detectors in `tests/pilot/`. Confirm that the change leaves no conflicting or duplicated instruction; drops no intent of the text it replaces, or moves it elsewhere, or states why it goes; keeps one clear order where several rules govern the same moment; and states the change in `content/guidance/global.md`'s size in the pull request. List the places checked in the pull request. A typo or formatting fix that changes no meaning needs only the search.
- Before changing host-specific behavior, check current official documentation and the installed CLI version. Mark documentary claims, runtime observations, and inferences separately.

## README and llms.txt

- `README.md` is the entry point for people: what Hive is, how to install it, the supported hosts, and how to update, recover, or uninstall it, plus a link to the documentation site once the site is published. Each flow's detail lives on the site, in `site/`, not in the README.
- Update `README.md` and `llms.txt` in the same pull request as any change that alters something they state, such as an installation command or flag, a supported host, a requirement, a user-facing manager command, or the repository's top-level folders.
- Update a page in `site/` in the same pull request as any change to the flow behavior that page describes. Pages describe behavior and link to the skill; they do not copy its rules.

## Language convention

- Write distributed CLI instructions and repository documentation intended to guide agents in English, including rules, skills, references, and implementation contracts.
- Write human-facing reports and deliverables in the language of the session, unless the user requests another language. Classify by intended purpose, not file extension or directory; an agent reading a human report does not change its audience.

## Complete research

When the user requests a "research completo", cover three fronts: the current Hive implementation on the base branch (verify reference freshness), external evidence from official documentation, research, developer blogs and firsthand community discussions, and optional reference sources when available.

Research findings stay in the conversation unless the user explicitly requests saving them or accepts a concrete retention proposal.

Delegate the three fronts to independent subagents in parallel when capacity permits. The main thread acts as an adversarial reviewer: check source support, challenge assumptions, resolve contradictions, distinguish observations from recommendations, and synthesize the result. Do not accept a subagent's conclusion solely on its assertion; disclose unavailable sources or incomplete coverage. Keep the investigation proportional to the question and preserve the normal authorization boundary: research does not authorize implementation, deployment, or model pilots.

## Workspace locations

Sessions in this repository require the deployed Hive global guidance, whose support-folder rules govern artifact placement here for every task, even when no skill is active. When a session does not show that guidance (for example, a host where Hive is not installed), read `content/guidance/global.md` before writing any artifact.

- Repo-scoped work uses this repository's `_support/`; `_support/workspace/` is git-ignored.
- When a task involves external reference repositories, inspiration sources, attribution or private project decisions, read `_support/workspace/local-context/reference-sources.local.md` and `decisions.local.md` first. Both are local and untracked: never copy their contents, including the names of the reference repositories, into tracked files.
- Worktrees go in the host's native path (`.claude/worktrees/`, `.grok/worktrees/`), which `.git/info/exclude` ignores; never create a `../tricell-hive-worktrees` sibling folder, which the user removed.
- For requested organization of existing material, evidence curation, promotion, or archiving, read `content/skills/workspace-conventions/SKILL.md`.
- See `_support/docs/architecture/workspace-and-artifacts.md` for explanations and examples.

## Scope and preservation

- A question, research task, or recommendation does not authorize implementation or unrelated fixes. Keep incidental findings separate from the requested work.
- Do not deploy or copy configuration into global tool directories, change global tool settings, commit, push, merge, or delete branches without the user's explicit instruction for that action. Research and recommendations are not that instruction; the delivery and review answers given in the session, or a `Delivery: direct-base` value in a repository's `## Hive` section, count as the user's explicit instruction for the effects the «Delivery» section of [the global guidance](content/guidance/global.md) assigns to them.
- The maintainer's preferred code-review mechanism here is Claude Code's `/code-review`: it informs the Recommended option of the review question and authorizes nothing by itself.
- Keep credentials out of tracked files, reports, and tool output; protect necessary local recovery copies and exclude them from Git.
- Preserve user-owned authentication, preferences, native histories, third-party tools, backups, and Engram data. Keep `.engram/config.json` project identity unless the user explicitly asks to change it.
- Do not add dependencies, a runtime, universal adapter, hook system, or build framework for configuration alone. Prefer the host's existing capability and document a verified limitation before introducing new machinery.

## Refreshing a local installation

Refresh a local installation only when the user asks, as the rule above requires. **Maintainer checkout only:** when `_support/workspace/maintainer.local` exists in this checkout (an ignored marker the maintainer creates; contributors' clones lack it), each commit or merged pull request that lands on `development` is that request. Refresh right after it lands, following the steps below for what it changed, and report the deployed commit. Without the marker, refresh only on request.

The binary and the deployed content update separately:

- **Manager changes** (`tooling/`, `integrations/`): rebuild the binary from this checkout with `go build -o "$(command -v hive)" ./tooling/cli`. `hive update` does not replace the binary. The build takes every commit up to `HEAD`, including other sessions' commits on the same branch: before building, name the `tooling/` and `integrations/` commits not yet in the installed binary, so a change another session has not delivered is not installed with yours.
- **Content changes** (`content/`): run `hive update` from this checkout, or add `--source <checkout>` elsewhere. It deploys the committed `HEAD`, never uncommitted edits, to every registered host, and asks for confirmation; from an agent shell without a terminal, after naming the commits not yet deployed, run `hive update --out FILE` with `FILE` in the host's temporary location and then `hive apply --plan FILE`, and tell the user that open sessions need a restart. Use `--dry-run` to preview. It deploys every committed change up to `HEAD`, not only the latest commit: before deploying, compare with the source commit the previous deploy printed and name the commits not yet deployed, so a commit the user chose to hold back is not published with another. Open sessions load the new guidance only after a restart, except that OpenCode applies edits to its global and upward-discovered `AGENTS.md` files before the next model request, and Codex 0.166 reloaded the global `AGENTS.md` in an open session within seconds of an update (runtime observation, one session, 2026-10-31); start a new OpenCode session for changed roles or skills.
- **Both:** rebuild first, since an older binary can reject newer content.
- **Offline-package installations:** `hive update` needs a Git checkout, so those users update by running the new package's `./install.sh`.

**Maintainer viewer restart:** when the maintainer marker exists and the integrated changes affect the viewer runtime or its embedded assets, refreshing the local installation also authorizes restarting an already running local viewer after rebuilding the binary. Before stopping it, record its port, state directory and TTL; restart with those same settings, preserve its registry and stored copies, and verify health, ingestion and the changed runtime or assets. Report the active revision and resulting service state. If the viewer was stopped, leave it stopped. Outside this case, restarting an existing viewer still requires an explicit user request.

`_support/docs/architecture/deployment-manager.md` holds the commands and their limits.

Refreshing a local installation is not a release. Cut a release only when the user asks; propose one in the closing report when the criterion in [CONTRIBUTING.md](CONTRIBUTING.md#releases-and-version-numbers) is met.

## Measurement

CLI behavior pilots are paused by user instruction. Do not resume them without explicit authorization. Recommend a pilot only when it is necessary to resolve a consequential behavior uncertainty that inspection or ordinary tests cannot answer; explain the expected evidence and keep the proposed scope minimal. This pause does not prohibit ordinary non-model tests.

A correction in distributed guidance, skills, or roles, meaning a new or stricter rule for a situation an existing rule already addresses, written because that rule was not followed or falls short in a real session, needs that failure observed in two independent real sessions, or the user's approval to adopt it at once. Independent sessions are distinct sessions, neither a resumption nor a child of the other; sessions in the same project count. Count only sessions that ran the current version of the governing rule: a change to that rule restarts the count. A correction that adds questions, confirmations, or other ceremony for the user needs the user's approval even when the recurrence is met; relaxing or removing a rule is the user's decision, resting on real-session evidence that the models in use behave correctly without it.

Record a single observed case as a watch item in one Engram entry under the topic `watch/<slug>`, with its count, the sessions observed (host, session ID, date), the rule version it counts against, and the occurrence that would justify the rule or reopen it. Issues, change records, and reports point to that entry instead of keeping their own count. After a guidance change is integrated, check the open issues and watch items it touches against the real sessions that ran it: when the failure did not recur in at least three episodes that exercised the rule, close the issue as a watch item, recording those episodes and the reopening condition in its entry. A new feature or requirement, including one taken from an external reference, and a gap, a situation no rule addresses, need no recurrence, even when a real session revealed the gap: the user's decision adopts them, and a plausible improvement alone does not. Fixing text that contradicts its own intent or the source it describes needs no recurrence.

A guidance change that fixes a failure observed in a real session names that session (host, session ID, date) and the observed failure in its commit message. Do not add a new regression case under `tests/fixtures/regression/` for it: those cases check a detector against synthetic traces, and no real session runs through them while pilots are paused. Keep the existing cases unchanged unless the user decides otherwise.


Before resuming a pilot or running any behavior test of a host, once the user authorizes it, read [the pilot procedure](_support/docs/harness-engineering/pilot-procedure.md): what to record, the delivery to use, memory isolation and the cleanup of test records.

Before analyzing a real session, read [the session review rubric](_support/docs/harness-engineering/session-review-rubric.md) and report against it. Add a row there when a review checks something new that the user wants checked again.

## Repository map

- `README.md` — purpose and working boundaries.
- `content/` — distributable Hive guidance, activity skills, and canonical agents; root `AGENTS.md` governs this repository only.
- `integrations/` — host-specific differences; `tooling/` — management interfaces and shared operations; `tests/` — verification.
- `site/` — the documentation site for people; it is not installed.
- `_support/docs/architecture/repository-and-distribution.md` — structure and managed global-instruction block design.
- `_support/docs/architecture/deployment-manager.md` — Go manager commands, ownership, recovery, and verification boundaries.
- `_support/docs/architecture/agent-delivery.md` — canonical roles, native profiles, and inline delivery limits.
- `_support/docs/architecture/workspace-and-artifacts.md` — workspace scope, artifact organization, retention, and hygiene.
- `_support/docs/harness-engineering/README.md` — durable research index and measurement standard.
- `_support/docs/harness-engineering/2026-09-20-portable-harness-research.md` — source-backed host comparison, limits, and evaluation approach.
- `_support/openspec/` — current requirements and change records; `_support/workspace/` — ignored working material whose versioning is undecided.

<!-- BEGIN:reui-agent-rules -->
## ReUI design rules (every UI edit, follow-ups included)

This project uses an existing design system: ReUI on shadcn/ui. Keep its look. Generic frontend advice (vary fonts or themes, gradients, page-load or staggered reveals, "2-3 intentional motions", Framer Motion) does not apply here, and other design skills lose where they disagree.

1. Use installed components first (`components/ui`, `components/reui`): tables `data-grid`, filtering `filters`, boards `kanban`, modals `Dialog` or `AlertDialog`, toasts `sonner`. Never set height, radius, padding, shadow or motion on them; use `size`, `variant`, `spacing`.
2. Color: semantic tokens. State: `<Badge variant="success-light">` from `@/components/reui/badge` at default size, or Alert variants. Raw palette only for categorical identity, with its `dark:` pair.
3. Spacing: `gap-*` on the parent: `gap-1` to `gap-2` in a row, `gap-3` to `gap-4` between groups, `gap-6` to `gap-8` between sections. Type: `text-sm` body, `text-xs` meta, nothing smaller; never resize FrameTitle or CardTitle.
4. Motion: installed parts already animate; never retime them. Yours: hover `transition-colors`; reveal `duration-200 ease-[cubic-bezier(0.22,1,0.36,1)]`; transform and opacity only; never `transition-all`, bounce, or a stagger on table rows or lists of 20+; at most 1 entrance per view; `MotionConfig reducedMotion="user"` plus `motion-reduce:`.
5. Every data region: empty state with an action, `Skeleton` loading, inline error with Retry.
6. Before UI code, write the plan from `.agents/skills/reui/SKILL.md` in your reply. Before you finish, run its self-check on the files you changed and fix every hit (2 rounds at most). Full rules: `.agents/skills/reui/SKILL.md`. If that folder is missing, call the `reui` MCP tool `get_agent_skill`.

Managed by the ReUI installer. Edits between the markers are replaced on the next run.
<!-- END:reui-agent-rules -->
