# ux-review-loop — Revisión UX en una sola pasada

| Campo | Valor |
| --- | --- |
| Estado | Completado · integrado en development (PR #366, #367 y #368) |
| Tracker | Ninguno |
| Git | tricell-hive · base `development` en `47362771` · rama propuesta `fix/ux-review-loop` |

## Decisiones

2026-11-04
- **D1 A** — ¿Cómo recibe el revisor UX los estados que debe revisar? El brief dice cómo llegar a cada estado (ruta del fixture, cuenta, cómo forzar el error, valor largo). El revisor captura lo que falta y lista como «no cubierto» lo que no pudo alcanzar, sin reprobar por eso.
- **D2 A** — ¿Quién relee una corrección de UI? Si la corrección responde a un hallazgo del revisor UX, se le manda un mensaje de seguimiento al mismo hijo, limitado a lo corregido. Si el host ya no lo conserva, se lanza un hijo nuevo limitado a esas regiones. Las demás correcciones las relee el principal, como hoy.
- **D3 A** — ¿Sobre qué emite el revisor su veredicto? Solo sobre lo que introduce el cambio en las regiones revisadas. Los defectos previos van en una lista aparte, con su prioridad, y no cambian el veredicto.
- El usuario aprobó adoptar las tres correcciones al pedir el plan. Su admisión según «Measurement» queda en la sección «Contexto verificado».

## Solicitud

El usuario, el 2026-11-04, después del análisis de las sesiones Codex de CAM-8, CAM-9 y COM-920: «/flow-plan sobre los tres hallazgos».

Los tres hallazgos, H16, H17 y H18 de ese análisis:

- **Cobertura al lanzar.** El revisor UX arranca sin los estados que su rol exige revisar siempre (error, valores largos, permisos) y responde «no aprobado por cobertura incompleta». El hilo principal fabrica la evidencia y vuelve a preguntar.
- **Relectura con un revisor nuevo.** Después de corregir, se lanza otro `hive-review-ux` completo en vez de releer solo lo corregido.
- **Veredicto mezclado con defectos previos.** El veredicto sale «no aprobado» solo por defectos que ya existían, que `flow-verify` envía al usuario y no a la ronda de correcciones.

## Objetivo

Hoy un `hive-review-ux` tarda de 1 a 2,5 min en su primer turno, pero la revisión ocupa entre 10 y 20 min de reloj y entre 6 y 24 M de tokens por idas y vueltas. Al terminar, el revisor recibe lo que necesita en su primer turno, la relectura de una corrección no lanza otro revisor completo y su veredicto distingue lo que introduce el cambio de lo que ya existía.

## Alcance y aceptación

**Dentro:**

- el rol `hive-review-ux`;
- el «Reviewer brief» de `delegation.md`;
- los pasos 2 y 3 de `flow-verify`;
- la página `site/.../flows/verify.md`;
- un delta de `workflow-lifecycle`.

**Fuera:**

- qué revisa el revisor UX: regiones, anchos, temas, auditoría axe e integridad de layout;
- cuándo se lanza: uno por cambio de UI;
- la revisión de código y su segunda vuelta;
- la guía global, que no cambia.

- AC1 — El «Reviewer brief» exige que el brief de un revisor UX incluya, junto a las capturas, cómo alcanzar cada estado que su rol revisa siempre: carga, vacío, error y los valores reales más largos (`hive-review-ux.md:10,12`). Si un estado no se puede alcanzar, el brief lo dice. · Hoy es falso porque `delegation.md:40` solo pide las capturas.
- AC2 — El rol indica que un estado que el revisor no puede alcanzar con los medios del brief se lista como «no cubierto», con lo que lo impidió. No hace fallar el veredicto ni lleva al revisor a detenerse a preguntar. · Hoy es falso porque el rol no menciona estados no cubiertos y en 4 casos el veredicto fue «no aprobado por cobertura».
- AC3 — `flow-verify` indica que el principal cubre en su propio recorrido cada estado «no cubierto», o lo reporta como no verificado. Lo reporta en la ronda de validación o, cuando la entrega no se detiene para validar, en el reporte final. El recorrido de despliegue (`deployment-checks.md`) lo reporta como no verificado en su reporte. · Hoy es falso porque ninguno de los dos trata ese caso.
- AC4 — `flow-verify` paso 3 indica cómo se relee una corrección:
  - de un hallazgo del revisor UX, con un mensaje de seguimiento al mismo hijo, limitado a las regiones y estados corregidos;
  - si el host no tiene una herramienta para enviarle un mensaje a ese hijo, o el mensaje devuelve un error, con un `hive-review-ux` nuevo; su brief lleva los hallazgos corregidos y solo las regiones y estados corregidos;
  - nunca con una revisión UX completa nueva.

  · Hoy es falso porque el paso 3 solo dice «Recheck and recapture corrections».
- AC5 — El rol indica que, ante un seguimiento tras una corrección, el revisor revisa solo las regiones y estados que se le nombran, contra sus hallazgos anteriores, y responde en un solo mensaje. · Hoy es falso porque el rol no trata los seguimientos.
- AC6 — El veredicto del rol se emite solo sobre los defectos que el cambio introdujo en las regiones revisadas:
  - es «aprobado» cuando los estados revisados no muestran ningún defecto introducido, de cualquier prioridad (`flow-verify` paso 3 corrige todo defecto de UI introducido);
  - es «no aprobado» cuando no se pudo revisar ningún estado, y el reporte dice que la revisión no corrió;
  - los defectos previos se reportan en una lista aparte, cada uno con su prioridad, y no cambian el veredicto.

  Siguen vigentes «el veredicto arranca como no aprobado» y «un defecto observado conserva su severidad». · Hoy es falso porque `hive-review-ux.md:12` no dice qué hace pasar el veredicto y hace fallar el veredicto con cualquier defecto observado.
- AC7 — La página `site/src/content/docs/flows/verify.md` describe los tres comportamientos: los medios en el brief, la relectura acotada y el veredicto sobre lo introducido. · Hoy es falso porque la página no los menciona.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

- **Sesiones revisadas.** 11 hijos `hive-review-ux` en Codex del 2026-11-03 y 2026-11-04, en 6 sesiones de cam y tricell-hive:
  - 7 primeros veredictos «no aprobado» solo por defectos previos;
  - 4 casos de cobertura incompleta: CAM-8 (2 turnos), COM-920 `qa_ux_review` y CAM-9 `cam9_ux_fix_readback`;
  - 1 revisor nuevo para releer correcciones, `cam9_ux_fix_readback`, con 232 llamadas y 23,7 M de tokens. Los seguimientos acotados al mismo hijo (cam `01a1a2e6…83822e` y tricell-hive `…2b6abb`) tardaron 1 min.
- **Sesiones que fundan cada hallazgo** (host, ID, fecha):
  - H16: Codex cam `01a1a844-a92a-7c93-b7a3-214c8be7c47d` (CAM-8, 2026-11-04), Codex cam `01a1a457-db8b-7380-b1de-2401ccbd7690` (COM-920, empezó el 2026-11-03 21:30 local) y Codex cam `01a1a8a8-b19f-7562-9db8-8319d3b6ade1` (CAM-9, 2026-11-04).
  - H17: Codex cam `01a1a8a8-b19f-7562-9db8-8319d3b6ade1` (CAM-9).
  - H18: Codex tricell-hive con padres `…43b7a5` (2026-11-03), `…2b6abb` (2026-11-03) y `…e6e4da` (2026-11-04), además de cam `01a1a2e6…83822e` (2026-11-03), `01a1a5b7…3d7fa1` (CAM-7) y COM-920.
- **Admisión según «Measurement».**
  - H18 cumple la recurrencia: seis sesiones independientes, todas posteriores a `6bf808f4` (2026-11-02), el último cambio del veredicto del rol.
  - H16 no cuenta como recurrencia: el «Reviewer brief» que lo rige cambió en `6b80fc65` (2026-11-04 16:30 local), después de que empezaran sus tres sesiones.
  - H17 tiene un caso claro.
  - H16 y H17 los adopta la aprobación del usuario. El usuario eligió «Corregir H16, H17 y H18 en Hive» con `/flow-plan sobre los tres hallazgos`; la respuesta que arranque el build de este plan confirma esa adopción.
- **Reglas vigentes:**
  - `content/agents/review/hive-review-ux.md`, pasos 2–4;
  - `content/skills/flow-verify/SKILL.md`, pasos 2 y 3;
  - `content/skills/flow-verify/references/ui-verification.md:5`, que no cambia: regula la relectura del principal;
  - `content/skills/flow-build/references/delegation.md:40`;
  - `site/src/content/docs/flows/verify.md:20-24`.
- **Lugares buscados:**
  - `tests/content/agent_boundary_test.go` solo fija el bloque de límites del rol, que no cambia;
  - `tests/content/spec_deltas_test.go` valida el delta;
  - `README.md:327` solo nombra «UX review»;
  - `content/skills/flow-deploy/references/deployment-checks.md:32` lanza `hive-review-ux` en el recorrido de un despliegue y apunta a `flow-verify` paso 3 para los defectos previos. Recibe la frase del «no cubierto» (T2) y su puntero sigue siendo válido;
  - `content/skills/unattended-delegation/SKILL.md:13` no le afecta;
  - `_support/docs/harness-engineering/session-review-rubric.md:43` describe el brief de un revisor y sigue siendo válido;
  - la guía global, `_support/docs/architecture/` y `tests/pilot/` no describen el veredicto ni el brief UX.
- **Seguimiento a un hijo por host.** Observado en Codex (`send_message` y `followup_task` a un hijo que ya terminó) y en Claude Code (`SendMessage` a un subagente). No verificado en Grok, Pi, OpenCode ni Cursor: por eso la alternativa se define por capacidad.

## Entrega

- **Entrega:** automática (D4 A). Verifico, reviso, hago push, abro el PR y lo integro en `development` sin parar para validar.
- **Revisión:** nativa, Claude Code `/code-review` (D5 A), sobre el candidato antes del push.
- **No autoriza:** una release, desplegar la instalación local fuera de lo que `AGENTS.md` fija para el checkout del mantenedor, ni escribir en trackers.

## Estado de revisión y avance

2026-11-04 — Plan escrito. La revisión del plan (`hive-review-plan`) encontró 10 huecos; se aplicaron todos:
- condición de «aprobado» y caso sin estados revisados (AC6);
- reporte del «no cubierto» sin parada de validación y en el recorrido de despliegue (AC3);
- permisos fuera de AC1, porque el rol no los revisa;
- alternativa por capacidad y brief con los hallazgos corregidos (AC4);
- comprobaciones exactas antes y después;
- `CHANGELOG.md`;
- admisión de H16 por aprobación;
- comandos del sitio;
- lugares buscados.

La base se actualizó a `47362771`, que añade dos commits del script de medición sin conflicto.

2026-11-04 — `TestSpecDeltasApplied` ya falla en la base desde `bb61f71f`: dos archivos del mismo día modifican un requisito y el orden por nombre no refleja el orden real. Se añadió T4 porque bloquea la verificación. Fue una decisión menor: el test acepta cualquiera de los deltas del mismo día, sin renombrar carpetas ni depender de git.
