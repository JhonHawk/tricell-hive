# Design

## Approaches

Se presentaron por hallazgo y el usuario eligió los recomendados el 2026-11-04.

- **Cobertura al lanzar (D1).**
  - A, elegido: el brief lleva los medios para alcanzar cada estado; el revisor captura lo que falta y lista lo inalcanzable como «no cubierto».
  - B: el paso 1 captura todos los estados antes de lanzarlo. Retrasa el lanzamiento y carga todo al principal.
  - C: solo listar «no cubierto», sin medios. Es más barato, pero deja estados sin revisar.
  - Se eligió A porque el rol ya permite capturar «lo que el brief no trae». Lo que faltaba era cómo llegar a esos estados.
- **Relectura (D2).**
  - A, elegido: un seguimiento al mismo hijo cuando se corrige un hallazgo suyo. Si el host no tiene una herramienta para enviarle un mensaje, o el mensaje devuelve un error, un hijo nuevo con los hallazgos corregidos y acotado a lo corregido. Las demás correcciones las relee el principal, como hoy.
  - B: que el principal relea siempre. Es más barato, pero el principal juzgaría la corrección de un hallazgo ajeno.
  - Se eligió A porque los seguimientos acotados observados tardaron 1 min y conservan la independencia.
- **Veredicto (D3).**
  - A, elegido: el veredicto juzga solo lo que introduce el cambio; los defectos previos van en una lista aparte.
  - B: solo etiquetar introducido o previo. No elimina los turnos de aclaración.
  - Se eligió A porque `flow-verify` paso 3 ya envía los defectos previos al usuario y no a la ronda de correcciones; el veredicto los mezclaba con el criterio de corrección.

## Proposed shape

Cada regla queda en un solo hogar y los demás lugares la señalan:

| Regla | Hogar | Punteros |
| --- | --- | --- |
| Qué lleva el brief del revisor UX (medios por estado, o cuáles no se alcanzan) | `delegation.md` «Reviewer brief» | `flow-verify` paso 2 ya apunta allí |
| Qué hace el revisor con un estado inalcanzable, con un seguimiento y con su veredicto | `hive-review-ux.md` pasos 2–4 | — |
| Qué hace el principal con un «no cubierto» y cómo se relee una corrección de un hallazgo UX | `flow-verify` pasos 2 y 3 | `ui-verification.md:5` sigue regulando la relectura del principal |
| Descripción para personas | `site/src/content/docs/flows/verify.md` | — |

Redacción orientativa, en inglés en los archivos:

- **Rol, paso 2:** «The verdict starts as not passed and judges only the defects the change introduced in the reviewed regions… Report every defect that was already there in a separate list, each with its priority; it does not change the verdict.» Se conservan «skeptical reviewer» y «keeps its severity».
- **Rol, paso 3:** «A state the brief's means cannot reach: list it as not covered with what blocked it; it does not fail the verdict, and you do not stop to ask for it.» Y para un seguimiento: «review only the regions and states it names, against your earlier findings, and answer in one reply.»
- **Delegación:** «A UX reviewer's brief also gives how to reach each state its role always checks… such as the fixture route, the account and how to force the error, or says which cannot be reached.»
- **flow-verify, paso 3:** relectura de una corrección de un hallazgo UX con un seguimiento al mismo hijo, o con un hijo nuevo acotado si el host no lo conserva, y nunca una revisión completa nueva. Cada «no cubierto» se cubre en el recorrido propio o se reporta como no verificado en la validación.

## Impact

- **Quién lee estas reglas.** Todos los hosts que despliegan el rol y las skills. El veredicto cambia de significado para quien lo lee, que es el hilo principal en `flow-verify`. No cambia ningún contrato de máquina, ni el visor, ni los tests de límites del rol.
- **Riesgo.** Que «solo lo introducido» se lea como permiso para callar los defectos previos. AC6 exige reportarlos y mantiene el escepticismo y la severidad.

## Recovery

Revertir el commit o PR. No toca datos, configuración desplegada ni estado de otros proyectos.
