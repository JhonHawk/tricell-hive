# Canonical agents and native delivery

Agent instructions have one source at `content/agents/<category>/<name>.md`. Categories organize authorship; all native destinations use the unique filename stem. Do not keep converted host copies in the repository or a tracked distribution directory.

`integrations/agents` parses a deliberately small frontmatter contract: `name`, `description`, `model_profile`, and `access_profile`; optional `claude_effort` preserves a role-specific Claude setting. Values are single-line strings, optionally JSON-quoted. The body is preserved. Unknown fields, malformed sources, and unknown profiles fail validation. `integrations/agent-profiles.json` is the single host mapping. The deployment manager converts in memory during planning and freezes the native bytes for apply and recovery.

## Roles and profiles

Twenty roles cover design, implementation, documentation, operations, verification, and review, including harness-engineering audits by `review-harness`. `solution-architect` includes contract design formerly assigned to `sdd-design`; `review-security` includes detection formerly assigned to `secrets-auditor`. Dedicated `sdd-spec-reviewer`, `sdd-product-critic`, `workspace-custodian`, and `prompt-engineer` remain deferred. The reference checkout is preserved.

| Profile | Claude | Codex | Pi | Grok | OpenCode V2 | Cursor CLI |
| --- | --- | --- | --- | --- | --- | --- |
| execution | Sonnet | Terra / high | Terra / high | Inherit | DeepSeek v4.1 Flash / max | Inherit |
| reasoning | Opus / high | Astra / medium | Astra / medium | Inherit | DeepSeek v4.1 Flash / max | Inherit |
| inherit | Inherit / high | Inherit / high | Inherit / high | Inherit | DeepSeek v4.1 Flash / max | Inherit |

Cursor inherits the parent model in every profile: its model IDs depend on the subscription plan, and its documentation states that Cursor replaces a configured model the plan does not include. Choosing per-profile models is deferred until role delivery is observed. Cursor is optional in `agent-profiles.json` so that releases frozen with the five original hosts remain installable for them; installing Cursor from such a release fails with `unsupported agent host "cursor"`.

Profiles are delivery defaults, not a replacement for the host's authentication, model loop, or permissions. The canonical JSON contains the exact model identifiers. A configured model is not proof of account availability or a successful run.

`observe` selects native read-oriented defaults where available. `implement` retains host permissions; `verify` also retains them because builds and tests may write artifacts. Role instructions constrain verification to evidence rather than source fixes. These categories are not universal security sandboxes: shell tools, external services, parent overrides, and native policy resolution still matter. Claude parent auto, acceptEdits or bypassPermissions modes can override a child's plan mode. Codex reapplies live parent permission overrides to children; a parent started with bypass flags can supersede an agent's sandbox default. Grok's documented permission field does not demonstrate enforcement of every mode in the installed build.

## Delegation and instruction delivery

`content/guidance/global.md` owns the shared delegation contract. Activity skills own their specific decomposition and review procedure; role descriptions identify stable responsibilities. Project guidance and situational references supply technical conventions. Keep these sources canonical rather than maintaining a second role-by-stack routing table.

A role's native configuration does not establish that it received or read a skill. The caller supplies project context, decisions and applicable resource identities or accessible paths. The child locates skills through the host catalog/loader or an explicit task path and reads the required resource before dependent work. Missing context blocks only the affected dependency. Hive does not implement portable skill preloading or filesystem-permission overrides.

For example, a backend implementation assignment identifies the authorized API change, root, owned files, compatibility contract and acceptance checks, then points to the available `flow-build/SKILL.md` and existing project conventions for the detected stack. A read-only reviewer receives the review question and relevant evidence instead of implementation authority. Neither assignment requires creating a new agent for its framework.

Role links use `skill:owner/path` logical resource identities, for example `skill:flow-plan/references/ui-planning.md`. They resolve from the named skill's discovered directory, not the native agent file. This is an authoring convention interpreted through instructions, not a new CLI URI handler. The release validator checks the owning skill and resource exist in the bundle; the renderer preserves the link without embedding content or a machine-specific path. See [instruction resources](instruction-resources.md) for authoring and validation limits.

This contract is authored guidance, not deterministic loading enforcement or a demonstrated quality improvement. The backend role covers the repository’s actual language and framework, including TypeScript. The former `ts-backend-developer` source is consolidated into `backend-developer`, retaining runtime validation, package/workspace inspection and shared-contract guidance. The former `react-developer` and `angular-developer` sources are consolidated into `frontend-developer`, with concise framework-conditional guidance for server/client boundaries, change detection, subscriptions, lifecycle and forms.

Retain separate implementation and review contracts: `test-engineer` authors tests while `sdd-verify` reports independent verification without fixing source; design and operational implementation remain different assignments. Database, performance and Kotlin Multiplatform roles retain their distinct evidence requirements. Security, UX and refutation reviews are selected for the relevant risk or question, not an obligatory review chain; state collection is a bounded assignment, not a mandatory workflow stage. No per-language backend roles or additional stack routing registry are needed for this catalog.

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

`review-plan` uses the existing observe profile and returns feedback without edits or gate execution. `flow-plan` owns its domain selection and reconciliation procedure; one canonical role can serve multiple bounded backend, frontend/UI, infrastructure or other domain assignments. Native observe controls vary by host and do not establish universal read-only isolation. The orchestrator remains the sole plan writer.

## Native selection and evidence

The shared delegation contract and per-host selection hints live in `content/guidance/global.md`. They are delivered together because Claude and Grok share the managed CLAUDE.md destination; separate host-specific block bytes would conflict there. These hints describe tool dialects, not a second responsibility router or executable adapter. Follow the session's actual tool schema before using a documented selector.

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
