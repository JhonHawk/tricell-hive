# Design

## Approaches

Solo una forma es sensata para la estructura: la que fijó `create-issue`, un `SKILL.md` corto con los pasos y una referencia por situación, que se lee solo cuando esa situación ocurre. Una skill sin referencias pasaría el límite de 16.000 bytes, y repartir sus pasos en otras skills dejaría el procedimiento sin un único hogar.

Lo que tenía más de una forma se resolvió como decisión del usuario (abajo).

## Decisions

- **D22 A** — Herramientas de documentos. Opciones: cualquier herramienta con Notion comprobada (elegida); solo Notion; Claude Docs primero. Se eligió la primera porque quien use otra herramienta no se queda sin skill y sirve en los siete hosts.
- **D23 A** — Congelar una versión. Opciones: copia en la misma herramienta (elegida); historial nativo; exportar `.md` y PDF siempre. La copia funciona igual en cualquier herramienta y se lee sin permisos especiales. Las mediciones confirmaron el costo anunciado: ninguna de las tres conexiones bloquea la copia, así que su inmutabilidad depende de que nadie la edite.
- **D24 A** — PDF. Opciones: navegador ya instalado (elegida); página imprimible que imprime el usuario; herramienta dedicada. Entrega el PDF sin instalar nada.
- **D25 A** — Ubicación del `.md` en un repositorio. Opciones: `_support/docs/requirements/<módulo>/` (elegida); `_support/sessions/` con fecha; raíz del repositorio.
- **D26 A** — Revisor. Opciones: rol nuevo `hive-review-prd` (elegida); ampliar `hive-review-plan`; hijo sin rol.
- **D30 A** — Entrada desde `flow-research`. Opciones: agregar la frase en sus pasos 0 y 5, los dos lugares donde está la de Figma (elegida); dejarlo sin cambios y confiar en la descripción de la skill. Sin la frase, un pedido dicho con otras palabras puede terminar en un documento escrito a mano. La frase sigue la forma de la que ya existe para Figma: una petición cuyo entregable es un documento de requerimientos sigue `create-prd` en el mismo turno. Reemplaza la decisión menor inicial de no tocar `flow-research`.
- **D31 A** — Alcance de la autorización de Notion: crear, renombrar y actualizar solo las páginas que la prueba crea.
- **D33 A** — Guía global. Opciones: nombrar en su lista de continuaciones el paso de research a `create-prd` (elegida); dejarla igual; que research se detenga con el comando. Reemplaza la decisión menor inicial de no tocar la guía global: dos lecturas independientes entendieron que el texto anterior prohibía el paso.
- **D35 A** — Presentación de la regla. Opciones: ficha por regla (elegida); tabla de tres columnas con la fuente y el veredicto en una lista aparte; lista con el detalle anidado. Reemplaza la línea o fila de tabla inicial, que el usuario rechazó al ver el ejemplo en Notion.

Decisiones menores, tomadas sin preguntar porque ninguna opción cambia el alcance, un contrato existente, datos guardados, el costo ni la reversibilidad:

- Los estados de una regla reutilizan los cuatro grupos con que Hive ya escribe reglas en un ticket, para que el PRD y el ticket usen las mismas palabras.
- El veredicto contra el código tiene cinco valores: `exists`, `new`, `contradicts`, `blocked`, `not checked`. «Falta» se parte en `new` (hay que construirla y nada lo impide) y `blocked` (necesita un dato, relación o permiso que no existe y el documento no prevé), porque en un PRD de algo nuevo casi todo falta y eso no es un problema.
- El PDF sale de un script determinista y no de HTML escrito por el agente, para que el PDF lleve exactamente el texto del `.md`.
- Donde la herramienta no tiene «duplicar» (Claude Docs), la copia es un documento nuevo creado con el contenido leído. Donde la conexión no edita en su lugar (el conector oficial de Google), cada versión es un documento nuevo.
- La copia congelada lleva en su título «no editar» y la skill nunca vuelve a escribirla; donde la herramienta ofrece bloqueo, la referencia lo nombra como paso opcional del usuario.
- El rol `hive-review-prd` usa el perfil de razonamiento, acceso de solo lectura y esfuerzo medio, como `hive-review-plan`.

## Proposed shape

### Archivos

```text
content/skills/create-prd/
  SKILL.md                         pasos; objetivo ≤ 9.000 bytes
  references/prd-format.md         secciones, línea de regla, estados, lista de comprobación, versiones
  references/incoming-document.md  estructurar un documento que llega y su cuestionario
  references/documents.md          dónde vive, operaciones por herramienta, congelar, modo .md y PDF
  references/technical-review.md   contraste con el código y paso a «listo para desarrollo»
  scripts/render_pdf.py            .md → HTML imprimible → PDF
content/agents/review/hive-review-prd.md
tests/skills/render_pdf_test.py
```

### Pasos de `SKILL.md`

1. **Elegir el uso**: redactar; documento que llega → `incoming-document.md`; actualizar con respuestas (versión nueva); revisión técnica → `technical-review.md`, solo con el repositorio.
2. **Leer antes de preguntar**: la fuente, el PRD existente con sus versiones congeladas, y con el código a mano, el código del módulo. Leer `documents.md` para saber dónde vive.
3. **Preguntar lo que falta**: tope de tres por ronda, con respuesta recomendada; una regla de negocio no decidida nunca se asume.
4. **Redactar** según `prd-format.md`.
5. **Comprobar** con la lista fija, en el mismo hilo.
6. **Revisión independiente**: un hijo `hive-review-prd`; lo que sostiene se aplica, y lo que necesita al dueño pasa a preguntas abiertas.
7. **Fijar el estado** según lo que la validación encontró.
8. **Escribir y congelar** según `documents.md`; fila nueva en el historial con la lista de cambios.
9. **Terminar**: enlace, estado, preguntas abiertas con quién las responde y supuestos; ofrece el siguiente paso (tickets con `create-issue`, diseño, plan) y se detiene.

### Ficha de una regla

Cada regla es una ficha propia, nunca una fila de una tabla (D35 A): un encabezado `R<n> · P0|P1|P2 · <estado>`, el texto de la regla en un párrafo propio, y una lista de dos puntos con la fuente y el veredicto contra el código. El historial de versiones es una lista, un punto por versión.

| Estado | Significa |
| --- | --- |
| decided | Su dueño la decidió; lleva quién y cuándo |
| assumption | Quien redactó la completó; vale mientras el dueño no diga otra cosa |
| to decide | Abierta; lleva quién la responde. Impide «aprobado por negocio» |
| cancelled | Reemplazada; lleva la versión y la regla que la reemplaza. La línea no se borra |

| Veredicto | Significa | Efecto |
| --- | --- | --- |
| exists | El código ya lo hace | Ninguno |
| new | No está en el código y nada lo impide | Ninguno |
| contradicts | El código hace otra cosa | Pregunta al dueño; impide «listo para desarrollo» hasta su respuesta |
| blocked | Falta un dato, relación o permiso que el documento no prevé | Igual que el anterior |
| not checked | Aún no se contrastó | Impide «listo para desarrollo» |

El veredicto se escribe en palabras que lee el autor; las ubicaciones del código van en la sección de revisión técnica, al final del documento.

### Estados del documento

```text
draft ──► iterating ──► approved by business ──► technical review ──► ready for development
            ▲                    │                       │
            └── corrección = versión nueva ◄─────────────┘
```

| Paso a | Lo permite |
| --- | --- |
| iterating | El documento se entregó a su autor con preguntas abiertas |
| approved by business | Lista de comprobación sin fallos, revisión independiente hecha, ninguna regla `to decide`, y aprobación explícita del usuario |
| technical review | Una sesión con el repositorio toma una versión aprobada por negocio |
| ready for development | Ninguna regla `not checked`; cada `contradicts` y `blocked` con la respuesta de su dueño |

### Operaciones por herramienta (`documents.md`)

| Operación | Notion | Google Docs | Claude Docs |
| --- | --- | --- | --- |
| Crear | Página con contenido Markdown bajo un padre | Archivo creado desde Markdown | Documento desde Markdown |
| Copia congelada | Duplicar, releer hasta que esté completa y renombrar | Copiar con nombre nuevo | Leer o exportar y crear otro |
| Editar la versión viva | Actualizar la página | Solo por la API; con el conector oficial, cada versión es un archivo nuevo | Actualizar el documento |
| Comentarios | Sí | Solo por la API | Sí |
| Bloqueo | Solo a mano en la aplicación | Solo por la API | Solo a mano (acceso de lectura) |
| Enlace | El de la respuesta al crear o duplicar; si no viene, releer la página | El identificador del archivo creado o copiado | El enlace que devuelve la creación |

Las filas de Google Docs vienen solo de su documentación y salen sin probar; la referencia lo dice.

Destino: antes de la primera escritura la skill usa el destino que el usuario nombró; si no nombró ninguno, pregunta dónde va, y donde la herramienta tiene borradores privados ese es el destino recomendado. Nunca escribe sobre un documento que no creó sin que el usuario lo señale.

La referencia nombra la operación de cada herramienta y remite a la ayuda actual de su CLI o conexión para el comando exacto, como hace `create-issue` con los trackers.

### Script de PDF

`render_pdf.py <prd.md> [--out <file.pdf>]`, solo con la biblioteca estándar de Python:

- convierte el subconjunto de Markdown que usa el formato del PRD (títulos, párrafos, listas, tablas, énfasis, código en línea, enlaces) en una página HTML con hoja de impresión A4;
- busca un navegador de la familia Chromium (Chrome, Edge, Chromium o Brave, por sus rutas habituales en macOS, Linux y Windows y por el `PATH`) y lo ejecuta sin ventana con `--print-to-pdf`;
- `--browser <ruta>` usa ese ejecutable y omite la búsqueda; una ruta que no existe cuenta como «sin navegador», lo que deja probar ese caso en una máquina que sí tiene uno;
- sale con 0 y la ruta del PDF, y borra la página intermedia, que escribe en la carpeta temporal del sistema; sin navegador, deja el `.html` junto al `.md`, sale con un código distinto de cero y escribe cómo imprimirlo;
- una construcción de Markdown que no reconoce sale como texto literal, nunca se omite;
- trata el `.md` como texto que no es de confianza, porque puede venir de un documento ajeno: escapa todo el texto de origen, de modo que el HTML que traiga se ve como texto; no emite imágenes ni otros recursos remotos; la página lleva una política de contenido que prohíbe scripts y cargas de red; y lanza el navegador con un directorio de perfil temporal que borra al terminar.

En el modo `.md`, cada entrega del documento vivo a su autor lleva `prd.pdf` al lado, y cada versión congelada su `prd-v<N>.pdf`. El PDF es un derivado que se regenera desde el `.md`: la referencia dice que se genera para compartir y que no hace falta versionarlo.

### Archivo en el repositorio

`_support/docs/requirements/<módulo>/prd.md` es un artefacto de soporte, como un reporte: la skill lo escribe sin rama, commit ni entrega, informa su ruta y que no está confirmado en Git, y solo lo confirma cuando el usuario lo pide.

### Rol `hive-review-prd`

Lee el PRD completo y su fuente; comprueba cada regla (un solo comportamiento, observable, sin términos vagos, sin dos lecturas), busca contradicciones entre reglas, casos sin cubrir (errores, permisos, estados vacíos, límites) y criterios que nadie puede observar; con el repositorio a mano, vuelve a contrastar una muestra de veredictos. Devuelve hallazgos con prioridad P0 a P3 y no edita el documento, el tracker ni la memoria.

Su encargo lleva el texto actual del documento, o la ruta de un `.md` local con él, nunca solo su enlace, porque el hijo puede no alcanzar la conexión de la herramienta de documentos; su fuente (la petición, el documento que llegó, las respuestas del autor) y el resultado de la lista de comprobación; nunca el razonamiento de quien redactó. El paso 6 de la skill enuncia esos tres insumos y remite a la regla del encargo de un revisor para lo que no se incluye.

## Impact

Cambia un comportamiento existente: `flow-research`, en su paso 0, envía a la skill las peticiones de un requerimiento; su archivo pasa de 10.245 bytes a unos 10.800, bajo el límite de 16.000. También cambia la guía global: su lista de continuaciones en el mismo turno nombra la de research a `create-prd` (D33 A), y `_support/docs/architecture/compaction-survival.md` repite esa lista. Lo demás es nuevo. `create-issue` ya lee «the requirement document with the rule it comes from» y sigue igual. Tocan producción solo en el sentido de que `hive update` los instala en cada host registrado.

## Recovery

Revertir el pull request quita la skill, el rol y las páginas; `hive update` los retira de la instalación. Las páginas que una sesión real haya creado en Notion se mueven a la papelera desde Notion.
