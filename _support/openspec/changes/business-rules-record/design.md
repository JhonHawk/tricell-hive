# Design

## Approaches

1. **Solo `flow-figma-design`.** La skill dibuja las tarjetas y escribe el ticket. Descartado: en la sesión de FSR-859, 14 de 20 reglas nacieron en plan y build, y un trabajo sin pantalla nunca pasa por la etapa.
2. **Regla común con escritura solo al cierre.** Una escritura por trabajo. Descartado por D1 A: el ticket quedaría atrasado durante el build, cuando otras sesiones lo leen.
3. **Regla común con escritura en cada aprobación y al cierre (elegido, D1 A).** Una referencia única; cada flujo lleva un puntero en el paso que la ejecuta. Reutiliza lo que ya existe: las líneas `rule:` de `examples.md`, las decisiones de `proposal.md` y los dos momentos en que Hive ya escribe el tracker (arranque del build y cierre).

## Decisions

- **Hogar de la regla común:** `content/skills/flow-plan/references/business-rules.md`. `flow-plan` es dueño del registro de decisiones del trabajo y su `SKILL.md` tiene espacio; `flow-close` (10 bytes libres) no puede serlo.
- **Hogar de las tarjetas:** `content/skills/flow-figma-design/references/rule-cards.md`, porque `SKILL.md` tiene 544 bytes libres. La definición de regla sale del paso 5 hacia la referencia común y deja un puntero, lo que libera unos 250 bytes.
- **Registro único por trabajo:** las líneas `rule:` viven en el registro del ejemplo (`examples.md`, no el `DESIGN.md` de la app) mientras el trabajo no tiene carpeta de cambio; el plan las copia a una sección `## Business rules` de `proposal.md`, que desde entonces gobierna (hoy `flow-plan` las escribe en la sección `UI` de `design.md` y el paso 7 de `flow-figma-design` manda copiar el registro entero ahí; ambos textos cambian para que las reglas tengan un solo lugar). La sección `UI` conserva el resto del registro del ejemplo: archivo, nodos, estados, tarjetas y candidatos. Si la etapa de diseño corre de nuevo con la carpeta ya escrita, anota en `examples.md` y la revisión del plan copia las líneas cambiadas. Un plan en la conversación usa su archivo de checkpoint, igual que un trabajo sin diseño ni plan.
- **El encabezado `## Business rules` se queda en inglés** en cualquier idioma de sesión, como `## Effects`, porque la referencia lo busca por ese nombre; el comentario de la plantilla lo dice.
- **Permiso para corregir tarjetas desde plan y build (D3).** La guía global lo declara junto al permiso del ticket: cubre solo cambiar nodos de tarjeta que el trabajo creó; nunca un frame ni un nodo ajeno.
- **Excepción a la regla de efectos críticos (D2 A).** Editar la descripción cambia datos que la tarea no creó en un sistema no local, que la guía global trata como efecto crítico con pregunta propia. La regla del ticket declara que la sección de reglas y su comentario no llevan pregunta, y `outside-the-repository.md` las nombra junto a los movimientos del ticket para que ningún plan las pida como clase de efecto.
- **Cierres que no escriben.** La escritura del cierre ocurre en el punto de finalización del trabajo, como lo define `flow-close` (el merge, o el pull request abierto bajo `ready`), para no depender de un archivo de checkpoint que la limpieza ya borró. Un trabajo abandonado no escribe. En esos casos y sin ticket, el reporte lista las reglas.
- **Momento «plan aprobado» = arranque del build.** La respuesta que aprueba el plan es la que arranca `flow-build`, y ahí Hive ya mueve el ticket a en curso: la escritura de reglas va en ese mismo paso.
- **Tarjetas después de la primera revisión.** El revisor hoy no recibe las reglas del dibujante para encontrar por su cuenta las que nadie resolvió. Las tarjetas se dibujan después de conciliar esa primera revisión y se revisan en la segunda pasada, que hoy solo corre si hubo correcciones y ahora corre también cuando hay tarjetas. Costo: una pasada de revisor más por etapa de diseño (unos 6 minutos en las sesiones observadas).
- **Research no escribe el ticket.** Sus decisiones entran al registro y llegan al ticket en el siguiente momento (diseño o arranque del build); `flow-research` sigue sin cambiar el tracker.
- **Un hijo nunca escribe el ticket ni Figma.** Reporta la regla que tuvo que asumir; el hilo principal la registra como `assumption` y la dice en su siguiente mensaje.
- **Tarea operativa.** La escritura de reglas ocurre en una petición cuyo entregable es el diseño cuando nombra su ticket; en las demás tareas operativas solo si el usuario la pide, como hoy mover o comentar.

## Proposed shape

### Línea del registro

`rule: <the rule> · screen <name, or none> · <status> · found by <drawer | reviewer | research | plan | build | verify>`

Estados:

- `proposed`, `assumption`, `scope`, `approved <date>: <answer>` — los actuales.
- `stated: <source>` — regla que la solicitud, el ticket o la documentación ya dicen; solo se anota en el registro de diseño, para la tarjeta, cuando el texto del frame no la dice. No va al ticket.
- `cancelled <date>: <reason> · replaced by <the rule that replaces it>` — la línea anterior se conserva; nunca se borra una línea.

Líneas de control en el mismo registro: `ticket section: <key> · <heading>`, `ticket write: <date> · <moment>`, `card: <screen> · <node>`, `card stale: <screen> · <rule>`, `frame stale: <node> · <rule>`.

### Sección del ticket

- Un encabezado que empieza con «Business rules ·» en el idioma del ticket, igual en todas las escrituras del trabajo y guardado en la línea `ticket section:` para encontrarlo.
- Grupos, solo los que tengan filas: decididas (la regla en palabras llanas y su fecha; nunca un ID solo), supuestos, por decidir y canceladas (la regla anterior citada, dónde estaba, el motivo y qué la reemplaza).
- Contiene las reglas que el ticket no decía y las cancelaciones de las que sí decía; no repite las reglas propias del ticket.
- Antes de escribir se lee la descripción; se reemplaza solo la sección y el resto queda byte a byte.
- Cada escritura agrega un comentario: una línea por regla agregada y, por cada cancelación, qué se canceló y por qué.
- La sección va entre su encabezado y una línea de cierre propia (el mismo texto más ` · end`); si falta alguna de las dos o se repite, no se edita la descripción y la sección va en el comentario. Si la sección que se escribiría es igual a la que el ticket ya tiene, no se escribe. Sin ticket, no se escribe y el reporte del cierre lista las reglas. Con tickets espejo, se escribe en cada uno en su tracker. Si la herramienta no puede editar la descripción, la sección completa va en el comentario y se reporta.
- Nunca un valor secreto ni rutas internas del repositorio.

### Tarjeta de reglas

- Un componente local del archivo, creado por el trabajo, fuera de los límites del frame, a su derecha, con un encabezado que dice que es anotación y no interfaz.
- Una fila por regla de la pantalla cuyo condición o consecuencia no está escrita en el texto visible del frame, con su origen: especificación (`stated`), decisión (`approved`), supuesto (`assumption`) o por decidir (`proposed`).
- Se dibuja después de conciliar la primera revisión; la segunda pasada del revisor la cubre.
- Al responderse una regla, su fila se actualiza en el mismo turno (paso 7).
- No se mueve un frame que el trabajo no creó para hacerle lugar: se pregunta antes.
- En una petición cuyo entregable es el diseño y que nombra frames existentes, las tarjetas van junto a esos frames.

### Dónde va cada puntero

| Archivo | Qué cambia |
| --- | --- |
| `content/guidance/global.md` «Scope and authorization» (`:21`) | La regla autoriza también la sección de reglas y su comentario, con sus tres momentos y el puntero, sin pregunta aunque cambien datos que la tarea no creó, y corregir las filas de tarjetas que el trabajo creó; la frase de la tarea operativa nombra la excepción del diseño. `:113` y `:111` cubren esa escritura en `Other trackers`. |
| `flow-plan/references/outside-the-repository.md` (`:3`, `:19`) | La sección de reglas y su comentario se nombran junto a «the ticket's moves». |
| `flow-research/SKILL.md` paso 1 | «ticket» pasa a nombrar sus comentarios, también cuando un hijo lo lee. |
| `flow-figma-design/SKILL.md` | Línea 8 (qué escribe la etapa), paso 5 (definición → puntero; tarjetas tras la revisión), paso 6 (segunda pasada cubre tarjetas), paso 7 (fila de la tarjeta, escritura al ticket y destino de la copia: las líneas `rule:` van al registro de reglas del plan), «A request for the design itself», «Unattended delegation». Lo que no quepa en 16,000 bytes pasa a `rule-cards.md` detrás del puntero. |
| `flow-plan/SKILL.md` | «UI decisions» paso 4 (la frase «write them beside the example's record in the `UI` section» pasa a mandar las líneas a `## Business rules` de `proposal.md`), «Change folder» (`proposal.md` lista las reglas), «Revise a plan in progress» (una regla reemplazada sigue la referencia). |
| `flow-plan/assets/change-folder/proposal.md` | Sección `## Business rules`, encabezado en inglés, con su comentario de formato. |
| `flow-build/SKILL.md` | Arranque: junto a mover el ticket. Decisión o supuesto a mitad del build: puntero. |
| `flow-verify/SKILL.md` | Una corrección del usuario en la validación que decide una regla: puntero. |
| `flow-close/SKILL.md` | Paso 3 gana la frase « Close its business rules as [business rules](../flow-plan/references/business-rules.md) sets.» (94 bytes). Para que quepa, el paso 5 deja de enumerar la limpieza y apunta entera a «Preservation» de la guía global («and stop and remove only what the global guidance's «Preservation» authorizes at a work item's close»), tras el hallazgo de la revisión de código de que el recorte planeado era una repetición parcial. El recorte planeado cambiaba «stop the processes and containers this work started, never a program other sessions may use, and remove its reproducible temporary files and checkpoint file, the containers it started with their volumes and the images it built, as the global guidance's «Preservation» sets and authorizes» por «stop the processes and containers this work started and remove its temporary files, checkpoint file, containers and images, within the limits the global guidance's «Preservation» sets and authorizes» (89 bytes menos). Los límites que salen (programas de otras sesiones, archivos reproducibles, volúmenes, imágenes que construyó) siguen enteros en `global.md` «Preservation», a la que el paso ya apunta. El archivo queda en 15,897 bytes. |
| `hive-review-ux.md`, `delegation.md` «Reviewer brief» | Las tarjetas entran solo al brief de la segunda pasada; qué revisa en ellas. En `hive-review-ux.md:10`, «as step 5 of flow-figma-design defines one» pasa a apuntar a `skill:flow-plan/references/business-rules.md`. |

## Impact

- Todos los proyectos que usan Hive con ticket reciben una escritura nueva en la descripción del ticket, hasta tres veces por trabajo, más un comentario por escritura.
- `flow-plan` deja de escribir las líneas `rule:` en la sección `UI` de `design.md`: las carpetas de cambio abiertas en otros proyectos las conservan ahí; el build las lee donde estén.
- La etapa de diseño suma una pasada de revisor cuando hay tarjetas.
- `content/guidance/global.md` crece; el pull request declara el cambio de tamaño.
- No cambia producción ni el manager.

## Recovery

Revertir el commit del cambio y desplegar de nuevo con `hive update --source`. Las secciones ya escritas en tickets de otros proyectos no se revierten con eso: el historial de cada tracker las deshace.
