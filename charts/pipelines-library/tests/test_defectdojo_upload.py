import re

import pytest

from .helpers import helm_template

BASE = """
global:
  dnsWildCard: "example.com"
"""

EXPECTED_UPLOADERS = {"image-scan", "image-scan-chart", "image-scan-remote", "security"}


def defectdojo_uploads():
    # A DefectDojo upload that is rejected (expired or wrong token) must fail the
    # step: otherwise the Task succeeds and publishes a link built from a "null" id.
    uploads = {}
    for name, task in helm_template(BASE).get("task", {}).items():
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
