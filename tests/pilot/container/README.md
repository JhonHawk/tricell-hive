# Reusable flash pilot container

T8 supplies the environment, not authorization to use bb1, log in or run models.
Host prerequisites: Docker, Bash and Python 3.9+; SSH/scp only for `upload`.
The image runs Linux/amd64 as UID/GID 1000 and uses Debian procps.

Pinned sources:

- Debian bookworm-slim amd64 manifest: `sha256:a4672c0cb26fbdde88e38fa2dfb6c681942306680e41e4378b28770b6e79ee91`.
- Engram 3.3.2 official Linux/amd64 release archive: SHA-256 `8694d3817b63c5fc5607b5919e88aa929bffa0bafda330513364e10c8ab61445`, checked both against the pinned value and release `checksums.txt` with `sha256sum -c`.
- npm `@opencode/cli-linux-x64@2.0.27`: integrity `sha512-QPGrzBSSs5ATG3nM4XjS4tmWXUB22Di36H0vfMuhKwk6i3PR3Nb4sm+l00gvTdcyCRohYsKwqjvD67T0DPcQLQ==`. Docker verifies its decoded SHA-512 before extraction.

Debian apt dependencies are not snapshot-pinned. A rebuild can change those
packages; record the resulting image ID and observed CLI versions. Preserve the
built image for reuse. `build` records its immutable local image ID; `create`
uses that ID rather than a mutable tag. Default tag: `hive-pilot-test`.

## Local Docker validation or execution on the authorized remote host

Use an explicit state directory outside source, containing manifests only.
All arguments below are placeholders, not permission to execute a pilot.

```bash
./bb1.sh --state-dir <state-directory> --context desktop-linux prepare <batch>
./bb1.sh --state-dir <state-directory> --context desktop-linux build <batch>
./bb1.sh --state-dir <state-directory> --context desktop-linux create <batch>
```

On bb1 omit `--context desktop-linux` and use its explicitly authorized Docker
context. `prepare` requires an explicit unique batch ID and writes a random
ownership token before creation. Names are `hive-pilot-<batch>` and
`hive-pilot-work-<batch>`. Both carry pilot, batch and ownership labels.
Creation records the inspected container ID and volume `CreatedAt` plus labels.
If a step is interrupted, keep the manifest: do not prepare the batch again.

The work volume owns `/work`, including `runs/` and `input/`. The shared
`hive-pilot-cache` owns only `/home/pilot/.cache` and survives normal destruction;
an existing volume without the pilot label is refused. Native OpenCode data
lives under the 0700 UID/GID 1000 tmpfs `/run/hive-pilot/secrets`. The runner's
per-run native homes use its `shadow-data/` subdirectory, separate from evidence.
No host mounts, published ports or Docker socket are used. Docker tmpfs can
reach host swap; this does not promise absence from swap or host access.

## Upload and stage

Compile the runner on the Mac, outside the image:

```bash
GOOS=linux GOARCH=amd64 go build -o <scratch>/pilot ./tests/pilot
```

Create a tar containing `candidate/`, `baseline/` and `fixtures/`. The baseline
must include its `.git` directory and have HEAD at `d5164f71` with clean payload
paths (`content`, `integrations/opencode/context-loader.js` and
`integrations/agent-profiles.json`); `stage` checks this before accepting it. Include only
pilot source inputs, no credentials, secrets, native data or support dumps.
Archives with links, absolute paths or traversal are rejected by `stage`.
`upload` sends the runner, bundle, manifest, launcher and Dockerfile via SSH;
it does not build, start a container, log in or run models remotely:

```bash
./bb1.sh --state-dir <state-directory> upload <batch> \
  --runner <scratch>/pilot --bundle <scratch>/sources.tar \
  --remote <ssh-host-alias> --remote-dir <absolute-remote-directory>
```

Invoke the copied launcher on the remote host with the uploaded manifest's
parent as `--state-dir`, then `build`, `create`, and:

```bash
./bb1.sh --state-dir <state-directory> stage <batch> \
  --runner <runner-file> --bundle <sources-tar>
```

`stage` writes input files as the container's non-root user and prints the
runner's actual `-h`. Verify those flags and input paths before using `run`.
The fixture directory remains in the supplied candidate layout; the runner
resolves its suite fixtures from `--source`, not a separate fixtures flag.

## User login, catalog and authorized runs

```bash
./bb1.sh --state-dir <state-directory> login <batch>
./bb1.sh --state-dir <state-directory> catalog <batch>
./bb1.sh --state-dir <state-directory> run <batch> --run-id <unique-run-id> -- \
  --case <fixture-case> --arm <A-or-B-or-C> --model <opencode-go-model-with-variant>
```

Login is interactive inside the container and connects the `opencode-go`
integration that the runner imports; a Console-only login leaves the Go models
unrouted. `catalog` also snapshots `https://models.dev/api.json` into the
container cache, because OpenCode 2 keeps no `models.json` and the runner
copies one into each shadow home. Never pass credentials in flags,
environment variables, manifests or files. Catalog checks the four planned Go
model IDs and retains the container's own public model cache. `run` owns the
isolation, output, source, host, suite and timeout flags; it accepts only case,
arm and model. Follow the retained plan's fixed schedule, quota, concurrency
limit and no-retry policy. No model execution is part of T8 verification.

## Collect, destroy and recover

```bash
./bb1.sh --state-dir <state-directory> collect <batch> --evidence <new-directory>
./bb1.sh --state-dir <state-directory> destroy <batch>
./bb1.sh --state-dir <state-directory> recover <batch> --evidence <new-directory>
```

Collection requires no active pilot/OpenCode/Engram process. It scans `/work`
for links, OpenCode databases and credential filenames before copying only
`runs/`; it compares the copy's SHA-256 hashes with hashes from the source and
adds `SHA256SUMS.json`. Failed copies leave resources and any partial copy
intact. Use a new evidence directory when explicitly resuming collection.
Do not collect raw credentials or deliberately put them into runner evidence.

`recover` collects then destroys, never rerunning a model. If creation was
interrupted or its container is missing, an owned work volume is mounted
read-only in a temporary, non-root helper from the same recorded pilot image.
The helper uses batch ownership labels, no network, ports, host binds or socket.
It applies the same credential refusal and source/copy checksum checks, and is
removed by ID even if collection fails. Only successful collection (or an
explicit check finding no run evidence) permits destruction of the owned work
volume. A prepared batch with no resources reports no evidence and is also
recoverable. It can start a
stopped batch's inert `sleep` container to inspect/copy evidence. A partial
manifest resolves exact resource names with all three ownership labels; recorded
IDs/creation timestamps must also match when present. Destruction validates
both resources before deleting either, attempts local provider logout when
available (with an account picker on an interactive terminal), removes by
inspected ID/name, and verifies absence. OpenCode 2.0.27 requires an account
identifier for noninteractive logout; without one the attempt fails, reports
that fact without credential output, and container deletion still discards the
tmpfs. Logout is not
provider revocation. It never removes the shared cache or calls `prune`.
If the runner reports an unverified process stop, keep the batch pending until
its process state has been independently resolved; do not use `destroy` to
claim that runner gate passed.

## Mac everyday Engram check

`count-projects.py` invokes `engram projects list` without overriding its store
and prints counts only. It checks the `flash-class-` prefix and optional exact
project IDs from result manifests. Exit failure means the check is unverified,
not zero. Confirm `engram version` is 3.3.2 and `engram --help` describes
`projects list`; Engram's subcommand help may still acquire its store lock.

```bash
python3 count-projects.py [<exact-project-id> ...]
```

Require zero prefix matches before the pilot and zero exact/prefix matches
afterward. This proves absence only in that everyday store. In this restricted
Mac session, Engram's mandatory instance identity lock prevents the everyday
store command from running; do not redirect it to a disposable store and claim
that checks the Mac store.

Docker option semantics were checked against current Docker documentation via
Context7 `/docker/docs` on 2026-10-12 (installed Docker 29.8.2):
https://docs.docker.com/engine/storage/tmpfs/ and
https://docs.docker.com/engine/storage/volumes/.

Deletion requires a collection receipt in the batch manifest: the evidence
directory, SHA256SUMS file hash and source file hash list. Before deleting,
`destroy` re-hashes `/work/runs` through the batch container or a read-only
helper, verifies the receipt and copied evidence, and refuses any mismatch.
An explicit current empty-evidence check also permits deletion when no receipt
exists. `recover` still collects first. Credential-name checks apply only to
`/work/runs`; candidate input filenames cannot block collection. Arm A uses
baseline guidance; arms B and C use candidate guidance.
