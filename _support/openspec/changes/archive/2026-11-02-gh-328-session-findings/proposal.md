# #328 — Hallazgos de la revisión: resumen del contrato completo y error alineado

| Campo | Valor |
| --- | --- |
| Estado | Completado · hallazgos no bloqueantes del cierre de #328, resueltos por decisión del usuario |
| Tracker | [JhonHawk/tricell-hive#328](https://github.com/JhonHawk/tricell-hive/issues/328) · cerrado |
| Git | `tricell-hive`, base `development` en `baeeaa9a`; rama `fix/gh-328-review-findings` |

## Decisiones

2026-11-02
- ¿Qué se hace con los hallazgos no bloqueantes del cierre de #328? Se resuelven en este mismo trabajo, con entrega automática y `/code-review`. La salida compacta en todo `hive view` (H17) queda como la decidió el plan.

## Alcance

- H14: `session_delete` y `recover_lead` pasan a un flujo `local`; `collab` queda con unirse, aceptar productor y pasar el foco.
- H15 y H26: los valores permitidos recorren `anyOf`, `oneOf` y `allOf` y se combinan por ruta; `actor.kind` se omite fuera de `local`, porque un productor nativo solo puede usar `native_producer` y así `--for delivery` cabe en 10.000 bytes.
- H16: el error local usa el orden del servidor: `invalid timeline input: invalid_text on reason`.
- H18: `--list`, `--section rules` y `--for` devuelven el error real en vez de ignorarlo o llamarlo flujo desconocido.
- H23: los obligatorios de elementos de lista aparecen con `[]`, como `segments[].planned_stage`.
- H24: una prueba falla si un nombre obligatorio no tiene propiedad.
- H25: el resumen de `event_patch` avisa que `reason` es obligatoria cuando cambia `stage_state`.
- H19, H22 y H27: las pruebas comparan contra el contrato, nombran las operaciones que se saltan y usan una lista explícita de operaciones `_patch`.
