# Diseño

## Enfoques

1. **Modo «reglas personales» de `harness-audit` con un script (elegido, D14 A; reemplaza D9 A).** La skill ayuda a definir la regla, muestra el texto y su diff, lo escribe en `~/.config/hive/rules.md` tras la aprobación del usuario, y un script lo copia a cada host donde Hive está instalado: como sección con marcadores propios en los cuatro archivos globales compartidos y como archivo propio en Cursor y Antigravity. Peldaño 2 de «Reuse before writing»: `harness-audit` ya conoce los archivos de instrucciones de cada host (`references/instruction-files.md`, `references/hierarchy.md`) y ya mantiene un archivo de toda la máquina (`HIVE.local.md`); el script reutiliza la revisión de secretos de `scripts/local_instructions.py` (`line_has_secret`). El manager ya preserva todo byte fuera de su bloque (`repository-and-distribution.md` «Installation and update contract» paso 2), así que no hay que tocarlo. ~6–8 archivos.
2. **Bloque personal desplegado por el manager (D9 A, reemplazado).** El patrón del bloque de voz en `tooling/management`, con estado v8, confirmación en terminal, validación de planes guardados y fila en `hive status`. Cuatro rondas de revisión lo dejaron en ~30 archivos. Descartado por D14 por su tamaño.
3. **Lectura al inicio de la sesión.** Una línea en `global.md` para leer un archivo de reglas al empezar. Descartado: depende de que el modelo lo lea, y el spec `local-instructions` reserva esos archivos para hechos de la máquina.
4. **Copia a mano sin skill.** Descartado: cinco o más copias que nada mantiene iguales.

## Decisiones

- **Orden y prioridad (D10 A, H2 A, D12 A).** La sección va justo después del bloque de voz, o del de Hive cuando no hay voz, y abre con la línea fijada en el delta: las reglas del usuario mandan sobre la guía de Hive, incluidas sus reglas de «Scope and authorization», salvo sus reglas de Secretos, que siguen valiendo (H2 A, que reemplaza en parte D11 B), y las instrucciones de cada proyecto mandan para ese proyecto. Coincide con la nota de `repository-and-distribution.md:92` de que Grok da prioridad al texto posterior (no verificada contra Grok). harness-audit HA-IF-08 pide que un override diga qué reemplaza y por qué; una línea general no lo hace, y la auditoría puede seguir preguntando por cada choque.
- **Quién escribe (D14 A, reemplaza la confirmación en terminal de D13 A).** Solo el modo, a pedido del usuario y tras su aprobación del texto en la conversación; la referencia del modo prohíbe escribir o sincronizar reglas personales en otro flujo o por iniciativa propia. Es una regla de guía, no un candado: un agente podría editar los archivos igual que hoy podría editar cualquier archivo global.
- **Fuente única (supuesto).** `rules.md` en `$XDG_CONFIG_HOME/hive/` o `~/.config/hive/`; el script siempre despliega desde ahí, y su modo `check` compara cada copia con ella.
- **Hosts destino (supuesto).** Los hosts con Hive instalado a nivel de usuario, detectados por el contenido administrado de Hive en cada destino (el marcador `TRICELL HIVE RULES:BEGIN` en los archivos compartidos, `hive.mdc` en Cursor, `hive-01.md` en Antigravity); Claude Code resuelve `CLAUDE_CONFIG_DIR` como el manager (`tooling/management/types.go:444`) y Grok siempre `~/.claude/CLAUDE.md` (`integrations/grok/grok.go:37`). `check` también reporta una copia personal en un host que ya no tiene Hive. Alternativa descartada: todos los hosts aunque no tengan Hive.
- **Cursor fuera de la carpeta personal.** El puntero de respaldo de Cursor para workspaces fuera de la carpeta personal nombra solo `hive.mdc` (`README.md:270`, `instruction-files.md:41`); esos workspaces no cargan `personal-rules.mdc`. Se documenta como límite.
- **Nombres de archivo (supuesto).** `personal-rules.mdc` y `personal-rules.md`, sin el prefijo `hive-`, porque son del usuario y no del manager.
- **Límite de 8,000 bytes y sin secretos (supuesto)**, con `line_has_secret` de `local_instructions.py` línea por línea (salta URLs y rutas, reconoce prefijos de tokens y aplica la prueba de entropía solo a tokens de 20 o más caracteres); el `secret_like` desnudo rechazaría identificadores de prosa como `autoMemoryDirectory`.
- **Alcance de escritura de `harness-audit` (por la revisión del plan).** `SKILL.md:12` prohíbe editar configuración fuera del proyecto y `apply.md:5` edita solo el proyecto actual; el modo nombra su excepción en `SKILL.md:12`, como ya hace con el archivo de la máquina `HIVE.local.md`, con sus destinos exactos, y su referencia dice que el pedido del usuario al modo es la instrucción explícita que `global.md` exige para un cambio de configuración global. El modo no corre bajo delegación desatendida (`SKILL.md:79`).
- **HA-HI-08 y HA-HI-11 no cambian (supuesto).** Hoy mandan una preferencia del usuario a `propose-global` o al archivo de instrucciones del nivel que la necesita; podrían mandar una preferencia de toda la máquina al modo nuevo, pero cambiar esas reglas de auditoría de memoria agranda el cambio que el usuario pidió más sencillo. Alternativa descartada: enrutarlas al modo en este cambio. Para revertirlo, se planea aparte.
- **Concurrencia con el manager.** El script no toma el lock de Hive; antes de reemplazar un archivo lo vuelve a leer y, si cambió desde su primera lectura (por ejemplo, un refresco del mantenedor), no lo escribe y lo nombra, para no devolver una copia vieja del bloque de Hive que el manager vería como editada a mano.
- **Posición.** El contenido de la sección sobrevive a `hive update`; si el bloque de Hive se borró a mano y el manager lo reinstala al final del archivo, la sección queda antes de él: `check` lo reporta y `deploy` la mueve.
- **Al desinstalar Hive (supuesto)**, las secciones y archivos personales se quedan: son del usuario; el modo `remove` del script los quita.

## Forma propuesta

- `content/skills/harness-audit/references/personal-rules.md`: el modo (entrevista, texto y diff, aprobación, `rules.md`, deploy, check, remove; quién puede escribir).
- `content/skills/harness-audit/scripts/personal_rules.py` con `deploy`, `check` y `remove`; marcadores `<!-- === PERSONAL RULES:BEGIN === -->` y `END`; inserción tras el final del bloque de voz o de Hive cuando la sección no existe, reemplazo en el lugar cuando existe; escritura atómica que conserva permisos y finales de línea.
- `content/skills/harness-audit/SKILL.md`: la línea de entrada del modo y su mención en `description`.
- `tests/skills/personal_rules_test.py`.

## Impacto

- Archivos globales de los siete hosts registrados en esta Mac (`hive status --scope user`, 2026-11-21).
- No toca `tooling/`, `integrations/` ni `content/guidance/global.md`.

## Recuperación

`personal_rules.py remove` quita las secciones y los dos archivos. Borrar `rules.md` no quita nada por sí solo.
