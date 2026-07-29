---
name: workspace-conventions
description: >
  Codex/opencode: load before creating/naming artifacts outside app source (_support,
  plans, specs, ADRs, contracts, evidence) or answering "what's next"/offering /flow-*
  with a ledger. Skip Claude Code (always-on rules). Triggers: _support/, *-specs, flow.
---

# workspace-conventions — router to workspace, session, and contract conventions

The full canonical rules live in `references/` (injected at build time from
`global/rules/workflow/` — single source of truth). This skill exists because Codex and
opencode have no conditional channel for intent-keyed policy: the always-on floor keeps
one trigger line; the complete conventions load here, when the situation is actually in
play.

## Routing table — read every row that matches the situation

| Situation | Read |
|---|---|
| Creating/moving/naming any artifact outside app source; `_support/` vs specs-repo routing; sessions; infra repo placement | `references/project-structure.md` |
| Generated-artifact naming/grouping, retention, evidence curation, versioning, legacy folder mappings (anything under `_support/`) | `references/support-artifacts.md` |
| New or changed cross-service contract (endpoint a frontend consumes, request/response shape between services, events/webhooks) | `references/cross-service-workflow.md` |

## Flow phase boundaries (harness addendum — not restated in the references)

- Actions owned by another convention get a pointer to it, never their execution plan
  inline: environment promotion and post-deploy verification follow the git-workflow
  promotion gates and `flow-core/references/promotion-playbook.md` — surface the next
  promotion as a recommendation the user confirms, never hand it back as an opaque
  user to-do.
- Suggestion surfaces (next steps, scope candidates, recommendations) obey the same
  boundaries: when a skill owns the action, the suggestion is the skill invocation
  phrased as an offer to run it, not the plan. **Every such offer names the direct
  route as its alternative — and when the current phase's artifact already holds what
  execution needs, the direct route goes first.** Never phrase the next skill as the
  only or "correct" next step: chained offers with no branch turn a small change into a
  full pipeline one acceptance at a time.
- In flow workspaces, an approved plan is a session artifact: it lives at
  `sessions/YYYY-MM-DD-<slug>/<slug>-plan.md` with `Status: planned` (a plan carrying
  `Session: no` declines the folder), the ledger `## Current handoff` is updated when
  artifacts are produced, and committing session artifacts at close is
  standing-authorized (`chore(sessions): <slug>`). `/flow-build` adopts and executes any
  session plan.

## Rules of use

- Route by what is ON DISK (a `_support/` folder, a `<project>-specs` sibling, a
  `_support/PROJECT.md` ledger), never by whether the prompt mentions the flow pack.
- A reference already loaded this session does not need reloading.
- No matching situation → this skill has nothing for the task; proceed without it.
