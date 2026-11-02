# Diseño — gh-204

## Contexto verificado (2026-10-26, `development` en `2e45e292`)

- **Lecturas recortadas.** Comprobado en el rollout de Codex `01a179e0`: 13 lecturas de Hive con salida "truncated output (original token count: N)" y topes elegidos por el modelo; por ejemplo, 6 500 tokens para `delegation.md`, y 1 600 o 2 200 para `flow-close/SKILL.md`. En Pi `01a177d3`, las lecturas con `cut -c1-1400` y `cut -c1-900` vienen del informe del subagente auditor, no las comprobé yo. Ver #204.
- **Guía actual** (`content/guidance/global.md`):
  - `:80` pide releer `SKILL.md` completo tras compactar.
  - `:81` agrupa en un script las llamadas de solo lectura que "returns only what the answer needs" y deja directa la que se lee completa.
  - `:86` resuelve recursos.
  - Ninguna regla prohíbe topes, recortes o filtros al leer una skill.
- **Tamaños.** Solo `SKILL.md` tiene límite: 16 000 bytes, en `tests/content/skill_size_test.go` y HA-SK-02. Pasan de 16 000 bytes cuatro referencias: `delivery-decisions.md` (21 084), `plan-format.md` (20 514), `diagram-grammar.md` (17 663) y `verification.md` (16 692). Quedan cerca `reporting.md` (15 797) y `ui-gate.md` (15 041). Margen de `SKILL.md`: `flow-build` 794 bytes y `flow-close` 680.
- **Distribución.** No hay manifiesto manual. `loadRelease` recorre `references/` (`tooling/management/plan.go:230-246`, `:325-358`). Un archivo que sale de la release se retira de los hosts (`ownership.go:203-232`), cubierto por `catalog_test.go:178-182`. Una copia editada a mano hace fallar `hive update` (`deployment-manager.md:436`).
- **Pruebas atadas a rutas.**
  - `tests/content/reading_footprint_test.go:60-65`: el escenario del ciclo de plan y build lista los archivos viejos y lee la línea base con `git show d5164f71:<ruta>`, así que una ruta nueva fuera de `moved` hace fallar la prueba.
  - `tests/pilot/flows.go:357` busca los encabezados literales `# Retained plan format` y `## What the document must carry` en `plan-format.md`.
- **Anclas.** Ninguna prueba las valida (`instruction-resources.md:49`). Enlaces con ancla a los archivos que se mueven: `#dedicated-review-timing` (9 en `content/` y 2 en `site/`), `#gate-selection-and-timing` (5 y 2), `#acceptance-criteria-and-task-markers` (4) y `#hand-off-a-ready-plan` (1).

## Diseño

### Regla de lectura completa (AC1, AC2)

La casa canónica es el bullet "Resolve required resources" de `global.md:86`. Debe estar en la guía global porque tiene que actuar antes de que se lea cualquier skill. Texto propuesto, al final de ese bullet:

> When you read a skill's `SKILL.md`, one of its references, or a role file, read it whole, also when a pointer names one of its sections or asks for only its applicable parts, since such a pointer says what to apply, not what to read: never set a read's output limit below the file's size, cut its lines, such as with a column limit, or filter it, whether you read it directly or inside a script; when a tool stops before the end, such as at a line or output limit, continue from where it stopped until the end.

- La frase "When you read…" no obliga a precargar referencias. Así sigue valiendo leer una referencia solo cuando su condición aplica: `global.md:80`, `instruction-resources.md:5` y HA-SK-02.
- `:81` (scripts) gana un puntero, porque hoy manda devolver "only what the answer needs" incluso al leer tres referencias:
  > Skill, reference and role files are read whole, as the resource rule below sets.
- `:80` no cambia: no contradice la regla nueva.
- Otros punteros, que solo nombran la regla:
  - `flow-research/references/delegation.md:39`, con alcance acotado: "The child reads the skills, references and role files among these sources whole, as the global guidance sets, before dependent work". Las guías del proyecto quedan fuera, en línea con "not whole manuals" de `:35`.
  - `_support/docs/architecture/instruction-resources.md:13`.
  - `AGENTS.md` (Guidance structure): menciona la prueba nueva de referencias.
  - `_support/docs/architecture/repository-and-distribution.md:61-64` y `deployment-manager.md#grok-read-whole-hook`. Hoy dicen que el hook de Grok obliga a leer completo "not what any guidance says"; pasan a decir que la guía lo pide y que el hook lo hace cumplir en Grok.
- **Antecedente.** Grok ya tiene un hook que quita `offset` y `limit` en las lecturas de `.md` (`integrations/grok/readhook/readhook.go`). Se adoptó porque una de cada tres sesiones de Grok leía parcialmente. Este cambio no lo toca: la regla cubre a los hosts sin hook, como Codex y Pi.
- **Impacto en `global.md`:** unos +560 bytes sobre 39 432. El PR lo declara.

### Límite de tamaño de referencias (AC3, AC5)

- `TestSkillReferenceSize` en `tests/content/skill_size_test.go`, con constante propia `skillReferenceLimit = 16000` y comentario propio (el motivo es el recorte de lecturas, no la compactación).
- Recorre `content/skills/*/references` con `filepath.WalkDir` y falla si no encuentra ningún archivo.
- HA-SK-02 suma "each reference stays under 16,000 bytes, because a model that truncates a long read loses its later rules". Su espejo `_support/docs/harness-engineering/harness-audit-rules.md:41`, que hoy no menciona ningún límite de bytes, suma los dos límites, con #204 en su columna de fuente. El ID no cambia, así que `tests/skills/harness_audit_rules_test.py` no se rompe.

### Separación por responsabilidad (AC4)

Las fronteras caen entre párrafos completos y no se reescribe ninguna regla. Cada archivo nuevo empieza con su condición de lectura. Los encabezados que se mueven conservan su texto, así que el ancla solo cambia de archivo.

| Origen | Queda | Archivos nuevos |
| --- | --- | --- |
| `flow-plan/references/delivery-decisions.md` | modos, repositorios de apoyo, varios repositorios, registro de cambio, merge y registro de la decisión (~13,3 KB) | `review-decisions.md`: la pregunta del mecanismo de revisión (L26-34) y `## Dedicated review timing` (~6,8 KB). `ui-verification-decision.md`: los valores `full`, `functional` y `none`, que siguen siendo su único hogar (~1,9 KB) |
| `flow-plan/references/plan-format.md` | intro, contenido obligatorio, hoja de control, plantilla y cierre (~14,9 KB), con los dos encabezados que busca `tests/pilot/flows.go:357` | `plan-criteria-and-markers.md`: AC, checkpoints y marcadores (~4,3 KB). `plan-handoff.md`: entrega de un plan listo (~2,0 KB) |
| `flow-build/references/verification.md` | todo salvo las autorizaciones previstas (~14,4 KB) | `foreseen-authorizations.md`, en `flow-build/references/`: el tramo de L27 desde "Ask for the authorizations the plan foresees…" (~2,5 KB) |
| `flow-report/references/diagram-grammar.md` | cabecera y atribución, §1, §2, §6, §7 y §9 (~6,6 KB) | `diagram-svg-construction.md`: §3, §4, §5 y §8 con la misma numeración, y con su atribución MIT y el enlace a la licencia (~11,4 KB) |

Decisiones menores, tomadas sin preguntar:
- **`foreseen-authorizations.md` se queda en `flow-build`.** Es donde vive hoy, y así no cambia el dueño de la regla. Revertir: moverlo a `flow-plan/references/` y actualizar sus 7 enlaces.
- **La separación de `delivery-decisions.md` lleva el tiempo de revisión a `review-decisions.md`.** Reúne las dos reglas de revisión y deja 2,7 KB de margen. La alternativa, dejar el tiempo en el núcleo, solo deja 1 KB.
- **Se corrigen dos referencias § que ya estaban desfasadas** en `flow-report`: `skeleton-explainer.html:235` y `scripts/verify_geometry.py:5-6,22`. Se tocan igual al remapear los §, y las cubren las pruebas que ya corren.

- **Las frases que atribuyen al núcleo la pregunta de revisión o el valor de UI se corrigen con un enlace a la parte nueva.** Hoy esas frases están en `flow-build/SKILL.md:24`, `harness-audit/SKILL.md:65`, `flow-verify/SKILL.md:22`, `unattended-delegation/SKILL.md:16` y la L5 del núcleo ("below"). Es la propuesta (b) de la revisión de consistencia. La alternativa (a), separar en cambio el bloque L14–24, dejaría falsas otras 6 frases: `AGENTS.md:57`, `project-settings.md:9`, `flow-close/SKILL.md:78`, `flow-build/SKILL.md:64`, `flow-verify/SKILL.md:63` y `questions.md:56`. El enlace que se suma a `flow-build/SKILL.md` cuesta unos 80 bytes de sus 794 de margen.
- **`foreseen-authorizations.md` se escribe sin referencias internas.** Su frase "When a gate this paragraph confirms…" pasa a decir "a gate that [Gate selection and timing](verification.md#gate-selection-and-timing) confirms". Es la única reescritura de una regla y se declara como excepción al diff de líneas.

Reglas que cruzan archivos:
- La "declaración de una línea" que define el núcleo (L5) solo se nombra desde `review-decisions.md` y `ui-verification-decision.md`, sin repetirse.
- El cierre de `Dedicated review timing` remite a la regla de promoción con un enlace explícito al núcleo, en lugar de "near the start of this reference".
- La regla de checkpoints conserva su "lives only here" en `plan-criteria-and-markers.md`. La plantilla de `plan-format.md` enlaza ahí para `Closes:`, `Checkpoint:` y "False at the base when".

Punteros en skills. `flow-close/SKILL.md` no se toca, porque le quedan 680 bytes de margen y sus frases siguen siendo ciertas. `flow-build/SKILL.md:24` solo recibe el enlace de la decisión de arriba. Cambian:
- `flow-plan/SKILL.md:22`: enlaza `review-decisions.md` y `ui-verification-decision.md` con su condición.
- `flow-plan/SKILL.md:40`, `:48` y `:74`: cambian de destino con el mismo texto.
- `flow-report/SKILL.md:27`: suma un puntero a `diagram-svg-construction.md` ("before writing or changing SVG markup").
- `flow-verify/SKILL.md` (×5) y `unattended-delegation/SKILL.md:16`: cambia el destino del ancla.

### Pruebas y documentación atadas a rutas

- `reading_footprint_test.go`: el escenario del ciclo suma las partes nuevas que lee a `cycle` y a `moved`. Esas partes son `review-decisions.md`, `plan-criteria-and-markers.md`, `plan-handoff.md` y `foreseen-authorizations.md`. `ui-verification-decision.md` no entra, porque el escenario no tiene UI. Los núcleos conservan su nombre y siguen en la línea base A, y una ruta nueva fuera de `moved` haría fallar `git show d5164f71:`.
- Sitio:
  - `site/src/content/docs/flows/build.md:48,50`.
  - `site/src/content/docs/flows/verify.md:28`.
  - `site/src/content/docs/flows/plan.md:34,69`.
- `_support/docs/architecture/instruction-resources.md:7,108`: el mapa del material de terceros nombra los dos archivos de diagramas.
- `CHANGELOG.md [Unreleased]`: una entrada, siguiendo el precedente del límite de `SKILL.md`.
