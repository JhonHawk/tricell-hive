# Entrega y revisión declaradas, y chequeos del ciclo de vida

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · fusionado en `development` por el [PR #68](https://github.com/JhonHawk/tricell-hive/pull/68) (`6a278a18`, 2026-10-18) por decisión del usuario, con la entrada de changelog añadida antes del merge |
| Tracker · GitHub Issues | Sin issue (D2-A): nace de la comparación con el artículo de Anthropic sobre revisión de código, registrada en Engram |
| Git | C1-A: rama de prueba nueva desde `development`, commit por tarea, despliegue local · sin push ni merge · tú pruebas en sesiones reales y decides el merge |
| Verificación | comprobaciones `rg` por criterio · `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` · tamaño de `global.md` |
| Siguiente paso | Ninguno: registro archivado al cierre. La prueba en sesiones reales de R1 y R2 (un repo con `Delivery` y `Review` declarados) queda como observación posterior al merge |

## Petición

> Me interesa que revisemos este ciclo de vida, comparandolo con hive y mejoras que podriamos tener para agilizar ciertos trabajos o mejorar la calidad de los resultados [Image #1] https://claude.dev/blog/solving-code-review-at-anthropic/ previamente lo revisamos, pero no sé si le dimos el enfoque correcto

> tambien toma en cuenta que nosotros no tenemos los recursos de anthropic :v

Respuestas del 2026-10-17: D1-A (plan con R1–R6, cada candidato como decisión del plan), D2-A (sin issue), R1-A, R2-A, R3–R6 (los cuatro), C1-A.

## Objetivo

La comparación del 2026-10-16 miró solo al revisor. La segunda lectura, capa por capa, muestra que Hive cubre las decisiones de todas las capas del artículo, pero reparte la atención humana al revés: el artículo clasifica el riesgo a partir del cambio y pide a una persona solo cuando el nivel lo exige; Hive pregunta al usuario el modo de entrega y el mecanismo de revisión antes de la primera edición en cada sesión y después automatiza. La tesis del artículo es que el recurso escaso es el tiempo humano por cambio; en Hive ese costo son las preguntas y las esperas por trabajo.

El cambio adopta lo que cabe en una guía, sin infraestructura (sin cola de merge, clasificadores, evals ni barridos nocturnos):

- **Agilizar.** R1: un valor `Delivery: automatic` o `interactive` declarado en el `## Hive` de un repo de código es el modo, dicho en una línea antes de editar, no preguntado. R2: un valor `Review` declarado (mecanismo o `none`) es el mecanismo, corrido como chequeo rutinario, no preguntado. Un repo con ambos valores arranca sin preguntas; hoy solo `Delivery: direct-base` lo hace, y solo en repos de soporte.
- **Calidad.** R3: antes de un merge, los checks deben haber corrido sobre la base actual; si la base avanzó, se actualiza la rama y se espera CI de nuevo (sustituto de la cola de merge). R4: `hive-review-code` reporta un defecto repetido como un solo hallazgo con todas sus ubicaciones. R5: el postmortem de la ruta urgente nombra la capa que debió atrapar el fallo y pone ahí la acción. R6: `hive-review-code` dice en su veredicto si el diff toca una frontera de seguridad, como segundo clasificador que dispara `hive-review-security`.

## Alcance y aceptación

Incluye:
- `content/skills/flow-plan/references/delivery-decisions.md` (R1, R2, R3 puntero, R5).
- `content/skills/flow-plan/references/project-settings.md` (valores nuevos y ejemplo).
- `content/guidance/global.md:66` (una declaración del repo autoriza el mecanismo y el modo que nombra).
- `content/skills/flow-build/SKILL.md:20`, `content/skills/flow-build/references/verification.md:33,43`, `content/skills/flow-plan/SKILL.md:18` y `content/skills/flow-close/SKILL.md:67` (punteros).
- `content/skills/git-workflow/SKILL.md` (R3, hogar de la regla).
- `content/agents/review/hive-review-code.md` (R4, R6).
- `tooling/cli/doctor_project.go`, `tooling/cli/tui_project_view.go` y sus pruebas: `hive doctor` y el formulario de proyecto aceptan `automatic` e `interactive`.
- `_support/docs/architecture/deployment-manager.md:77,110` y `site/src/content/docs/flows/build.md:41,64` (describen el comportamiento nuevo).

Queda fuera:
- **Rutas que siempre exigen persona** (lista de paths): la regla de UI que marca `automatic` no disponible ya cumple ese papel; una lista de paths se añadiría si dos sesiones reales la piden.
- **Las vigilancias A–D del 2026-10-16** (disposición de hallazgos, tasa de reverts, REVIEW.md en estándares, evals): siguen con conteo 0.
- **Cola de merge, flags, barridos nocturnos, votación de verificadores, clasificador**: necesitan infraestructura que Hive no provee.
- **README.md y llms.txt**: no cambia ningún comando ni flag, y ninguno de los dos menciona `Delivery` ni `direct-base`.

Restricciones:
- La suite local pasa: `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`.
- El texto distribuido se escribe para los modelos del perfil `execution`.
- Cada regla tiene un solo hogar; los demás sitios son punteros (AGENTS.md, "Guidance structure").
- Los commits citan las decisiones R1-A, R2-A, R3–R6 y C1-A; el tamaño de `global.md` se declara en el registro (base: 29 638 bytes).

- AC1. `delivery-decisions.md` dice que un valor `Delivery: automatic` o `Delivery: interactive` declarado en el `## Hive` del repo es el modo para el trabajo ahí y autoriza sus efectos: no se pregunta; una respuesta anterior de la sesión o una instrucción más estrecha del pedido lo anulan; antes de la primera edición el agente dice en una línea el modo y dónde se detiene, el mecanismo de revisión, el despliegue que dispare el destino y la ruta de cierre del registro (un PR de cierre sin revisión que el agente mergea); la regla que marca `automatic` no disponible para un cambio con validación humana sigue aplicando y entonces el trabajo va en `interactive` nombrando esa regla; la pregunta se hace solo cuando no hay valor declarado, también con varios repos (`:18,20`); `flow-plan/SKILL.md:18` y `flow-close/SKILL.md:67` remiten a ello. *Falso en la base cuando* `rg -c -F 'Delivery: automatic' content/skills/flow-plan/references/delivery-decisions.md` → 0.
- AC2. `delivery-decisions.md` dice que un valor `Review` declarado es la respuesta a la pregunta de revisión y autoriza el mecanismo: corre como chequeo rutinario del candidato en el punto que el modo fija, o ninguna revisión de código dedicada cuando es `none`, tras confirmar sin lanzarlo que el mecanismo existe en este host; un `Review` declarado, incluido `none`, autoriza también `hive-review-security` cuando el cambio cruza una frontera de `security-boundaries.md` (N4-A); `global.md:66` dice que un valor `Review` o `Delivery` declarado sí es autorización, remitiendo a `delivery-decisions.md`; `flow-build/SKILL.md:20` y `verification.md:33` remiten a ello. *Falso en la base cuando* `rg -c -F 'is the answer' content/skills/flow-plan/references/delivery-decisions.md` → 0 y `rg -c -F 'declared in the repository' content/guidance/global.md` → 0.
- AC3. `project-settings.md` lista `Review` (mecanismo o `none`) y `Delivery` (`direct-base` para soporte; `automatic` o `interactive` para código) y el ejemplo muestra `Delivery`. *Falso en la base cuando* `rg -c -F '`automatic`' content/skills/flow-plan/references/project-settings.md` → 0.
- AC4. `hive doctor` y el formulario de proyecto aceptan `Delivery: automatic` y `Delivery: interactive`, rechazan otros valores con un mensaje que lista los tres válidos, y el formulario cicla `none`, `direct-base`, `automatic`, `interactive` sin duplicar un valor ya presente en el archivo; `deployment-manager.md:77,110` lo dicen. *Falso en la base cuando* `go test ./tooling/cli/ -run 'TestDoctorProjectAcceptsDeliveryModes'` no encuentra la prueba, y la sección con `Delivery: automatic` produce `expected "direct-base", found "automatic"` (comprobado: `doctor_project.go:355`).
- AC5. `git-workflow/SKILL.md` exige, antes de un merge, que los checks que pasaron hayan corrido con la base actual, con una comprobación de solo lectura nombrada: si la base recibió commits después de la última corrida que pasó, actualizar la rama con un merge desde la base, esperar la corrida nueva y mergear solo si pasa, repitiendo una vez si la base vuelve a moverse; un rebase exigido por el repo necesita la autorización del force-push y, sin quien responda, deja el merge pendiente; con cola de merge, encolar y esperar su resultado; un verde sobre base vieja no cumple la condición. `delivery-decisions.md:24` remite a ello. *Falso en la base cuando* `rg -c -F 'stale base' content/skills/git-workflow/SKILL.md` → 0.
- AC6. `hive-review-code.md` dice que un mismo defecto en varios lugares se reporta una vez, nombrando la clase y cada ubicación. *Falso en la base cuando* `rg -c -F 'every location' content/agents/review/hive-review-code.md` → 0.
- AC7. `hive-review-code.md` termina su veredicto diciendo si el diff toca una frontera de `security-boundaries.md`, nombrándola, y `verification.md:43` incluye la frontera que el veredicto nombra entre las que añaden `hive-review-security`. *Falso en la base cuando* `rg -c -F 'security-boundaries.md' content/agents/review/hive-review-code.md` → 0.
- AC8. El postmortem de `delivery-decisions.md` nombra la capa que debió atrapar el fallo (un check, una prueba, un estándar, un criterio de revisión o una comprobación de despliegue) y pone ahí la acción preventiva con su dueño. *Falso en la base cuando* `rg -c -F 'should have caught' content/skills/flow-plan/references/delivery-decisions.md` → 0.
- AC9. `site/src/content/docs/flows/build.md` dice que con `Delivery` y `Review` declarados el agente los dice en una línea en lugar de preguntar, que puedes declararlos una vez en el `## Hive`, y que antes de mergear comprueba que los checks corrieron sobre la base actual. *Falso en la base cuando* `rg -c -F '## Hive' site/src/content/docs/flows/build.md` → 0.

## Entrega

- **Modo:** C1-A, elegido el 2026-10-17 para este repositorio: rama de prueba nueva desde `development` (propuesta: `feat/lifecycle-declared-delivery-review`), un commit por tarea cuando llega a `[x]`, despliegue local con `hive update` y `hive apply` (rebuild del binario primero, porque T2 cambia el manager). Sin push ni merge; tú pruebas en sesiones nuevas y decides el merge.
- **Revisión de código:** ninguna dedicada; el cambio de Go lo cubren sus pruebas y el resto es texto.
- **Registro:** esta carpeta queda sin versionar hasta el merge; `flow-close` la versiona por la ruta de cierre que elijas entonces.
