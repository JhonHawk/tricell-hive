# Menos pausas de confirmación en trabajo iterativo

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por PR #46, commit `16028d3575e7286cc028985cd2138c3268257add` · registro archivado |
| Tracker · GitHub Issues | Sin issue; la evidencia vive en este cambio y en Engram (`monitoring/codex-cam-excess-pauses-2026-10-11`) |
| Git | `feat/guidance-rework` integrada por [PR #46](https://github.com/JhonHawk/tricell-hive/pull/46) · publicación autorizada D3-A/D10-A el 2026-10-11 · cierre del registro por PR separado |
| Verificación | comprobaciones `rg` por criterio · `go test ./...` · medición en sesiones reales tras un despliegue que el usuario pida |
| Siguiente paso | Sin tareas técnicas pendientes de este cambio; la reorganización nueva continúa en su propio registro |

## Pedido

> Valida la sesión de codex en cam, sigue pasando el tema de que hace demasiada pausas,

Respuestas del usuario (2026-10-11):

- D1-A: aplicar R1–R3 con `flow-build` en la rama de prueba, sin commit (hecho antes de este plan; ver «Contexto verificado»).
- D2-A: planificar R4 ahora con `flow-plan`.
- D3-A: descartar H5 (ID reutilizado) y H6 (evidencia en `/var/folders`), conteo 1.
- D4-A: la continuidad sin preguntas se autoriza con una ruta iterativa explícita en la pregunta de ruta, no por inferencia.
- D5-A: el agente propone en la opción un número de vueltas acorde al caso, y la ruta para también cuando una vuelta no mejora la medición.

## Objetivo

En la sesión de Codex en cam `01a129d5` (2026-10-11), el agente se detuvo 13 veces para pedir una decisión y el usuario eligió la opción recomendada en todas. Cinco de esas pausas (D8–D12) fueron preguntas de ruta entre rondas de investigación e implementación, todas hacia un mismo objetivo (mejorar la precisión de un candidato local de CodeGraph) y todas con efectos solo locales y reversibles. `flow-research` ya permite llevar los hallazgos a una tarea más amplia autorizada sin otra aprobación, pero ninguna respuesta del usuario autorizó esa tarea más amplia: cada opción autorizaba un solo flujo, y la regla global «repeated approvals are not standing permission» empuja a preguntar en cada vuelta.

Este cambio añade una ruta iterativa a la pregunta de ruta. La opción declara el objetivo, los efectos que cubre (solo locales y reversibles), un número de vueltas que propone el agente y las condiciones de parada. Si el usuario la elige, el agente encadena rondas de `flow-research` y `flow-build` sin nuevas preguntas de ruta, informa en una línea por vuelta, y pregunta solo cuando la ruta se detiene. Resultado esperado: trabajo como el de cam sigue con una sola autorización y con un tope visible.

## Alcance y aceptación

Incluido:

- `content/skills/flow-research/SKILL.md:36`: la opción de ruta iterativa y sus paradas.
- `content/guidance/global.md`: la definición de unidad de trabajo (línea 34), la frase «The choice authorizes that flow only» (línea 31) y «Repeated approvals are not standing permission» (línea 42).
- `content/skills/flow-build/SKILL.md:75`: una vuelta verificada dentro de la ruta no es el punto de cierre.
- `site/src/content/docs/flows/research.md`, `site/src/content/docs/flows/build.md:31` y la entrada de `CHANGELOG.md` (R1–R4).

Excluido:

- Inferir la autorización de un objetivo sin que el usuario elija la ruta (D4-B, descartada).
- Un tope fijo de vueltas en Hive (D5-B) o ninguno (D5-C).
- Efectos fuera de lo local y reversible dentro de la ruta: commits, push, PR, instalación, escrituras en datos compartidos o configuración global siguen pidiendo autorización propia.
- La delegación desatendida (`unattended-delegation`), que no cambia.
- Detectores de `tests/pilot/` y casos de `tests/fixtures/regression/`.

Restricciones:

- `go test ./...` pasa; el tamaño de `global.md` se informa.
- Cada regla tiene un solo hogar: la ruta iterativa se define en `flow-research`; `global.md` solo la reconoce donde hoy contradice.
- Se conserva la intención de «A question, investigation, or proposal does not authorize implementation»: la ruta iterativa es una elección explícita del usuario.

Criterios de aceptación:

- AC1. La pregunta de ruta de `flow-research` puede ofrecer una ruta iterativa cuyo texto declara objetivo, efectos cubiertos (solo locales y reversibles), número de vueltas propuesto y condiciones de parada, y dice que elegirla encadena rondas de investigación e implementación sin nuevas preguntas de ruta. *Falso en la base cuando* `rg -n -i "iterative route" content/skills/flow-research/SKILL.md` no encuentra nada.
- AC2. `flow-research` define la vuelta (una investigación más una implementación con su medición) y enumera las paradas: objetivo cumplido, número de vueltas alcanzado, una vuelta sin mejora medible, un efecto que la ruta no cubre ni otro permiso autoriza, una vuelta que necesita `flow-plan`, un hallazgo que cambia objetivo o alcance, y una decisión que solo el usuario puede tomar; al parar con cambios en el árbol corre `flow-close` (fin o pausa), y sin cambios termina con la pregunta de ruta. *Falso en la base cuando* `rg -n -i "round limit|without a measurable improvement" content/skills/flow-research/SKILL.md` no encuentra nada.
- AC3. `global.md` cuenta una ruta iterativa elegida como unidad de trabajo, y `flow-build` no corre `flow-close` tras una vuelta verificada sino cuando la ruta para. *Falso en la base cuando* `rg -n -i "iterative route" content/guidance/global.md content/skills/flow-build/SKILL.md` no encuentra nada.
- AC4. `global.md` deja de contradecir la ruta: «The choice authorizes that flow only» y «Repeated approvals are not standing permission» reconocen la ruta iterativa elegida y mantienen fuera de ella commits, push, PR, despliegues, escrituras en el tracker y lectura de secretos. *Falso en la base cuando* la línea 42 dice solo «Repeated approvals are not standing permission.» sin excepción.
- AC5. `site/src/content/docs/flows/research.md` y `build.md` describen la ruta iterativa, y `CHANGELOG.md` registra R1–R4 bajo `[Unreleased]`. *Falso en la base cuando* `rg -n -i "iterative route"` sobre cada uno de esos tres archivos, por separado, no encuentra nada.

## Entrega final

La implementación quedó en `d5164f71` y luego recibió correcciones acotadas de revisión en `7ac9752` y `87290b9`. El usuario autorizó publicación e integración previa (D3-A) y construcción (D10-A); el PR #46 se fusionó en `16028d3575e7286cc028985cd2138c3268257add` el 2026-10-11 local. La ruta inicial `hold` que sigue en el historial ya no describe la entrega actual.

Revisión local con `hive-review-code`, solicitando OpenAI Sol, y observaciones automáticas de CodeRabbit corregidas; no P0/P1 en la revisión local. Go/vet/Python pasaron; checks de contenido/referencias y sitio pasaron tras los fixes. CodeRabbit aprobó el HEAD final. No hubo instalación global ni promoción a `master` en esta integración; un commit o merge no acredita mejora de conducta en sesiones reales.

## Contexto verificado histórico

- Sesión: `~/.codex/sessions/2026/10/11/rollout-2026-10-11T01-20-10-01a129d5-*.jsonl`, 16 turnos; la guía se recargó a las 08:14:21Z (versión f41da9b2). Las pausas D8–D12 citan la pregunta de ruta de `flow-research` y de los reportes de `flow-build`.
- `content/skills/flow-research/SKILL.md:36` ya dice «Within a broader authorized task, carry the findings into that task without treating this skill as another approval gate», pero no define cómo se autoriza esa tarea.
- `content/guidance/global.md:31` dice «The choice authorizes that flow only», y la línea 42, «Repeated approvals are not standing permission».
- R1–R3 aplicadas el 2026-10-11 sin commit, sobre la base `c332a814`: `global.md:42` (los hallazgos de una revisión hecha dentro de la implementación se corrigen dentro de ella), `flow-build/SKILL.md:18` (pregunta de revisión al inicio cuando no hay entrega por Git), `flow-build/references/verification.md` (tras la re-revisión, los arreglos locales sin contrato se corrigen sin otra ronda; fila «Local candidate»), `flow-plan/references/delivery-decisions.md` y `site/src/content/docs/flows/build.md`. `go test ./...` pasó; `global.md` subió de 46 561 a 46 683 bytes.
- Admisión de la regla (AGENTS.md, «Measurement»): decisión explícita del usuario (D2-A, D4-A, D5-A), con una sesión observada y su reporte de que «sigue pasando».

## Diseño

Ver [design.md](design.md).
