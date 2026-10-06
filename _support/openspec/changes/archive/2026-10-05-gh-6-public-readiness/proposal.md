# Preparar Hive para publicarlo (0.1.0)

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado el 2026-10-05 · integrado por el el PR #7 del repositorio previo a la publicación (merge `e61bf3c` tras la reescritura); `master` y [v0.1.0](https://github.com/JhonHawk/tricell-hive/releases/tag/v0.1.0) publicados · la visibilidad pública queda fuera (autorización aparte) |
| Tracker · GitHub Issues | • [#6 — Preparar el repositorio para publicarlo](https://github.com/JhonHawk/tricell-hive/issues/6) |
| Git | interactiva · rama → validación de la documentación y las imágenes → PR → `/code-review` → merge a `development` → reescritura del historial → `master` → `v0.1.0` |
| Verificación | suite local · prueba del paquete · instalación de prueba del paquete descargado |
| Siguiente paso | ninguno en este cambio |

## Objetivo

Hive se va a publicar, pero hoy alguien que llegue al repositorio no sabe cómo instalarlo, actualizarlo, recuperarlo o desinstalarlo en cada host. Tampoco hay un paquete descargable. Una IA que lea el repositorio no distingue lo que se instala en los hosts de lo que es material interno. Además, los archivos actuales y el historial contienen identificadores de proyectos de clientes, una ruta personal y referencias a una rama `master` que ya no existe.

Este cambio deja el repositorio listo para hacerse público:
- Instrucciones completas por host.
- Una entrada clara para agentes de IA (`llms.txt`, una sección en el README y una nota de audiencia en `AGENTS.md`).
- Una identidad visual.
- Un historial sin identificadores de clientes.
- Una rama `master` de versiones.
- La primera versión publicada, `v0.1.0`, con su paquete en GitHub Releases.

## Alcance

Incluido: los puntos H1–H5, A1–A4 y E1–E4 de #6, la identidad visual, y las decisiones D1–D12 descritas en [design.md](design.md).

Excluido:
- **Cambiar la visibilidad del repositorio y su rama principal en GitHub.** Necesita una autorización aparte (decisión del 2026-10-05).
- **Subir la imagen para redes sociales en la configuración de GitHub.** GitHub no ofrece API para eso; queda como acción manual del usuario.
- **Retirar o reconectar el comando `hive bootstrap`.** D8-A retira solo el script; el comando queda sin uso y se propondrá como ticket aparte.
- **Firmar y notarizar el binario para macOS.** `installer.md` ya lo aplaza; el README documenta el aviso de Gatekeeper.
- **Traducir los registros archivados en `_support/openspec/changes/archive/`.** Son registros escritos para personas en el idioma de la sesión. La reescritura del historial sí reemplaza en ellos los identificadores de clientes.
- **Limpiar `JhonHawk/tricell-hive-private`.** Se queda privado (decisión del usuario del 2026-10-05).

Restricciones que deben seguir cumpliéndose:
- La suite local completa pasa.
- `hive status` verifica la instalación después de `hive update`.
- `content/guidance/global.md` sigue bajo el límite de `tests/content/budget_test.go`.
- Ningún SHA propio citado en `history-and-provenance.md` deja de resolver después de la reescritura.
- Este registro y cualquier archivo nuevo usan solo los nombres de ejemplo. Los identificadores reales y las reglas de reemplazo viven en `_support/workspace/2026-10-05-gh-6-public-readiness/`, que Git ignora; abajo se cita como `<ws>`.

## Criterios de aceptación

`<patrón>` es `<ws>/client-names.pattern`, que reúne los identificadores de clientes, la clave de tickets y la ruta personal.

- AC1. Ningún archivo versionado en `HEAD` coincide con `<patrón>`. *Falso en la base cuando* `git grep -ciE -f <patrón> HEAD` lista archivos; hoy son 39.
- AC2. Ningún cambio ni mensaje del historial de GitHub coincide con `<patrón>`. *Falso en la base cuando* `git log --all -p | grep -ciE -f <patrón>` (hoy 373) o `git log --all --format=%B | grep -ciE -f <patrón>` (hoy 86) dan más de 0.
- AC3. `AGENTS.md`:
  - abre con una nota de audiencia;
  - declara `Environments: development → master`;
  - no cita `reference-sources.local.md`;
  - no contiene la excepción «until the rebuild merges into `master`»;
  - limita la regla de refresco al checkout del maintainer, según D11;
  - no usa `master` como referencia de investigación;
  - no menciona `bootstrap.sh`.

  *Falso en la base cuando* `grep -cE 'reference-sources.local|until the rebuild merges|bootstrap\.sh' AGENTS.md` da más de 0.
- AC4. Existe `llms.txt` en la raíz con un H1, un resumen en cita, secciones H2 con enlaces y una sección `Optional` para el material interno, y todos sus enlaces apuntan a archivos versionados. *Falso en la base cuando* `test -f llms.txt` falla.
- AC5. El README nombra los seis hosts y explica:
  - **instalar** desde el paquete de Releases, verificando su suma SHA-256 y usando el `bin/hive` extraído, y desde el código fuente;
  - **actualizar**: con un paquete nuevo, o con `git pull`, recompilar y `update`;
  - **verificar** con `status` y `doctor`;
  - **recuperarse** con `recover`;
  - **desinstalar** con `plan remove` y `apply`;
  - **requisitos de cada host**: la línea exacta para Cursor fuera de `## Hive`, que solo Claude y Codex aceptan instalación por proyecto, la compatibilidad de Grok, Grok y Cursor juntos, `pi-subagents` y `PI_CODING_AGENT_DIR` en Pi;
  - la posibilidad de un aviso de Gatekeeper en macOS si el paquete se descarga con el navegador, y cómo resolverlo. No se ha comprobado que ocurra.

  *Falso en la base cuando* `grep -c 'plan remove' README.md` da 0.
- AC6. El README tiene una sección para agentes de IA que nombra `content/` e `integrations/` como contenido instalable y `_support/` y `tests/` como material interno. *Falso en la base cuando* `grep -n '^## .*agent' README.md` no devuelve esa sección.
- AC7. Existen `CONTRIBUTING.md`, `SECURITY.md` y `CHANGELOG.md`, y este último tiene una entrada 0.1.0. *Falso en la base cuando* falta alguno.
- AC8. Cada paquete de `tooling/package` contiene `LICENSE` y `THIRD_PARTY_NOTICES.md`, y una prueba de `tooling/package` lo comprueba. *Falso en la base cuando* el archivo del paquete no los incluye (comprobado el 2026-10-05).
- AC9. `bootstrap.sh` ya no existe. Ningún documento ni mensaje del CLI lo propone como forma de instalar o recuperar. La prueba de que `install.sh` no usa la red sigue existiendo. *Falso en la base cuando* `test -f bootstrap.sh` tiene éxito.
- AC10. Están en inglés los archivos de la lista de T8 (documentos, la muestra de `skeleton-paper.html` y los comentarios de Go). Las entradas de prueba en español a propósito quedan fuera. *Falso en la base cuando* esos archivos contienen prosa en español.
- AC11. `_support/docs/architecture/installer.md` incluye `cursor` en la lista de hosts. *Falso en la base cuando* la línea 13 lo omite.
- AC12. `VERSION` contiene `0.1.0`. *Falso en la base cuando* contiene `0.0.2`.
- AC13. Existen `.github/assets/logo-light.png`, `logo-dark.png`, `banner.png` y `social-preview.png` (1280×640), y el README muestra `banner.png`. *Falso en la base cuando* `.github/` no existe.
- AC14. GitHub tiene la rama `master`, la etiqueta `v0.1.0` y el Release `v0.1.0` con 3 `.tar.gz` y 3 `.sha256`. Un paquete descargado verifica su suma y `./install.sh --dry-run` termina con éxito. *Falso en la base cuando* `gh release view v0.1.0` falla.

## Entrega

Decisión del usuario del 2026-10-05: D9-A, modo interactivo; D10-A, `/code-review` de Claude Code.

1. Rama `chore/gh-6-public-readiness` desde `development`, con commits locales y la suite completa en verde.
2. **Pausa para que el usuario valide** el README, `llms.txt`, la nota de `AGENTS.md` y las imágenes.
3. Después de esa validación:
   1. push, PR a `development` y `/code-review`;
   2. merge cuando la revisión esté resuelta y la suite local pase;
   3. instalar `git-filter-repo` con Homebrew, reescribir el historial y reemplazar `development` en GitHub;
   4. crear `master` y empujarla;
   5. empujar la etiqueta `v0.1.0` y publicar el Release;
   6. recompilar `hive` y ejecutar `hive update`.
4. Este registro se versiona una sola vez al cerrar, con `flow-close`. Se archiva mediante un PR pequeño a `development`, sin revisión dedicada, que integro yo, después de la reescritura y comprobando AC1 antes del commit.

Cambiar el texto de #5 (quitar el identificador del cliente) es una escritura en el tracker. Se pedirá en la pausa de validación.
