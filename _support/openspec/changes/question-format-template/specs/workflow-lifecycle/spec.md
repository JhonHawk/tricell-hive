## MODIFIED Requirements

### Requirement: Decisions carry labelled context

A decision `D<n>` shown to the user, and a session question shown under its fixed header, SHALL carry, under its title line, a `Situation:` line that says what happened or what was found, what was verified and what to check, with its target, in words the user has used or that the same message explains; under each option, an `If you choose it:` line that says what the user gets and what it costs them in time, money, risk or what stops working; and an `I recommend <letter>:` line whose reason is a consequence for the user and never a quotation of a rule or a skill. The three labels SHALL be written in the session language, while the IDs, fixed headers, letters and the `(★)` mark stay as they are. Each option SHALL open its own line as `<letter> - <name>`, with `(★)` right after the dash on the recommended one, or «(Recommended)», the marker a host's card asks for, at the end of the label, and no list marker before the letter. In a round written in the text, the decisions SHALL be separated by a `---` rule. The round's waiting line SHALL also say how many known decisions are still to come. In a native question card, the question SHALL carry the situation, each option's label SHALL be `<letter> - <name>` with the recommended one first and marked, each option's description the text of its `If you choose it:` line, or the option's effect in one line when it has none, and the starred option's description SHALL end with the reason of the recommendation. The "Open the viewer" option that a validation question carries SHALL have no `If you choose it:` line. A finding `H<n>` asked in a round carries its title and option lines without the three labelled lines. Every block of lettered replies the user answers, the analysis reply block, the next-step block and the close footer, SHALL open with a bold heading `▶ D<n> — <heading>` from the same series as the decisions, so its replies are cited as `D<n> <letter>`; the ID only names the block, and a reply to a next-step block keeps that block's effect. A close footer with no reply to offer keeps `▶ Closed` without an ID or letters.

#### Scenario: A rule requires the question

- **GIVEN** a rule requires the agent to ask before an effect
- **WHEN** the agent asks it as a decision
- **THEN** the recommendation line states the effect that needs the answer, such as its cost or what it changes
- **AND** it quotes no rule or skill as its reason

#### Scenario: A name the user was not given

- **WHEN** the situation would name a task ID, a file, a branch or an internal name the user has not been given
- **THEN** the same message explains it, or the situation leaves it out

#### Scenario: Two decisions written in the text

- **GIVEN** no waiting card carries the round
- **WHEN** the agent writes two decisions in the text
- **THEN** each option line opens with its letter and a dash, with no list marker before it
- **AND** a `---` rule separates the two decisions

#### Scenario: An analysis left open is cited later

- **GIVEN** an analysis ended with `▶ D5 — How do we continue?` and the user answered another question
- **WHEN** a later message refers to that open question
- **THEN** it names it as D5

## ADDED Requirements

### Requirement: The session audit flags card options out of format

`go run ./tests/pilot --audit-session` SHALL report `card_option_format` for Claude, Codex and Grok sessions, judging every native question card that waits for the answer, answered or not; Codex's `request_user_input_async` is not such a card. It SHALL fail a card when one of its option labels does not open with `<letter> - `; for a card of a single choice it SHALL also fail when no option carries the recommended mark, or when the mark sits on an option other than the first or on more than one. The recommended mark SHALL be `(★)` right after `<letter> - `, or «(Recommended)», in any letter case, at the end of the label. The evidence SHALL give the card's line and the question's number in the card, and never the question's text or its ID. A session with no question card SHALL be `not_observed`, and a child session SHALL not be judged.

#### Scenario: A card whose options lost their letters

- **WHEN** a card offers `Worktree propio (★)` and `Checkout principal`
- **THEN** `card_option_format` fails at that card's line and gives the question's number, without its text

#### Scenario: A star quoted inside another option

- **WHEN** a card offers `A - (★) Letra fija` and `B - Estricto: solo A - (★)`
- **THEN** `card_option_format` passes

#### Scenario: A card in its format

- **WHEN** every card of the session offers options such as `A - (★) Aprobar` and `B - Pedir correcciones`
- **THEN** `card_option_format` passes

#### Scenario: The host's recommended marker

- **WHEN** a Codex card offers `A - Siempre (Recommended)` and `B - Visibilidad actual`
- **THEN** `card_option_format` passes
