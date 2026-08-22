---
order: 90
targets: [claude]
---

## Spanish
- **User-facing output is Spanish in EVERY context — subagents and forked skills included.** Subagents never receive the session's `language` setting (docs: sub-agents inherit CLAUDE.md, not the full system prompt), so this line is what binds them; it matters most for forks whose report surfaces verbatim (e.g. `status-fetch`). Agent-to-agent payloads and config files keep their own rules (English). Residual gap: built-in Explore/Plan agents skip CLAUDE.md — their output is main-thread-consumed anyway.
- **Orthography is mandatory.** Always include proper accents (á, é, í, ó, ú, ñ, ü) in user-facing strings, error messages, labels, and comments written in Spanish. Watch the high-frequency misses: `reservación`, `vehículo`, `sesión`, `información`, `número`.
- **Dialect: Mexican Spanish.** Default to `tú` (informal) for all conversational communication with the user. Use `usted` only for formal client-facing copy (cotizaciones, RFP responses, proposals, formal emails to unknown audiences). Never use `vos` — voseo is not Mexican Spanish and reads as foreign.
- **Lexical preference:** when a Mexican-standard term differs from another regional term, prefer the Mexican one (e.g., `computadora` over `ordenador`). When both are accepted in Mexican usage, follow the user's lead.
