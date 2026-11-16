## ADDED Requirements

### Requirement: Decision rounds go through the card alone

On a host whose native question tool waits for the answer and gives each option a description field, the agent SHALL ask a round of decisions or findings through that tool without writing the decisions in the text before it: the text before the first card SHALL hold the context line, the report or explanation the round needs and the waiting line, which says what is awaited and how many decisions are still to come, and SHALL NOT hold a decision's title line, its `Situation:` line, its option lines or its `I recommend` line. Each card SHALL stand alone, as the format of a decision sets. When a card comes back without an answer (empty, dismissed, rejected, with a request to talk, or with free text that adds, removes or replaces a decision or an option), the agent SHALL ask no further card of that round and SHALL write, in its next message, every decision of the round still open in the full decision format, with a waiting line saying that "ok" applies every starred option and that a reply such as `D2 B` changes one, to be answered in the chat. When the host interrupts the turn while a card is open, so that the agent receives no result, the agent SHALL write the open round in reply to the user's next message, unless that message answers it. Later rounds SHALL go through the card again. Where no waiting tool carries the question, or the tool has no description field per option, the agent SHALL write the whole round in the text.

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

- **GIVEN** the user dismissed a card and the host ended the turn without a tool result
- **WHEN** the user's next message does not answer the round
- **THEN** the agent writes the open round in the text before anything else

#### Scenario: A card without option descriptions

- **GIVEN** a host whose question tool has no description field per option
- **WHEN** the agent asks a round
- **THEN** the whole round is written in the text, and the card follows it
