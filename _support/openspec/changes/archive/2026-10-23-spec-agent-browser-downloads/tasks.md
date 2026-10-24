# Tareas · spec-agent-browser-downloads

Commit base: `165d457a` (`development`).

### T1 — Requisito de descargas al día

- [x] Delta MODIFIED escrito y aplicado al spec vigente.

**Closes:** AC1.

**Locations:** `_support/openspec/specs/versioned-installation/spec.md` (requisito «Verified release package installation»).

**Verification:** `rg -c "only the Pi host's pi-subagents step may download" <spec>` no devuelve nada; `rg -n "agent-browser capability consented" <spec>` devuelve el escenario; `git diff` del spec toca solo ese requisito.

## Revisión y progreso

- 2026-10-23: cambio de solo spec; el registro se archiva en el mismo PR que aplica el delta, porque no hay implementación aparte.
