#!/bin/sh
# Online bootstrap entrypoint. The deployment process replaces only this fixed
# placeholder with the public HTTPS origin. The shipped script carries no
# runtime override for it: a test-only copy of this file may have its origin
# and protocol-restriction lines substituted the same way deployment
# substitutes the placeholder below (see tooling/distribution's shell-level
# tests), but that substitution never ships and no environment variable can
# reach it.
set -eu

origin='https://HIVE_BOOTSTRAP_ORIGIN.invalid'
max_manager_bytes=134217728

usage() {
  printf '%s\n' 'Usage: bootstrap.sh [--version MAJOR.MINOR.PATCH[-PRERELEASE]] [--dry-run]'
}

fail() {
  printf '%s\n' "$1" >&2
  exit 1
}

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
  set -- $base
  IFS=$old_ifs
  [ "$#" -eq 3 ] || return 1
  for part in "$@"; do
    case $part in ''|*[!0-9]*|0[0-9]*) return 1 ;; esac
  done
  if [ -n "$prerelease" ]; then
    old_ifs=$IFS
    IFS=.
    set -- $prerelease
    IFS=$old_ifs
    for part in "$@"; do
      case $part in ''|*[!0-9A-Za-z-]*) return 1 ;; esac
      case $part in 0[0-9]*) return 1 ;; esac
    done
  fi
}

validate_checksum() {
  case $1 in
    ????????* ) ;;
    * ) return 1 ;;
  esac
  [ "${#1}" -eq 64 ] || return 1
  case $1 in *[!0123456789abcdef]*) return 1 ;; esac
}

limit_file() {
  file=$1
  limit=$2
  bytes=$(wc -c < "$file") || return 1
  [ "$bytes" -le "$limit" ] || return 1
}

download() {
  resource=$1
  target=$2
  limit=$3
  case $resource in ''|/*|*'..'*|*'\\'*) return 1 ;; esac
  url="$origin/$resource"
  if command -v curl >/dev/null 2>&1; then
    final_url=$(curl --fail --silent --show-error --location --proto '=https' --connect-timeout 20 --max-time 60 --max-filesize "$limit" --output "$target" --write-out '%{url_effective}' "$url") || return 1
    case $final_url in "$origin"/*) ;; *) return 1 ;; esac
  elif command -v wget >/dev/null 2>&1; then
    # Refuse all wget redirects because this portable client cannot inspect
    # each hop while preserving the same-origin rule.
    wget --quiet --https-only --max-redirect=0 --timeout=60 --tries=1 --output-document="$target" "$url" || return 1
  else
    return 1
  fi
  limit_file "$target" "$limit"
}

requested_version=
dry_run=0
while [ "$#" -gt 0 ]; do
  case $1 in
    --version)
      [ "$#" -ge 2 ] || fail '--version requires a product version.'
      requested_version=$2
      shift 2
      ;;
    --dry-run)
      dry_run=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      usage >&2
      exit 1
      ;;
  esac
done

if [ -n "$requested_version" ] && ! valid_version "$requested_version"; then
  fail 'Invalid product version.'
fi

if [ "$dry_run" -eq 1 ]; then
  if [ -n "$requested_version" ]; then
    printf 'Bootstrap dry run for Hive %s. It will require curl or wget and shasum or sha256sum. No download or installation is performed.\n' "$requested_version"
  else
    printf '%s\n' 'Bootstrap dry run for the stable Hive release. It will require curl or wget and shasum or sha256sum. No download or installation is performed.'
  fi
  exit 0
fi

# [ -r /dev/tty ] / [ -w /dev/tty ] only check the device node's static
# permission bits (normally world read-write), so they report readable and
# writable even with no controlling terminal at all. Actually open it
# read-write, which fails with ENXIO when the process has no controlling TTY.
# "exec" is a special builtin: a failed redirection on it exits the *shell
# executing it* per POSIX, with no message, rather than just returning
# nonzero (observed with dash). Run it inside a subshell so that exit only
# ends the subshell; the parent script keeps running and reaches fail below.
if ( exec 3<>/dev/tty ) 2>/dev/null; then
  :
else
  fail 'A readable and writable terminal is required; use the offline package without a TTY.'
fi
if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
  fail 'curl or wget is required to download the verified manager.'
fi
if ! command -v shasum >/dev/null 2>&1 && ! command -v sha256sum >/dev/null 2>&1; then
  fail 'shasum or sha256sum is required to verify the manager.'
fi

case $(uname -s) in
  Darwin) platform_os=darwin ;;
  Linux) platform_os=linux ;;
  *) fail 'Unsupported operating system: use macOS or Linux.' ;;
esac
case $(uname -m) in
  arm64|aarch64) platform_arch=arm64 ;;
  x86_64|amd64) platform_arch=amd64 ;;
  *) fail 'Unsupported architecture: use ARM64 or AMD64.' ;;
esac

umask 077
scratch=$(mktemp -d "${TMPDIR:-/tmp}/hive-bootstrap.XXXXXX") || fail 'Could not create private bootstrap scratch space.'
trap 'rm -rf "$scratch"' EXIT HUP INT TERM

if [ -z "$requested_version" ]; then
  download stable.txt "$scratch/stable.txt" 128 || fail 'Could not download a valid stable release selector.'
  [ "$(wc -l < "$scratch/stable.txt")" -eq 1 ] || fail 'Invalid stable release selector.'
  IFS= read -r requested_version < "$scratch/stable.txt" || true
  valid_version "$requested_version" || fail 'Invalid stable release selector.'
fi

platform="$platform_os-$platform_arch"
base="versions/$requested_version/$platform"
printf 'Downloading a verified temporary Hive manager for %s. It will show the installer before any persistent change.\n' "$requested_version"
download "$base/hive" "$scratch/hive" "$max_manager_bytes" || fail 'Could not download the Hive manager.'
download "$base/hive.sha256" "$scratch/hive.sha256" 128 || fail 'Could not download the manager checksum.'
[ "$(wc -l < "$scratch/hive.sha256")" -le 1 ] || fail 'Invalid manager checksum.'
IFS= read -r expected < "$scratch/hive.sha256" || true
validate_checksum "$expected" || fail 'Invalid manager checksum.'
if command -v shasum >/dev/null 2>&1; then
  actual=$(shasum -a 256 "$scratch/hive")
else
  actual=$(sha256sum "$scratch/hive")
fi
actual=${actual%% *}
[ "$actual" = "$expected" ] || fail 'The manager checksum did not match.'
chmod 700 "$scratch/hive"

# The manager owns index download, safe archive extraction, consent, and private retention.
"$scratch/hive" bootstrap --origin "$origin" --version "$requested_version" --manager "$scratch/hive" --manager-sha256 "$expected" < /dev/tty
