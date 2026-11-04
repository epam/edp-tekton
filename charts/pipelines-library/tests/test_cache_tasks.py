import os
import shutil
import subprocess
import uuid

import pytest

from .helpers import helm_template

CONFIG = """
global:
  gitProviders:
    - github
"""

# Serves files from $FAKE_CACHE_SERVER: HEAD and GET by URL basename, POST -F path=... -F file=@... stores the file.
FAKE_CURL = """#!/usr/bin/env bash
out="" path="" file="" head=0 url=""
while [ $# -gt 0 ]; do
  case "$1" in
    -o) out="$2"; shift ;;
    -o*) out="${1#-o}" ;;
    -fsI|-I) head=1 ;;
    -F) case "$2" in path=*) path="${2#path=}" ;; file=@*) file="${2#file=@}" ;; esac; shift ;;
    -X) shift ;;
    http*) url="$1" ;;
  esac
  shift
done
if [ -n "$path" ]; then
  [ -f "$file" ] && cp "$file" "$FAKE_CACHE_SERVER/$path"
  exit 0
fi
name="$FAKE_CACHE_SERVER/${url##*/}"
[ -f "$name" ] || exit 22
[ "$head" = 1 ] && exit 0
cp "$name" "${out:-./${url##*/}}"
"""


def _gnu_tar_with_zstd():
    if shutil.which("zstd") is None:
        return False
    version = subprocess.run(["tar", "--version"], capture_output=True, text=True)
    return version.stdout.startswith("tar (GNU tar)")


pytestmark = pytest.mark.skipif(
    not _gnu_tar_with_zstd(), reason="cache task scripts need GNU tar and zstd"
)


@pytest.fixture(scope="module")
def scripts():
    tasks = helm_template(CONFIG)["task"]
    return {
        name: tasks[name]["spec"]["steps"][0]["script"]
        for name in ("get-cache", "save-cache")
    }


@pytest.fixture
def env(tmp_path):
    bin_dir = tmp_path / "bin"
    bin_dir.mkdir()
    curl = bin_dir / "curl"
    curl.write_text(FAKE_CURL)
    curl.chmod(0o755)

    server = tmp_path / "server"
    server.mkdir()

    # save-cache writes its archive to /tmp/<CACHE_NAME>.tar.zst; a unique name keeps parallel runs apart.
    name = f"cache-test-{uuid.uuid4().hex}"
    yield {
        **os.environ,
        "PATH": f"{bin_dir}:{os.environ['PATH']}",
        "FAKE_CACHE_SERVER": str(server),
        "CACHE_SERVER_URL": "http://tekton-cache:8080",
        "CACHE_NAME": name,
        "CACHE_MAX_AGE_DAYS": "0",
    }
    if os.path.exists(f"/tmp/{name}.tar.zst"):
        os.remove(f"/tmp/{name}.tar.zst")


def _run(script, workspace, env):
    subprocess.run(["bash", "-c", script], cwd=workspace, env=env, check=True,
                   capture_output=True)


def _pack(source, archive):
    subprocess.run(["tar", "c", "-I", "zstd", "-f", str(archive), "."], cwd=source,
                   check=True)


def _members(archive):
    listing = subprocess.run(["tar", "-t", "--zstd", "-f", str(archive)],
                             capture_output=True, text=True, check=True)
    return {line.rstrip("/") for line in listing.stdout.splitlines()}


def test_get_cache_leaves_no_archive_in_the_workspace(scripts, env, tmp_path):
    name = env["CACHE_NAME"]
    content = tmp_path / "content"
    (content / "github.com").mkdir(parents=True)
    (content / "github.com" / "module.go").write_text("package module\n")
    (content / f"{name}.tar.zst").write_bytes(b"archive packed by an earlier save-cache")
    _pack(content, tmp_path / "server" / f"{name}.tar.zst")

    workspace = tmp_path / "workspace"
    workspace.mkdir()
    _run(scripts["get-cache"], workspace, env)

    assert (workspace / "github.com" / "module.go").read_text() == "package module\n"
    assert not list(workspace.rglob("*.tar.zst")), (
        "get-cache must leave no cache archive in the workspace, save-cache packs the "
        "whole workspace into the next archive"
    )


def test_save_cache_does_not_pack_a_cache_archive(scripts, env, tmp_path):
    name = env["CACHE_NAME"]
    workspace = tmp_path / "workspace"
    (workspace / "github.com").mkdir(parents=True)
    (workspace / "github.com" / "module.go").write_text("package module\n")
    (workspace / f"{name}.tar.zst").write_bytes(b"stray cache archive")

    _run(scripts["save-cache"], workspace, env)

    members = _members(tmp_path / "server" / f"{name}.tar.zst")
    assert "./github.com/module.go" in members
    assert f"./{name}.tar.zst" not in members, (
        "save-cache must not pack the cache archive into itself"
    )
