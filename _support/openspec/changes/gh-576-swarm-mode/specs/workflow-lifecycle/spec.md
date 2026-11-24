## ADDED Requirements

### Requirement: Swarm mode coordinates several tickets through ticket owners

Hive SHALL offer a swarm mode, started only by the user's command or by a request in which the user names the mode, whose global guidance line links the swarm skill; the model SHALL NOT start it on its own. In it, one session coordinates two to four tickets. In one round before any worktree exists, the coordinator SHALL show each ticket's eligibility verdict (it passes the ready list of a ticket, its estimated scope stays within one repository and ten files, it writes no stored data and runs no migration, and for a screen change the project can serve it on an assigned port) and SHALL ask the user to confirm the eligible tickets, the delivery, the caps on simultaneous owners (at most four) and on runs per ticket, and the creation of one worktree per ticket; that answer SHALL be the authorization to create those worktrees. An ineligible ticket SHALL be named with its reason and left for its own session. For each ticket, one `hive-ticket-owner` child SHALL research and plan, writing the ticket's change folder in its worktree or, when the project's specs live in another repository, in the one change-folder path of that repository's main checkout the coordinator assigns to it, then stop and return its plan with every decision the user must take, the external effects, the assumed business rules, the paths it will write and each external fact as a command with its excerpt. After the coordinator's grouped rounds, the same child resumed with its history where the host documents that, or a new child of the role reading the plan and checkpoint, SHALL build and stop again with the exact head, commands, results and evidence. An owner MAY stop at any point to return a blocker, does inline what an ordinary flow would delegate to a child, and SHALL return each reversible decision it took alone as a `Ruling:` line, which the coordinator writes to the decision log. The coordinator SHALL move each ticket to in progress before its build, install dependencies in each worktree, create each ticket's execution record once its plan is approved, launch `hive-review-plan` only for a plan that is a risky change, verify each returned head through `validate-handoff`, an independent verifier and the ticket's checks before integrating it, commit a change folder written in another repository's checkout, publish each ticket on its own branch and pull request under the delivery in force, and keep a failing ticket from stopping the others. When a cap is reached, the coordinator SHALL launch no further owner and SHALL report each open ticket with its worktree, head and checkpoint. For a ticket that changes a screen, the coordinator SHALL assign a distinct port as an exclusive resource, and its owner SHALL serve its worktree on that port and take the before and after captures there. The swarm mode SHALL NOT be unattended delegation, and the user's silence SHALL NOT extend it.

#### Scenario: Three tickets, one ineligible
- **WHEN** the user starts the swarm mode with three tickets and one of them runs a migration
- **THEN** the round shows that ticket as ineligible with its reason, and after the user's answer the coordinator creates worktrees and launches owners only for the other two

#### Scenario: A decision found during the build
- **WHEN** an owner meets, during its build, a decision that changes a business rule
- **THEN** it stops and returns that decision, the coordinator asks the user, and the owner resumes or a new one continues from the plan and checkpoint

#### Scenario: The run cap is reached
- **WHEN** a ticket reaches its cap on runs before its second stop
- **THEN** no new owner starts for it and the report names its worktree, head and checkpoint

### Requirement: Ordinary flows carry no swarm exception

Every exception the swarm mode makes to ordinary rules SHALL live in the swarm skill and the `hive-ticket-owner` role, with one line of the global guidance naming and linking the mode and the execution-record script accepting the role. No other distributed file SHALL name the mode or the role, and a content test SHALL fail when one does.

#### Scenario: A pointer added to a flow
- **WHEN** a change adds a mention of the swarm mode to `flow-build/SKILL.md`
- **THEN** the content test fails and names that file

## MODIFIED Requirements

### Requirement: Local deliverable owner exception

Only a designated E2E local-owner role and the swarm mode's `hive-ticket-owner` role SHALL be allowed to create local commits, each within its assigned repository, worktree, branch and write paths, using the existing scoped-commit and secret-scan mechanism; the `hive-ticket-owner` role SHALL also write its ticket's change folder and checkpoint, in its worktree or in the one assigned change-folder path of another repository's main checkout, where it SHALL NOT commit. These named exceptions SHALL NOT expand ordinary implementation roles. The coordinator SHALL create and assign worktrees, arrange exclusive resources, launch independent verifiers and retain all publication, pull-request, tracker and integration operations. Owners SHALL NOT launch children, publish, integrate remotely, alter another assignment or infer authority from record fields.

#### Scenario: Local handoff
- **WHEN** an owner finishes an assigned unit
- **THEN** it returns the exact local head, command results, evidence and limitations, stops its own processes, and waits for the coordinator's next assignment rather than publishing its branch.

### Requirement: A plan that crosses domains gets one reviewer per domain

When a plan with a change folder has tasks that change files of two or more of these domains: server, client, guidance or infrastructure, or a data-entry screen where the user enters or chooses data, which also counts as client, the plan SHALL be reviewed by one `hive-review-plan` child per counted domain, launched together, with at most three children, the data-entry screen's included. Tests and documentation SHALL count with the domain they check or describe; repository-wide files, such as manifests, lockfiles, build configuration, the changelog and the change folder, SHALL count with the other files of the task that changes them, or with server when that task changes no other file; any other file no domain names SHALL count with server; and a task that changes files of two domains SHALL count in both. With four or more domains, the assignment with the fewest tasks SHALL be merged into the one with the next fewest, a merged assignment counting the sum of its tasks and taking its earliest domain's place, until at most three remain, a tie taking the later domain in that list, and the data-entry screen assignment SHALL never be merged. Each brief SHALL name its domain, the questions its domain examines and the contracts it shares with the other counted domains, and only the first assignment SHALL run the whole-plan checks of criteria, external facts and verification lines, the others being told to skip them. A finding several reviewers report SHALL be merged into one that names each reviewer. Any other plan with a change folder SHALL keep one reviewer, except a ticket's plan in the swarm mode, which SHALL get one only when it is a risky change. When a revision is reread, as the change-folder requirement sets, the reviewer of each domain whose tasks or criteria it changed SHALL reread it, the first assignment SHALL also reread a change to an external fact or a verification line, and every reviewer SHALL reread a change that belongs to no single domain.

#### Scenario: Server and client tasks

- **GIVEN** a plan whose tasks change a server endpoint and a web client's dialog where the user picks a time
- **WHEN** the plan is written
- **THEN** a server reviewer, a client reviewer and a data-entry screen reviewer are launched together
- **AND** each brief names the payload the server and the client share

#### Scenario: Four domains

- **GIVEN** a plan whose tasks change server, client, guidance and infrastructure files, with three server tasks, three client tasks, two guidance tasks and one infrastructure task
- **WHEN** the plan is written
- **THEN** three reviewers are launched, and infrastructure shares one assignment with guidance

#### Scenario: A plan in one domain

- **GIVEN** a plan whose tasks change only agent guidance files
- **WHEN** the plan is written
- **THEN** one `hive-review-plan` child reviews it

#### Scenario: A revision that touches one domain

- **GIVEN** a plan reviewed by a server and a client reviewer
- **WHEN** a revision changes a criterion that only a server task closes
- **THEN** the server reviewer rereads the revised folder, and the client reviewer does not

### Requirement: Root-owned execution record

Multiunit execution SHALL have one versioned local execution record linked from the existing checkpoint; the swarm mode SHALL keep one such record per ticket, created once the ticket's plan is approved, whose unit role is `hive-ticket-owner`, in the ignored support workspace of the coordinator's checkout even when the ticket's change folder lies in a worktree or another repository. The coordinator SHALL be its only writer. Record operations SHALL reject unknown versions, duplicate keys/IDs, invalid task/criterion/dependency references, cycles, conflicting active assignments and stale generations without losing the prior valid state. Machine paths, native IDs and raw run evidence SHALL remain outside versioned change records. Existing progress/checkpoint commands SHALL remain compatible.

#### Scenario: Stale update
- **WHEN** an update names a generation that no longer matches the record
- **THEN** the operation fails with a correction naming the stale state and preserves the current record.

### Requirement: Decision log for unattended and coordinated work
Unattended work, bounded E2E coordination and the swarm mode SHALL keep one decision log at `workspace/YYYY-MM-DD-<work>/decisions.tsv` of the main checkout's support home, a tab-separated file with the columns `ts`, `phase`, `decision`, `why`, `evidence` and `result`, one single-line row per decision and one per external effect, whose evidence names the target and the change and whose result names how to revert it. Evidence SHALL be a link or path, never a paragraph. Its format SHALL have one home, in `unattended-delegation`, which E2E coordination links; in E2E and the swarm mode only the coordinator writes it, and in the swarm mode it writes each `Ruling:` an owner returns as one row. The log SHALL NOT be committed or published unless the user asks, and the final report SHALL cite it by path. The task record SHALL keep only progress.

#### Scenario: An unattended run takes an external effect
- **WHEN** an unattended run changes a resource outside the repository within its authorized ceiling
- **THEN** `decisions.tsv` gains one row whose evidence names the target and the change and whose result names how to revert it, and the task record gains no copy of it
