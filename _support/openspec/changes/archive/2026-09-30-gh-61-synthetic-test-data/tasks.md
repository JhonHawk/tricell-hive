# Tareas

Base del cambio: `6230c7a` (`origin/development`). Worktree: `.claude/worktrees/gh-61b`, rama `feat/gh-61-synthetic-test-data`.

## T1 — `integrations/agents`: golden y perfiles sintéticos, y reglas sobre los roles reales

- [x] Golden sintético, `TestResolve…` con perfiles sintéticos, sin conteos de 20 y con la regla general de esfuerzo. Evidencia, revisada por el orquestador:
  - el golden pasa de 244.206 a 12.393 bytes, con 6 roles sintéticos en `testdata/roles`;
  - `testdata/profiles.json` es sintético y conserva las formas de los perfiles reales;
  - `TestRepositoryRolesRenderOnEveryHostAndCarryEffortOnlyWhereEmitted` reemplaza la tabla de esfuerzo;
  - ya no hay conteos de 20;
  - `TestProfilesReject*` usa `replaceOnce`, que falla si no encuentra el texto;
  - `go test ./integrations/agents` pasa.

  Implementó `hive-build-backend` (Sonnet, configurado).

**Closes:** AC3.

**Depends on:** ninguna.

**Locations:**
- `integrations/agents/render_golden_test.go`;
- `integrations/agents/agents_test.go`:
  - `:15`, `:22`, `:28` y `:251`;
  - las pruebas que leen el rol real `hive-research.md` y los perfiles reales (`:46-50`, `:76-80`, `:134-139` y `:346-366`);
  - `TestProfilesReject*` (`:125-189`), que reemplaza fragmentos del JSON real;
- `integrations/agents/resolve_test.go` (`:21`, `:62` y `:171`);
- `integrations/agents/testdata/`: el golden nuevo, `profiles.json` y los roles sintéticos que haga falta.

**Execution:** delegada a `hive-build-backend`; un paquete, sin cruce con T2 a T4.

**Test approach:** characterization. Se conservan las afirmaciones de lógica con entradas sintéticas, y las nuevas reglas sobre los datos reales se comprueban con mutaciones en T5.

**Verification:**
- `go test -count=1 ./integrations/agents` en verde.
- `wc -c integrations/agents/testdata/render.golden` da 40 KB o menos.
- `rg -n '\b20\b' integrations/agents/*_test.go` no devuelve conteos de roles.
- `rg -n '"\.\.", "\.\."' integrations/agents/*_test.go` solo devuelve las comprobaciones generales sobre los datos reales: todo rol se dibuja, el esfuerzo está presente o ausente según el host, y hay un verificador por host.

## T2 — `tooling/management`: perfiles sintéticos

- [x] `models_test.go`, `catalog_test.go` y `cursor_test.go` leen `testdata/agent-profiles.json`. Evidencia:
  - `sonnet`, `gpt-6` y `haiku` ya no aparecen en las pruebas de `management`;
  - `swapModel` falla si la constante sintética no existe;
  - las únicas lecturas del repositorio real que quedan son `references_test.go:63` y el recorrido de `sync_test.go:97`;
  - `go test ./tooling/management` pasa en 18 s.

  Implementó `hive-build-backend` (Sonnet, configurado).

**Closes:** AC1 (parte de `management`).

**Depends on:** ninguna.

**Locations:** `tooling/management/models_test.go:28,32,150,154,166,179,251`, `catalog_test.go:16,63`, `cursor_test.go:128`, y `tooling/management/testdata/agent-profiles.json` (nuevo).

**Execution:** delegada a `hive-build-backend`.

**Test approach:** characterization. Las afirmaciones de `EffectiveModels` exigen ahora los valores sintéticos. El `strings.ReplaceAll(... "sonnet" ...)`, que no hacía nada en silencio si el perfil cambiaba, pasa a una constante sintética.

**Verification:** `go test -count=1 ./tooling/management` en verde. `rg -n '"\.\.", "\.\."' tooling/management/*_test.go` solo devuelve `references_test.go` (el catálogo real de #71), y cada resultado se revisa a mano. Buscar `ProfilesSource` da falsos positivos: aparece también en las copias a la carpeta temporal.

## T3 — `tooling/cli`: perfiles sintéticos y listado de voces

- [x] Las pruebas de la vista de modelos, `writeUpdateCatalog` y los fixtures de bootstrap usan `tooling/cli/testdata/agent-profiles.json`, y `TestVoiceListShowsRepositoryVoices` exige al menos una voz. Evidencia:
  - ninguna prueba de `cli` lee el `agent-profiles.json` real;
  - solo quedan `voice_test.go:67` y `doctor_project_test.go:380`;
  - `go test ./tooling/cli` pasa en 42 s;
  - `minimal-source` conserva su perfil vacío, porque solo el bootstrap necesita el completo y lo sobrescribe.

  Implementó `hive-build-backend` (Sonnet, configurado).

**Closes:** AC1 (parte de `cli`), AC5 (parte de `cli`).

**Depends on:** ninguna.

**Locations:** `tooling/cli/tui_models_view_test.go:21,46,121-127,157,162,169,177,296`, `update_test.go:70`, `bootstrap_test.go:23,41,338,653`, `voice_test.go:66`, y `tooling/cli/testdata/`.

**Execution:** delegada a `hive-build-backend`.

**Test approach:** characterization.

**Verification:**
- `go test -count=1 ./tooling/cli` en verde.
- `rg -n '"\.\.", "\.\."|Abs\("\.\./\.\."\)' tooling/cli/*_test.go` solo devuelve `doctor_project_test.go:380` (`TestHiveSettingKeys*`) y la lectura de `voice list` en `voice_test.go`, y cada resultado se revisa a mano. `bootstrap_test.go:653` deja de leer el archivo real.

## T4 — `tests/content`: techo de `global.md` y voces por directorio

- [x] (`hive-verify-task`, Sonnet: AC4, AC5 en `tests/content` y AC6 cumplidos. Las mutaciones fallan en la base y pasan con el candidato. Quitar cualquiera de las 10 anclas hace fallar la prueba, y el techo corta en 44.337 bytes.) Techo con margen (D1-A), anclas de AC10 e IDs de voz tomados del directorio. Evidencia de la implementación:
  - `globalGuidanceBudget = 44336` (tamaño actual 43.312 más 1024), sin comprobación de encogimiento; la prueba se llama ahora `TestGlobalGuidanceStaysUnderBudget`;
  - 10 anclas del preámbulo, agrupadas por punto de AC10 (antes eran 16);
  - los IDs de voz se toman del directorio con `filepath.Glob`;
  - `go test ./tests/content/...` pasa.

**Closes:** AC4, AC5 (parte de `content`).

**Depends on:** ninguna.

**Locations:** `tests/content/budget_test.go`, `tests/content/voices_test.go:34,60`.

**Execution:** hilo principal, porque son dos archivos cortos.

**Test approach:** check.

**Verification:** `go test -count=1 ./tests/content/...` en verde. Las mutaciones de AC4 y AC5 se corren en T5.

## T5 — Mutaciones de datos reales

- [x] (`hive-verify-task`, Fable: AC1, AC2, AC4, AC5 y la restricción cumplidos, sobre la versión final tras `/code-review`. Las cuatro mutaciones de datos válidos terminan con código 0, y las cinco de datos inválidos hacen fallar su prueba. El candidato cambió a mitad de la corrida y se repitió toda la serie.) AC1, AC2, AC4 y AC5, y la restricción sobre datos reales inválidos, comprobados con mutaciones temporales en un worktree desechable.

**Closes:** AC1, AC2, AC4, AC5.

**Depends on:** T1 a T4.

**Locations:** un worktree de Git desechable (`git worktree add`) desde el candidato. No sirve una copia sin `.git`, porque `TestAppKeysEscAndBackspaceReturnFromPlaceholder/Project` falla sin mutación (`tui_test.go:457`). No se toca el worktree de trabajo.

**Execution:** la corre el verificador independiente (`hive-verify-task`), que ya reproduce comprobaciones.

**Test approach:** check.

**Verification.** Una corrida de `go test ./...` por mutación, deshaciendo cada una antes de la siguiente:
1. Sustituir cada valor `"model"` no vacío del `agent-profiles.json` real, con `inherit` incluido, por `<valor>-probe`: todo pasa (AC1). En la base fallan `TestRenderGolden`, tres pruebas de `TestResolve…`, cuatro de `models_test.go` y tres de la vista de modelos.
2. Añadir un rol real copiado con otro nombre: `sed 's/hive-review-plan/hive-probe-role/g' content/agents/review/hive-review-plan.md > content/agents/review/hive-probe-role.md`. Todo pasa (AC2).
3. Achicar `global.md` 2 KB: todo pasa (AC4).
4. Borrar `content/voices/mentor.md`: todo pasa (AC5).
5. Restricción de datos reales inválidos. Una mutación de cada tipo, y en cada caso falla la prueba correspondiente:
   - un rol con esfuerzo `turbo`;
   - un perfil `execution` de Claude sin `effort`;
   - un host sin perfil verificador;
   - `global.md` por encima del techo;
   - una ancla del preámbulo eliminada.

## T6 — Retirar la CI

- [x] (`hive-verify-task`, Sonnet: AC6 cumplido.) Se borra `.github/workflows/ci.yml`, y la documentación establece que la verificación es local. Evidencia de la implementación:
  - `git rm` del workflow; `.github/` queda vacío;
  - `deployment-manager.md` dice que no hay CI e incluye los comandos locales, con las pruebas de Python;
  - se quita la nota de `-timeout`, porque `-race` de `tooling/cli` ya tarda 116 s;
  - `installer.md:130` se actualiza.

**Closes:** AC6.

**Depends on:** ninguna.

**Locations:** `.github/workflows/ci.yml`, `_support/docs/architecture/deployment-manager.md` (sección «Verification»), `_support/docs/architecture/installer.md:130`.

**Execution:** hilo principal, porque es un cambio mecánico de configuración y documentación.

**Test approach:** check.

**Verification:**
- `test ! -e .github/workflows/ci.yml`.
- `rg -n 'CI|ci.yml|pull request' _support/docs/architecture/deployment-manager.md _support/docs/architecture/installer.md README.md` no devuelve menciones a la CI de este repositorio.
- `go test -count=1 ./tests/content/...` en verde.

## Verificación y revisión

| Gate | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Pruebas por paquete | T1 a T4, en local | Los comandos de cada tarea | Parte de la implementación |
| Mutaciones | T5, worktree desechable | `hive-verify-task` con otro modelo que el implementador | Parte de `flow-build` |
| Suite local | Una vez sobre el candidato final | `go vet ./...` y `go test ./...` | Parte de la implementación |
| Revisión de código | Antes del push | `/code-review` de Claude Code | Autorizada (decisión de la sesión) |
| Suite con `-race` | No aplica: el cambio no toca concurrencia (D3 de esta sesión: verificación local, `-race` solo para concurrencia) | — | — |

## Estado de la revisión y avance

`/code-review` (10 hallazgos, todos aplicados por `hive-build-backend`):
- se restauraron 3 anclas de AC10;
- `voice list` exige cada voz real;
- la regla de esfuerzo usa la clave exacta de cada host y se fusionó con la prueba de dibujo;
- la prueba de escape de Codex usa `synthetic-unicode`;
- el esfuerzo de `plain-role` es exacto;
- los tres archivos de perfiles sintéticos son idénticos;
- se reutiliza la constante de ruta en `cli`;
- el golden lee las rutas directamente.

Suite local sobre el candidato: `go vet ./...` y `go test ./...` en verde en 46,4 s, con las pruebas de Python OK. `gofmt -l` marca `tests/pilot/regression.go:1115`, que ya venía de la base: `gofmt` reescribiría las comillas invertidas dobles de un comentario como comillas tipográficas. No se toca; se informa como detalle menor.

Cierre (2026-09-30): commits `5a8483d` (pruebas) y `dd07758` (retirada de la CI), integrados con el PR #73 (merge `b8eb49a`). No hubo CI que esperar, y no hizo falta recompilar `hive` ni correr `hive update`, porque solo cambiaron pruebas, documentación y el workflow. #61 quedó cerrado con un comentario.

Plan escrito el 2026-09-30 sobre `6230c7a`. Lo revisó un `hive-review-plan` nativo, de solo lectura, en el dominio de código de pruebas en Go. Reprodujo las mutaciones de AC1 y AC2 sobre la base en una copia desechable.

Aceptados:
- **B1:** la regla de esfuerzo de D2-A pasa a exigir solo presencia o ausencia, porque los valores válidos ya los imponen los validadores; la mutación pasa a ser un perfil de Claude sin esfuerzo.
- **B2:** el antiguo AC6 ya se cumplía en la base y pasa a ser una restricción comprobada en T5; la retirada de la CI queda como AC6.
- **N1:** se añade `resolve_test.go:62`.
- **N2:** las búsquedas de T2 y T3 cambian de patrón, porque `ProfilesSource` daba falsos positivos y negativos.
- **N3:** las pruebas de render que leían `hive-research.md` y los perfiles reales, y `TestProfilesReject*`, pasan a entradas sintéticas.
- **N4:** unos 6 roles sintéticos, con las formas de perfil que el golden real cubría.
- **N5:** unas 9 anclas de voz.
- **N6:** T5 se corre en un worktree de Git.

T6 (retirar la CI) se añadió por decisión del usuario después de la revisión. Es mecánica y el orquestador la reconcilia sin otra ronda. Todas las correcciones aplican la propuesta del revisor, salvo B1, que ajusta D2-A dentro de su intención y se informa al usuario.
