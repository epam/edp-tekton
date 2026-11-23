import json
import os
import re
import shutil
import stat
import subprocess
import sys
from collections import defaultdict

import pytest

from .helpers import helm_template

BASE = """
global:
  dnsWildCard: "example.com"
"""

CODEBASE = "shop"
DD_URL = "http://defectdojo.example"
REIMPORT_URL = f"{DD_URL}/api/v2/reimport-scan/"

HEX = "8fb099199b9f2d70342674bd9dbccd3ed03a258f26bbd1d556822c6dfc60c317"
DIGEST = f"sha256:{HEX}"
SHORT_DIGEST = "sha256:8fb099199b9f"

# (image reference, engagement, version); a version of None is no version field.
EXAMPLES = [
    pytest.param(
        "reg.example.com/shop/app:2.4.1",
        "image:reg.example.com/shop/app",
        "2.4.1",
        id="tag",
    ),
    pytest.param("nginx:1.27", "image:nginx", "1.27", id="docker-hub-short-name"),
    pytest.param(
        f"reg.example.com/shop/app@{DIGEST}",
        "image:reg.example.com/shop/app",
        SHORT_DIGEST,
        id="digest",
    ),
    pytest.param(
        f"reg.example.com/shop/app@{HEX}",
        "image:reg.example.com/shop/app",
        SHORT_DIGEST,
        id="digest-without-algorithm",
    ),
    pytest.param(
        f"reg.example.com/shop/app:2.4.1@{DIGEST}",
        "image:reg.example.com/shop/app",
        "2.4.1",
        id="tag-and-digest",
    ),
    pytest.param(
        "reg.example.com:5000/shop/app",
        "image:reg.example.com:5000/shop/app",
        None,
        id="port-without-tag",
    ),
    pytest.param(
        "reg.example.com:5000/shop/app:1.0",
        "image:reg.example.com:5000/shop/app",
        "1.0",
        id="port-with-tag",
    ),
    pytest.param(
        "docker.io/library/nginx:1.27", "image:nginx", "1.27", id="docker-io-library"
    ),
    pytest.param(
        "docker.io/bitnami/redis:7.4",
        "image:bitnami/redis",
        "7.4",
        id="docker-io-namespace",
    ),
    pytest.param(
        "library/nginx:1.27", "image:nginx", "1.27", id="library-without-host"
    ),
    pytest.param(
        "Reg.example.com/Shop/App:1.0",
        "image:reg.example.com/shop/app",
        "1.0",
        id="upper-case",
    ),
    pytest.param(
        "index.docker.io/library/postgres", "image:postgres", None, id="index-docker-io"
    ),
    pytest.param(
        "registry-1.docker.io/valkey/valkey:9",
        "image:valkey/valkey",
        "9",
        id="registry-1-docker-io",
    ),
    pytest.param(
        "Docker.IO/Library/Nginx:1.27", "image:nginx", "1.27", id="upper-case-hub"
    ),
    pytest.param(
        "quay.io/library/app:1", "image:quay.io/library/app", "1", id="library-on-quay"
    ),
    pytest.param(
        "localhost/library/app:1",
        "image:localhost/library/app",
        "1",
        id="library-on-localhost",
    ),
    pytest.param(
        "localhost:5000/library/app",
        "image:localhost:5000/library/app",
        None,
        id="library-on-localhost-port",
    ),
    pytest.param(
        "reg.example.com/shop/app:Rel-1.0_RC",
        "image:reg.example.com/shop/app",
        "Rel-1.0_RC",
        id="tag-keeps-case",
    ),
    pytest.param("shop/app", "image:shop/app", None, id="no-tag-no-digest"),
]

STUB = """
import json
import os
import sys

name = os.path.basename(sys.argv[0])
args = sys.argv[1:]
with open(os.environ["STUB_CALLS"], "a") as calls:
    calls.write(json.dumps([name, *args]) + "\\n")
if name in ("trivy", "curl") and "-o" in args:
    with open(args[args.index("-o") + 1], "w") as out:
        out.write("{}")
if name == "grype":
    print("{}")
elif name == "curl" and "-w" in args:
    print(200, end="")
elif name == "jq":
    print(1)
"""

SHELLS = [
    pytest.param(
        ["dash"],
        id="dash",
        marks=pytest.mark.skipif(not shutil.which("dash"), reason="dash is missing"),
    ),
    pytest.param(
        ["busybox", "sh"],
        id="busybox",
        marks=pytest.mark.skipif(
            not shutil.which("busybox"),
            reason="busybox, the shell of the step images, is missing",
        ),
    ),
]

TASKS = ["image-scan-remote", "image-scan-chart"]


@pytest.fixture(scope="module")
def tasks():
    return helm_template(BASE)["task"]


class Pipeline:
    """Runs the steps of a rendered Task one after another in a shared workspace,
    with the scanners and DefectDojo replaced by stubs that record their calls."""

    def __init__(self, tmp_path, task, shell):
        self.task = task
        self.shell = shell
        self.root = tmp_path
        self.workdir = tmp_path / "source"
        self.workdir.mkdir()
        self.bin = tmp_path / "bin"
        self.bin.mkdir()
        self.calls_file = tmp_path / "calls.jsonl"
        self.calls_file.touch()
        for name in ("trivy", "grype", "curl", "jq"):
            stub = self.bin / name
            stub.write_text(f"#!{sys.executable}\n{STUB}")
            stub.chmod(stub.stat().st_mode | stat.S_IEXEC)

    @property
    def is_chart(self):
        return self.task["metadata"]["name"] == "image-scan-chart"

    def script(self, step_name):
        step = next(s for s in self.task["spec"]["steps"] if s["name"] == step_name)
        params = {p["name"]: p.get("default", "") for p in self.task["spec"]["params"]}
        params.update(CODEBASE_NAME=CODEBASE, CODEBASE_BRANCH="master")
        script = re.sub(
            r"\$\(params\.([\w-]+)\)", lambda m: params[m.group(1)], step["script"]
        )
        script = (
            script.replace(
                "$(results.SCAN_REPORT_URL.path)", str(self.root / "scan-report-url")
            )
            .replace("/workspace/", f"{self.root}/")
            .replace("/tmp/dd-response.json", str(self.root / "dd-response.json"))
            .replace("/grype ", f"{self.bin}/grype ")
        )
        assert "$(params." not in script and "$(results." not in script, step_name
        return script

    def run(self, step_name, *args):
        script = self.root / f"{step_name}.sh"
        script.write_text(self.script(step_name))
        env = {
            "PATH": f"{self.bin}:{os.environ['PATH']}",
            "STUB_CALLS": str(self.calls_file),
            "TRIVY_SERVER_URL": "",
            "DD_HOST_URL": DD_URL,
            "DD_TOKEN": "token",
        }
        result = subprocess.run(
            [*self.shell, str(script), *args],
            cwd=self.workdir,
            env=env,
            capture_output=True,
            text=True,
        )
        assert (
            result.returncode == 0
        ), f"{step_name} failed:\n{result.stdout}\n{result.stderr}"

    def scan(self, images):
        """Scan and upload the images the way the Task does."""
        if self.is_chart:
            (self.root / "images.txt").write_text("".join(f"{i}\n" for i in images))
            steps = [("trivy", ()), ("grype", ()), ("upload-report", ())]
        else:
            steps = [("trivy", images), ("upload-report", images)]
        for step_name, args in steps:
            self.run(step_name, *args)

    @property
    def scans_per_image(self):
        return 2 if self.is_chart else 1

    def reports(self):
        return sorted(p.name for p in self.workdir.glob("*-report-*.json"))

    def calls(self, tool):
        """The arguments of every call of a stubbed tool."""
        recorded = (
            json.loads(line) for line in self.calls_file.read_text().splitlines()
        )
        return [call[1:] for call in recorded if call[0] == tool]

    def uploads(self):
        """The form fields of every reimport request, by field name."""
        uploads = []
        for call in self.calls("curl"):
            if REIMPORT_URL not in call:
                continue
            fields = defaultdict(list)
            for flag, value in zip(call, call[1:]):
                if flag == "-F":
                    key, _, content = value.partition("=")
                    fields[key].append(content)
            uploads.append(fields)
        return uploads


@pytest.fixture(params=TASKS)
def task_name(request):
    return request.param


@pytest.fixture(params=SHELLS)
def pipeline(request, tasks, task_name, tmp_path):
    return Pipeline(tmp_path, tasks[task_name], request.param)


@pytest.mark.parametrize("reference,engagement,version", EXAMPLES)
def test_upload_sends_the_normalized_engagement_and_version(
    pipeline, reference, engagement, version
):
    pipeline.scan([reference])

    uploads = pipeline.uploads()
    assert len(uploads) == pipeline.scans_per_image
    for upload in uploads:
        assert upload["engagement_name"] == [engagement]
        assert upload["product_name"] == [CODEBASE]
        assert upload["test_title"] == upload["scan_type"]
        if version is None:
            assert "version" not in upload
        else:
            assert upload["version"] == [version]


@pytest.mark.parametrize(
    "reference",
    [
        "docker.io/library/nginx:1.27",
        f"reg.example.com/shop/app@{DIGEST}",
        f"reg.example.com/shop/app:2.4.1@{DIGEST}",
        "reg.example.com:5000/shop/app:1.0",
        "Reg.example.com/Shop/App:1.0",
    ],
    ids=["docker-hub", "digest", "tag-and-digest", "port-and-tag", "upper-case"],
)
def test_scanners_scan_the_reference_as_given(pipeline, reference):
    pipeline.scan([reference])

    trivy = pipeline.calls("trivy")
    assert [(call[0], call[-1]) for call in trivy] == [("image", reference)]
    if pipeline.is_chart:
        assert [call[0] for call in pipeline.calls("grype")] == [reference]


def test_two_images_of_one_registry_with_a_port_get_two_report_files(pipeline):
    images = ["reg.example.com:5000/app-a", "reg.example.com:5000/app-b"]

    pipeline.scan(images)

    tools = ["trivy", "grype"] if pipeline.is_chart else ["trivy"]
    expected = sorted(
        f"{tool}-report-reg.example.com:5000_{name}.json"
        for tool in tools
        for name in ("app-a", "app-b")
    )
    assert pipeline.reports() == expected
    uploads = pipeline.uploads()
    assert len(uploads) == len(expected)
    uploaded_files = sorted(
        upload["file"][0].removeprefix("@").split(";")[0] for upload in uploads
    )
    assert uploaded_files == expected
    assert {upload["engagement_name"][0] for upload in uploads} == {
        "image:reg.example.com:5000/app-a",
        "image:reg.example.com:5000/app-b",
    }
