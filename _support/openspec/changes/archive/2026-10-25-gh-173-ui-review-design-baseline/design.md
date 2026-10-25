# Diseño · la revisión de UI respeta el sistema de diseño de la librería

## Contexto verificado

Base: `development` en `9822ef84` (2026-10-25). Las líneas citadas son de esa revisión.

- `content/skills/flow-walkthrough/references/ui-review-criteria.md`:
  - `:3`: «Project guides and accepted patterns refine these criteria; they never excuse a failed check.»
  - `:8`: la severidad solo baja por una regla escrita.
  - `:9`: «The one exception, named in the UI gate, limits pre-existing findings…; it does not limit introduced or worsened findings».
  - `:37`: tokens y escala.
  - `:38`: anatomía de la librería.
  - `:39`: «a shared defect is a systemic finding, not a defense».
  - `:55`: «meaningful non-text contrast is at least 3:1».
  - `:65`: «High … or a WCAG AA failure».
  - `:73`: el veredicto no pasa con un Blocker o High introducido; «a suggested fix».
- `content/skills/flow-walkthrough/references/ui-gate.md`:
  - `:26`: el brief no pre-califica, pero pasa los documentos del proyecto; «fix introduced or worsened Blocker and High findings and re-review».
  - `:28`: la excepción de los preexistentes «narrows, by name, the rule… it changes nothing else»; «at least P2 for a WCAG AA failure».
  - `:34`: «When a gate child reports an introduced Blocker or High finding, fix it».
- `content/skills/flow-build/references/verification.md:44`: «Report P3 findings without fixing or recording them elsewhere».
- `content/skills/flow-close/references/incidental-findings.md:9`: las respuestas por hallazgo son ticket, declinar (memoria con conteo) y «no volver a mencionar» (memoria, suprimido).
- `content/skills/flow-plan/references/ui-planning.md` (sección `Keep project knowledge current`, `:39`–`:45`):
  - hogar de la guía: la documentación existente, o `_support/docs/frontend/ui-conventions.md`;
  - enlazarla desde la guía de agentes o reportar que falta.
- `content/agents/review/hive-review-ux.md:10`: lee la guía UI solo para cambios de pantalla, shell, layout o patrón.
- `content/agents/design/hive-design-ui.md`:
  - `:11`: «Inspect existing screens, brand assets, tokens…».
  - `:15`: «Verify keyboard focus, contrast…».
- `content/agents/development/hive-build-frontend.md:15`: lee `ui-planning.md` y la guía UI.
- `content/skills/flow-plan/references/ui-conventions-template.md`:
  - `:13`: «Shared foundations».
  - `:27`: «Responsive/accessibility expectations».
  - `:29`: «Decisions and exceptions».
- `site/src/content/docs/flows/build.md`:
  - `:39`: «breakage the change introduced or made worse is fixed as part of the work».
  - `:79`–`:82`: la pregunta frecuente sobre cambios de UI.
- `content/skills/flow-plan/references/delivery-decisions.md:10`: `automatic` nunca se detiene para validación humana.
- #159 (cerrado 2026-10-25) ya limita los preexistentes a las vistas tocadas y excluye los P3 de los tickets.

**Evidencia del caso.** Proyecto de cliente con shadcn/ui; sesiones de Claude Code (2026-10-24) y Pi (2026-10-25):
- Borde de controles 1.26:1 / 1.48:1, valores por defecto de shadcn `neutral`.
- Cambio a 3:1 rechazado por el usuario y registrado como regla no adoptada.
- Otros tickets editaron componentes instalados, tokens y colores sueltos sin registrar la desviación.

**Evidencia externa.** W3C, Understanding SC 1.4.11 (actualizado 2026-09-06), descargado y leído en esta sesión:
- No exige un límite visual del control salvo cuando no hay otra señal.
- Figuras 31 y 32: un input de texto con etiqueta, cuyo indicador es un borde inferior o completo, «that indicator must meet 3:1».
- Figura 33: el caret visible cuenta como señal de foco.
- Figura 44: un caret personalizado de bajo contraste falla.
- Figura 45: input sin etiqueta y borde #AAA, falla.
- Componentes deshabilitados y estilo del navegador sin modificar: exentos.

Valores de shadcn: `--border`/`--input` cerca de 1.25–1.5:1; contorno de checkbox/radio sin marcar en 1.26:1.

## Enfoque

La regla vive en un solo hogar, `ui-review-criteria.md`. `ui-gate.md` agrega lo que le toca al hilo principal: el destino de una desviación y de un hallazgo global. Los demás archivos solo apuntan a ella y, donde la regla cambia un elemento de otra, nombran ese elemento en el punto de uso, como pide `AGENTS.md`. No se agrega una clave en `## Hive` (D1-A) ni código.

### Texto propuesto para `ui-review-criteria.md`

**Línea 3.** Reemplazar la última frase por: «The [design baseline](#design-baseline) sets which project and library sources take precedence over these criteria.»

**Línea 9 (Stance).** Agregar después de la excepción de los preexistentes: «A second exception, the [design baseline](#design-baseline), lets the brief carry the project UI guide's recorded decisions and global findings, which this review skips as that section sets; it never skips an introduced or worsened finding.»

**Sección nueva `## Design baseline`, entre Stance y Prepare the comparison:**

1. **Read.** Before the checks, read the design baseline:
   - the UI library the application uses and its configuration, such as `components.json`;
   - its theme tokens and installed components;
   - the project UI guide, with its theme deviations, rules not adopted and global findings;
   - any accessibility requirement the project declares in that guide or in its `AGENTS.md`, or that the ticket requests.

   An absent guide or requirement is a fact to report, not a failed check.

2. **Precedence, highest first:**
   1. a declared accessibility requirement;
   2. the project's recorded decisions and UI guide;
   3. the UI library's defaults;
   4. the screen's accepted pattern, which decides composition only and never excuses a failed integrity, state or accessibility check;
   5. the checks below.

3. **Library default.** A library default is a value of a theme token, or a style the installed component sets inside its own file, that the project has not changed. Establish it with these checks:
   - (a) `git diff <base> <candidate>` on the library configuration, the theme file and the installed-component folder: anything the change alters there is introduced by the change, not a default;
   - (b) the value matches the library's published value for the declared configuration, such as `style` and `baseColor` in `components.json`, or the installed package when the library copies no files, read from the library's current documentation or registry without private identifiers in the request (D3-A); and `git log -p --follow` shows no later edit to it. When the published value cannot be read, or without Git history, as in a shallow clone, the value counts as a project value, and the shared ones form one global finding;
   - (c) a class or style set where the component is used, rather than inside it, is a project value;
   - (d) position, layering and composition, such as a sticky header covering focus, are always project values.

   When the origin cannot be established, treat the value as a project value.

4. **Not applicable.** A library default, or a value a recorded project decision sets, is correct for this review: the check that measures it records `not applicable: library default` or `not applicable: project decision <reference>`, never a finding. Without a declared accessibility requirement this holds even when the value misses a WCAG threshold. With one, a failure the requirement covers is a finding, and its fix is a theme deviation.

5. **Global finding.** A global finding is one whose cause lives in a shared source: a theme token, an installed or shared component, or the shell or layout, as the checks in 3 locate it. A class set on one view over a shared token is that view's finding, not a global one.
   - Before reporting a finding or recommending a change, locate its cause.
   - Report a global finding once, at its source, with the measured value and the affected views as examples, not once per view.
   - Skip a global finding that the project UI guide already records only when the change does not touch its source and the measurement matches the recorded value. Otherwise report it, classified as introduced or worsened as below.

**Línea 39.** Cambiar la segunda frase por: «A shared defect whose cause is not a library default or a recorded decision is a global finding, as the [design baseline](#design-baseline) sets, not a defense.»

**Línea 55 (Accessibility).** Reemplazar el primer punto por:

- Text contrast is at least 4.5:1, or 3:1 for large text (SC 1.4.3, AA), including placeholder text and text on tinted surfaces.
- Non-text contrast of at least 3:1 (SC 1.4.11, AA) applies to the visual information required to identify a component or its state:
  - the indicator that shows where a text input or select is, such as its border, bottom border or fill, even when a label is present;
  - the boundary or fill of any other control that has no visible text or contrasting icon of its own;
  - checkbox and radio outlines, state indicators and meaningful icons;
  - a focus indicator the author styled.

  A focused text input's visible caret counts as its focus cue; a custom low-contrast caret does not.
- A button or link identified by its own visible text or a contrasting icon needs no boundary contrast.
- Non-text contrast does not apply to decorative borders and separators, to disabled components, or to a user-agent style the author did not modify.
- Each contrast finding names:
  - the success criterion and its level;
  - the element;
  - the cue that identifies it;
  - the measured ratio.

**Sección Severity and attribution.** Agregar al final:

A suggested fix uses the project's existing tokens, variants and components. Mark a fix as a `theme deviation` when it would:
- change a theme token;
- edit an installed library component;
- use raw palette values.

The parent applies a theme deviation only with an explicit decision, as [the UI gate](ui-gate.md#the-independent-children) sets.

**Verdict.** Agregar dos cosas:
- El reporte lista los hallazgos globales aparte, una vez cada uno, con su fuente, su valor medido y las vistas de ejemplo.
- Un hallazgo cuyo único arreglo es una desviación del tema rechazada por una decisión registrada pasa a `not applicable: project decision <reference>`.

### Texto propuesto para `ui-gate.md`

En `## The independent children`:

- **Brief del revisor.** Agregar a la lista: «the design baseline sources: the UI library configuration and the project UI guide».
- **Línea 28.** Agregar a la excepción que la base de diseño de [UI review criteria](ui-review-criteria.md#design-baseline) es una segunda excepción con nombre a la regla de no decir qué ignorar, y que no cambia nada más.
- **Arreglos.** Tras «fix introduced or worsened Blocker and High findings and re-review», agregar el destino de las desviaciones del tema:
  - «A fix the reviewer marks as a `theme deviation` needs an explicit decision before it is applied. A declared accessibility requirement is that decision for the deviations its failures need. A deviation the project UI guide records is that decision too.»
  - «For any other deviation whose finding must be fixed, an introduced or worsened Blocker or High, ask the user, by following [questions](../../flow-plan/references/questions.md), in every delivery mode, since it is a design decision and not a validation. The gate stays not passed until the answer.»
  - «Under [unattended delegation](../../unattended-delegation/SKILL.md), this question is an exception to "decide every question yourself": do not apply the deviation or record it as a rule not adopted; report it as a pending decision with the gate not passed.»
  - «A deviation suggested for any other finding goes to the report as a proposal and blocks nothing.»
  - «Record an applied deviation in the project UI guide's theme deviations. Record a declined one under rules not adopted; the finding then becomes `not applicable: project decision`.»
  - «After its measurement, an applied deviation goes to the human validation of [flow-walkthrough](../SKILL.md) when the delivery mode has one; under `automatic`, the closing report lists it with its revert.»
- **Hallazgos globales.** Agregar:
  - «An introduced or worsened global finding follows the fix rule above.»
  - «Give a pre-existing global finding one disposition: a theme deviation decided as above; for P2 or higher, the answer to the close question under [incidental findings](../../flow-close/references/incidental-findings.md), where declining or "don't mention again" also records it under rules not adopted; or, for P3, including a Nit, a record as reported with no action, without a question. This record is an exception to the global guidance's rule that puts style nits only in the report: a global UI nit also goes in the project UI guide, once (D4-A).»
  - «Record that disposition in the project UI guide's global findings, with its source and measured value, so a later review on another view skips it.»
  - «When the application has no UI guide, create it as [UI planning](../../flow-plan/references/ui-planning.md#keep-project-knowledge-current) sets, and leave its accessibility requirement empty until someone decides it.»

**Línea 34.** Cambiar «When a gate child reports an introduced Blocker or High finding, fix it» por «When a gate child reports an introduced Blocker or High finding, fix it as [the independent children](#the-independent-children) sets for theme deviations».

### Puntos de uso de otras reglas

- **`verification.md:44`.** Agregar a la frase de P3: «except a global UI finding, which the UI gate records in the project UI guide without a question».
- **`incidental-findings.md`, sección `Which findings are listed and how`.** Una línea que apunta: para un hallazgo global de UI, declinar o «no volver a mencionar» también se registra en la guía UI del proyecto, como fija `ui-gate.md`.

### Roles y planificación

- **`hive-review-ux.md`, paso 1.** Para todo cambio con efecto de UI renderizado, lee la base de diseño que fija `ui-review-criteria.md` antes de revisar.
- **`hive-design-ui.md`:**
  - Paso 1: agregar la configuración de la librería de UI y la guía UI con sus desviaciones (la base de diseño).
  - Paso 3: «Verify keyboard focus, contrast as [UI review criteria](skill:flow-walkthrough/references/ui-review-criteria.md) defines it…».
  - Una frase más: una propuesta que es una desviación del tema se devuelve al hilo principal como decisión pendiente.
- **`hive-build-frontend.md`.** Una línea: mantener la base de diseño; no hacer una desviación del tema (token del tema, componente instalado, colores sueltos) salvo que el brief nombre la decisión explícita.
- **`ui-planning.md`, sección `Keep project knowledge current`.** Dos líneas:
  - el plan identifica la base de diseño de la aplicación, y una personalización planeada es una desviación del tema que se decide y se registra en la guía;
  - cuando un documento del proyecto ya registra la base de la librería o las desviaciones de tokens, como un archivo de tokens o de base del framework, ese documento es la guía para las secciones de la base de diseño. No se crea un segundo archivo; solo sin ninguno se usa `_support/docs/frontend/ui-conventions.md`, como ya dice la sección.

  Motivo: el Hive legacy no tenía convención de nombres para UI, solo de ubicación (`_support/docs/<app>/`, `<project>-specs/conventions/`). Los dos proyectos que la aplicaron en parte terminaron con un archivo que ya cumple ese papel (`_support/docs/conventions/design-tokens.md` en uno, `docs/architecture/heroui-baseline.md` en el otro), además de guías duplicadas y desactualizadas. Reusar el archivo que ya existe evita archivos que mantener.

Los términos `design baseline` y `theme deviation` aparecen textuales en cada archivo, para que la verificación pueda buscarlos.

### Plantilla de guía UI

En `Scope and sources`, agregar:
- «UI library baseline: <library, configuration file, theme>»
- «Accessibility requirement: <declared level and scope, or none: the library baseline governs; leave empty until decided>»

En `Decisions and exceptions`, agregar:
- «Theme deviations: <token or component, library value, project value, reason, origin>»
- «Rules not adopted: <rule, reason, decision>»
- «Global findings: <finding, source, measured value, disposition, link>»

### Sitio

- **`build.md:39`.** Agregar: «…except a fix that would customize the UI library's theme or components, which needs your decision or the project's».
- **`build.md:79`–`:82` (pregunta frecuente).** Agregar dos frases:
  - la revisión mantiene el sistema de diseño de la librería como base, y una personalización solo se aplica con tu decisión o la del proyecto;
  - un hallazgo global se reporta una vez y queda registrado en la guía UI del proyecto.

## Alternativas descartadas

- **Clave `Accessibility` en `## Hive` (D1-B):** el gestor lee esas claves (`tooling/cli/doctor_project.go`, la vista de proyecto del TUI), así que sumaba código y pruebas sin una necesidad medida.
- **Mención única de las brechas AA de la librería (D2-B):** el usuario prefirió que la base de la librería gobierne sin mención.
- **Etiqueta como señal suficiente para un input de texto:** las figuras 31 y 32 de W3C lo contradicen. Con D2-A el resultado práctico no cambia para los valores de la librería.
