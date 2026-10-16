# Evidencia del runner de pilotos: texto de shell, patrones y cobertura normalizada (#48, #49)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado por la [PR #63](https://github.com/JhonHawk/tricell-hive/pull/63) en `development` |
| Tracker · GitHub Issues | • [#48 — Piloto: la inspección de operaciones no puede aceptar corridas](https://github.com/JhonHawk/tricell-hive/issues/48)<br>• [#49 — Piloto: Coverage de las variantes A/B no puede salir observada](https://github.com/JhonHawk/tricell-hive/issues/49) |
| Git | automática · rama `feat/gh-48-49-pilot-runner-evidence` → `development` · sin parada humana |
| Verificación | `go test ./tests/pilot/` · `go vet ./...` y `go test ./...` · verificador por tarea · `/code-review` de Claude Code en la PR |
| Siguiente paso | Ninguno: cierre del registro con la PR #63 |

## Petición

> D1-A — Continuar con #48 y #49 ahora

(2026-10-16.) Acciones sugeridas de los issues, leídas el 2026-10-16:

> #48: Guardar el texto completo de `shell`/`execute` y las entradas de `glob`/`grep` del primario y los hijos en un archivo de evidencia privado (0600, fuera de `run.json`), para revisarlo tras el lote, en vez de retenerlo. Quitar `.git/` de la fixture o registrar en el inventario el tipo de cada entrada (archivo regular frente a enlace).

> #49: Confirmar el formato de salida de `read` en la evidencia retenida y, si se confirma, comparar el contenido normalizado (sin prefijos de línea) o por hash del contenido normalizado.

## Objetivo

El lote `fcg-20261012c` terminó con 0 de 48 corridas válidas para la inspección de operaciones, porque el runner retiene el texto de `shell` fuera de la gramática de lectura acotada (D22-A) y no conserva los patrones de `glob`/`grep`, y porque el inventario de la fixture no registra el tipo de cada entrada ni declara el `.git/` que el propio runner crea; y la cobertura de las variantes A/B quedó siempre `unverified` porque el detector exige el texto canónico byte a byte, que la herramienta `read` de OpenCode altera con prefijos de línea. Cualquier piloto futuro con este runner repetiría ambos resultados. El cambio conserva la evidencia que la inspección necesita, en privado, y compara la cobertura sobre contenido normalizado. No reanuda pilotos (siguen pausados) ni toca la guía distribuida.

## Alcance y aceptación

Incluido: `tests/pilot/` (runner, trazas, inspección, detector de cobertura y sus pruebas) y `tests/fixtures/flash-class-guidance/observer.js`.

Excluido: los casos de `tests/fixtures/regression/` (sin cambios, `AGENTS.md`); reanudar pilotos; cambios en guía distribuida.

Restricciones: `go test ./tests/pilot/` y la suite completa en verde; `TestFlashCommandSecurity` sigue garantizando que un comando ordinario de solo lectura no se retiene; el archivo de evidencia privado nunca entra en `run.json` ni en los artefactos que se publican; las pruebas usan fixtures sintéticas sin llamar a ningún modelo.

- AC1. Para cada llamada `shell`/`execute` cuyo texto queda fuera de la gramática de lectura acotada, y para cada llamada `glob`/`grep` del primario y de los hijos, el runner guarda el texto completo del comando o el patrón en un archivo de evidencia privado con modo `0600`, fuera de `run.json`, y la inspección de operaciones lee ese archivo para aceptar o rechazar la operación en vez de rechazarla por texto retenido. *Falso en la base cuando* `flash_class.go:282` marca `command_withheld` sin guardar el texto y `trace.go` no conserva patrones de `glob`/`grep`.
- AC2. El inventario de la fixture permite a la inspección probar que una búsqueda sobre la raíz no siguió un enlace simbólico: cada entrada registra su tipo (archivo regular o enlace), y las entradas del `.git/` que crea el runner quedan excluidas de la raíz de la fixture o inventariadas con su tipo. *Falso en la base cuando* `main.go:147-171` guarda solo digest y tamaño, y `fixture.json` no declara las entradas de `.git/`.
- AC3. La cobertura compara el texto de la referencia normalizado (sin los prefijos de número de línea que añade `read` de OpenCode, con espacio final recortado) contra el canónico, o por hash del texto normalizado, de modo que un hijo que leyó la referencia con `read` da `observed` en las variantes A y B, y una copia byte a byte sigue contando. *Falso en la base cuando* `observer.js:46-68` exige `text.includes(canonical)` y `flash_class.go:380` exige `r.Hash == canonicalHash`.

## Contexto verificado

- `tests/pilot/flash_class.go:267-285`: `commandWithheld` se fija cuando la herramienta es `execute` o cuando el comando original existía y la gramática lo vació; `trace.go:27` serializa `command_withheld`; `flash_protocol.go:30-115` define la gramática (`flashOperationCommand`, `flashReadOnlyArguments`, `rg`/`grep` en `:68`).
- `tests/pilot/main.go:147-171` (`inventory`): digest y tamaño por entrada, enlace como `symlink:<destino>`; `:209,249` crean `.git/pilot-remote.git` bajo la raíz; `:878` guarda `fixture.json`; `assess.go:47` lo lee.
- `tests/fixtures/flash-class-guidance/observer.js:46-68` (`inspectReference`): copias por `includes` del canónico o del prefijo inyectado; `flash_class.go:358-390` (`flashCoverage`).
- La evidencia del lote `fcg-20261012c` no está retenida en `_support/workspace/`; el formato de `read` de OpenCode se confirma en su código fuente o documentación (Context7 `/sst/opencode`) y se cubre con una prueba sintética.
- `go test ./tests/pilot/` tarda ~22 s en verde en la base.

## Diseño propuesto

- Evidencia privada: un archivo por corrida junto a `run.json` pero fuera de él (por ejemplo `operations-private.jsonl`, `0600`), con una línea por operación: sesión, agente, herramienta, texto completo del comando o patrón y la clasificación de la gramática. La inspección (`assess.go`) lo lee cuando `command_withheld` está presente y decide con el texto; `run.json` conserva solo el indicador y el nombre del archivo.
- Inventario: `item` añade `kind` (`file`/`symlink`); el runner excluye `.git/` de la raíz inventariada o lo inventaría con tipo; la inspección usa el tipo para la prueba de no seguir enlaces. El implementador elige entre excluir o inventariar según cuál deje la contención demostrable con menos código, y lo registra.
- Cobertura: `inspectReference` normaliza cada texto candidato (quita prefijos `^\s*\d+[|:]\s?` por línea y espacio final) antes de comparar; `hash` sobre el normalizado; `canonical` y `copies` siguen el mismo criterio; `flashCoverage` compara el hash normalizado.

## Entrega

Repositorio `tricell-hive`, base `development`, modo automático (D2-B, 2026-10-16) con `/code-review` de Claude Code (D3-A): commits por tarea verificada, push, PR a `development`, merge cuando pasen suite y revisión; registro versionado al cierre en la misma rama. Al integrar: cerrar #48 y #49. Sin release ni despliegue de contenido (no cambia `content/`).

## Estado de revisión y progreso

Base: `69dce947` (`development`, 2026-10-16). Sin revisión de plan delegada: alcance fijado por los dos issues y decisiones del usuario en la sesión.
