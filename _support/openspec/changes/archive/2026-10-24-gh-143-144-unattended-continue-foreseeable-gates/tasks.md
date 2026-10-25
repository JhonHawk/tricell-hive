# Tareas

Base: `6c0b5204` (2026-10-24). Rama `feat/gh-143-144-unattended-foreseeable-gates`, worktree `.claude/worktrees/gh-143-144`.

Convención de las comprobaciones: `rg -F -c "<frase>" <archivo>` debe dar `1` después del cambio. Antes del cambio no hay coincidencias en el archivo indicado. Ninguna frase existe en `content/` ni en `site/src/` en `6c0b5204`, salvo «Ask once for a set of actions», que ya está en `flow-close/references/reporting.md`; por eso su comprobación se limita a `verification.md` (búsqueda del 2026-10-24). Si el implementador ajusta el estilo de una frase, actualiza aquí la frase de comprobación antes de la verificación.

## T1 — Continuar, acusar recibo o preguntar por el estado mantiene la delegación desatendida (#143)

- [x] La skill y la guía global dicen lo mismo sobre cuándo expira la delegación.
Ronda 1: AC1 y AC2 cumplidos. Tras la ronda, la skill se alineó con `global.md` para que un mensaje solo informativo («reports something new») también termine la delegación. Ronda 2: AC1 y AC2 cumplidos sobre el texto final.

**Closes:** AC1, AC2.
**Depends on:** ninguna.
**Locations:** `content/skills/unattended-delegation/SKILL.md:8`, `content/guidance/global.md:34`.
**Execution:** hilo principal; son dos frases y el encargo sería más largo que el cambio.
**Test approach:** check.
- `rg -F "writes again" content/` y `rg -F "or the user returns" content/`: sin coincidencias.
- En `content/skills/unattended-delegation/SKILL.md`, una coincidencia de cada frase: «asks anything other than how the work is going», «A message in any language that only tells you to continue, acknowledges progress, or asks how the work is going» y «say that the delegation continues».
- En `content/guidance/global.md`, una coincidencia de «acknowledging progress, or asking how the work is going».

**Changes:** redacción de [design.md](design.md#redacción-propuesta). No toca la sección «Effects» ni la lista de fuentes.
**Verification:**
- Los comandos del enfoque de prueba dan lo esperado.
- Leer el párrafo de la skill y la viñeta global confirma que coinciden en los tres casos que mantienen la delegación y en los cinco que la terminan, y que ninguna frase amplía los efectos autorizados.

## T2 — Prueba costosa definida y autorizaciones previsibles pedidas en la ronda de entrega (#144)

- [x] La guía global define una prueba costosa. `verification.md` fija cuándo pedir las autorizaciones previsibles y solo exige evaluar las suites largas cuando son costosas. `flow-plan` remite a esa regla.

**Closes:** AC3, AC4, AC5, AC6.
**Depends on:** ninguna.
**Locations:** `content/guidance/global.md:68`, `content/skills/flow-build/references/verification.md:26` y `:39`, `content/skills/flow-plan/SKILL.md:36`, `content/skills/flow-plan/references/delivery-decisions.md:5`.
**Execution:** hilo principal; son frases cortas y el encargo sería más largo que el cambio.
**Test approach:** check.
- `rg -F "costly load/hardware" content/`: sin coincidencias.
- En `content/guidance/global.md`, una coincidencia de cada frase: «spends money or a third-party service's quota», «ordinary reads such as fetching dependencies do not count», «the project's CI, and ordinary reads such as fetching dependencies do not count» y «A local check of minutes with no external cost and no irreversible effect is routine».
- En `content/skills/flow-build/references/verification.md`, una coincidencia de cada frase: «while the user is present: in the round that settles the delivery and review questions», «Name each with its scope, target and limits», «an earlier answer or the request already settle those questions», «This is an exception to asking only at the decision point», «Ask once for a set of actions», «An authorization the user left unanswered stays unauthorized», «decide the ones never asked under its Effects section» y «costly as the global guidance defines it or can cause an irreversible effect». En `content/skills/flow-plan/references/delivery-decisions.md`, una coincidencia de «also when this round asks nothing».
- En `content/skills/flow-plan/SKILL.md`, una coincidencia de «ask the pending ones in the delivery and review round».
- `wc -c content/skills/flow-plan/SKILL.md` da 16 000 o menos.
- `rg -n "^## Gate selection and timing" content/skills/flow-build/references/verification.md` da una coincidencia; es la comprobación manual del ancla, porque ninguna prueba valida anclas.

**Changes:** redacción de [design.md](design.md#redacción-propuesta).
**Verification:**
- Los comandos del enfoque de prueba dan lo esperado.
- `go test ./tooling/management/... ./tests/content/...` en verde (enlaces relativos y tamaño de las skills).
- La lectura de `verification.md:26` confirma cuatro cosas: el momento es la ronda de entrega y revisión, no la pregunta de ruta; la excepción a `questions.md` está declarada; cada autorización va nombrada con su alcance, destino y límites y se agrupa según «Ask once»; y agrupar y el trato de producción se remiten a `reporting.md` sin repetirse.

## T3 — Una tarea agregada tarde se planea y verifica como cualquier otra (E1-B)

- [x] `plan-revision.md` dice cómo se trata una tarea agregada durante el build o tras la validación, incluida la que llega después de la revisión dedicada.

**Closes:** AC7, AC8.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-plan/references/plan-revision.md` (sección «Revise the change folder»).
**Execution:** hilo principal; una viñeta.
**Test approach:** check.
- En `content/skills/flow-plan/references/plan-revision.md`, una coincidencia de cada frase: «A task added during the build or after the user's validation», «criteria falsifiable at the base commit, a `Closes:` line and a test approach», «is verified by `hive-verify-task` before its commit», «never claims independent verification», «after the dedicated code review ran», «a task added during the build or after the user's validation, or a reconciliation on resume» y «anything they did not ask for is new behavior».

**Changes:** viñeta de [design.md](design.md#redacción-propuesta).
**Verification:**
- Los comandos del enfoque de prueba dan lo esperado.
- `go test ./tooling/management/...` en verde (los enlaces relativos a `plan-execution.md` y `verification.md` resuelven).

## T4 — Sitio y changelog

- [x] Tres páginas del sitio y el changelog describen los cambios.
Ronda 1: AC9 cumplido. Después se aplicaron dos observaciones del verificador en build.md: el verificador solo para tareas con prueba o chequeo nombrado, y el enlace a la guía global para la definición de costosa. Las frases de comprobación no cambian y el sitio compila.

**Closes:** AC9.
**Depends on:** T1, T2, T3.
**Locations:**
- `site/src/content/docs/index.mdx:90` (línea de `unattended-delegation`);
- `site/src/content/docs/flows/build.md` (lista de verificación cerca de la línea 41 y el párrafo de gates);
- `site/src/content/docs/flows/plan.md` (pregunta final del plan, cerca de la línea 64);
- `CHANGELOG.md` (`## [Unreleased]`, `### Changed`).

**Execution:** hilo principal; son tres frases del sitio y una entrada del changelog que dependen del texto final de T1–T3.
**Test approach:** check, con frases que el implementador escribe y anota aquí antes de la verificación. Frases anotadas: «acknowledges progress or asks how it is going does not end that» (index.mdx), «is planned and checked like any other» y «A check counts as costly when it spends money» (build.md), «are asked together with the delivery and review questions» (plan.md). Cada una debe faltar en la base y aparecer una vez.
- Una frase de `index.mdx` sobre continuar.
- Dos frases de `build.md`: prueba costosa y tarea agregada.
- Una frase de `plan.md`: autorizaciones junto con entrega y revisión.
- `npm run build` en `site/` en verde.

**Changes:** describir el comportamiento con enlace a la skill, sin copiar reglas (AGENTS.md, «README and llms.txt»). `README.md` y `llms.txt` no cambian, porque no describen estas reglas.
**Verification:** los comandos del enfoque de prueba dan lo esperado; la lectura de las tres páginas confirma que describen y no copian.

## Verificación y revisión

| Gate | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Verificación por tarea | Tras cada tarea, en el worktree | `hive-verify-task` con la carpeta del cambio, el ID y la base | Parte de la implementación |
| Suite local | Candidato final, en el worktree | `go vet ./...`, `go test ./...` y `npm run build` en `site/`, todo en verde | Parte de la implementación |
| Revisión dedicada | Sobre el PR, antes del merge | `/code-review` de Claude Code, una vez; una re-revisión limitada a los arreglos de P0/P1 que cambien lógica | E4 |
| Cuerpo del PR y commits | Al publicar | Cuerpo del PR: los lugares revisados por los tres mapas de impacto y la revisión del plan, el cambio de bytes de `global.md` y la evidencia por criterio. Mensajes de commit: las sesiones reales de [proposal.md](proposal.md) | AGENTS.md |

Sin superficie de UI: no aplican el gate de UI ni la validación humana visual.

## Estado de la revisión y progreso

- **Plan:** escrito el 2026-10-24 sobre `development @ 6c0b5204`.
- **Mapas de impacto:** tres hijos `hive-research`, uno por comportamiento, incorporados en [design.md](design.md).
- **Revisión del plan, ronda 1:** un hijo `hive-review-plan`, dominio de coherencia de la guía distribuida, sobre la revisión `1430821c4a95`. Coste: unos 123 000 tokens y 130 s. Siete hallazgos bloqueantes, todos aceptados:
  1. El momento de las autorizaciones pasa de la pregunta de ruta a la ronda de entrega y revisión (`reporting.md:98`, `flow-plan/SKILL.md:70`).
  2. Agrupar y el trato de producción se remiten a `reporting.md` en vez de repetirse.
  3. La excepción a `questions.md:11` se declara explícitamente.
  4. `global.md:34` alinea los casos que mantienen la delegación con la skill.
  5. La definición de costosa excluye el modelo de la sesión, sus hijos y la CI.
  6. Las comprobaciones usan frases exactas que faltan en la base, más lectura. Los enlaces se validan con `tooling/management` y el ancla a mano.
  7. Se añade `site/.../flows/plan.md`.
- **No bloqueantes aceptados:** las preguntas de estado mantienen la delegación; los ejemplos son válidos «in any language»; la verificación de tareas tardías ocurre «before its commit»; los IDs de sesión quedan en `proposal.md` para los commits.
- **Re-revisión (ronda única):** la hizo el mismo revisor sobre la revisión `9db9b9c45441`. Coste: unos 135 000 tokens y 39 s. Confirmó las siete correcciones y encontró un bloqueante nuevo: la regla de autorizaciones chocaba con la delegación desatendida y había perdido el «while the user is present». Lo resolví aplicando su propia propuesta, sin otra ronda (`plan-review.md`):
  - se restauró «while the user is present» y se añadió que, bajo delegación desatendida, esas autorizaciones se deciden según su sección «Effects»;
  - se ajustaron AC4 y las frases de comprobación de T2.

  Apliqué también dos no bloqueantes: «the one exception» pasa a «an exception», con alcance para el trabajo sin plan, y se corrigió la convención de comprobación sobre «Ask once for a set of actions».
- **Cobertura final:** coherencia de la guía distribuida, revisada en las dos rondas. Sin cambios materiales posteriores sin revisar, salvo esos ajustes, que siguen la propuesta del revisor. Sin superficie de UI, backend ni infraestructura.

- **Revisión de código (`/code-review`, PR #145):** 9 hallazgos de redacción y coherencia, todos introducidos por el cambio y todos aceptados, con revisión de AC3, AC4, AC5 y AC7:
  1. el silencio tras la ronda autorizaba lo no respondido;
  2. faltaba un puntero en `delivery-decisions.md`;
  3. no había ronda cuando entrega y revisión se reutilizan;
  4. «una pregunta con ID por cada una» chocaba con «Ask once»;
  5. las lecturas remotas normales contaban como costosas;
  6. las pruebas de hardware irreversibles perdían su evaluación;
  7. «Effects» no cubre gastar dinero bajo delegación;
  8. una corrección tras la validación podía saltarse la verificación;
  9. «When to read» de `plan-revision.md` no nombraba agregar una tarea.

  T2, T3 y T4 vuelven a `[?]` para re-verificar. Los arreglos van en commits adicionales `(T<n>)`, y la re-revisión se limita al diff de los arreglos.

- **Re-revisión de código (ronda única, `71592ae3..a6412273`):** de los 9 hallazgos anteriores, 4 quedaron resueltos y 5 parcialmente. Quedaron 10 puntos abiertos. Nueve los arreglé sin otra ronda, por ser locales, dentro del alcance y sin cambiar ningún contrato (`verification.md`):
  1. «Decisions» de la skill desatendida excluye las autorizaciones sin respuesta;
  2. las lecturas normales quedan fuera de toda la definición de costosa;
  3. una prueba con efecto irreversible no es rutinaria en `global.md`;
  4. «only the review question» incluye las autorizaciones previsibles;
  5. se define qué es corrección y qué es comportamiento nuevo;
  6. «When to read» incluye «after the user's validation»;
  7. el sitio distingue la prueba escrita primero;
  8. el sitio cubre el caso «or your request»;
  9. `plan.md` se alinea.
- **Hallazgo 7 (gastar dinero bajo delegación):** se retiró mi regla de omitir el gasto, porque contradecía el ejemplo de recurso en la nube de «Effects», que decidió el usuario. Queda como decisión del usuario.

- **Entrega:**
  - PR #145 fusionado con merge commit `2a100cd5` el 2026-10-24.
  - `go vet ./...`, `go test ./...` y `pnpm run build` en `site/` en verde sobre el candidato final `857f1056`.
  - `hive update` local aplicado con el plan guardado (resultado `0913501f…`).
  - Modelos: implementador en el hilo principal `claude-opus-5-5` (observado). Verificadores `hive-verify-task` lanzados con `model: opus` (configurado, no observado).
