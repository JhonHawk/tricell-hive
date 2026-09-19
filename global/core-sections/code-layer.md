---
order: 100
targets: [claude]
---

## Code Layer — identifiers always English

The `Spanish` rule above governs prose and UI strings; this governs code identifiers — a separate axis.

- **Identifiers are always English** — fields, types, functions, files, table/column names (`cashMethod`, `studentId`), and the identifiers inside specs: an OpenAPI property, schema field, or event key written in Spanish gets implemented verbatim and later costs a migration, not an edit.
- **Translate by domain meaning, not word-for-word:** pick the term a native practitioner of the domain uses (`accountsReceivable` for *cartera*, never `portfolio`); watch false friends.
- **Domain values MAY be Spanish** (enum values, RBAC keys, status constants: `efectivo`, `COLEGIATURA`) — consistent per bounded domain, never a mixed enum. i18n keys are English identifiers; the Spanish lives in the value.
- Judgment layer with the full examples and carve-out boundaries: `identifier-language.md` — read it via `language-rules` before naming; the `rule-delivery` hook holds a matching write until you do. Subagents carry it in their Role rules table.
