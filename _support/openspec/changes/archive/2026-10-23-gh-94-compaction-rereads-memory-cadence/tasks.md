# Tareas

Commit base: `ce6392ba` (rama de prueba, 2026-10-20).

Todas las tareas van en el hilo principal: son cambios de redacción en reglas distribuidas, el brief para delegarlas sería más largo que el cambio y T1 y T2 escriben las dos en `global.md`. Cada tarea cumple "Guidance structure" de `AGENTS.md`.

### T1 — Releer tras compactar (P1)

- [x] `global.md` (Proportionality) tiene la regla de P1.

**Closes:** AC1.

**Depends on:** ninguna.

**Locations:** `content/guidance/global.md` (viñeta de `:76`).

**Execution:** hilo principal.

**Test approach:** check: búsqueda y lectura.

**Changes:** según design.md, P1.

**Verification:**
- `rg -n -i "compact" content/guidance/global.md` encuentra la regla. En la base no hay ninguna coincidencia.
- Lectura de la regla:
  - relee solo la referencia cuya condición aplica a la siguiente acción;
  - no exime a `reporting.md` cuando aplica;
  - se limita al hilo que compactó;
  - relee el `SKILL.md` activo cuando sus pasos no están en el resumen;
  - no pide leer por secciones.

### T2 — Ritmo de escritura en memoria (P2)

- [x] `global.md` (Continuity) fija los momentos de escritura y su precedencia, y los pasos 3 y 5 de `flow-close` escriben en una pasada al cierre.

**Closes:** AC2, AC3.

**Depends on:** T1, porque las dos escriben en `global.md`.

**Locations:** `content/guidance/global.md` (después de `:91`), `content/skills/flow-close/SKILL.md` (pasos 3 y 5).

**Execution:** hilo principal.

**Test approach:** check: búsqueda y lectura.

**Changes:** según design.md, P2.

**Verification:**
- `rg -n "memory tools" content/guidance/global.md` encuentra la cláusula de precedencia. En la base no hay ninguna coincidencia.
- `rg -n "Each time you remove a worktree" content/skills/flow-close/SKILL.md` no devuelve nada.
- Lectura:
  - los cinco momentos de escritura, con el resumen condicionado y el cierre como punto durable;
  - la prohibición de escribir por tarea, hallazgo, arreglo o descubrimiento;
  - `:92` sigue igual;
  - el paso 3 y los conteos de recurrencia del paso 5 escriben en la pasada previa al reporte;
  - lo declinado o suprimido se guarda después de la respuesta del paso 7;
  - `questions.md:35` sigue siendo coherente.

### T3 — Entrada del CHANGELOG

- [x] `CHANGELOG.md`, `[Unreleased]` > Changed, describe P1 y P2 con las sesiones de origen y #94.

**Depends on:** T1, T2.

**Locations:** `CHANGELOG.md`.

**Execution:** hilo principal.

**Verification:** `rg -n "#94" CHANGELOG.md` encuentra la entrada.

## Verificación compartida

| Gate | Cuándo | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Pruebas de contenido | Candidato final | `go test ./tests/content/ ./tooling/management/` sin fallos | Parte de la implementación |
| Tamaño de `global.md` | Después de T1 y T2 | `wc -c content/guidance/global.md`, delta respecto de 31 550 bytes | Parte de la implementación |
| Sitio | Después de T3 | Desde `site/`: `pnpm build` sin errores, porque el sitio publica el CHANGELOG | Parte de la implementación |

Sin efecto de UI. Pilotos pausados. No se agregan casos de regresión.

## Revisión y progreso

- 2026-10-20: borrador con D18-A, D19-A y D20-A.
- 2026-10-20: revisión del plan sobre la versión inicial (un `hive-review-plan`, ~72 s, 98k tokens). Encontró un bloqueante y cuatro no bloqueantes, y apliqué todos tal como los propuso el revisor, sin otra ronda:
  - B1: las respuestas del paso 7 se guardan después de responder el usuario;
  - N1: el resumen de sesión queda condicionado y el cierre es el punto durable;
  - N2: se prohíbe escribir por hallazgo, arreglo o descubrimiento;
  - N3: se relee el `SKILL.md` activo;
  - N4: tamaño estimado de 700 a 900 bytes.

  Límites: no se confirmó cómo resume cada host al compactar, ni el texto del plugin dentro de las sesiones.
- 2026-10-20: T1 y T2 verificadas por `hive-verify-task` (AC1, AC2 y AC3 cumplidos). Cambios posteriores:
  - T1: la redacción pasa a "may no longer be in context".
  - T2: se agrega el momento "create or remove a worktree" por `git-workflow/SKILL.md:18`, un conflicto que el mapa de impacto no detectó y que la verificación encontró. Además, la escritura del cierre queda después del paso 5 y antes del reporte, aunque no se haya retirado ningún worktree.

  Las dos son aclaraciones que no cambian ningún criterio, así que no hubo re-verificación.
- 2026-10-20: gates compartidos.
  - `go test ./tests/content/ ./tooling/management/` pasó.
  - `pnpm build` del sitio pasó.
  - `global.md` quedó en 31 550 → 32 571 bytes (+1021, por encima de los 700–900 estimados por el quinto momento de escritura).
  - Implementador: hilo principal, Claude Opus 5.5, observado. Verificadores lanzados con `model: opus`, configurado y no observado.

- 2026-10-23: auditoría de sesiones entregada en la conversación. Se observó relectura focalizada tras compactación en Claude Hive, y guardados fuera de la cadencia en CAM y Facti. No se afirma eficacia general; el riesgo de precedencia del plugin permanece. La suite final compartida de Go/vet, 55 pruebas Python y check/lint/build del sitio pasó antes de publicar la rama. El usuario autorizó push/merge y PR posterior de cierre de ambos planes.

## Cierre del 2026-10-23

- Integrado por PR #96 (ajustes de validación y entrega), commit `ac146beb`. Verificación final: Go, vet, 55 pruebas Python, check/lint/build del sitio y diff sin errores; no hubo CI ni pilotos de modelos.
- Se conserva el riesgo de cumplimiento de modelos y de precedencia de memoria; la auditoría no certifica fiabilidad general. Los planes se archivan por D2 en el PR posterior de cierre.
- El contenido instalado corresponde a `ac146beb` en seis hosts; el visor cotidiano y el material previo no se tocaron.
