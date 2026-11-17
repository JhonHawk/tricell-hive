# Tasks

Base: `JhonHawk/tricell-hive`, `development` en `582997ab`; rama `feat/create-prd`.

T1 a T7 y T9 corren en el hilo principal: los textos de la skill comparten un vocabulario que T1 fija, y cada tarea edita como mucho tres archivos. T8 va a un hijo `hive-write-docs`, porque edita diez archivos, siete de ellos del sitio, y sus fuentes ya están escritas cuando empieza.

### T1 — Entrada de la skill y formato del PRD

- [x] `content/skills/create-prd/SKILL.md` y `references/prd-format.md` existen, con los nueve pasos, las secciones, la línea de regla, los estados, la lista de comprobación y las versiones.

**Closes:** AC2, AC3, AC5, AC11, AC25, AC26, AC29, AC30

**Depends on:** none

**Locations:** `content/skills/create-prd/SKILL.md`, `content/skills/create-prd/references/prd-format.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Escribe los pasos de `design.md` «Pasos de `SKILL.md`» y, en la referencia, las secciones del PRD, la línea de regla con sus estados y veredictos, la lista fija de comprobación, los estados del documento con lo que permite cada paso, y el historial de versiones con su lista de cambios. Antes de escribir, lee `_support/docs/harness-engineering/model-floor.md` y `_support/docs/architecture/instruction-resources.md`. No cambia `create-issue` ni la guía global.

**Verification:** `go test ./tests/content/...` pasa (frontmatter, tamaño ≤ 16.000 bytes y enlaces); `wc -c` de `SKILL.md` da 9.000 bytes o menos; `rg -n "at most three questions" content/skills/create-prd/SKILL.md` encuentra la línea. AC2, AC3, AC5, AC11, AC25, AC26, AC29 y AC30 se observan en la instalación temporal del candidato, que alcanza la verificación antes del push.

### T2 — Documento que llega

- [x] `references/incoming-document.md` dice cómo estructurar un documento ya escrito sin cambiar sus reglas y cómo se arma el cuestionario para su autor.

**Closes:** AC4

**Depends on:** T1

**Locations:** `content/skills/create-prd/references/incoming-document.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Lee el documento completo; pasa cada regla a una línea con su fuente apuntando a la sección o página del original; no corrige ni completa una regla del autor; marca `to decide` lo ambiguo, lo contradictorio y lo que falta; arma el cuestionario con una respuesta recomendada por pregunta, ordenado por cuánto cambia el documento; el original nunca se edita. El resultado empieza en `iterating`, nunca aprobado.

**Verification:** `go test ./tests/content/...` pasa; `rg -n "never edit" content/skills/create-prd/references/incoming-document.md` encuentra la regla del original. AC4 se observa en la instalación temporal del candidato.

### T3 — Dónde vive el documento y cómo se congela

- [x] `references/documents.md` cubre la herramienta de documentos, las operaciones por herramienta, la copia congelada y el modo `.md` con su ubicación y su PDF.

**Closes:** AC8, AC18, AC19, AC20, AC31

**Depends on:** T1, T4

**Locations:** `content/skills/create-prd/references/documents.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Nombra la capacidad («una herramienta de documentos en la que la sesión pueda escribir») y, en una tabla, la operación de Notion, Google Docs y Claude Docs para crear, copiar, editar, comentar y obtener el enlace, con los límites medidos en `proposal.md` «Verified context». Fija la copia congelada: título con versión, estado y fecha, nunca se vuelve a escribir, y el historial del documento vivo la enlaza. Fija el modo `.md`: `_support/docs/requirements/<módulo>/prd.md` dentro de un repositorio y la carpeta de trabajo fuera de uno, con `prd-v<N>.md` y su PDF generado con `scripts/render_pdf.py`. Antes de la primera escritura en la herramienta, lee qué puede escribir la cuenta y fija el destino como dice `design.md` «Operaciones por herramienta». Dice que las filas de Google Docs vienen de su documentación y están sin probar, que el archivo del repositorio se escribe sin commit, como dice `design.md` «Archivo en el repositorio», que cada entrega del documento vivo lleva su PDF al lado, y que el PDF es un derivado que no hace falta versionar.

**Verification:** `go test ./tests/content/...` pasa; `rg -n -F "docs/requirements" content/skills/create-prd/references/documents.md` encuentra la ruta. AC8, AC18, AC19, AC20 y AC31 se observan en la instalación temporal del candidato.

### T4 — Script de PDF

- [x] `scripts/render_pdf.py` convierte un `.md` de PRD en PDF con el navegador instalado, y sin navegador deja el HTML y sale con error.

**Closes:** AC21, AC22, AC28

**Depends on:** none

**Locations:** `content/skills/create-prd/scripts/render_pdf.py`, `tests/skills/render_pdf_test.py`

**Execution:** main thread

**Test approach:** tdd

**Changes:** Antes de escribir, lee `content/skills/flow-build/references/security-boundaries.md`: el script recibe texto que no es de confianza y ejecuta un proceso. Implementa el contrato de `design.md` «Script de PDF» solo con la biblioteca estándar de Python. El archivo de pruebas fija `sys.dont_write_bytecode = True`, como las demás pruebas de `tests/skills/`, para no dejar `__pycache__` en la carpeta de la skill. Las pruebas cubren: cada construcción del subconjunto de Markdown; que una construcción desconocida sale como texto literal; que todo el texto del `.md` está en el HTML; que un `<script>` y un `<img src="http…">` del `.md` salen escapados como texto y que la página lleva la política que prohíbe scripts y cargas de red; que el navegador se lanza con un directorio de perfil temporal; el código de salida, el mensaje y el `.html` que queda cuando no hay navegador (con `--browser` apuntando a una ruta que no existe); y, donde hay navegador, que el PDF existe y tiene páginas A4. No agrega dependencias ni activos fuera de `scripts/`.

**Verification:** `python3 -m unittest discover -s tests/skills -p 'render_pdf_test.py'` pasa; ejecutar el script sobre un `.md` de ejemplo en la carpeta temporal produce un PDF cuyas páginas miden 595 × 842 puntos y cuyo texto, leído con `pdftotext`, contiene cada línea de regla del `.md`. AC21, AC22 y AC28 se observan con el script de la instalación temporal del candidato.

### T5 — Revisión técnica

- [x] `references/technical-review.md` dice cómo se contrasta cada regla con el código y qué deja el documento «listo para desarrollo».

**Closes:** AC15, AC16, AC17

**Depends on:** T1

**Locations:** `content/skills/create-prd/references/technical-review.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Solo en una sesión con el repositorio y sobre una versión aprobada por negocio. Da un veredicto a cada regla, con las ubicaciones en la sección de revisión técnica del documento; delega la lectura a hijos `hive-research` cuando pasa los límites de delegación. Quien revisa y no es el autor no reescribe ninguna regla: escribe veredictos y, por cada `contradicts` o `blocked`, una pregunta con respuesta recomendada para el dueño, en un comentario donde la herramienta lo permite. El resultado es una versión nueva; pasa a «listo para desarrollo» solo sin `not checked` y con cada `contradicts` y `blocked` respondido.

**Verification:** `go test ./tests/content/...` pasa; `rg -n "not checked" content/skills/create-prd/references/technical-review.md` encuentra la condición. AC15, AC16 y AC17 se observan en la instalación temporal del candidato.

### T6 — Rol `hive-review-prd`

- [x] `content/agents/review/hive-review-prd.md` existe y el paso 6 de la skill lo lanza con el encargo de un revisor.

**Closes:** AC23, AC24

**Depends on:** T1

**Locations:** `content/agents/review/hive-review-prd.md`, `content/skills/create-prd/SKILL.md`, `_support/docs/architecture/agent-delivery.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Escribe el rol de `design.md` «Rol `hive-review-prd`» con `model_profile: reasoning`, `access_profile: observe` y `effort: medium`, sus enlaces a `prd-format.md` en la forma `skill:create-prd/references/prd-format.md`, y el párrafo de límites que llevan los demás roles. El paso 6 de la skill enuncia los tres insumos del encargo (el texto actual del documento o la ruta de un `.md` local con él, nunca solo su enlace; su fuente; y el resultado de la lista de comprobación) y remite a la regla del encargo de un revisor en `flow-build/references/delegation.md` para lo que no se incluye. En `agent-delivery.md`, el párrafo que abre con el número de roles (línea 9, que hoy dice veinte con 19 archivos) queda en veinte y nombra la revisión de requerimientos por `hive-review-prd`. No cambia `integrations/agent-profiles.json` ni el cargador de OpenCode.

**Verification:** `go test ./tests/content/... ./integrations/agents/...` pasa. AC23 y AC24 se observan en el rol y la skill de la instalación temporal del candidato.

### T7 — Entrada desde research, hogar canónico y revisión del comportamiento completo

- [x] Los pasos 0 y 5 de `flow-research` envían a la skill las peticiones de un requerimiento, la lista de hogares canónicos nombra la skill, y queda escrita la lista de lugares revisados que pide el `AGENTS.md` del proyecto.

**Closes:** AC13, AC27

**Depends on:** T1, T2, T3, T5, T6

**Locations:** `content/skills/flow-research/SKILL.md`, `_support/docs/architecture/instruction-resources.md`, `_support/openspec/changes/484-create-prd/tasks.md`

**Execution:** main thread; la búsqueda de lugares va a un hijo `hive-research`, porque pasa de cinco lecturas grandes

**Test approach:** check

**Changes:** En `flow-research` agrega la frase en los dos lugares donde está la de Figma: en el paso 0, una petición cuyo entregable es un documento de requerimientos se atiende siguiendo `create-prd` en el mismo turno, que hace sus propias preguntas; en el paso 5, lo mismo cuando el paso 3 ya hizo preguntas. La frase dice «sigue la skill», no «inicia un flujo», y nada más de `flow-research` cambia. Agrega la viñeta de `create-prd` junto a la de `create-issue`. Busca en la guía global, las skills con sus referencias, los roles, `site/`, `_support/docs/architecture/` y `tests/pilot/` cada lugar que enuncie o dependa de escribir o validar un requerimiento; confirma que no queda una instrucción duplicada o en conflicto y anota los lugares revisados en la sección de progreso, para el pull request. La guía global no cambia.

**Verification:** `go vet ./... && go test ./...` pasa; `rg -n "create-prd" content/skills/flow-research/SKILL.md _support/docs/architecture/instruction-resources.md` encuentra una línea en cada uno; `wc -c content/skills/flow-research/SKILL.md` da menos de 16.000; AC27 se observa en la instalación temporal del candidato. La comprobación de que la guía global no cambiaba se retiró con D33 A; T10 la cambia.

### T8 — Sitio y listas

- [x] El sitio tiene la página de `create-prd`, enlazada desde la barra lateral y las listas, y `llms.txt`, `CHANGELOG.md` y `CONTRIBUTING.md` la nombran.

**Closes:** AC13

**Depends on:** T1, T2, T3, T4, T5, T6, T7

**Locations:** `site/src/content/docs/flows/create-prd.md`, `site/astro.config.mjs`, `site/src/content/docs/flows/overview.md`, `site/src/content/docs/flows/research.md`, `site/src/content/docs/skills.md`, `site/src/content/docs/index.mdx`, `site/src/content/docs/getting-started.md`, `llms.txt`, `CHANGELOG.md`, `CONTRIBUTING.md`

**Execution:** hive-write-docs

**Test approach:** check

**Changes:** La página describe el comportamiento y enlaza la skill, sin copiar sus reglas, con la misma estructura que `flows/create-issue.md`. Las listas reciben una entrada cada una, y la página de research dice que una petición de un requerimiento sigue la skill. `README.md` no cambia: no lista skills.

**Verification:** `npm run build` en `site/` termina sin errores y genera `flows/create-prd/index.html`; `rg -n "create-prd" llms.txt CHANGELOG.md site/astro.config.mjs` encuentra una línea en cada uno.

### T10 — La guía global nombra la continuación de research a `create-prd`

- [x] La lista de continuaciones en el mismo turno de la guía global incluye la de research a `create-prd`.

**Closes:** AC32

**Depends on:** T7

**Locations:** `content/guidance/global.md`

**Execution:** main thread

**Test approach:** check

**Changes:** En «Questions and routes», la frase de las continuaciones pasa de cuatro a cinco y nombra la de research a `create-prd` para una petición cuyo entregable es un documento de requerimientos, junto a la de Figma. Ninguna otra frase de la guía global cambia; el pull request declara el cambio de tamaño.

**Verification:** `go test ./tests/content/... ./tooling/management/...` pasa; `rg -n "five continuations" content/guidance/global.md` encuentra la línea; `rg -n -i "four continuations" .` fuera de los archivos archivados no encuentra nada. AC32 se observa en la instalación temporal del candidato.

### T9 — Sesión real en Notion

- [ ] Tras integrar y refrescar la instalación, la skill instalada deja en Notion el PRD de `create-architecture` y su versión 1 congelada.

**Closes:** AC14

**Depends on:** T1, T2, T3, T4, T5, T6, T7, T8

**Locations:** ninguna en el repositorio

**Execution:** main thread

**Test approach:** check

**Changes:** Sigue la skill instalada de principio a fin: redacta el PRD desde las decisiones ya registradas de `create-architecture` (memoria del proyecto, tema `decision/create-architecture-skill`), corre el revisor, pide la aprobación del usuario y congela la versión 1. Solo con la clase de efecto de `proposal.md` «Effects» aprobada; antes de cada escritura anota aquí el destino, el cambio y cómo se revierte. No crea tickets.

**Verification:** Releer las dos páginas: la viva tiene las nueve secciones, cada regla con su línea completa y una fila de historial que enlaza la copia; la copia lleva versión, estado y fecha en el título. El cierre del trabajo y el archivo de esta carpeta esperan a T9: el ticket se mueve al integrar, pero la carpeta no se archiva con T9 abierta, y un defecto que T9 encuentre se corrige como una revisión de este mismo cambio, en un pull request nuevo.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | El comando de la tarea pasa | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./... && go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Site build | Final candidate, locally | `npm run build` en `site/` | Covered by the implementation |
| Installed content | Final candidate, locally, antes de la parada de validación | `hive plan install --scope user --hosts claude --home <tmp> --state-dir <tmp> --source <worktree> --out <tmp>/plan.json` y `hive apply --plan <tmp>/plan.json` en la carpeta temporal de la sesión; cada archivo de los criterios leído ahí, y el script de esa instalación produce el PDF. Si el instalador no acepta un directorio temporal vacío, estos criterios quedan sin verificar hasta el refresco posterior a la integración, y se dice | Covered by the implementation |
| Real session | Tras la instalación | T9 | La clase de efecto de Notion |
| Code review | Final candidate, before the validation stop | `/code-review` de Claude Code sobre el candidato | The review answer (D28 A) |
| Human validation | Tras la revisión, antes del push | El usuario lee la skill instalable y prueba el PDF | La respuesta de entrega (D27 A) |

Sin cambio de pantalla: no hay recorrido de interfaz ni revisión de experiencia.

Parada de validación: el usuario lee `content/skills/create-prd/SKILL.md` y sus cuatro referencias en la rama, y ejecuta `python3 content/skills/create-prd/scripts/render_pdf.py <un .md>` para abrir el PDF que produce.

## Checkpoint

- Updated: 2026-11-16
- Stage: verify · recorrido de criterios hecho en la instalación temporal; falta /code-review y la parada de validación
- Last commit: e727d326
- Open findings: none
- Next: /code-review, ronda de correcciones, parada de validación con D33
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/create-prd · feat/create-prd

### Evidence

- go vet ./... && go test ./... · pasa · cf758070
- python3 -m unittest discover -s tests/skills -p '*_test.py' · 99 pruebas pasan · cf758070
- go vet ./... && go test ./... · pasa · 930236d2
- python3 -m unittest discover -s tests/skills -p '*_test.py' · 99 pruebas pasan · 930236d2
- npm run build en site/ (node_modules del checkout principal enlazado) · 20 páginas, flows/create-prd/index.html generado · 930236d2
- instalación temporal del candidato (hive plan install --scope user --hosts claude --home TMP … + hive apply) · AC2-AC5, AC8, AC11, AC15-AC20, AC23-AC27, AC29-AC31 observados en los archivos instalados · 930236d2
- script instalado sobre un .md de ejemplo · PDF A4 (595x842), pdftotext contiene cada regla; --browser inexistente sale 3 y deja el .html; <script> e imagen remota salen como texto · 930236d2

## Review status and progress

2026-11-16 — Plan escrito.

2026-11-16 — Dos lecturas de `hive-review-plan` aplicadas antes de construir; el detalle está en `proposal.md` «Review status and progress». Nada construido se invalida.
2026-11-16 — T4 · 3fd7d41f · unverified: AC21, AC22 y AC28 en la instalación temporal del candidato (verificación)
2026-11-16 — T1 · aa245644 · unverified: AC2, AC3, AC5, AC11, AC25, AC26, AC29 y AC30 en la instalación temporal del candidato (verificación)
2026-11-16 — T2 · aa245644 · unverified: AC4 en la instalación temporal del candidato (verificación)
2026-11-16 — T3 · aa245644 · unverified: AC8, AC18, AC19, AC20 y AC31 en la instalación temporal del candidato (verificación)
2026-11-16 — T5 · aa245644 · unverified: AC15, AC16 y AC17 en la instalación temporal del candidato (verificación)
2026-11-16 — T6 · cf758070 · unverified: AC23 y AC24 en la instalación temporal del candidato (verificación)
2026-11-16 — T7 · adc8334f · unverified: AC27 en la instalación temporal del candidato (verificación); queda una decisión del usuario: si la guía global nombra el paso de research a create-prd
2026-11-16 — T8 · 9a88a42a · unverified: none

2026-11-16 — T7, lugares revisados por el barrido del comportamiento completo (hijo `hive-research`), para el pull request: `content/guidance/global.md` («Proportionality», «Questions and routes», «Scope and authorization», «Support folder», «Delivery»); `content/skills/**` con sus referencias (`create-issue`, `flow-research`, `flow-plan/references/business-rules.md`, `flow-close/references/artifact-layout.md`, `flow-build/references/delegation.md`, `flow-research/references/code-audit.md`); `content/agents/**`; `site/src/content/docs/**`; `_support/docs/architecture/**`, `README.md`, `llms.txt`, `AGENTS.md`; `tests/pilot/**` y `tests/fixtures/**`; y la coherencia interna de la skill. Corregido en el mismo cambio: la revisión técnica confirma antes de escribir; el archivo `.md` dentro de un repositorio se pregunta antes de la primera escritura y su commit espera al usuario; una regla de un documento que llega nunca se reescribe por la lista de comprobación; `draft` e `iterating` se distinguen por la lista; las referencias remiten a las rondas del paso 3 en vez de repetir el tope. Sin cambios, a propósito: `README.md` (no lista skills), `tests/pilot/protected.go` (su lista ya omite `create-issue`). Pendiente de decisión del usuario: `content/guidance/global.md:93` lista cuatro continuaciones en el mismo turno y no nombra la de research a `create-prd`. Hallazgo anotado, fuera de este cambio: `flow-plan/references/business-rules.md:80` dice que el plan actualiza un documento de requerimientos del repositorio y no remite a las versiones de `create-prd`.

2026-11-16 — T9, adelantado en parte a petición del usuario durante la validación («abre el Notion de ejemplo»), con la skill del candidato, idéntica a la de la instalación temporal. Efecto 1 (clase aprobada, D29 A y D31 A): destino, un borrador privado en el Notion conectado; cambio, crear la página «PRD · create-architecture» en su versión 1; cómo revertir, mover la página a la papelera desde Notion. La aprobación y la copia congelada quedan para después de integrar.
2026-11-16 — T10 · 6d30656d · unverified: AC32 en la instalación temporal del candidato (verificación)
