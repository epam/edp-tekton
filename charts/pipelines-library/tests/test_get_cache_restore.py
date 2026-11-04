import contextlib
import functools
import http.server
import os
import shutil
import subprocess
import threading

import pytest

from .helpers import helm_template

CONFIG = """
global:
  gitProviders:
    - github
"""

CACHE_NAME = "gitfusion"


def _unavailable_tools():
    missing = [
        tool for tool in ("bash", "curl", "tar", "zstd") if not shutil.which(tool)
    ]
    if missing:
        return f"the get-cache script needs {', '.join(missing)}"
    version = subprocess.run(["tar", "--version"], capture_output=True, text=True)
    if "GNU tar" not in version.stdout:
        return "the get-cache script is written for GNU tar, as in the Task image"
    return None


UNAVAILABLE = _unavailable_tools()
requires_tools = pytest.mark.skipif(
    UNAVAILABLE is not None, reason=UNAVAILABLE or "tools are available"
)


class _QuietHandler(http.server.SimpleHTTPRequestHandler):
    def log_message(self, *args):
        pass


class _FailingDownloadHandler(_QuietHandler):
    """HEAD finds the archive, GET answers 500."""

    def do_GET(self):
        self.send_error(500)


@contextlib.contextmanager
def _serve(directory, handler=_QuietHandler):
    server = http.server.ThreadingHTTPServer(
        ("127.0.0.1", 0), functools.partial(handler, directory=str(directory))
    )
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    try:
        yield f"http://127.0.0.1:{server.server_address[1]}"
    finally:
        server.shutdown()
        server.server_close()
        thread.join()


@pytest.fixture
def served(tmp_path):
    directory = tmp_path / "served"
    directory.mkdir()
    return directory


def _pack(source, archive):
    subprocess.run(
        ["tar", "c", "-I", "zstd -T1 -1", "-f", str(archive), "-C", str(source), "."],
        check=True,
    )


def _files(root):
    return sorted(
        str(path.relative_to(root)) for path in root.rglob("*") if path.is_file()
    )


def _run_get_cache(tmp_path, url):
    """Run the rendered get-cache script in an empty cache directory."""
    script = tmp_path / "get-cache.sh"
    script.write_text(
        helm_template(CONFIG)["task"]["get-cache"]["spec"]["steps"][0]["script"]
    )
    workspace = tmp_path / "workspace"
    workspace.mkdir()

    result = subprocess.run(
        ["bash", str(script)],
        cwd=workspace,
        env={**os.environ, "CACHE_SERVER_URL": url, "CACHE_NAME": CACHE_NAME},
        capture_output=True,
        text=True,
    )
    return workspace, result


@requires_tools
@pytest.mark.parametrize(
    "holds_a_copy_of_itself",
    [
        pytest.param(False, id="clean"),
        pytest.param(True, id="saved-by-an-earlier-version"),
    ],
)
def test_get_cache_restores_only_cache_data(tmp_path, served, holds_a_copy_of_itself):
    """save-cache packs the whole cache directory, so whatever get-cache leaves
    there goes into the next archive. The restored directory holds the cached data
    and nothing else, also when the archive on the cache server carries a copy of
    itself, as the archives saved by earlier versions do."""
    cached = tmp_path / "cached"
    (cached / "go-build" / "00").mkdir(parents=True)
    (cached / "go-build" / "00" / "object").write_bytes(os.urandom(64 * 1024))
    (cached / "mod" / "cache" / "download").mkdir(parents=True)
    (cached / "mod" / "cache" / "download" / "module.zip").write_bytes(os.urandom(4096))
    # a file of the cached data that happens to share the name of the archive
    (cached / "mod" / "testdata").mkdir()
    (cached / "mod" / "testdata" / f"{CACHE_NAME}.tar.zst").write_bytes(b"cached data")

    archive = served / f"{CACHE_NAME}.tar.zst"
    _pack(cached, archive)
    if holds_a_copy_of_itself:
        copy = cached / f"{CACHE_NAME}.tar.zst"
        copy.write_bytes(archive.read_bytes())
        _pack(cached, archive)
        copy.unlink()

    with _serve(served) as url:
        workspace, result = _run_get_cache(tmp_path, url)

    assert result.returncode == 0, result.stderr
    assert _files(workspace) == _files(cached), (
        "get-cache must restore exactly the cached data into the cache directory: "
        "an archive left there is packed into the next one by save-cache"
    )
    for name in _files(cached):
        assert (workspace / name).read_bytes() == (cached / name).read_bytes(), name


@requires_tools
def test_get_cache_without_an_archive_leaves_the_cache_directory_empty(
    tmp_path, served
):
    with _serve(served) as url:
        workspace, result = _run_get_cache(tmp_path, url)

    assert result.returncode == 0, result.stderr
    assert _files(workspace) == []


@requires_tools
def test_get_cache_fails_and_leaves_nothing_when_the_download_fails(tmp_path, served):
    cached = tmp_path / "cached"
    cached.mkdir()
    (cached / "object").write_bytes(os.urandom(4096))
    _pack(cached, served / f"{CACHE_NAME}.tar.zst")

    with _serve(served, _FailingDownloadHandler) as url:
        workspace, result = _run_get_cache(tmp_path, url)

    assert result.returncode != 0
    assert (
        _files(workspace) == []
    ), "a failed download must not leave a file in the cache directory"
