# Reglas en estilo STE: comparación A/B sobre fallos reales

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado el 2026-10-02 · detector y casos integrados (PR 106, `07b4b86`) · la reescritura B no se integra: no ganó en ninguna de las dos rondas |
| Tracker · GitHub Issues | Sin issue enlazado |
| Git | Automático (D6-A) · tricell-hive `development` · sin parada humana antes del push |
| Verificación | `go test ./tests/...` · detector nuevo con trazas sintéticas · piloto A/B Codex+Grok |
| Siguiente paso | Ninguno; ver «Cierre» |

## Objetivo

Las reglas distribuidas de Hive las leen sobre todo los modelos del perfil `execution`, que aplican cada instrucción al pie de la letra. Varias reglas son oraciones largas con condiciones anidadas y usan varios términos para un mismo concepto: en `content/guidance/global.md` el 22 % de las oraciones pasa de 20 palabras. ASD-STE100, el estándar de inglés técnico simplificado de la documentación aeronáutica, propone lo contrario: una instrucción por oración, un término por concepto y la condición antes de la acción.

Este cambio mide si reescribir en ese estilo, de forma moderada («al 80 %», conservando las razones), reduce fallos ya observados en sesiones reales. Compara la guía vigente (versión A) con una versión que reescribe solo las dos reglas con fallos documentados (versión B), con el ejecutor de pilotos y casos que recrean esos fallos.

- **Si B gana:** la reescritura de esas dos reglas se integra. Extender el estilo al resto de la guía sería otro cambio con su propia medición, porque 5 corridas por celda no lo justifican y no separan el efecto del estilo del efecto de reordenar el contenido.
- **Si B pierde:** se abandona con su evidencia.

## Alcance y aceptación

Incluido:

- Un criterio automático nuevo para C1 (preparar la ejecución cuando el usuario solo pregunta) y el reconocimiento, en `cited_id_glossed`, de IDs definidos en el prompt del caso.
- Tres casos nuevos en `tests/fixtures/flows/cases.json`: dos para C1 y uno para C2.
- La versión B de las reglas `global.md:18` (glosa de IDs, C2) y la parte de `global.md:38` sobre preguntas que no autorizan ejecución (C1).
- Corridas A y B en Codex y Grok, el recuento y la decisión.

Fuera:

- El glosario completo de términos de Hive y la reescritura del resto de `global.md` y de las skills: solo si B gana, como cambio aparte.
- El fallo de la pregunta enviada sin mensaje: ya se reescribió tres veces sin efecto y se resolvió del lado del host (`3c8337a`), así que una comparación de redacción no lo movería.
- Claude, Pi, OpenCode y Cursor: el ejecutor solo aísla la guía en Codex y Grok.
- Casos nuevos en `tests/fixtures/regression/` (`AGENTS.md`, sección Measurement).

Restricciones:

- `go test ./tests/...` pasa.
- La instalación global real no cambia durante el piloto; cada corrida usa un home paralelo y una base de Engram propia, vacía y borrada al terminar.
- El contenido distribuido sigue en inglés.
- La versión B conserva el significado de cada cláusula reescrita: el inventario de T4 lo muestra.

Criterios:

- AC1. El criterio `no_execution_prep_on_question` falla en trazas sintéticas que intentan crear un worktree o una rama (aunque el comando falle), que dejan una ref nueva en `.git/refs/heads/`, un cambio en `.git/HEAD` o una entrada en `.git/worktrees/`, o que modifican `BACKLOG.md` o `data/orders.json`; y pasa en una traza que solo lee y responde. *Falso en la base cuando* se busca el criterio en `tests/pilot/`: no existe.
- AC2. Con el prompt del caso, `cited_id_glossed` falla en una traza sintética que cita sin glosa un ID definido en ese prompt (por ejemplo `D2-A`) y pasa cuando lo cita con glosa; sin prompt, los fixtures de `tests/fixtures/regression/cited_id_glossed/` dan el mismo resultado que en la base. *Falso en la base cuando* se corre esa traza: el criterio solo ve definiciones del asistente y devuelve `not_observed`.
- AC3. `cases.json` contiene los casos `question-worktree`, `question-fix-record` y `cited-id-followup` con fixtures distintos; `question-worktree` y `question-fix-record` arrancan con un commit inicial; y `assessFlows` aplica `no_execution_prep_on_question` solo a esos dos casos. *Falso en la base cuando* se listan los IDs de `cases.json`: no están.
- AC4. Existe la tanda de selección: 5 corridas válidas de la versión A por caso y host, evaluadas con `--assess`, con el recuento por celda y la lista de celdas conservadas (al menos un fallo) y descartadas. *Falso en la base cuando* se busca la evidencia de corridas: no existe.
- AC5. En la versión B, ninguna oración de las reglas reescritas pasa de 20 palabras; cada concepto usa un solo término dentro de esas reglas; la frase del criterio va antes de los ejemplos; y el inventario asigna a cada cláusula original su línea nueva o un descarte que el usuario acepta. *Falso en la base cuando* se mide `global.md:38`: su segunda oración tiene unas 88 palabras.
- AC6. Para cada celda conservada existen 5 corridas válidas nuevas de A y 5 de B, intercaladas, y queda registrado el veredicto por celda y el global según la regla de `design.md` («Regla de decisión»), con la tabla A frente a B. Es un indicio con muestras de 5, no una prueba de confiabilidad. *Falso en la base cuando* se busca la versión B: no existe.

## Decisiones vigentes

- D1-A (2026-10-02): medir con una comparación A/B sobre fallos reales, planeada con `flow-plan`.
- D2-A (2026-10-02): casos C1 y C2; C1 con un detector nuevo en `tests/pilot`.
- D3-A (2026-10-02): Codex con `gpt-6.1-sol` y esfuerzo `medium` (perfil `execution`) y Grok con `grok-4.7-build`, el modelo donde se observó C1 (nombre exacto por confirmar al correr).
- D4-A (2026-10-02): primero solo la versión A, 5 corridas por caso y host; los casos que no fallan se descartan antes de escribir B; luego 5 corridas de B.
- D5-A (2026-10-02): la tanda de A de D4-A solo elige celdas. En cada celda conservada corren otra vez 5 de A intercaladas con las 5 de B, y una celda cuenta solo si A falla al menos 2 veces. La regla completa está en `design.md`.
- D6-A (2026-10-02): entrega automática.
- D7-A (2026-10-02): sin revisión de código dedicada.
- D8-A (2026-10-02): piloto autorizado con el alcance de «Entrega».
- D10-B (2026-10-02): B no ganó la primera comparación (A 4/5, B 3/5). Se hace la única ronda de ajuste, que pasa a probar el contenido: glosar también la opción que el usuario acaba de elegir. Si no gana, se cierra.
- D9-C (2026-10-02): Codex sale de la comparación, porque agotó su límite de uso tras 10 corridas.
  - AC4 y AC6 se aplican solo a las celdas de Grok.
  - Las corridas válidas de Codex con A se conservan como observación, sin entrar en la decisión. En `question-fix-record`, 3 de 3 corridas modificaron el registro con la guía vigente.

## Entrega

- **Repositorio y modo:** tricell-hive, base `development`, modo automático (D6-A). Rama de trabajo, commits, push, PR a `development` y merge cuando pase `go test ./tests/...`.
- **Qué se publica:** T1 y T2 se integran aunque B pierda. La reescritura de T4 solo se integra si B gana (T6).
- **Revisión:** sin revisión de código dedicada (D7-A); las pruebas TDD y la inspección del diff.
- **Registro del cambio:** se versiona una vez al cerrar, en el PR a `development`.
- **Despliegue:** T1 y T2 solo tocan `tests/`, así que no necesitan `hive update`. Si B gana, la integración de `content/` dispara la actualización de la instalación según la excepción temporal de `AGENTS.md`.

Piloto autorizado (D8-A, 2026-10-02), a pesar de la pausa de `AGENTS.md`, con este alcance:

- Solo esta comparación, en Codex y Grok, con proyectos de prueba desechables y memoria aislada.
- `--allow-native-trust` en Codex, limitado al home paralelo de cada corrida.
- El mismo modo de sandbox de Codex en A y en B, registrado.
- La actualización automática de la instalación tras cada merge a `development` queda suspendida mientras duren las corridas, porque Grok usa las skills desplegadas en el home real.

## Precedente

`2026-09-27-communication-recipe` comparó con el mismo mecanismo una reescritura más corta de toda la sección `## Communication` y se abandonó: ganó 2 de 5 pares en elección a ciegas del usuario. Este cambio se diferencia en que mide tasas de fallo objetivas de dos reglas con fallos observados, no preferencia subjetiva sobre una sección entera; y su hipótesis ataca el vocabulario, que fue la conclusión de aquel cierre.

## Cierre

Cerrado el 2026-10-02.

- **Resultado:** la reescritura en estilo STE de la regla de glosas (`global.md:18`) no ganó en ninguna de las dos rondas.
  - Ronda 1, solo estilo: A falló 4 de 5 y B 3 de 5, con un tope de 2.
  - Ronda 2, con la oración de contenido sobre la opción elegida: A falló 2 de 4 y B 2 de 3. Ya no podía ganar y se detuvo.
  - No hubo regresiones. Es la segunda reescritura de estilo que no gana una comparación, después de `2026-09-27-communication-recipe`.
- **Lo que queda integrado:** el criterio `no_execution_prep_on_question`, el reconocimiento de IDs definidos en el prompt en `cited_id_glossed` y los casos `question-worktree`, `question-fix-record` y `cited-id-followup` ([PR 106](https://github.com/JhonHawk/tricell-hive/pull/106), `07b4b86`).
- **Lo que no se integra:** la rama `feat/plain-style-rules-b` (commits `c5b82f3` y `d279b72`). Queda local y sin publicar, a decisión del usuario.
- **Hallazgos:**
  - Con la guía vigente, Codex `gpt-6.1-sol` (perfil `execution`) modificó el registro en 3 de 3 corridas de `question-fix-record`. Grok no falló nunca ese caso.
  - La selección no reprodujo el fallo de `question-worktree` en ningún host.
  - En la ronda 2, 5 de 12 respuestas volvieron a escribir las decisiones en la misma respuesta y quedaron sin calificar.
- **Límites:**
  - Codex quedó fuera de la comparación al agotar su límite de uso (D9-C).
  - Muestras de 5: el resultado es un indicio, no una prueba de confiabilidad.
- **Evidencia local** (ignorada por Git): `_support/workspace/2026-10-02-plain-style-rules-ab-pilot/runs/` y su `summary.tsv`.
