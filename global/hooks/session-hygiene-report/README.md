# session-hygiene-report

Hook `SessionStart` (solo-reporte) que detecta procesos probablemente fugados por
sesiones de agente anteriores y los presenta al inicio de la sesión nueva.

## Qué resuelve

El patrón documentado el 2026-07-23 en `family-jk`: 13 sesiones headless de
`agent-browser` huérfanas (17–27 h), un `next dev` de otro workspace ocupando el
puerto 3000 por 12 h. Las reglas "never orphan" (dev servers en `global/CLAUDE.md`,
sesiones de browser en `rules/tools/browser-automation.md`) son prompt-convention y
tienen un caso límite legítimo — "déjalo corriendo mientras reviso" termina el turno
con el server vivo por diseño, y nadie regresa a apagarlo. Este hook es el backstop:
la siguiente sesión fresca recoge lo que quedó.

## Comportamiento

- Corre solo en sesiones frescas (`startup`/`clear`); silencio en `resume`/`compact`.
  Sin input parseable (runner de Codex) → reporta por default.
- En el adaptador de Pi, `startup` no se interpreta por sí solo como una sesión fresca:
  la frescura se decide con las entradas de contexto activas. Un `resume` puede emitir el
  reporte fresco si esas entradas ya no existen; el adaptador lo encola como un
  `custom_message` nativo oculto (`triggerTurn: false`, sin `deliverAs`) y lo recarga una
  sola vez antes del trabajo del modelo. Pi 0.85.1 crea el archivo de sesión hasta después
  de la primera respuesta del asistente, así que no se promete durabilidad en disco antes
  de esa respuesta. La advisory pertenece al padre; los hijos no reciben una inyección nueva.
- **Detecta dos clases, ambas con antigüedad ≥ 2 h (`AGE_MIN=120`):**
  1. Procesos `agent-browser` / `Chrome for Testing` — reporta conteo y edad del
     más viejo, sugiere revisar `agent-browser session list` con el usuario y
     cerrar solo sesiones confirmadas stale. El mensaje prohíbe explícitamente
     `close --all` como reflejo: mataría sesiones de otros proyectos activos.
  2. Listeners TCP de runtimes dev (`node`/`bun`/`deno`/`next`/`pnpm`/`npm`/`yarn`/`vite`)
     en puertos típicos de desarrollo: 3000-3999, 4200-4299, 4321, 5173-5179,
     8080-8089 — reporta puerto, PID, edad y cwd del proceso.
- **Además (sin umbral de edad): hive-profile stale.** Si el repo raíz del `cwd`
  de la sesión tiene un `AGENTS.md` con bloque `hive-profile:start` (generado por
  `harness/hive-compile.py`), extrae el SHA del stamp (`hive@<sha>`) y lo compara
  contra el HEAD del checkout del hive (`HIVE_REPO`, default
  `~/Development/projects/tricell/tricell-hive`): stale = el hive avanzó tocando
  `global/rules/` o el clasificador (`git log --name-only <stamp>..HEAD`), o el
  stamp es desconocido para el checkout. Emite UNA línea advisory sugiriendo
  regenerar (ofrecer, nunca regenerar sin pedirlo); silencio si no hay checkout
  del hive en la máquina o el repo no lleva perfil. Clave de cooldown:
  `hive-profile:<sha>` — un stamp nuevo rompe el silencio del TTL.
- **Barrido máquina-completa, consciente de multi-proyecto.** El barrido cubre
  toda la máquina a propósito — la limpieza cruzada es el valor (el caso Umbrella).
  Pero el trabajo diario corre varios workspaces en paralelo, así que cada dev
  server se clasifica comparando su cwd con el `cwd` de la sesión que arranca:
  **this-workspace** → ofrecer limpieza directa; **other-workspace** → ofrecer
  limpieza pidiendo antes confirmar que ese proyecto está inactivo (puede ser el
  server activo de una sesión concurrente); sin `cwd` en el input (p.ej. Codex) →
  verificar con el usuario antes de cualquier limpieza.
- **Cooldown de 6 h con fingerprint (`TTL_MIN=360`).** Reporta a lo más una vez
  por TTL en toda la máquina — el state file
  (`~/.cache/session-hygiene-report.state`, override con `SESSION_HYGIENE_STATE`)
  lo comparten los deployments de Claude Code y Codex. Guarda identidades
  estables de los hallazgos (`browser:<pid>`, `dev:<pid>:<port>`, nunca edades):
  un hallazgo NUEVO respecto al último reporte rompe el silencio aunque el TTL
  no haya vencido; un set igual o menor calla hasta que venza. Evita que cada
  pane/sesión nueva re-reporte los mismos huérfanos.
- **Tercera clase (solo Claude Code, gated por `CLAUDECODE=1`): shell drift.**
  Verifica que el tool Bash siga en bash 5 (`CLAUDE_CODE_SHELL` en settings
  `env`). Checks deterministas: variable ausente en `settings.json`, binario
  faltante/no ejecutable, bash <5, y —para el caso "daemon anterior al
  setting"— la ausencia de `CLAUDE_CODE_SHELL` en el **entorno del propio
  hook**: los hooks heredan el env de settings (recarga en vivo según docs),
  así que settings-en-disco sin env-en-runtime es anómalo (versión vieja o
  recarga fallida) y se reporta como advertencia suave que instruye verificar
  in-band (`echo $BASH_VERSION` en el primer Bash call).
  **Nunca inferir drift de `~/.claude/shell-snapshots/`**: el snapshot se crea
  lazy en el PRIMER Bash call —después de los hooks de SessionStart— y se borra
  al salir limpio pero sobrevive crashes; ausencia y presencia son ruido
  (falso positivo estructural — corregido 2026-08-12). Exento del cooldown: un
  shell drifteado afecta cada comando. En Grok/Codex la sección completa se
  salta (`CLAUDECODE` ≠ 1; validado en runtime Grok).
- **Solo-reporte, nunca mata nada.** Un proceso listado puede pertenecer a otra
  sesión activa en paralelo; el mensaje instruye ofrecer la limpieza al usuario y
  no ejecutarla sin su confirmación.
- Si falta el script, el payload es inválido o el reporte advisory expira por timeout, Pi
  emite una advertencia y continúa. La ausencia del reporte nunca bloquea al padre ni cambia
  las guardas blocking existentes.
- **Encabezado adaptativo:** el intro y las instrucciones finales se componen
  según las clases realmente presentes — un reporte solo-drift ya no afirma
  "processes likely leaked".
- Silencio total cuando nada califica — cero ruido en sesiones limpias.
- Emite vía `hookSpecificOutput.additionalContext` (`SessionStart`), exit 0 siempre.

## Fuera de alcance (deliberado)

- **Contenedores Docker** — `docker ps` puede colgarse con el daemon caído, y las
  DBs de proyecto son longevas por diseño (falso positivo casi seguro).
- **Listeners no-dev** — servicios de sistema, daemons MCP, editores: fuera de los
  rangos de puerto listados.

## Registro

`/deploy-global` copia el `.sh` a `~/.claude/hooks/` y fusiona
`settings-config.json` en `~/.claude/settings.json` (paso 13). La variante Codex se
despliega por el paso 13d (genérico sobre `global/hooks/*/codex-hooks.json`): copia a
`~/.codex/hooks/` y merge en `~/.codex/hooks.json`; queda inerte hasta que Codex
confirme el trust prompt del comando (`[hooks.state]` en `~/.codex/config.toml`). En Pi,
el runtime padre adapta el mismo payload y ejecuta el script advisory bajo el runner
acotado de 10 segundos; la copia del script forma parte del despliegue del root de Pi.

## Prueba manual

```bash
export SESSION_HYGIENE_STATE=/tmp/shr-test.state; rm -f "$SESSION_HYGIENE_STATE"
echo '{"source":"startup"}' | ./session-hygiene-report.sh
# Con huérfanos ≥2h → JSON con additionalContext; limpio → sin salida, exit 0.
echo '{"source":"startup"}' | ./session-hygiene-report.sh  # repetición en TTL → silencio
echo '{"source":"resume"}' | ./session-hygiene-report.sh   # siempre silencio
# TTL vencido → reporta de nuevo:
{ echo 0; tail -n +2 "$SESSION_HYGIENE_STATE"; } > t && mv t "$SESSION_HYGIENE_STATE"
# Hallazgo nuevo dentro del TTL → reporta (quita una key del state y rerun):
grep -v 'browser:' "$SESSION_HYGIENE_STATE" > t && mv t "$SESSION_HYGIENE_STATE"
```
