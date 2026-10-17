# Engram como memoria principal del proyecto

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por [PR #64](https://github.com/JhonHawk/tricell-hive/pull/64) (`4dc73c28`), 2026-10-17 |
| Tracker · GitHub Issues | Sin issue: nace de la investigación del 2026-10-16 sobre memoria nativa y Engram |
| Git | rama de prueba activa `feat/delegation-threshold-global` · commit por tarea y despliegue local (C1-A) · sin push ni merge |
| Verificación | comprobaciones `rg` por criterio · `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` |
| Siguiente paso | Ninguno: cambio cerrado y archivado |

## Petición

> algo que creo que esta afectando en general, es que cada host tiene su propia memoria, hay actualización de engram o podemos hacer research sobre si 1) preferimos engram sobre la memoria nativa) 2) ponemos el puntero de la memoria nativa del host como una cita en la memoria de engram
>
> Es research

(2026-10-16.) Tras la investigación, respuesta a R2 (qué hacer con los hallazgos):

> R2-A Planear la regla (Recommended)

(2026-10-17.) La opción decía: "Engram es la fuente de decisiones, restricciones y estado cuando el host lo tiene. La memoria nativa guarda lo propio del host y punteros a topic_key de Engram. Si se contradicen, se verifica contra las fuentes actuales y se corrige la copia vieja. Incluye corregir tool-dependencies.md:22."

Durante la planificación:

> por cierto, ya arregle lo de codex

(2026-10-17: Codex tiene ahora el servidor MCP `engram`, el plugin con sus hooks y la función `memories` encendida.)

## Objetivo

Claude Code, Grok y ahora Codex guardan su propia memoria nativa además de Engram, y ninguna de las dos lee a la otra. En una muestra de 33 hechos de la semana del 12 al 16 de octubre, 11 estaban viejos o se contradecían entre las dos memorias. En cam, una restricción (no llamar a `POST /register`) se guardó en las dos, y su reversión del 16 de octubre llegó a Engram pero no a la memoria de Grok, hasta que el usuario lo corrigió a mano.

Hive no tiene hoy una regla que diga qué memoria manda. La guía dice "persistent memory" en `global.md:31`, `git-workflow` y `flow-close` sin decir cuál es. El cambio agrega esa regla en `global.md`, en la sección de continuidad:
- Engram es la memoria persistente cuando el host lo tiene.
- La memoria propia del host guarda solo lo que concierne al host y un puntero al `topic_key` de Engram para cada hecho del proyecto.
- Ante una contradicción, se verifica el estado actual antes de actuar y se corrige la entrada vieja.

## Alcance y aceptación

Incluye:
- Un párrafo nuevo en la sección "Continuity" de `content/guidance/global.md`.
- La fila de Engram en `_support/docs/architecture/tool-dependencies.md:22`.

Queda fuera:
- **Apagar la memoria nativa:** R2-B se descartó.
- **Guardar en Engram punteros a la memoria nativa:** era la opción 2 de la petición, y la investigación recomendó la dirección contraria.
- **Migrar o barrer las memorias nativas existentes:** `global.md:39` prohíbe migrar memoria; las entradas viejas se corrigen solo cuando una tarea las toca.
- **Registrar el origen de cada restricción en la memoria:** es H2, que el usuario decidió vigilar (R1-A).
- **Cambios de configuración en los hosts.**

Restricciones:
- La suite local pasa.
- El texto se escribe para los modelos del perfil `execution`, sin nombres de host.
- El commit nombra las sesiones de origen (host, ID y fecha) y la decisión R2-A, como pide AGENTS.md.
- El pull request declara cuánto crece `global.md`.

- AC1. "Continuity" en `global.md` dice que la memoria persistente de la guía es Engram cuando sus herramientas funcionan en la sesión y, si no, la memoria propia del host, que no son sus archivos de instrucciones. Si Engram está configurado pero no conecta, no autentica o no puede escribir, se informa y no se pasa en silencio a la memoria del host. *Falso en la base cuando* `rg -c -F 'Persistent memory in this guidance means Engram' content/guidance/global.md` no encuentra coincidencias.
- AC2. La misma sección dice que, cuando el agente escribe memoria y el host tiene las dos, las decisiones, restricciones y el estado del trabajo o de la entrega del proyecto van a Engram, cada uno con su propio `topic_key`. La memoria propia del host guarda solo lo que concierne a ese host y, para un hecho del proyecto, el puntero `engram:<project>/<topic_key>` en lugar de una copia. Esto prevalece sobre las instrucciones de memoria del host, salvo cuando el usuario pide guardar algo en la memoria propia del host. *Falso en la base cuando* `rg -c 'topic_key' content/guidance/global.md` no encuentra coincidencias.
- AC3. La misma sección dice que la captura automática del host puede copiar hechos del proyecto. Cuando dos memorias, o una memoria y una fuente actual, se contradicen, se actúa sobre una y se corrige una entrada solo si una fuente actual resuelve cuál está vieja; si no, se informan las dos sin editar. En Engram la corrección actualiza la entrada y sigue sus avisos de conflicto. En la memoria del host, donde se puede editar, reemplaza solo ese hecho por el puntero, y la regla autoriza esa edición. Solo se tocan las entradas de la tarea. *Falso en la base cuando* `rg -c 'disagree' content/guidance/global.md` no encuentra coincidencias.
- AC4. La fila de Engram en `tool-dependencies.md` nombra como hosts relevantes a Claude, Codex, Cursor, OpenCode y Pi, y dice que Grok lo recibe al leer la configuración MCP de Claude. *Falso en la base cuando* la línea 22 de `_support/docs/architecture/tool-dependencies.md` no contiene `Pi`.

## Entrega

- **Modo:** C1-A, elegido en esta sesión para el cambio anterior de este repositorio. Cada tarea se commitea en `feat/delegation-threshold-global` cuando llega a `[x]` y luego se despliega en local con `hive update`. Sin push ni merge.
- **Revisión de código:** ninguna dedicada, porque son cambios solo de texto, según la preferencia guardada para Hive.
- **Registro:** esta carpeta queda sin versionar hasta el merge, como dice `flow-close`.
