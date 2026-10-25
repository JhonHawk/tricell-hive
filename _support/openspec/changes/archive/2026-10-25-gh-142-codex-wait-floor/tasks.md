# Tareas

Base: `c85b4fce` (2026-10-25). Rama `feat/gh-142-codex-wait-floor`, worktree `.claude/worktrees/gh-142`.

## T1 — Los hijos devuelven solo su resultado final (D19-A)

- [x] `global.md`, los roles y `delegation.md` dicen la regla, y `agent-delivery.md` documenta la variante de Codex.
Verificado por `hive-verify-task` (AC1 cumplido). Antes del commit se corrigió la nota de Codex en `agent-delivery.md`, por observación del verificador: las esperas fueron de 60 000 ms o menos, hubo 5 de 10 000 ms; la sesión del 10-20 corrió con 0.163.0; y se dio la cadencia real de los hijos, unos 375 mensajes por sesión, el primero a los 40–64 s (mediana). Las frases de `global.md` y de los roles expresan la misma regla con redacción distinta, como fija el diseño. Commit `be039f17`, que nombra las sesiones.

**Closes:** AC1.
**Depends on:** ninguna.
**Locations:** `content/guidance/global.md:84`, los 19 roles de `content/agents/**/*.md` que tienen la frase de cierre, `content/skills/flow-research/references/delegation.md:39` y `:65`, `_support/docs/architecture/agent-delivery.md` (incluida la fila de `:150`).
**Execution:** hilo principal. Es la misma frase en varios archivos, más una nota.
**Test approach:** check.
- `rg -F -c "before it, send the parent at most one message, for a blocker that needs the parent's decision" content/agents` da 1 en cada uno de los 19 roles.
- `rg -F "any decision needed from the parent." content/agents` da 0.
- `global.md`: una coincidencia de «they return only their final result, and before it they send the parent at most one message».
- `delegation.md`: una coincidencia de «reports the unavailable source in its final reply» (`:39`).
- `delegation.md`: «not interim status» ausente, y una coincidencia de «as the global delegation rule sets».
- `agent-delivery.md`: una coincidencia de «min_wait_timeout_ms».

**Changes:** redacción de [design.md](design.md#redacción-de-la-guía-t1). `hive-review-plan.md` no tiene la frase de cierre y no cambia. El commit nombra las sesiones `01a15822` (2026-10-20) y `01a1679c` (2026-10-23) de cam en Codex y la falla observada (regla de admisión).
**Verification:** los comandos dan lo esperado, y `go test ./tests/content/... ./integrations/agents/...` pasa en verde (tamaño, enlaces y render de roles; actualizar el golden solo si cambia el render).

## T2 — Lector y editor de `[features.multi_agent_v2]`

- [x] `integrations/codex` clasifica `config.toml` y produce la edición y su inversa, sin dependencias.
Ronda 1, AC2: crear el archivo con 0600 no estaba asignado a ninguna tarea, porque el editor es puro y no escribe. Revisión del plan: esa cláusula pasa a T3. Además, una mutación mostró que ninguna prueba cubre un BOM seguido de la cabecera de la tabla; se corrige en una ronda de T2.
Ronda 2: AC2 cumplido en su alcance. Las dos pruebas nuevas de BOM fallan con la mutación. Commit `636c7bea`.

**Closes:** AC2.
**Depends on:** ninguna.
**Locations:** `integrations/codex/` (archivo nuevo junto a `features.go`, con sus pruebas), `integrations/codex/features_test.go` (caso de tabla hermana).
**Execution:** delegado a `hive-build-backend`, porque es una unidad con interfaz fija y archivos propios.
**Test approach:** tdd. Pruebas en tabla para cada fila de [Casos del editor](design.md#casos-del-editor-t2), más la inversa: quita solo lo registrado y lo que sigue igual, y quita la cabecera o el archivo solo si Hive los creó y quedaron vacíos. Se añade una prueba de que `FeatureEnabled` ignora `[features.multi_agent_v2]`.
**Changes:** [design.md](design.md#enfoque), punto 2.
**Verification:** `go test ./integrations/codex/...` en verde; `go vet ./integrations/...`.

## T3 — Plan, estado y journal (como el paquete de Pi)

- [x] `management` planea, aplica, registra y deshace la escritura sin guardar los bytes del archivo.
Ronda 1, AC3: la relectura y las comprobaciones de huella `Before` y `After` existen, pero ninguna prueba falla si se quitan; en la mutación del verificador todo siguió en verde. Se agregan dos pruebas por failpoint. En la misma ronda se corrigen dos hallazgos locales: una cabecera vacía que quedaba tras un `plan remove`, y un `plan remove` que fallaba entero si cambió `CODEX_HOME`.
Ronda 2: AC3 cumplido; las dos pruebas nuevas fallan con la mutación del verificador. Después, el hilo principal hizo alinear con el plan tres efectos de los arreglos incidentales: avisar las claves ausentes sin valores; borrar el respaldo solo si Hive quitó todas sus líneas en esa ejecución, sin anunciar nunca un respaldo que no existe; y clasificar con el resultado del editor, no con una expresión regular. Hay pruebas nuevas con RED y la suite está en verde, pero no pasó por otra ronda de verificación; lo cubre la revisión de código. El acceso `StoredCodexWaitFloor` es para el doctor de T4. Commit `750f252f`.

**Closes:** AC2 (crear el archivo con 0600 cuando no existe), AC3 (la escritura, la relectura y el registro), AC5.
**Depends on:** T2.
**Locations:** `tooling/management/types.go` (`Options`, `Plan`, `State`), `plan.go`, `apply.go` (`prepareTransaction`, `commitTransaction`, `Recover`), `migration.go` (`PlanUnchanged`), archivo nuevo junto a `pi_package.go`, y sus pruebas.
**Execution:** delegado a `hive-build-backend` cuando T2 esté verificada.
**Test approach:** tdd, imitando `pi_package_apply_test.go` y `pi_package_plan_test.go`:
- `Options.CodexWaitFloor` falso no produce nada, y verdadero lo planea;
- aplicar sin archivo lo crea con modo 0600;
- aplicar crea `config.toml.hive-backup` (0600) antes de escribir y no pisa uno que ya exista;
- aplicar escribe, relee y registra, y el journal no contiene los bytes del archivo;
- `plan remove` borra el respaldo solo tras revertir bien y solo si Hive lo creó;
- deshacer con `remove` quita solo lo registrado;
- un valor cambiado por el usuario se conserva y se informa;
- la huella obsoleta se rechaza;
- un plan guardado con `--out` se aplica;
- con un fallo inyectado entre la escritura y la confirmación, `Recover` quita las líneas aunque Codex haya editado otra parte del archivo y no deja `pending.json`;
- si las líneas registradas ya no están, se informa sin bloquear;
- la entrada del journal y el plan no tienen campo `Data` para `config.toml`, y la verificación usa las huellas `Before` y `After`;
- `prepareTransaction` conserva el registro;
- `PlanUnchanged` lo tiene en cuenta.

**Changes:** [design.md](design.md#enfoque), punto 3.
**Verification:** `go test ./tooling/management/...` en verde; `go test ./tooling/management -run 'Pi|Package|Recover'` sigue en verde.

## T4 — Oferta en instalar y actualizar, versión mínima y fila del doctor

- [x] La CLI ofrece la capacidad, la aplica por el plan y la muestra en el doctor.
Verificado por `hive-verify-task`: AC3 (la oferta), AC4, AC6 y AC7 cumplidos. En vivo y de solo lectura, sobre el home real, `hive doctor` muestra el ajuste aplicado y `update --dry-run` no ofrece nada; la huella de `config.toml` no cambió. Hubo un desvío de Locations, aceptado: `tui_hosts_view.go` (instalación desde la TUI), `agent_browser_test.go` y `agent_browser_tui_test.go`, porque sin ellos la TUI ignoraría la elección. Antes del commit se corrigió el paso siguiente del doctor cuando no hay registro de Hive, con su prueba. Commit `ae217a4b`.

**Closes:** AC3 (la oferta), AC4, AC6, AC7.
**Depends on:** T3.
**Locations:**
- `tooling/providers/types.go`, `catalog.go`, `run.go` y sus pruebas;
- `tooling/cli/provider_adapter.go`, `install.go`, `update.go`, `tui_update_view.go`, `doctor_integrations.go`, `tui_integrations_view.go` y sus pruebas.

**Execution:** delegado a `hive-build-backend` cuando T3 esté verificada.
**Test approach:** tdd, imitando `codex_question_card_test.go`, `codex_question_card_selection_test.go`, `update_test.go`, `doctor_integrations_test.go:747` y `providers_test.go`:
- con el archivo ya aplicado, `Detect` no consulta la versión y no ofrece nada, también con `--dry-run`;
- con el sondeo inyectado, la oferta aparece con 0.164 y no aparece, ni como paso manual, con 0.163 ni sin `codex`;
- no se ofrece en modo no interactivo;
- no se ofrece si ya está aplicada;
- tras elegir, `install` reconstruye el plan con `Options.CodexWaitFloor`;
- `hive update` la ofrece, recalcula `unchanged` y pide confirmación aunque el contenido no cambie;
- la fila del doctor aparece solo con Codex registrado y avisa con la tabla presente y Codex anterior a 0.164;
- el orden del catálogo se mantiene;
- `ValidateStep` acepta la capacidad sin paso planeado.

**Changes:** [design.md](design.md#enfoque), punto 4.
**Verification:** `go test ./tooling/providers/... ./tooling/cli/...` en verde.

## T5 — Documentación y changelog

- [x] La documentación sustituye la regla vieja y describe la capacidad.
Ronda 1, AC8: dos frases no coincidían con el código: `deployment-manager.md:76` («reads no host or provider configuration») e `installer.md:167-168` («a backup it did not create … reported»). Se corrigen en una ronda, junto con los menores que señaló el verificador.
Ronda 2: quedaron dos paréntesis invertidos sobre el dry-run (README y `deployment-manager.md:40`). D22-A: los corrigió el hilo principal y el mismo verificador revisó esas dos líneas: AC8 cumplido. El usuario señaló que esa pregunta sobraba; queda como corrección propuesta a `plan-execution.md`. Commit `3145a4c0`.

**Closes:** AC8.
**Depends on:** T4.
**Locations:** `README.md:150-175`, `llms.txt:44`, `_support/docs/architecture/installer.md:98-150`, `tool-dependencies.md:10-45`, `deployment-manager.md:36-42,76,229`, `tooling/providers/README.md:11-17`, comentarios de `tooling/providers/types.go:5-20`, `catalog.go:3-6`, `run.go:44-47`, `integrations/codex/features.go:14-18`, `tooling/cli/provider_adapter.go:13-17`, `CHANGELOG.md`.
**Execution:** delegado a `hive-write-docs`. Los comentarios de código, que pertenecen a los archivos de T4, los ajusta el implementador de T4.
**Test approach:** check.
- `rg -i -e "edits no host configuration" -e "does not edit Codex" -e "never edits Codex" README.md llms.txt _support/docs tooling` da 0. `CHANGELOG.md` se excluye porque sus entradas viejas son historia.
- `rg -F "lists five to seven integrations" _support/docs` da 0, o el número se actualiza.
- Una coincidencia de «min_wait_timeout_ms» en cada uno de README, llms.txt, `installer.md`, `tool-dependencies.md` y `tooling/providers/README.md`, y de «Codex wait floor» en `CHANGELOG.md`; ninguna está en la base.

**Changes:** [design.md](design.md#enfoque), punto 5. El CHANGELOG va en `[Unreleased]`; por la regla de la entrada de la tarjeta de preguntas, es una capacidad marcada por defecto y sube la versión menor.
**Verification:** los comandos dan lo esperado. Leer el README y `installer.md` confirma que describen la reversión y el respaldo.

## Verificación y revisión

| Gate | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Verificación por tarea | Después de cada tarea, en el worktree | `hive-verify-task`; veredicto por criterio | Parte de la implementación |
| Suite completa | Candidato final, local, sobre la base actual; no hay CI | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` en verde | Parte de la implementación |
| Prueba en vivo del instalador | Candidato final, con el binario construido desde el worktree en una ruta temporal | `hive update --dry-run` sobre el home real, de solo lectura: con la tabla ya aplicada (D17-A), la capacidad no se ofrece y `hive doctor` la muestra aplicada. Escribir, deshacer y recuperar los cubren las pruebas de T3 y T4, porque con `--home` sintético el binario no corre programas y no ofrece la capacidad | Parte de la implementación (local, sin escribir en el `~/.codex` real) |
| Revisión de código | Sobre el PR | `/code-review` de Claude Code (D8-A) | D8-A |

No hay autorizaciones previstas fuera de la entrega:
- El `config.toml` real ya tiene el ajuste (D17-A). Después del merge, `hive update` lo debe reconocer como aplicado sin escribir nada.
- Recompilar el binario tras el merge es parte de la actualización local, por la marca de mantenedor (AGENTS.md).

## Entrega

`automatic` (D7-A, vigente en esta sesión para este repositorio):
1. Rama de trabajo en un worktree.
2. Commits por tarea, push y PR a `development`.
3. `/code-review` y merge con la suite en verde.
4. #142 se cierra al integrarse.
5. Recompilar `hive` desde el checkout, porque el gestor cambia, y luego `hive update`.

El registro se versiona al cierre con un PR pequeño sin revisión dedicada, que hago yo.

## Estado de la revisión y progreso

**Revisión del plan, ronda 1 (2026-10-25).** Hubo dos revisores `hive-review-plan`: gestor (unos 133 k tokens, 176 s) y guía y documentación (unos 105 k tokens, 85 s). Todos los hallazgos se aceptaron y aplicaron:
- La elección entra por `Options.CodexWaitFloor` y el plan se reconstruye.
- `update` detecta, ofrece y confirma siempre.
- Journal por líneas, sin bytes de `config.toml`, con inversa también en `Recover`, sin bloquear.
- Gate en vivo de solo lectura más sondeo inyectado.
- `migration.go`, `prepareTransaction`, esquema 6 con `omitempty` y su limitación.
- AC5 pasa a `plan remove`; la fila del doctor aparece solo con Codex registrado; `continue` para Codex anterior a 0.164.
- Frase de bloqueo idéntica y precisa; `delegation.md:39`; justificación de admisión.
- Comprobaciones de T5 falsables y ubicaciones que faltaban.

**Re-revisión del gestor (única ronda; unos 143 k tokens, 50 s).** Las siete correcciones quedaron bien y dejaron tres hallazgos:
- Gate sintético inalcanzable: se quitó. Además, `Detect` comprueba «ya aplicado» antes de la versión. Aplicado tal como lo propuso el revisor.
- Confirmación de una entrada por líneas sin bytes: huellas `Before` y `After`, edición recalculada e inversa sobre el archivo actual. Aplicado tal como lo propuso el revisor.
- Sin respaldo, D16 queda incumplido: lo resolvió el usuario con D20-A, una copia junto al archivo, y se añadió a diseño, AC3 y pruebas de T3.

Cobertura: gestor (T2–T4 y gates) y guía y documentación (T1, T5). Ninguna corrección queda sin revisar; las de la re-revisión se concilian como propuestas del revisor o como decisión del usuario.

**Revisión de código y entrega (2026-10-25).**
- `/code-review` (medio) sobre `3145a4c0` dio 10 hallazgos. `hive-refute-claim` confirmó F1–F4 y descartó F5 por inalcanzable. Se corrigieron en `f6bb72c2`, `4f0105ec` y `0f5154b0`: F1–F4 y F6–F9, más F5 como refuerzo. F10, la duplicación, quedó como simplificación reportada.
- La re-revisión única dio 10 hallazgos. Ocho locales se corrigieron sin otra ronda, con pruebas vistas fallar antes, en `53d67a1c`, `93aa47f2` y el commit `(T4)` siguiente. El hallazgo 9 ya existía antes de este cambio, para el paquete de Pi, y se propone como ticket aparte.
- Suite final en verde: `go vet`, `go test` y `unittest` (64 pruebas). La prueba en vivo de solo lectura, con el binario final sobre el home real, mostró el ajuste aplicado y «no escrito por Hive»; `update --dry-run` no ofreció nada y la huella de `config.toml` no cambió.
- [PR #154](https://github.com/JhonHawk/tricell-hive/pull/154): merge `0bd7fc54`; #142 se cerró al integrarse. Se recompiló `~/.local/bin/hive` y `hive update` aplicó `0bd7fc54`.
- Modelos: hilo principal Opus 5.5 (observado); `hive-build-backend`, `hive-write-docs`, `hive-verify-task`, `hive-review-plan` y `hive-refute-claim` según su perfil configurado, no observado.
