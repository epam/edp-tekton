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
NAMESPACE = "krci-ci"
COMMIT = "0f1e2d3c4b5a69788796a5b4c3d2e1f001122334"
DD_URL = "http://defectdojo.example"
REIMPORT_URL = f"{DD_URL}/api/v2/reimport-scan/"
CODEBASE_TAG = f"app.edp.epam.com/codebase:{CODEBASE}"
TENANT_TAG = f"app.edp.epam.com/tenant:{NAMESPACE}"

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
UPLOAD_TASKS = [*TASKS, "security"]

# (image reference, oci-image tag value): the cases of TestNormalizeImage in
# krci-security-operator pkg/trivy/image_test.go, whose result is the oci-image tag
# of a runtime test.
OPERATOR_IMAGE_NAMES = [
    pytest.param(
        "registry.example.com:5000/krci/app:1.0.0",
        "registry.example.com:5000/krci/app",
        id="registry-with-port-and-tag",
    ),
    pytest.param(
        "registry.example.com:5000/krci/app",
        "registry.example.com:5000/krci/app",
        id="registry-with-port",
    ),
    pytest.param(
        "registry.example.com/krci/app@sha256:abc",
        "registry.example.com/krci/app",
        id="digest",
    ),
    pytest.param(
        "registry.example.com/krci/app:1.0.0@sha256:abc",
        "registry.example.com/krci/app",
        id="tag-and-digest",
    ),
    pytest.param(
        "Registry.Example.com/KRCI/App:1",
        "registry.example.com/krci/app",
        id="upper-case",
    ),
    pytest.param("docker.io/library/nginx:1.27", "nginx", id="docker-io-library"),
    pytest.param(
        "index.docker.io/library/postgres", "postgres", id="index-docker-io-library"
    ),
    pytest.param(
        "registry-1.docker.io/valkey/valkey:9",
        "valkey/valkey",
        id="registry-1-docker-io",
    ),
    pytest.param(
        "docker.io/epamedp/cd-pipeline-operator:2.35.0",
        "epamedp/cd-pipeline-operator",
        id="docker-io-user-image",
    ),
    pytest.param("nginx:1.27", "nginx", id="implicit-docker-hub"),
    pytest.param("quay.io/library/app", "quay.io/library/app", id="non-hub-library"),
]


@pytest.fixture(scope="module")
def tasks():
    return helm_template(BASE)["task"]


class Pipeline:
    """Runs the steps of a rendered Task one after another in a shared workspace,
    with the scanners and DefectDojo replaced by stubs that record their calls."""

    def __init__(self, tmp_path, task, shell):
        self.task = task
        self.shell = shell
        self.namespace = NAMESPACE
        self.params = {p["name"]: p.get("default", "") for p in task["spec"]["params"]}
        given = {
            "CODEBASE_NAME": CODEBASE,
            "CODEBASE_BRANCH": "master",
            "COMMIT_HASH": COMMIT,
        }
        self.params.update(
            {name: value for name, value in given.items() if name in self.params}
        )
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

    @property
    def is_security(self):
        return self.task["metadata"]["name"] == "security"

    def script(self, step_name):
        step = next(s for s in self.task["spec"]["steps"] if s["name"] == step_name)
        script = re.sub(
            r"\$\(params\.([\w-]+)\)",
            lambda m: self.params[m.group(1)],
            step["script"],
        )
        script = (
            script.replace("$(context.taskRun.namespace)", self.namespace)
            .replace(
                "$(results.SCAN_REPORT_URL.path)", str(self.root / "scan-report-url")
            )
            .replace("/workspace/", f"{self.root}/")
            .replace("/tmp/dd-response.json", str(self.root / "dd-response.json"))
            .replace("/grype ", f"{self.bin}/grype ")
        )
        for variable in ("$(params.", "$(results.", "$(context."):
            assert variable not in script, f"{step_name}: {variable}"
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

    def scan(self, images=()):
        """Scan and upload the images the way the Task does; the security task
        uploads the reports of its source scanners."""
        if self.is_security:
            steps = [("upload-report", ())]
        elif self.is_chart:
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


@pytest.fixture(params=UPLOAD_TASKS)
def upload_task_name(request):
    return request.param


@pytest.fixture(params=SHELLS)
def upload_pipeline(request, tasks, upload_task_name, tmp_path):
    return Pipeline(tmp_path, tasks[upload_task_name], request.param)


@pytest.fixture(params=SHELLS)
def security_pipeline(request, tasks, tmp_path):
    return Pipeline(tmp_path, tasks["security"], request.param)


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


def test_upload_tags_the_test(upload_pipeline):
    upload_pipeline.scan(["reg.example.com/shop/app:2.4.1"])

    tags = [CODEBASE_TAG, TENANT_TAG]
    if not upload_pipeline.is_security:
        tags.append("oci-image:reg.example.com/shop/app")
    uploads = upload_pipeline.uploads()
    assert uploads
    for upload in uploads:
        assert sorted(upload["tags"]) == sorted(tags)


@pytest.mark.parametrize("name", UPLOAD_TASKS)
def test_tasks_declare_the_params_their_steps_use(tasks, name):
    spec = tasks[name]["spec"]
    declared = {param["name"] for param in spec["params"]}
    used = set(re.findall(r"\$\(params\.([\w-]+)", json.dumps(spec["steps"])))

    assert {"CODEBASE_NAME", "COMMIT_HASH"} <= used
    assert used <= declared, f"{name} uses undeclared params {sorted(used - declared)}"


def test_security_upload_keeps_its_branch_product_and_engagement(security_pipeline):
    security_pipeline.scan()

    uploads = security_pipeline.uploads()
    assert [upload["scan_type"] for upload in uploads] == [
        ["Semgrep JSON Report"],
        ["Gitleaks Scan"],
    ]
    for upload in uploads:
        assert upload["product_name"] == [CODEBASE]
        assert upload["engagement_name"] == ["code-security-master"]
        assert upload["branch_tag"] == ["master"]
        assert upload["test_title"] == upload["scan_type"]
        assert "version" not in upload


@pytest.mark.parametrize("reference,name", OPERATOR_IMAGE_NAMES)
def test_image_tests_carry_the_oci_image_tag_of_the_operator(pipeline, reference, name):
    pipeline.scan([reference])

    uploads = pipeline.uploads()
    assert len(uploads) == pipeline.scans_per_image
    for upload in uploads:
        assert f"oci-image:{name}" in upload["tags"]


def test_upload_sends_the_commit_hash(upload_pipeline):
    upload_pipeline.scan(["reg.example.com/shop/app:2.4.1"])

    uploads = upload_pipeline.uploads()
    assert uploads
    for upload in uploads:
        assert upload["commit_hash"] == [COMMIT]


def test_upload_without_a_commit_sends_no_commit_hash(upload_pipeline):
    upload_pipeline.params["COMMIT_HASH"] = ""

    upload_pipeline.scan(["reg.example.com/shop/app:2.4.1"])

    uploads = upload_pipeline.uploads()
    assert uploads
    for upload in uploads:
        assert "commit_hash" not in upload


def test_codebase_and_tenant_tags_are_lower_case(upload_pipeline):
    upload_pipeline.params["CODEBASE_NAME"] = "Shop-API"
    upload_pipeline.namespace = "Krci-CI"

    upload_pipeline.scan(["reg.example.com/shop/app:2.4.1"])

    uploads = upload_pipeline.uploads()
    assert uploads
    for upload in uploads:
        assert "app.edp.epam.com/codebase:shop-api" in upload["tags"]
        assert "app.edp.epam.com/tenant:krci-ci" in upload["tags"]
