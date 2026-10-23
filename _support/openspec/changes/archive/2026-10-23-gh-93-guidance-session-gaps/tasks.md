# Tareas

Commit base: `a6255b6f` (development, 2026-10-19).

Todas las tareas se ejecutan en el hilo principal. Son cambios de redacción en reglas distribuidas, y el brief que necesitaría un hijo sería más largo que el cambio. Además, T1 y T2 escriben en `global.md`, y T3 y T4 en `ui-gate.md`, así que delegarlas cruzaría escritores.

Cada tarea hace lo que `AGENTS.md` (Guidance structure) pide para cambiar una regla distribuida:
- revisar todo el comportamiento que gobierna la regla;
- no dejar instrucciones contradictorias ni duplicadas;
- no perder la intención del texto reemplazado;
- escribir para los modelos del perfil `execution`.

### T1 — Aviso de un proceso propio sin mensaje (G1)

- [x] `reporting.md` tiene el paso nuevo 2, antes del paso de la pregunta en espera, con la renumeración de design.md, y `global.md:16`, `reporting.md:63` y `site/src/content/docs/flows/build.md:39` lo reflejan.

**Closes:** AC1.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-close/references/reporting.md` ("Report a running or completed item", condición de `:63`), `content/guidance/global.md:16`, `site/src/content/docs/flows/build.md:39`.

**Execution:** hilo principal (ver arriba).

**Test approach:** check: búsqueda y lectura.

**Changes:** según design.md G1. El encabezado se conserva.

**Verification:**
- Lectura dirigida del paso 2 nuevo y de los tres puntos que enrutan avisos:
  - el paso 3, que redirige al paso 5;
  - la condición de `:63`;
  - `questions.md:60`.
- Con eso, un aviso de un proceso que el agente detuvo a propósito, con o sin una pregunta abierta, termina sin mensaje.
- El paso 5 sigue aplicando a los demás avisos.
- `rg -n "#report-a-running-or-completed-item" content` sigue mostrando los cinco enlaces a la misma ancla.
- `rg -n "step [0-9]" content/skills/flow-close/references/reporting.md` no devuelve referencias a pasos con la numeración vieja.

### T2 — Procesos de los hijos en la limpieza (G2)

- [x] El paso 4 de flow-close y su autorización en `global.md:111` incluyen los procesos que iniciaron los hijos de la tarea, y las copias quedan alineadas.

**Closes:** AC2.

**Depends on:** T1, porque ambas escriben `global.md`.

**Locations:** `content/skills/flow-close/SKILL.md:8`, `:25` y `:87`; `content/skills/flow-build/SKILL.md:67` y `:82`; `content/guidance/global.md:111`, `content/skills/workspace-conventions/references/artifact-layout.md:43`, `site/src/content/docs/flows/close.md:44,66`.

**Execution:** hilo principal.

**Test approach:** check: búsqueda y lectura.

**Changes:** según design.md G2. `:87` queda como puntero al paso 4.

**Verification:**
- `rg -n "processes it started|task-started|started by|processes the task" content site` no devuelve ninguna frase que excluya a los hijos.
- Lectura de `flow-build/SKILL.md:67`: al volver un hijo, el hilo principal anota todo proceso que dejó corriendo, a propósito o por descuido. Detiene enseguida los que ya no se necesitan, y el paso 4 de flow-close detiene el resto al cierre.
- `rg -n "children" content/skills/flow-close/SKILL.md content/guidance/global.md` muestra la regla nueva en el paso 4 y en `:111`.
- Lectura: los servicios previos y los de pertenencia incierta siguen preservados.

### T3 — Runtime real en la inspección del implementador (G3)

- [x] "The implementer's inspection" exige al menos un chequeo acotado sobre la aplicación en ejecución (servidor de desarrollo o build), en un navegador real y sin dobles de prueba en las rutas que el cambio usa.

**Closes:** AC3.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-walkthrough/references/ui-gate.md:15`, `content/agents/development/hive-build-frontend.md` (punto 4, puntero).

**Execution:** hilo principal.

**Test approach:** check: búsqueda y lectura.

**Changes:** según design.md G3. La primera vista y los cambios de solo texto o estilo quedan fuera.

**Verification:**
- `rg -n "test doubles" content/skills/flow-walkthrough/references/ui-gate.md` encuentra la regla en la sección, y `rg -n "ui-gate.md" content/agents/development/hive-build-frontend.md` encuentra el puntero en el rol, sin copia de la regla.
- Lectura: la regla acepta el servidor de desarrollo o un build, y no exige un build por tarea.
- Lectura: el chequeo sigue siendo acotado, y el recorrido sigue perteneciendo a los hijos.

### T4 — Candidato fijo en el gate (G4)

- [x] El brief de cada hijo del gate nombra la identidad del candidato, el candidato se sirve desde un build fijo, y una corrección durante el gate produce un candidato nuevo sin cambiar el de los hijos en curso.

**Closes:** AC4.

**Depends on:** T3, porque ambas escriben `ui-gate.md`.

**Locations:** `content/skills/flow-walkthrough/references/ui-gate.md` ("The independent children", "When the gate runs"), `content/skills/flow-walkthrough/SKILL.md:38`, `content/skills/flow-plan/SKILL.md:48`, `content/skills/flow-plan/references/plan-format.md:121`.

**Execution:** hilo principal.

**Test approach:** check: búsqueda y lectura.

**Changes:** según design.md G4.

**Verification:**
- `rg -n "candidate identity|identity of the candidate" content/skills` aparece en `ui-gate.md`, `flow-plan/SKILL.md` y `plan-format.md`.
- `rg -n "write-tree" content/skills/flow-walkthrough/references/ui-gate.md` encuentra el comando de identidad.
- El comando ya se probó en un repositorio temporal: da un hash de árbol que incluye los archivos sin seguimiento, no toca el índice real, y `git diff <t1> <t2>` muestra el diff.
- Lectura: la re-revisión de `ui-gate.md:26` usa ese diff, y `flow-walkthrough/SKILL.md:38` marca como `stale` las grabaciones tras una corrección dentro del gate.

### T5 — Pregunta de validación del recorrido (G5)

- [x] La primera vista y la validación final ofrecen abrir (`hive view open`) o aprobar, y piden las correcciones en el campo libre. Sin campo libre, la pregunta va en texto. Los fallos de `open` están cubiertos.

**Closes:** AC5.

**Depends on:** ninguna.

**Locations:**
- `content/skills/flow-walkthrough/SKILL.md:24`, `:42`, `:44` y `:46-48`;
- `content/skills/flow-build/SKILL.md`, en el paso de entrega del recorrido;
- `site/src/content/docs/flows/build.md:75`;
- `_support/docs/architecture/deployment-manager.md:124`.

**Execution:** hilo principal.

**Test approach:** check: búsqueda y lectura.

**Changes:** según design.md G5, con las excepciones locales a `questions.md` nombradas en el punto de uso.

**Verification:**
- `rg -n -i "never open|never opens" content site CHANGELOG.md _support/docs/architecture` no devuelve ninguna prohibición absoluta de abrir el visor.
- Lectura de "Give the opening command":
  - las opciones Abrir el visor y Aprobar, con el efecto de Aprobar según la parada;
  - las correcciones en el campo libre;
  - los dos casos en texto;
  - las excepciones `:44`, `:46`, `:62` y `:68` nombradas;
  - los dos fallos de `open`, sin ejecutar comandos sugeridos y con `hive view start` prohibido.
- Lectura: `:24` y `:42` apuntan a esa sección.

### T7 — Confirmar una selección delegada (G6)

- [x] `questions.md`, condición "an answer arrives", exige listar y confirmar la selección que una respuesta deja al criterio del agente, y `global.md:20` y `site/src/content/docs/flows/research.md:51` lo reflejan.

**Closes:** AC6.

**Depends on:** ninguna. T1 y T2 ya están verificadas; sus líneas de `global.md` son otras.

**Locations:** `content/skills/flow-plan/references/questions.md`, `content/guidance/global.md:20`, `site/src/content/docs/flows/research.md:51`.

**Execution:** hilo principal.

**Test approach:** check: búsqueda y lectura.

**Changes:** según design.md G6.

**Verification:**
- `rg -n "what stays out" content/skills/flow-plan/references/questions.md content/guidance/global.md` encuentra la regla y la cláusula; en la base no hay ninguna coincidencia.
- Lectura de la regla:
  - lista lo que entra y lo que queda fuera, con glosa, antes de actuar;
  - la confirmación es una pregunta nueva, y va en la siguiente ronda si hay una;
  - elegir una opción concreta no la dispara;
  - quedan exceptuadas la ruta iterativa y el modo desatendido.

### T8 — Navegador asignado al implementador de UI delegado (H1)

- [x] "The implementer's inspection" exige que, cuando un hijo implementa el cambio, su brief le asigne el navegador para el chequeo de runtime real.

**Closes:** AC7.

**Depends on:** T3.

**Locations:** `content/skills/flow-walkthrough/references/ui-gate.md:15`.

**Execution:** hilo principal.

**Test approach:** check: búsqueda y lectura.

**Verification:** `rg -n "brief assigns the browser tool" content/skills/flow-walkthrough/references/ui-gate.md` encuentra la regla. En la base no hay ninguna coincidencia. Lectura: la regla remite a `browser-automation.md` para elegir el navegador, sin repetir sus criterios.

### T6 — Entrada del CHANGELOG

- [x] `CHANGELOG.md`, bajo `[Unreleased]` > Changed, tiene una entrada para #93 que resume G1–G5, con las sesiones de origen. La entrada vieja de `:82-84` ("never opens a browser, the page or the viewer by itself") refleja la opción "Abrir el visor".

**Depends on:** T1–T5 y T7.

**Locations:** `CHANGELOG.md`.

**Execution:** hilo principal.

**Test approach:** ninguno: es un cambio de documentación.

**Verification:** `rg -n "#93" CHANGELOG.md` encuentra la entrada, y `rg -n -i "never opens" CHANGELOG.md` no devuelve una prohibición absoluta de abrir el visor.

## Verificación compartida

| Gate | Cuándo y dónde | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite local | Candidato final, en el checkout de trabajo | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` sin fallos | Parte de la implementación |
| Sitio | Si T1, T2 o T5 tocan `site/` | Desde `site/`: `pnpm install --frozen-lockfile`, `pnpm check`, `pnpm lint` y `pnpm build` sin errores | Parte de la implementación |
| Tamaño de `global.md` | Tras T1, T2 y T7 | `go test -run TestGlobalGuidanceSize -v ./tests/content`; el delta frente a 31 201 bytes se informa en el PR | Parte de la implementación |
| Coherencia | Al final | Lectura de cada regla tocada contra sus punteros. El PR lista los lugares revisados (el mapa de design.md) | Parte de la implementación |

No hay efecto de UI renderizado: solo cambia texto de guía, así que no aplican el gate de UI ni `UI verification`. Los pilotos siguen pausados. No se agregan casos en `tests/fixtures/regression/`.

## Revisión y progreso

- 2026-10-19: borrador con G1–G5.
- 2026-10-19: revisión del plan en `05a5d16cdd7d` con dos `hive-review-plan`. G1–G2: 1 hallazgo bloqueante y 4 no bloqueantes (~89 s, 123k tokens). G3–G5: 1 bloqueante y 7 no bloqueantes (~118 s, 114k tokens). Todos se aceptaron y se corrigieron en design.md y tasks.md:
  - renumeración de pasos de reporting.md;
  - la detención a propósito no es fallo;
  - pertenencia de los procesos de los hijos en `flow-build:67`;
  - runtime real sin exigir build;
  - puntero en `hive-build-frontend`;
  - identidad por `git write-tree` con índice temporal;
  - un puerto y un directorio por candidato;
  - grabaciones `stale` en el gate;
  - tarjeta con dos opciones o texto;
  - efecto de Aprobar según la parada;
  - excepciones `:44`, `:46`, `:62` y `:68`;
  - no ejecutar comandos sugeridos.

  Límite aceptado: `agent-delivery.md` no documenta cómo trata cada host un turno sin texto. Queda como supuesto, con el paso 5 de respaldo. - 2026-10-19: re-revisión en `c9e9afb2200b` (~28 s y ~33 s). G1–G2: un bloqueante, que G2 solo anotaba los procesos dejados a propósito, y un no bloqueante, la casilla de T1. G3–G5: tres no bloqueantes: AC3 y T3 exigían un build, la línea correcta de `deployment-manager` es `:124`, y la identidad usa el commit solo con el árbol limpio. Todos se aplicaron tal como los propusieron los revisores, sin otra ronda, según `plan-review.md`. Sin hallazgos abiertos. Límite: cómo trata cada host un turno sin texto queda como supuesto, con el paso 5 de respaldo. D1 (qué hacer con los hallazgos), D2-A (correcciones en el campo libre) y D3-A (aplica en ambas paradas) registradas en proposal.md. Pendiente: revisión del plan y decisión de entrega.
- 2026-10-19: revisión durante el build. La entrada del CHANGELOG pasa a la tarea T6, porque `CHANGELOG.md` estaba en las Locations de las cinco tareas y eso impedía verificarlas en paralelo. Ningún criterio cambia. El borrador de la entrada, escrito con T1, queda bajo T6.
- 2026-10-19: revisión durante el build, por D7-A y D8-A. Se agregan AC6 y T7 (G6, confirmar una selección delegada), y T6 pasa a depender también de T7. Ningún criterio existente cambia. Queda pendiente una ronda de re-revisión para G6.
- 2026-10-19: re-revisión de G6 en `9fe9f2add43e` (~44 s): un bloqueante, la ruta iterativa, y tres no bloqueantes: modo desatendido, confirmación dentro de la siguiente ronda y verificación de T7. Todos se aplicaron tal como los propuso el revisor, sin otra ronda.
- 2026-10-19: gates compartidos sobre el candidato final.
  - La suite local pasó: `go vet ./...`, 20 paquetes `ok` en `go test ./...` y 55 pruebas en `tests/skills`, todo con salida 0.
  - El sitio pasó `pnpm check` (0 errores, 0 avisos), `lint` y `build` (10 páginas). Quedaron tres avisos de Vite y del contenido del sitio que no vienen de este cambio.
  - `global.md` pasó de 31 201 a 31 550 bytes (+349).
  - Modelos: el implementador fue el hilo principal, Claude Opus 5.5 (`claude-opus-5-5`), observado. Los verificadores `hive-verify-task` se lanzaron con `model: opus`, configurado y no observado.
- 2026-10-19: revisión al cierre, por D10-A. Se agregan AC7 y T8, el hallazgo H1 surgido en la verificación de T3. Es una sola línea en la casa canónica de G3, y no la re-reviso porque aplica la corrección propuesta.
- 2026-10-19: T8 verificada. Hay ocho commits en la rama de prueba (`d96e4021`…`83c2073a`), uno por tarea, armados por hunk. El despliegue local se hizo con `hive update --out` y `hive apply --plan` desde `83c2073a` a los seis hosts.

## Continuación y ajustes posteriores del 2026-10-23

- La auditoría de ocho sesiones principales (seis con trabajo sustantivo) se entregó en la conversación. La última Claude de Facti ejercitó runtime local real, identidad del candidato, navegador asignado y aprobación; no ejercitó todos los caminos de la guía. No se afirma una mejora general de fiabilidad.
- La suite de Go, `go vet`, 55 pruebas de skills y check/lint/build del sitio pasaron sobre `5fa7d308` antes de los ajustes posteriores. Los tres avisos del build del sitio coinciden con los ya documentados; no hubo errores.
- Ajustes P1–P3 aprobados por D1-A: verificación tras cambios posteriores a aprobación, mensaje con límite de espera y orden ejecutable de limpieza. Son correcciones posteriores a las tareas ya verificadas; la verificación final de estos ajustes queda pendiente antes de sus commits.
- Ejecución delegada sin historial: tres workers, con propiedad separada para candidato, espera y limpieza; sin commits ni efectos externos. El coordinador integra y verifica. No hubo revisión dedicada ni pilotos.
- D2 autoriza el PR posterior de cierre y su merge. El material previo en `_support/sessions/` queda fuera de esta entrega.

- Verificación final de P1–P3: conforme por un `hive-verify-change` nuevo, sin historial; no es una revisión de código dedicada. `go test ./...`, `go vet ./...`, 55 pruebas Python, `pnpm check`, `pnpm lint`, `pnpm build` (10 páginas) y `git diff --check` pasaron. Comprobó seis enlaces y contenido HTML generado. El worker de limpieza probó el orden en un fixture Git desechable y lo retiró. No se ejecutaron modelos ni pilotos; no se afirma cumplimiento conductual universal.
- Commits posteriores: `65c32d7` (mensaje de espera), `18ce037` (orden de limpieza) y `9e89f46` (evidencia posterior a aprobación). La rama contiene ahora 19 commits sobre la base. Guía global: 31 201 → 32 693 bytes (+1 492 contra `development`, sin crecimiento por estos tres ajustes).

## Cierre del 2026-10-23

- Integrado por PR #96 (ajustes de validación y entrega), commit `ac146beb`. Verificación final: Go, vet, 55 pruebas Python, check/lint/build del sitio y diff sin errores; no hubo CI ni pilotos de modelos.
- Se conserva el riesgo de cumplimiento de modelos y de precedencia de memoria; la auditoría no certifica fiabilidad general. Los planes se archivan por D2 en el PR posterior de cierre.
- El contenido instalado corresponde a `ac146beb` en seis hosts; el visor cotidiano y el material previo no se tocaron.
