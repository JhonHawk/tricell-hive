# Design

## Approaches

Se presentaron tres:

1. **Skill propia, sin etapa en el visor (elegido, D3 A).** La investigación recomienda la etapa en su bloque de siguiente paso; la sección se muda a la skill nueva; el plan recibe el ejemplo aprobado. Una skill, una página del sitio y entre 12 y 15 ediciones. Las propuestas se ven en Figma, no en el visor.
2. **Dejar la sección en `flow-research` y hacerla alcanzable.** De 4 a 6 archivos, pero sin etapa ni comando propio.
3. **Skill propia más etapa «design» en el visor.** Suma de 20 a 23 archivos del visor, rompe las pestañas abiertas con la versión anterior y una captura exige un commit candidato que el diseño todavía no tiene.

Se eligió el primero porque da la etapa con comando propio y arregla la ruta por tickets sin abrir el protocolo del visor; el tercero puede hacerse después como cambio aparte.

## Decisions

- **Nombre y alcance (D10 A).** `flow-figma-design`, solo Figma. Los pasos pueden nombrar lo concreto de Figma (publicar la librería, variables y componentes locales, las instrucciones propias del conector), que es donde la sesión de cam se desvió. El campo del registro de diseño conserva su nombre neutral, «Design file».
- **Sin Figma no hay etapa (D9, D10 A).** Desaparecen de la etapa las opciones «bocetos en texto» y «sin propuestas», y con ellas la tarjeta de oferta: quien no tiene conexión a Figma va de la investigación al plan, que sigue dibujando sus bocetos en texto. El ejemplo aprobado pasa a ser siempre un frame.
- **Entrada por bloque, no por tarjeta (D4 A).** Pasar de un flujo a otro ya se hace con un bloque de siguiente paso y un comando; la investigación recomienda `flow-figma-design` y deja «ir directo al plan» como otra respuesta.
- **Archivo no declarado (D5 A).** La etapa pide el enlace una vez como pregunta abierta en el texto; sin respuesta, termina con la ruta al plan.
- **Registro en `DESIGN.md` (D6 A).** Se pregunta en la misma ronda que la aprobación del ejemplo, con opciones que valen bajo cualquier respuesta de aprobación. La etapa no edita el `DESIGN.md`: guarda el sí en el registro del ejemplo y el plan de la pantalla agrega la tarea que escribe el campo, porque esa edición es un cambio del repositorio que necesita su entrega. Cuando el entregable era el diseño y no hay plan, el cierre la nombra como pendiente.
- **Quién autoriza escribir en Figma.** Al desaparecer la tarjeta de oferta, la respuesta del usuario que inicia la etapa (escribir su comando o elegirla en el bloque) es la instrucción para esas escrituras; una escritura que gasta dinero o cuota de pago conserva su propia pregunta.
- **Límites del hijo (D7 A).** Crea frames en la página del trabajo y componentes y variables locales al archivo (en Figma las variables viven en colecciones del archivo, no de una página), anota sus identificadores en el registro y borra solo lo que creó, incluidos los que solo usaban los ejemplos no elegidos.
- **Correcciones de cam (D8 A).** Cuatro pasos explícitos: cargar el sistema de diseño, encargo con los anchos, actualizar el registro al aprobarse y avisar los pasos manuales.

Decisiones menores tomadas sin preguntar:

- La petición cuyo entregable es el diseño mismo (replicar una pantalla en Figma) se muda a la skill nueva con el mismo comportamiento: se dibuja sin recomendación previa, no se reporta como hecha antes de la aprobación y cierra con `flow-close` como tarea operativa. La investigación continúa hacia la etapa en el mismo turno, como cuando la dibujaba ella, de modo que el usuario no gana una parada; la guía global nombra esa continuación.
- Bajo delegación desatendida la investigación no recomienda la etapa; si ya estaba en curso, conserva todos los ejemplos y reporta la elección como no resuelta, como hoy.
- El delta retira el requisito «Screen examples in research» y agrega «Screen proposals in Figma before the plan», en lugar de modificar el anterior, porque cambia de nombre y de responsable.
- La etapa no publica nada en el visor, como la investigación hoy.

## Proposed shape

`content/skills/flow-figma-design/SKILL.md`, con pasos numerados porque el orden importa:

1. **Entrada.** Se entra por el comando, desde el bloque de la investigación, o por una petición cuyo entregable es el diseño. Comprueba la conexión que escribe en Figma; sin ella no dibuja nada y reporta el bloqueo con lo que falta: para una pantalla que se va a construir, con la ruta al plan, y para una petición cuyo entregable es el diseño, sin otra ruta que conectar Figma.
2. **Cargar antes de dibujar.** Las instrucciones propias de la conexión de Figma cuando las trae, y el sistema de diseño del proyecto como fija `hive-ui-criteria`.
3. **Resolver el archivo.** El declarado en «Design file» del `DESIGN.md`, el nombrado en la conversación, o el pedido una vez.
4. **Avisar los pasos manuales.** Lo que el usuario tendrá que hacer en Figma, como publicar una librería o aceptar su actualización, antes del primer dibujo.
5. **Dibujar.** Dos o tres ejemplos cuando la composición está abierta y uno cuando está fijada, a ancho de escritorio, con tokens, componentes, textos reales y volumen realista, en una página con el nombre del trabajo. Un hijo `hive-design-ui` cuando el host da esa conexión a sus hijos; su encargo lleva el archivo, la página, los anchos (escritorio y 440 px) y los nodos que este trabajo creó.
6. **Elegir.** Una decisión con una opción por ejemplo; se redibuja con cada corrección; el frame de 440 px se aprueba aparte cuando la pantalla se sirve en móvil. En la misma ronda, la pregunta del registro en `DESIGN.md` cuando el archivo no está declarado.
7. **Registrar.** `examples.md` en la carpeta del trabajo con la herramienta, la ubicación del archivo, el nodo de cada frame, la imagen local y el estado, actualizado a «approved» con su fecha; se quitan los ejemplos no elegidos.
8. **Cerrar.** Bloque de siguiente paso hacia `flow-plan`; o `flow-close` cuando el entregable era el diseño.

Recableado:

- `flow-research`: se quita la sección; los pasos 0, 3 y 5 y el párrafo inicial dejan de apuntarla; el paso 5 gana la recomendación de la etapa con sus condiciones, también en la terminación de un ticket evaluado.
- `global.md` «Questions and routes»: la lista de transiciones gana «research to Figma design, Figma design to plan», y la cláusula que limita cuándo se ofrece el comando de un flujo gana «a design in Figma».
- `flow-research` paso 0: la petición cuyo entregable es el diseño continúa hacia la etapa en el mismo turno, o reporta el bloqueo sin conexión a Figma.
- `flow-plan` paso 4 nombra dónde está el registro (`examples.md` en la carpeta del trabajo bajo `_support/workspace/`), para un plan que empieza en otra sesión o tras una compactación, y agrega la tarea que escribe «Design file» cuando el registro trae esa instrucción.
- `delegation.md`: el dibujo deja de estar bajo «in research» y gana un punto propio para la etapa.
- `flow-plan` pasos 4 y 5 de «UI decisions», `flow-verify/SKILL.md:12`, `delegation.md:11`, `design-record.md:32`, `unattended-delegation/SKILL.md:14` y `hive-design-ui.md:11`: sus enlaces pasan a la skill nueva y su redacción deja de decir que la investigación dibuja.

## Copy

El nombre visible de la etapa en el bloque de siguiente paso: «diseño en Figma» en español, «Figma design» en inglés.

## Impact

Lo que usa lo que cambia está listado en «Contexto verificado» de la propuesta. No toca producción ni datos. El comportamiento de `flow-build` y `flow-verify` frente al ejemplo aprobado no cambia: siguen leyendo el registro.

El tamaño de `content/guidance/global.md` crece en unos 80 bytes (las dos transiciones y un procedimiento más en la cláusula); el número exacto va en el pull request.

## Recovery

Revertir el commit de integración devuelve la sección a `flow-research` y borra la skill; un `hive update` posterior la retira de los hosts. No queda nada que no se pueda deshacer.
