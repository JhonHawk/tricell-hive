# Fronteras entre flujos y skill `hive-core`

| Campo | Valor actual |
| --- | --- |
| Estado | Completed · fase 1 integrada en PR #242 (`96584b03`) y fase 2 en PR #243 (`25fe9d36`), 2026-10-28; fase 3 descartada (D16-B) |
| Tracker · GitHub Issues | • [#240 — Fronteras entre flujos y skill hive-core](https://github.com/JhonHawk/tricell-hive/issues/240) |
| Git | `interactive` por fase · 2 PRs a `development` (la fase 3 se descartó) · parada antes de cada push para tu validación |
| Verificación | `go test ./...` · checks de bytes y enlaces · `/code-review` por fase · refresco local tras cada merge |
| Siguiente paso | Ninguno: #240 cerrado con la integración de la fase 2 |

## Solicitud

Pedidos del usuario el 2026-10-27, textuales:

> Bien, vamos a analizar los flujos de Hive. […] Usamos la siguiente rubrica para todos (aunque me puedes apoyar a crecerla) […]

> Esto debe prepara un analisis de los hallazgos y su propuesta de resolución en formato de tabla, platicaremos de esto una vez que se acabe el research, guarda lo que consideres necesario, pero presentame lo relevante en este chat

> H1, de acuerdo, H12 y H17 es algo que practicabamos parciamente en el hive legacy, de hecho, varias cosas del global podrian irse a esta skill core  y cargar de forma lazy lo que necesite. 
>
> Haz research sobre el tema y ve como lo hacen los proyectos de referencia

> ahora, tambien hay que ver el valor real de un host vanilla, sumando a este research

> los hooks, siempre deben ser pensados para todos los hosts, no?

> ok, considerando esto, ahora si, comencemos con el plan sobre todos los temas

[#240](https://github.com/JhonHawk/tricell-hive/issues/240) (abierto el 2026-10-27 en esta sesión) resume los hallazgos y no tiene criterios propios distintos de los de abajo.

## Objetivo

El análisis de fronteras ([investigación](../../../sessions/2026-10-27-flow-boundary-analysis/flow-boundary-analysis.research.md), local y sin versionar) encontró 19 hallazgos, H1–H19. Los más graves:

- `flow-deploy` no autoriza de antemano la recuperación de producción.
- Un arreglo a mitad de una promoción arrastra commits ajenos.
- Hay reglas alojadas en la etapa equivocada.
- Tres archivos están en el tope de 16 KB y bloquean cualquier corrección.

Además, `global.md` pesa 42 KB, se carga en cada sesión y mezcla invariantes con procedimiento. La investigación sobre carga bajo demanda (Hive legacy, proyectos de referencia, documentación de los hosts y valor de un host sin Hive) muestra dos cosas. Las reglas que deben actuar sin lectura previa no pueden depender de que el modelo cargue una skill. El procedimiento, en cambio, sí puede vivir detrás de un disparador literal.

Este cambio hace tres cosas:

- Crea la skill compartida `hive-core` y le pasa las referencias transversales que hoy viven en `flow-plan`.
- Divide `global.md` en tres capas: invariantes siempre cargados, procedimiento en `hive-core` y reinyección tras compactar donde se compruebe pérdida.
- Corrige los hallazgos en sus hogares acordados.

El resultado esperado es una guía global de unos 15,5 KB, sin contradicciones entre etapas y con margen para corregirla.

## Alcance y aceptación

**Incluido:**
- **Fase 1:** recorte de tamaño (H7), tabla de qué sobrevive a la compactación por host (D9-A), skill `hive-core` con las cuatro referencias migradas (D6-B) y división de `global.md` en capas (D8-B, D12-A). Incluye los arreglos que caen sobre `global.md`: H4, H9 (parte global), H10, H15 (`global.md:78`) y H17 (resúmenes divergentes).
- **Fase 2:** los hallazgos restantes en sus hogares, según los fallos K1–K23 del análisis:
  - H1 (D5-A), H2, H3, H5, H6, H8, H9, H11, H13, H14, H15, H16, H18, H19.
  - Las correcciones de tracker de `plan-format.md:30`.
  - R2 de la revisión de sesiones reales (D22-A): el disparador de tool-scripts depende del host.
  - Menos fricción con el usuario (D23-A, D24 A–D, 2026-10-28): un efecto que el usuario pide sobre un objetivo concreto queda autorizado, y una respuesta de ruta o de release cubre lo que prescribe el procedimiento documentado del proyecto (FR1, FR2; reemplaza a R1); las notas de una línea van sin línea de contexto (FR3); una confirmación de sí o no puede ir en una línea (FR4); `questions.md` no se relee al llegar cada respuesta (FR5). Más FA1: los recordatorios de los hooks de memoria no cambian cuándo se escribe memoria.
- **Fase 3 (descartada por D16-B, 2026-10-28):** el hook de reinyección. Claude Code y Codex reinyectan el global completo tras compactar, y ninguna de las cuatro sesiones revisadas perdió una regla (ver Entrega).
- Actualizar `README.md`, `llms.txt`, `site/` y `_support/docs/architecture/` en el PR de cada fase que cambie lo que describen, como pide `AGENTS.md`.

**Excluido:**
- #205 (orden de entrada a build), #207 (preguntas en texto), #210 (revisión de PRs ajenos) y #188 (variantes en paralelo). Son temas aparte. Citan archivos que este cambio mueve o divide, así que sus rutas quedarán viejas tras la migración. Este cambio no escribe en esos issues.
- Un techo de tamaño para `global.md` (D7-B).
- Migrar a `hive-core` referencias distintas de las cuatro de D6-B. `delivery-decisions`, `review-decisions` y `ui-verification-decision` se quedan en `flow-plan` y solo se catalogan.

**Restricciones:**
- Cada `SKILL.md` y cada referencia queda en 16.000 B o menos (`tests/content/skill_size_test.go`).
- Cada regla tiene un solo hogar canónico (`AGENTS.md`, "Guidance structure").
- La autorización, la preservación, los secretos y la colocación básica de artefactos permanecen en el núcleo global (spec `context-delivery`, "Portable kernel and situational procedures").
- `go test ./...` pasa.
- No se nombran los repositorios de referencia en contenido versionado.

**Búsqueda en specs:** sobre `_support/openspec/specs/`, que tiene cuatro capacidades (`context-delivery`, `versioned-installation`, `walkthrough-viewer` y `documentation-site`) y ningún mapa de producto.
- **Términos buscados:** `compact`, `hook`, `PreToolUse`, `SessionStart`, `global.md|global core|kernel`, `catalog|inventory`, `skill`, los cuatro nombres de referencia, `flow-plan/references`, `context line|reporting`, `ticket|state`, `rollback|deploy|promot|Completion|incidental|open topic|authoriz`, `limited to|SHALL NOT`, `size|bytes|16,000`, `read whole` y `settings.json|host configuration`.
- **Requisitos leídos:**
  - En `context-delivery`: "Portable kernel and situational procedures", "Kernel references validated before installation", "Optional inactive OpenCode context resource", "Fixed-role context delivery without authority expansion" y "Declared source footprint without token claims".
  - En `versioned-installation`: "Optional operation capabilities" (lista cerrada), "Deployed skill resource links resolve to installed paths", "Deleted managed content is reinstalled" y "Project settings section editing".
  - En `documentation-site`: "Published documentation site" y "Flow pages describe without copying rules".
- **Resultado:**
  - Este cambio añade un requisito a `context-delivery` (ver `specs/`).
  - No modifica "Portable kernel", porque la capa 1 lo cumple.
  - El nombre `hive-core` no usa el prefijo `flow-`, así que `documentation-site` no cambia.
  - La fase 3, que se descartó, habría necesitado deltas en `context-delivery` y `versioned-installation`. Ya no aplica.
- No hay specs sobre estados de tickets, autorización de deploy o rollback, línea de contexto ni compactación.

**Aún no especificado:** nada. La fase 3 se descartó por D16-B (ver Entrega, 2026-10-28).

### Criterios de aceptación

**Fase 1**
- AC1. `_support/docs/architecture/compaction-survival.md` registra, para los seis hosts, si sobreviven a la compactación el archivo global de instrucciones y el cuerpo de una skill invocada, con su fuente (URL de documentación y versión, u observación). *Falso en la base cuando* el archivo no existe.
- AC2. `wc -c` da como máximo 14.500 B para `content/skills/flow-close/SKILL.md`, 15.100 B para `content/skills/flow-build/SKILL.md` y 15.800 B para `content/skills/flow-close/references/reporting.md`. *Falso en la base cuando* miden 15.850, 15.591 y 15.995.
- AC3. `questions.md`, `scope-changes.md`, `effect-authorization.md` y `project-settings.md` existen bajo `content/skills/hive-core/references/` y no bajo `content/skills/flow-plan/references/`, y todos sus enlaces entrantes resuelven en la validación de la release. *Falso en la base cuando* los cuatro están en `flow-plan/references/`.
- AC4. `content/guidance/global.md` mide como máximo 17.500 B (tope subido desde 16.500 por D20-A, 2026-10-27). Cada enlace que contiene (`](skill:` o `](<skill:`) está en una oración que dice literalmente cuándo leer el archivo enlazado, nombrando un acto o un evento ("Before a commit…, read…", "At a work-item close, read and run…"), y ese archivo existe. *Falso en la base cuando* mide 42.001 B.
- AC5. Cada regla que sale de `global.md` aparece una sola vez bajo `content/`, en su destino. Lo comprueba el check de unicidad sobre la lista de frases que T4 escribe en `design.md`. *Falso en la base cuando* esas reglas solo están en `global.md`.

**Fase 2**
- AC6 (H1). El paso 4 de `flow-deploy` pide la acción de recuperación de cada disparador dentro de la autorización de producción, con objetivo y reversa, y el paso 6 la ejecuta cuando esa respuesta la cubre. *Falso en la base cuando* `flow-deploy/SKILL.md:23` dice "runs only when an existing authorization covers it".
- AC7 (H2). Al reentrar tras un arreglo a mitad de corrida, `flow-deploy` lista los commits que no estaban en el candidato fijado y los trata como un cambio de alcance de clase Promoción. *Falso en la base cuando* solo `urgent-fixes.md:7` lista commits.
- AC8 (H11). `project-settings.md` pasa un ticket a su estado final solo cuando no queda ningún recorrido funcional pendiente en ese entorno. *Falso en la base cuando* `project-settings.md:29` lo mueve "when its ref holds the work" sin esa condición.
- AC9 (H3). `review-and-gates.md` aplica a los hallazgos de seguridad sobre un candidato sin integrar sus reglas de corrección P0–P3, refutación y re-revisión, con las definiciones de severidad enlazadas a `incidental-findings.md`, y `flow-verify` corre la revisión de seguridad en paralelo con la de código. *Falso en la base cuando* `flow-verify/SKILL.md:39` dice "in order" y `review-and-gates.md:16` dice "without waiting".
- AC10 (H6). "Dedicated review timing" tiene su hogar en `flow-verify/references/review-and-gates.md`, y `review-decisions.md` solo apunta a él. *Falso en la base cuando* la regla vive en `review-decisions.md:21`.
- AC11 (H5). `flow-build` manda leer el Scope y la inspección del implementador de `ui-gate.md` antes de reportar una tarea con efecto visual. *Falso en la base cuando* `flow-build/SKILL.md:40` enlaza `ui-gate` solo para los hijos.
- AC12 (H14). `flow-build` manda leer `verification.md` antes de elegir la selección de pruebas afectadas en cualquier build. *Falso en la base cuando* `flow-build/SKILL.md:22` lo limita a runtime, UI, rendimiento, hardware o checks costosos.
- AC13 (H6). `flow-build` declara la regla de commits para todo commit posterior del trabajo en cualquier etapa, y `flow-verify` apunta a ella. *Falso en la base cuando* la regla solo cubre "without a retained plan" y verify no la menciona.
- AC14 (H8). "When to run it" de `flow-close` incluye un disparador para la sesión que encuentra integrado el PR de un ítem cerrado bajo revisión humana, y para el movimiento final de `Completion`. *Falso en la base cuando* "the next session that sees the merge" no tiene disparador.
- AC15 (H9). La escritura de memoria de `flow-close` es un paso propio que va después de los hallazgos. `incidental-findings.md` e `instruction-resources.md` lo citan por su nombre, y `global.md` no cita ningún número de paso de `flow-close` (eso lo resuelve T4 al aplicar K23). *Falso en la base cuando* el paso 3 dice "after step 5" y `global.md:34` cita "step 5".
- AC16 (H6). "A validation stop or a question that waits is not a pause" vive en "When to run it" de `flow-close`, y `flow-build` y `flow-verify` apuntan ahí. *Falso en la base cuando* solo aparece en `flow-build/SKILL.md:68`.
- AC17 (H13). `diagnosis.md` limita las sondas y los fallos forzados de una investigación sin implementación autorizada al árbol de trabajo y a fixtures locales, y exige retirarlos o listarlos como pendientes antes de salir. *Falso en la base cuando* `diagnosis.md:11` solo los retira "after correction or mitigation".
- AC18 (H15). `project-settings.md` dice que una clave retirada solo informa la opción recomendada de la pregunta que antes respondía, sin contradecir a `review-decisions.md:17`. *Falso en la base cuando* `project-settings.md:11` dice "answers no question: ask as if it were absent".
- AC19 (H15). `flow-research` y `plan-format.md` resuelven `Tracker` como manda `project-settings.md`, y `project-settings.md` cubre los enlaces a issues y las integraciones configuradas como simples recomendaciones. *Falso en la base cuando* `flow-research/SKILL.md:42` y `plan-format.md:30` lo infieren.
- AC20 (H16). `scope-changes.md` manda a la pregunta de ruta un cambio nuevo durante la investigación, e indica qué pasa con los commits, la rama y el ticket de un build que vuelve a plan. *Falso en la base cuando* `scope-changes.md:13` remite a la pausa de `flow-close` y `:37` no dice nada del estado producido.
- AC21 (FR1 y FR2, D24-A; reemplaza a R1 de D22-A). AC21 cubre cinco puntos:
  - El punto **Critical** y la fila Critical de la tabla de `effect-authorization.md`, la frase de efecto crítico de `global.md` y la frase de efectos críticos de `questions.md:35` dicen lo mismo. Un efecto crítico queda autorizado sin otra pregunta cuando el mensaje del usuario lo pide sobre un objetivo concreto, sea porque lo nombra o porque aprueba la propuesta del mensaje inmediatamente anterior. Esa propuesta tuvo que nombrar el objetivo exacto, si se puede revertir y, cuando gasta, el costo y su techo. Ejemplo: "corre el script", después de describir lo que borra.
  - Un "ok" genérico a un plan con varios efectos no cubre los que ese plan no nombraba. Un efecto crítico que el agente propone sin esa aprobación, o uno sobre un objetivo que el mensaje no nombra, conserva su pregunta propia.
  - "What an answer covers" dice que una respuesta de ruta, de release o de despliegue que nombra una entrega cubre también los efectos que prescribe el procedimiento documentado del proyecto para esa entrega, sobre los objetivos que ese procedimiento nombra. Ejemplo: el merge con administrador o el force-push que pide su guía de release.
  - **Qué cuenta como procedimiento documentado:** un archivo versionado del repositorio o del workspace, leído en la sesión, que nombra el efecto y su objetivo. No cuentan la memoria, la documentación externa ni la salida de las herramientas.
  - **Quién autoriza:** el procedimiento solo acota los objetivos; la autorización la da la respuesta del usuario. `effect-authorization.md:28` ("a critical effect not already authorized for that exact target"), `reporting.md:98` entero, `delivery-decisions.md:9` y el desvío de `git-workflow/SKILL.md:42` lo dicen sin contradecirse. `git-workflow/SKILL.md:44`, el desvío de la ruta de promoción, no cambia.
  - "When to read" de `effect-authorization.md` y `hive-core/SKILL.md` agregan "or acting on an earlier grant", como ya dice `global.md:25`.
  - El paso 4 de `flow-deploy` enlaza `#what-an-answer-covers`.
  - La mitad de cobertura de `reporting.md:98` apunta a esa ancla en vez de repetir la regla.

  *Falso en la base cuando* se cumple todo esto:
  - `effect-authorization.md:22` dice "It holds even when the user's request names the effect" y `global.md:25` dice "even when the request names it".
  - `effect-authorization.md:26-28` no menciona el procedimiento documentado del proyecto, y `questions.md:35` pide pregunta propia para cualquier efecto crítico.
  - `reporting.md:98` repite la regla de cobertura sin enlazar `effect-authorization`.
  - `effect-authorization.md:5` y `hive-core/SKILL.md:12` omiten "acting on an earlier grant".
  - `flow-deploy/SKILL.md:21` enlaza la referencia sin ancla.
- AC23 (FR3, D24-B). La línea de contexto se exige en los informes, las entregas de resultado, los cierres y las preguntas. Una nota de progreso de una línea, como la línea entre tareas, una ronda, un tramo de espera o lo que resuelve una respuesta, va sin ella. Así lo dicen `global.md` (Communication), "Open with the context line" y el paso 5 de "Report a running or completed item" de `reporting.md`, y la línea de tramo de espera de `delegation.md`. *Falso en la base cuando* `global.md:13` dice "a one-line note included" y `reporting.md:13` dice "however short".
- AC24 (FR4, D24-C). En `questions.md`, una confirmación de sí o no sobre una acción que el mensaje describe, sin otra alternativa que hacerla o no, puede ser la propia línea de pausa: la etiqueta fija seguida de la pregunta, sin IDs, opciones ni tarjeta, aunque el host ofrezca una tarjeta nativa. Una decisión con alternativas reales conserva el formato completo. `questions.md:19` ("never as a closing sentence alone"), `:47` (condición de la tarjeta) y `:76` (línea de pausa "by ID") reconocen esta excepción. *Falso en la base cuando* `questions.md:23` dice "never a bare \"do you authorize it?\"".
- AC25 (FR5, D24-D). Ni `global.md`, ni `hive-core/SKILL.md`, ni "When to read" de `questions.md` mandan leer `questions.md` cuando llega una respuesta. La sección "Pause and handle the answer" no cambia. *Falso en la base cuando* los tres dicen "when an answer arrives".
- AC26 (FA1). La regla de cuándo escribir memoria de `global.md` dice que también prima sobre los recordatorios de los hooks de memoria, como un aviso tras compactar o uno periódico. *Falso en la base cuando* `global.md:83` solo nombra "memory tools' own save instructions".
- AC22 (R2, D22-A). El disparador de `tool-scripts` en `global.md`, y "When to read" de `tool-scripts.md`, solo aplican cuando el host ofrece una herramienta que corre un script capaz de llamar a otras herramientas. *Falso en la base cuando* `global.md:69` y `tool-scripts.md:5` aplican en cualquier host.

Restricciones que deben seguir valiendo en todas las fases: `go test ./...` pasa, cada skill y referencia mide 16.000 B o menos, y no se repiten reglas entre archivos.

## Entrega

- **Respuestas del 2026-10-27:**
  - D10-A: un cambio en tres fases, con un PR por fase.
  - D11-A: la skill se llama `hive-core`.
  - D13-A: entrega `interactive`.
  - D14-A: revisión con `/code-review` de Claude Code.
  - D15: autorizado abrir el issue paraguas (hecho: #240). Las pruebas de compactación por host no están autorizadas y se preguntan al llegar a T1.
- **Repositorio:** `tricell-hive`, base `development`.
- **Por fase:**
  1. Rama de trabajo desde `development`.
  2. Commits locales por tarea.
  3. Verificación y `/code-review` sobre el candidato local.
  4. Parada antes del push para tu validación.
  5. Tras validar: push, PR y merge una vez que pasen los checks.
- **Ticket (D16-B, 2026-10-27):** #240 pasa a In Progress al empezar la fase 1 y sigue abierto hasta la fase 3. Si T1 no encuentra pérdida en ningún host, o si las pruebas que siguen tampoco la encuentran, la fase 3 se descarta, se cierra #240 y el registro se cierra en ese momento. Los merges de las fases 1 y 2 no cierran el ticket. **Aplicado el 2026-10-28:** la fase 3 se descartó, así que #240 y el registro se cierran cuando se integre la fase 2.
- **Cierre del registro:** este registro se versiona una vez, cuando se cierra #240, con un PR pequeño sin revisión dedicada que mergea el agente.
- **Pruebas de compactación (D19-B, 2026-10-27):** autorizadas después de T1. Corre una sesión corta por host en un proyecto desechable del scratchpad, sin editar ninguna configuración real. Los resultados completan `compaction-survival.md`.
- **Tope de AC4 (D20-A, 2026-10-27):** 17.500 B. El de 16.500 B se estimó antes de fijar la capa 1 y no alcanzaba para conservar sus reglas.
- **Permiso para mover tickets (D17-A, 2026-10-27):** el invariante se queda en la capa 1 y el detalle va a `hive-core`. D17-A reemplaza a D12-A solo en U2.
- **Revisión de sesiones reales (D22-A, 2026-10-27; en la sesión se preguntó como D1):** se revisaron tres sesiones abiertas después del despliegue de la fase 1: carvento y cam en Claude Code 2.1.312 con Opus 5.5, y dened en Grok con grok-4.7-build-fast. Ninguna rompió rutas y ninguna se compactó. Lo que se decidió:
  - R1 (AC21) y R2 (AC22) entran en la fase 2. R1 se retiró el 2026-10-28 por D24-A, y AC21 ahora dice lo contrario (ver abajo). R1 era una corrección vista en una sola sesión; la adoptas ahora, como permite `AGENTS.md` (Measurement), porque toca producción. El commit nombra la sesión de Grok `01a180ab-1142-7f20-94f4-3cd2257186f3`. R2 corrige un disparador ambiguo y no necesita recurrencia.
  - Quedan en observación, con su conteo de sesiones:
    - H6: falta la línea de contexto en notas cortas de una línea. Conteo 3 (carvento, cam, dened). Ya cumple la recurrencia, pero no hay línea base anterior al despliegue; se reabre si una comparación con sesiones previas muestra que la separación en capas la empeoró.
    - H9: `questions.md` no se relee cuando llega la respuesta. Conteo 2, con el archivo todavía en contexto, así que el valor de corregirlo es bajo.
    - Memoria sin leer `memory.md`: conteo 1 (carvento).
    - `## Hive` resuelto sin leer `project-settings.md`: conteo 1 (dened).
    - Espera externa sin leer `delegation.md`: conteo 1 (dened).
    - Cada una se reabre con una segunda sesión independiente.
  - H7 (`evidence-tools.md` sin leer en dos sesiones de Claude) no cambia nada: el bloque CodeGraph del usuario, que está fuera de Hive, ya produce la conducta correcta.
- **Segunda revisión de sesiones y menos fricción (D23-A y D24 A–D, 2026-10-28):** se revisaron los tramos restantes de carvento, cam y dened, más la sesión Codex de cam `01a180de-c680-7f91-81d7-6c50a9e02ef1` con sus 12 hijos. Lo que se observó y lo que se decidió:
  - **Compactación:** carvento (Claude Code, línea 1944) y Codex (línea 1030) compactaron. Los dos hosts reinyectaron el global completo (carvento línea 1980, Codex línea 1031), y no se observó pérdida de ninguna regla de la capa 1. Grok no compactó (`compactionCount: 0`). Por D16-B se descarta la fase 3. #240 se cierra con la fase 2. No se observaron Pi, OpenCode ni Cursor.
  - **Lo que el usuario validó:** según el usuario, la sesión de Grok en dened "funcionó como esperaba". Ejecutó efectos críticos que su mensaje nombraba ("corre el script", "borra ambas") y efectos que la guía de release del proyecto prescribe (merge con administrador, force-push). Por eso R1 se retira y AC21 pasa a decir lo contrario. El usuario pidió quitar fricción y burocracia, y eligió FR1–FR5 para la fase 2.
  - **FA1 se adopta por recurrencia:** el hook de Engram empujó escrituras de memoria fuera de la regla de cuándo escribirla en dos sesiones independientes. En cam Claude `4afe6d47` hubo dos guardados a mitad de trabajo tras un `MEMORY REMINDER`. En Codex `01a180de` hubo dos `mem_session_summary` tras el aviso posterior a la compactación.
  - **No eran fallos:** los pushes directos a `master` en cam-specs y carvento-specs. Los dos repositorios declaran `Delivery: direct-base`.
  - **Observaciones que se cierran:** H6 (línea de contexto en notas) queda resuelta por FR3, y H9 (relectura de `questions.md`) por FR5.
  - **Observaciones que siguen abiertas, con conteo:**
    - Memoria escrita sin leer `memory.md`: 2 (carvento y Codex). Ya cumple la recurrencia, pero no hay una regla más estricta que proponer; FA1 quita la causa principal.
    - Afirmaciones de estado sin verificar: 3, en situaciones distintas (carvento tras compactar, dened y Codex).
    - Codex lee varios archivos en un solo `cat` y la salida llega truncada: 1.
    - flow-verify sin leer antes de un PR o un merge: 1 (carvento).
    - Tarjeta de pregunta sin mensaje visible: 1 (carvento).
    - `## Hive` resuelto sin leer `project-settings.md`: 1 (dened).
    - Espera externa sin leer `delegation.md`: 1 (dened).
- **Refresco local:** tras cada merge se refresca la instalación local (`go build` del binario si cambia `tooling/` o `integrations/`, luego `hive update`), porque existe `_support/workspace/maintainer.local`.
- **Revisión de seguridad:** con la fase 3 descartada, ninguna fase cruza un límite de seguridad. La revisión de seguridad corre con `flow-security` solo si la pides.
- **Revisión de UI:** sin efecto visual. Las páginas de `site/` solo cambian texto y enlaces.

## Estado

**Listo para implementar la fase 2.** Diseño, entrega y revisión del plan cerrados, incluida la re-revisión de D23-A y D24 del 2026-10-28.

**Re-revisión por D23-A y D24** (2026-10-28, un revisor `hive-review-plan` de coherencia de la guía, unos 153 k tokens y 2,7 min). Encontró cuatro bloqueantes:
- **B1:** FR2 chocaba con reglas que T11 no editaba.
- **B2:** FR1 chocaba con `questions.md:35` y con la tabla de `effect-authorization.md:16`.
- **B3:** AC24 chocaba con `questions.md:19,47,76`.
- **B4:** faltaba una comprobación por lectura de la frase nueva de `global.md:25`.

También dejó cinco observaciones no bloqueantes, N1–N5. Todo se aplicó con la propuesta del revisor, sin otra ronda, con una excepción en B1. El revisor proponía quitar "de ruta" de FR2. Lo conservé porque así lo eligió el usuario en D24-A, y tomé la alternativa que el propio revisor dio: editar enteras las reglas que chocan.

N5 queda como riesgo anotado. Al quitar la relectura al llegar una respuesta, la confirmación de selección de `questions.md:86` se apoya en la lectura anterior a la pregunta.

**Re-revisión** (revisión ba73a594aaca, los mismos tres revisores, unos 504 k tokens y 1,5 min en total): los 16 bloqueantes quedaron resueltos y el orden de los checkpoints es correcto. Aparecieron cuatro problemas nuevos, y todos los corregí aplicando la propuesta del propio revisor, sin otra ronda:
- La prueba en uso real usaba `--dry-run` con `--out`, combinación que no guarda nada, y `.target.path`, cuando el campo es `.target.Path`. Ahora usa `update --out` en un directorio temporal.
- El mensaje de `hive doctor` está fijado en cinco tests y en `deployment-manager.md:77`. Esos archivos se suman a T9.
- La nota de la fase 2 decía que T7 corría en paralelo. Ahora dice que va después de T6.
- AC4 ahora también revisa los punteros sin enlace.

**Límites que quedan:** no se ejecutó nada, y la redacción final de las reglas se comprueba en el build.

**Revisión del plan, ronda 1** (revisión 44d5a7cce859, 2026-10-27). Tres revisores `hive-review-plan`:
- coherencia de la guía: unos 189 k tokens, 5 min;
- manager y tests: 165 k, 3 min;
- entrega y orden: 116 k, 2 min.

Encontraron 16 bloqueantes en total. Los reconcilié así:
- **Decisiones del usuario:** D16-B (cierre de #240) y D17-A (permiso para mover tickets en la capa 1, que respeta la spec "Portable kernel").
- **Arreglos aplicados sin preguntar,** cada uno con su razón:
  - Más verificaciones, para que cada AC tenga su comprobación.
  - La comprobación de AC4 ahora busca todos los enlaces `](skill:`.
  - AC5 se apoya en una lista de unicidad, escrita en T4 antes de editar.
  - T8 corre después de T6 y T7, porque escribe en sus archivos.
  - T6 es dueña de todos los enlaces a `#dedicated-review-timing`.
  - Se agregan a las Locations de T4 y T9 los consumidores que faltaban.
  - Se respetan los límites de `TestRetiredHiveKeysMatchTheGuidance`.
  - Se actualizan el mapa `moved` y los escenarios de `reading_footprint_test.go`.
  - Se añaden el `rg` de anclas y la búsqueda en `site/` para T2.
  - Cada PR lleva su entrada en `CHANGELOG.md`.
  - Se añade un aviso para reiniciar las sesiones tras la fase 1.
  - El refresco despliega el commit del merge.
  - La prueba de uso real usa `--out plan.json`.
  - En `design.md`, cada regla se identifica por su texto y no por número de línea; se corrige la lista de descartes y se mantienen en la capa 1 la identidad de Engram y `:119`.
  - El delta de spec se acota a "read pointers".
  - Se corrige el razonamiento sobre las URLs a `master`: el sitio se publica desde `master`, así que no se rompe.
