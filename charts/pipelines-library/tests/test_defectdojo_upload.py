import re

import pytest

from .helpers import helm_template

BASE = """
global:
  dnsWildCard: "example.com"
"""

EXPECTED_UPLOADERS = {"image-scan", "image-scan-chart", "image-scan-remote", "security"}

RENDERED = helm_template(BASE)


def defectdojo_uploads():
    # A DefectDojo upload that is rejected (expired or wrong token) must fail the
    # step: otherwise the Task succeeds and publishes a link built from a "null" id.
    uploads = {}
    for name, task in RENDERED.get("task", {}).items():
        for step in task["spec"]["steps"]:
            if "reimport-scan" in step.get("script", ""):
                uploads[name] = step["script"]
    return uploads


UPLOADS = defectdojo_uploads()


def upload_command(script):
    lines = script.splitlines()
    index = next(i for i, line in enumerate(lines) if "reimport-scan" in line)
    return "\n".join(lines[max(index - 1, 0) : index + 1])


def test_all_defectdojo_uploaders_are_covered():
    assert EXPECTED_UPLOADERS <= set(UPLOADS)


@pytest.mark.parametrize("task_name,script", sorted(UPLOADS.items()), ids=sorted(UPLOADS))
def test_upload_captures_the_http_status_and_keeps_curl_errors(task_name, script):
    command = upload_command(script)
    assert "HTTP_CODE=$(curl" in command, f"{task_name} does not capture the upload HTTP status"
    assert "-w '%{http_code}'" in command, f"{task_name} does not read the upload HTTP status"
    # -s alone hides curl's own error (DNS, TLS, connection) under `set -e`
    assert "-sS" in command, f"{task_name} hides curl errors"


@pytest.mark.parametrize("task_name,script", sorted(UPLOADS.items()), ids=sorted(UPLOADS))
def test_upload_fails_the_step_on_a_non_2xx_answer(task_name, script):
    status_check = re.search(r"HTTP_CODE(%\?\?|\}\" -lt 200)[^\n]*(\n[^\n]*){0,3}exit 1", script)
    assert status_check, f"{task_name} does not exit 1 on a rejected upload"


@pytest.mark.parametrize("task_name,script", sorted(UPLOADS.items()), ids=sorted(UPLOADS))
def test_report_link_is_never_built_from_a_missing_id(task_name, script):
    assert "jq -r '.results[0].id')" not in script, f"{task_name} prints null for a missing id"
    for variable in re.findall(r"^\s*(\w+_ID)=\$\(curl", script, re.MULTILINE):
        assert "jq -r '.results[0].id // empty'" in script, f"{task_name} does not map a missing id to empty"
        assert f'[ -n "${{{variable}}}" ]' in script, f"{task_name} does not fail on an empty {variable}"


def scan_task_calls():
    # (pipeline, pipeline task, clone tasks of the pipeline) for every call of a scan task
    calls = []
    for name, pipeline in RENDERED.get("pipeline", {}).items():
        tasks = pipeline["spec"].get("tasks", [])
        clones = [t["name"] for t in tasks if t.get("taskRef", {}).get("name") == "git-clone"]
        for task in tasks:
            if task.get("taskRef", {}).get("name") in EXPECTED_UPLOADERS:
                calls.append((name, task, clones))
    return calls


SCAN_TASK_CALLS = scan_task_calls()


def test_all_scan_task_callers_are_covered():
    callers = {name for name, _, _ in SCAN_TASK_CALLS}
    providers = ["bitbucket", "gerrit", "github", "gitlab"]
    assert {f"{p}-security-scan" for p in providers} | {"image-scan-remote"} <= callers


@pytest.mark.parametrize(
    "pipeline_name,task,clones",
    SCAN_TASK_CALLS,
    ids=[f"{name}/{task['name']}" for name, task, _ in SCAN_TASK_CALLS],
)
def test_pipelines_pass_only_params_the_scan_task_declares(pipeline_name, task, clones):
    declared = {p["name"] for p in RENDERED["task"][task["taskRef"]["name"]]["spec"]["params"]}
    passed = {p["name"] for p in task.get("params", [])}
    assert passed <= declared, f"{pipeline_name}/{task['name']} passes {sorted(passed - declared)}"


@pytest.mark.parametrize(
    "pipeline_name,task,clones",
    SCAN_TASK_CALLS,
    ids=[f"{name}/{task['name']}" for name, task, _ in SCAN_TASK_CALLS],
)
def test_scan_tasks_get_the_commit_of_the_clone_task(pipeline_name, task, clones):
    params = {p["name"]: p["value"] for p in task.get("params", [])}
    if clones:
        assert len(clones) == 1, f"{pipeline_name} clones more than once"
        assert params.get("COMMIT_HASH") == f"$(tasks.{clones[0]}.results.commit)"
    else:
        assert "COMMIT_HASH" not in params, f"{pipeline_name} has no clone task"
