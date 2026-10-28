# Tareas — gh-256-long-running-loops

Base: `e158477c` (development), confirmada al iniciar el build el 2026-10-28. Rama `feat/gh-256-long-running-loops` en `.claude/worktrees/gh-256-long-running-loops`.

## T1 — Contrato de loop

- [x] `hive-core/references/loops.md` define el contrato según la sección 1 de `design.md`, y `hive-core/SKILL.md` lista su condición de lectura. `instruction-resources.md` lo suma a las casas canónicas, y `agent-delivery.md` nombra las primitivas por host.

**Closes:** AC1.

reopened by the revision D22-A: quitar las menciones a jobs de `loops.md`.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:**
- `content/skills/hive-core/references/loops.md`, que es nuevo;
- `content/skills/hive-core/SKILL.md`;
- `content/guidance/global.md`, con un puntero «Before starting or resuming work that repeats cycles until a condition…»;
- `_support/docs/architecture/instruction-resources.md`;
- `_support/docs/architecture/agent-delivery.md`.

**Execution:** hilo principal. Es la casa de la que dependen T2 a T4.

**Test approach:** check: `go test ./tests/content/...` (enlaces y tamaño) y la lectura de los siete elementos.

**Changes:** según la sección 1 de `design.md`.

**Verification:**
- `test -f content/skills/hive-core/references/loops.md`;
- `rg -n "loops.md" content/skills/hive-core/SKILL.md content/guidance/global.md _support/docs/architecture/instruction-resources.md` encuentra una línea en cada uno;
- la lectura muestra los siete elementos y que ninguno nombra un comando de host;
- `go test ./tests/content/...` pasa.

## T2 — Ruta iterativa y sin progreso

- [x] La ruta iterativa apunta a `loops.md`: el estado queda en disco, rige la regla de sin progreso y se suma la excepción de D17-A. `diagnosis.md:10` suma su puntero.

**Closes:** AC4.

**Checkpoint:** C2.

**Depends on:** T1.

**Locations:**
- `content/skills/flow-research/SKILL.md:48-58`;
- `content/skills/flow-research/references/diagnosis.md:10`, reescribiendo la cláusula «fixed counts… do not establish a useful stopping condition».

**Execution:** hilo principal.

**Test approach:** check.

**Changes:** según la sección 2 de `design.md`.

**Verification:**
- `rg -n "a round ends without a measurable improvement" content/skills/flow-research/SKILL.md` no encuentra nada, y `rg -n "fixed counts" content/skills/flow-research/references/diagnosis.md` tampoco;
- `rg -n "loops.md" content/skills/flow-research/SKILL.md content/skills/flow-research/references/diagnosis.md` encuentra una línea en cada uno;
- la lectura de la parada por decisión del usuario muestra la excepción desatendida;
- `go test ./tests/content/...` pasa.

## T3 — Retomar entre sesiones, sin jobs (revisada por D22-A)

- [x] Se quita todo lo de jobs declarados: la sección y las fuentes en `unattended-delegation`, `effect-authorization.md:29`, `global.md:32`, el destino `jobs/` y las menciones a jobs en `delivery-decisions.md`, `review-decisions.md` y `flow-build/SKILL.md:24`. `global.md:31` y `:58` vuelven al texto de la base. `delivery-decisions.md` conserva que el estado del loop nunca aporta respuestas, y la activación desatendida conserva la medida, el tope y el End del loop.
reopened by the revision D22-A

**Closes:** AC2, AC3.

**Checkpoint:** C2.

**Depends on:** T1.

**Locations:**
- `content/skills/unattended-delegation/SKILL.md:3`, `:8`, `:10`, `:12`, `:16`, la sección Declared jobs (que se quita) y `:35`;
- `content/skills/hive-core/references/effect-authorization.md:29`;
- `content/guidance/global.md:31`, `:32` y `:58`;
- `content/skills/workspace-conventions/references/artifact-layout.md:19-21`;
- `content/skills/flow-plan/references/delivery-decisions.md`, la condición «work resumes in a new session»;
- `content/skills/flow-plan/references/review-decisions.md:9` y `content/skills/flow-build/SKILL.md:24`.

**Execution:** hilo principal. Comparte `global.md` con T1.

**Test approach:** check.

**Changes:** según las secciones 3, 4 y 5 de `design.md`.

**Verification:**
- `rg -n -i "declared (scheduled )?job|jobs/" content` no encuentra nada;
- `git diff e158477c..HEAD -- content/guidance/global.md` no muestra cambios en `:31` ni en `:58`;
- `rg -n "loop's state never supplies" content/skills/flow-plan/references/delivery-decisions.md` encuentra la oración en el párrafo de registro (`:42`), la única casa de retomar;
- la lectura de `unattended-delegation` muestra tres fuentes de activación, que un turno programado o en segundo plano no activa nada (`:8`), la medida y el tope en `:10` y el End del loop; `rg -n "four sources" content` no encuentra nada;
- la lectura de `loops.md` (State y Resume) muestra el estado en disco, cómo se retoma, y que `waiting` impide reanudar sin el usuario;
- `go test ./tests/content/...` pasa.

## T4 — Sitio y CHANGELOG

- [x] Las páginas del sitio describen los loops y el estado, y el `CHANGELOG.md` registra el cambio. Las menciones a jobs se quitan.
reopened by the revision D22-A

**Closes:** AC5.

**Checkpoint:** C3.

**Depends on:** T2, T3.

**Locations:**
- `site/src/content/docs/flows/research.md:52`;
- `site/src/content/docs/flows/build.md:50` y `:78`;
- `site/src/content/docs/skills.md:33`;
- `site/src/content/docs/index.mdx:100`;
- `content/skills/hive-core/SKILL.md:3`;
- `llms.txt:32`;
- `CHANGELOG.md`.

**Execution:** hilo principal.

**Test approach:** check.

**Verification:**
- `rg -n -i "loop state" site/src/content/docs` encuentra las descripciones nuevas, y `rg -n -i "declared job|job file" site/src/content/docs CHANGELOG.md llms.txt` no encuentra nada;
- `rg -n "#256" CHANGELOG.md` encuentra la entrada.

## Verificación y revisión

- C1: T1. Cierra su propio checkpoint porque de él dependen tres tareas posteriores (regla de riesgo 1).
- C2: T2 y T3. T3 cierra C2 porque define un contrato persistido, el formato del job, y amplía la autorización entre sesiones (regla de riesgo 2).
- C3: T4.

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite completa | candidato final, local | `go test ./...` y `go vet ./...` pasan | implementación |
| Revisión dedicada | candidato local, antes del push | `/code-review` de Claude Code | D6-A |
| `hive update` | tras el merge | deploy del commit integrado | regla del mantenedor |

No hay UI ni entorno desplegado. Ningún criterio se verifica fuera de esta máquina.

## Revisión del plan y progreso

- Revisión del plan: pendiente.
- Progreso: nada implementado.

### Ronda 1 de la revisión del plan

- Dos `hive-review-plan` revisaron la revisión `dc23d9d54657`:
  - consistencia de guía: unos 151 s y 128 000 tokens; 5 bloqueantes y 3 menores;
  - autorización y operación: unos 77 s y 112 000 tokens; 3 bloqueantes y 4 menores.
- Todos se aplicaron sin preguntar, como correcciones convencionales: procedencia del job, estado terminal, corridas encimadas, retomar desde un plan o un job, punteros a `delivery-decisions`, `global.md:58` y `:31`, commits condicionados, `diagnosis.md`, AC5, checkpoints, casa única de las ubicaciones y lista del sitio.
- AC2 quedó acotado: el estado del loop no autoriza nada.
- Decisión del usuario: D19-A, el job llega como máximo a integrar en la rama base.
- Re-revisión: pendiente, una ronda con los mismos revisores.
- Re-revisión de autorización sobre `edc16f0a3db7`: unos 31 s y 120 000 tokens.
  - Un bloqueante residual: un agente podría escribir un job y hacer merge con `automatic`. Se aplicó la corrección que propuso el revisor: el job se crea solo a pedido y su contenido se valida antes de integrarlo.
  - Tres menores, aplicados tal como los propuso el revisor, sin otra ronda: corrida interrumpida, reporte terminal publicado una sola vez con un máximo de corridas o un techo total obligatorio, y deploy nombrado.
- Re-revisión de consistencia sobre `edc16f0a3db7`: unos 40 s y 143 000 tokens.
  - Dos bloqueantes: R1, el choque de D19-A con `unattended-delegation:24`, y R2, la excepción por sesión en dos casas.
  - Tres menores: las comprobaciones que no distinguían la base, la lista de alcance y la limpieza del estado terminal con la persona presente.
  - Todos se aplicaron tal como los propuso el revisor, sin otra ronda.
- Plan listo. No queda ninguna decisión abierta.

### Revisión D22-A (durante el build)

- El usuario quitó los jobs declarados de #256 tras preguntar qué es un job y si hace falta.
- Reemplaza D16-A, D19-A y D21.
- Cambió AC2 (se retoma solo desde el plan o la sesión), AC3 (la guía ya no nombra el job) y AC5. T1, T3 y T4 vuelven a `[/]`.
- El diseño investigado de jobs queda en `watch/declared-jobs`.
- La revisión cambia criterios y decisiones, así que corresponde una re-revisión del plan con el revisor de consistencia.
- Re-revisión de la revisión D22-A (consistencia, unos 35 s): un bloqueante (G1, AC2 sin verificar la parte de `waiting`) y dos menores (G2, «four sources»; G3, retomar en dos lugares). Los tres se aplicaron tal como los propuso el revisor, sin otra ronda: lectura de `loops.md` en T3, `:35` en «three» con su comprobación, y la regla de retomar fundida en `delivery-decisions.md:42`.

## Progreso del build

- **Commits** en `feat/gh-256-long-running-loops`, de `15c88354` a `4d1bc488` más el reordenamiento del CHANGELOG, integrados en el PR #257 (`5f0d8dc5`).
- **`hive-verify-task`:** T1 a T4 dieron `met` en sus checkpoints, y otra vez después de la revisión D22-A.
- **Suite:** `go test ./...` y `go vet ./...` en verde sobre el candidato final.
- **`/code-review`:** revisión completa, refutación, una re-revisión y la revisión del diff posterior. Todo se corrigió, como detalla el cuerpo del PR #257.
- **`hive update`:** aplicó `5f0d8dc5` y se verificó que `loops.md` quedó instalado.
