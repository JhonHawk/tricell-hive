# Design

## Approaches

Solo un camino es sensato para el hueco central: la regla de la librería copia la forma que ya tiene la del archivo (`flow-figma-design` paso 3: «the Figma file the user named in the conversation, otherwise the one the app's `DESIGN.md` declares»), con un tercer origen, la detección. Lo que tenía más de una forma se preguntó:

- Quién escribe `DESIGN.md` (D8): la etapa al momento, o seguir esperando a un plan. Elegido: la etapa.
- La primera pantalla (D9): sin pausa con aviso, o con pausa para el visto bueno. Elegido: sin pausa.
- El modelo del revisor de UX (D10): el perfil `reasoning`, heredar la sesión, o el mismo modelo con más esfuerzo. Elegido: `reasoning`.

## Decisions

- **Una referencia nueva es el hogar de la regla de la librería.** `flow-figma-design/SKILL.md` está a 22 bytes de su tope, así que `references/design-library.md` lleva completa la regla: qué librería obliga, la detección, el inventario, dibujar con sus piezas, la pregunta de registro y la escritura en `DESIGN.md`. El `SKILL.md` conserva punteros con la condición de lectura en los pasos 3, 5, 6 y 7, y pierde la frase de la librería del paso 5 y la pregunta del archivo del paso 6.
- **La primera pantalla se queda en el `SKILL.md`**, en el paso 5, porque es orden de pasos y no regla de la librería.
- **La detección no pregunta salvo empate.** Con varias librerías candidatas gana la que instancian los marcos hermanos; solo cuando eso deja ninguna o más de una se pregunta una vez, en el texto.
- **La escritura en `DESIGN.md` se limita a dos campos.** Sin `DESIGN.md`, la etapa no lo crea: detectar librería, versión y capa de personalización es otro trabajo; el sí queda pendiente y el plan, o el usuario, lo escribe.
- **La edición se entrega al momento, como cambio mecánico.** La pregunta de entrega va en la misma ronda de aprobación cuando ninguna respuesta está vigente para ese repositorio, porque la guía global la pide antes de la primera edición; con el sí, la etapa escribe, muestra el cambio y lo entrega con `git-workflow` bajo esa respuesta, aparte de cualquier plan o build posterior. Así la edición no queda sin commit en un checkout del que un plan posterior no la movería a su rama.
- **Un solo hogar para la escritura.** La referencia nueva es el hogar de lo que la etapa pregunta, escribe y entrega; `design-record.md` «Format» y «Writing» solo apuntan a ella nombrando los dos campos.
- **La detección depende de lo que la conexión lee.** Cuando la conexión no lista las librerías habilitadas del archivo, la detección usa las instancias de los marcos hermanos y los registros de las hermanas.
- **El revisor recibe la librería fijada, no la declarada.** Su hallazgo deja de depender de `DESIGN.md`.
- **Una revisión más por trabajo de varias pantallas.** La primera pantalla a su ancho de escritorio, el conjunto completo con esa pantalla incluida porque las demás reusan sus piezas, y la relectura de lo corregido con las tarjetas: tres pasadas del revisor en vez de dos. Un trabajo de una pantalla no cambia.
- **El registro nace antes.** `examples.md` se crea con el inventario antes del primer dibujo; sus líneas de marcos y estados se escriben al dibujar, como hoy.

## Proposed shape

`content/skills/flow-figma-design/references/design-library.md` (nuevo), en este orden:

1. **Which library binds** — conversación, `DESIGN.md`, detección (librería habilitada en el archivo según la conexión; librería de la que los marcos hermanos instancian componentes o enlazan variables; librería que nombra el registro de una hermana). Un enlace a otro archivo con componentes cuenta como nombrada. Desempate y la única pregunta.
2. **Inventory** — líneas `uses: <pieza> · <componente o variable de la librería> · <nodo de la hermana>` y `searched: <pieza> · <término> · found <componente> | not found`, escritas en `examples.md` antes del primer dibujo.
3. **Draw with it** — la regla que hoy vive en el paso 5: instancia o enlace, pieza local solo con `not found`, línea `candidate:`, el archivo que no puede usar la librería.
4. **Record it** — la pregunta única por archivo y librería no declarados, la escritura de los dos campos, el caso sin `DESIGN.md`, la entrega de la edición, y la delegación desatendida (no pregunta, no escribe).

Líneas del registro que el revisor recibe: el inventario y las `candidate:`.

## Impact

- Quien dibuja: una lectura más de los árboles de las hermanas y búsquedas en la librería antes de dibujar; una pasada más del revisor cuando hay varias pantallas.
- Quien revisa: corre en un modelo mayor en Claude y Codex, también en la revisión de UX de cada pantalla construida que hace `flow-verify`, no solo en la de marcos de Figma: más tiempo y cuota en cada cambio de UI. En Pi sube el esfuerzo de medio a alto; en los demás hosts que heredan el modelo de la sesión no cambia nada.
- El usuario: una pregunta que ahora cubre archivo y librería; un archivo modificado en el repo de la app tras su sí.
- `flow-plan`: deja de añadir la tarea del archivo salvo un sí pendiente.
- Sesiones abiertas: cargan la guía nueva al reiniciar.

## Recovery

`git revert` del commit de merge y un nuevo refresco de la instalación local devuelven la guía y el perfil anteriores. Las ediciones que una etapa ya haya hecho en el `DESIGN.md` de una app quedan en el repo de esa app y se revierten ahí.
