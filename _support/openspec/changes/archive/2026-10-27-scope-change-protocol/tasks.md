# Tareas

Commit base: `840807d9` (rama `feat/scope-change-protocol`, build iniciado el 2026-10-27).

`REF` abrevia `content/skills/flow-plan/references/scope-changes.md`.

### T1 — Referencia canónica de cambios de alcance

- [x] `REF` escrita según el diseño y registrada como hogar canónico.

**Closes:** AC1, AC2, AC3, AC13

**Checkpoint:** C1

**Depends on:** none

**Locations:** `REF` (nuevo), `_support/docs/architecture/instruction-resources.md:15-23`

**Execution:** hilo principal: la redacción fija el vocabulario que usan T2–T6.

**Test approach:** check — `rg` sobre el archivo y `go test ./tests/content`.

**Changes:** ver «Referencia canónica» en el diseño; en inglés; `## When to read` primero; ≤ 16 000 bytes.

**Verification:**
- `rg -n "^## When to read" REF` → 1 línea.
- `rg -n -i "correction" REF`, `rg -n -i "added task" REF`, `rg -n -i "new change" REF`, `rg -n -i "separate" REF` → al menos 1 línea cada uno (AC1).
- `rg -n -i "another ticket|another repository" REF` y `rg -n -i "stored data" REF` → al menos 1 línea cada uno (AC2).
- `rg -n -i "third" REF` → la propuesta de volver al plan (AC3).
- `rg -n "scope-changes.md" _support/docs/architecture/instruction-resources.md` → 1 línea que también contiene «before» o «when» con su condición (AC13).
- `wc -c REF` ≤ 16000; `go test ./tests/content` pasa.

### T2 — Punteros en guía global, plan y build

- [x] `global.md:21`, `plan-revision.md:11-12`, `flow-plan/SKILL.md:22` y `flow-build/SKILL.md:20` apuntan a la referencia sin repetir sus pasos.

**Closes:** AC2, AC4, AC5

**Checkpoint:** C2

**Depends on:** T1

**Locations:** `content/guidance/global.md:21`, `content/skills/flow-plan/references/plan-revision.md:11-12`, `content/skills/flow-plan/SKILL.md:22`, `content/skills/flow-build/SKILL.md:20`

**Execution:** hilo principal (texto acoplado a T1).

**Test approach:** check — `rg` y `go test ./tests/content`.

**Changes:** ver «Punteros» en el diseño; `flow-build` ≤ 16 000 bytes tras el cambio.

**Verification:**
- `rg -c "scope-changes.md" content/guidance/global.md content/skills/flow-plan/references/plan-revision.md content/skills/flow-plan/SKILL.md content/skills/flow-build/SKILL.md` → ≥ 1 en cada uno (AC2, AC4, AC5).
- `sed -n 21p content/guidance/global.md | rg -c "a correction"` → 1 y `sed -n 21p content/guidance/global.md | rg -c "an added task or a new change"` → 1 (AC4).
- `rg -n -i "a correction changes how the result" content/skills/flow-plan/references/plan-revision.md` → nada (AC5).
- `wc -c content/skills/flow-build/SKILL.md` ≤ 16000; `go test ./tests/content` pasa.

### T3 — Verify y primera vista

- [x] `flow-verify` clasifica cada pedido de la pausa, muestra el diff en rondas sin efecto visual y remite a la tercera ronda; `flow-walkthrough` define las rondas tras rechazar la primera vista y captura vistas solo en rondas con efecto visual.

**Closes:** AC6, AC7

**Checkpoint:** C2

**Depends on:** T1

**Locations:** `content/skills/flow-verify/SKILL.md:55-59`, `content/skills/flow-walkthrough/SKILL.md` («First rendered view», ~22-24, y `:52`)

**Execution:** hilo principal.

**Test approach:** check — `rg` y `go test ./tests/content`.

**Changes:** ver «Punteros» en el diseño.

**Verification:**
- `rg -n "scope-changes.md" content/skills/flow-verify/SKILL.md content/skills/flow-walkthrough/SKILL.md` → al menos 1 línea en cada uno (AC6, AC7).
- `rg -n -i "diff" content/skills/flow-verify/SKILL.md` → la ronda sin efecto visual (AC6).
- `rg -n -i "third" content/skills/flow-walkthrough/SKILL.md` o una línea que diga que las rondas tras el rechazo cuentan para la propuesta de la referencia (AC7).
- `rg -n "with a visual effect" content/skills/flow-walkthrough/SKILL.md` → la línea 52 corregida.
- `go test ./tests/content` pasa.

### T4 — Close y deploy

- [x] `flow-close` y `flow-deploy` apuntan a la referencia para cambios después de integrar y durante una promoción.

**Closes:** AC8, AC9

**Checkpoint:** C2

**Depends on:** T1

**Locations:** `content/skills/flow-close/SKILL.md` (cerca de la pregunta de cierre, ~41), `content/skills/flow-deploy/SKILL.md:8` («Entry and exit»), `REF` (línea de deploy, solo si T1 la dejó incompleta)

**Execution:** hilo principal.

**Test approach:** check — `rg` y `go test ./tests/content`.

**Changes:** ver «Punteros» en el diseño; `flow-close` ≤ 16 000 bytes tras el cambio.

**Verification:**
- `rg -n "scope-changes.md" content/skills/flow-close/SKILL.md content/skills/flow-deploy/SKILL.md` → al menos 1 línea en cada uno (AC8, AC9).
- `rg -n -i "integrat" REF` → la línea que dice que el registro integrado cierra normalmente y lo nuevo abre otro cambio (AC8).
- `rg -n -i "safe point" REF content/skills/flow-deploy/SKILL.md` → al menos 1 línea (AC9).
- `wc -c content/skills/flow-close/SKILL.md` ≤ 16000; `go test ./tests/content` pasa.

### T5 — Invariante mínima de la delegación desatendida

- [x] `global.md:36` queda como invariante mínima con puntero; la lista de mensajes que la mantienen vive solo en `unattended-delegation/SKILL.md`.

**Closes:** AC10

**Checkpoint:** C2

**Depends on:** none

**Locations:** `content/guidance/global.md:36`, `content/skills/unattended-delegation/SKILL.md:8` (solo lectura, para confirmar que conserva lo que sale de la guía global)

**Execution:** hilo principal.

**Test approach:** check — `rg` y `go test ./...`.

**Changes:** ver «`global.md:36`» en el diseño; comprobar que `global.md:20` sigue teniendo su referencia.

**Verification:**
- `sed -n 36p content/guidance/global.md | rg -c "Only these keep it|fits both lists"` → 0.
- `sed -n 36p content/guidance/global.md | rg -c "leaving"`, `… | rg -c "declared job"`, `… | rg -c "explicit delegation"`, `… | rg -c "route question"` → 1 cada uno (las cuatro fuentes).
- `sed -n 36p content/guidance/global.md | rg -c "changes its objective, scope or an answer in force"` → 1 (vencimiento) y `… | rg -c "irreversible"` → 1.
- `rg -n "Only these keep it|fits both lists" content/skills/unattended-delegation/SKILL.md` → ambas.
- `go test ./...` pasa.

### T6 — Diagrama, sitio y documentación

- [x] README con el diagrama nuevo; página `flows/overview` en el sitio y su menú; páginas de flujo y de inicio enlazadas; `llms.txt` actualizado; delta de `documentation-site` coherente.

**Closes:** AC11, AC12

**Checkpoint:** C2

**Depends on:** T1, T2, T3, T4, T5

**Locations:** `README.md:312-337`, `site/src/content/docs/flows/overview.md` (nuevo), `site/astro.config.mjs:74`, `site/src/content/docs/flows/{research,plan,build,verify,close,deploy}.md`, `site/src/content/docs/index.mdx` («The main flow»), `site/src/content/docs/skills.md` («Main flows»), `llms.txt`

**Execution:** delegado a `hive-write-docs`: las fuentes ya están verificadas y la escritura no se solapa con T1–T5, que habrán terminado. El brief nombra `REF`, los `SKILL.md` editados en T1–T5, el diseño y las rutas exactas que puede modificar.

**Test approach:** check — build del sitio y `rg` sobre el README.

**Changes:** ver «Diagrama y sitio» en el diseño; ninguna página copia texto de reglas.

**Verification:**
- En `site/`, `pnpm build` pasa y existe `dist/flows/overview/index.html` (AC12).
- `rg -n "flows/overview" site/astro.config.mjs site/src/content/docs/index.mdx site/src/content/docs/skills.md` → al menos 1 línea en cada uno.
- `rg -n "flows/overview" site/src/content/docs/flows/close.md site/src/content/docs/flows/deploy.md` → 1 línea en cada uno.
- En el bloque mermaid del README: `rg -n 'General rules' README.md`, `rg -n 'Change requested' README.md`, `rg -n 'Verify -\.->.*Build' README.md`, `rg -n 'Build -\.->.*Plan' README.md`, `rg -n 'Close -\.->.*(Build|Plan)' README.md`, `rg -n 'Deploy -\.->.*Build' README.md` → al menos 1 línea cada uno (AC11).
- `rg -n "scope-changes.md" README.md` → el enlace en prosa debajo del diagrama.
- Renderizado: ver «Diagrama y sitio» en el diseño.

## Verificación y revisión humana

- C1 — T1 — fija el vocabulario del que dependen T2–T6.
- C2 — T2, T3, T4, T5, T6 — cambios de texto que no cambian contratos; T6 es la última tarea.

| Puerta | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Pruebas de contenido | En cada tarea | `go test ./tests/content` pasa | Incluida en la implementación |
| Suite completa | Candidato final, local | `go test ./...` y `go vet ./...` pasan | Incluida |
| Build del sitio | T6 y candidato final | `pnpm build` en `site/` pasa | Incluida |
| Revisión de código | Candidato local, antes de la validación | `/code-review` de Claude Code; hallazgos corregidos | D10-A |
| Validación humana | Alto antes del push | Diff, diagrama renderizado (página en scratch o, si no procede, vista del PR antes del merge) y página del panorama | D9-A |

Verificación local en vivo: no aplica; no hay comportamiento en ejecución, solo texto de reglas y documentación. Ninguna prueba demuestra que un modelo siga la redacción; los pilotos siguen pausados.

## Estado de la revisión y progreso

- **Revisión del plan sobre `3a7ec381cdb1`** (2026-10-27), con dos `hive-review-plan`: coherencia de la guía (3 bloqueantes y 6 no bloqueantes) y sitio/README (2 bloqueantes y 4 no bloqueantes). Todos se aplicaron tal como los propuso el revisor, así que no hubo re-revisión. El coste de cada revisor no lo reporta el host.
  - B1: etiquetas de `global.md:21` sin solaparse.
  - B2: verificaciones por criterio.
  - B3 y sitio-1: `flows/close.md`, `flows/deploy.md` y `flow-walkthrough:52`.
  - Sitio-2: renderizado del diagrama antes del push.
  - N1: el cambio nuevo es solo «cambia el resultado».
  - N2: delegación desatendida.
  - N3: el vencimiento se queda en `global.md:36`.
  - N4: punteros en «dónde vuelve».
  - N5: punteros en `flow-build:20` y `flow-plan:22`.
  - N6: verificación de T5.
  - Sitio-3 a 6: panel con todas las secciones, patrones exactos, etiqueta del menú y enlaces de inicio, y `llms.txt`.
- **Decisiones tomadas sin preguntar:**
  - El diagrama se renderiza en una página de scratch con mermaid fijado desde CDN, con la vista del PR como respaldo. Motivo: no hay renderizador local instalado.
  - La etiqueta del menú es «Changes and loops». Motivo: evitar dos «Overview».
  - Ambas se revierten editando este archivo.
- **Decisiones en vigor:** D1-A, D2-A, D3-A, D4-A, D6-B, D8-B (con el ajuste sobre `global.md:72`), D9-A, D10-A, D11-A (PR #234 integrado en `840807d9` e instalación local refrescada).
- **Ruta elegida (2026-10-27):** D13-A (construir ahora); D12-A (la pregunta de revisión va como cambio pequeño aparte, después de este).

### Build (2026-10-27)

- **Commits:** T1 `51763fdc`, T5 `e9fbdc31`, T2 `e35e0703`, T3 `105ea38b` + `69a40a1c`, T4 `f08468fa`, T6 `e662b1a1` + `d390f2d6` + `1ed6ab4f`.
- **Verificación por tarea** (`hive-verify-task`, contexto nuevo sin historial):
  - T1 met (AC1, AC2 en su parte, AC3, AC13).
  - T2 met (AC2, AC4, AC5).
  - T3 met (AC7). AC6 met tras re-verificar: la elección entre diff y capturas pasó de depender del candidato a depender de cada corrección (`69a40a1c`).
  - T4 met (AC8, AC9).
  - T5 met (AC10). La comprobación literal «Only these keep it» del plan no coincide con la skill, que dice «Only these messages … keep the delegation». El contenido está; es un fallo de redacción del plan.
  - T6 met (AC11, AC12) tras una ronda: el panel de reglas pasó a un solo nodo, la vuelta Verify→Build se etiquetó «added task» porque una corrección se queda en Verify (desviación de la etiqueta «corrections» del diseño, por coherencia con la referencia), la primera vista se movió a Build en la tabla y se alineó la prosa del README.
- **Modelos:** implementador, hilo principal en Claude Opus 5.5 (observado). Verificadores y redactor de T6: perfil configurado de sus roles, no observado.
- **Renderizado del diagrama:** mermaid 11.17.2 desde jsDelivr en `agent-browser`; resultado «ok» con un SVG. Comprobación de dependencia: versión de más de 7 días (cooldown de pnpm: 10 080 min) y OSV sin avisos.
- **Puertas:** `go test ./...` y `go vet ./...` pasan sobre el candidato; `pnpm build` en `site/` pasa con 14 páginas y los avisos de siempre (`i18n` vacío, entrada `404`).
- **`/code-review` (medium, 2026-10-27) sobre `840807d9..1ed6ab4f`:** 10 hallazgos.
  - Corregidos los cuatro P1: en research una tarea añadida no autoriza implementar; la guía global vuelve a esperar a resolver el alcance antes del trabajo dependiente; quitar un ticket integrado de una promoción pide decidir si se revierte; después de integrar, una tarea añadida es trabajo propio.
  - Corregidos los cuatro P2: el puntero de la delegación desatendida ya no repite a medias su skill; `global.md:36` dice que cualquier mensaje no listado termina la delegación; hay punteros en `flow-research` y en «When to run it» de `flow-close`; las rondas se cuentan por etapa.
  - Descartado uno: un plan guardado absorbe crecimiento de alcance como tarea añadida. Es D4-A, y `plan-revision.md` mantiene la re-revisión y el aviso al usuario.
  - Reportado sin corregir un P3: el mapeo de etiquetas aparece en tres lugares.
  - Commits: T1 `b0e4a66a`, T2 `729148b9`, T4 `ee53342d`, T5 `03e8f488`, T6 `c92e6a7d`. Reabiertas T1, T2, T4, T5 y T6 por el arreglo después de su commit.
- **Re-revisión `/code-review` (medium) sobre `1ed6ab4f..c92e6a7d`:** 10 hallazgos.
  - Corregidos sin otra revisión, por ser locales y sin cambiar contratos (`7b124b3b`, `8103f6a7`, `c69ebbb0`): después de integrar, un cambio nuevo sigue la clase 3; una corrección es un seguimiento bajo la misma respuesta de entrega (cubre `direct-base`); una tarea añadida o un cambio nuevo terminan la delegación desatendida; qué resuelve el trabajo retenido; mantener un ticket en una promoción deja seguir la corrida; filas Research y Close del sitio; puntero de la pregunta de cierre.
  - Descartado: la excepción de la ronda de activación ya existía en la base y vive en la skill (evidencia del verificador de T5).
  - P3 reportado sin corregir: el puntero de `global.md:21` repite en mínimo la regla de retener el trabajo dependiente.
- **Re-verificación:** T1, T2, T4, T5 y T6 met tras sus correcciones. AC9 se reescribió para reflejar la corrección de la promoción; está pendiente de tu aprobación en la validación.
- **Detalles menores sin corregir:** una corrección después de integrar es un seguimiento, aunque la clase 1 diga que se queda en la etapa; la referencia no dice si la corrida espera mientras se responde mantener o revertir; la fila Research del sitio no menciona la parada de la ruta iterativa.
- **Puertas finales:** `go test ./...`, `go vet ./...`, `git diff --check` y `pnpm build` pasan sobre `c69ebbb0`. La guía global queda en 41 560 bytes (−67 respecto de la base).
- **Siguiente paso:** tu validación antes del push.
