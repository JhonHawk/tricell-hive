# Diseño

## Orden de las comprobaciones

El validador de cada lenguaje sigue el orden en que su sesión ficticia alcanzaba los campos, de modo que toda entrada con un solo error da el mismo código que antes. Los 26 casos de T1 lo comprueban en los dos lenguajes.

- Go: puntero exigido por el estado, `viewer_session_id`, `source`, `project_key`, `project_name`, `version`, `title`, `summary`, `state`, `created_at`, `updated_at`, tickets (registro, identidad duplicada, URL repetida y luego límite), `validation_count`, id del puntero y puntero permitido por el estado. Una diferencia que los casos no cubren: antes, `source` y `created_at` se validaban primero, a través del productor ficticio, así que una entrada con varios errores puede dar ahora otro primer código.
- TypeScript: campos exactos, id del puntero, puntero exigido, los mismos campos que Go más `archived` y `pinned`, tickets (forma y límite, luego identidad), puntero permitido y `validation_count`.

Cuatro reglas ya daban un código distinto en cada lenguaje antes de este cambio (`archived` con tipo inválido, `validation_count` fuera de rango y más de 100 tickets). Sus casos fijan solo el rechazo.

## Tickets: una función por lenguaje

La validación de la lista de tickets sale de `validateTimelineCatalog` (Go) y `validateTimelineCatalogs` (TypeScript) a una función propia, que usan tanto la sesión como la entrada del índice. Sus reglas son: registro válido, identidad duplicada, URL repetida y como máximo 100. Las entregas y los candidatos siguen en la validación del catálogo de la sesión y comparten con los tickets el mismo registro de identidades, como antes.

## Lo que no cambia

`ValidateTimelineSession`, `parseTimelineSession`, el formato del índice y lo que emite el servidor. Los dos decodificadores conservan su firma y su tipo de retorno.
