# hive-core-stage-packets — Las hojas viven en `hive-core` y cada flujo lee un paquete por etapa

| Campo | Valor |
| --- | --- |
| Estado | Borrador · plan revisado, listo para flow-build cuando el orden interno se cumpla |
| Tracker | Sin ticket |
| Git | `JhonHawk/tricell-hive`, base `development` en `7f6277e0` (2026-11-21); ramas propuestas `refactor/hive-core-sheets` (E1), `feat/stage-packets` (E2), `feat/stage-rules-reply` (E3), cada una en su worktree de `.claude/worktrees/` |
| Playbook | ninguno (reestructura de la distribución que además cambia lo que lee el agente) |

## Decisions

2026-11-21
- **Investigación A** — ¿Qué dirección toma la reestructura tras comparar Hive actual, Hive antes del corte (`2bb96478`) y pstack? La opción 4: las hojas viven en una sola skill oculta `hive-core`, el índice de `global.md` se queda, el manager genera al desplegar un paquete por etapa, cada flujo nombra cómo termina y el checkpoint guarda las reglas de la etapa.
- **Usuario** — El corte de referencia del reforge es `2bb96478` (development, 2026-11-18); `e005e617` queda como contraste pre-pstack.
- **D1 A** — ¿Dónde viven las hojas dentro de `hive-core`? En `hive-core/references/`. El manager no cambia para almacenarlas y la auditoría de costo ya cuenta esas lecturas.
- **D2 A** — ¿Qué pasa con `reporting-family-leaves`? Se construye antes, en su propio PR; después se mudan 60 hojas, no 62.
- **D3 A** — ¿Cómo se parte la entrega? Tres PRs en orden: mudanza (E1), paquetes (E2), checkpoint y línea de cierre (E3).
- **D4 A** — ¿Qué entra en el paquete de `flow-verify`? Lo que declara obligatorio, partido en un paquete de entrada y uno de salida.
- **Entrega A** — `automatic`. **Revisión A** — Claude Code `/code-review`.
- **D5 A** — ¿Qué nombre llevan los archivos de las hojas? Sin prefijo: `hive-core/references/<nombre>.md`, por ejemplo `questions.md`. Así `global.md` queda en ~29,018 B y deja margen para alinear el índice. Se reescriben también las menciones sueltas `hive-rule-<nombre>` fuera del archivo histórico.
- **D6 A** — ¿Qué meta mide el plan en lecturas? Como máximo 26 lecturas en la ruta plan→build. La prueba de huella compara bytes sin contar repeticiones y reporta, sin fallar por ellas, las hojas nuevas que traen los paquetes y las repeticiones.
- **D7 A** — ¿Quién alinea las líneas del índice con su hoja? Este plan, en T1, con la primera línea «Read» de cada hoja. `reporting-family-leaves` retira su T6.
- **D8 A** — ¿Qué lleva la línea de cierre de cada flujo? La fila de reporte que termina el flujo y solo lo que el flujo añade y la fila no tiene.
- **D9 A** — ¿Cuándo se refresca la instalación local? Una sola vez, tras mergear E3, avisando antes para cerrar o reiniciar las sesiones abiertas. Esto deja sin efecto, para este trabajo, el refresco tras cada merge que pide `AGENTS.md`.
- **D10 A** — ¿Cómo se evita que otra sesión despliegue E1 o E2 antes de tiempo? No se coordina: cada unidad debe funcionar desplegada sola. E1 es autocontenida, y en E2 cada flujo lee sus hojas si su paquete falta. El dry-run de cada unidad es su puerta, y el reporte de cada merge avisa que el próximo refresco de cualquier sesión obliga a reiniciar las sesiones abiertas.
- **Supuesto** — **Replaces:** el supuesto de más abajo sobre enlaces dentro de un paquete; este manda donde difieran. Un enlace con ancla a otro miembro del mismo paquete apunta a la hoja con su ancla, para no perder la sección; sin ancla, apunta al encabezado del miembro dentro del paquete. Las definiciones de enlace de estilo referencia (`[x]: ruta`) se reescriben igual. El paquete generado entra en la lista de archivos y anclas con la que se valida. Alternativa descartada: mandar el enlace con ancla al encabezado del miembro, que pierde la sección. Se revierte cambiando el renderizador.
- **Supuesto** — La limpieza de `__pycache__` cumple cinco condiciones: borra solo archivos regulares `*.pyc`; los comprueba con `Lstat`; rechaza un `__pycache__` que sea enlace; corre después de que la retirada quede registrada; reporta en la salida de `apply` cualquier directorio que conserva. Alternativa descartada: reportarlo en `status` o `doctor`. Se revierte quitando la limpieza.
- **Supuesto** — Si el texto del resultado de una lectura falta o está truncado, la auditoría reporta `not_observed` para esa hoja y no un fallo. Alternativa descartada: dar la hoja por no leída. Se revierte contándola como fallo.
- **Supuesto** — Cada paquete termina con la línea `<!-- hive packet end: <clave> · <miembro>,<miembro>,… -->`. La auditoría toma la lista de miembros de esa línea y cuenta el paquete como leído solo si el resultado de la lectura trae la línea y un encabezado exacto `## <miembro>` por cada miembro que nombra. Si falta la línea o algún encabezado, el resultado está truncado, también en el corte por la mitad de Codex, y la auditoría da `not_observed`. El nombre de un miembro es su ruta del manifiesto sin `.md`, por ejemplo `hive-core/references/questions`, así que no choca con otra hoja ni con una referencia de un flujo. Antes de comparar, la auditoría quita el número de línea que algunas herramientas anteponen, como el `N\t` de la herramienta Read de Claude Code. La comparación exacta evita confundir un miembro con una sección propia de una hoja, como `## Change folder`. Es una señal igual en todos los hosts. Alternativa descartada: contar encabezados contra un número, que no separa un miembro de una sección. Se revierte quitando la línea.
- **Supuesto** — La línea de entrada de cada flujo dice también: si la última línea del paquete no es su línea de fin, leer el resto. Así el agente no pierde reglas en silencio con una lectura truncada. Alternativa descartada: dejar la señal solo a la auditoría. Se revierte quitando la cláusula.
- **Supuesto** — Si el paquete que nombra `Rules:` falta tras una compactación, se leen las hojas que enlaza el flujo, igual que al entrar. Alternativa descartada: dejarlo solo en la línea de entrada. Se revierte quitando la cláusula.
- **Supuesto** — El paquete de `flow-plan` deja fuera `reuse-before-writing`, que la ruta plan→build ya recibe en el paquete de build, y conserva `plan-handoff`. Así ese paquete queda ≤ 23,000 B y la ruta en 26 lecturas. Alternativa descartada: sacar `plan-handoff`, que añade una lectura. Se revierte cambiando el manifiesto.
- **Supuesto** — El nombre `hive-core` se mantiene: ahora solo almacena, el índice sigue en `global.md` y no hay un segundo router. Alternativa descartada: otro nombre. Se revierte renombrando la carpeta y sus enlaces.
- **Supuesto** — La membresía de cada paquete se declara en `<flujo>/assets/stage.json`, un objeto cuyas claves cumplen `^stage(-[a-z]+)?$` y nombran cada paquete, con una lista ordenada de rutas como valor. Por ejemplo `{"stage": […]}`, o `{"stage-entry": […], "stage-exit": […]}` en `flow-verify`. Alternativa descartada: una tabla en Go o una lista plana, que no puede describir dos paquetes. Se revierte moviendo la lista al manager.
- **Supuesto** — Los principios `hive-principle-*`, su índice y `hive-typescript` no se mudan: `hive-typescript` debe seguir visible con `paths`. Alternativa descartada: mudarlos también. Se revierte con otra mudanza.
- **Supuesto** — En la auditoría de sesiones cuentan como lectura de la hoja tres rutas: la vieja `hive-rule-<nombre>/SKILL.md`, la nueva `hive-core/references/<nombre>.md` y un paquete cuyo encabezado `## <nombre>` la contiene. Así las sesiones anteriores siguen auditables y la auditoría no depende de los manifiestos actuales. Alternativa descartada: reescribir los fixtures o leer los manifiestos. Se revierte quitando las rutas nuevas.
- **Supuesto** — El disparador de cada hoja es su primera línea de cuerpo después del título `#`, que empieza por «Read». La prueba compara la línea del índice con esa línea en minúsculas y con los enlaces markdown reducidos a su texto. Alternativa descartada: conservar el frontmatter solo para el disparador. Se revierte devolviendo el frontmatter.
- **Supuesto** — Cada paquete mide como máximo 24,000 B y 1,000 líneas una vez generado. Codex corta a la mitad una salida mayor, y el hook de Grok lee entero hasta 64 KiB y 1,000 líneas. La prueba de contenido usa 23,000 B como tope de la suma de miembros, para dejar margen a encabezados y enlaces reescritos. Alternativa descartada: 35 KB, el máximo teórico de Codex, no confirmado. Se revierte cambiando los topes.
- **Supuesto** (reemplazado por el supuesto de enlaces con ancla de más arriba, que manda donde difieran) — Dentro de un paquete, un enlace a otro miembro del mismo paquete apunta al encabezado `## <miembro>` del propio paquete, para no invitar a releer la hoja. Un enlace con ancla dentro de un miembro, y cualquier enlace a una hoja que no está en el paquete, apuntan a la hoja en `hive-core`. El manager valida los enlaces del paquete generado. Alternativa descartada: apuntarlo todo a las hojas. Se revierte cambiando el renderizador.
- **Supuesto** — La línea `Rules:` del checkpoint usa la ruta relativa al almacén de skills (por ejemplo `flow-build/stage.md`), igual en todos los hosts. Alternativa descartada: la ruta absoluta de un host en un `tasks.md` versionado. Se revierte escribiendo la absoluta.
- **Supuesto** — Para revertir E2 se sigue este orden:
  1. revertir antes E3, que nombra los paquetes en `Rules:` y en `tasks.py`;
  2. desplegar el contenido revertido con el binario de E2, que retira los paquetes;
  3. reconstruir el binario.

  Un binario sin E2 rechaza el estado que registra paquetes. Alternativa descartada: reconstruir primero. Se revierte en sentido inverso.
- **Supuesto** — Al retirar una carpeta de skill de Hive, el manager borra también los `__pycache__` que solo contienen `.pyc` dentro de ella. Si no, `hive-rule-trackers/` queda sin `SKILL.md`. Alternativa descartada: dejarlo al usuario. Se revierte quitando esa limpieza.
- **Supuesto** — La tarea T6 de `reporting-family-leaves` se retira de ese plan antes de construirlo, por D7 A, conservando su corrección de `close-hygiene`.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Solicitud

> 1) retomar el hive-core, y todas las hojas se reintroducen ahi […] 2) Regresar al modo de hive antes del breakpoint […] reescribiendo las skills con las buenas practicas que nos da pstack […] y si, investiga la velocidad porfa
>
> o sea, en lugar de tener 80 skills con 1 hoja, tendriamos 1 kills con un router interno (otro de refuerzo) y 80 referencias o sheets
>
> que igual no se deja de investigar la opción 2, y si vale la pena que surga una opción 3 con lo mejor de ambas opciones

Respuesta elegida (A, 2026-11-21): adoptar la opción 4 como dirección.

## Objetivo

Hoy cada regla de momento es una skill oculta `content/skills/hive-rule-*/SKILL.md` (62 carpetas, 181,529 B). Un flujo enlaza entre 8 y 22 hojas. En sesiones reales fuera de este repositorio (2026-11-10 a 21), las lecturas de guía de Hive por sesión subieron tras el reforge de 12.9 a 35.4 en Codex y de 3.9 a 17.6 en Grok. Cada lectura es más chica (~1.1k tokens) y el total de tokens casi no cambió. El costo está en la latencia por llamada (~3–4 min por sesión) y en las relecturas tras compactar (324 en Grok). Además, Codex genera 72 `agents/openai.yaml` y Antigravity lista 62 hojas en su catálogo porque no puede ocultarlas.

Al terminar:
- las hojas son archivos de una sola skill oculta `hive-core`, con un hogar cada una;
- cada flujo de uso frecuente lee al entrar un paquete generado con las hojas que siempre necesita;
- el checkpoint nombra ese paquete, así que una compactación cuesta una relectura;
- cada flujo nombra la fila de reporte con la que termina.

## Alcance y aceptación

Dentro:
- la mudanza de las hojas a `hive-core/references/<nombre>.md`, con sus referencias y scripts propios en `references/<nombre>/` y `scripts/<nombre>/`;
- los enlaces y las menciones en `content/`, `tests/`, `tooling/`, `site/`, `_support/docs/`, `README.md`, `llms.txt`, `AGENTS.md` y `CHANGELOG.md`;
- la generación de paquetes en el manager, con los manifiestos de 8 flujos: hive-mode, flow-plan, flow-build, flow-verify (dos paquetes), flow-close, git-workflow, flow-deploy y unattended-delegation;
- la línea `Rules:` del checkpoint y la línea `Reply:` de 7 flujos;
- los deltas de `context-delivery`, `versioned-installation` y `workflow-lifecycle`.

Fuera:
- el contenido de las reglas, salvo lo que la mudanza exige: enlaces, la primera línea «Read» y la definición de hoja de `global.md`;
- `reporting-family-leaves`, que va antes;
- los principios y `hive-typescript`;
- paquetes para `create-*`;
- el archivo histórico de `_support/openspec/changes/archive/`, `_support/sessions/` y los fixtures de sesiones ya grabadas;
- pruebas piloto con modelos, que están pausadas;
- medir la mejora en sesiones reales: queda como seguimiento tras el despliegue.

Restricción de todas las tareas (no es criterio porque hoy se cumple): `content/guidance/global.md` ≤ 30,000 B, y cada `SKILL.md` y cada referencia ≤ 16,000 B.

- AC1 — Cada hoja de regla existe como `content/skills/hive-core/references/hive-rule-<x>.md` … **Retired:** D5 A cambia el nombre de archivo; lo reemplaza AC11.
- AC2 — Un plan de instalación del manager desde el candidato instala una carpeta `hive-core` con las hojas y retira las carpetas `hive-rule-*` registradas, sus enlaces de catálogo y los `__pycache__` que dejaron. Para Codex genera un solo `agents/openai.yaml` de regla (el de `hive-core`) en lugar de uno por hoja, y cada enlace `skill:` del bloque global desplegado resuelve a un archivo instalado · false today because el plan instala 60 carpetas de hoja en `6fa78bfd` y un `openai.yaml` por cada una · Observed in: la instalación local del mantenedor tras el refresco único posterior a E3 (`hive status`), apoyada antes por la prueba del manager sobre el contenido real y por el `--dry-run` de cada unidad
- AC3 — Una prueba de contenido falla cuando el disparador de una línea de `## Rule leaves` no es texto literal, en minúsculas y con los enlaces reducidos a su texto, de la primera línea de cuerpo de su hoja, que empieza por «Read». La prueba pasa en el candidato · false today because esa prueba no existe y solo 4 de 60 líneas citan su hoja literalmente · Observed in: `go test ./tests/content/...`, vista fallar con una línea alterada
- AC4 — `go run ./tests/pilot --audit-session` cuenta como lectura de la hoja … **Retired:** D5 A cambia la ruta; lo reemplaza AC12.
- AC5 — Para cada flujo con `assets/stage.json`, el manager instala un paquete por clave del manifiesto. Cada paquete:
  - es la concatenación, en el orden de la lista, de sus miembros sin frontmatter, cada uno bajo `## <miembro>`;
  - lleva los enlaces reescritos como fija «Decisions» y validados una vez generado;
  - mide ≤ 24,000 B y ≤ 1,000 líneas;
  - desaparece cuando su manifiesto o su flujo dejan la versión, se reescribe cuando cambia un miembro y se reinstala si se borra.

  Además, un archivo de paquete escrito a mano en el fuente se rechaza · false today because no existe ningún paquete ni manifiesto · Observed in: la instalación local tras el refresco único, apoyada antes por las pruebas del manager sobre el contenido real
- AC6 — Al entrar, cada uno de esos flujos dice que lea su paquete entero y que no vuelva a leer las hojas que contiene, y que si el paquete falta lea las hojas enlazadas. Conserva el enlace a cada hoja para los momentos fuera del paquete, y la definición de hoja de `global.md` admite el paquete como forma de leerla · false today because ningún flujo nombra un paquete · Observed in: `content/skills/<flujo>/SKILL.md` y `content/guidance/global.md` del candidato
- AC7 — El escenario plan/build declara como máximo 20 lecturas … **Retired:** D6 A fija la meta en 26; lo reemplaza AC13.
- AC8 — `tasks.py <carpeta> checkpoint --rules "<flujo>/stage.md · extra: <hojas>"` escribe la línea `- Rules:`, y `--principles "<nombres>"` escribe `- Principles:` con los principios de código que cambiaron una decisión, o `none` (D13 A). Las llamadas siguientes sin cada flag conservan su línea, y `check_change.py` acepta cada una bien formada y rechaza una mal formada. `--principles` une nombres sin duplicados y `none` nunca borra nombres. `flow-build` escribe `Principles:` cuando un principio cambia una decisión, o `none` al terminar si no existe, y `flow-close` la toma antes de borrar el checkpoint para la parte **Principles** del reporte final. La hoja del checkpoint y una línea de `global.md` dicen el orden tras compactar: checkpoint, `HIVE.local.md`, el paquete que nombra `Rules:`, sus extras y la memoria persistente. Ningún otro lugar de `content/` dice otro orden · false today because `tasks.py` rechaza `--rules` y `--principles` · Observed in: `python3 -I -m unittest discover -s tests/skills -p '*_test.py'` y los archivos de `content/` del candidato
- AC9 — `hive-mode`, `flow-plan`, `flow-build`, `flow-verify`, `flow-close`, `flow-deploy` y `git-workflow` tienen cada uno una línea `Reply:` con la fila de la hoja de reporte que termina el flujo y solo lo que el flujo añade y esa fila no tiene. Una prueba de contenido exige la línea · false today because ninguno la tiene · Observed in: `go test ./tests/content/...`
- AC10 — Ninguna mención de `hive-rule-<x>/SKILL.md` queda … **Retired:** el conteo estaba mal medido y D5 A cambia lo que se busca; lo reemplaza AC14.
- AC11 — **Replaces:** AC1. Cada hoja de regla existe como `content/skills/hive-core/references/<nombre>.md`, sin frontmatter, con su primera línea de cuerpo después del título empezando por «Read». No queda ningún directorio `content/skills/hive-rule-*`. `hive-core/SKILL.md` tiene `disable-model-invocation: true` y no lista las hojas, y la definición de hoja en `global.md` dice que una hoja es un archivo de `hive-core` · false today because existen 60 directorios `hive-rule-*` en `6fa78bfd`, y `hive-core` no existe · Observed in: `content/` del candidato y `go test ./tests/content/...`
- AC12 — **Replaces:** AC4. `go run ./tests/pilot --audit-session` cuenta como lectura de la hoja, para `leaf_read_before_action`, `report-endings` y el costo de lectura, la ruta vieja `hive-rule-<nombre>/SKILL.md`, la nueva `hive-core/references/<nombre>.md` y un paquete con su encabezado `## <nombre>`, leídos con la herramienta de lectura o por shell (`cat`, `sed`) · false today because solo reconoce `<hoja>/SKILL.md` y el shell compara con ese sufijo · Observed in: `go test ./tests/pilot/...` con fixtures nuevos de cada forma, de lectura y de shell
- AC13 — **Replaces:** AC7. El escenario plan/build de `reading_footprint_test.go` declara como máximo 26 lecturas de archivo (hoy 39). Reporta sus bytes sin repeticiones, las hojas nuevas que traen los paquetes y los bytes con repeticiones, y su comparación con `2bb96478` no falla por esas dos últimas cifras · false today because declara 39 lecturas y no conoce los paquetes · Observed in: la salida de `go test ./tests/content -run TestRequiredProcedureReadingFootprint -v`
- AC14 — **Replaces:** AC10. Las búsquedas `rg 'hive-rule-[a-z-]+/(SKILL\.md|references/|scripts/)'` y `rg -w 'hive-rule-[a-z-]+'` no encuentran nada en el candidato. Quedan excluidos el archivo histórico, `_support/sessions/`, los fixtures de sesiones grabadas, las carpetas de cambio abiertas y la lista de nombres viejos que la auditoría acepta · false today because la primera búsqueda encuentra 851 menciones en 156 archivos, medido en `7f6277e0` con esas exclusiones · Observed in: las dos búsquedas sobre el candidato **Retired:** las specs principales solo cambian al archivar, y la lista de nombres viejos no se puede excluir con un filtro; lo reemplaza AC15.
- AC15 — **Replaces:** AC14. Este comando no imprime nada sobre el candidato:

  `rg -n -e 'hive-rule-[a-z-]+/(SKILL\.md|references/|scripts/)' -e '\bhive-rule-[a-z-]+' <checkout> --glob '!_support/openspec/changes/**' --glob '!_support/openspec/specs/**' --glob '!_support/sessions/**' --glob '!tests/fixtures/sessions/**'`

  Excluye el archivo histórico, las carpetas de cambio, las specs principales (cambian con los deltas al archivar), `_support/sessions/` y los fixtures grabados. La auditoría arma los nombres viejos como `"hive-rule-" + nombre`, así que no los escribe literales. Los deltas de esta carpeta cubren cada requisito de las specs principales que nombra una hoja · false today because el comando imprime 801 líneas en 182 archivos, medido en `7f6277e0` · Observed in: el comando sobre el candidato, y los deltas de esta carpeta frente a `rg -n 'hive-rule-' _support/openspec/specs` **Retired:** no detecta la forma `hive-rule-*`; lo reemplaza AC16.
- AC16 — **Replaces:** AC15. Este comando no imprime nada sobre el candidato:

  `rg -n -e 'hive-rule-[a-z-]+/(SKILL\.md|references/|scripts/)' -e '\bhive-rule-[a-z-]+' -e 'hive-rule-\*' <checkout> --glob '!_support/openspec/changes/**' --glob '!_support/openspec/specs/**' --glob '!_support/sessions/**' --glob '!tests/fixtures/sessions/**' --glob '!CHANGELOG.md'`

  Además, los deltas de esta carpeta cubren cada requisito de las specs principales que nombra una hoja, también los que traiga `reporting-family-leaves` si se archiva antes. `CHANGELOG.md` se excluye porque sus versiones publicadas conservan los nombres viejos, y su entrada nueva los cita para describir la mudanza · false today because el comando imprime 802 líneas en `7f6277e0`, entre ellas la forma `hive-rule-*` de `README.md:35`, `llms.txt:34`, `site/src/content/docs/skills.md:28`, `instruction-resources.md:36` y `skills-and-agents.md:10` · Observed in: el comando sobre el candidato, y los deltas frente a `rg -n 'hive-rule-' _support/openspec/specs`

## Prerequisites

None: the build needs nothing from outside the repository.

Orden interno, que no es un prerrequisito externo:
- E1 empieza cuando estén en `development` dos cambios:
  - `reporting-family-leaves` (D2 A), ya sin su T6 (D7 A);
  - `conditional-plan-reviewers`, ya integrado (PR #561) y archivado (PR #562) en la base `7f6277e0`, comprobado el 2026-11-21 con `git log origin/development`.
- E1 rebasa sobre esa base, y T1 rehace allí el inventario de enlaces y la medida de `global.md`.

## Contexto verificado

- La investigación de 2026-11-21 (9 hijos) mide:
  - lecturas, relecturas y tiempos por host fuera de este repositorio;
  - en cuántos sitios se lee cada hoja (47 compartidas, 15 de una sola fase);
  - los costos de las opciones 1 a 3;
  - la supervivencia a la compactación por host: la guía global se reconstruye en los 7, y el cuerpo de una skill en ninguno, salvo los primeros 5,000 tokens en Claude Code.

  Engram: `decision/reforge-breakpoint-and-restructure-options`.
- El manager, inspeccionado sobre `7f6277e0`:
  - `plan.go:229,243,330-370,437` define las carpetas desplegadas. `collectResources` recorre carpetas anidadas y `validSource` las acepta.
  - `references.go:24-26,55-93,126` valida enlaces, solo de archivos de la versión.
  - `hidden_skills.go:58-102` genera las políticas de Codex.
  - `generated_hooks.go:40` marca el origen generado.
  - `ownership.go:64,139-299` registra y retira archivos; un binario más viejo rechaza un origen registrado que no conoce.
  - `apply.go:171-187` no borra carpetas no vacías.
  - `integrations/grok/readhook/readhook.go:18-21,97-117` lee entero hasta 64 KiB y 1,000 líneas.
  - `integrations/opencode/context_snapshot.go:16-17` no se ve afectado.
- Pruebas que asumen que una hoja es una skill:
  - en `tests/content`: `leaf_skill_test.go`, `single_home_test.go`, `leaf_links_test.go`, `skill_size_test.go`, `reading_footprint_test.go` (39 lecturas en plan/build y el tope de 182 B sobre `2bb96478`), `budget_test.go`, `rule_enforcers_test.go` y `skill_links_test.go`;
  - en `tests/pilot`: `session_audit.go:1706,1728`, `session_audit_endings.go:14` y `session_audit_cost.go:35`, más 13 fixtures;
  - en `tests/skills`: `run_check_test.py:12`, `jira_rest_test.py:14` y `work_time_test.py:10`.
- `global.md` mide 28,593 B. Con el prefijo `hive-rule-` en el nombre de archivo quedaría en 29,868 B; sin él (D5 A), en 29,018 B (`sed` sobre `7f6277e0`).
- Límites de lectura de un archivo por host, leídos el 2026-11-21:
  - **Claude Code:** 25,000 tokens (code.claude.com/docs/en/tools-reference).
  - **Grok:** 100,000 B, más el hook de Hive (`~/.grok/docs/user-guide/08-skills.md:270`).
  - **Pi:** 2,000 líneas o 50 KB (`truncate.js:10-11` de 1.5.2).
  - **OpenCode:** 2,000 líneas o 50 KB, registrado para v2.2.4.
  - **No medidos:**
    - el valor por defecto actual de `tool_output_token_limit` de Codex;
    - la lectura de Antigravity;
    - las lecturas parciales de Cursor, que son 55–59% de las lecturas según `_support/docs/architecture/compaction-survival.md:16`.

    Ninguno tiene una lectura de solo lectura sin correr un modelo. Los encuentra primero el refresco único tras E3. Después del despliegue se revisan las primeras sesiones reales de esos hosts para ver si leen los paquetes enteros. Es un seguimiento, no un watch: todavía no hay ningún caso observado.
  - Si un paquete se lee truncado, el agente pierde reglas de esa etapa. La alternativa de la línea de entrada lo cubre solo cuando falta el paquete. Si el seguimiento ve lecturas truncadas, se bajan los topes.

Comandos de verificación, corridos el 2026-11-21 sobre un clon local compartido de la base `7f6277e0` en el scratchpad:

- `go -C <checkout> test -count=1 ./...` · exit 0, 252 s · 2026-11-21
- `go -C <checkout> test -count=1 ./tests/content/...` · exit 0, 2 s · 2026-11-21
- `go -C <checkout> test -count=1 ./tests/content -run TestRequiredProcedureReadingFootprint -v` · exit 0; unión plan/build de 191,862 B frente a 184,489 B en `2bb96478` · 2026-11-21
- `go -C <checkout> test -count=1 ./tests/pilot/...` · exit 0, 68 s · 2026-11-21
- `go -C <checkout> test -count=1 ./tooling/management/...` · exit 0, 98 s · 2026-11-21
- `go -C <checkout> test -count=1 ./integrations/...` · exit 0, 2 s · 2026-11-21
- `go -C <checkout> test -count=1 ./tooling/cli/...` · exit 0, 235 s · 2026-11-21
- `python3 -I -m unittest discover -s <checkout>/tests/skills -p '*_test.py'` · exit 0, 314 tests, 90 s · 2026-11-21
- `go -C <checkout> vet ./tooling/... ./integrations/... ./tests/content/... ./tests/pilot/...` · exit 0 · 2026-11-21
- `rg -o 'hive-rule-[a-z-]+/SKILL\.md'`, sin el archivo histórico, `_support/sessions/` ni los fixtures de sesiones grabadas · 851 menciones en 156 archivos · 2026-11-21
- El comando completo de AC16, desde la raíz del clon · 802 líneas · 2026-11-21
- `python3 -I <flow-plan>/scripts/check_change.py <esta carpeta>` · exit 0 · 2026-11-21

No medidos:
- La prueba nueva de manifiestos (`TestStageManifests`); la encuentra primero T7, y su forma más cercana es la prueba de huella.
- `hive update --source <worktree> --dry-run`; necesita el candidato y el binario del worktree. La encuentra primero T4; si falla, E1 no se entrega.

## Effects

None: the tasks take no effect on an external system.

## Delivery

- **Entrega:** `automatic`, respondido el 2026-11-21. Cada PR (E1, E2, E3) se revisa, se pushea, se abre y se mergea sin parar, en orden.
- **Refresco local (D9 A y D10 A):** este trabajo refresca una sola vez, tras mergear E3.
  - Antes, aviso y espero a que cierres o reinicies las sesiones abiertas, incluida la paralela. Es la única parada dentro de `automatic`.
  - El refresco reconstruye primero el binario: un archivo temporal y luego `mv`. Después corre `hive update --source <checkout> --out <archivo>` y `hive apply --plan <archivo>`.
  - Si otra sesión refresca antes, cada unidad funciona sola. El reporte de cada merge lo avisa.
- **Cuerpo de cada PR:** lleva la lista de lugares revisados y el cambio de tamaño de `global.md`, como pide `AGENTS.md`.
- **Revisión de código:** Claude Code `/code-review` sobre cada candidato antes del push, respondido el 2026-11-21.
- **Lo que esto no autoriza:**
  - mergear PRs que este trabajo no escribió, como los de `reporting-family-leaves` o `conditional-plan-reviewers`;
  - promover a `master`;
  - cortar una release;
  - crear los worktrees sin preguntar al empezar el build.

## Review status and progress

- 2026-11-21 — Plan escrito tras D1–D4. Dos revisores `hive-review-plan`, uno de guía y otro del manager, encontraron:
  - guía: 5 P1, 7 P2 y 6 P3;
  - manager: 1 P1, 7 P2 y 4 P3.
- 2026-11-21 — Revisión del plan con D5–D9 y los hallazgos que se sostienen:
  - el formato del manifiesto;
  - los enlaces dentro del paquete;
  - la auditoría por encabezados y por shell;
  - el mapa de reescritura completo;
  - las pruebas Python en E1;
  - la definición de hoja en `global.md`;
  - el orden tras compactar;
  - la recuperación de E2;
  - los `__pycache__`;
  - las exclusiones y el conteo de AC14;
  - los límites de host no medidos.

  Invalida AC1, AC4, AC7 y AC10, que quedan retirados y reemplazados por AC11–AC14. Ninguna tarea estaba marcada.
- 2026-11-21 — Relectura de los dos revisores:
  - manager: 10 de 11 hallazgos resueltos y 1 en parte, más 1 P2 y 3 P3 nuevos;
  - guía: 14 de 18 resueltos y 3 en parte, más 1 P1, 2 P2 y 1 P3 nuevos, y 4 restos en `reporting-family-leaves`.
- 2026-11-21 — Segunda revisión, con D10 A, que aplica:
  - AC15, que reemplaza a AC14;
  - dos deltas MODIFIED de `workflow-lifecycle` para los requisitos que nombran `report-endings` y `background-wait`;
  - la membresía de `flow-plan`;
  - enlaces con ancla y definiciones de enlace;
  - la limpieza estricta de `__pycache__`;
  - `not_observed` en la auditoría;
  - revertir E3 antes que E2;
  - reconstruir el binario en el refresco;
  - el seguimiento en lugar del watch;
  - los restos de `reporting-family-leaves`.

  Invalida AC14. Ninguna tarea estaba marcada.
- 2026-11-21 — Segunda relectura:
  - manager: B y D resueltos, A cerrado por D10, 5 y C en parte, y nuevos E (P2) y F (P3);
  - guía: N1–N4 y P1-4 resueltos, P2-1 en parte, y nuevos G1 (P2), G2 y G3 (P3).
- 2026-11-21 — Tercera revisión:
  - AC16, que reemplaza a AC15 y añade la forma `hive-rule-*` y la exclusión de `CHANGELOG.md`;
  - la marca de fin de paquete como señal de truncado;
  - la alternativa tras compactar;
  - la regla de enlaces en el delta;
  - el binario del dry-run de E2;
  - la nota de G3 en T5;
  - el tope de crecimiento en la T6 de `reporting-family-leaves`.

  Invalida AC15. Ninguna tarea estaba marcada.
- 2026-11-21 — Tercera relectura:
  - guía: G1–G3, el marcador de AC6 y el tope de T6 resueltos, y nuevos H1 y H2 (P3);
  - manager: 5, E y F resueltos, C en parte (el corte por la mitad deja la línea de fin) y nuevo G (P3).
- 2026-11-21 — Cuarta revisión:
  - la línea de fin nombra los miembros, y la auditoría exige un encabezado exacto por cada uno (resuelve C y G);
  - la línea de entrada lee el resto si falta la línea de fin (H1);
  - el supuesto viejo de enlaces queda marcado como reemplazado (H2).

  No cambia ningún criterio. Cambia el delta de `versioned-installation`, así que el revisor del manager relee.
- 2026-11-21 — Cuarta relectura (manager): C y G resueltos en la regla, y nuevos H (P2), I y J (P3).
- 2026-11-21 — Quinta revisión:
  - se quita el número de línea antes de comparar, con un fixture de una lectura real de Claude Code (H);
  - el fixture de la sección propia pasa (I);
  - el nombre de cada miembro es su ruta sin `.md` (J).

  Cambia el delta de `versioned-installation`, así que el revisor del manager relee.
- 2026-11-21 — **D13 A** (durante los hallazgos de `reporting-family-leaves`): E3 añade además un campo `Principles:` al checkpoint, junto a `Rules:`, para que la línea de principios de `flow-build` sobreviva una compactación o una sesión nueva antes del cierre. AC8 y T8 se revisaron el 2026-11-21 durante el build, antes de E3, y se releyeron con un revisor de plan. `reporting-family-leaves` ya está integrado (PR #563, `9e8633ae`), así que E1 parte de esa base.
