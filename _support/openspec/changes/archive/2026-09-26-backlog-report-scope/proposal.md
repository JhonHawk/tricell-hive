# Alcance de la regla del backlog y tickets en líneas de lista

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `rebuild/harness-engineering` (`556c52b`) y desplegado como release `0db2adc637b6` |
| Tracker · GitHub Issues | Hueco del modelo de trazas: [#39 (juzgar el corte pedido)](https://github.com/JhonHawk/tricell-hive-private/issues/39) |
| Git | Commit `556c52b` con push a `rebuild/harness-engineering` |
| Verificación | `go vet ./...`, `go test -race ./...` y tests de skills en verde en un worktree limpio de `556c52b` · caso de regresión con reversión por rama · escaneo de fixtures vacío · prueba con Haiku |
| Siguiente paso | Ninguno en este cambio; la eficacia en Grok se observa en el monitoreo de sesiones |

## Objetivo

En la sesión de Grok 4.7 (high) `01a0dd4c` de ark (2026-09-26), el usuario pidió: "de los issues registrados, agrupa aquellos que puedan trabajarse en una sola sesión". La respuesta fue el reporte de backlog por módulo, en 8 párrafos que encadenaban de 4 a 8 tickets enlazados cada uno:

- **H1.** No respondió al corte pedido. La regla del backlog (`global.md`, Communication) no decía cuándo aplicaba, así que se disparó con cualquier respuesta de cinco tickets o más e impuso módulos, conteo por estado y estado de entrega. Los grupos reales quedaron enterrados.
- **H2.** Elementos paralelos en párrafos. Contradice "a list for parallel items" y también el prompt de sistema de Grok.
- **H3.** Descripciones telegráficas que no se entienden sin abrir el ticket.
- **H4.** Ruido de entrega (SHAs, PRs, conteo por estado) que no venía al caso.

## Alcance y aceptación

**Incluye:**
- `content/guidance/global.md`: la regla del backlog aplica a preguntas de estado o de listado del backlog. Otro corte estructura la respuesta y omite conteos y estado de entrega salvo que lo cambien. La regla de forma del chat pide un ticket, o un grupo que va junto, por línea de lista, con una descripción que se entienda sin abrir el ticket.
- `tests/content/budget_test.go`: presupuesto de 40635 a 41085 bytes (+450).
- El criterio determinista `ticket_ids_not_packed_in_prose`, con fixtures Grok derivados de la sesión.

**Excluye:**
- Un criterio para H1: exige entender la pregunta. Queda registrado en [#39 (juzgar el corte pedido)](https://github.com/JhonHawk/tricell-hive-private/issues/39).
- Endurecer la regla general de forma más allá de los tickets (R2, descartada por el usuario): un solo caso.
- Pilotos con modelos (pausados).

**Criterios de aceptación:**
- A1. La regla del backlog nombra su disparador (estado del proyecto o su backlog) y dice qué hacer ante otro corte.
- A2. La regla de forma pide una línea de lista por ticket o grupo, con descripción autosuficiente, y prohíbe encadenar tickets en un párrafo.
- A3. `ticket_ids_not_packed_in_prose` falla con el fixture derivado de `01a0dd4c` y pasa con su versión en lista. La reversión de sus ramas hace fallar el fixture que corresponde.
- A4. `go vet ./...` y `go test -race ./...` pasan, y el escaneo de fixtures no imprime nada.

## Decisiones del usuario (2026-09-26)

- **Ruta:** R1 con `flow-build`.
- **Entrega:** D1-A `direct-base`. Commits verificados directo en `rebuild/harness-engineering`, con push y un commit de cierre con el archivo del cambio.
- **Revisión:** D2-A, sin revisión dedicada. No toca contratos, seguridad ni datos, y los fixtures y la prueba de reversión prueban el criterio.
- **Despliegue:** D3-A, release nueva a los hosts globales tras el push.
- **Validación para modelos pequeños:** E1-A, aplicar las correcciones P1–P3 de la auditoría `prompt-audit` ([design.md](design.md#auditoría-para-modelos-pequeños-prompt-audit-2026-09-26)). E2-A, prueba mínima con Haiku: 2 corridas por variante, regla vieja y nueva (evidencia (historical evidence omitted from public history)).

El árbol de trabajo tiene cambios ajenos sin commitear (`versioned-installer-onboarding`, `tooling/**`, `VERSION`, `bootstrap.sh`). La entrega commitea solo las rutas de este cambio.
