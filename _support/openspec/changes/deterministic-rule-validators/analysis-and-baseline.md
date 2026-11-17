# Análisis del video `O_zcc6Xl12g`, contraste con Hive y línea base numérica (Antes vs. Después)

Este documento conserva dentro de la carpeta del cambio el análisis completo del video `https://www.youtube.com/watch?v=O_zcc6Xl12g` (*All Code Is Slop (Except Mine) — Lauren Tan, SpaceXSI*, AI Engineer, 2026-11-14), su contraste con `tricell-hive` en el commit base `f886538f` (`development`), los hallazgos `H1`–`H12` y las métricas exactas de **Antes** y **Después** para cualquier agente o sesión posterior.

- **Transcripción completa (470 líneas):** [assets/video-O_zcc6Xl12g-transcript.txt](assets/video-O_zcc6Xl12g-transcript.txt)
- **Metadatos y capítulos del video:** [assets/video-O_zcc6Xl12g-metadata.json](assets/video-O_zcc6Xl12g-metadata.json)

---

## 1. Síntesis del video por capítulo y contraste con Hive

1. **Pirámide invertida de confianza (`0:00`–`4:59`):**
   - **Tesis del video:** el trabajo meta del ingeniero de software es mantener o subir la calidad del código mientras escala el volumen de PRs generados por agentes. Ordena la inversión de mayor a menor determinismo:
     1. **Arquitectura, diseño de APIs y estructura de directorios.**
     2. **Sistemas de tipos, compiladores, reglas de lint y tests.**
     3. **Verificación en runtime y verificación formal.**
     4. **Prosa (`AGENTS.md`, reglas, skills) únicamente para decisiones que requieren juicio.**
   - **Contraste con Hive:** Hive ya aplica este principio para su propio mantenimiento (`AGENTS.md:21`: *"move a rule with no exceptions into a test, hook or script when one can enforce it"*), pero varias reglas operativas en `content/guidance/global.md` y `content/skills/` siguen expresadas como recetas manuales de shell en prosa (`H6`–`H12`), y aún falta un flujo que codifique errores repetidos en el código de los proyectos cliente (`H2`).
2. **El código como jardín y arquitectura exigible (`5:00`–`6:55`):**
   - **Tesis del video:** el código existente es el input más grande del agente; cualquier workaround, código deprecado o mal patrón se propaga porque el siguiente agente lo copia como convención. En su framework `Dune` prohibieron `useEffect` en código de producto y separaron `electron-renderer` y `electron-main` prohibiendo imports cruzados con análisis de grafo de dependencias y lints.
   - **Contraste con Hive:** `flow-build/SKILL.md` («Reuse before writing») y `flow-build/references/defect-fix.md` prohíben copiar workarounds y exigen corregir la causa raíz, pero la detección no deja tras de sí una barrera estática (lint o test de frontera arquitectónica) en el repositorio intervenido (`H2`).
3. **Reemplazar comentarios con reglas de lint accionables (`6:56`–`9:39`):**
   - **Tesis del video:** los agentes escriben comentarios extensos para justificar por qué aplicaron un parche rápido (ej. `setTimeout` / `requestAnimationFrame` para tapar un choque arquitectónico entre código imperativo y declarativo). Prohibieron el 99% de los comentarios y los sustituyeron por reglas de lint cuyo mensaje de error le explica al agente cómo corregir la causa raíz.
   - **Contraste con Hive:** Hive no audita de forma específica comentarios-excusa ni directivas de supresión (`eslint-disable`, `@ts-expect-error`, `nolint`) introducidas en el diff (`H3`).
4. **Verificación en runtime, *feature maps* y CLIs propios (`9:40`–`10:49`):**
   - **Tesis del video:** para que el agente ejecute y depure la aplicación por sí mismo sin adivinar el vocabulario de la UI ni gastar tokens redescubriendo cómo levantarla, empaquetan por repositorio un skill de verificación con un **feature map** (mapa de funcionalidades desde la vista del usuario) y un CLI/driver reutilizable y testeable.
   - **Contraste con Hive:** Hive cuenta con `flow-verify/SKILL.md`, `flow-verify/references/ui-verification.md`, el visor local (`hive view`), `HIVE.local.md` y `ACCOUNTS.md`, pero no genera ni mantiene un *feature map* ni un arnés de conducción por proyecto (`H1`).
5. **Verificación formal, máquinas de estado y property-based UI testing (`10:50`–`13:31`):**
   - **Tesis del video:** usan **TLA+** para model checking de flujos multi-dispositivo (notificaciones), **Lean** para probar invariantes de algoritmos (virtualización de listas sin parpadeo), **Bombadil** para property-based testing de UI, y un visualizador de máquinas de estado (`statechart`) para revisar transiciones junto con videos y capturas. Reportan que ~70% de sus tokens diarios se destinan a verificación.
   - **Contraste con Hive:** Hive tiene disciplina de tests de comportamiento (`flow-build/SKILL.md#tests`), auditoría de tests con mutación (`test-audit/SKILL.md`) y verificación con evidencia de corrida (`flow-verify/SKILL.md`), pero carece de referencias o skills para verificación formal / máquinas de estado (`H4`) y property-based UI testing (`H5`).
6. **Fuzzing post-release y bucle externo (`13:32`–`17:31`):**
   - **Tesis del video:** al cortar cada release, disparan un proyecto de agentes que hace fuzzing de la aplicación compilada usándola como usuario real, abre PRs de corrección y los fusiona cuando los cimientos determinísticos lo permiten.
   - **Contraste con Hive:** Hive supera este enfoque en gobernanza y seguridad: separa la autorización humana de los efectos (`interactive`, `automatic`, `ready`, `direct-base`, `hold`), protege secretos (`0700`/`0600`), mantiene higiene de `_support/` y memoria selectiva en Engram, y es portable entre 7 hosts (`claude`, `codex`, `cursor`, `grok`, `opencode`, `pi`, `antigravity`).

---

## 2. Catálogo completo de hallazgos (`H1`–`H12`)

### Huecos de skills y capacidades (Fases 2 y 3 — fuera de este cambio)
- **H1** — Generador y mantenedor de arnés de verificación y *feature map* por proyecto (`Launch`, `Doctor`, `Drive`, `Evidence`, `Cleanup` + mapa de features/rutas).
- **H2** — Escalera determinística de corrección de errores repetidos (arquitectura → tipos → lint con mensaje accionable → test → prosa + tabla regla↔enforcer).
- **H3** — Auditoría de comentarios-excusa y supresiones (`eslint-disable`, `@ts-expect-error`, `nolint`) en `hive-review-code` y `flow-build`.
- **H4** — Verificación formal de máquinas de estado y concurrencia (TLA+ / Lean / SMT + diagrama de estados).
- **H5** — Exploración basada en propiedades / fuzzing de UI en `flow-verify`.

### Reglas actuales de Hive que pasan a métodos determinísticos en este cambio (Fase 1: `H6`–`H12`)
- **H6 (`T2`, `AC2`)** — Validador y gestor determinístico de `HIVE.local.md`, `ACCOUNTS.md` y `_support/secrets/` en `tooling/cli/doctor_project.go` y `content/skills/harness-audit/scripts/local_instructions.py`.
- **H7 (`T3`, `AC3`)** — Verificador determinístico de antigüedad de paquetes (`>= 7` días) y advisories en `content/skills/flow-build/scripts/check_package.py`.
- **H8 (`T4`, `AC4`)** — Subcomando determinístico `commit` por rutas explícitas con preservación de índice, escaneo de secretos y validación de `Unreviewed:` en `content/skills/git-workflow/scripts/git_state.py`.
- **H9 (`T5`, `AC5`)** — Planificador y ejecutor determinístico de limpieza de ramas, worktrees y procesos en `content/skills/flow-close/scripts/cleanup_branches.py`.
- **H10 (`T6`, `AC6`)** — Validador de gramática de `## Business rules` en `content/skills/flow-plan/scripts/check_change.py` y actualizador por marcadores en `content/skills/flow-plan/scripts/business_rules.py`.
- **H11 (`T6`, `AC7`)** — Validadores estáticos de *ready list* para borradores de PRD (`content/skills/create-prd/scripts/check_prd.py`) y tickets (`content/skills/create-issue/scripts/check_issue.py`).
- **H12 (`T1`, `AC1`)** — Validación de anclas `#fragment` contra encabezados Markdown reales en `tooling/management/references.go`.

---

## 3. Línea base numérica: Antes vs. Después (medido en `f886538f`)

### 3.1 Huella de lectura declarada (`go test -v ./tests/content -run Footprint`)

| Escenario (`reading_footprint_test.go`) | Baseline editorial A (`d5164f71`) | Antes (`f886538f` actual) | Después (meta tras `T7`) | Delta esperado |
| --- | ---: | ---: | ---: | --- |
| `global-only` (bytes / caracteres) | 46,912 B / 46,898 chars | **52,508 B** / **52,385 chars** | `< 52,508 B` (medir al cerrar `T7`) | Reducción neta en cada sesión y turno |
| `simple-query` (bytes / caracteres) | 46,912 B / 46,898 chars | **52,508 B** / **52,385 chars** | `< 52,508 B` (medir al cerrar `T7`) | Reducción neta en consultas simples |
| `question` (bytes / caracteres) | 46,912 B / 46,898 chars | **52,508 B** / **52,385 chars** | `< 52,508 B` (medir al cerrar `T7`) | Reducción neta en rondas de preguntas |
| `plan-build-cycle` unión (bytes / caracteres) | 107,021 B / 106,987 chars | **133,441 B** / **133,211 chars** | `< 133,441 B` (medir al cerrar `T7`) | Reducción acumulada en el ciclo completo |

### 3.2 Tamaño por archivo afectado (límite de `skill_size_test.go`: `16,000` bytes para skills y referencias)

| Archivo | Antes (`f886538f`) — Bytes | Antes — Runes | Antes — Líneas | Margen actual vs. 16,000 B | Después (se completa en `T7`) |
| --- | ---: | ---: | ---: | ---: | --- |
| `content/guidance/global.md` | **52,508 B** | 52,385 | 149 | n/a (dividido en 4 reglas ≤ 22 KB en Antigravity) | Pendiente de medir en `T7` (`< 52,508 B`) |
| `content/skills/flow-close/SKILL.md` | **15,993 B** | 15,963 | 87 | **7 B** (al límite) | Pendiente de medir en `T7` (`<= 16,000 B`) |
| `content/skills/git-workflow/SKILL.md` | **15,950 B** | 15,938 | 36 | **50 B** (al límite) | Pendiente de medir en `T7` (`<= 16,000 B`) |
| `content/skills/flow-build/SKILL.md` | **15,150 B** | 15,127 | 70 | 850 B | Pendiente de medir en `T7` (`<= 16,000 B`) |
| `content/skills/flow-plan/SKILL.md` | **14,099 B** | 14,083 | 58 | 1,901 B | Pendiente de medir en `T7` (`<= 16,000 B`) |
| `content/skills/harness-audit/references/local-instructions.md` | **9,706 B** | 9,668 | 74 | 6,294 B | Pendiente de medir en `T7` (`<= 16,000 B`) |
| `content/skills/harness-audit/SKILL.md` | **9,427 B** | 9,419 | 79 | 6,573 B | Pendiente de medir en `T7` (`<= 16,000 B`) |
| `content/skills/flow-plan/references/business-rules.md` | **9,186 B** | 9,169 | 80 | 6,814 B | Pendiente de medir en `T7` (`<= 16,000 B`) |
| `content/skills/create-issue/SKILL.md` | **6,641 B** | 6,631 | 18 | 9,359 B | Pendiente de medir en `T7` (`<= 16,000 B`) |
| `content/skills/create-prd/SKILL.md` | **6,386 B** | 6,380 | 22 | 9,614 B | Pendiente de medir en `T7` (`<= 16,000 B`) |
| `content/skills/flow-close/references/branch-cleanup.md` | **5,235 B** | 5,229 | 50 | 10,765 B | Pendiente de medir en `T7` (`<= 16,000 B`) |

### 3.3 Tamaño por sección de `content/guidance/global.md` (Antes)

| Sección de `content/guidance/global.md` | Antes — Bytes | Antes — Líneas | Impactada en Fase 1 |
| --- | ---: | ---: | --- |
| `# Tricell Hive guidance` (encabezado) | 268 B | 4 | No |
| `## Language` | 650 B | 6 | No |
| `## Scope and authorization` | **6,820 B** | 15 | **Sí (`H7`: check de paquete ≥ 7 días)** |
| `## Preservation` | **2,818 B** | 8 | **Sí (`H9`: limpieza al cierre)** |
| `## Secrets` | **1,574 B** | 9 | **Sí (`H6`, `H8`: exclusión y escáner antes de commit)** |
| `## Test accounts` | 2,048 B | 9 | Indirecto (`H6` valida ausencia de secretos en `ACCOUNTS.md`) |
| `## Local instructions` | **9,218 B** | 11 | **Sí (`H6`: exclusión, 200 chars, 50 líneas, `local_instructions.py`)** |
| `## Evidence` | 360 B | 5 | No |
| `## Proportionality` | 3,237 B | 7 | No |
| `## Support folder (_support)` | 809 B | 4 | No |
| `## Questions and routes` | 14,258 B | 20 | No |
| `## Delivery` | 4,426 B | 13 | No |
| `## Project settings` + bloque `## Hive` | 1,895 B | 20 | No |
| `## Persistent memory` | 2,174 B | 6 | No |
| `## Scope changes` | 1,953 B | 12 | No |
| **Total `content/guidance/global.md`** | **52,508 B** | **149** | **4 secciones adelgazadas** |

### 3.4 Cobertura determinística (Antes vs. Después)

| Indicador de cobertura determinística | Antes (`f886538f`) | Después (al completar `T1`–`T7`) | Delta |
| --- | ---: | ---: | ---: |
| Scripts Python determinísticos en `content/skills/*/scripts/*.py` | **11 scripts** (en 6 skills) | **17 scripts** (en 8 skills) | **+6 scripts nuevos** y **2 extendidos** |
| Enlaces Markdown internos con `#fragment` validados en `references.go` | **0 de 133** (0%, ignorados) | **133 de 133** (100% de enlaces con `#fragment` en `content/`) | **+133 anclas validadas** |
| Superficies de proyecto auditadas por `hive doctor --project` | **2** (`AGENTS.md` `## Hive`, `<Specs>/changes/`) | **6** (`AGENTS.md` `## Hive`, `<Specs>/changes/`, `HIVE.local.md`, `ACCOUNTS.md`, `_support/secrets/`, `_support/scripts/`) | **+4 superficies** |
| Reglas de `H6`–`H12` respaldadas por validador/script determinístico | **0 de 7** (100% dependientes del prompt) | **7 de 7** (100% respaldadas por código y tests) | **+7 reglas determinísticas** |
