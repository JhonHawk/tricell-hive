# Design

## Approaches

- **A — Test con lista local (elegida, D13 A).** Las rutas `/Users/<nombre>/` se comprueban con un patrón commiteado; los nombres y prefijos de clientes, con un archivo ignorado por git. Reusa la comprobación de texto que corre en cada cambio de `content/` y `_support/`. Límite: la parte de nombres corre solo donde existe el archivo local.
- **B — Lista commiteada como hashes.** Corre en cualquier clon, pero nombres cortos se recuperan por diccionario.
- **C — Configuración local de gitleaks leída por `git_state.py commit`.** Bloquearía al commit, pero no está medido que `gitleaks stdin` lea una configuración local, y solo cubre commits hechos por ese script.

## Decisions

- El test lee archivos versionados y nuevos no ignorados, para fallar antes del commit y no solo después.
- Los mensajes de fallo nombran archivo, línea y número de regla, nunca el término: la salida de un test termina pegada en PRs y sesiones.
- El mapa de reemplazos vive en el mismo archivo local que la lista, para que el reemplazo de T3 y la comprobación de T1 no diverjan.

## Impact

- El archivo de cambios cambia texto sin cambiar estructura. `TestSpecDeltasApplied` compara el último delta archivado de cada requisito con el spec, así que el reemplazo es idéntico en los dos lados.
- El bundle del visor se regenera porque su digest incluye las fixtures y los tests del frontend; solo debe cambiar `sourceDigest`, porque el código de desarrollo no entra al bundle.
- Los commits y cuerpos de PR ya publicados conservan los identificadores; D11 A no reescribe el historial.

## Recovery

Revertir el commit del PR restaura el texto anterior. El archivo local no se versiona; borrarlo apaga la parte de nombres del test.
