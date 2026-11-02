# Protocolo único de cambio de alcance y diagrama de flujos

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el [PR #235](https://github.com/JhonHawk/tricell-hive/pull/235) (`36f1da1b`) el 2026-10-27 |
| Tracker · GitHub Issues | Sin issue: el trabajo nace de una investigación en la sesión (2026-10-27) |
| Git | `interactive` · rama de trabajo desde `development` · alto antes del push para tu validación |
| Verificación | `go test ./tests/content` · `go test ./...` · build del sitio · `/code-review` · tu validación |
| Siguiente paso | Ninguno en este cambio; la pregunta de revisión de código (D12-A) sigue como cambio aparte |

## Solicitud

> crea una rama nueva, commit y push y despues, research profundo [imagen del diagrama de flujos del README con un panel lateral «reglas generales»] este es nuestro flujo actual, cuales son reglas generales, cuales solo son procesos que corresponden a un flujo, donde se itera? es importante eso, donde se pueden solicitar cambios, y ampliar alcances sin romper el flujo o poder regresar a él

Respuestas que fijan el alcance (2026-10-27): D1-A «Planear protocolo y diagrama», D2-A «No guardar» la investigación, D3-A «Referencia nueva», D4-A «Por intención», D6-B «Proponer volver al plan», D8-B «Más quitar duplicaciones».

> y de hecho tambien hay que ajustar este tema del code-review, la formación de la pregunta no es estatica, si algo no necesita un code-review, eso se recomienda

Este segundo mensaje abre un tema relacionado; su destino se decide en D12 (ver «Temas abiertos»).

## Objetivo

Los flujos de Hive tienen unos 20 bucles con topes, pero ninguna regla única decide qué pasa cuando el usuario pide un cambio a mitad del trabajo: si es una corrección, una tarea añadida o un cambio nuevo, y a qué punto del flujo vuelve el trabajo. Hoy esa decisión está repartida entre `plan-revision.md` (solo con plan guardado), la regla de temas abiertos de la guía global y frases sueltas en verify; sin plan, durante deploy o después de integrar no hay ruta. Este cambio crea una regla canónica de clasificación, hace que cada flujo apunte a ella en su punto de reentrada, quita una duplicación de la guía global y redibuja el diagrama del README con el panel de reglas generales y los bucles reales, más una página de panorama en el sitio.

## Alcance y aceptación

**Incluido:** referencia nueva `content/skills/flow-plan/references/scope-changes.md`; punteros desde `global.md` (temas abiertos), `plan-revision.md`, `flow-plan`, `flow-build`, `flow-verify`, `flow-walkthrough`, `flow-close` y `flow-deploy`; reducción de `global.md:36` a su invariante mínima; diagrama del README; página `flows/overview` del sitio y las páginas de flujo que describen estos puntos; entrada en `_support/docs/architecture/instruction-resources.md`.

**Excluido:**
- Mover las 5 referencias generales (`questions`, `reporting`, `delegation`, `effect-authorization`, `project-settings`) a una ubicación neutral: unas 100 referencias y pruebas; queda como tema abierto con propuesta de ticket.
- `global.md:72` (aceptar sin verificar un control que el usuario pidió saltar): al releerla con la regla de AGENTS.md, ya es la invariante mínima que debe actuar sin `flow-verify` cargado; `flow-verify/SKILL.md:33` es su versión completa. Se deja igual (ajuste de D8-B, ver diseño).
- Mermaid en el sitio: el sitio no lo soporta y añadirlo sería una dependencia nueva; el sitio usa una tabla.
- Pilotos con modelos (pausados).

**Restricciones:** `go test ./...` y `go vet ./...` pasan; cada `SKILL.md` y referencia ≤ 16 000 bytes (`flow-build` tiene 532 bytes libres y `flow-close` 428); ninguna página del sitio copia texto de reglas; la guía global no crece en más de 300 bytes netos.

**Búsqueda en especificaciones:** se buscó en `_support/openspec/specs/` (context-delivery, documentation-site, versioned-installation, walkthrough-viewer) por global, flow-plan, README, llms, mermaid, reference, walkthrough, unattended, skip, correction y scope, y se leyeron todos sus `SHALL NOT`. Afecta a `documentation-site` «Published documentation site» (enumera las páginas del sitio) y «Flow pages describe without copying rules»; delta en `specs/documentation-site/spec.md`. `context-delivery` «Portable kernel and situational procedures» es una restricción que se cumple (las invariantes de autorización quedan en la guía global con condición de lectura explícita); sin delta. Nada más coincide.

- AC1. Existe `content/skills/flow-plan/references/scope-changes.md` con `## When to read` y las cuatro clases (corrección, tarea añadida, cambio nuevo, tema aparte con destino), decididas por la intención del trabajo y con ejemplos. *Falso en la base cuando* el archivo no existe.
- AC2. Un trabajo sin plan guardado pasa a `flow-plan` cuando el cambio añade otro ticket, otro repositorio, un contrato o datos guardados; `flow-build` apunta a la referencia donde decide que un cambio pequeño no necesita plan. *Falso en la base cuando* `rg scope-changes content/skills/flow-build/SKILL.md` no devuelve nada.
- AC3. A la tercera ronda de corrección sobre el mismo aspecto (en la pausa de validación o en la primera vista), el agente propone volver al plan como pregunta y el usuario decide. *Falso en la base cuando* ninguna regla de `content/` menciona una tercera ronda de corrección.
- AC4. La regla de temas abiertos de `global.md` conserva sus tres etiquetas, cada una ligada a clases que no se solapan (aclara = corrección; relacionado = tarea añadida o cambio nuevo; aparte = tema aparte), y remite a la referencia antes de actuar sobre un tema que cambia el trabajo en curso. *Falso en la base cuando* `global.md` no enlaza `scope-changes.md`.
- AC5. `plan-revision.md` deja de definir por su cuenta corrección, tarea añadida y cambio nuevo, y apunta a la referencia. *Falso en la base cuando* `plan-revision.md:11-12` contiene esas definiciones.
- AC6. En `flow-verify`, cada pedido durante la pausa de validación se clasifica con la referencia antes de su ronda, y una ronda de corrección sin efecto visual muestra el diff en vez de capturas. *Falso en la base cuando* `flow-verify/SKILL.md` no enlaza la referencia ni menciona el diff en las rondas.
- AC7. En `flow-walkthrough`, rechazar la primera vista abre rondas de corrección dentro del build que cuentan para la propuesta de AC3. *Falso en la base cuando* la sección «First rendered view» no dice qué pasa tras un rechazo.
- AC8. En `flow-close`, un requisito nuevo después de integrar y antes de cerrar se clasifica con la referencia; si es un cambio nuevo, el registro integrado cierra normalmente y lo nuevo abre otro cambio. *Falso en la base cuando* `flow-close/SKILL.md` no enlaza la referencia.
- AC9. En `flow-deploy`, pedir añadir o quitar tickets durante una promoción se clasifica con la referencia: un ticket integrado después del inicio va a la corrida siguiente; para quitar uno que ya está en la rama base se pregunta si se mantiene (la corrida sigue) o se revierte como cambio propio (la corrida para en un punto seguro y la siguiente empieza desde el primer paso). Revisado el 2026-10-27 tras la revisión de código: la promoción lleva toda la rama base, así que reiniciar no puede excluir un ticket integrado. *Falso en la base cuando* `flow-deploy/SKILL.md` no trata cambios del conjunto de tickets.
- AC10. `global.md:36` conserva solo las fuentes que activan la delegación desatendida (incluido el mensaje de que el usuario se va), su vencimiento (al completarse, o cuando el usuario cambia el objetivo, el alcance o una respuesta en vigor, pide otro trabajo o dice que pare), que nunca se activa por la duración de la tarea ni por silencio antes de las respuestas, que no permite acciones irreversibles ni amplía permisos, y el puntero a la skill; la lista de mensajes que la mantienen y la regla de «encaja en ambas listas» viven solo en `unattended-delegation/SKILL.md`. *Falso en la base cuando* `global.md:36` contiene «Only these keep it».
- AC11. El diagrama del README muestra un panel de reglas generales, las vueltas verify→build, build→plan, close→build/plan y deploy→build, y un nodo de «cambio pedido» que remite a la referencia. *Falso en la base cuando* el bloque mermaid del README no tiene esos nodos ni flechas.
- AC12. El sitio tiene la página `flows/overview` en el menú lateral, con una tabla de dónde se piden cambios en cada flujo y a qué punto vuelve el trabajo, sin copiar texto de reglas, y el build del sitio pasa. *Falso en la base cuando* `site/src/content/docs/flows/overview.md` no existe.
- AC13. `_support/docs/architecture/instruction-resources.md` lista `scope-changes.md` como hogar canónico con su condición de lectura. *Falso en la base cuando* no la menciona.

## Temas abiertos

- **Pregunta de revisión de código (tu mensaje del 2026-10-27):** destino pendiente de D12. `review-decisions.md:17` ya pide recomendar «Sin revisión dedicada» cuando el cambio no toca contratos, seguridad, permisos, datos guardados ni concurrencia y no cambia más comportamiento del que sus pruebas muestran; la causa de que en la práctica casi siempre se recomiende una revisión no está confirmada.
- **Mover las 5 referencias generales a una ubicación neutral:** fuera de este cambio; propuesta de ticket al cierre.
- **H2** (casos de cierre frente a `change-records.md`, re-verificación frente a `hive-verify-task`): en observación; se reabre si una sesión real muestra una contradicción.

## Entrega

- Repositorio `tricell-hive`, base `development`, modo `interactive` (D9-A, 2026-10-27): rama de trabajo, commits locales por tarea, pruebas, `/code-review` de Claude Code sobre el candidato local (D10-A) y alto antes del push para tu validación; después push, PR y merge en `development` sin CI (el repositorio la retiró): la condición de merge es la suite local completa sobre la base actual.
- El registro de cambio se versiona una vez, al cierre, en un PR pequeño sin revisión dedicada que integro yo.
- Tras cada merge en `development`, `hive update` refresca la instalación local (checkout del mantenedor, AGENTS.md).
- Sin promoción a `master` ni release.
