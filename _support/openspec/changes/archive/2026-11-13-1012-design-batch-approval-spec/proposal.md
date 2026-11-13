# design-batch-approval-spec — El registro recoge el delta de la aprobación por conjunto

| Campo | Valor |
| --- | --- |
| Estado | Completado |
| Tracker | ninguno |
| Git | JhonHawk/tricell-hive · base `development` en `17eaad3c` · rama `chore/design-batch-approval-spec` |

## Solicitud

Del usuario, 2026-11-13: resolver el hallazgo H10 de la validación de `business-rules-record`, la suite falla en `development` en `TestSpecDeltasApplied`.

## Objetivo

El PR #466 (`7006c9c4`) cambió en la spec `workflow-lifecycle` los requisitos «Screen proposals in Figma before the plan» y «Business rules that appear while drawing», para que una sola decisión apruebe el conjunto de pantallas terminado, y lo hizo sin delta. El último delta archivado de ambos, el de `2026-11-13-1008-business-rules-record`, dejó de coincidir con la spec y `go test ./tests/content/` falla. Este cambio registra el delta `MODIFIED` de esos dos requisitos tal como la spec los tiene hoy; no cambia la spec ni el contenido distribuido.

## Criterios

- **AC1** — El último delta archivado de «Screen proposals in Figma before the plan» y de «Business rules that appear while drawing» es igual al requisito de la spec viva, y `TestSpecDeltasApplied` pasa. Observed in: el repositorio.
