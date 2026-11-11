# Diseño

## Enfoques

Para dibujar hay un solo camino sensato: el hilo principal, por script y con un solo ejemplo, porque es lo que el usuario decidió y lo que produjo frames exactos en FSR-856. La alternativa de dar mejores instrucciones al subagente que dibuja no resuelve que el subagente arranque sin el contexto de la sesión, que fue la causa de que las decisiones se perdieran entre tickets.

Para evaluar se presentaron tres caminos (D1): ampliar `hive-review-ux` con un modo para frames (elegido: es de solo lectura, ya revisa contra requisitos y no suma un rol que desplegar en 7 hosts), un rol nuevo `hive-review-design` (rol exclusivo, ~4 archivos más y un rol más que mantener), o convertir `hive-design-ui` en revisor (tendría permisos de escritura y mezclaría diseñar con evaluar).

## Decisiones

- El dibujo queda en el hilo principal aunque sus lecturas sean grandes: la regla de costo de la delegación lo excluye de forma explícita, porque la calidad depende del contexto de la sesión.
- Un ejemplo por pantalla. Las pantallas hermanas (las del mismo módulo o navegación en el archivo y en el código) fijan el marco, el encabezado, la barra de filtros, la tabla, las acciones de fila y si un alta es página o diálogo. Alternativas: dos o tres, solo para la pantalla que el usuario nombra al pedirlas; el resto del flujo se dibuja una vez.
- Método por script, escrito por capacidad y no por nombre de herramienta: cuando la conexión ejecuta scripts en el archivo, cada escritura define una vez las piezas (contenedor con el layout automático de la herramienta, botón, campo, fila de tabla, badge, el marco de la app clonado de una pantalla existente) ligadas a las variables del archivo, y compone las pantallas con ellas; una corrección cambia la pieza o los datos y vuelve a dibujar ese frame.
- Registro y revisión: `examples.md` se escribe en cuanto se dibuja, con estado `drawn, under review`, para que los nodos creados queden registrados aunque haya una compactación; después, un `hive-review-ux` revisa el primer dibujo, y otro el frame de 440 px. Su brief se define en «Reviewer brief» de la regla de delegación (un solo lugar): nodos e imágenes de los frames, la petición, los nodos de las pantallas hermanas, las decisiones registradas y el ancho, sin el razonamiento de quien dibujó. El hilo principal corrige lo que se sostiene y pide una relectura solo de lo corregido, una vez; el estado pasa a `drawn, pending the user's approval` y lo que quede va en el mensaje al usuario. Una corrección que pide el usuario no recibe otra revisión: el usuario la ve en ese momento.
- Decisiones previas: se leen de **Decisions** en `DESIGN.md` y de las secciones `UI` de las carpetas de cambio del proyecto, archivadas incluidas; `_support/workspace/` no sirve como fuente porque está ignorado y su retención es local.
- Modo de frames del revisor: lo selecciona un brief con frames de un archivo de diseño. Lee los frames por la conexión, compara contra las hermanas y las decisiones, y mide la alineación (bordes compartidos, espaciados, marco idéntico a las hermanas, nodos fuera del layout automático o en posición absoluta donde las hermanas no los usan, texto que no cabe). No aplican: manejar el navegador, los estados de carga, vacío y error, el orden de teclado y foco, la auditoría automática, la lista de defectos de la base ni la captura para la guía. El veredicto pasa cuando los frames no muestran defectos frente a lo pedido, las hermanas y las decisiones. Sin conexión que lea el archivo, juzga por las imágenes del brief y reporta las medidas de alineación como no corridas; la etapa lo dice en el mensaje de aprobación. No verificado: que un subagente de solo lectura reciba lecturas de Figma en cada host (Claude Code, Codex, los demás); el respaldo por imágenes cubre el caso en que no.
- La descripción del rol `hive-review-ux` pasa a cubrir también los frames de un archivo de diseño, para que el enrutamiento lo encuentre.
- El título del requisito "Screen proposals in Figma before the plan" se conserva a propósito: renombrarlo rompería la fusión del delta, y el texto ya habla de ejemplos.
- `hive-design-ui` deja de recibir páginas de Figma: se quitan esas instrucciones de su rol porque ningún flujo lo manda a dibujar en un archivo de diseño. Sigue sirviendo a `flow-plan` para resolver composiciones en texto.
- Convenciones (D3): una corrección del usuario que aplica más allá de la pantalla se pregunta una vez en la ronda de aprobación; el sí se guarda en el registro y el plan agrega la tarea que la escribe en **Decisions** de `DESIGN.md`, con fecha y origen (el registro del ejemplo).

## Impacto

Los textos cambiados los consumen los agentes que cargan `flow-figma-design`, `flow-research`, `flow-plan`, `hive-ui-criteria` y la regla de delegación, y los roles `hive-review-ux` y `hive-design-ui` en todos los hosts tras `hive update`. Las sesiones abiertas siguen con la versión anterior hasta reiniciarse. No toca producción, la guía global ni el binario.

## Recuperación

Revertir el merge del PR y volver a desplegar con `hive update`. No hay datos que recuperar.
