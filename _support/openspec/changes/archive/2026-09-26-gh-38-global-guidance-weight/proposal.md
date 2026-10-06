# Poda del peso de `global.md`

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `rebuild/harness-engineering` (`08abd00`) y desplegado como release `baf1ec3c938e` |
| Tracker · GitHub Issues | • [#38 — Poda del peso de global.md](https://github.com/JhonHawk/tricell-hive-private/issues/38) |
| Git | Commit `08abd00` con push a `rebuild/harness-engineering` |
| Verificación | `go vet ./...` · `go test -race ./...` · validación de enlaces del release · revisión independiente de equivalencia (K4-A) |
| Siguiente paso | Ninguno en este cambio; el efecto en la atención no se midió (pilotos pausados) |

## Objetivo

`content/guidance/global.md` pesa 41177 bytes y se carga en cada sesión de los seis hosts. El [issue #38](https://github.com/JhonHawk/tricell-hive-private/issues/38) pide reducirlo sin perder lo que necesitan los modelos de ejecución: el perfil `execution` de `integrations/agent-profiles.json`, con `deepseek-v4.1-flash` como piso, y la clase Haiku, que pidió el usuario.

La auditoría del 2026-09-26 (tres `review-harness` en paralelo, método `prompt-audit` más `harness-audit`) encontró que la mayor parte es esencial. Las reglas de ubicación de artefactos no pueden depender de un skill (`AGENTS.md`). La parte de la delegación que toca al hilo principal debe seguir global, porque los hijos genéricos no reciben otra cosa. Y la entrega por skills o referencias llegó al modelo solo en ~45–70% de las sesiones medidas (`_support/docs/architecture/repository-and-distribution.md`). El cambio condensa sin cambiar el significado (C1) y mueve tres bloques situacionales a skills (C2), dejando en global el disparador y lo que falló en sesiones reales. De paso corrige dos contradicciones (V1 y V2).

## Alcance y aceptación

**Incluye:**
- C1: condensaciones, fusiones y eliminación de duplicados dentro de `global.md` ([design.md](design.md#c1--condensar-sin-cambiar-el-significado)).
- C2: el cuerpo del reporte de backlog pasa a `flow-research/references/backlog-report.md`; el contenido de `project.md` pasa a `change-records.md`; la frase `Review` de `:106` se quita porque `delivery-decisions.md:22,26` ya la cubre, y ahí se agrega que el ajuste `Review` cuenta como preferencia explícita.
- **Se queda sin cambio:** `:90`, las "shared dispatch evidence requirements" que cita `plan-review.md:26`.
- V1: `:42` deja de usar la rama base del workspace como respaldo.
- V2: `flow-build/references/verification.md:67` remite a la regla global de imágenes sin repetir la ubicación.
- `globalGuidanceBudget` baja al tamaño nuevo.

**Excluye:**
- Quitar reglas clasificadas como "conducta por defecto" (`:25`, `:26`, `:67`): no hay evidencia de que los modelos de ejecución se comporten bien sin ellas, y los pilotos siguen pausados.
- Mover reglas de ubicación, autorización, secretos o preservación.
- La presión de guardado del plugin de Engram (fuera del alcance de #38).

**Criterios de aceptación:**
- A1. Cada regla quitada, movida o condensada tiene su clasificación y la evidencia de por qué es segura en [design.md](design.md). Lo que se mueve conserva una sola casa canónica, con un enlace desde el skill que lo carga.
- A2. Ninguna regla de autorización, secretos, preservación ni ubicación de artefactos sale de la guía global. La revisión independiente (K4-A) confirma que ninguna reescritura pierde una restricción.
- A3. `globalGuidanceBudget` baja al tamaño nuevo, y `go test ./...` y `go vet ./...` pasan, incluida la validación de enlaces del release para los skills. El puntero `skill:` de `global.md` queda fuera de esa validación.
- A4. El reporte declara la reducción medida en bytes y lo que no se pudo medir por la pausa de pilotos.
- A5. V1 y V2 quedan resueltas sin contradicción con `:94`, `:105`, `:106` y `:143`.

## Decisiones del usuario (2026-09-26)

- **K1-B:** alcance C1 más C2, con el riesgo de carga de skills aceptado.
- **K2-A:** quitar el respaldo del workspace en `:42`.
- **K3-A:** incluir V2 en este cambio.
- **K4-A:** verificar la equivalencia con una revisión independiente, sin corridas de modelo.
- **Ruta:** `flow-plan` en esta sesión.

## Entrega

Respuesta del usuario (2026-09-26):
- **K6-A, `direct-base`:** un commit con las rutas de este cambio y su carpeta `openspec`, con push a `rebuild/harness-engineering`; después, el commit de cierre con el archivo del cambio, también con push. Sin PR.
- **K7-A, revisión:** solo la revisión de equivalencia K4-A (T4), con `review-refuter` sobre sus bloqueantes.
- **K8-A, despliegue:** release nueva a los seis hosts tras el push, desde un worktree limpio del commit, verificando el texto en los archivos de cada host y borrando el worktree después.
- **K5-A, test:** un test en `tests/content` comprueba que cada `skill:` citado en `global.md` existe (T5).

El árbol de trabajo tiene cambios ajenos sin commitear (`versioned-installer-onboarding`, `tooling/**`); la entrega commitea solo las rutas de este cambio.
