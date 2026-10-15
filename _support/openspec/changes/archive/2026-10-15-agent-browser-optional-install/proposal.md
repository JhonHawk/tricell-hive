# agent-browser como capacidad opcional del instalador

| Campo | Valor actual |
| --- | --- |
| Estado | Integrado en `development` el 2026-10-15 por el [PR #52](https://github.com/JhonHawk/tricell-hive/pull/52), merge `1307b1fb`; desplegado en local desde `development` y archivado el 2026-10-15. Estado previo (historia): En validación · T1–T5 verificadas con commits `b0fb62e0`–`673cbe4e` y correcciones de revisión `067d0453` en `feat/flash-class-guidance`; desplegado en local en `f8116af5` el 2026-10-13; sin push |
| Tracker · GitHub Issues | Sin issue asociado; relacionado: [#50](https://github.com/JhonHawk/tricell-hive/issues/50) (navegador bloqueado en Codex `observe`, fuera de este alcance) |
| Git | Integrado por merge del PR #52; la rama `feat/flash-class-guidance` y su worktree se eliminaron tras el merge |
| Verificación | `go vet ./...` · `go test ./...` · gate nativo: macOS solo descarga de Chrome con `HOME` temporal (G1-A), Linux en contenedor con Node 24; pendiente de autorización |
| Siguiente paso | Ninguno; registro cerrado |

Fecha inicial: 2026-10-13, hora local CST.

## Pedido

> hagamos research para ver si cambiamos agent-browser por devtools, playwrigth u otra opción como principal o acotar en cada opción que va como prioridad por situación, ademas de validar si hacerlo por cli, mcp, etc y qué pasaria si no se tiene la herramienta) la opción fácil es que el agente la instale

Decisiones del usuario (2026-10-13):

- R2-B: «que el instalador y hive doctor ofrezcan agent-browser como capacidad opcional al instalar Hive, con tu consentimiento».
- D1-A: instalar de verdad tras el consentimiento en el asistente.
- D2-A: `brew` en macOS y `npm` en Linux.
- D3 (texto libre): «segun yo, en pi-subagents, instalamos la latest» → sin versión fija, como `npm:pi-subagents` en la rama `feat/pi-subagents-unversioned-install` (commit `57d2fcab`).
- D4-A, sustituida por D4b-A al corregirse el dato (agent-browser sí detecta navegadores instalados): descargar Chrome solo si no hay ningún navegador en las rutas que revisa agent-browser.
- G1-A: en macOS el gate nativo prueba solo la descarga de Chrome con `HOME` temporal; la ruta de `brew` queda cubierta por pruebas con dobles.
- O1-A: la rama `feat/pi-subagents-unversioned-install` se integra antes; este cambio se escribe sobre su texto.

## Objetivo

La investigación del 2026-10-13 mantuvo `agent-browser` como herramienta principal (≈6.500 llamadas reales en cinco hosts frente a 188 a MCP; en la prueba R2-C fue 3 veces más rápido y con 5 veces menos salida que `@playwright/cli`). Pero Hive nunca la instala: si falta, el agente queda sin navegador a mitad de una tarea y la nueva regla de `browser-automation.md` solo puede pedir al usuario que la instale. Este cambio hace que `hive install` la ofrezca como capacidad opcional y, con el consentimiento del usuario, la instale y descargue Chrome, para que la herramienta esté presente antes de que una tarea la necesite.

## Alcance y aceptación

Incluye: catálogo y validación de proveedores (`tooling/providers`), adaptador, oferta sin versión y ejecutor en `tooling/cli`, fila de `hive doctor` y vista Integrations, especificación `versioned-installation`, `SECURITY.md`, documentación de arquitectura (`installer.md`, `tool-dependencies.md`, `deployment-manager.md`), `README.md`, `llms.txt` y `CHANGELOG.md`.

Excluye:

- Instalar o configurar servidores MCP de navegador.
- Desinstalar `agent-browser` o Chrome desde Hive (`deployment-manager.md:359`).
- `agent-browser install --with-deps` y cualquier `sudo`.
- Plataformas distintas de macOS arm64 y Linux amd64: allí la capacidad queda manual.
- El caso Codex `observe` ([#50](https://github.com/JhonHawk/tricell-hive/issues/50)).

Restricciones:

- `go vet ./...` y `go test ./...` pasan.
- Sesiones no interactivas, `--dry-run` y `--home` sintético no ejecutan ningún programa.
- `hive doctor` sigue siendo de solo lectura.
- Ningún comando usa shell ni `sudo`.
- Engram, Context7 y pi-subagents siguen manuales.
- Depende de que `feat/pi-subagents-unversioned-install` esté integrada en `development` y de que `feat/flash-class-guidance` la incorpore (O1-A, D-12).

Criterios (base: la punta de `feat/flash-class-guidance` tras incorporar `development` con la rama de pi-subagents integrada, D-12; al iniciar el build se vuelve a comprobar que cada criterio es falso allí; las observaciones de abajo se tomaron en `5d15446f`):

- AC1. En `hive install` interactivo en macOS arm64 con `brew`, o en Linux amd64 con Node ≥ 24 y prefijo global de npm escribible, cuando `agent-browser` no está en el PATH, el asistente ofrece `agent-browser` sin pedir versión, y su resumen muestra antes del consentimiento el comando exacto (`brew install agent-browser` o `npm install -g agent-browser`), que se instala la última versión, que `brew` instala Node y sus dependencias, y `agent-browser install` solo si no hay navegador (D-5). *Falso en la base cuando* `providers.Catalog()` no contiene `agent-browser`.
- AC2. Retirado: ya era cierto en la base; su comprobación quedó dentro de AC1 y AC4.
- AC3. Tras el consentimiento, cada comando se ejecuta sin shell, en su grupo de procesos, con `Stdin` vacío, límite de 15 min y salida guardada recortada a 64 KiB, y cada paso queda `verified` según su propia comprobación (D-7). *Falso en la base cuando* `ValidateStep` rechaza todo paso no manual (`tooling/providers/run.go:25`).
- AC4. Si falta un requisito (gestor, Node ≥ 24, prefijo escribible, plataforma), la capacidad aparece como manual con el motivo y el comando oficial. Si un comando falla, expira o se interrumpe, el paso queda `failed` con su mensaje visible en el detalle del resultado, el paso de Chrome queda `skipped` si el primero falló, el núcleo se conserva y el diario nunca queda `unknown`. Una ejecución posterior con `agent-browser` presente y sin navegador ofrece solo el paso de Chrome. *Falso en la base cuando* `showPartialOnboardingDetail` (`install.go:862-876`) solo muestra `ManualReason` y no existe ningún paso ejecutable.
- AC5. `hive doctor` y la vista Integrations, con `agent-browser` ausente o sin navegador, indican que `hive install` puede instalarlo, y la columna Record muestra el peor estado de sus pasos. *Falso en la base cuando* `doctor_integrations.go:241-253` imprime «Hive does not install it».
- AC6. La especificación, `SECURITY.md` y la documentación describen la capacidad, su receta, el consentimiento, la ausencia de shell y `sudo`, la versión sin fijar, la descarga del binario sin suma por npm, las plataformas probadas y la semántica de fallo, sin dejar ninguna frase que diga que Hive no instala ejecutables o que solo pi-subagents descarga. *Falso en la base cuando* `spec.md:62` limita el catálogo a tres y `SECURITY.md:32` dice «Hive does not install CLI executables».
- AC7. `ValidateStep` acepta para `agent-browser` solo los `argv` permitidos ligados a su ID de paso y un paso `agent-browser-manual`, y rechaza cualquier otro. *Falso en la base cuando* `ValidateStep` rechaza el proveedor `agent-browser`.

## Entrega

Respuestas del usuario (2026-10-13):

- E1-A: un commit por tarea verificada en `feat/flash-class-guidance`, sin la carpeta de cambio; sin push; recompilar el binario y desplegar en local solo a pedido; push, PR y merge después de las pruebas del usuario.
- E2-A: `/code-review` de Claude Code sobre el diff antes de dar el cambio por terminado (límite de seguridad: ejecución de procesos y cadena de suministro).
- E3-A: el plan queda listo y sin commit; el build empieza cuando `feat/pi-subagents-unversioned-install` esté integrada en `development` y `feat/flash-class-guidance` la incorpore.
- El gate nativo T5 sigue pendiente de autorización explícita al llegar a él.
