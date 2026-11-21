---
title: Skills
description: Find Hive's main flows and supporting skills for the work at hand.
---

Hive's skills supply procedures when a task needs them. The agent selects the
relevant skill; you can also name one in your request. The linked source files
hold the canonical rules.

## Main flows

For where each flow loops and where a requested change returns the work, see the [changes and loops overview](/flows/overview/).

- [Status](/flows/status/), a playbook of Hive mode: report the backlog and unfinished work, read-only, and recommend what to take next.
- [Hive mode](/flows/hive-mode/): the entry for every request to do work. It settles what you need, opens at most one of 13 playbooks for the kind of work, and also investigates a question with evidence or audits the codebase for what to improve.
- [Create design](/flows/create-design/): draw a new screen in Figma, correct frames you name, review a design someone else drew, or create or refresh `DESIGN.md`.
- [Plan](/flows/plan/): settle consequential decisions and define verifiable work.
- [Build](/flows/build/): implement authorized changes and hand over a built candidate.
- [Verify](/flows/verify/): check the candidate at the level you set, then integrate it.
- [Security review](/flows/security/): review a change's security boundaries, recommended when a change touches one.
- [Close](/flows/close/): record the outcome and clean up the work session.
- [Deploy](/flows/deploy/): promote an integrated change and check what is served.
- [Create issue](/flows/create-issue/): write a ticket someone else can work, break work down into a parent ticket with children, or check that a ticket is ready.
- [Create PRD](/flows/create-prd/): write the requirement document of a module or feature, restructure one someone else wrote, or check it against the code.

## Rule leaves

A rule that applies at one moment, such as staging a commit, moving a ticket or asking a round of questions, lives in a small file called a rule leaf, stored in the hidden `hive-core` skill (`references/<name>.md`). You do not pick one and the agent does not select one by its description. The [shared rules](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md) index every leaf under «Rule leaves», each with the moment that requires it, and the flow steps link the leaf where it applies. At that moment the agent reads the file and follows it. For example, the [stage and commit leaf](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/git-commit.md) is read before the first `git add` or commit, and the [questions leaf](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/questions.md) before a round of questions.

The `hive-core` skill is marked so a host does not offer it for automatic selection. Claude Code reads that mark from the skill; for Codex the manager generates a policy file, `agents/openai.yaml`, beside it; Antigravity lists it in its catalog with a one-line description. An update retires the earlier one-skill-per-leaf folders. For the other hosts Hive has not verified a hiding mechanism, so they may list `hive-core` too. See [Host limits](/host-limits/) for what each host does no matter what a skill says.

To save reads, eight flows read one stage packet when they start: a file the installer builds from the leaves that flow needs at that stage, such as the verification floor, the checks and the commit rules for Build. The agent reads the packet whole and does not read those leaves again; when the packet is missing, it reads the leaves the flow links. The leaves stay the only place where each rule is written.

Code principles are leaves too. The [principles index](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-principle-index/SKILL.md) lists each one with the moment it applies, and the build and code review steps link it. The [TypeScript skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-typescript/SKILL.md) is the exception: it stays visible to the host. Claude Code and Cursor list it only once the agent works with a `.ts` or `.tsx` file; the other hosts list it from the start. On every host the agent decides from its description whether to read it, and the principles index points to it during a build.

## Supporting skills

- [Shared rules (global guidance)](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md): it is always loaded and every flow uses it. It keeps the boundaries that must hold in every session (authorization, preservation, secrets, artifact placement, delivery modes) and indexes the rule leaves for the rest, such as how the agent asks you a decision, handles a change you request mid-work and reads the repository's `## Hive` settings.
- [Git workflow](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/git-workflow/SKILL.md): prepare authorized Git delivery while preserving unrelated work and scanning target paths before committing.
- [Visual reports](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-report/SKILL.md): present a report, comparison, or explainer as a designed page.
- [Starlight documentation sites, a mode of Visual reports](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-report/references/docs-site.md): scaffold, extend, or audit maintained documentation sites.
- [Harness audit](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/harness-audit/SKILL.md): audit agent instructions, skills, and role definitions, and the memory each host loads by where a session starts, including host memory and Engram records that restate the rules of Hive or the project, ask the decisions the findings need, and apply what you accept in the current project, including creating or updating a repository's `## Hive` section.
- [Local instructions, a mode of Harness audit](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/harness-audit/references/local-instructions.md): create or prune the project's `HIVE.local.md` and the machine file `~/.config/hive/HIVE.local.md` from what already worked in recent sessions, after asking you which facts to keep.
- [Security advisories, a playbook of Hive mode](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-mode/playbooks/dependency-security.md): check whether a published advisory affects your repositories, without changing them.
- [Test audit](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/test-audit/SKILL.md): trim one area's tests to one owning test per behavior, removing duplicates and tests tied to how the code is written, and prove with a separate review and deliberate code changes that nothing lost its only test.
- [UI criteria, a reference of Create design](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-design/references/ui-criteria.md): find a project's component library, its installed version and its customization layer, keep them in a verified `DESIGN.md`, and give design, build and review work measurable screen criteria that never override the project's own design system.
- [Walkthrough, a rule leaf of Verify](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/walkthrough.md): publish UI captures and their iteration to the [viewer](/viewer/), and walk a built screen for you to validate.
- [Project layout](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/project-layout/SKILL.md): say where a project's documentation, scripts and data belong, check a repository, a workspace or a monorepo against that convention, and clean up, reorganize or migrate it from a plan you confirm, with or without Hive. Where each script, Terraform file or old evidence folder goes is computed the same way on every run, the references to what moves are rewritten for you, and a move can be undone exactly.
- [Workspace repo](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/workspace-repo/SKILL.md): on your request, turn a workspace of several repositories into one coordinator repository that versions what they share, with each repository still its own clone inside it.
- [Unattended delegation](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/unattended-delegation/SKILL.md): keep bounded work going after you delegate it, say you are leaving, or go quiet once you have answered the route question and each delivery and review value is answered; feedback within the goal keeps it going, and the delivery questions still missing when you delegate or leave are asked at once without stopping the work, which stops at a verified local result without your answers. When it starts, it also states the amount you gave for unforeseen spend, one sum for a cost no answer covers, or asks for it in one line; without an amount it is zero, and the agent skips that spend and reports it. Within the amount, it takes only an effect it can undo with a command recorded first, lists each one in the final report, and skips the effects of a class you refused in the plan.
- [Engram workspace initialization, a mode of Harness audit](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/harness-audit/references/engram-workspace.md): establish an explicit workspace memory identity when requested.

For what each coding CLI limits no matter what the skills say, see [Host limits](/host-limits/).
