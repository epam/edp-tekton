import http.client
import urllib.parse

from .helpers import helm_template

CONFIG = """
global:
  gitProviders:
    - gitlab
"""

REQUIRED = {
    "GITLAB_HOST_URL": "git@example.com:mike/diaspora.git",
    "REPO_FULL_NAME": "mike/diaspora",
    "SHA": "aa922fe4",
    "TARGET_URL": "https://portal.example.com/run/1",
    "STATE": "running",
}


def _task():
    return helm_template(CONFIG)["task"]["gitlab-set-status"]


def _step(task):
    for step in task["spec"]["steps"]:
        if step["name"] == "set-status":
            return step
    raise AssertionError("set-status step not found")


def _script(**overrides):
    """Render the step script with every $(params.X) resolved, so the request
    logic can be executed instead of pattern-matched."""
    task = _task()
    values = {p["name"]: p.get("default", "") for p in task["spec"]["params"]}
    values.update(REQUIRED)
    values.update(overrides)

    script = _step(task)["script"]
    for name, value in values.items():
        script = script.replace(f"$(params.{name})", str(value))

    assert "$(params." not in script, "unresolved param left in script"
    return script


class _Response:
    status = 201

    def read(self):
        return b"{}"


def _run(monkeypatch, **overrides):
    """Execute the rendered script against a stubbed GitLab and return the
    request the task issued."""
    requests = []

    class _Connection:
        def __init__(self, host):
            self.host = host

        def request(self, method, url, headers=None):
            requests.append({"host": self.host, "method": method, "url": url})

        def getresponse(self):
            return _Response()

        def close(self):
            pass

    monkeypatch.setattr(http.client, "HTTPSConnection", _Connection)
    monkeypatch.setenv("GITLAB_TOKEN", "test-token")
    monkeypatch.delenv("QUEUE_CANCEL_REASON", raising=False)

    exec(
        compile(_script(**overrides), "gitlab-set-status", "exec"),
        {"__name__": "__main__"},
    )

    assert len(requests) == 1, "task must issue exactly one request"
    return requests[0]


def _query(request):
    return urllib.parse.parse_qs(
        urllib.parse.urlparse(request["url"]).query, keep_blank_values=True
    )


def _status_tasks(pipeline):
    spec = pipeline["spec"]
    for task in spec.get("tasks", []) + spec.get("finally", []):
        if task.get("taskRef", {}).get("name") == "gitlab-set-status":
            yield task


def test_ref_defaults_to_empty():
    params = {p["name"]: p.get("default") for p in _task()["spec"]["params"]}

    assert params["REF"] == ""


def test_ref_is_sent_as_the_gitlab_ref_query_parameter(monkeypatch):
    """Callers next to GitLab CI pass refs/merge-requests/<iid>/head to join the
    merge request pipeline."""
    request = _run(monkeypatch, REF="refs/merge-requests/1911/head")

    assert request["method"] == "POST"
    assert _query(request)["ref"] == ["refs/merge-requests/1911/head"]


def test_empty_ref_leaves_the_target_to_gitlab(monkeypatch):
    request = _run(monkeypatch)

    assert "ref" not in _query(request)


def test_stock_pipelines_do_not_set_ref():
    """Setting REF to the MR ref on every pipeline would orphan the status on
    repositories without GitLab CI. No stock pipeline sets REF."""
    checked = 0
    for name, pipeline in helm_template(CONFIG)["pipeline"].items():
        for task in _status_tasks(pipeline):
            params = {p["name"] for p in task.get("params", [])}
            assert "REF" not in params, name
            checked += 1

    assert checked > 0, "no gitlab-set-status callers rendered"
