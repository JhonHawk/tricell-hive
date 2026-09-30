# Tareas

Base del cambio: `8acdb6e` (`origin/development`). Worktree: `.claude/worktrees/gh-65`, rama `feat/gh-65-66-plain-file-change-messages`.

## T1 — Error tipado en el gestor

- [x] (`hive-verify-task`, Opus: cumplido; lo confirman seis mutaciones que hacen fallar sus pruebas. Implementada por `hive-build-backend`, Sonnet configurado. Añade `ManagedFileChangedError` con los tipos cambiado, bloque y ausente; el texto de «cambiado» sale de `legacy.ModifiedFileError`; `withPath` evita la ruta doble en `plan.go` y `voice.go`. Antes del cambio fallaban `TestOwnedReturnsTypedErrorWithPlainWords`, `TestVoiceConflictDoesNotRepeatThePath` y las dos pruebas de `cli`; después pasan. El orquestador corrigió el texto del bloque a «differs from what Hive wrote», porque el de voz decía «what voice wrote».) `owned()` devuelve el error tipado con sus tres redacciones, y ni `transformResource` ni el conflicto de voz repiten la ruta.

**Closes:** AC3.

**Depends on:** ninguna.

**Locations:**
- `tooling/management/files.go:257-283` (`owned`);
- `tooling/management/plan.go:425`;
- `tooling/management/voice.go:321`;
- `tooling/management/voice_review_test.go:209-224`: la aserción `HasSuffix(err.Error(), path)` pasa a exigir que la ruta aparezca una sola vez;
- pruebas nuevas en `tooling/management`;
- una prueba en `tooling/cli` para los comandos de texto.

**Execution:** delegada a `hive-build-backend`. Toca `tooling/management` y una sola prueba nueva en `tooling/cli`, que no se cruza con los archivos de T2.

**Test approach:** tdd. Primero las pruebas, que deben fallar en la base:
- una por cada redacción: archivo cambiado (skill, agente, enlace o tipo), bloque gestionado y archivo ausente;
- cada una exige el tipo con `errors.As`, la ruta una sola vez y su texto.

Después, el cambio.

**Verification:**
- `go test -count=1 ./tooling/management` en verde.
- En `tooling/cli`, una prueba con una skill compartida editada:
  - `plan remove` falla con el texto nuevo y la ruta una sola vez;
  - `hive install` agregando `cursor`, que no pasa por la búsqueda de instalaciones antiguas, falla con el mismo texto y `errors.As` lo reconoce como el tipo nuevo.

## T2 — El aviso neutro y su documentación

- [x] (`hive-verify-task`, Opus: cumplido; lo confirman seis mutaciones que hacen fallar sus pruebas. Implementada por `hive-build-backend`, Sonnet configurado. Antes del cambio fallaban las cuatro pruebas de aviso y la nueva `TestHostsViewUninstallAllWorksWithAUserFileAtALegacyPath`; después pasan. Se probó en una copia sin la prueba de T1, que todavía no compila. La verificación independiente espera a T1.) `legacyScanNote` usa la redacción de D3-A y el aviso genérico nuevo; `deployment-manager.md` distingue los dos casos.

**Closes:** AC1, AC2.

**Depends on:** ninguna.

**Locations:**
- `tooling/cli/tui_hosts_view.go:66-75`;
- `tooling/cli/tui_error_states_test.go`: `:63` (caso b), `:139` (aviso genérico), `:226` y `:358` (caso a), más una prueba nueva de Uninstall all en el caso (a);
- `_support/docs/architecture/deployment-manager.md`, párrafo de CLIs: las dos citas del aviso y la frase sobre rechazar quitar o instalar.

**Execution:** delegada a `hive-build-backend`, en paralelo con T1. Solo toca `tooling/cli` (sin la prueba de T1) y la documentación.

**Test approach:** tdd. Las pruebas del aviso exigen la redacción neutra y fallan en la base. La prueba nueva de Uninstall all usa hosts registrados y un archivo del usuario en una ruta antigua que el catálogo actual no gestiona, porque `legacyPathEnv` no tiene filas. Exige que el aviso se vea y que Uninstall all termine con «Hive removed».

**Verification:**
- `go test -count=1 ./tooling/cli` en verde.
- `rg -n 'installing or removing|removing or installing' tooling/cli _support/docs/architecture/deployment-manager.md` no devuelve nada.

## T3 — Agregar Cursor con una skill editada, en la vista

- [x] (`hive-verify-task`, Opus: cumplido; lo confirman seis mutaciones que hacen fallar sus pruebas. Implementada por el hijo de T2, reanudado. `TestHostsViewAddingCursorWithAChangedSkillRefusesOnce` falla con el texto antiguo y pasa con el nuevo; la prueba de `:90` usa el texto nuevo; `go test ./tooling/cli` pasa en 42,5 s en su sitio.) La vista muestra el rechazo de #66 una vez, sin cortar y sin el aviso duplicado, a 80×24 y a 120×40.

**Closes:** AC4.

**Depends on:** T1, que da el mensaje, y T2, que toca el mismo archivo de pruebas.

**Locations:** `tooling/cli/tui_error_states_test.go`: `:90` y una prueba nueva que usa `driftedEnv` para todos los hosts menos Cursor, `openDriftedCLIs`, `toggle("cursor")` y `a`. No hace falta un `lookPath` falso, porque `hostsTestDeps` ya da todos los hosts por detectados.

**Execution:** el hijo de T2 reanudado cuando termine T1, porque toca el mismo archivo.

**Test approach:** tdd. La prueba nueva falla en la base.

**Verification:** a los dos tamaños, la prueba nueva comprueba lo que exige AC4:
- «differs from what Hive expects there» aparece una sola vez en la pantalla aplanada;
- la ruta aparece una sola vez, contada con `squash`;
- `mustNotShow("press m to read all")` y `mustNotShow("modified managed skill")`;
- `assertFits` pasa.

Después, `go test -count=1 ./tooling/cli` en verde.

## T4 — Revisión de experiencia de uso y verificación de punta a punta

- [x] (Veredictos:
  - `hive-verify-change`: AC1, AC3 y AC4 cumplidos en los binarios reales, con capturas `verify-*`.
  - `hive-review-ux`: aceptable y sin regresiones, con dos observaciones medias en proceso de arreglo: M1, el rechazo no decía que no se hizo nada, así que se le antepone «Nothing was changed: »; M2, la frase del aviso pasa a «Removing may also be refused if Hive installed that file.», que refina D3-A. Capturas `a-*` y `b-*`. Las dos quedaron aplicadas: `refusedText` en `onRemovePlanned` y `failInstall`, y la frase en `legacyScanNote` y la documentación. Las pruebas exigen el prefijo; `go test ./...` y las pruebas de Python pasan.)

  Hay veredicto de `hive-review-ux` y de `hive-verify-change` sobre los estados (a) y (b) en la vista de CLIs.

**Closes:** AC1, AC4 (confirmación en la vista real).

**Depends on:** T1 a T3.

**Locations:** el binario del candidato y el de la base, compilado desde `git archive 8acdb6e` en un directorio temporal. Las capturas van a `_support/workspace/2026-09-30-gh-65-66-plain-file-change-messages/images/`.

**Execution:** dos hijos independientes en paralelo, `hive-review-ux` y `hive-verify-change`, cada uno con su propio servidor `tmux -L <nombre>` para no tocar el del usuario.

**Test approach:** check.

**Verification.** Entorno aislado, sin `--home`: con `--home` la vista ignora el `PATH` y no detecta un `cursor-agent` falso (`install.go:603-605`). Se usa:
- `HOME=<tmp>` y `--state-dir` explícito;
- `PATH=<bin-falso>:/usr/bin:/bin`, con `cursor-agent` falso en `<bin-falso>`;
- `CODEX_HOME`, `CLAUDE_CONFIG_DIR` y `XDG_STATE_HOME` sin definir, para no tocar la instalación real;
- el estado preparado con `hive install --hosts …` en ese mismo entorno.

Estados que se comprueban en 80×24 y 120×40, frente a la base:
- (a) aviso neutro y Uninstall all termina;
- (b) aviso neutro, quitar se rechaza con el mensaje nuevo y agregar Cursor muestra el rechazo una vez y completo.

La interfaz tiene un solo tema, así que no hay tema oscuro que revisar. Cada hijo devuelve capturas de texto y su veredicto.

## Verificación y revisión

| Gate | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Verificación por tarea | T1 a T3, en local | `hive-verify-task` con otro modelo que el implementador | Parte de `flow-build` |
| Vista real | T4, `tmux` privado | `hive-review-ux` y `hive-verify-change` | Parte de la implementación, porque hay efecto visible en la interfaz |
| Suite local | Una vez sobre el candidato final | `go vet ./...`, `go test ./...` y las pruebas de Python | Parte de la implementación |
| Revisión de código | Antes del push | `/code-review` de Claude Code | Autorizada (decisión de la sesión) |

## Estado de la revisión y avance

Cierre (2026-09-30):
- commits `57e6186` (gestor) y `2df1a68` (vista y documentación), integrados con el PR #75 (merge `3f24787`);
- suite local en verde; el repositorio no tiene CI;
- `hive` recompilado desde `3f24787`, y `hive doctor` muestra los seis CLIs verificados;
- #65 y #66 cerrados con un comentario.

Plan escrito el 2026-09-30 sobre `8acdb6e`. Lo revisaron dos `hive-review-plan` nativos, de solo lectura y en paralelo: uno del gestor (T1, AC3) y otro de la interfaz (T2 a T4, AC1, AC2 y AC4).

Aceptados, del gestor:
- **H1:** redacción propia para los bloques gestionados.
- **H2:** `voice.go:321` y `voice_review_test.go`.
- **H3:** la instalación se prueba con Cursor.
- **H4:** texto de archivo ausente fijado, igual que `doctor.go`.
- **H5:** se quita la promesa sobre `update` y Releases, y los errores de marcadores quedan excluidos.

Aceptados, de la interfaz:
- **F1:** el aviso engañaba en el caso (b). El usuario eligió D3-A, redacción neutra.
- **F2:** el entorno de T4 va sin `--home`.
- **F3:** ayudantes de T3 y aserciones exactas de AC4.
- **F4:** `:139` y la frase de la documentación.
- **F5:** entorno propio para Uninstall all.
- **F6:** coincide con H1 y H2.

Todas las correcciones aplican la propuesta de los revisores o una decisión del usuario, así que no hubo segunda ronda.

Después de la revisión de experiencia de uso, `/code-review` devolvió diez hallazgos, todos aplicados:
- **F1:** el aviso se oculta solo si el rechazo nombra la misma ruta (`scanPath` frente a `messagePath`).
- **F2 y F3:** textos propios para un bloque ausente y para el archivo de un bloque ausente.
- **F4:** la frase del aviso pasa a ser una oración aparte.
- **F5:** el registro ya cita el texto final.
- **F6:** texto propio cuando solo cambiaron los permisos.
- **F7:** la prueba de Cursor pasa por el comando real.
- **F8:** el prefijo pasa a «Nothing was changed. ».
- **F9:** `owned()` pone la ruta en todos sus errores y se elimina `withPath`.
- **F10:** un solo ayudante `squash`, y el texto esperado sale de `legacy.ModifiedFileError`.

`hive-verify-task` (Opus) volvió a verificar la versión final: AC1, AC3 y AC4 cumplidos, confirmados con 6 mutaciones y con la vista real en `tmux` a 80×24 (`reverify-b-add-cursor-80x24.txt`). También detectó que las correcciones habían devuelto el bloque de voz a «what voice wrote». El orquestador lo corrigió a «what Hive wrote» y lo fijó en `TestVoiceConflictDoesNotRepeatThePath`, que falla con el texto anterior. `management` y `cli` pasan.

Hallazgo aparte, sin ticket todavía: si falta un archivo gestionado y no hay copia de seguridad, no hay salida, porque instalar y quitar se rechazan.
