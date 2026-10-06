# Instalador de una línea en hive.tricell.tech

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado el 2026-10-05 · código en el [PR #12](https://github.com/JhonHawk/tricell-hive/pull/12) (merge `8154d5bc`); [v0.2.0](https://github.com/JhonHawk/tricell-hive/releases/tag/v0.2.0) en `master` (PRs #13 y #14); `hive.tricell.tech` en Cloudflare; repositorio público; AC1–AC8 cumplidos |
| Tracker · GitHub Issues | • [#11 — Instalador de una línea en hive.tricell.tech](https://github.com/JhonHawk/tricell-hive/issues/11) |
| Git | automática · rama → PR a `development` → `/code-review` → merge · Release y `master`, visibilidad del repositorio y Cloudflare con orden aparte |
| Verificación | pruebas del script (`go test ./tooling/distribution`) · `shellcheck` · suite local · prueba anónima en vivo tras publicar |
| Siguiente paso | ninguno en este cambio |

## Objetivo

Hoy, instalar Hive desde un Release exige descargar el paquete de la plataforma correcta y su `.sha256`, verificarlo, extraerlo y ejecutar `./install.sh`. Son cinco pasos manuales que un usuario nuevo tiene que leer en el README. Además, el repositorio es privado, así que esas descargas solo funcionan con una sesión de GitHub.

Este cambio ofrece la instalación con un solo comando, como hacen Codex y otras herramientas:

```sh
curl -fsSL https://hive.tricell.tech/install.sh | sh
```

El comando descarga un script pequeño que hace esos cinco pasos con las mismas verificaciones, deja el paquete en una carpeta estable para que `bin/hive` y la recuperación sigan disponibles, y después entrega el control al instalador del paquete, que sigue preguntando qué hosts instalar y pide confirmación antes de escribir. Para que funcione sin credenciales, el repositorio se hace público y `hive.tricell.tech` redirige al script.

## Alcance

Incluido:
- `get-hive.sh` en la raíz del repositorio, con el contrato de [design.md](design.md), y sus pruebas.
- Documentación: `README.md`, `_support/docs/architecture/installer.md`, `llms.txt` y una entrada en `CHANGELOG.md`.
- Delta de requisitos en `specs/versioned-installation/spec.md`.
- El siguiente Release (T3, D7-B): fija `VERSION`, actualiza el CHANGELOG, avanza `master`, crea la etiqueta, compila los paquetes y publica el Release. Se ejecuta solo con orden explícita.
- Configuración de `hive.tricell.tech` en Cloudflare (T4).
- Publicación del repositorio (T5), decidida (D1-A) pero ejecutada solo con orden explícita.

Excluido:
- **Paquete para macOS Intel.** El README ya lo declara no soportado; el script lo rechaza con el mismo mensaje que `install.sh`.
- **Firma de paquetes.** El checksum sigue siendo el control de integridad; viene del mismo origen, como ya documenta `installer.md`.
- **Poner `hive` en el `PATH`.** D3-A conserva el contrato actual: el script imprime la ruta de `bin/hive`.
- **Soporte de `wget`.** Las descargas de GitHub redirigen a otro dominio y `wget` no permite revisar cada salto; el propio comando de una línea ya requiere `curl`.
- **Reintroducir `hive bootstrap`.** El script es solo de shell; el binario no cambia.

Restricciones que deben seguir cumpliéndose:
- La suite local completa pasa (`go vet ./...`, `go test ./...`, pruebas de skills).
- El paquete no incluye `get-hive.sh`: `tooling/package/main.go` copia una lista fija de archivos y no se modifica.
- `install.sh` del paquete y el binario no cambian; la instalación sin red desde el paquete sigue igual.

## Criterios de aceptación

- AC1. En una terminal, en macOS arm64 o Linux arm64/amd64, `sh get-hive.sh` resuelve la última versión, descarga el paquete y su `.sha256`, los verifica, deja el paquete en `${XDG_DATA_HOME:-$HOME/.local/share}/hive/packages/hive-<versión>-<os>-<arch>/`, borra sus descargas temporales y ejecuta el `install.sh` de esa carpeta con la terminal como entrada y con los argumentos que recibió. *Falso en la base cuando* `get-hive.sh` no existe.
- AC2. `sh get-hive.sh --dry-run` no exige terminal: descarga, verifica, deja el paquete y ejecuta `install.sh --dry-run`. Si hay terminal, la entrada de `install.sh` viene de ella para que el usuario elija los hosts en la vista previa. Si no la hay, la entrada no cambia y la elección de hosts requiere `--hosts`, como en el instalador. *Falso en la base cuando* `get-hive.sh` no existe.
- AC3. Sin terminal y sin `--dry-run`, el script termina con error y un mensaje claro antes de hacer cualquier petición de red. *Falso en la base cuando* `get-hive.sh` no existe.
- AC4. Ante un checksum que no coincide, un `.sha256` mal formado o con otro nombre de archivo, una etiqueta de versión inválida o de otro repositorio, una redirección a un dominio no permitido, una descarga mayor que el límite, o un archivo con entradas fuera de su carpeta raíz, el script termina con error, no ejecuta `install.sh`, no deja carpeta parcial y no cambia una carpeta del mismo paquete que ya existiera. *Falso en la base cuando* `get-hive.sh` no existe.
- AC5. Una copia truncada del script, como la que deja una descarga interrumpida, no hace peticiones de red ni escribe archivos. *Falso en la base cuando* `get-hive.sh` no existe.
- AC6. `installer.md`, el README y `llms.txt` describen el comando de una línea, `--dry-run`, `--version`, la carpeta donde queda el paquete y sus requisitos; `installer.md` ya no dice que no hay instalador de una línea. *Falso en la base cuando* `rg -n 'There is no one-line online' _support/docs/architecture/installer.md` encuentra la frase.
- AC7. `curl -sI https://hive.tricell.tech/install.sh` responde `302` con `Location: https://raw.githubusercontent.com/JhonHawk/tricell-hive/master/get-hive.sh`. *Falso en la base cuando* `hive.tricell.tech` no resuelve (comprobado el 2026-10-05).
- AC8. Sin sesión de GitHub y sin terminal, `curl -fsSL https://hive.tricell.tech/install.sh | sh -s -- --dry-run --hosts claude` termina mostrando la vista previa del instalador para la última versión publicada. *Falso en la base cuando* el repositorio responde 404 sin credenciales (comprobado el 2026-10-05).

## Decisiones del usuario

Todas del 2026-10-05, en esta sesión:
- D1-A: publicar el repositorio. La ejecución requiere una orden explícita en el momento (T5).
- D2-A: URL `hive.tricell.tech`.
- D3-A: carpeta estable para el paquete, sin tocar el `PATH`.
- D4-A: entrega automática hacia `development`.
- D5-A: revisión con `/code-review` de Claude Code.
- D6-A: issue #11.
- D7-B: el script llega a `master` con el siguiente Release, no con una promoción suelta, para que `master` contenga solo versiones publicadas. La versión y la publicación se ejecutan con la orden de T3.

## Entrega

| Repositorio | Base | Modo | Incluye | Excluye |
| --- | --- | --- | --- | --- |
| `JhonHawk/tricell-hive` | `development` | automática (D4-A) | rama `feat/gh-11-one-line-installer`, commits, push, PR que cierra #11, `/code-review` sobre el diff, merge cuando pasen la suite local y la revisión, `hive update` local después del merge | Release y promoción a `master` (T3), cambio de visibilidad (T5), Cloudflare (T4) |

- La revisión `/code-review` (D5-A) corre sobre el PR antes del merge; los hallazgos se corrigen o se refutan con evidencia. Ejecutarla está autorizado por D5-A.
- Tras el merge a `development`, el cambio pasa a «En validación» con T3–T5 pendientes de orden. #11 se cierra con el merge, como pide la guía, porque el código queda integrado. La descripción del PR dice que el comando de una línea queda pendiente de T3–T5. El registro sigue abierto hasta que se cumplan AC7 y AC8.
- Este registro se versiona una sola vez, en el cierre con `flow-close` después de T5, con un PR pequeño a `development`, sin revisión dedicada, que yo fusiono. Si el usuario abandona T3–T5, el cierre registra AC7 y AC8 como no cumplidos.
- `hive update` después del merge: el cambio no toca `content/` ni el código del binario (en `tooling/distribution` solo entra un archivo de prueba), así que no hace falta recompilar y la actualización es prácticamente vacía. Se ejecuta igual, por la regla del checkout del mantenedor.
- El repositorio no tiene CI (retirada el 2026-09-30); los controles son la suite local.
- El Release con su promoción a `master` (T3), configurar Cloudflare (T4) y publicar el repositorio (T5) necesitan cada uno su propia orden. El orden recomendado está en [tasks.md](tasks.md).
