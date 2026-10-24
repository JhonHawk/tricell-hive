# Tareas — gh-134-criteria-approval-pr-evidence

Commit base: `3292d518` (`development`, 2026-10-24). Rama `feat/gh-134-criteria-approval-pr-evidence` en el worktree `.claude/worktrees/gh-134`, que también aloja esta carpeta del cambio (sin commit) mientras dura el trabajo.

## Diseño

- **Aprobación en la pregunta de ruta (D3-A).** El resumen del plan listo agrega el texto de cada `AC<n>` y el nombre de cada requisito que cambia, porque son la spec que el usuario aprueba. Las opciones de la pregunta de ruta no cambian, pero la de iniciar `flow-build` dice que aprueba los criterios mostrados por sus IDs (por ejemplo, «aprueba AC1–AC4»), porque la tarjeta nativa puede mostrarse sola (`questions.md:38`), sin pegar su texto. Una corrección en texto libre no responde la pregunta de ruta: se revisa el plan con [plan revision](../../../../content/skills/flow-plan/references/plan-revision.md), se re-revisa cuando lo exige `plan-review.md:39`, y luego se muestran los criterios revisados y se vuelve a hacer la pregunta de ruta. Con la ejecución ya autorizada, los criterios se muestran y el trabajo sigue sin preguntar.
- **Evidencia en el PR.** El PR que entrega el cambio agrega, por cada `AC<n>` que fijaron la conversación o el registro del cambio, la evidencia registrada bajo `verification.md:96` (criterio, comando o recorrido, resultado y limitación). «Lo no verificado» incluye las líneas `accepted unverified by the user` (`plan-format.md:37`) y «los supuestos» son los del plan (`plan-format.md:13`). Sin criterios, declara los checks que corrieron, lo no verificado y los supuestos. El PR que solo publica el registro del cambio (`flow-close/SKILL.md:75`) queda fuera. La plantilla del repositorio sigue mandando.
- Cada regla conserva un solo hogar: `flow-plan/SKILL.md` para el resumen y `git-workflow/SKILL.md` para el PR. Las páginas del sitio describen el comportamiento y enlazan, sin copiar la regla.

## Tareas

### T1 — El resumen del plan listo muestra los criterios y la ruta los aprueba

- [x] El párrafo del resumen del plan listo exige el texto de cada `AC<n>` y los nombres de los requisitos que cambian. Dice que iniciar `flow-build` aprueba esos criterios, que una corrección en texto libre revisa el plan antes de construir, y que con la ejecución ya autorizada se muestran y se continúa.

**Closes:** AC1, AC2.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-plan/SKILL.md:70`.

**Execution:** hilo principal; el cambio es una frase y su brief sería más largo que el cambio.

**Test approach:** check — `rg` sobre las frases nuevas y la prueba de tamaño de skills.

**Changes:** agregar los criterios y los requisitos a la lista del resumen; agregar a la pregunta de ruta que la opción de iniciar `flow-build` aprueba los criterios mostrados por sus IDs, y que una corrección en texto libre no responde la ruta sino que revisa el plan y vuelve a preguntar con los criterios revisados; extender la última frase para mostrar los criterios cuando la ejecución ya está autorizada.

**Verification:** `rg -n "AC<n>" content/skills/flow-plan/SKILL.md` muestra la línea del resumen con el texto de los criterios, su aprobación por IDs en la opción de ruta, la corrección que vuelve a preguntar y el caso autorizado; `wc -c content/skills/flow-plan/SKILL.md` da 16 000 o menos; `go test ./tests/content/ -run 'Size'` pasa.

### T2 — El cuerpo del PR declara la evidencia de cada criterio

- [x] El párrafo del cuerpo del PR exige, para el PR que entrega el cambio y con criterios `AC<n>` fijados en la conversación o el registro, cada criterio con su evidencia registrada, lo no verificado (incluido lo aceptado sin verificar) y los supuestos aceptados; sin criterios, los checks que corrieron, lo no verificado y los supuestos. Excluye el PR que solo publica el registro del cambio.

**Closes:** AC3.

**Depends on:** ninguna.

**Locations:** `content/skills/git-workflow/SKILL.md:32`.

**Execution:** hilo principal; mismo motivo que T1.

**Test approach:** check — `rg` sobre la frase nueva y la prueba de tamaño de skills.

**Changes:** ampliar el párrafo que empieza con «When a pull request is authorized» sin cambiar lo que ya pide (problema, comportamiento, checks, revert, plantilla).

**Verification:** `rg -n "pull request is authorized" content/skills/git-workflow/SKILL.md` muestra la evidencia por criterio con su fuente registrada, lo no verificado, los supuestos, el caso sin criterios y la exclusión del PR de cierre del registro; `go test ./tests/content/ -run 'Size'` pasa.

### T3 — El sitio y el changelog describen los dos cambios

- [x] `flows/plan.md` dice que el resumen muestra los criterios y que elegir construir los aprueba; `flows/build.md` dice qué declara el PR por criterio; `CHANGELOG.md` registra el cambio en `[Unreleased]`, con referencia a #134.

**Closes:** AC4.

**Depends on:** T1, T2 (el sitio describe el texto final).

**Locations:** `site/src/content/docs/flows/plan.md` (pregunta frecuente «Does a ready plan mean the agent may start building?»), `site/src/content/docs/flows/build.md` (lista de lo que hace el flujo), `CHANGELOG.md` (`[Unreleased]` → `Changed`).

**Execution:** hilo principal; depende del texto exacto de T1 y T2.

**Test approach:** check — `rg` y el build del sitio.

**Changes:** una o dos frases por página, describiendo el comportamiento y enlazando al skill sin copiar la regla; una entrada de changelog.

**Verification:** `rg -n -i "approve" site/src/content/docs/flows/plan.md` muestra la pregunta frecuente actualizada (en la base no hay coincidencia en esa pregunta); `rg -n -i "per criterion|each criterion" site/src/content/docs/flows/build.md` muestra la frase nueva (en la base no hay coincidencia); `npm run build` en `site/` termina sin errores.

## Verificación y revisión humana

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Verificación por tarea | Tras cada tarea, en la rama de trabajo | `hive-verify-task` con contexto limpio: `AC<n>: met` por criterio | Parte de la implementación |
| Gates compartidos | Candidato final, local | `go vet ./...` sin salida; `go test ./...` con todos los paquetes `ok`; `npm run build` en `site/` completo | Parte de la implementación |
| Validación humana | Antes del push (`interactive`) | El usuario lee el diff y, si lo pide, prueba la guía desplegada desde la rama en una sesión real | D4-A |

Sin revisión de código dedicada (D5-A), sin revisión de seguridad (no cruza ningún límite) y sin recorrido de UI (no hay efecto renderizado en la aplicación).

## Estado de la revisión y progreso

- Plan escrito el 2026-10-24. Revisión: un `hive-review-plan` (dominio: guía distribuida y sitio, único afectado) sobre `proposal.md` 175b63af0c7c y `tasks.md` 1157e24e8193; 80 s y unos 106 mil tokens. Sin bloqueantes; criterios falsos en la base y cubiertos.
- Seis sugerencias aceptadas tal como las propuso el revisor, sin re-revisión (`plan-review.md:39`): (1) excluir el PR de cierre del registro; (2) tomar la evidencia de lo registrado bajo `verification.md:96`, lo aceptado sin verificar y los supuestos del plan; (3) una corrección en texto libre no responde la ruta y se vuelve a preguntar; (4) la opción de ruta nombra los IDs que aprueba; (5) verificaciones de T3 que ya no pasan en la base; (6) criterios fijados en la conversación o el registro.
- Límites: el revisor no ejecutó pruebas ni leyó cada detector de `tests/pilot` (solo búsqueda de texto).
- Ruta D6-A (2026-10-24): `flow-build` iniciado; AC1–AC4 aprobados.
- T1 `[x]` (`440d6934`), T2 `[x]` (`9729bf13`), T3 `[x]` (`72af5f6a`), cada una con veredicto `met` de un `hive-verify-task` con contexto limpio. Modelo del implementador: Opus 5.5 (observado en el entorno de la sesión); modelo del verificador: configurado, no observado. Tras el veredicto de T1 se separó en su propia oración la descripción de la opción `flow-build`, sin cambiar el significado (nota del verificador).
- Gates compartidos sobre `72af5f6a`: `go vet ./...` sin salida; `go test ./...` con 20 paquetes `ok`; `pnpm run build` en `site/` (script `build`, 10 páginas) completo. Sin revisión de código dedicada (D5-A), sin revisión de seguridad (no cruza límites), sin recorrido de UI.
- Notas del verificador de T3, no aplicadas por ser matices que AC4 no pide: el sitio no menciona el caso con ejecución ya autorizada ni la exclusión del PR de cierre del registro.
- Tracker: GitHub Issues no tiene estado «en progreso» y el repositorio no usa etiqueta ni asignación para ello; #134 sigue abierto.
- Validación del usuario D7-A (2026-10-24). Push, [PR #135](https://github.com/JhonHawk/tricell-hive/pull/135) con la evidencia por criterio, merge `0d12f7c0` en `development` (base al día comprobada con `merge-base --is-ancestor`; sin CI registrada). #134 cerrado por el PR.
- Registro archivado y publicado por un PR pequeño de cierre.
