export const language = navigator.language.toLowerCase().startsWith("es")
  ? "es"
  : "en";
export const messages =
  language === "es"
    ? {
        indexFailure:
          "No se pudo actualizar la biblioteca. Puedes seguir leyendo la copia abierta y reintentar.",
        access: "Esta página necesita acceso",
        missing: "Ejecuta hive view open para abrir esta página con acceso.",
        rejected:
          "El acceso fue rechazado o el código de apertura caducó. Ejecuta hive view open para volver a abrir esta página.",
        pending: "Comprobando acceso…",
        down: "No se puede conectar con Hive. Inicia el visor con hive view start y vuelve a abrirlo con hive view open. La página se reconectará si el servidor responde aquí.",
        reconnected: "Reconectado con el servidor.",
        gone: "El recorrido abierto ya no está disponible. Se muestra otro recorrido.",
        pinned: "Copia conservada; no caducará.",
        unpinned: "La copia caducará con el resto.",
        deleted:
          "Se borró la copia de Hive. Los archivos originales se conservaron.",
        writeError: "No se pudo completar la operación. Inténtalo de nuevo.",
        dismiss: "Descartar",
        retry: "Reintentar",
        excluded: "Copia apartada",
        excludedBody:
          "El manifiesto fuente está versionado en Git. Su documento y sus archivos no se pueden abrir aquí. Puedes seguir conservando o borrando esta copia.",
        source: "Carpeta fuente",
        sourceLocal: "Archivo local; autoría no verificada",
        sourceUnknown: "Origen sin determinar; autoría no verificada",
        absent:
          "No se encontró la carpeta del repositorio. Sus copias se conservan.",
        noRepositories: "No hay repositorios registrados",
        register:
          "Ejecuta hive view register dentro del repositorio para registrar sus recorridos.",
        empty: "Aún no hay recorridos",
        emptyBody:
          "Los recorridos nuevos aparecen tras guardar walkthrough.json en _support/workspace/<session> del repositorio.",
        repositories: "Repositorios registrados",
        copy: "Copiar comando",
        copied: "Copiado",
        loadFailed: "No se pudo cargar este recorrido.",
        loading: "Cargando recorrido…",
        ttl: "Las copias caducan después de",
        unlessKept: "salvo que se conserven",
        ingestErrors: "Errores de ingesta",
        confirm: "¿Borrar esta copia?",
        confirmBody:
          "Solo se borra la copia que guarda Hive. Los archivos del repositorio no se modifican.",
        cancel: "Cancelar",
        remove: "Borrar copia",
        keeping: "Conservado; no caducará",
        sourceReasons: {
          git_query_failed: "Git no pudo comprobar la fuente.",
          git_output_incomplete: "Git devolvió un resultado incompleto.",
          source_unavailable: "La fuente no está disponible.",
          tracking_unavailable:
            "No se pudo comprobar el tracking de la fuente.",
        },
      }
    : {
        indexFailure:
          "Could not refresh the library. You can keep reading the open copy and retry.",
        access: "This page needs access",
        missing: "Run hive view open to open this page with access.",
        rejected:
          "Access was rejected or the opening code expired. Run hive view open to reopen this page.",
        pending: "Checking access…",
        down: "Cannot reach Hive. Start the viewer with hive view start, then reopen it with hive view open. This page reconnects if the server answers here again.",
        reconnected: "Reconnected to the server.",
        gone: "The open walkthrough is no longer available. Showing another walkthrough.",
        pinned: "Copy kept; it will not expire.",
        unpinned: "The copy will expire with the rest.",
        deleted:
          "The Hive copy was deleted. The original files were preserved.",
        writeError: "Could not complete the operation. Try again.",
        dismiss: "Dismiss",
        retry: "Retry",
        excluded: "Copy set aside",
        excludedBody:
          "The source manifest is tracked in Git. Its document and files cannot be opened here. Keeping or deleting this copy is still available.",
        source: "Source folder",
        sourceLocal: "Local file; authorship not verified",
        sourceUnknown: "Source undetermined; authorship not verified",
        absent: "The repository folder was not found. Its copies are kept.",
        noRepositories: "No repositories registered",
        register:
          "Run hive view register inside the repository to register its walkthroughs.",
        empty: "No walkthroughs yet",
        emptyBody:
          "New walkthroughs appear after saving walkthrough.json under the repository’s _support/workspace/<session>.",
        repositories: "Registered repositories",
        copy: "Copy command",
        copied: "Copied",
        loadFailed: "Could not load this walkthrough.",
        loading: "Loading walkthrough…",
        ttl: "Copies expire after",
        unlessKept: "unless kept",
        ingestErrors: "Ingest errors",
        confirm: "Delete this copy?",
        confirmBody:
          "Only the copy stored by Hive is removed. The repository files are not modified.",
        cancel: "Cancel",
        remove: "Delete copy",
        keeping: "Kept; it will not expire",
        sourceReasons: {
          git_query_failed: "Git could not check the source.",
          git_output_incomplete: "Git returned an incomplete result.",
          source_unavailable: "The source is unavailable.",
          tracking_unavailable: "Source tracking could not be checked.",
        },
      };
