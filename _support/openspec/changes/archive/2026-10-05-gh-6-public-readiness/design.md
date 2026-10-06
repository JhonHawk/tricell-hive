# Diseño

## Contexto verificado (2026-10-05, base `47a9032`)

- **Repositorio.** `JhonHawk/tricell-hive` es privado. Solo existe la rama `development`, que es la principal y no tiene protección (`gh api …/protection` devuelve 404). No hay etiquetas, `refs/pull`, carpeta `.github/` ni CI. Las pruebas son locales; [deployment-manager.md](../../../docs/architecture/deployment-manager.md) describe la suite.
- **Hosts soportados.** `tooling/cli/install.go:22` declara claude, codex, cursor, grok, opencode y pi.
  - Solo Claude y Codex aceptan instalación por proyecto (`deployment-manager.md:181`).
  - `install.sh` no pone `hive` en el PATH (`installer.md:6`).
  - La línea que necesita Cursor está en `content/skills/harness-audit/references/instruction-files.md:38-41` y va fuera de `## Hive`.
- **Paquete.**
  - `go run ./tooling/package --out DIR` copia el árbol de trabajo, no una etiqueta. Construye `hive-<versión>-<os>-<arch>.tar.gz` y su `.sha256` para macOS arm64 y Linux arm64 y amd64 (`main.go:22`), además de binarios sueltos y un `index.json` para un sitio web propio.
  - Rechaza reconstruir una versión que ya existe en la carpeta de salida.
  - Solo copia `content`, `integrations/agent-profiles.json` e `install.sh` (`main.go:113`). No incluye `LICENSE` ni avisos de terceros.
  - Los archivos de más no rompen la comprobación del instalador (`manifest.go:123`).
- **`bootstrap.sh`.** Espera un sitio web propio y rechaza las redirecciones a otro dominio (`bootstrap.sh:79-84`).
  - Lo usan el subcomando `hive bootstrap` y el mensaje de recuperación de `install.go:520`.
  - `tooling/distribution/bootstrap_shell_test.go:827` contiene también `TestInstallOfflinePathMakesNoNetworkRequests`, la única prueba de que `install.sh` no usa la red.
- **`hive update`.** Despliega el commit indicado (`--rev`, por omisión `HEAD`), nunca trae cambios del remoto (`tooling/cli/update.go:33,254`) y no reemplaza el binario.
  - `go build` sin `-ldflags` deja la versión en `dev` (`tooling/version/version.go:13`).
  - El commit de origen solo se escribe en un registro (`apply.go:543`).
- **Identificadores de clientes.** El inventario está en `<ws>` (ver [proposal.md](proposal.md#alcance)).
  - Son cinco nombres de proyecto, sus compuestos con guion, una clave de tickets y una ruta personal.
  - Línea base: 39 archivos en `HEAD` (incluido el archivo histórico), 373 líneas de cambios y 86 líneas de mensajes en el historial.
  - En `HEAD`, fuera del archivo histórico, aparecen en `content/guidance/global.md`, `content/skills/flow-plan/references/delivery-decisions.md`, `_support/docs/harness-engineering/harness-audit-rules.md`, `_support/docs/architecture/agent-delivery.md`, `tests/pilot/regression.go`, `tests/fixtures/regression/**` y `tooling/cli/tui_wrap_test.go:76`.
  - También hay identificadores de sesiones reales en `regression.go:1741`, en `tests/fixtures/regression/README.md` y en archivos `provenance.md`.
- **Límite de tamaño.** `global.md` mide 43,621 bytes y el límite es 43,622 (`tests/content/budget_test.go:12`). Cualquier reemplazo debe medir lo mismo que el original o menos.
- **Ruta personal en la prueba.** `tui_wrap_test.go` corta la cadena con `p[:120]`, así que el reemplazo debe tener al menos 120 caracteres, contener `/` y no contener `-`.
- **SHAs.** `history-and-provenance.md` es el único archivo que cita commits de este repositorio: 8 SHAs propios, más 7 del repositorio anterior.
- **Instrucciones cargadas por los agentes.** El `CLAUDE.md` raíz importa `@AGENTS.md`, y Codex, Cursor y OpenCode leen `AGENTS.md`. Lo que diga `AGENTS.md` aplica al agente de cualquiera que abra el repositorio.
- **`llms.txt`.** Según la especificación en [llmstxt.org](https://llmstxt.org/) (modificada el 2026-08-10, consultada por el revisor), lleva un H1, un resumen opcional en cita, secciones H2 con listas `[nombre](url): notas` y una sección `Optional` para lo que un agente puede omitir.
- **Imágenes.** Codex CLI 0.160.1 tiene la generación de imágenes activada (`image_generation stable true`). El plugin `codex:codex-rescue` delega tareas con escritura. *Supuesto por confirmar en T1:* dónde guarda Codex las imágenes.
- **`git-filter-repo`.** Según su [documentación oficial](https://github.com/newren/git-filter-repo/blob/main/Documentation/git-filter-repo.txt):
  - `--replace-text` acepta `regex:` con la sintaxis de Python, incluido el lookbehind de ancho fijo;
  - `--replace-message` usa el mismo formato;
  - no renombra rutas, y ninguna ruta del historial contiene los identificadores;
  - quita el remoto `origin` de la copia reescrita.

## Decisiones

| ID | Decisión (usuario, 2026-10-05) | Efecto |
| --- | --- | --- |
| D1-B | Código fuente y paquete | Hace falta publicar un paquete con `LICENSE` y los avisos de terceros. |
| D2-A | Nota de audiencia y `llms.txt` | `AGENTS.md` sigue siendo la guía de quien mantiene Hive. La entrada pública es el README más `llms.txt`. |
| D4-A | `master` como rama de versiones | `development` sigue siendo la base y `master` recibe promociones. Se declara `Environments: development → master`. |
| D5-A | GitHub Releases | Se publican los `.tar.gz` y sus `.sha256` en el Release `v<versión>`. |
| D6-A | 0.1.0 | Cambia `VERSION`, el README y `CHANGELOG.md`. |
| D7-B | Reescribir el historial | Se usa `git filter-repo` sobre el contenido de los archivos y los mensajes, y `development` se reemplaza a la fuerza en GitHub. |
| D8-A | Retirar `bootstrap.sh` | Se eliminan el script y sus pruebas, y la prueba de `install.sh` sin red pasa a su propio archivo (revisión del plan, sobre la intención de D8-A). El subcomando `hive bootstrap` queda sin uso y se propondrá un ticket. |
| D9-A | Entrega interactiva | Ver [proposal.md](proposal.md#entrega). |
| D10-A | `/code-review` | Se ejecuta sobre el PR antes del merge. |
| D11-A | Regla de refresco limitada | Sigue en `AGENTS.md`, pero solo aplica en el checkout del maintainer, así que el agente de un colaborador no la ejecuta. La segunda revisión del plan comprobó que `hive status` y el estado guardado no registran el checkout de origen, de modo que esa condición no se puede comprobar. **D13-A, confirmada por el usuario el 2026-10-05:** la regla aplica solo si existe `_support/workspace/maintainer.local`, un archivo que Git ignora y que se crea en el checkout del maintainer. |
| D12-A | Reemplazar la clave de tickets | La clave del cliente pasa a `ARK` en los archivos actuales y en el historial. |

## Enfoque

- **Nombres de ejemplo.**
  - La tabla está en `<ws>/replacements.local.txt` y la usan tanto T2 (`HEAD`) como T9 (historial), para que los textos coincidan.
  - El nombre principal y la clave de tickets pasan a `ark`/`ARK`, que miden lo mismo que el original, para que `global.md` no supere su límite. Los demás pasan a `globex`, `initech`, `umbrella` y `hooli`.
  - Las reglas reconocen el nombre solo o como parte de un compuesto con guion, con variantes de mayúsculas.
  - La ruta personal pasa a `/home/user/…`. En `tui_wrap_test.go` se escribe a mano con la longitud que exige la prueba.
  - Los identificadores de sesión pasan a UUID sintéticos solo en `HEAD`, porque no revelan nada.
- **Reescritura del historial (T9).**
  0. `git fetch origin` en el checkout principal, con `development` en el mismo commit que `origin/development` y `git status --porcelain --untracked-files=no` vacío. Así la copia incluye el merge del PR y no se pierden ediciones.
  1. Respaldo completo con `git bundle create <ws>/pre-rewrite.bundle --all`. `<ws>` sobrevive porque el checkout no se vuelve a clonar.
  2. Copia nueva con `git clone --no-local` en el scratchpad, conteo previo con los comandos de AC2, y revisión a mano de cualquier coincidencia que no sea de un cliente.
  3. `git filter-repo --replace-text <ws>/replacements.local.txt --replace-message <ws>/replacements.local.txt`.
  4. Commit encima con los 8 SHAs propios actualizados según `.git/filter-repo/commit-map`.
  5. Suite completa y comprobaciones de AC1 y AC2 en la copia.
  6. `git remote add origin …` y `git push --force origin development`.
  7. En el checkout principal, con las condiciones del paso 0 todavía cumplidas, `git fetch origin` y `git reset --hard origin/development`. El material ignorado (`_support/workspace/`) y este plan, que no está versionado, se conservan.
- **`AGENTS.md` (T3).**
  - Nota de audiencia: este archivo gobierna el mantenimiento de Hive en este repositorio y no es guía para usar Hive en otros proyectos.
  - La regla de refresco queda limitada según D11-A.
  - Se elimina la sección de fuentes privadas. El frente de investigación pasa a «la implementación actual en la rama base». La línea de instalaciones desde paquete deja de citar `bootstrap.sh`.
- **README y `llms.txt` (T4).**
  - El README queda como la entrada pública, con este orden: banner, qué es, instalación (paquete y código fuente), hosts con sus requisitos, actualizar, verificar, recuperar, desinstalar, flujo de trabajo, sección para agentes de IA y documentación.
  - `llms.txt` usa enlaces relativos mientras el repositorio sea privado, y pone `_support/` y `tests/` en `Optional`.
- **Paquete (T6).** `LICENSE` y `THIRD_PARTY_NOTICES.md` se agregan a la lista de copia de `main.go:113`. Los avisos reúnen la licencia BSD-3 de Go y las de los módulos de `go.mod`, tomadas de la caché de módulos (`go mod download -json`).
- **Identidad visual (T1).**
  - Hubo dos rondas de conceptos. El usuario descartó la primera (C1–C3, sobria) y pidió una colmena más tecnológica: neón rojo sobre negro al estilo de Tron: Ares, con el concepto de la instalación subterránea de Resident Evil.
  - **Elegido: C4** (2026-10-05): seis cámaras alrededor de un núcleo.
  - Como la imagen generada no estaba alineada, Codex reconstruye C4 como SVG con geometría calculada. Los archivos fuente son `hive-mark.svg` y `hive-logo.svg`.
  - Los PNG se renderizan con Chrome sin interfaz.
  - La alineación se comprueba comparando cada imagen con su reflejo horizontal: la diferencia media debe ser menor que 0.005.
  - El banner y la imagen para redes sociales llevan fondo negro propio. `logo-light.png` es una variante sin resplandor para fondos claros.
  - `.github/` no entra en el paquete.
- **Release (T10).** Se construye desde una copia limpia de `v0.1.0` y se suben solo los `*.tar.gz` y `*.tar.gz.sha256`. Mientras el repositorio sea privado, descargarlos requiere iniciar sesión.

## Recuperación

- Antes de reemplazar `development` en GitHub, `<ws>/pre-rewrite.bundle` permite restaurar las referencias anteriores.
- `tricell-hive-private` conserva el historial anterior.
- Un Release o una etiqueta se pueden borrar y volver a crear antes de que el repositorio sea público.
