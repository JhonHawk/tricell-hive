# Diseño — gh-249-mid-build-questions

## Contexto verificado

Leído en `development` `3bb1e623` el 2026-10-28.

- `content/guidance/global.md:29` ya ordena arreglar sin preguntar un defecto que bloquea el trabajo, como un check que falla. Las excepciones son un arreglo que corre en producción, un efecto no cubierto o crítico, y una decisión que el usuario podría tomar distinto.
- `content/skills/flow-plan/references/delivery-decisions.md:38` ya cubre, sin pregunta nueva, el PR de seguimiento que arregla el mismo ticket después del merge.
- `content/skills/flow-build/references/verification.md:56` es la regla que se lee al esperar el CI. Contradice a las dos anteriores para un fallo causado por el cambio.
- `content/skills/flow-build/references/foreseen-authorizations.md:9` pide en la ronda de entrega solo lo que el plan o la petición nombran. Lo demás se pregunta «when it arises».
- `content/skills/flow-plan/SKILL.md:40` exige el entorno y los efectos solo para los gates sustanciales. Su lista de readiness (`:44-54`) y las comprobaciones bloqueantes de `plan-review.md:22` no los piden.
- `content/skills/hive-core/references/questions.md:9-11` ordena preguntar en el punto de decisión. Ninguna regla distingue el build, en el que el usuario no está atento.
- `content/skills/flow-build/SKILL.md:20` manda preguntar la decisión que falta «before dependent edits».
- `content/skills/unattended-delegation/SKILL.md` ya decide por sí misma las preguntas mientras el usuario no está, y queda igual.

## Enfoque

### 1. Fallo de CI causado por el cambio (AC1)

En `verification.md:56` amplío el disparador para que cubra un run que falla en el PR antes del merge o en la base después de él (N1), y separo los dos casos que hoy comparten la última oración:

- Un fallo de infraestructura que vuelve a fallar tras el reintento sigue quedando pendiente, como ahora.
- Un fallo que viene del código del cambio es un defecto que bloquea el trabajo autorizado y se arregla bajo la regla global. Antes del merge, con un commit nuevo en el mismo PR. Después, con un PR de seguimiento bajo la misma entrega, como fija `delivery-decisions.md`.

El ticket sigue fuera de su estado de completado hasta que el run del arreglo pase. Solo se reporta como pendiente cuando la regla global no permite arreglarlo sin preguntar.

Es un cambio neto de unos 150 B, y el archivo tiene 620 B libres. Si no cabe, se acorta la redacción de la misma viñeta. No se mueve otra regla.

### 2. Entornos y efectos previstos en el plan (AC2, AC3)

- **Readiness en `flow-plan/SKILL.md`:** agrego un punto a la lista. Cada criterio que se verifica fuera de esta máquina nombra su entorno y los efectos, datos y accesos que su verificación necesita allí, clasificados como fija `effect-authorization.md`, para que la ronda de entrega los pida como fija `foreseen-authorizations.md`. Son unos 250 B, y el archivo tiene 1 013 B libres.
- **`plan-review.md:22`:** agrego el mismo caso a las comprobaciones bloqueantes que corre toda asignación, sea cual sea su dominio. Así lo revisa también un plan sin dominio de infraestructura.
- **`foreseen-authorizations.md:9`:** cambio «for work without a plan the request, already names» por lo que la petición y la verificación que necesita requieren, que el agente lista antes de preguntar (D4-A). Se mantiene «an effect neither covers is asked when it arises» para lo que de verdad no se podía prever.
- **`delivery-decisions.md:7`:** «the authorizations the plan or request foresees» pasa a nombrar el plan, o la petición y su verificación (B6).

Lo que queda en una sola casa:
- la clasificación de efectos, en `effect-authorization.md`;
- el momento de preguntar lo previsto, en `foreseen-authorizations.md`;
- la exigencia al plan, en la lista de readiness;
- la comprobación del revisor, en `plan-review.md`.

### 3. Decisiones a mitad de build (AC4, D7-B)

La casa es una condición nueva en «Decide whether a question is needed» de `questions.md`. Es la referencia que todo flujo lee antes de formular una decisión.

- **Condición:** una decisión que pertenece al usuario aparece durante un build, o durante la verificación del candidato, fuera de una parada de validación. No autoriza ningún efecto, cosa que siguen gobernando `effect-authorization.md` y `foreseen-authorizations.md`.
- **Exclusiones** (B1, B4):
  - una tarea añadida o un cambio nuevo, que se clasifican en `scope-changes.md` y siguen esa regla;
  - una decisión que otra regla reserva a la respuesta explícita del usuario, que sigue esa regla. Por ejemplo: la desviación de tema de `ui-gate.md`, las rondas agotadas de `plan-execution.md`, saltar u omitir un check (`flow-verify/SKILL.md`), un hook (`git-workflow/SKILL.md`) o las alertas de instalación de `security-boundaries.md`.
  - la delegación desatendida, que tiene su propio régimen.
- **Reversible** (N2): deshacer la decisión solo cambia trabajo de esta tarea en esta máquina, aún no publicado. Eso incluye una decisión de contrato interna del cambio (D8-A), y excluye los datos fuera de esta máquina y los contratos ya publicados. El término no se confunde con la clase Production de `effect-authorization.md`, porque la regla nombra «on this machine».
- **Acciones:**
  - **Reversible:** aplica la opción recomendada como supuesto, lo registra en el progreso del plan o en la conversación, y sigue.
  - **No reversible pero que no bloquea el resto:** sigue con el trabajo independiente y deja la decisión pendiente.
  - **Bloquea todo lo restante y no es reversible:** pregunta en el momento.
  - **Decisiones pendientes:** se preguntan juntas cuando ya no queda trabajo independiente.
  - **Confirmación de los supuestos (D9-A):**
    - bajo `interactive`, en la parada de validación;
    - bajo `automatic` y `direct-base`, después del merge, sin agregar paradas. El reporte de cierre los lista entre las decisiones tomadas sin preguntar (`reporting.md:70`), y la confirmación se pide en la misma ronda que la pregunta de cierre. Una corrección se clasifica allí como corrección (`scope-changes.md:40`) y va como PR de seguimiento bajo la misma entrega (`delivery-decisions.md:38`).
    - bajo `hold`, al entregar el resultado, también junto con la pregunta de cierre.
    - Un supuesto no es un ítem pendiente ni una elección abierta que impida decir que el trabajo está completo (R1). `reporting.md:84` gana un puntero breve que lo dice; el archivo tiene 207 B libres.
    - Un rechazo que llega después de la ronda de cierre no reabre nada: se clasifica como trabajo nuevo bajo `scope-changes.md` (N4).
- **Exclusión cerrada (N5):** en `questions.md`, las reglas que reservan la decisión al usuario se nombran una por una (las de la lista de exclusiones de arriba), sin un «such as».
- **Punteros:**
  - `questions.md:11` («ask only at that decision point») gana «except during a build, as the condition below sets» (B5).
  - En `flow-build/SKILL.md:20`, «or ask for the missing decision before dependent edits» pasa a resolver la decisión como fija `questions.md`.
  - `global.md:29` (D8-A): la excepción «a choice the user could make differently, such as scope, a contract or stored data» suma que, durante un build, una decisión reversible sigue la regla de `questions.md`. Es un puntero breve, y el PR reporta el tamaño de `global.md`.
- **Sin cambios:** `scope-changes.md` no cambia, porque la exclusión vive en la regla nueva. `delivery-decisions.md:24` tampoco: `automatic` sigue sin paradas, por D9-A.
- **Tamaño (N3):** si `questions.md` pasa de 16 000 B, acorto la redacción de la propia condición. No muevo otras reglas.

Riesgo: un supuesto rechazado obliga a rehacer trabajo. Queda en observación (`watch/build-assumptions-d7b`) con su condición para pasar a D7-A.

### 4. Sitio y changelog (AC5)

- **`site/src/content/docs/flows/build.md`:**
  - la viñeta de defectos (`:39`) menciona el fallo de CI;
  - la ronda antes de la primera edición (`:50`) menciona que el build sin plan lista lo que su verificación necesita;
  - la FAQ «What if the code no longer matches the plan?» (`:84`) y una viñeta de lo que hace describen el supuesto y dónde se confirma según el modo.
  - la FAQ de modos (`:78`) dice que bajo `automatic` los supuestos se confirman después del merge.
- **`site/src/content/docs/flows/plan.md`:** la FAQ de preguntas (`:70`) menciona que el plan nombra el entorno y lo que cada verificación necesita fuera de la máquina.
- **`CHANGELOG.md`:** una entrada en `[Unreleased] / Changed`.

Las páginas describen el comportamiento y enlazan a las referencias; no copian reglas.

## Alternativas descartadas

- **D7-A** (seguir sin asumir y juntar al final): el usuario eligió D7-B, y D7-A queda en observación.
- **Una regla para el bypass `--admin` en la pregunta de promoción (D3-A):** el usuario la rechazó por depender de cada usuario.
- **Diferir los contratos (D8-B):** el usuario eligió asumirlos mientras no estén publicados.
- **Una parada antes del push bajo `automatic` (D9-B):** el usuario eligió confirmar después del merge.
