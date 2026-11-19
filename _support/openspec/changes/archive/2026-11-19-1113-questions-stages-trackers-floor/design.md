# Design

## Approaches

**Hogar de preguntas (D31).** A: una hoja `hive-rule-questions` ≤7,000 bytes que absorbe las cuatro; B: hoja corta con referencias para los casos raros; C: cuatro hojas corregidas. Elegida A: Codex leyó dos de las cuatro hojas y no la de su caso; una sola lectura cubre todos los casos. Cuesta recortar cerca del 45 % de 12,873 bytes.

**Codex Default (D32).** A: ronda completa en texto; B: tarjeta que espera; C: forma del host. El usuario eligió la tarjeta que detiene el flujo (`request_user_input`, con descripción por opción). Riesgo: las instrucciones del modo Default de Codex reservan esa tarjeta para preguntas opcionales y pesan más que `AGENTS.md`; lo mide el criterio de AC2 y la auditoría de AC5.

**Piso de pruebas (D33).** A: hoja nueva; B: dentro de `flow-verify`; C: en cada playbook. Elegida A: plan y build eligen pruebas antes de verify; la hoja solo apunta a los hogares existentes porque `single_home_test.go` impide repetir sus frases.

**Trackers.** Una hoja corta más referencias por servicio para mover, comentar y adjuntar; `create-issue/references/trackers.md` conserva el procedimiento de creación y enlaza esas referencias. La caché de nombres y categorías de estado vive en Engram, `tracker/states/<tracker>/<prefijo>`, en lugar de una clave `## Hive`, porque se llena sola en la primera transición y no exige un cambio en cada repositorio; nunca reemplaza la lectura en vivo de las transiciones del ticket. `hive-rule-ticket-state` decide cuándo mover y enlaza la hoja de trackers para el cómo.

**Orden de entrega.** Primero los criterios del auditor (regla de `AGENTS.md`: el mecanismo antes de la prosa), luego preguntas y etapas, trackers y evidencia, y el piso.

## Proposed shape

`hive-rule-questions`: formato de decisión y hallazgo; preguntas de sesión con encabezado fijo; tabla de canal por host (tarjeta que espera con descripción → ronda solo en la tarjeta; sin descripción → texto y tarjeta; sin tarjeta que espere, hijo, desatendido o sin interfaz → texto); tarjeta sin respuesta, compactada; pregunta abierta de `hive-mode`.

`hive-rule-trackers`: antes de mover, leer en vivo las transiciones del ticket y elegir con la caché; ante una falla, la equivalente; guardar la caché; comentar y adjuntar evidencias con la herramienta del tracker. Referencias `jira.md`, `linear.md`, `github.md`.

`hive-rule-verification-floor`: una tabla tipo de cambio → lo que siempre corre, con enlaces a `hive-rule-tests`, `hive-rule-defect-fix`, `hive-rule-checks`, `hive-rule-blast-radius`, `hive-rule-ui-verification`, `hive-rule-before-captures`, `hive-review-ux`, `hive-verify-task`, `hive-rule-security-boundaries` y `hive-principle-explain-the-number`.

## Impact

Todas las sesiones de todos los hosts tras el despliegue: cambian las hojas que se leen al preguntar, al pasar de etapa, al mover tickets, al cerrar y al verificar. Los enlaces a las cuatro hojas borradas se buscan con `rg` en `content/`, `site/`, `_support/docs/` y `tests/`.

## Recovery

Cada unidad se revierte con `git revert` de su merge y `hive update`; el manager borra de los hosts las hojas que el release ya no trae. Las entradas de Engram de la caché se pueden borrar sin efecto en la guía.
