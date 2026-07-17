# Release notes — client-facing template

Used by the QA/prod promotion walk (`promotion-playbook.md`) to draft the client-facing
release notes from merged PRs + tracker states since the last promotion. The draft is **always reviewed by
the user before anything is sent** — this template produces a draft, never an outbound
message.

## Rules

- **Format: plain markdown, versioned** at `<project>-specs/releases/YYYY-MM-DD-<env>.md`.
  The real delivery channel is text (email, WhatsApp, Slack), so it stays copy-paste ready —
  no HTML, no `flow-report` — but it is a **durable record, not a throwaway draft**: it lives
  in the specs repo, reviewed in place before sending and committed at close as the history
  of what shipped. Never `_support/workspace/` (gitignored). **No specs repo** (standalone
  single repo, or a workspace before its specs repo exists) → version it in the deploy
  session's location per the session-capture detection rule
  (`<repo>/_support/sessions/<slug>/release-notes-<env>.md`); `git mv` into
  `<project>-specs/releases/` once a specs repo exists.
- **Client language, not engineering language.** Map PR titles and tracker items to what
  the client observes: "Ahora puede filtrar reservaciones por fecha", not
  "feat(api): add date-range query params". Drop internal-only changes (refactors, CI,
  dependency bumps) unless they have a visible effect worth stating (e.g. "el sistema
  responde más rápido al cargar el listado").
- **Tone: usted**, Mexican Spanish, full orthography — per the global Spanish rules for
  formal client-facing copy.
- **Empty sections are omitted**, not left with "N/A". A QA promotion with only fixes
  ships only "Correcciones".
- **Per-project overrides** live in an optional `Release notes` row of the project's
  PROJECT.md header table (channel, language, tone). Absent row = the defaults above.
  Overrides change tone/language/channel — never the section structure.

## Template (client-facing sections stay in Spanish verbatim)

```markdown
# Notas de versión — <Proyecto>
**Fecha:** <YYYY-MM-DD> · **Ambiente:** <QA / Producción>

## Novedades
- <funcionalidad nueva descrita en el lenguaje del cliente, sin jerga técnica>

## Correcciones
- <comportamiento visible que se corrigió, descrito desde la perspectiva del usuario>

## Requiere su atención
- <acción que el cliente debe tomar: limpiar caché, nuevas credenciales, capacitación>

## Próximamente
- <qué entra en la siguiente promoción, solo si está comprometido — no promesas blandas>
```

## Sourcing map

| Section | Source |
|---|---|
| Novedades | Merged PRs with `feat:` prefix + tracker items moved to done since last promotion |
| Correcciones | Merged PRs with `fix:` prefix + resolved bug items |
| Requiere su atención | Migration notes, config/credential changes, anything from the deploy's pre-gates that touches the client |
| Próximamente | Tracker items committed for the next cycle — exclude anything carrying the project's `Deferred marker` |
