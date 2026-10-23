# Skill `flow-walkthrough`: recorrido de UI con capturas, video y página de entrega; verificación de UI declarable por proyecto

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · fusionado en `development` por [#73](https://github.com/JhonHawk/tricell-hive/pull/73) (`bb4fd503`) el 2026-10-18, a decisión del usuario antes de las pruebas en sesiones reales previstas; desplegado en local |
| Tracker · GitHub Issues | • [#67 — Skill flow-walkthrough: recorrido de UI con capturas, video y página HTML de entrega](https://github.com/JhonHawk/tricell-hive/issues/67) |
| Git | E1: rama `feat/gh-67-flow-walkthrough-skill` desde `development` `b73bb16e` (tras fusionar la rama de prueba anterior en #68 y cerrar su registro en #69) · un commit por tarea verificada · sin push ni merge · despliegue local solo a petición (recompilar el binario primero: T6 toca el manager) |
| Verificación | `go vet ./...`, `go test ./...` · `python3 -m unittest discover -s tests/skills -p '*_test.py'` · `self_check.py` sobre el esqueleto y una página de muestra · grabación de prueba con `agent-browser record` leída con `ffprobe` · `pnpm -C site build` · tamaños de `global.md`, `verification.md` y la skill nueva |
| Siguiente paso | Ninguno: registro archivado |

## Petición

> Cursor en los cloud enviroments entrega un reporte como este [Image #1] en el cual pone las imagenes y un video del flujo que se corrio sobre la funcionalidad, creo que la etapa de revisión (local), puedes volverse una skill exclusiva que indique cómo hacer los recorridos y aprovechar las imagenes o incluso generar video de la funcionalidad para presentarsela al usuario.
>
> Analiza esto y dame una propuesto deep-research al respecto

> Consideremos que el medio de presentación tambien sea un html solo en darkmode como el de la captura de cursor

> tambien toma en cuenta, lo siguiente, ya que toma un poco más de tiempo el tema de la validación de ux y el in-vivo, creo que es momento para tomar la desición de si esto debe ofrecerse como un paso opcional en el proyecto (y tambien usar una bandera en las reglas del proyecto para omitir la pregunta)

Respuestas del 2026-10-17: D1-A (skill nueva que absorbe la sección de entrega de UI), D2-A (grabar por defecto en la entrega para validación humana, una pasada corta por flujo tras pasar la verificación), D3 sí (issue #67 abierto), D4-A (planear con `flow-plan`). Del 2026-10-18: la página HTML solo en modo oscuro como medio de presentación; P1-A (clave `UI verification` con `full`, `functional` y `none`); P2-A (sin clave, `flow-build` pregunta una vez por repositorio y sesión en la ronda de la pregunta de revisión).

## Objetivo

Hoy la entrega de una interfaz para que la valides vive dentro de la referencia de verificación de `flow-build`, en la sección "Human UI handoff", mezclada con las puertas de CI y despliegue en un archivo de 4.383 palabras. Produce capturas y un mensaje con pasos, pero no graba el flujo, no presenta una tabla de comprobaciones y no se puede invocar sola para una funcionalidad ya construida. Además, los dos hijos independientes de UI (`hive-review-ux` y `hive-verify-change`) son obligatorios para todo cambio con efecto visual, y sus rondas superan con frecuencia los cinco minutos (vigilancia V1-A; caso carvento del 2026-10-13: unos 73 y 40 minutos).

El cambio hace dos cosas. Primero, crea la skill `flow-walkthrough` como única casa del recorrido para validación humana: cuándo aplica, la primera vista rendida, las rondas de corrección y sus marcadores, las capturas, una pasada limpia grabada por flujo, y la entrega en dos medios: el mensaje con bloques por flujo y la línea de espera, y una página HTML solo en modo oscuro en la carpeta de trabajo, con la tabla de validación, las capturas y el video. Segundo, convierte la verificación independiente de UI en un ajuste del proyecto: la clave `UI verification` de `## Hive` con `full`, `functional` y `none`; sin clave, `flow-build` pregunta una vez por repositorio y sesión, en la misma ronda que la pregunta de revisión, y el manager valida la clave.

## Alcance y aceptación

Incluye:

- `content/skills/flow-walkthrough/SKILL.md`, nueva, con el procedimiento de `design.md` DS1–DS3.
- Revisión R1-A (2026-10-18): `content/skills/flow-walkthrough/references/ui-gate.md`, nueva casa de la puerta de UI (DS7), y `ui-review-criteria.md` movido de `flow-build/references/` a `flow-walkthrough/references/`; `verification.md` pierde la sección de puertas de UI salvo el puntero; repuntan `flow-build/SKILL.md`, `flow-plan/SKILL.md`, `plan-format.md`, `ui-planning.md`, `browser-automation.md`, `delivery-decisions.md` y `hive-review-ux.md`.
- `content/skills/flow-build/references/verification.md`: la sección "Human UI handoff" (líneas 65–79 en la base) queda reducida a un puntero; la fila "Local candidate" (línea 33) nombra la skill; el párrafo de la puerta de UI (línea 61) selecciona los hijos por el valor declarado o respondido (DS6).
- Punteros a esa sección: `content/skills/flow-build/SKILL.md` (líneas 34 y 75), `content/skills/flow-plan/SKILL.md` (línea 62), `content/agents/review/hive-review-ux.md` (paso 3).
- `content/agents/quality/hive-verify-change.md`: paso de grabación cuando el brief lo pide.
- `content/skills/flow-report/assets/skeleton-walkthrough.html`, nuevo esqueleto solo en modo oscuro, y `content/skills/flow-report/SKILL.md`: fila nueva en la tabla de activos, la frase que excluye los handoffs del HTML (línea 10) y la regla de archivos dentro del documento (línea 25).
- `content/skills/workspace-conventions/references/artifact-layout.md` y `content/skills/flow-close/SKILL.md` (paso 1 de "End the session"): grabaciones y página con el mismo ciclo de vida que las imágenes.
- Ajuste `UI verification`: `content/skills/flow-plan/references/project-settings.md`, `content/skills/flow-plan/references/delivery-decisions.md`, `content/skills/flow-build/SKILL.md` (líneas 20 y 36), `content/skills/flow-plan/SKILL.md` (línea 18).
- Manager: `tooling/cli/doctor_project.go`, `tooling/cli/tui_project_view.go` y sus pruebas; `_support/docs/architecture/deployment-manager.md`.
- `tests/pilot/flows.go` (lista `deployed-smoke`), `tests/pilot/flows_test.go`, `tests/fixtures/flows/cases.json` (prompt que nombra las skills) y `tests/fixtures/flows/README.md`.
- Documentación: `site/src/content/docs/flows/build.md` (párrafo de ajustes declarados y FAQ de cambios de interfaz), `site/src/content/docs/flows/close.md` (FAQ de qué borra) e `site/src/content/docs/index.mdx` (lista de skills de apoyo).

Queda fuera:

- Adjuntar capturas o video a un pull request de forma automática o preguntada. La skill documenta el comando y los límites de GitHub; se ejecuta solo cuando el usuario lo pide, porque publicar una captura no está autorizado por retenerla.
- Un visor o servidor web: la página se abre desde el sistema de archivos.
- Grabar bajo `Delivery: automatic`: ese modo no se detiene para validación humana (decisión G1-B del 2026-10-13) y no hay entrega que acompañar.
- `llms.txt`: lista solo las cuatro skills de ciclo (`llms.txt:28–31`); las de apoyo, como `flow-report`, no aparecen, y este cambio no altera nada que declare.
- Cambios en `content/guidance/global.md`: ninguna regla del núcleo cambia; la lista de claves de `## Hive` vive en `project-settings.md` y en la tabla del manager, que una prueba mantiene alineadas (`tooling/cli/doctor_project_test.go:455`).
- Pilotos de comportamiento: están pausados. La validación es tu uso en sesiones reales.
- `hive-verify-task` por tarea, la primera vista rendida y la parada de validación humana bajo `interactive` u `hold`: no cambian con ningún valor de `UI verification`.

Restricciones que deben mantenerse:

- `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasan; `pnpm -C site build` sale con 0.
- `content/guidance/global.md` no cambia por este cambio (4.563 palabras en la base). Nota: el commit `76408ee7` de la misma rama, decidido aparte (R1-A del 2026-10-18 tras auditar la sesión de cam), añade 20 palabras a la regla de idioma; el PR lo declara como cambio independiente.
- Cada regla conserva una sola casa: lo que se mueve a la skill desaparece de la referencia de verificación, y los puntos de uso solo apuntan.
- El texto distribuido va en inglés, con el criterio y el alcance de cada regla explícitos, para los modelos del perfil `execution`.

Criterios de aceptación:

- AC1. Existe `content/skills/flow-walkthrough/SKILL.md` y nombra: los dos disparadores (la parada de validación humana de `flow-build` y una petición directa), cuándo no aplica (`automatic`, cambios solo de texto o estilo), la primera vista rendida, las rondas de corrección y los marcadores, las capturas, la pasada limpia grabada por flujo con `record start` y `record stop`, la comprobación con `ffprobe`, qué pasa sin ffmpeg o ffprobe, la obsolescencia de una grabación tras una corrección, los bloques del mensaje con la línea de espera, y la página HTML con su destino. *Falso en la base cuando* el archivo no existe.
- AC2. La sección "Human UI handoff" de `verification.md` contiene solo un puntero a la skill: `rg -n "correction round per request|first rendered version|ARK-123" content/skills/flow-build/references/verification.md` no devuelve nada; `rg -l "verification.md#human-ui-handoff" content/` no devuelve nada (en la base devuelve `flow-build/SKILL.md:34` y `flow-plan/SKILL.md:62`); `flow-build/SKILL.md:75` y el paso 3 de `hive-review-ux.md` enlazan la skill para la entrega. *Falso en la base cuando* la primera búsqueda devuelve líneas y la segunda devuelve dos archivos.
- AC3. `hive-verify-change.md` tiene un paso que, cuando el brief lo pide, graba una pasada limpia por flujo después de que sus comprobaciones pasen y reporta "not recorded" tras dos fallos. *Falso en la base cuando* `rg -n "not recorded" content/agents/quality/hive-verify-change.md` no devuelve nada.
- AC4. Existe `content/skills/flow-report/assets/skeleton-walkthrough.html`, solo en modo oscuro en pantalla (`rg -c "data-theme|prefers-color-scheme" <archivo>` da 0; `color-scheme: dark` presente; el bloque `@media print` imprime en claro como los demás esqueletos), con marcadores para título y candidato, tabla de validación, un bloque por flujo con pasos, `<img>` y `<video controls preload="metadata">` con rutas relativas a `images/`, sección de acción requerida y pie de procedencia, sin atributos `on*`; `python3 content/skills/flow-report/scripts/self_check.py` sobre el esqueleto y sobre una página de muestra sale con 0; `flow-report/SKILL.md` lo lista en la tabla de activos, acota la exclusión de handoffs del HTML a los handoffs ordinarios y exceptúa los archivos hermanos del recorrido de la regla "keep … images … in the file". *Falso en la base cuando* el archivo no existe y `rg -n "walkthrough" content/skills/flow-report/SKILL.md` no devuelve nada.
- AC5. `artifact-layout.md` y el paso 1 de "End the session" de `flow-close/SKILL.md` tratan la página de recorrido, sus grabaciones y sus imágenes como una unidad: misma carpeta, se conservan juntas cuando el usuario selecciona la página y se borran juntas al fin de sesión en caso contrario. *Falso en la base cuando* `rg -n -i "recording" content/skills/workspace-conventions/references/artifact-layout.md content/skills/flow-close/SKILL.md` no devuelve nada.
- AC6. La FAQ de `flows/build.md` menciona la grabación y la página; la FAQ de `flows/close.md` menciona las grabaciones entre lo que se borra; `index.mdx` lista `flow-walkthrough` bajo "Supporting skills". *Falso en la base cuando* `rg -n "flow-walkthrough" site/src/content/docs` no devuelve nada.
- AC7. La lista `deployed-smoke` de `tests/pilot/flows.go`, su prueba, el prompt de `tests/fixtures/flows/cases.json` y `tests/fixtures/flows/README.md` nombran `flow-walkthrough`, y `go test ./tests/pilot/...` pasa. *Falso en la base cuando* `rg -n "flow-walkthrough" tests/` no devuelve nada.
- AC8. `project-settings.md` documenta la clave `UI verification` con `full`, `functional` y `none`, el efecto de cada valor, la regla de pregunta única por repositorio y sesión cuando falta, el aviso sobre `none` bajo `automatic`, y la línea en el ejemplo. *Falso en la base cuando* `rg -n "UI verification" content/skills/flow-plan/references/project-settings.md` no devuelve nada.
- AC9. El párrafo de la puerta de UI (en `verification.md` hasta T7; desde la revisión R1-A, en `flow-walkthrough/references/ui-gate.md`), `flow-build/SKILL.md:36`, la plantilla de `plan-format.md:121` y la puerta del candidato aceptado movida a la skill seleccionan los hijos por el valor declarado o respondido y ya no llaman obligatorios a los dos sin condición; `delivery-decisions.md` pregunta el valor en la ronda de la pregunta de revisión cuando no está declarado y reutiliza una respuesta explícita de la sesión para el mismo repositorio; `flow-plan/SKILL.md:18` lo nombra en el traspaso. *Falso en la base cuando* ``rg -n '`UI verification`' content/skills`` no devuelve nada (la frase suelta "UI verification" sí aparece en `verification.md:33` y `flow-plan/SKILL.md:36`), `rg -n "requires two independent children" content/skills/flow-build/references/verification.md` devuelve la línea 61, `rg -n "are required, not proposed" content/skills/flow-build/SKILL.md` devuelve la línea 36 y `rg -n "not a pending choice" content/skills/flow-plan/references/plan-format.md` devuelve la línea 121.
- AC10. `hive doctor` acepta `UI verification` con sus tres valores y avisa ante cualquier otro; el formulario de proyecto del TUI ofrece los tres; `go test ./tooling/...` pasa con una prueba para la clave nueva. *Falso en la base cuando* `rg -n "UI verification" tooling/` no devuelve nada.
- AC11. `flows/build.md` nombra `UI verification` junto a `Delivery` y `Review` en el párrafo de ajustes declarados, y `deployment-manager.md` la lista entre las claves validadas. *Falso en la base cuando* `rg -n "UI verification" site/src/content/docs/flows/build.md _support/docs/architecture/deployment-manager.md` no devuelve nada.
- AC12 (revisión R1-A del 2026-10-18). La puerta de UI tiene una sola casa, `content/skills/flow-walkthrough/references/ui-gate.md`, que contiene la inspección renderizada del implementador, la selección de hijos por valor con sus briefs y topes, y la puerta del candidato aceptado, y dice que aplica en todo modo de entrega; `ui-review-criteria.md` vive en `content/skills/flow-walkthrough/references/` y ya no en `flow-build/references/`; `verification.md` y la skill no restatan la puerta (`rg -n "Brief the reviewer|scoped to the diff" content/skills/flow-build/references/verification.md content/skills/flow-walkthrough/SKILL.md` no devuelve nada); ningún archivo de `content/`, `site/` ni `_support/docs/architecture/` enlaza `flow-build/references/ui-review-criteria`, y ningún enlace relativo a `ui-review-criteria.md` resuelve a `flow-build/references/` (`browser-automation.md:20` incluido); `flow-build/SKILL.md`, `flow-plan/SKILL.md`, `plan-format.md`, `ui-planning.md`, `browser-automation.md`, `delivery-decisions.md` y `hive-review-ux.md` enlazan `ui-gate.md`. *Falso en la base cuando* `test -f content/skills/flow-walkthrough/references/ui-gate.md` falla y `rg -n "Brief the reviewer" content/skills/flow-build/references/verification.md` devuelve la línea 61.

## Entrega

Respuestas del 2026-10-18:

- **E1 (modo):** primero comprobar si `feat/lifecycle-declared-delivery-review` está fusionada en `development`; si no, fusionarla (tu autorización explícita en E1) y cerrar su registro; después cortar una rama nueva desde `development` actualizado (propuesta: `feat/gh-67-flow-walkthrough-skill`), un commit por tarea verificada cuando llega a `[x]`, sin push ni merge de esta rama; tú pruebas en sesiones reales y decides el merge. Despliegue local solo cuando lo pidas, recompilando el binario antes porque T6 toca el manager.
- **E2-A (revisión):** ninguna dedicada; sin revisión de seguridad porque no se cruza ninguna frontera.
- **Q3-A:** con `UI verification: none`, un `hive-verify-change` lanzado solo para grabar.
- **R1-A:** `flow-build` arranca ahora.
- **Registro:** esta carpeta queda sin versionar hasta el merge; `flow-close` la versiona por la ruta de cierre que elijas entonces.
