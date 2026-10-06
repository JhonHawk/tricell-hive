# Tareas

Commit base: `0bef94fa2d3a807f967b76cf68630ae7ffbd1724` (rama `feat/gh-11-one-line-installer`, 2026-10-05).

Orden recomendado: T1 → T2 → entrega a `development` → T3 → T4 → T5. T4 puede hacerse antes que T3, porque mientras el repositorio sea privado la redirección solo lleva a un 404.

### T1 — Script `get-hive.sh` y sus pruebas

- [x] `get-hive.sh` en la raíz cumple el contrato de [design.md](design.md#contrato-de-get-hivesh) y lo prueban casos automáticos.

**Closes:** AC1, AC2, AC3, AC4, AC5.

**Depends on:** ninguna.

**Locations:** `get-hive.sh` (nuevo); `tooling/distribution/get_hive_test.go` (nuevo), junto a `install_offline_test.go`, que ya ejecuta scripts de shell desde Go. Referencias: `git show 6d4bd037^:bootstrap.sh` para la validación de versión, los límites, el checksum y la comprobación de la terminal, y `git show 'c753e394^:tooling/distribution/bootstrap_test.go'` para el patrón de copia de prueba.

**Execution:** delegada a `hive-build-infra`. Es una unidad con contrato cerrado y sin escrituras compartidas con T2. El encargo incluye este plan, el diseño, [security boundaries](../../../../content/skills/flow-build/references/security-boundaries.md) y `CONTRIBUTING.md`. Puede hacer commits locales en la rama `feat/gh-11-one-line-installer`; no hace push.

**Test approach:** tdd.

**Changes:**
1. Pruebas primero, con un servidor `httptest` en loopback que sirve `/releases/latest` (redirige a `/releases/tag/v9.9.9`), `/releases/download/v9.9.9/<archivo>` y un paquete falso generado en la prueba, cuyo `install.sh` registra sus argumentos y la línea que lee de su entrada. Cada caso cuenta las peticiones que recibe el servidor y corre con `sh` y con `dash`. Casos:
   - camino completo: argumentos reenviados, entrada desde el dispositivo de terminal de la prueba, descargas temporales borradas y `packages/` con solo `<etiqueta>`, sin carpetas de preparación ni anteriores ocultas (AC1);
   - `--dry-run` sin terminal (AC2);
   - sin terminal ni `--dry-run`, con el dispositivo en un directorio inexistente y cero peticiones (AC3);
   - integridad (AC4). En todos estos casos `install.sh` no corre y una carpeta previa del mismo paquete queda igual, comparando su contenido:
     - checksum distinto;
     - `.sha256` con dos líneas, con otro nombre o con caracteres no hexadecimales;
     - etiqueta inválida o de otro repositorio;
     - redirección a un dominio no permitido, incluidas una con el dominio permitido en la ruta (`http://localhost:<puerto>/127.0.0.1/...`) y otra con `@`;
     - descarga mayor que el límite;
     - archivo con entradas `../x`, con otra carpeta raíz o con un enlace simbólico;
   - reemplazo de un paquete existente con un `mv` falso en el `PATH` que falla en el segundo paso: la carpeta anterior queda restaurada (AC4);
   - `--version` sin petición a `/releases/latest`;
   - plataforma no soportada, con un `uname` falso en el `PATH`;
   - copias truncadas del script en cada byte de las últimas dos líneas y en puntos del cuerpo, con cero peticiones y sin archivos nuevos en el `HOME` de la prueba (AC5);
   - cada línea de sustitución aparece exactamente una vez y con su texto exacto esperado en el script publicado.
2. Escribir `get-hive.sh` hasta que pasen.
3. Comprobar el supuesto de redirección de GitHub del diseño con `gh` sobre v0.1.0 y ajustar `allowed_hosts` si difiere.

**Verification:**
- `go test ./tooling/distribution -run GetHive -count=1` pasa.
- `shellcheck -s sh get-hive.sh` no reporta nada.
- Prueba local con los paquetes reales: `gh release download v0.1.0` en una carpeta temporal, servida por un servidor en loopback con la misma estructura de rutas que GitHub. La copia de prueba del script, con `--version 0.1.0 --dry-run`, termina en la vista previa real de `install.sh` bajo un `HOME` temporal. Así se comprueban el `tar` de macOS y el `install.sh` real del paquete sin depender de la visibilidad del repositorio.

### T2 — Documentación y requisitos

- [x] La documentación describe el instalador de una línea y el delta de requisitos refleja el contrato final de T1.

**Closes:** AC6.

**Depends on:** T1, porque los textos citan el contrato que quede implementado.

**Locations:** `_support/docs/architecture/installer.md` (sección «Installing from a GitHub Release» y el párrafo «There is no one-line online installer»); `README.md` (sección «From a release package» e «Update»); `llms.txt`; `CHANGELOG.md` (sección sin publicar); `_support/openspec/changes/gh-11-one-line-installer/specs/versioned-installation/spec.md`.

**Execution:** hilo principal. Son textos cortos que dependen de lo que T1 entregue; el encargo sería más largo que el cambio.

**Test approach:** check, con `rg` y la suite de contenido.

**Changes:**
- `installer.md`:
  - el comando de una línea, con `--dry-run` y `--version`;
  - los requisitos (`curl`, `tar`, `shasum` o `sha256sum`);
  - la carpeta donde queda el paquete;
  - la lista de dominios permitidos;
  - que el checksum no es una firma;
  - la ruta manual como alternativa.
- README: el comando de una línea primero y la ruta manual después; en «Update», que basta con volver a ejecutarlo.
- `llms.txt`: el comando y `--dry-run` para agentes.
- CHANGELOG: entrada «Added» con referencia a #11.

**Verification:**
- `rg -n 'There is no one-line online' _support/docs/architecture/installer.md` no encuentra nada.
- `rg -n 'hive.tricell.tech/install.sh' README.md llms.txt _support/docs/architecture/installer.md` encuentra los tres archivos.
- `rg -n -e '--dry-run' -e '--version' -e 'hive/packages' _support/docs/architecture/installer.md README.md` encuentra los tres términos en ambos archivos.
- `go test ./tests/content -count=1` pasa.
- Ninguna prueba revisa los enlaces del README ni de `llms.txt`. Los enlaces nuevos se comprueban a mano: cada enlace relativo apunta a un archivo existente.

### T3 — Release con el instalador y promoción a `master` (requiere orden)

- [x] `master` contiene `get-hive.sh` como parte de un Release publicado (D7-B).

**Depends on:** merge de T1 y T2 en `development`.

**Execution:** hilo principal, con orden explícita del usuario. No está cubierta por D4-A. La orden fija el número de versión. Recomiendo `0.2.0`, porque SemVer pide subir la versión menor cuando se agrega funcionalidad; `0.1.1` también es posible si se prefiere tratarlo como un ajuste.

**Changes:** el mismo procedimiento que publicó 0.1.0 (gh-6, T10), con `<v>` como la versión acordada:
1. PR de preparación a `development`, sin revisión dedicada porque solo cambia texto, que yo fusiono:
   - `VERSION` pasa a `<v>`;
   - la sección `[Unreleased]` del CHANGELOG pasa a `[<v>] - <fecha>`, con la entrada «Added» de T2 y la entrada «Changed» ya pendiente;
   - queda un `## [Unreleased]` vacío encima;
   - se agrega el enlace de comparación de `<v>` si el archivo los usa.

   Anoto el SHA de ese merge como `<sha>`.
2. `git push origin <sha>:master`, que no es `development:master`, para que un merge posterior no entre en el Release. Hoy es un avance sin fusión, porque `master` es ancestro de `development`. Antes de etiquetar, `git rev-parse origin/master` debe dar `<sha>`.
3. `git tag -a v<v> <sha> -m 'Hive <v>'` y `git push origin v<v>`.
4. En una copia limpia de `v<v>` en el scratchpad, `go run ./tooling/package --out <scratch>/dist`.
5. `gh release create v<v> --verify-tag --notes-file <entrada <v> del CHANGELOG>`, con los `.tar.gz` y `.tar.gz.sha256` de `dist/versions/<v>/*/`.
6. Despliegue del Release en el checkout del mantenedor, desde la etiqueta y no desde `development`. Primero `go build -ldflags "-X tricell-hive/tooling/version.Current=<v>" -o "$(command -v hive)" ./tooling/cli` en la copia de `v<v>`. Después, `hive update --rev v<v>` en el checkout principal, que despliega el contenido de la etiqueta aunque `development` avance; `--rev` existe según `deployment-manager.md`. Así se hizo en T11 de gh-6, solo que entonces `development` coincidía con la etiqueta. Ese es el despliegue que cuenta como el del Release.
7. Enseguida, sin otros merges entre los pasos 1 y 7, un PR a `development` que vuelve a poner `VERSION=dev`, según la convención del CHANGELOG, que yo fusiono.

**Verification:**
- `git show origin/master:get-hive.sh | head -1` imprime `#!/bin/sh`.
- `git show origin/master:VERSION` da `<v>`.
- `gh release view v<v> --json assets -q '.assets | length'` da `6`.
- Contra el Release real con `gh release download` y la copia de prueba de T1, `--version <v> --dry-run` termina en la vista previa de `install.sh`.
- `hive --version` muestra `<v>`, y `hive status` muestra los recursos instalados y verificados en `<v>`.
- `git show origin/development:VERSION` da `dev` después del paso 7.

### T4 — Cloudflare para `hive.tricell.tech`

- [x] `hive.tricell.tech/install.sh` redirige al script en `master`.

**Closes:** AC7.

**Depends on:** ninguna técnica. El destino solo responde cuando T3 y T5 están hechas.

**Execution:** el usuario, en el panel de Cloudflare, porque aquí no hay token de Cloudflare ni `wrangler`. También puedo hacerlo yo por la API si el usuario guarda un token con alcance de DNS y reglas de la zona en `_support/secrets/`, con modo `0600`.

**Changes:** el registro DNS y la Single Redirect de [design.md](design.md#cloudflare-d2-a).

**Verification:** `dig +short hive.tricell.tech` devuelve direcciones de Cloudflare, y `curl -sI https://hive.tricell.tech/install.sh` muestra `302` y el `Location` de AC7.

### T5 — Publicar el repositorio y prueba en vivo (requiere orden)

- [x] El repositorio es público y el comando de una línea funciona sin credenciales.

**Closes:** AC8.

**Depends on:** T3 y T4.

**Execution:** hilo principal, con orden explícita del usuario en el momento (D1-A decide publicarlo, pero no autoriza el cambio de visibilidad por sí sola).

**Changes:**
1. Revisión previa a publicar, sin mostrar valores:
   - `gitleaks git --log-opts='--all' --redact --no-banner` sin hallazgos confirmados;
   - `git ls-files -ci --exclude-standard` vacío, es decir, ningún archivo ignorado está versionado;
   - el patrón de identificadores de clientes de gh-6 (`_support/workspace/2026-10-05-gh-6-public-readiness/client-names.pattern`) da 0 coincidencias en el historial de `development` y `master`, las issues y sus comentarios, los PR #7 y #8 con sus revisiones, y las notas del Release;
   - la wiki del repositorio se revisa o se desactiva.
2. Con la orden: `gh repo edit JhonHawk/tricell-hive --visibility public --accept-visibility-change-consequences`.
3. Activar el reporte privado de vulnerabilidades al que remite `SECURITY.md`, con `gh api -X PUT repos/JhonHawk/tricell-hive/private-vulnerability-reporting`. Lo cubre la misma orden de publicación.
4. Prueba anónima. `raw.githubusercontent.com` guarda en caché unos 5 minutos (`max-age=300`), así que si falla justo después de promover se repite pasado ese intervalo.

**Verification:**
- `curl -s -o /dev/null -w '%{http_code}' https://github.com/JhonHawk/tricell-hive` da `200`.
- `gh api repos/JhonHawk/tricell-hive/private-vulnerability-reporting -q .enabled` da `true`.
- En un shell sin `gh` ni credenciales, `curl -fsSL https://hive.tricell.tech/install.sh | sh -s -- --dry-run` muestra la vista previa del instalador.
- Recorrido del usuario, opcional: `curl -fsSL https://hive.tricell.tech/install.sh | sh -s -- --home "$(mktemp -d)" --state-dir "$(mktemp -d)"` en una terminal, que pregunta hosts y confirmación sin tocar la instalación real.

## Verificación compartida

| Control | Momento | Mecanismo y evidencia esperada | Autorización |
| --- | --- | --- | --- |
| Suite local | sobre el candidato final, antes del PR | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasan | incluida en la implementación |
| `/code-review` | sobre el PR, antes del merge | hallazgos corregidos o refutados con evidencia | D5-A |
| Prueba en vivo anónima | después de T3, T4 y T5 | el comando de una línea con `--dry-run` desde un shell sin credenciales | incluida en T5, que requiere orden |

No hay interfaz gráfica: no aplican `hive-review-ux` ni la revisión visual. `hive-verify-change` no se planea, porque la prueba en vivo de T5 es la verificación de punta a punta y depende de efectos externos que requieren orden.

## Revisión del plan y progreso

- Revisión del plan, 2026-10-05, con dos revisores `hive-review-plan` delegados como hijos nativos de solo lectura, sobre la revisión `1b8a107e8789`:
  - Seguridad del script (T1, AC1–AC5, delta de requisitos). Hallazgos bloqueantes:
    - B1: la lista de dominios se podía saltar comparando la URL con un patrón; ahora se extrae y compara el host.
    - B2: un corte justo después de `main` ejecutaba el script sin argumentos; ahora la última línea es `{ main "$@" || exit 1; }`.

    Se incorporaron también N1–N5: comprobar los valores reales del script publicado, la prueba sin terminal, el reemplazo seguro de una carpeta existente, el rechazo de enlaces y el aislamiento de `curl` y `tar`. Son propuestas del propio revisor, así que no hubo segunda ronda.
  - Infraestructura y entrega (T2–T5, AC6–AC8, entrega). Hallazgos bloqueantes:
    - B1: el momento del cierre; ahora se cierra después de T5.
    - B2: la recomendación de T3 chocaba con el alcance; el usuario eligió D7-B.

    Se incorporaron también N1–N5: la revisión previa a publicar, el reporte privado de vulnerabilidades, la caché de `raw`, la verificación de T2 y la nota sobre `hive update`.
  - Segunda y última ronda del mismo revisor de infraestructura sobre el nuevo T3: sin bloqueantes. Se incorporaron sus propuestas:
    - N6: fijar `<sha>` al promover;
    - N7: desplegar desde la etiqueta;
    - N8: la forma del CHANGELOG.

    El orquestador ajustó el paso 6 para usar `hive update --rev v<v>`, que existe según `deployment-manager.md`, en lugar de `--source` con una copia que no es checkout de Git.
- Límites de la revisión:
  - no se ejecutó nada;
  - quedan como supuestos para T1:
    - el comportamiento de bsdtar y GNU tar con enlaces;
    - `--max-filesize` sin longitud declarada;
    - la redirección real de GitHub;
  - no se pudo revisar el panel de Cloudflare.
- Progreso: plan redactado y revisado el 2026-10-05; #11 abierto. El usuario ordenó `flow-build` para T1 y T2 el 2026-10-05. Build iniciado sobre `0bef94fa` en `feat/gh-11-one-line-installer`; #11 asignado al usuario como señal de trabajo en curso, porque GitHub Issues no tiene un estado intermedio. T2 se redacta en paralelo sobre el contrato del diseño y se ajusta al resultado de T1.
- Modelos:
  - T1: implementador `hive-build-infra`, modelo `sonnet` (configurado, no observado); verificador `hive-verify-task`, modelo `opus` (configurado, no observado).
  - T2: implementador en el hilo principal, Opus 5.5 (observado); el verificador se lanza con `sonnet` para que no comparta modelo con el implementador.
- T1, primera entrega (`4c1658ea`), 2026-10-05:
  - RED observado: 24 de 24 pruebas fallan porque `get-hive.sh` no existe; el script se había redactado antes y se apartó para observar ese fallo.
  - GREEN y estabilidad: 12 roturas deliberadas del script hacen fallar sus pruebas; `shellcheck`, `go vet` y `gofmt` limpios.
  - Prueba con el paquete real de v0.1.0: llega a la vista previa con `--hosts claude`.
  - La redirección de GitHub va a `release-assets.githubusercontent.com`, que la lista de dominios ya admite.
  - Hallazgo: con el `install.sh` real, `--dry-run` sin `--hosts` necesita terminal para elegir los hosts. El orquestador decidió D1-A: con `--dry-run`, usar la terminal si abre. AC2, AC8, el diseño, el delta de requisitos y los documentos se ajustaron; el implementador aplica el cambio con sus pruebas.
  - Límites: GNU tar sin probar en Linux; `ENXIO` real sin probar; la petición real a `releases/latest` queda para T5.
- D1-A aplicado en `f42dd69f`. RED observado: la entrada de `install.sh` no venía de la terminal, tanto con archivo como con tubería. GREEN: `go test ./tooling/distribution` pasa.
- Verificación independiente, 2026-10-05:
  - T1, con `hive-verify-task` en `opus` (configurado): AC1 a AC5 cumplidos, citando `get_hive_test.go:549`, `:629` y `:1360`, que confirmé. Marcada `[x]`.
  - T1, hallazgos menores:
    - H1: un Mac Intel termina con un 404 en vez de un mensaje claro.
    - H2: ninguna prueba cubre la comparación por sufijo de dominio.

    Se corrigen en esta entrega con el mismo implementador.
  - T2, con `hive-verify-task` en `sonnet` (pasado al lanzarlo): AC6 cumplido. Se aplicaron sus precisiones: lista completa de comandos, `--help`, `XDG_DATA_HOME` relativo y la redacción de `llms.txt`. Commit `0eb1c7f7`. Marcada `[x]`.
- Suite local completa sobre el candidato con T1 y T2: pasa (`go vet`, 17 paquetes de Go y 23 pruebas de skills).
- H1 y H2 corregidos en `ca341d5a`: Mac Intel rechazado antes de cualquier petición y prueba del sufijo de dominio con los valores publicados.
- PR #12 y `/code-review` (nivel alto): 10 hallazgos, todos aceptados y corregidos.
  - En el script, `af287ea2`, con RED y GREEN observados:
    - F1 y F5: ventanas de señales durante el reemplazo de un paquete existente;
    - F2: las formas de escribir los flags que acepta Go;
    - F4: Rosetta;
    - F6: la prueba estructural;
    - F9: un único ayudante de `curl`.
  - En documentación, `537a6025`:
    - F3: `--dry-run` sí descarga y guarda el paquete;
    - F7: el orden del reemplazo;
    - F8: solo se revisa el host final, decisión de documentarlo en vez de revisar cada salto;
    - F10: las carpetas de versiones anteriores se acumulan.

  Después volví a correr `go test ./tooling/distribution ./tests/content` y `shellcheck`, y pasan. No hubo una segunda revisión de código.
- Merge del PR #12 en `8154d5bc`; #11 se cerró. `hive update` registró el commit sin cambios de archivos. La rama se borró en local y en el remoto.
- Pausa: T3–T5 esperaban orden del usuario.
- T3, orden del usuario del 2026-10-05 (D8-A, versión 0.2.0):
  - PR #13 de preparación, con merge en `3d882f87`. Además de `VERSION` y el CHANGELOG, actualiza la nota de versión del README y añade 0.2.x a la tabla de `SECURITY.md`; 0.1.x sigue soportada.
  - `master` avanzó sin fusión hasta `3d882f87`.
  - Etiqueta `v0.2.0` sobre `3d882f87`.
  - Paquetes compilados desde una copia limpia de la etiqueta.
  - [Release v0.2.0](https://github.com/JhonHawk/tricell-hive/releases/tag/v0.2.0): 6 archivos; es el último según `releases/latest` de la API.
  - Prueba con el paquete real de `darwin-arm64`: `shasum -c` lo acepta, y la copia de prueba de `get-hive.sh` sobre un servidor local, con `--version 0.2.0 --dry-run --hosts claude`, sale con 0 y llega a la vista previa.
  - `hive` recompilado como 0.2.0 y `hive update --rev v0.2.0` aplicado con `--out` y `apply`: 450 recursos instalados y verificados en 0.2.0.
  - PR #14 devolvió `VERSION=dev`, con merge en `7bdbe10e`.
  - No se aplicó `hive update` después de #14: no cambia ningún archivo, y solo reetiquetaría la instalación de 0.2.0 a `dev`. El despliegue del Release es el de la etiqueta, según N7.
  - Ramas borradas en local y en el remoto; la copia temporal se eliminó.
- Siguiente: T4 (Cloudflare, el usuario) y T5 (orden de publicar).
- T4, parcial, por pedido del usuario (2026-10-05, «el registro lo puedes agregar tú»), con la API de Cloudflare y el `CLOUDFLARE_API_TOKEN` del entorno, sin mostrar su valor:
  - Creado `AAAA hive.tricell.tech → 100::` con el proxy activado, id `965fc71691bf77707d48789c9114e802`. Para revertirlo: `DELETE /zones/1af51a81614ee067e428b2af0fa169bb/dns_records/965fc71691bf77707d48789c9114e802`.
  - La regla de redirección quedó bloqueada: el token no tiene permiso de reglas (`rulesets` devuelve el error 10000; `pagerules` y Workers no autorizados). La documentación de Cloudflare (Context7, `/cloudflare/cloudflare-docs`) pide el permiso «Zone > Single Redirect > Edit».
  - La zona no tenía reglas de redirección ni otro registro `hive`.
- T5, revisión previa (solo lectura, 2026-10-05). Todo sin hallazgos:
  - `gitleaks` sobre todo el historial: 0 hallazgos;
  - `git ls-files -ci --exclude-standard`: vacío;
  - patrón de clientes de gh-6: 0 coincidencias en el historial de todas las ramas (contenido, autores y mensajes), en los archivos versionados de `HEAD`, en las issues y sus comentarios, en los comentarios y revisiones de los PR y en las notas de los Releases.

  La wiki está activada pero vacía: su repositorio no existe. Se propone desactivarla con la publicación.
- T5, publicación por orden del usuario (D10-A, 2026-10-05):
  - Repositorio público, wiki desactivada y reporte privado de vulnerabilidades activado (`enabled=true`).
  - Sin credenciales: el repositorio y `raw/master/get-hive.sh` responden 200, y `releases/latest` redirige a `v0.2.0`.
  - Prueba anónima de punta a punta con `curl -fsSL <raw>/get-hive.sh | env -i … sh -s -- --dry-run --hosts claude`, con `HOME` y `TMPDIR` temporales: sale con 0, resuelve 0.2.0, descarga y verifica, deja el paquete y llega a la vista previa; `TMPDIR` queda vacío. Así quedan comprobadas en vivo la consulta HEAD de la última versión y la redirección de descargas.
  - AC8 quedaba pendiente solo de la URL `hive.tricell.tech`, es decir, de la regla de T4 (D9-A: el usuario amplía el permiso del token).
- T4, completada el 2026-10-05:
  - El usuario agregó «Zona > Redirección única > Editar» al mismo token.
  - Creé el conjunto de reglas `http_request_dynamic_redirect` (id `5c0037cfa02e47868ce6b6ca916eb77a`) con una sola regla, `hive_one_line_installer` (id `eda2accba0734d96b62d9493790d4f6d`): `302` hacia `raw/master/get-hive.sh`, sin conservar la query string.
  - Para revertir: `DELETE /zones/1af51a81614ee067e428b2af0fa169bb/rulesets/5c0037cfa02e47868ce6b6ca916eb77a/rules/eda2accba0734d96b62d9493790d4f6d` y borrar el registro DNS indicado arriba.
  - AC7: `curl -sI https://hive.tricell.tech/install.sh` muestra `HTTP/2 302`, el `location` esperado y `server: cloudflare`. Cualquier otra ruta de `hive.tricell.tech` responde 522, como se esperaba: el nombre no tiene origen.
- T5, prueba final:
  - AC8: `curl -fsSL https://hive.tricell.tech/install.sh | env -i … sh -s -- --dry-run --hosts claude`, sin credenciales y con `HOME` temporal, sale con 0, resuelve 0.2.0 y llega a la vista previa.
- Cierre (2026-10-05): el delta `ADDED: One-line online installation` se aplicó a `_support/openspec/specs/versioned-installation/spec.md`. El registro se archivó y se versiona con un PR pequeño a `development`.
- Siguiente paso: `flow-build` con esta carpeta, empezando por T1.
