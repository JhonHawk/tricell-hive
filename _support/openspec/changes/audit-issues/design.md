# audit-issues — diseño

## Enfoque

Solo hay una forma sensata: una skill nueva `audit-issues`, porque el usuario la nombró y porque ninguna de las existentes cubre esto. `create-issue` crea y valida un ticket a la vez; `ticket-status` solo reporta y no escribe. Una cuarta función dentro de `create-issue` la mezclaría con la creación y contradice el pedido.

La skill no cambia archivos del repositorio. Escribe en el tracker desde el hilo principal y delega la lectura de tickets y código a hijos `hive-research` cuando la pasada supera el criterio de [delegation](../../../../content/skills/hive-core/delegation/delegation.md) «When a child does the work», en carriles de hasta 15 tickets. Por eso no se enruta por `flow-build`: termina con su propuesta, la aprobación y las escrituras, como `create-issue`.

## Estructura

- `SKILL.md` (procedimiento, ≤16 KB):
  1. contar candidatos y sugerir la pasada (D3);
  2. leer cada ticket con sus comentarios;
  3. aplicar los checks 1–8 por ticket con su evidencia;
  4. formar grupos con los checks 9–14 y declarar su nivel (D9);
  5. comprobar el cierre automático de hijos (D12);
  6. mostrar el conjunto y pedir una sola aprobación;
  7. escribir de uno en uno y releer;
  8. reportar.

  También define:
  - que bajo delegación desatendida no escribe y reporta el conjunto sin resolver;
  - que la aprobación nombra cada escritura por ticket y es la instrucción sobre esos objetivos que pide la regla de efectos críticos.
- `references/checks.md`:
  - los 14 checks de abajo, con el umbral de evidencia de cada veredicto (D8, D10) y el estado «no confirmado»;
  - las exclusiones;
  - la excepción de autoría de D6, que nombra lo que cambia de `create-issue/references/existing-ticket.md` paso 4;
  - la excepción del grupo por causa raíz a `create-issue/references/breakdown.md` «The set».
- `references/tracker-writes.md`:
  - cómo mover un ticket bajo un padre (enlaza `create-issue/references/trackers.md` «Parent and children»);
  - cómo cerrar por tracker como duplicado, como no planeado o como ya corregido;
  - cómo escribir «related» por tracker;
  - cómo leer el cierre automático de hijos;
  - el comentario de evidencia, que enlaza `hive-core/tracker/trackers.md` «Comment and attach»;
  - el movimiento de estado, que enlaza «Move»;
  - la relectura y el alto ante un fallo (enlaza `breakdown.md` «Writes»).

## Los catorce checks

Por ticket. Cada veredicto lleva su evidencia: archivo y línea, commit o PR, o el ticket canónico.

1. **Vigente.** El código que cita el ticket (ruta, símbolo o comportamiento) existe en la base actual y se comporta como dice. Se lee contra la base, no contra el hilo del ticket: muchos reportes quedan obsoletos cuando commits posteriores quitan el código que citan ([Li y Zhong](https://zenodo.org/records/10405536)).
2. **Falso positivo.** El síntoma no se sostiene.
   - Se propone cerrar solo si el código muestra que el fallo no puede ocurrir (D10).
   - Un Bug cuya reproducción falla queda «no confirmado», porque las herramientas actuales reproducen entre 37 y 44 % de los casos ([Otter](https://arxiv.org/html/2502.05368v1), [AssertFlip](https://arxiv.org/html/2507.17542v2)).
   - Se reproduce solo cuando el veredicto cerraría un Bug (D5), y solo en un entorno local o desechable; si los pasos necesitan datos compartidos, el Bug queda «no confirmado».
3. **Ya corregido.** Un PR integrado o un commit cambió la ruta que falla, o una prueba la cubre. Se cita.
4. **Duplicado.** Se busca un ticket con el mismo resultado, abierto o cerrado, en todo el tracker y no solo en la pasada, como hace `create-issue` en su paso 2.
   - Solo es duplicado con evidencia firme: el mismo texto de error, los mismos pasos o la misma causa demostrada en el código (D8).
   - Sin esa evidencia, la relación es «related».
   - El canónico está abierto: si el ticket que coincide está cerrado, el abierto queda abierto y relacionado con él, como `create-issue` paso 2 relaciona un defecto que volvió.
   - Motivo: los estudios miden si el duplicado aparece entre los primeros resultados, no qué tan correcta es la fusión, y sobreestiman el rendimiento real entre 17 y 42 % ([Rakha et al.](https://sail.cs.queensu.ca/data/pdfs/TSE2017_RevisitingThePerformanceEvaluationOfAutomatedApproachesForTheRetrievalOfDuplicateIssueReports.pdf)).
   - La búsqueda de texto no rinde peor que los modelos neuronales ([Zhang et al.](https://ar5iv.labs.arxiv.org/html/2212.00548)), así que no hace falta un índice (D11).
5. **Cubierto por un cambio en curso.** Una carpeta de cambio abierta o un PR abierto ya cubre el ticket. El ticket no se cierra: se relaciona con el cambio y queda abierto hasta que el cambio se integre, que es cuando `hive-core/tracker/ticket-state.md` lo termina.
6. **Contradice una decisión registrada.** El registro de decisiones o un ADR del proyecto dice que el comportamiento es intencional. Es un falso positivo con esa evidencia: se propone cerrar como no planeado, con el enlace a la decisión.
7. **Tipo y severidad.** La etiqueta corresponde al efecto real; por ejemplo, un hallazgo de autorización etiquetado como deuda técnica. Este check solo informa: el reporte nombra la etiqueta que correspondería, y en un ticket ajeno el comentario lo dice. No se reetiqueta ningún ticket.
8. **Listo.** Se aplica la lista de listo de `create-issue/references/ticket-types.md`. Si falla, el veredicto es «no listo»: se reporta y no se corrige aquí.

**Exclusiones de escritura:**
- Un ticket en progreso, o con rama o PR abiertos, se reporta pero no se escribe.
- Un ticket que ya tiene padre no se mueve: su grupo reutiliza ese padre, o el ticket recibe una relación «related».

Por grupo:

9. **Clave.** Primero causa raíz, luego concepto de dominio (por ejemplo timezone) y luego módulo o vista. Cada grupo declara su nivel (D9). Un tema compartido sin causa ni módulo común da una relación «related», no un grupo: ninguna fuente respalda agrupar solo por tema, y la práctica agrupa por área ([GitLab](https://handbook.gitlab.com/handbook/engineering/infrastructure-platforms/developer-experience/bug-prioritization/)).
10. **Cabe en un cambio.**
    - Un grupo por causa raíz es un padre con hijos que un solo cambio cierra: un repositorio, un PR y como máximo 7 criterios (`create-issue` paso 3). Es una excepción a `breakdown.md` «The set», donde cada hijo se integra solo, y `checks.md` la dice ahí.
    - Si no cabe, el grupo pasa a ser por dominio o por módulo, con hijos que se entregan por separado.
11. **Orden.** Las relaciones blocked-by dentro del grupo.
12. **Padre existente.** Si un ticket o un paraguas ya cubre el grupo, se reutiliza, salvo que su propio veredicto lo cierre. Un padre nuevo sigue los pasos 2 y 4 a 6 de `create-issue`.
13. **Prioridad entre grupos.** El orden de `hive-mode/playbooks/ticket-status.md` «Report» punto 4.
14. **Sueltos.** Un ticket que no comparte causa ni área con otro queda solo.

## Veredictos por ticket

| Veredicto | Escritura |
| --- | --- |
| `vigente` | Ninguna, salvo moverlo bajo su padre |
| `cerrar: falso positivo` (incluye un comportamiento que una decisión registrada declara intencional) | Cierre como no planeado o cancelado, con la evidencia o el enlace a la decisión |
| `cerrar: ya corregido` | Cierre como completado, con el PR o commit |
| `cerrar: duplicado de <canónico>` | Cierre como duplicado, enlazado al canónico |
| `cubierto por <cambio>` | Relación con el cambio; el ticket sigue abierto |
| `no confirmado` | Ninguna |
| `no listo` | Ninguna |

Todos los cierres llevan un comentario con la evidencia.

Un ticket que el usuario no escribió y tiene veredicto de cierre no se cierra: recibe un comentario para su autor (D6). La autoría se decide como en `existing-ticket.md` paso 3; también cuenta como propio el ticket que el usuario declara suyo.

En Jira, un cierre se propone solo si se puede leer que el flujo tiene una transición para reabrir; si no la tiene o no se puede leer, el ticket recibe un comentario.

## Reutilización

Aprobación del conjunto y escrituras: `create-issue/references/breakdown.md`. Redacción de un padre nuevo: `create-issue` pasos 2 y 4 a 6. Representación de padre e hijos: `create-issue/references/trackers.md`. Movimiento de estado y comentarios: `hive-core/tracker/trackers.md`. Lista de listo: `ticket-types.md`. Prioridad: `ticket-status` «Report». Formato de la ronda: `hive-core/messages/questions.md`. Delegación: `hive-core/delegation/delegation.md`.
