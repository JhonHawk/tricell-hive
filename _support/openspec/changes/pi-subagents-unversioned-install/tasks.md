# Tareas: pi-subagents sin pin

Base del build: `bba2fd48` (`development`).

## Tasks

### T1 — Fuente Hive sin versión en código y pruebas

- [x] `HiveSubagentsSource` es `npm:pi-subagents`; plan, apply, recover y goldens de estado Hive usan esa cadena; el `package.json` sintético no finge 0.74.0.

**Closes:** AC1.

**Depends on:** none.

**Locations:** `integrations/pi/package_settings.go`; `tooling/management/pi_package.go`; `integrations/pi/package_settings_test.go`; `tooling/management/pi_package_apply_test.go`; `tooling/management/pi_package_plan_test.go`; `tooling/management/testdata/state_without_overrides.golden`.

**Execution:** hilo principal (un escritor, superficie acoplada).

**Test approach:** tdd.

**Changes:** cambiar la constante; en `syntheticApplyPi` escribir versión `0.0.0-synthetic`; actualizar aserciones y goldens que esperan la fuente Hive `@0.74.0`. Dejar fixtures de pin de usuario (`@0.67.0`, conflicto con `@0.74.0` y `extensions: []`). El caso `unpinned` ya existe y debe seguir en Present.

**Verification:** `go test ./integrations/pi ./tooling/management` — PASS; plan ausente con `Source == "npm:pi-subagents"`; goldens Hive con esa cadena; recover tras apply (incluido el caso interrumpido `TestInterruptedAfterPiInstallRecoverRemovesPackage`) invoca `remove` con `Source == "npm:pi-subagents"` y esa entrada desaparece. Seguir cubriendo `TestRecoverPreservesUserPinAfterInterruptedInstall` (pin de usuario no se quita).

### T2 — Requisito de instalación

- [x] El spec actual exige `pi install npm:pi-subagents` sin pin.

**Closes:** AC3.

**Depends on:** none.

**Locations:** `_support/openspec/changes/pi-subagents-unversioned-install/specs/versioned-installation/spec.md` (delta); al integrar, `_support/openspec/specs/versioned-installation/spec.md`.

**Execution:** hilo principal.

**Test approach:** check (el delta se revisa en el PR/cierre; las pruebas de T1 cubren el comportamiento).

**Changes:** MODIFIED del requisito «Pi user-scope install declares pi-subagents» y del escenario «Pi host selected without pi-subagents» (dejar de decir «pinned»).

**Verification:** el delta nombra `npm:pi-subagents` y no `@0.74.0`; `rg '0\\.74\\.0' _support/openspec/specs/versioned-installation/spec.md` no tiene aciertos una vez integrado. Durante el build, el delta en la carpeta de cambio es la fuente.

### T3 — Documentos de producto y arquitectura

- [x] README, SECURITY y arquitectura describen `pi install npm:pi-subagents` sin versión.

**Closes:** AC2.

**Depends on:** none.

**Locations:** `README.md`; `SECURITY.md`; `_support/docs/architecture/installer.md`; `_support/docs/architecture/agent-delivery.md`; `_support/docs/architecture/tool-dependencies.md`. `llms.txt` solo enlaza tool-dependencies; no duplicar el comando salvo que lo nombre.

**Execution:** hilo principal.

**Test approach:** check.

**Changes:** sustituir `@0.74.0` / «pinned source» por la fuente sin versión. En tool-dependencies, dejar de mostrar `npm:pi-subagents@<version>` como receta Hive.

**Verification:** `rg 'pi-subagents@0\\.74\\.0' README.md SECURITY.md _support/docs/architecture` — cero aciertos; cada archivo que hoy cita el comando cita `npm:pi-subagents`.

## Verification and human review

| Gate | Timing and environment | Mechanism and expected evidence | Authority / pending condition |
| --- | --- | --- | --- |
| Pruebas Go | Tras T1, candidato local | `go test ./integrations/pi ./tooling/management` PASS | Cubierto por la implementación |
| Revisión de implementación | Tras el diff, antes o con el commit | `hive-review-code`; `hive-refute-claim` en hallazgos que bloqueen | D7-A; consentimiento de ejecución al autorizar `flow-build` |
| In-vivo | No aplica | No hay UI ni host real; las pruebas usan `pi` falso | — |

## Review status and progress

Revisión del plan: `hive-review-plan` dominio backend/integración, candidato sha1 proposal fe9238ab / design 28abbc5d / tasks ecca874b / spec delta 0b97584f. Un hallazgo bloqueante H1 (T1 no exigía `Source` en recover interrumpido); aplicado como proponía el revisor, sin otra ronda. T2 `Changes` no nombra el segundo requisito; el delta sí lo incluye — no bloquea.

Implementador: grok-4.6. Verificadores T1–T3: grok-4.7. T1 `57d2fca`; T3 `cc4290c`; T2 en el registro. `hive-review-code` (`bba2fd48..cc4290cd`): 0 bloqueos; P3 test de constante (no bloquea).
