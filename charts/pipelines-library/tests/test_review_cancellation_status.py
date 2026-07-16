from .helpers import helm_template

CANCEL_REASON_ANNOTATION = "app.edp.epam.com/queue-cancel-reason"
SUPERSEDED_DESCRIPTION = "SUPERSEDED BY NEWER COMMIT"

ALL_PROVIDERS = """
global:
  gitProviders:
    - gitlab
    - github
    - bitbucket
"""


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


def get_finally_task(pipeline, name):
    for task in pipeline["spec"].get("finally", []):
        if task["name"] == name:
            return task
    return None


def get_param(task, name):
    for param in task["params"]:
        if param["name"] == name:
            return param["value"]
    raise AssertionError(f"param {name} not found")


def assert_cancellation_env(step):
    env = get_env(step, "QUEUE_CANCEL_REASON")
    field_path = env["valueFrom"]["fieldRef"]["fieldPath"]
    assert field_path == f"metadata.annotations['{CANCEL_REASON_ANNOTATION}']"


def test_set_status_tasks_derive_state_from_aggregate():
    # Single-reporter mode: the review finally task passes $(tasks.status) and the
    # script derives the state, covering success, failure and cancellation (caught
    # mid-task -> Failed, landed between tasks -> Completed) with no guard gaps.
    r = helm_template(ALL_PROVIDERS)

    canceled_states = {
        "gitlab-set-status": 'STATE = "canceled"',
        "github-set-status": 'state = "error"',
        "bitbucket-set-status": '"STOPPED"',
    }
    for task_name, canceled_state in canceled_states.items():
        step = get_set_status_step(r["task"][task_name])
        assert_cancellation_env(step)
        assert "PIPELINE_STATUS" in step["script"], task_name
        assert '"Succeeded"' in step["script"], task_name
        assert '("Failed", "Completed")' in step["script"], task_name
        assert canceled_state in step["script"], task_name
        assert SUPERSEDED_DESCRIPTION in step["script"], task_name


def test_review_pipelines_use_single_status_reporter():
    r = helm_template(ALL_PROVIDERS)

    checked = 0
    for name, pipeline in r["pipeline"].items():
        for provider in ["gitlab", "github", "bitbucket"]:
            reporter = get_finally_task(pipeline, f"{provider}-report-pipeline-status")
            if not name.startswith(f"{provider}-"):
                assert reporter is None, name
                continue
            if "-review" not in name:
                # Cancellation reporting is a review-pipeline concern; build
                # pipelines keep their guarded success/failure vote tasks.
                assert reporter is None, name
                success = get_finally_task(pipeline, f"{provider}-set-success-status")
                failure = get_finally_task(pipeline, f"{provider}-set-failure-status")
                assert bool(success) == bool(failure), name
                continue
            assert reporter is not None, name
            assert "when" not in reporter, name
            assert get_param(reporter, "PIPELINE_STATUS") == "$(tasks.status)", name
            # The reporter must be the only status finally task; the guarded pair
            # it replaces must be gone.
            assert get_finally_task(pipeline, f"{provider}-set-success-status") is None, name
            assert get_finally_task(pipeline, f"{provider}-set-failure-status") is None, name
            checked += 1

    assert checked > 0, "no review pipelines with a status reporter rendered"
