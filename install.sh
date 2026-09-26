#!/bin/sh
# Offline package entrypoint. Run the Go manager for every install operation.
set -eu
package_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)
case $(uname -s) in
  Darwin) platform_os=darwin ;;
  Linux) platform_os=linux ;;
  *) printf '%s\n' 'Sistema no soportado: usa macOS o Linux.' >&2; exit 1 ;;
esac
case $(uname -m) in
  arm64|aarch64) platform_arch=arm64 ;;
  x86_64|amd64) platform_arch=amd64 ;;
  *) printf '%s\n' 'Arquitectura no soportada: usa ARM64 o AMD64.' >&2; exit 1 ;;
esac
if [ ! -f "$package_dir/platform" ] || [ ! -f "$package_dir/bin/hive.sha256" ] || [ ! -f "$package_dir/release.json" ]; then
  printf '%s\n' 'Download and extract the complete Hive package for your system; the source archive does not include the executable.' >&2
  exit 1
fi
IFS= read -r expected_platform < "$package_dir/platform"
if [ "$expected_platform" != "$platform_os/$platform_arch" ]; then
  printf 'Este paquete es para %s; tu sistema es %s/%s.\n' "$expected_platform" "$platform_os" "$platform_arch" >&2
  exit 1
fi
if [ -L "$package_dir/bin" ] || [ -L "$package_dir/bin/hive" ] || [ ! -x "$package_dir/bin/hive" ]; then
  printf '%s\n' 'Executable is missing, linked, or not executable.' >&2
  exit 1
fi
IFS= read -r expected_hash < "$package_dir/bin/hive.sha256"
if command -v shasum >/dev/null 2>&1; then
  actual_hash=$(shasum -a 256 "$package_dir/bin/hive")
elif command -v sha256sum >/dev/null 2>&1; then
  actual_hash=$(sha256sum "$package_dir/bin/hive")
else
  printf '%s\n' 'Neither shasum nor sha256sum is available to verify the package.' >&2
  exit 1
fi
actual_hash=${actual_hash%% *}
if [ "$actual_hash" != "$expected_hash" ]; then
  printf '%s\n' 'El ejecutable no coincide con su checksum. Descarga nuevamente el paquete.' >&2
  exit 1
fi
exec "$package_dir/bin/hive" install "$@" --source "$package_dir"
