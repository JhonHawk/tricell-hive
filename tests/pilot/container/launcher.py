#!/usr/bin/env python3
"""Batch ownership and recovery for the explicitly authorized flash pilot."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import secrets
import shlex
import shutil
import subprocess
import sys
import tempfile

PILOT = "flash-class-guidance"
CACHE = "hive-pilot-cache"
MODELS = ("claude-haiku-5-5", "deepseek-v4.1-flash", "qwen3.8-flash", "glm-5.3-flash")
ROOT = Path(__file__).resolve().parent


def command(args, capture=True, **kwargs):
    result = subprocess.run(args, capture_output=capture, text=capture, **kwargs)
    if result.returncode:
        # Never relay arbitrary CLI stderr, which may contain authentication data.
        raise RuntimeError(f"command failed ({result.returncode}): {args[0]}")
    return result.stdout if capture else None


class Launcher:
    def __init__(self, args):
        self.args = args
        self.state = Path(args.state_dir).absolute()
        self.path = self.state / f"{args.batch}.json"
        self.docker = ["docker"] + (["--context", args.context] if args.context else [])
        self.data = None

    def call(self, *args, **kwargs):
        return command(self.docker + list(args), **kwargs)

    def inspect(self, kind, name):
        result = subprocess.run(self.docker + [kind, "inspect", name], capture_output=True, text=True)
        if result.returncode:
            if "no such" in result.stderr.lower() or "not found" in result.stderr.lower():
                return None
            raise RuntimeError("Docker inspection failed; refusing to infer absence")
        return json.loads(result.stdout)[0]

    def save(self):
        self.state.mkdir(mode=0o700, parents=True, exist_ok=True)
        with tempfile.NamedTemporaryFile(mode="w", dir=self.state, delete=False) as stream:
            temporary = Path(stream.name)
            json.dump(self.data, stream, indent=2)
            stream.write("\n")
            stream.flush()
            os.fsync(stream.fileno())
        os.replace(temporary, self.path)

    def load(self):
        self.data = json.loads(self.path.read_text())
        batch = self.args.batch
        if (self.data.get("batch") != batch or self.data.get("container") != f"hive-pilot-{batch}"
                or self.data.get("volume") != f"hive-pilot-work-{batch}"
                or not re.fullmatch(r"[0-9a-f]{64}", self.data.get("owner", ""))):
            raise RuntimeError("invalid batch manifest")
        if self.data.get("labels") != self.labels():
            raise RuntimeError("manifest labels disagree with batch ownership")

    def labels(self):
        return {"hive.pilot": PILOT, "hive.pilot.batch": self.args.batch,
                "hive.pilot.owner": self.data["owner"]}

    def label_args(self):
        return [arg for key, value in self.labels().items() for arg in ("--label", f"{key}={value}")]

    def owned(self, kind):
        name = self.data[kind]
        obj = self.inspect(kind, name)
        if obj is None:
            return None
        actual_name = obj["Name"].lstrip("/")
        labels = obj.get("Config", {}).get("Labels", {}) if kind == "container" else obj.get("Labels", {})
        if actual_name != name or any(labels.get(k) != v for k, v in self.labels().items()):
            raise RuntimeError(f"{kind} ownership mismatch; resource preserved")
        recorded_labels = self.data.get(f"{kind}_labels")
        if recorded_labels is not None and recorded_labels != labels:
            raise RuntimeError(f"{kind} recorded labels mismatch; resource preserved")
        recorded = self.data.get(f"{kind}_identity")
        identity = obj["Id"] if kind == "container" else obj["CreatedAt"]
        if recorded and recorded != identity:
            raise RuntimeError(f"{kind} identity mismatch; resource preserved")
        return obj

    def prepare(self):
        if self.path.exists():
            raise RuntimeError("manifest already exists")
        self.data = {"batch": self.args.batch, "owner": secrets.token_hex(32),
                     "container": f"hive-pilot-{self.args.batch}",
                     "volume": f"hive-pilot-work-{self.args.batch}"}
        for kind in ("container", "volume"):
            if self.inspect(kind, self.data[kind]):
                raise RuntimeError(f"{kind} name already exists")
        self.data["labels"] = self.labels()
        # Exclusive creation prevents a second prepare from replacing ownership.
        self.state.mkdir(mode=0o700, parents=True, exist_ok=True)
        with self.path.open("x") as stream:
            os.chmod(self.path, 0o600)
            json.dump(self.data, stream, indent=2)
            stream.flush()
            os.fsync(stream.fileno())
        print(self.path)

    def build(self):
        self.call("build", "--platform", "linux/amd64", "--progress", "plain", "-t", self.args.image,
                  str(ROOT), capture=False)
        image = json.loads(self.call("image", "inspect", self.args.image))[0]
        self.data["image"] = image["Id"]
        self.data["versions"] = self.call("run", "--rm", "--platform", "linux/amd64", *self.label_args(),
                                          image["Id"], "sh", "-c", "opencode --version && engram version").strip()
        self.save()
        print(json.dumps({k: self.data[k] for k in ("image", "versions")}))

    def create(self):
        for kind in ("container", "volume"):
            if self.inspect(kind, self.data[kind]):
                raise RuntimeError(f"{kind} name already exists")
        cache = self.inspect("volume", CACHE)
        if cache and cache.get("Labels", {}).get("hive.pilot") != PILOT:
            raise RuntimeError("shared cache ownership mismatch")
        if not cache:
            self.call("volume", "create", "--label", f"hive.pilot={PILOT}", "--label",
                      "hive.pilot.batch=shared-cache", "--label", "hive.pilot.owner=shared-cache", CACHE)
        self.call("volume", "create", *self.label_args(), self.data["volume"])
        volume = self.owned("volume")
        self.data["volume_identity"] = volume["CreatedAt"]
        self.data["volume_labels"] = volume["Labels"]
        self.save()
        image = self.data.get("image", self.args.image)
        self.call("create", "--platform", "linux/amd64", "--name", self.data["container"], "--init",
                  "--user", "1000:1000", "--network", "bridge", *self.label_args(),
                  "--tmpfs", "/run/hive-pilot/secrets:rw,noexec,nosuid,nodev,mode=0700,uid=1000,gid=1000",
                  "--mount", f"type=volume,src={self.data['volume']},dst=/work",
                  "--mount", f"type=volume,src={CACHE},dst=/home/pilot/.cache", image)
        container = self.owned("container")
        self.data["image"] = container["Image"]
        self.data["container_identity"] = container["Id"]
        self.data["container_labels"] = container["Config"]["Labels"]
        self.save()
        config, host = container["Config"], container["HostConfig"]
        if (config["User"] != "1000:1000" or host.get("PortBindings") or host.get("Binds")
                or not host.get("Init") or host.get("NetworkMode") != "bridge"
                or any(mount["Type"] != "volume" for mount in container["Mounts"])
                or "/run/hive-pilot/secrets" not in host.get("Tmpfs", {})):
            raise RuntimeError("container isolation inspection failed; resource preserved")
        self.call("start", container["Id"])
        self.call("exec", container["Id"], "mkdir", "-p", "/run/hive-pilot/secrets/real-data",
                  "/run/hive-pilot/secrets/shadow-data")
        print(container["Id"])

    def container_id(self):
        obj = self.owned("container")
        if not obj:
            raise RuntimeError("batch container missing")
        return obj["Id"]

    def upload(self):
        # The same launcher is then invoked on the remote host. No SSH is used
        # by any other subcommand; local Docker validation cannot reach bb1.
        if not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_.@-]*", self.args.remote or ""):
            raise RuntimeError("upload requires a plain SSH host alias")
        if not re.fullmatch(r"/[A-Za-z0-9_./-]+", self.args.remote_dir or ""):
            raise RuntimeError("upload requires an absolute remote directory without shell metacharacters")
        files = [Path(self.args.runner), Path(self.args.bundle)]
        if any(not file.is_file() for file in files):
            raise RuntimeError("runner or input bundle missing")
        destination = self.args.remote_dir.rstrip("/") + "/" + self.args.batch
        command(["ssh", "--", self.args.remote, "mkdir -p -- " + shlex.quote(destination)])
        command(["scp", "--", *(str(file) for file in files), str(self.path),
                 *(str(ROOT / name) for name in ("bb1.sh", "launcher.py", "Dockerfile")),
                 f"{self.args.remote}:{destination}/"], capture=False)
        print(destination)

    def stage(self):
        cid = self.container_id()
        # Reject absolute paths, traversal and links before extraction. Bundle
        # must contain candidate/, baseline/ and fixtures/ at its top level.
        import tarfile
        with tarfile.open(self.args.bundle) as archive:
            for member in archive.getmembers():
                path = Path(member.name)
                if path.is_absolute() or ".." in path.parts or not (member.isfile() or member.isdir()):
                    raise RuntimeError("unsafe input archive")
        for local, target in ((self.args.runner, "/work/input/pilot"), (self.args.bundle, "/work/input/sources.tar")):
            with open(local, "rb") as stream:
                self.call("exec", "-i", cid, "sh", "-c", 'cat > "$1"', "sh", target,
                          capture=False, stdin=stream)
        self.call("exec", cid, "tar", "--no-same-owner", "-xf", "/work/input/sources.tar", "-C", "/work/input")
        baseline = "/work/input/baseline"
        self.call("exec", cid, "test", "-d", baseline + "/.git")
        top = self.call("exec", cid, "git", "-C", baseline, "rev-parse", "--show-toplevel").strip()
        revision = self.call("exec", cid, "git", "-C", baseline, "rev-parse", "HEAD").strip()
        if top != baseline or not revision.startswith("d5164f71"):
            raise RuntimeError("baseline must be a Git checkout at d5164f71")
        dirty = self.call("exec", cid, "git", "-C", baseline, "status", "--porcelain", "--",
                          "content", "integrations/opencode/context-loader.js", "integrations/agent-profiles.json")
        if dirty.strip():
            raise RuntimeError("baseline payload must match its recorded commit")
        self.call("exec", cid, "chmod", "0755", "/work/input/pilot")
        self.call("exec", cid, "/work/input/pilot", "-h", capture=False)

    def login(self):
        # The runner imports the opencode-go integration (the pilot's provider);
        # a Console login alone leaves opencode-go models without a route.
        self.call("exec", "-it", self.container_id(), "opencode", "auth", "login", "--standalone", "opencode-go", capture=False)

    def catalog(self):
        # --standalone keeps no managed background service alive after the check.
        catalog = self.call("exec", self.container_id(), "opencode", "models", "--standalone")
        missing = [model for model in MODELS if f"opencode-go/{model}" not in catalog.splitlines()]
        if missing:
            raise RuntimeError("required catalog models missing: " + ", ".join(missing))
        # OpenCode 2 keeps no models.json cache, but the runner copies one into
        # each shadow home. Snapshot the public models.dev catalog it came from.
        cid = self.container_id()
        snapshot = "/home/pilot/.cache/opencode/models.json"
        self.call("exec", cid, "curl", "-fsSL", "-o", snapshot + ".tmp", "https://models.dev/api.json")
        go = json.loads(self.call("exec", cid, "cat", snapshot + ".tmp")).get("opencode-go", {}).get("models", {})
        absent = [model for model in MODELS if model not in go]
        if absent:
            self.call("exec", cid, "rm", "-f", snapshot + ".tmp")
            raise RuntimeError("models.dev snapshot lacks opencode-go models: " + ", ".join(absent))
        self.call("exec", cid, "mv", snapshot + ".tmp", snapshot)
        self.data["models_snapshot_sha256"] = self.call("exec", cid, "sha256sum", snapshot).split()[0]
        self.data["catalog_verified"] = True
        self.save()
        print("All four opencode-go models present")

    def run(self):
        if not self.data.get("catalog_verified"):
            raise RuntimeError("catalog must pass before running")
        extra = self.args.runner_args
        if extra and extra[0] == "--":
            extra = extra[1:]
        # Own the isolation and evidence destinations; callers supply only the
        # fixture, arm and model. No native configuration or credential flags.
        allowed = {"--case", "--arm", "--model"}
        if len(extra) != 6 or set(extra[::2]) != allowed:
            raise RuntimeError("run requires exactly --case, --arm and --model")
        if not re.fullmatch(r"[a-z0-9][a-z0-9-]{0,79}", self.args.run_id or ""):
            raise RuntimeError("run requires a safe unique --run-id")
        arm = dict(zip(extra[::2], extra[1::2]))["--arm"]
        if arm not in ("A", "B", "C"):
            raise RuntimeError("arm must be A, B or C")
        guidance = "/work/input/baseline" if arm == "A" else "/work/input/candidate"
        self.call("exec", self.container_id(), "/work/input/pilot", "--suite", PILOT,
                  "--host", "opencode", "--memory-store", "disposable-container",
                  "--native-data-root", "/run/hive-pilot/secrets/shadow-data",
                  "--guidance-source", guidance, "--source", "/work/input/candidate",
                  "--timeout", "3m", "--out", f"/work/runs/{self.args.run_id}", *extra, capture=False)

    def collect(self):
        cid = self.container_id()
        self.owned("volume")
        container = self.owned("container")
        if not container["State"]["Running"]:
            # Starting only the inert sleep command also permits inspection of
            # an interrupted, stopped container; it never retries the runner.
            self.call("start", cid)
        self.call("exec", cid, "sh", "-c",
                  'for name in pilot opencode engram; do pgrep -x "$name" >/dev/null && exit 1; done; exit 0')
        self.collect_from(cid)

    def collect_from(self, cid, allow_empty=False):
        expected = self.source_hashes(cid)
        if allow_empty and not expected:
            self.data["collection_receipt"] = {"verified_empty": True, "source_hashes": {}}
            self.save()
            print("Owned work volume checked; no run evidence to collect")
            return
        destination = Path(self.args.evidence or str(self.state / f"{self.args.batch}-evidence")).absolute()
        if destination.exists():
            raise RuntimeError("evidence destination already exists; select a new directory")
        destination.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
        staging = Path(tempfile.mkdtemp(prefix="flash-collect-", dir=destination.parent))
        try:
            self.call("cp", f"{cid}:/work/runs/.", str(staging))
            checksums = {}
            for file in sorted(staging.rglob("*")):
                if file.is_symlink() or (not file.is_file() and not file.is_dir()):
                    raise RuntimeError("unsafe evidence entry")
                if file.is_file():
                    checksums[str(file.relative_to(staging))] = hashlib.sha256(file.read_bytes()).hexdigest()
            # Compare against source hashes, not merely a hash of the copy.
            current = self.source_hashes(cid)
            if checksums != expected or current != expected:
                raise RuntimeError("evidence checksum mismatch")
            (staging / "SHA256SUMS.json").write_text(json.dumps(checksums, indent=2) + "\n")
            os.rename(staging, destination)
            self.data["evidence"] = str(destination)
            self.data["collection_receipt"] = {
                "evidence_directory": str(destination), "source_hashes": expected,
                "sha256sums_hash": hashlib.sha256((destination / "SHA256SUMS.json").read_bytes()).hexdigest(),
            }
            self.save()
            print(destination)
        except BaseException:
            # Keep partial evidence for diagnosis; batch resources remain owned.
            print(f"collection failed; partial copy retained at {staging}", file=sys.stderr)
            raise

    def source_hashes(self, cid):
        # Only generated evidence is copied; source inputs and tmpfs are excluded.
        self.call("exec", cid, "sh", "-c",
                  "if [ -L /work/runs ]; then exit 1; fi; "
                  "if [ ! -e /work/runs ]; then exit 0; fi; "
                  "test -d /work/runs || exit 1; "
                  "entries=$(find /work/runs -type l -o -iname 'opencode.db*' -o -iname '*credential*' "
                  "-o -iname 'auth*' -o -name '.env' -o -iname '*.sqlite*') || exit 1; test -z \"$entries\"")
        source = self.call("exec", cid, "sh", "-c",
                           "if [ -d /work/runs ]; then cd /work/runs && find . -type f -exec sha256sum {} +; fi")
        hashes = {}
        for line in source.splitlines():
            digest, name = line.split("  ", 1)
            hashes[name.removeprefix("./")] = digest
        return hashes

    def verify_receipt(self, cid):
        current = self.source_hashes(cid)
        receipt = self.data.get("collection_receipt")
        if not receipt:
            if current:
                raise RuntimeError("collection receipt required; resources preserved")
            self.data["collection_receipt"] = {"verified_empty": True, "source_hashes": {}}
            self.save()
            print("Work evidence verified empty before destruction")
            return
        if receipt.get("source_hashes") != current:
            raise RuntimeError("evidence changed since collection; resources preserved")
        if receipt.get("verified_empty") and not current:
            return
        destination = Path(receipt["evidence_directory"])
        checksum_file = destination / "SHA256SUMS.json"
        if destination.is_symlink() or checksum_file.is_symlink():
            raise RuntimeError("unsafe collection receipt")
        if hashlib.sha256(checksum_file.read_bytes()).hexdigest() != receipt.get("sha256sums_hash"):
            raise RuntimeError("collection checksum receipt changed")
        if json.loads(checksum_file.read_text()) != current:
            raise RuntimeError("collection hash list mismatch")
        copied = {}
        for file in destination.rglob("*"):
            if file.is_symlink() or (not file.is_file() and not file.is_dir()):
                raise RuntimeError("unsafe collected evidence")
            if file.is_file() and file != checksum_file:
                copied[str(file.relative_to(destination))] = hashlib.sha256(file.read_bytes()).hexdigest()
        if copied != current:
            raise RuntimeError("collected evidence changed; resources preserved")

    def destroy(self):
        # Validate ALL resources before deleting ANY, including partial manifests.
        container = self.owned("container")
        volume = self.owned("volume")
        if container:
            cid = container["Id"]
            if not container["State"]["Running"]:
                self.call("start", cid)
                container = self.owned("container")
            self.call("exec", cid, "sh", "-c",
                      'for name in pilot opencode engram; do pgrep -x "$name" >/dev/null && exit 1; done; exit 0')
            self.verify_receipt(cid)
        elif volume:
            self.recover_volume(volume, verify_only=True)
        if container:
            cid = container["Id"]
            if container["State"]["Running"]:
                self.call("exec", cid, "sh", "-c", 'for name in pilot opencode engram; do pgrep -x "$name" >/dev/null && exit 1; done; exit 0')
                help_text = self.call("exec", cid, "opencode", "auth", "--help")
                if "logout" in help_text:
                    interactive = sys.stdin.isatty() and sys.stdout.isatty()
                    logout_args = ["exec"] + (["-it"] if interactive else [])
                    logout_args += [cid, "opencode", "auth", "logout", "--standalone", "opencode"]
                    logout = subprocess.run(self.docker + logout_args,
                                            stdout=None if interactive else subprocess.DEVNULL,
                                            stderr=None if interactive else subprocess.DEVNULL)
                    if logout.returncode:
                        print("Local logout did not complete; container deletion discards its tmpfs")
            self.call("container", "rm", "--force", cid)
            if self.inspect("container", cid):
                raise RuntimeError("container deletion not verified")
        if volume:
            # Docker volume identity is its name; CreatedAt was checked above.
            self.call("volume", "rm", volume["Name"])
            if self.inspect("volume", volume["Name"]):
                raise RuntimeError("volume deletion not verified")
        self.data["destroyed"] = True
        self.save()
        print("Batch destroyed; shared cache preserved")

    def recover_volume(self, volume, verify_only=False):
        image = self.data.get("image", self.args.image)
        image_info = json.loads(self.call("image", "inspect", image))[0]
        if image_info["Config"].get("Labels", {}).get("hive.pilot") != PILOT:
            raise RuntimeError("recovery image is not a pilot image")
        helper = self.call("create", "--platform", "linux/amd64", "--init",
                           "--name", f"{self.data['container']}-recover-{secrets.token_hex(8)}",
                           "--user", "1000:1000", "--network", "none", "--read-only",
                           *self.label_args(), "--mount",
                           f"type=volume,src={volume['Name']},dst=/work,readonly,volume-nocopy",
                           image_info["Id"], "sleep", "infinity").strip()
        try:
            inspected = self.inspect("container", helper)
            host = inspected["HostConfig"]
            mounts = inspected["Mounts"]
            if (inspected["Config"]["User"] != "1000:1000" or host.get("PortBindings")
                    or host.get("Binds") or len(mounts) != 1
                    or mounts[0]["Type"] != "volume" or mounts[0]["Name"] != volume["Name"]
                    or mounts[0]["Destination"] != "/work" or mounts[0]["RW"]):
                raise RuntimeError("recovery helper isolation inspection failed")
            self.call("start", helper)
            if verify_only:
                self.verify_receipt(helper)
            else:
                self.collect_from(helper, allow_empty=True)
        finally:
            inspected = self.inspect("container", helper)
            if inspected:
                labels = inspected["Config"].get("Labels", {})
                if any(labels.get(key) != value for key, value in self.labels().items()):
                    raise RuntimeError("recovery helper ownership mismatch")
                self.call("container", "rm", "--force", helper)
                if self.inspect("container", helper):
                    raise RuntimeError("recovery helper deletion not verified")

    def recover(self):
        if self.owned("container") is None:
            volume = self.owned("volume")
            if volume is None:
                print("Batch container and work volume absent; no evidence to collect")
            else:
                self.recover_volume(volume)
        else:
            self.collect()
        self.destroy()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--state-dir", required=True, help="local manifest directory (not a credential store)")
    parser.add_argument("--context", help="Docker context; use desktop-linux for Mac validation")
    parser.add_argument("--image", default="hive-pilot-test")
    sub = parser.add_subparsers(dest="action", required=True)
    for action in ("prepare", "build", "create", "upload", "stage", "login", "catalog", "run", "collect", "destroy", "recover"):
        item = sub.add_parser(action)
        item.add_argument("batch", help="batch identifier, lowercase letters/digits/hyphens")
        if action in ("upload", "stage"):
            item.add_argument("--runner", required=True)
            item.add_argument("--bundle", required=True)
        if action == "upload":
            item.add_argument("--remote", required=True)
            item.add_argument("--remote-dir", required=True)
        if action in ("collect", "recover"):
            item.add_argument("--evidence")
        if action == "run":
            item.add_argument("--run-id", required=True)
    arguments = sys.argv[1:]
    runner_arguments = []
    if "--" in arguments:
        separator = arguments.index("--")
        runner_arguments = arguments[separator + 1:]
        arguments = arguments[:separator]
    args = parser.parse_args(arguments)
    if runner_arguments and args.action != "run":
        parser.error("trailing runner flags are valid only for run")
    args.runner_args = runner_arguments
    if not re.fullmatch(r"[a-z0-9][a-z0-9-]{0,63}", args.batch):
        parser.error("invalid batch identifier")
    launcher = Launcher(args)
    if args.action != "prepare":
        launcher.load()
    getattr(launcher, args.action)()


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, OSError, ValueError, KeyError) as error:
        print(f"error: {error}", file=sys.stderr)
        sys.exit(1)
