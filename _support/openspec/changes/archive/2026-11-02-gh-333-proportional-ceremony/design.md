# Diseño — #333

## Hogar canónico de cada regla

| Regla | Hogar canónico | Punteros que cambian |
| --- | --- | --- |
| Criterio de recomendación del modo de entrega (AC1) | `global.md` «Delivery», en la viñeta de la pregunta de entrega | `flow-plan` paso 4 y `flow-build` ya apuntan a «Delivery»: sin cambio |
| Tarea operativa (AC2) | `global.md` «Proportionality», junto a la definición de cambio mecánico | «Delivery» (exención de preguntas), «Scope and authorization» (ticket), `flow-build` (entrada, paso 1, paso 5), `flow-close` (disparador) |
| Plan en el chat o carpeta de cambio (AC3) | `flow-plan/SKILL.md` paso 3 y sección «Change folder» | `plan-handoff.md` (ruta y ruta de la carpeta), `flow-build` (checkpoint) |
| Slug sin ticket (AC4) | `flow-plan/SKILL.md` «Change folder» | plantilla `proposal.md` (título), `global.md:67` (ruta sin ticket) |
| Cuándo crear y cómo relacionar un ticket (AC5) | `global.md` «Scope changes», punto 4, que pasa a definir el destino de un tema aparte o un hallazgo | `flow-close` «Close question» (recomendación por hallazgo) |
| Tracker en tarea operativa (AC6) | `global.md` «Scope and authorization», viñeta del ticket | `flow-build` paso 1 |
| Disparadores y alcance del cierre (AC7) | `flow-close/SKILL.md` | `global.md` «Preservation» (autorización de limpieza local), `flow-build` paso 5 |

## Llamadas menores tomadas sin preguntar

- La carpeta de cambio también se escribe cuando el plan cambia un contrato, datos almacenados, seguridad o una dependencia: son los casos que llevan `hive-review-plan`, que revisa un plan guardado, y concretan «decisiones de peso que necesitan registro».
- También cuando el build delegará tareas a hijos: su brief nombra la carpeta de cambio y la tarea.
- Sin ticket, el comando de ruta lleva como argumento una descripción breve del trabajo o el slug de la carpeta; así se mantiene la regla de «argumento real».
- La limpieza local de contenedores, imágenes y dumps creados por la tarea entra en la autorización de «Preservation», como los temporales; los recursos remotos se preguntan porque destruir infraestructura o datos fuera de lo local es un efecto crítico.
- `hold` no recibe cierre hasta su entrega; los procesos que inició siguen la regla vigente de detenerse al cierre, sin regla nueva.
