# Tareas

### T1 — Caracterización, lectores únicos y reuso

- [x] Un test de caracterización en `tooling/management/versioning_test.go` fija la salida de `productFromSource` antes de tocar el código, en los siete casos de A3:
  - con `release.json`: `version` igual a `ProductVersion` y `artifact_id` igual al SHA-256 de los bytes del archivo;
  - con `release.json` y un `VERSION` ilegible (un directorio): el mismo resultado, porque `VERSION` no se lee;
  - con un `release.json` legado sin `product_version`: `nil`;
  - con `VERSION`: el `artifact_id` de la fórmula actual;
  - sin ninguno de los dos: `nil`;
  - con `VERSION` vacío: `nil`;
  - con `VERSION` inválido y sin manifiesto: error.

  **Valores esperados:** se calculan en el propio test con `crypto/sha256` y `json.MarshalIndent` (o la codificación de `encode` reescrita en el test), nunca con `hash`/`encode` de producción. Si no, alterar esas funciones pasaría inadvertido.

  **Detalle:** llamar a `productFromSource` directamente exige un `Release` con un `ID` hexadecimal de 64 caracteres, porque si no `validateProduct` lo rechaza.

  Debe pasar sobre el código actual.
- [x] `distribution.ReadManifestData` y `version.ReadSourceFile` según [design.md](design.md#contratos), cada uno con su test unitario:
  - `ReadManifestData` devuelve los mismos bytes que el archivo y un error `IsNotExist` si falta;
  - `ReadSourceFile` cubre los casos ausente, presente con espacios y vacío.
- [x] `ReadManifest`, `ReadSource`, `productFromSource` y `ValidateBootstrapPackage` usan los lectores nuevos. `rg -n "os.ReadFile" tooling/management/versioning.go` no devuelve nada, y `ValidateBootstrapPackage` ya no llama a `os.ReadFile`.

**Evidencia (T1):**
- **Delegación:** `backend-developer` nativo, sin commits, en las 7 rutas del brief.
- **Caracterización:** los 7 casos pasan sobre el código sin cambiar y después del cambio, con las mismas identidades.
- **RED:** `undefined: ReadManifestData` y `undefined: ReadSourceFile`, antes de implementar.
- **Reversión:** alterar la fórmula en la rama del manifiesto hace fallar sus 2 casos, y en la rama de `VERSION`, su caso. Ambas se restauraron.
- **Suite:** `go vet ./...` pasa, y `go test -race -count=1` pasa en `./tooling/...` (corrida del subagente) y en `./tests/...` (hilo principal), sobre el mismo candidato.
- **Decisión del implementador:** en la rama de `VERSION`, el error se revisa antes que `present`, para que un error de lectura no se tome por "ausente".
- **Hueco cerrado después:** en `/code-review` se agregó el caso de error de lectura de `VERSION`.

**Depende de:** nada.
**Ubicaciones:** `tooling/distribution/manifest.go`, `tooling/distribution/bootstrap.go`, `tooling/version/version.go`, `tooling/management/versioning.go` y sus `_test.go` (`tooling/distribution/manifest_test.go`, `tooling/version/version_test.go`, `tooling/management/versioning_test.go`).
**Ejecución:** delegado a `backend-developer`. Son tres paquetes con sus tests y la interfaz está fijada en el diseño. No se superpone con otro trabajo. Sin commits; el hilo principal revisa el diff.
**Enfoque de prueba:**
- `characterization` para `productFromSource` y `ValidateBootstrapPackage`: el comportamiento no cambia.
- `tdd` para los dos lectores nuevos: sus tests fallan primero por la función que falta.
**Verificación:**
- `go test -race -count=1 ./tooling/management/... ./tooling/distribution/... ./tooling/version/... ./tooling/package/...` pasa.
- El test de caracterización pasa igual antes y después del cambio.
- Si se altera a mano la fórmula del `artifact` en `versioning.go` (en cada rama), el test de caracterización falla. Después se restaura.

### T2 — Revisión, entrega y tracker

- [x] `go vet ./...` y `go test -race -count=1 ./...` pasan (ver la evidencia de T1).
- [x] `/code-review` sobre el diff antes del commit (E2-A). No marcó ningún hallazgo como bloqueante.
  - **Refutados (2):**
    - La carrera entre `Lstat` y `ReadFile` en `ReadManifestData`.
    - Las tres lecturas del manifiesto en `ValidateBootstrapPackage`.

    Los dos son anteriores al cambio, y el segundo está fuera de alcance por B1. Explotarlos exige poder escribir en el directorio del paquete mientras corre el proceso, y con ese acceso basta editar `release.json` directamente; cerrarlos no protege de nada.
  - **Aplicados (7):**
    - El caso "`VERSION` no se lee" ahora usa un `VERSION` ilegible, que es un directorio. Si se inyecta una lectura en la rama del manifiesto, fallan sus 2 casos.
    - El error "invalid product identity" queda fijado.
    - Test de `ReadSourceFile` con un error de lectura, lo que cierra el hueco declarado en T1.
    - Los comentarios de los tests ya no apuntan a `tasks.md`.
    - El comentario de `ReadManifestData` ya no nombra a sus llamadores.
    - Se quitó la rama redundante de `IsNotExist`.
    - Los conteos de casos del registro quedan en siete.
  - **Verificación tras las correcciones:** `go vet` y `go test -race` pasan en `management`, `version`, `distribution` y `package`. `go test -race -count=1 ./tooling/...` completo también pasa sobre el candidato final (`cli` tardó 511.7 s).
- [x] Commit `c2986f9`, solo con las 10 rutas del cambio (gitleaks sin hallazgos), y push a `rebuild/harness-engineering`; la cabeza remota es `c2986f9`.
- [x] Issue nuevo para M10: [#42](https://github.com/JhonHawk/tricell-hive-private/issues/42), marcado como diferido en el título, con la tabla de funciones y la condición para extraer el despacho.
- [x] #40 cerrado con un comentario que cita `c2986f9` y #42.
- [x] Cierre y archivo del cambio, con push (commit de archivo).

**Depende de:** T1.
**Ejecución:** hilo principal (integración, revisión, tracker y Git).
**Verificación:**
- La cabeza remota coincide con la local.
- `gh issue view 40 --json state` devuelve `CLOSED`.
- El issue de M10 existe.

## Verificación compartida

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Caracterización | T1, local | La misma identidad antes y después en los siete casos | Incluido en la implementación |
| Tests y vet | T2, local | `go vet ./...` y `go test -race ./...` pasan | Incluido en la implementación |
| Revisión de código | T2, antes del commit | `/code-review` sin bloqueantes abiertos | E2-A |
| Verificación in vivo | — | No aplica: el comportamiento observable no cambia, y los tests de `bootstrap`, `cli` e `install` ejercitan el camino real | — |

## Revisión y progreso

**Revisión del plan** (revisión `cb6ac831f0f6`): un `review-plan` nativo de solo lectura. Hay un solo dominio: identidad de producto en el tooling Go.

- **B1 aplicado:** A1 no se podía cumplir en `ValidateBootstrapPackage`, porque `VerifyIfPackaged` ya lee el manifiesto dos veces antes. Se reformula como "sin relectura para el digest", y `VerifyIfPackaged`/`verifyManifest` quedan fuera de alcance.
- **N1 aplicado:** dos casos más de caracterización (manifiesto con un `VERSION` inválido y manifiesto legado).
- **N2 aplicado:** los valores esperados se calculan sin el código de producción, y la alteración se hace en `versioning.go`.
- **N3 aplicado:** el error de `Lstat` se devuelve sin envolver.
- **N4 aplicado:** la variable `version` se renombra y el comentario se actualiza.

Todos son propuestas del propio revisor, aplicadas tal cual, así que se concilian sin otra ronda.

**Límites de la revisión:** no se ejecutaron tests, y el revisor leyó el resumen del issue que da el plan, no su texto completo.

**Estado:** listo para implementar, con la entrega decidida (E1-A, E2-A). La ejecución espera la confirmación para `flow-build`.
