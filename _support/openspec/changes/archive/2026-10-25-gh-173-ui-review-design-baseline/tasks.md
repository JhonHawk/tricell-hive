# Tareas · gh-173-ui-review-design-baseline

Base del cambio: `9822ef84` (`development`, 2026-10-25). Diseño: [design.md](design.md).

## Tareas

### T1 — Base de diseño, contraste preciso y hallazgos globales en los criterios de revisión

- [x] `ui-review-criteria.md` lleva los textos que fija [design.md](design.md#texto-propuesto-para-ui-review-criteriamd):
  - las líneas 3, 9, 39 y 55;
  - la sección `## Design baseline`, con la lectura, la precedencia, la definición de valor por defecto, «no aplica» y el hallazgo global;
  - los agregados a Severity y Verdict.

  Ronda 1 tras `/code-review`, AC4 no cumplido: la definición de `theme deviation` (`c2273640`) agregó la condición «where the project has a token for that purpose», que deja fuera de la desviación los colores sueltos sin token. Además, el término de verificación «second exception» ya no aparece.

**Closes:** AC1, AC2, AC3, AC4, AC5, AC9

**Checkpoint:** C1

**Depends on:** ninguna.

**Locations:** `content/skills/flow-walkthrough/references/ui-review-criteria.md`

**Execution:** hilo principal, porque fija la redacción a la que apuntan T2–T5 y el brief sería más largo que el cambio.

**Test approach:** check, con la búsqueda y la lectura de abajo.

**Verification:**
- `rg -n 'never excuse' content/` → sin coincidencias.
- `rg -n -i '## design baseline|second exception|not applicable: library default|not applicable: project decision|published value|git log -p --follow|global finding|theme deviation|SC 1.4.11|caret' content/skills/flow-walkthrough/references/ui-review-criteria.md` → al menos una coincidencia por término.
- Lectura de la sección contra cada criterio:
  - AC1: lectura de la base, precedencia con el patrón limitado a la composición, segunda excepción en la línea 9.
  - AC2: «no aplica» también sin umbral WCAG, las comprobaciones (a)–(d) y el origen no establecido como valor del proyecto.
  - AC3: input de texto o select con etiqueta a 3:1, botón o enlace con texto, exentos, caret visible.
  - AC4: arreglo dentro del sistema y desviación rechazada como «no aplica».
  - AC5: dónde se declara el requisito y su efecto.
  - AC9: causa global, clase de una vista no global, reporte único con valor medido, omisión solo si la fuente no cambia y la medición coincide.

### T2 — Destino de desviaciones y hallazgos globales en la etapa de UI, el revisor y los puntos de uso

- [x] Cambios en `ui-gate.md`:
  - pasa la base al brief;
  - nombra la segunda excepción en la línea 28;
  - condiciona el arreglo que es desviación a una decisión explícita, también en la línea 34;
  - da un solo destino a cada hallazgo global y lo registra en la guía UI, creada según `ui-planning.md`.
- [x] `verification.md:44` e `incidental-findings.md` nombran la excepción.
- [x] `hive-review-ux.md` lee la base en todo cambio con efecto de UI.

**Closes:** AC1, AC4, AC5, AC9

**Checkpoint:** C2

**Depends on:** T1.

**Locations:**
- `content/skills/flow-walkthrough/references/ui-gate.md`: sección `The independent children`, líneas 28 y 34.
- `content/skills/flow-build/references/verification.md:44`.
- `content/skills/flow-close/references/incidental-findings.md`: sección `Which findings are listed and how`.
- `content/agents/review/hive-review-ux.md`: paso 1.

**Execution:** hilo principal, porque son frases cortas acopladas a T1.

**Test approach:** check.

**Verification:**
- `rg -n -i 'theme deviation|global finding|design baseline' content/skills/flow-walkthrough/references/ui-gate.md content/agents/review/hive-review-ux.md content/skills/flow-build/references/verification.md content/skills/flow-close/references/incidental-findings.md` → coincidencias en los cuatro archivos.
- Lectura:
  - AC1: segunda excepción en `ui-gate.md:28`.
  - AC4: decisión explícita en `:26` y `:34`; pregunta en todo modo; rechazo registrado como regla no adoptada.
  - AC5: el requisito declarado cuenta como decisión; validación visual o informe de cierre bajo `automatic`.
  - AC9: los tres destinos por prioridad, el registro en la guía con valor medido y la guía creada según `ui-planning.md`.

### T3 — Diseño, implementación y planificación de UI apuntan a la misma base

- [x] `hive-design-ui.md` (pasos 1 y 3), `hive-build-frontend.md` y `ui-planning.md` (sección `Keep project knowledge current`) apuntan a la base de diseño y no hacen una desviación del tema sin decisión explícita, solo con punteros a `ui-review-criteria.md`.

**Closes:** AC6

**Checkpoint:** C2

**Depends on:** T1.

**Locations:** `content/agents/design/hive-design-ui.md`, `content/agents/development/hive-build-frontend.md`, `content/skills/flow-plan/references/ui-planning.md`

**Execution:** hilo principal, por ser una línea por archivo.

**Test approach:** check.

**Verification:**
- `rg -n -i 'design baseline' content/agents/design/hive-design-ui.md content/agents/development/hive-build-frontend.md content/skills/flow-plan/references/ui-planning.md` → coincidencias en los tres.
- `rg -n -i 'theme deviation'` sobre los mismos archivos → coincidencias en los tres.
- Lectura: ninguno repite los pasos de la regla, solo apunta a ella.

### T4 — Plantilla de guía UI

- [x] `ui-conventions-template.md` tiene los cinco espacios de [design.md](design.md#plantilla-de-guía-ui), con el requisito vacío hasta que se decida y el valor medido en los hallazgos globales.

**Closes:** AC7

**Checkpoint:** C2

**Depends on:** T1.

**Locations:** `content/skills/flow-plan/references/ui-conventions-template.md`

**Execution:** hilo principal.

**Test approach:** check.

**Verification:**
- `rg -n 'UI library baseline|Accessibility requirement|Theme deviations|Rules not adopted|Global findings' content/skills/flow-plan/references/ui-conventions-template.md` → 5 coincidencias.
- `rg -n 'measured value per theme|undecided'` sobre el mismo archivo → 2 coincidencias (revisado tras `/code-review`: el requisito sin decidir se registra como `undecided`).

### T5 — Sitio y changelog

- [x] `site/src/content/docs/flows/build.md` tiene la excepción en la línea 39 y la pregunta frecuente «What happens to a user interface change?» explica la base de la librería, la desviación explícita y el hallazgo global único, enlazando `ui-review-criteria.md`. `CHANGELOG.md` anota #173 en `[Unreleased]`.

**Closes:** AC8

**Checkpoint:** C2

**Depends on:** T1, T2.

**Locations:** `site/src/content/docs/flows/build.md` (líneas 39 y 79–82), `CHANGELOG.md`

**Execution:** hilo principal.

**Test approach:** check.

**Verification:**
- `rg -n -i 'design system|global finding' site/src/content/docs/flows/build.md` → coincidencias.
- Lectura de la línea 39: tiene la excepción.
- Lectura de la página: describe el comportamiento y enlaza la referencia sin copiar sus reglas.

## Verificación y revisión

- **C1 — T1:** T2–T5 dependen de su redacción, así que un defecto ahí se propaga.
- **C2 — T2–T5:** los punteros, la plantilla y el sitio se verifican juntos al final.

| Etapa | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Verificación por punto de control | C1 y C2, árbol local | `hive-verify-task` por punto de control con las búsquedas y lecturas de cada tarea | parte de la implementación |
| Suite local | candidato final | `go test ./...` pasa, incluido `tests/content/skill_size_test.go` | parte de la implementación |
| Consistencia de la guía | candidato final | búsqueda de cada regla cambiada en `content/`, `site/`, `_support/docs/architecture/`, `tests/pilot/` y `tests/content/`; lugares listados en el PR | parte de la implementación |
| Revisión de código | PR a `development` | `/code-review` | `Review: /code-review` declarado |

No hay efecto de UI renderizado: no corren `hive-review-ux` ni `hive-verify-change`. Las pruebas piloto siguen en pausa.

## Entrega

- **Modo:** `Delivery: automatic`, declarado en `AGENTS.md`. Rama `feat/gh-173-ui-review-design-baseline` desde `development`, commits por tarea al reportarla, push, PR a `development`, `/code-review`, y merge cuando pasen las verificaciones y se atiendan los hallazgos.
- **Lo que no incluye:** promoción a `master` y release.
- **Commit y ticket:** el mensaje nombra las sesiones reales (Claude Code 2026-10-24, la de COM-645 en el proyecto del cliente; Pi 2026-10-25, `01a172af`) y el fallo observado, como pide `AGENTS.md` (Measurement). Al integrar, #173 se cierra.
- **Instalación local:** como existe el marcador de mantenedor, al llegar a `development` se corre `hive update`; es un cambio de contenido.
- **Registro de cambio:** se cierra con un PR pequeño sin revisión dedicada.

## Estado de la revisión y progreso

**Primera ronda (revisión `4d98e8d954ee`).** Dos `hive-review-plan` en paralelo:

| Dominio | Hallazgos | Duración | Tokens |
| --- | --- | --- | --- |
| Consistencia de la guía | 7 bloqueantes y 2 menores | 109 s | 117 k |
| Frontend/UI y accesibilidad | 3 bloqueantes y 2 importantes | 112 s | 108 k |

Todos se aceptaron tras verificar sus referencias en `9822ef84` y en el texto de W3C. Las correcciones:
- **Excepción con nombre:** la base de diseño es la segunda excepción a «no decir qué ignorar» en `ui-review-criteria.md:9` y `ui-gate.md:28`.
- **Omitir un hallazgo global:** solo si la fuente no cambia y la medición coincide; la plantilla guarda el valor medido.
- **Arreglo obligatorio:** `ui-gate.md:34` y el veredicto pasan por la decisión de desviación. Una desviación rechazada pasa a «no aplica: decisión del proyecto».
- **Regla no adoptada:** es la respuesta de declinar o «no volver a mencionar», registrada además en la guía. Los P3 globales se registran sin pregunta, con la excepción nombrada en `verification.md:44`.
- **Guía UI nueva:** se crea según `ui-planning.md#keep-project-knowledge-current`, con el requisito vacío.
- **Sitio:** se suma la línea 39.
- **Contraste:** un input de texto o select con etiqueta necesita 3:1 en su indicador (figuras 31 y 32 de W3C), y el caret cuenta solo si es visible.
- **Valor por defecto:** se define con las comprobaciones (a)–(d).
- **Requisito declarado:** cuenta como la decisión de las desviaciones que exige; las demás se preguntan en todo modo.
- **Patrón aceptado:** decide solo la composición.
- **Búsquedas:** sin distinción de mayúsculas y con términos fijados.

**Decisión tomada sin preguntar:** un valor cuyo origen no se puede establecer cuenta como valor del proyecto. Es la opción conservadora que recomendó el revisor: no oculta una falla, y la alternativa, «no verificado», bloquearía todo veredicto. Se revierte cambiando la regla 3 de la base de diseño.

**Re-revisión (proposal `35be1bb6019f`, design `cc06da9c4fba`, tasks `8f56d0a6562e`).** Los mismos dos revisores; los 12 hallazgos de la primera ronda quedan resueltos.

| Dominio | Duración | Tokens |
| --- | --- | --- |
| Consistencia | 50 s | 138 k |
| UI y accesibilidad | 45 s | 118 k |

Hallazgos nuevos:
- **Delegación desatendida:** la pregunta de desviación es una excepción con nombre a «decide every question yourself»; queda como decisión pendiente con el gate sin pasar. Se aplica tal como lo propuso el revisor, sin otra ronda.
- **Hallazgo global introducido o empeorado:** sigue la regla de arreglo, y los destinos quedan solo para los preexistentes. Se aplica tal como se propuso.
- **Cuándo se pregunta una desviación:** solo si el hallazgo debe arreglarse (Blocker o High introducido o empeorado); otra desviación va al reporte como propuesta. Se aplica tal como se propuso.
- **D3 (valor por defecto de la librería):** pasa al usuario como decisión.
- **D4 (registro de Nit globales frente a `global.md:33`):** pasa al usuario como decisión.

Hay que sumar `content/skills/unattended-delegation/SKILL.md` y `content/guidance/global.md:33` a la búsqueda de consistencia del candidato final.

**Decisiones del usuario (2026-10-25):** D3-A (comparar con el valor publicado) y D4-A (registrar también los Nit globales). Ya están aplicadas en `design.md` y `proposal.md`.

**Convención legacy (investigada con `hive-research`, 124 s, 140 k tokens).** El Hive legacy no tenía nombres de archivo para UI, solo la regla de ubicación. Ambos proyectos ya tienen un archivo que registra la base y las desviaciones. Se resolvió sin preguntar con una segunda línea en `ui-planning.md` (T3): ese archivo es la guía para la base de diseño y no se crea otro. Motivo: es lo que pidió el usuario, «no dejan de ser más archivos por mantener». Se revierte quitando esa línea.

**Hallazgos de otros proyectos.** Se reportan para la sesión de cada proyecto; no se tocan desde aquí:
- **Proyecto con shadcn (carvento):**
  - «Zinc» desactualizado en cinco lugares;
  - el bloque HSL de Tailwind v3 en las dos FRONTEND-GUIDELINES §10;
  - accesibilidad repartida en tres archivos;
  - COM-707 contradice `design-tokens.md` §3;
  - checkbox y radio en 1.26:1;
  - `--destructive-hover`, `--destructive-foreground` y sonner sin registrar.
- **Proyecto con HeroUI (CAM):**
  - `heroui-baseline.md` y `cam-monorepo/AGENTS.md` siguen tratando el mock congelado como referencia visual;
  - `visual-identity.md` tiene un acento desactualizado.

**Implementación:** sin empezar.

## Progreso del build

- Worktree `.claude/worktrees/gh-173`, rama `feat/gh-173-ui-review-design-baseline` desde `9822ef84`.
- Commits: T1 `aa3c308e`, T2 `3ac500c0`, T3 `5fd5671b`, T4 `b725d955`, T5 `cc406135`.
- **C1:** T1 verificada por `hive-verify-task`, con AC1–AC5 y AC9 en lo que le toca: met (66 s, 105 k tokens).
- **C2:** T2–T5 verificadas por cuatro `hive-verify-task` en paralelo:

  | Tarea | Criterios | Duración | Tokens |
  | --- | --- | --- | --- |
  | T2 | AC1, AC4, AC5, AC9 | 52 s | 107 k |
  | T3 | AC6 | 37 s | 103 k |
  | T4 | AC7 | 31 s | 87 k |
  | T5 | AC8 | 40 s | 90 k |

- **Modelos:** implementador, el hilo principal (Opus 5.5, observado). Verificadores, perfil `verifier` del rol: configurado, no observado.
- **`go test ./...`:** pasa en el candidato `cc406135`, incluido `tests/content`.
- **Búsqueda de consistencia** («systemic finding», «never excuse», «meaningful non-text», nits solo en el reporte, «WCAG AA failure», «fixed as part of the work») en `content/`, `site/`, `_support/docs/architecture/`, `tests/pilot/`, `tests/content/`, `README.md` y `llms.txt`: sin contradicciones.
  - `global.md:33` y `unattended-delegation/SKILL.md:16` quedan exceptuados por nombre en `ui-gate.md`, sin editarlos.
  - `content/guidance/global.md` no cambia de tamaño.
- **Nit del verificador de T2:** en `verification.md:44`, el «they» que sigue a la excepción insertada se puede leer de dos formas. Es solo de estilo y no se corrige en este cambio.

**`/code-review` del PR #174 (sobre `cc406135`, 124 s, 111 k tokens).** Revisé los 10 hallazgos y los 10 se sostienen. Se corrigen con commits adicionales:
- **T1 `c2273640` (P1):**
  - un componente que el cambio instala se compara con el valor publicado;
  - se acepta el valor publicado actual o el de la versión instalada;
  - la postura nombra sus dos excepciones y se resuelve su contradicción;
  - una decisión o un valor por defecto solo excusa los checks que miden ese valor.
- **T1 `c2273640` (P2):** una sola definición de `theme deviation` que excluye los colores sueltos permitidos por la guía, y la comparación de hallazgos globales por tema.
- **T2 `ddc2c022` (P2):**
  - el ticket abierto por un hallazgo global también se registra en la guía;
  - el valor medido se guarda por tema;
  - el requisito sin decidir se registra como `undecided`;
  - la frase de P3 de `verification.md` se reordena.
- **T3 `df4e26da` (P2):** los punteros ya no repiten la definición de desviación.
- **T4 `8d6868eb` (P2):** `undecided` y valor medido por tema.

T1–T4 vuelven a `[?]`, y T5 también porque depende de T1 y T2. La verificación de T4 se ajustó al término `undecided`. Hay una re-revisión pendiente, limitada al diff de los arreglos (`cc406135..8d6868eb`).

**Re-revisión de `/code-review` (`cc406135..8d6868eb`, 131 s, 108 k tokens), la única ronda permitida.** Trajo 9 hallazgos y los 9 se sostienen. Son locales, están dentro del alcance y alinean el texto con AC4 y AC9, así que se corrigen sin otra revisión, como fija `verification.md:44`:
- **T1 `322c3238`:**
  - el arreglo de la ronda 1 de AC4;
  - la desviación cubre cambios implementados y sugeridos, tokens nuevos y colores sueltos que la guía no permita;
  - una entrada que el propio cambio agrega a la guía sin decisión no cuenta;
  - `undecided` queda con un solo hogar.
- **T2 `8db7f994`:**
  - toda respuesta sobre un hallazgo global va a los hallazgos globales de la guía;
  - el valor de un tema faltante se agrega al registro sin nuevo destino;
  - una guía nueva arranca con el requisito que declare el `AGENTS.md`.
- **T4 `74c26a26`:** la plantilla ya no repite qué significa `undecided`.
- **T5 `4b1399d4`:** el sitio y el changelog dicen qué checks exime un valor por defecto.

T1–T5 se vuelven a verificar.

**Segunda re-verificación (HEAD `4b1399d4`, 127 s, 137 k tokens):** AC1–AC9 met en T1–T5. Siguió una observación sobre T3: en `hive-build-frontend.md` la regla quedaba dentro del paso para cambios de pantalla o patrón. Se movió al paso 1 en `e709c0ef` (T3), y T3 se vuelve a verificar sola, porque ninguna otra tarea depende de ella.
- T3 re-verificada tras `e709c0ef`: AC6 met (43 s, 103 k tokens).
