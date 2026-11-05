# gh-373 — Skill de criterio de UI (`hive-ui-criteria`)

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build · plan revisado, pendiente de tu aprobación |
| Tracker | [#373](https://github.com/JhonHawk/tricell-hive/issues/373) · abierto · relacionado con #173 (la revisión de UI empujaba cambios al tema: faltaba la base del sistema de diseño) |
| Git | JhonHawk/tricell-hive · base `development` en `dd729041` · rama `feat/gh-373-ui-skill` en el worktree `.claude/worktrees/gh-373-ui-skill` |

## Decisiones

2026-11-05
- **Alcance** — ¿Qué cubre la skill? Los puntos H9–H18 del research (modos de pantalla, registro de producto y de diseño, mínimos de accesibilidad, colecciones, textos, casos extremos, refinamiento) más tres requisitos del usuario: detectar la librería de componentes y su versión, darle prioridad junto con su capa de personalización, y generar un DESIGN.md para no volver a investigar el sistema de diseño cada vez. Debe funcionar con cualquier librería (por ejemplo HeroUI, ReUI y shadcn en React, PrimeNG en Angular).
- **Worktree** — El trabajo se hace en un worktree, a pedido del usuario.
- **D1 A** — ¿Dónde vive el criterio? En una skill propia; `flow-plan` y los roles de diseño, construcción y revisión solo la enlazan.
- **D2 A** — ¿Dónde vive el DESIGN.md del proyecto? En la raíz de la app dueña de la UI; en un monorepo, uno por app junto a su manifiesto de paquetes.
- **D3 A** — ¿Qué formato usa el DESIGN.md? El de Google Labs (tokens en YAML y prosa) como base, más una sección de Hive con la librería, su versión instalada, la capa de personalización con rutas y las huellas de las fuentes; su `lint` corre solo si ya está instalado.
- **D4 A** — ¿Cuándo se crea o refresca el DESIGN.md? Dentro del plan de un cambio de UI: si falta o está desactualizado, el plan incluye una tarea que lo escribe y se entrega con el cambio; fuera de un plan, solo cuando el usuario lo pide.
- **D5 A** — ¿Qué juzga la revisión UX como defecto? Además de lo actual, los mínimos de accesibilidad y lo medible de colecciones y textos, con prioridad P0–P3 y solo cuando el cambio los introdujo; el modo de pantalla y el refinamiento quedan como guía de diseño; la revisión lee el DESIGN.md como guía del proyecto.
- **D9 A** — ¿Qué pasa si un mínimo de accesibilidad falla en un valor por defecto de la librería sin modificar? Se resuelve como en #173. Va a la lista de defectos que ya existían, con su motivo, y arreglarlo (tocar el tema o un componente instalado) vuelve como decisión. Los mínimos son defecto en lo que el cambio construye o personaliza, y en todo cuando el proyecto o el ticket declaran un requisito de accesibilidad; ese requisito encabeza la precedencia.
- **D10 A** — ¿Cuentan como defecto medido el movimiento reducido y la devolución del foco? Sí, pero solo en lo que el cambio construye, citando su fuente: WCAG 2.3.3, marcado como AAA, y la guía APG del diálogo modal.
- **D11 B** — ¿Qué créditos van a la página de inspiración? Solo Impeccable (metodología, Apache-2.0) y Google DESIGN.md (formato), que también se nombran en el aviso de la skill y en `instruction-resources.md`. Las demás referencias no se nombran en archivos versionados, como manda `AGENTS.md` «Workspace locations».
- **D6 A** — ¿Qué nombre lleva la skill? `hive-ui-criteria`: el prefijo evita choques con skills de terceros en el almacén compartido de skills, y «criteria» la distingue del rol `hive-design-ui`. Es la primera skill de Hive con prefijo.

## Pedido

El usuario, 2026-11-05: «En [un proyecto de referencia] hay metido un flujo nuevo, una skill de diseño, inspirada en [Impeccable] y en https://www.tasteskill.dev/ ¿puedes validar si tiene caso que tengamos nuestro skill propia? (yo creo que sí)». Y después: «también se debe respetar la librería de componentes del proyecto […] debe ser compatible con todos, detectar: 1) Versión de la librería de componentes 2) Dar prioridad a la misma y validar si hay ya una capa de personalización 3) […] se genera el DESIGN.md».

El ticket #373 (leído el 2026-11-05) no tiene criterios numerados: lista el alcance, lo que queda fuera y las decisiones para el plan, que son las de arriba. Fuera de alcance según el ticket: prohibiciones estéticas universales, direcciones de arte al azar, puntuaciones /40 o /20, hooks, CLI obligatoria, modo live, y explorar varias direcciones de diseño en el visor.

## Objetivo

Hoy Hive cubre la entrega de UI (decisiones de UI en el plan, `hive-design-ui`, revisión UX medida, visor), pero no tiene criterio para crear y refinar pantallas: `hive-design-ui` pide mejorar «hierarchy, typography, spacing» sin umbrales, tres roles y el plan citan una «project UI guide» que nadie define, y el plan solo carga la librería que nombra la guía del proyecto, sin detectar su versión ni su personalización.

Cuando el cambio esté hecho, una skill `hive-ui-criteria` detecta la librería, su versión y su capa de personalización en cualquier proyecto, guarda el resultado en un DESIGN.md verificable y da a plan, construcción y revisión criterios con umbrales observables que nunca pasan por encima del sistema de diseño del proyecto.

## Alcance y aceptación

Dentro: la skill nueva y sus referencias; los enlaces desde `flow-plan` «UI decisions», `hive-design-ui`, `hive-build-frontend` y `hive-review-ux`; el spec delta; el sitio, `llms.txt`, `CHANGELOG.md` y `instruction-resources.md`. Fuera: lo que el ticket deja fuera, cambiar el nombre de otras skills o roles, y cualquier cambio al binario o al instalador.

- AC1 — `content/skills/hive-ui-criteria/SKILL.md` existe con `name: hive-ui-criteria`, se instala con el resto del catálogo y cumple los límites de tamaño y frontmatter · falso hoy porque no hay ninguna skill con ese nombre en `content/skills/`.
- AC2 — La skill da un procedimiento para cualquier librería que encuentra la versión instalada en el lockfile o en el árbol de dependencias instalado, no en el rango del manifiesto, y la capa de personalización (tema o preset, variables CSS, `@theme`, componentes envoltorio, código de componentes copiado), con HeroUI v2/v3 y PrimeNG v20 como ejemplos, y exige leer la documentación de esa versión · falso hoy porque ningún archivo de `content/` habla de detectar la versión de la librería de UI.
- AC3 — Un solo lugar canónico, la sección de precedencia de la skill, fija este orden: un requisito de accesibilidad que declara el proyecto o el ticket, luego las decisiones registradas del proyecto, la capa de personalización, los valores por defecto de la librería y al final los de la skill. Los mínimos de accesibilidad son defecto en lo que el cambio construye o personaliza. Un valor por defecto de la librería sin modificar que falla un mínimo va a la lista de defectos que ya existían, y arreglarlo vuelve como decisión. Los roles apuntan a esa sección en vez de repetirla · falso hoy porque esa precedencia no está escrita en ningún archivo de `content/`.
- AC4 — La skill define el DESIGN.md: secciones y tokens del formato de Google que usa, la sección de Hive (librería, versión instalada, capa de personalización con rutas, huellas de las fuentes), que cada valor nombra su archivo fuente, cómo se detecta que está desactualizado, que nunca se sobreescribe sin proponer el cambio, y que va uno por app en su raíz · falso hoy porque la «project UI guide» que citan `hive-design-ui.md:11`, `hive-build-frontend.md:15` y `hive-review-ux.md:12` no tiene definición.
- AC5 — `flow-plan` «UI decisions» carga el sistema de diseño a través de `hive-ui-criteria` y, cuando el DESIGN.md falta o está desactualizado, el plan lleva una tarea que lo escribe o refresca; fuera de un plan solo se escribe cuando el usuario lo pide · falso hoy porque el paso 1 solo carga lo que la guía del proyecto nombra.
- AC6 — `hive-design-ui` y `hive-build-frontend` leen el DESIGN.md como guía del proyecto y enlazan la skill, y el paso 2 de `hive-design-ui` aplica los criterios de la skill en lugar de adjetivos sin umbral · falso hoy porque `hive-design-ui.md` paso 2 dice «improving hierarchy, typography, spacing» sin criterio.
- AC7 — `hive-review-ux` reporta como defecto P0–P3, cuando el cambio lo introdujo, cada check que la skill marca **[measured]**: los mínimos de accesibilidad y lo medible de colecciones, textos y casos extremos. Un valor por defecto de la librería sin modificar y una decisión registrada siguen sin ser desviación: si fallan un mínimo, van a la lista de defectos que ya existían. El modo de pantalla y el refinamiento no se juzgan · falso hoy porque el paso 2 de `hive-review-ux.md` no incluye esos checks.
- AC8 — Las referencias de la skill cubren los modos de pantalla, los mínimos con su umbral y el estándar que lo fija, colecciones, textos, casos extremos y refinamiento con la clasificación de desviaciones, sin prohibiciones estéticas universales, sin puntuaciones y sin exigir hooks ni CLI · falso hoy porque ese contenido no existe en `content/`.
- AC9 — La metodología se atribuye a Impeccable (Paul Bakaus, Apache-2.0) sin copiar archivos ni pasajes, y se cita el formato DESIGN.md de Google, en el aviso de la skill, en `instruction-resources.md` y en la página de inspiración del sitio; ninguno de esos lugares nombra otros proyectos de referencia · falso hoy porque ninguno de esos lugares menciona a Impeccable ni a Google DESIGN.md.
- AC10 — El sitio (`skills.md`, `flows/plan.md`, `flows/verify.md`), `llms.txt` y `CHANGELOG.md` describen la skill y el comportamiento nuevo del plan y de la revisión · falso hoy porque ninguno la menciona.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

- 2026-11-05, base `dd729041`: `content/skills/flow-plan/SKILL.md:17-26` («UI decisions»); `content/agents/design/hive-design-ui.md:11-15`; `content/agents/development/hive-build-frontend.md:11-17`; `content/agents/review/hive-review-ux.md:10-12`; `content/skills/flow-verify/references/ui-verification.md`; spec `workflow-lifecycle` «UX review in one pass»; `_support/docs/architecture/instruction-resources.md` (una casa canónica por procedimiento, material adaptado); `_support/docs/architecture/deployment-manager.md:489` (un directorio de skill ajeno con el mismo nombre es conflicto).
- `tests/content/skill_size_test.go` limita cada `SKILL.md` y cada referencia a 16 000 bytes; los catálogos de skills se leen por glob, no por una lista fija.
- Fuentes externas leídas: Impeccable `c8891a9a` (2026-11-05, Apache-2.0); las skills de diseño de un proyecto de referencia, cuyo nombre queda en las fuentes locales (2026-11-04); README de google-labs-code/design.md (alfa, Apache-2.0; secciones Overview, Colors, Typography, Layout, Elevation & Depth, Shapes, Components, Do's and Don'ts; tokens `colors`, `typography`, `rounded`, `spacing`, `components`; secciones desconocidas se conservan). No se leyó el `docs/spec.md` completo de Google.
- WCAG 2.2 (w3.org/TR/WCAG22 y la página Understanding de 2.5.8, leídas el 2026-11-05):
  - 1.4.1 Use of Color (A);
  - 1.4.4 Resize Text (AA, 200 %);
  - 1.4.10 Reflow (AA, 320 px CSS);
  - 1.4.11 Non-text Contrast (AA, 3:1);
  - 2.3.3 Animation from Interactions (**AAA**);
  - 2.4.3 Focus Order (A);
  - 2.4.11 Focus Not Obscured (AA);
  - 2.5.8 Target Size (Minimum) (AA, 24×24 px CSS, con las excepciones de espaciado, equivalente, en línea, control del agente de usuario y esencial).

  La guía WAI-ARIA APG del diálogo modal dice que, al cerrar, el foco vuelve al elemento que lo abrió, salvo que ese elemento ya no exista o que otro sea más lógico para el flujo.
- Revisión de la base de diseño (#173, `CHANGELOG.md:633-643`): un valor por defecto de la librería o una decisión registrada no es hallazgo, aunque quede bajo un umbral de WCAG, salvo que el proyecto declare un requisito de accesibilidad o el ticket lo pida; un arreglo que personaliza el tema o un componente instalado necesita decisión.
- Documentación de librerías vía Context7: HeroUI v2 personaliza el tema en el plugin `heroui()` de Tailwind y v3 con variables CSS sobre `@heroui/styles`; PrimeNG v20 usa `definePreset` de `@primeuix/themes`, `pt` y `cssLayer`.
- Decisión previa: `engram:tricell-hive/decisions/ui-review-declared-linter` (2026-11-02, PR #337) rechazó delegar la revisión de UI a Impeccable y dejó su detector como evidencia opcional.

## Entrega

- **D7 A** — ¿Cómo se entrega? `automatic`: revisión, push, PR contra `development` y merge sin parar; después, el refresco de la instalación local que fija `AGENTS.md` para el mantenedor.
- **D8 A** — ¿Qué revisión de código corre? La nativa de Claude Code, `/code-review`, sobre el candidato antes del push.

No autoriza: un release, un piloto de comportamiento con modelos ni crear DESIGN.md en otros proyectos.

## Estado de la revisión y avance

2026-11-05 — plan escrito. Una revisión `hive-review-plan` encontró 13 puntos: la colisión con #173, dos mínimos sin un estándar AA, créditos sin aprobar, un consumidor que faltaba, estados que el brief del revisor no daba, la regla de datos de demo, criterios sin umbral o ya verdaderos, verificaciones parciales, `index.mdx`, una comprobación de desactualización que repetía la detección, fuentes WCAG sin verificar, dependencias de más y la precedencia repetida. Se aplicaron todos; tres se resolvieron con D9–D11.
