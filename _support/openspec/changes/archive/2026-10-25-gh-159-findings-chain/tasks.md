# Tareas

Base: `e438e14b` (2026-10-25). Rama `feat/gh-159-findings-chain`, worktree `.claude/worktrees/gh-159` (`automatic`, E3-A). Build iniciado el 2026-10-25.

**Convención de las comprobaciones.** Igual que en [gh-157](../gh-157-test-guidance/tasks.md): una frase por cláusula, verificada en 0 antes de escribirla, más lectura del verificador.

## T1 — Mover el paso 5 de `flow-close` a una referencia

- [x] `content/skills/flow-close/references/incidental-findings.md` contiene el texto actual del paso 5 con el mismo sentido, solo con los enlaces relativos reescritos (`../flow-plan/…` → `../../flow-plan/…`, «step 3» → `../SKILL.md`); el paso 5 queda como una línea que nombra la regla y remite a la referencia.

**Depends on:** ninguna.
**Locations:** `content/skills/flow-close/SKILL.md:26`; referencia nueva; `_support/docs/architecture/instruction-resources.md` (agregar la referencia a la lista de hogares canónicos, porque pasa a ser destino entre skills: `ui-gate.md` y `ui-review-criteria.md` la enlazan en T2).
**Execution:** hilo principal: movimiento acoplado a T2.
**Test approach:** characterization: mismo sentido, solo enlaces reescritos.
Verificado por el orquestador (characterization): el texto movido coincide salvo tres enlaces reescritos; `flow-close/SKILL.md` pasa de 15 857 a 14 905 bytes; `go test ./tests/content/...` y `TestReleaseLinksResolveWithinPayload` pasan.

**Verification:** `diff` del párrafo movido contra la base con diferencias solo en los enlaces; `wc -c content/skills/flow-close/SKILL.md` ≤ 15 000; `go test ./tests/content/...` y `go test ./tooling/management -run TestReleaseLinksResolveWithinPayload` pasan.

## T2 — Qué hallazgos pueden volverse ticket y cómo se agrupan

- [x] La referencia aplica AC1, AC2 y AC8 y nombra los elementos de `global.md:32` que estrecha; `ui-gate.md:26` y `ui-review-criteria.md:69` remiten a ella.
Verificado por `hive-verify-task`: AC1, AC2 y AC8 cumplidos (`incidental-findings.md:11-20`, punteros en `ui-gate.md:26` y `ui-review-criteria.md:69`). En el mismo commit, el orquestador actualizó `instruction-resources.md:18` para nombrar los dos punteros nuevos, una observación del verificador. Observación sin bloqueo: las definiciones P0–P3 se copian de `hive-review-code.md` en vez de enlazarse.

**Closes:** AC1, AC2, AC8.
**Depends on:** T1.
**Locations:** `content/skills/flow-close/references/incidental-findings.md`; `content/skills/flow-walkthrough/references/ui-gate.md:26`; `content/skills/flow-walkthrough/references/ui-review-criteria.md:69`.
**Execution:** hilo principal.
**Test approach:** check: una frase por cláusula (P3 con enlace a `verification.md:43`, P0–P3 obligatorio, agrupar P2, duplicados por grupo, respuesta por miembro, confirmación renderizada) y lectura.
**Verification:** cada frase da 1; los dos punteros dan 1; `go test ./tests/content/...` y `TestReleaseLinksResolveWithinPayload` pasan.

## T3 — La pregunta de cierre no encadena tickets recién creados

- [x] La pregunta de cierre incorpora AC3 y ajusta el ejemplo de la línea 38.
Verificado por `hive-verify-task`: AC3 cumplido (`flow-close/SKILL.md:38`); el archivo pesa 15 071 bytes, dentro del límite de 16 000. El tope de 15 000 medía solo el movimiento de T1.

**Closes:** AC3.
**Depends on:** T1.
**Locations:** `content/skills/flow-close/SKILL.md:38-42`.
**Execution:** hilo principal.
**Test approach:** check: conteo de frase; tamaño ≤ 16 000 bytes.
**Verification:** la frase de AC3 da 1; «a defect found» ya no aparece sin la excepción; `go test ./tests/content/...` pasa.

## T4 — Severidad y límites de la revisión de UX

- [x] `ui-gate.md` lleva la tabla de AC4; `ui-review-criteria.md` y `ui-gate.md` llevan el alcance y la línea base de AC5, con la excepción nombrada; `verification.md:85` aplica AC5 al brief desplegado.
Verificado por `hive-verify-task`: AC4 y AC5 cumplidos (`ui-gate.md:28`, `ui-review-criteria.md:9,14,69`, `verification.md:85`). En el mismo commit, el orquestador agregó a las vistas tocadas «the screens the change makes reachable», un hueco que señaló el verificador frente a la regla de rutas que pasan a ser alcanzables (`verification.md:43`). La traducción de severidad quedó en prosa, no en tabla.

**Closes:** AC4, AC5.
**Depends on:** T2 (mismas líneas).
**Locations:** `content/skills/flow-walkthrough/references/ui-gate.md:23,26`; `content/skills/flow-walkthrough/references/ui-review-criteria.md:9,14,31,58-69`; `content/skills/flow-build/references/verification.md:85` (edición obligatoria).
**Execution:** hilo principal.
**Test approach:** check: frases de la tabla, del mínimo WCAG, de «solo para previos», de las vistas tocadas, del `git diff` vacío y del puntero en `verification.md:85`.
**Verification:** cada frase da 1; lectura de que la regla Blocker/High para introducidos no cambia; `go test ./tests/content/...` pasa.

## T5 — Umbral por rol y arreglos que crean superficie

- [x] `verification.md:70` aplica AC6; `flow-plan/SKILL.md` y `plan-revision.md` aplican AC7.
Verificado por `hive-verify-task`: AC6 y AC7 cumplidos (`verification.md:70`, `flow-plan/SKILL.md:16`, `plan-revision.md:11`); `go test ./tests/content/...` pasa.

**Closes:** AC6, AC7.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-build/references/verification.md:70`; `content/skills/flow-plan/SKILL.md` («Ground the decisions»); `content/skills/flow-plan/references/plan-revision.md`.
**Execution:** hilo principal.
**Test approach:** check: conteo de frases; `flow-plan/SKILL.md` ≤ 16 000 bytes.
**Verification:** las frases dan 1; `go test ./tests/content/...` pasa.

## T6 — Páginas del sitio

- [x] `close.md` y `build.md` describen AC1–AC5 y AC8.
Verificado por `hive-verify-task` en la ronda 2: AC9 cumplido (`close.md:47,49`, `build.md:37`). `astro check` no corrió: necesita una instalación no autorizada.
Ronda 1, AC9: `close.md:47` llama «sin confirmar» a un P3, y `build.md:37` dice que todo hallazgo introducido se arregla, cuando la skill solo exige arreglar Blocker y High introducidos o empeorados.

**Closes:** AC9.
**Depends on:** T2, T3, T4.
**Locations:** `site/src/content/docs/flows/close.md` paso 5 (línea 47) y paso 7 (línea 49); `site/src/content/docs/flows/build.md`, sección de revisión de UI.
**Execution:** delegado a `hive-write-docs`: fuentes y destino fijos.
**Test approach:** check: una frase por criterio en su archivo y paso, y lectura.
**Verification:** en `close.md` paso 5: P3, agrupar y renderizado (tres conteos ≥ 1); en el paso 7: no recomendar el ticket recién creado (≥ 1); en `build.md`: escala P0–P3, vistas tocadas y línea base local (tres conteos ≥ 1).

## Verificación compartida

| Comprobación | Cuándo | Mecanismo | Autoridad |
| --- | --- | --- | --- |
| `go vet ./...` y `go test ./...` | candidato final | salida limpia (incluye `TestReleaseLinksResolveWithinPayload`) | rutina de implementación |
| `/code-review` | candidato antes del merge | hallazgos verificados por la herramienta | `Review: /code-review` (E4-A) |
| Verificación por tarea | tras cada tarea | `hive-verify-task` | rutina de implementación |
| Lugares revisados y tamaño de `global.md` en el PR | al abrir el PR | lista de `AGENTS.md:24` | regla del repositorio |

Sin efecto en UI ni en tiempo de ejecución.

## Revisión y progreso

Revisión del plan (2026-10-25, un `hive-review-plan`, ~254 s, ~143k tokens): tres bloqueantes resueltos. B1: verificación del sitio por criterio. B2: AC4 solo para los previos, con tabla y mínimo WCAG. B3: AC5 con vistas definidas, «mismo árbol» por `git diff`, excepción nombrada y `verification.md:85` obligatorio. Aplicados S1–S6 y O1–O4; S6 deja AC4 como corrección pendiente de aprobación (E9).

Siguiente paso: `flow-build` (R1-A), primero de la secuencia gh-159 → gh-158 y gh-160.
