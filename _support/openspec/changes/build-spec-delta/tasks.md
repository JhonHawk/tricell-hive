# Tasks

Base: `development` en `d3d2b46c`, rama `fix/build-spec-delta`. Cambio sin plan; esta carpeta sigue `flow-build/references/spec-delta.md`.

### T1 — Puntero y referencia en `flow-build`

- [x] `flow-build` enlaza `references/spec-delta.md` con su condición y su paso 2 dice «When the work has a change folder»; la referencia y `instruction-resources.md` la describen.

**Closes:** AC1, AC2, AC4.

**Verification:** lectura de `flow-build` y `references/spec-delta.md`; `go test ./tests/content -run TestRequiredProcedureReadingFootprint`.

### T2 — Playbook en `hive-mode` y su delta

- [x] `hive-mode` registra el playbook en el `proposal.md` de cualquier carpeta de cambio; el delta de «Entry router with playbooks» lo dice.

**Closes:** AC3.

**Verification:** lectura de `hive-mode` «Record» y del delta; `go test ./tests/content`.

## Progreso

2026-11-20 — T1 · 7c0d3899 · unverified: none
2026-11-20 — T2 · 7c0d3899 · unverified: none
