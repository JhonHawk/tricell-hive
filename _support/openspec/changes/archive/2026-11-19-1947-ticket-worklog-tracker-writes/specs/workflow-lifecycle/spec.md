## MODIFIED Requirements

### Requirement: Verify complements and adversarial review option
Before the code review, `flow-verify` SHALL read the leaves that remove narrating comments, check the change's impact beyond the diff, and check the prose of the pull request, commits and documentation. A review of the candidate by another provider's agent SHALL run only when the user explicitly asks for it: no delivery round SHALL ask or suggest it and no agent SHALL probe other hosts before that request. When the user asks, the agent SHALL list the installed and signed-in command-line agents of other providers for the user to choose, under the reuse and spending limits the user set; when none is available, it SHALL report the review as unavailable.

#### Scenario: Delivery round without an adversarial question
- **WHEN** the agent asks the delivery and review questions for a work item
- **THEN** the round has no adversarial review question and no command checks other hosts' installation or sign-in

#### Scenario: The user asks for a review by another provider
- **WHEN** the user asks for an adversarial review of the candidate and picks an agent
- **THEN** verify runs it with that agent on that work item's candidate within its limit and reruns each reported finding before keeping it

### Requirement: Session questions carry fixed headers

The delivery mode, the code review mechanism and the security review questions SHALL carry a fixed header in the session language instead of a decision ID, with the rest of the decision format. A mechanical change SHALL get no code review question; every other change SHALL get a code review, so the question SHALL offer only review mechanisms, also under `Delivery: direct-base`.

#### Scenario: Delivery round of a UI fix

- **GIVEN** a work item that fixes a screen flow
- **WHEN** the agent asks the session questions
- **THEN** they read as **Delivery** and **Review** without `D<n>`, and the review options are mechanisms

### Requirement: Tracker states are discovered and cached

Before moving a ticket, the agent SHALL read that ticket's available transitions live, through the first tool that lists them in this order: the tracker's CLI, the host's authenticated MCP tool for that tracker, then the tracker's REST API with the user's global API token; it SHALL move to the state the project uses for that purpose by its exact name, and SHALL report a blocker instead of picking among states that share a category when none is recorded for that purpose. It SHALL keep the names and categories of the project's states in the Engram topic `tracker/states/<tracker>/<prefix>`, where `<prefix>` is the project's key prefix plus the ticket type when the workflow differs by type, written when it discovers them or after a failed transition, and SHALL use it to choose the state, never in place of the live read.

#### Scenario: A localized workflow

- **GIVEN** a Jira project whose in-progress state is named «En curso»
- **WHEN** a transition to "In Progress" fails
- **THEN** the agent reads the transitions, moves the ticket to «En curso» and writes the cache

## ADDED Requirements

### Requirement: Tracker writes follow the tools' capabilities

The agent SHALL write to a tracker through the first tool that has the needed capability, in this order: the tracker's CLI, the host's authenticated MCP tool, then the tracker's REST API with the user's global API token. The token SHALL live in a file the user creates once under Hive's global secret location, `~/.config/hive/secrets/` (folder `0700`, file `0600`); a script SHALL read it, and the agent SHALL check only that it exists, SHALL never write it, and SHALL ask the user to create it when it is missing. The token SHALL never appear in a command's arguments, output, an error, a message, a durable place or another provider's agent. Attachments and worklogs the CLI cannot write SHALL go through the next tool in that order.

#### Scenario: Attaching captures in Jira

- **GIVEN** `acli` lists but cannot upload attachments and the host has no MCP tool that uploads
- **WHEN** the close documents the ticket with its captures
- **THEN** the captures upload through the REST API with the global token and the token appears in no output

#### Scenario: The token file is missing

- **GIVEN** no `~/.config/hive/secrets/jira-api-token`
- **WHEN** a worklog must be written and no other tool can write it
- **THEN** the agent asks the user to create the file with its permissions and does not create it

### Requirement: Tickets record the time worked

For a ticket in a tracker with native time tracking, the agent SHALL add a worklog at the work item's delivery point (integration, the `ready` stop or the `hold` stop) to each ticket the session worked, after reading the worklogs its account already has on that ticket, holding only the converted time and no description, and SHALL update that same worklog at the close with the time added after delivery. Each ticket's time SHALL be the active time from that ticket's first message in each session that worked it, minus only the waits for the user; tickets worked together in a session SHALL each receive the full time. It SHALL be converted as the active time without the remote CI waits it recognizes times the factor (2.0 by default, between 1 and 3) plus those waits at 1. It SHALL be measured with `hive time` from the host's native transcript where the manager reads one, and from the agent's own timestamps otherwise; the plan's saved state SHALL keep the host and session ID of each session that worked the ticket so the delivering session measures them again. The worklog's ID, the real time, the factor and the ticket type SHALL stay in the work's close record in persistent memory and SHALL NOT appear in the tracker.

#### Scenario: A delivered Jira ticket

- **GIVEN** a session that researched, built and integrated a Jira ticket in 40 minutes of active time, 10 of them waiting on remote CI
- **WHEN** the work reaches its integration
- **THEN** the ticket gains one worklog of 70 minutes with no description, and the close record keeps the worklog ID, 40 minutes, factor 2.0 and the ticket type

#### Scenario: A tracker without time tracking

- **GIVEN** a GitHub issue
- **WHEN** the work reaches its delivery point
- **THEN** no worklog is written and nothing about time reaches the issue

### Requirement: `hive time` reports a session's active time

`hive time` SHALL print one JSON object with `result` (`ok`, `partial` or `no_session_file`), `host`, `session`, `since`, `until`, `active_seconds`, `user_wait_seconds`, `ci_wait_seconds`, `factor` and `converted_seconds`, rounded up to the minute. It SHALL read the whole transcript; when it cannot, it SHALL report `partial`. A host without a transcript the manager reads SHALL give `no_session_file` with exit code 0. A factor below 1 or above 3 SHALL exit with code 2. Only messages a person wrote and answers to a question tool SHALL end a user wait.

#### Scenario: A background CI wait in Claude Code

- **GIVEN** a Claude Code session that started `gh pr checks --watch` in the background and resumed on its task notification 6 minutes later
- **WHEN** `hive time` reads the session
- **THEN** those 6 minutes count as CI wait, not as user wait
