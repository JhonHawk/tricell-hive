# Design

## Approaches

- **A — `plan_layout.py` con `plan` y `apply` (elegido, D20 A).** Es el mismo patrón que `workspace_repo.py`. Reutiliza la enumeración de archivos y los ayudantes de `check_layout.py`, importándolo desde la carpeta del script, porque `python3 -I` no añade esa carpeta a `sys.path`. El comprobador sigue sin escribir nada.
- **B — Subcomandos dentro de `check_layout.py`.** Queda un solo archivo, pero el check pasa a escribir y crece a unas 900 líneas.
- **C — Tabla en prosa en `layout.md` aplicada por el agente.** Es lo que falló en carvento: el destino depende de cada corrida.

## Decisions

- **D20 A** — script nuevo con plan y apply.
- **D21 A** — grupos por propósito según quién llama al archivo.
- **D22 A** — la evidencia heredada sin versionar pasa al workspace.
- **D23 A** — la carpeta de infraestructura por defecto es `infra/`.

Supuestos menores, porque no cambian alcance, contratos, datos, costo ni reversibilidad:

- Un JSON o YAML de un árbol de infraestructura que ningún archivo versionado cita, por ruta o por nombre, sale como revisión en lugar de moverse: puede ser código muerto, como `github-actions-terraform-policy.json` en carvento.
- Las carpetas de conocimiento en la raíz de un repositorio de specs (`carvento-specs/contracts`) quedan fuera del planificador: no son material de soporte, y su destino lo fija la migración entre layouts. En el monorepo de carvento, la migración en curso ya las copió a `_support/contracts` y `_support/product`, donde el planificador sí las ve.

## Proposed shape

### Enumerar archivos (H20, compartido)

`list_files(base)` reemplaza a `walk_files`. Devuelve rutas ordenadas.

- **Dentro de un repo Git:** une estas dos listas:
  - `git ls-files -z --cached --others --exclude-standard -- <base>`;
  - para lo que está bajo `_support/`, también lo ignorado: `git ls-files -z --others --ignored --exclude-standard -- <base>/_support`, porque la convención ignora `_support/{scripts,workspace,data,secrets}` y los proyectos ignoran su `evidence/`.
- **Fuera de Git:** usa `os.walk` ordenado y no entra en directorios con `.git`, que son repos hijos.
- **Poda:** se descarta toda ruta con un componente en la lista fija `PRUNE` de AC16. Fuera de Git también se descarta un componente de `GATED` cuyo directorio padre tiene el manifiesto que lo produce; dentro de Git no, porque Git ya quitó la salida ignorada y un `build/` versionado es código fuente:
  - `target`: `Cargo.toml` o `pom.xml`;
  - `vendor`: `composer.json`, `go.mod` o `Gemfile`;
  - `dist` y `build`: `package.json`, `pyproject.toml`, `setup.py`, `build.gradle`, `build.gradle.kts` o `pubspec.yaml`;
  - `bin` y `obj`: un `*.csproj` o `*.fsproj`;
  - `deps`: `mix.exs`.
- **Secretos:** nunca se lista el contenido de `_support/secrets/`; solo se cuenta.

`check_scripts`, `check_index`, `check_undeclared_docs` y el planificador usan `list_files`.

### LAYOUT-08 (H22)

Es una regla `fail`. Aplica a un archivo versionado bajo `_support/<entry>/`, con `<entry>` fuera de `docs`, `openspec` y `sessions` (donde viven los scripts de una sola vez) y fuera de `scripts`, `secrets`, `workspace` y `data` (que ya reporta LAYOUT-02). El archivo debe cumplir una de dos condiciones:
- su extensión está en `CODE_EXTENSIONS ∪ {.tf, .tfvars, .hcl}`;
- es un `.json`, `.yaml` o `.yml` en una entrada de la familia IaC o en una entrada con algún `.tf`.

Un `README.md` de esa entrada no se reporta: el planificador lo mueve junto con su árbol.

La corrección dice: `run plan_layout.py plan --root <root> and move it with the project's code`.

`parse_layout` acepta las claves `docs`, `tools` e `infra`.

### Planificador `plan_layout.py`

```
plan_layout.py [plan] --root <dir> [--root <dir>…] [--override <src>=<dest>…] [--json] [--out FILE]
plan_layout.py apply --plan FILE
plan_layout.py verify --plan FILE
plan_layout.py undo --plan FILE
```

**Candidatos.** Se unen por ruta. Si una ruta llega por dos fuentes, cuenta la primera en este orden:
1. un archivo bajo `_support/scripts/` que `HIVE.local.md` no nombra (LAYOUT-01);
2. un archivo que reporta LAYOUT-08;
3. una entrada de `_support/` fuera de la convención (LAYOUT-04), tratada como carpeta según su tabla.

**Familia IaC.** Entradas `infra`, `infrastructure`, `terraform`, `deployments`, `aws`, `gcp`, `azure`, `iam` y `cloud`, o cualquier entrada cuyo árbol contiene un `.tf`. Una entrada de la familia se mueve como árbol completo, con una fila por archivo, incluidos `README.md`, `*.tfvars.example` y los JSON.

**Clase de un llamador.** Gana la primera que aplica a su ruta:
1. `ci`: `.github/workflows/`, `.gitlab-ci.yml`, `.circleci/`, `azure-pipelines.yml`, `bitbucket-pipelines.yml`, `Jenkinsfile`, `buildspec*.yml` o `amplify.yml`;
2. `hook`: `lefthook.yml`, `lefthook.yaml`, `.husky/`, `.pre-commit-config.yaml` o `.githooks/`;
3. `doc`: extensión `.md` o `.mdx`;
4. `package:P`: P es el directorio más cercano, sin contar la raíz del repo, que tiene uno de estos manifiestos: `package.json`, `go.mod`, `pyproject.toml`, `setup.py`, `composer.json`, `Gemfile`, `pom.xml`, `build.gradle`, `build.gradle.kts`, `Cargo.toml`, `*.csproj`, `pubspec.yaml` o `mix.exs`;
5. `root`: cualquier otra.

**Llamadores.** Son los archivos de las raíces que contienen un token de referencia que resuelve al candidato. Los tokens se definen en **Reescrituras**.
- Se excluyen los registros históricos, anclados a la raíz del repositorio para que una carpeta de código llamada `sessions/` no cuente: `openspec/changes/archive/`, `_support/openspec/changes/archive/`, `sessions/`, `_support/sessions/`, `evidence/`, `_support/evidence/`, `_support/workspace/`, `_support/reports/` y `CHANGELOG*`.
- También se excluyen las listas de ignorados (`.gitignore`, `.ignore`, `.dockerignore`, `.prettierignore` y cualquier `*ignore`): nombran patrones, no rutas que algo use, así que una entrada vieja ahí no rompe nada.
- Un candidato no cuenta como llamador para clasificar a otro. Sus referencias sí se reescriben.
- Una referencia de prefijo o glob en un archivo `ci` cuenta como llamador `ci` solo de las pruebas que cubre, porque es un descubrimiento de pruebas, como el `find _support/scripts -name '*.test.sh'`.

**Prueba.** Un nombre `*.test.*`, `*.spec.*`, `*_test.*` o `test_*`. Su sujeto es el mismo nombre sin el marcador, en el mismo directorio.

**Una sola vez.** Se parte el nombre por `-`, `_` y `.`. Es de una sola vez si algún token es `backfill`, `fix`, `cutover`, `migrate`, `migration`, `seed`, `cleanup`, `oneoff`, `once` o `repair`, o si el nombre contiene una clave `\b[A-Za-z]{2,10}-\d{2,}\b` cuyo prefijo no es un runtime, una codificación o una versión de la lista fija `NOT_KEYS` (`node`, `utf`, `python`, `sha`…).

**Casas.** Se resuelven una vez por raíz:
- **Raíz de herramientas:** `Layout: tools at`; si no, la primera que exista de `scripts`, `script`, `tools` o `hack`; si no, `scripts`.
- **Casa IaC:** `Layout: infra at`; si no, la primera que exista de `infra`, `infrastructure`, `terraform` o `deployments`; si no, `infra` (D23).
- **Casa de docs:** `Layout: docs at`; si no, `_support/docs`.
- **Carpeta de un paquete:** la primera que exista de `scripts`, `script` o `tools`; si no, `scripts`.
- **Grupos:** `ci`, `hooks` y `ops`. Si la raíz de herramientas ya tiene un alias, se usa ese:
  - `ci`: `ci`, `cicd`, `pipeline`, `pipelines`;
  - `hooks`: `hooks`, `git-hooks`, `githooks`;
  - `ops`: `ops`, `operations`, `runbooks`.

**Reglas de archivo.** Gana la primera que aplica. La subruta es la ruta tras `_support/scripts/` o tras `_support/<entrada>/`.

| # | Condición | Destino |
|---|---|---|
| R1 | Es prueba con sujeto entre los candidatos | La carpeta destino de su sujeto |
| R2 | Archivo de un árbol de la familia IaC | `<casa IaC>/<subruta>` si la entrada es alias de la casa IaC; si no, `<casa IaC>/<entrada>/<subruta>`. Un `.json`, `.yaml` o `.yml` que ningún archivo versionado fuera de su propia fila cita, por ruta o por nombre, pasa a `review` (`unreferenced`) |
| R3 | Prueba, `*.lib.*` o fixture sin sujeto | La carpeta destino que comparten más candidatos de los que cita; los empates se resuelven por orden de ruta. Si no cita candidatos, sigue con R4 |
| R4 | Todos sus llamadores son `package:P` con el mismo P; o no tiene llamadores de paquete y el primer componente de su subruta es igual al nombre base de exactamente un paquete P | `<P>/<carpeta del paquete>/<subruta>`, quitando ese primer componente cuando es igual al nombre base de P |
| R5 | Algún llamador `ci` | `<herramientas>/<ci>/<subruta>` |
| R6 | Algún llamador `hook` | `<herramientas>/<hooks>/<subruta>` |
| R7 | Solo llamadores `doc` | `<herramientas>/<ops>/<subruta>` |
| R8 | Algún llamador `root` | `<herramientas>/<subruta>` |
| R9 | Sin llamadores y de una sola vez | `_support/sessions/<fecha>-<trabajo>/<nombre>`. La fecha es la del primer commit si está versionado, o el mtime si no. El trabajo es el nombre sin extensión, en minúsculas y kebab |
| R10 | Ninguna anterior | `review` (`no rule`) sin destino |

**Carpetas de LAYOUT-04.** Gana la primera que aplica:

| # | Condición | Destino |
|---|---|---|
| F1 | Familia IaC | Archivo por archivo con R2 |
| F2 | `evidence`, `reports`, `backups` o `migration-data` que una línea no comentada de código o configuración versionada cita (`_support/<nombre>`) | `review` (`runtime path`) |
| F3 | `evidence` o `reports` | Una fila por subcarpeta hacia `_support/workspace/<fecha>-<trabajo>/evidence/`. La fecha es la primera `\d{4}-\d{2}-\d{2}` del nombre, o si no hay, el mtime más antiguo. El trabajo es el nombre sin esa fecha, en minúsculas y kebab. Los archivos sueltos van a `<fecha>-legacy-<entrada>`. La fila versionada es `review`, porque un movimiento a una carpeta ignorada borraría historia |
| F4 | `backups` o `migration-data` | `_support/data/<fecha más antigua>-<nombre>/` |
| F5 | Solo `.md`, `.mdx`, `.png`, `.jpg`, `.jpeg`, `.gif`, `.svg` o `.webp` | `<casa de docs>/<nombre>/`. Sus citas, incluso desde código, se reescriben |
| F6 | Cualquier otra, como `archive` o `security` | `review` (`no rule`) |

**Avisos de fila.** Van como `note` y no cambian el destino, salvo la colisión:
- el destino ya existe → `collision`, y la fila pasa a `review`;
- hay un archivo versionado con el mismo nombre base y otra extensión → `same name elsewhere`;
- dos paquetes con el mismo nombre base → `ambiguous package`, y R4 no aplica por componente.

**Reescrituras.** Se buscan tokens con límites, de modo que la forma más larga gana y una subcadena nunca se reescribe sola. Un token es la ruta con `_support/` que termina en el candidato o en un prefijo de candidatos, y en Markdown también un enlace relativo (`../infrastructure/main.tf`). Una sustitución delante (`$VAR/`, `"$VAR"/`, `$(git rev-parse --show-toplevel)/`) se conserva tal cual. Se resuelve así:
- tras una sustitución, `../x` → relativo al archivo que lo cita, y cualquier otro → relativo al repo;
- `./` o `../` al inicio → relativo al archivo que lo cita, luego al repo, donde corren CI y los scripts de paquete;
- `<repo>/` al inicio, con el nombre de otra raíz → relativo al workspace que la contiene;
- cualquier otro → relativo al repo.

La forma nueva se calcula desde la ubicación nueva del archivo que cita, también cuando ese archivo está dentro de una carpeta que se mueve. Un Markdown que se mueve recalcula además sus enlaces `../` para que sigan apuntando adonde apuntaban.

Casos:
- **Literal:** el token resuelve a un candidato con destino. Se reescribe con la misma forma y se guarda `file`, `line`, `sha256_line`, `old` y `new`.
- **Prefijo o glob:** resuelve a un directorio con candidatos. Se reescribe solo si todos los candidatos bajo él van a un mismo directorio que conserva su estructura. Si no, es `review` con la lista ordenada de directorios nuevos.
- **Sin resolver:** un token que no resuelve a nada existente se ignora.
- **Archivos que no se tocan:** se leen como bytes y se decodifican en UTF-8 estricto. Si fallan, o tienen un byte NUL, no se tocan y salen como `review` cuando contienen un token.
- **Al escribir:** se conservan los finales de línea y el modo del archivo.

**Rutas internas.** En código, una ruta relativa sin `_support/` suele ser relativa al directorio de trabajo, así que no se reescribe; se revisa así. Aplica a un candidato que se mueve y cuyo texto tiene, en cualquier línea, un ancla de ubicación propia: `BASH_SOURCE`, `dirname "$0"`, `$(dirname`, `__dirname`, `import.meta.url`, `__file__` o `Path(__file__)`. Cada literal `(\.\./)+[\w./-]*` del archivo se resuelve desde el directorio viejo y desde el nuevo. Si el destino resuelto cambia, sale una fila `review` con los dos.

**Índice.** Por cada herramienta movida por R4–R8 se emite una línea esqueleto con la ruta, el comando y su prueba. El agente completa para qué sirve.

**Salida.** Un JSON con:
- `version`;
- `roots`, con la ruta y el commit de `HEAD`;
- `rows`, con `source`, `fingerprint`, `tracked`, `rule`, `destination`, `notes` y `override`. `fingerprint` es el `sha256` de un archivo, o el `sha256` de la lista ordenada de `(ruta relativa, tamaño)` de una carpeta;
- `rewrites`;
- `reviews`;
- `layout_lines`.

Todo va ordenado y sin fechas de ejecución. El texto es una tabla Markdown para el manifiesto. Mismas entradas, misma salida byte a byte.

**Overrides.**
- `--override src=dest` reemplaza el destino de una fila y la marca `override: user`.
- `--home tools=<path>`, `--home infra=<path>` o `--home docs=<path>` cambia la casa de toda una categoría (R5–R8, R2 o F5), marca sus filas como `override: user` y añade a `layout_lines` la línea `Layout:` completa de esa raíz: sus claves existentes más la nueva, en el orden docs, tools, infra.

Es la única vía por la que nace una línea `Layout:` (H26).

**Raíz sin Git.** En una raíz que no es repositorio, como la raíz de un workspace, solo aplican R9, F3, F4, F5 y F6. Un archivo que R2 o R4–R8 colocarían sale como `review` (`no repository`), porque el código va en un repositorio y no en la raíz del workspace.

**R3, precisión.** Los candidatos relacionados con una prueba sin sujeto son los que ella cita. Los de un `*.lib.*` o fixture son los candidatos que lo citan.

**apply.**
1. Revalida todo y se niega sin tocar nada ante el primer fallo:
   - el `fingerprint` de cada fila;
   - cada línea a reescribir;
   - que ningún destino exista;
   - que el commit de cada raíz siga siendo el del plan.
2. Crea los directorios.
3. Mueve: `git mv` si el archivo o la carpeta están versionados, `os.replace` si no. Una carpeta se mueve entera. El plan ya mandó a revisión todo lo versionado cuyo destino Git ignora, porque ahí saldría de Git.
4. Reescribe línea por línea, sin tocar el resto del archivo, aunque tenga cambios locales de otro trabajo.
5. Escribe `<plan>.undo.json` antes del primer cambio y lo actualiza tras cada archivo reescrito y cada movimiento, con las carpetas que creó y las que vació. Si algo falla a mitad, se detiene y dice cómo deshacer con `undo`.
6. Imprime las filas `review` que quedan.

Nunca hace commit, push ni `git add` de archivos que no mueve.

**undo.** Invierte exactamente lo que registra `<plan>.undo.json`: deshace las reescrituras si la línea sigue siendo la nueva, y devuelve los movimientos con `git mv` o `os.replace`. Así no toca cambios ajenos, a diferencia de `git restore` o `git checkout -- .`.

**verify.**
- `git grep -F` de cada ruta vieja en cada raíz, fuera de los registros históricos, no da coincidencias.
- Cada herramienta movida tiene su nombre en el índice. Si no hay índice, falla con la corrección de crear `_support/docs/README.md`.
- Sale 1 con la lista si algo falla.

### Detector (H24)

`codexItem` reconoce `type == "additional_tools"`. Si una herramienta, o una dentro de un namespace, se llama `request_user_input`, y el `collaboration_mode` del `turn_context` es `default` o `plan`, marca `QuestionOffered`. Esos son los modos que su descripción declara: «only available in Default or Plan mode». `textRoundWithoutCard` trata `QuestionOffered` como prueba de permiso, igual que una llamada.

### Criterio nuevo (H25)

`review_question_native_option`:
- **Host:** solo Codex.
- **Pregunta de revisión:**
  - una línea de título de decisión (`D<n> — …`) en un mensaje del asistente que contiene `revisión de código`, `code review` o `revisión`, y entre cuyas opciones aparece `hive-review-code`;
  - o una pregunta de tarjeta con esos mismos textos.
- Debe cumplir las dos cosas: el título nombra una revisión de código, y sus opciones nombran un mecanismo (`hive-review-code`, `/review`, `Codex review`, sin revisión dedicada). Una decisión sobre aplicar hallazgos no es esa pregunta.
- **Falla** si ninguna de sus opciones contiene `/review` ni `Codex review`.
- **Pasa** si alguna la contiene.
- **No observado** sin pregunta de revisión.

## Impact

- **Consumidores de `check_layout.py`:**
  - el modo check de project-layout;
  - `workspace-repo/SKILL.md`, en sus pasos 1 y 4: el paso 4 exige que el check salga 0, y LAYOUT-08 puede romperlo;
  - `tests/skills/layout_check_test.py` y `tests/skills/workspace_repo_test.py`;
  - `rule-enforcers.md`.
- **LAYOUT-08** hace fallar un check que hoy pasa en proyectos con código bajo `_support/`, como carvento y facti; es el efecto buscado.
- **La enumeración nueva** quita falsos positivos y vuelve a ver lo ignorado bajo `_support/`.
- **El detector y el criterio** solo afectan a `--audit-session`.
- **Producción:** no se toca.

## Recovery

- Revertir el PR restaura la skill.
- En un proyecto, `apply` no hace commit y `undo --plan FILE` invierte exactamente lo que hizo, sin tocar cambios ajenos.
