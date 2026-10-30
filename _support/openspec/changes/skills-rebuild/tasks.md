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

- [ ] `flow-research` y sus referencias: `backlog-report`, `delegation`, `diagnosis`, `change-impact`, `pr-review`, `ticket-body`.
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

## Review status and progress

2026-10-30: F0 hecha en el checkout principal, sin commits. `go test ./...` pasa. `tests/skills` pasa 63 de 64: falla `walkthrough_render_test.test_reference_example_passes_the_validator` porque su ejemplo vivía en un cascarón; se resuelve en F5.
