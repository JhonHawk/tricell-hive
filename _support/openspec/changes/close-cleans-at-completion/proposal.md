# close-cleans-at-completion — el cierre limpia al llegar al merge y deja una sola respuesta pendiente

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `2eef1969` · rama propuesta `feat/close-cleans-at-completion` |
| Playbook | investigate → cambio de reglas (ninguno) |

## Decisiones

2026-11-21
- **Reply A (investigación)** — ¿Qué hacer con los cierres que dejan ramas y worktrees y piden un segundo cierre? Se adopta el híbrido: limpieza segura al llegar al merge, una sola respuesta pendiente, el fin de sesión corre lo pendiente y el helper reporta clones no registrados (`engram:tricell-hive/decision/close-safe-cleanup-at-completion`). Es un cambio de diseño que decide el usuario; además, la falla del fin de sesión se observó en sesiones independientes del piso (Grok `01a1fb69`, Grok `01a1f7e8`, Codex `01a1f86e-f093` y `01a1f515`).
- **Precisión del usuario (con la Reply A)** — Las evidencias (imágenes y videos) siguen aparte y sin versionar. El ticket recibe evidencia y comentario en el paso 1, que se reporta ahí mismo, antes de declarar el cierre; desde ese momento la evidencia ya cumplió su propósito y es segura de borrar.
- **D1 (respuesta libre)** — ¿Se versiona el contenido de sesión sin preguntar? No: en general no se versiona, porque envejece mal y las evidencias inflan el repositorio; deciden los criterios de retención que ya existen (`project-layout/references/layout.md` «Retain evidence»). Lo automático es OpenSpec: la carpeta de cambio se commitea con el código.
- **D2 (respuesta libre)** — ¿Qué entrega el lote? Solo archiva carpetas de cambio; no commitea material de soporte. Lo que se versiona va en la rama o el commit de la implementación.
- **Supuesto** — Como el paso 2 ya no pregunta, la limpieza completa corre al llegar al merge. Solo queda pendiente, con la respuesta «Close» del pie, lo que necesita al usuario: lo que el helper conserva por una razón que el usuario resuelve, un volcado o export, y los recursos remotos (efecto crítico). Alternativa descartada: diferir aún el paso 5 entero. Para revertirlo, se restaura la fila 1 de `close-modes.md`.
- **Supuesto** — La evidencia que el ticket o el cuerpo del PR ya tienen se borra al llegar al merge, con un inciso en `global.md` «Preservation» que lo autoriza; el resto del material que se conserva se mueve al checkout principal antes de quitar el worktree, como hoy. Alternativa descartada: esperar a la respuesta «Close». Para revertirlo, se quita el inciso y la acción vuelve a preguntarse.
- **Supuesto** — El inventario del paso 2 deja de proponer «Commit» y «Archive»: el contenido de sesión no se versiona por defecto (D1) y el lote no commitea (D2). Una promoción que «Retain evidence» propone, o una carpeta `sessions/` cerrada, se nombra en «Next» del reporte con el comando de `project-layout`, sin contar como pendiente. Alternativa descartada: commitearlas en el lote de archivo. Para revertirlo, se restauran las acciones y el paso 6 de `batch-archive.md`.
- **Supuesto** — Fin de sesión: `close-modes` enuncia el criterio (un mensaje del usuario que termina la sesión sin pedir más trabajo); la auditoría detecta un subconjunto de frases al inicio de un mensaje corto sin «?», documentado en `rule-enforcers.md`. Alternativa descartada: una lista cerrada en la guía. Para revertirlo, la guía nombra la lista.
- **Supuesto** — El reporte nombra cada rama conservada como «kept `<rama>`: <razón>», forma que la auditoría lee para su excepción. Alternativa descartada: buscar palabras libres. Para revertirlo, se quita la forma de `branch-cleanup.md` «Report».
- **Supuesto** — El fin de sesión corre los pasos pendientes y solo cuenta los archivos en espera, sin entregar el lote (la entrega del lote sigue siendo una respuesta o un pedido de higiene). Alternativa descartada: abrir el PR de archivo al salir. Para revertirlo, el fin de sesión añade el lote.
- **Supuesto** — Los clones no registrados se reportan y nunca se borran, porque su dueño es incierto. Alternativa descartada: borrarlos si están limpios y su rama está fusionada.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Pedido

> «A, pero recuerda que ahora el contenido de la sesión tambien se commitea, lo que recuerda que manteniamos aparte era el tema de evidencias, imagenes/videos , eso puede seguir asi, y recordando que los tickets se actualizan con evidencia y comentarios antes de declarar el cierre (y se reporte en ese mismo paso), o sea, si el agente reporta como siguiente paso cerrar, es porque la evidencia ya cumplió su proposito y es seguro eliminarla (aunque aun no lo haga)»

Origen: investigación de cierres posteriores al breakpoint `2bb96478` en fleetsystem, cam, carvento y kin, más una línea base anterior. Sin ticket.

## Objetivo

Hoy, al llegar al merge, el cierre deja pendientes el material de soporte (paso 2) y la limpieza de ramas y worktrees (paso 5) hasta una respuesta «Close»; el archivo va en una respuesta «Hygiene» aparte, y pedir higiene repite el cierre entero. En sesiones reales la respuesta no llega o el agente no corre lo pendiente al terminar la sesión, así que quedan ramas fusionadas, worktrees y clones. Con el cambio, el cierre limpia antes de su reporte, el pie ofrece a lo sumo una respuesta de higiene para los archivos de specs, y la auditoría marca como fallo lo que hoy marca como revisión.

## Alcance y aceptación

Dentro: `close-modes`, `flow-close`, `close-hygiene`, `support-material`, `batch-archive`, `report-endings`, `close-report`, `trackers`, `viewer-close`, `pr-captures`, `layout.md` «Retain evidence», `archive.md`, `inventory.md`, `branch-cleanup.md`, `unattended-delegation`, `global.md` «Preservation», `cleanup_branches.py` y su prueba, dos criterios de `--audit-session` con sus fixtures, la rúbrica, `rule-enforcers.md`, `workspace-and-artifacts.md`, la página del sitio de cierre, README, `llms.txt`, `CHANGELOG.md` y el delta de `workflow-lifecycle`. Fuera: el movimiento de tickets (sin regresión observada), las carpetas temporales de sesiones anteriores en `/var/folders`, y los cierres de pausa y `hold`.

- AC1 — Un cierre al llegar al merge borra la rama fusionada, su worktree limpio y las ramas fusionadas antiguas, y corta procesos y temporales, antes de escribir el reporte final y sin esperar respuesta · false today because `close-modes.md` fila 1 dice «Waits: Steps 2 and 5, recorded as pending» · Observed in: `~/.agents/skills/hive-core/references/close-modes.md` y `flow-close/SKILL.md` instalados tras `hive update`
- AC2 — El paso 2 no pregunta: borra la evidencia de `workspace/` cuyo adjunto muestra la lista de adjuntos del ticket leída en el paso 1, o cuyo URL está en el cuerpo del PR fusionado, y conserva sin versionar lo demás; una promoción propuesta se nombra en «Next»; `global.md` «Preservation» autoriza ese borrado · false today because `support-material.md` pide una decisión para toda acción distinta de keep, solo acepta el cuerpo del PR, y `global.md:32` no autoriza borrar material de verificación · Observed in: `support-material.md`, `layout.md` y el bloque global instalados
- AC3 — El registro `archive-pending/<id>` lleva solo la carpeta, el PR y el merge, y el lote no aplica acciones de soporte · false today because `batch-archive.md` paso 6 aplica commit, archivo y promoción del material · Observed in: `batch-archive.md` y `close-hygiene.md` instalados
- AC4 — Cuando nada necesita al usuario, el pie ofrece a lo sumo una respuesta de limpieza, «Hygiene» con el número de archivos pendientes; un pedido de higiene tras el cierre de una obra corre solo lo pendiente y el lote, y reporta solo eso, sin un segundo reporte final · false today because el pie ofrece «Close» mientras los pasos 2 y 5 esperan, y `close-modes.md:9` y `close-report.md:24` hacen correr «Every step» y escribir el reporte a un pedido de higiene · Observed in: `report-endings.md`, `close-modes.md` y `close-report.md` instalados
- AC5 — `--audit-session` falla `merged_branch_deleted` cuando el usuario termina la sesión («es todo», «terminamos», «por hoy», «that's all») después de un merge cuya rama no se borra luego · false today because la sesión Codex `01a1f86e-f093` (FSR-900) da `review` · Observed in: `go run ./tests/pilot --audit-session` sobre esa sesión real
- AC6 — `--audit-session` falla `close_before_branch_cleanup` cuando un reporte de cierre precede al borrado de una rama fusionada antes que él, y da review solo si ese reporte nombra la rama como conservada · false today because `closeBeforeCleanupStatus` es siempre `review` · Observed in: `--audit-session` sobre la sesión real Grok `01a1fb69`
- AC7 — `cleanup_branches.py` lista en `stray_clones` cada repositorio con su propio `.git` bajo las carpetas de worktrees cuyo `origin` coincide con el del repositorio, con su rama y si tiene cambios, y nunca lo borra · false today because su salida no tiene esa clave y `.worktrees/fsr-905-906-frontend` no aparece · Observed in: el helper sin `--apply` sobre `fleetsystem-frontend-monorepo` (solo lectura)

## Prerequisites

- [x] Archivos locales fuera del repositorio que observan AC5 a AC7: el rollout Codex `01a1f86e-f093`, la sesión Grok `01a1fb69` y el clon `fleetsystem/.worktrees/fsr-905-906-frontend` · owner: usuario (no borrarlos antes de T2 y T3) · needed by: T2, T3. Si el clon ya no existe, T2 observa AC7 sobre la fixture de su prueba.

## Contexto verificado

Leído el 2026-11-21 en `development` `2eef1969`: `close-modes.md`, `close-hygiene.md`, `flow-close/SKILL.md`, `support-material.md`, `batch-archive.md`, `branch-cleanup.md`, `report-endings.md` «The close footer», `layout.md` «Retain evidence», `session_audit.go` (`closeBeforeBranchCleanup`, `closeBeforeCleanupStatus`), `cleanup_branches.py` (`plan_cleanup`), los requisitos de cierre de `workflow-lifecycle/spec.md`. Ninguna carpeta de cambio pendiente bajo `_support/openspec/changes/`. Sesiones reales y su evidencia: investigación de esta sesión (cuatro hijos `hive-research`).

- `python3 -m unittest tests.skills.cleanup_branches_test -q` · OK, 12 s · 2026-11-21
- `go test ./tests/pilot/ -run 'TestAudit' -count=1` · ok, 0,4 s · 2026-11-21
- `go test ./tests/content/ -count=1` · ok, 2,3 s · 2026-11-21
- `pnpm --dir site build` · exit 0, 20 páginas, 3,4 s · 2026-11-21
- `go test ./...` · exit 0, 241 s · 2026-11-21
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` · OK, 91 s · 2026-11-21
- `go test ./tests/content/ -run 'TestStage|TestReadingFootprint' -count=1` · cubierto por `go test ./tests/content/` (ok); margen del paquete de cierre: miembros 22.099 B de 22.500 B; `global.md` 29.433 B de 30.000 B · 2026-11-21
- `rg -n 'Steps 2 and 5|\*\*Support\.\*\*|step 2.s question' content/` · 5 resultados (`flow-verify/SKILL.md:27`, que queda, y `close-modes.md:7`, `:9`, `:17`, `batch-archive.md:12`, que T1 quita) · 2026-11-21
- `rg -n 'support decision|support-material decision|only a capture whose URL|never a disposable input|replaced only when|counts as replaced only|only when its upload' content/ _support/docs/architecture/` · 9 resultados, que T1 y T5 quitan · 2026-11-21
- `python3 /Users/jmartinez/.agents/skills/flow-plan/scripts/check_change.py _support/openspec/changes/close-cleans-at-completion` · exit 0 · 2026-11-21
- `go run ./tests/pilot --audit-session <rollout Codex 01a1f86e-f093> --host codex` · `merged_branch_deleted` = review · 2026-11-21
- `go run ./tests/pilot --audit-session <Grok 01a1fb69 updates.jsonl> --host grok` · `close_before_branch_cleanup` = review · 2026-11-21
- `python3 -I content/skills/flow-close/scripts/cleanup_branches.py --repo /Users/jmartinez/Development/projects/comsys/fleetsystem/fleetsystem-frontend-monorepo --base development` · not measured: `stray_clones` lo añade T2

## Effects

None: the tasks take no effect on an external system.

## Entrega

`automatic` (respuesta de la sesión, 2026-11-21): revisión, push, PR y merge sin parar; después, el refresco de la instalación local que `AGENTS.md` autoriza en el checkout del mantenedor. Revisión de código: Native: Claude Code `/code-review`. No autoriza release, ni cambios en otros proyectos (los restos de fleetsystem y cam se limpian desde sus propias sesiones).

## Estado de la revisión y progreso

2026-11-21 — Plan escrito.
2026-11-21 — Revisión del plan con dos revisores (`hive-review-plan` servidor: 0 P0/P1, 4 P2, 10 P3, 224 s, 166 k tokens; guía: 2 P1, 7 P2, 3 P3, 225 s, 168 k tokens). Revisión 1 aplicada: `global.md` autoriza el borrado de evidencia (P1); «Close» queda solo para lo que necesita al usuario y commit/promoción van a «Next» (P1); T1 se divide en T1 y T5 con las hojas que dependían del criterio viejo; README, `llms.txt` y `CHANGELOG.md` entran a T4; la auditoría reconoce el helper por variable, la promoción antes del fin de sesión y la forma «kept `<rama>`»; `stray_clones` añade `unpushed` y compara `origin` normalizado; AC2 y AC4 reescritos.
2026-11-21 — Revisión 2, tras la relectura (guía: 2 P2, 2 P3; servidor: 1 P1, 3 P2, 1 P3): la excepción de rama conservada exige ``kept `<rama>` `` de una rama borrada; la promoción cuenta solo cuando la sesión muestra la cabeza; el fin de sesión repite la limpieza y lista lo pendiente con su pregunta; el delta modifica el requisito de capturas del PR y alinea el del helper; «Report» pasa a T5 y T2 y T3 dependen de T5. Segunda relectura: N1–N5 y N1–N4 corregidos; aplicados 4 P3 (rama del merge visible para la excepción de fin de sesión, campos de `stray_clones` contra «Report», forma «kept» en el último mensaje, fila sin espera solo para el fin de sesión sin cierre previo). Sin P0, P1 ni P2 abiertos.
