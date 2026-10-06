# Security policy

## Supported versions

| Version | Supported |
| --- | --- |
| 0.1.x | Yes |
| Earlier | No |

## Reporting a vulnerability

Report vulnerabilities privately through GitHub's private vulnerability
reporting: open the repository's **Security** tab and choose **Report a
vulnerability**. Do not report vulnerabilities in public issues, pull requests, or
discussions.

Include the affected version, the host CLI involved, steps to reproduce, and the
impact you observed. Do not include real credentials or tokens in a report.

## What Hive does and does not do

- Hive installs guidance, skills, and role definitions into a host CLI's native
  locations and records what it owns so it can detect drift and recover.
- Hive does not handle credentials. It does not authenticate providers, request or
  store secrets, and it leaves each CLI's authentication and history untouched.
- Installing from a complete package makes no network requests of its own, with
  one exception: when the Pi host is selected and Pi's settings do not already
  declare pi-subagents, Hive runs `pi install npm:pi-subagents@0.74.0`, which
  downloads from npm. Building from a source checkout may download Go modules,
  and `hive update` uses Git.
- Hive does not install CLI executables. Engram and Context7 are separate,
  manually configured tools.
- The manager writes only to the destinations shown in its preview, after
  confirmation, and keeps private backups for recovery.
