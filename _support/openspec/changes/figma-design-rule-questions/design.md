# Design

## Approaches

- **A — Dentro de la etapa, con salida a investigación (elegido, D6 A).** La regla se dibuja recomendada, se anota como propuesta y se pregunta aparte en la ronda de aprobación. Las que cambian el alcance llevan a investigación. Así la regla se decide viendo la pantalla, y solo sale de la etapa lo que cambia el alcance.
- **B — Siempre a investigación.** Separa más, pero cada duda obliga a salir, reentrar y redibujar. En FSR-858, D10–D16 tomaron unos 25 min.
- **C — Solo marcar y que el plan decida.** Hay menos preguntas al diseñar, pero la regla llega después de aprobar la pantalla, y un cambio obliga a volver a Figma.

## Decisions

- **Qué es una regla y cuándo está fijada.**
  - Una regla es un comportamiento que un frame muestra en una situación: qué exige una acción, quién la toma, qué estado sigue, qué bloquea o conserva, dónde vive un ajuste que aplica más allá de la pantalla.
  - Está fijada cuando la fijan la petición, los criterios del ticket, las decisiones de investigación, las decisiones registradas de la app o lo que ya implementan las pantallas hermanas o su código.
  - Criterio observable: se puede señalar en el frame y no aparece en esas fuentes.
- **Regla menor.** Es la que tiene opciones que no cambian alcance, contrato, datos guardados, costo, reversibilidad ni quién puede actuar. Se declara como supuesto y no se pregunta. Es el mismo criterio de decisión menor de la guía global, más «quién puede actuar».
- **Prueba de alcance de esta etapa.**
  - Una regla cambia el alcance si añade una capacidad no pedida, cambia datos o ajustes que usan otras pantallas o módulos, o cambia un permiso.
  - Es una prueba propia de la etapa, no una clase de «Scope changes»: allí una tarea añadida va al plan, y aquí la regla necesita investigación antes de dibujar.
  - No hay umbral por número de reglas, porque el número no dice si el alcance cambia.
  - Se aplica a las reglas de quien dibuja y a las del revisor.
- **Orden en la ronda.**
  - La decisión de la regla y la de su pantalla van en la misma ronda. La opción de aprobar la pantalla nombra la decisión de la regla y vale con cualquier respuesta, como permite la regla global de prerrequisitos.
  - Si la respuesta difiere de lo dibujado, se redibuja esa parte y la pantalla vuelve a pedirse, porque el usuario no aprueba algo que no vio.
- **Registro.**
  - Formato de línea: `rule: <regla> · screen <pantalla> · proposed | assumption | approved <fecha>: <respuesta> · found by drawer | reviewer`.
  - Va separada de las líneas de estado de pantalla, para que el plan no confunda una regla con una pantalla aprobada.
- **El revisor no recibe la lista de propuestas** de quien dibujó, para que su búsqueda sea independiente (D7 A). Las decisiones de investigación van entre sus requisitos, no como lista de «ya resuelto»; el hilo principal une ambas listas.
- **Vuelta de investigación.** La etapa retoma su registro y dibuja solo las pantallas detenidas. Las aprobadas siguen aprobadas.
- **Desatendido.** Las reglas quedan propuestas y se reportan sin resolver. Una regla de alcance igual detiene su pantalla, y el plan no decide una regla de alcance propuesta.

## Impact

- **Lectores del registro `examples.md`:** `flow-plan`, paso 4 de UI.
- **Revisor de frames:** `hive-review-ux` y el brief en `delegation.md`.
- **Transición nueva:** de diseño a investigación, en la lista de `global.md` línea 90.
- **Pruebas y detectores:** ninguna prueba ni detector de `tests/pilot/` lee estos textos. `agent_boundary_test.go:43` solo nombra el rol.

## Recovery

Revertir el merge restaura el comportamiento anterior. Las sesiones abiertas conservan los skills que cargaron al arrancar.
