# Diseño

## Contexto verificado (2026-10-25)

- **Transcripciones de cam en Codex** (`rollout-2026-10-20T01-07-13-01a15822-…`, `rollout-2026-10-23T01-14-31-01a1679c-…`).
  - La línea 1 trae, en las instrucciones base, «Avoid performing blocking sleep or wait calls longer than 60 seconds, as they may prevent you from communicating with the user for their duration» y «without a commentary update for more than 60 seconds».
  - La línea 4 trae, en el rol multiagente, «prefer longer waits (minutes) to avoid busy polling».
  - El hilo principal leyó `delegation.md` (líneas 22 y 1331) antes de su primera espera (línea 1610, 60 000 ms). Las 158 esperas usaron 60 000, con mediana de 23 s, porque cada mensaje de un hijo despierta la espera.
  - Los hijos llaman a `send_message` con destino `/root` unos 25 s después de empezar. Sus encargos y mensajes están cifrados.
- **Codex 0.164.0** (`codex --version`; código fuente de la etiqueta `rust-v0.164.0`).
  - `wait_agent`: por defecto 30 000 ms, mínimo 10 000 y máximo 3 600 000 (`core/src/config/mod.rs:258-260`). En v2, un valor menor al mínimo se sube al mínimo (`multi_agents_v2/wait.rs:57-65`).
  - `[features.multi_agent_v2]` acepta `min_wait_timeout_ms`, `max_wait_timeout_ms` y `default_wait_timeout_ms`, cada uno de 0 a 3 600 000. Al cargar se exige mín ≤ defecto ≤ máx (`config_mod.rs:3219-3232`). La tabla tiene `deny_unknown_fields`.
  - Una tabla sin `enabled` no cambia qué runtime elige el catálogo de modelos (`features.rs:961,1514-1518`).
  - En v2, la espera termina con cualquier mensaje o resultado final de un hijo, o cuando el usuario escribe (`multi_agents_spec.rs:306`).
  - `subagent_developer_instructions` «Overrides inherited developer instructions for subagents without role-specific instructions» (`fc.rs:313`).
  - No hay comando que guarde una clave anidada. `codex features` solo hace `list`, `enable` y `disable`, y `-c` vale para una sola ejecución.
  - La documentación oficial (https://developers.openai.com/codex/multi-agent, consultada el 2026-10-25) no menciona `wait_agent`. Lo de arriba viene del código fuente.
- **Hive.**
  - `integrations/codex/features.go`: `FeatureEnabled` ignora las tablas hermanas `[features.<otra>]` (`:59-67,73-97`). No tiene prueba para ese caso.
  - Regla fijada en `tooling/providers/types.go:19-20` («Hive never edits Codex's config»).
  - El precedente con propiedad y deshacer es el paquete de Pi: `pi_package.go:107-195`, `Plan.PiPackage` (`types.go:200`), `State.PiSubagentsSource` (`:133`), journal en `apply.go:17-36,393-400,668` y plan en `plan.go:606-645`.
  - Escritura atómica: `write()` en `files.go:82-122`, con `target.Safe` (`target.go:88-110`), que rechaza los enlaces simbólicos.
  - `update.go:65-66` y `tui_update_view.go:172-178` no muestran capacidades opcionales.
  - `doctor_integrations.go:169-179` no lee la configuración; la vista de integraciones reserva 7 filas (`tui_integrations_view.go`).

## Enfoque

1. **Guía (D19-A).** Una frase en `global.md:84`, que es su hogar. La misma frase se suma al cierre «Return evidence, limitations, and any decision needed from the parent.» que comparten los roles, para que en Codex llegue como instrucción de desarrollador. `delegation.md:65` remite a la guía global en vez de enunciarla. `agent-delivery.md` documenta la variante de Codex: el tope de 60 s de sus instrucciones base y el mínimo configurable.
2. **Lector y editor en `integrations/codex`.** Una función pura sobre los bytes del archivo:

   ```text
   WaitFloor(content []byte) (state, edit, err)
   state ∈ {Absent, Applied, Missing, Conflict, Unsupported}
   ```

   Inserta líneas sin tocar el resto. Reutiliza `stripTOMLComment`, `unquoteTOMLKey`, el corte del BOM y `maxConfigLine`. También da la operación inversa: dadas las líneas registradas, las quita solo si siguen iguales.
3. **Plan, estado y journal en `management`.**
   - La elección entra por `Options.CodexWaitFloor bool` (`types.go:243-246`), y `planCodexWaitFloor` la lee desde `plan.go`. Es lo que fija el plan; no es un campo que se añada después, porque `planID` se calcula al construir.
   - El plan lleva `Plan.CodexWaitFloor`: la ruta, las claves y valores previstos y la huella `Expected` del archivo. Un plan guardado con `--out` lleva solo eso.
   - La escritura es una entrada de journal de un tipo nuevo, «líneas». Guarda la huella previa y las líneas que inserta, nunca los bytes completos de `config.toml`, que puede tener secretos (por ejemplo en `[mcp_servers.*.env]`); los journals confirmados no se borran (`apply.go:18-24`).
   - Respaldo (D20-A): antes de la primera escritura, Hive copia el archivo a `config.toml.hive-backup` (0600) en la misma carpeta, con escritura atómica. Si esa copia ya existe, no la pisa: conserva la más antigua. El registro de `State` anota que Hive la creó. `plan remove` la borra solo después de revertir bien y solo si Hive la creó; si la reversión deja algo sin quitar, la conserva y lo informa. El journal nunca guarda esos bytes.
   - Cómo se confirma sin bytes guardados: la entrada lleva las huellas `Before` y `After` (`Fingerprint` ya existe). Al escribir, se comprueba que el archivo actual coincide con `Before`, se recalcula la edición con `WaitFloor` sobre ese contenido, se escribe con `write()` y se verifica contra `After` (en lugar de `same(cur, en.Before/After)` sobre bytes, `apply.go:376-401`). `readEntry` y `prepareEntries` (`migration.go:250`) reconocen el tipo nuevo.
   - Su inversa es por líneas, se calcula sobre el archivo actual y la usan tanto el deshacer como `Recover` (`apply.go:684-760`): quita solo las líneas registradas que siguen iguales. Si faltan o cambiaron, lo informa y sigue; nunca deja la transacción en «recovery conflict» por este archivo ni bloquea con `pending.json` (`apply.go:108-111,744-746`). Así no se atasca si Codex escribe el archivo entre la escritura y la recuperación.
   - Hay un registro en `State` con la ruta, las claves y valores escritos, si Hive creó la tabla y si creó el archivo. Se copia en `prepareTransaction` (`apply.go:205-216`) y entra en `PlanUnchanged` y su chequeo de «ya aplicado» (`migration.go:103-150`).
   - El esquema se queda en 6 con `omitempty`, como el precedente de `types.go:131-133`. Limitación aceptada: un binario anterior que reescriba el estado pierde el registro, y entonces `plan remove` deja las líneas en su sitio; se documenta.
   - `plan remove` (CLI y TUI, `tui_hosts.go:39`) deshace el registro cuando los hosts que quita incluyen `codex`.
4. **CLI.**
   - Una entrada al final del catálogo de proveedores. `adapter.Plan` no emite ningún paso externo para ella, y `ValidateStep` la acepta sin paso planeado; lo que hace la capacidad es fijar `Options.CodexWaitFloor`.
   - En `install.go`, el plan se vuelve a construir con `BuildPlan` después de `SelectProviders` (`:354,367`), y se recalcula `PlanUnchanged`.
   - En el adaptador, `Detect` lee primero el archivo y, si ya está aplicado, salta la oferta sin consultar la versión, también con `--dry-run`, donde no corre ningún programa. Si no lo está, consulta la versión (≥ 0.164 con un sondeo como `codexFeaturesProbe`). Con Codex anterior a 0.164 o sin `codex` hace un `continue` explícito, sin dejar un paso manual (`provider_adapter.go:138-141`). La marca `DefaultOn` en interactivo. El texto de efectos dice qué escribe y cómo revertirlo.
   - `update` (`planUpdate`, `update.go:113`) también detecta, ofrece y elige, solo en interactivo, y recalcula `unchanged` después de elegir. Cuando el plan escribe la configuración de Codex pide siempre confirmación, aunque el contenido de Hive no haya cambiado (`update.go:81-87`). La vista TUI de actualización sigue la misma ruta.
   - El doctor suma una fila que lee el archivo, que solo aparece con Codex registrado y que dice si existe `config.toml.hive-backup`. Avisa si la tabla existe y Codex es anterior a 0.164, porque esa configuración no cargaría.
   - Con `--home` sintético no corre ningún programa (`provider_adapter.go:78-96`), así que la oferta no aparece. Las pruebas inyectan el sondeo de versión.
5. **Documentación.** Sustituye la regla vieja y describe la capacidad, su reversión, qué guarda el journal (solo la huella y las líneas propias) y la limitación ante un binario anterior. El mínimo de espera de Codex aplica al runtime v2; en v1 no está verificado, y se dice.

Alternativa descartada: el modelo de la tarjeta de preguntas, que ejecuta un comando externo. No hay comando de Codex que guarde la clave y no deshace nada al desinstalar.

## Redacción de la guía (T1)

La misma oración, palabra por palabra, en `global.md` y en los roles; en los roles es una copia por el canal de instrucciones de desarrollador de Codex (D19-A), no un segundo hogar.

- `content/guidance/global.md:84`, después de «or launch children of their own»: «; they return only their final result, and before it they send the parent at most one message, for a blocker that needs the parent's decision, after which they return it as the final reply when no independent work remains».
- Los 19 roles con la frase de cierre: reemplazar «Return evidence, limitations, and any decision needed from the parent.» por «Return evidence, limitations, and any decision needed from the parent in one final reply; before it, send the parent at most one message, for a blocker that needs the parent's decision, and return it as the final reply when no independent work remains.» `hive-review-plan.md` no tiene esa frase y no recibe la copia; la regla le llega por `global.md`.
- `content/skills/flow-research/references/delegation.md:65`: reemplazar «A delegated child still returns its final result, not interim status.» por «A delegated child returns only its final result, as the global delegation rule sets.»
- `content/skills/flow-research/references/delegation.md:39`: precisar que una fuente que el hijo no puede leer se informa en la respuesta final, junto con lo que no pudo hacer, salvo que bloquee todo el encargo, y entonces es el aviso de bloqueo.
- `_support/docs/architecture/agent-delivery.md`: en la variante de Codex (incluida la fila de `:150`), explicar el tope de 60 s de las instrucciones base, la espera v2 que despierta con cada mensaje y la capacidad que fija el mínimo. Fuentes: versión 0.164.0 y la etiqueta `rust-v0.164.0`.
- Regla de admisión (AGENTS.md, Measurement): es una corrección observada en dos sesiones reales independientes (cam en Codex, `01a15822`, 2026-10-20, y `01a1679c`, 2026-10-23). Que los mensajes fueran avances es una inferencia por el tiempo (unos 25 s), porque están cifrados; D19-A fija dónde va la regla. El commit nombra host, IDs de sesión, fecha y la falla observada.

## Casos del editor (T2)

| Estado del archivo | Resultado |
| --- | --- |
| No existe | Crear con 0600: cabecera y dos líneas; registro `createdFile` |
| Tabla con los dos valores iguales | Aplicado; no escribe ni registra propiedad |
| Tabla sin las claves, o con otras claves | Insertar solo las claves que faltan; registrarlas |
| Tabla con un valor distinto, o un `max_wait_timeout_ms` < 600000 | Negarse con un paso manual |
| Forma con punto, tabla en línea, `[[ ]]`, tabla duplicada, `multi_agent_v2 = …` bajo `[features]`, comillas triples | Negarse con un paso manual |
| Enlace simbólico en el archivo o en la carpeta de Codex (por ejemplo `~/.codex`), sin permiso de escritura, más de 1 MiB | Negarse con un paso manual que nombra la causa. Un enlace en una carpeta de más arriba lo rechaza la escritura (`target.Safe`) al aplicar, no la oferta previa: en macOS `/tmp` y `/var` son enlaces y revisarlas impediría rutas legítimas (desviación aceptada en el build, 2026-10-25) |
| Siempre | Conservar BOM, CRLF o LF y comentarios; añadir el salto de línea final si falta; no repetir líneas del archivo en ningún mensaje |
