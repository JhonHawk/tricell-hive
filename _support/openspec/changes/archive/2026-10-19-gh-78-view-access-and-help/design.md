# Diseño

## Contexto verificado

Base inspeccionada: `0d2e51446476fc04b1664a6f7b4f507694bc7869`, igual al remoto `development`, 2026-10-19. Go declarado: 1.27.0.

- [view_command.go](../../../../../tooling/cli/view_command.go): `viewOpen` entrega `/#token=` a `deps.open`; `defaultViewOpener` usa `exec.Command(name, url)`. Token aleatorio nuevo en cada servidor.
- [view_command_test.go](../../../../../tooling/cli/view_command_test.go): `TestViewOpenNeedsARunningServer` pasó aislada y exige ese argumento; no abre navegador real.
- [server.go](../../../../../tooling/viewer/server.go): autoriza token en cabecera y mismo origen en escrituras; valida Host; JSON sin caché. Las capacidades de archivos dependen del token.
- [viewer.mjs](../../../../../tooling/viewer/static/viewer.mjs): `takeToken` es síncrona y la usan `main` y `hashchange`; guarda token en `sessionStorage`.
- [main.go](../../../../../tooling/cli/main.go): ayuda general ya contiene seis subcomandos. `view --help/-h` reprodujeron salida 1; `start --help` y `register -h`, salida 0.
- Fuentes consultadas vía Context7 `/mdn/content`, verificadas en MDN el 2026-10-19: [fragmentos](https://developer.mozilla.org/en-US/docs/Web/URI/Reference/Fragment), [replaceState](https://developer.mozilla.org/en-US/docs/Web/API/History/replaceState), [Origin](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Origin). Son APIs web sin versión de paquete. Un fragmento no viaja en la solicitud HTTP, pero sí forma parte del argumento local; reemplazar historial no borra observaciones previas. No se observó un robo real.

## C1 — Emisión y canje coordinados

Propuesta concreta de contrato:

1. `POST /api/access-codes`, cuerpo vacío, token mediante `X-Hive-Token` y `Origin` explícito del servidor en el cliente CLI. Validar método y controles existentes antes de emitir. Respuesta 200 JSON `{code: string}`, `Cache-Control: no-store`.
2. Código criptográfico de 32 bytes (64 caracteres hex), independiente del token; vigencia fija de 60 segundos desde emisión. Estado solo en memoria, compartido bajo mutex. Retirar vencidos al emitir y canjear; fijar un máximo de 64 códigos pendientes, rechazando nuevas emisiones con 429 hasta liberar cupo, sin invalidar códigos válidos de otras aperturas.
3. `POST /api/access-codes/exchange`, JSON `{code: string}` en cuerpo, máximo 1 KiB; exigir `Content-Type: application/json`, `Host` admitido y `Origin` exactamente igual a `http://` + Host. No aceptar origen ausente, `null` o extranjero; no habilitar CORS. Rechazar campos desconocidos, JSON adicional y códigos malformados. Esta es la única ruta que sustituye token por código como credencial.
4. Cuerpo/tipo inválidos: 400/415; método no permitido: 405; autenticación, origen o código ausente/inválido/vencido/consumido: 403; evitar ecos de valores. Validación de formato de código incorrecto puede responder 403 junto a cualquier código inválido. Comprobar caducidad (`now >= expires`) y retirar código dentro de una sola sección crítica antes de devolver 200 JSON `{token: string}` sin caché. Dos solicitudes concurrentes nunca obtienen dos éxitos. Si se pierde la respuesta, ejecutar `hive view open` nuevamente.
5. Códigos pendientes desaparecen al reiniciar. No cambian `server.json`, token del servidor, `sessionStorage` ni capacidades existentes. Emisión/canje no se registran con valores sensibles. Corregir el comentario de `Server` que actualmente afirma que no tiene estado.

La CLI usa cliente sin proxy ya existente, valida respuesta limitada y código antes de abrir URL. Los errores de emisión/abridor se convierten en mensajes acotados que nunca repiten URL con código ni cuerpo HTTP sensible. `open` conserva el comportamiento de servidor ausente. Un fallo del abridor deja el código vencer; no necesita otra API de revocación. Si un servidor anterior sigue activo y devuelve 404 al emitir, fallar sin abrir navegador ni recurrir a `#token`; indicar `hive view stop`, `hive view start` y `hive view open`. No reiniciar procesos cotidianos automáticamente. Probar este estado con servidor anterior sintético.

El código sigue siendo credencial portadora: un proceso local que lo robe antes del canje puede ganar la carrera. Este cambio mitiga exposición reutilizable; no promete aislamiento frente a procesos locales maliciosos. `Origin` no autentica clientes HTTP locales.

## C2 — Consumidor web

Extraer `#code` y limpiar URL antes del primer `await`; no guardar código. Canjear una vez, guardar solo token recibido y continuar lectura con cabecera existente. Un contador de generación compartido por inicio y `hashchange` impide que respuestas antiguas cambien token, estado de acceso o vista tras una apertura más reciente. La protección incluye lecturas/escrituras autenticadas anteriores: un índice viejo o un 403 tardío no puede aplicar datos ni ejecutar `rejectToken` sobre el acceso nuevo. No reintentar automáticamente un código consumible. Sin código, reutilizar token de la pestaña, como hoy; ante un código explícito fallido mostrar aviso de reabrir, sin ocultar el fallo usando un token previo. Recarga posterior al éxito conserva acceso.

Eliminar procesamiento de `#token` como credencial nueva. La coordinación CLI/servidor/JS se entrega en el mismo binario, sin transición dual. Pestañas con token vigente siguen usándolo hasta reinicio; tras reiniciar, reabrir con CLI actualizado. No migración de copias almacenadas ni modificación del servidor cotidiano durante pruebas. Para volver al binario anterior basta reconstruirlo desde la revisión anterior y, si el usuario lo decide, reiniciar su visor; no cambia el formato persistente.

## C3 — Ayuda

Un helper con `io.Writer` reutiliza el listado de ayuda general desde `runView`; reconocer `--help`, `-h`, `help` antes de resolución de opciones o estado. Mantener la convención de ayuda general y las ayudas por subcomando. No ampliar parser ni relajar argumentos de otros comandos.

## Interfaz

Área: visor `/`, módulo propio `viewer.mjs`. Se reutilizan shell, tarjeta de acceso, mensajes bilingües y acción `hive view open`. No se cambia composición, navegación, carrusel ni CSS. Referencia inspeccionada: fuente actual en la base; esta ronda no inspeccionó un render. Estados a comprobar en build: acceso satisfactorio, recarga, reapertura, código vencido/repetido, servidor reiniciado, fallo de red y apertura antigua que termina después de la nueva. Idiomas es/en; foco y acceso al comando de recuperación conservados. No requiere nueva guía visual.

## Documentación y especificación

Actualizar README y referencia de arquitectura para describir establecimiento de acceso mediante código, sin afirmar que se elimina todo riesgo local. Revisar `llms.txt` junto con README y modificarlo solo si cambian afirmaciones que contenga. Revisar `site/src/content/docs/flows/build.md`, que remite al visor, y actualizar si cambia su descripción. No reescribir diseños archivados. Añadir el delta [walkthrough-viewer](specs/walkthrough-viewer/spec.md) y aplicarlo al cerrar, no antes.
