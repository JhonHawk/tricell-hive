# pre-push-lint-reminder

Hook `PreToolUse` (no bloqueante) que recuerda re-correr el gate local de calidad
antes de un `git push`.

## Qué resuelve

El desliz clásico del loop de dev: editas código (p.ej. un fix tras un fallo de
CI), pusheas sin re-lintear, y el CI atrapa lo que un `pnpm lint` local de 3
segundos hubiera atrapado — quemando un ciclo de CI. El recordatorio cae en el
seam correcto (**push**, no commit: N commits → un push → un recordatorio) y
nunca bloquea.

## Comportamiento

- Solo reacciona a comandos que contienen `git push`. Cualquier otro comando →
  silencio total.
- Solo aplica en repos Node: detecta `package.json` en la raíz del repo y lee los
  scripts `lint` / `test`. Sin `package.json`, o sin esos scripts → silencio (no
  ruido en repos config-only o no-Node).
- Detecta el package manager (`pnpm-lock.yaml` → pnpm, `yarn.lock` → yarn,
  `package-lock.json` → npm; fallback pnpm) y nombra el comando exacto en el
  recordatorio.
- **Consciente de lefthook:** si el repo tiene el hook de git `pre-push`
  INSTALADO por lefthook (respetando `core.hooksPath` y worktrees) **y** algún
  config de lefthook (`lefthook.yml`/`.yaml`, variantes `.` y `-local`) declara
  una sección `pre-push:`, el gate determinístico ya cubre el push → silencio.
  Config presente pero sin `lefthook install` → no hay gate activo → sigue
  recordando.
- Inyecta el aviso vía `hookSpecificOutput.additionalContext` con
  `permissionDecision: "allow"` y `exit 0` — el modelo lo ve, el push procede.

Es un recordatorio, no un gate: depende de que el agente actúe. La defensa en
profundidad la da el cuerpo de `flow-build` (re-lint obligatorio tras cualquier fix
posterior a un push).

## Registro

`/deploy-global` copia el `.sh` a `~/.claude/hooks/` **y** fusiona automáticamente
el bloque de `settings-config.json` en `~/.claude/settings.json` (merge idempotente
y additive: añade lo que falta, no duplica en re-deploys, preserva tus preferencias
y hooks propios). No requiere paso manual. Si el merge fallara, el deploy deja
`settings.json` intacto y avisa para registrar a mano desde `settings-config.json`.

## Prueba manual

```bash
echo '{"tool_name":"Bash","tool_input":{"command":"git push origin development"}}' \
  | ./pre-push-lint-reminder.sh
# Repo Node con lint/test → JSON con additionalContext.
# Comando sin "git push", o repo sin package.json → sin salida, exit 0.
```
