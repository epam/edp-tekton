import http.client
import urllib.parse

import pytest

from .helpers import helm_template

CONFIG = """
global:
  gitProviders:
    - gitlab
"""


def _task():
    return helm_template(CONFIG)["task"]["gitlab-set-label"]


def _step(task):
    for step in task["spec"]["steps"]:
        if step["name"] == "set-label":
            return step
    raise AssertionError("set-label step not found")


def _script(**overrides):
    """Render the step script with every $(params.X) resolved, so the label logic
    can be executed instead of pattern-matched."""
    task = _task()
    values = {p["name"]: p.get("default", "") for p in task["spec"]["params"]}
    values.update(overrides)

    script = _step(task)["script"]
    for name, value in values.items():
        script = script.replace(f"$(params.{name})", str(value))

    assert "$(params." not in script, "unresolved param left in script"
    return script


class _Response:
    def __init__(self, status, body):
        self.status = status
        self._body = body

    def read(self):
        return self._body.encode()


def _run(monkeypatch, capsys, status=200, body="{}", **overrides):
    """Execute the rendered script against a stubbed GitLab, returning the exit
    code, stdout and the request the task issued (None if it made none)."""
    requests = []

    class _Connection:
        def __init__(self, host):
            self.host = host

        def request(self, method, url, headers=None):
            requests.append({"host": self.host, "method": method, "url": url, "headers": headers})

        def getresponse(self):
            return _Response(status, body)

        def close(self):
            pass

    monkeypatch.setattr(http.client, "HTTPSConnection", _Connection)
    monkeypatch.setenv("GITLAB_TOKEN", "test-token")

    code = 0
    try:
        exec(compile(_script(**overrides), "gitlab-set-label", "exec"), {"__name__": "__main__"})
    except SystemExit as exc:
        code = exc.code or 0

    return code, capsys.readouterr().out, (requests[0] if requests else None)


def _query(request):
    return urllib.parse.parse_qs(
        urllib.parse.urlparse(request["url"]).query, keep_blank_values=True
    )


# --- structure -------------------------------------------------------------


def test_task_is_rendered_with_expected_params():
    task = _task()

    assert task["metadata"]["name"] == "gitlab-set-label"

    defaults = {p["name"]: p.get("default") for p in task["spec"]["params"]}
    assert defaults["LABEL_PREFIX"] == "krci::"
    assert defaults["ADD_LABELS"] == ""
    assert defaults["REMOVE_LABELS"] == ""
    assert defaults["GITLAB_TOKEN_SECRET_NAME"] == "gitlab-api-secret"
    assert defaults["GITLAB_TOKEN_SECRET_KEY"] == "token"
    # required params carry no default
    assert defaults["REPO_FULL_NAME"] is None
    assert defaults["MR_IID"] is None


def test_token_is_wired_into_the_step_from_the_parameterized_secret():
    env = {e["name"]: e for e in _step(_task())["env"]}

    assert list(env) == ["GITLAB_TOKEN"], "the step takes no other environment"

    token = env["GITLAB_TOKEN"]["valueFrom"]["secretKeyRef"]
    assert token["name"] == "$(params.GITLAB_TOKEN_SECRET_NAME)"
    assert token["key"] == "$(params.GITLAB_TOKEN_SECRET_KEY)"


def test_task_is_guarded_by_the_tasks_feature_flag():
    rendered = helm_template(
        """
pipelines:
  deployableResources:
    tasks: false
"""
    )
    assert "gitlab-set-label" not in rendered.get("task", {})


# --- behaviour -------------------------------------------------------------


def test_vote_is_a_single_atomic_put_with_both_sides(monkeypatch, capsys):
    code, out, request = _run(
        monkeypatch,
        capsys,
        REPO_FULL_NAME="my-group/my-app",
        MR_IID="42",
        ADD_LABELS="passed",
        REMOVE_LABELS="running,failed",
    )

    assert code == 0
    assert request["method"] == "PUT"
    assert request["headers"]["Authorization"] == "Bearer test-token"

    # the project path is URL-encoded into a single path segment
    assert "/projects/my-group%2Fmy-app/merge_requests/42" in request["url"]

    query = _query(request)
    assert query["add_labels"] == ["krci::passed"]
    assert query["remove_labels"] == ["krci::running,krci::failed"]
    assert "PASSED" not in out


def test_labels_are_namespaced_by_the_prefix(monkeypatch, capsys):
    _, _, request = _run(
        monkeypatch,
        capsys,
        REPO_FULL_NAME="g/p",
        MR_IID="1",
        LABEL_PREFIX="krci-sanity::",
        ADD_LABELS=" passed ",
        REMOVE_LABELS="running, failed",
    )

    query = _query(request)
    assert query["add_labels"] == ["krci-sanity::passed"]
    assert query["remove_labels"] == ["krci-sanity::running,krci-sanity::failed"]


def test_add_only_and_remove_only_send_an_empty_counterpart(monkeypatch, capsys):
    _, _, add_only = _run(
        monkeypatch, capsys, REPO_FULL_NAME="g/p", MR_IID="1", ADD_LABELS="running"
    )
    assert _query(add_only)["add_labels"] == ["krci::running"]
    assert _query(add_only)["remove_labels"] == [""]

    _, _, remove_only = _run(
        monkeypatch, capsys, REPO_FULL_NAME="g/p", MR_IID="1", REMOVE_LABELS="running"
    )
    assert _query(remove_only)["add_labels"] == [""]
    assert _query(remove_only)["remove_labels"] == ["krci::running"]


def test_no_labels_configured_is_a_no_op(monkeypatch, capsys):
    code, out, request = _run(monkeypatch, capsys, REPO_FULL_NAME="g/p", MR_IID="1")

    assert code == 0
    assert request is None
    assert "nothing to do" in out


def test_non_2xx_response_fails_the_task_and_prints_the_body(monkeypatch, capsys):
    code, out, _ = _run(
        monkeypatch,
        capsys,
        status=404,
        body='{"message":"404 Not found"}',
        REPO_FULL_NAME="g/p",
        MR_IID="999",
        ADD_LABELS="passed",
    )

    assert code == 1
    assert "404 | Unable to set labels" in out
    assert "404 Not found" in out


@pytest.mark.parametrize("missing", [{"REPO_FULL_NAME": ""}, {"MR_IID": ""}])
def test_missing_required_params_fail_before_calling_gitlab(monkeypatch, capsys, missing):
    overrides = {"REPO_FULL_NAME": "g/p", "MR_IID": "1", "ADD_LABELS": "passed"}
    overrides.update(missing)

    code, out, request = _run(monkeypatch, capsys, **overrides)

    assert code == 1
    assert request is None
    assert "required" in out


def test_host_is_extracted_from_a_git_ssh_url(monkeypatch, capsys):
    _, _, request = _run(
        monkeypatch,
        capsys,
        GITLAB_HOST_URL="git@example.com:mike/diaspora.git",
        REPO_FULL_NAME="mike/diaspora",
        MR_IID="1",
        ADD_LABELS="passed",
    )

    assert request["host"] == "example.com"
