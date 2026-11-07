# Diseño

## Enfoques

Para entregar la guía global (37.738 bytes) a un host que corta cada archivo de reglas a 24.000 bytes se presentaron tres caminos:

- **Reglas siempre activas partidas (elegido, D2 A).** `~/.gemini/config/rules/hive-NN.md` con `trigger: always_on`, cortadas por encabezado `##`. La guía llega desde el primer turno sin depender de hooks; sigue el patrón del adaptador de Cursor (`~/.cursor/rules/hive.mdc`, bloque gestionado con frontmatter) y el de una referencia externa en producción. Costo: el gestor aprende a escribir una fuente en varios archivos.
- **Hook `SessionStart` con `injectSteps`.** Una sola fuente, pero sin tope verificado, alcance a subagentes desconocido, y una sesión sin guía si el hook falla.
- **Plugin de `agy`.** Mecanismo nativo, pero sus reglas tienen el mismo tope, `agy` es dueño de la instalación y el espacio de nombres puede renombrar las skills.

## Decisiones

- **Roles de solo lectura (D3 B).** El rol no lleva `tools`; la restricción es la instrucción del rol, como hoy en los hosts sin barrera técnica. `validateAccess` acepta para `antigravity` un `access.observe` vacío.
- **Modelo de los roles.** `model: "inherit"` en `integrations/agent-profiles.json`; Antigravity no admite esfuerzo por rol, así que `acceptsProfileEffort` y `acceptsRoleEffort` lo rechazan, como en Cursor. `hive models set` ofrece para Antigravity la lista fija `inherit`, `flash_lite`, `flash` y `pro` (doc de la skill integrada `run-workflow` de `agy` y una referencia externa), porque `agy` no tiene un comando que liste los niveles de agente.
- **Sonda de formato antes del código (T0).** La doc local de `agy` no describe `~/.gemini/config/rules/` ni `~/.gemini/config/agents/`; los describen la doc oficial y el changelog del binario. Una sonda con una regla y un agente de prueba confirma ambos antes de T2 y T3, dentro de las llamadas autorizadas.
- **Host opcional en los perfiles.** Como Cursor, `antigravity` es opcional en `ReadProfiles`, para que un release congelado sin la entrada siga instalándose.
- **Voz solo en la primera parte** (decisión menor): repetir el bloque de voz en cada parte suma bytes contra el tope sin aportar nada.

## Forma propuesta

### Partes de la guía

- Fuente: la guía reescrita que hoy produce `nextRecord` para un bloque (`global.md` con los enlaces de skills ya reescritos por `skillsDirFor`).
- Corte determinista en el plan: se acumulan secciones `## ` mientras la parte, con frontmatter, marcadores de bloque y (en la primera) la voz más larga posible, quede en 22.000 bytes o menos; margen de 2.000 bytes bajo el tope de 24.000 para ediciones de la guía. Solo cuentan los encabezados fuera de bloques de código: `global.md` tiene un `## Hive` dentro de un bloque ```` ```markdown ```` (líneas 108-118). Una sección que sola pase el límite hace fallar el plan con un error que la nombra (`global.md` no tiene encabezados `###`, así que no hay corte de segundo nivel). Medido el 2026-11-06 sobre `93ab65ee`: dos partes, de unos 16.600 y 21.100 bytes de guía.
- Destino: `<AntigravityHome>/rules/hive-01.md`, `hive-02.md`…; cada archivo lo crea Hive con `---\ntrigger: always_on\n---\n` y un bloque gestionado. Las partes comparten `GlobalSource` y llevan un índice de parte en el registro de estado.
- Un cambio en la cantidad de partes agrega los archivos nuevos y quita los que sobran, solo si son de Hive (registro y marcadores intactos).

### Skills, roles y hooks

- Skills: un enlace por skill `<AntigravityHome>/skills/<skill> -> ~/.agents/skills/<skill>`, mismo mecanismo que `claude.go:37-41`; el consumidor queda registrado en el almacén compartido. Un enlace ajeno con el mismo nombre detiene la instalación con el error de colisión existente.
- Roles: `<AntigravityHome>/agents/<rol>.md`, renderizados por `integrations/agents` con `name`, `description` y `model`.
- Hooks: un hook con nombre `hive` en la raíz de `<AntigravityHome>/hooks.json`. El editor reemplaza solo la clave `hive`, conserva byte a byte las demás claves (hoy `herdr`) y guarda la copia de respaldo que exige la registración de hooks.
  - Eventos: `SessionStart` → inicio de sesión (y entrada en el mapa de sesiones); `PreInvocation` → mensaje del usuario o actividad, que también refresca el mapa; `Stop` → fin de turno, nunca fin de sesión, porque `agy` lo emite en cada turno (`terminationReason: NO_TOOL_CALL`) y no tiene evento de fin de sesión. No se registra `PreToolUse`: su salida exige `decision` y un error bloquearía herramientas.
  - La carga no nombra su evento: el comando lo recibe por argumento (`hive hook antigravity <evento>`). El directorio de trabajo sale de `workspacePaths[0]`, no de `os.Getwd()`, porque `agy` corre los hooks en la carpeta de `hooks.json`.
  - Salida: en cada ejecución, incluidas las de error, argumentos desconocidos, pánico recuperado o tiempo agotado, `hive hook` imprime la salida mínima aceptada para el evento: `{}` en `SessionStart` y `PreInvocation`, y en `Stop` un `decision` distinto de `continue`. Es la única excepción a la regla «los hooks no imprimen nada», recogida en el delta de `walkthrough-viewer`.

## Impacto

- Listas de hosts: `types.go` (`validateHosts`), `update_contract.go`, `plan.go` (`skillsDirFor`, `resolve`, raíces de home), `integrations/agents/agents.go` (`ReadProfiles`, `knownHosts`, `validateAccess`), `cli/install.go` (`installerHosts`, `hostBinaries`: `agy`, `installRoots`), `cli/model_catalog.go`, `doctor_sessions.go`, `cli/hook_command.go`, `cli/hook_map.go`, `cli/view_identity_resolve.go` (`agy` → `antigravity`), `viewer/timeline_contract.go` y su copia TypeScript, `management/host_hooks.go`.
- Un plan guardado antes del cambio pasa a ser inválido por la nueva raíz de home (`plan.go:888`); se regenera, como pasó con Cursor.
- No toca producción: el despliegue real lo hace el refresco local del mantenedor tras integrar.

## Recuperación

`hive remove --hosts antigravity` quita las partes, los enlaces, los roles y la clave `hive` de `hooks.json`; el resto de `~/.gemini/` queda igual. Revertir el commit deja a Antigravity como host desconocido; un estado ya instalado se quita antes con el binario nuevo.
