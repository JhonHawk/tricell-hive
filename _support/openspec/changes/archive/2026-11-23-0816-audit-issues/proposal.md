# audit-issues — Skill que audita los tickets abiertos, cierra los que no aplican y agrupa los demás bajo padres

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #603, merge 6d1ab6f0 |
| Tracker | ninguno |
| Git | tricell-hive · base `development` (`48824e10`) · worktree `.claude/worktrees/audit-issues` · rama `feat/audit-issues` |
| Playbook | feature |

## Decisions

2026-11-22
- **D3** — ¿Qué tickets entran en una pasada? La skill primero cuenta los tickets candidatos y sugiere el alcance de cada pasada; el usuario lo elige. La búsqueda de duplicados recorre todo el tracker, no solo la pasada.
- **D4 A** — ¿Qué pasa con un ticket que no aplica (falso positivo, ya corregido o duplicado)? Se cierra como cancelado o duplicado, con un comentario que da la evidencia, después de una sola aprobación del conjunto.
- **D5 A** — ¿Qué tan a fondo se verifica cada ticket? Se lee el código; se reproduce solo cuando el veredicto cerraría un Bug.
- **D6 A** — ¿Puede la skill tocar tickets de otro autor? Puede moverlos bajo un padre dentro del conjunto aprobado; cerrar uno ajeno solo deja un comentario para su autor.
- **D8 (supuesto)** — Un duplicado se cierra solo con evidencia firme: el mismo error, los mismos pasos para reproducirlo o la misma causa demostrada en el código. Sin ella, la relación es «related». Descartado: cerrar por similitud de texto. Revertir: bajar el umbral en `references/checks.md`.
- **D9 (supuesto)** — Cada grupo declara su nivel. Un grupo por causa raíz es un solo cambio. Un grupo por dominio o por módulo es un padre cuyos hijos se entregan por separado. Descartado: tratar todos los grupos igual. Revertir: quitar el nivel.
- **D10 (supuesto)** — Un Bug se cierra como falso positivo solo si el código muestra que el fallo no puede ocurrir; una reproducción fallida lo deja «no confirmado». Descartado: cerrar con una reproducción fallida. Revertir: cambiar el veredicto en `references/checks.md`.
- **D11 (supuesto)** — Sin embeddings ni infraestructura nueva: búsqueda en el tracker y en el código, más el juicio del agente. Descartado: un índice vectorial. Revertir: agregarlo en otro cambio.
- **D12 (supuesto)** — Antes de la aprobación, la skill comprueba si el tracker cierra los hijos al cerrar el padre y lo dice en la propuesta. Descartado: ignorar esa configuración. Revertir: quitar ese paso.
- **Supuesto** — Los comandos de cierre por tracker (duplicado y cancelado) y las escrituras de padre viven en `audit-issues/references/tracker-writes.md`, que enlaza la sección de padres de `create-issue/references/trackers.md` en lugar de repetirla. Descartado: agregarlos a las hojas `hive-core/tracker/trackers/*.md`, que leen otros flujos que nunca cierran duplicados. Revertir: moverlos a esas hojas.
- **Supuesto** — Con hasta 10 tickets en la pasada, el hilo principal los lee. Con más, un hijo `hive-research` por carril (módulo, etiqueta o proyecto) de hasta 15 tickets; solo el hilo principal escribe en el tracker. Descartado: leer todo en el hilo principal. Revertir: cambiar los umbrales. **Retired:** contradecía el criterio de la hoja `delegation` (más de 5 lecturas grandes); lo reemplaza el supuesto de delegación de abajo.
- **D18 (supuesto, delegación)** — La lectura se delega según el criterio de `hive-core/delegation/delegation.md` «When a child does the work», que la skill enlaza sin repetir, en carriles de hasta 15 tickets; solo el hilo principal escribe en el tracker. Descartado: un umbral propio de 10 tickets. Revertir: volver a un umbral propio. **Replaces:** el supuesto de los 10 tickets.
- **D14 (supuesto, grupo por causa raíz)** — Un grupo por causa raíz es un padre con hijos que un solo cambio cierra. `checks.md` lo declara como excepción a `breakdown.md` «The set», donde cada hijo se integra solo. Descartado: un ticket canónico que lista a los miembros y cierra a los demás como cubiertos, porque cerraría tickets ajenos contra D6 y borraría el historial de cada reporte. Revertir: cambiar la excepción en `checks.md`.
- **D15 (supuesto, exclusiones)** — Un ticket en progreso o con rama o PR abiertos se reporta sin escribirlo. Un ticket que ya tiene padre no se mueve: su grupo reutiliza ese padre, o el ticket recibe «related». Descartado: reemplazar el padre, porque GitHub y Linear admiten uno solo. Revertir: permitir el cambio de padre dentro del conjunto.
- **D16 (supuesto, efectos al ejecutar)** — La aprobación del conjunto nombra cada escritura por ticket, y esa es la instrucción sobre objetivos nombrados que pide la regla de efectos críticos. Bajo delegación desatendida la skill no escribe y reporta el conjunto sin resolver, como `breakdown.md`. En Jira, un cierre se propone solo si existe una transición de reapertura; si no, comentario. Descartado: escribir sin la aprobación por ticket. Revertir: cambiar esas líneas de `SKILL.md`.
- **D17 (supuesto)** — Un ticket «cubierto por un cambio» queda abierto y relacionado con ese cambio hasta que se integre, y el check de tipo y severidad solo informa, sin reetiquetar. Descartado: cerrarlo antes de la integración. Revertir: cambiar los checks 5 y 7 de `checks.md`.
- **D18 (supuesto, detalles por tracker)** — «Related» es `relatedTo` en Linear, el vínculo «Relates» en Jira y un comentario con enlaces cruzados en GitHub, que no tiene esa relación. «Ya corregido» cierra como completado (`--reason completed` en GitHub, estado de tipo `completed` en Linear) con el PR o el commit. Si el cierre automático de hijos no se puede leer, la propuesta dice «desconocido» y nombra dónde se configura. Descartado: cerrar lo ya corregido como cancelado. Revertir: cambiar `tracker-writes.md`.
- **Supuesto** — Al contar, la skill sugiere pasadas de hasta 30 tickets por carril, empezando por el carril con más Bugs. Descartado: una pasada de todo el backlog. Revertir: cambiar el número.
- **Supuesto** — Texto original de Hive, sin nombrar ni acreditar repositorios de referencia. Descartado: un crédito en la página de inspiración, que cambia solo con tu aprobación. Revertir: agregar el crédito con esa aprobación.
- **Supuesto** — El enrutamiento es una línea en `hive-mode` antes del paso 1 y un límite en la descripción de `create-issue` («not for auditing or grouping existing tickets, which audit-issues does»). Descartado: un playbook nuevo o una ruta en `ticket-status`, que es de solo lectura. Revertir: quitar esas dos líneas.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «Lo siguiente es, una skill audit-issues, lo que hara esta skill es revisar todos los issues, y agruparlos bajo un issue padre, la intención es ver qué issues tan relacionados por dominio o modulo/vista y puede sen resueltos en conjunto. Por ejemplo, se levantaron 5 issues sobre el uso del timezone, en lugar de atacar uno por uno, podemos agruparlos y atacarlos todos en conjunto, tambien puede auditar cuales NO aplican por un falso positivo o tema similar» · «aunque tambien haz research en linea sobre cual seria la mejor forma de aplicar esto» (2026-11-22)

Origen: una sesión de Codex (`01a204d1`) en un proyecto cliente abrió 36 tickets sueltos desde `audit-comments`, sin padre; varios se solapan (watch `watch/comment-audit-tracker-ask-and-parent`).

## Objective

Hoy Hive crea tickets (`create-issue`), los desglosa al crearlos y reporta el backlog agrupado (`ticket-status`, solo lectura), pero no tiene un procedimiento para auditar los tickets ya abiertos. Al terminar, la skill `audit-issues` cuenta los candidatos y sugiere una pasada, verifica cada ticket contra el código actual, propone cerrar los que no aplican y agrupa los demás bajo padres por causa raíz, dominio o módulo. Todo eso lo escribe en el tracker solo después de una aprobación del conjunto.

## Scope and acceptance

Dentro:
- la skill `audit-issues` con `references/checks.md` y `references/tracker-writes.md`;
- la línea de enrutamiento en `hive-mode`, con su orden frente a `ticket-status`, y el límite en la descripción de `create-issue`;
- el catálogo (`llms.txt`, `site/src/content/docs/skills.md`, `index.mdx`);
- las páginas del sitio que describen esas rutas (`flows/hive-mode.md`, `flows/create-issue.md`, `flows/status.md`);
- `instruction-resources.md` (la entrada nueva y las líneas 23 y 24) y `CHANGELOG.md`.

Fuera: correr la skill en un tracker real (los pilotos están en pausa, y escribiría en datos compartidos); cambios a `audit-comments` (su paso 5 ya pide padre y aprobación); el criterio de `--audit-session` del watch; scripts de clustering o embeddings (D11); las hojas `hive-core/tracker/trackers/*.md`.

- AC1 — Instalar Hive desde este cambio entrega la skill `audit-issues`: `~/.agents/skills/audit-issues/SKILL.md` existe con su `name` y `description`, `~/.claude/skills/audit-issues` enlaza a ella, y `llms.txt`, `site/src/content/docs/skills.md` e `index.mdx` la listan · false today because ni `content/skills/audit-issues/` ni esas entradas existen en `48824e10` · Observed in: la instalación local tras el merge (`hive update --source <checkout>`, `ls -l ~/.claude/skills/audit-issues`) y los tres archivos del catálogo en `development`
- AC2 — Cada decisión D3, D4, D5, D6, D8, D9, D10, D11 y D12 tiene en `audit-issues` un paso que la cumple · **Retired:** juntaba nueve decisiones y su observación era una tabla del propio agente; lo reemplazan AC5 y AC6.
- AC3 — Un pedido de auditar, agrupar, deduplicar o limpiar los tickets abiertos llega a `audit-issues`: el `hive-mode/SKILL.md` instalado lo envía allí antes del paso 1, y la descripción del `create-issue` instalado lo excluye nombrando `audit-issues` · false today because ninguno de los dos menciona `audit-issues` en `48824e10` · Observed in: `~/.agents/skills/hive-mode/SKILL.md` y `~/.agents/skills/create-issue/SKILL.md` instalados tras el merge
- AC4 — `audit-issues/references/tracker-writes.md` instalado da, por tracker, cómo cerrar como duplicado y como no planeado dejando el vínculo al ticket canónico: GitHub con `gh issue close <n> --duplicate-of <m>` y `--reason "not planned"`, Linear con `duplicateOf` y un estado de tipo `canceled`, y Jira con la transición y el vínculo de duplicado que el proyecto tenga, leídos antes de escribir. También dice cómo cerrar lo ya corregido como completado, cómo escribir «related» en cada tracker, cómo leer el cierre automático de hijos («desconocido» cuando no se puede leer), y que cada ticket se relee después de escribirlo, reportando las diferencias sin repetir la escritura. El movimiento de estado y el comentario enlazan `hive-core/tracker/trackers.md` «Move» y «Comment and attach» sin repetirlos · false today because ninguna referencia de Hive cierra duplicados · Observed in: `~/.agents/skills/audit-issues/references/tracker-writes.md` instalado tras el merge
- AC5 — El `audit-issues/SKILL.md` instalado:
  - cuenta los candidatos y sugiere una pasada antes de leer código;
  - muestra el conjunto completo con el veredicto, la evidencia y el nivel de grupo de cada ticket, la opción de separar uno y el estado del cierre automático de hijos;
  - no escribe nada antes de una aprobación del conjunto, y después escribe de uno en uno y se detiene ante el primer fallo;
  - a un ticket ajeno con veredicto de cierre solo le comenta;
  - bajo delegación desatendida no escribe.

  · **Replaces:** AC2 · false today because la skill no existe · Observed in: `~/.agents/skills/audit-issues/SKILL.md` instalado tras el merge, con la línea de cada punto citada en «Evidence» del checkpoint
- AC6 — El `audit-issues/references/checks.md` instalado define:
  - los checks 1 a 14 de `design.md`;
  - que un duplicado exige el mismo error, los mismos pasos o la misma causa en el código, y que sin eso la relación es «related»;
  - que una reproducción fallida deja el Bug «no confirmado» y que cerrarlo como falso positivo exige que el código muestre que el fallo no puede ocurrir;
  - las exclusiones de tickets en progreso, con rama o PR, o con padre;
  - la excepción del grupo por causa raíz a `breakdown.md`, nombrándola.

  · **Replaces:** AC2 · false today because el archivo no existe · Observed in: `~/.agents/skills/audit-issues/references/checks.md` instalado tras el merge, con la línea de cada punto citada en «Evidence» del checkpoint

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Base `development` en `48824e10` (PR #602, renombre `audit-*` integrado), checkout limpio · `git log`, `git status` · 2026-11-22.
- Modelo de skill de auditoría: `content/skills/audit-comments/SKILL.md` (54 líneas; añadirla tocó solo su carpeta, y el catálogo vino aparte). Registro: `llms.txt:53-54`, `site/src/content/docs/skills.md:46-47`, `index.mdx:106-107`, `_support/docs/architecture/instruction-resources.md:28`.
- Reutiliza: `create-issue/references/breakdown.md` (aprobación del conjunto, escrituras de una en una, alto ante un fallo), `create-issue/references/trackers.md` «Parent and children» y «After writing», `create-issue/references/ticket-types.md` (lista de listo), `create-issue/references/existing-ticket.md` paso 3 (autoría), `hive-mode/playbooks/ticket-status.md` (agrupación por módulo y orden de prioridad).
- Hojas de tracker: `hive-core/tracker/trackers/{linear,github,jira}.md` no tratan cierres por duplicado · `rg` · 2026-11-22.

- La herramienta MCP de Linear `save_issue` expone `parentId`, `duplicateOf`, `relatedTo`, `state` · esquema de la herramienta en esta sesión · presente · 2026-11-22
- Linear: con el ajuste del equipo, cerrar el padre cierra los hijos pendientes · https://linear.app/docs/parent-and-sub-issues (investigación delegada) · documentado · 2026-11-22
- Linear: si `duplicateOf` cambia solo el estado, y si mover un ticket bajo un padre le cambia proyecto o ciclo · ninguna lectura sin escribir lo muestra · not measured; la skill relee cada escritura (T1), y la primera ejecución real lo observa · 2026-11-22
- GitHub CLI: `gh issue close` acepta `--duplicate-of` y `--reason {completed|not planned|duplicate}` · `gh issue close --help`, gh 2.105.0 · presente · 2026-11-22
- GitHub: hasta 100 sub-issues por padre y 8 niveles · https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/adding-sub-issues · documentado · 2026-11-22
- Jira: el tipo de vínculo de duplicado y la transición de cancelado dependen del proyecto · not measured; la skill los lee antes de escribir (T1) · 2026-11-22
- GitHub CLI: `gh issue edit` acepta `--parent`, `--remove-parent` y `--add-sub-issue` · `gh issue edit --help`, gh 2.105.0 · presente · 2026-11-22
- Linear: el ajuste que cierra los hijos con el padre no sale por MCP · `get_team` del MCP de Linear sobre un equipo real devuelve solo id, clave, nombre, visibilidad y fechas · ausente por MCP; la skill dice «desconocido» y nombra el ajuste del equipo (T1) · 2026-11-22
- Jira: si se puede leer, desde un ticket abierto, que el flujo tiene una transición para reabrir desde cancelado (la lectura de transiciones devuelve las del estado actual; el workflow puede pedir permisos de administración) · not measured; si no se puede leer, la skill solo comenta (T1) · 2026-11-22
- GitHub y Jira: si cerrar el padre cierra los hijos · not measured; la skill dice «desconocido» salvo que el tracker lo muestre (T1) · 2026-11-22

- `go test ./...` · pass, 3 min 18 s · 2026-11-22
- `go test ./tests/content/...` · pass, 9.9 s; pass también con esta carpeta · 2026-11-22
- `npm --prefix site run build` · pass, 20 páginas, 2.4 s · 2026-11-22
- `rg -n 'audit-issues' content/skills/hive-mode/SKILL.md content/skills/create-issue/SKILL.md` · sin coincidencias (exit 1) en la base · 2026-11-22
- `python3 -I ~/.agents/skills/flow-plan/scripts/check_change.py _support/openspec/changes/audit-issues` · exit 0, sin salida · 2026-11-22

## Effects

None: the tasks take no effect on an external system.

## Delivery

- **Dónde A** — Worktree `.claude/worktrees/audit-issues`, rama `feat/audit-issues` desde `development`. La carpeta del plan se mueve allí al empezar el build.
- **Delivery A** — `automatic`: revisión, push, PR y merge a `development` sin detenerse, y después se refresca la instalación local, porque existe el marcador de mantenedor.
- **Review A** — `Native: Claude Code /code-review`.
- No se autoriza ninguna escritura en un tracker. La skill no se ejecuta en este trabajo.

## Review status and progress

2026-11-22 — Plan escrito. La revisión `hive-review-plan` (dominio guidance) dio 1 P1, 8 P2 y 5 P3. Se aplicaron:
- H1: los 14 checks se escribieron en `design.md`.
- H2: la autoría y la lista de listo ahora enlazan `ticket-types.md`.
- H3, H4, H5 y H6: nuevos supuestos.
- H7: los padres nuevos siguen `create-issue`.
- H8: se agregaron páginas del sitio a T3.
- H9: nuevos hechos externos.
- H10: AC2 se retiró a favor de AC5 y AC6.
- H11: líneas de verificación.
- H12: enlaces a `trackers.md`.
- H13: orden frente a `ticket-status` en T2.
- H14: «ya corregido» cierra como completado.

2026-11-22 — La segunda lectura del revisor resolvió H1–H14 y encontró N1–N7. Se aplicaron todos:
- N1: el check 7 solo informa.
- N2: «cubierto por» deja el ticket abierto y relacionado con el cambio.
- N3: hecho de Jira `not measured`.
- N4: la decisión registrada es un falso positivo con su enlace.
- N5: T1 queda en el hilo principal, con su razón en `tasks.md`.
- N6: nombres privados generalizados.
- N7: «Evidence» es la del checkpoint.

No cambió ningún criterio ni el delta, así que no hace falta otra ronda.

2026-11-22 — `/code-review` (high) sobre `0dbc206d` dio 4 hallazgos importantes, 1 previo y 4 menores; se corrigieron todos en una ronda. El canónico de un duplicado está abierto. Un ticket que el conjunto cierra no se reutiliza como padre. Los hijos reciben hasta 15 tickets y devuelven sus filas en la respuesta. La reproducción corre solo en un entorno local o desechable. La base se lee del remoto. El comentario de `covered by` tiene su lugar en la aprobación y en el orden de escrituras. El check 7 no agrega escrituras. `hive-mode` suma «triage» y «consolidate». Se quitaron dos absolutos. El patrón del ledger que un rol de solo lectura no puede escribir sigue en `audit-comments` y `audit-tests` (hallazgo previo, fuera de este cambio).
