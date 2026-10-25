# Guía de pruebas: viejo contra nuevo, propiedades con oráculo y filtro de pruebas de poco valor

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por PR #161 (`4001d2e0`) el 2026-10-25 |
| Tracker · GitHub Issues | • [#157 — Guía de pruebas: comparar viejo contra nuevo al reemplazar, propiedades con oráculo y filtro de pruebas de poco valor](https://github.com/JhonHawk/tricell-hive/issues/157) |
| Git | `automatic` · PR a `development` con `/code-review` · sin parada humana |
| Verificación | `go test ./tests/content/...` · conteo de frases · `go vet ./...` y `go test ./...` sobre el candidato final |
| Siguiente paso | Ninguno; la promoción a `master` va aparte |

## Solicitud

> qué opinas de esto con respecto a Hive?

> Y tambien, sobre esos proyectos pruebas con son innecesarias por no detectar nada [...]

> Yyyyy tambien se une esto: [...] https://github.com/openclaw/openclaw/blob/main/.agents/skills/test-audit/SKILL.md

Respuesta a la ruta D2: «Ticket + flow-plan (Recomendada)» (2026-10-25).

Criterios del ticket [#157](https://github.com/JhonHawk/tricell-hive/issues/157), leídos el 2026-10-25:

> - **AC1:** la guía distribuida pide una comparación viejo contra nuevo con las mismas entradas cuando un cambio reemplaza o migra un escritor, parser, consulta o sistema, y dice qué ruido normalizar y qué hacer con cada diferencia.
> - **AC2:** la guía nombra cuándo usar una prueba por propiedades (lógica pura con oráculo independiente) y cuándo no (espacios finitos ya cubiertos por tablas), condicionada a que la librería exista o se proponga como dependencia.
> - **AC3:** antes de escribir una prueba, `hive-write-tests` y `tdd.md` piden comprobar que ninguna otra prueba ya es dueña del comportamiento y rechazan los patrones de poco valor listados.
> - **AC4:** la guía prohíbe agregar al código de producción un punto de entrada que solo usan las pruebas, y pide llevar la prueba al límite real.
> - **AC5:** las páginas de `site/` que describen la verificación reflejan el cambio.

## Objetivo

Hive solo pide pruebas por ejemplo. En dos proyectos reales faltó comparar el sistema viejo con el nuevo durante una migración, tres defectos de lógica pura los encontró la revisión en vez de una propiedad, y abundan pruebas que no pueden fallar por un cambio realista (mocks de módulos propios, la función bajo prueba como su propio oráculo, lectura del texto del código, puntos de entrada solo para pruebas). El cambio agrega a la guía cuándo usar una prueba diferencial o una por propiedades y un filtro que se aplica antes de escribir cada prueba, sin crear una skill nueva.

## Alcance y aceptación

Incluye `content/skills/flow-build/references/tdd.md` (hogar canónico de qué prueba escribir), `content/agents/quality/hive-write-tests.md`, `content/agents/review/hive-review-code.md`, `content/skills/flow-build/references/verification.md`, `content/skills/flow-plan/SKILL.md` (párrafo de reemplazo o migración), `content/skills/flow-build/SKILL.md:42` (descripción del puntero a `tdd.md`) y `site/src/content/docs/flows/build.md` y `plan.md`.

Fuera de alcance: una skill de auditoría y poda (la ruta iterativa de `flow-research` ya la cubre y la poda pertenece a cada proyecto); instalar librerías; cambiar `hive-verify-task` (ya pide que cada prueba citada pueda fallar).

Restricciones: `go test ./tests/content/...` sigue pasando; `content/guidance/global.md` no cambia de tamaño; ninguna regla se repite fuera de su hogar.

- AC1. `verification.md` pide, cuando un cambio reemplaza o migra un escritor, parser, consulta o sistema, comparar el viejo con el nuevo sobre las mismas entradas, normalizando el ruido (timestamps, IDs, orden), y tratar cada diferencia como defecto o cambio documentado. *Falso en la base cuando* `rg -i -c "same inputs" content/skills/flow-build/references/verification.md` no encuentra nada.
- AC2. `tdd.md` dice cuándo usar una prueba por propiedades (lógica pura con oráculo independiente: round-trip, idempotencia, invariantes, versión optimizada contra simple) y cuándo no (espacio finito ya cubierto por una tabla), y qué hacer cuando el proyecto no tiene la librería (según E1). *Falso en la base cuando* `rg -i -c "propert" content/skills/flow-build/references/tdd.md` da 0.
- AC3. `tdd.md` pide, antes de escribir una prueba, buscar la prueba que ya es dueña de ese comportamiento y no duplicarla, y rechaza la lista de patrones de poco valor: la función bajo prueba usada para calcular el valor esperado, un doble que entrega lo que se afirma, leer el texto del código fuente, aserciones siempre verdaderas, copiar del código una lista o inventario. *Falso en la base cuando* esas frases no aparecen en `tdd.md`.
- AC4. `tdd.md` prohíbe agregar al código de producción un punto de entrada solo para pruebas (un export, bandera o envoltorio que expone estado interno a las pruebas) y pide probar a través del límite real. Permite el punto de inyección en el límite del sistema (reloj, aleatoriedad, servicio externo) que producción también usa con la implementación real. *Falso en la base cuando* `rg -i -c "test-only entry point" content` da 0.
- AC5. `hive-write-tests` y `hive-review-code` remiten a `tdd.md` para el filtro y la lista de patrones en vez de enunciarlos: la frase «name the realistic change» de `hive-write-tests.md:13` pasa al filtro de `tdd.md`, y la lista de `hive-review-code.md:12` apunta a la de `tdd.md`. *Falso en la base cuando* `rg -c "tdd.md" content/agents/quality/hive-write-tests.md content/agents/review/hive-review-code.md` da 0 en ambos.
- AC6. El párrafo de reemplazo o migración de `flow-plan/SKILL.md` remite a la comparación de AC1 para que el plan la incluya en la verificación. *Falso en la base cuando* ese párrafo no menciona la comparación.
- AC7. `site/src/content/docs/flows/build.md` describe la comparación viejo contra nuevo, las pruebas por propiedades y el filtro, y `site/src/content/docs/flows/plan.md` dice que el plan de un reemplazo incluye la comparación; ambas enlazan la skill sin copiar sus reglas. *Falso en la base cuando* ninguna de las dos menciona estas prácticas.

## Decisiones

- **E1-A (2026-10-25):** sin librería de propiedades, la guía pide proponerla como dependencia de desarrollo en el plan, pasando por la revisión de dependencias; nunca instalarla a mitad de un build.
- **E3-A y E4-A (2026-10-25):** entrega `automatic` y revisión `/code-review`, que [gh-158](../gh-158-decision-cost/proposal.md) declara en `AGENTS.md`.

## Entrega

`automatic`: rama `feat/gh-157-test-guidance` desde `development` en un worktree, commits, push, PR a `development`, `/code-review` sobre el candidato y merge cuando pasan las comprobaciones. El registro de cambio se cierra con un PR pequeño sin revisión que el agente integra. No dispara despliegues. Tras el merge, el mantenedor refresca la instalación local con `hive update` (`AGENTS.md`, «Refreshing a local installation»).
