# Tareas

Commit base: `47a9032137933d3e40e91e2c1c9c127ecb2280d0` (rama `chore/gh-6-public-readiness`, 2026-10-05). `<ws>` y `<patrón>` se definen en [proposal.md](proposal.md).

### T1 — Identidad visual

- [x] Logotipo claro y oscuro, banner e imagen para redes sociales en `.github/assets/`, sobre el concepto que elija el usuario.

**Closes:** AC13 (los archivos; el banner en el README lo cierra T4).

**Depends on:** ninguna.

**Locations:** `.github/assets/`.

**Execution:** delegada a Codex con el plugin `codex:codex-rescue`, en modo escritura y limitada a `.github/assets/`. El usuario pidió Codex y solo Codex tiene aquí la generación de imágenes. La indicación a Codex no incluye identificadores de clientes.

**Changes:**
1. Prueba corta: un concepto guardado en `.github/assets/concepts/`, para confirmar dónde guarda Codex las imágenes.
2. Tres conceptos. Se pregunta al usuario en texto cuál prefiere; esta pregunta no espera a la pausa general.
3. Codex produce los archivos finales.
4. Se borra `concepts/` salvo que el usuario quiera conservarla.

**Verification:**
- `ls .github/assets/{logo-light,logo-dark,banner,social-preview}.png` lista los cuatro archivos.
- `file` los identifica como PNG.
- `sips -g pixelWidth -g pixelHeight .github/assets/social-preview.png` da 1280×640.

### T2 — Identificadores de clientes, sesiones y ruta personal en `HEAD`

- [x] Se aplica `<ws>/replacements.local.txt` a los archivos versionados fuera del archivo histórico; T9 cubre el resto. Los UUID de sesión pasan a sintéticos.

**Closes:** AC1 (fuera del archivo histórico).

**Depends on:** ninguna.

**Locations:** los archivos listados en [design.md](design.md#contexto-verificado-2026-10-05-base-47a9032), en «Identificadores de clientes».

**Execution:** hilo principal. Los reemplazos son mecánicos, pero la redacción afecta a la guía que se instala.

**Test approach:** check, con `go test ./tests/content/... ./tooling/cli/ -run 'Budget|Wrap|DoctorProject'`.

**Changes:**
- En `global.md`, conservar la forma `- Key: value` y el encabezado «Project settings», y no aumentar su tamaño en bytes.
- La nueva ruta de `tui_wrap_test.go` tiene 120 caracteres o más, contiene `/` y no contiene `-`.

**Verification:**
- `git grep -ciE -f <patrón> -- ':(exclude)_support/openspec/changes/archive'` no lista archivos.
- `wc -c content/guidance/global.md` da 43,621 o menos.
- Pasan las pruebas indicadas.

### T3 — `AGENTS.md`

- [x] Nota de audiencia al inicio; `Environments: development → master`; regla de refresco limitada con el archivo marcador (D13-A), que también se crea en este checkout; sección de fuentes privadas eliminada; frente de investigación sin `master`; línea de instalaciones desde paquete sin `bootstrap.sh`; mapa del repositorio corregido (`_support/sessions/`).

**Closes:** AC3.

**Depends on:** ninguna.

**Locations:** `AGENTS.md` (líneas 30, 54-61 y 95 de la base).

**Execution:** hilo principal, porque es la guía que gobierna este repositorio.

**Test approach:** check.

**Verification:**
- `grep -cE 'reference-sources.local|until the rebuild merges|bootstrap\.sh' AGENTS.md` da 0.
- `grep -c 'Environments: development → master' AGENTS.md` da 1.
- `grep -n 'master' AGENTS.md` solo devuelve la línea de `Environments`.
- La regla de refresco cita `_support/workspace/maintainer.local` como condición.
- `git check-ignore _support/workspace/maintainer.local` confirma que Git lo ignora.
- La nota de audiencia está antes del primer encabezado `##`.

### T4 — README y `llms.txt`

- [x] El README es la entrada pública, con todo lo que pide AC5, la sección de AC6 y el banner.
- [x] `llms.txt` sigue la especificación citada en [design.md](design.md#contexto-verificado-2026-10-05-base-47a9032).
- [x] `deployment-manager.md` tiene un ejemplo de `plan remove` a nivel de usuario.

**Closes:** AC4, AC5, AC6, AC13 (el banner en el README).

**Depends on:** T1, T3, T5 (versión), T7 (forma de instalar).

**Locations:** `README.md`, `llms.txt`, `_support/docs/architecture/deployment-manager.md`.

**Execution:** delegada a `hive-write-spec`, porque es documentación con fuentes identificadas. El hilo principal revisa el texto antes de la pausa.

**Test approach:** check. Cada comando del README se prueba con `--help` o `--dry-run`.

**Verification:** con un script en el scratchpad, que debe imprimir solo coincidencias:
- **AC5:** `README.md` contiene las cadenas nuevas `sha256`, `bin/hive`, `git pull`, `recover`, `plan remove`, `PI_CODING_AGENT_DIR`, `Gatekeeper`, `--scope project` y la línea exacta para Cursor de `instruction-files.md:38-41`. Los nombres de los hosts, `status`, `doctor` y `apply` ya están en la base; se conservan y se revisan a mano.
- **AC6:** existe un encabezado `## ` sobre agentes de IA que nombra `content/`, `integrations/`, `_support/` y `tests/`.
- **AC13:** `grep -n '.github/assets/banner.png' README.md` coincide.
- **AC4:** `llms.txt` empieza con `# `, tiene una línea `> `, al menos un `## ` y un `## Optional`.
- **Enlaces:** cada enlace relativo de `README.md` y `llms.txt` apunta a un archivo versionado.
- **Desinstalación:** `go run ./tooling/cli plan remove --hosts codex --scope user` produce un plan sin aplicarlo.

### T5 — Versión 0.1.0 y documentos estándar

- [x] `VERSION` contiene `0.1.0`.
- [x] Existen `CHANGELOG.md` (con la entrada 0.1.0), `CONTRIBUTING.md` (suite local, convenciones, idioma, sin CI) y `SECURITY.md` (reporte privado de vulnerabilidades en GitHub).

**Closes:** AC7, AC12.

**Depends on:** ninguna.

**Locations:** `VERSION`, `CHANGELOG.md`, `CONTRIBUTING.md`, `SECURITY.md`.

**Execution:** delegada junto con T4 a `hive-write-spec`.

**Test approach:** check.

**Verification:**
- `cat VERSION` muestra `0.1.0`.
- Existen los tres archivos y `grep -n '0.1.0' CHANGELOG.md` coincide.

### T6 — `LICENSE` y avisos de terceros en el paquete

- [x] Cada paquete contiene `LICENSE` y `THIRD_PARTY_NOTICES.md`.

**Closes:** AC8.

**Depends on:** ninguna.

**Locations:** `tooling/package/main.go:113`, `tooling/package/main_test.go`, `THIRD_PARTY_NOTICES.md`.

**Execution:** delegada a `hive-build-backend`. La interfaz está definida y no comparte archivos con otras tareas.

**Test approach:** tdd. La prueba nueva abre el `.tar.gz` y exige los dos archivos, y debe fallar en la base.

**Verification:**
- `go test ./tooling/package/...` pasa con la prueba nueva.
- `tar -tzf <paquete> | grep -E 'LICENSE|THIRD_PARTY'` lista los dos archivos.

### T7 — Retirar `bootstrap.sh` y corregir la documentación del instalador

- [x] `TestInstallOfflinePathMakesNoNetworkRequests` y sus helpers pasan a `tooling/distribution/install_offline_test.go`.
- [x] Se eliminan `bootstrap.sh` y el resto de `bootstrap_shell_test.go`.
- [x] El mensaje de `install.go:520` deja de proponer `bootstrap.sh`.
- [x] La ayuda de `main.go:85` marca `hive bootstrap` como sin uso.
- [x] `installer.md` y `deployment-manager.md` describen la instalación desde Releases, e `installer.md:13` incluye `cursor`.

**Closes:** AC9, AC11.

**Depends on:** ninguna.

**Locations:** `bootstrap.sh`, `tooling/distribution/bootstrap_shell_test.go`, `tooling/distribution/install_offline_test.go` (nuevo), `tooling/cli/install.go:520`, `tooling/cli/main.go:85`, `_support/docs/architecture/installer.md`, `_support/docs/architecture/deployment-manager.md`.

**Execution:** hilo principal. Son movimientos y textos pequeños. Eliminaciones previstas: `bootstrap.sh` y `bootstrap_shell_test.go`, una vez movida la prueba de `install.sh`.

**Test approach:** check, con `go test ./tooling/distribution/... ./tooling/cli/...`, incluida `TestInstallOfflinePathMakesNoNetworkRequests`.

**Verification:**
- `test -f bootstrap.sh` falla.
- `go test ./tooling/distribution/ -run TestInstallOfflinePathMakesNoNetworkRequests -v` pasa.
- `git grep -n 'bootstrap\.sh' -- ':(exclude)_support/openspec/changes/archive'` solo devuelve comentarios de `tooling/cli/bootstrap*.go` y de `install.go` sobre el camino sin uso, ningún texto que vea el usuario.
- `grep -n cursor _support/docs/architecture/installer.md` coincide en la lista de hosts.

### T8 — Traducir al inglés

- [x] Se traducen:
  - **documentos:** `_support/docs/harness-engineering/2026-09-20-workflow-map.md`, `_support/docs/history/history-and-provenance.md` y el texto de muestra de `content/skills/flow-report/assets/skeleton-paper.html`;
  - **comentarios de Go en `tooling/cli/`:** `{tui_app,doctor,bootstrap,tui,tui_hosts,tui_hosts_test,tui_hosts_view,tui_project_view,tui_releases_view_test,project_command,project_write}.go`;
  - **comentarios de Go en `tooling/management/`:** `{plan,voice,voice_test,types,migration,files_test}.go`;
  - **comentarios en `tests/pilot/`:** `{flows,guidance_variant,regression}.go`, sin tocar las entradas de prueba.
- [x] Cuando un comentario cita un encabezado en español de un registro archivado, se traduce y se marca como cita del registro.
  - ronda 1, AC10: no cumplido. `tooling/cli/tui_hosts_view.go:437` conservaba un encabezado en español sin marcar. Corregido en el hilo principal, junto con los `id` y `href` en español de `skeleton-paper.html` (hallazgo no bloqueante). Ronda 2, AC10: cumplido.

**Closes:** AC10.

**Depends on:** ninguna. `history-and-provenance.md` recibirá después los SHAs nuevos en T9.

**Locations:** los indicados.

**Execution:** delegada a `hive-write-spec`, porque es traducción de documentación y comentarios sin cambiar código.

**Test approach:** check.

**Verification:**
- `rg -n -i -w '(el|los|las|para|que|con|una|del)'` sobre los documentos, y sobre las líneas de comentario de los archivos Go (`rg '^\s*//'`), no devuelve prosa en español.
- Quedan fuera las entradas de prueba de `tests/pilot/{main,flows}.go`, los `cases.json`, `del` como etiqueta de HTML y el ejemplo en español citado a propósito en `tests/pilot/regression_test.go:788`, marcado como tal.
- `go vet ./...` pasa.

### T9 — Reescribir el historial (después del merge)

- [x] El historial de `development` en GitHub no coincide con `<patrón>`.

**Closes:** AC1 (archivo histórico), AC2.

**Depends on:** merge del PR con T1–T8.

**Locations:** el repositorio completo; `_support/docs/history/history-and-provenance.md` (SHAs propios).

**Execution:** hilo principal, porque reemplaza `development` a la fuerza y ese efecto no se delega.

**Test approach:** check.

**Changes:** los pasos 0–7 de [design.md](design.md#enfoque). El paso 0 garantiza que la copia incluye el merge del PR.

**Verification:**
- En la copia reescrita, `git log --all -p | grep -ciE -f <patrón>` y `git log --all --format=%B | grep -ciE -f <patrón>` dan 0, y `git grep -ciE -f <patrón> HEAD` no lista archivos.
- Cada SHA propio de `history-and-provenance.md` existe (`git cat-file -e`).
- La suite completa pasa sobre la copia.
- Después del push, `git ls-remote origin development` coincide con el `HEAD` de la copia.

### T10 — `master`, `v0.1.0` y el Release

- [x] Existen en GitHub la rama `master` (igual a `development`), la etiqueta anotada `v0.1.0` y el Release con 3 `.tar.gz` y 3 `.sha256`.

**Closes:** AC14.

**Depends on:** T9.

**Execution:** hilo principal, porque son efectos hacia fuera.

**Test approach:** check.

**Changes:**
1. `git push origin development:master`.
2. `git tag -a v0.1.0 origin/master -m 'Hive 0.1.0'` y `git push origin v0.1.0`.
3. Copia limpia de `v0.1.0` en el scratchpad y `go run ./tooling/package --out <scratch>/dist` desde ella.
4. `gh release create v0.1.0 --verify-tag --notes-file <entrada 0.1.0 de CHANGELOG.md>` con `dist/versions/0.1.0/*/*.tar.gz` y `dist/versions/0.1.0/*/*.tar.gz.sha256`.

**Verification:**
- `gh release view v0.1.0 --json assets` lista 6 archivos.
- `gh release download v0.1.0 -p '*darwin-arm64.tar.gz*'` y `shasum -a 256 -c` lo acepta.
- `./install.sh --dry-run` desde el paquete extraído termina con estado 0.
- `git ls-remote origin master` coincide con `development`.

### T11 — Refrescar la instalación local

- [x] La instalación local usa la versión 0.1.0.

**Depends on:** T10.

**Execution:** hilo principal.

**Changes:**
1. `go build -ldflags "-X tricell-hive/tooling/version.Current=0.1.0" -o "$(command -v hive)" ./tooling/cli`.
2. `hive update`, que pide confirmación.

**Verification:**
- `hive --version` muestra 0.1.0.
- `hive status --hosts claude,codex,cursor,grok,opencode,pi --scope user` muestra los recursos instalados y verificados en el commit desplegado.

## Verificación compartida

| Prueba | Cuándo | Mecanismo y evidencia | Autorización |
| --- | --- | --- | --- |
| Suite local | sobre la rama antes de la pausa y sobre la copia de T9 | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` sin fallos | incluida en la implementación |
| AC1 sobre el registro | antes del commit de cierre de `flow-close` | `git grep -ciE -f <patrón>` sobre los archivos del cierre no lista nada | incluida en el cierre |
| Validación humana | pausa antes del push | el usuario revisa en VS Code `README.md`, `llms.txt`, `AGENTS.md` y `.github/assets/`, y decide la edición de #5 | D9-A |
| Revisión de código | PR a `development` | `/code-review` de Claude Code; cada hallazgo bloqueante se resuelve o se refuta con evidencia | D10-A |
| Prueba del paquete | T10 | descarga, suma verificada y `install.sh --dry-run` | D9-A |

No hay interfaz gráfica, así que no aplican `hive-review-ux` ni la prueba en navegador. Gatekeeper con un paquete descargado por navegador queda sin verificar: T10 descarga con la línea de comandos.

## Estado de la revisión y avance

- **Ronda 1** sobre la revisión `e36a9e30bddc`, con dos `hive-review-plan` en paralelo (entrega y paquete; documentación y lectura por IA).
  - **Entrega:** 4 bloqueantes y 4 no bloqueantes, todos aceptados.
    - Reglas de reemplazo sin compuestos ni mayúsculas, y tres clientes más en el historial.
    - La etiqueta no se subía.
    - `go build` sin versión.
    - Volver a clonar perdía material ignorado.
    - El mensaje de `install.go:520`.
    - La prueba de `install.sh` sin red.
    - La ruta personal en el historial.
    - Detalles de construcción y subida.
  - **Documentación:** 6 bloqueantes y 4 no bloqueantes.
    - Aceptados: el límite de `global.md`, este registro con identificadores, verificaciones de AC4, AC5, AC6 y AC13, la lista de T8, los huecos por host, la sección `Optional`.
    - Decididos por el usuario: la regla de refresco (D11-A) y la clave de tickets (D12-A).
    - Descartado: el nombre del usuario en `voice_test.go:160` y `LICENSE`, porque es suyo y no de un cliente.
- **Ronda 2** sobre `e2b594863530`, con los mismos dos revisores:
  - Los hallazgos de la ronda 1 quedan corregidos, según las simulaciones de los revisores.
  - Nuevos, aplicados tal como los propusieron los revisores, sin otra ronda:
    - la regla para la clave de tickets en minúsculas, en `<ws>`;
    - el paso 0 de T9 (`fetch` y comparar con `origin` antes de copiar, para no perder el merge del PR);
    - `git status --porcelain` antes de `reset --hard`;
    - los patrones de archivos de T10;
    - dos archivos más en T8 y su excepción marcada;
    - comprobaciones de AC5 con cadenas nuevas;
    - el aviso de Gatekeeper presentado como posibilidad.
  - **Queda para el usuario:** la condición de D11-A no se puede comprobar (`hive status` no registra el checkout). La propuesta del archivo marcador `_support/workspace/maintainer.local` está pendiente de confirmar y bloquea solo esa parte de T3.
  - **Límite:** las reglas de reemplazo se simularon sobre el texto de `git log`, no con `git-filter-repo`, que no está instalado. El paso 2 de T9 repite el conteo sobre la copia reescrita real.
- **Avance:** build iniciado el 2026-10-05 sobre `47a9032`. T1–T8 verificadas. T1 se cerró con su propia comprobación: los 4 PNG existen, `social-preview.png` mide 1280×640 y el emblema es simétrico respecto a su reflejo (diferencia de 0.00007). La comprobación del logotipo completo (0.024) no aplica, porque las letras no son simétricas. Commits locales: `4a988d9`, `d9381eb`, `6d4bd03`. El usuario validó el 2026-10-05: se borraron los conceptos (V1) y se editó el texto de #5 (V2); el texto original está en el scratchpad de la sesión. Push y [PR #7](https://github.com/JhonHawk/tricell-hive/pull/7); `/code-review` en curso. Se instaló `git-filter-repo` (a40bce548d2c) y se ensayó en una copia desechable: de 402 líneas de cambios y 86 de mensajes con coincidencias se pasa a 0, y `HEAD` queda sin coincidencias (940 commits). `/code-review` dejó 9 hallazgos. Se corrigieron F1, F4, F6 y F8 (commit `f16d1b9`, antes de la reescritura). F2, F5 y F7 no se cambian por diseño, y F3 y F9 se proponen como tickets. Merge del PR #7 (`15730dc`) y reescritura: `development` pasó de `15730dc` a `65cc760` (forzado, con protección), con 942 commits y 0 coincidencias. Los SHAs propios de la historia se mapearon (13 resuelven; los 7 del repositorio anterior siguen sin resolver, como antes). Respaldo en `<ws>/pre-rewrite.bundle`. Checkout local reiniciado y limpiado con gc. T10: `master` = `65cc760`, etiqueta `v0.1.0`, [Release](https://github.com/JhonHawk/tricell-hive/releases/tag/v0.1.0) con 6 archivos; el paquete de macOS descargado verifica su suma y `install.sh --dry-run` termina con 0. T11: `hive` 0.1.0 recompilado y `hive update` aplicado (commit `65cc760`); `status` muestra 450 recursos instalados y verificados en 0.1.0. Pendiente de decidir: #1 también contiene identificadores de clientes, y V2 solo autorizó #5. Tras verificar T7 se agregó el delta de requisitos `specs/versioned-installation/spec.md` (se retira la entrada en línea y se agrega la instalación desde el paquete de Releases) y se corrigieron comentarios desactualizados (hallazgos H1–H5 del verificador).
- **Suite local (2026-10-05, rama con T2–T8):** `go vet ./...`, `go test ./...` (los 18 paquetes pasan) y los 23 tests de `tests/skills`: todo en verde. T4 quedó verificada (AC4, AC5, AC6 y la parte de AC13 que toca al README) y se corrigieron sus hallazgos H1–H4: el enlace `#hosts`, la recompilación sin depender del PATH, los perfiles de OpenCode y la frase sobre pi-subagents. La suite se repite si T4 o T1 cambian código; si solo cambian documentos o imágenes, no.
- **Hallazgo durante T5:** con Pi, la instalación desde un paquete ejecuta `pi install npm:pi-subagents@0.74.0` (`integrations/pi/package_settings.go:11`, `tooling/management/pi_package.go:28-35`). Por lo tanto, la frase "un paquete no usa la red" de `README.md:77` y `installer.md:62-66` es falsa. `SECURITY.md` ya está corregido; README e `installer.md` se corrigen en T4 y T7.
- **Para la publicación:** activar el reporte privado de vulnerabilidades cuando el repositorio sea público, porque `SECURITY.md` remite a él. La API devuelve 404 mientras el repositorio es privado.
- **Modelos:** hilo principal Opus 5.5 (observado). `hive-write-spec` y `hive-build-backend` usan sonnet y `hive-verify-task` usa opus (configurados, no observados). Las tareas del hilo principal (T2, T3, T7) se verifican con sonnet, elegido en el lanzamiento; las de los hijos, con el opus configurado.
