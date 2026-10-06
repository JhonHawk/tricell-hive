#!/bin/sh
# get-hive.sh: one-line online installer for Hive.
#
#   curl -fsSL https://hive.tricell.tech/install.sh | sh
#   curl -fsSL https://hive.tricell.tech/install.sh | sh -s -- --dry-run
#   curl -fsSL https://hive.tricell.tech/install.sh | sh -s -- --version 0.1.0 --hosts claude,codex
#
# It resolves a Hive release, downloads the package for this platform and its
# checksum from GitHub Releases, verifies both, keeps the package in a stable
# folder, and hands control to the package's own install.sh, which keeps asking
# which hosts to install and for confirmation. Only --version and --help are
# handled here; every other argument is forwarded to install.sh unchanged.
#
# The settings below are single fixed lines. The shipped script reads none of
# them from the environment or from arguments. A test-only copy of this file may
# have exactly these lines replaced (see tooling/distribution/get_hive_test.go);
# that substitution never ships.
#
# Everything else lives in functions and the last line is the only call: a
# download cut short can never run a partial script, and a cut right after the
# call could not lose its arguments, because the call is wrapped in braces and
# stays an unterminated block until the final brace arrives.

repo_url='https://github.com/JhonHawk/tricell-hive'
proto_restriction='=https'
allowed_hosts='github.com .githubusercontent.com'
tty_device=/dev/tty
max_archive_bytes=67108864
max_checksum_bytes=256

usage() {
  printf '%s\n' 'Usage: get-hive.sh [--version X.Y.Z[-PRERELEASE]] [--dry-run] [install.sh options]'
  printf '%s\n' 'Downloads and verifies a Hive release package, keeps it under the data folder'
  printf '%s\n' '(XDG_DATA_HOME or ~/.local/share, then hive/packages), and runs its install.sh.'
  printf '%s\n' 'Without --dry-run a terminal is required; a preview uses it when available. Other arguments go to install.sh.'
}

say() {
  printf '%s\n' "$1"
}

fail() {
  printf 'get-hive: %s\n' "$1" >&2
  exit 1
}

# Accepts MAJOR.MINOR.PATCH with an optional -PRERELEASE, as the packages are named.
valid_version() {
  candidate=$1
  [ "${#candidate}" -le 128 ] || return 1
  case $candidate in
    ''|*[!0-9A-Za-z.-]*|.*|*.) return 1 ;;
  esac
  base=${candidate%%-*}
  prerelease=
  if [ "$base" != "$candidate" ]; then
    prerelease=${candidate#*-}
    [ -n "$prerelease" ] || return 1
    case $prerelease in *..*|.*|*.) return 1 ;; esac
  fi
  old_ifs=$IFS
  IFS=.
  # shellcheck disable=SC2086 # intentional splitting on dots; characters were validated above
  set -- $base
  IFS=$old_ifs
  [ "$#" -eq 3 ] || return 1
  for part in "$@"; do
    case $part in ''|*[!0-9]*|0[0-9]*) return 1 ;; esac
  done
  if [ -n "$prerelease" ]; then
    old_ifs=$IFS
    IFS=.
    # shellcheck disable=SC2086 # intentional splitting on dots; characters were validated above
    set -- $prerelease
    IFS=$old_ifs
    for part in "$@"; do
      case $part in ''|*[!0-9A-Za-z-]*) return 1 ;; esac
      case $part in 0[0-9]*) return 1 ;; esac
    done
  fi
  return 0
}

need_command() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

# A URL is allowed only when its host is exactly an entry of allowed_hosts, or,
# for entries that start with a dot, ends with that entry. The host is taken
# from the authority part only. The whole URL is never matched against a
# pattern, because a pattern also matches across "/" and would accept a
# disallowed host that merely carries an allowed name in its path.
host_allowed() {
  url=$1
  case $url in
    http://*|https://*) ;;
    *) return 1 ;;
  esac
  authority=${url#*://}
  authority=${authority%%[/?#]*}
  case $authority in
    ''|*@*) return 1 ;;
  esac
  host=${authority%%:*}
  case $host in
    ''|.*|*[!0-9A-Za-z.-]*) return 1 ;;
  esac
  for entry in $allowed_hosts; do
    case $entry in
      .*)
        case $host in *"$entry") return 0 ;; esac
        ;;
      *)
        [ "$host" = "$entry" ] && return 0
        ;;
    esac
  done
  return 1
}

# fetch URL OUTPUT LIMIT: download with every restriction, then check the final
# host and the real size on disk (--max-filesize only sees a declared length).
# -q must be the first argument so a user's ~/.curlrc is never read.
fetch() {
  fetch_url=$1
  fetch_out=$2
  fetch_limit=$3
  final_url=$(curl -q --fail --silent --show-error --location --max-redirs 5 \
    --proto "$proto_restriction" --proto-redir "$proto_restriction" \
    --connect-timeout 20 --max-time 600 --max-filesize "$fetch_limit" \
    --output "$fetch_out" --write-out '%{url_effective}' "$fetch_url" </dev/null) || return 1
  host_allowed "$final_url" || {
    printf 'get-hive: refusing a redirect to a host that is not allowed: %s\n' "${final_url%%[?#]*}" >&2
    return 1
  }
  fetch_bytes=$(wc -c < "$fetch_out") || return 1
  fetch_bytes=$((fetch_bytes + 0))
  [ "$fetch_bytes" -le "$fetch_limit" ]
}

# The latest release is where GitHub redirects /releases/latest to.
latest_version() {
  latest_url=$(curl -q --fail --silent --show-error --location --max-redirs 5 --head \
    --proto "$proto_restriction" --proto-redir "$proto_restriction" \
    --connect-timeout 20 --max-time 60 \
    --output /dev/null --write-out '%{url_effective}' "$repo_url/releases/latest" </dev/null) || return 1
  host_allowed "$latest_url" || return 1
  tag_prefix="$repo_url/releases/tag/v"
  case $latest_url in
    "$tag_prefix"*) latest_found=${latest_url#"$tag_prefix"} ;;
    *) return 1 ;;
  esac
  valid_version "$latest_found" || return 1
  version=$latest_found
}

sha256_of() {
  if command -v shasum >/dev/null 2>&1; then
    digest_line=$(shasum -a 256 "$1") || return 1
  else
    digest_line=$(sha256sum "$1") || return 1
  fi
  digest=${digest_line%% *}
}

# The checksum file must be exactly one line: <64 lowercase hex>, two spaces, the
# exact archive name, and a newline.
read_expected_digest() {
  sum_file=$1
  sum_name=$2
  IFS= read -r sum_line < "$sum_file" || return 1
  case $sum_line in
    *"  $sum_name") expected=${sum_line%"  $sum_name"} ;;
    *) return 1 ;;
  esac
  [ "${#expected}" -eq 64 ] || return 1
  case $expected in *[!0123456789abcdef]*) return 1 ;; esac
  sum_bytes=$(wc -c < "$sum_file") || return 1
  sum_bytes=$((sum_bytes + 0))
  [ "$sum_bytes" -eq $((${#sum_line} + 1)) ]
}

# Before extracting: every entry must be the root folder or live inside it, with
# no ".." component and no absolute path, and the verbose listing must show only
# folders and regular files, as the package builder writes them.
check_listing() {
  list_archive=$1
  list_label=$2
  list_names=$3
  list_verbose=$4
  tar -tzf "$list_archive" > "$list_names" 2>/dev/null || return 1
  tar -tvzf "$list_archive" > "$list_verbose" 2>/dev/null || return 1
  [ -s "$list_names" ] || return 1
  while IFS= read -r entry; do
    case $entry in
      /*|*/../*|*/..|../*|..) return 1 ;;
    esac
    case $entry in
      "$list_label"|"$list_label"/|"$list_label"/*) ;;
      *) return 1 ;;
    esac
  done < "$list_names"
  while IFS= read -r entry; do
    case $entry in
      d*|-*) ;;
      *) return 1 ;;
    esac
  done < "$list_verbose"
  names_count=$(wc -l < "$list_names") || return 1
  verbose_count=$(wc -l < "$list_verbose") || return 1
  [ "$((names_count + 0))" -eq "$((verbose_count + 0))" ]
}

# After extracting: only the root folder, as a real folder, containing only real
# folders and regular files that are not hard links, and an install.sh.
check_extracted() {
  extract_stage=$1
  extract_label=$2
  stray=$(cd "$extract_stage" && find . ! -path . ! -path "./$extract_label" ! -path "./$extract_label/*") || return 1
  [ -z "$stray" ] || return 1
  [ -d "$extract_stage/$extract_label" ] && [ ! -L "$extract_stage/$extract_label" ] || return 1
  odd=$(find "$extract_stage/$extract_label" ! -type d ! -type f) || return 1
  [ -z "$odd" ] || return 1
  linked=$(find "$extract_stage/$extract_label" -type f -links +1) || return 1
  [ -z "$linked" ] || return 1
  [ -f "$extract_stage/$extract_label/install.sh" ] && [ ! -L "$extract_stage/$extract_label/install.sh" ]
}

# Runs on exit and on INT, TERM and HUP. It removes the downloads and the staging
# folder, and if the previous package was moved aside it puts it back, unless
# something else now occupies its place, in which case it is kept and reported.
# It never deletes a package folder that existed before this run.
cleanup() {
  if [ "$swap_active" -eq 1 ]; then
    swap_active=0
    if [ ! -e "$final_dir" ] && [ ! -L "$final_dir" ] && mv "$old_dir/$label" "$final_dir" 2>/dev/null; then
      rmdir "$old_dir" 2>/dev/null
      old_dir=
    else
      printf 'get-hive: the previous package was kept at %s\n' "$old_dir/$label" >&2
      old_dir=
    fi
  fi
  if [ -n "$dl_dir" ]; then
    rm -rf "$dl_dir"
    dl_dir=
  fi
  if [ -n "$stage_dir" ]; then
    rm -rf "$stage_dir"
    stage_dir=
  fi
  if [ -n "$old_dir" ]; then
    rm -rf "$old_dir"
    old_dir=
  fi
}

# Moves the staged package to its final place. An existing package is first
# renamed into a hidden folder, so a failure at any later point can restore it.
install_package() {
  [ -d "$packages_dir" ] || mkdir -p "$packages_dir" || fail "could not create $packages_dir"
  stage_dir=$(mktemp -d "$packages_dir/.stage.XXXXXX") || fail 'could not create a staging folder'
  say 'Verifying the package layout...'
  tar -xzf "$dl_dir/$archive" -C "$stage_dir" || fail 'could not extract the package'
  check_extracted "$stage_dir" "$label" || fail 'the package has an unexpected layout'
  if [ -e "$final_dir" ] || [ -L "$final_dir" ]; then
    old_dir=$(mktemp -d "$packages_dir/.old.XXXXXX") || fail 'could not create a folder for the previous package'
    mv "$final_dir" "$old_dir/$label" || fail "could not move the previous package aside: $final_dir"
    swap_active=1
  fi
  if [ -e "$final_dir" ] || [ -L "$final_dir" ]; then
    fail "$final_dir appeared while installing; leaving it untouched"
  fi
  mv "$stage_dir/$label" "$final_dir" || fail "could not move the package into $final_dir"
  swap_active=0
}

main() {
  version=
  version_given=0
  dry_run=0
  # Rotate the arguments once: consume --version, keep everything else in order.
  remaining=$#
  while [ "$remaining" -gt 0 ]; do
    arg=$1
    shift
    remaining=$((remaining - 1))
    case $arg in
      --help)
        usage
        return 0
        ;;
      --version)
        [ "$remaining" -gt 0 ] || fail '--version requires a value'
        version=$1
        version_given=1
        shift
        remaining=$((remaining - 1))
        ;;
      --dry-run)
        dry_run=1
        set -- "$@" "$arg"
        ;;
      *)
        set -- "$@" "$arg"
        ;;
    esac
  done
  if [ "$version_given" -eq 1 ]; then
    valid_version "$version" || fail "invalid version: $version"
  fi

  for tool in curl tar mktemp uname find wc mv rm rmdir mkdir; do
    need_command "$tool"
  done
  if ! command -v shasum >/dev/null 2>&1 && ! command -v sha256sum >/dev/null 2>&1; then
    fail 'required command not found: shasum or sha256sum'
  fi
  [ -n "${HOME:-}" ] || fail 'HOME is not set'

  case $(uname -s) in
    Darwin) platform_os=darwin ;;
    Linux) platform_os=linux ;;
    *) fail 'unsupported operating system: use macOS or Linux' ;;
  esac
  case $(uname -m) in
    arm64|aarch64) platform_arch=arm64 ;;
    x86_64|amd64) platform_arch=amd64 ;;
    *) fail 'unsupported architecture: use ARM64 or AMD64' ;;
  esac

  # Opening the device is the only reliable test: [ -r /dev/tty ] only reads the
  # node's permission bits, which look fine even with no controlling terminal.
  # "exec" with a failing redirection exits the shell that runs it, so it runs in
  # a subshell and only that subshell exits.
  # Decided up front, before any request. A preview (--dry-run) uses the terminal
  # when one opens, so the installer can ask which hosts to preview; without one
  # it leaves stdin alone and install.sh then needs --hosts. A real installation
  # requires the terminal.
  if ( exec 3<>"$tty_device" ) 2>/dev/null; then
    have_tty=1
  else
    have_tty=0
    if [ "$dry_run" -eq 0 ]; then
      fail 'a terminal is required so the installer can ask questions; run it from a terminal, add --dry-run to only preview, or install from the package manually'
    fi
  fi

  data_home=${XDG_DATA_HOME:-}
  case $data_home in
    /*) ;;
    *) data_home=$HOME/.local/share ;;
  esac
  packages_dir=$data_home/hive/packages

  dl_dir=
  stage_dir=
  old_dir=
  swap_active=0
  final_dir=
  label=
  saved_umask=$(umask)
  umask 077
  trap cleanup EXIT
  trap 'cleanup; exit 130' INT
  trap 'cleanup; exit 143' TERM
  trap 'cleanup; exit 129' HUP

  TAR_OPTIONS=
  export TAR_OPTIONS

  dl_dir=$(mktemp -d "${TMPDIR:-/tmp}/hive-get.XXXXXX") || fail 'could not create a private download folder'

  if [ "$version_given" -eq 0 ]; then
    say 'Looking up the latest Hive release...'
    latest_version || fail "could not resolve the latest release from $repo_url/releases/latest"
  fi

  label=hive-$version-$platform_os-$platform_arch
  archive=$label.tar.gz
  final_dir=$packages_dir/$label
  release_url=$repo_url/releases/download/v$version

  say "Downloading Hive $version for $platform_os-$platform_arch..."
  fetch "$release_url/$archive.sha256" "$dl_dir/$archive.sha256" "$max_checksum_bytes" || fail 'could not download the checksum'
  read_expected_digest "$dl_dir/$archive.sha256" "$archive" || fail 'the checksum file is malformed'
  fetch "$release_url/$archive" "$dl_dir/$archive" "$max_archive_bytes" || fail 'could not download the package'
  sha256_of "$dl_dir/$archive" || fail 'could not compute the package checksum'
  [ "$digest" = "$expected" ] || fail 'the package checksum did not match; nothing was installed'
  check_listing "$dl_dir/$archive" "$label" "$dl_dir/names.txt" "$dl_dir/verbose.txt" || fail 'the package archive is not safe to extract; nothing was installed'

  install_package

  package=$final_dir
  cleanup
  trap - EXIT INT TERM HUP
  umask "$saved_umask"
  say "Hive package kept at $package"
  say "Hive executable: $package/bin/hive"
  # exec replaces this shell, so the EXIT trap would not run: cleanup ran above.
  if [ "$have_tty" -eq 0 ]; then
    exec "$package/install.sh" "$@"
  fi
  exec "$package/install.sh" "$@" < "$tty_device"
}

# The only top-level command. A cut anywhere before the closing brace leaves an
# unterminated block, which fails to parse and runs nothing.
{ main "$@" || exit 1; }
