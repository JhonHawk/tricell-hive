## ADDED Requirements

### Requirement: Personal rules mode of harness-audit

When the user asks to define, change, sync, check or remove their own rules for every host, `harness-audit` SHALL run its personal rules mode. The agent SHALL help the user write the rule, SHALL show the full text and, for an existing source, its diff, and SHALL write the source `rules.md` in the Hive configuration folder (`$XDG_CONFIG_HOME/hive/` when that variable is set to an absolute path, `~/.config/hive/` otherwise) only after the user approves that text in the conversation. The user's request to this mode SHALL count as the explicit instruction the global guidance requires for a global configuration change, and `harness-audit`'s write scope SHALL name this mode's targets as its exception: `rules.md`, the marked section in each shared global file, and the two personal rule files. The agent SHALL NOT write or sync personal rules in any other flow, on its own initiative, or under unattended delegation.

The mode's script `harness-audit/scripts/personal_rules.py` SHALL deploy the source to each host where Hive is installed at user scope, found by Hive's own managed content in that host's global instructions: in the shared global files of Claude Code (`$CLAUDE_CONFIG_DIR/CLAUDE.md`, default `~/.claude/CLAUDE.md`) and Grok (always `~/.claude/CLAUDE.md`), Codex (`$CODEX_HOME/AGENTS.md`, default `~/.codex/AGENTS.md`), Pi (`$PI_CODING_AGENT_DIR/AGENTS.md`, default `~/.pi/agent/AGENTS.md`) and OpenCode (`$XDG_CONFIG_HOME/opencode/AGENTS.md`, default `~/.config/opencode/AGENTS.md`), as a section between its own markers placed immediately after Hive's voice block when present and otherwise immediately after Hive's block; in Cursor as the file `~/.cursor/rules/personal-rules.mdc` with `alwaysApply: true` frontmatter; and in Antigravity as the file `~/.gemini/config/rules/personal-rules.md` opening with `trigger: always_on` frontmatter. Each deployed copy SHALL open with this line, verbatim, followed by a blank line and the source text:

`These are the user's own rules for every project on this machine. Where they conflict with Hive's guidance, including its Secrets and Scope and authorization rules, these rules take precedence; where they conflict with a project's own instructions, the project's instructions hold for that project.`

Running the deploy again SHALL replace each section or file in place, never add a second one, and SHALL leave every byte outside its markers unchanged; when a section no longer sits right after Hive's block (or voice block), the deploy SHALL move it there. Just before replacing a file, the deploy SHALL read it again and, when it changed since its first read, SHALL stop without writing that file, name it, and exit non-zero after finishing the other files, so the mode runs the deploy again. A check run SHALL exit non-zero and name each host whose copy is missing, differs from the source, or is out of position, and each personal copy found in a host that no longer carries Hive's content. A remove run SHALL delete each section and both files. Before writing anything, the script SHALL refuse a source over 8,000 bytes, a source containing its own markers or Hive's managed-block markers, or a source with a line the secret check of `local_instructions.py` (`line_has_secret`) flags, naming the file and the cause. Hive's manager preserves every byte outside its managed blocks across `hive update`, so the section's content survives it; its position after a hand-deleted Hive block is reinstalled at the end of the file is not guaranteed, and the check reports it.

#### Scenario: Rules reach every installed host
- **GIVEN** the user approved the text and Hive is installed for all seven hosts
- **WHEN** the mode runs the deploy
- **THEN** the four shared files hold the personal section right after Hive's block (or voice block), opening with the precedence line
- **AND** `~/.cursor/rules/personal-rules.mdc` and `~/.gemini/config/rules/personal-rules.md` hold the same text with their frontmatter

#### Scenario: An edited source is deployed again
- **WHEN** the user approves a change to `rules.md` and the deploy runs again
- **THEN** each copy is replaced in place, no copy is duplicated, and bytes outside the markers are unchanged

#### Scenario: Hive updates after a deploy
- **WHEN** the user runs `hive update` after the deploy
- **THEN** every personal section is still present and unchanged, and the check exits 0

#### Scenario: Hive was removed from a host
- **GIVEN** a deploy wrote `~/.cursor/rules/personal-rules.mdc` and Hive was later uninstalled from Cursor
- **WHEN** the check runs
- **THEN** it exits non-zero naming that file as a copy outside the deployed hosts

#### Scenario: A copy drifted
- **GIVEN** the section in `~/.codex/AGENTS.md` was edited by hand
- **WHEN** the check runs
- **THEN** it exits non-zero naming Codex

#### Scenario: The source is refused
- **WHEN** `rules.md` holds 8,001 bytes, a Hive marker, or a token `line_has_secret` flags (a prose identifier such as `autoMemoryDirectory` is not refused)
- **THEN** the deploy exits non-zero naming `rules.md` and the cause, and no file changes
