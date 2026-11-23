# ledger-writer-role — diseño

## Enfoque

Solo hay una forma sensata, porque el usuario eligió un rol nuevo (D23 A). Su perfil es `implement` por su bloque de límites (D24). El rol se agrega como cualquier otro de `content/agents`; el instalador y las pruebas de render ya recorren esa carpeta, así que no hace falta código nuevo.

## El rol

`content/agents/review/hive-write-ledger.md`, con `model_profile: "execution"` y `access_profile: "implement"` (D25).

Descripción: «Classify each item of one audit lane by the criterion its brief gives and write one row per item to the ledger file the brief names, in the host's temporary location, editing nothing else. Use for the lanes of a comment or test audit that keep a ledger; not for research that returns its answer in the reply, which `hive-research` does.»

Pasos:
1. Lee el criterio que el brief adjunta (la sección de la skill y la hoja que nombre) y cada elemento de su carril con el código alrededor y su historial.
2. Escribe una fila por elemento en el archivo ledger que el brief nombra, en el formato que el brief da. Una fila sin evidencia toma el valor por defecto que fija el criterio del brief (D27). Nunca copia un secreto al ledger.
3. No edita ningún otro archivo, no instala dependencias, no corre suites que compartan un servicio local y no escribe en trackers.
4. Devuelve una respuesta corta: la ruta del ledger, los conteos por categoría, las filas sin evidencia y las listas que pida el brief.

Termina con el bloque de límites de `implement` que exige `agent_boundary_test.go`.

## Skills y hojas

- `audit-comments` paso 4: «one `hive-research` child» pasa a «one `hive-write-ledger` child»; el resto del brief no cambia.
- `audit-tests` paso 4: igual. Su descripción cambia «a read-only ledger per test» por «a ledger per test written without editing the tests».
- `hive-core/delegation/delegation.md` «When children run at the same time» (D28): tras la línea de hijos de solo lectura, «A child whose only write is its own output file in the host's temporary location, which no other child writes, runs side by side as a read-only one does.»

## Documentación

- `agent-delivery.md`:
  - línea 9: «Twenty-one roles» pasa a «Twenty-two roles» y nombra `hive-write-ledger`, las dos skills que lo usan y su cambio;
  - párrafo de acceso (línea 55): que el límite a un solo archivo depende solo de sus instrucciones, porque ningún host común limita la escritura a una carpeta, y que en Claude el modo del padre (`acceptEdits`, `bypassPermissions`) aplica al hijo.
- `instruction-resources.md` línea 28 (entrada de `audit-comments`): nombra el rol que escribe sus ledgers. `audit-tests` no tiene entrada propia allí; `agent-delivery.md` nombra a las dos skills que usan el rol.
- `CHANGELOG.md`: una entrada en «Added».
