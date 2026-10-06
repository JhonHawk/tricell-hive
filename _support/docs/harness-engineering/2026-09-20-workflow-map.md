# Hive workflow map

Review date: 2026-09-20. Status: documentary inventory of the previous Hive and proposals for the rebuild; it is not an implementation and does not reactivate its instructions.

## Purpose and scope

This document answers which jobs Hive covered, how they were carried out, and what we can learn from the reference repositories. A workflow is defined by its objective, inputs, activities, and closing evidence. A skill, an agent, or a command can take part in several workflows; none of them is a workflow by itself.

The main source is the local clone of `master` at `/path/to/reference-volume/dev-resources/tricell-hive-master`, commit `16e7d3357a3c41530d5e31460c3024872566f3c7`. The catalog describes what its sources prescribe; it does not guarantee that they ran correctly or that they are still installed. The external references were reviewed in their local checkouts, without updating remotes or testing integrations.

The rebuild rules and the user's authorization remain in force. The delegation, hook, publication, and structure requirements of the previous Hive are studied as reference material, not as instructions for this session.

## 1. The three entry routes

| Documented route | When it applies | Path | Output and limit |
|---|---|---|---|
| Research | Answering a question through code, documents, or external sources | Delimit the question → inspect evidence → cross-check → answer | A supported answer; a saved file only if requested. It implies neither implementation nor a formal plan. |
| Direct work | Obtaining an authorized result without selecting the formal development contract | Understand the deliverable and effects → explore what is needed → execute → check → close | A result proportional to the work. It may have a resumable record; it does not require a formal plan just because it is complex. |
| Formal development | Explicit request for `/flow-plan` and then `/flow-build`, or their equivalent | Explore → define → design/plan → implement → verify/review → deliver/close | A portable plan, with permissions and execution evidence kept distinct. Planning, implementing, and publishing are separate authorizations. |

Sources: [flow-research][h-research], [task-routing][h-routing], [common intake and direct route][h-agents], [flow-plan][h-plan], [flow-build][h-build].

Direct work also covers business/BI analysis, architecture proposals, and infrastructure operations. The type of deliverable determines the evidence needed; it does not turn these jobs into formal development. A brief query or a mechanical change does not need all of this structure.

## 2. Catalog of jobs Hive covered

The grouping below is a synthesis of the inventory, not a new list of mandatory commands.

| Job | Usual input | Essential sequence | Deliverable and closing criterion | Source in master |
|---|---|---|---|---|
| Investigate and compare | A question, claim, or pending decision | Pin down the question, look for evidence, cross-check contradictions and limits | An answer with sources; explicit uncertainty where evidence is missing | `flow-research`; `adversarial-research` as an explicit variant |
| Check status and pending items | An identified project or repository | Query Git, the declared tracker, PRs, and applicable deployments; reconcile observations | A dated status and verified pending items; querying does not decide what to run | `status-fetch` |
| Analyze data or propose architecture | Data, a business need, or constraints | Define metrics/assumptions or alternatives, analyze, and justify | Reproducible calculations or a proposal with trade-offs and open decisions | `agent-routing.md`, Common intake |
| Start a project | An idea, conversation, brief, or initial code | Refine requirements, resolve questions, set up the workspace and technical base as needed | Requirements, conventions, structure, and technical foundation; it does not mean deploying | `bootstrap-playbook.md` |
| Specify product and review business | Requirements or a scope change | Create/review the epic and product map, review rules and scenarios, derive tasks after the business gate | A specification and acceptance criteria; changes relative to what is already built are identified | `spec-writing-playbook.md` |
| Plan development | A bounded change that needs a formal contract | Explore files and contracts, resolve decisions/prerequisites, write tasks and verification, record approval | A plan with contract, authorization, and execution kept separate | `flow-plan` and `plan-format.md` |
| Implement and resume | A direct request or an authorized plan | Reconcile the real state, execute pending work, keep evidence, and resolve deviations | A bounded, verified change; when resuming, do not repeat work already applied | Direct route or `flow-build`, reconcile/execute |
| Diagnose and fix | A reproducible failure or unexpected behavior | Obtain evidence, locate the cause, apply the authorized fix, check for regressions | A supported cause and a verified fix, or the concrete limit of the diagnosis | `debugging.md`, `testing.md`, direct route/build |
| Review and verify | A diff, a built change, a spec, or criteria | Select the relevant controls, review, run checks, and record results | Findings or observable evidence; an approved review does not grant publication | `flow-build/references/verify-gate.md`; matrix in `agent-routing.md` |
| Audit debt and risks | Existing code or architecture | Review by lenses, validate findings, classify severity, and prioritize | Findings with evidence and proposed actions; an audit does not authorize fixing everything | `audit-playbook.md` |
| Deliver and promote | A verified change and an authorized destination | Reconcile permissions/Git state, meet the controls, publish or promote, check the environment | The observed terminal state, health/smoke checks, and applicable functional validation; delivery notes | `flow-build` CLOSE, `git-mechanics.md`, `promotion-playbook.md` |
| Migrate and maintain the workspace | A project predating the model or messy documentation | Inventory, classify, propose moves/repairs, apply what is authorized | A reconciled structure without losing decisions, references, or needed evidence | `migration-playbook.md`, `workspace-hygiene-playbook.md`, `workspace-archive` |
| Consolidate repositories | An explicit migration to a monorepo | Inspect dependencies and contracts, plan the cutover, execute and verify consumers | A bounded cutover with compatibility evidence; not confused with moving documents | `monorepo-cutover` |
| Communicate and document | Results, knowledge, or a documentation deliverable | Organize content, choose a format, generate and review readability/links | A verifiable document, report, or site; creating it does not authorize distributing it | `flow-report`, `starlight-docs-site` |

The playbooks are in `global/skills/flow-core/references/`; the rules cited by name are in `global/rules-situational/`. The [skills tree][h-skills] helps locate the other entries.

## 3. How they connect

There is no single mandatory chain for every request. These paths illustrate the documented connections:

- **New project:** need → requirements → project foundation → specification and business review → development → verification → authorized delivery.
- **Change to an existing product:** request → exploration → direct route or explicit formal plan → implementation → verification → delivery within the authorized scope.
- **Question or decision:** question → research → answer/proposal. It stops there; implementing requires another instruction that covers it.
- **Audit:** inventory → validated findings → prioritization → selection of fixes. Only the chosen fixes enter the development workflow.
- **Incident or bug:** evidence → diagnosis → authorized fix → check; promotion to the environment if it is part of the granted scope.
- **Resuming work:** read the record/handoff → compare with the current state → identify the real pending work → continue along the corresponding route.

Business review validates what the product must do. Technical review and tests validate the built change. The post-deployment check validates the delivered environment. These are different evidence; none automatically substitutes for the others.

## 4. Cross-cutting capabilities, not new workflows

| Capability | Role in the paths |
|---|---|
| Intake and scope control | Identify the deliverable, exclusions, authorized effects, closing evidence, and stopping point. |
| Memory and continuity | Recover decisions and context; cross-check them against current sources. `memory-policy`, `memory-sync`, and `engram-init-workspace` covered parts of this area. |
| Artifact organization | Separate code, durable documents, sessions, evidence, and temporary files. The historical paths in master do not replace the rebuild's current conventions. |
| Delegation and independent review | Distribute work when it adds value and the host allows it. It is not a new product phase and does not justify installing another runtime. |
| Technical rules | Guide tests, debugging, security, and stack according to the change. `language-rules` was an access surface, not a complete workflow. |
| Reports and handoff | Communicate results and leave enough information to continue. `flow-core` was a shared library, not an executable workflow. |
| Bounded autonomy | `unattended-delegation` handled delegated missions with limits; it did not turn every task into unattended work. |

## 5. What the reference repositories contribute

These comparisons refer to the local versions inspected. Their claims about savings, quality, or compatibility were not reproduced in this review.

| Reference | Documented pattern | What is worth studying for Hive | Difference or limit |
|---|---|---|---|
| **optional reference project** | ODD for everyday work; SDD by explicit choice, with proposal/spec/design/tasks and additional phases | Keep small work small; a single record for substantial work; separate research from implementation and preserve continuity | Its local SDD allows archiving without verification as a gate: `/sdd-verify` is an optional diagnostic and partial work can be archived with the pending items made explicit. It is not equivalent to Hive's formal `built → verified` gate. |
| **optional reference project** | Brainstorming → isolated environment → plan → execution → tests/review → branch closure | Executable specifications, systematic debugging, review against intent, and evidence before closing | Its README presents the path as mandatory and prescribes strong TDD practices. Do not carry over that obligation or its Git actions into the rebuild without a decision of our own. |
| **optional reference project** | Recognize → audit → validate findings → prioritize → write plans; delegated execution and reconciliation as options | A plan as a complete deliverable, with files, context, commands, and criteria; re-check the backlog before executing it | It is an audit/advisory reference and does not by itself cover Hive's whole cycle. The advantage of using cheap executors is a proposal of the project, not a result measured here. |
| **optional reference project** | Understand the problem and look for the sufficient solution: reuse, standard library, platform, and existing dependencies | Reduce unnecessary code and complexity without removing validation, security, or accessibility | It is mainly an implementation/review criterion, not a substitute for planning, authorizations, or verification. Its benchmarks do not prove an improvement for Hive. |

Sources: optional external research source, optional external research source, optional external research source, optional external research source, optional external research source.

The research on the [harness engineering corpus](2026-09-20-portable-harness-research.md), including the Uber ZIP, provides design and measurement criteria. By itself it neither defines our workflows nor shows that adopting these references would improve results.

## 6. Proposal for organizing the rebuild

**Proposal, pending a decision:** keep this map of needs and gradually choose which minimal procedures we implement. Do not restore the whole old catalog or combine all the frameworks.

A useful first cut would be:

1. **Investigate and decide:** an answer or proposal with evidence and a clear scope limit.
2. **Change and verify:** a proportional direct route; the formal plan stays available when it is chosen and adds value.
3. **Deliver and operate:** publish/promote only within the authorized scope, check the real state, and record pending items.
4. **Preserve and resume:** enough documentation and continuity, without accumulating scratch material as durable knowledge.

Bootstrap, specs, audit, migration, and documentation would remain specialized procedures selected as needed. This grouping is an organizational hypothesis, not four new skills or a commitment to implement.

The next step would be to choose a real path and its success criterion; then write the minimal guidance and test a few comparable cases. The existence of this map does not yet validate automatic selection of procedures or their operation across hosts.

## 7. Where to keep the references

Current location: `/path/to/reference-volume/dev-resources/reference/`. The user authorized moving the four references and confirmed that the volume is practically fixed on this Mac. The complete checkouts were moved, including `.git`, local files, and the `.jbcontextignore` marker of the parent directory. The four working trees ended up clean and kept their commits. The remotes were not updated.

| Repo | Local commit inspected | Approximate checkout size |
|---|---|---|
| optional reference project | `95edf9ff9172ca82f18ef34ccca2776b15348bb1` | 106 MB |
| optional reference project | `5bf4e78011075bcfc0dc295f0724994cd123ee71` | 8.5 MB |
| optional reference project | `cac56e1ebd3c279aa9153616cfeac7b174ab90f9` | 284 KB |
| optional reference project | `e3ba2aa6f1e6f0bc4d69eb09c9f0d0a93af56156` | 3.5 MB |
| optional reference project | `c55ee46073ed923f86ce59a5eb3b6d895095d1b7` (cloned 2026-09-24 from `optional reference project`; not yet inspected) | 3.2 MB |
| optional reference project | `e0881d2de397d5e9761d7b35ff5017d8f5ebf69b` (`github.com/pbakaus/optional reference project`; not yet inspected). Consulted only for UI/UX topics | — |
| optional reference project | `be0b51e8d0a14f7efd6f2fe5b3c2468daaecf362` (cloned 2026-09-24 from `github.com/686f6c61/optional reference project`; inspected). A Claude Code-only plugin based on hooks and a Python runtime, contrary to Hive's architecture; it is kept only for the checklists with no equivalent in the other references (`skills/{threat-model,compliance-check,evaluate-dependency,sbom-generate,incident-response}/SKILL.md`) and as a precedent for guards that block before executing (`hooks/secret-guard.py`, `dangerous-command-guard.py`, `evidence-guard.py`) if a measured failure ever justified them | 8.6 MB |

**Decision applied:** external references together on the resources volume; Hive under development stays in its usual workspace and the `master` clone stays at `dev-resources/tricell-hive-master`. The original directory was removed after verifying 3,505 file/link entries by content, file permissions, and link targets; the files were compared with SHA-256. No second copy or compatibility link was left.

Before the move, each repository was checked to have only its main worktree and no absolute symbolic links. The search for the previous path in Hive's active checkout and in the main Codex/Claude configurations found only this document; it was not an audit of all consumers on the Mac. Provenance links pinned to commits do not change. If the volume is not mounted, local lookup is temporarily unavailable.

## 8. Sources and limits of this review

- Hive: reading of the research, routing, planning, build, and common-library entries; bootstrap, specification, promotion, and hygiene procedures; skills inventory and audit/migration structure.
- References: README and relevant usage documentation of the four repositories. In particular, the ODD/SDD separation and the local verification/archiving policy of optional reference project were cross-checked.
- These workflows were not executed, integrations were not tested, benchmarks were not validated, and the full code of the external projects was not audited.
- The links below are pinned to inspected commits, not to moving branches. They are included for traceability; the reading itself was done locally.

[h-research]: https://github.com/JhonHawk/tricell-hive/blob/16e7d3357a3c41530d5e31460c3024872566f3c7/global/skills/flow-research/SKILL.md
[h-routing]: https://github.com/JhonHawk/tricell-hive/blob/16e7d3357a3c41530d5e31460c3024872566f3c7/global/skills/task-routing/SKILL.md
[h-agents]: https://github.com/JhonHawk/tricell-hive/blob/16e7d3357a3c41530d5e31460c3024872566f3c7/global/rules-situational/agent-routing.md
[h-plan]: https://github.com/JhonHawk/tricell-hive/blob/16e7d3357a3c41530d5e31460c3024872566f3c7/global/skills/flow-plan/SKILL.md
[h-build]: https://github.com/JhonHawk/tricell-hive/blob/16e7d3357a3c41530d5e31460c3024872566f3c7/global/skills/flow-build/SKILL.md
[h-skills]: https://github.com/JhonHawk/tricell-hive/tree/16e7d3357a3c41530d5e31460c3024872566f3c7/global/skills
