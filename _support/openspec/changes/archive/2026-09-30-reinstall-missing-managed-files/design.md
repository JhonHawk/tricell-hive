# Diseño

## Contexto verificado

Base: `0292ff4` (`development`), el 2026-09-30. Los datos vienen de una investigación con un subagente, de solo lectura; los números de línea son los de la base.

- **Dónde se decide que algo falta.** `owned()` (`tooling/management/files.go:258-296`) devuelve `ManagedFileMissing` o `ManagedBlockMissing`. Llega a esa función todo el que pasa por `transform()` (`files.go:299`):
  - `transformResource` (`resources.go:59`), desde `BuildPlan` (`plan.go:424`), `prepareTransaction` (`apply.go:234`) y `PlanUnchanged` (`migration.go:132`);
  - `prepareEntries` (`migration.go:373`, `:377`).
- **Voz.** `checkVoiceConflict` (`voice.go:319-334`) se llama desde `voice.go:456`, `:527` y `:590`. `BuildVoicePlan` falla con «no longer has a managed Hive block» (`voice.go:452-454`).
- **Estado.** `Status` (`plan.go:792`, `:851`) y `installationVersion` (`versioning.go:148`) marcan `drift`.
- **Lectura.** Un archivo ausente se lee como `snapshot{}` sin error (`files.go:33`, `resources.go:21`).
- **Directorios.** `missingDirs` (`apply.go:295`) ya vuelve a crear los directorios que falten.
- **Recuperación.** El diario guarda `Before` y `After`. Deshacer un archivo creado desde cero ya funciona: `Before` es «ausente» y `write()` lo borra (`apply.go:680`, `files.go:85`). Un cambio de quitar algo ausente tiene `Before == After`, así que no se escribe nada (`apply.go:360`).
- **Revalidación del plan.** `validatePlan` (`plan.go:623-656`) vuelve a construir el registro a partir de la huella `Expected`. No puede deducir desde el disco que algo faltaba, porque durante Apply ya contiene la imagen final. Por eso el plan tiene que llevar el dato.

## Enfoque

### Contrato

- `Change.Gone bool`, con JSON `gone,omitempty`. Lo mismo en `VoiceChange`.
- `BuildPlan` lo fija cuando `owned()` devuelve `*ManagedFileChangedError` con clase `ManagedFileMissing` o `ManagedBlockMissing`. Se detecta con un ayudante `isMissing(err)` que usa `errors.As`. Los planes viejos sin el campo conservan su ID.
- **`Gone` se comprueba antes de actuar (B2).** El ayudante solo lo acepta si `ch.Before != nil` y `isMissing(owned(s, *ch.Before, m))` es verdadero sobre la foto actual, que `prepareEntries` ya comparó con `Expected` (`migration.go:352`). Si no, falla cerrado con el error de `owned()`. Lo mismo vale para `VoiceChange.Gone`. Así, un plan viejo o manipulado no puede quitar el registro de un archivo que sigue en disco.
- **Formato (N1).** `Change` y `VoiceChange` viajan dentro de los planes guardados con `--out` y dentro del diario (`apply.go:16-33`), así que los dos pueden llevar `"gone": true`. Es un cambio aditivo con `omitempty`: los planes y diarios anteriores conservan su ID y su hash, y no hace falta migración. Límite al volver a una versión anterior: `decodeFile` rechaza campos desconocidos (`files.go:163`), así que un binario viejo falla cerrado con «unknown field» ante un plan o un diario pendiente que lleve `gone`. `state.json` no cambia.

### Instalar y actualizar

Un ayudante compartido recibe la acción del plan (`p.Action`), que tienen a mano sus cuatro llamadores: `plan.go:424`, `apply.go:234`, `migration.go:132` y `migration.go:369`. Con `ch.Gone` ya comprobado:

- Si la acción es `remove`, devuelve la foto sin tocar, sea cual sea `After` (B1). Al quitar un host de un recurso compartido, `After` es el registro con un consumidor menos, y reponerlo contradiría D1-A.
- Si la acción es instalar, llama a `transform(s, nil, ch.After, ...)`, es decir, a la ruta de una instalación nueva:
  - si falta el archivo, lo crea;
  - si falta el bloque, lo añade al final del archivo del usuario;
  - si falta el archivo que contenía el bloque, lo crea con el bloque;
  - si `After` es `nil`, porque la release nueva retira el recurso, no escribe nada.
- `nextRecord` recibe `gone`:
  - `CreatedFile = !s.Exists`;
  - `Leading` se calcula como en una instalación nueva;
  - se conservan los `Consumers` anteriores, para no perder a los demás hosts de un recurso compartido.
- `validatePlan` (`plan.go:632`, `:648`) acepta `old == nil || ch.Gone`.
- Para que la reposición cuente como cambio hace falta `|| ch.Gone` en `cli/install.go:791` y `migration.go:151`. En `migration.go:136` y `apply.go:238` sobra, porque `!same(cur, after)` ya es verdadero, pero no hace daño. Sin él, el resumen dice «already up to date» mientras escribe.
- `CreatedDirs` no duplica entradas (`apply.go:299`).

### Quitar (D1-A)

Si falta lo que había que quitar, el ayudante devuelve la foto tal cual y no escribe nada, también cuando `After` no es `nil` (B1). El registro se elimina, o pierde el consumidor que se quita. Un recurso compartido que otros hosts siguen usando queda sin reponer hasta el próximo `install` o `update`, y mientras tanto `status` lo marca como `drift`. El resumen de quitar (`tui_hosts.go:52-57`) puede listar lo ya borrado bajo «Files to remove»; se acepta así.

### Voz

- `checkVoiceConflict` acepta un bloque de voz ausente. Con `vc.Gone` ya comprobado, `composeVoiceStep` devuelve `base` si `After` es `nil` y hace `insertVoiceSpan` si `After` existe (N2). Sin `Gone` se comporta como hoy: si no, `voice off` dejaría de quitar bloques presentes.
- `addVoiceChangesForInstall` (`voice.go:560`) no salta el cambio cuando `Gone`.
- Se añade `|| vc.Gone` en `apply.go:259`, `migration.go:151` y `cli/voice.go:213`.
- `BuildVoicePlan` (`voice.go:452`), para `set`, omite las rutas sin bloque de Hive. Para `off`, una ruta sin bloque de Hive pero con bloque de voz quita el bloque de voz y su registro, para no dejarlo huérfano; si tampoco tiene bloque de voz, se omite. Solo falla si no queda ninguna ruta, con «run hive install first».
- Queda un caso sin resolver: si la instalación no puede generar la voz (release fija o voz retirada), el bloque ausente sigue en `state` y se ve como `drift`. No bloquea ninguna operación.

### Textos (D2-A)

- `driftRepairText` (`cli/doctor.go:431-437`) pasa a dos casos:
  - si el archivo se editó, deshacer el cambio o corregir los permisos;
  - si se borró, correr `hive install` o `hive update`, que lo reponen.

  Se actualiza su copia en `doctor_test.go:630`.
- `deployment-manager.md:71` y `:156` (la regla «must still match its installed bytes») pasan a decir que lo borrado se repone y lo editado se rechaza.
- `versioned-installation/spec.md` recibe un escenario «Deleted managed file» en la sección de preservación. En este cambio va como delta en `specs/versioned-installation/spec.md`.
- Los textos de «ausente» de `ManagedFileChangedError` siguen existiendo como protección cuando se aplica un plan viejo.

### Cambios tras `/code-review`

- **F1:** `validatePlan` aplica la comprobación de separador de una instalación nueva solo al instalar.
- **F3:** si faltan los marcadores pero el texto del bloque sigue en el archivo, se trata como editado y se rechaza, para no duplicar las reglas.
- **F4:** al quitar lo ya borrado, un archivo que Hive creó y queda vacío se borra, como en un quitar normal.
- **F5:** el bloque de Hive se repone delante del bloque de voz.
- **F8:** `voice set` registra en el plan qué archivos omite, y la CLI los muestra.
- **F6 y F7:** los resúmenes no cuentan como «a borrar» lo que ya no existe.
- **F9 y F10:** una sola comprobación compartida de `Gone`.
- **Límite de F5:** un bloque repuesto en un archivo sin salto de línea final queda con `Leading ""`; en la instalación original era `"\n"`. Un «quitar» posterior deja ese salto de línea en el archivo.
- **F2** queda como límite aceptado: con una release fijada o sin voces en la fuente, `install` no repone un bloque de voz borrado.

### Riesgos aceptados

- **Marcadores borrados, cuerpo conservado.** Tras F3 ya no es un riesgo: si el texto del bloque sigue en el archivo sin sus marcadores, se rechaza como edición.
- **Un solo marcador borrado.** Sigue contando como edición (marcadores mal formados) y se rechaza.
- **Borrar como forma de optar por no tener algo.** No existe tal mecanismo; el próximo `update` lo repone. Es lo que pidió el usuario.
