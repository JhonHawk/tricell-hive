# Diseño

## Contexto verificado

Inspeccionado el 2026-10-01 sobre `d4c014d`. Dos subagentes `hive-research` leyeron el código. Los datos que deciden el diseño se volvieron a leer en el hilo principal. Las rutas son relativas a la raíz del repositorio.

**Modelos** (lectura directa salvo donde dice inferencia):
- **Release:** el ID de la release resume los archivos, los bytes crudos de `agent-profiles.json` y la versión del renderizador (`tooling/management/types.go:227-236`). El snapshot `releases/<id>.json` guarda fuentes y perfiles, nunca archivos renderizados (`types.go:89-94`).
- **Esfuerzo:** `agents.Resolve` elige el perfil del rol, y el `effort` propio del rol reemplaza al del perfil (`integrations/agents/agents.go:278-282`). En OpenCode el esfuerzo viaja como sufijo `#variant` (`:284-287`).
  - `Render` llama a `Resolve` y recibe solo los bytes de perfiles.
  - Las dos se llaman desde `plan.go:517` (`nextRecord`) y desde `tooling/management/models.go:112` (`EffectiveModels`).
- **Plan:** `nextRecord` renderiza cada agente para sus consumidores y guarda el resultado en `Record.Managed`. `validatePlan` vuelve a calcular `nextRecord` y exige igualdad exacta (`plan.go:592-700`). Por eso cualquier ajuste tiene que viajar congelado dentro del plan.
- **Drift:** compara el archivo con `Record.Managed` (`files.go:277-284`), así que un archivo ajustado no se marca como drift.
- **Recibos:** el recibo de un CLI resume `hash(Record.Managed)` de sus recursos (`versioning.go:88-89`). `installationVersion` da `partial` cuando uno no coincide (`:144`). `updateProductState` borra el recibo cuando el plan no lleva `Product` (`:107`).
  - Un plan `install --release <id>` no lleva `Product`, porque `productFromSource` devuelve `nil` con `ReleaseID` (`versioning.go:46-48`). La vuelta atrás de hoy ya deja el CLI sin versión de producto.
- **Voz:** es el precedente.
  - `State.Voice` y `Plan.VoiceSetting` son campos `omitempty` sin cambio de esquema (`types.go:116-122`, `:160-168`).
  - `prepareTransaction` copia el estado campo por campo (`apply.go:205-217`): un campo nuevo que no se copie se pierde en el siguiente `apply`.
  - Al quitar el último consumidor, la elección se borra (`apply.go:280-289`).
- **Renderizador:** `nextRecord` falla con «agent renderer version changed; regenerate plan» cuando la release se renderizó con otra versión (`plan.go:514-515`). Reinstalar una release vieja hereda ese límite.
- **Validación:** no hay catálogo de modelos en el código. Un modelo solo se valida contra saltos de línea y NUL (`agents.go:158`). Los esfuerzos aceptados son `low|medium|high|xhigh|max` en los roles y además `ultra` en los perfiles (`agents.go:95-96`, `:161`). Grok y Cursor no aceptan esfuerzo (`:237-239`).

**`## Hive`** (lectura del subagente, contrastada en `doctor_project.go`):
- **Validador:** `parseHiveSection` (`tooling/cli/doctor_project.go:233-265`) hace lo siguiente:
  - busca la línea exacta `## Hive` e ignora los bloques de código;
  - termina la sección en la siguiente línea que empieza con `#`;
  - lee los elementos `- Key: value`, cortando en el primer `:`;
  - cuenta las secciones duplicadas.
- **Reglas que aplica `checkProject` (`:147-322`):**
  - resuelve la raíz con `git rev-parse --show-toplevel`;
  - lee hasta 1 MiB;
  - comprueba la rama con `checkBaseBranch` (`:381-402`);
  - exige `Delivery` = `direct-base` y `Hive guidance` = `required`.
- **Escrituras:** ningún código escribe un archivo del usuario sin marcadores de Hive. `management.write` no se exporta y está atado al journal. `target.Safe` rechaza enlaces simbólicos en la ruta (`integrations/target/target.go:88-110`).
- **Patrones de la TUI:** campo de texto `textinput` con `TextFocused()` en la vista Voice, y `confirmView` con Apply y Cancel (`tooling/cli/tui_views.go:472-536`). Las pruebas usan `appDriver` en proceso (`tooling/cli/tui_test.go`).
- **Guía:** sobre los valores (`content/guidance/global.md`, «Project settings»), dice que un valor visto solo en otra documentación es una recomendación, no el ajuste. Por eso las sugerencias van marcadas y el usuario las confirma.

## Diseño elegido

### Ajuste de modelo en el estado (D1-A, D2-A)

**Forma.** En `tooling/management/types.go`:

```go
type ModelOverride struct {
	Model  string `json:"model,omitempty"`  // "" keeps the release's model
	Effort string `json:"effort,omitempty"` // "" keeps the release's effort
}
// State and Plan gain the same field, keyed host → role:
ModelOverrides map[string]map[string]ModelOverride `json:"model_overrides,omitempty"`
```

Sin ajustes, el campo no aparece y `state.json` se codifica igual que antes.

**Regla de resolución.** `agents.Resolve` y `agents.Render` reciben un `*ModelOverride` del paquete `agents`. Con `nil` el resultado es el de hoy. Con un ajuste, se aplica después del esfuerzo propio del rol y antes del sufijo de OpenCode:
- un `Model` no vacío reemplaza al del perfil;
- un `Effort` no vacío reemplaza al efectivo.

Así el ajuste siempre gana, y OpenCode recibe `model#effort` igual que hoy. En OpenCode, cuando el ajuste trae `Effort`, primero se quita cualquier `#variant` del modelo efectivo. El perfil `inherit` trae `…#max`, y sin ese paso saldría `…#max#high`. Es una sola regla para el render, el plan y `EffectiveModels`.

**Validación** (`agents.ValidateOverride(host string, o ModelOverride) error`, usada al construir el plan y otra vez en `validatePlan`):
- **Ajuste vacío:** un ajuste sin `Model` ni `Effort` se rechaza; `prepareTransaction` nunca lo guarda.
- **Modelo, cuando no está vacío:** de 1 a 200 caracteres de `[A-Za-z0-9._:/@+=\[\]-]`. Un `Model` vacío significa «el de la release» y no se valida.
  - Admite `provider/model` y `name[effort=high]`.
  - Excluye `#`, espacios, comillas, barras invertidas, caracteres de control e invisibles.
  - Motivo: Pi lee la cabecera con un parser propio que no decodifica escapes (`agents.go:374`), y en Cursor no se ha observado que los acepte. La lista cerrada evita que el estado guarde un valor y el archivo muestre otro.
- **Esfuerzo:**
  - se acepta la lista de los perfiles (`low|medium|high|xhigh|max|ultra`, `agents.go:161`) en los CLIs que aceptan esfuerzo (`:237-239`), sin reglas por modelo, que el código no tiene;
  - vacío en Grok y Cursor.
- **OpenCode:** un esfuerzo exige un modelo efectivo no vacío (el del ajuste o el del perfil), porque sin modelo no hay dónde poner `#variant`. `ValidateOverride` no recibe el perfil, así que esta regla la aplica `Resolve`, que devuelve el error y por eso falla al construir el plan y al validarlo.
- **Rol:** debe existir en la release instalada en ese CLI. Lo comprueba el constructor del plan, no `agents`.
- **`validatePlan`:** comprueba además lo siguiente:
  - **Plan `install`:**
    - cada CLI de `Plan.ModelOverrides` pasa `validateHosts`, el conjunto de CLIs que ya usa `management`;
    - los ajustes de los CLIs que no están en el plan son iguales a los del estado;
    - cada ajuste de un CLI del plan es igual al del estado o nombra un rol de la release del plan.
  - **Planes `remove` y `voice`:** `Plan.ModelOverrides` debe ir vacío. `prepareTransaction` copia los ajustes del estado y, en un `remove`, borra los de los CLIs quitados.

**Constructor del plan.** La función es `management.BuildModelsPlan(o Options, host string, next map[string]ModelOverride) (Plan, error)`.
- `next` es el conjunto completo de ajustes de ese CLI después del cambio; vacío significa quitar todos.
- Construye un plan `install` para un solo CLI desde la release que tiene instalada, como `install --release <id>`.
  - Si sus agentes vienen de releases distintas, o la release se renderizó con otra versión del renderizador, falla con «Run hive update first; <host> has agents from more than one release» o con el error del renderizador. No hay fallback.
  - Rechaza el plan cuando hace algo más que reescribir archivos de agente:
    - un cambio con `Gone`, es decir, un recurso borrado a mano (`plan.go:401`);
    - un cambio con `Before != After` en un recurso que no es agente;
    - `Legacy` o `Migration` no vacíos (`scanMigration`, `plan.go:385`).

    El mensaje remite a `hive install` o `hive doctor`. Así `models set` nunca restaura ni migra nada (AC1).
- Pone `Plan.ModelOverrides` con los ajustes del estado para los demás CLIs y `next` para este.
- **Recibo:** conserva el `Product` del recibo actual del CLI cuando su `ReleaseID` coincide con la release reinstalada. Así `updateProductState` lo regenera con los hashes nuevos y `hive status` sigue en `verified` (AC2).
  - Sin recibo, o con otro `ReleaseID`, el plan sale sin `Product`, como la vuelta atrás de hoy. El CLI queda en `legacy`, igual que antes del ajuste.
  - La vuelta atrás (`install --release`) no cambia: sigue sin `Product`. Que la vuelta atrás conserve la versión de producto es otro problema, que existe hoy y queda fuera de este cambio.
  - `validateProduct` ya acepta un `Product` igual al que está en `State.Versions`.
- **Otros planes `install`:** `BuildPlan("install", …)`, que usan `update`, `install` y la vuelta atrás, copia `State.ModelOverrides` en el plan. Así cada reinstalación vuelve a aplicar los ajustes (AC2).
- **Planes `remove`:** no llevan ajustes.

**Plan y aplicación.**
- `nextRecord` pasa a `agents.Render` el ajuste de `p.ModelOverrides[c.Host][rol]`. El nombre del rol sale de la fuente del agente, el mismo que usa `EffectiveModels`.
- `validatePlan` valida cada ajuste del plan con `ValidateOverride`. Su recálculo de `nextRecord` ya cubre que el archivo coincida.
- `prepareTransaction` (`apply.go`) hace lo siguiente:
  - copia `ModelOverrides` del estado;
  - en un plan `install`, reemplaza los ajustes de los CLIs del plan por los del plan;
  - en un plan `remove`, borra los ajustes de los CLIs quitados;
  - poda los mapas internos vacíos y deja `nil` cuando no queda ninguno, para que `state.json` y `planID` no cambien sin ajustes;
  - marca `changed` cuando `ModelOverrides` cambia. Hoy solo lo marcan archivos, voz y producto (`apply.go:221-298`). Sin esto no se guardaría un cambio que ningún archivo refleja: un `reset` de un ajuste no aplicado, o un `set` igual al valor de la release.
- **Plan viejo:** `StateHash` ya rechaza un plan construido sobre otro estado. Se comprueba en `preflight` y otra vez bajo el bloqueo (`apply.go:85`, `:119`), y es el hash de los bytes de `state.json` (`files.go:195`). Una prueba lo fija para un ajuste cambiado entre `--out` y `apply`.
- **Recuperación:** `Recover` restaura `state.json` desde `BeforeState` (`apply.go:747`), así que no necesita nada nuevo.
- **Roles que desaparecen:** un ajuste de un rol que ya no está en la release no se aplica y se conserva en el estado, para que vuelva si el rol regresa. `hive models` lo lista como «not applied: role not in the installed release».

**Modelo efectivo.** `EffectiveModels` pasa el ajuste del estado a `Resolve`. `ModelRow` gana `Override bool`.

### Comandos de modelos

Todo lo de esta sección va en `tooling/cli/models.go`, que ya enruta `hive models`.

```
hive models [--home DIR] [--state-dir DIR]                 # unchanged
hive models set --host H --role R [--model M] [--effort E] [--home DIR] [--state-dir DIR] [--dry-run | --out FILE]
hive models reset --host H (--role R [--only model|effort] | --all) [--home DIR] [--state-dir DIR] [--dry-run | --out FILE]
```

- **`reset --only`:** quita solo una parte del ajuste del rol y conserva la otra. Si no queda ninguna parte, equivale a `reset --role R`. Es el equivalente en texto del borrado parcial del panel.

- **`set`:** combina el ajuste nuevo con el del estado.
  - Un campo que no se pasa queda como estaba en el ajuste guardado, o vacío si no había ajuste. Nunca se copia el valor de la release dentro del ajuste.
  - Así un ajuste solo de esfuerzo sigue los cambios de modelo de las releases futuras.
  - Así `set --effort` sobre el perfil `inherit` de OpenCode (`…#max`) no choca con la regla del `#`.
  - Pedir sin `--model` ni `--effort` es un error de uso.
- **Confirmación:** sigue el patrón de `voice set` (`tooling/cli/voice.go`): resumen, confirmación con terminal y `--dry-run`/`--out` sin ella. El resumen dice:
  - `<host> <role>: model A → B, effort X → Y`;
  - la ruta del archivo;
  - «Open sessions keep the previous model until they restart.»
- **Plan sin cambios:** un `set` o `reset` que no cambia ni archivos ni ajustes imprime «Nothing to change» y no escribe.
- **Modelo vacío:** `--model ""` se rechaza. Para volver al modelo de la release se usa `reset`.

### Edición en la vista Models

Patrón: la vista Voice (fila con ←→ y campo `textinput`) y `confirmView`. No hace falta un patrón nuevo.

La barra de ayuda está fuera del área de la vista; sale de `Keys()` (`tui_app.go:455-458`):

```
Models
[claude]  codex   grok   pi   opencode   cursor
  Role                 Profile    Model          Effort
> hive-review-code *   reasoning  opus           max
  hive-build-backend   execution  sonnet         high
  …
Enter edits the selected role; x resets one marked *.
Defaults come from integrations/agent-profiles.json in the release.
```

- **Filas fijas:** siguen siendo 6: título, CLIs, encabezado, posición y las dos líneas del pie de arriba, que sustituyen al pie de hoy (`tui_models_view.go:163-166`).
- **Lista:**
  - ↑↓ mueve el cursor `>`, que sustituye al desplazamiento con ↑↓. PgUp y PgDn siguen desplazando, y el `scrollBox` sigue al cursor.
  - ←→ cambia de CLI como hoy.
  - Después de aplicar o recargar, el cursor queda en el mismo rol.
- **Panel de edición** (Enter): está en la misma vista, no es una vista nueva. Ocupa 5 filas sobre la tabla: rol, `Model`, `Effort`, una fila de error y una en blanco. A 80×24 la tabla queda con 10 filas.
  - **`Model`:** campo de texto con el modelo efectivo sin el `#variant`.
  - **`Effort`:** se cambia con ←→ entre «release default» y la lista válida. En Grok y Cursor dice «not supported by <host>» y no recibe foco.
- **Qué envía Enter:** construye un `set` con lo que cambió respecto del valor efectivo.
  - El modelo solo va si el texto difiere del modelo efectivo. Así, cambiar solo el esfuerzo deja el mismo estado que `hive models set --effort` sin `--model`.
  - Un campo `Model` vacío significa «modelo de la release»: quita la parte del modelo del ajuste.
  - «release default» en el esfuerzo quita la parte del esfuerzo.
  - Si no queda nada del ajuste, el resultado es un `reset` del rol.
  - Cada resultado tiene su comando equivalente: un `set` con las partes que cambian, `reset --role R --only model|effort` o `reset --role R`. Las pruebas comparan el estado del panel con el del comando.
- **Teclas en el panel:**
  - Esc cierra solo el panel y vuelve a la lista: la vista devuelve `navNone`.
  - Backspace en el campo `Effort`, que no es de texto, también devuelve `navNone`. Si devolviera `action{}` como en Voice, la raíz cerraría la vista entera (`tui_app.go:416`).
  - Mientras el campo de texto tiene el foco, `TextFocused()` es verdadero, y Backspace, `r`, `x` y `e` escriben en el campo.
  - ↑↓ cambian de campo.
- **`x`:** en la lista, con un rol ajustado, abre la confirmación de `reset` de ese rol. Sin ajuste, no hace nada.
- **Confirmación y resultado:** se sigue el flujo de la vista Voice (`tui_voice_view.go:355-416`):
  - `confirmView` con el resumen del comando;
  - Cancel o Esc vuelven con «Cancelled. No changes applied.»;
  - un plan sin cambios no abre la confirmación y dice «Nothing to change»;
  - un error al aplicar con una operación pendiente abre la recuperación como en Voice;
  - si sale bien, la vista recarga y cierra el panel.
- **Errores:** un error de validación o del constructor ocupa la fila de error del panel, en una sola línea recortada con «…», como en Voice (`tui_voice_view.go:467-470`).
- **Barra de ayuda (`Keys()`), cabe en 80 columnas con `ctrl+c quit`:**
  - lista: `↑/↓ role · enter edit · x reset · ←/→ CLI · r reload · esc back`, que reemplaza a `scrollBinding` (PgUp y PgDn siguen funcionando aunque no se listen);
  - panel: `↑/↓ field · ←/→ effort · enter review · esc close`.

### Lista de modelos del CLI (D6-A, T7)

**Contexto verificado el 2026-10-01** (subagente `hive-research`, comandos ejecutados con límite de 15 s desde carpetas temporales; documentación vía Context7 y páginas oficiales):
- **Codex 0.159.1:** `codex debug models` está documentado ([developer-commands](https://learn.chatgpt.com/docs/developer-commands?surface=cli)). Tarda unos 1 s y devuelve unos 640 KB de JSON con 11 modelos. Cada uno trae `slug`, `visibility` (`list` o `hide`) y `supported_reasoning_levels`.
- **OpenCode 2.0.21:** `opencode models` imprime `provider/model`, 209 líneas. En la primera tanda devolvió tres veces salida vacía con código 0.
- **Pi 0.99.2:** `pi --list-models` imprime una tabla con columnas `provider model context max-out thinking images`.
- **Grok 1.0.45:** `grok models` imprime una línea de sesión y después los modelos, en líneas que empiezan con `-` o `*`.
- **Cursor 2026.09.28:** `cursor-agent models` está documentado ([parameters](https://cursor.com/docs/cli/reference/parameters)). Imprime `id - Name`, más un encabezado y un consejo.
- **Claude Code 2.1.286:** no tiene comando de listado. Los alias documentados son `fable`, `opus`, `sonnet` y `haiku`, más los ids completos ([model-config](https://code.claude.com/docs/en/model-config)).
- **optional reference project `b388eb31`:** consulta en vivo a OpenCode y a Codex, con una lista fija si Codex falla (`internal/model/codex_model.go:95-132`), y usa alias fijos para Claude (`claude_model.go:12-29`).

**Contrato** (`tooling/cli/model_catalog.go`, nuevo):

```go
type catalogModel struct {
	ID      string
	Efforts []string // empty: the host's general list applies
}
// listHostModels runs the host's listing command and parses its output.
func listHostModels(ctx context.Context, host string, run catalogRunner) ([]catalogModel, error)
```

- **`catalogRunner`:** recibe el nombre del binario y los argumentos fijos, y devuelve la salida estándar. La implementación real es un ejecutor propio (`newCatalogRunner(timeout, maxOutput)`), no una extensión del de `--version` (`doctor.go:83-115`): aquel mezcla la salida de error y recorta en vez de fallar, y extenderlo cambiaría `hive doctor`. Cumple estas condiciones:
  - `exec.LookPath` sobre el binario (`cursor-agent` para Cursor);
  - sin shell, con el entorno del proceso, porque el CLI lo necesita para autenticarse;
  - **carpeta de trabajo neutra (`cmd.Dir = os.TempDir()`):** OpenCode y Pi cargan configuración y extensiones del proyecto desde la carpeta actual, y la interfaz suele abrirse dentro de un repositorio;
  - **aislado de la terminal:** entrada estándar nula, `SysProcAttr{Setsid: true}`, y `cmd.Cancel` mata a todo el grupo de procesos;
  - 8 s de límite y `WaitDelay` de 1 s;
  - salida estándar acotada a 8 MiB, y error si se supera;
  - la salida de error se descarta.

  Las pruebas inyectan un runner falso con salidas grabadas. Los efectos que no se controlan quedan declarados: cachés y telemetría propios de cada CLI, y una posible consulta de red.
- **Argumentos fijos por CLI:**
  - `codex debug models`;
  - `opencode models`, con un reintento si la salida sale vacía y el código es 0;
  - `pi --list-models`;
  - `grok models`;
  - `cursor-agent models`.
- **Lectura de la salida:**
  - **Codex:** JSON, `models[].slug` con `visibility == "list"`. `Efforts` es `supported_reasoning_levels` filtrado por `ValidateOverride`. La implementación confirma la forma real del campo con la salida grabada.
  - **OpenCode:** una línea por modelo.
  - **Pi:** a partir de la tabla, se arma `provider/model` con las dos primeras columnas; la cabecera se omite.
  - **Grok:** se toman las líneas que empiezan con `-` o `*`, quitando la marca y el sufijo «(default)».
  - **Cursor:** el texto antes de ` - `, y solo en las líneas cuyo primer campo cumple la lista de caracteres.
- **Ids válidos:** se ofrecen solo los que pasan `ValidateOverride`. El resto se descarta sin aviso, y la cuenta de descartados queda en el error si no queda ninguno.
- **Claude:** no ejecuta nada y devuelve los alias fijos.
- **Con `--home`:** no se ejecuta nada y la vista muestra solo el valor actual y «Other…», igual que `detectInstallerHosts` (`install.go:603-604`).
- **Caché:**
  - El resultado se guarda en un puntero compartido que crea `appModel` una vez por sesión (`*modelCatalogCache`, por CLI), porque `appConfig` se copia por valor en cada vista (`tui_views.go:272`).
  - Solo se escribe en `Update`, con un mensaje que no pertenece a la vista (no `owned{v}`). Así una consulta que termina después de salir de Models no se pierde.
  - La primera apertura del panel de ese CLI lanza la consulta en segundo plano. Mientras carga, `NeedsSpinner()` es verdadero.
  - Un fallo no se guarda: la siguiente apertura del panel vuelve a intentar.
  - `hive models`, `hive doctor` y la vista Models mientras no se abre un panel nunca la ejecutan.
- **No se leen cachés internos de los CLIs** (`~/.codex/models_cache.json`, `~/.cache/opencode/models.json`): no están documentados.

**Selector en el panel** (forma final aceptada por el usuario en las rondas de corrección 2–4, 2026-10-01):
- **El campo Model** muestra su valor con `▸` y, con el foco, la pista atenuada «→ choose». Enter sigue revisando el cambio desde cualquier fila del panel.
- **→ o una letra** sobre Model abren el recuadro:
  - borde cuadrado, sin fondo propio y dos columnas más angosto que la pantalla;
  - en el borde superior, «Select model · <CLI> · <rol o grupo>» a la izquierda y «esc» a la derecha;
  - la letra queda como primer carácter de la búsqueda.
- **Contenido del recuadro:**
  - la línea `Search` filtra por nombre, id y proveedor, sin distinguir mayúsculas;
  - una línea de estado solo cuando hace falta: cargando, o «Model list unavailable: <motivo>»;
  - la lista:
    - Claude, Codex, Grok y Cursor muestran una lista plana;
    - OpenCode y Pi muestran una sección por proveedor, con nombre legible para los conocidos y el prefijo crudo si no;
    - una fila muestra el nombre legible si el CLI lo da (Codex `display_name`, Cursor), con el id atenuado; bajo un proveedor, el id va sin su prefijo;
    - el valor guardado es siempre el id completo;
    - el modelo actual aparece aunque el CLI no lo liste, marcado con `●`, y el cursor arranca en él;
  - la línea final tiene «release default» y «Other…», a los que se llega con ↓ y entre los que se cambia con ←→.
  - No hay «In use», «Recent», favoritos ni precios.
- **Teclas:**
  - Enter elige y vuelve al panel. Esc cierra el recuadro, luego el panel y luego la vista.
  - «Other…» convierte Model en el campo de texto, con la misma validación que el comando.
- **Recortes:** los textos se recortan por ancho visual, no por número de caracteres.
- **Esfuerzo:**
  - se compara por valor;
  - en Codex, elegir un modelo recorta los esfuerzos a los suyos y, si el elegido ya no está, vuelve a «release default»;
  - abrir el panel o recibir la lista no mueve el esfuerzo.
- **Confirmación** (rol, grupo, y el comando de texto por igual):
  - un título, seguido de una tabla Role/Model/Effort con `a → b` solo en las partes que cambian;
  - «Replaces the own override of <roles>.» con los roles que pierden una parte;
  - «Writes N files in <carpeta>», con el home abreviado como `~`;
  - la nota de sesiones abiertas.
- **El comando de texto no cambia:** `hive models set --model` acepta cualquier id válido.

### Grupos por función (D7-A, T8 y T9)

- **Grupo:** el segundo nivel de la ruta de la fuente del rol (`content/agents/<grupo>/<rol>.md`), que `agents.IsSource` ya exige (`integrations/agents/agents.go:47-50`). Hoy son `design` (2), `development` (4), `docs` (1), `ops` (1), `quality` (5) y `review` (7). `ModelRow` gana `Group`.
- **Vista:**
  - Por cada CLI, los roles se ordenan por grupo y luego por rol. Antes de cada grupo va una fila de cabecera con el nombre del grupo con mayúscula inicial y su modelo y esfuerzo comunes, o `mixed`.
  - La cabecera no lleva perfil.
  - El cursor recorre cabeceras y roles.
  - El pie cambia a: «Enter edits the selected role or group; x resets one marked *.» y «Defaults come from integrations/agent-profiles.json in the release.»
- **Edición de un grupo:**
  - Enter sobre la cabecera abre el mismo panel, con Role «<Group> group on <host> (N roles)».
  - **Valores iniciales:** el modelo y el esfuerzo comunes. Si son «mixed», el campo muestra «mixed» y significa «no tocar esa parte en el envío».
  - **Al aplicar:** cada rol del grupo queda con un ajuste que tiene solo las partes elegidas, y el ajuste propio que tuviera se reemplaza (D7-A). Elegir «release default» quita esa parte en todos los roles. Una parte que quedó «mixed» no se envía, así que esa parte del ajuste propio se pierde.
  - **Confirmación:** la del resumen de arriba. Elegir «release default» en las dos partes de un grupo se presenta como «Reset the <Group> group…».
- **Regla de OpenCode, igual para un rol, un grupo, la pantalla y el comando** (corregida tras el `/code-review` del PR #89):
  - cuando un ajuste trae modelo y no esfuerzo, `agents.Resolve` conserva el esfuerzo del perfil;
  - si ese esfuerzo viene incrustado en el modelo del perfil (el `#max` de `inherit`), `Resolve` lo traslada al reemplazar el modelo;
  - el ajuste guarda solo el modelo, así que el rol sigue los cambios de la release.

  Antes se copiaba el esfuerzo mostrado dentro del ajuste, lo que fijaba el valor de la release.
- **Cabeceras:** llevan `*` cuando algún rol del grupo tiene ajuste. El cursor las identifica con una clave propia, distinta de `curRole`.
- **`x` sobre una cabecera:** con algún rol ajustado, abre la confirmación de quitar los ajustes de todo el grupo.
- **Comandos:** `hive models set --host H --group G [--model M] [--effort E]` y `hive models reset --host H --group G [--only model|effort]`.
  - `--group` y `--role` son excluyentes.
  - Un grupo que no existe en la release instalada de ese CLI se rechaza con «group "G" is unknown for <host>; known groups: …».
  - Por debajo, se construye el conjunto completo de ajustes del CLI y se llama una vez a `BuildModelsPlan`: no hay estado nuevo.
- **`hive models`:** imprime los roles agrupados con una línea de cabecera por grupo (`  review  opus  high` o `mixed`). El golden de texto de la base se regenera a propósito. Una prueba quita las cabeceras y compara las filas con el golden de la base, así que las filas de los roles no cambian.
- **Tabla compartida:** la tabla de la vista no se agrupa en T8. T8 agrupa solo la salida de texto, sin tocar `modelTable`, que comparte la vista (`models.go:112`); la vista se agrupa en T9.

**Seguridad (T7):** ejecutar los CLIs es un límite nuevo. Hoy la interfaz solo ejecuta `--version`. Los controles son:
- binario detectado por nombre fijo y argumentos fijos, sin shell;
- límite de tiempo y de tamaño de salida, y la salida de error descartada;
- la salida se trata como dato no confiable: se filtra con la lista de caracteres de los modelos y nunca se muestra cruda;
- sin `--home`, un CLI recibe el entorno del usuario, porque lo necesita para su autenticación. No se pasan credenciales por argumentos.

Grok imprime una línea de sesión que se descarta.

### Escritura de `## Hive` (D3-A)

**Edición pura** (`tooling/cli/project_write.go`):

```go
// editHiveSection returns data with the ## Hive section's items set or unset.
func editHiveSection(data []byte, set []hiveItem, unset []string) ([]byte, error)
```

- **Lectura:** usa el mismo recorrido de `parseHiveSection`. Se extrae una función que devuelve también las posiciones de las líneas, para que el validador y el editor coincidan.
- **Sección existente:**
  - reemplaza la primera línea `- Key:` de cada clave dada;
  - quita las de `unset`;
  - agrega las claves nuevas después del último elemento, en el orden de `hiveSettingKeys`;
  - deja sin cambios todo lo demás.
- **Sin sección:** agrega `\n## Hive\n\n- …` al final. Si el archivo no termina en salto de línea, primero agrega uno.
- **Finales de línea:** si la primera línea del archivo termina en `\r\n`, todas las líneas nuevas usan `\r\n`.
- **Rechazos:** una sección duplicada, y una clave repetida dentro de la sección cuando es la que se edita.

**Validación antes de escribir.** `checkProject` lee del disco y devuelve texto para mostrar. `validateHiveSection` mezcla en una sola lista lo que bloquea, lo que solo avisa y las claves desconocidas (`doctor_project.go:140-345`). Por eso se extrae un validador sobre texto:

```go
// checkHiveText validates a section's items without reading the disk.
func checkHiveText(items []hiveItem, root string, git gitRunner) (blocking, warnings []string)
```

`validateHiveSection` pasa a usarlo, y sus pruebas actuales siguen pasando sin cambios.
- **Bloquean la escritura:**
  - un obligatorio que falta o queda vacío;
  - un valor con saltos de línea o caracteres de control, o cuyo `sanitizeLine(v) != v`, como ya hace `checkBaseBranch` (`doctor_project.go:386`). Así la vista previa muestra exactamente lo que se escribe;
  - `Delivery` o `Hive guidance` inválidos;
  - una rama que no pasa `check-ref-format`;
  - una clave desconocida en `--set`;
  - un `--unset` de una clave obligatoria.
- **Solo avisan:** una rama o una ruta de `Specs` que no existen, y una clave desconocida que ya estaba en el archivo, que se conserva (AC6, AC8).

**Escritura** (`writeProjectFile(path string, before []byte, existed bool, after []byte) error`):
- **Comprobaciones previas:**
  - `Lstat` rechaza un enlace simbólico y lo que no sea archivo regular. La raíz viene de `git rev-parse --show-toplevel`, que da la ruta física, y `AGENTS.md` está justo en ella, así que basta con mirar el propio archivo.
  - Se rechaza un archivo de más de 1 MiB.
- **Archivo que ya existía:** se vuelve a leer y debe ser idéntico a `before`; si no, se rechaza con «AGENTS.md changed since the preview». Luego viene la escritura atómica:
  - un temporal en el mismo directorio con `chmod` explícito a los permisos del original;
  - `fsync` del temporal;
  - `rename` y `fsync` del directorio.
- **Archivo nuevo** (`existed` falso):
  - la segunda lectura debe dar «no existe»;
  - se escribe el temporal con `chmod 0644` explícito, para que la umask no lo cambie;
  - se publica con `os.Link(temp, path)`, que falla si alguien creó el archivo mientras tanto;
  - se borra el temporal y se hace `fsync` del directorio.
- **Errores:** ante cualquier error, el temporal se borra.
- **Riesgo aceptado:** entre la segunda lectura y el `rename` se puede perder una edición concurrente, y un hardlink del archivo queda desvinculado. Sin un bloqueo compartido con los editores no hay forma de evitarlo.
- **Sin journal:** el archivo pertenece al proyecto y Git es su recuperación.

**Sugerencias** (`suggestHiveValues(root) map[string]string`): son consultas de solo lectura con `newGitRunner`.
- **Lectura de `origin`:**
  - la URL `https` se lee con `net/url` y se descartan el usuario y la contraseña;
  - la forma `git@host:owner/repo(.git)` se lee aparte;
  - el host debe ser exactamente `github.com`;
  - `owner` y `repo` deben cumplir `[A-Za-z0-9._-]+`.

  Cualquier otra forma no da sugerencia de `Tracker`. Así un token incrustado en la URL nunca llega a la pantalla ni al archivo.
- **`Project`:** el `repo` de la URL sin `.git`, o el nombre del directorio de la raíz.
- **`Base branch`:** `git symbolic-ref --short refs/remotes/origin/HEAD` sin `origin/`. Si el resultado no pasa `checkBaseBranch`, no se sugiere.
- **`Tracker`:** `GitHub Issues · <owner>/<repo>`.
- **`Specs`:** `_support/openspec` cuando es un directorio.
- Las opcionales no se sugieren.

**Aviso de `CLAUDE.md`:** si existe `<raíz>/CLAUDE.md` y ninguna de sus líneas, sin espacios alrededor, es `@AGENTS.md`, el resumen y la confirmación dicen: «CLAUDE.md does not import @AGENTS.md; Claude Code will not read this section.»
- Fuente: `content/skills/harness-audit/references/hierarchy.md:49`. Es una afirmación de esa referencia que no se volvió a comprobar en la documentación de Claude Code en esta sesión; T4 la comprueba con Context7 antes de fijar el texto.

**Comando** (en `tooling/cli/project_command.go`; `main.go` agrega `case "project"`):

```
hive project set [--project DIR] --set 'Key: value'… [--unset Key…] [--dry-run]
```

- **Resumen:** muestra la ruta, las líneas de la sección antes y después, los avisos y «AGENTS.md has uncommitted changes» cuando `git status --porcelain -- AGENTS.md` lo lista.
- **Con terminal:** pide confirmación.
- **Sin terminal:** solo acepta `--dry-run`. Sin él termina con «an interactive terminal is required to confirm; use --dry-run to preview», como `voice set` (`voice.go:178`). No hay `--out` porque no hay plan de Hive, ni `--yes` porque ningún comando lo usa hoy. Un script puede editar el archivo por su cuenta, porque es del proyecto.

### Formulario en la vista Project

```
Project · /path/to/repo/AGENTS.md
> Project        tricell-hive
  Base branch    development
  Tracker        GitHub Issues · JhonHawk/tricell-hive   suggested
  Specs          _support/openspec
  Environments
  Review
  Delivery       ‹ none ›
  Hive guidance  ‹ none ›

```

- **Apertura:** `e` abre el formulario desde la vista de validación.
- **Campos:**
  - ↑↓ cambian de campo, y los obligatorios van primero, en el orden de `hiveSettingKeys`;
  - los de texto son `textinput`, y un valor más ancho que el campo se desplaza dentro de él sin salirse de la línea;
  - `Delivery` y `Hive guidance` se cambian con ←→.
- **Sugerencia:** se muestra con «suggested» en tono `Muted` y cuenta como valor, salvo que el usuario la borre.
- **Teclas:**
  - Esc cierra el formulario sin escribir y vuelve a la validación, sin salir de la vista (`navNone`);
  - con un campo de texto enfocado, `TextFocused()` es verdadero, y Backspace, `r` y `e` escriben en el campo;
  - en `Delivery` y `Hive guidance`, Backspace devuelve `navNone` y no cierra la vista.
- **Barra de ayuda (`Keys()`):**
  - formulario: `↑/↓ field · ←/→ choice · enter review · esc close`;
  - validación: la de hoy más `e edit`.
- **Guardar:** Enter con un obligatorio vacío deja el foco en ese campo y muestra «<Key> is required» en una fila de error de una línea recortada con «…». Si no falta ninguno, Enter abre `confirmView` con el resumen del comando.
- **Aplicar:** `OnApply` devuelve `action{write: true, result: v}`, como Voice, pero sin recuperación, porque no hay journal.
  - Cancel o Esc vuelven al formulario con lo escrito intacto y «Cancelled. No changes applied.».
  - Un resultado sin cambios dice «Nothing to change» sin abrir la confirmación.
  - Un error al escribir, como «AGENTS.md changed since the preview», se muestra en la fila de error y conserva lo escrito.
  - Si sale bien, el formulario se cierra y la vista vuelve a cargar la validación.
- **Fuera de un repositorio:** sin raíz Git, `e` no hace nada y la vista muestra las dos líneas de hoy.
- **Tamaño:** a 80×24 caben las 8 filas, el título, la ruta ajustada y la ayuda.

## Especificación

El delta está en `specs/versioned-installation/spec.md` de este cambio:
- **MODIFIED «Interactive terminal interface»:** Models y Project dejan de ser solo de lectura.
- **MODIFIED «Read-only diagnostics»:** `hive doctor`, `hive models` sin subcomando y las vistas mientras no se edita siguen sin escribir.
- **ADDED «Per-role model overrides».**
- **ADDED «Project settings section editing».**

## Límites de seguridad

- **Datos no confiables:** los valores del usuario y de `AGENTS.md` se validan contra caracteres de control y saltos de línea antes de escribir, y se muestran con `sanitizeLine`.
- **Modelos:** nunca se interpolan en un shell. Van al frontmatter del agente por el mismo camino que los perfiles de hoy.
- **Procesos:**
  - `git` con argumentos fijos por `newGitRunner`, sin shell;
  - los comandos de listado de modelos de T7, con las condiciones de su contrato y solo al abrir el panel de Models.
- **Rutas:** la escritura sigue rechazando enlaces simbólicos y un archivo cambiado entre la vista previa y la escritura.
- **Dependencias:** ninguna nueva.

## Compatibilidad

- **Planes guardados con `--out`:** los de antes del cambio no tienen `ModelOverrides`, así que su `planID` no cambia y se siguen aplicando. Un plan nuevo con ajustes es rechazado por un gestor anterior, porque no conoce el campo y su `planID` no coincidiría. Ese es el comportamiento deseado.
- **Estado:** un estado con ajustes que lee un binario anterior pierde los ajustes en el siguiente `apply`, porque `prepareTransaction` no los copia.
  - Se acepta porque el binario lo reconstruye el usuario desde este checkout (`AGENTS.md`, «Refreshing a local installation») y no hay instalaciones publicadas.
  - Queda escrito en `deployment-manager.md`.
