# deploy-global step 5 — purge the settings.json hook block of a repo-managed
# orphan .sh (manifest-confirmed). Invoke: jq --arg bn "$orphan_basename" -f this-file.
# Matches entries whose inner .command references the basename; drops event keys
# left empty. User hooks (never in the manifest) are never passed here.
if .hooks then
  .hooks |= ( with_entries( .value |= map(
      select( ((.hooks // []) | map(.command) | any(contains($bn))) | not )
  )) | with_entries( select(.value | length > 0) ) )
else . end
