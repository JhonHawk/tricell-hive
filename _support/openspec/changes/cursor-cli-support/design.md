# Diseño — soporte de Cursor CLI

## Contexto verificado

Observado el 2026-09-25 con `cursor-agent` 2026.09.18-9a7762b, modelo Grok 4.6 High Fast (el configurado por defecto), `--print --mode ask`, fixtures desechables fuera de cualquier repositorio y códigos canario. Una corrida por caso.

| Superficie | Resultado | Fuente |
| --- | --- | --- |
| `AGENTS.md`, `CLAUDE.md` y `.cursor/rules` (`alwaysApply`) del proyecto | Cargados (3/3 canarios) | Corrida de inventario; [docs CLI](https://cursor.com/docs/cli/using) |
| `~/.claude/CLAUDE.md` | No cargado | Corrida de inventario |
| `~/.cursor/rules/*.mdc` | No cargado (`context7.mdc` ausente) | Corrida de inventario; hilo del foro de Cursor, staff 2026-07-02: "no ETA" |
| Canal global documentado | User Rules en Settings, no es archivo | [docs rules](https://cursor.com/docs/rules) |
| Skills | Descubre `.agents/skills`, `.claude/skills`, `.codex/skills` (changelog marzo 2026); los skills de Hive aparecen | Corrida de inventario; [changelog CLI](https://cursor.com/docs/cli/changelog) |
| Agentes del proyecto (`.cursor/agents`) | Visibles | Corrida de inventario |
| Agentes de usuario (`~/.claude/agents`, `~/.cursor/agents`) | No visibles | Corrida de inventario; Engram #11446 (bug del CLI desde mayo de 2026) |
| Puntero en `AGENTS.md` del proyecto | El agente leyó `~/.claude/CLAUDE.md` primero y respondió `_support/workspace/2026-09-25-<work>/`; el control respondió `/tmp/...` | Dos corridas en paralelo |

Gestor (inspeccionado en el working tree de `rebuild/harness-engineering`):

- Los hosts están fijos en el `switch` de `resolve` (`tooling/management/plan.go:23`), `validateHosts` (`tooling/management/types.go:163`, que también valida los consumidores al leer el estado en `ownership.go:72`), la lista de directorios de `plan.go:517` y la ayuda de `tooling/cli/main.go:28` y `:55`. `tooling/cli/setup.go:83-90` y el runner de `tests/pilot/` tampoco conocen Cursor; quedan fuera de este cambio.
- `State` y `Record` no guardan `Config` (`types.go:31-49`), así que agregar `CursorHome` no exige subir la versión del esquema. Los planes v4 guardados antes del cambio sí fallan con `invalid root` al aplicarse y hay que regenerarlos.
- Actualizar un recurso compartido exige incluir a todos sus consumidores en `--hosts` (`plan.go:467`). `~/.cursor/` ya existe y es de Cursor: Hive solo crea y retira `AGENTS.md` y `agents/<role>.md` dentro de él (`plan.go:426`, `files.go:280`, `apply.go:45-73`).
- `integrations/pi/pi.go` es el patrón más cercano: bloque en un archivo propio del host, skill compartido en `~/.agents/skills` y rechazo del scope proyecto.
- `integrations/agents/agents.go:126` exige exactamente cinco hosts en `agent-profiles.json`, y el renderizado y la validación de claves de acceso son por host (`agents.go:142-182`, `:222-280`).
- Contrato de bloques, consumidores compartidos y recuperación: [deployment-manager.md](../../../docs/architecture/deployment-manager.md).

Supuesto no verificado: Cursor no documenta una variable de entorno para mover `~/.cursor`. T1 lo comprueba en la documentación antes de fijar la ruta.

## Diseño elegido

### Guía global en `~/.cursor/AGENTS.md`

Nuevo adaptador `integrations/cursor` con destino `Kind: "block"` en `<cursor-home>/AGENTS.md`, con el mismo bloque gestionado y marcadores que los demás hosts. Cursor no lee ese archivo por sí mismo; lo lee porque el puntero lo indica. Si una versión futura de Cursor lo carga de forma nativa, el puntero condicional deja de dispararse y no hay duplicado.

Se prefiere una ruta propia de Cursor en vez de reutilizar `~/.claude/CLAUDE.md` para que el ciclo de vida de Cursor no dependa de que el consumidor Claude siga instalado (D1-A).

`target.Config` gana `CursorHome` (`json:"cursor_home,omitempty"`, para no cambiar el `planID` de journals anteriores), con valor por defecto `<home>/.cursor` y resolución canónica como los demás hosts. Solo tiene variable de entorno si T1 encuentra una documentada.

### Skills

El adaptador registra a Cursor como consumidor de los skills compartidos en `~/.agents/skills` (vía `target.ExpandSkills`, como Pi). Así, quitar otro consumidor no borra skills que Cursor sigue usando. No se crean alias en `~/.cursor/skills`.

### Roles en `~/.cursor/agents/` (D7-C)

El adaptador llama a `target.ExpandAgents` con destino `<cursor-home>/agents/<role>.md`, como Claude y Pi. Cursor pasa a ser el sexto host de `integrations/agent-profiles.json`. `Render` no necesita una rama nueva, y `plan.go` tampoco para agentes (`plan.go:366`, `:447-461` son genéricos).

Compatibilidad con versiones anteriores: `ReadProfiles` exige hoy exactamente cinco hosts (`agents.go:126-129`) y se llama al validar cada versión (`plan.go:295`). Si pasara a exigir seis, ninguna versión instalada antes de este cambio (las ocho retenidas tienen perfiles de cinco hosts) podría volver a instalarse con `--release`, rompiendo la vuelta atrás que documenta `deployment-manager.md:24`. Por eso `ReadProfiles` acepta los cinco hosts originales como obligatorios y `cursor` como opcional, completo si está presente. Cursor sobre una versión anterior falla en `Render` con `unsupported agent host "cursor"` (`agents.go:222-225`), un error claro. Al dejar de contar exactamente cinco hosts, `ReadProfiles` rechaza de forma explícita cualquier clave fuera de `{claude, codex, grok, pi, opencode, cursor}`, para que una errata no pase la validación.

Validación del perfil de Cursor: `cursor` entra en la lista de hosts que no admiten `effort` (`agents.go:142`), porque Cursor no documenta esa clave; su única clave de acceso es `readonly`, y solo con el booleano JSON `true` (omitirla ya equivale a `false`).

Formato de Cursor ([docs subagents](https://cursor.com/docs/subagents), consultado el 2026-09-25): frontmatter YAML con `name`, `description`, `model` (`inherit` o un ID con parámetros entre corchetes, como `claude-opus-5[effort=high]`), `readonly` (booleano) e `is_background`, seguido del cuerpo del rol. El renderizador actual escribe los valores con `json.Marshal`: produce claves ordenadas y escalares entre comillas (`model: "inherit"`), con `<`, `>` y `&` escapados como secuencias Unicode de JSON (por ejemplo, `<` como la barra invertida seguida de `u003c`). Es YAML válido y equivalente; que el parser de Cursor lo acepte está documentado como YAML pero no observado, y A8 lo comprueba.

Perfil inicial de Cursor en `agent-profiles.json`:

- **Modelos:** `execution`, `reasoning` e `inherit` renderizan `model: inherit`. Es un supuesto de este plan, no una decisión tuya: los IDs de modelo de Cursor dependen del plan de suscripción y la documentación advierte que Cursor reemplaza el modelo configurado cuando el plan no lo incluye. Elegir modelos por perfil queda para después de A8.
- **Acceso:** `observe` renderiza `readonly: true`; `implement` y `verify` no agregan campos. `readonly` es la única clave de acceso válida para Cursor.

Duplicados y precedencia, según la documentación: Cursor lee roles de usuario en `~/.cursor/agents/`, `~/.claude/agents/` y `~/.codex/agents/`, y cuando varios tienen el mismo nombre gana `.cursor/`. Los roles de Hive usan el mismo nombre en los tres, así que la copia de `~/.cursor/agents/` tapa a las otras dos. Que no aparezcan duplicados es una inferencia de esa precedencia, no una afirmación de la documentación ni una observación: el CLI 2026.09.18 no lista roles de usuario. A8 lo comprueba. Si se retira solo Cursor con Claude instalado, Cursor pasaría a los roles renderizados para Claude en `~/.claude/agents/`. Un hijo genérico que recibe el contrato leído del archivo no queda restringido por su `readonly: true`; `global.md:95` ya advierte que leer el frontmatter no aplica sus permisos.

`~/.cursor/agents/` ya existe y lo creó Cursor (vacío); Hive solo crea y retira sus archivos dentro de él.

**Delegación en `global.md` (D10-A, sustituye a la fila de Cursor):** la tabla por host sale de la guía siempre cargada. La regla por capacidad de `global.md:80` gana una frase: "A role the tool does not list counts as lacking role selection; read its contract from the installed role file." Cubre el caso "rol instalado pero no listado" de Cursor sin nombrar la herramienta. El piloto del [#30](https://github.com/JhonHawk/tricell-hive/issues/30) (2026-09-25, una corrida por celda) no mostró diferencias entre la guía con tabla y esta variante en los seis hosts; Cursor 2026.09.23 aplicó la alternativa correcta con el contrato leído de `~/.claude/agents/`. La tabla, con la fila de Cursor (`Task` con `subagent_type`, confirmado en el bundle `1218.index.js`), queda como referencia de mantenimiento en `_support/docs/architecture/agent-delivery.md`.

### Puntero en el `AGENTS.md` de cada proyecto

Texto canónico, en inglés porque es contenido de instrucciones:

```markdown
Cursor sessions: unless your loaded instructions contain the line "# Tricell Hive guidance" as a heading of its own, read `~/.cursor/AGENTS.md` before any other action and follow it. If that file is missing, say so and continue.
```

- **Ubicación:** fuera de `## Hive`, cerca del principio del archivo. El prefijo no es `Hive guidance:` para no confundirse con el ajuste `Hive guidance: required` (`global.md:99`), que usa HA-IF-13.
- **Condición:** depende del encabezado como línea propia, no de cualquier mención, porque el `AGENTS.md` de un proyecto puede mencionar la guía por su cuenta (este repo lo hace en la sección "Workspace locations") y el propio puntero la nombra dentro de una frase. En los hosts que ya cargan la guía, el bloque desplegado tiene `# Tricell Hive guidance` como línea propia (línea 2 en los cuatro archivos globales instalados), así que no leen el archivo. El prefijo `Cursor sessions:` orienta además a los demás hosts. A7 lo comprueba solo en las sesiones principales de Claude Code y Codex; no se afirma nada sobre hijos sin capa global.
- **Regla:** un repositorio con `Hive guidance: required` que se usa con Cursor necesita el puntero; sin él, esa declaración es falsa para las sesiones de Cursor.
- **Hogar canónico:** `content/skills/harness-audit/references/instruction-files.md`, como regla HA-IF-18 junto a HA-IF-17. Ausencia: `medium`, o `high` cuando el repositorio declara `Hive guidance: required`, por el criterio de `high` "a contradiction with consequence" (`SKILL.md:47`): esa declaración es falsa para las sesiones de Cursor. Propone la línea exacta. Como HA-IF-17, la ausencia no tiene outcome propio; es una laguna previa de la tabla de outcomes que este cambio no resuelve. Puntero existente: `keep`, para que una auditoría posterior no lo quite como puntero redundante (HA-HI-07). La documentación de T3 enlaza a esa referencia en vez de copiar el texto.
- **Detección de Cursor:** el usuario lo nombra, o hay historial del CLI para ese punto de entrada. Un directorio `.cursor/` en el repositorio es solo un indicio para preguntar (el IDE lo crea y el CLI no lo necesita). La conducta de carga de Cursor (qué lee y qué no) va en `content/skills/harness-audit/references/hierarchy.md`. El significado de `~/.cursor/projects/<ruta codificada>/` como historial no está verificado; T2 lo comprueba antes de citarlo.
- No va en `global.md`: Cursor no lo cargaría, y los demás hosts no lo necesitan.

### Fuera del diseño

Scope proyecto del gestor para Cursor (D2-B, descartado: chocaría con el bloque de Codex en el mismo `AGENTS.md`), reglas `.mdc` por proyecto (D1-B) y User Rules (D1-C).
