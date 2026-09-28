# Diseño

## Contexto verificado

Inspeccionado el 2026-09-27 sobre `d46de13` (`rebuild/harness-engineering`).

- **Comandos del gestor:** `run` en [`tooling/cli/main.go`](../../../../tooling/cli/main.go) atiende `--version`, `--help`, `setup`, `install` y `bootstrap` antes del `switch`, y `plan`, `apply`, `status` y `recover` dentro de él (líneas 26 a 174, 149 líneas). Escribe directamente en `os.Stdout`. No hay comando que liste releases ni que actualice desde un commit.
- **Pruebas de `run`:** [`tooling/cli/main_test.go`](../../../../tooling/cli/main_test.go) tiene tres pruebas y ningún helper para capturar la salida estándar. `plan`, `apply` y `status` no tienen pruebas de despacho en este paquete.
- **Instalación interactiva:** `runOnboardingWizard` en [`tooling/cli/install.go`](../../../../tooling/cli/install.go) arma el plan con `management.BuildPlan`, lo compara con `management.PlanUnchanged`, muestra `showInstallSummary` (líneas 649 a 723) y confirma con `confirmInstall` (línea 879). Sin terminal, `confirmInstall` falla con «an interactive terminal is required to confirm; use --dry-run to inspect». Los CLIs ya registrados se obtienen con `management.RegisteredHosts` (`install.go:475`).
- **CLIs registrados:** `RegisteredHosts` (`tooling/management/types.go:236`) lee los consumidores de `Records`, filtrados por alcance y por el home canónico. En cambio, `state.installations` no sirve para esto. `updateProductState` borra la instalación de un CLI cuando el plan no tiene identidad de producto (`versioning.go:107`), y eso pasa en todo `plan install --release` (`versioning.go:46`).
- **Extracción segura:** `distribution.Extract` en [`tooling/distribution/bootstrap.go`](../../../../tooling/distribution/bootstrap.go) (línea 246) expande un `.tar.gz` en `MkdirTemp(destination, ".hive-extract-")` y devuelve `<esa carpeta>/<prefijo>`. Tiene límites (128 MiB comprimido, 256 MiB extraído, 10 000 entradas) y solo acepta directorios y archivos regulares. Exige que la primera entrada sea el directorio raíz, y `archivePath` (línea 336) rechaza nombres que cambien con `path.Clean`.
- **Salida de `git archive`:** empieza con una cabecera global pax (`pax_global_header`, tipo `g`, con el commit en el comentario) y escribe los directorios con barra final (`hive-x/`). `Extract` la rechaza tal cual. Lo comprobaron el revisor del plan y una lectura con `tarfile` sobre `d46de13`.
- **Tamaño del repositorio:** `git archive --format=tar.gz HEAD` pesa 3,2 MB, con 402 archivos versionados, dentro de esos límites. El repositorio no versiona enlaces simbólicos ni tiene `.gitattributes`.
- **Snapshots de release:** `Apply` reescribe `releases/<id>.json` con archivo temporal y renombrado en cada aplicación con cambios. Falla con «release cache conflict» si los bytes difieren (`tooling/management/apply.go:338`). Un snapshot solo tiene `id`, `files`, `profiles` y `renderer`. Solo `loadRelease` lee esa carpeta, por un ID de 64 caracteres (`plan.go:101`); nada la recorre.
- **Resultado «sin cambios»:** `Apply` devuelve `"unchanged"` antes de escribir nada, con el bloqueo ya tomado (`apply.go:382` y `388`). En este repositorio, los commits que solo tocan `tooling/` o `_support/` producen una release idéntica.
- **Escritura atómica:** `write` en `tooling/management/files.go:89-120` usa archivo temporal, `Rename`, `fsync` del directorio y modo 0600.
- **ID de plan:** `planID` es el hash del plan codificado sin su ID (`types.go:118`). Un campo nuevo con `omitempty` deja igual el ID de los planes que no lo usan.
- **Sin Git en el gestor:** ningún paquete de `tooling/` ejecuta `git`; solo `tests/pilot/main.go`. La especificación `versioned-installation` exige que el paquete offline funcione sin Git.
- **Sin CI:** el repositorio no tiene `.github/workflows`. Las pruebas se ejecutan localmente.

## Diseño elegido

### `hive update`

```
hive update [--rev REV] [--source DIR] [--home DIR] [--state-dir DIR] [--dry-run] [--out FILE]
```

`--rev` vale `HEAD` por defecto y `--source` vale `.`, igual que en `plan` e `install`.

1. **Elegir los CLIs.** Se usa `management.RegisteredHosts` con alcance `user`. Si no hay ninguno, el comando falla y remite a `hive install`. Este paso va antes de ejecutar Git, para fallar sin extraer nada. Elegir todos los consumidores registrados cumple la regla de recursos compartidos: nunca queda un consumidor registrado sin elegir.
2. **Validar la revisión.** Un `--rev` vacío o que empieza por `-` se rechaza antes de ejecutar Git. Con el sufijo `^{commit}` añadido, eso basta para que no se lea como una opción.
3. **Ejecutar Git de forma controlada.**
   - `git` se busca en el `PATH`. Si falta, el error dice que `update` necesita Git y un checkout, y que el paquete offline se instala con `install.sh`.
   - Los procesos hijos se ejecutan con argumentos separados, sin shell, y sin las variables que `git rev-parse --local-env-vars` lista (`GIT_DIR`, `GIT_OBJECT_DIRECTORY`, `GIT_CONFIG_PARAMETERS` y el resto), porque tienen prioridad sobre `-C`. Ampliado tras `/code-review`: la versión inicial solo quitaba tres.
   - **Resolución:** `git -C <source> rev-parse --verify --quiet <rev>^{commit}`. Si `source` no es un checkout de Git o el commit no existe, falla nombrando la causa. El resultado es el hash completo (40 o 64 caracteres hexadecimales).
4. **Extraer el commit.**
   - `git -C <source> archive --format=tar --prefix=hive-<hash corto>/ <hash completo>`, leído con un límite de `distribution.MaxPackageBytes + 1` bytes.
   - Se pide `tar` y no `tar.gz` para no ejecutar el compresor que configure el repositorio (`tar.tar.gz.command`).
   - Un adaptador en `tooling/cli` recorre ese tar:
     - descarta las cabeceras globales pax (`tar.TypeXGlobalHeader`);
     - quita la barra final a los nombres de directorio;
     - copia las demás entradas y comprime el resultado en gzip en memoria.
   - `distribution.Extract(<gzip>, os.TempDir())` recibe esa salida. `Extract` no cambia: sigue siendo el límite de seguridad del bootstrap.
   - Se archiva por hash, no por el nombre que dio el usuario. Se extrae el commit completo, como en el despliegue manual, para no acoplar el comando a los archivos que lee el catálogo.
5. **Planear y limpiar.**
   - `management.BuildPlan("install", …)` con `Source` en la raíz extraída, los CLIs del paso 1 y el commit en el plan.
   - La limpieza borra `filepath.Dir(root)`, la carpeta `.hive-extract-*` completa, y no solo la raíz.
   - Se ejecuta con `defer` en todo camino de retorno después de la extracción, antes de pedir confirmación: el plan ya guarda una copia fija de los archivos.
   - Una señal que mate el proceso durante la extracción puede dejar la carpeta; el gestor no maneja señales hoy.
6. **Resumir.** Se reutiliza `showInstallSummary` sin capacidades opcionales y se añade una línea con el hash corto y la revisión pedida.
7. **Aplicar o guardar.**
   - **Con `--dry-run`:** termina tras el resumen, con o sin terminal.
   - **Con `--out FILE`:** `management.SavePlan` y un mensaje que indica `hive apply --plan FILE`, también cuando el plan no cambia nada, para que un script que encadena `update --out` y `apply` encuentre el archivo. Si el plan trae ediciones de legado, `SavePlan` ya rechaza guardarlo con su propio mensaje.
   - **Sin cambios** (`PlanUnchanged`), sin `--out`: informa que Hive ya está al día y llama a `management.Engine{}.Apply` sin preguntar. `Apply` no escribe ningún destino y registra el commit ([Commit de origen](#commit-de-origen)).
   - **Con terminal:** `confirmInstall(…, allowBack=false)` y `management.Engine{}.Apply`.
   - **Sin terminal y sin `--out`:** falla antes de pedir confirmación, con un mensaje que nombra `--dry-run` y `--out`.

`update` no ofrece capacidades opcionales ni abre el journal de onboarding. Un onboarding o una operación pendiente bloquean el plan como hoy, porque `BuildPlan` ya los comprueba.

Con un `--rev` antiguo, los atributos `export-ignore` o `export-subst` de ese commit alterarían el contenido extraído, igual que en el despliegue manual. Hoy no hay `.gitattributes`, y la documentación de T5 lo menciona.

**Por qué no reutilizar `runOnboardingWizard`:** su recorrido pregunta por CLIs y por capacidades opcionales, y hace la vinculación del instalador retenido de bootstrap. `update` no hace nada de eso. Un flujo propio y corto que reutiliza el resumen, la confirmación y `Apply` evita ramas condicionales dentro del asistente.

**Por qué un adaptador y no relajar `Extract`:** `Extract` protege el bootstrap en línea, que descarga archivos de la red. Adaptar la salida de Git en el lado del llamador deja intacto ese límite y sus pruebas.

### Commit de origen

- **En el plan:** `Plan` gana `SourceCommit string` con `json:"source_commit,omitempty"`. `validatePlan` acepta vacío o 40 o 64 caracteres hexadecimales en minúscula y rechaza cualquier otro valor.
- **Dónde se escribe:** dentro de `Engine.Apply`, con el bloqueo del gestor todavía tomado, en dos puntos. Por el bloqueo, dos `apply` concurrentes no pierden entradas en esta operación de leer, modificar y escribir.
  - Cuando el plan no cambia nada, justo antes de devolver `"unchanged"`. La release instalada ya coincide con el contenido de ese commit.
  - Después de que `commitTransaction` termina bien, junto a `finishApply`. Nunca dentro de `commitTransaction`: `Recover` no lo desharía y quedaría un commit registrado para una release que se deshizo.
- **Qué se escribe:** si el plan trae `SourceCommit` y `Release`, el gestor agrega `{commit, applied_at}` a `releases/<id>.commits.json` cuando ese commit no está ya en la lista. Usa `write` (atómica, 0600). El snapshot `releases/<id>.json` no se toca.
- **Por qué fuera del snapshot y del estado:** el snapshot se compara byte a byte. Meter el commit en `state.json` exigiría el esquema 7 y su migración. Un archivo aparte deja intactos el ID de la release y el esquema.
- **Si falla la escritura del registro:** la instalación ya está confirmada, así que el comando no falla. Muestra un aviso y la release queda sin ese commit.
- **Si se recupera un journal ya confirmado:** la release queda sin commit, porque `recover` no conoce el plan.
- **Releases anteriores:** las que ya existen no tienen registro y aparecen con la lista de commits vacía. No se reconstruye nada.

### `hive releases`

```
hive releases [--home DIR] [--state-dir DIR]
```

- **Función:** `management.Releases(o Options) ([]ReleaseEntry, error)` lee los `releases/<id>.json` con ID de 64 caracteres hexadecimales, lo que excluye los `*.commits.json`.
- **Qué muestra de cada release:** la fecha de modificación del snapshot, que corresponde a la última aplicación que lo escribió; su registro de commits si existe; y los consumidores de `state.Records` cuyo campo `Release` coincide.
- **`ReleaseEntry`:** `ID`, `LastWrittenAt` (RFC 3339), `Commits []string` (en el orden en que se aplicaron) y `Consumers []Consumer` (sin duplicados).
- **Orden:** del `LastWrittenAt` más reciente al más antiguo, y por ID cuando dos fechas coinciden.
- **Salida:** JSON indentado, igual que `status`, para que la TUI futura lo consuma.
- **Solo lectura.** Un snapshot con JSON inválido o un registro de commits ilegible hacen fallar el comando con la ruta del archivo; no se omiten en silencio.
- **Volver atrás:** a una release, sigue siendo `plan install --release <hash>`. A un commit, es `hive update --rev <commit>`, con el límite ya documentado: el gestor actual tiene que poder validar las fuentes de ese commit.

### Despacho de subcomandos (#42)

`run` resuelve `--version` y `--help` y delega en una función por subcomando: `setup`, `install`, `bootstrap`, `plan`, `apply`, `status`, `recover`, `update` y `releases`. El texto de ayuda pasa a su propia función. Cada función define sus opciones con su propio `flag.FlagSet`.

Antes del refactor, las pruebas de caracterización de T1 fijan el comportamiento visible de `plan`, `apply`, `status` y los errores de uso. Capturan la salida estándar con un helper basado en `os.Pipe`.

## Límites de seguridad

- **Ejecución de procesos:** Git se ejecuta con argumentos separados, sin shell y sin las variables `GIT_DIR`, `GIT_WORK_TREE` y `GIT_COMMON_DIR` heredadas. La revisión del usuario solo llega a `rev-parse`, después de rechazar el prefijo `-`, y `archive` recibe el hash ya resuelto. Se pide `tar` sin comprimir para no ejecutar un compresor configurado en el repositorio.
- **Archivo extraído:** es el propio checkout del usuario, una fuente de confianza. Aun así pasa por `distribution.Extract`, que rechaza rutas fuera del directorio, entradas que no sean archivos o directorios, y tamaños fuera de límite. El adaptador solo descarta cabeceras pax globales y normaliza la barra final de los directorios.
- **Carpeta temporal:** `Extract` la crea con `os.MkdirTemp` (modo 0700), y el comando la borra entera en todo camino de retorno.
- **Dependencias:** ninguna nueva.

## Compatibilidad

- **Productores y consumidores de `Plan`:** los dos son este mismo binario. Un plan guardado por un binario anterior no tiene `source_commit` y conserva su ID. Un plan con `source_commit` cargado por un binario anterior falla la comprobación de ID y hay que regenerarlo, igual que en cualquier cambio de formato de plan.
- **Registro de commits:** un binario anterior lo ignora, porque solo lee `releases/<id>.json` por ID.
