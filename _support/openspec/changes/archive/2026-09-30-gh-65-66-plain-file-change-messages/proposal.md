# Mensajes claros cuando un archivo de Hive se cambió a mano

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `development` el 2026-09-30 con el [PR #75](https://github.com/JhonHawk/tricell-hive-private/pull/75) (merge `3f24787`); #65 y #66 cerrados |
| Tracker · GitHub Issues | • [#65 — Vista CLIs: el aviso de archivo cambiado dice que quitar se rechaza, pero Uninstall all funciona](https://github.com/JhonHawk/tricell-hive-private/issues/65)<br>• [#66 — Vista CLIs: agregar Cursor con una skill editada muestra «modified managed skill» sin explicación](https://github.com/JhonHawk/tricell-hive-private/issues/66) |
| Git | `automatic` · rama `feat/gh-65-66-plain-file-change-messages` desde `development`, PR a `development`, `/code-review` antes del push, merge · sin CI |
| Verificación | pruebas por paquete · suite local · `hive-review-ux` y `hive-verify-change` en `tmux` |
| Siguiente paso | Ninguno en este cambio |

## Objetivo

Hay dos situaciones distintas en las que un archivo difiere de lo que Hive espera:

- **(a) Una ruta que Hive usó en una versión anterior y ya no gestiona.** Hoy solo instalar y actualizar rechazan este caso, porque solo ellos buscan instalaciones antiguas. Aun así, el aviso de la vista de CLIs dice que también se rechaza quitar, y `deployment-manager.md:101` lo repite. Es falso: Uninstall all funciona (#65).
- **(b) Un archivo que Hive gestiona hoy y que se editó a mano.** Instalar y quitar lo rechazan con un error crudo, por ejemplo `<ruta>: modified managed skill: <ruta>`. El mensaje repite la ruta, no dice en palabras del usuario qué pasó ni cómo salir, y en la vista de CLIs aparece junto al aviso, duplicado (#66).

Este cambio corrige el aviso para que nombre solo lo que de verdad se rechaza (D1-A). Además, el gestor pasa a devolver un error tipado en palabras llanas, con la ruta una sola vez (D2-A). Así la vista de CLIs, `hive install` y `hive plan remove` muestran el mismo mensaje claro, y la vista deja de duplicarlo. `hive update`, Releases y Voice imprimen el error tal cual, así que heredan el texto nuevo; no hay criterio aparte para ellos.

## Alcance y aceptación

Incluido:

- El aviso de la vista de CLIs (`legacyScanNote`, `tooling/cli/tui_hosts_view.go:66-75`), tanto el de archivo cambiado como el genérico, y la descripción de la vista en `deployment-manager.md`.
- Un error tipado para los cinco fallos de `owned()` (`tooling/management/files.go:257-283`): archivo gestionado ausente, enlace cambiado, tipo cambiado, skill o agente cambiado, y bloque gestionado cambiado o ausente. Incluye su mensaje en palabras llanas, y que `plan.go:425` no vuelva a anteponer la ruta cuando el error ya la lleva.

Excluido:

- Cambiar qué operaciones rechazan. Quitar sigue funcionando en el caso (a), porque la especificación dice que quitar solo borra lo que Hive gestiona. Instalar y quitar siguen rechazando en el caso (b).
- `hive doctor`, que ya explica el caso (b) (`doctor.go:436`).
- Los errores de marcadores de `blockRange` (`files.go:225-235`), que pasan por `owned()` sin tipo.
- Qué hacer cuando falta un archivo gestionado y no hay copia de seguridad: hoy instalar y quitar se rechazan y no hay salida. Se informa como hallazgo aparte.

Restricciones:

- Ninguna operación cambia de comportamiento: solo cambia el texto de los errores y del aviso.
- `errors.As` sigue permitiendo reconocer el error tipado.
- `go vet ./...` y `go test ./...` siguen en verde.

Criterios:

- AC1. El aviso de la vista de CLIs nombra instalar y actualizar como las operaciones que se rechazan, y avisa aparte de que quitar también puede rechazarse si Hive instaló ese archivo (D3-A, redacción refinada tras la revisión de experiencia de uso). El aviso genérico de fallo de búsqueda dice «installing or updating may be refused». En el caso (a), con hosts registrados y un archivo del usuario en una ruta antigua que el catálogo actual no gestiona, Uninstall all se aplica. *Falso en la base cuando* el aviso termina en «before installing or removing» (`tui_hosts_view.go:73`) y el genérico dice «installing or removing may be refused» (`:68`).
- AC2. `deployment-manager.md` describe que el caso (a) rechaza instalar y actualizar, y que quitar solo se rechaza para archivos que Hive gestiona hoy. *Falso en la base cuando* `deployment-manager.md:101` dice que quitar se rechaza hasta restaurar el archivo.
- AC3. Los fallos de `owned()` devuelven un error tipado que `errors.As` reconoce, con la ruta una sola vez y en palabras llanas, con tres redacciones:
  - un archivo cambiado dice «<ruta> differs from what Hive expects there; undo the change, restore it from a backup, or move your own file elsewhere»;
  - un bloque gestionado cambiado o ausente dice que el bloque de Hive en <ruta> difiere de lo que Hive escribió y que se deshaga el cambio dentro del bloque;
  - un archivo gestionado ausente dice «Hive installed <ruta> and it is no longer there; restore it from a backup», en línea con `doctor.go:436-437`.

  Se comprueba con una skill editada en `hive plan remove`, y en `hive install` agregando Cursor, que no pasa por la búsqueda de instalaciones antiguas. El conflicto de voz (`voice.go:321`) tampoco repite la ruta. *Falso en la base cuando* el mensaje es `<ruta>: modified managed skill: <ruta>`.
- AC4. En la vista de CLIs, a 80×24 y a 120×40, agregar Cursor con una skill compartida editada muestra el rechazo:
  - la frase «differs from what Hive expects there» aparece una sola vez en la pantalla aplanada;
  - la ruta aparece una sola vez, contada sobre el texto sin espacios;
  - no aparecen «press m to read all» ni «modified managed skill»;
  - `assertFits` pasa.

  *Falso en la base cuando* la vista muestra «modified managed skill» con la ruta dos veces, partida en 6 líneas a 80×24, debajo del aviso.

## Entrega

Siguen en vigor las decisiones de esta sesión: entrega automática, `/code-review` de Claude Code antes del push y sin CI. Decisiones del usuario del 2026-09-30:

- **D1-A:** se corrige el aviso.
- **D2-A:** el error tipado vive en el gestor.
- **D3-A:** el aviso usa una redacción neutra, porque la vista no distingue el caso (a) del (b). Texto final, tras la revisión de experiencia de uso (M2) y `/code-review` (F4): «… before installing or updating. Removing may also be refused if Hive installed that file.».

Secuencia:

1. Worktree con `feat/gh-65-66-plain-file-change-messages` desde `origin/development`.
2. T1 y T2 en paralelo, y después T3.
3. T4: revisión de experiencia de uso y verificación de punta a punta en `tmux`.
4. Suite local.
5. `/code-review`.
6. `gh pr create --base development`.
7. Merge.
8. Recompilar `hive`, porque cambia `tooling/`.
9. Cerrar #65 y #66.

El registro se cierra con un PR pequeño hacia `development`, sin revisión dedicada, al que yo hago merge.
