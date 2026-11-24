# Diseño — project-layout-skills

## Enfoques considerados (D19)

| Enfoque | Qué hace | Costo | Elegido |
| --- | --- | --- | --- |
| Una casa | `docs-cleanup` pasa a `project-layout` y absorbe `artifact-layout.md` y el layout de harness-audit | ~36 referencias; `/docs-cleanup` deja de existir | Sí (D19 A) |
| Extender sin mover | `docs-cleanup` suma referencias; las otras dos casas siguen | Tres casas para el mismo layout | No |
| Una casa sin la regla de escritura | flow-close conserva una regla corta al escribir | Dos textos sobre las mismas carpetas | No |

## Forma de `project-layout`

`SKILL.md` es un router corto, como en pstack (un `SKILL.md` que enruta, `references/` situacionales, `scripts/` con tests). Tres modos, cada uno con la condición que lo elige:

- **check:** corre `scripts/check_layout.py` y reporta; no escribe. Lo usan también `workspace-repo` y el cierre de un trabajo cuando revisa su material.
- **organizar o migrar:** el procedimiento actual de docs-cleanup (inventario, manifiesto, una confirmación, tag, aplicar, verificar), al que se suman los checks HA-PL, el manifiesto de migración, la pregunta del hub, la propuesta de `_support/docs` (D16) y el test de viabilidad del monorepo.
- **convención:** `references/layout.md` es la regla que `global.md` manda leer antes de escribir cualquier artefacto de soporte.

Referencias previstas, cada una ≤16.000 B:

- `references/layout.md` — sucesora de `artifact-layout.md`: raíz del proyecto, índice, hijos de `_support/` y si se versionan, código fuera de `_support/` en el repo dueño (scripts de base de datos, seeds y backfills en el backend; infraestructura en su carpeta), la prohibición de que el código de la aplicación lea o escriba `_support/`, retención de evidencia (sin cambios).
- `references/script-lifecycle.md` — los cinco niveles y qué sube cada uno; separado de `layout.md` porque solo se lee al promover o al organizar, no en cada escritura.
- `references/inventory.md` y `references/archive.md` — los actuales, con `data/` como protegido.
- `references/migrate.md` — los checks HA-PL movidos desde harness-audit, el manifiesto de migración, la pregunta del hub con sus cuatro opciones, el test de viabilidad y la parte compartida de una migración (depurar, procedencia por documento, chequeo de autonomía en copia aislada, retiro recuperable, entregar el destino antes de retirar el origen).

## Raíz del proyecto e índice

```text
<repo o workspace>/
├── AGENTS.md          reglas + una línea: «lee <casa de docs>/README.md primero»
├── HIVE.local.md      (ignorado)
└── _support/
    ├── docs/README.md ÍNDICE: una línea por documento, runbook y herramienta interna, con cuándo leerlo o usarlo
    ├── docs/<tema>/  docs/runbooks/
    ├── openspec/      versionado
    ├── sessions/      registros fechados; versionables si no son sensibles (sin cambio)
    ├── workspace/     ignorado
    ├── data/          ignorado y protegido (D17)
    ├── scripts/       ignorado: solo lo que nombra HIVE.local.md
    └── secrets/       ignorado
```

La casa de docs es `_support/docs`; un repo que ya usa otra recibe la propuesta de migrar y la conserva si el usuario no la acepta (D16). Esa excepción se declara en la cabecera del índice con una línea de gramática fija, en inglés porque un script la lee en proyectos ajenos: `Layout: docs at <path> · tools at <path>`, claves `docs` y `tools`, cada una opcional, separadas por ` · `, rutas relativas a la raíz. El modo migrar la escribe al declinar, en `<casa declarada>/README.md` (lo crea si falta) y actualiza la línea puntero de `AGENTS.md` a ese índice. `layout.md` dice que la línea manda sobre los destinos por defecto.

Dónde se busca el índice, en orden: `_support/docs/README.md`; `docs/README.md`; la ruta de una línea `Index: <path>` en `AGENTS.md`. Sin índice, es un hallazgo. Las herramientas internas viven en el código del repo dueño, en la carpeta que ya use (`scripts/`, `tools/`); asumido como decisión menor.

## Ciclo de vida de un script

| Nivel | Dónde | Qué lo sube |
| --- | --- | --- |
| 0 Desechable | scratch del host | nada; se borra al cerrar |
| 1 Corrido una vez | `_support/sessions/<fecha>-<trabajo>/` con salida, reversión y reporte | tocó datos compartidos |
| 2 Candidato | lo define flow-close «Candidates» (repetición dentro de un trabajo; se ofrece cuando su registro cuenta dos sesiones) | `script-lifecycle.md` enlaza, no lo repite |
| 3 Herramienta interna | código del repo dueño + línea en el índice | sin ID de ticket ni rutas absolutas, secretos desde su ubicación, `--dry-run` por defecto, idempotente, con test |
| 4 Mecanismo | lo elige defect-fix punto 5 (script, lint, test o hook) | `script-lifecycle.md` enlaza, no lo repite |

`script-lifecycle.md` es dueño de los niveles 0, 1 y 3; los niveles 2 y 4 son punteros a sus casas. flow-close `SKILL.md` está en 14.543 B: su cambio es un solo enlace en «Candidates».

## `check_layout.py`

Python con `-I`, sin dependencias, como `local_instructions.py`. Entrada: `--root <repo o workspace>` (repetible). Cada hallazgo: regla, ruta, corrección. Sale 1 con hallazgos. Reglas:

1. Script de `_support/scripts/` que ningún `HIVE.local.md` de esa raíz nombra por su ruta exacta o por la de su carpeta de primer nivel (nombrar una carpeta registra todo lo que contiene).
2. Archivo versionado bajo `_support/{scripts,secrets,workspace,data}`.
3. (revisión, no falla) Archivo de código versionado fuera de `_support/`, que no es test (una carpeta `tests/`, `test/`, `__tests__/`, `spec/` o `e2e/`, un sufijo `.test.`, `.spec.` o `_test.`, o un prefijo `test_`), en una línea que no es línea de comentario (marcas por lenguaje: `#`, `//`/`/*`/`*`, `--` solo en `.sql`) cita material de trabajo: `_support/{workspace,data,scripts,secrets,sessions,evidence,reports}`. Leer `_support/docs` o `_support/openspec` es leer conocimiento versionado y no cuenta; un comentario que cita un documento tampoco (la corrida real del 2026-11-18 dio ese falso positivo en project-a y project-g). Es código por extensión: `.js .mjs .cjs .ts .mts .tsx .jsx .py .go .rb .php .java .kt .cs .sh .sql .ps1`. Todo archivo que no es código por extensión queda fuera de la regla, cite lo que cite: `.gitignore`, `.ignore`, `.gitattributes`, `AGENTS.md`, `CLAUDE.md`, Markdown, `.engram/`, la configuración de CI y de hooks. Que el código de la aplicación en ejecución no lea `_support/` sigue siendo regla de juicio: por extensión el check no distingue aplicación de herramienta.
4. Entrada de primer nivel de `_support/` fuera de la convención (`docs`, `openspec`, `sessions`, `workspace`, `data`, `scripts`, `secrets`, `README.md`): se reporta para revisión, sin fallar, para clasificarla con el modo organizar.
5. Documento de la casa de docs ausente del índice, y enlace roto del índice. La casa de docs es `_support/docs`, o la que declare `docs at`; cuando `Layout:` declara `tools at <path>`, cada archivo directo de esa carpeta, salvo sus tests, también debe figurar en el índice.
6. En un coordinador: repo de la tabla ausente, con otro remoto, no ignorado o no reincluido en `.ignore`.
7. Una carpeta `docs/` con Markdown en la raíz que ninguna línea `Layout:` declara: se reporta como casa de docs no declarada, con la corrección de migrarla a `_support/docs` o declararla. Con la línea, no se reporta y la regla 5 se aplica a ella.

El script lee la línea `Layout:` antes de aplicar las reglas.

No revisa `Specs`: `hive doctor` ya lo hace (`doctor_project.go:469-477`), igual que el ignore de `_support/scripts` y `_support/secrets`; la fila de `rule-enforcers.md` nombra ambos mecanismos. Lee el estado con `git -C`, nunca abre `secrets/` ni el contenido de `data/`.

## `workspace-repo`

Procedimiento: (0) leer el test de viabilidad y la parte compartida de una migración de `project-layout/references/migrate.md`; (1) evaluar: los repos del workspace con su remoto y rama base, si pasa el test de viabilidad del monorepo (si pasa, lo recomienda en su lugar), qué hay en la raíz (`AGENTS.md`, `CLAUDE.md`, `HIVE.local.md`, `.engram`, `_support`), el destino del repo de specs (archivarlo o subirlo como coordinador conservando su historia), y si se crea el remoto, que es decisión del usuario; (2) correr `project-layout` en modo check y proponer qué partes de organizar o migrar correr; (3) `scripts/workspace_repo.py`: sin `--apply` imprime el plan, con `--apply` escribe `.gitignore` (hijos, `HIVE.local.md`, `.engram/` cuando existe, y `_support/{workspace,data,scripts,secrets}`), `.ignore` con `!<hijo>/`, la sección `## Repositorios` y el `## Hive` completo del coordinador y hace `git init`; nunca crea un remoto ni hace push. El `## Hive` lleva `Project`, `Base branch` (la rama inicial del coordinador), `Tracker` y `Specs: _support/openspec`; la skill pregunta el `Tracker` (propone el que comparten los hijos) y si declara `Delivery: direct-base` (lo recomienda porque deja commitear conocimiento directo a la base del coordinador, y dice que es lo único que autoriza efectos sin preguntar); el script recibe ambas respuestas como argumentos y no las decide. Subir el repo de specs como coordinador conservando su historia es manual: la skill lo describe y no lo automatiza; (4) modo setup: clona lo que falte según la tabla; (5) verificar con `check_layout.py` y una búsqueda desde la raíz.

Limitaciones que la skill declara: una sesión iniciada dentro de un hijo no carga el `AGENTS.md` del coordinador en Codex, Grok y Cursor, por eso cada hijo conserva su `## Hive` completo; los worktrees se crean dentro de cada hijo; no hay cambios atómicos entre repos.

## Manager

`hive doctor` en la raíz de un coordinador lo trata como cualquier repositorio (`rev-parse --show-toplevel`): exige `Project`, `Base branch`, `Tracker`, `Specs` (`doctor_project.go:42-47,459-465`) y bloquea un `HIVE.local.md` no ignorado (`:776-783`); por eso el `.gitignore` lo incluye y el test corre doctor sobre la salida real de `workspace_repo.py --apply`.

Viewer: la clave a la que quedan ligadas las capturas sale de `timelineProjectKey` (`tooling/viewer/timeline_store.go:672-740`), no de `resolveProject` (`index.go:94-118`, que solo da el nombre). `git init` sin remoto conserva la ruta como clave (`tooling/cli/view_command.go:734-739`); agregar `origin` la cambia y el store responde 409 `project_identity_changed` para una raíz con trabajos guardados, sin pérdida silenciosa. El test caracteriza ambos pasos; la skill avisa de ese efecto antes de que el usuario agregue el remoto. Si hiciera falta cambiar código, se agrega un delta MODIFIED a `walkthrough-viewer`.

Hosts ya instalados: el gestor retira los registros de catálogo ausentes de un release (`tooling/management/ownership.go:237-264`); la verificación confirma con `hive update --dry-run` que `docs-cleanup` aparece como retirado.

## Insumos copiados de memoria

Test de viabilidad del monorepo (los cuatro): (1) las mismas personas tienen escritura en todos los repos y ningún cliente o tercero necesita solo uno; (2) los despliegues se pueden elegir por ruta en un solo CI; (3) ningún repo tiene un contrato de publicación propio con consumidores externos (paquete publicado, fork, repo administrado por el cliente); (4) consolidar no exige fusionar historia de Git: se copia el estado vigente con procedencia.

Parte compartida de una migración (de la sesión Codex `01a1e864` en project-a, 2026-11-17): depurar primero (vigente contra histórico); procedencia por documento movido (commit y blob de origen, enlaces históricos al commit original); chequeo de autonomía desde el punto de entrada de la sesión en una copia aislada; retiro recuperable del origen; entregar el destino antes de retirar el origen.
