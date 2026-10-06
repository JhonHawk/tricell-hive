# Diseño

## Contexto verificado (2026-10-05)

- **Visibilidad.** `gh repo view` reporta `JhonHawk/tricell-hive` como `PRIVATE`. Sin credenciales, el repositorio, `releases/download/v0.1.0/...sha256` y `raw.githubusercontent.com/.../master/install.sh` responden 404.
- **Dominio.** `tricell.tech` usa los servidores de nombres `vin` y `april.ns.cloudflare.com`. No tiene registros A, AAAA ni CNAME en la raíz, `www`, `hive` ni `get`. No hay `wrangler` ni token de Cloudflare en esta máquina.
- **Release v0.1.0.** Tiene tres paquetes, `darwin-arm64` (7,9 MB), `linux-amd64` (8,2 MB) y `linux-arm64` (7,8 MB), cada uno con su `.sha256`. El `.sha256` es una sola línea, `<64 hex>  hive-0.1.0-darwin-arm64.tar.gz`. El archivo tiene una única carpeta raíz, `hive-0.1.0-darwin-arm64/`, con 132 entradas.
- **Paquete.** [tooling/package/main.go](../../../../tooling/package/main.go) copia una lista fija (`content`, `integrations/agent-profiles.json`, `install.sh`, `LICENSE`, `THIRD_PARTY_NOTICES.md`) y nombra el archivo con `archiveLabel` (`hive-<versión>-<os>-<arch>`).
- **Instalador del paquete.** [install.sh](../../../../install.sh) verifica la plataforma y el checksum del binario, y ejecuta `bin/hive install "$@" --source "$package_dir"`. El instalador solo aplica cambios con confirmación en una terminal; `--dry-run` no la necesita ([installer.md](../../../docs/architecture/installer.md#interaction)). El binario no localiza su propio paquete: el README pide conservarlo para la recuperación y para los comandos posteriores.
- **Precedente.** El `bootstrap.sh` retirado en 0.1.0 (`git show 6d4bd037^:bootstrap.sh`) validaba versiones, limitaba tamaños, verificaba checksums y comprobaba la terminal abriéndola con `( exec 3<>/dev/tty )`. Su comentario explica por qué `[ -r /dev/tty ]` no sirve. Descargaba un binario que después ejecutaba `hive bootstrap`; ese subcomando se retiró en `c753e394`. Sus pruebas sustituían las líneas de origen y protocolo en una copia del script usada solo por las pruebas, sin variables de entorno que alteren el script publicado.
- **Herramientas.** `gitleaks`, `shellcheck` y `dash` están instalados.
- **Documentación de Cloudflare** (Context7, `/cloudflare/cloudflare-docs`, consultada el 2026-10-05). Las Single Redirects requieren que el tráfico del nombre pase por el proxy de Cloudflare. Para un nombre sin servidor de origen se usa un registro de relleno con proxy, como `AAAA 100::`.
- **Supuesto por confirmar en T1.** Las descargas de Release de GitHub redirigen a un dominio bajo `githubusercontent.com` (hoy `release-assets.githubusercontent.com`). T1 lo comprueba contra v0.1.0 con `gh` mientras el repositorio siga privado, y contra la URL anónima en T5.

## Contrato de `get-hive.sh`

Uso:

```sh
curl -fsSL https://hive.tricell.tech/install.sh | sh                       # última versión
curl -fsSL https://hive.tricell.tech/install.sh | sh -s -- --dry-run       # solo vista previa
curl -fsSL https://hive.tricell.tech/install.sh | sh -s -- --version 0.1.0 --hosts claude,codex
```

- **Argumentos.** `--version <X.Y.Z[-pre]>` lo consume el script y se valida con la misma gramática que usaba `bootstrap.sh`. `--help` imprime el uso. Todo lo demás pasa sin cambios a `install.sh`. `--dry-run` también pasa, y además decide si hace falta terminal.
- **Requisitos.** `curl`, `tar` con gzip, `shasum` o `sha256sum`, `mktemp`, `uname` y `HOME` definido. Si falta alguno, el script termina con un error que lo nombra y antes de cualquier descarga.
- **Origen.** Es fijo en una sola línea: `repo_url='https://github.com/JhonHawk/tricell-hive'`. Ninguna variable de entorno ni argumento lo cambia. Todas las peticiones usan `curl -q --fail --proto '=https' --proto-redir '=https'`, con tiempo límite y `--max-filesize`, y comprueban el tamaño final con `wc -c`. `-q` va primero para que `~/.curlrc` no se lea; esto evita, por ejemplo, un `--insecure` del usuario. El script vacía `TAR_OPTIONS` antes de usar `tar`.
- **Dominios permitidos tras redirecciones.** De la URL efectiva final (`%{url_effective}`) se extrae el host: se quita el esquema, se corta en el primer `/`, `?` o `#`, se rechaza la URL si contiene `@` y se quita el puerto. Ese host debe ser exactamente `github.com` o terminar en `.githubusercontent.com`. Solo se comprueba el host final: los saltos intermedios deben usar HTTPS (`--proto-redir`), pero su host no se revisa; así quedó documentado tras `/code-review` (F8). No se compara la URL completa con un patrón de `case`, porque el `*` también acepta `/` y dejaría pasar `https://evil.example/x.githubusercontent.com/`. Cualquier otro host es un error.
- **Última versión.** Se pide `"$repo_url/releases/latest"` sin cuerpo y se lee la URL final. Debe ser exactamente `"$repo_url/releases/tag/v<versión>"` con una versión válida; si no, es un error. Así no se usa la API de GitHub ni su límite anónimo de peticiones.
- **Descargas.** `hive-<v>-<os>-<arch>.tar.gz` (límite de 64 MiB) y su `.sha256` (límite de 256 bytes) se descargan desde `"$repo_url/releases/download/v<v>/"`, a una carpeta temporal privada (`umask 077`, `mktemp -d`).
- **Checksum.** El `.sha256` debe ser una sola línea con exactamente `<64 hex minúsculas>  <nombre exacto del archivo>`. El digest calculado debe coincidir.
- **Validación del archivo antes de extraer.** Cada entrada de `tar -tzf` debe ser `<etiqueta>` o empezar por `<etiqueta>/`, sin componentes `..` ni rutas absolutas. Se extrae en una carpeta de preparación creada con `mktemp -d` dentro de `packages/`, para que mover el resultado sea un `mv` en el mismo sistema de archivos. Además, `tar -tvzf` debe mostrar solo carpetas (`d`) y archivos regulares (`-`), como produce `writeArchive` y exige `Extract` en Go. Después de extraer, `find` comprueba que no aparezca ningún otro tipo, y la preparación debe contener solo `<etiqueta>/`, como carpeta real y no enlace, con un `install.sh` regular. Los controles finos (checksum del binario, inventario completo, plataforma) siguen a cargo de `install.sh` y del binario.
- **Ubicación (D3-A).** `packages_dir="${XDG_DATA_HOME:-$HOME/.local/share}/hive/packages"`. Un `XDG_DATA_HOME` relativo se ignora, como pide la especificación XDG. Si ya existe `<etiqueta>/`, primero se renombra a una carpeta oculta temporal, después se mueve el paquete nuevo a su lugar y por último se borra la anterior. Justo antes de mover el nuevo se comprueba que `<etiqueta>` no exista ni sea un enlace (`-e` o `-L`); si existe, es un error, para evitar que `mv` meta el paquete dentro de esa carpeta. Si falla el segundo paso, o llega `INT`, `TERM` o `HUP` entre los dos, la trampa restaura la anterior y termina con `exit`, porque en POSIX una trampa sin `exit` deja que el script siga. Las carpetas de otras versiones no se tocan.
- **Terminal.** Sin `--dry-run`, el script abre la terminal con `( exec 3<>"$tty_device" )` antes de cualquier petición de red, con `tty_device=/dev/tty`. Si falla, termina con un error que remite a `--dry-run` o al paquete manual.
- **Entrega del control.** El script borra explícitamente su carpeta temporal de descargas, la de preparación y la anterior oculta, porque `exec` reemplaza al shell y la trampa `EXIT` no se ejecutaría. Después imprime dónde quedó el paquete y la ruta de `bin/hive`, y ejecuta `exec "$package/install.sh" "$@"`: con siempre con `< "$tty_device"` sin `--dry-run`. Con `--dry-run` también usa `< "$tty_device"` cuando la terminal abre; si no abre, deja la entrada como está. La sonda se hace antes de cualquier petición. Esto se cambió durante el build (D1-A del 2026-10-05): con el `install.sh` real de v0.1.0, `--dry-run` sin `--hosts` necesita una terminal para elegir los hosts, y con `curl | sh` la entrada es la tubería. El código de salida es el del instalador.
- **Errores.** Cada fallo imprime una línea en stderr y sale con un código distinto de 0. Una trampa `EXIT INT TERM HUP` borra la carpeta temporal de descargas y la de preparación, y nunca toca la carpeta final de un paquete que ya existía.
- **Protección contra descargas truncadas.** Todo el script vive en funciones; la última línea es `{ main "$@" || exit 1; }`. No basta con `main "$@"`: una copia cortada justo después de `main` llamaría a la función sin argumentos y perdería `--dry-run`. Con las llaves, cualquier corte antes de `}` deja un bloque sin cerrar que falla al interpretarse, y las líneas anteriores solo definen funciones.
- **Compatibilidad.** El script es POSIX `sh`, sin bashismos. Se prueba con `/bin/sh` y `dash`, y pasa `shellcheck -s sh`.

### Puntos de sustitución para las pruebas

Como hacía `bootstrap.sh`, las pruebas generan una copia del script en la que reemplazan líneas exactas, cada una definida una sola vez:
- `repo_url=`: por un servidor `httptest` en loopback;
- `allowed_hosts`: por `127.0.0.1`;
- la restricción de protocolo, por `http`;
- `tty_device=`: por un archivo de la prueba.

El script publicado no lee ninguna de ellas del entorno. Una prueba comprueba que cada línea aparece exactamente una vez y con su texto exacto esperado en el script publicado: origen `https://github.com/JhonHawk/tricell-hive`, protocolo `=https`, hosts `github.com` y `.githubusercontent.com`, y `/dev/tty`. Así, un valor erróneo en el script publicado no queda oculto por la sustitución.

Límite: la copia de prueba usa un archivo como terminal. La ausencia de terminal se simula con una ruta dentro de un directorio inexistente, porque `exec 3<>` crearía un archivo ausente. El error real de `/dev/tty` sin terminal de control (`ENXIO`) no se prueba automáticamente; el script retirado ya lo había comprobado con `dash`.

## Cloudflare (D2-A)

- **DNS.** `hive.tricell.tech`, `AAAA 100::`, con proxy activado. Es la dirección de descarte que la documentación de Cloudflare usa para nombres sin origen.
- **Redirección.** Una Single Redirect con la expresión `http.host eq "hive.tricell.tech" and http.request.uri.path eq "/install.sh"`, que responde `302` hacia `https://raw.githubusercontent.com/JhonHawk/tricell-hive/master/get-hive.sh` sin conservar la query string. Es `302` y no `301` para que cambiar el destino no quede en cachés. El resto de rutas de `hive.tricell.tech` no tienen regla.
- **Por qué `master`.** `master` es el entorno de publicación (`development → master`). El script que reciben los usuarios es el promovido, no el que está en integración.

## Alternativas descartadas

- **Un Worker o R2 en Cloudflare.** No hacen falta mientras el repositorio sea público y baste con una URL. Si más adelante se quieren URLs por plataforma, la regla se convierte en un Worker sin cambiar la URL pública.
- **La API de GitHub para resolver la última versión.** Limita las peticiones anónimas por IP; la redirección de `releases/latest` no tiene ese límite y no necesita interpretar JSON.
- **Nombres de paquete sin versión.** Permitirían usar `releases/latest/download/<nombre>`, pero obligan a cambiar `tooling/package` y a publicar un Release nuevo.

## Riesgos

- **Un cambio en cómo redirige GitHub** rompería la resolución de la última versión o la lista de dominios permitidos. El error sería visible y no ejecutaría nada; `--version` sigue funcionando sin resolver la última.
- **Supuesto: el `tar` de macOS (bsdtar) y GNU tar aceptan `-tzf` y `-xzf` de la misma forma** para estos archivos gzip. T1 lo prueba en macOS; en Linux queda sin probar hasta la primera instalación real, salvo que se use bb1 (opcional).
