# conditional-plan-reviewers — un revisor del plan por dominio cuando el plan cruza varios

| Campo | Valor |
| --- | --- |
| Estado | En verify |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `49ef7757` · rama propuesta `feat/conditional-plan-reviewers` |
| Playbook | ninguno |

## Decisiones

2026-11-21
- **D2 A** — ¿Revisores del plan en paralelo? Sí, en un cambio aparte y condicional: un revisor por dominio solo cuando el plan cruza dos o más, el ángulo de UX para pantallas de captura, y la re-revisión retomando al mismo revisor (`engram:tricell-hive/decision/conditional-parallel-plan-reviewers`).
- **D4** — ¿Qué cuenta como superficie? Sin respuesta: el usuario preguntó por el reparto anterior, y D7 la reemplaza.
- **D5 A** — ¿Cuántos revisores por ronda como máximo? Tres, contando el de UX; con más dominios se agrupan los de menos tareas.
- **D6 A** — ¿Cómo cabe el texto nuevo en el presupuesto de lectura? Recortando texto existente sin quitar instrucciones. Se respondió con una estimación equivocada (150–190 B; eran 365); la decisión vigente es D10.
- **D8 A** — **Replaces:** D6. ¿Cómo cabe, con las cifras reales? Con los recortes sin pérdida de `design.md` «Presupuesto» (entonces estimados en 199 B; medidos después en 189 B), más una tolerancia en `reading_footprint_test.go` igual al crecimiento medido de este cambio (+9 B), para dejar 15 B de margen a las correcciones.
- **D9 A** — **Replaces:** D8. (Vigente hasta D10.) ¿Con qué tolerancia, si los textos exactos crecen +22 B y no +9? Con +22 B, la misma regla de D8 con la cifra real; los 15 B de margen quedan para las correcciones.
- **D10 A** — **Replaces:** D9. ¿Con qué tolerancia, si los arreglos de `/code-review` suman +160 B? Con +182 B, el crecimiento medido de este cambio con sus arreglos; el margen sigue en 15 B.
- **D7 A** — **Replaces:** D4. ¿Cómo se reparten los revisores? Por dominio, como antes del 30 de octubre, con las etiquetas server, client, data-entry screen, guidance e infrastructure (tabla en `design.md`). Un dominio cuenta solo cuando alguna tarea cambia archivos suyos según su `Locations:`.
- **Supuesto** — Con varios revisores, relee una revisión solo el revisor cuyo dominio tiene una tarea cuyo texto, `Closes:` o verificación cambió, o un criterio que esa tarea cierra; un cambio que no cae en un solo dominio lo releen todos. Alternativa descartada: que relean todos siempre. Para revertirlo, se quita esa sección de `plan-review.md`.
- **Supuesto** — Retomar al revisor se generaliza en `hive-rule-reviewer-brief`, solo para revisores del plan y de UX, con la misma condición que ya usa el revisor de UX: un mensaje al mismo hijo, o un hijo nuevo con los hallazgos anteriores si el host no puede mandar el mensaje o este falla. Alternativa descartada: una regla aparte en «Revise». Para revertirlo, se restaura el párrafo de UX.
- **Supuesto** — Con cuatro o más dominios se juntan los de menos tareas, y en un empate el dominio que va después en la tabla se une al anterior. Una tarea que toca dos dominios cuenta en los dos. Un archivo que ninguna fila nombra cuenta como server. La asignación de la pantalla de captura nunca se junta. Para revertirlo, se cambian esas líneas de la referencia.
- **Supuesto** — Los hallazgos repetidos entre revisores se juntan en uno, anotando qué revisores lo encontraron. Alternativa descartada: reportarlos por separado. Para revertirlo, se quita esa frase de la referencia.
- **Supuesto** — No se agrega un revisor de otro proveedor o modelo, aunque la evidencia externa lo favorece: es otra decisión, con costo y autorización propios. Para revertirlo, se planea aparte.
- **Supuesto** — La rúbrica de revisión de sesiones cambia su fila de delegación de «un `hive-review-plan`» a la regla por dominio. Alternativa descartada: dejarla. Para revertirlo, se restaura la fila.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Pedido

> «por cierto, creo que tambien vale la pena que sean revisores paralelos no? con diferentes enfoques, ahorita simepre se lanza solo uno» — después, «hagamos research para ver si vale la pena o no sumarlo a este plan», D2 A, y `/flow-plan` desde el cierre del cambio anterior.

Sin ticket.

## Objetivo

Hoy cada plan con carpeta de cambio recibe un solo `hive-review-plan`, y cada revisión que cambia un criterio o el delta lanza otro revisor nuevo, sin el contexto del anterior. Cuando el plan cruza dominios con archivos propios, como un backend y un frontend, un solo revisor deja pasar lo que un revisor de cada lado encontraría.

Al terminar, un plan cuyas tareas cambian archivos de dos o más dominios, o una pantalla donde el usuario captura o elige datos, recibe un revisor por dominio, hasta tres, lanzados juntos. Los demás planes siguen con uno. La re-revisión se hace con un mensaje al mismo revisor, o con un revisor nuevo que recibe los hallazgos anteriores cuando el host no puede mandar ese mensaje.

## Alcance y aceptación

Dentro: `content/skills/flow-plan/references/plan-review.md` (nuevo), `content/skills/flow-plan/SKILL.md` pasos 3 (recorte) y 5, `content/skills/hive-rule-change-folder/SKILL.md` (párrafo inicial y «Revise» paso 4), `content/skills/hive-rule-reviewer-brief/SKILL.md` (releer), `content/agents/review/hive-review-plan.md` (recortes de los pasos 1, 3 y 4), `tests/content/reading_footprint_test.go` (tolerancia de D10), `site/src/content/docs/flows/plan.md`, `_support/docs/harness-engineering/session-review-rubric.md`, `CHANGELOG.md` y el delta de `workflow-lifecycle`: agrega «A plan that crosses domains gets one reviewer per domain» y modifica «Plan revisions keep the change folder coherent», cuyo escenario hoy pide «one `hive-review-plan` child».

Fuera: revisores de otro proveedor o modelo, cambiar el rol `hive-review-plan` (el reparto lo hace el hilo principal), el mapeo de OpenCode `hive-review-plan` → `flow-plan/SKILL.md`, y `flow-build/references/spec-delta.md` (un cambio sin plan sigue sin revisión del plan).

- AC1 — En `flow-plan` instalado, el paso 5 lanza un `hive-review-plan` por dominio, según la referencia `plan-review.md`, cuando las tareas cambian archivos de dos o más dominios o una pantalla donde el usuario captura o elige datos; si no, uno · falso hoy porque el paso 5 dice «have one `hive-review-plan` child review it» · Observed in: `~/.agents/skills/flow-plan/SKILL.md` tras `hive update`
- AC2 — La referencia `plan-review.md` instalada trae: la tabla de dominios con qué examina cada uno; que un dominio cuenta solo cuando una tarea cambia archivos suyos; el ángulo de UX para pantallas de captura; un máximo de tres revisores, agrupando los dominios de menos tareas; el brief de cada uno con los contratos que comparte con los demás; y que los hallazgos repetidos se juntan anotando quién los encontró · falso hoy porque el archivo no existe · Observed in: `~/.agents/skills/flow-plan/references/plan-review.md` tras `hive update`
- AC3 — En «Revise» instalado, el paso 4 manda la revisión a los revisores del plan según `hive-rule-reviewer-brief`, que la hace releer con un mensaje al mismo hijo, o con un hijo nuevo con los hallazgos anteriores cuando el host no puede mandarlo o el mensaje falla, y con varios revisores solo a los que nombra `plan-review.md` · falso hoy porque el paso 4 dice «have one `hive-review-plan` child review the revised folder» y el párrafo de releer de `hive-rule-reviewer-brief` solo habla del revisor de UX · Observed in: `~/.agents/skills/hive-rule-change-folder/SKILL.md` y `~/.agents/skills/hive-rule-reviewer-brief/SKILL.md` tras `hive update`

## Requisitos previos

None: the build needs nothing from outside the repository.

## Contexto verificado

- Evidencia interna (360 revisiones de plan reales, Claude Code y Grok, antes y después del 30 de octubre de 2026):
  - Los revisores por dominio encontraron cosas distintas cuando el plan cruzaba superficies independientes, como gh-192 y COM-838, y se repetían cuando leían los mismos archivos, como gh-89, la fila de COM-896 y el login de dened.
  - Una ronda con varios revisores costó unas 2,6 veces los tokens.
  - Retomar al mismo revisor costó 39 s y 149k tokens, frente a 250k–430k de uno nuevo.
- Evidencia externa: Claude Code Code Review, Cloudflare y Comfy usan revisores en paralelo con un paso que verifica y quita duplicados. Una segunda ronda de revisión bajó el F1 de 0,376 a 0,263 (arXiv 2603.16244). Ninguna fuente mide la revisión de planes.
- La referencia anterior, con su tabla de dominios, está en `_support/workspace/2026-10-30-skills-rebuild/baseline/content/skills/flow-plan/references/plan-review.md`. La quitó `7a37381b`.
- Presupuesto en la base:
  - lo que se lee entre plan y build suma 184474 B, frente a una línea base de 184489 B (15 B de margen); la lista contada está en `tests/content/reading_footprint_test.go` y no incluye referencias nuevas;
  - `hive-rule-change-folder` pesa 3967 B, con un límite de 4000 B;
  - `flow-plan/SKILL.md` pesa 9332 B, con un tope de 16 000 B.
- El mapeo de OpenCode `hive-review-plan` → `flow-plan/SKILL.md` (`integrations/opencode/context-loader.js:6`) no cambia: el revisor no reparte dominios.
- Hoy afirman «un solo revisor» estos lugares:
  - `flow-plan/SKILL.md:14`;
  - `hive-rule-change-folder/SKILL.md:13` y `:31`;
  - `site/.../flows/plan.md:47` y `:59`;
  - `session-review-rubric.md:64`;
  - `workflow-lifecycle/spec.md:164`.
- Comandos de verificación, ejecutados sobre la base `49ef7757` el 2026-11-21, en paralelo:
  - `go vet ./... && go test ./...` · pasa en 3 min 51 s
  - `go test ./tests/content/...` · pasa
  - `pnpm --dir site check && pnpm --dir site lint && pnpm --dir site build` · pasa en 3,7 s (0 errores)
  - `python3 -I ~/.agents/skills/flow-plan/scripts/check_change.py _support/openspec/changes/conditional-plan-reviewers` · sale 0 · 2026-11-21
  - `go test ./tests/content/ -run SpecDelta` · pasa · 2026-11-21

## Effects

None: the tasks take no effect on an external system.

## Entrega

`automatic` (en vigor en esta sesión y repositorio): revisión, push, pull request a `development` y merge sin detenerse; después, la actualización local de la instalación. Revisión de código: nativa, Claude Code `/code-review`. Sin revisión de seguridad.

## Estado de la revisión y avance

2026-11-21 — Plan escrito. Revisión de `hive-review-plan` (3 min 20 s, 154k tokens): 1 P1 (los borradores sumaban 365 B, no 150–190, y el único recorte de la hoja quitaba una instrucción), 4 P2 (sin regla de qué archivo es de qué dominio, ni de cuándo relee cada revisor; «cuando el host puede» no era observable; etiquetas distintas en cada sitio) y 6 P3. Aplicados: borradores compactados y medidos, tres recortes sin pérdida, la tabla de dominios con su regla de archivos, la condición de releer de `hive-rule-reviewer-brief` generalizada, el criterio de releer por dominio, el desempate, la prueba de enlaces correcta, las cifras y los lugares revisados. D6 volvió al usuario como D8.

2026-11-21 — Segunda ronda, retomando al mismo revisor (1 min 55 s, 181k tokens): los 8 hallazgos anteriores cerrados o casi. Quedaron 3 P2 nuevos y algunos P3:
- el párrafo de releer de `hive-rule-reviewer-brief` sumaba más de lo medido;
- los recortes del rol no tenían texto exacto;
- releer, generalizado, alcanzaba también la re-revisión de código;
- faltaba una fila para los archivos que ninguna nombra;
- la redacción del recorte de despliegue, el empate con la pantalla de captura, el escenario de cuatro dominios y textos sin actualizar.

Aplicados: textos exactos con sus bytes, releer limitado a revisores del plan y de UX, la regla «cuenta como server», la pantalla de captura que nunca se junta, el escenario con los conteos y el filtro de prueba ampliado. El neto medido es +22 B; D9 A lo acepta como tolerancia.

2026-11-21 — Tercera ronda, retomando al mismo revisor (56 s, 195k tokens): los hallazgos A–E de la segunda ronda cerrados y el plan listo para aprobar. Aplicados los 3 P3: D9 marca que reemplaza a D8; con cinco dominios se juntan asignaciones «hasta que queden tres»; el requisito MODIFIED nombra `hive-review-ux`.


2026-11-21 — Verify: `/code-review` dio 5 hallazgos importantes y 5 menores. Se corrigieron:
- «cuenta como server» partía planes de un solo dominio; los archivos de todo el repositorio ahora cuentan con su tarea;
- el paso 5 ahora lleva las reglas de conteo;
- una pantalla de captura cuenta también como client;
- releer cubre un plan revisado por una decisión y un revisor cuyo ID ya no existe;
- solo la primera asignación hace los chequeos de todo el plan;
- la regla de juntar dominios va de dos en dos;
- la descripción de la hoja y el requisito MODIFIED.

El enlace del sitio a `master` sigue la convención de las demás páginas. D10 A sube la tolerancia a +182 B.
