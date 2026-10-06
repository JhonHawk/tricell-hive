# Installer, release installation and legacy migration

Users download the complete archive for their operating system and architecture,
extract it, close affected CLI sessions, and run `./install.sh` from a terminal.
The automatic GitHub **Source code** archives do not contain an executable.
Installation does not require Go, Git, GitHub CLI, authentication in the terminal,
sudo, or PATH changes, and it downloads nothing except in the Pi case described
below. Private release access remains controlled
by the repository. Publishing a package is separate from building it.

```sh
./install.sh
./install.sh --dry-run
./install.sh --hosts codex,claude,grok,pi,opencode,cursor
```

The entrypoint verifies the package platform and binary checksum before invoking
the bundled manager. The manager verifies the complete package inventory before
planning. A source checkout remains usable through `go run ./tooling/cli install`.
The package installer never falls back to compiling or downloading missing files.
`hive --version` prints the product version of the running manager; a checkout
without `VERSION` or a release label is reported as a development build.

## Installing from a GitHub Release

Packages are published as GitHub Releases assets, one archive per supported
platform with its `.sha256` checksum file beside it. There is no one-line online
installer: `bootstrap.sh` was retired for the first public release (0.1.0). To
install, download the archive for your platform and its `.sha256` file from the
release page, verify the checksum, extract, and run the installer from a terminal.

```sh
shasum -a 256 -c hive-<version>-<os>-<arch>.tar.gz.sha256
tar -xzf hive-<version>-<os>-<arch>.tar.gz
cd hive-<version>-<os>-<arch>
./install.sh
```

The checksum file names the archive, so run `shasum -a 256 -c` in the directory
that holds both files. The checksum detects a corrupted or truncated download; it
is not an independent signature, because it comes from the same release page.

The `hive bootstrap` subcommand (an online hand-off that downloads and verifies a
package from a published origin) remains in the manager's code but no shipped
script invokes it. A follow-up decision will retire it or reconnect it. Until
then its recovery text for an interrupted online flow is unreachable in practice;
an interrupted offline installation is finished with `./install.sh` or
`hive recover`.

## Interaction

`install` detects CLI executables, registered consumers, and recognized legacy
resources. It displays selected hosts, destinations and private backup location,
then asks for one affirmative terminal confirmation. Cancellation, EOF and piped
input cannot apply changes. `--dry-run` is read-only and works without a terminal.
An already current installation exits without rewriting files or asking again.
No CLI executable is installed by Hive. `--hosts` selects configuration consumers;
shared legacy resources can require selecting additional consumers together.

After the hosts, the installer offers optional capabilities for the selected
hosts: Engram, Context7 and, only with the Pi host, pi-subagents. Their inventory
and limits are in [optional tool dependencies](tool-dependencies.md). This
release has no provider recipes: a selected capability is shown with its official
source and the reason it is manual, and no provider process runs. The one
exception is Pi: when the Pi host is selected in a user-scope install and Pi's
`settings.json` is missing or does not list pi-subagents, Hive itself runs
`pi install npm:pi-subagents@0.74.0` (the `pi` executable must be on `PATH`),
which downloads that package from npm. A settings file that already lists
pi-subagents makes Hive skip the step, and a conflicting entry stops the plan.
Package installation is therefore network-free except for this step. The core is
installed first; the installer then lists each selected capability with its
reason and next action and exits non-zero, so a manual capability is never
mistaken for an installed one, while the core remains installed.

If a transaction is pending, the installer offers recovery instead of installation.
After recovery, rerun the installer. Unknown or conflicting state is preserved.
Open new CLI sessions after success; filesystem verification does not demonstrate
model behavior or refresh existing sessions.

`--home` selects a synthetic home, ignoring environment path overrides and real
executables during discovery. `--state-dir` explicitly selects the private manager
state. These options support fixtures without touching the everyday installation.
Lower-level `plan`, `apply`, `status`, `remove` planning and `recover` remain available.
When migration edits include private legacy configuration, `plan --out` refuses to
save the plan outside protected state. Use `hive install` for that one-session
preview and apply flow; it keeps the plan in memory and journals private bytes
under the manager state directory.

## Legacy boundary

The embedded catalog recognizes Hive master revision
`16e7d3357a3c41530d5e31460c3024872566f3c7`. It checks known paths in effective and
historical default homes, the Claude path manifest, and shared/Pi hashed manifests.
Manifest claims never authorize arbitrary paths or content. Recognized bytes may
be migrated without a manifest. Unknown revisions, edited candidates or unsafe
links stop planning before installation writes.

Pi templates are compared after the historical `__HIVE_PI_ROOT__` substitution.
The known package patch is reversed only with exact pre/post hashes; the package
itself is preserved. Hook registrations are removed before their executables.
Shared resources require their affected consumers together. Extra user files,
credentials, histories, Engram, third-party configuration, backups and the legacy
source checkout are preserved. The old deployment script is never executed.

Migration uses the existing journal and recovery engine. State v5 records the
examined host/root coverage and detector version only after post-write verification.
Existing rebuild state without this receipt is still inspected. A recognized
reappearance of legacy files triggers another migration proposal. Old journal
hashes remain recoverable; older saved plans must be regenerated before applying.

macOS keeps `~/Library/Application Support/tricell-hive`. Linux defaults to
`$XDG_STATE_HOME/tricell-hive` or `~/.local/state/tricell-hive`; an existing state at
the historical macOS-style path is reused if it is the sole location. Two existing
locations require an explicit `--state-dir`; state is never silently moved.

## Building and verification

Maintainers need the Go version declared in `go.mod`. Build local packages with:

```sh
go run ./tooling/package --out ./dist
```

The builder freezes the source inputs, computes a source identity, and produces
macOS ARM64 and Linux ARM64/AMD64 archives with preserved executable modes and deterministic
archive metadata under `versions/<version>/`, plus raw managers, checksums and an
`index.json` release index (read only by the unused `hive bootstrap` subcommand). A version already present in
the output directory is refused before anything is built, so a published label is
never rebuilt with different content; build all platforms of a version in one run.
`--platforms linux/arm64` can select a subset. Building publishes nothing; a
release uploads only the `.tar.gz` archives and their `.sha256` files.
macOS Intel is unsupported and is rejected even when requested explicitly.

Run `go vet ./...` and `go test ./...` locally (and `go test -race ./...` for concurrency changes), followed by the actual
packaged installer in synthetic homes. Native execution coverage must be reported
separately from cross-compilation. Do not distribute an architecture merely because
its cross-build passed; validate installation and recovery in a compatible runtime.
Actual releases, signing/notarization arrangements if required by a recipient's
OS policy, and installation into real user homes remain separate delivery actions.
