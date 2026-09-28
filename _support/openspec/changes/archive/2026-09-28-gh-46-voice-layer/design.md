# Diseño

## Contexto verificado

Inspeccionado el 2026-09-27 sobre `6fd741c` (`rebuild/harness-engineering`). La investigación la hizo un subagente `sdd-explore`, y el hilo principal comprobó los puntos marcados.

- **Un solo par de marcadores:** `Begin` y `End` son constantes: `<!-- === TRICELL HIVE RULES:BEGIN === -->` y su cierre (`tooling/management/types.go:18-19`). `blockRange` (`files.go:190-219`) cuenta esas dos cadenas y falla si hay más de una o si están invertidas; un par con otro texto le es invisible. `managedBlock` (`files.go:220`) arma el bloque respetando los finales de línea del archivo. `owned` y `transform` (`files.go:229` y `:257`) comprueban y reemplazan ese único tramo. Comprobado por el hilo principal.
- **Un recurso por ruta:** `groups()` (`ownership.go:92-106`) agrupa los destinos por `Path` y falla con «incompatible shared destination» si dos destinos en la misma ruta difieren. `State.Records` es un mapa indexado por ruta (`types.go:52`). Los CLIs que comparten archivo, como Grok y Claude en `~/.claude/CLAUDE.md`, son consumidores del mismo registro. Comprobado.
- **La instalación retira lo que no está en la release:** con `action == "install"`, `desiredResources` agrega con `Retire: true` todo registro de `State.Records` cuya ruta no esté en la release nueva y cuyos consumidores estén elegidos (`ownership.go:179-208`). Un bloque de voz guardado como registro normal desaparecería en el siguiente `hive update`. Comprobado.
- **Precedente de ediciones fuera del catálogo:** la migración del Hive anterior agrega `Plan.Legacy []legacy.Edit`, y `overlayRead` (`migration.go:85-91`) compone esa edición con el cambio del bloque sobre el mismo archivo dentro de un plan. Esas ediciones guardan el archivo completo, así que `SavePlan` rechaza guardarlas porque contienen configuración privada (`plan.go:491-494`). Comprobado.
- **Estado fuera de las releases:** `State.Installations` (recibos de versión por consumidor, `versioning.go:20-35`) y `State.Migrations` se guardan en `state.json` sin pasar por `desiredResources`.
- **Validación del catálogo:** `validateRelease` rechaza cualquier archivo que contenga `Begin` o `End` (`plan.go:307-309`).
- **Paquete offline:** incluye `content/` completo (`tooling/package/main.go:113`), así que `content/voices/` viaja en el paquete sin cambios al empaquetador.
- **Esquema:** `state.json`, los planes y los journals usan el esquema 6. Los planes y journals se decodifican con `DisallowUnknownFields` (`decodeFile`, `files.go:134`), pero `state.json` no: `readState` usa `json.Unmarshal` (`files.go:152`) y solo rechaza versiones desconocidas (`:155`). Comprobado.
- **La transacción hoy:** `prepareTransaction` rearma el estado nuevo copiando solo `Records`, `CreatedDirs` y `Migrations` (`apply.go:182-187`). `prepareEntries` crea las entradas del journal solo desde `p.Changes` (`migration.go:323`). `validatePlan` acepta solo `install` y `remove`. `updateProductState` borra los recibos de versión de los CLIs elegidos cuando el plan no trae producto (`versioning.go:107`). `Recover` solo sabe invertir el bloque de Hive, y da por bueno un archivo cuyo bloque ya coincide con su estado anterior (`apply.go:534-559`). Comprobado.
- **`status`:** `StatusEntry` no tiene etiquetas JSON (`types.go:114`); `hive status` la imprime tal cual.

## Diseño elegido

### Bloque de voz

- **Marcadores propios:** `<!-- === TRICELL HIVE VOICE:BEGIN === -->` y `<!-- === TRICELL HIVE VOICE:END === -->`. Las funciones de `files.go` pasan a recibir el par de marcadores (`blockRange`, `managedBlock`, `owned`, `transform`), sin cambiar su comportamiento para el par de Hive; las pruebas de caracterización lo fijan antes.
- **Ubicación:** justo después de la línea de cierre del bloque de Hive. Si hace falta un separador, se registra como parte del tramo gestionado, igual que el separador del bloque de Hive.
- **Requisito:** solo se escribe en un archivo que ya tiene el bloque de Hive gestionado por este gestor. Sin él, `voice set` falla antes de escribir.
- **Validación:** `validateRelease` y el catálogo de voces rechazan cualquier texto que contenga alguno de los cuatro marcadores.

### Modelo

- **`Plan.Voice []VoiceChange`** (JSON `voice`, `omitempty`). Cada cambio lleva:
  - la ruta y sus consumidores;
  - el tramo gestionado antes y después (`Before`, `After *VoiceSpan`, cada uno con su texto generado y su `SourceHash`), igual que `Change` con `Record`.

  Guarda solo el tramo, nunca el archivo completo, así que un plan con voz sí se puede guardar con `--out`. La huella esperada del archivo es la del `Change` del mismo archivo si lo hay, o la propia si solo cambia la voz.
- **`Plan.VoiceSetting *VoiceSetting`** (JSON `voice_setting`, `omitempty`): la elección que `apply` escribirá en el estado. `voice off` la deja vacía y pide borrarla.
- **`State.Voice *VoiceSetting`** (JSON `voice`, `omitempty`) guarda la elección:
  - `ID`;
  - `Address`: `sir`, `name` o `none`;
  - `Name`: solo con `name`;
  - `Intensity`: `subtle` o `marked`.
- **`State.VoiceSpans map[string]VoiceSpan`** (JSON `voice_spans`, `omitempty`) guarda, por ruta, el tramo escrito, su `SourceHash` y sus consumidores. El hash va por tramo y no por home, para que un `plan install` de algunos CLIs no deje a los demás con el texto viejo sin enterarse.
- **Rutas de la voz:** salen de los registros de bloque de `State.Records`, no de `resolve()`. Así el tramo va siempre al archivo que el CLI lee de verdad, incluido el override de Codex y Grok frente a `CLAUDE_CONFIG_DIR`.
- **Esquema (D13-A):** se mantiene en 6, sin migración.
  - Un estado sin voz queda idéntico byte a byte.
  - Un gestor anterior sobre un estado con voz descarta en silencio `voice` y `voice_spans` (`readState` no rechaza campos desconocidos) y deja los tramos en los archivos.
  - El gestor nuevo encuentra entonces un tramo de voz que no está registrado y lo trata como conflicto sin pisarlo. Se documenta en T6.
- **Por qué no un tipo más de recurso:** exigiría quitar la regla de un recurso por ruta de `groups()` y excluir ese tipo del retiro de `desiredResources`. Con un campo aparte, el catálogo de la release y el retiro no se tocan, así que una release nunca puede borrar la voz.

### Generación del texto

- **Fuentes:**
  - `content/voices/preamble.md`: común a todas las voces. Se sostiene solo, sin depender de su posición, porque Grok da prioridad al texto que aparece después (`repository-and-distribution.md:56`) y este bloque va detrás del de Hive. Dice, en este orden:
    1. **Prioridad:** esta sección va después de las reglas de Hive pero nunca las anula; ante cualquier conflicto se sigue la regla de Hive.
    2. **Lo que la voz no cambia**, como lista explícita:
       - la primera frase lleva el resultado;
       - las secciones con etiqueta, las listas y las tablas;
       - el vocabulario llano, sin metáforas;
       - la glosa de los IDs;
       - el desacuerdo razonado, sin halagos ni deferencia;
       - el idioma de la sesión.

       Una cláusula general no basta para los modelos del perfil `execution`.
    3. **Dónde entra el tono:** en las transiciones, los cierres y la forma de tratamiento, nunca en el contenido de los hallazgos ni en las decisiones.
    4. **Cuándo ignorarla,** escrito desde el punto de vista de quien lee: si otro agente te lanzó, o si lo que escribes va a un archivo, commit, pull request, ticket, especificación u otro agente, ignora esta sección entera.

       El texto no puede hacer cumplir esa condición, así que un subagente que lee el archivo global puede adoptar la voz. Es un límite conocido: Pi documenta que sus hijos heredan el contexto global (`agent-delivery.md:65`).
  - `content/voices/<id>.md`: la descripción del estilo de cada voz.
- **Qué contiene el bloque:**
  1. El preámbulo.
  2. La descripción de la voz.
  3. Una línea de tratamiento:
     - `sir`: usar el tratamiento formal que eligió el usuario, en el idioma de la sesión, por ejemplo «señor», sin marcar género en el resto de las frases;
     - `name`: dirigirse al usuario por el nombre dado;
     - `none`: ningún tratamiento.
  4. Una línea de intensidad:
     - `subtle`: la voz se nota en la elección de palabras y en giros ocasionales;
     - `marked`: se nota en la mayoría de los mensajes.
- **Validación del nombre:** es entrada del usuario que termina en un archivo de instrucciones. Se aceptan letras, espacios, apóstrofos y guiones, hasta 40 caracteres, sin saltos de línea ni marcadores.
- **Voces disponibles:** son los archivos `content/voices/*.md` salvo `preamble.md`. Los ID son `jarvis`, `senior-direct` y `mentor`.

### Operaciones

- **`hive voice set <id> [--address sir|name|none] [--name NAME] [--intensity subtle|marked] [--source .] [--home DIR] [--state-dir DIR] [--dry-run] [--out FILE]`**
  - Por defecto, `--address none` y `--intensity subtle`.
  - Toma los archivos de los registros de bloque de los CLIs registrados en alcance de usuario.
  - Arma un plan con `Action` igual a `voice`, sin `Change`, con los cambios de voz y el `VoiceSetting`.
  - Lo confirma en la terminal con el mismo patrón de `update`: resumen, confirmación, `--dry-run` y `--out` sin terminal.
- **`hive voice off [...]`:** plan `voice` que quita cada tramo y borra `State.Voice`.
- **`hive voice list [--source .]`:** imprime cada ID con la primera línea de su descripción.
- **`plan install`, que es también la ruta de `hive update` (D12-A):** si `State.Voice` existe y la fuente trae `content/voices/`, genera el texto de la voz para cada archivo elegido.
  - Agrega un cambio de voz cuando el `SourceHash` del tramo difiere o cuando falta el tramo.
  - Con `--release ID` no hay texto de voz en la fuente, así que los tramos quedan intactos.
  - El resumen de `update` muestra una línea cuando el plan trae cambios de voz.
- **`plan remove` (D11-A):** quita el tramo de voz del archivo cuyo último consumidor se retira. Si ya no queda ningún tramo, borra `State.Voice`.
- **`status`:**
  - agrega una fila por tramo, con `Kind` igual a `voice`;
  - el campo nuevo `Voice` (con etiqueta JSON `omitempty`) lleva el ID, el tratamiento y la intensidad, por ejemplo `jarvis (sir, subtle)`;
  - `Status` vale `installed` o `drift` (tramo editado a mano), el mismo vocabulario que las demás filas;
  - el orden desempata por `Kind` cuando la ruta coincide.
- **Conflicto:** un tramo de voz editado a mano, o uno presente en el archivo pero sin registrar, hace fallar `voice set`, `voice off`, `update` y `plan remove` sobre ese archivo con un conflicto, y conserva la edición.

### Aplicación y recuperación

- **Validación del plan:**
  - `validatePlan` acepta `Action` igual a `voice` con cero `Change` y al menos un cambio de voz, y en `install` y `remove` también acepta cambios de voz;
  - comprueba que cada tramo nuevo no contiene marcadores y que el `Before` de cada cambio coincide con `State.VoiceSpans`;
  - `updateProductState` no toca los recibos de versión en un plan `voice`.
- **Estado nuevo:** `prepareTransaction` copia `Voice` y `VoiceSpans` al estado nuevo y les aplica los cambios de voz y el `VoiceSetting` del plan. El cálculo de `changed` (`apply.go:189` y `:231`) y `PlanUnchanged` cuentan los cambios de voz.
- **Una entrada del journal por archivo:** `prepareEntries` crea una entrada por cada ruta con `Change`, cambio de voz o ambos.
  - Imagen de antes: el archivo actual, leído con `overlayRead` en un plan de migración.
  - Imagen de después: el resultado de aplicar los dos cambios en orden.
  - El archivo se escribe una sola vez.
- **Orden de los dos cambios:**
  - al instalar o actualizar: primero el bloque de Hive y después la voz;
  - al quitar y al apagar la voz: primero la voz y después Hive. La recuperación invierte exactamente ese orden; ver «Recuperación».

  Así un archivo creado por Hive queda vacío y se borra al quitar los dos bloques.
- **Recuperación:** el inverso deshace exactamente la composición hacia adelante, así que su orden depende de la operación. Corregido tras la verificación de T2, que demostró con una prueba que un orden fijo de «voz primero» reconstruye con los bloques invertidos un archivo creado por Hive que se había borrado.
  - `install`: deshace primero la voz y después el bloque de Hive.
  - `remove` y `voice`: deshace primero Hive y después la voz.
  - El atajo de «el bloque ya coincide con su estado anterior» solo vale si el tramo de voz también coincide con su `Before`.
  - Se conserva el texto que el usuario escribió fuera de los bloques después de la interrupción.
- **Ubicación al insertar:** el tramo se inserta justo después de la línea de cierre del bloque de Hive, nunca al final del archivo. Así queda bien formado aunque el usuario tenga texto después del bloque de Hive o la última línea no termine en salto de línea.
- **Alcance de `install` y `remove`:** los cambios de voz se limitan al alcance de usuario y a los archivos cuyos consumidores incluyen un CLI elegido. Un CLI que recibe su bloque de Hive por primera vez en el mismo plan recibe también la voz, en la misma escritura.
- **Huella esperada:** se comprueba antes de escribir, como en los cambios actuales. Un plan viejo se rechaza.

## Límites de seguridad

- **Único dato externo:** el nombre del tratamiento, validado como arriba.
- **Sin procesos nuevos ni dependencias.**
- **Datos privados:** el plan y el journal guardan solo tramos gestionados y las imágenes de archivo que ya guardan hoy. No se agrega texto privado del usuario a un plan que se pueda guardar.

## Compatibilidad

- **`Plan` y `State` ganan campos con `omitempty`.** Los planes y estados sin voz conservan sus bytes y su ID.
- **El gestor anterior** lee un estado con voz pero descarta la voz (D13-A). Ocurre solo si se vuelve a un commit anterior del checkout; hay que ejecutar `hive voice off` antes.
- **El ID de las releases no cambia:** la voz no entra en el catálogo de la release.
