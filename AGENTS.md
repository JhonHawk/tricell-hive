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
- Put an activity's essential procedure in its `SKILL.md`. Link situational references directly from that file and state the condition that requires each read. A listed skill or reference is not proof it was selected, read, or followed.
- When authoring, moving, or removing distributed skills, agents, or their resources, read `_support/docs/architecture/instruction-resources.md` for canonical references, dependency checks, and runtime resolution.
- In distributed content, name the capability rather than a host's tool, such as "the host's native question tool", and give the fallback when it may be absent: tool names change across releases and modes. Keep a proper name only where the behavior itself differs by host and a capability-first rule has been shown insufficient; issue #30 found delegation needs no per-host table, so the host dialects live in `_support/docs/architecture/agent-delivery.md`.
- State a rule's criterion with its scope explicit, since current models apply instructions literally. Add the reason when it marks where the rule applies, and use examples only to disambiguate a boundary: a lone example reads as the rule's whole scope.
- Write distributed content for the least capable model that runs it: the `execution` profile models in `integrations/agent-profiles.json`, which run delegated roles on every host that sets one, and on OpenCode the main thread and every role outside the `reasoning` profile. A frontier model's prompting guide or default behavior alone does not justify removing a rule; removal needs evidence that those models behave correctly without it.
- Give each rule one canonical home. Avoid duplicate routers, fallback chains, generated copies, hooks, or adapters without a measured need.
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

Research findings stay in the conversation unless the user explicitly requests saving them or accepts a concrete retention proposal. The shared research-retention rule in `content/guidance/global.md` also applies here; requesting complete research does not request a document.

Delegate the three fronts to independent subagents in parallel when capacity permits. The main thread acts as an adversarial reviewer: check source support, challenge assumptions, resolve contradictions, distinguish observations from recommendations, and synthesize the result. Do not accept a subagent's conclusion solely on its assertion; disclose unavailable sources or incomplete coverage. Keep the investigation proportional to the question and preserve the normal authorization boundary: research does not authorize implementation, deployment, or model pilots.

## Workspace locations

Sessions in this repository require the deployed Hive global guidance, whose support-folder rules govern artifact placement here for every task, even when no skill is active. When a session does not show that guidance (for example, a host where Hive is not installed), read `content/guidance/global.md` before writing any artifact.

- Repo-scoped work uses this repository's `_support/`; `_support/workspace/` is git-ignored.
- For requested organization of existing material, evidence curation, promotion, or archiving, read `content/skills/workspace-conventions/SKILL.md`.
- See `_support/docs/architecture/workspace-and-artifacts.md` for explanations and examples.

## Scope and preservation

- A question, research task, or recommendation does not authorize implementation or unrelated fixes. Keep incidental findings separate from the requested work.
- Do not deploy or copy configuration into global tool directories, change global tool settings, commit, push, merge, or delete branches without the user's explicit instruction for that action. Research and recommendations are not that instruction.
- Keep credentials out of tracked files, reports, and tool output; protect necessary local recovery copies and exclude them from Git.
- Preserve user-owned authentication, preferences, native histories, third-party tools, backups, and Engram data. Keep `.engram/config.json` project identity unless the user explicitly asks to change it.
- Do not add dependencies, a runtime, universal adapter, hook system, or build framework for configuration alone. Prefer the host's existing capability and document a verified limitation before introducing new machinery.

## Refreshing a local installation

Refresh a local installation only when the user asks, as the rule above requires. **Maintainer checkout only:** when `_support/workspace/maintainer.local` exists in this checkout (an ignored marker the maintainer creates; contributors' clones lack it), each commit or merged pull request that lands on `development` is that request. Refresh right after it lands, following the steps below for what it changed, and report the deployed commit. Without the marker, refresh only on request.

The binary and the deployed content update separately:

- **Manager changes** (`tooling/`, `integrations/`): rebuild the binary from this checkout with `go build -o "$(command -v hive)" ./tooling/cli`. `hive update` does not replace the binary.
- **Content changes** (`content/`): run `hive update` from this checkout, or add `--source <checkout>` elsewhere. It deploys the committed `HEAD`, never uncommitted edits, to every registered host, and asks for confirmation. Use `--dry-run` to preview. Open sessions load the new guidance only after a restart, except that OpenCode applies edits to its global and upward-discovered `AGENTS.md` files before the next model request; start a new OpenCode session for changed roles or skills.
- **Both:** rebuild first, since an older binary can reject newer content.
- **Offline-package installations:** `hive update` needs a Git checkout, so those users update by running the new package's `./install.sh`.

`_support/docs/architecture/deployment-manager.md` holds the commands and their limits.

Refreshing a local installation is not a release. Cut a release only when the user asks; propose one in the closing report when the criterion in [CONTRIBUTING.md](CONTRIBUTING.md#releases-and-version-numbers) is met.

## Measurement

CLI behavior pilots are paused by user instruction. Do not resume them without explicit authorization. Recommend a pilot only when it is necessary to resolve a consequential behavior uncertainty that inspection or ordinary tests cannot answer; explain the expected evidence and keep the proposed scope minimal. This pause does not prohibit ordinary non-model tests.

A new or stricter rule in distributed guidance, skills, or roles needs a failure observed in two independent real sessions, or the user's explicit decision to adopt it without that evidence. An idea from an external reference meets the same bar; a plausible improvement alone does not. Record a single observed case as a watch item, in an issue or the maintainer's memory, with its count and the occurrence that would justify the rule. Fixing text that contradicts its own intent or the source it describes needs no recurrence.

A guidance change that fixes a failure observed in a real session names that session (host, session ID, date) and the observed failure in its commit message. Do not add a new regression case under `tests/fixtures/regression/` for it: those cases check a detector against synthetic traces, and no real session runs through them while pilots are paused. Keep the existing cases unchanged unless the user decides otherwise.

For behavior claims, record the host and installed version, resolved model, task, relevant instructions and references read, writes, human intervention, and terminal state. Separate guidance discovery, source reading, and task outcome. Use observable criteria and the least costly useful check. Use blinded human review when asserting subjective quality improvement; label model grading as exploratory. A single run, prompt-size estimate, or cached request does not establish reliability or improvement.

Pilots may use the deployed global Hive installation; use this delivery for new flow pilots and record the installed release and content hashes. Do not inject duplicate Hive instructions into their fixture projects. Keep test inputs and writes in disposable fixtures, record process-local overrides, and preserve unrelated global configuration. An isolated project-delivery experiment remains possible when its purpose and different conditions are explicit. Pilot authorization does not imply unrelated global deployment.

Before behavior tests, declare whether they use everyday or isolated memory. Use a fresh, verified-empty Engram data directory per run when prior fixture memories must be excluded, and disable cloud synchronization for that test process. Record the requested environment separately from observed memory routing; check for fixture leakage into the everyday store. Do not describe process-local memory isolation as isolation of all host history or filesystem access.

After tests finish, remove fictitious test records from Engram. Identify the exact fixture projects, sessions, and observation IDs from the run evidence before deletion; do not delete by a broad topic match or ambiguous project name. Preserve real Hive decisions, findings, and evaluation summaries. Wait for all test writers to finish, use Engram's supported deletion mechanism, verify the scoped records are gone, and record the cleanup outcome in the session report. If provenance is uncertain or scoped deletion is unavailable, report the unresolved records instead of deleting unrelated memory.

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
