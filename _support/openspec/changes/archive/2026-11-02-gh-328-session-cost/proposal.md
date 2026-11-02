# #328 — Costo de sesión: preguntas directas para acciones sueltas, compactar en transiciones y contrato del visor por flujo

| Campo | Valor |
| --- | --- |
| Estado | Completado · integrado por el PR #329 (`36ac21ee`) el 2026-11-02 |
| Tracker | [JhonHawk/tricell-hive#328](https://github.com/JhonHawk/tricell-hive/issues/328) · cerrado · relacionado con #324 (quitó la recomendación de compactar) y #319 (`work_key`) |
| Git | `tricell-hive`, base `development` en `5e414d77`; rama propuesta `feat/gh-328-session-cost` en `.claude/worktrees/gh-328-session-cost` |

## Decisiones

2026-11-02
- ¿Cuándo se ofrece el comando de un flujo de Hive? Solo cuando el paso siguiente necesita el procedimiento de ese flujo. Una acción mecánica suelta, como crear un ticket, se pregunta directamente («¿Quieres que lo haga?») y no se manda por `/flow-build`. (D2, ampliada durante este plan)
- ¿Qué acciones se preguntan directo? Las que no tocan archivos del repositorio: crear, mover o cerrar un ticket, borrar una rama integrada, guardar en memoria, abrir el visor. Los cambios en archivos, aunque sean mecánicos, siguen por `/flow-build`, que lleva la pregunta de entrega y la verificación. (D9 A)
- ¿Se vuelve a recomendar compactar? Sí, en dos momentos y nunca por porcentaje: al pasar de un plan aprobado al build en la misma sesión, y después de un cierre cuando lo siguiente es otro trabajo. (D5 A)
- ¿Cómo se baja el costo del visor? Con reglas de lectura única y un CLI que entrega el contrato por partes; los comandos de alto nivel con archivo de estado (fase 2) y un servidor MCP quedan fuera. (D6, D8 A)
- ¿Cómo se pide el contrato por partes? `hive view contract --for <flujo>` en versión corta, `--section rules` para el bloque común, `--list` con una línea por operación y `--operation` completo, todo en JSON sin sangría. (D8 A, más `--list` pedido después)
- ¿Un issue o varios? Uno solo, #328. (D7, cambiada después a un issue)
- ¿Se guarda el informe de la revisión como página? No. (D4 A)
- ¿Cómo se entrega? Automática: push, PR y merge con la CI en verde, sin parada de validación. (D10 A)
- ¿Qué revisión de código corre? Claude Code `/code-review`. (D11 A)

## Pedido

Del issue #328 (leído el 2026-11-02): tres cambios que salieron de revisar 16 sesiones de Claude Code (tricell-hive, cam y carvento, del 31-oct al 2-nov). Criterios en borrador del issue, citados:

- «AC1: A route that needs a missing ticket ends with the question whether to create it; no tracker write happens before the answer.»
- «AC2: The next-step block offers compaction at plan → build in the same session and after a close followed by another work item, and nowhere based on a percentage.»
- «AC3: `contract --for <flow>`, `--section rules`, `--list` and compact `--operation` exist, and a test fails when an operation belongs to no flow.»
- «AC4: `flow-walkthrough` and `viewer-publication.md` state the single-read rules above; `site/` pages and architecture docs that describe these behaviors are updated.»

Fuera del issue: la fase 2 del visor (verbos con archivo de estado), un servidor MCP y medir el efecto en el comportamiento del modelo (los pilotos están en pausa).

## Objetivo

Hoy:
- Una acción suelta, como crear un ticket, se ofrece con un bloque de siguiente paso que pide responder una frase o escribir un comando; en cam `32cffc1d` y tricell-hive `9187491a` eso costó una o dos respuestas extra.
- Nada recomienda compactar: los builds de cam, carvento y #319 empezaron con 207–324k tokens y un ticket nuevo de cam corrió sobre 543–602k del anterior.
- Publicar en el visor se llevó entre 5 y 16% de los tokens procesados: el contrato se pidió 16–33 veces por sesión, `viewer-publication.md` se releyó hasta 18 veces y `session get` corrió 13–43 veces.

Con el cambio: las acciones sueltas se preguntan en la ronda, el bloque de siguiente paso ofrece compactar en los dos momentos acordados, y el visor se publica leyendo una vez la referencia y el contrato del flujo en curso, que pesa entre 1k y 6k caracteres.

## Alcance y criterios

Entra: la regla de «Questions and routes» en `content/guidance/global.md` y sus punteros, `plan-handoff.md`, `flow-close`, las páginas del sitio de esos flujos, el comando `hive view contract` y su salida, el mensaje de error de una entrada rechazada, las reglas de `flow-walkthrough` y su referencia, la documentación del contrato y el delta de la spec `walkthrough-viewer`.

No entra: fase 2 del visor, MCP, cambios al protocolo v3 o a los esquemas, y cualquier cambio a qué hace `flow-build` con los cambios mecánicos en archivos.

- AC1 — `global.md` dice que el comando de un flujo se ofrece solo cuando el paso siguiente necesita el procedimiento de ese flujo, y que una acción suelta que no toca archivos del repositorio y no está ya autorizada por «Scope and authorization» (como crear o cerrar un ticket, borrar una rama integrada o guardar en memoria) se ofrece como respuesta del bloque cuando el mensaje termina en uno, o como decisión de la ronda cuando no; la respuesta autoriza solo esa acción; un flujo que necesita un ticket inexistente pregunta si lo crea y, ya creado, cierra con el bloque que nombra el ticket real · falso hoy porque `global.md:66` manda «recommend creating the ticket first» dentro del bloque de siguiente paso.
- AC2 — El bloque de siguiente paso ofrece, como respuesta opcional, compactar con el comando de compactación del host y lo que conservar, en dos momentos: al cerrar un plan antes del build en la misma sesión, y en el informe que sigue a un cierre cuando lo siguiente es otro trabajo; tras esa compactación el agente vuelve a cerrar con el mismo bloque y no empieza el flujo sin la respuesta; ninguna regla condiciona compactar a un porcentaje · falso hoy porque el PR #324 quitó la única recomendación y no hay otra.
- AC3 — `hive view contract --for <flujo>` (research, plan, build, verify, delivery, close, collab) devuelve en JSON sin sangría solo las operaciones de ese flujo más las comunes (crear y leer la sesión, registrar el ticket), cada una con su uso, campos obligatorios y opcionales, valores permitidos y ejemplo, en menos de 10.000 bytes por flujo; `--list` devuelve una entrada por operación con sus flujos y su uso; `--section rules` devuelve el bloque común; `--operation NAME` devuelve el mismo contenido que hoy sin sangría, salvo `reason`, que el esquema de `stage_start` y `stage_update` ahora marca obligatoria como ya exigía el validador; una prueba falla si una operación no tiene flujo · falso hoy porque `contract` solo acepta `--operation` y su salida va con sangría (`view_timeline.go:493`).
- AC4 — Una entrada rechazada por validación nombra el campo, como `invalid timeline input: reason: invalid_text` · falso hoy porque en esta sesión `stage start` sin `reason` respondió solo «invalid timeline input: invalid_text».
- AC5 — `flow-walkthrough/SKILL.md` y `viewer-publication.md` dicen: leer la referencia una vez por sesión y otra después de compactar; pedir `contract --for <flujo>` una vez por flujo y volver a pedirlo solo tras compactar o tras un error de validación; usar las versiones e IDs que devuelve cada cambio en vez de `session get`; releer la sesión completa antes de decirle al usuario que algo está visible, al cerrar y al retomar tras una interrupción o un conflicto de versión · falso hoy porque `SKILL.md:8` y `viewer-publication.md:3` piden leer la referencia «at each stage change» y `viewer-publication.md:37` pide releer después de cada publicación.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

- `content/guidance/global.md:66` (22.263 bytes) es la casa canónica del bloque de siguiente paso y de la frase del ticket inexistente; ningún test ni detector de `tests/pilot` depende de esa frase.
- `git show 83722449` (PR #324) quitó de `plan-handoff.md` la línea que recomendaba compactar al pasar el 40% de la ventana; su CHANGELOG lo explica.
- `flow-close` termina con la pregunta de cierre (una tarjeta), no con un bloque; su opción «Find the next ticket» corre `flow-status`, cuyo informe sí cierra con bloque.
- `tooling/viewer/timeline_protocol.go:51-110` arma el contrato: bloque común (`authority`, `error_codes`, `invariants`, `limits`, `localization`, `capabilities`) más 28 operaciones con `usage`, `input_schema`, `output_schema` y `example`. Medido el 2026-11-02: `stage_start` 18,7k caracteres con sangría y 12,0k sin ella; `session_get` 93k y 32k; el bloque común ~6,5k.
- `tooling/cli/view_timeline.go:96` imprime solo el código de `TimelineContractError`, aunque el error ya trae el campo (`timeline_contract.go:70`).
- La spec `walkthrough-viewer` exige «verify committed readback before claiming a visible update» (`spec.md` requisito «Proportional identity-aware publication procedure») y describe el contrato en «Self-describing protocol and confirmed-input hygiene».
- Investigación de Claude Code (2026-11-02): la opción «clear context» al aceptar un plan se ocultó en la v2.1.81 (CHANGELOG, 2026-03-20) por la ventana de 1M y los borrados accidentales; el modelo no recibe el uso de contexto; la documentación del host aconseja `/clear` entre trabajos no relacionados.

## Entrega

Automática (D10 A): reviso, hago push, abro el PR a `development` y lo fusiono con la CI en verde, sin parada de validación. Revisión de código: Claude Code `/code-review` sobre el diff del candidato (D11 A). El commit nombra las sesiones observadas (cam `32cffc1d`, carvento `8d09a5eb`, tricell-hive `98476eca`, `9187491a`, `7bcc87ae`, `da51e1c1`; Claude Code, 2026-11-02). Con el marcador de mantenedor, al integrarse se reconstruye el binario y se corre `hive update`. No autoriza una release.

## Revisión y avance

2026-11-02 — Plan escrito. La revisión del plan (`hive-review-plan`) encontró 1 bloqueante, 5 por corregir y 4 menores, todos aplicados: el límite de tamaño de `--for` pasó de 6.000 a 10.000 bytes según la medición (los ejemplos ocupan la mitad de cada entrada); cada `--for` suma las operaciones comunes; las acciones sueltas van como respuesta del bloque cuando hay bloque; se excluyen las escrituras al tracker que la guía ya autoriza; el comando de compactar se nombra por capacidad y la compactación en una parada de ruta no arranca el flujo; T1 suma `flow-research:24` e `instruction-resources.md:24`; el error usa el campo de `TimelineContractError`; T3 dice qué resúmenes lee cada skill; T2 suma su entrada de CHANGELOG y su prueba va en `view_timeline_v3_test.go`.
