# viewer-explicit-project-grouping — Agrupación del visor por identidad explícita

| Campo | Valor |
| --- | --- |
| Estado | Completed · PR #349 · development effe57fe |
| Tracker | Sin ticket; no solicitado |
| Git | tricell-hive · base development `2653b0b185572f9207d0b4196969891b21008f7c` · rama propuesta `feat/viewer-explicit-project-grouping` |

## Decisiones

2026-11-03
- ¿Cómo agrupar? Reutilizar el mismo `project_name` explícito válido en `.engram/config.json` de raíces registradas; sin configuración mantener la agrupación actual. Una edición posterior cambia la agrupación del historial, no sus identidades ni permisos. (D1 A)
- ¿Cómo presentar? Lista directa dentro del grupo, con origen en cada fila; conservar Sidebar y Collapsible instalados, sin subgrupos. (D2 A)

- ¿Cómo entregar? Interactiva: validación visual antes de push, PR y merge. (D3 A)
- ¿Qué revisión de código? Nativa Codex `/review`. (D4 A)
- ¿Dónde implementar? En worktree aislado con rama propia, por instrucción directa «y worktree»; crearlo al comenzar el build. (D5)

## Solicitud

El usuario reportó proyectos duplicados en el visor y pidió investigar las sesiones de su workspace.

> hagamos research para una propuesta

> talvez podamos validar cómo lo hace engram

> $flow-plan

## Objetivo

Dos raíces explícitamente configuradas como el mismo proyecto se muestran hoy en grupos distintos porque la navegación usa `project_key`. Añadir una agrupación de presentación derivada de la configuración explícita, preservando la identidad de publicación y colaboración; hacer registrable directamente una raíz no Git con esa configuración válida.

## Alcance y aceptación

Incluye lector validado de configuración, proyección opcional del índice, registro directo de raíz no Git, agrupación y origen en navegación, pruebas y documentación del visor. No modifica configuraciones de otros repositorios, Engram, datos históricos, claves de proyecto ni controles de colaboración. No incorpora descubrimiento recursivo, alias de Engram, herencia arbitraria de padres, nuevas dependencias, un editor de grupos ni agrupación entre repositorios por nombres de AGENTS.

- AC1 — Dos recorridos existentes con claves distintas y raíces registradas que declaran el mismo `project_name` válido aparecen en un solo grupo con ese nombre y conteo correcto. Falsador actual: `shell.tsx:332` agrupa por `projectKey`; la auditoría y el otro recorrido de el workspace original aparecen separados.
- AC2 — Cada fila de un grupo explícito muestra el origen del recorrido junto a la información secundaria; selección, búsqueda por título/ticket y navegación por teclado alcanzan cada recorrido sin un nivel adicional de expansión. Falsador actual: las filas no reciben ni representan origen en `ViewerSession`.
- AC3 — `hive view register --dir <root>` acepta una raíz no Git con configuración explícita válida y el segundo registro es idempotente. Falsador actual: `ResolveRoot` rechaza esa raíz salvo que la entrada esté dentro de `_support/workspace`.
- AC4 — En el siguiente índice actualizado, cambiar el nombre explícito reagrupa los recorridos; eliminarlo restaura sus grupos originales; volverlo inválido conserva esos grupos y muestra un aviso recuperable. Falsador actual: el índice no proyecta configuración de agrupación ni sus diagnósticos.
- AC5 — Un cliente que solicita `GET /api/index?navigation=grouped` recibe la proyección de agrupación versión 1, y uno sin ese parámetro conserva el contrato actual; el nuevo cliente admite una respuesta antigua sin proyección. Falsador actual: `Server.index` ignora la opción y nunca devuelve la proyección.

Invariantes asociados a todos los criterios: ningún ID, project_key, Source, permiso, recibo, documento, artefacto o archivo de recorrido cambia como efecto de agrupar; no se autoriza un join antes rechazado. La ausencia de configuración conserva el comportamiento actual. Nombres iguales de proyectos sin declaración explícita nunca son prueba de pertenencia. Esos invariantes ya existen parcialmente; son protección de regresión, no criterios nuevos que aparentan fallar en la base.

## Prerrequisitos

Ninguno externo: el build usa fixtures sintéticos y dependencias ya instaladas. No necesita modificar el workspace original ni consultar el servidor de Engram. La validación visual del usuario se define con la entrega.

## Contexto verificado

- Base local y remoto development verificados en `2653b0b1` durante el research.
- `tooling/viewer/registry.go`: resolución Git o entrada `_support/workspace`; igualdad exacta de raíz para registrar.
- `tooling/viewer/timeline_store.go:77,246,569`: project_key participa en hash, restauración y autorización; no se sustituye.
- `tooling/viewer/frontend/src/api.ts:170`: índice estricto que rechaza campos desconocidos. Por ello la extensión será solicitada explícitamente.
- `tooling/viewer/index.go:144`: lector existente limitado a 64 KiB, sin enlaces, usado para nombre alternativo; no valida una identidad de agrupación.
- `tooling/viewer/frontend/src/shell.tsx`: patrón Sidebar/Collapsible existente, estilo base-nova y lucide.
- Tests existentes seleccionados de registro, identidad y colaboradores pasaron durante el research; no prueban la propuesta.
- El visor local reportó revisión `0606c4f3` y modified=true; no se presume equivalencia completa con HEAD ni se reinicia para planear.
- Archivos no versionados ajenos detectados: master-promotion-guard y dos carpetas de sesiones; se preservan.

## Entrega

Entrega interactiva: validar el candidato visual con el usuario antes de push; tras aceptación, push, PR y merge hacia development. Revisión de código: nativa Codex `/review`, además de la revisión UX obligatoria. El plan no inicia implementación ni autoriza cambios en el workspace original o migración de datos. La publicación queda condicionada al inicio del build y a la validación visual. El maintainer refresh se realiza sólo cuando la entrega autorizada y AGENTS lo indiquen.

## Estado de revisión y progreso

2026-11-03 — Revisión independiente hive-review-plan completada. Un hallazgo P2 sobre nombres MODIFIED corregido: el delta ahora apunta a User-controlled walkthrough server y Hierarchical React viewer navigation y preserva sus obligaciones y escenarios. Entrega interactiva, Codex /review y worktree aislado confirmados. Sin implementación.
