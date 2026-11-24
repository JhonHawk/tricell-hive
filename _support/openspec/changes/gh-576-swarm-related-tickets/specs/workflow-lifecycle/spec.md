## MODIFIED Requirements

### Requirement: Swarm mode coordinates several tickets through ticket owners

Hive SHALL offer a swarm mode, started only by the user's command or by a request in which the user names the mode, whose global guidance line links the swarm skill; the model SHALL NOT start it on its own. In it, one session coordinates two to four tickets. In one round before any worktree exists, the coordinator SHALL show each ticket's eligibility verdict (it passes the ready list of a ticket, its estimated scope stays within one repository and ten files, it writes no stored data and runs no migration, and for a screen change the project can serve it on an assigned port, and when the project's specs live in another repository, that repository declares `Delivery: direct-base`), SHALL propose groups of related tickets, whose tickets share one repository, naming related tickets in different repositories as a pair of which the user keeps at most one, those whose estimated paths cross, that touch the same module or where one depends on another, each group with its reason and at most three tickets and ten estimated files in all, with tickets whose estimated paths cross, not counting files nearly every change touches such as a changelog, lockfiles, manifests and generated indexes, always in one group or not taken together, and with a recommendation of the ordinary flow when the batch forms a single group, and SHALL ask the user to confirm or change the groups, the eligible tickets, the delivery of each repository that holds their code, which SHALL be neither `hold` nor `direct-base` there while a separate specs repository keeps its declared `direct-base`, the caps on simultaneous owners (at most four) and on runs per group, and the creation of one worktree per group; that answer SHALL be the authorization to create those worktrees. The batch SHALL count two to four tickets, grouped or not, and a single ticket SHALL be a group of one. An ineligible ticket SHALL be named with its reason and left for its own session. For each group, one `hive-ticket-owner` child SHALL research and plan its tickets in one change folder whose criteria carry their ticket's key, writing it in its worktree or, when the project's specs live in another repository, in the one change-folder path of that repository's main checkout the coordinator assigns to it, then stop and return its plan, whose criteria are written `AC<n> (<key>)`, with every decision the user must take, the external effects, the assumed business rules, the paths it will write and each external fact as a command with its excerpt. After the coordinator's grouped rounds, the same child resumed with its history where the host documents that, or a new child of the role reading the plan and checkpoint, SHALL build and stop again with the exact head, commands, results and evidence. An owner MAY stop at any point to return a blocker, does inline what an ordinary flow would delegate to a child, and SHALL return each reversible decision it took alone as a `Ruling:` line, which the coordinator writes to the decision log. At the plan stop, and for each path a build-time blocker adds, the coordinator SHALL compare the group's paths to write with those of every other group not yet integrated, taken from its returned plan, else its record's allowed paths, else its round-0 estimates, and SHALL repeat it for every plan that arrives later; when they cross outside those common files, it SHALL stop the arriving group, stop the other at its next return, hold the publication of another group already past its build stop, and ask the user to regroup or set one aside; a conflict in a common file at integration SHALL be resolved when the base is merged before the merge. The coordinator SHALL move each ticket to in progress before its build, install dependencies in each worktree, create each group's execution record, one unit, once its plan is approved, launch `hive-review-plan` only for a plan that is a risky change, verify each returned head through `validate-handoff`, one `hive-verify-task` child covering all the group's criteria, recorded as the unit's verifier, and the group's checks before integrating it, commit a change folder written in another repository's checkout, publish each group on its own branch and a pull request that names each of its tickets under the delivery in force, move each ticket of the group when it integrates, and keep a failing group from stopping the others. When the cap on simultaneous owners is full, the other groups SHALL wait and start as running owners stop; when a group reaches its cap on runs, the coordinator SHALL launch no further owner for it and SHALL report it with its worktree, head and checkpoint, while the other groups continue. An owner whose change folder lies in its worktree SHALL commit it with its checkpoint before each return, and that path SHALL be among its unit's allowed paths. For a group with a ticket that changes a screen, the coordinator SHALL assign the group a distinct port as an exclusive resource, and its owner SHALL serve its worktree on that port and take the before and after captures there. The swarm mode SHALL NOT be unattended delegation, and the user's silence SHALL NOT extend it.

#### Scenario: Three tickets, one ineligible
- **WHEN** the user starts the swarm mode with three tickets and one of them runs a migration
- **THEN** the round shows that ticket as ineligible with its reason, and after the user's answer the coordinator creates worktrees and launches owners only for the other two

#### Scenario: Two tickets touch the same file
- **WHEN** two eligible tickets of the batch are estimated to change the same file
- **THEN** round 0 proposes them as one group with that reason, and after the user confirms, one owner plans and builds both in one worktree and one pull request names both tickets

#### Scenario: A decision found during the build
- **WHEN** an owner meets, during its build, a decision that changes a business rule
- **THEN** it stops and returns that decision, the coordinator asks the user, and the owner resumes or a new one continues from the plan and checkpoint

#### Scenario: The run cap is reached
- **WHEN** a group reaches its cap on runs before its second stop
- **THEN** no new owner starts for it and the report names its worktree, head and checkpoint

### Requirement: Local deliverable owner exception

Only a designated E2E local-owner role and the swarm mode's `hive-ticket-owner` role SHALL be allowed to create local commits, each within its assigned repository, worktree, branch and write paths, using the existing scoped-commit and secret-scan mechanism; the `hive-ticket-owner` role SHALL also write the change folder and checkpoint of its assigned tickets, in its worktree or in the one assigned change-folder path of another repository's main checkout, where it SHALL NOT commit. These named exceptions SHALL NOT expand ordinary implementation roles. The coordinator SHALL create and assign worktrees, arrange exclusive resources, launch independent verifiers and retain all publication, pull-request, tracker and integration operations. Owners SHALL NOT launch children, publish, integrate remotely, alter another assignment or infer authority from record fields.

#### Scenario: Local handoff
- **WHEN** an owner finishes an assigned unit
- **THEN** it returns the exact local head, command results, evidence and limitations, stops its own processes, and waits for the coordinator's next assignment rather than publishing its branch.

### Requirement: Root-owned execution record

Multiunit execution SHALL have one versioned local execution record linked from the existing checkpoint; the swarm mode SHALL keep one such record per group of tickets, created once the group's plan is approved, whose unit role is `hive-ticket-owner`, in the ignored support workspace of the coordinator's checkout even when the group's change folder lies in a worktree or another repository. The coordinator SHALL be its only writer. Record operations SHALL reject unknown versions, duplicate keys/IDs, invalid task/criterion/dependency references, cycles, conflicting active assignments and stale generations without losing the prior valid state. Machine paths, native IDs and raw run evidence SHALL remain outside versioned change records. Existing progress/checkpoint commands SHALL remain compatible.

#### Scenario: Stale update
- **WHEN** an update names a generation that no longer matches the record
- **THEN** the operation fails with a correction naming the stale state and preserves the current record.

### Requirement: A plan that crosses domains gets one reviewer per domain

When a plan with a change folder has tasks that change files of two or more of these domains: server, client, guidance or infrastructure, or a data-entry screen where the user enters or chooses data, which also counts as client, the plan SHALL be reviewed by one `hive-review-plan` child per counted domain, launched together, with at most three children, the data-entry screen's included. Tests and documentation SHALL count with the domain they check or describe; repository-wide files, such as manifests, lockfiles, build configuration, the changelog and the change folder, SHALL count with the other files of the task that changes them, or with server when that task changes no other file; any other file no domain names SHALL count with server; and a task that changes files of two domains SHALL count in both. With four or more domains, the assignment with the fewest tasks SHALL be merged into the one with the next fewest, a merged assignment counting the sum of its tasks and taking its earliest domain's place, until at most three remain, a tie taking the later domain in that list, and the data-entry screen assignment SHALL never be merged. Each brief SHALL name its domain, the questions its domain examines and the contracts it shares with the other counted domains, and only the first assignment SHALL run the whole-plan checks of criteria, external facts and verification lines, the others being told to skip them. A finding several reviewers report SHALL be merged into one that names each reviewer. Any other plan with a change folder SHALL keep one reviewer, except a group's plan in the swarm mode, which SHALL get one only when it is a risky change. When a revision is reread, as the change-folder requirement sets, the reviewer of each domain whose tasks or criteria it changed SHALL reread it, the first assignment SHALL also reread a change to an external fact or a verification line, and every reviewer SHALL reread a change that belongs to no single domain.

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
