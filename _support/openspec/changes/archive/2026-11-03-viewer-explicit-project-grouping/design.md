# Design

## Decisions

Reutilizar configuración explícita y lista directa (D1 A, D2 A). Decisiones técnicas propuestas: extensión opt-in del índice para preservar consumidores estrictos; agrupación calculada, sin migración; comparación exacta del nombre después de quitar espacios exteriores, sensible a mayúsculas, sin imitar alias ni normalización completa de Engram. Esto define la igualdad de presentación, no una nueva identidad de publicación.

## Proposed shape

### Configuration and registration

Leer únicamente `<registered source root>/.engram/config.json`; no ascender ni buscar hijos. La raíz de cada recorrido es su Source original canonizado, nunca la del último colaborador. Usar el mismo lector nuevo para agrupación y elegibilidad del registro no Git; no endurecer silenciosamente el lector legado de nombres visibles.

La carpeta .engram y el archivo deben ser reales y no enlaces, con archivo regular de hasta 64 KiB. Exigir un único objeto JSON y `project_name` string, no vacío tras trim, de hasta 200 caracteres Unicode, sin controles ni separadores `/` o `\`. Otros campos se ignoran y nunca se imprimen ni se usan. No truncar ni quitar controles para fabricar una identidad válida. Ausencia es normal; JSON inválido, tipo incorrecto, exceso de tamaño, enlace o fallo de lectura son diagnóstico fijo, sin contenido del archivo.

`ResolveRoot` mantiene el camino legacy de `_support/workspace` y Git. Si no encuentra raíz Git y la entrada exacta tiene configuración explícita válida, devuelve esa entrada canónica. No busca padres ni promueve hijos. Una configuración en un subdirectorio Git no sustituye el Git toplevel. Registro fallido por configuración inválida explica el archivo esperado sin escribir; volver a registrar una raíz no Git válida conserva la entrada y su added_at. Los registros existentes no se eliminan aunque después desaparezca o se invalide la configuración. La ruta legacy `_support/workspace` sigue funcionando sin Engram.

### Navigation contract

Mantener protocolo de publicación 3, schemas de sesión/índice legacy, almacenamiento y registry sin cambio. `GET /api/index` conserva exactamente su forma. Sólo `GET /api/index?navigation=grouped` incorpora el campo opcional `navigation`:

```json
{
  "navigation": {
    "version": 1,
    "entries": [{
      "viewer_session_id": "<existing-id>",
      "group_key": "explicit:<sha256-of-trimmed-name>",
      "group_name": "example-project",
      "origin_kind": "repository",
      "origin_label": "example-api",
      "config_status": "valid"
    }]
  }
}
```

`entries` contiene exactamente una entrada por sesión del mismo índice, incluidas archivadas; orden igual al índice. Una sola lectura de configuración por raíz distinta por petición, sin recursión ni escrituras, para que todos sus recorridos usen la misma observación. `group_key` para configuración válida: prefijo explicit y SHA-256 hex de los bytes UTF-8 del valor trim. Fallback: prefijo project y hash del project_key, con el project_name de la primera fila fallback de esa clave en el índice como nombre consistente para toda la petición. Los namespaces impiden colisiones con claves antiguas. Sin nombres derivados del archivo como rutas de disco.

`origin_kind` es `repository` cuando existe marcador Git válido en esa raíz, `workspace` para raíz accesible no Git y `unavailable` si la raíz no está disponible. `origin_label` es basename de la raíz original, con la ruta completa accesible desde la fuente ya presente en el índice para desambiguar nombres coincidentes. `config_status`: `valid`, `missing`, `invalid`, `unavailable`. Missing usa fallback sin aviso. Invalid/unavailable usan fallback con aviso localizado; no ocultan sesiones ni detienen publicación. La configuración de otra raíz con el mismo project_key no se hereda.

La proyección se recalcula en el refresco del índice; cambios sólo afectan navegación. No escribe los JSON históricos. No modifica `timelineProject`, binding, hash de sesión, elegibilidad de colaboradores, membresías, fuentes ni blobs. No usa el MCP, DB, CLI ni alias de Engram. Raíces registradas ajenas con el mismo nombre explícito sí se agrupan conforme a D1; esto no confiere permisos.

El nuevo frontend solicita la opción. Un servidor antiguo la ignora y devuelve el índice legacy: frontend sin navigation conserva navegación anterior. Un cliente antiguo no pide la opción y recibe legacy. Validar metadata como contrato independiente: versión conocida, IDs únicos y conjunto igual al índice recibido, tipos/enums y tamaños acotados. Una proyección inválida se descarta completa con aviso y grupos legacy; no debilitar el parser de sesiones ni ocultar recorridos válidos. IDs extra o duplicados nunca dirigen selección. Mantener la respuesta y sus metadata juntas durante polling; no mezclar versiones.

### Navigation preferences

Selección sigue por viewer_session_id. Expansión se guarda por nueva group_key; para fallback se consulta la preferencia legacy si falta la nueva. Primer grupo explícito sin preferencia se abre si cualquiera de sus grupos legacy estaba abierto o contiene la selección; después aplica preferencia explícita. Reagrupación conserva selección y abre su destino para que siga visible. No borrar preferencias ajenas. Filtros y conteos se calculan sobre las mismas filas visibles; archivados/pinned conservan sus reglas. No ampliar búsqueda a campos nuevos en esta entrega.

## UI

- Superficie operational; arquetipo existente de navegación lateral de recorridos (se conserva en vez de imponer otro patrón del catálogo).
- Container family Frame-led existente; densidad actual, sin alterar dimensiones del Sidebar ni estilos de componentes.
- Regiones: cabecera existente, buscador Input actual, aviso inline cuando corresponda, SidebarMenu con Collapsible por grupo y SessionNavigationButton por recorrido.
- Componentes: SidebarMenuButton variante y tamaño actuales por defecto; Collapsible instalado sin retiming; Input existente; Skeleton existente; Alert instalado variant=warning para agrupación degradada y destructive para error total; Button variant=outline size=sm para Retry (props verificados en los componentes instalados). No instalar componentes.
- Acción primaria de la lista: seleccionar recorrido. Expansión de grupo conserva su trigger; sin nuevo menú de gestión.
- Una línea meta text-xs con origen, ticket y fecha, conservando título text-sm. En grupos explícitos el origen aparece siempre. Workspace se localiza y la ruta/basename queda disponible en tooltip accesible; nombres largos truncan visualmente con texto completo accesible.
- Motion: sin entrada nueva, componentes conservan movimiento instalado.

Antes (fixture de dos recorridos):

```text
Proyectos                         [Archivo]
[Buscar título o ticket                  ]
▾ example (workspace .)                  1
  Auditoría de tests
  Sin ticket · 2 nov
▾ example (workspace ..)                 1
  EX-914
  EX-914 · 2 nov
```

Después (Sidebar + Collapsible, lista directa):

```text
Proyectos                         [Archivo]
[Buscar título o ticket                  ]
▾ example-project                       2
  Auditoría de tests
  Workspace · Sin ticket · 2 nov
  EX-914
  example-api · EX-914 · 2 nov
```

Estado degradado, mismos recorridos accesibles:

```text
[!] No se pudo leer la agrupación
    Se muestran los proyectos originales.
    [Reintentar]
▾ example (workspace .)                  1
  Auditoría de tests
```

Loading: Skeleton en estructura real de grupos/filas, conservar contenido previo durante refresco. Empty por búsqueda: texto actual y acción limpiar búsqueda; colección vacía: mensaje actual de publicación y acción Reintentar (no fabricar una sesión). Error total: Alert y Retry existentes; error de agrupación: aviso inline con Retry y datos legacy. Sin acceso: tratamiento existente del controlador, nunca fallback que evada autenticación. Permisos no cambian. Desktop 1440×900 y mobile 440×956, light/dark; teclado, truncado, expandir, selección, filtro, archivados y cambio de configuración.

## Copy

- `Workspace` / `Workspace`.
- `Source unavailable` / `Origen no disponible`.
- `Could not read project grouping` / `No se pudo leer la agrupación`.
- `Showing the original projects.` / `Se muestran los proyectos originales.`
- `Retry` / `Reintentar` (reutilizar clave existente).

Nombres y títulos del productor quedan verbatim; no convertir project_name en título decorativo ni quitar sufijos heurísticamente.

## Impact

Server.index, frontend api/controller/shell/types/tests, registry y helper de configuración; CLI register sólo por resolución/error. Espec walkthrough-viewer, manual viewer y descripción arquitectónica del manager. README y llms.txt se revisan y actualizan juntos sólo si cambia una afirmación existente; no se altera el formato de instrucciones globales ni las skills. No hay cambio de producción ni de otros repositorios durante build. El maintainer refresh, si aplica a la entrega autorizada, se rige por AGENTS y preserva puerto/state/TTL.

## Recovery

No hay migración. Volver al binario anterior conserva registros y publicaciones; el índice legacy sigue disponible. Las preferencias adicionales del navegador son aditivas. Volver a nombres de configuración anteriores restaura agrupaciones en el siguiente refresco, pero el build no edita esos archivos del usuario. No usar reset/flush, no eliminar registros para simular una solución.
