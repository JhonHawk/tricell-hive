# Vacíos de guía detectados en las sesiones del 2026-10-19

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en [PR #96, ajustes de validación y entrega](https://github.com/JhonHawk/tricell-hive/pull/96), commit `ac146beb`, el 2026-10-23 |
| Tracker · GitHub Issues | [#93 — avisos de procesos propios, procesos de hijos, runtime real y candidato fijo](https://github.com/JhonHawk/tricell-hive/issues/93) |
| Git | Integrado en `development` con merge commit; historial por tarea conservado; cierre de registros autorizado mediante PR posterior por D2 |
| Verificación | Búsqueda de impacto · suite local · lectura de coherencia |
| Siguiente paso | Sin tareas de implementación pendientes; conservar la observación y los límites de validación registrados |

## Solicitud

> Analiza las sesiones de hoy con grok en cam y de codex en tricell-hive (solo la ultima)
> - Grok, tuvo cierres muy raros, practicamente a medias
> - codex, fue sumamente lento en la implementación del cambio de front

Tras el análisis y la pregunta D1 (qué hacer con los hallazgos), el usuario respondió:

> Resolvamos de una vez los que pueden ser gaps en en nuestras reglas o skills y dejemos en observacion los demas

Esa respuesta adopta las reglas de los vacíos con evidencia de una sola sesión, como permite la sección Measurement de `AGENTS.md`. Los incumplimientos de reglas existentes quedaron como casos en observación en Engram (`watch/2026-10-19-session-analysis-watch-items`).

Durante el plan, el usuario añadió una regla nueva (G5):

> de hecho una regla nueva que va de mano con el visor, esa etapa no debe ser directamente "lo apruebas?" si no, si desea que se abra, que se apruebe directamente o pedir correcciones (por cierto, ojo con esta ultima, es una opción que genera un turno sin poder añadir detalles hast ael sigueinte)

## Objetivo

Dos sesiones reales mostraron cuatro situaciones que la guía no cubre o deja ambiguas. En Grok, al cerrar, el agente detuvo sus propios servidores y cada aviso de fin de proceso abrió un turno visible después del cierre (G1). Además, el primer cierre dejó encendidos los servidores que había iniciado un hijo verificador, porque la regla habla de "the processes it started" (G2). En Codex, las pruebas del implementador simulaban `fetch` y un Blocker que dejaba el visor sin funcionar apareció recién en el gate final de UI (G3). Ese gate, además, verificó contra un candidato que cambiaba mientras corría (G4).

El cambio ajusta cada regla en su lugar canónico para que la próxima sesión cierre sin ruido ni procesos huérfanos, y para que los defectos de runtime aparezcan antes del gate, que debe verificar un candidato identificado.

Durante el build, el usuario observó (G6):

> Algo que note en esta sesión y que evitaria que el usuario se quede a "ciegas" de lo que se esta haciendo es una doble confirmación, por ejemplo, me presentaste los hallazgos, yo te dije que ataquemos los son gaps y los demas se queden en observación, pero ya nunca me dijiste oficialmente cuales ibas a resolvar y cuales solo a observar, ese doble check con un "estas de acuerdo?" o corregir el camino, añadiria valor antes de que sea tarde

## Alcance

Incluye las cuatro reglas, sus punteros y los textos que dependen de ellas (guía global, skills, roles, `site/`, `_support/docs/architecture/`, detectores de `tests/pilot/`), según el mapa de impacto de `design.md`.

Excluye:
- Los casos en observación W1–W7 (progreso sin llamada a herramienta, walkthrough sin confirmar, turno sin siguiente paso, riesgo como recordatorio, ronda de corrección con contexto inflado, hilo principal sobrecargado, reintentos de pnpm). Son incumplimientos de reglas existentes o costos sin regla, con conteo 1.
- Cambios de comportamiento específicos de un host.
- Pilotos de comportamiento, pausados por instrucción del usuario.

Restricciones:
- Una sola casa por regla.
- La suite local sigue pasando.
- Los cambios en `content/guidance/global.md` se miden en bytes en el pull request.

## Criterios de aceptación

- AC1. La guía indica que el aviso de fin de un proceso que el propio agente detuvo y ya informó no produce mensaje para el usuario. **Falso en la referencia:** `reporting.md`, "Report a running or completed item", paso 4, exige responder cada aviso en una línea.
- AC2. La regla de limpieza de flow-close incluye entre los procesos de la tarea los que iniciaron sus hijos. **Falso en la referencia:** `flow-close/SKILL.md`, paso 4, dice "stop only the processes it started", sin mencionar hijos.
- AC3. La inspección del implementador de UI exige al menos un chequeo acotado que cargue la vista cambiada en un navegador real contra la aplicación en ejecución (servidor de desarrollo o build), sin dobles de prueba en las rutas que usa el cambio. **Falso en la referencia:** `ui-gate.md`, "The implementer's inspection", solo enumera `read`, consola o `eval`, sin exigir el runtime real.
- AC4. Cada hijo del gate de UI recibe en su brief la identidad del candidato que verifica, y la guía dice qué pasa con un hijo en curso cuando una corrección cambia el candidato. **Falso en la referencia:** `ui-gate.md` no nombra la identidad del candidato ni ese caso.
- AC6. Cuando una respuesta del usuario deja una selección al criterio del agente, por ejemplo qué hallazgos se resuelven y cuáles quedan en observación, la guía exige mostrar la selección (qué entra y qué no) y pedir confirmación o corrección antes de actuar. **Falso en la referencia:** `questions.md` ("an answer arrives") solo pide decir en una línea qué resuelve la respuesta.
- AC7. Cuando un hijo implementa un cambio de UI, la guía exige que su brief le asigne el navegador para el chequeo de runtime real de AC3. **Falso en la referencia:** `ui-gate.md` no lo pide, y `browser-automation.md:8` limita a un agente delegado a la herramienta que asigna su brief.
- AC5. En la primera vista y en la validación final de un recorrido, la pregunta ofrece abrirlo (el agente ejecuta `hive view open` y vuelve a preguntar) o aprobarlo, e indica que las correcciones se escriben en el campo libre de la tarjeta. Si el host no tiene campo libre, la pregunta va en texto. **Falso en la referencia:** `flow-walkthrough/SKILL.md` prohíbe abrir el visor ("Never open the page, the viewer, or a browser") y no define esas opciones.

## Decisiones

- D1, qué hacer con los hallazgos, 2026-10-19: corregir los vacíos de reglas y dejar el resto en observación (texto libre del usuario, citado arriba).
- D2-A, cómo se piden las correcciones, 2026-10-19: en el campo libre de la tarjeta, sin opción propia. Si el host no tiene campo libre, la pregunta va en texto.
- D4-A, entrega, 2026-10-19: rama de prueba `feat/gh-93-guidance-session-gaps` desde `development`, sin commits hasta que el usuario lo pida. El despliegue local desde la rama solo ocurre si el usuario lo pide; después, el usuario decide si hace merge o pide ajustes. El registro de cambio sigue el cierre que el usuario elija entonces.
- D5-A, revisión, 2026-10-19: sin revisión de código dedicada.
- D6-A, ruta, 2026-10-19: empezar `flow-build`.
- D7-A, separación de hallazgos, 2026-10-19: se resuelven G1–G5 y quedan en observación W1–W8 (`watch/2026-10-19-session-analysis-watch-items`, más W8, la grabación bajo CSP).
- D8-A, regla nueva, 2026-10-19: G6 entra en #93 como AC6 y T7, adoptada por decisión del usuario.
- D9-A, cierre, 2026-10-19: un commit por tarea en la rama de prueba, sin push, y despliegue local desde la rama para probar en sesiones reales.
- D10-A, hallazgo H1, 2026-10-19: agregar a #93 que el brief del implementador de UI delegado asigne el navegador (AC7, T8).
- D3-A, en qué paradas aplica, 2026-10-19: en la primera vista de una composición nueva y en la validación final del candidato.

## Continuación del 2026-10-23

- El usuario pidió auditar las sesiones de CAM, Facti e Hive desde el 20 de octubre y autorizó commits y push en la rama actual y merge a `development`. La investigación se entregó en la conversación; no se creó un documento de research.
- D1-A (ajustes posteriores): se aprueban P1, revalidar la evidencia afectada por cambios posteriores a aprobación y pedir aprobación nueva si cambia lo visible; P2, comunicar el límite de la espera; P3, salir de la rama y retirar el worktree limpio antes de borrarla. Los incumplimientos restantes quedan en observación, sin hooks ni reglas globales nuevas.
- D2 (cierre): se autoriza un PR pequeño posterior, sin revisión dedicada, con archivo de los dos planes y merge a `development` por el agente.
- Se conserva la decisión de no ejecutar revisión dedicada y la pausa de pilotos. La auditoría distingue resultados observados de mejora general no demostrada; la sesión de CAM activa al corte no tiene cierre evaluado.

## Cierre del 2026-10-23

- Integración confirmada en [PR #96, ajustes de validación y entrega](https://github.com/JhonHawk/tricell-hive/pull/96), `ac146beba43fe6c0de6409a9a5c8e6313d3b96fb`. Los issues cubiertos se cerraron y no tienen sub-issues. No hay CI configurado; las comprobaciones finales locales e independientes pasaron.
- Contenido local actualizado desde el commit integrado en los seis hosts registrados; estado de instalación comprobado: 562 entradas instaladas. No cambió el binario ni se reinició el visor cotidiano.
- D2 autoriza publicar y fusionar este archivo de cierre mediante un PR pequeño posterior contra `development`, sin revisión dedicada. No incluye promoción a `master`, release ni publicación del sitio.
