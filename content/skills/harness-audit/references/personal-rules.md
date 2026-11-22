# Define and deploy the user's personal rules

Read from [harness-audit](../SKILL.md) when the user asks to define, change, sync, check or remove their own rules for every host, such as how they want work placed or reviewed in any project. This mode replaces the audit procedure.

Personal rules are the user's own instructions for every project on this machine. They live in one source, `rules.md` in the Hive configuration folder (`$XDG_CONFIG_HOME/hive/` when that variable is an absolute path, `~/.config/hive/` otherwise), and [personal_rules.py](../scripts/personal_rules.py) copies them to each host where Hive is installed at user scope:

Targets: a marked section in `$CLAUDE_CONFIG_DIR/CLAUDE.md` (default `~/.claude/CLAUDE.md`, which Grok shares), in `$CODEX_HOME/AGENTS.md` (`AGENTS.override.md` when it is not empty; default `~/.codex/`), in `$PI_CODING_AGENT_DIR/AGENTS.md` (default `~/.pi/agent/`) and in `$XDG_CONFIG_HOME/opencode/AGENTS.md` (default `~/.config/opencode/`); the file `~/.cursor/rules/personal-rules.mdc` (always applied) for Cursor; and `~/.gemini/config/rules/personal-rules.md` (`trigger: always_on`) for Antigravity.

Each section sits right after Hive's block (or its voice block when present), between `PERSONAL RULES` markers, and opens with a fixed line: the user's rules take precedence over Hive's guidance, its Scope and authorization rules included, except Hive's Secrets rules, which still hold, and a project's own instructions hold over them for that project.

## Who writes them

Write or sync personal rules only in this mode, when the user asks for it, after the user approves the exact text in the conversation. The user's request is the explicit instruction the global guidance requires for a global configuration change; its targets are `rules.md`, the marked sections and the two rule files above, and nothing else outside the project. Never write them in another flow, on your own initiative, or under [unattended delegation](../../unattended-delegation/SKILL.md); a preference stated in passing during other work is not this request: offer this mode as a reply instead.

## Steps

1. **Read.** Read `rules.md` when it exists, and run `python3 -I <harness-audit-dir>/scripts/personal_rules.py check` to see which copies are missing, differ from it, are out of position, or sit in a host that no longer has Hive.
2. **Define.** Ask what the rule must make an agent do, when, and what it must not do; one open question per message. Write the rule as an instruction with its condition and an observable test, in English, without a secret value. Keep the whole file at or under 8,000 bytes: it enters every session on every host.
3. **Approve.** Show the full new text and, when `rules.md` exists, its diff. Write `rules.md` only after the user approves that text; a correction is not approval, so show the corrected text again.
4. **Deploy.** Run `python3 -I <harness-audit-dir>/scripts/personal_rules.py deploy`; it refuses, naming the cause and writing nothing, an oversized or secret-looking source, a symbolic-link target and a `personal-rules` rule file it did not write: report a refused target to the user instead of moving or deleting it. When it names a file that changed since it was read, run it again.
5. **Check.** Run `python3 -I <harness-audit-dir>/scripts/personal_rules.py check` and report each host with its result. Exit 0 means every copy matches `rules.md`.

To remove the rules, run `python3 -I <harness-audit-dir>/scripts/personal_rules.py remove` after the user confirms; it deletes each section and both rule files and leaves `rules.md` and Hive's blocks in place.

## Limits

- `hive update` keeps every byte outside Hive's block, so the sections survive it. When Hive's block was deleted by hand and reinstalled at the end of the file, the section is no longer right after it: `check` reports the position and `deploy` moves it.
- Uninstalling Hive leaves the sections and files in place, since they belong to the user; `check` names a copy in a host without Hive, or in a Codex file Codex no longer reads, and `remove` deletes it.
- Cursor loads `personal-rules.mdc` only in workspaces under the home folder, like Hive's own `hive.mdc`; the fallback pointer for other workspaces names only `hive.mdc`.
- An audit of a host's global layer reads one copy per host: copies in other hosts' files are not duplicates. A personal rule that conflicts with Hive's Secrets rules is a finding, since those still hold; one that conflicts with Scope and authorization is a finding for the user to confirm, since the precedence line lets it override them; other conflicts follow the precedence line.
