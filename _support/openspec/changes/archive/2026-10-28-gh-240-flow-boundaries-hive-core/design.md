# Diseño — gh-240

## Contexto verificado (`development` @ `4ef04e01`, 2026-10-27)

**Tamaños** (`wc -c`, en bytes):

| Archivo | Tamaño | Margen hasta 16.000 |
| --- | --- | --- |
| `global.md` | 42.001 | sin tope (no es una skill) |
| `flow-close/SKILL.md` | 15.850 | 150 |
| `flow-close/references/reporting.md` | 15.995 | 5 |
| `flow-build/SKILL.md` | 15.591 | 409 |
| `flow-plan/references/plan-format.md` | 15.027 | 973 |
| `flow-verify/references/ui-gate.md` | 14.965 | 1.035 |

Las cuatro referencias que migran suman 27.930 B.

**Despliegue.**
- Las skills se descubren por glob (`tooling/management/plan.go:230`); no hay ninguna lista cerrada de nombres en el manager.
- Toda skill necesita `SKILL.md` (`plan.go:414-418`). Si su frontmatter lleva `: ` o ` #`, va entre comillas (`tests/content/frontmatter_test.go`).
- Los enlaces `skill:<owner>/<path>` de `global.md` y de los roles se reescriben al instalar (`integrations/mdlinks/mdlinks.go`). Los enlaces relativos entre skills se copian tal cual.
- `tooling/management/references.go` valida, antes de escribir, que cada enlace resuelva contra la release. No valida anclas.

**Radio de impacto de mover las cuatro referencias.**
- **Archivos:** 49, fuera del archivo histórico: 32 en `content/skills`, más `global.md`, `content/agents/docs/hive-write-docs.md:19`, 8 en `site/`, 3 en `_support/docs/architecture/`, `README.md:355`, `tests/content/reading_footprint_test.go:53,58,81-82` y `tooling/cli/doctor_project_test.go:451` (con el comentario de `doctor_project.go:24`).
- **Nombre:** `hive-core` tiene 9 caracteres, igual que `flow-plan`, así que reescribir los enlaces no cambia el tamaño en bytes de `reporting.md`.
- **Anclas en uso que deben conservar su encabezado:** `questions.md#decide-whether-a-question-is-needed`, `questions.md#make-user-dependent-work-self-sufficient`, `questions.md#pause-and-handle-the-answer`, `effect-authorization.md#what-an-answer-covers` y `project-settings.md#later-completion-point`.
- **Enlaces de `site/` a `master`:** `site/` tiene 13 URLs absolutas a `blob/master/.../flow-plan/references/`. Pasan a la ruta nueva. El sitio se publica desde `master` en cada release (`CONTRIBUTING.md:136`), así que el sitio publicado no se rompe. Solo la vista en GitHub de `development` muestra enlaces rotos hasta la release, y eso se anota en el PR.

**Hooks que ya existen.**
- **Grok:** un `PreToolUse` read-whole, recurso gestionado (`integrations/grok/hook.go`, `tooling/cli/grok_hook.go`).
- **OpenCode:** el plugin `hive.context-loader`, que se instala inactivo y tiene dos roles fijos.
- **Otros hosts:** ninguno de los dos toca Claude Code, Codex, Pi ni Cursor.
- **Política:** `AGENTS.md:61` admite un mecanismo cuando hay una limitación verificada y documentada. `repository-and-distribution.md` dice que la comparación de septiembre "is historical, not a current prohibition on hooks".

**Evidencia sobre carga bajo demanda**, de la [investigación](../../../sessions/2026-10-27-flow-boundary-analysis/flow-boundary-analysis.research.md):
- **Hive legacy (OBSERVED):** el router se invocó 0 de 9 y 0 de 10 veces, y forzar la lectura de una referencia no hizo que la regla gobernara.
- **Proyectos de referencia:** ninguno pone reglas críticas en una skill que el modelo elija cargar.
- **Hosts:** ninguno de los seis garantiza que una skill se cargue. Claude Code relee el `CLAUDE.md` del proyecto tras compactar (sin verificar para el global de usuario) y recorta cada skill a 5.000 tokens, 25.000 en total.
- **Host sin Hive:** cubre solo lo genérico.

## Diseño elegido

### Tres capas (D8-B, D9-A, D12-A)

1. **Capa 1, `content/guidance/global.md`** (unos 15,5 KB):
   - **Contenido:** los invariantes que deben actuar sin lectura previa, cada uno con su forma mínima completa.
   - **Punteros:** cada uno es un disparador literal, "Before <acto observable>, read <ruta>".
   - **Ejemplos de invariantes:** idioma; que una pregunta no autoriza; los efectos más allá de lo local y su clase crítica; proyecto actual; datos externos; privacidad ante terceros; preservación; secretos; evidencia; puertas; definición de cambio mecánico; compactación; colocación básica de `_support` (exigida por la spec); cadencia de memoria.
2. **Capa 2, skill `hive-core`:**
   - **`SKILL.md`:** índice de condiciones de lectura. Su descripción dice qué contiene la skill y cuándo se lee, sin resumir sus procedimientos.
   - **`references/`:** las cuatro migradas, más las referencias nuevas que recogen el procedimiento que sale de `global.md`. Cada archivo mide 16.000 B o menos, y los contenidos se reparten así:
     - `language.md`: reanudar o reenviar registros, lista de identificadores, i18n.
     - `third-party-requests.md`: User-Agent, contacto y el caso "te piden identificarte".
     - `secrets.md`: formas de enmascarar y la exposición que queda solo en la sesión. La exposición en un lugar duradero o compartido se queda como invariante en la capa 1.
     - `memory.md`: corrección de entradas y conflictos entre memorias.
     - `mechanical-change.md`: solo la comprobación de si el código también sirve a conductas que el pedido no nombra. La definición se queda en la capa 1, porque `flow-verify/SKILL.md:20`, `tdd.md:9` y `review-decisions.md:9` la citan como "as the global guidance defines it".
     - `evidence-tools.md`: Context7 y CodeGraph (U8).
   - **Reparto final:** los nombres y el reparto definitivo los fija T4 dentro de estos límites.
3. **Capa 3, reinyección tras compactar:** diseñada para los seis hosts. Solo se activa donde T1 compruebe que se pierde la capa 1 o `hive-core`. La forma de entrega sigue el precedente:
   - si el host lee el hook de una carpeta (como Grok), va como recurso gestionado;
   - si hay que editar la configuración del usuario (`settings.json` o `config.toml`), va como capacidad opcional, igual que `codex-wait-floor`.

   **Descartada el 2026-10-28 por D16-B:** Claude Code y Codex reinyectan el global tras compactar, y en las sesiones revisadas no se perdió ninguna regla (ver `proposal.md`, Entrega). El diseño de arriba queda como referencia por si en otro host aparece una pérdida.

**Reglas que pasan a skills existentes:**

| Regla | Destino |
| --- | --- |
| Agrupación del backlog | `backlog-report.md` |
| Procedimiento de hallazgos incidentales (doble búsqueda, preguntar una vez al cierre) | `incidental-findings.md` |
| "Record promotion…" | `flow-deploy` |
| Excepción de la revisión del plan | `plan-review.md` |
| Identidad de Engram: el procedimiento. En la capa 1 se queda el invariante "una identidad de Engram solo se crea o cambia a través de `engram-init-workspace`; una identidad en conflicto bloquea toda escritura", porque es preservación según la spec | `engram-init-workspace` |
| Continuidad de los registros (reconciliar al reanudar) | `artifact-layout.md`, con puntero desde `change-records.md` |
| Pasos de limpieza. La autorización de borrar ramas y worktrees, con sus límites, se queda en la capa 1 | `flow-close` |
| Detalle de los hijos | `delegation.md` |
| Excepciones de las puertas | `flow-verify` |

**Se descartan**, identificados por su texto en `4ef04e01` (los cubre el host, se repiten o son metatexto):
- "Identify the requested outcome, scope, authorized effects and targets, and completion evidence…", de Scope. El host ya lo hace, y "discover facts yourself" queda en el invariante de preguntas.
- La frase sobre delegación desatendida dentro del párrafo "A question, investigation, or proposal does not authorize implementation…". La cubre el párrafo "Unattended work requires an explicit delegation…", que se queda en la capa 1 en forma corta.
- "Distinguish inspected facts, source claims, inference, and recommendations…", de Evidence. Ya está en `flow-research` y en el rol `hive-research`.
- "Use applicable tests and checks rather than adding ceremony or unrelated work.", de Proportionality. Es metatexto.
- "Planned work and requirements live in the `openspec` directory…", de Specs. Ya está en `change-records.md`.
- "For requested organization of existing support material…", de Retention. La descripción de `workspace-conventions` ya lo dice.

**También se descartan** (detectado al verificar T4): el ejemplo "one among several reasonable fixes" de la regla de defectos, porque la regla conserva el criterio "a choice the user could make differently"; y el motivo "since that output stays in the main thread's context" de la escalada a `hive-research`, porque solo explica la regla, que se conserva.

**No se descartan, contra lo que decía la primera versión:** la glosa de IDs (U6), "A question, investigation, or proposal does not authorize implementation" (invariante) y "Local retention is not versioning; eligibility for Git is not authorization to commit or push" (invariante de autorización). Las tres se quedan en la capa 1.

**Lista de unicidad:** T4 la escribe aquí antes de editar `global.md`. Lleva una frase textual y distintiva de cada regla que sale, con su destino. Es la entrada del check de AC5.

**Dudosas resueltas por D12-A:**

| Regla | Capa |
| --- | --- |
| U1, cadencia de memoria | 1 |
| U2, permiso para mover tickets | Invariante en la capa 1 y detalle (sub-issues, `Completion`) en `hive-core`, por D17-A, que reemplaza a D12-A solo para U2 |
| U3, no inventar restricciones | 1 |
| U4, nombrar los temas abiertos en cada reporte (una línea) | 1 |
| U5, arreglar sin preguntar un defecto que bloquea | 1 |
| U6, glosar IDs (unos 90 B) | 1 |
| U7, construir solo lo pedido | 1 |
| U8, Context7 y CodeGraph | `hive-core` |
| U9, colocación de `_support` | 1, por la spec |

**Reglas sin acto observable que se quedan como invariante en la capa 1:**
- escalar a delegación en la sexta búsqueda;
- compactación;
- fallo de una integración;
- secreto expuesto;
- activación por silencio;
- pregunta de ruta y fallo reportado;
- definición de cambio mecánico.

### Fallos de la réplica que este cambio aplica

| Fallo | Destino | Tarea |
| --- | --- | --- |
| K1 | "Una parada de validación no es pausa" pasa a `flow-close` "When to run it"; build y verify apuntan a ella | T8 |
| K2 | `reporting.md` es la definición completa de la línea de contexto y `global.md` deja el invariante mínimo | T4 |
| K3, K23 | El procedimiento de incidentales pasa a `incidental-findings.md`; se cita el paso de memoria por nombre | T4 y T8 |
| K4 | `reporting.md:96` queda como puntero a flow-research | T2 |
| K5 | Las sondas pasan a `diagnosis.md` | T9 |
| K6 | El momento de la revisión dedicada pasa a `review-and-gates.md` | T6 |
| K7 | Regla general de commits en `flow-build`; puntero en el Prepare de verify | T7 y T6 |
| K8 | `delivery-decisions.md:24` se queda en plan | — |
| K9 | `review-and-gates.md` cubre los hallazgos de seguridad antes del merge | T6 |
| K10 | Puntero de build a la inspección del implementador en `ui-gate.md` | T7 |
| K11, K12 | La regla de nueva superficie y el estado del build devuelto a plan van en `scope-changes.md` | T9 |
| K13 | Se quita "including deployment of existing code" | T7 |
| K14 | La regla de promoción pasa a `flow-deploy` paso 4 | T5 escribe en deploy; T9 deja el puntero en `delivery-decisions.md` |
| K15 | El registro de la promoción pasa a `flow-deploy`; una release que publica es efecto crítico | T4 (global) y T5 |
| K16 | Sin cambio | — |
| K17 | Corrección en `project-settings.md` (estado final de `Completion`) | T9 |
| K18 | Corrección en `project-settings.md:41` y `plan-format.md:30` | T9 |
| K19 | Corrección en `scope-changes.md:13,35` | T9 |
| K20 | Se quita "and its diagnosis reference" de `global.md:78` | T4 |
| K21 | Puntero en la entrada de deploy | T5 |
| K22 | Nueva redacción de `global.md:76,79` | T4 |

### Por qué este diseño

- **Mover solo procedimiento, no invariantes:** respeta la evidencia propia del legacy y la spec "Portable kernel".
- **Migración antes de los arreglos (D10-A):** así cada texto corregido se escribe una sola vez, ya en su hogar final.
- **`hive-core` en lugar de mover todo a `flow-plan`:** las referencias se usan en todas las etapas, y leerlas no debe parecer que arranca la planificación.

**Alternativa descartada:** una skill core cargada por su descripción, sin invariantes en la capa 1. Ningún host garantiza esa carga.

### Contratos que cambian

- **Rutas:** las de las cuatro referencias, ahora bajo `hive-core/references/`. Las anclas se conservan.
- **Guía global:** el contenido y la forma de `global.md`.
- **Spec:** un requisito nuevo en `context-delivery` (ver `specs/`).
- **Instalación:** `hive update` instala `hive-core` y retira de `flow-plan` las cuatro referencias. Las sesiones abiertas que ya leyeron la ruta vieja tienen que reiniciarse.

### Lista de unicidad (AC5, escrita en T4; actualizada tras /code-review)

Cada frase debe aparecer exactamente una vez bajo `content/`, en el destino de la regla que salió de `global.md`. Check: `rg -F -c --no-filename '<frase>' content | awk '{s+=$1} END {print s}'` imprime 1.

- "Apply this when resuming a record written in another language" → `skills/hive-core/references/language.md`
- "identifiers introduced in specifications" → `skills/hive-core/references/language.md`
- "even within session-language prose" → `skills/hive-core/references/language.md`
- "use English semantic i18n keys" → `skills/hive-core/references/language.md`
- "A `User-Agent` you set is generic" → `skills/hive-core/references/third-party-requests.md`
- "a pattern that misses one format leaks it" → `skills/hive-core/references/secrets.md`
- "its installation or an unlocked vault grants nothing" → `skills/hive-core/references/secrets.md`
- "reference documents mask all but what identifies the resource" → `skills/hive-core/references/secrets.md`
- "Do not mention the exposure or rotation unless the user raises credentials" → `skills/hive-core/references/secrets.md`
- "deleting the local copy does not contain it" → `skills/hive-core/references/secrets.md`
- "keep in the host's own memory only what concerns that host" → `skills/hive-core/references/memory.md`
- "Challenge an unsupported premise without flattery or automatic agreement" → `skills/hive-core/references/questions.md`
- "A host's automatic capture may still copy project facts" → `skills/hive-core/references/memory.md`
- "prefer Context7 as the first lookup" → `skills/hive-core/references/evidence-tools.md`
- "run `codegraph sync -q` first" → `skills/hive-core/references/evidence-tools.md`
- "path alias the root `tsconfig.json` does not declare" → `skills/hive-core/references/evidence-tools.md`
- "make three or more read-only calls that script can reach" → `skills/hive-core/references/tool-scripts.md`
- "save the failing output before rerunning" → `skills/flow-close/references/incidental-findings.md`
- "Put style or wording nits, and suspected issues marked unconfirmed" → `skills/flow-close/references/incidental-findings.md`
- "queried across all non-terminal states rather than by recent activity" → `skills/flow-research/references/backlog-report.md`
- "including other tickets, for the rest of the session" → `skills/hive-core/references/questions.md`
- "change each of its sub-issues that the same work covers to the same state" → `skills/hive-core/references/project-settings.md`
- "later ones receive changes only by promotion" → `skills/hive-core/references/project-settings.md`
- "is part of planning and needs no separate gate approval" → `skills/flow-plan/references/plan-review.md`
- "do not infer sibling scope, modify global Git ignores, or migrate memory" → `skills/engram-init-workspace/SKILL.md`
- "sends the parent at most one message" → `skills/flow-research/references/delegation.md`
- "retain superseded statements only as clearly labeled history" → `skills/workspace-conventions/references/artifact-layout.md`
- "this session's own model use, its children, the project's CI" → `skills/flow-build/references/verification.md`
- "a child still reads what its brief names" → `skills/flow-research/references/delegation.md`
- "leave the delivery and review questions to `flow-build`" → `skills/flow-research/SKILL.md`
- "An answer to a pending question is not a topic" → `skills/hive-core/references/scope-changes.md`
- "Record the promotion in the project's mechanism, without reopening the ticket" → `skills/flow-deploy/SKILL.md`
- "an existing documented destination such as a documentation repository" → `skills/workspace-conventions/references/artifact-layout.md`
