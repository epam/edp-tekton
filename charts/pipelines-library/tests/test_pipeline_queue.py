from .helpers import helm_template

ALL_PROVIDERS = """
global:
  gitProviders:
    - gitlab
    - github
    - bitbucket
    - gerrit
"""


def get_edp_interceptor(trigger):
    for interceptor in trigger["spec"]["interceptors"]:
        if interceptor["ref"]["name"] == "edp":
            return interceptor
    raise AssertionError("edp interceptor not found")


def get_params(interceptor):
    return {p["name"]: p["value"] for p in interceptor.get("params", [])}


def get_pipelinerun(trigger_template):
    return trigger_template["spec"]["resourcetemplates"][0]


def test_queue_disabled_by_default():
    r = helm_template(ALL_PROVIDERS)

    edp = get_edp_interceptor(r["trigger"]["gitlab-review"])
    assert "params" not in edp

    for provider in ["gitlab", "github", "bitbucket", "gerrit"]:
        run = get_pipelinerun(r["triggertemplate"][f"{provider}-review-template"])
        assert "status" not in run["spec"], provider


def test_queue_param_injected_into_review_triggers():
    r = helm_template(
        ALL_PROVIDERS
        + """
pipelines:
  queue:
    enabled: true
    """
    )

    for provider in ["gitlab", "github", "bitbucket"]:
        edp = get_edp_interceptor(r["trigger"][f"{provider}-review"])
        assert get_params(edp) == {"queuedStatusReporting": True}, provider

        build = get_edp_interceptor(r["trigger"][f"{provider}-build"])
        assert "params" not in build, provider

    # Gerrit review reporting is vote-based, the flag must not reach its trigger.
    gerrit = get_edp_interceptor(r["trigger"]["gerrit-review"])
    assert "params" not in gerrit

    # queue.enabled alone is status reporting only: the run lifecycle stays as-is.
    for provider in ["gitlab", "github", "bitbucket", "gerrit"]:
        review = get_pipelinerun(r["triggertemplate"][f"{provider}-review-template"])
        assert "status" not in review["spec"], provider


def test_queue_creates_review_pipelineruns_pending():
    r = helm_template(
        ALL_PROVIDERS
        + """
pipelines:
  queue:
    pendingPipelineRun: true
    """
    )

    # All providers, Gerrit included: queueing is VCS-agnostic, only the
    # commit status reporting is not.
    for provider in ["gitlab", "github", "bitbucket", "gerrit"]:
        review = get_pipelinerun(r["triggertemplate"][f"{provider}-review-template"])
        assert review["spec"]["status"] == "PipelineRunPending", provider

        build = get_pipelinerun(r["triggertemplate"][f"{provider}-build-template"])
        assert "status" not in build["spec"], provider


def get_gitlab_review_start_state(r, pipeline_name):
    pipeline = r["pipeline"][pipeline_name]
    start = pipeline["spec"]["tasks"][0]
    assert start["name"] == "report-pipeline-start-to-gitlab", pipeline_name

    return {p["name"]: p["value"] for p in start["params"]}["STATE"]


def test_gitlab_review_start_status_follows_queue():
    # GitLab rejects a pending -> pending transition in the same context, so with
    # the queue enabled (interceptor posts pending/QUEUED first) the review start
    # task must post `running`. Without it the long-standing `pending` default
    # is preserved for existing installations.
    pipelines = ["gitlab-go-gin-app-review", "gitlab-helm-charts-lib-review"]

    default = helm_template(
        """
global:
  gitProviders:
    - gitlab
    """
    )
    for pipeline_name in pipelines:
        assert get_gitlab_review_start_state(default, pipeline_name) == "pending", pipeline_name

    queued = helm_template(
        """
global:
  gitProviders:
    - gitlab
pipelines:
  queue:
    enabled: true
    """
    )
    for pipeline_name in pipelines:
        assert get_gitlab_review_start_state(queued, pipeline_name) == "running", pipeline_name


def test_queue_combines_with_cancel_in_progress():
    r = helm_template(
        """
global:
  gitProviders:
    - gitlab
    - gerrit
pipelines:
  cancelInProgress: true
  queue:
    enabled: true
    """
    )

    edp = get_edp_interceptor(r["trigger"]["gitlab-review"])
    assert get_params(edp) == {
        "cancelInProgress": True,
        "queuedStatusReporting": True,
    }

    gerrit = get_edp_interceptor(r["trigger"]["gerrit-review"])
    assert get_params(gerrit) == {"cancelInProgress": True}
