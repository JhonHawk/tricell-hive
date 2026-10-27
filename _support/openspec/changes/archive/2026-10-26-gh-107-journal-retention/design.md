# Diseño: retención de journals

## Verified context

Revisión inspeccionada: `development` en `269f1a92` (2026-10-26).

- **Contenido de un journal.** `journal` (`tooling/management/apply.go:27-40`) guarda `Plan`, `Entries` con `Before`/`After` de bytes completos, `BeforeState` y `AfterState`. Se escribe en `transactions/<id>.json` (`apply.go:614`).
- **Lo que pasa al confirmar.** `commitTransaction` marca la fase `committed`, guarda el journal y elimina solo `pending.json` (`apply.go:454-458`). `Recover` marca `recovered` y también conserva el journal. Si encuentra uno `committed` o `recovered`, solo quita `pending.json` (`apply.go:683-685`).
- **Quién lee journals fuera de los tests.** `Recover`, a través de `pending.json` (`apply.go:673`). Onboarding, por el `CoreID`: `discardUnstartedOnboarding` (`onboarding.go:249`) y `RecoverOnboarding` (`onboarding.go:443`), que con el padre en `prepared` o `core_pending` necesita el journal del core para saber si llegó a confirmarse. Nada más lee `transactions/` (`rg '"transactions"' tooling --glob '!*_test.go'`).
- **Dos caminos de aplicación.**
  - El no anidado: `Engine.Apply` con `!e.nested`, bajo el lock de `preflight` (`apply.go:587-598`). Lo usan `hive update`, `hive apply`, las vistas de la TUI y `Onboard` sin pasos (`onboarding.go:384-387`).
  - El anidado: `Onboard` con pasos toma el lock y aplica el core con `nested = true` (`onboarding.go:304-311`).
- **Avisos que no fallan la operación.** `recordSourceCommit` ya devuelve un aviso como texto tras el commit (`apply.go:548-560`), y `Apply` lo añade al resultado. Este es el patrón que se reutiliza.
- **Medición del 2026-10-26.** Son 262 journals y 2,9 GB, con una mediana de 12,3 MB por journal. `releases/` ocupa 173 MB y `onboarding/` tiene 0 archivos.
- **Tests que leen journals.** `codex_wait_floor_remove_test.go:264` necesita al menos un journal tras instalar. `catalog_test.go:279` lee el de una aplicación fallida. Con menos de 11 aplicaciones por test, ninguno pierde su journal; queda por confirmar al correr el paquete.

## Proposed design

**Función.** `pruneJournals(stateDir string) string` en `tooling/management`, junto a `finishApply`, con la constante `retainedJournals = 10`. Devuelve un aviso vacío o descriptivo y nunca un error. Se llama con el lock tomado:

1. Al final de `Engine.Apply` cuando `!e.nested`, después de `finishApply` y solo si la operación confirmó cambios y `finishApply` no falló (`apply.go:629`). Su aviso se une al de `recordSourceCommit`, añadido al ID del resultado como `"; " + aviso`.
2. Al final de `Onboard` con pasos, después de `finishOnboarding` (`onboarding.go:412`), que ya eliminó `onboarding-pending.json`, y todavía bajo su lock (`onboarding.go:395`). Corre también en la rama `partial`, antes de su `return`. El aviso se añade a `OnboardingResult.ID` igual que en `Apply`, que `install.go:457/469` ya imprime, sin campos nuevos en el struct persistido.

Estos dos puntos cubren todas las rutas que confirman un core: todo `startTransaction` vive en `Engine.Apply`, y `Recover` y `RecoverOnboarding` no confirman cores nuevos (revisión del plan, 2026-10-26).

**Reglas de selección.** Se aplican en este orden, y la poda es conservadora:

- **Qué archivos.** Solo entradas regulares de `transactions/` con nombre `<32 hex>.json`. Los temporales `.hive-write-*` siguen bajo `removeOrphanWrites`.
- **Qué se protege.** El journal de la operación que acaba de confirmar (su ID se pasa explícitamente), el ID de `pending.json` y el `CoreID` de `onboarding-pending.json`, si existen. Si alguno de esos archivos existe pero no se puede leer, la poda no borra nada y devuelve un aviso. En los dos puntos de llamada esos archivos ya se eliminaron, así que esta protección solo cubre estados raros; los tests la ejercitan llamando a `pruneJournals` directamente.
- **Cómo se lee la fase.** Con `os.Open` tras `target.Safe`, sin `read` ni `decodeFile`, que cargan el archivo entero. Un `json.Decoder` exige que las tres primeras claves del objeto sean exactamente `Version`, `ID` y `Phase` (orden de `apply.go:28-29`, serializado con `MarshalIndent` en `types.go:294`), con valores escalares. Si no aparecen así, el journal se conserva y no se sigue leyendo. Así se leen unos cientos de bytes por archivo. Los formatos v1–v5 usan las mismas claves (`shared_test.go:461`).
- **Qué se puede borrar.** Solo journals en fase `committed` o `recovered` cuyo `ID` coincida con el nombre del archivo.
- **Qué se conserva.** Los `retainedJournals` más recientes entre los borrables, ordenados por `ModTime` con desempate por nombre. `Recover` reescribe el journal recuperado (`apply.go:826-827`), así que su fecha cuenta como la de la recuperación, lo cual es aceptable. El resto se elimina con `os.Remove`. Si un borrado falla, se cuenta en el aviso y se sigue.

**Por qué al final y no al confirmar.** Borrar el journal al confirmar rompería `RecoverOnboarding`, que con el padre en `core_pending` necesita el journal confirmado del core (`onboarding.go:443`), y `discardUnstartedOnboarding`. Podar después de terminar, con exclusión explícita de lo pendiente, mantiene intactas ambas recuperaciones.

**Por qué por cantidad y no por días.** Al ritmo medido, 30 días equivalen a varios GB. Diez journals acotan a unos 120 MB y dejan una ventana para inspeccionar las últimas operaciones.

**Texto y documentación.** `Private backups: <dir>` (`tooling/cli/install.go:825`, `tooling/cli/tui_hosts.go:51` y el literal de `tui_wrap_test.go:175`) pasa a `Recovery journals (newest 10 kept): <dir>`. Se actualizan `deployment-manager.md:451` y `:456`, este último conservando que el historial de releases se retiene (los releases no se podan), y `README.md:146`, que menciona "private backup locations". La copia `config.toml.hive-backup` de Codex sigue siendo un backup y no cambia.

**Datos existentes.** Son archivos locales en la máquina de cada usuario, sin migración aparte. La primera operación tras actualizar poda lo acumulado (en la máquina medida, unos 252 archivos). No hay forma de restaurar lo borrado, y no se necesita: ningún código lo leía.
