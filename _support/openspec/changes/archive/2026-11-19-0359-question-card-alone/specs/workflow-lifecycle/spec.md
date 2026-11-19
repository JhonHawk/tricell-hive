## ADDED Requirements

### Requirement: Pending questions are not duplicated across tools

The agent SHALL NOT issue the same pending question through both asynchronous and waiting tools, or issue another card while its answer is pending. A tool's acceptance receipt SHALL NOT count as a user answer.

#### Scenario: An asynchronous question is still pending

- **GIVEN** a permitted asynchronous question was issued
- **WHEN** its answer is still pending
- **THEN** the agent does not send the same question again through a waiting tool or a second card

### Requirement: Decision information survives a required change of form

When the host requires text instead of a card or another question form, the agent SHALL retain the alternatives, their consequences and the recommended option with its reason, expressed in the form the host permits. It SHALL NOT interpret a different form as removing that information or as permission to use a prohibited tool.

#### Scenario: The host requires a permission request in concise text

- **GIVEN** the host requires text and prohibits the waiting tool for that permission request
- **WHEN** the agent asks permission
- **THEN** it states the recommendation and why, and the alternative with its consequence in the permitted form
- **AND** it does not issue a prohibited or duplicate card

### Requirement: Decision rounds go through the card alone

When the selected native question tool is permitted for that question, waits for the answer and gives each option a description field, the agent SHALL ask a round of decisions or findings through that tool without writing them out in the text before it: the text before the first card SHALL hold the context line, the report or explanation the round needs and the waiting line, which says what is awaited and how many decisions are still to come, and SHALL NOT hold a decision's or a finding's title line, `Situation:` line, option lines or `I recommend` line. Each card SHALL stand alone: its question SHALL carry the decision's or finding's ID and title with the situation. Where the host requires text or no permitted waiting tool carries the question, or the selected card has no description field per option, the agent SHALL use the host-required form or, where the host permits it, write the whole round in the text; when a card that follows a round written in the text comes back without an answer, the agent SHALL add one line naming the decisions still open and SHALL NOT write the round again. A plain acceptance SHALL apply the starred options only of decisions the user was shown, in a card that opened or in the text.

#### Scenario: A round answered through the card

- **GIVEN** a host whose permitted question tool waits and shows a description per option
- **WHEN** the agent asks a round of three decisions
- **THEN** the text before the card holds the context line, the report and the waiting line, with no decision block
- **AND** the card asks the three decisions and the turn ends with it

#### Scenario: A card without option descriptions

- **GIVEN** a host whose permitted question tool has no description field per option and permits the full round in text
- **WHEN** the agent asks a round
- **THEN** the whole round is written in the text, and the card follows it
- **AND** when that card comes back without an answer, one line names the decisions still open and the round is not written again

### Requirement: An unanswered card is handled by who left it unanswered

The host's required response SHALL take precedence. Where the host leaves that response to Hive, when a waiting card that carried the round alone comes back without an answer, the agent SHALL ask no further card of that round before acting as follows. The answers the card did return SHALL stand. When the card closed by itself, because the tool returned an empty answer or a timeout while the turn continued and its result held no rejection, dismissal or decline, the agent SHALL write, in its next message, every decision and finding of the round still open in its full format, those of the cards not yet asked included, with a waiting line saying that "ok" applies every starred option and that a reply such as `D2 B` changes one. When the user left it unanswered, because the tool returned a rejection, a dismissal or a decline, also as an error and also on a host that returns a decline after its own timeout, or the host ended the turn while the card was open, or the card returned a request to talk, answers to only some of its questions, or free text that adds, removes or replaces a decision or an option, the agent SHALL NOT write the round out in the text on its own and SHALL act on what the user said, in the card or in their next message when the host ended the turn, by the first of these that fits: answers by ID or a plain acceptance are applied and the decisions left open are asked again through the card; a cancel made by mistake, or a request for the questions, gets the same round again through the card with the same IDs; a statement that the direction is wrong, a change to what the decisions rest on, or free text that adds, removes or replaces a decision or an option, gets the open decisions it touches dropped, reworked and asked as a new round with new IDs; a request to drop the questions gets them dropped and not asked again, except one a later step cannot proceed without; a change of subject, a question or an investigation that leaves the decisions as they were gets its answer followed by one line naming the decisions still unanswered and how to get them back, which are asked again through the card when the user says so or asks to continue work that cannot go on without them. When the user has said nothing about the open decisions yet, as after a decline while the turn continues, a card answered in part or a request to talk that names no subject, the agent SHALL reply with that one line alone and wait. A decision asked again SHALL keep its ID, and one line saying which decisions return SHALL precede a card that returns. Later rounds SHALL go through the card again.

#### Scenario: A card closes by itself

- **GIVEN** a card that returns empty after the host's timeout and no native instruction requiring another response
- **WHEN** the agent replies
- **THEN** its message writes each decision still open with its situation, its options and its recommendation
- **AND** it asks none of that round's remaining cards
- **AND** the next round goes through the card

#### Scenario: The user cancels by mistake

- **GIVEN** the user dismissed a card and the host ended the turn
- **WHEN** the user's next message says it was a mistake or asks for the questions again
- **THEN** the agent asks the same round again through the card, with the same IDs and options

#### Scenario: The user cancels because the direction is wrong

- **GIVEN** the user dismissed a card
- **WHEN** the user's next message says the work is going in the wrong direction
- **THEN** the agent drops the open decisions, reworks them with what the user said and asks a new round with new IDs

#### Scenario: The user cancels and changes the subject

- **GIVEN** the user dismissed a card
- **WHEN** the user's next message asks a question that leaves the decisions as they were
- **THEN** the agent answers it and ends the reply with one line naming the decisions still unanswered
- **AND** it does not write the round out in the text
- **AND** it asks the round again through the card when the user says so

#### Scenario: The user cancels and answers by ID

- **GIVEN** the user dismissed a card of three decisions
- **WHEN** the user's next message answers two of them by their ID
- **THEN** those two answers apply and the third is asked again through the card

#### Scenario: The user cancels and asks to drop the questions

- **GIVEN** the user dismissed a card
- **WHEN** the user's next message says to drop those questions
- **THEN** the agent drops them and does not ask them again, unless a later step cannot proceed without one

#### Scenario: Free text replaces an option

- **GIVEN** a card whose answer is free text that replaces one option of a decision
- **WHEN** the agent replies
- **THEN** it reworks that decision with what the user said and asks it again through the card with a new ID

#### Scenario: The host directs continuation after an unanswered optional question

- **GIVEN** the host requires best judgment after an unanswered optional waiting card
- **WHEN** that card returns no answers
- **THEN** the agent follows that instruction without re-asking the question
- **AND** the empty response grants no permission for an action requiring authorization
