# Tareas · spec-agent-browser-downloads

Commit base: `165d457a` (`development`).

### T1 — Delta de agent-browser fusionado en el spec

- [x] Delta MODIFIED de tres requisitos escrito y aplicado al spec vigente.

**Closes:** AC1, AC2, AC3.

**Locations:** `_support/openspec/specs/versioned-installation/spec.md` (requisitos «Optional operation capabilities», «Verified release package installation» y «Read-only diagnostics»).

**Verification:** `rg -c "only the Pi host's pi-subagents step may download|no capability step downloads|0.74.0" <spec>` no devuelve nada; `rg -c` devuelve 1 para cada escenario («agent-browser on a supported platform», «agent-browser step fails», «Unsupported platform or missing prerequisite», «Safe modes run nothing», «agent-browser present without a browser», «agent-browser capability consented», «Planned CodeGraph step rejected», «CodeGraph detected without running it») y para «only on macOS arm64 with Homebrew»; `git diff 165d457a` del spec toca solo esos tres requisitos.

## Revisión y progreso

- 2026-10-23: cambio de solo spec; el registro se archiva en el mismo PR que aplica el delta, porque no hay implementación aparte.
- 2026-10-23: validado (D15-A); PR #116. `/code-review` (88 s, 97 695 tokens): 9 hallazgos. Su causa raíz propuesta (deltas que sobrescriben) no se sostiene: `git log` del spec muestra que el delta de agent-browser de 2026-10-15 nunca se aplicó. D16-A amplía el cambio a fusionar ese delta; eso resuelve los hallazgos 1–3 (cláusula de consentimiento, binario de GitHub), 7 (escenario de Pi) y 8 (puerta por plataforma). 4 y 5 (tabla del registro) quedan resueltos con esta actualización y la del cierre; 9 (cotejar con SECURITY.md) queda cubierto por usar la redacción aprobada el 2026-10-15, que coincide con `SECURITY.md:31-36`. 6 pasa a la incidencia de D17.
- 2026-10-23: delta fusionado; todas las comprobaciones de T1 pasan; el diff del spec toca solo los tres requisitos.
