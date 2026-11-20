## ADDED Requirements

### Requirement: The session audit flags an analysis that ends without the reply block

`go run ./tests/pilot --audit-session` SHALL report `analysis_reply_block` for Claude, Codex and Grok sessions. It SHALL fail a main-session turn whose last assistant text ends with two or more list items labeled in italics by something other than a single letter, or with lettered items and no line above them holding `▶` or a question in bold, and SHALL say in its evidence whether the installed `hive-rule-report-endings` was read earlier in the session. It SHALL NOT judge a turn that ends with a question card, a flow command or a decision round written in text, nor a child session's turns, and SHALL change no other criterion's result.

#### Scenario: Replies offered as an unlettered list
- **WHEN** an analysis ends with "Puedes responder:" followed by two bullets labeled in italics
- **THEN** `analysis_reply_block` fails at that line and says whether the report-endings rule was read

#### Scenario: The reply block in its format
- **WHEN** an analysis ends with a bold heading holding `▶` and lettered options with the recommended one starred
- **THEN** `analysis_reply_block` does not fail
