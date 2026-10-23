# Relecturas tras compactar y ritmo de escritura en memoria

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en [PR #96, ajustes de validación y entrega](https://github.com/JhonHawk/tricell-hive/pull/96), commit `ac146beb`, el 2026-10-23 |
| Tracker · GitHub Issues | [#94 — relecturas completas tras compactar y escritura en memoria sin retorno](https://github.com/JhonHawk/tricell-hive/issues/94) |
| Git | Integrado en `development` con merge commit; historial por tarea conservado; cierre de registros autorizado mediante PR posterior por D2 |
| Verificación | Búsqueda y lectura · pruebas de `tests/content` · tamaño de `global.md` |
| Siguiente paso | Sin tareas de implementación pendientes; conservar la observación y los límites de validación registrados |

## Solicitud

> se abrieron dos sesiones de grok en cam, las dos ultimas revisalas, siento que han tardado demasiado y se han equivocado

> Tambien analiza si algo de la metodología lo hacer ser lento, no solo si se equivoca, si no que haga algo qu eno haya tenido valor y solo haya inflado el tiempo

Tras el análisis, el usuario eligió D18-A: planear P1 y P2, dejar en observación los casos de una sola sesión y llevar P3 a una sesión de cam. También eligió D19-A, abrir un issue, y D20-A: Hive fija el ritmo de guardado y prevalece sobre las instrucciones del plugin.

## Objetivo

Dos sesiones de Grok en cam del 2026-10-20 (`01a1579f` y `01a157db`) perdieron minutos en dos pasos sin valor.

- **P1:** después de cada compactación, el agente releyó completas las referencias de la etapa. Las relecturas fueron entre el 42 y el 48 % de todo lo que leyó de la guía, y ese peso adelantó la compactación siguiente.
- **P2:** el agente escribió en Engram después de cada decisión o tarea y resolvió un juicio de conflicto por cada guardado. En una de las sesiones nunca leyó Engram. Ese ritmo lo impone el plugin de Engram; la guía de Hive solo dice dónde guardar.

El cambio fija en la guía global qué releer después de compactar y en qué momentos se escribe en memoria, para recortar ese tiempo sin perder lo que la memoria aporta entre sesiones.

## Alcance

Incluye dos reglas en `content/guidance/global.md` y el ajuste de `flow-close` (pasos 3 y 5) para que sus escrituras en memoria vayan en una sola pasada al cierre.

Excluye:
- los casos W16–W22, que siguen en observación (`watch/2026-10-20-grok-cam-method-cost`);
- P3, declarar Delivery, Review y UI verification en cam, que corresponde a una sesión de cam;
- cambiar la configuración del plugin de Engram;
- el hook de Grok `hive-read-whole-references`.

Restricciones: cada regla tiene una sola casa, la suite de contenido sigue pasando y el cambio de tamaño de `global.md` se informa.

## Criterios de aceptación

- AC1. La guía global indica que, después de que el host compacta o resume el contexto, el agente relee solo la referencia cuya condición aplica a su siguiente acción, no todas las de la etapa, y que eso aplica al hilo que compactó, no a sus hijos. **Falso en la referencia:** `content/guidance/global.md` no menciona la compactación (`rg -n -i "compact" content/guidance/global.md` no devuelve nada).
- AC2. La guía global fija los momentos de escritura en memoria persistente: una decisión del usuario, el plan listo (estado de trabajo, sin copiar el plan), el cierre de un work item y, cuando el usuario termina la sesión y la memoria admite resúmenes, el resumen de sesión. Prohíbe escribir por tarea, hallazgo, arreglo o descubrimiento entre esos momentos y declara que ese ritmo prevalece sobre las instrucciones de guardado de las herramientas de memoria. **Falso en la referencia:** `global.md:91` solo fija el destino.
- AC3. `flow-close` registra en una sola pasada de memoria, antes del reporte, los worktrees retirados y los conteos de recurrencia de los hallazgos. Lo que el usuario decline o suprima en la pregunta de cierre se guarda después de su respuesta, como escritura de decisión del usuario. **Falso en la referencia:** el paso 3 dice "Each time you remove a worktree … update its record".

## Decisiones

- D18-A, ruta, 2026-10-20: planear P1 y P2; el resto queda en observación y P3 va a cam.
- D19-A, issue, 2026-10-20: abrir #94.
- D22-A, cierre, 2026-10-20: un commit por tarea y despliegue local desde la rama.
- D21-A, ruta, 2026-10-20: empezar `flow-build`. Se reusan D4-A (rama de prueba sin commits hasta pedirlo) y D5-A (sin revisión dedicada).
- D20-A, P2 frente al plugin, 2026-10-20: la regla de Hive fija el ritmo y prevalece sobre las instrucciones de guardado de las herramientas de memoria. Riesgo aceptado: el plugin puede ganar igual, porque ningún host declara la precedencia entre canales.

## Continuación del 2026-10-23

La auditoría observó una relectura focalizada después de compactación en Claude Hive, compatible con el cambio, y guardados de avances fuera de la cadencia en dos sesiones independientes, CAM y Facti. No se presenta cumplimiento universal ni mejora de rendimiento. El riesgo de precedencia del plugin sigue reconocido. El usuario autorizó entrega a `development` y, en D2, un PR posterior de cierre para archivar ambos planes; se conserva la decisión de no hacer revisión dedicada.

## Cierre del 2026-10-23

- Integración confirmada en [PR #96, ajustes de validación y entrega](https://github.com/JhonHawk/tricell-hive/pull/96), `ac146beba43fe6c0de6409a9a5c8e6313d3b96fb`. Los issues cubiertos se cerraron y no tienen sub-issues. No hay CI configurado; las comprobaciones finales locales e independientes pasaron.
- Contenido local actualizado desde el commit integrado en los seis hosts registrados; estado de instalación comprobado: 562 entradas instaladas. No cambió el binario ni se reinició el visor cotidiano.
- D2 autoriza publicar y fusionar este archivo de cierre mediante un PR pequeño posterior contra `development`, sin revisión dedicada. No incluye promoción a `master`, release ni publicación del sitio.
