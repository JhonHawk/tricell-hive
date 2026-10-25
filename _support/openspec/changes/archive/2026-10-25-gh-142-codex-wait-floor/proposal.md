# Codex: mínimo de espera de 5 min y hijos que solo devuelven su resultado final (2026-10-25)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` el 2026-10-25 por el [PR #154](https://github.com/JhonHawk/tricell-hive/pull/154), merge `0bd7fc54` · #142 cerrado · binario recompilado y `hive update` aplicado · registro archivado el mismo día |
| Tracker · GitHub Issues | • [#142 — Codex: el hilo principal espera a sus hijos en tramos de 60 s y los hijos envían avances intermedios](https://github.com/JhonHawk/tricell-hive/issues/142) |
| Git | `automatic` (D7-A, vigente en esta sesión) · rama desde `development` · PR, `/code-review` (D8-A) y merge con la suite local en verde (no hay CI) |
| Verificación | `hive-verify-task` por tarea · `go vet ./...`, `go test ./...` y `unittest` sobre el candidato final · `/code-review` |
| Siguiente paso | Ninguno; registro cerrado. Queda por observar, en una sesión real de Codex, la espera de 5 min o más y que los hijos dejen de mandar avances |

## Solicitud

Pregunta de cierre de #151, 2026-10-25: D14-A, «Seguir con #142 ahora».

Respuestas en la investigación y el plan, 2026-10-25, en tarjetas nativas:
- D15-A: «Guía + ajuste de Codex».
- D16, en texto libre: «Ajustemos la regla que si edita la config de codex».
- D17-A: «Sí, con respaldo». Se aplicó el mismo día; ver Decisiones.
- D18-A: «Instalar y actualizar».
- D19-A: «Global + cierre de roles».

Criterios de aceptación de [#142](https://github.com/JhonHawk/tricell-hive/issues/142), leídos el 2026-10-25:

> - En una sesión real de Codex con hijos de larga duración, los `wait_agent` usan tramos de 5 minutos o más, y no hay ciclos de 60 s o menos.
> - En esa misma sesión, los hijos no envían avances intermedios al hilo principal, salvo para reportar un bloqueo.
> - El cambio de guía se apoya en documentación vigente de Codex sobre `wait_agent` y la mensajería entre agentes, con su versión.

Los dos primeros solo se pueden observar en una sesión real posterior al merge, porque los pilotos están pausados. Quedan como observación pendiente y no como criterio de este cambio (ver Alcance).

## Decisiones

- D15-A: corregir por los dos lados, la guía para los hijos y el ajuste del host.
- D16: Hive puede escribir la configuración de Codex. Antes no lo hacía, por la regla fijada en `tooling/providers/types.go:19` («Hive never edits Codex's config»), que este cambio sustituye. Solo escribe las claves que declara, con respaldo y relectura. Al desinstalar quita solo las claves que escribió, y solo si conservan su valor.
- D17-A: aplicado el 2026-10-25 en `~/.codex/config.toml` de la máquina del mantenedor, con la tabla `[features.multi_agent_v2]`, `min_wait_timeout_ms = 300000` y `default_wait_timeout_ms = 600000`. El respaldo es `~/.codex/config.toml.hive-142.bak` (0600), y `codex features list` sale con 0. Se revierte borrando la tabla o restaurando el respaldo. Se dejó fuera `subagent_developer_instructions`: el código fuente de Codex dice que «Overrides inherited developer instructions for subagents without role-specific instructions», así que no llega a los roles de Hive y reemplazaría las instrucciones de los hijos sin rol.
- D20-A (2026-10-25, tras la revisión del plan): el respaldo de D16 es una copia `config.toml.hive-backup` (0600) junto al archivo, en la carpeta de Codex, hecha antes de la primera escritura. `plan remove` la borra después de revertir bien, el doctor dice si existe y el journal de Hive no guarda esos bytes.
- D22-A (build, 2026-10-25): la segunda ronda fallida de T5 la corrigió el hilo principal y la revisó el mismo verificador. El usuario señaló que esa pregunta sobraba.
- D18-A: el ajuste se ofrece como capacidad opcional al instalar y en `hive update` mientras no esté aplicado. Va marcado por defecto en modo interactivo y no se ofrece sin terminal.
- D19-A: la regla de que los hijos devuelven solo su resultado final vive en `global.md:84` y se añade a la frase de cierre que comparten los roles.
- Decisiones del agente:
  - Valores: 300 000 y 600 000 ms, el tramo de 5 a 10 min de `delegation.md`.
  - Modelo de implementación: como el paquete de Pi (`Plan.PiPackage`, `State.PiSubagentsSource`, journal y deshacer), no como la tarjeta de preguntas, porque esta no puede deshacer y D16 pide revertir.
  - El ajuste se ofrece solo con Codex 0.164 o posterior: la tabla tiene `deny_unknown_fields` y un Codex anterior podría no cargar la configuración.
  - Si `config.toml` es un enlace simbólico, si un valor del usuario difiere o si aparece una forma de TOML ambigua, Hive se niega y da un paso manual. Nunca sobrescribe un valor del usuario.
  - Sin biblioteca de TOML: `deps_test.go` mantiene `management` sin dependencias externas.
- Delta de specs: no hace falta. Ninguna capability de `_support/openspec/specs/` cubre el instalador de capacidades opcionales, y el contrato queda en `installer.md` y `tool-dependencies.md`.

## Objetivo

En Codex, el hilo principal espera a sus hijos en tramos de 60 s y los hijos le mandan avances. Cada avance lo despierta y le hace volver a procesar entre 200 000 y 450 000 tokens. La causa son las instrucciones base de Codex: «Avoid performing blocking sleep or wait calls longer than 60 seconds», más un aviso al usuario cada 60 s como mucho. El modelo las prefirió a la regla de Hive, que había leído antes de su primera espera en los dos hilos de cam (10-20 y 10-23).

El cambio corrige los dos lados:
- **El host:** Codex sube por su cuenta cualquier espera menor al mínimo configurado, y Hive ofrece escribir ese mínimo.
- **La guía:** la regla de los hijos llega como instrucción de desarrollador por medio de sus roles.

La espera sigue terminando antes si un hijo acaba o el usuario escribe, así que el usuario no queda incomunicado.

## Alcance y aceptación

Incluye la guía (`global.md`, `delegation.md`, 19 roles), `integrations/codex`, `tooling/providers`, `tooling/management`, `tooling/cli` (instalar, actualizar, desinstalar, doctor, TUI) y la documentación (README, llms.txt, `installer.md`, `tool-dependencies.md`, `deployment-manager.md`, `agent-delivery.md`, `tooling/providers/README.md`, CHANGELOG).

Exclusiones:
- Otros hosts.
- `subagent_developer_instructions`, por el motivo de D17-A.
- Observar en una sesión real los dos primeros criterios del ticket: queda como observación tras el merge, y #142 se cierra al integrarse, como fija la regla de estado de tickets.
- Cortar una release: se propone al cierre.

Restricciones:
- `go vet ./...`, `go test ./...` y `unittest` en verde.
- Cada `SKILL.md` en 16 000 bytes o menos.
- `management` sigue sin dependencias externas.
- Ningún mensaje repite líneas de `config.toml`, que puede guardar secretos, y el journal no guarda sus bytes completos.
- El PR declara el tamaño de `global.md`.

- AC1. Los hijos devuelven solo su resultado final y avisan al padre antes solo por un bloqueo. Lo dicen `global.md:84` y la frase de cierre de los 19 roles que la comparten, y `delegation.md` remite a la guía global. *Falso en la base cuando* ningún rol ni `global.md` lo dicen y la frase solo está en la espera del hilo principal (`delegation.md:65`).
- AC2. Un lector y editor de `config.toml` clasifica cada estado y actúa en consecuencia:
  - Sin archivo: lo crea con 0600.
  - Con los dos valores iguales: ya aplicado, no escribe.
  - Con la tabla sin las claves: inserta solo las que faltan.
  - Con un valor distinto, una forma con punto, una tabla en línea, `[[ ]]`, una tabla duplicada, `multi_agent_v2` como valor, comillas triples o un enlace simbólico: se niega con un paso manual.

  En todos los casos conserva BOM, CRLF y comentarios. Pruebas en tabla. *Falso en la base cuando* `integrations/codex` no tiene ese editor.
- AC3. Con Codex 0.164 o posterior y el ajuste sin aplicar, `hive install` lo ofrece marcado por defecto en modo interactivo y no lo ofrece sin terminal. La elección entra en el plan por `Options.CodexWaitFloor`, y el plan se reconstruye después de elegir. Al aplicarlo, escribe mediante una entrada de journal por líneas que guarda solo la huella previa y las líneas propias, relee el archivo y registra en el estado lo que escribió. Antes de la primera escritura deja `config.toml.hive-backup` con 0600, sin pisar uno que ya exista. *Falso en la base cuando* el catálogo no tiene la capacidad.
- AC4. `hive update` la ofrece en modo interactivo mientras no esté aplicada y no la ofrece si ya lo está. Cuando la elección escribe la configuración, pide confirmación aunque el contenido de Hive no haya cambiado. *Falso en la base cuando* `update.go:65` no muestra capacidades opcionales.
- AC5. `hive plan remove`, y la desinstalación de la TUI, cuando sus hosts incluyen `codex`, quitan solo las líneas que Hive escribió y que conservan su valor. Quitan la cabecera solo si Hive la creó y quedó vacía, y borran el archivo solo si Hive lo creó y quedó vacío. Si el usuario cambió un valor, lo dejan y lo informan. `Recover` usa la misma inversa por líneas: si Codex editó el archivo entre la escritura y la recuperación, no deja la transacción bloqueada. *Falso en la base cuando* no existe ese registro de propiedad.
- AC6. Con Codex anterior a 0.164 o sin `codex`, no se ofrece, ni siquiera como paso manual, y no se escribe nada. *Falso en la base cuando* la capacidad no existe.
- AC7. Con Codex registrado, `hive doctor` muestra una fila con el estado del ajuste leyendo `config.toml`, y avisa si la tabla existe con Codex anterior a 0.164. Sin Codex registrado, la fila no aparece. *Falso en la base cuando* ninguna fila del doctor lee configuración (`doctor_integrations.go:169-179`).
- AC8. README, llms.txt, `installer.md`, `tool-dependencies.md`, `deployment-manager.md`, `tooling/providers/README.md`, `agent-delivery.md:150`, los comentarios de `types.go` y `catalog.go`, `agent-delivery.md` (la variante de Codex) y CHANGELOG describen la capacidad. Ya no dicen «Hive edits no host configuration file itself». *Falso en la base cuando* `README.md:169` lo dice.

**Not yet specified:** nada.
