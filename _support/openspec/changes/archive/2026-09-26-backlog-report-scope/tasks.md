# Tareas

### T1 — Regla del backlog y forma de tickets en `global.md`

- [x] `global.md` tiene los cambios de [design.md](design.md#guía-globalmd), incluidas las correcciones P1–P3 de la auditoría; el presupuesto sube a 41177 bytes (+542).

**Ejecución:** hilo principal.
**Verificación:** `go test -count=1 ./tests/content/...` pasa.

### T2 — Criterio `ticket_ids_not_packed_in_prose`

- [x] Criterio en `tests/pilot/regression.go` según [design.md](design.md#criterio-determinista-ticket_ids_not_packed_in_prose), registrado en `regressionCriteria` (seis criterios).
- [x] 8 fixtures Grok en `tests/fixtures/regression/ticket_ids_not_packed_in_prose/`, con `provenance.md` y fila en el README.

**Ejecución:** delegado a `test-engineer` (dos rondas), en las rutas `tests/pilot/regression*.go`, `tests/pilot/flows.go` y `tests/fixtures/regression/**`; sin commits.
**Verificación:**
- TDD rojo→verde: el test no compilaba sin la función.
- Reversión por rama: la exclusión de listas, el umbral, los encabezados y las negritas hacen fallar cada uno su fixture.
- `go test -race ./tests/pilot/...` pasa y el escaneo de fixtures sale vacío.
- Sobre las 4 salidas de Haiku: la primera versión daba falso positivo en las etiquetas de grupo; tras la corrección, 2 pasan y 2 quedan en `not_observed`.

### T3 — Verificación conjunta y entrega

- [x] `go vet ./...` y `go test -race ./...` en verde en un worktree limpio de `556c52b`; push a `rebuild/harness-engineering`; release `0db2adc637b6` desplegada en los seis hosts. El texto nuevo se verificó en `~/.claude/CLAUDE.md` (Claude y Grok), `~/.codex/AGENTS.md`, `~/.cursor/AGENTS.md`, `~/.config/opencode/AGENTS.md` y `~/.pi/agent/AGENTS.md`.

**Ejecución:** hilo principal. El árbol de trabajo no compila `tooling/cli` por trabajo ajeno sin commitear (`provider_adapter_test.go:257`), así que la verificación completa corre en el worktree del commit.
