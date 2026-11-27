"""Behavior of the sonarqube-* tasks with SonarQube and SonarCloud (EU and US regions).

The step scripts are taken from the rendered tasks and run with stub `curl`, `mvn`, `gradle` and
`dotnet`, so the tests check what the tasks do without a Sonar server: the mode and the region chosen
from the host name of the URL, the key and the organization read from the project, and the arguments
passed on to the scanners.
"""

import os
import re
import shutil
import stat
import subprocess

import pytest

from .helpers import helm_template

CONFIG = """
global:
  gitProviders:
    - github
"""

TASKS = ["sonarqube-general", "sonarqube-maven", "sonarqube-gradle", "sonarqube-dotnet"]

EU = "https://sonarcloud.io"
US = "https://sonarqube.us"
SONARQUBE = "http://sonar.sonar:9000"


def _unavailable_tools():
    missing = [tool for tool in ("bash", "sh", "sed", "tr", "grep", "jq") if not shutil.which(tool)]
    if missing:
        return f"the sonar task scripts need {', '.join(missing)}"
    return None


UNAVAILABLE = _unavailable_tools()
requires_tools = pytest.mark.skipif(UNAVAILABLE is not None, reason=UNAVAILABLE or "tools are available")


@pytest.fixture(scope="module")
def tasks():
    return helm_template(CONFIG)["task"]


def _step(tasks, task, name):
    for step in tasks[task]["spec"]["steps"]:
        if step["name"] == name:
            return step
    raise KeyError(f"{task} has no step {name}")


def _script(step):
    return step["script"] if "script" in step else step["args"][0]


def _secret_keys(step):
    return {
        env["valueFrom"]["secretKeyRef"]["key"]
        for env in step.get("env", [])
        if "secretKeyRef" in env.get("valueFrom", {})
        and env["valueFrom"]["secretKeyRef"]["name"] == "$(params.ci-sonarqube)"
    }


# ---------------------------------------------------------------------------------------------
# structure
# ---------------------------------------------------------------------------------------------


def test_secret_keeps_only_url_and_token(tasks):
    for task in TASKS:
        for step in tasks[task]["spec"]["steps"]:
            assert _secret_keys(step) <= {"url", "token"}, (task, step["name"])
            script = _script(step) if ("script" in step or "args" in step) else ""
            assert "SONAR_ORGANIZATION" not in script, (task, step["name"])
            assert "SONAR_PROVIDER" not in script, (task, step["name"])


def test_properties_step_of_general_task_gets_no_token(tasks):
    step = _step(tasks, "sonarqube-general", "prepare-sonar-project-properties")
    assert _secret_keys(step) == {"url"}


def test_maven_runs_through_a_shell_with_the_extra_commands_as_arguments(tasks):
    step = _step(tasks, "sonarqube-maven", "mvn-goals")
    assert step["command"] == ["/bin/sh", "-c"]
    assert step["args"][1] == "--"
    assert step["args"][2] == "$(params.EXTRA_COMMANDS[*])"
    assert 'exec /usr/bin/mvn -s "${MAVEN_SETTINGS:-/var/configmap/settings.xml}"' in step["args"][0]


# ---------------------------------------------------------------------------------------------
# running the scripts
# ---------------------------------------------------------------------------------------------

MAVEN_SETTINGS = """<settings>
  <profiles>
    <profile>
      <id>sonar</id>
      <properties>
        <sonar.token>
          ${env.SONAR_TOKEN}
        </sonar.token>
        <sonar.host.url>
          ${env.SONAR_HOST_URL}
        </sonar.host.url>
      </properties>
    </profile>
  </profiles>
</settings>
"""

STUB_FAKE = """#!/bin/sh
printf '[%s]' "$(basename "$0")"
for arg in "$@"; do printf ' <%s>' "$arg"; done
echo
echo "[env] SONAR_HOST_URL=${SONAR_HOST_URL-<unset>}"
"""

STUB_CURL = """#!/bin/sh
echo "$*" >> "$STUB_LOG"
case "$*" in
  *"-w "*) printf 200 ;;
  *authentication/validate*) echo "{\\"valid\\":${FAKE_TOKEN_VALID:-true}}" ;;
  *components/show*) echo '{"errors":[{"msg":"Component key not found"}]}' ;;
  *projects/create*) echo '{}' ;;
esac
"""

# The tasks are written for GNU sed (`sed -ir`, `\\s`); BSD sed on a developer machine needs a shim.
STUB_SED = """#!/bin/bash
args=()
for a in "$@"; do
  case "$a" in -ir) args+=(-i ""); continue ;; esac
  args+=("${a//\\\\s/[[:space:]]}")
done
exec /usr/bin/sed "${args[@]}"
"""


def _is_gnu_sed():
    return subprocess.run(["sed", "--version"], capture_output=True).returncode == 0


class Sandbox:
    """Workspace, TMPDIR and stub binaries for one run of a step script."""

    def __init__(self, root):
        self.root = root
        self.ws = root / "ws"
        self.tmp = root / "tmp"
        self.bin = root / "bin"
        self.settings = root / "settings.xml"
        self.settings.write_text(MAVEN_SETTINGS)
        for directory in (self.ws, self.tmp, self.bin):
            directory.mkdir()
        for name in ("mvn", "gradle", "dotnet"):
            self._stub(name, STUB_FAKE)
        self._stub("curl", STUB_CURL)
        if not _is_gnu_sed():
            self._stub("sed", STUB_SED)
        self.log = root / "curl.log"

    def _stub(self, name, body):
        path = self.bin / name
        path.write_text(body)
        path.chmod(path.stat().st_mode | stat.S_IEXEC)

    def write(self, name, content):
        (self.ws / name).write_text(content)

    def curl_log(self):
        return self.log.read_text() if self.log.exists() else ""

    def render(self, script, extra=""):
        text = script.replace("/tmp/", f"{self.tmp}/")
        text = text.replace("$(workspaces.source.path)", str(self.ws))
        text = text.replace("$(params.EXTRA_COMMANDS[*])", "")
        text = text.replace("$(params.EXTRA_COMMANDS)", extra)
        text = text.replace("$(params.PROJECT_DIR)", ".").replace("$(params.CONTEXT_DIR)", ".")
        text = text.replace("/usr/bin/mvn", "mvn")
        text = text.replace("/var/configmap/settings.xml", str(self.settings))
        return re.sub(r"\$\(params\.[A-Za-z_-]+\)", "", text)

    def env(self, url, **extra):
        env = {
            "PATH": f"{self.bin}:{os.environ['PATH']}",
            "TMPDIR": str(self.tmp),
            "STUB_LOG": str(self.log),
            "SONAR_HOST_URL": url,
            "SONAR_TOKEN": "token",
            "SONAR_PROJECT_KEY": "codebase",
            "SONAR_PROJECT_NAME": "codebase",
            "CI_USERNAME": "user",
            "CI_PASSWORD": "pass",
        }
        env.update(extra)
        return env

    def run(self, shell, script, url, extra="", args=(), **env):
        text = self.render(script, extra)
        if shell == "sh -c":
            cmd = ["sh", "-c", text, "--", *args]
        else:
            path = self.root / "script.sh"
            path.write_text(text)
            cmd = [shell, str(path)]
        return subprocess.run(cmd, cwd=self.ws, env=self.env(url, **env), capture_output=True, text=True, timeout=60)


@pytest.fixture
def sandbox(tmp_path):
    return Sandbox(tmp_path)


def _argv(result, tool):
    """The argument vector the stub `tool` printed, as a list."""
    for line in result.stdout.splitlines():
        if line.startswith(f"[{tool}]"):
            return re.findall(r"<([^>]*)>", line)
    raise AssertionError(f"{tool} was not called:\n{result.stdout}\n{result.stderr}")


def _stub_env(result, tool):
    lines = result.stdout.splitlines()
    for index, line in enumerate(lines):
        if line.startswith(f"[{tool}]"):
            return lines[index + 1]
    raise AssertionError(f"{tool} was not called")


# ---- mode and region, chosen from the host name ----------------------------------------------

MODE_CASES = [
    (EU, True, False),
    (EU + "/", True, False),
    ("https://sonarcloud.io:443/", True, False),
    ("https://SonarCloud.io", True, False),
    ("https://eu.sonarcloud.io", True, False),
    (US, True, True),
    (US + "/", True, True),
    (SONARQUBE, False, False),
    ("https://sonar.example.com", False, False),
    ("https://example.com/sonarcloud.io", False, False),
    ("https://sonarcloud.io.example.com", False, False),
    ("https://notsonarqube.us", False, False),
]


@requires_tools
@pytest.mark.parametrize("url,cloud,us", MODE_CASES)
def test_mode_and_region_follow_the_host_name(tasks, sandbox, url, cloud, us):
    script = _script(_step(tasks, "sonarqube-general", "prepare-project"))
    result = sandbox.run("bash", script, url)

    assert result.returncode == 0, result.stdout + result.stderr
    assert ("SonarCloud mode" in result.stdout) is cloud
    assert ("SonarCloud region: us" in result.stdout) is us
    log = sandbox.curl_log()
    if cloud:
        assert "/api/system/status" in log
        assert "components/show" not in log and "projects/create" not in log
    else:
        assert "components/show" in log


@requires_tools
def test_sonarqube_still_creates_a_missing_project(tasks, sandbox):
    script = _script(_step(tasks, "sonarqube-general", "prepare-project"))
    result = sandbox.run("bash", script, SONARQUBE)

    assert result.returncode == 0, result.stdout + result.stderr
    assert "Create project codebase" in result.stdout
    assert "projects/create?name=codebase&project=codebase" in sandbox.curl_log()


@requires_tools
@pytest.mark.parametrize("url", [EU, US])
def test_sonarcloud_rejects_an_invalid_token(tasks, sandbox, url):
    script = _script(_step(tasks, "sonarqube-general", "prepare-project"))
    result = sandbox.run("bash", script, url, FAKE_TOKEN_VALID="false")

    assert result.returncode != 0
    assert "Token isn't valid" in result.stdout
    assert "authentication/validate" in sandbox.curl_log()


# ---- general: sonar-project.properties ------------------------------------------------------

PROPS_STEP = ("sonarqube-general", "prepare-sonar-project-properties")
GENERAL_ENV = {
    "SONAR_QUALITYGATE_WAIT": "true",
    "BRANCH": "main",
    "KEY_ID": "1",
    "SOURCE_BRANCH": "feat",
    "TARGET_BRANCH": "main",
}


def _run_properties(tasks, sandbox, url, content=None):
    if content is not None:
        sandbox.write("sonar-project.properties", content)
    script = _script(_step(tasks, *PROPS_STEP))
    return sandbox.run("bash", script, url, **GENERAL_ENV)


def _properties(sandbox):
    return (sandbox.ws / "sonar-project.properties").read_text().splitlines()


@requires_tools
def test_general_keeps_the_key_and_organization_of_the_project_on_sonarcloud(tasks, sandbox):
    result = _run_properties(
        tasks, sandbox, EU + "/", "sonar.projectKey=Owner_app\nsonar.organization=owner\nsonar.exclusions=**/cmd/**\n"
    )

    assert result.returncode == 0, result.stdout + result.stderr
    lines = _properties(sandbox)
    assert "sonar.projectKey=Owner_app" in lines
    assert "sonar.organization=owner" in lines
    assert "sonar.projectKey=codebase" not in lines
    assert "sonar.host.url=https://sonarcloud.io" in lines
    assert "sonar.projectName=codebase" in lines
    assert not any(line.startswith("sonar.region") for line in lines)


@requires_tools
def test_general_sets_the_region_and_no_host_url_in_the_us_region(tasks, sandbox):
    result = _run_properties(tasks, sandbox, US, "sonar.projectKey=Owner_app\nsonar.organization=owner\n")

    assert result.returncode == 0, result.stdout + result.stderr
    lines = _properties(sandbox)
    assert "sonar.region=us" in lines
    assert not any(line.startswith("sonar.host.url") for line in lines)


@requires_tools
@pytest.mark.parametrize(
    "content,missing",
    [
        ("sonar.projectKey=Owner_app\n", "sonar.organization"),
        ("sonar.organization=owner\n", "sonar.projectKey"),
        (None, "sonar-project.properties"),
    ],
)
def test_general_reports_what_sonarcloud_needs(tasks, sandbox, content, missing):
    result = _run_properties(tasks, sandbox, EU, content)

    assert result.returncode != 0
    assert missing in result.stdout


@requires_tools
def test_general_on_sonarqube_replaces_the_key_with_the_codebase_name(tasks, sandbox):
    result = _run_properties(tasks, sandbox, SONARQUBE, "sonar.projectKey=something\nsonar.organization=x\n")

    assert result.returncode == 0, result.stdout + result.stderr
    lines = _properties(sandbox)
    assert "sonar.projectKey=codebase" in lines
    assert "sonar.host.url=http://sonar.sonar:9000" in lines
    assert not any(line.startswith("sonar.region") for line in lines)


@requires_tools
def test_general_on_sonarqube_creates_the_file_when_it_is_missing(tasks, sandbox):
    result = _run_properties(tasks, sandbox, SONARQUBE)

    assert result.returncode == 0, result.stdout + result.stderr
    lines = _properties(sandbox)
    assert "sonar.projectKey=codebase" in lines
    assert not any("organization" in line for line in lines)


# ---- maven ----------------------------------------------------------------------------------

MAVEN_ARGS = [
    "-Dsonar.projectKey=codebase",
    "-Dsonar.projectName=my app",
    "-Dsonar.host.url=https://host",
    "-Dsonar.branch.name=main",
    "verify",
    "sonar:sonar",
]
POM_WITH_KEYS = "<properties><sonar.projectKey>Owner_app</sonar.projectKey><sonar.organization>owner</sonar.organization></properties>"


def _run_maven(tasks, sandbox, url, pom=POM_WITH_KEYS):
    sandbox.write("pom.xml", pom)
    script = _script(_step(tasks, "sonarqube-maven", "mvn-goals"))
    return sandbox.run("sh -c", script, url, args=MAVEN_ARGS)


@requires_tools
def test_maven_on_sonarcloud_drops_only_the_codebase_key(tasks, sandbox):
    result = _run_maven(tasks, sandbox, EU)

    assert result.returncode == 0, result.stdout + result.stderr
    assert _argv(result, "mvn") == ["-s", str(sandbox.settings)] + [
        "-Dsonar.projectName=my app",
        "-Dsonar.host.url=https://host",
        "-Dsonar.branch.name=main",
        "verify",
        "sonar:sonar",
    ]
    assert "WARNING" not in result.stdout


@requires_tools
def test_maven_in_the_us_region_drops_the_host_url_and_sets_the_region(tasks, sandbox):
    result = _run_maven(tasks, sandbox, US)

    assert result.returncode == 0, result.stdout + result.stderr
    argv = _argv(result, "mvn")
    assert argv[0] == "-s" and argv[1] != str(sandbox.settings)
    assert argv[2:] == [
        "-Dsonar.projectName=my app",
        "-Dsonar.branch.name=main",
        "verify",
        "sonar:sonar",
        "-Dsonar.region=us",
    ]
    assert _stub_env(result, "mvn") == "[env] SONAR_HOST_URL=<unset>"

    # the copy of the Maven settings loses sonar.host.url only: the scanner rejects it with the region
    copy = open(argv[1]).read()
    assert "sonar.host.url" not in copy and "SONAR_HOST_URL" not in copy
    assert "sonar.token" in copy and "<id>sonar</id>" in copy
    assert "sonar.host.url" in sandbox.settings.read_text()


@requires_tools
def test_maven_on_sonarqube_passes_the_arguments_unchanged(tasks, sandbox):
    result = _run_maven(tasks, sandbox, SONARQUBE, pom="<project/>")

    assert result.returncode == 0, result.stdout + result.stderr
    assert _argv(result, "mvn") == ["-s", str(sandbox.settings)] + MAVEN_ARGS
    assert "WARNING" not in result.stdout


@requires_tools
def test_maven_warns_when_the_pom_has_no_sonar_properties_on_sonarcloud(tasks, sandbox):
    result = _run_maven(tasks, sandbox, EU, pom="<project/>")

    assert result.returncode == 0, result.stdout + result.stderr
    assert "WARNING: sonar.projectKey is not found in pom.xml" in result.stdout
    assert "WARNING: sonar.organization is not found in pom.xml" in result.stdout


# ---- gradle ---------------------------------------------------------------------------------

GRADLE_EXTRA = "-Dsonar.projectKey=codebase -Dsonar.host.url=${SONAR_HOST_URL} sonar"
BUILD_GRADLE_WITH_KEYS = "sonarqube { properties { property 'sonar.projectKey', 'k'\nproperty 'sonar.organization', 'o' } }"


def _run_gradle(tasks, sandbox, url, build_gradle=BUILD_GRADLE_WITH_KEYS):
    if build_gradle is not None:
        sandbox.write("build.gradle", build_gradle)
    script = _script(_step(tasks, "sonarqube-gradle", "gradle-tasks"))
    return sandbox.run("bash", script, url, extra=GRADLE_EXTRA)


@requires_tools
def test_gradle_on_sonarcloud_clears_only_the_codebase_key(tasks, sandbox):
    result = _run_gradle(tasks, sandbox, EU)

    assert result.returncode == 0, result.stdout + result.stderr
    init = f"{sandbox.tmp}/sonar-init.gradle"
    assert (sandbox.tmp / "sonar-init.gradle").read_text().splitlines() == ["System.clearProperty('sonar.projectKey')"]
    argv = _argv(result, "gradle")
    assert ["-I", init] == argv[argv.index(init) - 1 : argv.index(init) + 1]
    assert "-Dsonar.host.url=https://sonarcloud.io" in argv
    assert "WARNING" not in result.stdout


@requires_tools
def test_gradle_in_the_us_region_clears_the_host_url_and_sets_the_region(tasks, sandbox):
    result = _run_gradle(tasks, sandbox, US)

    assert result.returncode == 0, result.stdout + result.stderr
    init = (sandbox.tmp / "sonar-init.gradle").read_text().splitlines()
    assert init == [
        "System.clearProperty('sonar.projectKey')",
        "System.clearProperty('sonar.host.url')",
        "System.setProperty('sonar.region', 'us')",
    ]
    assert _stub_env(result, "gradle") == "[env] SONAR_HOST_URL=<unset>"


@requires_tools
def test_gradle_on_sonarqube_has_no_init_script(tasks, sandbox):
    result = _run_gradle(tasks, sandbox, SONARQUBE, build_gradle=None)

    assert result.returncode == 0, result.stdout + result.stderr
    assert not (sandbox.tmp / "sonar-init.gradle").exists()
    assert "-Dsonar.projectKey=codebase" in _argv(result, "gradle")
    assert "WARNING" not in result.stdout


@requires_tools
def test_gradle_warns_when_the_build_has_no_sonar_properties_on_sonarcloud(tasks, sandbox):
    result = _run_gradle(tasks, sandbox, EU, build_gradle="plugins {}")

    assert result.returncode == 0, result.stdout + result.stderr
    assert "WARNING: sonar.projectKey is not found in build.gradle" in result.stdout
    assert "WARNING: sonar.organization is not found in build.gradle" in result.stdout


# ---- dotnet ---------------------------------------------------------------------------------

DOTNET_EXTRA = """dotnet sonarscanner begin /d:sonar.host.url=${SONAR_HOST_URL} /k:${SONAR_PROJECT_KEY} /n:${SONAR_PROJECT_NAME}
dotnet build
dotnet sonarscanner end /d:sonar.token=${SONAR_TOKEN}"""

XML_SINGLE_LINE = (
    "<SonarQubeAnalysisProperties>\r\n"
    '  <Property Name="sonar.projectKey">Owner_app</Property>\r\n'
    '  <Property Name="sonar.organization"> owner </Property>\r\n'
    "</SonarQubeAnalysisProperties>\r\n"
)

XML_MULTI_LINE = """<SonarQubeAnalysisProperties xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <Property xmlns:x="a"
            Name="sonar.projectKey">
    Owner_app
  </Property>
  <Property Name="sonar.organization"
  >owner</Property>
</SonarQubeAnalysisProperties>
"""


def _run_dotnet(tasks, sandbox, url, xml=None):
    if xml is not None:
        sandbox.write("SonarQube.Analysis.xml", xml)
    script = _script(_step(tasks, "sonarqube-dotnet", "sonar-scanner"))
    return sandbox.run("sh", script, url, extra=DOTNET_EXTRA)


def _dotnet_calls(result):
    return [re.findall(r"<([^>]*)>", line) for line in result.stdout.splitlines() if line.startswith("[dotnet]")]


@requires_tools
@pytest.mark.parametrize("xml", [XML_SINGLE_LINE, XML_MULTI_LINE])
def test_dotnet_on_sonarcloud_takes_the_key_and_organization_from_the_settings_file(tasks, sandbox, xml):
    result = _run_dotnet(tasks, sandbox, EU + "/", xml)

    assert result.returncode == 0, result.stdout + result.stderr
    begin, build, end = _dotnet_calls(result)
    assert begin == [
        "sonarscanner",
        "begin",
        "/o:owner",
        "/d:sonar.host.url=https://sonarcloud.io",
        "/k:Owner_app",
        "/n:codebase",
    ]
    assert build == ["build"]
    assert end == ["sonarscanner", "end", "/d:sonar.token=token"]


@requires_tools
def test_dotnet_in_the_us_region_drops_the_host_url_and_sets_the_region(tasks, sandbox):
    result = _run_dotnet(tasks, sandbox, US, XML_SINGLE_LINE)

    assert result.returncode == 0, result.stdout + result.stderr
    begin = _dotnet_calls(result)[0]
    assert begin == ["sonarscanner", "begin", "/o:owner", "/d:sonar.region=us", "/k:Owner_app", "/n:codebase"]


@requires_tools
def test_dotnet_keeps_the_wrapper_out_of_the_shared_tmp(tasks, sandbox):
    result = _run_dotnet(tasks, sandbox, EU, XML_SINGLE_LINE)

    assert result.returncode == 0, result.stdout + result.stderr
    assert not (sandbox.tmp / "dotnet").exists()


@requires_tools
@pytest.mark.parametrize(
    "xml",
    [
        '<SonarQubeAnalysisProperties><Property Name="sonar.projectKey">k</Property></SonarQubeAnalysisProperties>',
        '<SonarQubeAnalysisProperties><Property Name="sonar.organization">o</Property></SonarQubeAnalysisProperties>',
        None,
    ],
)
def test_dotnet_reports_what_sonarcloud_needs(tasks, sandbox, xml):
    result = _run_dotnet(tasks, sandbox, EU, xml)

    assert result.returncode != 0
    assert "SonarQube.Analysis.xml" in result.stdout
    assert not _dotnet_calls(result)


@requires_tools
def test_dotnet_on_sonarqube_leaves_the_commands_alone(tasks, sandbox):
    result = _run_dotnet(tasks, sandbox, SONARQUBE, XML_SINGLE_LINE)

    assert result.returncode == 0, result.stdout + result.stderr
    begin = _dotnet_calls(result)[0]
    assert begin == [
        "sonarscanner",
        "begin",
        "/d:sonar.host.url=http://sonar.sonar:9000",
        "/k:codebase",
        "/n:codebase",
    ]
