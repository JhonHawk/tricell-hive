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
- **Detecta dos clases, ambas con antigüedad ≥ 2 h (`AGE_MIN=120`):**
  1. Procesos `agent-browser` / `Chrome for Testing` — reporta conteo y edad del
     más viejo, sugiere revisar `agent-browser session list` con el usuario y
     cerrar solo sesiones confirmadas stale. El mensaje prohíbe explícitamente
     `close --all` como reflejo: mataría sesiones de otros proyectos activos.
  2. Listeners TCP de runtimes dev (`node`/`bun`/`deno`/`next`/`pnpm`/`npm`/`yarn`/`vite`)
     en puertos típicos de desarrollo: 3000-3999, 4200-4299, 4321, 5173-5179,
     8080-8089 — reporta puerto, PID, edad y cwd del proceso.
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
- **Solo-reporte, nunca mata nada.** Un proceso listado puede pertenecer a otra
  sesión activa en paralelo; el mensaje instruye ofrecer la limpieza al usuario y
  no ejecutarla sin su confirmación.
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
confirme el trust prompt del comando (`[hooks.state]` en `~/.codex/config.toml`).

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
