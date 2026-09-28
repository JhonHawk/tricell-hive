# Diseño

## Contexto verificado

- La sección `## Communication` de [global.md](../../../../content/guidance/global.md) (líneas 14–32 en el commit `074398b`) ocupa 12 041 bytes de ~41,5 KB. Casi todo son obligaciones de formato, acumuladas por correcciones hechas a partir de sesiones reales.
- Otras partes de la guía citan la sección:
  - Proportionality (línea 85) remite al «in-progress item described under Communication».
  - Retention and cleanup (línea 154) y la pregunta de cierre piden una «cleanup label».
  - `flow-build/SKILL.md` pide «Ask the global close question» y «its cleanup line».
  - `flow-report/SKILL.md` remite a «the Communication rules».
- Tests: `tests/content/budget_test.go` solo mide el tamaño de `global.md` y `skill_links_test.go` comprueba los links `skill:`. Ningún test compara el texto de la sección.
- Los criterios de regresión evalúan trazas de sesiones, no el texto de la guía:
  - En `tests/pilot/regression.go`: `question_after_detail`, `cited_id_glossed`, `no_bare_url` y `ticket_ids_not_packed_in_prose`.
  - En `flows.go`: `close_question_after_report`, que solo se evalúa en el caso `close-sequence`.
- Research del 2026-09-27 (resumen en Engram, `research/prose-communication-modes`):
  - La misma estructura aparece en Grok, Codex, Claude y OpenCode con cuatro modelos distintos, así que la impone la guía.
  - Los calcos de la jerga de la guía se repiten en tres o más CLIs.
  - Anthropic recomienda ejemplos positivos antes que prohibiciones. Los estudios de instrucciones en conflicto (arXiv 2502.15851 y 2606.22470) muestran que el cumplimiento baja cuando las reglas se acumulan.
- Supuesto sin medir: que la versión nueva mejore la prosa. Lo decide la preferencia del usuario a ciegas (AC4).
- Límite aceptado: el piloto no ejercita esperas al usuario, decisiones con ID, el ítem «en curso», la delegación ni la exclusión de autorizaciones en las opciones de ruta. Tampoco corre los modelos del perfil `execution`. Contra esas pérdidas protege el inventario de T1, no el piloto.

## Enfoque

Reescribir la sección en cinco bloques cortos, en inglés, con frases positivas. Una prohibición se mantiene solo donde la prohibición es la regla misma: tickets por párrafo, efectos agrupados en una opción de ruta, el silencio no autoriza.

1. **Voz:** abrir con el resultado; prosa clara en el idioma de la sesión; sin jerga interna, con los términos de la guía dichos en palabras comunes; evidencia en lugar de pasos; links con etiqueta, una vez por ticket y con la URL corta.
2. **Forma:**
   - prosa para causas y razonamiento, lista para elementos paralelos, tabla solo para comparar 3+ elementos en 2+ atributos;
   - un ticket por línea;
   - etiquetas en negrita solo en reportes con tres o más áreas;
   - la línea de limpieza siempre que haya limpieza;
   - ítem «en curso» primero cuando queda trabajo corriendo;
   - IDs con prefijo por lista, glosados cada vez que se citan fuera de su mensaje;
   - estado en palabras, no emojis; diagramas a un archivo;
   - link al documento retenido en lugar de reproducirlo;
   - oferta de `flow-report` cuando el usuario va a compartir o conservar el material;
   - el formato de una skill tiene precedencia.
3. **Preguntar:**
   - decisiones con ID, efecto y opción recomendada;
   - reusar las decisiones vigentes (el silencio no autoriza);
   - una sola pregunta para un conjunto de acciones;
   - peticiones autosuficientes, con un bloque por flujo;
   - detalle antes de la pregunta nativa; si el turno termina, repetir la pregunta en texto y tomar la primera respuesta.
4. **Cerrar:**
   - abrir el reporte con el punto alcanzado y lo que no incluye; nombrar los secretos que llegaron a un lugar duradero;
   - las categorías de cierre con su definición condensada: Pending con un solo dueño (lo de otro proyecto va a su propia sesión), Next step, Recommendations con el flujo de cada ruta y sin efectos agrupados, Reminders en una línea (se vuelven recomendación si esperar tiene un costo concreto);
   - distinguir lo hecho, lo que continúa y lo que necesita autorización nueva;
   - la pregunta de cierre con sus opciones, los criterios de sesión nueva, lo que queda abierto y sus excepciones.
5. **Ejemplo:** un único bloque cercado bajo «Example completion report:», de unos 600 bytes, con placeholders (`<ticket>`, `<commit>`) en lugar de frases hechas.
   - Se presenta como modelo solo para reportes de cierre. Las respuestas de estado, las preguntas y las esperas siguen la receta.
   - La prosa del mensaje va en el idioma de la sesión.
   - El propio ejemplo cumple las reglas que se miden: link corto una vez, línea de limpieza y pregunta de cierre con el flujo nombrado.

Las reglas de asesoría (razonar, cuestionar premisas, prerrequisitos frente a bloqueos) y el disparador del reporte de backlog se condensan sin cambiar su alcance.

Se descartó mover partes de la sección a una skill: la sección gobierna todas las respuestas, incluso sin skill activa, y `AGENTS.md` exige que la autorización y la forma básica no dependan de elegir una skill.

## Comparación a ciegas

- **Brazos:**
  - A: `git archive` del commit base.
  - B: el mismo `git archive` más los dos archivos que cambia T1.
  - Antes de correr, `diff -r` entre ambos muestra solo esos dos archivos y el `GuidanceBlockHash` de `run.json` difiere entre brazos.
  - `budget_test.go` no se instala; el contenido que ven los hosts difiere solo en `global.md`.
- **Casos:**
  - `backlog-status` (nuevo): un tracker local declarado en el `AGENTS.md` de la fixture con `## Hive`, un `BACKLOG.md` con 6–8 tickets en 2–3 módulos con formato `ABC-123` y `expected.skill_read: "flow-research"`.
  - `direct-build`: reporte de cierre tras un cambio pequeño.
  - `adaptive-plan`: recomendación y pregunta sin Git, en lugar de `close-sequence`, para no necesitar push ni abrir el sandbox de Codex. `close_question_after_report` no se evalúa en estos casos; la pregunta la juzga el usuario.
- **Hosts y modelos:** Grok y Codex. Un paso previo lee el modelo configurado en cada CLI para `--configured-model`, y se pasa `--effort` (`high` en Grok, `medium` en Codex). `--timeout 600s` en las 12 corridas.
- **Cegado:** el extractor toma el último texto del asistente y la pregunta de cada `events.json`. Reemplaza por un marcador neutro las rutas de la fixture y de la corrida, que incluyen la letra del brazo. Mezcla el orden de cada par y guarda la clave aparte.
- **Límite de aislamiento:** Grok usa su `GROK_HOME` real para la autenticación, y ahí escribe su historial nativo. Ese historial no se borra; queda declarado.
