---
name: workspace-conventions
description: >
  Load BEFORE creating/naming artifacts outside app source (_support, plans, specs, ADRs,
  contracts, evidence, sessions), naming any infra resource (bucket, cluster, service,
  security group, DB, subdomain, env branch), designing a cross-service contract, or
  answering "what's next"/offering /flow-* with a ledger. Every harness, Claude Code
  included: these rules are path-scoped there and load only after a matching file is read,
  which is usually too late to pick a name or a location.
  Triggers: _support/, *-specs, flow, IaC, naming, contract.
---

# workspace-conventions — router to workspace, session, contract, and naming conventions

This skill exists because Codex and Grok have no conditional channel for intent-keyed
policy: their always-on floor keeps one trigger line and the complete conventions load
here, when the situation is actually in play.

**Claude Code needs it too now.** `project-structure`, `session-capture`, `infra-naming`
and `cross-service-workflow` carry `paths:` there, so they load only once a matching file
is READ — and naming an infra resource or choosing where an artifact goes usually happens
BEFORE any such file is open. That is the gap this skill covers: invoke it when the
situation applies, not when a file happens to match.

References are injected at build time from `global/rules/workflow/` into `references/`.

## Routing table — read every row that matches the situation

| Situation | Read |
|---|---|
| Creating/moving/naming any artifact outside app source; `_support/` vs specs-repo routing; infra repo placement | `project-structure.md` |
| Which subfolder it lands in; session folders (`sessions/YYYY-MM-DD-<slug>/`), execution-vs-intention, raw-out-of-git, scripts and plans placement | `session-capture.md` |
| Generated-artifact naming/grouping, retention, evidence curation, versioning, legacy folder mappings (anything under `_support/`) | `support-artifacts.md` |
| New or changed cross-service contract (endpoint a frontend consumes, request/response shape between services, events/webhooks) | `cross-service-workflow.md` |
| Naming ANY infra resource — bucket, cluster, ECS service, security group, DB, secret path, subdomain, env branch, repo — or adding one to a project's naming table | `infra-naming.md` |

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
  `_support/PROJECT.md` ledger, `*.tf`/`Dockerfile`), never by whether the prompt mentions
  the flow pack.
- Naming is the row to read EARLY: a wrong infra name costs a recreate + migrate, not an
  edit — read `infra-naming.md` before proposing the name, not after creating it.
- A reference already loaded this session does not need reloading.
- **Every file in the table above lives in THIS skill's `references/`.** Rules cited by name
  from inside those files usually belong to another skill — read them there, never under
  this one: `memory-routing.md` → `memory-policy/references/`, language and framework rules
  → `language-rules/references/`, `unattended-autonomy-mode.md` →
  `unattended-delegation/references/`. A path guessed under the wrong skill fails silently
  and the answer proceeds without the rule.
- No matching situation → this skill has nothing for the task; proceed without it.
