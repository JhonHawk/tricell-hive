# Tareas

### T1 — Condensar y fusionar en `global.md` (C1, B1–B3 en global, V1)

- [x] `global.md` tiene L1–L3, M1, M2, M5, S1–S8, P1–P8, el bullet B1, la línea `:124` de B2, B3 y V1 según [design.md](design.md). M4 y S9 quedan sin cambio.

**Depende de:** nada.
**Ubicación:** `content/guidance/global.md`.
**Ejecución:** hilo principal. Un solo archivo con un solo escritor, y el texto exacto ya está en el diseño: un brief sería tan largo como el cambio.
**Enfoque de prueba:** `check`. Es texto de guía; no hay comportamiento ejecutable que probar con TDD.
**Verificación:** `rg -n 'Use clear prose|Reuse authorization while|otherwise from the workspace|A declared .Review.' content/guidance/global.md` no devuelve nada; `rg -n 'shared dispatch|For consequential delegated work' content/guidance/global.md` sigue encontrando `:90`; `wc -c content/guidance/global.md` se registra antes y después.

### T2 — Recursos que reciben lo movido (B1, B2, B3, V2)

- [x] Nuevo `content/skills/flow-research/references/backlog-report.md` con el resto de `:21`, y el enlace con su condición en `flow-research/SKILL.md` (B1).
- [x] La frase de `project.md` en la sección "Files" de `content/skills/flow-plan/references/change-records.md` (B2).
- [x] "A `Review` setting in the repository's `## Hive` section is an explicit preference." en `content/skills/flow-plan/references/delivery-decisions.md:26` (B3).
- [x] `content/skills/flow-build/references/verification.md:67` remite a la regla global de imágenes (V2).

**Depende de:** T1, por el texto que sale de `global.md`.
**Ejecución:** hilo principal; son cuatro ediciones de pocas líneas, acopladas al texto de T1.
**Enfoque de prueba:** `check`. `TestRepositoryCatalogueInstructionReferences` (`tooling/management/references_test.go:60`) valida el enlace de `flow-research/SKILL.md`. El puntero `skill:` de `global.md` queda fuera de esa validación (`references.go:17-20`).
**Verificación:**
- `rg -n 'backlog-report.md' content/skills/flow-research/SKILL.md content/guidance/global.md` muestra el enlace y la instrucción de lectura.
- `rg -n 'session folder' content/skills/flow-build/references/verification.md` no devuelve la ubicación vieja.
- La validación de enlaces corre en T3.

### T5 — Test de punteros `skill:` en `global.md` (K5-A)

- [x] Un test en `tests/content/` extrae de `content/guidance/global.md` cada destino `skill:<owner>/<path>` y comprueba que exista `content/skills/<owner>/<path>`.

**Depende de:** nada; se escribe antes que B1 en T1 y T2.
**Ejecución:** hilo principal; es un test corto junto a `budget_test.go`.
**Enfoque de prueba:** `tdd`. Primero el test, con el puntero de B1 ya escrito en `global.md` y la referencia todavía sin crear: debe fallar nombrando `flow-research/references/backlog-report.md`. Después se crea la referencia y pasa.
**Verificación:** `go test -count=1 ./tests/content/...` pasa por rojo y luego por verde; además, renombrar a mano la referencia en el worktree de T3 lo hace fallar (se revierte después).

### T3 — Presupuesto y verificación conjunta

- [x] `globalGuidanceBudget` en `tests/content/budget_test.go` igual al nuevo `wc -c` de `global.md`.
- [x] En un worktree en `e7906a4` dentro del scratchpad, con solo el diff de las rutas de este cambio aplicado (`git diff -- content/ tests/content/ | git -C <wt> apply`), pasan:
  - `go vet ./...`;
  - `go test -race ./...`, incluida `TestRepositoryCatalogueInstructionReferences`;
  - `python3 -m unittest discover -s tests/skills -p '*_test.py'`.
  
  Así la verificación no depende del commit ni del trabajo ajeno en `tooling/**`.

**Depende de:** T1, T2.
**Ejecución:** hilo principal.
**Verificación:** los tres comandos en verde; la reducción en bytes queda registrada en `proposal.md` (A4); el worktree se elimina al terminar.

### T4 — Revisión independiente de equivalencia (K4-A)

- [x] Un `review-harness` compara cada regla vieja con la nueva, sobre el diff de T1–T2, buscando restricciones perdidas, cambios de disparador o de alcance, y reglas de autorización, secretos, preservación o ubicación que hayan salido de global. Los hallazgos bloqueantes pasan por `review-refuter`.

**Depende de:** T1–T3.
**Ejecución:** delegado, solo lectura. El `review-harness` está pensado para auditar archivos de instrucciones; el hilo principal no puede revisar su propia reescritura de forma independiente.
**Verificación:** cada hallazgo queda corregido, o refutado con evidencia, antes de la entrega. Se revisan en especial L3, S1, S3, P6 y B1.

## Verificación y revisión humana

| Gate | Momento | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Tests, vet, enlaces | T3 | Worktree en `e7906a4` con el diff aplicado | Incluido en la implementación |
| Equivalencia (K4-A) | T4 | `review-harness` con el diff; bloqueantes pasan por `review-refuter` | Decidido por el usuario (K4-A) |
| Revisión de código | Antes de entregar | Mecanismo de la pregunta de entrega | Pendiente |
| Despliegue | Después de integrar | Manager de Hive, release nueva | Pendiente de la pregunta de entrega |

No hay superficie de UI ni prueba local en vivo aplicable: el cambio es texto de guía. Las corridas con modelos siguen pausadas; la mejora de atención no se afirma sin medirla (A4).

## Estado de revisión y progreso

**Revisión del plan (2026-09-26, revisión `5ef1616a6d6d`):** dos `review-plan` en paralelo, uno de semántica de la guía y otro de distribución y verificación.

**Bloqueante, aceptado:**
- **H1, en ambos dominios:** S9 dejaba sin carga "the shared dispatch evidence requirements" que cita `plan-review.md:26`. S9 sale del plan y `:90` queda sin cambio.

**Semántica, aceptados:**
- **H2:** M4 sin cambio, porque la condición "consequential" se extendía a todo el bullet.
- **H3:** M2 reducido a la frase de PR.
- **H4:** B1 conserva en global las prohibiciones del agrupamiento.
- **H5:** B2 va a la sección "Files".
- **H6:** P6 conserva "automatically" y "migrate".
- **H7:** L3 conserva la condición y la frase de separación.
- **H8:** S5 cita `:72`.
- **Menores:** "prefer" en S3, "to the user" en S1, y la preferencia explícita de `Review` en B3.

**Distribución, aceptados:**
- **H2:** localizador `skill:`; queda registrado que el puntero de `global.md` no se valida.
- **H3 y H4:** verificación en un worktree en `e7906a4` con el diff aplicado.
- **H5:** coincide con la H5 de semántica.
- **Propuesta abierta:** un test que compruebe que cada `skill:` citado en `global.md` existe. Es una adición y la decide el usuario.

**Sin segunda ronda:** todas las correcciones aplican lo que propusieron los propios revisores, o quitan un cambio en vez de introducir uno nuevo.

**Implementación (2026-09-26):**
- **T1:** hecho. `global.md` pasa de 41177 a 38119 bytes (−3058 B, −7,4%). Tres ajustes respecto del diseño:
  - el bullet B1 conserva "rather than by recent activity", que es el fallo de a143621 (Grok devolvió un backlog parcial);
  - S4 conserva "the orchestrator selects the relevant domains";
  - de Naming salen además "Change records use the fixed names of the change folder" y "add internal stages only when useful". El primero lo cubren `:121` y `change-records.md`; el segundo es una concesión opcional.
- **T5:** hecho con TDD. `TestGlobalGuidanceSkillLinksResolve` falló primero porque no existía `flow-research/references/backlog-report.md`, y pasó al crearla. En el worktree de T3, renombrar la referencia lo vuelve a hacer fallar.
- **T2:** hecho. `rg 'session folder'` en `verification.md` no devuelve nada.
- **T3:** hecho, en un worktree en `e7906a4` con solo los 8 archivos de este cambio copiados. Pasan:
  - `go vet ./...`;
  - `go test -race -count=1 ./...`, incluida `TestRepositoryCatalogueInstructionReferences`;
  - los tests de skills en Python.
- **T4:** hecho. El `review-harness` revisó las siete rutas contra `e7906a4` y confirmó todas las condensaciones menos siete (H1–H7), además de las diferencias con el diseño:
  - **H1, bloqueante:** se perdió el umbral de cinco tickets. `review-refuter` lo confirmó en parte: es un cambio de significado, no una contradicción estricta. Se restauró en global y en la referencia.
  - **H2–H7:** se aplicaron con el texto que propuso el revisor:
    - resolución de recursos para cualquier enlace, no solo `skill:`;
    - "describe current behavior";
    - definición del current-requirements home;
    - "Give it a brief";
    - disparador del enlace de PR;
    - "or record it as below".
  - **Además:** "Tool-defined filenames keep their names", la mejora opcional que el revisor sugirió para P6.
  - Los cambios de S4 y B1 respecto del diseño se aceptaron como correctos.
- **Medición final:** `global.md` pasa de 41177 a 38352 bytes (−2825 B, −6,9%). Nuevo `backlog-report.md`: 1,3 KB, que se carga solo con `flow-research`.
- **Verificación final:** worktree en `e7906a4` con los 8 archivos. `go vet ./...` y `go test -race -count=1 ./...` pasan, incluidos `tests/content` y `tooling/management`; los tests de skills también.
- **Sin segunda ronda:** las correcciones aplican lo que propuso el propio revisor.
