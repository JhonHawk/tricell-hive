# public-tree-private-identifiers — Sacar identificadores privados del árbol público y bloquearlos

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build |
| Tracker | ninguno · decisión `decision/public-repo-client-identifiers` (Engram) |
| Git | tricell-hive · `development` en `cf490341` · rama propuesta `fix/public-tree-private-identifiers`, en el checkout principal |
| Playbook | feature |

## Decisiones

2026-11-24
- **D11 A** — ¿Qué hacemos con los datos de clientes en el repositorio público? Reemplazarlos en el árbol actual por ejemplos neutros y bloquear que vuelvan; el historial de git no se reescribe.
- **D13 A** — ¿Cómo se bloquea sin publicar los nombres? Un test en `tests/content` con dos partes: las rutas `/Users/<nombre>/` fuera de una lista de ejemplos se comprueban con un patrón commiteado, y los nombres y prefijos de clientes se leen de un archivo local ignorado por git; sin ese archivo, esa parte se salta.
- **D14 A** — ¿La regla cubre también mensajes de commit y textos de PR? Sí: commits, PRs y comentarios citan sesión, host, modelo y fecha, sin proyecto ni ticket de cliente; queda solo en prosa, sin chequeo automático.
- **Entrega** — `automatic` y `/code-review` de Claude Code, las respuestas vigentes en esta sesión.
- **Supuesto** — Cada nombre de cliente se reemplaza por un nombre neutro fijo (`org-a`, `project-a`…) y cada prefijo de ticket de cliente por uno neutro (`PRJ-`, `PRK-`…), conservando el número; el mapa vive en el archivo local. Descartado: borrar las menciones, que rompería los ejemplos y los tests que las usan.
- **Supuesto** — La ruta `/Users/<usuario real>/` pasa a `/Users/me/`, que entra en la lista de ejemplos permitidos. Descartado: `~/`, que cambia la forma de rutas que algunos tests comparan.
- **Supuesto** — Los IDs de sesión se conservan: son aleatorios, no identifican a un cliente, y root `AGENTS.md` «Measurement» exige citarlos como evidencia. Descartado: reemplazarlos, que rompería esas citas.
- **Supuesto** — Los proyectos propios de Tricell (de la organización dueña del repositorio) quedan fuera; solo cuentan los de clientes. Descartado: incluirlos, que no expone a terceros.
- **Supuesto** — El archivo de cambios entra en el reemplazo, aunque D10 podría retirarlo después: el reemplazo es un script y no depende de esa decisión. Descartado: esperar a D10, que deja publicado el material mientras tanto.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Solicitud

D11 A (2026-11-24): «Limpiar el árbol y bloquear que vuelva a pasar: reemplazar nombres de clientes, tickets y rutas por ejemplos neutros, empezando por `content/` y el README, que se distribuyen; y añadir una regla de gitleaks o un test que rechace esos patrones en archivos versionados». D13 A eligió el test con lista local.

## Objetivo

Hoy el repositorio es público y su árbol tiene nombres de proyectos de clientes, claves de tickets de clientes con títulos reales (también en la guía distribuida `content/skills/hive-mode/playbooks/ticket-status.md`) y la ruta local del usuario. Ninguna comprobación lo impide. Al terminar, el árbol no tiene esos identificadores y la comprobación de texto falla si un archivo versionado o nuevo los vuelve a traer.

## Alcance y aceptación

Dentro: los ~235 archivos del árbol que hoy los contienen (la mayoría del archivo de cambios, ~16 de `tooling/viewer`, el resto en `content/`, `README.md`, `site/`, `CHANGELOG.md`, `_support/docs`, `_support/sessions`, specs, carpetas de cambio activas y tests), el test nuevo, el archivo local con la lista y el mapa, la regeneración del bundle del visor, una línea en root `AGENTS.md` y una fila en `rule-enforcers.md`. Fuera: el historial de git, los cuerpos de PRs y commits ya publicados, la memoria local, los IDs de sesión y los proyectos propios.

- AC1 — Con la lista local presente, el test nuevo no encuentra ningún nombre ni prefijo de cliente ni el nombre del usuario local, en ninguna forma, ni ruta `/Users/<nombre>/` fuera de los ejemplos, en el contenido ni en la ruta de los archivos versionados y nuevos no ignorados · false today because `git grep` de los términos de la lista da 235 archivos en `cf490341` · Observed in: `go test -count=1 -run TestNoPrivateIdentifiers ./tests/content/` en la candidata y el mismo `git grep` con 0 archivos
- AC2 — Un archivo nuevo, no ignorado, con un nombre de cliente o una ruta `/Users/<usuario real>/` hace fallar la comprobación de texto con el archivo, la línea y la corrección, sin imprimir el término privado · false today because no existe ninguna comprobación · Observed in: la comprobación de texto con un archivo de prueba sin trackear, borrado después
- AC3 — Sin la lista local, el test sigue comprobando las rutas `/Users/` y dice que se saltó la parte de nombres · false today because no existe el test · Observed in: `go test -v -run TestNoPrivateIdentifiers ./tests/content/` con la lista movida fuera temporalmente
- AC4 — La guía instalada no trae claves ni títulos de tickets de clientes · false today because `content/skills/hive-mode/playbooks/ticket-status.md:27-35` los trae · Observed in: `~/.agents/skills/hive-mode/playbooks/ticket-status.md` tras el merge y `hive update`

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

- El repositorio es público (`gh repo view --json visibility`, 2026-11-24).
- gitleaks sobre todo el historial: 7 hallazgos, todos valores de prueba o falsos positivos (2026-11-24). Las fixtures de sesiones son sintéticas.
- Archivos afectados por carpeta, de `git grep -l` con los términos de la lista local en `cf490341`: 189 archivo de cambios, 21 `tooling/viewer`, 4 `_support/sessions`, 4 `_support/docs`, 3 specs, 3 `content/skills`, 3 carpetas de cambio activas, 2 `tooling/cli`, 2 `tests/skills`, 1 `tests/pilot`, 1 `site/src`, `README.md`, `CHANGELOG.md`.
- `_support/workspace/` está ignorado (`.gitignore:35`).
- `CONTRIBUTING.md:45-56`: un cambio en `tooling/viewer/frontend/` regenera y commitea `tooling/viewer/static/app/`; el digest incluye tests y fixtures del frontend.
- `go test -count=1 ./tests/content/...` · ok, 10 s · 2026-11-24
- `go test -count=1 ./tooling/cli/` · ok, 148 s · 2026-11-24
- `go test -count=1 ./tooling/viewer/...` · ok, 5 s · 2026-11-24
- `pnpm --dir tooling/viewer/frontend` `check`, `lint`, `test`, `verify:dist` · ok, 22 s con `install` · 2026-11-24
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` · OK, 90 s · 2026-11-24
- `go test -count=1 ./tests/pilot/` · ok, 64 s · 2026-11-24
- La comprobación de texto de root `AGENTS.md` «Checks by path» · ok, 92 s · 2026-11-24
- `pnpm --dir site check`, `lint`, `build` · ok · 2026-11-24

## Effects

None: the tasks take no effect on an external system.

## Entrega

`automatic`: revisión con `/code-review`, push, PR a `development` y merge; luego la actualización de la instalación local que autoriza `AGENTS.md` en esta copia de mantenedor. Revisión de seguridad: `hive-review-security` sobre la candidata (respuesta del 2026-11-24). No autoriza reescribir el historial ni editar PRs publicados.

## Estado de revisión y progreso

2026-11-24 — Plan escrito.
2026-11-24 — Revisión del plan: tres `hive-review-plan` (servidor, guía y cliente), ninguno con P0. Se aplicó:
- Los ejemplos permitidos suman `x` y `u`.
- La lista local usa términos literales por palabra entera, prefijos de ticket y el nombre de usuario, con un formato fijo. El test la busca en el checkout principal, recorre todo el árbol y también las rutas de archivo.
- El caso de prueba propio del test planta un término inventado.
- T2 suma los tests de Python. T3 ordena la secuencia del bundle y verifica con la suite completa.
- T4 parte la fila de `rule-enforcers.md` sin nombre de función entre backticks, suma qué forma neutra usar, qué hacer con un cliente nuevo, las excepciones, y los punteros de la rúbrica y `CONTRIBUTING.md`.
- D14 A: la regla cubre también mensajes de commit, textos de PR y comentarios.
