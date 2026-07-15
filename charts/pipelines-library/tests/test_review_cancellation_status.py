from .helpers import helm_template

CANCEL_REASON_ANNOTATION = "app.edp.epam.com/queue-cancel-reason"
SUPERSEDED_DESCRIPTION = "SUPERSEDED BY NEWER COMMIT"


def get_set_status_step(task):
    for step in task["spec"]["steps"]:
        if step["name"] == "set-status":
            return step
    raise AssertionError("set-status step not found")


def get_env(step, name):
    for env in step.get("env", []):
        if env["name"] == name:
            return env
    raise AssertionError(f"env {name} not found")


def assert_cancellation_env(step):
    env = get_env(step, "QUEUE_CANCEL_REASON")
    field_path = env["valueFrom"]["fieldRef"]["fieldPath"]
    assert field_path == f"metadata.annotations['{CANCEL_REASON_ANNOTATION}']"


def test_gitlab_set_status_reports_canceled_for_cancelled_runs():
    r = helm_template(
        """
global:
  gitProviders:
    - gitlab
    """
    )

    step = get_set_status_step(r["task"]["gitlab-set-status"])
    assert_cancellation_env(step)
    assert 'STATE == "failed" and CANCEL_REASON' in step["script"]
    assert 'STATE = "canceled"' in step["script"]
    assert SUPERSEDED_DESCRIPTION in step["script"]


def test_github_set_status_reports_error_for_cancelled_runs():
    r = helm_template(
        """
global:
  gitProviders:
    - github
    """
    )

    step = get_set_status_step(r["task"]["github-set-status"])
    assert_cancellation_env(step)
    # GitHub commit statuses have no canceled state; error + description is used
    assert 'state == "failure" and cancel_reason' in step["script"]
    assert 'state = "error"' in step["script"]
    assert SUPERSEDED_DESCRIPTION in step["script"]


def test_bitbucket_set_status_reports_stopped_for_cancelled_runs():
    r = helm_template(
        """
global:
  gitProviders:
    - bitbucket
    """
    )

    step = get_set_status_step(r["task"]["bitbucket-set-status"])
    assert_cancellation_env(step)
    assert 'state == "FAILED" and cancel_reason' in step["script"]
    assert 'state = "STOPPED"' in step["script"]
    assert '(CANCELED)' in step["script"]
    assert SUPERSEDED_DESCRIPTION in step["script"]
