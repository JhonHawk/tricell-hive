# Decisión: no adoptar optional reference project (ecosystem configurator); portar 2 ideas sueltas

> **Fecha de decisión:** 2026-07-11 · **Promovido a registro durable:** 2026-07-12
> **Fuente:** investigación adversarial (research + Codex), reporte efímero `_support/workspace/optional reference project-adversarial-research-2026-07-11.html`.

## Decisión

**No adoptar el binario optional reference project** (configurador Go que inyecta memoria/workflow/skills/permisos en ~15 harnesses). Sí portar dos ideas puntuales del análisis.

## Razones del rechazo

1. **Solapamiento masivo:** 6/10 de sus componentes ya están cubiertos por este hub (deploy-global, flow pack, Engram, reglas path-scoped).
2. **Engram es literalmente el mismo proyecto** (Gentleman-Programming): adoptar optional reference project duplicaría la gestión de la misma DB SQLite desde dos gestores.
3. **Colisión de workflows spec-driven:** su workflow y el flow pack escribirían instrucciones concurrentes en el mismo `CLAUDE.md`.
4. **Riesgo de doble gestión** de archivos que `/deploy-global` posee.

## Ideas portadas / pendientes

| Idea | Estado |
|---|---|
| A — Snapshot-before-write en `/deploy-global` (backup `tar.gz`, poda a 5) | ✅ Promovida — ya implementada en `.claude/skills/deploy-global/SKILL.md` |
| B — Convención `model:`/`effort:` por fase en agentes/skills del flow pack (exploración = modelo barato, diseño = modelo caro; el frontmatter ya lo soporta) | ⏳ Pendiente — falta declarar la convención; sin registro previo a este documento |
