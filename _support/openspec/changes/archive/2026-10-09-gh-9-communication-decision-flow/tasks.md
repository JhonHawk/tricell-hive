# Tareas: flujo de decisiones y pausas

Commit base: `9c285026` (`development`, 2026-10-08), rama `fix/gh-9-communication-decision-flow`.

## Tareas

### T1 — Reglas de elección y pausa en la guía global

- [x] `content/guidance/global.md` aplica las reglas 1 a 8 de [design.md](design.md#invariante-y-pausa), conserva las intenciones reubicadas y cabe en el presupuesto sin subirlo.

**Closes:** AC1, AC2, AC3, AC4, AC5, AC8, AC9

**Depends on:** ninguna.

**Locations:** `content/guidance/global.md:16`, `:17`, `:21`, `:22`, `:23`, `:28`, `:30` y `:33`; `tests/content/budget_test.go:12`; `CHANGELOG.md` (`[Unreleased]`).

**Execution:** hilo principal. Es la redacción central de la que dependen T2 a T4.

**Test approach:** check. Una búsqueda por criterio, la lectura de cada regla y la prueba de presupuesto.

**Changes:** reescribir los párrafos según el diseño, con las eliminaciones que nombra la sección [Tamaño](design.md#invariante-y-pausa), y agregar la entrada de `CHANGELOG.md`. Las reglas fuera del alcance no se tocan. Subir `globalGuidanceBudget` en `tests/content/budget_test.go:12` solo lo necesario, hasta 46.040 como máximo (D5-A).

**Verification:** cada búsqueda sobre `content/guidance/global.md`:
- AC1: `rg -n 'not the option list' …` no devuelve nada. Leer la regla 1 y confirmar que nombra decisión, ruta, cierre, oferta y sí o no, y que exige el contexto, las opciones con efecto y la recomendada aunque existan copias.
- AC2: `rg -n 'count as having no native question tool' …` no devuelve nada, y `rg -n 'that waits for the answer, also ask' …` encuentra la regla 2.
- AC3: `rg -n 'in its label when the tool has no description field' …` encuentra la regla.
- AC4: `rg -n 'dismissed or declined question is no answer' …` encuentra la regla.
- AC5: `rg -n -i 'pause line' …` encuentra la regla 3, y `rg -n 'ask it once with its options' …` no devuelve nada.
- AC8: `rg -n 'several openings apply' …` encuentra la precedencia, y `rg -n 'before the detail that supports it' …` no devuelve nada.
- AC9: `rg -n 'must weigh stay in the message' …` encuentra la regla, y `rg -n 'a linked file supports it, never replaces it' …` no devuelve nada (un solo hogar).
- Intenciones conservadas: cada una de estas búsquedas, por separado, encuentra su frase en la regla nueva y no solo en texto de la base fuera del alcance: `a response holding only the tool call is the failure`, `so no tool call separates them`, `split across questions`, `no "Other" option or reply example`, `never a second report` (dentro de `:23` reescrito) y `Take the first answer that arrives`.
- Tamaño: `go test ./tests/content/...` pasa, presupuesto incluido, y `tests/content/budget_test.go` sube la constante a 46.040 como máximo (D5-A).

### T2 — Oferta de recomendaciones en la investigación

- [x] `flow-research` termina cada ronda de hallazgos sobre trabajo posible con la línea de pausa que ofrece las recomendaciones, y las presenta como elección cuando el usuario acepta.

**Closes:** AC6

**Depends on:** T1, por las reglas 1 a 3.

**Locations:** `content/skills/flow-research/SKILL.md:30`, `:32` y `:36`.

**Execution:** hilo principal. Es un cambio corto que depende de la redacción de T1.

**Test approach:** check.

**Changes:** como describe [design.md](design.md#investigación-d2-a).

**Verification:**
- `rg -n 'one line of plain text' content/skills/flow-research/SKILL.md` no devuelve nada.
- `rg -n -i 'offer to show recommendations|show the recommendations' content/skills/flow-research/SKILL.md` encuentra la oferta.
- Leer `:30`–`:36` y confirmar tres cosas: las opciones de ruta (`flow-plan`, `flow-build`, detenerse) solo aparecen después de aceptar o pedir continuar; la pregunta de ticket lleva su propio ID en la pausa; las alternativas que responden la pregunta investigada siguen en los hallazgos.
- `go test ./tests/content/...` pasa.

### T3 — Pausa de plan y preguntas de entrega

- [x] `flow-plan` presenta de inmediato, en un bloque con la etiqueta de pausa, las decisiones que cambian el diseño, y `delivery-decisions.md` deja las opciones siempre en el mensaje.

**Closes:** AC7

**Depends on:** T1.

**Locations:** `content/skills/flow-plan/SKILL.md` (sección «Ground the decisions») y `content/skills/flow-plan/references/delivery-decisions.md:14` y `:30`.

**Execution:** hilo principal.

**Test approach:** check.

**Changes:** como describe [design.md](design.md#plan-d3-a).

**Verification:**
- `rg -n -i 'pause label' content/skills/flow-plan/SKILL.md` encuentra la regla. Leerla y confirmar que pide la explicación, las opciones con su efecto y la recomendada en el mismo bloque, la línea de pausa al final y la detención del trabajo que depende de la decisión.
- `rg -n 'applies to options in a native question tool|in a text question, omit it' content/skills/flow-plan/references/delivery-decisions.md` no devuelve nada.
- `go test ./tests/content/...` pasa.

### T4 — Validación manual y cierre

- [x] El traspaso de UI, las rondas de corrección, la validación de la entrega y la pregunta de cierre usan la regla de elección y la línea de pausa.

**Closes:** AC10

**Depends on:** T1.

**Locations:** `content/skills/flow-build/references/verification.md:65`, `:67` y `:77`; `content/skills/flow-close/SKILL.md:26` (paso 7); `content/guidance/global.md:23` (comprobación).

**Execution:** hilo principal.

**Test approach:** check.

**Changes:** como describe [design.md](design.md#construcción-y-cierre). Antes de editar, verificar en la documentación de Claude Code, Codex, Grok, OpenCode y Cursor si los hijos en segundo plano siguen corriendo mientras una tarjeta espera la respuesta, y registrar la fuente en el progreso.

**Verification:**
- `rg -n 'question in text|without the native question tool' content/skills/flow-build/references/verification.md` no devuelve nada. Si T4 registra una excepción por capacidad porque la verificación por host no se cumple, su redacción no usa esas dos frases.
- `rg -n 'never the native tool, on every host' content/guidance/global.md` no devuelve nada.
- `rg -n -i 'pause line' content/skills/flow-build/references/verification.md content/skills/flow-close/SKILL.md` encuentra la pausa en ambos archivos.
- `go test ./tests/content/...` pasa.

### T5 — Tabla de tarjetas de preguntas por host

- [x] `agent-delivery.md` tiene una tabla nueva de tarjetas de preguntas verificada el 2026-10-08, y los párrafos que afirmaban «opciones solo en la tarjeta» quedan como historia.

**Closes:** AC11

**Depends on:** ninguna; puede correr en paralelo con T1.

**Locations:** `_support/docs/architecture/agent-delivery.md`: la tabla de la línea ~151 (solo su columna «Native question», que pasa a remitir a la tabla nueva) y los párrafos `:139–143` y los demás sobre las preguntas de Codex y Grok.

**Execution:** delegado a `hive-write-spec`. Es documentación con las fuentes ya reunidas en [design.md](design.md#contexto-verificado) y no comparte escritura con T1. El brief lleva esa sección, los números de issue, las reglas de etiquetas de [design.md](design.md#hosts) y la orden de no tocar las demás columnas de la tabla existente.

**Test approach:** check.

**Changes:** como describe [design.md](design.md#hosts). La tabla nueva lleva una fila por host: Claude Code 2.1.295, Codex 0.161.0, Grok 1.0.50, OpenCode v2.0.24, Cursor 2026.10.01 y Pi 1.0.4. Cada fila trae cinco datos con su etiqueta de evidencia:
- la herramienta y sus campos;
- si espera la respuesta;
- qué queda en el historial al responder o rechazar;
- el costo de duplicar las opciones;
- los issues abiertos.

**Verification:**
- Leer la tabla nueva y comprobar, fila por fila, que los seis hosts traen los cinco datos con su etiqueta.
- `rg -n 'keeps the option list in the tool' _support/docs/architecture/agent-delivery.md` solo aparece dentro de un pasaje marcado como historia.
- La fila de OpenCode de la tabla nueva nombra `question`, en lugar de «None found».

### T6 — Sitio de documentación

- [x] Las páginas de investigación, plan y cierre describen la oferta de recomendaciones, la pausa de plan y la línea de pausa.

**Closes:** AC12

**Depends on:** T2, T3 y T4.

**Locations:** `site/src/content/docs/flows/research.md:37` y `:54`, `plan.md:57–58`, `close.md:44` y `:57–58`.

**Execution:** delegado a `hive-write-spec` tras T4. Las páginas describen el comportamiento y enlazan la skill sin copiar sus reglas (AGENTS.md, «README and llms.txt»). `README.md` y `llms.txt` no describen este comportamiento y no cambian (comprobado con `rg` en la base).

**Test approach:** check.

**Changes:** reescribir los pasajes citados.

**Verification:**
- `rg -n 'names the route it recommends' site/src/content/docs/flows/research.md` no devuelve nada.
- `rg -n -i 'waiting for your decision' site/src/content/docs/flows/research.md site/src/content/docs/flows/plan.md site/src/content/docs/flows/close.md` encuentra una coincidencia en cada página; en la base no hay ninguna.
- Leer los tres pasajes y confirmar que describen la oferta (`research.md`), el bloque de pausa (`plan.md`) y la línea de pausa en el cierre (`close.md`).
- El sitio construye (comprobación compartida).

### T7 — Scripts de medición

- [x] La medición se puede repetir con scripts que solo reportan conteos, en `_support/workspace/2026-10-08-communication-decision-flow/`.

**Depends on:** ninguna.

**Locations:** el directorio temporal de esta sesión (`bare.py` y los scripts del subagente de sesiones reales, si siguen ahí) y el destino citado.

**Execution:** hilo principal.

**Changes:** como describe [design.md](design.md#medición):
- `bare.py` acepta fechas de inicio y fin.
- Ningún script escribe texto de las sesiones en el directorio retenido.
- El README nombra los hosts y el período, y dice qué números de la línea base reproduce cada script y cuáles no.

**Verification:**
- `python3 -I _support/workspace/2026-10-08-communication-decision-flow/bare.py 2026-10-01 2026-10-09` imprime un conteo coherente con la línea base. Resultado: `214 86` a las 18:45. El 208 de 82 se tomó a las 17:55 del mismo día, y la diferencia son tarjetas posteriores del 2026-10-08, incluida esta sesión. Con fin el 2026-10-08 (es decir, hasta el día 7) imprime `200 78`. El README lo registra.
- El directorio no tiene archivos `.json`: `norm2.py` escribe su volcado en el directorio temporal que recibe como argumento. `git check-ignore` confirma que la carpeta está ignorada (`.gitignore:35`).

## Verificación y revisión humana

| Comprobación | Momento | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Batería de CONTRIBUTING | candidato final, local | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasan | incluida en la implementación |
| Sitio | candidato final, local | `pnpm --dir site build` termina sin errores (`site/node_modules` presente) | incluida |
| Lectura del usuario | parada antes del push | el usuario lee el diff de `content/` y del sitio y valida la redacción; si pide cambios, se aplican en la misma rama y se repite esta parada | modo interactivo (E1-A) |
| Medición posterior | una semana después del despliegue | repetir T7 sobre sesiones reales y comparar con la línea base | se ofrecerá al cerrar; no autorizada aún |

No aplican la verificación de UI ni la verificación en vivo: el cambio es solo de texto, y las pruebas con modelos están en pausa.

## Revisión del plan y progreso

- **Revisión 68a195f2de0d**, dos revisores `hive-review-plan` en paralelo:
  - **Redacción de la guía:** 6 hallazgos, 3 de ellos altos. Se aceptaron todos:
    - la regla 2 conserva la frase del fallo de #9;
    - la regla 1 cubre toda elección;
    - se reubican las cinco intenciones que se perdían;
    - se resuelven las contradicciones de repetición, etiquetas y ticket;
    - las eliminaciones necesarias para el tamaño quedan nombradas;
    - las verificaciones van separadas por criterio.
  - **Entrega, pruebas y documentación:** 6 hallazgos medios y 3 menores. Se aceptaron todos:
    - entrada en `CHANGELOG.md`;
    - tabla nueva con sus propias versiones y las etiquetas del archivo;
    - los párrafos `:139–143` pasan a historia;
    - verificación de T6 con una frase que no existe en la base;
    - scripts de T7 que solo reportan conteos y aceptan fechas;
    - sesiones completas para los commits;
    - batería completa de CONTRIBUTING;
    - condición para reanudar tras la parada.
  - **Sin cambios necesarios:** el revisor confirmó que ninguna prueba existente afirma el texto que se elimina y que los detectores `question_after_detail` y `close_question_after_report` siguen siendo compatibles.
  - Una decisión de la revisión: se conserva «no reply example» de `de91b153`, así que la línea de pausa nombra lo que se pregunta por ID con su glosa, sin listar respuestas.
- **Re-revisión 6d6d1e2d68d9** (mismo revisor de redacción, única ronda):
  - Confirma resueltos F1 a F4 y las verificaciones por criterio.
  - Tres hallazgos nuevos, reconciliados con la propuesta del propio revisor:
    - N2: la regla 2 recupera «so no tool call separates them», el orden que evita el fallo de #9, y T1 lo busca.
    - N3: las intenciones conservadas se buscan por separado, y la excepción de T4 no puede usar las frases que se buscan.
    - Rule 3 dice «by ID … when it has one».
  - N1, el tamaño, sigue abierto. El borrador sumaba unos +700 bytes frente a un margen de 394. Con las reglas 2, 3 y 5 acortadas y una eliminación más en `:23`, la estimación baja a unos +480. Se pasó al usuario como decisión D5 (cómo cabe la guía), sin otra ronda. Respuesta: D5-A, subir el presupuesto hasta 512 bytes.
- **Progreso:**
  - Build iniciado sobre `9c285026` en `fix/gh-9-communication-decision-flow`.
  - Modelos: el implementador es el hilo principal, Claude Opus 5.5 (observado). `hive-verify-task` está configurado con `opus`, el mismo modelo, así que se lanza con `fable` (Claude Fable 5.1) mediante la opción de modelo del host.
  - T1 implementada: `global.md` mide 45.572 bytes y el presupuesto sube a 46.040 (D5-A). Las búsquedas por criterio dan el resultado esperado y `go test ./tests/content/...` pasa.
  - T1 verificada por `hive-verify-task` (Fable 5.1): AC1, AC2, AC3, AC4, AC5, AC8 y AC9 se cumplen. El orquestador abrió `global.md:23` y confirmó la línea de pausa. El presupuesto queda en 46.040, el máximo que autoriza D5-A, con 468 bytes de margen.
  - T2 verificada (Fable 5.1): AC6 se cumple; el orquestador releyó `flow-research/SKILL.md:36`.
  - T3 verificada (Fable 5.1): AC7 se cumple; el orquestador releyó `flow-plan/SKILL.md:16`. Se deja anotado, sin cambio, que «pause label» (el encabezado «Pausa») y la etiqueta de la línea de pausa son dos etiquetas distintas por diseño.
  - T4, verificación previa: la documentación de Claude Code dice que los subagentes en segundo plano «run concurrently» y que las preguntas «stay open until you answer them» ([Subagents](https://code.claude.com/docs/en/sub-agents.md), [Tools reference](https://code.claude.com/docs/en/tools-reference)). No dice cómo interactúa el aviso de un hijo que termina con una pregunta pendiente, y en los demás hosts no se verificó. Por eso, mientras haya hijos corriendo, la elección queda solo en el mensaje y el turno termina; la tarjeta se usa cuando no corre ningún hijo. Ninguna de las dos ramas nombra un host.
  - T4 verificada (Fable 5.1): AC10 se cumple. Se aplicó su observación menor: `verification.md:77` dice ahora «a native question card that waits for the answer», igual que `:65`.
  - T6 la escribió `hive-write-spec` y la verificó `hive-verify-task` (Fable 5.1): AC12 se cumple. Se aplicaron sus dos observaciones menores: `plan.md` dice que el trabajo dependiente espera, y `research.md` nombra también el pedido de hacer el trabajo.
  - Commits locales: `75fe4ce5` (guía, skills, sitio, presupuesto y CHANGELOG) y `4c75483b` (tabla de hosts).
  - Batería de CONTRIBUTING sobre el candidato: `go vet ./...` y `go test ./...` pasan, y las 23 pruebas de Python dan OK.
  - T5 verificada (Fable 5.1): AC11 se cumple. Los estados de los ocho issues se comprobaron en vivo. Las cinco celdas sin etiqueta de evidencia que señaló el verificador se corrigieron en el hilo principal; se agregó la etiqueta `report` para los informes de la comunidad. La tabla de presentación conserva sus versiones antiguas en las otras columnas, como manda el diseño.
  - T5 la escribió `hive-write-spec`. El hilo principal completó con datos verificados de la investigación de hosts las celdas que había dejado como «Not verified»: Claude Code respondida, campos y límites de Codex, espera e historial de Grok, OpenCode y Cursor, y la extensión de Pi.
- **Despliegue desde la rama (pedido del usuario, 2026-10-08):** se hizo push de `fix/gh-9-communication-decision-flow` (`4c75483b`), y `hive update --out` + `hive apply` instalaron la release `ff98f6208738` en los seis hosts. `hive status` marca los 486 archivos como `installed`. El usuario pidió quedarse en la rama por si hay más cambios.
- **Agregado fuera de las tareas (R1-A, usuario, 2026-10-08):** `3f7d15a6` suma a `AGENTS.md` («Guidance structure») la regla de revisar todo el comportamiento antes de cambiar una regla. Es un cambio de documentación, verificado por lectura; está publicado en la rama y no requiere despliegue.
- **Agregado fuera de las tareas (D1-A, usuario, 2026-10-08):** `722c7fe6` corrige el cierre por tarea que se vio en la sesión de Grok cam `01a11cbc`. Un trabajo terminado es el ticket o el alcance autorizado; la opción de continuar cubre lo que falta del plan con una sola pregunta de permisos; no se recomienda sesión nueva a mitad de un plan. Verificación: lectura y revisión de todo el comportamiento (lugares listados en el commit), `go test ./tests/content/...`, las pruebas de Python y el build del sitio pasan; `global.md` mide 45.751 bytes. Publicado en la rama y desplegado con la release `96d434eb`.
- **Agregado fuera de las tareas (D1-A, usuario, 2026-10-08):** `809acee3` corrige las paradas de la sesión de Grok cam `01a11cbc`. La línea de avance va con la siguiente acción; una pregunta o aviso que llega a mitad del trabajo se responde en una línea y se sigue; todo proceso externo lanzado lleva su aviso de término. `global.md` mide 46.022 de 46.040 bytes (18 de margen). Pasan las pruebas de contenido y de Python, y el build del sitio. Publicado y desplegado con la release `b371d6dd`.
- **Agregado fuera de las tareas (D1-A, usuario, 2026-10-08):** `8c69c3ef`. El aviso de término nombra el monitor o la herramienta de vigilancia del host, y `flow-build` vuelve a listar las tareas que faltan en la lista nativa al retomar o continuar un plan. `global.md` mide 46.036 de 46.040 bytes. Pasan las pruebas y el build del sitio. Publicado y desplegado con la release `e2507d48`.
- **Pausa (2026-10-08):** el usuario cerró la sesión. La siguiente continuará en la misma rama con otra premisa. Los commits de la rama son `75fe4ce5`, `4c75483b`, `3f7d15a6`, `722c7fe6`, `809acee3` y `8c69c3ef`, todos en `origin`. A `global.md` le quedan 4 bytes de presupuesto.
- **Siguiente paso (anterior):** el usuario lee el diff (`git diff 9c285026..4c75483b`). Si lo valida: push, PR a `development`, merge y `hive update`. Si pide cambios: se aplican en la rama, se repiten las comprobaciones y se vuelve a parar.
- **Cierre (2026-10-09, C1-A del usuario, que autorizó PR y merge):**
  - Se fusionó `origin/development` en la rama (`d4bbf0f7`), con el conflicto de `CHANGELOG.md` resuelto conservando ambas entradas. Las entradas de los cuatro commits de la segunda premisa se agregaron en `fb1c865e`.
  - Sobre `fb1c865e` pasaron `go vet ./...`, `go test ./...`, las pruebas de Python y `pnpm --dir site build`. gitleaks no encontró nada en los commits de la rama.
  - [PR #36](https://github.com/JhonHawk/tricell-hive/pull/36) fusionado (`3b872726`). #9 y #10 quedaron cerrados.
  - `hive update` desde `development` dio «unchanged»: la instalación ya coincidía. Los archivos de skills comparados son iguales al contenido de `development`.
