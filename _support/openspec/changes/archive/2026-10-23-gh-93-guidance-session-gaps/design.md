# Diseño

## Contexto verificado

El mapa de impacto se hizo el 2026-10-19 sobre `a6255b6f` (development). Abarca `content/`, `site/`, `_support/docs/`, `tests/`, `tooling/` e `integrations/`.

- **Pruebas:** ninguna afirma sobre el texto de estas reglas. `tests/content/reading_footprint_test.go` y `TestGlobalGuidanceSize` solo registran tamaños. `tooling/management/references_test.go` valida los enlaces entre archivos, pero no las anclas.
- **Tamaño de `content/guidance/global.md`:** 31 201 bytes, 4 785 palabras y 113 líneas. G1 y G2 lo tocan; G3, G4 y G5 no.
- **Suite local** (CONTRIBUTING.md): `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'`. Si se toca `site/`, además, desde `site/`: `pnpm install --frozen-lockfile`, `pnpm check`, `pnpm lint` y `pnpm build`.
- **Evidencia de cada vacío:** está en el [issue #93](https://github.com/JhonHawk/tricell-hive/issues/93) y en su comentario sobre G5. Las sesiones son Grok 1.0.64 `01a15706` y Codex 0.163.0 `01a156af`, ambas del 2026-10-19.
- **`hive view open`** (`tooling/cli/view_command.go:597-616`):
  - abre la raíz del visor, no la página del recorrido;
  - con el visor apagado imprime "The viewer is not running…" y sale con código 0;
  - sus errores sugieren comandos (`hive view stop`, `hive view start`).

## Diseño propuesto

### G1 — Aviso de un proceso que el propio agente detuvo

**Casa canónica:** `content/skills/flow-close/references/reporting.md`, sección "Report a running or completed item". El encabezado se conserva, así que los cinco enlaces a su ancla siguen válidos.

**Paso nuevo 2:** cuando el aviso informa la salida de un proceso que el agente detuvo a propósito, como al limpiar o al reiniciar un servidor, el turno no lleva mensaje y sigue en el paso 6. El código de salida que causó esa detención, como 143, no es un fallo.

**Renumeración:**

| Paso actual | Paso nuevo | Cambio |
| --- | --- | --- |
| 1 | 1 | Sin cambio |
| 2 (pregunta esperando) | 3 | Dirige al paso 5 |
| 3 | 4 | Sin cambio |
| 4 (una línea) | 5 | Sin cambio |
| 5 (continuar o terminar) | 6 | Sin cambio |

El paso 2 va antes del 3, así que un aviso de este tipo con una pregunta abierta tampoco lleva mensaje. Ese es justamente el escenario de Grok.

**Textos que cambian con la renumeración:**
- La condición de `reporting.md:63` cita "step 5 above" y exceptúa también el turno del paso 2.
- `global.md:16` recibe la excepción mínima, porque hoy dice que ese turno "is that note".
- `site/src/content/docs/flows/build.md:39` describe el caso.

**Supuesto sobre los hosts:** `agent-delivery.md` no documenta qué hace cada host con un turno que termina sin texto. El paso dice: "si el host exige una respuesta, aplica el paso 5". Se descartó evitar el aviso en origen, deteniendo el monitor antes de matar el proceso, porque depende del host.

### G2 — Procesos que iniciaron los hijos

**Casa canónica:** `content/skills/flow-close/SKILL.md`, paso 4. "The processes it started" pasa a incluir los que iniciaron sus hijos.

**Cómo se sabe qué es de la tarea:** `flow-build/SKILL.md:67` ya manda comprobar, cuando vuelve un hijo, que no dejó procesos corriendo. Se amplía así:
- todo proceso que el hijo deja corriendo se anota con su PID o puerto en el progreso del plan, o en el reporte si no hay plan, lo haya dejado a propósito o por descuido, como los servidores de un verificador;
- los que ya no se necesitan se detienen enseguida;
- los demás cuentan como de la tarea y se detienen al cierre.

Siguen preservados los servicios que ya existían y los de pertenencia incierta. Un proceso anotado no es de pertenencia incierta.

**Autorización:** viene de `global.md:111`, que recibe el mismo ajuste. Quien detiene es el hilo principal dentro de su tarea, así que no hay autoridad nueva para el hijo.

**Textos alineados:**
- `flow-close/SKILL.md:8` ("task-started processes") y `:87`, que pasa a ser puntero al paso 4.
- `flow-build/SKILL.md:82` ("stop only task-started temporary processes").
- `artifact-layout.md:43`.
- `site/src/content/docs/flows/close.md:44` y `:66`.

### G3 — La inspección del implementador ejerce el runtime real

**Casa canónica:** `ui-gate.md`, sección "The implementer's inspection". Antes de informar la tarea, al menos uno de sus chequeos acotados carga la vista cambiada en un navegador real, contra la aplicación en ejecución. Puede ser el servidor de desarrollo o un build. Las rutas que usa el cambio (red, almacenamiento, APIs nativas del navegador) se ejercen sin dobles de prueba, y el chequeo revisa la consola y la vista.

- Se aceptan datos semilla y cuentas locales preparadas, como ya dice `verification.md:51`.
- No exige un build por tarea, ni producción, ni el recorrido completo; el recorrido sigue a cargo de los hijos.
- Quedan fuera los cambios de solo texto o un valor de estilo (`ui-gate.md:7`).
- `verification.md:51` no cambia.

**Puntero en el rol:** `content/agents/development/hive-build-frontend.md`, punto 4, apunta a esa sección. El implementador delegado no lee `ui-gate.md` por sí mismo.

### G4 — Candidato fijo para los hijos del gate

**Casa canónica:** `ui-gate.md`, secciones "The independent children" y "When the gate runs".

**Identidad del candidato:**
- Si `git status --porcelain` está vacío, es el commit `HEAD`.
- Si no, que es lo habitual porque `flow-build/SKILL.md:61` deja la UI sin commit hasta el gate, es el hash de árbol que da este comando, sin tocar el índice real ni el árbol de trabajo:

  ```sh
  idx=$(mktemp); cp "$(git rev-parse --git-path index)" "$idx"
  GIT_INDEX_FILE="$idx" git add -A && GIT_INDEX_FILE="$idx" git write-tree; rm -f "$idx"
  ```

- Con esa identidad, `git diff <árbol-1> <árbol-2>` da el diff entre candidatos que pide `ui-gate.md:26`.
- El brief de cada hijo nombra esa identidad, y el hijo la repite en su reporte.

**Servir:** cada candidato se sirve desde su propio directorio de build y su propio puerto, que las ediciones posteriores no reconstruyen. Nunca desde un servidor con recarga en caliente. El brief nombra el estado que comparten los candidatos, como la base de datos y las cuentas.

**Corrección durante un gate en curso:**
1. La corrección produce un candidato nuevo.
2. Los hijos en curso terminan sobre el suyo.
3. Al volver, el diff entre candidatos se revisa de nuevo solo con el hijo que reportó, según `ui-gate.md:26`.
4. Las grabaciones de los flujos que tocan archivos cambiados quedan `stale`. `flow-walkthrough/SKILL.md:38` extiende su regla de rondas de corrección a estas correcciones.

Se descartó cancelar hijos, porque `delegation.md` no tiene esa regla y la capacidad depende del host.

**Precedente:** `plan-review.md:26` ("Keep the candidate stable while that round runs").

**Listas del brief:** `flow-plan/SKILL.md:48` y `plan-format.md:121` agregan "candidate identity".

### G6 — Confirmar una selección delegada

**Casa canónica:** `content/skills/flow-plan/references/questions.md`, "Pause and handle the answer", condición "an answer arrives, including free text". Cuando la respuesta deja una selección al criterio del agente, como qué hallazgos resolver y cuáles observar o qué elementos incluir, el agente no actúa sobre ella hasta confirmarla:
- lista la selección resultante, con lo que entra y lo que queda fuera, cada elemento con su glosa;
- pregunta si el usuario está de acuerdo, como una pregunta nueva con ID, con la recomendación de seguir así y las correcciones en texto libre.

Una respuesta que elige una opción concreta no es una selección delegada.

**Cómo se pregunta y excepciones** (re-revisión del 2026-10-19):
- **Pregunta nueva:** la confirmación es una pregunta nueva, no repetir una respuesta ya dada (`questions.md`, "Re-ask an answered question only when…").
- **Sin ronda extra:** si de todos modos sigue otra pregunta, la lista y la confirmación van en esa misma ronda.
- **Ruta iterativa:** una selección dentro de los límites de una ruta iterativa autorizada se informa en la línea de progreso de la ronda, porque la regla global ya autoriza esas rondas.
- **Modo desatendido:** la selección se registra con las decisiones tomadas, como fija `unattended-delegation`.
- **Copia en `global.md:20`:** lleva la condición, la lista de lo que entra y lo que queda fuera, la confirmación antes de actuar y un puntero a las excepciones.

**Copias que se corrigen**, porque `AGENTS.md` pide corregir una copia cuando cambia la regla que repite:
- `global.md:20` ("After each answer, state what it settles or invalidates before acting") agrega la cláusula mínima;
- `site/src/content/docs/flows/research.md:51` describe el caso.

### G5 — Pregunta de validación del recorrido

**Casa canónica:** `content/skills/flow-walkthrough/SKILL.md`, sección "Give the opening command". El encabezado se conserva. La primera vista (`:24`) y la validación final (`:42`) apuntan a ella (D3-A).

**Opciones de la tarjeta:**
- **Abrir el visor:** el agente ejecuta `hive view open`, sin otro comando, y vuelve a hacer la misma pregunta, sin cambios y con el mismo ID. Se ofrece solo si `hive view register` tuvo éxito. El mensaje conserva la línea con la ruta de la página, porque `open` abre la raíz del visor.
- **Aprobar:** su efecto depende de la parada.
  - En la primera vista, acepta la composición y el build sigue hasta el gate del candidato construido (`ui-gate.md:25`).
  - En la validación final, es el aceptar de `:54`, que dispara el gate del candidato aceptado.
- **Correcciones** (D2-A): no tienen opción. La pregunta dice que se escriben en el campo libre de la tarjeta.

**Cuándo va en texto:**
- cuando la herramienta de preguntas no muestra en su esquema o documentación una respuesta libre;
- cuando no se ofrece "Abrir el visor", porque la tarjeta quedaría con una sola opción y los hosts piden al menos dos (`agent-delivery.md:197-198`).

En texto, la pregunta lleva sus opciones y una línea que dice que las correcciones se escriben en la respuesta.

**Excepciones locales a `questions.md`,** nombradas en el punto de uso:
- `:44`: hay una condición nueva para preguntar en texto.
- `:46`: se permite la línea de correcciones.
- `:62` y `:68`: volver a preguntar tras abrir no cuenta como repetir una pregunta. "Abrir el visor" es una acción intermedia, no una respuesta que resuelva la pregunta.
- `:64` no aplica, porque la pregunta no cambia.
- `questions.md` en sí no cambia.

**Fallos de `hive view open`:**
- **Visor apagado** (la salida dice "not running", con código 0): el agente informa en una línea que el usuario puede iniciarlo con `hive view start` y vuelve a preguntar.
- **Otro error:** el agente informa el error en una línea y vuelve a preguntar en texto, sin "Abrir el visor".
- En ningún caso el agente ejecuta los comandos que sugiere la salida, ni reintenta. `hive view start` sigue prohibido (`:44`).

**Regla de no abrir:** `:48` queda así: abrir solo con `hive view open` y solo cuando el usuario elige esa opción. Sigue prohibido abrir la página o un navegador por cualquier otra vía, incluido el navegador del agente.

**Seguridad:** sin cambio respecto a #78. El abridor recibe un código de un solo uso de 60 s, que no aparece en la salida (`view_command_test.go:459-463`). La exposición breve en `ps` es la misma que si el comando lo ejecuta el usuario. El PR lo menciona.

**Textos alineados:**
- `flow-build/SKILL.md`, en el paso de entrega del recorrido, que pasa a ser puntero.
- `site/src/content/docs/flows/build.md:75`.
- `CHANGELOG.md:92-94`.
- `_support/docs/architecture/deployment-manager.md:124` ("The user starts and stops it; agents register…"), que agrega que el agente ejecuta `open` cuando el usuario lo elige.
- `README.md:355-361` no cambia: "Agents never start it" sigue siendo cierto.
- `global.md` no se toca.
