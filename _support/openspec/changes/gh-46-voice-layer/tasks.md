# Tareas

**Commit base:** `6fd741c`. Worktree `.claude/worktrees/gh-46-voice`, rama local `feat/gh-46-voice-layer`; se publica directo a `rebuild/harness-engineering` (D7-A). La suite completa con `-race` pasó en `390a0cb`; `6fd741c` solo agrega documentación.

## Orden y ejecución

- **Primera ronda, en paralelo:** T1 (el gestor, en `tooling/management`) y T3 (los textos de voz, en `content/voices`) escriben en archivos distintos.
- **Después:** T2 depende de T1 y la hace el mismo hijo en secuencia. T4 depende de T2 y T3.
- **Al final:** T5 (validación humana) y T6 (documentación y prueba local) van en el hilo principal.

Cada hijo recibe `AGENTS.md`, este cambio y [security boundaries](../../../../content/skills/flow-build/references/security-boundaries.md) (T2 y T4, por el nombre que escribe el usuario).

### T1 — Marcadores por parámetro y catálogo de voces en el gestor

- [x] Las funciones de bloque reciben el par de marcadores, y el gestor lee, valida y genera el texto de una voz.
  - Ronda 1, AC9 y AC10 no cumplidos: `RenderVoice` acepta el ID `preamble`; las pruebas no dependen de la protección contra rutas con `../`, no fijan cada marcador por separado y no afirman que el bloque empieza con el preámbulo ni el orden y la cantidad de las líneas de tratamiento e intensidad.
  - **Ronda 2:** AC9 y AC10 cumplidos. `review-task` repitió las roturas deliberadas del código que antes nadie detectaba, y ahora cada una hace fallar una prueba. `go test -race ./tooling/management` pasa. En la ronda también se agregaron: el nombre del bloque en los errores (con los textos de Hive idénticos a la base), la versión del formato dentro del hash y el rechazo de un nombre con un tratamiento que no es `name`.

**Closes:** AC9 (la validación del catálogo y de los parámetros) y AC10 (que el bloque generado empiece con el preámbulo).

**Depends on:** ninguna.

**Locations:**
- `tooling/management/types.go`: marcadores de voz, `VoiceSetting` y `VoiceSpan`.
- `tooling/management/files.go`: `blockRange`, `managedBlock`, `owned`, `transform`.
- `tooling/management/plan.go`: `validateRelease`.
- Nuevos: `tooling/management/voice.go` y `tooling/management/voice_test.go`.

**Execution:** delegada a `backend-developer`, porque la interfaz ya está fijada en el diseño.

**Test approach:** characterization para parametrizar las funciones de bloque; tdd para el catálogo y la generación del texto.

**Changes:**
- **Caracterización primero:** fijar con pruebas `blockRange`, `managedBlock`, `owned` y `transform` para el par de Hive, incluidos CRLF, separador, archivo creado por Hive y marcadores duplicados o invertidos. Luego parametrizar por el par de marcadores sin cambiar esas pruebas.
- **Catálogo, con pruebas escritas antes del código:**
  - `ListVoices(source)` y `RenderVoice(source, VoiceSetting) (body []byte, sourceHash string, err error)`, según [generación del texto](design.md#generación-del-texto).
  - Errores ante una voz desconocida, `name` sin nombre, un nombre inválido o un texto con marcadores.
  - La salida de `RenderVoice` empieza con el preámbulo y sigue con la voz, el tratamiento y la intensidad.
  - `validateRelease` rechaza también los marcadores de voz.

**Verification:** `go test -race ./tooling/management` pasa. Las pruebas de caracterización pasan igual antes y después de parametrizar, sin modificarlas entre medio. Las pruebas nuevas del catálogo fallan en la base.

### T2 — Plan, aplicación, recuperación y estado de la voz

- [x] El gestor arma, aplica y recupera cambios de voz, los conserva y regenera en `plan install`, los quita en `plan remove` y los informa en `status`.
  - Ronda 1, AC2 y AC3 no cumplidos: la voz se agrega al final del archivo y no justo tras el bloque de Hive; sin salto de línea final el archivo queda mal formado y `voice off` falla. También: `install` regenera la voz fuera de los CLIs y del alcance elegidos, un CLI nuevo no recibe la voz en el mismo plan, y un bloque de voz sin registrar no provoca conflicto.
  - **Ronda 2:** AC2 a AC8 cumplidos en la capa del gestor. `review-task` repitió 14 roturas deliberadas del código; 13 las detectaban las pruebas, y la que sobrevivía (reconstruir la voz al final del archivo al recuperar) la cubre ahora `TestVoiceOffRecoveryKeepsVoiceAfterHiveBlock`, tomada de su prueba P16.
  - **Correcciones de la ronda:**
    - la voz se inserta tras el bloque de Hive;
    - solo toca los CLIs y el alcance elegidos;
    - un CLI nuevo la recibe en el mismo plan;
    - un bloque sin registrar provoca conflicto;
    - los consumidores se reducen al quitar un CLI;
    - `status` ordena por ruta, tipo y CLI;
    - se rechazan rutas duplicadas;
    - la recuperación ya no da un conflicto falso con archivos nuevos.
  - **Límite menor pendiente (H7):** si Grok se suma a un archivo compartido que ya tiene voz, `status --hosts grok` no muestra la fila de voz hasta el siguiente cambio. Se corrige en T4.

**Closes:** AC2, AC3, AC4, AC5, AC6, AC7 y AC8, en la capa del gestor.

**Depends on:** T1.

**Locations:**
- `tooling/management/types.go`: `Plan.Voice`, `State.Voice` y `State.VoiceSpans`.
- `tooling/management/plan.go`: la nueva `BuildVoicePlan`, `BuildPlan` para install y remove, `PlanUnchanged`, `validatePlan` y `Status`.
- `tooling/management/apply.go`: `prepareTransaction`, la composición por archivo, la confirmación y `Recover`.
- `tooling/management/voice.go` y `voice_test.go`.

**Execution:** delegada, el mismo hijo que T1, en secuencia.

**Test approach:** tdd.

**Changes:** según [modelo](design.md#modelo), [operaciones](design.md#operaciones) y [aplicación y recuperación](design.md#aplicación-y-recuperación). Cada caso se prueba antes de implementarlo, sobre homes sintéticos:
- **Activar y apagar:** set y off en Codex y Claude dejan los bytes idénticos (AC2 y AC3).
- **Release con voz:** una release nueva con la voz sin cambios no toca el tramo; con el texto cambiado, lo regenera con la misma elección (AC4).
- **Desinstalar:** remove de un consumidor único y de un archivo compartido entre Grok y Claude (AC5).
- **`status`:** con `ok` y con `drift` (AC6).
- **Edición a mano:** conflicto en set, off, install y remove (AC7).
- **Recuperación:** interrupción en cada punto de escritura con el patrón de fallos inyectados de las pruebas actuales; después, `Recover` (AC8). Incluye:
  - un archivo que solo cambia la voz;
  - un archivo que cambia los dos bloques;
  - una edición del usuario fuera de los bloques hecha después de la interrupción, que debe conservarse.
- **Estado:** un plan `voice` conserva los recibos de versión de los CLIs, e `install` y `remove` conservan `Voice` y `VoiceSpans` cuando no tocan la voz.
- **Orden:** quitar Hive y la voz de un archivo creado por Hive lo borra.
- **Formatos:** un plan con voz se guarda y se carga con `SavePlan` y `LoadPlan`; un plan y un estado sin voz conservan su ID y sus bytes.

**Verification:** `go test -race ./tooling/management` pasa, y las pruebas nuevas fallan en la base.

### T3 — Textos de las tres voces

- [x] `content/voices/` tiene `preamble.md`, `jarvis.md`, `senior-direct.md` y `mentor.md`, en inglés.
  - **Evidencia:** `review-task` da cumplida la parte de contenido de AC10. Las pruebas fallaron antes de escribir los textos, con la carpeta ausente, y pasan después.
  - **Corregido después de la verificación, con las propuestas del verificador:**
    - La prueba ahora exige las frases que fijan cada regla (la lista de destinos, «only in transitions», «without metaphors», «without flattery or deference» y «keep plain wording»); antes seis eliminaciones del preámbulo pasaban sin fallar.
    - Las tres voces ya no dan estilo al contenido del desacuerdo.
    - Mentor ya no admite elogios.
    - La cortesía de Jarvis nunca va antes del resultado.

**Closes:** AC10.

**Depends on:** ninguna.

**Locations:** `content/voices/` y una prueba en `tests/content/`.

**Execution:** hilo principal. El tono es la parte subjetiva del cambio y el brief sería más largo que los textos.

**Test approach:** check.

**Changes:**
- **Preámbulo:** según [generación del texto](design.md#generación-del-texto).
- **Cada voz:** describe su estilo (registro, ritmo, recursos permitidos y lo que evita) sin copiar diálogos de películas ni pedir prosa telegráfica. Recuerda que las reglas de Hive mandan en la estructura y el vocabulario.
- **Prueba en `tests/content/`:** comprueba que el preámbulo contiene cada punto de AC10 (prioridad, la lista de lo que no cambia, dónde entra el tono y cuándo ignorarla) y que ningún texto de voz contiene marcadores.

**Verification:** `go test ./tests/content` pasa, y la prueba nueva falla en la base porque `content/voices/` no existe.

### T4 — Comando `hive voice` y la voz en `hive update`

- [x] `hive voice list|set|off` funciona de punta a punta, y el resumen de `update` muestra los cambios de voz.
  - Ronda 1, AC4 no cumplido: la prueba de la línea de voz del resumen buscaba «Voice» y lo encontraba en la ruta temporal, así que pasaba aunque se quitara la línea. Además, `voice set` guarda el tratamiento y la intensidad vacíos en lugar de `none` y `subtle`, y un plan de voz sin cambios de voz se aceptaba.
  - **Ronda 2:** AC1 a AC4 y AC9 cumplidos. `review-task` repitió las roturas deliberadas del código y cada una hace fallar una prueba; también confirmó que los cambios de voz que no cambian nada no escriben archivos ni dan cambios falsos, y que la recuperación funciona.
  - **Correcciones de la ronda:**
    - la prueba exige la línea exacta del resumen y su ausencia cuando no corresponde;
    - `BuildVoicePlan` guarda `none` y `subtle` por defecto;
    - un plan de voz sin cambios de voz se rechaza;
    - hay pruebas de rechazo en la confirmación, del caso sin cambios y de que no se escriba nada sin Hive.
  - **Después de la verificación:** el resumen de `voice set` contaba y listaba también los archivos que no cambian (N1). Lo corregí en el hilo principal; la prueba `TestVoiceSummaryCountsOnlyFilesThatChange` falló antes del arreglo y pasa después.
  - **Incluido:** la corrección H7 del gestor (los consumidores del tramo se amplían cuando un CLI se suma a un archivo compartido).

**Closes:** AC1, AC2, AC3, AC4 (la línea de voz en el resumen de `update`) y AC9 (los errores del comando).

**Depends on:** T2, T3.

**Locations:**
- Nuevos: `tooling/cli/voice.go` y `tooling/cli/voice_test.go`.
- El despacho y la ayuda en `tooling/cli/main.go`, manteniendo `run` en 40 líneas o menos.
- `tooling/cli/update.go`: la línea de voz del resumen.

**Execution:** delegada a `backend-developer`.

**Test approach:** tdd.

**Changes:** según [operaciones](design.md#operaciones), con el mismo patrón de confirmación, `--dry-run` y `--out` de `update`. Pruebas sobre un home de prueba con la fuente del repositorio:
- `list` muestra las tres voces;
- `set` en modo interactivo, en `--dry-run` y con `--out` seguido de `apply`;
- `off`;
- los errores de AC9;
- el resumen de `update` con una voz cuyo texto cambió.

**Verification:** `go test -race ./tooling/cli -run 'Voice|Update'` pasa, `go test ./tooling/cli` pasa, y `awk '/^func run\(/,/^}/' tooling/cli/main.go | wc -l` da 40 o menos.

### T5 — Validación humana de las voces

- [x] El usuario aprueba el texto de las tres voces y un mensaje de ejemplo en cada una, o pide cambios. Aprobado por el usuario el 2026-09-27: «Apruebo las tres». Vio el resumen de cada voz y el mismo reporte de avance en tono neutro, Jarvis con «señor», Senior directo y Mentor, todos con intensidad sutil.

**Depends on:** T3.

**Execution:** hilo principal.

**Changes:** mostrar al usuario, para cada voz:
- el bloque generado con cada intensidad;
- el mismo mensaje de ejemplo (un reporte de avance corto con secciones etiquetadas) escrito en esa voz.

Aplicar los cambios que pida.

**Verification:** la aprobación explícita del usuario queda anotada aquí con su fecha.

**Límite:** los ejemplos los escribe el hilo principal para mostrar el tono; no miden si un modelo de ejecución con la voz cargada respeta las reglas, porque los pilotos están pausados. El seguimiento es revisar la primera sesión real con la voz activa: las secciones con etiqueta, la glosa de los IDs y el tono de los reportes de los subagentes.

### T6 — Documentación y prueba local

- [?] `deployment-manager.md` documenta `hive voice`, y una prueba local sobre un home de prueba confirma el ciclo completo.
  - **Evidencia del hilo principal (2026-09-28):** el ciclo completo se ejecutó con un binario compilado del worktree, sobre `--home` y `--state-dir` temporales:
    - `install` de Codex y Claude;
    - `voice list` mostró las tres voces;
    - `voice set jarvis --address sir`, primero con `--dry-run` y luego con `--out` y `apply`, puso un solo bloque justo después del de Hive, con sus líneas de tratamiento e intensidad;
    - `status` mostró filas `voice installed jarvis (sir, subtle)`;
    - `update --dry-run` sobre un clon con `jarvis.md` cambiado mostró «Voice files to regenerate: 2»;
    - una edición a mano apareció como `drift` y bloqueó `voice off`;
    - al restaurarla, `voice off` dejó los dos archivos idénticos byte a byte a su estado tras `install` (`cmp`) y sin `voice` ni `voice_spans` en el estado.

    El estado real del usuario no se tocó.
  - **Arreglo encontrado en la prueba:** el mensaje de conflicto de voz no nombraba el archivo. Ahora dice `voice block conflict in <ruta>`, con la prueba `TestVoiceConflictNamesTheFile`; su rojo fue de compilación, por el cambio de firma.

**Closes:** AC11.

**Depends on:** T4, T5.

**Locations:** `_support/docs/architecture/deployment-manager.md`.

**Execution:** hilo principal.

**Test approach:** check.

**Changes:**
- Documentar `hive voice`, el bloque y sus marcadores, la relación con `update` y `plan remove`, y la incompatibilidad con el gestor anterior.
- Ejecutar el ciclo con `go run ./tooling/cli` sobre `--home` y `--state-dir` temporales:
  - `install` de Codex y Claude;
  - `voice set jarvis --address sir`;
  - `status`;
  - `update --dry-run` con el texto de la voz cambiado localmente;
  - `voice off`;
  - comparar los bytes de los archivos con los del principio.

**Verification:**
- `rg -n "hive voice" _support/docs/architecture/deployment-manager.md` devuelve la sección nueva.
- Los archivos del home de prueba quedan idénticos a su estado tras `install`.
- El estado real del usuario no se toca.

## Verificación y revisión humana

| Comprobación | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Pruebas | Al cerrar cada tarea y al final, en local | `go vet ./...` y `go test -race ./...` en verde; no hay CI | Parte de la implementación |
| Verificación por tarea | Tras T1, T2, T3, T4 y T6 | `review-task` contra sus `AC<n>` | Parte de la implementación |
| Validación humana del tono | T5, antes de la revisión de código | El usuario lee los bloques y los ejemplos | Parte del plan; decide el usuario |
| Revisión de código | Antes del push | `/code-review` de Claude Code (D8-A) | Ya decidida |
| Activar la voz en los CLIs reales | Después de integrar, si el usuario lo pide | `hive voice set` en una terminal; escribe en la configuración global | Necesita autorización explícita |

## Estado de la revisión y avance

- **Revisión del plan** (2026-09-27): versión revisada `1408cbf268b1`. Dos revisores `review-plan` en paralelo, como rol nativo de Claude Code y de solo lectura.
  - **Modelo del gestor:** seis hallazgos que bloqueaban, todos corregidos con la propuesta del revisor.
    - La transacción perdía la voz, borraba recibos de versión y no tenía entradas del journal para la voz.
    - `Recover` solo sabía invertir el bloque de Hive.
    - El orden de los dos bloques al instalar y al quitar se contradecía.
    - El diseño afirmaba en falso que un gestor anterior fallaba ante un estado con voz; se decidió D13-A: mantener el esquema 6 y documentarlo.
    - Faltaba de dónde sale el texto de la voz al aplicar, y el `SourceHash` pasa a guardarse por tramo.
    - AC4 no estaba en el `Closes:` de T4 y AC6 no coincidía con el diseño.

    Se incorporaron las sugerencias sobre `status`, el cálculo de `changed`, las rutas tomadas de `Records` y la lectura con `overlayRead`.
  - **Contenido y reglas:** tres hallazgos que bloqueaban, corregidos.
    - Grok da prioridad al texto posterior, así que el preámbulo declara la prioridad por sí mismo.
    - La cláusula general se reemplazó por una lista explícita de lo que la voz no cambia, y AC10 se amplió.
    - La condición para ignorar la voz se escribe para el agente que la lee, y su límite queda registrado.

    Se incorporaron el tratamiento sin marcar género y el límite de lo que mide T5. La opción de «señora» o de un tratamiento neutro queda como sugerencia para el usuario.
  - **Sin nueva ronda:** todas las correcciones aplican propuestas de los revisores, salvo D13, que decidió el usuario.
- **Avance:** plan en borrador; ninguna tarea empezada.
- **Siguiente paso:** `flow-build`, con T1 (gestor) y T3 (textos) en paralelo.
