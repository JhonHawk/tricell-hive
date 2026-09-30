# CI en pull requests a `development` y pruebas locales sin fsync ni `-race` por defecto

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `development` el 2026-09-29 con el [PR #69](https://github.com/JhonHawk/tricell-hive/pull/69) (merge `31cdd43`), con la CI en verde; #62 y #64 cerrados |
| Tracker · GitHub Issues | • [#64 — chore: run -race only when it adds value, add CI, and scope tests/pilot](https://github.com/JhonHawk/tricell-hive/issues/64)<br>• [#62 — test(management): skip fsync in tests to cut package time](https://github.com/JhonHawk/tricell-hive/issues/62) |
| Git | `automatic` · publicar `development` desde la punta de `rebuild/harness-engineering`, rama `feat/gh-64-ci-and-fast-local-tests`, PR a `development`, `/code-review` antes del push, CI en verde, merge · sin parada humana |
| Verificación | `go vet ./...` · `go test ./...` · un `go test -race` final en la CI del PR · medición de tiempos contra la base · pruebas de Python |
| Siguiente paso | Ninguno en este cambio. La primera CI tardó 27 min, así que #60 y #61 son el siguiente recorte de tiempo |

## Objetivo

Las pruebas tardan más que los cambios que verifican. Hay tres causas. `README.md:38` prescribe `go test -race ./...`, que en macOS lleva `tooling/cli` a 17–19 min. El repositorio no tiene CI, así que `flow-build` exige correr la suite completa en local antes de cada entrega. Y cada escritura del gestor llama a fsync, que en macOS es un vaciado completo a disco (`F_FULLFSYNC`): cuesta 42 s en `tooling/management` y 63 s en `tooling/cli` por corrida.

Este cambio desactiva fsync solo dentro de los binarios de prueba y deja la verificación local por defecto en `go vet ./...` más `go test ./...`. Añade una CI de GitHub que corre `go vet`, `go test -race` y las pruebas de Python en cada pull request hacia `development`. `development` pasa a ser la rama base del repositorio y reemplaza a `rebuild/harness-engineering`. El resultado esperado: en local la suite completa baja de unos 4,5 min a unos 2,5 min sin `-race`, y la corrida con `-race` pasa a hacerse una vez por pull request en la CI, no en cada ronda del agente.

## Alcance y aceptación

Incluido:

- Opción de prueba que desactiva fsync, activada desde los `TestMain` de `tooling/management` y `tooling/cli` (D1-A).
- Que `TestNewFlowCasesHaveDistinctFixturesAndContracts` se salte cuando falta `node`, como ya hace su prueba hermana. Es un hallazgo incidental de la corrida en Linux que la CI necesita resuelto.
- El workflow de GitHub Actions que corre en pull requests hacia `development` (D2).
- Crear `development` y declararla como base en `AGENTS.md`, incluida la excepción temporal para refrescar la instalación local (D4-A).
- Actualizar la verificación documentada en `README.md`, `deployment-manager.md` e `installer.md`.

Excluido:

- `tests/pilot` sigue en la corrida normal (D3-A): tarda 6 s y apartarlo no ahorra nada. Ese punto de #64 se cierra con la medición.
- #59, #60, #61 y #63: quedan para después. Siguen siendo útiles para `tooling/cli`.
- Protección de ramas, cambiar la rama por defecto de GitHub (hoy `master`), borrar `rebuild/harness-engineering` o integrar en `master`.
- El contenido distribuido (`content/`): `flow-build` ya dice cómo tratar un proyecto con CI y sin ella.

Restricciones:

- El binario `hive` que corre fuera de las pruebas sigue sincronizando cada escritura y cada cambio de enlace.
- `go vet ./...` y `go test ./...` siguen pasando en macOS, y en Linux dentro de la CI.
- No se modifica ningún caso de `tests/fixtures/regression/`.

Criterios:

- AC1. En macOS, `go test -count=1 ./tooling/management` tarda 25 s o menos y `go test -count=1 ./tooling/cli` 170 s o menos, en esta máquina. *Falso en la base cuando* se miden 53 s y 204 s, respectivamente (medidos el 2026-09-29 en `a5f3d1f`).
- AC2. Una prueba de `tooling/management` comprueba que la sincronización está activa por defecto y ejecuta, con ella activa, una escritura de archivo y un reemplazo de enlace. Fuera de los archivos `_test.go`, nada llama a la función que la desactiva. *Falso en la base cuando* no existe ninguna opción ni ninguna prueba así.
- AC3. Sin `node` en el `PATH`, `go test -count=1 ./tests/pilot/` pasa y la prueba de casos de flujo aparece como saltada. *Falso en la base cuando* `TestNewFlowCasesHaveDistinctFixturesAndContracts` falla con `exec: "node": executable file not found`.
- AC4. Un pull request hacia `development` dispara el workflow `CI`, que corre `go vet ./...`, `go test -race ./...` y las pruebas de Python, y termina en verde. Un push directo a cualquier rama no lo dispara. *Falso en la base cuando* el repositorio no tiene `.github/workflows/`.
- AC5. `README.md`, `_support/docs/architecture/deployment-manager.md` e `_support/docs/architecture/installer.md` establecen como verificación local `go vet ./...` más `go test ./...`. Reservan `-race` para cambios que tocan concurrencia y para la CI. *Falso en la base cuando* `README.md:38` prescribe `go test -race ./...`.
- AC6. `origin/development` existe y contiene la punta de `origin/rebuild/harness-engineering` (`git merge-base --is-ancestor` sale con 0). `AGENTS.md` declara `Base branch: development`, y su excepción de refresco nombra `development`. *Falso en la base cuando* la rama no existe y `AGENTS.md:8` declara `rebuild/harness-engineering`.

## Entrega

Decisiones del usuario del 2026-09-29, en esta sesión:

- **D1-A:** la opción que desactiva fsync cubre `tooling/management` y `tooling/cli`.
- **D2:** CI en las máquinas de GitHub, solo en pull requests hacia `development`, siguiendo la convención de los otros proyectos.
- **D3-A:** `tests/pilot` sigue en la corrida normal.
- **D4-A:** `development` reemplaza a `rebuild/harness-engineering` como rama base; `rebuild/harness-engineering` queda congelada como historia.
- **D5-A:** entrega automática.
- **D6-A:** `/code-review` de Claude Code antes del push.

Secuencia:

1. Crear `development` desde `origin/rebuild/harness-engineering` y publicarla. Es la primera escritura remota y la autorizan D4-A y D5-A.
2. Crear la rama `feat/gh-64-ci-and-fast-local-tests` desde `development`, en un worktree, e implementar T1–T4.
3. Pasar `/code-review` sobre el diff y resolver o refutar sus hallazgos. Push y `gh pr create --base development`; la rama por defecto es `master` y sin `--base` el PR iría ahí. El PR ya corre el workflow nuevo, porque en `pull_request` GitHub usa el workflow del commit de merge.
4. Comprobar que `rebuild/harness-engineering` no recibió commits nuevos (T0), esperar a que la CI termine en verde y hacer merge.
5. Cambiar el checkout principal a `development` y actualizarlo. Recompilar `hive`, porque cambió `tooling/`. No hace falta `hive update`, porque `content/` no cambia. Actualizar las memorias locales que nombran `rebuild/harness-engineering` como base, sobre todo `feedback_agent_owns_base_sync_redeploy.md`, para que las sesiones futuras actualicen `development`.
6. Cerrar #62 y #64 a mano, con un comentario que enlace el PR y registre D3-A. Las palabras clave de cierre de un PR solo actúan sobre la rama por defecto (`master`), así que aquí no sirven.

El registro del cambio se versiona una vez, al cerrar, con `flow-close`: un commit en `development` mediante un PR pequeño hacia `development`, sin revisión dedicada, al que yo hago merge. Ese PR solo toca `_support/openspec/**`, así que la CI se salta a propósito por `paths-ignore`.
