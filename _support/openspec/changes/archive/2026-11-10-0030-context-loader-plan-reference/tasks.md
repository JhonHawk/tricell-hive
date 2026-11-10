# Tareas

### T1 — Igualar la documentación y la spec al código

- [x] `agent-delivery.md` y `deployment-manager.md` nombran `flow-plan/SKILL.md`
- [x] El requisito de `context-delivery` nombra `flow-plan/SKILL.md`, con su delta `MODIFIED`

**Closes:** AC1

**Verificación:** `go test ./tests/content/ -run TestSpecDeltasApplied` y la búsqueda de `plan-review.md` fuera del archivo histórico y de los fixtures de piloto.
