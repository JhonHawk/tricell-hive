## MODIFIED Requirements

### Requirement: The session audit flags an analysis that ends without the reply block

`go run ./tests/pilot --audit-session` SHALL report `analysis_reply_block` for Claude, Codex and Grok sessions. It SHALL fail a main-session turn whose last assistant text ends with two or more list items labeled in italics by something other than a single letter, or with lettered items and no line above them holding `▶` or a question in bold. It SHALL also fail a main-session turn that loaded `hive-mode`, through the skill call or a read of its installed `SKILL.md`, whose last assistant text offers no replies as a list, has six or more non-blank lines and holds no `▶`, no question in bold among its last three non-blank lines and no flow command, when that turn changed no file outside the host's temporary location and delivered no work. A file change counts through a write or delete tool the host did not refuse, an output redirection, an in-place `sed` or `perl` edit, `find -delete` or a writer run by `find -exec`, or the files `tee`, `mv`, `rm`, `rmdir`, `mkdir`, `touch` and `truncate` name and the destination of `cp`, `ln` and `install`, including behind `bash -c`, `env`, `sudo`, `nohup` or `xargs`. The shell line is read as tokens, so text inside quotes and the body of a here-document fed to a program other than a shell are not commands. A relative path resolves against an earlier `cd` in the same line, and a path it cannot resolve counts as outside. Delivering work means a git subcommand that changes the working tree, the history or a remote, `gh pr create`, `gh pr merge` or `git_state.py commit`. A shell command whose text the trace withholds entirely, or that cannot be tokenized, counts as work; one the trace keeps as a template is read with each interpolated value unknown. Its evidence SHALL say whether the installed `hive-rule-report-endings` was read earlier in the session. It SHALL NOT judge a turn that ends with a question card, a flow command or a decision round written in text, nor a child session's turns, and SHALL change no other criterion's result.

#### Scenario: Replies offered as an unlettered list
- **WHEN** an analysis ends with "Puedes responder:" followed by two bullets labeled in italics
- **THEN** `analysis_reply_block` fails at that line and says whether the report-endings rule was read

#### Scenario: The reply block in its format
- **WHEN** an analysis ends with a bold heading holding `▶` and lettered options with the recommended one starred
- **THEN** `analysis_reply_block` does not fail

#### Scenario: An analysis that offers no replies
- **WHEN** a turn that loaded `hive-mode` and edited no file ends a report of six or more lines with "Queda pendiente tu visto bueno" and no block
- **THEN** `analysis_reply_block` fails at that line, says the analysis ended without a reply block and whether the report-endings rule was read

#### Scenario: A short answer or a work turn
- **WHEN** a turn ends with fewer than six non-blank lines, changed a file outside the host's temporary location before its last text, for example with `sed -i`, or did not load `hive-mode`
- **THEN** `analysis_reply_block` does not judge that turn for a missing block
