#!/usr/bin/env bash
# sessions-archive.sh — classify session folders under a specs repo's `sessions/`
# and (with --apply) move the closed ones older than N days to `sessions/archived/`.
#
# Deterministic, token-free: the /workspace-archive skill runs it and reads its
# output. Dry-run by default — nothing moves or is written without --apply, and
# the script never commits or pushes; it stages and prints the suggested commit.
#
#   sessions-archive.sh [--sessions <dir>] [--days N] [--apply] [--normalize] [--json]
#
# Resolution of `sessions/` (no --sessions): cwd is a specs repo (./sessions/
# exists) → that; otherwise walk up for `_support/PROJECT.md` and take
# `<root>/*-specs/sessions` or `<root>/_support/sessions` (same rule as the
# flow-context hook). Nothing resolved → exit 2.
#
# Archive mode buckets (in output order): archivable · stalled · review · recent.
# Normalize mode (one-off migration of legacy `Status:` lines): resolved ·
# unresolved · status-inside-contract. Only the `Status:` line of a plan is
# ever rewritten, and only outside the hive-plan contract block.
#
# Age = days since the last commit touching the folder, skipping the commits
# this script itself suggests for --normalize (`chore(sessions): normalize …`):
# a one-off Status migration must not make 50 closed sessions look fresh.
#
# Exit codes: 0 ok · 1 an --apply step failed · 2 usage / resolution error.

set -uo pipefail

readonly CANON_STATUSES="draft planned building built verified"
readonly INDEX_SIGNAL_RE='✅|finaliz|done|conclu|verified|prod'
readonly LEGACY_UNRESOLVED_RE='fase|phase|parcial|partial|pending|diferid'
readonly LEGACY_VERIFIED_RE='verified|done|completed|shipped|concluded|finaliz|qa-verified'
readonly LEGACY_BUILDING_RE='in-progress|in progress|building'
readonly CONTRACT_START='<!-- hive-plan:contract:start -->'
readonly CONTRACT_END='<!-- hive-plan:contract:end -->'
readonly NORMALIZE_COMMIT_PREFIX='chore(sessions): normalize'

usage() {
  cat >&2 <<'EOF'
Uso: sessions-archive.sh [--sessions <dir>] [--days N] [--apply] [--normalize] [--json]

  --sessions <dir>  Carpeta sessions/ a procesar (default: se resuelve desde cwd).
  --days N          Umbral de edad en días (default: 15).
  --apply           Ejecuta (git mv / reescritura). Sin él, solo reporta.
  --normalize       Migra líneas `Status:` legacy a un estado canónico (no archiva).
  --json            Salida JSON en lugar de la tabla en español.
EOF
}

die() { printf 'sessions-archive: %s\n' "$*" >&2; exit 2; }

# ─── arguments ───────────────────────────────────────────────────────────────
sessions_arg=""; days=15; apply=0; normalize=0; json=0
while [ $# -gt 0 ]; do
  case "$1" in
    --sessions) [ $# -ge 2 ] || die "--sessions requiere un directorio"; sessions_arg="$2"; shift 2 ;;
    --sessions=*) sessions_arg="${1#--sessions=}"; shift ;;
    --days) [ $# -ge 2 ] || die "--days requiere un número"; days="$2"; shift 2 ;;
    --days=*) days="${1#--days=}"; shift ;;
    --apply) apply=1; shift ;;
    --normalize) normalize=1; shift ;;
    --json) json=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) usage; die "argumento desconocido: $1" ;;
  esac
done
[[ "$days" =~ ^[0-9]+$ ]] || die "--days debe ser un entero no negativo (recibido: $days)"

# ─── resolve sessions home ───────────────────────────────────────────────────
resolve_sessions() {
  # prints the sessions dir or nothing
  local dir ledger root specs
  if [ -n "$sessions_arg" ]; then
    [ -d "$sessions_arg" ] || die "--sessions: no existe el directorio $sessions_arg"
    (cd "$sessions_arg" && pwd -P)
    return
  fi
  if [ -d "$PWD/sessions" ]; then
    (cd "$PWD/sessions" && pwd -P)
    return
  fi
  dir="$PWD"; ledger=""
  while [ -n "$dir" ] && [ "$dir" != "/" ]; do
    if [ -f "$dir/_support/PROJECT.md" ]; then ledger="$dir/_support/PROJECT.md"; break; fi
    dir=$(dirname "$dir")
  done
  [ -n "$ledger" ] || return 0
  root=${ledger%/_support/PROJECT.md}
  specs=$(find "$root" -maxdepth 2 -type d -path '*-specs/sessions' 2>/dev/null | sort | head -1)
  if [ -n "$specs" ]; then
    (cd "$specs" && pwd -P)
  elif [ -d "$root/_support/sessions" ]; then
    (cd "$root/_support/sessions" && pwd -P)
  fi
}

sessions=$(resolve_sessions)
[ -n "$sessions" ] || die "no se resolvió sessions/: no hay ./sessions en cwd ni un ledger _support/PROJECT.md hacia arriba (usa --sessions <dir>)"

repo_root=$(git -C "$sessions" rev-parse --show-toplevel 2>/dev/null || true)
if [ "$apply" = "1" ] && [ -z "$repo_root" ]; then
  die "--apply requiere que $sessions esté dentro de un repo git (git mv / git add)"
fi

# ─── helpers ─────────────────────────────────────────────────────────────────
# Ages are CALENDAR-day differences (midnight to midnight), never seconds since
# now: BSD `date -j -f '%Y-%m-%d'` fills the missing time with the current
# time-of-day, which made a 30-day-old folder read 29 or 30 by the second.
date_to_epoch() {
  # YYYY-MM-DD → epoch at local midnight (BSD date first, GNU fallback)
  date -j -f '%Y-%m-%d %H:%M:%S' "$1 00:00:00" +%s 2>/dev/null || date -d "$1 00:00:00" +%s 2>/dev/null
}

epoch_to_date() {
  # epoch → YYYY-MM-DD (BSD date first, GNU fallback)
  date -r "$1" +%Y-%m-%d 2>/dev/null || date -d "@$1" +%Y-%m-%d 2>/dev/null
}

file_mtime() {
  stat -f %m "$1" 2>/dev/null || stat -c %Y "$1" 2>/dev/null
}

today=$(date +%Y-%m-%d)
today_epoch=$(date_to_epoch "$today")

folder_age() {
  # folder_age <abs folder> → prints "<days> <git|mtime>"
  local folder="$1" d epoch="" latest=0 m
  if [ -n "$repo_root" ]; then
    d=$(git -C "$repo_root" log -1 --format=%cs --invert-grep --fixed-strings --grep="$NORMALIZE_COMMIT_PREFIX" -- "$folder" 2>/dev/null || true)
    [ -n "$d" ] && epoch=$(date_to_epoch "$d")
  fi
  if [ -n "$epoch" ]; then
    printf '%s git' "$(( (today_epoch - epoch) / 86400 ))"
    return
  fi
  while IFS= read -r m; do
    [ -n "$m" ] || continue
    [ "$m" -gt "$latest" ] && latest="$m"
  done < <(find "$folder" -type f -print0 2>/dev/null | while IFS= read -r -d '' f; do file_mtime "$f"; done)
  [ "$latest" -gt 0 ] || latest=$(file_mtime "$folder")
  epoch=$(date_to_epoch "$(epoch_to_date "$latest")")
  printf '%s mtime' "$(( (today_epoch - epoch) / 86400 ))"
}

plan_file() {
  # plan_file <folder> <slug> → path of the plan or nothing (prefers <slug>-plan.md)
  local folder="$1" slug="$2"
  if [ -f "$folder/$slug-plan.md" ]; then printf '%s' "$folder/$slug-plan.md"; return; fi
  find "$folder" -maxdepth 1 -type f -name '*-plan.md' 2>/dev/null | sort | head -1
}

status_line_no() {
  # first `^Status:` line number, or nothing
  grep -n -m1 -E '^Status:' "$1" 2>/dev/null | cut -d: -f1
}

status_value() {
  # status_value <plan> → the trimmed value after `Status:` on its first line
  grep -m1 -E '^Status:' "$1" 2>/dev/null | sed -E 's/^Status:[[:space:]]*//; s/[[:space:]]+$//; s/\r$//'
}

is_canon() {
  local s
  for s in $CANON_STATUSES; do [ "$1" = "$s" ] && return 0; done
  return 1
}

status_in_contract() {
  # status_in_contract <plan> <status line no> → 0 when the line sits inside the contract block
  local plan="$1" ln="$2" start end
  start=$(grep -n -m1 -F "$CONTRACT_START" "$plan" 2>/dev/null | cut -d: -f1)
  end=$(grep -n -m1 -F "$CONTRACT_END" "$plan" 2>/dev/null | cut -d: -f1)
  [ -n "$start" ] && [ -n "$end" ] && [ "$ln" -gt "$start" ] && [ "$ln" -lt "$end" ]
}

# ─── README index: Estado cell for a slug ────────────────────────────────────
readme="$sessions/README.md"
index_state_for() {
  # index_state_for <folder name> <slug> → prints the Estado cell of the matching row, or nothing
  [ -f "$readme" ] || return 0
  awk -v folder="$1" -v slug="$2" '
    function trim(s) { sub(/^[ \t]+/, "", s); sub(/[ \t]+$/, "", s); return s }
    BEGIN { slug_col = 0; state_col = 0; found = 0 }
    /^[ \t]*\|/ {
      n = split($0, cells, "|")
      # cells[1] is the text before the first pipe (empty); real cells start at 2
      if (!found) {
        for (i = 2; i <= n; i++) {
          c = tolower(trim(cells[i]))
          if (c == "slug" || c == "sesión" || c == "sesion" || c == "session" || c == "carpeta") slug_col = i
          if (c == "estado" || c == "state" || c == "status") state_col = i
        }
        if (slug_col && state_col) { found = 1; next }
      }
      if (!found) next
      if (index(cells[slug_col], "`" folder "`") || index(cells[slug_col], "`" slug "`")) {
        print trim(cells[state_col]); exit
      }
    }
  ' "$readme"
}

# ─── collect candidates ──────────────────────────────────────────────────────
items=""   # JSONL, one object per candidate
add_item() { items="${items}$1"$'\n'; }

while IFS= read -r folder; do
  [ -n "$folder" ] || continue
  name=$(basename "$folder")
  case "$name" in archived|previously) continue ;; esac
  [[ "$name" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}-.+$ ]] || continue
  slug="${name:11}"

  plan=$(plan_file "$folder" "$slug")
  status="noplan"; status_raw=""; plan_rel=""; status_ln=""; in_contract=false
  if [ -n "$plan" ]; then
    plan_rel="${plan#"$sessions"/}"
    status_ln=$(status_line_no "$plan")
    if [ -n "$status_ln" ]; then
      status_raw=$(status_value "$plan")
      if is_canon "$status_raw"; then status="$status_raw"; else status="legacy"; fi
      status_in_contract "$plan" "$status_ln" && in_contract=true
    else
      status="noplan"   # a plan file without any Status line counts as no plan
    fi
  fi

  read -r age age_source < <(folder_age "$folder")

  index_state=""; index_signal=false
  if [ "$status" = "noplan" ]; then
    index_state=$(index_state_for "$name" "$slug")
    if [ -n "$index_state" ] && printf '%s' "$index_state" | grep -qiE "$INDEX_SIGNAL_RE"; then
      index_signal=true
    fi
  fi

  # bucket + reason (archive mode)
  bucket=""; reason=""
  case "$status" in
    verified)
      if [ "$age" -gt "$days" ]; then bucket=archivable; reason="verified"; else bucket=recent; fi ;;
    noplan)
      if [ "$index_signal" = true ]; then
        if [ "$age" -gt "$days" ]; then bucket=archivable; reason="sin plan; índice: $index_state"; else bucket=recent; fi
      else
        bucket=review
        if [ -n "$index_state" ]; then reason="sin plan; índice sin señal de cierre: $index_state"
        else reason="sin plan y sin fila en el índice"; fi
      fi ;;
    legacy)
      bucket=review; reason="Status legacy: $status_raw" ;;
    planned|building|built)
      if [ "$age" -gt "$days" ]; then bucket=stalled; reason="$status"; else bucket=active; fi ;;
    draft)
      if [ "$age" -gt "$days" ]; then bucket=review; reason="draft sin aprobar (> $days días)"; else bucket=active; fi ;;
  esac

  # normalize verdict (only legacy plans)
  norm_to=""; norm_reason=""
  if [ "$status" = "legacy" ]; then
    lower=$(printf '%s' "$status_raw" | tr '[:upper:]' '[:lower:]')
    if [ "$in_contract" = true ]; then
      norm_to=""; norm_reason="Status dentro del bloque de contrato; no se toca"
    elif printf '%s' "$lower" | grep -qE "$LEGACY_UNRESOLVED_RE"; then
      norm_reason="menciona fase/parcial/pendiente/diferido: requiere revisión humana"
    elif printf '%s' "$lower" | grep -qE "$LEGACY_VERIFIED_RE"; then
      norm_to=verified
    elif printf '%s' "$lower" | grep -qE "$LEGACY_BUILDING_RE"; then
      norm_to=building
    else
      norm_reason="sin correspondencia en el mapa legacy"
    fi
  fi

  add_item "$(jq -nc \
    --arg folder "$name" --arg slug "$slug" --arg status "$status" --arg status_raw "$status_raw" \
    --arg plan "$plan_rel" --argjson status_line "${status_ln:-0}" --argjson in_contract "$in_contract" \
    --argjson age "$age" --arg age_source "$age_source" \
    --arg index_state "$index_state" --argjson index_signal "$index_signal" \
    --arg bucket "$bucket" --arg reason "$reason" \
    --arg norm_to "$norm_to" --arg norm_reason "$norm_reason" \
    '{folder: $folder, slug: $slug, status: $status,
      status_raw: (if $status_raw == "" then null else $status_raw end),
      plan: (if $plan == "" then null else $plan end),
      status_line: (if $status_line == 0 then null else $status_line end),
      status_in_contract: $in_contract,
      age_days: $age, age_source: $age_source,
      index_state: (if $index_state == "" then null else $index_state end),
      index_signal: $index_signal,
      bucket: $bucket, reason: $reason,
      normalize_to: (if $norm_to == "" then null else $norm_to end),
      normalize_reason: (if $norm_reason == "" then null else $norm_reason end)}')"
done < <(find "$sessions" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort)

# ─── apply ───────────────────────────────────────────────────────────────────
apply_errors=0
moved=""          # newline list of folder names actually moved
readme_rewritten=false
normalized=""     # newline list of plan paths rewritten

escape_re() { printf '%s' "$1" | sed -E 's/[][\/.*^$]/\\&/g'; }

if [ "$apply" = "1" ] && [ "$normalize" = "0" ]; then
  mkdir -p "$sessions/archived"
  while IFS= read -r name; do
    [ -n "$name" ] || continue
    if [ -e "$sessions/archived/$name" ]; then
      printf 'sessions-archive: ya existe archived/%s; se omite\n' "$name" >&2
      apply_errors=$((apply_errors + 1)); continue
    fi
    if git -C "$repo_root" mv "$sessions/$name" "$sessions/archived/$name" 2>/dev/null; then
      moved="${moved}${name}"$'\n'
      if [ -f "$readme" ]; then
        esc=$(escape_re "$name")
        # ](X/ · ](./X/ · ](X) · ](./X)  →  ](archived/X…
        sed -E -i.bak "s#\]\((\./)?${esc}(/|\))#](archived/${name}\2#g" "$readme" && rm -f "$readme.bak"
        readme_rewritten=true
      fi
    else
      printf 'sessions-archive: git mv falló para %s\n' "$name" >&2
      apply_errors=$((apply_errors + 1))
    fi
  done < <(printf '%s' "$items" | jq -r 'select(.bucket == "archivable") | .folder')
  if [ "$readme_rewritten" = true ]; then
    git -C "$repo_root" add -- "$readme" 2>/dev/null || apply_errors=$((apply_errors + 1))
  fi
fi

if [ "$apply" = "1" ] && [ "$normalize" = "1" ]; then
  while IFS=$'\t' read -r plan_rel ln canon raw; do
    [ -n "$plan_rel" ] || continue
    plan="$sessions/$plan_rel"
    tmp=$(mktemp)
    if awk -v ln="$ln" -v canon="$canon" -v raw="$raw" -v today="$today" '
        NR == ln { print "Status: " canon; print "<!-- legacy status (normalized " today "): " raw " -->"; next }
        { print }
      ' "$plan" > "$tmp" && mv "$tmp" "$plan"; then
      normalized="${normalized}${plan_rel}"$'\n'
      git -C "$repo_root" add -- "$plan" 2>/dev/null || apply_errors=$((apply_errors + 1))
    else
      rm -f "$tmp"
      printf 'sessions-archive: no se pudo reescribir %s\n' "$plan_rel" >&2
      apply_errors=$((apply_errors + 1))
    fi
  done < <(printf '%s' "$items" | jq -r 'select(.normalize_to != null) | [.plan, .status_line, .normalize_to, .status_raw] | @tsv')
fi

# ─── report ──────────────────────────────────────────────────────────────────
moved_json=$(printf '%s' "$moved" | jq -R -s 'split("\n") | map(select(length > 0))')
normalized_json=$(printf '%s' "$normalized" | jq -R -s 'split("\n") | map(select(length > 0))')
moved_count=$(printf '%s' "$moved_json" | jq 'length')
commit_msg=""
if [ "$normalize" = "0" ]; then
  n=$(printf '%s' "$items" | jq -s 'map(select(.bucket == "archivable")) | length')
  [ "$apply" = "1" ] && n="$moved_count"
  commit_msg="chore(sessions): archive $n closed session(s) older than $days days"
else
  n=$(printf '%s' "$items" | jq -s 'map(select(.normalize_to != null)) | length')
  [ "$apply" = "1" ] && n=$(printf '%s' "$normalized_json" | jq 'length')
  commit_msg="$NORMALIZE_COMMIT_PREFIX $n legacy plan status line(s)"
fi

mode=archive; [ "$normalize" = "1" ] && mode=normalize
applied=false; [ "$apply" = "1" ] && applied=true

report=$(printf '%s' "$items" | jq -s \
  --arg mode "$mode" --arg sessions "$sessions" --argjson days "$days" --argjson apply "$applied" \
  --arg today "$today" --argjson moved "$moved_json" --argjson normalized "$normalized_json" \
  --argjson readme_rewritten "$readme_rewritten" --arg commit "$commit_msg" --argjson errors "$apply_errors" '
  def strip: map(del(.bucket, .reason, .normalize_to, .normalize_reason));
  if $mode == "archive" then
    {mode: $mode, sessions: $sessions, days: $days, apply: $apply, today: $today,
     total: length,
     archivable: map(select(.bucket == "archivable")) | map(del(.normalize_to, .normalize_reason)),
     stalled:    map(select(.bucket == "stalled"))    | map(del(.normalize_to, .normalize_reason)),
     review:     map(select(.bucket == "review"))     | map(del(.normalize_to, .normalize_reason)),
     recent:     {count: (map(select(.bucket == "recent")) | length), folders: map(select(.bucket == "recent") | .folder)},
     active:     {count: (map(select(.bucket == "active")) | length), folders: map(select(.bucket == "active") | .folder)},
     applied: (if $apply then {moved: $moved, readme_rewritten: $readme_rewritten, errors: $errors} else null end),
     commit_message: $commit}
  else
    {mode: $mode, sessions: $sessions, apply: $apply, today: $today,
     total: length,
     resolved:   map(select(.status == "legacy" and .normalize_to != null)),
     unresolved: map(select(.status == "legacy" and .normalize_to == null and .status_in_contract == false)),
     in_contract: map(select(.status == "legacy" and .status_in_contract == true)),
     applied: (if $apply then {normalized: $normalized, errors: $errors} else null end),
     commit_message: $commit}
  end')

if [ "$json" = "1" ]; then
  printf '%s\n' "$report"
else
  printf '%s' "$report" | jq -r '
    def age: "\(.age_days) d" + (if .age_source == "mtime" then " (mtime)" else "" end);
    def estado: if .status == "legacy" then "legacy: \(.status_raw)" elif .status == "noplan" then "sin plan" else .status end;
    def cell: gsub("\\|"; "\\|");
    def table(rows; header; sep; f): if (rows | length) == 0 then "_(ninguna)_" else header, sep, (rows[] | f) end;
    "## Sesiones — \(.sessions)",
    "",
    "Modo: \(.mode) · umbral: \(.days // "—") días · \(if .apply then "APLICADO" else "dry-run (nada se movió ni se escribió)" end) · fecha: \(.today) · candidatas: \(.total)",
    "",
    (if .mode == "archive" then
      "### 1. Archivables (verified o cerradas en el índice, > \(.days) días)",
      "",
      table(.archivable; "| Carpeta | Estado | Edad | Índice |"; "|---|---|---|---|";
            "| `\(.folder)` | \(estado) | \(age) | \(.index_state // "—" | cell) |"),
      "",
      "### 2. Estancadas — no archivables (planned/building/built, > \(.days) días)",
      "",
      table(.stalled; "| Carpeta | Estado | Edad |"; "|---|---|---|";
            "| `\(.folder)` | \(estado) | \(age) |"),
      "",
      "### 3. Requieren revisión (Status legacy, sin plan y sin señal en el índice, draft vencido)",
      "",
      table(.review; "| Carpeta | Estado | Edad | Motivo |"; "|---|---|---|---|";
            "| `\(.folder)` | \(estado | cell) | \(age) | \(.reason | cell) |"),
      "",
      "### 4. Recientes",
      "",
      "Cerradas hace ≤ \(.days) días: \(.recent.count). Activas (draft/planned/building/built ≤ \(.days) días): \(.active.count).",
      (if .applied != null then
        "",
        "### Aplicado",
        "",
        "Movidas a `archived/`: \(.applied.moved | length)" + (if (.applied.moved | length) > 0 then " (" + (.applied.moved | map("`\(.)`") | join(", ")) + ")" else "" end),
        "Índice `README.md` reescrito: \(if .applied.readme_rewritten then "sí" else "no" end) · errores: \(.applied.errors)",
        "Todo queda staged; sin commit. Commit sugerido:",
        "",
        "    \(.commit_message)"
      else
        "",
        "Con `--apply`: `git mv` de cada archivable a `archived/`, reescritura de enlaces en `README.md`, todo staged. Commit sugerido:",
        "",
        "    \(.commit_message)"
      end)
    else
      "### 1. Resueltas (Status legacy → canónico)",
      "",
      table(.resolved; "| Carpeta | Plan | Status actual | → | Edad |"; "|---|---|---|---|---|";
            "| `\(.folder)` | `\(.plan)` | \(.status_raw | cell) | \(.normalize_to) | \(age) |"),
      "",
      "### 2. Sin resolver (revisión humana o del custodian; no se tocan)",
      "",
      table(.unresolved; "| Carpeta | Plan | Status actual | Motivo |"; "|---|---|---|---|";
            "| `\(.folder)` | `\(.plan)` | \(.status_raw | cell) | \(.normalize_reason | cell) |"),
      "",
      "### 3. Status dentro del bloque de contrato (no se tocan; corregir a mano)",
      "",
      table(.in_contract; "| Carpeta | Plan | Línea |"; "|---|---|---|";
            "| `\(.folder)` | `\(.plan)` | \(.status_line) |"),
      (if .applied != null then
        "",
        "### Aplicado",
        "",
        "Planes reescritos: \(.applied.normalized | length) · errores: \(.applied.errors). Cada plan conserva la línea original como comentario `<!-- legacy status (normalized \(.today)): … -->` bajo `Status:`.",
        "Todo queda staged; sin commit. Commit sugerido:",
        "",
        "    \(.commit_message)"
      else
        "",
        "Con `--normalize --apply`: reemplaza la línea `Status:` por el canónico y añade el comentario legacy debajo; staged, sin commit. Commit sugerido:",
        "",
        "    \(.commit_message)"
      end)
    end)'
fi

[ "$apply_errors" -eq 0 ] || exit 1
exit 0
