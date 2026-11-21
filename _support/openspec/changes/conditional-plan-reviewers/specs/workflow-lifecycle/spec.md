## ADDED Requirements

### Requirement: A plan that crosses domains gets one reviewer per domain

When a plan with a change folder has tasks that change files of two or more of these domains: server, client, guidance or infrastructure, or a data-entry screen where the user enters or chooses data, the plan SHALL be reviewed by one `hive-review-plan` child per counted domain, launched together, with at most three children, the data-entry screen's included. Tests and documentation SHALL count with the domain they check or describe, a file no domain names SHALL count with server, and a task that changes files of two domains SHALL count in both. With four or more domains, the domains with the fewest tasks SHALL share one assignment until at most three remain, a tie joining the later domain in that list to the earlier, and the data-entry screen assignment SHALL never be merged. Each brief SHALL name its domain, the questions its domain examines and the contracts it shares with the other counted domains. A finding several reviewers report SHALL be merged into one that names each reviewer. Any other plan with a change folder SHALL keep one reviewer. On a revision, a reviewer SHALL reread it when the revision changed the text, `Closes:` or verification of a task in its domain, or a criterion such a task closes, and every reviewer SHALL reread a change that belongs to no single domain.

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

## MODIFIED Requirements

### Requirement: Plan revisions keep the change folder coherent
A revision of a written plan that has a change folder SHALL edit in the same pass every artifact that states what changed, mark a replaced decision or criterion with `**Replaces:** <old IDs>` and a dropped criterion with `**Retired:** <reason>` without deleting the old line, reopen each completed task the revision invalidates, and run the coherence check (`flow-plan/scripts/check_change.py`, which also validates the canonical grammar of lines under `## Business rules` in `proposal.md` and of the delivery-unit boxes under `## Delivery units` in `tasks.md` when run without `--archive`) until it exits 0 before work continues; a revision that changes a criterion or the delta SHALL get one review round from the plan's reviewers, each `hive-review-plan` or `hive-review-ux` reviewer reread by one follow-up message to the same child, or, when the host has no tool to message it or the message returns an error, by a new child of its role given the earlier findings.

#### Scenario: A task still names a replaced decision
- **GIVEN** a change folder whose revision replaced a decision and marked the new one with `**Replaces:**`
- **WHEN** the coherence check runs and a task still names the replaced decision
- **THEN** the check prints that task's file and line and exits 1, and the revision is not complete until it exits 0

#### Scenario: A revision changes a criterion
- **GIVEN** a change folder whose plan was already reviewed
- **WHEN** a revision changes a criterion or the delta
- **THEN** the plan's reviewer rereads the revised folder in one round, through a follow-up message to the same child or a new child given its earlier findings, and a gap that survives the round goes to the user as a decision

#### Scenario: Malformed business rule line in proposal
- **GIVEN** a change folder whose `proposal.md` holds a line under `## Business rules` that violates the canonical prefix or field grammar
- **WHEN** `check_change.py` runs without `--archive`
- **THEN** the check prints the file and line of the malformed entry and exits 1

#### Scenario: A live box without its predicate
- **GIVEN** a `tasks.md` whose `## Delivery units` section holds `- [ ] live: <scenario>` without `· pass: <predicate>` and without `n/a:`
- **WHEN** `check_change.py` runs without `--archive`
- **THEN** the check prints the file and line and exits 1
