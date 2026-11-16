# Design

## Approaches

Decididos con el usuario el 2026-11-16 (D22):

- **A — Skill como procedimiento, con punteros (elegido).** El cuándo proponer un ticket y la autorización del usuario siguen en la guía global. El cómo (buscar uno existente, su contenido, su nivel, cómo se relaciona con su origen, la prueba de «listo») se muda a la skill, y los flujos que lo enunciaban dejan un puntero. Tras la respuesta del usuario, el flujo sigue la skill en el mismo turno. Deja cada regla en un solo sitio, que es lo que pide el `AGENTS.md` del repositorio, y cumple que build y close lleguen a la skill.
- **B — Ruta con comando propio.** Crear un ticket pasaría a ser una ruta entre flujos, con su bloque de siguiente paso. Control explícito, pero un comando por ticket en medio de un build o de un cierre.
- **C — Skill independiente.** Sin tocar las reglas existentes. Menos archivos, pero las reglas quedarían en dos sitios y ningún flujo llegaría a la skill.

## Decisions

- **Los flujos llegan por la guía global.** `flow-build`, `flow-verify` y `test-audit` ya remiten a la regla de tickets de «Scope changes». Al apuntar esa regla a la skill, llegan a ella sin editarse. `flow-build` está a 850 bytes de su límite y no se toca.
- **Dos momentos, un hogar.** Buscar un ticket existente hace falta dos veces: antes de proponer uno (para no proponer un duplicado) y antes de crearlo. Los pasos viven en la skill; `flow-research` y `flow-close` conservan solo «comprueba si ya existe, como fija `create-issue`».
- **Lo propio de la auditoría se queda en la auditoría.** El cuerpo de un ticket de hallazgo añade los campos del hallazgo, sus ubicaciones con el commit auditado y los comandos que comprueban el arreglo. Es una excepción de un solo punto de uso y se queda en `code-audit.md`.
- **Borrador solo donde nadie eligió aún (D29 A).** Usada sola, o al desglosar, la skill muestra el borrador o el lote y escribe tras la confirmación. Cuando la respuesta del usuario ya eligió el ticket dentro de un flujo, como «H1 ticket» en un cierre, lo crea y muestra el ticket creado con su enlace. Sin supervisión crea el ticket suelto de un hallazgo sin confirmación, como hoy.
- **Cierre según quién la llama.** Usada sola termina en el ticket y ofrece el siguiente paso. Seguida desde un flujo, devuelve el control a ese flujo: un cierre debe poder terminar con su pie después de crear los tickets de sus hallazgos, y una auditoría con su bloque.
- **Las cuatro formas de relación se mudan enteras.** Bloqueo cuando un ticket espera a otro, sub-issue para una parte aplazada del alcance del origen, relacionado en otro caso, y un comentario con la URL en cada ticket cuando el origen está en otro tracker.
- **Un lote es un padre con hijos.** Varios tickets de hallazgos elegidos uno a uno son tickets sueltos. Los hijos de un desglose que toca varios repositorios se crean todos en el tracker principal y cada uno nombra su repositorio.
- **Sin `## Hive`, se pregunta.** Una sesión sin repositorio no tiene tracker declarado: la skill pregunta en cuál y en qué repositorio crear antes de buscar o escribir.
- **Sin nombres de comandos en el contenido.** Las referencias nombran el concepto del tracker (sub-issue, relación de bloqueo, tipo de issue) y piden usar «the tracker's own CLI or tool», como ya hace la guía global, porque las opciones de cada CLI cambian entre versiones.
- **Sin repositorio, sin pista.** Una sesión sin el código a la vista omite la pista de solución y lo dice en el ticket; el resto del procedimiento es igual.
- **El umbral de 7 criterios no tiene fuente.** Se eligió para dar una prueba comprobable; el primer uso real y las sesiones siguientes dirán si se ajusta.

## Proposed shape

```
content/skills/create-issue/
  SKILL.md                        entradas, pasos y cierre
  references/ticket-types.md      campos por tipo, lista de «listo», pista, autoría
  references/breakdown.md         señales de desglose, el lote, su aprobación, las escrituras
  references/trackers.md          padre e hijos, dependencias y bloqueo por tracker; permisos
  references/existing-ticket.md   validar un ticket que ya existe
```

Pasos del `SKILL.md`:

1. **Elegir la entrada.** Crear, desglosar o validar un ticket existente; esta última lee `existing-ticket.md` y termina como fija.
2. **Leer antes de preguntar.** La fuente (hallazgo, pedido o documento de requisitos), el código que toca cuando hay repositorio, y cada tracker declarado en busca de un ticket que ya exista. Si existe, lo reporta y no crea otro.
3. **Decidir el nivel.** Cuando se cumple una señal de desglose, lee `breakdown.md`.
4. **Preguntar lo que falta.** Hasta tres preguntas por ronda, cada una con respuesta recomendada; lo demás queda como supuesto escrito en el ticket.
5. **Redactar.** Con la plantilla de su tipo, de `ticket-types.md`.
6. **Comprobar.** Contra la lista de «listo». Una regla de negocio sin decidir deja el ticket bloqueado, como fija `trackers.md`.
7. **Mostrar y confirmar.** Usada sola o al desglosar: el borrador, o el lote completo, y una confirmación. Cuando la respuesta del usuario ya eligió el ticket dentro de un flujo, este paso se omite.
8. **Escribir.** En el tracker principal, relacionándolo con su origen, y releyendo lo creado.
9. **Cerrar.** Reporta el ticket con su enlace y su estado (listo o bloqueado). Usada sola, ofrece el siguiente paso y se detiene; seguida desde un flujo, vuelve al paso siguiente de ese flujo.

Qué cambia en cada lugar que hoy enuncia el comportamiento:

| Lugar | Hoy | Queda |
| --- | --- | --- |
| `global.md:95` | Crear un ticket es una acción suelta, ofrecida como respuesta o preguntada | Igual, más: una vez respondido, crearlo sigue `create-issue` |
| `global.md:149` | Cuándo proponer; buscar uno existente; relacionarlo con su origen | Cuándo proponer, y un puntero a la skill para crear, desglosar o validar |
| `global.md:116` | Los tickets nuevos, solo en el tracker principal | Sin cambio; la skill lo enlaza |
| `flow-research/SKILL.md:20` | Enuncia la búsqueda antes de proponer | Puntero a la skill |
| `flow-research/SKILL.md:34` | Crear un ticket es una acción suelta | Igual, más el puntero |
| `flow-research/SKILL.md:36–38` | Enuncia la prueba de «listo» | Conserva impacto, complejidad y dependencias; remite a la lista de la skill |
| `code-audit.md:71` | Remite a la regla global, incluida la búsqueda, y define el cuerpo | Remite a la skill; conserva lo propio del hallazgo |
| `flow-close/SKILL.md:79–85` | Enuncia la búsqueda y que la regla global rige duplicados y relación | Puntero a la skill; el texto neto no crece |
| `unattended-delegation/SKILL.md:14` | Permite abrir un ticket por un hallazgo | Igual, acotado a un ticket suelto, nunca un lote |
| `flow-plan/SKILL.md` | No nombra la pista | Una frase: la pista de un ticket es un dato por verificar |
| `flow-build`, `flow-verify`, `test-audit` | Remiten a la regla global | Sin cambio |

No cambian: mover un ticket de estado, el comentario de pendientes, ni la escritura de reglas de negocio al ticket.

## Copy

Secciones de un ticket. La referencia las nombra en inglés; el ticket se escribe en el idioma de la sesión.

| Tipo | Secciones |
| --- | --- |
| Todos | Title · Outcome (qué y por qué) · Done when (criterios) · Out of scope · Open questions and assumptions · Dependencies · Hint, not binding (opcional) |
| Story | más For whom and why |
| Bug | más Steps to reproduce · Expected and observed · Environment and version · Evidence |
| Task | más The change · Verify with; la deuda técnica añade What it costs today |
| Spike | más Question · Timebox · Deliverable · Decision it feeds |

Lista de «listo», ocho pruebas:

1. El título y una frase de resultado nombran el cambio sin detallar la implementación.
2. Cada criterio es observable y otra persona puede verificarlo.
3. Cada criterio dice una sola regla.
4. No hay términos vagos ni marcadores «por definir».
5. Hay una línea de fuera de alcance.
6. Las preguntas de negocio abiertas están listadas; una sin resolver deja el ticket bloqueado.
7. Un bug trae pasos, esperado contra observado y entorno.
8. El trabajo cabe en un cambio que se integra de una vez; un spike tiene límite de tiempo y entregable.

Límites de tamaño, para que el ticket no reemplace al plan: como mucho 7 criterios; una pista de hasta cinco ubicaciones y una frase; lo que el ticket necesita de un documento enlazado se resume dentro, en hasta tres frases, en vez de solo enlazarlo.

Fuentes: INVEST (Wake), EARS (Mavin), Example Mapping (Wynne), Connextra y Cohn para historias y spikes, el syllabus CTFL de ISTQB y Bettenburg et al. (2008) para bugs, y las guías de GitHub, Devin, Codex, Cursor y Claude Code para tickets que trabaja un agente. Dos estudios de 2026 miden que una solución sugerida y la localización predicen que un agente resuelva el ticket, y que los tickets largos y los que solo enlazan bajan esa probabilidad. No se adapta texto de esas fuentes, así que no hace falta archivo de licencia.

## Impact

- **Catálogo:** de 13 a 14 skills. La descripción nueva añade entre 400 y 600 bytes a cada sesión de cada host.
- **Guía global:** se acorta, porque salen los pasos de búsqueda y de relación. El pull request declara el tamaño antes y después.
- **`flow-close/SKILL.md`:** está a 113 bytes de su límite. El cambio sustituye texto; la tarea comprueba que no supere 16.000 bytes.
- **Riesgo que solo el uso real resuelve:** que un modelo siga el puntero a la skill en medio de un cierre largo. El primer uso real (T7) da una observación; no es una prueba.
- **Tracker público:** los tres tickets de T7 se crean en un repositorio público.

## Recovery

Revertir el pull request y volver a ejecutar `hive update --source <checkout>` restaura las reglas anteriores en cada host. Los tickets de T7 se cierran como no planeados; el usuario es administrador del repositorio y puede borrarlos.
