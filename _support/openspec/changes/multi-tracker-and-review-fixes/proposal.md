# multi-tracker-and-review-fixes — Varios trackers por repositorio y tres correcciones de la revisión de sesiones

| Campo | Valor |
| --- | --- |
| Estado | Completado en la rama · sin integrar a development |
| Tracker | ninguno |
| Git | tricell-hive · rama `feat/gh-311-minimal-research-entry` sobre `origin/development` 048ad0c3 con `4a6b0ea6` (IDs al frente, rondas visibles, compactación) sin integrar |

## Decisiones

2026-11-04
- **D1 A** — ¿Dónde va el ID en un punto de lista? Al frente en negritas, con guion y descripción en palabras (ya en `4a6b0ea6`).
- **D2 A** — ¿Cómo se corrige el resumen tras compactar? Reforzando la cláusula aunque el hook lo marque crítico (ya en `4a6b0ea6`).
- **D3** — ¿Cómo se entrega? Commit sobre la rama actual y push; sin PR mientras el usuario no lo pida (D5 C).
- **D6 A** — ¿Se toca el recordatorio de guardado del plugin de memoria? No; se observa en la próxima revisión.
- Lista de ajustes: soportar más de un tracker por repositorio con uno principal, y corregir H11, H12 y H13.
- **D7 A** — ¿Cómo se declara más de un tracker? `Tracker` sigue siendo el principal; una clave opcional `Other trackers` declara los demás con sus prefijos de clave, y `Completion` acepta un valor por tracker con prefijo, validados por `hive doctor` y el editor.
- **D8 A** — ¿Qué revisión corre antes del push? Claude Code `/code-review`, porque el cambio toca lógica y la forma de `## Hive`.
- **D9 A** — ¿Qué hacen los flujos con los trackers que no son el principal? Leen y buscan en todos los declarados, mueven y comentan cada ticket en su tracker (espejos incluidos) y crean tickets nuevos solo en el principal; declarar un tracker en `## Hive` autoriza esos movimientos y comentarios.
- **D11 A** — ¿Dónde escribe `hive project set` la clave en una sección existente? Al final, como cualquier clave nueva; el orden de claves aplica a una sección nueva.
- **D10 A** — ¿Qué se hace con los identificadores de clientes en lo versionado? Marcadores en todo lo que se commitea, incluida esta carpeta; el detalle real queda en memoria privada.

## Solicitud

«añadamos a la lista de estos ajustes el tema de usar más de un tracker entonces, obvio uno es el principal», y planear H11, H12 y H13 junto con los trackers en un solo cambio.

## Objetivo

Hoy `## Hive` declara un solo tracker y un solo `Completion`. Un proyecto con Linear como tracker principal y el Jira de un cliente como segundo no mueve los tickets de Jira: llegaron a producción y el usuario los movió a mano; en otro proyecto, sin `## Hive`, tuvo que pedirle al agente que revisara Jira. Al terminar, un repositorio declara sus demás trackers y sus estados por prefijo de clave, y los flujos leen, buscan, mueven y comentan cada ticket en el tracker al que pertenece.

Además: Codex pregunta sin una tarjeta que espere (H11), los briefs de revisores llevan conclusiones del productor (H12) y `hive view work tasks register` falla sin decir qué campo (H13).

## Alcance y aceptación

Dentro: la clave `Other trackers`, `Completion` por tracker, su validación y edición en el manager, las reglas que leen, buscan, mueven, comentan y crean tickets, `harness-audit` HA-IF-17, la arquitectura, el sitio y el CHANGELOG; las correcciones de H11, H12 y H13. Fuera: configurar los proyectos afectados (después, con `/harness-audit` en cada uno), el PR a `development` y el despliegue.

- AC1 — `hive doctor --project` sobre un repositorio con `Other trackers: Jira · example.atlassian.net · keys XYZ` y `Completion: XYZ: QA on qa · Done on production` no reporta hallazgos en esas dos claves · falso hoy porque reporta `Other trackers: unknown key`.
- AC2 — `hive doctor` avisa, sin bloquear, cuando un valor de `Other trackers` no lleva `keys <PREFIJO>`, cuando un prefijo se repite, cuando un prefijo de `Completion` no está declarado, o cuando un valor de `Completion` tras `; ` no tiene la forma `<state> on <ref> · <state> on <ref>` · falso hoy porque esos chequeos no existen y el patrón actual se aplica a la línea entera, sin separar valores.
- AC3 — La vista Project del TUI muestra `Other trackers` en el orden de las claves con su forma como ayuda, y `hive project set` la escribe en el orden de las claves cuando la sección es nueva, y al final de una sección existente como cualquier clave nueva (D11 A); si el valor no cumple la forma, la muestra en el aviso · falso hoy porque la clave es desconocida.
- AC4 — La guía global define `Other trackers` y `Completion` por tracker en «Project settings», e indica que los flujos leen y buscan en todos los trackers declarados, mueven y comentan cada ticket en el tracker que declara su prefijo (el principal para claves sin prefijo declarado) y crean tickets nuevos solo en el principal · falso hoy porque el texto habla de un solo tracker.
- AC5 — HA-IF-17 indica reportar un tracker nombrado en las instrucciones del repositorio fuera de `## Hive` que no esté en `Tracker` ni en `Other trackers`, y una regla del proyecto que exija confirmar escrituras a un tracker declarado como decisión del usuario (mantenerla o quitarla), nunca como `stale` · falso hoy porque HA-IF-17 solo revisa las claves.
- AC6 — «Questions and routes» dice que una ronda hecha solo en texto, o solo con una herramienta que regresa al instante, es la falla, salvo las excepciones de su punto siguiente · falso hoy porque solo pide usar la herramienta que espera.
- AC7 — La línea canónica del «Reviewer brief» de `delegation.md` prohíbe pasar al revisor una lista de hechos o conclusiones «ya verificados» que no debe volver a comprobar, conciliada con «nothing settled» de la misma referencia · falso hoy porque no nombra ese patrón.
- AC8 — `hive view work tasks register` con una clave de tarea repetida responde `invalid_input` con la regla y la ruta, sin valores, por ejemplo `tasks[1].key: duplicate task key` · falso hoy porque responde solo «invalid semantic JSON».
- AC9 — `getting-started`, las páginas `deploy` y `close` del sitio, `deployment-manager.md` y el CHANGELOG describen `Other trackers` y `Completion` por tracker con marcadores · falso hoy porque no la mencionan.

## Prerrequisitos

Ninguno: el build no necesita nada fuera del repositorio.

## Contexto verificado

- `tooling/cli/doctor_project.go`: `hiveSettingKeys` (8 claves), `completionPattern` aplicado a la línea entera, clave desconocida como aviso. `project_write.go` y `tui_project_view.go` recorren `hiveSettingKeys`; `TestHiveSettingKeysAppearInTheGuidance` exige que cada clave aparezca en «Project settings».
- `tooling/cli/view_work.go:234-235`: todo error de decodificación o validación de `tasks register` se convierte en «invalid semantic JSON»; el decodificador compartido de `tooling/viewer` descarta el detalle. Observado en dos sesiones del 2026-11-04 (Codex y Claude Code).
- `content/skills/flow-close/SKILL.md` mide 15 804 bytes (límite 16 000).
- Revisión de sesiones del 2026-11-04: H11 en tres sesiones de Codex; H12 en dos de Claude Code y una de Grok.

## Entrega

Commit sobre `feat/gh-311-minimal-research-entry` y push (D3, D5 C); sin PR ni despliegue. Revisión de código: Claude Code `/code-review` sobre el candidato antes del push (D8 A). Antes del push se nombran los commits de la rama que no son de esta obra.

## Estado de la revisión y avance

2026-11-04 — plan escrito y revisado por un hijo `hive-review-plan`: 13 hallazgos; aplicados los de dependencia de T1 con la guía, gramática, motivo de AC1, rutas de H13 sin tocar el visor, lugares faltantes, orden de T1 y T2, HA-IF-17 como decisión del usuario y cumplimiento del commit; D9 y D10 resueltos por el usuario.
