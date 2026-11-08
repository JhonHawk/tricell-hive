# Diseño

## Enfoques

Para medir quién califica se presentaron tres enfoques (capa sobre `tests/pilot`, solo procedimiento escrito, adoptar `eval.mjs`); se eligió el primero (D1 A) y su construcción pasó a `model-qualification-tool` (D13 A). Para la regla y el documento del piso hay una sola forma sensata: editar la guidance del repositorio y crear el documento, porque la regla vive solo en `AGENTS.md` y no se distribuye.

## Decisiones

- D1 a D17, con su texto en `proposal.md`.
- Decisiones menores tomadas sin preguntar, para cumplir D3, D7, D12, D14–D17:
  - **Dos conjuntos.** El **objetivo de escritura** son los modelos con nombre de los perfiles de roles (`execution`, `focused`, `reasoning`, `verifier`): ahí están los frontier (`opus`, `gpt-6.1-sol`, `grok-4.6`) y los modelos que corren roles delegados (`sonnet`, `gpt-6.1-terra`). El **piso** es ese objetivo más la clase Haiku/flash de D7: sus fallos observados cuentan, pero no son objetivo de escritura (D12).
  - **Pertenencia por modelo.** Un modelo está en el piso por su identidad, sin importar el host: se quitan el prefijo de proveedor y la variante (`opencode-go/claude-haiku-5-5#high` → `claude-haiku-5-5`), y cada alias (`opus`, `sonnet`) se lista con los IDs completos a los que resuelve hoy y la fecha en que se verificó; pertenecer es ser uno de esos IDs. Así una sesión de Pi con `opus` cuenta, y los hosts que heredan el modelo no quedan sin evidencia.
  - Una sesión cuyo modelo servido no se ve (las trazas de Codex y OpenCode no lo registran todavía) cuenta con el modelo configurado, nombrado como tal.
  - Un directorio o lote de corridas cuenta como una observación como máximo, y nunca el brazo sin guidance.
  - La lista vive en `_support/docs/harness-engineering/model-floor.md`, no en `integrations/agent-profiles.json`: ese archivo lo lee el manager y un manager anterior rechaza claves nuevas. Un test de contenido la alinea con `agent-profiles.json` y con `flashModels`.
  - `model-floor.md` es el hogar canónico de la lista, el estado, la normalización y el criterio de calificación. La regla de recurrencia sigue en `AGENTS.md` «Measurement»; `AGENTS.md` «Guidance structure», `pilot-procedure.md`, el índice de `harness-engineering`, la rúbrica, README, `llms.txt` y el sitio solo apuntan a uno u otro.
  - El enlace del sitio usa la forma `blob/master` del resto del sitio; `master` ya tiene `_support/docs/harness-engineering`, y el archivo nuevo resuelve tras la promoción.

## Forma propuesta

### `AGENTS.md`

- «Guidance structure», en lugar de la línea 22 (AC27, AC34): el contenido distribuido se escribe claro para el objetivo de escritura de `model-floor.md`; una regla dirigida a un modelo fuera de ese objetivo necesita la recurrencia de «Measurement»; la guía de prompting de un modelo frontier o su conducta por defecto no basta para quitar una regla, y la evidencia para quitar está en «Measurement». Sin umbral ni conjunto de modelos propio: un solo hogar.
- «Measurement»:
  - Línea 86 (AC35, AC36, AC37): «needs that failure observed in two independent real sessions, or the user's approval to adopt it at once» pasa a dos observaciones independientes en candidatos del piso, al menos una en una sesión real y la otra, si no, en una corrida autorizada de calificación o de piloto que corrió la versión vigente de la regla, nunca el brazo sin guidance, con cada directorio o lote como una observación como máximo (D14, D15), y conserva la aprobación del usuario como atajo; un fallo solo fuera del piso no cuenta; pertenencia por modelo según `model-floor.md`; modelo configurado cuando el servido no se ve; se conserva «Count only sessions that ran the current version», extendido a corridas; relajar o quitar una regla sigue siendo decisión del usuario, con evidencia de sesiones reales de los candidatos del piso, y las corridas autorizadas suman pero no bastan solas (D16).
  - Línea 88 (AC31): el watch item registra host, modelo, sesión o directorio de corrida, y fecha; las corridas autorizadas que ejercitan la regla cuentan entre los tres episodios que cierran un watch item.
  - Línea 90 (AC31): el commit de una corrección nombra la sesión o la corrida con host, modelo y fecha.

### `model-floor.md` (AC38)

| Host | Modelo | Esfuerzos | Origen | Objetivo de escritura | Estado | Verificado | Referencia |
| --- | --- | --- | --- | --- | --- | --- | --- |
| claude | `opus` | `high` | `reasoning`, `verifier` | sí | provisional | ninguna | — |
| claude | `sonnet` | `high` | `execution`, `focused` | sí | provisional | ninguna | — |
| codex | `gpt-6.1-sol` | `medium`, `high` | `execution`, `reasoning`, `verifier` | sí | provisional | ninguna | — |
| codex | `gpt-6.1-terra` | `medium` | `focused` | sí | provisional | ninguna | — |
| grok | `grok-4.6` | — | `verifier` | sí | provisional | ninguna | — |
| opencode | `opencode-go/claude-haiku-5-5` | `high` | piso del 2026-10-11 (`flashModels`) | no | provisional | ninguna | — |
| opencode | `opencode-go/deepseek-v4.1-flash` | `high` | piso del 2026-10-11 (`flashModels`) | no | provisional | ninguna | — |
| opencode | `opencode-go/qwen3.8-flash` | `medium` | piso del 2026-10-11 (`flashModels`) | no | provisional | ninguna | — |
| opencode | `opencode-go/glm-5.3-flash` | `high` | piso del 2026-10-11 (`flashModels`) | no | provisional | ninguna | — |

Más: los IDs completos de cada alias con su fecha de verificación, la normalización de modelos, qué cuenta para la recurrencia (puntero a `AGENTS.md` «Measurement»), el criterio de calificación (D2 con D11, D9), que la herramienta está pendiente en `model-qualification-tool` y que correrla es un piloto con autorización y tope de gasto, y los límites conocidos (Claude, Grok y Pi sin brazo sin guidance; Codex y OpenCode sin modelo servido en la traza).

### Test de contenido (`tests/content/model_floor_test.go`, AC38)

Lee la tabla de `model-floor.md` y comprueba: estado `provisional` o `qualified` en cada fila; que cada modelo con nombre (no `inherit` ni vacío) de `execution`, `focused`, `reasoning` y `verifier` en `agent-profiles.json` tenga una fila de su host, marcada como objetivo de escritura, que incluya cada esfuerzo con nombre; que cada entrada de `flashModels` en `tests/pilot/flash_class.go` (leída con una expresión regular sobre el archivo, porque es una variable no exportada de otro paquete) tenga su fila con el esfuerzo de la variante; que cada alias tenga al menos un ID completo y una fecha. Casos con copias modificadas: una fila quitada, un esfuerzo cambiado y un alias sin ID fallan.

### Rúbrica de revisión de sesiones (AC33)

- Una fila: antes de contar una sesión para la recurrencia, comprobar que su modelo, servido o, si no se ve, el configurado, esté en el piso de `model-floor.md`, y registrarlo.
- La línea 58 («a correction to distributed guidance needs two independent sessions or the user's approval») pasa a apuntar a `AGENTS.md` «Measurement» sin repetir la regla.

## Impacto

- `rg "least capable"`: solo `AGENTS.md:22`. `AGENTS.md:21` («literal and smaller models act on what they can check») queda igual: es una razón para criterios observables, no un objetivo de escritura.
- `AGENTS.md:86`, `:88` y `:90` cambian juntos: son la regla de recurrencia y sus registros. `session-review-rubric.md:58` es la única copia de esa regla fuera de `AGENTS.md`.
- `README.md:29` y `site/src/content/docs/index.mdx:37-38` se alinean.
- Ningún cambio en `content/`, `tooling/`, `integrations/` ni `tests/pilot`: no hay despliegue.

## Recuperación

Revertir el commit. No hay datos ni contratos persistidos.
