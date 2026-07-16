from .helpers import helm_template


def get_set_status_step(task):
    for step in task["spec"]["steps"]:
        if step["name"] == "set-status":
            return step
    raise AssertionError("set-status step not found")


def test_gitlab_set_status_tolerates_invalid_transition():
    """A stuck run leaves the GitLab check in an unfinished state (pending/running);
    posting the same state again fails with 400 "Cannot transition status". The task
    must treat that conflict as already reported instead of failing the pipeline."""
    r = helm_template(
        """
global:
  gitProviders:
    - gitlab
    """
    )

    script = get_set_status_step(r["task"]["gitlab-set-status"])["script"]
    assert '"Cannot transition status" in body' in script
    assert "already reports an active state" in script
