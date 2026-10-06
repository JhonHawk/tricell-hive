# Identidad de producto: una sola lectura del manifiesto y de `VERSION`

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `rebuild/harness-engineering` (`c2986f9`); #40 cerrado y M10 movido a #42 |
| Tracker · GitHub Issues | • [#40 — Instalador: deuda pendiente del audit (identidad de versión y funciones largas)](https://github.com/JhonHawk/tricell-hive-private/issues/40) |
| Git | `direct-base` a `rebuild/harness-engineering` · sin PR · sin release |
| Verificación | test de caracterización del `artifact_id` · `go vet ./...` · `go test -race ./...` · `/code-review` |
| Siguiente paso | Ninguno en este cambio; M10 queda en [#42](https://github.com/JhonHawk/tricell-hive-private/issues/42), diferido |

## Objetivo

El audit del instalador versionado dejó M9 pendiente. `productFromSource` ([versioning.go:45](../../../../tooling/management/versioning.go)) lee `release.json` dos veces: una para la estructura y otra para el hash que se guarda como `artifact_id`. Además lee `VERSION` con su propio `os.ReadFile` en vez del lector de `tooling/version`. `ValidateBootstrapPackage` ([bootstrap.go:101](../../../../tooling/distribution/bootstrap.go)) repite la misma doble lectura del manifiesto. El cambio deja una sola lectura del manifiesto y un solo lector de `VERSION`, sin cambiar ninguna identidad calculada.

## Alcance y aceptación

**Incluye:**
- `distribution.ReadManifestData`, que devuelve la estructura y los bytes en una sola lectura, reusada por `productFromSource` y `ValidateBootstrapPackage`.
- `version.ReadSourceFile`, que distingue un `VERSION` ausente de uno presente, reusada por `ReadSource` y `productFromSource`.
- Un test de caracterización de `productFromSource` que fije la identidad antes del cambio.
- Un issue nuevo para M10, marcado como diferido.
- Cerrar #40 al integrar.

**Excluye:**
- **M10 (funciones largas):** pasa a un issue propio, porque el propio #40 recomienda extraerlas cuando se vuelvan a tocar y no con un refactor aislado (D1-A).
- **Cambiar la semántica de `VERSION`:** un archivo vacío sigue significando "sin versión" para la identidad, y "inválido" para `ReadSource`, como hoy.
- **Release:** el cambio toca el código del gestor, no el contenido que se instala.

**Criterios de aceptación:**
- A1. Ni `productFromSource` ni `ValidateBootstrapPackage` releen `release.json` para calcular el digest: el `artifact_id` sale de los mismos bytes que se decodificaron. Las lecturas de `VerifyIfPackaged`/`verifyManifest`, anteriores a este paso, quedan fuera de alcance.
- A2. `VERSION` se lee en un solo lugar (`tooling/version`), y `productFromSource` ya no llama a `os.ReadFile` directamente.
- A3. Para un mismo árbol, `productFromSource` devuelve la misma identidad (`version`, `artifact_id`, `release_id`) antes y después, en siete casos:
  - con manifiesto;
  - con manifiesto y un `VERSION` ilegible, que no se lee;
  - con un manifiesto legado sin `product_version`, que da `nil` sin recurrir a `VERSION`;
  - con `VERSION`;
  - sin ninguno de los dos;
  - con `VERSION` vacío.

  - con un `VERSION` inválido y sin manifiesto, que sigue dando el error "invalid product identity".
- A4. `go vet ./...` y `go test -race ./...` pasan, y `/code-review` no deja hallazgos bloqueantes sin resolver o refutar.

## Decisiones del usuario (2026-09-26)

- **D1-A:** solo M9. M10 pasa a un issue diferido, y #40 se cierra al integrar M9.
- **E1-A, `direct-base`:** un commit con las rutas del cambio y esta carpeta, con push a `rebuild/harness-engineering`; después, el commit de cierre con el archivo del cambio, también con push. Sin PR.
- **E2-A, revisión:** `/code-review` de Claude Code sobre el diff antes del commit.
- **Tracker:** D1-A autoriza abrir el issue de M10 y cerrar #40; la guía global autoriza cerrar #40 al integrar.
