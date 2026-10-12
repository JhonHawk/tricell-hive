# Tareas: ruta iterativa (R4)

Base del cambio: `c332a814` (`feat/guidance-rework`, 2026-10-11) más R1–R3 entonces sin commit. Implementación integrada por [PR #46](https://github.com/JhonHawk/tricell-hive/pull/46) en `16028d3575e7286cc028985cd2138c3268257add`; registro archivado. La entrega inicial sin commits que sigue debajo es histórica y fue sustituida por la entrega final de `proposal.md`.

## Entrega y verificación iniciales (historial)

- Modo `hold` en `feat/guidance-rework`: árbol de trabajo verificado, sin commit, push ni PR hasta que el usuario los pida. Sin revisión dedicada: cambio solo de texto de la guía (convención del usuario, rama de prueba).
- Verificación por tarea: `hive-verify-task` según `flow-build`, comparando modelos antes de lanzarlo.
- Puerta compartida: `go test ./...` una vez sobre el candidato final; informa el tamaño de `global.md`.
- Medición en sesiones reales: solo tras un despliegue que el usuario pida.

## Tareas

### T1 — Ruta iterativa en `flow-research`

- [x] Párrafo nuevo tras `flow-research/SKILL.md:36` con el texto de [design.md](design.md#texto-propuesto): definición de vuelta, opción, límites, paradas en lista y cierre al parar.

**Closes:** AC1, AC2.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-research/SKILL.md:36`.

**Execution:** hilo principal; la redacción está acoplada con T2 y T4, y el brief sería más largo que el cambio.

**Test approach:** check.

**Verification:** `rg -n -i "iterative route" content/skills/flow-research/SKILL.md` y `rg -n -i "round limit|without a measurable improvement" content/skills/flow-research/SKILL.md` encuentran el párrafo nuevo. Leerlo: vuelta completa, objetivo y medición, límite de efectos de `global.md`, tope propuesto, las siete paradas en lista y `flow-close` o pregunta de ruta al parar.

### T2 — `global.md` reconoce la ruta

- [x] Las líneas 31, 34 y 42 de `content/guidance/global.md` aplican el texto de [design.md](design.md#texto-propuesto).

**Closes:** AC3 (parte global), AC4.

**Depends on:** T1.

**Locations:** `content/guidance/global.md:31`, `:34`, `:42`.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check.

**Verification:** `rg -n -i "iterative route" content/guidance/global.md` encuentra las líneas 31, 34 y 42. Leer la 31: la lista de efectos excluidos sigue entera y la ruta es la excepción nombrada a la elección de flujo. Leer la 42: la excepción se limita a la ruta elegida y sus límites, sin «only».

### T4 — `flow-build` no cierra tras cada vuelta

- [x] `content/skills/flow-build/SKILL.md:75` añade la frase de [design.md](design.md#texto-propuesto), y `site/src/content/docs/flows/build.md:31` lo describe.

**Closes:** AC3 (parte de `flow-build`), AC5 (parte de `build.md`).

**Depends on:** T1.

**Locations:** `content/skills/flow-build/SKILL.md:75`, `site/src/content/docs/flows/build.md:31`.

**Execution:** hilo principal; una frase acoplada con T1.

**Test approach:** check.

**Verification:** `rg -n -i "iterative route" content/skills/flow-build/SKILL.md` encuentra la línea 75; `rg -n -i "iterative route" site/src/content/docs/flows/build.md` encuentra la línea 31.

### T3 — Sitio de investigación y CHANGELOG

- [x] `site/src/content/docs/flows/research.md` describe la ruta iterativa; `CHANGELOG.md` registra R1–R4 bajo `[Unreleased]`.

**Closes:** AC5 (partes de `research.md` y `CHANGELOG.md`).

**Depends on:** T1, T2, T4.

**Locations:** `site/src/content/docs/flows/research.md` (sección «What it produces»), `CHANGELOG.md` (`[Unreleased]`: R4 en «Added», R1–R3 en «Changed»).

**Execution:** hilo principal; texto corto que depende de la redacción final.

**Test approach:** check.

**Verification:** `rg -n -i "iterative route"` sobre `site/src/content/docs/flows/research.md` y, por separado, sobre `CHANGELOG.md` encuentra cada uno. Leer la entrada de `CHANGELOG.md`: nombra también R1 (arreglos de revisión dentro de la implementación), R2 (pregunta de revisión sin entrega por Git) y R3 (sin otra ronda tras la re-revisión). El sitio describe comportamiento y enlaza al skill sin copiar la regla.

## Revisión del plan y progreso

Revisado el 2026-10-11 por `hive-review-plan` (dominio: guía distribuida), sin hallazgos bloqueantes. Aplicados F1 (T4, `flow-build` no cierra tras cada vuelta), F2 (vuelta completa), F3 (límite de efectos de `global.md`), F4 (un permiso vigente cubre el commit), F5 (vuelta pequeña y parada por `flow-plan`), F6 (sin «only»), F7 (cierre o pregunta de ruta al parar), F8 (comprobación de AC5 por archivo), N1 (párrafo propio con lista) y N2 (contenido del CHANGELOG). N3 aceptado sin cambio (ver design.md). Las correcciones aplican propuestas del revisor tal cual, sin nueva revisión. Construcción iniciada el 2026-10-11. Modelos: implementador = hilo principal, `claude-opus-5-5` (observado); `hive-verify-task` configurado `opus` (configurado, no observado), igual al implementador, así que se lanza con `sonnet` por la opción de modelo del host. T1–T4 verificadas (`[x]`), todos los criterios `met`. Tras la verificación de T3 se añadió la parada «necesita un plan» al CHANGELOG y al sitio, como señaló el verificador. Puerta compartida: `go test ./...` pasó; `global.md` mide 46 912 bytes (46 561 antes de R1–R4).
