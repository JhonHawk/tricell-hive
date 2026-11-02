# Diseño

## Un solo orden para la entrada del índice

Los dos lenguajes siguen el orden que ya usa Go: decodificación estricta, puntero exigido por el estado, cabecera, límite y registros de tickets, identidades de tickets, `validation_count`, id del puntero y puntero permitido por el estado. TypeScript mueve el id del puntero y `validation_count` a esas posiciones. Una entrada con un solo fallo da el mismo código que hoy en ambos lenguajes; una con varios puede cambiar de primer código en TypeScript.

## Cabecera compartida

- Go: `validateTimelineSummaryHeader(timelineSummaryHeader)` con los campos de cabecera; `ValidateTimelineSession` y `validateTimelineIndexEntry` construyen el valor y la llaman en el mismo punto donde hoy hacen los chequeos.
- TypeScript: `summaryHeader(obj)` con los mismos campos más `archived`/`pinned`; la llaman `parseTimelineSession` y `parseTimelineIndexEntry`.

## Tickets: validar una vez, límite con el código de Go

- Go: `addTickets` solo registra identidades y URL; la validación de cada ticket queda en `timelineTickets`. `ValidateTimelineSession` llama a `timelineTickets` antes de `validateTimelineCatalog`, como ya hacía TypeScript, así que en Go una sesión con más de 100 tickets y un duplicado pasa de `duplicate_identity` a `array_required_or_limit`.
- TypeScript: `addTickets` deja de llamar a `validateTimelineTicketRecord`. `tickets()` con `full` (sesión, entrada del índice y bootstrap) falla con `array_required_or_limit` en el campo `tickets` cuando la lista pasa de 100; las referencias de eventos (`ticket_refs`) conservan `invalid_array`. Se quita el caso especial de `parseTimelineIndexEntry`.
- El límite usa una constante con nombre: la existente `TimelineMaxCatalogRecords` en Go y una nueva `TIMELINE_LIMITS.catalogRecords` en TypeScript, con el mismo nombre de concepto.

## Test determinista

El falso `pi` escribe el PID en un archivo temporal y lo renombra al destino (`mv` es atómico en el mismo directorio), así que el archivo aparece ya con su contenido.
