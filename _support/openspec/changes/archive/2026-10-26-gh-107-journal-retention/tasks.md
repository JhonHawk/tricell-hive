# Tareas: retención de journals

Commit base: `269f1a92` (`development`). Worktree `.claude/worktrees/gh-107-journal-retention`, rama `fix/gh-107-journal-retention`; el registro del cambio vive en este worktree.

## Tasks

### T1 — Podar journals terminados al final de cada aplicación

- [x] `pruneJournals` existe y se llama al final de `Apply` no anidado y de `Onboard` con pasos, según [el diseño](design.md#proposed-design).

**Closes:** AC1, AC2.

**Checkpoint:** C1.

**Depends on:** ninguna.

**Locations:** `tooling/management/apply.go` (`Apply`, `finishApply`), `tooling/management/onboarding.go` (`Onboard`, `finishOnboarding`), un test nuevo `tooling/management/journal_retention_test.go`.

**Execution:** hilo principal. Es una función pequeña acoplada al lock y a los dos caminos de aplicación, y el brief para un hijo sería casi tan largo como el cambio.

**Test approach:** tdd.

**Changes:** cada ciclo es rojo y luego verde, de uno en uno. Los ciclos 1, 5 y 6 pasan por `Apply` u `Onboard`. Los ciclos 2, 3, 4 y 7 llaman directamente a `pruneJournals` sobre un directorio sembrado, porque `preflight` rechaza un `pending.json` o un `onboarding-pending.json` (`apply.go:110-117`, `onboarding.go:115-123`):
1. Doce aplicaciones con cambios dejan 10 journals, y el último está entre ellos.
2. Un journal antiguo nombrado por `pending.json` se conserva. Se fabrica con el patrón de `shared_test.go:467`.
3. Un journal en fase `prepared` y otro cuyas tres primeras claves no son `Version`, `ID` y `Phase` se conservan.
4. El `CoreID` de un `onboarding-pending.json` presente se conserva.
5. Doce `Onboard` con pasos y un runner doble dejan 10 journals, también cuando el resultado es `partial`.
6. Con `transactions/` sin permiso de escritura (`chmod 0500` antes de la duodécima aplicación), `Apply` devuelve éxito y su resultado lleva el aviso de poda.
7. Un `pending.json` ilegible no borra nada y devuelve un aviso.

**Verification:** `go test ./tooling/management/ -run 'JournalRetention|Prune' -count=1` pasa, y antes de implementar cada ciclo falla por la causa esperada. Después, `go test ./tooling/management/... -count=1` pasa sin cambios en los tests existentes.

### T2 — Nombrar la carpeta como journals de recuperación y documentar la retención

- [x] Salida, documentación y delta de especificación describen la retención.

**Closes:** AC3, AC4.

**Checkpoint:** C2.

**Depends on:** T1.

**Locations:** `tooling/cli/install.go:825`, `tooling/cli/tui_hosts.go:51`, `tooling/cli/tui_wrap_test.go:175`, `README.md:146`, `_support/docs/architecture/deployment-manager.md:451` y `:456`, `specs/versioned-installation/spec.md` de este cambio.

**Execution:** hilo principal, por ser texto y documentación que dependen del comportamiento de T1.

**Test approach:** check: el test existente de la salida del instalador y `rg`.

**Changes:** aplicar el texto de [el diseño](design.md#proposed-design). Revisar `llms.txt` y `site/` en busca de "backup" referido a `transactions/`; hoy no hay coincidencias, se vuelve a comprobar. El delta ya está escrito y se ajusta si T1 cambia algún detalle.

**Verification:**
- `rg -n 'Private backups' tooling README.md _support/docs` no devuelve nada.
- `rg -n 'Recovery journals \(newest 10 kept\)' tooling/cli/install.go tooling/cli/tui_hosts.go` devuelve las dos líneas, o un test de la salida del instalador lo afirma.
- `go test ./tooling/cli/ -run 'Install|Hosts|Wrap' -count=1` pasa.
- `deployment-manager.md` ya no contiene "does not prune them" referido a journals. `rg -n '10 most recent' _support/docs/architecture/deployment-manager.md` devuelve la regla de retención, que dice que nunca se borra un journal que una recuperación necesite. El texto conserva que el historial de releases se retiene.

## Verification and human review

- C1 — T1 — cierra un grupo porque borra datos guardados en el directorio de estado (criterio 2).
- C2 — T2 — última tarea `check`.

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite completa | Candidato final, local | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'`, todo en verde. Sin `-race`: no hay concurrencia nueva, la poda corre bajo el lock existente | Implementación |
| Revisión dedicada | Candidato local, tras la suite y antes del push | Claude Code `/code-review`; sus hallazgos se corrigen o se refutan con evidencia | D4-A (2026-10-26) |
| In-vivo local, antes del merge | Candidato final, binario en el scratchpad | `go build -o <scratch>/hive ./tooling/cli`, luego `hive install --home <scratch>/home --state-dir <scratch>/state --hosts codex`. Se siembran 250 journals `committed` sintéticos de 12 MB en `<scratch>/state/transactions/` y se fuerza un cambio con `hive plan remove`/`install` medido con `time`. Se espera ≤ 10 journals terminados, una operación que tarda lo mismo (±1 s) que con la carpeta vacía y ninguna escritura fuera del scratchpad | Implementación (local, datos sintéticos) |
| In-vivo local, tras el merge | Refresco de AGENTS.md | `go build -o "$(command -v hive)" ./tooling/cli` y `hive update`. Antes y después: `ls ~/Library/Application\ Support/tricell-hive/transactions \| wc -l` y `du -sh` de esa carpeta. Se espera ≤ 10 journals terminados y unos 120 MB, y `hive doctor` sin operación pendiente | Refresco del mantenedor (AGENTS.md) más D1-A |

Antes del merge, el in-vivo usa un home y un directorio de estado sintéticos (`--home` ignora las rutas del usuario, `install.go:185`), nunca el estado real: `hive update` sobre una copia del estado escribiría en los hosts reales. El refresco tras el merge borra unos 2,8 GB sin vuelta atrás; por eso el in-vivo sintético corre primero.

## Review status and progress

- **Revisión del plan:** revisión 773a9e46111b, un `hive-review-plan` del dominio gestor y datos locales, que también hizo la comprobación de toda la lista de tareas. Unos 146 s y 115 k tokens. Cuatro bloqueantes y cuatro no bloqueantes, todos aceptados tal como los propuso el revisor, así que no hizo falta otra ronda:
  - Aviso de `Onboard`: va en `OnboardingResult.ID`, y se poda también en `partial`.
  - Ciclos de test que `preflight` hace inalcanzables: llaman a `pruneJournals` directamente. El éxito con aviso se prueba con un `transactions/` no escribible.
  - Verificación positiva de AC3 (texto del instalador) y AC4 (documentación de la retención).
  - Lectura acotada a las tres primeras claves.
  - No bloqueantes: proteger el ID propio, desempatar por nombre, no podar si `finishApply` falla, e in-vivo sintético antes del merge.
- **Sin cobertura:** reflexión en los llamadores (no hay índice CodeGraph; la búsqueda fue con `rg`).
- **Decisiones tomadas sin preguntar:**
  - El aviso en el ID del resultado: es el patrón existente de `recordSourceCommit` y no toca structs persistidos.
  - El fallo de poda se prueba con `chmod` y no con un failpoint nuevo: así no se añade una entrada solo para tests.
- **Verificación del candidato `d702b9ee`:**
  - T1 y T2 verificadas por `hive-verify-task`: AC1–AC4 cumplidos.
  - Suite completa en verde: 20 paquetes Go y 64 tests de skills.
  - In-vivo sintético: 251 journals (2,9 GB lógicos) pasan a 10 (108 MB); la operación tarda 0,85 s, frente a 0,83 s con la carpeta vacía. El estado real no se tocó (mismo `mtime`).
  - `/code-review` (esfuerzo medio, verifica sus propios hallazgos) sobre `d702b9ee`: 9 hallazgos.
    - Corregidos en `a7acfbd7` y `5092bd01`: el 10 escrito a mano en la CLI ahora sale de `management.RetainedJournals`; el aviso de poda nombra la primera causa; la frase "backups use 0600" de `deployment-manager.md` ya no es ambigua.
    - Descartado, "borra la única copia de bytes del usuario": Hive solo reemplaza su bloque gestionado (`deployment-manager.md:436`), el retiro legacy se niega a tocar archivos modificados y el `config.toml` de Codex va por líneas, con `config.toml.hive-backup`. Además, la poda es la decisión D1-A del usuario.
    - Solo reportados, de severidad menor o decisión de diseño: orden por `mtime` (un journal recuperado o un directorio copiado sin `-t` puede desordenar los finalizados, que nadie lee); aviso añadido al ID de `Onboard` (mismo patrón que `Apply` y `recordSourceCommit`; solo se muestra); protecciones inalcanzables desde los llamadores actuales (defensa deliberada del diseño); validación de nombre duplicada con `onboardingRecordPattern`; journals en fase no terminada que nunca se podan (restricción del plan).
    - Sin re-revisión: ninguna corrección era P0 ni P1 de lógica. Tras corregir, `go vet ./tooling/...` y los tests de `tooling/management` y `tooling/cli` siguen en verde.
  - Implementador: Claude Opus 5.5 (sesión). Verificadores: perfil configurado, no observado.
- **Integración:** [PR #222](https://github.com/JhonHawk/tricell-hive/pull/222), merge `b9cb57cb` en `development`. Después se recompiló `~/.local/bin/hive` desde ese árbol.
- **In-vivo tras el merge: no corrió.** Este merge no trae cambios de `content/`, así que `hive update` saldría `unchanged` y no podaría. El estado real sigue con 263 journals hasta la próxima operación que cambie algo. Lo cubre el in-vivo sintético.
- **Progreso de T1:** hecha en `a57a2057`.
  - Rojo observado por la causa esperada en los ciclos 1 a 5.
  - Los ciclos 6 y 7 pasaron a la primera porque su código ya existía de los ciclos 1 y 2. Una mutación que desactiva cada uno hace fallar su test.
  - `go vet` y `go test ./tooling/management/...` en verde (25 s).
- **Desviación del plan (ciclo 6):** con `transactions/` en 0500 antes de la aplicación, falla el propio journal de la operación (`startTransaction`), no la poda. El test usa `Onboard`: el runner quita el permiso después de que el core confirme y antes de la poda. Comprueba que el resultado es `completed` con el aviso. El cableado del aviso en `Apply` sigue el patrón de `recordSourceCommit` y no tiene una prueba de extremo a extremo propia.
- Siguiente paso: `flow-build` con T1 (poda de journals).
