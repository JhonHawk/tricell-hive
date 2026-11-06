# Diseño

## Enfoques

Se presentaron tres el 2026-11-05:

- **Skill propia (elegido, D1 A).** Un `SKILL.md` con el procedimiento (detección, precedencia, registro de diseño) y referencias que se cargan según la necesidad. `flow-plan` y los roles de diseño, construcción y revisión solo la enlazan, y el usuario la puede invocar directamente. Es una sola casa canónica, como pide `instruction-resources.md`. Costo: una skill más en el catálogo.
- **Referencias dentro de `flow-plan`.** No suma skill, pero los roles de construcción y revisión leerían archivos del plan, y no se podría invocar sola.
- **Ampliar los roles y `ui-verification.md`.** No suma archivos, pero duplica el criterio en tres lugares y el hilo principal del plan no lo leería.

## Decisiones

Las decisiones D1–D6 están en `proposal.md`. Supuestos menores (no cambian alcance, contrato, datos guardados, costo ni reversibilidad):

- La atribución va en `references/notice.md`, sin enlazarla desde `SKILL.md` para que no se lea en cada uso, siguiendo el precedente de `test-audit/references/license-openclaw.md`. Nombra solo a Impeccable (metodología, Apache-2.0) y a Google DESIGN.md (formato), como fija D11 B. Como el texto es original y no se copia ningún pasaje, no hace falta incluir la licencia: las condiciones de redistribución de la licencia solo aplican al material adaptado.
- Umbrales con el estándar que los fija:
  - 24×24 px CSS de objetivo táctil, con las excepciones de 2.5.8 (AA), como mínimo;
  - 44 px como valor por defecto de la skill cuando el tamaño de la librería lo permite, sin que su ausencia sea defecto;
  - reajuste a 320 px CSS de ancho (1.4.10, AA) y texto al 200 % (1.4.4, AA);
  - la longitud de línea la fija la librería o el proyecto, no la skill.
- **Valores por defecto de la skill.** Son solo los valores de guía de `criteria.md` que no son mínimos, por ejemplo los 44 px de objetivo táctil. Por eso ocupan el último nivel de la precedencia.
- Lo que juzga la revisión sigue teniendo su casa en el paso 2 de `hive-review-ux.md`, que pasa a apuntar a los checks medidos de la skill. `ui-verification.md` solo suma el ancho de 320 px para el check de reajuste, y `delegation.md` «Reviewer brief», los estados que esos checks necesitan.

## Forma propuesta

### Archivos de la skill

```
content/skills/hive-ui-criteria/
├── SKILL.md                    # procedimiento y tabla de carga de referencias
└── references/
    ├── design-record.md        # detección, precedencia, formato y ciclo de vida del DESIGN.md
    ├── criteria.md             # modos, mínimos, colecciones, textos, casos extremos
    ├── refinement.md           # simplificar, endurecer, pulir; clasificación de desviaciones
    └── notice.md               # atribución (no enlazada desde SKILL.md)
```

Cada archivo pesa como mucho 16 000 bytes (`tests/content/skill_size_test.go`). Todo se escribe en inglés y para el modelo menos capaz que lo ejecuta: criterios con umbral o una prueba observable, sin palabras como «claro», «limpio» o «consistente» sin definir.

### `SKILL.md`

- **Frontmatter.** `name: hive-ui-criteria`. La descripción dice cuándo usarla: diseñar, construir, refinar o revisar una pantalla web o de producto en un proyecto con o sin librería de componentes, y crear o refrescar el DESIGN.md del proyecto. Las interfaces de terminal no entran.
- **Pasos.**
  1. Buscar el DESIGN.md en la raíz de la app dueña de la UI. Si existe, leerlo primero y comprobar que no esté desactualizado, como dice `design-record.md`.
  2. Si falta o está desactualizado, detectar librería, versión instalada, capa de personalización y herramientas de la librería con el procedimiento de `design-record.md`.
     - Durante un plan, `flow-plan` agrega la tarea que lo escribe o refresca (D4 A).
     - Fuera de un plan, escribirlo solo cuando el usuario lo pide.
     - Mientras no exista, usar lo detectado sin guardarlo.
  3. Aplicar la precedencia (D9 A). Esta sección es la casa canónica de la regla:
     1. un requisito de accesibilidad que declara el proyecto (en el DESIGN.md o en su guía) o el ticket;
     2. las decisiones registradas del proyecto;
     3. la capa de personalización;
     4. los valores por defecto de la librería;
     5. los valores por defecto de la skill.

     Los mínimos de accesibilidad de `criteria.md` son defecto en lo que el cambio construye o personaliza. Un valor por defecto de la librería sin modificar que falla un mínimo no es defecto del cambio: va a la lista de defectos que ya existían, con su motivo. Cuando hay un requisito de accesibilidad declarado, los mínimos se aplican a todo.

     Un arreglo que altere un token, o el tamaño, radio, espaciado o movimiento de un componente de la librería, vuelve como decisión pendiente. Esta regla hoy la repiten `hive-design-ui` y `hive-build-frontend`, que pasan a apuntar aquí.
  4. Decidir el modo de cada pantalla con `criteria.md`: operar por defecto en apps con librería; leer para documentación; persuadir solo para páginas públicas.
  5. Cargar solo las referencias que pide la tarea, según una tabla de situaciones: crear o cambiar una pantalla, revisarla, refinarla, o crear o refrescar el DESIGN.md.
- **Contrato de salida.** Las decisiones y su justificación según el registro, lo que se detectó con sus rutas, los checks que de verdad se hicieron y lo que no se pudo verificar.

### `design-record.md`

- **Detección (para cualquier librería).**
  1. Encontrar el manifiesto de paquetes de la app dueña de la UI; en un monorepo, la app que importan las pantallas que cambian.
  2. Identificar la librería de componentes por sus dependencias y sus archivos de configuración, por ejemplo `components.json` en shadcn y ReUI, el plugin de Tailwind en HeroUI v2 o `providePrimeNG` en PrimeNG. Si hay más de una, vale la que importan las pantallas que cambian.
  3. Tomar la versión instalada del lockfile (`package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `bun.lock`) o del `package.json` instalado del paquete, nunca del rango del manifiesto. En las librerías que copian el código al proyecto (shadcn, ReUI), registrar el estilo y el registro de `components.json` y las versiones de las primitivas que usan.
  4. Encontrar la capa de personalización y registrar sus rutas:
     - el preset o tema, por ejemplo `definePreset` en PrimeNG v20 o el plugin `heroui()` en HeroUI v2;
     - las variables CSS, `@theme` o el tema de Tailwind, por ejemplo las variables CSS sobre `@heroui/styles` en HeroUI v3;
     - las sobreescrituras globales, como `pt` o `cssLayer` en PrimeNG;
     - los componentes envoltorio propios y el código de componentes copiado.
  5. Si están instalados el servidor MCP o la skill de la librería, usarlos antes de leer archivos.
  6. Leer la documentación de esa versión (Context7, o la oficial) antes de afirmar cualquier API o token.
- **Formato del DESIGN.md (D3 A).**
  - Frontmatter YAML de Google: `version: alpha`, `name`, `description`, y `colors`, `typography`, `rounded`, `spacing` y `components` solo con valores encontrados.
  - Secciones de prosa en el orden de Google: Overview, Colors, Typography, Layout, Elevation & Depth, Shapes, Components, Do's and Don'ts. Se pueden omitir; las que estén van en ese orden.
  - Una sección final `## Hive`, que el formato conserva porque guarda las secciones desconocidas, con:
    - **Library:** nombre, versión instalada y archivo de donde salió.
    - **Customization layer:** tipo y ruta de cada pieza.
    - **Library tools:** MCP o skill, si existen.
    - **Product:** usuarios, tareas principales, modo de cada área y restricciones; es el registro de producto (H10).
    - **Decisions:** desviaciones aprobadas respecto a la librería, con fecha y origen. Son las «decisiones registradas» que la revisión respeta.
    - **Sources:** cada archivo fuente con su huella (`git hash-object <ruta>`). Los lockfiles no llevan huella, porque cambian con cualquier dependencia; su lectura cubre solo la versión.
    - **Watch:** los globs de la capa de personalización que se miran para detectar piezas nuevas, por ejemplo `src/styles/**/*.css` o `src/app/theme/**`.
  - Cada valor del frontmatter y de la prosa nombra su archivo fuente. Ningún valor se inventa. Si no hay sistema de diseño, se dice así y se propone uno mínimo como propuesta, nunca como tokens del proyecto.
  - El documento lleva un ejemplo completo en un bloque de código.
- **Ubicación (D2 A).** En la raíz de la app dueña de la UI, junto a su manifiesto; uno por app en un monorepo.
- **Desactualizado.** El chequeo no repite la detección: compara solo lo registrado. Está desactualizado cuando:
  - la versión que da el lockfile difiere de la registrada;
  - la huella de un archivo de **Sources** cambió o el archivo ya no existe;
  - un glob de **Watch** encuentra un archivo que no está en **Sources**.

  En ese caso hay que nombrar qué cambió y proponer el refresco. Nunca se usan en silencio los tokens desactualizados.
- **Escritura.** Nunca se sobreescribe sin más: el cambio se muestra como diff dentro de la tarea del plan o del pedido del usuario. Si el `lint` de `@google/design.md` ya está instalado, se corre y sus errores se corrigen. Nunca se instala ni se ejecuta con un descargador de paquetes de un solo uso.

### `criteria.md`

Cada check lleva su umbral o su observación y una marca **[measured]** cuando la revisión UX lo juzga como defecto (D5 A). Lo que no lleva la marca es guía de diseño.

- **Modos de pantalla** (sin marca).
  - Operar: escaneo, consistencia y controles estándar por encima de la expresión; el estilo del producto solo aporta tipografía, paleta, densidad y un rasgo propio, nunca el layout ni los controles.
  - Leer: una columna tranquila, con la longitud de línea que fija el proyecto.
  - Persuadir: la oferta y una acción que funcione en la primera pantalla.
- **Mínimos de accesibilidad** (todos **[measured]**, con el alcance que fija la precedencia):
  - el estado nunca se indica solo con color (WCAG 1.4.1, A);
  - objetivo táctil de 24×24 px CSS o una de las excepciones de 2.5.8 (AA);
  - reajuste a 320 px CSS (1.4.10, AA) y texto al 200 % (1.4.4, AA) sin perder contenido ni funciones;
  - el foco no queda tapado (2.4.11, AA);
  - contraste no textual de 3:1 en el indicador de foco y en un estado que se transmite solo por ese contraste (1.4.11, AA);
  - el movimiento que dispara una interacción y no es esencial respeta `prefers-reduced-motion` (2.3.3, **AAA**; D10 A);
  - el foco vuelve al elemento que abrió el overlay al cerrarlo, salvo que ya no exista (guía WAI-ARIA APG del diálogo modal, con 2.4.3; D10 A).

  El contraste de texto, el orden de teclado y los nombres accesibles ya los mide la revisión y no se repiten.
- **Colecciones.**
  - **[measured]**:
    - el alcance de la búsqueda está nombrado;
    - hay un conteo de resultados;
    - los filtros activos se ven y se pueden quitar uno a uno;
    - hay un control para reiniciar los filtros;
    - «sin resultados» conserva la búsqueda y ofrece cómo ampliarla;
    - un catálogo vacío es un estado distinto, con una acción;
    - cuando la app enruta, la búsqueda, los filtros, el orden y la página viven en la URL;
    - al volver de un detalle se restauran la página, el scroll y el foco.
  - Guía: tabla, tarjetas o lista según la tarea de comparar o de escanear.
- **Textos.**
  - **[measured]**:
    - la etiqueta de un botón nombra la acción o su resultado;
    - una confirmación destructiva nombra la acción, nunca «OK» o «Sí» a secas;
    - un error dice qué falló y cómo recuperarse;
    - un placeholder nunca es la única etiqueta;
    - en el código que llega a producción, ningún dato de demostración se presenta como real. Los datos de prueba o sembrados de un walkthrough no cuentan.
  - Guía: la etiqueta anticipa el resultado.
- **Casos extremos.**
  - **[measured]**:
    - con más de 1000 elementos, una lista pagina o virtualiza: el DOM nunca muestra todos los elementos a la vez;
    - mientras un envío está pendiente, su control queda deshabilitado y solo sale una petición en el registro de red.
  - Guía: los gestos interrumpidos.

  Los valores largos y el texto un 30–50 % más largo ya los cubre la revisión.

### `refinement.md` (sin marca)

- **Simplificar.** Una acción primaria por región, los estados que de verdad se alcanzan y quitar la decoración que no lleva información.
- **Endurecer.** Errores, permisos, datos vacíos, extremos e interrupciones.
- **Pulir.** Recorrer el camino completo y clasificar cada desviación del registro de diseño como:
  - **falta un token:** se propone agregarlo al registro;
  - **caso suelto:** se corrige hacia el token existente;
  - **no encaja con el concepto:** vuelve como decisión pendiente;
  - **defecto local:** se corrige.

### Consumidores (solo punteros; ninguno repite pasos de la skill)

- `flow-plan/SKILL.md` «UI decisions», paso 1:
  - cargar el sistema de diseño con [hive-ui-criteria](../hive-ui-criteria/SKILL.md);
  - cuando el DESIGN.md falta o está desactualizado, agregar una tarea que lo escribe o refresca, que se entrega con el cambio.
La precedencia y la regla de que tocar el tema o un componente de la librería vuelve como decisión pendiente pasan a tener su casa canónica en la sección de precedencia de la skill. Los roles que hoy las repiten las reemplazan por un puntero a esa sección.

- `hive-design-ui.md`:
  - paso 1: la guía del proyecto es el DESIGN.md que define [hive-ui-criteria](skill:hive-ui-criteria/SKILL.md), y la frase sobre desviaciones del tema se cambia por el puntero a la precedencia;
  - paso 2: aplicar `criteria.md` y `refinement.md` en lugar de los adjetivos actuales.
- `hive-build-frontend.md`:
  - paso 1: la frase «change no token, edit no library component…» se cambia por el puntero a la precedencia;
  - paso 3: la «project UI guide» pasa a ser el DESIGN.md, con el enlace a la skill.
- `hive-review-ux.md`:
  - paso 1 (línea 10): la «project UI guide» pasa a ser el DESIGN.md;
  - paso 2: juzgar contra el DESIGN.md y reportar como defecto P0–P3, cuando el cambio lo introdujo, cada check **[measured]** de [criteria](skill:hive-ui-criteria/references/criteria.md); «a library default or a recorded decision is not a deviation» se cambia por el puntero a la precedencia.
- `flow-verify/references/ui-verification.md`: el ancho de 320 px CSS se suma solo para el check de reajuste de las regiones cambiadas.
- `flow-build/references/delegation.md` «Reviewer brief»: el brief del revisor UX también da cómo llegar a los estados que piden los checks medidos de la vista: «sin resultados» distinto del catálogo vacío, una lista de más de 1000 elementos, y volver desde un detalle. Si no se pueden alcanzar, dice cuáles.

## Impacto

- **Uso de la «project UI guide».** Se buscó el valor literal: aparece en `hive-design-ui.md:11`, `hive-build-frontend.md:15` y `hive-review-ux.md:10,12`. Todas cambian.
- **Catálogos del sitio.** `site/src/content/docs/index.mdx:89-98` y `skills.md` listan las skills de apoyo. El README no las lista.
- **«UI decisions».** Lo enlazan `hive-build-frontend.md:15`, `plan-handoff.md:12` y `flow-walkthrough/SKILL.md:10`. Los dos últimos no cambian, porque el paso 1 sigue en el mismo lugar.
- **Revisión UX.** El spec `workflow-lifecycle` «UX review in one pass» no cambia: el veredicto y el alcance por regiones siguen igual; solo crece la lista de checks.
- **Instalador.** Ningún catálogo del instalador ni de los tests nombra skills una por una: se leen por glob. El instalador trata como conflicto un directorio ajeno con el mismo nombre (`deployment-manager.md:489`), y por eso el prefijo.
- **Proyectos de usuarios.** El primer plan de UI en cada proyecto agrega un DESIGN.md versionado, que se entrega en su PR. No hay efecto en producción.

## Recuperación

Revertir el merge quita la skill y los punteros. Los DESIGN.md que ya se hayan creado en otros proyectos se quedan ahí: son archivos de esos repositorios y no se borran desde Hive.
