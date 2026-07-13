# Decisión: rechazo de RTK ("Rust Token Killer") como hook global

> **Fecha de decisión:** 2026-07-11 · **Promovido a registro durable:** 2026-07-12
> **Fuente:** investigación adversarial (research + Codex sobre código fuente v0.42.4), reporte efímero `_support/workspace/rtk-adversarial-research-2026-07-11.html`.

## Decisión

**No adoptar RTK** — ni como hook `PreToolUse` global que comprime salidas de Bash, ni en adopción parcial por repo.

## Razones

1. **Bypass del permission gate (severidad HIGH):** GHSA-7gxq-fvfc-g327.
2. **Envenenamiento de revisión vía filtros locales del repo:** CVE-2026-45792 — un repo hostil puede definir filtros que alteran lo que el agente ve.
3. **Falsos "clean" por filtrado lossy:** historial de salidas comprimidas que ocultan errores reales.
4. **Choque directo con `global/rules/quality/debugging.md`:** leer el error completo y nunca confiar en el banner de éxito son incompatibles con un compresor lossy entre el comando y el agente.
5. **Colisión de propiedad con `/deploy-global`:** RTK gestiona `~/.claude/CLAUDE.md` y `settings.json`, que este hub posee y deploya.

## Condiciones de revisita

Reconsiderar SOLO si se cumplen ambas:

- El harness primario pasa a un modelo de 200k de contexto con billing metered (la presión de tokens vuelve material), **y**
- Existen benchmarks independientes de task-success con ≥6 meses sin advisories de la clase permission/filtering.

En ese caso, empezar por `rtk <cmd>` manual pinneado en repos no-cliente — nunca el hook global.
