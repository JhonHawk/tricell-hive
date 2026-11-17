## ADDED Requirements

### Requirement: Requirement documents are written and validated through one procedure

The workflow guidance SHALL keep the procedure for writing, restructuring, updating and technically reviewing a requirement document in the `create-prd` skill. The skill SHALL write the document with these sections in order: problem, goals, non-goals, user stories, requirements, acceptance criteria, success metrics, open questions and version history, and SHALL NOT write interface design, the technical how or a schedule. Each requirement SHALL be one rule on one line with a stable code that is never reused, a priority of P0, P1 or P2, a state (decided, assumption, to decide or cancelled), its source and its verdict against the code (exists, new, contradicts, blocked or not checked). The skill SHALL ask at most three questions per round and SHALL never assume a business rule the user does not decide. Used on its own, it SHALL end at the document and only offer the next step. `flow-research` SHALL serve a request whose deliverable is a requirement document by following the skill in the same turn, whether or not it asked questions first, and the global guidance SHALL name that hand-over among the continuations that need no reply from the user.

#### Scenario: A requirement is drafted
- **WHEN** the user asks for the requirement of a module or feature
- **THEN** the skill reads the source, asks at most three questions per round, writes the document in the fixed sections with each requirement as a rule line, and reports its link, its state, its open questions with who answers each and its assumptions.

#### Scenario: A request enters through research
- **WHEN** a request whose deliverable is a requirement document reaches `flow-research`
- **THEN** research follows the `create-prd` skill in the same turn, and the skill asks its own questions.

#### Scenario: A document arrives already written
- **WHEN** the user hands over a requirement document someone else wrote
- **THEN** the skill returns it in the fixed sections without changing any of its author's rules, with each rule pointing to where it stood in the original, delivers a questionnaire for the author with a recommended answer per question, and never edits the original.

#### Scenario: An undecided business rule
- **WHEN** a rule belongs to someone else or the user does not know it yet
- **THEN** the rule is written as to decide, with who answers it, and is never filled in.

#### Scenario: A correction arrives
- **WHEN** answers or corrections change the document
- **THEN** the skill writes a new version with its list of changes, and a cancelled rule keeps its line with the version that cancelled it and the rule that replaces it.

### Requirement: A requirement document's state is set by validation

The state of a requirement document SHALL be one of draft, iterating, approved by business, technical review and ready for development, and SHALL be set by what validation found, never by its author's label. The document SHALL reach approved by business only when the fixed checklist passes, an independent `hive-review-prd` child has reviewed it, no rule is still to decide and the user approves it explicitly. It SHALL reach ready for development only through the technical review, run in a session with the repository, when no rule is not checked and every rule the code contradicts or blocks has its owner's answer. A technical reviewer who is not the author SHALL NOT rewrite a rule.

#### Scenario: Approval by business
- **WHEN** the checklist passes, the independent review has run, no rule is to decide and the user approves
- **THEN** the document is approved by business and that version is frozen.

#### Scenario: A rule is still to decide
- **WHEN** the user asks to approve a document that holds a rule still to decide
- **THEN** the document stays iterating and the skill reports the rule and who answers it.

#### Scenario: Technical review with the repository
- **WHEN** a session with the repository takes a version approved by business
- **THEN** the skill gives each rule a verdict against the code, writes a question with a recommended answer for each contradiction or block, and marks the document ready for development only when none is unanswered and none is not checked.

#### Scenario: A session without the repository
- **WHEN** a session without the repository is asked to mark a document ready for development
- **THEN** the skill does not, and says the technical review needs a session with the repository.

### Requirement: A requirement document lives in a shared document and its approved versions are frozen

The skill SHALL write the document in a document tool the session can write to, at the destination the user names, asking for it when none was named, and SHALL name for Notion, Google Docs and Claude Docs how to create, copy, comment and obtain the link. Without such a tool it SHALL write the document as a Markdown file, at `_support/docs/requirements/<module>/prd.md` inside a repository and in the working folder outside one, and render it to PDF beside it with a browser already installed, installing nothing. The rendering script SHALL treat the Markdown file as untrusted text: it SHALL escape any HTML the file carries, emit a page that loads no remote resource and runs no script, and launch the browser with a temporary profile. At each approval it SHALL freeze the version as a copy in the same tool, titled with its version, state and date, which it never writes again, and link it from the live document's version history; in the Markdown mode the copy SHALL be `prd-v<N>.md` with its PDF.

#### Scenario: A document tool is connected
- **WHEN** the session can write to a document tool
- **THEN** the document is created there and each approved version is a titled copy linked from the live document's history.

#### Scenario: No document tool, inside a repository
- **WHEN** the session has no document tool and works inside a repository
- **THEN** the document is written to `_support/docs/requirements/<module>/prd.md` and each approved version to `prd-v<N>.md` with its PDF beside it, without a commit, and the skill reports the path and that it is not committed.

#### Scenario: A document that carries HTML
- **WHEN** the Markdown file holds a script tag or a remote image
- **THEN** the PDF shows them as text, nothing is fetched and no script runs.

#### Scenario: No installed browser
- **WHEN** the PDF is rendered on a machine with no Chromium-family browser
- **THEN** the script leaves the printable page, exits with a non-zero code and says how to print it.
