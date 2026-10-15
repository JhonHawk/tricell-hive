#!/usr/bin/env python3
"""Explicit local Docker checks for T8; no models, login or SSH."""
import importlib.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

sys.dont_write_bytecode = True
ROOT = Path(__file__).resolve().parent
DOCKER = shutil.which("docker")
PREFIX = "flash-t8-" + os.urandom(4).hex()
STATE = Path(tempfile.mkdtemp(prefix="flash-t8-verify-", dir=os.environ["TMPDIR"]))
RESOURCES = []
CACHE_CREATED = False


def docker(*args, check=True):
    return subprocess.run([DOCKER, "--context", "desktop-linux", *args], capture_output=True,
                          text=True, check=check)


def launch(action, batch, *args, success=True, env=None):
    result = subprocess.run([str(ROOT / "bb1.sh"), "--state-dir", str(STATE), "--context",
                             "desktop-linux", action, batch, *args], capture_output=True, text=True, env=env)
    assert (result.returncode == 0) == success, (action, result.returncode, result.stderr)
    print(f"{action} {batch}: {'passed' if success else 'refused as expected'}")
    return result


def manifest(batch):
    return STATE / (batch + ".json")


def create_batch(suffix):
    batch = PREFIX + "-" + suffix
    launch("prepare", batch)
    launch("create", batch)
    container = json.loads(docker("container", "inspect", "hive-pilot-" + batch).stdout)[0]
    volume = json.loads(docker("volume", "inspect", "hive-pilot-work-" + batch).stdout)[0]
    RESOURCES.extend((("container", container["Id"]), ("volume", volume["Name"])))
    return batch, container, volume


def write_dummy(cid):
    script = '#!/bin/sh\nmkdir -p /work/runs/dummy\nprintf "dummy evidence\\n" > /work/runs/dummy/evidence.txt\n'
    subprocess.run([DOCKER, "--context", "desktop-linux", "exec", "-i", cid, "sh", "-c",
                    'cat > /work/input/fake-runner; chmod +x /work/input/fake-runner'],
                   input=script, text=True, check=True)
    docker("exec", cid, "/work/input/fake-runner")


def main():
    global CACHE_CREATED
    CACHE_CREATED = docker("volume", "inspect", "hive-pilot-cache", check=False).returncode != 0
    for phase in ("nothing-created", "volume-only"):
        batch = PREFIX + "-" + phase
        launch("prepare", batch)
        data = json.loads(manifest(batch).read_text())
        if phase == "volume-only":
            labels = [arg for key, value in data["labels"].items()
                      for arg in ("--label", f"{key}={value}")]
            name = docker("volume", "create", *labels, data["volume"]).stdout.strip()
            RESOURCES.append(("volume", name))
        result = launch("recover", batch)
        assert ("no run evidence to collect" if phase == "volume-only" else "no evidence to collect") in result.stdout
        assert docker("container", "inspect", data["container"], check=False).returncode != 0
        assert docker("volume", "inspect", data["volume"], check=False).returncode != 0
        assert json.loads(manifest(batch).read_text())["destroyed"]
        assert not (STATE / (batch + "-evidence")).exists()
        print(f"recover: {phase} reports no evidence and removes only owned resources")
    batch, container, volume = create_batch("partial")
    cid = container["Id"]
    host = container["HostConfig"]
    assert container["Config"]["User"] == "1000:1000"
    assert not host["PortBindings"] and not host["Binds"] and host["Init"]
    assert all(m["Type"] == "volume" and m["Destination"] != "/var/run/docker.sock" for m in container["Mounts"])
    assert "mode=0700" in host["Tmpfs"]["/run/hive-pilot/secrets"]
    assert "uid=1000" in host["Tmpfs"]["/run/hive-pilot/secrets"]
    assert container["Config"]["Labels"]["hive.pilot"] == "flash-class-guidance"
    assert container["Config"]["Labels"]["hive.pilot.batch"] == batch
    print("inspect: no ports/binds/socket, UID 1000, init, bridge, owned volumes and 0700 tmpfs")
    # Name collision against Docker, without a pre-existing manifest.
    other = STATE / "other"
    result = subprocess.run([str(ROOT / "bb1.sh"), "--state-dir", str(other), "--context",
                             "desktop-linux", "prepare", batch], capture_output=True)
    assert result.returncode and not (other / (batch + ".json")).exists()
    print("prepare: existing Docker name refused before writing manifest")
    original = json.loads(manifest(batch).read_text())
    changed = dict(original, owner="0" * 64)
    changed["labels"] = dict(original["labels"], **{"hive.pilot.owner": "0" * 64})
    manifest(batch).write_text(json.dumps(changed))
    launch("destroy", batch, success=False)
    changed = dict(original, container_identity="0" * 64)
    manifest(batch).write_text(json.dumps(changed))
    launch("destroy", batch, success=False)
    assert docker("container", "inspect", cid, check=False).returncode == 0
    print("destroy: token and ID mismatches preserve both resources")
    partial = dict(original)
    del partial["container_identity"]
    del partial["container_labels"]
    manifest(batch).write_text(json.dumps(partial))
    write_dummy(cid)
    # Labeled controls belong to this test but not the recovered batch.
    control = PREFIX + "-control"
    control_labels = ["--label", "hive.pilot=flash-class-guidance", "--label",
                      "hive.pilot.batch=" + control, "--label", "hive.pilot.owner=" + control]
    control_id = docker("create", "--platform", "linux/amd64", "--name", "hive-pilot-" + control,
                        *control_labels, "hive-pilot-test").stdout.strip()
    RESOURCES.append(("container", control_id))
    control_volume = docker("volume", "create", *control_labels, "hive-pilot-work-" + control).stdout.strip()
    RESOURCES.append(("volume", control_volume))
    cache_before = json.loads(docker("volume", "inspect", "hive-pilot-cache").stdout)[0]
    launch("recover", batch)
    assert docker("container", "inspect", cid, check=False).returncode != 0
    assert docker("volume", "inspect", volume["Name"], check=False).returncode != 0
    evidence = STATE / (batch + "-evidence")
    assert json.loads((evidence / "SHA256SUMS.json").read_text())["dummy/evidence.txt"]
    assert docker("container", "inspect", control_id, check=False).returncode == 0
    assert docker("volume", "inspect", control_volume, check=False).returncode == 0
    assert json.loads(docker("volume", "inspect", "hive-pilot-cache").stdout)[0] == cache_before
    print("recover: partial manifest collected/checksummed; controls and shared cache unchanged")
    batch, container, volume = create_batch("copy-fail")
    write_dummy(container["Id"])
    wrapper = STATE / "bin"
    wrapper.mkdir()
    (wrapper / "docker").write_text('#!/bin/sh\nfor arg do\n  if [ "$arg" = cp ]; then exit 79; fi\ndone\nexec ' + DOCKER + ' "$@"\n')
    (wrapper / "docker").chmod(0o755)
    env = dict(os.environ, PATH=str(wrapper) + os.pathsep + os.environ["PATH"])
    launch("recover", batch, success=False, env=env)
    assert docker("container", "inspect", container["Id"], check=False).returncode == 0
    assert docker("volume", "inspect", volume["Name"], check=False).returncode == 0
    print("recover: injected docker cp failure preserves container and volume")
    launch("recover", batch)
    for phase in ("orphan-evidence", "orphan-copy-fail"):
        batch, container, volume = create_batch(phase)
        write_dummy(container["Id"])
        data = json.loads(manifest(batch).read_text())
        assert container["Config"]["Labels"] == data["container_labels"]
        assert container["Id"] == data["container_identity"]
        docker("container", "rm", "--force", container["Id"])
        if phase == "orphan-copy-fail":
            launch("recover", batch, success=False, env=env)
            assert docker("volume", "inspect", volume["Name"], check=False).returncode == 0
            remaining = docker("container", "ls", "-a", "--filter", "label=hive.pilot.batch=" + batch,
                               "--format", "{{.ID}}").stdout.strip()
            assert not remaining
            print("recover: orphan copy failure preserves volume and removes read-only helper")
        launch("recover", batch)
        evidence = STATE / (batch + "-evidence")
        import hashlib
        copied = evidence / "dummy/evidence.txt"
        assert copied.read_text() == "dummy evidence\n"
        assert json.loads((evidence / "SHA256SUMS.json").read_text())["dummy/evidence.txt"] == hashlib.sha256(copied.read_bytes()).hexdigest()
        assert docker("volume", "inspect", volume["Name"], check=False).returncode != 0
        assert not docker("container", "ls", "-a", "--filter", "label=hive.pilot.batch=" + batch,
                          "--format", "{{.ID}}").stdout.strip()
        print("recover: orphan evidence collected with verified checksums before volume destruction")
    batch, container, volume = create_batch("orphan-credential")
    write_dummy(container["Id"])
    docker("exec", container["Id"], "touch", "/work/runs/auth.json")
    data = json.loads(manifest(batch).read_text())
    assert container["Config"]["Labels"] == data["container_labels"]
    assert container["Id"] == data["container_identity"]
    docker("container", "rm", "--force", container["Id"])
    launch("recover", batch, success=False)
    assert docker("volume", "inspect", volume["Name"], check=False).returncode == 0
    assert not (STATE / (batch + "-evidence")).exists()
    assert not docker("container", "ls", "-a", "--filter", "label=hive.pilot.batch=" + batch,
                      "--format", "{{.ID}}").stdout.strip()
    print("recover: orphan credential filename refused; volume preserved and helper removed")
    batch, container, volume = create_batch("receipt")
    cid = container["Id"]
    write_dummy(cid)
    docker("exec", cid, "touch", "/work/input/credential_process.go")
    launch("destroy", batch, success=False)
    assert docker("volume", "inspect", volume["Name"], check=False).returncode == 0
    launch("collect", batch)
    receipt = json.loads(manifest(batch).read_text())["collection_receipt"]
    assert receipt["evidence_directory"] and receipt["sha256sums_hash"] and receipt["source_hashes"]
    docker("exec", cid, "sh", "-c", "printf changed >> /work/runs/dummy/evidence.txt")
    launch("destroy", batch, success=False)
    assert docker("container", "inspect", cid, check=False).returncode == 0
    launch("collect", batch, "--evidence", str(STATE / "receipt-updated"))
    docker("container", "rm", "--force", cid)
    launch("destroy", batch)
    assert docker("volume", "inspect", volume["Name"], check=False).returncode != 0
    print("destroy: missing/stale receipts refused; matching receipt verified through orphan helper")
    batch, container, volume = create_batch("receipt-present")
    write_dummy(container["Id"])
    launch("collect", batch)
    launch("destroy", batch)
    print("destroy: matching collection receipt permits deletion with container present")
    batch, container, volume = create_batch("empty-destroy")
    launch("destroy", batch)
    assert json.loads(manifest(batch).read_text())["collection_receipt"]["verified_empty"]
    print("destroy: verified-empty batch permits deletion")
    batch, container, volume = create_batch("arms-stage")
    cid = container["Id"]
    launch("upload", batch, "--remote=-bad", "--remote-dir", "/tmp/pilot",
           "--runner", "/nonexistent", "--bundle", "/nonexistent", success=False)
    print("upload: leading-hyphen remote rejected without SSH")
    runner = STATE / "fake-pilot"
    runner.write_text('#!/bin/sh\nwhile [ "$#" -gt 0 ]; do\n  case "$1" in\n    --guidance-source) shift; guidance=$1;;\n    --out) shift; output=$1;;\n  esac\n  shift\ndone\nmkdir -p "$output"\nprintf "%s\\n" "$guidance" > "$output/guidance.txt"\n')
    with runner.open("rb") as stream:
        subprocess.run([DOCKER, "--context", "desktop-linux", "exec", "-i", cid, "sh", "-c",
                        'cat > /work/input/pilot; chmod +x /work/input/pilot'], stdin=stream, check=True)
    data = json.loads(manifest(batch).read_text())
    data["catalog_verified"] = True  # Fake runner only; never a catalog/model claim.
    manifest(batch).write_text(json.dumps(data))
    for arm in ("A", "B", "C"):
        launch("run", batch, "--run-id", "arm-" + arm.lower(), "--", "--case", "dummy",
               "--arm", arm, "--model", "dummy")
        recorded = docker("exec", cid, "cat", "/work/runs/arm-" + arm.lower() + "/guidance.txt").stdout.strip()
        assert recorded == ("/work/input/baseline" if arm == "A" else "/work/input/candidate")
    print("run: fake runner observed baseline for A and candidate for B/C")
    import tarfile
    import io
    bundle = STATE / "missing-git.tar"
    with tarfile.open(bundle, "w") as archive:
        entry = tarfile.TarInfo("baseline/dummy.txt")
        entry.size = 5
        archive.addfile(entry, io.BytesIO(b"dummy"))
    launch("stage", batch, "--runner", str(runner), "--bundle", str(bundle), success=False)
    docker("exec", cid, "git", "init", "/work/input/baseline")
    docker("exec", cid, "git", "-C", "/work/input/baseline", "add", "dummy.txt")
    docker("exec", cid, "git", "-C", "/work/input/baseline", "-c", "user.name=Pilot Test",
           "-c", "user.email=pilot@example.invalid", "commit", "-m", "Dummy baseline")
    wrong_bundle = STATE / "wrong-git.tar"
    with wrong_bundle.open("wb") as output:
        subprocess.run([DOCKER, "--context", "desktop-linux", "exec", cid, "tar", "-cf", "-",
                        "-C", "/work/input", "baseline"], stdout=output, check=True)
    launch("stage", batch, "--runner", str(runner), "--bundle", str(wrong_bundle), success=False)
    print("stage: missing .git and a different baseline commit refused")
    launch("recover", batch)
    # Independently exercise a batch label mismatch with the right ownership token.
    batch = PREFIX + "-label"
    launch("prepare", batch)
    data = json.loads(manifest(batch).read_text())
    bad = docker("create", "--platform", "linux/amd64", "--name", data["container"],
                 "--label", "hive.pilot=flash-class-guidance", "--label", "hive.pilot.batch=wrong-batch",
                 "--label", "hive.pilot.owner=" + data["owner"], "hive-pilot-test").stdout.strip()
    RESOURCES.append(("container", bad))
    launch("destroy", batch, success=False)
    assert docker("container", "inspect", bad, check=False).returncode == 0
    print("destroy: different batch label refused")
    spec = importlib.util.spec_from_file_location("count_projects", ROOT / "count-projects.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    assert module.parse_projects("No projects found.\n") == []
    assert module.parse_projects("Projects (2):\n  flash-class-example   1 obs  1 session  0 prompts\n  other  2 obs  3 sessions  1 prompt\n") == ["flash-class-example", "other"]
    print("count parser: empty and table formats; prefix and exact IDs remain distinct")
    print("scratch evidence and manifests: " + str(STATE))


if __name__ == "__main__":
    try:
        main()
    finally:
        # Only IDs/names recorded by this invocation; never a label-wide purge.
        for kind, identity in sorted(RESOURCES, key=lambda item: item[0] != "container"):
            inspected = docker(kind, "inspect", identity, check=False)
            if inspected.returncode:
                continue
            obj = json.loads(inspected.stdout)[0]
            labels = obj["Config"]["Labels"] if kind == "container" else obj["Labels"]
            if labels.get("hive.pilot") != "flash-class-guidance":
                raise RuntimeError("cleanup ownership changed")
            args = ("--force", identity) if kind == "container" else (identity,)
            docker(kind, "rm", *args)
            assert docker(kind, "inspect", identity, check=False).returncode != 0
        if CACHE_CREATED and docker("volume", "inspect", "hive-pilot-cache", check=False).returncode == 0:
            docker("volume", "rm", "hive-pilot-cache")
        print("test resources removed by recorded identity; no test process left running")
