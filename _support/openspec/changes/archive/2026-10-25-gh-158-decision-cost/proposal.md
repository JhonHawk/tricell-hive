# Costo visible en las decisiones

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por PR #164 (`66837399`) el 2026-10-25 |
| Tracker · GitHub Issues | • [#158 — Costo invisible en las decisiones: opciones sin tamaño, siguiente ticket sin estimar y planes que contradicen reglas globales](https://github.com/JhonHawk/tricell-hive/issues/158) |
| Git | `automatic` · PR a `development` con `/code-review` · sin parada humana |
| Verificación | `go test ./tests/content/...` · conteo de frases · `go vet ./...` y `go test ./...` sobre el candidato final |
| Siguiente paso | Ninguno; la promoción a `master` va aparte |

## Solicitud

> Y tambien, sobre esos proyectos [...] ademas de por ejemplo, en este mismo proyecto hay otra sesión de claude trabajando y para cosas que consideraba sencillas o claras, lleva cerca de dos horas trabajando, eso tambien es un cuello de botella

> otro detalle a agregar y es muy notorio en la sesión actual de claude en cam, un circulo vicioso en el tema de hallazgos -> resuelven -> mas hallazgos, y casi todos son detalles menores

Respuesta a la ruta D3: «Ticket + flow-plan (Recomendada)» (2026-10-25).

Criterios del ticket [#158](https://github.com/JhonHawk/tricell-hive/issues/158), leídos el 2026-10-25:

> - **AC1:** una pregunta cuyas opciones difieren en un orden de magnitud de tareas o tiempo muestra la estimación en cada opción.
> - **AC2:** la opción de cierre «seguir con el siguiente ticket» dice el tamaño estimado de ese ticket.
> - **AC3:** la revisión del plan, para planes que editan guía distribuida, reporta como hallazgo una opción o tarea que contradice un invariante de `global.md`.
> - **AC4:** `AGENTS.md` declara `Delivery` y `Review` en `## Hive`.

## Objetivo

En la sesión `14a79ad0` de este repositorio (143 min), más de la mitad del tiempo se fue en trabajo cuyo costo nadie mostró antes de elegirlo: encadenar un ticket que resultó ser una función en Go y una opción de diseño que agregó un editor de TOML. Además, una opción del plan contradecía una regla global y la misma decisión se preguntó dos veces. El cambio hace visible el tamaño de cada opción cuando las alternativas difieren mucho, hace que la revisión del plan detecte contradicciones con la guía vigente y declara la entrega y la revisión de este repositorio para no preguntarlas en cada sesión.

## Alcance y aceptación

Incluye `content/skills/flow-plan/references/questions.md` (hogar del formato de las opciones), `content/skills/flow-close/SKILL.md` (pregunta de cierre), `content/skills/flow-plan/references/plan-review.md`, `AGENTS.md` (`## Hive` y la línea 57) y las páginas `site/src/content/docs/flows/plan.md` y `close.md`.

El ciclo de hallazgos menores de la sesión de cam va en [#159](https://github.com/JhonHawk/tricell-hive/issues/159) (E5-A), con su propio cambio.

Fuera de alcance: dejar que una tarea dependiente arranque antes de verificar su predecesora (`plan-execution.md:26`), corrección con un solo caso, en vigilancia; reducir la verificación, la revisión del plan o `/code-review`.

Restricciones: `go test ./tests/content/...` sigue pasando; `content/guidance/global.md` no cambia de tamaño; cada regla vive en un solo hogar.

- AC1. Las Actions de la condición de costo de `questions.md:33` piden que cada opción diga su tamaño estimado (número de tareas y un rango de tiempo) con su base (criterios del ticket, archivos leídos), o «sin estimar: no investigado», nunca un número sin base, cuando se cumple un disparador observable: la opción abre o encadena otro ticket, agrega un componente, lenguaje o dependencia, o su tamaño es al menos diez veces el de otra opción. *Falso en la base cuando* `rg -i -F -c "not investigated" content/skills/flow-plan/references/questions.md` da 0.
- AC2. La pregunta de cierre de `flow-close/SKILL.md` da una vez el tamaño estimado del ticket recomendado, según AC1, para las opciones de seguir ahora y de nueva sesión; al continuar un plan en curso, el tamaño son sus tareas restantes. *Falso en la base cuando* la sección de la pregunta de cierre no menciona tamaño.
- AC3. `plan-review.md` asigna a un plan que edita guía de agentes (archivos de instrucciones como `AGENTS.md` o `CLAUDE.md` de cada nivel, skills con sus referencias, definiciones de agentes y la documentación que los describe) un revisor de «consistencia de la guía». Ese revisor busca en ese corpus todas las frases que enuncian cada regla que el plan cambia y reporta como bloqueante, con archivo y línea, la que seguiría contradiciéndola después de las ediciones del plan, y también la opción o tarea del plan que contradice una regla vigente que el plan no cambia. Una decisión registrada del usuario nunca se reporta como violación, pero sí exige editar el texto hermano. *Falso en la base cuando* la tabla de dominios de `plan-review.md` no tiene esa asignación.
- AC4. La sección `## Hive` de `AGENTS.md` declara `Delivery: automatic` y `Review: /code-review`, y `AGENTS.md:57` («without the user's explicit instruction») dice que un valor declarado de `Delivery` o `Review` es esa instrucción, remitiendo a `delivery-decisions.md`. *Falso en la base cuando* `rg -c "^- (Delivery: automatic|Review: /code-review)$" AGENTS.md` da 0.
- AC5. `site/src/content/docs/flows/plan.md` y `close.md` describen AC1–AC3 enlazando las skills. *Falso en la base cuando* ninguna menciona el tamaño estimado ni la revisión de consistencia.

## Decisiones

- **E2-A (2026-10-25), tras investigación:** la comprobación solo rige en planes que editan guía, con un revisor asignado de consistencia sobre el corpus resultante (AC3). Una comprobación general habría chocado con decisiones deliberadas del usuario: en un proyecto de cliente, un plan que borraba migraciones aplicadas por decisión registrada. El error que la motiva, una frase hermana que sigue contradiciendo una regla cambiada, se repitió en al menos siete sesiones de este repositorio y lo encontró tarde `/code-review`.
- **E3-A (2026-10-25):** `Delivery: automatic`.
- **E4-A (2026-10-25):** `Review: /code-review`. En otros hosts ese comando no existe y `delivery-decisions.md:26` vuelve a preguntar la revisión; no es una contradicción.
- **Clasificación según AGENTS.md, «Measurement»:** AC1 endurece la condición de costo de `questions.md:33`, así que es una corrección con un solo caso observado (sesión Claude Code `14a79ad0`, 2026-10-25: un ticket encadenado y una opción con editor de TOML sin tamaño). Aprobada por el usuario sin segunda sesión: E9-A (2026-10-25). AC2 y AC3 son huecos que adoptan D3-A y E2-A. Los commits que cierren AC1 nombran esa sesión, host y fecha.

## Entrega

`automatic`: rama `feat/gh-158-decision-cost` desde `development` en un worktree, commits, push, PR a `development`, `/code-review` sobre el candidato y merge cuando pasan las comprobaciones. El registro de cambio se cierra con un PR pequeño sin revisión que el agente integra. No dispara despliegues. Tras el merge, `hive update` refresca la instalación local. Empieza después del merge de gh-159, que libera espacio en `flow-close/SKILL.md` y edita la misma pregunta de cierre. El PR lista los lugares revisados y la variación de tamaño de `global.md` (0), como pide `AGENTS.md:24`.
