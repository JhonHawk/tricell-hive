# plan-runs-verification-commands — el plan ejecuta los comandos de verificación que escribe

| Campo | Valor |
| --- | --- |
| Estado | En build |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `04a1f27b` · rama propuesta `feat/plan-runs-verification-commands` |
| Playbook | ninguno |

## Decisiones

2026-11-21
- **Reply A** — ¿Se adopta ya o se espera una segunda observación? Se adopta como vacío: ninguna regla pide ejecutar los comandos de verificación de un plan. Basta la decisión del usuario, sin recurrencia (`engram:tricell-hive/decision/plan-runs-its-verification-commands`).
- **D1 A** — ¿Qué enfoque? Extender lo existente: `flow-plan` paso 3 es el hogar de la regla; «Revise a plan in progress», el rol `hive-review-plan`, la plantilla `proposal.md` y `plan-handoff` solo la apuntan. Sin chequeo por script y sin fusionar la hoja de datos externos.
- **D2 A** — ¿Se suman a este plan los revisores del plan en paralelo? No: van en un cambio aparte, después de este. Un revisor por superficie solo cuando el plan cruza dos o más superficies de código independientes, el ángulo de UX en formularios de captura, y la re-revisión retomando al mismo revisor.
- **D3 A** — ¿Hasta cuánto puede tardar un comando que el plan ejecuta completo? 5 minutos según su última corrida medida, local o en CI; uno más largo se ejecuta solo en su forma más corta con el mismo runner, por ejemplo un archivo de prueba.
- **Corrección del usuario (2026-11-21, durante el build)** — «se ejecuten los comandos en paralelo». Los comandos que no comparten base de datos, puerto, lock ni carpeta de salida arrancan juntos y se esperan juntos; los demás, uno tras otro. Cambia la regla y el delta, no el resultado: entra como revisión del plan.
- **Supuesto** — Una prueba que escribe en una base de datos local de pruebas que ya está corriendo sí se ejecuta durante el plan; solo queda fuera la base que no es local, como la regla global de efectos críticos. Alternativa descartada: excluir toda base de datos, lo que habría dejado sin ejecutar la prueba Jest contra Postgres del caso FSR-937. Para revertirlo, se quita «que no es local» de la regla.
- **Supuesto** — La sesión de revisión agrega una fila a `session-review-rubric.md` para observar la regla en sesiones reales. Alternativa descartada: no agregarla. Para revertirlo, se borra la fila.
- **Supuesto** — `rule-enforcers.md` agrega la regla como `judgment`, porque ningún archivo muestra si un comando se ejecutó. Alternativa descartada: no listarla. Para revertirlo, se borra la fila.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Pedido

> «no crees que nos falta incentivar mayor investigación o metodologia de esta en la creación del plan?» — y luego «A» al cierre que proponía planear el cambio de `flow-plan` (pasos 1 y 5) con la comparación completa.

Sin ticket.

## Objetivo

Hoy un plan escribe en el campo `Verification:` de cada tarea un comando que nadie ejecuta antes de la revisión. Cuando el comando no puede correr (el runner es otro, falta una dependencia, el módulo no carga), lo descubre el revisor del plan, o el build. Y cada corrección tras la revisión puede traer otro comando sin probar. En la sesión de Grok `01a1fb69` (fleetsystem, FSR-937/938, 2026-11-20) eso costó dos rondas extra de `hive-review-plan`, de 7 a 10 minutos cada una.

Al terminar, el plan ejecuta sobre la base cada comando de verificación que escribe, o su forma existente más cercana, y lo registra en su contexto verificado. Un comando que no puede correr se corrige antes de la revisión. Una revisión del plan que cambia un comando lo ejecuta otra vez antes de la ronda siguiente. El revisor del plan señala el comando que no tiene su línea.

## Alcance y aceptación

Dentro: `content/skills/flow-plan/SKILL.md`, `content/skills/flow-plan/references/plan-handoff.md`, `content/skills/flow-plan/assets/change-folder/proposal.md`, `content/skills/hive-rule-change-folder/SKILL.md`, `content/agents/review/hive-review-plan.md`, `tests/content/single_home_test.go`, `site/src/content/docs/flows/plan.md`, `_support/docs/harness-engineering/session-review-rubric.md`, `_support/docs/harness-engineering/rule-enforcers.md`, `CHANGELOG.md`, y el delta de `workflow-lifecycle`.

Fuera: los revisores del plan en paralelo (D2 A, cambio aparte), un chequeo en `check_change.py` (D1 B, descartado), fusionar `hive-rule-external-facts` (D1 C, descartado), cualquier cambio en `global.md`, y un detector en `tests/pilot`.

- AC1 — En `flow-plan` instalado, el paso 3 pide ejecutar sobre la base cada comando que nombra un `Verification:` antes de escribirlo, o su forma existente más cercana más una prueba desechable, fuera del repositorio cuando el runner acepta una ruta, que solo importe los módulos que importará la prueba nueva; no ejecutar lo que cambia archivos versionados, una base de datos que no es local o recursos compartidos; ejecutar solo en su forma más corta, con el mismo runner, un comando cuya última corrida medida pasó de 5 minutos; arrancar juntos, en segundo plano, los comandos que no comparten base de datos, puerto, lock ni carpeta de salida, y los demás uno tras otro; registrar una línea por comando en el contexto verificado o en el resumen, o `not measured` con la tarea que lo encuentra primero; y corregir antes de la revisión un comando cuyo script, runner o dependencia falta, o cuyo módulo no carga · falso hoy porque `rg -n 'Verification' ~/.agents/skills/flow-plan/SKILL.md` no devuelve ningún paso así · Observed in: `~/.agents/skills/flow-plan/SKILL.md` tras `hive update`
- AC2 — En «Revise a plan in progress» instalado, una revisión que cambia el comando de verificación de una tarea, o aquello en lo que se apoya, lo ejecuta otra vez y actualiza su línea antes de la ronda de `hive-review-plan` · falso hoy porque el paso 3 de esa sección solo corre `check_change.py` · Observed in: `~/.agents/skills/hive-rule-change-folder/SKILL.md` tras `hive update`
- AC3 — El rol `hive-review-plan` instalado reporta como hallazgo el comando de verificación de una tarea que no tiene su línea en el contexto verificado, o cuya línea lo da por ejecutado aunque la base no tenga lo que necesita, sin ejecutarlo · falso hoy porque el rol solo revisa las líneas de datos externos · Observed in: `~/.claude/agents/hive-review-plan.md` y `~/.grok/agents/hive-review-plan.md` tras `hive update`
- AC4 — `go test ./tests/content/...` falla cuando la frase canónica de la regla aparece en otro archivo de `content/` además de `flow-plan/SKILL.md` · falso hoy porque `single_home_test.go` no tiene esa frase · Observed in: la salida de `go test ./tests/content/...` con una copia temporal de la frase en `content/skills/hive-rule-change-folder/SKILL.md`, borrada después

## Requisitos previos

None: the build needs nothing from outside the repository.

## Contexto verificado

- Sesión Grok `01a1fb69-6507-79e1-874d-14f8c3a0976f`, `grok-4.7-build-fast` alto (en el piso de modelos, `model-floor.md` línea 36). Rondas de `hive-review-plan` a las 18:32, 18:46 y 18:54 (hora local). La edición de las 18:45–18:46 prescribió Vitest; la ronda 2 encontró que Fleet Manager corre `node --test`. La edición de las 18:54 cambió a `node --test`; la ronda 3 encontró que Node no carga el módulo. Ninguna edición ejecutó el comando. Fuente: `updates.jsonl` líneas 1310–1499 y `subagents/*/output.json`.
- El caso se ve con lectura: `git -C fleetsystem-frontend-monorepo show 1024669:apps/fleet-manager/package.json` da `"test": "node --test test/*.test.mjs"` y no lista Vitest (2026-11-21).
- `hive-rule-change-folder` «Revise» paso 4 pide una ronda de revisión por revisión que cambie un criterio o el delta: las tres rondas de esa sesión cumplían la regla.
- La regla vecina es `hive-rule-external-facts`: mide datos externos y los escribe como «dato · comando · resultado o not measured · fecha» en el contexto verificado. El revisor revisa esas líneas sin ejecutar. Su requisito es `workflow-lifecycle` «A plan measures the external facts it depends on».
- `flow-plan/SKILL.md` pesa 7860 bytes, con tope de 16 000. `global.md` pesa 28 593, con tope de 30 000, y no se toca.
- El cambio pendiente `reporting-family-leaves` (en espera, sin integrar) también edita `plan-handoff.md` (sus líneas 21 y 28; este cambio, el punto 7 en la línea 16), `single_home_test.go` (los dos agregan entradas a `homePhrases`), `session-review-rubric.md` y el spec `workflow-lifecycle` (requisitos ADDED con otros nombres). Los conflictos serían de merge, sin contradicción: el que entre segundo se rebasa sobre el primero.
- Comandos de verificación de este plan, ejecutados sobre la base `04a1f27b` el 2026-11-21:
  - `go test ./tests/content/...` · pasa en 1,6 s
  - `go vet ./... && go test ./...` · pasa en 4 min 14 s
  - `pnpm --dir site check && pnpm --dir site lint && pnpm --dir site build` · pasa en 4 s (0 errores, 20 páginas)
  - `python3 -I ~/.agents/skills/flow-plan/scripts/check_change.py _support/openspec/changes/plan-runs-verification-commands` · sale 0

## Effects

None: the tasks take no effect on an external system.

## Entrega

`automatic` (respuesta de esta sesión, 2026-11-21): revisión, push, pull request a `development` y merge sin detenerse; después, la actualización local de la instalación, porque este checkout tiene la marca de mantenedor. Revisión de código: nativa, Claude Code `/code-review`. Sin revisión de seguridad: el cambio no toca ningún límite de seguridad. Esto no autoriza un release ni la promoción a `master`.

## Estado de la revisión y avance

2026-11-21 — Plan escrito. Revisión de `hive-review-plan` (2 min 39 s, 123k tokens): 1 P1, 7 P2 y 5 P3. Aplicados en el texto de la regla, el delta, T1, T2 y este contexto: el límite pasa a ser por efecto y no por ubicación (P1); la prueba desechable va fuera del repositorio cuando se puede; un servicio que el build levanta queda `not measured`; las observaciones y los recorridos no son comandos; el resumen de un plan en la conversación lleva las líneas; la plantilla usa un comentario aparte; se agregan `CHANGELOG.md` y las secciones y checks del sitio; el nombre de la prueba es `TestRulesHaveOneHome`; se separan los pasos 4 y 5 de `flow-plan`; el puntero de «Revise» aplica en la siguiente ronda, cuando la hay; se amplía el cruce con `reporting-family-leaves`; se listan los lugares revisados. El umbral de tiempo de los comandos durante el plan va al usuario (D3), y la revisión en paralelo por áreas está en investigación (D2).

2026-11-21 — D2 A y D3 A registrados. Re-revisión retomando al mismo revisor (44 s, 137k tokens): los 13 hallazgos anteriores cerrados; 1 P2 nuevo (AC1 no cubría el límite de 5 minutos de D3) y 3 P3 (el umbral nombra la regla de 30 minutos que reemplaza al planear; redacción del servicio que el build levanta; el párrafo de costo cita D3), todos aplicados. Se acota «base de datos» a la que no es local (supuesto). Sin más rondas.

2026-11-21 — Revisión durante el build: corrección del usuario (comandos en paralelo) y texto compactado por `reading_footprint_test.go` y `leaf_skill_test.go`; no invalidó tareas. Tercera ronda, retomando al mismo revisor (64 s, 161k tokens): 2 P2 (la forma más cercana solo cubría pruebas; faltaban «esperar a todos» y «los demás uno tras otro») y 3 P3 (actualizar la línea, la nota del diseño, esta línea de avance), aplicados en `flow-plan` paso 3 con el presupuesto medido (−12 B). Sin otra ronda: no cambió un criterio ni el delta.

2026-11-21 — Verify: `/code-review` de la rama (3 importantes, 6 menores). Se corrigieron los 3 importantes y los menores 4 y 7 (redacciones que un modelo literal malinterpreta), más el sitio (8 y 9), en `33c4b85b`. El delta ganó la excepción «lo que agrega una tarea» y «no cambiar de rama ni hacer stash», como pedía el hallazgo 3. La revisión limitada a esos arreglos (`hive-review-code`) dio 2 P3 que hacían falsos AC1 y AC3 («its modules» ambiguo; el revisor eximía una línea «pasa» de algo que agrega una tarea), corregidos sin otra revisión. Quedan para el cierre los menores 5 (prueba desechable sin imports del proyecto) y 6 (borrar la prueba desechable en cualquier salida).
