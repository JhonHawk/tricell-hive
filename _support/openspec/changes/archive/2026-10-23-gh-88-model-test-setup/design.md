# Diseño y evidencia

## Base inspeccionada

Fecha: 2026-10-23. Revisión: `7b0775c803a05f1ac97732bec13b246a13e37222`. Go 1.27.0, macOS arm64. La única entrada previa de `git status` era `_support/sessions/` sin seguimiento, que se preserva.

Una corrida `go test -race -json -count=1 -timeout=12m ./tooling/cli` pasó: paquete 316,680 s, tiempo total 319,812 s, 685 pruebas principales. Los 128 nombres principales que empiezan por `TestModels` sumaron 190,61 s; los otros 557, 124,52 s. No sumar subtests y padres: sus tiempos se solapan. La prueba del flag de caché `TestTestLogFlagStillExists` pasó.

Los cuatro casos principales más lentos fueron el catálogo grande en seis hosts (31,63 s), equivalencia interfaz/comando (20,53 s), cambios de grupo frente a cambios individuales (12,06 s) y tamaños de pantalla (10,31 s). Los tres últimos conservan su escenario actual.

Una segunda corrida aisló el catálogo grande con `-race -count=1 -run '^TestModelsViewFitsAndScrollsWithManyRolesOnSixCLIs$' -cpuprofile <temporary-profile> -o <temporary-test-binary>`: padre 32,21 s; subtests 80×24 y 120×40, 0,24 y 0,19 s. La diferencia aproximada de 31,78 s corresponde a preparación fuera de ambos subtests, no a cada acción visual.

El perfil CPU registró 30,81 s de muestras: `installViaText` acumuló 22,86 s, `validatePlan` 22,21 s y `target.Safe` 18,30 s. Son tiempos inclusivos de CPU, no porcentajes independientes del tiempo de pared. Apoya coste de validación/archivos en la instalación de esta fixture; no prueba una causa para todo el paquete ni una regresión respecto del 19 de octubre.

Evidencia local de esta sesión, no publicar archivos crudos:

- `/var/folders/dg/0d963h2158s10jggj2_k2w_00000gn/T/hive-issue88-bvn5cov9.jsonl`: corrida completa.
- `/var/folders/dg/0d963h2158s10jggj2_k2w_00000gn/T/hive-issue88-profile-0ee8bbu6/`: JSON, perfil CPU y binario exacto del caso aislado.
- `/var/folders/dg/0d963h2158s10jggj2_k2w_00000gn/T/hive-issue88-sample-cf01jak0.txt`: muestra previa de dos segundos; sus conteos inclusivos no son proporciones del tiempo total.

La evidencia temporal podría desaparecer; las cifras anteriores son el resumen reproducible. Para una comparación usar el comando, la revisión y las condiciones indicadas, sin competir con otras suites. Documentación consultada mediante Context7: Go `go1.27.1`, salida JSON de pruebas; la instalación local es 1.27.0.

## Decisiones técnicas

**D-A — Dividir el plan sólo en la preparación del catálogo grande.** En [la prueba de catálogo](../../../../../tooling/cli/tui_models_view_test.go), reemplazar su instalación combinada por una llamada a `installDirect` por cada elemento de `installerHosts`. El helper existente de [pruebas de actualización](../../../../../tooling/cli/update_test.go) construye un plan real y llama `Engine.Apply`; no simula estado. Se conserva el mismo home exclusivo de la prueba, fuente de 22 roles, configuración y recorrido visual.

El camino actual usa adaptadores core sin pasos externos; `Engine.Onboard` delega directamente en `Apply` en esa condición. `desiredResources` resuelve los hosts del plan, y `commitTransaction` vuelve a validar el plan antes de cada escritura. Planes menores deberían evitar parte del trabajo multiplicado; los seis commits y el estado creciente tienen coste propio, por lo que el ahorro requiere medición.

Cada instalación conserva los registros anteriores. Habrá seis transacciones en vez de una; este test comprueba la vista instalada, no la historia transaccional. Las pruebas que sí comprueban instalación combinada/historia permanecen iguales. Comprobar explícitamente `selectedModelsHost == installerHosts[i]` y que cada host contiene los 22 nombres, además de preservar las assertions existentes; así la preparación no puede omitir silenciosamente hosts o roles.

**D-B — Hosts mínimos sólo en escenarios de un host.** Cambiar argumentos locales, sin alterar helpers ni expectativas. `toHost` puede seguir verificando el host seleccionado; no sustituirlo por un índice que presuponga Claude primero.

| Prueba | Archivo en `tooling/cli/` | Preparación actual → propuesta |
| --- | --- | --- |
| `TestModelsViewMarksAnOverriddenRoleAndKeepsTheMarkNextToTheCursor` | `tui_models_edit_test.go` | cinco hosts → `claude` |
| `TestModelsViewFlatListsHaveNoSectionHeaders` | `tui_models_picker_test.go` | cuatro por subtest → `host` del subtest |
| `TestModelsViewProviderHeadersAreReadableWithAFallback` | `tui_models_picker_test.go` | `claude,opencode` → `opencode` |
| `TestModelsViewTwoHundredNineModelsReachTheLastOneWithPageDown` | `tui_models_picker_test.go` | `claude,opencode` → `opencode` |
| `TestModelsViewListDoesNotRepeatAnEffectiveModelTheCLIListsAndKeepsCatalogOrder` | `tui_models_picker_test.go` | `claude,codex` → `codex` |
| `TestModelsConfirmationFitsEightyColumnsWithLongOpenCodeIDs` | `tui_models_review_fixes_test.go` | `claude,opencode` → `opencode` |
| `TestModelsConfirmationOfASevenRoleGroupKeepsOneRowPerRole` | `tui_models_review_fixes_test.go` | `claude,opencode` → `opencode` |
| `TestModelsViewCodexEffortResetSaysSoInThePanel` | `tui_models_followups_test.go` | `claude,codex` → `codex` |
| `TestModelsViewPickerShowsThePositionAmongTheModels` | `tui_models_followups_test.go` | dos por tamaño → `codex` en ambos tamaños |
| `TestModelsViewXOnAGroupWithoutOverrideSaysThereIsNothingToReset` | `tui_models_followups_test.go` | cinco hosts → `claude` |

Estos diez padres sumaron 9,84 s en la corrida completa. Aun eliminar todo ese tiempo no alcanzaría el objetivo de cinco minutos; la mejora principal debe venir de D-A (planes menores en el catálogo grande).

**D-C — Preservar el escenario multihost.** No alterar `TestModelsViewApplyEqualsTheMatchingCommand`, su `twin` ni `assertTwin`; `TestModelsCommandSetGroupEqualsTheRolesOneByOne`; `TestModelsViewBrowsingNeverQueriesAList`; `TestModelsViewPickerFitsWith209ModelsAndLongNames`; ni la navegación a OpenCode de `TestModelsViewEnterOpensThePanelWithTheEffectiveModel`.

Se descartan clones con rebasing: las rutas viven en claves de registros, targets, consumidores y configuración; el hook Grok incrusta rutas dentro de payloads con hash; los journals guardan planes/estados e integridad. `copyTree` es para fuentes y no preserva modos/symlinks. También se descartan cachés globales y eliminación de validaciones productivas: no son necesarias para probar esta alternativa.
