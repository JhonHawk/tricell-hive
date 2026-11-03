# Tareas

Base: `development` en `5fa27f76`. Se construye en el checkout actual, rama `fix/post-0.7.0-corrections`, que ya lleva `26883dcc`; los dos trabajos salen en un solo PR (D9 B).

Todas las tareas corren en el hilo principal: son ediciones de texto chicas sobre archivos que se leen juntos, y delegar una costaría más que hacerla. Las referencias a `flow-close` se dan por su texto, porque sus números de línea difieren entre `development` y `26883dcc`.

### T1 — Modo `ready` en la guía global

- [x] «Delivery» define `ready` con sus efectos, y la recomendación de revisión lo incluye.

**Closes:** AC1, AC8

**Depends on:** ninguna

**Locations:** `content/guidance/global.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** una viñeta nueva tras `automatic`, y «always under `automatic` or `ready`» en la pregunta de revisión. No cambia la regla de commits ajenos ni la lista de claves de `## Hive`.

**Verification:**
- `rg -n '`ready`' content/guidance/global.md` muestra la viñeta, con «non-draft» y «never merge, enable auto-merge or add it to a merge queue», y la recomendación de revisión.
- `go test ./tests/content/...` pasa.
- Se anota la diferencia de bytes de `global.md`.

### T2 — Revisores y bots, automatización de merge, base al día y REST en `git-workflow`

- [x] El paso 6 lee la automatización de merge.
- [x] El paso 7 trata revisores humanos y bots, con su filtro de frescura y su ronda única.
- [x] El paso 8 vale para la parada de `ready` y no la frena un bot pendiente.
- [x] El paso 9 lee y espera por REST cuando GraphQL se agota.

**Closes:** AC3, AC4, AC5, AC6, AC11, AC12, AC14

**Depends on:** ninguna

**Locations:** `content/skills/git-workflow/SKILL.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** amplía los pasos 6 a 9 como fija `design.md` «Forma propuesta». No cambia:
- la regla del merge admin;
- «never poll for comments»;
- la regla de force-push;
- el texto de los hilos humanos, salvo el título del paso.

**Verification:** releer los pasos 6 a 9 contra AC3–AC6, AC11, AC12 y AC14, con una frase observable por criterio. Cada criterio tiene su comprobación:
- AC4: el paso 7 nombra `commit_id`, la hora del push y la excepción de los hilos humanos.
- AC5: el paso 8 cita «not ready».
- AC6: el paso 9 nombra los cinco endpoints REST y `gh run watch`.
- AC11: el paso 7 nombra `Unreviewed:` y remite a `flow-verify` paso 3.
- AC12: el paso 6 incluye «merge» en la lista de automatización.
- AC14: el paso 7 trata los commits ajenos en la rama del PR.

Además, `go test ./tests/content/...` pasa (tamaño ≤ 16.000 bytes).

### T3 — Fin de la integración bajo `ready`

- [x] `flow-verify` paso 6 se detiene sin merge bajo `ready`, después de la comprobación de base, y apunta a `git-workflow` para las revisiones.

**Closes:** AC2

**Depends on:** T2 (los pasos a los que apunta)

**Locations:** `content/skills/flow-verify/SKILL.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** solo el paso 6. No repite el reporte, que vive en `flow-close`. La espera bloqueante del CI y el diagnóstico desde el log quedan igual.

**Verification:**
- El paso 6 nombra la parada bajo `ready` y remite a `git-workflow` pasos 7 y 8.
- `rg -n 'auto-merge|merge queue' content/skills/flow-verify/SKILL.md content/guidance/global.md` muestra la prohibición solo en la guía global.
- `go test ./tests/content/...` pasa.

### T4 — Punto de cierre, visor y reporte de `ready` en `flow-close`

- [x] El punto de cierre incluye `ready`.
- [x] El paso 1 deja el visor en `in_review` en ese caso.
- [x] El reporte final nombra el PR, el commit de cabeza, el commit de base comprobado, el CI y los bots pendientes.

**Closes:** AC7, AC13, AC2 (el reporte de la parada), AC5 (el commit de base en el reporte)

**Depends on:** ninguna

**Locations:** `content/skills/flow-close/SKILL.md`, y `content/skills/flow-walkthrough/references/viewer-publication.md` solo si su sección «Close and input hygiene» contradice `in_review`

**Execution:** hilo principal

**Test approach:** check

**Changes:** la frase del punto de cierre, el paso 1 y las viñetas **Integrated** y **Next** del reporte final.

**Verification:**
- `rg -n 'ready' content/skills/flow-close/SKILL.md` muestra las tres partes.
- `hive view contract --operation session_patch` lista `in_review` como estado aceptado.

### T5 — Sitio, puntero de la documentación y changelog

- [x] `verify.md` y `close.md` describen la parada de `ready`.
- [x] `CHANGELOG.md` tiene la entrada en `[Unreleased]`.

**Closes:** AC9

**Depends on:** T1, T3, T4 (el comportamiento que describen)

**Locations:**
- `site/src/content/docs/flows/verify.md`
- `site/src/content/docs/flows/close.md`
- `CHANGELOG.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:**
- Las páginas describen el comportamiento y enlazan la skill, sin copiar sus reglas.
- La entrada del changelog va en `Added`; corresponde subir la versión menor en el próximo release, porque es una opción nueva que el usuario elige.

**Verification:**
- `npm run build` en `site/` pasa.

### T6 — Hogares canónicos al día en `instruction-resources.md`

- [x] Cada entrada de «Runtime resolution» apunta al hogar actual de su regla, o sale de la lista cuando la regla ya no existe; el párrafo de `hive-core` se reescribe o se quita.

**Closes:** AC10

**Depends on:** T1, T2, T4 (la entrada de entrega apunta a la «Delivery» ya editada)

**Locations:** `_support/docs/architecture/instruction-resources.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** solo la sección «Runtime resolution» (líneas 13 a 39 hoy). Para cada regla, se busca con `rg` su texto en `content/` y se apunta a la skill, la referencia o la sección de `global.md` que la tiene. No cambia las secciones de autoría, validación de release ni el cargador de OpenCode.

**Verification:** un recorrido de las rutas `content/...` y `<skill>/references|assets/...` que nombra el documento, con `test -e` para cada una, no deja ninguna sin resolver; `rg -n 'hive-core' _support/docs/architecture/instruction-resources.md` solo queda si nombra algo que existe.

## Verificación y revisión humana

| Chequeo | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar cada tarea, en local | El comando de la tarea pasa | Cubierta por la implementación |
| Suite completa | Candidato final, en local | `go test ./...` | Cubierta por la implementación |
| Build del sitio | Candidato final, en local | `npm run build` en `site/` | Cubierta por la implementación |
| Revisión de coherencia | Candidato final | Lista de los lugares revisados y bytes de `global.md`, en el cuerpo del PR. Los lugares: la guía global, las skills y sus referencias, los roles, `site/`, `_support/docs/architecture/` y los detectores de `tests/pilot/` | Cubierta por la implementación |
| Revisión de código | Candidato final, antes de la parada de validación | Claude Code `/code-review` y el commit revisado | La respuesta de revisión (D8 A) |

Validación humana (`interactive`): el usuario lee el diff de estos archivos:
- `content/guidance/global.md`
- `content/skills/{git-workflow,flow-verify,flow-close}/SKILL.md`
- las dos páginas del sitio

Comprueba que una persona que elige `ready` entendería qué hace el agente y dónde se detiene.

## Estado de la revisión y avance

2026-11-02 — Plan escrito y revisado; los hallazgos de `hive-review-plan` se aplicaron, como detalla `proposal.md`.
- T1 hecha · `4aaba3fe` · `global.md` +339 bytes (21.924 → 22.263).
- T2 hecha · `c5a67dd3` · `git-workflow` 9.406 → 11.342 bytes; `gh run watch` usa REST, observado con `GH_DEBUG=api` sobre una corrida pública.
- T3 hecha · `b1cda93f`.
- T4 hecha · `932881b7` · `hive view contract` acepta `in_review`.
- T5 hecha · `e18b9e54` · `npm run build` en `site/` pasa (17 páginas).
- T6 hecha · `fe095359` · todas las rutas de `content/` que nombra el documento existen; los ejemplos de autoría usan `plan-handoff.md`.

Checkpoint: build terminado; `go test ./...` pasa (20 paquetes); sin workflows de CI que reproducir. Siguiente: `flow-verify` (revisión con `/code-review` y validación del usuario antes del push). Hallazgo aparte: `_support/openspec/specs/context-delivery/spec.md:86` todavía exige la skill `hive-core`, disuelta en `5e6e8907`.

2026-11-02 — Verify:
- AC1–AC14 comprobados por observación con `rg` sobre `fe095359`.
- El usuario aprobó el candidato (D12 A).
- `/code-review` (medium, `origin/development...HEAD`) dio 10 candidatos. Se corrigieron 8 en `f6f32661`:
  - CI sobre el nuevo head antes del merge o de la parada;
  - `in_review` en `viewer-publication.md`;
  - un walkthrough legado se lee cerrado en todos los casos;
  - la advertencia de rebase vuelve a la viñeta de base al día;
  - el paso 10 cierra tras la parada de `ready`;
  - `original_commit_id` para los comentarios de revisión;
  - lectura por REST con `user.type`;
  - sin segunda revisión para los hallazgos posteriores al push.
- Descartado: `Delivery: ready` en `tooling` (`doctor_project.go:397` ya lo marca como inesperado).
- A los hallazgos del cierre: la doble ejecución de `flow-close` (ver abajo).
- Segunda revisión, limitada a `f6f32661`, en curso.

Checkpoint: verify paso 3, segunda revisión. Siguiente: aplicar lo bloqueante con `Unreviewed:`, hacer push, abrir el PR a `development` y hacer el merge.

Hallazgos para el cierre:
- `flow-close` corre en la parada de `ready` y otra vez cuando una sesión posterior encuentra el PR mergeado. Ese patrón ya existía con el revisor humano.
- `context-delivery/spec.md:86` todavía exige la skill `hive-core`.
