# Presupuesto de tamaño de las skills (2026-10-23)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` el 2026-10-23 por el [PR #126](https://github.com/JhonHawk/tricell-hive/pull/126), merge `f44468af`; registro archivado el mismo día |
| Tracker · GitHub Issues | Sin issue: sale de decisiones del usuario en la revisión de referencias (D22, D24, D25) |
| Git | `automatic` (D9) · integrado por merge del PR #126; rama `feat/skill-size-budget` eliminada tras el merge |
| Verificación | `hive-verify-task` por tarea · `go vet ./...`, `go test ./...`, pruebas de skills y build del sitio sobre el candidato final · sin revisión dedicada (D10) |
| Siguiente paso | Ninguno; registro cerrado |

## Solicitud

> bien, casi terminamos.
> - Actualiza gentle-ai, superpowers y ponytail
> - mismo tipo de revisión que llevamos hasta el momento

Respuestas del usuario (2026-10-23, tarjetas nativas):

- D22 (skills que se cortan tras la compactación): «Eso y achicar flow-build», es decir, releer la skill completa tras una compactación, achicar `flow-build` y añadir a `harness-audit` un límite en bytes.
- D24 (límite): «16.000 bytes + prueba».
- D25 (qué sale): primero «Analicemoslo y contratemoslo a detalle, y hagamos una tabla para verlo por secciones»; tras la tabla por secciones, «B1–B15, P1 y P2».
- D26 (tras la revisión del plan, que mostró que partes de B6, B7, B11, B12 y B15 no tenían otro hogar): «Mover el TDD a una referencia».

## Objetivo

La documentación oficial de Claude Code (`code.claude.com/docs/en/skills`, «Skill content lifecycle», leída el 2026-10-23 con Claude Code 2.1.310) dice que, al compactar la conversación, vuelve a adjuntar cada skill invocada conservando solo sus primeros 5.000 tokens, con 25.000 tokens compartidos. `flow-build/SKILL.md` pesa 28.911 bytes (unos 7.200 tokens estimados) y su sección «Verify and close» empieza en el carácter 22.348, así que se pierde tras una compactación; la regla de `global.md:78` no lo evita porque solo pide releer la skill «cuando sus pasos no están en el resumen», y el principio adjuntado oculta lo que falta. El cambio hace que la skill activa se relea completa tras una compactación, mueve a referencias condicionadas las reglas situacionales de `flow-build` y `flow-plan` hasta dejar cada `SKILL.md` en 16.000 bytes o menos, y fija ese límite con una prueba y con la regla `HA-SK-02` de `harness-audit`.

## Alcance y aceptación

Incluido: `content/guidance/global.md:78`; `content/skills/flow-build/SKILL.md` y sus referencias nuevas `plan-execution.md` y `tools-and-dependencies.md`; `content/skills/flow-plan/SKILL.md` y sus referencias nuevas `interface-changes.md` y `plan-revision.md`; las referencias existentes que reciben reglas o punteros (`flow-build/references/verification.md`, `flow-research/references/delegation.md`, `flow-walkthrough/SKILL.md`, `flow-walkthrough/references/ui-gate.md`, `flow-plan/references/questions.md`, `plan-review.md`, `plan-format.md`, `delivery-decisions.md`); los enlaces a `#revise-a-plan-in-progress`; `content/skills/harness-audit/references/skills-and-agents.md` (HA-SK-02); `flow-close/SKILL.md:25`, `flow-walkthrough/SKILL.md:62`, `plan-format.md:40` y `_support/docs/architecture/instruction-resources.md:62`, cuyos punteros o afirmaciones cambian; la referencia nueva `flow-build/references/tdd.md` (D26); una prueba nueva en `tests/content/`; la lista de lecturas de `tests/content/reading_footprint_test.go`; `AGENTS.md` («Guidance structure»); `CHANGELOG.md`.

Excluido: cambiar qué hacen las reglas (solo cambia dónde viven y cuándo se leen); dividir `flow-build` en dos skills; reglas en observación como G1 (datos reconstruibles); páginas del sitio, que describen comportamiento y no secciones.

Restricciones: ninguna regla se pierde ni queda con dos hogares (`AGENTS.md`, «Guidance structure»); cada puntero nombra la situación exacta en que se lee; `go vet ./...`, `go test ./...`, las pruebas de skills y el build del sitio siguen pasando; la validación de enlaces de instrucciones pasa.

**No especificado todavía:** el número exacto de tokens de cada skill; no hay contador disponible (sin clave de API ni librería). Se especificaría con un contador de tokens de Claude; mientras, el límite de 16.000 bytes deja margen para entre 3,2 y 4 caracteres por token.

- AC1. La regla de compactación de `global.md` pide releer completo el `SKILL.md` de la skill activa tras una compactación, porque un host puede volver a adjuntar solo su principio. *Falso en la base cuando* `rg -c -i 'only (its|the) (start|beginning)' content/guidance/global.md` no encuentra nada.
- AC2. Cada `content/skills/*/SKILL.md` pesa 16.000 bytes o menos. *Falso en la base cuando* `wc -c content/skills/*/SKILL.md` muestra `flow-build` con 28.911 y `flow-plan` con 17.326.
- AC3. Una prueba de `tests/content/` falla si algún `content/skills/*/SKILL.md` supera 16.000 bytes y nombra el archivo. *Falso en la base cuando* `rg -n '16000|16_000|16,000' tests/content/` no encuentra nada.
- AC4. `HA-SK-02` fija el límite de 16.000 bytes por `SKILL.md` junto al de líneas, con la razón: un host puede conservar solo el principio de una skill tras una compactación. *Falso en la base cuando* `rg -c '16,000' content/skills/harness-audit/references/skills-and-agents.md` no encuentra nada.
- AC5. `flow-build/references/plan-execution.md` contiene el registro del plan, los marcadores y los seis pasos de `hive-verify-task`, la verificación en paralelo, los commits por tarea, el cierre de las otras tareas y el brief de un hijo para una tarea de plan (B1–B6), y `flow-build/SKILL.md` la enlaza con la condición «while you execute or resume a retained plan, before you start, mark, verify, or commit any of its tasks». *Falso en la base cuando* el archivo no existe.
- AC6. `flow-plan/references/interface-changes.md` contiene la regla de cambios de interfaz de `flow-plan` (P1) y la revisión de transiciones de `flow-build` (B10), y ambos `SKILL.md` la enlazan con la condición de un cambio de interfaz. *Falso en la base cuando* el archivo no existe.
- AC7. `flow-plan/references/plan-revision.md` contiene «Revise a plan in progress» (P2), y ningún archivo de `content/` enlaza ya a esa sección por su ancla. *Falso en la base cuando* el archivo no existe y `rg -c '#revise-a-plan-in-progress' content/` encuentra 7 enlaces en 6 archivos (seis desde otros archivos y uno interno en `flow-plan/SKILL.md:18`).
- AC8. `flow-build/references/tools-and-dependencies.md` contiene las reglas de instalar herramientas y añadir dependencias (B13), y `flow-build/SKILL.md` la enlaza con su condición. *Falso en la base cuando* el archivo no existe.
- AC9. Ninguna regla movida (B1–B14, P1, P2 y el TDD) se pierde ni queda con dos hogares: para cada pieza que se traslada, la frase distintiva que lista `tasks.md` aparece en su destino y ya no en el `SKILL.md` de origen; para cada pieza cuyo destino ya decía la regla (B7 en su parte de UI, B8, B9 y B11 en su parte ya cubierta), el origen conserva una sola frase de puntero; las partes que no tienen otro hogar (B15, el resto de B7, B11 y B12, y el brief sin plan de B6) siguen en `flow-build/SKILL.md`. *Falso en la base cuando* las frases distintivas aparecen hoy en `flow-build/SKILL.md` o `flow-plan/SKILL.md`.
- AC12. `flow-build/references/tdd.md` contiene las reglas de TDD de `flow-build/SKILL.md:44-46`, y tanto `flow-build/SKILL.md` como `flow-plan/SKILL.md:36` («the build skill's TDD criterion») la enlazan, con la condición «before you change code or tests» en `flow-build` y «when you choose a task's test approach» en `flow-plan`; también la enlazan `flow-walkthrough/SKILL.md:60` y `flow-build/references/verification.md:3`. *Falso en la base cuando* el archivo no existe.
- AC10. `AGENTS.md` («Guidance structure») dice que cada `SKILL.md` se queda en 16.000 bytes o menos y que la prueba lo comprueba. *Falso en la base cuando* `rg -c '16,000' AGENTS.md` no encuentra nada.
- AC11. `CHANGELOG.md` tiene en `[Unreleased]` la entrada de este cambio. *Falso en la base cuando* `awk '/^## \[Unreleased\]/{f=1;next} /^## \[/{f=0} f' CHANGELOG.md | rg -c -i 'compact'` no encuentra nada.

## Entrega

- Modo `automatic` (D9): rama `feat/skill-size-budget` desde `development`, un commit por tarea verificada, push, PR a `development` con los lugares revisados y el cambio de tamaño de `global.md`, merge cuando pasen las pruebas locales y `hive update`.
- Revisión: sin revisión dedicada (D10).
- Cierre del registro: un PR pequeño de cierre sin revisión dedicada, que fusiono yo.
- No incluye release ni promoción a `master`.
