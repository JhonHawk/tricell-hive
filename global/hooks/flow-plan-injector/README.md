# flow-plan-injector

UserPromptSubmit hook (Claude Code only), gated to `permission_mode == "plan"` AND a flow
workspace (ledger `_support/PROJECT.md`), once per session, ~8 factual lines: capture is
automatic on approval, the `Session: yes/no` opt-out line, initiative shaping for
multi-session scope, the findings convention, and no author-time git semantics (that delta
is `/flow-build`'s adoption gate).

Supersedes the retired `flow-route-reminder` (2026-07-09 → 2026-07-10): its per-prompt
routing payload had documented failure modes (ignored injected instructions, false
positives, token pollution — anthropics/claude-code #19659/#17804); its mechanics (ledger
walk-up, stdin cwd, once-per-session marker, exit 0) live on here and in
`flow-plan-capture`. If the `permission_mode` field is absent from the payload, the
injector stays silent by design — capture works without it.

Deployed by `/deploy-global` (script → `~/.claude/hooks/`, block merged into
`settings.json` from `settings-config.json`).
