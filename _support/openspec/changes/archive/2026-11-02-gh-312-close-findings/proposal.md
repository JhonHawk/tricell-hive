# Hallazgos del cierre de #312 — spec de hive-core y redacción

| Campo | Valor |
| --- | --- |
| Estado | Completed · se integra con su propio PR el 2026-11-02 |
| Tracker | [#312](https://github.com/JhonHawk/tricell-hive/issues/312) · cerrada · hallazgos de su cierre |
| Git | `JhonHawk/tricell-hive`, base `development` en `e2787f26`, rama `fix/gh-312-close-findings` |

## Decisiones

2026-11-02
- ¿Se resuelven los hallazgos del cierre de #312? Sí, todos en un cambio chico, sin plan. (pedido del usuario)
- ¿Cómo se entrega? `interactive`, con Claude Code `/code-review` antes de la validación. (D14 A, D15 A)

## Pedido

> resolvamos los hallazgos porfa, se ven chicos

## Alcance

- H17 (P2): el requisito «Layered global core with a shared core skill» exigía la skill `hive-core`, disuelta en `5e6e8907`; se retira y lo reemplaza «Layered global core with owning skills».
- P3 de `git-workflow`: el bot pendiente tampoco frena la parada de `ready`; cuenta como bot la cuenta con la que publica la automatización de revisión; un comentario de bot cuenta si apunta al commit de cabeza o se creó después del último push; el paso 9 ya no repite los endpoints de revisiones del paso 7; el paso 10 cierra también al abrir un PR con revisor humano.
- P3 de `flow-close` y del manifiesto: un walkthrough legado sigue `in_progress` mientras su PR espera; la sesión que encuentra el PR mergeado solo hace lo que esperaba al merge.

## Entrega

`interactive` (D14 A) con `/code-review` (D15 A). Este registro va directo a `archive/` porque se integra en el mismo PR que el cambio.
