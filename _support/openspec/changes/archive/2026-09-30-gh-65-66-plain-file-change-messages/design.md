# Diseño

## Contexto verificado

Base: `8acdb6e` (`development`), el 2026-09-30. Los datos vienen de una investigación con un subagente, que reprodujo los dos estados; las rutas clave las comprobó el orquestador.

- **Aviso.** `legacyScanNote` (`tooling/cli/tui_hosts_view.go:66-75`) arma el aviso cuando la búsqueda de instalaciones antiguas devuelve `*legacy.ModifiedFileError` (`tooling/legacy/scan.go:38-46`, `:329`). En ese caso añade «before installing or removing»; para cualquier otro fallo de búsqueda dice «installing or removing may be refused» (`:68`). `changedFilePhrase` (`:62`) es la frase que el aviso comparte con el rechazo, y `scanNoteLines` (`:912`) oculta la línea «Detail:» cuando el mensaje de error la contiene.
- **Quién busca instalaciones antiguas.** `scanLegacy` (`tooling/management/migration.go:521`) corre solo cuando la acción es `install` en el ámbito de usuario (`migration.go:50`, `plan.go:354`), en `validateMigration`, en la comprobación posterior a la escritura (`apply.go:391`) y en `DetectLegacyHosts`. Por eso `plan remove`, Uninstall all, `recover` y las voces no rechazan el caso (a). `BuildPlan("remove")` devuelve nil en ese estado; está verificado en vivo.
- **Caso (b).** `owned()` (`tooling/management/files.go:257-283`) devuelve cinco errores crudos:
  - `managed file is missing: <ruta>`;
  - `modified managed symlink: <ruta>`;
  - `managed resource type changed: <ruta>`;
  - `modified managed skill: <ruta>`, para skills y agentes;
  - `modified or missing managed block: <ruta>`.

  `transformResource` los envuelve como `"%s: %w", t.Path` (`plan.go:425`), y por eso la ruta sale dos veces. Los llama `BuildPlan` para instalar y para quitar (`plan.go:424`), `Apply` (`apply.go:234`) y la migración (`migration.go:132`, `:369`).
- **Cómo muestra la vista los errores.** Imprime `err.Error()` sin traducir en `failInstall` (`tui_hosts_view.go:596-602`), `onRemovePlanned` (`:497`) y `finish` (`:290-294`).
- **Dependencias de esos textos.** Solo `tui_error_states_test.go:90` fija «modified managed skill», y un comentario lo menciona en `voice_review_test.go:209`. Ningún código compara esos textos.
- **Especificación.** `deployment-manager.md`, en la sección de preservación, dice que quitar solo borra el tramo gestionado y deja intactos los archivos del usuario fuera de los destinos gestionados. `installer.md:85-90` dice que los candidatos editados detienen el plan antes de escribir al instalar. `versioned-installation/spec.md` no tiene requisitos sobre estos mensajes.

## Enfoque

### Aviso (D1-A y D3-A)

La vista no puede distinguir los dos casos: la búsqueda informa como `legacy.ModifiedFileError` tanto una ruta antigua como un archivo gestionado que se editó (`migration.go:521-533`). Por eso el aviso es neutro:

- el de archivo cambiado dice «<ruta> differs from what Hive expects there. Undo the change, restore it from a backup, or move your own file elsewhere before installing or updating. Removing may also be refused if Hive installed that file.» (redacción final tras M2 y F4);
- el genérico dice «Hive could not check for a legacy installation, so installing or updating may be refused. The Detail line says why.».

En `deployment-manager.md`, en el párrafo de CLIs, se cambian las dos citas del aviso y la frase que dice que quitar o instalar se rechaza hasta restaurar el archivo. Queda así: el caso (a) rechaza instalar y actualizar; quitar se rechaza con su propio mensaje solo cuando el archivo es uno que Hive gestiona hoy.

### Error tipado (D2-A)

En `tooling/management` se añade un tipo exportado, por ejemplo `ManagedFileChangedError`, con `Path` y una clase que distingue tres redacciones. `owned()` lo devuelve en vez de `fmt.Errorf`:

- **Archivo cambiado** (skill, agente, enlace o tipo cambiado): «<ruta> differs from what Hive expects there; undo the change, restore it from a backup, or move your own file elsewhere». Es el mismo texto que `legacy.ModifiedFileError`. Si se puede, sale de una sola fuente: `management` importa ya `legacy`, así los dos textos no se separan con el tiempo.
- **Bloque gestionado cambiado o ausente**, dentro de un archivo del usuario como `CLAUDE.md` o `AGENTS.md`: «the Hive block in <ruta> differs from what Hive wrote; undo the change inside the block». Mover o restaurar el archivo entero perdería las ediciones del usuario. Hoy `owned()` junta el bloque ausente (`a < 0`) con el contenido distinto (`files.go:278-280`); la redacción puede cubrir los dos o separarlos.
- **Archivo gestionado ausente**: «Hive installed <ruta> and it is no longer there; restore it from a backup», igual que `driftRepairText` (`doctor.go:436-437`).

Las tres redacciones contienen la ruta una sola vez. La de «archivo cambiado» contiene `changedFilePhrase`, así que `scanNoteLines` (`tui_hosts_view.go:912`) oculta el aviso cuando el rechazo lo repite.

`transformResource` (`plan.go:425`) y `voice.go:321` no anteponen la ruta cuando el error es de este tipo. Con otros errores la siguen anteponiendo, porque algunos, como los de marcadores de `blockRange`, no la llevan. `apply.go:234` y la migración (`migration.go:132`, `:369`) devuelven el error sin envolver, así que ganan la ruta. `errors.As` funciona en todas las capas, porque todas usan `%w`.

Se descarta traducir el mensaje solo en la vista (D2-B), porque los comandos de texto seguirían mostrando el error crudo.

### Vista

No hace falta tocar el código de la vista. Las pruebas se ajustan:

- `tui_error_states_test.go:63` corresponde al caso (b), y `:226` y `:358` al caso (a). Las tres exigen la redacción neutra, y `:139` exige el aviso genérico nuevo.
- `:90` deja de fijar «modified managed skill».
- Una prueba nueva reproduce #66 con `driftedEnv` para todos los hosts menos Cursor, `openDriftedCLIs`, `toggle("cursor")` y `a`. `hostsTestDeps` ya da todos los hosts por detectados, así que no hace falta un `lookPath` falso. Comprueba las afirmaciones de AC4 a los dos tamaños.
- Para Uninstall all en el caso (a) no sirve `legacyPathEnv`, que no tiene filas y por tanto no tiene tecla `u`. Hace falta un entorno con hosts registrados y un archivo del usuario en una ruta antigua que el catálogo actual no gestione.
