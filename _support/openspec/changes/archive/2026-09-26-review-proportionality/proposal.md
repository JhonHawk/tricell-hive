# Revisión de código proporcional

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `rebuild/harness-engineering` (`282145f`) |
| Tracker · GitHub Issues | Sin issue; sale de la investigación del 2026-09-26 sobre revisión proporcional |
| Git | Commit `282145f` con push a `rebuild/harness-engineering` (el usuario aprobó la redacción el 2026-09-26) |
| Verificación | Lecturas con `rg` · `go test ./tests/content/...` · `./integrations/agents/...` · tests de skills · `go vet` y `go test -race` una vez · `/code-review` |
| Siguiente paso | Ninguno en este cambio; el despliegue a los hosts se ofrece aparte |

## Objetivo

Hive trata la revisión dedicada de código como algo que se decide en cada entrega, sin importar el riesgo ni el tamaño del cambio:

- **Siempre se pregunta:** `delivery-decisions.md:18` pide elegir un mecanismo "for every mode including hold".
- **Sin regla para P2 y P3:** nada dice qué hace el orquestador con los hallazgos P2 y P3 que define `review-code.md:16`.
- **Sin tope de rondas:** la re-revisión de la implementación no tiene límite, mientras que la del plan tiene una sola ronda (`plan-review.md:34`).

El Hive legacy tenía las tres piezas y el rebuild no las trajo.

En la sesión del instalador (`c1a83bea`) la revisión encontró defectos reales: faltaba el subcomando `hive bootstrap`, `hive recover` estaba roto y `sdd-verify` detectó un panic. Pero también se corrigieron todos los hallazgos sin priorizar, incluido el pulido de UX, y hubo una cuarta pasada para verificar las correcciones.

La evidencia apunta a revisar según el riesgo y acotar los hallazgos:

- **Google:** aprueba cuando el cambio "mejora la salud del código", no cuando queda perfecto.
- **Guía de Claude Code:** un reviewer al que se le piden huecos "usually report some, even when the work is sound".
- **Ruido de los reviewers con IA:** algunos reviewers automáticos basados en LLM tienen menos de 10% de precisión.
- **Proyectos de referencia:** optional reference project no revisa los cambios pasivos, optional reference project y optional reference project limitan las rondas, y optional reference project advierte "do not run it in a loop until it comes back clean".

El cambio define cuándo se recomienda una revisión dedicada, qué hallazgos se corrigen según su prioridad, cuántas rondas se permiten y qué reporta el reviewer.

## Alcance y aceptación

**Incluye:**
- `content/skills/flow-plan/references/delivery-decisions.md:18-25`: criterio de recomendación (Q1 con D1-B y D2-A) y opción "No dedicated review".
- `content/skills/flow-build/references/verification.md:31-32` y `:39`: filas de gates que respetan la decisión de revisión (B1), tratamiento de los hallazgos por prioridad (Q2), una sola re-revisión (Q3) y un reviewer por candidato (Q6).
- `content/agents/review/review-code.md`: no reportar hallazgos cuya única corrección sea código defensivo o un test para un caso imposible, y cerrar con el conteo por prioridad (Q4).
- `content/guidance/global.md:105`: el `Review` declarado por el proyecto es el mecanismo recomendado cuando aplica una revisión dedicada; `tests/content/budget_test.go` se ajusta al tamaño.

**Excluye:**
- **Q5 (tests tras corregir):** ya lo cubre `flow-build/SKILL.md:54` desde `86473b8`, que pide volver a correr solo lo afectado y repetir la suite solo si una edición puede afectar lo que cubre.
- **Revisión de UI** (`verification.md:57`): conserva su propia regla Blocker/High y sus hijos obligatorios.
- **Revisión de planes:** ya tiene su tope de una ronda.
- **Caso de regresión en `tests/pilot`:** el cambio sale de la investigación, no de una falla observada. En `c1a83bea` la revisión fue útil; el exceso estuvo en cómo se trataron los hallazgos.
- **Pilotos con modelos:** siguen pausados.
- **Despliegue a los hosts:** se ofrece aparte.

**Criterios de aceptación:**
- A1. `delivery-decisions.md` recomienda una revisión dedicada solo cuando el cambio toca un contrato público o persistido, seguridad, permisos o datos almacenados, o concurrencia, o cuando cambia más comportamiento del que sus tests y el in vivo local pueden mostrar correcto. Si no hay disparador y la guía del proyecto no exige revisión, recomienda "No dedicated review" y dice qué le falta al cambio. Las filas `verification.md:31-32` no vuelven a proponer la revisión que ya se decidió. La opción existe siempre y el usuario puede elegir un mecanismo igual.
- A2. `verification.md` pide corregir P0 y P1 confirmados o plausibles (o verificados por el propio mecanismo), corregir P2 solo si es local y está dentro del alcance (si no, sigue la regla de hallazgos incidentales) y reportar P3 sin corregirlo; los preexistentes siguen la regla de hallazgos incidentales.
- A3. `verification.md` permite como máximo una re-revisión, limitada a las correcciones y solo cuando una corrección de P0 o P1 cambió lógica o un contrato. Lo que siga abierto, incluido un hallazgo nuevo de esa ronda, va al usuario; una revisión que aprueba cierra la revisión hasta que una edición posterior cambie lógica o un contrato.
- A4. `verification.md` pide un solo reviewer de código por candidato y agrega `review-security` solo cuando el cambio cruza un límite de `security-boundaries.md`, dentro de la autorización de revisión acordada.
- A5. `review-code.md` no reporta hallazgos cuya única corrección proteja contra un caso que el cambio no puede producir, o lo pruebe (una entrada que cruza un límite de confianza puede producir cualquier caso), y termina con el conteo de hallazgos por prioridad.
- A6. `global.md:105` subordina el `Review` declarado a que aplique una revisión dedicada, y `globalGuidanceBudget` coincide con el tamaño medido.
- A7. Pasan `go test -count=1 ./tests/content/... ./integrations/agents/...`, los tests de skills, `go vet ./...` y `go test -race -count=1 ./...`, salvo fallas del trabajo ajeno en `tooling/**`, que se reportan con su salida.

## Decisiones del usuario (2026-09-26)

- **Ruta:** `flow-plan` con Q1–Q6.
- **D1-B:** no hay umbral numérico; que un cambio sea "grande" lo juzga el agente según si sus tests y el in vivo alcanzan a mostrarlo correcto.
- **D2-A:** sin disparador, la pregunta de entrega recomienda "No dedicated review" con su razón y el usuario puede pedir revisión igual.

## Entrega

Respuesta del usuario (2026-09-26):
- **Modo:** `direct-base`. Implementar y verificar, y parar para que el usuario lea la redacción en el diff. Después, un commit con solo las rutas de este cambio y esta carpeta, con push a `rebuild/harness-engineering`. El archivo del cambio va en un segundo commit con push.
- **Revisión:** `/code-review` de Claude Code sobre el diff, antes del commit.
- **Despliegue:** fuera de este plan; se ofrece al terminar.

El árbol tiene cambios sin commitear de la sesión del instalador (`tooling/**`, `README.md`, `_support/docs/architecture/installer.md` y otros). La entrega agrega al commit solo las rutas de este cambio.
