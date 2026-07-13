# deploy-global step 13 — idempotent additive merge of repo-managed hook blocks
# into ~/.claude/settings.json. Invoke: jq --argjson m "$managed" -f this-file.
# Identity is the inner .command string: managed entries are upserted (old
# version dropped, new appended); user entries never match and are preserved.
reduce ($m | to_entries[]) as $evt (.;
  .hooks[$evt.key] = (
    ($evt.value | map(.hooks[].command)) as $mcmds
    | ((.hooks[$evt.key] // [])
        | map( select(
            ((.hooks // []) | map(.command)) as $ec
            | ($ec | any(. as $c | ($mcmds | index($c)) != null)) | not
          )))
      + $evt.value
  ))
