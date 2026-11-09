# Consumo del visor de Hive por sesión

Fecha de inicio: 2026-11-02 · Estado: línea base y medición posterior a #328 hechas; falta la nueva medición con 3 sesiones completas (ver «Siguiente paso»). Mediciones de Codex del 2026-11-03 y el script `measure-delegation.py` en `../2026-11-03-delegation-cost/`.

Los proyectos distintos de Hive aparecen como «proyecto B» y «proyecto C», y las sesiones van sin su ID.

## Pregunta y decisión

¿Cuánto del consumo de tokens se va en usar el visor? Sirve para saber si #328 (contrato por flujo y lectura única de la referencia; PR #329, #331) bajó el costo. Decisión D3 A: guardar el método y volver a medir cuando haya unas 3 sesiones completas que publiquen con las reglas nuevas.

## Método

`measure-viewer-cost.py` (esta carpeta) lee los transcripts principales de Claude Code (`~/.claude/projects/*/<id>.jsonl`, sin subagentes) y da dos medidas sobre los tokens procesados (input + cache creation + cache read + output de cada llamada al modelo):

- **Llamadas al visor:** tokens de las llamadas que emitieron un `hive view …` o leyeron archivos de `flow-walkthrough`, repartidos si la respuesta mezcló otras herramientas. Es el mismo criterio del 5–16 % de #328 y lo reproduce (14,8 % y 33 % en sus dos sesiones de referencia).
- **Contexto arrastrado:** tamaño estimado (caracteres / 4) de cada llamada y su resultado, multiplicado por las llamadas que siguen hasta la siguiente compactación.

No cubre Codex ni OpenCode: ninguna de sus sesiones usó el visor el 2026-11-02 después del despliegue.

## Línea base (13 sesiones, 2026-10-30 → 2026-11-02, todas antes de #328 salvo las dos últimas)

| Inicio (UTC) | Proyecto | Llamadas | Procesados | Llamadas al visor | Arrastrado |
| --- | --- | --- | --- | --- | --- |
| 10-30 06:42 | hive | 1678 | 647,5M | 3,4 % | 0,8 % |
| 11-02 05:06 | hive | 216 | 60,8M | 14,8 % | 5,2 % |
| 11-02 06:07 | hive | 33 | 3,7M | 5,3 % | 2,0 % |
| 11-02 06:33 | hive | 145 | 24,1M | 4,3 % | 1,0 % |
| 11-02 07:06 | hive | 497 | 147,4M | 8,8 % | 3,3 % |
| 11-02 07:09 | proyecto B | 393 | 124,4M | 4,6 % | 1,8 % |
| 11-02 07:57 | proyecto C | 322 | 104,0M | 11,6 % | 4,8 % |
| 11-02 09:26 | hive | 56 | 8,1M | 33,2 % | 2,6 % |
| 11-02 09:46 | hive | 242 | 81,0M | 10,9 % | 5,6 % |
| 11-02 10:33 | proyecto B | 348 | 132,5M | 2,0 % | 1,6 % |
| 11-02 10:34 | proyecto C | 222 | 48,2M | 4,8 % | 3,2 % |
| 11-02 17:03 | hive | 193 | 56,3M | 2,4 % | 1,3 % |
| 11-03 01:09 | proyecto B | 34 (en curso) | 3,9M | 7,5 % | 5,7 % |

Total: 1.442M procesados; llamadas al visor 80,8M (5,6 %); arrastrado 30,3M (2,1 %); en conjunto ~7–8 %.

Antes de #328 el costo venía de pedir el contrato (completo o por operación, 9–27 veces por sesión), las publicaciones, releer `viewer-publication.md` y `session get` (hasta 16 por sesión).

Las dos sesiones posteriores al despliegue (contenido desplegado el 2026-11-02 hacia 10:46–10:58 hora local) no sirven aún de comparación: la de hive de las 17:03 no publicó porque el visor estaba caído (5 llamadas de estado) y la del proyecto B de las 01:09 estaba en curso, con un solo `contract --for`.

## Corrección del script y medición posterior a #328 (2026-11-02, noche)

El script ahora ignora el texto escrito dentro de heredocs, cuenta lecturas de `flow-walkthrough` hechas con `cat`/`sed`/`head` y cuenta los scripts ayudantes que la sesión escribió para llamar a `hive view`. Con esa versión:

- Línea base (11 sesiones que arrancaron antes del despliegue): llamadas al visor 5,7 %, arrastrado 2,1 %.
- Después de #328 (7 sesiones de Claude Code que arrancaron desde las 11:00): 6,0 % y 2,9 %. Solo dos publicaron: una con 15,5 % y 2,3 %, y otra con 11,0 % y 6,1 % con la sesión completa. Las demás no publicaron (visor caído o sin disparador).
- OpenCode (flush del visor) ~2,7 % en total; Grok 0 %.

Conclusión: la lectura única de la referencia se cumple, pero el costo de las sesiones que publican no bajó. Las causas observadas fueron lecturas de contrato fuera del brief (`--for build` dos veces y 5 `--operation`) y un ayudante que relee la sesión antes de cada escritura. De fondo, cada publicación es una llamada del modelo que reprocesa ~180k de contexto; esto apunta a la fase 2 de #328 (verbos de alto nivel). La fila de la rúbrica que mide esto entró en el PR #338.

## Nueva medición (2026-11-08)

- Hive, después de #374 (24 sesiones desde 2026-11-05 06:00): llamadas al visor 7,2 %, arrastrado 1,9 % (896M procesados). Sesiones que publican 6–17 %, una corta 30 %. Comandos: 124 `work sync`, 24 `work open`, 20 `work contract`, 19 `status`, 18 `work tasks register`, 17 `work close`, 14 `work task update`; ningún comando con imágenes.
- Proyectos B y C (14 sesiones 2026-10-25 → 11-04, todas antes de #374; ninguna sesión posterior usó el visor): 4,9 % y 2,8 % (1.098M). De 186 comandos del visor, 4 llevan imágenes, más un ayudante (`publish_views.py`, 9 llamadas en una sesión del proyecto B): ~7 % de las llamadas son capturas; el resto es contrato, `session get`, estado y ciclo de vida.
- Conclusión: el uso que el usuario valora (capturas y su iteración) es una fracción pequeña del costo.
- Decisión: viewer-capture-only (PR #431, 69575ff7) reduce el visor a capturas con su iteración (`hive view work capture`).

## Siguiente paso

Después de unas 3 sesiones reales con capturas posteriores al despliegue de #431 (2026-11-08), correr el script con `-newermt` en esa fecha y comparar ambas medidas con esta tabla.
