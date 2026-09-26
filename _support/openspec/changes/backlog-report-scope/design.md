# Diseño

## Guía (`global.md`)

- **Regla de forma del chat:** al final de la frase de tablas, listas y prosa se agrega: "Put each ticket on its own list line, or each set of tickets meant to be worked together, with a description the reader understands without opening the ticket; never put three or more tickets in one paragraph." Vive en la regla de forma y no en la del backlog porque el fallo aparece en cualquier respuesta que cite tickets.
- **Regla del backlog:** abre con el disparador y la excepción juntos, antes del procedimiento: "When the user asks for the project's status or a listing of its backlog, give the backlog report below. When they ask for another cut of the tickets, such as those that fit one session, those blocking a release, or a single module's tickets, structure the answer by that cut instead, and include state counts or delivery state only when they change the answer." Después sigue "The backlog report groups five or more tickets by product module…". La frase "Give every ticket its ID and a one-line description" se reduce a "Give every ticket its ID": la descripción ya la exige la regla de forma, y así la regla tiene un solo hogar.
- **Por qué los ejemplos de corte son tres:** un ejemplo solo se leería como el alcance completo de la regla (`AGENTS.md`, Guidance structure).

### Auditoría para modelos pequeños (`prompt-audit`, 2026-09-26)

Se aplicó `shared/prompt-audit.md` de la skill `claude-api` a la primera redacción. El modelo objetivo fue el perfil de ejecución más débil (`deepseek-v4.1-flash` en OpenCode) y la clase Haiku, que pidió el usuario.

- **P1:** el disparador "status or its backlog" atrapaba una pregunta de agrupación, y la excepción iba al final de un bullet largo. Se movió junto al disparador y se acotó a "a listing of its backlog".
- **P2:** "several" y "each group of tickets that belongs together" eran vagos. Se reemplazaron por "each set of tickets meant to be worked together" y "three or more", que coincide con el criterio.
- **P3:** "unless they change it" pasó a "unless they change the answer".
- **P4:** la regla nace de un solo caso ("recency trap"). Solo se señaló: acota una regla existente que causó el fallo y la hace verificable un criterio determinista.

La prueba con Haiku está en la evidencia (historical evidence omitted from public history).

## Criterio determinista `ticket_ids_not_packed_in_prose`

- **Superficie:** el texto de cada mensaje del asistente. Antes de buscar se enmascaran los bloques de código, el código en línea, la parte URL de los enlaces, los encabezados y el contenido de las negritas. Los dos últimos son etiquetas: una etiqueta puede nombrar los miembros de un conjunto que va junto.
- **Párrafo:** bloque separado por líneas en blanco. Quedan fuera las líneas de lista (`-`, `*`, `+`, `N.` o `N)` seguido de espacio) con sus continuaciones indentadas, y las de tabla (`|`).
- **Falla:** un párrafo de prosa con tres IDs de ticket distintos o más (`\b[A-Z][A-Z0-9]{1,9}-\d+\b`) fuera de sus etiquetas. La evidencia nombra el mensaje y los IDs, sin el texto que los rodea.
- **Límites declarados:**
  - No detecta números de issue estilo `#123`.
  - No revisa el texto de la pregunta nativa.
  - No cuenta los IDs en negrita dentro de la prosa.
  - Un párrafo que compara legítimamente tres tickets cuenta como fallo. Es coherente con la regla.
  - H1 (seguir el corte pedido) no se expresa. Queda en [#39 (juzgar el corte pedido)](https://github.com/JhonHawk/tricell-hive/issues/39).
