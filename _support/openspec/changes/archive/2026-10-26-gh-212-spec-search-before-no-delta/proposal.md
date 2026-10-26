# Specs que contradicen el código y búsqueda en specs antes de declarar «sin delta» (2026-10-26)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` el 2026-10-26 por el [PR #215](https://github.com/JhonHawk/tricell-hive/pull/215), merge `678010d2`; #212 cerrado |
| Tracker · GitHub Issues | • [#212 — Specs de versioned-installation contradicen el código y los planes no buscan requisitos afectados antes de declarar sin delta](https://github.com/JhonHawk/tricell-hive/issues/212) |
| Git | `automatic` (D1-A) · rama `feat/gh-212-spec-search-before-no-delta` desde `development` · registro por PR de cierre pequeño, integrado por el agente |
| Verificación | comprobaciones `rg` y aplicación de deltas en una copia temporal por tarea · `go vet ./...`, `go test ./...` y `npm run build` en `site/` · `/code-review` de Claude Code (D2-A) |
| Siguiente paso | Ninguno; registro cerrado |

## Solicitud

Mensajes del usuario, 2026-10-26:

> Qué opinas de esto? [Imagen: primera página de *Spec-Driven Development as Prescribed and as Practiced*, arXiv:2608.25202v4]

Respuestas siguientes, 2026-10-26, en tarjetas nativas:

- «Lectura + contraste (Recomendado)»: leer el paper completo y contrastarlo con Hive.
- «Seguir investigando»: clasificar los 40 requisitos y revisar si los registros guardan el porqué.
- «R2 specs + regla»: abrir el ticket y seguir con `flow-plan` para corregir las specs y añadir la regla de búsqueda, adoptando la corrección con un solo caso observado → [#212](https://github.com/JhonHawk/tricell-hive/issues/212).

Criterios de aceptación de #212, leídos el 2026-10-26:

> - AC1: el requisito de capacidades opcionales incluye la capacidad del tiempo mínimo de espera de Codex, y su lista coincide con el catálogo.
> - AC2: el requisito de diagnóstico permite la lectura de `config.toml` de Codex que hace la fila del tiempo mínimo de espera, sin repetir su contenido, y sigue prohibiendo cualquier otra lectura de configuración de hosts.
> - AC3: `documentation-site` ya no afirma una página por cada skill de flujo, o nombra cuáles no la tienen.
> - AC4: el requisito de overrides por rol dice que Pi no acepta `ultra`.
> - AC5: los cuatro cambios llegan como deltas MODIFIED de un change record archivado, y `TestSpecDeltasApplied` pasa.
> - AC6: `change-records.md` pide, antes de declarar que un cambio no necesita delta, buscar en las specs los requisitos afectados, incluidas listas cerradas y `SHALL NOT`, y anotar en `proposal.md` qué se buscó.

## Objetivo

Dos requisitos de `versioned-installation` dicen hoy algo falso. La lista de capacidades opcionales se declara cerrada y le falta el tiempo mínimo de espera de Codex. El diagnóstico promete no leer configuración de hosts, pero lee `config.toml` de Codex para esa misma capacidad. Otros dos requisitos quedaron imprecisos: «una página por skill de flujo» en el sitio y la lista de esfuerzos, que no dice que Pi rechaza `ultra`. Quien lea esas specs, persona o agente, recibe un contrato que el código ya no cumple.

La causa de los dos primeros es una decisión de planificación. El plan de gh-142 declaró que ninguna spec cubría el instalador de capacidades opcionales, y sí había una. La regla ya pedía un delta cuando un cambio altera un requisito que una spec declara, pero no pedía buscar antes de descartarlo, y la prueba de deltas no puede detectarlo porque compara cada spec con sus deltas archivados, no con el código.

Después del cambio, las cuatro specs describen el código actual, y un plan que declare «sin delta» deja escrito qué buscó en las specs y qué requisitos leyó, incluidas las listas cerradas y las prohibiciones.

## Alcance y aceptación

**Incluido:**

- deltas MODIFIED de *Optional operation capabilities*, *Read-only diagnostics* y *Per-role model overrides* en `versioned-installation`, y de *Published documentation site* en `documentation-site`;
- la regla de búsqueda en `content/skills/flow-plan/references/change-records.md`, su hogar canónico;
- una frase en `site/src/content/docs/flows/plan.md`, que describe el change folder;
- `CHANGELOG.md`.

**Excluido:**

- **Una regla sobre cuánto detalle de implementación admite una spec:** los 40 requisitos revisados coinciden con el código salvo estos cuatro, y la deriva vino de listas cerradas y prohibiciones, no de detalle volátil.
- **Un chequeo automático de deriva entre specs y código después de cerrar:** la evidencia apunta a prevenir al planificar. Queda en vigilancia.
- **Enlazar cada requisito con el cambio que lo creó, y hacer obligatorias las alternativas descartadas en `design.md`:** sin fallo observado. Quedan en vigilancia.
- **`global.md`:** `global.md:104` ya apunta a `change-records.md`, donde vive la regla. `plan-format.md` solo recibe un puntero al registro de la búsqueda, añadido tras la revisión de código; #204 ya está integrado en la base (`f14eb791`).
- **Las ediciones pequeñas sin change folder:** la regla aplica al decidir el delta de un change record; una edición sin plan no escribe deltas.
- **El puntero roto de `tests/fixtures/regression/no_bare_url/provenance.md:3`:** AGENTS.md pide no tocar esos casos sin decisión del usuario. Se informa aparte.
**Búsqueda en specs que decidió los deltas** (la regla nueva aplicada a este cambio): se buscaron `codex-wait-floor`, `CodexWaitFloor`, `optional capabilit`, `config.toml`, `doctor`, `ultra`, `effort`, `one page per flow`, `flows/` y `plan remove` en `_support/openspec/specs/`; se leyeron los cuatro requisitos que se modifican. Ningún otro requisito los nombra. `context-delivery` y `walkthrough-viewer` no cambian.

**Restricciones:**

- La suite local sigue pasando (`go vet ./...`, `go test ./...`) y el sitio construye.
- Los requisitos no tocados quedan byte a byte iguales al aplicar los deltas.
- La regla nueva no repite pasos de otra regla y no aparece en otro archivo salvo como puntero.

**Criterios de aceptación:**

- AC1. *Optional operation capabilities* nombra `codex-wait-floor` en su lista cerrada y describe qué escribe, qué rechaza y cómo se deshace, y esa lista coincide con los IDs de `tooling/providers/catalog.go`. *Falso en la base cuando* la spec lista seis capacidades y el catálogo siete.
- AC2. *Read-only diagnostics* permite solo la lectura de `config.toml` de Codex para la fila del tiempo mínimo de espera, sin mostrar texto del archivo, y mantiene la prohibición para cualquier otra configuración de hosts. *Falso en la base cuando* la spec prohíbe toda lectura de configuración de hosts y `codexWaitFloorRow` lee `config.toml`.
- AC3. *Published documentation site* deja de afirmar una página por cada skill de flujo y nombra las skills sin página. *Falso en la base cuando* la spec dice «one page per flow skill» y `flow-report` y `flow-walkthrough` no tienen página.
- AC4. *Per-role model overrides* rechaza `ultra` en Pi en el texto y en su escenario. *Falso en la base cuando* la spec acepta `ultra` en cualquier host que admita esfuerzo y `integrations/agents/agents.go` lo rechaza en Pi.
- AC5. Aplicados en una copia temporal, los deltas cambian solo esos cuatro requisitos, cada nombre MODIFIED coincide con un requisito existente y la copia pasa la misma comprobación que `TestSpecDeltasApplied`. *Falso en la base cuando* no existe ningún delta para esos requisitos.
- AC6. `change-records.md` pide, antes de decidir qué requisitos altera un change record o que no necesita delta, y antes de aplicar la exención de «implementation, tooling…», buscar lo que el cambio añade o altera, y la lista, el catálogo o el informe al que se suma un elemento nuevo, en el hogar de requisitos vigentes y en los documentos antiguos que aún rigen el área, leer cada requisito o regla del mapa de producto que esas búsquedas encuentren, incluidos los que enumeran un conjunto cerrado y los que prohíben algo, y anotar en `proposal.md` los términos buscados y los requisitos leídos. *Falso en la base cuando* `change-records.md` solo da el criterio del delta, sin búsqueda ni registro.

El criterio AC5 de #212 pide además que el change record quede archivado y que `TestSpecDeltasApplied` pase sobre el repositorio: eso ocurre en el cierre (`flow-close`), después de integrar, y queda como condición del cierre, no de una tarea.

## Entrega

Respuestas del 2026-10-26: D1-A, entrega `automatic` con el registro versionado en un PR de cierre pequeño sin revisión, integrado por el agente; D2-A, `/code-review` de Claude Code sobre el candidato local; D3-A, arrancar el build (aprueba AC1–AC6); D4-A, quitar el escenario nuevo del tiempo mínimo de espera de Codex tras tres verificaciones fallidas.

Requisitos de AGENTS.md («Guidance structure» y «Measurement») para la publicación, con cualquier modo:

- el cuerpo del PR lista los sitios revisados (el inventario de [design.md](design.md#contexto-verificado-2026-10-26-development-en-2e45e292)) y dice que `content/guidance/global.md` no cambia de tamaño (0 bytes);
- el commit de la regla nombra la falla (gh-142, 2026-10-25), dice que el ID de esa sesión no se pudo recuperar, nombra esta sesión (Claude Code, `abdd6ea1-e8bb-4ebc-8d12-6284d022d322`, 2026-10-26) y que el usuario aprobó adoptar la corrección con un solo caso.

## Temas abiertos

- **Opinión sobre el paper:** respondida en la conversación; este cambio es su destino.
- **Chequeo de deriva después de cerrar:** en vigilancia; se reabre si una spec vuelve a contradecir el código tras un plan que buscó en las specs.
- **Regla de longevidad del detalle en specs:** descartada por la evidencia de los 40 requisitos.
- **Enlace requisito → cambio y alternativas obligatorias:** en vigilancia, sin fallo observado.
- **Puntero roto en el caso de regresión `no_bare_url`:** solo informado; depende de la decisión del usuario sobre esos casos.
- **Que la revisión del plan compruebe la búsqueda anotada:** ningún dominio de `plan-review.md` la revisa hoy. En vigilancia hasta que se integre #204, que modifica ese archivo; se reabre si un plan anota una búsqueda incompleta y nadie lo detecta.
