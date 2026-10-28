# Tareas de larga duración: contrato de loop y estado entre sesiones

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el [PR #257](https://github.com/JhonHawk/tricell-hive/pull/257) (`5f0d8dc5`) |
| Tracker · GitHub Issues | • [#256 — Tareas de larga duración: contrato de loop, estado entre sesiones y jobs declarados](https://github.com/JhonHawk/tricell-hive/issues/256) |
| Git | automatic · PR a `development` y merge · registro en un PR de cierre pequeño que integra el agente · sin parada humana |
| Verificación | comprobaciones `rg` por criterio · `go test ./...` · `/code-review` de Claude Code |
| Siguiente paso | ninguno; registro archivado |

## Solicitud

> vamos a ajustar un poco la premisa, que no se considere solo con human in the loop: [Image #1] si no que podemos adaptar hive para tareas long runing, analizalo consideranod el contexto actual, para ver si solo son pequeños ajustes o gates o ya es un alcance mayor

La imagen es el post de Claude sobre «Turn-based loops»: «Claude judges when it's completed the tasks or needs additional context… Exits when Claude judges the task complete. Or the effort budget runs out.»

Respuestas en las tarjetas (2026-10-28; D16-A, D19-A y D21 quedaron sin efecto por D22-A):
- **Ruta:** D14-A (planear todo).
- **Diseño:** D15-A (una referencia nueva en `hive-core`), D16-A (el job es un archivo versionado que trae su propia entrega), D17-A (en la ruta desatendida manda la delegación), D18-A (dos ciclos sin mejora llevan a un cambio de enfoque y luego a parar) y D19-A (un job llega como máximo a integrar en la rama base).
- **Entrega reutilizada de esta sesión:** D5-A (automatic) y D6-A (`/code-review`).
- **Durante el build:** D22-A, quitar los jobs declarados de #256. El trabajo programado con autoridad queda en observación (`watch/declared-jobs`).

## Objetivo

Hive está pensado para pausas con una persona presente. Para el trabajo largo tiene piezas sueltas: la ruta iterativa, la delegación desatendida, las rondas de arreglo y las esperas. No tiene un contrato de loop. El cambio crea uno, que no depende del host: un objetivo con su medida, un tope, la verificación de cada ciclo, una regla de «sin progreso», el estado en disco y la escalada. Le da a una tarea larga un estado en disco con el que retomarla en otra sesión sin volver a preguntar lo ya respondido. Y define el «declared job» como un archivo versionado que trae sus respuestas y autorizaciones, para correr sin una persona usando la primitiva de cada host.

## Evidencia

**Repositorio** (inspeccionado por un hijo `hive-research`; las citas clave las comprobó el hilo principal, en `development` `e158477c`):
- la ruta iterativa en `content/skills/flow-research/SKILL.md:48-58`, con una sola ronda plana como parada en `:52` y la decisión del usuario como parada en `:56`;
- `content/skills/unattended-delegation/SKILL.md`: el job declarado sin definir en `:8`, la activación sin medida ni tope en `:10`, la prohibición de crear un ledger en `:12`, la preautorización que solo existe cuando la persona está presente en `:16`, y «decide yourself» en `:20`;
- `content/guidance/global.md:31-32` y `:83`;
- `content/skills/hive-core/references/effect-authorization.md:29`: las autorizaciones no pasan de una sesión a otra;
- `content/skills/hive-core/references/questions.md:19`: las respuestas de flujo duran lo que dura la sesión;
- `content/skills/flow-research/references/diagnosis.md:10`.

**Fuentes externas:**
- «Getting started with loops», https://claude.com/blog/getting-started-with-loops, 2026-06-30: cuatro tipos de loop, por turnos, por objetivo, por tiempo y proactivo;
- documentación de Claude Code 2.1.312: `/goal`, scheduled tasks, routines, hooks y workflows;
- Codex 0.165.0, con `/goal` documentado;
- Anthropic, «Effective harnesses for long-running agents», 2025-11-26: archivo de progreso, commits como puntos de control y un ciclo por vez;
- Pi 1.4.0, OpenCode 2.2.0 y Grok 1.0.72 sin una primitiva de loop verificada; Cursor ofrece el seguimiento por hook Stop.

**Antecedente:** el Hive anterior al rebuild tenía «standing jobs» (2026-09-13): una fila en el ledger y la política de proponer sin integrar.

## Alcance y aceptación

**Incluido:**
- `hive-core/references/loops.md`, que es nuevo, y su condición de lectura en `hive-core/SKILL.md`;
- la ruta iterativa de `flow-research`, `diagnosis.md:10`, `unattended-delegation/SKILL.md` y `effect-authorization.md:29`;
- `global.md`: un puntero, y quitar el job de `:32`;
- `delivery-decisions.md` y `review-decisions.md:9`: retomar solo desde un plan;
- `reporting.md`, solo si hace falta un puntero; `hive-core/SKILL.md:3`, `llms.txt:32` y `_support/docs/architecture/agent-delivery.md`;
- `artifact-layout.md`: el destino del estado;
- `_support/docs/architecture/instruction-resources.md`;
- las páginas del sitio que describen estos flujos, y el `CHANGELOG.md`.

**Excluido, con su razón:**
- un scheduler o listener propio de Hive: se usan los de cada host (D16-A);
- relajar los invariantes de seguridad: los efectos críticos o irreversibles, el camino de promoción y las autorizaciones sin responder siguen deteniendo el loop;
- las reglas de procesos colgados (`verification.md:53`) y las esperas (`delegation.md:67`): miden un proceso, no los ciclos de un loop.

**Restricciones:**
- `go test ./...` pasa, con cada referencia por debajo de 16 000 B;
- cada regla tiene una sola casa;
- se reporta el tamaño de `global.md`, que hoy mide 17 668 B;
- `questions.md` no cambia, porque le quedan 8 B libres.

**Búsqueda de specs:** busqué «loop», «unattended», «scheduled», «job», «hive-core» y «SHALL NOT» en `_support/openspec/specs/*/spec.md`. Leí «Layered global core with a shared core skill» (`context-delivery`), que nombra cuatro referencias compartidas sin cerrar la lista y exige que cada procedimiento tenga una sola casa con su condición de lectura. Ningún requisito describe loops ni jobs, así que el cambio no lleva deltas.

**Criterios de aceptación:**
- AC1. `content/skills/hive-core/references/loops.md` define el contrato de loop: un objetivo con su medida, un tope (ciclos, tiempo o tokens), la verificación de cada ciclo, la regla de sin progreso (D18-A), el estado en disco con sus campos y cómo se retoma, y la escalada. Nombra la capacidad que aporta cada host, con su respaldo, y no un comando. `hive-core/SKILL.md` lista su condición de lectura. *Falso en la base cuando* el archivo no existe.
- AC2. Un loop deja su estado en disco y se retoma desde él en otra sesión o tras una compactación. El estado no autoriza nada: las respuestas de entrega y revisión salen solo del plan o de la sesión, como hoy, y `delivery-decisions.md` lo dice. Un estado `waiting` impide reanudar los ciclos sin el usuario. *Falso en la base cuando* `content/skills/hive-core/references/loops.md` no existe y `rg -n "loop's state never supplies" content/skills/flow-plan/references/delivery-decisions.md` no encuentra nada.
- AC3. La guía ya no nombra un «declared job» (D22-A). Un turno programado o en segundo plano no activa la delegación desatendida, y `global.md:32` y `unattended-delegation/SKILL.md:3` y `:8` listan tres fuentes. *Falso en la base cuando* `rg -n "declared (scheduled )?job" content` encuentra `global.md:32` y `unattended-delegation/SKILL.md:3` y `:8`.
- AC4. En la ruta iterativa, una ronda plana ya no la para sola: aplica la regla de sin progreso (D18-A). Si la ruta corre desatendida, las decisiones del usuario siguen las de la delegación (D17-A). *Falso en la base cuando* `flow-research/SKILL.md:52` dice «a round ends without a measurable improvement» y `:56` para ante «a decision is genuinely the user's» sin excepción.
- AC5. El sitio describe el contrato de loop y el estado en disco sin copiar reglas. *Falso en la base cuando* `rg -n -i "loop state" site/src/content/docs` no encuentra nada.

## Entrega

- **Repositorio:** `tricell-hive`, base `development`, en la rama propuesta `feat/gh-256-long-running-loops` dentro de un worktree.
- **Modo:** automatic (D5-A), con `/code-review` (D6-A) sobre el candidato local; después push, PR, merge y `hive update`.
- **Cierre del registro:** un PR pequeño, sin revisión, que integra el agente.
- **Ticket:** #256 se cierra al integrar.
- **Implementación:** pendiente de la pregunta de build.
