# Delegación desatendida: vencimiento acotado, ronda de preautorización y anuncio visible (2026-10-26)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` el 2026-10-26 por el [PR #201](https://github.com/JhonHawk/tricell-hive/pull/201), merge `5cb8ff6b`; #200 cerrado |
| Tracker · GitHub Issues | • [#200 — La delegación desatendida se apaga con feedback y se activa sin preguntas de entrega](https://github.com/JhonHawk/tricell-hive/issues/200) |
| Git | `automatic` (D4-A) · rama `feat/gh-200-unattended-preauthorization` desde `development` · sin parada antes del push |
| Verificación | comprobaciones `rg` por tarea · `go vet ./...`, `go test ./...` y `npm run build` en `site/` sobre el candidato final · `/code-review` de Claude Code (D5-A) |
| Siguiente paso | Ninguno; registro cerrado |

## Solicitud

Mensaje del usuario, 2026-10-26:

> Sobre la skill de trabajo desatendido, alguna vecez lo lanzo y me quedo uno sminutos, al ver la sesión dor retro, y eso la desactiva antes de siguiera irme, haz research al respecto porque incluso podria lanzar las preguntas para preautorizar flujos, tengo sesión abierta de pi en carvente y claude en cam

Respuestas siguientes, 2026-10-26, en tarjetas nativas:

- D1: adoptar A (vencimiento acotado), B (ronda de preautorización) y C (anuncio visible).
- D2: abrir el ticket y pasar a plan → [#200](https://github.com/JhonHawk/tricell-hive/issues/200).
- D3-A: la ronda de preautorización se pregunta en texto, sin una herramienta que bloquee el turno, y el agente sigue trabajando.
- D4-A: entrega `automatic`. D5-A: `/code-review` de Claude Code.

#200 no tiene criterios de aceptación propios; su sección «Comportamiento esperado» se traduce en los AC de abajo.

## Objetivo

Hoy, cuando el usuario anuncia que se va pero se queda unos minutos mirando la sesión, la delegación desatendida falla de tres maneras:

- **Vencimiento amplio.** Cualquier mensaje que «gives a new instruction» o «reports something new» la apaga. Un aviso de un recurso o una corrección menor, dentro del mismo objetivo, la termina si se lee al pie de la letra.
- **Sin punto de entrega.** El «me voy» la activa aunque nadie haya preguntado entrega, revisión ni efectos previstos. La delegación solo llega «up to the delivery point the answers chose», así que el trabajo acaba en una pregunta que nadie contesta.
- **Activación invisible.** La activación queda solo en el razonamiento, y el usuario no sabe si la delegación sigue vigente.

Después del cambio:

- el feedback dentro del objetivo se toma como insumo y la delegación sigue;
- al activarse, el agente pregunta de inmediato, en texto y sin bloquear, lo que falta para llegar a la entrega, y mientras tanto sigue trabajando;
- la activación y su estado se ven en cada mensaje.

## Alcance y aceptación

**Incluido:**

- el primer párrafo y la sección de activación de `content/skills/unattended-delegation/SKILL.md`;
- una sección nueva en esa misma skill;
- el bullet «Unattended work» de `content/guidance/global.md`;
- las condiciones de tarjeta y de texto de `content/skills/flow-plan/references/questions.md`;
- una frase puntero en `content/skills/flow-build/references/verification.md`, `content/skills/flow-plan/references/delivery-decisions.md` y `content/skills/flow-build/SKILL.md`;
- `site/src/content/docs/index.mdx`, `site/src/content/docs/skills.md` y `site/src/content/docs/flows/build.md`;
- `CHANGELOG.md`.

**Excluido:**

- **Los efectos que autoriza la delegación** (sección «Effects»): no cambian.
- **Las fuentes de activación**, incluida la del silencio tras la ruta (decisión W1-B del 2026-10-18): no cambian.
- **Los detectores de `tests/pilot/`**: solo listan el nombre de la skill (`tests/pilot/flows.go:550`), y los pilotos siguen en pausa.
- **Deltas de requisitos**: ninguna spec de `_support/openspec/specs/` describe este comportamiento, igual que en #143/#144.

**Restricciones:**

- `content/skills/unattended-delegation/SKILL.md` queda en 16 000 bytes o menos (`tests/content/skill_size_test.go`). Hoy mide 5 885.
- El PR declara el tamaño de `content/guidance/global.md` antes y después. Hoy mide 39 046 bytes.
- La suite existente sigue en verde.
- Cada regla tiene un solo hogar: los otros archivos solo apuntan a ella.

**Decisiones menores tomadas sin preguntar:**

- Una pregunta del usuario sobre el trabajo, como «¿por qué elegiste X?», también mantiene la delegación. Es otra forma de feedback, y la regla actual ya la mantenía para las preguntas de estado.
- La lista de mensajes que mantienen la delegación es cerrada. Cualquier otro mensaje la termina, como hoy.
- La respuesta a la ronda de preautorización mantiene la delegación. Sin esto, contestar la ronda la apagaría.

**Criterios:**

- AC1. La skill y la guía global dicen que un mensaje que aporta información, un recurso, feedback o una corrección dentro del objetivo y el alcance, o que pregunta sobre el trabajo, mantiene la delegación. El agente lo aplica como insumo y dice que la delegación sigue. Un mensaje que cae en las dos listas, como una corrección que cambia el modo de entrega, la termina. *Falso en la base cuando* `content/skills/unattended-delegation/SKILL.md:8` la termina con «reports something new» o «gives a new instruction».
- AC2. La skill dice qué hacer cuando una delegación explícita o un «me voy» se activa con un objetivo que llega a build o entrega y faltan respuestas:
  - preguntar de inmediato, en una sola ronda, la entrega, la revisión, `UI verification` y las autorizaciones previstas, con los efectos críticos con su destino y tope;
  - hacerlo en texto, sin una herramienta que bloquee el turno, y seguir trabajando;
  - mantener la delegación con la respuesta, que además fija el punto de entrega;
  - sin respuesta, llegar a un candidato local verificado como lo deja `hold`, sin revisión dedicada ni commit, y listar la ronda pendiente con sus mismos ID en el reporte final.

  *Falso en la base cuando* la skill no tiene esa sección (`rg -F "pre-authorization round" content/` sin coincidencias).
- AC3. `questions.md` excluye la ronda de preautorización de la tarjeta que espera y la incluye entre los casos que se preguntan en texto, con una línea de pausa que dice que el trabajo sigue mientras tanto. `verification.md`, `delivery-decisions.md` y `flow-build/SKILL.md` remiten a la skill para el momento de esa ronda, y la sección «Decisions» de la skill ya no manda decidir sus preguntas. *Falso en la base cuando* ninguno de esos archivos menciona la ronda.
- AC4. La skill exige anunciar la activación en texto visible: la fuente, la cita, el objetivo y el vencimiento. Cada mensaje al usuario mientras corre muestra el estado de la delegación, y el mensaje con que termina dice que terminó y por qué. *Falso en la base cuando* `content/skills/unattended-delegation/SKILL.md:10` no exige texto visible ni estado por mensaje.
- AC5. `index.mdx`, `skills.md` y `flows/build.md` del sitio describen el vencimiento acotado y la ronda de preautorización, y `CHANGELOG.md` tiene la entrada. *Falso en la base cuando* `site/src/content/docs/index.mdx:97` dice «anything else hands control back to you».

## Entrega

- **Repositorio:** `tricell-hive`, base `development` (`## Hive` de `AGENTS.md`).
- **Modo `automatic` (D4-A):**
  - rama `feat/gh-200-unattended-preauthorization` en un worktree;
  - commits por tarea;
  - `/code-review` de Claude Code sobre el candidato local (D5-A), con hallazgos corregidos o refutados con evidencia;
  - push, PR a `development` y merge con los chequeos en verde. Este repo no tiene CI (memoria `project_local_tests_no_ci`; la suite local es la puerta).
- **Después del merge:**
  - `hive update` desde el checkout principal, porque existe `_support/workspace/maintainer.local`;
  - #200 se cierra al integrarse en `development`;
  - el registro se archiva con `flow-close`.
- **Sin despliegue:** `development` no publica nada; el sitio se publica desde `master`, que necesita promoción aparte.
- **Autorizaciones previstas:** ninguna más allá de la entrega. Todo es local hasta el push.
