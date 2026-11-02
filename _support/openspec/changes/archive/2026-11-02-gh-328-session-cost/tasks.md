# Tareas

Base: `tricell-hive` `development` en `5e414d77`; rama `feat/gh-328-session-cost` en `.claude/worktrees/gh-328-session-cost`. T1 y T3 son texto de la guía y corren en el hilo principal, porque exigen revisar todo el comportamiento que gobiernan (root `AGENTS.md`). T2 es código Go y pruebas, con un brief más corto que el trabajo: corre en un hijo `hive-build-backend` en su propio worktree. Las tres son independientes: T3 nombra las opciones de T2, que este plan ya fija.

### T1 — Acciones sueltas como pregunta y compactar en transiciones

- [ ] La viñeta del bloque de siguiente paso en `global.md` dice cuándo se ofrece el comando de un flujo, que una acción suelta fuera de los archivos del repositorio se pregunta en la ronda, cómo se pregunta por un ticket inexistente, y la respuesta opcional de compactar en los dos momentos.
- [ ] `plan-handoff.md` y `flow-close` («Find the next ticket») apuntan a esa regla sin repetirla; `flow-research` (líneas 10 y 24) acepta que la respuesta a una pregunta autorice su acción.
- [ ] La regla dice que una compactación en una parada de ruta termina otra vez con el mismo bloque, sin chocar con `global.md:99`, y nombra el comando por capacidad.
- [ ] Las páginas `site/src/content/docs/flows/{plan,close,status,research}.md` describen el comportamiento y enlazan la skill; CHANGELOG con la entrada.

**Closes:** AC1, AC2

**Depends on:** ninguna

**Locations:** `content/guidance/global.md`, `content/skills/flow-plan/references/plan-handoff.md`, `content/skills/flow-close/SKILL.md`, `content/skills/flow-research/SKILL.md`, `site/src/content/docs/flows/`, `_support/docs/architecture/instruction-resources.md:19,24`, `CHANGELOG.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** cambia el texto de la regla y sus punteros; no cambia qué hace `flow-build` con los cambios en archivos ni la tarjeta de cierre.

**Verification:** `go test ./tests/content/...` pasa (tamaño de skills y enlaces); `rg -n "recommend creating the ticket first|40% of the model" content site` no devuelve nada; lectura de la viñeta contra AC1 y AC2; tamaño de `global.md` antes y después.

### T2 — Contrato del visor por partes, salida compacta y error con campo

- [ ] `hive view contract` acepta `--for FLOW`, `--list` y `--section rules`, excluyentes con `--operation`, con la forma de `design.md`.
- [ ] El mapa flujo → operación vive en `tooling/viewer/timeline_protocol.go` y una prueba falla si una operación no tiene flujo o nombra uno desconocido.
- [ ] La salida de `viewTimelineOutput` es JSON compacto; `--operation` conserva su contenido.
- [ ] Un error de validación imprime `invalid timeline input: <campo>: <código>`.
- [ ] Ayuda (`main.go:385`), `README.md`, `llms.txt` (revisión), `deployment-manager.md` y una entrada de CHANGELOG describen las opciones nuevas.
- [ ] Antes de fijar el límite, medir cada `--for`; si alguno pasa de 10.000 bytes, reportarlo antes de recortar contenido.

**Closes:** AC3, AC4

**Depends on:** ninguna

**Locations:** `tooling/viewer/timeline_protocol.go`, `tooling/viewer/timeline_protocol_test.go`, `tooling/cli/view_timeline.go`, `tooling/cli/view_timeline_v3_test.go`, `tooling/cli/main.go`, `CHANGELOG.md`, `README.md`, `llms.txt`, `_support/docs/architecture/deployment-manager.md`

**Execution:** hijo `hive-build-backend` en worktree

**Test approach:** tdd

**Changes:** cambia la lectura del contrato y el formato de la salida; no cambia el protocolo v3, los esquemas, el servidor ni las operaciones de escritura.

**Verification:** `go test ./tooling/...` pasa; `hive view contract --for <cada flujo> | wc -c` < 10000; `--list` trae 28 entradas; `--operation stage_start` sin sangría y con el mismo contenido que hoy (comparado con `python3 -m json.tool`); una entrada sin `reason` en `stage start` devuelve `invalid timeline input: reason: invalid_text`.

### T3 — Reglas de lectura única del visor

- [ ] `flow-walkthrough/SKILL.md` y `viewer-publication.md` fijan la lectura única de la referencia, el contrato por flujo (qué resumen lee cada skill), el uso de los acuses en vez de `session get` y los momentos de relectura; los anclas que otras skills enlazan se conservan.
- [ ] `site/src/content/docs/viewer.md` revisada; el delta de `walkthrough-viewer` queda listo para aplicar al cerrar.

**Closes:** AC5

**Depends on:** ninguna

**Locations:** `content/skills/flow-walkthrough/SKILL.md`, `content/skills/flow-walkthrough/references/viewer-publication.md`, `site/src/content/docs/viewer.md`, `_support/openspec/changes/gh-328-session-cost/specs/walkthrough-viewer/spec.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** cambia cuándo se lee y relee; no cambia qué se publica ni el orden de las etapas.

**Verification:** `go test ./tests/content/...` pasa; `rg -n "at each stage change" content/skills/flow-walkthrough` no devuelve nada; los anclas `#open-the-viewer`, `#register-once-reference-by-id` y `#close-and-input-hygiene` siguen existiendo; lectura contra AC5.

## Verificación y revisión humana

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar cada tarea, local | El comando de su verificación pasa | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go test ./...` | Cubierto por la implementación |
| Revisión de la guía | Candidato final | Lista de lugares revisados y tamaño de `global.md` en el PR | Cubierto por la implementación |
| Revisión de código | Candidato final, antes del push | Claude Code `/code-review` sobre el diff | La respuesta de revisión (D11 A) |

Sin recorrido de UI: el cambio no toca pantallas.

## Revisión y avance

2026-11-02 — Plan escrito.
2026-11-02 — T1 hecha en `e90df24f`: regla en `global.md:66` (22.263 → 23.577 bytes), punteros, sitio. `tests/content` pasa. CHANGELOG pendiente de commit con T2.
2026-11-02 — T3 hecha en `412cf79a`: lectura única en `flow-walkthrough` y su referencia; anclas intactas; `site/viewer.md` no describe la cadencia de lectura, sin cambios. La frase «at each stage change» sigue, ahora para cuándo se sigue el procedimiento, no para cuándo se lee.

2026-11-02 — T2 hecha en `44da6248` por un hijo `hive-build-backend`; `hive-verify-task` dio AC3 y AC4 cumplidos y un defecto menor (`--for=`, `--section=` y `--list=false` imprimían el contrato completo), corregido en el hilo principal con su prueba. Tamaños de `--for`: research 6.241, plan 6.237, build 8.742, verify 8.415, delivery 9.532, close 1.475, collab 2.237 bytes; `--operation` con el mismo contenido que el binario instalado. Suite local (`go vet`, `go test ./...`, pruebas Python) en verde; una corrida previa falló en `tooling/cli` sin nombre de prueba y no se repitió en dos corridas.

2026-11-02 — Verify: AC1–AC5 comprobados sobre `44da6248`. `/code-review` (medium, `development...HEAD`) dio 10 hallazgos. Bloqueantes corregidos en `06057b8c`: H10 `const` ignorado en los valores permitidos; H11 `reason` marcada opcional en `stage_start` y `stage_update` (nueva prueba contra los validadores); H12 obligatorios anidados ausentes; H13 «closing a ticket» en conflicto con la escritura ya autorizada al integrar. No bloqueantes, al cierre: H14 `collab` incluye `session_delete` y `recover_lead`, que solo usa el visor local; H15 enums bajo `oneOf`/`allOf` no se recorren; H16 el error local (`campo: código`) y el del servidor (`código on campo`) usan órdenes distintos; H17 la salida compacta aplica a todo `hive view`, como decidió el diseño; H18 `List`/`Rules` ignoran errores y cada `--for` arma el contrato completo; H19 la prueba del CLI fija 28 operaciones. Segunda revisión (`44da6248..06057b8c`): dos bloqueantes introducidos por las correcciones, arreglados en `61c1bdf9` con su prueba y la línea `Unreviewed`: H20 una `const` entera salía como el texto «3»; H21 `reason` solo era obligatoria en el resumen y no en el contrato completo (ahora el esquema la exige). Descartado tras probarlo: que los obligatorios anidados estén sobredeclarados (quitar cada uno del ejemplo lo rechaza el validador). No bloqueantes, al cierre: H22 la prueba de etiquetas salta en silencio las operaciones cuyo ejemplo no valida; H23 los obligatorios de elementos de lista (`segments`, `tabs`) no se listan; H24 un nombre requerido sin propiedad se descarta sin error; H25 `event_patch` exige `reason` solo cuando cambia `stage_state` y el resumen no lo dice; H26 los valores de ramas `anyOf`/`oneOf` no se combinan (el `kind` de `actor` no aparece); H27 la excepción `_patch` de la prueba es heurística. La segunda revisión no cubre lo que la primera no vio. Suite local en verde sobre `61c1bdf9`.

2026-11-02 — Integrado por el PR #329 (`36ac21ee`); #328 cerrado. Sin CI en el repositorio; sin revisiones en el PR. Delta aplicado a `specs/walkthrough-viewer/spec.md` y carpeta archivada.
