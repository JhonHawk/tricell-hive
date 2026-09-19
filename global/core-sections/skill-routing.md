---
order: 170
targets: [claude]
---

### Skill Auto-invocation

**Situational policy lives behind a router skill, so not invoking it is the same as not having the rule.** Load the matching router BEFORE acting — not after, and not "if it turns out to be needed". If a router plausibly covers the situation, read it; being wrong costs one read, skipping it costs the rule. These thoughts mean the check is being rationalized away, not that it is unnecessary:

| Thought | Reality |
|---|---|
| "This is a simple edit" | Simple edits are where conventions get silently broken. |
| "I already know this convention" | Conventions change and are per-project. Read the current one. |
| "I'll check the convention after writing it" | Then the wrong name is already in a migration. |
| "I read that rule earlier in the session" | Fine — a reference already loaded is not reloaded. |
| "The task is too small to route" | Size decides delegation, never whether the rule applies. |

Consulting the router is never the blocking step: read it and keep going in the same turn.

**No rule loads by file kind — read it before writing.** The `rule-delivery` hook HOLDS the first write matching a rule and names the file to read: read it, re-issue the call, it passes. Reading it up front avoids the hold, and nothing fires before the file exists (naming identifiers, choosing a path) — so before creating the first file of a kind in a session (`.tsx`, `.py`, `.tf`, a migration, a Dockerfile), read the matching `~/.claude/skills/language-rules/references/` file. Rules already in context need no re-read; the hold **re-arms after a compaction or `/clear`**, so treat the rule as absent and read it again.

**The routers and the act that fires each one.** Nothing else loads them; a router not invoked is a rule you do not have.

| Router | Fires on — the observable act | What it governs |
|---|---|---|
| `task-routing` | the first `Write`/`Edit` on project code OR the first `git` command of the session — whichever comes first — or before the first delegation | who takes the task, delegation gates, plan gap analysis; git mechanics: branch, session mode, commit semantics, PRs, promotion, close |
| `memory-policy` | the first `mem_*` call of the session, and the close-time summary | project identity, save cadence, invalidation, tracker sync |
| `workspace-conventions` | writing a file outside application source, typing an infra resource name, or stating in an answer/plan where an artifact, script, report, or doc will live (a path or folder named in prose is the act) | `_support`, specs, ADRs, contracts, naming, cross-service shapes |
| `status-fetch` | about to answer "what's pending / where are we" without having read git yet | live external state |
| `language-rules` | Grok — the first `Write`/`Edit` of code; Claude Code — about to drive a browser (language rows arrive through the `rule-delivery` hold on the first matching write) | full language conventions; `browser-automation-reference.md` |

`flow-report` is not in the table: it is a renderer, not a router, and its trigger is a property of the answer rather than an act of yours — `rules/quality/communication-format.md` is canonical for it and the conditions are never restated elsewhere.

The gates those routers' domains carry stay always-on and need no skill: what a git verb authorizes, protected branches, force-push and production promotion live in `git-workflow.md`.

**A trigger is written as an act, never as an intent.** The test: could a third party reading the transcript say whether the moment happened? "At edit-intent" fails it — it needs introspection, and a model that does not recognize the moment never reaches the rule. "The first `Write`/`Edit` on project code" passes: it is in the log. This governs the table above, every `description:` in a skill's frontmatter, and any rule whose trigger is a moment rather than a file.

- **`/simplify` (Claude Code built-in) may auto-invoke at a change-group's green seam** — affected tests passing, before the commit/diff-presentation boundary — scoped to the just-changed code, and declared when run. Quality cleanup only (reuse, simplification, efficiency, altitude); never a substitute for review. It edits the working tree — its edits are part of the change-group and reach the user through the same commit/diff gate as the rest. Claude Code only: Codex/Grok/opencode have no such skill — the same four dimensions reach them through the `review-code` agent.

- **User-gated skills are OFFERED, never invoked.** A `disable-model-invocation` skill (`/flow-plan`, `/flow-build`, `/deploy-global`, `/adversarial-research`, `/agents-md-primary`) is never executed uninvited. `flow-core` and `flow-report` are pack infrastructure, not gated commands; `/memory-sync` is deliberately ungated — model-invocable where its owning rules command it, offered (not run) otherwise.
- **Process knowledge is not a command.** The specs, bootstrap, migration, audit and workspace-hygiene procedures live as playbooks under `flow-core/references/`; portable planning is the explicit `/flow-plan` skill. Apply procedures conversationally and proportionally to the change; never announce a stage or manufacture an artifact to justify one.
- **Nothing promotes a task into process.** File count, changed lines, subagent count, or perceived risk never select a plan, a spec, a `_support/` artifact, or a flow command — the user's explicit request does. Risk escalates *verification* (fresh reviewer, refuters — `agent-routing.md`), never process. **Carve-out — the direct route's task record** (`agent-routing.md > Direct Route`): working state in gitignored `_support/workspace/tasks/`, created when the exploration enumerates 2+ implementation steps, never approved, never a plan.
- **Offering `/flow-plan` or `/flow-build`.** Offer `/flow-plan` once, in prose, when no portable contract exists and the user either asked to plan or, in a flow workspace (ledger at or above cwd), handed over a batch of tracker tickets or a multi-task objective; a new batch reopens it. Offer `/flow-build` when an approved contract with implementation authority and pending tasks exists. Name the direct route as the alternative — first when the plan already holds what execution needs. **The offer precedes any executor dispatch or tracker state change — never follows one;** a no takes the direct route, which still needs the explicit implementation verb. An offer selects nothing; a no is sticky for that batch.
- **Never offer a production promotion or `/deploy-global` as a next step.** Production promotions are git-workflow gates that always confirm; `/deploy-global` is the user's explicit call.
