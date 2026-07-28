import re

from .helpers import helm_template

ALL_PROVIDERS = """
global:
  gitProviders:
    - gitlab
    - github
    - bitbucket
    - gerrit
"""

PROVIDERS = ["gitlab", "github", "bitbucket", "gerrit"]

# The templates that scaffold PipelineRuns whose pipelines carry a finally status
# reporter. Autotest/security templates are deliberately absent: those pipelines have
# no finally block, so there is no reporter for a spent budget to skip.
REPORTING_TEMPLATES = [f"{p}-{t}-template" for p in PROVIDERS for t in ["review", "build"]]


def get_pipelinerun(trigger_template):
    return trigger_template["spec"]["resourcetemplates"][0]


def parse_duration(value):
    match = re.fullmatch(r"(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?", value)
    assert match, f"unparsable duration {value!r}"
    hours, minutes, seconds = (int(g or 0) for g in match.groups())
    return hours * 3600 + minutes * 60 + seconds


def test_reporting_templates_set_both_timeouts_by_default():
    r = helm_template(ALL_PROVIDERS)

    for name in REPORTING_TEMPLATES:
        timeouts = get_pipelinerun(r["triggertemplate"][name])["spec"].get("timeouts")
        assert timeouts is not None, name
        assert set(timeouts) == {"pipeline", "finally"}, name


def test_finally_budget_is_reserved_within_the_pipeline_budget():
    """Tekton rejects finally > pipeline; a finally of zero leaves the reporter no room
    to run, which is the failure this configuration exists to prevent."""
    r = helm_template(ALL_PROVIDERS)

    for name in REPORTING_TEMPLATES:
        timeouts = get_pipelinerun(r["triggertemplate"][name])["spec"]["timeouts"]
        pipeline = parse_duration(timeouts["pipeline"])
        finally_ = parse_duration(timeouts["finally"])

        assert 0 < finally_ < pipeline, name


def test_default_tasks_budget_matches_tekton_default_timeout():
    """Tekton computes the tasks budget as `pipeline - finally`, so carving the finally
    window out of a 60m pipeline would silently shorten every build. The default adds
    the reservation on top, keeping the 60m tasks get today from
    default-timeout-minutes."""
    r = helm_template(ALL_PROVIDERS)

    for name in REPORTING_TEMPLATES:
        timeouts = get_pipelinerun(r["triggertemplate"][name])["spec"]["timeouts"]
        tasks_budget = parse_duration(timeouts["pipeline"]) - parse_duration(timeouts["finally"])

        assert tasks_budget == 60 * 60, name


def test_timeouts_are_overridable():
    r = helm_template(
        ALL_PROVIDERS
        + """
pipelines:
  timeouts:
    pipeline: 3h0m0s
    finally: 15m0s
"""
    )

    for name in REPORTING_TEMPLATES:
        timeouts = get_pipelinerun(r["triggertemplate"][name])["spec"]["timeouts"]
        assert timeouts == {"pipeline": "3h0m0s", "finally": "15m0s"}, name


def test_timeouts_omitted_unless_both_values_are_set():
    """With only one of the two set Tekton cannot derive a tasks budget, so it never
    cancels overrunning tasks and the reporter can still be skipped - a half-config
    reads as protection that is not there, so nulling either value drops the block."""
    for override in [
        "timeouts: null",
        "timeouts:\n    pipeline: null",
        "timeouts:\n    finally: null",
    ]:
        r = helm_template(
            ALL_PROVIDERS
            + f"""
pipelines:
  {override}
"""
        )

        for name in REPORTING_TEMPLATES:
            spec = get_pipelinerun(r["triggertemplate"][name])["spec"]
            assert "timeouts" not in spec, f"{name} with {override!r}"


def test_non_reporting_templates_are_left_alone():
    """Autotest and security pipelines have no finally reporter, so the reservation
    would buy them nothing."""
    r = helm_template(ALL_PROVIDERS)

    for provider in PROVIDERS:
        for name in [f"{provider}-run-autotests", f"{provider}-security-template"]:
            spec = get_pipelinerun(r["triggertemplate"][name])["spec"]
            assert "timeouts" not in spec, name
