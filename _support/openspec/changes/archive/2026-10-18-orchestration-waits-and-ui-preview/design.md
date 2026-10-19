# Diseño: cuatro momentos, una casa cada uno

## Contexto verificado

**Evidencia de la sesión cam.** Sesión de Claude Code `e2c8d39f` del proyecto cam, 2026-10-18, modelo opus-5-5, con Hive desplegado `9c131d3b` (y `flow-walkthrough` leído en `df6e550e`). Lo que se comprobó en el transcript:

- **Hijo detenido:** el hijo `Verify T4` (`hive-verify-task`) cerró su turno con «Waiting for the monitor to report» a las 23:13:05Z. Su transcript no tiene eventos hasta las 00:31:07Z, cuando se lo reanudó a mano, y en ese tiempo el orquestador escribió «sigue en curso» cuatro veces.
- **Commit antes de la primera vista:** T6 se comprometió a las 17:01 hora local (`38c8cf88`), antes de la primera vista (17:05) y de la revisión de interfaz (18:07), y acabó con cuatro commits.
- **Validación pedida por el usuario:** el HTML del recorrido se abrió a las 00:46:23Z, después de que el usuario escribiera «abre el html para revisar el recorrido» (`queue-operation` de las 00:46:19Z). La primera vista solo tuvo PNG sueltos.
- **Tarjeta y cola:** con la tarjeta S1 abierta, un aviso de un subagente quedó en cola de 01:19:48Z a 01:26:12Z sin abrir turno.

**Reglas que se mapearon** (rama `feat/gh-67-flow-walkthrough-skill`, `8c15f59a`):

- **Turno de notificación (M1):**
  - `reporting.md:45-47`: el ítem en curso es incondicional al final de cada turno con trabajo corriendo.
  - `reporting.md:91`: una línea solo cuando hay trabajo propio pendiente.
  - `delegation.md:61`: manda el ítem en curso completo en cada turno de notificación.
  - `delegation.md:65`: una línea por tramo de espera bloqueante.
  - `flow-walkthrough/SKILL.md:24`: «nothing is repeated for them», excepción local con tarjeta.
  - `flow-build/SKILL.md:67`: vida del hijo solo tras silencio largo.
  - `hive-verify-task.md:14`: primer plano. El hijo de cam lo incumplió.
- **Apertura del mensaje (M2):**
  - «Lead with the result» en `global.md:16` y `reporting.md:9`.
  - Ítem en curso primero (`reporting.md:9,47`).
  - Reporte de cierre «it is complete» (`reporting.md:51`).
  - Bloque «Pausa» (`flow-plan/SKILL.md:18`).
  - Línea de pausa al final (`questions.md:60`).
  - La tarjeta se sostiene sola (`questions.md:38`).
  - La primera frase lleva el resultado en las voces (`content/voices/preamble.md:7`, fijada por `tests/content/voices_test.go:40`).
- **Antes de la validación humana (M3):**
  - commit al llegar a `[x]` (`flow-build/SKILL.md:61`);
  - suite completa solo si no corre CI (`flow-build/SKILL.md:73`, `verification.md:67,70`);
  - revisión de código y de seguridad tras la entrega (`verification.md:33,43`, `delivery-decisions.md:9,26`, `ui-gate.md:26`);
  - primera vista como capturas sueltas (`flow-walkthrough/SKILL.md:24,30`);
  - el usuario abre la página (`flow-walkthrough/SKILL.md:44`, `site/.../flows/build.md:73`).
  - Ningún texto de `content/` abre una página hoy.
- **Brief (M4):**
  - `delegation.md:27,35,37`;
  - `flow-build/SKILL.md:34`;
  - `ui-gate.md:21`;
  - `plan-review.md:26`, que ya pasa rutas de skills a los revisores del plan.

Los briefs de cam no pasaron `heroui-pro-design-taste` ni `heroui-baseline.md`, aunque la regla condicional del proyecto los activaba (fuente: el propio orquestador de cam, no verificado en el repositorio cam).

**Pruebas que verifican texto:** ningún detector ni prueba Go verifica el texto de estas reglas. `tests/content/voices_test.go:40` fija la primera frase del preámbulo de voces. Las anclas no las valida ninguna prueba.

## Por qué por momentos

El usuario pidió evaluar los seis cambios en conjunto. Varios gobiernan el mismo instante y, escritos por separado, chocarían:

- **C2, C3 y C6 (M1):** los tres deciden qué escribir cuando un aviso abre un turno. C2 pide una línea, C3 pide comprobar y reanudar en vez de «una línea y seguir», y C6 pide la línea de contexto.
- **C6 (M2):** compite con otras cinco aperturas por la primera línea.
- **C1 y C4 (M3):** los dos deciden qué pasa antes de preguntarte. C1 abre la página; C4 corre la suite y la seguridad y fija el commit.
- **C5 (M4):** repite con otro alcance lo que ya dicen `delegation.md:35`, `ui-gate.md:21` y `plan-review.md:26`.

Cada momento tiene una sola casa, que enuncia el orden completo; los demás lugares solo apuntan. Así lo pide `AGENTS.md` (sección Guidance structure).

## M1 — El turno que abre una notificación (C2, C3 y la forma de C6)

**Casa:** `reporting.md`, sección «Report a running or completed item». Sigue teniendo el ancla `#report-a-running-or-completed-item`, porque la citan `delegation.md` y `unattended-delegation/SKILL.md`. Lleva una condición nueva antes de la del ítem en curso.

- **Condición:** un turno lo abre una notificación de un subagente o de un proceso en segundo plano, no un mensaje del usuario.
- **Acciones, en orden:**
  1. Si el resultado o el aviso dice que el hijo paró mientras espera trabajo propio, ese hijo está detenido. Aplica la comprobación de [delegation](…#check-returned-work) antes de escribir nada.
  2. Si el resultado decide algo, trae un bloqueo nuevo o cambia lo que se le pide al usuario, repórtalo como manda el resto de esta referencia, con el ítem en curso cuando otro trabajo sigue corriendo.
     - Si una pregunta espera respuesta y el resultado no la cambia, se toma la salida 3 aunque el resultado decida otra cosa (D2-A, 2026-10-18): la pregunta sigue tal como se hizo y no se repite.
     - Si el resultado cambia lo que se pide, la nueva petición es una pregunta nueva con ID nuevo, que reemplaza a la anterior y termina con su propia línea de pausa, como fija `questions.md`.
  3. Si nada de eso, responde en una línea: la línea de contexto (M2) seguida del cambio. No re-listes lo que sigue corriendo ni repitas una pregunta abierta o su pausa. Esa línea no es un segundo reporte.
  4. Si queda trabajo propio, continúa con la siguiente llamada de herramienta. Si no, termina el turno.
- **Ítem en curso:** su condición actual («a turn ends while work is still running») deja de aplicarse solo a la salida 3.
- **Intención que se conserva:** la frase de `flow-walkthrough/SKILL.md:24` («their results wait for the answer») queda absorbida por la salida 2: los resultados que no cambian la pregunta abierta no se presentan como una nueva petición.
- **«Work-item boundary»:** pierde la cláusula «answer a question or notification that arrives meanwhile in one line and continue». Conserva la de pregunta, con la forma de una línea de M2, y apunta a esta condición para las notificaciones.
- **Excepción en `questions.md`:** en «Pause and handle the answer» (`questions.md:60` y `:64`), en el punto de uso, nombrando lo que cambia:
  - la línea de una salida 3, con una pregunta abierta, no lleva línea de pausa y no cuenta como segundo reporte; con una pregunta abierta, todo resultado que no la cambia toma esa salida (D2-A);
  - una petición que cambia es una pregunta nueva con ID nuevo, que reemplaza a la anterior.
- **Ejemplo resuelto:** el de `delegation.md:81` adopta la forma de la línea de contexto.

**Hijo detenido:** vive en `delegation.md`, sección «Check returned work», con una condición nueva en términos de capacidad. Un hijo que vuelve, o un aviso del host que dice que se detuvo, mientras espera trabajo propio (como un vigilante, una espera o un comando en segundo plano) se ha detenido y su resultado puede ser provisional. Se comprueba en el acto si siguen vivos él y ese trabajo, se reanuda el mismo hijo o se repite la comprobación, y nunca se lo reporta como en curso. La frase literal del host de Claude Code va solo en `agent-delivery.md`.

**Punteros:**

- `delegation.md:61` cambia «ends its turn with the in-progress item» por «handles that turn as [reporting](…#report-a-running-or-completed-item) sets for a turn a notification starts».
- `delegation.md:65` mantiene la línea por tramo de espera bloqueante, que no es un turno de notificación, pero en forma de línea de contexto: etapa · ticket · avance, y lo que sigue corriendo y qué decide.
- `flow-walkthrough/SKILL.md:24` cambia «their results wait for the answer, and nothing is repeated for them» por un puntero a la condición de `reporting.md`. Esta edición la hace T3, dueña del archivo.
- `flow-build/SKILL.md:67` conserva la regla del silencio largo y la de «confirm it left no running processes», y apunta a la condición de hijo detenido para el caso en que el propio hijo avisa.

Los roles no cambian: diez ya prohíben esperar en segundo plano, y `hive-verify-task` lo prohíbe y falló. La protección vive en el orquestador.

## M2 — Cómo abre cada mensaje (C6)

**Casa:** `reporting.md`, sección nueva «Open with the context line», justo después de «Lead with the result».

- **Condición:** un mensaje informa progreso, entrega un resultado de una etapa o pide una decisión o validación.
- **Acciones:**
  - Su primera línea es la línea de contexto: `<stage> · <ticket(s)> · <progress>`, en el idioma de la sesión y en negrita.
    - La etapa es la del flujo: investigación, plan, construcción, validación de UI, cierre.
    - El ticket va por ID, o el nombre del cambio cuando no hay ticket.
    - El avance es el contador propio de la etapa cuando existe: tareas verificadas de las totales, subagentes que volvieron de los lanzados, ronda de las acordadas. Si la etapa no tiene contador, se omite.
  - Después va la apertura que corresponda, con el orden actual de «Lead with the result»: lo que corre, luego el trabajo completado, luego la respuesta; o el bloque «Pausa» del plan. Luego el cuerpo y, al final, la línea de pausa.
  - **Notas de una línea:** una nota de progreso de una línea, es decir, un mensaje que es solo esa línea, es la línea de contexto seguida del cambio, en esa misma línea. Una frase de una línea dentro de un mensaje mayor no lleva línea de contexto propia, como el modo de entrega antes de la primera edición (`delivery-decisions.md:5`), la decisión menor (`questions.md:11`) o un valor declarado (`flow-plan/SKILL.md:18`). Así quedan, sin editar sus textos:
    - la nota entre tareas y la respuesta a una pregunta que llega mientras hay trabajo (`reporting.md:91`);
    - la ronda de una ruta iterativa (`flow-build/SKILL.md:79`, `flow-research/SKILL.md:42`);
    - el aviso antes de un lanzamiento bloqueante (`delegation.md:57`);
    - la línea por tramo de espera (`delegation.md:65`);
    - lo que decide una respuesta (`questions.md:68`, `flow-plan/SKILL.md:63`);
    - la salida 3 de M1.
  - Una tarjeta nativa, que se muestra sola, empieza su pregunta con etapa y ticket.
  - La línea de contexto es un encabezado, no la primera oración. La casa lo dice expresamente, y `content/voices/preamble.md:7` añade «after any context line», así que la prueba de voces (`tests/content/voices_test.go:40`, que fija una subcadena) no cambia.
- **Excepciones:** el saludo, la respuesta corta directa sin trabajo en curso, el retorno de un subagente al orquestador y los archivos.

**Invariante en `global.md`.** Se reescribe la viñeta de Communication que hoy dice «Lead with the answer or outcome…» para añadir la invariante completa:

- todo mensaje de progreso o de decisión al usuario abre con una línea de contexto (etapa · ticket · avance), y un mensaje que es solo una nota de una línea es esa línea con el cambio;
- un turno que abre una notificación sin cambios para el usuario es esa nota, sin re-listar lo que corre ni repetir una pregunta abierta;
- no aplica al retorno de un subagente ni a una respuesta corta sin trabajo en curso.

Va en `global.md` porque `global.md:17` exime de leer `reporting.md` a una respuesta corta, y porque los hijos también cargan `global.md`. El cambio de tamaño se informa.

**Punteros:**

- `questions.md` paso 4 de «Use a waiting native card»: «the question opens with the stage and ticket that [reporting](…#open-with-the-context-line) sets».
- `flow-plan/SKILL.md:18` (bloque «Pausa») no cambia: el orden de M2 lo coloca después de la línea de contexto.
- **Ejemplo resuelto:** el de `reporting.md:95-109` abre con la línea de contexto.

**Anclas fijadas:** `#open-with-the-context-line` (nueva), `#report-a-running-or-completed-item` (se conserva), `#check-returned-work` y `#deliver-and-resolve-instructions` (existentes en `delegation.md`).

## M3 — Antes de pedir la validación humana (C1, C4 y D1-A)

**Casa del orden:** `flow-build/SKILL.md`, en el párrafo de verificación (hoy línea 71), que ya dice qué hace falta antes de reportar completo un cambio de UI. Se añade la secuencia antes de pedir al usuario que valide un resultado, para cualquier modo que se detenga a validar:

1. la revisión de interfaz que la `UI verification` selecciona;
2. la suite completa de la aplicación afectada, en local sobre el candidato, aunque la CI la vaya a correr después. Bajo `interactive` la CI corre tras el push, que viene después de la validación;
3. `hive-review-security` cuando el cambio cruza una frontera de [security boundaries](…), con la autorización que ya fija `verification.md:43`: la revisión acordada o la que da un valor `Review` declarado. La que el veredicto del revisor de código nombra después sigue en el momento de ese revisor. Si un arreglo de seguridad toca archivos de UI, la revisión de interfaz se repite sobre ese diff, con los hijos que el valor selecciona, antes de pedir la validación;
4. la entrega de `flow-walkthrough`, que abre la página antes de preguntar.

La revisión de código dedicada sigue donde la pone el modo de entrega, porque el usuario adoptó mover solo la de seguridad. Si una corrida local no avanza, como fija `verification.md:70`, se para y el mensaje de validación la nombra como no corrida. Las correcciones que el usuario pida durante la validación siguen cambiando código; la regla evita los cambios que salen de puertas que pudieron correr antes.

**Commit (D1-A):** vive en `flow-build/SKILL.md:61`.

- **Cuándo espera:** una tarea con efecto de UI renderizado llega a `[x]` y queda sin commit hasta que pasan dos cosas: la primera vista del cambio, cuando se mostró, fue aceptada, y la revisión de interfaz del cambio pasó.
  - Bajo `UI verification: none`, que no lanza hijos, en vez de la revisión de interfaz se espera la comprobación renderizada del implementador que fija `ui-gate.md`.
  - Bajo `automatic`, se espera solo la revisión de interfaz.
- **Quién más espera:** una tarea sin UI que comparte archivos con una de esas tareas, o que depende de ella, también espera.
- **Cómo se compromete:** después, un commit por tarea, en el orden de las tareas. Las correcciones de las rondas de esa revisión de interfaz entran en el commit de su tarea.
- **Arreglos sobre una tarea ya comprometida:** los de la suite completa o de la revisión de seguridad (pasos 2 y 3), y las correcciones de la validación humana, que reabren la tarea. Llevan un commit adicional `(T<n>)` por cada tarea afectada, tras su verificación o tras la puerta del candidato aceptado, nunca un amend, porque `git-workflow/SKILL.md:24` prohíbe reescribir el historial. Es una excepción a «every commit holds one verified task and its fix rounds», y la regla la nombra.
- **Si el trabajo se pausa antes:** el reporte nombra esas tareas como verificadas sin commit.
- **Tareas sin efecto de UI:** se comprometen al llegar a `[x]`, como hoy.

Para el usuario, la frase que lo resume es «first rendered view and the UI gate».

**Casa de la página:** `flow-walkthrough/SKILL.md`.

- **«First rendered view»:** la primera vista se escribe en `walkthrough.html`, junto a `images/`, con el llenado parcial del esqueleto. Se corre `self_check.py`, se abre la página como fija «Open the page» y después va el mensaje con la pausa y la tarjeta. La página final del recorrido reemplaza ese archivo, así que `artifact-layout.md` y `flow-close` no cambian. La skill solo nombra el llenado parcial; la lista va en el comentario del esqueleto, para cuidar el tope de palabras.
- **Llenado parcial** (en el comentario de uso del esqueleto):
  - **Se conserva:** la cabecera, con el candidato como `working tree on <base>` porque todavía no hay commit; una `section.flow` por vista cambiada con solo su galería, sin pasos ni vídeo, para que el lightbox siga encontrando `.gallery button[data-full]`; «Action required» con la pregunta de composición; el entorno; la procedencia.
  - **Se borra:** la tabla de validación, los pasos y la grabación.
- **«Open the page»:** párrafo nuevo bajo «Deliver», que comparten la primera vista y la página final.
  - Abre el archivo local en el visor predeterminado del sistema; se nombra la capacidad, no la herramienta.
  - Solo el HTML, nunca capturas sueltas ni una carpeta.
  - Una vez por versión nueva de la página, antes del mensaje que pide la decisión o la validación.
  - El mensaje da siempre la ruta absoluta completa, porque abrir no prueba que el usuario la vea.
  - Cuando el host no puede abrirla (remoto, sin interfaz, o con un rechazo del sandbox o de un permiso), el mensaje lo dice.
  - Nunca publica: la sección «Share» sigue igual.
- **Ajustes de texto:** `:44` pierde «the user opens the page in a browser or from the editor», y `:30` pide mostrar las capturas a través de la página.
- **Fila de `flow-report`:** la fila de `flow-report/SKILL.md:22` nombra también la página de primera vista.

**Punteros a la casa del orden:**

- `ui-gate.md`, «When the gate runs»: solo la viñeta de `:25` y la última frase de `:26`. La viñeta apunta a la secuencia de `flow-build` y conserva el orden primera vista → revisión; la frase deja la revisión de código donde la pone el modo. El resto de `:26`, la puerta del candidato aceptado a la que apunta `flow-walkthrough/SKILL.md:50,52`, no se toca.
- `verification.md:33` (fila «Local candidate»), `verification.md:43` (seguridad desacoplada del revisor de código, «at the point flow-build sets») y `delivery-decisions.md:9` (`interactive`: «local commits, verification and the gates flow-build runs before validation, then stop before push»).

## M4 — Qué lleva el brief de un subagente (C5)

**Casa:** `delegation.md`, sección «Deliver and resolve instructions» (hoy `:35`). Se extiende la frase de lo que pasa el padre: «the skills and project documents that the project's rules direct for this kind of work, evaluating each conditional rule against what this host exposes now, by identity or path». La lista que ya existe («project-guidance paths, task decisions, existing project conventions, and the exact skill or resource identities») queda como el marco general.

**Punteros:**

- `flow-build/SKILL.md:34`, en el brief de una tarea del plan: «and the skills and project documents that [delegation](…#deliver-and-resolve-instructions) sets».
- `flow-build/SKILL.md:67` (M1): el puntero a la condición de hijo detenido es `delegation.md#check-returned-work`.
- `ui-gate.md:21`, en el brief de revisores: el mismo puntero, y «passing them is not pre-rating findings».
- `plan-review.md:26` no cambia: ya pasa rutas de skills.

## Decisiones tomadas sin preguntar

Cada una se puede revertir editando la casa de su momento:

- M2 va en `reporting.md` con una frase invariante en `global.md`, y no en `global.md` completo, para no repetir la regla.
- La respuesta de una línea de M1 es la propia línea de contexto con el cambio, no dos elementos.
- La tarjeta lleva etapa y ticket, sin avance, porque el avance ya va en el mensaje.
- El avance es el contador propio de la etapa, y se omite si la etapa no tiene uno.
- La primera vista reutiliza `walkthrough.html`.
- Bajo `automatic`, el commit de una tarea de UI espera la revisión de interfaz; bajo `UI verification: none`, la comprobación renderizada del implementador.
- Una tarea sin UI que comparte archivos con una de UI sin commit, o depende de ella, también espera.
- Las correcciones de la validación humana van en un commit adicional por tarea reabierta, sin amend.
- Un cambio que pide algo distinto mientras otra pregunta espera es una pregunta nueva con ID nuevo, que reemplaza a la anterior.
- La página siempre lleva su ruta absoluta en el mensaje, aunque se haya abierto.
- La suite local corre antes de validar aunque haya CI.
- El tope de `ui-gate.md` sube de 850 a 880 palabras.

## Revisión del plan

Ronda 1 sobre `51751a7641f2`, en dos dominios: coherencia del conjunto y orden y ejecución. Las correcciones aplicadas son:

- Excepción en `questions.md` para la línea de pausa y para el reemplazo de una pregunta.
- La salida 2 de M1 conserva el ítem en curso.
- La intención de `flow-walkthrough:24` queda absorbida en M1.
- La invariante de `global.md` lleva sus excepciones.
- Todas las notas de una línea siguen la forma de M2.
- Aclaración de la línea de contexto en `preamble.md`.
- Anclas fijadas.
- Commits de D1-A: rondas de la revisión de interfaz, correcciones de validación, archivos compartidos, `UI verification: none` y pausa.
- La revisión de seguridad conserva su autorización y su caso posterior al revisor de código; un arreglo de seguridad que toca la UI repite la revisión de interfaz.
- La ruta absoluta siempre en el mensaje.
- Edición de `ui-gate.md` acotada.
- Llenado parcial en el comentario del esqueleto, con la estructura que espera el lightbox.
- Ejemplos resueltos y páginas del sitio añadidos a las tareas.
- Verificaciones ampliadas para AC8–AC11.

## Fuera de alcance

- C7: el issue #71.
- Legislar sobre que una tarjeta abierta retenga los avisos: comportamiento observado una vez en Claude Code y no documentado.
- Aislar las sesiones de navegador entre revisores: queda como punto a vigilar.
- Cambiar los roles.
- Añadir un campo de referencias a las tareas en `plan-format.md`.
