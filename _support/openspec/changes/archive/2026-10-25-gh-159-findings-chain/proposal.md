# Tickets encadenados por hallazgos menores

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por PR #162 (`27fab1ff`) el 2026-10-25 |
| Tracker · GitHub Issues | • [#159 — Tickets encadenados por hallazgos menores: P3 ofrecidos como tickets, uno por defecto y revisiones de UX sin límite](https://github.com/JhonHawk/tricell-hive/issues/159) |
| Git | `automatic` · PR a `development` con `/code-review` · sin parada humana |
| Verificación | `go test ./tests/content/...` · conteo de frases · `go vet ./...` y `go test ./...` sobre el candidato final |
| Siguiente paso | Ninguno; la promoción a `master` va aparte |

## Solicitud

> otro detalle a agregar y es muy notorio en la sesión actual de claude en cam, un circulo vicioso en el tema de hallazgos -> resuelven -> mas hallazgos, y casi todos son detalles menores

> El tema es que: 1) Son tickets estrechamente relacionados con lo que ya se esta trabajando 2) cada defecto menor lo considera un ticket 3) Cada que resuelve un ticket, encuentra otros ~5 hallazgos, entonces termina resolviendo lo que acaba de encontrar, esos son los tickets encadenados

Respuestas: E5-A «Ticket + correcciones» y E6-A «G1 y G2 a #159; G3 vigilar» (2026-10-25).

Criterios del ticket [#159](https://github.com/JhonHawk/tricell-hive/issues/159), leídos el 2026-10-25:

> - **AC1:** los puntos de uso que proponen tickets por hallazgos (`flow-close` paso 5, `ui-gate.md`, `ui-review-criteria.md`) excluyen P3 y detalles de texto.
> - **AC2:** la guía dice qué hacer con varios hallazgos menores ligados al trabajo en curso en vez de abrir uno por defecto.
> - **AC3:** la pregunta de cierre no recomienda por defecto un ticket abierto en ese mismo cierre a partir de los hallazgos de la revisión.
> - **AC4:** la guía traduce la escala de severidad de la revisión de UX a P0–P3 por impacto observable.
> - **AC5:** la revisión de UX limita los hallazgos previos a las vistas cambiadas, y la de QA no repite lo revisado en local sobre el mismo árbol.
> - **AC6:** el umbral de costo de `verification.md:70` distingue el rol de la ronda.
> - **AC7:** la guía dice cómo tratar un arreglo cercano agregado al plan que crea una superficie de interacción nueva (G1).
> - **AC8:** un hallazgo de UI solo se propone como ticket después de confirmarlo renderizado (G2).
> - **AC9:** las páginas de `site/` afectadas reflejan el cambio.

## Objetivo

En una sesión real de un proyecto de cliente (2026-10-25, ~247 min), cada ticket de UI terminaba con una revisión de UX de pantalla completa, repetida en QA, que reportaba de 7 a 11 hallazgos menores y previos. El cierre proponía un ticket por cada uno, incluidos P3 y detalles de texto, y recomendaba seguir con el recién creado, así que la sesión se dedicó a resolver lo que ella misma acababa de encontrar. Unos 100 minutos se fueron en eso. El cambio corta la cadena en cuatro puntos: qué hallazgos pueden volverse ticket, cómo se agrupan, qué se recomienda como siguiente y cuánto abarca la revisión de UX.

## Alcance y aceptación

Incluye `content/skills/flow-close/SKILL.md` y una referencia nueva `content/skills/flow-close/references/incidental-findings.md` (el paso 5 no cabe: `SKILL.md` tiene 15 857 de 16 000 bytes), `content/skills/flow-walkthrough/references/ui-gate.md`, `content/skills/flow-walkthrough/references/ui-review-criteria.md`, `content/skills/flow-build/references/verification.md`, `content/skills/flow-plan/SKILL.md` y las páginas `site/src/content/docs/flows/close.md` y `build.md`.

Fuera de alcance: G3 (verificador que ignora su parada), en vigilancia; reducir la verificación funcional o la revisión del plan; los tickets ya abiertos en el proyecto del cliente.

Restricciones: `go test ./tests/content/...` y `TestReleaseLinksResolveWithinPayload` (`tooling/management`) siguen pasando; `flow-close/SKILL.md` ≤ 15 000 bytes tras T1; `content/guidance/global.md` no cambia (variación 0, declarada en el PR); cada regla vive en un solo hogar. La referencia nueva estrecha `global.md:32` en el cierre y nombra cada elemento que cambia, como exige la regla de excepciones de `AGENTS.md`: un P3 con escenario no se propone, los hallazgos se proponen agrupados y un hallazgo de UI sin confirmar renderizado no se propone. El filtro del cierre es la compuerta de tickets.

Correcciones aprobadas por el usuario sin segunda sesión (AGENTS.md, «Measurement»): L1 (AC1) y L4 (AC6). AC4 (traducción de severidad, L2) también es una corrección de `verification.md:43`, aprobada en E9-A. El resto son huecos que adoptan E5-A y E6-A.

- AC1. La referencia `flow-close/references/incidental-findings.md` aplica a todo hallazgo propuesto la regla P3 de `verification.md:43`, enlazándola (un P3 no se propone como ticket), y remite a `global.md:32` para los detalles de estilo o redacción; `ui-gate.md` y `ui-review-criteria.md` remiten a la referencia para disponer de los hallazgos previos. *Falso en la base cuando* `flow-close/SKILL.md:26` no menciona P3.
- AC2. La referencia exige un P0–P3 para cada hallazgo propuesto, con las definiciones de `hive-review-code.md:16`, y agrupa en una sola propuesta los P2 con la misma área o causa. Busca duplicados en el tracker por grupo, registra la respuesta al grupo para cada miembro y trata la reaparición de un solo miembro como recurrencia de ese miembro. *Falso en la base cuando* el paso 5 no menciona agrupar.
- AC3. La pregunta de cierre no marca como recomendado un ticket abierto en ese mismo cierre a partir de los hallazgos del trabajo, salvo que sea P0 o P1 o bloquee el siguiente ticket o un pendiente del trabajo. El ejemplo «a defect found» de `flow-close/SKILL.md:38` se ajusta a esa excepción. *Falso en la base cuando* la sección de la pregunta de cierre no lo dice.
- AC4. `ui-gate.md` traduce la severidad de la revisión de UX a P0–P3 solo para disponer de los hallazgos previos: Blocker→P0/P1, High→P1/P2, Medium→P2/P3, Nit→P3, con mínimo P2 para una falla WCAG AA o una tarea imposible con teclado o tecnología asistiva. Los hallazgos introducidos o empeorados siguen la regla Blocker/High de `ui-gate.md:26` y el veredicto de `ui-review-criteria.md`. *Falso en la base cuando* `ui-gate.md` no menciona P0–P3.
- AC5. La revisión de UX reporta hallazgos previos solo en las vistas que el cambio toca: pantallas con archivos o componentes cambiados, más las regiones que el cambio desplaza. Una revisión en un entorno desplegado cuyo árbol coincide con el revisado en local (`git diff <identidad local> <commit desplegado>` vacío en las rutas de UI afectadas) recibe el informe local como línea base y no vuelve a reportar sus hallazgos. Si el diff no está vacío, revisa las vistas afectadas por la diferencia. Esta excepción nombra la prohibición de decir qué ignorar (`ui-gate.md:26`, `ui-review-criteria.md:9`) y el alcance de pantalla completa (`ui-review-criteria.md:14,31`); el brief desplegado de `verification.md:85` la aplica. *Falso en la base cuando* `ui-review-criteria.md:69` trata todo hallazgo sin línea base como introducido sin excepción.
- AC6. El umbral de `verification.md:70` distingue `hive-verify-task` (5 min) de `hive-verify-change` y `hive-review-ux` (20 min). *Falso en la base cuando* la línea agrupa los tres roles en 5 minutos.
- AC7. `flow-plan/SKILL.md` dice que un arreglo cercano propuesto como agregado al plan que crea una superficie de interacción nueva (una parada de foco, un control, una ruta) lleva su propio criterio y entra en el alcance de la revisión de UI, o queda fuera del plan; `plan-revision.md` aplica lo mismo a los agregados durante el build. *Falso en la base cuando* `flow-plan/SKILL.md` no menciona superficie de interacción.
- AC8. La referencia propone un hallazgo de UI como ticket solo después de verlo renderizado; uno visto solo en el código queda como sin confirmar en el reporte. *Falso en la base cuando* el paso 5 no menciona renderizado.
- AC9. `site/src/content/docs/flows/close.md` y `build.md` describen AC1–AC5 y AC8 enlazando las skills. *Falso en la base cuando* ninguna menciona P3 ni agrupar hallazgos.

## Decisiones

- **E5-A, E6-A (2026-10-25):** ver la solicitud.
- **E3-A y E4-A (2026-10-25):** entrega `automatic` y revisión `/code-review`, declaradas por [gh-158](../gh-158-decision-cost/proposal.md).
- **E9-A (2026-10-25):** AC4 (L2) aprobado como corrección sin segunda sesión. **R1-A:** AC1–AC9 aprobados; build autorizado.
- **Decisiones menores tomadas sin preguntar:** 20 min para `hive-verify-change` y `hive-review-ux` (AC6): una revisión de UX completa tardó entre 6 y 20 min en la sesión observada, y la de 66 min, anómala, debe seguir saltando. Paso 5 a una referencia propia: es la única forma de que quepan AC1, AC2 y AC8 en el límite de tamaño. Mientras `development` no se promueva a `master`, los enlaces del sitio a la referencia nueva darán 404, porque `close.md` enlaza blobs de `master`.

## Entrega

`automatic`: rama `feat/gh-159-findings-chain` desde `development` en un worktree, commits, push, PR a `development`, `/code-review` sobre el candidato y merge cuando pasan las comprobaciones. Registro cerrado con un PR pequeño sin revisión. Sin despliegues. Tras el merge, `hive update`. Va antes que gh-158 y gh-160, que editan `flow-close/SKILL.md`. El PR lista los lugares revisados y la variación de tamaño de `global.md` (0).
