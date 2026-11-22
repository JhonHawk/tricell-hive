## ADDED Requirements

### Requirement: The session audit fails a plan that ends without a build route

`go run ./tests/pilot --audit-session` SHALL report `plan_without_build_route` for Claude, Codex and Grok sessions. A plan stage SHALL run from a load of `flow-plan`, through the skill call or a read of its installed `SKILL.md`, to the next load of a different flow's skill or the end of the session; a further load of `flow-plan`, a search event or a `grep` or `rg` of a flow's `SKILL.md` SHALL NOT end it. The criterion SHALL pass a stage in which an assistant message holds the command of `flow-build` or `create-design` in any host's syntax, such as `/flow-build`, `$flow-build` or `**/flow-build**`; SHALL fail a stage that a later flow load ends with no such message, giving the line of the `flow-plan` load, except that it SHALL report review when that load is in the turn of the stage's latest `flow-plan` load with no question card between them, a continuation the user may have asked for, or follows a load of the installed `unattended-delegation` skill; SHALL report review for a stage the end of the session ends with no such message, since the session may still be open; SHALL report not_observed when the session loads no `flow-plan`; and SHALL NOT judge a child session.

#### Scenario: A plan kept in the conversation that never names the build command

- **GIVEN** a Codex session that loads `flow-plan`, writes the plan in a message with no flow command, and later loads `flow-build` after the user answers a delivery round
- **WHEN** the audit runs on it
- **THEN** `plan_without_build_route` fails with the line of the `flow-plan` load

#### Scenario: A plan that ends with the next-step block

- **GIVEN** a session whose plan message ends with a `▶` block naming `/flow-build`
- **WHEN** the audit runs on it
- **THEN** `plan_without_build_route` passes

#### Scenario: A plan and its build in one turn

- **GIVEN** a session whose user asks to plan and build without stopping, and that loads `flow-plan` and then `flow-build` in the same turn with no flow command in its messages
- **WHEN** the audit runs on it
- **THEN** `plan_without_build_route` reports review

#### Scenario: A message that names only the plan command

- **GIVEN** a plan stage whose only flow command in an assistant message is `/flow-plan`, ended by a load of `flow-build`
- **WHEN** the audit runs on it
- **THEN** `plan_without_build_route` fails

### Requirement: The session audit reads a round written with numbered questions

In `go run ./tests/pilot --audit-session`, a round title SHALL match on a single line, and a decision round written as text SHALL also be a block whose title line is numbered and ends in a question, optionally followed by a short parenthetical note, such as `**1. ¿Dónde trabajamos?**`, when at least one option line marked `(★)` follows it before the next title and the message has no `D<n>` title or session header, whose Situation may list numbered questions; a numbered question with no such option SHALL remain a sentence that `prose_choice_without_card` reads. An option line SHALL count as the recommended one when `(★)` follows its letter with or without a separator, as in `- **A (★) · Checkout actual.**`. On Codex, `review_question_native_option` SHALL judge a review question whose options name at least two review mechanisms, and one that asks for the code review itself, such as «¿Cómo revisamos el código?» or a title under the fixed **Review** header, whose options name at least one; a security review question with one mechanism SHALL NOT be judged.

#### Scenario: The delivery round written as numbered questions

- **GIVEN** a Codex session in Default mode with `request_user_input` offered, whose message asks «**1. ¿Dónde trabajamos?**», «**2. ¿Hasta dónde entregamos?**» and «**3. ¿Cómo revisamos el código?**», each with a starred lettered option and the third with only `hive-review-code`, and no card after it
- **WHEN** the audit runs on it
- **THEN** `text_round_without_card`, `decision_round_format` and `review_question_native_option` fail with that message's line

#### Scenario: An analysis that lists numbered questions

- **GIVEN** an analysis whose message lists «**1. ¿Qué falló?**» and «**2. ¿Por qué?**» with no option lines
- **WHEN** the audit runs on it
- **THEN** none of those three criteria treats the message as a round

### Requirement: The session audit reads an imperative confirmation and quoted replies as prose choices

`prose_choice_without_card` SHALL also read as a question a sentence with no question mark that opens with, or follows a comma or colon with, an imperative confirmation such as «confirma», «confírmame» or "confirm", and SHALL fail it when it names a session choice and no card follows in its turn. It SHALL report review for a message that offers two or more replies in quotation marks after a phrase such as «puedes responder», «responde» or "you can reply", when no card follows in its turn.

#### Scenario: Delivery asked as an imperative confirmation

- **GIVEN** a Codex session with the card offered, whose message says «Para arrancar, confirma `en el checkout actual, entrega local (hold)`.» and ends its turn without a card
- **WHEN** the audit runs on it
- **THEN** `prose_choice_without_card` fails with that message's line

#### Scenario: Replies offered as quoted phrases

- **GIVEN** a message that says «Puedes responder «apruebo», «corrige el filtro» o «abre el visor»» with no card after it and no other prose choice in the session
- **WHEN** the audit runs on it
- **THEN** `prose_choice_without_card` reports review

#### Scenario: A sentence that only reports a confirmation

- **GIVEN** a message that says «El log confirma que la entrega terminó.»
- **WHEN** the audit runs on it
- **THEN** `prose_choice_without_card` does not count it

### Requirement: The session audit reads a stage packet printed inside a Codex exec cell

When a Codex `exec` cell's output holds as many plain-text items as the cell has shell commands, the audit SHALL pair each item with the shell command at its position, and only the reading of stage packets SHALL use it, under the same completeness rule as any packet read: the end line, every heading it names, no truncation marker and no filter. With a different count, no shell command of that cell SHALL get a plain-text item.

#### Scenario: A stage packet read inside an exec cell

- **GIVEN** a Codex `exec` cell with two shell commands, the second of which runs `cat` on `hive-mode/stage.md`, and an output with two plain-text items, the second holding the packet with its end line and every heading
- **WHEN** a later analysis ends without the reply block
- **THEN** the evidence of `analysis_reply_block` says that the `report-endings` rule was read

#### Scenario: A filtered read and a partial read in the same cell

- **GIVEN** a Codex `exec` cell whose first command greps the headings and end line of a stage packet and whose second prints part of it
- **WHEN** the audit judges the leaves that packet holds
- **THEN** neither command counts as a whole read

### Requirement: The session audit judges the Try it block on the close report

`try_it_block_in_report` SHALL judge the last assistant message of the last turn that loaded `flow-close`, through the skill call, a read of its installed `SKILL.md` or a shell read of it, a `sed` range included, after the interface change, when the session has one and no later interface change that reaches `flow-verify` follows that turn; and the session's last assistant message otherwise.

#### Scenario: Other work after the close report

- **GIVEN** a session that changes a screen, verifies it, loads `flow-close` with a `sed` range and writes a close report without Try it, then works on something else
- **WHEN** the audit runs on it
- **THEN** `try_it_block_in_report` fails with the line of the close report
