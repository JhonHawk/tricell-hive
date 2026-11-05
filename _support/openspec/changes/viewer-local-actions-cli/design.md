# Diseño

## Enfoque (D13 A)

Opciones consideradas:

1. **CLI como `local_viewer` con origen declarado (elegida).** Da paridad completa sin reglas de autoridad nuevas; la barrera entre persona y agente pasa del código a la guía, como ya ocurre con un push o un merge.
2. **Autoridad nueva para productores nativos.** Reasignación automática por inactividad o PR fusionado; choca con la spec («no inferir que una sesión murió») y no cubre borrar.
3. **Solicitud por CLI y aprobación en el visor.** Mantiene la barrera pero sigue necesitando la UI.

Las cinco acciones son todas las que hoy solo hace el navegador: el servidor expone tres rutas de escritura con `local_viewer` (`PATCH` de la sesión para título, archivo y fijación, `DELETE` y `POST …/recover-lead`).

## Contrato

- El actor `local_viewer` acepta `origin` opcional, `browser` o `cli`, por defecto `browser`. Es procedencia declarada, no autenticación: la CLI y el navegador usan el mismo token.
- Dónde queda el origen (D18 A): en el evento público de renombrar y de reasignar; en el recibo administrativo de archivar, restaurar y fijar; en una auditoría privada del almacén que sobrevive al borrado. Solo reasignar lleva motivo.
- Capacidad (D17 A): la revisión del contrato no sube. `/api/health` anuncia `capabilities: ["local_actions_cli"]`; la CLI la lee antes de enviar y, si falta, no envía y pide actualizar el visor. Un visor viejo además rechaza `origin` como campo desconocido sin cambiar nada.
- El validador y el tipo del actor en el frontend aceptan `origin`; sin eso la UI dejaría de abrir trabajos con acciones de la CLI. Los casos compartidos de `tooling/viewer/testdata/timeline_contract/` cubren el actor con y sin origen.
- El esquema publicado por `hive view contract` (`timeline_protocol.go`) admite `origin`, y sus textos de uso dejan de decir «Local viewer only».
- Vuelta atrás: un binario anterior no lee un registro guardado con `origin`; volver atrás después de usar los comandos exige restaurar la copia del estado. Se documenta en `deployment-manager.md`.

## Reintento (D19 A)

Sin diario: tras una respuesta incierta, el comando relee el trabajo y dice si la acción quedó. La versión esperada hace fallar un reintento de renombrar, archivar o reasignar, y un trabajo borrado responde 410, así que nada se aplica dos veces. La spec exime a estos comandos del diario de publicación.

## Comandos

```text
hive view work rename  --work-id ID --title TEXT
hive view work archive --work-id ID
hive view work restore --work-id ID
hive view work pin     --work-id ID
hive view work unpin   --work-id ID
hive view work recover --work-id ID --request-id ID --reason TEXT
hive view work delete  --work-id ID --confirm ID
```

- No llevan `--host` ni `--native-session-id`: no actúan como productor. El encabezado del contrato embebido deja de decir que toda mutación los requiere.
- `recover` toma la solicitud pendiente indicada, del mismo proyecto. Antes, la sesión nueva pide unirse con `hive view producer request`; después, para cerrar, completa la etapa activa y cierra.
- `delete` exige `--confirm` igual a `--work-id`.
- `hive view session update` y `hive view session delete`, que hoy siempre fallan para estos campos, apuntan a los comandos nuevos en su error.
- Las órdenes de publicación nativa siguen rechazando un actor `local_viewer` en su JSON.

## Guía

`viewer-publication.md` es la casa de la regla: un agente ejecuta estas acciones solo cuando el usuario las pide nombrando el trabajo. Borrar es un efecto crítico que se pregunta cada vez, aun cuando el usuario lo pidió sobre ese trabajo: esta excepción cambia la frase «unless the user asks for it on a named target» de «Scope and authorization» solo para borrar trabajos del visor. La reasignación sigue sin inferir que la sesión conductora terminó. Se corrigen el paso 3 («Pass your verified `--host`… on mutations») y «Report that required human action», y la frase que prohíbe a los publicadores el actor `local_viewer`.

Al reiniciar el visor tras integrar, su token cambia y las pestañas abiertas pierden acceso: el reporte lo dice y ofrece «abre el visor».
