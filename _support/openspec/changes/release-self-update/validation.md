# Validación de la actualización completa de Hive

Estado: candidato listo para validación humana; rama local `feat/release-self-update`.
Código probado: `f6d8cf915590d8f85572a66785bb5c1e696f9e68`, base integrada `a67feac4`.

La migración inicial usa el instalador compatible y deja un comando estable. Después `hive update` actualiza binario e integraciones, sin Git ni Go. Se conservan selección exacta de versión, preview, plan offline, modo source explícito y recuperación según el límite del commit. La publicación de una release no forma parte de esta validación.

## Comprobaciones

| Comprobación | Resultado observado |
| --- | --- |
| `go test ./...` | Todos los paquetes pasan; CLI241.901s, gestión77.647s, viewer62.345s |
| CLI normal tras fixes de revisión | PASS190.606s |
| `go test -race -timeout=18m ./...` | 21/21 paquetes `ok`; CLI 605.134s, gestión 272.560s; sin informes de carrera ni fallos |
| Integridad del resultado race | Conjunto de paquetes igual a `go list ./...`, ninguno ausente; proceso terminado observado mediante kqueue |
| `go vet ./...` | PASS |
| Python | 57 pruebas PASS,2.571s |
| Sitio | Astro check sin errores; build estático de 17 páginas. Warnings existentes de i18n/404 vacíos |
| Linux ARM64 y AMD64 | Compilación offline con CGO=0 PASS; ELF estáticos. No ejecución nativa Linux |
| Checker del change folder y diff | PASS |
| Texto global distribuido | 35954 bytes antes y después; delta0 |

La sesión del ejecutor original se perdió durante el reinicio del daemon. La suite race continuó, no se repitió: se observó la salida de su proceso y se comprobó el resultado de cada uno de sus 21 paquetes. No se conserva el código de salida del wrapper original.

## Criterios y evidencia

| Criterio | Evidencia |
| --- | --- |
| AC1 — Actualización completa sin Git/Go | E2E migra legacy→paquete1.0.0→1.2.3, vacía PATH, ejecuta comando estable y comprueba Claude+Codex |
| AC2 — Versión y compatibilidad | Owners de resolución estable/exacta, downgrade explícito, prerelease y protocolo/esquema |
| AC3 — Entrada remota verificada | Casos de checksum, redirecciones, truncado, traversal, enlaces y límites |
| AC4 — Preview/cancelación sin cambios | Comparación de home/almacén y programa TUI real con Ctrl-C/SIGINT/QuitMsg; espera al worker |
| AC5 — Plan offline congelado | E2E aplica con transporte que falla ante cualquier petición; rechazos de bytes/destinos/estado alterados |
| AC6 — Migración y propiedad | Instalación repetida, PATH sin cambios, colisiones y desinstalación con enlace alterado |
| AC7 — Preservación y proveedores | Owners existentes de drift, recursos compartidos, voces y overrides; release no programa proveedores |
| AC8 — Antes del commit | Fallos inyectados restauran core y conservan el comando anterior |
| AC9 — Después del commit | Journal durable asociado, respuesta perdida, retención frente a pruning y routing previo a esquema 999 |
| AC10 — Concurrencia | Locks, mutadores competidores, ausencia de fugas y lectura con operación pendiente |
| AC11 — Source/dev | Commit excluye ediciones locales, selectors explícitos y cero peticiones de baredev/noTTY |
| AC12 — CLI/TUI | Preparación, confirmación, Apply protegido, éxito con `/hooks`, error/reintento y recuperación; UX aprobó navegación corregida |
| AC13 — Preview del ejecutable final | Seis hosts muestran destino final ausente; apply exige paquete definitivo exacto |
| AC14 — Pi incompleto | Preflight/apply bloquean antes de writes, preservan settings y no invocan Pi |

Los verificadores independientes de tareas confirmaron esos contratos en sus owners. El fix H12, posterior al último verificador, tiene regresión local red→green y forma parte de la suite race final.

## Revisión y límites

- Revisión nativa Codex y un follow-up ejecutados. H9 (builder de paquete al aceptar wait-floor) y H10 (avisos de hooks) corregidos.
- El follow-up encontró H12: un core extendido con el mismo esquema de estado podía ser bloqueado por el cálculo local de notas. Se corrigió con una regresión local; `97c86edc` lleva `Unreviewed` porque no se ejecutó una tercera revisión.
- UX independiente y follow-up aprobaron H11: Shift+Tab retrocede por Mode/Source/Revision y por Mode/Version. Terminal real a 80×24; 40×24 muestra la protección de tamaño mínimo.
- Recuperación probada con interrupciones lógicas y validación de fsync/journals en código; no se simuló pérdida física de energía.
- La preocupación de mensajes extremadamente largos truncados no se reprodujo como defecto. No se midió expansión de copy de 30–50%. No aplica auditoría web/axe a esta terminal.
- No se descargaron ni ejecutaron paquetes públicos en las pruebas. Se usaron binarios locales reales y dobles sólo en límites externos. No se modificó la instalación real ni se reinició el visor.

## Qué validar

En el visor están los campos Source/Release, la protección del builddev, tamaño mínimo y los renders de preparación, confirmación, éxito y recuperación. Los primeros proceden de PTY real con dos hosts sintéticos; los segundos del árbol TUI integrado con ejecutor real y HTTP simulado.

Revisar el selector Mode, las versiones/destinos antes de Apply y la instrucción de `/hooks` al terminar. La elección humana pendiente decide si seguir con push, PR y merge a development (D2 A), o ajustar el flujo. Una release pública y la promoción a master requieren otra instrucción.
