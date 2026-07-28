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
    # script derives the state from the aggregate. A cancel-reason annotation wins
    # regardless of whether the aggregate landed on Failed (caught mid-task) or
    # Completed (landed between tasks); absent a cancel reason, Completed is a
    # legitimate when-guard-skip success shape, not a cancellation.
    r = helm_template(ALL_PROVIDERS)

    canceled_states = {
        "gitlab-set-status": 'STATE = "canceled"',
        "github-set-status": 'state = "error"',
        "bitbucket-set-status": '"STOPPED"',
    }
    completed_as_success = {
        "gitlab-set-status": 'STATE, DESCRIPTION = "success", "PASSED"',
        "github-set-status": 'state, description = "success", "Pipeline (PASSED)"',
        "bitbucket-set-status": 'state, name, description = "SUCCESSFUL", "Pipeline (PASSED)", "Review Pipeline"',
    }
    for task_name, canceled_state in canceled_states.items():
        step = get_set_status_step(r["task"][task_name])
        assert_cancellation_env(step)
        assert "PIPELINE_STATUS" in step["script"], task_name
        assert '"Succeeded"' in step["script"], task_name
        assert "CANCEL_REASON" in step["script"] or "cancel_reason" in step["script"], task_name
        assert canceled_state in step["script"], task_name
        assert SUPERSEDED_DESCRIPTION in step["script"], task_name
        # Completed without a cancel reason derives success, not a cancellation.
        assert step["script"].count(completed_as_success[task_name]) >= 1, task_name


def test_pipelines_use_single_status_reporter():
    r = helm_template(ALL_PROVIDERS)

    checked = 0
    for name, pipeline in r["pipeline"].items():
        for provider in ["gitlab", "github", "bitbucket"]:
            reporter = get_finally_task(pipeline, f"{provider}-report-pipeline-status")
            if not name.startswith(f"{provider}-"):
                assert reporter is None, name
                continue
            # Review and build pipelines report through one aggregate-driven
            # task; the guarded success/failure pair must not exist anywhere.
            assert get_finally_task(pipeline, f"{provider}-set-success-status") is None, name
            assert get_finally_task(pipeline, f"{provider}-set-failure-status") is None, name
            if "-review" in name:
                assert reporter is not None, name
            if reporter is None:
                continue
            assert "when" not in reporter, name
            assert get_param(reporter, "PIPELINE_STATUS") == "$(tasks.status)", name
            checked += 1

    assert checked > 0, "no pipelines with a status reporter rendered"
