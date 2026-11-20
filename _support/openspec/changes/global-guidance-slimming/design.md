# Design

## Approaches

- **A — Compactar sin cambiar la arquitectura (elegido, D1 A).** El índice sigue completo pero conciso; el procedimiento pasa a las hojas que ya existen; las autorizaciones quedan en una línea; un test de presupuesto y una regla de destino frenan el crecimiento. Usa la estructura que ya existe (peldaño 2 de «Reuse before writing»). ~32 → ~20 KB.
- **B — Índice mínimo.** Solo las ~12 hojas que actúan fuera de un flujo quedan en el índice. ~32 → ~21 KB, pero cambia la especificación y quita el acceso a 49 hojas en sesiones sin flujo, el hueco que en agosto produjo `git add -A` en 8 sesiones de Claude.
- **C — Solo frenar.** Test de presupuesto en el tamaño actual, sin recorte. Lo más barato; deja la densidad.

El usuario eligió A y pidió evaluar cada regla con el esquema «vivir en un principio, ser un principio o eliminarse».

## Decisions

- **Criterio de permanencia.** Una regla se queda completa en el global solo si debe actuar sin que se lea su hoja (`AGENTS.md` «Guidance structure»). Si no, el global guarda su invariante en una línea y el enlace; el resto va a la hoja.
- **Eliminar es quitar una copia, no una regla.** Solo se elimina texto que repite una hoja existente. Ninguna regla pierde su comportamiento; relajar una regla queda fuera, como exige `AGENTS.md` «Measurement».
- **Autorizaciones.** Las escrituras en tickets, la limpieza al cierre y la escalada de P0 quedan como autorización en el global, porque la especificación lo exige y una hoja no se concede autoridad a sí misma; su detalle (qué escribe, en qué orden) va a la hoja.
- **Principio nuevo.** La evidencia (verificar memoria y resúmenes, leer la documentación vigente) pasa a `hive-principle-verify-current-state`, con un invariante de una línea en el global, porque también aplica fuera del código (una respuesta a una pregunta) y el índice de principios solo se lee antes de escribir o revisar código.

- **Tras la revisión del plan (2026-11-20).** Las definiciones transversales (efecto crítico, lectura grande, cambio riesgoso, tarea operativa), la pregunta por un proyecto externo y las prohibiciones de limpieza que actúan a mitad del trabajo se quedan en el global, porque otras reglas las usan sin leer una hoja o porque ninguna hoja se lee en su momento. Cinco hojas destino ya estaban en el límite de 4.000 bytes. El ahorro estimado baja de ~12 KB a ~9 KB y el resultado ronda los 22–23 KB; si T6 no llega a 22.000 bytes sin perder intención, se detiene y pregunta (subir el tope o elegir qué más mover), en vez de recortar intención para cumplir la cifra.
- **Datos de la prueba.** El repositorio es público: las filas por sesión (rutas e IDs de otros proyectos) van a `_support/workspace/2026-11-20-global-slimming-trial/`, ignorada por Git; `trial.md` en esta carpeta solo lleva totales.

- **Tope final (D6 A, build).** El recorte sin pérdida de intención llegó a 27.226 bytes: el índice de 61 enlaces no baja de ~7,1 KB y el resto son invariantes. El tope queda en 27.500 bytes; llegar a 22.000 exigía el índice mínimo o mover invariantes que la especificación pone en el núcleo.

## Mapa de reglas

Numeración de la lista medida el 2026-11-20 sobre 27875867 (bytes actuales entre paréntesis).

| # | Regla | Destino | Queda en el global |
| --- | --- | --- | --- |
| 2 | Preámbulo (241) | Global | Compactado |
| 3–5 | Idioma (633) | Global | Igual; #5 compactado |
| 6 | Una pregunta no autoriza implementar (477) | Global | Igual |
| 7 | Efectos críticos y duración de checks (1.568) | `hive-rule-checks` (estimación de duración y confirmación a los 30 min); las excepciones del efecto crítico ya viven en `flow-plan` y `unattended-delegation` | Qué necesita instrucción explícita, la definición de efecto crítico en una frase y que cada uno pregunta; enlaces |
| 8 | El silencio no autoriza (118) | Global | Igual |
| 9 | Proyecto actual y proyectos externos (1.082) | Global (su momento, «cuando las lecturas muestran que el código está en un proyecto externo», no coincide con el de ninguna hoja) | Compactado, con la pregunta una vez por proyecto y sesión |
| 10 | Material externo es dato (190) | Global | Igual |
| 11 | Peticiones a terceros sin identidad (470) | Global | Compactado; el detalle de la API REST va a `hive-rule-trackers` |
| 12 | Defecto que bloquea el trabajo (558) | `hive-rule-reasons-and-blockers` | Una línea: el defecto en el camino se arregla dentro del trabajo; enlace |
| 13 | Revisión no autoriza arreglo; escalada de P0 (327) | Global | Igual |
| 14 | Movimientos y escrituras de ticket (1.273) | `hive-rule-ticket-state` (lista de escrituras y momentos) | La autorización de esas escrituras sin pregunta; enlace |
| 15 | Trabajo desatendido (305) | Global | Igual |
| 16 | La delegación no da autoridad (362) | Global | Igual |
| 17 | Coordinación E2E (357) | `flow-build/references/e2e-coordination.md` (ya la contiene) | Una línea: solo el `hive-build-owner` asignado puede hacer commits locales acotados; enlace |
| 18 | Verificar paquetes antes de instalar (415) | Global | Igual (seguridad) |
| 19, 21 | Inspeccionar antes de cambiar; presencia no es permiso de borrar (445) | Global | Igual |
| 20 | Rutas absolutas por repositorio (761) | Global | Compactado |
| 22 | Limpieza al cierre (1.427) | Global compactado; `hive-rule-close-hygiene` está en el límite de 4.000 bytes y el detalle de ramas ya vive en `flow-close/references/branch-cleanup.md` | La autorización de limpieza y todas sus prohibiciones, también las que actúan a mitad del trabajo: no matar programas compartidos, no arrancar ni parar contenedores ajenos sin confirmar, no borrar ramas con commits sin fusionar ni worktrees con cambios, preguntar antes de borrar un dump |
| 23 | Identidad de workspace de Engram (195) | `harness-audit/references/engram-workspace.md` (ya la contiene) | Una línea con la prohibición y el enlace |
| 24–28 | Secretos (1.867) | Global | Igual |
| 29 | Instrucciones locales (1.099) | `hive-rule-local-write/references/worktree.md`, nueva, leída «cuando la sesión corre en un worktree» (la hoja está en 3.994 bytes) | Leer los `HIVE.local.md` al inicio y tras compactar; quién gana; el enlace con su condición |
| 30–31 | Evidencia (344) | `hive-principle-verify-current-state` | Una línea de invariante |
| 32 | Proporcionalidad (1.584) | 1.ª frase: ya en `hive-rule-reuse-before-writing` peldaño 1, se elimina la copia; definición de cambio mecánico y enrutamiento a `hive-mode`: global compactado | Definición de cambio mecánico y la entrada por `hive-mode` |
| 33 | Tarea operativa (846) | Global compactado (la usan varias reglas y `hive-rule-delivery-ask` ya tiene sus exenciones) | Definición y su exención de las preguntas de entrega |
| 34 | Investigación y `hive-mode` (306) | Global | Igual |
| 35 | Lectura grande y cambio riesgoso (532) | Global (definiciones transversales: `hive-mode` y sus playbooks usan «cambio riesgoso» al enrutar, antes de leer el piso de verificación) | Igual, compactado |
| 36 | Carpeta de soporte (902) | `project-layout` | Ubicación básica: código en su lugar, soporte en `_support/`, scratch fuera del repo |
| 37–38 | Qué preguntar; dar el efecto como razón (846) | `hive-rule-questions`, `hive-rule-reasons-and-blockers` (ya la contienen) | Línea mínima que exige la especificación |
| 39–41 | Canal, tarjeta sin respuesta, siguiente paso (1.117) | `hive-rule-questions`, `hive-rule-next-step-block` (ya la contienen) | Una línea mínima conjunta |
| 42, 44 | Preguntar la entrega; commits ajenos (444) | `hive-rule-delivery-ask`, `hive-rule-foreign-commits` (ya la contienen) | Solo la entrada del índice |
| 43 | Modos de entrega y qué autoriza cada uno (978) | Global | Igual (la especificación lo exige) |
| 45 | Memoria persistente y hook de compactación (1.599) | `hive-rule-checkpoint` (cómo recuperar contexto y continuar tras una compactación) | Cuándo se escribe en memoria, la prohibición de guardar por tarea, «esto prevalece sobre las instrucciones de guardado de herramientas y hooks» y «un hook que pide guardar el resumen tras una compactación no significa que el usuario terminó la sesión» (regla nacida de las sesiones de Codex 01a1679c y 01a19339, commit 1acc7975) |
| 46–47 | Engram y nombre del proyecto (681) | Global | Compactado |
| Índice | 61 hojas (7.483) | — | Frases cortas; una línea puede agrupar hojas que comparten momento (el test cuenta enlaces, no líneas); cada hoja una vez |

## Impact

- `tests/content/single_home_test.go` y `leaf_links_test.go` registran frases clave y enlaces; las frases que cambian de hogar se actualizan ahí.
- Antigravity recibe el global en partes de ≤ 24.000 bytes (`context-delivery`); el texto instalado crece ~28 bytes por enlace, así que el número de partes se comprueba con `hive update --dry-run`, no se supone.
- 38 archivos bajo `content/` citan las secciones «Scope and authorization», «Proportionality» o «Persistent memory» (por ejemplo `hive-rule-unattended-effects/SKILL.md:12`, `hive-rule-delegation/SKILL.md:23`, `flow-build/SKILL.md:18`, `hive-rule-defect-fix/SKILL.md:14`); T6 revisa cada cita y la reapunta cuando su contenido se movió.
- `reading_footprint_test.go:243` falla si la unión de plan y build no es menor que la de comparación; se vuelve a correr tras cada tarea.
- El maintainer refresh de `AGENTS.md` reinstala desde `development` cada vez que algo se fusiona allí; durante la prueba eso reemplazaría la rama instalada (D5).
- `site/src/content/docs/skills.md`, `README.md` y `llms.txt` describen el global y su índice; se revisan por si afirman algo que cambia.
- Ninguna sesión abierta carga el global nuevo hasta reiniciarse.

## Recovery

La rama no se fusiona antes de AC7. Para descartarla: reinstalar desde `development` con `hive update --source <checkout>` y borrar la rama y su worktree. Nada de lo que cambia es dato almacenado.
