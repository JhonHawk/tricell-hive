# Diseño

## Contexto verificado (2026-09-26, `edf7dcf`)

- **La pregunta de revisión:** `delivery-decisions.md:18` exige resolver el mecanismo de revisión "for every mode including hold". Las opciones (`:20-23`) son el nativo del host, `review-code` con refuter, lo que declare `AGENTS.md` y "Other"; no hay opción "sin revisión". `:25` siempre recomienda un mecanismo.
- **La tabla de gates:** en `verification.md:30`, "Build iteration" ya dice que la autoinspección del diff "is not a dedicated independent code-review gate". En `:31`, "Local candidate" dice "Propose dedicated code review separately unless already authorized".
- **Rondas:** `verification.md:37` dice "do not assume unlimited rounds"; `:86` dice "Bound review/repair loops by material findings and explicit stopping conditions", sin un número.
- **Refuter:** `verification.md:39` manda al refuter los P0, P1 y los hallazgos bloqueantes; no dice nada sobre P2 ni P3.
- **UI:** `verification.md:57` tiene su propia regla para la revisión de UI: "fix introduced or worsened Blocker and High findings and re-review".
- **Prioridades:** `review-code.md:16` define P0–P3 ("P2 should be fixed but does not block; P3 is minor"). `:14` ya exige un escenario de falla concreto.
- **Ajuste `Review`:** `global.md:105` dice "A declared `Review` is the recommended option when asking about review, not a substitute for the question".
- **`review-security`:** ningún archivo de `content/` lo dispara fuera de su propia descripción.
- **Hive legacy** (`global/rules-situational/agent-routing.md:134-171`): tenía una matriz pasivo/estándar/sensible, "two rounds per verification cycle" y "An approving review closes review for the changeset".

## Enfoque

La revisión dedicada se recomienda por disparadores de riesgo, sin umbral numérico (D1-B). Cuando no hay disparador, la pregunta de entrega recomienda omitirla, pero deja elegir (D2-A). Los hallazgos se corrigen según su prioridad y hay una sola ronda de re-revisión, siguiendo el modelo de la revisión de planes. El reviewer filtra en origen los hallazgos que la guía de Claude Code identifica como fuente de sobreingeniería.

Se descartó un tope de dos rondas como el del legacy: la revisión de planes ya usa una sola ronda con escalamiento al usuario, y conviene una regla coherente entre ambas. Q5 no se edita porque `flow-build/SKILL.md:54` ya lo cubre.

## Redacción distribuida (inglés)

### `delivery-decisions.md` (nueva viñeta después de `:23`)

> - Offer “No dedicated review” as its own option: the implementer's own diff inspection and the project's checks, without an independent reviewer.

### `delivery-decisions.md:25` (reemplaza la primera frase)

"Mark the available mechanism explicitly preferred by applicable project guidance as Recommended; otherwise recommend the verified native option, or `review-code` where the host has none." pasa a:

> Recommend a dedicated review when the change touches a public or persisted contract, security, permissions, or stored data, or concurrency, or changes more behavior than its tests and local in-vivo checks can show to be correct. Then mark the available mechanism explicitly preferred by applicable project guidance as Recommended; otherwise recommend the verified native option, or `review-code` where the host has none. When the change has none of these and project guidance does not require review, recommend “No dedicated review” and name what the change lacks; the user may still choose a mechanism.

### `verification.md` (nuevo párrafo después de `:39`)

> When a mechanism reports its own severity levels, map each finding to P0–P3 by the same observable impact before applying these rules. Fix P0 and P1 findings that are confirmed or plausible after refutation, or verified by a mechanism that checks its own findings. Fix a P2 finding the change introduced when the fix is local and inside the change's scope; otherwise handle it under the shared incidental-findings rule. Report P3 findings without fixing or recording them elsewhere; they are not defects under the shared incidental-findings rule. Dispose of pre-existing findings under the shared incidental-findings rule. Re-review at most once, limited to the fixes, and only when a P0 or P1 fix changed logic or a contract; a finding still open after that round, including a new one it reports, goes to the user as a decision. An approving review closes review until a later edit changes logic or a contract. Use one code reviewer per candidate, and add `review-security` only when the change crosses a boundary in [security boundaries](security-boundaries.md) and within the agreed review authorization.

### `verification.md:31-32` (filas de la tabla de gates)

- `:31` "Propose dedicated code review separately unless already authorized." pasa a "Propose dedicated code review separately unless the delivery decision already settled it."
- `:32` "Review the identified candidate with the selected mechanism and inspect actual CI coverage." pasa a "Review the identified candidate with the selected mechanism, if one was selected, and inspect actual CI coverage."

### `review-code.md:14` (punto 3, antes de "Report nothing rather than guess")

> Do not report a finding whose only fix guards against, or tests, a case the change cannot produce; input that crosses a trust boundary can produce any case.

### `review-code.md:20` (cierre)

"End with a one-line verdict: whether the change is correct as written, with its main reason." pasa a:

> End with the number of findings at each priority and a one-line verdict: whether the change is correct as written, with its main reason.

### `global.md:105`

"A declared `Review` is the recommended option when asking about review, not a substitute for the question." pasa a:

> A declared `Review` is the recommended mechanism when a dedicated review applies, not a substitute for the question.

## Riesgos

- **Disparador por criterio (D1-B):** un modelo débil puede juzgar como "cubierto por tests" un cambio grande que no lo está. Lo atenúa la lista explícita de disparadores de riesgo y la posibilidad de que el usuario elija un mecanismo en la misma pregunta.
- **Hallazgos P2 mal clasificados:** un defecto real etiquetado como P2 no bloquea. `review-code.md:16` define P1 como "must be fixed before merge"; la calidad de la etiqueta depende del reviewer, y el refuter solo actúa sobre P0 y P1.
- **Una sola ronda:** si la re-revisión encuentra un P0 nuevo, no se abre otra ronda: pasa al usuario como decisión, igual que en `plan-review.md:34`.
