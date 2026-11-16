## ADDED Requirements

### Requirement: Decision rounds go through the card alone

On a host whose native question tool waits for the answer and gives each option a description field, the agent SHALL ask a round of decisions or findings through that tool without writing them out in the text before it: the text before the first card SHALL hold the context line, the report or explanation the round needs and the waiting line, which says what is awaited and how many decisions are still to come, and SHALL NOT hold a decision's or a finding's title line, `Situation:` line, option lines or `I recommend` line. Each card SHALL stand alone: its question SHALL carry the decision's or finding's ID and title with the situation. When such a card comes back without an answer (empty, dismissed, rejected, including a dismissal the tool reports as an error, answered only in part, with a request to talk, or with free text that adds, removes or replaces a decision or an option), the agent SHALL ask no further card of that round and SHALL, in its next message, first answer what the user said when there was a request to talk or free text, and then write every decision and finding of the round still open in its full format, those of the cards not yet asked included, with a waiting line saying that "ok" applies every starred option and that a reply such as `D2 B` changes one, to be answered in the chat. When the host ends the turn with the dismissal, whether or not the tool returns a result, the agent SHALL write the open round in reply to the user's next message, after answering what that message says, unless it answers each open decision by its ID or sets the round aside. A plain acceptance SHALL apply the starred options only of decisions the user was shown, in a card that opened or in the text. Later rounds SHALL go through the card again. Where no waiting tool carries the question, or the tool has no description field per option, the agent SHALL write the whole round in the text; when a card that follows a round written in the text comes back without an answer, the agent SHALL add one line naming the decisions still open and SHALL NOT write the round again.

#### Scenario: A round answered through the card

- **GIVEN** a host whose question tool waits and shows a description per option
- **WHEN** the agent asks a round of three decisions
- **THEN** the text before the card holds the context line, the report and the waiting line, with no decision block
- **AND** the card asks the three decisions and the turn ends with it

#### Scenario: A card closes without an answer

- **GIVEN** a card that returns empty after the host's timeout
- **WHEN** the agent replies
- **THEN** its message writes each decision still open with its situation, its options and its recommendation
- **AND** it asks none of that round's remaining cards
- **AND** the next round goes through the card

#### Scenario: The host interrupts the turn

- **GIVEN** the user dismissed a card and the host ended the turn, with or without a tool result
- **WHEN** the user's next message does not answer the round
- **THEN** the agent answers that message and then writes the open round in the text

#### Scenario: A card without option descriptions

- **GIVEN** a host whose question tool has no description field per option
- **WHEN** the agent asks a round
- **THEN** the whole round is written in the text, and the card follows it
- **AND** when that card comes back without an answer, one line names the decisions still open and the round is not written again
