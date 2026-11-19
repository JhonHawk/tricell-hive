## MODIFIED Requirements

### Requirement: Decision rounds go through the card alone

When the selected native question tool waits for the answer and gives each option a description field, the agent SHALL ask a round of decisions or findings through that tool without writing them out in the text before it, also on a host whose own instructions reserve that tool for optional questions, such as Codex in its Default mode: the text before the first card SHALL hold the context line, the report or explanation the round needs and the waiting line, which says what is awaited and how many decisions are still to come, and SHALL NOT hold a decision's or a finding's title line, `Situation:` line, option lines or `I recommend` line. Each card SHALL stand alone: its question SHALL carry the decision's or finding's ID and title with the situation, or the fixed header of a session question. Where no waiting tool carries the question, or the selected card has no description field per option, the agent SHALL write the whole round in the text in its full format; when a card that follows a round written in the text comes back without an answer, the agent SHALL add one line naming the decisions still open and SHALL NOT write the round again. A plain acceptance SHALL apply the starred options only of decisions the user was shown, in a card that opened or in the text.

#### Scenario: A round answered through the card

- **GIVEN** a host whose question tool waits and shows a description per option
- **WHEN** the agent asks a round of three decisions
- **THEN** the text before the card holds the context line, the report and the waiting line, with no decision block
- **AND** the card asks the three decisions and the turn ends with it

#### Scenario: Codex in its Default mode

- **GIVEN** a Codex session in Default mode that offers `request_user_input`
- **WHEN** the agent asks the delivery round before its first edit
- **THEN** it asks the round through `request_user_input` and the turn waits for the answer
- **AND** it does not write the round as a plain-text question

#### Scenario: A card without option descriptions

- **GIVEN** a host whose question tool has no description field per option and permits the full round in text
- **WHEN** the agent asks a round
- **THEN** the whole round is written in the text, and the card follows it
- **AND** when that card comes back without an answer, one line names the decisions still open and the round is not written again

### Requirement: Decision information survives a required change of form

When no waiting tool carries a question and the host requires another form than Hive's decision format, the agent SHALL retain the decision's ID or fixed header, the alternatives with their letters, their consequences and the recommended option marked `(★)` with its reason, expressed in the form the host permits. It SHALL NOT interpret a different form as removing that information.

#### Scenario: The host requires a permission request in concise text

- **GIVEN** the host requires text for a permission request and no waiting tool carries it
- **WHEN** the agent asks permission
- **THEN** it states the recommended option with `(★)` and why, and the alternative with its letter and consequence
- **AND** it does not issue a duplicate card

### Requirement: Decisions carry labelled context

A decision `D<n>` shown to the user, and a session question shown under its fixed header, SHALL carry, under its title line, a `Situation:` line that says what happened or what was found, what was verified and what to check, with its target, in words the user has used or that the same message explains; under each option, an `If you choose it:` line that says what the user gets and what it costs them in time, money, risk or what stops working; and an `I recommend <letter>:` line whose reason is a consequence for the user and never a quotation of a rule or a skill. The three labels SHALL be written in the session language, while the IDs, fixed headers, letters and the `(★)` mark stay as they are. The round's waiting line SHALL also say how many known decisions are still to come. In a native question card, the question SHALL carry the situation, each option's description the text of its `If you choose it:` line, or the option's effect in one line when it has none, and the starred option's description SHALL end with the reason of the recommendation. The "Open the viewer" option that a validation question carries SHALL have no `If you choose it:` line. A finding `H<n>` asked in a round carries its title and option lines without the three labelled lines, and the replies of a next-step block keep their own format.

#### Scenario: A rule requires the question

- **GIVEN** a rule requires the agent to ask before an effect
- **WHEN** the agent asks it as a decision
- **THEN** the recommendation line states the effect that needs the answer, such as its cost or what it changes
- **AND** it quotes no rule or skill as its reason

#### Scenario: A name the user was not given

- **WHEN** the situation would name a task ID, a file, a branch or an internal name the user has not been given
- **THEN** the same message explains it, or the situation leaves it out

### Requirement: Entry router with playbooks
Hive SHALL provide an entry skill, `hive-mode`, that replaces `flow-research`. It SHALL read before asking, play the request back, ask what is missing and settle what the user needs, and SHALL NOT edit repository files itself; it MAY hand the work to `flow-build` in the same turn only for a small understood change that the user's own request named with its target, never after a reply to its question or proposal. It SHALL classify the work and open at most one playbook under `playbooks/<kind>.md`, record its name in the change folder's `proposal.md` (written by `flow-plan`) or, for work without a change folder, in a `- Playbook:` line of the checkpoint file, and, when no playbook fits or the work is large and cross-cutting, follow the general route from research to plan or build. A playbook names the order of the stages, what is specific to that kind and the leaves it uses; the stages (`flow-figma-design`, `flow-plan`, `flow-build`, `flow-verify`, `flow-close`, `flow-deploy`) SHALL remain commands that read the playbook's specific steps.

#### Scenario: A reported defect with a fix request
- **WHEN** the user reports a defect and asks for its fix
- **THEN** `hive-mode` opens the `bugfix` playbook, and build and verify fix the defect at its cause as «Defects are fixed at their cause» sets, with its regression evidence.

#### Scenario: An investigation confirms a defect
- **GIVEN** work running the `investigate` playbook
- **WHEN** the findings confirm a defect and the user asks for its fix
- **THEN** the work hands off explicitly to the `bugfix` playbook as a scope change, and the new playbook is recorded

### Requirement: Verify complements and adversarial review option
Before the code review, `flow-verify` SHALL read the leaves that remove narrating comments, check the change's impact beyond the diff, and check the prose of the pull request, commits and documentation. The adversarial cross-vendor review SHALL be asked as its own session question for each work item that gets a code review, stating that it spends another provider's quota and listing the installed and signed-in command-line agents of other providers for the user to choose, under the reuse and spending limits the user set for it; when none is available, verify SHALL report it as unavailable and run the chosen review only.

#### Scenario: The adversarial review is chosen
- **WHEN** the user answers yes to the adversarial review for a work item and picks an agent
- **THEN** verify runs it with that agent on that work item's candidate within its limit, reruns each reported finding before keeping it, and runs no such review on work items where it was not chosen.

## ADDED Requirements

### Requirement: Moving between stages shows the next-step block

Every move from one flow to the next SHALL end the report with the next-step block and its command, except the continuations its rule names. A reply to a question or the acceptance of a proposal SHALL only define or redirect scope and SHALL NOT start a flow. Research SHALL continue into build in the same turn only when the user's own request named the change and its target.

#### Scenario: A ticket picked for research

- **GIVEN** the user picks a ticket for research without asking for a change
- **WHEN** the agent's proposal is accepted with a reply such as "1"
- **THEN** the next message ends with the next-step block toward `flow-build` or `flow-plan`
- **AND** no build starts until the user's reply names it

### Requirement: Session questions carry fixed headers

The delivery mode, the code review mechanism and the adversarial review question SHALL carry a fixed header in the session language instead of a decision ID, with the rest of the decision format. The code review question SHALL offer no-review only for a mechanical change. The adversarial review SHALL list the installed and signed-in agents of other providers for the user to choose.

#### Scenario: Delivery round of a UI fix

- **GIVEN** a work item that fixes a screen flow
- **WHEN** the agent asks the session questions
- **THEN** they read as **Delivery** and **Review** without `D<n>`, and the review options are mechanisms

### Requirement: Tracker states are discovered and cached

Before moving a ticket, the agent SHALL read that ticket's available transitions from its tracker, and SHALL take the equivalent state when a transition fails. It SHALL keep the names and categories of the project's states in the Engram topic `tracker/states/<tracker>/<prefix>`, where `<prefix>` is the project's key prefix plus the ticket type when the workflow differs by type, written when it discovers them or after a failed transition, and SHALL use it to choose the state, never in place of the live read.

#### Scenario: A localized workflow

- **GIVEN** a Jira project whose in-progress state is named «En curso»
- **WHEN** a transition to "In Progress" fails
- **THEN** the agent reads the transitions, moves the ticket to «En curso» and writes the cache

### Requirement: A ticket is documented with evidence before cleanup

At a work item's close, the agent SHALL document its ticket with the evidence the work produced, such as captures and test output, before any step that removes support material, temporary files or worktrees.

#### Scenario: Captures before cleanup

- **GIVEN** a closed UI fix with before and after captures in its workspace
- **WHEN** the close runs
- **THEN** the captures reach the ticket before the cleanup removes them

### Requirement: Verification floor by kind of change

One rule SHALL set, per kind of change (mechanical, defect fix, feature, UI, risky, performance), the checks that always run, linking each check's own rule. The code review SHALL run on every non-mechanical change. A security review SHALL be recommended for a risky change that touches a security boundary.

#### Scenario: A UI defect fix under hold

- **GIVEN** a defect fix to a screen delivered under `hold`
- **WHEN** it is verified
- **THEN** its regression test runs failing then passing, the code review and the UX review run, and before and after captures are taken

### Requirement: The final report carries evidence and a try-it block

The final report SHALL carry a defect's regression test output seen failing and then passing, and, when the change can be seen or run, a block with the URL or command, the test account's index row, the steps and the expected result, in every delivery mode.

#### Scenario: Hold delivery

- **GIVEN** a change delivered under `hold`
- **WHEN** the agent reports it
- **THEN** the report holds the failing and passing output and the try-it block
