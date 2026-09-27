# Diseño

## Contexto verificado (sobre `b89bce4`)

- `productFromSource` ([versioning.go:45-80](../../../../tooling/management/versioning.go)) tiene tres ramas:
  - **Con `release.json`:** `version = m.ProductVersion` y `artifact = hash(bytes de release.json)`.
  - **Sin manifiesto pero con `VERSION`:** `version = TrimSpace(VERSION)` y `artifact = hash(version + "\n" + encode(r))`.
  - **Sin ninguno de los dos:** identidad `nil`.

  Un `VERSION` vacío también da `nil`. Un valor inválido lo rechaza `validateProduct` con "invalid product identity". Cualquier otro error de lectura se propaga.
- `version.ReadSource` ([version.go:23](../../../../tooling/version/version.go)) convierte la ausencia en `"dev"` y rechaza un valor inválido o vacío. Su único llamador es `tooling/package/main.go:67`.
- `distribution.ReadManifest` ([manifest.go:85](../../../../tooling/distribution/manifest.go)) comprueba con `Lstat` que el manifiesto sea un archivo regular, lo lee y lo decodifica. Cuando falta, devuelve el error `os.IsNotExist` intacto. `ValidateBootstrapPackage` ([bootstrap.go:101-120](../../../../tooling/distribution/bootstrap.go)) lo llama y después relee el archivo para `Digest`.
- `hash` (`tooling/management/types.go:97`) y `distribution.Digest` son SHA-256 en hexadecimal.
- **Cobertura actual:** `tooling/management/versioning_test.go` solo cubre la rama de `VERSION`. Ningún test de `management` fija el `artifact_id` de la rama del manifiesto.

## Contratos

**`distribution.ReadManifestData(root string) (Manifest, []byte, error)`:**
- Hace lo mismo que `ReadManifest` y además devuelve los bytes leídos.
- Mantiene los mismos errores, en el mismo orden: la ausencia devuelve el error de `Lstat` **sin envolver**, para que `os.IsNotExist` lo siga reconociendo (`versioning.go:58`, `manifest.go:67` y, de forma indirecta, `cli/install.go:162` dependen de eso), y hay error si no es un archivo regular o si el JSON es inválido.
- `ReadManifest` pasa a ser `m, _, err := ReadManifestData(root)`. Los demás llamadores no cambian.

**`version.ReadSourceFile(root string) (value string, present bool, err error)`:**
- Devuelve el contenido de `VERSION` sin espacios al inicio y al final, sin validar.
- Si el archivo falta: `("", false, nil)`.
- Cualquier otro error de lectura se propaga sin envolver.
- `ReadSource` pasa a ser: si el archivo no está, `"dev"`; si está, valida el valor como hoy (vacío o inválido da error).

**`productFromSource`:**
- Usa `ReadManifestData` en la rama del manifiesto y `ReadSourceFile` en la de `VERSION`.
- Conserva las tres ramas, el cálculo del `artifact` y la validación posterior con `validateProduct`: un `VERSION` vacío sigue dando `nil`, y uno inválido, "invalid product identity".
- Conserva también el orden: con manifiesto presente, `VERSION` no se lee. Un manifiesto sin `ProductVersion` da `nil` sin recurrir a `VERSION`.
- La variable local `version` (`versioning.go:49`) tapa el paquete `version`: se renombra, por ejemplo a `productVersion`.
- El comentario de `versioning.go:40-42` pasa a citar `ReadManifestData` y `ReadSourceFile`.

**`ValidateBootstrapPackage`:** usa `ReadManifestData` y calcula `Digest` sobre los bytes devueltos; se elimina la segunda lectura.

**Por qué no se valida `VERSION` dentro de `ReadSourceFile`:** los dos llamadores tratan distinto el valor vacío (identidad nula frente a error). Validar en el lector obligaría a cambiar uno de los dos comportamientos, y ese cambio está fuera de alcance.
