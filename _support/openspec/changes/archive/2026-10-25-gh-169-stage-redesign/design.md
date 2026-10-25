# Diseño

## Contexto verificado

Leído en `development` `f4946d63` (incluye #167, #168 y #173) el 2026-10-25, con dos investigaciones delegadas de esta sesión.

- **Verificación dentro de build.** `content/skills/flow-build/SKILL.md:52-69` («Verify and close», ~5,1 KB) tiene la secuencia previa a validar (puerta de UI, suite completa, oferta de seguridad, walkthrough), las reglas de suite y evidencia y el cierre. `flow-build/SKILL.md` mide 15 168 bytes.
- **`verification.md`** (26 694 bytes, 24 referencias en `content/`): mezcla selección de evidencia (`:7-33`, `:70-74`, `:87-97`), revisión de código y entrega humana (`:34-35`, `:42-46`, `:50-56`, `:58-62`, `:64-66`) y despliegue (`:36-37`, `:60` recorrido desplegado, `:74`, `:77-86`).
- **Puerta de UI repartida.** `flow-walkthrough/SKILL.md` (15 540 bytes) dice ser dueña de la puerta (`:3`, `:8`) y tiene las rondas de corrección y la puerta del candidato aceptado (`:60-66`); `references/ui-gate.md` (14 430 bytes) y `references/ui-review-criteria.md` (13 684 bytes) son lógica de verificación. Presentación, visor y primera vista renderizada (`:22-56`, `:68+`) son de apoyo. El visor depende de la ruta `content/skills/flow-walkthrough/` (`tooling/viewer/*`, `tests/skills/walkthrough_*`).
- **Despliegue repartido.** `verification.md:36-37,60,74,77-86` (~5 KB), `delivery-decisions.md:40-46` (arreglos urgentes, ~1,65 KB), `git-workflow/SKILL.md:44` (ruta de promoción) y `global.md:78` (puntero «before … promotion, or deployment, read flow-build»).
- **Seguridad.** «Offer the security review» vive en `flow-build/references/security-boundaries.md:12-18` (#167); el rol `hive-review-security` existe; `security-advisory` es otra cosa (barrido de avisos entre repositorios). Claude Code trae una skill propia `security-review`; no se encontró catálogo de skills propias de los otros hosts (sin verificar).
- **Distribución.** Las skills se descubren por carpeta (`tooling/management/plan.go:230`); no hay registro. Pruebas que fijan estructura: `tests/pilot/flows.go:548,550`, `flows_test.go:221,756`, `protected.go:56`, `tests/content/reading_footprint_test.go:52-79`, `tests/fixtures/flows/cases.json`. Enlaces y tamaños los comprueban `tests/content/skill_links_test.go` y `skill_size_test.go`.
- **Skill core (descartada).** `flow-core` del Hive anterior (`abad754d^`, 26 archivos, ~231 KB, `user-invocable: false`); la lectura medida de cadenas enrutador → referencia fue ~50–73 % (memoria de #24).

## Enfoque

**Etapas y contratos.** Cada skill de etapa abre con un párrafo «Entry and exit» de dos o tres frases: qué necesita para empezar (y que se puede entrar directamente cuando se cumple) y qué entrega a la siguiente. El orden se lee de los contratos; no hay índice ni enrutador.

| Etapa | Entrada | Salida |
| --- | --- | --- |
| `flow-research` | Una pregunta, una falla o una propuesta | Hallazgos con evidencia y la pregunta de ruta |
| `flow-plan` | Un objetivo con alcance | Plan listo (carpeta de cambio) y decisiones de entrega |
| `flow-build` | Trabajo autorizado (plan o pedido) | Candidato construido: commits (o árbol con `hold`), tareas `[x]` por puntos de control, pruebas afectadas en verde, primera vista aceptada cuando aplica; nombra lo que debe `flow-verify`. Cuando verify se omite, integra él con `git-workflow`: en un cambio mecánico informa los chequeos que ya lo cubren; si lo pidió el usuario, nombra los chequeos omitidos como aceptados sin verificar (D21-A) |
| `flow-verify` | Candidato identificado, criterios y valores de revisión resueltos | Nivel corrido y cada chequeo no cumplido nombrado; candidato aceptado e integrado según el modo de entrega, con `git-workflow` (o árbol verificado con `hold`) |
| `flow-security` (a pedido, dentro de verify o después del merge) | Pedido explícito del usuario y un candidato o commit | Hallazgos priorizados y cobertura; no corrige |
| `flow-deploy` (etapa aparte, después de integrar; puede abarcar varios tickets) | Candidato integrado e identificado, entornos declarados, autorización por entorno | Por entorno: identidad servida, salud y aceptación funcional, o «no corrido» |
| `flow-close` | Punto de cierre alcanzado, abandono, pausa o pedido de higiene | Registro cerrado, limpieza e informe |

**Nivel de `flow-verify` (D17-A).** Sin clave nueva. Al entrar, una línea dice si corre o se omite y con qué nivel. Se omite solo en un cambio mecánico (definición de `global.md`) o cuando el usuario lo pide; lo omitido se nombra en el informe. El nivel combina los valores que ya existen, en una tabla única dentro de `flow-verify/SKILL.md`:

| Chequeo | Corre cuando | Lo fija |
| --- | --- | --- |
| Pruebas afectadas y evidencia | siempre que verify corre | — |
| Suite completa local | cambio no mecánico | regla actual de suite |
| Prueba en vivo local | el comportamiento cambiado lo pide; si parece innecesaria, se ofrece | regla actual de `verification.md` |
| Puerta de UI | efecto de UI renderizado | `UI verification` (`full`, `functional`, `none`) |
| Revisión de código dedicada | según el modo de entrega | `Review` o la respuesta a su pregunta |
| Revisión de seguridad | solo si el usuario la pidió | `flow-security` |

El usuario puede subir o bajar el nivel por pedido («verify completo», «sin revisión de UI», «sin verify»); el pedido vale para esa obra. Omitir verify o bajar su nivel, por pedido del usuario, cuenta como aceptar sin verificar cada chequeo omitido (D21-A): se nombran uno por uno al entrar y en el informe, y un chequeo aceptado así no es una puerta pendiente. Solo un mensaje del propio usuario lo pide: la delegación desatendida no puede omitir verify ni bajar su nivel. D21-A nunca exime la CI, la protección de ramas ni una revisión que el proyecto exija. Las frases que hoy dicen lo contrario sin esa distinción se ajustan en T2: `global.md:71`, `delivery-decisions.md:24` y `verification.md:50` (que pasa a `review-and-gates.md`).

**Integración (D24-A).** Las reglas de push, PR, revisión, merge y espera de CI siguen solo en `git-workflow` y `delivery-decisions.md`. Las ejecuta la última etapa que corre antes de `flow-close`: `flow-verify` en el flujo normal; `flow-build` cuando verify se omite. La revisión de código dedicada corre dentro de `flow-verify` en el punto que fija el modo de entrega. `flow-build` no la corre: allí solo quedan los chequeos del implementador y la verificación por tarea. Las preguntas de qué se ejecuta (modo de entrega, mecanismo de revisión, `UI verification`, autorizaciones previstas) se siguen resolviendo donde hoy, con `delivery-decisions.md`: al cerrar el plan, o al empezar `flow-build` cuando no hay plan, y los valores declarados en `## Hive` se anuncian sin preguntar. `flow-verify` no agrega preguntas: lee esas respuestas, anuncia el nivel y las ejecuta; solo cuando se entra directo a verify sin plan ni build hace esa misma ronda una vez, y solo para los valores que falten. Un `Review` declarado es el mecanismo por defecto, no una exigencia: cuando el usuario pide omitir verify, la revisión cuenta como aceptada sin verificar (D21-A), salvo que el repositorio la exija con una protección de rama o una revisión requerida; en un cambio mecánico no hace falta revisión dedicada, como hoy.

**Movimientos.**

- `flow-verify/SKILL.md` (nuevo, ~7–8 KB): contrato, decisión de nivel, secuencia previa a validar (de `flow-build/SKILL.md:54-61`), suite y evidencia del candidato final (de `:63`), pila verificada para validación humana (de `:65`), rondas de corrección y puerta del candidato aceptado (de `flow-walkthrough/SKILL.md:60-66`).
- `flow-verify/references/review-and-gates.md` (nuevo): de `verification.md` las filas «Local candidate» y «Publication / integration» y los párrafos `:42-46`, `:50-56`, `:58-62` (sin la frase del recorrido desplegado) y `:64-66`.
- `flow-verify/references/ui-gate.md` y `ui-review-criteria.md`: `git mv` desde `flow-walkthrough/references/` (D19-A). Sus enlaces relativos `../SKILL.md` (`ui-gate.md:3,19,28,36,38,46`) se corrigen a `../../flow-walkthrough/SKILL.md` o a `../SKILL.md` de `flow-verify` según lo que nombran; los enlaces entrantes, incluidas las formas relativas `references/ui-gate.md` de `flow-walkthrough/SKILL.md`, se corrigen.
- `flow-deploy/SKILL.md` (nuevo) y `flow-deploy/references/deployment-checks.md`: de `verification.md:36-37`, la frase de `:60`, `:74`, `:77-86`; `flow-deploy/references/urgent-fixes.md`: de `delivery-decisions.md:40-46`. `git-workflow/SKILL.md:44` conserva el invariante de promoción y remite a `flow-deploy`; `global.md:78` cambia su puntero de despliegue a `flow-deploy` sin crecer.
- `flow-security/SKILL.md` (nuevo, D22-A): cómo correr la revisión pedida, que hoy está en `security-boundaries.md:18` (antes de validar o antes del push o merge, límites posteriores del mismo cambio, pedido después del merge, puerta de UI tras un arreglo), más el brief a `hive-review-security`, el informe y el alcance; remite al paso 5 de `security-advisory` para alcanzabilidad, sin copiarlo. `security-boundaries.md` conserva los límites, la verificación de dependencias y la regla de ofrecer (`:14-16`), que remite a `flow-security` para correrla; «a boundary this reference covers» sigue siendo cierto porque la oferta no se mueve.
- `verification.md` queda en `flow-build` como referencia compartida (24 referencias) y remite a las referencias nuevas.
- Frases de propiedad que quedarían falsas y se reescriben: `verification.md:58,66`; `ui-gate.md:3,19,36`; `flow-build/SKILL.md:34,61,67-69` (entrega y cierre pasan a depender de D24-A); `plan-revision.md:11,13`; `content/agents/review/hive-review-ux.md:14`; `plan-format.md:47` («flow-build owns the transitions»: pasa a «flow-build y flow-verify»); `global.md:78` (antes de un commit, PR o merge se lee la etapa que integra: `flow-verify`, o `flow-build` cuando verify se omite; antes de una promoción o despliegue, `flow-deploy`); `flow-walkthrough/SKILL.md:13-14` («only the UI gate runs» nombra la puerta de `flow-verify`); `project-settings.md:9` sigue diciendo que `flow-build` pregunta `UI verification`, porque la pregunta se queda en la entrada de build; `delivery-decisions.md:9` (la cadena interactiva nombra `flow-verify`); `plan-execution.md:40` (los arreglos de suite o validación se commitean igual, en la etapa que los pide).
- `flow-build/SKILL.md` queda en ~11–11,5 KB; `flow-walkthrough/SKILL.md` pierde ~2 KB.

**Descripciones (borrador, sin solaparse).**

- `flow-verify`: «Verify a built candidate before it is accepted: affected tests, runtime checks, the UI gate, the full suite and the dedicated review, at the level the project's settings and the request set. Use when a candidate exists and needs evidence; not for writing code, deploying, or checking a deployed environment.»
- `flow-deploy`: «Promote an integrated candidate through the declared environments and verify what is served: rollback triggers, served-identity and health checks, and urgent production fixes. Use when the user asks to promote or deploy, or for a production defect that cannot wait; not for committing, opening pull requests for, or merging the work itself, though it opens and merges the promotion pull request.» (texto ajustado por D27-A)
- `flow-security`: «Review a change's security boundaries with an independent reviewer and report prioritized findings without remediating. Use only when the user asks for a security review; not for advisory sweeps across repositories or for checking a package before installing it.»
- `flow-build` deja «promotion, or deployment» y «verify results, and close»; `flow-walkthrough` deja «own the UI gate» y remite a `flow-verify`.

## Riesgos

- Elección de skill por descripción: «verify», «deploy» y «close» se pueden pisar; los borradores excluyen lo de las otras. Sin pilotos no se mide.
- Lectura: una etapa más significa una lectura más de `SKILL.md` en un ciclo completo; a cambio, `flow-build` se lee más corta. `reading_footprint_test.go` lo fija.
- Muchos punteros: `tooling/management/references_test.go` (`TestRepositoryCatalogueInstructionReferences`) detecta archivos de enlace rotos, pero no anclas de encabezado (`references.go:12`); las anclas que se mueven (`#offer-the-security-review`, `#urgent-production-fixes`) se comprueban con `rg`. `skill_links_test.go` solo cubre los enlaces `skill:` de `global.md`.
- Copias instaladas viejas: `hive update` debe retirar `flow-walkthrough/references/ui-gate.md` de los hosts; se comprueba con `hive update --dry-run` sobre el candidato final.
