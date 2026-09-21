# Canonical agents and native delivery

Agent instructions have one source at `content/agents/<category>/<name>.md`. Categories organize authorship; all native destinations use the unique filename stem. Do not keep converted host copies in the repository or a tracked distribution directory.

`integrations/agents` parses a deliberately small frontmatter contract: `name`, `description`, `model_profile`, and `access_profile`; optional `claude_effort` preserves a role-specific Claude setting. Values are single-line strings, optionally JSON-quoted. The body is preserved. Unknown fields, malformed sources, and unknown profiles fail validation. `integrations/agent-profiles.json` is the single host mapping. The deployment manager converts in memory during planning and freezes the native bytes for apply and recovery.

## Roles and profiles

Eighteen roles cover design, implementation, documentation, operations, verification, and review. `solution-architect` includes contract design formerly assigned to `sdd-design`; `review-security` includes detection formerly assigned to `secrets-auditor`. Dedicated `sdd-spec-reviewer`, `sdd-product-critic`, `workspace-custodian`, and `prompt-engineer` remain deferred. The reference checkout is preserved.

| Profile | Claude | Codex | Pi | Grok | OpenCode V2 |
| --- | --- | --- | --- | --- | --- |
| execution | Sonnet | Terra / high | Terra / high | Inherit | DeepSeek v4.1 Flash / max |
| reasoning | Opus / high | Astra / medium | Astra / medium | Inherit | DeepSeek v4.1 Flash / max |
| inherit | Inherit / high | Inherit / high | Inherit / high | Inherit | DeepSeek v4.1 Flash / max |

Profiles are delivery defaults, not a replacement for the host's authentication, model loop, or permissions. The canonical JSON contains the exact model identifiers. A configured model is not proof of account availability or a successful run.

`observe` selects native read-oriented defaults where available. `implement` retains host permissions; `verify` also retains them because builds and tests may write artifacts. Role instructions constrain verification to evidence rather than source fixes. These categories are not universal security sandboxes: shell tools, external services, parent overrides, and native policy resolution still matter. Codex reapplies live parent permission overrides to children; a parent started with bypass flags can supersede an agent's sandbox default. Grok's documented permission field does not demonstrate enforcement of every mode in the installed build.

## Delegation and instruction delivery

`content/guidance/global.md` owns the shared delegation contract. Activity skills own their specific decomposition and review procedure; role descriptions identify stable responsibilities. Project guidance and situational references supply technical conventions. Keep these sources canonical rather than maintaining a second role-by-stack routing table.

A role's native configuration does not establish that it received or read an activity skill. The caller supplies accessible source paths in the assignment; the child reads the applicable sources before dependent work and reports missing required context. The current renderer does not implement a portable skill-preload mechanism. Native inheritance settings are host-specific, not a substitute for this handoff.

For example, a backend implementation assignment identifies the authorized API change, root, owned files, compatibility contract and acceptance checks, then points to the available `flow-build/SKILL.md` and existing project conventions for the detected stack. A read-only reviewer receives the review question and relevant evidence instead of implementation authority. Neither assignment requires creating a new agent for its framework.

Database and security assignments also supply the accessible `flow-build` skill directory: those roles read its `references/data-changes.md` or `references/security-boundaries.md` when applicable. These paths are relative to that skill directory, not the rendered native agent file. Shared procedures stay in the skill bundle; the renderer neither copies them into each role nor resolves them at runtime.

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

Existing project scope is retained for Claude and Codex, under `.claude/agents` and `.codex/agents`. Other hosts remain user-scope only in this manager. Pi requires its already configured `pi-subagents` extension; Hive does not install an extension implicitly. Pi uses append mode and explicitly inherits project context, global context, and skills. Its installed lightweight parser requires comma-separated tool exclusions rather than JSON/YAML flow arrays; descriptions use literal blocks to preserve escaping. Grok extends its native prompt and retains AGENTS discovery. OpenCode uses V2 `permissions`, never the legacy singular schema.

The manager owns only its recorded destinations. Unmanaged collisions and edits to managed content must remain conflicts, not reasons to overwrite user files. Private runtime recovery snapshots are transaction data, not additional authored agent sources.

## Evidence and limits

Inspected on 2026-09-21: Codex 0.155.1, Claude Code 2.1.278, Grok Build 1.0.40, Pi 0.86.1 with pi-subagents 0.67.0, and OpenCode 2.0.9. Native contract sources:

- [Codex subagents](https://learn.chatgpt.com/docs/agent-configuration/subagents)
- [Claude Code subagents](https://code.claude.com/docs/en/sub-agents)
- [Grok agent configuration](https://github.com/xai-org/grok-build/blob/main/crates/codegen/xai-grok-agent/README.md)
- [OpenCode V2 agents](https://opencode.ai/v2/docs/agents)
- Installed pi-subagents `docs/agents.md`, `docs/models.md`, and parser source.

Documentation establishes intended formats. Parser tests establish serialization and selected native loading properties. Filesystem status establishes installation only. CLI behavior pilots remain paused; none of those checks establish role selection, instruction adherence, or model quality.
