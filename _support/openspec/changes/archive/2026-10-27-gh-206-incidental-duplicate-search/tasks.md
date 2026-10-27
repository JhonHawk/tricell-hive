# Tareas

Commit base: `20646413` (`development`, 2026-10-27).

### T1 — La referencia nombra las dos búsquedas y se aplica al proponer

- [x] `incidental-findings.md` se lee antes de proponer un hallazgo durante el trabajo o al cerrar; su estrechamiento (línea 13) se aplica en los dos momentos; su viñeta *Duplicate check* remite a las búsquedas en tracker y memoria persistente de la regla global, la del tracker una por propuesta y la de memoria una por miembro.
  reopened by the code-review fix d5e4c565

**Closes:** AC1, AC2.

**Checkpoint:** C1.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-close/references/incidental-findings.md` («When to read», línea 13 y viñeta *Duplicate check*); `_support/docs/architecture/instruction-resources.md:18`.

**Execution:** hilo principal (edición pequeña y acoplada al resto del cambio).

**Test approach:** check — las comprobaciones `rg` de la verificación.

**Changes:** según el diseño, H1 y H2.

**Verification:**
- `rg -n 'Duplicate check' content/skills/flow-close/references/incidental-findings.md`: la línea nombra el tracker y la memoria persistente, y dice que la de memoria va por miembro.
- `rg -n -A2 '## When to read' content/skills/flow-close/references/incidental-findings.md` y `rg -n 'narrows' content/skills/flow-close/references/incidental-findings.md`: la condición y el estrechamiento incluyen proponer durante el trabajo.
- `go test ./tests/content/...` pasa.

### T2 — Un fallo esquivado es un hallazgo incidental

- [x] La regla global dice que un fallo de un comando o herramienta que no bloquea el trabajo y se esquiva es un hallazgo incidental que se reporta con el fallo y el rodeo, que solo se propone como defecto cuando está en el proyecto actual, y que el de otro proyecto se reporta para la sesión de ese proyecto.
  reopened by the code-review fix d0e6d1e9

**Closes:** AC3.

**Checkpoint:** C2.

**Depends on:** ninguna.

**Locations:** `content/guidance/global.md:34`.

**Execution:** hilo principal.

**Test approach:** check — `rg`.

**Changes:** según el diseño, H4; una frase junto a «Pre-existing and intermittent test failures…».

**Verification:** `sed -n 34p content/guidance/global.md | rg -c -i 'work around'` devuelve 1 y la frase dice «does not block», limita la propuesta de defecto al proyecto actual y manda el de otro proyecto a su sesión; `wc -c content/guidance/global.md` se anota para el PR.

### T3 — Sitio y changelog

- [x] La página Close describe la búsqueda de duplicados en tracker y memoria; la página Build dice que un hallazgo propuesto durante el build pasa por esa misma búsqueda y que un fallo esquivado se reporta como hallazgo; `CHANGELOG.md` registra el cambio.
  reopened by the code-review fix bdb9513d

**Closes:** AC4.

**Checkpoint:** C3.

**Depends on:** T1, T2.

**Locations:** `site/src/content/docs/flows/close.md:47` (paso 5); `site/src/content/docs/flows/build.md` (viñetas junto a la línea 39); `CHANGELOG.md` (sección sin publicar).

**Execution:** hilo principal.

**Test approach:** check — `rg`.

**Changes:** describir el comportamiento y enlazar la regla, sin copiarla.

**Verification:** `rg -n -i 'memory' site/src/content/docs/flows/close.md` muestra la búsqueda de duplicados en el paso 5; `rg -n -i 'work(s|ed)? around' site/src/content/docs/flows/build.md` devuelve la nueva viñeta; `rg -n '#206' CHANGELOG.md` devuelve la entrada.

## Verificación y revisión humana

- C1 — T1 — cambia una referencia distribuida que leen todos los hosts (contrato público de Hive).
- C2 — T2 — cambia la guía global distribuida (contrato público de Hive).
- C3 — T3 — describe lo que C1 y C2 dejaron verificado.

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad / condición pendiente |
| --- | --- | --- | --- |
| Suite completa | candidato final, local | `go test ./...` pasa | incluida en la implementación |
| Revisión dedicada | candidato local, antes del push | `/code-review` de Claude Code (D9-A); sin hallazgos P0–P1 abiertos | D9-A |
| In vivo local | no aplica | el cambio es texto de guía; los pilotos están pausados por AGENTS.md | — |
| Despliegue local | tras integrar en `development` | `hive update` desde este checkout | marcador `maintainer.local` (AGENTS.md) |

## Estado de revisión y progreso

**Revisión del plan** (revisión `90019e2cb7bf`, dominio Coherencia de guía más el chequeo de orden; un `hive-review-plan`, ~86 s, ~112 k tokens): 4 bloqueantes y 3 no bloqueantes. Todos se aplicaron tal como los propuso el revisor, así que no hubo segunda ronda:
- B1: la búsqueda en memoria se hace por miembro, no por grupo (diseño H1, T1).
- B2: la línea 13 de la referencia entra en T1.
- B3: H4 se limita a fallos que no bloquean el trabajo, manda los del host o de otro proyecto por `global.md:30` y hereda el alcance de la viñeta (se quitó «en cualquier tarea» de AC3).
- B4: checkpoints C1, C2 y C3, uno por tarea.
- N1: la base pasa a `20646413`; los criterios siguen siendo falsos ahí.
- N2: la comprobación de AC3 busca «work around» en la línea 34.
- N3: la página Build menciona la búsqueda de duplicados durante el build.

Decisión tomada sin preguntar: quitar «en cualquier tarea» de H4, porque ampliar el alcance de toda la viñeta sería una decisión aparte y el caso de Codex (`hive update` después de integrar) cae dentro del trabajo autorizado.

**Progreso (2026-10-27):**
- T1 `38c3456d`, T2 `d9819dcb`, T3 `6f351594`: los tres `[x]`, verificados por `hive-verify-task` (contexto limpio; modelo configurado, no observado; implementador: opus-5-5 de la sesión).
- AC3 se ajustó al texto implementado: el fallo esquivado solo se propone como defecto cuando está en el proyecto actual; el de otro proyecto va a la sesión de ese proyecto.
- Desviación: T3 empezó antes de que T1 y T2 estuvieran `[x]` (no comparten archivos); las verificaciones de T1 y T2 dieron todo `met`, así que T3 no necesitó reverificarse.
- Observaciones de redacción de los verificadores, sin escenario de fallo: la viñeta *Duplicate check* no menciona el caso sin tracker; «propose it as a defect» podría leerse como si excluyera arreglarlo en este trabajo; la página Build omite «silently» y no dice que solo se propone el fallo del proyecto actual; la entrada del changelog va en *Fixed* aunque H4 es un hueco.
- `go test ./...` pasa en el candidato `6f351594`.
- `/code-review` (medium) sobre `6f351594`: 10 hallazgos, ninguno P0/P1. Corregidos 8 en `d5e4c565` (T1), `d0e6d1e9` (T2), `bdb9513d` (T3); quedan como notas P3: un fallo reportado y no propuesto no se puede suprimir, y el paso 5 de `flow-close` conserva su puntero más estrecho. Las tareas se reabrieron y se reverificaron: AC1–AC4 `met`.
- Notas de redacción de la reverificación, sin escenario de fallo: «at once» frente a «in your next message» de la regla de integraciones; la sección Actions de la referencia sigue escrita para la pregunta de cierre; la página Close no dice que un hallazgo rechazado se vuelve a proponer con su conteo.
- Merge de `origin/development` (`20646413`→`d3285724`) en la rama y `go test ./...` verde en el candidato final.
- [PR #231](https://github.com/JhonHawk/tricell-hive/pull/231) fusionado en `development` (`7372c065`); #206 cerrado por el PR.
