## MODIFIED Requirements

### Requirement: Decision rounds go through the card alone

When the selected native question tool waits for the answer and gives each option a description field, the agent SHALL ask a round of decisions or findings through that tool without writing them out in the text before it, also on a host whose own instructions reserve that tool for optional questions, such as Codex in its Default mode: the text before the first card SHALL hold the context line, the report or explanation the round needs and the waiting line, which says what is awaited and how many decisions are still to come, and SHALL NOT hold a decision's or a finding's title line, `Situation:` line, option lines or `I recommend` line. Each card SHALL stand alone: its question SHALL carry the decision's or finding's ID and title with the situation, or the fixed header of a session question. While Codex offers `request_user_input`, the agent SHALL NOT ask a round through `request_user_input_async`, and SHALL ask the delivery, review and validation questions and every other decision through `request_user_input` although the Default-mode instructions send required input to one plain-text question; only the host's sandbox or approval escalations follow those instructions' form for permission requests. Where no waiting tool carries the question, or the selected card has no description field per option, the agent SHALL write the whole round in the text in its full format; when a card that follows a round written in the text comes back without an answer, the agent SHALL add one line naming the decisions still open and SHALL NOT write the round again. A plain acceptance SHALL apply the starred options only of decisions the user was shown, in a card that opened or in the text.

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

#### Scenario: Codex offers both question tools

- **GIVEN** a Codex session in Default mode that offers `request_user_input` and `request_user_input_async`
- **WHEN** the user asks the agent to send the pending delivery and review questions
- **THEN** it asks them through `request_user_input`
- **AND** it does not call `request_user_input_async` for them

#### Scenario: A card without option descriptions

- **GIVEN** a host whose question tool has no description field per option and permits the full round in text
- **WHEN** the agent asks a round
- **THEN** the whole round is written in the text, and the card follows it
- **AND** when that card comes back without an answer, one line names the decisions still open and the round is not written again

### Requirement: The session audit reads an imperative confirmation and quoted replies as prose choices

`prose_choice_without_card` SHALL also read as a question a sentence with no question mark that opens with, or follows a comma or colon with, an imperative confirmation such as «confirma», «confírmame» or "confirm", and SHALL fail it when it names a session choice and no card follows in its turn. It SHALL fail a sentence with no question mark that recommends, such as «recomiendo» or "I recommend", outside quotation marks, a code span that starts with a delivery mode (`interactive`, `automatic`, `ready`, `hold` or `direct-base`) as a whole word when no card follows in its turn, judging each sentence on its own, and SHALL report review for a sentence that names a session choice as still to confirm, such as «queda por confirmar la entrega» or "the delivery still needs confirming", when no card follows in its turn. It SHALL report review for a message that offers two or more replies in quotation marks after a phrase such as «puedes responder», «responde» or "you can reply", when no card follows in its turn.

#### Scenario: Delivery asked as an imperative confirmation

- **GIVEN** a Codex session with the card offered, whose message says «Para arrancar, confirma `en el checkout actual, entrega local (hold)`.» and ends its turn without a card
- **WHEN** the audit runs on it
- **THEN** `prose_choice_without_card` fails with that message's line

#### Scenario: Delivery left pending with a recommendation

- **GIVEN** a Codex session with the card offered, whose message says «Queda por confirmar la entrega y las revisiones de la candidata. Recomiendo **`automatic, Codex, sí`** para integrar a `development` tras CI.» and ends its turn without a card
- **WHEN** the audit runs on it
- **THEN** `prose_choice_without_card` fails with that message's line

#### Scenario: Delivery left pending without a recommendation

- **GIVEN** a message that says «Queda por confirmar la entrega.» with no card after it and no other prose choice in the session
- **WHEN** the audit runs on it
- **THEN** `prose_choice_without_card` reports review

#### Scenario: Replies offered as quoted phrases

- **GIVEN** a message that says «Puedes responder «apruebo», «corrige el filtro» o «abre el visor»» with no card after it and no other prose choice in the session
- **WHEN** the audit runs on it
- **THEN** `prose_choice_without_card` reports review

#### Scenario: A sentence that only reports a confirmation

- **GIVEN** a message that says «El log confirma que la entrega terminó.» or «Registré la entrega automática a `development`.»
- **WHEN** the audit runs on it
- **THEN** `prose_choice_without_card` does not count it

### Requirement: The session audit fails a plan that ends without a build route

`go run ./tests/pilot --audit-session` SHALL report `plan_without_build_route` for Claude, Codex and Grok sessions. A plan stage SHALL run from a load of `flow-plan`, through the skill call, a Codex user message that carries the skill as typed by the user, such as `$flow-plan`, or a read of its installed `SKILL.md`, to the next load of a different flow's skill or the end of the session; a further load of `flow-plan`, a search event or a `grep` or `rg` of a flow's `SKILL.md` SHALL NOT end it. The criterion SHALL pass a stage in which an assistant message holds the command of `flow-build` or `create-design` in any host's syntax, such as `/flow-build`, `$flow-build` or `**/flow-build**`, together with a `▶` line; SHALL fail a stage in which the command appears only in messages without a `▶` line at the start of a line, giving the line of the first such message, unless one of the review cases below for a stage another flow load ends applies; SHALL fail a stage that a later flow load ends with no such message, giving the line of the `flow-plan` load, except that it SHALL report review when that load is in the turn of the stage's latest `flow-plan` load with no question card between them, a continuation the user may have asked for, or follows a load of the installed `unattended-delegation` skill; SHALL report review for a stage the end of the session ends with no such message, since the session may still be open; SHALL report not_observed when the session loads no `flow-plan`; and SHALL NOT judge a child session.

#### Scenario: A plan kept in the conversation that never names the build command

- **GIVEN** a Codex session that loads `flow-plan`, writes the plan in a message with no flow command, and later loads `flow-build` after the user answers a delivery round
- **WHEN** the audit runs on it
- **THEN** `plan_without_build_route` fails with the line of the `flow-plan` load

#### Scenario: A plan that ends with the next-step block

- **GIVEN** a session whose plan message ends with a `▶` block naming `/flow-build`
- **WHEN** the audit runs on it
- **THEN** `plan_without_build_route` passes

#### Scenario: A plan route without the next-step block

- **GIVEN** a Codex session whose user types `$flow-plan` and whose plan message ends «**Siguiente paso · Implementar COM-933**» followed by `` `$flow-build COM-933` `` with no `▶` line
- **WHEN** the audit runs on it
- **THEN** `plan_without_build_route` fails with that message's line

#### Scenario: A plan and its build in one turn

- **GIVEN** a session whose user asks to plan and build without stopping, and that loads `flow-plan` and then `flow-build` in the same turn with no flow command in its messages
- **WHEN** the audit runs on it
- **THEN** `plan_without_build_route` reports review

#### Scenario: A message that names only the plan command

- **GIVEN** a plan stage whose only flow command in an assistant message is `/flow-plan`, ended by a load of `flow-build`
- **WHEN** the audit runs on it
- **THEN** `plan_without_build_route` fails

## ADDED Requirements

### Requirement: The session audit fails an asynchronous round while the waiting card is offered

`go run ./tests/pilot --audit-session` SHALL report `async_card_when_waiting_offered` for Codex sessions. For each call to `request_user_input_async` in a main session it SHALL fail when the session offers `request_user_input` in any of its turns, and SHALL fail when it does not and no assistant text earlier in the same turn is a round written in the text; a call at or after the start of unattended delegation SHALL fail only when no round is written in the text earlier in its turn; otherwise it SHALL pass. It SHALL report not_observed for a session with no such call, SHALL NOT judge a child session, and its evidence SHALL name the line, never the text.

#### Scenario: Async card with the waiting card offered

- **GIVEN** a Codex session in Default mode that offers `request_user_input` and calls `request_user_input_async` for the delivery, review and security questions
- **WHEN** the audit runs on it
- **THEN** `async_card_when_waiting_offered` fails with the line of that call

#### Scenario: Async card as the only tool, round written first

- **GIVEN** a Codex session that offers only `request_user_input_async` and writes the round in the text in full before calling it in the same turn
- **WHEN** the audit runs on it
- **THEN** `async_card_when_waiting_offered` passes

#### Scenario: Async card as the only tool, no round in the text

- **GIVEN** a Codex session that offers only `request_user_input_async` and calls it after a one-line message
- **WHEN** the audit runs on it
- **THEN** `async_card_when_waiting_offered` fails with the line of that call
