# Comunicación como receta positiva

| Campo | Valor actual |
| --- | --- |
| Estado | Listo para implementar · implementación pendiente de tu visto bueno |
| Tracker · GitHub Issues | Sin issue enlazado |
| Git | Directo a `rebuild/harness-engineering` (D2-A) · despliegue solo si gana la comparación a ciegas |
| Verificación | `go test ./tests/...` · inventario de reglas aceptado · piloto A/B Grok+Codex · elección a ciegas |
| Siguiente paso | `flow-build` desde T1, con tu visto bueno |

## Objetivo

Las sesiones de ark y sample-project escriben con una estructura idéntica en cuatro CLIs distintas (etiquetas en negrita por área, IDs, glosas, categorías de cierre) y con jerga en inglés traducida a medias. El research del 2026-09-27 atribuye esa sensación de «formulario» a la sección `## Communication` de [global.md](../../../../content/guidance/global.md): ocupa ~12 KB, casi todo en forma de obligaciones de formato, y la evidencia externa (guía de prompting de Anthropic, estudios de instrucciones en conflicto, la prueba de optional reference project) indica que una receta positiva con un ejemplo funciona mejor que una pila de reglas.

Este cambio reescribe la sección como una receta breve con un ejemplo corto de reporte de cierre. No pierde ningún comportamiento sin que el usuario acepte el descarte, y solo se despliega si el usuario prefiere la versión nueva en una comparación a ciegas.

## Alcance y aceptación

Incluye toda la sección `## Communication` (D12-A): prosa, formato, decisiones, esperas al usuario, reportes de cierre, categorías de cierre y pregunta de cierre. Excluye las demás secciones de `global.md` y las skills. Las referencias externas a la sección se conservan: el ítem «en curso» que cita Proportionality, la «cleanup label» de Retention and cleanup, la «global close question» de `flow-build` y el encabezado `Communication` que cita `flow-report`.

Restricciones que deben seguir siendo ciertas:

- Cada frase de la sección actual tiene su línea en la versión nueva o un descarte con motivo que el usuario acepta antes del piloto (T1). El inventario es la protección real contra pérdidas, porque el piloto solo ejercita una parte de las reglas.
- Se conservan los nombres de los que depende el resto de la guía: `Communication`, las categorías Pending, Next step, Recommendations y Reminders, «close question» y la línea de limpieza (que se sigue dando aunque el reporte tenga menos de 3 áreas).
- `go test ./tests/...` pasa y el presupuesto de `global.md` baja con el archivo.
- El contenido distribuido sigue en inglés.

Criterios:

- AC1. La sección `## Communication` ocupa como máximo 8 400 bytes, un 30 % menos que hoy. *Falsa en la base cuando* se mide: 12 041 bytes.
- AC2. La sección contiene un único ejemplo, un bloque cercado bajo la línea «Example completion report:», presentado como modelo solo para reportes de cierre. *Falsa en la base cuando* se busca esa línea: no existe.
- AC3. Las etiquetas en negrita por área se piden solo para reportes con tres o más áreas. *Falsa en la base cuando* se lee la línea 17: las pide en todo reporte de progreso o cierre.
- AC4. En la comparación a ciegas, el usuario prefiere la versión nueva en al menos 4 de 6 pares. «Sin preferencia» cuenta en contra de la versión nueva, y un par en el que alguna corrida no terminó se repite una vez. Es la preferencia del usuario, no una afirmación de fiabilidad (D14-A). *Falsa en la base cuando* no existe versión nueva ni pares.
- AC5. Las 12 corridas terminan (`completed`) y, dentro de cada celda (caso × host), ninguno de estos criterios que pasa en el brazo A falla en el brazo B: `question_after_detail`, `cited_id_glossed`, `no_bare_url`, `ticket_ids_not_packed_in_prose` y `Final writes within authorized fixture scope`. *Falsa en la base cuando* no hay corridas del brazo B.

## Entrega

- Repositorio y rama: `tricell-hive`, `rebuild/harness-engineering`, commit y push directos (D2-A, vigente en la sesión). Este registro se versiona en el mismo repositorio al quedar listo el plan, al desplegar y al archivar.
- Revisión de código: sin revisión dedicada (D3-A, vigente); la revisión del plan sí aplica.
- Piloto autorizado por el usuario solo para esta comparación (D11-A, 2026-09-27): Grok y Codex, 3 casos, dos brazos, una corrida por celda (D14-A), proyectos de prueba desechables, memoria aislada. `--allow-native-trust` autorizado en las corridas de Codex (D13-A): la entrada de confianza vive solo en el home paralelo de cada corrida.
- Despliegue a las 6 CLIs: solo si se cumplen AC4 y AC5. Si no, una ronda de ajuste y otra comparación; si vuelve a perder, se abandona el cambio y se archiva con el motivo.
