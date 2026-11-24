# Tasks

Base `development` en `cf490341`, rama `fix/public-tree-private-identifiers` en el checkout principal. T1, T2 y T4 corren en el hilo principal: el hilo principal tiene el archivo local con los términos y el mapa, y T2 y T4 tocan pocos archivos. T3 edita más de 200 archivos con un script y corre en un hijo `hive-build-backend`, que recibe la ruta del archivo local.

### T1 — Test `TestNoPrivateIdentifiers` y la lista local

- [x] El test falla en la base por los identificadores de hoy, sin imprimir ningún término privado

**Closes:** AC2, AC3

**Depends on:** none

**Locations:** `tests/content/private_identifiers_test.go` (nuevo), `_support/workspace/local-context/private-identifiers.local` (local, ignorado; no se commitea)

**Execution:** main thread

**Test approach:** tdd

**Changes:**
- **Archivo local:** una línea por término, `término literal<TAB>reemplazo`.
  - Se compara sin distinguir mayúsculas y como palabra entera.
  - Una línea `prefix:<PREFIJO><TAB><reemplazo>` cubre un prefijo de ticket, con o sin guion, seguido de dígitos. El número se conserva.
  - Incluye el nombre de usuario local, sin la ruta, con reemplazo `me`. Así cubre también `-Users-<usuario>-` y `/home/<usuario>/`.
  - Al escribirlo, el progreso registra el recuento de archivos por número de línea, nunca el término.
- **Ubicación de la lista:** el test la busca en el checkout principal, a partir de `git rev-parse --git-common-dir`, como hace `tests/content/audit_comments_template_test.go:69`. Así un worktree también la encuentra.
- **Qué recorre:** `git -C <toplevel> ls-files -z --cached --others --exclude-standard`, con el top level de `git rev-parse --show-toplevel`. Revisa el contenido y también la ruta de cada archivo, saltando binarios y su propio archivo.
- **Rutas personales:** falla por cada `/Users/[A-Za-z0-9._-]+/` cuyo nombre no esté en la lista commiteada de ejemplos: `alice`, `someone`, `test`, `me`, `example`, `user`, `you`, `x`, `u`.
- **Términos de la lista:** con el archivo local presente, falla también por cada término.
  - El mensaje nombra `archivo:línea`, el número de línea de la regla y la corrección («reemplázalo por el ejemplo neutro que da esa línea del archivo local»), nunca el término.
  - Una línea malformada del archivo falla con su número de línea, sin repetir su contenido.
- **Sin el archivo local:** registra con `t.Log` que saltó esa parte y no falla por ella.
- **Función de escaneo:** recibe la raíz, la lista de archivos y los términos. El caso de prueba propio no lee la lista real; usa un término inventado y comprueba tres cosas:
  - que un archivo con ese término falla, con archivo, línea y número de regla;
  - que el mensaje no contiene el término;
  - que una palabra que solo contiene el término pasa.

**Verification:** `go test -count=1 -run TestNoPrivateIdentifiers ./tests/content/` falla en la base por rutas y términos, y su caso propio pasa

### T2 — Guía distribuida y páginas públicas

- [x] `content/`, `README.md` y `site/` no traen identificadores de clientes

**Closes:** AC4

**Depends on:** T1

**Locations:** `content/skills/hive-mode/playbooks/ticket-status.md`, `content/skills/audit-comments/assets/ratchet/check-comments.test.mjs` y el tercer archivo de `content/skills` que el test de T1 señale, `README.md`, `site/src/content/docs/flows/hive-mode.md`

**Execution:** main thread

**Test approach:** check

**Changes:** El ejemplo de estado de tickets usa claves y títulos neutros (`ABC-123`, «Exportar facturas»). El ejemplo de comando del README usa `ABC-123`. Los demás reciben el reemplazo del mapa. La forma de los ejemplos no cambia.

**Verification:** la comprobación de texto de root `AGENTS.md` «Checks by path» pasa; `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa (ruta `content/skills/**/assets/**`); `pnpm --dir site check`, `lint`, `build` pasan

### T3 — Reemplazo en el resto del árbol

- [x] El test de T1 pasa con la lista local presente

**Closes:** AC1

**Depends on:** T1, T2

**Locations:** `_support/openspec/changes/archive/`, `_support/openspec/specs/`, carpetas de cambio activas, `_support/docs/`, `_support/sessions/`, `CHANGELOG.md`, `tests/skills/`, `tests/pilot/`, `tooling/cli/`, `tooling/viewer/frontend/src/`, `tooling/viewer/testdata/`, `tooling/viewer/registry_test.go` y el bundle regenerado `tooling/viewer/static/app/`. El recuento por carpeta lo da el test de T1.

**Execution:** hive-build-backend

**Test approach:** check

**Changes:**
- Un script de un solo uso, fuera del repositorio, aplica el mapa del archivo local, por palabra entera y sin distinguir mayúsculas.
- Cambia en la misma pasada el spec, los deltas archivados y los deltas activos. Dos encabezados de requisito llevan un término, y así `TestSpecDeltasApplied` sigue en verde.
- No toca IDs de sesión ni los proyectos propios.
- El bundle no se edita a mano. Orden en el visor:
  1. reemplazar las fuentes;
  2. `pnpm --dir tooling/viewer/frontend install --frozen-lockfile`;
  3. `check`, `lint` y `test`;
  4. `build`;
  5. `verify:dist`;
  6. un segundo `build` y `git diff --exit-code tooling/viewer/static/app`;
  7. `go test ./tooling/viewer/...`.
- Solo debe cambiar `sourceDigest` en `bundle-manifest.json`. Un asset JS renombrado indica que cambió una fuente de producción, y detiene la tarea.

**Verification:**
- La suite completa pasa: `go vet ./...`, `go test -count=1 ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'`. Es el check que «Checks by path» asigna a `tooling/**`.
- La secuencia del visor de arriba pasa.
- El test de T1 pasa con la lista local.

### T4 — Regla y enforcer

- [x] Root `AGENTS.md` dice que el repositorio es público y qué no entra, y `rule-enforcers.md` nombra el test

**Closes:** AC2

**Depends on:** T1

**Locations:** `AGENTS.md` «Scope and preservation», `_support/docs/harness-engineering/rule-enforcers.md`, `_support/docs/harness-engineering/session-review-rubric.md:3`, `CONTRIBUTING.md` «Branches and commits»

**Execution:** main thread

**Test approach:** check

**Changes:**
- **Una línea en root `AGENTS.md`** que diga:
  - el repositorio es público;
  - los archivos versionados no traen nombres de proyectos ni claves de tickets de clientes ni la ruta o el nombre del usuario local;
  - se escriben en forma neutra: `/Users/me/`, `ABC-123`, `org-a`, `project-a`;
  - un cliente nuevo se agrega al archivo local antes de mencionarlo;
  - los IDs de sesión y los proyectos propios están permitidos;
  - lo aplican las pruebas de contenido (la comprobación de texto y la suite completa).
  
  La línea enlaza el punto de «Workspace locations» que ya gobierna `local-context/`, en vez de repetirlo. Por D14 A, la línea cubre también mensajes de commit, textos de PR y comentarios, que citan sesión, host, modelo y fecha sin proyecto ni ticket de cliente; root `AGENTS.md` «Measurement», que pide nombrar observaciones en el commit, recibe la misma precisión en su frase.
- **`rule-enforcers.md`:** dos filas. La de rutas `/Users/` queda como aplicada en cualquier checkout. La de nombres queda como aplicada solo donde existe el archivo local. La ruta del test va sin un nombre de función entre backticks, porque `rule_enforcers_test.go:92-101` lo leería como criterio.
- **Rúbrica, línea 3:** un puntero a la línea nueva, para citar el ID de sesión y no el proyecto ni el ticket.
- **`CONTRIBUTING.md` «Branches and commits»:** un puntero a la misma línea.

**Verification:** la comprobación de texto de root `AGENTS.md` «Checks by path» pasa

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar cada tarea, en el checkout | El comando de la tarea pasa | Cubierto por la implementación |
| Full suite | Candidata final | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Cubierto por la implementación |
| Recuento en el árbol (AC1) | Candidata final | `TestNoPrivateIdentifiers` con la lista local, y `git grep -l` de sus términos: 0 archivos | Cubierto por la implementación |
| Code review | Candidata final, antes del push | `/code-review` de Claude Code | La respuesta de revisión |
| Security review | Candidata final, antes del push | `hive-review-security` | La respuesta de seguridad |
| Guía instalada (AC4) | Tras el merge y `hive update` | Leer `~/.agents/skills/hive-mode/playbooks/ticket-status.md` | La entrega `automatic` y la actualización de mantenedor |

## Checkpoint

- Updated: 2026-11-24
- Stage: flow-verify · 6 · push y PR
- Last commit: f47696cb
- Open findings: H12 · El test no revisa el destino de un symlink versionado (preexistente; hoy no hay symlinks) · al cierre; H13 · La segunda alternativa de prefijo distingue mayúsculas según cómo se escriba en la lista · al cierre; H14 · Prefijo corto con espacio o _ puede chocar con palabras comunes; no ocurre con la lista actual · al cierre; H15 · Con -run/-skip que excluyan el subtest, los nombres no se revisan sin aviso · al cierre; H16 · Plurales y uniones camelCase del término no se detectan · al cierre; H17 · Nombre del mantenedor en tooling/management/voice_test.go e IP de red local en un registro archivado · al cierre
- Next: push, PR, merge, hive update, AC4, flow-close; luego retomar D10
- Rules: flow-build/stage.md, flow-verify/stage-entry.md · extra: security-boundaries
- Principles: none

### Evidence

- before captures: none (no screen)
- go vet ./... · exit 0 · a17d57b4
- go test -count=1 ./... · exit 0 · a17d57b4
- python unittest tests/skills · OK 416 · a17d57b4
- texto AGENTS.md Checks by path · exit 0 · a17d57b4 (tras T4)
- pil_scan · files 0 lines 0; TestNoPrivateIdentifiers ok · 4700a9ba
- UUIDs intactos en el diff de T3 (0 quitados, 0 añadidos)
- visor: install/check/lint/test/build/verify:dist/2º build sin diff; solo sourceDigest cambió (hijo T3)
- local CI: sin jobs de tests (master-source-branch.yml, CI-only)
- AC1: TestNoPrivateIdentifiers ok con lista; escaneo del revisor de seguridad: 0 formas restantes · f47696cb
- AC2: sonda sin trackear falla con archivo:línea y regla, término no impreso · 6150b566
- AC3: sin lista pasa y salta la parte de nombres · 6150b566
- go test ./... ok, go vet ok · e16a16d9; tests/content ok · f47696cb
- /code-review 1 (4 important) y hive-review-security (1 P1, 1 P2) arreglados en e16a16d9; /code-review 2 (3 important) arreglados en f47696cb (Unreviewed)

## Review status and progress

2026-11-24 — Plan escrito.
2026-11-24 — T1 · e42503bd · unverified: AC1 hasta T3
2026-11-24 — T2 · 72162233 · unverified: AC4 tras el merge y hive update
2026-11-24 — T3 · 4700a9ba · unverified: none
2026-11-24 — T4 · a17d57b4 · unverified: none
2026-11-24 — Verificación: AC1–AC3 observados. `/code-review` y `hive-review-security` encontraron formas que la primera pasada no veía (minúsculas, `_`, espacio, prefijo suelto) y que el test imprimía términos de rutas; corregido en e16a16d9. La corrección de claves de 6150b566 rompió un `<proyecto>-<hash>`, restaurado. La segunda revisión encontró tres fallos en esos arreglos, corregidos en f47696cb con `Unreviewed:`. Los menores quedan como H12–H17.
