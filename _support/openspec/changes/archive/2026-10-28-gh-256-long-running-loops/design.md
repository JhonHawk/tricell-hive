# Diseño — gh-256-long-running-loops

## 1. Contrato de loop (`hive-core/references/loops.md`, D15-A)

**Cuándo se lee:** antes de empezar o retomar un trabajo que repite ciclos hasta cumplir una condición: la ruta iterativa de `flow-research`, un trabajo bajo delegación desatendida que repite ciclos (por ejemplo, «dejar el CI en verde» o «migrar los N archivos») y una primitiva de loop del host que el usuario pide.

**Qué contiene**, en este orden y cada parte en una sola casa:

1. **Objetivo y medida.** Un estado final observable y la comprobación que lo muestra, por ejemplo un comando con su resultado esperado. Cuando el host ofrece un evaluador de objetivo, la comprobación se escribe para que su salida quede a la vista en el transcript.
2. **Tope.** Ciclos, tiempo o tokens, declarados al empezar. Si el host acepta una cláusula de tope en el objetivo, se incluye. Al alcanzar el tope, el loop para y se reporta.
3. **Verificación de cada ciclo.** Se corre la medida y se registra el resultado. Cuando el resultado decide algo consecuente, se prefiere una verificación con contexto nuevo.
4. **Sin progreso (D18-A).** Si dos ciclos seguidos no mejoran la medida, se cambia de enfoque una vez. Si el ciclo siguiente tampoco mejora, se para.
5. **Estado en disco**, actualizado en cada ciclo y leído antes que nada al retomar.
   - Con plan, va en la sección de progreso de `tasks.md`. Sin plan, va en el destino de «saved state» de `artifact-layout.md`. La ubicación tiene una sola casa, `artifact-layout.md`, y `loops.md` apunta a ella (N1).
   - Registra el objetivo y su medida, el tope, la corrida que lo escribió, los ciclos hechos con la medida de cada uno, las decisiones tomadas, cada efecto con su reversión, el siguiente paso, el ticket y un estado: `running`, `done` (objetivo cumplido) o `waiting for the user: <qué>` (cualquier otra parada, el tope incluido).
   - Un estado `waiting` es terminal: no se corren ciclos hasta que el usuario lo limpie, editando el estado o, si está presente, con su mensaje de retomar. Una corrida reconoce su propio `running` y para ante el `running` reciente de otra sin tocar el estado.
   - Un estado `running` que encuentra otra corrida sin terminar para (N4 de autorización). El estado registra cuándo empezó su último ciclo, y un `running` más viejo que el tope de tiempo del trabajo, o que un día cuando el tope no es de tiempo, cuenta como interrumpido: se registra `waiting for the user: interrupted run` y se reporta.
   - No autoriza nada (B2 de autorización). No es un registro de investigación retenido (N3 de consistencia). No usa Engram por ciclo (`global.md:83`).
   - Los commits sirven como puntos de control solo cuando la respuesta de entrega los autoriza (B3 de consistencia).
6. **Escalada.** Al parar, se registra el estado final. Si el objetivo no se cumplió, se reporta qué se intentó, qué falló, la medida actual y qué se necesita, y después corre `flow-close`, como pausa.
7. **Primitivas del host.** Se nombra la capacidad, no el comando: un evaluador de objetivo con tope de turnos, un programador por intervalo o agenda, un hook de parada que fuerza a continuar y una espera con señal de finalización.
   - El respaldo, cuando el host no ofrece la primitiva, es correr los ciclos en la sesión con las esperas de `delegation.md` y el estado en disco.
   - Los nombres por host van en `_support/docs/architecture/agent-delivery.md`.

## 2. Ruta iterativa y diagnóstico (AC4)

- **`flow-research/SKILL.md:48-58`:**
  - apunta a `loops.md` para el tope, la verificación, el estado y la regla de sin progreso;
  - el estado queda en disco;
  - «a round ends without a measurable improvement» se reemplaza por la regla de sin progreso;
  - la parada «a decision is genuinely the user's» suma la excepción de D17-A: bajo delegación desatendida, se decide como fija esa skill;
  - se mantiene la parada por un efecto no cubierto, como un commit.
- **`diagnosis.md:10`:** se reescribe «fixed counts of fixes or deployments do not establish a useful stopping condition» (B4 de consistencia). La idea pasa a ser que un conteo de intentos no dice si hay progreso, y que un diagnóstico que corre en ciclos sigue la regla de sin progreso de `loops.md`.

## 3. Sin jobs declarados (D22-A, que reemplaza D16-A, D19-A y D21)

- La guía deja de nombrar el «declared job»:
  - `global.md:32` y `unattended-delegation/SKILL.md:3` y `:8` listan tres fuentes de activación;
  - un turno programado o en segundo plano sigue sin activar nada: puede investigar, reportar o dejar cambios en `hold`.
- Se quitan la sección «Declared jobs», la cláusula de `effect-authorization.md:29`, el destino `jobs/` de `artifact-layout.md` y las menciones a jobs en `loops.md`, `delivery-decisions.md`, `review-decisions.md`, `flow-build/SKILL.md:24`, el sitio y el CHANGELOG.
- `global.md:31` y `:58` vuelven a su texto de la base.
- El diseño de jobs ya investigado queda en `watch/declared-jobs`: el archivo, la procedencia, los topes y la rama de estado `hive-job/<job>`.

## 4. Retomar entre sesiones (AC2)

- El loop se retoma desde su estado en disco (sección 1).
- **`delivery-decisions.md:42`**, el párrafo de registro, suma una oración: el estado del loop nunca aporta las respuestas de entrega, revisión y `UI verification`, y sin ese registro una sesión nueva vuelve a preguntar. Es la única casa de la regla de retomar (G3).
- `review-decisions.md:9` apunta a esa condición.

## 5. Delegación desatendida (`unattended-delegation/SKILL.md`)

- **`:10`:** la activación declara la medida y el tope cuando el trabajo repite ciclos, como fija `loops.md`.
- **`:12`:** el estado del loop no es un ledger nuevo.
- **`:35` (End):** registra el estado final y escala según `loops.md`, y «four sources» pasa a «three sources» (G2).

## 6. Sitio y documentación

**Sitio:**
- `site/src/content/docs/flows/research.md:52`: el estado en disco y la regla de sin progreso;
- `site/src/content/docs/flows/build.md:50`: al retomar en una sesión nueva, las respuestas salen solo de un plan guardado, nunca del estado del loop;
- `site/src/content/docs/skills.md:33` y `site/src/content/docs/index.mdx:100`: sin cambios por jobs.

**Guía:**
- `hive-core/SKILL.md:3`: su descripción suma los loops;
- `llms.txt:32`: la lista de temas de `hive-core`, si nombra los temas.

**Documentación interna:**
- `_support/docs/architecture/instruction-resources.md`: la casa canónica de `loops.md`;
- `_support/docs/architecture/agent-delivery.md`: las primitivas por host;
- `CHANGELOG.md`.

## Alternativas descartadas

- **D15-B**, el contrato dentro de `unattended-delegation`: no se leería en un loop con la persona presente.
- **D16-A, D19-A y D21:** jobs declarados con su propia entrega, alcance hasta integrar y estado en una rama propia. Los reemplazó D22-A durante el build.
- **D17-B**, que la ruta pare aunque corra desatendida.
- **D18-B**, un solo ciclo plano.
- **Tomar las respuestas de entrega del estado del loop:** ese archivo lo escribe el agente y no puede autorizar nada (B2 de autorización).
