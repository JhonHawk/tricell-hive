# Fases

Base del cambio: `bb9604c5` (`development`, 2026-10-30). Rama `refactor/flow-ceremony`, en el checkout principal.

Cada fase de reconstrucción sigue el mismo método:

1. Leer la línea base de la skill en `baseline/` y listar sus reglas.
2. Clasificar cada regla con el filtro del inventario, buscando su procedencia (`git log -S`, el commit que nombra la sesión, los watch items).
3. Llevar esa tabla al usuario, que decide qué se queda, qué se acorta y qué se va.
4. Reescribir el `SKILL.md` y las referencias que sobreviven; borrar los cascarones que se descarten y corregir sus enlaces.
5. Actualizar la página de `site/` de ese flujo y correr las pruebas.

Una referencia compartida se rehace en la primera fase que la necesita y las siguientes la reutilizan.

### F0 — Preparación

- [x] Respaldo literal de `content/` y del `AGENTS.md` del repo en `_support/workspace/2026-10-30-skills-rebuild/baseline/` (ignorado por Git).
- [x] Inventario con uso medido, decisión por skill y filtro de reglas.
- [x] Fusión de `workspace-archive` en `workspace-conventions`, y de `flow-security` en `flow-verify/references/security-review.md`.
- [x] Seis skills de flujo reducidas a cascarón (frontmatter y encabezados).

### F1 — Flujo objetivo y decisiones transversales

- [x] Diseño de las skills: **híbrido estricto**, elegido por el usuario el 2026-10-30 tras investigar la evidencia externa (documentación de los hosts, estudios, caché) y la interna (lecturas en las sesiones desde el 2026-10-20). Queda así:
  - La guía global lleva solo los invariantes que toda sesión necesita, cada uno completo: secretos, autorización de efectos, el contrato mínimo de preguntas y reporte, y preservación. Desaparecen sus punteros «antes de X, lee Y».
  - Cada flujo lleva su camino principal completo en su `SKILL.md`, de ≤ ~4 500 tokens, con lo crítico al principio, porque Claude Code solo vuelve a adjuntar los primeros 5 000 tokens tras compactar.
  - Las referencias quedan solo para ramas raras, con un disparador explícito, a un solo nivel y sin cadenas.
  - El contrato de cada subagente vive en su rol; el hilo principal solo conserva el criterio de reparto.
  - `hive-core` se disuelve: lo universal va a la guía global, lo de un flujo a ese flujo, y `language` a los roles.
  - Se mantienen la regla de leer completo y el hook de Grok, y se añade «un archivo por llamada» por el truncado de Codex.
  - Cambia la sección «Guidance structure» de `AGENTS.md` (F7).
- [x] Diagrama del flujo mínimo, de la entrada al cierre (`design.md`).
- [x] Decidir las ceremonias transversales (D1–D5 en `design.md`, elegidas el 2026-10-30):
  - el verificador por tarea (siempre, por riesgo o nunca);
  - los registros de progreso y de delegación;
  - la refutación y la re-revisión del code review;
  - los niveles de la UI gate y el montaje de la base aparte;
  - las preguntas de entrega, revisión y UI (cuántas y cuándo).
- [ ] Ubicar los pasos baratos que faltaron: la región de acciones en el plan, la primera vista renderizada y el rol `hive-design-ui`.

**Closes:** base de AC1 y AC3. Sin código.

### F2 — research

- [x] Inventario de reglas: 141 reglas clasificadas con su procedencia (`_support/workspace/2026-10-30-skills-rebuild/f2-rule-inventory.md`, sin seguimiento).
- [x] Decisiones del usuario (2026-10-30):
  - Las reglas de categoría 4 (sin incidente detrás) se quitan del camino principal y de la guía global; solo se quedan en las referencias raras.
  - La revisión de un PR ajeno sigue D3: sin `hive-refute-claim` de rutina.
  - Se quita la línea de retrabajo del reporte de estado.
  - La guía global recibe ~1.3 KB netos de `reporting`, sin pasar de 19 KB.
- [x] **Cambio de enfoque (2026-10-30, decisión del usuario): skills en su mínima expresión, para evaluar de poco a más.**
  - Los seis flujos y `hive-core` quedan con su entrada, los pasos en orden y sus preguntas (~1.1–1.7 KB cada uno), sin referencias.
  - La guía global queda solo con los invariantes de seguridad: idioma, alcance y autorización, preservación y secretos (6.1 KB).
  - Una regla vuelve solo cuando una sesión real muestre que hace falta. El inventario F2 y `baseline/` son el catálogo de donde sale.
  - Este enfoque reemplaza el método de reconstrucción por fase de F2–F7.
  - Reporte: `_support/sessions/2026-10-30-skills-rebuild/skills-minimal.report.html`.
- [ ] Referencias compartidas que estrena: `flow-close/references/reporting.md` y `work-item-boundary.md` (research ya reporta y cierra turnos).
- [ ] Mantener los arreglos de #303: opciones por ticket en la ruta del backlog y listados acotados en el hilo principal.
- [ ] `site/.../flows/research.md`.

### F3 — plan

- [ ] `flow-plan` y sus 13 referencias, entre ellas `delivery-decisions`, `review-decisions`, `ui-verification-decision`, `ui-planning`, `plan-format`, `plan-review`, `change-records` e `interface-changes`.
- [ ] Decidir dónde se fija la región de acciones y la jerarquía de botones (AC3).
- [ ] `hive-review-plan`: cuántos revisores y cuándo.
- [ ] `site/.../flows/plan.md`.

### F4 — build

- [ ] `flow-build` y sus 9 referencias, entre ellas `plan-execution`, `verification`, `tdd`, `security-boundaries` y `browser-automation`.
- [ ] La primera vista renderizada dentro del build (AC3).
- [ ] El commit por tarea y el verificador por tarea, según lo decidido en F1.
- [ ] Salida de verificación silenciosa: un ✓ si pasa, la salida completa si falla, con fail-fast cuando la herramienta lo permita (HumanLayer, *Context-efficient backpressure*).
- [ ] Un checklist antes de dar por terminado el build, que verifique contra los criterios de la tarea (LangChain, *Improving deep agents with harness engineering*).
- [ ] `site/.../flows/build.md`.

### F5 — verify y walkthrough

- [ ] `flow-verify`: `review-and-gates`, `runtime-gates`, `ui-gate`, `ui-review-criteria`; integrar la referencia `security-review.md` ya movida.
- [ ] `flow-walkthrough` y `walkthrough-manifest`. Restituir el ejemplo JSON que exige `walkthrough_render_test`.
- [ ] Qué pasa con un hallazgo visual Medium introducido: hoy se puede recomendar aprobarlo (watch `watch/v1-recommends-known-medium`, si se registra).
- [ ] Primero los chequeos deterministas y después la revisión con modelo, solo donde D1, D3 y D4 la piden (Fowler, *Harness Engineering*). La revisión de UX califica el resultado, no el camino.
- [ ] Una revisión omitida deja anotada su razón, como `review_skipped_reason` (idea de un kit de orquestación listado en `reference-sources.local.md`).
- [ ] `site/.../flows/verify.md` y `security.md`.

### F6 — close

- [ ] `flow-close`: `branch-cleanup`, `incidental-findings`, `walkthrough-close`, y lo que quede de `reporting` tras F2.
- [ ] Evaluar un archivo de relevo al compactar o pausar, porque la compactación no deja la sesión en limpio (Anthropic, *Harness design for long-running apps*). Se adopta solo si en las sesiones baja la tasa de relecturas (hoy 34 % en Grok y 65 % en Codex).
- [ ] `site/.../flows/close.md`.

### F7 — Transversal

- [ ] `global.md`: dejar solo los invariantes universales y quitar los punteros «antes de X, lee Y»; reportar su tamaño (meta: < 200 líneas; Codex corta `AGENTS.md` en 32 KiB).
- [ ] Disolver `hive-core`: repartir sus 13 referencias entre la guía global, los flujos y los roles; corregir los enlaces de `git-workflow`, `harness-audit`, `flow-deploy`, `unattended-delegation` y `security-advisory`.
- [ ] `AGENTS.md` del repo: reescribir «Guidance structure» para el híbrido estricto, y ajustar `tests/content/skill_size_test.go` si el límite pasa a medirse en tokens.
- [ ] La guía global no cambia de bytes dentro de una sesión, porque el caché depende de un prefijo estable (Manus); ~150–200 instrucciones y < 300 líneas (HumanLayer, *Writing a good CLAUDE.md*).
- [ ] Una regla que se sigue fallando pese al énfasis pasa a un hook (documentación de Claude Code, *best practices*), según el criterio de medición de `AGENTS.md`.
- [ ] Cada rol lleva el brief de cuatro partes: objetivo, formato de salida, herramientas y límites (Anthropic, *multi-agent research system*).
- [ ] Roles de `content/agents/` que lanzan los flujos (`hive-verify-task`, `hive-verify-change`, `hive-review-*`, `hive-refute-claim`, `hive-read-state`).
- [ ] Alinear `flow-deploy` y `unattended-delegation` con el flujo nuevo.
- [ ] `README.md`, `llms.txt`, `site/.../skills.md`, `index.mdx`, `overview.md`; delta de `documentation-site` por `flows/security/`.
- [ ] Pruebas: `tests/content`, `tests/skills`, fixtures de `tests/fixtures/flows` y listas de `tests/pilot`.

### F8 — Medición e integración

- [ ] Desplegar la rama solo cuando el usuario lo indique.
- [ ] Segunda ronda vanilla y sesiones con la versión nueva en tickets comparables; anotar las filas en la rúbrica (AC5).
- [ ] Protocolo de medición tomado de un curso MIT listado en `reference-sources.local.md` (conservar el aviso de licencia si se copia texto):
  - variantes sin Hive, con Hive completo y con Hive ligero sobre el mismo commit;
  - modelo, esfuerzo y herramientas fijos, y el orden de las variantes rotado;
  - ≥ 3 pares de tareas antes de concluir;
  - un recibo por corrida: aceptado, evidencia, reintentos, correcciones humanas, tokens y veredicto (mantener, simplificar o quitar).
- [ ] Criterios de Anthropic (*Demystifying evals*, *infrastructure noise*): tareas sacadas de fallas reales, varias corridas por tarea, calificar el resultado y no el camino, anotar la configuración del entorno, y tratar como ruido una diferencia de menos de 3 puntos.
- [ ] Métricas de ceremonia: llamadas por turno, tokens con y sin caché, y la tasa de aprobación de cada gate (una gate que aprueba ~95 % solo pone sellos). Pruebas de remoción periódicas de reglas, bajo `decisions/rule-removal-criterion`.
- [ ] Contrastar los scripts propios con las herramientas de telemetría de sesiones listadas en `reference-sources.local.md`, pasando antes el chequeo de dependencias. Ninguna cubre los seis hosts por sí sola.
- [ ] PR a `development` con los lugares revisados, el tamaño de `global.md` y el resultado de la medición.

### F9 — Carga de skills por host (research del 2026-10-31)

Base: investigación en los 6 hosts (documentación, código de Codex, OpenCode y Pi, transcripciones locales y X). La guía es la única capa que sobrevive a la compactación en todos; el cuerpo de una skill es contexto de un turno; las referencias son resultados de herramienta que se pierden o se leen a medias; la autoinvocación falla seguido y el comando explícito garantiza la carga, no la obediencia.

- [x] **A. Paradas en prosa con el comando del host.** El bloque «▶ Siguiente paso» de «Questions and routes» trae el comando listo para escribir: `/flow-x` en Claude Code, Grok y Cursor; `$flow-x` en Codex; `/skill:flow-x` en Pi; `@flow-x` en OpenCode. La tarjeta de build desaparece: `plan-handoff.md` termina con `/flow-build`, o con el comando de compactación y luego `/flow-build` por encima del 40 %. Las decisiones dentro de una fase siguen en tarjeta. El README recomienda arrancar cada flujo con su comando.
- [x] **B. Autoridad de la guía global.** Abrir `global.md` con una línea que diga que sus reglas aplican en toda sesión y no son preferencias opcionales, porque Grok la encuadra como `user_rules` «follow if appropriate».
- [ ] **C. Tamaño y orden de las skills.** Skills de flujo por debajo de 8 KB (Codex recorta `$skill` a 8 KB) y lo obligatorio en las primeras ~60 líneas (Cursor lee 60–120). Las referencias solo para lecturas en un momento fijo.
- [x] **D. Instalación en Cursor.** La sesión de Cursor del 2026-10-31 (`97f259a4`) confirmó que el bloque de Hive en `~/.cursor/AGENTS.md` no se inyecta. Probar `~/.cursor/rules/hive.mdc` con `alwaysApply` u otra ruta que el CLI cargue; cambiar `integrations/cursor`, sus pruebas y `agent-delivery.md`.
- [x] **E. Documentación de arquitectura.** Corregir `compaction-survival.md` (Claude: la sonda no reinyectó la skill; Grok: nunca reinyecta skills; Codex: el fragmento explícito se descarta; OpenCode: se quitan las skills; Pi: resultados recortados a 2000 caracteres al resumir) y añadir a `agent-delivery.md` la tabla de carga de skills por host.
- [ ] **F. Skills sin revisar.** `git-workflow` (12.9 KB), `harness-audit` (10.9 KB), `unattended-delegation` (9.4 KB) y `flow-deploy` (8.1 KB) pasan de 8 KB; además `security-advisory`, `workspace-conventions`, `flow-report`, `starlight-docs-site`, `engram-init-workspace` y los 20 roles.
- [x] **G. La plantilla manda sobre el historial.** En `flow-plan`: las carpetas de cambio previas son contexto, no modelo; un solo `hive-review-plan` aunque el historial registre varios (`watch/plan-reviewers-copied-from-archive`, conteo 2).
- [ ] **H. Medición.** Una sesión por host con comandos explícitos; anotar carga, adherencia y relecturas en la rúbrica.

- [x] **I. Entrega en la guía global.** Los modos de entrega y la pregunta de entrega y revisión pasan de `flow-plan` a la sección `## Delivery` de `global.md`; el trabajo que no deja nada que commitear no lleva ninguna de las dos (Grok `01a1913d`, cam, 2026-10-31).
- [x] **J. Índices y respuesta «ok».** `D<n>` con su línea de título y una línea por opción, `A - (★) <descripción>` con la recomendada primero; `H<n>`, `T<n>` y `AC<n>`; una ronda con dos o más decisiones o hallazgos va en texto, «ok» aplica las ★ y «D2 B» cambia una.
- [x] **K. Un solo comando en la ruta.** El bloque «▶ Siguiente paso» escribe el comando una vez, en la sintaxis del host, con un argumento real; si el ticket no existe, recomienda crearlo primero (OpenCode `ses_e6ea0f80`, carvento, 2026-10-31).

- [x] **L. Tarjetas por ronda.** Una ronda de decisiones o hallazgos se escribe en texto y luego va en una sola llamada a la herramienta nativa (hasta su límite), con `A - (★)` en la etiqueta y la explicación en la descripción; previews para artefactos; entrevista por rondas en research y plan; línea de avance `✓ T3 · 3/5` en build y «Next» en el reporte final; respaldo en texto donde la herramienta falta o falla (research completo del 2026-10-31).

- [x] **M. Bloqueos con solución y validación autosuficiente.** Un bloqueo termina con la solución más chica y se pregunta como decisión cuando necesita autorización; la validación dice qué revisar y cómo y ofrece la prueba por el camino real; vuelve la regla de defectos que `flow-deploy` cita (Codex `01a1916d`, cam, 2026-10-31).
- [x] **N. Herramienta que espera, pregunta con contexto.** Se usa la herramienta de preguntas que espera la respuesta (la que regresa al instante solo como extra mientras sigue trabajo) y el turno termina con la pregunta; el título de cada decisión es la pregunta con lo verificado y qué revisar, cada opción dice su efecto; en decisiones de peso, razón, costo e incertidumbre, y solo lo verificado se afirma. Recuperado de la vieja `questions.md` (Codex `01a19339`, cam, 2026-10-31; `watch/codex-async-card-duplicate` con 2 sesiones; decisión D12, D14, D15).
- [x] **O. Revisión de UX de la pantalla completa.** `hive-review-ux` revisa lo que muestra el cambio y cada región que su layout afecta, no la pantalla entera (1280/390/320 px por tema, valores más largos y texto 30–50 % más largo), con comprobaciones medibles de integridad, y reporta todo defecto de esas regiones; `flow-verify` muestra los defectos previos de regiones afectadas en la ronda de validación, antes del push. Recuperado de la vieja `ui-review-criteria.md` (Codex `01a19339`, cam, 2026-10-31: el revisor aprobó con un nombre de 8 líneas y el móvil a dos columnas en su propia captura; decisión D18, D19). `AGENTS.md`: Codex 0.166 recarga el `AGENTS.md` global en sesiones abiertas (D17).
- [x] **P. Tarjeta que vuelve vacía.** Una línea que remite a la ronda ya escrita, sin repetir las decisiones, y la tarjeta sigue en las rondas siguientes; el respaldo en texto aplica a una herramienta que vuelve de inmediato. Codex `01a19339`, cam, 2026-10-31: en modo Default la tarjeta se resolvió vacía dos veces a los 120,0 s (60 s ocultos + 60 s de cuenta regresiva, según el código) y el agente repitió el bloque D15 (decisión D21).
- [x] **Q. Hook de Engram tras compactar.** «Persistent memory» nombra el caso: la orden de un hook de guardar el resumen compactado no es el cierre de la sesión; se recupera el contexto con lecturas. Codex `01a1679c` (2026-10-23) y `01a19339` (2026-10-31): el plugin Engram 0.1.5 inyecta «FIRST: Call mem_session_summary… MANDATORY» sin opción para desactivarlo, y el agente lo obedeció (decisión D23).
- [x] **R. Pregunta de revisión de código.** «Delivery» fija el título, los nombres de las opciones (`Native: <host> <command>` solo si el host tiene una, `Agent: hive-review-code`, `No dedicated review`; lo que ya corre en cada PR, como Bugbot, va al contexto), cuándo recomendar una revisión y cuál (la preferida del proyecto si corre en este host, si no la nativa, si no el agente). Reemplaza la vieja `review-decisions.md`, perdida en el corte; 84 preguntas reales mostraban nombres, orden y recomendación variables. Codex `01a19479`, hive, 2026-10-31: recomendó Claude `/code-review`, que Codex no corre (decisiones D27, D28).
- [x] **T. Commits ajenos en la entrega.** «Delivery» exige comparar la rama con su base antes de la validación o del push, y nombrar y preguntar por los commits de otro trabajo. Codex `01a19479`, hive, 2026-10-31: el PR #304 mergeó 57 commits de la reconstrucción de skills sin decirlo (decisión D32, 85186e94).
- [x] **U. `git-workflow` al mínimo.** De 13.2 KB a unos 9 KB en 10 pasos:
  - quita lo que repetía «Scope and authorization» y «Support folder»;
  - corrige 4 punteros rotos (flow-research step 2, dos secciones de flow-verify que ya no existen y la frase hacia flow-close);
  - suma el análisis de la situación: la rama comparada con su base, no tocar el checkout de otra sesión y preguntar con propuestas cuando el estado no resuelve una elección de Git.
  Quedan staging por archivo, `commit --only`, hooks, método de merge, base al día, ramas de entorno, cuerpo del PR y revisores humanos (decisión D36 y corrección del usuario del 2026-11-01).
- [x] **V. Skills que no se rehicieron.** Se rehacen como `git-workflow`: `unattended-delegation`, `flow-deploy` con sus referencias, `security-advisory`, `harness-audit` (sus pasos de aplicar pasan a `references/apply.md` y la memoria a `references/memory.md`), `workspace-conventions`, `engram-init-workspace`, `starlight-docs-site` y `flow-report`. Se corrigen unos 30 punteros que apuntaban a reglas que ya no existen y se quita lo que repetía la guía global. El modo desatendido deja de pedir el ajuste retirado «UI verification» y corre el revisor de UX. 8 KB queda como meta y no como tope (decisión D42 y aclaración del usuario).
- [x] **W. Reglas perdidas en el recorte, recuperadas en versión mínima donde se usan** (D43):
  - en la guía global, qué guarda la memoria de cada host y la excepción de idioma para valores de dominio;
  - en el modo desatendido, el contrato de ciclos;
  - en `flow-deploy`, la revisión del PR de promoción, incluidos los commits con `Unreviewed:`, y el paso a `Completion`;
  - en `urgent-fixes`, la mitigación de incidentes;
  - en `flow-build`, las condiciones para abortar un cambio de datos.
- [x] **X. Roles.**
  - Se retira `hive-refute-claim` (D44).
  - Todos los revisores usan P0–P3, definido en «Findings» de `flow-close` (D45).
  - Los hijos nunca hacen commit (D46).
  - La comprobación de producción pasa a `hive-read-state` (D47).
  - Se corrigen frontend (los datos de ejemplo del plan quedan exentos), plan (un revisor para todo el plan), UX (no vuelve a capturar), verificación (solo a pedido) y otros textos de referencias borradas.
- [x] **Y. Ideas del artículo de Microsoft sobre modernización de código con agentes (2026-10-26)** (decisión D52):
  - `flow-build`: ante una prueba que falla, decidir primero si está mal el código, el criterio o la prueba, y cambiar o quitar una prueba solo cuando es ella la que está mal;
  - `hive-review-code`: revisa si el diff es mínimo y coherente con las convenciones de los archivos que toca, con P2 para lo que la tarea no pidió;
  - `flow-close`: los P2 que comparten zona o causa se proponen como un solo grupo, una regla que se había perdido en el recorte.
- [x] **S. Retomar tras compactar.** «Persistent memory» pide leer el checkpoint del trabajo antes de responder o terminar el turno después de una compactación, y seguir desde su siguiente paso sin volver a reportar lo ya hecho; la regla de `flow-build` y `flow-verify` no sobrevive a la compactación porque el texto de la skill no queda en el historial. Codex `01a19479`, hive, 2026-10-31: la compactación de las 23:27 dejó como último mensaje un ajuste ya resuelto; el agente lo reportó otra vez y cerró el turno a mitad de la verificación (decisión D30; watch `watch/stop-after-midturn-compaction`).

Orden: A, B y G primero (contenido, sin manager); D con su sonda; E; F skill por skill; el `site/` al final, todo junto.

## Review status and progress

2026-10-30: F0 hecha en el checkout principal, sin commits. `go test ./...` pasa. `tests/skills` pasa 63 de 64: falla `walkthrough_render_test.test_reference_example_passes_the_validator` porque su ejemplo vivía en un cascarón; se resuelve en F5.
