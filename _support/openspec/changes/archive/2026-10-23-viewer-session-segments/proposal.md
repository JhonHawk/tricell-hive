# Visor: sesiones con tramos y enlaces de entrega

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado mediante PR #124 (`59e2e627`) |
| Seguimiento · GitHub Issues | Sin incidencia vinculada; no se ha autorizado crear una |
| Git | Interactivo · `development` · validación visual antes del push |
| Verificación | Contratos Python/Go · frontend · compilado local · UX y funcional · Codex y seguridad |
| Siguiente paso | Confirmación de copia final pendiente en visor cotidiano; evidencia y worktree conservados |

## Solicitud

2026-10-23, acompañada de dos imágenes de referencia de Orca:

> Vamos a trabajar esto en un worktree porfa, es sobre el visor [Image #1] [Image #2] claro que primero necesitamos una propuesta, algo que vamos a hacer es:
>
> - Antes, si trabajabamos varios tickets y "continueaciones" durante la misma sesión, se separaban, por eso la UI anterior del sidebar era como una carpeta, ahora todo debe entrar visualmente en una sesión y tener un separador interno
>
> [sesión 1]
> ----
> [continuación ticket #NN]
>
> y asi, dame una propuesta al respecto,
> Finalmente, tambien agregemos una sección en el header que muestre el link/acceso al PR donde se implemento (o commit) de lo que se habla

El usuario eligió la ruta «Planear» y después confirmó «D2-A»: fijar, archivar/restaurar y eliminar por tramo. En la misma sesión eligió entrega interactiva, `codex review` y verificación UI completa. D6 (ruta de ejecución): el usuario eligió iniciar flow-build con este plan.

Nueva instrucción del usuario durante la implementación:

> veo que se sigue generando el .html, pero este solo debe hacerse si el binario no esta instalado (la intención del visor es justamente ahorrarnos el uso de tokens en la creación del html)

## Objetivo y alcance

Presentar una conversación como una sola sesión navegable, con sus distintos trabajos o continuaciones apilados y separados dentro de ella. Ofrecer enlaces explícitos al PR o commit del que habla cada trabajo y reunirlos en el header de la sesión. Con `hive` disponible en PATH, publicar sólo el manifiesto JSON y registrarlo en el visor; generar HTML standalone únicamente sin binario. No se afirma un ahorro medido de tokens: el HTML anterior ya se generaba por script.

Incluye agrupación, carga y selección de tramos, controles independientes, renderer compartido seguro para montajes simultáneos, contrato opcional de entregas y sus productores/documentación. La implementación se realizará en un worktree desde `development`. No incluye consultar estados de checks/merge a proveedores, correlacionar históricos por heurísticas, reescribir walkthroughs históricos, ni operaciones conjuntas sobre la sesión. La publicación, promoción a `master`, release y reinicio del visor cotidiano son efectos distintos; sólo la entrega a `development` y su refresco local condicionado están acordados aquí.

## Criterios de aceptación

- AC1. Dos copias normales con el mismo proyecto, host y `session.id`, aunque pertenezcan a tickets distintos, aparecen en una sola entrada del sidebar y cuentan como una sesión. Proyectos, hosts o IDs distintos permanecen separados. *Falso en la revisión local inspeccionada:* cada clave física produce una entrada independiente.
- AC2. Al abrir una sesión con dos tramos, ambos contenidos se muestran apilados en orden cronológico, con separadores y encabezados propios; anclas, pestañas de capturas, carrusel, foco e impresión de uno no controlan al otro ni dejan listeners tras desmontar. *Falso en la revisión inspeccionada:* sólo se monta un documento y sus IDs se repiten entre montajes simultáneos.
- AC3. Archivar/restaurar, fijar y eliminar un tramo de una sesión agregada afectan sólo a su clave física; los hermanos conservan sus valores, documentos cargados y recursos. El filtro de archivados deja visible la sesión mientras quede un tramo visible. *Falso en la revisión inspeccionada:* no existe una sesión agregada con acciones por tramo.
- AC4. Una sesión con entregas explícitas presenta enlaces accesibles en el header y en el tramo correspondiente, deduplicados por URL en el header; sin metadatos muestra «Sin entrega registrada». Dos enlaces caben directamente y los demás quedan accesibles en «Ver todas». *Falso en la revisión inspeccionada:* no existe contrato ni sección de enlaces de entrega.
- AC5. Python y Go aceptan el nuevo `delivery` opcional con el mismo contrato de host ASCII/punycode o IP, manteniendo Unicode en etiquetas y rutas, rechazan entradas/URLs inválidas sin publicar copias parciales y continúan aceptando manifiestos anteriores sin el campo; una reingestión inválida conserva los bytes y el hash de la última copia válida. *Falso en la revisión inspeccionada:* ambos rechazan `delivery` como desconocido.
- AC6. La producción y cierre del walkthrough actual permiten registrar sus PRs/commits conocidos en `delivery`, conservando el candidato mostrado y sin modificar otros tramos ni atribuir una entrega no verificada. *Falso en la revisión inspeccionada:* el procedimiento y contrato no disponen de metadatos de entrega.
- AC7. Buscar un ticket de un tramo encuentra la sesión y permite llegar a ese tramo; la selección por clave física sobrevive a refresco y los saltos internos siguen funcionando al archivar/eliminar un hermano, sin respuestas tardías que sustituyan otra sesión. *Falso en la revisión inspeccionada:* la navegación sólo selecciona copias independientes, sin saltos dentro de una sesión.
- AC8. El viewer y la página standalone renderizan entregas válidas como enlaces seguros con etiqueta tratada como texto, `target="_blank"` y `rel="noopener noreferrer"`; no crean un `href` para una entrega inválida aunque se invoque directamente el renderer. *Falso en la revisión inspeccionada:* no existe renderizado de entregas ni defensa de destinos externos en ese límite.

- AC9. La disponibilidad del ejecutable `hive` determina la salida del publicador: con binario valida/actualiza JSON sin construir, crear ni modificar HTML, incluso si el daemon está apagado; sin binario genera el fallback HTML. El cierre/confirmación mantiene sus resultados y conservación de evidencia en ambos modos. *Falso en la base:* el script construye y publica HTML incondicionalmente.

- AC10. Elegir conservar `walkthrough.json` conserva también sus imágenes/grabaciones referenciadas, igual que elegir la página cuando existe; el cierre respeta esa unidad. *Falso en la base:* sólo seleccionar la página garantiza conservar dependencias, mientras JSON se trata como archivo individual.

## Restricciones que deben conservarse

Claves físicas, capacidades y autenticación por copia; exclusión de fuentes tracked; advertencias de procedencia; estado de trabajo explícito por tramo; TTL y pin independientes; ausencia de cambios a fuentes por las acciones del visor. Mantener renderer/estilos compartidos con páginas offline y bundle generado vigente. No dependencias nuevas ni pilotos de modelos. Preservar `_support/sessions/`, que ya estaba sin seguimiento al comenzar.

## Entrega acordada

- D1 (ruta): planificación elegida por el usuario el 2026-10-23; D6 autorizó la construcción.
- D2-A (acciones): individuales por tramo, sin menú de operaciones conjuntas.
- D3 (Git): interactivo para `tricell-hive` hacia `development`, con worktree y rama de trabajo; commits y checks locales; detenerse antes del push para validar el resultado. Después de validación: push, PR y merge cuando pasen los controles acordados. Rama creada: `feat/viewer-session-segments`; worktree `/Users/jmartinez/.codex/worktrees/fa01/tricell-hive`.
- D4 (código): `codex review`, disponible según la ayuda del CLI inspeccionada; revisar candidato local antes de pedir validación. Contrastar hallazgos bloqueantes y añadir revisión de seguridad para el límite de URLs externas. Las correcciones posteriores se rigen por los límites de revisión y renovación de evidencia de Hive.
- D5 (UI): completa, con `hive-review-ux` y `hive-verify-change` independientes; conserva vigencia para el resto de esta sesión en este repositorio.
- `_support/workspace/maintainer.local` existe: al integrar en `development`, reconstruir manager antes de `hive update` por cambiar `tooling/` y `content/`, usando el checkout integrado. No reiniciar el daemon cotidiano sin autorización expresa. Un runtime viejo puede rechazar el campo nuevo; conservar la evidencia que no pueda confirmar.
- D9 (salida de walkthrough): instrucción explícita del usuario: JSON y registro con `hive` instalado/disponible en PATH; HTML sólo sin binario. Presencia del proceso, errores de registro o confirmación no habilitan fallback. Preservar HTML previos sin tratarlos como salida actual.
- D8 (primera composición): aprobada por el usuario; autoriza continuar controles del compilado, todavía no satisface validación final anterior al push.
- D7-A (hosts de entrega): el usuario eligió admitir hosts ASCII, incluidos punycode e IPs; Unicode sigue permitido en etiquetas y rutas. Se corrige y verifica este contrato acotado, sin normalización IDN directa.
- El registro se mantiene sin commit mientras corre el trabajo. Tras integración y controles aplicables, aplicar deltas y archivarlo mediante una pequeña PR de cierre sin revisión dedicada, fusionada por el agente. No promoción a `master` ni release en esta entrega.

## Fuentes y estado

Estado final: validación humana D10 aprobada; PR #124 integrado en `development` (`59e2e627`). Binario y contenido local actualizados desde ese commit, sin reinicio del visor cotidiano. Los párrafos siguientes describen la investigación inicial histórica.


Revisión local observada: `4ed215dbf3947b476cee0eea482b69de71123a16`, rama `development`, 2026-10-23. No se ha afirmado frescura frente al remoto; el build la reconciliará antes de crear el worktree. El [diseño](design.md) contiene contexto y contratos; [tareas](tasks.md) contiene ejecución, controles y revisión.
