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

## 3. Línea base numérica: Antes vs. Después (medido en `f886538f` / `e005e617` y tras `T7`)

### 3.1 Huella de lectura declarada (`go test -v ./tests/content -run Footprint`)

| Escenario (`reading_footprint_test.go`) | Baseline editorial A (`d5164f71`) | Antes (`f886538f` / `e005e617`) | Después (medido tras `T7`) | Delta vs. `f886538f` (`e005e617`) |
| --- | ---: | ---: | ---: | --- |
| `global-only` (bytes / caracteres) | 46,912 B / 46,898 chars | **52,508 B** / **52,385 chars** | **52,379 B** / **52,256 chars** | **-129 B (-129 chars)** en cada sesión |
| `simple-query` (bytes / caracteres) | 46,912 B / 46,898 chars | **52,508 B** / **52,385 chars** | **52,379 B** / **52,256 chars** | **-129 B (-129 chars)** |
| `question` (bytes / caracteres) | 46,912 B / 46,898 chars | **52,508 B** / **52,385 chars** | **52,379 B** / **52,256 chars** | **-129 B (-129 chars)** |
| `plan-build-cycle` unión (bytes / caracteres) | 107,021 B / 106,987 chars | **133,441 B** (`130,621 B` en `e005e617`) | **130,722 B** / **130,498 chars** | **-2,719 B vs. `f886538f`** (`+101 B` vs. `e005e617` por enlaces a scripts) |

### 3.2 Tamaño por archivo afectado (límite de `skill_size_test.go`: `16,000` bytes para skills y referencias)

| Archivo | Antes (`f886538f` / `e005e617`) — Bytes | Antes — Runes | Antes — Líneas | Después (`T7`) — Bytes / Runes / Líneas | Delta (Bytes) |
| --- | ---: | ---: | ---: | ---: | ---: |
| `content/guidance/global.md` | **52,508 B** | 52,385 | 149 | **52,379 B** / 52,256 runes / 149 líneas | **-129 B** |
| `content/skills/flow-close/SKILL.md` | **15,993 B** (`13,172 B` en `e005e617`) | 15,963 | 87 (`75` en `e005e617`) | **13,178 B** / 13,154 runes / 75 líneas | **-2,815 B vs. `f886538f`** (`+6 B` vs. `e005e617`) |
| `content/skills/git-workflow/SKILL.md` | **15,950 B** | 15,938 | 36 | **15,869 B** / 15,857 runes / 36 líneas | **-81 B** (margen sube de 50 B a 131 B) |
| `content/skills/flow-build/SKILL.md` | **15,150 B** | 15,127 | 70 | **15,265 B** / 15,242 runes / 70 líneas | **+115 B** (enlace a `check_package.py`) |
| `content/skills/flow-plan/SKILL.md` | **14,099 B** | 14,083 | 58 | **14,209 B** / 14,193 runes / 58 líneas | **+110 B** (alcance de `check_change.py`) |
| `content/skills/harness-audit/references/local-instructions.md` | **9,706 B** | 9,668 | 74 | **9,959 B** / 9,921 runes / 74 líneas | **+253 B** (enlace a `local_instructions.py`) |
| `content/skills/harness-audit/SKILL.md` | **9,427 B** | 9,419 | 79 | **9,521 B** / 9,513 runes / 79 líneas | **+94 B** (enlace a `local_instructions.py`) |
| `content/skills/flow-plan/references/business-rules.md` | **9,186 B** | 9,169 | 80 | **9,383 B** / 9,366 runes / 80 líneas | **+197 B** (enlace a `business_rules.py`) |
| `content/skills/create-issue/SKILL.md` | **6,641 B** | 6,631 | 18 | **6,814 B** / 6,804 runes / 18 líneas | **+173 B** (enlace a `check_issue.py`) |
| `content/skills/create-prd/SKILL.md` | **6,386 B** | 6,380 | 22 | **6,531 B** / 6,525 runes / 22 líneas | **+145 B** (enlace a `check_prd.py`) |
| `content/skills/flow-close/references/branch-cleanup.md` | **5,235 B** | 5,229 | 50 | **5,468 B** / 5,462 runes / 50 líneas | **+233 B** (enlace a `cleanup_branches.py`) |

### 3.3 Tamaño por sección de `content/guidance/global.md` (Antes vs. Después)

| Sección de `content/guidance/global.md` | Antes — Bytes | Antes — Líneas | Después (`T7`) — Bytes | Delta (Bytes) |
| --- | ---: | ---: | ---: | ---: |
| `# Tricell Hive guidance` (encabezado) | 268 B | 4 | 268 B | 0 B |
| `## Language` | 650 B | 6 | 650 B | 0 B |
| `## Scope and authorization` (`H7`: `check_package.py`) | **6,820 B** | 15 | **6,881 B** | +61 B (agrega link `skill:flow-build/scripts/check_package.py`) |
| `## Preservation` | **2,818 B** | 8 | **2,818 B** | 0 B |
| `## Secrets` (`H6`, `H8`: `local_instructions.py` + `git_state.py`) | **1,574 B** | 9 | **1,735 B** | +161 B (agrega links `skill:`) |
| `## Test accounts` | 2,048 B | 9 | 2,048 B | 0 B |
| `## Local instructions` (`H6`: `local_instructions.py`) | **9,218 B** | 11 | **8,867 B** | **-351 B** |
| `## Evidence` | 360 B | 5 | 360 B | 0 B |
| `## Proportionality` | 3,237 B | 7 | 3,237 B | 0 B |
| `## Support folder (_support)` | 809 B | 4 | 809 B | 0 B |
| `## Questions and routes` | 14,258 B | 20 | 14,258 B | 0 B |
| `## Delivery` | 4,426 B | 13 | 4,426 B | 0 B |
| `## Project settings` + bloque `## Hive` | 1,895 B | 20 | 1,895 B | 0 B |
| `## Persistent memory` | 2,174 B | 6 | 2,174 B | 0 B |
| `## Scope changes` | 1,953 B | 12 | 1,953 B | 0 B |
| **Total `content/guidance/global.md`** | **52,508 B** | **149** | **52,379 B** | **-129 B** |

### 3.4 Cobertura determinística (Antes vs. Después)

| Indicador de cobertura determinística | Antes (`f886538f`) | Después (medido tras `T1`–`T7`) | Delta |
| --- | ---: | ---: | ---: |
| Scripts Python determinísticos en `content/skills/*/scripts/*.py` | **11 scripts** (en 6 skills) | **17 scripts** (en 8 skills) | **+6 scripts nuevos** y **2 extendidos** |
| Enlaces Markdown internos con `#fragment` validados en `references.go` | **0 de 133** (0%, ignorados de 528 enlaces) | **134 de 134** (100% validados de 557 enlaces internos) | **+134 anclas validadas** |
| Superficies de proyecto auditadas por `hive doctor --project` | **2** (`AGENTS.md` `## Hive`, `<Specs>/changes/`) | **6** (`AGENTS.md` `## Hive`, `<Specs>/changes/`, `HIVE.local.md`, `ACCOUNTS.md`, `_support/secrets/`, `_support/scripts/`) | **+4 superficies** |
| Reglas de `H6`–`H12` respaldadas por validador/script determinístico | **0 de 7** (100% dependientes del prompt) | **7 de 7** (100% respaldadas por código y tests) | **+7 reglas determinísticas** |
