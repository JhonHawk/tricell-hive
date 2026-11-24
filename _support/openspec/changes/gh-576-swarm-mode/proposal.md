# gh-576-swarm-mode — modo swarm: una sesión coordina 2–4 tickets con un dueño por ticket

| Campo | Valor |
| --- | --- |
| Estado | Borrador · plan en revisión |
| Tracker | [#576](https://github.com/JhonHawk/tricell-hive/issues/576) · abierto · enhancement |
| Git | tricell-hive · base `development` en `d5d9ba65` · rama propuesta `feat/swarm-mode` |
| Playbook | feature |

## Decisiones

2026-11-23
- **D10 A** — ¿Dónde viven las excepciones del modo? En un skill `swarm` aparte, activado solo por el usuario, que guarda todas las excepciones y usa las mecánicas de E2E por referencia; los flujos normales no cambian. Descartados: ampliar `e2e-coordination.md` y un flag en `## Hive`.
- **D11 B** — ¿Qué tickets admite la fase 1? Pequeños y listos, con UI incluida; los que no califican van a su propia sesión.
- **D12 A** — ¿Cómo se topa el gasto de un lote? Por conteo, fijado en la ronda 0: owners simultáneos (máximo 4) y corridas por ticket; al llegar al tope no se lanza nada más y cada worktree queda con su head y su checkpoint.
- **D13 A** — ¿Cada plan de ticket pasa por `hive-review-plan`? Solo los planes riesgosos según «Proportionality»; los demás confían en la verificación independiente del build.
- **D14 A** — ¿Cómo no chocan los servidores de UI? El coordinador asigna a cada ticket de UI un puerto libre como recurso exclusivo; un ticket de UI cuyo proyecto no deja cambiar el puerto no califica.
- **D15 A** — ¿Qué rol hace research, plan y build de un ticket? Un rol nuevo, `hive-ticket-owner`, definido solo para el modo; `hive-build-owner` queda intacto.
- **D16 A** — ¿Con qué perfil de modelo corre ese rol? `reasoning`.
- **H1 A** — El registro de ejecución de E2E rechaza el rol nuevo y pide un plan antes de que exista: ¿cómo se verifica la entrega de cada owner? Se amplía `tasks.py` para aceptar `hive-ticket-owner` como rol de unidad, y el registro de cada ticket se crea al aprobar su plan; `validate-handoff` sirve tal cual.
- **H3 A** — ¿Dónde vive el registro de ejecución de cada ticket? En el `_support/workspace` del checkout del coordinador, junto a `decisions.tsv`, con las mismas exigencias de ignorado y sin seguimiento; T5 amplía la validación de ubicación para eso.
- **H2 B** — ¿Dónde escribe el owner su change folder cuando el Specs vive en otro repositorio? En la única ruta de change folder de ese checkout principal que el coordinador le asigna; no hace commits ahí: los hace el coordinador.
- **Supuesto** — Las líneas `Ruling:` no van al checkpoint (`tasks.py checkpoint` las borra y el spec deja el registro de tareas solo para el progreso): el owner las devuelve en cada parada y el coordinador las escribe en el log de decisiones. Alternativa descartada: `--check "Ruling: …"` bajo Evidence. Se revierte cambiando el formato de retorno.
- **Supuesto** — «Pequeño» es: un solo repositorio y un alcance estimado de 10 archivos o menos, sin datos guardados ni migración; «listo» es pasar la lista de ticket listo de `create-issue`. Alternativa descartada: el «cambio pequeño y entendido» de «Proportionality», que no lleva plan. Se revierte cambiando el umbral en el skill.
- **Supuesto** — El modo arranca con el comando del usuario o cuando el usuario lo nombra: la línea global enlaza `skill:swarm/SKILL.md` con esa condición, porque `disable-model-invocation` impide que Claude Code cargue el skill por su descripción. Alternativa descartada: solo el comando. Se revierte quitando la condición de la línea.
- **Supuesto** — El nombre es `swarm`, palabra del usuario. Alternativa descartada: `multi-ticket`. Se revierte con un renombre antes de integrar.
- **Supuesto** — Tras una parada, el coordinador retoma al mismo hijo con su historial donde el host lo documenta (Claude Code, Antigravity) y en los demás relanza un hijo del rol que lee el plan y el checkpoint del worktree. Alternativa descartada: relanzar siempre. Se revierte en una línea del procedimiento.
- **Supuesto** — sbx queda fuera de la fase 1 (agente dentro del sandbox sin medir, CVE-2026-18171 abierto, token en texto plano en su configuración). Alternativa descartada: incluirlo. Se revierte con un cambio de fase 2.
- **Supuesto** — Ronda 0 de activación: una ronda confirma los tickets, el veredicto de cada uno, la entrega, los topes y la creación de sus worktrees; esa respuesta es la autorización de crearlos. Alternativa descartada: crearlos al aprobar cada plan. Se revierte moviendo la creación después de la parada 1.
- **Supuesto** — Una rama y un PR por ticket según la entrega en vigor; la falla de uno no detiene a los demás. Alternativa descartada: una rama de integración común. Se revierte con una opción de integración en la ronda 0.
- **Supuesto** — Un ticket que escribe datos guardados o corre una migración no califica en la fase 1: la base de datos de desarrollo compartida solo se lee. Alternativa descartada: una base por ticket, que requiere sbx. Se revierte en fase 2.
- **Supuesto** — El skill lleva `disable-model-invocation: true`, que el manager ya traduce a `allow_implicit_invocation: false` en Codex (`tooling/management/hidden_skills.go:14`); HA-SK-04 de `audit-harness` se reescribe en términos generales para no marcarlo. Alternativa descartada: un mecanismo nuevo por host. Se revierte quitando la línea del frontmatter.
- **Supuesto** — La página del sitio es una entrada en `site/src/content/docs/skills.md`, sin página propia. Alternativa descartada: una página en `flows/`. Se revierte añadiéndola.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «y que no afecte como funciona ya nuestro flujo no? y si tuvieramos un modo swarm que nos permitiera habilitar esas excepciones (y que las escepciones solo estuvieran ahi?)»
> «o puede ser un hijo completo que haga el research, plan + build, lo que sigue siendo separado son las verificaciones»
> «quiero que planeemos para poderlo probar en una sesión real»

Criterios de #576 (leídos el 2026-11-23): una sola sesión con 2 a 4 tickets, un plan por ticket con preguntas agrupadas; cada ticket en un worktree o un sandbox `sbx`, nombrado en su checkpoint; agentes solapados en el tiempo; dos Postgres en el 5432 a la vez; una rama y un PR por ticket, y la falla de uno no detiene a los demás; el hilo principal vuelve a correr las verificaciones antes de integrar. Fuera de alcance según #576: otra máquina o la nube, más de 4 tickets, resolver conflictos de merge entre tickets. Esta fase deja fuera además `sbx` y los dos Postgres (supuesto sbx).

## Objective

Hoy cada ticket pasa por plan, build, verificación y cierre de uno en uno, y trabajar varios a la vez exige una sesión por ticket. Las reglas que lo impedirían en una sola sesión (un hijo no planea, no escribe el plan, no hace commits; el principal toma las capturas previas y crea los worktrees solo para la obra en curso) protegen al flujo normal y deben seguir igual. Con el cambio, el usuario activa el modo swarm, confirma 2 a 4 tickets en una ronda y un hijo `hive-ticket-owner` por ticket hace research, plan y build en su worktree, con paradas donde el coordinador le pregunta al usuario; la verificación sigue siendo independiente y el coordinador publica cada ticket. Todas las excepciones viven en el skill del modo y en el rol, y fuera de ellos los flujos no cambian.

## Scope and acceptance

Dentro: el skill `swarm` y sus referencias, el rol `hive-ticket-owner` con su bloque de límites, una línea de `content/guidance/global.md`, la aceptación del rol en `tasks.py`, la redacción general de HA-SK-04 en `audit-harness`, las pruebas de contenido que fijan el alcance del modo y del rol, el delta de `workflow-lifecycle`, la documentación (sitio, `llms.txt`, `CHANGELOG.md`, `instruction-resources.md`, `agent-delivery.md`) y las filas de la rúbrica de sesiones para auditar la prueba real. Fuera: `sbx`, cambios en `tooling/` e `integrations/`, cambios en el texto de `flow-plan`, `flow-build`, `flow-verify`, `flow-close`, `hive-mode` y las hojas de `hive-core`, y más de 4 tickets.

- AC1 — Una prueba de contenido falla cuando un archivo distribuido bajo `content/` nombra el modo swarm o `hive-ticket-owner` fuera de `content/skills/swarm/`, `content/agents/development/hive-ticket-owner.md`, una sola línea de `content/guidance/global.md` y `content/skills/flow-plan/scripts/tasks.py` · false today because no existe esa prueba: añadir hoy un puntero al modo en `flow-build/SKILL.md` no falla nada · Observed in: `go test -count=1 ./tests/content/...` sobre el candidato, con un caso negativo que añade la mención a un flujo y la ve fallar
- AC2 — El skill `swarm` no se invoca solo y se alcanza desde la guía global: su `SKILL.md` tiene `disable-model-invocation: true`, la línea de `content/guidance/global.md` lo enlaza para cuando el usuario nombra el modo, y la instalación en Codex deja `swarm/agents/openai.yaml` con `allow_implicit_invocation: false` · false today because no existe el skill · Observed in: `~/.agents/skills/swarm/agents/openai.yaml` (con `allow_implicit_invocation: false`), `~/.agents/skills/swarm/SKILL.md` y el enlace `~/.claude/skills/swarm`, tras el refresco de mantenedor que sigue a la integración
- AC3 — En la ronda 0, una sola ronda muestra para cada ticket su veredicto (lista de ticket listo, un repositorio y ≤ 10 archivos estimados, sin datos guardados ni migración; para UI, puerto asignable) y pide confirmar los tickets aptos, la entrega, los topes (owners simultáneos ≤ 4, corridas por ticket) y la creación de sus worktrees; un ticket no apto se nombra con su razón y se envía a su propia sesión · false today because no hay entrada para varios tickets: hoy cada ticket pasa solo por `hive-mode` · Observed in: la primera sesión real del modo en Claude Code tras el despliegue, revisada con la rúbrica de sesiones **Retired:** reemplazado por AC11 (elegibilidad con Specs `direct-base` y sin `hold`)
- AC4 — El rol `hive-ticket-owner` (perfil `reasoning`, acceso `implement`) escribe su change folder y su checkpoint en su worktree o en la única ruta de change folder asignada en el checkout del repositorio de Specs, y hace commits locales con el helper de commit con ámbito solo en su worktree, rama y rutas asignadas; no hace commits en el repositorio de Specs, no hace push, no abre PRs, no escribe en el tracker ni en la memoria, no lanza hijos y no crea worktrees; su bloque de límites lo fija la prueba de límites de roles · false today because no existe el rol y la prueba de límites solo admite commits de `hive-build-owner` · Observed in: `go test -count=1 ./tests/content/...` sobre el candidato
- AC5 — El owner de cada ticket termina su primera corrida en la parada 1 con el plan escrito y, en el formato de texto de `questions`, cada decisión que los criterios a–f envían al usuario, los efectos externos, las reglas de negocio supuestas, las rutas que escribirá y cada hecho externo como comando con su extracto; puede devolver un bloqueo en cualquier momento; cada retorno trae sus líneas `Ruling:`, que el coordinador escribe en `decisions.tsv` · false today because ningún hijo planea ni devuelve decisiones del plan · Observed in: la primera sesión real del modo, en los retornos de los owners y en el `decisions.tsv` del lote
- AC6 — Tras la parada 1 el coordinador agrupa las decisiones de todos los tickets en rondas, mueve cada ticket a en progreso, crea su registro de ejecución, lanza `hive-review-plan` solo para un plan riesgoso y retoma o relanza al owner; tras la parada 2 corre `validate-handoff`, lanza un `hive-verify-task` por ticket que cubre todos sus criterios y es el verificador del registro (y `hive-review-ux` si el ticket cambia una pantalla), vuelve a correr las verificaciones del ticket y publica cada ticket en su rama y su PR según la entrega; la falla de un ticket no detiene a los demás · false today because no existe coordinación de varios tickets · Observed in: la primera sesión real del modo
- AC7 — Al llegar a un tope de la ronda 0, el coordinador no lanza más owners, y cada ticket abierto queda con su worktree, su head y su checkpoint nombrados en el reporte · false today because nada limita los hijos de un lote: la sesión de Pi del 2026-11-23 (`01a206d8`) murió a mitad por créditos · Observed in: la primera sesión real del modo con un tope de corridas que se alcance **Retired:** reemplazado por AC12 (topes separados)
- AC8 — Un ticket de UI recibe del coordinador un puerto distinto como recurso exclusivo; su owner sirve el worktree en ese puerto, toma las capturas previas y posteriores en él y las devuelve por ruta; el coordinador las publica · false today because las capturas previas las toma el hilo principal en su checkout · Observed in: la primera sesión real del modo con un ticket de UI
- AC9 — La página de skills y la vista general de flujos del sitio, `llms.txt`, `CHANGELOG.md`, `instruction-resources.md` y `agent-delivery.md` nombran el modo y el rol, y el sitio compila · false today because ninguno los nombra · Observed in: `pnpm check`, `pnpm lint` y `pnpm build` en `site/` y la comprobación de texto sobre el candidato
- AC10 — `tasks.py` acepta un registro cuya unidad tiene rol `hive-ticket-owner` y que vive en el `_support/workspace` ignorado del checkout del coordinador aunque el change folder esté en un worktree o en otro repositorio; sigue rechazando otros roles y un registro rastreado, no ignorado o fuera de un `_support/workspace`; y `validate-handoff` comprueba esa unidad como a un owner de E2E · false today because `tasks.py:622` exige `hive-build-owner` y `tasks.py:473-490` exige el registro en el repositorio del change folder · Observed in: `python3 -m unittest discover -s tests/skills -p '*_test.py'` sobre el candidato
- AC11 — En la ronda 0, una sola ronda muestra para cada ticket su veredicto (lista de ticket listo, un repositorio y ≤ 10 archivos estimados, sin datos guardados ni migración; para UI, puerto asignable; con Specs en otro repositorio, que ese repositorio declare `Delivery: direct-base`) y pide confirmar los tickets aptos, la entrega sin `hold`, los topes (owners simultáneos ≤ 4, corridas por ticket) y la creación de sus worktrees; un ticket no apto se nombra con su razón y se envía a su propia sesión · **Replaces:** AC3 · false today because no hay entrada para varios tickets: hoy cada ticket pasa solo por `hive-mode` · Observed in: la primera sesión real del modo en Claude Code tras el despliegue, revisada con la rúbrica de sesiones **Retired:** reemplazado por AC13 (también excluye `direct-base`)
- AC12 — Con el tope de owners simultáneos lleno, los demás tickets esperan y cada uno arranca cuando un owner se detiene; un ticket que llega a su tope de corridas no recibe más owners y el reporte nombra su worktree, su head y su checkpoint, mientras los demás siguen · **Replaces:** AC7 · false today because nada limita los hijos de un lote: la sesión de Pi del 2026-11-23 (`01a206d8`) murió a mitad por créditos · Observed in: la primera sesión real del modo con un tope que se alcance
- AC13 — En la ronda 0, una sola ronda muestra para cada ticket su veredicto (lista de ticket listo, un repositorio y ≤ 10 archivos estimados, sin datos guardados ni migración; para UI, puerto asignable; con Specs en otro repositorio, que ese repositorio declare `Delivery: direct-base`) y pide confirmar los tickets aptos, la entrega de cada repositorio con código de los tickets sin `hold` ni `direct-base` (un repositorio de Specs aparte conserva su `direct-base`), los topes (owners simultáneos ≤ 4, corridas por ticket) y la creación de sus worktrees; un ticket no apto se nombra con su razón y se envía a su propia sesión · **Replaces:** AC11 · false today because no hay entrada para varios tickets: hoy cada ticket pasa solo por `hive-mode` · Observed in: la primera sesión real del modo en Claude Code tras el despliegue, revisada con la rúbrica de sesiones

## Prerequisites

None: the build needs nothing from outside the repository.

La observación de AC5, AC6, AC8, AC12 y AC13 necesita después del despliegue una sesión real del usuario con 2 a 4 tickets aptos (propuesta: PRJ-931, PRJ-933 y PRJ-944 si su veredicto los admite); no es un piloto y no la ejecuta este trabajo.

## Verified context

- Investigación de esta sesión (2026-11-23): 93 sesiones de ciclo completo desde el 2026-11-04; pico de contexto del hilo principal con mediana de 447k en proyectos de producto y 8 de 34 sobre 800k; 64–71 % con preguntas después de empezar a editar. Base de D11 y de la salida por bloqueo.
- Sesión de Pi `01a206d8` (project-b, 2026-11-23): 103 hijos, hasta 20 a la vez, 31 tickets integrados, ~$184 en ~75 min, cortada por créditos; otra sesión la terminó desde git. Base de D12 y AC12.
- `hive-build-owner` tiene acceso `implement` y su excepción de commits la fija `tests/content/agent_boundary_test.go:44` y `tests/content/single_home_test.go:116`; `content/guidance/global.md:24` nombra solo E2E.
- `content/skills/flow-plan/scripts/tasks.py:622` exige `unit role must be hive-build-owner`, y `cmd_checkpoint` (`tasks.py:284–347`) reescribe la cabecera del checkpoint conservando solo sus líneas conocidas.
- `tooling/management/hidden_skills.go:14` traduce `disable-model-invocation: true` a la política de invocación de Codex de forma genérica.
- Ningún archivo bajo `content/`, `site/src`, `README.md` ni `llms.txt` nombra hoy «swarm» (`rg -i swarm`, 2026-11-23).
- `content/guidance/global.md` pesa 25,094 de 30,000 bytes; `content/skills/hive-mode/SKILL.md`, 15,003 de 16,000.
- Retomar un hijo con su historial: documentado en Claude Code (`SendMessage`) y Antigravity; no documentado en Codex, Cursor, OpenCode y Grok (docs oficiales leídas el 2026-11-23).

- Comprobación de texto de `AGENTS.md` «Checks by path» (sus cuatro comandos) · pasa sobre la base, 83 s · 2026-11-23
- `go test -count=1 ./tests/content/` · pasa sobre la base, 12 s · 2026-11-23
- `go test -count=1 -run 'TestAgentBoundaryBlock|TestRulesHaveOneHome' ./tests/content/` · pasa sobre la base (3 pruebas) · 2026-11-23
- `go test -count=1 -run 'TestSwarm' ./tests/content/` · not measured: la prueba la añade T3 · 2026-11-23
- `pnpm --dir site check`, `pnpm --dir site lint`, `pnpm --dir site build` · pasan sobre la base (0 hints; 14 archivos; 20 páginas), 4 s · 2026-11-23
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` · pasa sobre la base, 412 pruebas, 89 s · 2026-11-23

## Effects

None: the tasks take no effect on an external system.

## Delivery

`interactive` (respuesta de la ronda de entrega, 2026-11-23): parada de validación antes del push para leer el skill y el rol; después push, PR a `development` y merge. Revisión de código nativa de Claude Code (`/code-review`), como prefiere `AGENTS.md`; `hive-review-security` también corre sobre el candidato (respuesta Seguridad A). Esto no autoriza el refresco de la instalación local más allá de lo que `AGENTS.md` fija para el checkout del mantenedor, ni la sesión real de prueba.

## Review status and progress

2026-11-23 — Segunda relectura: un P2 introducido (la exclusión de `direct-base` alcanzaba al repositorio de Specs aparte); aplicado tal como lo propuso el revisor, sin otra relectura: la exclusión se limita a los repositorios con código de los tickets.
2026-11-23 — Relectura de la revisión en build: sin P0–P2; tres P3 arreglados por decisión H4 A (adelanto del commit del Specs aparte nombrado, re-pregunta de entrega nombrada, `direct-base` excluido en la ronda 0); el puntero del registro ya lo evita `5274596a`.
2026-11-23 — Revisión en build (ronda de correcciones de verify): `/code-review` y `hive-review-security` encontraron que `validate-handoff` rechazaría el worktree con el change folder sin commit, que el registro admite un solo verificador por unidad con todos sus criterios, que el tope de simultáneos cortaba el lote, que `hold` seguía haciendo commits (F1) y que un Specs sin `direct-base` se escribía igual (F2). AC3 → AC11, AC7 → AC12; AC6 y el delta ajustados; T1 cierra AC11 y AC12.
2026-11-23 — Plan escrito. Revisión del plan (`hive-review-plan`, ~314 s, 183k tokens): 2 P1, 7 P2 y 5 P3. H1 A y H2 B decididos por el usuario; el resto aplicado: decisiones `Ruling` al log de decisiones, deltas MODIFIED de revisión de plan, registro y log, activación por la línea global, AC2 observado en la instalación real, líneas de verificación por «Checks by path», umbral de «pequeño», HA-SK-04, excepciones nombradas, deberes del coordinador, `flows/overview.md`, AC10 y T5 nuevos. Relectura (~61 s, 210k tokens): resuelto lo anterior; H3 A decidido por el usuario (registro en el checkout del coordinador, T5 ampliada); rutas de instalación de AC2 corregidas a `~/.agents/skills/` y `global.md:23` nombrada en T1 y T3.
