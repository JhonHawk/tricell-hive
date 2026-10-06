# Pruebas sin datos fijados del repositorio (resto de #61)

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `development` el 2026-09-30 con el [PR #73](https://github.com/JhonHawk/tricell-hive-private/pull/73) (merge `b8eb49a`); #61 cerrado |
| Tracker · GitHub Issues | • [#61 — test: use fixtures instead of the real repository tree and data](https://github.com/JhonHawk/tricell-hive-private/issues/61) (parte de datos fijados; la de velocidad se entregó en #71) |
| Git | `automatic` · rama `feat/gh-61-synthetic-test-data` desde `development`, PR a `development`, `/code-review` antes del push, CI en verde, merge · sin parada humana |
| Verificación | `go vet ./...` · `go test ./...` en local · pruebas de mutación de datos reales · sin CI (T6) |
| Siguiente paso | Ninguno en este cambio |

## Objetivo

Varias pruebas leen el `agent-profiles.json` y los roles reales y exigen sus valores exactos. Por eso, un cambio deliberado de contenido rompe pruebas que no protegen nada de ese cambio. Por ejemplo, el cambio de modelo 9aed6e5 (GPT-6.1 Sol) obligó a editar cinco archivos de prueba, entre ellos un golden de 244 KB. Otras cuentan 20 roles o 3 voces con números fijos. Y el límite de `global.md` falla también cuando el archivo se achica: se editó en 58 de 102 commits.

Este cambio pasa esas pruebas a datos sintéticos. Sobre los datos reales deja solo comprobaciones generales que valen para cualquier contenido válido. Resultado esperado: cambiar un modelo, añadir o quitar un rol o una voz, o achicar `global.md` ya no exige tocar pruebas, y la lógica sigue cubierta con entradas sintéticas.

## Alcance y aceptación

Incluido:

- `integrations/agents`:
  - un golden de render con unos 6 roles sintéticos y perfiles sintéticos;
  - fuera los conteos fijos de 20 roles;
  - la tabla de esfuerzo pasa a una regla general (D2-A);
  - `TestResolve…` con perfiles sintéticos.
- `tooling/management`: `models_test.go`, `catalog_test.go` y `cursor_test.go` con perfiles sintéticos.
- `tooling/cli`:
  - las pruebas de la vista de modelos, `writeUpdateCatalog` y los fixtures de bootstrap con perfiles sintéticos;
  - `TestVoiceListShowsRepositoryVoices` exige al menos una voz y ninguna `preamble`.
- `tests/content`:
  - el límite de `global.md` pasa a ser solo un techo, con unos 1024 bytes de margen (D1-A);
  - las frases fijas de las voces se reducen a una ancla por cada punto de AC10 de `gh-46-voice-layer` (unas 9), y los IDs de voz se toman del directorio.

Se conservan como comprobaciones sobre los datos reales:
- todo rol real se dibuja en los 6 hosts, con esfuerzo válido en Claude, Codex y Pi, y sin esfuerzo en Grok, OpenCode y Cursor;
- los perfiles reales pasan `ReadProfiles`, cubren los 6 hosts y tienen un perfil verificador en cada uno;
- el techo de `global.md`, las anclas del preámbulo y el tamaño de cada voz;
- `TestHiveSettingKeys*` y `TestRepositoryCatalogueInstructionReferences`.

Excluido:
- #60 y #63;
- cualquier cambio en `content/` o en el código de producción.

Restricciones:

- La lógica que cubría cada prueba sigue cubierta con entradas sintéticas. No se borra ninguna afirmación de comportamiento sin que otra la sustituya.
- `go vet ./...` y `go test ./...` siguen en verde.
- Los datos reales inválidos se siguen detectando. T5 lo comprueba con una mutación de cada tipo, y todas deben fallar:
  - un rol que no se dibuja en algún host (esfuerzo `turbo` en un rol);
  - un perfil de Claude sin `effort` (regla general de D2-A);
  - un host sin perfil verificador;
  - `global.md` por encima del techo;
  - una ancla del preámbulo eliminada.

  Era AC6 en la primera versión del plan. Pasó a restricción porque ya se cumple en la base.

Criterios. AC1, AC2 y AC5 se comprueban con mutaciones temporales de los datos reales en un worktree desechable, que se deshacen después:

- AC1. Si se sustituyen los modelos en `integrations/agent-profiles.json` (por ejemplo `sonnet` por `sonnet-next` y `gpt-6.1-sol` por `gpt-7`), `go test ./...` pasa. *Falso en la base cuando* fallan `TestResolveReturnsProfileModelAndProfileEffort`, `TestRenderGolden`, `models_test.go` y las pruebas de la vista de modelos.
- AC2. Si se añade un rol real válido (`content/agents/<grupo>/hive-probe-role.md`, copia de un rol existente con otro nombre), `go test ./...` pasa. *Falso en la base cuando* fallan los conteos de 20 roles, la tabla de esfuerzo y el golden.
- AC3. El golden de render se genera solo con roles y perfiles sintéticos y ocupa como mucho 40 KB. Sigue fijando byte a byte la salida de los 6 hosts. *Falso en la base cuando* `render.golden` ocupa 244 KB y se genera con los 20 roles reales.
- AC4. Achicar `global.md` en más de 1024 bytes no hace fallar ninguna prueba, y pasar del techo sí. El techo queda en el tamaño actual más unos 1024 bytes. *Falso en la base cuando* achicar más de 1024 bytes hace fallar `TestGlobalGuidanceStaysWithinRatchetedBudget`.
- AC5. Si se borra una de las voces reales (`content/voices/mentor.md`), `go test ./...` pasa. *Falso en la base cuando* fallan `tests/content/voices_test.go:60` y `TestVoiceListShowsRepositoryVoices`.
- AC6. El repositorio no tiene CI, y la documentación de verificación dice que las pruebas corren en local: `go vet ./...` y `go test ./...`, con `-race` solo para cambios de concurrencia. *Falso en la base cuando* existe `.github/workflows/ci.yml` y `deployment-manager.md` describe la CI en cada PR.

## Entrega

Siguen en vigor las decisiones de esta sesión: entrega automática y `/code-review` de Claude Code antes del push. Decisiones del usuario del 2026-09-30:

- **D1-A:** techo de `global.md` con unos 1024 bytes de margen. Consecuencia aceptada: un crecimiento pequeño deja de exigir tocar la prueba y la señal para #47 se debilita.
- **D2-A:** regla general de esfuerzo en lugar de la tabla. Ajuste tras la revisión del plan: la regla exige que haya esfuerzo en Claude, Codex y Pi y que no lo haya en Grok, OpenCode y Cursor. El conjunto de valores válidos lo imponen `Parse` y `ReadProfiles`: con una lista propia, la regla habría rechazado contenido válido como `ultra`.
- **D3:** se retira la CI de GitHub (añadida en #69). La verificación es local, y bb1 queda como opción para pruebas concretas, no como regla. Motivo del usuario: el 90 % de los cambios serán de contenido. Esto se añade como T6 y AC6.

Secuencia:

1. Worktree con `feat/gh-61-synthetic-test-data` desde `origin/development`.
2. T1 a T4 en paralelo.
3. T5, las mutaciones.
4. `/code-review`.
5. `gh pr create --base development`.
6. Merge. Con T6 ya no hay CI que esperar.
7. Recompilar `hive` si cambió `tooling/`; solo cambian pruebas, así que no hace falta `hive update`.
8. Cerrar #61.

El registro se cierra con un PR pequeño hacia `development`, sin revisión dedicada, al que yo hago merge.
