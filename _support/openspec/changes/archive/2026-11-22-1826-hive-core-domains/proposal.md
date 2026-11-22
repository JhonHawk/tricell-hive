# hive-core-domains — Hojas y principios de hive-core en carpetas de dominio

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #587, d4f1eedb |
| Tracker | ninguno |
| Git | tricell-hive · base `development` (`da2a4f27`) · se construye sobre `docs/inspiration-credits-index` (`657c1ea0`), que ya quitó los `references/notice.md` |
| Playbook | refactor |

## Decisions

2026-11-22
- **D3 A** — ¿Qué hacemos con los `hive-principle-*`? Moverlos a `hive-core` sin cambiar su texto; `hive-typescript` sigue como skill.
- **D4 A** — ¿Bajo qué dominio vive `principles/`? Dentro de `hive-core`, con índice y reglas de forma propios, y no en `references/` (pedido explícito del usuario). Descartadas: una skill propia `hive-principles/` y subcarpetas por tema.
- **D5 B** — ¿Las hojas de `references/` se reparten por dominio en este cambio o en otro? En este: `hive-core` queda con N carpetas de dominio al mismo nivel y sin `references/` («en lugar de 1 references, serían N cantidad de carpetas al mismo nivel, agrupados por dominios»).
- **D6 A** — ¿Cómo sabe el manager qué carpetas de dominio instalar? Por un manifiesto `hive-core/assets/domains.json` que el manager y los tests de contenido leen; el manager instala y valida solo las carpetas listadas. Descartados: nombres fijos en Go (lista en dos lugares) y cualquier subcarpeta (copiaría carpetas no declaradas a todos los hosts).
- **D7 A** — ¿Apruebas la tabla de dominios? Sí, tal como está en `design.md` «Proposed shape».
- **D9 A** — ¿Cómo arreglamos `TestStageManifests`, que sumaba los bytes de los miembros de cada paquete contra un proxy de 22.500 B? Midiendo el paquete generado: un test en `tooling/management` renderiza los paquetes del contenido real con el código del manager y exige 500 B de margen bajo el tope de 24.000 B; el test de contenido deja de sumar miembros. Motivo: la mudanza subió los miembros de flow-build a 22.560 B solo por rutas entre dominios, mientras el paquete generado bajó de 23.641 a 23.427 B.
- **D10 A** — ¿Qué revisa `/code-review`? Solo lo no mecánico: el manager, el auditor, los tests de contenido, los deltas de spec y la prosa reescrita. Las 78 mudanzas quedan cubiertas por la comprobación de texto idéntico fuera de los enlaces y por la instalación real con todos los enlaces resueltos.
- **supuesto** — Las carpetas de dominio reemplazan el nivel `references/` y no van debajo de él. Alternativa descartada: `references/<dominio>/`, que suma ~750 B a `global.md`, que solo tiene 383 libres.
- **supuesto** — Los principios pierden el prefijo: `principles/boundary-discipline.md`. Así coinciden con los nombres que ya escribe la línea `- Principles:`.
- **supuesto** — `hive-core/scripts/<hoja>/` no se mueve: `scripts` no es un dominio y sus 3 scripts tienen rutas que invocan flujos y tests.
- **supuesto** — Se construye sobre `docs/inspiration-credits-index`, porque en `development` aún existen las carpetas `references/<hoja>/notice.md`.
- **supuesto** — El retiro de las rutas viejas no lleva código nuevo: es genérico (`tooling/management/ownership.go:257-285`, `apply.go:171-187`). Sí lleva un test nuevo en T1, porque ningún test actual retira archivos de una subcarpeta de una skill que se queda ni una skill `hive-principle-*` con su `agents/openai.yaml`.

## Business rules

None: el trabajo no adopta ninguna regla de negocio.

## Request

«Planear la mudanza a `hive-core/principles/` sin tocar su contenido, dejando `hive-typescript` como skill» (D3 A). «No quiero que se vaya a la carpeta de references, quiero que se vaya a una de principles y se manejen diferente, bajo el dominio que les toca» (D4 A). «En lugar de 1 references, serían N cantidad de carpetas al mismo nivel, agrupados por dominios», en este mismo cambio (D5 B). Fuera de alcance: el texto de cualquier hoja o principio salvo sus enlaces, y `hive-typescript`.

## Objective

Hoy las 60 hojas de regla están juntas en `hive-core/references/`, y los 9 principios con su índice son 10 skills ocultas aparte. Al terminar, `hive-core` tiene una carpeta por dominio (`messages`, `git`, `plan`, `build`, `delegation`, `ui`, `review`, `business`, `close`, `local`, `tracker`, `principles`), declaradas en `assets/domains.json`, cada hoja en la suya, y ninguna skill `hive-principle-*`. El agente sigue leyendo cada hoja por la ruta que le dan el índice, el flujo o el paquete; solo cambia la ruta.

## Scope and acceptance

Dentro: el manager, que instala y valida las carpetas que declara el manifiesto; mover 60 hojas, sus 5 carpetas de referencias propias (8 archivos) y los 10 principios; reescribir cada enlace en `content/`, `site/`, `README.md`, `llms.txt`, `AGENTS.md`, `_support/docs/` y los `assets/stage.json`; la descripción de `hive-core`; los deltas de `context-delivery` y `versioned-installation`; los tests de contenido y del manager, y el auditor de sesiones; la entrada de `CHANGELOG.md`. Fuera: el texto de las hojas salvo enlaces, `hive-core/scripts/`, las líneas `- Rules:` y `- Principles:`, y los registros archivados de `_support/openspec/changes/archive/`.

- AC1 — El manager, desde un checkout y desde una release guardada, instala cada carpeta que lista el `assets/domains.json` de una skill, valida los enlaces Markdown que hay en ellas, acepta estado registrado con rutas de dominio, no instala una carpeta de primer nivel que no esté listada (salvo `references`, `scripts`, `assets` y `playbooks`), y rechaza la release, nombrando manifiesto y motivo, cuando el manifiesto está mal formado, tiene una clave desconocida, repite un nombre, usa uno reservado (esos cuatro y `agents`), nombra una carpeta que no existe o vacía, o la release trae archivos en una carpeta no listada · false today because `tooling/management/plan.go:246` solo recoge esas cuatro carpetas fijas y no lee ningún manifiesto · Observed in: `go test ./tooling/management/... -count=1` con los casos nuevos
- AC2 — `content/skills/hive-core/` tiene las 12 carpetas de dominio de `design.md`, declaradas en `assets/domains.json`, cada hoja en la suya con su texto igual salvo los destinos de sus enlaces, y no tiene `references/`; no queda ninguna carpeta `content/skills/hive-principle-*` · false today because `content/skills/hive-core/references/` tiene 60 hojas y existen 10 carpetas `hive-principle-*` · Observed in: el árbol `content/` del candidato y `git diff -M --stat HEAD`, que muestra cada hoja como renombrada
- AC3 — `rg -l 'hive-core/references|hive-principle-' content site README.md llms.txt AGENTS.md _support/docs tooling --glob '!*_test.go'` no devuelve nada (en `_support/docs/` ningún archivo con rutas viejas es un registro fechado, así que todos se corrigen); y cada enlace resuelve tras instalar. Quedan fuera del comando, porque conservan las rutas viejas a propósito: los tests y fixtures (`*_test.go`, `tests/fixtures/`), `tests/pilot/` (reconoce rutas viejas), `CHANGELOG.md` (narra el cambio), `_support/openspec/` (specs hasta el archivado, esta carpeta y los archivados) · false today because `rg` encuentra 659 apariciones en 95 archivos de `content/` · Observed in: ese comando, y `go test ./tooling/management/ -run 'TestRealContentInstallsHiveCoreAndResolvesEveryGlobalLink|Stage' -count=1`, que instala el contenido real, valida cada enlace y genera los paquetes
- AC4 — `content/guidance/global.md` mide menos de 29.617 bytes · false today because mide 29.617 · Observed in: `wc -c content/guidance/global.md`
- AC5 — Los tests de contenido fallan, nombrando el archivo, cuando existe `hive-core/references/`, una carpeta de primer nivel de `hive-core` que no está en `domains.json` ni es `assets` o `scripts`, un dominio vacío, un nombre de hoja repetido entre dominios, una carpeta `hive-rule-*` o `hive-principle-*`, un principio con frontmatter, de más de 4.000 bytes, sin «Read», ausente o repetido en `principles/index.md`, un texto distribuido que nombra una skill `hive-principle-*` o la ruta `hive-core/references/`, o un archivo de referencias propio de una hoja (`<dominio>/<hoja>/*.md`) de más de 16.000 bytes · false today because `tests/content/leaf_skill_test.go` busca las hojas en `references/` y los principios como skills, y `retired_names_test.go` no conoce el prefijo `hive-principle-` · Observed in: `go test ./tests/content/... -count=1` y sus casos de violación
- AC6 — El auditor de sesiones reconoce la lectura de una hoja por su ruta nueva y sigue reconociendo las dos rutas anteriores · false today because sus patrones solo aceptan `hive-rule-<n>/SKILL.md` y `hive-core/references/<n>.md` · Observed in: `go test ./tests/pilot/... -count=1`, con un caso de sesión que lee una hoja por su ruta nueva
- AC7 — Una actualización desde una instalación con `hive-core/references/` y carpetas `hive-principle-*` las retira, con el `agents/openai.yaml` de Codex de cada skill retirada, e instala las carpetas de dominio · false today because la release actual instala `references/` y las skills de principios · Observed in: el test de retiro nuevo de T1 (instalación previa con `hive-core/references/aa.md` y una skill oculta `hive-principle-x`, actualización a un dominio, y una segunda actualización que no cambia nada); la instalación local refrescada (binario y luego contenido) cuando llegue a `development` lo confirma en hosts reales, y bajo `hold` esa parte queda sin verificar

## Prerequisites

None: la construcción no necesita nada de fuera del repositorio.

## Verified context

- `global.md` mide 29.617 bytes (tope 30.000); con la tabla de dominios aplicada a sus 85 rutas mediría 29.210 · `wc -c` y sustitución medida en esta sesión · 2026-11-22
- La tabla de dominios cubre las 60 hojas, sin faltantes ni sobrantes · comparación con `ls content/skills/hive-core/references/*.md` · 2026-11-22
- El manager recoge solo `references`, `scripts`, `assets` y `playbooks` de cada skill (`tooling/management/plan.go:246`), acepta solo esas en `validSource` (`plan.go:373`, también aplicado al estado registrado en `ownership.go:64`) y valida enlaces solo en `/references/` y `/playbooks/` (`references.go:25`) · lectura del código · 2026-11-22
- El retiro de skills que salen del contenido es genérico · `tooling/management/ownership.go:257-283` · 2026-11-22
- `rg -o 'hive-core/references|hive-principle-' content` · 659 apariciones en 95 archivos · 2026-11-22
- `rg -l 'hive-principle-' content site README.md llms.txt` · 26 archivos · 2026-11-22
- Uso real de los principios desde 2026-11-19: Codex en 49 sesiones, Grok en 24, Claude Code solo en tricell-hive · escaneo de transcripciones · 2026-11-22
- `go test ./tooling/management/... -count=1` · ok (108 s) · 2026-11-22
- `go test ./tooling/management/ -run 'TestRealContentInstallsHiveCoreAndResolvesEveryGlobalLink|Stage' -count=1` · ok (30 s) · 2026-11-22
- `go test ./tests/content/... -count=1` · ok (2,3 s) · 2026-11-22
- `go test ./tests/pilot/... -count=1` · ok (64 s) · 2026-11-22
- `go test ./... -count=1` · ok · 2026-11-22
- `python3 -I -m unittest discover -s tests/skills -p '*_test.py'` · 412 tests OK (94 s) · 2026-11-22

## Effects

None: las tareas no producen efectos en sistemas externos.

## Delivery

Entrega `hold`, vigente en esta sesión: sin commit ni publicación; el resultado es el working tree verificado sobre `docs/inspiration-credits-index`. Revisión de código: Claude Code `/code-review` sobre el candidato, antes de la parada de validación (respuesta Review A, 2026-11-22). Esta entrega no autoriza push, PR, merge ni refrescar la instalación local. Cuando se instale, el binario se reconstruye antes de desplegar el contenido, porque un binario anterior no conoce el manifiesto de dominios.

## Review status and progress

2026-11-22 — plan escrito como `principles-into-hive-core`; ampliado y renombrado a `hive-core-domains` por D5 B.
2026-11-22 — revisión de `hive-review-plan` sobre la versión de solo principios: 1 P1 (el manager solo instala cuatro carpetas fijas), 3 P2 (recuperación con binario viejo, guarda de nombres retirados, documentos y CHANGELOG sin tarea) y 5 P3 (conteo, enlaces en tres principios, descripción de `hive-core`, patrón de `index.md`, nombre de escenario). Todos aplicados en esta versión; D6 decide el mecanismo del manager.
2026-11-22 — segunda ronda, dos revisores. Servidor: 0 P1, 5 P2 (AC7 sin prueba, manifiesto no validado en una release guardada, test de la CLI con ruta por segmentos, métrica de lecturas del auditor, AC3 inalcanzable) y 3 P3 (`git diff -M --stat HEAD`, reservar `agents`, estado registrado y recuperación). Guía: 0 P1, 2 P2 (reescritura de enlaces de principios, tope de 16.000 bytes de las referencias propias) y 4 P3 (registros fechados, guarda de texto plano, conteo de carpetas, clasificación del registro de lecturas). Todos aplicados.
2026-11-22 — build: D9 A cambia T4, que mueve el límite de tamaño de los paquetes a `tooling/management/stage_packet_margin_test.go`; no cambia ningún criterio.
