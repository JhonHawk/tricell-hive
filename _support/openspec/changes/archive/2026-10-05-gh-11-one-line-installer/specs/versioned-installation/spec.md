## ADDED Requirements

### Requirement: One-line online installation
The repository SHALL provide `get-hive.sh`, a POSIX shell script that `https://hive.tricell.tech/install.sh` redirects to, which installs the latest published release, or the one named by `--version`, from GitHub Releases. Its origin SHALL be a fixed line that no environment variable or argument can change, every request and redirect hop SHALL use HTTPS, and the final URL of each request SHALL be on `github.com` or a subdomain of `githubusercontent.com`. It SHALL verify the downloaded package against its `.sha256` file, reject archive entries outside the package's single root directory, keep the package in `${XDG_DATA_HOME:-$HOME/.local/share}/hive/packages/<label>/` without changing `PATH`, and then run that package's `install.sh` with the remaining arguments, reading from the controlling terminal. Without `--dry-run` it SHALL require a controlling terminal before any network request. Its whole body SHALL run from a function called on its last line, so a truncated download executes nothing. The script SHALL NOT be part of the package.

#### Scenario: Integrity failure
- **WHEN** the checksum does not match, the `.sha256` file is malformed or names another file, the release tag is invalid, a redirect leaves the allowed hosts, a download exceeds its limit, or an archive entry falls outside the package root
- **THEN** the script exits non-zero without running `install.sh`, leaves no partial package directory, and keeps an existing directory for the same package unchanged.

#### Scenario: No terminal
- **WHEN** the script runs without a controlling terminal and without `--dry-run`
- **THEN** it exits non-zero before any network request.

#### Scenario: Preview
- **WHEN** the script runs with `--dry-run`
- **THEN** it downloads, verifies, and keeps the package, and runs `install.sh --dry-run`, reading from the controlling terminal when one opens and leaving its input unchanged otherwise.
