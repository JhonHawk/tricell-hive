# Canonical agents and native delivery

Agent instructions have one source at `content/agents/<category>/<name>.md`. Categories organize authorship; all native destinations use the unique filename stem. Do not keep converted host copies in the repository or a tracked distribution directory.

`integrations/agents` parses a deliberately small frontmatter contract: `name`, `description`, `model_profile`, and `access_profile`; optional `effort` (`low`, `medium`, `high`, `xhigh`, or `max`) overrides the profile's effort for that role where the host can render it. Optional `effort_claude` overrides only Claude Code's effort and leaves other hosts on their profile defaults. OpenCode encodes effort as a model variant instead of a separate field. Values are single-line strings, optionally JSON-quoted. The body is preserved. Unknown fields, malformed sources, and unknown profiles fail validation. `integrations/agent-profiles.json` is the single host mapping. The deployment manager converts in memory during planning and freezes the native bytes for apply and recovery.

## Roles and profiles

Twenty roles cover design, implementation, documentation, operations, verification, and review, including harness-engineering audits by `hive-review-harness` and per-task verification of a retained plan by `hive-verify-task` (change `per-task-verification`, 2026-09-26). `hive-design-architecture` includes contract design formerly assigned to `sdd-design` and cloud infrastructure design formerly assigned to `cloud-architect`, which had no recorded invocation on any of the six hosts through 2026-09-25; `hive-review-security` includes detection formerly assigned to `secrets-auditor`. Dedicated `sdd-spec-reviewer`, `sdd-product-critic`, `workspace-custodian`, and `prompt-engineer` remain deferred. The reference checkout is preserved.

| Profile | Claude | Codex | Pi | Grok | OpenCode V2 | Cursor CLI |
| --- | --- | --- | --- | --- | --- | --- |
| execution | Sonnet / high | GPT-6.1 Sol / medium | GPT-6.1 Sol / medium | Inherit | GPT-6.1 Sol via GitHub Copilot / medium | Inherit |
| reasoning | Opus / high | GPT-6.1 Sol / high | GPT-6.1 Sol / high | Inherit | GPT-6.1 Sol via GitHub Copilot / high | Inherit |
| inherit | Inherit / high | Inherit / high | Inherit / high | Inherit | DeepSeek v4.1 Flash / max | Inherit |
| verifier | Opus / high | GPT-6 Astra / high | xai/grok-4.7 / high | Grok 4.6 (no effort) | github-copilot/claude-opus-5.5 / high | Inherit |

The `verifier` model profile, used by `hive-verify-task` (change `verifier-model-and-validation-handoff`, 2026-09-29), is distinct from the `verify` access profile, which the same role also uses and which only selects permissions. `verifier` is optional when reading profiles, so releases frozen with three model profiles remain readable, and required in the repository profiles; resolving a role that asks for it against a host without it fails with an error naming the host and the profile. Catalog evidence, 2026-09-29: `pi --list-models` lists `xai grok-4.7`; `opencode models` lists `github-copilot/claude-opus-5.5`; on 2026-09-30 the local catalog cache (`~/.cache/opencode/models.json`) listed `high` among that model's `reasoning_options` efforts — the cache declares efforts under `reasoning_options`, not under a `variants` key — and the running service decoded the installed `hive-verify-task` role's `#high` reference into variant `high`, so only an observed run of that role on that model remains unverified; the Codex model comes from Codex's `models_cache.json`. Pi finding: its resolver (`@earendil-works/pi-coding-agent` `core/model-resolver.js` lines 447-454) accepts a model id missing from its catalog when the provider is known, warning "Using custom model id", so `openai-codex/gpt-6.1-sol` stays in Pi's `execution` and `reasoning` profiles although `pi --list-models` does not list it.

OpenCode renders the configured base model and effort as `provider/model#variant`: its execution roles use `github-copilot/gpt-6.1-sol#medium`, its reasoning roles use `github-copilot/gpt-6.1-sol#high`, and role exceptions replace that suffix. The local OpenCode 2.0.19 catalog listed that model and its `low`, `medium`, `high`, `xhigh`, and `max` variants on 2026-09-29. Its `inherit` profile remains the fixed `opencode-go/deepseek-v4.1-flash#max`; role exceptions do not change a legacy profile that already embeds a variant and declares no effort. The GitHub Copilot model depends on the user's plan and credentials. When it cannot run, `flow-build` leaves the task unverified and asks the user rather than substituting a child on the implementer's model. Before each launch `flow-build` compares the implementer's model with the verifier's, treats an unknown or session-inheriting model as the same, and when they match launches the verifier with another model through the launch's model option or asks the user once. Cursor's `verifier` profile inherits the session model, so there that comparison chooses the model at launch; Grok's `grok-4.6` is a configured model that the host may still override.

Cursor inherits the parent model in every profile: its model IDs depend on the subscription plan, and its documentation states that Cursor replaces a configured model the plan does not include. Choosing per-profile models is deferred until role delivery is observed. Cursor is optional in `agent-profiles.json` so that releases frozen with the five original hosts remain installable for them; installing Cursor from such a release fails with `unsupported agent host "cursor"`.

### Effort

Every profile that can carry an effort declares one, so no role inherits the session effort on Claude, Codex, Pi, or the configured OpenCode execution and reasoning profiles: a user lowering the session effort to stay in the loop would otherwise lower delegated children that nobody steers. A role declares `effort` only when its task needs a different level on every host that can represent it; the value is absolute and replaces the profile's. Current exceptions: `hive-review-security` `max`, where effort gains most for security work ([Spending your effort](https://claude.dev/blog/spending-your-effort/)); `hive-read-state` `low`; `hive-review-plan` and `hive-write-spec` `medium`.

| Host | Rendered key | Levels |
| --- | --- | --- |
| Claude Code | `effort` | `low`–`max`; the active model clamps an unsupported level to its highest supported one ([model configuration](https://code.claude.com/docs/en/model-config)) |
| Codex | `model_reasoning_effort` | `low`–`max`; behavior for a level the model does not support is undocumented, and `max` is not documented for GPT-6 Astra ([subagents](https://learn.chatgpt.com/docs/agent-configuration/subagents)) |
| Pi + pi-subagents | `thinking` | `off`–`max` (0.67.0 `src/shared/model-info.ts`); the level is appended to the resolved model without a capability check, so provider handling is unverified |
| OpenCode V2 | `model: provider/model#variant` | `low`–`max`; generated from the profile effort, with role exceptions replacing the variant only for profiles that declare a base model and effort |
| Grok, Cursor | none | Role exceptions have no effect. Grok has no effort field, so children inherit the session effort; Cursor binds effort to a concrete model, which Hive leaves as `inherit`, with the same result |

A level missing from a host's scale would render as that host's highest documented level; the configured Claude, Codex, Pi, and OpenCode profiles use `low`–`max`. The mapping follows the host's scale, not the model's: a role on the `inherit` profile, such as `hive-review-security`, runs on the session's model, whose support for `max` the renderer cannot know. Codex and Pi use GPT-6.1 Sol at `medium` for execution and `high` for reasoning; their `verifier` profiles use GPT-6 Astra and Grok 4.7 at `high`. The configured Pi identifier remains subject to provider and account availability.

Profiles are delivery defaults, not a replacement for the host's authentication, model loop, or permissions. The canonical JSON contains the exact model identifiers. A configured model is not proof of account availability or a successful run.

`observe` selects native read-oriented defaults where available. `implement` retains host permissions; `verify` also retains them because builds and tests may write artifacts. Role instructions constrain verification to evidence rather than source fixes. These categories are not universal security sandboxes: shell tools, external services, parent overrides, and native policy resolution still matter. Claude parent auto, acceptEdits or bypassPermissions modes can override a child's plan mode. Codex reapplies live parent permission overrides to children; a parent started with bypass flags can supersede an agent's sandbox default. Grok's documented permission field does not demonstrate enforcement of every mode in the installed build.

## Delegation and instruction delivery

`content/guidance/global.md` owns the shared delegation contract. Activity skills own their specific decomposition and review procedure; role descriptions identify stable responsibilities. Project guidance and situational references supply technical conventions. Keep these sources canonical rather than maintaining a second role-by-stack routing table.

A role's native configuration does not establish that it received or read a skill. The caller supplies project context, decisions and applicable resource identities or accessible paths. The child locates skills through the host catalog/loader or an explicit task path and reads the required resource before dependent work. Missing context blocks only the affected dependency. Hive does not implement portable skill preloading or filesystem-permission overrides.

For example, a backend implementation assignment identifies the authorized API change, root, owned files, compatibility contract and acceptance checks, then points to the available `flow-build/SKILL.md` and existing project conventions for the detected stack. A read-only reviewer receives the review question and relevant evidence instead of implementation authority. Neither assignment requires creating a new agent for its framework.

Role links use `skill:owner/path` logical resource identities, for example `skill:flow-plan/references/ui-planning.md`. They resolve from the named skill's discovered directory, not the native agent file. This is an authoring convention interpreted through instructions, not a new CLI URI handler. The release validator checks the owning skill and resource exist in the bundle; the renderer preserves the link without embedding content or a machine-specific path. See [instruction resources](instruction-resources.md) for authoring and validation limits.

This contract is authored guidance, not deterministic loading enforcement or a demonstrated quality improvement. The backend role covers the repository’s actual language and framework, including TypeScript. The former `ts-backend-developer` source is consolidated into `hive-build-backend`, retaining runtime validation, package/workspace inspection and shared-contract guidance. The former `react-developer` and `angular-developer` sources are consolidated into `hive-build-frontend`, with concise framework-conditional guidance for server/client boundaries, change detection, subscriptions, lifecycle and forms.

Retain separate implementation and review contracts: `hive-write-tests` authors tests while `hive-verify-change` reports independent verification without fixing source; design and operational implementation remain different assignments. Database, performance and Kotlin Multiplatform roles retain their distinct evidence requirements. Security, UX and refutation reviews are selected for the relevant risk or question, not an obligatory review chain; state collection is a bounded assignment, not a mandatory workflow stage. No per-language backend roles or additional stack routing registry are needed for this catalog.

## Native destinations

| Host | User directory | Format |
| --- | --- | --- |
| Claude Code | `<Claude home>/agents/` | Markdown with native YAML |
| Codex | `<Codex home>/agents/` | TOML |
| Grok Build | `<Grok home>/agents/` | Markdown with camelCase native YAML |
| Pi + pi-subagents | `<Pi agent home>/agents/` | Markdown with native YAML |
| OpenCode V2 | `<OpenCode home>/agents/` | Markdown with `mode: subagent` |
| Cursor CLI | `<Cursor home>/agents/` | Markdown with native YAML (`readonly: true` for `observe`) |

Existing project scope is retained for Claude and Codex, under `.claude/agents` and `.codex/agents`. Other hosts remain user-scope only in this manager.

Cursor documents user subagents in `~/.cursor/agents/`, `~/.claude/agents/`, and `~/.codex/agents/`, with `.cursor/` winning a name conflict ([Cursor subagents](https://cursor.com/docs/subagents), 2026-09-25). Hive roles share their names across these locations, so the Cursor copy should shadow the others; that no duplicates appear is inferred from this precedence, not observed. Cursor CLI 2026.09.18 listed no user-level subagents at all, so its sessions read the role contract from `<Cursor home>/agents/<role>.md` and pass it to a generic child, as the delegation rule in `global.md` says; that child is not restricted by the file's `readonly` field. Removing only the Cursor consumer while Claude stays installed would leave Cursor on the roles rendered for Claude in `~/.claude/agents/`. Pi requires its already configured `pi-subagents` extension; Hive does not install an extension implicitly. Pi uses append mode and explicitly inherits project context, global context, and skills. Its installed lightweight parser requires comma-separated tool exclusions rather than JSON/YAML flow arrays; descriptions use literal blocks to preserve escaping. Grok extends its native prompt and retains AGENTS discovery. OpenCode uses V2 `permissions`, never the legacy singular schema.

The manager owns only its recorded destinations. Unmanaged collisions and edits to managed content must remain conflicts, not reasons to overwrite user files. Private runtime recovery snapshots are transaction data, not additional authored agent sources.

## Evidence and limits

Inspected on 2026-09-21: Codex 0.155.1, Claude Code 2.1.278, Grok Build 1.0.40, Pi 0.86.1 with pi-subagents 0.67.0, and OpenCode 2.0.9. Native contract sources:

- [Codex subagents](https://learn.chatgpt.com/docs/agent-configuration/subagents)
- [Claude Code subagents](https://code.claude.com/docs/en/sub-agents)
- [Grok agent configuration](https://github.com/xai-org/grok-build/blob/main/crates/codegen/xai-grok-agent/README.md)
- [OpenCode V2 agents](https://opencode.ai/v2/docs/agents)
- Installed pi-subagents `docs/agents.md`, `docs/models.md`, and parser source.
- [Cursor subagents](https://cursor.com/docs/subagents), consulted 2026-09-25 for `cursor-agent` 2026.09.18-9a7762b; its installed bundle names the delegation tool `Task`.

Documentation establishes intended formats. Parser tests establish serialization and selected native loading properties. Filesystem status establishes installation only. CLI behavior pilots remain paused; none of those checks establish role selection, instruction adherence, or model quality.

## Saved-plan review

`hive-review-plan` uses the existing observe profile and returns feedback without edits or gate execution. `flow-plan` owns its domain selection and reconciliation procedure; one canonical role can serve multiple bounded backend, frontend/UI, infrastructure or other domain assignments. Native observe controls vary by host and do not establish universal read-only isolation. The orchestrator remains the sole plan writer.

## Native selection and evidence

The shared delegation contract lives in `content/guidance/global.md` as a capability-first rule: select the installed role through the child-launch tool's own selector, and when the tool does not list the role, give a bounded generic child the contract read from the installed role file and disclose the fallback. The per-host table that used to follow it was removed from the always-loaded block after [issue #30](https://github.com/JhonHawk/tricell-hive/issues/30): on 2026-09-25, all six hosts delegated the same way with and without it, one run per cell. The dialects below are maintainer reference, not distributed guidance; follow the session's actual tool schema.

| Host | Child selection observed |
| --- | --- |
| Claude Code | The exposed delegation tool (`Agent` or `Task`) with the installed role's `subagent_type`; 2.1.278 announced `Task` and executed `Agent`. |
| Codex | `spawn_agent` with `agent_type` set to the installed TOML role. Some hosted APIs offer model overrides without this role selector. |
| Grok Build | `spawn_subagent` with `subagent_type` when exposed; 1.0.39–1.0.41 dropped the parameter, so a discovered agent did not establish that the selector exists. |
| Pi + pi-subagents | `subagent` with `agent` set to the installed role ID. |
| OpenCode V2 | `subagent` with `agent` set to the configured agent ID; V1 used `Task`/`subagent_type`. |
| Cursor CLI | `Task` with `subagent_type`; 2026.09.18 and 2026.09.23 list only project-level and built-in types, so sessions fall back to a generic child with the role contract. |

`global.md` asks for a background launch of a child expected to outlast one stretch of the wait, because a foreground launch blocks the parent's turn and the stretch status line cannot be written ([issue #44](https://github.com/JhonHawk/tricell-hive/issues/44): a 27.8-minute foreground child on OpenCode, sample-project session `ses_f1e800deeffe`, 2026-09-27). Every host offers one; inspected 2026-09-30 from documentation and installed binaries, without model runs:

| Host | Background launch | Completion reaches the parent as |
| --- | --- | --- |
| Claude Code 2.1.286 | `Agent` with `run_in_background`; the default for interactive spawns since 2.1.278 | A notification in a later turn |
| Codex 0.159.1 | `spawn_agent` returns at once | `wait_agent` with a timeout; push delivery to a later turn is undocumented |
| Grok Build 1.0.45 | `spawn_subagent`; a foreground child over its time budget is moved to the background | An automatic notification, or `wait_commands_or_subagents` |
| Pi 0.87.1 + pi-subagents 0.67.0 | `subagent` with `async`, on by default | A message that triggers a new turn |
| OpenCode V2 2.0.20 | `subagent` with `background: true`; a running foreground call can be moved with `POST /api/session/{id}/background` | A synthetic parent message `<subagent sessionID=… state=…>` with `metadata.source: "subagent"`, which wakes the parent |
| Cursor CLI 2026.09.28 | Only per role, through `is_background` in the agent file | Undocumented; a `subagentStop` hook exists |

Non-interactive runs are unverified on every host. On OpenCode, `opencode run --format json` prints no synthetic message and returns when the parent finishes, so a child that completes later probably goes unreported.

Isolating a guidance variant per process for such pilots: Claude Code `--settings` with `claudeMdExcludes` plus `--append-system-prompt`; Codex and Pi a shadow `CODEX_HOME` or `PI_CODING_AGENT_DIR` with symlinks; Grok a shadow `HOME` with the real `GROK_HOME`; OpenCode 2 needs `opencode run --standalone`, because its background service ignores process environment overrides; Cursor a project `.cursor/rules` file.

Ids as observed, before the 2026-09-29 rename to `hive-<verb>-<object>`.

| Host inspected | Evidence through 2026-09-22 | Remaining limit |
| --- | --- | --- |
| Claude Code 2.1.278 | Agent(subagent_type: review-plan) observed; child Opus, role and references read | Child effort/mode not independently observed; parent mode can override plan |
| Codex 0.155.1 | spawn_agent(agent_type: review-plan) observed; child Astra/medium despite Terra parent | Ephemeral pilot parent failed; persisted parent worked. Parent bypass overrode read-only sandbox |
| Grok Build 1.0.38 / 1.0.40 | 1.0.38 interactive session selected frontend-developer natively with grok-4.6; 1.0.40 pilots exposed no subagent_type | review-plan has not been directly tested on 1.0.38; version and launch surface differ between observations |
| Pi 0.86.1 / pi-subagents 0.67.0 | subagent(agent: review-plan) observed; child Astra/medium read role and reference | write/edit exclusions do not exclude shell or memory tools; a memory write occurred in the isolated store |
| OpenCode V2 2.0.9 | Native call and child session record confirm review-plan with DeepSeek/max | edit deny is configured, not mutation-tested; no universal write isolation |
| Cursor CLI 2026.09.18 | Project-level subagents listed; user-level ones (`~/.claude/agents`) not listed | Rendered Cursor roles, their precedence, and the generic-child fallback are not yet observed |

Distinguish installed bytes, catalog discovery, actual native selection, resource reading and effective controls. Preserve these observations in the existing task/review record when relevant; no additional ledger is required. Ordinary packaging tests do not establish model compliance.

The bounded dispatch verification (historical evidence omitted from public history) records actual calls, effective models, pilot-runner corrections and memory hygiene. These observations do not establish reliability or model quality.

## Native user questions

Codex 0.157.0 (VS Code, Default mode) asks through `request_user_input_async`, which returns before the answer, or `send_user_message_question`; the synchronous `request_user_input` is not available in `codex exec`. Observed on 2026-09-25, not documented: a question answered while the turn was still open reached the user (7 cases), but a question followed by the end of the turn before any answer was not visible, and the user asked for it again. Evidence: `<private-local-session-evidence>` lines 1727–1748 and `<private-local-session-evidence>` lines 4060–4231. The same sample-project session later repeated a pending close question in text before ending the turn, and the user answered by typing its option ID (lines 6599–6634). The global guidance therefore keeps the turn working while a question is unanswered and repeats a pending question with its options when the turn ends first. The pilot trace model cannot express this case yet ([#34](https://github.com/JhonHawk/tricell-hive/issues/34)).

## Presentation capabilities

The global guidance names presentation capabilities rather than host tools: labeled Markdown links, reports shaped by content, ASCII diagrams in the chat with Mermaid and SVG kept out, and the native task list rule in `flow-build`. This table records what each host offers, checked on 2026-09-26 against the installed versions (change `report-readability`). Evidence labels: **doc** official documentation or changelog, **obs** observed locally (binary, `--help`, or a user session), **inf** inferred from code structure.

| Host · version | Task list | Native question | Labeled links | Mermaid in chat | Publishable page |
| --- | --- | --- | --- | --- | --- |
| Claude Code 2.1.283 | `TaskCreate`/`TaskUpdate` (doc) | `AskUserQuestion` with per-option `preview` (doc) | Label shown and clickable in the user's session (obs, 2026-09-26); issues [#26390](https://github.com/anthropics/claude-code/issues/26390) and [#37808](https://github.com/anthropics/claude-code/issues/37808) claiming raw URLs are closed not-planned and duplicate | No (doc: not found) | `Artifact`, paid plans only (doc) |
| Codex CLI 0.157.0 | `update_plan` (obs, matches [base instructions](https://github.com/openai/codex)) | `request_user_input_async` or `send_user_message_question` in the TUI (obs, section above); not in `codex exec` | OSC 8 with label; base prompt prefers clickable Markdown links (obs) | No | No |
| Grok Build 1.0.38 | `TodoWrite` and Plan Mode (obs) | `ask_user_question` with focused-option preview (obs) | OSC 8 for http, https, mailto when the terminal supports it (obs) | Yes, `ui.render_mermaid` (doc, obs) | No |
| OpenCode v2.0.18 | None; v2 removed the todo tool (doc) | None found | Label when hyperlinks are negotiated, otherwise `label (url)` (inf) | No | No |
| Cursor Agent 2026.09.23 | `todo_write` (obs) | Question tool in Plan, Ask, and Debug modes (doc) | OSC 8 tokenizer in the renderer (obs, inf for label handling) | Yes, as ASCII ([changelog](https://cursor.com/changelog/cli-feb-18-2026), doc) | No |
| Pi 0.87.1 | None built in; extension example only (obs) | None built in; `ctx.ui.select` for extensions (doc) | OSC 8, `PI_HYPERLINKS` override (doc) | Yes, `markdown.mermaid` default `streaming` (doc) | No |

Mermaid and SVG stay out of the chat because the model cannot tell whether its host renders them and three of six show the source. An ASCII diagram in a code block needs no rendering, so it reads the same on every host; the 2026-09-26 rule banned all chat diagrams and with them the ASCII diagrams the previous Hive drew by default, which the user asked to restore on 2026-09-29. The 70-character width keeps a diagram from wrapping in a narrow terminal. Markdown table rendering also varies: Codex's base prompt discourages tables unless requested, and OpenCode issues report misaligned tables, so the guidance limits tables to comparisons with at most four short columns. No pilot measured these rules; later session monitoring is the evidence.
