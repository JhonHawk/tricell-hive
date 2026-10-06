# Pruebas de `tooling/cli` sin el árbol real y sin la espera de 60 s

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `development` el 2026-09-30 con el [PR #71](https://github.com/JhonHawk/tricell-hive-private/pull/71) (merge `c08388e`), con la CI en verde en 8 min 20 s; #59 cerrado y #61 abierto con la parte de datos fijados |
| Tracker · GitHub Issues | • [#61 — test: use fixtures instead of the real repository tree and data](https://github.com/JhonHawk/tricell-hive-private/issues/61) (solo la parte de velocidad, D1-A)<br>• [#59 — test(cli): no_CLI_hosts update subtest waits about 60 s](https://github.com/JhonHawk/tricell-hive-private/issues/59) |
| Git | `automatic` · rama `feat/gh-61-fast-cli-test-sources` desde `development`, PR a `development`, `/code-review` antes del push, CI en verde, merge · sin parada humana |
| Verificación | `go vet ./...` · `go test ./...` · tiempos con y sin `-race` · CI del PR |
| Siguiente paso | Ninguno en este cambio. Lo que queda de #61 son las pruebas atadas a datos del repositorio |

## Objetivo

La CI tarda 27 min, y 24 de ellos son `tooling/cli` con `-race`. Hay dos causas. Unas 19 pruebas de esa suite instalan y aplican el árbol real del repositorio, y con `-race` cada una tarda unos 30 s. Además, una subprueba espera 60 s por un defecto de la propia prueba. Ninguna de esas pruebas afirma nada sobre el contenido real: prueban el asistente de instalación, los planes que caducan, la recuperación, el bootstrap y los textos. El contenido real queda cubierto por el validador y por una prueba dedicada.

Este cambio mueve esas pruebas a una fuente mínima fija en `tooling/cli/testdata/`, como es idiomático en Go, y arregla la espera. Deja una sola prueba que instala y aplica el contenido real. Además, protege el gancho de fsync de #69 con `testing.Testing()`. Resultado esperado: `tooling/cli` baja de ~140 s a menos de 45 s sin `-race`, y la CI de ~27 min a menos de 12.

## Alcance y aceptación

Incluido:

- El ayudante `openMenuEntry` dibuja la vista antes de seguir (#59).
- Una fuente mínima fija en `tooling/cli/testdata/` que reemplaza a `minimalCatalogSource`, generada hoy con `sync.OnceValues` y borrada en `TestMain` (D2-A).
- Las pruebas de `tooling/cli` que instalan o leen el árbol real pasan a la fuente mínima o a roles sintéticos: `installArgs`, `interfaceTestOptions`, los fixtures de bootstrap, `provider_adapter_test.go`, `main_test.go`, `tui_test.go` y el catálogo de la vista de modelos.
- Una sola prueba, en `tooling/management`, que instala y aplica el catálogo real.
- `DisableDiskSyncForTests` entra en pánico fuera de un binario de prueba (D3-A).

Excluido. Queda en #61, que sigue abierto:

- el golden de agentes, la tabla de esfuerzos y las pruebas de modelos atadas a perfiles reales;
- `TestVoiceListShowsRepositoryVoices` (`voice_test.go:66`) y las frases fijas de las voces;
- el límite de tamaño de `global.md` en las dos direcciones.

También quedan fuera:

- `TestHiveSettingKeys*`, que compara el código con `global.md` a propósito;
- #60 (paralelismo) y #63 (juntar pruebas duplicadas).

Restricciones:

- Cada prueba movida conserva sus afirmaciones de comportamiento. No se borra ninguna prueba ni se debilita ninguna afirmación.
- El binario `hive` no cambia de comportamiento, salvo el pánico de D3-A, que solo ocurre si código de producción llama al gancho.
- `go vet ./...` y `go test ./...` siguen en verde.

Criterios:

- AC1. Sin `-race`, `TestUpdateViewErrorsShowInsideTheView/no_CLI_hosts` tarda 2 s o menos. *Falso en la base cuando* tarda 61 s (medido el 2026-09-30 en `3530d13`).
- AC2. Ninguna prueba de `tooling/cli` instala ni planifica el árbol real, ni lee los agentes reales. Se comprueba con dos búsquedas:
  - `rg -n 'Abs\("\.\./\.\."\)' tooling/cli/*_test.go` solo devuelve `voice_test.go`;
  - `rg -n '"\.\.", "\.\.", "content"' tooling/cli/*_test.go` solo devuelve `doctor_project_test.go:380`, la lectura de `global.md`.

  Siguen permitidas esa lectura y las de `agent-profiles.json` en `update_test.go:70` y `tui_models_view_test.go:46`: son datos fijados que D1-A deja en #61. *Falso en la base cuando* la primera búsqueda devuelve 20 usos y la segunda encuentra también `tui_models_view_test.go:389`.
- AC3. La fuente mínima es un árbol fijo en `tooling/cli/testdata/`. Ya no existen `minimalCatalogSource` ni su borrado en `TestMain`. *Falso en la base cuando* `tui_hosts_test.go` define `minimalCatalogSource` con `sync.OnceValues` y la borra en `TestMain`.
- AC4. Una prueba de `tooling/management` instala y aplica el catálogo real del repositorio en un home temporal. Comprueba que el estado resultante no tiene diferencias y que existen archivos gestionados de cada tipo: la guía global, una skill y un agente. *Falso en la base cuando* `TestRepositoryCatalogueInstructionReferences` solo planifica.
- AC5. Llamar a `DisableDiskSyncForTests` fuera de un binario de prueba provoca un pánico con un mensaje claro. Una prueba lo comprueba sustituyendo el predicado de binario de prueba. *Falso en la base cuando* la función solo guarda el valor.
- AC6. En esta máquina y corriendo solo, `go test -count=1 ./tooling/cli` tarda 60 s o menos, y con `-race` 240 s o menos. *Falso en la base cuando* se miden 140 s y 839 s. La prueba de concepto del revisor dio 48,1 s sin `-race`. Las pruebas que el plan no toca ya suman unos 48 s y corren en serie, porque #60 queda fuera.
- AC7. En la CI del PR, el paso `go test -race` tarda 12 min o menos. *Falso en la base cuando* tardó 26 min 14 s en la corrida 36673081482.

## Entrega

Las decisiones de esta sesión siguen en vigor para este repositorio y esta rama destino: entrega automática (D5-A) y `/code-review` de Claude Code antes del push (D6-A). Decisiones del usuario del 2026-09-30:

- **D1-A:** solo la parte de velocidad de #61.
- **D2-A:** fuente mínima en `testdata/`.
- **D3-A:** chequeo con `testing.Testing()`.

Secuencia:

1. Worktree con `feat/gh-61-fast-cli-test-sources` desde `origin/development`.
2. T1 a T3.
3. `/code-review`.
4. `gh pr create --base development`.
5. CI en verde, T4 y merge.
6. Recompilar `hive`, porque cambia `tooling/management`.
7. Cerrar #59. Comentar en #61 lo hecho y lo que queda.

El registro se cierra con un PR pequeño hacia `development`, sin revisión dedicada, al que yo hago merge. La CI se lo salta por `paths-ignore`.
